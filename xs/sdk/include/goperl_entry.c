/* Entry table of one native XS module: the one symbol the loader resolves,
 * and the only symbol a module object needs to keep global.
 *
 * Compile it once per module, together with the module's XS objects:
 *
 *   cc -I<sdk> -DGOPERL_ENTRY_BOOT=boot_Demo__XS \
 *      -DGOPERL_ENTRY_TABLE=goperl_entry_Demo__XS -c goperl_entry.c
 *
 * Both names end in the Perl package with "::" spelled "__", exactly as in
 * the module's boot_ symbol. Both are whole identifiers: this file pastes
 * nothing onto them. The slot order is the loader's contract (entryNames in
 * xs/entry.go). */
#include "EXTERN.h"
#include "perl.h"
#include "XSUB.h"
#include "perliol.h"

#include <locale.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#if defined(__linux__)
#include <errno.h>
#include <grp.h>
#include <unistd.h>
#endif

#if !defined(GOPERL_ENTRY_BOOT) || !defined(GOPERL_ENTRY_TABLE)
#error "GOPERL_ENTRY_BOOT and GOPERL_ENTRY_TABLE must name the module's boot function and entry table"
#endif

/* The C library functions that change the ids of the process (the
 * runtime calls them on Linux only), each slot empty elsewhere. */
#if defined(__linux__)
#define GOPERL_ENTRY_LINUX(fn) (const void *)fn
#else
#define GOPERL_ENTRY_LINUX(fn) (const void *)0
#endif

XS_EXTERNAL(GOPERL_ENTRY_BOOT);

GOPERL_C const void *const GOPERL_ENTRY_TABLE[29];

GOPERL_C __attribute__((visibility("default"), used))
const void *const GOPERL_ENTRY_TABLE[29] = {
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
    (const void *)fflush,
    (const void *)setlocale,
    GOPERL_ENTRY_LINUX(setuid),
    GOPERL_ENTRY_LINUX(seteuid),
    GOPERL_ENTRY_LINUX(setreuid),
    GOPERL_ENTRY_LINUX(setresuid),
    GOPERL_ENTRY_LINUX(setgid),
    GOPERL_ENTRY_LINUX(setegid),
    GOPERL_ENTRY_LINUX(setregid),
    GOPERL_ENTRY_LINUX(setresgid),
    GOPERL_ENTRY_LINUX(setgroups),
    GOPERL_ENTRY_LINUX(__errno_location),
};
