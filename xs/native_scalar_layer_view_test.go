//go:build darwin || linux

package xs_test

import "testing"

// The view native code gets of an in-memory handle (IoIFP of a handle opened
// on a scalar) has the :scalar layer's flags, as FileHandle::Fmode reads
// them (PerlIOBase(f)->flags & PERLIO_F_CANREAD), and as the layer's buffer
// what the layer's own functions give: the scalar's string. The layer says
// it is buffered, but its struct is no PerlIOBuf; the interpreter read a
// PerlIOBuf's buffer fields past the end of it, from whatever lay there,
// and the view died "guest_mem: pointer outside linear memory" in some runs
// (FileHandle-Fmode's t/basic.t and t/binmode.t) and showed garbage in the
// others.
const scalarLayerViewSource = `#include "EXTERN.h"
#include "perl.h"
#include "XSUB.h"
#include "perliol.h"
/* flags(fh): "r" and "w" as the top layer's flags allow, as FileHandle::Fmode
 * asks for a handle of fileno -1 */
XS(XS_MySl_flags) {
    dXSARGS;
    PerlIO *f;
    PERL_UNUSED_VAR(items);
    f = IoIFP(sv_2io(ST(0)));
    if (!PerlIOValid(f)) croak("not a valid handle");
    ST(0) = sv_2mortal(newSVpvf("%s%s", (PerlIOBase(f)->flags & PERLIO_F_CANREAD) ? "r" : "",
                                (PerlIOBase(f)->flags & PERLIO_F_CANWRITE) ? "w" : ""));
    XSRETURN(1);
}
/* buffer(fh): the top layer's buffer as the view has it: what is read and
 * what is left, and its size */
XS(XS_MySl_buffer) {
    dXSARGS;
    PerlIO *f;
    PerlIOBuf *b;
    PERL_UNUSED_VAR(items);
    f = IoIFP(sv_2io(ST(0)));
    if (!(PerlIOBase(f)->tab->kind & PERLIO_K_BUFFERED)) croak("not buffered");
    b = PerlIOSelf(f, PerlIOBuf);
    if (!b->buf || !b->ptr || !b->end)
        ST(0) = sv_2mortal(newSVpvf("none %lu", (unsigned long)b->bufsiz));
    else
        ST(0) = sv_2mortal(newSVpvf("read [%.*s] left [%.*s] size %lu", (int)(b->ptr - b->buf), b->buf,
                                    (int)(b->end - b->ptr), b->ptr, (unsigned long)b->bufsiz));
    XSRETURN(1);
}
XS_EXTERNAL(boot_MySl) {
    dXSBOOTARGSXSAPIVERCHK;
    newXS_deffile("MySl::flags", XS_MySl_flags);
    newXS_deffile("MySl::buffer", XS_MySl_buffer);
    Perl_xs_boot_epilog(aTHX_ ax);
}
`

func TestNativeViewOfAnInMemoryHandle(t *testing.T) {
	so := buildModuleSource(t, "MySl", scalarLayerViewSource)
	got := requireModule(t, "MySl", so, `my @r;
my $text = "hello world";
open my $in, '<', \$text or die;
read $in, my $two, 2;
push @r, "in " . MySl::flags($in) . " " . MySl::buffer($in);
my $out_text = "";
open my $out, '>', \$out_text or die;
print $out "abc";
push @r, "out " . MySl::flags($out) . " " . MySl::buffer($out);
open my $both, '+<', \$text or die;
push @r, "both " . MySl::flags($both);
join "; ", @r`)
	// the flags are perl's; the buffer is the :scalar layer's own view of its
	// string (PerlIO_get_base, get_ptr, get_cnt, get_bufsiz)
	if want := "in r read [he] left [llo world] size 11; out w none 0; both rw"; got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}
