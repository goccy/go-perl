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
	"strings"
	"text/template"

	"github.com/goccy/go-perl/internal/mod"
)

// Build compiles script into a standalone binary at out (default: the
// script's basename without its extension, in the current directory). The
// embedded app tree contains the script itself and the project's lib/ (when
// present); the project's CPAN dependencies are linked in.
//
// The GPERL_DEV_REPLACE environment variable ("module=path,module=path")
// adds replace directives to the generated module — the way to build against
// an unreleased go-perl/perlwasm2go checkout.
func Build(script, out string) error { return buildBinary(script, out, nil) }

// buildBinary builds script into a standalone binary: a program that runs
// it, or — when psgiDefaults is non-nil — a server for the PSGI application
// it is, with those settings as defaults. A server carries the whole
// project directory (an application reads templates and static files next
// to itself); a program carries the script and lib/.
func buildBinary(script, out string, psgiDefaults []string) (err error) {
	if script, err = filepath.Abs(script); err != nil {
		return err
	}
	if _, err := os.Stat(script); err != nil {
		return err
	}
	projectDir := filepath.Dir(script)
	if out == "" {
		out = strings.TrimSuffix(filepath.Base(script), filepath.Ext(script))
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
	appZip, incDirs, err := buildAppZip(projectDir, script, psgiDefaults != nil, out)
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
	}); err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(work, "main.go"), mainSrc.Bytes(), 0o644); err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(work, "go.mod"), goModSrc(), 0o644); err != nil {
		return err
	}

	// The dependencies enter through a workspace: the generated link set
	// stands in for the go-perl/cpan package the interpreter imports.
	gowork := "off"
	err = withDeps(func(m *mod.Manager) error {
		deps, err := loadDeps(context.Background(), m, projectDir, mod.BuildTarget())
		if err != nil || deps.lib == nil {
			return err
		}
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
	env := os.Environ()
	// A cgo build compiles and links with the same C toolchain the native
	// modules were built with.
	if ccWords, arWords, tcErr := cToolchain(mod.BuildTarget()); tcErr == nil {
		env = append(env, "CC="+strings.Join(ccWords, " "), "AR="+strings.Join(arWords, " "))
	}
	for _, step := range []struct {
		gowork string
		argv   []string
	}{
		{"off", []string{"go", "mod", "tidy"}},
		{gowork, []string{"go", "build", "-o", out, "."}},
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
	addFile := func(name, host string) error {
		data, err := os.ReadFile(host)
		if err != nil {
			return err
		}
		w, err := zw.Create(name)
		if err != nil {
			return err
		}
		_, err = w.Write(data)
		return err
	}
	addTree := func(name, host string) error {
		return filepath.WalkDir(host, func(path string, d fs.DirEntry, err error) error {
			if err != nil || d.IsDir() {
				return err
			}
			rel, rErr := filepath.Rel(host, path)
			if rErr != nil {
				return rErr
			}
			return addFile(name+"/"+filepath.ToSlash(rel), path)
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

// goModSrc renders the generated program's go.mod. Versions resolve through
// `go mod tidy`; GPERL_DEV_REPLACE injects replace directives for unreleased
// checkouts.
func goModSrc() []byte {
	var b strings.Builder
	// The generated program imports this package, which is internal to
	// go-perl: its module path sits below go-perl's so that it may.
	b.WriteString("module " + generatedModulePath + "\n\ngo 1.25.0\n\nrequire github.com/goccy/go-perl v0.0.0\n")
	if env := os.Getenv("GPERL_DEV_REPLACE"); env != "" {
		b.WriteString("\n")
		for _, pair := range strings.Split(env, ",") {
			mod, path, ok := strings.Cut(strings.TrimSpace(pair), "=")
			if !ok {
				continue
			}
			fmt.Fprintf(&b, "replace %s => %s\n", mod, path)
		}
	}
	return []byte(b.String())
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
{{- if not .PSGI}}
	"context"
{{- end}}
	_ "embed"
{{- if not .PSGI}}
	"errors"
{{- end}}
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
{{if not .PSGI}}
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
		if f.FileInfo().IsDir() {
			continue
		}
		target := filepath.Join(appDir, filepath.FromSlash(f.Name))
		if !strings.HasPrefix(target, appDir+string(os.PathSeparator)) {
			fatal(fmt.Errorf("embedded path escapes app dir: %s", f.Name))
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
		if err := os.WriteFile(target, data, 0o644); err != nil {
			fatal(err)
		}
	}
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

// cleanup removes the unpacked app dir; os.Exit skips defers, so every
// exit path calls it explicitly.
var cleanup = func() {}

func fatal(err error) {
	fmt.Fprintln(os.Stderr, err)
	cleanup()
	os.Exit(1)
}
`))
