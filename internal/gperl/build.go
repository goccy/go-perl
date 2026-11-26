package gperl

// gperl build: produce a self-contained Go binary that embeds the script,
// its CPAN dependencies, and the interpreter. The generated program
// behaves like the perl command (host filesystem, stdio, environment): at
// startup it unpacks the embedded app tree into a private temp directory,
// runs the script from there, and cleans up on exit.

import (
	"archive/zip"
	"bytes"
	"context"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"runtime/debug"
	"strings"
	"text/template"

	"github.com/goccy/go-perl/internal/mod"
)

// Build compiles script into a standalone binary at out (default: the
// script's basename without its extension, in the current directory). The
// embedded app tree contains the script itself and the project's lib/ (when
// present); the project's CPAN dependencies are linked in.
//
// The generated module requires the go-perl version gperl was built from.
// The GPERL_DEV_REPLACE environment variable ("module=path,module=path")
// adds replace directives to it — the way to build against an unreleased
// go-perl/perlwasm2go checkout.
func Build(script, out string) error { return buildBinary(script, out, nil) }

// buildBinary builds script into a standalone binary: a program that runs
// it, or — when psgiDefaults is non-nil — a server for the PSGI application
// it is, with those settings as defaults, or — when script is "" — the test
// binary of the project in the current directory (default out:
// <directory>.test). A server and a test binary carry the whole project
// directory (an application reads templates and static files next to
// itself, a test its data under t/); a program carries the script and lib/.
func buildBinary(script, out string, psgiDefaults []string) (err error) {
	config, err := readBuildConfig()
	if err != nil {
		return err
	}
	test := script == ""
	var projectDir string
	if test {
		if projectDir, err = os.Getwd(); err != nil {
			return err
		}
		if out == "" {
			out = filepath.Base(projectDir) + ".test"
		}
	} else {
		if script, err = filepath.Abs(script); err != nil {
			return err
		}
		if _, err := os.Stat(script); err != nil {
			return err
		}
		projectDir = filepath.Dir(script)
		if out == "" {
			out = strings.TrimSuffix(filepath.Base(script), filepath.Ext(script))
		}
	}
	if out, err = filepath.Abs(out); err != nil {
		return err
	}

	work, err := os.MkdirTemp("", "gperl-build-")
	if err != nil {
		return err
	}
	defer func() { err = errors.Join(err, os.RemoveAll(work)) }()
	// A workspace names its modules by their real paths, and the temp
	// directory is reached through a symlink on some systems.
	if work, err = filepath.EvalSymlinks(work); err != nil {
		return err
	}

	scriptName := filepath.Base(script)
	appZip, incDirs, err := buildAppZip(projectDir, script, psgiDefaults != nil || test, out)
	if err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(work, "app.zip"), appZip, 0o644); err != nil {
		return err
	}
	var mainSrc bytes.Buffer
	if err := mainTemplate.Execute(&mainSrc, map[string]any{
		"ScriptPath":   "app/" + scriptName,
		"Inc":          incDirs,
		"PSGI":         psgiDefaults != nil,
		"PSGIDefaults": psgiDefaults,
		"Test":         test,
	}); err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(work, "main.go"), mainSrc.Bytes(), 0o644); err != nil {
		return err
	}
	info, _ := debug.ReadBuildInfo() // nil without build information, which goModSrc reports
	gomod, err := goModSrc(info, os.Getenv("GPERL_DEV_REPLACE"))
	if err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(work, "go.mod"), gomod, 0o644); err != nil {
		return err
	}

	// The dependencies enter through a workspace: the generated link set
	// stands in for the go-perl/cpan package the interpreter imports.
	gowork := "off"
	target := config.target()
	if declaresDeps(projectDir) {
		target = nativeTarget(target)
	}
	native := false
	err = withDeps(func(m *mod.Manager) error {
		deps, err := loadDeps(context.Background(), m, projectDir, target)
		if err != nil || len(deps.lib) == 0 {
			return err
		}
		native = len(deps.nativeDirs) > 0
		links, err := generateLinkSet(m, deps)
		if err != nil {
			return err
		}
		gowork = filepath.Join(work, "go.work")
		return links.writeGoWork(gowork, work)
	})
	if err != nil {
		return err
	}
	goBin, err := goCommand()
	if err != nil {
		return err
	}
	ccWords, arWords, tcErr := cToolchain(target)
	// cgo links native modules in (GPERLNATIVE=link). Without native
	// modules a build for this machine has cgo when it has a C compiler, as
	// go build does; a build for another platform needs no C compiler.
	cgo := config.link && (native || (config.target() == hostPlatform() && tcErr == nil))
	if cgo && tcErr != nil {
		return fmt.Errorf("link the native modules in: %w", tcErr)
	}
	env := config.goEnv(cgo)
	// A cgo build compiles and links with the same C toolchain the native
	// modules were built with.
	if tcErr == nil {
		env = append(env, "CC="+strings.Join(ccWords, " "), "AR="+strings.Join(arWords, " "))
	}
	for _, step := range []struct {
		gowork string
		argv   []string
	}{
		{"off", []string{"go", "mod", "tidy"}},
		{gowork, append(buildFlags(target), "-o", out, ".")},
	} {
		cmd := exec.Command(goBin, step.argv[1:]...)
		cmd.Dir = work
		cmd.Env = append(append([]string{}, env...), "GOWORK="+step.gowork)
		cmd.Stdout = os.Stderr
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("%s: %w", strings.Join(step.argv, " "), err)
		}
	}
	fmt.Fprintf(os.Stderr, "gperl: built %s\n", out)
	return nil
}

// buildAppZip assembles the embedded app tree: the script at /app/<name>
// and the project's lib/ — or, with whole, everything in the project
// directory but hidden entries, the dependency declarations, and out.
// Returns the zip bytes and the guest @INC additions, most specific first.
func buildAppZip(projectDir, script string, whole bool, out string) ([]byte, []string, error) {
	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)
	// add stores host at name with its mode: a file (what a symbolic link
	// to one names), or a directory as an entry of its own, so that an empty
	// one stays. No time goes with it: the same tree makes the same binary.
	add := func(name, host string, dir bool) error {
		info, err := os.Stat(host)
		if err != nil {
			return err
		}
		hdr := &zip.FileHeader{Name: name, Method: zip.Deflate}
		hdr.SetMode(info.Mode())
		var data []byte
		switch {
		case info.IsDir() != dir:
			return fmt.Errorf("%s: a symbolic link to a directory cannot be embedded", host)
		case dir:
			hdr.Name, hdr.Method = name+"/", zip.Store
		case !info.Mode().IsRegular():
			return fmt.Errorf("%s: not a regular file or a directory", host)
		default:
			if data, err = os.ReadFile(host); err != nil {
				return err
			}
		}
		w, err := zw.CreateHeader(hdr)
		if err != nil {
			return err
		}
		_, err = w.Write(data)
		return err
	}
	addFile := func(name, host string) error { return add(name, host, false) }
	addTree := func(name, host string) error {
		return filepath.WalkDir(host, func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			rel, err := filepath.Rel(host, path)
			if err != nil {
				return err
			}
			entry := name
			if rel != "." {
				entry += "/" + filepath.ToSlash(rel)
			}
			return add(entry, path, d.IsDir())
		})
	}

	inc := []string{}
	if whole {
		// Everything but what is not the application's: hidden entries,
		// the dependency declarations, and the binary being built.
		entries, err := os.ReadDir(projectDir)
		if err != nil {
			return nil, nil, err
		}
		for _, e := range entries {
			host := filepath.Join(projectDir, e.Name())
			switch {
			case strings.HasPrefix(e.Name(), "."), e.Name() == mod.CPANfile, e.Name() == mod.SnapshotFile, host == out:
				continue
			case e.IsDir():
				err = addTree("app/"+e.Name(), host)
			default:
				err = addFile("app/"+e.Name(), host)
			}
			if err != nil {
				return nil, nil, err
			}
		}
	} else if err := addFile("app/"+filepath.Base(script), script); err != nil {
		return nil, nil, err
	}
	if st, err := os.Stat(filepath.Join(projectDir, "lib")); err == nil && st.IsDir() {
		if !whole {
			if err := addTree("app/lib", filepath.Join(projectDir, "lib")); err != nil {
				return nil, nil, err
			}
		}
		inc = append(inc, "app/lib")
	}
	inc = append(inc, "app")
	if err := zw.Close(); err != nil {
		return nil, nil, err
	}
	return buf.Bytes(), inc, nil
}

// generatedModulePath is the module path of a program gperl generates.
const generatedModulePath = "github.com/goccy/go-perl/internal/gperl/app"

// goPerlModule is the module path of go-perl.
const goPerlModule = "github.com/goccy/go-perl"

// goModSrc renders the generated program's go.mod. It requires the go-perl
// version in info, gperl's build information; the rest resolves through
// `go mod tidy`. devReplace ("module=path,module=path") adds replace
// directives for unreleased checkouts.
func goModSrc(info *debug.BuildInfo, devReplace string) ([]byte, error) {
	var replaces strings.Builder
	replaced := false
	for _, pair := range strings.Split(devReplace, ",") {
		mod, path, ok := strings.Cut(strings.TrimSpace(pair), "=")
		if !ok {
			continue
		}
		fmt.Fprintf(&replaces, "replace %s => %s\n", mod, path)
		replaced = replaced || strings.TrimSpace(mod) == goPerlModule
	}
	version, err := goPerlVersion(info)
	if err != nil {
		if !replaced {
			return nil, err
		}
		// A replaced module comes from its checkout whatever version is
		// required.
		version = "v0.0.0"
	}
	var b strings.Builder
	// The generated program imports this package, which is internal to
	// go-perl: its module path sits below go-perl's so that it may.
	b.WriteString("module " + generatedModulePath + "\n\ngo 1.25.0\n\nrequire " + goPerlModule + " " + version + "\n")
	if replaces.Len() > 0 {
		b.WriteString("\n" + replaces.String())
	}
	return []byte(b.String()), nil
}

// goPerlVersion returns the version of go-perl in info: its main module,
// the module of gperl's main package, is go-perl.
func goPerlVersion(info *debug.BuildInfo) (string, error) {
	if info == nil {
		return "", fmt.Errorf("this binary has no build information: the %s version to build with is unknown", goPerlModule)
	}
	m := info.Main
	// A build outside a tagged commit has no version or "(devel)", one with
	// uncommitted changes carries build metadata ("+dirty"), and a replaced
	// module's code is not its version's: none is a version to require.
	if m.Version == "" || m.Version == "(devel)" || strings.Contains(m.Version, "+") || m.Replace != nil {
		return "", fmt.Errorf("this binary is built from %s %q, not a release: install a released gperl, or set GPERL_DEV_REPLACE=%s=<checkout> to build with a local checkout", goPerlModule, m.Version, goPerlModule)
	}
	return m.Version, nil
}

var mainTemplate = template.Must(template.New("main").Parse(`// Code generated by gperl build. DO NOT EDIT.

// The program is the process and has its descriptor numbers: the Go runtime
// must not hold a file of its own under the lowest of them (see
// cmd/gperl).
//
//go:debug containermaxprocs=0
package main

import (
	"archive/zip"
	"bytes"
{{- if not (or .PSGI .Test)}}
	"context"
{{- end}}
	_ "embed"
{{- if not (or .PSGI .Test)}}
	"errors"
{{- end}}
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
{{if not (or .PSGI .Test)}}
	perl "github.com/goccy/go-perl"
{{- end}}
	"github.com/goccy/go-perl/internal/gperl"
)

//go:embed app.zip
var appZip []byte

const scriptRel = {{printf "%q" .ScriptPath}}

var incRel = []string{ {{range .Inc}}{{printf "%q" .}}, {{end}} }

func main() {
	// The embedded app tree unpacks into a private temp directory; the
	// script itself runs with the HOST filesystem, like the perl command.
	appDir, err := os.MkdirTemp("", "gperl-app-")
	if err != nil {
		fatal(err)
	}
	cleanup = func() { os.RemoveAll(appDir) }
	defer cleanup()
	zr, err := zip.NewReader(bytes.NewReader(appZip), int64(len(appZip)))
	if err != nil {
		fatal(err)
	}
	for _, f := range zr.File {
		target := filepath.Join(appDir, filepath.FromSlash(f.Name))
		if !strings.HasPrefix(target, appDir+string(os.PathSeparator)) {
			fatal(fmt.Errorf("embedded path escapes app dir: %s", f.Name))
		}
		if f.FileInfo().IsDir() {
			if err := os.MkdirAll(target, 0o755); err != nil {
				fatal(err)
			}
			continue
		}
		if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
			fatal(err)
		}
		rc, err := f.Open()
		if err != nil {
			fatal(err)
		}
		data, err := io.ReadAll(rc)
		rc.Close()
		if err != nil {
			fatal(err)
		}
		if err := os.WriteFile(target, data, f.Mode().Perm()); err != nil {
			fatal(err)
		}
	}
{{- if .Test}}
	status := gperl.RunTestBinary(filepath.Join(appDir, "app"), os.Args[1:])
	cleanup()
	os.Exit(status)
}
{{- else}}
	scriptPath := filepath.Join(appDir, filepath.FromSlash(scriptRel))
	incDirs := make([]string, 0, len(incRel))
	for _, rel := range incRel {
		incDirs = append(incDirs, filepath.Join(appDir, filepath.FromSlash(rel)))
	}
{{- if .PSGI}}
	status := gperl.PSGIMain(scriptPath, incDirs, []string{ {{range .PSGIDefaults}}{{printf "%q" .}}, {{end}} }, os.Args[1:])
	cleanup()
	os.Exit(status)
}
{{- else}}
	cfg, err := gperl.HostConfig()
	if err != nil {
		fatal(err)
	}
	cfg.Stdin = os.Stdin
	cfg.Stdout = os.Stdout
	cfg.Stderr = os.Stderr
	cfg.Env = os.Environ()
	p, err := perl.New(cfg)
	if err != nil {
		fatal(err)
	}
	status, err := p.RunProgram(context.Background(), perl.Program{
		File: scriptPath,
		Inc:  incDirs,
		Args: os.Args[1:],
	})
	// The status is final: the run flushed PerlIO and went through global
	// destruction. Close gives back what the host holds for the instance.
	p.Close()
	var killed *perl.SignalError
	if errors.As(err, &killed) {
		// A signal ended the program: report it the way a shell reports a
		// signal death.
		cleanup()
		os.Exit(128 + killed.Signal)
	}
	if err != nil {
		fatal(err)
	}
	cleanup()
	os.Exit(status)
}
{{- end}}
{{- end}}

// cleanup removes the unpacked app dir; os.Exit skips defers, so every
// exit path calls it explicitly.
var cleanup = func() {}

func fatal(err error) {
	fmt.Fprintln(os.Stderr, err)
	cleanup()
	os.Exit(1)
}
`))
