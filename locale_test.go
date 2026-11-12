package perl_test

import (
	"testing"

	perl "github.com/goccy/go-perl"
)

// The C library has the data of the locales "C" (also "POSIX") and
// "C.UTF-8". A program that asks for another locale by name is told that it
// is not there, as on a system where that locale is not installed, instead
// of being given "C" under the name it asked for.
func TestSetlocaleRefusesLocalesWithoutData(t *testing.T) {
	r := runProgram(t, perl.Config{Env: []string{"LANG=C"}}, perl.Program{Code: code(
		`use POSIX qw(setlocale localeconv LC_ALL LC_CTYPE LC_NUMERIC LC_MONETARY);`,
		`for my $name ("de_DE", "de_DE.UTF-8", "de_DE.utf8", "ru_RU.ISO8859-5", "en_US", "en_US.UTF-8", "C.UTF-8;de_DE;C;C;C;C", "no such locale") {`,
		`    for my $category ([LC_ALL, "all"], [LC_NUMERIC, "numeric"], [LC_MONETARY, "monetary"], [LC_CTYPE, "ctype"]) {`,
		`        my $got = setlocale($category->[0], $name);`,
		`        print "$name ($category->[1]): accepted as $got\n" if defined $got;`,
		`    }`,
		`}`,
		`print "after the refusals: ", setlocale(LC_ALL), "\n";`,
		`for my $name ("C", "POSIX", "C.UTF-8", "C.utf8") {`,
		`    my $got = setlocale(LC_ALL, $name);`,
		`    print "$name: ", defined $got ? $got : "refused", "; numeric ", setlocale(LC_NUMERIC), "\n";`,
		`}`,
		`print "by position: ", setlocale(LC_ALL, "C.UTF-8;C;C;C;C;C") // "refused", "\n";`,
		`print "one category: ", setlocale(LC_CTYPE, "C") // "refused", " then ", setlocale(LC_ALL), "\n";`,
		`print "decimal point: ", localeconv()->{decimal_point}, "\n";`,
	)}, "")
	if r.stderr != "" {
		t.Errorf("stderr: %q", r.stderr)
	}
	r.want(t, 0, "after the refusals: C\n"+
		"C: C; numeric C\n"+
		"POSIX: C; numeric C\n"+
		"C.UTF-8: C.UTF-8;C;C;C;C;C; numeric C\n"+
		"C.utf8: C.UTF-8;C;C;C;C;C; numeric C\n"+
		"by position: C.UTF-8;C;C;C;C;C\n"+
		"one category: C then C\n"+
		"decimal point: .\n")
}

// The locale of the environment is not one the program asked for by name: a
// user whose environment names a locale the C library has no data for gets
// the locale it has for that character encoding, under its own name, and no
// complaint when the interpreter starts.
func TestLocaleOfTheEnvironment(t *testing.T) {
	for _, tc := range []struct {
		env  []string
		want string
	}{
		{[]string{"LANG=ja_JP.UTF-8"}, "all=C.UTF-8;C;C;C;C;C ctype=C.UTF-8 numeric=C codeset=UTF-8 utf8=1\n"},
		{[]string{"LANG=de_DE.utf8@euro"}, "all=C.UTF-8;C;C;C;C;C ctype=C.UTF-8 numeric=C codeset=UTF-8 utf8=1\n"},
		{[]string{"LANG=de_DE"}, "all=C ctype=C numeric=C codeset=ASCII utf8=0\n"},
		{[]string{"LANG=de_DE.ISO-8859-1"}, "all=C ctype=C numeric=C codeset=ASCII utf8=0\n"},
		{[]string{"LANG=C", "LC_ALL=en_US.UTF-8", "EXPECT_LC_ALL=1"}, "all=C.UTF-8;C;C;C;C;C ctype=C.UTF-8 numeric=C codeset=UTF-8 utf8=1\n"},
		{[]string{"LANG=en_US.UTF-8", "LC_CTYPE=C"}, "all=C ctype=C numeric=C codeset=ASCII utf8=0\n"},
		{[]string{"LANG=C", "LC_CTYPE=fr_FR.UTF-8"}, "all=C.UTF-8;C;C;C;C;C ctype=C.UTF-8 numeric=C codeset=UTF-8 utf8=1\n"},
		{[]string{"LANG=POSIX"}, "all=C ctype=C numeric=C codeset=ASCII utf8=0\n"},
		// An environment that names no locale means "C", and stays as it
		// is: no variable is added to it.
		{nil, "all=C ctype=C numeric=C codeset=ASCII utf8=0\n"},
		{[]string{"LANG="}, "all=C ctype=C numeric=C codeset=ASCII utf8=0\n"},
	} {
		r := runProgram(t, perl.Config{Env: tc.env}, perl.Program{Code: code(
			`use POSIX qw(setlocale LC_ALL LC_CTYPE LC_NUMERIC);`,
			`use I18N::Langinfo qw(langinfo CODESET);`,
			`my $all = setlocale(LC_ALL, "");`,
			`print "all=", $all // "refused", " ctype=", setlocale(LC_CTYPE), " numeric=", setlocale(LC_NUMERIC),`,
			`    " codeset=", langinfo(CODESET), " utf8=", (${^UTF8LOCALE} ? 1 : 0), "\n";`,
			`print "LC_ALL was added to the environment\n" if exists $ENV{LC_ALL} && !$ENV{EXPECT_LC_ALL};`,
		)}, "")
		if r.stderr != "" {
			t.Errorf("%v: stderr %q", tc.env, r.stderr)
		}
		if r.status != 0 || r.stdout != tc.want {
			t.Errorf("%v: status=%d\n got %q\nwant %q", tc.env, r.status, r.stdout, tc.want)
		}
	}
}
