package gperl

// gperl xs build: compile an XS distribution against the native XS SDK and
// place the resulting shared library where gperl run / built binaries load
// it from.
//
// The build rides the dist's OWN build system — `perl Makefile.PL && make`
// (ExtUtils::MakeMaker) or `perl Build.PL && ./Build` (Module::Build) —
// exactly the way the CPAN toolchain normally builds XS, with two twists:
// every perl in the pipeline is the EMBEDDED interpreter (Makefile.PL,
// Build.PL, and cpanm run IN-PROCESS; the perl children that make rules
// and $^X re-invocations spawn exec a shim that re-enters this executable
// as `gperl run`, with a preloaded %Config overlay describing the host C
// toolchain), and the compiler's perl-header search path is redirected to
// the SDK (materialized from the copy embedded in the xs package).
// Generated headers, xsubpp runs, extra C sources, and typemaps all keep
// working. The only command the pipeline execs is make (cc runs under
// it) — no perl needs to be installed. The RUNTIME requirement stays
// zero: gperl run only dlopens the prebuilt artifacts.
//
// Layout produced under the project directory:
//
//	local/xs/<goos>_<goarch>/<Module-Name>.so   the native module(s)
//	local/lib/perl5/...                         the dist's pure-Perl half
//	                                            (blib/lib, same tree cpanm
//	                                            would install into)

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"sort"
	"strings"

	perl "github.com/goccy/go-perl"
	"github.com/goccy/go-perl/xs"
)

// xsArchTag names the per-architecture native-module directory
// (local/xs/<tag>). The tag follows the RUNNING binary — its dlopen must
// match — which is also the architecture gperl build produces by default.
func xsArchTag() string { return xs.ArchTag() }

// xsDir returns the project's native-module directory for this
// architecture (creating nothing).
func xsDir(projectDir string) string {
	return filepath.Join(projectDir, "local", "xs", xsArchTag())
}

// XSBuild builds each XS distribution (a source directory or a .tar.gz)
// against the SDK and installs the artifacts into projectDir/local. It
// returns the Perl module names that gained a native library.
func XSBuild(projectDir string, dists []string) ([]string, error) {
	projectDir, err := filepath.Abs(projectDir)
	if err != nil {
		return nil, err
	}
	sdkDir := filepath.Join(projectDir, "local", "xs", ".sdk")
	if err := xs.WriteSDK(sdkDir); err != nil {
		return nil, fmt.Errorf("materialize SDK headers: %w", err)
	}
	var modules []string
	for _, dist := range dists {
		mods, err := xsBuildOne(projectDir, sdkDir, dist)
		if err != nil {
			return modules, fmt.Errorf("%s: %w", dist, err)
		}
		modules = append(modules, mods...)
	}
	return modules, nil
}

func xsBuildOne(projectDir, sdkDir, dist string) ([]string, error) {
	work, err := os.MkdirTemp("", "gperl-xs-")
	if err != nil {
		return nil, err
	}
	defer os.RemoveAll(work)

	// Build in a scratch copy: the dist's build system writes Makefile,
	// objects, and blib into its tree, and the source should stay pristine.
	srcDir, err := xsStageSource(dist, work)
	if err != nil {
		return nil, err
	}

	// The build is DRIVEN by the embedded interpreter: this process runs
	// Makefile.PL/Build.PL itself, and the shim (for the children make or
	// $^X spawns) re-enters this executable. Every run preloads a %Config
	// overlay describing the host toolchain (the interpreter's own
	// %Config is guest-true wasm32). No perl needs to be installed.
	shim, err := xsWritePerlShim(work)
	if err != nil {
		return nil, err
	}
	// The compiler's CORE include directory is the SDK, via a fake
	// archlib whose CORE symlinks to it (both ExtUtils::MakeMaker and
	// Module::Build spell it $Config{archlibexp}/CORE).
	fakeArch := filepath.Join(work, "fakearchlib")
	if err := os.MkdirAll(fakeArch, 0o755); err != nil {
		return nil, err
	}
	if err := os.Symlink(sdkDir, filepath.Join(fakeArch, "CORE")); err != nil {
		return nil, err
	}
	stdlibDir, err := perl.ExtractStdlib()
	if err != nil {
		return nil, err
	}
	overlay, err := xsWriteConfigOverlay(work, sdkDir, stdlibDir, fakeArch)
	if err != nil {
		return nil, err
	}
	env := append(os.Environ(), xsArchEnv()...)
	env = append(env,
		"GPERL_PERL_PRELOAD="+overlay,
		"GPERL_PERL_EXE="+shim,
		"PERL="+shim,
		"PATH="+filepath.Dir(shim)+string(os.PathListSeparator)+os.Getenv("PATH"),
	)
	// Resolve the dist's OWN build-time dependencies (configure_requires /
	// build_requires — Module::Build, Devel::PPPort, ...) into a scratch
	// lib for the build. Runtime deps of the PROJECT stay a cpanfile
	// concern.
	buildLib := filepath.Join(work, "buildlib")
	if err := xsInstallBuildDeps(srcDir, buildLib); err != nil {
		return nil, err
	}
	if dirExists(filepath.Join(buildLib, "lib", "perl5")) {
		env = append(env, "PERL5LIB="+filepath.Join(buildLib, "lib", "perl5"))
	}
	switch {
	case fileExists(filepath.Join(srcDir, "Makefile.PL")):
		if err := runPerlInProcess(srcDir, env, os.Stdout, "Makefile.PL",
			"PERL="+shim, "FULLPERL="+shim); err != nil {
			return nil, err
		}
		if err := xsRedirectMakefile(filepath.Join(srcDir, "Makefile"), sdkDir); err != nil {
			return nil, err
		}
		if err := runIn(srcDir, env, "make"); err != nil {
			return nil, err
		}
	case fileExists(filepath.Join(srcDir, "Build.PL")):
		if err := runPerlInProcess(srcDir, env, os.Stdout, "Build.PL"); err != nil {
			return nil, err
		}
		if err := runPerlInProcess(srcDir, env, os.Stdout, "Build"); err != nil {
			return nil, err
		}
	default:
		return nil, fmt.Errorf("no Makefile.PL or Build.PL found (not an XS distribution?)")
	}

	return xsInstallBlib(projectDir, filepath.Join(srcDir, "blib"))
}

// xsStageSource copies a dist directory (or extracts a .tar.gz) into the
// scratch dir and returns the dist root.
func xsStageSource(dist, work string) (string, error) {
	st, err := os.Stat(dist)
	if err != nil {
		return "", err
	}
	dst := filepath.Join(work, "src")
	if st.IsDir() {
		if err := copyTree(dist, dst); err != nil {
			return "", err
		}
		if err := xsNeutralizePPPort(dst); err != nil {
			return "", err
		}
		return dst, nil
	}
	if err := extractTarGz(dist, dst); err != nil {
		return "", err
	}
	if err := xsNeutralizePPPort(dst); err != nil {
		return "", err
	}
	// A CPAN tarball unpacks into a single Dist-Name-1.23 directory.
	entries, err := os.ReadDir(dst)
	if err != nil {
		return "", err
	}
	if len(entries) == 1 && entries[0].IsDir() {
		return filepath.Join(dst, entries[0].Name()), nil
	}
	return dst, nil
}

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
		stub := "/* ppport.h neutralized by gperl xs build: the goperl XS SDK\n" +
			" * presents a current perl API, so the portability backfills\n" +
			" * must not activate. */\n" +
			"#ifndef _P_P_PORTABILITY_H_\n" +
			"#define _P_P_PORTABILITY_H_\n" +
			"#endif\n"
		fmt.Fprintf(os.Stderr, "gperl: neutralized %s\n", path)
		return os.WriteFile(path, []byte(stub), 0o644)
	})
}

// xsArchEnv yields the environment that makes the host toolchain target
// the running binary's architecture (Apple's perl honors ARCHFLAGS for
// extension builds).
func xsArchEnv() []string {
	if runtime.GOOS != "darwin" {
		return nil
	}
	arch := map[string]string{"amd64": "x86_64", "arm64": "arm64"}[runtime.GOARCH]
	if arch == "" {
		return nil
	}
	return []string{"ARCHFLAGS=-arch " + arch}
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

// xsInstallBlib copies the built native libraries into local/xs/<arch> and
// the pure-Perl half into local/lib/perl5, returning the module names.
func xsInstallBlib(projectDir, blib string) ([]string, error) {
	autoRoot := filepath.Join(blib, "arch", "auto")
	var modules []string
	err := filepath.WalkDir(autoRoot, func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return nil //nolint:nilerr // a dist without arch objects is legal
		}
		switch filepath.Ext(path) {
		case ".so", ".bundle", ".dylib", ".dll":
		default:
			return nil
		}
		rel, err := filepath.Rel(autoRoot, path)
		if err != nil {
			return err
		}
		// auto/Devel/NYTProf/NYTProf.bundle -> Devel::NYTProf
		module := strings.ReplaceAll(filepath.Dir(rel), string(filepath.Separator), "::")
		dstDir := xsDir(projectDir)
		if err := os.MkdirAll(dstDir, 0o755); err != nil {
			return err
		}
		dst := filepath.Join(dstDir, strings.ReplaceAll(module, "::", "-")+".so")
		if err := copyFile(path, dst); err != nil {
			return err
		}
		modules = append(modules, module)
		return nil
	})
	if err != nil {
		return nil, err
	}
	if len(modules) == 0 {
		return nil, fmt.Errorf("build produced no native library under blib/arch/auto")
	}
	// The .pm half goes where cpanm would put it, so @INC already covers it.
	libDst := filepath.Join(projectDir, "local", "lib", "perl5")
	if dirExists(filepath.Join(blib, "lib")) {
		if err := copyTree(filepath.Join(blib, "lib"), libDst); err != nil {
			return nil, err
		}
	}
	return modules, nil
}

// xsInstallBuildDeps runs `cpanm --installdeps` for the dist into lib on
// the embedded interpreter. Run from a neutral directory: on a
// case-insensitive filesystem cpanm resolves bare module names against
// look-alike local paths.
func xsInstallBuildDeps(srcDir, lib string) error {
	neutral, err := os.MkdirTemp("", "gperl-xsdeps-")
	if err != nil {
		return err
	}
	defer os.RemoveAll(neutral)
	if err := runCpanmArgs(neutral, neutral, []string{"-L", lib, "--notest", "--installdeps", srcDir}); err != nil {
		return fmt.Errorf("resolve build dependencies (cpanm --installdeps): %w", err)
	}
	return nil
}

// hostToolchain is the %Config overlay describing the HOST C toolchain:
// the embedded interpreter's own %Config is guest-true (wasm32), and the
// build must emit host-native compile/link rules. These are the platform
// conventions a native perl build would have configured.
func hostToolchain() map[string]string {
	common := map[string]string{
		"cc": "cc", "ld": "cc", "ar": "ar", "full_ar": "ar",
		"ranlib": "ranlib", "make": "make",
		"ccflags":  "-fno-strict-aliasing -fstack-protector-strong -DPERL_USE_SAFE_PUTENV",
		"optimize": "-O2", "ldflags": "", "cccdlflags": "-fPIC",
		"ccdlflags": "", "obj_ext": ".o", "exe_ext": "", "lib_ext": ".a",
		"usedl": "define", "dlsrc": "dl_dlopen.xs", "useshrplib": "false",
	}
	switch runtime.GOOS {
	case "darwin":
		common["osname"] = "darwin"
		common["so"] = "dylib"
		common["dlext"] = "bundle"
		common["lddlflags"] = "-bundle -undefined dynamic_lookup -fstack-protector-strong"
		common["cccdlflags"] = "" // PIC is the darwin default
	default: // linux and friends
		common["osname"] = runtime.GOOS
		common["so"] = "so"
		common["dlext"] = "so"
		common["lddlflags"] = "-shared"
	}
	return common
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

// xsWriteConfigOverlay writes the preload every build-driving perl run
// executes first: it replaces the guest-true toolchain/path keys of
// %Config with the host ones, so ExtUtils::MakeMaker and Module::Build
// generate a build for THIS machine while everything else about %Config
// (version, ivsize, ...) stays guest-true.
func xsWriteConfigOverlay(work, sdkDir, stdlibDir, fakeArch string) (string, error) {
	tc := hostToolchain()
	if sdkDir != "" {
		tc["ccflags"] = tc["ccflags"] + " -I" + sdkDir
	}
	// paths: modules and xsubpp resolve from the extracted stdlib; the
	// compiler's CORE directory is the SDK (via the fake archlib).
	tc["privlibexp"] = stdlibDir
	tc["privlib"] = stdlibDir
	tc["archlibexp"] = fakeArch
	tc["archlib"] = fakeArch
	tc["sitelibexp"] = ""
	tc["sitearchexp"] = ""
	tc["vendorlibexp"] = ""
	tc["vendorarchexp"] = ""
	tc["installprivlib"] = stdlibDir
	tc["installarchlib"] = fakeArch

	var b strings.Builder
	b.WriteString("use Config ();\n")
	b.WriteString("my %over = (\n")
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
	// %Config is a tied read-only view; replace it with a writable copy
	// carrying the overlay.
	b.WriteString("my %full = %Config::Config;\n")
	b.WriteString("untie %Config::Config;\n")
	b.WriteString("%Config::Config = (%full, %over);\n")
	// Build tools branch on $^O, not just $Config{osname}: the build must
	// look like the host it targets.
	b.WriteString("$^O = '" + tc["osname"] + "';\n1;\n")

	path := filepath.Join(work, "config-overlay.pl")
	if err := os.WriteFile(path, []byte(b.String()), 0o644); err != nil {
		return "", err
	}
	return path, nil
}

func runIn(dir string, env []string, name string, args ...string) error {
	cmd := exec.Command(name, args...)
	cmd.Dir = dir
	cmd.Env = env
	cmd.Stdout = os.Stderr
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("%s %s: %w", name, strings.Join(args, " "), err)
	}
	return nil
}

func fileExists(path string) bool {
	st, err := os.Stat(path)
	return err == nil && !st.IsDir()
}

func dirExists(path string) bool {
	st, err := os.Stat(path)
	return err == nil && st.IsDir()
}

func copyFile(src, dst string) error {
	data, err := os.ReadFile(src)
	if err != nil {
		return err
	}
	return os.WriteFile(dst, data, 0o755)
}

func copyTree(src, dst string) error {
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
				return copyTree(resolved, target)
			}
			path = resolved
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		info, err := d.Info()
		mode := fs.FileMode(0o644)
		if err == nil && info.Mode()&0o100 != 0 {
			mode = 0o755
		}
		return os.WriteFile(target, data, mode)
	})
}

func extractTarGz(archive, dst string) error {
	f, err := os.Open(archive)
	if err != nil {
		return err
	}
	defer f.Close()
	gz, err := gzip.NewReader(f)
	if err != nil {
		return fmt.Errorf("open %s: %w (only .tar.gz distributions are supported)", archive, err)
	}
	defer gz.Close()
	tr := tar.NewReader(gz)
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}
		if filepath.Clean(hdr.Name) == "." {
			continue
		}
		target := filepath.Join(dst, hdr.Name)
		// containment: a crafted entry ("..", "../x", absolute) must not
		// escape the extraction root
		if !strings.HasPrefix(target, filepath.Clean(dst)+string(os.PathSeparator)) {
			return fmt.Errorf("archive entry escapes extraction dir: %q", hdr.Name)
		}
		switch hdr.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(target, 0o755); err != nil {
				return err
			}
		case tar.TypeReg:
			if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
				return err
			}
			out, err := os.OpenFile(target, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, fs.FileMode(hdr.Mode)&0o777)
			if err != nil {
				return err
			}
			if _, err := io.Copy(out, tr); err != nil { //nolint:gosec // trusted local archive
				out.Close()
				return err
			}
			out.Close()
		}
	}
}
