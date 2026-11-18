/* go-perl native XS SDK: the part of perl's public XS API that needs no
 * host-side model of its own. Each definition is either the interpreter's
 * own macro body (sv.h, cv.h, gv.h, handy.h) expressed over the flag-word
 * and slot operations of the guest, a direct call of the interpreter
 * function of the same name, or a pure C helper that never touches the
 * interpreter (number scanning, UTF-8 inspection, libc aliases).
 *
 * Included at the end of perl.h; not meant to be included on its own. */
#ifndef GOPERL_XS_SDK_PERLAPI_H
#define GOPERL_XS_SDK_PERLAPI_H
#pragma GCC system_header

/* A function the SDK does not provide fails when the module is loaded: the
 * loader resolves every symbol of a module then (RTLD_NOW), and the error
 * names the function. So a call of a function without a declaration is
 * what it is on perl, a warning: a module that calls a function of its own
 * that it does not declare builds and works. */

/* ---- guest operation selectors (mirror perl-wasm's perl_xs_helper) ------ */

/* SAVE_MISC selectors. */
enum {
    GOPERL_SAVE_FREESV = 0,
    GOPERL_SAVE_MORTALIZESV = 1,
    GOPERL_SAVE_ITEM = 2,
    GOPERL_SAVE_ARY = 3,
    GOPERL_SAVE_HASH = 4,
    GOPERL_SAVE_GP = 5,
    GOPERL_SAVE_HINTS = 6,
    GOPERL_SAVE_COMPILE_WARNINGS = 7,
    GOPERL_SAVE_DEFSV = 8
};

/* SV_SLOT_SET selectors. */
enum {
    GOPERL_SVSLOT_RV = 0,
    GOPERL_SVSLOT_STASH = 1,
    GOPERL_SVSLOT_IV = 2,
    GOPERL_SVSLOT_UV = 3,
    GOPERL_SVSLOT_NV = 4,
    GOPERL_SVSLOT_PV = 5,
    GOPERL_SVSLOT_LEN = 6,
    GOPERL_SVSLOT_GV_STASH = 7,
    GOPERL_SVSLOT_GV_CVGEN = 8
};

/* SV_CALL selectors. */
enum {
    GOPERL_SVCALL_TAINT_ENV = 54,
    GOPERL_SVCALL_TIED = 56,
    GOPERL_SVCALL_PV_AS = 57,
    GOPERL_SVCALL_PV_FORCE_AS = 58,
    GOPERL_SVCALL_MG_GET = 0,
    GOPERL_SVCALL_MG_CLEAR = 1,
    GOPERL_SVCALL_MG_FREE = 2,
    GOPERL_SVCALL_MG_MAGICAL = 3,
    GOPERL_SVCALL_MG_SIZE = 4,
    GOPERL_SVCALL_BACKOFF = 5,
    GOPERL_SVCALL_UNREF = 6,
    GOPERL_SVCALL_RVUNWEAKEN = 7,
    GOPERL_SVCALL_UTF8_ENCODE = 8,
    GOPERL_SVCALL_LEN_UTF8 = 9,
    GOPERL_SVCALL_DEC = 10,
    GOPERL_SVCALL_TAINTED = 11,
    GOPERL_SVCALL_TAINT = 12,
    GOPERL_SVCALL_UNTAINT = 13,
    GOPERL_SVCALL_REF = 14,
    GOPERL_SVCALL_CV_CONST_SV = 15,
    GOPERL_SVCALL_CV_UNDEF = 16,
    GOPERL_SVCALL_POS_U2B = 17,
    GOPERL_SVCALL_POS_B2U = 18,
    GOPERL_SVCALL_SET_UNDEF = 19,
    GOPERL_SVCALL_SHRINK_TO_CUR = 20,
    GOPERL_SVCALL_AV_DELETE = 21,
    GOPERL_SVCALL_VOK = 22,
    GOPERL_SVCALL_HV_USEDKEYS = 23,
    GOPERL_SVCALL_HV_TOTALKEYS = 24,
    GOPERL_SVCALL_HV_FILL = 25,
    GOPERL_SVCALL_HV_MAX = 26,
    GOPERL_SVCALL_AV_MAX = 27,
    GOPERL_SVCALL_UTF8_UPGRADE = 28,
    GOPERL_SVCALL_NEWSVSV = 29,
    GOPERL_SVCALL_NEWAVAV = 30,
    GOPERL_SVCALL_NEWAVHV = 31,
    GOPERL_SVCALL_SETRV_INC = 32,
    GOPERL_SVCALL_SETRV_NOINC = 33,
    GOPERL_SVCALL_SETSV = 34,
    GOPERL_SVCALL_CATSV = 35,
    GOPERL_SVCALL_CMP = 36,
    GOPERL_SVCALL_EQ = 37,
    GOPERL_SVCALL_MRO_LINEAR_ISA = 38,
    GOPERL_SVCALL_MRO_METHOD_CHANGED = 39,
    GOPERL_SVCALL_2BOOL = 40,
    GOPERL_SVCALL_2IV = 41,
    GOPERL_SVCALL_2UV = 42,
    GOPERL_SVCALL_2NV = 43,
    GOPERL_SVCALL_2PV = 44,
    GOPERL_SVCALL_CLEAR_UNDEF = 45,
    GOPERL_SVCALL_IO_TYPE = 46,
    GOPERL_SVCALL_IO_FLAGS = 47,
    GOPERL_SVCALL_IO_LINES = 48,
    GOPERL_SVCALL_SETPV_BUFSIZE = 49,
    GOPERL_SVCALL_THROW = 50,
    GOPERL_SVCALL_2PVBYTE = 51,
    GOPERL_SVCALL_HV_KSPLIT = 52,
    GOPERL_SVCALL_UNI_DISPLAY = 53
};

/* ---- globs without a name, random numbers, comparison in the locale ---------- */

/* newGVgen: a new glob under a name nobody has, in the package (CALL 18). */
#define newGVgen_flags(pack, flags)                                              \
    ({ const char *const goperl_newGVgen_pack_ = (pack);                         \
       (GV *)GOPERL_SV(goperl_do_op3(_gof, GOPERL_OP_CALL, 0, (uint32_t)(flags), 18, \
                                     goperl_newGVgen_pack_,                       \
                                     __builtin_strlen(goperl_newGVgen_pack_))); })
#define newGVgen(pack) newGVgen_flags((pack), 0)
/* The random numbers of the interpreter (CALL 19): Drand01 is the next one
 * of the generator rand() uses, seedDrand01 seeds it, seed() is what srand
 * without an argument seeds it with, and PL_srand_called says whether it
 * has been seeded. */
typedef U32 Rand_seed_t;
static NV goperl_drand01(goperl_frame_t *f) __attribute__((unused));
static NV goperl_drand01(goperl_frame_t *f) {
    const uint64_t bits = goperl_do_op3(f, GOPERL_OP_CALL, 0, 0, (uint64_t)(19), 0, 0);
    double d;
    memcpy(&d, &bits, sizeof d);
    return (NV)d;
}
#define Drand01() goperl_drand01(_gof)
#define seedDrand01(x) \
    ((void)goperl_op3(GOPERL_OP_CALL, 0, ((uint64_t)(Rand_seed_t)(x) << 2) | 1, 19))
#define seed() ((UV)goperl_op3(GOPERL_OP_CALL, 0, 2, 19))
#define Perl_seed(ctx) seed()
#define GOPERL_PL_SRAND_CALLED 78
#define GOPERL_PL_SV_COUNT 79
#define GOPERL_PL_SIG_PENDING 80
#define PL_srand_called (*goperl_plint_ref(_gof, GOPERL_PL_SRAND_CALLED))
/* How many SVs the interpreter has, and whether signals wait to be
 * delivered: read only. */
#define PL_sv_count ((IV)GOPERL_PLVAR(GOPERL_PL_SV_COUNT))
/* The interpreter's own variable, in its memory: native code reads it, and
 * takes its address to watch it (Data::HashMap::Shared, Data::ReqRep::Shared
 * compare it before and after a wait). */
#define GOPERL_PL_SIG_PENDING_ADDR 117
#define GOPERL_PL_PSIG_PEND_ADDR 118
static volatile int *goperl_sig_pending_ref(goperl_frame_t *f) {
    volatile int *p = (volatile int *)goperl_guest_mem(
        f, goperl_xs_op(f, GOPERL_OP_PLVAR_GET, GOPERL_PL_SIG_PENDING_ADDR, 0, 0, 0));
    if (!p) goperl_croakf(f, "PL_sig_pending: the interpreter did not give its address");
    return p;
}
#define PL_sig_pending (*goperl_sig_pending_ref(_gof))
/* the pending count of each signal, the interpreter's array */
#define PL_psig_pend \
    ((int *)goperl_guest_mem(_gof, goperl_xs_op(_gof, GOPERL_OP_PLVAR_GET, GOPERL_PL_PSIG_PEND_ADDR, 0, 0, 0)))

/* The names and numbers of the signals, as perl's Configure finds them on
 * the system native code runs on (config.h's SIG_NAME, SIG_NUM and
 * SIG_SIZE): a name by number for every number up to the highest, NUMn
 * where the system has none, and then the other names of signals that
 * have one already. Native code works with the numbers of the system it
 * runs on, which an interpreter running as a process of it uses too. */
#if defined(__APPLE__)
#define SIG_NAME "ZERO", "HUP", "INT", "QUIT", "ILL", "TRAP", "ABRT", "EMT", "FPE", \
    "KILL", "BUS", "SEGV", "SYS", "PIPE", "ALRM", "TERM", "URG", "STOP", "TSTP",   \
    "CONT", "CHLD", "TTIN", "TTOU", "IO", "XCPU", "XFSZ", "VTALRM", "PROF",         \
    "WINCH", "INFO", "USR1", "USR2", "IOT", 0
#define SIG_NUM 0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, \
    19, 20, 21, 22, 23, 24, 25, 26, 27, 28, 29, 30, 31, 6, 0
#define SIG_SIZE 33
#elif defined(__linux__)
#define SIG_NAME "ZERO", "HUP", "INT", "QUIT", "ILL", "TRAP", "ABRT", "BUS", "FPE", \
    "KILL", "USR1", "SEGV", "USR2", "PIPE", "ALRM", "TERM", "STKFLT", "CHLD",      \
    "CONT", "STOP", "TSTP", "TTIN", "TTOU", "URG", "XCPU", "XFSZ", "VTALRM",       \
    "PROF", "WINCH", "IO", "PWR", "SYS", "NUM32", "NUM33", "RTMIN", "NUM35",        \
    "NUM36", "NUM37", "NUM38", "NUM39", "NUM40", "NUM41", "NUM42", "NUM43",         \
    "NUM44", "NUM45", "NUM46", "NUM47", "NUM48", "NUM49", "NUM50", "NUM51",         \
    "NUM52", "NUM53", "NUM54", "NUM55", "NUM56", "NUM57", "NUM58", "NUM59",         \
    "NUM60", "NUM61", "NUM62", "NUM63", "RTMAX", "IOT", "CLD", "POLL", 0
#define SIG_NUM 0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, \
    19, 20, 21, 22, 23, 24, 25, 26, 27, 28, 29, 30, 31, 32, 33, 34, 35, 36, 37,   \
    38, 39, 40, 41, 42, 43, 44, 45, 46, 47, 48, 49, 50, 51, 52, 53, 54, 55, 56,   \
    57, 58, 59, 60, 61, 62, 63, 64, 6, 17, 29, 0
#define SIG_SIZE 68
#endif
#ifdef SIG_SIZE
static const char *const goperl_sig_name_v[] __attribute__((unused)) = {SIG_NAME};
static const int goperl_sig_num_v[] __attribute__((unused)) = {SIG_NUM};
#define PL_sig_name ((char **)goperl_sig_name_v)
#define PL_sig_num ((int *)goperl_sig_num_v)
#endif
/* sv_cmp by the rules of the locale in effect (SV_CALL 60). */
#define sv_cmp_locale_flags(a, b, flags) \
    ((I32)(int64_t)GOPERL_SVCALL2(60, a, b, flags))
#define sv_cmp_locale(a, b) sv_cmp_locale_flags((a), (b), SV_GMAGIC)
/* as a function, for code that hands it to sortsv */
static I32 goperl_sv_cmp_locale_fn(pTHX_ SV *a, SV *b) __attribute__((unused));
#define Perl_sv_cmp_locale goperl_sv_cmp_locale_fn

/* ---- names that are other names -------------------------------------------- */

/* The errors of the compilation in progress, by the name it had before
 * the parser was a structure. */
#define PL_error_count (PL_parser->error_count)

/* hv_common_key_len is hv_common for a key given as a string, whose length
 * is negative when the key is UTF-8 (hv_fetch's convention). */
#define hv_common_key_len(hv, key, klen_i32, action, val, hash)            \
    ({ const I32 goperl_hckl_len_ = (I32)(klen_i32);                        \
       hv_common((hv), NULL, (key),                                         \
                 (STRLEN)(goperl_hckl_len_ < 0 ? -goperl_hckl_len_          \
                                               : goperl_hckl_len_),         \
                 (goperl_hckl_len_ < 0 ? HVhek_UTF8 : 0), (action), (val),  \
                 (hash)); })
#define Perl_hv_common_key_len(ctx, hv, key, klen, action, val, hash) \
    hv_common_key_len((hv), (key), (klen), (action), (val), (hash))
/* The slots of a glob made on demand, by their old names: each returns the
 * glob. */
#define gv_SVadd(gv) ({ GV *const goperl_gva_ = (GV *)(gv); (void)GvSVn(goperl_gva_); goperl_gva_; })
#define gv_AVadd(gv) ({ GV *const goperl_gva_ = (GV *)(gv); (void)GvAVn(goperl_gva_); goperl_gva_; })
#define gv_HVadd(gv) ({ GV *const goperl_gva_ = (GV *)(gv); (void)GvHVn(goperl_gva_); goperl_gva_; })
#define gv_IOadd(gv) ({ GV *const goperl_gva_ = (GV *)(gv); (void)GvIOn(goperl_gva_); goperl_gva_; })
/* hv_magic is sv_magic on a hash. */
#define hv_magic(hv, gv, how) sv_magic((SV *)(hv), (SV *)(gv), (how), NULL, 0)
/* Whether an array owns its elements, and whether it has yet to take them
 * over (@_). */
#define AvREAL(av) ((SvFLAGS(av) & SVpav_REAL) != 0)
#define AvREIFY(av) ((SvFLAGS(av) & SVpav_REIFY) != 0)
#define AvREALISH(av) (SvFLAGS(av) & (SVpav_REAL | SVpav_REIFY))
/* Whether an array counts a reference for each element (av.h). */
#define AvREAL_on(av) ((void)GOPERL_FLAGS_CHANGE(av, SVpav_REAL, 0))
#define AvREAL_off(av) ((void)GOPERL_FLAGS_CHANGE(av, 0, SVpav_REAL))
#define AvREAL_only(av) ((void)GOPERL_FLAGS_CHANGE(av, SVpav_REAL, SVpav_REIFY))
#define AvREIFY_on(av) ((void)GOPERL_FLAGS_CHANGE(av, SVpav_REIFY, 0))
#define AvREIFY_off(av) ((void)GOPERL_FLAGS_CHANGE(av, 0, SVpav_REIFY))
#define AvREIFY_only(av) ((void)GOPERL_FLAGS_CHANGE(av, SVpav_REIFY, SVpav_REAL))
/* What is in a pad slot: a glob, or a constant. */
#define IS_PADGV(v) ((v) && isGV(v))
#define IS_PADCONST(v) \
    ((v) && (SvREADONLY(v) || (SvIsCOW(v) && !SvLEN(v))))
/* The mortals of native code are the interpreter's, which makes room for
 * each as it comes; the C stack of native code is the host's. */
#define EXTEND_MORTAL(n) NOOP
#define PERL_STACK_OVERFLOW_CHECK() NOOP
#define dXSUB_SYS dNOOP
#define STRUCT_OFFSET(s, m) offsetof(s, m)
#define MAXARG (PL_op->op_private & OPpARG4_MASK)
#define kUNOP cUNOPx(kid)
/* perl's own malloc, which a module built with -Dmalloc=Perl_malloc
 * -Dfree=Perl_mfree (MakeMaker's PERL_EXTMALLOC_DEF) calls for all of its
 * memory: the C library's. (A macro is not expanded inside its own
 * expansion, so `free` here is the function.) A source file that does not
 * include perl.h calls them as functions, which every module has: one
 * copy in the module (weak), each calling the C library's (the names of
 * the C library's functions are the module's -D names everywhere else). */
#if defined(malloc) || defined(calloc) || defined(realloc) || defined(free)
#define GOPERL_LIBC_ALLOC_RENAMED 1
#endif
#pragma push_macro("malloc")
#pragma push_macro("calloc")
#pragma push_macro("realloc")
#pragma push_macro("free")
#undef malloc
#undef calloc
#undef realloc
#undef free
#ifdef __cplusplus
extern "C" {
#endif
#ifdef GOPERL_LIBC_ALLOC_RENAMED
/* the system's header declared them under the -D names */
void *malloc(size_t);
void *calloc(size_t, size_t);
void *realloc(void *, size_t);
void free(void *);
#endif
GOPERL_C __attribute__((weak, visibility("hidden"))) void *(Perl_malloc)(size_t n) { return malloc(n); }
GOPERL_C __attribute__((weak, visibility("hidden"))) void *(Perl_calloc)(size_t n, size_t size) {
    return calloc(n, size);
}
GOPERL_C __attribute__((weak, visibility("hidden"))) void *(Perl_realloc)(void *p, size_t n) {
    return realloc(p, n);
}
GOPERL_C __attribute__((weak, visibility("hidden"))) void (Perl_mfree)(void *p) { free(p); }
#ifdef __cplusplus
}
#endif
#pragma pop_macro("free")
#pragma pop_macro("realloc")
#pragma pop_macro("calloc")
#pragma pop_macro("malloc")
#ifndef Perl_malloc
#define Perl_malloc(n) malloc(n)
#endif
#ifndef Perl_calloc
#define Perl_calloc(n, size) calloc((n), (size))
#endif
#ifndef Perl_realloc
#define Perl_realloc(p, n) realloc((p), (n))
#endif
#define Perl_mfree(p) free(p)
/* sv_usepvn_flags: the buffer becomes the SV's string, with set magic when
 * asked for. */
#define sv_usepvn_flags(sv, p, len, flags)                              \
    ({ GOPERL_ONCE(goperl_sv_usepvn_flags_sv_, sv);                      \
       const U32 goperl_sv_usepvn_flags_f_ = (U32)(flags);               \
       sv_usepvn(goperl_sv_usepvn_flags_sv_, (p), (len));                \
       if (goperl_sv_usepvn_flags_f_ & SV_SMAGIC)                        \
           SvSETMAGIC(goperl_sv_usepvn_flags_sv_); })

/* GV_CALL selectors: the ones below FIRST_CV take a glob, the rest a CV. */
enum {
    GOPERL_GVCALL_SVN = 0,
    GOPERL_GVCALL_AVN = 1,
    GOPERL_GVCALL_HVN = 2,
    GOPERL_GVCALL_ION = 3,
    GOPERL_GVCALL_CV_SET = 4,
    GOPERL_GVCALL_FLAGS = 5,
    GOPERL_GVCALL_CVGEN = 6,
    GOPERL_GVCALL_FORM = 7,
    GOPERL_GVCALL_FULLNAME = 8,
    GOPERL_GVCALL_GP = 9,
    GOPERL_GVCALL_FILE = 10,
    GOPERL_GVCALL_LINE = 11,
    GOPERL_GVCALL_NAME_UTF8 = 12,
    GOPERL_GVCALL_SV_SET = 13,
    GOPERL_GVCALL_AV_SET = 14,
    GOPERL_GVCALL_HV_SET = 15,
    GOPERL_GVCALL_FIRST_CV = 32,
    GOPERL_GVCALL_CV_GV_SET = 32,
    GOPERL_GVCALL_CV_STASH_SET = 33,
    GOPERL_GVCALL_CV_FLAGS = 34,
    GOPERL_GVCALL_CV_OUTSIDE = 35,
    GOPERL_GVCALL_CV_PROTO = 36,
    GOPERL_GVCALL_CVGEN_SET = 16,
    GOPERL_GVCALL_STASH_SET = 17,
    GOPERL_GVCALL_IO_SET = 18
};

/* What the interpreter holds in a field that is an lvalue (see
 * GOPERL_LV_IV), and the store of one. 0, 2, 3 and 4 are GOPERL_SVSLOT_RV,
 * _IV, _UV and _NV; 2, 3 and 5 name the hash, the array and the scalar of
 * a glob for GV_PTR. */
/* A block of this module's allocators (see goperl_block_t) put in the
 * buffer slot of an SV stays the module's pointer there, as any pointer of
 * this process (a structure behind an SV), until the SV makes it its
 * string: a length (SvCUR_set, SvCUR) or the flag of a string (SvPOK_on,
 * SvPOK_only, SvFLAGS) given to it, in whatever order the module gives
 * them, or the size of its buffer (SvLEN_set, SvLEN) given to an SV that is
 * a string. Then a block of the interpreter's with the same size and
 * contents takes its place, which the interpreter owns. The size of the
 * buffer is perl's sign that the block is the SV's from then on: the
 * module's block is freed when an outermost native call of this module's
 * returns (so that a write through the module's own pointer before then
 * does not reach freed memory). Without it the module's block stays the
 * module's to free. The size given to an SV that is no string waits for
 * the SV to become one: a structure behind an SV with the size of its
 * buffer (Hash::SharedMem's handle: SvPV_set, SvLEN_set, no flag of a
 * string) stays the module's pointer, which the module goes on writing
 * through and reads back with SvPVX. The pending ones are the SVs given
 * such a block in this native call. */
typedef struct goperl_pending_block {
    struct goperl_pending_block *next;
    uint64_t tok;
    void *p;
    size_t n;
    int copied;
    int owned; /* the size of the buffer was given before the SV was a string */
} goperl_pending_block_t;
GOPERL_C __attribute__((weak, visibility("hidden"))) goperl_pending_block_t *goperl_pending_blocks_v = 0;
typedef struct goperl_handed {
    struct goperl_handed *next;
    void *p;
} goperl_handed_t;
GOPERL_C __attribute__((weak, visibility("hidden"))) goperl_handed_t *goperl_handed_v = 0;
static void goperl_blocks_handed_free(void) {
    while (goperl_handed_v) {
        goperl_handed_t *const h = goperl_handed_v;
        goperl_handed_v = h->next;
        free(h->p);
        free(h);
    }
    while (goperl_pending_blocks_v) {
        goperl_pending_block_t *const b = goperl_pending_blocks_v;
        goperl_pending_blocks_v = b->next;
        free(b);
    }
    goperl_freed_bufs_forget();
    /* a native call that died left them (one that returned took them) */
    while (goperl_pending_views_v) {
        goperl_pending_view_t *const v = goperl_pending_views_v;
        goperl_pending_views_v = v->next;
        free(v);
    }
}
static void goperl_pending_block_note(goperl_frame_t *f, uint64_t tok, const void *p) {
    size_t n = 0;
    (void)f;
    if (!goperl_block_size(p, &n)) return;
    goperl_pending_block_t *b = (goperl_pending_block_t *)calloc(1, sizeof *b);
    if (!b) return; /* stays a pointer of this process's: refused as a string */
    b->tok = tok;
    b->p = (void *)p;
    b->n = n;
    b->next = goperl_pending_blocks_v;
    goperl_pending_blocks_v = b;
}
/* The buffer of the SV tok when it is a copy the interpreter was given of a
 * buffer of native code's (goperl_pv_copy_in), 0 otherwise. What is kept
 * for the SV says so: kept[0] is native code's pointer, kept[1] the copy's
 * address while the SV's slot holds that address (kept[1] is the room of
 * the buffer only while the slot holds GOPERL_PV_HOST, which no buffer
 * is). An entry left from an SV that was at the address before counts
 * only if the SV there now has a buffer at the address of that SV's copy,
 * and a buffer native code freed in this call is no copy (see
 * goperl_freed_buf_t). */
static uint64_t goperl_pv_copy_of(goperl_frame_t *f, uint64_t tok, void **kept) {
    if (!kept || !kept[1]) return 0;
    const uint64_t pv = goperl_xs_opx(f, GOPERL_OP_SV_PVX_RAW, tok, 0, 0, 0, 0, 0);
    if (!pv || pv == GOPERL_PV_HOST || (uint64_t)(uintptr_t)kept[1] != pv) return 0;
    return goperl_freed_buf_is(pv) ? 0 : pv;
}
/* The interpreter's copy of native code's buffer p, as the buffer of the
 * SV tok: need bytes, the first ncopy of them p's, then a NUL for a string
 * (nul). The SV owns the copy (SvLEN is its size) and frees it with
 * itself. A copy the SV has already of a buffer of native code's is used
 * again, grown when it is too small: native code that points the SV at its
 * buffer again (Crypt::SecretBuffer's get magic does on every read, and
 * again after it moved the buffer) gives perl the bytes there now and frees
 * no buffer of the SV's, as on perl, where the SV owned none. 0 when there
 * is no memory for it (the SV is left as it was). */
static uint64_t goperl_pv_copy_in(goperl_frame_t *f, uint64_t tok, const void *p, size_t need,
                                  size_t ncopy, int nul) {
    void **kept = f->api->pv_host(f, tok, 1);
    if (!kept) return 0;
    uint64_t g = goperl_pv_copy_of(f, tok, kept);
    uint64_t len = g ? goperl_xs_opx(f, GOPERL_OP_SV_LEN_BUF, tok, 0, 0, 0, 0, 0) : 0;
    if (!g || len < need) {
        g = goperl_xs_opx(f, GOPERL_OP_MEM_REALLOC, g, need ? need : 1, 0, 0, 0, 0);
        if (!g) return 0;
        len = need ? need : 1;
    }
    char *const d = (char *)goperl_guest_mem(f, g);
    if (ncopy) memcpy(d, p, ncopy);
    if (nul) d[ncopy] = 0;
    goperl_xs_opx(f, GOPERL_OP_SV_SLOT_SET, tok, g, (uint64_t)(5 /* GOPERL_SVSLOT_PV */), 0, 0, 0);
    goperl_xs_opx(f, GOPERL_OP_SV_SLOT_SET, tok, len, (uint64_t)(6 /* GOPERL_SVSLOT_LEN */), 0, 0, 0);
    kept[0] = (void *)p;
    kept[1] = (void *)(uintptr_t)g;
    goperl_pv_copies_g = 1;
    return g;
}
/* The SV tok makes its pending block its string (owned: the buffer's size
 * was given): the copy is made the first time, the module's block goes
 * when the SV owns it. */
static void goperl_pending_block_take(goperl_frame_t *f, uint64_t tok, int owned) {
    goperl_pending_block_t **pp = &goperl_pending_blocks_v, *b;
    while ((b = *pp) && b->tok != tok) pp = &b->next;
    if (!b) return;
    if (!b->copied) {
        void **kept = goperl_pv_host_of(f, tok);
        const svtype t = (svtype)goperl_xs_opx(f, GOPERL_OP_SV_TYPE, tok, 0, 0, 0, 0, 0);
        if (!kept || kept[0] != b->p || t < SVt_PV || t == SVt_INVLIST || t == SVt_REGEXP ||
            (t > SVt_PVMG && t != SVt_PVLV)) {
            /* another pointer by now, or no string: the module's pointer */
            *pp = b->next;
            free(b);
            return;
        }
        if (!goperl_pv_copy_in(f, tok, b->p, b->n ? b->n : 1, b->n, 0))
            goperl_croakf(f, "Out of memory!");
        b->copied = 1;
    }
    if (owned || b->owned) {
        /* the copy is the SV's own buffer from now on, as the block was on
         * perl: no copy of native code's buffer any more */
        void **kept = f->api->pv_host(f, tok, 0);
        if (kept) kept[1] = 0;
        goperl_handed_t *h = (goperl_handed_t *)malloc(sizeof(goperl_handed_t));
        if (h) {
            goperl_block_forget(b->p, 0);
            h->p = b->p;
            h->next = goperl_handed_v;
            goperl_handed_v = h;
        }
        *pp = b->next;
        free(b);
    }
}
static void *goperl_pending_blocks_any(void) { return goperl_pending_blocks_v; }
/* Before an operation that makes an SV a string (a length, the flag of a
 * string) the SV's pending block is handed over. */
static void goperl_pending_block_string(goperl_frame_t *f, int32_t op, uint64_t a, uint64_t b) {
    int makes_string = 0;
    if (op == GOPERL_OP_SV_CUR_SET) makes_string = b != 0;
    else if (op == GOPERL_OP_SV_POK_ON) makes_string = 1;
    else if (op == GOPERL_OP_SV_FLAGS) makes_string = ((uint32_t)b & (SVf_POK | SVp_POK)) != 0;
    if (makes_string) goperl_pending_block_take(f, a, 0);
}
/* A view: native code sets the buffer slot of an SV to memory of its own
 * that no block of its allocators is (a string of a library's, with SvLEN
 * 0: SWISH-3 wraps libxml2's text so for the regex engine, MarpaX-ESLIF
 * hands the input to Perl code so), and makes the SV a string. perl reads
 * the bytes where they are; the interpreter here cannot reach them, so the
 * SV gets a copy of them, in a buffer of the interpreter's that it owns and
 * frees with the SV, when it becomes a string (POK with a length), and its
 * SvPVX is that buffer from then on.
 *
 * A copy is what perl reads only while native code leaves its bytes alone:
 * a change native code makes to its buffer afterwards would reach perl, and
 * does not reach the copy; nothing tells such a change here. And perl
 * writes into a buffer it does not own in place, through a view that is
 * not read-only (File::Map's mapping of a file); such writes would land in
 * the copy. So the scalar has to be read-only by the end of the native call,
 * or by the next operation that is not on it (MarpaX-ESLIF makes it so
 * after its length, SWISH-3 before), or it fails with the refusal as
 * before. Which scalar waits for that is kept (goperl_pv_view_g): it is
 * decided by a flag native code may set after the copy. */
GOPERL_C __attribute__((weak, visibility("hidden"))) uint64_t goperl_pv_view_g = 0;
static int goperl_pv_view_busy;
/* The operations on the scalar that finish making it, which wait for its
 * read-only flag. */
static int goperl_pv_view_building(int32_t op) {
    switch (op) {
    case GOPERL_OP_SV_FLAGS: case GOPERL_OP_SV_POK_ON: case GOPERL_OP_SV_POK_ONLY:
    case GOPERL_OP_SV_CUR_SET: case GOPERL_OP_SV_CUR: case GOPERL_OP_SV_SLOT_SET:
    case GOPERL_OP_SV_READONLY_ON: case GOPERL_OP_SV_UTF8_ON: case GOPERL_OP_SV_UTF8_OFF:
    case GOPERL_OP_SV_PVX_RAW: case GOPERL_OP_SV_LEN_BUF: case GOPERL_OP_SV_INFO:
    case GOPERL_OP_SV_TYPE: case GOPERL_OP_SV_UPGRADE: case GOPERL_OP_SV_REFCNT:
        return 1;
    }
    return 0;
}
/* The refusal of a string over native code's memory that perl could write
 * through (see goperl_pv_view_g): before an operation it dies; at the end of
 * the native call (op -1), which has returned already, the loader reports
 * it. */
static const char goperl_pv_view_refusal[] =
    "the string of a scalar cannot be in native memory on go-perl: native code set "
    "the scalar's buffer (SvPVX, SvPV_set) to an address outside the interpreter's "
    "memory and made it a string that perl can write to; gperl gives the "
    "interpreter a copy of such a string only when it is read-only";
static void goperl_pv_view_refuse(goperl_frame_t *f, int32_t op) {
    goperl_pv_view_busy = 0;
    if (op >= 0) goperl_croakf(f, "%s", goperl_pv_view_refusal);
    if (!f->failed) {
        snprintf(f->err, sizeof f->err, "%s", goperl_pv_view_refusal);
        f->failed = 1;
    }
}
static void goperl_pending_view_drop(uint64_t tok) {
    goperl_pending_view_t **pp = &goperl_pending_views_v;
    while (*pp) {
        goperl_pending_view_t *const v = *pp;
        if (v->tok == tok) {
            *pp = v->next;
            free(v);
        } else {
            pp = &v->next;
        }
    }
}
/* The SVs given native code's memory while they were strings get their
 * copies, of the length they have now, before an operation that does not
 * make them (op on a) and at the end of the native call (op -1). Their
 * making is over then, so one perl could write through is refused at once,
 * and keeps the string it had. 0 when one was refused. */
static int goperl_pending_views_take(goperl_frame_t *f, int32_t op, uint64_t a) {
    goperl_pending_view_t **pp = &goperl_pending_views_v;
    while (*pp) {
        goperl_pending_view_t *const v = *pp;
        if (op >= 0 && a == v->tok && goperl_pv_view_building(op)) {
            pp = &v->next;
            continue;
        }
        *pp = v->next;
        const uint64_t tok = v->tok, oldcur = v->oldcur;
        const void *const p = v->p;
        free(v);
        if (!(goperl_xs_opx(f, GOPERL_OP_SV_INFO, tok, 0, 0, 0, 0, 0) & GOPERL_INFO_READONLY)) {
            /* the scalar keeps the string it had */
            goperl_xs_opx(f, GOPERL_OP_SV_CUR_SET, tok, oldcur, 0, 0, 0, 0);
            goperl_pv_view_refuse(f, op);
            return 0;
        }
        const uint64_t cur = goperl_xs_opx(f, GOPERL_OP_SV_CUR, tok, 0, 0, 0, 0, 0);
        if (!goperl_pv_copy_in(f, tok, p, (size_t)cur + 1, (size_t)cur, 1)) {
            goperl_pv_view_busy = 0;
            goperl_croakf(f, "Out of memory!");
        }
    }
    return 1;
}
static void goperl_pv_view_hook(goperl_frame_t *f, int32_t op, uint64_t a, uint64_t b) {
    if (goperl_pv_view_busy) return;
    goperl_pv_view_busy = 1;
    if (goperl_pending_views_v && !goperl_pending_views_take(f, op, a)) return;
    if (goperl_pv_view_g && !(a == goperl_pv_view_g && goperl_pv_view_building(op))) {
        const uint64_t tok = goperl_pv_view_g;
        goperl_pv_view_g = 0;
        if (!(goperl_xs_opx(f, GOPERL_OP_SV_INFO, tok, 0, 0, 0, 0, 0) & GOPERL_INFO_READONLY)) {
            /* the scalar is left no string, its buffer slot native code's
             * pointer again, and the copy is freed */
            const uint64_t copy = goperl_xs_opx(f, GOPERL_OP_SV_PVX_RAW, tok, 0, 0, 0, 0, 0);
            goperl_xs_opx(f, GOPERL_OP_SV_FLAGS, tok,
                          (uint64_t)(uint32_t)(SVf_POK | SVp_POK | SVf_UTF8) << 32, 0, 0, 0, 0);
            goperl_xs_opx(f, GOPERL_OP_SV_CUR_SET, tok, 0, 0, 0, 0, 0);
            goperl_xs_opx(f, GOPERL_OP_SV_SLOT_SET, tok, GOPERL_PV_HOST, (uint64_t)(5 /* GOPERL_SVSLOT_PV */), 0, 0, 0);
            goperl_xs_opx(f, GOPERL_OP_SV_SLOT_SET, tok, 0, (uint64_t)(6 /* GOPERL_SVSLOT_LEN */), 0, 0, 0);
            {
                void **kept = f->api->pv_host(f, tok, 0);
                if (kept) kept[1] = 0; /* the room of no buffer: no copy */
            }
            if (copy && copy != GOPERL_PV_HOST) goperl_xs_opx(f, GOPERL_OP_MEM_FREE, copy, 0, 0, 0, 0, 0);
            goperl_pv_view_refuse(f, op);
            return;
        }
    }
    if (op == GOPERL_OP_SV_CUR_SET || op == GOPERL_OP_SV_POK_ON || op == GOPERL_OP_SV_POK_ONLY ||
        (op == GOPERL_OP_SV_FLAGS && ((uint32_t)b & (SVf_POK | SVp_POK)))) {
        void **kept = f->api->pv_host(f, a, 0);
        if (kept && kept[0] &&
            goperl_xs_opx(f, GOPERL_OP_SV_PVX_RAW, a, 0, 0, 0, 0, 0) == GOPERL_PV_HOST) {
            /* a string after the operation: a length while the flag of a
             * string is on, or the flag with a length; of a type that has
             * a string (the interpreter's refusal asks the same) */
            const uint64_t info = goperl_xs_opx(f, GOPERL_OP_SV_INFO, a, 0, 0, 0, 0, 0);
            const uint64_t cur = op == GOPERL_OP_SV_CUR_SET
                                     ? ((info & (GOPERL_INFO_POK | GOPERL_INFO_POKp)) ? b : 0)
                                     : goperl_xs_opx(f, GOPERL_OP_SV_CUR, a, 0, 0, 0, 0, 0);
            const uint64_t t = goperl_xs_opx(f, GOPERL_OP_SV_TYPE, a, 0, 0, 0, 0, 0);
            if (cur && t >= SVt_PV && t != SVt_PVAV && t != SVt_PVHV && t != SVt_PVCV &&
                t != SVt_PVFM && t != SVt_PVIO && t != SVt_PVGV) {
                if (!goperl_pv_copy_in(f, a, kept[0], (size_t)cur + 1, (size_t)cur, 1)) {
                    goperl_pv_view_busy = 0;
                    goperl_croakf(f, "Out of memory!");
                }
                if (!(info & GOPERL_INFO_READONLY)) goperl_pv_view_g = a;
            }
        }
    }
    goperl_pv_view_busy = 0;
}
/* Whether the SV tok is a string (the flag of one and a length) of a type
 * that has a string: the interpreter takes no pointer of this process's
 * into the buffer slot of such an SV (GOPERL_PV_HOST), as no byte of it
 * could be read. The length, 0 for no string. */
static uint64_t goperl_sv_is_string(goperl_frame_t *f, uint64_t tok) {
    const uint64_t info = goperl_xs_opx(f, GOPERL_OP_SV_INFO, tok, 0, 0, 0, 0, 0);
    if (!(info & (GOPERL_INFO_POK | GOPERL_INFO_POKp)) || (info & (GOPERL_INFO_ROK | GOPERL_INFO_ISGV)))
        return 0;
    const uint64_t t = goperl_xs_opx(f, GOPERL_OP_SV_TYPE, tok, 0, 0, 0, 0, 0);
    if (t < SVt_PV || t == SVt_PVAV || t == SVt_PVHV || t == SVt_PVCV || t == SVt_PVFM ||
        t == SVt_PVIO)
        return 0;
    return goperl_xs_opx(f, GOPERL_OP_SV_CUR, tok, 0, 0, 0, 0, 0);
}
/* Stores a pointer in the buffer slot of an SV: one into the interpreter's
 * memory as its address there, any other with the loader (a block of this
 * module's allocators waits to be handed over, goperl_pending_block_t; a
 * view gets a copy when it becomes a string, goperl_pv_view_hook).
 *
 * An SV that is a string already cannot wait for a length or the flag of a
 * string: a block gets its copy at once (Encode's Unicode.xs moves the
 * string of its result to a smaller block of its own with SvPV_set, then
 * hands the block over with SvLEN_set; Crypt::SecretBuffer points its
 * string at its buffer again on every read), other memory when its length
 * is known (goperl_pending_view_t).
 *
 * A copy the SV owned only as the copy of an earlier buffer of native
 * code's goes when native code gives the SV another buffer, and the SV
 * owns no buffer then, as it owned none on perl. */
static void goperl_sv_pvx_store(goperl_frame_t *f, uint64_t tok, const void *p) {
    uint64_t g = goperl_guest_addr(f, p);
    if (goperl_pending_views_v) goperl_pending_view_drop(tok);
    if (p && !g) {
        goperl_pending_block_note(f, tok, p);
        const uint64_t cur = goperl_sv_is_string(f, tok);
        if (cur) {
            size_t n = 0;
            if (goperl_block_size(p, &n)) {
                if (!goperl_pv_copy_in(f, tok, p, n ? n : 1, n, 0)) goperl_croakf(f, "Out of memory!");
                for (goperl_pending_block_t *b = goperl_pending_blocks_v; b; b = b->next)
                    if (b->tok == tok) b->copied = 1;
            } else {
                goperl_pending_view_t *const v = (goperl_pending_view_t *)malloc(sizeof *v);
                if (!v) goperl_croakf(f, "Out of memory!");
                v->tok = tok;
                v->p = p;
                v->oldcur = cur;
                v->next = goperl_pending_views_v;
                goperl_pending_views_v = v;
                goperl_pv_views_g = 1;
            }
            return;
        }
    }
    void **kept = f->api->pv_host(f, tok, p && !g);
    if (p && !g && !kept) goperl_croakf(f, "out of memory for the buffer slot of an SV");
    const uint64_t old = goperl_pv_copy_of(f, tok, kept);
    if (old && old != g) {
        goperl_xs_opx(f, GOPERL_OP_SV_SLOT_SET, tok, 0, (uint64_t)(6 /* GOPERL_SVSLOT_LEN */), 0, 0, 0);
        goperl_xs_opx(f, GOPERL_OP_MEM_FREE, old, 0, 0, 0, 0, 0);
        kept[1] = 0;
    }
    if (p && !g) {
        kept[0] = (void *)p;
        kept[1] = 0;
        g = GOPERL_PV_HOST;
        goperl_pv_views_g = 1;
    }
    goperl_xs_opx(f, GOPERL_OP_SV_SLOT_SET, tok, g, (uint64_t)(5 /* GOPERL_SVSLOT_PV */), 0, 0, 0);
}
/* SvLEN, read and stored: the interpreter's, except for an SV whose buffer
 * slot holds a pointer of this process's (see GOPERL_PV_HOST). Only an SV
 * the interpreter has no length for can be one. */
static uint64_t goperl_sv_len_load(goperl_frame_t *f, uint64_t tok) {
    const uint64_t len = goperl_do_op(f, GOPERL_OP_SV_LEN_BUF, tok, 0, 0, 0);
    if (len) return len;
    void **kept = goperl_pv_host_of(f, tok);
    return kept ? (uint64_t)(uintptr_t)kept[1] : 0;
}
/* The size native code gives the buffer of an SV whose buffer is the copy
 * of one of native code's (goperl_pv_copy_in) is that buffer's: the copy's
 * own stays the SV's, which frees the copy with itself. */
static uint64_t goperl_sv_is_string(goperl_frame_t *f, uint64_t tok);
static void goperl_sv_len_store(goperl_frame_t *f, uint64_t tok, uint64_t len) {
    if (goperl_pending_blocks_v && len) {
        if (goperl_sv_is_string(f, tok)) {
            goperl_pending_block_take(f, tok, 1);
        } else {
            for (goperl_pending_block_t *b = goperl_pending_blocks_v; b; b = b->next)
                if (b->tok == tok) b->owned = 1;
        }
    }
    void **kept = goperl_pv_host_of(f, tok);
    if (kept) {
        kept[1] = (void *)(uintptr_t)len;
        return;
    }
    if (goperl_pv_copy_of(f, tok, f->api->pv_host(f, tok, 0))) return;
    goperl_xs_opx(f, GOPERL_OP_SV_SLOT_SET, tok, len, (uint64_t)(6 /* GOPERL_SVSLOT_LEN */), 0, 0, 0);
}
static uint64_t goperl_cvfield_load(goperl_frame_t *f, uint64_t tok, int field);
static void goperl_cvfield_store(goperl_frame_t *f, uint64_t tok, int field, uint64_t val);
/* CvXSUB of an XSUB of the interpreter's (B's, POSIX's): a function of
 * this process that calls the sub through the interpreter, as calling the
 * XSUB's function itself would (Runops::Trace in B::OPCheck calls
 * B::UNOP::first so). The mark and the arguments are the caller's, the
 * results land where the XSUB would leave them, and the call is in the
 * caller's context (a list one in void context, where an XSUB still
 * leaves what it returns). */
static void goperl_interp_xsub(pTHX_ CV *cv) {
    goperl_frame_t *const f = _gof;
    (void)goperl_do_call(f, (SV *)cv, 0, GIMME_V == G_SCALAR ? G_SCALAR : G_LIST);
}

static void *goperl_sv_any_value(goperl_frame_t *f, SV *sv);
static uint64_t goperl_lv_load_field(goperl_frame_t *f, uint64_t tok, int field) {
    if (GOPERL_LV_IS_IOWORD(field)) { /* FIELD 24 */
        const int w = field - GOPERL_LV_IOWORD;
        const uint64_t v = goperl_do_op3(f, GOPERL_OP_FIELD, tok, (uint64_t)w, (uint64_t)(24), 0, 0);
        if (GOPERL_IOWORD_IS_NAME(w)) return v ? (uint64_t)(uintptr_t)goperl_guest_mem(f, v) : 0;
        return v;
    }
    switch (field) {
    case GOPERL_LV_IV:
        return goperl_do_op3(f, GOPERL_OP_SV_SLOT_GET, tok, 0, (uint64_t)(2), 0, 0);
    case GOPERL_LV_UV:
        return goperl_do_op3(f, GOPERL_OP_SV_SLOT_GET, tok, 0, (uint64_t)(3), 0, 0);
    case GOPERL_LV_NV:
        return goperl_do_op3(f, GOPERL_OP_SV_SLOT_GET, tok, 0, (uint64_t)(4), 0, 0);
    case GOPERL_LV_RV:
        /* the slot, as perl's SvRV reads it whatever the flags say (not
         * "the target if SvROK", which is SvROK's question) */
        return goperl_do_op3(f, GOPERL_OP_SV_SLOT_GET, tok, 0, (uint64_t)(0 /* GOPERL_SVSLOT_RV */), 0, 0);
    case GOPERL_LV_CVXSUB: {
        void *const fn = goperl_api_cv_xsub(f, tok);
        if (fn || !(goperl_do_op(f, GOPERL_OP_CV_INFO, tok, 0, 0, 0) & 1)) return (uint64_t)(uintptr_t)fn;
        return (uint64_t)(uintptr_t)(void *)goperl_interp_xsub;
    }
    case GOPERL_LV_GVSTASH:
        return goperl_do_op(f, GOPERL_OP_GV_PTR, tok, 0, 0, 0);
    case GOPERL_LV_GVHV:
        return goperl_do_op(f, GOPERL_OP_GV_PTR, tok, 2, 0, 0);
    case GOPERL_LV_GVAV:
        return goperl_do_op(f, GOPERL_OP_GV_PTR, tok, 3, 0, 0);
    case GOPERL_LV_GVSV:
        return goperl_do_op(f, GOPERL_OP_GV_PTR, tok, 5, 0, 0);
    case GOPERL_LV_REFCNT:
        return goperl_do_op(f, GOPERL_OP_SV_REFCNT, tok, 0, 0, 0);
    case GOPERL_LV_PVX:
        return (uint64_t)(uintptr_t)goperl_sv_pvx(f, (SV *)(uintptr_t)tok);
    case GOPERL_LV_LEN:
        return goperl_sv_len_load(f, tok);
    case GOPERL_LV_AVFILLP: /* FIELD 6 */
        return goperl_do_op3(f, GOPERL_OP_FIELD, tok, 0, (uint64_t)(6), 0, 0);
    case GOPERL_LV_LVTYPE: /* FIELD 9 */
        return goperl_do_op3(f, GOPERL_OP_FIELD, tok, 0, (uint64_t)(9), 0, 0);
    case GOPERL_LV_LVTARG: /* FIELD 10 */
        return goperl_do_op3(f, GOPERL_OP_FIELD, tok, 0, (uint64_t)(10), 0, 0);
    case GOPERL_LV_CVOUTSIDE:
        return goperl_do_op3(f, GOPERL_OP_GV_CALL, tok, 0, (uint64_t)(GOPERL_GVCALL_CV_OUTSIDE), 0, 0);
    case GOPERL_LV_CVFLAGS:
    case GOPERL_LV_CVROOT:
    case GOPERL_LV_CVSTART:
    case GOPERL_LV_CVOUTSIDE_SEQ:
    case GOPERL_LV_CVPADLIST:
    case GOPERL_LV_CVDEPTH:
    case GOPERL_LV_CVFILE:
    case GOPERL_LV_OBJMAXFIELD:
        return goperl_cvfield_load(f, tok, field);
    case GOPERL_LV_FLAGS:
        return goperl_do_op(f, GOPERL_OP_SV_FLAGS, tok, 0, 0, 0);
    case GOPERL_LV_STASH:
        return goperl_do_op(f, GOPERL_OP_SV_STASH, tok, 0, 0, 0);
    case GOPERL_LV_GVFLAGS:
        return goperl_do_op3(f, GOPERL_OP_GV_CALL, tok, 0, (uint64_t)(GOPERL_GVCALL_FLAGS), 0, 0);
    case GOPERL_LV_HVRITER: /* FIELD 2 */
        return goperl_do_op3(f, GOPERL_OP_FIELD, tok, 0, (uint64_t)(2), 0, 0);
    case GOPERL_LV_IOIFP:
        return goperl_do_op3(f, GOPERL_OP_PERLIO, tok, 0, (uint64_t)(GOPERL_PIO_IFP), 0, 0);
    case GOPERL_LV_IOOFP:
        return goperl_do_op3(f, GOPERL_OP_PERLIO, tok, 0, (uint64_t)(GOPERL_PIO_OFP), 0, 0);
    case GOPERL_LV_IOTYPE:
        return (uint64_t)(unsigned char)goperl_do_op3(f, GOPERL_OP_SV_CALL, tok, 0, (uint64_t)(GOPERL_SVCALL_IO_TYPE), 0, 0);
    case GOPERL_LV_IOFLAGS:
        return (uint64_t)(U8)goperl_do_op3(f, GOPERL_OP_SV_CALL, tok, 0, (uint64_t)(GOPERL_SVCALL_IO_FLAGS), 0, 0);
    case GOPERL_LV_GVIO:
        return goperl_do_op(f, GOPERL_OP_GV_PTR, tok, 6, 0, 0);
    case GOPERL_LV_HVEITER: /* FIELD 4: the entry itself */
        return (uint64_t)(uintptr_t)goperl_he_at(
            f, goperl_do_op3(f, GOPERL_OP_FIELD, tok, 0, (uint64_t)(4), 0, 0));
    case GOPERL_LV_SVANY:
        return (uint64_t)(uintptr_t)goperl_sv_any_value(f, (SV *)(uintptr_t)tok);
    default:
        return goperl_do_op3(f, GOPERL_OP_GV_CALL, tok, 0, (uint64_t)(GOPERL_GVCALL_CVGEN), 0, 0);
    }
}
static void goperl_lv_store_field(goperl_frame_t *f, uint64_t tok, int field,
                                  uint64_t val) {
    if (GOPERL_LV_IS_IOWORD(field)) { /* FIELD 25, which in the high half */
        const int w = field - GOPERL_LV_IOWORD;
        const uint64_t c = (uint64_t)(25) | ((uint64_t)(uint32_t)w << 32);
        if (GOPERL_IOWORD_IS_NAME(w) && val) {
            const char *name = (const char *)(uintptr_t)val;
            const uint64_t g = goperl_guest_addr(f, name);
            if (g) /* a name in the interpreter's memory already */
                goperl_xs_opx(f, GOPERL_OP_FIELD, tok, g, c, 0, 0, 0);
            else /* a copy of the text: the interpreter's from now on */
                goperl_xs_opx(f, GOPERL_OP_FIELD, tok, 0, c, name, __builtin_strlen(name) + 1, 0);
        } else {
            goperl_xs_opx(f, GOPERL_OP_FIELD, tok, val, c, 0, 0, 0);
        }
        return;
    }
    switch (field) {
    case GOPERL_LV_IV:
        goperl_xs_opx(f, GOPERL_OP_SV_SLOT_SET, tok, val, (uint64_t)(2), 0, 0, 0);
        break;
    case GOPERL_LV_UV:
        goperl_xs_opx(f, GOPERL_OP_SV_SLOT_SET, tok, val, (uint64_t)(3), 0, 0, 0);
        break;
    case GOPERL_LV_NV:
        goperl_xs_opx(f, GOPERL_OP_SV_SLOT_SET, tok, val, (uint64_t)(4), 0, 0, 0);
        break;
    case GOPERL_LV_RV:
        goperl_xs_opx(f, GOPERL_OP_SV_SLOT_SET, tok, val, 0 /* GOPERL_SVSLOT_RV */, 0, 0, 0);
        break;
    case GOPERL_LV_CVXSUB:
        /* the sub would call itself through the interpreter, for ever */
        if (val == (uint64_t)(uintptr_t)(void *)goperl_interp_xsub)
            goperl_croakf(f, "go-perl gives native code no access to this interpreter "
                             "internal: the C function of an XSUB of the interpreter's, "
                             "given to another sub");
        goperl_api_cv_set_xsub(f, tok, (void *)(uintptr_t)val);
        break;
    case GOPERL_LV_GVSTASH:
        goperl_xs_opx(f, GOPERL_OP_GV_CALL, tok, val, (uint64_t)(GOPERL_GVCALL_STASH_SET), 0, 0, 0);
        break;
    case GOPERL_LV_GVHV:
        goperl_xs_opx(f, GOPERL_OP_GV_CALL, tok, val, (uint64_t)(GOPERL_GVCALL_HV_SET), 0, 0, 0);
        break;
    case GOPERL_LV_GVAV:
        goperl_xs_opx(f, GOPERL_OP_GV_CALL, tok, val, (uint64_t)(GOPERL_GVCALL_AV_SET), 0, 0, 0);
        break;
    case GOPERL_LV_GVSV:
        goperl_xs_opx(f, GOPERL_OP_GV_CALL, tok, val, (uint64_t)(GOPERL_GVCALL_SV_SET), 0, 0, 0);
        break;
    case GOPERL_LV_PVX:
        goperl_sv_pvx_store(f, tok, (const void *)(uintptr_t)val);
        break;
    case GOPERL_LV_LEN:
        goperl_sv_len_store(f, tok, val);
        break;
    case GOPERL_LV_AVFILLP: /* FIELD 7 */
        goperl_xs_opx(f, GOPERL_OP_FIELD, tok, val, (uint64_t)(7), 0, 0, 0);
        break;
    case GOPERL_LV_LVTYPE: /* FIELD 18 */
        goperl_xs_opx(f, GOPERL_OP_FIELD, tok, (uint64_t)(unsigned char)val, (uint64_t)(18), 0, 0, 0);
        break;
    case GOPERL_LV_LVTARG: /* FIELD 19 */
        goperl_xs_opx(f, GOPERL_OP_FIELD, tok, val, (uint64_t)(19), 0, 0, 0);
        break;
    case GOPERL_LV_SVANY: {
        /* the body pointer is a pointer of the interpreter's: an address in
         * its memory, as &sv->sv_u and the layouts give one; the head is
         * the interpreter's memory too, and takes it as it is */
        uint64_t *const head = (uint64_t *)goperl_guest_mem(f, tok);
        if (!head || val >= (uint64_t)*f->memsize)
            goperl_croakf(f, "SvANY: native code stored a body pointer that is not an address "
                             "of the interpreter's memory; go-perl cannot give a value a body "
                             "of native code's");
        head[0] = val;
        break;
    }
    case GOPERL_LV_REFCNT: /* FIELD 0: the count itself, nothing is freed */
        goperl_xs_opx(f, GOPERL_OP_FIELD, tok, (uint32_t)val, 0 /* REFCNT_SET */, 0, 0, 0);
        break;
    case GOPERL_LV_CVOUTSIDE: /* FIELD 16: the plain store */
        goperl_xs_opx(f, GOPERL_OP_FIELD, tok, val, (uint64_t)(16), 0, 0, 0);
        break;
    case GOPERL_LV_CVFLAGS:
    case GOPERL_LV_CVROOT:
    case GOPERL_LV_CVSTART:
    case GOPERL_LV_CVOUTSIDE_SEQ:
    case GOPERL_LV_CVPADLIST:
    case GOPERL_LV_CVDEPTH:
    case GOPERL_LV_CVFILE:
    case GOPERL_LV_OBJMAXFIELD:
        goperl_cvfield_store(f, tok, field, val);
        break;
    case GOPERL_LV_FLAGS: /* every bit cleared, the word's set */
        goperl_xs_op(f, GOPERL_OP_SV_FLAGS, tok,
                     ((uint64_t)0xffffffffu << 32) | (uint32_t)val, 0, 0);
        break;
    case GOPERL_LV_STASH:
        goperl_xs_opx(f, GOPERL_OP_SV_SLOT_SET, tok, val, (uint64_t)(GOPERL_SVSLOT_STASH), 0, 0, 0);
        break;
    case GOPERL_LV_GVFLAGS:
        /* b = the flags to set, hi = the ones to clear (every one) */
        goperl_xs_opx(f, GOPERL_OP_GV_CALL, tok, (uint32_t)val,
                      GOPERL_GVCALL_FLAGS | ((uint64_t)0xffffffffu << 32), 0, 0, 0);
        break;
    case GOPERL_LV_HVRITER: /* FIELD 3 */
        goperl_xs_opx(f, GOPERL_OP_FIELD, tok, (uint64_t)(int64_t)(int32_t)(uint32_t)val, (uint64_t)(3), 0, 0, 0);
        break;
    case GOPERL_LV_HVEITER: /* FIELD 5: the entry's address */
        goperl_xs_opx(f, GOPERL_OP_FIELD, tok,
                      goperl_he_addr(f, (const HE *)(uintptr_t)val, "HvEITER"), (uint64_t)(5),
                      0, 0, 0);
        break;
    case GOPERL_LV_IOIFP: /* FIELD 20 */
    case GOPERL_LV_IOOFP: /* FIELD 21 */
        goperl_xs_opx(f, GOPERL_OP_FIELD, tok, goperl_pio_tok(f, (const void *)(uintptr_t)val), (uint64_t)(field == GOPERL_LV_IOIFP ? 20 : 21), 0, 0, 0);
        break;
    case GOPERL_LV_IOTYPE: /* FIELD 22 */
        goperl_xs_opx(f, GOPERL_OP_FIELD, tok, (uint64_t)(unsigned char)val, (uint64_t)(22), 0, 0, 0);
        break;
    case GOPERL_LV_IOFLAGS: /* FIELD 23 */
        goperl_xs_opx(f, GOPERL_OP_FIELD, tok, (uint64_t)(U8)val, (uint64_t)(23), 0, 0, 0);
        break;
    case GOPERL_LV_GVIO:
        goperl_xs_opx(f, GOPERL_OP_GV_CALL, tok, val, (uint64_t)(GOPERL_GVCALL_IO_SET), 0, 0, 0);
        break;
    default:
        goperl_xs_opx(f, GOPERL_OP_GV_CALL, tok, (uint32_t)val, (uint64_t)(GOPERL_GVCALL_CVGEN_SET), 0, 0, 0);
        break;
    }
}

/* SV_STR selectors. */
enum {
    GOPERL_SVSTR_CATPVN = 0,
    GOPERL_SVSTR_NEWSVRV = 1,
    GOPERL_SVSTR_DOES = 2,
    GOPERL_SVSTR_FETCHMETH = 3,
    GOPERL_SVSTR_FETCHMETH_AUTOLOAD = 4,
    GOPERL_SVSTR_FETCHMETHOD = 5,
    GOPERL_SVSTR_GV_FETCHPVN = 6,
    GOPERL_SVSTR_GET_CVN = 7,
    GOPERL_SVSTR_WHICHSIG = 8,
    GOPERL_SVSTR_COP_HINT = 9,
    GOPERL_SVSTR_SETENV = 10,
    /* 11 to 14: the formatter's (perl.h) */
    GOPERL_SVSTR_TAINT_PROPER = 15,
    /* 16: the formatter's (perl.h) */
    GOPERL_SVSTR_GV_NAME_SET = 17,
    GOPERL_SVSTR_HV_NAME_SET = 18
};

/* PLVAR_GET/SET ids beyond the shadowed ones of perl.h. */
#define GOPERL_PL_ERRGV 29
#define GOPERL_PL_DEFSTASH 30
#define GOPERL_PL_INCGV 31
#define GOPERL_PL_STDERRGV 32
#define GOPERL_PL_DEFOUTGV 33
#define GOPERL_PL_RS 34
#define GOPERL_PL_IN_EVAL 35
#define GOPERL_PL_PHASE 36
#define GOPERL_PL_TAINTING 37
#define GOPERL_PL_TAINTED 38
#define GOPERL_PL_DIRTY 39
#define GOPERL_PL_SAVESTACK_IX 40
#define GOPERL_PL_TMPS_IX 41
#define GOPERL_PL_STDINGV 42
#define GOPERL_PL_ARGVGV 43
#define GOPERL_PL_ENVGV 44
#define GOPERL_PL_LOCALIZING 45
#define GOPERL_PL_BEGINAV 46
#define GOPERL_PL_UNITCHECKAV 47
#define GOPERL_PL_SV_ZERO 48
#define GOPERL_PL_SV_PLACEHOLDER 49
#define GOPERL_PL_STRTAB 50
#define GOPERL_PL_STATGV 51
#define GOPERL_PL_TMPS_FLOOR 52
#define GOPERL_PL_ORIGFILENAME 53
#define GOPERL_PL_UNSAFE 54
#define GOPERL_PL_DESTRUCT_LEVEL 55
#define GOPERL_PL_LASTGV 56
#define GOPERL_PL_COP_SEQMAX 57
#define GOPERL_PL_ERRSV 58
#define GOPERL_PL_DEFSV 59
#define GOPERL_PL_ERRNO 60

#define GOPERL_SVCALL(fn, sv, b)                                           \
    goperl_do_op3(_gof, GOPERL_OP_SV_CALL, GOPERL_TOK(sv), (uint64_t)(b), (uint64_t)(fn), 0, 0)
#define GOPERL_GVCALL(fn, gv, b)                                           \
    goperl_do_op3(_gof, GOPERL_OP_GV_CALL, GOPERL_TOK(gv), (uint64_t)(b), (uint64_t)(fn), 0, 0)
#define GOPERL_SVSTR(fn, sv, flags, s, len)                                \
    ({ const uint64_t goperl_SVSTR_len_ = (uint64_t)(len);                 \
       goperl_do_op3(_gof, GOPERL_OP_SV_STR, GOPERL_TOK(sv), ((uint64_t)(U32)(flags) << 32) | (uint32_t)goperl_SVSTR_len_, (uint64_t)(fn), (s), goperl_SVSTR_len_); })
/* The calls of two SVs with flags: b is the other SV, the flags are the
 * 32-bit value in c's high half (hi). */
#define GOPERL_SVCALL2(fn, sv, other, hi)                                    \
    goperl_do_op3(_gof, GOPERL_OP_SV_CALL, GOPERL_TOK(sv), GOPERL_TOK(other), \
                  (uint64_t)(uint32_t)(fn) | ((uint64_t)(U32)(hi) << 32), 0, 0)
#define GOPERL_GVCALL2(fn, gv, b, hi)                                        \
    goperl_do_op3(_gof, GOPERL_OP_GV_CALL, GOPERL_TOK(gv), (uint64_t)(b),     \
                  (uint64_t)(uint32_t)(fn) | ((uint64_t)(U32)(hi) << 32), 0, 0)
#define GOPERL_PLVAR(id) \
    goperl_do_op(_gof, GOPERL_OP_PLVAR_GET, (uint64_t)(id), 0, 0, 0)
#define GOPERL_PLVAR_STORE(id, v) \
    ((void)goperl_do_op(_gof, GOPERL_OP_PLVAR_SET, (uint64_t)(id), (uint64_t)(v), 0, 0))

/* ---- compiler and configuration vocabulary ------------------------------ */

#ifdef __cplusplus
#define START_EXTERN_C extern "C" {
#define END_EXTERN_C }
#else
#define START_EXTERN_C
#define END_EXTERN_C
#endif
#define EXT extern
#define dEXT
#define EXTCONST extern const
#define dEXTCONST const
#define PERL_CALLCONV
#define PERL_CALLCONV_NO_RET
#define PERL_STATIC_INLINE static inline
#define PERL_STATIC_INLINE_NO_RET static inline
#define PERL_STATIC_NO_RET static
#define PERL_STATIC_FORCE_INLINE static inline __attribute__((always_inline))
#define __attribute__noreturn__ __attribute__((noreturn))
#define __attribute__unused__ __attribute__((unused))
#define __attribute__pure__ __attribute__((pure))
#define __attribute__malloc__ __attribute__((__malloc__))
#define __attribute__nonnull__(a) __attribute__((nonnull(a)))
#define __attribute__warn_unused_result__ __attribute__((warn_unused_result))
#define __attribute__deprecated__ __attribute__((deprecated))
#define __attribute__always_inline__ __attribute__((always_inline))
#define HASATTRIBUTE_UNUSED
#define HASATTRIBUTE_NORETURN
#define PERL_USE_GCC_BRACE_GROUPS
#define ASSUME(x) assert(x)
#define NOT_REACHED assert(0)
#define NORETURN_FUNCTION_END NOT_REACHED
/* The expression is evaluated, and its value dropped: it is how code calls
 * a function for its effect when the compiler insists on the result being
 * used (PERL_UNUSED_RESULT(Gconvert(nv, NV_DIG, 0, buf))). */
#define PERL_UNUSED_RESULT(v)                         \
    do {                                              \
        __typeof__(v) goperl_unused_result_ = (v);    \
        (void)sizeof(goperl_unused_result_);          \
    } while (0)
#define PERL_BITFIELD8 U8
#define PERL_BITFIELD16 U16
#define PERL_BITFIELD32 U32
#define STATIC_ASSERT_DECL(cond) _Static_assert(cond, #cond)
#define STATIC_ASSERT_STMT(cond) STATIC_ASSERT_DECL(cond)
#define NULLOK
#define NN
#define NNDD
#define NULLOKDD

#define PERL_VERSION_EQ(r, v, s) \
    (PERL_DECIMAL_VERSION == PERL_VERSION_DECIMAL(r, v, s))
#define PERL_VERSION_NE(r, v, s) \
    (PERL_DECIMAL_VERSION != PERL_VERSION_DECIMAL(r, v, s))

/* The host's C geometry. IVSIZE/UVSIZE and the IV/UV limits (perl.h)
 * describe the interpreter's integers, because they decide what a value
 * can hold once it is stored in an SV; those are 64 bits wide as well. */
/* What config.h says about the host: the headers and library functions a
 * unix C library always has. */
#define I_STDBOOL
#define I_STDLIB
#define I_STRING
#define I_UNISTD
#define I_STDINT
#define I_INTTYPES
#define I_LIMITS
#define I_FLOAT
#define I_MATH
#define I_FCNTL
#define I_LOCALE
#define I_TIME
#define I_SYS_TIME
#define I_SYS_TYPES
#define I_SYS_STAT
#define I_SYS_PARAM
#define I_SYS_IOCTL
#define I_SYS_WAIT
#define I_SYS_SOCKET
#define I_SYS_SELECT
#define I_SYS_UN
#define I_SYS_RESOURCE
#define I_NETINET_IN
#define I_ARPA_INET
#define I_NETDB
#define I_DIRENT
#define I_DLFCN
#define I_PWD
#define I_GRP
#define I_TERMIOS
#define I_POLL
#define HAS_LONG_LONG
#define HAS_LONG_DOUBLE
#define HAS_STRERROR
#define HAS_SNPRINTF
#define HAS_VSNPRINTF
#define HAS_STRTOLL
#define HAS_STRTOULL
#define HAS_STRTOL
#define HAS_STRTOUL
#define HAS_STRTOD
#define HAS_ISNAN
#define HAS_ISINF
#define HAS_ISFINITE
#define HAS_SELECT
#define HAS_POLL
#define HAS_SYSCONF
#define HAS_GETPAGESIZE
#define HAS_UNAME
#define HAS_GETHOSTNAME
#define HAS_TIMES
#define HAS_USLEEP
#define HAS_NANOSLEEP
#define HAS_MKSTEMP
#define HAS_MMAP
#define HAS_MUNMAP
#define HAS_MSYNC
#define HAS_SIGACTION
#define HAS_FCNTL
#define HAS_MEMCPY
#define HAS_MEMMOVE
#define HAS_MEMSET
#define HAS_MEMCMP
#define HAS_MEMCHR
#define STDC_HEADERS

#define PTRSIZE __SIZEOF_POINTER__
#define LONGSIZE __SIZEOF_LONG__
#define INTSIZE __SIZEOF_INT__
#define SHORTSIZE __SIZEOF_SHORT__
#define NVSIZE 8
#define NVTYPE double
#define IVTYPE int64_t
#define UVTYPE uint64_t
#define I8TYPE int8_t
#define U8TYPE uint8_t
#define I16TYPE int16_t
#define U16TYPE uint16_t
#define I32TYPE int32_t
#define U32TYPE uint32_t
#define I64TYPE int64_t
#define U64TYPE uint64_t
#define HAS_QUAD
#define Quad_t int64_t
#define Uquad_t uint64_t
typedef int64_t I64;
typedef uint64_t U64;
#define PTRV uintptr_t
/* The order of the bytes of a UV, as perl's config.h has it. */
#if __BYTE_ORDER__ == __ORDER_LITTLE_ENDIAN__
#define BYTEORDER 0x12345678
#else
#define BYTEORDER 0x87654321
#endif
#define MEM_SIZE size_t
#define Malloc_t void *
#define Free_t void
#define Signal_t void
/* A signal handler as perl's own are: one argument (perl does not use
 * three-argument handlers). */
typedef Signal_t (*Sighandler_t)(int);
typedef Signal_t (*Sighandler1_t)(int);
typedef Signal_t (*Sighandler3_t)(int, siginfo_t *, void *);
#define Uid_t uid_t
#define Gid_t gid_t
#define Mode_t mode_t
#define Stat(fname, bufptr) stat((fname), (bufptr))
#define Fstat(fd, bufptr) fstat((fd), (bufptr))
#define Strtol strtol
#define Strtoul strtoul
#define Atol atol
#define Atoul(s) strtoul((s), 0, 10)
#define Strtod strtod
#define Null(type) ((type)0)
#define Nullop ((OP *)0)
#define PERL_INT_MAX INT_MAX
#define PERL_INT_MIN INT_MIN
#define PERL_UINT_MAX UINT_MAX
#define PERL_UINT_MIN 0U
#define PERL_SHORT_MAX SHRT_MAX
#define PERL_SHORT_MIN SHRT_MIN
#define PERL_USHORT_MAX USHRT_MAX
#define PERL_USHORT_MIN 0
#define PERL_UCHAR_MAX UCHAR_MAX
#define PERL_UCHAR_MIN 0
#define PERL_ULONG_MIN 0UL
#define PERL_UQUAD_MIN 0
#define PERL_LONG_MAX LONG_MAX
#define PERL_LONG_MIN LONG_MIN
#define PERL_ULONG_MAX ULONG_MAX
#define PERL_QUAD_MAX INT64_MAX
#define PERL_QUAD_MIN INT64_MIN
#define PERL_UQUAD_MAX UINT64_MAX
#define SSize_t_MAX ((SSize_t)(~(size_t)0 >> 1))
#define MEM_SIZE_MAX (~(size_t)0)
#define NV_INF ((NV)INFINITY)
#define NV_NAN ((NV)NAN)
#define NV_MANT_DIG DBL_MANT_DIG
#define NV_MAX_EXP DBL_MAX_EXP
#define NV_MIN_EXP DBL_MIN_EXP
#define NV_EPSILON DBL_EPSILON
#include <limits.h>

#define PTR2nat(p) ((PTRV)PTR2IV(p))
#define PTR2NV(p) ((NV)PTR2IV(p))
#define DPTR2FPTR(t, p) ((t)PTR2nat(p))
#define FPTR2DPTR(t, p) ((t)PTR2nat(p))

/* The saturating NV casts of numeric.c (Perl_cast_iv, Perl_cast_uv). The
 * limits are written the way perl.h writes IV_MAX_P1 and UV_MAX_P1: an NV
 * cannot hold IV_MAX + 1 computed in integers, but it holds twice the half
 * exactly. */
#define GOPERL_IV_MAX_P1 ((NV)((IV_MAX >> 1) + 1) * 2.0)
#define GOPERL_UV_MAX_P1 ((NV)((UV_MAX >> 1) + 1) * 2.0)
static inline IV goperl_cast_iv(NV f) {
    if (f < GOPERL_IV_MAX_P1) return f < (NV)IV_MIN ? IV_MIN : (IV)f;
    if (f < GOPERL_UV_MAX_P1) return (IV)(UV)f;
    return f > 0 ? (IV)UV_MAX : 0; /* NaN */
}
static inline UV goperl_cast_uv(NV f) {
    if (f < 0.0) return f < (NV)IV_MIN ? (UV)IV_MIN : (UV)(IV)f;
    if (f < GOPERL_UV_MAX_P1) return (UV)f;
    return f > 0 ? UV_MAX : 0; /* NaN */
}
#define I_V(n) goperl_cast_iv((NV)(n))
#define U_V(n) goperl_cast_uv((NV)(n))
#define I_32(n) ((I32)goperl_cast_iv((NV)(n)))
#define U_32(n) ((U32)goperl_cast_uv((NV)(n)))
#define U_I(n) ((unsigned int)goperl_cast_uv((NV)(n)))
#define U_S(n) ((U16)goperl_cast_uv((NV)(n)))

#define Perl_isnan(x) isnan(x)
#define Perl_isinf(x) isinf(x)
#define Perl_isfinite(x) isfinite(x)
#define Perl_signbit(x) signbit(x)
#define Perl_frexp(x, y) frexp((x), (y))
#define Perl_ldexp(x, y) ldexp((x), (y))
#define Perl_log(x) log(x)
#define Perl_log10(x) log10(x)
#define Perl_exp(x) exp(x)
#define Perl_sqrt(x) sqrt(x)
#define Perl_sin(x) sin(x)
#define Perl_cos(x) cos(x)
#define Perl_atan2(x, y) atan2((x), (y))
#define PERL_ABS(x) ((x) < 0 ? -(x) : (x))
#define PERL_MAX(a, b) ((a) > (b) ? (a) : (b))
#define PERL_MIN(a, b) ((a) < (b) ? (a) : (b))
#define C_ARRAY_END(a) ((a) + (sizeof(a) / sizeof((a)[0])))

/* A perl built without ithreads compiles its locking macros away; the
 * embedded interpreter is such a perl. */
#define MUTEX_INIT(m) NOOP
#define MUTEX_LOCK(m) NOOP
#define MUTEX_UNLOCK(m) NOOP
#define MUTEX_DESTROY(m) NOOP
#define COND_INIT(c) NOOP
#define COND_SIGNAL(c) NOOP
#define COND_BROADCAST(c) NOOP
#define COND_WAIT(c, m) NOOP
#define COND_DESTROY(c) NOOP
#define OP_REFCNT_LOCK NOOP
#define OP_REFCNT_UNLOCK NOOP
#define OP_CHECK_MUTEX_LOCK NOOP
#define OP_CHECK_MUTEX_UNLOCK NOOP
#define SvLOCK(sv) NOOP
#define SvUNLOCK(sv) NOOP
#define SvSHARE(sv) NOOP
#define LOCK_DOLLARZERO_MUTEX NOOP
#define UNLOCK_DOLLARZERO_MUTEX NOOP

#define dTHXR dTHX
#define aTHXR aTHX
#define aTHXR_ aTHX_
#define dTHXoa(x) dTHXa(x)
#define PERL_SET_CONTEXT(i) NOOP
#define PERL_SET_THX(i) NOOP
/* perl without MULTIPLICITY has no context to get (DBI prints it, and
 * Net::DNS::Native compares the one it kept with it) */
#define PERL_GET_THX ((void *)NULL)
#define PERL_GET_INTERP PERL_GET_CONTEXT
#define dSAVEDERRNO int goperl_saved_errno
#define dSAVE_ERRNO int goperl_saved_errno = errno
#define SAVE_ERRNO (goperl_saved_errno = errno)
#define RESTORE_ERRNO (errno = goperl_saved_errno)

/* ---- memory (handy.h) ---------------------------------------------------- */

#define MEM_WRAP_CHECK(n, t)                                              \
    ((void)((sizeof(t) > 1 && (size_t)(n) > MEM_SIZE_MAX / sizeof(t))     \
                ? (goperl_croakf(_gof, "%s", "panic: memory wrap"), 0)    \
                : 0))
#define MEM_WRAP_CHECK_(n, t) MEM_WRAP_CHECK(n, t),
#define Renewc(v, n, t, c) \
    ({ __typeof__(v) *const goperl_Renewc_v_ = &(v); \
       GOPERL_ONCE(goperl_Renewc_n_, n); \
       *goperl_Renewc_v_ = (MEM_WRAP_CHECK_(goperl_Renewc_n_, t)(c *)goperl_saferealloc((void *)*goperl_Renewc_v_, (size_t)goperl_Renewc_n_ * sizeof(t))); })
#define PoisonWith(d, n, t, b) \
    ({ GOPERL_ONCE(goperl_PoisonWith_n_, n); \
       (MEM_WRAP_CHECK_(goperl_PoisonWith_n_, t)(void) memset((char *)(d), (U8)(b), (size_t)goperl_PoisonWith_n_ * sizeof(t))); })
#define PoisonNew(d, n, t) PoisonWith(d, n, t, 0xAB)
#define PoisonFree(d, n, t) PoisonWith(d, n, t, 0xEF)
#define Poison(d, n, t) PoisonFree(d, n, t)
#define CopyD(s, d, n, t) \
    ({ GOPERL_ONCE(goperl_CopyD_n_, n); \
       (MEM_WRAP_CHECK_(goperl_CopyD_n_, t) memcpy((char *)(d), (const char *)(s), (size_t)goperl_CopyD_n_ * sizeof(t))); })
#define MoveD(s, d, n, t) \
    ({ GOPERL_ONCE(goperl_MoveD_n_, n); \
       (MEM_WRAP_CHECK_(goperl_MoveD_n_, t) memmove((char *)(d), (const char *)(s), (size_t)goperl_MoveD_n_ * sizeof(t))); })
#define ZeroD(d, n, t) \
    ({ GOPERL_ONCE(goperl_ZeroD_n_, n); \
       (MEM_WRAP_CHECK_(goperl_ZeroD_n_, t) memset((char *)(d), 0, (size_t)goperl_ZeroD_n_ * sizeof(t))); })
#define safesysmalloc(n) malloc((size_t)(n))
#define safesyscalloc(n, sz) calloc((size_t)(n), (size_t)(sz))
#define safesysrealloc(p, n) realloc((void *)(p), (size_t)(n))
#define safesysfree(p) free((void *)(p))
#define PerlMem_malloc(n) malloc((size_t)(n))
#define PerlMem_calloc(n, sz) calloc((size_t)(n), (size_t)(sz))
#define PerlMem_realloc(p, n) realloc((void *)(p), (size_t)(n))
#define PerlMem_free(p) free((void *)(p))
#define PerlMemShared_realloc(p, n) realloc((void *)(p), (size_t)(n))
#define savesharedpv(p) savepv(p)
#define savesharedpvn(p, n) savepvn((p), (n))
#define savepvs(s) savepvn("" s "", sizeof(s) - 1)
static char *goperl_savesvpv(goperl_frame_t *f, SV *sv) {
    uint64_t len = 0;
    const char *p = goperl_sv_pv_(f, sv, &len);
    return savepvn(p ? p : "", len);
}
#define savesvpv(sv) goperl_savesvpv(_gof, (SV *)(sv))
#define memNE(a, b, n) (memcmp((a), (b), (n)) != 0)
#define memEQs(s1, l, s2) \
    ({ GOPERL_ONCE(goperl_memEQs_l_, l); \
       (sizeof(s2) - 1 == (size_t)goperl_memEQs_l_) && memcmp((s1), ("" s2 ""), goperl_memEQs_l_) == 0; })
#define memNEs(s1, l, s2) (!memEQs(s1, l, s2))
#define strLT(a, b) (strcmp((a), (b)) < 0)
#define strLE(a, b) (strcmp((a), (b)) <= 0)
#define strGT(a, b) (strcmp((a), (b)) > 0)
#define strGE(a, b) (strcmp((a), (b)) >= 0)

/* ---- character classes (handy.h, ASCII range) --------------------------- */

#ifndef isASCII
#define isASCII(c) ((UV)(c) < 128)
#endif
#ifndef isOCTAL
#define isOCTAL(c) \
    ({ GOPERL_ONCE(goperl_isOCTAL_c_, c); \
       goperl_isOCTAL_c_ >= '0' && goperl_isOCTAL_c_ <= '7'; })
#endif
#ifndef isIDFIRST
#define isIDFIRST(c) \
    ({ GOPERL_ONCE(goperl_isIDFIRST_c_, c); \
       isALPHA(goperl_isIDFIRST_c_) || goperl_isIDFIRST_c_ == '_'; })
#endif
#ifndef isIDCONT
#define isIDCONT(c) isWORDCHAR(c)
#endif
#ifndef isALNUMC
#define isALNUMC(c) \
    ({ GOPERL_ONCE(goperl_isALNUMC_c_, c); \
       isALPHA(goperl_isALNUMC_c_) || isDIGIT(goperl_isALNUMC_c_); })
#endif
#ifndef toFOLD
#define toFOLD(c) toLOWER(c)
#endif
#define UNI_TO_NATIVE(ch) ((UV)(ch))
#define NATIVE_TO_UNI(ch) ((UV)(ch))
#define UTF8_IS_CONTINUED(c) (((U8)(c)) & 0x80)

/* ---- the flag word (sv.h) ------------------------------------------------ */

#define SVpgv_GP SVp_SCREAM
#define SVphv_CLONEABLE SVp_SCREAM
#define SVf_OK \
    (SVf_IOK | SVf_NOK | SVf_POK | SVf_ROK | SVp_IOK | SVp_NOK | SVp_POK | SVpgv_GP)
#define SVf_THINKFIRST \
    (SVf_READONLY | SVf_PROTECT | SVf_ROK | SVf_FAKE | SVs_RMG | SVf_IsCOW)

/* The flag word after the change, as the value of perl's macros is (they
 * are assignments to SvFLAGS: `RETVAL = SvREADONLY_on(sv)`). */
#define GOPERL_FLAGS_CHANGE(sv, set, clear)                               \
    ((U32)goperl_op0(GOPERL_OP_SV_FLAGS, GOPERL_TOK(sv),                  \
                     ((uint64_t)(U32)(clear) << 32) | (U32)(set)))

#define SvOOK(sv) (SvFLAGS(sv) & SVf_OOK)
#define sv_backoff(sv) ((void)GOPERL_SVCALL(GOPERL_SVCALL_BACKOFF, sv, 0))
#define SvOOK_off(sv) \
    ({ GOPERL_ONCE(goperl_SvOOK_off_sv_, sv); \
       (void)(SvOOK(goperl_SvOOK_off_sv_) && GOPERL_SVCALL(GOPERL_SVCALL_BACKOFF, goperl_SvOOK_off_sv_, 0)); })
/* SvOK_off: clear the value flags, then give back an offset string start. */
static void goperl_sv_ok_off(goperl_frame_t *f, SV *sv, U32 clear) {
    U32 flags = (U32)goperl_do_op(f, GOPERL_OP_SV_FLAGS, GOPERL_TOK(sv),
                                  (uint64_t)clear << 32, 0, 0);
    if (flags & SVf_OOK)
        goperl_do_op3(f, GOPERL_OP_SV_CALL, GOPERL_TOK(sv), 0, (uint64_t)(GOPERL_SVCALL_BACKOFF), 0, 0);
}
#define SvOK_off(sv) \
    goperl_sv_ok_off(_gof, (SV *)(sv), SVf_OK | SVf_IVisUV | SVf_UTF8)
#define SvOK_off_exc_UV(sv) goperl_sv_ok_off(_gof, (SV *)(sv), SVf_OK | SVf_UTF8)
#define SvIOK_on(sv) GOPERL_FLAGS_CHANGE(sv, SVf_IOK | SVp_IOK, 0)
#define SvIOK_off(sv) GOPERL_FLAGS_CHANGE(sv, 0, SVf_IOK | SVp_IOK | SVf_IVisUV)
#define SvIOK_only(sv) \
    ({ GOPERL_ONCE(goperl_SvIOK_only_sv_, sv); \
       (SvOK_off(goperl_SvIOK_only_sv_), SvIOK_on(goperl_SvIOK_only_sv_)); })
#define SvIOK_only_UV(sv) \
    ({ GOPERL_ONCE(goperl_SvIOK_only_UV_sv_, sv); \
       (SvOK_off_exc_UV(goperl_SvIOK_only_UV_sv_), SvIOK_on(goperl_SvIOK_only_UV_sv_)); })
#define SvIOK_notUV(sv) ((SvFLAGS(sv) & (SVf_IOK | SVf_IVisUV)) == SVf_IOK)
#define SvIsUV_on(sv) GOPERL_FLAGS_CHANGE(sv, SVf_IVisUV, 0)
#define SvIsUV_off(sv) GOPERL_FLAGS_CHANGE(sv, 0, SVf_IVisUV)
#define SvNOK_on(sv) GOPERL_FLAGS_CHANGE(sv, SVf_NOK | SVp_NOK, 0)
#define SvNOK_off(sv) GOPERL_FLAGS_CHANGE(sv, 0, SVf_NOK | SVp_NOK)
#define SvNOK_only(sv) \
    ({ GOPERL_ONCE(goperl_SvNOK_only_sv_, sv); \
       (SvOK_off(goperl_SvNOK_only_sv_), SvNOK_on(goperl_SvNOK_only_sv_)); })
#define SvNIOK_off(sv) \
    GOPERL_FLAGS_CHANGE(sv, 0, SVf_IOK | SVf_NOK | SVp_IOK | SVp_NOK | SVf_IVisUV)
#define SvPOK_off(sv) GOPERL_FLAGS_CHANGE(sv, 0, SVf_POK | SVp_POK)
/* the private flags alone, as tied and other magical values carry them */
#define SvIOKp_on(sv) GOPERL_FLAGS_CHANGE(sv, SVp_IOK, 0)
#define SvNOKp_on(sv) GOPERL_FLAGS_CHANGE(sv, SVp_NOK, 0)
#define SvPOKp_on(sv) GOPERL_FLAGS_CHANGE(sv, SVp_POK, 0)
/* how far the public flags lie below the private ones (SVp_IOK >> PRIVSHIFT
 * is SVf_IOK) */
#define PRIVSHIFT 4
#define SvPOK_only(sv) \
    GOPERL_FLAGS_CHANGE(sv, SVf_POK | SVp_POK, (SVf_OK | SVf_IVisUV | SVf_UTF8) & ~(SVf_POK | SVp_POK))
#define SvPOK_only_UTF8(sv) \
    GOPERL_FLAGS_CHANGE(sv, SVf_POK | SVp_POK, (SVf_OK | SVf_IVisUV) & ~(SVf_POK | SVp_POK))
#define SvROK_on(sv) GOPERL_FLAGS_CHANGE(sv, SVf_ROK, 0)
#define SvROK_off(sv) GOPERL_FLAGS_CHANGE(sv, 0, SVf_ROK)
#define SvTEMP(sv) (SvFLAGS(sv) & SVs_TEMP)
#define SvTEMP_on(sv) GOPERL_FLAGS_CHANGE(sv, SVs_TEMP, 0)
#define SvTEMP_off(sv) GOPERL_FLAGS_CHANGE(sv, 0, SVs_TEMP)
#define SvOBJECT_on(sv) GOPERL_FLAGS_CHANGE(sv, SVs_OBJECT, 0)
#define SvOBJECT_off(sv) GOPERL_FLAGS_CHANGE(sv, 0, SVs_OBJECT)
#define SvGMAGICAL(sv) (SvFLAGS(sv) & SVs_GMG)
#define SvGMAGICAL_on(sv) GOPERL_FLAGS_CHANGE(sv, SVs_GMG, 0)
#define SvGMAGICAL_off(sv) GOPERL_FLAGS_CHANGE(sv, 0, SVs_GMG)
#define SvSMAGICAL(sv) (SvFLAGS(sv) & SVs_SMG)
#define SvSMAGICAL_on(sv) GOPERL_FLAGS_CHANGE(sv, SVs_SMG, 0)
#define SvSMAGICAL_off(sv) GOPERL_FLAGS_CHANGE(sv, 0, SVs_SMG)
#define SvRMAGICAL_on(sv) GOPERL_FLAGS_CHANGE(sv, SVs_RMG, 0)
#define SvRMAGICAL_off(sv) GOPERL_FLAGS_CHANGE(sv, 0, SVs_RMG)
#define SvMAGICAL(sv) (SvFLAGS(sv) & (SVs_GMG | SVs_SMG | SVs_RMG))
#define SvMAGICAL_on(sv) GOPERL_FLAGS_CHANGE(sv, SVs_GMG | SVs_SMG | SVs_RMG, 0)
#define SvMAGICAL_off(sv) GOPERL_FLAGS_CHANGE(sv, 0, SVs_GMG | SVs_SMG | SVs_RMG)
#define SvFAKE(sv) (SvFLAGS(sv) & SVf_FAKE)
#define SvFAKE_on(sv) GOPERL_FLAGS_CHANGE(sv, SVf_FAKE, 0)
#define SvFAKE_off(sv) GOPERL_FLAGS_CHANGE(sv, 0, SVf_FAKE)
#define SvTHINKFIRST(sv) (SvFLAGS(sv) & SVf_THINKFIRST)
#define SvPADTMP(sv) (SvFLAGS(sv) & SVs_PADTMP)
#define SvPADTMP_on(sv) GOPERL_FLAGS_CHANGE(sv, SVs_PADTMP, 0)
#define SvPADTMP_off(sv) GOPERL_FLAGS_CHANGE(sv, 0, SVs_PADTMP)
#define SvPADSTALE(sv) (SvFLAGS(sv) & SVs_PADSTALE)
#define SVs_PADMY 0
#define SvPADMY(sv) (!(SvFLAGS(sv) & SVs_PADTMP))
#define SvPADMY_on(sv) SvPADTMP_off(sv)
#define SvWEAKREF(sv) \
    ((SvFLAGS(sv) & (SVf_ROK | SVprv_WEAKREF)) == (SVf_ROK | SVprv_WEAKREF))
#define SvIsCOW(sv) (SvFLAGS(sv) & SVf_IsCOW)
#define SvIsCOW_on(sv) GOPERL_FLAGS_CHANGE(sv, SVf_IsCOW, 0)
#define SvIsCOW_off(sv) GOPERL_FLAGS_CHANGE(sv, 0, SVf_IsCOW)
#define SvIsCOW_shared_hash(sv) \
    ({ GOPERL_ONCE(goperl_SvIsCOW_shared_hash_sv_, sv); \
       (SvFLAGS(goperl_SvIsCOW_shared_hash_sv_) & (SVf_POK | SVf_IsCOW)) == (SVf_POK | SVf_IsCOW) && SvLEN(goperl_SvIsCOW_shared_hash_sv_) == 0; })
#define SvPROTECT(sv) (SvFLAGS(sv) & SVf_PROTECT)
/* The shared key a shared-hash scalar's string is (hv.h): the string is the
 * key of the HEK in the interpreter's memory, where its length, hash and
 * flags are read (Sereal::Encoder dedupes keys by it). */
#define SvSHARED_HEK_FROM_PV(pvx) ((HEK *)((const char *)(pvx) - STRUCT_OFFSET(HEK, hek_key)))
/* The magic of a v-string (sv.h): its spelling in mg_ptr and mg_len. */
#define SvVSTRING_mg(sv) (SvMAGICAL(sv) ? mg_find((SV *)(sv), PERL_MAGIC_vstring) : (MAGIC *)NULL)
#define SvIMMORTAL(sv) \
    ({ GOPERL_ONCE(goperl_SvIMMORTAL_sv_, sv); \
       SvREADONLY(goperl_SvIMMORTAL_sv_) && ((SV *)goperl_SvIMMORTAL_sv_ == &PL_sv_undef || (SV *)goperl_SvIMMORTAL_sv_ == &PL_sv_yes || (SV *)goperl_SvIMMORTAL_sv_ == &PL_sv_no || (SV *)goperl_SvIMMORTAL_sv_ == &PL_sv_zero || (SV *)goperl_SvIMMORTAL_sv_ == &PL_sv_placeholder); })
#define SvVOK(sv) (GOPERL_SVCALL(GOPERL_SVCALL_VOK, sv, 0) != 0)
#define SvGAMAGIC(sv) \
    ({ GOPERL_ONCE(goperl_SvGAMAGIC_sv_, sv); \
       SvGMAGICAL(goperl_SvGAMAGIC_sv_) || SvAMAGIC(goperl_SvGAMAGIC_sv_); })
#define assert_not_ROK(sv)
#define assert_not_glob(sv)

/* ---- magic ---------------------------------------------------------------- */

#define mg_get(sv) ((int)(int64_t)GOPERL_SVCALL(GOPERL_SVCALL_MG_GET, sv, 0))
#define mg_clear(sv) ((int)(int64_t)GOPERL_SVCALL(GOPERL_SVCALL_MG_CLEAR, sv, 0))
/* Every magic of the SV goes, the svt_free of each running (native code's
 * too: its MAGICs go with the interpreter's anchor of them). */
#define mg_free(sv) ((int)(int64_t)GOPERL_SVCALL(GOPERL_SVCALL_MG_FREE, sv, 0))
/* The flags come from the vtables of the chain as they are now: the
 * interpreter is told of the hooks of native code's MAGICs first (setting
 * the head the chain has gives the host a look at it), as a module may have
 * given one a vtable since it was attached (pTk's Tcl_ObjMagic sets
 * mg_virtual on what mg_find found, then calls mg_magical). A chain with no
 * uvar magic left (Variable::Magic's dispell unlinks its own) has the
 * interpreter's uvar magic that stood for it go too (bit 0). */
static void goperl_mg_magical(goperl_frame_t *f, SV *sv) __attribute__((unused));
static void goperl_mg_magical(goperl_frame_t *f, SV *sv) {
    goperl_flush(f);
    MAGIC *const head = goperl_api_magic_chain(f, GOPERL_TOK(sv));
    if (head) goperl_api_magic_set_head(f, GOPERL_TOK(sv), head);
    int uvar = 0;
    for (const MAGIC *mg = head; mg && !uvar; mg = mg->mg_moremagic)
        uvar = mg->mg_type == 'U'; /* PERL_MAGIC_uvar */
    (void)GOPERL_SVCALL(GOPERL_SVCALL_MG_MAGICAL, sv, uvar ? 0 : 1);
}
#define mg_magical(sv) goperl_mg_magical(_gof, (SV *)(sv))
#define mg_size(sv) ((U32)GOPERL_SVCALL(GOPERL_SVCALL_MG_SIZE, sv, 0))
#define SvGETMAGIC(sv) \
    ({ GOPERL_ONCE(goperl_SvGETMAGIC_sv_, sv); \
       (void)(SvGMAGICAL(goperl_SvGETMAGIC_sv_) && mg_get(goperl_SvGETMAGIC_sv_)); })
/* The extension magic of an SV lives in the host mirror chain. */
static MAGIC *goperl_mg_findext(goperl_frame_t *f, const SV *sv, int type,
                                const MGVTBL *vtbl) __attribute__((unused));
static MAGIC *goperl_mg_findext(goperl_frame_t *f, const SV *sv, int type,
                                const MGVTBL *vtbl) {
    if (!sv) return 0;
    goperl_flush(f);
    for (MAGIC *mg = goperl_api_magic_chain(f, GOPERL_TOK(sv)); mg;
         mg = mg->mg_moremagic)
        if (mg->mg_type == (char)type && mg->mg_virtual == vtbl) return mg;
    return 0;
}
#define mg_findext(sv, type, vtbl) \
    goperl_mg_findext(_gof, (const SV *)(sv), (type), (const MGVTBL *)(vtbl))

/* ---- taint ---------------------------------------------------------------- */

#define PL_tainting (*goperl_plint_ref(_gof, GOPERL_PL_TAINTING))
#define PL_tainted (*goperl_plint_ref(_gof, GOPERL_PL_TAINTED))
#define TAINTING_get PL_tainting
#define TAINT_get PL_tainted
#define TAINT ((void)(PL_tainting && (PL_tainted = 1)))
#define TAINT_NOT ((void)(PL_tainted = 0))
#define sv_tainted(sv) (GOPERL_SVCALL(GOPERL_SVCALL_TAINTED, sv, 0) != 0)
#define sv_taint(sv) ((void)GOPERL_SVCALL(GOPERL_SVCALL_TAINT, sv, 0))
#define sv_untaint(sv) ((void)GOPERL_SVCALL(GOPERL_SVCALL_UNTAINT, sv, 0))
#define SvTAINTED(sv) \
    ({ GOPERL_ONCE(goperl_SvTAINTED_sv_, sv); \
       SvMAGICAL(goperl_SvTAINTED_sv_) && sv_tainted(goperl_SvTAINTED_sv_); })
#define SvTAINTED_on(sv) ((void)(TAINTING_get && (sv_taint(sv), 0)))
#define SvTAINTED_off(sv) ((void)(TAINTING_get && (sv_untaint(sv), 0)))
#define SvTAINT(sv) ((void)(TAINTING_get && TAINT_get && (sv_taint(sv), 0)))

/* ---- value accessors with explicit get-magic control --------------------- */

static IV goperl_sv_iv_flags(goperl_frame_t *f, SV *sv, U32 flags) {
    return (IV)(int64_t)goperl_do_op3(f, GOPERL_OP_SV_CALL, GOPERL_TOK(sv), flags, (uint64_t)(GOPERL_SVCALL_2IV), 0, 0);
}
static UV goperl_sv_uv_flags(goperl_frame_t *f, SV *sv, U32 flags) {
    return (UV)goperl_do_op3(f, GOPERL_OP_SV_CALL, GOPERL_TOK(sv), flags, (uint64_t)(GOPERL_SVCALL_2UV), 0, 0);
}
static NV goperl_sv_nv_flags(goperl_frame_t *f, SV *sv, U32 flags) {
    union {
        uint64_t u;
        double d;
    } cvt;
    cvt.u = goperl_do_op3(f, GOPERL_OP_SV_CALL, GOPERL_TOK(sv), flags, (uint64_t)(GOPERL_SVCALL_2NV), 0, 0);
    return cvt.d;
}
static char *goperl_sv_pv_flags(goperl_frame_t *f, SV *sv, STRLEN *lenp,
                                U32 flags) {
    return goperl_str_op(f, GOPERL_OP_SV_CALL, GOPERL_TOK(sv), flags, GOPERL_SVCALL_2PV, lenp);
}
#define SvIV_nomg(sv) goperl_sv_iv_flags(_gof, (SV *)(sv), 0)
#define SvUV_nomg(sv) goperl_sv_uv_flags(_gof, (SV *)(sv), 0)
#define SvNV_nomg(sv) goperl_sv_nv_flags(_gof, (SV *)(sv), 0)
#define SvTRUE_nomg(sv) (GOPERL_SVCALL(GOPERL_SVCALL_2BOOL, sv, 0) != 0)
#define SvTRUE_NN(sv) SvTRUE(sv)
#define SvTRUE_nomg_NN(sv) SvTRUE_nomg(sv)
#define SvTRUEx(sv) SvTRUE(sv)
#define SvPV_flags(sv, lenvar, flags) \
    GOPERL_PV_LEN(lenvar, goperl_sv_pv_flags(_gof, (SV *)(sv), (STRLEN *)&goperl_pvlen_, (U32)(flags)))
#define SvPV_flags_const(sv, lenvar, flags) \
    ((const char *)SvPV_flags(sv, lenvar, flags))
#define SvPV_nomg(sv, lenvar) SvPV_flags(sv, lenvar, 0)
#define SvPV_nomg_const(sv, lenvar) ((const char *)SvPV_flags(sv, lenvar, 0))
#define SvPV_nomg_nolen(sv) goperl_sv_pv_flags(_gof, (SV *)(sv), 0, 0)
#define SvPV_nomg_const_nolen(sv) ((const char *)SvPV_nomg_nolen(sv))
#define SvPVx(sv, lenvar) SvPV(sv, lenvar)
#define SvPVx_nolen(sv) SvPV_nolen(sv)
#define SvPVx_const(sv, lenvar) SvPV_const(sv, lenvar)
#define SvPVx_nolen_const(sv) SvPV_nolen_const(sv)
#define SvPVX_mutable(sv) SvPVX(sv)
#define sv_2iv_flags(sv, flags) goperl_sv_iv_flags(_gof, (SV *)(sv), (U32)(flags))
#define sv_2uv_flags(sv, flags) goperl_sv_uv_flags(_gof, (SV *)(sv), (U32)(flags))
#define sv_2nv_flags(sv, flags) goperl_sv_nv_flags(_gof, (SV *)(sv), (U32)(flags))
#define sv_2pv_flags(sv, lenp, flags) \
    goperl_sv_pv_flags(_gof, (SV *)(sv), (STRLEN *)(lenp), (U32)(flags))
#define sv_2iv(sv) sv_2iv_flags(sv, SV_GMAGIC)
#define sv_2uv(sv) sv_2uv_flags(sv, SV_GMAGIC)
#define sv_2nv(sv) sv_2nv_flags(sv, SV_GMAGIC)
#define sv_2pv(sv, lenp) sv_2pv_flags(sv, lenp, SV_GMAGIC)
#define sv_2pv_nolen(sv) sv_2pv_flags(sv, 0, SV_GMAGIC)
#define sv_2bool(sv) (GOPERL_SVCALL(GOPERL_SVCALL_2BOOL, sv, SV_GMAGIC) != 0)
#define sv_2bool_flags(sv, flags) (GOPERL_SVCALL(GOPERL_SVCALL_2BOOL, sv, flags) != 0)
#define sv_pvn(sv, lenp) sv_2pv_flags(sv, lenp, SV_GMAGIC)
#define SvIV_please(sv) \
    ({ GOPERL_ONCE(goperl_SvIV_please_sv_, sv); \
       (void)(!(SvFLAGS(goperl_SvIV_please_sv_) & (SVf_IOK | SVp_IOK)) && (SvFLAGS(goperl_SvIV_please_sv_) & (SVf_NOK | SVf_POK)) && (SvIV(goperl_SvIV_please_sv_), 0)); })
#define SvIV_please_nomg(sv) \
    ({ GOPERL_ONCE(goperl_SvIV_please_nomg_sv_, sv); \
       !(SvFLAGS(goperl_SvIV_please_nomg_sv_) & (SVf_IOK | SVp_IOK)) && (SvFLAGS(goperl_SvIV_please_nomg_sv_) & (SVf_NOK | SVf_POK)) ? (sv_2iv_flags(goperl_SvIV_please_nomg_sv_, 0), SvIOK(goperl_SvIV_please_nomg_sv_)) : SvIOK(goperl_SvIV_please_nomg_sv_); })
#define POPp SvPV_nolen(POPs)
#define POPpbytex SvPVbyte_nolen(POPs)
#define TOPi ((IV)SvIV(TOPs))
#define TOPn ((NV)SvNV(TOPs))
#define TOPp SvPV_nolen(TOPs)
#define TOPm1s (*(sp - 1))
#define SETs(s) (*sp = (SV *)(s))

/* ---- body slots (sv.h) --------------------------------------------------- */

#define GOPERL_SLOT_SET(slot, sv, v)                                       \
    ((void)goperl_do_op3(_gof, GOPERL_OP_SV_SLOT_SET, GOPERL_TOK(sv), (uint64_t)(v), (uint64_t)(slot), 0, 0))
#define SvRV_set(sv, rv) GOPERL_SLOT_SET(GOPERL_SVSLOT_RV, sv, GOPERL_TOK(rv))
#define SvSTASH_set(sv, stash) \
    GOPERL_SLOT_SET(GOPERL_SVSLOT_STASH, sv, GOPERL_TOK(stash))
#define SvIV_set(sv, v) GOPERL_SLOT_SET(GOPERL_SVSLOT_IV, sv, (int64_t)(v))
#define SvUV_set(sv, v) GOPERL_SLOT_SET(GOPERL_SVSLOT_UV, sv, (UV)(v))
static void goperl_sv_nv_set(goperl_frame_t *f, SV *sv, NV v) {
    union {
        uint64_t u;
        double d;
    } cvt;
    cvt.d = v;
    goperl_do_op3(f, GOPERL_OP_SV_SLOT_SET, GOPERL_TOK(sv), cvt.u, (uint64_t)(GOPERL_SVSLOT_NV), 0, 0);
}
#define SvNV_set(sv, v) goperl_sv_nv_set(_gof, (SV *)(sv), (NV)(v))
#define SvLEN_set(sv, n)                                            \
    ({ GOPERL_ONCE(goperl_SvLEN_set_sv_, sv);                        \
       const uint64_t goperl_SvLEN_set_n_ = (uint64_t)(STRLEN)(n);   \
       goperl_flush(_gof);                                          \
       goperl_sv_len_store(_gof, GOPERL_TOK(goperl_SvLEN_set_sv_),   \
                           goperl_SvLEN_set_n_); })
/* The string buffer of an SV is interpreter memory: a buffer the
 * interpreter handed out (SvPVX, SvGROW) or NULL. Any other pointer is one
 * the module keeps in the slot for itself (see GOPERL_PV_HOST). */
#define SvPV_set(sv, p)                                            \
    ({ GOPERL_ONCE(goperl_SvPV_set_sv_, sv);                       \
       goperl_flush(_gof);                                         \
       goperl_sv_pvx_store(_gof, GOPERL_TOK(goperl_SvPV_set_sv_),  \
                           (const void *)(p)); })
#define SvPV_shrink_to_cur(sv) \
    ((void)GOPERL_SVCALL(GOPERL_SVCALL_SHRINK_TO_CUR, sv, 0))
/* The buffer made n bytes long, as sv.h has it: the interpreter's realloc
 * of the interpreter's buffer (saferealloc of an interpreter address). */
#define SvPV_renew(sv, n)                                                  \
    STMT_START {                                                           \
        SV *const goperl_SvPV_renew_sv_ = (SV *)(sv);                      \
        const STRLEN goperl_SvPV_renew_n_ = (STRLEN)(n);                   \
        SvLEN_set(goperl_SvPV_renew_sv_, goperl_SvPV_renew_n_);            \
        SvPV_set(goperl_SvPV_renew_sv_,                                    \
                 (char *)saferealloc((Malloc_t)SvPVX(goperl_SvPV_renew_sv_), \
                                     (MEM_SIZE)goperl_SvPV_renew_n_));     \
    }                                                                      \
    STMT_END
#define SvPVCLEAR(sv) sv_setpvs((sv), "")
#define Sv_Grow(sv, n) sv_grow((sv), (n))
#define SvGROW_mutable(sv, n) SvGROW((sv), (n))
#define sv_set_undef(sv) ((void)GOPERL_SVCALL(GOPERL_SVCALL_SET_UNDEF, sv, 0))
static char *goperl_sv_setpv_bufsize(goperl_frame_t *f, SV *sv, STRLEN cur,
                                     STRLEN len) {
    /* b = the size of the buffer, hi = cur */
    uint64_t gp = goperl_do_op3(f, GOPERL_OP_SV_CALL, GOPERL_TOK(sv), (uint64_t)len,
                                GOPERL_SVCALL_SETPV_BUFSIZE | ((uint64_t)(uint32_t)cur << 32),
                                0, 0);
    return gp ? (char *)goperl_guest_mem(f, gp) : 0;
}
#define sv_setpv_bufsize(sv, cur, len) \
    goperl_sv_setpv_bufsize(_gof, (SV *)(sv), (STRLEN)(cur), (STRLEN)(len))
#define sv_setbool(sv, b) sv_setsv((sv), boolSV(b))
#define sv_setbool_mg(sv, b) sv_setsv_mg((sv), boolSV(b))
#define sv_set_true(sv) sv_setsv((sv), &PL_sv_yes)
#define sv_set_false(sv) sv_setsv((sv), &PL_sv_no)
#define newSVbool(b) newSVsv(boolSV(b))
#define newSV_true() newSVsv(&PL_sv_yes)
#define newSV_false() newSVsv(&PL_sv_no)

/* ---- SV functions --------------------------------------------------------- */

#define sv_setsv_flags(d, s, flags) \
    ((void)GOPERL_SVCALL2(GOPERL_SVCALL_SETSV, d, s, flags))
#define sv_setsv_nomg(d, s) sv_setsv_flags(d, s, SV_DO_COW_SVSETSV)
#define SvSetSV_nosteal(d, s) \
    ({ GOPERL_ONCE(goperl_SvSetSV_nosteal_d_, d); \
       GOPERL_ONCE(goperl_SvSetSV_nosteal_s_, s); \
       (void)((SV *)goperl_SvSetSV_nosteal_d_ != (SV *)goperl_SvSetSV_nosteal_s_ && (sv_setsv_flags(goperl_SvSetSV_nosteal_d_, goperl_SvSetSV_nosteal_s_, SV_GMAGIC | SV_NOSTEAL | SV_DO_COW_SVSETSV), 0)); })
#define SvSetSV_nosteal_and(d, s, finally)                                \
    STMT_START {                                                          \
        GOPERL_ONCE(goperl_SvSetSV_nosteal_and_d_, d);                    \
        GOPERL_ONCE(goperl_SvSetSV_nosteal_and_s_, s);                    \
        if ((SV *)goperl_SvSetSV_nosteal_and_d_ != (SV *)goperl_SvSetSV_nosteal_and_s_) { \
            sv_setsv_flags(goperl_SvSetSV_nosteal_and_d_, goperl_SvSetSV_nosteal_and_s_, SV_GMAGIC | SV_NOSTEAL | SV_DO_COW_SVSETSV); \
            finally;                                                      \
        }                                                                 \
    }                                                                     \
    STMT_END
#define SvSetMagicSV_nosteal(d, s)                                        \
    STMT_START {                                                          \
        GOPERL_ONCE(goperl_SvSetMagicSV_nosteal_d_, d);                   \
        SvSetSV_nosteal_and(goperl_SvSetMagicSV_nosteal_d_, s,            \
                            SvSETMAGIC(goperl_SvSetMagicSV_nosteal_d_));  \
    }                                                                     \
    STMT_END
#define sv_catsv_flags(d, s, flags) \
    ((void)GOPERL_SVCALL2(GOPERL_SVCALL_CATSV, d, s, flags))
#define sv_catsv_nomg(d, s) sv_catsv_flags(d, s, 0)
#define sv_catsv_mg(d, s) sv_catsv_flags(d, s, SV_GMAGIC | SV_SMAGIC)
#define sv_catpvn_flags(d, p, len, flags) \
    ((void)GOPERL_SVSTR(GOPERL_SVSTR_CATPVN, d, flags, (p), (len)))
#define sv_catpvn_nomg(d, p, len) sv_catpvn_flags(d, p, len, 0)
#define sv_catpvn_nomg_maybeutf8(d, p, len, is_utf8) \
    sv_catpvn_flags(d, p, len, (is_utf8) ? SV_CATUTF8 : SV_CATBYTES)
#define sv_catpv_flags(d, p, flags) \
    ({ GOPERL_ONCE(goperl_sv_catpv_flags_p_, p); \
       sv_catpvn_flags(d, goperl_sv_catpv_flags_p_, __builtin_strlen(goperl_sv_catpv_flags_p_), flags); })
#define sv_catpv_nomg(d, p) \
    ({ GOPERL_ONCE(goperl_sv_catpv_nomg_p_, p); \
       sv_catpvn_flags(d, goperl_sv_catpv_nomg_p_, __builtin_strlen(goperl_sv_catpv_nomg_p_), 0); })
#define sv_catpv_mg(d, p) \
    ({ GOPERL_ONCE(goperl_sv_catpv_mg_p_, p); \
       sv_catpvn_flags(d, goperl_sv_catpv_mg_p_, __builtin_strlen(goperl_sv_catpv_mg_p_), SV_GMAGIC | SV_SMAGIC); })
#define sv_catpvs_nomg(d, s) sv_catpvn_flags(d, "" s "", sizeof(s) - 1, 0)
#define sv_catpvs_mg(d, s) \
    sv_catpvn_flags(d, "" s "", sizeof(s) - 1, SV_GMAGIC | SV_SMAGIC)
#define sv_catpvs_flags(d, s, flags) \
    sv_catpvn_flags(d, "" s "", sizeof(s) - 1, flags)
#define sv_setpv_mg(sv, p) \
    ({ GOPERL_ONCE(goperl_sv_setpv_mg_sv_, sv); \
       (sv_setpv((goperl_sv_setpv_mg_sv_), (p)), SvSETMAGIC(goperl_sv_setpv_mg_sv_)); })
#define sv_setpvs_mg(sv, s) \
    ({ GOPERL_ONCE(goperl_sv_setpvs_mg_sv_, sv); \
       (sv_setpvs((goperl_sv_setpvs_mg_sv_), s), SvSETMAGIC(goperl_sv_setpvs_mg_sv_)); })
#define sv_setpvf_mg(sv, ...) \
    ({ GOPERL_ONCE(goperl_sv_setpvf_mg_sv_, sv); \
       (sv_setpvf((goperl_sv_setpvf_mg_sv_), __VA_ARGS__), SvSETMAGIC(goperl_sv_setpvf_mg_sv_)); })
#define sv_catpvf_mg(sv, ...) \
    ({ GOPERL_ONCE(goperl_sv_catpvf_mg_sv_, sv); \
       (sv_catpvf((goperl_sv_catpvf_mg_sv_), __VA_ARGS__), SvSETMAGIC(goperl_sv_catpvf_mg_sv_)); })
#define sv_cmp_flags(a, b, flags) \
    ((I32)(int64_t)GOPERL_SVCALL2(GOPERL_SVCALL_CMP, a, b, flags))
#define sv_eq_flags(a, b, flags) \
    ((I32)GOPERL_SVCALL2(GOPERL_SVCALL_EQ, a, b, flags))
/* The numeric comparison of perl 5.44 (overloading included; SV_GMAGIC
 * runs get magic). */
#define sv_numcmp_flags(a, b, flags) \
    ((I32)(int64_t)GOPERL_SVCALL2(62 /* GOPERL_SVCALL_NUMCMP */, a, b, flags))
#define sv_numcmp(a, b) sv_numcmp_flags((a), (b), SV_GMAGIC)
#define sv_dec(sv) ((void)GOPERL_SVCALL(GOPERL_SVCALL_DEC, sv, 0))
#define sv_dec_nomg(sv) ((void)GOPERL_SVCALL(GOPERL_SVCALL_DEC, sv, 1))
#define sv_len_utf8(sv) ((STRLEN)GOPERL_SVCALL(GOPERL_SVCALL_LEN_UTF8, sv, 0))
#define sv_utf8_encode(sv) ((void)GOPERL_SVCALL(GOPERL_SVCALL_UTF8_ENCODE, sv, 0))
#define sv_utf8_upgrade_flags(sv, flags) \
    ((STRLEN)GOPERL_SVCALL(GOPERL_SVCALL_UTF8_UPGRADE, sv, (U32)(flags)))
#define sv_utf8_upgrade_nomg(sv) sv_utf8_upgrade_flags(sv, 0)
/* upgraded with room for extra bytes more and the NUL after them */
#define sv_utf8_upgrade_flags_grow(sv, flags, extra)                        \
    ({ SV *const goperl_sufg_sv_ = (SV *)(sv);                              \
       const STRLEN goperl_sufg_extra_ = (STRLEN)(extra);                   \
       const STRLEN goperl_sufg_len_ = sv_utf8_upgrade_flags(goperl_sufg_sv_, (flags)); \
       if (SvLEN(goperl_sufg_sv_) < SvCUR(goperl_sufg_sv_) + goperl_sufg_extra_ + 1) \
           SvGROW(goperl_sufg_sv_, SvCUR(goperl_sufg_sv_) + goperl_sufg_extra_ + 1); \
       goperl_sufg_len_; })
#define sv_unref_flags(sv, flags) \
    ((void)GOPERL_SVCALL(GOPERL_SVCALL_UNREF, sv, (U32)(flags)))
#define sv_unref(sv) sv_unref_flags(sv, 0)
#define sv_rvunweaken(sv) \
    ({ GOPERL_ONCE(goperl_sv_rvunweaken_sv_, sv); \
       ((void)GOPERL_SVCALL(GOPERL_SVCALL_RVUNWEAKEN, goperl_sv_rvunweaken_sv_, 0), (SV *)(goperl_sv_rvunweaken_sv_)); })
#define sv_ref(dst, sv, ob)                                               \
    goperl_sv_ref(_gof, (SV *)(dst), (const SV *)(sv), (int)(ob))
static SV *goperl_sv_ref(goperl_frame_t *f, SV *dst, const SV *sv, int ob)
    __attribute__((unused));
static SV *goperl_sv_ref(goperl_frame_t *f, SV *dst, const SV *sv, int ob) {
    /* the interpreter's sv_ref(NULL, ...): a new mortal with the name */
    SV *r = GOPERL_SV(goperl_do_op3(f, GOPERL_OP_SV_CALL, GOPERL_TOK(sv), (uint64_t)ob,
                                    (uint64_t)(GOPERL_SVCALL_REF), 0, 0));
    if (!dst) return r;
    /* perl fills the caller's SV; here the name is copied into it, and the
     * mortal goes with the other temporaries (it is not the caller's to
     * release) */
    goperl_do_op(f, GOPERL_OP_SV_SETSV, GOPERL_TOK(dst), GOPERL_TOK(r), 0, 0);
    return dst;
}
#define sv_does_pvn(sv, name, len, flags) \
    (GOPERL_SVSTR(GOPERL_SVSTR_DOES, sv, flags, (name), (len)) != 0)
#define sv_does_pv(sv, name, flags) \
    ({ GOPERL_ONCE(goperl_sv_does_pv_name_, name); \
       sv_does_pvn(sv, goperl_sv_does_pv_name_, __builtin_strlen(goperl_sv_does_pv_name_), flags); })
#define sv_does(sv, name) \
    ({ GOPERL_ONCE(goperl_sv_does_name_, name); \
       sv_does_pvn(sv, goperl_sv_does_name_, __builtin_strlen(goperl_sv_does_name_), 0); })
#define sv_pos_u2b_flags(sv, uoffset, lenp, flags) \
    goperl_sv_pos_u2b(_gof, (SV *)(sv), (STRLEN)(uoffset), (STRLEN *)(lenp))
static STRLEN goperl_sv_pos_u2b(goperl_frame_t *f, SV *sv, STRLEN uoffset,
                                STRLEN *lenp) __attribute__((unused));
static STRLEN goperl_sv_pos_u2b(goperl_frame_t *f, SV *sv, STRLEN uoffset,
                                STRLEN *lenp) {
    STRLEN boff = (STRLEN)goperl_do_op3(f, GOPERL_OP_SV_CALL, GOPERL_TOK(sv), uoffset,
                                        GOPERL_SVCALL_POS_U2B, 0, 0);
    if (lenp) {
        STRLEN bend = (STRLEN)goperl_do_op3(f, GOPERL_OP_SV_CALL, GOPERL_TOK(sv),
                                            uoffset + *lenp, GOPERL_SVCALL_POS_U2B, 0, 0);
        *lenp = bend - boff;
    }
    return boff;
}
#define sv_pos_b2u_flags(sv, offset, flags) \
    ((STRLEN)GOPERL_SVCALL(GOPERL_SVCALL_POS_B2U, sv, (STRLEN)(offset)))
static void goperl_sv_pos_u2b_i32(goperl_frame_t *f, SV *sv, I32 *offsetp,
                                  I32 *lenp) __attribute__((unused));
static void goperl_sv_pos_u2b_i32(goperl_frame_t *f, SV *sv, I32 *offsetp,
                                  I32 *lenp) {
    STRLEN ulen = lenp ? (STRLEN)*lenp : 0;
    *offsetp = (I32)goperl_sv_pos_u2b(f, sv, (STRLEN)*offsetp, lenp ? &ulen : 0);
    if (lenp) *lenp = (I32)ulen;
}
#define sv_pos_u2b(sv, offsetp, lenp) \
    goperl_sv_pos_u2b_i32(_gof, (SV *)(sv), (offsetp), (lenp))
#define sv_pos_b2u(sv, offsetp) \
    ({ GOPERL_ONCE(goperl_sv_pos_b2u_offsetp_, offsetp); \
       ((void)(*(goperl_sv_pos_b2u_offsetp_) = (I32)sv_pos_b2u_flags(sv, *(goperl_sv_pos_b2u_offsetp_), SV_GMAGIC))); })
#define newSVrv(rv, classname)                                            \
    goperl_newSVrv(_gof, (SV *)(rv), (const char *)(classname))
static SV *goperl_newSVrv(goperl_frame_t *f, SV *rv, const char *classname) {
    return GOPERL_SV(goperl_do_op3(f, GOPERL_OP_SV_STR, GOPERL_TOK(rv), ((uint64_t)(classname ? 0 : 1) << 32) |
            (uint32_t)(classname ? __builtin_strlen(classname) : 0), (uint64_t)(GOPERL_SVSTR_NEWSVRV), classname ? classname : "", classname ? __builtin_strlen(classname) : 0));
}
#define sv_setref_pvn(rv, classname, pv, n) \
    ({ GOPERL_ONCE(goperl_sv_setref_pvn_rv_, rv); \
       (sv_setpvn(newSVrv((goperl_sv_setref_pvn_rv_), (classname)), (pv), (n)), (SV *)(goperl_sv_setref_pvn_rv_)); })
#define sv_setref_pvs(rv, classname, s) \
    sv_setref_pvn(rv, classname, "" s "", sizeof(s) - 1)
#define sv_setref_uv(rv, classname, uv) \
    ({ GOPERL_ONCE(goperl_sv_setref_uv_rv_, rv); \
       (sv_setuv(newSVrv((goperl_sv_setref_uv_rv_), (classname)), (uv)), (SV *)(goperl_sv_setref_uv_rv_)); })
#define sv_setref_nv(rv, classname, nv) \
    ({ GOPERL_ONCE(goperl_sv_setref_nv_rv_, rv); \
       (sv_setnv(newSVrv((goperl_sv_setref_nv_rv_), (classname)), (nv)), (SV *)(goperl_sv_setref_nv_rv_)); })
#define sv_setrv_inc(sv, ref) \
    ((void)GOPERL_SVCALL(GOPERL_SVCALL_SETRV_INC, sv, GOPERL_TOK(ref)))
#define sv_setrv_noinc(sv, ref) \
    ((void)GOPERL_SVCALL(GOPERL_SVCALL_SETRV_NOINC, sv, GOPERL_TOK(ref)))
#define sv_setrv_inc_mg(sv, ref) \
    ({ GOPERL_ONCE(goperl_sv_setrv_inc_mg_sv_, sv); \
       (sv_setrv_inc(goperl_sv_setrv_inc_mg_sv_, ref), SvSETMAGIC(goperl_sv_setrv_inc_mg_sv_)); })
#define sv_setrv_noinc_mg(sv, ref) \
    ({ GOPERL_ONCE(goperl_sv_setrv_noinc_mg_sv_, sv); \
       (sv_setrv_noinc(goperl_sv_setrv_noinc_mg_sv_, ref), SvSETMAGIC(goperl_sv_setrv_noinc_mg_sv_)); })
#define newSVsv_flags(sv, flags) \
    GOPERL_SV(GOPERL_SVCALL(GOPERL_SVCALL_NEWSVSV, sv, (U32)(flags)))
#define newSVsv_nomg(sv) newSVsv_flags(sv, SV_NOSTEAL)
#define sv_newref(sv) SvREFCNT_inc(sv)
#define newAVav(av) ((AV *)GOPERL_SV(GOPERL_SVCALL(GOPERL_SVCALL_NEWAVAV, av, 0)))
#define newAVhv(hv) ((AV *)GOPERL_SV(GOPERL_SVCALL(GOPERL_SVCALL_NEWAVHV, hv, 0)))
#define av_delete(av, key, flags)                                         \
    GOPERL_SV(goperl_do_op3(_gof, GOPERL_OP_SV_CALL, GOPERL_TOK(av),       \
                            (uint64_t)(int64_t)(SSize_t)(key),             \
                            GOPERL_SVCALL_AV_DELETE | ((uint64_t)(U32)(flags) << 32), 0, 0))
#define AvMAX(av) ((SSize_t)(int64_t)GOPERL_SVCALL(GOPERL_SVCALL_AV_MAX, av, 0))
#define HvUSEDKEYS(hv) ((STRLEN)GOPERL_SVCALL(GOPERL_SVCALL_HV_USEDKEYS, hv, 0))
#define HvTOTALKEYS(hv) ((STRLEN)GOPERL_SVCALL(GOPERL_SVCALL_HV_TOTALKEYS, hv, 0))
#define HvFILL(hv) ((STRLEN)GOPERL_SVCALL(GOPERL_SVCALL_HV_FILL, hv, 0))
#define HvMAX(hv) ((STRLEN)GOPERL_SVCALL(GOPERL_SVCALL_HV_MAX, hv, 0))
#define hv_existss(hv, key) hv_exists((hv), "" key "", sizeof(key) - 1)
#define hv_deletes(hv, key, flags) \
    hv_delete((hv), "" key "", sizeof(key) - 1, (flags))
#define mro_get_linear_isa(stash) \
    ((AV *)GOPERL_SV(GOPERL_SVCALL(GOPERL_SVCALL_MRO_LINEAR_ISA, stash, 0)))
#define mro_method_changed_in(stash) \
    ((void)GOPERL_SVCALL(GOPERL_SVCALL_MRO_METHOD_CHANGED, stash, 0))
/* The type character and the flags of an IO are lvalues (`IoTYPE(io) =
 * '<'`, `IoFLAGS(io) |= IOf_FLUSH`), as its handles are (perl.h). */
#define IoTYPE(io) \
    (*(char *)goperl_lv_ref(_gof, (SV *)(io), GOPERL_LV_IOTYPE, GOPERL_LV_AT))
#define IoFLAGS(io) \
    (*(U8 *)goperl_lv_ref(_gof, (SV *)(io), GOPERL_LV_IOFLAGS, GOPERL_LV_AT))
/* The other words of an IO's body, lvalues as on perl
 * (`IoLINES(io) = 1`, `IoBOTTOM_NAME(io) = savepv(name)`). */
#define GOPERL_IOWORD_LV(io, type, w) \
    (*(type *)goperl_lv_ref(_gof, (SV *)(io), GOPERL_LV_IOWORD + (w), GOPERL_LV_AT))
#define IoLINES(io) GOPERL_IOWORD_LV(io, IV, 0)
#define IoPAGE(io) GOPERL_IOWORD_LV(io, IV, 1)
#define IoPAGE_LEN(io) GOPERL_IOWORD_LV(io, IV, 2)
#define IoLINES_LEFT(io) GOPERL_IOWORD_LV(io, IV, 3)
#define IoTOP_GV(io) GOPERL_IOWORD_LV(io, GV *, 4)
#define IoFMT_GV(io) GOPERL_IOWORD_LV(io, GV *, 5)
#define IoBOTTOM_GV(io) GOPERL_IOWORD_LV(io, GV *, 6)
#define IoTOP_NAME(io) GOPERL_IOWORD_LV(io, char *, 7)
#define IoFMT_NAME(io) GOPERL_IOWORD_LV(io, char *, 8)
#define IoBOTTOM_NAME(io) GOPERL_IOWORD_LV(io, char *, 9)

/* ---- do_openn: the open behind perl's open function ----------------------- */

static bool goperl_do_openn(goperl_frame_t *f, GV *gv, const char *oname,
                            I32 len, int as_raw, int rawmode, int rawperm,
                            const void *supplied_fp, SV **svp, I32 num)
    __attribute__((unused));
static bool goperl_do_openn(goperl_frame_t *f, GV *gv, const char *oname,
                            I32 len, int as_raw, int rawmode, int rawperm,
                            const void *supplied_fp, SV **svp, I32 num) {
    const size_t count = (svp && num > 0) ? (size_t)num : 0;
    const size_t mode_len = len > 0 ? (size_t)len : 0;
    /* words: [as_raw][rawmode][rawperm][supplied handle][count], the
     * arguments, then the mode text */
    const goperl_gword_t head[5] = {(goperl_gword_t)(int64_t)as_raw,
                                    (goperl_gword_t)(int64_t)rawmode,
                                    (goperl_gword_t)(int64_t)rawperm,
                                    goperl_pio_tok(f, supplied_fp), (goperl_gword_t)count};
    const size_t size = sizeof head + count * 8 + mode_len;
    char *req = (char *)malloc(size + 1);
    if (!req) goperl_croakf(f, "%s", "Out of memory!");
    memcpy(req, head, sizeof head);
    for (size_t i = 0; i < count; i++) {
        const goperl_gword_t tok = GOPERL_TOK(svp[i]);
        memcpy(req + sizeof head + i * 8, &tok, 8);
    }
    if (mode_len) memcpy(req + sizeof head + count * 8, oname, mode_len);
    req[size] = '\0';
    /* goperl_do_op's contract, with the request released before a death
     * inside the open unwinds this call */
    goperl_flush(f);
    const int32_t failed_before = f->failed;
    const uint64_t opened =
        goperl_xs_op(f, GOPERL_OP_DO_OPENN, GOPERL_TOK(gv), size, req, size);
    goperl_gen_bump();
    free(req);
    if (!failed_before && f->failed && f->jb) goperl_raise(f);
    if (!opened) goperl_errno_from_interp(f);
    return opened != 0;
}
#define do_openn(gv, oname, len, as_raw, rawmode, rawperm, supplied_fp, svp, num) \
    goperl_do_openn(_gof, (GV *)(gv), (oname), (I32)(len), (as_raw), (rawmode),   \
                    (rawperm), (supplied_fp), (svp), (I32)(num))
#define do_open(gv, name, len, as_raw, rawmode, rawperm, supplied_fp) \
    do_openn((gv), (name), (len), (as_raw), (rawmode), (rawperm), (supplied_fp), (SV **)0, 0)

/* ---- tie magic --------------------------------------------------------------
 * Tie magic is the interpreter's and is not mirrored. SvTIED_mg answers
 * with a MAGIC that says what a module asks of it: that the SV is tied,
 * by which kind of tie, and to what object (SvTIED_obj). It lasts as long
 * as the outermost native call. */
static MAGIC *goperl_sv_tied_mg(goperl_frame_t *f, SV *sv, int how)
    __attribute__((unused));
static MAGIC *goperl_sv_tied_mg(goperl_frame_t *f, SV *sv, int how) {
    const uint64_t tied = goperl_do_op3(f, GOPERL_OP_SV_CALL, GOPERL_TOK(sv), (uint64_t)(int64_t)how, (uint64_t)(GOPERL_SVCALL_TIED), 0, 0);
    if (!tied) return 0;
    MAGIC *mg = (MAGIC *)goperl_arena_alloc(sizeof(MAGIC));
    memset(mg, 0, sizeof(MAGIC));
    mg->mg_type = (char)how;
    mg->mg_obj = GOPERL_SV(tied);
    return mg;
}
#define SvTIED_mg(sv, how) goperl_sv_tied_mg(_gof, (SV *)(sv), (int)(how))
#ifndef SvTIED_obj
#define SvTIED_obj(sv, mg) ((mg)->mg_obj)
#endif

/* ---- formatted strings ---------------------------------------------------- */

static void goperl_sv_vsetpvfn(goperl_frame_t *f, SV *sv, const char *pat,
                               STRLEN patlen, va_list *args, SV **svargs,
                               Size_t sv_count, int cat)
    __attribute__((unused));
static void goperl_sv_vsetpvfn(goperl_frame_t *f, SV *sv, const char *pat,
                               STRLEN patlen, va_list *args, SV **svargs,
                               Size_t sv_count, int cat) {
    goperl_text_t t;
    if (svargs || sv_count)
        goperl_croakf(f, "sv_v%spvfn: formatting from an SV argument list is "
                         "not available to native XS modules",
                      cat ? "cat" : "set");
    /* the format is patlen bytes, not necessarily terminated there */
    char *fmt = (char *)malloc(patlen + 1);
    if (!fmt) goperl_croakf(f, "%s", "Out of memory!");
    memcpy(fmt, pat, patlen);
    fmt[patlen] = '\0';
    goperl_text_init(&t);
    goperl_vtext_p(f, &t, fmt, args);
    free(fmt);
    if (cat)
        goperl_text_op(f, &t, GOPERL_OP_SV_STR, GOPERL_TOK(sv),
                       ((uint64_t)(uint32_t)(SV_GMAGIC | (t.utf8 ? SV_CATUTF8 : SV_CATBYTES)) << 32) | (uint32_t)t.len,
                       0 /* GOPERL_SVSTR_CATPVN */);
    else
        goperl_text_op(f, &t, GOPERL_OP_SV_STR, GOPERL_TOK(sv), GOPERL_TEXT_ARG(&t, 0), (uint64_t)(GOPERL_SVSTR_SETPVN));
}
#define sv_vsetpvfn(sv, pat, patlen, args, svargs, n, tainted) \
    goperl_sv_vsetpvfn(_gof, (SV *)(sv), (pat), (patlen), (args), (svargs), (n), 0)
#define sv_vcatpvfn(sv, pat, patlen, args, svargs, n, tainted) \
    goperl_sv_vsetpvfn(_gof, (SV *)(sv), (pat), (patlen), (args), (svargs), (n), 1)
#define sv_vsetpvf(sv, pat, args) \
    ({ GOPERL_ONCE(goperl_sv_vsetpvf_pat_, pat); \
       goperl_sv_vsetpvfn(_gof, (SV *)(sv), (goperl_sv_vsetpvf_pat_), __builtin_strlen(goperl_sv_vsetpvf_pat_), (args), 0, 0, 0); })
#define sv_vcatpvf(sv, pat, args) \
    ({ GOPERL_ONCE(goperl_sv_vcatpvf_pat_, pat); \
       goperl_sv_vsetpvfn(_gof, (SV *)(sv), (goperl_sv_vcatpvf_pat_), __builtin_strlen(goperl_sv_vcatpvf_pat_), (args), 0, 0, 1); })
#define sv_vsetpvf_mg(sv, pat, args) \
    ({ GOPERL_ONCE(goperl_sv_vsetpvf_mg_sv_, sv); \
       (sv_vsetpvf(goperl_sv_vsetpvf_mg_sv_, pat, args), SvSETMAGIC(goperl_sv_vsetpvf_mg_sv_)); })
#define sv_vcatpvf_mg(sv, pat, args) \
    ({ GOPERL_ONCE(goperl_sv_vcatpvf_mg_sv_, sv); \
       (sv_vcatpvf(goperl_sv_vcatpvf_mg_sv_, pat, args), SvSETMAGIC(goperl_sv_vcatpvf_mg_sv_)); })

/* vcroak without a format dies with $@, as croak(NULL) does. */
static void goperl_vcroak_args(goperl_frame_t *f, const char *pat, va_list *args)
    __attribute__((noreturn, unused));
static void goperl_vcroak_args(goperl_frame_t *f, const char *pat, va_list *args) {
    goperl_text_t t;
    if (!pat) goperl_croak_errsv(f);
    goperl_text_init(&t);
    goperl_vtext_p(f, &t, pat, args);
    goperl_croak_text(f, &t);
}
#define vcroak(pat, args) goperl_vcroak_args(_gof, (pat), (args))
static void goperl_vwarn_args(goperl_frame_t *f, const char *pat, va_list *args)
    __attribute__((unused));
static void goperl_vwarn_args(goperl_frame_t *f, const char *pat, va_list *args) {
    goperl_text_t t;
    goperl_text_init(&t);
    goperl_vtext_p(f, &t, pat, args);
    goperl_text_op(f, &t, GOPERL_OP_SV_STR, 0, GOPERL_TEXT_ARG(&t, 0), (uint64_t)(GOPERL_SVSTR_WARN));
}
#define vwarn(pat, args) goperl_vwarn_args(_gof, (pat), (args))
static char *goperl_vform(goperl_frame_t *f, const char *pat, va_list *args)
    __attribute__((unused));
static char *goperl_vform(goperl_frame_t *f, const char *pat, va_list *args) {
    char *buf = goperl_form_bufs_v[goperl_form_ix_v++ % GOPERL_FORM_BUFS];
    goperl_text_t t;
    goperl_text_init(&t);
    goperl_vtext_p(f, &t, pat, args);
    const size_t n = t.len < GOPERL_FORM_LEN - 1 ? t.len : GOPERL_FORM_LEN - 1;
    memcpy(buf, t.p, n);
    buf[n] = '\0';
    goperl_text_free(&t);
    return buf;
}
#define vform(pat, args) goperl_vform(_gof, (pat), (args))
#define die_sv(sv) croak_sv(sv)

/* ---- globs and subs (gv.h, cv.h) ----------------------------------------- */

#define GvSVn(gv) GOPERL_SV(GOPERL_GVCALL(GOPERL_GVCALL_SVN, gv, 0))
#define GvAVn(gv) ((AV *)GOPERL_SV(GOPERL_GVCALL(GOPERL_GVCALL_AVN, gv, 0)))
#define GvHVn(gv) ((HV *)GOPERL_SV(GOPERL_GVCALL(GOPERL_GVCALL_HVN, gv, 0)))
#define GvIOn(gv) ((IO *)GOPERL_SV(GOPERL_GVCALL(GOPERL_GVCALL_ION, gv, 0)))
/* The IO of a glob is an lvalue (`GvIOp(gv) = newIO()`): the plain store
 * of gv.h. */
#define GvIOp(gv) \
    (*(IO **)goperl_lv_ref(_gof, (SV *)(gv), GOPERL_LV_GVIO, GOPERL_LV_AT))
#define GvFORM(gv) ((CV *)GOPERL_SV(GOPERL_GVCALL(GOPERL_GVCALL_FORM, gv, 0)))
/* An expression whose value is the CV stored, as gv.h's assignment is
 * (Package::Stash::XS tests it: `if ((GvCV_set(g, cv))) GvIMPORTED_CV_on(g)`). */
#define GvCV_set(gv, cv)                                                   \
    ({ GOPERL_ONCE(goperl_GvCV_set_cv_, cv);                               \
       (void)GOPERL_GVCALL(GOPERL_GVCALL_CV_SET, gv, GOPERL_TOK(goperl_GvCV_set_cv_)); \
       (CV *)(goperl_GvCV_set_cv_); })
/* An lvalue as on perl (`GvFLAGS(gv) ^= flag`): the whole word is stored. */
#define GvFLAGS(gv) \
    (*(U32 *)goperl_lv_ref(_gof, (SV *)(gv), GOPERL_LV_GVFLAGS, GOPERL_LV_AT))
#define GOPERL_GVFLAGS_CHANGE(gv, set, clear)                             \
    ((void)GOPERL_GVCALL2(GOPERL_GVCALL_FLAGS, gv, (U32)(set), (clear)))
#define GvMULTI(gv) (GvFLAGS(gv) & GVf_MULTI)
#define GvMULTI_on(gv) GOPERL_GVFLAGS_CHANGE(gv, GVf_MULTI, 0)
#define GvMULTI_off(gv) GOPERL_GVFLAGS_CHANGE(gv, 0, GVf_MULTI)
#define GvINTRO(gv) (GvFLAGS(gv) & GVf_INTRO)
#define GvINTRO_on(gv) GOPERL_GVFLAGS_CHANGE(gv, GVf_INTRO, 0)
#define GvINTRO_off(gv) GOPERL_GVFLAGS_CHANGE(gv, 0, GVf_INTRO)
#define GvASSUMECV(gv) (GvFLAGS(gv) & GVf_ASSUMECV)
#define GvASSUMECV_on(gv) GOPERL_GVFLAGS_CHANGE(gv, GVf_ASSUMECV, 0)
#define GvIMPORTED(gv) (GvFLAGS(gv) & GVf_IMPORTED)
#define GvIMPORTED_on(gv) GOPERL_GVFLAGS_CHANGE(gv, GVf_IMPORTED, 0)
#define GvIMPORTED_CV(gv) (GvFLAGS(gv) & GVf_IMPORTED_CV)
#define GvIMPORTED_CV_on(gv) GOPERL_GVFLAGS_CHANGE(gv, GVf_IMPORTED_CV, 0)
#define GvIMPORTED_SV_on(gv) GOPERL_GVFLAGS_CHANGE(gv, GVf_IMPORTED_SV, 0)
#define GvIMPORTED_AV_on(gv) GOPERL_GVFLAGS_CHANGE(gv, GVf_IMPORTED_AV, 0)
#define GvIMPORTED_HV_on(gv) GOPERL_GVFLAGS_CHANGE(gv, GVf_IMPORTED_HV, 0)
#define GvCVGEN(gv) \
    (*(U32 *)goperl_lv_ref(_gof, (SV *)(gv), GOPERL_LV_GVCVGEN, GOPERL_LV_AT))
/* The name of a glob is changed with gv_name_set; GvNAME is not an lvalue
 * (nor is it on perl). */
#define gv_name_set(gv, name, len, flags) \
    ((void)GOPERL_SVSTR(GOPERL_SVSTR_GV_NAME_SET, gv, flags, (name), (len)))
/* The name of a stash, given (or taken away, with a null name) as perl's
 * hv_name_set does it (Package::Stash::XS names the stash it makes for an
 * anonymous package "__ANON__"); HvNAME is not an lvalue. flags:
 * HVhek_UTF8; bit 31 of what crosses says that there is no name. */
#define hv_name_set(hv, name, len, flags)                                  \
    ({ GOPERL_ONCE(goperl_hv_name_set_name_, name);                        \
       (void)GOPERL_SVSTR(GOPERL_SVSTR_HV_NAME_SET, hv,                     \
                          (U32)(flags) | (goperl_hv_name_set_name_ ? 0u : 0x80000000u), \
                          goperl_hv_name_set_name_, goperl_hv_name_set_name_ ? (len) : 0); })
/* GvGP is an identity: modules test it and compare it, never look inside. */
#define GvGP(gv) ((void *)(uintptr_t)GOPERL_GVCALL(GOPERL_GVCALL_GP, gv, 0))
#define GvFILE(gv) \
    ((char *)goperl_intern_op(_gof, GOPERL_OP_GV_CALL, GOPERL_TOK(gv), 0, GOPERL_GVCALL_FILE))
#define GvLINE(gv) ((U32)GOPERL_GVCALL(GOPERL_GVCALL_LINE, gv, 0))
#define GvNAMEUTF8(gv) (GOPERL_GVCALL(GOPERL_GVCALL_NAME_UTF8, gv, 0) != 0)
#define GvNAME_get(gv) GvNAME(gv)
#define GvNAMELEN_get(gv) GvNAMELEN(gv)
static void goperl_gv_fullname4(goperl_frame_t *f, SV *sv, const GV *gv,
                                const char *prefix, int keepmain,
                                int effective) __attribute__((unused));
static void goperl_gv_fullname4(goperl_frame_t *f, SV *sv, const GV *gv,
                                const char *prefix, int keepmain,
                                int effective) {
    /* b = the SV, hi = keepmain<<1 | effective */
    goperl_do_op3(f, GOPERL_OP_GV_CALL, GOPERL_TOK(gv), GOPERL_TOK(sv),
                  GOPERL_GVCALL_FULLNAME |
                      ((uint64_t)((keepmain ? 2 : 0) | (effective ? 1 : 0)) << 32),
                  0, 0);
    if (prefix && *prefix) {
        /* the interpreter writes the prefix first; here it goes in front
         * of the name the interpreter produced */
        SV *tmp = GOPERL_SV(goperl_do_op(f, GOPERL_OP_NEW_PVN, 0, __builtin_strlen(prefix),
                                         prefix, __builtin_strlen(prefix)));
        goperl_do_op(f, GOPERL_OP_SV_CATSV, GOPERL_TOK(tmp), GOPERL_TOK(sv), 0, 0);
        goperl_do_op(f, GOPERL_OP_SV_SETSV, GOPERL_TOK(sv), GOPERL_TOK(tmp), 0, 0);
        goperl_do_op(f, GOPERL_OP_SV_REFCNT_DEC, GOPERL_TOK(tmp), 0, 0, 0);
    }
}
#define gv_fullname4(sv, gv, prefix, keepmain) \
    goperl_gv_fullname4(_gof, (SV *)(sv), (const GV *)(gv), (prefix), (keepmain), 0)
#define gv_efullname4(sv, gv, prefix, keepmain) \
    goperl_gv_fullname4(_gof, (SV *)(sv), (const GV *)(gv), (prefix), (keepmain), 1)
#define gv_fullname3(sv, gv, prefix) gv_fullname4(sv, gv, prefix, TRUE)
#define gv_efullname3(sv, gv, prefix) gv_efullname4(sv, gv, prefix, TRUE)

#define GOPERL_METH_FLAGS(level, flags) \
    (((U32)(uint8_t)(int8_t)(level) << 24) | ((U32)(flags)&0xFFFFFF))
#define gv_fetchmeth_pvn(stash, name, len, level, flags)                  \
    ((GV *)GOPERL_SV(GOPERL_SVSTR(GOPERL_SVSTR_FETCHMETH, stash,          \
                                  GOPERL_METH_FLAGS(level, flags), (name), (len))))
#define gv_fetchmeth(stash, name, len, level) \
    gv_fetchmeth_pvn(stash, name, len, level, 0)
#define gv_fetchmeth_pv(stash, name, level, flags) \
    ({ GOPERL_ONCE(goperl_gv_fetchmeth_pv_name_, name); \
       gv_fetchmeth_pvn(stash, goperl_gv_fetchmeth_pv_name_, __builtin_strlen(goperl_gv_fetchmeth_pv_name_), level, flags); })
#define gv_fetchmeth_pvn_autoload(stash, name, len, level, flags)         \
    ((GV *)GOPERL_SV(GOPERL_SVSTR(GOPERL_SVSTR_FETCHMETH_AUTOLOAD, stash, \
                                  GOPERL_METH_FLAGS(level, flags), (name), (len))))
#define gv_fetchmeth_autoload(stash, name, len, level) \
    gv_fetchmeth_pvn_autoload(stash, name, len, level, 0)
#define gv_fetchmethod_pvn_flags(stash, name, len, flags)                 \
    ((GV *)GOPERL_SV(GOPERL_SVSTR(GOPERL_SVSTR_FETCHMETHOD, stash, flags, \
                                  (name), (len))))
#define gv_fetchmethod_pv_flags(stash, name, flags) \
    ({ GOPERL_ONCE(goperl_gv_fetchmethod_pv_flags_name_, name); \
       gv_fetchmethod_pvn_flags(stash, goperl_gv_fetchmethod_pv_flags_name_, __builtin_strlen(goperl_gv_fetchmethod_pv_flags_name_), flags); })
#define gv_fetchmethod_flags(stash, name, flags) \
    ({ GOPERL_ONCE(goperl_gv_fetchmethod_flags_name_, name); \
       gv_fetchmethod_pvn_flags(stash, goperl_gv_fetchmethod_flags_name_, __builtin_strlen(goperl_gv_fetchmethod_flags_name_), flags); })
#define gv_fetchpvn_flags(name, len, flags, svt) \
    ({ GOPERL_ONCE(goperl_gv_fetchpvn_flags_len_, len); \
       ((GV *)GOPERL_SV(goperl_do_op3(_gof, GOPERL_OP_SV_STR, svt, ((uint64_t)(U32)(flags) << 32) | (uint32_t)(goperl_gv_fetchpvn_flags_len_), (uint64_t)(GOPERL_SVSTR_GV_FETCHPVN), (name), (uint64_t)(goperl_gv_fetchpvn_flags_len_)))); })
#define gv_fetchpvn(name, len, flags, svt) gv_fetchpvn_flags(name, len, flags, svt)
#define get_cvn_flags(name, len, flags) \
    ({ GOPERL_ONCE(goperl_get_cvn_flags_len_, len); \
       ((CV *)GOPERL_SV(goperl_do_op3(_gof, GOPERL_OP_SV_STR, 0, ((uint64_t)(U32)(flags) << 32) | (uint32_t)(goperl_get_cvn_flags_len_), (uint64_t)(GOPERL_SVSTR_GET_CVN), (name), (uint64_t)(goperl_get_cvn_flags_len_)))); })
#define whichsig_pvn(sig, len) \
    ({ GOPERL_ONCE(goperl_whichsig_pvn_len_, len); \
       ((I32)(int64_t)goperl_do_op3(_gof, GOPERL_OP_SV_STR, 0, (uint32_t)(goperl_whichsig_pvn_len_), (uint64_t)(GOPERL_SVSTR_WHICHSIG), (sig), (uint64_t)(goperl_whichsig_pvn_len_))); })
#define whichsig_pv(sig) \
    ({ GOPERL_ONCE(goperl_whichsig_pv_sig_, sig); \
       whichsig_pvn((goperl_whichsig_pv_sig_), __builtin_strlen(goperl_whichsig_pv_sig_)); })
#define whichsig(sig) whichsig_pv(sig)

#define CvFLAGS(cv) ((U32)GOPERL_GVCALL(GOPERL_GVCALL_CV_FLAGS, cv, 0))
#define GOPERL_CVFLAGS_CHANGE(cv, set, clear)                             \
    ((void)GOPERL_GVCALL2(GOPERL_GVCALL_CV_FLAGS, cv, (U32)(set), (clear)))
#define CvCONST(cv) (CvFLAGS(cv) & CVf_CONST)
#define CvCONST_on(cv) GOPERL_CVFLAGS_CHANGE(cv, CVf_CONST, 0)
#define CvCONST_off(cv) GOPERL_CVFLAGS_CHANGE(cv, 0, CVf_CONST)
#define CvANON(cv) (CvFLAGS(cv) & CVf_ANON)
#define CvANON_on(cv) GOPERL_CVFLAGS_CHANGE(cv, CVf_ANON, 0)
#define CvANON_off(cv) GOPERL_CVFLAGS_CHANGE(cv, 0, CVf_ANON)
#define CvNAMED(cv) (CvFLAGS(cv) & CVf_NAMED)
#define CvLVALUE(cv) (CvFLAGS(cv) & CVf_LVALUE)
#define CvLVALUE_on(cv) GOPERL_CVFLAGS_CHANGE(cv, CVf_LVALUE, 0)
#define CvLVALUE_off(cv) GOPERL_CVFLAGS_CHANGE(cv, 0, CVf_LVALUE)
/* CvMETHOD is the :method attribute, which perl now calls
 * CvNOWARN_AMBIGUOUS; a method of a class is CvIsMETHOD. */
#define CvMETHOD(cv) (CvFLAGS(cv) & CVf_METHOD)
#define CvMETHOD_on(cv) GOPERL_CVFLAGS_CHANGE(cv, CVf_METHOD, 0)
#define CvMETHOD_off(cv) GOPERL_CVFLAGS_CHANGE(cv, 0, CVf_METHOD)
#define CvNOWARN_AMBIGUOUS(cv) CvMETHOD(cv)
#define CvNOWARN_AMBIGUOUS_on(cv) CvMETHOD_on(cv)
#define CvNOWARN_AMBIGUOUS_off(cv) CvMETHOD_off(cv)
#define CvIsMETHOD(cv) (CvFLAGS(cv) & CVf_IsMETHOD)
#define CvIsMETHOD_on(cv) GOPERL_CVFLAGS_CHANGE(cv, CVf_IsMETHOD, 0)
#define CvIsMETHOD_off(cv) GOPERL_CVFLAGS_CHANGE(cv, 0, CVf_IsMETHOD)
#define CvCLONE(cv) (CvFLAGS(cv) & CVf_CLONE)
#define CvCLONED(cv) (CvFLAGS(cv) & CVf_CLONED)
#define CvUNIQUE(cv) (CvFLAGS(cv) & CVf_UNIQUE)
#define CvNODEBUG(cv) (CvFLAGS(cv) & CVf_NODEBUG)
#define CvNODEBUG_on(cv) GOPERL_CVFLAGS_CHANGE(cv, CVf_NODEBUG, 0)
#define CvNODEBUG_off(cv) GOPERL_CVFLAGS_CHANGE(cv, 0, CVf_NODEBUG)
#define CvWEAKOUTSIDE(cv) (CvFLAGS(cv) & CVf_WEAKOUTSIDE)
#define CvDYNFILE(cv) (CvFLAGS(cv) & CVf_DYNFILE)
#define CvCVGV_RC(cv) (CvFLAGS(cv) & CVf_CVGV_RC)
#define CvLEXICAL(cv) (CvFLAGS(cv) & CVf_LEXICAL)
#define CvOUTSIDE(cv) \
    ((CV *)GOPERL_SV(GOPERL_GVCALL(GOPERL_GVCALL_CV_OUTSIDE, cv, 0)))
#define CvGV_set(cv, gv) \
    ((void)GOPERL_GVCALL(GOPERL_GVCALL_CV_GV_SET, cv, GOPERL_TOK(gv)))
#define CvSTASH_set(cv, stash) \
    ((void)GOPERL_GVCALL(GOPERL_GVCALL_CV_STASH_SET, cv, GOPERL_TOK(stash)))
#define CvPROTO(cv) \
    ((const char *)goperl_intern_op(_gof, GOPERL_OP_GV_CALL, GOPERL_TOK(cv), 0, \
                                    GOPERL_GVCALL_CV_PROTO))
#define CvPROTOLEN(cv) \
    ((STRLEN)goperl_op_second(_gof, GOPERL_OP_GV_CALL, GOPERL_TOK(cv), 0, GOPERL_GVCALL_CV_PROTO))
#define cv_const_sv(cv) GOPERL_SV(GOPERL_SVCALL(GOPERL_SVCALL_CV_CONST_SV, cv, 0))
#define cv_undef(cv) ((void)GOPERL_SVCALL(GOPERL_SVCALL_CV_UNDEF, cv, 0))

/* The interpreter keeps a sub's prototype in the string slot of its CV;
 * newXSproto is newXS followed by that assignment. */
static CV *goperl_newXS_proto(goperl_frame_t *f, const char *name,
                              void (*fn)(pTHX_ CV *), const char *proto)
    __attribute__((unused));
static CV *goperl_newXS_proto(goperl_frame_t *f, const char *name,
                              void (*fn)(pTHX_ CV *), const char *proto) {
    CV *cv = goperl_newXS(f, name, fn);
    if (cv && proto)
        goperl_do_op(f, GOPERL_OP_SV_SETPVN, GOPERL_TOK(cv), __builtin_strlen(proto),
                     proto, __builtin_strlen(proto));
    return cv;
}
#define newXSproto(name, fn, file, proto) \
    goperl_newXS_proto(_gof, (name), (fn), (proto))
#define newXSproto_portable(name, fn, file, proto) \
    goperl_newXS_proto(_gof, (name), (fn), (proto))
#define newXS_flags(name, fn, file, proto, flags) \
    goperl_newXS_proto(_gof, (name), (fn), (proto))

/* ---- compile-time hints of a COP ----------------------------------------- */

#define CopHINTS_get(c) (goperl_cop_fields((COP *)(c))->cop_hints)
#define IN_BYTES (CopHINTS_get(PL_curcop) & HINT_BYTES)
#define IN_LOCALE_RUNTIME (CopHINTS_get(PL_curcop) & HINT_LOCALE)
#define cop_hints_fetch_pvn(cop, key, keylen, hash, flags) \
    ({ GOPERL_ONCE(goperl_cop_hints_fetch_pvn_keylen_, keylen); \
       GOPERL_SV(goperl_do_op3(_gof, GOPERL_OP_SV_STR, (cop)->gop, ((uint64_t)(U32)(flags) << 32) | (uint32_t)(goperl_cop_hints_fetch_pvn_keylen_), (uint64_t)(GOPERL_SVSTR_COP_HINT), (key), (uint64_t)(goperl_cop_hints_fetch_pvn_keylen_))); })
#define cop_hints_fetch_pv(cop, key, hash, flags) \
    ({ GOPERL_ONCE(goperl_cop_hints_fetch_pv_key_, key); \
       cop_hints_fetch_pvn(cop, goperl_cop_hints_fetch_pv_key_, __builtin_strlen(goperl_cop_hints_fetch_pv_key_), hash, flags); })
#define cop_hints_fetch_pvs(cop, key, flags) \
    cop_hints_fetch_pvn(cop, "" key "", sizeof(key) - 1, 0, flags)

/* ---- interpreter variables (read) ---------------------------------------- */

#define PL_errgv ((GV *)GOPERL_SV(GOPERL_PLVAR(GOPERL_PL_ERRGV)))
#define PL_defstash ((HV *)GOPERL_SV(GOPERL_PLVAR(GOPERL_PL_DEFSTASH)))
#define PL_incgv ((GV *)GOPERL_SV(GOPERL_PLVAR(GOPERL_PL_INCGV)))
/* The globs of $a and $b of the sort block that is running: a module that
 * calls a block with the two (List::MoreUtils' pairwise and its sorts)
 * saves them, points them at the caller's *a and *b, and sets the scalars. */
#define PL_firstgv (*(GV **)goperl_plsv_ref(_gof, GOPERL_PL_FIRSTGV))
#define PL_secondgv (*(GV **)goperl_plsv_ref(_gof, GOPERL_PL_SECONDGV))
#define PL_stderrgv ((GV *)GOPERL_SV(GOPERL_PLVAR(GOPERL_PL_STDERRGV)))
#define PL_stdingv ((GV *)GOPERL_SV(GOPERL_PLVAR(GOPERL_PL_STDINGV)))
#define PL_defoutgv (*(GV **)goperl_plsv_ref(_gof, GOPERL_PL_DEFOUTGV))
#define PL_argvgv ((GV *)GOPERL_SV(GOPERL_PLVAR(GOPERL_PL_ARGVGV)))
#define PL_envgv ((GV *)GOPERL_SV(GOPERL_PLVAR(GOPERL_PL_ENVGV)))
#define PL_statgv ((GV *)GOPERL_SV(GOPERL_PLVAR(GOPERL_PL_STATGV)))
#define PL_last_in_gv ((GV *)GOPERL_SV(GOPERL_PLVAR(GOPERL_PL_LASTGV)))
#define PL_rs (*goperl_plsv_ref(_gof, GOPERL_PL_RS))
#define PL_beginav (*(AV **)goperl_plsv_ref(_gof, GOPERL_PL_BEGINAV))
#define PL_unitcheckav (*(AV **)goperl_plsv_ref(_gof, GOPERL_PL_UNITCHECKAV))
#define PL_strtab (*(HV **)goperl_plsv_ref(_gof, GOPERL_PL_STRTAB))
#define PL_sv_zero (*GOPERL_SV(GOPERL_PLVAR(GOPERL_PL_SV_ZERO)))
#define PL_sv_placeholder (*GOPERL_SV(GOPERL_PLVAR(GOPERL_PL_SV_PLACEHOLDER)))
/* an lvalue: Syntax::Keyword::Defer marks the eval it pushes (PL_in_eval =
 * EVAL_INEVAL|EVAL_KEEPERR) */
#define PL_in_eval (*goperl_plint_ref(_gof, GOPERL_PL_IN_EVAL))
#define PL_dirty (GOPERL_PLVAR(GOPERL_PL_DIRTY) != 0)
#define PL_savestack_ix ((I32)(int64_t)GOPERL_PLVAR(GOPERL_PL_SAVESTACK_IX))
#define PL_tmps_ix ((SSize_t)(int64_t)GOPERL_PLVAR(GOPERL_PL_TMPS_IX))
#define PL_tmps_floor ((SSize_t)(int64_t)GOPERL_PLVAR(GOPERL_PL_TMPS_FLOOR))
#define PL_localizing (*goperl_plint_ref(_gof, GOPERL_PL_LOCALIZING))
#define GOPERL_PL_STATUSVALUE 62
#define PL_statusvalue (*goperl_plint_ref(_gof, GOPERL_PL_STATUSVALUE))
/* The position pos() reports for a scalar, in characters or with
 * SV_POSBYTES in bytes: sv_regex_global_pos_get says whether it has one,
 * _set gives it one, _clear takes it away (perl 5.42's API; Term-TermKey
 * uses it, and walks the regex_global magic itself only without it).
 * CALL 28, 29, 30. */
#ifndef SV_POSBYTES
#define SV_POSBYTES SV_CATBYTES
#endif
static bool goperl_regex_global_pos_get(goperl_frame_t *f, SV *sv, STRLEN *posp, U32 flags)
    __attribute__((unused));
static bool goperl_regex_global_pos_get(goperl_frame_t *f, SV *sv, STRLEN *posp, U32 flags) {
    uint64_t pos = 0;
    const uint64_t found = goperl_do_opx(f, GOPERL_OP_CALL, GOPERL_TOK(sv), 0,
                                         (uint64_t)28 | ((uint64_t)flags << 32), 0, 0, &pos);
    if (found && posp) *posp = (STRLEN)pos;
    return found != 0;
}
#define sv_regex_global_pos_get(sv, posp, flags) \
    goperl_regex_global_pos_get(_gof, (SV *)(sv), (posp), (U32)(flags))
#define sv_regex_global_pos_set(sv, pos, flags)                                     \
    ((void)goperl_do_op3(_gof, GOPERL_OP_CALL, GOPERL_TOK((SV *)(sv)), (uint64_t)(pos), \
                         (uint64_t)29 | ((uint64_t)(U32)(flags) << 32), 0, 0))
#define sv_regex_global_pos_clear(sv) \
    ((void)goperl_do_op3(_gof, GOPERL_OP_CALL, GOPERL_TOK((SV *)(sv)), 0, (uint64_t)30, 0, 0))
/* The seed of the interpreter's hash function, PERL_HASH_SEED_BYTES bytes
 * (Data-HashMap seeds its own hash with it). It does not change while the
 * program runs; the copy is this module's (CALL 31). */
#define PERL_HASH_SEED_BYTES 32
static const U8 *goperl_hash_seed(goperl_frame_t *f) __attribute__((unused));
static const U8 *goperl_hash_seed(goperl_frame_t *f) {
    static U8 seed[PERL_HASH_SEED_BYTES];
    static int have;
    if (!have) {
        const goperl_gword_t *w = (const goperl_gword_t *)goperl_guest_mem(
            f, goperl_do_op3(f, GOPERL_OP_CALL, 0, 0, (uint64_t)31, 0, 0));
        if (!w || w[0] != PERL_HASH_SEED_BYTES)
            goperl_croakf(f, "PL_hash_seed: the interpreter's seed is not %d bytes",
                          PERL_HASH_SEED_BYTES);
        memcpy(seed, w + 1, sizeof seed);
        have = 1;
    }
    return seed;
}
#define PL_hash_seed (goperl_hash_seed(_gof))
/* is_lvalue_sub: how the sub being run was called, as an lvalue or not
 * (the OPpLVAL_INTRO bits of perl's); LVRET: whether the XSUB's result is
 * an lvalue's, as perl's pp.h defines it on the entersub that called it
 * (CALL 32). */
#define is_lvalue_sub() \
    ((I32)(int64_t)goperl_do_op3(_gof, GOPERL_OP_CALL, 0, 0, (uint64_t)32, 0, 0))
#define LVRET (goperl_do_op3(_gof, GOPERL_OP_CALL, 0, 1, (uint64_t)32, 0, 0) != 0)
/* the status of the program as perl's perl.h gives it where it is not
 * VMS: $? itself */
#define STATUS_UNIX PL_statusvalue
#define STATUS_CURRENT STATUS_UNIX
#define STATUS_EXIT STATUS_UNIX
/* cv_name: the name of a sub, "Package::name" (only "name" with
 * CV_NAME_NOTQUAL), set into sv or a new mortal; a value that is no sub
 * and no glob is itself, as perl's does (CALL 27: CV_NAME). */
#ifndef CV_NAME_NOTQUAL
#define CV_NAME_NOTQUAL 1
#endif
#define cv_name(cv, sv, flags)                                                   \
    ((SV *)GOPERL_SV(goperl_do_op3(_gof, GOPERL_OP_CALL, GOPERL_TOK((SV *)(cv)), \
                                   GOPERL_TOK((SV *)(sv)),                       \
                                   (uint64_t)27 | ((uint64_t)(U32)(flags) << 32), 0, 0)))
#define PL_perl_destruct_level (*goperl_plint_ref(_gof, GOPERL_PL_DESTRUCT_LEVEL))
#define PL_unsafe (GOPERL_PLVAR(GOPERL_PL_UNSAFE) != 0)
#define PL_cop_seqmax ((U32)GOPERL_PLVAR(GOPERL_PL_COP_SEQMAX))
#define PL_origfilename \
    ((char *)goperl_intern_op(_gof, GOPERL_OP_PLVAR_GET, GOPERL_PL_ORIGFILENAME, 0, 0))
#define PL_stack_max (PL_stack_base + _gof->stcap - 1)
/* The scratch length variable of old XS code (`SvPV(sv, PL_na)`). */
#define PL_na (GOPERL_SHARED->na)

enum perl_phase {
    PERL_PHASE_CONSTRUCT = 0,
    PERL_PHASE_START = 1,
    PERL_PHASE_CHECK = 2,
    PERL_PHASE_INIT = 3,
    PERL_PHASE_RUN = 4,
    PERL_PHASE_END = 5,
    PERL_PHASE_DESTRUCT = 6
};
#define PL_phase ((enum perl_phase)GOPERL_PLVAR(GOPERL_PL_PHASE))
static const char *const goperl_phase_names_v[] __attribute__((unused)) = {
    "CONSTRUCT", "START", "CHECK", "INIT", "RUN", "END", "DESTRUCT", 0};
#define PL_phase_names goperl_phase_names_v

/* The interpreter's message strings (perl.h). */
#define PL_no_modify "Modification of a read-only value attempted"
#define PL_no_func "The %s function is unimplemented"
#define PL_memory_wrap "panic: memory wrap"
#define PL_no_mem "Out of memory!\n"
#define PL_no_usym "Can't use an undefined value as %s reference"
#define PL_no_symref "Can't use string (\"%.32s\") as %s ref while \"strict refs\" in use"
#define PL_no_wrongref "Can't use %s ref as %s ref"
#define PL_no_aelem "Modification of non-creatable array value attempted, subscript %d"
#define PL_no_helem_sv \
    "Modification of non-creatable hash value attempted, subscript \"%" SVf "\""
#define PL_no_security "Insecure dependency in %s%s"
/* the interpreter's strings of true and false, which are the buffers of
 * its booleans (code tells core false by SvPVX(sv) == PL_No, as
 * Sereal::Encoder does) */
#define GOPERL_PL_YES 128
#define GOPERL_PL_NO 129
#define PL_Yes ((const char *)goperl_guest_mem(_gof, GOPERL_PLVAR(GOPERL_PL_YES)))
#define PL_No ((const char *)goperl_guest_mem(_gof, GOPERL_PLVAR(GOPERL_PL_NO)))

/* ---- calls ----------------------------------------------------------------- */

#define perl_call_sv(sv, flags) call_sv((sv), (flags))
#define perl_call_pv(name, flags) call_pv((name), (flags))
#define perl_call_method(name, flags) call_method((name), (flags))
#define perl_eval_sv(sv, flags) eval_sv((sv), (flags))
#define perl_eval_pv(code, croak_on_error) eval_pv((code), (croak_on_error))
#define perl_get_sv(name, flags) get_sv((name), (flags))
#define perl_get_av(name, flags) get_av((name), (flags))
#define perl_get_hv(name, flags) get_hv((name), (flags))
#define perl_get_cv(name, flags) get_cv((name), (flags))

static I32 goperl_call_argv(goperl_frame_t *f, const char *name, I32 flags,
                            char **argv) __attribute__((unused));
static I32 goperl_call_argv(goperl_frame_t *f, const char *name, I32 flags,
                            char **argv) {
    goperl_pushmark(f, (I32)(f->psp - f->stbase));
    while (*argv) {
        goperl_stack_extend(f, f->psp, 1);
        uint64_t tok = goperl_do_op(f, GOPERL_OP_NEW_PVN, 0, __builtin_strlen(*argv),
                                    *argv, __builtin_strlen(*argv));
        *++f->psp = GOPERL_SV(goperl_do_op(f, GOPERL_OP_SV_MORTAL, tok, 0, 0, 0));
        argv++;
    }
    return goperl_call_pv(f, name, flags);
}
#define call_argv(name, flags, argv) \
    goperl_call_argv(_gof, (name), (flags), (argv))
#define perl_call_argv(name, flags, argv) call_argv((name), (flags), (argv))

/* require_pv evaluates `require q<NUL>name<NUL>` and leaves a failure in
 * $@, as perl.c does. */
static void goperl_require_pv(goperl_frame_t *f, const char *pv)
    __attribute__((unused));
static void goperl_require_pv(goperl_frame_t *f, const char *pv) {
    static const char head[] = "require q";
    size_t n = __builtin_strlen(pv), o = sizeof(head) - 1;
    char *code = (char *)malloc(o + n + 2);
    if (!code) goperl_croakf(f, "%s", PL_no_mem);
    memcpy(code, head, o);
    code[o++] = '\0';
    memcpy(code + o, pv, n);
    o += n;
    code[o++] = '\0';
    uint64_t tok = goperl_do_op(f, GOPERL_OP_NEW_PVN, 0, o, code, o);
    free(code);
    SV *sv = GOPERL_SV(goperl_do_op(f, GOPERL_OP_SV_MORTAL, tok, 0, 0, 0));
    SV **keep = f->psp;
    goperl_eval_sv(f, sv, G_DISCARD);
    f->psp = keep;
}
#define require_pv(pv) goperl_require_pv(_gof, (pv))

/* A module that runs a callback on "a new stack" gets one in the
 * interpreter, of the type it names: the callback runs on it (a `last` or
 * a goto in it finds no context of the module's caller's to leave to, as
 * on perl). The module keeps working on the one host stack meanwhile: the
 * position to come back to is kept on the mark stack. */
#define PUSHSTACKi(type)                                                  \
    STMT_START {                                                          \
        PL_stack_sp = sp;                                                 \
        goperl_pushmark(_gof, (I32)(sp - PL_stack_base));                 \
        (void)GOPERL_INTERNAL(_gof, GOPERL_INT_PUSHSTACK, 0, (I32)(type)); \
    }                                                                     \
    STMT_END
#define PUSHSTACK PUSHSTACKi(PERLSI_UNKNOWN)
#define POPSTACK                                                          \
    STMT_START {                                                          \
        (void)GOPERL_INTERNAL(_gof, GOPERL_INT_POPSTACK, 0, 0);            \
        PL_stack_sp = PL_stack_base + goperl_popmark(_gof);               \
    }                                                                     \
    STMT_END

#define XST_mUV(i, v) (ST(i) = sv_2mortal(newSVuv((UV)(v))))
#define XST_mPVN(i, v, n) (ST(i) = sv_2mortal(newSVpvn((v), (n))))
#define XSRETURN_UV(v)                             \
    STMT_START {                                   \
        XST_mUV(0, v);                             \
        XSRETURN(1);                               \
    }                                              \
    STMT_END
#define XSRETURN_PVN(v, n)                         \
    STMT_START {                                   \
        XST_mPVN(0, v, n);                         \
        XSRETURN(1);                               \
    }                                              \
    STMT_END
#define mXPUSHn(n) XPUSHs(sv_2mortal(newSVnv(n)))
#define dUNDERBAR dNOOP
#define UNDERBAR DEFSV
#define UVf UVuf

/* ---- scopes and saves ------------------------------------------------------ */

#define ENTER_with_name(name) ENTER
#define LEAVE_with_name(name) LEAVE
#define GOPERL_SAVE(sel, b) \
    goperl_do_op(_gof, GOPERL_OP_SAVE_MISC, (uint64_t)(sel), (uint64_t)(b), 0, 0)
#define save_freesv(sv) ((void)GOPERL_SAVE(GOPERL_SAVE_FREESV, GOPERL_TOK(sv)))
#define save_mortalizesv(sv) \
    ((void)GOPERL_SAVE(GOPERL_SAVE_MORTALIZESV, GOPERL_TOK(sv)))
#define SAVEMORTALIZESV(sv) save_mortalizesv(sv)
#define save_item(sv) ((void)GOPERL_SAVE(GOPERL_SAVE_ITEM, GOPERL_TOK(sv)))
#define save_ary(gv) ((AV *)GOPERL_SV(GOPERL_SAVE(GOPERL_SAVE_ARY, GOPERL_TOK(gv))))
#define save_hash(gv) ((HV *)GOPERL_SV(GOPERL_SAVE(GOPERL_SAVE_HASH, GOPERL_TOK(gv))))
#define save_gp(gv, empty)                                                \
    ((void)goperl_do_op3(_gof, GOPERL_OP_SAVE_MISC, GOPERL_SAVE_GP, GOPERL_TOK(gv), \
                         (uint64_t)(U32)(empty), 0, 0))
#define SAVEHINTS() ((void)GOPERL_SAVE(GOPERL_SAVE_HINTS, 0))
#define SAVECOMPILEWARNINGS() ((void)GOPERL_SAVE(GOPERL_SAVE_COMPILE_WARNINGS, 0))
/* IV and long are pointer-width on the hosts the SDK supports. */
typedef char goperl_iv_is_pointer_width
    [(sizeof(IV) == sizeof(void *) && sizeof(long) == sizeof(void *)) ? 1 : -1]
    __attribute__((unused));
#define SAVEIV(i) goperl_save_int8(_gof, (void *)&(i))
#define SAVELONG(l) goperl_save_int8(_gof, (void *)&(l))
#define SAVESTRLEN(l) goperl_save_int8(_gof, (void *)&(l))
#define SAVEPPTR(p) goperl_save_vptr(_gof, (void **)&(p))
#define save_int(p) goperl_save_i32(_gof, (int32_t *)(p))
#define save_I32(p) goperl_save_i32(_gof, (int32_t *)(p))
#define save_long(p) goperl_save_int8(_gof, (void *)(p))
#define save_iv(p) goperl_save_int8(_gof, (void *)(p))
#define save_strlen(p) goperl_save_int8(_gof, (void *)(p))
#define Perl_save_strlen(p) save_strlen(p)
/* $@ made the empty string, as perl's CLEAR_ERRSV leaves it */
#define CLEAR_ERRSV() STMT_START { sv_setpvs(ERRSV, ""); SvPOK_only(ERRSV); } STMT_END
#define CvSIGNATURE(cv) (CvFLAGS(cv) & CVf_SIGNATURE)
/* the iteration variable of a loop context: the address of its pad slot
 * (the interpreter's memory), of its glob's scalar, or of the reference */
#define CxPADLOOP(c) ((c)->cx_type & CXp_FOR_PAD)
#define CxITERVAR(c)                                                         \
    (CxPADLOOP(c) ? (c)->blk_loop.itervar_u.svp                              \
                  : ((c)->cx_type & CXp_FOR_GV) ? &GvSV((c)->blk_loop.itervar_u.gv) \
                                                : (SV **)&(c)->blk_loop.itervar_u.gv)

/* ---- number scanning (numeric.c) ------------------------------------------ */

static UV goperl_grok_radix(goperl_frame_t *f, const char *start,
                            STRLEN *len_p, I32 *flags, NV *result, int shift)
    __attribute__((unused));
static UV goperl_grok_radix(goperl_frame_t *f, const char *start,
                            STRLEN *len_p, I32 *flags, NV *result,
                            int shift) {
    static const char *const names[] = {"", "binary", "", "octal", "hexadecimal"};
    const char *s = start;
    const char *e = start + *len_p;
    const I32 in = *flags;
    const UV max_div = UV_MAX >> shift;
    const unsigned base = 1u << shift;
    UV value = 0;
    NV value_nv = 0;
    int overflowed = 0;
    if (!(in & PERL_SCAN_DISALLOW_PREFIX) && e - s >= 1) {
        const char letter = shift == 4 ? 'x' : shift == 1 ? 'b' : 'o';
        if ((s[0] | 0x20) == letter)
            s++;
        else if (e - s >= 2 && s[0] == '0' && (s[1] | 0x20) == letter)
            s += 2;
    }
    for (; s < e; s++) {
        unsigned digit;
        const char c = *s;
        if (c >= '0' && c <= '9')
            digit = (unsigned)(c - '0');
        else if (shift == 4 && (c | 0x20) >= 'a' && (c | 0x20) <= 'f')
            digit = (unsigned)((c | 0x20) - 'a' + 10);
        else
            digit = base;
        if (digit >= base) {
            if (c == '_' && (in & PERL_SCAN_ALLOW_UNDERSCORES) && s + 1 < e) {
                const char n = s[1];
                unsigned nd = (n >= '0' && n <= '9')
                                  ? (unsigned)(n - '0')
                                  : (shift == 4 && (n | 0x20) >= 'a' && (n | 0x20) <= 'f')
                                        ? 10u
                                        : base;
                if (nd < base) continue;
            }
            if (!(in & PERL_SCAN_SILENT_ILLDIGIT) && ckWARN(WARN_DIGIT)) {
                if (shift == 3 && (c == '8' || c == '9'))
                    goperl_warnf(f, "Illegal octal digit '%c' ignored", c);
                else if (shift != 3)
                    goperl_warnf(f, "Illegal %s digit '%c' ignored", names[shift], c);
            }
            break;
        }
        if (!overflowed) {
            if (value <= max_div) {
                value = (value << shift) | digit;
                continue;
            }
            overflowed = 1;
            value_nv = (NV)value;
            if (ckWARN_d(WARN_OVERFLOW))
                goperl_warnf(f, "Integer overflow in %s number", names[shift]);
        }
        value_nv = value_nv * (NV)base + (NV)digit;
    }
    *len_p = (STRLEN)(s - start);
    if (!overflowed) {
        *flags = 0;
        return value;
    }
    *flags = PERL_SCAN_GREATER_THAN_UV_MAX;
    if (result) *result = value_nv;
    return UV_MAX;
}
#define grok_hex(start, len_p, flags, result) \
    goperl_grok_radix(_gof, (start), (len_p), (flags), (result), 4)
#define grok_oct(start, len_p, flags, result) \
    goperl_grok_radix(_gof, (start), (len_p), (flags), (result), 3)
#define grok_bin(start, len_p, flags, result) \
    goperl_grok_radix(_gof, (start), (len_p), (flags), (result), 1)
static NV goperl_scan_radix(goperl_frame_t *f, const char *start, STRLEN len,
                            STRLEN *retlen, int shift) __attribute__((unused));
static NV goperl_scan_radix(goperl_frame_t *f, const char *start, STRLEN len,
                            STRLEN *retlen, int shift) {
    NV rnv = 0;
    I32 flags = PERL_SCAN_ALLOW_UNDERSCORES;
    const UV ruv = goperl_grok_radix(f, start, &len, &flags, &rnv, shift);
    *retlen = len;
    return (flags & PERL_SCAN_GREATER_THAN_UV_MAX) ? rnv : (NV)ruv;
}
#define scan_hex(start, len, retlen) goperl_scan_radix(_gof, (start), (len), (retlen), 4)
#define scan_oct(start, len, retlen) goperl_scan_radix(_gof, (start), (len), (retlen), 3)
#define scan_bin(start, len, retlen) goperl_scan_radix(_gof, (start), (len), (retlen), 1)

/* ---- UTF-8 inspection (utf8.h, inline.h) ---------------------------------- */

#define UVCHR_SKIP(uv) \
    ({ GOPERL_ONCE(goperl_UVCHR_SKIP_uv_, uv); \
       (UV)goperl_UVCHR_SKIP_uv_ < 0x80 ? 1 : (UV)goperl_UVCHR_SKIP_uv_ < 0x800 ? 2 : (UV)goperl_UVCHR_SKIP_uv_ < 0x10000 ? 3 : (UV)goperl_UVCHR_SKIP_uv_ < 0x200000 ? 4 : (UV)goperl_UVCHR_SKIP_uv_ < 0x4000000 ? 5 : (UV)goperl_UVCHR_SKIP_uv_ < 0x80000000 ? 6 : (UV)goperl_UVCHR_SKIP_uv_ < ((UV)1 << 36) ? 7 : 13; })
#define UNISKIP(uv) UVCHR_SKIP(uv)
#define OFFUNISKIP(uv) UVCHR_SKIP(uv)
#define utf8_to_uvchr(s, retlen) \
    ({ GOPERL_ONCE(goperl_utf8_to_uvchr_s_, s); \
       utf8_to_uvchr_buf((goperl_utf8_to_uvchr_s_), (const U8 *)(goperl_utf8_to_uvchr_s_) + UTF8SKIP(goperl_utf8_to_uvchr_s_), (retlen)); })
#define utf8_to_uvuni(s, retlen) utf8_to_uvchr((s), (retlen))
#define utf8_to_uvuni_buf(s, e, retlen) utf8_to_uvchr_buf((s), (e), (retlen))
#define utf8n_to_uvuni(s, len, retlen, flags) \
    utf8n_to_uvchr((s), (len), (retlen), (flags))
static bool goperl_is_utf8_string_loclen(const U8 *s, STRLEN len,
                                         const U8 **ep, STRLEN *el)
    __attribute__((unused));
static bool goperl_is_utf8_string_loclen(const U8 *s, STRLEN len,
                                         const U8 **ep, STRLEN *el) {
    const U8 *p = s;
    const U8 *const end = s + len;
    STRLEN chars = 0;
    while (p < end) {
        STRLEN n = UTF8SKIP(p);
        if (n > (STRLEN)(end - p) || !goperl_is_utf8_string(p, n)) break;
        p += n;
        chars++;
    }
    if (ep) *ep = p;
    if (el) *el = chars;
    return p == end;
}
#define is_utf8_string_loclen(s, len, ep, el) \
    goperl_is_utf8_string_loclen((const U8 *)(s), (STRLEN)(len), (ep), (el))
#define is_utf8_string_loc(s, len, ep) \
    goperl_is_utf8_string_loclen((const U8 *)(s), (STRLEN)(len), (ep), 0)
static bool goperl_is_invariant_string(const U8 *s, STRLEN len)
    __attribute__((unused));
static bool goperl_is_invariant_string(const U8 *s, STRLEN len) {
    for (STRLEN i = 0; i < len; i++)
        if (s[i] & 0x80) return 0;
    return 1;
}
#define is_utf8_invariant_string(s, len) \
    goperl_is_invariant_string((const U8 *)(s), (STRLEN)(len))
/* ... and where the first byte that is not lies (the end when there is
 * none) */
static bool goperl_is_invariant_string_loc(const U8 *s, STRLEN len, const U8 **ep)
    __attribute__((unused));
static bool goperl_is_invariant_string_loc(const U8 *s, STRLEN len, const U8 **ep) {
    for (STRLEN i = 0; i < len; i++)
        if (s[i] & 0x80) {
            if (ep) *ep = s + i;
            return 0;
        }
    if (ep) *ep = s + len;
    return 1;
}
#define is_utf8_invariant_string_loc(s, len, ep) \
    goperl_is_invariant_string_loc((const U8 *)(s), (STRLEN)(len), (const U8 **)(ep))
/* grok_atoUV (numeric.c): the decimal digits at pv as a UV, without a
 * sign, leading zeros (other than "0" itself), or overflow; endptr, when
 * not NULL, bounds the digits on entry and says where they ended. */
static bool goperl_grok_atoUV(const char *pv, UV *valptr, const char **endptr)
    __attribute__((unused));
static bool goperl_grok_atoUV(const char *pv, UV *valptr, const char **endptr) {
    const char *s = pv;
    const char *const end = endptr ? *endptr : pv + __builtin_strlen(pv);
    if (!end || end <= s || *s < '0' || *s > '9') return 0;
    UV val = (UV)(*s++ - '0');
    if (s < end && *s >= '0' && *s <= '9') {
        if (val == 0) return 0; /* a leading zero */
        while (s < end && *s >= '0' && *s <= '9') {
            const UV digit = (UV)(*s++ - '0');
            if (val < UV_MAX / 10 || (val == UV_MAX / 10 && digit <= UV_MAX % 10))
                val = val * 10 + digit;
            else
                return 0;
        }
    }
    if (endptr == NULL) {
        if (*s) return 0;
    } else {
        *endptr = s;
    }
    *valptr = val;
    return 1;
}
#define grok_atoUV(pv, valptr, endptr) goperl_grok_atoUV((pv), (valptr), (endptr))
#define isinfnan(nv) (isinf((NV)(nv)) || isnan((NV)(nv)))
#define is_ascii_string(s, len) is_utf8_invariant_string(s, len)
#define is_invariant_string(s, len) is_utf8_invariant_string(s, len)
/* bytes_from_utf8 (utf8.c): a new buffer with the native bytes when every
 * character fits one, the argument itself otherwise. */
static U8 *goperl_bytes_from_utf8(const U8 *s, STRLEN *lenp, bool *is_utf8p)
    __attribute__((unused));
static U8 *goperl_bytes_from_utf8(const U8 *s, STRLEN *lenp, bool *is_utf8p) {
    if (!*is_utf8p) return (U8 *)s;
    const U8 *const end = s + *lenp;
    for (const U8 *p = s; p < end; p++) {
        if (*p < 0x80) continue;
        if ((*p != 0xC2 && *p != 0xC3) || p + 1 >= end || (p[1] & 0xC0) != 0x80)
            return (U8 *)s;
        p++;
    }
    U8 *d = (U8 *)malloc(*lenp + 1);
    U8 *o = d;
    if (!d) return (U8 *)s;
    for (const U8 *p = s; p < end; p++) {
        if (*p < 0x80) {
            *o++ = *p;
        } else {
            *o++ = (U8)(((p[0] & 0x03) << 6) | (p[1] & 0x3F));
            p++;
        }
    }
    *o = '\0';
    *lenp = (STRLEN)(o - d);
    *is_utf8p = 0;
    return d;
}
#define bytes_from_utf8(s, lenp, is_utf8p) \
    goperl_bytes_from_utf8((const U8 *)(s), (lenp), (is_utf8p))

/* ---- more of the stack, string and message vocabulary -------------------- */

#define XPUSHi(i) XPUSHs(sv_2mortal(newSViv(i)))
#define XPUSHu(u) XPUSHs(sv_2mortal(newSVuv(u)))
#define XPUSHn(n) XPUSHs(sv_2mortal(newSVnv(n)))
#define XPUSHp(p, l) XPUSHs(sv_2mortal(newSVpvn((p), (l))))
#define PUSHmortal PUSHs(sv_newmortal())
#define XPUSHmortal XPUSHs(sv_newmortal())
#define dTARG SV *targ
#define SvREFCNT_inc_void_NN(sv) ((void)SvREFCNT_inc(sv))
static char *goperl_sv_pv_force_nolen(goperl_frame_t *f, SV *sv)
    __attribute__((unused));
static char *goperl_sv_pv_force_nolen(goperl_frame_t *f, SV *sv) {
    STRLEN len = 0;
    return goperl_sv_pv_force(f, sv, &len);
}
#define SvPV_force_nolen(sv) goperl_sv_pv_force_nolen(_gof, (SV *)(sv))
#define SvPV_force_mutable(sv, lenvar) SvPV_force(sv, lenvar)
#define SvPV_mutable(sv, lenvar) SvPV(sv, lenvar)
/* The bytes of a string: a string held as UTF-8 is converted to its byte
 * form first, and one with a character above 255 has none ("Wide character"). */
static char *goperl_sv_pvbyte(goperl_frame_t *f, SV *sv, STRLEN *lenp,
                              U32 flags) __attribute__((unused));
static char *goperl_sv_pvbyte(goperl_frame_t *f, SV *sv, STRLEN *lenp,
                              U32 flags) {
    return goperl_str_op(f, GOPERL_OP_SV_CALL, GOPERL_TOK(sv), flags, GOPERL_SVCALL_2PVBYTE,
                         lenp);
}
#define SvPVbyte(sv, lenvar) \
    GOPERL_PV_LEN(lenvar, goperl_sv_pvbyte(_gof, (SV *)(sv), (STRLEN *)&goperl_pvlen_, SV_GMAGIC))
#define SvPVbyte_nolen(sv) goperl_sv_pvbyte(_gof, (SV *)(sv), 0, SV_GMAGIC)
#define SvPVbytex(sv, lenvar) SvPVbyte(sv, lenvar)
#define SvPVbytex_nolen(sv) SvPVbyte_nolen(sv)
#define SvPVbyte_nomg(sv, lenvar) \
    GOPERL_PV_LEN(lenvar, goperl_sv_pvbyte(_gof, (SV *)(sv), (STRLEN *)&goperl_pvlen_, 0))
#define sv_2pvbyte(sv, lenp) \
    goperl_sv_pvbyte(_gof, (SV *)(sv), (lenp), SV_GMAGIC)
#define sv_2pvbyte_nolen(sv) SvPVbyte_nolen(sv)
#define sv_2pvbyte_flags(sv, lenp, flags) \
    goperl_sv_pvbyte(_gof, (SV *)(sv), (lenp), (U32)(flags))
#define sv_pvbyte(sv) SvPVbyte_nolen(sv)

/* The string of an SV in an encoding — UTF-8, or bytes — as the
 * interpreter gives it: get magic runs once or not at all (the _nomg
 * forms), a value that cannot be changed in place is converted in a
 * mortal copy, and the "or null" forms have no string for an undefined
 * value. how is GOPERL_PV_UTF8 or GOPERL_PV_BYTES, with GOPERL_PV_OR_NULL
 * for those; flags is SV_GMAGIC or 0. */
#define GOPERL_PV_UTF8 1u
#define GOPERL_PV_BYTES 2u
#define GOPERL_PV_OR_NULL 0x100u
static char *goperl_sv_pv_as(goperl_frame_t *f, SV *sv, STRLEN *lenp,
                             uint32_t how, U32 flags) __attribute__((unused));
static char *goperl_sv_pv_as(goperl_frame_t *f, SV *sv, STRLEN *lenp,
                             uint32_t how, U32 flags) {
    /* b = SV_GMAGIC or 0, hi = how */
    return goperl_str_op(f, GOPERL_OP_SV_CALL, GOPERL_TOK(sv), flags & SV_GMAGIC,
                         GOPERL_SVCALL_PV_AS | ((uint64_t)how << 32), lenp);
}
/* The SV made a string of its own that may be written to, as it is (how 0)
 * or in an encoding. */
static char *goperl_sv_pv_force_as(goperl_frame_t *f, SV *sv, STRLEN *lenp,
                                   uint32_t how, U32 flags)
    __attribute__((unused));
static char *goperl_sv_pv_force_as(goperl_frame_t *f, SV *sv, STRLEN *lenp,
                                   uint32_t how, U32 flags) {
    /* b = the flags, hi = how */
    return goperl_str_op(f, GOPERL_OP_SV_CALL, GOPERL_TOK(sv), flags,
                         GOPERL_SVCALL_PV_FORCE_AS | ((uint64_t)how << 32), lenp);
}
#define SvPVutf8(sv, lenvar) \
    GOPERL_PV_LEN(lenvar, goperl_sv_pv_as(_gof, (SV *)(sv), (STRLEN *)&goperl_pvlen_, GOPERL_PV_UTF8, SV_GMAGIC))
#define SvPVutf8_nomg(sv, lenvar) \
    GOPERL_PV_LEN(lenvar, goperl_sv_pv_as(_gof, (SV *)(sv), (STRLEN *)&goperl_pvlen_, GOPERL_PV_UTF8, 0))
#define SvPVutf8_nolen(sv) \
    goperl_sv_pv_as(_gof, (SV *)(sv), 0, GOPERL_PV_UTF8, SV_GMAGIC)
#define SvPVutf8x(sv, lenvar) SvPVutf8(sv, lenvar)
#define SvPVutf8_or_null(sv, lenvar) \
    GOPERL_PV_LEN(lenvar, goperl_sv_pv_as(_gof, (SV *)(sv), (STRLEN *)&goperl_pvlen_,              GOPERL_PV_UTF8 | GOPERL_PV_OR_NULL, SV_GMAGIC))
#define SvPVutf8_or_null_nomg(sv, lenvar) \
    GOPERL_PV_LEN(lenvar, goperl_sv_pv_as(_gof, (SV *)(sv), (STRLEN *)&goperl_pvlen_,              GOPERL_PV_UTF8 | GOPERL_PV_OR_NULL, 0))
#define SvPVbyte_or_null(sv, lenvar) \
    GOPERL_PV_LEN(lenvar, goperl_sv_pv_as(_gof, (SV *)(sv), (STRLEN *)&goperl_pvlen_,              GOPERL_PV_BYTES | GOPERL_PV_OR_NULL, SV_GMAGIC))
#define SvPVbyte_or_null_nomg(sv, lenvar) \
    GOPERL_PV_LEN(lenvar, goperl_sv_pv_as(_gof, (SV *)(sv), (STRLEN *)&goperl_pvlen_,              GOPERL_PV_BYTES | GOPERL_PV_OR_NULL, 0))
#define sv_2pvutf8_flags(sv, lenp, flags) \
    goperl_sv_pv_as(_gof, (SV *)(sv), (lenp), GOPERL_PV_UTF8, (U32)(flags))
#define sv_2pvutf8(sv, lenp) sv_2pvutf8_flags(sv, lenp, SV_GMAGIC)
#define sv_2pvutf8_nolen(sv) SvPVutf8_nolen(sv)
#define sv_pvutf8(sv) SvPVutf8_nolen(sv)
#define sv_2pv_nomg(sv, lenp) sv_2pv_flags(sv, lenp, 0)
#define SvPV_flags_mutable(sv, lenvar, flags) SvPV_flags(sv, lenvar, flags)
#define SvPV_flags_const_nolen(sv, flags) \
    ((const char *)goperl_sv_pv_flags(_gof, (SV *)(sv), 0, (U32)(flags)))

#define SvPV_force_flags(sv, lenvar, flags) \
    GOPERL_PV_LEN(lenvar, goperl_sv_pv_force_as(_gof, (SV *)(sv), (STRLEN *)&goperl_pvlen_, 0, (U32)(flags)))
#define SvPV_force_flags_nolen(sv, flags) \
    goperl_sv_pv_force_as(_gof, (SV *)(sv), 0, 0, (U32)(flags))
#define SvPV_force_flags_mutable(sv, lenvar, flags) \
    SvPV_force_flags(sv, lenvar, flags)
#define SvPV_force_nomg(sv, lenvar) SvPV_force_flags(sv, lenvar, 0)
#define SvPV_force_nomg_nolen(sv) SvPV_force_flags_nolen(sv, 0)
#define SvPVx_force(sv, lenvar) SvPV_force(sv, lenvar)
#define SvPVutf8_force(sv, lenvar) \
    GOPERL_PV_LEN(lenvar, goperl_sv_pv_force_as(_gof, (SV *)(sv), (STRLEN *)&goperl_pvlen_, GOPERL_PV_UTF8,  SV_GMAGIC))
#define SvPVutf8x_force(sv, lenvar) SvPVutf8_force(sv, lenvar)
#define SvPVbyte_force(sv, lenvar) \
    GOPERL_PV_LEN(lenvar, goperl_sv_pv_force_as(_gof, (SV *)(sv), (STRLEN *)&goperl_pvlen_, GOPERL_PV_BYTES,  SV_GMAGIC))
#define SvPVbytex_force(sv, lenvar) SvPVbyte_force(sv, lenvar)
#define sv_pvn_force_flags(sv, lenp, flags) \
    goperl_sv_pv_force_as(_gof, (SV *)(sv), (lenp), 0, (U32)(flags))
#define sv_pvn_force(sv, lenp) sv_pvn_force_flags(sv, lenp, SV_GMAGIC)
#define sv_pvn_force_nomg(sv, lenp) sv_pvn_force_flags(sv, lenp, 0)
#define sv_pvutf8n_force(sv, lenp)                                  \
    goperl_sv_pv_force_as(_gof, (SV *)(sv), (lenp), GOPERL_PV_UTF8, \
                          SV_GMAGIC)
#define sv_pvbyten_force(sv, lenp)                                   \
    goperl_sv_pv_force_as(_gof, (SV *)(sv), (lenp), GOPERL_PV_BYTES, \
                          SV_GMAGIC)
#define SvPVXx(sv) SvPVX(sv)
#define SvPVCLEAR_FRESH(sv) SvPVCLEAR(sv)

/* my_setenv: the environment of the interpreter (what %ENV's magic and the
 * programs it starts see) and the C environment of the host process (what
 * getenv in native code sees) are two, and both follow. */
static void goperl_my_setenv(goperl_frame_t *f, const char *nam,
                             const char *val) __attribute__((unused));
static void goperl_my_setenv(goperl_frame_t *f, const char *nam,
                             const char *val) {
    size_t nl = __builtin_strlen(nam), vl = val ? __builtin_strlen(val) : 0;
    char *buf = (char *)malloc(nl + vl + 2);
    if (!buf) goperl_croakf(f, "Out of memory!");
    memcpy(buf, nam, nl + 1);
    memcpy(buf + nl + 1, val ? val : "", vl + 1);
    goperl_do_op3(f, GOPERL_OP_SV_STR, 0, ((uint64_t)(val ? 0 : 1) << 32) | (uint32_t)(nl + vl + 2), (uint64_t)(GOPERL_SVSTR_SETENV), buf, nl + vl + 2);
    free(buf);
    if (val)
        setenv(nam, val, 1);
    else
        unsetenv(nam);
}
#define my_setenv(nam, val) goperl_my_setenv(_gof, (nam), (val))
/* sv_uni_display: ssv's characters, escaped for display, into dsv. */
#ifndef UNI_DISPLAY_ISPRINT
#define UNI_DISPLAY_ISPRINT 0x0001
#define UNI_DISPLAY_BACKSLASH 0x0002
#define UNI_DISPLAY_BACKSPACE 0x0004
#define UNI_DISPLAY_QQ (UNI_DISPLAY_ISPRINT | UNI_DISPLAY_BACKSLASH | UNI_DISPLAY_BACKSPACE)
#define UNI_DISPLAY_REGEX (UNI_DISPLAY_ISPRINT | UNI_DISPLAY_BACKSLASH)
#endif
static char *goperl_sv_uni_display(goperl_frame_t *f, SV *dsv, SV *ssv,
                                   STRLEN pvlim, UV flags)
    __attribute__((unused));
static char *goperl_sv_uni_display(goperl_frame_t *f, SV *dsv, SV *ssv,
                                   STRLEN pvlim, UV flags) {
    if (pvlim > 0xFFFFFF) pvlim = 0xFFFFFF;
    /* b = ssv, hi = pvlim<<8 | flags */
    return goperl_str_op(f, GOPERL_OP_SV_CALL, GOPERL_TOK(dsv), GOPERL_TOK(ssv),
                         GOPERL_SVCALL_UNI_DISPLAY |
                             ((((uint64_t)pvlim << 8) | (flags & 0xFF)) << 32),
                         0);
}
#define sv_uni_display(dsv, ssv, pvlim, flags) \
    goperl_sv_uni_display(_gof, (SV *)(dsv), (SV *)(ssv), (STRLEN)(pvlim), (UV)(flags))
/* memCHRs: whether the character is one of those of a string literal. */
#define memCHRs(lit, c) ((const char *)memchr("" lit "", (c), sizeof(lit) - 1))
#define Select_fd_set_t fd_set *
#define PERLDB_ALL                                                   \
    (PERLDBf_SUB | PERLDBf_LINE | PERLDBf_NOOPT | PERLDBf_INTER |    \
     PERLDBf_SUBLINE | PERLDBf_SINGLE | PERLDBf_NAMEEVAL |           \
     PERLDBf_NAMEANON | PERLDBf_SAVESRC)
#define PERLDB_SUB (PL_perldb & PERLDBf_SUB)
#define PERLDB_LINE (PL_perldb & PERLDBf_LINE)
#define PERLDB_NOOPT (PL_perldb & PERLDBf_NOOPT)
#define PERLDB_INTER (PL_perldb & PERLDBf_INTER)
#define PERLDB_SUBLINE (PL_perldb & PERLDBf_SUBLINE)
#define PERLDB_SINGLE (PL_perldb & PERLDBf_SINGLE)
#define PERLDB_SUB_NN (PL_perldb & PERLDBf_NONAME)
#define PERLDB_GOTO (PL_perldb & PERLDBf_GOTO)
#define PERLDB_NAMEEVAL (PL_perldb & PERLDBf_NAMEEVAL)
#define PERLDB_NAMEANON (PL_perldb & PERLDBf_NAMEANON)
#define PERLDB_SAVESRC (PL_perldb & PERLDBf_SAVESRC)
#define hv_ksplit(hv, newmax)                                             \
    ((void)goperl_do_op3(_gof, GOPERL_OP_SV_CALL, GOPERL_TOK(hv), (uint64_t)(int64_t)(newmax), (uint64_t)(GOPERL_SVCALL_HV_KSPLIT), 0, 0))
#define SV_CHECK_THINKFIRST(sv)                                       \
    STMT_START {                                                      \
        GOPERL_ONCE(goperl_SV_CHECK_THINKFIRST_sv_, sv);              \
        if (SvTHINKFIRST(goperl_SV_CHECK_THINKFIRST_sv_))             \
            sv_force_normal_flags(goperl_SV_CHECK_THINKFIRST_sv_, 0); \
    }                                                                 \
    STMT_END
#define SV_CHECK_THINKFIRST_COW_DROP(sv)                              \
    STMT_START {                                                      \
        GOPERL_ONCE(goperl_SV_CHECK_THINKFIRST_COW_DROP_sv_, sv);     \
        if (SvTHINKFIRST(goperl_SV_CHECK_THINKFIRST_COW_DROP_sv_))    \
            sv_force_normal_flags(goperl_SV_CHECK_THINKFIRST_COW_DROP_sv_, SV_COW_DROP_PV); \
    }                                                                 \
    STMT_END
#define newSVpvn_utf8(s, len, u) newSVpvn_flags((s), (len), (u) ? SVf_UTF8 : 0)
/* my_strlcpy (util.c): copies at most size-1 bytes, always terminates, and
 * returns the length of the source. */
static Size_t goperl_strlcpy(char *dst, const char *src, Size_t size)
    __attribute__((unused));
static Size_t goperl_strlcpy(char *dst, const char *src, Size_t size) {
    Size_t length = __builtin_strlen(src);
    if (size > 0) {
        Size_t copy = length >= size ? size - 1 : length;
        memcpy(dst, src, copy);
        dst[copy] = '\0';
    }
    return length;
}
#define my_strlcpy goperl_strlcpy
#define croak_no_modify() croak("%s", PL_no_modify)
#define Perl_croak_no_modify() croak("%s", PL_no_modify)
#define croak_memory_wrap() croak("%s", PL_memory_wrap)
#define Perl_croak_memory_wrap() croak("%s", PL_memory_wrap)
#define croak_nocontext croak
#define die_nocontext croak
#define warn_nocontext warn
#define Perl_warn_nocontext warn
#define Perl_die Perl_croak
#define Perl_newSVpvf(...) goperl_newSVpvf(_gof, __VA_ARGS__)
#define Perl_sv_setpvf(...) goperl_sv_setpvf(_gof, __VA_ARGS__)
#define Perl_sv_catpvf(...) goperl_sv_catpvf(_gof, __VA_ARGS__)
#define warner(cat, ...) Perl_warner(aTHX_(cat), __VA_ARGS__)

/* The fields of a hash key (hv.h); the flags byte follows the key's NUL. */
#define HEK_KEY(hek) ((hek)->hek_key)
#define HEK_LEN(hek) ((hek)->hek_len)
#define HEK_HASH(hek) ((hek)->hek_hash)
static inline __attribute__((unused)) unsigned char *goperl_hek_flags(HEK *hek) {
    return (unsigned char *)hek->hek_key + hek->hek_len + 1;
}
#define HEK_FLAGS(hek) (*goperl_hek_flags((HEK *)(hek)))
#define HEK_UTF8(hek) (HEK_FLAGS(hek) & HVhek_UTF8)
#define HEK_WASUTF8(hek) (HEK_FLAGS(hek) & HVhek_WASUTF8)

/* The stash a COP was compiled in; the interpreter keeps it by name too. */
#define CopSTASH(c) gv_stashpv(CopSTASHPV(c), 0)
#define CopSTASH_eq(c, hv) (CopSTASH(c) == (HV *)(hv))


/* ---- fields of interpreter structures ------------------------------------ */

/* FIELD selectors. */
enum {
    GOPERL_FIELD_REFCNT_SET = 0,
    GOPERL_FIELD_HVAUX_BACKREFS = 1,
    GOPERL_FIELD_HVAUX_RITER = 2,
    GOPERL_FIELD_HVAUX_RITER_SET = 3,
    GOPERL_FIELD_HVAUX_EITER = 4,
    GOPERL_FIELD_HVAUX_EITER_SET = 5,
    GOPERL_FIELD_AV_FILLP = 6,
    GOPERL_FIELD_AV_FILLP_SET = 7,
    GOPERL_FIELD_CV_DEPTH_SET = 8,
    GOPERL_FIELD_LV_TYPE = 9,
    GOPERL_FIELD_LV_TARG = 10,
    GOPERL_FIELD_SHARED_HASH = 11,
    GOPERL_FIELD_OP_SV = 12,
    GOPERL_FIELD_OP_GV = 13,
    GOPERL_FIELD_IS_FREED = 14
};
#define GOPERL_FIELD(sel, tok, b)                                           \
    goperl_op3(GOPERL_OP_FIELD, GOPERL_TOK(tok), (b), sel)

#define VOL volatile

/* The reference count is an lvalue (`--SvREFCNT(sv)`, which frees nothing).
 * A U32, as in the interpreter: the low half of the slot. */
#undef SvREFCNT
#define SvREFCNT(sv) \
    (*(U32 *)goperl_lv_ref(_gof, (SV *)(sv), GOPERL_LV_REFCNT, GOPERL_LV_AT))

/* The type and the target of an lvalue scalar are lvalues
 * (`LvTYPE(sv) = '~'; LvTARG(sv) = SvREFCNT_inc(sv);`). The slot holds
 * eight bytes, of which the type is the first. */
#define LvTYPE(sv) \
    (*(char *)goperl_lv_ref(_gof, (SV *)(sv), GOPERL_LV_LVTYPE, GOPERL_LV_AT))
#define LvTARG(sv) \
    (*(SV **)goperl_lv_ref(_gof, (SV *)(sv), GOPERL_LV_LVTARG, GOPERL_LV_AT))
#define SvSHARED_HASH(sv) ((U32)GOPERL_FIELD(GOPERL_FIELD_SHARED_HASH, sv, 0))
#define SvIS_FREED(sv) (GOPERL_FIELD(GOPERL_FIELD_IS_FREED, sv, 0) != 0)
#define HvRITER_get(hv) ((I32)(int64_t)GOPERL_FIELD(GOPERL_FIELD_HVAUX_RITER, hv, 0))
#define HvRITER_set(hv, r) \
    ((void)GOPERL_FIELD(GOPERL_FIELD_HVAUX_RITER_SET, hv, (uint64_t)(int64_t)(r)))
/* Where the iterator of a hash is: lvalues as on perl (`HvRITER(hv) =
 * saved_riter; HvEITER(hv) = saved_eiter;`). The entry is the
 * interpreter's own, which a module may keep from one call to the next. */
#define HvRITER(hv) \
    (*(I32 *)goperl_lv_ref(_gof, (SV *)(hv), GOPERL_LV_HVRITER, GOPERL_LV_AT))
#define HvEITER(hv) \
    (*(HE **)goperl_lv_ref(_gof, (SV *)(hv), GOPERL_LV_HVEITER, GOPERL_LV_AT))
#define HvEITER_get(hv) \
    goperl_he_at(_gof, GOPERL_FIELD(GOPERL_FIELD_HVAUX_EITER, hv, 0))
#define HvEITER_set(hv, e)                                                  \
    ({ const uint64_t goperl_HvEITER_set_e_ =                               \
           goperl_he_addr(_gof, (const HE *)(e), "HvEITER_set");            \
       (void)GOPERL_FIELD(GOPERL_FIELD_HVAUX_EITER_SET, hv, goperl_HvEITER_set_e_); })

/* ---- ops ----------------------------------------------------------------- */

#define cSVOP cSVOPx(PL_op)
#define cSVOP_sv cSVOPx_sv(PL_op)
#define cSVOPo_sv cSVOPx_sv(o)
#define cGVOPx_gv(o) ((GV *)cSVOPx(o)->op_sv)
#define cGVOP_gv cGVOPx_gv(PL_op)
#define cGVOPo_gv cGVOPx_gv(o)
#define cCOP cCOPx(PL_op)
#define cCOPo cCOPx(o)
#define cPADOPx(o) ((PADOP *)(o))
#undef cPMOPx
#define cPMOPx(o) ((PMOP *)(o))
#define cMETHOPx_meth(o) (cMETHOPx(o)->op_u.op_meth_sv)
#define cMETHOPo_meth cMETHOPx_meth(o)
#define cMETHOP_meth cMETHOPx_meth(PL_op)
typedef struct op PMOP;

/* The class of an op, as an OA_ value: the interpreter's for an op of its
 * own (a custom op has the class it was registered with), and the class of
 * the type for one a module is still building. */
static U32 goperl_op_class(const OP *o) __attribute__((unused));
static U32 goperl_op_class(const OP *o) {
    static const U32 by_class[] = {OA_BASEOP, OA_BASEOP, OA_UNOP, OA_BINOP,
                                   OA_LOGOP, OA_LISTOP, OA_PMOP, OA_SVOP,
                                   OA_PADOP, OA_PVOP_OR_SVOP, OA_LOOP, OA_COP,
                                   OA_METHOP, OA_UNOP_AUX};
    if (o->op_type == OP_CUSTOM) {
        /* the class its pp function was registered with */
        const goperl_hooks_t *h = goperl_hooks_v;
        for (int32_t i = 0; h && i < h->xop_n; i++)
            if (h->xops[i].pp == o->op_ppaddr) return XopENTRY(h->xops[i].xop, xop_class);
    }
    if (o->gop_class && (size_t)(o->gop_class - 1) < sizeof(by_class) / sizeof(by_class[0]))
        return by_class[o->gop_class - 1];
    return o->op_type < OP_max ? (PL_opargs[o->op_type] & OA_CLASS_MASK) : OA_BASEOP;
}
#define OP_CLASS(o) goperl_op_class((const OP *)(o))
#undef OA_LIST
#define OA_LIST 2
#define OPpTRANS_FROM_UTF OPpTRANS_USE_SVOP
#define OPpTRANS_TO_UTF OPpTRANS_CAN_FORCE_UTF8

#define RETURNX(x) return (x, PUTBACK, NORMAL)
#define RETPUSHYES RETURNX(PUSHs(&PL_sv_yes))
#define RETPUSHNO RETURNX(PUSHs(&PL_sv_no))
#define RETPUSHUNDEF RETURNX(PUSHs(&PL_sv_undef))
#define RETSETYES RETURNX(SETs(&PL_sv_yes))
#define RETSETNO RETURNX(SETs(&PL_sv_no))
#define RETSETUNDEF RETURNX(SETs(&PL_sv_undef))

/* ---- COPs ---------------------------------------------------------------- */

/* CALL selectors. */
enum {
    GOPERL_CALL_PV_DISPLAY = 0,
    GOPERL_CALL_PV_ESCAPE = 1,
    GOPERL_CALL_PV_PRETTY = 2,
    GOPERL_CALL_CV_CLONE = 3,
    GOPERL_CALL_NEW_WARNINGS = 4,
    GOPERL_CALL_GETCWD_SV = 5,
    GOPERL_CALL_COP_SET = 6,
    GOPERL_CALL_MESS_SV = 7
};
#define GOPERL_CALL(sel, tok, b)                                            \
    goperl_op3(GOPERL_OP_CALL, GOPERL_TOK(tok), (b), sel)
#define GOPERL_PL_COMPILING 63
#define GOPERL_PL_WARN_ALL 64
#define GOPERL_PL_WARN_NONE 65
#define GOPERL_PL_ERRORS 66
#define GOPERL_PL_SCOPESTACK 67
#define GOPERL_PL_SAVEBEGIN 68
#define GOPERL_PL_MARKSTACK 77

/* PL_compiling is the COP the compiler fills in for the statement it is
 * compiling. */
#define PL_compiling \
    (*goperl_cop_shadow(_gof, GOPERL_PLVAR(GOPERL_PL_COMPILING)))
#define IN_PERL_COMPILETIME cBOOL(PL_curcop == &PL_compiling)
#define IN_PERL_RUNTIME cBOOL(PL_curcop != &PL_compiling)

/* The warnings of a COP are a buffer of the interpreter's, one of its two
 * markers, or NULL: cop_warnings holds what the interpreter holds. */
#undef pWARN_STD
#undef pWARN_ALL
#undef pWARN_NONE
#define pWARN_STD ((char *)0)
#define pWARN_ALL \
    ((char *)goperl_guest_mem(_gof, GOPERL_PLVAR(GOPERL_PL_WARN_ALL)))
#define pWARN_NONE \
    ((char *)goperl_guest_mem(_gof, GOPERL_PLVAR(GOPERL_PL_WARN_NONE)))
#define specialWARN(x)                                         \
    ({ const char *const goperl_specialWARN_x_ = (x);          \
       goperl_specialWARN_x_ == pWARN_STD ||                   \
           goperl_specialWARN_x_ == pWARN_ALL ||               \
           goperl_specialWARN_x_ == pWARN_NONE; })
/* The size of a warnings buffer and the buffer with every category on:
 * the interpreter's warnings.h. */
#define WARNsize 21
#define WARN_ALLstring \
    "\125\125\125\125\125\125\125\125\125\125\125\125\125\125\125\125\125\125\125\125\125"
#define WARN_NONEstring \
    "\0\0\0\0\0\0\0\0\0\0\0\0\0\0\0\0\0\0\0\0\0"
static char *goperl_new_warnings_bitfield(goperl_frame_t *f, char *buffer,
                                          const char *bits, STRLEN size)
    __attribute__((unused));
static char *goperl_new_warnings_bitfield(goperl_frame_t *f, char *buffer,
                                          const char *bits, STRLEN size) {
    /* the buffer to reuse is the interpreter's, or none (a marker is not
     * a buffer) */
    const uintptr_t d = (uintptr_t)buffer - (uintptr_t)f->membase;
    uint64_t old = 0;
    if (buffer && d < *f->memsize && buffer != pWARN_ALL && buffer != pWARN_NONE)
        old = (uint64_t)d;
    return (char *)goperl_guest_mem(
        f, goperl_do_op3(f, GOPERL_OP_CALL, old, (uint64_t)size, (uint64_t)(GOPERL_CALL_NEW_WARNINGS), bits, (uint64_t)size));
}
#define Perl_new_warnings_bitfield(...) goperl_new_warnings_bitfield(_gof, __VA_ARGS__)
#define new_warnings_bitfield(buffer, bits, size) \
    goperl_new_warnings_bitfield(_gof, (buffer), (bits), (size))

/* The line of a COP: the interpreter's COP is given it at once, as a
 * module sets the line of PL_curcop for the message it croaks with next
 * (Quote::Code names the line a here-doc began on); a COP of the module's
 * own making keeps it. */
static __attribute__((unused)) U32 goperl_cop_line_set(COP *c, U32 line) {
    COP *const o = goperl_cop_fields(c);
    goperl_frame_t *const f = goperl_cur_frame_v;
    o->cop_line = line;
    if (o->gop && f && goperl_op_is_shadow(f, (const OP *)o)) {
        goperl_do_op3(f, GOPERL_OP_CALL, o->gop, line, GOPERL_CALL_COP_SET | ((uint64_t)0 << 32), 0, 0);
        if (GOPERL_OPENT(o)->base) GOPERL_OPENT(o)->base->cop_line = line;
    }
    return line;
}
#define CopLINE_set(c, l) goperl_cop_line_set((COP *)(c), (U32)(l))
#define CopLINE_inc(c) \
    ({ GOPERL_ONCE(goperl_CopLINE_inc_c_, c); goperl_cop_line_set((COP *)goperl_CopLINE_inc_c_, CopLINE(goperl_CopLINE_inc_c_) + 1); })
#define CopLINE_dec(c) \
    ({ GOPERL_ONCE(goperl_CopLINE_dec_c_, c); goperl_cop_line_set((COP *)goperl_CopLINE_dec_c_, CopLINE(goperl_CopLINE_dec_c_) - 1); })
static void goperl_cop_file_set(goperl_frame_t *f, COP *c, const char *pv)
    __attribute__((unused));
static void goperl_cop_file_set(goperl_frame_t *f, COP *c, const char *pv) {
    if (c->gop)
        goperl_do_op3(f, GOPERL_OP_CALL, c->gop, 0, GOPERL_CALL_COP_SET | ((uint64_t)4 << 32), pv,
                      __builtin_strlen(pv));
    c->cop_file = goperl_intern(pv, __builtin_strlen(pv));
}
#define CopFILE_set(c, pv) goperl_cop_file_set(_gof, (c), (pv))
static void goperl_cop_stash_set(goperl_frame_t *f, COP *c, HV *hv)
    __attribute__((unused));
static void goperl_cop_stash_set(goperl_frame_t *f, COP *c, HV *hv) {
    if (c->gop)
        goperl_do_op3(f, GOPERL_OP_CALL, c->gop, GOPERL_TOK(hv),
                      GOPERL_CALL_COP_SET | ((uint64_t)3 << 32), 0, 0);
    const char *name = hv ? HvNAME(hv) : 0;
    c->cop_stashpv = name ? goperl_intern(name, __builtin_strlen(name)) : "";
}
#define CopSTASH_set(c, hv) goperl_cop_stash_set(_gof, (c), (HV *)(hv))
#define CopFILESV(c) GvSV(CopFILEGV(c))
#define CopFILEAV(c) GvAV(CopFILEGV(c))

/* ---- the context stack ----------------------------------------------------- */

/* (the CX selectors are with the context shadows in perl.h) */
/* dounwind pops contexts: interpreter code that works on the stacks as a
 * pp function left them. */
static void goperl_dounwind(goperl_frame_t *f, I32 cxix) __attribute__((unused));
static void goperl_dounwind(goperl_frame_t *f, I32 cxix) {
    goperl_do_op3(f, GOPERL_OP_CX, 0, (uint64_t)(int64_t)cxix, (uint64_t)(GOPERL_CX_DOUNWIND), 0, 0);
}
#define dounwind(cxix) goperl_dounwind(_gof, (cxix))
#define Perl_dounwind(...) goperl_dounwind(_gof, __VA_ARGS__)
#define cxstack_max \
    ((I32)(int64_t)goperl_op3(GOPERL_OP_CX, 0, 0, GOPERL_CX_MAX))
#define CX_CUR() (&cxstack[cxstack_ix])

/* Contexts native code pushes and pops itself (Syntax::Keyword::Defer
 * pretends an eval {} around a defer block): the interpreter's cx_pushblock
 * and the others, on its current stackinfo, with the stack as native code
 * left it. A context given back to one of them is one of those the SDK
 * shows now (from cx_pushblock or cxstack), and the top one, as perl's
 * functions require. */
static I32 goperl_cx_top_index(goperl_frame_t *f, const PERL_CONTEXT *cx, const char *what) {
    PERL_SI *const si = goperl_si_cur(f);
    if (!cx || cx < si->si_cxstack || cx > si->si_cxstack + si->si_cxix)
        goperl_croakf(f, "%s: native code gave a context that is not one of the "
                         "interpreter's contexts as the SDK shows them now", what);
    return (I32)(cx - si->si_cxstack);
}
static PERL_CONTEXT *goperl_cx_pushblock(goperl_frame_t *f, U8 type, U8 gimme, SV **sp,
                                         I32 saveix) __attribute__((unused));
static PERL_CONTEXT *goperl_cx_pushblock(goperl_frame_t *f, U8 type, U8 gimme, SV **sp,
                                         I32 saveix) {
    (void)sp; /* the stack native code left, which the interpreter has now */
    const goperl_gword_t w[3] = {type, gimme, (goperl_gword_t)(int64_t)saveix};
    const I32 ix = (I32)(int64_t)goperl_do_op3(f, GOPERL_OP_CX, 0, sizeof w,
                                               (uint64_t)(GOPERL_CX_PUSHBLOCK),
                                               (const char *)w, sizeof w);
    return &goperl_si_cur(f)->si_cxstack[ix];
}
static void goperl_cx_pusheval(goperl_frame_t *f, PERL_CONTEXT *cx, OP *retop, SV *namesv)
    __attribute__((unused));
static void goperl_cx_pusheval(goperl_frame_t *f, PERL_CONTEXT *cx, OP *retop, SV *namesv) {
    const goperl_gword_t w[3] = {(goperl_gword_t)goperl_cx_top_index(f, cx, "cx_pusheval"),
                                 retop ? retop->gop : 0, GOPERL_TOK(namesv)};
    goperl_do_op3(f, GOPERL_OP_CX, 0, sizeof w, (uint64_t)(GOPERL_CX_PUSHEVAL),
                  (const char *)w, sizeof w);
}
static void goperl_cx_pop(goperl_frame_t *f, PERL_CONTEXT *cx, int sel, const char *what)
    __attribute__((unused));
static void goperl_cx_pop(goperl_frame_t *f, PERL_CONTEXT *cx, int sel, const char *what) {
    const I32 ix = goperl_cx_top_index(f, cx, what);
    goperl_do_op3(f, GOPERL_OP_CX, 0, (uint64_t)(int64_t)ix, (uint64_t)sel, 0, 0);
}
#define cx_pushblock(type, gimme, sp, saveix) \
    goperl_cx_pushblock(_gof, (U8)(type), (U8)(gimme), (SV **)(sp), (I32)(saveix))
static void goperl_cx_pushtry(goperl_frame_t *f, PERL_CONTEXT *cx, OP *retop)
    __attribute__((unused));
static void goperl_cx_pushtry(goperl_frame_t *f, PERL_CONTEXT *cx, OP *retop) {
    const goperl_gword_t w[2] = {(goperl_gword_t)goperl_cx_top_index(f, cx, "cx_pushtry"),
                                 retop ? retop->gop : 0};
    goperl_do_op3(f, GOPERL_OP_CX, 0, sizeof w, (uint64_t)(GOPERL_CX_PUSHTRY), (const char *)w,
                  sizeof w);
}
#define cx_pushtry(cx, retop) goperl_cx_pushtry(_gof, (PERL_CONTEXT *)(cx), (OP *)(retop))
#define cx_pusheval(cx, retop, namesv) \
    goperl_cx_pusheval(_gof, (PERL_CONTEXT *)(cx), (OP *)(retop), (SV *)(namesv))
#define cx_popeval(cx) goperl_cx_pop(_gof, (PERL_CONTEXT *)(cx), GOPERL_CX_POPEVAL, "cx_popeval")
#define cx_popblock(cx) \
    goperl_cx_pop(_gof, (PERL_CONTEXT *)(cx), GOPERL_CX_POPBLOCK, "cx_popblock")
#define CX_POP(cx) goperl_cx_pop(_gof, (PERL_CONTEXT *)(cx), GOPERL_CX_POP, "CX_POP")

/* ---- the regular expression API ------------------------------------------- */

/* REGEXP selectors. */
enum {
    GOPERL_RX_OF_SV = 0,
    GOPERL_RX_COMP = 1,
    GOPERL_RX_EXEC = 2,
    GOPERL_RX_BODY = 3,
    GOPERL_RX_OFFS = 4,
    GOPERL_RX_COMP_PV = 5
};
#define GOPERL_RX(sel, tok, b)                                              \
    goperl_op3(GOPERL_OP_REGEXP, GOPERL_TOK(tok), (b), sel)

/* A REGEXP is a value of the interpreter, as an SV is. Its body, which
 * perl's macros read through ReANY (and old code through SvANY), is a copy
 * here: what the interpreter held when the macro ran. */
typedef struct goperl_sv REGEXP;
typedef struct regexp_engine {
    int goperl_opaque;
} regexp_engine;
typedef struct regexp_paren_pair {
    SSize_t start;
    SSize_t end;
    SSize_t start_tmp;
} regexp_paren_pair;
typedef struct regexp {
    const regexp_engine *engine;
    REGEXP *mother_re;
    HV *paren_names;
    U32 extflags;
    U32 nparens;
    SSize_t minlen;
    SSize_t minlenret;
    STRLEN gofs;
    U32 intflags;
    U32 lastparen;
    U32 lastcloseparen;
    regexp_paren_pair *offs;
    char *subbeg;
    SSize_t sublen;
    SSize_t suboffset;
    SSize_t subcoffset;
    U32 compflags;
    char *precomp;
    STRLEN prelen;
    char *wrapped;
    STRLEN wraplen;
} regexp;

/* The body of a regexp (ReANY, SvANY) is one struct, as perl's is: code
 * keeps the pointer across matches and reads the offsets of each match there
 * (Lucy's RegexTokenizer takes it once and runs pregexec in a loop). So a
 * regexp has one body for the call of native code, in the arena as the
 * other shadows of the call are; ReANY fills it again from the interpreter,
 * and so does every match of that regexp made through the perl API. */
typedef struct goperl_rxbody {
    struct goperl_rxbody *next;
    uint64_t tok;
    U32 cap; /* the groups r has room for (nparens + 1) */
    regexp r; /* then the offsets of the groups */
} goperl_rxbody_t;
#define goperl_rxbodies_v (*(goperl_rxbody_t **)&GOPERL_SHARED->rx_bodies)
static goperl_rxbody_t *goperl_rxbody_find(uint64_t tok) __attribute__((unused));
static goperl_rxbody_t *goperl_rxbody_find(uint64_t tok) {
    for (goperl_rxbody_t *b = goperl_rxbodies_v; b; b = b->next)
        if (b->tok == tok) return b;
    return 0;
}
static regexp *goperl_reany(goperl_frame_t *f, REGEXP *rx) __attribute__((unused));
static regexp *goperl_reany(goperl_frame_t *f, REGEXP *rx) {
    const goperl_gword_t *w = (const goperl_gword_t *)goperl_guest_mem(
        f, goperl_do_op3(f, GOPERL_OP_REGEXP, GOPERL_TOK(rx), 0, (uint64_t)(GOPERL_RX_BODY), 0, 0));
    if (!w) return 0;
    goperl_gword_t c[20];
    memcpy(c, w, sizeof c);
    goperl_rxbody_t *body = goperl_rxbody_find(GOPERL_TOK(rx));
    if (!body || body->cap < (U32)c[4] + 1) {
        /* a body too small for the groups of a regexp now at the address
         * stays where it is: code may still hold it */
        const size_t n = sizeof(goperl_rxbody_t) + sizeof(regexp_paren_pair) * ((size_t)c[4] + 1);
        goperl_rxbody_t *nb = (goperl_rxbody_t *)goperl_arena_alloc(n);
        if (!nb) goperl_croakf(f, "Out of memory!");
        memset(nb, 0, n);
        nb->tok = GOPERL_TOK(rx);
        nb->cap = (U32)c[4] + 1;
        nb->next = goperl_rxbodies_v;
        goperl_rxbodies_v = nb;
        body = nb;
    }
    regexp *r = &body->r;
    /* an engine is known by its address in the interpreter */
    r->engine = (const regexp_engine *)(uintptr_t)c[0];
    r->mother_re = (REGEXP *)GOPERL_SV(c[1]);
    r->paren_names = (HV *)GOPERL_SV(c[2]);
    r->extflags = (U32)c[3];
    r->nparens = (U32)c[4];
    r->minlen = (SSize_t)(int64_t)c[5];
    r->minlenret = (SSize_t)(int64_t)c[6];
    r->gofs = (STRLEN)c[7];
    r->intflags = (U32)c[8];
    r->lastparen = (U32)c[9];
    r->lastcloseparen = (U32)c[10];
    r->subbeg = c[11] ? (char *)goperl_guest_mem(f, c[11]) : 0;
    r->sublen = (SSize_t)(int64_t)c[12];
    r->suboffset = (SSize_t)(int64_t)c[13];
    r->subcoffset = (SSize_t)(int64_t)c[14];
    r->compflags = (U32)c[15];
    r->precomp = c[16] ? (char *)goperl_guest_mem(f, c[16]) : 0;
    r->prelen = (STRLEN)c[17];
    r->wrapped = c[18] ? (char *)goperl_guest_mem(f, c[18]) : 0;
    r->wraplen = (STRLEN)c[19];
    r->offs = (regexp_paren_pair *)(r + 1);
    for (U32 i = 0; i <= r->nparens; i++) {
        /* -> the start of the group; second: its end */
        uint64_t end = 0;
        const uint64_t start = goperl_do_opx(f, GOPERL_OP_REGEXP, GOPERL_TOK(rx), i,
                                             GOPERL_RX_OFFS, 0, 0, &end);
        r->offs[i].start = (SSize_t)(int64_t)start;
        r->offs[i].end = (SSize_t)(int64_t)end;
        r->offs[i].start_tmp = -1;
    }
    return r;
}
#define ReANY(rx) goperl_reany(_gof, (REGEXP *)(rx))

#define SvRX(sv) ((REGEXP *)GOPERL_SV(GOPERL_RX(GOPERL_RX_OF_SV, sv, 0)))
#define pregcomp(pattern, flags) \
    ((REGEXP *)GOPERL_SV(GOPERL_RX(GOPERL_RX_COMP, pattern, (U32)(flags))))
#define re_compile(pattern, flags) pregcomp((pattern), (flags))
#define CALLREGCOMP(pattern, flags) pregcomp((pattern), (flags))
#define pregfree(rx) SvREFCNT_dec((SV *)(rx))
#define ReREFCNT_dec(rx) SvREFCNT_dec((SV *)(rx))
#define ReREFCNT_inc(rx) ((REGEXP *)SvREFCNT_inc((SV *)(rx)))
#define RX_ENGINE(rx) (ReANY(rx)->engine)
#define RX_EXTFLAGS(rx) (ReANY(rx)->extflags)
#define RX_COMPFLAGS(rx) (ReANY(rx)->compflags)
#define RX_PRECOMP(rx) (ReANY(rx)->precomp)
#define RX_PRECOMP_const(rx) ((const char *)RX_PRECOMP(rx))
#define RX_PRELEN(rx) (ReANY(rx)->prelen)
#define RX_WRAPPED(rx) (ReANY(rx)->wrapped)
#define RX_WRAPPED_const(rx) ((const char *)RX_WRAPPED(rx))
#define RX_WRAPLEN(rx) (ReANY(rx)->wraplen)
#define RX_NPARENS(rx) (ReANY(rx)->nparens)
#define RX_LASTPAREN(rx) (ReANY(rx)->lastparen)
#define RX_LASTCLOSEPAREN(rx) (ReANY(rx)->lastcloseparen)
#define RX_OFFS(rx) (ReANY(rx)->offs)
#define RX_SUBBEG(rx) (ReANY(rx)->subbeg)
#define RX_SUBLEN(rx) (ReANY(rx)->sublen)
#define RX_SUBOFFSET(rx) (ReANY(rx)->suboffset)
#define RX_SUBCOFFSET(rx) (ReANY(rx)->subcoffset)
#define RX_MINLEN(rx) (ReANY(rx)->minlen)
#define RX_MINLENRET(rx) (ReANY(rx)->minlenret)
#define RX_GOFS(rx) (ReANY(rx)->gofs)
/* The macros of regexp.h that take the body (a regexp, from ReANY), and
 * the offsets as perl 5.38 and later name them. */
#define RXp_ENGINE(prog) ((prog)->engine)
#define RXp_EXTFLAGS(prog) ((prog)->extflags)
#define RXp_COMPFLAGS(prog) ((prog)->compflags)
#define RXp_NPARENS(prog) ((prog)->nparens)
#define RXp_LASTPAREN(prog) ((prog)->lastparen)
#define RXp_LASTCLOSEPAREN(prog) ((prog)->lastcloseparen)
#define RXp_OFFSp(prog) ((prog)->offs)
#define RXp_OFFS_START(prog, n) (RXp_OFFSp(prog)[(n)].start)
#define RXp_OFFS_END(prog, n) (RXp_OFFSp(prog)[(n)].end)
#define RXp_OFFS_VALID(prog, n) \
    (RXp_OFFSp(prog)[(n)].end != -1 && RXp_OFFSp(prog)[(n)].start != -1)
#define RXp_SUBBEG(prog) ((prog)->subbeg)
#define RXp_SUBLEN(prog) ((prog)->sublen)
#define RXp_SUBOFFSET(prog) ((prog)->suboffset)
#define RXp_SUBCOFFSET(prog) ((prog)->subcoffset)
#define RXp_MINLEN(prog) ((prog)->minlen)
#define RXp_MINLENRET(prog) ((prog)->minlenret)
#define RXp_GOFS(prog) ((prog)->gofs)
#define RXp_PAREN_NAMES(prog) ((prog)->paren_names)
#define RX_OFFSp(rx) RXp_OFFSp(ReANY(rx))
#define RX_OFFS_START(rx, n) RXp_OFFS_START(ReANY(rx), (n))
#define RX_OFFS_END(rx, n) RXp_OFFS_END(ReANY(rx), (n))
#define RX_OFFS_VALID(rx, n) RXp_OFFS_VALID(ReANY(rx), (n))
#define RX_UTF8(rx) SvUTF8((SV *)(rx))
#define RX_MATCH_UTF8(rx) cBOOL(RX_EXTFLAGS(rx) & RXf_MATCH_UTF8)
/* The character set of a pattern is a field of its flags, where the
 * interpreter has it (RXf_PMf_CHARSET, its lowest bit the shift). */
#define RXf_PMf_CHARSET_SHIFT_ (__builtin_ctz(RXf_PMf_CHARSET))
#define get_regex_charset(flags) \
    ((regex_charset)(((flags) & RXf_PMf_CHARSET) >> RXf_PMf_CHARSET_SHIFT_))
typedef enum {
    REGEX_DEPENDS_CHARSET = 0,
    REGEX_LOCALE_CHARSET,
    REGEX_UNICODE_CHARSET,
    REGEX_ASCII_RESTRICTED_CHARSET,
    REGEX_ASCII_MORE_RESTRICTED_CHARSET
} regex_charset;
#define set_regex_charset(flagsp, cs) \
    (*(flagsp) = (*(flagsp) & ~RXf_PMf_CHARSET) | ((U32)(cs) << RXf_PMf_CHARSET_SHIFT_))
/* The modifiers of a pattern as regexp.h spells them, and the flags a
 * pattern is compiled with (op_reg_common.h): a pattern's modifiers are
 * rebuilt from its flags in the order of INT_PAT_MODS (Sereal::Encoder). */
#define ONCE_PAT_MODS "o"
#define KEEPCOPY_PAT_MODS "p"
#define NOCAPTURE_PAT_MODS "n"
#define EXEC_PAT_MODS "e"
#define LOOP_PAT_MODS "gc"
#define NONDESTRUCT_PAT_MODS "r"
#define LOCALE_PAT_MODS "l"
#define UNICODE_PAT_MODS "u"
#define DEPENDS_PAT_MODS "d"
#define ASCII_RESTRICT_PAT_MODS "a"
#define ASCII_MORE_RESTRICT_PAT_MODS "aa"
#define STD_PAT_MODS "msixxn"
#define CHARSET_PAT_MODS ASCII_RESTRICT_PAT_MODS DEPENDS_PAT_MODS LOCALE_PAT_MODS UNICODE_PAT_MODS
#define INT_PAT_MODS STD_PAT_MODS KEEPCOPY_PAT_MODS
#define EXT_PAT_MODS ONCE_PAT_MODS KEEPCOPY_PAT_MODS NOCAPTURE_PAT_MODS
#define QR_PAT_MODS STD_PAT_MODS EXT_PAT_MODS CHARSET_PAT_MODS
#define M_PAT_MODS QR_PAT_MODS LOOP_PAT_MODS
#define S_PAT_MODS M_PAT_MODS EXEC_PAT_MODS NONDESTRUCT_PAT_MODS
#define RXf_PMf_COMPILETIME                                                            \
    (RXf_PMf_MULTILINE | RXf_PMf_SINGLELINE | RXf_PMf_FOLD | RXf_PMf_EXTENDED |          \
     RXf_PMf_EXTENDED_MORE | RXf_PMf_KEEPCOPY | RXf_PMf_NOCAPTURE | RXf_PMf_CHARSET |    \
     RXf_PMf_STRICT)
#define RXf_PMf_FLAGCOPYMASK (RXf_PMf_COMPILETIME | RXf_PMf_SPLIT)

/* pregexec matches in a string of the interpreter's memory (the buffer of
 * the SV it is given, as a rule). A string that lies elsewhere is matched
 * in a copy the interpreter keeps for the match, and the offsets it reports
 * are the same either way. */
static I32 goperl_pregexec(goperl_frame_t *f, REGEXP *rx, char *stringarg,
                           char *strend, char *strbeg, SSize_t minend,
                           SV *screamer, U32 nosave) __attribute__((unused));
static I32 goperl_pregexec(goperl_frame_t *f, REGEXP *rx, char *stringarg,
                           char *strend, char *strbeg, SSize_t minend,
                           SV *screamer, U32 nosave) {
    goperl_gword_t w[6];
    const uintptr_t b = (uintptr_t)strbeg - (uintptr_t)f->membase;
    const uintptr_t e = (uintptr_t)strend - (uintptr_t)f->membase;
    if (b < *f->memsize && e < *f->memsize) {
        w[0] = (goperl_gword_t)((uintptr_t)stringarg - (uintptr_t)f->membase);
        w[1] = (goperl_gword_t)e;
        w[2] = (goperl_gword_t)b;
    } else {
        SV *copy = sv_2mortal(newSVpvn(strbeg, (STRLEN)(strend - strbeg)));
        if (screamer && SvUTF8(screamer)) SvUTF8_on(copy);
        const uint64_t base = goperl_do_op(f, GOPERL_OP_SV_PVX_RAW, GOPERL_TOK(copy), 0, 0, 0);
        w[2] = base;
        w[0] = base + (uint64_t)(stringarg - strbeg);
        w[1] = base + (uint64_t)(strend - strbeg);
        screamer = copy;
    }
    w[3] = (goperl_gword_t)(int64_t)minend;
    w[4] = GOPERL_TOK(screamer);
    w[5] = nosave;
    const I32 r = (I32)(int64_t)goperl_do_op3(f, GOPERL_OP_REGEXP, GOPERL_TOK(rx), sizeof w, (uint64_t)(GOPERL_RX_EXEC), (const char *)w, sizeof w);
    /* the body code holds says what this match found */
    if (goperl_rxbody_find(GOPERL_TOK(rx))) (void)goperl_reany(f, rx);
    return r;
}
#define pregexec(rx, stringarg, strend, strbeg, minend, screamer, nosave)   \
    goperl_pregexec(_gof, (REGEXP *)(rx), (char *)(stringarg),              \
                    (char *)(strend), (char *)(strbeg), (SSize_t)(minend),  \
                    (SV *)(screamer), (U32)(nosave))

/* ---- strings for display --------------------------------------------------- */

/* pv_display, pv_escape and pv_pretty are the interpreter's: the request is
 * [count][max][flags][len][n1][n2], the two colour strings, the text. */
static char *goperl_pv_show(goperl_frame_t *f, int fn, SV *dsv, const char *pv,
                            STRLEN count, STRLEN len, STRLEN max,
                            STRLEN *escaped, const char *c1, const char *c2,
                            U32 flags) __attribute__((unused));
static char *goperl_pv_show(goperl_frame_t *f, int fn, SV *dsv, const char *pv,
                            STRLEN count, STRLEN len, STRLEN max,
                            STRLEN *escaped, const char *c1, const char *c2,
                            U32 flags) {
    const size_t n1 = c1 ? __builtin_strlen(c1) : 0, n2 = c2 ? __builtin_strlen(c2) : 0;
    const goperl_gword_t w[6] = {(goperl_gword_t)count, (goperl_gword_t)max, flags,
                                 (goperl_gword_t)len, (goperl_gword_t)n1, (goperl_gword_t)n2};
    const size_t total = sizeof w + n1 + n2 + count;
    char *req = (char *)malloc(total);
    if (!req) goperl_croakf(f, "Out of memory!");
    memcpy(req, w, sizeof w);
    if (n1) memcpy(req + sizeof w, c1, n1);
    if (n2) memcpy(req + sizeof w + n1, c2, n2);
    if (count) memcpy(req + sizeof w + n1 + n2, pv, count);
    /* -> where the result is; second: how much of the text was taken */
    uint64_t taken = 0;
    const uint64_t r = goperl_do_opx(f, GOPERL_OP_CALL, GOPERL_TOK(dsv), (uint64_t)total,
                                     (uint64_t)(uint32_t)fn, req, (uint64_t)total, &taken);
    free(req);
    if (escaped) *escaped = (STRLEN)taken;
    return r ? (char *)goperl_guest_mem(f, r) : 0;
}
#define pv_display(dsv, pv, cur, len, pvlim)                                \
    goperl_pv_show(_gof, GOPERL_CALL_PV_DISPLAY, (dsv), (pv), (cur), (len), \
                   (pvlim), 0, 0, 0, 0)
#define pv_escape(dsv, str, count, max, escaped, flags)                    \
    goperl_pv_show(_gof, GOPERL_CALL_PV_ESCAPE, (dsv), (str), (count), 0,  \
                   (max), (escaped), 0, 0, (flags))
#define pv_pretty(dsv, str, count, max, start_color, end_color, flags)     \
    goperl_pv_show(_gof, GOPERL_CALL_PV_PRETTY, (dsv), (str), (count), 0,  \
                   (max), 0, (start_color), (end_color), (flags))
#define PERL_PV_ESCAPE_QUOTE 0x000001
#define PERL_PV_PRETTY_QUOTE PERL_PV_ESCAPE_QUOTE
#define PERL_PV_PRETTY_ELLIPSES 0x000002
#define PERL_PV_PRETTY_LTGT 0x000004
#define PERL_PV_PRETTY_EXACTSIZE 0x000008
#define PERL_PV_ESCAPE_UNI 0x000100
#define PERL_PV_ESCAPE_UNI_DETECT 0x000200
#define PERL_PV_ESCAPE_NONASCII 0x000400
#define PERL_PV_ESCAPE_FIRSTCHAR 0x000800
#define PERL_PV_ESCAPE_ALL 0x001000
#define PERL_PV_ESCAPE_NOBACKSLASH 0x002000
#define PERL_PV_ESCAPE_NOCLEAR 0x004000
#define PERL_PV_PRETTY_NOCLEAR PERL_PV_ESCAPE_NOCLEAR
#define PERL_PV_ESCAPE_RE 0x008000
#define PERL_PV_ESCAPE_DWIM 0x010000
#define PERL_PV_PRETTY_DUMP (PERL_PV_PRETTY_ELLIPSES | PERL_PV_PRETTY_QUOTE)
#define PERL_PV_PRETTY_REGPROP \
    (PERL_PV_PRETTY_ELLIPSES | PERL_PV_PRETTY_LTGT | PERL_PV_ESCAPE_RE | \
     PERL_PV_ESCAPE_NONASCII)

/* ---- more calls ------------------------------------------------------------- */

#define cv_clone(cv) ((CV *)GOPERL_SV(GOPERL_CALL(GOPERL_CALL_CV_CLONE, cv, 0)))
static CV *goperl_cv_clone_fn(pTHX_ CV *proto) __attribute__((unused));
static CV *goperl_cv_clone_fn(pTHX_ CV *proto) { return cv_clone(proto); }
#define Perl_cv_clone goperl_cv_clone_fn
#define getcwd_sv(sv) (GOPERL_CALL(GOPERL_CALL_GETCWD_SV, sv, 0) != 0)
#define mess_sv(sv, consume) \
    GOPERL_SV(GOPERL_CALL(GOPERL_CALL_MESS_SV, sv, (consume) ? 1 : 0))

/* ---- more interpreter variables --------------------------------------------- */

#define PL_errors (*goperl_plsv_ref(_gof, GOPERL_PL_ERRORS))
/* The interpreter's mark stack, where it is: what a context's
 * blk_oldmarksp indexes. (The marks of the running native function are its
 * own: PL_markstack_ptr, TOPMARK.) A mark there is an index of the
 * interpreter's stack, which PL_stack_base is by index (goperl_stack_below). */
static inline __attribute__((unused)) I32 *goperl_markstack_guest(goperl_frame_t *f) {
    goperl_stack_below(f);
    return (I32 *)goperl_guest_mem(f, GOPERL_PLVAR(GOPERL_PL_MARKSTACK));
}
#define PL_markstack (goperl_markstack_guest(_gof))
#define PL_savebegin (*(bool *)goperl_plint_ref(_gof, GOPERL_PL_SAVEBEGIN))
/* The scope stack is the interpreter's array, read where it is. */
#define PL_scopestack \
    ((I32 *)goperl_guest_mem(_gof, GOPERL_PLVAR(GOPERL_PL_SCOPESTACK)))
#undef PL_sub_generation
#define PL_sub_generation \
    (*(U32 *)goperl_plint_ref(_gof, GOPERL_PL_SUB_GENERATION))


/* ---- DBM filters (the macros of perl.h the DBM extensions are written
 * with: a filter is a sub called with the key or value in $_) ------------ */

#define DBM_setFilter(db_type, code)                 \
    STMT_START {                                     \
        if (db_type) RETVAL = sv_mortalcopy(db_type); \
        ST(0) = RETVAL;                              \
        if (db_type && (code == &PL_sv_undef)) {     \
            SvREFCNT_dec_NN(db_type);                \
            db_type = NULL;                          \
        } else if (code) {                           \
            if (db_type)                             \
                sv_setsv(db_type, code);             \
            else                                     \
                db_type = newSVsv(code);             \
        }                                            \
    }                                                \
    STMT_END

#define DBM_ckFilter(arg, type, name)                              \
    STMT_START {                                                   \
        if (db->type) {                                            \
            if (db->filtering) {                                   \
                croak("recursion detected in %s", name);           \
            }                                                      \
            ENTER;                                                 \
            SAVETMPS;                                              \
            SAVEINT(db->filtering);                                \
            db->filtering = TRUE;                                  \
            SAVE_DEFSV;                                            \
            if (name[7] == 's') arg = newSVsv(arg);                \
            DEFSV_set(arg);                                        \
            SvTEMP_off(arg);                                       \
            PUSHMARK(SP);                                          \
            PUTBACK;                                               \
            (void)perl_call_sv(db->type, G_DISCARD);               \
            SPAGAIN;                                               \
            PUTBACK;                                               \
            FREETMPS;                                              \
            LEAVE;                                                 \
            if (name[7] == 's') {                                  \
                arg = sv_2mortal(arg);                             \
            }                                                      \
        }                                                          \
    }                                                              \
    STMT_END


/* ---- the overload table of a stash ----------------------------------------- */

#define AMTf_AMAGIC 1
#define AMT_AMAGIC(amt) ((amt)->flags & AMTf_AMAGIC)
#define AMT_AMAGIC_on(amt) ((amt)->flags |= AMTf_AMAGIC)
#define AMT_AMAGIC_off(amt) ((amt)->flags &= ~AMTf_AMAGIC)
/* As the interpreter keeps it, with room for every overload method it has
 * (the table is given with their number). A copy: good until the
 * outermost native call returns. */
#define GOPERL_AMT_METHODS 128
typedef struct am_table {
    U8 flags;
    U8 fallback;
    U16 spare;
    U32 was_ok_sub;
    CV *table[GOPERL_AMT_METHODS];
} AMT;
static void *goperl_amt(goperl_frame_t *f, SV *stash) {
    /* CALL 10: AMT -> [flags][fallback][was_ok_sub][n], n subs */
    const goperl_gword_t *w = (const goperl_gword_t *)goperl_guest_mem(
        f, goperl_do_op3(f, GOPERL_OP_CALL, GOPERL_TOK(stash), 0, (uint64_t)(10), 0, 0));
    if (!w) return 0;
    AMT *amt = (AMT *)goperl_arena_alloc(sizeof(AMT));
    if (!amt) goperl_croakf(f, "Out of memory!");
    if (w[3] > GOPERL_AMT_METHODS)
        goperl_croakf(f, "the interpreter has more overload methods than the "
                         "go-perl XS SDK's AMT");
    amt->flags = (U8)w[0];
    amt->fallback = (U8)w[1];
    amt->was_ok_sub = (U32)w[2];
    for (goperl_gword_t i = 0; i < w[3]; i++) amt->table[i] = (CV *)GOPERL_SV(w[4 + i]);
    return amt;
}

/* ---- small additions --------------------------------------------------------- */

#define sv_catpvn_mg(sv, p, l)                                    \
    ({ GOPERL_ONCE(goperl_sv_catpvn_mg_sv_, sv);                  \
       sv_catpvn(goperl_sv_catpvn_mg_sv_, (p), (l));              \
       SvSETMAGIC(goperl_sv_catpvn_mg_sv_); })
static NV goperl_my_atof(goperl_frame_t *f, const char *s) __attribute__((unused));
static NV goperl_my_atof(goperl_frame_t *f, const char *s) {
    union {
        uint64_t u;
        NV d;
    } cvt;
    /* CALL 9: ATOF */
    cvt.u = goperl_do_op3(f, GOPERL_OP_CALL, 0, 0, (uint64_t)(9), s, __builtin_strlen(s));
    return cvt.d;
}
#define my_atof(s) goperl_my_atof(_gof, (s))
#define Atof my_atof
/* my_atof3 (numeric.c): the number in the len bytes at orig (its strlen
 * when 0) into *value, as the interpreter reads one; -> where it ended, or
 * NULL where perl answers NULL (*value then as perl leaves it). */
static char *goperl_my_atof3(goperl_frame_t *f, const char *orig, NV *value, STRLEN len)
    __attribute__((unused));
static char *goperl_my_atof3(goperl_frame_t *f, const char *orig, NV *value, STRLEN len) {
    union {
        uint64_t u;
        NV d;
    } cvt;
    uint64_t end = 0;
    if (len == 0) len = __builtin_strlen(orig);
    cvt.d = *value;
    /* CALL 9: ATOF; hi 1: my_atof3 from the NV in b; second: 1 + the bytes */
    cvt.u = goperl_do_opx(f, GOPERL_OP_CALL, 0, cvt.u, (uint64_t)9 | ((uint64_t)1 << 32), orig,
                          len, &end);
    *value = cvt.d;
    return end ? (char *)orig + (end - 1) : (char *)0;
}
#define my_atof3(orig, value, len) \
    goperl_my_atof3(_gof, (const char *)(orig), (NV *)(value), (STRLEN)(len))
#define my_atof2(orig, value) my_atof3((orig), (value), 0)
/* call_list: the subs of the array (BEGIN, UNITCHECK, ...) run and taken
 * off it, as perl.c's. */
#define GOPERL_CALL_CALL_LIST 100
#define call_list(oldscope, av)                                             \
    ((void)(goperl_optree_flush(_gof),                                      \
            goperl_do_op3(_gof, GOPERL_OP_CALL, GOPERL_TOK(av),             \
                          (uint64_t)(int64_t)(I32)(oldscope), (uint64_t)GOPERL_CALL_CALL_LIST, 0, 0)))
/* hv_scalar: what a hash is in scalar context (its count of keys, or what
 * a tied hash's SCALAR says), a mortal. */
#define GOPERL_CALL_HV_SCALAR 101
#define hv_scalar(hv)                                                       \
    ((SV *)(uintptr_t)goperl_do_op3(_gof, GOPERL_OP_CALL, GOPERL_TOK(hv), 0, \
                                    (uint64_t)GOPERL_CALL_HV_SCALAR, 0, 0))
/* packlist and unpackstring (pp_pack.c): pack's and unpack's work on the
 * values from beglist up to endlist, and on the bytes from s up to strend,
 * whose values are pushed on the stack. */
#define GOPERL_CALL_PACKLIST 102
#define GOPERL_CALL_UNPACKSTRING 103
static void goperl_packlist(goperl_frame_t *f, SV *cat, const char *pat, const char *patend,
                            SV **beglist, SV **endlist) __attribute__((unused));
static void goperl_packlist(goperl_frame_t *f, SV *cat, const char *pat, const char *patend,
                            SV **beglist, SV **endlist) {
    const size_t plen = (size_t)(patend - pat), n = endlist > beglist ? (size_t)(endlist - beglist) : 0;
    const size_t size = 2 * sizeof(goperl_gword_t) + n * sizeof(goperl_gword_t) + plen;
    char *const buf = (char *)malloc(size ? size : 1);
    if (!buf) goperl_croakf(f, "Out of memory!");
    goperl_gword_t w[2] = {(goperl_gword_t)plen, (goperl_gword_t)n};
    memcpy(buf, w, sizeof w);
    for (size_t i = 0; i < n; i++) {
        const goperl_gword_t tok = GOPERL_TOK(beglist[i]);
        memcpy(buf + sizeof w + i * sizeof tok, &tok, sizeof tok);
    }
    memcpy(buf + sizeof w + n * sizeof(goperl_gword_t), pat, plen);
    goperl_do_op3(f, GOPERL_OP_CALL, GOPERL_TOK(cat), 0, (uint64_t)GOPERL_CALL_PACKLIST, buf, size);
    free(buf);
}
#define packlist(cat, pat, patend, beglist, endlist)                        \
    goperl_packlist(_gof, (SV *)(cat), (const char *)(pat), (const char *)(patend), \
                    (SV **)(beglist), (SV **)(endlist))
static SSize_t goperl_unpackstring(goperl_frame_t *f, const char *pat, const char *patend,
                                   const char *s, const char *strend, U32 flags)
    __attribute__((unused));
static SSize_t goperl_unpackstring(goperl_frame_t *f, const char *pat, const char *patend,
                                   const char *s, const char *strend, U32 flags) {
    const goperl_gword_t plen = (goperl_gword_t)(patend - pat);
    const size_t slen = strend > s ? (size_t)(strend - s) : 0;
    const size_t size = sizeof plen + (size_t)plen + slen;
    char *const buf = (char *)malloc(size);
    if (!buf) goperl_croakf(f, "Out of memory!");
    memcpy(buf, &plen, sizeof plen);
    memcpy(buf + sizeof plen, pat, (size_t)plen);
    memcpy(buf + sizeof plen + plen, s, slen);
    goperl_stk(f);
    uint64_t second = 0;
    const uint64_t words =
        goperl_do_opx(f, GOPERL_OP_CALL, 0, 0,
                      (uint64_t)GOPERL_CALL_UNPACKSTRING | ((uint64_t)flags << 32), buf, size,
                      &second);
    free(buf);
    const I32 count = (I32)(second & 0x7FFFFFFF);
    goperl_call_results_take(f, words, count);
    return (SSize_t)count;
}
#define unpackstring(pat, patend, s, strend, flags)                         \
    goperl_unpackstring(_gof, (const char *)(pat), (const char *)(patend), \
                        (const char *)(s), (const char *)(strend), (U32)(flags))
/* the embedding calls: native code runs inside the program's interpreter
 * and cannot make or end one */
#define GOPERL_NO_EMBEDDING(name)                                           \
    (goperl_croakf(goperl_cur_frame_v, "%s: go-perl does not let native code make or end an " \
                   "interpreter", name), 0)
#define perl_alloc() ((PerlInterpreter *)(uintptr_t)GOPERL_NO_EMBEDDING("perl_alloc"))
#define perl_construct(i) ((void)(i), (void)GOPERL_NO_EMBEDDING("perl_construct"))
#define perl_parse(i, x, c, v, e) \
    ((void)(i), (void)(x), (void)(c), (void)(v), (void)(e), GOPERL_NO_EMBEDDING("perl_parse"))
#define perl_run(i) ((void)(i), GOPERL_NO_EMBEDDING("perl_run"))
#define perl_destruct(i) ((void)(i), GOPERL_NO_EMBEDDING("perl_destruct"))
#define perl_free(i) ((void)(i), (void)GOPERL_NO_EMBEDDING("perl_free"))
#define perl_clone(i, flags) \
    ((void)(i), (void)(flags), (PerlInterpreter *)(uintptr_t)GOPERL_NO_EMBEDDING("perl_clone"))
#define HV_ITERNEXT_WANTPLACEHOLDERS 0x01
#define hv_iternext_flags(hv, flags)                                         \
    goperl_he_at(_gof, goperl_op0(GOPERL_OP_HV_ITERNEXT, GOPERL_TOK(hv),     \
                                  (uint64_t)(int64_t)(flags)))
#define CvSPECIAL_on(cv)                          \
    ({ GOPERL_ONCE(goperl_CvSPECIAL_on_cv_, cv);  \
       GOPERL_CVFLAGS_CHANGE(goperl_CvSPECIAL_on_cv_, CVf_UNIQUE, 0); \
       GOPERL_FLAGS_CHANGE(goperl_CvSPECIAL_on_cv_, SVf_FAKE, 0); (void)0; })
#define CvSPECIAL(cv) \
    ((CvFLAGS(cv) & CVf_UNIQUE) && (SvFLAGS((SV *)(cv)) & SVf_FAKE))
#define HeKFLAGS(he) HEK_FLAGS(HeKEY_hek(he))
#define HvSHAREKEYS_on(hv) GOPERL_FLAGS_CHANGE(hv, SVphv_SHAREKEYS, 0)
#define HvSHAREKEYS_off(hv) GOPERL_FLAGS_CHANGE(hv, 0, SVphv_SHAREKEYS)
#define HvSHAREKEYS(hv) (SvFLAGS((SV *)(hv)) & SVphv_SHAREKEYS)
#define UV_MAX_P1 (4.0 * (1 + ((UV_MAX) >> 2)))
#define UV_MAX_P1_HALF (2.0 * (1 + ((UV_MAX) >> 2)))
#define IV_MAX_P1 (2.0 * (1 + ((IV_MAX) >> 1)))
#define start_subparse(is_format, flags) \
    goperl_start_subparse_(_gof, (is_format), (flags))
#define newATTRSUB(floor, o, proto, attrs, block)                         \
    goperl_newattrsub(_gof, (floor), (OP *)(o), (OP *)(proto), (OP *)(attrs), \
                      (OP *)(block))
#define newSUB(floor, o, proto, block) newATTRSUB((floor), (o), (proto), 0, (block))
#define cxinc() ((I32)(int64_t)goperl_tree(_gof, GOPERL_TREE_CXINC, 0, 0))
static I32 goperl_cxinc_fn(pTHX) __attribute__((unused));
static I32 goperl_cxinc_fn(pTHX) { return cxinc(); }
#define Perl_cxinc goperl_cxinc_fn
#define find_runcv(db_seqp)                                              \
    ({ U32 *const goperl_find_runcv_p_ = (db_seqp);                      \
       if (goperl_find_runcv_p_) *goperl_find_runcv_p_ = 0;              \
       (CV *)GOPERL_SV(goperl_tree(_gof, GOPERL_TREE_FIND_RUNCV, 0, 0)); })
static void goperl_lex_stuff_sv(goperl_frame_t *f, SV *sv, U32 flags)
    __attribute__((unused));
static void goperl_lex_stuff_sv(goperl_frame_t *f, SV *sv, U32 flags) {
    const goperl_gword_t w[2] = {GOPERL_TOK(sv), flags};
    goperl_parser_push(f);
    goperl_tree(f, GOPERL_TREE_LEX_STUFF_SV, 2, w);
    goperl_parser_pull(f);
}
#define lex_stuff_sv(sv, flags) goperl_lex_stuff_sv(_gof, (SV *)(sv), (U32)(flags))
static PADOFFSET goperl_pad_add_name_sv(goperl_frame_t *f, SV *name, U32 flags,
                                        HV *typestash, HV *ourstash)
    __attribute__((unused));
static PADOFFSET goperl_pad_add_name_sv(goperl_frame_t *f, SV *name, U32 flags,
                                        HV *typestash, HV *ourstash) {
    const goperl_gword_t w[4] = {GOPERL_TOK(name), flags,
                           GOPERL_TOK(typestash),
                           GOPERL_TOK(ourstash)};
    return (PADOFFSET)goperl_tree(f, GOPERL_TREE_PAD_ADD_NAME_SV, 4, w);
}
#define pad_add_name_sv(name, flags, typestash, ourstash)               \
    goperl_pad_add_name_sv(_gof, (SV *)(name), (U32)(flags),            \
                           (HV *)(typestash), (HV *)(ourstash))


/* ---- a parser of its own, COP saves ------------------------------------------ */

/* lex_start makes a parser for a string (or for nothing); the scope that
 * was open when it was called ends it. */
static void goperl_lex_start(goperl_frame_t *f, SV *line, void *rsfp, U32 flags)
    __attribute__((unused));
static void goperl_lex_start(goperl_frame_t *f, SV *line, void *rsfp, U32 flags) {
    if (rsfp)
        goperl_croakf(f, "lex_start with a file handle is not supported by "
                         "the go-perl XS SDK");
    goperl_parser_push(f);
    /* CALL 14: LEX_START */
    goperl_do_op3(f, GOPERL_OP_CALL, GOPERL_TOK(line), flags, (uint64_t)(14), 0, 0);
    goperl_parser_pull(f);
}
#define lex_start(line, rsfp, flags) \
    goperl_lex_start(_gof, (SV *)(line), (void *)(rsfp), (U32)(flags))
/* The line and the file of a COP are saved by the interpreter, which puts
 * them back when the scope ends (CALL 13: 0 the line, 1 the file). */
#define SAVECOPLINE(c)                                              \
    ((void)goperl_op3(GOPERL_OP_CALL, (c)->gop, 0, 13))
#define SAVECOPFILE_FREE(c)                                         \
    ((void)goperl_op3(GOPERL_OP_CALL, (c)->gop, 1, 13))
#define SAVECOPFILE(c) SAVECOPFILE_FREE(c)
/* The stash of a COP, saved the same way (CALL 13 with 2): the scope's end
 * puts it back in the interpreter's COP and in its copy here. */
#define SAVECOPSTASH_FREE(c)                                        \
    ({ COP *const goperl_scsf_c_ = (c);                              \
       (void)goperl_op3(GOPERL_OP_CALL, goperl_scsf_c_->gop, 2, 13); \
       goperl_hostsave_ptr(_gof, (void **)&goperl_scsf_c_->cop_stashpv); })
#define SAVECOPSTASH(c) SAVECOPSTASH_FREE(c)

/* ---- a few more fields --------------------------------------------------------- */

#define GvREFCNT(gv) ((U32)GOPERL_FIELD(15, gv, 0)) /* FIELD_GV_REFCNT */
/* CvOUTSIDE is an lvalue (`CvOUTSIDE(cv) = NULL`: the plain store). */
#undef CvOUTSIDE
#define CvOUTSIDE(cv) \
    (*(CV **)goperl_lv_ref(_gof, (SV *)(cv), GOPERL_LV_CVOUTSIDE, GOPERL_LV_AT))
#define CvWEAKOUTSIDE(cv) (CvFLAGS(cv) & CVf_WEAKOUTSIDE)
#define CvWEAKOUTSIDE_on(cv) GOPERL_CVFLAGS_CHANGE(cv, CVf_WEAKOUTSIDE, 0)
#define CvWEAKOUTSIDE_off(cv) GOPERL_CVFLAGS_CHANGE(cv, 0, CVf_WEAKOUTSIDE)
#define isWORDCHAR_utf8_safe(p, e)              \
    goperl_charclass_uv(_gof, GOPERL_CC_WORDCHAR, \
                        goperl_utf8_cp(_gof, (const U8 *)(p), (const U8 *)(e)))


/* ---- ops with auxiliary data ----------------------------------------------------
 * What the aux pointer of an op means is the op type's: an argument check
 * has a struct of three fields, an argument element a number in place of a
 * pointer, a custom op whatever its module made. The interpreter is given
 * what it runs itself (the first two), in its own memory; the data of a
 * custom op stays with the module. */
#define GOPERL_TREE_NEWUNOP_AUX 23
static GOPERL_OP_BUILDER OP *goperl_newunop_aux(goperl_frame_t *f, I32 type, I32 flags, OP *first,
                              UNOP_AUX_item *aux) __attribute__((unused));
static GOPERL_OP_BUILDER OP *goperl_newunop_aux(goperl_frame_t *f, I32 type, I32 flags, OP *first,
                              UNOP_AUX_item *aux) {
    goperl_gword_t w[7] = {(goperl_gword_t)(int64_t)type, (goperl_gword_t)(int64_t)flags,
                           goperl_optok(f, first), 0, 0, 0, 0};
    switch (type) {
    case OP_ARGELEM:
        w[3] = 1;
        w[4] = (goperl_gword_t)(uintptr_t)aux; /* a number in place of the pointer */
        break;
    case OP_ARGCHECK:
        if (aux) {
            const struct op_argcheck_aux *ca = (const struct op_argcheck_aux *)aux;
            w[3] = 2;
            w[4] = (goperl_gword_t)ca->params;
            w[5] = (goperl_gword_t)ca->opt_params;
            w[6] = (goperl_gword_t)(unsigned char)ca->slurpy;
        }
        break;
    case OP_CUSTOM:
        break;
    default:
        goperl_croakf(f, "newUNOP_AUX of %s is not supported by the go-perl XS SDK",
                      PL_op_name[type]);
    }
    OP *o = goperl_tree_op(f, goperl_tree(f, GOPERL_TREE_NEWUNOP_AUX, 7, w));
    if (type == OP_ARGCHECK && aux) Safefree(aux); /* the op has its own */
    if (type == OP_CUSTOM) o->op_aux = aux;
    if (type == OP_ARGELEM) o->op_aux = aux;
    return o;
}
#define newUNOP_AUX(type, flags, first, aux) \
    goperl_newunop_aux(_gof, (type), (flags), (OP *)(first), (UNOP_AUX_item *)(aux))
/* An SV in an item is an address of the interpreter's, and a module's own
 * items hold the same handles. */
#define UNOP_AUX_item_sv(item) GOPERL_SV((item)->uv)

/* ---- pads: goperl_compiler.h ---------------------------------------------------- */
#define GOPERL_PL_CURPAD 69
#define GOPERL_PL_COMPPAD 70
#define GOPERL_PL_COMPPAD_NAME 71
#define KEY_sigvar 0xFFFF
#define SvPADSTALE_on(sv) GOPERL_FLAGS_CHANGE(sv, SVs_PADSTALE, 0)
#define SvPADSTALE_off(sv) GOPERL_FLAGS_CHANGE(sv, 0, SVs_PADSTALE)
#define SvPADTMP_on(sv) GOPERL_FLAGS_CHANGE(sv, SVs_PADTMP, 0)
#define SvPADTMP_off(sv) GOPERL_FLAGS_CHANGE(sv, 0, SVs_PADTMP)


#define CvCLONE_on(cv) GOPERL_CVFLAGS_CHANGE(cv, CVf_CLONE, 0)
#define CvCLONE_off(cv) GOPERL_CVFLAGS_CHANGE(cv, 0, CVf_CLONE)
#define CvCLONE(cv) (CvFLAGS(cv) & CVf_CLONE)
/* FIELD 17: the aux flags of a stash (0 for a hash without aux) */
#define HvAUXf_IS_CLASS 0x4
#define HvSTASH_IS_CLASS(hv) ((GOPERL_FIELD(17, hv, 0) & HvAUXf_IS_CLASS) != 0)
#define gv_init_sv(gv, stash, namesv, flags)                              \
    ({ STRLEN goperl_gv_init_sv_len_;                                     \
       SV *const goperl_gv_init_sv_name_ = (SV *)(namesv);                \
       const char *const goperl_gv_init_sv_p_ =                           \
           SvPV(goperl_gv_init_sv_name_, goperl_gv_init_sv_len_);         \
       gv_init_pvn((gv), (stash), goperl_gv_init_sv_p_,                   \
                   goperl_gv_init_sv_len_,                                \
                   (flags) | (SvUTF8(goperl_gv_init_sv_name_) ? SVf_UTF8 : 0)); })


/* ---- MULTICALL -------------------------------------------------------------------
 * Calls one sub many times (the block of a first { ... } @list): the sub
 * is given no arguments, reads $_ or $a and $b as the caller set them, and
 * leaves what it returns on the stack, where the caller reads it before
 * the next call. perl saves the cost of entering the sub each time; here
 * each MULTICALL is a call of the sub.
 *
 * As in perl, the calls run on a stack of their own, from PUSH_MULTICALL
 * to POP_MULTICALL: what a call returns is PL_stack_base[1] and up to
 * PL_stack_sp (a block in list context: List::MoreUtils' pairwise), and
 * the stack the caller had, with its arguments, is as it was afterwards
 * (a pointer to the arguments taken before PUSH_MULTICALL stays good: a
 * sort that calls the block rearranges them in place).
 *
 * As in perl, MULTICALL needs nothing but multicall_cop: a module hands
 * that one variable to a function of its own and says MULTICALL there. So
 * what the call needs — the sub, the context the variable gimme had at
 * PUSH_MULTICALL, the caller's stack — is kept in a record that
 * multicall_cop points into.
 *
 * A function that leaves by croak between the two leaves the frame on the
 * stack of the calls, which is given up with the frame. */
typedef struct {
    OP op; /* what multicall_cop points to */
    CV *cv;
    U8 gimme;
    SV **outer_base; /* the caller's stack */
    SV **outer_sp;
    int64_t outer_cap;
} goperl_multicall_t;
#define dMULTICALL                                                   \
    goperl_multicall_t goperl_multicall_v __attribute__((unused));   \
    OP *multicall_cop __attribute__((unused)) =                      \
        (memset(&goperl_multicall_v, 0, sizeof(goperl_multicall_v)), \
         &goperl_multicall_v.op);                                    \
    bool multicall_oldcatch __attribute__((unused)) = 0
static void goperl_multicall_push(goperl_frame_t *f, goperl_multicall_t *m, CV *cv, U8 gimme)
    __attribute__((unused));
static void goperl_multicall_push(goperl_frame_t *f, goperl_multicall_t *m, CV *cv, U8 gimme) {
    enum { first_cap = 64 };
    goperl_stk(f);
    /* the stack of a pp function is the interpreter's own, kept in step
     * with it: another one cannot stand in for it */
    if (f->pp_frame)
        goperl_croakf(f, "MULTICALL in a pp function is not supported by the go-perl XS SDK");
    SV **calls = (SV **)malloc(first_cap * sizeof(SV *));
    if (!calls) goperl_croakf(f, "Out of memory for the stack of a MULTICALL");
    /* the entry below the first value is undef, as on every stack of
     * perl's: *PL_stack_sp is undef after a call that returned nothing */
    calls[0] = &PL_sv_undef;
    m->cv = cv;
    m->gimme = gimme;
    m->outer_base = f->stbase;
    m->outer_sp = f->psp;
    m->outer_cap = f->stcap;
    f->stbase = calls;
    f->stcap = first_cap;
    f->psp = calls;
    /* the scope of the calls: what is saved until POP_MULTICALL is undone
     * there (perl's CX_LEAVE_SCOPE of the multicall's context) */
    ENTER;
}
#define PUSH_MULTICALL_FLAGS(the_cv, flags)                                    \
    goperl_multicall_push(_gof, (goperl_multicall_t *)multicall_cop,           \
                          (CV *)(the_cv), (U8)(gimme))
#define PUSH_MULTICALL(the_cv) PUSH_MULTICALL_FLAGS((the_cv), 0)
static void goperl_multicall(goperl_frame_t *f, goperl_multicall_t *m)
    __attribute__((unused));
static void goperl_multicall(goperl_frame_t *f, goperl_multicall_t *m) {
    (void)f;
    PL_stack_sp = PL_stack_base;
    {
        dSP;
        PUSHMARK(SP);
        PUTBACK;
        /* perl's MULTICALL runs the ops of the sub and leaves the stack as
         * the last of them left it: the value of the last statement is
         * there in void context too (List::MoreUtils' qsort calls its
         * block in the context qsort itself was called in, and reads the
         * result). A call in void context would drop it, so the block is
         * called for one value then. */
        (void)call_sv((SV *)m->cv, (m->gimme == G_VOID ? G_SCALAR : m->gimme) | G_NOARGS);
    }
}
#define MULTICALL goperl_multicall(_gof, (goperl_multicall_t *)multicall_cop)
static void goperl_multicall_pop(goperl_frame_t *f, goperl_multicall_t *m)
    __attribute__((unused));
static void goperl_multicall_pop(goperl_frame_t *f, goperl_multicall_t *m) {
    if (!m->outer_base) return;
    LEAVE;
    free(f->stbase);
    f->stbase = m->outer_base;
    f->stcap = m->outer_cap;
    f->psp = m->outer_sp;
    m->outer_base = 0;
}
#define POP_MULTICALL goperl_multicall_pop(_gof, (goperl_multicall_t *)multicall_cop)
#define CHANGE_MULTICALL_FLAGS(the_cv, flags) \
    STMT_START { ((goperl_multicall_t *)multicall_cop)->cv = (CV *)(the_cv); } STMT_END

/* ---- pp.h: taking from the stack ---------------------------------------------- */
#define dTOPss SV *sv = TOPs
#define dPOPss SV *sv = POPs
#define dTOPiv IV value = TOPi
#define dPOPiv IV value = POPi
#define dTOPnv NV value = TOPn
#define dPOPnv NV value = POPn
#define dTOPuv UV value = TOPu
#define dPOPuv UV value = POPu
#define dPOPTOPssrl SV *right = POPs; SV *left = TOPs
#define dPOPTOPnnrl NV right = POPn; NV left = TOPn
#define dPOPTOPiirl IV right = POPi; IV left = TOPi
#ifndef TOPi
#define TOPi ((IV)SvIV(TOPs))
#define TOPu ((UV)SvUV(TOPs))
#define TOPn ((NV)SvNV(TOPs))
#endif
#ifndef POPi
#define POPi ((IV)SvIV(POPs))
#define POPu ((UV)SvUV(POPs))
#define POPn ((NV)SvNV(POPs))
#endif


/* ---- subs built by a module -------------------------------------------------- */
#define GOPERL_TREE_PARSE_SUBSIGNATURE 24
#define GOPERL_TREE_NEWMYSUB 25
#define GOPERL_TREE_CV_NAME 26
static inline __attribute__((unused)) OP *goperl_parse_subsignature(
    goperl_frame_t *f, U32 flags) {
    const goperl_gword_t w[1] = {flags};
    goperl_parser_push(f);
    const uint64_t tok = goperl_tree(f, GOPERL_TREE_PARSE_SUBSIGNATURE, 1, w);
    goperl_parser_pull(f);
    return goperl_tree_op(f, tok);
}
#define parse_subsignature(flags) goperl_parse_subsignature(_gof, (U32)(flags))
static inline __attribute__((unused)) CV *goperl_newmysub(
    goperl_frame_t *f, I32 floor, OP *o, OP *proto, OP *attrs, OP *block) {
    const goperl_gword_t w[5] = {(goperl_gword_t)(int64_t)floor, goperl_optok(f, o),
                                 goperl_optok(f, proto), goperl_optok(f, attrs),
                                 goperl_optok(f, block)};
    return (CV *)GOPERL_SV(goperl_tree(f, GOPERL_TREE_NEWMYSUB, 5, w));
}
#define newMYSUB(floor, o, proto, attrs, block)                         \
    goperl_newmysub(_gof, (I32)(floor), (OP *)(o), (OP *)(proto),       \
                    (OP *)(attrs), (OP *)(block))

/* A key as perl shares it: here the key itself, laid out like perl's HEK
 * (NUL after the bytes, then the flags byte), for as long as the outermost
 * native call runs; the interpreter shares the key when it is given to it
 * (CvNAME_HEK_set). A negative length says the key is UTF-8. A key whose
 * hash the caller does not give (0: the names of stashes, globs and subs)
 * gets the interpreter's (PERL_HASH), which the interpreter's own keys
 * have. */
static HEK *goperl_hek_of(goperl_frame_t *f, const char *key, I32 len,
                          int utf8, U32 hash) __attribute__((unused));
static HEK *goperl_hek_of(goperl_frame_t *f, const char *key, I32 len,
                          int utf8, U32 hash) {
    HEK *hek = (HEK *)goperl_arena_alloc(sizeof(HEK) + (size_t)len + 2);
    if (!hek) goperl_croakf(f, "share_hek: out of memory");
    if (!hash) hash = goperl_key_hash(f, key, (uint64_t)len);
    hek->hek_hash = hash;
    hek->hek_len = len;
    memcpy(hek->hek_key, key, (size_t)len);
    hek->hek_key[len] = 0;
    hek->hek_key[len + 1] = (char)(utf8 ? 1 : 0); /* HVhek_UTF8 */
    return hek;
}
#define share_hek(key, len, hash)                                          \
    ({ const I32 goperl_share_hek_len_ = (I32)(len);                       \
       goperl_hek_of(_gof, (key),                                          \
                     goperl_share_hek_len_ < 0 ? -goperl_share_hek_len_    \
                                               : goperl_share_hek_len_,    \
                     goperl_share_hek_len_ < 0, (U32)(hash)); })
/* The name a sub has by itself (a sub that is not in a glob), or NULL. */
static HEK *goperl_cv_name_hek(goperl_frame_t *f, CV *cv)
    __attribute__((unused));
static HEK *goperl_cv_name_hek(goperl_frame_t *f, CV *cv) {
    /* -> the name, or 0; second: length<<1 | utf8 */
    const goperl_gword_t w[4] = {GOPERL_TOK(cv), 0, 0, 0};
    goperl_optree_flush(f);
    uint64_t second = 0;
    const uint64_t r = goperl_do_opx(f, GOPERL_OP_TREE, GOPERL_TREE_CV_NAME, sizeof w, 0,
                                     (const char *)w, sizeof w, &second);
    if (!r) return 0;
    const I32 len = (I32)(second >> 1);
    return goperl_hek_of(f, (const char *)goperl_guest_mem(f, r), len, (int)(second & 1), 0);
}
#define CvNAME_HEK(cv) goperl_cv_name_hek(_gof, (CV *)(cv))
static void goperl_cv_name_hek_set(goperl_frame_t *f, CV *cv, HEK *hek)
    __attribute__((unused));
static void goperl_cv_name_hek_set(goperl_frame_t *f, CV *cv, HEK *hek) {
    const size_t len = (size_t)hek->hek_len;
    /* the four words, then the name */
    char *req = (char *)malloc(32 + len + 1);
    if (!req) goperl_croakf(f, "Out of memory!");
    const goperl_gword_t w[4] = {GOPERL_TOK(cv), 1,
                                 (goperl_gword_t)(hek->hek_key[len + 1] & 1),
                                 (goperl_gword_t)len};
    memcpy(req, w, 32);
    memcpy(req + 32, hek->hek_key, len);
    req[32 + len] = 0;
    goperl_optree_flush(f);
    goperl_do_op(f, GOPERL_OP_TREE, GOPERL_TREE_CV_NAME, 32 + len + 1, req,
                 32 + len + 1);
    free(req);
}
#define CvNAME_HEK_set(cv, hek) \
    goperl_cv_name_hek_set(_gof, (CV *)(cv), (HEK *)(hek))


/* The op_multiparam_aux of an op of the interpreter's (at g in its memory)
 * as a block of this process's. It is read field by field where it is: the
 * interpreter is LP64 like the host, so its structure has the layout the
 * SDK's has here, which the assertions below pin (perl.cc asserts the same
 * offsets of the interpreter's); the pointers in it are the interpreter's
 * addresses. */
#define GOPERL_MPA_OFF(field) ((uint32_t)offsetof(struct op_multiparam_aux, field))
#define GOPERL_MPN_OFF(field) ((uint32_t)offsetof(struct op_multiparam_named_aux, field))
typedef char goperl_multiparam_aux_layout
    [(GOPERL_MPA_OFF(min_args) == 0 && GOPERL_MPA_OFF(n_positional) == 8 &&
      GOPERL_MPA_OFF(n_named) == 16 && GOPERL_MPA_OFF(slurpy) == 24 &&
      GOPERL_MPA_OFF(slurpy_padix) == 32 && GOPERL_MPA_OFF(param_padix) == 40 &&
      GOPERL_MPA_OFF(named) == 48 && GOPERL_MPN_OFF(padix) == 0 &&
      GOPERL_MPN_OFF(namepv) == 8 && GOPERL_MPN_OFF(namelen) == 16 &&
      GOPERL_MPN_OFF(namehash) == 24 && sizeof(struct op_multiparam_named_aux) == 32)
         ? 1 : -1] __attribute__((unused));
static void *goperl_multiparam_aux(goperl_frame_t *f, uint64_t g64) {
    if (!g64) return 0;
    const uint64_t g = g64;
    const size_t n_positional = (size_t)goperl_g64(f, g + GOPERL_MPA_OFF(n_positional));
    const size_t n_named = (size_t)goperl_g64(f, g + GOPERL_MPA_OFF(n_named));
    struct op_multiparam_aux *aux = (struct op_multiparam_aux *)malloc(
        sizeof(*aux) + n_positional * sizeof(PADOFFSET) +
        n_named * sizeof(struct op_multiparam_named_aux));
    if (!aux) goperl_croakf(f, "out of memory for the parameters of a signature");
    aux->min_args = (size_t)goperl_g64(f, g + GOPERL_MPA_OFF(min_args));
    aux->n_positional = n_positional;
    aux->n_named = n_named;
    aux->slurpy = *(const char *)(f->membase + g + GOPERL_MPA_OFF(slurpy));
    aux->slurpy_padix = (PADOFFSET)goperl_g64(f, g + GOPERL_MPA_OFF(slurpy_padix));
    aux->param_padix = (PADOFFSET *)(aux + 1);
    aux->named = (struct op_multiparam_named_aux *)(aux->param_padix + n_positional);
    const uint64_t padix = goperl_g64(f, g + GOPERL_MPA_OFF(param_padix));
    for (size_t i = 0; padix && i < n_positional; i++)
        aux->param_padix[i] =
            (PADOFFSET)goperl_g64(f, padix + GOPERL_GUEST_PTRSIZE * (uint64_t)i);
    const uint64_t named = goperl_g64(f, g + GOPERL_MPA_OFF(named));
    for (size_t i = 0; named && i < n_named; i++) {
        const uint64_t n = named + (uint64_t)(i * sizeof(struct op_multiparam_named_aux));
        aux->named[i].padix = (PADOFFSET)goperl_g64(f, n + GOPERL_MPN_OFF(padix));
        aux->named[i].namepv =
            (const char *)goperl_guest_mem(f, goperl_g64(f, n + GOPERL_MPN_OFF(namepv)));
        aux->named[i].namelen = (STRLEN)goperl_g64(f, n + GOPERL_MPN_OFF(namelen));
        aux->named[i].namehash = goperl_g32(f, n + GOPERL_MPN_OFF(namehash));
        /* the bit field is the lowest bit of the byte after namehash */
        aux->named[i].is_required =
            (*(const U8 *)(f->membase + n + GOPERL_MPN_OFF(namehash) + 4) & 1) != 0;
    }
    if (!named) aux->named = 0;
    return aux;
}

/* A pad name in a message: %" PNf " with PNfARG(pn). */
#define PNf UTF8f
#define PNfARG(pn) (int)1, (UV)PadnameLEN(pn), (void *)PadnamePV(pn)

static I32 goperl_sv_cmp_locale_fn(pTHX_ SV *a, SV *b) {
    return sv_cmp_locale(a, b);
}

/* ---- more of perl's public API (op.h, cv.h, gv.h, hv.h, pp.h, XSUB.h,
 * util.c, utf8.c) ---------------------------------------------------------- */

/* The description of an op; a custom op's is its XOP's (op.h). */
#define OP_DESC(o)                                                          \
    ({ GOPERL_ONCE(goperl_OP_DESC_o_, (const OP *)(o));                     \
       goperl_OP_DESC_o_->op_type == OP_CUSTOM                              \
           ? XopENTRY(goperl_custom_op_xop(_gof, goperl_OP_DESC_o_), xop_desc) \
           : PL_op_desc[goperl_OP_DESC_o_->op_type]; })
/* The number of ops the interpreter has (intrpvar.h's PL_maxo, which the
 * interpreter never changes). */
#define PL_maxo MAXO
/* The definition of a pp function (pp.h). */
#ifndef PP
#define PP(s) OP *Perl_##s(pTHX)
#endif
/* Where a missing argument of a format was missed: the op running. */
static const char *goperl_op_desc_now(goperl_frame_t *f) {
    (void)f;
    return PL_op ? OP_DESC(PL_op) : "sv_vcatpvfn()";
}

/* The class of an op, as B sees it (dump.c's op_class). */
typedef enum {
    OPclass_NULL,
    OPclass_BASEOP,
    OPclass_UNOP,
    OPclass_BINOP,
    OPclass_LOGOP,
    OPclass_LISTOP,
    OPclass_PMOP,
    OPclass_SVOP,
    OPclass_PADOP,
    OPclass_PVOP,
    OPclass_LOOP,
    OPclass_COP,
    OPclass_METHOP,
    OPclass_UNOP_AUX
} OPclass;
static OPclass goperl_op_class_of(const OP *o) __attribute__((unused));
static OPclass goperl_op_class_of(const OP *o) {
    bool custom = 0;
    if (!o) return OPclass_NULL;
    if (o->op_type == 0) {
        if (o->op_targ == OP_NEXTSTATE || o->op_targ == OP_DBSTATE) return OPclass_COP;
        return (o->op_flags & OPf_KIDS) ? OPclass_UNOP : OPclass_BASEOP;
    }
    if (o->op_type == OP_SASSIGN)
        return (o->op_private & OPpASSIGN_BACKWARDS) ? OPclass_UNOP : OPclass_BINOP;
    if (o->op_type == OP_AELEMFAST) return OPclass_SVOP;
    if (o->op_type == OP_CUSTOM) custom = 1;
    switch (OP_CLASS(o)) {
    case OA_BASEOP: return OPclass_BASEOP;
    case OA_UNOP: return OPclass_UNOP;
    case OA_BINOP: return OPclass_BINOP;
    case OA_LOGOP: return OPclass_LOGOP;
    case OA_LISTOP: return OPclass_LISTOP;
    case OA_PMOP: return OPclass_PMOP;
    case OA_SVOP: return OPclass_SVOP;
    case OA_PADOP: return OPclass_PADOP;
    case OA_PVOP_OR_SVOP:
        return (!custom && (o->op_private & OPpTRANS_USE_SVOP)) ? OPclass_SVOP : OPclass_PVOP;
    case OA_LOOP: return OPclass_LOOP;
    case OA_COP: return OPclass_COP;
    case OA_BASEOP_OR_UNOP:
        return (o->op_flags & OPf_KIDS) ? OPclass_UNOP : OPclass_BASEOP;
    case OA_FILESTATOP:
        return (o->op_flags & OPf_KIDS) ? OPclass_UNOP
               : (o->op_flags & OPf_REF) ? OPclass_SVOP : OPclass_BASEOP;
    case OA_LOOPEXOP:
        if (o->op_flags & OPf_STACKED) return OPclass_UNOP;
        if (o->op_flags & OPf_SPECIAL) return OPclass_BASEOP;
        return OPclass_PVOP;
    case OA_METHOP: return OPclass_METHOP;
    case OA_UNOP_AUX: return OPclass_UNOP_AUX;
    }
    return OPclass_BASEOP;
}
#define op_class(o) goperl_op_class_of((const OP *)(o))

/* Flags of subs, globs, hashes and pad names that are bits of the words
 * above (cv.h, gv.h, hv.h, pad.h, sv.h). */
#define CvAUTOLOAD(cv) (CvFLAGS(cv) & CVf_AUTOLOAD)
#define CvAUTOLOAD_on(cv) GOPERL_CVFLAGS_CHANGE(cv, CVf_AUTOLOAD, 0)
#define CvAUTOLOAD_off(cv) GOPERL_CVFLAGS_CHANGE(cv, 0, CVf_AUTOLOAD)
#ifndef GvIMPORTED_SV
#define GvIMPORTED_SV(gv) (GvFLAGS(gv) & GVf_IMPORTED_SV)
#endif
#ifndef GvIMPORTED_AV
#define GvIMPORTED_AV(gv) (GvFLAGS(gv) & GVf_IMPORTED_AV)
#endif
#ifndef GvIMPORTED_HV
#define GvIMPORTED_HV(gv) (GvFLAGS(gv) & GVf_IMPORTED_HV)
#endif
#define GvIMPORTED_off(gv) ((void)GOPERL_GVFLAGS_CHANGE(gv, 0, GVf_IMPORTED))
#define GvIMPORTED_SV_off(gv) ((void)GOPERL_GVFLAGS_CHANGE(gv, 0, GVf_IMPORTED_SV))
#define GvIMPORTED_AV_off(gv) ((void)GOPERL_GVFLAGS_CHANGE(gv, 0, GVf_IMPORTED_AV))
#define GvIMPORTED_HV_off(gv) ((void)GOPERL_GVFLAGS_CHANGE(gv, 0, GVf_IMPORTED_HV))
#define GvIMPORTED_CV_off(gv) ((void)GOPERL_GVFLAGS_CHANGE(gv, 0, GVf_IMPORTED_CV))
/* No glob is in a pad on a perl without threads, and nothing is compiled
 * into a scalar (gv.h, sv.h outside the core). */
#define GvIN_PAD(gv) 0
#define GvIN_PAD_on(gv) NOOP
#define GvIN_PAD_off(gv) NOOP
#define SvCOMPILED(sv) 0
#define SvCOMPILED_on(sv)
#define SvCOMPILED_off(sv)
#define HvLAZYDEL(hv) (SvFLAGS(hv) & SVphv_LAZYDEL)
#define HvLAZYDEL_on(hv) ((void)GOPERL_FLAGS_CHANGE(hv, SVphv_LAZYDEL, 0))
#define HvLAZYDEL_off(hv) ((void)GOPERL_FLAGS_CHANGE(hv, 0, SVphv_LAZYDEL))
#ifndef SvPAD_TYPED
#define SvPAD_TYPED PadnameHasTYPE
#endif
#ifndef SvPAD_OUR
#define SvPAD_OUR(pn) cBOOL(PadnameOURSTASH(pn))
#endif
#define PAD_FAKELEX_ANON 1
#define PAD_FAKELEX_MULTI 2
#define PERL_SIGNALS_UNSAFE_FLAG 0x0001

/* The decimal digits an integer can have, and the version as a string
 * (perl.h). */
#define IV_BITS (IVSIZE * 8)
#define UV_BITS (UVSIZE * 8)
#ifndef BIT_DIGITS
#define BIT_DIGITS(N) (((N) * 146) / 485 + 1)
#endif
/* the digits of an integer type, and the characters of its decimal string
 * with a sign and the NUL (Stats::LikeR: `char tmp[TYPE_DIGITS(UV) + 1]`) */
#define TYPE_DIGITS(T) BIT_DIGITS(sizeof(T) * 8)
#define TYPE_CHARS(T) (TYPE_DIGITS(T) + 2)
#define IV_DIG (BIT_DIGITS(IV_BITS))
#define UV_DIG (BIT_DIGITS(UV_BITS))
#define PERL_VERSION_STRING \
    STRINGIFY(PERL_REVISION) "." STRINGIFY(PERL_VERSION) "." STRINGIFY(PERL_SUBVERSION)

/* The interpreter is not a DEBUGGING build: the debugging statements of
 * every letter compile to nothing (perl.h). */
#ifndef DEBUG_p
#define DEBUG_p(a)
#endif
#ifndef DEBUG_s
#define DEBUG_s(a)
#endif
#ifndef DEBUG_t
#define DEBUG_t(a)
#endif
#ifndef DEBUG_o
#define DEBUG_o(a)
#endif
#ifndef DEBUG_c
#define DEBUG_c(a)
#endif
#ifndef DEBUG_m
#define DEBUG_m(a)
#endif
#ifndef DEBUG_f
#define DEBUG_f(a)
#endif
#ifndef DEBUG_r
#define DEBUG_r(a)
#endif
#ifndef DEBUG_u
#define DEBUG_u(a)
#endif
#ifndef DEBUG_U
#define DEBUG_U(a)
#endif
#ifndef DEBUG_X
#define DEBUG_X(a)
#endif
#ifndef DEBUG_D
#define DEBUG_D(a)
#endif
#ifndef DEBUG_S
#define DEBUG_S(a)
#endif
#ifndef DEBUG_T
#define DEBUG_T(a)
#endif
#ifndef DEBUG_R
#define DEBUG_R(a)
#endif
#ifndef DEBUG_J
#define DEBUG_J(a)
#endif
#ifndef DEBUG_C
#define DEBUG_C(a)
#endif
#ifndef DEBUG_A
#define DEBUG_A(a)
#endif
#ifndef DEBUG_q
#define DEBUG_q(a)
#endif
#ifndef DEBUG_M
#define DEBUG_M(a)
#endif
#ifndef DEBUG_B
#define DEBUG_B(a)
#endif
#ifndef DEBUG_L
#define DEBUG_L(a)
#endif
#ifndef DEBUG_i
#define DEBUG_i(a)
#endif
#ifndef DEBUG_y
#define DEBUG_y(a)
#endif

/* The target of a pp function or an XSUB set to a value and left in the
 * place of the top of the stack (pp.h). */
#ifndef SETi
#define SETi(i) STMT_START { sv_setiv(TARG, (IV)(i)); SvSETMAGIC(TARG); SETs(TARG); } STMT_END
#endif
#ifndef SETu
#define SETu(u) STMT_START { sv_setuv(TARG, (UV)(u)); SvSETMAGIC(TARG); SETs(TARG); } STMT_END
#endif
#ifndef SETn
#define SETn(n) STMT_START { sv_setnv(TARG, (NV)(n)); SvSETMAGIC(TARG); SETs(TARG); } STMT_END
#endif
#ifndef SETp
#define SETp(p, l) STMT_START { sv_setpvn(TARG, (p), (l)); SvSETMAGIC(TARG); SETs(TARG); } STMT_END
#endif
#ifndef SETTARG
#define SETTARG STMT_START { SvSETMAGIC(TARG); SETs(TARG); } STMT_END
#endif

/* xsubpp's INTERFACE: the C function an XSUB stands for is kept in its
 * XSANY (XSUB.h). */
#define XSINTERFACE_CVT(ret, name) ret (*name)()
#define XSINTERFACE_CVT_ANON(ret) ret (*)()
#define dXSFUNCTION(ret) XSINTERFACE_CVT(ret, XSFUNCTION)
#define XSINTERFACE_FUNC(ret, cv, f) ((XSINTERFACE_CVT_ANON(ret))(f))
#define XSINTERFACE_FUNC_SET(cv, f) \
    CvXSUBANY(cv).any_dxptr = (void (*)(pTHX_ void *))(f)

/* A new scalar of a shared key (hv.h): the key's bytes, shared; a key
 * that was UTF-8 and is kept downgraded comes back upgraded, and a key
 * that is a scalar is that scalar's copy. */
#ifndef HEf_SVKEY
#define HEf_SVKEY -2
#endif
static SV *goperl_newSVhek(goperl_frame_t *f, const HEK *hek) __attribute__((unused));
static SV *goperl_newSVhek(goperl_frame_t *f, const HEK *hek) {
    (void)f;
    if (!hek) return &PL_sv_undef;
    if (HEK_LEN(hek) == HEf_SVKEY) return newSVsv(*(SV *const *)HEK_KEY(hek));
    if (HEK_FLAGS(hek) & HVhek_WASUTF8) {
        SV *sv = newSVpvn(HEK_KEY(hek), HEK_LEN(hek));
        sv_utf8_upgrade(sv);
        return sv;
    }
    return newSVpvn_share(HEK_KEY(hek), HEK_UTF8(hek) ? -(I32)HEK_LEN(hek) : (I32)HEK_LEN(hek),
                          HEK_HASH(hek));
}
#define newSVhek(hek) goperl_newSVhek(_gof, (const HEK *)(hek))
#define newSVhek_mortal(hek) sv_2mortal(newSVhek(hek))

/* An array with room for size elements and none yet (av.h). */
#define newAV_alloc_x(size)                                                 \
    ({ AV *const goperl_newAV_alloc_av_ = newAV();                          \
       av_extend(goperl_newAV_alloc_av_, (SSize_t)(size) - 1);              \
       goperl_newAV_alloc_av_; })
#define newAV_alloc_xz(size) newAV_alloc_x(size)
#define newSV_type_mortal(type) sv_2mortal(newSV_type(type))
#define sv_setpvn_fresh(sv, p, l) sv_setpvn((sv), (p), (l))
#define lex_stuff_pv(pv, flags)                                             \
    ({ const char *const goperl_lex_stuff_pv_ = (pv);                       \
       lex_stuff_pvn(goperl_lex_stuff_pv_, __builtin_strlen(goperl_lex_stuff_pv_), (flags)); })

/* A hint by a key that is a scalar (cop_hints_fetch_sv): its bytes, and
 * whether they are UTF-8. */
#ifndef REFCOUNTED_HE_KEY_UTF8
#define REFCOUNTED_HE_KEY_UTF8 0x00000001
#endif
#define cop_hints_fetch_sv(cop, keysv, hash, flags)                         \
    ({ SV *const goperl_chf_key_ = (keysv);                                 \
       STRLEN goperl_chf_len_;                                              \
       const char *const goperl_chf_pv_ = SvPV(goperl_chf_key_, goperl_chf_len_); \
       cop_hints_fetch_pvn((cop), goperl_chf_pv_, goperl_chf_len_, (hash),  \
                           (flags) | (SvUTF8(goperl_chf_key_) ? REFCOUNTED_HE_KEY_UTF8 : 0)); })

/* One continuation byte more of a UTF-8 sequence (utf8.h). */
#ifndef UTF_ACCUMULATION_SHIFT
#define UTF_ACCUMULATION_SHIFT 6
#endif
#ifndef UTF_CONTINUATION_MASK
#define UTF_CONTINUATION_MASK ((U8)0x3f)
#endif
#ifndef NATIVE_UTF8_TO_I8
#define NATIVE_UTF8_TO_I8(ch) ((U8)(ch))
#endif
#ifndef I8_TO_NATIVE_UTF8
#define I8_TO_NATIVE_UTF8(ch) ((U8)(ch))
#endif
#define UTF8_ACCUMULATE(old, new) \
    (((old) << UTF_ACCUMULATION_SHIFT) | (NATIVE_UTF8_TO_I8(new) & UTF_CONTINUATION_MASK))
/* The marks and masks of the bytes of a sequence, and the code points
 * that are surrogates (utf8.h); on this ASCII platform a Latin-1 character
 * is itself. */
#define UTF_CONTINUATION_MARK (~(0x40 | UTF_CONTINUATION_MASK) & 0xff)
#define UTF_START_MARK(len) ((U8) ~(0xFF >> (len)))
#define UTF_START_MASK(len) (0xFF >> (len))
#define UNICODE_IS_SURROGATE(uv) \
    ((UV)(uv) >= UNICODE_SURROGATE_FIRST && (UV)(uv) <= UNICODE_SURROGATE_LAST)
#define LATIN1_TO_NATIVE(ch) ((U8)(ch))
#define NATIVE_TO_LATIN1(ch) ((U8)(ch))
#define ASCII_TO_NATIVE(ch) LATIN1_TO_NATIVE(ch)
#define NATIVE_TO_ASCII(ch) NATIVE_TO_LATIN1(ch)
#define NATIVE_TO_UTF(ch) NATIVE_UTF8_TO_I8(ch)
#define UTF_TO_NATIVE(ch) I8_TO_NATIVE_UTF8(ch)
/* the encoder's flags put together, and their other names (utf8.h) */
#define UNICODE_WARN_ILLEGAL_C9_INTERCHANGE (UNICODE_WARN_SURROGATE | UNICODE_WARN_SUPER)
#define UNICODE_WARN_ILLEGAL_INTERCHANGE (UNICODE_WARN_ILLEGAL_C9_INTERCHANGE | UNICODE_WARN_NONCHAR)
#define UNICODE_DISALLOW_ILLEGAL_C9_INTERCHANGE (UNICODE_DISALLOW_SURROGATE | UNICODE_DISALLOW_SUPER)
#define UNICODE_DISALLOW_ILLEGAL_INTERCHANGE \
    (UNICODE_DISALLOW_ILLEGAL_C9_INTERCHANGE | UNICODE_DISALLOW_NONCHAR)
#define UNICODE_WARN_ABOVE_31_BIT UNICODE_WARN_PERL_EXTENDED
#define UNICODE_DISALLOW_ABOVE_31_BIT UNICODE_DISALLOW_PERL_EXTENDED
#define UNICODE_GOT_SURROGATE UNICODE_DISALLOW_SURROGATE
#define UNICODE_GOT_NONCHAR UNICODE_DISALLOW_NONCHAR
#define UNICODE_GOT_SUPER UNICODE_DISALLOW_SUPER
#define UNICODE_GOT_PERL_EXTENDED UNICODE_DISALLOW_PERL_EXTENDED

/* Bytes compared with a UTF-8 string of the same characters (utf8.c):
 * -1 or +1 when one is a prefix of the other, -2 or +2 when they differ
 * (and -2 for UTF-8 that does not come down to bytes). */
static int goperl_bytes_cmp_utf8(goperl_frame_t *f, const U8 *b, STRLEN blen, const U8 *u,
                                 STRLEN ulen) __attribute__((unused));
static int goperl_bytes_cmp_utf8(goperl_frame_t *f, const U8 *b, STRLEN blen, const U8 *u,
                                 STRLEN ulen) {
    const U8 *const bend = b + blen;
    const U8 *const uend = u + ulen;
    while (b < bend && u < uend) {
        U8 c = *u++;
        if (c >= 0x80) {
            if (c == 0xc2 || c == 0xc3) {
                if (u < uend) {
                    const U8 c1 = *u++;
                    if ((c1 & 0xc0) == 0x80) {
                        c = (U8)(((c & 0x03) << 6) | (c1 & 0x3f));
                    } else {
                        Perl_ck_warner_d(WARN_UTF8,
                            "Malformed UTF-8 character: \\x%02x\\x%02x (unexpected "
                            "non-continuation byte 0x%02x, immediately after start "
                            "byte 0x%02x; need 2 bytes, got 1)%s%s",
                            u[-2], c1, c1, u[-2], PL_op ? " in " : "", PL_op ? OP_DESC(PL_op) : "");
                        return -2;
                    }
                } else {
                    Perl_ck_warner_d(WARN_UTF8,
                        "Malformed UTF-8 character (unexpected end of string)%s%s",
                        PL_op ? " in " : "", PL_op ? OP_DESC(PL_op) : "");
                    return -2;
                }
            } else {
                return -2;
            }
        }
        if (*b != c) return *b < c ? -2 : +2;
        ++b;
    }
    if (b == bend && u == uend) return 0;
    return b < bend ? +1 : -1;
}
#define bytes_cmp_utf8(b, blen, u, ulen) \
    goperl_bytes_cmp_utf8(_gof, (const U8 *)(b), (STRLEN)(blen), (const U8 *)(u), (STRLEN)(ulen))

/* A struct tm with the zone of now, and the normalising of one without
 * the zone (util.c's init_tm and mini_mktime). */
static void goperl_init_tm(struct tm *ptm) __attribute__((unused));
static void goperl_init_tm(struct tm *ptm) {
    const time_t now = time(NULL);
    struct tm t;
    if (localtime_r(&now, &t)) *ptm = t;
}
#define init_tm(ptm) goperl_init_tm(ptm)
static void goperl_mini_mktime(struct tm *ptm) __attribute__((unused));
static void goperl_mini_mktime(struct tm *ptm) {
    enum { DAYS_PER_YEAR_ = 365, DAYS_PER_QYEAR_ = 4 * 365 + 1,
           DAYS_PER_CENT_ = 25 * (4 * 365 + 1) - 1,
           DAYS_PER_QCENT_ = 4 * (25 * (4 * 365 + 1) - 1) + 1,
           SECS_PER_HOUR_ = 60 * 60, SECS_PER_DAY_ = 24 * 60 * 60,
           YEAR_ADJUST_ = 4 * 153 / 5 + 1, WEEKDAY_BIAS_ = 6 };
    int yearday, secs, month, mday, year, jday, odd_cent, odd_year;
    year = 1900 + ptm->tm_year;
    month = ptm->tm_mon;
    mday = ptm->tm_mday;
    jday = 0;
    if (month >= 2) month += 2;
    else month += 14, year--;
    yearday = DAYS_PER_YEAR_ * year + year / 4 - year / 100 + year / 400;
    yearday += month * 153 / 5 + mday + jday;
    if ((unsigned)ptm->tm_sec <= 60) {
        secs = 0;
    } else {
        secs = ptm->tm_sec;
        ptm->tm_sec = 0;
    }
    secs += 60 * ptm->tm_min;
    secs += SECS_PER_HOUR_ * ptm->tm_hour;
    if (secs < 0) {
        if (secs - (secs / SECS_PER_DAY_ * SECS_PER_DAY_) < 0) {
            yearday += (secs / SECS_PER_DAY_) - 1;
            secs -= SECS_PER_DAY_ * (secs / SECS_PER_DAY_ - 1);
        } else {
            yearday += (secs / SECS_PER_DAY_);
            secs -= SECS_PER_DAY_ * (secs / SECS_PER_DAY_);
        }
    } else if (secs >= SECS_PER_DAY_) {
        yearday += (secs / SECS_PER_DAY_);
        secs %= SECS_PER_DAY_;
    }
    ptm->tm_hour = secs / SECS_PER_HOUR_;
    secs %= SECS_PER_HOUR_;
    ptm->tm_min = secs / 60;
    secs %= 60;
    ptm->tm_sec += secs;
    jday = yearday;
    yearday -= YEAR_ADJUST_;
    year = (yearday / DAYS_PER_QCENT_) * 400;
    yearday %= DAYS_PER_QCENT_;
    odd_cent = yearday / DAYS_PER_CENT_;
    year += odd_cent * 100;
    yearday %= DAYS_PER_CENT_;
    year += (yearday / DAYS_PER_QYEAR_) * 4;
    yearday %= DAYS_PER_QYEAR_;
    odd_year = yearday / DAYS_PER_YEAR_;
    year += odd_year;
    yearday %= DAYS_PER_YEAR_;
    if (!yearday && (odd_cent == 4 || odd_year == 4)) {
        month = 1;
        yearday = 29;
    } else {
        yearday += YEAR_ADJUST_;
        month = yearday * 5 / 153;
        yearday -= month * 153 / 5;
        if (month > 13) {
            month -= 14;
            year++;
        } else {
            month -= 2;
        }
    }
    ptm->tm_year = year - 1900;
    if (yearday) {
        ptm->tm_mday = yearday;
        ptm->tm_mon = month;
    } else {
        ptm->tm_mday = 31;
        ptm->tm_mon = month - 1;
    }
    year--;
    yearday = year * DAYS_PER_YEAR_ + year / 4 - year / 100 + year / 400;
    yearday += 14 * 153 / 5 + 1;
    ptm->tm_yday = jday - yearday;
    ptm->tm_wday = (jday + WEEKDAY_BIAS_) % 7;
}
#define mini_mktime(ptm) goperl_mini_mktime(ptm)
#define Perl_mini_mktime(ptm) goperl_mini_mktime(ptm)

/* ---- more interpreter variables (intrpvar.h) ------------------------------- */

#define GOPERL_PL_ARGVOUTGV 100
#define GOPERL_PL_GLOBALSTASH 101
#define GOPERL_PL_SIGNALS 102
#define GOPERL_PL_OP_MASK 103
#define GOPERL_PL_LASTSTATVAL 104
#define GOPERL_PL_LASTSTYPE 105
#define GOPERL_PL_STATNAME 106
#define GOPERL_PL_CURSTACK 107
#define GOPERL_PL_MAINSTACK 108
/* Lvalues like the others (goperl_plint_ref): read from the interpreter,
 * stored back before its next operation, and saved by the interpreter
 * (SAVEVPTR(PL_op_mask), save_hptr(&PL_defstash)). */
#define PL_argvoutgv (*(GV **)goperl_plsv_ref(_gof, GOPERL_PL_ARGVOUTGV))
#define PL_globalstash (*(HV **)goperl_plsv_ref(_gof, GOPERL_PL_GLOBALSTASH))
#define PL_signals (*(U32 *)goperl_plint_ref(_gof, GOPERL_PL_SIGNALS))
#define PL_laststatval (*(I32 *)goperl_plint_ref(_gof, GOPERL_PL_LASTSTATVAL))
#define PL_laststype (*(U16 *)goperl_plint_ref(_gof, GOPERL_PL_LASTSTYPE))
#define PL_statname (*goperl_plsv_ref(_gof, GOPERL_PL_STATNAME))
#define PL_curstack ((AV *)GOPERL_SV(GOPERL_PLVAR(GOPERL_PL_CURSTACK)))
#define PL_mainstack ((AV *)GOPERL_SV(GOPERL_PLVAR(GOPERL_PL_MAINSTACK)))
#undef PL_defstash
#define PL_defstash (*(HV **)goperl_plsv_ref(_gof, GOPERL_PL_DEFSTASH))
#undef PL_tmps_floor
#define PL_tmps_floor (*(SSize_t *)goperl_plint_ref(_gof, GOPERL_PL_TMPS_FLOOR))

/* ---- the save stack and the temporaries stack -----------------------------
 * The temporaries stack is the interpreter's array of SV tokens, read and
 * written in place. The save stack's entries are shown in native code's
 * pointers (see goperl_ss_view): an SV, an array, a pad is its token, which
 * is what an SV* is here; a location a save names, or a string, is the
 * address of the interpreter's memory as this process has it, so that it
 * compares with PL_curpad and SvPVX; a destructor of native code's is a
 * function native code may call (goperl_dtor_tramp), which runs it; an op,
 * a COP or a function of the interpreter's stays the interpreter's
 * address. Native code that takes entries off (PL_savestack_ix -= n) or
 * puts them on (dSS_ADD ... SS_ADD_END) moves the index, which is an
 * lvalue written back before the next operation, as the entries it wrote
 * are. Room is the interpreter's to make (the stacks move when they grow):
 * SSCHECK and EXTEND_MORTAL, and dSS_ADD for the SS_MAXPUSH entries perl
 * keeps spare above the top. */
#define GOPERL_PL_SAVESTACK_ADDR 119
#define GOPERL_PL_TMPS_STACK_ADDR 120
#define GOPERL_PL_SAVESTACK_MAX 121
#define GOPERL_PL_TMPS_MAX 122
#define GOPERL_PL_CATCH 123
#define GOPERL_CALL_SS_ROOM 36
/* The interpreter's JMPENV that is the top of its chain, as an address of
 * its memory: native code compares it with an eval context's (whose
 * cur_top_env it is when the eval was entered under it). */
#define GOPERL_PL_TOP_ENV 125
#define PL_top_env ((JMPENV *)(uintptr_t)GOPERL_PLVAR(GOPERL_PL_TOP_ENV))
/* Where the interpreter goes on after a die an eval caught (die_unwind sets
 * them; a docatch of native code clears them and runs from the op): the op
 * as a shadow, and the JMPENV as the interpreter's address. Lvalues. */
#define GOPERL_PL_RESTARTOP 74
#define GOPERL_PL_RESTARTJMPENV 127
static OP **goperl_restartop_ref(goperl_frame_t *f) __attribute__((unused));
static OP **goperl_restartop_ref(goperl_frame_t *f) {
    goperl_plshadow_t *s = &goperl_plshadow_v[GOPERL_PL_RESTARTOP];
    if (!s->live || s->val == s->shadow) {
        const uint64_t tok =
            goperl_do_op(f, GOPERL_OP_PLVAR_GET, (uint64_t)GOPERL_PL_RESTARTOP, 0, 0, 0);
        s->val = s->shadow = (uint64_t)(uintptr_t)(tok ? goperl_op_shadow(f, tok) : 0);
        goperl_plshadow_set_live(s);
    }
    return (OP **)&s->val;
}
#define PL_restartop (*goperl_restartop_ref(_gof))
#define PL_restartjmpenv (*(JMPENV **)goperl_plsv_ref(_gof, GOPERL_PL_RESTARTJMPENV))
/* The PMOP of the last match, as the interpreter's address of it (native
 * code keeps it and puts it back, as Future::AsyncAwait does across an
 * await; it is no op shadow to look into). */
#define PL_curpm (*(PMOP **)goperl_plsv_ref(_gof, GOPERL_PL_CURPM))
#define GOPERL_CALL_TMPS_ROOM 37
#undef PL_savestack_ix
#define PL_savestack_ix (*(I32 *)goperl_plint_ref(_gof, GOPERL_PL_SAVESTACK_IX))
#undef PL_tmps_ix
#define PL_tmps_ix (*(SSize_t *)goperl_plint_ref(_gof, GOPERL_PL_TMPS_IX))
static ANY *goperl_ss_view(goperl_frame_t *f);
#define PL_savestack (goperl_ss_view(_gof))
#define PL_savestack_max ((I32)(int64_t)GOPERL_PLVAR(GOPERL_PL_SAVESTACK_MAX))
#define PL_tmps_stack ((SV **)goperl_guest_mem(_gof, GOPERL_PLVAR(GOPERL_PL_TMPS_STACK_ADDR)))
#define PL_tmps_max ((SSize_t)(int64_t)GOPERL_PLVAR(GOPERL_PL_TMPS_MAX))
#define CATCH_GET (GOPERL_PLVAR(GOPERL_PL_CATCH) != 0)
/* room for n entries above the top of the save stack (and the SS_MAXPUSH
 * spare ones perl keeps above it); -> the save stack as native code sees it */
static ANY *goperl_ss_room(goperl_frame_t *f, int n) {
    (void)goperl_do_op3(f, GOPERL_OP_CALL, 0, (uint64_t)(int64_t)n, (uint64_t)GOPERL_CALL_SS_ROOM, 0, 0);
    return goperl_ss_view(f);
}
#define SSCHECK(need) ((void)goperl_ss_room(_gof, (int)(need)))
#define SSGROW(need) SSCHECK(need)
#define EXTEND_MORTAL(n)                                                        \
    ((void)goperl_do_op3(_gof, GOPERL_OP_CALL, 0, (uint64_t)(int64_t)(n),         \
                         (uint64_t)GOPERL_CALL_TMPS_ROOM, 0, 0))
/* the save types of this perl (scope_types.h) */
#define SAVE_TIGHT_SHIFT 6
#define SAVE_MASK 0x3F
#define SS_MAXPUSH 4
#define SAVEt_ALLOC 0
#define SAVEt_CLEARPADRANGE 1
#define SAVEt_CLEARSV 2
#define SAVEt_REGCONTEXT 3
#define SAVEt_PADSV_NULL 4
#define SAVEt_TMPSFLOOR 5
#define SAVEt_BOOL 6
#define SAVEt_COMPILE_WARNINGS 7
#define SAVEt_CURCOP_WARNINGS 8
#define SAVEt_COMPPAD 9
#define SAVEt_FREECOPHH 10
#define SAVEt_FREEOP 11
#define SAVEt_FREEPV 12
#define SAVEt_FREESV 13
#define SAVEt_I16 14
#define SAVEt_I32_SMALL 15
#define SAVEt_I8 16
#define SAVEt_INT_SMALL 17
#define SAVEt_MORTALIZESV 18
#define SAVEt_NSTAB 19
#define SAVEt_OP 20
#define SAVEt_PARSER 21
#define SAVEt_STACK_POS 22
#define SAVEt_READONLY_OFF 23
#define SAVEt_FREEPADNAME 24
#define SAVEt_STRLEN_SMALL 25
#define SAVEt_FREERCPV 26
#define SAVEt_FREE_REXC_STATE 27
#define SAVEt_PADSV 28
#define SAVEt_AV 29
#define SAVEt_DESTRUCTOR 30
#define SAVEt_DESTRUCTOR_X 31
#define SAVEt_GENERIC_PVREF 32
#define SAVEt_GENERIC_SVREF 33
#define SAVEt_GP 34
#define SAVEt_GVSV 35
#define SAVEt_HINTS 36
#define SAVEt_HPTR 37
#define SAVEt_HV 38
#define SAVEt_I32 39
#define SAVEt_INT 40
#define SAVEt_ITEM 41
#define SAVEt_IV 42
#define SAVEt_PPTR 43
#define SAVEt_SAVESWITCHSTACK 44
#define SAVEt_SHARED_PVREF 45
#define SAVEt_SPTR 46
#define SAVEt_STRLEN 47
#define SAVEt_SV 48
#define SAVEt_SVREF 49
#define SAVEt_VPTR 50
#define SAVEt_ADELETE 51
#define SAVEt_APTR 52
#define SAVEt_RCPV 53
#define SAVEt_HELEM 54
#define SAVEt_PADSV_AND_MORTALIZE 55
#define SAVEt_SET_SVFLAGS 56
#define SAVEt_GVSLOT 57
#define SAVEt_AELEM 58
#define SAVEt_DELETE 59
#define SAVEt_HINTS_HH 60
/* The words of each save type below its type word, the arguments in the
 * order perl pushes them (scope.c): R as it is (an integer, the token of an
 * SV, an array, a hash, a glob or a pad; an op, a COP, a parser, a
 * refcounted_he, a function of the interpreter's, a value of no known
 * kind), A an address of the interpreter's memory (a location saved, a
 * string), F and P a destructor's function and its argument. SAVEt_ALLOC
 * and SAVEt_REGCONTEXT have as many raw words below the type word as it
 * says above SAVE_TIGHT_SHIFT. */
static const char *const goperl_ss_kinds[61] = {
    "", "", "", "", "",                /* ALLOC CLEARPADRANGE CLEARSV REGCONTEXT PADSV_NULL */
    "R", "A", "A", "A", "R",           /* TMPSFLOOR BOOL COMPILE_WARNINGS CURCOP_WARNINGS COMPPAD */
    "R", "R", "A", "R", "A",           /* FREECOPHH FREEOP FREEPV FREESV I16 */
    "A", "A", "A", "R", "R",           /* I32_SMALL I8 INT_SMALL MORTALIZESV NSTAB */
    "R", "R", "R", "R", "R",           /* OP PARSER STACK_POS READONLY_OFF FREEPADNAME */
    "A", "A", "R", "R",                /* STRLEN_SMALL FREERCPV FREE_REXC_STATE PADSV */
    "RR", "RR", "FP", "AA", "AR",      /* AV DESTRUCTOR DESTRUCTOR_X GENERIC_PVREF GENERIC_SVREF */
    "RR", "RR", "RR", "RA", "RR",      /* GP GVSV HINTS HPTR HV */
    "RA", "RA", "RR", "RA", "AA",      /* I32 INT ITEM IV PPTR */
    "RR", "AA", "RA", "RA", "RR",      /* SAVESWITCHSTACK SHARED_PVREF SPTR STRLEN SV */
    "AR", "RA", "RR", "RA", "AA",      /* SVREF VPTR ADELETE APTR RCPV */
    "RRR", "RRR", "RRR", "RAR", "RRR", /* HELEM PADSV_AND_MORTALIZE SET_SVFLAGS GVSLOT AELEM */
    "ARR", "RRR"                       /* DELETE HINTS_HH */
};
/* The function the interpreter's save stack names for a destructor of the
 * host's, with the id (1-based) as the argument (perl-wasm PLVAR 130). Its
 * address is the interpreter's binary's, the same in every instance. */
#define GOPERL_PL_DTOR_THUNK 130
static uint64_t goperl_dtor_thunk_addr(goperl_frame_t *f) {
    static uint64_t addr;
    if (!addr) addr = goperl_xs_op(f, GOPERL_OP_PLVAR_GET, (uint64_t)GOPERL_PL_DTOR_THUNK, 0, 0, 0);
    return addr;
}
/* A destructor of the host's as native code that reads it off the save
 * stack sees it: called with the id, it runs the destructor and gives the
 * id up, as the interpreter's call of the thunk does when the scope ends
 * (Future::AsyncAwait runs the finally block of a sub it cancels so); put
 * on the save stack again, it is the thunk with the id. */
static void goperl_dtor_tramp(pTHX_ void *idp) {
    const uint32_t id1 = (uint32_t)(uintptr_t)idp;
    if (id1 == 0 || id1 > GOPERL_DTOR_MAX) return;
    const goperl_dtor_t d = goperl_dtor_v[id1 - 1];
    goperl_dtor_v[id1 - 1].used = 0;
    if (!d.used || !d.fn) return;
    if (d.used == GOPERL_DTOR_PLAIN)
        ((void (*)(void *))(void (*)(void))d.fn)(d.arg);
    else
        d.fn(aTHX_ d.arg);
}
/* The save stack as native code sees it: the entries below the
 * interpreter's top, each word in native code's form (goperl_ss_kinds),
 * read when native code first looks at the save stack after an operation,
 * and what it was then, for the write-back before the next operation. One
 * for each thread of the module's files (a weak, hidden definition each
 * file has, which the link makes one). */
typedef struct goperl_ssview {
    ANY *view;       /* the entries in native code's form, room for cap */
    ANY *conv;       /* the same as loaded */
    int32_t cap;
    int32_t n;       /* the interpreter's PL_savestack_ix when loaded */
    uint64_t base_g; /* the address of the interpreter's save stack then */
    uint64_t gen;    /* the generation it was loaded in */
    void *inst;      /* the instance it was loaded from */
} goperl_ssview_t;
GOPERL_C __attribute__((weak, visibility("hidden"))) __thread goperl_ssview_t goperl_ssview_tl = {0};
/* The kinds of the words of the group whose type word is at t (the words
 * t - argc .. t - 1, below them the raw words of an allocation): -> the
 * index of the group's lowest word. */
static int32_t goperl_ss_group(goperl_frame_t *f, const ANY *ss, int32_t t, const char **kinds) {
    const UV uv = ss[t].any_uv;
    const U8 type = (U8)(uv & SAVE_MASK);
    if (type > SAVEt_HINTS_HH)
        goperl_croakf(f, "the save stack has an entry of save type %u, which this perl does not have",
                      (unsigned)type);
    *kinds = goperl_ss_kinds[type];
    int32_t lo = t - (int32_t)strlen(*kinds);
    if (type == SAVEt_ALLOC || type == SAVEt_REGCONTEXT) lo -= (int32_t)(uv >> SAVE_TIGHT_SHIFT);
    if (lo < 0) goperl_croakf(f, "the save stack has an entry that goes below its bottom");
    return lo;
}
static ANY *goperl_ss_view(goperl_frame_t *f) {
    goperl_ssview_t *const v = &goperl_ssview_tl;
    if (v->view && v->inst == f->inst && v->gen == goperl_gen_v) return v->view;
    /* not operations of the funnel: nothing is to be written back now, and
     * the generation stays */
    const uint64_t base = goperl_xs_op(f, GOPERL_OP_PLVAR_GET, GOPERL_PL_SAVESTACK_ADDR, 0, 0, 0);
    const int32_t max = (int32_t)(int64_t)goperl_xs_op(f, GOPERL_OP_PLVAR_GET, GOPERL_PL_SAVESTACK_MAX, 0, 0, 0);
    const int32_t n = (int32_t)(int64_t)goperl_xs_op(f, GOPERL_OP_PLVAR_GET, GOPERL_PL_SAVESTACK_IX, 0, 0, 0);
    const int32_t cap = max + 1 + SS_MAXPUSH;
    if (!base || n < 0 || n > cap) goperl_croakf(f, "the interpreter did not give its save stack");
    if (cap > v->cap) {
        ANY *const nv = (ANY *)realloc(v->view, sizeof(ANY) * (size_t)cap);
        if (!nv) goperl_croakf(f, "Out of memory!");
        v->view = nv;
        ANY *const nc = (ANY *)realloc(v->conv, sizeof(ANY) * (size_t)cap);
        if (!nc) goperl_croakf(f, "Out of memory!");
        v->conv = nc;
        v->cap = cap;
    }
    ANY *const ss = v->view;
    memcpy(ss, goperl_guest_mem(f, base), sizeof(ANY) * (size_t)n);
    memset(ss + n, 0, sizeof(ANY) * (size_t)(v->cap - n));
    /* each group from the top down, its words in native code's form */
    const uint64_t thunk = goperl_dtor_thunk_addr(f);
    for (int32_t t = n - 1; t >= 0;) {
        const char *kinds;
        const int32_t lo = goperl_ss_group(f, ss, t, &kinds);
        const int32_t a = t - (int32_t)strlen(kinds);
        for (int32_t k = 0; kinds[k]; k++) {
            ANY *const e = &ss[a + k];
            if (kinds[k] == 'A' && e->any_uv)
                e->any_ptr = (void *)(f->membase + e->any_uv);
            else if (kinds[k] == 'F' && e->any_uv == thunk)
                e->any_dxptr = goperl_dtor_tramp;
        }
        t = lo - 1;
    }
    memcpy(v->conv, ss, sizeof(ANY) * (size_t)v->cap);
    v->n = n;
    v->base_g = base;
    v->gen = goperl_gen_v;
    v->inst = f->inst;
    return ss;
}
/* An address of the interpreter's memory in a word of the save stack, as
 * the interpreter has it. */
static uint64_t goperl_ss_addr_out(goperl_frame_t *f, const void *p) {
    if (!p) return 0;
    const uint64_t g = goperl_guest_addr(f, p);
    if (!g)
        goperl_croakf(f, "an entry of the save stack was given a pointer that is not in the "
                         "interpreter's memory, where the interpreter cannot reach it");
    return g;
}
/* What native code wrote into the save stack goes to the interpreter
 * (before every operation, goperl_flush): the words that differ from what
 * was loaded, and every word between the top it was loaded with and the top
 * native code moved it to, each in the interpreter's form. A destructor of
 * native code's own pushed raw is registered as save_destructor_x does. */
static void goperl_ss_view_flush(goperl_frame_t *f) {
    goperl_ssview_t *const v = &goperl_ssview_tl;
    if (!v->view || v->inst != f->inst || v->gen != goperl_gen_v) return;
    const goperl_plshadow_t *const s = &goperl_plshadow_v[GOPERL_PL_SAVESTACK_IX];
    const int32_t top = s->live ? (int32_t)(int64_t)s->val : v->n;
    if (top < 0 || top > v->cap) goperl_croakf(f, "PL_savestack_ix is beyond the save stack");
    int32_t low = top;
    for (int32_t i = 0; i < top; i++)
        if (i >= v->n || memcmp(&v->view[i], &v->conv[i], sizeof(ANY))) {
            low = i;
            break;
        }
    if (low == top) return;
    ANY *const ss = v->view;
    uint64_t *const g = (uint64_t *)goperl_guest_mem(f, v->base_g);
    const uint64_t thunk = goperl_dtor_thunk_addr(f);
    for (int32_t t = top - 1; t >= low;) {
        const char *kinds;
        const int32_t lo = goperl_ss_group(f, ss, t, &kinds);
        const int32_t a = t - (int32_t)strlen(kinds);
        for (int32_t i = (lo > low ? lo : low); i <= t; i++) {
            if (i < v->n && !memcmp(&ss[i], &v->conv[i], sizeof(ANY))) continue;
            const char kind = i >= a && i < t ? kinds[i - a] : 'R';
            uint64_t w = ss[i].any_uv;
            if (kind == 'A') {
                w = goperl_ss_addr_out(f, ss[i].any_ptr);
            } else if (kind == 'F' || kind == 'P') {
                /* the pair: the thunk with an id, as the view showed it, or
                 * a destructor of native code's own, registered now */
                const int32_t fi = kind == 'F' ? i : i - 1;
                void (*const fn)(pTHX_ void *) = ss[fi].any_dxptr;
                void *const arg = ss[fi + 1].any_ptr;
                uint64_t gf = (uint64_t)(uintptr_t)fn, ga = (uint64_t)(uintptr_t)arg;
                if (fn == goperl_dtor_tramp) {
                    gf = thunk;
                } else if ((uint64_t)(uintptr_t)fn != thunk) {
                    int32_t id = -1;
                    for (int32_t k = 0; k < GOPERL_DTOR_MAX && id < 0; k++) {
                        const int32_t j = (goperl_dtor_hint_v + k) % GOPERL_DTOR_MAX;
                        if (!goperl_dtor_v[j].used) id = j;
                    }
                    if (id < 0) goperl_croakf(f, "save_destructor_x: destructor table full");
                    goperl_dtor_hint_v = id + 1;
                    goperl_dtor_v[id].fn = fn;
                    goperl_dtor_v[id].arg = arg;
                    goperl_dtor_v[id].used = GOPERL_DTOR_X;
                    gf = thunk;
                    ga = (uint64_t)(id + 1);
                    ss[fi].any_dxptr = goperl_dtor_tramp;
                    ss[fi + 1].any_ptr = (void *)(uintptr_t)ga;
                }
                g[fi] = gf;
                g[fi + 1] = ga;
                v->conv[fi] = ss[fi];
                v->conv[fi + 1] = ss[fi + 1];
                if (kind == 'F') i++;
                continue;
            }
            g[i] = w;
            v->conv[i] = ss[i];
        }
        t = lo - 1;
    }
    if (top > v->n) v->n = top;
}
/* the raw pushes of scope.h: entries written in place above the top, the
 * index moved at the end (room: SSCHECK first, or the spare entries) */
#define dSS_ADD                                         \
    I32 ix = PL_savestack_ix;                           \
    ANY *ssp = goperl_ss_room(_gof, SS_MAXPUSH) + ix
#define SS_ADD_INT(i) ((ssp++)->any_i32 = (I32)(i))
#define SS_ADD_LONG(i) ((ssp++)->any_long = (long)(i))
#define SS_ADD_BOOL(p) ((ssp++)->any_bool = (p))
#define SS_ADD_IV(i) ((ssp++)->any_iv = (IV)(i))
#define SS_ADD_UV(u) ((ssp++)->any_uv = (UV)(u))
#define SS_ADD_PTR(p) ((ssp++)->any_ptr = (void *)(p))
#define SS_ADD_DPTR(p) ((ssp++)->any_dptr = (p))
#define SS_ADD_DXPTR(p) ((ssp++)->any_dxptr = (p))
#define SS_ADD_END(need) (ix += (need), PL_savestack_ix = ix)
#define SSPUSHINT(i) (PL_savestack[PL_savestack_ix++].any_i32 = (I32)(i))
#define SSPUSHLONG(i) (PL_savestack[PL_savestack_ix++].any_long = (long)(i))
#define SSPUSHBOOL(p) (PL_savestack[PL_savestack_ix++].any_bool = (p))
#define SSPUSHIV(i) (PL_savestack[PL_savestack_ix++].any_iv = (IV)(i))
#define SSPUSHUV(u) (PL_savestack[PL_savestack_ix++].any_uv = (UV)(u))
#define SSPUSHPTR(p) (PL_savestack[PL_savestack_ix++].any_ptr = (void *)(p))
#define SSPUSHDPTR(p) (PL_savestack[PL_savestack_ix++].any_dptr = (p))
#define SSPUSHDXPTR(p) (PL_savestack[PL_savestack_ix++].any_dxptr = (p))
#define SSPOPINT (PL_savestack[--PL_savestack_ix].any_i32)
#define SSPOPLONG (PL_savestack[--PL_savestack_ix].any_long)
#define SSPOPBOOL (PL_savestack[--PL_savestack_ix].any_bool)
#define SSPOPIV (PL_savestack[--PL_savestack_ix].any_iv)
#define SSPOPUV (PL_savestack[--PL_savestack_ix].any_uv)
#define SSPOPPTR (PL_savestack[--PL_savestack_ix].any_ptr)
#define SSPOPDPTR (PL_savestack[--PL_savestack_ix].any_dptr)
#define SSPOPDXPTR (PL_savestack[--PL_savestack_ix].any_dxptr)
/* the pushes of a pointer, two, and an integer with a pointer, with the
 * save type on top (scope.c) */
static void goperl_save_pushptr(goperl_frame_t *f, const void *ptr, int type) {
    ANY *ss = goperl_ss_room(f, 2);
    I32 *const ixp = (I32 *)goperl_plint_ref(f, GOPERL_PL_SAVESTACK_IX);
    ss[*ixp].any_ptr = (void *)ptr;
    ss[*ixp + 1].any_uv = (UV)type;
    *ixp += 2;
}
static void goperl_save_pushptrptr(goperl_frame_t *f, const void *p1, const void *p2, int type) {
    ANY *ss = goperl_ss_room(f, 3);
    I32 *const ixp = (I32 *)goperl_plint_ref(f, GOPERL_PL_SAVESTACK_IX);
    ss[*ixp].any_ptr = (void *)p1;
    ss[*ixp + 1].any_ptr = (void *)p2;
    ss[*ixp + 2].any_uv = (UV)type;
    *ixp += 3;
}
static void goperl_save_pushi32ptr(goperl_frame_t *f, I32 i, const void *ptr, int type) {
    ANY *ss = goperl_ss_room(f, 3);
    I32 *const ixp = (I32 *)goperl_plint_ref(f, GOPERL_PL_SAVESTACK_IX);
    ss[*ixp].any_i32 = i;
    ss[*ixp + 1].any_ptr = (void *)ptr;
    ss[*ixp + 2].any_uv = (UV)type;
    *ixp += 3;
}
#define save_pushptr(ptr, type) goperl_save_pushptr(_gof, (const void *)(ptr), (int)(type))
#define Perl_save_pushptr(ptr, type) save_pushptr(ptr, type)
#define save_pushptrptr(p1, p2, type) \
    goperl_save_pushptrptr(_gof, (const void *)(p1), (const void *)(p2), (int)(type))
#define Perl_save_pushptrptr(p1, p2, type) save_pushptrptr(p1, p2, type)
#define save_pushi32ptr(i, ptr, type) \
    goperl_save_pushi32ptr(_gof, (I32)(i), (const void *)(ptr), (int)(type))
#define Perl_save_pushi32ptr(i, ptr, type) save_pushi32ptr(i, ptr, type)
#define savestack_grow() SSCHECK(128)
#define Perl_savestack_grow() savestack_grow()
#define savestack_grow_cnt(n) SSCHECK(n)
#define Perl_savestack_grow_cnt(n) savestack_grow_cnt(n)
/* A string freed when the scope ends. A string of the interpreter's memory
 * goes on the save stack as perl's save_freepv pushes it (an entry taken
 * off is pushed again as it was: Future::AsyncAwait suspends and resumes a
 * sub's saves so). A block of the module's allocators, or of the C
 * library, which the interpreter cannot reach, is freed by the module when
 * the scope ends. */
static void goperl_savefreepv_cb(pTHX_ void *p);
static void goperl_save_freepv(goperl_frame_t *f, char *pv) __attribute__((unused));
static void goperl_save_freepv(goperl_frame_t *f, char *pv) {
    if (!pv || goperl_guest_addr(f, pv))
        goperl_save_pushptr(f, pv, SAVEt_FREEPV);
    else
        goperl_save_destructor(f, goperl_savefreepv_cb, (void *)pv, GOPERL_DTOR_CALL_OWN);
}
#define save_freepv(pv) goperl_save_freepv(_gof, (char *)(pv))
/* save_alloc (scope.c): room for size bytes on the interpreter's save stack
 * after pad bytes, in whole entries, under a SAVEt_ALLOC entry; -> where
 * the room begins, in bytes from the bottom of the save stack (Scope::Upper
 * pads the save stack with it so that its destructor lands where it wants) */
static SSize_t goperl_save_alloc(goperl_frame_t *f, SSize_t size, I32 pad) __attribute__((unused));
static SSize_t goperl_save_alloc(goperl_frame_t *f, SSize_t size, I32 pad) {
    const UV elems = 1 + (UV)((size + pad - 1) / (SSize_t)sizeof(ANY));
    const UV elems_shifted = elems << SAVE_TIGHT_SHIFT;
    if ((elems_shifted >> SAVE_TIGHT_SHIFT) != elems)
        goperl_croakf(f, "panic: save_alloc elems %lu out of range (%ld-%ld)", (unsigned long)elems,
                      (long)size, (long)pad);
    ANY *const ss = goperl_ss_room(f, (int)(elems + 1));
    I32 *const ixp = (I32 *)goperl_plint_ref(f, GOPERL_PL_SAVESTACK_IX);
    const SSize_t start = pad + (SSize_t)*ixp * (SSize_t)sizeof(ANY);
    *ixp += (I32)elems;
    ss[*ixp].any_uv = (UV)SAVEt_ALLOC | elems_shifted;
    *ixp += 1;
    return start;
}
#define save_alloc(size, pad) goperl_save_alloc(_gof, (SSize_t)(size), (I32)(pad))
#define Perl_save_alloc(size, pad) save_alloc(size, pad)
/* leave_scope (scope.c): the saves above base are undone, as LEAVE_SCOPE */
#define leave_scope(base) LEAVE_SCOPE(base)
#define Perl_leave_scope(base) leave_scope(base)
/* PL_DBgv, the glob *DB::DB (NULL unless the debugger is in), as it is
 * now (a value: the variables with a slot to store into are those below
 * GOPERL_PLSHADOW_MAX) */
#define GOPERL_PL_DBGV 131
#define PL_DBgv ((GV *)GOPERL_SV(GOPERL_PLVAR(GOPERL_PL_DBGV)))
/* The mask of the ops a compilation refuses (Safe's opmask), MAXO bytes in
 * the interpreter's memory, or NULL: in this process, a pointer into that
 * memory (where the buffer of a string is, SvPVX). */
#define OP_MASK_BUF_SIZE (MAXO + 100)
static __attribute__((unused)) char **goperl_op_mask_ref(goperl_frame_t *f) {
    goperl_plshadow_t *s = &goperl_plshadow_v[GOPERL_PL_OP_MASK];
    if (!s->live || s->val == s->shadow) {
        s->val = (uint64_t)(uintptr_t)goperl_guest_mem(
            f, goperl_xs_op(f, GOPERL_OP_PLVAR_GET, GOPERL_PL_OP_MASK, 0, 0, 0));
        s->shadow = s->val;
        goperl_plshadow_set_live(s);
    }
    return (char **)&s->val;
}
#define PL_op_mask (*goperl_op_mask_ref(_gof))

/* The variables of the parser that perl's own headers leave to ppport.h,
 * which defines them over PL_parser on every perl since 5.9.5; the SDK
 * neutralizes a module's ppport.h (see ppport.h), so they are here, with
 * ppport's check that there is a parser. */
static __attribute__((unused)) yy_parser *goperl_parser_present(goperl_frame_t *f,
                                                                const char *file, int line) {
    yy_parser *p = PL_parser;
    if (!p) goperl_croakf(f, "panic: PL_parser == NULL in %s:%d", file, line);
    return p;
}
#define GOPERL_PARSER_VAR(var) (goperl_parser_present(_gof, __FILE__, __LINE__)->var)
#ifndef PL_expect
#define PL_expect GOPERL_PARSER_VAR(expect)
#endif
#ifndef PL_copline
#define PL_copline GOPERL_PARSER_VAR(copline)
#endif
#ifndef PL_rsfp
#define PL_rsfp GOPERL_PARSER_VAR(rsfp)
#endif
#ifndef PL_rsfp_filters
#define PL_rsfp_filters GOPERL_PARSER_VAR(rsfp_filters)
#endif
#ifndef PL_linestr
#define PL_linestr GOPERL_PARSER_VAR(linestr)
#endif
#ifndef PL_bufptr
#define PL_bufptr GOPERL_PARSER_VAR(bufptr)
#endif
#ifndef PL_bufend
#define PL_bufend GOPERL_PARSER_VAR(bufend)
#endif
#ifndef PL_lex_state
#define PL_lex_state GOPERL_PARSER_VAR(lex_state)
#endif
#ifndef PL_lex_stuff
#define PL_lex_stuff GOPERL_PARSER_VAR(lex_stuff)
#endif
#ifndef PL_tokenbuf
#define PL_tokenbuf GOPERL_PARSER_VAR(tokenbuf)
#endif
#ifndef PL_in_my
#define PL_in_my GOPERL_PARSER_VAR(in_my)
#endif
#ifndef PL_in_my_stash
#define PL_in_my_stash GOPERL_PARSER_VAR(in_my_stash)
#endif
#ifndef PL_error_count
#define PL_error_count GOPERL_PARSER_VAR(error_count)
#endif
/* the others the parser has, read through the copy (what a module writes
 * into them does not go back) */
#ifndef PL_oldbufptr
#define PL_oldbufptr GOPERL_PARSER_VAR(oldbufptr)
#endif
#ifndef PL_oldoldbufptr
#define PL_oldoldbufptr GOPERL_PARSER_VAR(oldoldbufptr)
#endif
#ifndef PL_linestart
#define PL_linestart GOPERL_PARSER_VAR(linestart)
#endif
#ifndef PL_last_uni
#define PL_last_uni GOPERL_PARSER_VAR(last_uni)
#endif
#ifndef PL_last_lop
#define PL_last_lop GOPERL_PARSER_VAR(last_lop)
#endif
#ifndef PL_last_lop_op
#define PL_last_lop_op GOPERL_PARSER_VAR(last_lop_op)
#endif
#ifndef PL_lex_inwhat
#define PL_lex_inwhat GOPERL_PARSER_VAR(lex_inwhat)
#endif
#ifndef PL_lex_brackets
#define PL_lex_brackets GOPERL_PARSER_VAR(lex_brackets)
#endif
#ifndef PL_lex_formbrack
#define PL_lex_formbrack GOPERL_PARSER_VAR(lex_formbrack)
#endif
#ifndef PL_lex_repl
#define PL_lex_repl GOPERL_PARSER_VAR(lex_repl)
#endif
#ifndef PL_multi_start
#define PL_multi_start GOPERL_PARSER_VAR(multi_start)
#endif
#ifndef PL_multi_end
#define PL_multi_end GOPERL_PARSER_VAR(multi_end)
#endif
#ifndef PL_multi_open
#define PL_multi_open GOPERL_PARSER_VAR(multi_open)
#endif
#ifndef PL_multi_close
#define PL_multi_close GOPERL_PARSER_VAR(multi_close)
#endif
#ifndef PL_preambling
#define PL_preambling GOPERL_PARSER_VAR(preambling)
#endif
/* ppport.h's name for a variable of the parser, with its check that there
 * is one, which a module that defines a PL_ name of its own uses
 * (signatures: PL_oldbufptr) */
#ifndef D_PPP_my_PL_parser_var
#define D_PPP_my_PL_parser_var(var) GOPERL_PARSER_VAR(var)
#endif

/* What is pending is delivered (the handler of a signal), as between two
 * ops: despatch_signals, which PERL_ASYNC_CHECK runs. */
#define despatch_signals() goperl_async(_gof)
#define Perl_despatch_signals(...) goperl_async(_gof)

/* save_re_context: the match variables of the code calling native code are
 * kept from the matches of the code it calls (CALL 22). */
#define save_re_context() ((void)goperl_do_op3(_gof, GOPERL_OP_CALL, 0, 0, (uint64_t)(22), 0, 0))

/* sv_peek: perl's short description of a value (dump.c), the text of a
 * mortal of the interpreter (CALL 23). */
static __attribute__((unused)) char *goperl_sv_peek(goperl_frame_t *f, SV *sv) {
    return (char *)goperl_guest_mem(
        f, goperl_do_op3(f, GOPERL_OP_CALL, GOPERL_TOK(sv), 0, (uint64_t)(23), 0, 0));
}
#define sv_peek(sv) goperl_sv_peek(_gof, (SV *)(sv))

/* sv_recode_to_utf8: a string of bytes decoded by an encoding object into
 * characters (its decode method), in place (CALL 24). */
static __attribute__((unused)) char *goperl_sv_recode_to_utf8(goperl_frame_t *f, SV *sv, SV *encoding) {
    return (char *)goperl_guest_mem(
        f, goperl_do_op3(f, GOPERL_OP_CALL, GOPERL_TOK(sv), GOPERL_TOK(encoding), (uint64_t)(24), 0, 0));
}
#define sv_recode_to_utf8(sv, encoding) goperl_sv_recode_to_utf8(_gof, (SV *)(sv), (SV *)(encoding))
#define Perl_sv_peek(...) GOPERL_LONG_CALL(sv_peek, GOPERL_THX_DROP(__VA_ARGS__))

/* CALLREGEXEC: the engine's exec, regexec_flags with the flags as given
 * (REXEC_COPY_STR, REXEC_IGNOREPOS, ...) and its data (RX EXEC_FLAGS). */
#define GOPERL_RX_EXEC_FLAGS 6
#ifndef REXEC_COPY_STR
#define REXEC_COPY_STR 0x01
#define REXEC_CHECKED 0x02
#define REXEC_SCREAM 0x04
#define REXEC_IGNOREPOS 0x08
#define REXEC_NOT_FIRST 0x10
#define REXEC_COPY_SKIP_PRE 0x20
#define REXEC_COPY_SKIP_POST 0x40
#define REXEC_FAIL_ON_UNDERFLOW 0x80
#endif
static __attribute__((unused)) I32 goperl_regexec_flags(goperl_frame_t *f, REGEXP *rx, char *stringarg,
                                                        char *strend, char *strbeg, SSize_t minend,
                                                        SV *screamer, void *data, U32 flags) {
    goperl_gword_t w[7];
    const uintptr_t b = (uintptr_t)strbeg - (uintptr_t)f->membase;
    const uintptr_t e = (uintptr_t)strend - (uintptr_t)f->membase;
    if (data) goperl_croakf(f, "CALLREGEXEC: the data argument of the engine is not available to native XS modules");
    if (b < *f->memsize && e < *f->memsize) {
        w[0] = (goperl_gword_t)((uintptr_t)stringarg - (uintptr_t)f->membase);
        w[1] = (goperl_gword_t)e;
        w[2] = (goperl_gword_t)b;
    } else {
        SV *copy = sv_2mortal(newSVpvn(strbeg, (STRLEN)(strend - strbeg)));
        if (screamer && SvUTF8(screamer)) SvUTF8_on(copy);
        const uint64_t base = goperl_do_op(f, GOPERL_OP_SV_PVX_RAW, GOPERL_TOK(copy), 0, 0, 0);
        w[2] = base;
        w[0] = base + (uint64_t)(stringarg - strbeg);
        w[1] = base + (uint64_t)(strend - strbeg);
        screamer = copy;
    }
    w[3] = (goperl_gword_t)(int64_t)minend;
    w[4] = GOPERL_TOK(screamer);
    w[5] = flags;
    w[6] = 0;
    const I32 r = (I32)(int64_t)goperl_do_op3(f, GOPERL_OP_REGEXP, GOPERL_TOK(rx), sizeof w, (uint64_t)(GOPERL_RX_EXEC_FLAGS), (const char *)w, sizeof w);
    /* the body code holds says what this match found */
    if (goperl_rxbody_find(GOPERL_TOK(rx))) (void)goperl_reany(f, rx);
    return r;
}
#define CALLREGEXEC(rx, stringarg, strend, strbeg, minend, sv, data, flags)          \
    goperl_regexec_flags(_gof, (REGEXP *)(rx), (char *)(stringarg), (char *)(strend), \
                         (char *)(strbeg), (SSize_t)(minend), (SV *)(sv), (void *)(data), \
                         (U32)(flags))
#define regexec_flags(rx, stringarg, strend, strbeg, minend, sv, data, flags) \
    CALLREGEXEC(rx, stringarg, strend, strbeg, minend, sv, data, flags)

#include "goperl_compiler.h"

#endif /* GOPERL_XS_SDK_PERLAPI_H */
