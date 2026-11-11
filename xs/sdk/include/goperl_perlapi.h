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
    GOPERL_GVCALL_STASH_SET = 17
};

/* The glob and CV fields that are lvalues (see GOPERL_LV_CVXSUB). */
static uint64_t goperl_lv_load_field(goperl_frame_t *f, uint64_t tok, int field) {
    switch (field) {
    case GOPERL_LV_CVXSUB:
        return (uint64_t)(uintptr_t)goperl_api_cv_xsub(f, tok);
    case GOPERL_LV_GVSTASH:
        return goperl_do_op(f, GOPERL_OP_GV_PTR, tok, 0, 0, 0);
    default:
        return goperl_do_op(f, GOPERL_OP_GV_CALL,
                            ((uint64_t)GOPERL_GVCALL_CVGEN << 32) | (uint32_t)tok,
                            0, 0, 0);
    }
}
static void goperl_lv_store_field(goperl_frame_t *f, uint64_t tok, int field,
                                  uint64_t val) {
    switch (field) {
    case GOPERL_LV_CVXSUB:
        goperl_api_cv_set_xsub(f, tok, (void *)(uintptr_t)val);
        break;
    case GOPERL_LV_GVSTASH:
        goperl_xs_op(f, GOPERL_OP_GV_CALL,
                     ((uint64_t)GOPERL_GVCALL_STASH_SET << 32) | (uint32_t)tok,
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
    GOPERL_SVSTR_TAINT_PROPER = 15
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
#define PERL_UNUSED_RESULT(v) ((void)sizeof(v))
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
 * describe the interpreter's 32-bit integers instead, because they decide
 * what a value can hold once it is stored in an SV. */
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
#define BYTEORDER 0x1234
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

#define PTR2nat(p) ((PTRV)(p))
#define PTR2NV(p) ((NV)PTR2nat(p))
#define DPTR2FPTR(t, p) ((t)PTR2nat(p))
#define FPTR2DPTR(t, p) ((t)PTR2nat(p))

/* The saturating NV casts of numeric.c, at the interpreter's integer width. */
static inline IV goperl_cast_iv(NV f) {
    if (f < (NV)IV_MAX + 1.0) return f < (NV)IV_MIN ? IV_MIN : (IV)f;
    if (f < (NV)UV_MAX + 1.0) return (IV)(int32_t)(uint32_t)f;
    return f > 0 ? (IV)(int32_t)(uint32_t)UV_MAX : 0; /* NaN */
}
static inline UV goperl_cast_uv(NV f) {
    if (f < 0.0) return f < (NV)IV_MIN ? (UV)(uint32_t)IV_MIN : (UV)(uint32_t)(int32_t)f;
    if (f < (NV)UV_MAX + 1.0) return (UV)f;
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
#define PERL_GET_INTERP PERL_GET_CONTEXT
#define my_perl goperl_cur_frame_v
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

#define GOPERL_FLAGS_CHANGE(sv, set, clear)                               \
    ((void)goperl_op0(GOPERL_OP_SV_FLAGS, GOPERL_TOK(sv),                 \
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
#define SvLEN_set(sv, n) GOPERL_SLOT_SET(GOPERL_SVSLOT_LEN, sv, (STRLEN)(n))
/* The string buffer of an SV is interpreter memory, so the only pointer a
 * module can store there is one the interpreter handed out (SvPVX, SvGROW)
 * or NULL. */
static void goperl_sv_pv_set(goperl_frame_t *f, SV *sv, const char *p) {
    const uint64_t g = goperl_guest_addr(f, p);
    if (p && !g)
        goperl_croakf(f, "SvPV_set: the buffer is not interpreter memory");
    goperl_do_op(f, GOPERL_OP_SV_SLOT_SET,
                 ((uint64_t)GOPERL_SVSLOT_PV << 32) | (uint32_t)GOPERL_TOK(sv), g,
                 0, 0);
}
#define SvPV_set(sv, p) goperl_sv_pv_set(_gof, (SV *)(sv), (const char *)(p))
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
    SV *r = GOPERL_SV(goperl_do_op(
        f, GOPERL_OP_SV_CALL,
        ((uint64_t)GOPERL_SVCALL_REF << 32) | (uint32_t)GOPERL_TOK(sv),
        (uint64_t)ob, 0, 0));
    if (!dst) return r;
    /* the interpreter fills the caller's SV; here the fresh one is copied
     * into it and released */
    goperl_do_op(f, GOPERL_OP_SV_SETSV, GOPERL_TOK(dst), GOPERL_TOK(r), 0, 0);
    goperl_do_op(f, GOPERL_OP_SV_REFCNT_DEC, GOPERL_TOK(r), 0, 0, 0);
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
#define IoTYPE(io) ((char)GOPERL_SVCALL(GOPERL_SVCALL_IO_TYPE, io, 0))
#define IoFLAGS(io) ((U8)GOPERL_SVCALL(GOPERL_SVCALL_IO_FLAGS, io, 0))
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
    if (!args || svargs || sv_count)
        goperl_croakf(f, "sv_v%spvfn: formatting from an SV argument list is "
                         "not available to native XS modules",
                      cat ? "cat" : "set");
    /* the format is patlen bytes, not necessarily terminated there */
    char *fmt = (char *)malloc(patlen + 1);
    if (!fmt) goperl_croakf(f, "%s", "Out of memory!");
    memcpy(fmt, pat, patlen);
    fmt[patlen] = '\0';
    goperl_text_init(&t);
    goperl_vtext(f, &t, fmt, *args);
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
    if (!pat) goperl_croak_errsv(f);
    goperl_vcroak(f, pat, *args);
}
#define vcroak(pat, args) goperl_vcroak_args(_gof, (pat), (args))
#define vwarn(pat, args) goperl_vwarn(_gof, (pat), *(args))
static char *goperl_vform(goperl_frame_t *f, const char *pat, va_list *args)
    __attribute__((unused));
static char *goperl_vform(goperl_frame_t *f, const char *pat, va_list *args) {
    char *buf = goperl_form_bufs_v[goperl_form_ix_v++ % GOPERL_FORM_BUFS];
    goperl_vfmt(f, buf, GOPERL_FORM_LEN, pat, *args);
    return buf;
}
#define vform(pat, args) goperl_vform(_gof, (pat), (args))
#define die_sv(sv) croak_sv(sv)

/* ---- globs and subs (gv.h, cv.h) ----------------------------------------- */

#define GvSVn(gv) GOPERL_SV(GOPERL_GVCALL(GOPERL_GVCALL_SVN, gv, 0))
#define GvAVn(gv) ((AV *)GOPERL_SV(GOPERL_GVCALL(GOPERL_GVCALL_AVN, gv, 0)))
#define GvHVn(gv) ((HV *)GOPERL_SV(GOPERL_GVCALL(GOPERL_GVCALL_HVN, gv, 0)))
#define GvIOn(gv) ((IO *)GOPERL_SV(GOPERL_GVCALL(GOPERL_GVCALL_ION, gv, 0)))
#define GvIOp(gv) ((IO *)GvIO(gv))
#define GvFORM(gv) ((CV *)GOPERL_SV(GOPERL_GVCALL(GOPERL_GVCALL_FORM, gv, 0)))
#define GvCV_set(gv, cv) \
    ((void)GOPERL_GVCALL(GOPERL_GVCALL_CV_SET, gv, (uint32_t)GOPERL_TOK(cv)))
#define GvFLAGS(gv) ((U32)GOPERL_GVCALL(GOPERL_GVCALL_FLAGS, gv, 0))
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
#define GvCVGEN(gv) (*(U32 *)goperl_lv_ref(_gof, (SV *)(gv), GOPERL_LV_GVCVGEN))
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
#define CvMETHOD(cv) (CvFLAGS(cv) & CVf_IsMETHOD)
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
#define PL_strtab ((HV *)GOPERL_SV(GOPERL_PLVAR(GOPERL_PL_STRTAB)))
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
#define SAVEIV(i) goperl_hostsave_ptr(_gof, (void **)&(i))
#define SAVELONG(l) goperl_hostsave_ptr(_gof, (void **)&(l))
#define SAVESTRLEN(l) goperl_hostsave_ptr(_gof, (void **)&(l))
#define SAVEPPTR(p) goperl_hostsave_ptr(_gof, (void **)&(p))
#define save_int(p) goperl_hostsave_i32(_gof, (int32_t *)(p))
#define save_I32(p) goperl_hostsave_i32(_gof, (int32_t *)(p))
#define save_long(p) goperl_hostsave_ptr(_gof, (void **)(p))
#define save_iv(p) goperl_hostsave_ptr(_gof, (void **)(p))

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
       (UV)goperl_UVCHR_SKIP_uv_ < 0x80 ? 1 : (UV)goperl_UVCHR_SKIP_uv_ < 0x800 ? 2 : (UV)goperl_UVCHR_SKIP_uv_ < 0x10000 ? 3 : (UV)goperl_UVCHR_SKIP_uv_ < 0x200000 ? 4 : (UV)goperl_UVCHR_SKIP_uv_ < 0x4000000 ? 5 : (UV)goperl_UVCHR_SKIP_uv_ < 0x80000000 ? 6 : 7; })
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
#define SvPVbyte_force(sv, lenvar) \
    ({ GOPERL_ONCE(goperl_SvPVbyte_force_sv_, sv); \
       (sv_utf8_downgrade((SV *)(goperl_SvPVbyte_force_sv_), 0), SvPV_force(goperl_SvPVbyte_force_sv_, lenvar)); })
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
#define croak_memory_wrap() croak("%s", PL_memory_wrap)
#define croak_nocontext croak
#define die_nocontext croak
#define warn_nocontext warn
#define Perl_warn_nocontext warn
#define Perl_die Perl_croak
#define Perl_newSVpvf goperl_newSVpvf
#define Perl_sv_setpvf goperl_sv_setpvf
#define Perl_sv_catpvf goperl_sv_catpvf
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

#endif /* GOPERL_XS_SDK_PERLAPI_H */
