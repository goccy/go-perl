//go:build darwin || linux

package xs_test

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"testing"

	perl "github.com/goccy/go-perl"
)

// buildLayerView compiles testdata/perliol/LayerView.c, a module written
// against <perliol.h>, as buildPerlAPI does PerlAPI.c.
func buildLayerView(t testing.TB) string {
	t.Helper()
	cc, err := exec.LookPath("cc")
	if err != nil {
		t.Skip("no host C compiler available")
	}
	so := filepath.Join(t.TempDir(), "LayerView.so")
	args := []string{"-shared", "-fPIC", "-Wall", "-Werror=implicit-function-declaration",
		"-I", "sdk/include", "-o", so, "testdata/perliol/LayerView.c", "-lm"}
	if runtime.GOOS == "darwin" {
		switch runtime.GOARCH {
		case "arm64":
			args = append([]string{"-arch", "arm64"}, args...)
		case "amd64":
			args = append([]string{"-arch", "x86_64"}, args...)
		}
	} else {
		args = append([]string{"-Wl,-z,defs", "-D_GNU_SOURCE"}, args...)
	}
	if out, err := exec.Command(cc, args...).CombinedOutput(); err != nil {
		t.Fatalf("compile LayerView.so: %v\n%s", err, out)
	}
	return so
}

// Code that walks the layers of a handle through <perliol.h> sees the
// interpreter's layers: their names, whether their kind is buffered, the
// flags of each (UTF-8, CRLF); and the handle functions work on the
// handle, and on a layer of it (PerlIONext), as given.
func TestNativeLayersOfAHandle(t *testing.T) {
	p := newHostPerl(t)
	so := buildLayerView(t)
	loadXSModule(t, p, "LayerView", so)
	libdir := t.TempDir()
	if err := os.WriteFile(filepath.Join(libdir, "LayerView.pm"),
		[]byte("package LayerView;\nuse XSLoader;\nXSLoader::load(__PACKAGE__);\n1;\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	file := filepath.Join(t.TempDir(), "data.txt")
	if err := os.WriteFile(file, []byte("abc\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	if r, err := p.Eval(ctx, `unshift @INC, '`+libdir+`'; require LayerView; 1;`); err != nil || r.Error != nil {
		t.Fatalf("setup: err=%v error=%v", err, r.Error)
	}
	if r, err := p.Eval(ctx, `sub __t_file { our $FILE = shift; 1 } 1;`); err != nil || r.Error != nil {
		t.Fatalf("file helper: err=%v error=%v", err, r.Error)
	}
	if _, err := p.Call(ctx, "__t_file", perl.ValueOf(file)); err != nil {
		t.Fatalf("pass file: %v", err)
	}
	// the names from the top, as PerlIO::get_layers gives them from the
	// bottom (with "utf8" after a layer that has the flag)
	const names = `sub __t_names { my $fh = shift; my @want = grep { $_ ne "utf8" } reverse PerlIO::get_layers($fh); my @got = map { (split m{/})[0] } grep { m{/} } split /,/, LayerView::layers($fh); "@got" eq "@want" ? "same" : "[@got] vs [@want]" } `
	runEvalCases(t, p, []evalCase{
		// the descriptor native code is given is the system's (which
		// fileno numbers its own way), the same for the handle and for
		// the layer below its top
		{"raw", names + `open my $fh, "<:raw", our $FILE or die $!; my $r = LayerView::layers($fh); join ",", __t_names($fh), $r =~ m{^perlio/1/0/0,unix/0/0/0,} ? "perlio over unix" : $r, $r =~ /,(\d+),\1,0\z/ ? "descriptors" : $r`, "same,perlio over unix,descriptors"},
		{"utf8", names + `open my $fh, "<:raw:utf8", our $FILE or die $!; my $r = LayerView::layers($fh); join ",", __t_names($fh), $r =~ m{^\w+/\d/1/} ? "utf8 on the top" : $r`, "same,utf8 on the top"},
		{"crlf", names + `open my $fh, "<:raw:crlf", our $FILE or die $!; my $r = LayerView::layers($fh); join ",", __t_names($fh), $r =~ m{^crlf/1/\d/1,} ? "crlf on the top" : $r`, "same,crlf on the top"},
		{"at the end", `open my $fh, "<:raw", our $FILE or die $!; my @l = <$fh>; my $r = LayerView::layers($fh); $r =~ /,(\d)\z/; $1`, "1"},
	})
}
