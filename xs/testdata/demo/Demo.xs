#define PERL_NO_GET_CONTEXT
#include "EXTERN.h"
#include "perl.h"
#include "XSUB.h"

/* A deliberately tiny XS module: one pure function and one that touches
 * the interpreter (creates an SV), so the import surface shows both the
 * data (GOT.mem) and function (GOT.func) relocation kinds. */

MODULE = Demo::XS  PACKAGE = Demo::XS

int
add(a, b)
    int a
    int b
  CODE:
    RETVAL = a + b;
  OUTPUT:
    RETVAL

SV *
greet(name)
    const char *name
  CODE:
    RETVAL = newSVpvf("hello, %s", name);
  OUTPUT:
    RETVAL
