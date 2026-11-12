package gperl

import (
	"bytes"
	"os"
	"strings"
	"testing"
)

// overlayRun runs a perl command line the way a child of the XS build runs:
// with the build's %Config overlay as its preload.
func overlayRun(t *testing.T, argv ...string) string {
	t.Helper()
	overlay, err := xsWriteConfigOverlay(t.TempDir(), "/the/sdk", "/the/stdlib", "/the/arch", "/the/perl")
	if err != nil {
		t.Fatal(err)
	}
	var out, errOut bytes.Buffer
	status, err := runPerlCLI(argv, perlRunOpts{
		env:    append(os.Environ(), "GPERL_PERL_PRELOAD="+overlay),
		stdin:  strings.NewReader(""),
		stdout: &out,
		stderr: &errOut,
	})
	if err != nil || status != 0 {
		t.Fatalf("perl %q = (%d, %v), stderr %q", argv, status, err, errOut.String())
	}
	return out.String()
}

// The build's overlay changes values of %Config and nothing else about it:
// it stays the tied, read-only view of the Config object.
func TestConfigOverlayKeepsConfigTied(t *testing.T) {
	got := overlayRun(t, "-MConfig", "-e", `
print "tied to: ", ref(tied %Config), "\n";
print "overlaid: cc=$Config{cc} privlibexp=$Config{privlibexp} archlibexp=$Config{archlibexp} perlpath=$Config{perlpath}\n";
print "the interpreter's own: ivsize=$Config{ivsize} version=", ($Config{version} eq sprintf("%vd", $^V) ? "this perl's" : $Config{version}), "\n";
print "exists: ", (exists $Config{cc} ? 1 : 0), (exists $Config{no_such_key} ? 1 : 0), "\n";
print "assignment: ", (eval { $Config{cc} = "other"; 1 } ? "accepted" : "refused"), "\n";
print "after it: cc=$Config{cc}\n";
print "listed: ", (scalar(grep { $_ eq "cc" } keys %Config) ? "cc among the keys" : "cc missing"), "\n";
`)
	want := "tied to: Config\n" +
		"overlaid: cc=cc privlibexp=/the/stdlib archlibexp=/the/arch perlpath=/the/perl\n" +
		"the interpreter's own: ivsize=4 version=this perl's\n" +
		"exists: 10\n" +
		"assignment: refused\n" +
		"after it: cc=cc\n" +
		"listed: cc among the keys\n"
	if got != want {
		t.Fatalf("got:\n%s\nwant:\n%s", got, want)
	}
}
