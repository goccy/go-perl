//go:build darwin || linux

package xs_test

import "testing"

// A buffered layer of native code in the shape of PerlIO::eol: its own
// Get_base allocates the buffer with Newxz (the buffer moves into the
// interpreter's memory, which frees it in perl's PerlIOBuf_close), its Open
// calls perl's PerlIOBuf_open and keeps PerlIOArg in its own struct, its
// Write doubles every byte through perl's PerlIOBuf_write, its Fill fills
// with perl's PerlIOBuf_fill and works on the buffer, its Clearerr resets its
// own counter before perl's PerlIOBase_clearerr, and its Popped frees what
// it kept. Three rounds read the file, so that a buffer freed twice or a
// struct left behind would show. Before, PerlIOArg did not exist and a
// layer could customize only Pushed and Fill. What is expected is what perl
// 5.36 on Linux gives with the same module.
func TestNativeLayerOwnBufferAndOpen(t *testing.T) {
	so := buildModuleSource(t, "TwiceLayer", `#include "EXTERN.h"
#include "perl.h"
#include "XSUB.h"
#include "perliol.h"
#include <string.h>
/* :twice(NAME), a buffered layer in the shape of PerlIO::eol: its own
 * Get_base allocates the buffer (twice the size perl's would be), its Open
 * calls perl's PerlIOBuf_open and then keeps the layer's argument
 * (PerlIOArg) in its own struct, its Write doubles every byte through
 * perl's PerlIOBuf_write, its Fill fills with perl's PerlIOBuf_fill and
 * keeps every other byte, its Clearerr resets its own counters before
 * perl's, and its Popped frees what it kept */
typedef struct {
    PerlIOBuf base;
    char *name;
    IV fills;
} PerlIOTwice;
static IV twice_pushed(pTHX_ PerlIO *f, const char *mode, SV *arg, PerlIO_funcs *tab) {
    if (PerlIOBase(PerlIONext(f))->flags & PERLIO_F_UTF8)
        PerlIOBase(f)->flags |= PERLIO_F_UTF8;
    else
        PerlIOBase(f)->flags &= ~PERLIO_F_UTF8;
    return PerlIOBuf_pushed(aTHX_ f, mode, arg, tab);
}
static STDCHAR *twice_get_base(pTHX_ PerlIO *f) {
    PerlIOBuf *b = PerlIOSelf(f, PerlIOBuf);
    if (!b->buf) {
        if (!b->bufsiz) b->bufsiz = 1024;
        Newxz(b->buf, b->bufsiz * 2, STDCHAR);
        b->ptr = b->buf;
        b->end = b->ptr;
    }
    return b->buf;
}
static SSize_t twice_write(pTHX_ PerlIO *f, const void *vbuf, Size_t count) {
    const STDCHAR *in = (const STDCHAR *)vbuf;
    STDCHAR *two;
    Size_t i;
    SSize_t done;
    Newx(two, count * 2 + 1, STDCHAR);
    for (i = 0; i < count; i++) two[2 * i] = two[2 * i + 1] = in[i];
    done = PerlIOBuf_write(aTHX_ f, two, count * 2);
    Safefree(two);
    return done < 0 ? done : done / 2;
}
static IV twice_fill(pTHX_ PerlIO *f) {
    IV code = PerlIOBuf_fill(aTHX_ f);
    PerlIOBuf *b = PerlIOSelf(f, PerlIOBuf);
    STDCHAR *p, *q;
    if (code) return code;
    for (p = q = b->ptr; p + 1 < b->end; p += 2) *q++ = *p;
    b->end = q;
    PerlIOSelf(f, PerlIOTwice)->fills++;
    return 0;
}
static void twice_clearerr(pTHX_ PerlIO *f) {
    if (PerlIOValid(f)) PerlIOSelf(f, PerlIOTwice)->fills = 0;
    PerlIOBase_clearerr(aTHX_ f);
}
static SV *twice_getarg(pTHX_ PerlIO *f, CLONE_PARAMS *param, int flags) {
    PerlIOTwice *t = PerlIOSelf(f, PerlIOTwice);
    PERL_UNUSED_ARG(param);
    PERL_UNUSED_ARG(flags);
    return newSVpvf("%s:%d", t->name ? t->name : "-", (int)t->fills);
}
static PerlIO *twice_open(pTHX_ PerlIO_funcs *self, PerlIO_list_t *layers, IV n, const char *mode,
                          int fd, int imode, int perm, PerlIO *old, int narg, SV **args) {
    SV *arg = PerlIOArg;
    PerlIO *f = PerlIOBuf_open(aTHX_ self, layers, n, mode, fd, imode, perm, old, narg, args);
    if (f) PerlIOSelf(f, PerlIOTwice)->name = savepv(arg && SvOK(arg) ? SvPV_nolen(arg) : "none");
    return f;
}
static IV twice_popped(pTHX_ PerlIO *f) {
    PerlIOTwice *t = PerlIOSelf(f, PerlIOTwice);
    Safefree(t->name);
    t->name = NULL;
    return PerlIOBuf_popped(aTHX_ f);
}
static PERLIO_FUNCS_DECL(twice_funcs) = {
    sizeof(PerlIO_funcs), "twice", sizeof(PerlIOTwice), PERLIO_K_BUFFERED | PERLIO_K_UTF8,
    twice_pushed, twice_popped, twice_open, PerlIOBase_binmode, twice_getarg,
    PerlIOBase_fileno, PerlIOBuf_dup, PerlIOBuf_read, PerlIOBuf_unread, twice_write,
    PerlIOBuf_seek, PerlIOBuf_tell, PerlIOBuf_close, PerlIOBuf_flush, twice_fill,
    PerlIOBase_eof, PerlIOBase_error, twice_clearerr, PerlIOBase_setlinebuf,
    twice_get_base, PerlIOBuf_bufsiz, PerlIOBuf_get_ptr, PerlIOBuf_get_cnt,
    PerlIOBuf_set_ptrcnt,
};
XS_EXTERNAL(boot_TwiceLayer) {
    dXSBOOTARGSXSAPIVERCHK;
    PERL_UNUSED_VAR(items);
    PerlIO_define_layer(aTHX_ PERLIO_FUNCS_CAST(&twice_funcs));
    Perl_xs_boot_epilog(aTHX_ ax);
}
`)
	program := `require File::Temp;
my ($tmp, $name) = File::Temp::tempfile(UNLINK => 1);
close $tmp;
my @r;
open my $o, ">:raw:twice(out)", $name or die "open >: $!";
print $o "ab\n";
print $o "x" x 3000, "\n";
push @r, join("+", PerlIO::get_layers($o));
close $o or die "close: $!";
push @r, -s $name;
open my $raw, "<:raw", $name or die;
push @r, scalar(<$raw>);
close $raw;
for my $round (1 .. 3) {
    open my $i, "<:raw:twice(in$round)", $name or die "open <: $!";
    my @lines = <$i>;
    push @r, scalar(@lines), $lines[0], length($lines[1]);
    push @r, join("+", PerlIO::get_layers($i));
    push @r, eof($i) ? "eof" : "more";
    $i->clearerr;
    push @r, join("+", PerlIO::get_layers($i));
    close $i;
}
my $s = join "|", @r;
$s =~ s/\n/\\n/g;
$s`
	const want = "unix+perlio+twice(out:0)|6008|aabb\\n|2|ab\\n|3001|unix+perlio+twice(in1:6)|eof|unix+perlio+twice(in1:0)|2|ab\\n|3001|unix+perlio+twice(in2:6)|eof|unix+perlio+twice(in2:0)|2|ab\\n|3001|unix+perlio+twice(in3:6)|eof|unix+perlio+twice(in3:0)"
	if got := requireModule(t, "TwiceLayer", so, program); got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}
