/* Entry table of one native XS module: the one symbol the loader resolves,
 * and the only symbol a module object needs to keep global.
 *
 * Compile it once per module, together with the module's XS objects:
 *
 *   cc -I<sdk> -DGOPERL_ENTRY_MODULE=Demo__XS -c goperl_entry.c
 *
 * GOPERL_ENTRY_MODULE is the Perl package with "::" spelled "__", exactly as
 * in the module's boot_ symbol. The slot order is the loader's contract
 * (entryNames in xs/entry.go). */
#include "EXTERN.h"
#include "perl.h"
#include "XSUB.h"
#include "perliol.h"

#include <stdlib.h>
#include <string.h>

#ifndef GOPERL_ENTRY_MODULE
#error "GOPERL_ENTRY_MODULE must name the module (package with :: spelled __)"
#endif

#define GOPERL_ENTRY_BOOT CAT2(boot_, GOPERL_ENTRY_MODULE)
#define GOPERL_ENTRY_TABLE CAT2(goperl_entry_, GOPERL_ENTRY_MODULE)

XS_EXTERNAL(GOPERL_ENTRY_BOOT);

GOPERL_C const void *const GOPERL_ENTRY_TABLE[17];

GOPERL_C __attribute__((visibility("default"), used))
const void *const GOPERL_ENTRY_TABLE[17] = {
    (const void *)GOPERL_ENTRY_BOOT,
    (const void *)__goperl_xs_init,
    (const void *)__goperl_xs_layout,
    (const void *)__goperl_pp_invoke,
    (const void *)__goperl_shared_clone,
    (const void *)__goperl_shared_release,
    (const void *)__goperl_svt_invoke,
    (const void *)__goperl_xs_invoke,
    (const void *)__goperl_xs_epilogue,
    (const void *)__goperl_keyword_invoke,
    (const void *)__goperl_perlio_invoke,
    (const void *)__goperl_dtor_invoke,
    (const void *)__goperl_host_invoke,
    (const void *)strlen,
    (const void *)memcpy,
    (const void *)malloc,
    (const void *)free,
};
