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

/* A function the SDK does not provide must stop the build: with an implicit
 * declaration the module would compile, link (an XS module is linked with
 * undefined symbols allowed) and fail only when it is loaded. */
#pragma GCC diagnostic error "-Wimplicit-function-declaration"

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
#define newGVgen_flags(pack, flags)                                         \
    ({ const char *const goperl_newGVgen_pack_ = (pack);                    \
       (GV *)GOPERL_SV(goperl_ops(GOPERL_OP_CALL, (uint64_t)18 << 32,       \
                                  (uint64_t)(uint32_t)(flags),              \
                                  goperl_newGVgen_pack_,                    \
                                  strlen(goperl_newGVgen_pack_))); })
#define newGVgen(pack) newGVgen_flags((pack), 0)
/* The random numbers of the interpreter (CALL 19): Drand01 is the next one
 * of the generator rand() uses, seedDrand01 seeds it, seed() is what srand
 * without an argument seeds it with, and PL_srand_called says whether it
 * has been seeded. */
typedef U32 Rand_seed_t;
static NV goperl_drand01(goperl_frame_t *f) __attribute__((unused));
static NV goperl_drand01(goperl_frame_t *f) {
    const uint64_t bits = goperl_do_op(f, GOPERL_OP_CALL, (uint64_t)19 << 32, 0, 0, 0);
    double d;
    memcpy(&d, &bits, sizeof d);
    return (NV)d;
}
#define Drand01() goperl_drand01(_gof)
#define seedDrand01(x) \
    ((void)goperl_op0(GOPERL_OP_CALL, (uint64_t)19 << 32, ((uint64_t)(Rand_seed_t)(x) << 2) | 1))
#define seed() ((UV)goperl_op0(GOPERL_OP_CALL, (uint64_t)19 << 32, 2))
#define Perl_seed(ctx) seed()
#define GOPERL_PL_SRAND_CALLED 78
#define GOPERL_PL_SV_COUNT 79
#define GOPERL_PL_SIG_PENDING 80
#define PL_srand_called (*goperl_plint_ref(_gof, GOPERL_PL_SRAND_CALLED))
/* How many SVs the interpreter has, and whether signals wait to be
 * delivered: read only. */
#define PL_sv_count ((IV)GOPERL_PLVAR(GOPERL_PL_SV_COUNT))
#define PL_sig_pending ((int)GOPERL_PLVAR(GOPERL_PL_SIG_PENDING))

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
    ((I32)(int64_t)GOPERL_SVCALL(60, a, GOPERL_SV_PAIR(flags, b)))
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
__attribute__((weak, visibility("hidden"))) void *(Perl_malloc)(size_t n) { return malloc(n); }
__attribute__((weak, visibility("hidden"))) void *(Perl_calloc)(size_t n, size_t size) {
    return calloc(n, size);
}
__attribute__((weak, visibility("hidden"))) void *(Perl_realloc)(void *p, size_t n) {
    return realloc(p, n);
}
__attribute__((weak, visibility("hidden"))) void (Perl_mfree)(void *p) { free(p); }
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
/* Stores a pointer in the buffer slot of an SV: one into the interpreter's
 * memory as its address there, any other with the loader. */
static void goperl_sv_pvx_store(goperl_frame_t *f, uint64_t tok, const void *p) {
    uint64_t g = goperl_guest_addr(f, p);
    if (p && !g) {
        void **kept = f->api->pv_host(f, tok, 1);
        if (!kept) goperl_croakf(f, "out of memory for the buffer slot of an SV");
        kept[0] = (void *)p;
        kept[1] = 0;
        g = GOPERL_PV_HOST;
    }
    goperl_xs_op(f, GOPERL_OP_SV_SLOT_SET,
                 ((uint64_t)5 /* GOPERL_SVSLOT_PV */ << 32) | (uint32_t)tok, g,
                 0, 0);
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
static void goperl_sv_len_store(goperl_frame_t *f, uint64_t tok, uint64_t len) {
    void **kept = goperl_pv_host_of(f, tok);
    if (kept) {
        kept[1] = (void *)(uintptr_t)len;
        return;
    }
    goperl_xs_op(f, GOPERL_OP_SV_SLOT_SET,
                 ((uint64_t)6 /* GOPERL_SVSLOT_LEN */ << 32) | (uint32_t)tok, len,
                 0, 0);
}
static uint64_t goperl_cvfield_load(goperl_frame_t *f, uint64_t tok, int field);
static void goperl_cvfield_store(goperl_frame_t *f, uint64_t tok, int field, uint64_t val);

/* The entry a hash's iterator is at (HvEITER) is kept by modules beyond the
 * call that read it, to be put back later (Hash::SafeKeys,
 * Hash::StoredIterator); the shadows of other calls live in the call's
 * arena, released when it returns. Each entry of an instance therefore has
 * one shadow that lasts, here, with its value as it was last read. */
typedef struct {
    void *inst;
    uint64_t tok;
    HE *he;
} goperl_eiter_t;
__attribute__((weak, visibility("hidden"))) goperl_eiter_t *goperl_eiter_v = 0;
__attribute__((weak, visibility("hidden"))) uint32_t goperl_eiter_cap_v = 0;
__attribute__((weak, visibility("hidden"))) uint32_t goperl_eiter_n_v = 0;
__attribute__((weak, visibility("hidden"))) pthread_mutex_t goperl_eiter_lock_v =
    PTHREAD_MUTEX_INITIALIZER;
/* The lasting shadow of the entry, made when there is none (NULL when
 * memory ran out); several instances may run native code at once. */
static HE *goperl_eiter_find(void *inst, uint64_t tok) __attribute__((unused));
static HE *goperl_eiter_find(void *inst, uint64_t tok) {
    HE *he = 0;
    pthread_mutex_lock(&goperl_eiter_lock_v);
    if (2 * (goperl_eiter_n_v + 1) > goperl_eiter_cap_v) {
        uint32_t ncap = goperl_eiter_cap_v ? 2 * goperl_eiter_cap_v : 64;
        goperl_eiter_t *n = (goperl_eiter_t *)calloc(ncap, sizeof(*n));
        if (!n) goto out;
        for (uint32_t i = 0; i < goperl_eiter_cap_v; i++) {
            goperl_eiter_t *e = &goperl_eiter_v[i];
            if (!e->he) continue;
            uint32_t h = (uint32_t)((e->tok ^ (uintptr_t)e->inst) * 2654435761u) & (ncap - 1);
            while (n[h].he) h = (h + 1) & (ncap - 1);
            n[h] = *e;
        }
        free(goperl_eiter_v);
        goperl_eiter_v = n;
        goperl_eiter_cap_v = ncap;
    }
    {
        uint32_t h = (uint32_t)((tok ^ (uintptr_t)inst) * 2654435761u) & (goperl_eiter_cap_v - 1);
        while (goperl_eiter_v[h].he &&
               (goperl_eiter_v[h].tok != tok || goperl_eiter_v[h].inst != inst))
            h = (h + 1) & (goperl_eiter_cap_v - 1);
        goperl_eiter_t *e = &goperl_eiter_v[h];
        if (!e->he && (e->he = (HE *)calloc(1, sizeof(HE))) != 0) {
            e->inst = inst;
            e->tok = tok;
            e->he->goperl_tok = tok;
            goperl_eiter_n_v++;
        }
        he = e->he;
    }
out:
    pthread_mutex_unlock(&goperl_eiter_lock_v);
    return he;
}
static HE *goperl_eiter_shadow(goperl_frame_t *f, uint64_t hvtok, uint64_t tok)
    __attribute__((unused));
static HE *goperl_eiter_shadow(goperl_frame_t *f, uint64_t hvtok, uint64_t tok) {
    if (!tok) return 0;
    HE *he = goperl_eiter_find(f->inst, tok);
    if (!he) goperl_croakf(f, "%s", "Out of memory!");
    /* the key and the value now (the address may be another entry's by
     * now); the key is the shadow's own, as the shadow is */
    HEK *old = he->hent_hek;
    he->hent_hek = 0;
    free(old);
    const HEK *k = goperl_he_hek(f, he);
    const size_t ksize = sizeof(HEK) + (size_t)k->hek_len + 2;
    HEK *own = (HEK *)malloc(ksize);
    if (!own) goperl_croakf(f, "%s", "Out of memory!");
    memcpy(own, k, ksize);
    he->hent_hek = own;
    he->he_valu.hent_val = GOPERL_SV(goperl_value_out(
        f, goperl_do_op(f, GOPERL_OP_HV_ITERVAL, hvtok, tok, 0, 0)));
    return he;
}

static uint64_t goperl_lv_load_field(goperl_frame_t *f, uint64_t tok, int field) {
    switch (field) {
    case GOPERL_LV_IV:
        return goperl_do_op(f, GOPERL_OP_SV_SLOT_GET,
                            ((uint64_t)2 << 32) | (uint32_t)tok, 0, 0, 0);
    case GOPERL_LV_UV:
        return goperl_do_op(f, GOPERL_OP_SV_SLOT_GET,
                            ((uint64_t)3 << 32) | (uint32_t)tok, 0, 0, 0);
    case GOPERL_LV_NV:
        return goperl_do_op(f, GOPERL_OP_SV_SLOT_GET,
                            ((uint64_t)4 << 32) | (uint32_t)tok, 0, 0, 0);
    case GOPERL_LV_RV:
        /* the slot, as perl's SvRV reads it whatever the flags say (not
         * "the target if SvROK", which is SvROK's question) */
        return goperl_do_op(f, GOPERL_OP_SV_SLOT_GET,
                            ((uint64_t)0 /* GOPERL_SVSLOT_RV */ << 32) | (uint32_t)tok,
                            0, 0, 0);
    case GOPERL_LV_CVXSUB:
        return (uint64_t)(uintptr_t)goperl_api_cv_xsub(f, tok);
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
        return goperl_do_op(f, GOPERL_OP_FIELD, ((uint64_t)6 << 32) | (uint32_t)tok, 0, 0, 0);
    case GOPERL_LV_LVTYPE: /* FIELD 9 */
        return goperl_do_op(f, GOPERL_OP_FIELD, ((uint64_t)9 << 32) | (uint32_t)tok, 0, 0, 0);
    case GOPERL_LV_LVTARG: /* FIELD 10 */
        return goperl_do_op(f, GOPERL_OP_FIELD, ((uint64_t)10 << 32) | (uint32_t)tok, 0, 0, 0);
    case GOPERL_LV_CVOUTSIDE:
        return goperl_do_op(f, GOPERL_OP_GV_CALL,
                            ((uint64_t)GOPERL_GVCALL_CV_OUTSIDE << 32) | (uint32_t)tok,
                            0, 0, 0);
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
        return goperl_do_op(f, GOPERL_OP_GV_CALL,
                            ((uint64_t)GOPERL_GVCALL_FLAGS << 32) | (uint32_t)tok, 0, 0, 0);
    case GOPERL_LV_HVRITER: /* FIELD 2 */
        return goperl_do_op(f, GOPERL_OP_FIELD, ((uint64_t)2 << 32) | (uint32_t)tok, 0, 0, 0);
    case GOPERL_LV_IOIFP:
        return goperl_do_op(f, GOPERL_OP_PERLIO,
                            ((uint64_t)GOPERL_PIO_IFP << 32) | (uint32_t)tok, 0, 0, 0);
    case GOPERL_LV_IOOFP:
        return goperl_do_op(f, GOPERL_OP_PERLIO,
                            ((uint64_t)GOPERL_PIO_OFP << 32) | (uint32_t)tok, 0, 0, 0);
    case GOPERL_LV_IOTYPE:
        return (uint64_t)(unsigned char)goperl_do_op(f, GOPERL_OP_SV_CALL, ((uint64_t)GOPERL_SVCALL_IO_TYPE << 32) | (uint32_t)tok, 0, 0, 0);
    case GOPERL_LV_IOFLAGS:
        return (uint64_t)(U8)goperl_do_op(f, GOPERL_OP_SV_CALL, ((uint64_t)GOPERL_SVCALL_IO_FLAGS << 32) | (uint32_t)tok, 0, 0, 0);
    case GOPERL_LV_GVIO:
        return goperl_do_op(f, GOPERL_OP_GV_PTR, tok, 6, 0, 0);
    case GOPERL_LV_HVEITER: /* FIELD 4 */
        return (uint64_t)(uintptr_t)goperl_eiter_shadow(
            f, tok, goperl_do_op(f, GOPERL_OP_FIELD, ((uint64_t)4 << 32) | (uint32_t)tok, 0, 0, 0));
    default:
        return goperl_do_op(f, GOPERL_OP_GV_CALL,
                            ((uint64_t)GOPERL_GVCALL_CVGEN << 32) | (uint32_t)tok,
                            0, 0, 0);
    }
}
static void goperl_lv_store_field(goperl_frame_t *f, uint64_t tok, int field,
                                  uint64_t val) {
    switch (field) {
    case GOPERL_LV_IV:
        goperl_xs_op(f, GOPERL_OP_SV_SLOT_SET,
                     ((uint64_t)2 << 32) | (uint32_t)tok, val, 0, 0);
        break;
    case GOPERL_LV_UV:
        goperl_xs_op(f, GOPERL_OP_SV_SLOT_SET,
                     ((uint64_t)3 << 32) | (uint32_t)tok, val, 0, 0);
        break;
    case GOPERL_LV_NV:
        goperl_xs_op(f, GOPERL_OP_SV_SLOT_SET,
                     ((uint64_t)4 << 32) | (uint32_t)tok, val, 0, 0);
        break;
    case GOPERL_LV_RV:
        goperl_xs_op(f, GOPERL_OP_SV_SLOT_SET, (uint32_t)tok, (uint32_t)val, 0,
                     0);
        break;
    case GOPERL_LV_CVXSUB:
        goperl_api_cv_set_xsub(f, tok, (void *)(uintptr_t)val);
        break;
    case GOPERL_LV_GVSTASH:
        goperl_xs_op(f, GOPERL_OP_GV_CALL,
                     ((uint64_t)GOPERL_GVCALL_STASH_SET << 32) | (uint32_t)tok,
                     (uint32_t)val, 0, 0);
        break;
    case GOPERL_LV_GVHV:
        goperl_xs_op(f, GOPERL_OP_GV_CALL,
                     ((uint64_t)GOPERL_GVCALL_HV_SET << 32) | (uint32_t)tok,
                     (uint32_t)val, 0, 0);
        break;
    case GOPERL_LV_GVAV:
        goperl_xs_op(f, GOPERL_OP_GV_CALL,
                     ((uint64_t)GOPERL_GVCALL_AV_SET << 32) | (uint32_t)tok,
                     (uint32_t)val, 0, 0);
        break;
    case GOPERL_LV_GVSV:
        goperl_xs_op(f, GOPERL_OP_GV_CALL,
                     ((uint64_t)GOPERL_GVCALL_SV_SET << 32) | (uint32_t)tok,
                     (uint32_t)val, 0, 0);
        break;
    case GOPERL_LV_PVX:
        goperl_sv_pvx_store(f, tok, (const void *)(uintptr_t)val);
        break;
    case GOPERL_LV_LEN:
        goperl_sv_len_store(f, tok, val);
        break;
    case GOPERL_LV_AVFILLP: /* FIELD 7 */
        goperl_xs_op(f, GOPERL_OP_FIELD, ((uint64_t)7 << 32) | (uint32_t)tok, val, 0, 0);
        break;
    case GOPERL_LV_LVTYPE: /* FIELD 18 */
        goperl_xs_op(f, GOPERL_OP_FIELD, ((uint64_t)18 << 32) | (uint32_t)tok,
                     (uint64_t)(unsigned char)val, 0, 0);
        break;
    case GOPERL_LV_LVTARG: /* FIELD 19 */
        goperl_xs_op(f, GOPERL_OP_FIELD, ((uint64_t)19 << 32) | (uint32_t)tok,
                     (uint32_t)val, 0, 0);
        break;
    case GOPERL_LV_REFCNT: /* FIELD 0: the count itself, nothing is freed */
        goperl_xs_op(f, GOPERL_OP_FIELD, (uint32_t)tok, (uint32_t)val, 0, 0);
        break;
    case GOPERL_LV_CVOUTSIDE: /* FIELD 16: the plain store */
        goperl_xs_op(f, GOPERL_OP_FIELD, ((uint64_t)16 << 32) | (uint32_t)tok,
                     (uint32_t)val, 0, 0);
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
        goperl_xs_op(f, GOPERL_OP_SV_SLOT_SET,
                     ((uint64_t)GOPERL_SVSLOT_STASH << 32) | (uint32_t)tok,
                     (uint32_t)goperl_value_in(f, val), 0, 0);
        break;
    case GOPERL_LV_GVFLAGS:
        goperl_xs_op(f, GOPERL_OP_GV_CALL,
                     ((uint64_t)GOPERL_GVCALL_FLAGS << 32) | (uint32_t)tok,
                     ((uint64_t)0xffffffffu << 32) | (uint32_t)val, 0, 0);
        break;
    case GOPERL_LV_HVRITER: /* FIELD 3 */
        goperl_xs_op(f, GOPERL_OP_FIELD, ((uint64_t)3 << 32) | (uint32_t)tok,
                     (uint64_t)(int64_t)(int32_t)(uint32_t)val, 0, 0);
        break;
    case GOPERL_LV_HVEITER: /* FIELD 5: the entry the shadow stands for */
        goperl_xs_op(f, GOPERL_OP_FIELD, ((uint64_t)5 << 32) | (uint32_t)tok,
                     val ? ((HE *)(uintptr_t)val)->goperl_tok : 0, 0, 0);
        break;
    case GOPERL_LV_IOIFP: /* FIELD 20 */
    case GOPERL_LV_IOOFP: /* FIELD 21 */
        goperl_xs_op(f, GOPERL_OP_FIELD,
                     ((uint64_t)(field == GOPERL_LV_IOIFP ? 20 : 21) << 32) | (uint32_t)tok,
                     goperl_pio_tok(f, (const void *)(uintptr_t)val), 0, 0);
        break;
    case GOPERL_LV_IOTYPE: /* FIELD 22 */
        goperl_xs_op(f, GOPERL_OP_FIELD, ((uint64_t)22 << 32) | (uint32_t)tok,
                     (uint64_t)(unsigned char)val, 0, 0);
        break;
    case GOPERL_LV_IOFLAGS: /* FIELD 23 */
        goperl_xs_op(f, GOPERL_OP_FIELD, ((uint64_t)23 << 32) | (uint32_t)tok,
                     (uint64_t)(U8)val, 0, 0);
        break;
    case GOPERL_LV_GVIO:
        goperl_xs_op(f, GOPERL_OP_GV_CALL,
                     ((uint64_t)GOPERL_GVCALL_IO_SET << 32) | (uint32_t)tok,
                     (uint32_t)val, 0, 0);
        break;
    default:
        goperl_xs_op(f, GOPERL_OP_GV_CALL,
                     ((uint64_t)GOPERL_GVCALL_CVGEN_SET << 32) | (uint32_t)tok,
                     (uint32_t)val, 0, 0);
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
    GOPERL_SVSTR_GV_NAME_SET = 17
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
    goperl_do_op(_gof, GOPERL_OP_SV_CALL,                                  \
                 ((uint64_t)(fn) << 32) | (uint32_t)GOPERL_TOK(sv),        \
                 (uint64_t)(b), 0, 0)
#define GOPERL_GVCALL(fn, gv, b)                                           \
    goperl_do_op(_gof, GOPERL_OP_GV_CALL,                                  \
                 ((uint64_t)(fn) << 32) | (uint32_t)GOPERL_TOK(gv),        \
                 (uint64_t)(b), 0, 0)
#define GOPERL_SVSTR(fn, sv, flags, s, len)                                \
    ({ const uint64_t goperl_SVSTR_len_ = (uint64_t)(len);                 \
       goperl_do_op(_gof, GOPERL_OP_SV_STR,                                \
                    ((uint64_t)(fn) << 32) | (uint32_t)GOPERL_TOK(sv),     \
                    ((uint64_t)(U32)(flags) << 32) | (uint32_t)goperl_SVSTR_len_, (s), \
                    goperl_SVSTR_len_); })
/* b of the two-SV calls: flags<<32 | the other SV. */
#define GOPERL_SV_PAIR(flags, sv) \
    (((uint64_t)(U32)(flags) << 32) | (uint32_t)GOPERL_TOK(sv))
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
        goperl_do_op(f, GOPERL_OP_SV_CALL,
                     ((uint64_t)GOPERL_SVCALL_BACKOFF << 32) | (uint32_t)GOPERL_TOK(sv),
                     0, 0, 0);
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
#define SvIsCOW_shared_hash(sv) \
    ({ GOPERL_ONCE(goperl_SvIsCOW_shared_hash_sv_, sv); \
       (SvFLAGS(goperl_SvIsCOW_shared_hash_sv_) & (SVf_POK | SVf_IsCOW)) == (SVf_POK | SVf_IsCOW) && SvLEN(goperl_SvIsCOW_shared_hash_sv_) == 0; })
#define SvPROTECT(sv) (SvFLAGS(sv) & SVf_PROTECT)
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
#define mg_magical(sv) ((void)GOPERL_SVCALL(GOPERL_SVCALL_MG_MAGICAL, sv, 0))
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
    return (IV)(int64_t)goperl_do_op(
        f, GOPERL_OP_SV_CALL,
        ((uint64_t)GOPERL_SVCALL_2IV << 32) | (uint32_t)GOPERL_TOK(sv), flags, 0, 0);
}
static UV goperl_sv_uv_flags(goperl_frame_t *f, SV *sv, U32 flags) {
    return (UV)goperl_do_op(
        f, GOPERL_OP_SV_CALL,
        ((uint64_t)GOPERL_SVCALL_2UV << 32) | (uint32_t)GOPERL_TOK(sv), flags, 0, 0);
}
static NV goperl_sv_nv_flags(goperl_frame_t *f, SV *sv, U32 flags) {
    union {
        uint64_t u;
        double d;
    } cvt;
    cvt.u = goperl_do_op(
        f, GOPERL_OP_SV_CALL,
        ((uint64_t)GOPERL_SVCALL_2NV << 32) | (uint32_t)GOPERL_TOK(sv), flags, 0, 0);
    return cvt.d;
}
static char *goperl_sv_pv_flags(goperl_frame_t *f, SV *sv, STRLEN *lenp,
                                U32 flags) {
    uint64_t r = goperl_do_op(
        f, GOPERL_OP_SV_CALL,
        ((uint64_t)GOPERL_SVCALL_2PV << 32) | (uint32_t)GOPERL_TOK(sv), flags, 0, 0);
    if (lenp) *lenp = (STRLEN)(uint32_t)r;
    if (!(r >> 32)) return 0;
    return (char *)goperl_guest_mem(f, r >> 32);
}
#define SvIV_nomg(sv) goperl_sv_iv_flags(_gof, (SV *)(sv), 0)
#define SvUV_nomg(sv) goperl_sv_uv_flags(_gof, (SV *)(sv), 0)
#define SvNV_nomg(sv) goperl_sv_nv_flags(_gof, (SV *)(sv), 0)
#define SvTRUE_nomg(sv) (GOPERL_SVCALL(GOPERL_SVCALL_2BOOL, sv, 0) != 0)
#define SvTRUE_NN(sv) SvTRUE(sv)
#define SvTRUE_nomg_NN(sv) SvTRUE_nomg(sv)
#define SvTRUEx(sv) SvTRUE(sv)
#define SvPV_flags(sv, lenvar, flags) \
    goperl_sv_pv_flags(_gof, (SV *)(sv), (STRLEN *)&(lenvar), (U32)(flags))
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
    ((void)goperl_do_op(_gof, GOPERL_OP_SV_SLOT_SET,                       \
                        ((uint64_t)(slot) << 32) | (uint32_t)GOPERL_TOK(sv), \
                        (uint64_t)(v), 0, 0))
#define SvRV_set(sv, rv) GOPERL_SLOT_SET(GOPERL_SVSLOT_RV, sv, (uint32_t)GOPERL_TOK(rv))
#define SvSTASH_set(sv, stash) \
    GOPERL_SLOT_SET(GOPERL_SVSLOT_STASH, sv, (uint32_t)GOPERL_TOK(stash))
#define SvIV_set(sv, v) GOPERL_SLOT_SET(GOPERL_SVSLOT_IV, sv, (int64_t)(v))
#define SvUV_set(sv, v) GOPERL_SLOT_SET(GOPERL_SVSLOT_UV, sv, (UV)(v))
static void goperl_sv_nv_set(goperl_frame_t *f, SV *sv, NV v) {
    union {
        uint64_t u;
        double d;
    } cvt;
    cvt.d = v;
    goperl_do_op(f, GOPERL_OP_SV_SLOT_SET,
                 ((uint64_t)GOPERL_SVSLOT_NV << 32) | (uint32_t)GOPERL_TOK(sv),
                 cvt.u, 0, 0);
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
#define SvPVCLEAR(sv) sv_setpvs((sv), "")
#define Sv_Grow(sv, n) sv_grow((sv), (n))
#define SvGROW_mutable(sv, n) SvGROW((sv), (n))
#define sv_set_undef(sv) ((void)GOPERL_SVCALL(GOPERL_SVCALL_SET_UNDEF, sv, 0))
static char *goperl_sv_setpv_bufsize(goperl_frame_t *f, SV *sv, STRLEN cur,
                                     STRLEN len) {
    uint64_t gp = goperl_do_op(
        f, GOPERL_OP_SV_CALL,
        ((uint64_t)GOPERL_SVCALL_SETPV_BUFSIZE << 32) | (uint32_t)GOPERL_TOK(sv),
        ((uint64_t)(uint32_t)cur << 32) | (uint32_t)len, 0, 0);
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
    ((void)GOPERL_SVCALL(GOPERL_SVCALL_SETSV, d, GOPERL_SV_PAIR(flags, s)))
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
    ((void)GOPERL_SVCALL(GOPERL_SVCALL_CATSV, d, GOPERL_SV_PAIR(flags, s)))
#define sv_catsv_nomg(d, s) sv_catsv_flags(d, s, 0)
#define sv_catsv_mg(d, s) sv_catsv_flags(d, s, SV_GMAGIC | SV_SMAGIC)
#define sv_catpvn_flags(d, p, len, flags) \
    ((void)GOPERL_SVSTR(GOPERL_SVSTR_CATPVN, d, flags, (p), (len)))
#define sv_catpvn_nomg(d, p, len) sv_catpvn_flags(d, p, len, 0)
#define sv_catpvn_nomg_maybeutf8(d, p, len, is_utf8) \
    sv_catpvn_flags(d, p, len, (is_utf8) ? SV_CATUTF8 : SV_CATBYTES)
#define sv_catpv_flags(d, p, flags) \
    ({ GOPERL_ONCE(goperl_sv_catpv_flags_p_, p); \
       sv_catpvn_flags(d, goperl_sv_catpv_flags_p_, strlen(goperl_sv_catpv_flags_p_), flags); })
#define sv_catpv_nomg(d, p) \
    ({ GOPERL_ONCE(goperl_sv_catpv_nomg_p_, p); \
       sv_catpvn_flags(d, goperl_sv_catpv_nomg_p_, strlen(goperl_sv_catpv_nomg_p_), 0); })
#define sv_catpv_mg(d, p) \
    ({ GOPERL_ONCE(goperl_sv_catpv_mg_p_, p); \
       sv_catpvn_flags(d, goperl_sv_catpv_mg_p_, strlen(goperl_sv_catpv_mg_p_), SV_GMAGIC | SV_SMAGIC); })
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
    ((I32)(int64_t)GOPERL_SVCALL(GOPERL_SVCALL_CMP, a, GOPERL_SV_PAIR(flags, b)))
#define sv_eq_flags(a, b, flags) \
    ((I32)GOPERL_SVCALL(GOPERL_SVCALL_EQ, a, GOPERL_SV_PAIR(flags, b)))
/* The numeric comparison of perl 5.44 (overloading included; SV_GMAGIC
 * runs get magic). */
#define sv_numcmp_flags(a, b, flags) \
    ((I32)(int64_t)GOPERL_SVCALL(62 /* GOPERL_SVCALL_NUMCMP */, a, GOPERL_SV_PAIR(flags, b)))
#define sv_numcmp(a, b) sv_numcmp_flags((a), (b), SV_GMAGIC)
#define sv_dec(sv) ((void)GOPERL_SVCALL(GOPERL_SVCALL_DEC, sv, 0))
#define sv_len_utf8(sv) ((STRLEN)GOPERL_SVCALL(GOPERL_SVCALL_LEN_UTF8, sv, 0))
#define sv_utf8_encode(sv) ((void)GOPERL_SVCALL(GOPERL_SVCALL_UTF8_ENCODE, sv, 0))
#define sv_utf8_upgrade_flags(sv, flags) \
    ((STRLEN)GOPERL_SVCALL(GOPERL_SVCALL_UTF8_UPGRADE, sv, (U32)(flags)))
#define sv_utf8_upgrade_nomg(sv) sv_utf8_upgrade_flags(sv, 0)
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
    SV *r = GOPERL_SV(goperl_do_op(
        f, GOPERL_OP_SV_CALL,
        ((uint64_t)GOPERL_SVCALL_REF << 32) | (uint32_t)GOPERL_TOK(sv),
        (uint64_t)ob, 0, 0));
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
       sv_does_pvn(sv, goperl_sv_does_pv_name_, strlen(goperl_sv_does_pv_name_), flags); })
#define sv_does(sv, name) \
    ({ GOPERL_ONCE(goperl_sv_does_name_, name); \
       sv_does_pvn(sv, goperl_sv_does_name_, strlen(goperl_sv_does_name_), 0); })
#define sv_pos_u2b_flags(sv, uoffset, lenp, flags) \
    goperl_sv_pos_u2b(_gof, (SV *)(sv), (STRLEN)(uoffset), (STRLEN *)(lenp))
static STRLEN goperl_sv_pos_u2b(goperl_frame_t *f, SV *sv, STRLEN uoffset,
                                STRLEN *lenp) __attribute__((unused));
static STRLEN goperl_sv_pos_u2b(goperl_frame_t *f, SV *sv, STRLEN uoffset,
                                STRLEN *lenp) {
    uint64_t a = ((uint64_t)GOPERL_SVCALL_POS_U2B << 32) | (uint32_t)GOPERL_TOK(sv);
    STRLEN boff = (STRLEN)goperl_do_op(f, GOPERL_OP_SV_CALL, a, uoffset, 0, 0);
    if (lenp) {
        STRLEN bend = (STRLEN)goperl_do_op(f, GOPERL_OP_SV_CALL, a,
                                           uoffset + *lenp, 0, 0);
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
    return GOPERL_SV(goperl_do_op(
        f, GOPERL_OP_SV_STR,
        ((uint64_t)GOPERL_SVSTR_NEWSVRV << 32) | (uint32_t)GOPERL_TOK(rv),
        ((uint64_t)(classname ? 0 : 1) << 32) |
            (uint32_t)(classname ? strlen(classname) : 0),
        classname ? classname : "", classname ? strlen(classname) : 0));
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
    ((void)GOPERL_SVCALL(GOPERL_SVCALL_SETRV_INC, sv, (uint32_t)GOPERL_TOK(ref)))
#define sv_setrv_noinc(sv, ref) \
    ((void)GOPERL_SVCALL(GOPERL_SVCALL_SETRV_NOINC, sv, (uint32_t)GOPERL_TOK(ref)))
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
    GOPERL_SV(GOPERL_SVCALL(GOPERL_SVCALL_AV_DELETE, av,                  \
                            ((uint64_t)(U32)(flags) << 32) |              \
                                (uint32_t)(int32_t)(key)))
#define AvMAX(av) ((SSize_t)(int64_t)GOPERL_SVCALL(GOPERL_SVCALL_AV_MAX, av, 0))
/* the caller has flushed */
static int32_t goperl_av_slots(goperl_frame_t *f, uint64_t avtok) {
    return (int32_t)(int64_t)goperl_xs_op(
               f, GOPERL_OP_SV_CALL,
               ((uint64_t)GOPERL_SVCALL_AV_MAX << 32) | (uint32_t)avtok, 0, 0,
               0) +
           1;
}
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
#define IoLINES(io) ((IV)(int64_t)GOPERL_SVCALL(GOPERL_SVCALL_IO_LINES, io, 0))

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
    const uint32_t head[5] = {(uint32_t)as_raw, (uint32_t)rawmode,
                              (uint32_t)rawperm,
                              (uint32_t)(uintptr_t)supplied_fp,
                              (uint32_t)count};
    const size_t size = sizeof head + count * 4 + mode_len;
    char *req = (char *)malloc(size + 1);
    if (!req) goperl_croakf(f, "%s", "Out of memory!");
    memcpy(req, head, sizeof head);
    for (size_t i = 0; i < count; i++) {
        const uint32_t tok = (uint32_t)GOPERL_TOK(svp[i]);
        memcpy(req + sizeof head + i * 4, &tok, 4);
    }
    if (mode_len) memcpy(req + sizeof head + count * 4, oname, mode_len);
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
    const uint64_t tied = goperl_do_op(
        f, GOPERL_OP_SV_CALL,
        ((uint64_t)GOPERL_SVCALL_TIED << 32) | (uint32_t)GOPERL_TOK(sv),
        (uint64_t)(int64_t)how, 0, 0);
    if (!tied) return 0;
    MAGIC *mg = (MAGIC *)goperl_arena_alloc(sizeof(MAGIC));
    memset(mg, 0, sizeof(MAGIC));
    mg->mg_type = (char)how;
    mg->mg_obj = GOPERL_SV((uint32_t)tied);
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
        goperl_text_op(f, &t, GOPERL_OP_SV_STR, (uint32_t)GOPERL_TOK(sv),
                       ((uint64_t)(uint32_t)(SV_GMAGIC | (t.utf8 ? SV_CATUTF8 : SV_CATBYTES)) << 32) | (uint32_t)t.len);
    else
        goperl_text_op(f, &t, GOPERL_OP_SV_STR,
                       ((uint64_t)GOPERL_SVSTR_SETPVN << 32) | (uint32_t)GOPERL_TOK(sv),
                       GOPERL_TEXT_ARG(&t, 0));
}
#define sv_vsetpvfn(sv, pat, patlen, args, svargs, n, tainted) \
    goperl_sv_vsetpvfn(_gof, (SV *)(sv), (pat), (patlen), (args), (svargs), (n), 0)
#define sv_vcatpvfn(sv, pat, patlen, args, svargs, n, tainted) \
    goperl_sv_vsetpvfn(_gof, (SV *)(sv), (pat), (patlen), (args), (svargs), (n), 1)
#define sv_vsetpvf(sv, pat, args) \
    ({ GOPERL_ONCE(goperl_sv_vsetpvf_pat_, pat); \
       goperl_sv_vsetpvfn(_gof, (SV *)(sv), (goperl_sv_vsetpvf_pat_), strlen(goperl_sv_vsetpvf_pat_), (args), 0, 0, 0); })
#define sv_vcatpvf(sv, pat, args) \
    ({ GOPERL_ONCE(goperl_sv_vcatpvf_pat_, pat); \
       goperl_sv_vsetpvfn(_gof, (SV *)(sv), (goperl_sv_vcatpvf_pat_), strlen(goperl_sv_vcatpvf_pat_), (args), 0, 0, 1); })
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
    goperl_text_op(f, &t, GOPERL_OP_SV_STR, (uint64_t)GOPERL_SVSTR_WARN << 32,
                   GOPERL_TEXT_ARG(&t, 0));
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
#define GvCV_set(gv, cv) \
    ((void)GOPERL_GVCALL(GOPERL_GVCALL_CV_SET, gv, (uint32_t)GOPERL_TOK(cv)))
/* An lvalue as on perl (`GvFLAGS(gv) ^= flag`): the whole word is stored. */
#define GvFLAGS(gv) \
    (*(U32 *)goperl_lv_ref(_gof, (SV *)(gv), GOPERL_LV_GVFLAGS, GOPERL_LV_AT))
#define GOPERL_GVFLAGS_CHANGE(gv, set, clear)                             \
    ((void)GOPERL_GVCALL(GOPERL_GVCALL_FLAGS, gv,                         \
                         ((uint64_t)(U32)(clear) << 32) | (U32)(set)))
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
/* GvGP is an identity: modules test it and compare it, never look inside. */
#define GvGP(gv) ((void *)(uintptr_t)GOPERL_GVCALL(GOPERL_GVCALL_GP, gv, 0))
#define GvFILE(gv) \
    ((char *)goperl_intern_packed(_gof, GOPERL_GVCALL(GOPERL_GVCALL_FILE, gv, 0)))
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
    goperl_do_op(f, GOPERL_OP_GV_CALL,
                 ((uint64_t)GOPERL_GVCALL_FULLNAME << 32) | (uint32_t)GOPERL_TOK(gv),
                 ((uint64_t)((keepmain ? 2 : 0) | (effective ? 1 : 0)) << 32) |
                     (uint32_t)GOPERL_TOK(sv),
                 0, 0);
    if (prefix && *prefix) {
        /* the interpreter writes the prefix first; here it goes in front
         * of the name the interpreter produced */
        SV *tmp = GOPERL_SV(goperl_do_op(f, GOPERL_OP_NEW_PVN, 0, strlen(prefix),
                                         prefix, strlen(prefix)));
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
       gv_fetchmeth_pvn(stash, goperl_gv_fetchmeth_pv_name_, strlen(goperl_gv_fetchmeth_pv_name_), level, flags); })
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
       gv_fetchmethod_pvn_flags(stash, goperl_gv_fetchmethod_pv_flags_name_, strlen(goperl_gv_fetchmethod_pv_flags_name_), flags); })
#define gv_fetchmethod_flags(stash, name, flags) \
    ({ GOPERL_ONCE(goperl_gv_fetchmethod_flags_name_, name); \
       gv_fetchmethod_pvn_flags(stash, goperl_gv_fetchmethod_flags_name_, strlen(goperl_gv_fetchmethod_flags_name_), flags); })
#define gv_fetchpvn_flags(name, len, flags, svt) \
    ({ GOPERL_ONCE(goperl_gv_fetchpvn_flags_len_, len); \
       ((GV *)GOPERL_SV(goperl_do_op( _gof, GOPERL_OP_SV_STR, ((uint64_t)GOPERL_SVSTR_GV_FETCHPVN << 32) | (uint32_t)(svt), ((uint64_t)(U32)(flags) << 32) | (uint32_t)(goperl_gv_fetchpvn_flags_len_), (name), (uint64_t)(goperl_gv_fetchpvn_flags_len_)))); })
#define gv_fetchpvn(name, len, flags, svt) gv_fetchpvn_flags(name, len, flags, svt)
#define get_cvn_flags(name, len, flags) \
    ({ GOPERL_ONCE(goperl_get_cvn_flags_len_, len); \
       ((CV *)GOPERL_SV(goperl_do_op( _gof, GOPERL_OP_SV_STR, (uint64_t)GOPERL_SVSTR_GET_CVN << 32, ((uint64_t)(U32)(flags) << 32) | (uint32_t)(goperl_get_cvn_flags_len_), (name), (uint64_t)(goperl_get_cvn_flags_len_)))); })
#define whichsig_pvn(sig, len) \
    ({ GOPERL_ONCE(goperl_whichsig_pvn_len_, len); \
       ((I32)(int64_t)goperl_do_op(_gof, GOPERL_OP_SV_STR, (uint64_t)GOPERL_SVSTR_WHICHSIG << 32, (uint32_t)(goperl_whichsig_pvn_len_), (sig), (uint64_t)(goperl_whichsig_pvn_len_))); })
#define whichsig_pv(sig) \
    ({ GOPERL_ONCE(goperl_whichsig_pv_sig_, sig); \
       whichsig_pvn((goperl_whichsig_pv_sig_), strlen(goperl_whichsig_pv_sig_)); })
#define whichsig(sig) whichsig_pv(sig)

#define CvFLAGS(cv) ((U32)GOPERL_GVCALL(GOPERL_GVCALL_CV_FLAGS, cv, 0))
#define GOPERL_CVFLAGS_CHANGE(cv, set, clear)                             \
    ((void)GOPERL_GVCALL(GOPERL_GVCALL_CV_FLAGS, cv,                      \
                         ((uint64_t)(U32)(clear) << 32) | (U32)(set)))
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
    ((void)GOPERL_GVCALL(GOPERL_GVCALL_CV_GV_SET, cv, (uint32_t)GOPERL_TOK(gv)))
#define CvSTASH_set(cv, stash) \
    ((void)GOPERL_GVCALL(GOPERL_GVCALL_CV_STASH_SET, cv, (uint32_t)GOPERL_TOK(stash)))
#define CvPROTO(cv) \
    ((const char *)goperl_intern_packed(_gof, GOPERL_GVCALL(GOPERL_GVCALL_CV_PROTO, cv, 0)))
#define CvPROTOLEN(cv) \
    ((STRLEN)(uint32_t)GOPERL_GVCALL(GOPERL_GVCALL_CV_PROTO, cv, 0))
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
        goperl_do_op(f, GOPERL_OP_SV_SETPVN, GOPERL_TOK(cv), strlen(proto),
                     proto, strlen(proto));
    return cv;
}
#define newXSproto(name, fn, file, proto) \
    goperl_newXS_proto(_gof, (name), (fn), (proto))
#define newXSproto_portable(name, fn, file, proto) \
    goperl_newXS_proto(_gof, (name), (fn), (proto))
#define newXS_flags(name, fn, file, proto, flags) \
    goperl_newXS_proto(_gof, (name), (fn), (proto))

/* ---- compile-time hints of a COP ----------------------------------------- */

#define CopHINTS_get(c) ((c)->cop_hints)
#define IN_BYTES (CopHINTS_get(PL_curcop) & HINT_BYTES)
#define IN_LOCALE_RUNTIME (CopHINTS_get(PL_curcop) & HINT_LOCALE)
#define cop_hints_fetch_pvn(cop, key, keylen, hash, flags) \
    ({ GOPERL_ONCE(goperl_cop_hints_fetch_pvn_keylen_, keylen); \
       GOPERL_SV(goperl_do_op( _gof, GOPERL_OP_SV_STR, ((uint64_t)GOPERL_SVSTR_COP_HINT << 32) | (uint32_t)(cop)->gop, ((uint64_t)(U32)(flags) << 32) | (uint32_t)(goperl_cop_hints_fetch_pvn_keylen_), (key), (uint64_t)(goperl_cop_hints_fetch_pvn_keylen_))); })
#define cop_hints_fetch_pv(cop, key, hash, flags) \
    ({ GOPERL_ONCE(goperl_cop_hints_fetch_pv_key_, key); \
       cop_hints_fetch_pvn(cop, goperl_cop_hints_fetch_pv_key_, strlen(goperl_cop_hints_fetch_pv_key_), hash, flags); })
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
#define PL_in_eval ((U8)GOPERL_PLVAR(GOPERL_PL_IN_EVAL))
#define PL_dirty (GOPERL_PLVAR(GOPERL_PL_DIRTY) != 0)
#define PL_savestack_ix ((I32)(int64_t)GOPERL_PLVAR(GOPERL_PL_SAVESTACK_IX))
#define PL_tmps_ix ((SSize_t)(int64_t)GOPERL_PLVAR(GOPERL_PL_TMPS_IX))
#define PL_tmps_floor ((SSize_t)(int64_t)GOPERL_PLVAR(GOPERL_PL_TMPS_FLOOR))
#define PL_localizing (*goperl_plint_ref(_gof, GOPERL_PL_LOCALIZING))
#define GOPERL_PL_STATUSVALUE 62
#define PL_statusvalue (*goperl_plint_ref(_gof, GOPERL_PL_STATUSVALUE))
#define PL_perl_destruct_level (*goperl_plint_ref(_gof, GOPERL_PL_DESTRUCT_LEVEL))
#define PL_unsafe (GOPERL_PLVAR(GOPERL_PL_UNSAFE) != 0)
#define PL_cop_seqmax ((U32)GOPERL_PLVAR(GOPERL_PL_COP_SEQMAX))
#define PL_origfilename \
    ((char *)goperl_intern_packed(_gof, GOPERL_PLVAR(GOPERL_PL_ORIGFILENAME)))
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
#define PL_Yes "1"
#define PL_No ""

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
        uint64_t tok = goperl_do_op(f, GOPERL_OP_NEW_PVN, 0, strlen(*argv),
                                    *argv, strlen(*argv));
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
    size_t n = strlen(pv), o = sizeof(head) - 1;
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

/* A module that runs a callback on "a new stack" keeps working on the one
 * host stack: the position to come back to is kept on the mark stack. */
#define PUSHSTACKi(type)                                                  \
    STMT_START {                                                          \
        PL_stack_sp = sp;                                                 \
        goperl_pushmark(_gof, (I32)(sp - PL_stack_base));                 \
    }                                                                     \
    STMT_END
#define PUSHSTACK PUSHSTACKi(PERLSI_UNKNOWN)
#define POPSTACK                                                          \
    STMT_START { PL_stack_sp = PL_stack_base + goperl_popmark(_gof); }    \
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
    ((void)GOPERL_SAVE(GOPERL_SAVE_GP,                                    \
                       ((uint64_t)(U32)(empty) << 32) | (uint32_t)GOPERL_TOK(gv)))
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
    uint64_t r = goperl_do_op(
        f, GOPERL_OP_SV_CALL,
        ((uint64_t)GOPERL_SVCALL_2PVBYTE << 32) | (uint32_t)GOPERL_TOK(sv),
        flags, 0, 0);
    if (lenp) *lenp = (STRLEN)(uint32_t)r;
    if (!(r >> 32)) return 0;
    return (char *)goperl_guest_mem(f, r >> 32);
}
#define SvPVbyte(sv, lenvar) \
    goperl_sv_pvbyte(_gof, (SV *)(sv), &(lenvar), SV_GMAGIC)
#define SvPVbyte_nolen(sv) goperl_sv_pvbyte(_gof, (SV *)(sv), 0, SV_GMAGIC)
#define SvPVbytex(sv, lenvar) SvPVbyte(sv, lenvar)
#define SvPVbytex_nolen(sv) SvPVbyte_nolen(sv)
#define SvPVbyte_nomg(sv, lenvar) \
    goperl_sv_pvbyte(_gof, (SV *)(sv), &(lenvar), 0)
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
    uint64_t r = goperl_do_op(
        f, GOPERL_OP_SV_CALL,
        ((uint64_t)GOPERL_SVCALL_PV_AS << 32) | (uint32_t)GOPERL_TOK(sv),
        ((uint64_t)how << 32) | (flags & SV_GMAGIC), 0, 0);
    if (lenp) *lenp = (STRLEN)(uint32_t)r;
    if (!(r >> 32)) return 0;
    return (char *)goperl_guest_mem(f, r >> 32);
}
/* The SV made a string of its own that may be written to, as it is (how 0)
 * or in an encoding. */
static char *goperl_sv_pv_force_as(goperl_frame_t *f, SV *sv, STRLEN *lenp,
                                   uint32_t how, U32 flags)
    __attribute__((unused));
static char *goperl_sv_pv_force_as(goperl_frame_t *f, SV *sv, STRLEN *lenp,
                                   uint32_t how, U32 flags) {
    uint64_t r = goperl_do_op(
        f, GOPERL_OP_SV_CALL,
        ((uint64_t)GOPERL_SVCALL_PV_FORCE_AS << 32) | (uint32_t)GOPERL_TOK(sv),
        ((uint64_t)how << 32) | flags, 0, 0);
    if (lenp) *lenp = (STRLEN)(uint32_t)r;
    if (!(r >> 32)) return 0;
    return (char *)goperl_guest_mem(f, r >> 32);
}
#define SvPVutf8(sv, lenvar) \
    goperl_sv_pv_as(_gof, (SV *)(sv), &(lenvar), GOPERL_PV_UTF8, SV_GMAGIC)
#define SvPVutf8_nomg(sv, lenvar) \
    goperl_sv_pv_as(_gof, (SV *)(sv), &(lenvar), GOPERL_PV_UTF8, 0)
#define SvPVutf8_nolen(sv) \
    goperl_sv_pv_as(_gof, (SV *)(sv), 0, GOPERL_PV_UTF8, SV_GMAGIC)
#define SvPVutf8x(sv, lenvar) SvPVutf8(sv, lenvar)
#define SvPVutf8_or_null(sv, lenvar)                         \
    goperl_sv_pv_as(_gof, (SV *)(sv), &(lenvar),             \
                    GOPERL_PV_UTF8 | GOPERL_PV_OR_NULL, SV_GMAGIC)
#define SvPVutf8_or_null_nomg(sv, lenvar)                    \
    goperl_sv_pv_as(_gof, (SV *)(sv), &(lenvar),             \
                    GOPERL_PV_UTF8 | GOPERL_PV_OR_NULL, 0)
#define SvPVbyte_or_null(sv, lenvar)                         \
    goperl_sv_pv_as(_gof, (SV *)(sv), &(lenvar),             \
                    GOPERL_PV_BYTES | GOPERL_PV_OR_NULL, SV_GMAGIC)
#define SvPVbyte_or_null_nomg(sv, lenvar)                    \
    goperl_sv_pv_as(_gof, (SV *)(sv), &(lenvar),             \
                    GOPERL_PV_BYTES | GOPERL_PV_OR_NULL, 0)
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
    goperl_sv_pv_force_as(_gof, (SV *)(sv), &(lenvar), 0, (U32)(flags))
#define SvPV_force_flags_nolen(sv, flags) \
    goperl_sv_pv_force_as(_gof, (SV *)(sv), 0, 0, (U32)(flags))
#define SvPV_force_flags_mutable(sv, lenvar, flags) \
    SvPV_force_flags(sv, lenvar, flags)
#define SvPV_force_nomg(sv, lenvar) SvPV_force_flags(sv, lenvar, 0)
#define SvPV_force_nomg_nolen(sv) SvPV_force_flags_nolen(sv, 0)
#define SvPVx_force(sv, lenvar) SvPV_force(sv, lenvar)
#define SvPVutf8_force(sv, lenvar)                                    \
    goperl_sv_pv_force_as(_gof, (SV *)(sv), &(lenvar), GOPERL_PV_UTF8, \
                          SV_GMAGIC)
#define SvPVutf8x_force(sv, lenvar) SvPVutf8_force(sv, lenvar)
#define SvPVbyte_force(sv, lenvar)                                     \
    goperl_sv_pv_force_as(_gof, (SV *)(sv), &(lenvar), GOPERL_PV_BYTES, \
                          SV_GMAGIC)
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
    size_t nl = strlen(nam), vl = val ? strlen(val) : 0;
    char *buf = (char *)malloc(nl + vl + 2);
    if (!buf) goperl_croakf(f, "Out of memory!");
    memcpy(buf, nam, nl + 1);
    memcpy(buf + nl + 1, val ? val : "", vl + 1);
    goperl_do_op(f, GOPERL_OP_SV_STR, (uint64_t)GOPERL_SVSTR_SETENV << 32,
                 ((uint64_t)(val ? 0 : 1) << 32) | (uint32_t)(nl + vl + 2), buf,
                 nl + vl + 2);
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
    uint64_t r = goperl_do_op(
        f, GOPERL_OP_SV_CALL,
        ((uint64_t)GOPERL_SVCALL_UNI_DISPLAY << 32) | (uint32_t)GOPERL_TOK(dsv),
        ((((uint64_t)pvlim << 8) | (flags & 0xFF)) << 32) |
            (uint32_t)GOPERL_TOK(ssv),
        0, 0);
    if (!(r >> 32)) return 0;
    return (char *)goperl_guest_mem(f, r >> 32);
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
    ((void)goperl_do_op(_gof, GOPERL_OP_SV_CALL,                          \
                        ((uint64_t)GOPERL_SVCALL_HV_KSPLIT << 32) |       \
                            (uint32_t)GOPERL_TOK(hv),                     \
                        (uint64_t)(int64_t)(newmax), 0, 0))
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
    Size_t length = strlen(src);
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
    goperl_op0(GOPERL_OP_FIELD,                                             \
               ((uint64_t)(sel) << 32) | (uint32_t)GOPERL_TOK(tok), (b))

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
 * saved_riter; HvEITER(hv) = saved_eiter;`). An entry is a shadow that
 * lasts beyond the call (goperl_eiter_shadow). */
#define HvRITER(hv) \
    (*(I32 *)goperl_lv_ref(_gof, (SV *)(hv), GOPERL_LV_HVRITER, GOPERL_LV_AT))
#define HvEITER(hv) \
    (*(HE **)goperl_lv_ref(_gof, (SV *)(hv), GOPERL_LV_HVEITER, GOPERL_LV_AT))
#define HvEITER_get(hv)                                                     \
    ({ GOPERL_ONCE(goperl_HvEITER_get_hv_, hv);                             \
       goperl_eiter_shadow(_gof, GOPERL_TOK(goperl_HvEITER_get_hv_),        \
                           GOPERL_FIELD(GOPERL_FIELD_HVAUX_EITER,           \
                                        goperl_HvEITER_get_hv_, 0)); })
#define HvEITER_set(hv, e)                                                  \
    ({ HE *const goperl_HvEITER_set_e_ = (e);                               \
       (void)GOPERL_FIELD(GOPERL_FIELD_HVAUX_EITER_SET, hv,                 \
                          goperl_HvEITER_set_e_ ? goperl_HvEITER_set_e_->goperl_tok : 0); })

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
    goperl_op0(GOPERL_OP_CALL,                                              \
               ((uint64_t)(sel) << 32) | (uint32_t)GOPERL_TOK(tok), (b))
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
    uint32_t old = 0;
    if (buffer && d < f->memcap && buffer != pWARN_ALL && buffer != pWARN_NONE)
        old = (uint32_t)d;
    return (char *)goperl_guest_mem(
        f, goperl_do_op(f, GOPERL_OP_CALL,
                        ((uint64_t)GOPERL_CALL_NEW_WARNINGS << 32) | old,
                        (uint64_t)size, bits, (uint64_t)size));
}
#define Perl_new_warnings_bitfield(...) goperl_new_warnings_bitfield(_gof, __VA_ARGS__)
#define new_warnings_bitfield(buffer, bits, size) \
    goperl_new_warnings_bitfield(_gof, (buffer), (bits), (size))

#define CopLINE_set(c, l) ((c)->cop_line = (l))
#define CopLINE_inc(c) (++(c)->cop_line)
#define CopLINE_dec(c) (--(c)->cop_line)
static void goperl_cop_file_set(goperl_frame_t *f, COP *c, const char *pv)
    __attribute__((unused));
static void goperl_cop_file_set(goperl_frame_t *f, COP *c, const char *pv) {
    if (c->gop)
        goperl_do_op(f, GOPERL_OP_CALL,
                     ((uint64_t)GOPERL_CALL_COP_SET << 32) | (uint32_t)c->gop,
                     (uint64_t)4 << 32, pv, strlen(pv));
    c->cop_file = goperl_intern(pv, strlen(pv));
}
#define CopFILE_set(c, pv) goperl_cop_file_set(_gof, (c), (pv))
static void goperl_cop_stash_set(goperl_frame_t *f, COP *c, HV *hv)
    __attribute__((unused));
static void goperl_cop_stash_set(goperl_frame_t *f, COP *c, HV *hv) {
    if (c->gop)
        goperl_do_op(f, GOPERL_OP_CALL,
                     ((uint64_t)GOPERL_CALL_COP_SET << 32) | (uint32_t)c->gop,
                     ((uint64_t)3 << 32) | (uint32_t)GOPERL_TOK(hv), 0, 0);
    const char *name = hv ? HvNAME(hv) : 0;
    c->cop_stashpv = name ? goperl_intern(name, strlen(name)) : "";
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
    goperl_pp_stack_push(f);
    goperl_do_op(f, GOPERL_OP_CX, (uint64_t)GOPERL_CX_DOUNWIND << 32,
                 (uint64_t)(int64_t)cxix, 0, 0);
    f->st_stale = f->pp_frame;
}
#define dounwind(cxix) goperl_dounwind(_gof, (cxix))
#define Perl_dounwind(...) goperl_dounwind(_gof, __VA_ARGS__)
#define cxstack_max \
    ((I32)(int64_t)goperl_op0(GOPERL_OP_CX, (uint64_t)GOPERL_CX_MAX << 32, 0))
#define CX_CUR() (&cxstack[cxstack_ix])

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
    goperl_op0(GOPERL_OP_REGEXP,                                            \
               ((uint64_t)(sel) << 32) | (uint32_t)GOPERL_TOK(tok), (b))

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

static regexp *goperl_reany(goperl_frame_t *f, REGEXP *rx) __attribute__((unused));
static regexp *goperl_reany(goperl_frame_t *f, REGEXP *rx) {
    const uint32_t *w = (const uint32_t *)goperl_guest_mem(
        f, goperl_do_op(f, GOPERL_OP_REGEXP,
                        ((uint64_t)GOPERL_RX_BODY << 32) | (uint32_t)GOPERL_TOK(rx),
                        0, 0, 0));
    if (!w) return 0;
    uint32_t c[20];
    memcpy(c, w, sizeof c);
    regexp *r = (regexp *)goperl_arena_alloc(
        sizeof(regexp) + sizeof(regexp_paren_pair) * ((size_t)c[4] + 1));
    if (!r) goperl_croakf(f, "Out of memory!");
    /* an engine is known by its address in the interpreter */
    r->engine = (const regexp_engine *)(uintptr_t)c[0];
    r->mother_re = (REGEXP *)GOPERL_SV(c[1]);
    r->paren_names = (HV *)GOPERL_SV(c[2]);
    r->extflags = c[3];
    r->nparens = c[4];
    r->minlen = (SSize_t)(int32_t)c[5];
    r->minlenret = (SSize_t)(int32_t)c[6];
    r->gofs = c[7];
    r->intflags = c[8];
    r->lastparen = c[9];
    r->lastcloseparen = c[10];
    r->subbeg = c[11] ? (char *)goperl_guest_mem(f, c[11]) : 0;
    r->sublen = (SSize_t)(int32_t)c[12];
    r->suboffset = (SSize_t)(int32_t)c[13];
    r->subcoffset = (SSize_t)(int32_t)c[14];
    r->compflags = c[15];
    r->precomp = c[16] ? (char *)goperl_guest_mem(f, c[16]) : 0;
    r->prelen = c[17];
    r->wrapped = c[18] ? (char *)goperl_guest_mem(f, c[18]) : 0;
    r->wraplen = c[19];
    r->offs = (regexp_paren_pair *)(r + 1);
    for (U32 i = 0; i <= r->nparens; i++) {
        const uint64_t se = goperl_do_op(
            f, GOPERL_OP_REGEXP,
            ((uint64_t)GOPERL_RX_OFFS << 32) | (uint32_t)GOPERL_TOK(rx), i, 0, 0);
        r->offs[i].start = (SSize_t)(int32_t)(se >> 32);
        r->offs[i].end = (SSize_t)(int32_t)se;
        r->offs[i].start_tmp = -1;
    }
    return r;
}
#define ReANY(rx) goperl_reany(_gof, (REGEXP *)(rx))
static void *goperl_svany(goperl_frame_t *f, SV *sv) __attribute__((unused));
static void *goperl_svany(goperl_frame_t *f, SV *sv) {
    if (sv && SvTYPE(sv) == SVt_REGEXP) return goperl_reany(f, (REGEXP *)sv);
    return 0;
}
#undef SvANY
#define SvANY(sv) goperl_svany(_gof, (SV *)(sv))

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
#define get_regex_charset(flags) \
    ((regex_charset)(((flags) & RXf_PMf_CHARSET) >> 5))
typedef enum {
    REGEX_DEPENDS_CHARSET = 0,
    REGEX_LOCALE_CHARSET,
    REGEX_UNICODE_CHARSET,
    REGEX_ASCII_RESTRICTED_CHARSET,
    REGEX_ASCII_MORE_RESTRICTED_CHARSET
} regex_charset;
#define set_regex_charset(flagsp, cs) \
    (*(flagsp) = (*(flagsp) & ~RXf_PMf_CHARSET) | ((U32)(cs) << 5))

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
    uint32_t w[6];
    const uintptr_t b = (uintptr_t)strbeg - (uintptr_t)f->membase;
    const uintptr_t e = (uintptr_t)strend - (uintptr_t)f->membase;
    if (b < f->memcap && e < f->memcap) {
        w[0] = (uint32_t)((uintptr_t)stringarg - (uintptr_t)f->membase);
        w[1] = (uint32_t)e;
        w[2] = (uint32_t)b;
    } else {
        SV *copy = sv_2mortal(newSVpvn(strbeg, (STRLEN)(strend - strbeg)));
        if (screamer && SvUTF8(screamer)) SvUTF8_on(copy);
        const uint32_t base = (uint32_t)goperl_do_op(f, GOPERL_OP_SV_PVX_RAW,
                                                     GOPERL_TOK(copy), 0, 0, 0);
        w[2] = base;
        w[0] = base + (uint32_t)(stringarg - strbeg);
        w[1] = base + (uint32_t)(strend - strbeg);
        screamer = copy;
    }
    w[3] = (uint32_t)(int32_t)minend;
    w[4] = (uint32_t)GOPERL_TOK(screamer);
    w[5] = nosave;
    return (I32)(int64_t)goperl_do_op(
        f, GOPERL_OP_REGEXP,
        ((uint64_t)GOPERL_RX_EXEC << 32) | (uint32_t)GOPERL_TOK(rx), sizeof w,
        (const char *)w, sizeof w);
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
    const size_t n1 = c1 ? strlen(c1) : 0, n2 = c2 ? strlen(c2) : 0;
    const size_t total = 24 + n1 + n2 + count;
    char *req = (char *)malloc(total ? total : 1);
    if (!req) goperl_croakf(f, "Out of memory!");
    const uint32_t w[6] = {(uint32_t)count, (uint32_t)max, flags,
                           (uint32_t)len, (uint32_t)n1, (uint32_t)n2};
    memcpy(req, w, 24);
    if (n1) memcpy(req + 24, c1, n1);
    if (n2) memcpy(req + 24 + n1, c2, n2);
    if (count) memcpy(req + 24 + n1 + n2, pv, count);
    /* -> how much of the text was taken << 32 | where the result is */
    const uint64_t r = goperl_do_op(
        f, GOPERL_OP_CALL, ((uint64_t)fn << 32) | (uint32_t)GOPERL_TOK(dsv),
        (uint64_t)total, req, (uint64_t)total);
    free(req);
    if (escaped) *escaped = (STRLEN)(r >> 32);
    return (uint32_t)r ? (char *)goperl_guest_mem(f, (uint32_t)r) : 0;
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
 * own: PL_markstack_ptr, TOPMARK.) */
#define PL_markstack \
    ((I32 *)goperl_guest_mem(_gof, GOPERL_PLVAR(GOPERL_PL_MARKSTACK)))
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
    const uint32_t *w = (const uint32_t *)goperl_guest_mem(
        f, goperl_do_op(f, GOPERL_OP_CALL,
                        ((uint64_t)10 << 32) | (uint32_t)GOPERL_TOK(stash), 0, 0, 0));
    if (!w) return 0;
    AMT *amt = (AMT *)goperl_arena_alloc(sizeof(AMT));
    if (!amt) goperl_croakf(f, "Out of memory!");
    if (w[3] > GOPERL_AMT_METHODS)
        goperl_croakf(f, "the interpreter has more overload methods than the "
                         "go-perl XS SDK's AMT");
    amt->flags = (U8)w[0];
    amt->fallback = (U8)w[1];
    amt->was_ok_sub = w[2];
    for (uint32_t i = 0; i < w[3]; i++) amt->table[i] = (CV *)GOPERL_SV(w[4 + i]);
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
    cvt.u = goperl_do_op(f, GOPERL_OP_CALL, (uint64_t)9 << 32, 0, s, strlen(s));
    return cvt.d;
}
#define my_atof(s) goperl_my_atof(_gof, (s))
#define Atof my_atof
#define HV_ITERNEXT_WANTPLACEHOLDERS 0x01
#define hv_iternext_flags(hv, flags)                                         \
    ({ GOPERL_ONCE(goperl_hv_iternext_flags_hv_, hv);                        \
       goperl_he_shadow(_gof, (HV *)(goperl_hv_iternext_flags_hv_),          \
                        goperl_op0(GOPERL_OP_HV_ITERNEXT,                    \
                                   GOPERL_TOK(goperl_hv_iternext_flags_hv_), \
                                   (uint64_t)(int64_t)(flags))); })
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
    const uint32_t w[2] = {(uint32_t)GOPERL_TOK(sv), flags};
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
    const uint32_t w[4] = {(uint32_t)GOPERL_TOK(name), flags,
                           (uint32_t)GOPERL_TOK(typestash),
                           (uint32_t)GOPERL_TOK(ourstash)};
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
    goperl_do_op(f, GOPERL_OP_CALL, ((uint64_t)14 << 32) | (uint32_t)GOPERL_TOK(line),
                 flags, 0, 0);
    goperl_parser_pull(f);
}
#define lex_start(line, rsfp, flags) \
    goperl_lex_start(_gof, (SV *)(line), (void *)(rsfp), (U32)(flags))
/* The line and the file of a COP are saved by the interpreter, which puts
 * them back when the scope ends (CALL 13: 0 the line, 1 the file). */
#define SAVECOPLINE(c)                                              \
    ((void)goperl_op0(GOPERL_OP_CALL,                               \
                      ((uint64_t)13 << 32) | (uint32_t)(c)->gop, 0))
#define SAVECOPFILE_FREE(c)                                         \
    ((void)goperl_op0(GOPERL_OP_CALL,                               \
                      ((uint64_t)13 << 32) | (uint32_t)(c)->gop, 1))
#define SAVECOPFILE(c) SAVECOPFILE_FREE(c)
/* The stash of a COP, saved the same way (CALL 13 with 2): the scope's end
 * puts it back in the interpreter's COP and in its copy here. */
#define SAVECOPSTASH_FREE(c)                                        \
    ({ COP *const goperl_scsf_c_ = (c);                              \
       (void)goperl_op0(GOPERL_OP_CALL,                              \
                        ((uint64_t)13 << 32) | (uint32_t)goperl_scsf_c_->gop, 2); \
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
static OP *goperl_newunop_aux(goperl_frame_t *f, I32 type, I32 flags, OP *first,
                              UNOP_AUX_item *aux) __attribute__((unused));
static OP *goperl_newunop_aux(goperl_frame_t *f, I32 type, I32 flags, OP *first,
                              UNOP_AUX_item *aux) {
    uint32_t w[7] = {(uint32_t)type, (uint32_t)flags, goperl_optok(f, first), 0, 0, 0, 0};
    switch (type) {
    case OP_ARGELEM:
        w[3] = 1;
        w[4] = (uint32_t)(uintptr_t)aux;
        break;
    case OP_ARGCHECK:
        if (aux) {
            const struct op_argcheck_aux *ca = (const struct op_argcheck_aux *)aux;
            w[3] = 2;
            w[4] = (uint32_t)ca->params;
            w[5] = (uint32_t)ca->opt_params;
            w[6] = (uint32_t)(unsigned char)ca->slurpy;
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
/* An SV in an item fills its low half (the interpreter's pointers are 32
 * bits wide, and a module's own items hold the same handles). */
#define UNOP_AUX_item_sv(item) GOPERL_SV((uint32_t)(item)->uv)

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
    const uint32_t w[1] = {flags};
    goperl_parser_push(f);
    const uint64_t tok = goperl_tree(f, GOPERL_TREE_PARSE_SUBSIGNATURE, 1, w);
    goperl_parser_pull(f);
    return goperl_tree_op(f, tok);
}
#define parse_subsignature(flags) goperl_parse_subsignature(_gof, (U32)(flags))
static inline __attribute__((unused)) CV *goperl_newmysub(
    goperl_frame_t *f, I32 floor, OP *o, OP *proto, OP *attrs, OP *block) {
    const uint32_t w[5] = {(uint32_t)floor, goperl_optok(f, o),
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
 * (CvNAME_HEK_set). A negative length says the key is UTF-8. */
static HEK *goperl_hek_of(goperl_frame_t *f, const char *key, I32 len,
                          int utf8, U32 hash) __attribute__((unused));
static HEK *goperl_hek_of(goperl_frame_t *f, const char *key, I32 len,
                          int utf8, U32 hash) {
    HEK *hek = (HEK *)goperl_arena_alloc(sizeof(HEK) + (size_t)len + 2);
    if (!hek) goperl_croakf(f, "share_hek: out of memory");
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
    const uint32_t w[4] = {(uint32_t)GOPERL_TOK(cv), 0, 0, 0};
    const uint64_t r = goperl_tree(f, GOPERL_TREE_CV_NAME, 4, w);
    if (!r) return 0;
    const I32 len = (I32)((uint32_t)r >> 1);
    return goperl_hek_of(f, (const char *)goperl_guest_mem(f, r >> 32), len,
                         (int)(r & 1), 0);
}
#define CvNAME_HEK(cv) goperl_cv_name_hek(_gof, (CV *)(cv))
static void goperl_cv_name_hek_set(goperl_frame_t *f, CV *cv, HEK *hek)
    __attribute__((unused));
static void goperl_cv_name_hek_set(goperl_frame_t *f, CV *cv, HEK *hek) {
    const size_t len = (size_t)hek->hek_len;
    char *req = (char *)malloc(16 + len + 1);
    if (!req) goperl_croakf(f, "Out of memory!");
    const uint32_t w[4] = {(uint32_t)GOPERL_TOK(cv), 1,
                           (uint32_t)(hek->hek_key[len + 1] & 1),
                           (uint32_t)len};
    memcpy(req, w, 16);
    memcpy(req + 16, hek->hek_key, len);
    req[16 + len] = 0;
    goperl_optree_flush(f);
    goperl_do_op(f, GOPERL_OP_TREE, GOPERL_TREE_CV_NAME, 16 + len + 1, req,
                 16 + len + 1);
    free(req);
}
#define CvNAME_HEK_set(cv, hek) \
    goperl_cv_name_hek_set(_gof, (CV *)(cv), (HEK *)(hek))


static void *goperl_multiparam_aux(goperl_frame_t *f, const uint32_t *g) {
    if (!g) return 0;
    const size_t n_positional = g[1], n_named = g[2];
    struct op_multiparam_aux *aux = (struct op_multiparam_aux *)malloc(
        sizeof(*aux) + n_positional * sizeof(PADOFFSET) +
        n_named * sizeof(struct op_multiparam_named_aux));
    if (!aux) goperl_croakf(f, "out of memory for the parameters of a signature");
    aux->min_args = g[0];
    aux->n_positional = n_positional;
    aux->n_named = n_named;
    aux->slurpy = (char)(g[3] & 0xFF);
    aux->slurpy_padix = (PADOFFSET)g[4];
    aux->param_padix = (PADOFFSET *)(aux + 1);
    aux->named = (struct op_multiparam_named_aux *)(aux->param_padix + n_positional);
    const uint32_t *padix = (const uint32_t *)goperl_guest_mem(f, g[5]);
    for (size_t i = 0; padix && i < n_positional; i++)
        aux->param_padix[i] = (PADOFFSET)padix[i];
    const uint32_t *named = (const uint32_t *)goperl_guest_mem(f, g[6]);
    for (size_t i = 0; named && i < n_named; i++) {
        const uint32_t *n = named + i * 5;
        aux->named[i].padix = (PADOFFSET)n[0];
        aux->named[i].namepv = (const char *)goperl_guest_mem(f, n[1]);
        aux->named[i].namelen = (STRLEN)n[2];
        aux->named[i].namehash = n[3];
        aux->named[i].is_required = (n[4] & 1) != 0;
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
       lex_stuff_pvn(goperl_lex_stuff_pv_, strlen(goperl_lex_stuff_pv_), (flags)); })

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

/* What is pending is delivered (the handler of a signal), as between two
 * ops: despatch_signals, which PERL_ASYNC_CHECK runs. */
#define despatch_signals() goperl_async(_gof)
#define Perl_despatch_signals(...) goperl_async(_gof)

/* save_re_context: the match variables of the code calling native code are
 * kept from the matches of the code it calls (CALL 22). */
#define save_re_context() ((void)goperl_do_op(_gof, GOPERL_OP_CALL, (uint64_t)22 << 32, 0, 0, 0))

/* sv_peek: perl's short description of a value (dump.c), the text of a
 * mortal of the interpreter (CALL 23). */
static __attribute__((unused)) char *goperl_sv_peek(goperl_frame_t *f, SV *sv) {
    return (char *)goperl_guest_mem(
        f, goperl_do_op(f, GOPERL_OP_CALL, ((uint64_t)23 << 32) | (uint32_t)GOPERL_TOK(sv), 0, 0, 0));
}
#define sv_peek(sv) goperl_sv_peek(_gof, (SV *)(sv))

/* sv_recode_to_utf8: a string of bytes decoded by an encoding object into
 * characters (its decode method), in place (CALL 24). */
static __attribute__((unused)) char *goperl_sv_recode_to_utf8(goperl_frame_t *f, SV *sv, SV *encoding) {
    return (char *)goperl_guest_mem(
        f, goperl_do_op(f, GOPERL_OP_CALL, ((uint64_t)24 << 32) | (uint32_t)GOPERL_TOK(sv),
                        (uint32_t)GOPERL_TOK(encoding), 0, 0));
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
    uint32_t w[7];
    const uintptr_t b = (uintptr_t)strbeg - (uintptr_t)f->membase;
    const uintptr_t e = (uintptr_t)strend - (uintptr_t)f->membase;
    if (data) goperl_croakf(f, "CALLREGEXEC: the data argument of the engine is not available to native XS modules");
    if (b < f->memcap && e < f->memcap) {
        w[0] = (uint32_t)((uintptr_t)stringarg - (uintptr_t)f->membase);
        w[1] = (uint32_t)e;
        w[2] = (uint32_t)b;
    } else {
        SV *copy = sv_2mortal(newSVpvn(strbeg, (STRLEN)(strend - strbeg)));
        if (screamer && SvUTF8(screamer)) SvUTF8_on(copy);
        const uint32_t base = (uint32_t)goperl_do_op(f, GOPERL_OP_SV_PVX_RAW,
                                                     GOPERL_TOK(copy), 0, 0, 0);
        w[2] = base;
        w[0] = base + (uint32_t)(stringarg - strbeg);
        w[1] = base + (uint32_t)(strend - strbeg);
        screamer = copy;
    }
    w[3] = (uint32_t)(int32_t)minend;
    w[4] = (uint32_t)GOPERL_TOK(screamer);
    w[5] = flags;
    w[6] = 0;
    return (I32)(int64_t)goperl_do_op(
        f, GOPERL_OP_REGEXP,
        ((uint64_t)GOPERL_RX_EXEC_FLAGS << 32) | (uint32_t)GOPERL_TOK(rx), sizeof w,
        (const char *)w, sizeof w);
}
#define CALLREGEXEC(rx, stringarg, strend, strbeg, minend, sv, data, flags)          \
    goperl_regexec_flags(_gof, (REGEXP *)(rx), (char *)(stringarg), (char *)(strend), \
                         (char *)(strbeg), (SSize_t)(minend), (SV *)(sv), (void *)(data), \
                         (U32)(flags))
#define regexec_flags(rx, stringarg, strend, strbeg, minend, sv, data, flags) \
    CALLREGEXEC(rx, stringarg, strend, strbeg, minend, sv, data, flags)

#include "goperl_compiler.h"

#endif /* GOPERL_XS_SDK_PERLAPI_H */
