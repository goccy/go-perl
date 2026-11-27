package perl_test

import (
	"strings"
	"testing"

	perl "github.com/goccy/go-perl"
	"github.com/goccy/go-perl/internal"
)

// siteProgram reports what the interpreter says about its site directories.
var siteProgram = perl.Program{Code: code(
	`use Config;`,
	`print "tied=", ref(tied %Config), "\n";`,
	`print "$_=$Config{$_}\n" for qw(siteprefix sitelib sitelibexp installsitelib sitearch sitearchexp installsitearch sitebin installsitescript installsiteman1dir siteman3dir sitelib_stem);`,
	`print "config_vars: "; Config::config_vars("installsitelib");`,
	`my ($stdlib) = grep { -e "$_/strict.pm" } @INC;`,
	`print "inc=", join(" ", map { $_ eq $stdlib ? "STDLIB" : $_ } @INC), "\n";`,
	`print "version=", sprintf("%vd", $^V), "\n";`,
)}

func siteReport(t *testing.T, cfg perl.Config) map[string]string {
	t.Helper()
	r := runProgram(t, cfg, siteProgram, "")
	if r.status != 0 {
		t.Fatalf("status=%d stdout=%q stderr=%q", r.status, r.stdout, r.stderr)
	}
	got := map[string]string{}
	for _, line := range strings.Split(strings.TrimRight(r.stdout, "\n"), "\n") {
		key, value, _ := strings.Cut(line, "=")
		got[key] = value
	}
	return got
}

// An interpreter that is given a site directory has its site directories
// under it: %Config names them under every key, and the two with modules
// are in @INC before the standard library, the architecture-specific one
// first.
func TestSiteDirectoriesOfAnInstance(t *testing.T) {
	cfg := perl.Config{SiteDir: "/opt/site"}
	got := siteReport(t, cfg)
	lib := "/opt/site/site_perl/" + got["version"]
	arch := lib + "/" + internal.ArchName()
	want := map[string]string{
		"tied":       "Config",
		"siteprefix": "/opt/site", "sitelib_stem": "/opt/site/site_perl",
		"sitelib": lib, "sitelibexp": lib, "installsitelib": lib,
		"sitearch": arch, "sitearchexp": arch, "installsitearch": arch,
		"sitebin": "/opt/site/bin", "installsitescript": "/opt/site/bin",
		"installsiteman1dir": "/opt/site/man/man1", "siteman3dir": "/opt/site/man/man3",
		"config_vars: installsitelib": "'" + lib + "';",
		"inc":                         arch + " " + lib + " STDLIB",
	}
	for key, value := range want {
		if got[key] != value {
			t.Errorf("%s = %q, want %q", key, got[key], value)
		}
	}
	if got["version"] == "" {
		t.Errorf("the interpreter did not say its version")
	}

	// A clone and a second instance have the same.
	p, err := perl.New(cfg)
	if err != nil {
		t.Fatal(err)
	}
	defer p.Close()
	clone, err := p.Clone()
	if err != nil {
		t.Fatal(err)
	}
	defer clone.Close()
	r, err := clone.Eval(t.Context(), `require Config; "$Config::Config{sitelib} @INC[0, 1]"`)
	if err != nil || r.Error != nil {
		t.Fatalf("Eval: %v %v", err, r.Error)
	}
	s, err := perl.As[perl.ScalarValue](r.Value)
	if err != nil {
		t.Fatal(err)
	}
	if want := lib + " " + arch + " " + lib; s.String() != want {
		t.Errorf("clone: %q, want %q", s.String(), want)
	}
}

// An interpreter that is given none has none: @INC is the standard library
// alone and %Config keeps the directories of the build's configuration —
// in memory and on the host alike.
func TestNoSiteDirectoriesByDefault(t *testing.T) {
	for name, cfg := range map[string]perl.Config{"in memory": {}, "on the host": hostConfig(t)} {
		t.Run(name, func(t *testing.T) {
			got := siteReport(t, cfg)
			lib := "/usr/lib/perl5/site_perl/" + got["version"]
			want := map[string]string{
				"tied":       "Config",
				"siteprefix": "/usr", "sitelib": lib, "sitelibexp": lib, "installsitelib": lib,
				"sitearch": lib + "/" + internal.ArchName(), "sitearchexp": lib + "/" + internal.ArchName(),
				"sitebin": "/usr/bin",
				"inc":     "STDLIB",
			}
			for key, value := range want {
				if got[key] != value {
					t.Errorf("%s = %q, want %q", key, got[key], value)
				}
			}
		})
	}
}
