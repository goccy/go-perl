//go:build darwin || linux

package xs_test

import "testing"

// A buffered layer of native code in the shape of PerlIO::gzip and
// PerlIO::subfile: its own Pushed (perl's PerlIOBuf_pushed first, an
// autopop with the argument "pop", a :perlio pushed under it when the layer
// below is not buffered), Popped (frees its own state, pops the :perlio it
// pushed), Fill (reads the buffer of the layer below in place), Flush
// (writes through the layer below), Close (its own flush through the handle,
// then perl's PerlIOBuf_close), Getarg (get_layers, and the argument a dup
// pushes the layer with) and Tell (perl's PerlIOBuf_tell less where the
// layer started). An XSUB that looks at the layer's struct while Perl code
// it calls reads the same handle leaves the reading where Perl code left it.
// Before, a layer could customize only Pushed and Fill, and
// PerlIO_push, PerlIO_pop and PerlIO_perlio stopped the build ("go-perl
// gives native code no access to this interpreter internal"). What is
// expected is what perl 5.36 on Linux gives with the same module.
func TestNativeLayerOwnSlotsAndStack(t *testing.T) {
	so := buildModuleSource(t, "RotLayer", `#include "EXTERN.h"
#include "perl.h"
#include "XSUB.h"
#include "perliol.h"
#include <stdlib.h>
#include <string.h>
/* :rot(N), a buffered layer in the shape of PerlIO::gzip and
 * PerlIO::subfile: its own Pushed (perl's PerlIOBuf_pushed first; "pop" pops
 * the layer again at once; a layer below that is not buffered gets a
 * :perlio pushed under this one), Popped (frees what Pushed allocated, pops
 * the :perlio it pushed), Fill (reads the buffer of the layer below in
 * place), Flush (writes through the layer below), Close (its own flush
 * through the handle, then perl's PerlIOBuf_close), Getarg and Tell (perl's
 * PerlIOBuf_tell less where the layer started) */
typedef struct {
    PerlIOBuf base;
    int shift;
    int pushed_below;
    Off_t start;
    char *scratch;
} PerlIORot;
static STDCHAR rot_char(STDCHAR c, int shift) {
    if (c >= 'a' && c <= 'z') return (STDCHAR)('a' + (c - 'a' + shift) % 26);
    if (c >= 'A' && c <= 'Z') return (STDCHAR)('A' + (c - 'A' + shift) % 26);
    return c;
}
static IV rot_pushed(pTHX_ PerlIO *f, const char *mode, SV *arg, PerlIO_funcs *tab) {
    PerlIORot *r = PerlIOSelf(f, PerlIORot);
    STRLEN len;
    const char *a = (arg && SvOK(arg)) ? SvPV(arg, len) : "13";
    IV code = PerlIOBuf_pushed(aTHX_ f, mode, &PL_sv_undef, tab);
    if (code) return code;
    if (strEQ(a, "pop")) {
        PerlIO_pop(aTHX_ f);
        return 0;
    }
    r->shift = atoi(a) % 26;
    r->start = PerlIOBuf_tell(aTHX_ f);
    Newx(r->scratch, 16, char);
    strcpy(r->scratch, "scratch");
    if (PerlIOBase(f)->flags & PERLIO_F_CANREAD) {
        PerlIO *below = PerlIONext(f);
        if (!(PerlIOBase(below)->flags & PERLIO_F_FASTGETS)) {
            if (!PerlIO_push(aTHX_ below, &PerlIO_perlio, "r", &PL_sv_undef)) return -1;
            r->pushed_below = 1;
        }
    }
    return 0;
}
static IV rot_popped(pTHX_ PerlIO *f) {
    PerlIORot *r = PerlIOSelf(f, PerlIORot);
    Safefree(r->scratch);
    r->scratch = NULL;
    if (r->pushed_below) {
        r->pushed_below = 0;
        PerlIO_pop(aTHX_ PerlIONext(f));
    }
    return PerlIOBuf_popped(aTHX_ f);
}
static IV rot_fill(pTHX_ PerlIO *f) {
    PerlIORot *r = PerlIOSelf(f, PerlIORot);
    PerlIOBuf *b = &r->base;
    PerlIO *n = PerlIONext(f);
    SSize_t avail, i;
    STDCHAR *p;
    if (!b->buf) PerlIO_get_base(f);
    b->ptr = b->end = b->buf;
    avail = PerlIO_get_cnt(n);
    if (avail <= 0) {
        if (PerlIO_fill(n) != 0) {
            PerlIOBase(f)->flags |= PerlIO_error(n) ? PERLIO_F_ERROR : PERLIO_F_EOF;
            return -1;
        }
        avail = PerlIO_get_cnt(n);
    }
    if (avail > (SSize_t)b->bufsiz) avail = (SSize_t)b->bufsiz;
    p = PerlIO_get_ptr(n);
    for (i = 0; i < avail; i++) b->buf[i] = rot_char(p[i], r->shift);
    PerlIO_set_ptrcnt(n, p + avail, PerlIO_get_cnt(n) - avail);
    b->end = b->buf + avail;
    PerlIOBase(f)->flags |= PERLIO_F_RDBUF;
    return 0;
}
static IV rot_flush(pTHX_ PerlIO *f) {
    PerlIORot *r = PerlIOSelf(f, PerlIORot);
    PerlIOBuf *b = &r->base;
    if (PerlIOBase(f)->flags & PERLIO_F_WRBUF) {
        STDCHAR *p;
        SSize_t n = b->ptr - b->buf;
        for (p = b->buf; p < b->ptr; p++) *p = rot_char(*p, r->shift);
        if (PerlIO_write(PerlIONext(f), b->buf, n) != n) return -1;
        b->ptr = b->end = b->buf;
        PerlIOBase(f)->flags &= ~PERLIO_F_WRBUF;
        return PerlIO_flush(PerlIONext(f));
    }
    return PerlIOBuf_flush(aTHX_ f);
}
static IV rot_close(pTHX_ PerlIO *f) {
    IV code = PerlIO_flush(f);
    code |= PerlIOBuf_close(aTHX_ f);
    return code ? -1 : 0;
}
static SV *rot_getarg(pTHX_ PerlIO *f, CLONE_PARAMS *param, int flags) {
    PERL_UNUSED_ARG(param);
    PERL_UNUSED_ARG(flags);
    return newSVpvf("%d", PerlIOSelf(f, PerlIORot)->shift);
}
static Off_t rot_tell(pTHX_ PerlIO *f) {
    return PerlIOBuf_tell(aTHX_ f) - PerlIOSelf(f, PerlIORot)->start;
}
static PERLIO_FUNCS_DECL(rot_funcs) = {
    sizeof(PerlIO_funcs), "rot", sizeof(PerlIORot), PERLIO_K_BUFFERED,
    rot_pushed, rot_popped, PerlIOBuf_open, PerlIOBase_binmode, rot_getarg,
    PerlIOBase_fileno, PerlIOBuf_dup, PerlIOBuf_read, PerlIOBuf_unread, PerlIOBuf_write,
    PerlIOBuf_seek, rot_tell, rot_close, rot_flush, rot_fill,
    PerlIOBase_eof, PerlIOBase_error, PerlIOBase_clearerr, PerlIOBase_setlinebuf,
    PerlIOBuf_get_base, PerlIOBuf_bufsiz, PerlIOBuf_get_ptr, PerlIOBuf_get_cnt,
    PerlIOBuf_set_ptrcnt,
};
/* peek(fh, cb): the top layer of the handle as native code sees it (its
 * table's name and the struct's shift), with Perl code reading the same
 * handle in between (cb) */
XS(XS_RotLayer_peek) {
    dXSARGS;
    PerlIO *f;
    const char *name;
    int shift;
    PERL_UNUSED_VAR(items);
    f = IoIFP(sv_2io(ST(0)));
    name = PerlIOBase(f)->tab->name;
    shift = strEQ(name, "rot") ? PerlIOSelf(f, PerlIORot)->shift : -1;
    PUSHMARK(SP);
    PUTBACK;
    call_sv(ST(1), G_DISCARD | G_NOARGS);
    SPAGAIN;
    ST(0) = sv_2mortal(newSVpvf("%s:%d", name, shift));
    XSRETURN(1);
}
XS_EXTERNAL(boot_RotLayer) {
    dXSBOOTARGSXSAPIVERCHK;
    PERL_UNUSED_VAR(items);
    newXS("RotLayer::peek", XS_RotLayer_peek, __FILE__);
    PerlIO_define_layer(aTHX_ PERLIO_FUNCS_CAST(&rot_funcs));
    Perl_xs_boot_epilog(aTHX_ ax);
}
`)
	program := `require File::Temp;
my ($tmp, $name) = File::Temp::tempfile(UNLINK => 1);
close $tmp;
my @r;
open my $o, ">:raw:rot(1)", $name or die "open >: $!";
print $o "Hello, World\n" x 3;
print $o "abc" x 5000, "\n";
close $o or die "close: $!";
open my $raw, "<:raw", $name or die;
my $line = <$raw>;
close $raw;
push @r, $line;
open my $i, "<:raw:rot(25)", $name or die "open <: $!";
my @lines = <$i>;
push @r, scalar(@lines), $lines[0], length($lines[3]), substr($lines[3], 0, 6);
close $i;
open my $u, "<:unix:rot(25)", $name or die "open unix: $!";
push @r, join("+", PerlIO::get_layers($u));
push @r, scalar(<$u>), tell($u);
open my $d, "<&", $u or die "dup: $!";
push @r, join("+", map { defined $_ ? $_ : "-" } PerlIO::get_layers($d, details => 1));
close $d;
close $u;
open my $k, "<:raw:rot(25)", $name or die "open peek: $!";
my $first = <$k>;
my $second;
push @r, RotLayer::peek($k, sub { $second = <$k> });
my $third = <$k>;
push @r, join("/", map { substr($_, 0, 5) } $first, $second, $third), tell($k);
close $k;
open my $p, "<:raw:rot(pop)", $name or die "open pop: $!";
push @r, join("+", PerlIO::get_layers($p)), scalar(<$p>);
close $p;
binmode STDOUT;
my $s = join "|", @r;
$s =~ s/\n/\\n/g;
$s`
	const want = "Ifmmp, Xpsme\\n|4|Hello, World\\n|15001|abcabc|unix+perlio+rot(25)|Hello, World\\n|13|unix+-+2098176+perlio+-+4195328+rot+25+4195328|rot:25|Hello/Hello/Hello|39|unix+perlio|Ifmmp, Xpsme\\n"
	if got := requireModule(t, "RotLayer", so, program); got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}
