package gperl

// Building the distributions of the module directory.
//
// A distribution whose metadata allows a static install is used as
// extracted. Every other one is built once per machine: its own build
// system (Makefile.PL + make, or Build.PL + Build) runs in a scratch copy,
// driven by the embedded interpreter, and what it produces is kept in the
// distribution's tree —
//
//	gperl/lib/                     the built module tree (blib/lib)
//	gperl/bin/                     its scripts
//	gperl/MYMETA.json              the prereqs configure settled on
//	gperl/abi<N>/                  the Go package of its native modules:
//	  go.mod, xs.go                  links the .a files, exports the entry tables
//	  <goos>_<goarch>/auto/...       <Name>.a, <Name>.so, <Name>.ldflags

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"sync"

	perl "github.com/goccy/go-perl"
	"github.com/goccy/go-perl/internal/mod"
	"github.com/goccy/go-perl/xs"
)

// cpanModulePath is the Go module path prefix of a distribution's native
// package.
const cpanModulePath = "github.com/goccy/go-perl/cpan"

// syncDeps makes the project's dependencies ready to use on target: resolved (when
// resolve is set or the project has no snapshot yet), fetched, and built.
// Building can reveal prereqs that only a distribution's configure step
// knows, so resolution and building alternate until nothing new appears.
func syncDeps(ctx context.Context, m *mod.Manager, projectDir, target string, resolve bool, upgrade ...string) (*mod.Snapshot, error) {
	projectDir, err := filepath.Abs(projectDir)
	if err != nil {
		return nil, err
	}
	if !fileExists(filepath.Join(projectDir, mod.SnapshotFile)) {
		if !fileExists(filepath.Join(projectDir, mod.CPANfile)) {
			return &mod.Snapshot{}, nil
		}
		resolve = true
	}
	if !resolve {
		snap, err := m.Download(ctx, projectDir)
		if err != nil {
			return nil, err
		}
		_, err = buildSnapshot(ctx, m, snap, target)
		return snap, err
	}
	for {
		snap, err := m.Tidy(ctx, projectDir, upgrade...)
		if err != nil {
			return nil, err
		}
		built, err := buildSnapshot(ctx, m, snap, target)
		if err != nil || built == 0 {
			return snap, err
		}
		upgrade = nil
	}
}

// buildSnapshot builds every distribution of the snapshot that is not
// ready yet, dependencies first, independent ones in parallel. It returns
// how many it built.
func buildSnapshot(ctx context.Context, m *mod.Manager, snap *mod.Snapshot, target string) (int, error) {
	type node struct {
		dist *mod.Distribution
		tree string
		deps []int
	}
	nodes := make([]node, len(snap.Distributions))
	index := map[string]int{}
	pending := 0
	for i := range snap.Distributions {
		d := &snap.Distributions[i]
		tree, err := mod.TreePath(m.Dir, d.Pathname)
		if err != nil {
			return 0, err
		}
		nodes[i] = node{dist: d, tree: tree}
		index[d.Pathname] = i
		if !mod.Built(tree, target) {
			pending++
		}
	}
	if pending == 0 {
		return 0, nil
	}
	for i := range nodes {
		for _, req := range nodes[i].dist.Requirements {
			if p := snap.Provider(req.Name); p != nil && index[p.Pathname] != i {
				nodes[i].deps = append(nodes[i].deps, index[p.Pathname])
			}
		}
	}

	done := make([]bool, len(nodes))
	built := 0
	for remaining := len(nodes); remaining > 0; {
		var wave []int
		for i := range nodes {
			if done[i] {
				continue
			}
			ready := true
			for _, dep := range nodes[i].deps {
				ready = ready && done[dep]
			}
			if ready {
				wave = append(wave, i)
			}
		}
		if len(wave) == 0 {
			// A dependency cycle: build what is left in snapshot order.
			for i := range nodes {
				if !done[i] {
					wave = append(wave, i)
					break
				}
			}
		}
		var (
			wg    sync.WaitGroup
			mu    sync.Mutex
			first error
			slots = make(chan struct{}, runtime.NumCPU())
		)
		for _, i := range wave {
			if mod.Built(nodes[i].tree, target) {
				continue
			}
			built++
			wg.Add(1)
			slots <- struct{}{}
			go func() {
				defer wg.Done()
				defer func() { <-slots }()
				// Everything the distribution may load while it builds:
				// the trees of all the distributions before it.
				var inc []string
				for j := range nodes {
					if done[j] {
						inc = append(inc, mod.LibDir(nodes[j].tree))
					}
				}
				if err := buildDist(ctx, m, nodes[i].tree, target, inc); err != nil {
					mu.Lock()
					if first == nil {
						first = fmt.Errorf("build %s: %w", nodes[i].dist.Name, err)
					}
					mu.Unlock()
				}
			}()
		}
		wg.Wait()
		if first != nil {
			return built, first
		}
		for _, i := range wave {
			done[i] = true
		}
		remaining -= len(wave)
	}
	return built, nil
}

// sdkDirs records the module directories whose SDK headers this process
// has materialized: path -> *sdkState.
var sdkDirs sync.Map

type sdkState struct {
	once sync.Once
	err  error
}

// sdkDirIn materializes the SDK headers in the module directory, once per
// process and directory, and returns their path.
func sdkDirIn(m *mod.Manager) (string, error) {
	path := filepath.Join(m.Dir, "cache", "sdk", filepath.Base(mod.ABIDir("")))
	entry, _ := sdkDirs.LoadOrStore(path, &sdkState{})
	state := entry.(*sdkState)
	state.once.Do(func() { state.err = xs.WriteSDK(path) })
	return path, state.err
}

// buildDist makes one distribution ready on target. inc lists the module
// directories its build may load from.
func buildDist(ctx context.Context, m *mod.Manager, tree, target string, inc []string) (err error) {
	meta, err := m.Meta(tree)
	if err != nil {
		return err
	}
	// A static install is "copy lib/"; with a share directory it also
	// installs files a build step would place, so those are built.
	if meta.StaticInstall && !meta.HasXS && !dirExists(filepath.Join(tree, "share")) {
		return mod.MarkBuilt(tree, "")
	}

	work, err := os.MkdirTemp("", "gperl-build-")
	if err != nil {
		return err
	}
	defer func() { err = errors.Join(err, os.RemoveAll(work)) }()
	src := filepath.Join(work, "src")
	if err := copyDistSource(tree, src); err != nil {
		return err
	}
	if err := xsNeutralizePPPort(src); err != nil {
		return err
	}
	sdkDir, err := sdkDirIn(m)
	if err != nil {
		return fmt.Errorf("materialize SDK headers: %w", err)
	}
	perlShim, err := xsWritePerlShim(work)
	if err != nil {
		return err
	}
	ldShim, err := writeLdShim(work)
	if err != nil {
		return err
	}
	stdlibDir, err := perl.ExtractStdlib()
	if err != nil {
		return err
	}
	fakeArch, err := xsStandInArchlib(work, sdkDir, stdlibDir)
	if err != nil {
		return err
	}
	// A distribution without native code never runs the compiler, so a
	// machine without one can still build it.
	ccWords, arWords, err := cToolchain(target)
	if err != nil {
		if meta.HasXS {
			return err
		}
		ccWords, arWords = []string{"cc"}, []string{"ar"}
	}
	cc, ar := strings.Join(ccWords, " "), strings.Join(arWords, " ")
	goos, _, _ := strings.Cut(target, "_")
	tc := targetToolchain(goos, cc, ar)
	// The module link is gperl's (ld.go).
	tc["ld"] = ldShim
	overlay, err := xsWriteConfigOverlay(work, tc, sdkDir, stdlibDir, fakeArch, perlShim)
	if err != nil {
		return err
	}
	env := append(os.Environ(), xsArchEnv(target)...)
	env = append(env,
		"CC="+cc,
		"AR="+ar,
		ldTargetEnv+"="+target,
		"GPERL_PERL_PRELOAD="+overlay,
		"GPERL_PERL_EXE="+perlShim,
		"PERL="+perlShim,
		"PERL_MM_USE_DEFAULT=1",
		"PERL5LIB="+strings.Join(inc, string(os.PathListSeparator)),
		ldSDKEnv+"="+sdkDir,
		"PATH="+filepath.Dir(perlShim)+string(os.PathListSeparator)+os.Getenv("PATH"),
	)
	// As under cpanm, the build finds the modules a distribution carries
	// in its own directory (inc/) unless the user says otherwise.
	if _, ok := os.LookupEnv("PERL_USE_UNSAFE_INC"); !ok {
		env = append(env, "PERL_USE_UNSAFE_INC=1")
	}

	var log bytes.Buffer
	// A program the interpreter runs is the process (its working directory
	// and environment are the process's), and distributions build side by
	// side: each perl of a build is a process of its own, as the ones make
	// starts are.
	perlRun := func(args ...string) error {
		cmd := exec.CommandContext(ctx, perlShim, args...)
		cmd.Dir, cmd.Env, cmd.Stdout, cmd.Stderr = src, env, &log, &log
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("perl %s: %w", strings.Join(args, " "), err)
		}
		return nil
	}
	// As under cpanm, Build.PL has configured the distribution when it exits
	// 0 and writes Build; otherwise Makefile.PL, if there is one, does.
	var buildPLErr error
	useBuildPL := fileExists(filepath.Join(src, "Build.PL"))
	if useBuildPL {
		if buildPLErr = perlRun("Build.PL"); buildPLErr == nil && !fileExists(filepath.Join(src, "Build")) {
			buildPLErr = errors.New("perl Build.PL wrote no Build")
		}
		useBuildPL = buildPLErr == nil || !fileExists(filepath.Join(src, "Makefile.PL"))
	}
	switch {
	case useBuildPL:
		if err = buildPLErr; err == nil {
			err = perlRun("Build")
		}
	case fileExists(filepath.Join(src, "Makefile.PL")):
		if err = perlRun("Makefile.PL", "PERL="+perlShim, "FULLPERL="+perlShim); err == nil {
			if err = xsRedirectMakefile(filepath.Join(src, "Makefile"), sdkDir); err == nil {
				mkPath, lookErr := exec.LookPath("make")
				if lookErr != nil {
					return fmt.Errorf("the distribution builds with make, which was not found on PATH: %w", lookErr)
				}
				mk := exec.CommandContext(ctx, mkPath)
				mk.Dir, mk.Env, mk.Stdout, mk.Stderr = src, env, &log, &log
				if err = mk.Run(); err != nil {
					err = fmt.Errorf("make: %w", err)
				}
			}
		}
		if err != nil && buildPLErr != nil {
			err = errors.Join(buildPLErr, err)
		}
	default:
		err = errors.New("no Build.PL or Makefile.PL")
	}
	if err != nil {
		// The whole log stays beside where the artifacts would be; the
		// error carries its end.
		note := filepath.Join(mod.TargetDir(tree, target), buildErrorLog)
		if werr := os.MkdirAll(filepath.Dir(note), 0o755); werr != nil {
			return errors.Join(err, werr)
		}
		if werr := os.WriteFile(note, log.Bytes(), 0o644); werr != nil {
			return errors.Join(err, werr)
		}
		return fmt.Errorf("%w (full log: %s)\n%s", err, note, tail(log.String(), 40))
	}
	return harvest(src, tree, target)
}

// buildErrorLog names the file that keeps the log of a distribution's
// failed build for a target.
const buildErrorLog = "build-error.log"

// copyDistSource copies a distribution tree for building, without what
// earlier builds left in it.
func copyDistSource(tree, dst string) error {
	entries, err := os.ReadDir(tree)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(dst, 0o755); err != nil {
		return err
	}
	for _, e := range entries {
		if e.Name() == filepath.Base(mod.BuildDir(tree)) {
			continue
		}
		from, to := filepath.Join(tree, e.Name()), filepath.Join(dst, e.Name())
		if e.IsDir() {
			err = copyTree(from, to)
		} else {
			err = copyFileMode(from, to)
		}
		if err != nil {
			return err
		}
	}
	return nil
}

// copyFileMode copies one file, keeping its mode.
func copyFileMode(src, dst string) error {
	info, err := os.Stat(src)
	if err != nil {
		return err
	}
	data, err := os.ReadFile(src)
	if err != nil {
		return err
	}
	return os.WriteFile(dst, data, info.Mode().Perm())
}

// tail returns the last n lines of s.
func tail(s string, n int) string {
	lines := strings.Split(strings.TrimRight(s, "\n"), "\n")
	if len(lines) > n {
		lines = lines[len(lines)-n:]
	}
	return strings.Join(lines, "\n")
}

// harvest moves what the build produced (blib/) into the distribution's
// tree and marks it built.
func harvest(src, tree, target string) error {
	blib := filepath.Join(src, "blib")
	build := mod.BuildDir(tree)
	if err := os.MkdirAll(build, 0o755); err != nil {
		return err
	}
	for from, to := range map[string]string{"lib": "lib", "script": "bin", "bin": "bin"} {
		if dirExists(filepath.Join(blib, from)) {
			if err := copyTree(filepath.Join(blib, from), filepath.Join(build, to)); err != nil {
				return err
			}
		}
	}
	if mymeta := filepath.Join(src, "MYMETA.json"); fileExists(mymeta) {
		if err := copyFileMode(mymeta, filepath.Join(build, "MYMETA.json")); err != nil {
			return err
		}
	}

	// Native modules: blib/arch/auto/<Module/Path>/<Name>.{a,so,ldflags}.
	auto := filepath.Join(blib, "arch", "auto")
	native := false
	err := filepath.WalkDir(auto, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			if errors.Is(err, fs.ErrNotExist) {
				return fs.SkipAll
			}
			return err
		}
		if d.IsDir() {
			return nil
		}
		rel, err := filepath.Rel(auto, path)
		if err != nil {
			return err
		}
		switch filepath.Ext(path) {
		case ".so":
			if !fileExists(strings.TrimSuffix(path, ".so") + ".a") {
				return fmt.Errorf("%s was not linked by gperl (the distribution overrides the linker), so it cannot be linked statically", rel)
			}
			native = true
		case ".a", ".ldflags":
		case ".bundle", ".dylib", ".dll":
			return fmt.Errorf("%s was not linked by gperl (the distribution overrides the linker)", rel)
		default:
			return nil
		}
		dst := filepath.Join(mod.TargetDir(tree, target), "auto", rel)
		if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
			return err
		}
		return copyFileMode(path, dst)
	})
	if err != nil {
		return err
	}
	if !native {
		return mod.MarkBuilt(tree, "")
	}
	if err := writeBridge(tree); err != nil {
		return err
	}
	return mod.MarkBuilt(tree, target)
}

// nativeModule is one native module of a distribution on one target.
type nativeModule struct {
	target  string // "darwin_arm64"
	archive string // slash path below the target directory: auto/Cpanel/JSON/XS/XS.a
	libs    []string
}

// perlPackage returns the module's Perl package ("Cpanel::JSON::XS").
func (n nativeModule) perlPackage() string {
	dir := filepath.ToSlash(filepath.Dir(strings.TrimPrefix(n.archive, "auto/")))
	return strings.ReplaceAll(dir, "/", "::")
}

// nativeModules lists the native modules built into the tree, for every
// target present.
func nativeModules(tree string) ([]nativeModule, error) {
	abi := mod.ABIDir(tree)
	targets, err := os.ReadDir(abi)
	if errors.Is(err, fs.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var out []nativeModule
	for _, t := range targets {
		if !t.IsDir() {
			continue
		}
		root := filepath.Join(abi, t.Name())
		err := filepath.WalkDir(filepath.Join(root, "auto"), func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				if errors.Is(err, fs.ErrNotExist) {
					return fs.SkipAll
				}
				return err
			}
			if d.IsDir() || filepath.Ext(path) != ".a" {
				return nil
			}
			rel, err := filepath.Rel(root, path)
			if err != nil {
				return err
			}
			n := nativeModule{target: t.Name(), archive: filepath.ToSlash(rel)}
			flags, err := os.ReadFile(strings.TrimSuffix(path, ".a") + ".ldflags")
			if err != nil && !errors.Is(err, fs.ErrNotExist) {
				return err
			}
			n.libs = strings.Fields(string(flags))
			out = append(out, n)
			return nil
		})
		if err != nil {
			return nil, err
		}
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].target != out[j].target {
			return out[i].target < out[j].target
		}
		return out[i].archive < out[j].archive
	})
	return out, nil
}

// distImportPath returns the Go import path of the tree's native package.
func distImportPath(tree string) string {
	name, _, _ := strings.Cut(filepath.Base(tree), "@")
	return cpanModulePath + "/" + name
}

// writeBridge writes the Go package that links the tree's native modules:
// a go.mod and an xs.go with one #cgo LDFLAGS line per target and module,
// exporting each module's entry table.
func writeBridge(tree string) error {
	modules, err := nativeModules(tree)
	if err != nil {
		return err
	}
	var b strings.Builder
	b.WriteString("// Code generated by gperl. DO NOT EDIT.\n\npackage dist\n\n/*\n")
	packages := map[string]bool{}
	for _, n := range modules {
		fmt.Fprintf(&b, "#cgo %s LDFLAGS: ${SRCDIR}/%s/%s", strings.Join(targetTags(n.target), ","), n.target, n.archive)
		for i := 0; i < len(n.libs); i++ {
			// cgo accepts a fixed set of linker flags; the ones naming
			// libraries are all a module's .ldflags carries that it needs.
			switch lib := n.libs[i]; {
			case lib == "-framework" && i+1 < len(n.libs):
				b.WriteString(" -framework " + n.libs[i+1])
				i++
			case strings.HasPrefix(lib, "-l"), strings.HasPrefix(lib, "-L"), lib == "-pthread":
				b.WriteString(" " + lib)
			}
		}
		b.WriteString("\n")
		packages[n.perlPackage()] = true
	}
	names := make([]string, 0, len(packages))
	for name := range packages {
		names = append(names, name)
	}
	sort.Strings(names)
	for _, name := range names {
		fmt.Fprintf(&b, "extern const void *const goperl_entry_%s[12];\n", strings.ReplaceAll(name, "::", "__"))
	}
	b.WriteString("*/\nimport \"C\"\n\nimport \"unsafe\"\n\n")
	b.WriteString("// Entries maps each Perl package of the distribution to its entry table.\n")
	b.WriteString("var Entries = map[string]unsafe.Pointer{\n")
	for _, name := range names {
		fmt.Fprintf(&b, "\t%q: unsafe.Pointer(&C.goperl_entry_%s),\n", name, strings.ReplaceAll(name, "::", "__"))
	}
	b.WriteString("}\n")
	abi := mod.ABIDir(tree)
	gomod := "module " + distImportPath(tree) + "\n\ngo 1.25.0\n"
	if err := os.WriteFile(filepath.Join(abi, "go.mod"), []byte(gomod), 0o644); err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(abi, "xs.go"), []byte(b.String()), 0o644)
}
