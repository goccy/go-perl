//go:build darwin || linux

package gperl_test

import (
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

// perl's programs are where %Config says they are, beside the perl command,
// as they are beside an installed perl: a program that runs perldoc to show
// its manual (Pod::Usage at -verbose 2) or prove from beside $^X finds them,
// and they run on this perl.
func TestPerlProgramsAreBesideThePerlCommand(t *testing.T) {
	out, status := hostPerl(t, "-MConfig", "-e", `
(my $dir = $^X) =~ s{/[^/]*\z}{};
print "scriptdir: ", ($Config{scriptdir} eq $dir && $Config{scriptdirexp} eq $dir ? "beside the perl command" : "$Config{scriptdir} (\$^X is $^X)"), "\n";
print "bin: ", ($Config{bin} eq $dir && $Config{binexp} eq $dir ? "beside the perl command" : $Config{bin}), "\n";
for my $name (qw(prove perldoc pod2usage podchecker pod2man pod2text enc2xs piconv h2xs xsubpp corelist perlthanks)) {
    print "$name: ", (-x "$Config{scriptdir}/$name" ? "there" : "missing"), "\n";
}
open(my $fh, '<', "$Config{scriptdir}/prove") or die "prove: $!";
my $line = <$fh>;
print "interpreter: ", ($line eq "#!$^X\n" ? "the perl command" : $line), "\n";
my $version = qx("$Config{scriptdir}/prove" --version);
print "prove: ", ($version =~ /\ATAP::Harness v[\d.]+ and Perl v5\.44\.0\n\z/ ? "runs on this perl" : $version), "\n";
`)
	want := "scriptdir: beside the perl command\n" +
		"bin: beside the perl command\n" +
		"prove: there\nperldoc: there\npod2usage: there\npodchecker: there\npod2man: there\npod2text: there\n" +
		"enc2xs: there\npiconv: there\nh2xs: there\nxsubpp: there\ncorelist: there\nperlthanks: there\n" +
		"interpreter: the perl command\n" +
		"prove: runs on this perl\n"
	if status != 0 || out != want {
		t.Fatalf("status %d, output:\n%s\nwant:\n%s", status, out, want)
	}
}

// perl's programs run on this perl also with another home directory, whose
// cache has no link of its own yet: their #! line names the link that was
// made under the home directory of the time (a test runs a program with
// HOME set to a directory of its own).
func TestPerlProgramsRunWithAnotherHome(t *testing.T) {
	home := t.TempDir()
	out, status := hostPerl(t, "-MConfig", "-e", `
my $home = shift;
delete $ENV{XDG_CACHE_HOME};
local $ENV{HOME} = $home;
my $version = qx("$Config{scriptdir}/prove" --version);
print "prove: ", ($version =~ /\ATAP::Harness v[\d.]+ and Perl v5\.44\.0\n\z/ ? "runs on this perl" : $version), "\n";
`, home)
	if want := "prove: runs on this perl\n"; status != 0 || out != want {
		t.Fatalf("status %d, output:\n%s\nwant:\n%s", status, out, want)
	}
}

// A user who has no cache directory to write still has a perl command: it
// is beside the standard library, in this user's directory of the
// temporary directory, and it runs as perl.
func TestPerlCommandWithoutACacheDirectory(t *testing.T) {
	// A home that is a file has no cache directory, for root too.
	home := filepath.Join(t.TempDir(), "home")
	if err := os.WriteFile(home, nil, 0o644); err != nil {
		t.Fatal(err)
	}
	tmp := t.TempDir()
	cmd := perlProcess("-e", `print $^X; exit(system($^X, '-e', 'exit 3') >> 8)`)
	cmd.Env = append(cmd.Env, "HOME="+home, "XDG_CACHE_HOME=", "TMPDIR="+tmp)
	var errOut strings.Builder
	cmd.Stderr = &errOut
	out, err := cmd.Output()
	if cmd.ProcessState == nil || cmd.ProcessState.ExitCode() != 3 {
		t.Fatalf("perl: %v (want status 3 from $^X)\n%s", err, errOut.String())
	}
	link := string(out)
	root := filepath.Join(tmp, "go-perl-"+strconv.Itoa(os.Getuid()))
	if !strings.HasPrefix(link, root+string(filepath.Separator)+"perl-") || filepath.Base(link) != "perl" {
		t.Errorf("$^X = %q, want a perl link in %s", link, root)
	}
}

// A program that shows its whole manual (Pod::Usage at -verbose 2) has
// perldoc print it, as on perl.
func TestPodUsageShowsTheManualThroughPerldoc(t *testing.T) {
	script := filepath.Join(t.TempDir(), "tool")
	src := "use Pod::Usage;\npod2usage(-verbose => 2, -exitval => 0);\n__END__\n\n=head1 NAME\n\ntool - does a thing\n\n=head1 OPTIONS\n\nThe options of the tool.\n\n=cut\n"
	if err := os.WriteFile(script, []byte(src), 0o644); err != nil {
		t.Fatal(err)
	}
	out, status := hostPerl(t, script)
	if status != 0 || !strings.Contains(out, "tool - does a thing") || !strings.Contains(out, "The options of the tool.") {
		t.Fatalf("status %d, output:\n%s", status, out)
	}
}

// perl's manual pages are where perl installs them, in the pod directory of
// its library: perldoc finds the description of a function in perlfunc and
// a program browsing the documentation finds perl.pod.
func TestPerlManualPagesAreInstalled(t *testing.T) {
	out, status := hostPerl(t, "-MConfig", "-e", `
for my $page (qw(perl perlfunc perldiag perlvar perlre perlapi perlmodlib perluniprops)) {
    print "$page: ", (-f "$Config{privlibexp}/pod/$page.pod" ? "there" : "missing"), "\n";
}
my $doc = qx("$Config{scriptdir}/perldoc" -T -f sprintf);
print "perldoc -f: ", ($doc =~ /sprintf FORMAT, LIST/ ? "found" : $doc), "\n";
`)
	want := "perl: there\nperlfunc: there\nperldiag: there\nperlvar: there\nperlre: there\n" +
		"perlapi: there\nperlmodlib: there\nperluniprops: there\n" +
		"perldoc -f: found\n"
	if status != 0 || out != want {
		t.Fatalf("status %d, output:\n%s\nwant:\n%s", status, out, want)
	}
}
