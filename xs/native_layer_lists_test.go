//go:build darwin || linux

package xs_test

import "testing"

// The interpreter's lists of layers, handles and open flags as layers of
// native code use them, in the shapes of PerlIO::fse, PerlIO::code and
// PerlIO::Util: a pseudo-layer whose Open walks the list of layers it gets
// (cur, array[i].funcs), opens through the last layer below it that has an
// Open and applies the ones above with PerlIO_apply_layera; a buffered
// layer that opens a handle of its own with PerlIO_allocate and pushes
// itself; PL_known_layers and PL_def_layerlist walked as perl's lists;
// PerlIOUnix_oflags compared with this system's O_ values; the flags the
// Open of the walking layer gets, as this system's; PerlIO_openn
// of a sysopen mode and do_open of sysopen given this system's flags, and
// PerlIO_fdupopen; and a layer in the shape of :tee that keeps a second
// handle in its struct (opened with PerlIO_openn, or the IoOFP of another
// handle), writes to it from its Write slot, and is walked from an XSUB
// in a later call. Before, each of those names stopped the build, and the
// open flags crossed as the other side's bits. What is expected is what
// perl 5.36 on Linux gives with the same module.
func TestNativeLayerListsAndKeptHandles(t *testing.T) {
	so := buildModuleSource(t, "LayerLists", `#include "EXTERN.h"
#include "perl.h"
#include "XSUB.h"
#include "perliol.h"
#include <fcntl.h>
#include <string.h>
/* The interpreter's lists of layers and handles as native code uses them
 * (the shapes of PerlIO::Util, PerlIO::fse and PerlIO::code): */

/* :walk, a pseudo-layer whose Open walks the list of layers it got
 * (layers->array[i].funcs, ->cur), notes the open flags it got (as this
 * system's), opens through the last one below it that has an Open, and
 * applies the ones above that with PerlIO_apply_layera */
static char walked[256];
static char walk_flags[64];
static IV walk_pushed(pTHX_ PerlIO *f, const char *mode, SV *arg, PerlIO_funcs *tab) {
    PERL_UNUSED_ARG(f);
    PERL_UNUSED_ARG(mode);
    PERL_UNUSED_ARG(arg);
    PERL_UNUSED_ARG(tab);
    return 0;
}
static PerlIO *walk_open(pTHX_ PerlIO_funcs *self, PerlIO_list_t *layers, IV n, const char *mode,
                         int fd, int imode, int perm, PerlIO *f, int narg, SV **args) {
    PerlIO_funcs *tab = NULL;
    IV i = n;
    strcpy(walk_flags, (imode & O_ACCMODE) == O_RDWR     ? "RDWR"
                       : (imode & O_ACCMODE) == O_WRONLY ? "WRONLY"
                                                         : "RDONLY");
    if (imode & O_CREAT) strcat(walk_flags, "|CREAT");
    if (imode & O_TRUNC) strcat(walk_flags, "|TRUNC");
    if (imode & O_APPEND) strcat(walk_flags, "|APPEND");
    walked[0] = 0;
    for (i = 0; i < layers->cur; i++) {
        if (i) strcat(walked, ",");
        strcat(walked, layers->array[i].funcs == self ? "SELF" : layers->array[i].funcs->name);
    }
    i = n;
    while (--i >= 0) {
        tab = layers->array[i].funcs;
        if (tab && tab->Open) break;
    }
    if (!tab || !tab->Open) return NULL;
    f = tab->Open(aTHX_ tab, layers, i, mode, fd, imode, perm, f, narg, args);
    if (f && ++i < n && i != n) {
        /* the layers after the one that opened, up to this one */
        if (PerlIO_apply_layera(aTHX_ f, mode, layers, i, n) != 0) {
            PerlIO_close(f);
            f = NULL;
        }
    }
    return f;
}
static PERLIO_FUNCS_DECL(walk_funcs) = {
    sizeof(PerlIO_funcs), "walk", 0, PERLIO_K_UTF8 | PERLIO_K_MULTIARG, walk_pushed, NULL, walk_open,
};

/* :src(TEXT), a buffered layer at the bottom of a handle of its own
 * (PerlIO_allocate), serving TEXT from its own Fill */
typedef struct {
    PerlIOBuf base;
    SV *text;
    STRLEN pos;
} PerlIOSrc;
static IV src_pushed(pTHX_ PerlIO *f, const char *mode, SV *arg, PerlIO_funcs *tab) {
    PerlIOSrc *s = PerlIOSelf(f, PerlIOSrc);
    IV code = PerlIOBuf_pushed(aTHX_ f, mode, arg, tab);
    if (code) return code;
    s->text = newSVsv(arg && SvOK(arg) ? arg : &PL_sv_no);
    s->pos = 0;
    return 0;
}
static IV src_popped(pTHX_ PerlIO *f) {
    PerlIOSrc *s = PerlIOSelf(f, PerlIOSrc);
    SvREFCNT_dec(s->text);
    s->text = NULL;
    return PerlIOBuf_popped(aTHX_ f);
}
static IV src_fill(pTHX_ PerlIO *f) {
    PerlIOSrc *s = PerlIOSelf(f, PerlIOSrc);
    PerlIOBuf *b = &s->base;
    STRLEN len;
    const char *p = SvPV(s->text, len);
    STRLEN n = len - s->pos;
    if (!b->buf) PerlIO_get_base(f);
    b->ptr = b->end = b->buf;
    if (!n) {
        PerlIOBase(f)->flags |= PERLIO_F_EOF;
        return -1;
    }
    if (n > 7) n = 7; /* small fills */
    Copy(p + s->pos, b->buf, n, STDCHAR);
    s->pos += n;
    b->end = b->buf + n;
    PerlIOBase(f)->flags |= PERLIO_F_RDBUF;
    return 0;
}
static PerlIO *src_open(pTHX_ PerlIO_funcs *self, PerlIO_list_t *layers, IV n, const char *mode,
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
    if (f)
        PerlIO_close(f);
    else
        f = PerlIO_allocate(aTHX);
    if ((f = PerlIO_push(aTHX_ f, self, mode, args[0]))) PerlIOBase(f)->flags |= PERLIO_F_OPEN;
    return f;
}
static PERLIO_FUNCS_DECL(src_funcs) = {
    sizeof(PerlIO_funcs), "src", sizeof(PerlIOSrc), PERLIO_K_BUFFERED | PERLIO_K_RAW,
    src_pushed, src_popped, src_open, PerlIOBase_binmode, NULL,
    PerlIOBase_fileno, NULL, PerlIOBuf_read, PerlIOBuf_unread, NULL,
    NULL, PerlIOBuf_tell, PerlIOBuf_close, PerlIOBuf_flush, src_fill,
    PerlIOBase_eof, PerlIOBase_error, PerlIOBase_clearerr, PerlIOBase_setlinebuf,
    PerlIOBuf_get_base, PerlIOBuf_bufsiz, PerlIOBuf_get_ptr, PerlIOBuf_get_cnt,
    PerlIOBuf_set_ptrcnt,
};

/* :keep(TARGET), a layer in the shape of PerlIO::Util's :tee: it keeps a
 * second handle in its struct (one it opens itself with PerlIO_openn for a
 * path, or the IoOFP of a handle it is given), writes to it from its Write
 * slot, and closes the one it opened when it is popped */
typedef struct {
    PerlIOl base;
    PerlIO *out;
    int own;
} PerlIOKeep;
static IV keep_pushed(pTHX_ PerlIO *f, const char *mode, SV *arg, PerlIO_funcs *tab) {
    PerlIOKeep *k = PerlIOSelf(f, PerlIOKeep);
    if (!arg || !SvOK(arg)) {
        SETERRNO(EINVAL, LIB_INVARG);
        return -1;
    }
    if (SvROK(arg)) {
        k->out = IoOFP(sv_2io(arg));
        k->own = 0;
    } else {
        k->out = PerlIO_openn(aTHX_ NULL, "w", -1, 0, 0, NULL, 1, &arg);
        k->own = 1;
    }
    if (!k->out) return -1;
    return PerlIOBase_pushed(aTHX_ f, mode, arg, tab);
}
static IV keep_popped(pTHX_ PerlIO *f) {
    PerlIOKeep *k = PerlIOSelf(f, PerlIOKeep);
    if (k->own && k->out) PerlIO_close(k->out);
    k->out = NULL;
    return PerlIOBase_popped(aTHX_ f);
}
static SSize_t keep_write(pTHX_ PerlIO *f, const void *vbuf, Size_t count) {
    PerlIOKeep *k = PerlIOSelf(f, PerlIOKeep);
    if (k->out && PerlIO_write(k->out, vbuf, count) != (SSize_t)count) return -1;
    return PerlIO_write(PerlIONext(f), vbuf, count);
}
static IV keep_flush(pTHX_ PerlIO *f) {
    PerlIOKeep *k = PerlIOSelf(f, PerlIOKeep);
    if (k->out && PerlIO_flush(k->out) != 0) return -1;
    return PerlIO_flush(PerlIONext(f));
}
static PERLIO_FUNCS_DECL(keep_funcs) = {
    sizeof(PerlIO_funcs), "keep", sizeof(PerlIOKeep), PERLIO_K_RAW,
    keep_pushed, keep_popped, NULL, PerlIOBase_binmode, NULL,
    PerlIOBase_fileno, NULL, NULL, NULL, keep_write,
    NULL, NULL, PerlIOBase_close, keep_flush, NULL,
    PerlIOBase_eof, PerlIOBase_error, PerlIOBase_clearerr, PerlIOBase_setlinebuf,
    NULL, NULL, NULL, NULL, NULL,
};
XS(XS_LayerLists_push_keep) {
    dXSARGS;
    PerlIO *f;
    PERL_UNUSED_VAR(items);
    f = IoOFP(sv_2io(ST(0)));
    if (!PerlIO_push(aTHX_ f, PERLIO_FUNCS_CAST(&keep_funcs), "w", ST(1)))
        croak("push: %s", strerror(errno));
    XSRETURN(0);
}
/* the handle a :keep layer kept, in a later call: its layers, whether it
 * can write, and whether it has a descriptor */
XS(XS_LayerLists_kept) {
    dXSARGS;
    PerlIO *f, *out;
    SV *r;
    PERL_UNUSED_VAR(items);
    f = IoOFP(sv_2io(ST(0)));
    if (!PerlIOValid(f) || strNE(PerlIOBase(f)->tab->name, "keep")) croak("no :keep on top");
    out = PerlIOSelf(f, PerlIOKeep)->out;
    r = newSVpvs("");
    if (!PerlIOValid(out)) {
        sv_catpvs(r, "invalid");
    } else {
        PerlIO *l;
        for (l = out; PerlIOValid(l); l = PerlIONext(l))
            sv_catpvf(r, "%s%s", l == out ? "" : "+", PerlIOBase(l)->tab->name);
        sv_catpvf(r, ":%s:%s", PerlIOBase(out)->flags & PERLIO_F_CANWRITE ? "w" : "-",
                  PerlIO_fileno(out) >= 0 ? "fd" : "nofd");
    }
    ST(0) = sv_2mortal(r);
    XSRETURN(1);
}
XS(XS_LayerLists_walked) {
    dXSARGS;
    PERL_UNUSED_VAR(items);
    ST(0) = sv_2mortal(newSVpvf("%s:%s", walked, walk_flags));
    XSRETURN(1);
}
/* which of some layers PL_known_layers has, and what PL_def_layerlist is */
XS(XS_LayerLists_known) {
    dXSARGS;
    static const char *const names[] = {"unix", "perlio", "crlf", "utf8", "bytes", "raw", "pop",
                                        "walk", "src", "nosuch"};
    SV *out = newSVpvs("");
    IV i;
    size_t k;
    PERL_UNUSED_VAR(items);
    for (k = 0; k < sizeof names / sizeof names[0]; k++) {
        int found = 0;
        for (i = 0; i < PL_known_layers->cur; i++)
            if (strEQ(PL_known_layers->array[i].funcs->name, names[k])) found = 1;
        sv_catpvf(out, "%s%s=%d", k ? "," : "", names[k], found);
    }
    sv_catpvs(out, ";");
    for (i = 0; i < PL_def_layerlist->cur; i++)
        sv_catpvf(out, "%s%s", i ? "," : "", PL_def_layerlist->array[i].funcs->name);
    ST(0) = sv_2mortal(out);
    XSRETURN(1);
}
/* PerlIOUnix_oflags as this system's open flags */
XS(XS_LayerLists_oflags) {
    dXSARGS;
    int fl;
    SV *out;
    PERL_UNUSED_VAR(items);
    fl = PerlIOUnix_oflags(SvPV_nolen(ST(0)));
    out = newSVpv((fl & O_ACCMODE) == O_RDWR ? "RDWR" : (fl & O_ACCMODE) == O_WRONLY ? "WRONLY" : "RDONLY", 0);
    if (fl & O_CREAT) sv_catpvs(out, "|CREAT");
    if (fl & O_TRUNC) sv_catpvs(out, "|TRUNC");
    if (fl & O_APPEND) sv_catpvs(out, "|APPEND");
    if (fl & O_EXCL) sv_catpvs(out, "|EXCL");
    ST(0) = sv_2mortal(out);
    XSRETURN(1);
}
/* a file written through PerlIO_openn with this system's open flags (a
 * sysopen: the mode starts with '#'), then appended to through do_open as
 * sysopen does, through a duplicate of that handle (PerlIO_fdupopen), and
 * through PerlIO_openn with :walk on top, whose Open gets the flags */
XS(XS_LayerLists_write) {
    dXSARGS;
    SV *path;
    PerlIO *f, *d;
    GV *gv;
    PERL_UNUSED_VAR(items);
    path = ST(0);
    f = PerlIO_openn(aTHX_ NULL, "#w", -1, O_WRONLY | O_CREAT | O_TRUNC, 0644, NULL, 1, &path);
    if (!f) croak("openn: %s", strerror(errno));
    PerlIO_write(f, "one\n", 4);
    PerlIO_close(f);
    gv = gv_fetchpvs("LayerLists::OUT", GV_ADD, SVt_PVIO);
    if (!do_open(gv, SvPV_nolen(path), (I32)SvCUR(path), TRUE, O_WRONLY | O_APPEND, 0644, NULL))
        croak("do_open: %s", strerror(errno));
    f = IoOFP(GvIOp(gv));
    PerlIO_write(f, "two\n", 4);
    PerlIO_flush(f);
    d = PerlIO_fdupopen(aTHX_ f, NULL, 0);
    if (!d) croak("fdupopen: %s", strerror(errno));
    PerlIO_write(d, "three\n", 6);
    PerlIO_close(d);
    f = PerlIO_openn(aTHX_ ":raw:perlio:walk", "#a", -1, O_WRONLY | O_APPEND, 0644, NULL, 1, &path);
    if (!f) croak("openn walk: %s", strerror(errno));
    PerlIO_write(f, "four\n", 5);
    PerlIO_close(f);
    XSRETURN(0);
}
XS_EXTERNAL(boot_LayerLists) {
    dXSBOOTARGSXSAPIVERCHK;
    PERL_UNUSED_VAR(items);
    newXS("LayerLists::walked", XS_LayerLists_walked, __FILE__);
    newXS("LayerLists::known", XS_LayerLists_known, __FILE__);
    newXS("LayerLists::oflags", XS_LayerLists_oflags, __FILE__);
    newXS("LayerLists::write", XS_LayerLists_write, __FILE__);
    newXS("LayerLists::push_keep", XS_LayerLists_push_keep, __FILE__);
    newXS("LayerLists::kept", XS_LayerLists_kept, __FILE__);
    PerlIO_define_layer(aTHX_ PERLIO_FUNCS_CAST(&walk_funcs));
    PerlIO_define_layer(aTHX_ PERLIO_FUNCS_CAST(&src_funcs));
    PerlIO_define_layer(aTHX_ PERLIO_FUNCS_CAST(&keep_funcs));
    Perl_xs_boot_epilog(aTHX_ ax);
}
`)
	program := `require File::Temp;
my ($tmp, $name) = File::Temp::tempfile(UNLINK => 1);
print $tmp "line one\nline two\n";
close $tmp;
my @r;
open my $w, "<:raw:perlio:walk", $name or die "open walk: $!";
push @r, LayerLists::walked(), join("+", PerlIO::get_layers($w)), scalar(<$w>);
close $w;
open my $s, "<:src", "served by\na layer\n" or die "open src: $!";
my @lines = <$s>;
push @r, join("+", PerlIO::get_layers($s)), scalar(@lines), $lines[1];
close $s;
push @r, LayerLists::known();
push @r, map { "$_=" . LayerLists::oflags($_) } "r", "r+", "w", "a+";
my ($t2, $out) = File::Temp::tempfile(UNLINK => 1);
close $t2;
LayerLists::write($out);
close LayerLists::OUT;
push @r, LayerLists::walked();
open my $in, "<", $out or die;
push @r, join("", <$in>), sprintf("%o", (stat $out)[2] & 0777);
close $in;
my ($t3, $copy) = File::Temp::tempfile(UNLINK => 1);
close $t3;
my ($t4, $main) = File::Temp::tempfile(UNLINK => 1);
close $t4;
my ($t5, $sidep) = File::Temp::tempfile(UNLINK => 1);
close $t5;
open my $o, ">", $main or die;
LayerLists::push_keep($o, $copy);
open my $side, ">", $sidep or die;
open my $o2, ">>", $main or die;
LayerLists::push_keep($o2, $side);
print $o "to both\n";
print $o2 "and to the side\n";
push @r, join("+", PerlIO::get_layers($o)), LayerLists::kept($o), LayerLists::kept($o2);
close $o;
close $o2;
close $side;
for my $p ($main, $copy, $sidep) {
    open my $i, "<", $p or die;
    push @r, join("", <$i>);
}
my $r = join "|", @r;
$r =~ s/\n/\\n/g;
$r`
	const want = "unix,perlio,raw,perlio,SELF:RDONLY|unix+perlio+perlio|line one\\n|src|2|a layer\\n|unix=1,perlio=1,crlf=1,utf8=1,bytes=1,raw=1,pop=1,walk=1,src=1,nosuch=0;unix,perlio|r=RDONLY|r+=RDWR|w=WRONLY|CREAT|TRUNC|a+=RDWR|CREAT|APPEND|unix,perlio,raw,perlio,SELF:WRONLY|APPEND|one\\ntwo\\nthree\\nfour\\n|600|unix+perlio+keep|perlio+unix:w:fd|perlio+unix:w:fd|to both\\nand to the side\\n|to both\\n|and to the side\\n"
	if got := requireModule(t, "LayerLists", so, program); got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}
