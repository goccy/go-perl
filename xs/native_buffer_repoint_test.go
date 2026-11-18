//go:build darwin || linux

package xs_test

import "testing"

// Native code that points the buffer of an SV that is a string already at
// memory of its own. Encode's Unicode.xs moves the string of its result to a
// smaller block (Safefree of the old buffer, SvPV_set, then SvLEN_set, which
// hands the block over); Crypt::SecretBuffer's stringify scalar is pointed
// at the module's buffer (SvLEN 0) by its get magic on every read, at "" when
// there is no buffer, and at the new buffer when the module moves it. The
// interpreter refused each of those stores ("the string of a scalar cannot
// be in native memory"): the SV was a string while its slot held the
// transient GOPERL_PV_HOST. What is expected is what perl 5.36 on Linux gives
// with the same module.
func TestNativeBufferOfAStringPointedAgain(t *testing.T) {
	so := buildModuleSource(t, "RePoint", `#include "EXTERN.h"
#include "perl.h"
#include "XSUB.h"
#include <string.h>
/* Encode's Unicode.xs shrink_buffer: the string of a result moves to a
 * smaller block of the module's own, which SvLEN_set hands over */
XS(XS_RePoint_shrunk) {
    dXSARGS;
    STRLEN len;
    const char *s;
    SV *result;
    PERL_UNUSED_VAR(items);
    s = SvPV(ST(0), len);
    result = newSV(len + 100);
    sv_setpvn(result, s, len);
    if (SvLEN(result) > 42 + SvCUR(result)) {
        char *newpv;
        STRLEN newlen = 1 + SvCUR(result);
        New(0, newpv, newlen, char);
        Copy(SvPVX(result), newpv, newlen, char);
        Safefree(SvPVX(result));
        SvPV_set(result, newpv);
        SvLEN_set(result, newlen);
    }
    ST(0) = sv_2mortal(result);
    ST(1) = sv_2mortal(newSViv((IV)(SvLEN(result) - SvCUR(result))));
    XSRETURN(2);
}
/* Crypt::SecretBuffer's stringify scalar: get magic points it at the
 * module's buffer (SvLEN 0) on every read, or at "" without one; the
 * module grows the buffer (a new block, the old one wiped and freed) and
 * points the scalar at the new one, and sets its length when it appends */
typedef struct {
    char *data;
    size_t len, cap;
    SV *view;
} secret;
static void secret_realloc(secret *s, size_t cap) {
    char *old = s->data;
    if (cap) {
        Newxz(s->data, cap, char);
        if (old) memcpy(s->data, old, cap < s->cap ? cap : s->cap);
    } else {
        s->data = NULL;
    }
    if (old) {
        memset(old, 0, s->cap);
        Safefree(old);
    }
    s->cap = cap;
    if (s->len > cap) s->len = cap;
    if (s->view) {
        SvPVX(s->view) = s->data;
        SvCUR(s->view) = s->len;
    }
}
static int secret_view_get(pTHX_ SV *sv, MAGIC *mg) {
    secret *s = (secret *)mg->mg_ptr;
    if (s->data) {
        if (s->cap <= s->len) {
            secret_realloc(s, (s->len + 64) & ~(size_t)63);
            s->data[s->len] = 0;
        }
        SvPVX(sv) = s->data;
        SvCUR(sv) = s->len;
    } else {
        SvPVX(sv) = "";
        SvCUR(sv) = 0;
    }
    SvPOK_on(sv);
    SvUTF8_off(sv);
    SvREADONLY_on(sv);
    return 0;
}
static MGVTBL secret_view_vtbl = {secret_view_get, 0, 0, 0, 0};
static secret the_secret;
XS(XS_RePoint_view) {
    dXSARGS;
    secret *s = &the_secret;
    MAGIC mg;
    PERL_UNUSED_VAR(items);
    if (!s->view) {
        s->view = newSV(0);
        sv_magicext(s->view, NULL, PERL_MAGIC_ext, &secret_view_vtbl, (const char *)s, 0);
    }
    mg.mg_ptr = (char *)s;
    secret_view_get(aTHX_ s->view, &mg);
    ST(0) = s->view;
    XSRETURN(1);
}
XS(XS_RePoint_append) {
    dXSARGS;
    secret *s = &the_secret;
    STRLEN len;
    const char *p;
    PERL_UNUSED_VAR(items);
    p = SvPV(ST(0), len);
    if (s->len + len + 1 > s->cap) secret_realloc(s, (s->len + len + 1 + 63) & ~(size_t)63);
    memcpy(s->data + s->len, p, len);
    s->len += len;
    s->data[s->len] = 0;
    if (s->view) SvCUR(s->view) = s->len;
    XSRETURN(0);
}
XS(XS_RePoint_clear) {
    dXSARGS;
    PERL_UNUSED_VAR(items);
    the_secret.len = 0;
    secret_realloc(&the_secret, 0);
    XSRETURN(0);
}
XS_EXTERNAL(boot_RePoint) {
    dXSBOOTARGSXSAPIVERCHK;
    PERL_UNUSED_VAR(items);
    newXS("RePoint::shrunk", XS_RePoint_shrunk, __FILE__);
    newXS("RePoint::view", XS_RePoint_view, __FILE__);
    newXS("RePoint::append", XS_RePoint_append, __FILE__);
    newXS("RePoint::clear", XS_RePoint_clear, __FILE__);
    Perl_xs_boot_epilog(aTHX_ ax);
}
`)
	program := `my @r;
for my $n (1, 50, 300) {
    my ($s, $d) = RePoint::shrunk("x" x $n);
    push @r, length($s) . ":$d:" . ($s eq "x" x $n ? "same" : "differs");
}
my $bad = 0;
for my $i (1 .. 2000) {
    my ($s) = RePoint::shrunk("ab$i" x 3);
    $bad++ if $s ne "ab$i" x 3;
}
push @r, "bad=$bad";
my $v = \RePoint::view();
push @r, "[$$v]";
RePoint::append("hello");
push @r, "[$$v]";
RePoint::append(" world" x 20);
push @r, length("$$v") . ":" . substr("$$v", 0, 17);
for (1 .. 500) { my $t = "$$v"; $bad++ if length($t) != 125 }
push @r, "reads=$bad";
RePoint::clear();
push @r, "[$$v]";
RePoint::append("again");
push @r, "[$$v]";
join ",", @r`
	const want = "1:1:same,50:1:same,300:1:same,bad=0,[],[hello],125:hello world world,reads=0,[],[again]"
	if got := requireModule(t, "RePoint", so, program); got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}
