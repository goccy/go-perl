//go:build darwin || linux

package xs_test

import (
	"bytes"
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	perl "github.com/goccy/go-perl"
	goperlfs "github.com/goccy/go-perl/fs"
)

// Native code runs in the locale the program's environment names, as under
// perl, which sets it when it starts (setlocale(LC_ALL, "")) and keeps
// LC_NUMERIC "C" for itself: GLib converts the arguments it parses from the
// locale's character set, Text::CharWidth measures a character with mbtowc
// and wcwidth (Unicode-Block). A locale the system has no data for falls
// back as perl's start-up does: to the locale LC_ALL names, then LANG,
// then "C" (perl 5.36 reports the same in each case). Before, the C
// library of the process stayed in "C" whatever the environment said:
// GLib's option parser failed ("Invalid byte sequence in conversion
// input") and U+3111 measured 1. The locale is the process's and is set
// with the first native module, so each case runs in a process of its own.
const localeModule = `#include "EXTERN.h"
#include "perl.h"
#include "XSUB.h"
#include <locale.h>
#include <langinfo.h>
#include <stdlib.h>
#include <wchar.h>
XS(XS_MyLocale_report) {
    dXSARGS;
    const char *numeric = setlocale(LC_NUMERIC, NULL);
    const char *codeset = nl_langinfo(CODESET);
    /* U+3111 BOPOMOFO LETTER ZH, a wide character */
    const char bytes[] = "\xe3\x84\x91";
    wchar_t wc;
    int width = mbtowc(&wc, bytes, 3) == 3 ? wcwidth(wc) : -2;
    PERL_UNUSED_VAR(items);
    ST(0) = sv_2mortal(newSVpvf("numeric %s, codeset %s, %s, width %d", numeric ? numeric : "?",
                                strcmp(codeset, "UTF-8") == 0 ? "UTF-8" : "other",
                                MB_CUR_MAX > 1 ? "multibyte" : "single byte", width));
    XSRETURN(1);
}
XS_EXTERNAL(boot_MyLocale) {
    dXSBOOTARGSXSAPIVERCHK;
    newXS_deffile("MyLocale::report", XS_MyLocale_report);
    Perl_xs_boot_epilog(aTHX_ ax);
}
`

func TestNativeCodeRunsInTheLocaleOfTheEnvironment(t *testing.T) {
	if os.Getenv("GOPERL_LOCALE_CHILD") != "" {
		localeChild(t)
		return
	}
	so := buildModuleSource(t, "MyLocale", localeModule)
	for _, tc := range []struct{ name, env, want string }{
		{"LANG of UTF-8", "LANG=C.UTF-8", "numeric C, codeset UTF-8, multibyte, width 2"},
		{"LC_ALL over LANG", "LC_ALL=C.UTF-8|LANG=C", "numeric C, codeset UTF-8, multibyte, width 2"},
		{"LC_NUMERIC stays C", "LANG=C.UTF-8|LC_NUMERIC=C.UTF-8", "numeric C, codeset UTF-8, multibyte, width 2"},
		{"no locale named", "", "numeric C, codeset other, single byte, width -2"},
		{"a locale the system lacks", "LANG=xx_XX.UTF-8", "numeric C, codeset other, single byte, width -2"},
		{"a category the system lacks falls back to LANG", "LC_TIME=xx_XX.UTF-8|LANG=C.UTF-8", "numeric C, codeset UTF-8, multibyte, width 2"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			cmd := exec.Command(os.Args[0], "-test.run=^TestNativeCodeRunsInTheLocaleOfTheEnvironment$")
			cmd.Env = append(os.Environ(), "GOPERL_LOCALE_CHILD="+so, "GOPERL_LOCALE_ENV="+tc.env)
			var out, errOut bytes.Buffer
			cmd.Stdout, cmd.Stderr = &out, &errOut
			err := cmd.Run()
			got := out.String()
			if i := strings.Index(got, "result: "); i >= 0 {
				got = got[i+len("result: "):]
				if j := strings.Index(got, "\n"); j >= 0 {
					got = got[:j]
				}
			}
			if err != nil || got != tc.want {
				t.Fatalf("err=%v got %q, want %q\nstdout %q\nstderr %q", err, got, tc.want, out.String(), errOut.String())
			}
		})
	}
}

// localeChild loads the module into an interpreter whose environment is
// GOPERL_LOCALE_ENV ("|"-separated), the first native module of the
// process, and prints what it reports.
func localeChild(t *testing.T) {
	stdlib, err := perl.ExtractStdlib()
	if err != nil {
		t.Fatalf("extract stdlib: %v", err)
	}
	var env []string
	if e := os.Getenv("GOPERL_LOCALE_ENV"); e != "" {
		env = strings.Split(e, "|")
	}
	p, err := perl.New(perl.Config{FS: goperlfs.NewHostFS(), StdlibDir: stdlib, NativeModules: true, Env: env})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	defer p.Close()
	libdir := t.TempDir()
	auto := filepath.Join(libdir, "auto", "MyLocale")
	if err := os.MkdirAll(auto, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(os.Getenv("GOPERL_LOCALE_CHILD"), filepath.Join(auto, "MyLocale.so")); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(libdir, "MyLocale.pm"),
		[]byte("package MyLocale;\nuse XSLoader;\nXSLoader::load(__PACKAGE__);\n1;\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	r, err := p.Eval(context.Background(), `unshift @INC, '`+libdir+`'; require MyLocale; MyLocale::report()`)
	if err != nil {
		t.Fatalf("Eval: %v", err)
	}
	if r.Error != nil {
		t.Fatalf("Eval died: %v", r.Error)
	}
	os.Stdout.WriteString("result: " + resultStr(r) + "\n")
}
