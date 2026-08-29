package gperl

// gperl xs build: compile an XS distribution against the native XS SDK and
// place the resulting shared library where gperl run / built binaries load
// it from.
//
// The build rides the dist's OWN build system — `perl Makefile.PL && make`
// (ExtUtils::MakeMaker) or `perl Build.PL && ./Build` (Module::Build) — run
// with the HOST perl, exactly the way the CPAN toolchain normally builds XS.
// Only the compiler's perl-header search path is redirected to the SDK
// (materialized from the copy embedded in the xsnative package), so
// generated headers, xsubpp runs, extra C sources, and typemaps all keep
// working. Build-time requirements: host perl, make, and a C compiler —
// the same set MakeMaker needs anywhere. The RUNTIME requirement stays
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
	"strings"

	perl "github.com/goccy/go-perl"
	"github.com/goccy/go-perl/xsnative"
)

// XSArchTag names the per-architecture native-module directory
// (local/xs/<tag>). The tag follows the RUNNING binary — its dlopen must
// match — which is also the architecture gperl build produces by default.
func XSArchTag() string { return runtime.GOOS + "_" + runtime.GOARCH }

// xsDir returns the project's native-module directory for this
// architecture (creating nothing).
func xsDir(projectDir string) string {
	return filepath.Join(projectDir, "local", "xs", XSArchTag())
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
	if err := xsnative.WriteSDK(sdkDir); err != nil {
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

	env := append(os.Environ(), xsArchEnv()...)
	// Resolve the dist's OWN build-time dependencies (configure_requires /
	// build_requires — Module::Build::XSUtil, Devel::PPPort, ...) into a
	// scratch lib for the host perl that drives the build. Runtime deps of
	// the PROJECT stay a cpanfile concern.
	buildLib := filepath.Join(work, "buildlib")
	if err := xsInstallBuildDeps(srcDir, buildLib); err != nil {
		return nil, err
	}
	if dirExists(filepath.Join(buildLib, "lib", "perl5")) {
		env = append(env, "PERL5LIB="+filepath.Join(buildLib, "lib", "perl5"))
	}
	switch {
	case fileExists(filepath.Join(srcDir, "Makefile.PL")):
		if err := runIn(srcDir, env, "perl", "Makefile.PL"); err != nil {
			return nil, err
		}
		if err := xsRedirectMakefile(filepath.Join(srcDir, "Makefile"), sdkDir); err != nil {
			return nil, err
		}
		if err := runIn(srcDir, env, "make"); err != nil {
			return nil, err
		}
	case fileExists(filepath.Join(srcDir, "Build.PL")):
		// Module::Build compiles with -I$Config{archlibexp}/CORE, so a
		// fake archlib whose CORE is the SDK redirects it. Apple's system
		// perl additionally patches ExtUtils::CBuilder to spell every
		// include dir as -iwithsysroot (prefixing the platform SDK root),
		// which no override can survive — the SDK therefore ALSO rides in
		// verbatim through ccflags.
		fakeArch := filepath.Join(work, "fakearchlib")
		if err := os.MkdirAll(fakeArch, 0o755); err != nil {
			return nil, err
		}
		if err := os.Symlink(sdkDir, filepath.Join(fakeArch, "CORE")); err != nil {
			return nil, err
		}
		ccflags, err := hostPerlConfig("ccflags")
		if err != nil {
			return nil, err
		}
		if err := runIn(srcDir, env, "perl", "Build.PL",
			"--config", "archlibexp="+fakeArch,
			"--config", "ccflags="+ccflags+" -I"+sdkDir); err != nil {
			return nil, err
		}
		if err := runIn(srcDir, env, "perl", "Build"); err != nil {
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
		return dst, nil
	}
	if err := extractTarGz(dist, dst); err != nil {
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

// LoadXS registers every native module under dir (a local/xs/<arch>
// directory) with the instance. Registration is cheap — each module's boot
// runs lazily when Perl code first `use`s it.
func LoadXS(p *perl.Perl, dir string) error {
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	for _, e := range entries {
		if e.IsDir() || filepath.Ext(e.Name()) != ".so" {
			continue
		}
		module := strings.ReplaceAll(strings.TrimSuffix(e.Name(), ".so"), "-", "::")
		if err := xsnative.Load(p, module, filepath.Join(dir, e.Name())); err != nil {
			return fmt.Errorf("load native module %s: %w", module, err)
		}
	}
	return nil
}

// xsInstallBuildDeps runs `cpanm --installdeps` for the dist into lib
// (cpanm bootstraps from cpanmin.us when not installed, matching gperl
// get). Run from a neutral directory: on a case-insensitive filesystem
// cpanm resolves bare module names against look-alike local paths.
func xsInstallBuildDeps(srcDir, lib string) error {
	neutral, err := os.MkdirTemp("", "gperl-xsdeps-")
	if err != nil {
		return err
	}
	defer os.RemoveAll(neutral)
	argv, err := cpanmArgv(neutral)
	if err != nil {
		return err
	}
	args := append(argv[1:], "-L", lib, "--notest", "--installdeps", srcDir)
	cmd := exec.Command(argv[0], args...)
	cmd.Dir = neutral
	cmd.Stdout = os.Stderr
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("resolve build dependencies (cpanm --installdeps): %w", err)
	}
	return nil
}

// hostPerlConfig reads one %Config value from the host perl.
func hostPerlConfig(key string) (string, error) {
	out, err := exec.Command("perl", "-MConfig", "-e", "print $Config{"+key+"}").Output()
	if err != nil {
		return "", fmt.Errorf("read host perl %%Config{%s}: %w", key, err)
	}
	return strings.TrimSpace(string(out)), nil
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
		name := filepath.Clean(hdr.Name)
		if name == "." || strings.HasPrefix(name, ".."+string(filepath.Separator)) || filepath.IsAbs(name) {
			continue
		}
		target := filepath.Join(dst, name)
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
