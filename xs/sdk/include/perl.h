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
#include <math.h>
#include <float.h>
#include <ctype.h>
#include <errno.h>
#include <float.h>
#include <sys/param.h>
#include <sys/stat.h>
#include <sys/times.h>
#include <sys/time.h>
#include <time.h>
#include <unistd.h>
#include <sys/time.h>
#include <pthread.h>

#define GOPERL_XS_ABI 9u

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

/* Capability macros real perl.h inherits from config.h. The SDK targets
 * ordinary darwin/linux hosts, so declare what those always have. */
#ifndef HAS_GETTIMEOFDAY
#define HAS_GETTIMEOFDAY
#endif

/* A dist's bundled ppport.h is a portability layer for REAL perl headers;
 * against this SDK it must not activate. Its include guard is pre-defined
 * so the #include turns into a no-op. */
#define _P_P_PORTABILITY_H_

/* Ancient K&R-compat prototype wrapper old dist headers still use
 * (DBI's DBIXS.h: `void (*fn) _((args));`). ppport used to supply it. */
#ifndef _
#define _(args) args
#endif

typedef int32_t I32;
typedef uint32_t U32;
typedef int16_t I16;
typedef uint16_t U16;
typedef int8_t I8;
typedef uint8_t U8;
typedef int64_t IV;
typedef int64_t Off_t;
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
 * width. Never dereference at runtime. The struct carries one dummy field
 * so legacy code that does SV pointer arithmetic or names sv_flags (the
 * pre-5.18 overload-flag arena walk, see the v5 compat section) still
 * COMPILES; no live path reads it. */
struct goperl_sv { uint32_t sv_flags; };
typedef struct goperl_sv SV;
typedef struct goperl_sv CV;
typedef struct goperl_sv AV;
typedef struct goperl_sv HV;
typedef struct goperl_sv GV;
typedef struct goperl_sv IO;
typedef struct goperl_interp PerlInterpreter;
/* HE/HEK are REAL structs (host shadows): dists both consume HEs from
 * iteration/fetch_ent AND fabricate their own with New + HeSVKEY_set
 * (Cpanel::JSON::XS), then read either kind back through the same macros —
 * so member access must work. A shadow made from a guest HE carries the
 * guest token; its key bytes materialize lazily into an arena HEK laid out
 * like perl's (NUL after the bytes, then the flags byte). */
typedef struct hek {
    U32 hek_hash;
    I32 hek_len;
    char hek_key[1];
} HEK;
#define cBOOL(x) ((bool)!!(x))
#define DO_UTF8(sv) SvUTF8(sv)
typedef struct goperl_he HE;
struct goperl_he {
    HE *hent_next;
    HEK *hent_hek;
    union {
        SV *hent_val;
    } he_valu;
    uint64_t goperl_tok; /* guest HE token; 0 = fabricated by the dist */
};

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
#define OPf_KIDS 4

/* Opcode numbers and names, pinned to the embedded perl (generated). */
#include "goperl_opnames.h"

/* Context types (cop.h, real perl 5.42 values). */
#define CXTYPEMASK 0xf
#define CXt_NULL 0
#define CXt_WHEN 1
#define CXt_BLOCK 2
#define CXt_GIVEN 3
#define CXt_LOOP_ARY 4
#define CXt_LOOP_LAZYSV 5
#define CXt_LOOP_LAZYIV 6
#define CXt_LOOP_LIST 7
#define CXt_LOOP_PLAIN 8
#define CXt_SUB 9
#define CXt_FORMAT 10
#define CXt_EVAL 11
#define CXt_SUBST 12
#define CXt_DEFER 13

/* Debugger flags ($^P) and exit flags (real perl 5.42 values). */
#define PERLDBf_SUB 0x01
#define PERLDBf_LINE 0x02
#define PERLDBf_NOOPT 0x04
#define PERLDBf_INTER 0x08
#define PERLDBf_SUBLINE 0x10
#define PERLDBf_SINGLE 0x20
#define PERLDBf_NONAME 0x40
#define PERLDBf_GOTO 0x80
#define PERLDBf_NAMEEVAL 0x100
#define PERLDBf_NAMEANON 0x200
#define PERLDBf_SAVESRC 0x400
#define PERLDBf_SAVESRC_NOSUBS 0x800
#define PERLDBf_SAVESRC_INVALID 0x1000
#define PERL_EXIT_DESTRUCT_END 0x02
#define GV_ADDWARN 0x04

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
/* The current activation frame lives in the goperl_cur_frame_v global,
 * maintained at every entry/exit point (dXSARGS, xs_leave, the pp-hook and
 * destructor invokers). `_gof` — the name every SDK macro reads — is an
 * alias for it, so interpreter macros work in ANY function, with or without
 * a pTHX parameter (real XS is full of THX-less helpers: Moose's
 * mop_get_code_info and friends). pTHX still declares the ABI parameter the
 * loader passes, but it is vestigial — always equal to the global while the
 * XSUB runs. */
#define pTHX goperl_frame_t *goperl_frame_arg __attribute__((unused))
#define pTHX_ goperl_frame_t *goperl_frame_arg __attribute__((unused)),
#define aTHX goperl_cur_frame_v
#define aTHX_ goperl_cur_frame_v,
#define _gof goperl_cur_frame_v
/* dTHX binds goperl_frame_arg locally, so a following dXSARGS re-parses
 * the CURRENT activation (DBI's dbixst_bounce_method re-enters the
 * caller's frame after restoring its mark) instead of doing XSUB-entry
 * bookkeeping — dXSARGS skips entry work when arg == current. */
#define dTHX \
    goperl_frame_t *const goperl_frame_arg __attribute__((unused)) = \
        goperl_cur_frame_v
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

/* ---- host-side OP model --------------------------------------------------
 *
 * One struct serves three roles:
 *   - a scratch OP built by the module itself (Zero + assign, the pp_flop
 *     "fake op" idiom): gop == 0;
 *   - a SHADOW of a real guest OP (gop == its token), materialized through
 *     the introspection helper ops and deduplicated in a registry so
 *     pointer identity comparisons behave like real OP pointers;
 *   - a COP view of the same shadow (cop_* fields filled when the op is a
 *     nextstate/dbstate), so (COP*)op casts work as in real perl.
 *
 * Scalar fields are filled at shadow creation. Pointer fields (op_next,
 * op_sibparent, op_first, op_redoop) are filled when the shadow is
 * materialized "full"; a thin shadow reached through another shadow's
 * pointer has them NULL, so a deep optree walk degrades (stops early)
 * rather than crashing. */
struct op;
typedef struct op OP;
typedef struct op COP;
typedef struct op UNOP;
typedef struct op LOOP;
typedef struct op BASEOP;
typedef struct op BINOP;
typedef struct op LISTOP;
typedef struct op LOGOP;
typedef struct op SVOP;
typedef struct op PVOP;
typedef struct op METHOP;
typedef struct op PADOP;
typedef struct op UNOP_AUX;
typedef OP *(*Perl_ppaddr_t)(pTHX);
typedef union {
    SV *sv;
    UV uv;
    IV iv;
    char *pv;
    SSize_t ssize;
} UNOP_AUX_item;
/* BASEOP mirrors real perl's role: a dist declaring its own op struct
 * (struct myop { BASEOP OP *op_first; OP *op_other; ... }) gets the same
 * field offsets as the built-in classes, because struct op is exactly
 * BASEOP followed by the class members in real perl's order (op_last and
 * op_other share an offset, as they do in the interpreter). */
#define BASEOP \
    OP *op_next; \
    OP *op_sibparent; \
    Perl_ppaddr_t op_ppaddr; \
    UV op_targ; \
    U16 op_type; \
    U16 op_spare; \
    U8 op_flags; \
    U8 op_private; \
    /* shadow bookkeeping (part of the SDK base) */ \
    U8 gop_full; \
    U8 gop_iscop; \
    U8 op_moresib; /* op_sibparent holds a sibling, not the parent */ \
    uint64_t gop; /* guest OP token (0 = module-built scratch op) */ \
    struct goperl_op_base *gop_base; /* write-back baseline */
struct op {
    BASEOP
    OP *op_first; /* UNOP view */
    union {
        OP *op_last;  /* BINOP/LISTOP view */
        OP *op_other; /* LOGOP view — same offset, as in real perl */
    };
    OP *op_redoop; /* LOOP view */
    U32 cop_line;
    U32 cop_hints;
    char *cop_warnings; /* only ever compared/assigned via pWARN_* */
    const char *cop_file;    /* interned */
    const char *cop_stashpv; /* interned */
    SV *op_sv;     /* SVOP view */
    char *op_pv;   /* PVOP view */
    UNOP_AUX_item *op_aux; /* UNOP_AUX view */
    SV *op_rclass_sv;      /* METHOP view (5.24+) */
    UV op_rclass_targ;
    union {
        SV *op_meth_sv; /* METHOP view */
    } op_u;
};
#define cUNOPx(o) ((UNOP *)(o))
#define cUNOPo cUNOPx(o)
#define cLOOPx(o) ((LOOP *)(o))
#define cCOPx(o) ((COP *)(o))
#define cBINOPx(o) ((BINOP *)(o))
#define cBINOPo cBINOPx(o)
#define cLISTOPx(o) ((LISTOP *)(o))
#define cLISTOPo cLISTOPx(o)
#define cLOGOPx(o) ((LOGOP *)(o))
#define cLOGOPo cLOGOPx(o)
#define cLOGOP cLOGOPx(PL_op)
#define cSVOPx(o) ((SVOP *)(o))
#define cSVOPo cSVOPx(o)
#define cSVOPx_sv(o) (cSVOPx(o)->op_sv)
#define cPVOPx(o) ((PVOP *)(o))
#define cMETHOPx(o) ((METHOP *)(o))
#define cUNOP_AUXx(o) ((UNOP_AUX *)(o))
#define OpSIBLING(o) goperl_op_sibling(_gof, (OP *)(o))
#define OpHAS_SIBLING(o) (OpSIBLING(o) != 0)
#define OpMORESIB_set(o, sib) \
    (((OP *)(o))->op_moresib = 1, ((OP *)(o))->op_sibparent = (OP *)(sib))
#define OpLASTSIB_set(o, parent) \
    (((OP *)(o))->op_moresib = 0, ((OP *)(o))->op_sibparent = (OP *)(parent))
#define OpMAYBESIB_set(o, sib, parent) \
    (((OP *)(o))->op_sibparent = (sib) ? (OP *)(sib) : (OP *)(parent), \
     ((OP *)(o))->op_moresib = (sib) ? 1 : 0)
#define CopLINE(c) ((c)->cop_line)
#define CopFILE(c) ((c)->cop_file)
#define OutCopFILE(c) ((c)->cop_file)
#define CopSTASHPV(c) ((char *)(c)->cop_stashpv)

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
    /* v4: register a host pp function for an op type (the PL_ppaddr proxy
     * diff detected a module writing that slot). */
    void (*pp_hook_set)(goperl_frame_t *, int32_t op_type, void *fn);
    /* v5: replace the head of an SV's mirror MAGIC chain (SvMAGIC_set) —
     * the chain lives host-side, owned by the loader. */
    void (*magic_set_head)(goperl_frame_t *, uint64_t sv, MAGIC *mg);
    /* v6: the loader-owned cross-module shared state block (zeroed,
     * GOPERL_SHARED_RESERVED bytes — must hold goperl_shared_t). */
    void *shared_raw;
    /* v7: register a host PerlIO layer (name + funcs table); returns the
     * funcs id the guest proxy layer carries. */
    uint32_t (*perlio_def)(goperl_frame_t *, const char *name, void *funcs);
    /* map one guest op (by token) to a host pp function for pp-hook
     * dispatch; the guest op's ppaddr is patched separately (OP_SET) */
    void (*perop_pp_set)(goperl_frame_t *, uint64_t optok, void *fn);
} goperl_api_t;
#define GOPERL_SHARED_RESERVED 262144

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
    uint64_t hook_op_tok;    /* +104 pp-hook frames: the op being executed */
    void *prev_frame;        /* +112 frame nesting (goperl_cur_frame_v) */
    void *inst;              /* +120 this INSTANCE's shared-state block */
    uint64_t st[GOPERL_XS_STACK];   /* +128 the perl stack (tokens) */
    int32_t marks[GOPERL_XS_MARKS]; /* mark offsets into st */
    uint64_t tmp[GOPERL_XS_TMPS];   /* slots backing SV**-returning macros */
    char err[512];
};

/* ---- shared module state (weak: one copy across a multi-TU .so) --------- */

__attribute__((weak)) const goperl_api_t *goperl_api_v = 0;

/* ---- the current frame, per THREAD --------------------------------------
 * "The current activation" is per-thread process-wide state: several
 * instances may run XS at once on different goroutine threads (cloned
 * workers), and a THX-less helper in ANY loaded .so must find the frame of
 * the call running on ITS thread. A per-.so TLS variable cannot carry the
 * value across .so boundaries (weak symbols do not coalesce across shared
 * objects) and a single global races across threads, so the loader-owned
 * process block (api->shared_raw) holds a small thread-id keyed table. */
typedef struct goperl_tframe {
    void *tid; /* pthread_t of the owning thread; 0 = free slot */
    goperl_frame_t *f;
} goperl_tframe_t;
#define GOPERL_TFRAME_SLOTS 512
typedef struct goperl_process {
    goperl_tframe_t tf[GOPERL_TFRAME_SLOTS];
} goperl_process_t;
#define GOPERL_PROCESS ((goperl_process_t *)goperl_api_v->shared_raw)

static goperl_frame_t **goperl_curframe_slot(void) {
    void *tid = (void *)pthread_self();
    goperl_tframe_t *tf = GOPERL_PROCESS->tf;
    uint32_t h = (uint32_t)(((uintptr_t)tid >> 4) % GOPERL_TFRAME_SLOTS);
    for (uint32_t i = 0; i < GOPERL_TFRAME_SLOTS; i++) {
        goperl_tframe_t *e = &tf[(h + i) % GOPERL_TFRAME_SLOTS];
        if (e->tid == tid) return &e->f;
        if (e->tid == 0) {
            /* claim: threads only ever insert their OWN slot, so a lost
             * race means someone else claimed a different tid here */
            if (__sync_bool_compare_and_swap(&e->tid, (void *)0, tid))
                return &e->f;
            if (e->tid == tid) return &e->f;
        }
    }
    /* table full: unrecoverable configuration error */
    abort();
}
#define goperl_cur_frame_v (*goperl_curframe_slot())

/* This INSTANCE's shared block, reached through the current frame (the
 * loader stamps every frame with its instance's block). */
#define GOPERL_SHARED ((goperl_shared_t *)goperl_cur_frame_v->inst)

/* Host save-stack backing SAVESPTR/save_hptr on HOST memory locations,
 * unwound by the LEAVE that closes the scope (or by croak unwind). */
typedef struct goperl_hostsave {
    void **loc;
    void *val;
    int32_t scope;
    int32_t width; /* 0 = pointer-width; 4 = a saved I32/int */
} goperl_hostsave_t;
#define GOPERL_HOSTSAVE_MAX 256

/* Control-flow state SHARED ACROSS MODULES. Weak globals only coalesce
 * within one shared object; with several SDK-built .so files loaded
 * (DBI + HTTP::Parser::XS), each had its own current-frame/save-stack and
 * a hook running in one module dereferenced another's stale frame. The
 * loader owns ONE block (api->shared) that every module aliases. */
typedef struct goperl_avmirror {
    uint64_t avtok;
    uint64_t *buf;
    uint64_t *shadow;
    int32_t len;
    int32_t cap;
} goperl_avmirror_t;
typedef struct goperl_ivmirror {
    uint64_t svtok;
    IV val;
    IV shadow;
} goperl_ivmirror_t;
#define GOPERL_OPREG_BUCKETS 1024
typedef struct goperl_shared {
    goperl_frame_t *cur_frame; /* unused since ABI 9 (per-thread table) */
    int32_t xs_depth;
    int32_t scope;
    int32_t hostsave_n;
    int32_t modglobal_seeded;
    int32_t *markstack_ptr;
    goperl_frame_t *markstack_owner;
    uint64_t svcur_sv;
    STRLEN svcur_val;
    STRLEN svcur_orig;
    uint64_t svmagic_sv;
    MAGIC *svmagic_val;
    MAGIC *svmagic_orig;
    char *svpvx_slot;
    STRLEN svlen_slot;
    int32_t avmirror_n;
    int32_t ivmirror_n;
    goperl_hostsave_t hostsave[GOPERL_HOSTSAVE_MAX];
    goperl_avmirror_t avmirrors[64];
    goperl_ivmirror_t ivmirrors[32];
    /* v8 parse surface — one copy for ALL loaded modules (weak globals do
     * not coalesce across .so files; see the v6 shared-state lesson) */
    void *kw_head;    /* host keyword-plugin chain head */
    void *infix_head; /* host infix-plugin chain head */
    void *opreg[GOPERL_OPREG_BUCKETS]; /* op shadow registry buckets */
    void *opext_head;                  /* persistent op-shadow list */
    int32_t opreg_persist_mode;
    uint32_t hints_slot; /* writable PL_hints */
    uint32_t hints_base;
    int32_t hints_live;
    void *parser_ptr;    /* yy_parser* (NULL = guest parser absent) */
    uint32_t parser_pvg; /* guest address of the linestr PV */
    int32_t pad0;
    unsigned char parser_shadow[128]; /* the yy_parser instance */
    /* Per-INSTANCE interpreter state that lived in per-.so weak globals
     * until concurrent instances existed (cloned workers): every one of
     * these is mutated at runtime and must be as private per instance as
     * the interpreter it mirrors. Raw byte areas; the owning code aliases
     * them through its own types (typedefs appear later in this header). */
    unsigned char plshadow[32 * 16];      /* goperl_plshadow_t[32] */
    char form_bufs[8][1024];              /* SVf formatter rotation */
    int32_t form_ix;
    int32_t pad1;
    int64_t amagic_generation;
    void *curcop_slot;                    /* COP* lvalue slot */
    void *intern[256];                    /* string-intern buckets */
    unsigned char dtors[4096 * 24];       /* goperl_dtor_t[4096] */
    int32_t dtor_hint;
    int32_t pad2;
    unsigned char ssnew[4096 * 8];        /* goperl_ssnew_t[4096] */
    int32_t ssnew_hint;
    int32_t pad3;
    void *arena_head;                     /* context-shadow arena */
    void *si_cache;                       /* PERL_SI* */
    uint64_t gen;                         /* bumped per guest op */
    uint64_t si_gen;
} goperl_shared_t;
#define goperl_hostsave_v (GOPERL_SHARED->hostsave)
#define goperl_hostsave_n (GOPERL_SHARED->hostsave_n)
#define goperl_scope_v (GOPERL_SHARED->scope)

/* AV body mirrors backing AvARRAY: per-AV stable buffers of tokens, with a
 * shadow copy for detecting host writes to flush back (refcount-neutral). */
#define GOPERL_AVMIRROR_MAX 64
#define goperl_avmirrors_v (GOPERL_SHARED->avmirrors)
#define goperl_avmirror_n (GOPERL_SHARED->avmirror_n)

/* IV mirrors backing the lvalue SvIVX idiom (`SvIVX(counter)++`): the value
 * lives host-side between guest operations and any change is flushed back
 * as sv_setiv before the next one. */
#define GOPERL_IVMIRROR_MAX 32
#define goperl_ivmirrors_v (GOPERL_SHARED->ivmirrors)
#define goperl_ivmirror_n (GOPERL_SHARED->ivmirror_n)

/* PL_ppaddr proxy: every slot is the same generic trampoline. */
__attribute__((weak)) Perl_ppaddr_t goperl_ppaddr_v[OP_max];

/* Native-XSUB nesting depth: when the top-level XSUB returns, the AV body
 * mirrors are dropped (guest AVs may be freed between calls; a stale
 * mirror must never flush into reused memory). */
#define goperl_xs_depth_v (GOPERL_SHARED->xs_depth)

/* form() rotating buffers. */
#define GOPERL_FORM_BUFS 8
#define GOPERL_FORM_LEN 1024
#define goperl_form_bufs_v (GOPERL_SHARED->form_bufs)
#define goperl_form_ix_v (GOPERL_SHARED->form_ix)
typedef char goperl_form_fits
    [(sizeof(((goperl_shared_t *)0)->form_bufs) ==
      GOPERL_FORM_BUFS * GOPERL_FORM_LEN)
         ? 1
         : -1] __attribute__((unused));

__attribute__((weak, visibility("default"), used)) uint32_t
__goperl_xs_init(const goperl_api_t *api) {
    if (!api || api->abi != GOPERL_XS_ABI) return 0;
    if (!api->shared_raw ||
        sizeof(goperl_shared_t) > GOPERL_SHARED_RESERVED)
        return 0;
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
    GOPERL_OP_AV_STORE_RAW = 97,
    GOPERL_OP_PP_HOOK_SET = 98,
    GOPERL_OP_RUN_ORIGINAL = 99,
    GOPERL_OP_SAVE_DESTRUCTOR = 100,
    GOPERL_OP_OP_FIELDS = 101,
    GOPERL_OP_OP_PTR = 102,
    GOPERL_OP_COP_LINE = 103,
    GOPERL_OP_COP_FILE = 104,
    GOPERL_OP_COP_STASHPV = 105,
    GOPERL_OP_CV_INFO = 106,
    GOPERL_OP_CV_PTR = 107,
    GOPERL_OP_GV_PTR = 108,
    GOPERL_OP_GV_NAME = 109,
    GOPERL_OP_SV_ISGV_GP = 110,
    GOPERL_OP_HV_NAME = 111,
    GOPERL_OP_SI_GET = 112,
    GOPERL_OP_CX_FIELDS = 113,
    GOPERL_OP_CX_PTR = 114,
    GOPERL_OP_OP_NAME_STR = 115,
    GOPERL_OP_CV_FILE = 116,
    GOPERL_OP_GV_FETCHFILE = 117,
    GOPERL_OP_HV_CLEAR = 118,
    GOPERL_OP_HV_DELETE = 119,
    GOPERL_OP_AV_EXISTS = 120,
    GOPERL_OP_SV_UTF8_OFF = 121,
    GOPERL_OP_SAVE_SCALAR = 122,
    GOPERL_OP_SV_READONLY_ON = 123,
    GOPERL_OP_EVAL_PV = 124,
    GOPERL_OP_AV_UNSHIFT = 125,
    GOPERL_OP_NEW_CONSTSUB = 126,
    GOPERL_OP_SV_PVX_RAW = 127,
    GOPERL_OP_HV_PKG_GEN = 128,
    GOPERL_OP_SV_REFCNT = 129,
    GOPERL_OP_GV_INIT = 130,
    GOPERL_OP_GV_AMG = 131,
    GOPERL_OP_SV_AMAGIC_SET = 132,
    GOPERL_OP_SV_MAGIC_SET_HOOK = 133,
    GOPERL_OP_GV_FETCHMETHOD = 134,
    GOPERL_OP_SV_READONLY_OFF = 135,
    GOPERL_OP_SV_PV_FORCE = 136,
    GOPERL_OP_SV_CHOP = 137,
    GOPERL_OP_SV_INSERT = 138,
    GOPERL_OP_SV_UTF8_DECODE = 139,
    GOPERL_OP_SV_UTF8_DOWNGRADE = 140,
    GOPERL_OP_MG_SET = 141,
    GOPERL_OP_GIMME_V = 142,
    GOPERL_OP_CKWARN = 143,
    GOPERL_OP_SV_LEN_BUF = 144,
    GOPERL_OP_HV_STORE_KLEN = 145,
    GOPERL_OP_HV_FETCH_KLEN = 146,
    GOPERL_OP_HV_DELETE_ENT = 147,
    GOPERL_OP_AV_POP = 148,
    GOPERL_OP_COP_HINTS = 149,
    GOPERL_OP_SV_IS_BOOL = 150,
    GOPERL_OP_SV_POK_ONLY = 151,
    GOPERL_OP_EVAL_SV = 152,
    GOPERL_OP_AV_SHIFT = 153,
    GOPERL_OP_SV_2IO = 154,
    GOPERL_OP_SV_FORCE_NORMAL = 155,
    GOPERL_OP_PERLIO_OPEN = 156,
    GOPERL_OP_PERLIO_CLOSE = 157,
    GOPERL_OP_PERLIO_PUTS = 158,
    GOPERL_OP_PERLIO_FLUSH = 159,
    GOPERL_OP_PERLIO_WRITE = 160,
    GOPERL_OP_IO_OFP = 161,
    GOPERL_OP_SV_CUR = 162,
    GOPERL_OP_SV_MAGIC_STD = 163,
    GOPERL_OP_PERLIO_DEF_LAYER = 164,
    GOPERL_OP_PERLIO_NEXT_READ = 165,
    GOPERL_OP_PERLIO_NEXT_FASTGETS = 166,
    GOPERL_OP_PERLIO_NEXT_GETCNT = 167,
    GOPERL_OP_PERLIO_NEXT_GETPTR = 168,
    GOPERL_OP_PERLIO_NEXT_SETPTRCNT = 169,
    GOPERL_OP_PERLIO_NEXT_FILL = 170,
    GOPERL_OP_PERLIO_STATE = 171,
    GOPERL_OP_OPTREE_NEW = 172,
    GOPERL_OP_OPTREE_MISC = 173,
    GOPERL_OP_OP_SET = 174,
    GOPERL_OP_LEX = 175,
    GOPERL_OP_PARSE = 176,
    GOPERL_OP_PARSER_GET = 177,
    GOPERL_OP_PARSER_SET = 178,
    GOPERL_OP_BLOCK = 179,
    GOPERL_OP_PAD = 180,
    GOPERL_OP_KEYWORD_ENABLE = 181,
    GOPERL_OP_CHARCLASS = 182,
    GOPERL_OP_SAVE_MISC = 183,
    GOPERL_OP_SV_CLASSIFY = 184
};

/* PLVAR_GET/SET ids. */
#define GOPERL_PL_DIEHOOK 1
#define GOPERL_PL_WARNHOOK 2
#define GOPERL_PL_SV_UNDEF 3
#define GOPERL_PL_SV_YES 4
#define GOPERL_PL_SV_NO 5
#define GOPERL_PL_CURCOP 6
#define GOPERL_PL_OP 7
#define GOPERL_PL_PERLDB 8
#define GOPERL_PL_DBSUB 9
#define GOPERL_PL_DBSINGLE 10
#define GOPERL_PL_ENDAV 11
#define GOPERL_PL_CHECKAV 12
#define GOPERL_PL_INITAV 13
#define GOPERL_PL_MAIN_CV 14
#define GOPERL_PL_DEBSTASH 15
#define GOPERL_PL_SAWAMPERSAND 16
#define GOPERL_PL_SCOPESTACK_IX 17
#define GOPERL_PL_EXIT_FLAGS 18
#define GOPERL_PL_BASETIME 19
#define GOPERL_PL_MODGLOBAL 20
#define GOPERL_PL_MINUS_C 21
#define GOPERL_PL_CURSTASH 22
#define GOPERL_PL_DOWARN 23
#define GOPERL_PL_HINTS 24
#define GOPERL_PL_SUB_GENERATION 25
#define GOPERL_PL_DEFGV 26
#define GOPERL_PL_HINTGV 27
#define GOPERL_PL_COMPCV 28

#define GOPERL_TOK(sv) ((uint64_t)(uintptr_t)(sv))
#define GOPERL_SV(tok) ((SV *)(uintptr_t)(tok))

/* ---- the guest-op funnel: every operation flushes host state first ------ */

static void goperl_flush(goperl_frame_t *f);
static void goperl_hostsave_ptr(goperl_frame_t *f, void **loc);
static void goperl_hostsave_i32(goperl_frame_t *f, int32_t *loc);
static void goperl_svcur_writeback(goperl_frame_t *f);
static void goperl_gen_bump(void);

static uint64_t goperl_do_op(goperl_frame_t *f, int32_t op, uint64_t a,
                             uint64_t b, const char *s, uint64_t slen) {
    goperl_flush(f);
    uint64_t r = goperl_api_v->xs_op(f, op, a, b, s, slen);
    /* any guest operation may have changed interpreter state the context
     * mirrors reflect */
    goperl_gen_bump();
    return r;
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
        if (e->width == 4)
            *(int32_t *)e->loc = (int32_t)(intptr_t)e->val;
        else
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
#define UTF8f "\002GU8\002"
#define UTF8fARG(u, l, p) ((int)(u)), ((UV)(l)), ((const char *)(p))
#define IVdf "lld"
#define UVuf "llu"
#define UVxf "llx"
#define UVXf "llX"
#define NVef "e"
#define NVff "f"
#define NVgf "g"
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
        if (strncmp(fmt, UTF8f, sizeof(UTF8f) - 1) == 0) {
            (void)va_arg(ap, int); /* is-utf8 flag: bytes append either way */
            UV ulen = va_arg(ap, UV);
            const char *up = va_arg(ap, const char *);
            for (UV i = 0; up && i < ulen; i++) GOPERL_PUTC(up[i]);
            fmt += sizeof(UTF8f) - 1;
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
static const char *Perl_form(pTHX_ const char *fmt, ...)
    __attribute__((unused));
static const char *Perl_form(pTHX_ const char *fmt, ...) {
    char *buf = goperl_form_bufs_v[goperl_form_ix_v++ % GOPERL_FORM_BUFS];
    va_list ap;
    va_start(ap, fmt);
    goperl_vfmt(_gof, buf, GOPERL_FORM_LEN, fmt, ap);
    va_end(ap);
    return buf;
}
static const char *Perl_form_nocontext(const char *fmt, ...)
    __attribute__((unused));
static const char *Perl_form_nocontext(const char *fmt, ...) {
    char *buf = goperl_form_bufs_v[goperl_form_ix_v++ % GOPERL_FORM_BUFS];
    va_list ap;
    va_start(ap, fmt);
    goperl_vfmt(goperl_cur_frame_v, buf, GOPERL_FORM_LEN, fmt, ap);
    va_end(ap);
    return buf;
}

/* croak() and warn() (no context arg) are variadic macros over the local
 * frame. Perl_croak/Perl_warn/Perl_warner are called with aTHX_ — which now
 * expands to a real argument — so they must be actual variadic FUNCTIONS
 * taking pTHX_ (a macro would glue `aTHX_ fmt` into one argument). */
#define croak(...) goperl_croakf(goperl_cur_frame_v, __VA_ARGS__)
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
#define warn(...) goperl_warnf(goperl_cur_frame_v, __VA_ARGS__)
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
#define SvPV_nomg(sv, lenvar) SvPV((sv), lenvar)
#define SvPV_nomg_const(sv, lenvar) SvPV_const((sv), lenvar)
#define SvPV_nomg_nolen(sv) SvPV_nolen(sv)
/* SvPVX is the RAW buffer pointer (no get-magic, no stringification):
 * code grows the SV and writes through it. NULL when the SV has no
 * buffer yet — grow first, as real perl requires. */
static char *goperl_sv_pvx(goperl_frame_t *f, SV *sv) {
    uint64_t gp =
        goperl_do_op(f, GOPERL_OP_SV_PVX_RAW, GOPERL_TOK(sv), 0, 0, 0);
    if (!gp) return 0;
    return (char *)goperl_api_v->guest_mem(f, gp);
}
/* SvPVX reads the real buffer pointer; it is also ASSIGNABLE because old
 * XS (DBI) stashes a file pointer into a CV's PVX slot as write-only
 * bookkeeping — such slot assignments land in a scratch and are
 * discarded (modern newXS already records the file). */
#define goperl_svpvx_slot_v (GOPERL_SHARED->svpvx_slot)
static char **goperl_svpvx_lv(goperl_frame_t *f, SV *sv) {
    goperl_svpvx_slot_v = goperl_sv_pvx(f, sv);
    return &goperl_svpvx_slot_v;
}
#define SvPVX(sv) (*goperl_svpvx_lv(_gof, (SV *)(sv)))
#define SvPVX_const(sv) ((const char *)goperl_sv_pvx(_gof, (SV *)(sv)))
/* SvCUR is WRITABLE (DBI trims strings with --SvCUR(sv)): reads load a
 * synced slot, and goperl_flush writes a changed slot back through
 * SV_CUR_SET before the next guest operation. */
#define goperl_svcur_sv_v (GOPERL_SHARED->svcur_sv)
#define goperl_svcur_val_v (GOPERL_SHARED->svcur_val)
#define goperl_svcur_orig_v (GOPERL_SHARED->svcur_orig)
static void goperl_svcur_writeback(goperl_frame_t *f) {
    if (goperl_svcur_sv_v && goperl_svcur_val_v != goperl_svcur_orig_v) {
        uint64_t sv = goperl_svcur_sv_v;
        goperl_svcur_sv_v = 0; /* break the flush recursion */
        goperl_api_v->xs_op(f, GOPERL_OP_SV_CUR_SET, sv,
                            (uint64_t)goperl_svcur_val_v, 0, 0);
    }
    goperl_svcur_sv_v = 0;
}
static STRLEN *goperl_svcur_lv(goperl_frame_t *f, SV *sv) {
    if (goperl_svcur_sv_v != GOPERL_TOK(sv)) {
        goperl_svcur_writeback(f);
        goperl_svcur_val_v = (STRLEN)goperl_do_op(f, GOPERL_OP_SV_CUR,
                                                  GOPERL_TOK(sv), 0, 0, 0);
        goperl_svcur_orig_v = goperl_svcur_val_v;
        goperl_svcur_sv_v = GOPERL_TOK(sv);
    }
    return &goperl_svcur_val_v;
}
#define SvCUR(sv) (*goperl_svcur_lv(_gof, (SV *)(sv)))
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
static void goperl_sv_setpvf(goperl_frame_t *f, SV *sv, const char *fmt, ...)
    __attribute__((unused));
static void goperl_sv_setpvf(goperl_frame_t *f, SV *sv, const char *fmt,
                             ...) {
    char buf[2048];
    va_list ap;
    va_start(ap, fmt);
    goperl_vfmt(f, buf, sizeof buf, fmt, ap);
    va_end(ap);
    goperl_do_op(f, GOPERL_OP_SV_SETPVN, GOPERL_TOK(sv), strlen(buf), buf,
                 strlen(buf));
}
#define sv_setpvf(sv, ...) goperl_sv_setpvf(_gof, (SV *)(sv), __VA_ARGS__)
static void goperl_sv_catpvf(goperl_frame_t *f, SV *sv, const char *fmt, ...)
    __attribute__((unused));
static void goperl_sv_catpvf(goperl_frame_t *f, SV *sv, const char *fmt,
                             ...) {
    char buf[2048];
    va_list ap;
    va_start(ap, fmt);
    goperl_vfmt(f, buf, sizeof buf, fmt, ap);
    va_end(ap);
    goperl_do_op(f, GOPERL_OP_SV_CATPVN, GOPERL_TOK(sv), strlen(buf), buf,
                 strlen(buf));
}
#define sv_catpvf(sv, ...) goperl_sv_catpvf(_gof, (SV *)(sv), __VA_ARGS__)
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
    int utf8 = klen < 0;
    if (utf8) klen = -klen;
    uint64_t tok = goperl_do_op(f, GOPERL_OP_HV_FETCH_KLEN, GOPERL_TOK(hv),
                                ((uint64_t)(utf8 ? 1 : 0) << 32) |
                                    (uint32_t)(lval ? 1 : 0),
                                key, (uint64_t)(uint32_t)klen);
    if (!tok && lval) {
        /* An lvalue fetch only fails on a hash being torn down (core's
         * hv_common bails on SvIS_FREED). XS deref-s the result without
         * checking (DBI's set_err during global destruction), so hand
         * back a mortal undef slot: reads see undef, writes evaporate
         * with the dying hash — which is where they were headed anyway. */
        uint64_t u = goperl_do_op(f, GOPERL_OP_NEW_SV, 0, 0, 0, 0);
        goperl_do_op(f, GOPERL_OP_SV_MORTAL, u, 0, 0, 0);
        return goperl_tmp_slot(f, u);
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
/* hv_store returns the stored slot (fetched back after the store) so
 * callers that check it for success — real perl returns NULL only on
 * failure — behave correctly (Moose's collect_all_symbols does). */
static SV **goperl_hv_store(goperl_frame_t *f, HV *hv, const char *key,
                            I32 klen, SV *sv) {
    /* Negative klen = UTF-8 key (real hv_store convention). */
    int utf8 = klen < 0;
    if (utf8) klen = -klen;
    uint64_t tok = goperl_do_op(f, GOPERL_OP_HV_STORE_KLEN, GOPERL_TOK(hv),
                                ((uint64_t)(utf8 ? 1 : 0) << 32) |
                                    (uint32_t)GOPERL_TOK(sv),
                                key, (uint64_t)(uint32_t)klen);
    if (!tok) return 0;
    return goperl_tmp_slot(f, tok);
}
#define hv_store(hv, key, klen, sv, hash) \
    goperl_hv_store(_gof, (HV *)(hv), (key), (I32)(klen), (SV *)(sv))
#define hv_stores(hv, key, sv) hv_store((hv), "" key "", sizeof(key) - 1, (sv), 0)
#define hv_exists_ent(hv, keysv, hash)                            \
    ((I32)goperl_op0(GOPERL_OP_HV_EXISTS_ENT, GOPERL_TOK(hv),     \
                     GOPERL_TOK(keysv)))
#define hv_iterinit(hv) \
    ((I32)goperl_op0(GOPERL_OP_HV_ITERINIT, GOPERL_TOK(hv), 0))
#define HvKEYS(hv) hv_iterinit(hv)
/* hv_fetch_ent / hv_store_ent / hv_iternext / HeVAL / hv_iterkeysv /
 * hv_iterval live in the HE-shadow section further down (they need the
 * arena allocator). */
#define HvNAME(hv)                                     \
    ((char *)goperl_intern_packed(                     \
        _gof, goperl_op0(GOPERL_OP_HV_NAME, GOPERL_TOK(hv), 0)))
#define HvNAME_get(hv) HvNAME(hv)

/* ---- stashes / globals / blessing --------------------------------------- */

#define gv_stashpv(name, flags)                                          \
    ((HV *)GOPERL_SV(goperl_ops(GOPERL_OP_GV_STASHPV,                    \
                                (uint64_t)(int64_t)(flags), 0, (name),   \
                                strlen(name))))
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
#define goperl_amagic_generation_v (*(IV *)&GOPERL_SHARED->amagic_generation)
#define PL_amagic_generation goperl_amagic_generation_v

#define PL_op (_gof->plop)

/* Wider interpreter-variable surface (shadow slots, see goperl_plint_ref).
 * Reads are always fresh; assignments flush before the next guest op. */
#define PL_perldb (*goperl_plint_ref(_gof, GOPERL_PL_PERLDB))
#define PL_DBsub (*(GV **)goperl_plsv_ref(_gof, GOPERL_PL_DBSUB))
#define PL_DBsingle (*goperl_plsv_ref(_gof, GOPERL_PL_DBSINGLE))
#define PL_endav (*(AV **)goperl_plsv_ref(_gof, GOPERL_PL_ENDAV))
#define PL_checkav (*(AV **)goperl_plsv_ref(_gof, GOPERL_PL_CHECKAV))
#define PL_initav (*(AV **)goperl_plsv_ref(_gof, GOPERL_PL_INITAV))
#define PL_main_cv (*(CV **)goperl_plsv_ref(_gof, GOPERL_PL_MAIN_CV))
#define PL_debstash (*(HV **)goperl_plsv_ref(_gof, GOPERL_PL_DEBSTASH))
#define PL_sawampersand ((U8) * goperl_plint_ref(_gof, GOPERL_PL_SAWAMPERSAND))
#define PL_scopestack_ix ((I32) * goperl_plint_ref(_gof, GOPERL_PL_SCOPESTACK_IX))
#define PL_exit_flags (*goperl_plint_ref(_gof, GOPERL_PL_EXIT_FLAGS))
#define PL_basetime ((Time_t) * goperl_plint_ref(_gof, GOPERL_PL_BASETIME))
/* PL_modglobal carries Time::HiRes's exported C hooks (Time::NVtime /
 * Time::U2time), which consumers call as raw function pointers — a guest
 * pointer is uncallable from the host, so the first host access re-seeds
 * those two entries with HOST implementations of the same wallclock (the
 * pointers round-trip through the SDK's PTR2IV/INT2PTR registry). */
static NV goperl_nvtime_fn(void) {
    struct timeval tv;
    gettimeofday(&tv, 0);
    return (NV)tv.tv_sec + (NV)tv.tv_usec / 1e6;
}
static void goperl_u2time_fn(pTHX_ UV *r) {
    struct timeval tv;
    gettimeofday(&tv, 0);
    r[0] = (UV)tv.tv_sec;
    r[1] = (UV)tv.tv_usec;
}
static SV **goperl_plsv_ref(goperl_frame_t *f, int id);
static HV **goperl_modglobal_ref(goperl_frame_t *f) {
    HV **ref = (HV **)goperl_plsv_ref(f, GOPERL_PL_MODGLOBAL);
    if (!GOPERL_SHARED->modglobal_seeded && *ref) {
        GOPERL_SHARED->modglobal_seeded = 1;
        uint64_t iv;
        iv = goperl_do_op(
            f, GOPERL_OP_NEW_IV,
            goperl_api_v->ptr_encode(f, (void *)goperl_nvtime_fn), 0, 0, 0);
        goperl_do_op(f, GOPERL_OP_HV_STORE_KLEN, GOPERL_TOK(*ref),
                     (uint32_t)iv, "Time::NVtime",
                     sizeof("Time::NVtime") - 1);
        iv = goperl_do_op(
            f, GOPERL_OP_NEW_IV,
            goperl_api_v->ptr_encode(f, (void *)goperl_u2time_fn), 0, 0, 0);
        goperl_do_op(f, GOPERL_OP_HV_STORE_KLEN, GOPERL_TOK(*ref),
                     (uint32_t)iv, "Time::U2time",
                     sizeof("Time::U2time") - 1);
    }
    return ref;
}
#define PL_modglobal (*goperl_modglobal_ref(_gof))
#define PL_minus_c ((int)*goperl_plint_ref(_gof, GOPERL_PL_MINUS_C))
#define PL_curstash (*(HV **)goperl_plsv_ref(_gof, GOPERL_PL_CURSTASH))

/* The current COP, as a host shadow (cached per guest COP token). */
#define PL_curcop (*goperl_curcop_ref(_gof))

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

/* Interpreter-variable shadows beyond the die/warn hooks: read-through on
 * every access, written back before the next guest operation. Slots are
 * indexed by the GOPERL_PL_* id. */
#define GOPERL_PLSHADOW_MAX 32
typedef struct {
    uint64_t val;
    U8 dirty;
} goperl_plshadow_t;
#define goperl_plshadow_v ((goperl_plshadow_t *)GOPERL_SHARED->plshadow)
typedef char goperl_plshadow_fits
    [(sizeof(goperl_plshadow_t) * GOPERL_PLSHADOW_MAX <=
      sizeof(((goperl_shared_t *)0)->plshadow))
         ? 1
         : -1] __attribute__((unused));

static void goperl_plshadow_flush(goperl_frame_t *f) {
    for (int i = 0; i < GOPERL_PLSHADOW_MAX; i++)
        if (goperl_plshadow_v[i].dirty) {
            goperl_plshadow_v[i].dirty = 0;
            goperl_api_v->xs_op(f, GOPERL_OP_PLVAR_SET, (uint64_t)i,
                                goperl_plshadow_v[i].val, 0, 0);
        }
}

static void goperl_ppaddr_sync(goperl_frame_t *f);

#define goperl_hints_slot_v (GOPERL_SHARED->hints_slot)
#define goperl_hints_base_v (GOPERL_SHARED->hints_base)
#define goperl_hints_live_v (GOPERL_SHARED->hints_live)

static void goperl_flush(goperl_frame_t *f) {
    goperl_svcur_writeback(f);
    if (goperl_hints_live_v && goperl_hints_slot_v != goperl_hints_base_v) {
        goperl_hints_base_v = goperl_hints_slot_v;
        goperl_api_v->xs_op(f, GOPERL_OP_PLVAR_SET, (uint64_t)GOPERL_PL_HINTS,
                            (uint64_t)goperl_hints_slot_v, 0, 0);
    }
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
    goperl_plshadow_flush(f);
    goperl_ivmirror_flush(f);
    goperl_avmirror_flush(f);
    goperl_ppaddr_sync(f);
}

/* Take a writable reference to an interpreter variable: refresh from the
 * guest, mark pending so the (possibly modified) value flushes back. */
static IV *goperl_plint_ref(goperl_frame_t *f, int id) {
    goperl_plshadow_t *s = &goperl_plshadow_v[id];
    if (!s->dirty)
        s->val = goperl_api_v->xs_op(f, GOPERL_OP_PLVAR_GET, (uint64_t)id, 0,
                                     0, 0);
    s->dirty = 1;
    return (IV *)&s->val;
}
static SV **goperl_plsv_ref(goperl_frame_t *f, int id) {
    return (SV **)goperl_plint_ref(f, id);
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

/* PL_markstack_ptr is an assignable pointer some dists arithmetic on
 * directly (DBI bumps it to un-pop the mark its caller's dXSARGS took).
 * The pointer lives in a synced global; every mark operation adopts any
 * adjustment back into the frame's mark index first. */
#define goperl_markstack_ptr_v (GOPERL_SHARED->markstack_ptr)
#define goperl_markstack_owner_v (GOPERL_SHARED->markstack_owner)
static void goperl_marks_adopt(goperl_frame_t *f) {
    if (goperl_markstack_owner_v == f && goperl_markstack_ptr_v)
        f->markidx = (int32_t)(goperl_markstack_ptr_v - f->marks);
}
static void goperl_marks_publish(goperl_frame_t *f) {
    goperl_markstack_owner_v = f;
    goperl_markstack_ptr_v = f->marks + f->markidx;
}
static int32_t **goperl_markstack_ref(goperl_frame_t *f) {
    goperl_marks_publish(f);
    return &goperl_markstack_ptr_v;
}
#define PL_markstack_ptr (*goperl_markstack_ref(_gof))
static I32 goperl_popmark(goperl_frame_t *f) {
    goperl_marks_adopt(f);
    I32 r = f->markidx >= 0 ? f->marks[f->markidx--] : 0;
    goperl_marks_publish(f);
    return r;
}
static void goperl_pushmark(goperl_frame_t *f, I32 off) {
    goperl_marks_adopt(f);
    if (f->markidx + 1 >= GOPERL_XS_MARKS)
        goperl_croakf(f, "mark stack overflow in native XS");
    f->marks[++f->markidx] = off;
    goperl_marks_publish(f);
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
/* PerlIO is a MACRO so <perliol.h> (the layer-author view) can replace it
 * with the layered handle type; plain dists keep the FILE* model. */
#define PerlIO FILE
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

static void goperl_arena_free_all(void);
static void goperl_opreg_clear(void);

static void goperl_xs_leave(goperl_frame_t *f) {
    /* capture the frame to make current BEFORE the host-save unwind: the
     * entry protocol parked f->prev_frame's PRE-entry value there, and the
     * unwind below writes it back (for a fresh frame that value is 0) */
    goperl_frame_t *leave_prev = (goperl_frame_t *)f->prev_frame;
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
        goperl_arena_free_all(); /* context-stack shadows */
        goperl_opreg_clear();    /* guest op addresses recycle across evals */
        /* a croak path may leave the scope counter high; at rest it is 0 */
        goperl_scope_v = 0;
    }
    /* restore LAST: the shared-state macros above resolve through the
     * current frame, and the caller's frame can be NULL (thread at rest) */
    goperl_cur_frame_v = leave_prev;
}

/* One protocol for every activation, INCLUDING an XSUB invoked directly
 * by C pointer from another XSUB on the SAME frame (DBI's dispatch calls
 * (*CvXSUB(cv))(aTHX_ cv)): the previous prev_frame/jb are parked on the
 * host save-stack (above this activation's base, so this activation's
 * xs_leave restores them), prev_frame points at the frame the leave must
 * make current again — the caller for a fresh entry, the frame ITSELF for
 * a same-frame nested call — and depth++ balances the leave's decrement
 * either way. */
#define dXSARGS                                            \
    jmp_buf _gof_jb;                                       \
    {                                                      \
        /* current-frame FIRST: the shared-state macros below resolve  \
         * through it, and on a fresh thread the slot starts NULL */   \
        goperl_frame_t *goperl_prev_cur = goperl_cur_frame_v;          \
        goperl_cur_frame_v = goperl_frame_arg;             \
        int32_t goperl_entry_base = goperl_hostsave_n;     \
        goperl_hostsave_i32(goperl_frame_arg,              \
                            &goperl_frame_arg->hostsave_base); \
        goperl_hostsave_ptr(goperl_frame_arg,              \
                            (void **)&goperl_frame_arg->prev_frame); \
        goperl_hostsave_ptr(goperl_frame_arg,              \
                            (void **)&goperl_frame_arg->jb); \
        goperl_frame_arg->prev_frame =                     \
            (goperl_frame_arg == goperl_prev_cur)          \
                ? (void *)goperl_frame_arg                 \
                : (void *)goperl_prev_cur;                 \
        goperl_xs_depth_v++;                               \
        goperl_frame_arg->hostsave_base = goperl_entry_base; \
        goperl_frame_arg->jb = (void *)&_gof_jb;           \
        if (setjmp(_gof_jb)) {                             \
            goperl_hostsave_unwind_to(                     \
                goperl_frame_arg, goperl_frame_arg->hostsave_base); \
            goperl_xs_leave(goperl_frame_arg);             \
            /* Real XSUBs are void; the sole non-void dXSARGS user \
             * (DBI's bounce) is reached nested, where a croak      \
             * unwinds to the OUTER activation's setjmp instead. */ \
            _Pragma("GCC diagnostic push")                 \
            _Pragma("GCC diagnostic ignored \"-Wreturn-mismatch\"") \
            return;                                        \
            _Pragma("GCC diagnostic pop")                  \
        }                                                  \
    }                                                      \
    dSP;                                                   \
    dAXMARK;                                               \
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

/* CV/GV introspection (real guest lookups; names are interned host copies
 * with stable addresses, matching perl's stable-storage guarantees). */
#define CvISXSUB(cv) \
    ((int)(goperl_op0(GOPERL_OP_CV_INFO, GOPERL_TOK(cv), 0) & 1))
#define CvDEPTH(cv) \
    ((I32)(goperl_op0(GOPERL_OP_CV_INFO, GOPERL_TOK(cv), 0) >> 32))
#define CvSTART(cv) \
    goperl_op_shadow_full(_gof, goperl_op0(GOPERL_OP_CV_PTR, GOPERL_TOK(cv), 0))
#define CvROOT(cv) \
    goperl_op_shadow_full(_gof, goperl_op0(GOPERL_OP_CV_PTR, GOPERL_TOK(cv), 3))
#define CvGV(cv) \
    ((GV *)GOPERL_SV(goperl_op0(GOPERL_OP_CV_PTR, GOPERL_TOK(cv), 1)))
#define CvSTASH(cv) \
    ((HV *)GOPERL_SV(goperl_op0(GOPERL_OP_CV_PTR, GOPERL_TOK(cv), 2)))
#define CvFILE(cv)                                     \
    ((char *)goperl_intern_packed(                     \
        _gof, goperl_op0(GOPERL_OP_CV_FILE, GOPERL_TOK(cv), 0)))
#define GvSTASH(gv) \
    ((HV *)GOPERL_SV(goperl_op0(GOPERL_OP_GV_PTR, GOPERL_TOK(gv), 0)))
#define GvCVu(gv) \
    ((CV *)GOPERL_SV(goperl_op0(GOPERL_OP_GV_PTR, GOPERL_TOK(gv), 1)))
#define GvCV(gv) GvCVu(gv)
#define GvHV(gv) \
    ((HV *)GOPERL_SV(goperl_op0(GOPERL_OP_GV_PTR, GOPERL_TOK(gv), 2)))
#define GvAV(gv) \
    ((AV *)GOPERL_SV(goperl_op0(GOPERL_OP_GV_PTR, GOPERL_TOK(gv), 3)))
#define GvSV(gv) \
    GOPERL_SV(goperl_op0(GOPERL_OP_GV_PTR, GOPERL_TOK(gv), 5))
#define GvIO(gv) \
    GOPERL_SV(goperl_op0(GOPERL_OP_GV_PTR, GOPERL_TOK(gv), 6))
#define GvNAME(gv)                                     \
    ((char *)goperl_intern_packed(                     \
        _gof, goperl_op0(GOPERL_OP_GV_NAME, GOPERL_TOK(gv), 0)))
#define isGV_with_GP(sv) \
    (goperl_op0(GOPERL_OP_SV_ISGV_GP, GOPERL_TOK(sv), 0) != 0)

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
 * macro argument, so this must swallow whatever it gets. It is the boot
 * XSUB's RETURN path (modern xsubpp emits it instead of XSRETURN_YES), so
 * it must do what XSRETURN does — set the return value and, crucially,
 * run the frame-exit bookkeeping; skipping that leaks the native nesting
 * depth and the per-activation caches never reset. */
#define Perl_xs_boot_epilog(...)                     \
    STMT_START {                                     \
        ST(0) = &PL_sv_yes;                          \
        PL_stack_sp = PL_stack_base + ax;            \
        goperl_xs_leave(_gof);                       \
    }                                                \
    STMT_END
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
    e->width = 0;
}
static void goperl_hostsave_i32(goperl_frame_t *f, int32_t *loc) {
    if (goperl_hostsave_n >= GOPERL_HOSTSAVE_MAX)
        goperl_croakf(f, "host save stack overflow in native XS");
    goperl_hostsave_t *e = &goperl_hostsave_v[goperl_hostsave_n++];
    e->loc = (void **)loc;
    e->val = (void *)(intptr_t)*loc;
    e->scope = goperl_scope_v;
    e->width = 4;
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
static I32 goperl_call_pv(goperl_frame_t *f, const char *name, I32 flags) {
    uint64_t cv =
        goperl_do_op(f, GOPERL_OP_GET_CV, GV_ADD, 0, name, strlen(name));
    return goperl_do_call(f, (SV *)(uintptr_t)cv, 0, flags);
}
#define call_pv(name, flags) goperl_call_pv(_gof, (name), (flags))
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

/* ---- OP shadow registry --------------------------------------------------
 * Shadows of real guest OPs, deduplicated by token so host pointer
 * identity matches guest OP identity. Registry entries live for the
 * process (guest optrees are effectively stable while a module profiles
 * them; freed-and-reused guest OPs would alias, which is documented). */

typedef struct goperl_opreg_ent {
    struct goperl_opreg_ent *next;
    U8 persist; /* survives opreg_clear (module-owned extended op) */
    OP op;      /* last: an adopted entry extends past it (NewOpSz) */
} goperl_opreg_ent_t;
#define goperl_opreg_persist_mode_v (GOPERL_SHARED->opreg_persist_mode)
#define goperl_opreg_v ((goperl_opreg_ent_t **)GOPERL_SHARED->opreg)

/* The registry deduplicates within one top-level native activation (so OP
 * pointer identity holds where code compares them) but is dropped when the
 * outermost frame returns: guest optrees are freed and their addresses
 * reused across evals, so cross-activation caching would alias stale
 * shadows onto new ops. */
static void goperl_opreg_clear(void) {
    for (int i = 0; i < GOPERL_OPREG_BUCKETS; i++) {
        goperl_opreg_ent_t *e = goperl_opreg_v[i];
        while (e) {
            goperl_opreg_ent_t *n = e->next;
            if (!e->persist) {
                free(e->op.gop_base);
                free(e);
            } else {
                /* module structures still point at this shadow (custom op
                 * state read at runtime); keep the memory, drop the link */
                e->next = 0;
            }
            e = n;
        }
        goperl_opreg_v[i] = 0;
    }
}

static void goperl_op_base_capture(OP *o);

/* Long-lived op shadows: entries whose guest op carries the per-op host
 * pp hook survive registry clears (the module reads their state when the
 * compiled code runs). A hit is validated against the guest (the op still
 * dispatches to the host) so a recycled op address never aliases. */
#define GOPERL_OPPTR_IS_HOOKED 7
typedef struct goperl_opext_ent {
    struct goperl_opext_ent *next;
    uint64_t tok;
    OP *op;
} goperl_opext_ent_t;
#define goperl_opext_v (*(goperl_opext_ent_t **)&GOPERL_SHARED->opext_head)

static void goperl_opext_put(uint64_t tok, OP *o) {
    for (goperl_opext_ent_t *e = goperl_opext_v; e; e = e->next)
        if (e->tok == tok) {
            e->op = o;
            return;
        }
    goperl_opext_ent_t *e =
        (goperl_opext_ent_t *)calloc(1, sizeof(goperl_opext_ent_t));
    if (!e) return;
    e->tok = tok;
    e->op = o;
    e->next = goperl_opext_v;
    goperl_opext_v = e;
}

static OP *goperl_op_shadow(goperl_frame_t *f, uint64_t tok) {
    if (!tok) return 0;
    uint32_t h = (uint32_t)(tok >> 3) % GOPERL_OPREG_BUCKETS;
    for (goperl_opreg_ent_t *e = goperl_opreg_v[h]; e; e = e->next)
        if (e->op.gop == tok) return &e->op;
    for (goperl_opext_ent_t **p = &goperl_opext_v; *p;) {
        goperl_opext_ent_t *e = *p;
        if (e->tok == tok) {
            if (goperl_do_op(f, GOPERL_OP_OP_PTR, tok,
                             GOPERL_OPPTR_IS_HOOKED, 0, 0)) {
                /* relink into the bucket so identity holds this activation */
                goperl_opreg_ent_t *re =
                    (goperl_opreg_ent_t *)((char *)e->op -
                                           offsetof(goperl_opreg_ent_t, op));
                re->next = goperl_opreg_v[h];
                goperl_opreg_v[h] = re;
                return e->op;
            }
            /* the guest op is gone (freed / address recycled): drop */
            *p = e->next;
            free(e);
            continue;
        }
        p = &e->next;
    }
    goperl_opreg_ent_t *e =
        (goperl_opreg_ent_t *)calloc(1, sizeof(goperl_opreg_ent_t));
    if (!e) goperl_croakf(f, "op shadow: out of memory");
    e->persist = (U8)goperl_opreg_persist_mode_v;
    e->op.gop = tok;
    uint64_t fields =
        goperl_do_op(f, GOPERL_OP_OP_FIELDS, tok, 0, 0, 0);
    e->op.op_type = (U16)(fields & 0xFFFF);
    e->op.op_flags = (U8)(fields >> 16);
    e->op.op_private = (U8)(fields >> 24);
    e->op.op_targ = (UV)(uint32_t)(fields >> 32);
    e->next = goperl_opreg_v[h];
    goperl_opreg_v[h] = e;
    return &e->op;
}

/* String interning: stable host copies of guest identity strings (file
 * names, package names). Stability matters: modules compare and retain
 * these pointers (real perl hands out pointers into stable structures). */
#define GOPERL_INTERN_BUCKETS 256
typedef struct goperl_intern_ent {
    struct goperl_intern_ent *next;
    char s[1];
} goperl_intern_ent_t;
#define goperl_intern_v ((goperl_intern_ent_t **)GOPERL_SHARED->intern)
typedef char goperl_intern_fits
    [(sizeof(void *) * GOPERL_INTERN_BUCKETS <=
      sizeof(((goperl_shared_t *)0)->intern))
         ? 1
         : -1] __attribute__((unused));

static const char *goperl_intern(const char *p, size_t n) {
    uint32_t h = 2166136261u;
    for (size_t i = 0; i < n; i++) h = (h ^ (unsigned char)p[i]) * 16777619u;
    h %= GOPERL_INTERN_BUCKETS;
    for (goperl_intern_ent_t *e = goperl_intern_v[h]; e; e = e->next)
        if (strlen(e->s) == n && memcmp(e->s, p, n) == 0) return e->s;
    goperl_intern_ent_t *e =
        (goperl_intern_ent_t *)malloc(sizeof(goperl_intern_ent_t) + n);
    if (!e) return "";
    memcpy(e->s, p, n);
    e->s[n] = '\0';
    e->next = goperl_intern_v[h];
    goperl_intern_v[h] = e;
    return e->s;
}

/* Intern a guest (ptr<<32|len) packed string; NULL when packed is 0. */
static const char *goperl_intern_packed(goperl_frame_t *f, uint64_t packed) {
    if (!packed) return 0;
    const char *p = (const char *)goperl_api_v->guest_mem(f, packed >> 32);
    if (!p) return 0;
    return goperl_intern(p, (size_t)(uint32_t)packed);
}

static void goperl_op_fill_cop(goperl_frame_t *f, OP *o) {
    o->gop_iscop = 1;
    o->cop_line = (U32)goperl_do_op(f, GOPERL_OP_COP_LINE, o->gop, 0, 0, 0);
    o->cop_hints = (U32)goperl_do_op(f, GOPERL_OP_COP_HINTS, o->gop, 0, 0, 0);
    o->cop_file = goperl_intern_packed(
        f, goperl_do_op(f, GOPERL_OP_COP_FILE, o->gop, 0, 0, 0));
    o->cop_stashpv = goperl_intern_packed(
        f, goperl_do_op(f, GOPERL_OP_COP_STASHPV, o->gop, 0, 0, 0));
    /* dists compare these with strEQ and %s them without null checks */
    if (!o->cop_file) o->cop_file = "";
    if (!o->cop_stashpv) o->cop_stashpv = "";
}

static void goperl_op_pull(goperl_frame_t *f, OP *o) {
    uint64_t tok = o->gop;
    uint64_t fields = goperl_do_op(f, GOPERL_OP_OP_FIELDS, tok, 0, 0, 0);
    o->op_type = (U16)(fields & 0xFFFF);
    o->op_flags = (U8)(fields >> 16);
    o->op_private = (U8)(fields >> 24);
    o->op_targ = (UV)(uint32_t)(fields >> 32);
    o->op_next = goperl_op_shadow(
        f, goperl_do_op(f, GOPERL_OP_OP_PTR, tok, 0, 0, 0));
    o->op_first = goperl_op_shadow(
        f, goperl_do_op(f, GOPERL_OP_OP_PTR, tok, 2, 0, 0));
    o->op_redoop = goperl_op_shadow(
        f, goperl_do_op(f, GOPERL_OP_OP_PTR, tok, 3, 0, 0));
    /* op_last / op_other share storage; the guest returns whichever the
     * op's class carries */
    OP *lastish = goperl_op_shadow(
        f, goperl_do_op(f, GOPERL_OP_OP_PTR, tok, 4, 0, 0));
    if (!lastish)
        lastish = goperl_op_shadow(
            f, goperl_do_op(f, GOPERL_OP_OP_PTR, tok, 5, 0, 0));
    o->op_last = lastish;
    /* the direct sibling link, kept shadow-local for locally-relinked
     * chains: moresib mirrors OpHAS_SIBLING, sibparent the target */
    o->op_moresib =
        (U8)goperl_do_op(f, GOPERL_OP_OP_PTR, tok, 6, 0, 0);
    o->op_sibparent =
        o->op_moresib ? goperl_op_shadow(f, goperl_do_op(
                                                f, GOPERL_OP_OP_PTR, tok, 1,
                                                0, 0))
                      : 0;
    goperl_op_base_capture(o);
}

static OP *goperl_op_shadow_full(goperl_frame_t *f, uint64_t tok) {
    OP *o = goperl_op_shadow(f, tok);
    if (!o || o->gop_full) return o;
    o->gop_full = 1;
    goperl_op_pull(f, o);
    int t = o->op_type ? o->op_type : (int)o->op_targ;
    if (!o->gop_iscop && (t == OP_NEXTSTATE || t == OP_DBSTATE))
        goperl_op_fill_cop(f, o);
    return o;
}

/* Refresh after a guest optree operation: the guest may have relinked or
 * retyped the participants. */
static void goperl_op_refresh(goperl_frame_t *f, OP *o) {
    if (o && o->gop && o->gop_full) goperl_op_pull(f, o);
}
static void goperl_op_refresh_tok(goperl_frame_t *f, uint64_t tok) {
    if (!tok) return;
    goperl_op_refresh(f, goperl_op_shadow(f, tok));
}

/* Materialize a token the CALLER knows is a COP (PL_curcop, blk_oldcop):
 * static interpreter COPs like PL_compiling have op_type 0, so cop-ness
 * cannot be inferred from the type. */
static COP *goperl_cop_shadow(goperl_frame_t *f, uint64_t tok) {
    OP *o = goperl_op_shadow_full(f, tok);
    if (o && !o->gop_iscop) goperl_op_fill_cop(f, o);
    return (COP *)o;
}

/* In an embedded interpreter PL_curcop CAN be NULL: eval_sv driven from C
 * enters with no compiling COP, so the eval context's blk_oldcop is NULL
 * and leaveeval restores that. Native perl code dereferences PL_curcop
 * without checks, so hand it a stable dummy COP (line 0, package main)
 * instead — the same shape modules already handle for optimized-away
 * COPs. */
__attribute__((weak)) OP goperl_nullcop_v = {
    0, 0, 0, 0, OP_NEXTSTATE, 0, 0, 0, /* base scalars */
    1 /* gop_full */, 1 /* gop_iscop */, 0 /* moresib */, 0 /* gop */,
    0 /* gop_base */, 0 /* op_first */, {0}, 0 /* op_redoop */,
    0 /* cop_line */, 0 /* cop_hints */, 0 /* cop_warnings */, "", "main"};
static COP *goperl_curcop(goperl_frame_t *f) {
    uint64_t tok = goperl_do_op(f, GOPERL_OP_PLVAR_GET, GOPERL_PL_CURCOP, 0,
                                0, 0);
    if (!tok) return (COP *)&goperl_nullcop_v;
    return goperl_cop_shadow(f, tok);
}

/* PL_curcop must be an LVALUE: dists save/override it around a block
 * (Cpanel::JSON::XS points it at a hint-stripped copy while sorting).
 * Reads refresh the slot from the guest, so an override only holds until
 * the next PL_curcop read — which is exactly the Cpanel pattern (assign,
 * run host-local code, restore). Note the override cannot influence GUEST
 * execution (guest ops read the real interpreter's curcop); the one
 * observable deviation is `use bytes` callers of canonical JSON encoding,
 * whose sort comparisons keep byte semantics. */
#define goperl_curcop_slot_v (*(COP **)&GOPERL_SHARED->curcop_slot)
static COP **goperl_curcop_ref(goperl_frame_t *f) {
    goperl_curcop_slot_v = goperl_curcop(f);
    return &goperl_curcop_slot_v;
}

static OP *goperl_op_sibling(goperl_frame_t *f, OP *o) {
    if (!o) return 0;
    /* a full shadow carries the (possibly locally rewritten) sibling link */
    if (o->gop_full) return o->op_moresib ? o->op_sibparent : 0;
    if (!o->gop) return o->op_moresib ? o->op_sibparent : 0;
    return goperl_op_shadow_full(
        f, goperl_do_op(f, GOPERL_OP_OP_PTR, o->gop, 1, 0, 0));
}

#define OP_NAME(o) \
    (PL_op_name[(o)->op_type ? (o)->op_type : (int)(o)->op_targ])

/* Op-tree dumping needs interpreter internals the SDK does not model. */
static void goperl_do_op_dump(I32 level, void *fp, const OP *o)
    __attribute__((unused));
static void goperl_do_op_dump(I32 level, void *fp, const OP *o) {
    (void)level;
    fprintf((FILE *)fp, "<op dump unavailable in the go-perl XS SDK: %s>\n",
            o ? OP_NAME(o) : "(null)");
}
#define do_op_dump(l, f, o) goperl_do_op_dump((l), (void *)(f), (const OP *)(o))

/* ---- PL_ppaddr proxy / pp dispatch ---------------------------------------
 * Every slot of the proxy table is a distinct stub carrying its op type
 * (goperl_ppslots.h), so both idioms work:
 *   - a module CALLS a slot (run_original_op after saving the table):
 *     dispatch runs the guest's real pp for that type;
 *   - a module WRITES a slot (installing a pp hook): the funnel notices
 *     the diff and routes guest-side executions of that type back to the
 *     module's function (see goperl_ppaddr_sync).
 * Calls with a module-built scratch op (gop == 0, the pp_flop fake-op
 * idiom) run the op on the guest stack via RUN_PP instead. */

static OP *goperl_pp_dispatch(pTHX_ int type);
#include "goperl_ppslots.h"

__attribute__((weak)) Perl_ppaddr_t goperl_ppaddr_base_v[OP_max];
__attribute__((weak)) uint8_t goperl_pp_installed_v[OP_max];
__attribute__((weak)) int32_t goperl_ppaddr_inited_v = 0;

static OP *goperl_pp_dispatch(pTHX_ int type) {
    OP *cur = _gof->plop;
    if (cur && cur->gop) {
        /* real guest op under execution: run the true pp for `type` */
        uint64_t next = goperl_do_op(_gof, GOPERL_OP_RUN_ORIGINAL,
                                     (uint64_t)(uint32_t)type, 0, 0, 0);
        return goperl_op_shadow_full(_gof, next);
    }
    /* module-built scratch op */
    I32 nargs;
    switch (type) {
    case OP_FLOP:
        nargs = 2;
        break;
    default:
        goperl_croakf(_gof, "pp dispatch: scratch op %d is not supported",
                      type);
    }
    SV **base = (SV **)_gof->st;
    if ((I32)(_gof->psp - base) < nargs)
        goperl_croakf(_gof, "pp dispatch: stack underflow");
    char buf[64 * 4];
    for (I32 i = 0; i < nargs; i++) {
        uint32_t tok =
            (uint32_t)(uint64_t)(uintptr_t)_gof->psp[i - nargs + 1];
        memcpy(buf + i * 4, &tok, 4);
    }
    uint32_t flags = cur ? cur->op_flags : 0;
    uint64_t packed = goperl_do_op(
        _gof, GOPERL_OP_RUN_PP,
        ((uint64_t)flags << 32) | (uint32_t)type,
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
    if (!goperl_ppaddr_inited_v) {
        goperl_ppaddr_inited_v = 1;
        GOPERL_PP_SLOT_INIT(goperl_ppaddr_v);
        memcpy(goperl_ppaddr_base_v, goperl_ppaddr_v,
               sizeof(goperl_ppaddr_base_v));
    }
    return goperl_ppaddr_v;
}
#define PL_ppaddr (goperl_ppaddr_get(_gof))

/* Detect module writes to the proxy table and (de)register them as guest
 * pp hooks. Called from the funnel flush. Uses the raw vtable (never the
 * funnel) to avoid recursion. */
static void goperl_ppaddr_sync(goperl_frame_t *f) {
    if (!goperl_ppaddr_inited_v) return;
    for (int i = 0; i < OP_max; i++) {
        int hooked = goperl_ppaddr_v[i] != goperl_ppaddr_base_v[i];
        if (hooked && !goperl_pp_installed_v[i]) {
            goperl_pp_installed_v[i] = 1;
            goperl_api_v->pp_hook_set(f, i, (void *)goperl_ppaddr_v[i]);
        } else if (!hooked && goperl_pp_installed_v[i]) {
            goperl_pp_installed_v[i] = 0;
            goperl_api_v->pp_hook_set(f, i, 0);
        }
    }
}

/* ---- save-stack destructors ----------------------------------------------
 * save_destructor_x registers (fn, arg) host-side; the guest savestack
 * carries the id and fires it back (callback method -4) when the scope
 * pops — normal exit and die unwinding both included, like real perl. */
#define GOPERL_DTOR_MAX 4096
typedef struct {
    void (*fn)(pTHX_ void *);
    void *arg;
    U8 used;
} goperl_dtor_t;
#define goperl_dtor_v ((goperl_dtor_t *)GOPERL_SHARED->dtors)
#define goperl_dtor_hint_v (GOPERL_SHARED->dtor_hint)
typedef char goperl_dtor_fits
    [(sizeof(goperl_dtor_t) * GOPERL_DTOR_MAX <=
      sizeof(((goperl_shared_t *)0)->dtors))
         ? 1
         : -1] __attribute__((unused));

static void goperl_save_destructor_x(goperl_frame_t *f,
                                     void (*fn)(pTHX_ void *), void *arg) {
    int32_t id = -1;
    for (int32_t k = 0; k < GOPERL_DTOR_MAX; k++) {
        int32_t i = (goperl_dtor_hint_v + k) % GOPERL_DTOR_MAX;
        if (!goperl_dtor_v[i].used) {
            id = i;
            break;
        }
    }
    if (id < 0) goperl_croakf(f, "save_destructor_x: destructor table full");
    goperl_dtor_hint_v = id + 1;
    goperl_dtor_v[id].fn = fn;
    goperl_dtor_v[id].arg = arg;
    goperl_dtor_v[id].used = 1;
    goperl_do_op(f, GOPERL_OP_SAVE_DESTRUCTOR, (uint64_t)(uint32_t)(id + 1),
                 0, 0, 0);
}
#define save_destructor_x(fn, p) goperl_save_destructor_x(_gof, (fn), (p))
#define SAVEDESTRUCTOR_X(fn, p) goperl_save_destructor_x(_gof, (fn), (p))

/* Loader entry: run a fired destructor (from guest scope pop). */
__attribute__((weak, visibility("default"), used)) void
__goperl_dtor_invoke(goperl_frame_t *f, uint32_t id1) {
    if (id1 == 0 || id1 > GOPERL_DTOR_MAX) return;
    jmp_buf jb;
    void *prev_jb = f->jb;
    goperl_frame_t *prev_frame = goperl_cur_frame_v;
    f->jb = (void *)&jb;
    /* current frame before the table read: the destructor table is
     * per-instance state reached through it */
    goperl_cur_frame_v = f;
    goperl_dtor_t d = goperl_dtor_v[id1 - 1];
    goperl_dtor_v[id1 - 1].used = 0;
    if (!d.used || !d.fn) {
        f->jb = prev_jb;
        goperl_cur_frame_v = prev_frame;
        return;
    }
    if (setjmp(jb)) {
        fprintf(stderr, "goperl native XS: croak in save-stack destructor: %s\n",
                f->err);
        f->failed = 0;
        f->jb = prev_jb;
        goperl_cur_frame_v = prev_frame;
        return;
    }
    d.fn(f, d.arg);
    f->jb = prev_jb;
    goperl_cur_frame_v = prev_frame;
}

/* Loader entry: close any activation levels an XSUB left open. xsubpp's
 * PPCODE trailer is `PUTBACK; return;` with no XSRETURN, so the entry
 * bookkeeping (depth, saved prev_frame/jb/hostsave_base) is still armed
 * when the C function returns; the loader calls this after every XSUB.
 * Each xs_leave unwinds that activation's host saves, which restores the
 * previous prev_frame/jb/base — so the loop also settles bare-returning
 * NESTED direct calls, one level at a time. */
__attribute__((weak, visibility("default"), used)) uint32_t
__goperl_xs_epilogue(goperl_frame_t *f) {
    uint32_t n = 0;
    while (goperl_cur_frame_v == f && goperl_xs_depth_v > 0 && n < 64) {
        goperl_frame_t *prev = (goperl_frame_t *)f->prev_frame;
        goperl_xs_leave(f);
        n++;
        if (prev != f) break; /* settled to the caller; stop */
    }
    return n;
}

/* Loader entry: run one mirror-MAGIC hook (svt_set / svt_free — the
 * int (*)(pTHX_ SV*, MAGIC*) shape) with the current frame installed and a
 * croak guard armed. The loader calls this instead of the raw function
 * pointer so interpreter macros inside the hook see the right frame. */
__attribute__((weak, visibility("default"), used)) uint32_t
__goperl_svt_invoke(goperl_frame_t *f, void *fnv, uint64_t sv, void *mg) {
    int (*fn)(pTHX_ SV *, MAGIC *) = (int (*)(pTHX_ SV *, MAGIC *))fnv;
    jmp_buf jb;
    void *prev_jb = f->jb;
    goperl_frame_t *prev_frame = goperl_cur_frame_v;
    f->jb = (void *)&jb;
    goperl_cur_frame_v = f;
    if (setjmp(jb)) {
        fprintf(stderr, "goperl native XS: croak in magic hook: %s\n", f->err);
        f->failed = 0;
        f->jb = prev_jb;
        goperl_cur_frame_v = prev_frame;
        return 0;
    }
    fn(f, (SV *)(uintptr_t)sv, (MAGIC *)mg);
    f->jb = prev_jb;
    goperl_cur_frame_v = prev_frame;
    return 1;
}

/* Loader entry: run a module pp hook for one guest op execution.
 * Returns the next-op token (0 = NULL); on croak, f->failed is set and
 * the message is in f->err. */
__attribute__((weak, visibility("default"), used)) uint64_t
__goperl_pp_invoke(goperl_frame_t *f, void *fnv) {
    OP *(*fn)(pTHX) = (OP * (*)(pTHX)) fnv;
    jmp_buf jb;
    f->jb = (void *)&jb;
    f->prev_frame = (void *)goperl_cur_frame_v;
    goperl_cur_frame_v = f;
    goperl_xs_depth_v++;
    f->hostsave_base = goperl_hostsave_n;
    if (setjmp(jb)) {
        goperl_hostsave_unwind_to(f, f->hostsave_base);
        goperl_xs_leave(f);
        return 0;
    }
    f->plop = goperl_op_shadow_full(f, f->hook_op_tok);
    if (f->plop && f->plop->gop) {
        /* a per-op-hooked shadow lives across activations; the guest may
         * have relinked it since (peephole, other compile passes) */
        goperl_opreg_ent_t *pe =
            (goperl_opreg_ent_t *)((char *)f->plop -
                                   offsetof(goperl_opreg_ent_t, op));
        if (pe->persist) goperl_op_refresh(f, f->plop);
    }
    OP *next = fn(f);
    uint64_t tok = next ? next->gop : 0;
    goperl_xs_leave(f);
    return tok;
}

/* ---- savestack scratch allocations (SSNEW / SSPTR) -----------------------
 * Real perl hands out scratch memory ON the savestack, reclaimed at scope
 * pop. Here the block is host memory and an internal save-stack destructor
 * (registered BEFORE any user destructor in the same scope, so it fires
 * AFTER it) frees the block at the same point. */
#define GOPERL_SSNEW_MAX 4096
typedef struct {
    void *mem;
} goperl_ssnew_t;
#define goperl_ssnew_v ((goperl_ssnew_t *)GOPERL_SHARED->ssnew)
#define goperl_ssnew_hint_v (GOPERL_SHARED->ssnew_hint)
typedef char goperl_ssnew_fits
    [(sizeof(goperl_ssnew_t) * GOPERL_SSNEW_MAX <=
      sizeof(((goperl_shared_t *)0)->ssnew))
         ? 1
         : -1] __attribute__((unused));

static void goperl_ssnew_free_cb(pTHX_ void *p) {
    int32_t ix = (int32_t)(intptr_t)p;
    (void)_gof;
    if (ix >= 0 && ix < GOPERL_SSNEW_MAX && goperl_ssnew_v[ix].mem) {
        free(goperl_ssnew_v[ix].mem);
        goperl_ssnew_v[ix].mem = 0;
    }
}

static SSize_t goperl_ssnew(goperl_frame_t *f, size_t size) {
    int32_t ix = -1;
    for (int32_t k = 0; k < GOPERL_SSNEW_MAX; k++) {
        int32_t i = (goperl_ssnew_hint_v + k) % GOPERL_SSNEW_MAX;
        if (!goperl_ssnew_v[i].mem) {
            ix = i;
            break;
        }
    }
    if (ix < 0) goperl_croakf(f, "SSNEW: scratch table full");
    goperl_ssnew_hint_v = ix + 1;
    goperl_ssnew_v[ix].mem = calloc(1, size);
    if (!goperl_ssnew_v[ix].mem) goperl_croakf(f, "SSNEW: out of memory");
    goperl_save_destructor_x(f, goperl_ssnew_free_cb,
                             (void *)(intptr_t)ix);
    return (SSize_t)ix;
}
#define SSNEW(size) goperl_ssnew(_gof, (size_t)(size))
#define SSNEWa(size, align) goperl_ssnew(_gof, (size_t)(size))
#define SSPTR(ix, type) ((type)goperl_ssnew_v[(ix)].mem)
#define MEM_ALIGNBYTES 8

/* ---- context stack shadows -----------------------------------------------
 * PERL_CONTEXT / PERL_SI mirrors materialized from the live interpreter,
 * refreshed whenever any guest operation may have changed them (a
 * generation counter bumped by the funnel). Allocations go to an arena
 * freed when the outermost native frame returns. */

typedef struct goperl_si PERL_SI;
typedef struct goperl_context {
    U8 cx_type;
    COP *blk_oldcop;
    struct {
        CV *cv;
    } blk_sub;
    struct {
        OP *my_op;
    } blk_loop;
} PERL_CONTEXT;
struct goperl_si {
    uint64_t tok;
    PERL_CONTEXT *si_cxstack;
    I32 si_cxix;
    I32 si_type;
    PERL_SI *si_prev;
};
#undef CxTYPE
#define CxTYPE(cx) ((cx)->cx_type)
#define CxMULTICALL(cx) 0

typedef struct goperl_arena_blk {
    struct goperl_arena_blk *next;
} goperl_arena_blk_t;
#define goperl_arena_v (*(goperl_arena_blk_t **)&GOPERL_SHARED->arena_head)
#define goperl_si_cache_v (*(PERL_SI **)&GOPERL_SHARED->si_cache)
/* zero-start is safe: the si cache is only trusted when non-NULL */
#define goperl_gen_v (GOPERL_SHARED->gen)
#define goperl_si_gen_v (GOPERL_SHARED->si_gen)

static void goperl_gen_bump(void) { goperl_gen_v++; }

static void *goperl_arena_alloc(size_t n) {
    goperl_arena_blk_t *b =
        (goperl_arena_blk_t *)calloc(1, sizeof(goperl_arena_blk_t) + n);
    if (!b) return 0;
    b->next = goperl_arena_v;
    goperl_arena_v = b;
    return (void *)(b + 1);
}

static void goperl_arena_free_all(void) {
    while (goperl_arena_v) {
        goperl_arena_blk_t *n = goperl_arena_v->next;
        free(goperl_arena_v);
        goperl_arena_v = n;
    }
    goperl_si_cache_v = 0;
    goperl_si_gen_v = 0;
}

/* ---- HE shadows (see the HE/HEK typedefs near the top) ------------------ */

#ifndef HVhek_UTF8
#define HVhek_UTF8 0x01
#endif

static HE *goperl_he_shadow(goperl_frame_t *f, HV *hv, uint64_t tok) {
    if (!tok) return 0;
    HE *he = (HE *)goperl_arena_alloc(sizeof(HE));
    if (!he) goperl_croakf(f, "HE shadow: out of memory");
    he->goperl_tok = tok;
    he->he_valu.hent_val =
        hv ? GOPERL_SV(goperl_do_op(f, GOPERL_OP_HV_ITERVAL, GOPERL_TOK(hv),
                                    tok, 0, 0))
           : GOPERL_SV(goperl_do_op(f, GOPERL_OP_HE_VAL, tok, 0, 0, 0));
    return he;
}

/* The key bytes materialize lazily, laid out like a real perl HEK: the
 * bytes, a NUL, then the flags byte (HVhek_UTF8). */
static HEK *goperl_he_hek(goperl_frame_t *f, HE *he) {
    if (he->hent_hek) return he->hent_hek;
    uint64_t keytok = goperl_do_op(f, GOPERL_OP_HV_ITERKEYSV, he->goperl_tok,
                                   0, 0, 0);
    uint64_t len = 0;
    const char *p = keytok ? goperl_api_v->sv_pv(f, keytok, &len) : 0;
    HEK *hek = (HEK *)goperl_arena_alloc(sizeof(HEK) + (size_t)len + 2);
    if (!hek) goperl_croakf(f, "HEK shadow: out of memory");
    hek->hek_hash = 0;
    hek->hek_len = (I32)len;
    if (p) memcpy(hek->hek_key, p, (size_t)len);
    hek->hek_key[len] = 0;
    hek->hek_key[len + 1] =
        (keytok &&
         (goperl_do_op(f, GOPERL_OP_SV_INFO, keytok, 0, 0, 0) &
          GOPERL_INFO_UTF8))
            ? HVhek_UTF8
            : 0;
    he->hent_hek = hek;
    return hek;
}

#define hv_fetch_ent(hv, keysv, lval, hash)                              \
    goperl_he_shadow(_gof, (HV *)0,                                      \
                     goperl_op0(GOPERL_OP_HV_FETCH_ENT, GOPERL_TOK(hv),  \
                                (uint32_t)GOPERL_TOK(keysv) |            \
                                    ((uint64_t)((lval) ? 1 : 0) << 32)))
#define hv_store_ent(hv, keysv, sv, hash)                                 \
    goperl_he_shadow(_gof, (HV *)0,                                       \
                     goperl_op0(GOPERL_OP_HV_STORE_ENT, GOPERL_TOK(hv),   \
                                (GOPERL_TOK(keysv) << 32) |               \
                                    (uint32_t)GOPERL_TOK(sv)))
#define hv_iternext(hv)                                                  \
    goperl_he_shadow(_gof, (HV *)(hv),                                   \
                     goperl_op0(GOPERL_OP_HV_ITERNEXT, GOPERL_TOK(hv), 0))
#define hv_iternext_sv hv_iternext
#define HeVAL(he) ((he)->he_valu.hent_val)
#define HeNEXT(he) ((he)->hent_next)
#define HeKEY_hek(he) goperl_he_hek(_gof, (HE *)(he))
#define HeKEY(he) ((void *)HeKEY_hek(he)->hek_key)
#define HeKEY_sv(he) (*(SV **)HeKEY(he))
#define HeKLEN(he) (HeKEY_hek(he)->hek_len)
#define HeHASH(he) (HeKEY_hek(he)->hek_hash)
#define HeKUTF8(he)                                                       \
    (HeKEY_hek(he)->hek_key[HeKEY_hek(he)->hek_len + 1] & HVhek_UTF8)
#define HeKWASUTF8(he) 0
#define HeUTF8(he) \
    ((HeKLEN(he) == HEf_SVKEY) ? SvUTF8(HeKEY_sv(he)) : (U32)HeKUTF8(he))
#define HePV(he, lenvar) ((lenvar) = (STRLEN)HeKLEN(he), (char *)HeKEY(he))
#define HeSVKEY(he) \
    ((HeKLEN(he) == HEf_SVKEY) ? *(SV **)HeKEY(he) : (SV *)0)
#define HeSVKEY_set(he, sv)                                     \
    ((he)->hent_hek->hek_len = HEf_SVKEY,                       \
     *(SV **)(he)->hent_hek->hek_key = (sv))

/* The key as an SV: guest-backed shadows go through the guest (a mortal
 * copy, utf8-faithful); SVKEY entries return the stored SV. */
static SV *goperl_hv_iterkeysv(goperl_frame_t *f, HE *he) {
    if (he->goperl_tok)
        return GOPERL_SV(goperl_do_op(f, GOPERL_OP_HV_ITERKEYSV,
                                      he->goperl_tok, 0, 0, 0));
    if (he->hent_hek && he->hent_hek->hek_len == HEf_SVKEY)
        return *(SV **)he->hent_hek->hek_key;
    if (he->hent_hek) {
        SV *sv = GOPERL_SV(goperl_do_op(
            f, GOPERL_OP_NEW_PVN, 0, (uint64_t)he->hent_hek->hek_len,
            he->hent_hek->hek_key, (uint64_t)he->hent_hek->hek_len));
        goperl_do_op(f, GOPERL_OP_SV_MORTAL, GOPERL_TOK(sv), 0, 0, 0);
        return sv;
    }
    return 0;
}
#define hv_iterkeysv(he) goperl_hv_iterkeysv(_gof, (HE *)(he))
#define HeSVKEY_force(he) hv_iterkeysv(he)
#define hv_iterval(hv, he) HeVAL(he)

static PERL_SI *goperl_si_materialize(goperl_frame_t *f, uint64_t si_tok,
                                      int depth) {
    if (depth > 16) return 0;
    uint64_t tok = si_tok
                       ? si_tok
                       : goperl_do_op(f, GOPERL_OP_SI_GET, 0, 0, 0, 0);
    if (!tok) return 0;
    PERL_SI *si = (PERL_SI *)goperl_arena_alloc(sizeof(PERL_SI));
    if (!si) goperl_croakf(f, "stackinfo shadow: out of memory");
    si->tok = tok;
    si->si_type =
        (I32)(int64_t)goperl_do_op(f, GOPERL_OP_SI_GET, tok, 2, 0, 0);
    si->si_cxix =
        (I32)(int64_t)goperl_do_op(f, GOPERL_OP_SI_GET, tok, 3, 0, 0);
    I32 n = si->si_cxix + 1;
    si->si_cxstack = (PERL_CONTEXT *)goperl_arena_alloc(
        sizeof(PERL_CONTEXT) * (size_t)(n > 0 ? n : 1));
    if (!si->si_cxstack) goperl_croakf(f, "context shadow: out of memory");
    for (I32 i = 0; i < n; i++) {
        PERL_CONTEXT *cx = &si->si_cxstack[i];
        uint64_t t = goperl_do_op(f, GOPERL_OP_CX_FIELDS, tok,
                                  (uint64_t)(int64_t)i, 0, 0);
        cx->cx_type = (U8)t;
        cx->blk_oldcop = goperl_cop_shadow(
            f, goperl_do_op(f, GOPERL_OP_CX_PTR, tok,
                            ((uint64_t)(uint32_t)i << 8) | 0, 0, 0));
        cx->blk_sub.cv = (CV *)GOPERL_SV(goperl_do_op(
            f, GOPERL_OP_CX_PTR, tok, ((uint64_t)(uint32_t)i << 8) | 1, 0,
            0));
        cx->blk_loop.my_op = goperl_op_shadow_full(
            f, goperl_do_op(f, GOPERL_OP_CX_PTR, tok,
                            ((uint64_t)(uint32_t)i << 8) | 2, 0, 0));
    }
    uint64_t prev =
        goperl_do_op(f, GOPERL_OP_SI_GET, tok, 1, 0, 0);
    si->si_prev = prev ? goperl_si_materialize(f, prev, depth + 1) : 0;
    return si;
}

static PERL_SI *goperl_si_cur(goperl_frame_t *f) {
    if (goperl_si_cache_v && goperl_si_gen_v == goperl_gen_v)
        return goperl_si_cache_v;
    goperl_si_cache_v = goperl_si_materialize(f, 0, 0);
    goperl_si_gen_v = goperl_gen_v;
    if (!goperl_si_cache_v)
        goperl_croakf(f, "cannot materialize the context stack");
    return goperl_si_cache_v;
}
#define PL_curstackinfo (goperl_si_cur(_gof))
#define cxstack (goperl_si_cur(_gof)->si_cxstack)
#define cxstack_ix (goperl_si_cur(_gof)->si_cxix)

/* ---- MAGIC --------------------------------------------------------------- */

#define sv_magicext(sv, obj, how, vtbl, name, namlen)                     \
    (goperl_flush(_gof),                                                  \
     goperl_api_v->magic_ext(_gof, GOPERL_TOK(sv), GOPERL_TOK(obj),       \
                             (int32_t)(how), (const void *)(vtbl),        \
                             (const char *)(name), (int64_t)(namlen)))
/* SvMAGIC is WRITABLE (DBI splices tie magic by assigning the chain
 * head): reads load a synced slot; a changed slot is written back through
 * magic_set_head on the next SvMAGIC use. Guest-attached magic (ties) is
 * not mirrored, so such splices see only host-attached entries. */
#define goperl_svmagic_sv_v (GOPERL_SHARED->svmagic_sv)
#define goperl_svmagic_val_v (GOPERL_SHARED->svmagic_val)
#define goperl_svmagic_orig_v (GOPERL_SHARED->svmagic_orig)
static MAGIC **goperl_svmagic_lv(goperl_frame_t *f, SV *sv) {
    if (goperl_svmagic_sv_v && goperl_svmagic_val_v != goperl_svmagic_orig_v)
        goperl_api_v->magic_set_head(f, goperl_svmagic_sv_v,
                                     goperl_svmagic_val_v);
    goperl_flush(f);
    goperl_svmagic_val_v = goperl_api_v->magic_chain(f, GOPERL_TOK(sv));
    goperl_svmagic_orig_v = goperl_svmagic_val_v;
    goperl_svmagic_sv_v = GOPERL_TOK(sv);
    return &goperl_svmagic_val_v;
}
#define SvMAGIC(sv) (*goperl_svmagic_lv(_gof, (SV *)(sv)))
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
     goperl_api_v->magic_del(_gof, GOPERL_TOK(sv), (int32_t)(how), 0), \
     (void)goperl_op0(GOPERL_OP_SV_UNMAGIC, GOPERL_TOK(sv),            \
                      (uint64_t)(uint32_t)(how)),                      \
     0)
#define sv_unmagicext(sv, how, vtbl)                                   \
    (goperl_flush(_gof),                                               \
     goperl_api_v->magic_del(_gof, GOPERL_TOK(sv), (int32_t)(how),     \
                             (const void *)(vtbl)),                    \
     0)
/* Replace the mirror chain's head (the SvMAGIC_set idiom for unlinking an
 * entry the module found via SvMAGIC). The unlinked node's cleanup stays
 * with the module (Safefree), exactly as with real perl. */
#define SvMAGIC_set(sv, mg) \
    (goperl_flush(_gof),    \
     goperl_api_v->magic_set_head(_gof, GOPERL_TOK(sv), (MAGIC *)(mg)))
#define PERL_MAGIC_backref '<'
/* Plain sv_magic attaches BEHAVIORAL core magic (PERL_MAGIC_tied being
 * the load-bearing case: DBI ties its outer handles) — that must land on
 * the real guest SV, not the host mirror, or the interpreter never
 * dispatches FETCH/STORE. Custom-vtbl magic stays host-side via
 * sv_magicext. */
static void goperl_sv_magic(goperl_frame_t *f, SV *sv, SV *obj, int how,
                            const char *name, I32 namlen) {
    goperl_do_op(f, GOPERL_OP_SV_MAGIC_STD, GOPERL_TOK(sv),
                 ((uint64_t)(uint32_t)how << 32) | (uint32_t)GOPERL_TOK(obj),
                 name ? name : "", name ? (uint64_t)namlen : 0);
    /* Mirror it too: XS reads the attachment back with mg_find (DBI
     * resolves its inner handle through the tie's mg_obj). */
    goperl_flush(f);
    goperl_api_v->magic_ext(f, GOPERL_TOK(sv), GOPERL_TOK(obj), how, 0,
                            name, name ? (int64_t)namlen : 0);
}
#define sv_magic(sv, obj, how, name, namlen)                          \
    goperl_sv_magic(_gof, (SV *)(sv), (SV *)(obj), (int)(how),        \
                    (const char *)(name), (I32)(namlen))
/* Any-magic test: the anchor for mirrored magic marks the guest SV
 * RMAGICAL, so this covers every SV whose magic the host can see. */
#define SvMAGICAL(sv) SvRMAGICAL(sv)

/* ---- v5: the Class::MOP/Moose surface ----------------------------------- */

#define SvREFCNT(sv) \
    ((U32)goperl_op0(GOPERL_OP_SV_REFCNT, GOPERL_TOK(sv), 0))

static HV *goperl_gv_stashsv(goperl_frame_t *f, SV *name, I32 flags) {
    uint64_t len = 0;
    const char *p = goperl_api_v->sv_pv(f, GOPERL_TOK(name), &len);
    return (HV *)GOPERL_SV(goperl_do_op(f, GOPERL_OP_GV_STASHPV,
                                        (uint64_t)(int64_t)flags, 0, p, len));
}
#define gv_stashsv(sv, flags) goperl_gv_stashsv(_gof, (SV *)(sv), (I32)(flags))

/* Expand a stash stub/constant slot into a real glob in place. */
#define gv_init_pvn(gv, stash, name, len, flags)                          \
    ((void)goperl_ops(GOPERL_OP_GV_INIT, GOPERL_TOK(gv),                  \
                      ((uint64_t)(U32)(flags) << 32) |                    \
                          (GOPERL_TOK(stash) & 0xFFFFFFFFu),              \
                      (name), (len)))
#define gv_init(gv, stash, name, len, multi) \
    gv_init_pvn((gv), (stash), (name), (len), (multi) ? GV_ADDMULTI : 0)

/* Overload (amagic) surface: Gv_AMG asks the guest whether the stash has a
 * freshly-updated overload table; SvAMAGIC_on/off flip the flag through the
 * guest so the stash-level bookkeeping stays coherent. */
#define Gv_AMG(stash) \
    ((int)goperl_op0(GOPERL_OP_GV_AMG, GOPERL_TOK(stash), 0))
#define SvAMAGIC_on(sv) \
    ((void)goperl_op0(GOPERL_OP_SV_AMAGIC_SET, GOPERL_TOK(sv), 1))
#define SvAMAGIC_off(sv) \
    ((void)goperl_op0(GOPERL_OP_SV_AMAGIC_SET, GOPERL_TOK(sv), 0))

/* mro package generation, read the way Moose's mop.c reads it:
 * HvAUX(stash)->xhv_mro_meta->pkg_gen. Only that chain is modeled — HvAUX
 * returns a scratch one-deep view refreshed from the guest on every call. */
struct mro_meta {
    U32 pkg_gen;
    U32 cache_gen; /* always 0: method-cache users revalidate every time */
};
struct xpvhv_aux { struct mro_meta *xhv_mro_meta; };
static struct {
    struct mro_meta meta;
    struct xpvhv_aux aux;
} goperl_hvaux_v;
static struct xpvhv_aux *goperl_hvaux(goperl_frame_t *f, HV *hv) {
    uint64_t r = goperl_do_op(f, GOPERL_OP_HV_PKG_GEN, GOPERL_TOK(hv), 0, 0, 0);
    if (r >> 32) {
        goperl_hvaux_v.meta.pkg_gen = (U32)r;
        goperl_hvaux_v.aux.xhv_mro_meta = &goperl_hvaux_v.meta;
    } else {
        goperl_hvaux_v.aux.xhv_mro_meta = 0;
    }
    return &goperl_hvaux_v.aux;
}
#define HvAUX(hv) goperl_hvaux(_gof, (HV *)(hv))
/* HvMROMETA never returns NULL in core; hand out the shim meta. */
static struct mro_meta *goperl_hvmrometa(goperl_frame_t *f, HV *hv) {
    struct xpvhv_aux *aux = goperl_hvaux(f, hv);
    return aux->xhv_mro_meta ? aux->xhv_mro_meta
                             : (goperl_hvaux_v.meta.pkg_gen = 0,
                                goperl_hvaux_v.meta.cache_gen = 0,
                                &goperl_hvaux_v.meta);
}
#define HvMROMETA(hv) goperl_hvmrometa(_gof, (HV *)(hv))

/* The SDK's hv/he ops take the key bytes and ignore precomputed hashes, so
 * prehashing collapses to zero. */
#define PERL_HASH(hash, str, len) ((hash) = 0)

/* Compiled-regex test (ppport's NEED_SvRX fallback defers to this). */
#define SvRXOK(sv) \
    (SvROK(sv) && SvTYPE(SvRV((SV *)(sv))) == SVt_REGEXP)

#define DEFSV get_sv("_", GV_ADD)

/* Length-taking stash lookup (the guest op copies exactly slen bytes). */
#define gv_stashpvn(name, len, flags)                                     \
    ((HV *)GOPERL_SV(goperl_ops(GOPERL_OP_GV_STASHPV,                     \
                                (uint64_t)(int64_t)(I32)(flags), 0,       \
                                (name), (uint64_t)(len))))
#define gv_stashpvs(name, flags) \
    gv_stashpvn("" name "", sizeof(name) - 1, (flags))

/* Method resolution with ISA walk (returns the method's GV, or NULL). */
#define gv_fetchmethod(stash, name)                                       \
    ((GV *)GOPERL_SV(goperl_ops(GOPERL_OP_GV_FETCHMETHOD,                 \
                                GOPERL_TOK(stash), 0, (name),             \
                                strlen(name))))
#define gv_fetchmethod_autoload(stash, name, autoload)                    \
    ((GV *)GOPERL_SV(goperl_ops(GOPERL_OP_GV_FETCHMETHOD,                 \
                                GOPERL_TOK(stash), (autoload) ? 1 : 0,    \
                                (name), strlen(name))))

/* sv_reftype: the "HASH"/"ARRAY"/"Class::Name" string. Interned host
 * copies, stable for the process like every SDK-returned name. */
static const char *goperl_sv_reftype(goperl_frame_t *f, SV *sv, int ob) {
    if (ob) {
        uint64_t stash = goperl_do_op(f, GOPERL_OP_SV_STASH, GOPERL_TOK(sv),
                                      0, 0, 0);
        if (stash) {
            const char *n = goperl_intern_packed(
                f, goperl_do_op(f, GOPERL_OP_HV_NAME, stash, 0, 0, 0));
            if (n && *n) return n;
            return "__ANON__";
        }
    }
    switch ((int)goperl_do_op(f, GOPERL_OP_SV_TYPE, GOPERL_TOK(sv), 0, 0, 0)) {
    case SVt_PVAV: return "ARRAY";
    case SVt_PVHV: return "HASH";
    case SVt_PVCV: return "CODE";
    case SVt_PVGV: return "GLOB";
    case SVt_PVIO: return "IO";
    case SVt_REGEXP: return "Regexp";
    case SVt_PVFM: return "FORMAT";
    default:
        return goperl_op0(GOPERL_OP_SV_RV, GOPERL_TOK(sv), 0) ? "REF"
                                                              : "SCALAR";
    }
}
#define sv_reftype(sv, ob) \
    ((char *)goperl_sv_reftype(_gof, (SV *)(sv), (ob)))

/* libm passthroughs (real perl defines these to the C library too). */
#define Perl_floor(x) floor(x)
#define Perl_ceil(x) ceil(x)
#define Perl_fmod(x, y) fmod((x), (y))
#define Perl_pow(x, y) pow((x), (y))
#define Perl_modf(x, y) modf((x), (y))
#define Perl_strtod(s, e) strtod((s), (e))

/* MY_CXT: per-interpreter module state. One interpreter per process here,
 * so it collapses to one static struct. pTHX_ is NON-empty in this SDK, so
 * pMY_CXT must be a real (ignored) parameter for `(pTHX_ pMY_CXT)`
 * signatures to stay valid C. */
#define PERL_GET_CONTEXT goperl_cur_frame_v
#define START_MY_CXT static my_cxt_t goperl_my_cxt_v;
#define MY_CXT goperl_my_cxt_v
#define dMY_CXT \
    struct goperl_dmycxt_s *goperl_dmycxt_unused __attribute__((unused)) = 0
#define dMY_CXT_SV dMY_CXT
#define MY_CXT_INIT \
    memset(&goperl_my_cxt_v, 0, sizeof goperl_my_cxt_v)
#define MY_CXT_CLONE MY_CXT_INIT
#define pMY_CXT my_cxt_t *goperl_mycxt_unused __attribute__((unused))
#define pMY_CXT_ pMY_CXT,
#define _pMY_CXT , pMY_CXT
#define aMY_CXT (&goperl_my_cxt_v)
#define aMY_CXT_ aMY_CXT,
#define _aMY_CXT , aMY_CXT
#define MY_CXT_KEY "goperl::_guts" XS_VERSION

/* HePV is defined with the HE-shadow machinery below. */

/* Legacy arena-walk compat: pre-5.18 perls tracked overload flags per-RV,
 * and old XS (Moose's ToInstance) walks PL_sv_arenaroot to fix up every
 * reference after toggling. On modern perl the flag lives on the STASH, so
 * one SvAMAGIC_on covers all references and the walk is vestigial — a NULL
 * arena root makes it compile to a no-op loop, which is the CORRECT result
 * here, not an approximation. */

/* ---- v6 groundwork: the Cpanel::JSON::XS / HTML::Parser surface -------- */

#ifndef NV_MAX
#define NV_MAX DBL_MAX
#define NV_MIN DBL_MIN
#define NV_DIG DBL_DIG
#endif
/* The GUEST's integer geometry: wasm32 perl builds with 32-bit IV/UV
 * (ivsize=4). The SDK's IV/UV C types stay 64-bit for token plumbing, but
 * every SEMANTIC decision a dist makes — "does this number fit an IV?",
 * IVSIZE branches — must follow the guest, or values silently truncate
 * when they cross (Cpanel::JSON::XS decoding 2^32+ integers). */
#ifndef UV_MAX
#define IVSIZE 4
#define UVSIZE 4
#define UV_MAX ((UV)UINT32_MAX)
#define UV_MIN 0
#define IV_MAX ((IV)INT32_MAX)
#define IV_MIN ((IV)INT32_MIN)
#endif
#define Gconvert(x, n, t, b) snprintf((b), 64, "%.*g", (int)(n), (double)(x))
#define PTR2ul(p) ((unsigned long)(uintptr_t)(p))
#define safecalloc(n, sz) calloc((size_t)(n), (size_t)(sz))
#define SVfARG(sv) ((SV *)(sv))
#define GOPERL_INFO_IOKp 131072
#define GOPERL_INFO_NOKp 262144
#define SvIOKp(sv) ((SvFLAGS(sv) & GOPERL_INFO_IOKp) != 0)
#define SvNOKp(sv) ((SvFLAGS(sv) & GOPERL_INFO_NOKp) != 0)
#define SvIsBOOL(sv) \
    ((int)goperl_op0(GOPERL_OP_SV_IS_BOOL, GOPERL_TOK(sv), 0))
/* nomg variants: the SDK's accessors go through the guest's public API,
 * which applies get-magic — acceptable for these dists' uses. */
#define SvIV_nomg(sv) SvIV(sv)
#define SvUV_nomg(sv) SvUV(sv)
#define SvNV_nomg(sv) SvNV(sv)
#define SvNVX(sv) SvNV(sv)
#define SvPOK_only(sv) \
    ((void)goperl_op0(GOPERL_OP_SV_POK_ONLY, GOPERL_TOK(sv), 0))
#define SvPOK_only_UTF8(sv) (SvPOK_only(sv), SvUTF8_on(sv))
#define SvPVutf8_force(sv, lenvar) \
    (sv_utf8_upgrade((SV *)(sv)), SvPV_force((sv), lenvar))
#define SvPVutf8(sv, lenvar) (sv_utf8_upgrade((SV *)(sv)), SvPV((sv), lenvar))
#define SvSetMagicSV(dst, src) (sv_setsv((dst), (src)), (void)mg_set(dst))
#define SvSetSV(dst, src) sv_setsv((dst), (src))
/* Taint support is off in the embedded build (like -DNO_TAINT_SUPPORT). */
#define SvTAINTED(sv) 0
#define SvTAINT(sv) ((void)0)
#define SvTAINTED_on(sv) ((void)0)
#define TAINT_IF(x) ((void)(x))
#define TAINT_NOT ((void)0)
#define SvTRUEx(sv) SvTRUE(sv)
#define SvTRUE_nomg(sv) SvTRUE(sv)
#define HvAMAGIC(hv) ((int)goperl_op0(GOPERL_OP_GV_AMG, GOPERL_TOK(hv), 0))

/* hv_common: only the JUST_SV fetch/store shapes dists actually use. */
#define HV_FETCH_ISSTORE 0x04
#define HV_FETCH_ISEXISTS 0x08
#define HV_FETCH_LVALUE 0x10
#define HV_FETCH_JUST_SV 0x20
#define HV_DELETE 0x40
static void *goperl_hv_common(goperl_frame_t *f, HV *hv, SV *keysv,
                              const char *key, STRLEN klen, int kflags,
                              int action, SV *val, U32 hash) {
    (void)hash;
    char kbuf[8];
    if (keysv && !key) {
        uint64_t l = 0;
        key = goperl_api_v->sv_pv(f, GOPERL_TOK(keysv), &l);
        klen = (STRLEN)l;
        (void)kbuf;
    }
    if (action & HV_FETCH_ISSTORE)
        return (void *)GOPERL_SV(goperl_do_op(
            f, GOPERL_OP_HV_STORE_KLEN, GOPERL_TOK(hv),
            ((uint64_t)((kflags & HVhek_UTF8) ? 1 : 0) << 32) |
                (uint32_t)GOPERL_TOK(val),
            key, (uint64_t)klen));
    return (void *)GOPERL_SV(goperl_do_op(
        f, GOPERL_OP_HV_FETCH_KLEN, GOPERL_TOK(hv),
        ((uint64_t)((kflags & HVhek_UTF8) ? 1 : 0) << 32) |
            (uint32_t)((action & HV_FETCH_LVALUE) ? 1 : 0),
        key, (uint64_t)klen));
}
#define hv_common(hv, keysv, key, klen, kflags, action, val, hash)        \
    goperl_hv_common(_gof, (HV *)(hv), (SV *)(keysv), (key),              \
                     (STRLEN)(klen), (int)(kflags), (int)(action),        \
                     (SV *)(val), (U32)(hash))

/* xsubpp emits this for ATTRS: sections — apply attribute strings to a
 * freshly bound CV by driving the guest's attributes::import. */
static void apply_attrs_string(const char *stashpv, CV *cv,
                               const char *attrstr, STRLEN len) {
    goperl_frame_t *f = goperl_cur_frame_v;
    goperl_do_op(f, GOPERL_OP_EVAL_PV, 1, 0, "require attributes",
                 sizeof("require attributes") - 1);
    uint32_t toks[16];
    int n = 0;
    size_t pl = strlen(stashpv);
    uint64_t pkgsv = goperl_do_op(f, GOPERL_OP_NEW_PVN, 0, pl, stashpv, pl);
    goperl_do_op(f, GOPERL_OP_SV_MORTAL, pkgsv, 0, 0, 0);
    toks[n++] = (uint32_t)pkgsv;
    uint64_t rv =
        goperl_do_op(f, GOPERL_OP_NEW_RV_INC, GOPERL_TOK(cv), 0, 0, 0);
    goperl_do_op(f, GOPERL_OP_SV_MORTAL, rv, 0, 0, 0);
    toks[n++] = (uint32_t)rv;
    const char *p = attrstr, *e = attrstr + (len ? len : strlen(attrstr));
    while (p < e && n < 15) {
        while (p < e && *p == ' ') p++;
        const char *q = p;
        while (q < e && *q != ' ') q++;
        if (q > p) {
            uint64_t s2 = goperl_do_op(f, GOPERL_OP_NEW_PVN, 0,
                                       (uint64_t)(q - p), p, (uint64_t)(q - p));
            goperl_do_op(f, GOPERL_OP_SV_MORTAL, s2, 0, 0, 0);
            toks[n++] = (uint32_t)s2;
        }
        p = q;
    }
    uint64_t cvtok = goperl_do_op(f, GOPERL_OP_GET_CV, 0, 0,
                                  "attributes::import",
                                  sizeof("attributes::import") - 1);
    if (!cvtok) return;
    char buf[sizeof toks];
    memcpy(buf, toks, (size_t)n * 4);
    goperl_do_op(f, GOPERL_OP_CALL_SV,
                 ((uint64_t)(uint32_t)(G_VOID | G_DISCARD) << 32) |
                     (uint32_t)cvtok,
                 (uint64_t)(n * 4), buf, (uint64_t)(n * 4));
}

/* warnings-state stubs: cop_warnings is modeled as an opaque slot dists
 * assign pWARN_* into around host-local code; the guest's real warnings
 * state is reached through ckWARN/Perl_warner instead. */
#define pWARN_NONE ((char *)0)
#define pWARN_STD ((char *)0)
#define pWARN_ALL ((char *)1)
#define WARN_UTF8 44 /* perl 5.42 warnings.h */
#ifndef I32_MAX
#define I32_MAX INT32_MAX
#define I32_MIN INT32_MIN
#define I16_MAX INT16_MAX
#define U32_MAX UINT32_MAX
#endif
#define MUTABLE_PTR(p) ((void *)(p))
#define MUTABLE_SV(p) ((SV *)(p))
#define MUTABLE_AV(p) ((AV *)(p))
#define MUTABLE_HV(p) ((HV *)(p))
#define MUTABLE_CV(p) ((CV *)(p))
#define MUTABLE_GV(p) ((GV *)(p))
#define CvNODEBUG_on(cv) ((void)(cv))
#define PERL_LOADMOD_NOIMPORT 0x1
#define GV_NO_SVGMAGIC 0
#define get_cvs(name, flags) get_cv("" name "", (flags))
#define UNI_DISPLAY_ISPRINT 0x1
#define UNI_DISPLAY_BACKSLASH 0x2
#define UNI_DISPLAY_QQ (UNI_DISPLAY_ISPRINT | UNI_DISPLAY_BACKSLASH)
/* pv_uni_display: escape a UTF-8 buffer for an error message into dsv. */
static char *goperl_pv_uni_display(goperl_frame_t *f, SV *dsv, const U8 *spv,
                                   STRLEN len, STRLEN pvlim, U32 flags) {
    (void)flags;
    char out[1024];
    size_t o = 0;
    STRLEN shown = 0;
    for (STRLEN i = 0; i < len && o < sizeof out - 8; i++) {
        U8 c = spv[i];
        if (shown++ >= pvlim) {
            memcpy(out + o, "...", 3);
            o += 3;
            break;
        }
        if (c == '\\' || c == '"') {
            out[o++] = '\\';
            out[o++] = (char)c;
        } else if (c >= 0x20 && c < 0x7f) {
            out[o++] = (char)c;
        } else {
            o += (size_t)snprintf(out + o, sizeof out - o, "\\x{%02x}", c);
        }
    }
    goperl_do_op(f, GOPERL_OP_SV_SETPVN, GOPERL_TOK(dsv), (uint64_t)o, out,
                 (uint64_t)o);
    return (char *)goperl_sv_pv_(f, dsv, 0);
}
#define pv_uni_display(dsv, spv, len, pvlim, flags)                     \
    goperl_pv_uni_display(_gof, (SV *)(dsv), (const U8 *)(spv),         \
                          (STRLEN)(len), (STRLEN)(pvlim), (U32)(flags))

/* ---- the DBI surface ---------------------------------------------------- */

#define CopFILEGV(cop) ((GV *)0) /* file GV not modeled; CopFILE has the name */
#define DEFSV_set(sv) sv_setsv(DEFSV, (SV *)(sv))
#define die Perl_croak_nocontext
#define perl_require_pv(pv)                                              \
    do {                                                                 \
        char goperl_req_buf[512];                                        \
        snprintf(goperl_req_buf, sizeof goperl_req_buf, "require %s",    \
                 (pv));                                                  \
        goperl_do_op(_gof, GOPERL_OP_EVAL_PV, 0, 0, goperl_req_buf,      \
                     strlen(goperl_req_buf));                            \
    } while (0)
/* Host-local save of plain C variables: restored at XSUB return via the
 * host save-stack (same mechanism as SAVEVPTR). */
#define SAVEI32(i) goperl_hostsave_i32(_gof, (int32_t *)&(i))
#define SAVEINT(i) goperl_hostsave_i32(_gof, (int32_t *)&(i))
#define SAVEI8(i) SAVEI32(i)
#define SAVEBOOL(i) SAVEI32(i)
#define SvOBJECT_off(sv) ((void)(sv)) /* flag flips w/o value change */
#define SvOOK_off(sv) ((void)(sv))
#define SvPOK_off(sv) sv_setsv((SV *)(sv), &PL_sv_undef)
#define SvRMAGICAL_on(sv) ((void)(sv))
#define SvSMAGICAL(sv) SvRMAGICAL(sv)
#define DEBUG_l(x) ((void)0)
#define DEBUG_x(x) ((void)0)
#define DEBUG_v_TEST 0
#define aTHXo aTHX
#define aTHXo_ aTHX_
#define pTHXo pTHX
#define pTHXo_ pTHX_
#define MSPAGAIN SPAGAIN
typedef void (*XSUBADDR_t)(pTHX_ CV *cv);
/* localize $_ guest-side (save_scalar on *main::_) */
static void goperl_save_defsv(goperl_frame_t *f) {
    uint64_t gv = goperl_do_op(f, GOPERL_OP_GV_FETCH, GV_ADD,
                               (uint64_t)SVt_PV, "main::_",
                               sizeof("main::_") - 1);
    if (gv) goperl_do_op(f, GOPERL_OP_SAVE_SCALAR, gv, 0, 0, 0);
}
#define SAVE_DEFSV goperl_save_defsv(_gof)
#define SvIOK_UV(sv) \
    ((SvFLAGS(sv) & (SVf_IOK | GOPERL_INFO_ISUV)) == \
     (SVf_IOK | GOPERL_INFO_ISUV))
#define SvPVutf8_nolen(sv) (sv_utf8_upgrade((SV *)(sv)), SvPV_nolen(sv))
#define XST_mIV(i, v) (ST(i) = sv_2mortal(newSViv((IV)(v))))
#define XST_mNV(i, v) (ST(i) = sv_2mortal(newSVnv((NV)(v))))
#define XST_mPV(i, v) (ST(i) = sv_2mortal(newSVpv((v), 0)))
#define XST_mUNDEF(i) (ST(i) = &PL_sv_undef)
#define XST_mNO(i) (ST(i) = &PL_sv_no)
#define XST_mYES(i) (ST(i) = &PL_sv_yes)
/* av_fill(-1) IS av_clear in core; av_undef/hv_undef additionally drop the
 * container's identity, which nothing host-side can observe — emptying is
 * the faithful effect here. */
#define av_clear(av) av_fill((av), -1)
#define av_undef(av) av_fill((av), -1)
#define hv_undef(hv) ((void)goperl_op0(GOPERL_OP_HV_CLEAR, GOPERL_TOK(hv), 0))
static AV *goperl_get_av(goperl_frame_t *f, const char *name, I32 flags) {
    uint64_t gv = goperl_do_op(f, GOPERL_OP_GV_FETCH, (uint64_t)(int64_t)flags,
                               (uint64_t)SVt_PVAV, name, strlen(name));
    if (!gv) return 0;
    return (AV *)GOPERL_SV(
        goperl_do_op(f, GOPERL_OP_GV_PTR, gv, 3 /* GvAV */, 0, 0));
}
#define get_av(name, flags) goperl_get_av(_gof, (name), (I32)(flags))
#define Nullfp ((PerlIO *)0)
#define Nullsv ((SV *)0)
#define Nullhv ((HV *)0)
#define Nullav ((AV *)0)
#define Nullcv ((CV *)0)
#define Nullgv ((GV *)0)
#define Nullch ((char *)0)
#define PERL_GET_THX goperl_cur_frame_v
#define dTHR dNOOP
#define TAINT ((void)0)
#define TAINT_ENV() ((void)0)
#define TAINT_PROPER(s) ((void)0)
#define I_STDARG 1
#define POPi ((IV)SvIV(POPs))
#define POPu ((UV)SvUV(POPs))
#define POPn ((NV)SvNV(POPs))
#define POPl POPi
#define POPpx SvPV_nolen(POPs)
#define SvGMAGICAL(sv) SvRMAGICAL(sv)
/* Writable taint state (DBI assigns PL_tainted around trace calls); the
 * embedded build has no taint support, so the value is inert. */
__attribute__((weak)) int goperl_pl_tainted_v = 0;
__attribute__((weak)) int goperl_pl_tainting_v = 0;
__attribute__((weak)) int goperl_pl_dirty_v = 0;
#undef PL_tainted
#undef PL_tainting
#undef PL_dirty
#define PL_tainted goperl_pl_tainted_v
#define PL_tainting goperl_pl_tainting_v
#define PL_dirty goperl_pl_dirty_v
/* Assignable no-op: teardown depth is the embedder's business. */
__attribute__((weak)) int goperl_pl_destruct_level_v = 0;
#define PL_perl_destruct_level goperl_pl_destruct_level_v
#define perl_get_sv(name, flags) get_sv((name), (flags))
#define perl_get_av(name, flags) get_av((name), (flags))
#define perl_get_hv(name, flags) get_hv((name), (flags))
#define perl_get_cv(name, flags) get_cv((name), (flags))


#define PL_ptr_table ((void *)0)
#define PL_sub_generation \
    ((U32)goperl_op0(GOPERL_OP_PLVAR_GET, GOPERL_PL_SUB_GENERATION, 0))
#define SVt_PVBM SVt_PVMG /* pre-5.10 name */
#define av_shift(av) \
    GOPERL_SV(goperl_op0(GOPERL_OP_AV_SHIFT, GOPERL_TOK(av), 0))
/* Object exceptions degrade to their string form across the SDK boundary
 * (croak carries a message, not an SV). */
static void goperl_croak_sv(goperl_frame_t *f, SV *sv)
    __attribute__((noreturn, unused));
static void goperl_croak_sv(goperl_frame_t *f, SV *sv) {
    uint64_t len = 0;
    const char *p = goperl_api_v->sv_pv(f, GOPERL_TOK(sv), &len);
    goperl_croakf(f, "%.*s", (int)(len > 400 ? 400 : len), p ? p : "");
}
#define croak_sv(sv) goperl_croak_sv(_gof, (SV *)(sv))
#define warn_sv(sv)                                       \
    do {                                                  \
        STRLEN goperl_wl;                                 \
        const char *goperl_wp = SvPV((SV *)(sv), goperl_wl); \
        Perl_warn(aTHX_ "%.*s", (int)goperl_wl, goperl_wp);  \
    } while (0)
static void goperl_gv_efullname(goperl_frame_t *f, SV *sv, GV *gv) {
    uint64_t stash = goperl_do_op(f, GOPERL_OP_GV_PTR, GOPERL_TOK(gv), 0, 0, 0);
    const char *pkg = stash ? goperl_intern_packed(
                                  f, goperl_do_op(f, GOPERL_OP_HV_NAME, stash,
                                                  0, 0, 0))
                            : 0;
    const char *name = goperl_intern_packed(
        f, goperl_do_op(f, GOPERL_OP_GV_NAME, GOPERL_TOK(gv), 0, 0, 0));
    char buf[512];
    snprintf(buf, sizeof buf, "%s::%s", pkg && *pkg ? pkg : "main",
             name ? name : "?");
    goperl_do_op(f, GOPERL_OP_SV_SETPVN, GOPERL_TOK(sv), strlen(buf), buf,
                 strlen(buf));
}
#define gv_efullname(sv, gv) goperl_gv_efullname(_gof, (SV *)(sv), (GV *)(gv))
#define gv_efullname3(sv, gv, prefix) gv_efullname((sv), (gv))
#define gv_efullname4(sv, gv, prefix, keepmain) gv_efullname((sv), (gv))
static char *goperl_hv_iterkey(goperl_frame_t *f, HE *he, I32 *retlen) {
    HEK *hek = goperl_he_hek(f, he);
    *retlen = hek->hek_len;
    return hek->hek_key;
}
#define hv_iterkey(he, retlen) goperl_hv_iterkey(_gof, (HE *)(he), (retlen))
#define sv_2nv(sv) SvNV(sv)
#define sv_2uv(sv) SvUV(sv)
#define sv_2pv(sv, lenp) ((char *)goperl_sv_pv_(_gof, (SV *)(sv), (uint64_t *)(lenp)))
#define sv_2io(sv) \
    ((IO *)GOPERL_SV(goperl_op0(GOPERL_OP_SV_2IO, GOPERL_TOK(sv), 0)))
#define sv_force_normal(sv) \
    ((void)goperl_op0(GOPERL_OP_SV_FORCE_NORMAL, GOPERL_TOK(sv), 0))
#define sv_force_normal_flags(sv, flags) sv_force_normal(sv)

/* PerlIO HANDLES (not layers). The SDK's PerlIO is a host FILE* (stderr/
 * stdout above); a GUEST handle (sv_2io -> IoOFP, PerlIO_open) is wrapped
 * into a host FILE* with funopen/fopencookie whose write/close forward to
 * guest ops — so PerlIO_printf and friends work uniformly on both. */
static int goperl_iowrap_write_(void *tokp, const char *buf, int n);
static int goperl_iowrap_close_(void *tokp);
#if defined(__APPLE__)
static int goperl_iowrap_write_cb(void *c, const char *b, int n) {
    return goperl_iowrap_write_(c, b, n);
}
static int goperl_iowrap_close_cb(void *c) { return goperl_iowrap_close_(c); }
#endif
static int goperl_iowrap_write_(void *tokp, const char *buf, int n) {
    uint64_t tok = *(uint64_t *)tokp;
    return (int)(int64_t)goperl_do_op(goperl_cur_frame_v,
                                      GOPERL_OP_PERLIO_WRITE, tok,
                                      (uint64_t)n, buf, (uint64_t)n);
}
static int goperl_iowrap_close_(void *tokp) {
    uint64_t tok = *(uint64_t *)tokp;
    int r = (int)(int64_t)goperl_do_op(goperl_cur_frame_v,
                                       GOPERL_OP_PERLIO_CLOSE, tok, 0, 0, 0);
    free(tokp);
    return r;
}
static FILE *goperl_perlio_wrap(uint64_t tok) {
    if (!tok) return 0;
    uint64_t *cookie = (uint64_t *)malloc(sizeof(uint64_t));
    *cookie = tok;
#if defined(__APPLE__)
    FILE *fp = funopen(cookie, 0, goperl_iowrap_write_cb, 0,
                       goperl_iowrap_close_cb);
#else
    cookie_io_functions_t io = {0};
    io.write = (cookie_write_function_t *)goperl_iowrap_write_;
    io.close = (cookie_close_function_t *)goperl_iowrap_close_;
    FILE *fp = fopencookie(cookie, "w", io);
#endif
    if (fp) setvbuf(fp, 0, _IONBF, 0);
    return fp;
}
#define IoOFP(io)                                              \
    goperl_perlio_wrap(                                        \
        goperl_op0(GOPERL_OP_IO_OFP, GOPERL_TOK(io), 0))
#define IoIFP(io) IoOFP(io)
static PerlIO *goperl_perlio_open(goperl_frame_t *f, const char *path,
                                  const char *mode) {
    char buf[1024];
    size_t ml = strlen(mode), pl = strlen(path);
    if (ml + pl + 2 > sizeof buf) return 0;
    memcpy(buf, mode, ml + 1);
    memcpy(buf + ml + 1, path, pl + 1);
    return goperl_perlio_wrap(goperl_do_op(f, GOPERL_OP_PERLIO_OPEN, 0, 0,
                                           buf, ml + 1 + pl + 1));
}
#define PerlIO_open(path, mode) goperl_perlio_open(_gof, (path), (mode))
#define PerlIO_close(io) fclose(io)
#define PerlIO_flush(io) fflush(io)
#define PerlIO_write(io, buf, n) \
    ((SSize_t)fwrite((buf), 1, (size_t)(n), (io)))
#define PerlIO_puts(io, s2) fputs((s2), (io))
#define PerlIO_setlinebuf(io) ((void)0)
static void goperl_perlio_vprintf_v(FILE *fp, const char *fmt, va_list ap)
    __attribute__((unused));
static void goperl_perlio_vprintf_v(FILE *fp, const char *fmt, va_list ap) {
    char buf[2048];
    goperl_vfmt(goperl_cur_frame_v, buf, sizeof buf, fmt, ap);
    fputs(buf, fp);
}
#define PerlIO_vprintf(io, fmt, ap) (goperl_perlio_vprintf_v((io), (fmt), (ap)), 0)

/* eval_sv: results delivered like call_sv (the guest packs them into a
 * mortal AV; the SDK pushes them onto the host stack model). */
static I32 goperl_eval_sv(goperl_frame_t *f, SV *sv, I32 flags) {
    uint64_t packed =
        goperl_do_op(f, GOPERL_OP_EVAL_SV,
                     ((uint64_t)(uint32_t)flags << 32) |
                         (uint32_t)GOPERL_TOK(sv),
                     0, 0, 0);
    int died = (int)(packed >> 63);
    I32 count = (I32)((packed >> 32) & 0x7FFFFFFF);
    uint64_t avtok = (uint32_t)packed;
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
#define eval_sv(sv, flags) goperl_eval_sv(_gof, (SV *)(sv), (I32)(flags))


/* ASCII ctype (real perl's are ASCII-range too for these dists' uses). */
#ifndef toLOWER
#define toLOWER(c) ((c) >= 'A' && (c) <= 'Z' ? (c) + 32 : (c))
#define toUPPER(c) ((c) >= 'a' && (c) <= 'z' ? (c)-32 : (c))
#endif
#ifndef isALPHA
#define isALPHA(c) \
    (((c) >= 'a' && (c) <= 'z') || ((c) >= 'A' && (c) <= 'Z'))
#define isDIGIT(c) ((c) >= '0' && (c) <= '9')
#define isALNUM(c) (isALPHA(c) || isDIGIT(c) || (c) == '_')
#define isSPACE(c) \
    ((c) == ' ' || (c) == '\t' || (c) == '\n' || (c) == '\r' || \
     (c) == '\f' || (c) == '\v')
#define isWORDCHAR(c) isALNUM(c)
#define isPRINT(c) ((c) >= 0x20 && (c) < 0x7f)
#define isGRAPH(c) ((c) > 0x20 && (c) < 0x7f)
#define isUPPER(c) ((c) >= 'A' && (c) <= 'Z')
#define isLOWER(c) ((c) >= 'a' && (c) <= 'z')
#define isPUNCT(c) (isGRAPH(c) && !isALNUM(c))
#define isXDIGIT(c) \
    (isDIGIT(c) || ((c) >= 'a' && (c) <= 'f') || ((c) >= 'A' && (c) <= 'F'))
#define isCNTRL(c) (((U8)(c)) < 0x20 || (c) == 0x7f)
#define isBLANK(c) ((c) == ' ' || (c) == '\t')
#define isPSXSPC(c) isSPACE(c)
#endif

/* perl's hex digit table: low nibble lowercase, +16 for uppercase. */
static const char PL_hexdigit[] = "0123456789abcdef0123456789ABCDEF";

/* UTF-8 toolkit — self-contained byte math on host-visible buffers. */
#define UTF8_MAXLEN 13
#define UTF8_MAXBYTES UTF8_MAXLEN
#define UTF8_IS_START(c) (((U8)(c)) >= 0xc2)
#define UTF8_IS_CONTINUATION(c) ((((U8)(c)) & 0xc0) == 0x80)
#define UTF8_CHECK_ONLY 0x1
#define UTF8_DISALLOW_SUPER 0x2
#define UTF8_ALLOW_ANY 0x4
#define UTF8_DISALLOW_SURROGATE 0x8
#define UTF8_WARN_SURROGATE 0
#define UTF8_WARN_SUPER 0
#define UTF8_DISALLOW_ILLEGAL_INTERCHANGE \
    (UTF8_DISALLOW_SURROGATE | UTF8_DISALLOW_SUPER)
static const U8 goperl_utf8skip_v[16] = {1, 1, 1, 1, 1, 1, 1, 1,
                                         1, 1, 1, 1, 2, 2, 3, 4};
#define UTF8SKIP(s) ((STRLEN)goperl_utf8skip_v[((U8) * (const U8 *)(s)) >> 4])
static STRLEN goperl_utf8_length(const U8 *s, const U8 *e) {
    STRLEN n = 0;
    while (s < e) {
        s += UTF8SKIP(s);
        n++;
    }
    return n;
}
#define utf8_length(s, e) goperl_utf8_length((const U8 *)(s), (const U8 *)(e))
#define utf8_distance(a, b) \
    ((IV)goperl_utf8_length((const U8 *)(b), (const U8 *)(a)))
static U8 *goperl_utf8_hop(const U8 *s, SSize_t off) {
    while (off-- > 0) s += UTF8SKIP(s);
    return (U8 *)s;
}
#define utf8_hop(s, off) goperl_utf8_hop((const U8 *)(s), (off))
static UV goperl_utf8n_to_uvchr(const U8 *s, STRLEN curlen, STRLEN *retlen,
                                U32 flags) {
    STRLEN n = curlen ? UTF8SKIP(s) : 0;
    if (n == 0 || n > curlen) goto fail;
    if (n == 1) {
        if (retlen) *retlen = 1;
        return s[0];
    }
    {
        UV uv = s[0] & (0x7f >> n);
        for (STRLEN i = 1; i < n; i++) {
            if (!UTF8_IS_CONTINUATION(s[i])) goto fail;
            uv = (uv << 6) | (s[i] & 0x3f);
        }
        /* non-minimal encodings are malformed, as in the real decoder */
        if ((n == 2 && uv < 0x80) || (n == 3 && uv < 0x800) ||
            (n == 4 && uv < 0x10000))
            goto fail;
        if ((flags & UTF8_DISALLOW_SURROGATE) && uv >= 0xD800 && uv <= 0xDFFF)
            goto fail;
        if ((flags & UTF8_DISALLOW_SUPER) && uv > 0x10FFFF)
            goto fail;
        if (retlen) *retlen = n;
        return uv;
    }
fail:
    if (retlen) *retlen = (STRLEN)-1;
    return 0;
}
#define utf8n_to_uvchr(s, len, retlen, flags) \
    goperl_utf8n_to_uvchr((const U8 *)(s), (len), (retlen), (flags))
#define utf8_to_uvchr_buf(s, e, retlen) \
    utf8n_to_uvchr((s), (STRLEN)((const U8 *)(e) - (const U8 *)(s)), (retlen), 0)
static U8 *goperl_uvchr_to_utf8(U8 *d, UV uv) {
    if (uv < 0x80) {
        *d++ = (U8)uv;
    } else if (uv < 0x800) {
        *d++ = (U8)(0xc0 | (uv >> 6));
        *d++ = (U8)(0x80 | (uv & 0x3f));
    } else if (uv < 0x10000) {
        *d++ = (U8)(0xe0 | (uv >> 12));
        *d++ = (U8)(0x80 | ((uv >> 6) & 0x3f));
        *d++ = (U8)(0x80 | (uv & 0x3f));
    } else {
        *d++ = (U8)(0xf0 | (uv >> 18));
        *d++ = (U8)(0x80 | ((uv >> 12) & 0x3f));
        *d++ = (U8)(0x80 | ((uv >> 6) & 0x3f));
        *d++ = (U8)(0x80 | (uv & 0x3f));
    }
    return d;
}
#define uvchr_to_utf8(d, uv) goperl_uvchr_to_utf8((U8 *)(d), (UV)(uv))
#define uvchr_to_utf8_flags(d, uv, flags) uvchr_to_utf8((d), (uv))
#define uvuni_to_utf8(d, uv) uvchr_to_utf8((d), (uv))
#define uvuni_to_utf8_flags(d, uv, flags) uvchr_to_utf8((d), (uv))
static U8 *goperl_bytes_to_utf8(const U8 *s, STRLEN *lenp) {
    STRLEN len = *lenp, out = 0;
    for (STRLEN i = 0; i < len; i++) out += s[i] < 0x80 ? 1 : 2;
    U8 *d = (U8 *)malloc(out + 1), *p = d;
    for (STRLEN i = 0; i < len; i++) p = goperl_uvchr_to_utf8(p, s[i]);
    *p = 0;
    *lenp = out;
    return d;
}
#define bytes_to_utf8(s, lenp) goperl_bytes_to_utf8((const U8 *)(s), (lenp))
static U8 *goperl_utf8_to_bytes(U8 *s, STRLEN *lenp) {
    U8 *d = s, *p = s, *e = s + *lenp;
    while (p < e) {
        STRLEN rl;
        UV uv = goperl_utf8n_to_uvchr(p, (STRLEN)(e - p), &rl, 0);
        if (rl == (STRLEN)-1 || uv > 0xff) return 0;
        *d++ = (U8)uv;
        p += rl;
    }
    *lenp = (STRLEN)(d - s);
    return s;
}
#define utf8_to_bytes(s, lenp) goperl_utf8_to_bytes((U8 *)(s), (lenp))

/* Guest-backed SV/HV/AV additions. */
#define goperl_svlen_slot_v (GOPERL_SHARED->svlen_slot)
static STRLEN *goperl_svlen_lv(goperl_frame_t *f, SV *sv) {
    goperl_svlen_slot_v =
        (STRLEN)goperl_do_op(f, GOPERL_OP_SV_LEN_BUF, GOPERL_TOK(sv), 0, 0, 0);
    return &goperl_svlen_slot_v;
}
#define SvLEN(sv) (*goperl_svlen_lv(_gof, (SV *)(sv)))
#define SvREADONLY_off(sv) \
    ((void)goperl_op0(GOPERL_OP_SV_READONLY_OFF, GOPERL_TOK(sv), 0))
static char *goperl_sv_pv_force(goperl_frame_t *f, SV *sv, STRLEN *lenp) {
    uint64_t r = goperl_do_op(f, GOPERL_OP_SV_PV_FORCE, GOPERL_TOK(sv), 0, 0, 0);
    if (lenp) *lenp = (STRLEN)(uint32_t)r;
    return (char *)goperl_api_v->guest_mem(f, r >> 32);
}
#define SvPV_force(sv, lenvar) goperl_sv_pv_force(_gof, (SV *)(sv), &(lenvar))
#define SvPV_force_nomg(sv, lenvar) SvPV_force(sv, lenvar)
#define sv_chop(sv, p) \
    ((void)goperl_op0(GOPERL_OP_SV_CHOP, GOPERL_TOK(sv), \
                      (uint64_t)((const char *)(p)-SvPVX_const(sv))))
#define sv_insert(big, off, len, little, littlelen)                      \
    ((void)goperl_ops(GOPERL_OP_SV_INSERT, GOPERL_TOK(big),              \
                      ((uint64_t)(off) << 32) | (uint32_t)(len),         \
                      (little), (uint64_t)(littlelen)))
#define sv_utf8_decode(sv) \
    ((bool)goperl_op0(GOPERL_OP_SV_UTF8_DECODE, GOPERL_TOK(sv), 0))
#define sv_utf8_downgrade(sv, fail_ok) \
    ((bool)goperl_op0(GOPERL_OP_SV_UTF8_DOWNGRADE, GOPERL_TOK(sv), \
                      (fail_ok) ? 1 : 0))
#define mg_set(sv) ((void)goperl_op0(GOPERL_OP_MG_SET, GOPERL_TOK(sv), 0), 0)
#define SvSETMAGIC(sv) ((void)mg_set(sv))
#define sv_setiv_mg(sv, i) (sv_setiv((sv), (i)), (void)mg_set(sv))
#define sv_setnv_mg(sv, n) (sv_setnv((sv), (n)), (void)mg_set(sv))
#define sv_setpvn_mg(sv, p, n) (sv_setpvn((sv), (p), (n)), (void)mg_set(sv))
#define sv_setsv_mg_2(sv, s2) (sv_setsv((sv), (s2)), (void)mg_set(sv))
#define GIMME_V ((U8)goperl_op0(GOPERL_OP_GIMME_V, 0, 0))
#define GIMME GIMME_V
#define ckWARN(w) ((int)goperl_op0(GOPERL_OP_CKWARN, (uint64_t)(w), 0))
#define ckWARN_d(w) ((int)goperl_op0(GOPERL_OP_CKWARN, (uint64_t)(w), 1))
#define Perl_ck_warner Perl_warner
#define Perl_ck_warner_d Perl_warner
#define PL_dowarn \
    ((U8)goperl_op0(GOPERL_OP_PLVAR_GET, GOPERL_PL_DOWARN, 0))
/* PL_hints is an LVALUE (keyword plugins |= HINT_BLOCK_SCOPE): reads pull
 * the guest value into a slot, and goperl_flush writes a changed slot back
 * (op PLVAR_SET) before the next guest operation. */
static U32 *goperl_hints_ref(goperl_frame_t *f) {
    goperl_hints_slot_v = (U32)goperl_do_op(f, GOPERL_OP_PLVAR_GET,
                                            GOPERL_PL_HINTS, 0, 0, 0);
    goperl_hints_base_v = goperl_hints_slot_v;
    goperl_hints_live_v = 1;
    return &goperl_hints_slot_v;
}
#define PL_hints (*goperl_hints_ref(_gof))
#define G_WARN_ON 1
#define HINT_BYTES 0x00000008
#define SVt_RV SVt_IV /* pre-5.12 name */
#define PERL_MAGIC_overload_table 'c'
#define GOPERL_INFO_AMAGIC 65536
#define SVf_AMAGIC GOPERL_INFO_AMAGIC
#define av_pop(av) GOPERL_SV(goperl_op0(GOPERL_OP_AV_POP, GOPERL_TOK(av), 0))
#define av_top_index(av) av_len(av)
#define av_tindex(av) av_len(av)
#define av_count(av) ((STRLEN)(av_len(av) + 1))
#define hv_delete_ent(hv, keysv, flags, hash)                              \
    GOPERL_SV(goperl_op0(GOPERL_OP_HV_DELETE_ENT, GOPERL_TOK(hv),          \
                         ((uint64_t)(uint32_t)(flags) << 32) |             \
                             (uint32_t)GOPERL_TOK(keysv)))
static I32 goperl_hv_exists(goperl_frame_t *f, HV *hv, const char *key,
                            I32 klen) {
    uint64_t k = goperl_do_op(f, GOPERL_OP_NEW_PVN, 0, (uint64_t)klen, key,
                              (uint64_t)klen);
    goperl_do_op(f, GOPERL_OP_SV_MORTAL, k, 0, 0, 0);
    return (I32)goperl_do_op(f, GOPERL_OP_HV_EXISTS_ENT, GOPERL_TOK(hv), k,
                             0, 0);
}
#define hv_exists(hv, key, klen) \
    goperl_hv_exists(_gof, (HV *)(hv), (key), (I32)(klen))

/* Old-style allocation names and misc compat. */
#define NEWSV(id, len) newSV(len)
#define New(id, ptr, n, t) Newx(ptr, n, t)
#define Newz(id, ptr, n, t) Newxz(ptr, n, t)
#define MEMBER_TO_FPTR(x) (x)
#define Perl_deb(...) ((void)0)
#define Strerror(e) strerror(e)
#define SvUOK(sv) ((SvFLAGS(sv) & GOPERL_INFO_ISUV) != 0)
/* Dists pair these with a value write (sv_setiv/SvIV_set), which already
 * sets the guest flags — the standalone flag flip is then a no-op. */
#define SvIOK_on(sv) ((void)(sv))
#define SvIOK_off(sv) ((void)(sv))
#define SvIV_set(sv, v) sv_setiv((SV *)(sv), (IV)(v))
#define SvNV_set(sv, v) sv_setnv((SV *)(sv), (NV)(v))
#define SvUV_set(sv, v) sv_setuv((SV *)(sv), (UV)(v))
#define PerlProc_getpid() ((int)getpid())
#define PerlProc_getuid() ((int)getuid())
#define PerlEnv_getenv(n) getenv(n)
#define SvTEMP(sv) 0 /* no host-visible TEMP flag; "never steal" is safe */
#define XSRETURN_NV(v)                            \
    STMT_START {                                  \
        ST(0) = sv_2mortal(newSVnv((NV)(v)));     \
        XSRETURN(1);                              \
    }                                             \
    STMT_END
#define DEBUG_v(x) ((void)0)
#define SvOK_off(sv) sv_setsv((SV *)(sv), &PL_sv_undef)
/* Tie magic stays guest-side (not mirrored); every hv/av op above goes
 * through the real perl API, which honors it — so "not visibly tied" makes
 * dists take their plain path, which is then still tie-correct. */
#define SvTIED_mg(sv, how) ((MAGIC *)0)
#define sv_2pv_flags(sv, lenp, flags) ((char *)goperl_sv_pv_(_gof, (SV *)(sv), (uint64_t *)(lenp)))
#define sv_cmp_flags(a, b, flags) sv_cmp((a), (b))
#define sv_copypv(dsv, ssv)                            \
    do {                                               \
        STRLEN goperl_cl;                              \
        const char *goperl_cp = SvPV((ssv), goperl_cl); \
        sv_setpvn((dsv), goperl_cp, goperl_cl);        \
        if (SvUTF8(ssv)) SvUTF8_on(dsv);               \
        else SvUTF8_off(dsv);                          \
    } while (0)
/* Overload (amagic) calls: the *_amg indexes below are perl 5.42's
 * overload.h values — the guest passes them straight to amagic_call. */
enum {
    bool__amg = 8,
    numer_amg = 9,
    string_amg = 10,
    eq_amg = 0x15,
    ne_amg = 0x16,
    seq_amg = 0x1b,
    sne_amg = 0x1c
};
#define AMG_CALLunary(sv, meth)                                     \
    GOPERL_SV(goperl_op0(GOPERL_OP_AMAGIC_CALL,                     \
                         ((uint64_t)(meth) << 32) |                 \
                             (uint32_t)GOPERL_TOK(sv),              \
                         0))
#define AMG_CALLun(sv, meth) AMG_CALLunary((sv), meth##_amg)
/* require through the guest (Perl_load_module's common use). */
static void goperl_load_module(goperl_frame_t *f, U32 flags, SV *name,
                               SV *ver, ...) {
    (void)flags;
    (void)ver;
    uint64_t len = 0;
    const char *p = goperl_api_v->sv_pv(f, GOPERL_TOK(name), &len);
    char buf[512];
    snprintf(buf, sizeof buf, "require %.*s", (int)len, p ? p : "");
    goperl_do_op(f, GOPERL_OP_EVAL_PV, 1, 0, buf, strlen(buf));
}
/* Real functions, not macros: `Perl_load_module(aTHX_ flags, ...)` puts
 * aTHX_ inside the FIRST macro argument (commas from expansions come too
 * late for argument splitting), so only a function absorbs it correctly. */
static void Perl_load_module(pTHX_ U32 flags, SV *name, SV *ver, ...)
    __attribute__((unused));
static void Perl_load_module(pTHX_ U32 flags, SV *name, SV *ver, ...) {
    goperl_load_module(goperl_cur_frame_v, flags, name, ver);
}
static void load_module(U32 flags, SV *name, SV *ver, ...)
    __attribute__((unused));
static void load_module(U32 flags, SV *name, SV *ver, ...) {
    goperl_load_module(goperl_cur_frame_v, flags, name, ver);
}

#define PL_sv_arenaroot ((SV *)0)
#define SvANY(sv) ((void *)0)
#define SVTYPEMASK 0xff

/* ---- memory helpers ------------------------------------------------------ */

#define Newx(ptr, n, t) ((ptr) = (t *)malloc((size_t)(n) * sizeof(t)))
#define Newxz(ptr, n, t) ((ptr) = (t *)calloc((size_t)(n), sizeof(t)))
#define Newxc(ptr, n, t, c) ((ptr) = (c *)malloc((size_t)(n) * sizeof(t)))
#define safemalloc(n) malloc((size_t)(n))
#define safefree(p) free((void *)(p))
#define SETERRNO(e, x) (errno = (e))
#define Renew(ptr, n, t) ((ptr) = (t *)realloc((void *)(ptr), (size_t)(n) * sizeof(t)))
#define Safefree(p) free((void *)(p))
#define Copy(s, d, n, t) ((void)memcpy((d), (s), (size_t)(n) * sizeof(t)))
#define Move(s, d, n, t) ((void)memmove((d), (s), (size_t)(n) * sizeof(t)))
#define Zero(d, n, t) ((void)memset((d), 0, (size_t)(n) * sizeof(t)))
#define StructCopy(s, d, t) (*(d) = *(s))
/* ---- assorted core-API surface (interpreter-hooking dists) --------------- */

typedef size_t Size_t;
typedef pid_t Pid_t;
#ifndef MAXPATHLEN
#define MAXPATHLEN 4096
#endif
#define Nullsv ((SV *)0)
#define Nullcv ((CV *)0)
#define Nullgv ((GV *)0)
#define Nullch ((char *)0)
#define Nullav ((AV *)0)
#define Nullhv ((HV *)0)
#define CPERLscope(x) x
#define memzero(d, n) memset((d), 0, (n))
#define Newc(i, p, n, t, c) ((p) = (c *)malloc((size_t)(n) * sizeof(t)))
#define my_snprintf snprintf
#define my_vsnprintf vsnprintf
#define HINT_STRICT_REFS 0x00000002
#define PERLSI_UNKNOWN (-1)
#define PERLSI_UNDEF 0
#define PERLSI_MAIN 1
#define PERLSI_MAGIC 2
#define PERLSI_SORT 3
#define PERLSI_SIGNAL 4
#define PERLSI_OVERLOAD 5
#define PERLSI_DESTROY 6
#define SvGMAGICAL(sv) 0
#define mg_get(sv) ((void)(sv), 0)
#define sv_free(sv) SvREFCNT_dec(sv)
#define CxTRYBLOCK(cx) 0
#define strnEQ(a, b, n) (strncmp((a), (b), (n)) == 0)
#define strnNE(a, b, n) (strncmp((a), (b), (n)) != 0)
typedef U16 OPCODE;
typedef int opcode;
#define CvFLAGS(cv) 0 /* diagnostic-only flag word; not modeled */
#define GvEGV(gv) \
    ((GV *)GOPERL_SV(goperl_op0(GOPERL_OP_GV_PTR, GOPERL_TOK(gv), 4)))
#define GvNAMELEN(gv) \
    ((U32)(uint32_t)goperl_op0(GOPERL_OP_GV_NAME, GOPERL_TOK(gv), 0))
#define hv_clear(hv) ((void)goperl_op0(GOPERL_OP_HV_CLEAR, GOPERL_TOK(hv), 0))
#define hv_delete(hv, key, klen, flags)                                    \
    GOPERL_SV(goperl_ops(GOPERL_OP_HV_DELETE, GOPERL_TOK(hv),              \
                         (uint64_t)(uint32_t)(flags), (key),               \
                         (uint64_t)(klen)))
#define Perl_debug_log stderr
#define do_sv_dump(level, fp, sv, nest, maxnest, dumpops, pvlim)           \
    fprintf((FILE *)(fp), "<sv dump unavailable in the go-perl XS SDK>\n")
/* dTHXa: the context arg is a real-perl interpreter pointer; here context
 * is the innermost live native frame, and the arg is NOT evaluated (its
 * type may not even exist without MULTIPLICITY). */
#define dTHXa(a) dTHX

/* grok_number, minimally: decimal unsigned parse (perl numeric.c). */
#define IS_NUMBER_IN_UV 0x01
#define IS_NUMBER_GREATER_THAN_UV_MAX 0x02
#define IS_NUMBER_NOT_INT 0x04
#define IS_NUMBER_NEG 0x08
#define IS_NUMBER_INFINITY 0x10
#define IS_NUMBER_NAN 0x20
static int goperl_grok_number(const char *pv, STRLEN len, UV *valuep)
    __attribute__((unused));
static int goperl_grok_number(const char *pv, STRLEN len, UV *valuep) {
    uint64_t v = 0;
    STRLEN i = 0;
    if (len == 0) return 0;
    for (; i < len && pv[i] >= '0' && pv[i] <= '9'; i++) {
        v = v * 10 + (uint64_t)(pv[i] - '0');
        /* overflow past the GUEST's 32-bit UV: report like real perl
         * (GREATER_THAN_UV_MAX, IN_UV cleared) so callers take their
         * NV path instead of storing a truncated integer. */
        if (v > 0xFFFFFFFFull)
            return IS_NUMBER_GREATER_THAN_UV_MAX;
    }
    if (i == 0 || i != len) return 0;
    if (valuep) *valuep = (UV)v;
    return IS_NUMBER_IN_UV;
}
#define grok_number(pv, len, vp) goperl_grok_number((pv), (len), (vp))

/* ninstr: find the FIRST occurrence of little in big (perl util.c). */
static char *ninstr(const char *big, const char *bigend, const char *little,
                    const char *lend) __attribute__((unused));
static char *ninstr(const char *big, const char *bigend, const char *little,
                    const char *lend) {
    const STRLEN littlelen = (STRLEN)(lend - little);
    if (littlelen == 0) return (char *)big;
    for (; big + (SSize_t)littlelen <= bigend; big++)
        if (memcmp(big, little, littlelen) == 0) return (char *)big;
    return 0;
}

/* hv_iternextsv: iterate returning value with key pv (interned copy). */
static SV *goperl_hv_iternextsv(goperl_frame_t *f, HV *hv, char **keyp,
                                I32 *retlen) __attribute__((unused));
static SV *goperl_hv_iternextsv(goperl_frame_t *f, HV *hv, char **keyp,
                                I32 *retlen) {
    uint64_t he =
        goperl_do_op(f, GOPERL_OP_HV_ITERNEXT, GOPERL_TOK(hv), 0, 0, 0);
    if (!he) return 0;
    uint64_t keysv =
        goperl_do_op(f, GOPERL_OP_HV_ITERKEYSV, he, 0, 0, 0);
    uint64_t klen = 0;
    const char *kp = goperl_api_v->sv_pv(f, keysv, &klen);
    const char *interned = kp ? goperl_intern(kp, (size_t)klen) : "";
    if (keyp) *keyp = (char *)interned;
    if (retlen) *retlen = (I32)klen;
    return GOPERL_SV(
        goperl_do_op(f, GOPERL_OP_HV_ITERVAL, GOPERL_TOK(hv), he, 0, 0));
}
#define hv_iternextsv(hv, keyp, retlenp) \
    goperl_hv_iternextsv(_gof, (HV *)(hv), (keyp), (retlenp))

#define av_exists(av, ix) \
    ((int)goperl_op0(GOPERL_OP_AV_EXISTS, GOPERL_TOK(av), (uint64_t)(int64_t)(ix)))
#define av_unshift(av, n) \
    ((void)goperl_op0(GOPERL_OP_AV_UNSHIFT, GOPERL_TOK(av), (uint64_t)(int64_t)(n)))
#define instr(big, little) strstr((big), (little))
#define SvUTF8_off(sv) \
    ((void)goperl_op0(GOPERL_OP_SV_UTF8_OFF, GOPERL_TOK(sv), 0))
#define SvREADONLY_on(sv) \
    ((void)goperl_op0(GOPERL_OP_SV_READONLY_ON, GOPERL_TOK(sv), 0))
#define SvTEMP_off(sv) ((void)(sv)) /* buffer-steal hint; not modeled */
#define sv_setuv_mg(sv, v) (sv_setuv((sv), (v)), (void)mg_set(sv))
#define C_ARRAY_LENGTH(a) (sizeof(a) / sizeof((a)[0]))
#define save_scalar(gv) \
    GOPERL_SV(goperl_op0(GOPERL_OP_SAVE_SCALAR, GOPERL_TOK(gv), 0))
#define eval_pv(code, croak_on_err)                                     \
    GOPERL_SV(goperl_ops(GOPERL_OP_EVAL_PV,                             \
                         (uint64_t)(int64_t)(croak_on_err), 0, (code),  \
                         strlen(code)))
#define newCONSTSUB(stash, name, sv)                                     \
    ((CV *)GOPERL_SV(goperl_ops(GOPERL_OP_NEW_CONSTSUB,                  \
                                GOPERL_TOK(stash), GOPERL_TOK(sv),       \
                                (name), strlen(name))))
#define SVf_UTF8 0x20000000 /* only consumed by newSVpvn_flags below */
static SV *goperl_newSVpvn_flags(goperl_frame_t *f, const char *p,
                                 STRLEN len, U32 flags)
    __attribute__((unused));
static SV *goperl_newSVpvn_flags(goperl_frame_t *f, const char *p,
                                 STRLEN len, U32 flags) {
    goperl_flush(f);
    uint64_t tok = goperl_api_v->new_pvn(f, p ? p : "", (uint64_t)len);
    if (flags & SVf_UTF8)
        goperl_do_op(f, GOPERL_OP_SV_UTF8_ON, tok, 0, 0, 0);
    if (flags & SVs_TEMP) tok = goperl_api_v->sv_mortal(f, tok);
    return GOPERL_SV(tok);
}
#define newSVpvn_flags(p, len, flags) \
    goperl_newSVpvn_flags(_gof, (p), (STRLEN)(len), (U32)(flags))
#define XSRETURN_IV(iv)                            \
    STMT_START {                                   \
        ST(0) = sv_2mortal(newSViv(iv));           \
        XSRETURN(1);                               \
    }                                              \
    STMT_END
#define XSRETURN_PV(pv)                            \
    STMT_START {                                   \
        ST(0) = sv_2mortal(newSVpv((pv), 0));      \
        XSRETURN(1);                               \
    }                                              \
    STMT_END

#define saferealloc(p, n) realloc((p), (size_t)(n))
#ifndef NV_DIG
#define NV_DIG DBL_DIG
#endif
#define SvPVbyte(sv, lenvar) SvPV((sv), (lenvar))
#define SvPVbyte_nolen(sv) SvPV_nolen(sv)
/* sv_usepvn: real perl STEALS the malloc'd buffer as the SV's PV. Here the
 * bytes are copied into the guest SV and the host buffer is freed — same
 * ownership contract, same observable bytes (SvPVX hands back a stable
 * pointer into the copy). Code that later frees memory reached THROUGH the
 * stored bytes must not be driven through this shim. */
static void goperl_sv_usepvn(goperl_frame_t *f, SV *sv, char *p, STRLEN len)
    __attribute__((unused));
static void goperl_sv_usepvn(goperl_frame_t *f, SV *sv, char *p, STRLEN len) {
    goperl_do_op(f, GOPERL_OP_SV_SETPVN, GOPERL_TOK(sv), (uint64_t)len, p,
                 (uint64_t)len);
    free(p);
}
#define sv_usepvn(sv, p, len) goperl_sv_usepvn(_gof, (SV *)(sv), (p), (len))
#define SvPV_set(sv, p) ((void)(sv)) /* buffer detach; nothing to detach */
#define SvLEN_set(sv, l) ((void)(sv))

/* sv_inc: numeric increment (string auto-increment is not modeled). */
static void goperl_sv_inc(goperl_frame_t *f, SV *sv) {
    goperl_flush(f);
    int64_t v = goperl_api_v->sv_iv(f, GOPERL_TOK(sv));
    goperl_do_op(f, GOPERL_OP_SV_SETIV, GOPERL_TOK(sv), (uint64_t)(v + 1), 0,
                 0);
}
#define sv_inc(sv) goperl_sv_inc(_gof, (SV *)(sv))

#define gv_fetchfile_flags(name, namelen, flags)                          \
    ((GV *)GOPERL_SV(goperl_ops(GOPERL_OP_GV_FETCHFILE, 0,                \
                                (uint64_t)(namelen), (name),              \
                                (uint64_t)(namelen))))
#define gv_fetchfile(name) gv_fetchfile_flags((name), strlen(name), 0)

/* my_strlcat (perl's bundled strlcat) */
static Size_t my_strlcat(char *dst, const char *src, Size_t size)
    __attribute__((unused));
static Size_t my_strlcat(char *dst, const char *src, Size_t size) {
    Size_t used = strlen(dst);
    Size_t length = strlen(src);
    if (size > 0 && used < size - 1) {
        Size_t copy = size - used - 1;
        if (copy > length) copy = length;
        memcpy(dst + used, src, copy);
        dst[used + copy] = '\0';
    }
    return used + length;
}

/* rninstr: find the LAST occurrence of little in big (perl util.c). */
static char *rninstr(const char *big, const char *bigend, const char *little,
                     const char *lend) __attribute__((unused));
static char *rninstr(const char *big, const char *bigend, const char *little,
                     const char *lend) {
    const STRLEN littlelen = (STRLEN)(lend - little);
    const char *s, *ss;
    if (littlelen == 0) return (char *)bigend;
    for (s = bigend - littlelen; s >= big; s--) {
        if (*s != *little) continue;
        for (ss = s + 1; ss < s + (SSize_t)littlelen; ss++)
            if (*ss != little[ss - s]) break;
        if (ss == s + (SSize_t)littlelen) return (char *)s;
    }
    return 0;
}

/* tryAMAGICunDEREF (the pre-5.14 idiom pp_entersub copies use): replace
 * *sp with the overloaded dereference when the SV overloads it. */
#define tryAMAGICunDEREF(meth)                                            \
    STMT_START {                                                          \
        if (SvAMAGIC(*sp)) {                                              \
            SV *goperl_amg_r =                                            \
                goperl_amagic_deref_call(_gof, *sp, CAT2(meth, _amg));    \
            if (goperl_amg_r) *sp = goperl_amg_r;                         \
        }                                                                 \
    }                                                                     \
    STMT_END

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

/* MY_CXT lives in the v5 section above: unlike real perl without ithreads,
 * pTHX_ is NON-empty here, so pMY_CXT must be a real (ignored) parameter —
 * `void` would make `(pTHX_ pMY_CXT)` signatures invalid C. */

/* Thread-clone stubs: go-perl builds without ithreads, so CLONE hooks are
 * never invoked; these only need to compile (CLONE_PARAMS is defined above
 * the MGVTBL). */
#define CLONEf_KEEP_PTR_TABLE 0
#define sv_dup(sv, param) ((void)(param), (SV *)(sv))
#define sv_dup_inc(sv, param) ((void)(param), SvREFCNT_inc((SV *)(sv)))

/* ==== the parse surface (keyword plugins, lexer, optree building) ========
 *
 * XS::Parse::Keyword-style dists hook the tokenizer: a keyword plugin runs
 * DURING guest compilation, reads the source buffer through PL_parser,
 * consumes text with lex_*, recurses into the guest parser with parse_*,
 * and hands back an optree it built with the new*OP constructors.
 *
 * Model:
 *   - PL_parser is a host SHADOW of the guest parser. Its buffer pointers
 *     point straight into guest linear memory (the linestr PV), so pointer
 *     walks and comparisons behave like real perl. The shadow is pulled
 *     from the guest at every sync point (plugin entry and after every
 *     lex_ / parse_ call, any of which may refill or realloc the buffer)
 *     and the writable fields (bufptr, in_my, error_count) are pushed back
 *     before every guest call and at plugin exit.
 *   - the new-op constructors build REAL guest ops immediately and return
 *     shadows. Field writes land on the shadow; every shadow carries a
 *     baseline copy (gop_base) and dirty fields are written back to the
 *     guest at the next optree sync point (a constructor consuming the op,
 *     or plugin exit).
 *   - an op_ppaddr write pointing at a module function makes that ONE op
 *     dispatch to the host: the guest op's op_ppaddr is patched to the
 *     pp-hook trampoline and the loader maps the op token to the function
 *     (types are never hooked wholesale, so e.g. every other OP_EQ in the
 *     program keeps running natively).
 *   - keyword plugins registered with wrap_keyword_plugin form a host
 *     chain; the guest's PL_keyword_plugin wrapper forwards candidate
 *     words over the loader (reserved method -7) and falls through to the
 *     guest's own chain when every host plugin declines.
 */

typedef U32 line_t;
typedef UV PADOFFSET;
#define NOT_IN_PAD ((PADOFFSET)-1)
#define NOLINE ((line_t)4294967295UL)

/* keywords.h (KEY_* ids) ships in the SDK include dir but is NOT included
 * here: real perl.h doesn't include it either, and dists (Moose's mop.h)
 * define their own KEY_* identifiers. Dists that want it include it
 * explicitly, as XS::Parse::Keyword does. */
#include "goperl_opargs.h"

/* op flags (values from the real op.h; op_flags crosses to the guest) */
#define OPf_WANT 3
#define OPf_WANT_VOID 1
#define OPf_WANT_SCALAR 2
#define OPf_WANT_LIST 3
#define OPf_PARENS 8
#define OPf_REF 16
#define OPf_MOD 32
#define OPf_STACKED 64
#define OPf_SPECIAL 128
#define OP_GIMME(op, dfl) \
    (((op)->op_flags & OPf_WANT) ? ((op)->op_flags & OPf_WANT) : (dfl))

/* op class + trait bits for PL_opargs (values from the real op.h) */
#define OA_MARK 1
#define OA_FOLDCONST 2
#define OA_RETSCALAR 4
#define OA_TARGET 8
#define OA_OTHERINT 16
#define OA_DANGEROUS 32
#define OA_DEFGV 64
#define OA_TARGLEX 128
#define OA_CLASS_SHIFT 13
#define OA_CLASS_MASK (15 << OA_CLASS_SHIFT)
#define OA_BASEOP (0 << OA_CLASS_SHIFT)
#define OA_UNOP (1 << OA_CLASS_SHIFT)
#define OA_BINOP (2 << OA_CLASS_SHIFT)
#define OA_LOGOP (3 << OA_CLASS_SHIFT)
#define OA_LISTOP (4 << OA_CLASS_SHIFT)
#define OA_PMOP (5 << OA_CLASS_SHIFT)
#define OA_SVOP (6 << OA_CLASS_SHIFT)
#define OA_PADOP (7 << OA_CLASS_SHIFT)
#define OA_PVOP_OR_SVOP (8 << OA_CLASS_SHIFT)
#define OA_LOOP (9 << OA_CLASS_SHIFT)
#define OA_COP (10 << OA_CLASS_SHIFT)
#define OA_BASEOP_OR_UNOP (11 << OA_CLASS_SHIFT)
#define OA_FILESTATOP (12 << OA_CLASS_SHIFT)
#define OA_LOOPEXOP (13 << OA_CLASS_SHIFT)
#define OA_METHOP (14 << OA_CLASS_SHIFT)
#define OA_UNOP_AUX (15 << OA_CLASS_SHIFT)

#define SVs_PADTMP 0x00020000
#define CVf_ANON 0x0080

/* ---- PL_parser shadow --------------------------------------------------- */

typedef struct yy_parser {
    SV *linestr;   /* guest SV handle */
    char *bufptr;  /* host views into the guest linestr PV */
    char *oldbufptr;
    char *bufend;
    char *linestart;
    I32 error_count;
    U16 in_my;
    U32 lex_flags;
    line_t preambling;
} yy_parser;

#define LEX_STUFF_UTF8 0x00000001
#define LEX_KEEP_PREVIOUS 0x00000002
#define LEX_IGNORE_UTF8_HINTS 0x00000002 /* lex_flags bit, parser.h */
#define PARSE_OPTIONAL 0x00000001

typedef char goperl_parser_fits_shared
    [(sizeof(yy_parser) <= 128) ? 1 : -1] __attribute__((unused));
#define goperl_parser_v (*(yy_parser *)GOPERL_SHARED->parser_shadow)
#define goperl_parser_ptr_v (*(yy_parser **)&GOPERL_SHARED->parser_ptr)
#define goperl_parser_pvg_v (GOPERL_SHARED->parser_pvg)

/* parser field ids shared with the guest op (PARSER_GET/PARSER_SET) */
enum {
    GOPERL_PARSER_PRESENT = 0,
    GOPERL_PARSER_LINESTR = 1,
    GOPERL_PARSER_BUFPTR = 2,
    GOPERL_PARSER_BUFEND = 3,
    GOPERL_PARSER_OLDBUFPTR = 4,
    GOPERL_PARSER_LINESTART = 5,
    GOPERL_PARSER_ERROR_COUNT = 6,
    GOPERL_PARSER_IN_MY = 7,
    GOPERL_PARSER_LEX_FLAGS = 8,
    GOPERL_PARSER_PREAMBLING = 9,
    GOPERL_PARSER_PV = 10
};

static char *goperl_parser_hptr(goperl_frame_t *f, uint32_t g, char *hbase) {
    if (!g) return 0;
    return hbase + (g - goperl_parser_pvg_v);
}

static void goperl_parser_pull(goperl_frame_t *f) {
    if (!goperl_do_op(f, GOPERL_OP_PARSER_GET, GOPERL_PARSER_PRESENT, 0, 0,
                      0)) {
        goperl_parser_ptr_v = 0;
        return;
    }
    yy_parser *p = &goperl_parser_v;
    goperl_parser_pvg_v = (uint32_t)goperl_do_op(f, GOPERL_OP_PARSER_GET,
                                                 GOPERL_PARSER_PV, 0, 0, 0);
    char *hbase =
        goperl_parser_pvg_v
            ? (char *)goperl_api_v->guest_mem(f, goperl_parser_pvg_v)
            : 0;
    p->linestr = GOPERL_SV(
        goperl_do_op(f, GOPERL_OP_PARSER_GET, GOPERL_PARSER_LINESTR, 0, 0, 0));
    p->bufptr = goperl_parser_hptr(
        f,
        (uint32_t)goperl_do_op(f, GOPERL_OP_PARSER_GET, GOPERL_PARSER_BUFPTR,
                               0, 0, 0),
        hbase);
    p->bufend = goperl_parser_hptr(
        f,
        (uint32_t)goperl_do_op(f, GOPERL_OP_PARSER_GET, GOPERL_PARSER_BUFEND,
                               0, 0, 0),
        hbase);
    p->oldbufptr = goperl_parser_hptr(
        f,
        (uint32_t)goperl_do_op(f, GOPERL_OP_PARSER_GET,
                               GOPERL_PARSER_OLDBUFPTR, 0, 0, 0),
        hbase);
    p->linestart = goperl_parser_hptr(
        f,
        (uint32_t)goperl_do_op(f, GOPERL_OP_PARSER_GET,
                               GOPERL_PARSER_LINESTART, 0, 0, 0),
        hbase);
    p->error_count = (I32)goperl_do_op(f, GOPERL_OP_PARSER_GET,
                                       GOPERL_PARSER_ERROR_COUNT, 0, 0, 0);
    p->in_my = (U16)goperl_do_op(f, GOPERL_OP_PARSER_GET, GOPERL_PARSER_IN_MY,
                                 0, 0, 0);
    p->lex_flags = (U32)goperl_do_op(f, GOPERL_OP_PARSER_GET,
                                     GOPERL_PARSER_LEX_FLAGS, 0, 0, 0);
    p->preambling = (line_t)goperl_do_op(f, GOPERL_OP_PARSER_GET,
                                         GOPERL_PARSER_PREAMBLING, 0, 0, 0);
    goperl_parser_ptr_v = p;
}

static void goperl_parser_push(goperl_frame_t *f) {
    yy_parser *p = goperl_parser_ptr_v;
    if (!p) return;
    if (p->bufptr && goperl_parser_pvg_v) {
        char *hbase = (char *)goperl_api_v->guest_mem(f, goperl_parser_pvg_v);
        goperl_do_op(f, GOPERL_OP_PARSER_SET, GOPERL_PARSER_BUFPTR,
                     goperl_parser_pvg_v + (uint32_t)(p->bufptr - hbase), 0,
                     0);
    }
    goperl_do_op(f, GOPERL_OP_PARSER_SET, GOPERL_PARSER_ERROR_COUNT,
                 (uint64_t)(uint32_t)p->error_count, 0, 0);
    goperl_do_op(f, GOPERL_OP_PARSER_SET, GOPERL_PARSER_IN_MY,
                 (uint64_t)p->in_my, 0, 0);
}

#define PL_parser (goperl_parser_ptr_v)

/* ---- optree write-back (shadow -> guest) -------------------------------- */

/* Baseline of the guest-visible writable fields, captured at shadow fill
 * and after every flush; the diff against the live shadow is what gets
 * written back. */
struct goperl_op_base {
    OP *op_next;
    OP *op_sibparent;
    U8 op_moresib;
    OP *op_first;
    OP *op_last;
    OP *op_other;
    Perl_ppaddr_t op_ppaddr;
    UV op_targ;
    U16 op_type;
    U8 op_flags;
    U8 op_private;
};

/* guest field selectors for OP_SET */
enum {
    GOPERL_OPF_NEXT = 0,
    GOPERL_OPF_SIBPARENT = 1,
    GOPERL_OPF_FIRST = 2,
    GOPERL_OPF_LAST = 3,
    GOPERL_OPF_OTHER = 4,
    GOPERL_OPF_TARG = 5,
    GOPERL_OPF_TYPE = 6,
    GOPERL_OPF_FLAGS = 7,
    GOPERL_OPF_PRIVATE = 8,
    GOPERL_OPF_PPADDR_HOOK = 9, /* patch this op's ppaddr to the pp-hook
                                   trampoline (per-op host dispatch) */
    GOPERL_OPF_MORESIB = 10
};

static void goperl_op_base_capture(OP *o) {
    if (!o->gop_base)
        o->gop_base = (struct goperl_op_base *)calloc(
            1, sizeof(struct goperl_op_base));
    struct goperl_op_base *b = o->gop_base;
    b->op_next = o->op_next;
    b->op_sibparent = o->op_sibparent;
    b->op_moresib = o->op_moresib;
    b->op_first = o->op_first;
    b->op_last = o->op_last;
    b->op_other = o->op_other;
    b->op_ppaddr = o->op_ppaddr;
    b->op_targ = o->op_targ;
    b->op_type = o->op_type;
    b->op_flags = o->op_flags;
    b->op_private = o->op_private;
}

static void goperl_perop_pp_register(goperl_frame_t *f, OP *o);

static void goperl_op_setfield(goperl_frame_t *f, OP *o, int sel,
                               uint64_t val) {
    goperl_do_op(f, GOPERL_OP_OP_SET, o->gop, ((uint64_t)sel << 32) | val, 0,
                 0);
}

static uint32_t goperl_op_tok32(goperl_frame_t *f, OP *kid) {
    if (!kid) return 0;
    if (!kid->gop)
        goperl_croakf(f, "optree write-back: scratch op (no guest identity) "
                         "linked into a guest tree");
    return (uint32_t)kid->gop;
}

static void goperl_op_flush_one(goperl_frame_t *f, OP *o) {
    struct goperl_op_base *b = o->gop_base;
    if (!o->gop || !b) return;
    if (b->op_next != o->op_next)
        goperl_op_setfield(f, o, GOPERL_OPF_NEXT,
                           goperl_op_tok32(f, o->op_next));
    if (b->op_sibparent != o->op_sibparent || b->op_moresib != o->op_moresib)
        goperl_op_setfield(f, o,
                           GOPERL_OPF_SIBPARENT |
                               ((int)(o->op_moresib ? 1 : 0) << 16),
                           goperl_op_tok32(f, o->op_sibparent));
    if (b->op_first != o->op_first)
        goperl_op_setfield(f, o, GOPERL_OPF_FIRST,
                           goperl_op_tok32(f, o->op_first));
    if (b->op_last != o->op_last)
        goperl_op_setfield(f, o, GOPERL_OPF_LAST,
                           goperl_op_tok32(f, o->op_last));
    if (b->op_other != o->op_other)
        goperl_op_setfield(f, o, GOPERL_OPF_OTHER,
                           goperl_op_tok32(f, o->op_other));
    if (b->op_targ != o->op_targ)
        goperl_op_setfield(f, o, GOPERL_OPF_TARG, (uint32_t)o->op_targ);
    if (b->op_type != o->op_type)
        goperl_op_setfield(f, o, GOPERL_OPF_TYPE, o->op_type);
    if (b->op_flags != o->op_flags)
        goperl_op_setfield(f, o, GOPERL_OPF_FLAGS, o->op_flags);
    if (b->op_private != o->op_private)
        goperl_op_setfield(f, o, GOPERL_OPF_PRIVATE, o->op_private);
    if (b->op_ppaddr != o->op_ppaddr) goperl_perop_pp_register(f, o);
    goperl_op_base_capture(o);
}

static void goperl_optree_flush(goperl_frame_t *f) {
    for (int i = 0; i < GOPERL_OPREG_BUCKETS; i++)
        for (goperl_opreg_ent_t *e = goperl_opreg_v[i]; e; e = e->next)
            goperl_op_flush_one(f, &e->op);
}

/* ---- per-op pp hooks ----------------------------------------------------
 * A module writing o->op_ppaddr points ONE op at a host function. The
 * guest op's ppaddr is patched to the pp-hook trampoline and the loader
 * maps this op's token to the function, so only this op round-trips; the
 * op type as a whole keeps running natively. */

static void goperl_perop_pp_register(goperl_frame_t *f, OP *o) {
    if (!goperl_api_v->perop_pp_set)
        goperl_croakf(f, "op_ppaddr write needs a newer go-perl loader");
    goperl_api_v->perop_pp_set(f, o->gop, (void *)o->op_ppaddr);
    goperl_op_setfield(f, o, GOPERL_OPF_PPADDR_HOOK, 0);
    /* the module reads this shadow's state whenever the op executes, which
     * can be long after this activation's registry is cleared */
    goperl_opreg_ent_t *e =
        (goperl_opreg_ent_t *)((char *)o - offsetof(goperl_opreg_ent_t, op));
    e->persist = 1;
    goperl_opext_put(o->gop, o);
}

/* ---- op constructors ---------------------------------------------------- */

/* class ids shared with the guest OPTREE_NEW op */
enum {
    GOPERL_OPC_BASEOP = 0,
    GOPERL_OPC_UNOP = 1,
    GOPERL_OPC_BINOP = 2,
    GOPERL_OPC_LISTOP = 3,
    GOPERL_OPC_LOGOP = 4,
    GOPERL_OPC_SVOP = 5,
    GOPERL_OPC_PVOP = 6,
    GOPERL_OPC_GVOP = 7,
    GOPERL_OPC_METHOP_NAMED = 8,
    GOPERL_OPC_STATEOP = 9,
    GOPERL_OPC_CONDOP = 10,
    GOPERL_OPC_SLICEOP = 11
};

static OP *goperl_newop(goperl_frame_t *f, int cls, I32 type, I32 flags,
                        uint64_t k1, uint64_t k2) {
    char kids[16];
    memcpy(kids, &k1, 8);
    memcpy(kids + 8, &k2, 8);
    goperl_optree_flush(f);
    uint64_t tok = goperl_do_op(
        f, GOPERL_OP_OPTREE_NEW,
        ((uint64_t)(uint32_t)cls << 32) | (uint32_t)type,
        (uint64_t)(uint32_t)flags, kids, 16);
    /* the constructor relinked its kids (sibling/parent) guest-side */
    goperl_op_refresh_tok(f, k1);
    goperl_op_refresh_tok(f, k2);
    return goperl_op_shadow_full(f, tok);
}

#define newOP(type, flags) goperl_newop(_gof, GOPERL_OPC_BASEOP, (type), (flags), 0, 0)
#define newUNOP(type, flags, first) \
    goperl_newop(_gof, GOPERL_OPC_UNOP, (type), (flags), \
                 (first) ? ((OP *)(first))->gop : 0, 0)
#define newBINOP(type, flags, first, last) \
    goperl_newop(_gof, GOPERL_OPC_BINOP, (type), (flags), \
                 (first) ? ((OP *)(first))->gop : 0, \
                 (last) ? ((OP *)(last))->gop : 0)
#define newLISTOP(type, flags, first, last) \
    goperl_newop(_gof, GOPERL_OPC_LISTOP, (type), (flags), \
                 (first) ? ((OP *)(first))->gop : 0, \
                 (last) ? ((OP *)(last))->gop : 0)
#define newLOGOP(type, flags, first, other) \
    goperl_newop(_gof, GOPERL_OPC_LOGOP, (type), (flags), \
                 (first) ? ((OP *)(first))->gop : 0, \
                 (other) ? ((OP *)(other))->gop : 0)
#define newSVOP(type, flags, sv) \
    goperl_newop(_gof, GOPERL_OPC_SVOP, (type), (flags), GOPERL_TOK(sv), 0)
#define newGVOP(type, flags, gv) \
    goperl_newop(_gof, GOPERL_OPC_GVOP, (type), (flags), GOPERL_TOK(gv), 0)
#define newMETHOP_named(type, flags, name_sv) \
    goperl_newop(_gof, GOPERL_OPC_METHOP_NAMED, (type), (flags), \
                 GOPERL_TOK(name_sv), 0)
#define newSTATEOP(flags, label, o) \
    goperl_newop(_gof, GOPERL_OPC_STATEOP, 0, (flags), \
                 (uint64_t)(uintptr_t)(label), (o) ? ((OP *)(o))->gop : 0)
#define newCONDOP(flags, cond, tru, fls) \
    goperl_newcondop(_gof, (flags), (OP *)(cond), (OP *)(tru), (OP *)(fls))
#define newSLICEOP(flags, subscript, listval) \
    goperl_newop(_gof, GOPERL_OPC_SLICEOP, 0, (flags), \
                 (subscript) ? ((OP *)(subscript))->gop : 0, \
                 (listval) ? ((OP *)(listval))->gop : 0)

static OP *goperl_newcondop(goperl_frame_t *f, I32 flags, OP *cond, OP *tru,
                            OP *fls) {
    char kids[24];
    uint64_t k1 = cond ? cond->gop : 0, k2 = tru ? tru->gop : 0,
             k3 = fls ? fls->gop : 0;
    memcpy(kids, &k1, 8);
    memcpy(kids + 8, &k2, 8);
    memcpy(kids + 16, &k3, 8);
    goperl_optree_flush(f);
    uint64_t tok =
        goperl_do_op(f, GOPERL_OP_OPTREE_NEW,
                     ((uint64_t)(uint32_t)GOPERL_OPC_CONDOP << 32),
                     (uint64_t)(uint32_t)flags, kids, 24);
    return goperl_op_shadow_full(f, tok);
}

static OP *goperl_newpadop_unsup(goperl_frame_t *f) {
    goperl_croakf(f, "newPADOP is not supported by the go-perl XS SDK");
    return 0;
}
#define newPADOP(type, flags, sv) goperl_newpadop_unsup(_gof)
static OP *goperl_newunop_aux_unsup(goperl_frame_t *f) {
    goperl_croakf(f, "newUNOP_AUX is not supported by the go-perl XS SDK");
    return 0;
}
#define newUNOP_AUX(type, flags, first, aux) goperl_newunop_aux_unsup(_gof)

/* ---- op list/manipulation helpers --------------------------------------- */

/* selectors shared with the guest OPTREE_MISC op */
enum {
    GOPERL_OPM_APPEND_ELEM = 0,
    GOPERL_OPM_APPEND_LIST = 1,
    GOPERL_OPM_PREPEND_ELEM = 2,
    GOPERL_OPM_CONVERT_LIST = 3,
    GOPERL_OPM_CONTEXTUALIZE = 4,
    GOPERL_OPM_SCOPE = 5,
    GOPERL_OPM_LINKLIST = 6,
    GOPERL_OPM_FREE = 7,
    GOPERL_OPM_NULL = 8,
    GOPERL_OPM_FORCE_LIST = 9,
    GOPERL_OPM_SIBLING_SPLICE = 10,
    GOPERL_OPM_OP_CLASS = 11
};

static OP *goperl_op_misc3(goperl_frame_t *f, int sel, I32 aux, U32 aux2,
                           OP *a, OP *b) {
    char toks[16];
    uint64_t t1 = a ? a->gop : 0, t2 = b ? b->gop : 0;
    if ((a && !a->gop) || (b && !b->gop))
        goperl_croakf(f, "optree helper: scratch op has no guest identity");
    memcpy(toks, &t1, 8);
    memcpy(toks + 8, &t2, 8);
    goperl_optree_flush(f);
    uint64_t tok = goperl_do_op(f, GOPERL_OP_OPTREE_MISC,
                                ((uint64_t)(uint32_t)sel << 32) |
                                    (uint32_t)aux,
                                (uint64_t)aux2, toks, 16);
    if (sel == GOPERL_OPM_FREE) return 0;
    goperl_op_refresh(f, a);
    goperl_op_refresh(f, b);
    return tok ? goperl_op_shadow_full(f, tok) : 0;
}
#define goperl_op_misc2(f, sel, aux, a, b)     goperl_op_misc3((f), (sel), (aux), 0, (a), (b))

/* op_sibling_splice: parent/start identify the splice point, del_count
 * kids come out, the insert chain goes in; returns the deleted chain. */
static OP *goperl_op_sibling_splice(goperl_frame_t *f, OP *parent, OP *start,
                                    int del_count, OP *insert) {
    char toks[24];
    uint64_t t1 = parent ? parent->gop : 0, t2 = start ? start->gop : 0,
             t3 = insert ? insert->gop : 0;
    if ((parent && !parent->gop) || (start && !start->gop) ||
        (insert && !insert->gop))
        goperl_croakf(f, "op_sibling_splice: scratch op has no guest identity");
    memcpy(toks, &t1, 8);
    memcpy(toks + 8, &t2, 8);
    memcpy(toks + 16, &t3, 8);
    goperl_optree_flush(f);
    uint64_t tok = goperl_do_op(
        f, GOPERL_OP_OPTREE_MISC,
        ((uint64_t)(uint32_t)GOPERL_OPM_SIBLING_SPLICE << 32) |
            (uint32_t)del_count,
        0, toks, 24);
    goperl_op_refresh(f, parent);
    goperl_op_refresh(f, start);
    goperl_op_refresh(f, insert);
    return tok ? goperl_op_shadow_full(f, tok) : 0;
}

#define op_append_elem(type, first, last) \
    goperl_op_misc2(_gof, GOPERL_OPM_APPEND_ELEM, (type), (OP *)(first), (OP *)(last))
#define op_append_list(type, first, last) \
    goperl_op_misc2(_gof, GOPERL_OPM_APPEND_LIST, (type), (OP *)(first), (OP *)(last))
#define op_prepend_elem(type, first, last) \
    goperl_op_misc2(_gof, GOPERL_OPM_PREPEND_ELEM, (type), (OP *)(first), (OP *)(last))
#define op_convert_list(type, flags, o) \
    goperl_op_misc3(_gof, GOPERL_OPM_CONVERT_LIST, (type), (U32)(flags), \
                    (OP *)(o), 0)
#define op_contextualize(o, context) \
    goperl_op_misc2(_gof, GOPERL_OPM_CONTEXTUALIZE, (context), (OP *)(o), 0)
#define op_scope(o) goperl_op_misc2(_gof, GOPERL_OPM_SCOPE, 0, (OP *)(o), 0)
#define op_linklist(o) goperl_op_misc2(_gof, GOPERL_OPM_LINKLIST, 0, (OP *)(o), 0)
#define LINKLIST(o) op_linklist((OP *)(o))
#define op_free(o) ((void)goperl_op_misc2(_gof, GOPERL_OPM_FREE, 0, (OP *)(o), 0))
#define op_null(o) ((void)goperl_op_misc2(_gof, GOPERL_OPM_NULL, 0, (OP *)(o), 0))
#define op_sibling_splice(parent, start, del_count, insert) \
    goperl_op_sibling_splice(_gof, (OP *)(parent), (OP *)(start), \
                             (del_count), (OP *)(insert))

/* ---- lexer bridge ------------------------------------------------------- */

enum {
    GOPERL_LEX_READ_SPACE = 0,
    GOPERL_LEX_PEEK_UNICHAR = 1,
    GOPERL_LEX_READ_UNICHAR = 2,
    GOPERL_LEX_READ_TO = 3,
    GOPERL_LEX_BUFUTF8 = 4,
    GOPERL_LEX_STUFF_PVN = 5,
    GOPERL_LEX_STUFF_SV = 6
};

static int64_t goperl_lex(goperl_frame_t *f, int sel, uint64_t b,
                          const char *s, uint64_t slen) {
    goperl_parser_push(f);
    int64_t r = (int64_t)goperl_do_op(f, GOPERL_OP_LEX,
                                      (uint64_t)(uint32_t)sel, b, s, slen);
    goperl_parser_pull(f);
    return r;
}

#define lex_read_space(flags) ((void)goperl_lex(_gof, GOPERL_LEX_READ_SPACE, (uint64_t)(uint32_t)(flags), 0, 0))
#define lex_peek_unichar(flags) ((I32)goperl_lex(_gof, GOPERL_LEX_PEEK_UNICHAR, (uint64_t)(uint32_t)(flags), 0, 0))
#define lex_read_unichar(flags) ((I32)goperl_lex(_gof, GOPERL_LEX_READ_UNICHAR, (uint64_t)(uint32_t)(flags), 0, 0))
#define lex_bufutf8() ((bool)goperl_lex(_gof, GOPERL_LEX_BUFUTF8, 0, 0, 0))
#define lex_stuff_pvn(pv, len, flags) \
    ((void)goperl_lex(_gof, GOPERL_LEX_STUFF_PVN, \
                      ((uint64_t)(uint32_t)(flags) << 32) | (uint32_t)(len), \
                      (pv), (len)))
#define lex_stuff_pvs(pv, flags) lex_stuff_pvn("" pv "", sizeof(pv) - 1, (flags))

static void goperl_lex_read_to(goperl_frame_t *f, char *ptr) {
    yy_parser *p = goperl_parser_ptr_v;
    if (!p || !p->bufptr || ptr < p->bufptr || ptr > p->bufend)
        goperl_croakf(f, "lex_read_to: pointer outside the lexer buffer");
    char *hbase = (char *)goperl_api_v->guest_mem(f, goperl_parser_pvg_v);
    goperl_lex(f, GOPERL_LEX_READ_TO,
               goperl_parser_pvg_v + (uint32_t)(ptr - hbase), 0, 0);
}
#define lex_read_to(ptr) goperl_lex_read_to(_gof, (ptr))

/* ---- parser recursion --------------------------------------------------- */

enum {
    GOPERL_PARSE_BLOCK = 0,
    GOPERL_PARSE_TERMEXPR = 1,
    GOPERL_PARSE_LISTEXPR = 2,
    GOPERL_PARSE_ARITHEXPR = 3,
    GOPERL_PARSE_FULLEXPR = 4,
    GOPERL_PARSE_FULLSTMT = 5,
    GOPERL_PARSE_STMTSEQ = 6,
    GOPERL_PARSE_BARESTMT = 7,
    GOPERL_PARSE_LABEL = 8
};

static OP *goperl_parse(goperl_frame_t *f, int sel, U32 flags) {
    goperl_parser_push(f);
    goperl_optree_flush(f);
    uint64_t tok = goperl_do_op(f, GOPERL_OP_PARSE, (uint64_t)(uint32_t)sel,
                                (uint64_t)flags, 0, 0);
    goperl_parser_pull(f);
    return tok ? goperl_op_shadow_full(f, tok) : 0;
}

#define parse_block(flags) goperl_parse(_gof, GOPERL_PARSE_BLOCK, (flags))
#define parse_termexpr(flags) goperl_parse(_gof, GOPERL_PARSE_TERMEXPR, (flags))
#define parse_listexpr(flags) goperl_parse(_gof, GOPERL_PARSE_LISTEXPR, (flags))
#define parse_arithexpr(flags) goperl_parse(_gof, GOPERL_PARSE_ARITHEXPR, (flags))
#define parse_fullexpr(flags) goperl_parse(_gof, GOPERL_PARSE_FULLEXPR, (flags))
#define parse_fullstmt(flags) goperl_parse(_gof, GOPERL_PARSE_FULLSTMT, (flags))
#define parse_stmtseq(flags) goperl_parse(_gof, GOPERL_PARSE_STMTSEQ, (flags))
#define parse_barestmt(flags) goperl_parse(_gof, GOPERL_PARSE_BARESTMT, (flags))

/* ---- blocks and pads ---------------------------------------------------- */

static I32 goperl_block_start(goperl_frame_t *f, int full) {
    goperl_parser_push(f);
    I32 r = (I32)goperl_do_op(f, GOPERL_OP_BLOCK, 0,
                              (uint64_t)(uint32_t)full, 0, 0);
    goperl_parser_pull(f);
    return r;
}
static OP *goperl_block_end(goperl_frame_t *f, I32 floor, OP *seq) {
    char toks[8];
    uint64_t t1 = seq ? seq->gop : 0;
    memcpy(toks, &t1, 8);
    goperl_parser_push(f);
    goperl_optree_flush(f);
    uint64_t tok = goperl_do_op(f, GOPERL_OP_BLOCK, 1,
                                (uint64_t)(uint32_t)floor, toks, 8);
    goperl_parser_pull(f);
    goperl_op_refresh(f, seq);
    return tok ? goperl_op_shadow_full(f, tok) : 0;
}
#define block_start(full) goperl_block_start(_gof, (full))
#define block_end(floor, seq) goperl_block_end(_gof, (floor), (OP *)(seq))
#define Perl_block_start goperl_block_start
#define Perl_block_end goperl_block_end

enum {
    GOPERL_PAD_ALLOC = 0,
    GOPERL_PAD_ADD_NAME_PVN = 1,
    GOPERL_PAD_FINDMY_PVN = 2,
    GOPERL_PAD_INTRO_MY = 3
};

static PADOFFSET goperl_pad_alloc(goperl_frame_t *f, I32 optype, U32 tmptype) {
    return (PADOFFSET)goperl_do_op(
        f, GOPERL_OP_PAD, GOPERL_PAD_ALLOC,
        ((uint64_t)(uint32_t)optype << 32) | tmptype, 0, 0);
}
static PADOFFSET goperl_pad_add_name_pvn(goperl_frame_t *f, const char *name,
                                         STRLEN len, U32 flags, HV *typestash,
                                         HV *ourstash) {
    if (typestash || ourstash)
        goperl_croakf(f, "pad_add_name_pvn: typed/our pad names are not "
                         "supported by the go-perl XS SDK");
    return (PADOFFSET)goperl_do_op(f, GOPERL_OP_PAD, GOPERL_PAD_ADD_NAME_PVN,
                                   ((uint64_t)flags << 32) | (uint32_t)len,
                                   name, (uint64_t)len);
}
static PADOFFSET goperl_pad_findmy_pvn(goperl_frame_t *f, const char *name,
                                       STRLEN len, U32 flags) {
    return (PADOFFSET)goperl_do_op(f, GOPERL_OP_PAD, GOPERL_PAD_FINDMY_PVN,
                                   ((uint64_t)flags << 32) | (uint32_t)len,
                                   name, (uint64_t)len);
}
#define pad_alloc(optype, tmptype) goperl_pad_alloc(_gof, (optype), (tmptype))
#define Perl_pad_alloc goperl_pad_alloc
#define pad_add_name_pvn(name, len, flags, typestash, ourstash) \
    goperl_pad_add_name_pvn(_gof, (name), (len), (flags), (typestash), (ourstash))
#define pad_findmy_pvn(name, len, flags) \
    goperl_pad_findmy_pvn(_gof, (name), (len), (flags))
#define intro_my() \
    ((U32)goperl_do_op(_gof, GOPERL_OP_PAD, GOPERL_PAD_INTRO_MY, 0, 0, 0))
static U32 goperl_intro_my(goperl_frame_t *f) {
    return (U32)goperl_do_op(f, GOPERL_OP_PAD, GOPERL_PAD_INTRO_MY, 0, 0, 0);
}
#define Perl_intro_my goperl_intro_my
#define padadd_NO_DUP_CHECK 0x04
#define padadd_STATE 0x02

/* the compiling pad surface start_subparse-based pieces need */
static I32 goperl_start_subparse_unsup(goperl_frame_t *f) {
    goperl_croakf(f, "start_subparse (anon-sub keyword pieces) is not "
                     "supported by the go-perl XS SDK");
    return 0;
}
#define start_subparse(outside, flags) goperl_start_subparse_unsup(_gof)
static CV *goperl_newattrsub_unsup(goperl_frame_t *f) {
    goperl_croakf(f, "newATTRSUB is not supported by the go-perl XS SDK");
    return 0;
}
#define newATTRSUB(floor, o, proto, attrs, block) goperl_newattrsub_unsup(_gof)
#define CvLVALUE_on(cv) ((void)(cv))

/* ---- keyword and infix plugins ------------------------------------------ */

#define KEYWORD_PLUGIN_DECLINE 0
#define KEYWORD_PLUGIN_STMT 1
#define KEYWORD_PLUGIN_EXPR 2

typedef int (*Perl_keyword_plugin_t)(pTHX_ char *keyword_ptr,
                                     STRLEN keyword_len, OP **op_ptr);

/* The host plugin chain head. Each wrap_keyword_plugin call pushes its
 * function and hands back the previous head (or the declining terminator)
 * as "next"; when the whole host chain declines, the guest wrapper falls
 * through to the guest's own plugin chain. */
static int goperl_kw_decline(pTHX_ char *kw, STRLEN kwlen, OP **op_ptr) {
    (void)_gof;
    (void)kw;
    (void)kwlen;
    (void)op_ptr;
    return KEYWORD_PLUGIN_DECLINE;
}
#define goperl_kw_head_v (*(Perl_keyword_plugin_t *)&GOPERL_SHARED->kw_head)
/* PL_keyword_plugin only appears in wrap_keyword_plugin-style compare/
 * assign bookkeeping; hand dists the chain head slot. */
#define PL_keyword_plugin (goperl_kw_head_v)

static void goperl_wrap_keyword_plugin(goperl_frame_t *f,
                                       Perl_keyword_plugin_t func,
                                       Perl_keyword_plugin_t *var) {
    if (*var) return;
    *var = goperl_kw_head_v ? goperl_kw_head_v : goperl_kw_decline;
    goperl_kw_head_v = func;
    goperl_do_op(f, GOPERL_OP_KEYWORD_ENABLE, 0, 0, 0, 0);
}
#define wrap_keyword_plugin(func, var) \
    goperl_wrap_keyword_plugin(_gof, (func), (var))

/* Infix plugins (5.38+ core surface). The chain registration works; the
 * guest-side wrapper reports any HOST claim of an operator loudly instead
 * of running it (custom infix operators are not bridged yet — XPK's own
 * parsing of core operators never reaches this). */
struct Perl_custom_infix;
enum Perl_custom_infix_precedence {
    INFIX_PREC_LOW = 10,
    INFIX_PREC_LOGICAL_OR_LOW = 30,
    INFIX_PREC_LOGICAL_AND_LOW = 40,
    INFIX_PREC_ASSIGN = 50,
    INFIX_PREC_LOGICAL_OR = 70,
    INFIX_PREC_LOGICAL_AND = 80,
    INFIX_PREC_REL = 90,
    INFIX_PREC_ADD = 110,
    INFIX_PREC_MUL = 130,
    INFIX_PREC_POW = 150,
    INFIX_PREC_HIGH = 170
};
struct Perl_custom_infix {
    enum Perl_custom_infix_precedence prec;
    void (*parse)(pTHX_ SV **opdata, struct Perl_custom_infix *);
    OP *(*build_op)(pTHX_ SV **opdata, OP *lhs, OP *rhs,
                    struct Perl_custom_infix *);
};
typedef STRLEN (*Perl_infix_plugin_t)(pTHX_ char *opname, STRLEN oplen,
                                      struct Perl_custom_infix **infix_ptr);
static STRLEN goperl_infix_decline(pTHX_ char *opname, STRLEN oplen,
                                   struct Perl_custom_infix **infix_ptr) {
    (void)_gof;
    (void)opname;
    (void)oplen;
    (void)infix_ptr;
    return 0;
}
#define goperl_infix_head_v (*(Perl_infix_plugin_t *)&GOPERL_SHARED->infix_head)
#define PL_infix_plugin (goperl_infix_head_v)
static void goperl_wrap_infix_plugin(goperl_frame_t *f,
                                     Perl_infix_plugin_t func,
                                     Perl_infix_plugin_t *var) {
    if (*var) return;
    *var = goperl_infix_head_v ? goperl_infix_head_v : goperl_infix_decline;
    goperl_infix_head_v = func;
    goperl_do_op(f, GOPERL_OP_KEYWORD_ENABLE, 1, 0, 0, 0);
}
#define wrap_infix_plugin(func, var) \
    goperl_wrap_infix_plugin(_gof, (func), (var))

/* Loader entry: run the host keyword-plugin chain for one candidate word.
 * req: [u32 kind (0 keyword / 1 infix)][word bytes]; resp on success:
 * [1][u32 ret][u64 op token]; on croak the frame error is set. */
__attribute__((weak, visibility("default"), used)) int32_t
__goperl_keyword_invoke(goperl_frame_t *f, const char *word, uint32_t wordlen,
                        uint32_t kind, uint32_t *ret_out, uint64_t *op_out) {
    jmp_buf jb;
    void *prev_jb = f->jb;
    goperl_frame_t *prev_frame = goperl_cur_frame_v;
    f->jb = (void *)&jb;
    f->prev_frame = (void *)prev_frame;
    goperl_cur_frame_v = f;
    goperl_xs_depth_v++;
    f->hostsave_base = goperl_hostsave_n;
    /* shadows made while compiling a keyword describe optree nodes that
     * outlive this activation (the module stores raw pointers to them and
     * reads them when the compiled code RUNS): allocate them persistently */
    int32_t prev_persist = goperl_opreg_persist_mode_v;
    goperl_opreg_persist_mode_v = 1;
    if (setjmp(jb)) {
        goperl_opreg_persist_mode_v = prev_persist;
        goperl_hostsave_unwind_to(f, f->hostsave_base);
        goperl_xs_depth_v--;
        goperl_cur_frame_v = prev_frame;
        f->jb = prev_jb;
        return -1; /* croak: message in f->err */
    }
    int32_t rc = 0;
    if (kind == 0 && goperl_kw_head_v) {
        goperl_parser_pull(f);
        /* the plugin may keep pointers into the word; give it a stable
         * writable copy (real perl hands the tokenizer's buffer) */
        char wbuf[256];
        if (wordlen >= sizeof(wbuf))
            goperl_croakf(f, "keyword too long");
        memcpy(wbuf, word, wordlen);
        wbuf[wordlen] = 0;
        OP *op = 0;
        int ret = goperl_kw_head_v(f, wbuf, wordlen, &op);
        if (ret != KEYWORD_PLUGIN_DECLINE) {
            goperl_optree_flush(f);
            goperl_parser_push(f);
            *ret_out = (uint32_t)ret;
            *op_out = op ? op->gop : 0;
            rc = 1;
        } else {
            goperl_parser_push(f);
        }
    } else if (kind == 1 && goperl_infix_head_v) {
        struct Perl_custom_infix *inf = 0;
        char wbuf[256];
        if (wordlen >= sizeof(wbuf)) goperl_croakf(f, "operator too long");
        memcpy(wbuf, word, wordlen);
        wbuf[wordlen] = 0;
        if (goperl_infix_head_v(f, wbuf, wordlen, &inf) != 0)
            goperl_croakf(f,
                          "custom infix operator '%s' is not supported by "
                          "the go-perl XS SDK yet",
                          wbuf);
    }
    goperl_opreg_persist_mode_v = prev_persist;
    goperl_xs_depth_v--;
    goperl_cur_frame_v = prev_frame;
    f->jb = prev_jb;
    return rc;
}

/* ---- custom op registration --------------------------------------------- */

typedef struct xop {
    const char *xop_name;
    const char *xop_desc;
    U32 xop_class;
    void (*xop_peep)(pTHX_ OP *o, OP *oldop);
} XOP;
#define XOPf_xop_name 0x01
#define XOPf_xop_desc 0x02
#define XOPf_xop_class 0x04
#define XOPf_xop_peep 0x08
#define XopFLAGS(xop) 0
#define XopENTRY_set(xop, which, to) ((xop)->which = (to))
#define XopENTRY(xop, which) ((xop)->which)
#define XopDISABLE(xop, which) ((xop)->which = 0)
/* Host bookkeeping only: names/descs feed diagnostics; the guest sees
 * custom ops as OP_CUSTOM either way. */
#define GOPERL_XOP_MAX 64
typedef struct {
    Perl_ppaddr_t pp;
    const XOP *xop;
} goperl_xop_ent_t;
__attribute__((weak)) goperl_xop_ent_t goperl_xops_v[GOPERL_XOP_MAX];
__attribute__((weak)) int32_t goperl_xop_n_v = 0;
static void goperl_custom_op_register(goperl_frame_t *f, Perl_ppaddr_t pp,
                                      const XOP *xop) {
    (void)f;
    if (goperl_xop_n_v < GOPERL_XOP_MAX) {
        goperl_xops_v[goperl_xop_n_v].pp = pp;
        goperl_xops_v[goperl_xop_n_v].xop = xop;
        goperl_xop_n_v++;
    }
}
#define Perl_custom_op_register goperl_custom_op_register
#define custom_op_register(pp, xop) \
    goperl_custom_op_register(_gof, (pp), (xop))

/* ---- character classification via the guest ----------------------------- */

enum {
    GOPERL_CC_IDFIRST = 0,
    GOPERL_CC_IDCONT = 1,
    GOPERL_CC_WORDCHAR = 2,
    GOPERL_CC_SPACE = 3,
    GOPERL_CC_DIGIT = 4,
    GOPERL_CC_ALPHA = 5
};
static int goperl_charclass_uv(goperl_frame_t *f, int cls, UV cp) {
    return (int)goperl_do_op(f, GOPERL_OP_CHARCLASS,
                             (uint64_t)(uint32_t)cls, (uint64_t)cp, 0, 0);
}
static UV goperl_utf8_cp(goperl_frame_t *f, const U8 *p, const U8 *e) {
    /* decode one UTF-8 char (the buffer is guest linestr text, always
     * well-formed by the time the lexer hands it out) */
    (void)f;
    UV c = *p;
    if (c < 0x80 || p + 1 > e) return c;
    int n = c >= 0xF0 ? 4 : c >= 0xE0 ? 3 : c >= 0xC2 ? 2 : 1;
    if (n == 1 || p + n > e) return c;
    c &= (0x7F >> n);
    for (int i = 1; i < n; i++) c = (c << 6) | (p[i] & 0x3F);
    return c;
}
#define isIDFIRST_uni(c) goperl_charclass_uv(_gof, GOPERL_CC_IDFIRST, (UV)(c))
#define isIDCONT_uni(c) goperl_charclass_uv(_gof, GOPERL_CC_IDCONT, (UV)(c))
#define isALNUM_uni(c) goperl_charclass_uv(_gof, GOPERL_CC_WORDCHAR, (UV)(c))
#define isSPACE_uni(c) goperl_charclass_uv(_gof, GOPERL_CC_SPACE, (UV)(c))
#define isIDFIRST_utf8_safe(p, e) \
    goperl_charclass_uv(_gof, GOPERL_CC_IDFIRST, \
                        goperl_utf8_cp(_gof, (const U8 *)(p), (const U8 *)(e)))
#define isIDCONT_utf8_safe(p, e) \
    goperl_charclass_uv(_gof, GOPERL_CC_IDCONT, \
                        goperl_utf8_cp(_gof, (const U8 *)(p), (const U8 *)(e)))
#define isALNUM_utf8_safe(p, e) \
    goperl_charclass_uv(_gof, GOPERL_CC_WORDCHAR, \
                        goperl_utf8_cp(_gof, (const U8 *)(p), (const U8 *)(e)))

/* ---- assorted compat the parse dists use -------------------------------- */

#define NUM2PTR(type, num) ((type)(uintptr_t)(num))
#define PerlMemShared_malloc(size) malloc(size)
#define PerlMemShared_calloc(count, size) calloc((count), (size))
#define PerlMemShared_free(p) free(p)

#define SAVEFREESV(sv) \
    ((void)goperl_do_op(_gof, GOPERL_OP_SAVE_MISC, 0, GOPERL_TOK(sv), 0, 0))

/* pp-function return conventions (host pp fns run under the pp-hook
 * bridge; PL_op is the shadow of the executing guest op) */
#define NORMAL (PL_op->op_next)
#define RETURN return (PUTBACK, NORMAL)
#define RETURNOP(o) return (PUTBACK, (o))

#define mXPUSHu(u) XPUSHs(sv_2mortal(newSVuv(u)))
#define mXPUSHi(i) XPUSHs(sv_2mortal(newSViv(i)))
#define mXPUSHp(p, l) XPUSHs(sv_2mortal(newSVpvn((p), (l))))

/* sv class checks against the guest: mode 0 = derived_from name string
 * (s = name bytes), mode 1 = derived_from an SV name (s = 8-byte token),
 * mode 2 = sv_isa_sv (s = 8-byte token). */
static int goperl_sv_classify(goperl_frame_t *f, SV *sv, int mode, U32 flags,
                              const char *name, STRLEN len, SV *namesv) {
    /* b layout (shared with the guest): len<<32 | flags<<8 | mode */
    if (mode == 0)
        return (int)goperl_do_op(f, GOPERL_OP_SV_CLASSIFY, GOPERL_TOK(sv),
                                 ((uint64_t)(uint32_t)len << 32) |
                                     ((uint64_t)(flags & 0xFFFFFFu) << 8),
                                 name, (uint64_t)len);
    char tokbuf[8];
    uint64_t nt = GOPERL_TOK(namesv);
    memcpy(tokbuf, &nt, 8);
    return (int)goperl_do_op(f, GOPERL_OP_SV_CLASSIFY, GOPERL_TOK(sv),
                             ((uint64_t)(flags & 0xFFFFFFu) << 8) |
                                 (uint32_t)(uint8_t)mode,
                             tokbuf, 8);
}
#define sv_derived_from_pvn(sv, name, len, flags) \
    goperl_sv_classify(_gof, (SV *)(sv), 0, (U32)(flags), (name), (len), 0)
#define sv_derived_from_sv(sv, namesv, flags) \
    goperl_sv_classify(_gof, (SV *)(sv), 1, (U32)(flags), 0, 0, (SV *)(namesv))
#define sv_isa_sv(sv, namesv) \
    goperl_sv_classify(_gof, (SV *)(sv), 2, 0, 0, 0, (SV *)(namesv))

/* stash-name metadata via the guest (mode 3 = name length in bytes,
 * mode 4 = name is UTF-8) */
#define HvNAMELEN(hv) \
    ((I32)goperl_sv_classify(_gof, (SV *)(hv), 3, 0, 0, 0, 0))
#define HvNAMEUTF8(hv) \
    goperl_sv_classify(_gof, (SV *)(hv), 4, 0, 0, 0, 0)
#define HvNAMELEN_get(hv) HvNAMELEN(hv)

struct op_argcheck_aux {
    UV params;
    UV opt_params;
    char slurpy;
};

#define GOPERL_PAD_SETSV 4
#define PAD_SETSV(ix, sv) \
    ((void)goperl_do_op(_gof, GOPERL_OP_PAD, GOPERL_PAD_SETSV, \
                        ((uint64_t)GOPERL_TOK(sv) << 32) | (uint32_t)(ix), \
                        0, 0))

static GV *goperl_gv_fetchsv(goperl_frame_t *f, SV *name, I32 flags,
                             int svt) {
    STRLEN len;
    const char *pv = SvPV((SV *)name, len);
    return (GV *)GOPERL_SV(goperl_do_op(f, GOPERL_OP_GV_FETCH,
                                        (uint64_t)(int64_t)flags,
                                        (uint64_t)(int64_t)svt, pv,
                                        (uint64_t)len));
}
#define gv_fetchsv(name, flags, svt) goperl_gv_fetchsv(_gof, (name), (flags), (svt))

static void goperl_scan_version_unsup(goperl_frame_t *f) {
    goperl_croakf(f, "scan_version is not supported by the go-perl XS SDK");
}
#define scan_version(s, rv, qv) \
    (goperl_scan_version_unsup(_gof), (const char *)0)

#define op_dump(o) goperl_do_op_dump(0, stderr, (const OP *)(o))

/* call-checker surface (registered but not bridged; croaks on use) */
typedef OP *(*Perl_call_checker)(pTHX_ OP *o, GV *namegv, SV *ckobj);
static void goperl_call_checker_unsup(goperl_frame_t *f) {
    goperl_croakf(f, "cv_set_call_checker is not supported by the go-perl "
                     "XS SDK");
}
#define cv_set_call_checker(cv, ckfun, ckobj) goperl_call_checker_unsup(_gof)
#define cv_set_call_checker_flags(cv, ckfun, ckobj, flags) \
    goperl_call_checker_unsup(_gof)
static OP *goperl_ck_entersub_unsup(goperl_frame_t *f) {
    goperl_croakf(f, "ck_entersub_args_proto_or_list is not supported by "
                     "the go-perl XS SDK");
    return 0;
}
#define ck_entersub_args_proto_or_list(o, namegv, ckobj) \
    goperl_ck_entersub_unsup(_gof)

/* interpreter globals the parse surface reads */
#define PL_defgv \
    ((GV *)(uintptr_t)goperl_do_op(_gof, GOPERL_OP_PLVAR_GET, \
                                   GOPERL_PL_DEFGV, 0, 0, 0))
#define PL_hintgv \
    ((GV *)(uintptr_t)goperl_do_op(_gof, GOPERL_OP_PLVAR_GET, \
                                   GOPERL_PL_HINTGV, 0, 0, 0))
#define PL_compcv \
    ((CV *)(uintptr_t)goperl_do_op(_gof, GOPERL_OP_PLVAR_GET, \
                                   GOPERL_PL_COMPCV, 0, 0, 0))
#define HINT_UTF8 0x00800000

/* ---- module-defined op structs (BASEOP extensions) ---------------------- */

/* NewOpSz backs a dist's own op struct (struct myop { BASEOP ... }): the
 * host allocation is an oversized PERSISTENT shadow-registry entry (the
 * dist keeps state in the extension fields and reads it at runtime, long
 * after this activation's registry is cleared) adopted onto a fresh raw
 * guest op, so linking it into a guest optree works like any other op. */
enum { GOPERL_OPC_RAW = 12 };
static OP *goperl_newopsz(goperl_frame_t *f, size_t size) {
    if (size < sizeof(OP)) size = sizeof(OP);
    uint64_t tok = goperl_do_op(f, GOPERL_OP_OPTREE_NEW,
                                ((uint64_t)GOPERL_OPC_RAW << 32), 0, 0, 0);
    uint32_t h = (uint32_t)(tok >> 3) % GOPERL_OPREG_BUCKETS;
    goperl_opreg_ent_t *e = (goperl_opreg_ent_t *)calloc(
        1, offsetof(goperl_opreg_ent_t, op) + size);
    if (!e) goperl_croakf(f, "NewOpSz: out of memory");
    e->persist = 1;
    e->op.gop = tok;
    e->op.gop_full = 1;
    goperl_op_base_capture(&e->op);
    e->next = goperl_opreg_v[h];
    goperl_opreg_v[h] = e;
    return &e->op;
}
#define NewOpSz(m, var, size) ((var) = (OP *)goperl_newopsz(_gof, (size)))
#define NewOp(m, var, n, type) \
    ((var) = (type *)goperl_newopsz(_gof, sizeof(type) * (n)))

/* ---- runtime pad access (pp functions) ---------------------------------- */

#define GOPERL_PAD_SV_FETCH 5
static SV *goperl_pad_sv_fetch(goperl_frame_t *f, PADOFFSET ix) {
    return GOPERL_SV(goperl_do_op(f, GOPERL_OP_PAD, GOPERL_PAD_SV_FETCH,
                                  (uint64_t)ix, 0, 0));
}
#define PAD_SV(ix) goperl_pad_sv_fetch(_gof, (PADOFFSET)(ix))
#define dTARGET SV *targ = PAD_SV(PL_op->op_targ)
#define GETTARGET dTARGET

/* numeric/string equality with overload control */
#define SV_SKIP_OVERLOAD (1 << 13)
#define sv_numeq_flags(sv1, sv2, flags) \
    goperl_sv_classify(_gof, (SV *)(sv1), 5, (U32)(flags), 0, 0, (SV *)(sv2))
#define sv_streq_flags(sv1, sv2, flags) \
    goperl_sv_classify(_gof, (SV *)(sv1), 6, (U32)(flags), 0, 0, (SV *)(sv2))
#define sv_numeq(sv1, sv2) sv_numeq_flags((sv1), (sv2), 0)
#define sv_streq(sv1, sv2) sv_streq_flags((sv1), (sv2), 0)

#define cPMOPx(o) ((OP *)(o))

#define pad_add_name_pvs(name, flags, typestash, ourstash) \
    pad_add_name_pvn("" name "", sizeof(name) - 1, (flags), (typestash), \
                     (ourstash))

/* free host memory when the guest scope pops (rides the guest savestack
 * through the SDK's host-destructor bridge) */
static void goperl_savefreepv_cb(pTHX_ void *p) {
    (void)_gof;
    free(p);
}
#define SAVEFREEPV(p) SAVEDESTRUCTOR_X(goperl_savefreepv_cb, (void *)(p))

/* op_private flag values (from the guest perl's generated opcode.h) */
#define OPpCONST_STRICT 0x08
#define OPpCONST_BARE 0x20
#define OPpLVAL_INTRO 0x80
#define HINT_BLOCK_SCOPE 0x00000100

#endif /* GOPERL_XS_SDK_PERL_H */
