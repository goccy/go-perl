package perl_test

import (
	"strings"
	"testing"

	perl "github.com/goccy/go-perl"
)

// %Config names the directory the standard library is in at run time —
// wherever the library was put, on the host or in memory — under every
// spelling of the library directory, and stays the tied hash it is.
func TestConfigNamesTheLibraryDirectory(t *testing.T) {
	program := perl.Program{Code: code(
		`use Config;`,
		`print "tied=", ref(tied %Config), "\n";`,
		`for my $key (qw(privlibexp privlib archlibexp archlib installprivlib installarchlib)) {`,
		`    print "$key holds the library=", (-e "$Config{$key}/strict.pm" ? 1 : 0), "\n";`,
		`}`,
		`print "one directory=", ($Config{privlibexp} eq $Config{archlib} ? 1 : 0), "\n";`,
		`my ($loaded) = $INC{"strict.pm"} =~ m{\A(.*?)/*strict\.pm\z}s;`,
		`print "the directory strict.pm was loaded from=", ($Config{privlibexp} eq ($loaded eq "" ? "/" : $loaded) ? 1 : 0), "\n";`,
		`print "config_vars: "; Config::config_vars("privlibexp");`,
		`print "expected:    privlibexp='$Config{privlibexp}';\n";`,
		`print "version=", ($Config{version} eq sprintf("%vd", $^V) ? "this perl's" : $Config{version}), "\n";`,
	)}
	for name, cfg := range map[string]perl.Config{"in memory": {}, "on the host": hostConfig(t)} {
		t.Run(name, func(t *testing.T) {
			r := runProgram(t, cfg, program, "")
			lines := strings.Split(strings.TrimRight(r.stdout, "\n"), "\n")
			if r.status != 0 || len(lines) != 12 {
				t.Fatalf("status=%d stdout=%q stderr=%q", r.status, r.stdout, r.stderr)
			}
			want := []string{
				"tied=Config",
				"privlibexp holds the library=1",
				"privlib holds the library=1",
				"archlibexp holds the library=1",
				"archlib holds the library=1",
				"installprivlib holds the library=1",
				"installarchlib holds the library=1",
				"one directory=1",
				"the directory strict.pm was loaded from=1",
			}
			if got := lines[:9]; strings.Join(got, "\n") != strings.Join(want, "\n") {
				t.Fatalf("got:\n  %s\nwant:\n  %s", strings.Join(got, "\n  "), strings.Join(want, "\n  "))
			}
			if got, want := strings.TrimPrefix(lines[9], "config_vars: "), strings.TrimPrefix(lines[10], "expected:    "); got != want {
				t.Fatalf("config_vars prints %q, %%Config holds %q", got, want)
			}
			if lines[11] != "version=this perl's" {
				t.Fatalf("%s", lines[11])
			}
		})
	}
}

// The diagnostics pragma finds the explanations of perl's messages in the
// library directory %Config names.
func TestDiagnosticsFindsItsMessages(t *testing.T) {
	for name, cfg := range map[string]perl.Config{"in memory": {}, "on the host": hostConfig(t)} {
		t.Run(name, func(t *testing.T) {
			r := runProgram(t, cfg, perl.Program{Code: code(
				`use diagnostics;`,
				`print "loaded\n";`,
				`my $explained = diagnostics::splainthis("Use of uninitialized value in addition (+) at x line 1.\n");`,
				`print "explains a message=", (defined $explained && length $explained ? 1 : 0), "\n";`,
			)}, "")
			if r.status != 0 || !strings.HasPrefix(r.stdout, "loaded\n") {
				t.Fatalf("status=%d stdout=%q stderr=%q", r.status, r.stdout, r.stderr)
			}
			if !strings.Contains(r.stderr, "uninitialized") && !strings.Contains(r.stdout, "explains a message=1") {
				t.Fatalf("no explanation: stdout=%q stderr=%q", r.stdout, r.stderr)
			}
		})
	}
}
