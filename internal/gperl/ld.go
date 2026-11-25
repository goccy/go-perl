package gperl

// The linker of a distribution build.
//
// A distribution's own build system (ExtUtils::MakeMaker, Module::Build)
// compiles the XS sources and then runs $Config{ld} to produce the
// loadable module. The %Config overlay points ld at a wrapper that
// re-enters this executable, so the final step is gperl's: the module's
// objects are merged into ONE relocatable object in which only the
// module's entry table keeps its name, and both artifacts come from that
// object —
//
//	auto/<Module/Path>/<Name>.a        linked statically into a Go binary
//	auto/<Module/Path>/<Name>.so       loaded at run time
//	auto/<Module/Path>/<Name>.ldflags  the libraries either needs
//	auto/<Module/Path>/<Name>.exports  the names the .so exports
//
// With nothing else visible, two modules can carry the same helper names
// and still link into one binary. A module's references to what a native
// module of its dependencies exports are given that module's names.
//
// The C compiler is $CC (default cc) and the archiver $AR (default ar);
// for an ELF target nothing else is needed, so a cross compiler such as
// `zig cc -target x86_64-linux-gnu` builds modules for another platform.

import (
	"debug/elf"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"slices"
	"strings"
)

// ldEnv marks a process as a linker re-entry; ldArg0 is its first argument.
const (
	ldEnv  = "GPERL_INTERNAL_LD"
	ldArg0 = "gperl-internal-ld"
	// ldSDKEnv carries the SDK header directory to the linker, ldTargetEnv
	// the target the module is built for ("linux_amd64"), and ldNativeEnv
	// the target directories of the distributions the module's distribution
	// depends on.
	ldSDKEnv    = "GPERL_INTERNAL_SDK"
	ldTargetEnv = "GPERL_INTERNAL_TARGET"
	ldNativeEnv = "GPERL_INTERNAL_NATIVE"
)

func init() {
	if os.Getenv(ldEnv) == "" || len(os.Args) < 2 || os.Args[1] != ldArg0 {
		return
	}
	if err := linkModule(os.Args[2:], os.Getenv(ldSDKEnv), os.Getenv(ldTargetEnv), filepath.SplitList(os.Getenv(ldNativeEnv))); err != nil {
		fmt.Fprintf(os.Stderr, "gperl ld: %v\n", err)
		os.Exit(1)
	}
	os.Exit(0)
}

// writeLdShim writes the executable the build system knows as the linker.
func writeLdShim(work string) (string, error) {
	exe, err := os.Executable()
	if err != nil {
		return "", err
	}
	shim := filepath.Join(work, "gperl-ld")
	script := "#!/bin/sh\n" + ldEnv + "=1\nexport " + ldEnv + "\nexec \"" + exe + "\" " + ldArg0 + " \"$@\"\n"
	return shim, os.WriteFile(shim, []byte(script), 0o755)
}

// toolCommand returns the command line of a build tool: the words of the
// environment variable name, or fallback when it is unset.
func toolCommand(name, fallback string) []string {
	if words := strings.Fields(os.Getenv(name)); len(words) > 0 {
		return words
	}
	return []string{fallback}
}

// ldCommand is a parsed link command line.
type ldCommand struct {
	out    string
	inputs []string // objects and static archives
	libs   []string // flags naming libraries and where to find them
	module string   // mangled Perl package ("Cpanel__JSON__XS"), "" when not a module link
}

// parseLd sorts a link command line into what the module link needs. The
// module is recognized by its output path, which every build system
// spells .../auto/<Module/Path>/<Name>.so.
func parseLd(args []string) ldCommand {
	var cmd ldCommand
	for i := 0; i < len(args); i++ {
		a := args[i]
		switch {
		case a == "-o" && i+1 < len(args):
			i++
			cmd.out = args[i]
		case a == "-framework" && i+1 < len(args):
			cmd.libs = append(cmd.libs, a, args[i+1])
			i++
		case strings.HasPrefix(a, "-l"), strings.HasPrefix(a, "-L"), strings.HasPrefix(a, "-F"),
			strings.HasPrefix(a, "-Wl,"), a == "-pthread":
			cmd.libs = append(cmd.libs, a)
		case strings.HasSuffix(a, ".o"), strings.HasSuffix(a, ".a"):
			cmd.inputs = append(cmd.inputs, a)
		case strings.HasSuffix(a, ".dylib"), strings.HasSuffix(a, ".so"), strings.HasSuffix(a, ".tbd"):
			cmd.libs = append(cmd.libs, a)
		}
	}
	slash := filepath.ToSlash(cmd.out)
	if i := strings.LastIndex(slash, "/auto/"); i >= 0 && strings.HasSuffix(slash, ".so") {
		cmd.module = strings.ReplaceAll(filepath.ToSlash(filepath.Dir(slash[i+len("/auto/"):])), "/", "__")
	}
	return cmd
}

// linkModule performs one link for target; natives are the target
// directories of the distributions the module's distribution depends on.
// Anything that is not a module link (a configure probe, a helper
// executable) goes to the C compiler unchanged.
func linkModule(args []string, sdkDir, target string, natives []string) error {
	cmd := parseLd(args)
	cc := toolCommand("CC", "cc")
	if cmd.module == "" || len(cmd.inputs) == 0 {
		return run(append(cc, args...)...)
	}
	if sdkDir == "" {
		return errors.New(ldSDKEnv + " is not set")
	}
	if target == "" {
		target = runtime.GOOS + "_" + runtime.GOARCH
	}
	cc = append(cc, archFlags(target)...)
	base := strings.TrimSuffix(cmd.out, ".so")
	if err := os.MkdirAll(filepath.Dir(cmd.out), 0o755); err != nil {
		return err
	}
	entry := base + ".entry.o"
	if err := run(append(cc, "-fPIC", "-w", "-fno-sanitize=undefined", "-I"+sdkDir,
		"-DGOPERL_ENTRY_BOOT=boot_"+cmd.module, "-DGOPERL_ENTRY_TABLE=goperl_entry_"+cmd.module,
		"-c", "-o", entry, filepath.Join(sdkDir, "goperl_entry.c"))...); err != nil {
		return err
	}
	// The loadable module is the link the build system asked for: the
	// interpreter's loaders find its boot function by name, and a module
	// that lends its symbols to others (dl_load_flags) keeps them.
	if err := run(append(cc, args...)...); err != nil {
		return err
	}
	imports, err := moduleImports(cc, target, cmd.out, cmd.libs, natives)
	if err != nil {
		return err
	}
	merged := base + ".merged.o"
	inputs := append(append([]string{}, cmd.inputs...), entry)
	exports, err := mergeObjects(cc, target, merged, "goperl_entry_"+cmd.module, inputs, imports)
	if err != nil {
		return err
	}
	if err := os.WriteFile(base+".exports", []byte(strings.Join(exports, "\n")), 0o644); err != nil {
		return err
	}
	archive := base + ".a"
	if err := os.Remove(archive); err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}
	if err := run(append(toolCommand("AR", "ar"), "rcs", archive, merged)...); err != nil {
		return err
	}
	deps, err := staticDeps(cc, target, cmd.out, cmd.libs)
	if err != nil {
		return err
	}
	if err := os.WriteFile(base+".ldflags", []byte(strings.Join(append(cmd.libs, deps...), "\n")), 0o644); err != nil {
		return err
	}
	return errors.Join(os.Remove(entry), os.Remove(merged))
}

// staticDeps returns what a static link of the module so needs beyond
// libs: the shared libraries so depends on as the linker recorded them
// (DT_NEEDED), and theirs in turn, as -l flags, each before the libraries
// it depends on, since a static link reads each archive once. Only a
// library whose shared object (as the linker takes one: an ELF shared
// object of the module's machine and class) and archive (lib<name>.a) are
// in the directories a -l searches (libs' -L, then the compiler's own) is
// named; one without stays out, and a static link that needs it fails with
// the linker's own error. A darwin module, never linked statically, needs
// none.
func staticDeps(cc []string, target, so string, libs []string) ([]string, error) {
	if strings.HasPrefix(target, "darwin_") {
		return nil, nil
	}
	dirs := libraryDirs(cc, libs)
	// lookup returns the first file named name in dirs, "" when there is
	// none.
	lookup := func(name string) string {
		for _, dir := range dirs {
			if path := filepath.Join(dir, name); fileExists(path) {
				return path
			}
		}
		return ""
	}
	var order []string // each library after the ones it depends on
	err := walkNeeded(so, dirs, func(name string, _ *elf.File) error {
		if lookup("lib"+name+".a") != "" {
			order = append(order, "-l"+name)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	slices.Reverse(order)
	return order, nil
}

// libraryDirs returns the directories a -l searches: libs' -L, then the
// compiler's own.
func libraryDirs(cc, libs []string) []string {
	var dirs []string
	for _, lib := range libs {
		if dir, ok := strings.CutPrefix(lib, "-L"); ok && dir != "" {
			dirs = append(dirs, dir)
		}
	}
	return append(dirs, compilerLibraryDirs(strings.Join(cc, " "))...)
}

// walkNeeded calls fn, once each, for the shared libraries so depends on
// as the linker recorded them (DT_NEEDED) and theirs in turn, each after
// the ones it depends on, with its -l name: those a -l can name whose
// shared object, as the linker takes one, is in dirs.
func walkNeeded(so string, dirs []string, fn func(name string, f *elf.File) error) error {
	module, err := elf.Open(so)
	if err != nil {
		return fmt.Errorf("%s: %w", so, err)
	}
	// sharedObject opens the first file named soname in dirs that the
	// linker takes for the module, skipping the others as it does (a
	// linker script, another machine's library); nil when there is none.
	sharedObject := func(soname string) (*elf.File, string, error) {
		for _, dir := range dirs {
			path := filepath.Join(dir, soname)
			if !fileExists(path) {
				continue
			}
			f, err := elf.Open(path)
			var notELF *elf.FormatError
			if errors.As(err, &notELF) {
				continue
			}
			if err != nil {
				return nil, "", fmt.Errorf("%s: %w", path, err)
			}
			if f.Type == elf.ET_DYN && f.Machine == module.Machine && f.Class == module.Class {
				return f, path, nil
			}
			if err := f.Close(); err != nil {
				return nil, "", fmt.Errorf("%s: %w", path, err)
			}
		}
		return nil, "", nil
	}
	seen := map[string]bool{}
	// visit walks what f depends on, then calls fn for f unless it is the
	// module (name "").
	var visit func(f *elf.File, path, name string) error
	visit = func(f *elf.File, path, name string) (err error) {
		defer func() {
			if cerr := f.Close(); cerr != nil {
				err = errors.Join(err, fmt.Errorf("%s: %w", path, cerr))
			}
		}()
		needed, err := f.ImportedLibraries()
		if err != nil {
			return fmt.Errorf("%s: %w", path, err)
		}
		for _, soname := range needed {
			if seen[soname] {
				continue
			}
			seen[soname] = true
			// -l<name> finds lib<name>.so[.<version>] and lib<name>.a: a
			// library named otherwise (the dynamic loader) has no -l.
			rest, isLib := strings.CutPrefix(soname, "lib")
			depName, _, isSO := strings.Cut(rest, ".so")
			if !isLib || !isSO || depName == "" {
				continue
			}
			dep, file, err := sharedObject(soname)
			if err != nil {
				return err
			}
			if dep == nil {
				continue
			}
			if err := visit(dep, file, depName); err != nil {
				return err
			}
		}
		if name == "" {
			return nil
		}
		if err := fn(name, f); err != nil {
			return fmt.Errorf("%s: %w", path, err)
		}
		return nil
	}
	return visit(module, so, "")
}

// moduleImports returns the names privatizeELF gives what the module so
// takes from a native module that exports it (its .exports, in the auto/
// directory of one of natives): the name with that module's suffix. A name
// a library so depends on defines stays, as perl resolves it there; one
// that two modules export is an error. A darwin module imports nothing.
func moduleImports(cc []string, target, so string, libs, natives []string) (map[string]string, error) {
	if strings.HasPrefix(target, "darwin_") || len(natives) == 0 {
		return nil, nil
	}
	exporters := map[string][]string{} // name -> the modules exporting it
	for _, native := range natives {
		auto := filepath.Join(native, "auto")
		err := filepath.WalkDir(auto, func(path string, d fs.DirEntry, err error) error {
			if err != nil || d.IsDir() || filepath.Ext(path) != ".exports" {
				return err
			}
			rel, err := filepath.Rel(auto, filepath.Dir(path))
			if err != nil {
				return err
			}
			data, err := os.ReadFile(path)
			if err != nil {
				return err
			}
			// The module's suffix, as parseLd and mergeObjects make it.
			module := strings.ReplaceAll(filepath.ToSlash(rel), "/", "__")
			for _, name := range strings.Fields(string(data)) {
				exporters[name] = append(exporters[name], module)
			}
			return nil
		})
		if err != nil {
			return nil, err
		}
	}
	f, err := elf.Open(so)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", so, err)
	}
	syms, err := f.ImportedSymbols()
	if err := errors.Join(err, f.Close()); err != nil {
		return nil, fmt.Errorf("%s: %w", so, err)
	}
	imports := map[string]string{}
	for _, s := range syms {
		if len(exporters[s.Name]) > 0 {
			imports[s.Name] = ""
		}
	}
	if len(imports) == 0 {
		return nil, nil
	}
	err = walkNeeded(so, libraryDirs(cc, libs), func(_ string, lib *elf.File) error {
		defined, err := lib.DynamicSymbols()
		for _, s := range defined {
			if s.Section != elf.SHN_UNDEF {
				delete(imports, s.Name)
			}
		}
		return err
	})
	if err != nil {
		return nil, err
	}
	for name := range imports {
		modules := exporters[name]
		if len(modules) > 1 {
			return nil, fmt.Errorf("%s: %s is exported by native modules %s", so, name, strings.Join(modules, " and "))
		}
		imports[name] = name + "..gperl." + modules[0]
	}
	return imports, nil
}

// archFlags pins a darwin toolchain to the target's architecture (the
// compiler's default follows the process tree, not the Go target).
func archFlags(target string) []string {
	goos, goarch, _ := strings.Cut(target, "_")
	if goos != "darwin" {
		return nil
	}
	if arch := map[string]string{"amd64": "x86_64", "arm64": "arm64"}[goarch]; arch != "" {
		return []string{"-arch", arch}
	}
	return nil
}

// mergeObjects links inputs into one relocatable object in which only
// symbol keeps its name, and returns what privatizeELF does; imports
// renames its undefined symbols. A Mach-O object exports nothing and
// imports nothing.
func mergeObjects(cc []string, target, out, symbol string, inputs []string, imports map[string]string) ([]string, error) {
	if strings.HasPrefix(target, "darwin_") {
		// Mach-O: the linker itself can leave one symbol exported.
		args := append([]string{"ld", "-r"}, archFlags(target)...)
		return nil, run(append(append(args, "-exported_symbol", "_"+symbol, "-o", out), inputs...)...)
	}
	// ELF: the compiler driver merges, and the symbols are made private
	// here — no binutils needed, whatever the host.
	if err := run(append(append(cc, "-r", "-nostdlib", "-o", out), inputs...)...); err != nil {
		return nil, err
	}
	return privatizeELF(out, symbol, "..gperl."+strings.TrimPrefix(symbol, "goperl_entry_"), imports)
}

// run executes a toolchain command line, passing its output through.
func run(argv ...string) error {
	c := exec.Command(argv[0], argv[1:]...)
	c.Stdout, c.Stderr = os.Stderr, os.Stderr
	if err := c.Run(); err != nil {
		return fmt.Errorf("%s: %w", strings.Join(argv, " "), err)
	}
	return nil
}
