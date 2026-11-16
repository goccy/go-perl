//go:build darwin || linux

package xs_test

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

// An entry of a hash the interpreter hands to native code is the entry
// itself: a value stored into HeVAL is the hash's value (Package::Stash::XS
// gives a stash's entry its glob so), an entry is the same from one call
// to the next, and the chains of HvARRAY's buckets hold the entries, whose
// values are read and written in place. The SDK handed out copies before:
// a store into one was lost.
func TestNativeEntriesOfAHashInPlace(t *testing.T) {
	p := newHostPerl(t)
	perlAPIInstance(t, p)
	runEvalCases(t, p, []evalCase{
		{"store into a new entry", `my %h; PerlAPI::entry_assign(\%h, "k", 42); join ",", map { "$_=$h{$_}" } sort keys %h`, "k=42"},
		{"store into an entry", `my %h = (k => 1, j => 2); PerlAPI::entry_assign(\%h, "k", "v"); join ",", map { "$_=$h{$_}" } sort keys %h`, "j=2,k=v"},
		{"same entry", `my %h = (a => 1); join ",", PerlAPI::entry_same(\%h, "a"), PerlAPI::entry_same(\%h, "a")`, "first,same"},
		{"buckets in place", `my %h = map { $_ => $_ } 1 .. 50; my $r = PerlAPI::bucket_values(\%h); my $s = 0; $s += $_ for values %h; "$r,$s"`, "50:1275:50,2550"},
	})
}

// Entries a module makes of its own read through the macros of hv.h as
// the interpreter's do: copies in a string buffer of the interpreter's
// pointing at a key in the same buffer that is a scalar (Cpanel::JSON::XS
// copies a tied hash's entries so), and copies in memory of the process
// with the key's bytes.
func TestNativeEntriesOfTheModulesOwn(t *testing.T) {
	p := newHostPerl(t)
	perlAPIInstance(t, p)
	runEvalCases(t, p, []evalCase{
		{"plain hash", `PerlAPI::own_entries({b => 2, a => 1, c => 3})`, "a=1,b=2,c=3|a=1,b=2,c=3"},
		{"tied hash", `package Tied { sub TIEHASH { bless { x => 10, y => 20 } } sub FETCH { $_[0]{$_[1]} } sub FIRSTKEY { my @k = sort keys %{$_[0]}; $k[0] } sub NEXTKEY { my @k = sort keys %{$_[0]}; for my $i (0 .. $#k) { return $k[$i + 1] if $k[$i] eq $_[1] } undef } } tie my %t, "Tied"; PerlAPI::own_entries(\%t)`, "x=10,y=20|x=10,y=20"},
	})
}

// The loader refuses a module whose layout of the interpreter's entries
// of hashes is not the interpreter's (it reads and writes them in place),
// and says which.
func TestNativeLayoutDisagreementRefused(t *testing.T) {
	cc, err := exec.LookPath("cc")
	if err != nil {
		t.Skip("no host C compiler available")
	}
	dir := t.TempDir()
	so := filepath.Join(dir, "Layout.so")
	args := []string{"-shared", "-fPIC", "-I", "sdk/include", `-DXS_VERSION="1.00"`, "-o", so,
		"testdata/layout/Layout.c", "testdata/layout/wrong.c"}
	if runtime.GOOS == "darwin" {
		switch runtime.GOARCH {
		case "arm64":
			args = append([]string{"-arch", "arm64"}, args...)
		case "amd64":
			args = append([]string{"-arch", "x86_64"}, args...)
		}
	}
	if out, err := exec.Command(cc, args...).CombinedOutput(); err != nil {
		t.Fatalf("compile Layout.so: %v\n%s", err, out)
	}
	p := newHostPerl(t)
	xsdir := t.TempDir()
	if err := os.Symlink(so, filepath.Join(xsdir, "Layout.so")); err != nil {
		t.Fatal(err)
	}
	err = p.AddXSDir(xsdir)
	if err == nil || !strings.Contains(err.Error(), "sizeof(HE) = 32, the interpreter has 24") {
		t.Fatalf("AddXSDir: %v, want the layout disagreement", err)
	}
	r, err := p.Eval(context.Background(), `defined &Layout::hello ? "defined" : "not defined"`)
	if err != nil {
		t.Fatalf("Eval: %v", err)
	}
	if got := resultStr(r); got != "not defined" {
		t.Fatalf("got %q after the refusal", got)
	}
}

// The link of an entry of the interpreter's is an address of its memory:
// native code that follows hent_next as a pointer does not compile (HeNEXT
// gives the next entry).
func TestNativeEntryLinkIsNotAPointer(t *testing.T) {
	cc, err := exec.LookPath("cc")
	if err != nil {
		t.Skip("no host C compiler available")
	}
	dir := t.TempDir()
	for name, body := range map[string]string{
		"follow": "HE *next_of(HE *he) { return he->hent_next; }",
		"heNEXT": "HE *next_of(HE *he) { return HeNEXT(he); }",
	} {
		src := filepath.Join(dir, name+".c")
		code := "#include \"EXTERN.h\"\n#include \"perl.h\"\n#include \"XSUB.h\"\n" + body + "\n"
		if err := os.WriteFile(src, []byte(code), 0o644); err != nil {
			t.Fatal(err)
		}
		out, err := exec.Command(cc, "-fsyntax-only", "-I", "sdk/include", src).CombinedOutput()
		if name == "follow" && err == nil {
			t.Errorf("he->hent_next compiled as a pointer")
		}
		if name == "heNEXT" && err != nil {
			t.Errorf("HeNEXT did not compile: %v\n%s", err, out)
		}
	}
}
