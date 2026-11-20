package perl_test

import (
	"context"
	"testing"
	"testing/fstest"

	perl "github.com/goccy/go-perl"
	goperlfs "github.com/goccy/go-perl/fs"
	"github.com/goccy/go-perl/internal"
)

// useCPANLib installs a module tree the way a generated go-perl/cpan
// package does, for the duration of the test.
func useCPANLib(t *testing.T) {
	t.Helper()
	internal.SetCPANLib(fstest.MapFS{
		"Greeting.pm":       {Data: []byte("package Greeting;\nuse Greeting::Word;\nsub hello { Greeting::Word::word() . ', ' . $_[0] }\n1;\n")},
		"Greeting/Word.pm":  {Data: []byte("package Greeting::Word;\nsub word { 'hello' }\n1;\n")},
		"Greeting/data.txt": {Data: []byte("payload")},
	})
	t.Cleanup(func() { internal.SetCPANLib(nil) })
}

func evalString(t *testing.T, p *perl.Perl, src string) string {
	t.Helper()
	r, err := p.Eval(context.Background(), src)
	if err != nil {
		t.Fatalf("Eval(%s): %v", src, err)
	}
	if r.Error != nil {
		t.Fatalf("Eval(%s): %v", src, r.Error)
	}
	return resultStr(r)
}

// The module tree the cpan package carries is on @INC of an instance on
// the sandboxed default filesystem and of one on the host filesystem.
func TestCPANLibIsOnINC(t *testing.T) {
	useCPANLib(t)
	stdlib, err := perl.ExtractStdlib()
	if err != nil {
		t.Fatal(err)
	}
	for name, cfg := range map[string]perl.Config{
		"default": {},
		"host":    {FS: goperlfs.NewHostFS(), StdlibDir: stdlib},
	} {
		t.Run(name, func(t *testing.T) {
			p, err := perl.New(cfg)
			if err != nil {
				t.Fatal(err)
			}
			defer p.Close()
			if got := evalString(t, p, `use Greeting; Greeting::hello("gopher")`); got != "hello, gopher" {
				t.Fatalf("Greeting::hello = %q", got)
			}
		})
	}
}

// Without a carried tree an instance has no such directory.
func TestNoCPANLibByDefault(t *testing.T) {
	p, err := perl.New(perl.Config{})
	if err != nil {
		t.Fatal(err)
	}
	defer p.Close()
	if got := evalString(t, p, `(-e "/.goperl" ? "present" : "absent") . (grep({ m{\.goperl} } @INC) ? "+inc" : "")`); got != "absent" {
		t.Fatalf("default instance: %q", got)
	}
}
