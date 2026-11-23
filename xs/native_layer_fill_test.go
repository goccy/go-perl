//go:build darwin || linux

package xs_test

import "testing"

// A buffered layer of native code in the shape of PerlIO::unicodeeol: it
// looks for PERLIO_LAYERS before it registers, names :perlio's unread in
// its table, takes the UTF-8 flag of the layer below in Pushed, and fills
// with :perlio's fill in its own Fill before it works on what was read.
// Each of those failed: PERLIO_LAYERS was not defined, PerlIOBuf_unread was
// no name the SDK gave, the layer below showed no flags, and PerlIOBuf_fill
// filled nothing. What is expected is what perl 5.36 on Linux gives with the
// same module.
func TestNativeLayerFillsWithPerlioFill(t *testing.T) {
	so := buildModuleSource(t, "Upper", `#include "EXTERN.h"
#include "perl.h"
#include "XSUB.h"
#include "perlio.h"
#include "perliol.h"
/* :upper, a buffered layer in the shape of PerlIO::unicodeeol: Pushed takes
 * the UTF-8 flag of the layer below, Fill fills with :perlio's fill and
 * then works on what it read, the table names :perlio's unread */
static IV upper_pushed(pTHX_ PerlIO *f, const char *mode, SV *arg, PerlIO_funcs *tab) {
    if (PerlIOBase(PerlIONext(f))->flags & PERLIO_F_UTF8)
        PerlIOBase(f)->flags |= PERLIO_F_UTF8;
    else
        PerlIOBase(f)->flags &= ~PERLIO_F_UTF8;
    return PerlIOBuf_pushed(aTHX_ f, mode, arg, tab);
}
static IV upper_fill(pTHX_ PerlIO *f) {
    IV code = PerlIOBuf_fill(aTHX_ f);
    PerlIOBuf *b = PerlIOSelf(f, PerlIOBuf);
    STDCHAR *p;
    if (code) return code;
    for (p = b->ptr; p < b->end; p++)
        if (*p >= 'a' && *p <= 'z') *p -= 'a' - 'A';
    return 0;
}
static PERLIO_FUNCS_DECL(upper_funcs) = {
    sizeof(PerlIO_funcs), "upper", sizeof(PerlIOBuf), PERLIO_K_BUFFERED | PERLIO_K_UTF8,
    upper_pushed, PerlIOBuf_popped, PerlIOBuf_open, PerlIOBase_binmode, NULL,
    PerlIOBase_fileno, PerlIOBuf_dup, PerlIOBuf_read, PerlIOBuf_unread, PerlIOBuf_write,
    PerlIOBuf_seek, PerlIOBuf_tell, PerlIOBuf_close, PerlIOBuf_flush, upper_fill,
    PerlIOBase_eof, PerlIOBase_error, PerlIOBase_clearerr, PerlIOBase_setlinebuf,
    PerlIOBuf_get_base, PerlIOBuf_bufsiz, PerlIOBuf_get_ptr, PerlIOBuf_get_cnt,
    PerlIOBuf_set_ptrcnt,
};
XS_EXTERNAL(boot_Upper) {
    dXSBOOTARGSXSAPIVERCHK;
    PERL_UNUSED_VAR(items);
#ifdef PERLIO_LAYERS
    PerlIO_define_layer(aTHX_ PERLIO_FUNCS_CAST(&upper_funcs));
#else
    croak("PerlIO layers are not supported");
#endif
    Perl_xs_boot_epilog(aTHX_ ax);
}
`)
	program := `require File::Temp;
my ($tmp, $name) = File::Temp::tempfile(UNLINK => 1);
binmode $tmp;
print $tmp "hello\nworld \xc3\xa9\n";
close $tmp;
my @got;
for my $layers (':raw:upper', ':raw:utf8:upper') {
    open my $in, "<$layers", $name or die "open $layers: $!";
    my @lines = <$in>;
    close $in;
    chomp @lines;
    push @got, join("|", map { utf8::is_utf8($_) ? "utf8:" . length($_) . ":" . substr($_, 0, 5) : "bytes:" . length($_) . ":" . substr($_, 0, 5) } @lines);
}
join ",", @got`
	if got, want := requireModule(t, "Upper", so, program), "bytes:5:HELLO|bytes:8:WORLD,utf8:5:HELLO|utf8:7:WORLD"; got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}
