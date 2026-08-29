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
 * v3 models enough of the interpreter surface for stack-driven, MAGIC-using,
 * pp-poking XS (the Text::Xslate class of module):
 *
 *   - pTHX plumbing: `pTHX` expands to the call frame, so every function
 *     that threads perl context (`pTHX_` / `aTHX_`) carries the frame.
 *     XSUBs have perl's real signature: void xsub(pTHX_ CV* cv).
 *   - The perl argument stack is UNIFIED with the frame: PL_stack_base /
 *     PL_stack_sp / marks live host-side (SV tokens widened to 64-bit), and
 *     guest operations exchange explicit token lists. dSP/PUSHMARK/PUTBACK/
 *     SPAGAIN and the xsubpp macro set behave like the real ones.
 *   - MAGIC chains are host-side mirror structs (vtbl identity comparisons
 *     work); a guest-side anchor keeps lifetimes aligned — freeing the SV
 *     runs the module's svt_free through the loader.
 *   - PL_ppaddr is a proxy table of one generic trampoline; running a
 *     scratch OP (the pp_flop "fake op" idiom) executes the real pp in the
 *     guest.
 *   - PL_warnhook/PL_diehook are read-through/write-back shadows, flushed
 *     before any guest operation that can run Perl code.
 *   - AvARRAY returns a host mirror of the array body; host writes through
 *     the mirror are flushed back (raw, refcount-neutral) before the next
 *     guest operation.
 *
 * croak() is a host-local setjmp/longjmp back to the XSUB entry — matching
 * perl's own semantics of croak longjmping over XS frames — after which the
 * loader reports the message to the interpreter, which raises the Perl-level
 * die.
 *
 * Multi-TU dists: the shared state below is declared weak, so linking
 * several xsubpp outputs into one .so shares one copy.
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
#include <stdlib.h>
#include <string.h>
#include <sys/stat.h>
#include <time.h>
#include <unistd.h>

#define GOPERL_XS_ABI 3u

/* The perl this SDK targets (the embedded interpreter's version). */
#define PERL_REVISION 5
#define PERL_VERSION 42
#define PERL_SUBVERSION 2
#define PERL_BCDVERSION 0x5042002
#define PERL_VERSION_DECIMAL(r, v, s) ((r)*1000000 + (v)*1000 + (s))
#define PERL_DECIMAL_VERSION \
    PERL_VERSION_DECIMAL(PERL_REVISION, PERL_VERSION, PERL_SUBVERSION)
#define PERL_VERSION_GE(r, v, s) \
    (PERL_DECIMAL_VERSION >= PERL_VERSION_DECIMAL(r, v, s))
#define PERL_VERSION_GT(r, v, s) \
    (PERL_DECIMAL_VERSION > PERL_VERSION_DECIMAL(r, v, s))
#define PERL_VERSION_LE(r, v, s) \
    (PERL_DECIMAL_VERSION <= PERL_VERSION_DECIMAL(r, v, s))
#define PERL_VERSION_LT(r, v, s) \
    (PERL_DECIMAL_VERSION < PERL_VERSION_DECIMAL(r, v, s))

/* A dist's bundled ppport.h is a portability layer for REAL perl headers;
 * against this SDK it must not activate. Its include guard is pre-defined
 * so the #include turns into a no-op. */
#define _P_P_PORTABILITY_H_

typedef int32_t I32;
typedef uint32_t U32;
typedef int16_t I16;
typedef uint16_t U16;
typedef int8_t I8;
typedef uint8_t U8;
typedef int64_t IV;
typedef uint64_t UV;
typedef double NV;
typedef size_t STRLEN;
typedef int64_t SSize_t;
#ifndef __cplusplus
#include <stdbool.h>
#endif
#ifndef TRUE
#define TRUE 1
#define FALSE 0
#endif

/* Opaque handles: guest SVs are linear-memory offsets widened to pointer
 * width. Never dereference. */
typedef struct goperl_sv SV;
typedef struct goperl_sv CV;
typedef struct goperl_sv AV;
typedef struct goperl_sv HV;
typedef struct goperl_sv GV;
typedef struct goperl_he HE;

/* svtype: the REAL perl 5.42 enum values, and the guest reports raw
 * SvTYPE, so comparisons like SvTYPE(sv) <= SVt_PVMG agree exactly. */
typedef enum {
    SVt_NULL = 0,
    SVt_IV = 1,
    SVt_NV = 2,
    SVt_PV = 3,
    SVt_INVLIST = 4,
    SVt_PVIV = 5,
    SVt_PVNV = 6,
    SVt_PVMG = 7,
    SVt_REGEXP = 8,
    SVt_PVGV = 9,
    SVt_PVLV = 10,
    SVt_PVAV = 11,
    SVt_PVHV = 12,
    SVt_PVCV = 13,
    SVt_PVFM = 14,
    SVt_PVIO = 15,
    SVt_PVOBJ = 16,
    SVt_LAST = 17
} svtype;

/* Context / call flags (real perl 5.42 values; they cross to the guest). */
#define G_VOID 1
#define G_SCALAR 2
#define G_LIST 3
#define G_ARRAY G_LIST
#define G_WANT 3
#define G_DISCARD 0x4
#define G_EVAL 0x8
#define G_NOARGS 0x10
#define G_KEEPERR 0x20
#define G_METHOD 0x80

/* gv.h flags (real values). */
#define GV_ADD 0x01
#define GV_ADDMULTI 0x02
#define GV_NOADD_NOINIT 0x00 /* unused placeholder */

/* overload (amagic) — real perl 5.42 values used by deref overloading. */
#define fallback_amg 0
#define to_sv_amg 1
#define to_av_amg 2
#define to_hv_amg 3
#define to_gv_amg 4
#define to_cv_amg 5
#define AMGf_noright 1
#define AMGf_unary 8

/* scope.h */
#define SAVEf_KEEPOLDELEM 2

/* op.h want bits. */
#define OPf_WANT 3
#define OPf_WANT_VOID 1
#define OPf_WANT_SCALAR 2
#define OPf_WANT_LIST 3

/* Opcode numbers this SDK knows how to run through RUN_PP_SCRATCH. */
#define OP_FLOP 182
#define OP_max 443

/* MAGIC type characters (mg_vtable.h). */
#define PERL_MAGIC_ext '~'
#define PERL_MAGIC_extvalue '^'
#define PERL_MAGIC_tied 'P'
#define PERL_MAGIC_taint 't'
#define MGf_REFCOUNTED 2
#define MGf_COPY 8
#define MGf_DUP 0x10
#define MGf_LOCAL 0x20
#define HEf_SVKEY -2

/* SvFLAGS speaks the SDK's synthetic bitset (guest op SV_INFO): SvFLAGS(a)
 * & (SVf_POK|SVf_IOK|SVf_NOK) style tests work because both sides use the
 * same synthetic values. These are NOT perl's real flag bits. */
#define SVf_OK 1
#define SVf_ROK 2
#define SVf_POK 4
#define GOPERL_INFO_NIOK 8
#define GOPERL_INFO_UTF8 16
#define GOPERL_INFO_ISOBJ 32
#define GOPERL_INFO_READONLY 64
#define GOPERL_INFO_ISCV 128
#define GOPERL_INFO_ISGV 256
#define SVf_IOK 512
#define SVf_NOK 1024
#define GOPERL_INFO_POKp 2048
#define GOPERL_INFO_NIOKp 4096
#define GOPERL_INFO_ISUV 8192
#define GOPERL_INFO_RMAGICAL 16384
#define GOPERL_INFO_OBJECT 32768
#define SVs_TEMP 0x00080000 /* only consumed by newSVpvs_flags below */

/* ---- host-side MAGIC mirror (layout is ABI: the loader writes it) ------- */

struct goperl_frame;
typedef struct goperl_frame goperl_frame_t;

/* pTHX: perl context threading. The "interpreter" is the live call frame. */
#define pTHX goperl_frame_t *_gof
#define pTHX_ goperl_frame_t *_gof,
#define aTHX _gof
#define aTHX_ _gof,
#define dTHX goperl_frame_t *const _gof = goperl_cur_frame_v
#define pTHX_1 2
#define pTHX_2 3
#define pTHX_3 4
#define pTHX_4 5
/* Format checking is off: SVf is a custom marker the compiler must not
 * try to type-check. */
#define __attribute__format__(x, y, z)

/* Thread-clone parameter type (go-perl builds without ithreads, so CLONE
 * hooks never run; the type exists so vtables and CLONE code compile). */
typedef struct goperl_clone_params {
    int unused;
} CLONE_PARAMS;

struct magic;
typedef struct magic MAGIC;
typedef struct mgvtbl {
    int (*svt_get)(pTHX_ SV *sv, MAGIC *mg);
    int (*svt_set)(pTHX_ SV *sv, MAGIC *mg);
    U32 (*svt_len)(pTHX_ SV *sv, MAGIC *mg);
    int (*svt_clear)(pTHX_ SV *sv, MAGIC *mg);
    int (*svt_free)(pTHX_ SV *sv, MAGIC *mg);
    int (*svt_copy)(pTHX_ SV *sv, MAGIC *mg, SV *nsv, const char *name,
                    I32 namlen);
    int (*svt_dup)(pTHX_ MAGIC *mg, CLONE_PARAMS *param);
    int (*svt_local)(pTHX_ SV *nsv, MAGIC *mg);
} MGVTBL;

/* Field offsets are ABI (the loader allocates and fills these):
 * mg_moremagic 0, mg_virtual 8, mg_private 16, mg_type 18, mg_flags 19,
 * mg_len 20 (I32), mg_obj 24, mg_ptr 32; sizeof == 40. */
struct magic {
    MAGIC *mg_moremagic;
    MGVTBL *mg_virtual;
    U16 mg_private;
    char mg_type;
    U8 mg_flags;
    I32 mg_len;
    SV *mg_obj;
    char *mg_ptr;
};

/* ---- host-side OP model (for the scratch-op / PL_ppaddr idiom) ---------- */

struct op;
typedef struct op OP;
typedef OP *(*Perl_ppaddr_t)(pTHX);
struct op {
    OP *op_next;
    OP *op_sibparent;
    Perl_ppaddr_t op_ppaddr;
    UV op_targ;
    U16 op_type;
    U16 op_spare;
    U8 op_flags;
    U8 op_private;
};

/* ---- the vtable the loader injects (field order is ABI) ----------------- */

typedef struct goperl_api {
    uint32_t abi;
    uint32_t reserved;
    int64_t (*sv_iv)(goperl_frame_t *, uint64_t sv);
    const char *(*sv_pv)(goperl_frame_t *, uint64_t sv, uint64_t *lenp);
    uint64_t (*new_iv)(goperl_frame_t *, int64_t v);
    uint64_t (*new_pvn)(goperl_frame_t *, const char *p, uint64_t len);
    uint64_t (*sv_mortal)(goperl_frame_t *, uint64_t sv);
    void (*register_xs)(goperl_frame_t *, const char *name, void *xsub);
    uint64_t (*xs_op)(goperl_frame_t *, int32_t op, uint64_t a, uint64_t b,
                      const char *s, uint64_t slen);
    uint64_t (*ptr_encode)(goperl_frame_t *, void *p);
    void *(*ptr_decode)(goperl_frame_t *, uint64_t id);
    /* v3 */
    void *(*guest_mem)(goperl_frame_t *, uint64_t gptr);
    uint64_t (*new_xs)(goperl_frame_t *, const char *name, void *xsub);
    void *(*cv_any)(goperl_frame_t *, uint64_t cv);
    void *(*cv_xsub)(goperl_frame_t *, uint64_t cv);
    MAGIC *(*magic_ext)(goperl_frame_t *, uint64_t sv, uint64_t obj,
                        int32_t how, const void *vtbl, const char *ptr,
                        int64_t len);
    MAGIC *(*magic_chain)(goperl_frame_t *, uint64_t sv);
    void (*magic_del)(goperl_frame_t *, uint64_t sv, int32_t how,
                      const void *vtbl);
} goperl_api_t;

/* ---- one XSUB activation (layout is ABI, mirrored by the loader) -------- */

#define GOPERL_XS_STACK 512
#define GOPERL_XS_MARKS 64
#define GOPERL_XS_TMPS 16

struct goperl_frame {
    const goperl_api_t *api; /* +0 */
    const char *subname;     /* +8   fully qualified, for usage croaks */
    void *jb;                /* +16  jmp_buf* armed by dXSARGS */
    int32_t items;           /* +24  arg count at entry */
    int32_t failed;          /* +28 */
    SV **psp;                /* +32  PL_stack_sp (points into st) */
    int32_t markidx;         /* +40 */
    int32_t hostsave_base;   /* +44  host save-stack depth at XSUB entry */
    int32_t hook_dirty;      /* +48  bit0 diehook, bit1 warnhook pending */
    int32_t tmpidx;          /* +52  rotating slot cursor */
    uint64_t hook_val[2];    /* +56  pending hook assignments */
    uint64_t imm[3];         /* +72  cached &PL_sv_undef/yes/no tokens */
    OP *plop;                /* +96  PL_op host shadow */
    uint64_t st[GOPERL_XS_STACK];   /* +104 the perl stack (tokens) */
    int32_t marks[GOPERL_XS_MARKS]; /* mark offsets into st */
    uint64_t tmp[GOPERL_XS_TMPS];   /* slots backing SV**-returning macros */
    char err[512];
};

/* ---- shared module state (weak: one copy across a multi-TU .so) --------- */

__attribute__((weak)) const goperl_api_t *goperl_api_v = 0;
__attribute__((weak)) goperl_frame_t *goperl_cur_frame_v = 0;

/* Host save-stack backing SAVESPTR/save_hptr on HOST memory locations,
 * unwound by the LEAVE that closes the scope (or by croak unwind). */
typedef struct goperl_hostsave {
    void **loc;
    void *val;
    int32_t scope;
} goperl_hostsave_t;
#define GOPERL_HOSTSAVE_MAX 256
__attribute__((weak)) goperl_hostsave_t goperl_hostsave_v[GOPERL_HOSTSAVE_MAX];
__attribute__((weak)) int32_t goperl_hostsave_n = 0;
__attribute__((weak)) int32_t goperl_scope_v = 0;

/* AV body mirrors backing AvARRAY: per-AV stable buffers of tokens, with a
 * shadow copy for detecting host writes to flush back (refcount-neutral). */
typedef struct goperl_avmirror {
    uint64_t avtok;
    uint64_t *buf;
    uint64_t *shadow;
    int32_t len;
    int32_t cap;
} goperl_avmirror_t;
#define GOPERL_AVMIRROR_MAX 64
__attribute__((weak)) goperl_avmirror_t goperl_avmirrors_v[GOPERL_AVMIRROR_MAX];
__attribute__((weak)) int32_t goperl_avmirror_n = 0;

/* IV mirrors backing the lvalue SvIVX idiom (`SvIVX(counter)++`): the value
 * lives host-side between guest operations and any change is flushed back
 * as sv_setiv before the next one. */
typedef struct goperl_ivmirror {
    uint64_t svtok;
    IV val;
    IV shadow;
} goperl_ivmirror_t;
#define GOPERL_IVMIRROR_MAX 32
__attribute__((weak)) goperl_ivmirror_t goperl_ivmirrors_v[GOPERL_IVMIRROR_MAX];
__attribute__((weak)) int32_t goperl_ivmirror_n = 0;

/* PL_ppaddr proxy: every slot is the same generic trampoline. */
__attribute__((weak)) Perl_ppaddr_t goperl_ppaddr_v[OP_max];

/* Native-XSUB nesting depth: when the top-level XSUB returns, the AV body
 * mirrors are dropped (guest AVs may be freed between calls; a stale
 * mirror must never flush into reused memory). */
__attribute__((weak)) int32_t goperl_xs_depth_v = 0;

/* form() rotating buffers. */
#define GOPERL_FORM_BUFS 8
#define GOPERL_FORM_LEN 1024
__attribute__((weak)) char goperl_form_bufs_v[GOPERL_FORM_BUFS][GOPERL_FORM_LEN];
__attribute__((weak)) int32_t goperl_form_ix_v = 0;

__attribute__((weak, visibility("default"), used)) uint32_t
__goperl_xs_init(const goperl_api_t *api) {
    if (!api || api->abi != GOPERL_XS_ABI) return 0;
    goperl_api_v = api;
    return GOPERL_XS_ABI;
}

/* ---- guest op numbers (mirror perl-wasm's perl_xs_helper) --------------- */
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
    GOPERL_OP_SV_TYPE = 19,
    GOPERL_OP_SV_ISA = 20,
    GOPERL_OP_SV_DERIVED_FROM = 21,
    GOPERL_OP_SETREF_IV = 22,
    GOPERL_OP_SV_SETSV = 23,
    GOPERL_OP_SV_SETSV_NOMG = 24,
    GOPERL_OP_SV_SETSV_MG = 25,
    GOPERL_OP_SV_SETIV = 26,
    GOPERL_OP_SV_SETUV = 27,
    GOPERL_OP_SV_SETNV = 28,
    GOPERL_OP_SV_SETPVN = 29,
    GOPERL_OP_SV_NV = 30,
    GOPERL_OP_NEW_NV = 31,
    GOPERL_OP_NEW_SVSV = 32,
    GOPERL_OP_SV_MORTALCOPY = 33,
    GOPERL_OP_SV_CATSV = 34,
    GOPERL_OP_SV_CATSV_NOMG = 35,
    GOPERL_OP_SV_CATPVN = 36,
    GOPERL_OP_SV_TRUE = 37,
    GOPERL_OP_SV_INFO = 38,
    GOPERL_OP_SV_CUR_SET = 39,
    GOPERL_OP_SV_GROW = 40,
    GOPERL_OP_SV_EQ = 41,
    GOPERL_OP_SV_CMP = 42,
    GOPERL_OP_SV_RVWEAKEN = 43,
    GOPERL_OP_SV_2CV = 44,
    GOPERL_OP_GET_SV = 45,
    GOPERL_OP_GET_CV = 46,
    GOPERL_OP_GV_FETCH = 47,
    GOPERL_OP_AV_STORE = 48,
    GOPERL_OP_AV_EXTEND = 49,
    GOPERL_OP_AV_FILL = 50,
    GOPERL_OP_AV_READ = 51,
    GOPERL_OP_HV_FETCH_ENT = 53,
    GOPERL_OP_HV_STORE_ENT = 54,
    GOPERL_OP_HV_EXISTS_ENT = 55,
    GOPERL_OP_HE_VAL = 56,
    GOPERL_OP_HV_ITERINIT = 57,
    GOPERL_OP_HV_ITERNEXT = 58,
    GOPERL_OP_HV_ITERKEYSV = 59,
    GOPERL_OP_HV_ITERVAL = 60,
    GOPERL_OP_ENTER = 61,
    GOPERL_OP_LEAVE = 62,
    GOPERL_OP_SAVETMPS = 63,
    GOPERL_OP_FREETMPS = 64,
    GOPERL_OP_CALL_SV = 65,
    GOPERL_OP_CALL_METHOD = 66,
    GOPERL_OP_ERRSV = 67,
    GOPERL_OP_SAVE_OP = 68,
    GOPERL_OP_RUN_PP = 69,
    GOPERL_OP_MAGIC_ATTACH = 70,
    GOPERL_OP_MAGIC_ID = 71,
    GOPERL_OP_MAGIC_UNATTACH = 72,
    GOPERL_OP_PLVAR_GET = 73,
    GOPERL_OP_PLVAR_SET = 74,
    GOPERL_OP_SV_UTF8_ON = 75,
    GOPERL_OP_SV_POK_ON = 76,
    GOPERL_OP_SV_DUMP = 77,
    GOPERL_OP_SV_REFCNT_DEC = 78,
    GOPERL_OP_SV_UNMAGIC = 79,
    GOPERL_OP_SAVE_HOOK = 80,
    GOPERL_OP_NEW_XS = 81,
    GOPERL_OP_SV_STASH = 82,
    GOPERL_OP_NEW_SV = 83,
    GOPERL_OP_SV_UPGRADE = 84,
    GOPERL_OP_SV_UTF8_UPGRADE = 85,
    GOPERL_OP_SAVE_DELETE = 86,
    GOPERL_OP_SAVE_HELEM = 87,
    GOPERL_OP_NEW_PVN_SHARE = 88,
    GOPERL_OP_SV_LEN = 89,
    GOPERL_OP_NEW_HVHV = 90,
    GOPERL_OP_AV_MAKE = 91,
    GOPERL_OP_LOOKS_LIKE_NUMBER = 92,
    GOPERL_OP_SV_AMAGIC = 93,
    GOPERL_OP_AMAGIC_CALL = 94,
    GOPERL_OP_WARN = 95,
    GOPERL_OP_GET_HV = 96,
    GOPERL_OP_AV_STORE_RAW = 97
};

/* PLVAR_GET/SET ids. */
#define GOPERL_PL_DIEHOOK 1
#define GOPERL_PL_WARNHOOK 2
#define GOPERL_PL_SV_UNDEF 3
#define GOPERL_PL_SV_YES 4
#define GOPERL_PL_SV_NO 5

#define GOPERL_TOK(sv) ((uint64_t)(uintptr_t)(sv))
#define GOPERL_SV(tok) ((SV *)(uintptr_t)(tok))

/* ---- the guest-op funnel: every operation flushes host state first ------ */

static void goperl_flush(goperl_frame_t *f);

static uint64_t goperl_do_op(goperl_frame_t *f, int32_t op, uint64_t a,
                             uint64_t b, const char *s, uint64_t slen) {
    goperl_flush(f);
    return goperl_api_v->xs_op(f, op, a, b, s, slen);
}

#define goperl_op0(op, a, b) \
    (goperl_do_op(_gof, (op), (uint64_t)(a), (uint64_t)(b), 0, 0))
#define goperl_ops(op, a, b, s, l)                                    \
    (goperl_do_op(_gof, (op), (uint64_t)(a), (uint64_t)(b), (s),      \
                  (uint64_t)(l)))

/* croak: format into the frame and longjmp back to the XSUB entry. */
static void goperl_vfmt(goperl_frame_t *f, char *dst, size_t dstsize,
                        const char *fmt, va_list ap);

static void goperl_hostsave_unwind_to(goperl_frame_t *f, int32_t base) {
    (void)f;
    while (goperl_hostsave_n > base) {
        goperl_hostsave_t *e = &goperl_hostsave_v[--goperl_hostsave_n];
        *e->loc = e->val;
    }
}

static void goperl_raise(goperl_frame_t *f) __attribute__((noreturn, unused));
static void goperl_raise(goperl_frame_t *f) {
    f->failed = 1;
    if (!f->jb) {
        /* No XSUB entry to unwind to (e.g. a croak inside svt_free
         * teardown): surface the message and stop instead of jumping
         * through a null buffer. */
        fprintf(stderr, "goperl native XS: croak outside an XSUB: %s\n",
                f->err);
        abort();
    }
    longjmp(*(jmp_buf *)f->jb, 1);
}

static void goperl_croakf(goperl_frame_t *f, const char *fmt, ...)
    __attribute__((noreturn, unused));
static void goperl_croakf(goperl_frame_t *f, const char *fmt, ...) {
    va_list ap;
    va_start(ap, fmt);
    goperl_vfmt(f, f->err, sizeof f->err, fmt, ap);
    va_end(ap);
    goperl_raise(f);
}

/* ---- SVf-aware formatting ----------------------------------------------
 * SVf is a marker sequence; "%"SVf in a format consumes an SV* vararg and
 * appends its PV. Everything else is parsed one conversion at a time and
 * handed to snprintf with the right va_arg type. */
#define SVf "\002GSV\002"
#define IVdf "lld"
#define UVuf "llu"
#define UVxf "llx"
#define UVof "llo"
#define NVff "f"
#define NVgf "g"

static const char *goperl_sv_pv_(goperl_frame_t *f, SV *sv, uint64_t *lenp) {
    goperl_flush(f);
    return goperl_api_v->sv_pv(f, GOPERL_TOK(sv), lenp);
}

static void goperl_vfmt(goperl_frame_t *f, char *dst, size_t dstsize,
                        const char *fmt, va_list ap) {
    size_t o = 0;
    if (dstsize == 0) return;
#define GOPERL_PUTC(c)                    \
    do {                                  \
        if (o + 1 < dstsize) dst[o] = (c); \
        o++;                              \
    } while (0)
    while (*fmt) {
        if (*fmt != '%') {
            GOPERL_PUTC(*fmt++);
            continue;
        }
        fmt++;
        if (*fmt == '%') {
            GOPERL_PUTC('%');
            fmt++;
            continue;
        }
        if (strncmp(fmt, SVf, sizeof(SVf) - 1) == 0) {
            SV *sv = va_arg(ap, SV *);
            uint64_t len = 0;
            const char *p = sv ? goperl_sv_pv_(f, sv, &len) : "(null)";
            if (!p) p = "";
            if (!sv) len = strlen(p);
            for (uint64_t i = 0; i < len; i++) GOPERL_PUTC(p[i]);
            fmt += sizeof(SVf) - 1;
            continue;
        }
        /* one standard conversion: %[flags][width][.prec][len]conv; a `*`
         * width/precision is materialized from the vararg into digits. */
        char spec[48];
        size_t sn = 0;
        spec[sn++] = '%';
        while (*fmt && strchr("-+ #0", *fmt) && sn < 8) spec[sn++] = *fmt++;
        if (*fmt == '*') {
            fmt++;
            sn += (size_t)snprintf(spec + sn, sizeof spec - sn, "%d",
                                   va_arg(ap, int));
        } else {
            while (*fmt >= '0' && *fmt <= '9' && sn < 14) spec[sn++] = *fmt++;
        }
        if (*fmt == '.') {
            spec[sn++] = *fmt++;
            if (*fmt == '*') {
                fmt++;
                sn += (size_t)snprintf(spec + sn, sizeof spec - sn, "%d",
                                       va_arg(ap, int));
            } else {
                while (*fmt >= '0' && *fmt <= '9' && sn < 24)
                    spec[sn++] = *fmt++;
            }
        }
        int lmod = 0; /* 0 none, 1 l, 2 ll, 3 z, 4 h */
        if (fmt[0] == 'l' && fmt[1] == 'l') {
            lmod = 2;
            spec[sn++] = *fmt++;
            spec[sn++] = *fmt++;
        } else if (*fmt == 'l') {
            lmod = 1;
            spec[sn++] = *fmt++;
        } else if (*fmt == 'z') {
            lmod = 3;
            spec[sn++] = *fmt++;
        } else if (*fmt == 'h') {
            lmod = 4;
            spec[sn++] = *fmt++;
            if (*fmt == 'h') spec[sn++] = *fmt++;
        }
        char conv = *fmt ? *fmt++ : 0;
        if (!conv) break;
        spec[sn++] = conv;
        spec[sn] = '\0';
        char chunk[512];
        chunk[0] = '\0';
        switch (conv) {
        case 'd':
        case 'i':
            if (lmod == 2) snprintf(chunk, sizeof chunk, spec, va_arg(ap, long long));
            else if (lmod == 1) snprintf(chunk, sizeof chunk, spec, va_arg(ap, long));
            else if (lmod == 3) snprintf(chunk, sizeof chunk, spec, va_arg(ap, size_t));
            else snprintf(chunk, sizeof chunk, spec, va_arg(ap, int));
            break;
        case 'u':
        case 'o':
        case 'x':
        case 'X':
            if (lmod == 2) snprintf(chunk, sizeof chunk, spec, va_arg(ap, unsigned long long));
            else if (lmod == 1) snprintf(chunk, sizeof chunk, spec, va_arg(ap, unsigned long));
            else if (lmod == 3) snprintf(chunk, sizeof chunk, spec, va_arg(ap, size_t));
            else snprintf(chunk, sizeof chunk, spec, va_arg(ap, unsigned int));
            break;
        case 'c':
            snprintf(chunk, sizeof chunk, spec, va_arg(ap, int));
            break;
        case 's': {
            const char *sp = va_arg(ap, const char *);
            snprintf(chunk, sizeof chunk, spec, sp ? sp : "(null)");
            break;
        }
        case 'p':
            snprintf(chunk, sizeof chunk, spec, va_arg(ap, void *));
            break;
        case 'e':
        case 'E':
        case 'f':
        case 'F':
        case 'g':
        case 'G':
            snprintf(chunk, sizeof chunk, spec, va_arg(ap, double));
            break;
        default:
            snprintf(chunk, sizeof chunk, "<%%%c?>", conv);
            break;
        }
        for (const char *cp = chunk; *cp; cp++) GOPERL_PUTC(*cp);
    }
    dst[o < dstsize ? o : dstsize - 1] = '\0';
#undef GOPERL_PUTC
}

static const char *goperl_form(goperl_frame_t *f, const char *fmt, ...)
    __attribute__((unused));
static const char *goperl_form(goperl_frame_t *f, const char *fmt, ...) {
    char *buf = goperl_form_bufs_v[goperl_form_ix_v++ % GOPERL_FORM_BUFS];
    va_list ap;
    va_start(ap, fmt);
    goperl_vfmt(f, buf, GOPERL_FORM_LEN, fmt, ap);
    va_end(ap);
    return buf;
}
#define form(...) goperl_form(_gof, __VA_ARGS__)

/* croak() and warn() (no context arg) are variadic macros over the local
 * frame. Perl_croak/Perl_warn/Perl_warner are called with aTHX_ — which now
 * expands to a real argument — so they must be actual variadic FUNCTIONS
 * taking pTHX_ (a macro would glue `aTHX_ fmt` into one argument). */
#define croak(...) goperl_croakf(_gof, __VA_ARGS__)
static void Perl_croak(pTHX_ const char *fmt, ...)
    __attribute__((noreturn, unused));
static void Perl_croak(pTHX_ const char *fmt, ...) {
    va_list ap;
    va_start(ap, fmt);
    goperl_vfmt(_gof, _gof->err, sizeof _gof->err, fmt, ap);
    va_end(ap);
    goperl_raise(_gof);
}
static void Perl_croak_nocontext(const char *fmt, ...)
    __attribute__((noreturn, unused));
static void Perl_croak_nocontext(const char *fmt, ...) {
    goperl_frame_t *f = goperl_cur_frame_v;
    va_list ap;
    va_start(ap, fmt);
    goperl_vfmt(f, f->err, sizeof f->err, fmt, ap);
    va_end(ap);
    goperl_raise(f);
}

static void goperl_warnf(goperl_frame_t *f, const char *fmt, ...)
    __attribute__((unused));
static void goperl_warnf(goperl_frame_t *f, const char *fmt, ...) {
    char buf[1024];
    va_list ap;
    va_start(ap, fmt);
    goperl_vfmt(f, buf, sizeof buf, fmt, ap);
    va_end(ap);
    goperl_do_op(f, GOPERL_OP_WARN, 0, 0, buf, strlen(buf));
}
#define warn(...) goperl_warnf(_gof, __VA_ARGS__)
static void Perl_warn(pTHX_ const char *fmt, ...) __attribute__((unused));
static void Perl_warn(pTHX_ const char *fmt, ...) {
    char buf[1024];
    va_list ap;
    va_start(ap, fmt);
    goperl_vfmt(_gof, buf, sizeof buf, fmt, ap);
    va_end(ap);
    goperl_do_op(_gof, GOPERL_OP_WARN, 0, 0, buf, strlen(buf));
}
static void Perl_warner(pTHX_ U32 cat, const char *fmt, ...)
    __attribute__((unused));
static void Perl_warner(pTHX_ U32 cat, const char *fmt, ...) {
    char buf[1024];
    va_list ap;
    (void)cat;
    va_start(ap, fmt);
    goperl_vfmt(_gof, buf, sizeof buf, fmt, ap);
    va_end(ap);
    goperl_do_op(_gof, GOPERL_OP_WARN, 0, 0, buf, strlen(buf));
}
#define packWARN(a) (a)
#define WARN_MISC 4
#define WARN_UNINITIALIZED 30

/* ---- basic SV constructors / accessors ---------------------------------- */

#define SvIV(sv) \
    (goperl_flush(_gof), goperl_api_v->sv_iv(_gof, GOPERL_TOK(sv)))
/* SvIVX is an LVALUE in the wild (`SvIVX(counter)++`): host IV mirror. */
#define SvIVX(sv) (*goperl_ivx_ref(_gof, (SV *)(sv)))
#define SvIVx(sv) SvIV(sv)
#define SvUV(sv) ((UV)SvIV(sv))
#define SvUVX(sv) SvUV(sv)
#define SvUVx(sv) SvUV(sv)
#define sv_2iv(sv) ((IV)SvIV(sv))
#define SvIOK_only(sv) ((void)(sv)) /* representation hint; SETIV flush covers it */

static double goperl_sv_nv(goperl_frame_t *f, SV *sv) {
    union {
        uint64_t u;
        double d;
    } cvt;
    cvt.u = goperl_do_op(f, GOPERL_OP_SV_NV, GOPERL_TOK(sv), 0, 0, 0);
    return cvt.d;
}
#define SvNV(sv) goperl_sv_nv(_gof, (SV *)(sv))
#define SvNVx(sv) SvNV(sv)

#define SvPV_nolen(sv) ((char *)goperl_sv_pv_(_gof, (SV *)(sv), 0))
#define SvPV_nolen_const(sv) goperl_sv_pv_(_gof, (SV *)(sv), 0)
#define SvPV(sv, lenvar) \
    ((char *)goperl_sv_pv_(_gof, (SV *)(sv), (uint64_t *)&(lenvar)))
#define SvPV_const(sv, lenvar) \
    goperl_sv_pv_(_gof, (SV *)(sv), (uint64_t *)&(lenvar))
#define SvPVX(sv) SvPV_nolen(sv)
#define SvPVX_const(sv) SvPV_nolen_const(sv)
#define SvCUR(sv) \
    ((STRLEN)(uint32_t)goperl_op0(GOPERL_OP_SV_PV, GOPERL_TOK(sv), 0))
#define SvCUR_set(sv, n) \
    ((void)goperl_op0(GOPERL_OP_SV_CUR_SET, GOPERL_TOK(sv), (uint64_t)(n)))
#define SvEND(sv) (SvPVX(sv) + SvCUR(sv))

static char *goperl_sv_grow(goperl_frame_t *f, SV *sv, STRLEN n) {
    uint64_t gp = goperl_do_op(f, GOPERL_OP_SV_GROW, GOPERL_TOK(sv),
                               (uint64_t)n, 0, 0);
    return (char *)goperl_api_v->guest_mem(f, gp);
}
#define SvGROW(sv, n) goperl_sv_grow(_gof, (SV *)(sv), (STRLEN)(n))
#define sv_grow(sv, n) SvGROW((sv), (n))

#define sv_2mortal(sv)                                        \
    ((SV *)(uintptr_t)(goperl_flush(_gof),                    \
                       goperl_api_v->sv_mortal(_gof, GOPERL_TOK(sv))))
#define newSViv(v)                                            \
    ((SV *)(uintptr_t)(goperl_flush(_gof),                    \
                       goperl_api_v->new_iv(_gof, (int64_t)(v))))
#define newSVpvn(p, l)                                        \
    ((SV *)(uintptr_t)(goperl_flush(_gof),                    \
                       goperl_api_v->new_pvn(_gof, (p), (uint64_t)(l))))
#define newSVpvs(s) newSVpvn("" s "", sizeof(s) - 1)
#define newSVuv(v) GOPERL_SV(goperl_op0(GOPERL_OP_NEW_UV, (uint64_t)(v), 0))
#define newSV(len) GOPERL_SV(goperl_op0(GOPERL_OP_NEW_SV, (uint64_t)(len), 0))
#define newSVsv(sv) GOPERL_SV(goperl_op0(GOPERL_OP_NEW_SVSV, GOPERL_TOK(sv), 0))
#define sv_mortalcopy(sv) \
    GOPERL_SV(goperl_op0(GOPERL_OP_SV_MORTALCOPY, GOPERL_TOK(sv), 0))
#define sv_newmortal() sv_2mortal(newSV(0))

static SV *goperl_newSVnv(goperl_frame_t *f, double d) {
    union {
        uint64_t u;
        double d;
    } cvt;
    cvt.d = d;
    return GOPERL_SV(goperl_do_op(f, GOPERL_OP_NEW_NV, cvt.u, 0, 0, 0));
}
#define newSVnv(v) goperl_newSVnv(_gof, (double)(v))

static SV *goperl_newSVpv(goperl_frame_t *f, const char *s, STRLEN len) {
    if (len == 0) len = s ? strlen(s) : 0;
    goperl_flush(f);
    return (SV *)(uintptr_t)goperl_api_v->new_pvn(f, s ? s : "", (uint64_t)len);
}
#define newSVpv(s, l) goperl_newSVpv(_gof, (s), (STRLEN)(l))

static SV *goperl_newSVpvf(goperl_frame_t *f, const char *fmt, ...)
    __attribute__((unused));
static SV *goperl_newSVpvf(goperl_frame_t *f, const char *fmt, ...) {
    char buf[2048];
    va_list ap;
    va_start(ap, fmt);
    goperl_vfmt(f, buf, sizeof buf, fmt, ap);
    va_end(ap);
    goperl_flush(f);
    return (SV *)(uintptr_t)goperl_api_v->new_pvn(f, buf, strlen(buf));
}
#define newSVpvf(...) goperl_newSVpvf(_gof, __VA_ARGS__)

static SV *goperl_vnewSVpvf(goperl_frame_t *f, const char *fmt, va_list *ap)
    __attribute__((unused));
static SV *goperl_vnewSVpvf(goperl_frame_t *f, const char *fmt, va_list *ap) {
    char buf[2048];
    goperl_vfmt(f, buf, sizeof buf, fmt, *ap);
    goperl_flush(f);
    return (SV *)(uintptr_t)goperl_api_v->new_pvn(f, buf, strlen(buf));
}
#define vnewSVpvf(fmt, ap) goperl_vnewSVpvf(_gof, (fmt), (ap))

static SV *goperl_mess(goperl_frame_t *f, const char *fmt, ...)
    __attribute__((unused));
static SV *goperl_mess(goperl_frame_t *f, const char *fmt, ...) {
    char buf[2048];
    va_list ap;
    va_start(ap, fmt);
    goperl_vfmt(f, buf, sizeof buf, fmt, ap);
    va_end(ap);
    goperl_flush(f);
    return (SV *)(uintptr_t)goperl_api_v->sv_mortal(
        f, goperl_api_v->new_pvn(f, buf, strlen(buf)));
}
#define mess(...) goperl_mess(_gof, __VA_ARGS__)

#define newSVpvn_share(p, l, h) \
    GOPERL_SV(goperl_ops(GOPERL_OP_NEW_PVN_SHARE, 0,                     \
                         (uint64_t)(int64_t)(l), (p),                    \
                         (uint64_t)((int64_t)(l) < 0 ? -(int64_t)(l)     \
                                                     : (int64_t)(l))))
#define newSVpvs_share(s) newSVpvn_share("" s "", sizeof(s) - 1, 0)

static SV *goperl_newSV_type(goperl_frame_t *f, svtype t) {
    SV *sv = GOPERL_SV(goperl_do_op(f, GOPERL_OP_NEW_SV, 0, 0, 0, 0));
    if (t != SVt_NULL)
        goperl_do_op(f, GOPERL_OP_SV_UPGRADE, GOPERL_TOK(sv), (uint64_t)t, 0, 0);
    return sv;
}
#define newSV_type(t) goperl_newSV_type(_gof, (t))

static SV *goperl_newSVpvs_flags_(goperl_frame_t *f, const char *s,
                                  STRLEN len, U32 flags) {
    goperl_flush(f);
    uint64_t tok = goperl_api_v->new_pvn(f, s, (uint64_t)len);
    if (flags & SVs_TEMP) tok = goperl_api_v->sv_mortal(f, tok);
    return GOPERL_SV(tok);
}
#define newSVpvs_flags(s, flags) \
    goperl_newSVpvs_flags_(_gof, "" s "", sizeof(s) - 1, (flags))

/* ---- setters / string ops ---------------------------------------------- */

#define sv_setsv(d, s) \
    ((void)goperl_op0(GOPERL_OP_SV_SETSV, GOPERL_TOK(d), GOPERL_TOK(s)))
#define sv_setsv_nomg(d, s) \
    ((void)goperl_op0(GOPERL_OP_SV_SETSV_NOMG, GOPERL_TOK(d), GOPERL_TOK(s)))
#define sv_setsv_mg(d, s) \
    ((void)goperl_op0(GOPERL_OP_SV_SETSV_MG, GOPERL_TOK(d), GOPERL_TOK(s)))
#define sv_setiv(sv, v) \
    ((void)goperl_op0(GOPERL_OP_SV_SETIV, GOPERL_TOK(sv), (uint64_t)(int64_t)(v)))
#define sv_setuv(sv, v) \
    ((void)goperl_op0(GOPERL_OP_SV_SETUV, GOPERL_TOK(sv), (uint64_t)(v)))
static void goperl_sv_setnv(goperl_frame_t *f, SV *sv, double d) {
    union {
        uint64_t u;
        double d;
    } cvt;
    cvt.d = d;
    goperl_do_op(f, GOPERL_OP_SV_SETNV, GOPERL_TOK(sv), cvt.u, 0, 0);
}
#define sv_setnv(sv, v) goperl_sv_setnv(_gof, (SV *)(sv), (double)(v))
#define sv_setpvn(sv, p, l) \
    ((void)goperl_ops(GOPERL_OP_SV_SETPVN, GOPERL_TOK(sv), (uint64_t)(l), (p), (uint64_t)(l)))
#define sv_setpv(sv, p) sv_setpvn((sv), (p), strlen(p))
#define sv_setpvs(sv, s) sv_setpvn((sv), "" s "", sizeof(s) - 1)
#define sv_catsv(d, s) \
    ((void)goperl_op0(GOPERL_OP_SV_CATSV, GOPERL_TOK(d), GOPERL_TOK(s)))
#define sv_catsv_nomg(d, s) \
    ((void)goperl_op0(GOPERL_OP_SV_CATSV_NOMG, GOPERL_TOK(d), GOPERL_TOK(s)))
#define sv_catpvn(d, p, l) \
    ((void)goperl_ops(GOPERL_OP_SV_CATPVN, GOPERL_TOK(d), (uint64_t)(l), (p), (uint64_t)(l)))
#define sv_catpv(d, p) sv_catpvn((d), (p), strlen(p))
#define sv_catpvs(d, s) sv_catpvn((d), "" s "", sizeof(s) - 1)

/* ---- flags / predicates -------------------------------------------------- */

#define SvFLAGS(sv) ((U32)goperl_op0(GOPERL_OP_SV_INFO, GOPERL_TOK(sv), 0))
#define SvOK(sv) ((SvFLAGS(sv) & SVf_OK) != 0)
#define SvPOK(sv) ((SvFLAGS(sv) & SVf_POK) != 0)
#define SvIOK(sv) ((SvFLAGS(sv) & SVf_IOK) != 0)
#define SvNOK(sv) ((SvFLAGS(sv) & SVf_NOK) != 0)
#define SvNIOK(sv) ((SvFLAGS(sv) & GOPERL_INFO_NIOK) != 0)
#define SvPOKp(sv) ((SvFLAGS(sv) & GOPERL_INFO_POKp) != 0)
#define SvNIOKp(sv) ((SvFLAGS(sv) & GOPERL_INFO_NIOKp) != 0)
#define SvIsUV(sv) ((SvFLAGS(sv) & GOPERL_INFO_ISUV) != 0)
#define SvUTF8(sv) ((SvFLAGS(sv) & GOPERL_INFO_UTF8) != 0)
#define SvREADONLY(sv) ((SvFLAGS(sv) & GOPERL_INFO_READONLY) != 0)
#define SvOBJECT(sv) ((SvFLAGS(sv) & GOPERL_INFO_OBJECT) != 0)
#define SvRMAGICAL(sv) ((SvFLAGS(sv) & GOPERL_INFO_RMAGICAL) != 0)
#define sv_isobject(sv) ((SvFLAGS(sv) & GOPERL_INFO_ISOBJ) != 0)
#define SvTRUE(sv) (goperl_op0(GOPERL_OP_SV_TRUE, GOPERL_TOK(sv), 0) != 0)
#define SvTRUE_nomg(sv) SvTRUE(sv)
#define sv_true(sv) SvTRUE(sv)
#define SvROK(sv) (goperl_op0(GOPERL_OP_SV_RV, GOPERL_TOK(sv), 0) != 0)
#define SvRV(sv) GOPERL_SV(goperl_op0(GOPERL_OP_SV_RV, GOPERL_TOK(sv), 0))
#define SvTYPE(sv) \
    ((svtype)goperl_op0(GOPERL_OP_SV_TYPE, GOPERL_TOK(sv), 0))
#define isGV(sv) (SvTYPE((SV *)(sv)) == SVt_PVGV)
#define SvGETMAGIC(sv) ((void)(sv))
#define SvSETMAGIC(sv) ((void)(sv))
#define SvIV_please(sv) ((void)(sv))
#define SvTAINTED_off(sv) ((void)(sv))
#define TAINT_NOT ((void)0)
#define SvUTF8_on(sv) \
    ((void)goperl_op0(GOPERL_OP_SV_UTF8_ON, GOPERL_TOK(sv), 0))
#define SvPOK_on(sv) \
    ((void)goperl_op0(GOPERL_OP_SV_POK_ON, GOPERL_TOK(sv), 0))
#define SvUPGRADE(sv, t) \
    ((void)goperl_op0(GOPERL_OP_SV_UPGRADE, GOPERL_TOK(sv), (uint64_t)(t)), 0)
#define sv_upgrade(sv, t) \
    ((void)goperl_op0(GOPERL_OP_SV_UPGRADE, GOPERL_TOK(sv), (uint64_t)(t)))
#define sv_utf8_upgrade(sv) \
    ((void)goperl_op0(GOPERL_OP_SV_UTF8_UPGRADE, GOPERL_TOK(sv), 0))
#define sv_len(sv) ((STRLEN)goperl_op0(GOPERL_OP_SV_LEN, GOPERL_TOK(sv), 0))
#define sv_eq(a, b) \
    ((I32)goperl_op0(GOPERL_OP_SV_EQ, GOPERL_TOK(a), GOPERL_TOK(b)))
#define sv_cmp(a, b) \
    ((I32)(int64_t)goperl_op0(GOPERL_OP_SV_CMP, GOPERL_TOK(a), GOPERL_TOK(b)))
#define sv_rvweaken(sv) \
    ((void)goperl_op0(GOPERL_OP_SV_RVWEAKEN, GOPERL_TOK(sv), 0), (SV *)(sv))
#define sv_dump(sv) ((void)goperl_op0(GOPERL_OP_SV_DUMP, GOPERL_TOK(sv), 0))
#define looks_like_number(sv) \
    ((I32)goperl_op0(GOPERL_OP_LOOKS_LIKE_NUMBER, GOPERL_TOK(sv), 0))
#define SvAMAGIC(sv) \
    (goperl_op0(GOPERL_OP_SV_AMAGIC, GOPERL_TOK(sv), 0) != 0)
#define SvSTASH(sv) \
    ((HV *)GOPERL_SV(goperl_op0(GOPERL_OP_SV_STASH, GOPERL_TOK(sv), 0)))

#define SvREFCNT_inc(sv) \
    GOPERL_SV(goperl_op0(GOPERL_OP_REFCNT_INC, GOPERL_TOK(sv), 0))
#define SvREFCNT_inc_NN(sv) SvREFCNT_inc(sv)
#define SvREFCNT_inc_simple(sv) SvREFCNT_inc(sv)
#define SvREFCNT_inc_simple_NN(sv) SvREFCNT_inc(sv)
#define SvREFCNT_inc_void(sv) ((void)SvREFCNT_inc(sv))
#define SvREFCNT_inc_simple_void(sv) ((void)SvREFCNT_inc(sv))
#define SvREFCNT_inc_simple_void_NN(sv) ((void)SvREFCNT_inc(sv))
#define SvREFCNT_dec(sv) \
    ((void)goperl_op0(GOPERL_OP_SV_REFCNT_DEC, GOPERL_TOK(sv), 0))
#define SvREFCNT_dec_NN(sv) SvREFCNT_dec(sv)

#define newRV_inc(sv) \
    GOPERL_SV(goperl_op0(GOPERL_OP_NEW_RV_INC, GOPERL_TOK(sv), 0))
#define newRV(sv) newRV_inc(sv)
static SV *goperl_newRV_noinc(goperl_frame_t *f, SV *sv) {
    /* rv takes the caller's reference: inc via newRV then drop the extra. */
    SV *rv = GOPERL_SV(
        goperl_do_op(f, GOPERL_OP_NEW_RV_INC, GOPERL_TOK(sv), 0, 0, 0));
    goperl_do_op(f, GOPERL_OP_SV_REFCNT_DEC, GOPERL_TOK(sv), 0, 0, 0);
    return rv;
}
#define newRV_noinc(sv) goperl_newRV_noinc(_gof, (SV *)(sv))

#define amagic_call(sv, right, method, flags)                             \
    GOPERL_SV(goperl_op0(GOPERL_OP_AMAGIC_CALL,                           \
                         ((uint64_t)(uint32_t)(method) << 32) |           \
                             (uint32_t)GOPERL_TOK(sv),                    \
                         (uint64_t)(flags)))
#define amagic_deref_call(ref, method) \
    goperl_amagic_deref_call(_gof, (SV *)(ref), (method))
static SV *goperl_amagic_deref_call(goperl_frame_t *f, SV *ref, int method)
    __attribute__((unused));
static SV *goperl_amagic_deref_call(goperl_frame_t *f, SV *ref, int method) {
    SV *tmpsv = 0;
    for (;;) {
        if (!goperl_do_op(f, GOPERL_OP_SV_AMAGIC, GOPERL_TOK(ref), 0, 0, 0))
            break;
        SV *r = GOPERL_SV(goperl_do_op(
            f, GOPERL_OP_AMAGIC_CALL,
            ((uint64_t)(uint32_t)method << 32) | (uint32_t)GOPERL_TOK(ref),
            (uint64_t)(AMGf_noright | AMGf_unary), 0, 0));
        if (!r) break;
        tmpsv = r;
        if (!goperl_do_op(f, GOPERL_OP_SV_RV, GOPERL_TOK(tmpsv), 0, 0, 0))
            goperl_croakf(f,
                          "Overloaded dereference did not return a reference");
        if (tmpsv == ref) return tmpsv;
        {
            uint64_t ra = goperl_do_op(f, GOPERL_OP_SV_RV, GOPERL_TOK(tmpsv),
                                       0, 0, 0);
            uint64_t rb = goperl_do_op(f, GOPERL_OP_SV_RV, GOPERL_TOK(ref), 0,
                                       0, 0);
            if (ra == rb) return tmpsv;
        }
        ref = tmpsv;
    }
    return tmpsv ? tmpsv : ref;
}

/* ---- AV / HV ------------------------------------------------------------ */

#define newAV() ((AV *)GOPERL_SV(goperl_op0(GOPERL_OP_NEW_AV, 0, 0)))
#define newHV() ((HV *)GOPERL_SV(goperl_op0(GOPERL_OP_NEW_HV, 0, 0)))
#define newHVhv(hv) \
    ((HV *)GOPERL_SV(goperl_op0(GOPERL_OP_NEW_HVHV, GOPERL_TOK(hv), 0)))
#define av_push(av, sv) \
    ((void)goperl_op0(GOPERL_OP_AV_PUSH, GOPERL_TOK(av), GOPERL_TOK(sv)))
#define av_len(av) \
    ((SSize_t)(int64_t)goperl_op0(GOPERL_OP_AV_LEN, GOPERL_TOK(av), 0))
#define AvFILL(av) av_len(av)
#define AvFILLp(av) ((I32)av_len(av))
#define av_extend(av, n) \
    ((void)goperl_op0(GOPERL_OP_AV_EXTEND, GOPERL_TOK(av), (uint64_t)(int64_t)(n)))
#define av_fill(av, n) \
    ((void)goperl_op0(GOPERL_OP_AV_FILL, GOPERL_TOK(av), (uint64_t)(int64_t)(n)))
#define av_store(av, i, sv)                                                    \
    ((void)goperl_op0(GOPERL_OP_AV_STORE, GOPERL_TOK(av),                      \
                      ((uint64_t)(uint32_t)(int32_t)(i) << 32) |               \
                          (uint32_t)GOPERL_TOK(sv)),                           \
     (SV **)0)
#define AvREAL_on(av) ((void)(av))
#define AvREAL_only(av) ((void)(av))
#define AvREIFY_off(av) ((void)(av))

static SV **goperl_tmp_slot(goperl_frame_t *f, uint64_t tok) {
    uint64_t *slot = &f->tmp[(f->tmpidx++) % GOPERL_XS_TMPS];
    *slot = tok;
    return (SV **)slot;
}

static SV **goperl_av_fetch(goperl_frame_t *f, AV *av, SSize_t i, I32 lval) {
    uint64_t tok = goperl_do_op(f, lval ? GOPERL_OP_AV_READ : GOPERL_OP_AV_FETCH,
                                GOPERL_TOK(av), (uint64_t)(int64_t)i, 0, 0);
    if (!tok) return 0;
    return goperl_tmp_slot(f, tok);
}
#define av_fetch(av, i, lval) goperl_av_fetch(_gof, (AV *)(av), (i), (lval))

#define av_make(n, ary) goperl_av_make(_gof, (SSize_t)(n), (ary))
static AV *goperl_av_make(goperl_frame_t *f, SSize_t n, SV **ary) {
    char buf[64 * 4];
    if (n < 0 || n > 64)
        goperl_croakf(f, "av_make: %lld elements exceed the SDK limit",
                      (long long)n);
    for (SSize_t i = 0; i < n; i++) {
        uint32_t tok = (uint32_t)GOPERL_TOK(ary[i]);
        memcpy(buf + i * 4, &tok, 4);
    }
    return (AV *)GOPERL_SV(goperl_do_op(f, GOPERL_OP_AV_MAKE, 0,
                                        (uint64_t)(n * 4), buf,
                                        (uint64_t)(n * 4)));
}

static SV **goperl_hv_fetch(goperl_frame_t *f, HV *hv, const char *key,
                            I32 klen, I32 lval) {
    (void)lval;
    uint64_t tok = goperl_do_op(f, GOPERL_OP_HV_FETCH, GOPERL_TOK(hv), 0, key,
                                (uint64_t)(uint32_t)klen);
    if (!tok && lval) {
        /* autovivify like the real lval fetch: store a fresh undef */
        uint64_t nsv = goperl_do_op(f, GOPERL_OP_NEW_SV, 0, 0, 0, 0);
        goperl_do_op(f, GOPERL_OP_HV_STORE, GOPERL_TOK(hv), nsv, key,
                     (uint64_t)(uint32_t)klen);
        tok = goperl_do_op(f, GOPERL_OP_HV_FETCH, GOPERL_TOK(hv), 0, key,
                           (uint64_t)(uint32_t)klen);
    }
    if (!tok) return 0;
    return goperl_tmp_slot(f, tok);
}
#define hv_fetch(hv, key, klen, lval) \
    goperl_hv_fetch(_gof, (HV *)(hv), (key), (I32)(klen), (lval))
/* hv_fetchs is 3-arg (hv, literal, lval) in perl proper, but some dists
 * use it 2-arg; dispatch on argument count. */
#define GOPERL_HV_FETCHS_3(hv, key, lval) \
    goperl_hv_fetch(_gof, (HV *)(hv), ("" key ""), (I32)(sizeof(key) - 1), (lval))
#define GOPERL_HV_FETCHS_2(hv, key) GOPERL_HV_FETCHS_3(hv, key, 0)
#define GOPERL_HV_FETCHS_GET(_1, _2, _3, NAME, ...) NAME
#define hv_fetchs(...)                                                 \
    GOPERL_HV_FETCHS_GET(__VA_ARGS__, GOPERL_HV_FETCHS_3,              \
                         GOPERL_HV_FETCHS_2)(__VA_ARGS__)
#define hv_store(hv, key, klen, sv, hash)                              \
    ((void)goperl_ops(GOPERL_OP_HV_STORE, GOPERL_TOK(hv),              \
                      GOPERL_TOK(sv), (key), (uint64_t)(klen)),        \
     (SV **)0)
#define hv_stores(hv, key, sv) hv_store((hv), "" key "", sizeof(key) - 1, (sv), 0)
#define hv_fetch_ent(hv, keysv, lval, hash)                              \
    ((HE *)(uintptr_t)goperl_op0(                                        \
        GOPERL_OP_HV_FETCH_ENT, GOPERL_TOK(hv),                          \
        (uint32_t)GOPERL_TOK(keysv) |                                    \
            ((uint64_t)((lval) ? 1 : 0) << 32)))
#define hv_store_ent(hv, keysv, sv, hash)                                 \
    ((HE *)(uintptr_t)goperl_op0(GOPERL_OP_HV_STORE_ENT, GOPERL_TOK(hv),  \
                                 (GOPERL_TOK(keysv) << 32) |              \
                                     (uint32_t)GOPERL_TOK(sv)))
#define hv_exists_ent(hv, keysv, hash)                            \
    ((I32)goperl_op0(GOPERL_OP_HV_EXISTS_ENT, GOPERL_TOK(hv),     \
                     GOPERL_TOK(keysv)))
#define HeVAL(he) \
    (*goperl_tmp_slot(_gof, goperl_op0(GOPERL_OP_HE_VAL, GOPERL_TOK(he), 0)))
#define hv_iterinit(hv) \
    ((I32)goperl_op0(GOPERL_OP_HV_ITERINIT, GOPERL_TOK(hv), 0))
#define HvKEYS(hv) hv_iterinit(hv)
#define hv_iternext(hv) \
    ((HE *)(uintptr_t)goperl_op0(GOPERL_OP_HV_ITERNEXT, GOPERL_TOK(hv), 0))
#define hv_iterkeysv(he) \
    GOPERL_SV(goperl_op0(GOPERL_OP_HV_ITERKEYSV, GOPERL_TOK(he), 0))
#define hv_iterval(hv, he) \
    GOPERL_SV(goperl_op0(GOPERL_OP_HV_ITERVAL, GOPERL_TOK(hv), GOPERL_TOK(he)))
#define HvNAME(hv) "(stash)"

/* ---- stashes / globals / blessing --------------------------------------- */

#define gv_stashpv(name, flags)                                          \
    ((HV *)GOPERL_SV(goperl_ops(GOPERL_OP_GV_STASHPV,                    \
                                (uint64_t)(int64_t)(flags), 0, (name),   \
                                strlen(name))))
#define gv_stashpvs(name, flags) gv_stashpv("" name "", (flags))
#define sv_bless(rv, stash) \
    GOPERL_SV(goperl_op0(GOPERL_OP_SV_BLESS, GOPERL_TOK(rv), GOPERL_TOK(stash)))
#define sv_isa(sv, name) \
    ((int)goperl_ops(GOPERL_OP_SV_ISA, GOPERL_TOK(sv), 0, (name), strlen(name)))
#define sv_derived_from(sv, name)                                          \
    ((int)goperl_ops(GOPERL_OP_SV_DERIVED_FROM, GOPERL_TOK(sv), 0, (name), \
                     strlen(name)))
#define get_sv(name, flags)                                              \
    GOPERL_SV(goperl_ops(GOPERL_OP_GET_SV, (uint64_t)(int64_t)(flags), 0, \
                         (name), strlen(name)))
#define get_cv(name, flags)                                              \
    ((CV *)GOPERL_SV(goperl_ops(GOPERL_OP_GET_CV,                        \
                                (uint64_t)(int64_t)(flags), 0, (name),   \
                                strlen(name))))
#define get_cvs(name, flags) get_cv("" name "", (flags))
#define get_hv(name, flags)                                              \
    ((HV *)GOPERL_SV(goperl_ops(GOPERL_OP_GET_HV,                        \
                                (uint64_t)(int64_t)(flags), 0, (name),   \
                                strlen(name))))
#define gv_fetchpv(name, flags, svt)                                     \
    ((GV *)GOPERL_SV(goperl_ops(GOPERL_OP_GV_FETCH,                      \
                                (uint64_t)(int64_t)(flags),              \
                                (uint64_t)(int64_t)(svt), (name),        \
                                strlen(name))))
#define gv_fetchpvs(name, flags, svt) gv_fetchpv("" name "", (flags), (svt))

static SV *goperl_sv_2cv(goperl_frame_t *f, SV *sv, HV **stash, GV **gvp,
                         I32 lref) {
    if (stash) *stash = 0;
    if (gvp) *gvp = 0;
    return GOPERL_SV(goperl_do_op(f, GOPERL_OP_SV_2CV, GOPERL_TOK(sv),
                                  (uint64_t)(int64_t)lref, 0, 0));
}
#define sv_2cv(sv, stash, gvp, lref) \
    ((CV *)goperl_sv_2cv(_gof, (SV *)(sv), (stash), (gvp), (lref)))

/* T_PTROBJ: guest IVs are 32-bit, so native object pointers live in the
 * loader's registry and only the id crosses. sv_setref_pv/INT2PTR are the
 * two ends of that path; PTR2IV exists for symmetry. */
#define sv_setref_pv(rv, classname, ptr)                                  \
    GOPERL_SV(goperl_ops(GOPERL_OP_SETREF_IV, GOPERL_TOK(rv),             \
                         goperl_api_v->ptr_encode(_gof, (void *)(ptr)),   \
                         (classname), strlen(classname)))
#define sv_setref_iv(rv, classname, iv)                                   \
    GOPERL_SV(goperl_ops(GOPERL_OP_SETREF_IV, GOPERL_TOK(rv),             \
                         (uint64_t)(int64_t)(iv), (classname),            \
                         strlen(classname)))
#define INT2PTR(type, iv) ((type)goperl_api_v->ptr_decode(_gof, (uint64_t)(iv)))
#define PTR2IV(p) ((IV)goperl_api_v->ptr_encode(_gof, (void *)(p)))
#define PTR2UV(p) ((UV)PTR2IV(p))

/* ---- interpreter variables ---------------------------------------------- */

static SV **goperl_immortal(goperl_frame_t *f, int ix /* 0 undef 1 yes 2 no */) {
    if (!f->imm[ix])
        f->imm[ix] = goperl_do_op(f, GOPERL_OP_PLVAR_GET,
                                  (uint64_t)(GOPERL_PL_SV_UNDEF + ix), 0, 0, 0);
    return (SV **)&f->imm[ix];
}
#define PL_sv_undef (**goperl_immortal(_gof, 0))
#define PL_sv_yes (**goperl_immortal(_gof, 1))
#define PL_sv_no (**goperl_immortal(_gof, 2))
#define boolSV(b) ((b) ? &PL_sv_yes : &PL_sv_no)
#define ERRSV GOPERL_SV(goperl_op0(GOPERL_OP_ERRSV, 0, 0))

/* PL_diehook / PL_warnhook: read-through, write-back shadows. Reads always
 * consult the guest unless an assignment is pending; assignments are
 * flushed before the next guest operation (goperl_flush). */
static SV **goperl_hook_ref(goperl_frame_t *f, int which /* 0 die, 1 warn */) {
    if (!(f->hook_dirty & (1 << which)))
        f->hook_val[which] = goperl_api_v->xs_op(
            f, GOPERL_OP_PLVAR_GET,
            (uint64_t)(which == 0 ? GOPERL_PL_DIEHOOK : GOPERL_PL_WARNHOOK), 0,
            0, 0);
    f->hook_dirty |= (1 << which); /* the caller may assign through this */
    return (SV **)&f->hook_val[which];
}
#define PL_diehook (*goperl_hook_ref(_gof, 0))
#define PL_warnhook (*goperl_hook_ref(_gof, 1))

/* PL_amagic_generation is gone from modern perl (overload caches are
 * invalidated automatically); a dummy keeps `PL_amagic_generation++`
 * compiling as the harmless hint it now is. */
__attribute__((weak)) IV goperl_amagic_generation_v = 0;
#define PL_amagic_generation goperl_amagic_generation_v

#define PL_op (_gof->plop)

/* ---- the flush: pending host state -> guest, before any guest op -------- */

static void goperl_avmirror_flush(goperl_frame_t *f) {
    for (int32_t m = 0; m < goperl_avmirror_n; m++) {
        goperl_avmirror_t *mir = &goperl_avmirrors_v[m];
        for (int32_t i = 0; i < mir->len; i++) {
            if (mir->buf[i] != mir->shadow[i]) {
                goperl_api_v->xs_op(f, GOPERL_OP_AV_STORE_RAW, mir->avtok,
                                    ((uint64_t)(uint32_t)i << 32) |
                                        (uint32_t)mir->buf[i],
                                    0, 0);
                mir->shadow[i] = mir->buf[i];
            }
        }
    }
}

static void goperl_ivmirror_flush(goperl_frame_t *f) {
    for (int32_t i = 0; i < goperl_ivmirror_n; i++) {
        goperl_ivmirror_t *m = &goperl_ivmirrors_v[i];
        if (m->val != m->shadow) {
            goperl_api_v->xs_op(f, GOPERL_OP_SV_SETIV, m->svtok,
                                (uint64_t)m->val, 0, 0);
            m->shadow = m->val;
        }
    }
}

static void goperl_flush(goperl_frame_t *f) {
    if (f->hook_dirty) {
        int32_t d = f->hook_dirty;
        f->hook_dirty = 0;
        if (d & 1)
            goperl_api_v->xs_op(f, GOPERL_OP_PLVAR_SET,
                                (uint64_t)GOPERL_PL_DIEHOOK, f->hook_val[0], 0,
                                0);
        if (d & 2)
            goperl_api_v->xs_op(f, GOPERL_OP_PLVAR_SET,
                                (uint64_t)GOPERL_PL_WARNHOOK, f->hook_val[1],
                                0, 0);
    }
    goperl_ivmirror_flush(f);
    goperl_avmirror_flush(f);
}

/* The SvIVX lvalue proxy: refreshes (after flushing pending state) and
 * returns a host slot whose changes flush back before the next guest op. */
static IV *goperl_ivx_ref(goperl_frame_t *f, SV *sv) {
    uint64_t tok = GOPERL_TOK(sv);
    goperl_ivmirror_t *m = 0;
    for (int32_t i = 0; i < goperl_ivmirror_n; i++)
        if (goperl_ivmirrors_v[i].svtok == tok) {
            m = &goperl_ivmirrors_v[i];
            break;
        }
    if (!m) {
        if (goperl_ivmirror_n >= GOPERL_IVMIRROR_MAX)
            goperl_croakf(f, "SvIVX: too many live IV mirrors");
        m = &goperl_ivmirrors_v[goperl_ivmirror_n++];
        m->svtok = tok;
    }
    goperl_flush(f);
    m->val = (IV)goperl_api_v->sv_iv(f, tok);
    m->shadow = m->val;
    return &m->val;
}

/* ---- AvARRAY mirrors ----------------------------------------------------- */

static SV **goperl_av_array(goperl_frame_t *f, AV *av) {
    uint64_t tok = GOPERL_TOK(av);
    goperl_avmirror_t *mir = 0;
    for (int32_t i = 0; i < goperl_avmirror_n; i++)
        if (goperl_avmirrors_v[i].avtok == tok) {
            mir = &goperl_avmirrors_v[i];
            break;
        }
    if (!mir) {
        if (goperl_avmirror_n >= GOPERL_AVMIRROR_MAX)
            goperl_croakf(f, "AvARRAY: too many live array mirrors");
        mir = &goperl_avmirrors_v[goperl_avmirror_n++];
        mir->avtok = tok;
        mir->cap = 64;
        mir->buf = (uint64_t *)malloc((size_t)mir->cap * 8);
        mir->shadow = (uint64_t *)malloc((size_t)mir->cap * 8);
        mir->len = 0;
    }
    goperl_flush(f); /* host writes first, then refresh from the guest */
    int32_t len =
        (int32_t)(int64_t)goperl_api_v->xs_op(f, GOPERL_OP_AV_LEN, tok, 0, 0, 0) +
        1;
    if (len > mir->cap) {
        int32_t ncap = mir->cap;
        while (ncap < len) ncap *= 2;
        mir->buf = (uint64_t *)realloc(mir->buf, (size_t)ncap * 8);
        mir->shadow = (uint64_t *)realloc(mir->shadow, (size_t)ncap * 8);
        mir->cap = ncap;
    }
    for (int32_t i = 0; i < len; i++) {
        mir->buf[i] = goperl_api_v->xs_op(f, GOPERL_OP_AV_FETCH, tok,
                                          (uint64_t)(int64_t)i, 0, 0);
        mir->shadow[i] = mir->buf[i];
    }
    mir->len = len;
    return (SV **)mir->buf;
}
#define AvARRAY(av) goperl_av_array(_gof, (AV *)(av))

/* sortsv over a mirror (or any host SV* array): perl's stable sort is
 * approximated with insertion sort — the arrays this surface sorts are
 * method/key lists. The mirror flush propagates the permutation. */
typedef I32 (*SVCOMPARE_t)(pTHX_ SV *a, SV *b);
static void goperl_sortsv(goperl_frame_t *f, SV **array, size_t n,
                          SVCOMPARE_t cmp) {
    for (size_t i = 1; i < n; i++) {
        SV *key = array[i];
        size_t j = i;
        while (j > 0 && cmp(f, array[j - 1], key) > 0) {
            array[j] = array[j - 1];
            j--;
        }
        array[j] = key;
    }
}
#define sortsv(a, n, c) goperl_sortsv(_gof, (a), (size_t)(n), (c))
static I32 goperl_sv_cmp_fn(pTHX_ SV *a, SV *b) {
    return (I32)(int64_t)goperl_op0(GOPERL_OP_SV_CMP, GOPERL_TOK(a),
                                    GOPERL_TOK(b));
}
#define Perl_sv_cmp goperl_sv_cmp_fn

/* ---- the perl stack ------------------------------------------------------ */

#define PL_stack_base ((SV **)_gof->st)
#define PL_stack_sp (_gof->psp)

static I32 goperl_popmark(goperl_frame_t *f) {
    return f->markidx >= 0 ? f->marks[f->markidx--] : 0;
}
static void goperl_pushmark(goperl_frame_t *f, I32 off) {
    if (f->markidx + 1 >= GOPERL_XS_MARKS)
        goperl_croakf(f, "mark stack overflow in native XS");
    f->marks[++f->markidx] = off;
}
static void goperl_stack_extend(goperl_frame_t *f, SV **sp, SSize_t n) {
    if ((sp - (SV **)f->st) + n >= GOPERL_XS_STACK)
        goperl_croakf(f, "perl stack overflow in native XS (max %d)",
                      GOPERL_XS_STACK);
}

#define dSP SV **sp = PL_stack_sp
#define SP sp
#define MARK mark
#define POPMARK goperl_popmark(_gof)
#define TOPMARK (_gof->marks[_gof->markidx])
#define PUSHMARK(p) goperl_pushmark(_gof, (I32)((p)-PL_stack_base))
#define dMARK SV **mark = PL_stack_base + POPMARK
#define dORIGMARK const I32 origmark = (I32)(mark - PL_stack_base)
#define ORIGMARK (PL_stack_base + origmark)
#define dAX const I32 ax = (I32)(mark - PL_stack_base + 1)
#define dAXMARK             \
    I32 ax = POPMARK + 1;   \
    SV **mark = PL_stack_base + ax - 1
#define dITEMS I32 items = (I32)(sp - mark)
#define EXTEND(p, n) goperl_stack_extend(_gof, (p), (SSize_t)(n))
#define PUSHs(s) (*++sp = (SV *)(s))
#define XPUSHs(s)                 \
    STMT_START {                  \
        EXTEND(sp, 1);            \
        *++sp = (SV *)(s);        \
    }                             \
    STMT_END
#define POPs (*sp--)
#define TOPs (*sp)
#define PUTBACK PL_stack_sp = sp
#define SPAGAIN sp = PL_stack_sp
#define mPUSHs(s) PUSHs(sv_2mortal(s))
#define mXPUSHs(s) XPUSHs(sv_2mortal(s))
#define mPUSHi(i) PUSHs(sv_2mortal(newSViv(i)))
#define mPUSHu(u) PUSHs(sv_2mortal(newSVuv(u)))
#define mPUSHn(n) PUSHs(sv_2mortal(newSVnv(n)))
#define mPUSHp(p, l) PUSHs(sv_2mortal(newSVpvn((p), (l))))
#define PUSHi(i) PUSHs(sv_2mortal(newSViv(i)))
#define PUSHu(u) PUSHs(sv_2mortal(newSVuv(u)))
#define PUSHn(n) PUSHs(sv_2mortal(newSVnv(n)))
#define PUSHp(p, l) PUSHs(sv_2mortal(newSVpvn((p), (l))))

/* ---- xsubpp entry/exit macros ------------------------------------------- */

#define STATIC static
#define STATIC_INLINE static inline
#define NOOP ((void)0)
#define dNOOP ((void)0)
#define dVAR dNOOP
#define PERL_ASYNC_CHECK() NOOP
#define strNE(a, b) (strcmp((a), (b)) != 0)
#define UTF8_IS_INVARIANT(c) ((U8)(c) < 0x80)
#define UTF8_EIGHT_BIT_HI(c) ((U8)(0xC0 | ((U8)(c) >> 6)))
#define UTF8_EIGHT_BIT_LO(c) ((U8)(0x80 | ((U8)(c)&0x3F)))

/* PerlIO: the SDK routes prints to the host's stdio, through the SVf-aware
 * formatter (fprintf itself cannot digest the SVf marker). */
#define PerlIO_stderr() stderr
#define PerlIO_stdout() stdout
static void goperl_perlio_printf(FILE *fp, const char *fmt, ...)
    __attribute__((unused));
static void goperl_perlio_printf(FILE *fp, const char *fmt, ...) {
    char buf[2048];
    va_list ap;
    va_start(ap, fmt);
    goperl_vfmt(goperl_cur_frame_v, buf, sizeof buf, fmt, ap);
    va_end(ap);
    fputs(buf, fp);
}
#define PerlIO_printf goperl_perlio_printf
#define PerlIO_stdoutf(...) goperl_perlio_printf(stdout, __VA_ARGS__)

/* XCPT: local exception catching (the NO_XSLOCKS surface). A croak inside
 * the TRY block longjmps to the local buffer instead of the XSUB entry. */
#define dXCPT                        \
    jmp_buf goperl_xcpt_jb;          \
    void *goperl_xcpt_prev = 0;      \
    int goperl_xcpt_rc = 0
#define XCPT_TRY_START                               \
    goperl_xcpt_prev = _gof->jb;                     \
    _gof->jb = (void *)&goperl_xcpt_jb;              \
    goperl_xcpt_rc = setjmp(goperl_xcpt_jb);         \
    if (goperl_xcpt_rc == 0)
#define XCPT_TRY_END                 \
    _gof->jb = goperl_xcpt_prev;     \
    if (goperl_xcpt_rc) _gof->failed = 0;
#define XCPT_CATCH if (goperl_xcpt_rc != 0)
/* The caught message still sits in the frame's err buffer. */
#define XCPT_RETHROW goperl_croakf(_gof, "%s", _gof->err)
#define PERL_UNUSED_VAR(v) ((void)(v))
#define PERL_UNUSED_ARG(v) ((void)(v))
#define PERL_UNUSED_DECL __attribute__((unused))
#define PERL_UNUSED_CONTEXT ((void)_gof)
#define STMT_START do
#define STMT_END while (0)
#ifndef LIKELY
#define LIKELY(x) (!!(x))
#define UNLIKELY(x) (!!(x))
#endif
#ifdef __cplusplus
#define EXTERN_C extern "C"
#else
#define EXTERN_C extern
#endif
#define CAT2_(a, b) a##b
#define CAT2(a, b) CAT2_(a, b)
#define STRINGIFY_(n) #n
#define STRINGIFY(n) STRINGIFY_(n)
#define STR_WITH_LEN(s) ("" s ""), (sizeof(s) - 1)
#define CALL_FPTR(fptr) (*(fptr))

#define XSPROTO(name) void name(pTHX_ CV *cv)
#define XS(name) XSPROTO(name)
#define XS_EXTERNAL(name) \
    __attribute__((visibility("default"))) void name(pTHX_ CV *cv)
#define XS_INTERNAL(name) static void name(pTHX_ CV *cv)

static void goperl_xs_leave(goperl_frame_t *f) {
    goperl_flush(f); /* pending mirror writes must land before teardown */
    /* Host-location saves (SAVEVPTR/save_hptr) made at the XSUB's top-level
     * scope have no host LEAVE to restore them; real perl restores them when
     * the CALLER's scope pops, which the host cannot observe. Restoring at
     * XSUB return is the closest safe point — the saved locations typically
     * point into this activation's now-dying C stack. */
    goperl_hostsave_unwind_to(f, f->hostsave_base);
    if (--goperl_xs_depth_v <= 0) {
        goperl_xs_depth_v = 0;
        for (int32_t i = 0; i < goperl_avmirror_n; i++) {
            free(goperl_avmirrors_v[i].buf);
            free(goperl_avmirrors_v[i].shadow);
            goperl_avmirrors_v[i].buf = 0;
            goperl_avmirrors_v[i].shadow = 0;
            goperl_avmirrors_v[i].avtok = 0;
        }
        goperl_avmirror_n = 0;
        goperl_ivmirror_n = 0;
        /* a croak path may leave the scope counter high; at rest it is 0 */
        goperl_scope_v = 0;
    }
}

#define dXSARGS                              \
    jmp_buf _gof_jb;                         \
    _gof->jb = (void *)&_gof_jb;             \
    goperl_cur_frame_v = _gof;               \
    goperl_xs_depth_v++;                     \
    _gof->hostsave_base = goperl_hostsave_n; \
    if (setjmp(_gof_jb)) {                   \
        goperl_hostsave_unwind_to(_gof, _gof->hostsave_base); \
        goperl_xs_leave(_gof);               \
        return;                              \
    }                                        \
    dSP;                                     \
    dAXMARK;                                 \
    dITEMS

#define dXSTARG SV *targ __attribute__((unused)) = sv_2mortal(newSV(0))
#define TARG targ
#define PUSHTARG PUSHs(TARG)
#define XSprePUSH (sp = PL_stack_base + ax - 1)
#define XSRETURN(off)                                       \
    STMT_START {                                            \
        PL_stack_sp = PL_stack_base + ax + (off)-1;         \
        goperl_xs_leave(_gof);                              \
        return;                                             \
    }                                                       \
    STMT_END
#define XSRETURN_EMPTY XSRETURN(0)
#define XSRETURN_YES                  \
    STMT_START {                      \
        ST(0) = &PL_sv_yes;           \
        XSRETURN(1);                  \
    }                                 \
    STMT_END
#define XSRETURN_NO                   \
    STMT_START {                      \
        ST(0) = &PL_sv_no;            \
        XSRETURN(1);                  \
    }                                 \
    STMT_END
#define XSRETURN_UNDEF                \
    STMT_START {                      \
        ST(0) = &PL_sv_undef;         \
        XSRETURN(1);                  \
    }                                 \
    STMT_END
#define ST(n) (PL_stack_base[ax + (n)])

/* Predefining the assert guard makes xsubpp skip its own S_croak_xs_usage
 * fallback (which needs CvGV/GvNAME/HvNAME internals this SDK does not
 * model); the SDK's version croaks with the frame's qualified sub name. */
#define PERL_ARGS_ASSERT_CROAK_XS_USAGE
#define croak_xs_usage(cv_ignored, params) \
    goperl_croakf(_gof, "Usage: %s(%s)", _gof->subname, params)
#define CvGV(cv) ((GV *)(cv))
#define GvNAME(gv) (_gof->subname)

/* CvXSUBANY / XSANY / alias dispatch: per-CV host storage via the loader. */
typedef union goperl_any {
    void *any_ptr;
    IV any_iv;
    I32 any_i32;
} ANY;
#define CvXSUBANY(cv) \
    (*(ANY *)goperl_api_v->cv_any(_gof, GOPERL_TOK(cv)))
#define XSANY CvXSUBANY(cv)
#define dXSI32 I32 ix = XSANY.any_i32
#define CvXSUB(cv) \
    ((void (*)(pTHX_ CV *))goperl_api_v->cv_xsub(_gof, GOPERL_TOK(cv)))

/* newXS family: every native XSUB is registered through the loader, which
 * assigns the dispatch id and returns the guest CV token. Prototypes and
 * file names are not modeled. */
static CV *goperl_newXS(goperl_frame_t *f, const char *name,
                        void (*fn)(pTHX_ CV *)) {
    goperl_flush(f);
    return (CV *)(uintptr_t)goperl_api_v->new_xs(f, name, (void *)fn);
}
#define newXS(name, fn, file) goperl_newXS(_gof, (name), (fn))
#define newXS_deffile(name, fn) goperl_newXS(_gof, (name), (fn))
#define newXS_flags(name, fn, file, proto, flags) goperl_newXS(_gof, (name), (fn))
#define newXSproto(name, fn, file, proto) goperl_newXS(_gof, (name), (fn))
#define newXSproto_portable(name, fn, file, proto) goperl_newXS(_gof, (name), (fn))
#define Perl_newXS_deffile(name, fn) goperl_newXS(_gof, (name), (fn))

/* boot support */
#define XS_APIVERSION_BOOTCHECK dNOOP
#define XS_VERSION_BOOTCHECK dNOOP
#define dXSBOOTARGSXSAPIVERCHK dXSARGS
#define dXSBOOTARGSAPIVERCHK dXSARGS
/* Called as Perl_xs_boot_epilog(aTHX_ ax): aTHX_ folds into the single
 * macro argument, so this must swallow whatever it gets. */
#define Perl_xs_boot_epilog(...) dNOOP
#define DECL_BOOT(name) EXTERN_C XS(CAT2(boot_, name))
#define CALL_BOOT(name)                             \
    STMT_START {                                    \
        PUSHMARK(SP);                               \
        CALL_FPTR(CAT2(boot_, name))(aTHX_ cv);     \
    }                                               \
    STMT_END

/* ---- scopes, saves, calls ------------------------------------------------ */

#define ENTER                                        \
    STMT_START {                                     \
        goperl_scope_v++;                            \
        goperl_op0(GOPERL_OP_ENTER, 0, 0);           \
    }                                                \
    STMT_END
static void goperl_leave(goperl_frame_t *f) {
    goperl_do_op(f, GOPERL_OP_LEAVE, 0, 0, 0, 0);
    while (goperl_hostsave_n > 0 &&
           goperl_hostsave_v[goperl_hostsave_n - 1].scope >= goperl_scope_v) {
        goperl_hostsave_t *e = &goperl_hostsave_v[--goperl_hostsave_n];
        *e->loc = e->val;
    }
    goperl_scope_v--;
}
#define LEAVE goperl_leave(_gof)
#define SAVETMPS ((void)goperl_op0(GOPERL_OP_SAVETMPS, 0, 0))
#define FREETMPS ((void)goperl_op0(GOPERL_OP_FREETMPS, 0, 0))

static void goperl_hostsave_ptr(goperl_frame_t *f, void **loc) {
    if (goperl_hostsave_n >= GOPERL_HOSTSAVE_MAX)
        goperl_croakf(f, "host save stack overflow in native XS");
    goperl_hostsave_t *e = &goperl_hostsave_v[goperl_hostsave_n++];
    e->loc = loc;
    e->val = *loc;
    e->scope = goperl_scope_v;
}
#define save_hptr(hvp) goperl_hostsave_ptr(_gof, (void **)(hvp))
#define save_aptr(avp) goperl_hostsave_ptr(_gof, (void **)(avp))
#define SAVEVPTR(v) goperl_hostsave_ptr(_gof, (void **)&(v))

/* SAVESPTR/SAVEGENERICSV on PL_diehook/PL_warnhook route to the guest's
 * own save stack (they must be restored by the matching guest LEAVE, in
 * the interpreter). Any other argument is a host location. */
static void goperl_save_sv_slot(goperl_frame_t *f, SV **loc, int generic) {
    if (loc == (SV **)&f->hook_val[0])
        goperl_do_op(f, GOPERL_OP_SAVE_HOOK, GOPERL_PL_DIEHOOK,
                     (uint64_t)generic, 0, 0);
    else if (loc == (SV **)&f->hook_val[1])
        goperl_do_op(f, GOPERL_OP_SAVE_HOOK, GOPERL_PL_WARNHOOK,
                     (uint64_t)generic, 0, 0);
    else
        goperl_hostsave_ptr(f, (void **)loc);
}
#define SAVESPTR(s) goperl_save_sv_slot(_gof, &(s), 0)
#define SAVEGENERICSV(s) goperl_save_sv_slot(_gof, &(s), 1)

#define SAVEOP() ((void)goperl_op0(GOPERL_OP_SAVE_OP, 0, 0))
#define save_op() SAVEOP()
#define SAVEDELETE(hv, key, klen) goperl_save_delete(_gof, (hv), (key), (klen))
static void goperl_save_delete(goperl_frame_t *f, HV *hv, char *key,
                               I32 klen) {
    I32 abslen = klen < 0 ? -klen : klen;
    goperl_do_op(f, GOPERL_OP_SAVE_DELETE, GOPERL_TOK(hv),
                 (uint64_t)(int64_t)klen, key, (uint64_t)abslen);
    free(key); /* SAVEDELETE takes ownership of the savepvn'd key */
}
#define save_helem_flags(hv, keysv, svp, flags)                        \
    ((void)(svp),                                                      \
     (void)goperl_op0(GOPERL_OP_SAVE_HELEM, GOPERL_TOK(hv),            \
                      ((uint64_t)(uint32_t)(flags) << 32) |            \
                          (uint32_t)GOPERL_TOK(keysv)))
#define save_helem(hv, keysv, svp) save_helem_flags((hv), (keysv), (svp), 0)

/* call_sv / call_method: arguments between the last PUSHMARK and the stack
 * top cross as a token list; results come back in a mortal AV and are
 * pushed where perl would leave them. G_EVAL is forced across the boundary
 * (an uncaught guest die cannot unwind host C frames); when the caller did
 * not ask for G_EVAL the error is re-raised here as a croak. */
static I32 goperl_do_call(goperl_frame_t *f, SV *sv, const char *method,
                          I32 flags) {
    I32 markoff = goperl_popmark(f);
    SV **base = (SV **)f->st;
    SV **mark = base + markoff;
    I32 nargs = (I32)(f->psp - mark);
    char buf[GOPERL_XS_STACK * 4 + 256];
    size_t o = 0;
    if (method) {
        size_t ml = strlen(method) + 1;
        if (ml > 255) ml = 255;
        memcpy(buf, method, ml);
        buf[ml - 1] = '\0';
        o = ml;
    }
    for (I32 i = 0; i < nargs; i++) {
        uint32_t tok = (uint32_t)(uint64_t)(uintptr_t)mark[i + 1];
        memcpy(buf + o, &tok, 4);
        o += 4;
    }
    uint64_t packed;
    if (method)
        packed = goperl_do_op(f, GOPERL_OP_CALL_METHOD,
                              ((uint64_t)(uint32_t)flags << 32), o, buf, o);
    else
        packed = goperl_do_op(f, GOPERL_OP_CALL_SV,
                              ((uint64_t)(uint32_t)flags << 32) |
                                  (uint32_t)GOPERL_TOK(sv),
                              o, buf, o);
    int died = (int)(packed >> 63);
    I32 count = (I32)((packed >> 32) & 0x7FFFFFFF);
    uint64_t avtok = (uint32_t)packed;
    f->psp = mark; /* args consumed */
    goperl_stack_extend(f, f->psp, count);
    for (I32 i = 0; i < count; i++) {
        uint64_t tok = goperl_api_v->xs_op(f, GOPERL_OP_AV_FETCH, avtok,
                                           (uint64_t)(int64_t)i, 0, 0);
        *++f->psp = (SV *)(uintptr_t)tok;
    }
    if (died && !(flags & G_EVAL)) {
        uint64_t errtok = goperl_api_v->xs_op(f, GOPERL_OP_ERRSV, 0, 0, 0, 0);
        uint64_t len = 0;
        const char *msg = goperl_api_v->sv_pv(f, errtok, &len);
        goperl_croakf(f, "%.*s", (int)(len > 400 ? 400 : len), msg ? msg : "");
    }
    return count;
}
#define call_sv(sv, flags) goperl_do_call(_gof, (SV *)(sv), 0, (flags))
static I32 Perl_call_sv(pTHX_ SV *sv, I32 flags) __attribute__((unused));
static I32 Perl_call_sv(pTHX_ SV *sv, I32 flags) {
    return goperl_do_call(_gof, sv, 0, flags);
}
#define call_method(name, flags) goperl_do_call(_gof, 0, (name), (flags))

/* do_join: join mark[1]..*sp with delim into sv (doop.c semantics). */
static void goperl_do_join(goperl_frame_t *f, SV *sv, SV *delim, SV **mark,
                           SV **sp) __attribute__((unused));
static void goperl_do_join(goperl_frame_t *f, SV *sv, SV *delim, SV **mark,
                           SV **sp) {
    goperl_do_op(f, GOPERL_OP_SV_SETPVN, GOPERL_TOK(sv), 0, "", 0);
    for (SV **p = mark + 1; p <= sp; p++) {
        if (p != mark + 1)
            goperl_do_op(f, GOPERL_OP_SV_CATSV, GOPERL_TOK(sv),
                         GOPERL_TOK(delim), 0, 0);
        goperl_do_op(f, GOPERL_OP_SV_CATSV, GOPERL_TOK(sv), GOPERL_TOK(*p), 0,
                     0);
    }
}
#define do_join(sv, delim, mark, sp) \
    goperl_do_join(_gof, (sv), (delim), (mark), (sp))

/* ---- PL_ppaddr proxy / scratch-op execution ------------------------------ */

static OP *goperl_pp_trampoline(pTHX) {
    OP *o = _gof->plop;
    if (!o) goperl_croakf(_gof, "pp trampoline: PL_op is not set");
    I32 nargs;
    switch (o->op_type) {
    case OP_FLOP:
        nargs = 2;
        break;
    default:
        goperl_croakf(_gof, "pp trampoline: op %d is not supported by the SDK",
                      (int)o->op_type);
    }
    SV **base = (SV **)_gof->st;
    if ((I32)(_gof->psp - base) < nargs)
        goperl_croakf(_gof, "pp trampoline: stack underflow");
    char buf[64 * 4];
    for (I32 i = 0; i < nargs; i++) {
        uint32_t tok =
            (uint32_t)(uint64_t)(uintptr_t)_gof->psp[i - nargs + 1];
        memcpy(buf + i * 4, &tok, 4);
    }
    uint64_t packed = goperl_do_op(
        _gof, GOPERL_OP_RUN_PP,
        ((uint64_t)(uint32_t)o->op_flags << 32) | (uint32_t)o->op_type,
        (uint64_t)(nargs * 4), buf, (uint64_t)(nargs * 4));
    I32 count = (I32)((packed >> 32) & 0x7FFFFFFF);
    uint64_t avtok = (uint32_t)packed;
    _gof->psp -= nargs;
    goperl_stack_extend(_gof, _gof->psp, count);
    for (I32 i = 0; i < count; i++) {
        uint64_t tok = goperl_api_v->xs_op(_gof, GOPERL_OP_AV_FETCH, avtok,
                                           (uint64_t)(int64_t)i, 0, 0);
        *++_gof->psp = (SV *)(uintptr_t)tok;
    }
    return 0;
}

static Perl_ppaddr_t *goperl_ppaddr_get(goperl_frame_t *f) {
    (void)f;
    if (!goperl_ppaddr_v[0])
        for (int i = 0; i < OP_max; i++) goperl_ppaddr_v[i] = goperl_pp_trampoline;
    return goperl_ppaddr_v;
}
#define PL_ppaddr (goperl_ppaddr_get(_gof))

/* ---- MAGIC --------------------------------------------------------------- */

#define sv_magicext(sv, obj, how, vtbl, name, namlen)                     \
    (goperl_flush(_gof),                                                  \
     goperl_api_v->magic_ext(_gof, GOPERL_TOK(sv), GOPERL_TOK(obj),       \
                             (int32_t)(how), (const void *)(vtbl),        \
                             (const char *)(name), (int64_t)(namlen)))
#define SvMAGIC(sv) \
    (goperl_flush(_gof), goperl_api_v->magic_chain(_gof, GOPERL_TOK(sv)))
static MAGIC *goperl_mg_find(goperl_frame_t *f, const SV *sv, int type) {
    goperl_flush(f);
    for (MAGIC *mg = goperl_api_v->magic_chain(f, GOPERL_TOK(sv)); mg;
         mg = mg->mg_moremagic)
        if (mg->mg_type == (char)type) return mg;
    return 0;
}
#define mg_find(sv, type) goperl_mg_find(_gof, (sv), (type))
#define sv_unmagic(sv, how)                                            \
    (goperl_flush(_gof),                                               \
     goperl_api_v->magic_del(_gof, GOPERL_TOK(sv), (int32_t)(how), 0), 0)
#define sv_unmagicext(sv, how, vtbl)                                   \
    (goperl_flush(_gof),                                               \
     goperl_api_v->magic_del(_gof, GOPERL_TOK(sv), (int32_t)(how),     \
                             (const void *)(vtbl)),                    \
     0)

/* ---- memory helpers ------------------------------------------------------ */

#define Newx(ptr, n, t) ((ptr) = (t *)malloc((size_t)(n) * sizeof(t)))
#define Newxz(ptr, n, t) ((ptr) = (t *)calloc((size_t)(n), sizeof(t)))
#define Renew(ptr, n, t) ((ptr) = (t *)realloc((void *)(ptr), (size_t)(n) * sizeof(t)))
#define Safefree(p) free((void *)(p))
#define Copy(s, d, n, t) ((void)memcpy((d), (s), (size_t)(n) * sizeof(t)))
#define Move(s, d, n, t) ((void)memmove((d), (s), (size_t)(n) * sizeof(t)))
#define Zero(d, n, t) ((void)memset((d), 0, (size_t)(n) * sizeof(t)))
#define StructCopy(s, d, t) (*(d) = *(s))
/* Filesystem shims: XS-level stat/unlink map to the HOST filesystem. With
 * go-perl's default Config the interpreter sees the host filesystem too, so
 * the views agree; a sandboxed Config can diverge from what Perl-level I/O
 * in the same module sees. */
typedef time_t Time_t;
typedef struct stat Stat_t;
#define PerlLIO_stat(path, buf) stat((path), (buf))
#define PerlLIO_lstat(path, buf) lstat((path), (buf))
#define PerlLIO_unlink(path) unlink(path)
#define PerlLIO_open(path, flags) open((path), (flags))

static char *goperl_savepvn(const char *p, size_t n) {
    char *out = (char *)malloc(n + 1);
    if (out) {
        memcpy(out, p, n);
        out[n] = '\0';
    }
    return out;
}
#define savepvn(p, n) goperl_savepvn((p), (size_t)(n))
#define savepv(p) goperl_savepvn((p), strlen(p))
#define memEQ(a, b, n) (memcmp((a), (b), (n)) == 0)
#define strEQ(a, b) (strcmp((a), (b)) == 0)

/* UTF-8 validation, host-side (pure function). */
static int goperl_is_utf8_string(const U8 *s, STRLEN len)
    __attribute__((unused));
static int goperl_is_utf8_string(const U8 *s, STRLEN len) {
    STRLEN i = 0;
    while (i < len) {
        U8 c = s[i];
        STRLEN n;
        if (c < 0x80) n = 1;
        else if ((c & 0xE0) == 0xC0) n = 2;
        else if ((c & 0xF0) == 0xE0) n = 3;
        else if ((c & 0xF8) == 0xF0) n = 4;
        else return 0;
        if (i + n > len) return 0;
        for (STRLEN k = 1; k < n; k++)
            if ((s[i + k] & 0xC0) != 0x80) return 0;
        if (n == 2 && c < 0xC2) return 0; /* overlong */
        i += n;
    }
    return 1;
}
#define is_utf8_string(s, len) goperl_is_utf8_string((s), (len))

/* MY_CXT: the non-threaded model (matching real perl without ithreads):
 * one static struct per translation unit. */
#define START_MY_CXT static my_cxt_t my_cxt_v;
#define dMY_CXT dNOOP
#define MY_CXT my_cxt_v
#define MY_CXT_INIT dNOOP
#define MY_CXT_CLONE dNOOP
#define pMY_CXT_
#define aMY_CXT_
#define pMY_CXT void
#define aMY_CXT

/* Thread-clone stubs: go-perl builds without ithreads, so CLONE hooks are
 * never invoked; these only need to compile (CLONE_PARAMS is defined above
 * the MGVTBL). */
#define CLONEf_KEEP_PTR_TABLE 0
#define sv_dup(sv, param) ((void)(param), (SV *)(sv))
#define sv_dup_inc(sv, param) ((void)(param), SvREFCNT_inc((SV *)(sv)))

#endif /* GOPERL_XS_SDK_PERL_H */
