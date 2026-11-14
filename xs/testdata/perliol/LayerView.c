/* A module that walks the layers of a handle the way File::Map and
 * PerlIO::Layers do, through <perliol.h>: PerlIOBase(f)->tab->name and
 * kind, PerlIOBase(f)->flags, PerlIONext(f); and that gives a handle, and a
 * layer of one, to the handle functions. Written in the shape xsubpp
 * generates. */
#include "EXTERN.h"
#include "perl.h"
#include "XSUB.h"
#include "perliol.h"

/* layers(fh): for each layer from the top, "name/b/u/c": whether its kind
 * is buffered and its flags say UTF-8 and CRLF; then the descriptor of the
 * handle and that of the layer below the top, and whether the handle is at
 * its end. */
XS(XS_LayerView_layers) {
    dXSARGS;
    PerlIO *fh, *l;
    SV *out;
    if (items != 1) croak_xs_usage(cv, "fh");
    fh = IoIFP(sv_2io(ST(0)));
    out = sv_2mortal(newSVpvs(""));
    for (l = fh; *l; l = PerlIONext(l))
        sv_catpvf(out, "%s/%d/%d/%d,", PerlIOBase(l)->tab->name,
                  (PerlIOBase(l)->tab->kind & PERLIO_K_BUFFERED) ? 1 : 0,
                  (PerlIOBase(l)->flags & PERLIO_F_UTF8) ? 1 : 0,
                  (PerlIOBase(l)->flags & PERLIO_F_CRLF) ? 1 : 0);
    sv_catpvf(out, "%d,%d,%d", PerlIO_fileno(fh),
              *fh && *PerlIONext(fh) ? PerlIO_fileno(PerlIONext(fh)) : -1,
              (int)PerlIO_eof(fh));
    ST(0) = out;
    XSRETURN(1);
}

XS_EXTERNAL(boot_LayerView) {
    dXSBOOTARGSXSAPIVERCHK;
    PERL_UNUSED_VAR(items);
    newXS_deffile("LayerView::layers", XS_LayerView_layers);
    Perl_xs_boot_epilog(aTHX_ ax);
}
