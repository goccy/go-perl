package perl_test

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	perl "github.com/goccy/go-perl"
	goperlfs "github.com/goccy/go-perl/fs"
)

// TestDefaultFSIsSandboxed pins the library default: an instance built from
// a bare Config gets a private in-memory filesystem — host files are
// invisible, writes stay inside the instance, and instances do not share
// state. Tools that want perl-like behavior opt in with fs.NewHostFS().
func TestDefaultFSIsSandboxed(t *testing.T) {
	secret := filepath.Join(t.TempDir(), "secret.txt")
	if err := os.WriteFile(secret, []byte("host data"), 0o644); err != nil {
		t.Fatal(err)
	}

	p, err := perl.New(perl.Config{})
	if err != nil {
		t.Fatal(err)
	}
	defer p.Close()
	ctx := context.Background()

	// Host files are not visible.
	r, err := p.Eval(ctx, `-e '`+secret+`' ? "visible" : "hidden"`)
	if err != nil || r.Error != nil {
		t.Fatalf("host visibility probe: err=%v error=%v", err, r.Error)
	}
	if r.Value.String() != "hidden" {
		t.Fatalf("sandboxed instance sees host file: %q", r.Value.String())
	}

	// Writes land in the private FS and read back within the instance...
	r, err = p.Eval(ctx, `
		open my $fh, '>', '/scratch.txt' or die "open: $!";
		print $fh "private";
		close $fh;
		open my $in, '<', '/scratch.txt' or die "reopen: $!";
		local $/; my $got = <$in>;
		close $in;
		$got;
	`)
	if err != nil || r.Error != nil {
		t.Fatalf("private write/read: err=%v error=%v", err, r.Error)
	}
	if r.Value.String() != "private" {
		t.Fatalf("private FS round trip = %q", r.Value.String())
	}

	// ...but never reach the host disk.
	if _, err := os.Stat("/scratch.txt"); err == nil {
		t.Fatal("guest write leaked to the host filesystem")
	}

	// A second instance shares nothing with the first.
	p2, err := perl.New(perl.Config{})
	if err != nil {
		t.Fatal(err)
	}
	defer p2.Close()
	r, err = p2.Eval(ctx, `-e '/scratch.txt' ? "shared" : "isolated"`)
	if err != nil || r.Error != nil {
		t.Fatalf("isolation probe: err=%v error=%v", err, r.Error)
	}
	if r.Value.String() != "isolated" {
		t.Fatalf("instances share filesystem state: %q", r.Value.String())
	}
}

// TestHostFSOptIn pins the perl-like mode gperl uses: the host filesystem
// via the fs package, with the stdlib extracted onto it.
func TestHostFSOptIn(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "data.txt")
	if err := os.WriteFile(file, []byte("from host"), 0o644); err != nil {
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
	ctx := context.Background()

	r, err := p.Eval(ctx, `
		open my $fh, '<', '`+file+`' or die "open: $!";
		local $/; my $got = <$fh>;
		close $fh;
		open my $out, '>', '`+filepath.Join(dir, "out.txt")+`' or die "write: $!";
		print $out uc($got);
		close $out;
		$got;
	`)
	if err != nil || r.Error != nil {
		t.Fatalf("host read/write: err=%v error=%v", err, r.Error)
	}
	if r.Value.String() != "from host" {
		t.Fatalf("host read = %q", r.Value.String())
	}
	data, err := os.ReadFile(filepath.Join(dir, "out.txt"))
	if err != nil || !strings.EqualFold(string(data), "from host") {
		t.Fatalf("host write-back = %q, err=%v", data, err)
	}
}
