/* go-perl native XS SDK — the perl.h an XS module is compiled against when
 * it targets go-perl's embedded interpreter as a HOST-NATIVE shared library.
 *
 * This is NOT perl's own perl.h. The real header exposes the interpreter's
 * struct layouts and lets XS macros compile into direct memory access; that
 * can never match go-perl, whose interpreter lives in a wasm32 linear memory
 * (4-byte pointers, offsets as handles). Here every SV is an opaque token
 * and every operation is a call through a vtable the loader injects at
 * dlopen time (__goperl_xs_init). Anything this SDK does not cover fails AT
 * COMPILE TIME — never silently at runtime.
 *
 * Calling convention: every XSUB has the signature
 *     void xsub(goperl_frame_t *_gof)
 * The frame carries the argument/return SV tokens (widened to 64-bit), the
 * item count, and the croak channel. croak() is a host-local
 * setjmp/longjmp back to the XSUB entry — matching perl's own semantics of
 * croak longjmping over XS frames — after which the loader reports the
 * message to the interpreter, which raises the Perl-level die.
 *
 * ABI: GOPERL_XS_ABI below must equal the loader's; __goperl_xs_init
 * refuses a mismatch. */
#ifndef GOPERL_XS_SDK_PERL_H
#define GOPERL_XS_SDK_PERL_H

#include <assert.h>
#include <setjmp.h>
#include <stdarg.h>
#include <stddef.h>
#include <stdint.h>
#include <stdio.h>
#include <string.h>

#define GOPERL_XS_ABI 1u

/* The perl this SDK targets (the embedded interpreter's version). */
#define PERL_REVISION 5
#define PERL_VERSION 42
#define PERL_SUBVERSION 2

typedef int32_t I32;
typedef uint32_t U32;
typedef int64_t IV;
typedef uint64_t UV;
typedef double NV;
typedef size_t STRLEN;
typedef int64_t SSize_t;

/* Opaque handles: guest SVs are linear-memory offsets widened to pointer
 * width. Never dereference. */
typedef struct goperl_sv SV;
typedef struct goperl_sv CV;
typedef struct goperl_sv AV;
typedef struct goperl_sv HV;
typedef struct goperl_sv GV;

#define GOPERL_XS_MAXSTACK 64

struct goperl_frame;

/* The vtable the loader injects. Field order is ABI. */
typedef struct goperl_api {
    uint32_t abi;
    uint32_t reserved;
    int64_t (*sv_iv)(struct goperl_frame *, uint64_t sv);
    const char *(*sv_pv)(struct goperl_frame *, uint64_t sv, uint64_t *lenp);
    uint64_t (*new_iv)(struct goperl_frame *, int64_t v);
    uint64_t (*new_pvn)(struct goperl_frame *, const char *p, uint64_t len);
    uint64_t (*sv_mortal)(struct goperl_frame *, uint64_t sv);
    void (*register_xs)(struct goperl_frame *, const char *name, void *xsub);
} goperl_api_t;

/* One XSUB activation. Field order/sizes are ABI (mirrored by the loader). */
typedef struct goperl_frame {
    const goperl_api_t *api;
    const char *subname; /* fully qualified, for usage croaks */
    void *jb;            /* jmp_buf* armed by dXSARGS */
    int32_t items;
    int32_t nret;
    int32_t failed;
    int32_t reserved;
    uint64_t st[GOPERL_XS_MAXSTACK];
    char err[512];
} goperl_frame_t;

/* The module-global vtable, set by __goperl_xs_init. Single-TU model: the
 * SDK currently assumes the module's XS code lives in one translation unit
 * (xsubpp output); multi-TU dists need this hoisted into a tiny glue .c. */
static const goperl_api_t *goperl_api_v;

__attribute__((visibility("default"), used)) uint32_t
__goperl_xs_init(const goperl_api_t *api) {
    if (!api || api->abi != GOPERL_XS_ABI) return 0;
    goperl_api_v = api;
    return GOPERL_XS_ABI;
}

/* croak: format into the frame and longjmp back to the XSUB entry. */
static void goperl_croakf(goperl_frame_t *f, const char *fmt, ...)
    __attribute__((noreturn, unused, format(printf, 2, 3)));
static void goperl_croakf(goperl_frame_t *f, const char *fmt, ...) {
    va_list ap;
    va_start(ap, fmt);
    vsnprintf(f->err, sizeof f->err, fmt, ap);
    va_end(ap);
    f->failed = 1;
    longjmp(*(jmp_buf *)f->jb, 1);
}

static SV *goperl_newSVpvf(goperl_frame_t *f, const char *fmt, ...)
    __attribute__((unused, format(printf, 2, 3)));
static SV *goperl_newSVpvf(goperl_frame_t *f, const char *fmt, ...) {
    char buf[1024];
    va_list ap;
    va_start(ap, fmt);
    int n = vsnprintf(buf, sizeof buf, fmt, ap);
    va_end(ap);
    if (n < 0) n = 0;
    if ((size_t)n >= sizeof buf) n = (int)sizeof buf - 1;
    return (SV *)(uintptr_t)goperl_api_v->new_pvn(f, buf, (uint64_t)n);
}

/* ---- the macro surface xsubpp-generated code compiles against ---- */

#define STATIC static
#define dNOOP ((void)0)
#define aTHX_
#define PERL_UNUSED_VAR(v) ((void)(v))
#define PERL_UNUSED_ARG(v) ((void)(v))

#define XSPROTO(name) void name(goperl_frame_t *_gof)
#define XS(name) XSPROTO(name)
#define XS_EXTERNAL(name) __attribute__((visibility("default"))) XSPROTO(name)
#define XS_INTERNAL(name) static XSPROTO(name)

#define dXSARGS                                   \
    jmp_buf _gof_jb;                              \
    I32 ax = 0;                                   \
    I32 items;                                    \
    _gof->jb = (void *)&_gof_jb;                  \
    if (setjmp(_gof_jb)) return;                  \
    items = _gof->items;                          \
    PERL_UNUSED_VAR(ax)

#define ST(n) (*(SV **)&_gof->st[(n)])
#define dXSTARG dNOOP
#define XSprePUSH (_gof->nret = 0)
#define PUSHi(v)                                                              \
    (_gof->st[_gof->nret] = goperl_api_v->sv_mortal(                          \
         _gof, goperl_api_v->new_iv(_gof, (int64_t)(v))),                     \
     _gof->nret++)
#define XSRETURN(k)                    \
    do {                               \
        _gof->nret = (int32_t)(k);     \
        return;                        \
    } while (0)
#define XSRETURN_EMPTY XSRETURN(0)

/* Predefining the assert guard makes xsubpp skip its own S_croak_xs_usage
 * fallback (which needs CvGV/GvNAME internals this SDK does not model). */
#define PERL_ARGS_ASSERT_CROAK_XS_USAGE
#define croak_xs_usage(cv_ignored, params) \
    goperl_croakf(_gof, "Usage: %s(%s)", _gof->subname, params)
#define croak(...) goperl_croakf(_gof, __VA_ARGS__)

#define SvIV(sv) (goperl_api_v->sv_iv(_gof, (uint64_t)(uintptr_t)(sv)))
#define SvPV_nolen(sv) \
    ((char *)goperl_api_v->sv_pv(_gof, (uint64_t)(uintptr_t)(sv), 0))
#define SvPV(sv, lenvar)                                                     \
    ((char *)goperl_api_v->sv_pv(_gof, (uint64_t)(uintptr_t)(sv),            \
                                 (uint64_t *)&(lenvar)))
#define sv_2mortal(sv)                                                       \
    ((SV *)(uintptr_t)goperl_api_v->sv_mortal(_gof,                          \
                                              (uint64_t)(uintptr_t)(sv)))
#define newSViv(v) ((SV *)(uintptr_t)goperl_api_v->new_iv(_gof, (int64_t)(v)))
#define newSVpvn(p, l)                                                       \
    ((SV *)(uintptr_t)goperl_api_v->new_pvn(_gof, (p), (uint64_t)(l)))
#define newSVpvf(...) goperl_newSVpvf(_gof, __VA_ARGS__)

/* boot support */
#define dXSBOOTARGSXSAPIVERCHK                                \
    jmp_buf _gof_jb;                                          \
    I32 ax = 0;                                               \
    I32 items = _gof->items;                                  \
    void *cv = 0;                                             \
    _gof->jb = (void *)&_gof_jb;                              \
    if (setjmp(_gof_jb)) return;                              \
    PERL_UNUSED_VAR(ax)
#define Perl_xs_boot_epilog(ax_ignored) ((void)(ax_ignored))
#define Perl_newXS_deffile(name, fn) \
    goperl_api_v->register_xs(_gof, (name), (void *)(fn))

#endif /* GOPERL_XS_SDK_PERL_H */
