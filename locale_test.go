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
		"C.UTF-8: C.UTF-8; numeric C.UTF-8\n"+
		"C.utf8: C.UTF-8; numeric C.UTF-8\n"+
		"by position: C.UTF-8;C;C;C;C;C\n"+
		"one category: C then C\n"+
		"decimal point: .\n")
}

// The interpreter starts in the locale of the environment. A user whose
// environment names a locale the C library has no data for gets the locale
// it has for that character encoding, and no complaint at the start.
//
// A program that asks for the locale of the environment itself
// (setlocale(LC_ALL, "")) is answered like one that asks by name: the call
// fails when the environment names a locale there is no data for, and the
// locale stays as it was. Test suites find out that way whether a locale is
// installed (libintl-perl: `setlocale(LC_ALL, "") ? "" : "locale missing"`).
func TestLocaleOfTheEnvironment(t *testing.T) {
	const utf8, ctype, plain = "start=C.UTF-8", "start=C.UTF-8;C;C;C;C;C", "start=C"
	const others = "C;C.UTF-8;C.UTF-8;C.UTF-8;C.UTF-8;C.UTF-8"
	for _, tc := range []struct {
		env  []string
		want string
	}{
		{[]string{"LANG=ja_JP.UTF-8"}, utf8 + " asked=refused all=C.UTF-8 ctype=C.UTF-8 numeric=C.UTF-8 codeset=UTF-8 utf8=1\n"},
		{[]string{"LANG=de_DE.utf8@euro"}, utf8 + " asked=refused all=C.UTF-8 ctype=C.UTF-8 numeric=C.UTF-8 codeset=UTF-8 utf8=1\n"},
		{[]string{"LANG=de_DE"}, plain + " asked=refused all=C ctype=C numeric=C codeset=ASCII utf8=0\n"},
		{[]string{"LANG=de_DE.ISO-8859-1"}, plain + " asked=refused all=C ctype=C numeric=C codeset=ASCII utf8=0\n"},
		{[]string{"LANG=C", "LC_ALL=en_US.UTF-8", "EXPECT_LC_ALL=1"}, utf8 + " asked=refused all=C.UTF-8 ctype=C.UTF-8 numeric=C.UTF-8 codeset=UTF-8 utf8=1\n"},
		{[]string{"LANG=en_US.UTF-8", "LC_CTYPE=C"}, "start=" + others + " asked=refused all=" + others + " ctype=C numeric=C.UTF-8 codeset=ASCII utf8=0\n"},
		{[]string{"LANG=C", "LC_CTYPE=fr_FR.UTF-8"}, ctype + " asked=refused all=C.UTF-8;C;C;C;C;C ctype=C.UTF-8 numeric=C codeset=UTF-8 utf8=1\n"},
		// Locales there is data for are given to the program too.
		{[]string{"LANG=POSIX"}, plain + " asked=C all=C ctype=C numeric=C codeset=ASCII utf8=0\n"},
		{[]string{"LANG=C.UTF-8"}, utf8 + " asked=C.UTF-8 all=C.UTF-8 ctype=C.UTF-8 numeric=C.UTF-8 codeset=UTF-8 utf8=1\n"},
		{[]string{"LANG=de_DE", "LC_ALL=C.UTF-8", "EXPECT_LC_ALL=1"}, utf8 + " asked=C.UTF-8 all=C.UTF-8 ctype=C.UTF-8 numeric=C.UTF-8 codeset=UTF-8 utf8=1\n"},
		// An environment that names no locale means "C", and stays as it
		// is: no variable is added to it.
		{nil, plain + " asked=C all=C ctype=C numeric=C codeset=ASCII utf8=0\n"},
		{[]string{"LANG="}, plain + " asked=C all=C ctype=C numeric=C codeset=ASCII utf8=0\n"},
	} {
		r := runProgram(t, perl.Config{Env: tc.env}, perl.Program{Code: code(
			`use POSIX qw(setlocale LC_ALL LC_CTYPE LC_NUMERIC);`,
			`use I18N::Langinfo qw(langinfo CODESET);`,
			`my $start = setlocale(LC_ALL);`,
			`my $asked = setlocale(LC_ALL, "");`,
			`print "start=$start asked=", $asked // "refused", " all=", setlocale(LC_ALL), " ctype=", setlocale(LC_CTYPE),`,
			`    " numeric=", setlocale(LC_NUMERIC), " codeset=", langinfo(CODESET), " utf8=", (${^UTF8LOCALE} ? 1 : 0), "\n";`,
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

// A category set to C.UTF-8 reads as C.UTF-8, whichever it is, as on
// glibc: libintl-perl takes a category that reads as "C" for one without
// translations (setlocale(LC_MESSAGES, "C.UTF-8") and LANGUAGE give a
// program its translations).
func TestLocaleCategoryKeepsItsName(t *testing.T) {
	r := runProgram(t, perl.Config{Env: []string{"LANG=C"}}, perl.Program{Code: code(
		`use POSIX qw(setlocale LC_ALL LC_MESSAGES LC_NUMERIC);`,
		`print "messages: ", setlocale(LC_MESSAGES, "C.UTF-8") // "refused", ", reads ", setlocale(LC_MESSAGES), "\n";`,
		`print "numeric reads ", setlocale(LC_NUMERIC), "\n";`,
		`print "all reads ", setlocale(LC_ALL), "\n";`,
	)}, "")
	if r.stderr != "" {
		t.Errorf("stderr: %q", r.stderr)
	}
	r.want(t, 0, "messages: C.UTF-8, reads C.UTF-8\n"+
		"numeric reads C\n"+
		"all reads C;C;C;C;C;C.UTF-8\n")
}

// One category asked for from the environment is refused or given by
// itself, and a refusal changes nothing: not the categories that came
// before it in a request for all of them either.
func TestLocaleOfTheEnvironmentByCategory(t *testing.T) {
	r := runProgram(t, perl.Config{Env: []string{"LANG=C.UTF-8", "LC_NUMERIC=de_DE.UTF-8"}}, perl.Program{Code: code(
		`use POSIX qw(setlocale LC_ALL LC_CTYPE LC_NUMERIC);`,
		`setlocale(LC_ALL, "C");`,
		`print "numeric: ", setlocale(LC_NUMERIC, "") // "refused", "\n";`,
		`print "all: ", setlocale(LC_ALL, "") // "refused", ", is ", setlocale(LC_ALL), "\n";`,
		`print "ctype: ", setlocale(LC_CTYPE, "") // "refused", ", all is ", setlocale(LC_ALL), "\n";`,
	)}, "")
	if r.stderr != "" {
		t.Errorf("stderr: %q", r.stderr)
	}
	r.want(t, 0, "numeric: refused\n"+
		"all: refused, is C\n"+
		"ctype: C.UTF-8, all is C.UTF-8;C;C;C;C;C\n")
}
