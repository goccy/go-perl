package gperl

// gperl build: produce a self-contained Go binary that embeds the script,
// its vendored pure-Perl modules, and the interpreter. The generated program
// behaves like the perl command (host filesystem, stdio, environment): at
// startup it unpacks the embedded app tree into a private temp directory,
// runs the script from there, and cleans up on exit.

import (
	"archive/zip"
	"bytes"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"text/template"
)

// Build compiles script into a standalone binary at out (default: the
// script's basename without its extension, in the current directory). The
// embedded app tree contains the script itself, the project's lib/ (when
// present), and the vendored local/lib/perl5 (resolved from the cpanfile
// first when needed).
//
// The GPERL_DEV_REPLACE environment variable ("module=path,module=path")
// adds replace directives to the generated module — the way to build against
// an unreleased go-perl/perlwasm2go checkout.
func Build(script, out string) error {
	script, err := filepath.Abs(script)
	if err != nil {
		return err
	}
	if _, err := os.Stat(script); err != nil {
		return err
	}
	projectDir := filepath.Dir(script)
	if err := EnsureDeps(projectDir); err != nil {
		return err
	}
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
	defer os.RemoveAll(work)

	scriptName := filepath.Base(script)
	appZip, incDirs, err := buildAppZip(projectDir, script)
	if err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(work, "app.zip"), appZip, 0o644); err != nil {
		return err
	}
	var mainSrc bytes.Buffer
	if err := mainTemplate.Execute(&mainSrc, map[string]any{
		"ScriptPath": "app/" + scriptName,
		"Inc":        incDirs,
	}); err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(work, "main.go"), mainSrc.Bytes(), 0o644); err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(work, "go.mod"), goModSrc(), 0o644); err != nil {
		return err
	}

	for _, argv := range [][]string{
		{"go", "mod", "tidy"},
		{"go", "build", "-o", out, "."},
	} {
		cmd := exec.Command(argv[0], argv[1:]...)
		cmd.Dir = work
		cmd.Stdout = os.Stderr
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("%s: %w", strings.Join(argv, " "), err)
		}
	}
	// Native XS modules cannot be embedded (they are dlopen'd host
	// libraries); ship them next to the binary, where the generated main
	// looks by default.
	if srcXS := xsDir(projectDir); dirExists(srcXS) {
		dstXS := out + ".xs"
		if err := os.RemoveAll(dstXS); err != nil {
			return err
		}
		if err := copyTree(srcXS, dstXS); err != nil {
			return err
		}
		fmt.Fprintf(os.Stderr, "gperl: native XS modules copied to %s\n", dstXS)
	}
	fmt.Fprintf(os.Stderr, "gperl: built %s\n", out)
	return nil
}

// buildAppZip assembles the embedded app tree: the script at /app/<name>,
// the project's lib/ and the vendored local/lib/perl5 when they exist.
// Returns the zip bytes and the guest @INC additions, most specific first.
func buildAppZip(projectDir, script string) ([]byte, []string, error) {
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

	if err := addFile("app/"+filepath.Base(script), script); err != nil {
		return nil, nil, err
	}
	inc := []string{}
	if lib := localLib(projectDir); lib != "" {
		if err := addTree("app/local/lib/perl5", lib); err != nil {
			return nil, nil, err
		}
		inc = append(inc, "app/local/lib/perl5")
	}
	if st, err := os.Stat(filepath.Join(projectDir, "lib")); err == nil && st.IsDir() {
		if err := addTree("app/lib", filepath.Join(projectDir, "lib")); err != nil {
			return nil, nil, err
		}
		inc = append(inc, "app/lib")
	}
	inc = append(inc, "app")
	if err := zw.Close(); err != nil {
		return nil, nil, err
	}
	return buf.Bytes(), inc, nil
}

// goModSrc renders the generated program's go.mod. Versions resolve through
// `go mod tidy`; GPERL_DEV_REPLACE injects replace directives for unreleased
// checkouts.
func goModSrc() []byte {
	var b strings.Builder
	b.WriteString("module gperl.invalid/app\n\ngo 1.25.0\n\nrequire github.com/goccy/go-perl v0.0.0\n")
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
package main

import (
	"archive/zip"
	"bytes"
	"context"
	_ "embed"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	perl "github.com/goccy/go-perl"
	"github.com/goccy/go-perl/gperl"
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
	p, err := perl.New(perl.Config{
		HostFS: true,
		Stdin:  os.Stdin,
		Stdout: os.Stdout,
		Stderr: os.Stderr,
		Env:    os.Environ(),
	})
	if err != nil {
		fatal(err)
	}
	// Native XS modules built by gperl xs build: GPERL_XS_DIR wins, then
	// <binary>.xs next to the executable, then the project-layout dir.
	xsCandidates := []string{os.Getenv("GPERL_XS_DIR")}
	if exe, exeErr := os.Executable(); exeErr == nil {
		xsCandidates = append(xsCandidates, exe+".xs")
	}
	xsCandidates = append(xsCandidates, filepath.Join("local", "xs", gperl.XSArchTag()))
	for _, dir := range xsCandidates {
		if dir == "" {
			continue
		}
		if err := gperl.LoadXS(p, dir); err != nil {
			fatal(err)
		}
	}
	err = p.RunFile(context.Background(), scriptPath, incDirs, os.Args[1:])
	if err == nil {
		p.Close()
		return
	}
	if code, ok := perl.ExitCode(err); ok {
		// A guest exit() unwound cleanly; Close flushes PerlIO and runs
		// END blocks before the process reports the status.
		p.Close()
		cleanup()
		os.Exit(code)
	}
	var pe *perl.PerlError
	if errors.As(err, &pe) {
		msg := pe.Message
		if len(msg) == 0 || msg[len(msg)-1] != '\n' {
			msg += "\n"
		}
		fmt.Fprint(os.Stderr, msg)
		p.Close()
		cleanup()
		os.Exit(255)
	}
	fatal(err)
}

// cleanup removes the unpacked app dir; os.Exit skips defers, so every
// exit path calls it explicitly.
var cleanup = func() {}

func fatal(err error) {
	fmt.Fprintln(os.Stderr, err)
	cleanup()
	os.Exit(1)
}
`))
