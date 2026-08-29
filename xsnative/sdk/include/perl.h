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

#define GOPERL_XS_ABI 2u

/* The perl this SDK targets (the embedded interpreter's version). */
#define PERL_REVISION 5
#define PERL_VERSION 42
#define PERL_SUBVERSION 2

/* A dist's bundled ppport.h is a portability layer for REAL perl headers;
 * against this SDK it must not activate. Its include guard is pre-defined
 * so the #include turns into a no-op. */
#define _P_P_PORTABILITY_H_

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
    /* v2: the generic SV micro-operation (op selects; see the loader for the
     * table), plus the host-pointer registry backing T_PTROBJ - guest IVs
     * are 32-bit, so native object pointers cross as registry ids. */
    uint64_t (*xs_op)(struct goperl_frame *, int32_t op, uint64_t a,
                      uint64_t b, const char *s, uint64_t slen);
    uint64_t (*ptr_encode)(struct goperl_frame *, void *p);
    void *(*ptr_decode)(struct goperl_frame *, uint64_t id);
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
    /* Scratch slot backing the pointer-returning fetch macros (*av_fetch /
     * *hv_fetchs): the fetched token parks here so the macro can hand back
     * an SV** the caller immediately dereferences. One slot suffices for
     * xsubpp-style immediate-deref usage. */
    uint64_t tmp;
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

/* v2 op numbers (mirror perl-wasm's perl_xs_helper). */
enum {
    GOPERL_OP_SV_IV = 1,
    GOPERL_OP_SV_PV = 2,
    GOPERL_OP_NEW_IV = 3,
    GOPERL_OP_NEW_PVN = 4,
    GOPERL_OP_SV_MORTAL = 5,
    GOPERL_OP_NEW_UV = 6,
    GOPERL_OP_NEW_AV = 7,
    GOPERL_OP_NEW_HV = 8,
    GOPERL_OP_NEW_RV_INC = 9,
    GOPERL_OP_AV_PUSH = 10,
    GOPERL_OP_AV_LEN = 11,
    GOPERL_OP_AV_FETCH = 12,
    GOPERL_OP_HV_STORE = 13,
    GOPERL_OP_HV_FETCH = 14,
    GOPERL_OP_REFCNT_INC = 15,
    GOPERL_OP_GV_STASHPV = 16,
    GOPERL_OP_SV_BLESS = 17,
    GOPERL_OP_SV_RV = 18,
    GOPERL_OP_SV_TYPE_CLASS = 19,
    GOPERL_OP_SV_ISA = 20,
    GOPERL_OP_SV_DERIVED_FROM = 21,
    GOPERL_OP_SETREF_IV = 22
};

#define GOPERL_TOK(sv) ((uint64_t)(uintptr_t)(sv))
#define GOPERL_SV(tok) ((SV *)(uintptr_t)(tok))
#define goperl_op0(op, a, b)     (goperl_api_v->xs_op(_gof, (op), (uint64_t)(a), (uint64_t)(b), 0, 0))
#define goperl_ops(op, a, b, s, l)     (goperl_api_v->xs_op(_gof, (op), (uint64_t)(a), (uint64_t)(b), (s), (uint64_t)(l)))

#define STMT_START do
#define STMT_END while (0)
#define SvGETMAGIC(sv) ((void)(sv))
#define Perl_croak_nocontext(...) goperl_croakf(_gof, __VA_ARGS__)
#define Perl_croak(...) goperl_croakf(_gof, __VA_ARGS__)
#define CvGV(cv) (cv)
#define GvNAME(gv) (_gof->subname)

/* SvTYPE speaks our own stable classification (the guest translates from
 * perl's svtype), so both halves of `SvTYPE(x) == SVt_PVHV` agree. */
#define SVt_PVAV 1
#define SVt_PVHV 2
#define SVt_PVCV 3
#define SvTYPE(sv) ((int)goperl_op0(GOPERL_OP_SV_TYPE_CLASS, GOPERL_TOK(sv), 0))
#define SvROK(sv) (goperl_op0(GOPERL_OP_SV_RV, GOPERL_TOK(sv), 0) != 0)
#define SvRV(sv) GOPERL_SV(goperl_op0(GOPERL_OP_SV_RV, GOPERL_TOK(sv), 0))
#define SvPVX(sv) SvPV_nolen(sv)
#define SvIVX(sv) SvIV(sv)
#define SvUV(sv) ((UV)SvIV(sv))
#define sv_newmortal() sv_2mortal(newSVpvn("", 0))

#define newSVuv(v) GOPERL_SV(goperl_op0(GOPERL_OP_NEW_UV, (uint64_t)(v), 0))
static SV *goperl_newSVpv(goperl_frame_t *f, const char *s, STRLEN len)
    __attribute__((unused));
static SV *goperl_newSVpv(goperl_frame_t *f, const char *s, STRLEN len) {
    if (len == 0) len = s ? strlen(s) : 0;
    return (SV *)(uintptr_t)goperl_api_v->xs_op(f, GOPERL_OP_NEW_PVN, 0,
                                                (uint64_t)len, s, (uint64_t)len);
}
#define newSVpv(s, l) goperl_newSVpv(_gof, (s), (STRLEN)(l))
#define newAV() ((AV *)GOPERL_SV(goperl_op0(GOPERL_OP_NEW_AV, 0, 0)))
#define newHV() ((HV *)GOPERL_SV(goperl_op0(GOPERL_OP_NEW_HV, 0, 0)))
#define newRV_inc(sv) GOPERL_SV(goperl_op0(GOPERL_OP_NEW_RV_INC, GOPERL_TOK(sv), 0))
#define newRV(sv) newRV_inc(sv)
#define SvREFCNT_inc(sv) GOPERL_SV(goperl_op0(GOPERL_OP_REFCNT_INC, GOPERL_TOK(sv), 0))

#define av_push(av, sv) ((void)goperl_op0(GOPERL_OP_AV_PUSH, GOPERL_TOK(av), GOPERL_TOK(sv)))
#define av_len(av) ((SSize_t)(int64_t)goperl_op0(GOPERL_OP_AV_LEN, GOPERL_TOK(av), 0))
#define av_fetch(av, i, lval)     (_gof->tmp = goperl_op0(GOPERL_OP_AV_FETCH, GOPERL_TOK(av), (uint64_t)(int64_t)(i)),      _gof->tmp ? (SV **)&_gof->tmp : (SV **)0)

#define hv_store(hv, key, klen, sv, hash)     ((void)goperl_ops(GOPERL_OP_HV_STORE, GOPERL_TOK(hv), GOPERL_TOK(sv), (key), (uint64_t)(klen)))
#define hv_stores(hv, key, sv) hv_store((hv), "" key "", sizeof(key) - 1, (sv), 0)
/* hv_fetchs is used in the wild both 2-arg (literal, standard) and 3-arg
 * (key, len - some dists); tolerate both, length recomputed from the key. */
#define hv_fetchs(hv, key, ...)     (_gof->tmp = goperl_ops(GOPERL_OP_HV_FETCH, GOPERL_TOK(hv), 0, (key), strlen(key)),      _gof->tmp ? (SV **)&_gof->tmp : (SV **)0)
#define hv_fetch(hv, key, klen, lval)     (_gof->tmp = goperl_ops(GOPERL_OP_HV_FETCH, GOPERL_TOK(hv), 0, (key), (uint64_t)(klen)),      _gof->tmp ? (SV **)&_gof->tmp : (SV **)0)

#define gv_stashpv(name, flags)     ((HV *)GOPERL_SV(goperl_ops(GOPERL_OP_GV_STASHPV, (uint64_t)(int64_t)(flags), 0, (name), strlen(name))))
#define sv_bless(rv, stash)     GOPERL_SV(goperl_op0(GOPERL_OP_SV_BLESS, GOPERL_TOK(rv), GOPERL_TOK(stash)))
#define sv_isa(sv, name)     ((int)goperl_ops(GOPERL_OP_SV_ISA, GOPERL_TOK(sv), 0, (name), strlen(name)))
#define sv_derived_from(sv, name)     ((int)goperl_ops(GOPERL_OP_SV_DERIVED_FROM, GOPERL_TOK(sv), 0, (name), strlen(name)))

/* T_PTROBJ: guest IVs are 32-bit, so native object pointers live in the
 * loader's registry and only the id crosses. sv_setref_pv/INT2PTR are the
 * two ends of that path; PTR2IV exists for symmetry. */
#define sv_setref_pv(rv, classname, ptr)     GOPERL_SV(goperl_ops(GOPERL_OP_SETREF_IV, GOPERL_TOK(rv),                          goperl_api_v->ptr_encode(_gof, (void *)(ptr)),                          (classname), strlen(classname)))
#define INT2PTR(type, iv) ((type)goperl_api_v->ptr_decode(_gof, (uint64_t)(iv)))
#define PTR2IV(p) ((IV)goperl_api_v->ptr_encode(_gof, (void *)(p)))
#define PTR2UV(p) ((UV)PTR2IV(p))
#define UVxf "llx"

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
