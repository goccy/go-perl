//go:build darwin || linux

package xs_test

import "testing"

// Layers of native code with no struct of their own (size 0), in the
// shapes of PerlIO::locale and PerlIO::if: one pushes perl's :encoding in
// its Pushed (PerlIO_find_layer with STR_WITH_LEN as PerlIO::bom calls it,
// PerlIO_push) and opens through
// PerlIOBase_open; the other walks the handle's layers (their tables' kinds,
// their flags, PerlIONext), applies layers when its test holds, and opens by
// calling the Open of the layer below it in the list (PerlIO_layer_fetch,
// the table's Open slot) and pushing itself with PerlIOArg, and sets errno
// when its argument is missing ($! after the open). Before, each of those
// names stopped the build. What is expected is what perl 5.36 on
// Linux gives with the same module.
func TestNativePseudoLayers(t *testing.T) {
	so := buildModuleSource(t, "PseudoLayers", `#include "EXTERN.h"
#include "perl.h"
#include "XSUB.h"
#include "perliol.h"
#include <string.h>
/* Two layers with no struct of their own (size 0), in the shapes of
 * PerlIO::locale and PerlIO::if. :utf8enc pushes perl's :encoding(UTF-8)
 * in its Pushed (PerlIO_find_layer, PerlIO_push), and opens through
 * PerlIOBase_open. :when(TEST,LAYERS) applies LAYERS when TEST holds for the
 * handle's layers (walked through PerlIONext: their tables' kinds and their
 * flags), and opens by calling the Open of the layer below it in the list
 * (PerlIO_layer_fetch) and pushing itself with its argument (PerlIOArg). */
static IV utf8enc_pushed(pTHX_ PerlIO *f, const char *mode, SV *arg, PerlIO_funcs *tab) {
    PerlIO_funcs *encoding = PerlIO_find_layer(aTHX_ STR_WITH_LEN("encoding"), 1);
    SV *name = sv_2mortal(newSVpvs("UTF-8"));
    PERL_UNUSED_ARG(arg);
    PERL_UNUSED_ARG(tab);
    return PerlIO_push(aTHX_ f, encoding, mode, name) == f ? 0 : -1;
}
static PERLIO_FUNCS_DECL(utf8enc_funcs) = {
    sizeof(PerlIO_funcs), "utf8enc", 0, 0, utf8enc_pushed, NULL, PerlIOBase_open,
};
static int when_buffered(pTHX_ PerlIO *f) {
    PerlIO *t = f;
    const PerlIOl *l;
    while (t && (l = *t)) {
        if (l->tab && l->tab->kind & PERLIO_K_BUFFERED && !(l->flags & PERLIO_F_UNBUF)) return 1;
        t = PerlIONext(t);
    }
    return 0;
}
static int when_crlf(pTHX_ PerlIO *f) {
    PerlIO *t = f;
    const PerlIOl *l;
    while (t && (l = *t)) {
        if (l->tab && l->tab->kind & PERLIO_K_CANCRLF && l->flags & PERLIO_F_CRLF) return 1;
        t = PerlIONext(t);
    }
    return 0;
}
static IV when_pushed(pTHX_ PerlIO *f, const char *mode, SV *arg, PerlIO_funcs *tab) {
    const char *argstr, *delim, *layer;
    int negate, hold;
    PERL_UNUSED_ARG(tab);
    if (!PerlIOValid(f) || !arg || !SvOK(arg)) {
        SETERRNO(EINVAL, LIB_INVARG);
        return -1;
    }
    argstr = SvPV_nolen(arg);
    delim = strchr(argstr, ',');
    if (!delim) {
        SETERRNO(EINVAL, LIB_INVARG);
        return -1;
    }
    negate = argstr[0] == '!';
    if (negate) argstr++;
    layer = delim + 1;
    if (delim - argstr == 8 && memEQ(argstr, "buffered", 8))
        hold = when_buffered(aTHX_ f);
    else if (delim - argstr == 4 && memEQ(argstr, "crlf", 4))
        hold = when_crlf(aTHX_ f);
    else
        Perl_croak(aTHX_ "No such test '%.*s' known", (int)(delim - argstr), argstr);
    if (hold != negate) PerlIO_apply_layers(aTHX_ f, mode, layer);
    return 0;
}
static PerlIO *when_open(pTHX_ PerlIO_funcs *self, PerlIO_list_t *layers, IV n, const char *mode,
                         int fd, int imode, int perm, PerlIO *old, int narg, SV **args) {
    PerlIO_funcs *const tab = PerlIO_layer_fetch(aTHX_ layers, n - 1, NULL);
    if (tab && tab->Open) {
        PerlIO *ret = (*tab->Open)(aTHX_ tab, layers, n - 1, mode, fd, imode, perm, old, narg, args);
        if (ret && PerlIO_push(aTHX_ ret, self, mode, PerlIOArg) == NULL) {
            PerlIO_close(ret);
            return NULL;
        }
        return ret;
    }
    SETERRNO(EINVAL, LIB_INVARG);
    return NULL;
}
static PERLIO_FUNCS_DECL(when_funcs) = {
    sizeof(PerlIO_funcs), "when", 0, PERLIO_K_UTF8 | PERLIO_K_MULTIARG, when_pushed, NULL,
    when_open,
};
XS_EXTERNAL(boot_PseudoLayers) {
    dXSBOOTARGSXSAPIVERCHK;
    PERL_UNUSED_VAR(items);
    PerlIO_define_layer(aTHX_ PERLIO_FUNCS_CAST(&utf8enc_funcs));
    PerlIO_define_layer(aTHX_ PERLIO_FUNCS_CAST(&when_funcs));
    Perl_xs_boot_epilog(aTHX_ ax);
}
`)
	program := `require File::Temp;
my ($tmp, $name) = File::Temp::tempfile(UNLINK => 1);
binmode $tmp;
print $tmp "caf\xc3\xa9\r\nline2\r\n";
close $tmp;
my @r;
open my $e, "<:utf8enc", $name or die "open utf8enc: $!";
my $l = <$e>;
push @r, join("+", PerlIO::get_layers($e)), length($l), utf8::is_utf8($l) ? "chars" : "bytes";
close $e;
open my $w, "<:when(buffered,:crlf)", $name or die "open when: $!";
push @r, join("+", PerlIO::get_layers($w)), length(scalar <$w>);
close $w;
open my $n, "<:raw:when(!buffered,:crlf)", $name or die "open when not: $!";
push @r, join("+", PerlIO::get_layers($n)), length(scalar <$n>);
close $n;
open my $b, "<:raw", $name or die;
binmode $b, ":when(buffered,:utf8enc)" or die "binmode: $!";
push @r, join("+", PerlIO::get_layers($b)), length(scalar <$b>);
close $b;
push @r, eval { open my $x, "<:when(odd,:crlf)", $name; 1 } ? "opened" : $@ =~ /No such test 'odd' known/ ? "refused" : $@;
require Errno;
$! = 0;
push @r, (open my $y, "<:when", $name) ? "opened" : ($! == Errno::EINVAL() ? "EINVAL" : 0 + $!);
my $s = join "|", @r;
$s =~ s/\n/\\n/g;
$s =~ s/\r/\\r/g;
$s`
	const want = "unix+perlio+encoding(utf-8-strict)+utf8|6|chars|unix+perlio+crlf|6|unix+perlio|7|unix+perlio+encoding(utf-8-strict)+utf8|6|refused|EINVAL"
	if got := requireModule(t, "PseudoLayers", so, program); got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}
