//go:build darwin || linux

package perl_test

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	perl "github.com/goccy/go-perl"
	goperlfs "github.com/goccy/go-perl/fs"
)

// readdir lists a directory of the host's in the order the system gives
// its names, with "." and ".." where the system puts them, as perl's
// readdir does: the names other than "." and ".." are in the order
// Readdirnames reads them from the system, and "." and ".." are listed once
// each. The names were sorted, with "."
// and ".." first (PerlIO::Util's t/10_dir.t compares readdir with its own
// reading of the directory). A directory of the sandbox, which has no
// order of its own, lists "." and ".." first and then its names sorted.
func TestReaddirOrder(t *testing.T) {
	var names []string
	for i := 0; i < 64; i++ {
		names = append(names, fmt.Sprintf("f%02d", (i*37)%64))
	}
	ctx := context.Background()
	t.Run("host", func(t *testing.T) {
		dir := t.TempDir()
		for _, name := range names {
			if err := os.WriteFile(filepath.Join(dir, name), nil, 0o644); err != nil {
				t.Fatal(err)
			}
		}
		d, err := os.Open(dir)
		if err != nil {
			t.Fatal(err)
		}
		sys, err := d.Readdirnames(-1)
		if err := errors.Join(err, d.Close()); err != nil {
			t.Fatal(err)
		}
		stdlib, err := perl.ExtractStdlib()
		if err != nil {
			t.Fatal(err)
		}
		p, err := perl.New(perl.Config{FS: goperlfs.NewHostFS(), StdlibDir: stdlib})
		if err != nil {
			t.Fatal(err)
		}
		defer p.Close()
		r, err := p.Eval(ctx, `opendir my $d, '`+dir+`' or die "opendir: $!"; join " ", readdir $d`)
		if err != nil || r.Error != nil {
			t.Fatalf("err=%v error=%v", err, r.Error)
		}
		listed := strings.Fields(resultStr(r))
		others := slices.DeleteFunc(slices.Clone(listed), func(name string) bool { return name == "." || name == ".." })
		if !slices.Equal(others, sys) || len(listed)-len(others) != 2 || !slices.Contains(listed, ".") || !slices.Contains(listed, "..") {
			t.Errorf("readdir lists\n%s\nthe system lists, besides . and ..\n%s", strings.Join(listed, " "), strings.Join(sys, " "))
		}
	})
	t.Run("sandbox", func(t *testing.T) {
		p, err := perl.New(perl.Config{})
		if err != nil {
			t.Fatal(err)
		}
		defer p.Close()
		r, err := p.Eval(ctx, `mkdir '/listed' or die "mkdir: $!";
for my $name (qw(`+strings.Join(names, " ")+`)) { open my $fh, '>', "/listed/$name" or die "open: $!" }
opendir my $d, '/listed' or die "opendir: $!"; join " ", readdir $d`)
		if err != nil || r.Error != nil {
			t.Fatalf("err=%v error=%v", err, r.Error)
		}
		want := ". .."
		for i := 0; i < 64; i++ {
			want += fmt.Sprintf(" f%02d", i)
		}
		if got := resultStr(r); got != want {
			t.Errorf("readdir lists\n%s\nwant\n%s", got, want)
		}
	})
}
