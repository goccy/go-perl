//go:build darwin || linux

package xs_test

import "testing"

// Native code that compiles a pattern sets its character set in the flags
// it hands CALLREGCOMP with set_regex_charset (Sereal::Decoder does so for
// the /d, /l, /u and /a modifiers it reads), and reads it back with
// get_regex_charset. The field lies where the interpreter has it: before,
// the macros shifted the set to another bit, and /u came out as /p.
const regexCharsetSource = `#include "EXTERN.h"
#include "perl.h"
#include "XSUB.h"
/* comp(pattern, modifiers): a qr// object compiled with the flags the
 * modifiers name, and the character set its flags say */
XS(XS_MyRc_comp) {
    dXSARGS;
    SV *pat = ST(0);
    const char *mods = SvPV_nolen(ST(1));
    U32 flags = 0;
    int as = 0;
    REGEXP *rx;
    SV *rv;
    PERL_UNUSED_VAR(items);
    for (; *mods; mods++) {
        switch (*mods) {
        case 'd': set_regex_charset(&flags, REGEX_DEPENDS_CHARSET); break;
        case 'l': set_regex_charset(&flags, REGEX_LOCALE_CHARSET); break;
        case 'u': set_regex_charset(&flags, REGEX_UNICODE_CHARSET); break;
        case 'a':
            set_regex_charset(&flags, ++as == 2 ? REGEX_ASCII_MORE_RESTRICTED_CHARSET
                                                : REGEX_ASCII_RESTRICTED_CHARSET);
            break;
        case 'i': flags |= PMf_FOLD; break;
        }
    }
    rx = CALLREGCOMP(pat, flags);
    rv = sv_bless(newRV_noinc((SV *)rx), gv_stashpvs("Regexp", 0));
    ST(0) = sv_2mortal(rv);
    ST(1) = sv_2mortal(newSViv((IV)get_regex_charset(flags)));
    XSRETURN(2);
}
XS_EXTERNAL(boot_MyRc) {
    dXSBOOTARGSXSAPIVERCHK;
    newXS_deffile("MyRc::comp", XS_MyRc_comp);
    Perl_xs_boot_epilog(aTHX_ ax);
}
`

func TestNativeCodeSetsARegexCharset(t *testing.T) {
	so := buildModuleSource(t, "MyRc", regexCharsetSource)
	got := requireModule(t, "MyRc", so, `use re qw(regexp_pattern);
my @r;
for my $case (["\\w", "u", "\xe9"], ["\\w", "a", "\xe9"], ["\\w", "d", "\xe9"],
              ["foo\xdf", "ui", "FOOSS"], ["k", "aai", "\x{212a}"], ["\\w", "l", "a"]) {
    my ($pat, $mods, $text) = @$case;
    my ($re, $cs) = MyRc::comp($pat, $mods);
    my (undef, $flags) = regexp_pattern($re);
    push @r, sprintf "%s:%s cs %d %s", $mods, join("", sort split //, $flags), $cs,
        ($text =~ $re) ? "matches" : "no";
}
join "; ", @r`)
	// perl: the set is the pattern's (/u matches é as \w, /a and /d do
	// not; /ui folds ß to SS; /aai keeps KELVIN SIGN from k)
	want := "u:u cs 2 matches; a:a cs 3 no; d: cs 0 no; ui:iu cs 2 matches; aai:aai cs 4 no; l:l cs 1 matches"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}
