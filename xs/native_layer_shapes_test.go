//go:build darwin || linux

package xs_test

import "testing"

// Layers of native code in the shapes of PerlIO::Util's :dir, :reverse
// and :tee: a layer that says it is buffered and keeps its buffer in its
// own struct, answering pointers into it from its Get_ptr and Get_base,
// with perl's Read (PerlIOBase_read through them) and a handle of its own;
// one that says it is buffered and keeps its buffer in a scalar while a
// scratch area of its struct holds what it read; and one that writes to a
// second handle and whose Dup pushes itself with a struct of its own as
// the mode, which its Pushed takes, and whose Getarg answers PerlIO_sv_dup.
// A push on a handle that is closed answers NULL, and vwarner gives a
// warning of a category. Before, the first two had their structs read as
// perl's buffer layer's, the Dup was looked for on the new handle, the
// mode reached the Pushed as a copy of the struct's bytes up to a zero, the
// push died, and vwarner and PerlIO_sv_dup were not there. What is
// expected is what perl 5.36 on Linux gives with the same module.
func TestNativeLayerShapesOfPerlIOUtil(t *testing.T) {
	so := buildModuleSource(t, "LayerShapes", `#include "EXTERN.h"
#include "perl.h"
#include "XSUB.h"
#include "perliol.h"
#include <stdarg.h>
#include <string.h>
/* Layers of native code in the shapes of PerlIO::Util's :dir, :reverse
 * and :tee: */

/* :inner(N), a reading layer that keeps its buffer in its own struct (as
 * :dir): it says it is buffered and has its own Get_base, Get_ptr, Get_cnt,
 * Bufsiz and Set_ptrcnt; its Read is perl's (NULL: PerlIOBase_read through
 * them); it opens a handle of its own with PerlIO_allocate; its Fill makes
 * the lines "line N" down to "line 1" */
typedef struct {
    PerlIOl base;
    IV left;
    STDCHAR buf[96];
    STDCHAR *ptr;
    STDCHAR *end;
} PerlIOInner;
static IV inner_pushed(pTHX_ PerlIO *f, const char *mode, SV *arg, PerlIO_funcs *tab) {
    PerlIOInner *s = PerlIOSelf(f, PerlIOInner);
    s->left = arg && SvOK(arg) ? SvIV(arg) : 0;
    s->ptr = s->end = s->buf;
    return PerlIOBase_pushed(aTHX_ f, mode, arg, tab);
}
static PerlIO *inner_open(pTHX_ PerlIO_funcs *self, PerlIO_list_t *layers, IV n, const char *mode,
                          int fd, int imode, int perm, PerlIO *f, int narg, SV **args) {
    PERL_UNUSED_ARG(layers);
    PERL_UNUSED_ARG(n);
    PERL_UNUSED_ARG(fd);
    PERL_UNUSED_ARG(imode);
    PERL_UNUSED_ARG(perm);
    if (narg != 1) {
        SETERRNO(EINVAL, LIB_INVARG);
        return NULL;
    }
    if (PerlIOValid(f))
        PerlIO_close(f);
    else
        f = PerlIO_allocate(aTHX);
    return PerlIO_push(aTHX_ f, self, mode, args[0]);
}
static IV inner_fill(pTHX_ PerlIO *f) {
    PerlIOInner *s = PerlIOSelf(f, PerlIOInner);
    if (s->left <= 0 || s->left > 9) {
        PerlIOBase(f)->flags |= PERLIO_F_EOF;
        s->ptr = s->end = s->buf;
        return -1;
    }
    memcpy(s->buf, "line ", 5);
    s->buf[5] = (STDCHAR)('0' + s->left--);
    s->buf[6] = '\n';
    s->ptr = s->buf;
    s->end = s->buf + 7;
    PerlIOBase(f)->flags |= PERLIO_F_RDBUF;
    return 0;
}
static STDCHAR *inner_get_base(pTHX_ PerlIO *f) { return PerlIOSelf(f, PerlIOInner)->buf; }
static STDCHAR *inner_get_ptr(pTHX_ PerlIO *f) { return PerlIOSelf(f, PerlIOInner)->ptr; }
static SSize_t inner_get_cnt(pTHX_ PerlIO *f) {
    return PerlIOSelf(f, PerlIOInner)->end - PerlIOSelf(f, PerlIOInner)->ptr;
}
static Size_t inner_bufsiz(pTHX_ PerlIO *f) {
    return PerlIOSelf(f, PerlIOInner)->end - PerlIOSelf(f, PerlIOInner)->buf;
}
static void inner_set_ptrcnt(pTHX_ PerlIO *f, STDCHAR *ptr, SSize_t cnt) {
    PERL_UNUSED_ARG(cnt);
    PerlIOSelf(f, PerlIOInner)->ptr = ptr;
}
static PERLIO_FUNCS_DECL(inner_funcs) = {
    sizeof(PerlIO_funcs), "inner", sizeof(PerlIOInner), PERLIO_K_BUFFERED | PERLIO_K_RAW,
    inner_pushed, NULL, inner_open, PerlIOBase_binmode, NULL,
    NULL, NULL, NULL, NULL, NULL,
    NULL, NULL, NULL, NULL, inner_fill,
    NULL, NULL, NULL, NULL,
    inner_get_base, inner_bufsiz, inner_get_ptr, inner_get_cnt, inner_set_ptrcnt,
};

/* :upper, a reading layer over another that keeps its buffer in a scalar
 * (as :reverse): it says it is buffered, reads into a scratch area of its
 * own struct and gives perl the upper-cased bytes from the scalar's
 * buffer */
typedef struct {
    PerlIOl base;
    STDCHAR scratch[96];
    SV *bufsv;
    STDCHAR *ptr;
    STDCHAR *end;
} PerlIOUpper;
static IV upper_pushed(pTHX_ PerlIO *f, const char *mode, SV *arg, PerlIO_funcs *tab) {
    PerlIOUpper *s = PerlIOSelf(f, PerlIOUpper);
    s->bufsv = newSV(64);
    sv_setpvs(s->bufsv, "");
    s->ptr = s->end = SvPVX(s->bufsv);
    return PerlIOBase_pushed(aTHX_ f, mode, arg, tab);
}
static IV upper_popped(pTHX_ PerlIO *f) {
    PerlIOUpper *s = PerlIOSelf(f, PerlIOUpper);
    SvREFCNT_dec(s->bufsv);
    s->bufsv = NULL;
    return PerlIOBase_popped(aTHX_ f);
}
static IV upper_fill(pTHX_ PerlIO *f) {
    PerlIOUpper *s = PerlIOSelf(f, PerlIOUpper);
    SSize_t n = PerlIO_read(PerlIONext(f), s->scratch, 5), i;
    if (n <= 0) {
        PerlIOBase(f)->flags |= n < 0 ? PERLIO_F_ERROR : PERLIO_F_EOF;
        return -1;
    }
    for (i = 0; i < n; i++)
        if (s->scratch[i] >= 'a' && s->scratch[i] <= 'z') s->scratch[i] -= 'a' - 'A';
    sv_setpvn(s->bufsv, (const char *)s->scratch, (STRLEN)n);
    s->ptr = SvPVX(s->bufsv);
    s->end = s->ptr + n;
    PerlIOBase(f)->flags |= PERLIO_F_RDBUF;
    return 0;
}
static STDCHAR *upper_get_base(pTHX_ PerlIO *f) { return SvPVX(PerlIOSelf(f, PerlIOUpper)->bufsv); }
static STDCHAR *upper_get_ptr(pTHX_ PerlIO *f) { return PerlIOSelf(f, PerlIOUpper)->ptr; }
static SSize_t upper_get_cnt(pTHX_ PerlIO *f) {
    return PerlIOSelf(f, PerlIOUpper)->end - PerlIOSelf(f, PerlIOUpper)->ptr;
}
static Size_t upper_bufsiz(pTHX_ PerlIO *f) { return SvCUR(PerlIOSelf(f, PerlIOUpper)->bufsv); }
static void upper_set_ptrcnt(pTHX_ PerlIO *f, STDCHAR *ptr, SSize_t cnt) {
    PERL_UNUSED_ARG(cnt);
    PerlIOSelf(f, PerlIOUpper)->ptr = ptr;
}
static PERLIO_FUNCS_DECL(upper_funcs) = {
    sizeof(PerlIO_funcs), "upper", sizeof(PerlIOUpper), PERLIO_K_BUFFERED | PERLIO_K_RAW,
    upper_pushed, upper_popped, NULL, PerlIOBase_binmode, NULL,
    NULL, NULL, NULL, NULL, NULL,
    NULL, NULL, NULL, NULL, upper_fill,
    NULL, NULL, NULL, NULL,
    upper_get_base, upper_bufsiz, upper_get_ptr, upper_get_cnt, upper_set_ptrcnt,
};

/* :pair(PATH), a writing layer that writes to a second handle too (as
 * :tee): its Dup duplicates the layers below and the second handle and
 * pushes itself on the new handle with a struct of its own as the mode,
 * which its Pushed takes when there is no argument; its Getarg answers
 * PerlIO_sv_dup of its argument */
typedef struct {
    PerlIOl base;
    SV *arg;
    PerlIO *out;
} PerlIOPair;
static IV pair_pushed(pTHX_ PerlIO *f, const char *mode, SV *arg, PerlIO_funcs *tab) {
    PerlIOPair *s = PerlIOSelf(f, PerlIOPair);
    const PerlIOPair *proto = (mode && !arg) ? (const PerlIOPair *)mode : NULL;
    PERL_UNUSED_ARG(tab);
    if (!PerlIOValid(PerlIONext(f))) {
        SETERRNO(EBADF, SS_IVCHAN);
        return -1;
    }
    if (proto) {
        s->arg = proto->arg;
        s->out = proto->out;
    } else {
        if (!arg || !SvOK(arg)) {
            SETERRNO(EINVAL, LIB_INVARG);
            return -1;
        }
        s->arg = newSVsv(arg);
        s->out = PerlIO_openn(aTHX_ NULL, "a", -1, 0, 0, NULL, 1, &s->arg);
    }
    if (!s->out) return -1;
    PerlIOBase(f)->flags = PerlIOBase(PerlIONext(f))->flags;
    return 0;
}
static IV pair_popped(pTHX_ PerlIO *f) {
    PerlIOPair *s = PerlIOSelf(f, PerlIOPair);
    if (s->out) PerlIO_close(s->out);
    s->out = NULL;
    SvREFCNT_dec(s->arg);
    s->arg = NULL;
    return 0;
}
static SSize_t pair_write(pTHX_ PerlIO *f, const void *vbuf, Size_t count) {
    PerlIOPair *s = PerlIOSelf(f, PerlIOPair);
    if (PerlIO_write(s->out, vbuf, count) != (SSize_t)count) return -1;
    return PerlIO_write(PerlIONext(f), vbuf, count);
}
static IV pair_flush(pTHX_ PerlIO *f) {
    PerlIOPair *s = PerlIOSelf(f, PerlIOPair);
    if (s->out && PerlIO_flush(s->out) != 0) return -1;
    return PerlIO_flush(PerlIONext(f));
}
static SV *pair_getarg(pTHX_ PerlIO *f, CLONE_PARAMS *param, int flags) {
    PERL_UNUSED_ARG(flags);
    return PerlIO_sv_dup(aTHX_ PerlIOSelf(f, PerlIOPair)->arg, param);
}
/* a handle's top layer's Dup (perl's PerlIOBase_dup when it has none) on
 * a handle given or a new one, as :tee duplicates */
static PerlIO *layer_dup(pTHX_ PerlIO *newfp, PerlIO *oldfp, CLONE_PARAMS *param, int flags) {
    PerlIO *(*dup)(pTHX_ PerlIO *, PerlIO *, CLONE_PARAMS *, int);
    if (!PerlIOValid(oldfp)) {
        SETERRNO(EBADF, SS_IVCHAN);
        return NULL;
    }
    dup = PerlIOBase(oldfp)->tab->Dup;
    if (!newfp) newfp = PerlIO_allocate(aTHX);
    if (!dup) dup = PerlIOBase_dup;
    return dup(aTHX_ newfp, oldfp, param, flags);
}
static PerlIO *pair_dup(pTHX_ PerlIO *f, PerlIO *o, CLONE_PARAMS *param, int flags) {
    f = layer_dup(aTHX_ f, PerlIONext(o), param, flags);
    if (f) {
        PerlIOPair proto;
        proto.arg = PerlIO_sv_dup(aTHX_ PerlIOSelf(o, PerlIOPair)->arg, param);
        proto.out = layer_dup(aTHX_ NULL, PerlIOSelf(o, PerlIOPair)->out, param, flags);
        f = PerlIO_push(aTHX_ f, PerlIOBase(o)->tab, (const char *)&proto, NULL);
    }
    return f;
}
static PERLIO_FUNCS_DECL(pair_funcs) = {
    sizeof(PerlIO_funcs), "pair", sizeof(PerlIOPair), PERLIO_K_RAW | PERLIO_K_MULTIARG,
    pair_pushed, pair_popped, NULL, PerlIOBase_binmode, pair_getarg,
    PerlIOBase_fileno, pair_dup, NULL, NULL, pair_write,
    NULL, NULL, PerlIOBase_close, pair_flush, NULL,
    PerlIOBase_eof, PerlIOBase_error, PerlIOBase_clearerr, PerlIOBase_setlinebuf,
    NULL, NULL, NULL, NULL, NULL,
};

/* a push on no handle (an IO without one): PerlIO_push answers NULL */
XS(XS_LayerShapes_push_none) {
    dXSARGS;
    PerlIO *f;
    PERL_UNUSED_VAR(items);
    f = IoIFP(sv_2io(ST(0)));
    ST(0) = sv_2mortal(newSVpv(PerlIO_push(aTHX_ f, PERLIO_FUNCS_CAST(&upper_funcs), "r", NULL)
                                   ? "pushed" : "null", 0));
    XSRETURN(1);
}
/* vwarner, as PerlIO::Util's warnif gives it its arguments */
static void warn_layer(pTHX_ const char *fmt, ...) {
    va_list args;
    va_start(args, fmt);
    vwarner(packWARN(WARN_LAYER), fmt, &args);
    va_end(args);
}
XS(XS_LayerShapes_warn) {
    dXSARGS;
    PERL_UNUSED_VAR(items);
    warn_layer(aTHX_ "layer %s says %d", SvPV_nolen(ST(0)), 42);
    XSRETURN(0);
}
XS_EXTERNAL(boot_LayerShapes) {
    dXSBOOTARGSXSAPIVERCHK;
    PERL_UNUSED_VAR(items);
    newXS("LayerShapes::push_none", XS_LayerShapes_push_none, __FILE__);
    newXS("LayerShapes::warn", XS_LayerShapes_warn, __FILE__);
    PerlIO_define_layer(aTHX_ PERLIO_FUNCS_CAST(&inner_funcs));
    PerlIO_define_layer(aTHX_ PERLIO_FUNCS_CAST(&upper_funcs));
    PerlIO_define_layer(aTHX_ PERLIO_FUNCS_CAST(&pair_funcs));
    Perl_xs_boot_epilog(aTHX_ ax);
}
`)
	program := `require File::Temp;
require IO::Handle;
my @r;
open my $in, "<:inner", 3 or die "open inner: $!";
my @lines = <$in>;
push @r, join("+", PerlIO::get_layers($in)), scalar(@lines), join("", @lines);
close $in;
my ($t1, $src) = File::Temp::tempfile(UNLINK => 1);
print $t1 "hello, upper world\n";
close $t1;
open my $up, "<", $src or die;
binmode $up, ":upper" or die "binmode upper: $!";
push @r, join("+", PerlIO::get_layers($up)), scalar(<$up>);
close $up;
my ($t2, $main) = File::Temp::tempfile(UNLINK => 1);
close $t2;
my ($t3, $side) = File::Temp::tempfile(UNLINK => 1);
close $t3;
open my $p, ">:pair($side)", $main or die "open pair: $!";
print $p "direct\n";
$p->flush;
open my $d, ">&", $p or die "dup: $!";
my @dl = PerlIO::get_layers($d, details => 1);
(my $dls = join("+", PerlIO::get_layers($d))) =~ s/\Q$side\E/SIDE/;
push @r, $dls, $dl[-3], ($dl[-2] eq $side ? "same arg" : "arg $dl[-2]");
print $d "via dup\n";
close $d;
print $p "direct again\n";
close $p;
for my $f ($main, $side) {
    open my $i, "<", $f or die;
    push @r, join("", <$i>);
}
open my $closed, "<", $src or die;
close $closed;
push @r, LayerShapes::push_none($closed);
my @w;
{
    local $SIG{__WARN__} = sub { push @w, $_[0] };
    LayerShapes::warn("inner");
}
(my $w = $w[0]) =~ s/ at .*//s;
push @r, scalar(@w), $w;
my $r = join "|", @r;
$r =~ s/\n/\\n/g;
$r`
	const want = "inner|3|line 3\\nline 2\\nline 1\\n|unix+perlio+upper|HELLO, UPPER WORLD\\n|unix+perlio+pair(SIDE)|pair|same arg|direct\\nvia dup\\ndirect again\\n|direct\\nvia dup\\ndirect again\\n|null|1|layer inner says 42"
	if got := requireModule(t, "LayerShapes", so, program); got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}
