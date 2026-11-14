/* go-perl native XS SDK: the version of the embedded interpreter, under the
 * names perl's patchlevel.h gives it. XS sources include this header by
 * name to branch on the perl they are built for. */
#ifndef GOPERL_XS_SDK_PATCHLEVEL_H
#define GOPERL_XS_SDK_PATCHLEVEL_H

#define PERL_REVISION 5
#define PERL_VERSION 44
#define PERL_SUBVERSION 0
#define PERL_BCDVERSION 0x5044000

#define PERL_API_REVISION 5
#define PERL_API_VERSION 44
#define PERL_API_SUBVERSION 0

/* The names from before perl 5.6. */
#define PATCHLEVEL PERL_VERSION
#define SUBVERSION PERL_SUBVERSION

#endif /* GOPERL_XS_SDK_PATCHLEVEL_H */

/* The patches applied locally to the interpreter (none), for code that
 * includes this header itself; perl.h includes it with
 * PERL_PATCHLEVEL_H_IMPLICIT, as perl's does, and defines no array. */
#if !defined(PERL_PATCHLEVEL_H_IMPLICIT) && !defined(LOCAL_PATCH_COUNT)
static const char *const local_patches[] = {NULL, NULL};
#define LOCAL_PATCH_COUNT ((int)(sizeof(local_patches) / sizeof(local_patches[0]) - 2))
#endif
