package gperl

import (
	"bytes"
	"os"
	"os/exec"
	"runtime"
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
		"the interpreter's own: ivsize=8 version=this perl's\n" +
		"exists: 10\n" +
		"assignment: refused\n" +
		"after it: cc=cc\n" +
		"listed: cc among the keys\n"
	if got != want {
		t.Fatalf("got:\n%s\nwant:\n%s", got, want)
	}
}

// The preload is how the interpreter was started, not something the program
// loaded: %INC does not list it.
func TestPreloadLeavesNoTraceInINC(t *testing.T) {
	got := overlayRun(t, "-e", `print join(",", sort grep { !/^(?:Config|strict|warnings|vars|Config_heavy)\b/ } keys %INC), "|", scalar(grep { /overlay/ } %INC), "\n";`)
	if want := "|0\n"; got != want {
		t.Fatalf("%%INC after the preload = %q, want %q", got, want)
	}
}

// The keys of the build overlay that identify the C toolchain describe the
// compiler of this machine, by its own account: a build script that looks
// at $Config{gccversion} or $Config{ccname} to choose flags, or runs
// $Config{cpprun}, must not find the compiler the interpreter was built
// with.
func TestConfigOverlayDescribesTheHostCompiler(t *testing.T) {
	cc, err := exec.LookPath("cc")
	if err != nil {
		t.Skip("no host C compiler")
	}
	macros, err := exec.Command(cc, "-dM", "-E", "-x", "c", os.DevNull).Output()
	if err != nil {
		t.Skipf("the host C compiler does not list its macros: %v", err)
	}
	wantName := "gcc"
	if strings.Contains(string(macros), "#define __clang__ ") {
		wantName = "clang"
	}
	raw := overlayRun(t, "-e", `
for my $key (qw(cc ccname gccversion cpp cpprun cppstdin cppflags libc libs ccsymbols)) {
    print "$key=", ($Config::Config{$key} // "undef"), "\n";
}
print "perl_version=", (defined $Config::Config{perl_version} ? "defined" : "undef"), "\n";
print "preprocessed=", (qx(echo __VERSION__ | $Config::Config{cpprun} -) =~ /"\Q$Config::Config{gccversion}\E"/ ? "by that compiler" : "by another"), "\n";
`)
	got := map[string]string{}
	for _, line := range strings.Split(strings.TrimSpace(raw), "\n") {
		key, value, _ := strings.Cut(line, "=")
		got[key] = value
	}
	if got["ccname"] != wantName {
		t.Errorf("ccname = %q, the host compiler is %s", got["ccname"], wantName)
	}
	if got["gccversion"] == "" || strings.Contains(got["gccversion"], "wasi") {
		t.Errorf("gccversion = %q, want the host compiler's __VERSION__", got["gccversion"])
	}
	if wantName == "gcc" && strings.Contains(strings.ToLower(got["gccversion"]), "clang") {
		t.Errorf("gccversion = %q names clang, the host compiler is gcc", got["gccversion"])
	}
	for _, key := range []string{"cpp", "cpprun", "cppstdin"} {
		if got[key] != got["cc"]+" -E" {
			t.Errorf("%s = %q, want the host compiler's preprocessor (%q)", key, got[key], got["cc"]+" -E")
		}
	}
	for _, key := range []string{"cppflags", "libc", "libs", "ccsymbols"} {
		if strings.Contains(got[key], "wasi") || strings.Contains(got[key], "__arm__") {
			t.Errorf("%s = %q describes the compiler the interpreter was built with", key, got[key])
		}
	}
	if got["preprocessed"] != "by that compiler" {
		t.Errorf("$Config{cpprun} is not the compiler $Config{gccversion} names: %v", got)
	}
	// perl has no such key (some vendors' builds add one): a build script
	// that reads it gets what it gets from perl.
	if got["perl_version"] != "undef" {
		t.Errorf("$Config{perl_version} is %s, perl leaves it undefined", got["perl_version"])
	}
}

// perl -V prints the summary of the configuration and what the binary was
// built with, under the overlay as without it: the overlay gives $^O the
// host's name, and Config::_V's check that the library was built for the
// system that runs it is made with the interpreter's own system. The
// summary shows %Config as the overlay leaves it.
func TestConfigOverlayLeavesPerlDashVWorking(t *testing.T) {
	got := overlayRun(t, "-V")
	for _, want := range []string{
		"Summary of my perl5 (revision 5 version 44",
		"    osname=" + runtime.GOOS + "\n",
		"    cc='cc'\n",
		"Characteristics of this binary (from libperl):",
		"  @INC:\n",
	} {
		if !strings.Contains(got, want) {
			t.Errorf("perl -V under the overlay does not print %q:\n%s", want, got)
		}
	}
	status, plain, stderr := runCLI(t, t.TempDir(), "", "-V")
	if status != 0 || !strings.Contains(plain, "Summary of my perl5") || !strings.Contains(plain, "  @INC:\n") {
		t.Errorf("perl -V without the overlay: status=%d stderr=%q stdout:\n%s", status, stderr, plain)
	}
	// $^O stays the host's for the rest of the program.
	if got := overlayRun(t, "-V", "-e", `print "after: $^O\n"`); !strings.HasSuffix(got, "after: "+runtime.GOOS+"\n") {
		t.Errorf("$^O after the summary, under the overlay: %q", got[max(0, len(got)-40):])
	}
}
