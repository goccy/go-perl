package gperl

// The C toolchain of a distribution build (dist.go).
//
// The build rides the dist's OWN build system — `perl Makefile.PL && make`
// (ExtUtils::MakeMaker) or `perl Build.PL && ./Build` (Module::Build) —
// exactly the way the CPAN toolchain normally builds XS, with two twists:
// every perl in the pipeline is the EMBEDDED interpreter (Makefile.PL and
// Build.PL run IN-PROCESS; the perl children that make rules and $^X
// re-invocations spawn exec a shim that re-enters this executable as
// `gperl run`, with a preloaded %Config overlay describing the target's C
// toolchain), and the compiler's perl-header search path is redirected to
// the SDK (materialized from the copy embedded in the xs package).
// Generated headers, xsubpp runs, extra C sources, and typemaps all keep
// working. The only command the pipeline execs is make (cc runs under
// it) — no perl needs to be installed.

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"sort"
	"strings"
	"time"
)

// xsNeutralizePPPort overwrites every bundled ppport.h in the staged tree
// with an inert stub. ppport.h is a portability shim for REAL perl headers;
// against the SDK (which presents a current perl API) its backfills must
// not activate. Pre-defining its include guard is not enough: some
// Devel::PPPort builds emit definitions (pMY_CXT and friends) BEFORE the
// guarded section, which clobber the SDK's own. Anything a dist genuinely
// needed from ppport then surfaces as an honest missing-symbol error to
// fix in the SDK instead of a silent wrong definition.
func xsNeutralizePPPort(root string) error {
	return filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil || d.IsDir() || d.Name() != "ppport.h" {
			return err
		}
		stub := "/* ppport.h neutralized by gperl: the goperl XS SDK\n" +
			" * presents a current perl API, so the portability backfills\n" +
			" * must not activate. */\n" +
			"#ifndef _P_P_PORTABILITY_H_\n" +
			"#define _P_P_PORTABILITY_H_\n" +
			"#endif\n"
		fmt.Fprintf(os.Stderr, "gperl: neutralized %s\n", path)
		return os.WriteFile(path, []byte(stub), 0o644)
	})
}

// xsArchEnv yields the environment that makes a darwin toolchain build for
// the target's architecture (Apple's perl honors ARCHFLAGS for extension
// builds).
func xsArchEnv(target string) []string {
	if flags := archFlags(target); len(flags) > 0 {
		return []string{"ARCHFLAGS=" + strings.Join(flags, " ")}
	}
	return nil
}

var perlIncLine = regexp.MustCompile(`(?m)^(PERL_INC) = .*$`)

// xsRedirectMakefile points the generated Makefile's perl-header path at
// the SDK. Only PERL_INC (the compiler search path) moves; PERL_INCDEP
// stays on the real perl CORE — it is a Makefile prerequisite path only,
// and some templates resolve it under a platform sysroot where the SDK
// does not exist. Apple's perl templates reference PERL_INC through
// -iwithsysroot (which prefixes the platform SDK root), so that spelling
// is rewritten to a plain -I as well.
func xsRedirectMakefile(makefile, sdkDir string) error {
	data, err := os.ReadFile(makefile)
	if err != nil {
		return err
	}
	out := perlIncLine.ReplaceAll(data, []byte("$1 = "+sdkDir))
	out = []byte(strings.ReplaceAll(string(out),
		`-iwithsysroot "$(PERL_INC)"`, `-I"$(PERL_INC)"`))
	return os.WriteFile(makefile, out, 0o644)
}

// targetToolchain is the %Config overlay describing the C toolchain of the
// target (goos), with cc and ar as its compiler and archiver command lines:
// the embedded interpreter's own %Config is guest-true (wasm64), and the
// build must emit compile/link rules for the platform the module runs on.
// These are the platform conventions a native perl build would have
// configured, and what the compiler says of itself.
func targetToolchain(goos, cc, ar string) map[string]string {
	common := toolchainStatic(goos, cc, ar)
	// Everything else that identifies the C toolchain is the
	// compiler's too: build scripts choose flags by the compiler's name and
	// version (gcc or clang), run its preprocessor, and look for libraries
	// where it looks. The interpreter's own values describe the compiler
	// the interpreter was built with, on the machine it was built on.
	for key, value := range hostCompiler(common["cc"]) {
		common[key] = value
	}
	// No sanitizer instrumentation: a module must not need a runtime
	// library the Go link that takes it in does not provide (zig cc
	// instruments undefined behavior by default).
	common["ccflags"] += " -fno-sanitize=undefined"
	return common
}

// toolchainStatic is the part of targetToolchain that needs no asking of
// the compiler: the programs of the toolchain, their flags and the names of
// what they make, by the platform's conventions.
func toolchainStatic(goos, cc, ar string) map[string]string {
	common := map[string]string{
		"cc": cc, "ld": cc, "ar": ar, "full_ar": ar,
		"ranlib": "ranlib", "make": "make",
		"ccflags":  "-fno-strict-aliasing -fstack-protector-strong -DPERL_USE_SAFE_PUTENV",
		"optimize": "-O2", "ldflags": "", "cccdlflags": "-fPIC",
		"ccdlflags": "", "obj_ext": ".o", "exe_ext": "", "lib_ext": ".a",
		"usedl": "define", "dlsrc": "dl_dlopen.xs", "useshrplib": "false",
	}
	switch goos {
	case "darwin":
		common["osname"] = "darwin"
		common["so"] = "dylib"
		// The embedded XSLoader looks for auto/Module/Module.so on every
		// platform; a module is a dynamic library under that name, which
		// dlopen loads as it loads a bundle.
		common["dlext"] = "so"
		common["lddlflags"] = "-shared -undefined dynamic_lookup -fstack-protector-strong"
		common["cccdlflags"] = "" // PIC is the darwin default
	default: // linux and friends
		// glibc hides its extensions unless asked; perl's own builds
		// there compile everything with _GNU_SOURCE.
		common["ccflags"] += " -D_GNU_SOURCE"
		common["osname"] = goos
		common["so"] = "so"
		common["dlext"] = "so"
		common["lddlflags"] = "-shared"
	}
	// the compiler's preprocessor, and the libraries a program is linked
	// with
	common["cpp"], common["cpprun"], common["cppstdin"] = cc+" -E", cc+" -E", cc+" -E"
	common["cppminus"], common["cpplast"], common["cppflags"] = "-", "-", common["ccflags"]
	common["libs"], common["perllibs"] = "-lm", "-lm"
	return common
}

// runToolchain is the part of %Config every program gperl runs is given in
// place of the interpreter's: the host's C toolchain, which code that runs
// it at run time (Filter::cpp's preprocessor, Inline::C, Affix::Build)
// needs, where the interpreter's values name the compiler it was built
// with on the machine it was built on. The keys that need the compiler
// asked (hostCompiler) are left out, as asking costs processes at every
// start; so is osname, which is the host's already ($^O, which the
// interpreter is told when it starts).
func runToolchain() map[string]string {
	tc := toolchainStatic(runtime.GOOS, "cc", "ar")
	delete(tc, "osname")
	return tc
}

// hostCompiler returns the %Config keys that describe the C compiler cc of
// this machine, as the compiler itself reports them: the macros it
// predefines say what it is (-dM -E), and it names the directories it
// searches for libraries and the C library it links (-print-search-dirs,
// -print-file-name). What it does not say stays empty rather than
// describing another compiler.
func hostCompiler(cc string) map[string]string {
	keys := map[string]string{
		"ccname": filepath.Base(strings.Fields(cc)[0]), "gccversion": "", "ccversion": "", "gccosandvers": "",
		"ccsymbols": "", "cppsymbols": "", "cppccsymbols": "",
		"libc": "", "gnulibc_version": "",
	}
	macros := compilerMacros(cc)
	if _, gnu := macros["__GNUC__"]; gnu {
		// perl's Configure takes __VERSION__ of a compiler that says it is
		// GNU C, which clang says too.
		keys["gccversion"] = strings.Trim(macros["__VERSION__"], `"`)
		keys["ccname"] = "gcc"
	}
	if _, clang := macros["__clang__"]; clang {
		keys["ccname"] = "clang"
	}
	if major, minor := compilerMacros(cc, "-include", "features.h")["__GLIBC__"], compilerMacros(cc, "-include", "features.h")["__GLIBC_MINOR__"]; major != "" && minor != "" {
		keys["gnulibc_version"] = major + "." + minor
	}
	if runtime.GOOS != "darwin" {
		if out, err := compilerCommand(cc, "-print-file-name=libc.so.6").Output(); err == nil {
			// a name without a directory is the compiler's way of saying
			// that it did not find the file
			if path := strings.TrimSpace(string(out)); filepath.IsAbs(path) && fileExists(path) {
				keys["libc"] = filepath.Clean(path)
			}
		}
	}
	if dirs := compilerLibraryDirs(cc); len(dirs) > 0 {
		list := strings.Join(dirs, " ")
		keys["libpth"], keys["glibpth"], keys["libspath"] = list, list, list
	}
	return keys
}

// compilerCommand is the command line cc (which may carry arguments of its
// own, as a cross compiler's does) with args.
func compilerCommand(cc string, args ...string) *exec.Cmd {
	words := strings.Fields(cc)
	return exec.Command(words[0], append(words[1:], args...)...)
}

// compilerMacros returns the macros cc predefines for C (with the extra
// arguments), name to replacement text; none when cc cannot be asked.
func compilerMacros(cc string, extra ...string) map[string]string {
	args := append([]string{"-dM", "-E", "-x", "c"}, extra...)
	out, err := compilerCommand(cc, append(args, os.DevNull)...).Output()
	if err != nil {
		return nil
	}
	macros := map[string]string{}
	for _, line := range strings.Split(string(out), "\n") {
		// "#define NAME replacement": the format of -dM
		rest, isDefine := strings.CutPrefix(line, "#define ")
		if !isDefine {
			continue
		}
		name, value, _ := strings.Cut(rest, " ")
		macros[name] = value
	}
	return macros
}

// compilerLibraryDirs returns the directories cc searches for libraries
// that exist, in its order.
func compilerLibraryDirs(cc string) []string {
	out, err := compilerCommand(cc, "-print-search-dirs").Output()
	if err != nil {
		return nil
	}
	var dirs []string
	seen := map[string]bool{}
	for _, line := range strings.Split(string(out), "\n") {
		// "libraries: =dir:dir:...": the format of -print-search-dirs
		list, isLibraries := strings.CutPrefix(line, "libraries: =")
		if !isLibraries {
			continue
		}
		for _, dir := range filepath.SplitList(list) {
			if dir == "" {
				continue
			}
			dir = filepath.Clean(dir)
			if !seen[dir] && dirExists(dir) {
				seen[dir] = true
				dirs = append(dirs, dir)
			}
		}
	}
	return dirs
}

// xsWritePerlShim writes the executable the build system knows as "perl":
// a wrapper that re-enters this gperl binary's embedded interpreter.
func xsWritePerlShim(work string) (string, error) {
	exe, err := os.Executable()
	if err != nil {
		return "", err
	}
	shim := filepath.Join(work, "perl")
	script := "#!/bin/sh\n" +
		"GPERL_INTERNAL_PERL_CLI=1\n" +
		"export GPERL_INTERNAL_PERL_CLI\n" +
		"exec \"" + exe + "\" run \"$@\"\n"
	if err := os.WriteFile(shim, []byte(script), 0o755); err != nil {
		return "", err
	}
	return shim, nil
}

// xsStandInArchlib makes the directory an XS build is told is the
// interpreter's architecture library. The compiler's CORE include
// directory is the SDK, via a CORE in it that links to the SDK (both
// ExtUtils::MakeMaker and Module::Build spell it $Config{archlibexp}/CORE);
// and since the SDK has a config.h, as a perl's headers do, the Makefile
// MakeMaker writes depends on <archlib>/Config.pm, which is the standard
// library's.
func xsStandInArchlib(work, sdkDir, stdlibDir string) (string, error) {
	dir := filepath.Join(work, "fakearchlib")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", err
	}
	if err := os.Symlink(sdkDir, filepath.Join(dir, "CORE")); err != nil {
		return "", err
	}
	if err := os.Symlink(filepath.Join(stdlibDir, "Config.pm"), filepath.Join(dir, "Config.pm")); err != nil {
		return "", err
	}
	return dir, nil
}

// xsWriteConfigOverlay writes the preload every build-driving perl run
// executes first: it replaces the guest-true toolchain/path keys of
// %Config with those of the toolchain tc (targetToolchain), so
// ExtUtils::MakeMaker and Module::Build generate a build for its target
// while everything else about %Config (version, ivsize, ...) stays
// guest-true.
func xsWriteConfigOverlay(work string, tc map[string]string, sdkDir, stdlibDir, fakeArch, perlPath string) (string, error) {
	if sdkDir != "" {
		tc["ccflags"] = tc["ccflags"] + " -I" + sdkDir
	}
	// paths: modules and xsubpp resolve from the extracted stdlib; the
	// compiler's CORE directory is the SDK (via the fake archlib).
	tc["privlibexp"] = stdlibDir
	tc["privlib"] = stdlibDir
	tc["archlibexp"] = fakeArch
	tc["archlib"] = fakeArch
	tc["vendorlibexp"] = ""
	tc["vendorarchexp"] = ""
	// The perl that programs written during the build must name on their
	// #! line (ExtUtils::MakeMaker's and ExtUtils::Helpers' script fixing)
	// is the one running the build.
	tc["perlpath"] = perlPath
	tc["startperl"] = "#!" + perlPath

	// The values go into %Config::Override, which the standard library's
	// Config.pm reports in place of its own in every view of %Config (the
	// hash, config_sh, perl -V) once it is loaded. The preload loads no
	// module itself: the program finds %INC and %Config as perl leaves them
	// before it runs (a test that assigns to %Config::Config before loading
	// Config, a -I directory that holds strict.pm, a debugger module that
	// looks at what is loaded), and Config is loaded when the program loads
	// it.
	var b strings.Builder
	b.WriteString("%Config::Override = (\n")
	keys := make([]string, 0, len(tc))
	for k := range tc {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		v := strings.ReplaceAll(tc[k], "'", "\\'")
		b.WriteString("    '" + k + "' => '" + v + "',\n")
	}
	b.WriteString(");\n")
	// Build tools branch on $^O, not just $Config{osname}: the build must
	// look like the host it targets.
	b.WriteString("$^O = '" + tc["osname"] + "';\n1;\n")

	path := filepath.Join(work, "config-overlay.pl")
	if err := os.WriteFile(path, []byte(b.String()), 0o644); err != nil {
		return "", err
	}
	return path, nil
}

func fileExists(path string) bool {
	st, err := os.Stat(path)
	return err == nil && !st.IsDir()
}

func dirExists(path string) bool {
	st, err := os.Stat(path)
	return err == nil && st.IsDir()
}

// errLinkInsideCopy is copyTree's error for a symbolic link to a directory
// the copy is already inside (one that holds the link, or one a loop of
// links leads back to): the copy would never end.
var errLinkInsideCopy = errors.New("a link to a directory the copy is already inside")

// copyTree copies src to dst, following symbolic links. A link to a
// directory the copy is already inside fails with errLinkInsideCopy.
func copyTree(src, dst string) error {
	root, err := filepath.Abs(src)
	if err != nil {
		return err
	}
	if root, err = filepath.EvalSymlinks(root); err != nil {
		return err
	}
	return copyTreeFollowing(root, dst, []string{root})
}

// copyTreeFollowing copies src, a path without symbolic links, to dst.
// following lists the directories whose copy holds this one: the root of
// the copy and each linked directory followed to reach src.
func copyTreeFollowing(src, dst string, following []string) error {
	return filepath.WalkDir(src, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}
		target := filepath.Join(dst, rel)
		if d.IsDir() {
			return os.MkdirAll(target, 0o755)
		}
		info, err := d.Info()
		if err != nil {
			return err
		}
		modTime := info.ModTime()
		if d.Type()&fs.ModeSymlink != 0 {
			resolved, err := filepath.EvalSymlinks(path)
			if err != nil {
				return err
			}
			st, err := os.Stat(resolved)
			if err != nil {
				return err
			}
			if st.IsDir() {
				for _, dir := range following {
					rel, err := filepath.Rel(resolved, dir)
					if err != nil {
						return err
					}
					if filepath.IsLocal(rel) {
						return fmt.Errorf("%s -> %s: %w", path, resolved, errLinkInsideCopy)
					}
				}
				return copyTreeFollowing(resolved, target, append(following, resolved))
			}
			path, modTime = resolved, st.ModTime()
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		mode := fs.FileMode(0o644)
		if info.Mode()&0o100 != 0 {
			mode = 0o755
		}
		if err := os.WriteFile(target, data, mode); err != nil {
			return err
		}
		return keepModTime(target, modTime)
	})
}

// keepModTime gives the copy at path the modification time of the file it
// copies, as tar extraction and ExtUtils::Install do: a build's make
// compares the times of a distribution's files.
func keepModTime(path string, modTime time.Time) error {
	return os.Chtimes(path, time.Time{}, modTime)
}
