/* go-perl native XS SDK — the perl.h an XS module is compiled against when
 * it targets go-perl's embedded interpreter as a HOST-NATIVE shared library.
 *
 * This is NOT perl's own perl.h. The real header exposes the interpreter's
 * struct layouts and lets XS macros compile into direct memory access; that
 * can never match go-perl, whose interpreter lives in a wasm64 linear memory
 * (offsets as handles). Here every SV is an opaque token
 * and every operation is a call through a vtable the loader injects at
 * dlopen time (__goperl_xs_init). Anything this SDK does not cover fails AT
 * COMPILE TIME — never silently at runtime.
 *
 * v3 models enough of the interpreter surface for stack-driven, MAGIC-using,
 * pp-poking XS (the Text::Xslate class of module):
 *
 *   - Context: as in perl built without MULTIPLICITY, pTHX is void and
 *     aTHX nothing; what a native call works on is its frame, which the
 *     loader's entries make current (goperl_cur_frame_v, per thread).
 *     XSUBs have perl's signature there: void xsub(CV* cv).
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
/* The SDK's headers are compiled as the system's are: a module built
 * with warnings made errors (-Werror, -Wextra,
 * -Werror=declaration-after-statement) compiles against them as it does
 * against perl's, whose headers are declarations and macros only; the
 * functions and initializers of these headers warn of nothing in the
 * module's code. Each header says so of itself. */
#pragma GCC system_header

/* The loader's entry points and the state the translation units of a module
 * share have C linkage in whatever language includes the header: a C++
 * module that includes it outside extern "C" (Algorithm::LibLinear) would
 * otherwise define them under mangled names the loader does not find, and a
 * module made of C and C++ units would have two of each. */
#ifdef __cplusplus
#define GOPERL_C extern "C"
#else
#define GOPERL_C
#endif

/* The stdio stream of a handle (PerlIO_findFILE, PerlIO_exportFILE) is built
 * on fopencookie where there is no funopen; glibc and musl declare it only
 * under _GNU_SOURCE, which has to be in effect before the first system
 * header. perl's own Linux builds compile XS with it defined too. */
#if !defined(__APPLE__) && !defined(_GNU_SOURCE)
#  define _GNU_SOURCE 1
#endif

/* The interpreter is not a DEBUGGING build: as perl's perl.h does, the
 * assertions of the C library are off in the code compiled against it
 * (assert() is the perl API's assertion, which such a build leaves out). */
#if !defined(DEBUGGING) && !defined(NDEBUG)
#  define NDEBUG 1
#endif
#include <assert.h>
#include <setjmp.h>
#include <stdarg.h>
#include <stddef.h>
#include <stdint.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <strings.h>
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
/* perl's own perl.h pulls these in on every unix; XS sources rely on that
 * (htonl without <arpa/inet.h>, open flags without <fcntl.h>, ...). */
#include <sys/types.h>
#include <inttypes.h>
#include <limits.h>
#include <fcntl.h>
#include <signal.h>
#include <dlfcn.h>
#include <locale.h>
#include <dirent.h>
#include <sys/ioctl.h>
#include <sys/wait.h>
#include <sys/socket.h>
#include <sys/select.h>
#include <netinet/in.h>
#include <arpa/inet.h>
#include <netdb.h>
#include <sys/file.h>

/* The catch points of native code (an XSUB's entry, XCPT, JMPENV, the
 * loader's other entries) and the jump to them. On macOS they are _setjmp
 * and _longjmp, which keep no signal state. setjmp and longjmp there also
 * keep and put back the signal mask and whether the thread runs on its
 * alternate signal stack, and on arm64 setjmp leaves the words longjmp takes
 * the second from as they were: a croak marked the thread as running on
 * that stack whenever they held an odd number, and the next signal with
 * SA_ONSTACK (the Go runtime's preemption) was delivered on the stack the
 * thread was running on, which the Go runtime refuses ("non-Go code set up
 * signal handler without SA_ONSTACK flag"). perl's own catch points keep no
 * signal state either (JMPENV_PUSH is sigsetjmp(env, 0)), nor do setjmp
 * and longjmp on Linux. */
#if defined(__APPLE__)
#define GOPERL_SETJMP(jb) _setjmp(jb)
#define GOPERL_LONGJMP(jb, v) _longjmp((jb), (v))
#else
#define GOPERL_SETJMP(jb) setjmp(jb)
#define GOPERL_LONGJMP(jb, v) longjmp((jb), (v))
#endif

/* A module's own setjmp, and sigsetjmp asked to keep the signal mask, on
 * macOS on arm64. longjmp there (and siglongjmp of such a buffer) puts back
 * whether the thread runs on its alternate signal stack from bit 0 of the
 * buffer's word 47, which setjmp does not write: a jump from a buffer that
 * held an odd number there (one on a stack that held it) marked the thread
 * as on that stack while it ran on its own, and the next signal with
 * SA_ONSTACK, the Go runtime's preemption among them, ended the process.
 * So the word gets the thread's state as sigaltstack reports it before the
 * C library's setjmp runs, which is what setjmp keeps on every other
 * system: right on the thread's own stack and in a handler running on the
 * alternate one. The call stays a call of setjmp itself, so the compiler
 * treats it as one; only its argument is the buffer after the store.
 * (setjmp)(jb) is the C library's alone. */
#if defined(__APPLE__) && defined(__aarch64__)
#define GOPERL_JB_ONSTACK_WORD 47
static inline int *goperl_jb_onstack(int *jb) {
    stack_t ss;
    jb[GOPERL_JB_ONSTACK_WORD] =
        sigaltstack(NULL, &ss) == 0 && (ss.ss_flags & SS_ONSTACK) ? 1 : 0;
    return jb;
}
static inline int *goperl_jb_onstack_if(int *jb, int savemask) {
    return savemask ? goperl_jb_onstack(jb) : jb;
}
#define setjmp(jb) setjmp(goperl_jb_onstack(jb))
#define sigsetjmp(jb, savemask) \
    sigsetjmp(goperl_jb_onstack_if((jb), (savemask)), (savemask))
#endif

#define GOPERL_XS_ABI 34u

/* The width of a pointer, a size_t and an SSize_t of the interpreter, as
 * its structures hold them in its memory: it is LP64 (wasm64), like the
 * hosts the SDK supports. An address of its memory, which may be bigger
 * than 4 GiB, is handed over whole, as a 64-bit token. */
#define GOPERL_GUEST_PTRSIZE 8

/* Signal actions of the process.
 *
 * The process a module is loaded into has a handler of its own for every
 * signal, which runs on an alternate signal stack and takes the three
 * arguments of an SA_SIGINFO handler; a signal that reaches it in any other
 * way ends the process. A handler that runs on the stack of whatever the
 * thread was running corrupts that stack: the process's threads run on
 * small stacks of its runtime (libev's SIGCHLD handler, installed without
 * SA_ONSTACK, broke the stacks of the runtime's own threads when a child
 * process ended).
 *
 * A module that sets a signal aside for a moment keeps the handler it found
 * — the pointer alone — and puts that back with the flags it has for
 * handlers of its own:
 *
 *     old = signal(SIGCHLD, SIG_DFL);  ...;  signal(SIGCHLD, old);
 *
 * which is right for a process whose handlers are all like the module's.
 * So sigaction and signal, as a module calls them, keep the action each
 * signal was found with, and a handler that goes back goes back with all of
 * that action. A handler that is the module's own is installed as the
 * module asks, on the alternate stack every thread of the process has.
 *
 * They are functions of the module itself (hidden: its own calls bind to
 * them, and nothing outside it sees them), not macros: a module may #undef
 * signal and sigaction after perl.h, as EV does, and still installs its
 * handlers through them. The C library's sigaction is the next one after
 * the module (RTLD_NEXT).
 *
 * goperl_seen_action is what the module was shown, by signal: a handler
 * pointer does not tell what action it came from. */
#ifndef NSIG
#define NSIG 65
#endif
/* glibc's declarations end in __THROW: in C++ an exception specification
 * the definition has to repeat; in C attributes, which gcc does not take
 * after the declarator of a definition */
#if defined(__cplusplus) && defined(__THROW)
#define GOPERL_LIBC_THROW __THROW
#else
#define GOPERL_LIBC_THROW
#endif
GOPERL_C __attribute__((weak, visibility("hidden"))) struct sigaction goperl_seen_action[NSIG] = {0};
GOPERL_C __attribute__((weak, visibility("hidden"))) int (*goperl_libc_sigaction)(
    int, const struct sigaction *, struct sigaction *) = 0;

/* The action of the process's runtime: the handler it has for every
 * signal, SIGURG's among them, which nothing else takes. 0 when it has
 * none there. */
static int goperl_runtime_handler(struct sigaction *rt) {
    return goperl_libc_sigaction(SIGURG, NULL, rt) == 0 && (rt->sa_flags & SA_SIGINFO) &&
           rt->sa_sigaction;
}
/* Whether an action is the runtime's. */
static int goperl_runtime_action(const struct sigaction *a) {
    struct sigaction rt;
    return (a->sa_flags & SA_SIGINFO) && goperl_runtime_handler(&rt) &&
           rt.sa_sigaction == a->sa_sigaction;
}
static int goperl_program_action(int sig);

/* A signal whose action is the runtime's is the program's: its action is
 * shown as the program has it, the default (SIG_DFL) or ignored (SIG_IGN),
 * or the runtime's handler when the program has a handler of its own
 * (EV::ClickHouse ignores SIGPIPE when it finds the default action). The
 * default native code puts back is the runtime's handler again, which takes
 * the program's default action (IO::Tty sets SIGCHLD to the default around
 * grantpt and puts back what it found). */
GOPERL_C __attribute__((weak, visibility("hidden"))) int sigaction(
    int sig, const struct sigaction *act, struct sigaction *old) GOPERL_LIBC_THROW {
    struct sigaction current, with;
    if (!goperl_libc_sigaction) {
        goperl_libc_sigaction = (int (*)(int, const struct sigaction *, struct sigaction *))
            dlsym(RTLD_NEXT, "sigaction");
        if (!goperl_libc_sigaction) {
            errno = ENOSYS;
            return -1;
        }
    }
    if (sig <= 0 || sig >= NSIG)
        return goperl_libc_sigaction(sig, act, old);
    if (goperl_libc_sigaction(sig, NULL, &current) == 0 &&
        current.sa_handler != SIG_DFL && current.sa_handler != SIG_IGN)
        goperl_seen_action[sig] = current;
    if (act != NULL && act->sa_handler != SIG_IGN) {
        if (act->sa_handler == SIG_DFL) {
            if (goperl_runtime_handler(&with)) act = &with;
        } else {
            if (act->sa_handler == goperl_seen_action[sig].sa_handler) {
                with = goperl_seen_action[sig];
            } else {
                with = *act;
                with.sa_flags |= SA_ONSTACK;
            }
            act = &with;
        }
    }
    const int r = goperl_libc_sigaction(sig, act, old);
    if (r == 0 && old != NULL && goperl_runtime_action(old)) {
        const int program = goperl_program_action(sig);
        if (program == 0 || program == 1) {
            memset(old, 0, sizeof *old);
            sigemptyset(&old->sa_mask);
            old->sa_handler = program ? SIG_IGN : SIG_DFL;
        }
    }
    return r;
}

/* a handler named by a type of its own: in the declarator signal's own,
 * `void (*signal(int, void (*)(int)))(int) throw()` gives the exception
 * specification to the handler signal returns, not to signal, and g++
 * refuses that as another declaration than glibc's */
typedef void (*goperl_sighandler_t)(int);
GOPERL_C __attribute__((weak, visibility("hidden"))) goperl_sighandler_t signal(
    int sig, goperl_sighandler_t handler) GOPERL_LIBC_THROW {
    struct sigaction act, old;
    memset(&act, 0, sizeof(act));
    act.sa_handler = handler;
    sigemptyset(&act.sa_mask);
    act.sa_flags = SA_RESTART;
    if (sigaction(sig, &act, &old) != 0)
        return SIG_ERR;
    return old.sa_handler;
}

/* The perl this SDK targets (the embedded interpreter's version). */
#define PERL_PATCHLEVEL_H_IMPLICIT
#include "patchlevel.h"
#undef PERL_PATCHLEVEL_H_IMPLICIT
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

/* What the C environment has, as perl's config.h says it. */
#include "config.h"

/* A dist's bundled ppport.h is a portability layer for REAL perl headers;
 * against this SDK it must not activate. Its include guard is pre-defined
 * so the #include turns into a no-op. */
#define _P_P_PORTABILITY_H_

/* Ancient K&R-compat prototype wrapper old dist headers still use
 * (DBI's DBIXS.h: `void (*fn) _((args));`). ppport used to supply it. */
#ifndef _
#define _(args) args
#endif
/* More of what perl.h keeps for modules written for old compilers (Tk's
 * Event.xs picks the ANSI form of a variadic function by STANDARD_C). */
#define CAN_PROTOTYPE
#define STANDARD_C

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
/* HE and HEK are perl's own (hv.h), laid out as the interpreter has them:
 * an entry the interpreter hands out (hv_fetch_ent, hv_store_ent,
 * hv_iternext, HvEITER, HvARRAY) is the entry itself, where it lies in the
 * interpreter's memory, so HeVAL reads and writes the hash's value slot
 * (a value is an SV token, the address the interpreter has for it) and
 * an entry keeps its identity from one call to the next. Modules also make
 * entries of their own (Cpanel::JSON::XS, for the entries of a tied hash,
 * in a string buffer: HeSVKEY_set, hent_hek assigned) and read either kind
 * through the same macros.
 *
 * The pointers an entry of the interpreter holds are addresses of its
 * memory. hent_next is a struct of one member, so that native code that
 * follows it as a pointer (he->hent_next->...) does not compile; HeNEXT
 * gives the next entry. hent_hek keeps its pointer type, for the modules
 * that store a pointer of this process there (Cpanel::JSON::XS); HeKEY_hek
 * tells the two kinds apart (goperl_he_ptr). The loader compares this
 * layout with the interpreter's when it loads a module (__goperl_xs_layout)
 * and refuses a module that disagrees. */
typedef struct hek {
    U32 hek_hash;
    I32 hek_len;
    char hek_key[1];
} HEK;
#define cBOOL(x) ((bool)!!(x))
#define DO_UTF8(sv) SvUTF8(sv)
typedef struct goperl_he HE;
typedef struct goperl_he_link {
    uint64_t goperl_addr; /* an address of the interpreter's memory, or 0 */
} goperl_he_link_t;
struct goperl_he {
    goperl_he_link_t hent_next;
    HEK *hent_hek;
    union {
        SV *hent_val;
        size_t hent_refcount;
    } he_valu;
};

/* svtype: the REAL perl 5.44 enum values, and the guest reports raw
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

/* Context / call flags (real perl 5.44 values; they cross to the guest). */
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

/* overload (amagic): the index of every operator a class can overload, in
 * the order of perl 5.44's overload.h (fallback_amg is 0, regexp_amg 0x4a).
 * The interpreter takes them as they are (amagic_call). */
enum {
    fallback_amg, to_sv_amg, to_av_amg, to_hv_amg, to_gv_amg, to_cv_amg,
    inc_amg, dec_amg, bool__amg, numer_amg, string_amg, not_amg, copy_amg,
    abs_amg, neg_amg, iter_amg, int_amg, lt_amg, le_amg, gt_amg, ge_amg,
    eq_amg, ne_amg, slt_amg, sle_amg, sgt_amg, sge_amg, seq_amg, sne_amg,
    nomethod_amg, add_amg, add_ass_amg, subtr_amg, subtr_ass_amg, mult_amg,
    mult_ass_amg, div_amg, div_ass_amg, modulo_amg, modulo_ass_amg,
    pow_amg, pow_ass_amg, lshift_amg, lshift_ass_amg, rshift_amg,
    rshift_ass_amg, band_amg, band_ass_amg, sband_amg, sband_ass_amg,
    bor_amg, bor_ass_amg, sbor_amg, sbor_ass_amg, bxor_amg, bxor_ass_amg,
    sbxor_amg, sbxor_ass_amg, ncmp_amg, scmp_amg, compl_amg, scompl_amg,
    atan2_amg, cos_amg, sin_amg, exp_amg, log_amg, sqrt_amg, repeat_amg,
    repeat_ass_amg, concat_amg, concat_ass_amg, smart_amg, ftest_amg,
    regexp_amg, max_amg_code
};
#define AMGf_noright 1
#define AMGf_noleft 2
#define AMGf_assign 4 /* the operator has a mutator form ($x .= ...) */
#define AMGf_unary 8
#define AMGf_numeric 0x0010
#define AMGf_want_list 0x0040
#define AMGf_numarg 0x0080
#define AMGf_force_scalar 0x0100

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

/* Context types (cop.h, real perl 5.44 values). */
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

/* Debugger flags ($^P) and exit flags (real perl 5.44 values). */
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

/* Flag and type constants are the interpreter's own (SvFLAGS returns the
 * real flag word, and flag arguments cross to the interpreter unchanged). */
#include "goperl_consts.h"

/* The SV_INFO guest op answers the common predicates (SvOK, SvPOK, ...) in
 * one round trip, as this private bitset. */
#define GOPERL_INFO_OK 1
#define GOPERL_INFO_ROK 2
#define GOPERL_INFO_POK 4
#define GOPERL_INFO_NIOK 8
#define GOPERL_INFO_UTF8 16
#define GOPERL_INFO_ISOBJ 32
#define GOPERL_INFO_READONLY 64
#define GOPERL_INFO_ISCV 128
#define GOPERL_INFO_ISGV 256
#define GOPERL_INFO_IOK 512
#define GOPERL_INFO_NOK 1024
#define GOPERL_INFO_POKp 2048
#define GOPERL_INFO_NIOKp 4096
#define GOPERL_INFO_ISUV 8192
#define GOPERL_INFO_RMAGICAL 16384
#define GOPERL_INFO_OBJECT 32768
#define GOPERL_INFO_AMAGIC 65536
#define GOPERL_INFO_IOKp 131072
#define GOPERL_INFO_NOKp 262144

/* ---- host-side MAGIC mirror (layout is ABI: the loader writes it) ------- */

struct goperl_frame;
typedef struct goperl_frame goperl_frame_t;

/* The interpreter's context is that of a perl built without MULTIPLICITY,
 * as perl's default build is: functions take no context argument (pTHX is
 * void, aTHX is nothing), and a module compiles as it does against such a
 * perl, also where it picks its declarations by MULTIPLICITY or
 * PERL_IMPLICIT_CONTEXT (SWIG's wrappers, qsort comparators).
 *
 * What a native call works on is its frame, which the loader's entry
 * points make current before they call a module's function: the current
 * frame lives in goperl_cur_frame_v (per thread), maintained at every entry
 * and exit, and `_gof`, the name every SDK macro reads, is an alias for it.
 * So the interpreter's macros work in any function, as they do in perl
 * without MULTIPLICITY. */
#define pTHX void
#define pTHX_
#define aTHX
#define aTHX_
#define dTHX dNOOP
#define dTHXa(a) dNOOP
/* the context of the system interface, which a perl without
 * PERL_IMPLICIT_SYS does not have (Tk's Event.xs declares it) */
#define dTHXs dNOOP
#define aTHXa(a) NOOP
#define _gof goperl_cur_frame_v
#define pTHX_1 1
#define pTHX_2 2
#define pTHX_3 3
#define pTHX_4 4
#define pTHX_5 5
#define pTHX_6 6
#define pTHX_7 7
#define pTHX_8 8
#define pTHX_9 9
#define pTHX_10 10
#define pTHX_11 11
#define pTHX_12 12
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
/* An item of an op's auxiliary data. The items of an op of the
 * interpreter's (a multideref's) are its memory, where an item is as wide
 * as here and a pad offset is too (both are LP64); a pointer there is an
 * address of the interpreter's, which UNOP_AUX_item_sv reads an SV from. */
typedef union {
    SV *sv;
    UV uv;
    IV iv;
    char *pv;
    SSize_t ssize;
    UV pad_offset;
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
    U8 gop_class;  /* the interpreter's class of the op, plus one */ \
    U8 op_opt;      /* the flag bits of the interpreter's op, to read */ \
    U8 op_slabbed; \
    U8 op_savefree; \
    U8 op_static; \
    U8 op_folded; \
    uint64_t gop; /* guest OP token (0 = module-built scratch op) */
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
    OP *op_nextop; /* LOOP view */
    OP *op_lastop;
    U32 cop_seq;
    U32 op_aux_n; /* the items op_aux has */
    struct cop_feature_t {
        U32 bits[1]; /* COP_FEATURE_SIZE of feature.h */
    } cop_features;
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
#define cUNOP_AUXo cUNOP_AUXx(o)
#define cUNOP_AUX cUNOP_AUXx(PL_op)
#define OpSIBLING(o) goperl_op_sibling(_gof, (OP *)(o))
#define OpHAS_SIBLING(o) (OpSIBLING(o) != 0)
#define OpMORESIB_set(o, sib) \
    ({ OP *const goperl_OpMORESIB_set_o_ = (OP *)(o); \
       goperl_OpMORESIB_set_o_->op_moresib = 1; \
       goperl_OpMORESIB_set_o_->op_sibparent = (OP *)(sib); })
#define OpLASTSIB_set(o, parent) \
    ({ OP *const goperl_OpLASTSIB_set_o_ = (OP *)(o); \
       goperl_OpLASTSIB_set_o_->op_moresib = 0; \
       goperl_OpLASTSIB_set_o_->op_sibparent = (OP *)(parent); })
#define OpMAYBESIB_set(o, sib, parent) \
    ({ OP *const goperl_OpMAYBESIB_set_o_ = (OP *)(o); \
       OP *const goperl_OpMAYBESIB_set_sib_ = (OP *)(sib); \
       goperl_OpMAYBESIB_set_o_->op_sibparent = \
           goperl_OpMAYBESIB_set_sib_ ? goperl_OpMAYBESIB_set_sib_ : (OP *)(parent); \
       goperl_OpMAYBESIB_set_o_->op_moresib = goperl_OpMAYBESIB_set_sib_ ? 1 : 0; })
/* The fields of a COP are read through goperl_cop_fields, which fills them
 * for a shadow at the edge of what the interpreter described (see
 * goperl_op_apply): a link of another op, such as the op_next a pp
 * function returns, has its type and none of its COP fields yet. */
static COP *goperl_cop_fields(COP *c);
#define CopLINE(c) (goperl_cop_fields((COP *)(c))->cop_line)
#define CopFILE(c) (goperl_cop_fields((COP *)(c))->cop_file)
#define OutCopFILE(c) CopFILE(c)
#define CopSTASHPV(c) ((char *)goperl_cop_fields((COP *)(c))->cop_stashpv)

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
                      uint64_t c, const char *s, uint64_t slen);
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
    /* the descriptor of the host process behind a descriptor of the
     * interpreter, or -1 when it has none */
    int64_t (*host_fd)(goperl_frame_t *, int64_t guest_fd);
    /* `CvXSUB(cv) = fn`: the CV calls the native function fn from now on */
    void (*cv_set_xsub)(goperl_frame_t *, uint64_t cv, void *fn);
    /* What is kept for an SV whose string-buffer slot a module gave a
     * pointer of this process's (see GOPERL_PV_HOST): two words, the
     * pointer and the room the module says the buffer has (SvLEN). Made
     * on request, NULL when the SV has none. */
    void **(*pv_host)(goperl_frame_t *, uint64_t sv, int32_t make);
    /* v27: run the operation whose arguments are in the frame's operation
     * box (see goperl_xs_op); 1 when it returned (the result is in the
     * box), 0 when it failed (the frame says how). */
    uint64_t (*xs_op_box)(goperl_frame_t *);
} goperl_api_t;
#define GOPERL_SHARED_RESERVED 262144

/* ---- one XSUB activation (layout is ABI, mirrored by the loader) -------- */

#define GOPERL_XS_STACK 512
#define GOPERL_XS_MARKS 64
#define GOPERL_XS_TMPS 16

/* Slots of SV** results beyond the frame's own (see goperl_tmp_slot): a
 * block of them, after the one made before it. */
#define GOPERL_TMP_BLOCK 256
typedef struct goperl_tmpblock {
    struct goperl_tmpblock *prev;
    uint64_t slot[GOPERL_TMP_BLOCK];
} goperl_tmpblock_t;

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
    int32_t tmpidx;          /* +52  the SV** result slots handed out */
    uint64_t hook_val[2];    /* +56  pending hook assignments */
    uint64_t imm[3];         /* +72  cached &PL_sv_undef/yes/no tokens */
    OP *plop;                /* +96  PL_op host shadow */
    uint64_t hook_op_tok;    /* +104 pp-hook frames: the op being executed */
    void *prev_frame;        /* +112 frame nesting (goperl_cur_frame_v) */
    void *inst;              /* +120 this INSTANCE's shared-state block */
    SV **stbase;             /* +128 the perl stack: st below, or a heap
                              * block once it outgrew st (the loader frees
                              * it when the frame ends) */
    int64_t stcap;           /* +136 its capacity in slots */
    uint64_t st[GOPERL_XS_STACK];   /* +144 the initial stack (tokens) */
    int32_t marks[GOPERL_XS_MARKS]; /* mark offsets into st */
    uint64_t tmp[GOPERL_XS_TMPS];   /* the first slots of SV** results */
    char err[512];
    int32_t xerrno; /* 0, or 1 + the errno the XSUB left, in the
                     * interpreter's numbering (see goperl_errno_leave) */
    char *membase; /* the interpreter's linear memory: where it starts in
                    * this process (it never moves) */
    const uint64_t *memsize; /* and where the loader keeps its size, which
                              * grows while the interpreter runs, also
                              * during a native call: read it at every use.
                              * The range reserved past the size holds
                              * nothing of this process's and is not
                              * readable. */
    /* The perl stack of a pp function (see goperl_pp_stack_pull): st and
     * marks are a copy of the interpreter's stack and mark stack. */
    int32_t pp_frame;  /* the frame runs a pp function */
    int32_t st_stale;  /* the interpreter's stack has moved on since the copy */
    int32_t gmark_lo;  /* the interpreter's index of marks[0] */
    int32_t mark_seen; /* markidx when the copy was last in line */
    uint64_t *st_seen; /* the values then (index 0 unused) */
    int64_t st_seen_n; /* and their number */
    int32_t marks_seen[GOPERL_XS_MARKS];
    char *pp_state;    /* what the function made of the stack, for the loader */
    int64_t pp_state_len;
    int32_t native_call; /* the frame of a native call the interpreter made
                          * (an XSUB, a pp function, a hook), not of a
                          * function the loader runs on the side */
    int32_t pad5;
    void *entry_jb; /* the catch point of the loader's entry while no XSUB
                     * has taken that activation over (dXSARGS); the frame
                     * of a module built before the entry existed ends at
                     * pad5 for that module, which never looks here */
    uint64_t opbox; /* the address of the instance's operation box in its
                     * memory, 0 when it has none (see goperl_xs_op) */
    goperl_tmpblock_t *tmp_more; /* the last block of SV** result slots
                                  * beyond tmp (malloc'd; the loader frees
                                  * the blocks when the frame ends) */
};

/* ---- shared module state (weak: one copy across a multi-TU .so) ---------
 * The module's own globals are hidden: a module loaded with its symbols
 * global (dl_load_flags 0x01, for modules that link against the functions
 * it exports) must not lend them to the modules loaded after it, whose
 * references would otherwise bind to its copies. */

GOPERL_C __attribute__((weak, visibility("hidden"))) const goperl_api_t *goperl_api_v = 0;

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
    /* set in a child of fork(2) native code made (see fork below): the
     * block is the child's copy, so only the child sees it */
    int32_t forked;
} goperl_process_t;
#define GOPERL_PROCESS ((goperl_process_t *)goperl_api_v->shared_raw)

/* ---- fork(2) in native code ----------------------------------------------
 * The interpreter runs inside a Go program, whose runtime keeps threads of
 * its own; fork(2) copies the process with only the calling thread, so in
 * the child the runtime, and with it the interpreter, cannot run (it waits
 * for threads that are not there, uses descriptors the child closes). A
 * child that execs, or does its work in C alone and exits, is fine. A
 * child that calls the interpreter (perl API functions, a callback into
 * Perl code) ends instead, at its first such call, with a message: perl
 * would run that child, gperl cannot. fork is the module's own (hidden), as
 * sigaction is; it marks the child's copy of the process block. */
GOPERL_C __attribute__((weak, visibility("hidden"))) pid_t (*goperl_libc_fork)(void) = 0;
GOPERL_C __attribute__((weak, visibility("hidden"))) pid_t fork(void) GOPERL_LIBC_THROW {
    if (!goperl_libc_fork) {
        goperl_libc_fork = (pid_t (*)(void))dlsym(RTLD_NEXT, "fork");
        if (!goperl_libc_fork) {
            errno = ENOSYS;
            return -1;
        }
    }
    const pid_t pid = goperl_libc_fork();
    if (pid == 0 && goperl_api_v) GOPERL_PROCESS->forked = 1;
    return pid;
}
static void goperl_forked_exit(void) __attribute__((noreturn, unused));
static void goperl_forked_exit(void) {
    static const char msg[] =
        "native code forked this process with fork(2), and its child called "
        "the interpreter: a child of fork(2) made by native code cannot run "
        "Perl code under gperl (the Go runtime does not survive fork); the "
        "child exits\n";
    ssize_t w = write(2, msg, sizeof msg - 1);
    (void)w;
    _exit(255);
}
/* Ends a child of native code's fork(2) that calls the interpreter. */
#define goperl_forked_check()                                         \
    do {                                                              \
        if (goperl_api_v && GOPERL_PROCESS->forked) goperl_forked_exit(); \
    } while (0)

/* ---- errno ------------------------------------------------------------------
 * The host's errno is native code's own: a call into the loader (which runs
 * the Go runtime on this thread, and the interpreter with its separate errno)
 * leaves it as it was, as a call of a perl API function leaves it alone
 * unless it fails. What an XSUB leaves in it when it returns becomes the
 * interpreter's errno, so that `$!` after the call says what the C code set:
 * the XSUB starts with a value no C code sets (GOPERL_ERRNO_UNSET), and
 * anything else found at its end was put there by it. */
#define GOPERL_ERRNO_UNSET (-0x676f70)
#define GOPERL_KEEP_ERRNO(type, call) \
    do {                              \
        const int goperl_e_ = errno;  \
        type goperl_r_ = (call);      \
        errno = goperl_e_;            \
        return goperl_r_;             \
    } while (0)
#define GOPERL_KEEP_ERRNO_VOID(call) \
    do {                             \
        const int goperl_e_ = errno; \
        (call);                      \
        errno = goperl_e_;           \
    } while (0)
/* The operation box of an instance (perl-wasm's struct goperl_opbox): an
 * operation's operands, written here into the interpreter's memory, and
 * its result, which the interpreter leaves there, with the second value
 * of an operation that answers two (result2). An operation goes through it
 * with nothing encoded or allocated on the way; a string longer than the
 * box holds goes through the loader's encoding instead, and that operation
 * leaves its second value in the box as well.
 *
 * An operation takes three operands, a, b and c, and a string; a token (an
 * address of the interpreter's memory) is always an operand of its own,
 * whole. The selector of the operations that have them (SV_CALL, GV_CALL,
 * SV_STR, CALL, INTERNAL, FIELD, REGEXP, PERLIO, CX, SV_SLOT_SET/GET) is c,
 * with a 32-bit value that goes with b in c's high half. */
#define GOPERL_OPBOX_STR 4096
typedef struct goperl_opbox {
    uint64_t a;
    uint64_t b;
    uint64_t c;
    uint64_t result;
    uint64_t result2;
    int32_t op;
    uint32_t slen;
    char s[GOPERL_OPBOX_STR];
} goperl_opbox_t;
typedef char goperl_opbox_layout[(__builtin_offsetof(goperl_opbox_t, c) == 16 &&
                                  __builtin_offsetof(goperl_opbox_t, result) == 24 &&
                                  __builtin_offsetof(goperl_opbox_t, result2) == 32 &&
                                  __builtin_offsetof(goperl_opbox_t, op) == 40 &&
                                  __builtin_offsetof(goperl_opbox_t, slen) == 44 &&
                                  __builtin_offsetof(goperl_opbox_t, s) == 48)
                                     ? 1
                                     : -1] __attribute__((unused));
static inline __attribute__((unused)) uint64_t goperl_xs_op_box(
    goperl_frame_t *f, int32_t op, uint64_t a, uint64_t b, uint64_t c,
    const char *s, uint32_t slen) {
    goperl_opbox_t *const box = (goperl_opbox_t *)(f->membase + f->opbox);
    box->a = a;
    box->b = b;
    box->c = c;
    box->op = op;
    box->slen = slen;
    if (slen) memcpy(box->s, s, slen);
    return goperl_api_v->xs_op_box(f) ? box->result : 0;
}
/* One operation; with second, the second value of an operation that
 * answers two is stored there (0 when the operation failed). */
static void goperl_pending_block_string(goperl_frame_t *f, int32_t op, uint64_t a, uint64_t b);
static void *goperl_pending_blocks_any(void);
static inline __attribute__((unused)) uint64_t goperl_xs_opx(
    goperl_frame_t *f, int32_t op, uint64_t a, uint64_t b, uint64_t c,
    const char *s, uint64_t slen, uint64_t *second) {
    goperl_forked_check();
    if (goperl_pending_blocks_any()) goperl_pending_block_string(f, op, a, b);
    const int goperl_e_ = errno;
    const int32_t failed_before = f->failed;
    const uint64_t r = slen <= GOPERL_OPBOX_STR
                           ? goperl_xs_op_box(f, op, a, b, c, s, (uint32_t)slen)
                           : goperl_api_v->xs_op(f, op, a, b, c, s, slen);
    if (second)
        *second = f->failed && !failed_before
                      ? 0
                      : ((const goperl_opbox_t *)(f->membase + f->opbox))->result2;
    errno = goperl_e_;
    return r;
}
static inline __attribute__((unused)) uint64_t goperl_xs_op(
    goperl_frame_t *f, int32_t op, uint64_t a, uint64_t b, const char *s,
    uint64_t slen) {
    return goperl_xs_opx(f, op, a, b, 0, s, slen, 0);
}
/* Where an address of the interpreter's memory is in this process. No
 * address (0) is none here either: the interpreter answers a question
 * about something that does not exist with 0. The memory never moves, so
 * an address below its size as it is now is turned here; the loader turns
 * any other, and fails the frame when the memory does not hold it (a
 * stray address of native code's: the pages reserved past the size are
 * not readable). The failure unwinds to the nearest catch point, as a
 * perl API function's does (goperl_do_opx), so native code never follows
 * the null pointer the call answers then; without a catch point armed the
 * caller gets the null pointer and the loader reports the failure. */
static void goperl_raise(goperl_frame_t *f) __attribute__((noreturn, unused));
static inline __attribute__((unused)) void *goperl_guest_mem(
    goperl_frame_t *f, uint64_t gptr) {
    if (!gptr) return 0;
    if (f->membase && gptr < *f->memsize) return f->membase + gptr;
    {
        const int goperl_e_ = errno;
        const int32_t failed_before = f->failed;
        void *const p = goperl_api_v->guest_mem(f, gptr);
        errno = goperl_e_;
        if (!failed_before && f->failed && f->jb) goperl_raise(f);
        return p;
    }
}
/* The interpreter's memory, read and written where it is: a word of four
 * bytes at an address of it (a U32, an I32, an int), a word of eight bytes
 * (a pointer stored there, a size_t, an SSize_t: GOPERL_GUEST_PTRSIZE), and
 * a byte. An address read as a word of eight bytes is a token. */
typedef char goperl_guest_ptrsize_is_8[GOPERL_GUEST_PTRSIZE == 8 ? 1 : -1] __attribute__((unused));
static inline __attribute__((unused)) uint32_t goperl_g32(const goperl_frame_t *f, uint64_t g) {
    uint32_t v;
    memcpy(&v, f->membase + g, 4);
    return v;
}
static inline __attribute__((unused)) void goperl_g32_set(goperl_frame_t *f, uint64_t g, uint32_t v) {
    memcpy(f->membase + g, &v, 4);
}
static inline __attribute__((unused)) uint64_t goperl_g64(const goperl_frame_t *f, uint64_t g) {
    uint64_t v;
    memcpy(&v, f->membase + g, 8);
    return v;
}
static inline __attribute__((unused)) void goperl_g64_set(goperl_frame_t *f, uint64_t g, uint64_t v) {
    memcpy(f->membase + g, &v, 8);
}
static inline __attribute__((unused)) uint64_t goperl_api_new_xs(
    goperl_frame_t *f, const char *name, void *xsub) {
    GOPERL_KEEP_ERRNO(uint64_t, goperl_api_v->new_xs(f, name, xsub));
}
static inline __attribute__((unused)) void *goperl_api_cv_any(
    goperl_frame_t *f, uint64_t cv) {
    GOPERL_KEEP_ERRNO(void *, goperl_api_v->cv_any(f, cv));
}
static inline __attribute__((unused)) void *goperl_api_cv_xsub(
    goperl_frame_t *f, uint64_t cv) {
    GOPERL_KEEP_ERRNO(void *, goperl_api_v->cv_xsub(f, cv));
}
static inline __attribute__((unused)) MAGIC *goperl_api_magic_ext(
    goperl_frame_t *f, uint64_t sv, uint64_t obj, int32_t how,
    const void *vtbl, const char *ptr, int64_t len) {
    GOPERL_KEEP_ERRNO(MAGIC *,
                      goperl_api_v->magic_ext(f, sv, obj, how, vtbl, ptr, len));
}
static inline __attribute__((unused)) MAGIC *goperl_api_magic_chain(
    goperl_frame_t *f, uint64_t sv) {
    GOPERL_KEEP_ERRNO(MAGIC *, goperl_api_v->magic_chain(f, sv));
}
static inline __attribute__((unused)) void goperl_api_magic_del(
    goperl_frame_t *f, uint64_t sv, int32_t how, const void *vtbl) {
    GOPERL_KEEP_ERRNO_VOID(goperl_api_v->magic_del(f, sv, how, vtbl));
}
static inline __attribute__((unused)) void goperl_api_magic_set_head(
    goperl_frame_t *f, uint64_t sv, MAGIC *mg) {
    GOPERL_KEEP_ERRNO_VOID(goperl_api_v->magic_set_head(f, sv, mg));
}
static inline __attribute__((unused)) void goperl_api_pp_hook_set(
    goperl_frame_t *f, int32_t op_type, void *fn) {
    GOPERL_KEEP_ERRNO_VOID(goperl_api_v->pp_hook_set(f, op_type, fn));
}
static inline __attribute__((unused)) void goperl_api_perop_pp_set(
    goperl_frame_t *f, uint64_t optok, void *fn) {
    GOPERL_KEEP_ERRNO_VOID(goperl_api_v->perop_pp_set(f, optok, fn));
}
static inline __attribute__((unused)) uint32_t goperl_api_perlio_def(
    goperl_frame_t *f, const char *name, void *funcs) {
    GOPERL_KEEP_ERRNO(uint32_t, goperl_api_v->perlio_def(f, name, funcs));
}
static inline __attribute__((unused)) int64_t goperl_api_host_fd(
    goperl_frame_t *f, int64_t guest_fd) {
    GOPERL_KEEP_ERRNO(int64_t, goperl_api_v->host_fd(f, guest_fd));
}
static inline __attribute__((unused)) void goperl_api_cv_set_xsub(
    goperl_frame_t *f, uint64_t cv, void *fn) {
    GOPERL_KEEP_ERRNO_VOID(goperl_api_v->cv_set_xsub(f, cv, fn));
}
/* The interpreter's number for a host errno value. Its C library has a
 * number for every POSIX name; a host value it has no name for (a few are
 * specific to one system) is carried marked: the bit 0x4000 over the
 * host's number, which a program that sees the host's numbers gets back as
 * that number and any other program as EIO. */
static int goperl_errno_guest(int e) __attribute__((unused));
static int goperl_errno_guest(int e) {
    if (e == 0) return 0;
#ifdef E2BIG
    if (e == E2BIG) return 1;
#endif
#ifdef EACCES
    if (e == EACCES) return 2;
#endif
#ifdef EADDRINUSE
    if (e == EADDRINUSE) return 3;
#endif
#ifdef EADDRNOTAVAIL
    if (e == EADDRNOTAVAIL) return 4;
#endif
#ifdef EAFNOSUPPORT
    if (e == EAFNOSUPPORT) return 5;
#endif
#ifdef EAGAIN
    if (e == EAGAIN) return 6;
#endif
#ifdef EALREADY
    if (e == EALREADY) return 7;
#endif
#ifdef EBADF
    if (e == EBADF) return 8;
#endif
#ifdef EBADMSG
    if (e == EBADMSG) return 9;
#endif
#ifdef EBUSY
    if (e == EBUSY) return 10;
#endif
#ifdef ECANCELED
    if (e == ECANCELED) return 11;
#endif
#ifdef ECHILD
    if (e == ECHILD) return 12;
#endif
#ifdef ECONNABORTED
    if (e == ECONNABORTED) return 13;
#endif
#ifdef ECONNREFUSED
    if (e == ECONNREFUSED) return 14;
#endif
#ifdef ECONNRESET
    if (e == ECONNRESET) return 15;
#endif
#ifdef EDEADLK
    if (e == EDEADLK) return 16;
#endif
#ifdef EDESTADDRREQ
    if (e == EDESTADDRREQ) return 17;
#endif
#ifdef EDOM
    if (e == EDOM) return 18;
#endif
#ifdef EDQUOT
    if (e == EDQUOT) return 19;
#endif
#ifdef EEXIST
    if (e == EEXIST) return 20;
#endif
#ifdef EFAULT
    if (e == EFAULT) return 21;
#endif
#ifdef EFBIG
    if (e == EFBIG) return 22;
#endif
#ifdef EHOSTUNREACH
    if (e == EHOSTUNREACH) return 23;
#endif
#ifdef EIDRM
    if (e == EIDRM) return 24;
#endif
#ifdef EILSEQ
    if (e == EILSEQ) return 25;
#endif
#ifdef EINPROGRESS
    if (e == EINPROGRESS) return 26;
#endif
#ifdef EINTR
    if (e == EINTR) return 27;
#endif
#ifdef EINVAL
    if (e == EINVAL) return 28;
#endif
#ifdef EIO
    if (e == EIO) return 29;
#endif
#ifdef EISCONN
    if (e == EISCONN) return 30;
#endif
#ifdef EISDIR
    if (e == EISDIR) return 31;
#endif
#ifdef ELOOP
    if (e == ELOOP) return 32;
#endif
#ifdef EMFILE
    if (e == EMFILE) return 33;
#endif
#ifdef EMLINK
    if (e == EMLINK) return 34;
#endif
#ifdef EMSGSIZE
    if (e == EMSGSIZE) return 35;
#endif
#ifdef EMULTIHOP
    if (e == EMULTIHOP) return 36;
#endif
#ifdef ENAMETOOLONG
    if (e == ENAMETOOLONG) return 37;
#endif
#ifdef ENETDOWN
    if (e == ENETDOWN) return 38;
#endif
#ifdef ENETRESET
    if (e == ENETRESET) return 39;
#endif
#ifdef ENETUNREACH
    if (e == ENETUNREACH) return 40;
#endif
#ifdef ENFILE
    if (e == ENFILE) return 41;
#endif
#ifdef ENOBUFS
    if (e == ENOBUFS) return 42;
#endif
#ifdef ENODEV
    if (e == ENODEV) return 43;
#endif
#ifdef ENOENT
    if (e == ENOENT) return 44;
#endif
#ifdef ENOEXEC
    if (e == ENOEXEC) return 45;
#endif
#ifdef ENOLCK
    if (e == ENOLCK) return 46;
#endif
#ifdef ENOLINK
    if (e == ENOLINK) return 47;
#endif
#ifdef ENOMEM
    if (e == ENOMEM) return 48;
#endif
#ifdef ENOMSG
    if (e == ENOMSG) return 49;
#endif
#ifdef ENOPROTOOPT
    if (e == ENOPROTOOPT) return 50;
#endif
#ifdef ENOSPC
    if (e == ENOSPC) return 51;
#endif
#ifdef ENOSYS
    if (e == ENOSYS) return 52;
#endif
#ifdef ENOTCONN
    if (e == ENOTCONN) return 53;
#endif
#ifdef ENOTDIR
    if (e == ENOTDIR) return 54;
#endif
#ifdef ENOTEMPTY
    if (e == ENOTEMPTY) return 55;
#endif
#ifdef ENOTRECOVERABLE
    if (e == ENOTRECOVERABLE) return 56;
#endif
#ifdef ENOTSOCK
    if (e == ENOTSOCK) return 57;
#endif
#ifdef ENOTSUP
    if (e == ENOTSUP) return 58;
#endif
#ifdef ENOTTY
    if (e == ENOTTY) return 59;
#endif
#ifdef ENXIO
    if (e == ENXIO) return 60;
#endif
#ifdef EOVERFLOW
    if (e == EOVERFLOW) return 61;
#endif
#ifdef EOWNERDEAD
    if (e == EOWNERDEAD) return 62;
#endif
#ifdef EPERM
    if (e == EPERM) return 63;
#endif
#ifdef EPIPE
    if (e == EPIPE) return 64;
#endif
#ifdef EPROTO
    if (e == EPROTO) return 65;
#endif
#ifdef EPROTONOSUPPORT
    if (e == EPROTONOSUPPORT) return 66;
#endif
#ifdef EPROTOTYPE
    if (e == EPROTOTYPE) return 67;
#endif
#ifdef ERANGE
    if (e == ERANGE) return 68;
#endif
#ifdef EROFS
    if (e == EROFS) return 69;
#endif
#ifdef ESPIPE
    if (e == ESPIPE) return 70;
#endif
#ifdef ESRCH
    if (e == ESRCH) return 71;
#endif
#ifdef ESTALE
    if (e == ESTALE) return 72;
#endif
#ifdef ETIMEDOUT
    if (e == ETIMEDOUT) return 73;
#endif
#ifdef ETXTBSY
    if (e == ETXTBSY) return 74;
#endif
#ifdef EXDEV
    if (e == EXDEV) return 75;
#endif
#ifdef ENOTCAPABLE
    if (e == ENOTCAPABLE) return 76;
#endif
    if (e > 0 && e < 0x4000) return 0x4000 | e;
    return 29; /* EIO */
}
/* The host's errno value for a number of the interpreter's. */
static int goperl_errno_host(int e) __attribute__((unused));
static int goperl_errno_host(int e) {
    /* an error of the host's that the interpreter carries marked */
    if (e >= 0x4000 && e < 0x8000) return e - 0x4000;
    switch (e) {
    case 0: return 0;
#ifdef E2BIG
    case 1: return E2BIG;
#endif
#ifdef EACCES
    case 2: return EACCES;
#endif
#ifdef EADDRINUSE
    case 3: return EADDRINUSE;
#endif
#ifdef EADDRNOTAVAIL
    case 4: return EADDRNOTAVAIL;
#endif
#ifdef EAFNOSUPPORT
    case 5: return EAFNOSUPPORT;
#endif
#ifdef EAGAIN
    case 6: return EAGAIN;
#endif
#ifdef EALREADY
    case 7: return EALREADY;
#endif
#ifdef EBADF
    case 8: return EBADF;
#endif
#ifdef EBADMSG
    case 9: return EBADMSG;
#endif
#ifdef EBUSY
    case 10: return EBUSY;
#endif
#ifdef ECANCELED
    case 11: return ECANCELED;
#endif
#ifdef ECHILD
    case 12: return ECHILD;
#endif
#ifdef ECONNABORTED
    case 13: return ECONNABORTED;
#endif
#ifdef ECONNREFUSED
    case 14: return ECONNREFUSED;
#endif
#ifdef ECONNRESET
    case 15: return ECONNRESET;
#endif
#ifdef EDEADLK
    case 16: return EDEADLK;
#endif
#ifdef EDESTADDRREQ
    case 17: return EDESTADDRREQ;
#endif
#ifdef EDOM
    case 18: return EDOM;
#endif
#ifdef EDQUOT
    case 19: return EDQUOT;
#endif
#ifdef EEXIST
    case 20: return EEXIST;
#endif
#ifdef EFAULT
    case 21: return EFAULT;
#endif
#ifdef EFBIG
    case 22: return EFBIG;
#endif
#ifdef EHOSTUNREACH
    case 23: return EHOSTUNREACH;
#endif
#ifdef EIDRM
    case 24: return EIDRM;
#endif
#ifdef EILSEQ
    case 25: return EILSEQ;
#endif
#ifdef EINPROGRESS
    case 26: return EINPROGRESS;
#endif
#ifdef EINTR
    case 27: return EINTR;
#endif
#ifdef EINVAL
    case 28: return EINVAL;
#endif
#ifdef EIO
    case 29: return EIO;
#endif
#ifdef EISCONN
    case 30: return EISCONN;
#endif
#ifdef EISDIR
    case 31: return EISDIR;
#endif
#ifdef ELOOP
    case 32: return ELOOP;
#endif
#ifdef EMFILE
    case 33: return EMFILE;
#endif
#ifdef EMLINK
    case 34: return EMLINK;
#endif
#ifdef EMSGSIZE
    case 35: return EMSGSIZE;
#endif
#ifdef EMULTIHOP
    case 36: return EMULTIHOP;
#endif
#ifdef ENAMETOOLONG
    case 37: return ENAMETOOLONG;
#endif
#ifdef ENETDOWN
    case 38: return ENETDOWN;
#endif
#ifdef ENETRESET
    case 39: return ENETRESET;
#endif
#ifdef ENETUNREACH
    case 40: return ENETUNREACH;
#endif
#ifdef ENFILE
    case 41: return ENFILE;
#endif
#ifdef ENOBUFS
    case 42: return ENOBUFS;
#endif
#ifdef ENODEV
    case 43: return ENODEV;
#endif
#ifdef ENOENT
    case 44: return ENOENT;
#endif
#ifdef ENOEXEC
    case 45: return ENOEXEC;
#endif
#ifdef ENOLCK
    case 46: return ENOLCK;
#endif
#ifdef ENOLINK
    case 47: return ENOLINK;
#endif
#ifdef ENOMEM
    case 48: return ENOMEM;
#endif
#ifdef ENOMSG
    case 49: return ENOMSG;
#endif
#ifdef ENOPROTOOPT
    case 50: return ENOPROTOOPT;
#endif
#ifdef ENOSPC
    case 51: return ENOSPC;
#endif
#ifdef ENOSYS
    case 52: return ENOSYS;
#endif
#ifdef ENOTCONN
    case 53: return ENOTCONN;
#endif
#ifdef ENOTDIR
    case 54: return ENOTDIR;
#endif
#ifdef ENOTEMPTY
    case 55: return ENOTEMPTY;
#endif
#ifdef ENOTRECOVERABLE
    case 56: return ENOTRECOVERABLE;
#endif
#ifdef ENOTSOCK
    case 57: return ENOTSOCK;
#endif
#ifdef ENOTSUP
    case 58: return ENOTSUP;
#endif
#ifdef ENOTTY
    case 59: return ENOTTY;
#endif
#ifdef ENXIO
    case 60: return ENXIO;
#endif
#ifdef EOVERFLOW
    case 61: return EOVERFLOW;
#endif
#ifdef EOWNERDEAD
    case 62: return EOWNERDEAD;
#endif
#ifdef EPERM
    case 63: return EPERM;
#endif
#ifdef EPIPE
    case 64: return EPIPE;
#endif
#ifdef EPROTO
    case 65: return EPROTO;
#endif
#ifdef EPROTONOSUPPORT
    case 66: return EPROTONOSUPPORT;
#endif
#ifdef EPROTOTYPE
    case 67: return EPROTOTYPE;
#endif
#ifdef ERANGE
    case 68: return ERANGE;
#endif
#ifdef EROFS
    case 69: return EROFS;
#endif
#ifdef ESPIPE
    case 70: return ESPIPE;
#endif
#ifdef ESRCH
    case 71: return ESRCH;
#endif
#ifdef ESTALE
    case 72: return ESTALE;
#endif
#ifdef ETIMEDOUT
    case 73: return ETIMEDOUT;
#endif
#ifdef ETXTBSY
    case 74: return ETXTBSY;
#endif
#ifdef EXDEV
    case 75: return EXDEV;
#endif
#ifdef ENOTCAPABLE
    case 76: return ENOTCAPABLE;
#endif
    }
    return EIO;
}
/* The end of an XSUB: an errno it set is noted in the frame for the loader. */
static void goperl_errno_leave(goperl_frame_t *f) {
    if (errno != GOPERL_ERRNO_UNSET) f->xerrno = 1 + goperl_errno_guest(errno);
}

static goperl_frame_t **goperl_curframe_slot(void) {
    goperl_forked_check();
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
    int32_t width; /* 0 = pointer-width; 1, 2, 4 = an integer of that size;
                    * GOPERL_HOSTSAVE_GENERIC = an SV the location holds a
                    * reference to (SAVEGENERICSV) */
} goperl_hostsave_t;
#define GOPERL_HOSTSAVE_GENERIC 9
#define GOPERL_HOSTSAVE_MAX 256
/* The scope of the saves an activation's entry makes (GOPERL_XS_OPEN: the
 * previous activation's frame, catch point and base): no LEAVE reaches it,
 * only the activation's end restores them. An XSUB may leave its caller's
 * scope (Guard's scope_guard: `LEAVE; SAVEDESTRUCTOR_X(...); ENTER;`), and
 * that LEAVE must not undo the bookkeeping of the native calls running
 * under it, its own or that of the XSUB whose callback called it. */
#define GOPERL_HOSTSAVE_ACTIVATION INT32_MIN

/* Control-flow state SHARED ACROSS MODULES. Weak globals only coalesce
 * within one shared object; with several SDK-built .so files loaded
 * (DBI + HTTP::Parser::XS), each had its own current-frame/save-stack and
 * a hook running in one module dereferenced another's stale frame. The
 * loader owns ONE block (api->shared) that every module aliases. */
/* The body of an array as AvARRAY designates it: the tokens of its slots,
 * and what the interpreter held when they were loaded. The slot of the
 * field (GOPERL_LV_AVBODY) holds the address of slots. */
typedef struct goperl_avbody {
    int32_t len; /* the slots the array had room for when it was loaded */
    int32_t cap;
    uint64_t *shadow;
    uint64_t slots[];
} goperl_avbody_t;
/* One field an accessor macro designated as an lvalue (SvIVX, SvRV, GvSV,
 * ...): val is the storage, shadow what the interpreter held when it was
 * loaded, users the evaluations of a macro that designate it and are not
 * known to be over. */
typedef struct goperl_lvslot {
    uint64_t val;
    uint64_t shadow;
    uint64_t svtok;
    int32_t field; /* a GOPERL_LV_ id */
    int32_t users;
    struct goperl_lvslot *free_next;
} goperl_lvslot_t;
#define GOPERL_LVCHUNK 64
typedef struct goperl_lvchunk {
    struct goperl_lvchunk *next;
    goperl_lvslot_t slots[GOPERL_LVCHUNK];
} goperl_lvchunk_t;
/* One evaluation of an accessor macro: the expansion, the activation of the
 * function it is expanded in, and the slot it designates. */
typedef struct goperl_lvuse {
    const void *site;
    uintptr_t frame;
    goperl_lvslot_t *slot;
} goperl_lvuse_t;
/* The slots no evaluation designates any more, kept a while longer for code
 * that holds on to the address of one (see goperl_lv_ref). */
#define GOPERL_LV_RECENT 16
typedef struct goperl_lvstate {
    goperl_lvuse_t *uses; /* in the order they were made */
    int32_t nuses;
    int32_t capuses;
    goperl_lvslot_t *recent[GOPERL_LV_RECENT]; /* oldest first */
    int32_t nrecent;
    int32_t pad;
    goperl_lvslot_t *free;
    goperl_lvchunk_t *chunks;
} goperl_lvstate_t;
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
    int32_t pad_hostsave;
    goperl_hostsave_t hostsave[GOPERL_HOSTSAVE_MAX];
    goperl_lvstate_t *lvstate; /* the lvalue slots accessor macros handed out */
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
    uint64_t parser_pvg; /* guest address of the linestr PV */
    unsigned char parser_shadow[128]; /* the yy_parser instance */
    /* Per-INSTANCE interpreter state that lived in per-.so weak globals
     * until concurrent instances existed (cloned workers): every one of
     * these is mutated at runtime and must be as private per instance as
     * the interpreter it mirrors. Raw byte areas; the owning code aliases
     * them through its own types (typedefs appear later in this header). */
    unsigned char plshadow[128 * 24];     /* goperl_plshadow_t[128] */
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
    STRLEN na; /* PL_na */
    void *pio_files; /* the stdio streams PerlIO_findFILE handed out */
    void *hooks; /* goperl_hooks_t: PL_check, the peephole hooks, call_atexit */
    void *callcks; /* goperl_callcks_t: the call checkers modules installed */
    void *curcop_seen; /* what PL_curcop held when its slot was loaded */
    uint64_t parser_gen; /* gen + 1 when the copy of the parser was made */
    void *compiler; /* goperl_compiler_state_t: the shadows of the compiler's
                     * structures (goperl_compiler.h). New fields go last: a
                     * module built before them leaves them alone. */
    void *statcache; /* goperl_statcache_t: PL_statcache (goperl_compiler.h) */
    void *fntab; /* goperl_fntab_t: functions of native code the interpreter
                  * runs as its own, by id (goperl_compiler.h) */
    void *pio_views; /* goperl_pio_slot_t: the views of handles made for the
                      * native call (perliol.h), in the arena */
    uint64_t plshadow_live[2]; /* the slots of plshadow handed out, a bit
                                * each: what the flush before every
                                * operation looks at */
    void *rx_bodies; /* goperl_rxbody_t: the bodies of regexps ReANY
                      * handed out in this call of native code, in the
                      * arena (goperl_perlapi.h) */
    MGVTBL interp_magic_vtbl; /* all zero: the vtable of the copies of the
                               * interpreter's own magic that SvMAGIC shows
                               * (goperl_svmagic_lv), which tells them from
                               * the chain native code attached */
} goperl_shared_t;
#define goperl_hostsave_v (GOPERL_SHARED->hostsave)
#define goperl_hostsave_n (GOPERL_SHARED->hostsave_n)
#define goperl_scope_v (GOPERL_SHARED->scope)

/* Fields as lvalues (`SvIVX(counter)++`, `&SvNVX(sv)`, `SvRV(rv) = sv`,
 * `GvSV(gv) = NULL`, `--SvREFCNT(sv)`): the accessor designates host storage
 * loaded from the interpreter, and a value that differs from what was
 * loaded is written back before the next guest operation.
 *
 * C does not say whether the left or the right side of an assignment is
 * evaluated first, so the store may come any number of guest operations
 * after the accessor ran (`SvRV(clone) = clone_of(SvRV(ref))`, where the
 * right side is a recursion through this very line). A slot therefore stays
 * under that comparison until the evaluation that made it is over, which
 * is known in two ways: the same expansion of the macro is evaluated again
 * by the same activation of the function it is written in (control cannot
 * come back to an expression that is still being evaluated), or a function
 * above that activation on the C stack evaluates one (the activation has
 * returned). Evaluations that designate the same field get the same
 * storage, so two addresses of one field are one address. A slot that no
 * evaluation designates any more is not given up at once: the last
 * GOPERL_LV_RECENT of them stay under the comparison, for code that keeps
 * the address of a field and stores through it later
 * (`IV *n = counter_of(sv); ... *n += 1;`). An address kept while that many
 * other fields were designated and given up no longer designates the
 * field.
 *
 * What is loaded and written back is the field and nothing else, as with
 * the macros of sv.h: no flag changes, nothing is converted, and a
 * read-only scalar is not refused (a module that keeps a number of its own
 * in the slot of a scalar it has made read-only for the program goes on
 * updating it there). */
#define GOPERL_LV_IV 0
#define GOPERL_LV_UV 1
#define GOPERL_LV_NV 2
#define GOPERL_LV_CVXSUB 3
#define GOPERL_LV_GVSTASH 4
#define GOPERL_LV_GVCVGEN 5
#define GOPERL_LV_RV 6
#define GOPERL_LV_GVSV 7
#define GOPERL_LV_GVAV 8
#define GOPERL_LV_GVHV 9
#define GOPERL_LV_REFCNT 10
#define GOPERL_LV_CVOUTSIDE 11
/* AvARRAY: the slot holds the address of the array's body in this process
 * (a goperl_avbody_t's slots), whose elements are what is written back. */
#define GOPERL_LV_AVBODY 12
/* SvPVX and SvLEN: the buffer of a string (its address in this process)
 * and the room it has. */
#define GOPERL_LV_PVX 13
#define GOPERL_LV_LEN 14
/* The index of the last element of an array, and the type and the target
 * of an lvalue scalar (PVLV). */
#define GOPERL_LV_AVFILLP 15
#define GOPERL_LV_LVTYPE 16
#define GOPERL_LV_LVTARG 17
/* The fields of a sub (goperl_compiler.h), and its body as SvANY designates
 * it: the slot holds the address of a body of perl's fields here, whose
 * fields are what is written back. */
#define GOPERL_LV_CVFLAGS 18
#define GOPERL_LV_CVROOT 19
#define GOPERL_LV_CVSTART 20
#define GOPERL_LV_CVOUTSIDE_SEQ 21
#define GOPERL_LV_CVPADLIST 22
#define GOPERL_LV_CVDEPTH 23
#define GOPERL_LV_CVFILE 24
#define GOPERL_LV_CVBODY 25
/* The array of the fields of an object (ObjectFIELDS): the slot holds the
 * address of the fields here, whose elements are written back; and their
 * number less one (ObjectMAXFIELD). */
#define GOPERL_LV_OBJFIELDS 26
#define GOPERL_LV_OBJMAXFIELD 27
/* The flag word of a scalar (SvFLAGS), its stash (SvSTASH), the flags of a
 * glob (GvFLAGS), and where the iterator of a hash is (HvRITER, and HvEITER,
 * whose slot holds the address of the entry's shadow here). */
#define GOPERL_LV_FLAGS 28
#define GOPERL_LV_STASH 29
#define GOPERL_LV_GVFLAGS 30
#define GOPERL_LV_HVRITER 31
#define GOPERL_LV_HVEITER 32
/* The handles, the type character and the flags of an IO (IoIFP, IoOFP,
 * IoTYPE, IoFLAGS), and the IO of a glob (GvIOp): the plain stores of
 * sv.h and gv.h, which count no reference. */
#define GOPERL_LV_IOIFP 33
#define GOPERL_LV_IOOFP 34
#define GOPERL_LV_IOTYPE 35
#define GOPERL_LV_IOFLAGS 36
#define GOPERL_LV_GVIO 37
/* The other words of an IO's body, plain stores too (Filter::Util::Call
 * keeps its state in them): the line counts, the format globs and the
 * format names, in the order of the guest's GOPERL_IOWORD_ (FIELD 24/25).
 * A name stored is copied into the interpreter, which frees it with the
 * IO, as perl frees the savepv it is given. */
#define GOPERL_LV_IOWORD 38 /* to 47 */
#define GOPERL_LV_IS_IOWORD(field) ((field) >= GOPERL_LV_IOWORD && (field) < GOPERL_LV_IOWORD + 10)
#define GOPERL_IOWORD_IS_NAME(w) ((w) >= 7)
/* The slots whose value is the address of a body here (goperl_compiler.h). */
#define GOPERL_LV_IS_BODY(field) \
    ((field) == GOPERL_LV_CVBODY || (field) == GOPERL_LV_OBJFIELDS)
static uint64_t goperl_lv_load_field(goperl_frame_t *f, uint64_t tok, int field);
static void goperl_lv_store_field(goperl_frame_t *f, uint64_t tok, int field,
                                  uint64_t val);
#define goperl_lvstate_v (GOPERL_SHARED->lvstate)
/* Which evaluation an accessor macro is: an object of its own for every
 * expansion, and the frame of the function the expansion is in. */
#define GOPERL_LV_AT                                              \
    ({ static const char goperl_lv_site_ = 0;                     \
       (const void *)&goperl_lv_site_; }),                        \
    (uintptr_t)__builtin_frame_address(0)

/* PL_ppaddr proxy: every slot is the same generic trampoline. */
GOPERL_C __attribute__((weak, visibility("hidden"))) Perl_ppaddr_t goperl_ppaddr_v[OP_max] = {0};

/* Native-XSUB nesting depth: when the top-level XSUB returns, the lvalue
 * slots are dropped (the interpreter may free what they designate between
 * calls; a stale slot must never be written into reused memory). */
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

GOPERL_C __attribute__((weak, visibility("default"), used)) uint32_t
__goperl_xs_init(const goperl_api_t *api) {
    if (!api || api->abi != GOPERL_XS_ABI) return 0;
    if (!api->shared_raw ||
        sizeof(goperl_shared_t) > GOPERL_SHARED_RESERVED)
        return 0;
    goperl_api_v = api;
    return GOPERL_XS_ABI;
}

/* The structures of the interpreter's that native code reads and writes
 * where they lie in its memory (HE, HEK), as this SDK lays them out: pairs
 * of an index of the interpreter's table of its layout (perl-wasm's
 * GOPERL_INT_LAYOUT) and the value the module was compiled with. The loader
 * compares them with the interpreter's when it loads the module, and
 * refuses a module that disagrees. */
enum {
    GOPERL_LAY_HE_NEXT = 23, GOPERL_LAY_HE_HEK = 24, GOPERL_LAY_HE_VALU = 25,
    GOPERL_LAY_HE_SIZE = 26, GOPERL_LAY_HEK_HASH = 27, GOPERL_LAY_HEK_LEN = 28,
    GOPERL_LAY_HEK_KEY = 29
};
GOPERL_C __attribute__((weak, visibility("default"), used)) const uint32_t *
__goperl_xs_layout(uint32_t *n) {
    static const uint32_t pairs[] = {
        GOPERL_LAY_HE_NEXT, (uint32_t)offsetof(HE, hent_next),
        GOPERL_LAY_HE_HEK, (uint32_t)offsetof(HE, hent_hek),
        GOPERL_LAY_HE_VALU, (uint32_t)offsetof(HE, he_valu),
        GOPERL_LAY_HE_SIZE, (uint32_t)sizeof(HE),
        GOPERL_LAY_HEK_HASH, (uint32_t)offsetof(HEK, hek_hash),
        GOPERL_LAY_HEK_LEN, (uint32_t)offsetof(HEK, hek_len),
        GOPERL_LAY_HEK_KEY, (uint32_t)offsetof(HEK, hek_key),
    };
    *n = (uint32_t)(sizeof(pairs) / sizeof(pairs[0]) / 2);
    return pairs;
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
    GOPERL_OP_SV_CLASSIFY = 184,
    GOPERL_OP_SV_FLAGS = 185,
    GOPERL_OP_SV_SLOT_SET = 186,
    GOPERL_OP_SV_CALL = 187,
    GOPERL_OP_GV_CALL = 188,
    GOPERL_OP_SV_STR = 189,
    GOPERL_OP_SV_SLOT_GET = 190,
    GOPERL_OP_CALLER_CX = 191,
    GOPERL_OP_PERLIO = 192,
    GOPERL_OP_MEM_FREE = 193,
    GOPERL_OP_MEM_REALLOC = 194,
    GOPERL_OP_DO_OPENN = 195,
    GOPERL_OP_FIELD = 196,
    GOPERL_OP_REGEXP = 197,
    GOPERL_OP_HOOK = 198,
    GOPERL_OP_PP_STACK = 199,
    GOPERL_OP_CX = 200,
    GOPERL_OP_OP_DUMP = 201,
    GOPERL_OP_CALL = 202,
    GOPERL_OP_TREE = 203,
    GOPERL_OP_INTERNAL = 204
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
#define GOPERL_PL_FIRSTGV 81
#define GOPERL_PL_SECONDGV 82

/* GOPERL_ONCE binds the value of a macro argument, so that a macro standing
 * in for a function of the perl API evaluates the argument once, as a call
 * of the function would (`av_push(av, sv_rvweaken(newRV_inc(sv)))` makes
 * one reference, not two). An array argument (a string literal) decays to
 * a pointer. */
#ifdef __cplusplus
#define GOPERL_ONCE(var, expr) auto var = (expr)
#else
#define GOPERL_ONCE(var, expr) __typeof__((void)0, (expr)) var = (expr)
#endif

#define GOPERL_TOK(sv) ((uint64_t)(uintptr_t)(sv))
/* A word of the records that cross as words: the strings of words an
 * operation takes, the answers the interpreter lays out as words in its
 * memory, the requests and responses of its calls of the loader. A word is
 * as wide as an address, which it may hold. */
typedef uint64_t goperl_gword_t;
#define GOPERL_SV(tok) ((SV *)(uintptr_t)(tok))

/* ---- the guest-op funnel: every operation flushes host state first ------ */

static void goperl_flush(goperl_frame_t *f);
static uint64_t *goperl_lv_ref(goperl_frame_t *f, SV *sv, int field,
                               const void *site, uintptr_t frame);
static void goperl_hostsave_ptr(goperl_frame_t *f, void **loc);
static void goperl_hostsave_i32(goperl_frame_t *f, int32_t *loc);
static void goperl_svcur_writeback(goperl_frame_t *f);
static void goperl_gen_bump(void);

static void goperl_raise(goperl_frame_t *f) __attribute__((noreturn, unused));

static void goperl_pads_refresh(goperl_frame_t *f);
/* An operation of the perl API's: the funnel every one goes through. With
 * second, the second value of an operation that answers two is stored
 * there, taken out of the box before anything else can use it. */
static uint64_t goperl_do_opx(goperl_frame_t *f, int32_t op, uint64_t a,
                              uint64_t b, uint64_t c, const char *s,
                              uint64_t slen, uint64_t *second) {
    const int32_t failed_before_flush = f->failed;
    goperl_flush(f);
    /* what the flush found wrong (native code unlinked the interpreter's
     * magic from the chain SvMAGIC showed) stops the operation before it
     * runs: it may be the free of what native code meant to take off */
    if (!failed_before_flush && f->failed && f->jb) goperl_raise(f);
    const int32_t failed_before = f->failed;
    uint64_t r = goperl_xs_opx(f, op, a, b, c, s, slen, second);
    /* any guest operation may have changed interpreter state the context
     * mirrors reflect, and the pads native code holds shadows of */
    goperl_gen_bump();
    goperl_pads_refresh(f);
    /* The interpreter died inside the operation (the loader left the
     * message in the frame). A perl API function does not return to its
     * caller in that case: it unwinds to the nearest catch point, which
     * here is the XSUB entry or an enclosing XCPT_TRY. Without one armed
     * (a callback the loader runs outside an XSUB) the caller goes on with
     * the zero result and the loader reports the failure afterwards. */
    if (!failed_before && f->failed && f->jb) goperl_raise(f);
    return r;
}
/* An operation of two operands (c is 0), and one of three. */
static inline __attribute__((unused)) uint64_t goperl_do_op(
    goperl_frame_t *f, int32_t op, uint64_t a, uint64_t b, const char *s,
    uint64_t slen) {
    return goperl_do_opx(f, op, a, b, 0, s, slen, 0);
}
static inline __attribute__((unused)) uint64_t goperl_do_op3(
    goperl_frame_t *f, int32_t op, uint64_t a, uint64_t b, uint64_t c,
    const char *s, uint64_t slen) {
    return goperl_do_opx(f, op, a, b, c, s, slen, 0);
}

#define goperl_op0(op, a, b) \
    (goperl_do_op(_gof, (op), (uint64_t)(a), (uint64_t)(b), 0, 0))
#define goperl_ops(op, a, b, s, l)                                    \
    (goperl_do_op(_gof, (op), (uint64_t)(a), (uint64_t)(b), (s),      \
                  (uint64_t)(l)))
#define goperl_op3(op, a, b, c) \
    (goperl_do_op3(_gof, (op), (uint64_t)(a), (uint64_t)(b), (uint64_t)(c), 0, 0))

/* croak: format into the frame and longjmp back to the XSUB entry. */
static void goperl_vfmt(goperl_frame_t *f, char *dst, size_t dstsize,
                        const char *fmt, va_list ap);

/* A saved location gets back what it held: as many bytes as were saved.
 * The location of SAVEGENERICSV owns a reference to what it holds, as
 * save_generic_svref has it: the value it holds now is released, and the
 * saved one, whose reference was kept, is put back. */
static void goperl_hostsave_restore(goperl_frame_t *f, const goperl_hostsave_t *e) {
    switch (e->width) {
    case 1: *(int8_t *)e->loc = (int8_t)(intptr_t)e->val; break;
    case 2: *(int16_t *)e->loc = (int16_t)(intptr_t)e->val; break;
    case 4: *(int32_t *)e->loc = (int32_t)(intptr_t)e->val; break;
    case GOPERL_HOSTSAVE_GENERIC: {
        void *const now = *e->loc;
        *e->loc = e->val;
        if (now)
            goperl_xs_op(f, GOPERL_OP_SV_REFCNT_DEC, (uint64_t)(uintptr_t)now, 0, 0, 0);
        break;
    }
    default: *e->loc = e->val; break;
    }
}

static void goperl_hostsave_unwind_to(goperl_frame_t *f, int32_t base) {
    while (goperl_hostsave_n > base) {
        goperl_hostsave_t *e = &goperl_hostsave_v[--goperl_hostsave_n];
        goperl_hostsave_restore(f, e);
    }
}

/* The jump out of native code that ends a native call with a failure.
 * The scopes the call opened are left first (HOOK 5), while its C frames
 * exist: a destructor it registered for a variable of its own
 * (SAVEDESTRUCTOR_X(cleanup, &local)) runs on that variable, as it does
 * when perl unwinds to an eval before it jumps. */
static void goperl_jump_out(goperl_frame_t *f) __attribute__((noreturn));
static void goperl_jump_out(goperl_frame_t *f) {
    if (f->native_call) {
        const int32_t failed = f->failed;
        f->failed = 0; /* the frame is still in use while the scopes go */
        goperl_xs_op(f, GOPERL_OP_HOOK, 5, 0, 0, 0);
        f->failed = failed;
    }
    GOPERL_LONGJMP(*(jmp_buf *)f->jb, 1);
}

/* The interpreter left an operation by a non-local exit (it died inside
 * it, and what catches that is further out than this native call): the
 * call ends without a message of its own, and the interpreter goes on
 * unwinding once the call has returned. */
#define GOPERL_FAILED_JUMP 3
static void goperl_raise(goperl_frame_t *f) {
    if (f->failed != GOPERL_FAILED_JUMP) f->failed = 1;
    if (!f->jb) {
        /* No XSUB entry to unwind to (e.g. a croak inside svt_free
         * teardown): surface the message and stop instead of jumping
         * through a null buffer. */
        fprintf(stderr, "goperl native XS: croak outside an XSUB: %s\n",
                f->err);
        abort();
    }
    goperl_jump_out(f);
}

/* The loader's direct accessors. They follow the contract of goperl_do_op
 * (minus the flush, which their callers do): when the interpreter dies
 * inside one, the call does not return. */
static inline __attribute__((unused)) void goperl_direct_end(
    goperl_frame_t *f, int32_t failed_before) {
    goperl_gen_bump();
    if (!failed_before && f->failed && f->jb) goperl_raise(f);
}
#define GOPERL_DIRECT_END(f, failed_before) goperl_direct_end((f), (failed_before))
static int64_t goperl_direct_sv_iv(goperl_frame_t *f, uint64_t sv)
    __attribute__((unused));
static int64_t goperl_direct_sv_iv(goperl_frame_t *f, uint64_t sv) {
    const int32_t failed_before = f->failed;
    const int goperl_e = errno;
    int64_t r = goperl_api_v->sv_iv(f, sv);
    errno = goperl_e;
    GOPERL_DIRECT_END(f, failed_before);
    return r;
}
static const char *goperl_direct_sv_pv(goperl_frame_t *f, uint64_t sv,
                                       uint64_t *lenp) __attribute__((unused));
static const char *goperl_direct_sv_pv(goperl_frame_t *f, uint64_t sv,
                                       uint64_t *lenp) {
    const int32_t failed_before = f->failed;
    const int goperl_e = errno;
    const char *r = goperl_api_v->sv_pv(f, sv, lenp);
    errno = goperl_e;
    GOPERL_DIRECT_END(f, failed_before);
    return r;
}
static uint64_t goperl_direct_new_iv(goperl_frame_t *f, int64_t v)
    __attribute__((unused));
static uint64_t goperl_direct_new_iv(goperl_frame_t *f, int64_t v) {
    const int32_t failed_before = f->failed;
    const int goperl_e = errno;
    uint64_t r = goperl_api_v->new_iv(f, v);
    errno = goperl_e;
    GOPERL_DIRECT_END(f, failed_before);
    return r;
}
static uint64_t goperl_direct_new_pvn(goperl_frame_t *f, const char *p,
                                      uint64_t len) __attribute__((unused));
static uint64_t goperl_direct_new_pvn(goperl_frame_t *f, const char *p,
                                      uint64_t len) {
    const int32_t failed_before = f->failed;
    const int goperl_e = errno;
    uint64_t r = goperl_api_v->new_pvn(f, p, len);
    errno = goperl_e;
    GOPERL_DIRECT_END(f, failed_before);
    return r;
}
static uint64_t goperl_direct_sv_mortal(goperl_frame_t *f, uint64_t sv)
    __attribute__((unused));
static uint64_t goperl_direct_sv_mortal(goperl_frame_t *f, uint64_t sv) {
    const int32_t failed_before = f->failed;
    const int goperl_e = errno;
    uint64_t r = goperl_api_v->sv_mortal(f, sv);
    errno = goperl_e;
    GOPERL_DIRECT_END(f, failed_before);
    return r;
}

static void goperl_croakf(goperl_frame_t *f, const char *fmt, ...)
    __attribute__((noreturn, unused));
static void goperl_croak_errsv(goperl_frame_t *f)
    __attribute__((noreturn, unused));
static void goperl_vcroak(goperl_frame_t *f, const char *fmt, va_list ap)
    __attribute__((noreturn, unused));
static void goperl_croakf(goperl_frame_t *f, const char *fmt, ...) {
    va_list ap;
    if (!fmt) goperl_croak_errsv(f);
    va_start(ap, fmt);
    goperl_vcroak(f, fmt, ap);
}


/* ---- SVf-aware formatting ----------------------------------------------
 * SVf is a marker sequence; "%"SVf in a format consumes an SV* vararg and
 * appends its PV. Everything else is parsed one conversion at a time and
 * handed to snprintf with the right va_arg type. */
#define SVf "\002GSV\002"
#define UTF8f "\002GU8\002"
/* A shared key in a message: %" HEKf " with HEKfARG(hek). */
#define HEKf "\002GHK\002"
#define HEKf256 HEKf
#define HEKfARG(hek) ((const HEK *)(hek))
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
    return goperl_direct_sv_pv(f, GOPERL_TOK(sv), lenp);
}

/* SV_STR selectors of the formatter's consumers (the others are listed
 * in goperl_perlapi.h). Their flags: bit 0, the text is UTF-8. */
#define GOPERL_SVSTR_NEWPVN 11 /* bit 1: the new SV is mortal */
#define GOPERL_SVSTR_WARN 12
#define GOPERL_SVSTR_THROW 13
#define GOPERL_SVSTR_SETPVN 14
#define GOPERL_SVSTR_WARNER 16 /* the SV slot is the category; flags bits
                                * 1-2: warner, ck_warner, ck_warner_d */
#define GOPERL_TEXT_UTF8 1
#define GOPERL_TEXT_MORTAL 2

/* The text a format produced: its bytes, and whether they are UTF-8. Like
 * the SV perl formats into, it starts as one character per byte and
 * becomes UTF-8 when an argument is: what it holds is upgraded then, and
 * every later piece that is not UTF-8 is upgraded as it is appended. */
typedef struct goperl_text {
    char *p;
    size_t len;
    size_t cap;
    int utf8;
    char first[512];
} goperl_text_t;

static void goperl_text_init(goperl_text_t *t) {
    t->p = t->first;
    t->len = 0;
    t->cap = sizeof t->first;
    t->utf8 = 0;
    t->first[0] = '\0';
}
static void goperl_text_free(goperl_text_t *t) {
    if (t->p != t->first) free(t->p);
    t->p = t->first;
}
/* Room for n more bytes and the terminator. */
static void goperl_text_room(goperl_text_t *t, size_t n) {
    if (t->len + n + 1 <= t->cap) return;
    size_t cap = t->cap * 2;
    while (cap < t->len + n + 1) cap *= 2;
    char *np = (char *)malloc(cap);
    if (!np) abort();
    memcpy(np, t->p, t->len + 1);
    if (t->p != t->first) free(t->p);
    t->p = np;
    t->cap = cap;
}
static void goperl_text_put(goperl_text_t *t, const char *src, size_t n,
                            int is_utf8) {
    if (is_utf8 && !t->utf8) {
        size_t high = 0;
        for (size_t i = 0; i < t->len; i++)
            if ((unsigned char)t->p[i] >= 0x80) high++;
        if (high) {
            goperl_text_room(t, high);
            size_t from = t->len, to = t->len + high;
            t->p[to] = '\0';
            while (from > 0) {
                unsigned char c = (unsigned char)t->p[--from];
                if (c >= 0x80) {
                    t->p[--to] = (char)(0x80 | (c & 0x3f));
                    t->p[--to] = (char)(0xc0 | (c >> 6));
                } else {
                    t->p[--to] = (char)c;
                }
            }
            t->len += high;
        }
        t->utf8 = 1;
    }
    if (t->utf8 && !is_utf8) {
        goperl_text_room(t, 2 * n);
        char *w = t->p + t->len;
        for (size_t i = 0; i < n; i++) {
            unsigned char c = (unsigned char)src[i];
            if (c >= 0x80) {
                *w++ = (char)(0xc0 | (c >> 6));
                *w++ = (char)(0x80 | (c & 0x3f));
            } else {
                *w++ = (char)c;
            }
        }
        t->len = (size_t)(w - t->p);
    } else {
        goperl_text_room(t, n);
        if (n) memcpy(t->p + t->len, src, n);
        t->len += n;
    }
    t->p[t->len] = '\0';
}

/* One standard conversion of the next argument, of the given type,
 * rendered by snprintf into a buffer of the size it needs. */
#define GOPERL_VA(app, type) \
    ((app) ? va_arg(*(app), type) : (goperl_missing = 1, (type)0))
#define GOPERL_TEXT_CONV(t, spec, app, type)                             \
    do {                                                                 \
        type goperl_arg = GOPERL_VA(app, type);                          \
        GOPERL_TEXT_VALUE(t, spec, goperl_arg);                          \
    } while (0)
#define GOPERL_TEXT_VALUE(t, spec, arg)                                  \
    do {                                                                 \
        char goperl_small[128];                                          \
        int goperl_n = snprintf(goperl_small, sizeof goperl_small, (spec), (arg)); \
        if (goperl_n < 0) break;                                         \
        if ((size_t)goperl_n < sizeof goperl_small) {                    \
            goperl_text_put((t), goperl_small, (size_t)goperl_n, 0);     \
        } else {                                                         \
            char *goperl_big = (char *)malloc((size_t)goperl_n + 1);     \
            if (!goperl_big) abort();                                    \
            snprintf(goperl_big, (size_t)goperl_n + 1, (spec), (arg));   \
            goperl_text_put((t), goperl_big, (size_t)goperl_n, 0);       \
            free(goperl_big);                                            \
        }                                                                \
    } while (0)

/* The text of a format and its arguments: a pointer to them, as perl's
 * functions that take a va_list * have it (what is taken is gone from the
 * caller's list), or NULL for none, where every conversion is a missing
 * argument: perl's empty string, or zero, and a warning in the category
 * "missing". */
static void Perl_ck_warner(pTHX_ U32 cat, const char *fmt, ...);
static const char *goperl_op_desc_now(goperl_frame_t *f);
static void goperl_vtext_p(goperl_frame_t *f, goperl_text_t *t, const char *fmt,
                           va_list *app) {
    int goperl_missing = 0;
    while (*fmt) {
        if (*fmt != '%') {
            const char *run = fmt;
            while (*fmt && *fmt != '%') fmt++;
            goperl_text_put(t, run, (size_t)(fmt - run), 0);
            continue;
        }
        fmt++;
        if (*fmt == '%') {
            goperl_text_put(t, "%", 1, 0);
            fmt++;
            continue;
        }
        if (strncmp(fmt, SVf, sizeof(SVf) - 1) == 0) {
            SV *sv = GOPERL_VA(app, SV *);
            fmt += sizeof(SVf) - 1;
            if (!sv) {
                if (app) goperl_text_put(t, "(null)", 6, 0);
                continue;
            }
            uint64_t len = 0;
            const char *sp = goperl_sv_pv_(f, sv, &len);
            /* the flag is read after the string: making the string is what
             * settles it (an overloaded object, a number) */
            int is_utf8 = (goperl_do_op(f, GOPERL_OP_SV_INFO, GOPERL_TOK(sv),
                                        0, 0, 0) &
                           GOPERL_INFO_UTF8) != 0;
            if (sp) goperl_text_put(t, sp, (size_t)len, is_utf8);
            continue;
        }
        if (strncmp(fmt, UTF8f, sizeof(UTF8f) - 1) == 0) {
            int is_utf8 = GOPERL_VA(app, int);
            UV ulen = GOPERL_VA(app, UV);
            const char *up = GOPERL_VA(app, const char *);
            if (up) goperl_text_put(t, up, (size_t)ulen, is_utf8 != 0);
            fmt += sizeof(UTF8f) - 1;
            continue;
        }
        /* a shared key (HEKf): its bytes, UTF-8 by the flags byte that
         * follows them and their NUL */
        if (strncmp(fmt, HEKf, sizeof(HEKf) - 1) == 0) {
            const HEK *hek = GOPERL_VA(app, const HEK *);
            fmt += sizeof(HEKf) - 1;
            if (hek)
                goperl_text_put(t, hek->hek_key, (size_t)hek->hek_len,
                                (hek->hek_key[hek->hek_len + 1] & 1) != 0);
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
                                   GOPERL_VA(app, int));
        } else {
            while (*fmt >= '0' && *fmt <= '9' && sn < 14) spec[sn++] = *fmt++;
        }
        if (*fmt == '.') {
            spec[sn++] = *fmt++;
            if (*fmt == '*') {
                fmt++;
                sn += (size_t)snprintf(spec + sn, sizeof spec - sn, "%d",
                                       GOPERL_VA(app, int));
            } else {
                while (*fmt >= '0' && *fmt <= '9' && sn < 24)
                    spec[sn++] = *fmt++;
            }
        }
        int goperl_lmod = 0; /* 0 none, 1 l, 2 ll (or perl's q), 3 z, 4 h, 5 t,
                              * 6 j, 7 L */
        if (fmt[0] == 'l' && fmt[1] == 'l') {
            goperl_lmod = 2;
            spec[sn++] = *fmt++;
            spec[sn++] = *fmt++;
        } else if (*fmt == 'q') {
            goperl_lmod = 2;
            spec[sn++] = 'l';
            spec[sn++] = 'l';
            fmt++;
        } else if (*fmt == 't') {
            goperl_lmod = 5;
            spec[sn++] = *fmt++;
        } else if (*fmt == 'j') {
            goperl_lmod = 6;
            spec[sn++] = *fmt++;
        } else if (*fmt == 'L') {
            goperl_lmod = 7;
            spec[sn++] = *fmt++;
        } else if (*fmt == 'l') {
            goperl_lmod = 1;
            spec[sn++] = *fmt++;
        } else if (*fmt == 'z') {
            goperl_lmod = 3;
            spec[sn++] = *fmt++;
        } else if (*fmt == 'h') {
            goperl_lmod = 4;
            spec[sn++] = *fmt++;
            if (*fmt == 'h') spec[sn++] = *fmt++;
        }
        char conv = *fmt ? *fmt++ : 0;
        if (!conv) break;
        spec[sn++] = conv;
        spec[sn] = '\0';
        switch (conv) {
        case 'd':
        case 'i':
            if (goperl_lmod == 2) GOPERL_TEXT_CONV(t, spec, app, long long);
            else if (goperl_lmod == 1) GOPERL_TEXT_CONV(t, spec, app, long);
            else if (goperl_lmod == 3) GOPERL_TEXT_CONV(t, spec, app, size_t);
            else if (goperl_lmod == 5) GOPERL_TEXT_CONV(t, spec, app, ptrdiff_t);
            else if (goperl_lmod == 6) GOPERL_TEXT_CONV(t, spec, app, intmax_t);
            else GOPERL_TEXT_CONV(t, spec, app, int);
            break;
        case 'u':
        case 'o':
        case 'x':
        case 'X':
            if (goperl_lmod == 2) GOPERL_TEXT_CONV(t, spec, app, unsigned long long);
            else if (goperl_lmod == 1) GOPERL_TEXT_CONV(t, spec, app, unsigned long);
            else if (goperl_lmod == 3 || goperl_lmod == 5) GOPERL_TEXT_CONV(t, spec, app, size_t);
            else if (goperl_lmod == 6) GOPERL_TEXT_CONV(t, spec, app, uintmax_t);
            else GOPERL_TEXT_CONV(t, spec, app, unsigned int);
            break;
        case 'c':
            GOPERL_TEXT_CONV(t, spec, app, int);
            break;
        case 's': {
            const char *sp = GOPERL_VA(app, const char *);
            if (!sp) sp = app ? "(null)" : "";
            GOPERL_TEXT_VALUE(t, spec, sp);
            break;
        }
        case 'p': {
            /* perl's %p is the address in hexadecimal digits, as %lx is
             * (with 0x before them only by the # flag); the C library's
             * puts 0x before them */
            spec[sn - 1] = 'l';
            spec[sn] = 'x';
            spec[sn + 1] = '\0';
            const unsigned long pv = (unsigned long)(uintptr_t)GOPERL_VA(app, void *);
            GOPERL_TEXT_VALUE(t, spec, pv);
            break;
        }
        case 'e':
        case 'E':
        case 'f':
        case 'F':
        case 'g':
        case 'G':
            if (goperl_lmod == 7) GOPERL_TEXT_CONV(t, spec, app, long double);
            else GOPERL_TEXT_CONV(t, spec, app, double);
            break;
        default: {
            char unknown[8];
            int un = snprintf(unknown, sizeof unknown, "<%%%c?>", conv);
            goperl_text_put(t, unknown, (size_t)un, 0);
            break;
        }
        }
    }
    if (goperl_missing)
        Perl_ck_warner(WARN_MISSING, "Missing argument in %s",
                       goperl_op_desc_now(f));
}
static void goperl_vtext(goperl_frame_t *f, goperl_text_t *t, const char *fmt,
                         va_list ap) {
    va_list goperl_ap;
    va_copy(goperl_ap, ap);
    goperl_vtext_p(f, t, fmt, &goperl_ap);
    va_end(goperl_ap);
}

/* The formatted text in a buffer of fixed size, cut where it does not
 * fit (form's rotating buffers, messages on the standard error). */
static void goperl_vfmt(goperl_frame_t *f, char *dst, size_t dstsize,
                        const char *fmt, va_list ap) {
    goperl_text_t t;
    if (dstsize == 0) return;
    goperl_text_init(&t);
    goperl_vtext(f, &t, fmt, ap);
    size_t n = t.len < dstsize - 1 ? t.len : dstsize - 1;
    memcpy(dst, t.p, n);
    dst[n] = '\0';
    goperl_text_free(&t);
}

/* The op argument that carries a text's length and flags. */
static inline __attribute__((unused)) uint64_t goperl_text_arg(
    const goperl_text_t *t, uint32_t extra) {
    return ((uint64_t)((t->utf8 ? GOPERL_TEXT_UTF8 : 0) | extra) << 32) |
           (uint32_t)t->len;
}
#define GOPERL_TEXT_ARG(t, extra) goperl_text_arg((t), (uint32_t)(extra))

/* Runs a string operation of the interpreter on a text and releases the
 * text. The contract is goperl_do_op's: when the interpreter dies inside
 * the operation the call does not return, and the text is released
 * before the unwinding. */
static uint64_t goperl_text_op(goperl_frame_t *f, goperl_text_t *t, int32_t op,
                               uint64_t a, uint64_t b, uint64_t c) {
    goperl_flush(f);
    const int32_t failed_before = f->failed;
    uint64_t r = goperl_xs_opx(f, op, a, b, c, t->p, t->len, 0);
    goperl_gen_bump();
    goperl_text_free(t);
    if (!failed_before && f->failed && f->jb) goperl_raise(f);
    return r;
}

/* A new SV holding the formatted text, UTF-8 when the text is. */
static uint64_t goperl_vtext_sv(goperl_frame_t *f, int mortal, const char *fmt,
                                va_list ap) {
    goperl_text_t t;
    goperl_text_init(&t);
    goperl_vtext(f, &t, fmt, ap);
    return goperl_text_op(f, &t, GOPERL_OP_SV_STR, 0, GOPERL_TEXT_ARG(&t, mortal ? GOPERL_TEXT_MORTAL : 0), (uint64_t)(GOPERL_SVSTR_NEWPVN));
}

/* warn: the interpreter gets the text and its encoding, and does the
 * rest (the place of the caller, $SIG{__WARN__}). */
static void goperl_vwarn(goperl_frame_t *f, const char *fmt, va_list ap)
    __attribute__((unused));
static void goperl_vwarn(goperl_frame_t *f, const char *fmt, va_list ap) {
    goperl_text_t t;
    goperl_text_init(&t);
    goperl_vtext(f, &t, fmt, ap);
    goperl_text_op(f, &t, GOPERL_OP_SV_STR, 0, GOPERL_TEXT_ARG(&t, 0), (uint64_t)(GOPERL_SVSTR_WARN));
}

/* croak: the exception is the formatted text as a string of the
 * interpreter, whole and with its encoding. It is handed over here and the
 * frame ends as one that threw a value, like croak_sv. The frame keeps the
 * start of the text for the catch points that only report a message. */
static void goperl_croak_text(goperl_frame_t *f, goperl_text_t *tp)
    __attribute__((noreturn));
static void goperl_vcroak(goperl_frame_t *f, const char *fmt, va_list ap) {
    goperl_text_t t;
    goperl_text_init(&t);
    goperl_vtext(f, &t, fmt, ap);
    goperl_croak_text(f, &t);
}
static void goperl_croak_text(goperl_frame_t *f, goperl_text_t *tp) {
    size_t n = tp->len < sizeof f->err - 1 ? tp->len : sizeof f->err - 1;
    memcpy(f->err, tp->p, n);
    f->err[n] = '\0';
    if (!f->jb) {
        goperl_text_free(tp);
        goperl_raise(f);
    }
    goperl_flush(f);
    goperl_xs_opx(f, GOPERL_OP_SV_STR, 0, GOPERL_TEXT_ARG(tp, 0), (uint64_t)(GOPERL_SVSTR_THROW), tp->p, tp->len, 0);
    goperl_text_free(tp);
    f->failed = 2; /* GOPERL_FAILED_THROWN */
    goperl_jump_out(f);
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

/* croak() and warn() are variadic macros over the current frame;
 * Perl_croak, Perl_warn and Perl_warner are variadic functions with perl's
 * prototypes. */
#define croak(...) goperl_croakf(goperl_cur_frame_v, __VA_ARGS__)
static void Perl_croak(pTHX_ const char *fmt, ...)
    __attribute__((noreturn, unused));
static void Perl_croak(pTHX_ const char *fmt, ...) {
    va_list ap;
    if (!fmt) goperl_croak_errsv(_gof);
    va_start(ap, fmt);
    goperl_vcroak(_gof, fmt, ap);
}
static void Perl_croak_nocontext(const char *fmt, ...)
    __attribute__((noreturn, unused));
static void Perl_croak_nocontext(const char *fmt, ...) {
    goperl_frame_t *f = goperl_cur_frame_v;
    va_list ap;
    if (!fmt) goperl_croak_errsv(f);
    va_start(ap, fmt);
    goperl_vcroak(f, fmt, ap);
}

static void goperl_warnf(goperl_frame_t *f, const char *fmt, ...)
    __attribute__((unused));
static void goperl_warnf(goperl_frame_t *f, const char *fmt, ...) {
    va_list ap;
    va_start(ap, fmt);
    goperl_vwarn(f, fmt, ap);
    va_end(ap);
}
#define warn(...) goperl_warnf(goperl_cur_frame_v, __VA_ARGS__)
static void Perl_warn(pTHX_ const char *fmt, ...) __attribute__((unused));
static void Perl_warn(pTHX_ const char *fmt, ...) {
    va_list ap;
    va_start(ap, fmt);
    goperl_vwarn(_gof, fmt, ap);
    va_end(ap);
}
static void Perl_warner(pTHX_ U32 cat, const char *fmt, ...)
    __attribute__((unused));
static void goperl_vwarner(goperl_frame_t *f, int kind, U32 cat,
                           const char *fmt, va_list ap) {
    goperl_text_t t;
    goperl_text_init(&t);
    goperl_vtext(f, &t, fmt, ap);
    goperl_text_op(f, &t, GOPERL_OP_SV_STR, cat, GOPERL_TEXT_ARG(&t, (uint32_t)kind << 1), (uint64_t)(GOPERL_SVSTR_WARNER));
}
/* A warning of a category is the interpreter's to give: it is fatal where
 * the caller's `use warnings FATAL` says so, and ck_warner and ck_warner_d
 * give it only where the category is enabled (or not disabled). */
static void Perl_warner(pTHX_ U32 cat, const char *fmt, ...) {
    va_list ap;
    va_start(ap, fmt);
    goperl_vwarner(_gof, 0, cat, fmt, ap);
    va_end(ap);
}
static void Perl_ck_warner(pTHX_ U32 cat, const char *fmt, ...)
    __attribute__((unused));
static void Perl_ck_warner(pTHX_ U32 cat, const char *fmt, ...) {
    va_list ap;
    va_start(ap, fmt);
    goperl_vwarner(_gof, 1, cat, fmt, ap);
    va_end(ap);
}
static void Perl_ck_warner_d(pTHX_ U32 cat, const char *fmt, ...)
    __attribute__((unused));
static void Perl_ck_warner_d(pTHX_ U32 cat, const char *fmt, ...) {
    va_list ap;
    va_start(ap, fmt);
    goperl_vwarner(_gof, 2, cat, fmt, ap);
    va_end(ap);
}
#define packWARN(a) (a)
#define packWARN2(a, b) ((a) | ((b) << 8))
#define packWARN3(a, b, c) ((a) | ((b) << 8) | ((c) << 16))
#define packWARN4(a, b, c, d) ((a) | ((b) << 8) | ((c) << 16) | ((d) << 24))

/* ---- basic SV constructors / accessors ---------------------------------- */

#define SvIV(sv) \
    (goperl_flush(_gof), goperl_direct_sv_iv(_gof, GOPERL_TOK(sv)))
/* SvIVX is an LVALUE in the wild (`SvIVX(counter)++`): host IV mirror. */
#define SvIVX(sv) \
    (*(IV *)goperl_lv_ref(_gof, (SV *)(sv), GOPERL_LV_IV, GOPERL_LV_AT))
#define SvIVx(sv) SvIV(sv)
#define SvUV(sv) ((UV)goperl_op0(GOPERL_OP_SV_IV, GOPERL_TOK(sv), 1))
#define SvUVX(sv) \
    (*(UV *)goperl_lv_ref(_gof, (SV *)(sv), GOPERL_LV_UV, GOPERL_LV_AT))
#define SvUVx(sv) SvUV(sv)

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

/* The SvPV forms with a length give it to a variable of the module's, which
 * may be of any integer type: perl's macros assign the length (`int len;
 * SvPV(sv, len)`, Digest::OAT). The call writes it into a STRLEN of its
 * own, and the variable is assigned from that; written through a pointer
 * to the variable, the length's eight bytes went past an int. */
#define GOPERL_PV_LEN(lenvar, call)                \
    ({ STRLEN goperl_pvlen_ = 0;                   \
       char *const goperl_pvp_ = (char *)(call);   \
       (lenvar) = goperl_pvlen_;                   \
       goperl_pvp_; })
#define SvPV_nolen(sv) ((char *)goperl_sv_pv_(_gof, (SV *)(sv), 0))
#define SvPV_nolen_const(sv) goperl_sv_pv_(_gof, (SV *)(sv), 0)
#define SvPV(sv, lenvar) \
    GOPERL_PV_LEN(lenvar, ((char *)goperl_sv_pv_(_gof, (SV *)(sv), (uint64_t *)&goperl_pvlen_)))
#define SvPV_const(sv, lenvar) \
    GOPERL_PV_LEN(lenvar, goperl_sv_pv_(_gof, (SV *)(sv), (uint64_t *)&goperl_pvlen_))
/* SvPVX is the RAW buffer pointer (no get-magic, no stringification):
 * code grows the SV and writes through it. NULL when the SV has no
 * buffer yet — grow first, as real perl requires. */
/* The buffer slot of an SV holds whatever pointer a module puts there, as
 * in sv.h: a module keeps a structure of its own behind an SV that way
 * (`sv = newSV(0); sv_upgrade(sv, SVt_PV); SvPVX(sv) = ptr;`) and gets it
 * back with SvPVX. The interpreter's slot is an address of its own memory
 * and cannot hold a pointer of this process's, so such a pointer is kept by
 * the loader, for the SV, and the interpreter's slot holds GOPERL_PV_HOST:
 * an address no buffer has, which says where the pointer is. The
 * interpreter leaves the slot of an SV alone that is no string and owns no
 * buffer (SvLEN 0), which is what makes this use of it possible on any
 * perl.
 *
 * A module that gives such an SV a length of buffer as well (DBI:
 * `SvPVX(cv) = savepv(file); SvLEN(cv) = 1;`) hands the memory over: perl
 * frees it with the SV. The interpreter here cannot free memory of this
 * process, so the length is kept with the pointer, the interpreter's own
 * stays 0, and the memory is not freed when the SV goes. */
#define GOPERL_PV_HOST 1u
/* What is kept for sv when its buffer slot holds a pointer of this
 * process's, NULL when it holds none (an entry may be left from an SV that
 * was at the address before: the interpreter's slot says). */
static void **goperl_pv_host_of(goperl_frame_t *f, uint64_t tok) {
    void **kept = f->api->pv_host(f, tok, 0);
    if (!kept) return 0;
    return goperl_xs_op(f, GOPERL_OP_SV_PVX_RAW, tok, 0, 0, 0) == GOPERL_PV_HOST
               ? kept
               : 0;
}
static char *goperl_sv_pvx(goperl_frame_t *f, SV *sv) {
    uint64_t gp =
        goperl_do_op(f, GOPERL_OP_SV_PVX_RAW, GOPERL_TOK(sv), 0, 0, 0);
    if (!gp) return 0;
    if (gp == GOPERL_PV_HOST) {
        void **kept = f->api->pv_host(f, GOPERL_TOK(sv), 0);
        return kept ? (char *)kept[0] : 0;
    }
    return (char *)goperl_guest_mem(f, gp);
}
/* SvPVX is the buffer of the string, and an lvalue: a module takes the
 * buffer over (SvPVX(sv) = NULL; SvLEN(sv) = 0) or gives the SV another
 * one, as with SvPV_set. */
#define SvPVX(sv) \
    (*(char **)goperl_lv_ref(_gof, (SV *)(sv), GOPERL_LV_PVX, GOPERL_LV_AT))
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
        goperl_xs_op(f, GOPERL_OP_SV_CUR_SET, sv,
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
#define SvEND(sv) \
    ({ GOPERL_ONCE(goperl_SvEND_sv_, sv); \
       SvPVX(goperl_SvEND_sv_) + SvCUR(goperl_SvEND_sv_); })

static char *goperl_sv_grow(goperl_frame_t *f, SV *sv, STRLEN n) {
    uint64_t gp = goperl_do_op(f, GOPERL_OP_SV_GROW, GOPERL_TOK(sv),
                               (uint64_t)n, 0, 0);
    return (char *)goperl_guest_mem(f, gp);
}
#define SvGROW(sv, n) goperl_sv_grow(_gof, (SV *)(sv), (STRLEN)(n))
#define sv_grow(sv, n) SvGROW((sv), (n))

#define sv_2mortal(sv)                                        \
    ((SV *)(uintptr_t)(goperl_flush(_gof),                    \
                       goperl_direct_sv_mortal(_gof, GOPERL_TOK(sv))))
#define newSViv(v)                                            \
    ((SV *)(uintptr_t)(goperl_flush(_gof),                    \
                       goperl_direct_new_iv(_gof, (int64_t)(v))))
/* Without a string (a NULL pointer) the new SV is undefined, whatever the
 * length says: newSVpvn(NULL, 0) is how XS code returns undef. */
static SV *goperl_newSVpvn(goperl_frame_t *f, const char *p, STRLEN len)
    __attribute__((unused));
static SV *goperl_newSVpvn(goperl_frame_t *f, const char *p, STRLEN len) {
    if (!p) return GOPERL_SV(goperl_do_op(f, GOPERL_OP_NEW_SV, 0, 0, 0, 0));
    goperl_flush(f);
    return (SV *)(uintptr_t)goperl_direct_new_pvn(f, p, (uint64_t)len);
}
#define newSVpvn(p, l) goperl_newSVpvn(_gof, (p), (STRLEN)(l))
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
    if (!s) return GOPERL_SV(goperl_do_op(f, GOPERL_OP_NEW_SV, 0, 0, 0, 0));
    if (len == 0) len = __builtin_strlen(s);
    goperl_flush(f);
    return (SV *)(uintptr_t)goperl_direct_new_pvn(f, s, (uint64_t)len);
}
#define newSVpv(s, l) goperl_newSVpv(_gof, (s), (STRLEN)(l))

static SV *goperl_newSVpvf(goperl_frame_t *f, const char *fmt, ...)
    __attribute__((unused));
static SV *goperl_newSVpvf(goperl_frame_t *f, const char *fmt, ...) {
    va_list ap;
    va_start(ap, fmt);
    uint64_t sv = goperl_vtext_sv(f, 0, fmt, ap);
    va_end(ap);
    return (SV *)(uintptr_t)sv;
}
#define newSVpvf(...) goperl_newSVpvf(_gof, __VA_ARGS__)

static SV *goperl_vnewSVpvf(goperl_frame_t *f, const char *fmt, va_list *ap)
    __attribute__((unused));
static SV *goperl_vnewSVpvf(goperl_frame_t *f, const char *fmt, va_list *ap) {
    goperl_text_t t;
    goperl_text_init(&t);
    goperl_vtext_p(f, &t, fmt, ap);
    return GOPERL_SV(goperl_text_op(f, &t, GOPERL_OP_SV_STR, 0, GOPERL_TEXT_ARG(&t, 0), (uint64_t)(GOPERL_SVSTR_NEWPVN)));
}
#define vnewSVpvf(fmt, ap) goperl_vnewSVpvf(_gof, (fmt), (ap))

static SV *goperl_mess(goperl_frame_t *f, const char *fmt, ...)
    __attribute__((unused));
/* mess: the formatted text, and where the program is (" at FILE line N.")
 * when the text does not end with a newline, as perl's mess_sv adds it. */
static SV *goperl_mess(goperl_frame_t *f, const char *fmt, ...) {
    va_list ap;
    va_start(ap, fmt);
    uint64_t sv = goperl_vtext_sv(f, 1, fmt, ap);
    va_end(ap);
    return GOPERL_SV(goperl_do_op3(f, GOPERL_OP_CALL, sv, 1, (uint64_t)(7 /* MESS_SV */), 0, 0));
}
#define mess(...) goperl_mess(_gof, __VA_ARGS__)
/* vmess: mess with the arguments of a va_list * (or NULL). */
static SV *goperl_vmess(goperl_frame_t *f, const char *fmt, va_list *ap)
    __attribute__((unused));
static SV *goperl_vmess(goperl_frame_t *f, const char *fmt, va_list *ap) {
    goperl_text_t t;
    goperl_text_init(&t);
    goperl_vtext_p(f, &t, fmt, ap);
    const uint64_t sv = goperl_text_op(f, &t, GOPERL_OP_SV_STR, 0, GOPERL_TEXT_ARG(&t, GOPERL_TEXT_MORTAL), (uint64_t)(GOPERL_SVSTR_NEWPVN));
    return GOPERL_SV(goperl_do_op3(f, GOPERL_OP_CALL, sv, 1, (uint64_t)(7 /* MESS_SV */), 0, 0));
}
#define vmess(fmt, ap) goperl_vmess(_gof, (fmt), (ap))

#define newSVpvn_share(p, l, h) \
    ({ GOPERL_ONCE(goperl_newSVpvn_share_l_, l); \
       GOPERL_SV(goperl_ops(GOPERL_OP_NEW_PVN_SHARE, 0, (uint64_t)(int64_t)(goperl_newSVpvn_share_l_), (p), (uint64_t)((int64_t)(goperl_newSVpvn_share_l_) < 0 ? -(int64_t)(goperl_newSVpvn_share_l_) : (int64_t)(goperl_newSVpvn_share_l_)))); })
#define newSVpvs_share(s) newSVpvn_share("" s "", sizeof(s) - 1, 0)

static SV *goperl_newSV_type(goperl_frame_t *f, svtype t) {
    SV *sv = GOPERL_SV(goperl_do_op(f, GOPERL_OP_NEW_SV, 0, 0, 0, 0));
    if (t != SVt_NULL)
        goperl_do_op(f, GOPERL_OP_SV_UPGRADE, GOPERL_TOK(sv), (uint64_t)t, 0, 0);
    return sv;
}
#define newSV_type(t) goperl_newSV_type(_gof, (t))
/* An IO, as sv.h makes one (blessed into IO::File by the upgrade). */
#define newIO() ((IO *)newSV_type(SVt_PVIO))

static SV *goperl_newSVpvs_flags_(goperl_frame_t *f, const char *s,
                                  STRLEN len, U32 flags) {
    goperl_flush(f);
    uint64_t tok = goperl_direct_new_pvn(f, s, (uint64_t)len);
    if (flags & SVs_TEMP) tok = goperl_direct_sv_mortal(f, tok);
    return GOPERL_SV(tok);
}
#define newSVpvs_flags(s, flags) \
    goperl_newSVpvs_flags_(_gof, "" s "", sizeof(s) - 1, (flags))

/* ---- setters / string ops ---------------------------------------------- */

#define sv_setsv(d, s) \
    ((void)goperl_op0(GOPERL_OP_SV_SETSV, GOPERL_TOK(d), GOPERL_TOK(s)))
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
/* sv_setpvn and sv_setpv: a null pointer makes the SV undefined, as in
 * perl (xsubpp sets a `char *` return value with sv_setpv, and a function
 * that returns NULL is how XS code returns undef). */
static void goperl_sv_setpvn(goperl_frame_t *f, SV *sv, const char *p,
                             STRLEN len) __attribute__((unused));
static void goperl_sv_setpvn(goperl_frame_t *f, SV *sv, const char *p,
                             STRLEN len) {
    if (!p) { /* GOPERL_SVCALL_SET_UNDEF */
        goperl_do_op3(f, GOPERL_OP_SV_CALL, GOPERL_TOK(sv), 0, (uint64_t)(19), 0, 0);
        return;
    }
    goperl_do_op(f, GOPERL_OP_SV_SETPVN, GOPERL_TOK(sv), (uint64_t)len, p,
                 (uint64_t)len);
}
static void goperl_sv_setpv(goperl_frame_t *f, SV *sv, const char *p)
    __attribute__((unused));
static void goperl_sv_setpv(goperl_frame_t *f, SV *sv, const char *p) {
    goperl_sv_setpvn(f, sv, p, p ? __builtin_strlen(p) : 0);
}
#define sv_setpvn(sv, p, l) \
    goperl_sv_setpvn(_gof, (SV *)(sv), (p), (STRLEN)(l))
#define sv_setpv(sv, p) goperl_sv_setpv(_gof, (SV *)(sv), (p))
#define sv_setpvs(sv, s) sv_setpvn((sv), "" s "", sizeof(s) - 1)
static void goperl_sv_setpvf(goperl_frame_t *f, SV *sv, const char *fmt, ...)
    __attribute__((unused));
static void goperl_sv_setpvf(goperl_frame_t *f, SV *sv, const char *fmt,
                             ...) {
    goperl_text_t t;
    va_list ap;
    goperl_text_init(&t);
    va_start(ap, fmt);
    goperl_vtext(f, &t, fmt, ap);
    va_end(ap);
    goperl_text_op(f, &t, GOPERL_OP_SV_STR, GOPERL_TOK(sv), GOPERL_TEXT_ARG(&t, 0), (uint64_t)(GOPERL_SVSTR_SETPVN));
}
#define sv_setpvf(sv, ...) goperl_sv_setpvf(_gof, (SV *)(sv), __VA_ARGS__)
static void goperl_sv_catpvf(goperl_frame_t *f, SV *sv, const char *fmt, ...)
    __attribute__((unused));
static void goperl_sv_catpvf(goperl_frame_t *f, SV *sv, const char *fmt,
                             ...) {
    goperl_text_t t;
    va_list ap;
    goperl_text_init(&t);
    va_start(ap, fmt);
    goperl_vtext(f, &t, fmt, ap);
    va_end(ap);
    /* appended as what it is: sv_catpvn_flags upgrades whichever of the
     * two sides is not UTF-8 when the other is */
    goperl_text_op(f, &t, GOPERL_OP_SV_STR, GOPERL_TOK(sv),
                   ((uint64_t)(uint32_t)(SV_GMAGIC | (t.utf8 ? SV_CATUTF8 : SV_CATBYTES)) << 32) | (uint32_t)t.len,
                   0 /* GOPERL_SVSTR_CATPVN */);
}
#define sv_catpvf(sv, ...) goperl_sv_catpvf(_gof, (SV *)(sv), __VA_ARGS__)
#define sv_catsv(d, s) \
    ((void)goperl_op0(GOPERL_OP_SV_CATSV, GOPERL_TOK(d), GOPERL_TOK(s)))
#define sv_catpvn(d, p, l) \
    ({ GOPERL_ONCE(goperl_sv_catpvn_l_, l); \
       ((void)goperl_ops(GOPERL_OP_SV_CATPVN, GOPERL_TOK(d), (uint64_t)(goperl_sv_catpvn_l_), (p), (uint64_t)(goperl_sv_catpvn_l_))); })
/* sv_catpv: nothing is appended for a null pointer, as in perl. */
static void goperl_sv_catpv(goperl_frame_t *f, SV *sv, const char *p)
    __attribute__((unused));
static void goperl_sv_catpv(goperl_frame_t *f, SV *sv, const char *p) {
    if (!p) return;
    goperl_do_op(f, GOPERL_OP_SV_CATPVN, GOPERL_TOK(sv), (uint64_t)__builtin_strlen(p),
                 p, (uint64_t)__builtin_strlen(p));
}
#define sv_catpv(d, p) goperl_sv_catpv(_gof, (SV *)(d), (p))
#define sv_catpvs(d, s) sv_catpvn((d), "" s "", sizeof(s) - 1)

/* ---- flags / predicates -------------------------------------------------- */

#define GOPERL_SVINFO(sv) \
    ((U32)goperl_op0(GOPERL_OP_SV_INFO, GOPERL_TOK(sv), 0))
/* SvFLAGS is the interpreter's real flag word, and an lvalue as on perl
 * (`SvFLAGS(sv) |= SVf_BREAK`, `SvFLAGS(av) = saved_flags`): what is stored
 * into it is the whole word, as it is. A U32: the low half of the slot. */
#define SvFLAGS(sv) \
    (*(U32 *)goperl_lv_ref(_gof, (SV *)(sv), GOPERL_LV_FLAGS, GOPERL_LV_AT))
#define SvOK(sv) ((GOPERL_SVINFO(sv) & GOPERL_INFO_OK) != 0)
#define SvPOK(sv) ((GOPERL_SVINFO(sv) & GOPERL_INFO_POK) != 0)
#define SvIOK(sv) ((GOPERL_SVINFO(sv) & GOPERL_INFO_IOK) != 0)
#define SvNOK(sv) ((GOPERL_SVINFO(sv) & GOPERL_INFO_NOK) != 0)
#define SvNIOK(sv) ((GOPERL_SVINFO(sv) & GOPERL_INFO_NIOK) != 0)
#define SvPOKp(sv) ((GOPERL_SVINFO(sv) & GOPERL_INFO_POKp) != 0)
#define SvNIOKp(sv) ((GOPERL_SVINFO(sv) & GOPERL_INFO_NIOKp) != 0)
#define SvIsUV(sv) ((GOPERL_SVINFO(sv) & GOPERL_INFO_ISUV) != 0)
/* SvUTF8 is the flag itself, as on perl: code passes it on in the flags of
 * a call (gv_stashpvn(name, len, GV_ADD | SvUTF8(sv)), newSVpvn_flags(p,
 * len, SvUTF8(sv) | SVs_TEMP)). */
#define SvUTF8(sv) ((GOPERL_SVINFO(sv) & GOPERL_INFO_UTF8) ? (U32)SVf_UTF8 : (U32)0)
#define SvREADONLY(sv) ((GOPERL_SVINFO(sv) & GOPERL_INFO_READONLY) != 0)
#define SvOBJECT(sv) ((GOPERL_SVINFO(sv) & GOPERL_INFO_OBJECT) != 0)
#define SvRMAGICAL(sv) ((GOPERL_SVINFO(sv) & GOPERL_INFO_RMAGICAL) != 0)
/* sv_isobject runs the get magic of its argument, as perl's does; the flag
 * query itself runs none. */
#define sv_isobject(sv)                                    \
    ({ GOPERL_ONCE(goperl_sv_isobject_sv_, sv);            \
       int goperl_sv_isobject_ = 0;                        \
       if (goperl_sv_isobject_sv_) {                       \
           SvGETMAGIC(goperl_sv_isobject_sv_);             \
           goperl_sv_isobject_ =                           \
               (GOPERL_SVINFO(goperl_sv_isobject_sv_) & GOPERL_INFO_ISOBJ) != 0; \
       }                                                   \
       goperl_sv_isobject_; })
#define SvTRUE(sv) (goperl_op0(GOPERL_OP_SV_TRUE, GOPERL_TOK(sv), 0) != 0)
#define sv_true(sv) SvTRUE(sv)
#define SvROK(sv) (goperl_op0(GOPERL_OP_SV_RV, GOPERL_TOK(sv), 0) != 0)
/* SvRV is an lvalue (`SvRV(rv) = sv`, with SvROK_on before or after). */
#define SvRV(sv) \
    (*(SV **)goperl_lv_ref(_gof, (SV *)(sv), GOPERL_LV_RV, GOPERL_LV_AT))
#define SvTYPE(sv) \
    ((svtype)goperl_op0(GOPERL_OP_SV_TYPE, GOPERL_TOK(sv), 0))
#define isGV(sv) (SvTYPE((SV *)(sv)) == SVt_PVGV)
#define SvUTF8_on(sv) \
    ((void)goperl_op0(GOPERL_OP_SV_UTF8_ON, GOPERL_TOK(sv), 0))
#define SvPOK_on(sv) \
    ((void)goperl_op0(GOPERL_OP_SV_POK_ON, GOPERL_TOK(sv), 0))
/* SvUPGRADE only ever raises the type; sv_upgrade itself refuses to lower
 * it. */
#define SvUPGRADE(sv, t) \
    ({ GOPERL_ONCE(goperl_SvUPGRADE_sv_, sv); \
       GOPERL_ONCE(goperl_SvUPGRADE_t_, t); \
       (void)(SvTYPE(goperl_SvUPGRADE_sv_) >= goperl_SvUPGRADE_t_ || (goperl_op0(GOPERL_OP_SV_UPGRADE, GOPERL_TOK(goperl_SvUPGRADE_sv_), (uint64_t)goperl_SvUPGRADE_t_), 1)); })
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
    ({ GOPERL_ONCE(goperl_sv_rvweaken_sv_, sv); \
       ((void)goperl_op0(GOPERL_OP_SV_RVWEAKEN, GOPERL_TOK(goperl_sv_rvweaken_sv_), 0), (SV *)(goperl_sv_rvweaken_sv_)); })
#define sv_dump(sv) ((void)goperl_op0(GOPERL_OP_SV_DUMP, GOPERL_TOK(sv), 0))
#define looks_like_number(sv) \
    ((I32)goperl_op0(GOPERL_OP_LOOKS_LIKE_NUMBER, GOPERL_TOK(sv), 0))
#define SvAMAGIC(sv) \
    (goperl_op0(GOPERL_OP_SV_AMAGIC, GOPERL_TOK(sv), 0) != 0)
/* The stash of an object, an lvalue as on perl (`SvSTASH(sv) = NULL`): the
 * store is the slot's, as SvSTASH_set's is (no reference is counted). */
#define SvSTASH(sv) \
    (*(HV **)goperl_lv_ref(_gof, (SV *)(sv), GOPERL_LV_STASH, GOPERL_LV_AT))

/* SvREFCNT_inc and SvREFCNT_dec take any pointer: on perl the macro casts
 * its argument to an SV. A module can pass a pointer to a structure of its
 * own (Marpa::R2 counts the users of its recognizer structure that way),
 * and perl then counts in that structure's memory, in the word where the
 * head of an SV has its count. Such a pointer is not a value of the
 * interpreter and must never reach it as one: its type says what it is,
 * and the count is kept where perl would keep it. */
typedef struct goperl_foreign_head {
    void *any;
    U32 refcnt;
    U32 flags;
} goperl_foreign_head_t;
static inline __attribute__((unused)) SV *goperl_foreign_refcnt_inc(void *p) {
    if (p) ((goperl_foreign_head_t *)p)->refcnt++;
    return (SV *)p;
}
/* perl goes on to free an SV whose count ends; there is no SV here to
 * free, so the count just ends. */
static inline __attribute__((unused)) void goperl_foreign_refcnt_dec(void *p) {
    if (p && ((goperl_foreign_head_t *)p)->refcnt > 0)
        ((goperl_foreign_head_t *)p)->refcnt--;
}
#ifdef __cplusplus
extern "C++" { /* the header is usually included inside extern "C" */
template <class T> static inline int goperl_is_value_ptr(T) { return 1; }
template <class T> static inline int goperl_is_value_ptr(T *) { return 0; }
static inline int goperl_is_value_ptr(struct goperl_sv *) { return 1; }
static inline int goperl_is_value_ptr(const struct goperl_sv *) { return 1; }
static inline int goperl_is_value_ptr(void *) { return 1; }
static inline int goperl_is_value_ptr(const void *) { return 1; }
}
#define GOPERL_IS_VALUE_PTR(p) goperl_is_value_ptr(p)
#else
#define GOPERL_IS_VALUE_PTR(p)           \
    _Generic((p), struct goperl_sv *: 1, \
             const struct goperl_sv *: 1, void *: 1, const void *: 1, default: 0)
#endif
#define SvREFCNT_inc(sv) \
    ({ GOPERL_ONCE(goperl_SvREFCNT_inc_sv_, sv); \
       GOPERL_IS_VALUE_PTR(goperl_SvREFCNT_inc_sv_) \
           ? GOPERL_SV(goperl_op0(GOPERL_OP_REFCNT_INC, GOPERL_TOK(goperl_SvREFCNT_inc_sv_), 0)) \
           : goperl_foreign_refcnt_inc((void *)(uintptr_t)(goperl_SvREFCNT_inc_sv_)); })
#define SvREFCNT_inc_NN(sv) SvREFCNT_inc(sv)
#define SvREFCNT_inc_simple(sv) SvREFCNT_inc(sv)
#define SvREFCNT_inc_simple_NN(sv) SvREFCNT_inc(sv)
#define SvREFCNT_inc_void(sv) ((void)SvREFCNT_inc(sv))
#define SvREFCNT_inc_simple_void(sv) ((void)SvREFCNT_inc(sv))
#define SvREFCNT_inc_simple_void_NN(sv) ((void)SvREFCNT_inc(sv))
#define SvREFCNT_dec(sv) \
    ({ GOPERL_ONCE(goperl_SvREFCNT_dec_sv_, sv); \
       if (GOPERL_IS_VALUE_PTR(goperl_SvREFCNT_dec_sv_)) \
           (void)goperl_op0(GOPERL_OP_SV_REFCNT_DEC, GOPERL_TOK(goperl_SvREFCNT_dec_sv_), 0); \
       else \
           goperl_foreign_refcnt_dec((void *)(uintptr_t)(goperl_SvREFCNT_dec_sv_)); \
       (void)0; })
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

/* amagic_call: the overloaded operator of two operands (the right one is
 * undef for the operators of one). NULL when neither overloads it. */
#define amagic_call(left, right, method, flags)                             \
    GOPERL_SV(goperl_op3(GOPERL_OP_SV_CALL, GOPERL_TOK(left), GOPERL_TOK(right), \
                         61 /* GOPERL_SVCALL_AMAGIC */ |                      \
                             ((uint64_t)(((uint32_t)(method) & 0xff) |        \
                                         (((uint32_t)(flags) & 0xffff) << 8)) \
                              << 32)))
#define amagic_deref_call(ref, method) \
    goperl_amagic_deref_call(_gof, (SV *)(ref), (method))
static SV *goperl_amagic_deref_call(goperl_frame_t *f, SV *ref, int method)
    __attribute__((unused));
static SV *goperl_amagic_deref_call(goperl_frame_t *f, SV *ref, int method) {
    SV *tmpsv = 0;
    for (;;) {
        if (!goperl_do_op(f, GOPERL_OP_SV_AMAGIC, GOPERL_TOK(ref), 0, 0, 0))
            break;
        SV *r = GOPERL_SV(goperl_do_op3(f, GOPERL_OP_AMAGIC_CALL, GOPERL_TOK(ref),
                                        (uint64_t)(AMGf_noright | AMGf_unary),
                                        (uint64_t)(int64_t)method, 0, 0));
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
/* AvFILLp is the array's own index of its last element, and an lvalue
 * (`AvFILLp(av) = -1` empties an array whose elements the code has taken
 * out itself). */
#define AvFILLp(av) \
    (*(SSize_t *)goperl_lv_ref(_gof, (SV *)(av), GOPERL_LV_AVFILLP, GOPERL_LV_AT))
#define av_extend(av, n) \
    ((void)goperl_op0(GOPERL_OP_AV_EXTEND, GOPERL_TOK(av), (uint64_t)(int64_t)(n)))
#define av_fill(av, n) \
    ((void)goperl_op0(GOPERL_OP_AV_FILL, GOPERL_TOK(av), (uint64_t)(int64_t)(n)))
/* The slot an SV** result (av_fetch, hv_fetch, av_store, ...) points at:
 * one of its own for every result, which holds its value until the native
 * call returns, as perl's points into the array or hash for as long as
 * that is not changed (Number::Closest::XS keeps the slots of the closest
 * elements while it fetches the rest). The first slots are the frame's
 * own, the next ones are in blocks the loader frees with the frame. */
static SV **goperl_tmp_slot(goperl_frame_t *f, uint64_t tok) {
    const int32_t i = f->tmpidx;
    uint64_t *slot;
    if (i < GOPERL_XS_TMPS) {
        slot = &f->tmp[i];
    } else {
        const int32_t off = (i - GOPERL_XS_TMPS) % GOPERL_TMP_BLOCK;
        if (i == INT32_MAX) goperl_croakf(f, "Out of memory!");
        if (off == 0) {
            goperl_tmpblock_t *b = (goperl_tmpblock_t *)malloc(sizeof(goperl_tmpblock_t));
            if (!b) goperl_croakf(f, "Out of memory!");
            b->prev = f->tmp_more;
            f->tmp_more = b;
        }
        slot = &f->tmp_more->slot[off];
    }
    f->tmpidx = i + 1;
    *slot = tok;
    return (SV **)slot;
}

/* The SV** of av_fetch, hv_fetch, av_store and hv_store is the slot perl's
 * gave: the element's own in the array, HeVAL of the entry, a field of the
 * temporary perl makes for a tied container. It is in the interpreter's
 * memory, so what native code stores through it is the container's (an
 * element swapped or replaced, `*svp = sv`, without a reference counted, as
 * on perl), and what it reads through it is what the container holds now,
 * after an av_store of the same element too. It is good for as long as
 * perl's would be: until the container's storage is reallocated or freed.
 * A slot the operation could not give stays a temporary of the call's
 * (goperl_tmp_slot). */
static SV **goperl_live_slot(goperl_frame_t *f, uint64_t tok, uint64_t addr) {
    SV **slot = addr ? (SV **)goperl_guest_mem(f, addr) : 0;
    return slot ? slot : goperl_tmp_slot(f, tok);
}

/* av_store returns the slot when the array took the value and NULL when
 * it did not (a tied array hands it to STORE): callers release the value
 * themselves on NULL, so the result has to tell the two apart. */
static SV **goperl_av_store(goperl_frame_t *f, AV *av, SSize_t i, SV *sv) {
    uint64_t tok, addr = 0;
    goperl_do_op3(f, GOPERL_OP_AV_STORE, GOPERL_TOK(av), GOPERL_TOK(sv), (uint64_t)(int64_t)i,
                  0, 0);
    tok = goperl_do_opx(f, GOPERL_OP_AV_FETCH, GOPERL_TOK(av), (uint64_t)(int64_t)i, 0, 0, 0,
                        &addr);
    if (tok != GOPERL_TOK(sv)) return 0;
    return goperl_live_slot(f, tok, addr);
}
#define av_store(av, i, sv) \
    goperl_av_store(_gof, (AV *)(av), (SSize_t)(i), (SV *)(sv))

static SV **goperl_av_fetch(goperl_frame_t *f, AV *av, SSize_t i, I32 lval) {
    uint64_t addr = 0;
    uint64_t tok = goperl_do_opx(f, lval ? GOPERL_OP_AV_READ : GOPERL_OP_AV_FETCH, GOPERL_TOK(av),
                                 (uint64_t)(int64_t)i, 0, 0, 0, &addr);
    if (!tok) return 0;
    return goperl_live_slot(f, tok, addr);
}
#define av_fetch(av, i, lval) goperl_av_fetch(_gof, (AV *)(av), (i), (lval))

#define av_make(n, ary) goperl_av_make(_gof, (SSize_t)(n), (ary))
static AV *goperl_av_make(goperl_frame_t *f, SSize_t n, SV **ary) {
    goperl_gword_t buf[64];
    if (n < 0 || n > 64)
        goperl_croakf(f, "av_make: %lld elements exceed the SDK limit",
                      (long long)n);
    for (SSize_t i = 0; i < n; i++) buf[i] = GOPERL_TOK(ary[i]);
    return (AV *)GOPERL_SV(goperl_do_op(f, GOPERL_OP_AV_MAKE, 0,
                                        (uint64_t)(n * 8), (const char *)buf,
                                        (uint64_t)(n * 8)));
}

/* The length of a key (any bytes, so its length crosses with it) and
 * whether it is UTF-8, as the keyed hash operations take them in c. */
#define GOPERL_HV_KLEN(utf8, klen) \
    (((uint64_t)((utf8) ? 1 : 0) << 32) | (uint32_t)(klen))
static SV **goperl_hv_fetch(goperl_frame_t *f, HV *hv, const char *key,
                            I32 klen, I32 lval) {
    int utf8 = klen < 0;
    if (utf8) klen = -klen;
    uint64_t addr = 0;
    uint64_t tok = goperl_do_opx(f, GOPERL_OP_HV_FETCH_KLEN, GOPERL_TOK(hv), lval ? 1 : 0,
                                 GOPERL_HV_KLEN(utf8, klen), key, (uint64_t)(uint32_t)klen, &addr);
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
    return goperl_live_slot(f, tok, addr);
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
    uint64_t addr = 0;
    uint64_t tok = goperl_do_opx(f, GOPERL_OP_HV_STORE_KLEN, GOPERL_TOK(hv), GOPERL_TOK(sv),
                                 GOPERL_HV_KLEN(utf8, klen), key, (uint64_t)(uint32_t)klen, &addr);
    if (!tok) return 0;
    return goperl_live_slot(f, tok, addr);
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
 * hv_iterval live in the section on entries of hashes further down. */
#define HvNAME(hv) \
    ((char *)goperl_intern_op(_gof, GOPERL_OP_HV_NAME, GOPERL_TOK(hv), 0, 0))
#define HvNAME_get(hv) HvNAME(hv)

/* ---- stashes / globals / blessing --------------------------------------- */

#define gv_stashpv(name, flags) \
    ({ GOPERL_ONCE(goperl_gv_stashpv_name_, name); \
       ((HV *)GOPERL_SV(goperl_ops(GOPERL_OP_GV_STASHPV, (uint64_t)(int64_t)(flags), 0, (goperl_gv_stashpv_name_), __builtin_strlen(goperl_gv_stashpv_name_)))); })
#define sv_bless(rv, stash) \
    GOPERL_SV(goperl_op0(GOPERL_OP_SV_BLESS, GOPERL_TOK(rv), GOPERL_TOK(stash)))
#define sv_isa(sv, name) \
    ({ GOPERL_ONCE(goperl_sv_isa_name_, name); \
       ((int)goperl_ops(GOPERL_OP_SV_ISA, GOPERL_TOK(sv), 0, (goperl_sv_isa_name_), __builtin_strlen(goperl_sv_isa_name_))); })
#define sv_derived_from(sv, name) \
    ({ GOPERL_ONCE(goperl_sv_derived_from_name_, name); \
       ((int)goperl_ops(GOPERL_OP_SV_DERIVED_FROM, GOPERL_TOK(sv), 0, (goperl_sv_derived_from_name_), __builtin_strlen(goperl_sv_derived_from_name_))); })
#define get_sv(name, flags) \
    ({ GOPERL_ONCE(goperl_get_sv_name_, name); \
       GOPERL_SV(goperl_ops(GOPERL_OP_GET_SV, (uint64_t)(int64_t)(flags), 0, (goperl_get_sv_name_), __builtin_strlen(goperl_get_sv_name_))); })
#define get_cv(name, flags) \
    ({ GOPERL_ONCE(goperl_get_cv_name_, name); \
       ((CV *)GOPERL_SV(goperl_ops(GOPERL_OP_GET_CV, (uint64_t)(int64_t)(flags), 0, (goperl_get_cv_name_), __builtin_strlen(goperl_get_cv_name_)))); })
#define get_cvs(name, flags) get_cv("" name "", (flags))
#define get_hv(name, flags) \
    ({ GOPERL_ONCE(goperl_get_hv_name_, name); \
       ((HV *)GOPERL_SV(goperl_ops(GOPERL_OP_GET_HV, (uint64_t)(int64_t)(flags), 0, (goperl_get_hv_name_), __builtin_strlen(goperl_get_hv_name_)))); })
#define gv_fetchpv(name, flags, svt) \
    ({ GOPERL_ONCE(goperl_gv_fetchpv_name_, name); \
       ((GV *)GOPERL_SV(goperl_ops(GOPERL_OP_GV_FETCH, (uint64_t)(int64_t)(flags), (uint64_t)(int64_t)(svt), (goperl_gv_fetchpv_name_), __builtin_strlen(goperl_gv_fetchpv_name_)))); })
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

/* Pointers in IVs are plain casts, as in perl's own headers (see the note
 * on the integer geometry below for how the 64 bits cross). */
static SV *goperl_sv_setref_iv(goperl_frame_t *f, SV *rv,
                               const char *classname, int64_t iv)
    __attribute__((unused));
static SV *goperl_sv_setref_iv(goperl_frame_t *f, SV *rv,
                               const char *classname, int64_t iv) {
    /* newSVrv (selector 1 of SV_STR; flag bit 0: no class), then the value */
    uint64_t sv = goperl_do_op3(f, GOPERL_OP_SV_STR, GOPERL_TOK(rv), ((uint64_t)(classname ? 0 : 1) << 32) |
            (uint32_t)(classname ? __builtin_strlen(classname) : 0), (uint64_t)(1), classname ? classname : "", classname ? __builtin_strlen(classname) : 0);
    goperl_do_op(f, GOPERL_OP_SV_SETIV, sv, (uint64_t)iv, 0, 0);
    return rv;
}
static SV *goperl_sv_setref_pv(goperl_frame_t *f, SV *rv,
                               const char *classname, void *ptr)
    __attribute__((unused));
static SV *goperl_sv_setref_pv(goperl_frame_t *f, SV *rv,
                               const char *classname, void *ptr) {
    if (!ptr) { /* a null pointer makes rv undefined, not a reference */
        goperl_do_op(f, GOPERL_OP_SV_SETSV, GOPERL_TOK(rv), f->imm[0], 0, 0);
        return rv;
    }
    return goperl_sv_setref_iv(f, rv, classname, (int64_t)(intptr_t)ptr);
}
#define sv_setref_pv(rv, classname, ptr) \
    goperl_sv_setref_pv(_gof, (SV *)(rv), (classname), (void *)(ptr))
#define sv_setref_iv(rv, classname, iv) \
    goperl_sv_setref_iv(_gof, (SV *)(rv), (classname), (int64_t)(iv))
/* A pointer as a number and back. An op of the interpreter's is a number
 * as on perl, its address in the interpreter, with GOPERL_OP_NUMBER_TAG set:
 * the number of the op's B::OP object too (B is the interpreter's, and
 * gives the same numbers). So a module that keeps an op as a number, or
 * takes the op of a B object (B::Utils, Devel::Cover's report), gets the
 * op's shadow back from INT2PTR. An op of a module's own that the
 * interpreter has no copy of (one on the C stack, in the C heap) is its
 * address here, made a number by PTR2IV or by a cast, as is a shadow cast
 * to a number: no address of this process has the tag, wherever the
 * process puts its memory, so INT2PTR tells the two apart by the number
 * alone. Defined below, with the op shadows. */
#define GOPERL_OP_NUMBER_TAG ((UV)1 << 62)
static IV goperl_op_number(const OP *o);
static OP *goperl_op_of_number(IV iv);
#ifdef __cplusplus
extern "C++" {
template <class T> static inline IV goperl_ptr_number(T p) { return (IV)(intptr_t)p; }
static inline IV goperl_ptr_number(OP *o) { return goperl_op_number(o); }
static inline IV goperl_ptr_number(const OP *o) { return goperl_op_number(o); }
template <class T> struct goperl_number_ptr {
    static T of(IV iv) { return (T)(uintptr_t)iv; }
};
template <> struct goperl_number_ptr<OP *> {
    static OP *of(IV iv) { return goperl_op_of_number(iv); }
};
template <> struct goperl_number_ptr<const OP *> {
    static const OP *of(IV iv) { return goperl_op_of_number(iv); }
};
}
#define GOPERL_PTR_NUMBER(p) goperl_ptr_number(p)
#define INT2PTR(type, iv) (goperl_number_ptr<type>::of((IV)(iv)))
#else
#define GOPERL_PTR_NUMBER(p)                                        \
    _Generic((p), OP *: goperl_op_number((const OP *)(uintptr_t)(p)), \
             const OP *: goperl_op_number((const OP *)(uintptr_t)(p)), \
             default: (IV)(intptr_t)(p))
#define INT2PTR(type, iv)                                           \
    _Generic((type)0, OP *: (type)(uintptr_t)goperl_op_of_number((IV)(iv)), \
             const OP *: (type)(uintptr_t)goperl_op_of_number((IV)(iv)), \
             default: (type)(uintptr_t)(iv))
#endif
#define PTR2IV(p) GOPERL_PTR_NUMBER(p)
#define PTR2UV(p) ((UV)GOPERL_PTR_NUMBER(p))

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
#define ERRSV (*goperl_plsv_ref(_gof, GOPERL_PL_ERRSV))

/* PL_diehook / PL_warnhook: read-through, write-back shadows. Reads always
 * consult the guest unless an assignment is pending; assignments are
 * flushed before the next guest operation (goperl_flush). */
static SV **goperl_hook_ref(goperl_frame_t *f, int which /* 0 die, 1 warn */) {
    if (!(f->hook_dirty & (1 << which)))
        f->hook_val[which] = goperl_xs_op(
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

/* PL_op is the op the interpreter is running: for a pp function the op it
 * was called for, which the frame has, and for an XSUB the call that called
 * it, which is asked for when the XSUB looks (Class::XSAccessor gives the
 * call a pp function of its own from inside the accessor). It is an
 * lvalue: a module points it at an op of its own around a call of a pp
 * function. */
static OP *goperl_op_shadow(goperl_frame_t *f, uint64_t tok);
static OP **goperl_plop_ref(goperl_frame_t *f) __attribute__((unused));
static OP **goperl_plop_ref(goperl_frame_t *f) {
    if (!f->plop) {
        const uint64_t tok =
            goperl_do_op(f, GOPERL_OP_PLVAR_GET, (uint64_t)GOPERL_PL_OP, 0, 0, 0);
        if (tok) f->plop = goperl_op_shadow(f, tok);
    }
    return &f->plop;
}
#define PL_op (*goperl_plop_ref(_gof))

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
 * Time::U2time) once Time::HiRes is loaded, which consumers call as raw
 * function pointers — a guest pointer is uncallable from the host, so the
 * first host access after Time::HiRes put them there replaces those two
 * entries with HOST implementations of the same wallclock. Before that
 * they are not there, as with perl (UUID requires Time::HiRes and looks). */
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
    if (!GOPERL_SHARED->modglobal_seeded && *ref &&
        goperl_do_op3(f, GOPERL_OP_HV_FETCH_KLEN, GOPERL_TOK(*ref), 0,
                      GOPERL_HV_KLEN(0, sizeof("Time::U2time") - 1), "Time::U2time",
                      sizeof("Time::U2time") - 1)) {
        GOPERL_SHARED->modglobal_seeded = 1;
        uint64_t iv;
        iv = goperl_do_op(
            f, GOPERL_OP_NEW_IV,
            (uint64_t)(uintptr_t)goperl_nvtime_fn, 0, 0, 0);
        goperl_do_op3(f, GOPERL_OP_HV_STORE_KLEN, GOPERL_TOK(*ref), iv,
                      GOPERL_HV_KLEN(0, sizeof("Time::NVtime") - 1), "Time::NVtime",
                      sizeof("Time::NVtime") - 1);
        iv = goperl_do_op(
            f, GOPERL_OP_NEW_IV,
            (uint64_t)(uintptr_t)goperl_u2time_fn, 0, 0, 0);
        goperl_do_op3(f, GOPERL_OP_HV_STORE_KLEN, GOPERL_TOK(*ref), iv,
                      GOPERL_HV_KLEN(0, sizeof("Time::U2time") - 1), "Time::U2time",
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

#define GOPERL_AVBODY_OF(m) \
    ((goperl_avbody_t *)((char *)(uintptr_t)(m)->val - \
                         offsetof(goperl_avbody_t, slots)))

/* The elements of an array's body that were stored into go to the
 * interpreter (as they are: the store does not count a reference). */
static void goperl_avbody_store(goperl_frame_t *f, goperl_lvslot_t *m) {
    goperl_avbody_t *b = GOPERL_AVBODY_OF(m);
    for (int32_t i = 0; i < b->len; i++) {
        if (b->slots[i] == b->shadow[i]) continue;
        b->shadow[i] = b->slots[i];
        goperl_xs_opx(f, GOPERL_OP_AV_STORE_RAW, m->svtok, b->slots[i], (uint64_t)i, 0, 0, 0);
    }
}

/* A slot is given up: the body it holds, if it is an array's or a sub's. */
static void goperl_lvbody_release(goperl_lvslot_t *m);
static void goperl_lvslot_release(goperl_lvslot_t *m) {
    if (GOPERL_LV_IS_BODY(m->field)) {
        goperl_lvbody_release(m);
        return;
    }
    if (m->field != GOPERL_LV_AVBODY || !m->val) return;
    goperl_avbody_t *b = GOPERL_AVBODY_OF(m);
    free(b->shadow);
    free(b);
    m->val = m->shadow = 0;
}

static void goperl_lvbody_store(goperl_frame_t *f, goperl_lvslot_t *m);
static void goperl_lvbody_load(goperl_frame_t *f, goperl_lvslot_t *m);
static void goperl_lvslot_store(goperl_frame_t *f, goperl_lvslot_t *m) {
    if (m->field == GOPERL_LV_AVBODY) {
        if (m->val) goperl_avbody_store(f, m);
        return;
    }
    if (GOPERL_LV_IS_BODY(m->field)) {
        goperl_lvbody_store(f, m);
        return;
    }
    if (m->val == m->shadow) return;
    m->shadow = m->val;
    goperl_lv_store_field(f, m->svtok, m->field, m->val);
}

static void goperl_lvslot_flush(goperl_frame_t *f) {
    goperl_lvstate_t *st = goperl_lvstate_v;
    if (!st) return;
    /* a slot with several users is looked at once for each: the second
     * look finds it written back */
    for (int32_t i = 0; i < st->nuses; i++)
        goperl_lvslot_store(f, st->uses[i].slot);
    for (int32_t i = 0; i < st->nrecent; i++)
        goperl_lvslot_store(f, st->recent[i]);
}

/* Forgets every lvalue slot (the outermost native call has returned, and
 * with it every expression that could still store through one). */
static void goperl_lvslot_reset(void) {
    goperl_lvstate_t *st = goperl_lvstate_v;
    if (!st) return;
    for (int32_t i = 0; i < st->nuses; i++)
        goperl_lvslot_release(st->uses[i].slot);
    for (int32_t i = 0; i < st->nrecent; i++)
        goperl_lvslot_release(st->recent[i]);
    st->nuses = 0;
    st->nrecent = 0;
    st->free = 0;
    for (goperl_lvchunk_t *c = st->chunks; c;) {
        goperl_lvchunk_t *next = c->next;
        if (c == st->chunks) {
            for (int32_t i = 0; i < GOPERL_LVCHUNK; i++) {
                c->slots[i].free_next = st->free;
                st->free = &c->slots[i];
            }
            c->next = 0;
        } else {
            free(c);
        }
        c = next;
    }
}

/* Interpreter-variable shadows beyond the die/warn hooks: read-through on
 * every access, written back before the next guest operation. Slots are
 * indexed by the GOPERL_PL_* id. */
#define GOPERL_PLSHADOW_MAX 128
typedef struct {
    uint64_t val;
    uint64_t shadow; /* what the interpreter held when val was loaded */
    U8 live;         /* the slot has been handed out */
} goperl_plshadow_t;
#define goperl_plshadow_v ((goperl_plshadow_t *)GOPERL_SHARED->plshadow)
typedef char goperl_plshadow_fits
    [(sizeof(goperl_plshadow_t) * GOPERL_PLSHADOW_MAX <=
      sizeof(((goperl_shared_t *)0)->plshadow))
         ? 1
         : -1] __attribute__((unused));

/* What a stored value of a variable whose slot holds a shadow is to the
 * interpreter (goperl_compiler.h). */
static int goperl_plshadow_out(goperl_frame_t *f, int id, goperl_plshadow_t *s, uint64_t *out);
static void goperl_curpad_own_flush(goperl_frame_t *f);
/* A slot is handed out: the flush before every operation looks at it from
 * now on. */
static inline __attribute__((unused)) void goperl_plshadow_set_live(goperl_plshadow_t *s) {
    goperl_shared_t *const sh = GOPERL_SHARED;
    const size_t id = (size_t)(s - (goperl_plshadow_t *)sh->plshadow);
    s->live = 1;
    sh->plshadow_live[id >> 6] |= (uint64_t)1 << (id & 63);
}
static void goperl_plshadow_flush(goperl_frame_t *f) {
    goperl_curpad_own_flush(f);
    /* the frame's instance block, looked up once: this runs before every
     * operation, over the slots handed out */
    goperl_shared_t *const sh = (goperl_shared_t *)f->inst;
    goperl_plshadow_t *const tab = (goperl_plshadow_t *)sh->plshadow;
    for (int w = 0; w < 2; w++) {
        for (uint64_t m = sh->plshadow_live[w]; m; m &= m - 1) {
            const int i = w * 64 + __builtin_ctzll(m);
            goperl_plshadow_t *s = &tab[i];
            if (s->live && s->val != s->shadow) {
                uint64_t out = 0;
                if (!goperl_plshadow_out(f, i, s, &out)) continue;
                s->shadow = s->val;
                goperl_xs_op(f, GOPERL_OP_PLVAR_SET, (uint64_t)i, out, 0, 0);
            }
        }
    }
}

static void goperl_ppaddr_sync(goperl_frame_t *f);
static void goperl_hooks_sync(goperl_frame_t *f);
static void goperl_bhk_sync(goperl_frame_t *f);

#define goperl_hints_slot_v (GOPERL_SHARED->hints_slot)
#define goperl_hints_base_v (GOPERL_SHARED->hints_base)
#define goperl_hints_live_v (GOPERL_SHARED->hints_live)

static void goperl_parser_push(goperl_frame_t *f);
static void goperl_curcop_writeback(goperl_frame_t *f);
static void goperl_pads_flush(goperl_frame_t *f);
static void goperl_cx_writeback(goperl_frame_t *f);
static void goperl_statcache_writeback(goperl_frame_t *f);
/* The head of an SV's chain of magic that native code stored through
 * SvMAGIC (Event and Tk link a MAGIC of their own onto the end of a chain
 * that was empty): what the loader keeps is that head from then on.
 *
 * svmagic_orig is the head SvMAGIC showed. When that is a copy of the
 * interpreter's magic (see goperl_svmagic_lv), the head of native code's
 * chain behind the copies is kept in the zeroed MAGIC after the last copy,
 * and every type the copies have must still be on the chain: native code
 * that unlinked one meant to take the interpreter's magic off the value,
 * which this cannot do, and the call fails saying so. */
static MAGIC *goperl_magic_chain_of_native(goperl_frame_t *f, MAGIC *mg);
static int goperl_magic_is_interp_copy(goperl_frame_t *f, const MAGIC *mg);
/* Whether native code took the chain SvMAGIC showed off the value whole,
 * the interpreter's magic with it: Tk's trace of a variable sets SvMAGIC
 * to NULL, attaches a magic of its own with sv_magic (which would not add
 * a second of its type) and puts the chain back with the new magic at its
 * end. Until it is put back, the chain is native code's alone. */
static int goperl_svmagic_detached(goperl_frame_t *f) {
    goperl_shared_t *const sh = (goperl_shared_t *)f->inst;
    if (!sh || !sh->svmagic_sv || !goperl_magic_is_interp_copy(f, sh->svmagic_orig)) return 0;
    for (const MAGIC *mg = sh->svmagic_val; mg; mg = mg->mg_moremagic)
        if (goperl_magic_is_interp_copy(f, mg)) return 0;
    return 1;
}
/* The head of native code's chain on the value is head from now on. */
static void goperl_svmagic_head(goperl_frame_t *f, uint64_t sv, MAGIC *head) {
    if (goperl_api_magic_chain(f, sv) != head) goperl_api_magic_set_head(f, sv, head);
}
/* settle: the outermost native call ends, and a chain native code took off
 * whole is not put back any more */
static void goperl_svmagic_writeback_at(goperl_frame_t *f, int settle) {
    goperl_shared_t *const sh = (goperl_shared_t *)f->inst;
    if (!sh || !sh->svmagic_sv) return;
    MAGIC *const shown = sh->svmagic_orig;
    MAGIC *const val = sh->svmagic_val;
    if (!goperl_magic_is_interp_copy(f, shown)) {
        if (val == shown) return;
        MAGIC *const head = goperl_magic_chain_of_native(f, val);
        goperl_svmagic_head(f, sh->svmagic_sv, head);
        sh->svmagic_orig = head;
        if (head != val) sh->svmagic_val = head;
        return;
    }
    if (!settle && goperl_svmagic_detached(f)) {
        /* the native part of the chain is what native code made it; the
         * interpreter's magic stays where it is until the chain is put
         * back or the call ends */
        goperl_svmagic_head(f, sh->svmagic_sv, goperl_magic_chain_of_native(f, val));
        return;
    }
    /* every type of the interpreter's magic shown is still on the chain (a
     * copy of it: the one shown, or one an earlier SvMAGIC of the value
     * showed, which a MAGIC native code put before the head links to) */
    MAGIC *end = shown;
    for (; goperl_magic_is_interp_copy(f, end); end++) {
        int on = 0;
        for (MAGIC *mg = val; mg && !on; mg = mg->mg_moremagic)
            on = goperl_magic_is_interp_copy(f, mg) && mg->mg_type == end->mg_type;
        if (on) continue;
        sh->svmagic_sv = 0; /* the slot is done with */
        if (!f->failed) {
            snprintf(f->err, sizeof f->err,
                     "SvMAGIC: native code unlinked magic '%c' that the interpreter keeps on "
                     "the value from its chain; gperl cannot take the interpreter's own "
                     "magic off a value that way (sv_unmagic can)",
                     end->mg_type);
            f->failed = 1;
        }
        return;
    }
    /* end is the zeroed MAGIC after the copies */
    MAGIC *const head = goperl_magic_chain_of_native(f, val);
    goperl_svmagic_head(f, sh->svmagic_sv, head);
    /* whether the copies are still the front of the chain, as shown */
    MAGIC *c = shown;
    for (MAGIC *mg = val; mg == c && c != end; mg = mg->mg_moremagic) c++;
    if (c == end) {
        end->mg_moremagic = head;
    } else {
        /* native code linked a MAGIC of its own before or among the
         * copies, which are off its chain now */
        sh->svmagic_orig = head;
        sh->svmagic_val = head;
    }
}
static void goperl_svmagic_writeback(goperl_frame_t *f) { goperl_svmagic_writeback_at(f, 0); }
static void goperl_flush(goperl_frame_t *f) {
    goperl_svmagic_writeback(f);
    goperl_svcur_writeback(f);
    goperl_curcop_writeback(f);
    goperl_parser_push(f);
    goperl_cx_writeback(f);
    goperl_statcache_writeback(f);
    if (goperl_hints_live_v && goperl_hints_slot_v != goperl_hints_base_v) {
        goperl_hints_base_v = goperl_hints_slot_v;
        goperl_xs_op(f, GOPERL_OP_PLVAR_SET, (uint64_t)GOPERL_PL_HINTS,
                            (uint64_t)goperl_hints_slot_v, 0, 0);
    }
    if (f->hook_dirty) {
        int32_t d = f->hook_dirty;
        f->hook_dirty = 0;
        if (d & 1)
            goperl_xs_op(f, GOPERL_OP_PLVAR_SET,
                                (uint64_t)GOPERL_PL_DIEHOOK, f->hook_val[0], 0,
                                0);
        if (d & 2)
            goperl_xs_op(f, GOPERL_OP_PLVAR_SET,
                                (uint64_t)GOPERL_PL_WARNHOOK, f->hook_val[1],
                                0, 0);
    }
    goperl_plshadow_flush(f);
    goperl_lvslot_flush(f);
    goperl_pads_flush(f);
    goperl_ppaddr_sync(f);
    goperl_hooks_sync(f);
    goperl_bhk_sync(f);
}

/* Take a writable reference to an interpreter variable. The slot of a
 * variable never moves; it is refreshed from the interpreter unless a
 * store through it is still to be written back, and from then on every
 * flush writes back a value that differs from the one loaded - also when
 * the store comes after guest operations of the same expression (the
 * right-hand side of the assignment evaluated after the left). */
static void goperl_croakf(goperl_frame_t *f, const char *fmt, ...)
    __attribute__((noreturn, unused));
static IV *goperl_plint_ref(goperl_frame_t *f, int id) {
    /* every variable has a slot of its own: an id beyond the table would
     * be kept in what follows it */
    if (id < 0 || id >= GOPERL_PLSHADOW_MAX)
        goperl_croakf(f, "interpreter variable %d has no slot", id);
    goperl_plshadow_t *s = &goperl_plshadow_v[id];
    if (!s->live || s->val == s->shadow) {
        s->val = goperl_xs_op(f, GOPERL_OP_PLVAR_GET, (uint64_t)id, 0,
                                     0, 0);
        s->shadow = s->val;
        goperl_plshadow_set_live(s);
    }
    return (IV *)&s->val;
}
static SV **goperl_plsv_ref(goperl_frame_t *f, int id) {
    return (SV **)goperl_plint_ref(f, id);
}

/* An evaluation is over. A slot without users stays under the comparison
 * of the flush for a while (see GOPERL_LV_IV), in place of the oldest such
 * slot, which is given up.
 *
 * The body of an array is given up last: code keeps the address AvARRAY
 * gave it for as long as it works with the array (a template engine keeps
 * the base of its current frame), and reads the scalar fields of many
 * values meanwhile (AvFILLp, SvIVX, SvRV). The oldest slot that is no body
 * goes first; a body goes only when the slots kept are all bodies. */
static void goperl_lvuse_end(goperl_frame_t *f, goperl_lvstate_t *st,
                             goperl_lvslot_t *m) {
    if (--m->users > 0) return;
    if (st->nrecent == GOPERL_LV_RECENT) {
        int32_t at = 0;
        for (int32_t i = 0; i < st->nrecent; i++)
            if (st->recent[i]->field != GOPERL_LV_AVBODY) {
                at = i;
                break;
            }
        goperl_lvslot_t *old = st->recent[at];
        goperl_lvslot_store(f, old);
        goperl_lvslot_release(old);
        old->free_next = st->free;
        st->free = old;
        memmove(&st->recent[at], &st->recent[at + 1],
                (size_t)(st->nrecent - at - 1) * sizeof(st->recent[0]));
        st->nrecent--;
    }
    st->recent[st->nrecent++] = m;
}

/* The number of slots the array has room for (AvMAX + 1); defined with the
 * calls it is made of. */
static int32_t goperl_av_slots(goperl_frame_t *f, uint64_t avtok)
    __attribute__((unused));

/* AvARRAY is the array's body: every slot the array has room for, the ones
 * beyond its last element NULL. Code that extends an array and then looks
 * which slots are filled (AvARRAY(av)[i] ? ... : ...) reads them. The body
 * is loaded from the interpreter at every evaluation (what was stored into
 * it has been written back by then): the interpreter says where the body
 * is in its memory (SV_CALL 59, AV_BODY), and the slots are read there —
 * as they are, without the magic of a tied array, which is what AvARRAY
 * is. */
static void goperl_avbody_load(goperl_frame_t *f, goperl_lvslot_t *m) {
    uint64_t n = 0;
    const uint64_t where = goperl_xs_opx(f, GOPERL_OP_SV_CALL, m->svtok, 0, (uint64_t)(59), 0, 0, &n);
    if (!where) {
        /* an array with no body, as a tied array has: AvARRAY is NULL,
         * which is how code tells it to fetch the elements instead
         * (Tree::RB::XS) */
        if (m->val) {
            goperl_avbody_t *const old = GOPERL_AVBODY_OF(m);
            free(old->shadow);
            free(old);
        }
        m->val = m->shadow = 0;
        return;
    }
    int32_t len = (int32_t)n;
    int32_t slots = goperl_av_slots(f, m->svtok);
    if (slots < len) slots = len;
    /* the slots are SV pointers of the interpreter's: words of its width.
     * Every slot the array has room for is read as it is in memory, past
     * its last element too: code that takes elements out by lowering
     * AvFILLp reads the slot it has just left behind
     * (`AvARRAY(av)[AvFILLp(av)--]`, Future::XS), and perl leaves what
     * was there in place. */
    const uint64_t *const guest =
        slots > 0 ? (const uint64_t *)goperl_guest_mem(f, where) : 0;
    len = guest ? slots : 0;
    goperl_avbody_t *b = m->val ? GOPERL_AVBODY_OF(m) : 0;
    if (!b || slots > b->cap) {
        int32_t ncap = b ? b->cap : 16;
        while (ncap < slots) ncap *= 2;
        uint64_t *oshadow = b ? b->shadow : 0;
        goperl_avbody_t *nb = (goperl_avbody_t *)realloc(
            b, sizeof(goperl_avbody_t) + (size_t)ncap * 8);
        if (!nb) goperl_croakf(f, "out of memory for the body of an array");
        b = nb;
        m->val = m->shadow = (uint64_t)(uintptr_t)b->slots;
        b->len = 0;
        b->shadow = oshadow;
        uint64_t *nshadow = (uint64_t *)realloc(oshadow, (size_t)ncap * 8);
        if (!nshadow) goperl_croakf(f, "out of memory for the body of an array");
        b->shadow = nshadow;
        b->cap = ncap;
    }
    for (int32_t i = 0; i < len; i++) b->slots[i] = b->shadow[i] = guest[i];
    for (int32_t i = len; i < slots; i++) b->slots[i] = b->shadow[i] = 0;
    /* every slot is written back when it was stored into, the ones beyond
     * the last element too */
    b->len = slots;
}

/* The lvalue of a field (see GOPERL_LV_IV): pending stores are written
 * back, the evaluations that are known to be over are ended, and the slot
 * of the field is found or made, and loaded from the interpreter. */
static uint64_t *goperl_lv_ref(goperl_frame_t *f, SV *sv, int field,
                               const void *site, uintptr_t frame) {
    uint64_t tok = GOPERL_TOK(sv);
    goperl_lvslot_t *m = 0;
    goperl_flush(f);
    goperl_lvstate_t *st = goperl_lvstate_v;
    if (!st) {
        st = (goperl_lvstate_t *)calloc(1, sizeof(goperl_lvstate_t));
        if (!st) goperl_croakf(f, "out of memory for an lvalue slot");
        goperl_lvstate_v = st;
    }
    /* made by functions that have returned: the stack grows downwards, and
     * this function's frame is above theirs */
    while (st->nuses > 0 && st->uses[st->nuses - 1].frame < frame)
        goperl_lvuse_end(f, st, st->uses[--st->nuses].slot);
    /* an earlier evaluation of this expansion in this activation */
    for (int32_t i = st->nuses - 1; i >= 0; i--) {
        goperl_lvuse_t *u = &st->uses[i];
        if (u->site != site || u->frame != frame) continue;
        goperl_lvslot_t *ended = u->slot;
        memmove(u, u + 1, (size_t)(st->nuses - i - 1) * sizeof(*u));
        st->nuses--;
        goperl_lvuse_end(f, st, ended);
        break;
    }
    /* the slot the field has */
    for (int32_t i = st->nuses - 1; i >= 0 && !m; i--)
        if (st->uses[i].slot->svtok == tok && st->uses[i].slot->field == field)
            m = st->uses[i].slot;
    for (int32_t i = st->nrecent - 1; i >= 0 && !m; i--)
        if (st->recent[i]->svtok == tok && st->recent[i]->field == field) {
            m = st->recent[i];
            memmove(&st->recent[i], &st->recent[i + 1],
                    (size_t)(st->nrecent - i - 1) * sizeof(st->recent[0]));
            st->nrecent--;
        }
    if (st->nuses == st->capuses) {
        int32_t ncap = st->capuses ? st->capuses * 2 : 64;
        goperl_lvuse_t *nu = (goperl_lvuse_t *)realloc(
            st->uses, (size_t)ncap * sizeof(st->uses[0]));
        if (!nu) goperl_croakf(f, "out of memory for an lvalue slot");
        st->uses = nu;
        st->capuses = ncap;
    }
    if (!m) {
        if (!st->free) {
            goperl_lvchunk_t *c =
                (goperl_lvchunk_t *)calloc(1, sizeof(goperl_lvchunk_t));
            if (!c) goperl_croakf(f, "out of memory for an lvalue slot");
            c->next = st->chunks;
            st->chunks = c;
            for (int32_t i = 0; i < GOPERL_LVCHUNK; i++) {
                c->slots[i].free_next = st->free;
                st->free = &c->slots[i];
            }
        }
        m = st->free;
        st->free = m->free_next;
        m->svtok = tok;
        m->field = field;
        m->users = 0;
        m->val = m->shadow = 0;
    }
    m->users++;
    st->uses[st->nuses].site = site;
    st->uses[st->nuses].frame = frame;
    st->uses[st->nuses].slot = m;
    st->nuses++;
    if (field == GOPERL_LV_AVBODY) {
        goperl_avbody_load(f, m);
        return &m->val;
    }
    if (GOPERL_LV_IS_BODY(field)) {
        goperl_lvbody_load(f, m);
        return &m->val;
    }
    /* the load is a guest operation, which flushes: the slot is clean (the
     * flush above wrote it back) and stays so until the value is in */
    uint64_t v = goperl_lv_load_field(f, tok, field);
    m->val = m->shadow = v;
    return &m->val;
}

/* ---- AvARRAY ----------------------------------------------------------------- */

#define AvARRAY(av)                                                       \
    ((SV **)(uintptr_t)*goperl_lv_ref(_gof, (SV *)(av), GOPERL_LV_AVBODY, \
                                      GOPERL_LV_AT))

/* sortsv and sortsv_flags over a mirror (or any host SV* array): a
 * stable merge sort, as perl's (pp_sort.c), descending for SORTf_DESC;
 * the mirror flush propagates the permutation. */
typedef I32 (*SVCOMPARE_t)(pTHX_ SV *a, SV *b);
#ifndef SORTf_DESC
#define SORTf_DESC 1
#define SORTf_STABLE 2
#define SORTf_UNSTABLE 8
#endif
static void goperl_sortsv_flags(goperl_frame_t *f, SV **array, size_t n,
                                SVCOMPARE_t cmp, U32 flags) {
    const int desc = (flags & SORTf_DESC) != 0;
    if (n < 2) return;
    SV **tmp = (SV **)malloc(n * sizeof(SV *));
    if (!tmp) goperl_croakf(f, "%s", "Out of memory!");
    /* bottom-up: runs of width w merged into the other buffer, whose role
     * then swaps */
    SV **from = array, **to = tmp;
    for (size_t w = 1; w < n; w *= 2) {
        for (size_t lo = 0; lo < n; lo += 2 * w) {
            const size_t mid = lo + w < n ? lo + w : n;
            const size_t hi = lo + 2 * w < n ? lo + 2 * w : n;
            size_t i = lo, j = mid, k = lo;
            while (i < mid && j < hi) {
                I32 c = cmp(from[i], from[j]);
                if (desc) c = -c;
                to[k++] = c <= 0 ? from[i++] : from[j++];
            }
            while (i < mid) to[k++] = from[i++];
            while (j < hi) to[k++] = from[j++];
        }
        SV **x = from;
        from = to;
        to = x;
    }
    if (from != array) memcpy(array, from, n * sizeof(SV *));
    free(tmp);
}
static void goperl_sortsv(goperl_frame_t *f, SV **array, size_t n,
                          SVCOMPARE_t cmp) {
    goperl_sortsv_flags(f, array, n, cmp, 0);
}
#define sortsv_flags(a, n, c, flags) \
    goperl_sortsv_flags(_gof, (a), (size_t)(n), (c), (U32)(flags))
#define sortsv(a, n, c) goperl_sortsv(_gof, (a), (size_t)(n), (c))
static I32 goperl_sv_cmp_fn(pTHX_ SV *a, SV *b) {
    return (I32)(int64_t)goperl_op0(GOPERL_OP_SV_CMP, GOPERL_TOK(a),
                                    GOPERL_TOK(b));
}
#define Perl_sv_cmp goperl_sv_cmp_fn

/* ---- the perl stack ------------------------------------------------------ */

/* The frame with its stack in line with the interpreter's: the stack of a
 * pp function is a copy that falls behind when interpreter code runs on
 * the real one, and is fetched again when it is next looked at. */
static void goperl_pp_stack_pull(goperl_frame_t *f);
static inline goperl_frame_t *goperl_stk(goperl_frame_t *f) {
    if (__builtin_expect(f->st_stale, 0)) goperl_pp_stack_pull(f);
    return f;
}
#define PL_stack_base (goperl_stk(_gof)->stbase)
#define PL_stack_sp (goperl_stk(_gof)->psp)

/* PL_markstack_ptr is an assignable pointer some dists arithmetic on
 * directly (DBI bumps it to un-pop the mark its caller's dXSARGS took).
 * The pointer lives in a synced global; every mark operation adopts any
 * adjustment back into the frame's mark index first. */
#define goperl_markstack_ptr_v (GOPERL_SHARED->markstack_ptr)
#define goperl_markstack_owner_v (GOPERL_SHARED->markstack_owner)
static void goperl_marks_adopt(goperl_frame_t *f) {
    goperl_stk(f);
    if (goperl_markstack_owner_v == f && goperl_markstack_ptr_v)
        f->markidx = (int32_t)(goperl_markstack_ptr_v - f->marks);
}
static void goperl_marks_publish(goperl_frame_t *f) {
    goperl_markstack_owner_v = f;
    goperl_markstack_ptr_v = f->marks + f->markidx;
}
/* Makes a frame the current one. The mark pointer is the current frame's
 * from then on: what the frame that ran before left there is not this
 * frame's, and the owner's address does not tell the two apart, because
 * frames are short-lived and a new one may lie where an earlier one was. A
 * pointer the frame that was current had moved is taken into its marks
 * first. An activation that opens on the frame that is current already (an
 * XSUB called by C pointer from another) keeps the pointer: its caller may
 * have moved it for it. */
static void goperl_frame_enter(goperl_frame_t *f) {
    goperl_frame_t *prev = goperl_cur_frame_v;
    if (prev == f) return;
    if (prev && goperl_markstack_owner_v == prev && goperl_markstack_ptr_v &&
        !prev->st_stale)
        prev->markidx = (int32_t)(goperl_markstack_ptr_v - prev->marks);
    goperl_cur_frame_v = f;
    goperl_marks_publish(f);
}
/* The pointer as an lvalue: what the code moved it to before (Inline's
 * void XSUBs push a mark with `temp = PL_markstack_ptr++` and compare the
 * pointer with temp after the call) is taken into the frame's marks first,
 * so that a read sees it. */
static int32_t **goperl_markstack_ref(goperl_frame_t *f) {
    goperl_marks_adopt(f);
    goperl_marks_publish(f);
    return &goperl_markstack_ptr_v;
}
#define PL_markstack_ptr (*goperl_markstack_ref(_gof))
static void goperl_croakf(goperl_frame_t *f, const char *fmt, ...)
    __attribute__((noreturn, unused));
static I32 goperl_popmark(goperl_frame_t *f) {
    goperl_marks_adopt(f);
    /* marks[0] of a pp function's frame stands for a mark of the
     * interpreter's that was not handed over */
    if (f->pp_frame && f->markidx <= 0 && f->gmark_lo > 0)
        goperl_croakf(f, "a pp function popped more marks than the go-perl "
                         "XS SDK hands it (%d)", GOPERL_XS_MARKS / 2);
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
/* Makes room for n more values above sp, moving the stack to a larger
 * block when it is full, and returns sp in the stack as it now is (the
 * frame's own stack pointer is moved along). As with perl's EXTEND, a
 * pointer into the stack taken before (MARK) is stale after it; ST(n) and
 * ORIGMARK are not, they go through the base. */
static SV **goperl_stack_extend(goperl_frame_t *f, SV **sp, SSize_t n) {
    int64_t used = (int64_t)(sp - f->stbase);
    if (n < 0) n = 0;
    if (used + n + 1 < f->stcap) return sp;
    int64_t ncap = f->stcap * 2;
    if (ncap < used + n + 64) ncap = used + n + 64;
    SV **nb = (SV **)malloc((size_t)ncap * sizeof(SV *));
    if (!nb) goperl_croakf(f, "Out of memory during stack extend");
    memcpy(nb, f->stbase, (size_t)f->stcap * sizeof(SV *));
    int64_t psp = (int64_t)(f->psp - f->stbase);
    if (f->stbase != (SV **)f->st) free(f->stbase);
    f->stbase = nb;
    f->stcap = ncap;
    f->psp = nb + psp;
    return nb + used;
}

#define dSP SV **sp = PL_stack_sp
#define SP sp
#define MARK mark
#define POPMARK goperl_popmark(_gof)
#define TOPMARK (goperl_stk(_gof)->marks[_gof->markidx])
#define PUSHMARK(p) goperl_pushmark(_gof, (I32)((p)-PL_stack_base))
#define dMARK SV **mark = PL_stack_base + POPMARK
#define dORIGMARK const I32 origmark = (I32)(mark - PL_stack_base)
#define ORIGMARK (PL_stack_base + origmark)
#define dAX const I32 ax = (I32)(mark - PL_stack_base + 1)
#define dAXMARK             \
    I32 ax = POPMARK + 1;   \
    SV **mark = PL_stack_base + ax - 1
#define dITEMS I32 items = (I32)(sp - mark)
#define EXTEND(p, n) \
    ({ __typeof__(p) *const goperl_EXTEND_p_ = &(p); \
       *goperl_EXTEND_p_ = goperl_stack_extend(_gof, *goperl_EXTEND_p_, (SSize_t)(n)); \
       (void)0; })
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
/* a declaration, as perl's: it may come before other declarations */
#define dNOOP struct Perl___notused_struct
#define dVAR dNOOP
/* Where native code that runs long lets the interpreter deliver what is
 * pending: the handler of a signal, an interruption of the program. */
static void goperl_croak_errsv(goperl_frame_t *f) __attribute__((noreturn));
static inline __attribute__((unused)) void goperl_async(goperl_frame_t *f) {
    /* GOPERL_CALL_ASYNC: 1 when a handler died, which is this code's die */
    if (goperl_do_op3(f, GOPERL_OP_CALL, 0, 0, (uint64_t)(16), 0, 0))
        goperl_croak_errsv(f);
}
#define PERL_ASYNC_CHECK() goperl_async(_gof)
#define strNE(a, b) (strcmp((a), (b)) != 0)
/* A character or a byte that is the same one byte in UTF-8: the value is
 * not cut to a byte (a code point above 0xFF is not invariant). */
#define UTF8_IS_INVARIANT(c) UVCHR_IS_INVARIANT(c)
/* A code point that is the same one byte in UTF-8 (utf8.h; ASCII). */
#define OFFUNI_IS_INVARIANT(cp) ((UV)(cp) < 0x80)
#define UVCHR_IS_INVARIANT(cp) OFFUNI_IS_INVARIANT(cp)
#define UNI_IS_INVARIANT(cp) UVCHR_IS_INVARIANT(cp)
#define NATIVE_BYTE_IS_INVARIANT(c) UVCHR_IS_INVARIANT(c)
/* The bytes the case change of one character can take (utf8.h). */
#define UTF8_MAXBYTES_CASE 13
#define UTF8_EIGHT_BIT_HI(c) ((U8)(0xC0 | ((U8)(c) >> 6)))
#define UTF8_EIGHT_BIT_LO(c) ((U8)(0x80 | ((U8)(c)&0x3F)))

/* ---- PerlIO handles ---------------------------------------------------------
 * A PerlIO* is a handle of the interpreter: an opaque token, as an SV* is.
 * Every operation on one runs in the interpreter, on its layers and its
 * buffers, so what native code reads or writes is in sequence with what
 * the Perl program reads and writes on the same handle, whatever the handle
 * is (a file, a socket, a scalar opened as a file, a tied or layered one).
 *
 * PerlIO is a MACRO so <perliol.h> (the layer-author view) can replace it
 * with the layered handle type. */
struct goperl_perlio;
#define PerlIO struct goperl_perlio
enum {
    GOPERL_PIO_STDIN = 0,
    GOPERL_PIO_STDOUT = 1,
    GOPERL_PIO_STDERR = 2,
    GOPERL_PIO_FILENO = 3,
    GOPERL_PIO_READ = 4,
    GOPERL_PIO_EOF = 5,
    GOPERL_PIO_ERROR = 6,
    GOPERL_PIO_CLEARERR = 7,
    GOPERL_PIO_SEEK_SET = 8,
    GOPERL_PIO_SEEK_CUR = 9,
    GOPERL_PIO_SEEK_END = 10,
    GOPERL_PIO_TELL = 11,
    GOPERL_PIO_GETC = 12,
    GOPERL_PIO_UNGETC = 13,
    GOPERL_PIO_UNREAD = 14,
    GOPERL_PIO_FDOPEN = 15,
    GOPERL_PIO_BINMODE = 16,
    GOPERL_PIO_SETLINEBUF = 17,
    GOPERL_PIO_IFP = 18,
    GOPERL_PIO_OFP = 19,
    GOPERL_PIO_VALID = 20,
    GOPERL_PIO_FAST_GETS = 21,
    GOPERL_PIO_HAS_CNTPTR = 22,
    GOPERL_PIO_CANSET_CNT = 23,
    GOPERL_PIO_HAS_BASE = 24,
    GOPERL_PIO_GET_CNT = 25,
    GOPERL_PIO_GET_PTR = 26,
    GOPERL_PIO_GET_BASE = 27,
    GOPERL_PIO_GET_BUFSIZ = 28,
    GOPERL_PIO_SET_PTRCNT = 29,
    GOPERL_PIO_FILL = 30,
    GOPERL_PIO_APPLY_LAYERS = 31,
    GOPERL_PIO_LAYERS = 32,
    GOPERL_PIO_FIND_FILE = 33,
    GOPERL_PIO_EXPORT_FILE = 34,
    GOPERL_PIO_RELEASE_FILE = 35
};
#define GOPERL_PIO_READ_MAX 65536
#define GOPERL_PL_ERRNO 60
/* A handle operation of the interpreter failed. Native code looks for the
 * reason where it looks after any failed call: in errno. */
static void goperl_errno_from_interp(goperl_frame_t *f) __attribute__((unused));
static void goperl_errno_from_interp(goperl_frame_t *f) {
    const int e = (int)(int64_t)goperl_xs_op(f, GOPERL_OP_PLVAR_GET,
                                             GOPERL_PL_ERRNO, 0, 0, 0);
    if (e) errno = goperl_errno_host(e);
}
static uint64_t goperl_pio(goperl_frame_t *f, int sel, const void *h,
                           uint64_t b, const char *s, uint64_t slen)
    __attribute__((unused));
/* A handle is the interpreter's (an address of its memory), or a view of
 * one that code which walks layers is given (perliol.h): the views made
 * for the native call say which handle of the interpreter each of their
 * slots stands for (a handle, or the next field of a layer, which is the
 * handle of the layer below). Which of the two a value is cannot be told
 * by where it points: the C library's heap may lie below the size of the
 * interpreter's memory. */
typedef struct goperl_pio_slot {
    struct goperl_pio_slot *next;
    const void *slot;
    uint64_t tok;
} goperl_pio_slot_t;
#define goperl_pio_views_v (*(goperl_pio_slot_t **)&GOPERL_SHARED->pio_views)
static inline __attribute__((unused)) const goperl_pio_slot_t *goperl_pio_view_of(
    const void *h) {
    if (!h || !goperl_cur_frame_v) return 0;
    for (const goperl_pio_slot_t *m = goperl_pio_views_v; m; m = m->next)
        if (m->slot == h) return m;
    return 0;
}
static inline __attribute__((unused)) uint64_t goperl_pio_tok(goperl_frame_t *f,
                                                              const void *h) {
    (void)f;
    const goperl_pio_slot_t *m = goperl_pio_view_of(h);
    return m ? m->tok : (uint64_t)(uintptr_t)h;
}
static uint64_t goperl_pio(goperl_frame_t *f, int sel, const void *h,
                           uint64_t b, const char *s, uint64_t slen) {
    return goperl_do_op3(f, GOPERL_OP_PERLIO, goperl_pio_tok(f, h), b, (uint64_t)(sel), s, slen);
}
#define GOPERL_PIO_H(sel) \
    ((PerlIO *)(uintptr_t)goperl_pio(_gof, (sel), 0, 0, 0, 0))
#define PerlIO_stdin() GOPERL_PIO_H(GOPERL_PIO_STDIN)
#define PerlIO_stdout() GOPERL_PIO_H(GOPERL_PIO_STDOUT)
#define PerlIO_stderr() GOPERL_PIO_H(GOPERL_PIO_STDERR)

/* Reads through the interpreter in pieces of its transfer buffer; a short
 * piece is the end of the input or an error, as a short PerlIO_read is. */
static SSize_t goperl_pio_read(goperl_frame_t *f, const void *h, void *buf,
                               size_t count) __attribute__((unused));
static SSize_t goperl_pio_read(goperl_frame_t *f, const void *h, void *buf,
                               size_t count) {
    size_t total = 0;
    while (total < count) {
        size_t want = count - total;
        if (want > GOPERL_PIO_READ_MAX) want = GOPERL_PIO_READ_MAX;
        uint64_t got = 0;
        const uint64_t r = goperl_do_opx(f, GOPERL_OP_PERLIO, goperl_pio_tok(f, h), want,
                                         GOPERL_PIO_READ, 0, 0, &got);
        const int64_t n = (int64_t)got;
        if (n < 0) {
            if (total) return (SSize_t)total;
            goperl_errno_from_interp(f);
            return -1;
        }
        if (n > 0)
            memcpy((char *)buf + total, goperl_guest_mem(f, r), (size_t)n);
        total += (size_t)n;
        if ((size_t)n < want) break;
    }
    return (SSize_t)total;
}
static SSize_t goperl_pio_write(goperl_frame_t *f, const void *h,
                                const void *buf, size_t count)
    __attribute__((unused));
static SSize_t goperl_pio_write(goperl_frame_t *f, const void *h,
                                const void *buf, size_t count) {
    const SSize_t n = (SSize_t)(int64_t)goperl_do_op(
        f, GOPERL_OP_PERLIO_WRITE, (uint64_t)(uintptr_t)h,
        (uint64_t)count, (const char *)buf, (uint64_t)count);
    if (n < 0) goperl_errno_from_interp(f);
    return n;
}
static int goperl_pio_vprintf(goperl_frame_t *f, const void *h,
                              const char *fmt, va_list ap)
    __attribute__((unused));
static int goperl_pio_vprintf(goperl_frame_t *f, const void *h,
                              const char *fmt, va_list ap) {
    goperl_text_t t;
    goperl_text_init(&t);
    goperl_vtext(f, &t, fmt, ap);
    int n = (int)goperl_pio_write(f, h, t.p, t.len);
    goperl_text_free(&t);
    return n;
}
static int goperl_perlio_printf(const void *h, const char *fmt, ...)
    __attribute__((unused));
static int goperl_perlio_printf(const void *h, const char *fmt, ...) {
    va_list ap;
    va_start(ap, fmt);
    int n = goperl_pio_vprintf(goperl_cur_frame_v, h, fmt, ap);
    va_end(ap);
    return n;
}
static int goperl_perlio_stdoutf(const char *fmt, ...) __attribute__((unused));
static int goperl_perlio_stdoutf(const char *fmt, ...) {
    goperl_frame_t *f = goperl_cur_frame_v;
    va_list ap;
    va_start(ap, fmt);
    int n = goperl_pio_vprintf(
        f, (void *)(uintptr_t)goperl_pio(f, GOPERL_PIO_STDOUT, 0, 0, 0, 0), fmt,
        ap);
    va_end(ap);
    return n;
}
/* The descriptor of the host process behind a handle, which is what native
 * code hands to the host's system calls (ioctl, fstat, mmap, a library's
 * set_fd): the interpreter's own descriptor numbers mean nothing outside
 * it. -1 for a handle with no descriptor, and for one whose descriptor is
 * not a descriptor of the host (a file of an in-memory filesystem). */
static int goperl_pio_fileno(goperl_frame_t *f, const void *h)
    __attribute__((unused));
static int goperl_pio_fileno(goperl_frame_t *f, const void *h) {
    int64_t fd = (int64_t)goperl_pio(f, GOPERL_PIO_FILENO, h, 0, 0, 0);
    if (fd < 0) return -1;
    return (int)goperl_api_host_fd(f, fd);
}
static IV goperl_pio_seek(goperl_frame_t *f, const void *h, Off_t off,
                          int whence) __attribute__((unused));
static IV goperl_pio_seek(goperl_frame_t *f, const void *h, Off_t off,
                          int whence) {
    int sel = whence == SEEK_SET   ? GOPERL_PIO_SEEK_SET
              : whence == SEEK_CUR ? GOPERL_PIO_SEEK_CUR
              : whence == SEEK_END ? GOPERL_PIO_SEEK_END
                                   : -1;
    if (sel < 0) {
        errno = EINVAL;
        return -1;
    }
    const IV r = (IV)(int64_t)goperl_pio(f, sel, h, (uint64_t)off, 0, 0);
    if (r != 0) goperl_errno_from_interp(f);
    return r;
}
/* The handle functions that answer 0 for success and something else for
 * a failure (tell: a negative position). */
static int64_t goperl_pio_status(goperl_frame_t *f, int sel, const void *h)
    __attribute__((unused));
static int64_t goperl_pio_status(goperl_frame_t *f, int sel, const void *h) {
    const int64_t r = (int64_t)goperl_pio(f, sel, h, 0, 0, 0);
    if (sel == GOPERL_PIO_TELL ? r < 0 : r != 0) goperl_errno_from_interp(f);
    return r;
}
static int goperl_pio_flush(goperl_frame_t *f, const void *h)
    __attribute__((unused));
static int goperl_pio_flush(goperl_frame_t *f, const void *h) {
    const int r = (int)(int64_t)goperl_do_op(
        f, GOPERL_OP_PERLIO_FLUSH, (uint64_t)(uintptr_t)h, 0, 0, 0);
    if (r != 0) goperl_errno_from_interp(f);
    return r;
}
/* PerlIO_binmode and PerlIO_apply_layers are functions that take the
 * interpreter (aTHX_) on perl, and are called that way. */
static int PerlIO_binmode(pTHX_ const void *h, int iotype, int omode,
                          const char *names) __attribute__((unused));
static int PerlIO_binmode(pTHX_ const void *h, int iotype, int omode,
                          const char *names) {
    const uint64_t ok = goperl_pio(
        _gof, GOPERL_PIO_BINMODE, h,
        ((uint64_t)(uint32_t)iotype << 32) | (uint32_t)omode,
        names ? names : "", names ? __builtin_strlen(names) : 0);
    if (!ok) goperl_errno_from_interp(_gof);
    return ok ? 1 : 0;
}
static int PerlIO_apply_layers(pTHX_ const void *h, const char *mode,
                               const char *names) __attribute__((unused));
static int PerlIO_apply_layers(pTHX_ const void *h, const char *mode,
                               const char *names) {
    const size_t ml = mode ? __builtin_strlen(mode) : 0, nl = names ? __builtin_strlen(names) : 0;
    char *req = (char *)malloc(ml + nl + 2);
    if (!req) goperl_croakf(_gof, "%s", "Out of memory!");
    if (ml) memcpy(req, mode, ml);
    req[ml] = '\0';
    if (nl) memcpy(req + ml + 1, names, nl);
    req[ml + 1 + nl] = '\0';
    goperl_flush(_gof);
    const int32_t failed_before = _gof->failed;
    const int r = (int)(int64_t)goperl_xs_opx(_gof, GOPERL_OP_PERLIO, (uint64_t)(uintptr_t)h, 0,
                                              GOPERL_PIO_APPLY_LAYERS, req, ml + 1 + nl + 1, 0);
    goperl_gen_bump();
    free(req);
    if (!failed_before && _gof->failed && _gof->jb) goperl_raise(_gof);
    if (r != 0) goperl_errno_from_interp(_gof);
    return r;
}
/* PerlIO_openn: the open behind every open of the interpreter's, with the
 * layers, the mode, a descriptor (-1 for none), the open(2) mode and
 * permissions, the handle to reuse (or NULL) and the arguments (a path, a
 * reference to a scalar for :scalar, ...). The handle is the
 * interpreter's (Hyperman opens its request bodies with ":scalar"). */
#define GOPERL_PIO_OPENN 36
#define GOPERL_PIO_ISUTF8 37
static PerlIO *PerlIO_openn(pTHX_ const char *layers, const char *mode, int fd,
                            int imode, int perm, PerlIO *old, int narg, SV **args)
    __attribute__((unused));
static PerlIO *PerlIO_openn(pTHX_ const char *layers, const char *mode, int fd,
                            int imode, int perm, PerlIO *old, int narg, SV **args) {
    if (!mode) mode = "r";
    if (!layers) layers = "";
    if (narg < 0) narg = 0;
    const size_t ll = __builtin_strlen(layers), ml = __builtin_strlen(mode);
    const size_t len = 4 * 8 + (size_t)narg * 8 + ll + 1 + ml + 1;
    char *req = (char *)malloc(len);
    if (!req) goperl_croakf(_gof, "%s", "Out of memory!");
    const goperl_gword_t head[4] = {(goperl_gword_t)(int64_t)fd, (goperl_gword_t)(int64_t)imode,
                                    (goperl_gword_t)(int64_t)perm, (goperl_gword_t)narg};
    memcpy(req, head, sizeof head);
    for (int i = 0; i < narg; i++) {
        const goperl_gword_t tok = GOPERL_TOK(args[i]);
        memcpy(req + sizeof head + (size_t)i * 8, &tok, 8);
    }
    memcpy(req + sizeof head + (size_t)narg * 8, layers, ll + 1);
    memcpy(req + sizeof head + (size_t)narg * 8 + ll + 1, mode, ml + 1);
    PerlIO *const io =
        (PerlIO *)(uintptr_t)goperl_pio(_gof, GOPERL_PIO_OPENN, old, (uint64_t)len, req, len);
    free(req);
    if (!io) goperl_errno_from_interp(_gof);
    return io;
}
/* Whether the handle's top layer takes characters (:utf8). */
#define PerlIO_isutf8(io) ((int)goperl_pio(_gof, GOPERL_PIO_ISUTF8, (io), 0, 0, 0))
#define PerlIO_fileno(io) goperl_pio_fileno(_gof, (io))
#define PerlIO_read(io, buf, n) \
    goperl_pio_read(_gof, (io), (buf), (size_t)(n))
#define PerlIO_write(io, buf, n) \
    goperl_pio_write(_gof, (io), (buf), (size_t)(n))
static int goperl_pio_puts(goperl_frame_t *f, const void *h, const char *s2)
    __attribute__((unused));
static int goperl_pio_puts(goperl_frame_t *f, const void *h, const char *s2) {
    return (int)goperl_pio_write(f, h, s2, __builtin_strlen(s2));
}
#define PerlIO_puts(io, s2) goperl_pio_puts(_gof, (io), (s2))
#define PerlIO_putc(io, c) \
    ({ GOPERL_ONCE(goperl_PerlIO_putc_c_, c); \
       (goperl_pio_write(_gof, (io), &(char){(char)(goperl_PerlIO_putc_c_)}, 1) == 1 ? (int)(unsigned char)(goperl_PerlIO_putc_c_) : EOF); })
#define PerlIO_printf goperl_perlio_printf
#define PerlIO_stdoutf goperl_perlio_stdoutf
#define PerlIO_vprintf(io, fmt, ap) goperl_pio_vprintf(_gof, (io), (fmt), (ap))
#define PerlIO_eof(io) ((int)goperl_pio(_gof, GOPERL_PIO_EOF, (io), 0, 0, 0))
#define PerlIO_error(io) \
    ((int)(int64_t)goperl_pio(_gof, GOPERL_PIO_ERROR, (io), 0, 0, 0))
#define PerlIO_clearerr(io) \
    ((void)goperl_pio(_gof, GOPERL_PIO_CLEARERR, (io), 0, 0, 0))
#define PerlIO_seek(io, off, whence) \
    goperl_pio_seek(_gof, (io), (Off_t)(off), (whence))
#define PerlIO_tell(io) ((Off_t)goperl_pio_status(_gof, GOPERL_PIO_TELL, (io)))
#define PerlIO_rewind(io) \
    ({ GOPERL_ONCE(goperl_PerlIO_rewind_io_, io); \
       ((void)(goperl_pio_seek(_gof, (goperl_PerlIO_rewind_io_), 0, SEEK_SET), PerlIO_clearerr(goperl_PerlIO_rewind_io_))); })
#define PerlIO_getc(io) \
    ((int)(int64_t)goperl_pio(_gof, GOPERL_PIO_GETC, (io), 0, 0, 0))
#define PerlIO_ungetc(io, c) \
    ((int)(int64_t)goperl_pio(_gof, GOPERL_PIO_UNGETC, (io), (uint64_t)(int64_t)(c), 0, 0))
#define PerlIO_unread(io, buf, n) \
    ({ GOPERL_ONCE(goperl_PerlIO_unread_n_, n); \
       ((SSize_t)(int64_t)goperl_pio(_gof, GOPERL_PIO_UNREAD, (io), (uint64_t)(goperl_PerlIO_unread_n_), (const char *)(buf), (uint64_t)(goperl_PerlIO_unread_n_))); })
#define PerlIO_setlinebuf(io) \
    ((void)goperl_pio(_gof, GOPERL_PIO_SETLINEBUF, (io), 0, 0, 0))
/* The read buffer of a handle is the interpreter's own: the pointers are
 * into its memory, valid until the next operation on the handle. */
#ifndef STDCHAR
#define STDCHAR char
#endif
static STDCHAR *goperl_pio_bufptr(goperl_frame_t *f, int sel, const void *h)
    __attribute__((unused));
static STDCHAR *goperl_pio_bufptr(goperl_frame_t *f, int sel, const void *h) {
    uint64_t g = goperl_pio(f, sel, h, 0, 0, 0);
    if (!g || (int64_t)g < 0) return 0;
    return (STDCHAR *)goperl_guest_mem(f, g);
}
static void goperl_pio_set_ptrcnt(goperl_frame_t *f, const void *h,
                                  STDCHAR *ptr, SSize_t cnt)
    __attribute__((unused));
static void goperl_pio_set_ptrcnt(goperl_frame_t *f, const void *h,
                                  STDCHAR *ptr, SSize_t cnt) {
    /* ptr is a host address inside the buffer: back to the interpreter's */
    uint64_t g = goperl_pio(f, GOPERL_PIO_GET_PTR, h, 0, 0, 0);
    if (!g || (int64_t)g < 0) return;
    STDCHAR *cur = (STDCHAR *)goperl_guest_mem(f, g);
    uint64_t target = (uint64_t)((int64_t)g + (ptr - cur));
    goperl_do_op3(f, GOPERL_OP_PERLIO, goperl_pio_tok(f, h), target,
                  GOPERL_PIO_SET_PTRCNT | ((uint64_t)(uint32_t)(int32_t)cnt << 32), 0, 0);
}
#define USE_PERLIO
#define PerlIO_fast_gets(io) \
    ((int)goperl_pio(_gof, GOPERL_PIO_FAST_GETS, (io), 0, 0, 0))
#define PerlIO_has_cntptr(io) \
    ((int)goperl_pio(_gof, GOPERL_PIO_HAS_CNTPTR, (io), 0, 0, 0))
#define PerlIO_canset_cnt(io) \
    ((int)goperl_pio(_gof, GOPERL_PIO_CANSET_CNT, (io), 0, 0, 0))
#define PerlIO_has_base(io) \
    ((int)goperl_pio(_gof, GOPERL_PIO_HAS_BASE, (io), 0, 0, 0))
#define PerlIO_get_cnt(io) \
    ((SSize_t)(int64_t)goperl_pio(_gof, GOPERL_PIO_GET_CNT, (io), 0, 0, 0))
#define PerlIO_get_bufsiz(io) \
    ((SSize_t)(int64_t)goperl_pio(_gof, GOPERL_PIO_GET_BUFSIZ, (io), 0, 0, 0))
#define PerlIO_get_ptr(io) goperl_pio_bufptr(_gof, GOPERL_PIO_GET_PTR, (io))
#define PerlIO_get_base(io) goperl_pio_bufptr(_gof, GOPERL_PIO_GET_BASE, (io))
#define PerlIO_set_ptrcnt(io, ptr, cnt) \
    goperl_pio_set_ptrcnt(_gof, (io), (STDCHAR *)(ptr), (SSize_t)(cnt))
#define PerlIO_fill(io) ((int)goperl_pio_status(_gof, GOPERL_PIO_FILL, (io)))
#define PerlIO_flush(io) goperl_pio_flush(_gof, (io))

/* XCPT: local exception catching (the NO_XSLOCKS surface). A croak inside
 * the TRY block longjmps to the local buffer instead of the XSUB entry. */
#define dXCPT                        \
    jmp_buf goperl_xcpt_jb;          \
    void *goperl_xcpt_prev = 0;      \
    int32_t goperl_xcpt_failed = 0;  \
    int goperl_xcpt_rc = 0
#define XCPT_TRY_START                               \
    goperl_xcpt_prev = _gof->jb;                     \
    _gof->jb = (void *)&goperl_xcpt_jb;              \
    goperl_xcpt_rc = GOPERL_SETJMP(goperl_xcpt_jb);  \
    if (goperl_xcpt_rc == 0)
#define XCPT_TRY_END                                 \
    _gof->jb = goperl_xcpt_prev;                     \
    if (goperl_xcpt_rc) {                            \
        goperl_xcpt_failed = _gof->failed;           \
        _gof->failed = 0;                            \
    }
#define XCPT_CATCH if (goperl_xcpt_rc != 0)
/* XCPT_RETHROW goes on with what was caught, as croak(NULL) does on perl:
 * the same exception, whole — the value the interpreter holds for a croak
 * of native code, or the jump the interpreter itself was making. Only a
 * failure that has nothing but the message in the frame is thrown as that
 * message. */
static void goperl_xcpt_rethrow(goperl_frame_t *f, int32_t failed)
    __attribute__((noreturn, unused));
static void goperl_xcpt_rethrow(goperl_frame_t *f, int32_t failed) {
    if (failed == 2 /* GOPERL_FAILED_THROWN */ || failed == GOPERL_FAILED_JUMP) {
        if (!f->jb) goperl_raise(f);
        f->failed = failed;
        goperl_jump_out(f);
    }
    goperl_croakf(f, "%s", f->err);
}
#define XCPT_RETHROW goperl_xcpt_rethrow(_gof, goperl_xcpt_failed)

/* JMPENV: the place native code catches what ends the code it calls before
 * it returns (cop.h), the way the XCPT macros above do. JMPENV_PUSH gives
 * 0, and gives 3 again when a croak of native code, or a die of the
 * interpreter inside an operation, comes back to it; from there until
 * JMPENV_POP, what the code calls runs as usual (another such end comes
 * back to the same place, as on perl). JMPENV_JUMP goes on with what was
 * caught, whole, as XCPT_RETHROW does; with nothing caught it dies with
 * $@. The interpreter has unwound to the eval that catches the exception
 * by the time it comes back here, as perl's die_unwind has before its
 * jump. */
typedef struct goperl_jmpenv {
    jmp_buf je_buf;
    void *je_prev;     /* the frame's place to jump to before the push */
    int je_ret;        /* 0, or 3 once something came back here */
    int32_t je_failed; /* how the native call was failing when it came */
} JMPENV;
#define dJMPENV JMPENV cur_env
#define JMPENV_PUSH(v)                                 \
    do {                                               \
        cur_env.je_prev = _gof->jb;                    \
        cur_env.je_failed = 0;                         \
        _gof->jb = (void *)&cur_env.je_buf;            \
        if (GOPERL_SETJMP(cur_env.je_buf)) {           \
            cur_env.je_ret = 3;                        \
            cur_env.je_failed = _gof->failed;          \
            _gof->failed = 0;                          \
        } else {                                       \
            cur_env.je_ret = 0;                        \
        }                                              \
        (v) = cur_env.je_ret;                          \
    } while (0)
#define JMPENV_POP (_gof->jb = cur_env.je_prev)
static void goperl_jmpenv_jump(goperl_frame_t *f, int v, int32_t failed)
    __attribute__((noreturn, unused));
static void goperl_jmpenv_jump(goperl_frame_t *f, int v, int32_t failed) {
    (void)v;
    if (!failed) goperl_croak_errsv(f);
    goperl_xcpt_rethrow(f, failed);
}
#define JMPENV_JUMP(v) goperl_jmpenv_jump(_gof, (v), cur_env.je_failed)
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
/* The name of the enclosing function, as handy.h gives it. */
#define FUNCTION__ __func__
#define SAFE_FUNCTION__ __func__
#define STR_WITH_LEN(s) ("" s ""), (sizeof(s) - 1)
#define CALL_FPTR(fptr) (*(fptr))

/* as perl's XSUB.h: the CV an XSUB is called for may go unused */
#define XSPROTO(name) void name(pTHX_ CV *cv __attribute__((unused)))
/* as perl's XSUB.h: an exported XSUB has C linkage in C++ (the boot
 * function XSLoader looks up by name among them) */
#ifdef __cplusplus
#define XS_EXTERNAL(name) \
    extern "C" __attribute__((visibility("default"))) void name(pTHX_ CV *cv __attribute__((unused)))
#else
#define XS_EXTERNAL(name) \
    __attribute__((visibility("default"))) void name(pTHX_ CV *cv __attribute__((unused)))
#endif
#define XS(name) XS_EXTERNAL(name)
#define XS_INTERNAL(name) static void name(pTHX_ CV *cv __attribute__((unused)))

static void goperl_arena_free_all(void);
static void goperl_blocks_handed_free(void);
static void goperl_opreg_clear(void);
static void goperl_opreg_enter(void);
static void goperl_pads_rest(goperl_frame_t *f);
static void goperl_optree_flush(goperl_frame_t *f);
static void goperl_hooks_rest(goperl_frame_t *f);

static void goperl_xs_leave(goperl_frame_t *f) {
    /* capture the frame to make current BEFORE the host-save unwind: the
     * entry protocol parked f->prev_frame's PRE-entry value there, and the
     * unwind below writes it back (for a fresh frame that value is 0) */
    goperl_frame_t *leave_prev = (goperl_frame_t *)f->prev_frame;
    /* the entry's activation ends here if no XSUB took it over */
    if (f->entry_jb == f->jb) f->entry_jb = 0;
    goperl_errno_leave(f); /* what the XSUB left in errno */
    if (goperl_xs_depth_v <= 1) goperl_svmagic_writeback_at(f, 1);
    goperl_flush(f); /* pending mirror writes must land before teardown */
    /* Host-location saves (SAVEVPTR/save_hptr) made at the XSUB's top-level
     * scope have no host LEAVE to restore them; real perl restores them when
     * the CALLER's scope pops, which the host cannot observe. Restoring at
     * XSUB return is the closest safe point — the saved locations typically
     * point into this activation's now-dying C stack. */
    goperl_hostsave_unwind_to(f, f->hostsave_base);
    /* a restored location can be the slot of an interpreter variable
     * (SAVEI32(PL_statusvalue)): the restored value is the interpreter's */
    goperl_plshadow_flush(f);
    /* what was written into op shadows: Perl code runs next, the caller's
     * even when native code called it (Devel::Declare's handler reads the
     * line of PL_compiling its toke_skipspace moved on) */
    goperl_optree_flush(f);
    if (--goperl_xs_depth_v <= 0) {
        goperl_xs_depth_v = 0;
        goperl_lvslot_flush(f);
        goperl_lvslot_reset();
        goperl_blocks_handed_free();
        goperl_arena_free_all(); /* context-stack shadows */
        goperl_opreg_clear();
        goperl_pads_rest(f);
        goperl_hooks_rest(f);
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
/* Runs when the function that declared dXSARGS returns. An XSUB that ends
 * with a bare `return` (xsubpp's PPCODE trailer is `PUTBACK; return;`)
 * skips XSRETURN's frame-exit bookkeeping; the frame's catch point would
 * then keep naming the jmp_buf of a function that is gone, and a croak in
 * the caller of a directly called XSUB (DBI's dispatch calls driver XSUBs
 * by C pointer) would jump into it. The activation is still open exactly
 * when the frame's catch point is this function's own buffer. */
static void goperl_xs_autoleave(jmp_buf *jb) {
    goperl_frame_t *f = goperl_cur_frame_v;
    if (f && f->jb == (void *)jb) goperl_xs_leave(f);
}

/* Opens an activation on the frame: the frame becomes current, what the
 * previous activation had in it is parked on the host save-stack, and the
 * catch point of croak is the buffer declared here. */
#define GOPERL_XS_OPEN(frame, prev_cur) GOPERL_XS_OPEN_WITH(frame, prev_cur, GOPERL_XS_OPEN_RETURN)
/* GOPERL_XS_OPEN with what a croak does in the function, once the
 * activation is closed */
#define GOPERL_XS_OPEN_WITH(frame, prev_cur, after_croak)  \
    {                                                      \
        int32_t goperl_entry_base = goperl_hostsave_n;     \
        errno = GOPERL_ERRNO_UNSET;                        \
        goperl_hostsave_i32((frame), &(frame)->hostsave_base); \
        goperl_hostsave_ptr((frame), (void **)&(frame)->prev_frame); \
        goperl_hostsave_ptr((frame), (void **)&(frame)->jb); \
        for (int32_t goperl_i = goperl_entry_base; goperl_i < goperl_hostsave_n; goperl_i++) \
            goperl_hostsave_v[goperl_i].scope = GOPERL_HOSTSAVE_ACTIVATION; \
        (frame)->prev_frame = ((frame) == (prev_cur))      \
                                  ? (void *)(frame)        \
                                  : (void *)(prev_cur);    \
        goperl_xs_depth_v++;                               \
        goperl_opreg_enter();                              \
        (frame)->native_call = 1;                          \
        (frame)->hostsave_base = goperl_entry_base;        \
        (frame)->jb = (void *)&_gof_jb;                    \
        if (GOPERL_SETJMP(_gof_jb)) {                      \
            /* xs_leave unwinds the host saves itself, after reading \
             * the frame to return to: that link is one of them */  \
            goperl_xs_leave(frame);                        \
            after_croak                                    \
        }                                                  \
    }
/* A croak in an activation dXSARGS opened (inside another one, the
 * entry's at least, which returns: __goperl_xs_invoke) goes on to the
 * outer activation, as perl's croak goes past every function of native
 * code to the interpreter: the function that declared dXSARGS does not
 * return to its caller. A function of a module's own that returns a value
 * (XS::JIT's, Inline::CPP's with variable arguments) would hand its
 * caller whatever its return register held. */
#define GOPERL_XS_OPEN_RETURN goperl_raise(goperl_cur_frame_v);

/* dXSARGS works on the current frame. An XSUB the loader called through
 * its entry (__goperl_xs_invoke) takes the activation the entry opened:
 * croak has its catch point there. Every other dXSARGS opens an activation
 * of its own on the frame: an XSUB called by C pointer from another (DBI's
 * dispatch), a module's own function (Inline::C's Inline_Stack_Vars).
 *
 * It is declarations only, as perl's is: what it does is the initializer
 * of one of them, so that a module built with
 * -Werror=declaration-after-statement (C89 style) compiles. */
#define dXSARGS                                            \
    jmp_buf _gof_jb __attribute__((cleanup(goperl_xs_autoleave))); \
    const int goperl_xs_opened __attribute__((unused)) = ({ \
        goperl_frame_t *const goperl_frame_here = goperl_cur_frame_v; \
        if (goperl_frame_here->entry_jb &&                 \
            goperl_frame_here->entry_jb == goperl_frame_here->jb) \
            goperl_frame_here->entry_jb = 0;               \
        else                                               \
            GOPERL_XS_OPEN(goperl_frame_here, goperl_frame_here) \
        1;                                                 \
    });                                                    \
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
#define CvFILE(cv) \
    ((char *)goperl_intern_op(_gof, GOPERL_OP_CV_FILE, GOPERL_TOK(cv), 0, 0))
#define GvSTASH(gv) \
    (*(HV **)goperl_lv_ref(_gof, (SV *)(gv), GOPERL_LV_GVSTASH, GOPERL_LV_AT))
#define GvCVu(gv) \
    ((CV *)GOPERL_SV(goperl_op0(GOPERL_OP_GV_PTR, GOPERL_TOK(gv), 1)))
/* GvCV is the sub of the glob whatever the glob is; GvCVu leaves out the
 * one a method lookup cached there. gv_fetchmethod returns such cache
 * entries, and their callers take the sub with GvCV. */
#define GvCV(gv) \
    ((CV *)GOPERL_SV(goperl_op0(GOPERL_OP_GV_PTR, GOPERL_TOK(gv), 7)))
/* The scalar, array and hash slots of a glob are lvalues (`GvSV(gv) = sv`):
 * the plain store of gv.h, which counts no reference. */
#define GvHV(gv) \
    (*(HV **)goperl_lv_ref(_gof, (SV *)(gv), GOPERL_LV_GVHV, GOPERL_LV_AT))
#define GvAV(gv) \
    (*(AV **)goperl_lv_ref(_gof, (SV *)(gv), GOPERL_LV_GVAV, GOPERL_LV_AT))
#define GvSV(gv) \
    (*(SV **)goperl_lv_ref(_gof, (SV *)(gv), GOPERL_LV_GVSV, GOPERL_LV_AT))
#define GvIO(gv) \
    GOPERL_SV(goperl_op0(GOPERL_OP_GV_PTR, GOPERL_TOK(gv), 6))
#define GvNAME(gv) \
    ((char *)goperl_intern_op(_gof, GOPERL_OP_GV_NAME, GOPERL_TOK(gv), 0, 0))
#define isGV_with_GP(sv) \
    (goperl_op0(GOPERL_OP_SV_ISGV_GP, GOPERL_TOK(sv), 0) != 0)

/* CvXSUBANY / XSANY / alias dispatch: per-CV host storage via the loader. */
/* perl's ANY, every member (the slot of a sub's XSANY holds eight bytes). */
typedef union goperl_any {
    void *any_ptr;
    SV *any_sv;
    SV **any_svp;
    GV *any_gv;
    AV *any_av;
    HV *any_hv;
    OP *any_op;
    char *any_pv;
    char **any_pvp;
    I32 any_i32;
    U32 any_u32;
    IV any_iv;
    UV any_uv;
    long any_long;
    bool any_bool;
    size_t any_size;
    ssize_t any_ssize;
    STRLEN any_strlen;
    void (*any_dptr)(void *);
    void (*any_dxptr)(pTHX_ void *);
} ANY;
#define CvXSUBANY(cv) \
    (*(ANY *)goperl_api_cv_any(_gof, GOPERL_TOK(cv)))
#define XSANY CvXSUBANY(cv)
#define dXSI32 I32 ix = XSANY.any_i32
#define CvXSUB(cv) \
    (*(void (**)(pTHX_ CV *))goperl_lv_ref(_gof, (SV *)(cv), GOPERL_LV_CVXSUB, \
                                           GOPERL_LV_AT))

/* newXS family: every native XSUB is registered through the loader, which
 * assigns the dispatch id and returns the guest CV token. Prototypes and
 * file names are not modeled. */
static CV *goperl_newXS(goperl_frame_t *f, const char *name,
                        void (*fn)(pTHX_ CV *)) {
    goperl_flush(f);
    /* a sub named BEGIN, INIT, CHECK, UNITCHECK or END is one of perl's
     * blocks, which the warnings of PL_curcop say about (Tk turns them off
     * around its INIT): what native code wrote into the COP first */
    goperl_optree_flush(f);
    return (CV *)(uintptr_t)goperl_api_new_xs(f, name, (void *)fn);
}
#define newXS(name, fn, file) goperl_newXS(_gof, (name), (fn))
#define newXS_deffile(name, fn) goperl_newXS(_gof, (name), (fn))
#define Perl_newXS_deffile(name, fn) goperl_newXS(_gof, (name), (fn))

/* boot support */
#define XS_APIVERSION_BOOTCHECK dNOOP
#define XS_VERSION_BOOTCHECK dNOOP
#define dXSBOOTARGSXSAPIVERCHK dXSARGS
#define dXSBOOTARGSAPIVERCHK dXSARGS
/* Called as Perl_xs_boot_epilog(aTHX_ ax). It is the boot
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
        goperl_hostsave_restore(f, e);
    }
    goperl_scope_v--;
}
#define LEAVE goperl_leave(_gof)
#define SAVETMPS ((void)goperl_op0(GOPERL_OP_SAVETMPS, 0, 0))
#define FREETMPS ((void)goperl_op0(GOPERL_OP_FREETMPS, 0, 0))

/* What a save is of: a pointer that holds a reference (SAVESPTR), a
 * pointer that holds none (SAVEVPTR, SAVEPPTR, save_hptr), a pointer whose
 * reference the save owns (SAVEGENERICSV), an integer. A location of the
 * interpreter's is saved by the interpreter, which puts it back when its
 * scope ends (goperl_save_routed, goperl_compiler.h); any other is saved
 * here. */
enum {
    GOPERL_SAVEK_SPTR = 0,
    GOPERL_SAVEK_VPTR = 1,
    GOPERL_SAVEK_GENERIC = 2,
    GOPERL_SAVEK_INT = 3
};
static int goperl_save_routed(goperl_frame_t *f, void *loc, int kind, int width);
static void goperl_hostsave_ptr(goperl_frame_t *f, void **loc) {
    if (goperl_hostsave_n >= GOPERL_HOSTSAVE_MAX)
        goperl_croakf(f, "host save stack overflow in native XS");
    goperl_hostsave_t *e = &goperl_hostsave_v[goperl_hostsave_n++];
    e->loc = loc;
    e->val = *loc;
    e->scope = goperl_scope_v;
    e->width = 0;
}
/* SAVEVPTR and the other saves of a pointer that holds no reference. */
static void goperl_save_vptr(goperl_frame_t *f, void **loc) __attribute__((unused));
static void goperl_save_vptr(goperl_frame_t *f, void **loc) {
    if (!goperl_save_routed(f, (void *)loc, GOPERL_SAVEK_VPTR, (int)sizeof(void *)))
        goperl_hostsave_ptr(f, loc);
}
/* SAVEIV, SAVESTRLEN and the other saves of an integer of 8 bytes. */
static void goperl_save_int8(goperl_frame_t *f, void *loc) __attribute__((unused));
static void goperl_save_int8(goperl_frame_t *f, void *loc) {
    if (!goperl_save_routed(f, loc, GOPERL_SAVEK_INT, 8))
        goperl_hostsave_ptr(f, (void **)loc);
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
/* SAVEI32 and the other saves of an integer of 4 bytes. */
static void goperl_save_i32(goperl_frame_t *f, int32_t *loc) __attribute__((unused));
static void goperl_save_i32(goperl_frame_t *f, int32_t *loc) {
    if (!goperl_save_routed(f, (void *)loc, GOPERL_SAVEK_INT, 4)) goperl_hostsave_i32(f, loc);
}
static void goperl_hostsave_small(goperl_frame_t *f, void *loc, int width)
    __attribute__((unused));
static void goperl_hostsave_small(goperl_frame_t *f, void *loc, int width) {
    if (goperl_save_routed(f, loc, GOPERL_SAVEK_INT, width)) return;
    if (goperl_hostsave_n >= GOPERL_HOSTSAVE_MAX)
        goperl_croakf(f, "host save stack overflow in native XS");
    goperl_hostsave_t *e = &goperl_hostsave_v[goperl_hostsave_n++];
    e->loc = (void **)loc;
    e->val = (void *)(intptr_t)(width == 1 ? *(int8_t *)loc : *(int16_t *)loc);
    e->scope = goperl_scope_v;
    e->width = width;
}
#define save_hptr(hvp) goperl_save_vptr(_gof, (void **)(hvp))
#define save_aptr(avp) goperl_save_vptr(_gof, (void **)(avp))
#define SAVEVPTR(v) goperl_save_vptr(_gof, (void **)&(v))

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
    else if (!goperl_save_routed(f, (void *)loc,
                                 generic ? GOPERL_SAVEK_GENERIC : GOPERL_SAVEK_SPTR,
                                 (int)sizeof(void *))) {
        goperl_hostsave_ptr(f, (void **)loc);
        if (generic) goperl_hostsave_v[goperl_hostsave_n - 1].width = GOPERL_HOSTSAVE_GENERIC;
    }
}
#define SAVESPTR(s) goperl_save_sv_slot(_gof, (SV **)&(s), 0)
#define SAVEGENERICSV(s) goperl_save_sv_slot(_gof, (SV **)&(s), 1)

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
     (void)goperl_op3(GOPERL_OP_SAVE_HELEM, GOPERL_TOK(hv),            \
                      GOPERL_TOK(keysv), (uint32_t)(flags)))
#define save_helem(hv, keysv, svp) save_helem_flags((hv), (keysv), (svp), 0)

/* call_sv / call_method: arguments between the last PUSHMARK and the stack
 * top cross as a token list; results come back in a mortal AV and are
 * pushed where perl would leave them. G_EVAL is forced across the boundary
 * (an uncaught guest die cannot unwind host C frames); when the caller did
 * not ask for G_EVAL the error is re-raised here as a croak. */
static I32 goperl_do_call(goperl_frame_t *f, SV *sv, const char *method,
                          I32 flags) {
    I32 markoff = goperl_popmark(f);
    SV **base = f->stbase;
    SV **mark = base + markoff;
    I32 nargs = (I32)(f->psp - mark);
    char *buf = (char *)malloc((size_t)(nargs > 0 ? nargs : 0) * 8 + 256);
    size_t o = 0;
    if (!buf) goperl_croakf(f, "Out of memory!");
    if (method) {
        size_t ml = __builtin_strlen(method) + 1;
        if (ml > 255) ml = 255;
        memcpy(buf, method, ml);
        buf[ml - 1] = '\0';
        o = ml;
    }
    /* the arguments, a word each */
    for (I32 i = 0; i < nargs; i++) {
        const goperl_gword_t tok = GOPERL_TOK(mark[i + 1]);
        memcpy(buf + o, &tok, 8);
        o += 8;
    }
    /* the code called reads the ops as native code left them */
    goperl_optree_flush(f);
    /* -> the mortal array of the results; second: died<<63 | count */
    uint64_t second = 0;
    const uint64_t avtok =
        goperl_do_opx(f, method ? GOPERL_OP_CALL_METHOD : GOPERL_OP_CALL_SV,
                      method ? 0 : GOPERL_TOK(sv), o, (uint64_t)(uint32_t)flags, buf, o, &second);
    free(buf);
    int died = (int)(second >> 63);
    I32 count = (I32)(second & 0x7FFFFFFF);
    f->psp = f->stbase + markoff; /* args consumed */
    goperl_stack_extend(f, f->psp, count);
    for (I32 i = 0; i < count; i++) {
        uint64_t tok = goperl_xs_op(f, GOPERL_OP_AV_FETCH, avtok,
                                           (uint64_t)(int64_t)i, 0, 0);
        *++f->psp = (SV *)(uintptr_t)tok;
    }
    if (died && !(flags & G_EVAL)) goperl_croak_errsv(f);
    return count;
}
#define call_sv(sv, flags) goperl_do_call(_gof, (SV *)(sv), 0, (flags))
static I32 goperl_call_pv(goperl_frame_t *f, const char *name, I32 flags) {
    uint64_t cv =
        goperl_do_op(f, GOPERL_OP_GET_CV, GV_ADD, 0, name, __builtin_strlen(name));
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
        if (__builtin_strlen(e->s) == n && memcmp(e->s, p, n) == 0) return e->s;
    goperl_intern_ent_t *e =
        (goperl_intern_ent_t *)malloc(sizeof(goperl_intern_ent_t) + n);
    if (!e) return "";
    memcpy(e->s, p, n);
    e->s[n] = '\0';
    e->next = goperl_intern_v[h];
    goperl_intern_v[h] = e;
    return e->s;
}

/* An operation that answers a string of the interpreter's memory (its
 * address, 0 for none, and its length as the second value), the string
 * interned; NULL for none. */
static const char *goperl_intern_op(goperl_frame_t *f, int32_t op, uint64_t a, uint64_t b,
                                    uint64_t c) {
    uint64_t len = 0;
    const uint64_t addr = goperl_do_opx(f, op, a, b, c, 0, 0, &len);
    if (!addr) return 0;
    const char *p = (const char *)goperl_guest_mem(f, addr);
    if (!p) return 0;
    return goperl_intern(p, (size_t)len);
}
/* An operation that answers a string of the interpreter's memory: where it
 * is in this process (NULL for none), and its length in *lenp (0 for none). */
static char *goperl_str_op(goperl_frame_t *f, int32_t op, uint64_t a, uint64_t b, uint64_t c,
                           STRLEN *lenp) __attribute__((unused));
static char *goperl_str_op(goperl_frame_t *f, int32_t op, uint64_t a, uint64_t b, uint64_t c,
                           STRLEN *lenp) {
    uint64_t len = 0;
    const uint64_t addr = goperl_do_opx(f, op, a, b, c, 0, 0, &len);
    if (lenp) *lenp = addr ? (STRLEN)len : 0;
    return addr ? (char *)goperl_guest_mem(f, addr) : 0;
}
/* The second value of an operation that answers two, when its first is not
 * 0 (of one that answers a string: the length); 0 otherwise. */
static uint64_t goperl_op_second(goperl_frame_t *f, int32_t op, uint64_t a, uint64_t b,
                                 uint64_t c) __attribute__((unused));
static uint64_t goperl_op_second(goperl_frame_t *f, int32_t op, uint64_t a, uint64_t b,
                                 uint64_t c) {
    uint64_t second = 0;
    return goperl_do_opx(f, op, a, b, c, 0, 0, &second) ? second : 0;
}

/* ---- OP shadow registry --------------------------------------------------
 * An op of the interpreter is a host struct here, its shadow: one per op,
 * for as long as the op exists, so that a module can keep the pointer and
 * compare it (an op map keyed by OP *, a saved PL_check hook argument) the
 * way it does on perl. The interpreter says which ops were freed; their
 * shadows are then given up and their memory serves for the shadows of
 * later ops, as the slab of the op does in the interpreter.
 *
 * The fields of a shadow are what the interpreter held when the shadow was
 * last described. The op a hook is handed is described with everything
 * below it (an op checker, the op of a pp function) or everything that can
 * be reached from it (a peephole hook, the start of a sub), in one answer
 * of the interpreter; an op at the edge of that (the op_next of a subtree)
 * has its type, flags and target, and links that are not yet known (NULL)
 * until an accessor (OpSIBLING, CvROOT, ...) or a hook describes it. What
 * a module writes into a shadow goes back to the interpreter before the
 * next operation on op trees, and when the native call ends. */

enum {
    GOPERL_OPDUMP_ONE = 0,
    GOPERL_OPDUMP_SUBTREE = 1,
    GOPERL_OPDUMP_REACH = 2,
    GOPERL_OPDUMP_FREED = 3,
    GOPERL_OPDUMP_IS_COP = 0x80
};
enum {
    GOPERL_OPREC_TOK = 0,
    GOPERL_OPREC_NEXT = 1,
    GOPERL_OPREC_SIBPARENT = 2,
    GOPERL_OPREC_FIRST = 3,
    GOPERL_OPREC_LAST = 4,
    GOPERL_OPREC_OTHER = 5,
    GOPERL_OPREC_REDOOP = 6,
    GOPERL_OPREC_NEXTOP = 7,
    GOPERL_OPREC_LASTOP = 8,
    GOPERL_OPREC_TARG = 9,
    GOPERL_OPREC_TYPE = 10,
    GOPERL_OPREC_BITS = 11,
    GOPERL_OPREC_LINE = 12,
    GOPERL_OPREC_HINTS = 13,
    GOPERL_OPREC_FILE = 14,
    GOPERL_OPREC_FILE_LEN = 15,
    GOPERL_OPREC_STASH = 16,
    GOPERL_OPREC_STASH_LEN = 17,
    GOPERL_OPREC_SV = 18,
    GOPERL_OPREC_PV = 19,
    GOPERL_OPREC_WARNINGS = 20,
    GOPERL_OPREC_SEQ = 21,
    GOPERL_OPREC_AUX = 22,
    GOPERL_OPREC_AUX_N = 23,
    GOPERL_OPREC_FEATURES = 24,
    GOPERL_OPREC_WORDS = 25
};
#define GOPERL_OPBIT_MORESIB 1
#define GOPERL_OPBIT_COP 2
#define GOPERL_OPBIT_HOOKED 4
#define GOPERL_OPBIT_EDGE 8
#define GOPERL_OPBIT_RAW 16
#define GOPERL_OPBIT_OPT 0x10000
#define GOPERL_OPBIT_SLABBED 0x20000
#define GOPERL_OPBIT_SAVEFREE 0x40000
#define GOPERL_OPBIT_STATIC 0x80000
#define GOPERL_OPBIT_FOLDED 0x100000
#define GOPERL_OPBIT_TYPEHOOK 0x200000
#define GOPERL_PL_OP_FREED 72

typedef struct goperl_opreg_ent {
    struct goperl_opreg_ent *next;        /* in its bucket, or among the spare */
    struct goperl_opreg_ent *active_next; /* among the ones with a baseline */
    U8 persist; /* its size is a module's (NewOpSz): the memory is not reused */
    U8 active;
    U16 ext; /* a persistent one: the bytes of the module's struct past OP */
    uint32_t gen; /* the generation its links were described in */
    struct goperl_op_base *base; /* what was written back last (a write-back
                                  * baseline: see goperl_op_base_capture) */
    void *aux_own; /* the aux of the op in this process's layout (a signature's
                    * parameters), which lasts as long as the shadow: code
                    * keeps the op and reads its aux in a later call */
    OP op;        /* last: an entry of a module's own op struct extends past it */
} goperl_opreg_ent_t;
typedef struct goperl_opreg {
    goperl_opreg_ent_t **buckets;
    uint32_t nbuckets;
    uint32_t count;
    goperl_opreg_ent_t *active; /* shadows with a baseline (see goperl_op_base) */
    goperl_opreg_ent_t *spare;  /* shadows of freed ops */
    uint32_t gen;      /* moves on at every entry of native code */
    uint64_t freed_at; /* where the interpreter counts the freed ops it has
                        * not reported yet (an address of its memory) */
} goperl_opreg_t;
#define goperl_opreg_v (*(goperl_opreg_t **)&GOPERL_SHARED->opreg[0])
#define GOPERL_OPENT(o) \
    ((goperl_opreg_ent_t *)((char *)(o) - offsetof(goperl_opreg_ent_t, op)))

static void goperl_op_base_capture(OP *o);
static int goperl_raw_links(const OP *o);
static void goperl_optree_flush(goperl_frame_t *f);
static Perl_ppaddr_t *goperl_ppaddr_get(goperl_frame_t *f);
static Perl_ppaddr_t goperl_ppaddr_base(goperl_frame_t *f, int type);

static goperl_opreg_t *goperl_opreg_get(goperl_frame_t *f) {
    goperl_opreg_t *r = goperl_opreg_v;
    if (r) return r;
    r = (goperl_opreg_t *)calloc(1, sizeof(goperl_opreg_t));
    if (r) {
        r->nbuckets = 1024;
        r->buckets = (goperl_opreg_ent_t **)calloc(r->nbuckets, sizeof(void *));
    }
    if (!r || !r->buckets) goperl_croakf(f, "op shadow: out of memory");
    r->gen = 1;
    goperl_opreg_v = r;
    return r;
}

/* Native code is entered (an XSUB, a pp function, a hook): what the shadows
 * say about links is from before, and is asked for again when needed; so
 * is everything else that is kept until the interpreter moves on (the
 * context mirrors, the copy of the parser). */
static void goperl_gen_bump(void);
static void goperl_opreg_enter(void) {
    if (goperl_opreg_v) goperl_opreg_v->gen++;
    goperl_gen_bump();
    /* the interpreter ran since native code last looked at the pads */
    goperl_pads_refresh(goperl_cur_frame_v);
}

static uint32_t goperl_opreg_hash(const goperl_opreg_t *r, uint64_t tok) {
    return ((uint32_t)(tok >> 3) * 2654435761u) % r->nbuckets;
}

static goperl_opreg_ent_t *goperl_opreg_find(goperl_opreg_t *r, uint64_t tok) {
    for (goperl_opreg_ent_t *e = r->buckets[goperl_opreg_hash(r, tok)]; e; e = e->next)
        if (e->op.gop == tok) return e;
    return 0;
}

static void goperl_opreg_insert(goperl_frame_t *f, goperl_opreg_t *r,
                                goperl_opreg_ent_t *e) {
    if (r->count >= r->nbuckets * 2) {
        const uint32_t n = r->nbuckets * 4;
        goperl_opreg_ent_t **nb = (goperl_opreg_ent_t **)calloc(n, sizeof(void *));
        if (!nb) goperl_croakf(f, "op shadow: out of memory");
        const uint32_t old = r->nbuckets;
        goperl_opreg_ent_t **ob = r->buckets;
        r->buckets = nb;
        r->nbuckets = n;
        for (uint32_t i = 0; i < old; i++)
            for (goperl_opreg_ent_t *x = ob[i]; x;) {
                goperl_opreg_ent_t *next = x->next;
                const uint32_t h = goperl_opreg_hash(r, x->op.gop);
                x->next = nb[h];
                nb[h] = x;
                x = next;
            }
        free(ob);
    }
    const uint32_t h = goperl_opreg_hash(r, e->op.gop);
    e->next = r->buckets[h];
    r->buckets[h] = e;
    r->count++;
}

static goperl_opreg_ent_t *goperl_opreg_new(goperl_frame_t *f, goperl_opreg_t *r,
                                            uint64_t tok) {
    goperl_opreg_ent_t *e = r->spare;
    if (e) {
        r->spare = e->next;
        memset(e, 0, sizeof(*e));
    } else {
        e = (goperl_opreg_ent_t *)calloc(1, sizeof(goperl_opreg_ent_t));
        if (!e) goperl_croakf(f, "op shadow: out of memory");
    }
    e->op.gop = tok;
    goperl_opreg_insert(f, r, e);
    return e;
}

/* The shadows of the ops the interpreter freed are given up. */
static void goperl_opreg_sync(goperl_frame_t *f, goperl_opreg_t *r) {
    if (!r->freed_at)
        r->freed_at = goperl_xs_op(f, GOPERL_OP_PLVAR_GET, GOPERL_PL_OP_FREED, 0, 0, 0);
    if (!r->freed_at || !f->membase) return;
    if (goperl_g32(f, r->freed_at) == 0) return;
    /* -> the tokens of the freed ops, a word each; second: their number */
    uint64_t n = 0;
    const uint64_t where =
        goperl_xs_opx(f, GOPERL_OP_OP_DUMP, 0, GOPERL_OPDUMP_FREED, 0, 0, 0, &n);
    const goperl_gword_t *toks = (const goperl_gword_t *)goperl_guest_mem(f, where);
    for (uint64_t i = 0; toks && i < n; i++) {
        const uint64_t tok = toks[i];
        goperl_opreg_ent_t **p = &r->buckets[goperl_opreg_hash(r, tok)];
        for (; *p; p = &(*p)->next)
            if ((*p)->op.gop == tok) break;
        goperl_opreg_ent_t *e = *p;
        if (!e) continue;
        *p = e->next;
        r->count--;
        e->op.gop = 0; /* nothing of it is written back any more */
        if (e->active || e->persist) continue; /* given up when the list is */
        free(e->base);
        e->base = 0;
        free(e->aux_own);
        e->aux_own = 0;
        e->next = r->spare;
        r->spare = e;
    }
}

/* The native call that started it all has returned: no shadow has a
 * baseline any more. */
static void goperl_opreg_clear(void) {
    goperl_opreg_t *r = goperl_opreg_v;
    if (!r) return;
    for (goperl_opreg_ent_t *e = r->active; e;) {
        goperl_opreg_ent_t *next = e->active_next;
        free(e->base);
        e->base = 0;
        e->active = 0;
        e->active_next = 0;
        if (!e->op.gop && !e->persist) { /* its op was freed meanwhile */
            free(e->aux_own);
            e->aux_own = 0;
            e->next = r->spare;
            r->spare = e;
        }
        e = next;
    }
    r->active = 0;
    r->gen++;
}

/* The shadow has a baseline from here on: what a module writes into it is
 * seen against it and written back. */
static void goperl_op_activate(goperl_opreg_t *r, goperl_opreg_ent_t *e) {
    goperl_op_base_capture(&e->op);
    if (e->active) return;
    e->active = 1;
    e->active_next = r->active;
    r->active = e;
}

static goperl_opreg_ent_t *goperl_op_ent(goperl_frame_t *f, goperl_opreg_t *r,
                                         uint64_t tok) {
    if (!tok) return 0;
    goperl_opreg_ent_t *e = goperl_opreg_find(r, tok);
    return e ? e : goperl_opreg_new(f, r, tok);
}
static OP *goperl_op_link(goperl_frame_t *f, goperl_opreg_t *r, uint64_t tok) {
    goperl_opreg_ent_t *e = goperl_op_ent(f, r, tok);
    return e ? &e->op : 0;
}

/* The parameters of a signature, read from the interpreter's struct (g: its
 * words there: min_args, n_positional, n_named, slurpy, slurpy_padix, the
 * address of the pad indexes, the address of the named parameters, each of
 * which is padix, the address of the name, its length, its hash and the
 * required bit). The copy is the caller's to free. */
struct op_multiparam_aux;
static void *goperl_multiparam_aux(goperl_frame_t *f, uint64_t g);

/* One record of the interpreter's answer becomes the state of its shadow. */
static void goperl_op_apply(goperl_frame_t *f, goperl_opreg_t *r,
                            const goperl_gword_t *w) {
    goperl_opreg_ent_t *e = goperl_op_ent(f, r, w[GOPERL_OPREC_TOK]);
    OP *o = &e->op;
    const uint32_t bits = (uint32_t)w[GOPERL_OPREC_BITS];
    const int edge = (bits & GOPERL_OPBIT_EDGE) != 0;
    if (edge && e->active) return; /* described already, perhaps written to */
    o->op_type = (U16)(w[GOPERL_OPREC_TYPE] & 0xFFFF);
    o->op_flags = (U8)(w[GOPERL_OPREC_TYPE] >> 16);
    o->op_private = (U8)(w[GOPERL_OPREC_TYPE] >> 24);
    o->op_targ = (UV)w[GOPERL_OPREC_TARG];
    o->gop_class = (U8)(((bits >> 8) & 0xFF) + 1);
    o->op_opt = (bits & GOPERL_OPBIT_OPT) != 0;
    o->op_slabbed = (bits & GOPERL_OPBIT_SLABBED) != 0;
    o->op_savefree = (bits & GOPERL_OPBIT_SAVEFREE) != 0;
    o->op_static = (bits & GOPERL_OPBIT_STATIC) != 0;
    o->op_folded = (bits & GOPERL_OPBIT_FOLDED) != 0;
    if (bits & GOPERL_OPBIT_TYPEHOOK)
        /* made while a module had its type: PL_ppaddr's function of the
         * type, the module's (perl's newOP copies PL_ppaddr[type]) */
        o->op_ppaddr = o->op_type < OP_max ? goperl_ppaddr_get(f)[o->op_type] : 0;
    else if (!(bits & GOPERL_OPBIT_HOOKED))
        /* the interpreter's own pp function of its type, whatever a module
         * put in PL_ppaddr after the op was made: the slot the table had
         * for the type before (`o->op_ppaddr == orig_ppaddr[OP_X]`) */
        o->op_ppaddr = o->op_type < OP_max ? goperl_ppaddr_base(f, o->op_type) : 0;
    if (edge) {
        goperl_op_activate(r, e);
        return;
    }
    o->op_next = goperl_op_link(f, r, w[GOPERL_OPREC_NEXT]);
    o->op_sibparent = goperl_op_link(f, r, w[GOPERL_OPREC_SIBPARENT]);
    o->op_moresib = (bits & GOPERL_OPBIT_MORESIB) ? 1 : 0;
    if (bits & GOPERL_OPBIT_RAW) {
        /* a module's own op struct: the interpreter holds the links perl
         * reads in it (see goperl_raw_links), and everything else of it is
         * the module's, here */
        const int links = goperl_raw_links(o);
        if (links >= 1) o->op_first = goperl_op_link(f, r, w[GOPERL_OPREC_FIRST]);
        if (links >= 2) o->op_other = goperl_op_link(f, r, w[GOPERL_OPREC_OTHER]);
        o->gop_full = 1;
        e->gen = r->gen;
        goperl_op_activate(r, e);
        return;
    }
    o->op_first = goperl_op_link(f, r, w[GOPERL_OPREC_FIRST]);
    /* op_last and op_other share storage: an op has one of them */
    o->op_last = goperl_op_link(f, r, w[GOPERL_OPREC_LAST] ? w[GOPERL_OPREC_LAST]
                                                           : w[GOPERL_OPREC_OTHER]);
    o->op_redoop = goperl_op_link(f, r, w[GOPERL_OPREC_REDOOP]);
    o->op_nextop = goperl_op_link(f, r, w[GOPERL_OPREC_NEXTOP]);
    o->op_lastop = goperl_op_link(f, r, w[GOPERL_OPREC_LASTOP]);
    if (o->op_type == OP_CUSTOM && (bits & GOPERL_OPBIT_HOOKED) && o->gop_full &&
        ((bits >> 8) & 0xFF) == 13 /* OPclass_UNOP_AUX */) {
        /* a custom op with auxiliary data: the data is the module's, here */
        e->gen = r->gen;
        goperl_op_activate(r, e);
        return;
    }
    o->op_sv = GOPERL_SV(w[GOPERL_OPREC_SV]);
    o->op_u.op_meth_sv = o->op_sv;
    o->op_pv = w[GOPERL_OPREC_PV]
                   ? (char *)goperl_guest_mem(f, w[GOPERL_OPREC_PV])
                   : 0;
    if (w[GOPERL_OPREC_AUX] && o->op_type == OP_MULTIPARAM) {
        free(e->aux_own);
        e->aux_own = goperl_multiparam_aux(f, w[GOPERL_OPREC_AUX]);
        o->op_aux = (UNOP_AUX_item *)e->aux_own;
    } else if (w[GOPERL_OPREC_AUX]) {
        /* the items of a multideref, which are the interpreter's memory */
        o->op_aux = (UNOP_AUX_item *)goperl_guest_mem(f, w[GOPERL_OPREC_AUX]);
        o->op_aux_n = (U32)w[GOPERL_OPREC_AUX_N];
    } else if (o->op_type == OP_ARGELEM) {
        /* the index of the argument, kept where the pointer would be */
        o->op_aux = (UNOP_AUX_item *)(uintptr_t)w[GOPERL_OPREC_AUX_N];
    }
    if (bits & GOPERL_OPBIT_COP) {
        o->gop_iscop = 1;
        o->cop_line = (U32)w[GOPERL_OPREC_LINE];
        o->cop_hints = (U32)w[GOPERL_OPREC_HINTS];
        o->cop_seq = (U32)w[GOPERL_OPREC_SEQ];
        o->cop_features.bits[0] = (U32)w[GOPERL_OPREC_FEATURES];
        o->cop_warnings = w[GOPERL_OPREC_WARNINGS]
                              ? (char *)goperl_guest_mem(f, w[GOPERL_OPREC_WARNINGS])
                              : 0;
        /* dists compare these with strEQ and %s them without null checks */
        const char *p = w[GOPERL_OPREC_FILE]
                            ? (const char *)goperl_guest_mem(f, w[GOPERL_OPREC_FILE])
                            : 0;
        o->cop_file = p ? goperl_intern(p, (size_t)w[GOPERL_OPREC_FILE_LEN]) : "";
        p = w[GOPERL_OPREC_STASH]
                ? (const char *)goperl_guest_mem(f, w[GOPERL_OPREC_STASH])
                : 0;
        o->cop_stashpv = p ? goperl_intern(p, (size_t)w[GOPERL_OPREC_STASH_LEN]) : "";
    }
    o->gop_full = 1;
    e->gen = r->gen;
    goperl_op_activate(r, e);
}

/* Describes the op of the token and, by the mode, the ops around it; what
 * modules wrote into shadows goes to the interpreter first. */
static OP *goperl_op_describe(goperl_frame_t *f, uint64_t tok, int mode) {
    if (!tok) return 0;
    goperl_opreg_t *r = goperl_opreg_get(f);
    goperl_optree_flush(f);
    const uint64_t cap = (mode & 0x7F) == GOPERL_OPDUMP_REACH ? 4000000u : 400000u;
    /* -> the records, GOPERL_OPREC_WORDS words each; second: their number */
    uint64_t n = 0;
    const uint64_t where = goperl_do_opx(f, GOPERL_OP_OP_DUMP, tok, (cap << 8) | (uint64_t)mode,
                                         0, 0, 0, &n);
    goperl_opreg_sync(f, r);
    const goperl_gword_t *w = (const goperl_gword_t *)goperl_guest_mem(f, where);
    if (!w || !n) goperl_croakf(f, "op shadow: the interpreter did not describe the op");
    /* the ops at the edge first, so that the links of the others find them
     * with their types */
    for (uint64_t i = 0; i < n; i++)
        if (w[i * GOPERL_OPREC_WORDS + GOPERL_OPREC_BITS] & GOPERL_OPBIT_EDGE)
            goperl_op_apply(f, r, w + i * GOPERL_OPREC_WORDS);
    for (uint64_t i = 0; i < n; i++)
        if (!(w[i * GOPERL_OPREC_WORDS + GOPERL_OPREC_BITS] & GOPERL_OPBIT_EDGE))
            goperl_op_apply(f, r, w + i * GOPERL_OPREC_WORDS);
    goperl_opreg_ent_t *e = goperl_opreg_find(r, tok);
    return e ? &e->op : 0;
}

/* The shadow of an op, whatever is known of it. */
static OP *goperl_op_shadow(goperl_frame_t *f, uint64_t tok) {
    if (!tok) return 0;
    goperl_opreg_t *r = goperl_opreg_get(f);
    goperl_opreg_sync(f, r);
    goperl_opreg_ent_t *e = goperl_opreg_find(r, tok);
    if (e) return &e->op;
    return goperl_op_describe(f, tok, GOPERL_OPDUMP_ONE);
}

/* An op as a number: its address in the interpreter with the tag, or for
 * an op of the module's own its address here (see PTR2IV). */
static IV goperl_op_number(const OP *o) {
    return !o ? 0 : o->gop ? (IV)(o->gop | GOPERL_OP_NUMBER_TAG) : (IV)(intptr_t)o;
}
/* The op of a number (see PTR2IV): with the tag, the interpreter's op at
 * the rest of the number, as its shadow; without it, the op at that address
 * here (a module's own, or a shadow that was cast to a number). Outside a
 * call of native code there is no interpreter to ask: no op. */
static OP *goperl_op_of_number(IV iv) {
    goperl_frame_t *const f = goperl_cur_frame_v;
    if (!iv) return 0;
    if (!((UV)iv & GOPERL_OP_NUMBER_TAG)) return (OP *)(intptr_t)iv;
    return f ? goperl_op_shadow(f, (uint64_t)((UV)iv & ~GOPERL_OP_NUMBER_TAG)) : 0;
}

/* The shadow of an op with what is below it, as the interpreter has it
 * now (asked once for every entry of native code). */
static OP *goperl_op_shadow_full(goperl_frame_t *f, uint64_t tok) {
    if (!tok) return 0;
    goperl_opreg_t *r = goperl_opreg_get(f);
    goperl_opreg_sync(f, r);
    goperl_opreg_ent_t *e = goperl_opreg_find(r, tok);
    if (e && e->op.gop_full && e->gen == r->gen) return &e->op;
    return goperl_op_describe(f, tok, GOPERL_OPDUMP_SUBTREE);
}

/* The same with every op that can be reached from it: what code that walks
 * op_next chains (a peephole hook) is handed. */
static OP *goperl_op_shadow_reach(goperl_frame_t *f, uint64_t tok) {
    return goperl_op_describe(f, tok, GOPERL_OPDUMP_REACH);
}

static void goperl_op_pull(goperl_frame_t *f, OP *o) {
    goperl_op_describe(f, o->gop, GOPERL_OPDUMP_ONE |
                                      (o->gop_iscop ? GOPERL_OPDUMP_IS_COP : 0));
}

/* Refresh after a guest optree operation: the guest may have relinked or
 * retyped the participants, or freed one (op_append_list frees the list
 * whose kids it moved into the first; a constructor frees an operand it
 * folds into a constant). The ops the interpreter reports freed are given
 * up first, and a participant left without a shadow is not asked for: its
 * memory is the slab's or the allocator's again and holds anything. */
static void goperl_op_refresh(goperl_frame_t *f, OP *o) {
    if (!o || !o->gop) return;
    goperl_opreg_sync(f, goperl_opreg_get(f));
    if (o->gop) goperl_op_pull(f, o); /* a shadow given up has no token */
}
static void goperl_op_refresh_tok(goperl_frame_t *f, uint64_t tok) {
    if (!tok) return;
    goperl_opreg_t *r = goperl_opreg_get(f);
    goperl_opreg_sync(f, r);
    if (goperl_opreg_find(r, tok)) goperl_op_describe(f, tok, GOPERL_OPDUMP_ONE);
}

/* Materialize a token the CALLER knows is a COP (PL_curcop, blk_oldcop):
 * static interpreter COPs like PL_compiling have op_type 0, so cop-ness
 * cannot be inferred from the type. */
static COP *goperl_cop_shadow(goperl_frame_t *f, uint64_t tok) {
    if (!tok) return 0;
    goperl_opreg_t *r = goperl_opreg_get(f);
    goperl_opreg_sync(f, r);
    goperl_opreg_ent_t *e = goperl_opreg_find(r, tok);
    if (e && e->op.gop_iscop && e->gen == r->gen) return (COP *)&e->op;
    return (COP *)goperl_op_describe(f, tok,
                                     GOPERL_OPDUMP_ONE | GOPERL_OPDUMP_IS_COP);
}

/* In an embedded interpreter PL_curcop CAN be NULL: eval_sv driven from C
 * enters with no compiling COP, so the eval context's blk_oldcop is NULL
 * and leaveeval restores that. Native perl code dereferences PL_curcop
 * without checks, so hand it a stable dummy COP (line 0, package main)
 * instead — the same shape modules already handle for optimized-away
 * COPs. */
GOPERL_C __attribute__((weak, visibility("hidden"))) OP goperl_nullcop_v = {
    0, 0, 0, 0, OP_NEXTSTATE, 0, 0, 0, /* base scalars */
    1 /* gop_full */, 1 /* gop_iscop */, 0 /* moresib */, 12 /* a COP */,
    0, 0, 0, 0, 0 /* the flag bits */, 0 /* gop */,
    0 /* op_first */, {0}, 0 /* op_redoop */,
    0 /* cop_line */, 0 /* cop_hints */, 0 /* cop_warnings */, "", "main"};
static COP *goperl_curcop(goperl_frame_t *f) {
    uint64_t tok = goperl_do_op(f, GOPERL_OP_PLVAR_GET, GOPERL_PL_CURCOP, 0,
                                0, 0);
    if (!tok) return (COP *)&goperl_nullcop_v;
    return goperl_cop_shadow(f, tok);
}

/* PL_curcop is an lvalue: a module points it at another COP of the
 * program (the caller's, so that a message it croaks with names the
 * caller's line) or saves and overrides it around a block. Every
 * evaluation loads the slot from the interpreter, and a COP of the
 * interpreter's that the module stored there since is written back before
 * the next operation (goperl_flush), like any other lvalue. A COP the
 * module made up in its own memory is no op of the interpreter's and stays
 * here: such an override holds for the module's own reads until the next
 * evaluation (Cpanel::JSON::XS points PL_curcop at a copy with other hints
 * while it sorts; the comparisons the interpreter runs for it keep the
 * hints of the real COP). */
#define goperl_curcop_slot_v (*(COP **)&GOPERL_SHARED->curcop_slot)
#define goperl_curcop_seen_v (*(COP **)&GOPERL_SHARED->curcop_seen)
static void goperl_curcop_writeback(goperl_frame_t *f) {
    COP *const now = goperl_curcop_slot_v;
    if (!goperl_curcop_seen_v || now == goperl_curcop_seen_v) return;
    /* before the operation: it flushes again */
    goperl_curcop_seen_v = now;
    if (now && ((const OP *)now)->gop)
        goperl_xs_op(f, GOPERL_OP_PLVAR_SET, (uint64_t)GOPERL_PL_CURCOP,
                     ((const OP *)now)->gop, 0, 0);
}
static COP **goperl_curcop_ref(goperl_frame_t *f) {
    COP *const cop = goperl_curcop(f);
    goperl_curcop_slot_v = goperl_curcop_seen_v = cop;
    return &goperl_curcop_slot_v;
}

/* Whether the shadow of an op of the interpreter's has what the
 * interpreter has now: it was described in full since native code was last
 * entered. A shadow kept from an earlier call has the links of then — that
 * of an op a module's check function saw while the op was being built has
 * no op_next yet, which the compiler set after (autobox's method calls,
 * whose entersub Want looks for by its op_next) — and is described again. */
static inline int goperl_op_current(const OP *o) {
    if (!o->gop_full) return 0;
    const goperl_opreg_t *const r = goperl_opreg_v;
    return !r || GOPERL_OPENT(o)->gen == r->gen;
}

/* A COP at the edge of what was described gets its fields now, as an op
 * gets its links in OpSIBLING. Outside a call of native code there is no
 * interpreter to ask, and the shadow is as it is. */
static COP *goperl_cop_fields(COP *c) {
    goperl_frame_t *const f = goperl_cur_frame_v;
    if (c && c->gop && f && !goperl_op_current((const OP *)c))
        goperl_op_describe(f, c->gop, GOPERL_OPDUMP_ONE);
    return c;
}

static OP *goperl_op_sibling(goperl_frame_t *f, OP *o) {
    if (!o) return 0;
    /* an op at the edge of what was described gets its links now */
    if (o->gop && !goperl_op_current(o)) goperl_op_describe(f, o->gop, GOPERL_OPDUMP_SUBTREE);
    OP *sib = o->op_moresib ? o->op_sibparent : 0;
    if (sib && sib->gop && !goperl_op_current(sib))
        goperl_op_describe(f, sib->gop, GOPERL_OPDUMP_SUBTREE);
    return sib;
}

#define OP_NAME(o) \
    ({ GOPERL_ONCE(goperl_OP_NAME_o_, o); \
       PL_op_name[goperl_OP_NAME_o_->op_type ? goperl_OP_NAME_o_->op_type : (int)goperl_OP_NAME_o_->op_targ]; })

/* Op-tree dumping needs interpreter internals the SDK does not model. */
static void goperl_do_op_dump(I32 level, void *fp, const OP *o)
    __attribute__((unused));
static void goperl_do_op_dump(I32 level, void *fp, const OP *o) {
    (void)level;
    goperl_perlio_printf(fp,
                         "<op dump unavailable in the go-perl XS SDK: %s>\n",
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
 * A call with an op of the module's own making (gop == 0) runs the pp
 * function of the type on a copy of this function's stack (RUN_PP). */

static OP *goperl_pp_dispatch(pTHX_ int type);
static OP *goperl_newop_dispatch(pTHX_ int type, OP *o);
static void goperl_pp_stack_push(goperl_frame_t *f);
#include "goperl_ppslots.h"

GOPERL_C __attribute__((weak, visibility("hidden"))) Perl_ppaddr_t goperl_ppaddr_base_v[OP_max] = {0};
GOPERL_C __attribute__((weak, visibility("hidden"))) uint8_t goperl_pp_installed_v[OP_max] = {0};
GOPERL_C __attribute__((weak, visibility("hidden"))) int32_t goperl_ppaddr_inited_v = 0;

static OP *goperl_pp_dispatch(pTHX_ int type) {
    OP *cur = _gof->plop;
    if (cur && cur->gop) {
        /* real guest op under execution: run the true pp for `type`, on
         * the stack as this function leaves it */
        goperl_pp_stack_push(_gof);
        /* the op is PL_op while its pp runs (it is already, unless a run
         * loop of native code made it PL_op) */
        uint64_t next = goperl_do_op(_gof, GOPERL_OP_RUN_ORIGINAL,
                                     (uint64_t)(uint32_t)type, cur->gop, 0, 0);
        _gof->st_stale = _gof->pp_frame;
        return goperl_op_shadow(_gof, next);
    }
    /* An op of the module's own making (`BINOP myop; ... PL_op = (OP *)&myop;
     * PL_ppaddr[myop.op_type](aTHX)`: Tie::Hash::Indexed adds to a value
     * the way `+=` does, a range is expanded with pp_flop). The
     * interpreter runs the pp function of the type for an op with this
     * one's flags, private flags and target, on a stack that holds what
     * this function's holds — every value and every mark — and what is on
     * the stack afterwards, and how many marks the function took, is this
     * function's stack from then on. */
    goperl_frame_t *const f = _gof;
    goperl_marks_adopt(f);
    const I32 n = (I32)(f->psp - f->stbase); /* the values are st[1..n] */
    const I32 nmarks = f->markidx + 1 > 0 ? f->markidx + 1 : 0;
    /* words: [marks][op_targ][an offset for each mark][the values] */
    const size_t bytes = sizeof(goperl_gword_t) * (size_t)(2 + nmarks + (n > 0 ? n : 0));
    goperl_gword_t *const w = (goperl_gword_t *)malloc(bytes);
    if (!w) goperl_croakf(f, "Out of memory!");
    w[0] = (goperl_gword_t)nmarks;
    w[1] = cur ? (goperl_gword_t)cur->op_targ : 0;
    for (I32 i = 0; i < nmarks; i++) w[2 + i] = (goperl_gword_t)(int64_t)f->marks[i];
    for (I32 i = 0; i < n; i++) w[2 + nmarks + i] = GOPERL_TOK(f->stbase[1 + i]);
    const uint64_t op = ((uint64_t)(cur ? cur->op_private : 0) << 40) |
                        ((uint64_t)(cur ? cur->op_flags : 0) << 32) |
                        (uint32_t)type;
    goperl_flush(f);
    const int32_t failed_before = f->failed;
    /* -> the array of what is on the stack; second: taken<<32 | count */
    uint64_t second = 0;
    const uint64_t avtok =
        goperl_xs_opx(f, GOPERL_OP_RUN_PP, op, bytes, 0, (const char *)w, bytes, &second);
    goperl_gen_bump();
    free(w);
    if (!failed_before && f->failed && f->jb) goperl_raise(f);
    const I32 count = (I32)(uint32_t)second;
    const I32 taken = (I32)(second >> 32);
    f->psp = f->stbase;
    goperl_stack_extend(f, f->psp, count);
    for (I32 i = 0; i < count; i++) {
        uint64_t tok = goperl_xs_op(f, GOPERL_OP_AV_FETCH, avtok,
                                    (uint64_t)(int64_t)i, 0, 0);
        *++f->psp = (SV *)(uintptr_t)tok;
    }
    f->markidx -= taken;
    goperl_marks_publish(f);
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
/* The slot PL_ppaddr had for a type before any module wrote it. */
static Perl_ppaddr_t goperl_ppaddr_base(goperl_frame_t *f, int type) {
    goperl_ppaddr_get(f);
    return goperl_ppaddr_base_v[type];
}

/* Detect module writes to the proxy table and (de)register them as guest
 * pp hooks. Called from the funnel flush. Uses the raw vtable (never the
 * funnel) to avoid recursion. */
static void goperl_ppaddr_sync(goperl_frame_t *f) {
    if (!goperl_ppaddr_inited_v) return;
    for (int i = 0; i < OP_max; i++) {
        int hooked = goperl_ppaddr_v[i] != goperl_ppaddr_base_v[i];
        if (hooked && !goperl_pp_installed_v[i]) {
            goperl_pp_installed_v[i] = 1;
            goperl_api_pp_hook_set(f, i, (void *)goperl_ppaddr_v[i]);
        } else if (!hooked && goperl_pp_installed_v[i]) {
            goperl_pp_installed_v[i] = 0;
            goperl_api_pp_hook_set(f, i, 0);
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
    U8 used; /* 1: fn(aTHX_ arg), SAVEDESTRUCTOR_X; 2: fn(arg), SAVEDESTRUCTOR;
              * 3: fn(aTHX_ arg) that releases memory of this process (the
              * buffer of SAVEFREEPV or SSNEW, a deferred save): the copies
              * of an instance share that memory, so a copy leaves its
              * release to the original (see __goperl_shared_clone) */
} goperl_dtor_t;
#define GOPERL_DTOR_X 1
#define GOPERL_DTOR_PLAIN 2
#define GOPERL_DTOR_CALL_OWN 3
#define goperl_dtor_v ((goperl_dtor_t *)GOPERL_SHARED->dtors)
#define goperl_dtor_hint_v (GOPERL_SHARED->dtor_hint)
typedef char goperl_dtor_fits
    [(sizeof(goperl_dtor_t) * GOPERL_DTOR_MAX <=
      sizeof(((goperl_shared_t *)0)->dtors))
         ? 1
         : -1] __attribute__((unused));

static void goperl_save_destructor(goperl_frame_t *f,
                                   void (*fn)(pTHX_ void *), void *arg,
                                   U8 kind) {
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
    goperl_dtor_v[id].used = kind;
    goperl_do_op(f, GOPERL_OP_SAVE_DESTRUCTOR, (uint64_t)(uint32_t)(id + 1),
                 0, 0, 0);
}
#define save_destructor_x(fn, p) \
    goperl_save_destructor(_gof, (fn), (p), GOPERL_DTOR_X)
#define SAVEDESTRUCTOR_X(fn, p) \
    goperl_save_destructor(_gof, (fn), (p), GOPERL_DTOR_X)
/* SAVEDESTRUCTOR's function takes the argument alone (void (*)(void *)). */
#define save_destructor(fn, p)                                           \
    goperl_save_destructor(_gof, (void (*)(pTHX_ void *))(void (*)(void))(fn), \
                           (void *)(p), GOPERL_DTOR_PLAIN)
#define SAVEDESTRUCTOR(fn, p) save_destructor((fn), (p))

/* ---- saves of native code's own locations at the end of a native call --
 * An XSUB runs inside the scope perl's entersub opens for it, so what it
 * saved without a scope of its own is put back when it returns (see
 * goperl_xs_leave). A keyword plugin, a pp function, an op checker or a
 * call checker runs inside a scope of the interpreter's that goes on after
 * it returns: what it saved is put back when that scope ends, as perl does
 * (Object::Pad saves the class's method scope while a method is compiled).
 * At the end of such a call each save it left becomes a destructor of that
 * scope, which puts the location back then; a location on the call's own C
 * stack, which is gone once it returns, is put back at once. */
typedef struct goperl_deferred_save {
    goperl_hostsave_t e;
} goperl_deferred_save_t;
static void goperl_deferred_save_cb(pTHX_ void *p) {
    goperl_deferred_save_t *d = (goperl_deferred_save_t *)p;
    goperl_hostsave_restore(_gof, &d->e);
    free(d);
}
/* The range of the C stack of the calling thread below the frame given. */
static int goperl_on_call_stack(const void *loc, const void *entry_frame) {
    uintptr_t lo = 0;
#if defined(__APPLE__)
    const pthread_t self = pthread_self();
    lo = (uintptr_t)pthread_get_stackaddr_np(self) - (uintptr_t)pthread_get_stacksize_np(self);
#else
    pthread_attr_t attr;
    void *addr = 0;
    size_t size = 0;
    if (pthread_getattr_np(pthread_self(), &attr) == 0) {
        if (pthread_attr_getstack(&attr, &addr, &size) == 0) lo = (uintptr_t)addr;
        pthread_attr_destroy(&attr);
    }
#endif
    return lo && (uintptr_t)loc >= lo && (uintptr_t)loc < (uintptr_t)entry_frame;
}
static void goperl_hostsave_settle(goperl_frame_t *f, int32_t base, const void *entry_frame) {
    /* the destructors go to the interpreter's save stack in the order the
     * saves were made, so that they run in reverse, as the saves would */
    int32_t n = goperl_hostsave_n;
    for (int32_t i = base; i < n; i++) {
        goperl_hostsave_t *e = &goperl_hostsave_v[i];
        if (goperl_on_call_stack(e->loc, entry_frame)) continue;
        goperl_deferred_save_t *d = (goperl_deferred_save_t *)malloc(sizeof(*d));
        if (!d) goperl_croakf(f, "Out of memory!");
        d->e = *e;
        e->loc = 0; /* handed over */
        goperl_save_destructor(f, goperl_deferred_save_cb, d, GOPERL_DTOR_CALL_OWN);
    }
    /* what is on the call's stack is put back now, last saved first */
    while (goperl_hostsave_n > base) {
        goperl_hostsave_t *e = &goperl_hostsave_v[--goperl_hostsave_n];
        if (e->loc) goperl_hostsave_restore(f, e);
    }
}

/* Loader entry: run a fired destructor (from guest scope pop). */
GOPERL_C __attribute__((weak, visibility("default"), used)) void
__goperl_dtor_invoke(goperl_frame_t *f, uint32_t id1) {
    if (id1 == 0 || id1 > GOPERL_DTOR_MAX) return;
    jmp_buf jb;
    void *prev_jb = f->jb;
    goperl_frame_t *prev_frame = goperl_cur_frame_v;
    f->jb = (void *)&jb;
    /* current frame before the table read: the destructor table is
     * per-instance state reached through it */
    goperl_frame_enter(f);
    goperl_dtor_t d = goperl_dtor_v[id1 - 1];
    goperl_dtor_v[id1 - 1].used = 0;
    if (!d.used || !d.fn) {
        f->jb = prev_jb;
        goperl_cur_frame_v = prev_frame;
        return;
    }
    if (GOPERL_SETJMP(jb)) {
        /* the destructor died: the frame keeps the failure, which the
         * loader hands to the interpreter, and the scope's end dies with
         * it there */
        f->jb = prev_jb;
        goperl_cur_frame_v = prev_frame;
        return;
    }
    if (d.used == GOPERL_DTOR_PLAIN)
        ((void (*)(void *))(void (*)(void))d.fn)(d.arg);
    else
        d.fn(d.arg);
    /* what the destructor stored (PAD_SVl(ix) = NULL) is the interpreter's
     * before it goes on with the scope */
    goperl_flush(f);
    f->jb = prev_jb;
    goperl_cur_frame_v = prev_frame;
}

/* Loader entry: call an XSUB. The frame is current and has a catch point
 * before the function's first statement, so an XSUB written without dXSARGS
 * (Params::Classify's read the stack with dMARK and dSP alone) reads the
 * stack, reaches the interpreter and croaks like any other. The dXSARGS of
 * an XSUB that has one takes this activation over. */
GOPERL_C __attribute__((weak, visibility("default"), used)) void
__goperl_xs_invoke(goperl_frame_t *f, void *fnv, uint64_t cv) {
    void (*fn)(pTHX_ CV *) = (void (*)(pTHX_ CV *))fnv;
    jmp_buf _gof_jb __attribute__((cleanup(goperl_xs_autoleave)));
    goperl_frame_t *prev_cur = goperl_cur_frame_v;
    goperl_frame_enter(f);
    /* the outermost activation: a croak returns to the loader, also when
     * the module is C++ */
    GOPERL_XS_OPEN_WITH(f, prev_cur, return;)
    f->entry_jb = (void *)&_gof_jb;
    fn((CV *)(uintptr_t)cv);
}

/* Loader entry: close any activation levels an XSUB left open. xsubpp's
 * PPCODE trailer is `PUTBACK; return;` with no XSRETURN, so the entry
 * bookkeeping (depth, saved prev_frame/jb/hostsave_base) is still armed
 * when the C function returns; the loader calls this after every XSUB.
 * Each xs_leave unwinds that activation's host saves, which restores the
 * previous prev_frame/jb/base — so the loop also settles bare-returning
 * NESTED direct calls, one level at a time. */
GOPERL_C __attribute__((weak, visibility("default"), used)) uint32_t
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

/* Loader entry: run one mirror-MAGIC hook (svt_free, svt_local — the
 * int (*)(pTHX_ SV*, MAGIC*) shape) with the current frame installed and a
 * croak guard armed. The loader calls this instead of the raw function
 * pointer so interpreter macros inside the hook see the right frame. It
 * returns 0 when the hook croaked, with the message in f->err: the loader
 * has the free or the `local` die with it, as perl lets it out. */
GOPERL_C __attribute__((weak, visibility("default"), used)) uint32_t
__goperl_svt_invoke(goperl_frame_t *f, void *fnv, uint64_t sv, void *mg) {
    int (*fn)(pTHX_ SV *, MAGIC *) = (int (*)(pTHX_ SV *, MAGIC *))fnv;
    jmp_buf jb;
    void *prev_jb = f->jb;
    goperl_frame_t *prev_frame = goperl_cur_frame_v;
    f->jb = (void *)&jb;
    goperl_frame_enter(f);
    if (GOPERL_SETJMP(jb)) {
        f->failed = 0;
        f->jb = prev_jb;
        goperl_cur_frame_v = prev_frame;
        return 0;
    }
    fn((SV *)(uintptr_t)sv, (MAGIC *)mg);
    goperl_flush(f);
    f->jb = prev_jb;
    goperl_cur_frame_v = prev_frame;
    return 1;
}

/* Loader entry: run a module pp hook for one guest op execution.
 * Returns the next-op token (0 = NULL); on croak, f->failed is set and
 * the message is in f->err. */
/* ---- the perl stack of a pp function --------------------------------------
 * A pp function reads and writes the argument stack and the mark stack. A
 * module's own runs here, on a copy of both that the loader made from the
 * interpreter's when it called the function: st[i] is the interpreter's
 * PL_stack_base[i], and marks[1..] are its top marks (marks[0] stands for
 * the mark below those, the interpreter's index gmark_lo). st_seen and
 * marks_seen are that copy as it was when the two were last in line.
 *
 * What the function made of its copy goes to the interpreter when the
 * function returns (pp_state, with the next op) and before every operation
 * that runs interpreter code on the stack (the pp function it stands in
 * for, dounwind); after such an operation the copy is behind (st_stale)
 * and is fetched again when the function next looks at it. */
enum { GOPERL_PPSTACK_PULL = 0, GOPERL_PPSTACK_PUSH = 1 };

static void goperl_pp_seen_capture(goperl_frame_t *f) {
    const int64_t n = (int64_t)(f->psp - f->stbase);
    uint64_t *seen = (uint64_t *)realloc(f->st_seen, (size_t)(n + 1) * 8);
    if (!seen) goperl_croakf(f, "Out of memory for the stack of a pp function");
    memcpy(seen, f->stbase, (size_t)(n + 1) * 8);
    f->st_seen = seen;
    f->st_seen_n = n;
    f->mark_seen = f->markidx;
    memcpy(f->marks_seen, f->marks, sizeof(f->marks_seen));
}

/* What the copy became since it was last in line, as the interpreter takes
 * it, in words: [depth][lo] and the values above lo, [top mark][mlo] and
 * the marks above mlo. NULL when nothing changed. The caller frees it. */
static char *goperl_pp_state_build(goperl_frame_t *f, int64_t *lenp) {
    goperl_marks_adopt(f);
    const int64_t depth = (int64_t)(f->psp - f->stbase);
    const int64_t common = depth < f->st_seen_n ? depth : f->st_seen_n;
    int64_t lo = common;
    for (int64_t i = 1; i <= common; i++)
        if ((uint64_t)(uintptr_t)f->stbase[i] != f->st_seen[i]) {
            lo = i - 1;
            break;
        }
    const int32_t mcommon = f->markidx < f->mark_seen ? f->markidx : f->mark_seen;
    int32_t mlo = mcommon < 0 ? 0 : mcommon;
    for (int32_t j = 1; j <= mcommon; j++)
        if (f->marks[j] != f->marks_seen[j]) {
            mlo = j - 1;
            break;
        }
    if (depth == f->st_seen_n && lo == depth && f->markidx == f->mark_seen &&
        mlo == f->markidx)
        return 0;
    const int32_t midx = f->markidx < 0 ? 0 : f->markidx;
    const size_t len = (size_t)(4 + (depth - lo) + (midx - mlo)) * sizeof(goperl_gword_t);
    goperl_gword_t *w = (goperl_gword_t *)malloc(len);
    if (!w) goperl_croakf(f, "Out of memory for the stack of a pp function");
    size_t k = 0;
    w[k++] = (goperl_gword_t)depth;
    w[k++] = (goperl_gword_t)lo;
    for (int64_t i = lo + 1; i <= depth; i++) w[k++] = GOPERL_TOK(f->stbase[i]);
    w[k++] = (goperl_gword_t)(int64_t)(f->gmark_lo + midx);
    w[k++] = (goperl_gword_t)(int64_t)(f->gmark_lo + mlo);
    for (int32_t j = mlo + 1; j <= midx; j++) w[k++] = (goperl_gword_t)(int64_t)f->marks[j];
    *lenp = (int64_t)len;
    return (char *)w;
}

/* Before interpreter code runs on the stack: it gets what the function
 * made of its copy. */
static void goperl_pp_stack_push(goperl_frame_t *f) {
    if (!f->pp_frame || f->st_stale) return;
    int64_t len = 0;
    char *state = goperl_pp_state_build(f, &len);
    if (!state) return;
    goperl_xs_op(f, GOPERL_OP_PP_STACK, GOPERL_PPSTACK_PUSH, (uint64_t)len,
                 state, (uint64_t)len);
    free(state);
    goperl_pp_seen_capture(f);
}

/* The copy is fetched again, as words: [depth][top mark][n] the n top
 * marks, then the values. */
static void goperl_pp_stack_pull(goperl_frame_t *f) {
    f->st_stale = 0;
    uint64_t nwords = 0;
    const uint64_t where =
        goperl_xs_opx(f, GOPERL_OP_PP_STACK, GOPERL_PPSTACK_PULL, 0, 0, 0, 0, &nwords);
    const goperl_gword_t *w = (const goperl_gword_t *)goperl_guest_mem(f, where);
    if (!w || nwords < 3)
        goperl_croakf(f, "the interpreter did not hand over its stack");
    const int64_t depth = (int64_t)w[0];
    const int32_t top = (int32_t)(int64_t)w[1];
    int32_t n = (int32_t)w[2];
    if (n > GOPERL_XS_MARKS - 1) n = GOPERL_XS_MARKS - 1;
    if (depth + 64 >= f->stcap) {
        const int64_t ncap = depth + 128;
        SV **nb = (SV **)malloc((size_t)ncap * sizeof(SV *));
        if (!nb) goperl_croakf(f, "Out of memory during stack extend");
        if (f->stbase != (SV **)f->st) free(f->stbase);
        f->stbase = nb;
        f->stcap = ncap;
    }
    f->stbase[0] = 0;
    for (int64_t i = 0; i < depth; i++)
        f->stbase[1 + i] = GOPERL_SV(w[3 + (int32_t)w[2] + i]);
    f->psp = f->stbase + depth;
    f->gmark_lo = top - n;
    f->marks[0] = 0;
    for (int32_t j = 0; j < n; j++)
        f->marks[1 + j] = (int32_t)(int64_t)w[3 + ((int32_t)w[2] - n) + j];
    f->markidx = n;
    goperl_marks_publish(f);
    goperl_pp_seen_capture(f);
}

/* Loader entry: run a module pp function for one guest op execution.
 * Returns the next-op token (0 = NULL), and leaves what the function made
 * of the stack in the frame; on croak, f->failed is set and the message is
 * in f->err. */
GOPERL_C __attribute__((weak, visibility("default"), used)) uint64_t
__goperl_pp_invoke(goperl_frame_t *f, void *fnv) {
    OP *(*fn)(pTHX) = (OP * (*)(pTHX)) fnv;
    jmp_buf jb;
    f->jb = (void *)&jb;
    f->prev_frame = (void *)goperl_cur_frame_v;
    goperl_frame_enter(f);
    goperl_xs_depth_v++;
    goperl_opreg_enter();
    f->native_call = 1;
    f->hostsave_base = goperl_hostsave_n;
    if (GOPERL_SETJMP(jb)) {
        goperl_hostsave_unwind_to(f, f->hostsave_base);
        free(f->st_seen);
        f->st_seen = 0;
        goperl_xs_leave(f);
        return 0;
    }
    goperl_pp_seen_capture(f);
    f->plop = goperl_op_describe(f, f->hook_op_tok, GOPERL_OPDUMP_SUBTREE);
    OP *next = fn();
    uint64_t tok = next ? next->gop : 0;
    goperl_hostsave_settle(f, f->hostsave_base, __builtin_frame_address(0));
    if (!f->st_stale) f->pp_state = goperl_pp_state_build(f, &f->pp_state_len);
    free(f->st_seen);
    f->st_seen = 0;
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

/* The end of the block's scope releases the block rather than freeing it:
 * perl's savestack keeps what was there until the stack grows over it
 * again, and code reads a block right after an op that left its scope
 * (Devel::NYTProf looks there whether the destructor it saved beside the
 * block has run). The block goes when its slot is taken again. A released
 * block's pointer has its low bit set (calloc's blocks are aligned): which
 * blocks were released cannot be told otherwise. */
#define GOPERL_SSNEW_RELEASED ((uintptr_t)1)
static void *goperl_ssnew_block(void *mem) {
    return (void *)((uintptr_t)mem & ~GOPERL_SSNEW_RELEASED);
}
static void goperl_ssnew_free_cb(pTHX_ void *p) {
    int32_t ix = (int32_t)(intptr_t)p;
    (void)_gof;
    if (ix >= 0 && ix < GOPERL_SSNEW_MAX && goperl_ssnew_v[ix].mem)
        goperl_ssnew_v[ix].mem =
            (void *)((uintptr_t)goperl_ssnew_v[ix].mem | GOPERL_SSNEW_RELEASED);
}

static SSize_t goperl_ssnew(goperl_frame_t *f, size_t size) {
    int32_t ix = -1;
    for (int32_t k = 0; k < GOPERL_SSNEW_MAX; k++) {
        int32_t i = (goperl_ssnew_hint_v + k) % GOPERL_SSNEW_MAX;
        void *mem = goperl_ssnew_v[i].mem;
        if (!mem || ((uintptr_t)mem & GOPERL_SSNEW_RELEASED)) {
            free(goperl_ssnew_block(mem));
            goperl_ssnew_v[i].mem = 0;
            ix = i;
            break;
        }
    }
    if (ix < 0) goperl_croakf(f, "SSNEW: scratch table full");
    goperl_ssnew_hint_v = ix + 1;
    goperl_ssnew_v[ix].mem = calloc(1, size);
    if (!goperl_ssnew_v[ix].mem) goperl_croakf(f, "SSNEW: out of memory");
    goperl_save_destructor(f, goperl_ssnew_free_cb, (void *)(intptr_t)ix,
                           GOPERL_DTOR_CALL_OWN);
    return (SSize_t)ix;
}
#define SSNEW(size) goperl_ssnew(_gof, (size_t)(size))
#define SSNEWa(size, align) goperl_ssnew(_gof, (size_t)(size))
#define SSPTR(ix, type) ((type)goperl_ssnew_block(goperl_ssnew_v[(ix)].mem))
#define SSNEWt(n, t) SSNEW((n) * sizeof(t))
#define SSNEWat(n, t, align) SSNEWa((n) * sizeof(t), (align))
#define SSPTRt(ix, t) SSPTR((ix), t *)
#define MEM_ALIGNBYTES 8

/* ---- context stack shadows -----------------------------------------------
 * PERL_CONTEXT / PERL_SI mirrors materialized from the live interpreter,
 * refreshed whenever any guest operation may have changed them (a
 * generation counter bumped by the funnel). Allocations go to an arena
 * freed when the outermost native frame returns. */

typedef struct goperl_si PERL_SI;
/* A context as the interpreter has it. The members of the kinds of block
 * (sub, format, eval, loop) are apart here, where the interpreter overlays
 * them: each is filled for a context of its kind. What native code writes
 * into a context does not reach the interpreter. */
typedef struct goperl_context {
    U8 cx_type; /* CXt_ type and CXp_ bits, as in the interpreter */
    U8 blk_gimme;
    U16 blk_u16;
    I32 blk_oldsaveix;
    SSize_t blk_oldsp;
    I32 blk_oldmarksp;
    I32 blk_oldscopesp;
    SSize_t blk_old_tmpsfloor;
    COP *blk_oldcop;
    OP *blk_oldpm;
    struct {
        CV *cv;
        OP *retop;
        I32 olddepth;
        AV *savearray;
        AV *prevcomppad;
    } blk_sub;
    struct {
        CV *cv;
        OP *retop;
    } blk_format;
    struct {
        OP *retop;
        SV *old_namesv;
        SV *cur_text;
        CV *cv;
        OP *old_eval_root;
    } blk_eval;
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
#define CXp_MULTICALL 0x10
#define CXp_HASARGS 0x20
#define CXp_SUB_RE 0x40
#define CXp_SUB_RE_FAKE 0x80
#define CXp_REAL 0x20
#define CXp_EVALBLOCK 0x40
#define CXp_TRY 0x80
#define CXp_TRYBLOCK CXp_EVALBLOCK
#define CXp_FOR_DEF 0x10
#define CXp_FOR_LVREF 0x20
#define CXp_FOR_GV 0x40
#define CXp_FOR_PAD 0x80
#define CXp_ONCE 0x10
#define CXp_FINALLY 0x20
#define CxTYPE(cx) ((cx)->cx_type & CXTYPEMASK)
#define CxTYPE_is_LOOP(cx) \
    ({ GOPERL_ONCE(goperl_CxTYPE_is_LOOP_cx_, cx); \
       CxTYPE(goperl_CxTYPE_is_LOOP_cx_) >= CXt_LOOP_ARY && CxTYPE(goperl_CxTYPE_is_LOOP_cx_) <= CXt_LOOP_PLAIN; })
#define CxMULTICALL(cx) (((cx)->cx_type & CXp_MULTICALL) != 0)
#define CxHASARGS(cx) (((cx)->cx_type & CXp_HASARGS) != 0)
/* The lvalue flags (OPpLVAL_INTRO, OPpENTERSUB_INARGS) a sub was called
 * with, kept in the low byte of blk_u16 (cop.h). */
#define CxLVAL(cx) (0 + ((U8)((cx)->blk_u16)))
/* What an eval context keeps in blk_u16 (cop.h): the PL_in_eval it saved,
 * whether it holds a reference on cur_text, and the type of its eval op. */
#define CxOLD_IN_EVAL(cx) (((cx)->blk_u16) & 0x3F)
#define CxEVAL_TXT_REFCNTED(cx) (((cx)->blk_u16) & 0x40)
#define CxOLD_OP_TYPE(cx) (((cx)->blk_u16) >> 7)
#define CxREALEVAL(cx) \
    (((cx)->cx_type & (CXTYPEMASK | CXp_REAL)) == (CXt_EVAL | CXp_REAL))
#define CxEVALBLOCK(cx) \
    (((cx)->cx_type & (CXTYPEMASK | CXp_EVALBLOCK)) == (CXt_EVAL | CXp_EVALBLOCK))
#define CxTRYBLOCK(cx) CxEVALBLOCK(cx)
#define CxTRY(cx) \
    (((cx)->cx_type & (CXTYPEMASK | CXp_TRY)) == (CXt_EVAL | CXp_TRY))
#define CxFOREACH(cx) \
    ({ GOPERL_ONCE(goperl_CxFOREACH_cx_, cx); \
       CxTYPE(goperl_CxFOREACH_cx_) >= CXt_LOOP_ARY && CxTYPE(goperl_CxFOREACH_cx_) <= CXt_LOOP_LIST; })

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
    goperl_pio_views_v = 0; /* they were in the arena */
    GOPERL_SHARED->rx_bodies = 0; /* so were they */
    /* and so may the chain SvMAGIC showed be (its copies of the
     * interpreter's magic): the slot was written back before */
    GOPERL_SHARED->svmagic_sv = 0;
    GOPERL_SHARED->svmagic_val = 0;
    GOPERL_SHARED->svmagic_orig = 0;
}

/* ---- entries of hashes (see the HE/HEK typedefs near the top) ---------- */

#ifndef HVhek_UTF8
#define HVhek_UTF8 0x01
#endif

/* Whether p points into the interpreter's memory as it is now (it never
 * shrinks, and never moves). */
static inline __attribute__((unused)) int goperl_in_memory(const goperl_frame_t *f,
                                                           const void *p) {
    return f->membase && (uintptr_t)p - (uintptr_t)f->membase < *f->memsize;
}
/* The entry of the interpreter's at an address of its memory, where it
 * lies (NULL for 0). */
static inline __attribute__((unused)) HE *goperl_he_at(goperl_frame_t *f, uint64_t g) {
    return (HE *)goperl_guest_mem(f, g);
}
/* A pointer an entry holds (hent_next, hent_hek), as a pointer of this
 * process. An entry outside the interpreter's memory is a module's own and
 * holds pointers of this process. One inside it holds addresses of the
 * memory, except where a module stored a pointer of this process there:
 * Cpanel::JSON::XS makes the entries of a tied hash in a string buffer and
 * points them at keys in the same buffer. The number alone tells the two
 * apart: every address of the memory is below the memory's ceiling, and
 * the memory lies at or above its ceiling in this process (the loader
 * loads no native code into an instance whose memory does not: see
 * memoryAboveCeiling in xs.go), so an address of the memory is never also
 * a pointer into it. Any other number is an address, and goperl_guest_mem
 * fails the call when the memory does not hold it. */
static inline __attribute__((unused)) void *goperl_he_ptr(goperl_frame_t *f, const HE *he,
                                                          uint64_t v) {
    if (!v || !goperl_in_memory(f, he) || goperl_in_memory(f, (const void *)(uintptr_t)v))
        return (void *)(uintptr_t)v;
    return goperl_guest_mem(f, v);
}
static inline __attribute__((unused)) HEK *goperl_he_hek(goperl_frame_t *f, const HE *he) {
    return (HEK *)goperl_he_ptr(f, he, (uint64_t)(uintptr_t)he->hent_hek);
}
static inline __attribute__((unused)) HE *goperl_he_next(goperl_frame_t *f, const HE *he) {
    return (HE *)goperl_he_ptr(f, he, he->hent_next.goperl_addr);
}
/* Whether the interpreter can read the entry itself: one of its memory
 * whose key is an address of its memory (an entry of the interpreter's,
 * not one a module made with a key of this process's). */
static inline __attribute__((unused)) int goperl_he_readable(const goperl_frame_t *f,
                                                             const HE *he) {
    return goperl_in_memory(f, he) && !goperl_in_memory(f, (const void *)he->hent_hek);
}
/* The address of an entry of the interpreter's memory, for the
 * interpreter to keep (HvEITER): an entry of this process's cannot be
 * handed to it. */
static uint64_t goperl_he_addr(goperl_frame_t *f, const HE *he, const char *what)
    __attribute__((unused));
static uint64_t goperl_he_addr(goperl_frame_t *f, const HE *he, const char *what) {
    if (!he) return 0;
    if (!goperl_in_memory(f, he))
        goperl_croakf(f, "%s: an entry that is not in the interpreter's memory", what);
    return (uint64_t)((uintptr_t)he - (uintptr_t)f->membase);
}

#define hv_fetch_ent(hv, keysv, lval, hash)                              \
    goperl_he_at(_gof, goperl_op3(GOPERL_OP_HV_FETCH_ENT, GOPERL_TOK(hv), \
                                  GOPERL_TOK(keysv), (lval) ? 1 : 0))
#define hv_store_ent(hv, keysv, sv, hash)                                 \
    goperl_he_at(_gof, goperl_op3(GOPERL_OP_HV_STORE_ENT, GOPERL_TOK(hv),  \
                                  GOPERL_TOK(keysv), GOPERL_TOK(sv)))
#define hv_iternext(hv) \
    goperl_he_at(_gof, goperl_op0(GOPERL_OP_HV_ITERNEXT, GOPERL_TOK(hv), 0))
#define hv_iternext_sv hv_iternext
/* hv.h's, over goperl_he_hek and goperl_he_next */
#define HeVAL(he) ((he)->he_valu.hent_val)
#define HeNEXT(he) goperl_he_next(_gof, (const HE *)(he))
#define HeKEY_hek(he) goperl_he_hek(_gof, (const HE *)(he))
#define HeKEY(he) HEK_KEY(HeKEY_hek(he))
#define HeKEY_sv(he) (*(SV **)HeKEY(he))
#define HeKLEN(he) HEK_LEN(HeKEY_hek(he))
#define HeHASH(he) HEK_HASH(HeKEY_hek(he))
#define HeKUTF8(he) HEK_UTF8(HeKEY_hek(he))
#define HeKWASUTF8(he) HEK_WASUTF8(HeKEY_hek(he))
#define HeKLEN_UTF8(he)                                                    \
    ({ HEK *const goperl_HeKLEN_UTF8_hek_ = HeKEY_hek(he);                 \
       HEK_UTF8(goperl_HeKLEN_UTF8_hek_) ? -HEK_LEN(goperl_HeKLEN_UTF8_hek_) \
                                         : HEK_LEN(goperl_HeKLEN_UTF8_hek_); })
#define HeUTF8(he)                                                         \
    ({ HEK *const goperl_HeUTF8_hek_ = HeKEY_hek(he);                      \
       (HEK_LEN(goperl_HeUTF8_hek_) == HEf_SVKEY)                          \
           ? (U32)SvUTF8(*(SV **)HEK_KEY(goperl_HeUTF8_hek_))              \
           : (U32)HEK_UTF8(goperl_HeUTF8_hek_); })
#define HePV(he, lp)                                                       \
    ({ HEK *const goperl_HePV_hek_ = HeKEY_hek(he);                        \
       (HEK_LEN(goperl_HePV_hek_) == HEf_SVKEY)                            \
           ? SvPV(*(SV **)HEK_KEY(goperl_HePV_hek_), lp)                   \
           : ((lp = HEK_LEN(goperl_HePV_hek_)), HEK_KEY(goperl_HePV_hek_)); })
#define HeSVKEY(he)                                                        \
    ({ HEK *const goperl_HeSVKEY_hek_ = HeKEY_hek(he);                     \
       (goperl_HeSVKEY_hek_ && HEK_LEN(goperl_HeSVKEY_hek_) == HEf_SVKEY)  \
           ? *(SV **)HEK_KEY(goperl_HeSVKEY_hek_) : (SV *)0; })
#define HeSVKEY_set(he, sv)                                                \
    ({ HEK *const goperl_HeSVKEY_set_hek_ = HeKEY_hek(he);                 \
       HEK_LEN(goperl_HeSVKEY_set_hek_) = HEf_SVKEY;                       \
       *(SV **)HEK_KEY(goperl_HeSVKEY_set_hek_) = (sv); })

/* The key as a new mortal scalar (hv_iterkeysv): the interpreter's own
 * for an entry it can read; for one a module made, of the key it points
 * at (newSVhek). */
static SV *goperl_newSVhek(goperl_frame_t *f, const HEK *hek) __attribute__((unused));
static SV *goperl_hv_iterkeysv(goperl_frame_t *f, const HE *he) __attribute__((unused));
static SV *goperl_hv_iterkeysv(goperl_frame_t *f, const HE *he) {
    if (goperl_he_readable(f, he))
        return GOPERL_SV(goperl_do_op(f, GOPERL_OP_HV_ITERKEYSV,
                                      (uint64_t)((uintptr_t)he - (uintptr_t)f->membase),
                                      0, 0, 0));
    const HEK *const hek = goperl_he_hek(f, he);
    if (!hek) return 0;
    SV *const sv = goperl_newSVhek(f, hek);
    goperl_do_op(f, GOPERL_OP_SV_MORTAL, GOPERL_TOK(sv), 0, 0, 0);
    return sv;
}
#define hv_iterkeysv(he) goperl_hv_iterkeysv(_gof, (const HE *)(he))
#define HeSVKEY_force(he) hv_iterkeysv(he)
/* The value of the entry an iteration is at: the interpreter's for an
 * entry it can read (the value of a tied hash's is fetched); HeVAL of
 * one a module made. */
static SV *goperl_hv_iterval(goperl_frame_t *f, HV *hv, const HE *he) __attribute__((unused));
static SV *goperl_hv_iterval(goperl_frame_t *f, HV *hv, const HE *he) {
    if (goperl_he_readable(f, he))
        return GOPERL_SV(goperl_do_op(f, GOPERL_OP_HV_ITERVAL, GOPERL_TOK(hv),
                                      (uint64_t)((uintptr_t)he - (uintptr_t)f->membase),
                                      0, 0));
    return he->he_valu.hent_val;
}
#define hv_iterval(hv, he) goperl_hv_iterval(_gof, (HV *)(hv), (const HE *)(he))

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
    if (n < 0) n = 0;
    /* the contexts, the one above them, then what each context was when it
     * was read (what native code stores into one goes to the interpreter:
     * goperl_cx_writeback) */
    si->si_cxstack = (PERL_CONTEXT *)goperl_arena_alloc(
        sizeof(PERL_CONTEXT) * (2 * (size_t)n + 1));
    if (!si->si_cxstack) goperl_croakf(f, "context shadow: out of memory");
    for (I32 i = 0; i < n; i++) {
        PERL_CONTEXT *cx = &si->si_cxstack[i];
        /* CX_DUMP: the fields as words (see the interpreter side) */
        const goperl_gword_t *w = (const goperl_gword_t *)goperl_guest_mem(
            f, goperl_do_op3(f, GOPERL_OP_CX, tok, (uint64_t)(int64_t)i, 0 /* GOPERL_CX_DUMP */,
                             0, 0));
        goperl_gword_t c[16];
        if (!w) goperl_croakf(f, "context shadow: no such context");
        memcpy(c, w, sizeof c); /* the next operation reuses the words */
        memset(cx, 0, sizeof(*cx));
        cx->cx_type = (U8)c[0];
        cx->blk_gimme = (U8)c[1];
        cx->blk_u16 = (U16)c[2];
        cx->blk_oldsaveix = (I32)c[3];
        cx->blk_oldsp = (SSize_t)(int64_t)c[4];
        cx->blk_oldmarksp = (I32)c[5];
        cx->blk_oldscopesp = (I32)c[6];
        cx->blk_old_tmpsfloor = (SSize_t)(int64_t)c[9];
        cx->blk_oldcop = goperl_cop_shadow(f, c[7]);
        cx->blk_oldpm = goperl_op_shadow(f, c[8]);
        switch (cx->cx_type & CXTYPEMASK) {
        case CXt_SUB:
            cx->blk_sub.cv = (CV *)GOPERL_SV(c[10]);
            cx->blk_sub.retop = goperl_op_shadow_full(f, c[11]);
            cx->blk_sub.olddepth = (I32)c[12];
            cx->blk_sub.savearray = (AV *)GOPERL_SV(c[13]);
            cx->blk_sub.prevcomppad = (AV *)GOPERL_SV(c[14]);
            break;
        case CXt_FORMAT:
            cx->blk_format.cv = (CV *)GOPERL_SV(c[10]);
            cx->blk_format.retop = goperl_op_shadow_full(f, c[11]);
            /* the sub of a format is asked for as a sub's */
            cx->blk_sub.cv = cx->blk_format.cv;
            break;
        case CXt_EVAL:
            cx->blk_eval.retop = goperl_op_shadow_full(f, c[10]);
            cx->blk_eval.old_namesv = GOPERL_SV(c[11]);
            cx->blk_eval.cur_text = GOPERL_SV(c[12]);
            cx->blk_eval.cv = (CV *)GOPERL_SV(c[13]);
            cx->blk_eval.old_eval_root = goperl_op_shadow(f, c[14]);
            break;
        default:
            cx->blk_loop.my_op = goperl_op_shadow_full(f, c[10]);
            break;
        }
        si->si_cxstack[n + 1 + i] = *cx;
    }
    /* The context above the top is the one perl popped last, which it
     * leaves as it was: code that has just left a scope reads it (the free
     * of Variable::Magic asks whether an eval was being left). It is shown
     * as the interpreter has it, with its type and numbers but none of its
     * ops or scalars, which may be gone; what native code stores into it
     * goes nowhere, as nothing reads it back. */
    {
        PERL_CONTEXT *ghost = &si->si_cxstack[n];
        memset(ghost, 0, sizeof(*ghost));
        const goperl_gword_t *w = (const goperl_gword_t *)goperl_guest_mem(
            f, goperl_do_op3(f, GOPERL_OP_CX, tok, (uint64_t)(int64_t)n, 0 /* GOPERL_CX_DUMP */,
                             0, 0));
        if (w) {
            goperl_gword_t c[10];
            memcpy(c, w, sizeof c);
            ghost->cx_type = (U8)c[0];
            ghost->blk_gimme = (U8)c[1];
            ghost->blk_u16 = (U16)c[2];
            ghost->blk_oldsaveix = (I32)c[3];
            ghost->blk_oldsp = (SSize_t)(int64_t)c[4];
            ghost->blk_oldmarksp = (I32)c[5];
            ghost->blk_oldscopesp = (I32)c[6];
            ghost->blk_old_tmpsfloor = (SSize_t)(int64_t)c[9];
        }
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

/* CX selectors. */
enum { GOPERL_CX_DUMP = 0, GOPERL_CX_DOUNWIND = 1, GOPERL_CX_MAX = 2, GOPERL_CX_SET = 3 };

/* What native code stored into the contexts it was shown goes to the
 * interpreter's contexts, field by field, before the next operation and
 * before the contexts are read again. Native code may store into a context
 * it was shown before an operation (one that reads PL_op, say): the
 * context is the interpreter's still when the interpreter has it as it was
 * shown, and then it is written; one that is not (popped, or another in
 * its place) is left alone. */
static uint64_t goperl_cx_op_tok(goperl_frame_t *f, const OP *o) {
    if (o && !o->gop)
        goperl_croakf(f, "an op stored into a context is not the interpreter's");
    return o ? o->gop : 0;
}
/* A context as the words CX_DUMP gives (fields 0 to 14; a signed field
 * sign-extended to the word, as the interpreter stores it). */
static void goperl_cx_words(const PERL_CONTEXT *cx, goperl_gword_t *w) {
    memset(w, 0, 15 * sizeof *w);
    w[0] = cx->cx_type;
    w[1] = cx->blk_gimme;
    w[2] = cx->blk_u16;
    w[3] = (goperl_gword_t)(int64_t)cx->blk_oldsaveix;
    w[4] = (goperl_gword_t)(int64_t)cx->blk_oldsp;
    w[5] = (goperl_gword_t)(int64_t)cx->blk_oldmarksp;
    w[6] = (goperl_gword_t)(int64_t)cx->blk_oldscopesp;
    w[7] = cx->blk_oldcop ? cx->blk_oldcop->gop : 0;
    w[8] = cx->blk_oldpm ? cx->blk_oldpm->gop : 0;
    w[9] = (goperl_gword_t)(int64_t)cx->blk_old_tmpsfloor;
    switch (cx->cx_type & CXTYPEMASK) {
    case CXt_SUB:
        w[10] = GOPERL_TOK(cx->blk_sub.cv);
        w[11] = cx->blk_sub.retop ? cx->blk_sub.retop->gop : 0;
        w[12] = (goperl_gword_t)(int64_t)cx->blk_sub.olddepth;
        w[13] = GOPERL_TOK(cx->blk_sub.savearray);
        w[14] = GOPERL_TOK(cx->blk_sub.prevcomppad);
        break;
    case CXt_FORMAT:
        w[10] = GOPERL_TOK(cx->blk_format.cv);
        w[11] = cx->blk_format.retop ? cx->blk_format.retop->gop : 0;
        break;
    case CXt_EVAL:
        w[10] = cx->blk_eval.retop ? cx->blk_eval.retop->gop : 0;
        w[11] = GOPERL_TOK(cx->blk_eval.old_namesv);
        w[12] = GOPERL_TOK(cx->blk_eval.cur_text);
        w[13] = GOPERL_TOK(cx->blk_eval.cv);
        w[14] = cx->blk_eval.old_eval_root ? cx->blk_eval.old_eval_root->gop : 0;
        break;
    default:
        if (CxTYPE_is_LOOP(cx))
            w[10] = cx->blk_loop.my_op ? cx->blk_loop.my_op->gop : 0;
        break;
    }
}
/* The interpreter has the context as it was shown. */
static int goperl_cx_unchanged(goperl_frame_t *f, uint64_t si_tok, I32 ix,
                               const PERL_CONTEXT *was) {
    goperl_gword_t want[15];
    const goperl_gword_t *now = (const goperl_gword_t *)goperl_guest_mem(
        f, goperl_xs_opx(f, GOPERL_OP_CX, si_tok, (uint64_t)(int64_t)ix, 0 /* GOPERL_CX_DUMP */,
                         0, 0, 0));
    if (!now) return 0;
    goperl_cx_words(was, want);
    return !memcmp(now, want, sizeof want);
}
static void goperl_cx_set(goperl_frame_t *f, uint64_t si_tok, I32 ix,
                          uint32_t field, uint64_t value) {
    const goperl_gword_t w[3] = {(goperl_gword_t)(int64_t)ix, field, value};
    goperl_xs_opx(f, GOPERL_OP_CX, si_tok, sizeof w, (uint64_t)(GOPERL_CX_SET), (const char *)w,
                  sizeof w, 0);
}
static void goperl_cx_writeback(goperl_frame_t *f) {
    if (!goperl_si_cache_v) return;
    const int fresh = goperl_si_gen_v == goperl_gen_v;
    for (PERL_SI *si = goperl_si_cache_v; si; si = si->si_prev) {
        const I32 n = si->si_cxix + 1;
        for (I32 i = 0; i < n; i++) {
            PERL_CONTEXT *cx = &si->si_cxstack[i], *was = &si->si_cxstack[n + 1 + i];
            if (!memcmp(cx, was, sizeof *cx)) continue;
            if (!fresh && !goperl_cx_unchanged(f, si->tok, i, was)) {
                *cx = *was;
                continue;
            }
#define GOPERL_CX_FIELD(id, member, value)                             \
            if (cx->member != was->member)                             \
                goperl_cx_set(f, si->tok, i, (id), (uint64_t)(value));
            GOPERL_CX_FIELD(0, cx_type, cx->cx_type)
            GOPERL_CX_FIELD(1, blk_gimme, cx->blk_gimme)
            GOPERL_CX_FIELD(2, blk_u16, cx->blk_u16)
            GOPERL_CX_FIELD(3, blk_oldsaveix, cx->blk_oldsaveix)
            GOPERL_CX_FIELD(4, blk_oldsp, cx->blk_oldsp)
            GOPERL_CX_FIELD(5, blk_oldmarksp, cx->blk_oldmarksp)
            GOPERL_CX_FIELD(6, blk_oldscopesp, cx->blk_oldscopesp)
            GOPERL_CX_FIELD(7, blk_oldcop, goperl_cx_op_tok(f, cx->blk_oldcop))
            GOPERL_CX_FIELD(8, blk_oldpm, goperl_cx_op_tok(f, cx->blk_oldpm))
            GOPERL_CX_FIELD(9, blk_old_tmpsfloor, cx->blk_old_tmpsfloor)
            switch (was->cx_type & CXTYPEMASK) {
            case CXt_SUB:
                GOPERL_CX_FIELD(10, blk_sub.cv, GOPERL_TOK(cx->blk_sub.cv))
                GOPERL_CX_FIELD(11, blk_sub.retop, goperl_cx_op_tok(f, cx->blk_sub.retop))
                GOPERL_CX_FIELD(12, blk_sub.olddepth, cx->blk_sub.olddepth)
                GOPERL_CX_FIELD(13, blk_sub.savearray, GOPERL_TOK(cx->blk_sub.savearray))
                GOPERL_CX_FIELD(14, blk_sub.prevcomppad, GOPERL_TOK(cx->blk_sub.prevcomppad))
                break;
            case CXt_FORMAT:
                GOPERL_CX_FIELD(10, blk_format.cv, GOPERL_TOK(cx->blk_format.cv))
                GOPERL_CX_FIELD(11, blk_format.retop, goperl_cx_op_tok(f, cx->blk_format.retop))
                break;
            case CXt_EVAL:
                GOPERL_CX_FIELD(10, blk_eval.retop, goperl_cx_op_tok(f, cx->blk_eval.retop))
                GOPERL_CX_FIELD(11, blk_eval.old_namesv, GOPERL_TOK(cx->blk_eval.old_namesv))
                GOPERL_CX_FIELD(12, blk_eval.cur_text, GOPERL_TOK(cx->blk_eval.cur_text))
                GOPERL_CX_FIELD(13, blk_eval.cv, GOPERL_TOK(cx->blk_eval.cv))
                GOPERL_CX_FIELD(14, blk_eval.old_eval_root, goperl_cx_op_tok(f, cx->blk_eval.old_eval_root))
                break;
            default:
                if (CxTYPE_is_LOOP(was))
                    GOPERL_CX_FIELD(10, blk_loop.my_op, goperl_cx_op_tok(f, cx->blk_loop.my_op))
                break;
            }
#undef GOPERL_CX_FIELD
            *was = *cx;
        }
    }
}

/* caller_cx: the interpreter finds the frame (skipping the debugger's own
 * and crossing stackinfos, as caller does); the answer is the shadow of
 * that context. */
static const PERL_CONTEXT *goperl_caller_cx(goperl_frame_t *f, I32 level,
                                            const PERL_CONTEXT **dbcxp)
    __attribute__((unused));
static const PERL_CONTEXT *goperl_caller_cx(goperl_frame_t *f, I32 level,
                                            const PERL_CONTEXT **dbcxp) {
    /* -> the stackinfo of the frame, 0 for none; second: its index + 1 */
    uint64_t found[2], index1[2] = {0, 0};
    found[0] = goperl_do_opx(f, GOPERL_OP_CALLER_CX, (uint64_t)(int64_t)level, 0, 0, 0, 0,
                             &index1[0]);
    found[1] = dbcxp ? goperl_do_opx(f, GOPERL_OP_CALLER_CX, (uint64_t)(int64_t)level, 1, 0,
                                     0, 0, &index1[1])
                     : 0;
    const PERL_CONTEXT *out[2] = {0, 0};
    for (int i = 0; i < 2; i++) {
        if (!found[i]) continue;
        for (PERL_SI *si = goperl_si_cur(f); si; si = si->si_prev)
            if (si->tok == found[i]) {
                I32 ix = (I32)index1[i] - 1;
                if (ix >= 0 && ix <= si->si_cxix) out[i] = &si->si_cxstack[ix];
                break;
            }
    }
    if (dbcxp) *dbcxp = out[1];
    return out[0];
}
#define caller_cx(level, dbcxp) goperl_caller_cx(_gof, (level), (dbcxp))
#define cxstack (goperl_si_cur(_gof)->si_cxstack)
#define cxstack_ix (goperl_si_cur(_gof)->si_cxix)

/* ---- MAGIC --------------------------------------------------------------- */

static void goperl_svmagic_changed(goperl_frame_t *f, uint64_t sv);
static MAGIC *goperl_sv_magicext(goperl_frame_t *f, SV *sv, SV *obj, int how,
                                 const MGVTBL *vtbl, const char *name, I32 namlen) {
    goperl_flush(f);
    MAGIC *const mg = goperl_api_magic_ext(f, GOPERL_TOK(sv), GOPERL_TOK(obj), (int32_t)how,
                                           (const void *)vtbl, name, (int64_t)namlen);
    goperl_svmagic_changed(f, GOPERL_TOK(sv));
    return mg;
}
#define sv_magicext(sv, obj, how, vtbl, name, namlen)                                  \
    goperl_sv_magicext(_gof, (SV *)(sv), (SV *)(obj), (int)(how), (const MGVTBL *)(vtbl), \
                       (const char *)(name), (I32)(namlen))
/* SvMAGIC is WRITABLE (DBI splices tie magic by assigning the chain
 * head): reads load a synced slot; a changed slot is written back through
 * magic_set_head at the next flush (before the next operation, and when the
 * native call ends).
 *
 * The chain it shows is the SV's whole chain: copies of the MAGIC the
 * interpreter itself keeps there (a tie, the UTF-8 cache, ...), as
 * goperl_mg_find gives one, then the chain native code attached, which is
 * the loader's. Clone walks it to copy a value's magic, ties among it. A
 * type native code's chain has (sv_magic attaches the interpreter's magic
 * and a mirror of it) is shown once, as the mirror. A copy is good until
 * the outermost native call returns; its vtable is interp_magic_vtbl (all
 * zero), which is how the write-back and SvMAGIC_set leave the copies out
 * of the chain native code keeps. */
#define goperl_svmagic_sv_v (GOPERL_SHARED->svmagic_sv)
#define goperl_svmagic_val_v (GOPERL_SHARED->svmagic_val)
#define goperl_svmagic_orig_v (GOPERL_SHARED->svmagic_orig)
/* The SDK itself changed the chain of sv: an SvMAGIC view of it is done
 * with, as a write-back would put the chain the view showed back over the
 * change (Variable::Magic casts with sv_magicext and reads SvMAGIC after);
 * a chain native code took off whole stays native code's, with the change
 * in front (see goperl_svmagic_detached). */
static void goperl_svmagic_changed(goperl_frame_t *f, uint64_t sv) {
    if (goperl_svmagic_sv_v != sv) return;
    if (goperl_svmagic_detached(f))
        goperl_svmagic_val_v = goperl_api_magic_chain(f, sv);
    else
        goperl_svmagic_sv_v = 0;
}
#define GOPERL_CALL_MG_LIST 26
static int goperl_magic_is_interp_copy(goperl_frame_t *f, const MAGIC *mg) {
    return mg && mg->mg_virtual == &((goperl_shared_t *)f->inst)->interp_magic_vtbl;
}
/* The chain native code keeps of one SvMAGIC shows: the nodes that are not
 * copies of the interpreter's magic, in their order. A node of native
 * code's linked to a copy (a MAGIC put before the head of the chain
 * SvMAGIC showed) is linked past the copies instead, so that native code's
 * chain holds none of them. */
static MAGIC *goperl_magic_chain_of_native(goperl_frame_t *f, MAGIC *mg) {
    MAGIC *head = 0;
    MAGIC **link = &head;
    for (; mg; mg = mg->mg_moremagic) {
        if (goperl_magic_is_interp_copy(f, mg)) continue;
        if (*link != mg) *link = mg;
        link = &mg->mg_moremagic;
    }
    if (*link) *link = 0;
    return head;
}
static MAGIC **goperl_svmagic_lv(goperl_frame_t *f, SV *sv) {
    /* a chain native code took off whole is native code's until it puts
     * it back (see goperl_svmagic_detached) */
    if (goperl_svmagic_sv_v == GOPERL_TOK(sv) && goperl_svmagic_detached(f)) {
        goperl_flush(f);
        return &goperl_svmagic_val_v;
    }
    goperl_flush(f);
    MAGIC *const native = goperl_api_magic_chain(f, GOPERL_TOK(sv));
    MAGIC *head = native;
    /* CALL 26: MG_LIST */
    const goperl_gword_t *w = (const goperl_gword_t *)goperl_guest_mem(
        f, goperl_do_op3(f, GOPERL_OP_CALL, GOPERL_TOK(sv), 0, (uint64_t)(GOPERL_CALL_MG_LIST), 0, 0));
    if (w && w[0] > 0) {
        const uint64_t all = w[0];
        uint64_t n = 0;
        /* a zeroed MAGIC before the copies and one after them bound them
         * (see goperl_svmagic_writeback) */
        MAGIC *copies = (MAGIC *)goperl_arena_alloc((size_t)(all + 2) * sizeof(MAGIC));
        if (!copies) goperl_croakf(f, "Out of memory!");
        copies++;
        for (uint64_t i = 0; i < all; i++) {
            goperl_gword_t c[6];
            memcpy(c, w + 1 + i * 6, sizeof c);
            /* sv_magic of native code attached it and mirrored it */
            int mirrored = 0;
            for (const MAGIC *m = native; m && !mirrored; m = m->mg_moremagic)
                mirrored = m->mg_type == (char)c[0];
            if (mirrored) continue;
            MAGIC *mg = &copies[n++];
            mg->mg_virtual = &((goperl_shared_t *)f->inst)->interp_magic_vtbl;
            mg->mg_type = (char)c[0];
            mg->mg_flags = (U8)c[1];
            mg->mg_len = (I32)c[2];
            mg->mg_obj = GOPERL_SV(c[3]);
            mg->mg_private = (U16)c[5];
            if (mg->mg_len == HEf_SVKEY)
                mg->mg_ptr = (char *)GOPERL_SV(c[4]); /* the name is an SV */
            else
                mg->mg_ptr = c[4] ? (char *)goperl_guest_mem(f, c[4]) : 0;
        }
        for (uint64_t i = 0; i < n; i++)
            copies[i].mg_moremagic = i + 1 < n ? &copies[i + 1] : native;
        copies[n].mg_moremagic = native; /* the head behind the copies */
        if (n) head = copies;
    }
    goperl_svmagic_val_v = head;
    goperl_svmagic_orig_v = head;
    goperl_svmagic_sv_v = GOPERL_TOK(sv);
    return &goperl_svmagic_val_v;
}
#define SvMAGIC(sv) (*goperl_svmagic_lv(_gof, (SV *)(sv)))
/* The magic of a type on an SV: what native code attached, or else what
 * the interpreter itself keeps there (the back references of a weakened
 * referent, the overload table of a stash, the taint of a value, ...). The
 * latter is a copy, good until the outermost native call returns; its
 * mg_ptr is the interpreter's memory, except for the overload table, which
 * is given as the AMT it is (see goperl_amt). */
static void *goperl_arena_alloc(size_t n);
static void *goperl_amt(goperl_frame_t *f, SV *stash);
static MAGIC *goperl_mg_find(goperl_frame_t *f, const SV *sv, int type) {
    if (!sv) return 0;
    goperl_flush(f);
    for (MAGIC *mg = goperl_api_magic_chain(f, GOPERL_TOK(sv)); mg;
         mg = mg->mg_moremagic)
        if (mg->mg_type == (char)type) return mg;
    /* CALL 8: MG_FIND */
    const goperl_gword_t *w = (const goperl_gword_t *)goperl_guest_mem(
        f, goperl_do_op3(f, GOPERL_OP_CALL, GOPERL_TOK(sv), (uint64_t)(unsigned char)type, (uint64_t)(8), 0, 0));
    if (!w) return 0;
    goperl_gword_t c[6];
    memcpy(c, w, sizeof c);
    MAGIC *mg = (MAGIC *)goperl_arena_alloc(sizeof(MAGIC));
    if (!mg) goperl_croakf(f, "Out of memory!");
    mg->mg_type = (char)c[0];
    mg->mg_flags = (U8)c[1];
    mg->mg_len = (I32)c[2];
    mg->mg_obj = GOPERL_SV(c[3]);
    mg->mg_private = (U16)c[5];
    if (mg->mg_len == HEf_SVKEY)
        mg->mg_ptr = (char *)GOPERL_SV(c[4]); /* the name is an SV */
    else if (type == 'c') /* PERL_MAGIC_overload_table */
        mg->mg_ptr = (char *)goperl_amt(f, (SV *)sv);
    else
        mg->mg_ptr = c[4] ? (char *)goperl_guest_mem(f, c[4]) : 0;
    return mg;
}
#define mg_find(sv, type) goperl_mg_find(_gof, (sv), (type))
#define sv_unmagic(sv, how) \
    ({ GOPERL_ONCE(goperl_sv_unmagic_sv_, sv); GOPERL_ONCE(goperl_sv_unmagic_how_, how); \
       (goperl_flush(_gof), goperl_api_magic_del(_gof, GOPERL_TOK(goperl_sv_unmagic_sv_), (int32_t)(goperl_sv_unmagic_how_), 0), goperl_svmagic_changed(_gof, GOPERL_TOK(goperl_sv_unmagic_sv_)), (void)goperl_op0(GOPERL_OP_SV_UNMAGIC, GOPERL_TOK(goperl_sv_unmagic_sv_), (uint64_t)(uint32_t)(goperl_sv_unmagic_how_)), 0); })
#define sv_unmagicext(sv, how, vtbl)                                   \
    ({ GOPERL_ONCE(goperl_sv_unmagicext_sv_, sv);                      \
       (goperl_flush(_gof),                                            \
        goperl_api_magic_del(_gof, GOPERL_TOK(goperl_sv_unmagicext_sv_), (int32_t)(how), \
                             (const void *)(vtbl)),                    \
        goperl_svmagic_changed(_gof, GOPERL_TOK(goperl_sv_unmagicext_sv_)), 0); })
/* Replace the mirror chain's head (the SvMAGIC_set idiom for unlinking an
 * entry the module found via SvMAGIC). The unlinked node's cleanup stays
 * with the module (Safefree), exactly as with real perl. */
static void goperl_magic_unlinked(goperl_frame_t *f, int type) __attribute__((noreturn));
static void goperl_magic_unlinked(goperl_frame_t *f, int type) {
    goperl_croakf(f,
                  "SvMAGIC: native code unlinked magic '%c' that the interpreter keeps on "
                  "the value from its chain; gperl cannot take the interpreter's own "
                  "magic off a value that way (sv_unmagic can)",
                  type);
}
/* The chain mg starts, as the head of sv's: every type of the
 * interpreter's magic the chain SvMAGIC showed of sv has is still on it (a
 * copy of it), as the write-back of the slot asks of an assignment to it. */
static void goperl_svmagic_set(goperl_frame_t *f, SV *sv, MAGIC *mg) {
    goperl_flush(f);
    const goperl_shared_t *const sh = (const goperl_shared_t *)f->inst;
    if (sh && sh->svmagic_sv == GOPERL_TOK(sv))
        for (const MAGIC *c = sh->svmagic_orig; goperl_magic_is_interp_copy(f, c); c++) {
            int on = 0;
            for (const MAGIC *m = mg; m && !on; m = m->mg_moremagic)
                on = goperl_magic_is_interp_copy(f, m) && m->mg_type == c->mg_type;
            if (!on) goperl_magic_unlinked(f, c->mg_type);
        }
    goperl_api_magic_set_head(f, GOPERL_TOK(sv), goperl_magic_chain_of_native(f, mg));
    goperl_svmagic_changed(f, GOPERL_TOK(sv));
}
#define SvMAGIC_set(sv, mg) goperl_svmagic_set(_gof, (SV *)(sv), (MAGIC *)(mg))
#define PERL_MAGIC_backref '<'
/* Plain sv_magic attaches BEHAVIORAL core magic (PERL_MAGIC_tied being
 * the load-bearing case: DBI ties its outer handles) — that must land on
 * the real guest SV, not the host mirror, or the interpreter never
 * dispatches FETCH/STORE. Custom-vtbl magic stays host-side via
 * sv_magicext. */
/* uvar magic: the two functions and the index a module hands to
 * sv_magic(sv, NULL, PERL_MAGIC_uvar, (char *)&uf, ...). The interpreter
 * calls the host for them (see GOPERL_HOST_UVAR); the module's own struct
 * is what the MAGIC's mg_ptr holds here. */
struct ufuncs {
    I32 (*uf_val)(pTHX_ IV, SV *);
    I32 (*uf_set)(pTHX_ IV, SV *);
    IV uf_index;
};
#define PERL_MAGIC_uvar 'U'
/* perl's vtable of uvar magic: the functions of the MAGIC's struct ufuncs
 * with its index. A module that attaches the magic without the struct and
 * fills the MAGIC in afterwards (Tk's variable traces) puts this one in
 * it. */
static int goperl_uvar_get(SV *sv, MAGIC *mg) {
    const struct ufuncs *uf = (const struct ufuncs *)mg->mg_ptr;
    if (uf && uf->uf_val) uf->uf_val(uf->uf_index, sv);
    return 0;
}
static int goperl_uvar_set(SV *sv, MAGIC *mg) {
    const struct ufuncs *uf = (const struct ufuncs *)mg->mg_ptr;
    if (uf && uf->uf_set) uf->uf_set(uf->uf_index, sv);
    return 0;
}
GOPERL_C __attribute__((weak, visibility("hidden"))) MGVTBL PL_vtbl_uvar = {
    goperl_uvar_get, goperl_uvar_set, 0, 0, 0, 0, 0, 0};
static void goperl_safefree(void *p) __attribute__((unused));
static void goperl_sv_magic(goperl_frame_t *f, SV *sv, SV *obj, int how,
                            const char *name, I32 namlen) {
    if (how == PERL_MAGIC_uvar) {
        /* The name is the module's struct ufuncs, copied (namlen is its
         * size) or kept (0). A shorter one is not, nor is none: a module
         * that attaches the magic with some other name, or with none,
         * puts its own vtable or struct in the MAGIC afterwards (SWIG's
         * variables, Tk's traces), and the interpreter then reports both
         * reads and writes, for that vtable to handle. */
        uint64_t calls = 3;
        if (name && (namlen == 0 || namlen >= (I32)sizeof(struct ufuncs))) {
            const struct ufuncs *uf = (const struct ufuncs *)name;
            calls = (uf->uf_val ? 1 : 0) | (uf->uf_set ? 2 : 0);
        }
        /* CALL 11: UVAR; bit 0: uf_val, bit 1: uf_set */
        goperl_do_op3(f, GOPERL_OP_CALL, GOPERL_TOK(sv), calls, (uint64_t)(11), 0, 0);
    } else if (how == PERL_MAGIC_utf8 && name && namlen < 0) {
        /* the UTF-8 cache of positions, which the magic owns from now on
         * and frees (Clone copies one into a block of its own): the
         * interpreter keeps a copy of its own, and the block is freed as
         * the magic would. Nothing is mirrored: mg_find finds the
         * interpreter's. */
        goperl_do_op3(f, GOPERL_OP_SV_MAGIC_STD, GOPERL_TOK(sv), GOPERL_TOK(obj),
                      (uint64_t)(uint32_t)how | ((uint64_t)1 << 32), name,
                      PERL_MAGIC_UTF8_CACHESIZE * 2 * sizeof(STRLEN));
        goperl_safefree((void *)name);
        return;
    } else if (how == PERL_MAGIC_ext) {
        /* extension magic has no behaviour of the interpreter's: the
         * mirror below is all of it. A second MAGIC in the interpreter
         * would be one more to take off when the module takes its own off
         * from inside a hook (pTk's TclObj_set does sv_unmagic of '~'),
         * and the interpreter's mg_set would step onto the freed one. */
    } else if (name && namlen == HEf_SVKEY) {
        /* the name is an SV, which the interpreter's magic holds a
         * reference to: its token, in the bytes of the text */
        const uint64_t tok = GOPERL_TOK((SV *)name);
        goperl_do_op3(f, GOPERL_OP_SV_MAGIC_STD, GOPERL_TOK(sv), GOPERL_TOK(obj),
                      (uint64_t)(uint32_t)how | ((uint64_t)2 << 32), (const char *)&tok,
                      sizeof tok);
    } else {
        /* A name of a length of 0 or below other than those is a pointer
         * the magic keeps as it is, into this process's memory: the
         * interpreter's magic goes without it, and the mirror below, which
         * mg_find finds, keeps it. */
        goperl_do_op3(f, GOPERL_OP_SV_MAGIC_STD, GOPERL_TOK(sv), GOPERL_TOK(obj),
                      (uint64_t)(int64_t)how, name && namlen > 0 ? name : "",
                      name && namlen > 0 ? (uint64_t)namlen : 0);
    }
    /* Mirror it too: XS reads the attachment back with mg_find (DBI
     * resolves its inner handle through the tie's mg_obj). */
    goperl_flush(f);
    goperl_api_magic_ext(f, GOPERL_TOK(sv), GOPERL_TOK(obj), how, 0,
                            name, name ? (int64_t)namlen : 0);
    /* on a chain native code took off whole, SvMAGIC is the new magic in
     * front of what native code had there, as perl's sv_magic leaves it */
    goperl_svmagic_changed(f, GOPERL_TOK(sv));
}
#define sv_magic(sv, obj, how, name, namlen)                          \
    goperl_sv_magic(_gof, (SV *)(sv), (SV *)(obj), (int)(how),        \
                    (const char *)(name), (I32)(namlen))
/* Any-magic test: the anchor for mirrored magic marks the guest SV
 * RMAGICAL, so this covers every SV whose magic the host can see. */

/* ---- v5: the Class::MOP/Moose surface ----------------------------------- */

#define SvREFCNT(sv) \
    ((U32)goperl_op0(GOPERL_OP_SV_REFCNT, GOPERL_TOK(sv), 0))

static HV *goperl_gv_stashsv(goperl_frame_t *f, SV *name, I32 flags) {
    uint64_t len = 0;
    const char *p = goperl_direct_sv_pv(f, GOPERL_TOK(name), &len);
    /* the name is characters when the SV is (perl's gv_stashsv) */
    if (SvUTF8(name)) flags |= SVf_UTF8;
    return (HV *)GOPERL_SV(goperl_do_op(f, GOPERL_OP_GV_STASHPV,
                                        (uint64_t)(int64_t)flags, 0, p, len));
}
#define gv_stashsv(sv, flags) goperl_gv_stashsv(_gof, (SV *)(sv), (I32)(flags))

/* Expand a stash stub/constant slot into a real glob in place. */
#define gv_init_pvn(gv, stash, name, len, flags)                          \
    ((void)goperl_do_op3(_gof, GOPERL_OP_GV_INIT, GOPERL_TOK(gv), GOPERL_TOK(stash), \
                         (U32)(flags), (name), (uint64_t)(len)))
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

/* The mro generations of a stash, read the way Moose's mop.c reads one
 * (HvAUX(stash)->xhv_mro_meta->pkg_gen) and the way DBI reads the other
 * (HvMROMETA(stash)->cache_gen, which says whether a method it looked up
 * and kept is still the one the stash resolves to). Only these chains are
 * modeled: a scratch one-deep view refreshed from the guest on every call. */
struct mro_meta {
    U32 pkg_gen;
    U32 cache_gen;
};
struct xpvhv_aux {
    struct mro_meta *xhv_mro_meta;
    AV *xhv_backreferences;
    I32 xhv_riter;
};
static struct {
    struct mro_meta meta;
    struct xpvhv_aux aux;
} goperl_hvaux_v;
static struct xpvhv_aux *goperl_hvaux(goperl_frame_t *f, HV *hv) {
    uint64_t r = goperl_do_op(f, GOPERL_OP_HV_PKG_GEN, GOPERL_TOK(hv), 0, 0, 0);
    if (r >> 32) {
        goperl_hvaux_v.meta.pkg_gen = (U32)r;
        goperl_hvaux_v.meta.cache_gen = (U32)goperl_do_op(
            f, GOPERL_OP_HV_PKG_GEN, GOPERL_TOK(hv), 1, 0, 0);
        goperl_hvaux_v.aux.xhv_mro_meta = &goperl_hvaux_v.meta;
    } else {
        goperl_hvaux_v.aux.xhv_mro_meta = 0;
    }
    /* FIELD 1 and 2: the back references and the iterator position */
    goperl_hvaux_v.aux.xhv_backreferences = (AV *)GOPERL_SV(goperl_do_op3(f, GOPERL_OP_FIELD, GOPERL_TOK(hv), 0, (uint64_t)(1), 0, 0));
    goperl_hvaux_v.aux.xhv_riter = (I32)(int64_t)goperl_do_op3(f, GOPERL_OP_FIELD, GOPERL_TOK(hv), 0, (uint64_t)(2), 0, 0);
    return &goperl_hvaux_v.aux;
}
#define HvAUX(hv) goperl_hvaux(_gof, (HV *)(hv))
/* HvMROMETA never returns NULL in core: a stash without the meta gets one. */
static struct mro_meta *goperl_hvmrometa(goperl_frame_t *f, HV *hv) {
    uint64_t r = goperl_do_op(f, GOPERL_OP_HV_PKG_GEN, GOPERL_TOK(hv), 1, 0, 0);
    goperl_hvaux_v.meta.pkg_gen = (U32)(r >> 32);
    goperl_hvaux_v.meta.cache_gen = (U32)r;
    return &goperl_hvaux_v.meta;
}
#define HvMROMETA(hv) goperl_hvmrometa(_gof, (HV *)(hv))

/* The hash of a key, as the keys of the interpreter's hashes have it
 * (HeHASH, HEK_HASH, SvSHARED_HASH): a module compares one it computed with
 * those (XS::Parse::Sublike finds named parameters by hash). The
 * interpreter computes it, with its hash function and seed. The hash
 * operations take the key's bytes and compute it themselves, whatever hash
 * a module hands them. */
static U32 goperl_key_hash(goperl_frame_t *f, const char *key, uint64_t len)
    __attribute__((unused)); /* goperl_compiler.h */
#define PERL_HASH(hash, str, len) \
    ((hash) = goperl_key_hash(_gof, (const char *)(str), (uint64_t)(len)))

/* Compiled-regex test: what SvRX finds, a regexp or a reference to one
 * (perl's get_re_arg, which runs get magic). */
#define SvRXOK(sv) (SvRX(sv) != 0)

/* $_ is a slot of the interpreter that holds an SV: DEFSV reads it,
 * DEFSV_set (and an assignment to DEFSV) puts another SV there, and
 * SAVE_DEFSV has the scope put the present one back. None of them copies
 * a value: code that swaps $_ around a callback relies on the SV that was
 * there being untouched. */
#define DEFSV (*goperl_plsv_ref(_gof, GOPERL_PL_DEFSV))

/* Length-taking stash lookup (the guest op copies exactly slen bytes). */
#define gv_stashpvn(name, len, flags)                                     \
    ((HV *)GOPERL_SV(goperl_ops(GOPERL_OP_GV_STASHPV,                     \
                                (uint64_t)(int64_t)(I32)(flags), 0,       \
                                (name), (uint64_t)(len))))
#define gv_stashpvs(name, flags) \
    gv_stashpvn("" name "", sizeof(name) - 1, (flags))

/* Method resolution with ISA walk (returns the method's GV, or NULL). */
#define gv_fetchmethod(stash, name) \
    ({ GOPERL_ONCE(goperl_gv_fetchmethod_name_, name); \
       ((GV *)GOPERL_SV(goperl_ops(GOPERL_OP_GV_FETCHMETHOD, GOPERL_TOK(stash), 1, (goperl_gv_fetchmethod_name_), __builtin_strlen(goperl_gv_fetchmethod_name_)))); })
#define gv_fetchmethod_autoload(stash, name, autoload) \
    ({ GOPERL_ONCE(goperl_gv_fetchmethod_autoload_name_, name); \
       ((GV *)GOPERL_SV(goperl_ops(GOPERL_OP_GV_FETCHMETHOD, GOPERL_TOK(stash), (autoload) ? 1 : 0, (goperl_gv_fetchmethod_autoload_name_), __builtin_strlen(goperl_gv_fetchmethod_autoload_name_)))); })

/* sv_reftype: the "HASH"/"ARRAY"/"Class::Name" string. Interned host
 * copies, stable for the process like every SDK-returned name. */
static const char *goperl_sv_reftype(goperl_frame_t *f, SV *sv, int ob) {
    if (ob) {
        uint64_t stash = goperl_do_op(f, GOPERL_OP_SV_STASH, GOPERL_TOK(sv),
                                      0, 0, 0);
        if (stash) {
            const char *n = goperl_intern_op(f, GOPERL_OP_HV_NAME, stash, 0, 0);
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

/* The interpreter as perl names it (PL_curinterp, PERL_GET_CONTEXT): the
 * instance the current native call is in, a pointer nothing reads through
 * (NULL on a thread no native call of the interpreter's is on). */
static inline __attribute__((unused)) PerlInterpreter *goperl_interp_cur(void) {
    goperl_frame_t *const f = goperl_cur_frame_v;
    return f ? (PerlInterpreter *)f->inst : (PerlInterpreter *)0;
}
#define PERL_GET_CONTEXT goperl_interp_cur()
#define PL_curinterp goperl_interp_cur()

/* MY_CXT: module state, as perl without MULTIPLICITY keeps it: one static
 * struct, and no context argument (pMY_CXT is void). */
#define START_MY_CXT static my_cxt_t my_cxt;
#define MY_CXT my_cxt
#define dMY_CXT dNOOP
#define dMY_CXT_SV dNOOP
#define dMY_CXT_INTERP(interp) dNOOP
#define MY_CXT_INIT memset(&my_cxt, 0, sizeof my_cxt)
#define MY_CXT_INIT_INTERP(interp) MY_CXT_INIT
#define MY_CXT_CLONE MY_CXT_INIT
#define pMY_CXT void
#define pMY_CXT_
#define _pMY_CXT
#define aMY_CXT
#define aMY_CXT_
#define _aMY_CXT
/* MY_CXT_KEY is the module's, as in perl's perl.h, which does not define
 * it: a module tells by it whether it keeps state in MY_CXT (Time::HiRes
 * defines it only where it does, and runs MY_CXT_INIT `#ifdef MY_CXT_KEY`). */

/* HePV is defined with the HE-shadow machinery below. */


/* ---- v6 groundwork: the Cpanel::JSON::XS / HTML::Parser surface -------- */

#ifndef NV_MAX
#define NV_MAX DBL_MAX
#define NV_MIN DBL_MIN
#define NV_DIG DBL_DIG
#endif
/* The interpreter's integer geometry, which every SEMANTIC decision a dist
 * makes follows ("does this number fit an IV?", IVSIZE branches): its IV
 * and UV are 64 bits wide (use64bitint), the same as the C types IV and UV
 * here. An integer crosses as it is in both directions (sv_setiv, newSViv,
 * sv_setref_pv going in; SvIV, SvUV, SvIVX coming out), a host pointer
 * stored in one with a plain cast included. */
#ifndef UV_MAX
#define IVSIZE 8
#define UVSIZE 8
#define UV_MAX ((UV)UINT64_MAX)
#define UV_MIN 0
#define IV_MAX ((IV)INT64_MAX)
#define IV_MIN ((IV)INT64_MIN)
#endif
#define Gconvert(x, n, t, b) snprintf((b), 64, "%.*g", (int)(n), (double)(x))
#define PTR2ul(p) ((unsigned long)GOPERL_PTR_NUMBER(p))
#define safecalloc(n, sz) goperl_block_alloc((size_t)(n) * (size_t)(sz), 1)
#define SVfARG(sv) ((SV *)(sv))
#define SvIOKp(sv) ((GOPERL_SVINFO(sv) & GOPERL_INFO_IOKp) != 0)
#define SvNOKp(sv) ((GOPERL_SVINFO(sv) & GOPERL_INFO_NOKp) != 0)
#define SvIsBOOL(sv) \
    ((int)goperl_op0(GOPERL_OP_SV_IS_BOOL, GOPERL_TOK(sv), 0))
/* nomg variants: the SDK's accessors go through the guest's public API,
 * which applies get-magic — acceptable for these dists' uses. */
#define SvNVX(sv) \
    (*(NV *)goperl_lv_ref(_gof, (SV *)(sv), GOPERL_LV_NV, GOPERL_LV_AT))
#define SvSetMagicSV(dst, src) \
    ({ GOPERL_ONCE(goperl_SvSetMagicSV_dst_, dst); \
       (sv_setsv((goperl_SvSetMagicSV_dst_), (src)), (void)mg_set(goperl_SvSetMagicSV_dst_)); })
#define SvSetSV(dst, src) sv_setsv((dst), (src))
/* Taint support is off in the embedded build (like -DNO_TAINT_SUPPORT). */
#define TAINT_IF(c) ((void)((c) && PL_tainting && (PL_tainted = 1)))
#define SvTRUEx(sv) SvTRUE(sv)
#define HvAMAGIC(hv) ((int)goperl_op0(GOPERL_OP_GV_AMG, GOPERL_TOK(hv), 0))

/* hv_common: every way into a hash by one function, as hv.h's macros use
 * it. What it returns depends on the action, as in perl: whether the key
 * exists; the value that was deleted; and for a fetch or a store either
 * the entry (HE *: hv_fetch_ent, hv_store_ent) or, with HV_FETCH_JUST_SV,
 * the place of the value in the hash (SV **: hv_fetch, hv_store). */
#define HV_FETCH_ISSTORE 0x04
#define HV_FETCH_ISEXISTS 0x08
#define HV_FETCH_LVALUE 0x10
#define HV_FETCH_JUST_SV 0x20
#define HV_DELETE 0x40
static I32 goperl_hv_exists(goperl_frame_t *f, HV *hv, const char *key,
                            I32 klen);
static I32 goperl_hv_exists_flags(goperl_frame_t *f, HV *hv, const char *key, I32 klen,
                                  int action);
static void *goperl_hv_common(goperl_frame_t *f, HV *hv, SV *keysv,
                              const char *key, STRLEN klen, int kflags,
                              int action, SV *val, U32 hash) {
    (void)hash;
    if (keysv && !key) {
        uint64_t l = 0;
        key = goperl_direct_sv_pv(f, GOPERL_TOK(keysv), &l);
        klen = (STRLEN)l;
        if (goperl_do_op(f, GOPERL_OP_SV_INFO, GOPERL_TOK(keysv), 0, 0, 0) &
            GOPERL_INFO_UTF8)
            kflags |= HVhek_UTF8;
    }
    /* a key's length is negative for the functions below when the key is
     * UTF-8 */
    const I32 signed_klen = (kflags & HVhek_UTF8) ? -(I32)klen : (I32)klen;
    if (action & HV_DELETE)
        return (void *)GOPERL_SV(goperl_do_op3(
            f, GOPERL_OP_HV_DELETE, GOPERL_TOK(hv), (uint64_t)(uint32_t)(action & ~HV_DELETE),
            (uint64_t)(int64_t)signed_klen, key, (uint64_t)klen));
    if (action & HV_FETCH_ISEXISTS)
        return goperl_hv_exists_flags(f, hv, key, signed_klen, action) ? (void *)hv : 0;
    if (action & HV_FETCH_JUST_SV) {
        if (action & HV_FETCH_ISSTORE)
            return (void *)goperl_hv_store(f, hv, key, signed_klen, val);
        return (void *)goperl_hv_fetch(f, hv, key, signed_klen,
                                       (action & HV_FETCH_LVALUE) ? 1 : 0);
    }
    /* the entry: by a key that is a scalar */
    if (!keysv) {
        keysv = GOPERL_SV(goperl_do_op(f, GOPERL_OP_NEW_PVN, 0, klen, key, klen));
        goperl_do_op(f, GOPERL_OP_SV_MORTAL, GOPERL_TOK(keysv), 0, 0, 0);
        if (kflags & HVhek_UTF8)
            goperl_do_op(f, GOPERL_OP_SV_UTF8_ON, GOPERL_TOK(keysv), 0, 0, 0);
    }
    if (action & HV_FETCH_ISSTORE)
        return (void *)goperl_he_at(
            f, goperl_do_op3(f, GOPERL_OP_HV_STORE_ENT, GOPERL_TOK(hv), GOPERL_TOK(keysv),
                             GOPERL_TOK(val), 0, 0));
    /* HV_FETCH_EMPTY_HE: an entry the fetch makes has no value yet, as
     * List::Util's uniqnum tells a new key by */
    return (void *)goperl_he_at(
        f, goperl_do_op3(f, GOPERL_OP_HV_FETCH_ENT, GOPERL_TOK(hv), GOPERL_TOK(keysv),
                         ((action & HV_FETCH_LVALUE) ? 1 : 0) |
                             ((action & HV_FETCH_LVALUE) && (action & HV_FETCH_EMPTY_HE) ? 2 : 0),
                         0, 0));
}
#define hv_common(hv, keysv, key, klen, kflags, action, val, hash)        \
    goperl_hv_common(_gof, (HV *)(hv), (SV *)(keysv), (key),              \
                     (STRLEN)(klen), (int)(kflags), (int)(action),        \
                     (SV *)(val), (U32)(hash))
/* the long names code calls them by (JavaBin: Perl_hv_common(aTHX_ ...));
 * aTHX_ is nothing here */
#define Perl_hv_common(...) hv_common(__VA_ARGS__)
#define Perl_hv_common_key_len(...) hv_common_key_len(__VA_ARGS__)

/* xsubpp emits this for ATTRS: sections. It is `use attributes PACKAGE,
 * \\&code, ATTRS` (op.c): attributes->import(PACKAGE, \\&code, ATTRS). */
static void apply_attrs_string(const char *stashpv, CV *cv,
                               const char *attrstr, STRLEN len) {
    goperl_frame_t *f = goperl_cur_frame_v;
    goperl_do_op(f, GOPERL_OP_EVAL_PV, 1, 0, "require attributes",
                 sizeof("require attributes") - 1);
    goperl_gword_t toks[16];
    int n = 0;
    size_t pl = __builtin_strlen(stashpv);
    uint64_t classsv = goperl_do_op(f, GOPERL_OP_NEW_PVN, 0,
                                    sizeof("attributes") - 1, "attributes",
                                    sizeof("attributes") - 1);
    goperl_do_op(f, GOPERL_OP_SV_MORTAL, classsv, 0, 0, 0);
    toks[n++] = classsv;
    uint64_t pkgsv = goperl_do_op(f, GOPERL_OP_NEW_PVN, 0, pl, stashpv, pl);
    goperl_do_op(f, GOPERL_OP_SV_MORTAL, pkgsv, 0, 0, 0);
    toks[n++] = pkgsv;
    uint64_t rv =
        goperl_do_op(f, GOPERL_OP_NEW_RV_INC, GOPERL_TOK(cv), 0, 0, 0);
    goperl_do_op(f, GOPERL_OP_SV_MORTAL, rv, 0, 0, 0);
    toks[n++] = rv;
    const char *p = attrstr, *e = attrstr + (len ? len : __builtin_strlen(attrstr));
    while (p < e && n < 15) {
        while (p < e && *p == ' ') p++;
        const char *q = p;
        while (q < e && *q != ' ') q++;
        if (q > p) {
            uint64_t s2 = goperl_do_op(f, GOPERL_OP_NEW_PVN, 0,
                                       (uint64_t)(q - p), p, (uint64_t)(q - p));
            goperl_do_op(f, GOPERL_OP_SV_MORTAL, s2, 0, 0, 0);
            toks[n++] = s2;
        }
        p = q;
    }
    uint64_t cvtok = goperl_do_op(f, GOPERL_OP_GET_CV, 0, 0,
                                  "attributes::import",
                                  sizeof("attributes::import") - 1);
    if (!cvtok) return;
    goperl_do_op3(f, GOPERL_OP_CALL_SV, cvtok, (uint64_t)(n * 8),
                  (uint64_t)(uint32_t)(G_VOID | G_DISCARD), (const char *)toks,
                  (uint64_t)(n * 8));
}

/* warnings-state stubs: cop_warnings is modeled as an opaque slot dists
 * assign pWARN_* into around host-local code; the guest's real warnings
 * state is reached through ckWARN/Perl_warner instead. */
#define pWARN_NONE ((char *)0)
#define pWARN_STD ((char *)0)
#define pWARN_ALL ((char *)1)
#define WARN_UTF8 44 /* perl 5.44 warnings.h (categories the SDK uses sit below the 5.44 renumbering) */
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
#define get_cvs(name, flags) get_cv("" name "", (flags))
/* The flags of pv_uni_display and sv_uni_display, as utf8.h has them. */
#define UNI_DISPLAY_ISPRINT 0x0001
#define UNI_DISPLAY_BACKSLASH 0x0002
#define UNI_DISPLAY_BACKSPACE 0x0004
#define UNI_DISPLAY_QQ (UNI_DISPLAY_ISPRINT | UNI_DISPLAY_BACKSLASH | UNI_DISPLAY_BACKSPACE)
#define UNI_DISPLAY_REGEX (UNI_DISPLAY_ISPRINT | UNI_DISPLAY_BACKSLASH)
static char *goperl_sv_uni_display(goperl_frame_t *f, SV *dsv, SV *ssv,
                                   STRLEN pvlim, UV flags);
/* pv_uni_display: the interpreter's, over a copy of the bytes (a string of
 * the interpreter's of their length, which sv_uni_display shows). */
static char *goperl_pv_uni_display(goperl_frame_t *f, SV *dsv, const U8 *spv,
                                   STRLEN len, STRLEN pvlim, U32 flags) {
    SV *const tmp = GOPERL_SV(goperl_direct_sv_mortal(
        f, GOPERL_TOK(goperl_newSVpvn(f, (const char *)spv, len))));
    return goperl_sv_uni_display(f, dsv, tmp, pvlim, flags);
}
#define pv_uni_display(dsv, spv, len, pvlim, flags)                     \
    goperl_pv_uni_display(_gof, (SV *)(dsv), (const U8 *)(spv),         \
                          (STRLEN)(len), (STRLEN)(pvlim), (U32)(flags))

/* ---- the DBI surface ---------------------------------------------------- */

/* CopFILEGV: the glob "_<file" of the file the COP was compiled from, whose
 * scalar holds the file name. */
#define CopFILEGV(cop) gv_fetchfile(CopFILE(cop))
#define DEFSV_set(sv) ((void)(DEFSV = (SV *)(sv)))
#define die Perl_croak_nocontext
/* The old name of require_pv: the argument is the name of a file, as the
 * one require takes in a string ("Math/Prime/Util/PP.pm"). */
#define perl_require_pv(pv) require_pv(pv)
/* Host-local save of plain C variables: restored at XSUB return via the
 * host save-stack (same mechanism as SAVEVPTR). */
#define SAVEI32(i) goperl_save_i32(_gof, (int32_t *)&(i))
#define SAVEINT(i) goperl_save_i32(_gof, (int32_t *)&(i))
#define SAVEI8(i) goperl_hostsave_small(_gof, &(i), 1)
#define SAVEI16(i) goperl_hostsave_small(_gof, &(i), 2)
#define SAVEBOOL(i) goperl_hostsave_small(_gof, &(i), 1)
#define DEBUG_l(x) ((void)0)
#define DEBUG_x(x) ((void)0)
#define DEBUG_v_TEST 0
#define aTHXo aTHX
#define aTHXo_ aTHX_
#define pTHXo pTHX
#define pTHXo_ pTHX_
/* MSPAGAIN and MEXTEND: the stack may have moved (a callback or the
 * EXTEND itself made it grow), so the mark is taken again as well. */
#define MSPAGAIN             \
    STMT_START {             \
        sp = PL_stack_sp;    \
        mark = ORIGMARK;     \
    }                        \
    STMT_END
#define MEXTEND(p, n)                                          \
    STMT_START {                                               \
        const SSize_t goperl_markoff = mark - PL_stack_base;   \
        EXTEND(p, n);                                          \
        mark = PL_stack_base + goperl_markoff;                 \
    }                                                          \
    STMT_END
typedef void (*XSUBADDR_t)(pTHX_ CV *cv);
/* SAVE_DEFSV: the interpreter saves the slot (SAVEGENERICSV of it). */
static void goperl_save_defsv(goperl_frame_t *f) {
    goperl_do_op(f, GOPERL_OP_SAVE_MISC, 8 /* GOPERL_SAVE_DEFSV */, 0, 0,
                 0);
}
#define SAVE_DEFSV goperl_save_defsv(_gof)
#define SvIOK_UV(sv) \
    ((GOPERL_SVINFO(sv) & (GOPERL_INFO_IOK | GOPERL_INFO_ISUV)) == \
     (GOPERL_INFO_IOK | GOPERL_INFO_ISUV))
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
                               (uint64_t)SVt_PVAV, name, __builtin_strlen(name));
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
#define dTHR dNOOP
/* The taint checks are the interpreter's: it dies (or, under -t, warns)
 * when tainted data is in use, naming the operation. */
#define TAINT_ENV() \
    ((void)(PL_tainting && (GOPERL_SVCALL(GOPERL_SVCALL_TAINT_ENV, 0, 0), 0)))
#define TAINT_PROPER(s) \
    ({ GOPERL_ONCE(goperl_taint_proper_s_, s); \
       (void)(PL_tainting && (GOPERL_SVSTR(GOPERL_SVSTR_TAINT_PROPER, 0, 0, (goperl_taint_proper_s_), __builtin_strlen(goperl_taint_proper_s_)), 0)); })
#define I_STDARG 1
#define POPi ((IV)SvIV(POPs))
#define POPu ((UV)SvUV(POPs))
#define POPn ((NV)SvNV(POPs))
#define POPl POPi
#define POPpx SvPV_nolen(POPs)
/* Writable taint state (DBI assigns PL_tainted around trace calls); the
 * embedded build has no taint support, so the value is inert. */
#undef PL_tainted
#undef PL_tainting
#undef PL_dirty
/* Assignable no-op: teardown depth is the embedder's business. */
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
/* croak_sv dies with the SV itself: an exception object reaches the Perl
 * code that catches it as the same object. The value is handed to the
 * interpreter here and the frame is marked as ending with a thrown value;
 * the loader then makes the interpreter die with it. The text goes into
 * the frame as well, for the catch points that only report a message. */
#define GOPERL_FAILED_THROWN 2
static void goperl_croak_sv(goperl_frame_t *f, SV *sv) {
    goperl_flush(f);
    /* The value is taken before anything else happens to it. A reference
     * is not stringified at all, as perl does not when it dies with one:
     * its string overloading is the program's code, which may run anything
     * (and clear $@, the very value being thrown, on the way). */
    if (f->jb)
        goperl_xs_opx(f, GOPERL_OP_SV_CALL, GOPERL_TOK(sv), 0, (uint64_t)(50 /* GOPERL_SVCALL_THROW */), 0, 0, 0);
    if (goperl_xs_op(f, GOPERL_OP_SV_INFO, GOPERL_TOK(sv), 0, 0, 0) &
        GOPERL_INFO_ROK) {
        snprintf(f->err, sizeof f->err, "died with a reference");
    } else {
        uint64_t len = 0;
        const char *p = goperl_api_v->sv_pv(f, GOPERL_TOK(sv), &len);
        snprintf(f->err, sizeof f->err, "%.*s", (int)(len > 400 ? 400 : len),
                 p ? p : "");
    }
    if (!f->jb) goperl_raise(f);
    f->failed = GOPERL_FAILED_THROWN;
    goperl_jump_out(f);
}
/* croak(NULL): the exception is whatever $@ holds. */
static void goperl_croak_errsv(goperl_frame_t *f) {
    goperl_croak_sv(f, GOPERL_SV(goperl_xs_op(f, GOPERL_OP_ERRSV, 0, 0,
                                                    0, 0)));
}
#define croak_sv(sv) goperl_croak_sv(_gof, (SV *)(sv))
#define warn_sv(sv) Perl_warn(aTHX_ "%" SVf, SVfARG((SV *)(sv)))
static void goperl_gv_efullname(goperl_frame_t *f, SV *sv, GV *gv) {
    uint64_t stash = goperl_do_op(f, GOPERL_OP_GV_PTR, GOPERL_TOK(gv), 0, 0, 0);
    const char *pkg = stash ? goperl_intern_op(f, GOPERL_OP_HV_NAME, stash, 0, 0) : 0;
    const char *name = goperl_intern_op(f, GOPERL_OP_GV_NAME, GOPERL_TOK(gv), 0, 0);
    char buf[512];
    snprintf(buf, sizeof buf, "%s::%s", pkg && *pkg ? pkg : "main",
             name ? name : "?");
    goperl_do_op(f, GOPERL_OP_SV_SETPVN, GOPERL_TOK(sv), __builtin_strlen(buf), buf,
                 __builtin_strlen(buf));
}
#define gv_efullname(sv, gv) goperl_gv_efullname(_gof, (SV *)(sv), (GV *)(gv))
/* The key of an entry (hv.c's hv_iterkey): its bytes where the entry's key
 * has them, or a key that is a scalar's string. */
static char *goperl_hv_iterkey(goperl_frame_t *f, const HE *he, I32 *retlen) {
    HEK *const hek = goperl_he_hek(f, he);
    if (hek->hek_len == HEf_SVKEY) {
        SV *keysv;
        memcpy(&keysv, hek->hek_key, sizeof keysv);
        uint64_t len = 0;
        char *const p = (char *)goperl_direct_sv_pv(f, GOPERL_TOK(keysv), &len);
        *retlen = (I32)len;
        return p;
    }
    *retlen = hek->hek_len;
    return hek->hek_key;
}
#define hv_iterkey(he, retlen) goperl_hv_iterkey(_gof, (const HE *)(he), (retlen))
#define sv_2io(sv) \
    ((IO *)GOPERL_SV(goperl_op0(GOPERL_OP_SV_2IO, GOPERL_TOK(sv), 0)))
#define sv_force_normal(sv) \
    ((void)goperl_op0(GOPERL_OP_SV_FORCE_NORMAL, GOPERL_TOK(sv), 0))
#define sv_force_normal_flags(sv, flags) sv_force_normal(sv)

/* The handles of an IO, lvalues (`IoIFP(io) = PerlIO_fdopen(fd, "r")`):
 * the plain store of sv.h, which neither opens nor closes anything. */
#define IoIFP(io) \
    (*(PerlIO **)goperl_lv_ref(_gof, (SV *)(io), GOPERL_LV_IOIFP, GOPERL_LV_AT))
#define IoOFP(io) \
    (*(PerlIO **)goperl_lv_ref(_gof, (SV *)(io), GOPERL_LV_IOOFP, GOPERL_LV_AT))

/* ---- stdio streams of handles (PerlIO_findFILE, PerlIO_exportFILE) ---------
 * As perl does, the interpreter flushes the handle, opens a stream on its
 * descriptor and pushes a :stdio layer with that stream on the handle (the
 * stream of a :stdio layer the handle has already is the one found). The
 * stream native code gets is one of this process on the same descriptor,
 * when the descriptor of that number here is the interpreter's (the same
 * device and inode: so it is for a program that is the host process):
 * opened on a duplicate of it, so that giving the stream up closes only
 * the duplicate, and it shares the file's offset with the handle. Both
 * streams are unbuffered, so that what the C code and the program write
 * reaches the descriptor in the order they write it, as through perl's one
 * stream.
 *
 * Where the descriptor is not this process's (an interpreter with
 * descriptors of its own), or the handle has none (an in-memory file), the
 * stream is one whose reads, writes and seeks are the handle's, done in the
 * interpreter: funopen where there is one, fopencookie elsewhere. It has no
 * descriptor (fileno gives -1). The interpreter may die inside one of its
 * callbacks; the C library is on the stack then, so the callback reports
 * failure to it instead of unwinding, and the loader reports the death
 * when the XSUB returns. */
typedef struct goperl_piofile {
    struct goperl_piofile *next;
    uint64_t tok; /* the handle */
    FILE *fp;
    int found;        /* in the list PerlIO_findFILE answers from */
    int own;          /* fp is on a descriptor (closing it does not free this) */
    goperl_gword_t gstream; /* the interpreter's stream of the handle's :stdio layer */
} goperl_piofile_t;
#define goperl_pio_files_v (*(goperl_piofile_t **)&GOPERL_SHARED->pio_files)
static uint64_t goperl_piofile_op(goperl_piofile_t *c, int32_t op, uint64_t a,
                                  uint64_t b, uint64_t sel, const char *s, uint64_t slen,
                                  uint64_t *second, int *failed) {
    goperl_frame_t *f = goperl_cur_frame_v;
    const int32_t failed_before = f->failed;
    goperl_flush(f);
    uint64_t r = goperl_xs_opx(f, op, a, b, sel, s, slen, second);
    goperl_gen_bump();
    *failed = !failed_before && f->failed;
    (void)c;
    return r;
}
static int goperl_piofile_read(void *cookie, char *buf, int n) {
    goperl_piofile_t *c = (goperl_piofile_t *)cookie;
    int failed, total = 0;
    while (total < n) {
        int want = n - total;
        if (want > GOPERL_PIO_READ_MAX) want = GOPERL_PIO_READ_MAX;
        uint64_t n = 0;
        const uint64_t r = goperl_piofile_op(c, GOPERL_OP_PERLIO, c->tok, (uint64_t)want,
                                             GOPERL_PIO_READ, 0, 0, &n, &failed);
        const int32_t got = (int32_t)(int64_t)n;
        if (failed || got < 0) return total ? total : -1;
        if (got > 0)
            memcpy(buf + total, goperl_guest_mem(goperl_cur_frame_v, r), (size_t)got);
        total += got;
        if (got < want) break;
    }
    return total;
}
static int goperl_piofile_write(void *cookie, const char *buf, int n) {
    goperl_piofile_t *c = (goperl_piofile_t *)cookie;
    int failed;
    int64_t r = (int64_t)goperl_piofile_op(c, GOPERL_OP_PERLIO_WRITE, c->tok,
                                           (uint64_t)n, 0, buf, (uint64_t)n, 0,
                                           &failed);
    return (failed || r < 0) ? -1 : (int)r;
}
static int64_t goperl_piofile_seek(void *cookie, int64_t off, int whence) {
    goperl_piofile_t *c = (goperl_piofile_t *)cookie;
    int failed;
    int sel = whence == SEEK_SET   ? GOPERL_PIO_SEEK_SET
              : whence == SEEK_CUR ? GOPERL_PIO_SEEK_CUR
              : whence == SEEK_END ? GOPERL_PIO_SEEK_END
                                   : -1;
    if (sel < 0) return -1;
    int64_t r = (int64_t)goperl_piofile_op(c, GOPERL_OP_PERLIO, c->tok, (uint64_t)off,
                                           (uint64_t)sel, 0, 0, 0, &failed);
    if (failed || r < 0) return -1;
    r = (int64_t)goperl_piofile_op(c, GOPERL_OP_PERLIO, c->tok, 0, GOPERL_PIO_TELL, 0, 0, 0,
                                   &failed);
    return failed ? -1 : r;
}
/* Closing the stream gives it up; the handle stays open. */
static int goperl_piofile_close(void *cookie) {
    free(cookie);
    return 0;
}
#if defined(__APPLE__)
static fpos_t goperl_piofile_seek_cb(void *c, fpos_t off, int whence) {
    return (fpos_t)goperl_piofile_seek(c, (int64_t)off, whence);
}
#else
static ssize_t goperl_piofile_read_cb(void *c, char *buf, size_t n) {
    return goperl_piofile_read(c, buf, n > INT_MAX ? INT_MAX : (int)n);
}
static ssize_t goperl_piofile_write_cb(void *c, const char *buf, size_t n) {
    /* fopencookie takes 0 for a failed write */
    int r = goperl_piofile_write(c, buf, n > INT_MAX ? INT_MAX : (int)n);
    return r < 0 ? 0 : r;
}
static int goperl_piofile_seek_cb(void *c, off64_t *off, int whence) {
    int64_t r = goperl_piofile_seek(c, (int64_t)*off, whence);
    if (r < 0) return -1;
    *off = (off64_t)r;
    return 0;
}
#endif
/* The interpreter's stream of the handle (FIND_FILE or EXPORT_FILE), in
 * *gstream (0 when there is none); the stream of a descriptor here that is
 * the same file, or 0. */
/* Whether the descriptor fd of this process is the one the words of
 * FIND_FILE/EXPORT_FILE describe: the same file at the same offset (or
 * with none, as a pipe). */
static int goperl_pio_same_fd(const goperl_gword_t *w, int fd) {
    /* [descriptor][stream][st_dev][st_ino][offset] */
    const uint64_t dev = w[2];
    const uint64_t ino = w[3];
    const int64_t off = (int64_t)w[4];
    struct stat st;
    return (dev || ino) && fstat(fd, &st) == 0 && (uint64_t)st.st_dev == dev &&
           (uint64_t)st.st_ino == ino && (int64_t)lseek(fd, 0, SEEK_CUR) == off;
}
static FILE *goperl_pio_stdio(goperl_frame_t *f, const void *h, int sel,
                              const char *mode, goperl_gword_t *gstream) {
    const char *m = mode ? mode : "";
    const goperl_gword_t *w = (const goperl_gword_t *)goperl_guest_mem(
        f, goperl_pio(f, sel, h, 0, m, __builtin_strlen(m) + 1));
    *gstream = 0;
    if (!w || (int32_t)w[0] < 0) return 0;
    *gstream = w[1];
    if (!goperl_pio_same_fd(w, (int)w[0])) return 0;
    const int fd = dup((int)w[0]);
    if (fd < 0) return 0;
    if (!*m) {
        /* the mode the descriptor was opened with, as perl's
         * PerlIO_modestr answers for the handle */
        const int fl = fcntl(fd, F_GETFL);
        const int acc = fl & O_ACCMODE;
        m = acc == O_RDONLY   ? "r"
            : acc == O_WRONLY ? ((fl & O_APPEND) ? "a" : "w")
                              : ((fl & O_APPEND) ? "a+" : "r+");
    }
    FILE *fp = fdopen(fd, m);
    if (!fp) {
        close(fd);
        return 0;
    }
    setvbuf(fp, 0, _IONBF, 0);
    return fp;
}
static FILE *goperl_pio_export(goperl_frame_t *f, const void *h,
                               const char *mode, int found)
    __attribute__((unused));
static FILE *goperl_pio_export(goperl_frame_t *f, const void *h,
                               const char *mode, int found) {
    if (!h || !goperl_pio(f, GOPERL_PIO_VALID, h, 0, 0, 0)) return 0;
    goperl_piofile_t *c = (goperl_piofile_t *)calloc(1, sizeof *c);
    if (!c) return 0;
    c->tok = (uint64_t)(uintptr_t)h;
    c->fp = goperl_pio_stdio(f, h, found ? GOPERL_PIO_FIND_FILE : GOPERL_PIO_EXPORT_FILE,
                             found ? 0 : mode, &c->gstream);
    if (c->fp) {
        c->own = 1;
        if (found) {
            c->found = 1;
            c->next = goperl_pio_files_v;
            goperl_pio_files_v = c;
        }
        return c->fp;
    }
#if defined(__APPLE__)
    (void)mode;
    c->fp = funopen(c, goperl_piofile_read, goperl_piofile_write,
                    goperl_piofile_seek_cb, goperl_piofile_close);
#else
    cookie_io_functions_t io;
    io.read = goperl_piofile_read_cb;
    io.write = goperl_piofile_write_cb;
    io.seek = goperl_piofile_seek_cb;
    io.close = goperl_piofile_close;
    c->fp = fopencookie(c, (mode && *mode) ? mode : "r+", io);
#endif
    if (!c->fp) {
        free(c);
        return 0;
    }
    setvbuf(c->fp, 0, _IONBF, 0);
    if (found) {
        c->found = 1;
        c->next = goperl_pio_files_v;
        goperl_pio_files_v = c;
    }
    return c->fp;
}
/* Closes a stream this code made, and the record of it. */
static void goperl_piofile_free(goperl_piofile_t *c) {
    const int own = c->own;
    fclose(c->fp); /* frees c when it is a stream over the handle */
    if (own) free(c);
}
/* PerlIO_findFILE: the stream of the handle's :stdio layer, made on first
 * use. A stream made before stands for the handle as long as the handle
 * still has the :stdio layer it was made with and, for a stream on a
 * descriptor, that descriptor is still the handle's (a handle closed and
 * another opened at its address is not the same handle, even when its
 * stream happens to be where the old one was). */
static FILE *goperl_pio_find(goperl_frame_t *f, const void *h)
    __attribute__((unused));
static FILE *goperl_pio_find(goperl_frame_t *f, const void *h) {
    uint64_t tok = (uint64_t)(uintptr_t)h;
    for (goperl_piofile_t **pp = &goperl_pio_files_v; *pp; pp = &(*pp)->next) {
        goperl_piofile_t *c = *pp;
        if (c->tok != tok) continue;
        if (!goperl_pio(f, GOPERL_PIO_VALID, h, 0, 0, 0)) return 0;
        if (c->gstream) {
            const goperl_gword_t *w = (const goperl_gword_t *)goperl_guest_mem(
                f, goperl_pio(f, GOPERL_PIO_FIND_FILE, h, 0, "", 1));
            if (!w || (int32_t)w[0] < 0 || w[1] != c->gstream ||
                (c->own && !goperl_pio_same_fd(w, fileno(c->fp)))) {
                *pp = c->next;
                goperl_piofile_free(c);
                break;
            }
        }
        clearerr(c->fp);
        return c->fp;
    }
    return goperl_pio_export(f, h, "r+", 1);
}
/* PerlIO_releaseFILE: the stream is given up, and the :stdio layer that
 * came with it is popped (the handle goes on with the layers below). */
static void goperl_pio_release(goperl_frame_t *f, const void *h, FILE *fp)
    __attribute__((unused));
static void goperl_pio_release(goperl_frame_t *f, const void *h, FILE *fp) {
    uint64_t tok = (uint64_t)(uintptr_t)h;
    for (goperl_piofile_t **pp = &goperl_pio_files_v; *pp; pp = &(*pp)->next) {
        goperl_piofile_t *c = *pp;
        if (c->tok == tok && c->fp == fp) {
            *pp = c->next;
            if (c->gstream)
                goperl_pio(f, GOPERL_PIO_RELEASE_FILE, h, c->gstream, 0, 0);
            goperl_piofile_free(c);
            return;
        }
    }
    if (fp) fclose(fp);
}
/* The handle is closed (PerlIO_close): the stream PerlIO_findFILE made for
 * it goes; the handle's :stdio layer is closed with the handle. */
static void goperl_pio_forget(goperl_frame_t *f, const void *h) {
    uint64_t tok = (uint64_t)(uintptr_t)h;
    (void)f;
    for (goperl_piofile_t **pp = &goperl_pio_files_v; *pp; pp = &(*pp)->next) {
        goperl_piofile_t *c = *pp;
        if (c->tok == tok) {
            *pp = c->next;
            goperl_piofile_free(c);
            return;
        }
    }
}
#define PerlIO_exportFILE(io, mode) goperl_pio_export(_gof, (io), (mode), 0)
#define PerlIO_findFILE(io) goperl_pio_find(_gof, (io))
#define PerlIO_releaseFILE(io, fp) goperl_pio_release(_gof, (io), (fp))

static PerlIO *goperl_perlio_open(goperl_frame_t *f, const char *path,
                                  const char *mode) {
    char buf[1024];
    size_t ml = __builtin_strlen(mode), pl = __builtin_strlen(path);
    if (ml + pl + 2 > sizeof buf) return 0;
    memcpy(buf, mode, ml + 1);
    memcpy(buf + ml + 1, path, pl + 1);
    PerlIO *const io = (PerlIO *)(uintptr_t)goperl_do_op(
        f, GOPERL_OP_PERLIO_OPEN, 0, 0, buf, ml + 1 + pl + 1);
    if (!io) goperl_errno_from_interp(f);
    return io;
}
static int goperl_pio_close(goperl_frame_t *f, const void *h)
    __attribute__((unused));
static int goperl_pio_close(goperl_frame_t *f, const void *h) {
    goperl_pio_forget(f, h);
    const int r = (int)(int64_t)goperl_do_op(
        f, GOPERL_OP_PERLIO_CLOSE, (uint64_t)(uintptr_t)h, 0, 0, 0);
    if (r != 0) goperl_errno_from_interp(f);
    return r;
}
#define PerlIO_open(path, mode) goperl_perlio_open(_gof, (path), (mode))
#define PerlIO_close(io) goperl_pio_close(_gof, (io))
/* PerlIO_fdopen: a handle on a descriptor. For a program that is the host
 * process the interpreter's descriptors are the process's, so the
 * descriptor can be one this code made (a socket, a pipe, a terminal): the
 * handle reads and writes it, and closing the handle closes it. */
static PerlIO *goperl_pio_fdopen(goperl_frame_t *f, int fd, const char *mode)
    __attribute__((unused));
static PerlIO *goperl_pio_fdopen(goperl_frame_t *f, int fd, const char *mode) {
    if (!mode) mode = "r";
    PerlIO *const io = (PerlIO *)(uintptr_t)goperl_pio(
        f, GOPERL_PIO_FDOPEN, 0, (uint64_t)(int64_t)fd, mode,
        __builtin_strlen(mode) + 1);
    if (!io) goperl_errno_from_interp(f);
    return io;
}
#define PerlIO_fdopen(fd, mode) goperl_pio_fdopen(_gof, (fd), (mode))
/* PerlIO_importFILE: a handle on the descriptor of a stdio stream, open
 * for what mode says, or for what the descriptor is open for when mode is
 * null or empty. What the stream has buffered for output is written first;
 * from then on the handle is the way to the descriptor, and closing it
 * closes the descriptor (the stream is not to be used afterwards, as with
 * perl). */
static PerlIO *goperl_pio_import(goperl_frame_t *f, FILE *fp, const char *mode)
    __attribute__((unused));
static PerlIO *goperl_pio_import(goperl_frame_t *f, FILE *fp, const char *mode) {
    const int fd = fp ? fileno(fp) : -1;
    if (fd < 0) {
        errno = EBADF;
        return 0;
    }
    if (!mode || !*mode) {
        const int flags = fcntl(fd, F_GETFL);
        if (flags < 0) return 0;
        switch (flags & O_ACCMODE) {
        case O_WRONLY: mode = (flags & O_APPEND) ? "a" : "w"; break;
        case O_RDWR: mode = (flags & O_APPEND) ? "a+" : "r+"; break;
        default: mode = "r"; break;
        }
    }
    fflush(fp);
    return goperl_pio_fdopen(f, fd, mode);
}
#define PerlIO_importFILE(fp, mode) goperl_pio_import(_gof, (fp), (mode))

/* eval_sv: results delivered like call_sv (the guest packs them into a
 * mortal AV; the SDK pushes them onto the host stack model). */
static I32 goperl_eval_sv(goperl_frame_t *f, SV *sv, I32 flags) {
    /* the code run reads the ops as native code left them */
    goperl_optree_flush(f);
    /* -> the mortal array of the results; second: died<<63 | count */
    uint64_t second = 0;
    const uint64_t avtok = goperl_do_opx(f, GOPERL_OP_EVAL_SV, GOPERL_TOK(sv), 0,
                                         (uint64_t)(uint32_t)flags, 0, 0, &second);
    int died = (int)(second >> 63);
    I32 count = (I32)(second & 0x7FFFFFFF);
    goperl_stack_extend(f, f->psp, count);
    for (I32 i = 0; i < count; i++) {
        uint64_t tok = goperl_xs_op(f, GOPERL_OP_AV_FETCH, avtok,
                                           (uint64_t)(int64_t)i, 0, 0);
        *++f->psp = (SV *)(uintptr_t)tok;
    }
    /* eval_sv traps the error in $@; only G_RETHROW asks for it to be
     * thrown again (perl.c) */
    if (died && (flags & G_RETHROW)) goperl_croak_errsv(f);
    return count;
}
#define eval_sv(sv, flags) goperl_eval_sv(_gof, (SV *)(sv), (I32)(flags))


/* ASCII ctype (real perl's are ASCII-range too for these dists' uses). */
#ifndef toLOWER
#define toLOWER(c) \
    ({ GOPERL_ONCE(goperl_toLOWER_c_, c); \
       goperl_toLOWER_c_ >= 'A' && goperl_toLOWER_c_ <= 'Z' ? goperl_toLOWER_c_ + 32 : goperl_toLOWER_c_; })
#define toUPPER(c) \
    ({ GOPERL_ONCE(goperl_toUPPER_c_, c); \
       goperl_toUPPER_c_ >= 'a' && goperl_toUPPER_c_ <= 'z' ? goperl_toUPPER_c_ - 32 : goperl_toUPPER_c_; })
#endif
#ifndef isALPHA
#define isALPHA(c) \
    ({ GOPERL_ONCE(goperl_isALPHA_c_, c); \
       (goperl_isALPHA_c_ >= 'a' && goperl_isALPHA_c_ <= 'z') || (goperl_isALPHA_c_ >= 'A' && goperl_isALPHA_c_ <= 'Z'); })
#define isDIGIT(c) \
    ({ GOPERL_ONCE(goperl_isDIGIT_c_, c); \
       goperl_isDIGIT_c_ >= '0' && goperl_isDIGIT_c_ <= '9'; })
#define isALNUM(c) \
    ({ GOPERL_ONCE(goperl_isALNUM_c_, c); \
       isALPHA(goperl_isALNUM_c_) || isDIGIT(goperl_isALNUM_c_) || goperl_isALNUM_c_ == '_'; })
#define isSPACE(c) \
    ({ GOPERL_ONCE(goperl_isSPACE_c_, c); \
       goperl_isSPACE_c_ == ' ' || goperl_isSPACE_c_ == '\t' || goperl_isSPACE_c_ == '\n' || goperl_isSPACE_c_ == '\r' || goperl_isSPACE_c_ == '\f' || goperl_isSPACE_c_ == '\v'; })
#define isWORDCHAR(c) isALNUM(c)
#define isPRINT(c) \
    ({ GOPERL_ONCE(goperl_isPRINT_c_, c); \
       goperl_isPRINT_c_ >= 0x20 && goperl_isPRINT_c_ < 0x7f; })
#define isGRAPH(c) \
    ({ GOPERL_ONCE(goperl_isGRAPH_c_, c); \
       goperl_isGRAPH_c_ > 0x20 && goperl_isGRAPH_c_ < 0x7f; })
#define isUPPER(c) \
    ({ GOPERL_ONCE(goperl_isUPPER_c_, c); \
       goperl_isUPPER_c_ >= 'A' && goperl_isUPPER_c_ <= 'Z'; })
#define isLOWER(c) \
    ({ GOPERL_ONCE(goperl_isLOWER_c_, c); \
       goperl_isLOWER_c_ >= 'a' && goperl_isLOWER_c_ <= 'z'; })
#define isPUNCT(c) \
    ({ GOPERL_ONCE(goperl_isPUNCT_c_, c); \
       isGRAPH(goperl_isPUNCT_c_) && !isALNUM(goperl_isPUNCT_c_); })
#define isXDIGIT(c) \
    ({ GOPERL_ONCE(goperl_isXDIGIT_c_, c); \
       isDIGIT(goperl_isXDIGIT_c_) || (goperl_isXDIGIT_c_ >= 'a' && goperl_isXDIGIT_c_ <= 'f') || (goperl_isXDIGIT_c_ >= 'A' && goperl_isXDIGIT_c_ <= 'F'); })
#define isCNTRL(c) \
    ({ GOPERL_ONCE(goperl_isCNTRL_c_, c); \
       ((U8)goperl_isCNTRL_c_) < 0x20 || goperl_isCNTRL_c_ == 0x7f; })
#define isBLANK(c) \
    ({ GOPERL_ONCE(goperl_isBLANK_c_, c); \
       goperl_isBLANK_c_ == ' ' || goperl_isBLANK_c_ == '\t'; })
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
/* An overlong sequence (more bytes than its character needs) is malformed
 * unless one of these allows it; the first makes it U+FFFD, the second
 * gives the character it spells. */
#define UTF8_ALLOW_LONG 0x10
#define UTF8_ALLOW_LONG_AND_ITS_VALUE (UTF8_ALLOW_LONG | 0x20)
#define UTF8_DISALLOW_SURROGATE 0x8
#define UTF8_WARN_SURROGATE 0
#define UTF8_WARN_SUPER 0
#define UTF8_DISALLOW_ILLEGAL_INTERCHANGE \
    (UTF8_DISALLOW_SURROGATE | UTF8_DISALLOW_SUPER)
/* what Unicode's Corrigendum #9 allows to be interchanged: the
 * noncharacters too, which these flags never refuse */
#define UTF8_DISALLOW_ILLEGAL_C9_INTERCHANGE \
    (UTF8_DISALLOW_SURROGATE | UTF8_DISALLOW_SUPER)
#define UTF8_WARN_ILLEGAL_C9_INTERCHANGE (UTF8_WARN_SURROGATE | UTF8_WARN_SUPER)
static const U8 goperl_utf8skip_v[16] = {1, 1, 1, 1, 1, 1, 1, 1,
                                         1, 1, 1, 1, 2, 2, 3, 4};
/* The length a start byte announces. Above U+1FFFFF perl's extended
 * UTF-8 goes on: five bytes (F8-FB), six (FC, FD), seven (FE), thirteen
 * (FF). */
static inline __attribute__((unused)) STRLEN goperl_utf8skip(U8 c) {
    if (c < 0xF8) return goperl_utf8skip_v[c >> 4];
    return c < 0xFC ? 5 : c < 0xFE ? 6 : c == 0xFE ? 7 : 13;
}
#define UTF8SKIP(s) goperl_utf8skip(*(const U8 *)(s))
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
        /* the skip table gives 1 for a continuation byte too; only an
         * invariant is a character by itself */
        if (s[0] & 0x80) goto fail;
        if (retlen) *retlen = 1;
        return s[0];
    }
    {
        UV uv = s[0] & (0x7f >> n);
        for (STRLEN i = 1; i < n; i++) {
            if (!UTF8_IS_CONTINUATION(s[i])) goto fail;
            uv = (uv << 6) | (s[i] & 0x3f);
        }
        /* non-minimal encodings are malformed, as in the real decoder,
         * unless the caller allows them */
        if ((n == 2 && uv < 0x80) || (n == 3 && uv < 0x800) ||
            (n == 4 && uv < 0x10000) || (n == 5 && uv < 0x200000) ||
            (n == 6 && uv < 0x4000000) || (n >= 7 && uv < 0x80000000)) {
            if (!(flags & UTF8_ALLOW_LONG)) goto fail;
            if ((flags & UTF8_ALLOW_LONG_AND_ITS_VALUE) != UTF8_ALLOW_LONG_AND_ITS_VALUE)
                uv = 0xFFFD;
        }
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
    ({ GOPERL_ONCE(goperl_utf8_to_uvchr_buf_s_, s); \
       utf8n_to_uvchr((goperl_utf8_to_uvchr_buf_s_), (STRLEN)((const U8 *)(e) - (const U8 *)(goperl_utf8_to_uvchr_buf_s_)), (retlen), 0); })
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
    } else if (uv < 0x200000) {
        *d++ = (U8)(0xf0 | (uv >> 18));
        *d++ = (U8)(0x80 | ((uv >> 12) & 0x3f));
        *d++ = (U8)(0x80 | ((uv >> 6) & 0x3f));
        *d++ = (U8)(0x80 | (uv & 0x3f));
    } else {
        /* perl's extended UTF-8: the start byte says how many bytes, and
         * each of the others carries six bits */
        const int n = uv < 0x4000000 ? 5 : uv < 0x80000000 ? 6
                      : uv < ((UV)1 << 36) ? 7 : 13;
        *d++ = n == 5 ? (U8)(0xf8 | (uv >> 24))
               : n == 6 ? (U8)(0xfc | (uv >> 30))
               : n == 7 ? (U8)0xfe : (U8)0xff;
        for (int i = n - 2; i >= 0; i--)
            *d++ = (U8)(0x80 | (i * 6 < 64 ? ((uv >> (i * 6)) & 0x3f) : 0));
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
#define SvLEN(sv) \
    (*(STRLEN *)goperl_lv_ref(_gof, (SV *)(sv), GOPERL_LV_LEN, GOPERL_LV_AT))
#define SvREADONLY_off(sv) GOPERL_FLAGS_CHANGE((sv), 0, SVf_READONLY)
static char *goperl_sv_pv_force(goperl_frame_t *f, SV *sv, STRLEN *lenp) {
    return goperl_str_op(f, GOPERL_OP_SV_PV_FORCE, GOPERL_TOK(sv), 0, 0, lenp);
}
#define SvPV_force(sv, lenvar) \
    GOPERL_PV_LEN(lenvar, goperl_sv_pv_force(_gof, (SV *)(sv), (STRLEN *)&goperl_pvlen_))
#define sv_chop(sv, p) \
    ({ GOPERL_ONCE(goperl_sv_chop_sv_, sv); \
       ((void)goperl_op0(GOPERL_OP_SV_CHOP, GOPERL_TOK(goperl_sv_chop_sv_), (uint64_t)((const char *)(p)-SvPVX_const(goperl_sv_chop_sv_)))); })
#define sv_insert(big, off, len, little, littlelen)                      \
    ((void)goperl_do_op3(_gof, GOPERL_OP_SV_INSERT, GOPERL_TOK(big), (uint64_t)(off), \
                         (uint64_t)(len), (little), (uint64_t)(littlelen)))
#define sv_utf8_decode(sv) \
    ((bool)goperl_op0(GOPERL_OP_SV_UTF8_DECODE, GOPERL_TOK(sv), 0))
#define sv_utf8_downgrade(sv, fail_ok) \
    ((bool)goperl_op0(GOPERL_OP_SV_UTF8_DOWNGRADE, GOPERL_TOK(sv), \
                      (fail_ok) ? 1 : 0))
#define mg_set(sv) ((void)goperl_op0(GOPERL_OP_MG_SET, GOPERL_TOK(sv), 0), 0)
#define SvSETMAGIC(sv) ((void)mg_set(sv))
#define sv_setiv_mg(sv, i) \
    ({ GOPERL_ONCE(goperl_sv_setiv_mg_sv_, sv); \
       (sv_setiv((goperl_sv_setiv_mg_sv_), (i)), (void)mg_set(goperl_sv_setiv_mg_sv_)); })
#define sv_setnv_mg(sv, n) \
    ({ GOPERL_ONCE(goperl_sv_setnv_mg_sv_, sv); \
       (sv_setnv((goperl_sv_setnv_mg_sv_), (n)), (void)mg_set(goperl_sv_setnv_mg_sv_)); })
#define sv_setpvn_mg(sv, p, n) \
    ({ GOPERL_ONCE(goperl_sv_setpvn_mg_sv_, sv); \
       (sv_setpvn((goperl_sv_setpvn_mg_sv_), (p), (n)), (void)mg_set(goperl_sv_setpvn_mg_sv_)); })
#define sv_setsv_mg_2(sv, s2) \
    ({ GOPERL_ONCE(goperl_sv_setsv_mg_2_sv_, sv); \
       (sv_setsv((goperl_sv_setsv_mg_2_sv_), (s2)), (void)mg_set(goperl_sv_setsv_mg_2_sv_)); })
#define GIMME_V ((U8)goperl_op0(GOPERL_OP_GIMME_V, 0, 0))
#define GIMME GIMME_V
#define ckWARN(w) ((int)goperl_op0(GOPERL_OP_CKWARN, (uint64_t)(w), 0))
#define ckWARN_d(w) ((int)goperl_op0(GOPERL_OP_CKWARN, (uint64_t)(w), 1))
#define ckWARN2(a, b) ckWARN(packWARN2(a, b))
#define ckWARN3(a, b, c) ckWARN(packWARN3(a, b, c))
#define ckWARN4(a, b, c, d) ckWARN(packWARN4(a, b, c, d))
#define ckWARN2_d(a, b) ckWARN_d(packWARN2(a, b))
#define ckWARN3_d(a, b, c) ckWARN_d(packWARN3(a, b, c))
#define ckWARN4_d(a, b, c, d) ckWARN_d(packWARN4(a, b, c, d))
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
/* The overload magic of perls before 5.18, as ppport.h defines it for the
 * modules that still name it (Package::Stash::XS puts it on a %OVERLOAD it
 * makes): the interpreter has no such magic and refuses it, as perl does
 * ("Don't know how to handle magic of type \\101"). */
#define PERL_MAGIC_overload 'A'
#define av_pop(av) \
    GOPERL_SV(goperl_op0(GOPERL_OP_AV_POP, GOPERL_TOK(av), 0))
#define av_top_index(av) av_len(av)
#define av_tindex(av) av_len(av)
#define av_count(av) ((STRLEN)(av_len(av) + 1))
#define hv_delete_ent(hv, keysv, flags, hash)                              \
    GOPERL_SV(goperl_op3(GOPERL_OP_HV_DELETE_ENT, GOPERL_TOK(hv),          \
                         GOPERL_TOK(keysv), (uint64_t)(int64_t)(I32)(flags)))
/* A negative klen is the length of a key that is UTF-8, as for hv_fetch
 * and hv_store. */
static I32 goperl_hv_exists(goperl_frame_t *f, HV *hv, const char *key,
                            I32 klen) {
    return goperl_hv_exists_flags(f, hv, key, klen, 0);
}
/* action: hv_common's, of which HV_DISABLE_UVAR_XKEY is taken (a uvar
 * magic's function that looks a key up in its own hash asks for it, or it
 * would be called again for ever: English::Name) */
static I32 goperl_hv_exists_flags(goperl_frame_t *f, HV *hv, const char *key, I32 klen,
                                  int action) {
    const int utf8 = klen < 0;
    const uint64_t len = utf8 ? (uint64_t)-(int64_t)klen : (uint64_t)klen;
    uint64_t k = goperl_do_op(f, GOPERL_OP_NEW_PVN, 0, len, key, len);
    goperl_do_op(f, GOPERL_OP_SV_MORTAL, k, 0, 0, 0);
    if (utf8) goperl_do_op(f, GOPERL_OP_SV_UTF8_ON, k, 0, 0, 0);
    return (I32)goperl_do_op3(f, GOPERL_OP_HV_EXISTS_ENT, GOPERL_TOK(hv), k,
                              (uint64_t)(uint32_t)(action & HV_DISABLE_UVAR_XKEY), 0, 0);
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
#define SvUOK(sv) SvIOK_UV(sv)
/* Dists pair these with a value write (sv_setiv/SvIV_set), which already
 * sets the guest flags — the standalone flag flip is then a no-op. */
/* The process id is the program's. A fork child is a process of its own
 * to the program although it runs inside the host process (perl's own
 * fork emulation gives XS code the same view of getpid), so a module that
 * asks who it is gets the answer $$ gives. Outside a call from the
 * interpreter there is no program to ask, and the host's id is the answer. */
#define GOPERL_PL_PID 61
static pid_t goperl_getpid(void) __attribute__((unused));
static pid_t goperl_getpid(void) {
    goperl_frame_t *f = goperl_cur_frame_v;
    if (!f) return getpid();
    return (pid_t)(int64_t)goperl_do_op(f, GOPERL_OP_PLVAR_GET, GOPERL_PL_PID,
                                        0, 0, 0);
}

/* ---- waiting for children ------------------------------------------------
 * The children of the program are the interpreter's: the processes its
 * fork made (pseudo-processes, which run inside this process and which
 * the C library knows nothing of) and the commands it started, which the
 * runtime waits for itself. So waitpid, wait, wait3 and wait4, as a module
 * calls them, are the interpreter's waitpid, as Perl's waitpid is: native
 * code sees the program's children (an event loop's child watcher reaps a
 * fork child with waitpid(-1, ...)), and a wait for any child never takes
 * a command from under the runtime. A child native code started itself is
 * found by its pid, as perl finds it.
 *
 * They are functions of the module itself (hidden), as sigaction is. Where
 * the interpreter cannot be asked, the C library is asked for a child
 * named by its pid, and a wait for any child finds none (ECHILD): in a
 * child of fork(2) native code made, which is not the program; on a thread
 * that runs no call of the interpreter's (a thread of the module's own);
 * in a signal handler, which may have stopped the interpreter anywhere. A
 * program that is not the host process (an interpreter embedded in a Go
 * program) has no children of the host's: there a child native code
 * started, named by its pid, is the C library's too.
 *
 * What is reported is a child that ended; the interpreter's children are
 * not reported stopped or continued (WUNTRACED, WCONTINUED). wait3 and
 * wait4 report no resource usage of the child (all zero). The status word
 * is the host's: a signal is the host's number for it. An error is the
 * interpreter's errno, which crosses as its number and is the host's
 * errno for it here. */
#include <sys/resource.h>
#define GOPERL_INT_WAITPID 35 /* a: the pid; b: 1 for WNOHANG -> the pid
                               * reaped (0: none has ended), second: its
                               * status word; or -1, second: the errno */
#define GOPERL_ERRNO_GUEST_ECHILD 12
#if defined(__cplusplus) && defined(__THROWNL)
#define GOPERL_LIBC_THROWNL __THROWNL
#else
#define GOPERL_LIBC_THROWNL
#endif
GOPERL_C __attribute__((weak, visibility("hidden"))) pid_t (*goperl_libc_wait4)(
    pid_t, int *, int, struct rusage *) = 0;
/* The host's number for a signal of the interpreter's (Linux's numbers, on
 * every host). */
static int goperl_signal_host(int sig) __attribute__((unused));
static int goperl_signal_host(int sig) {
    switch (sig) {
    case 1: return SIGHUP;
    case 2: return SIGINT;
    case 3: return SIGQUIT;
    case 4: return SIGILL;
    case 5: return SIGTRAP;
    case 6: return SIGABRT;
    case 7: return SIGBUS;
    case 8: return SIGFPE;
    case 9: return SIGKILL;
    case 10: return SIGUSR1;
    case 11: return SIGSEGV;
    case 12: return SIGUSR2;
    case 13: return SIGPIPE;
    case 14: return SIGALRM;
    case 15: return SIGTERM;
    case 17: return SIGCHLD;
    case 18: return SIGCONT;
    case 19: return SIGSTOP;
    case 20: return SIGTSTP;
    case 21: return SIGTTIN;
    case 22: return SIGTTOU;
    case 23: return SIGURG;
    case 24: return SIGXCPU;
    case 25: return SIGXFSZ;
    case 26: return SIGVTALRM;
    case 27: return SIGPROF;
    case 28: return SIGWINCH;
    case 29: return SIGIO;
    case 31: return SIGSYS;
    }
    return sig;
}
/* The host's status word for one of the interpreter's. */
static int goperl_wait_status_host(int st) {
    const int low = st & 0x7f;
    if (low == 0) return st & 0xffff;
    if (low == 0x7f) return (goperl_signal_host((st >> 8) & 0xff) << 8) | 0x7f;
    return goperl_signal_host(low) | (st & 0x80);
}
/* The frame of the call of the interpreter's this thread runs, or NULL: a
 * thread that runs none has no slot, and none is made for it. */
static goperl_frame_t *goperl_thread_frame(void) {
    void *tid = (void *)pthread_self();
    goperl_tframe_t *tf = GOPERL_PROCESS->tf;
    uint32_t h = (uint32_t)(((uintptr_t)tid >> 4) % GOPERL_TFRAME_SLOTS);
    for (uint32_t i = 0; i < GOPERL_TFRAME_SLOTS; i++) {
        goperl_tframe_t *e = &tf[(h + i) % GOPERL_TFRAME_SLOTS];
        if (e->tid == tid) return e->f;
        if (e->tid == 0) return NULL;
    }
    return NULL;
}
/* Whether the thread runs a signal handler: the process's handlers all run
 * on the alternate signal stack of the thread (see sigaction above). */
static int goperl_in_signal_handler(void) {
    stack_t ss;
    return sigaltstack(NULL, &ss) == 0 && (ss.ss_flags & SS_ONSTACK) != 0;
}
/* The interpreter's number for a signal of the host's. */
static int goperl_signal_guest(int sig) {
    for (int g = 1; g < 32; g++)
        if (goperl_signal_host(g) == sig) return g;
    return sig;
}
/* The program's action of a signal of the host's, as its C library has it:
 * 0 the default, 1 ignored, 2 a handler; -1 where the interpreter cannot
 * be asked (see goperl_wait_child). */
#define GOPERL_INT_SIGACTION 40 /* b: a signal -> 0 default, 1 ignored, 2 a handler */
static int goperl_program_action(int sig) {
    goperl_frame_t *const f =
        goperl_api_v && !GOPERL_PROCESS->forked && !goperl_in_signal_handler()
            ? goperl_thread_frame()
            : NULL;
    if (!f) return -1;
    const int e = errno;
    const int r = (int)goperl_xs_opx(f, GOPERL_OP_INTERNAL, 0, (uint64_t)goperl_signal_guest(sig),
                                     GOPERL_INT_SIGACTION, 0, 0, 0);
    errno = e;
    return r;
}
static pid_t goperl_wait_child(pid_t pid, int *status, int options, struct rusage *usage) {
    if (!goperl_libc_wait4) {
        goperl_libc_wait4 = (pid_t (*)(pid_t, int *, int, struct rusage *))dlsym(RTLD_NEXT, "wait4");
        if (!goperl_libc_wait4) {
            errno = ENOSYS;
            return -1;
        }
    }
    goperl_frame_t *const f =
        goperl_api_v && !GOPERL_PROCESS->forked && !goperl_in_signal_handler()
            ? goperl_thread_frame()
            : NULL;
    if (f) {
        uint64_t second = 0;
        const int64_t got = (int64_t)goperl_xs_opx(
            f, GOPERL_OP_INTERNAL, (uint64_t)(int64_t)pid, (options & WNOHANG) ? 1 : 0,
            GOPERL_INT_WAITPID, 0, 0, &second);
        if (got >= 0) {
            if (usage) memset(usage, 0, sizeof *usage);
            if (got > 0 && status) *status = goperl_wait_status_host((int)(uint32_t)second);
            return (pid_t)got;
        }
        if ((int)second != GOPERL_ERRNO_GUEST_ECHILD || pid <= 0 ||
            (pid_t)(int64_t)goperl_xs_opx(f, GOPERL_OP_PLVAR_GET, GOPERL_PL_PID, 0, 0, 0, 0, 0) ==
                getpid()) {
            errno = goperl_errno_host((int)second);
            return -1;
        }
    } else if (pid <= 0) {
        errno = ECHILD;
        return -1;
    }
    return goperl_libc_wait4(pid, status, options, usage);
}
GOPERL_C __attribute__((weak, visibility("hidden"))) pid_t waitpid(pid_t pid, int *status,
                                                                   int options) {
    return goperl_wait_child(pid, status, options, NULL);
}
GOPERL_C __attribute__((weak, visibility("hidden"))) pid_t wait(int *status) {
    return goperl_wait_child(-1, status, 0, NULL);
}
GOPERL_C __attribute__((weak, visibility("hidden"))) pid_t wait3(
    int *status, int options, struct rusage *usage) GOPERL_LIBC_THROWNL {
    return goperl_wait_child(-1, status, options, usage);
}
GOPERL_C __attribute__((weak, visibility("hidden"))) pid_t wait4(
    pid_t pid, int *status, int options, struct rusage *usage) GOPERL_LIBC_THROWNL {
    return goperl_wait_child(pid, status, options, usage);
}
/* ---- waits with a signal mask -----------------------------------------------
 * epoll_pwait, ppoll and pselect put the signal mask they are given in
 * place for the wait alone, so that a signal the program blocks otherwise
 * ends the wait (IO::Async::Loop::Epoll blocks SIGCHLD with sigprocmask
 * and waits with a mask that lets it in). The program's signals and their
 * mask are the interpreter's, not the kernel's: so the mask is the
 * program's while the wait runs, a signal for the program ends the wait
 * with EINTR as any signal does a wait of native code, and what arrived is
 * delivered, with the wait's mask still in place, before the mask before
 * is put back, as the kernel delivers it. The mask is a sigset_t of the
 * interpreter's (POSIX::SigSet's, which native code hands on as it is) or
 * one of the host's. Where the interpreter cannot be asked (see
 * goperl_wait_child) the C library's wait runs as it is.
 *
 * They are functions of the module itself (hidden), as sigaction is. */
#define GOPERL_INT_WAIT_MASK 41 /* a: 1 begin, 0 end; b: the mask or the address of
                                 * a sigset_t of the interpreter's (c's high half 1)
                                 * -> the mask before; second: a handler ran */
static goperl_frame_t *goperl_wait_frame(const void *mask) {
    return mask && goperl_api_v && !GOPERL_PROCESS->forked && !goperl_in_signal_handler()
               ? goperl_thread_frame()
               : NULL;
}
/* The wait's mask in place: the mask before in *old; 1 when a signal it
 * lets in was delivered at once, which ends the wait before it begins. */
static int goperl_wait_mask_begin(goperl_frame_t *f, const sigset_t *mask, uint64_t *old) {
    uint64_t b = 0, form = 0;
    const char *const m = (const char *)mask;
    if (f->membase && m >= f->membase && m < f->membase + *f->memsize) {
        b = (uint64_t)(m - f->membase);
        form = 1;
    } else {
        for (int g = 1; g <= 64; g++)
            if (sigismember(mask, goperl_signal_host(g)) == 1) b |= (uint64_t)1 << (g - 1);
    }
    const int e = errno;
    uint64_t ran = 0;
    *old = goperl_xs_opx(f, GOPERL_OP_INTERNAL, 1, b, GOPERL_INT_WAIT_MASK | (form << 32), 0, 0,
                         &ran);
    errno = e;
    return ran != 0;
}
/* The wait has ended with r: what arrived is delivered, the mask before is
 * put back, and a wait that found nothing ends with EINTR when a handler
 * ran. */
static int goperl_wait_mask_end(goperl_frame_t *f, uint64_t old, int r) {
    const int e = errno;
    uint64_t ran = 0;
    (void)goperl_xs_opx(f, GOPERL_OP_INTERNAL, 0, old, GOPERL_INT_WAIT_MASK, 0, 0, &ran);
    errno = e;
    if (r == 0 && ran) {
        errno = EINTR;
        return -1;
    }
    return r;
}
GOPERL_C __attribute__((weak, visibility("hidden"))) int (*goperl_libc_pselect)(
    int, fd_set *, fd_set *, fd_set *, const struct timespec *, const sigset_t *) = 0;
GOPERL_C __attribute__((weak, visibility("hidden"))) int pselect(
    int n, fd_set *r, fd_set *w, fd_set *x, const struct timespec *timeout,
    const sigset_t *mask) {
    if (!goperl_libc_pselect) {
        goperl_libc_pselect = (int (*)(int, fd_set *, fd_set *, fd_set *, const struct timespec *,
                                       const sigset_t *))dlsym(RTLD_NEXT, "pselect");
        if (!goperl_libc_pselect) {
            errno = ENOSYS;
            return -1;
        }
    }
    goperl_frame_t *const f = goperl_wait_frame(mask);
    if (!f) return goperl_libc_pselect(n, r, w, x, timeout, mask);
    uint64_t old;
    if (goperl_wait_mask_begin(f, mask, &old)) {
        (void)goperl_wait_mask_end(f, old, 0);
        errno = EINTR;
        return -1;
    }
    return goperl_wait_mask_end(f, old, goperl_libc_pselect(n, r, w, x, timeout, NULL));
}
#ifdef __linux__
#include <poll.h>
#include <sys/epoll.h>
GOPERL_C __attribute__((weak, visibility("hidden"))) int (*goperl_libc_ppoll)(
    struct pollfd *, nfds_t, const struct timespec *, const sigset_t *) = 0;
GOPERL_C __attribute__((weak, visibility("hidden"))) int ppoll(
    struct pollfd *fds, nfds_t nfds, const struct timespec *timeout, const sigset_t *mask) {
    if (!goperl_libc_ppoll) {
        goperl_libc_ppoll = (int (*)(struct pollfd *, nfds_t, const struct timespec *,
                                     const sigset_t *))dlsym(RTLD_NEXT, "ppoll");
        if (!goperl_libc_ppoll) {
            errno = ENOSYS;
            return -1;
        }
    }
    goperl_frame_t *const f = goperl_wait_frame(mask);
    if (!f) return goperl_libc_ppoll(fds, nfds, timeout, mask);
    uint64_t old;
    if (goperl_wait_mask_begin(f, mask, &old)) {
        (void)goperl_wait_mask_end(f, old, 0);
        errno = EINTR;
        return -1;
    }
    return goperl_wait_mask_end(f, old, goperl_libc_ppoll(fds, nfds, timeout, NULL));
}
GOPERL_C __attribute__((weak, visibility("hidden"))) int (*goperl_libc_epoll_pwait)(
    int, struct epoll_event *, int, int, const sigset_t *) = 0;
GOPERL_C __attribute__((weak, visibility("hidden"))) int epoll_pwait(
    int epfd, struct epoll_event *events, int maxevents, int timeout, const sigset_t *mask) {
    if (!goperl_libc_epoll_pwait) {
        goperl_libc_epoll_pwait = (int (*)(int, struct epoll_event *, int, int,
                                           const sigset_t *))dlsym(RTLD_NEXT, "epoll_pwait");
        if (!goperl_libc_epoll_pwait) {
            errno = ENOSYS;
            return -1;
        }
    }
    goperl_frame_t *const f = goperl_wait_frame(mask);
    if (!f) return goperl_libc_epoll_pwait(epfd, events, maxevents, timeout, mask);
    uint64_t old;
    if (goperl_wait_mask_begin(f, mask, &old)) {
        (void)goperl_wait_mask_end(f, old, 0);
        errno = EINTR;
        return -1;
    }
    return goperl_wait_mask_end(f, old,
                                goperl_libc_epoll_pwait(epfd, events, maxevents, timeout, NULL));
}
#endif

#define getpid goperl_getpid
#define PerlProc_getpid() ((int)goperl_getpid())
#define PerlProc_getuid() ((int)getuid())
#define PerlEnv_getenv(n) getenv(n)
#define XSRETURN_NV(v)                            \
    STMT_START {                                  \
        ST(0) = sv_2mortal(newSVnv((NV)(v)));     \
        XSRETURN(1);                              \
    }                                             \
    STMT_END
#define DEBUG_v(x) ((void)0)
#define sv_copypv(dsv, ssv) \
    ({ GOPERL_ONCE(goperl_sv_copypv_dsv_, dsv); GOPERL_ONCE(goperl_sv_copypv_ssv_, ssv); \
       do { STRLEN goperl_cl; const char *goperl_cp = SvPV((goperl_sv_copypv_ssv_), goperl_cl); sv_setpvn((goperl_sv_copypv_dsv_), goperl_cp, goperl_cl); if (SvUTF8(goperl_sv_copypv_ssv_)) SvUTF8_on(goperl_sv_copypv_dsv_); else SvUTF8_off(goperl_sv_copypv_dsv_); } while (0); })
#define AMG_CALLunary(sv, meth)                                     \
    GOPERL_SV(goperl_op3(GOPERL_OP_AMAGIC_CALL, GOPERL_TOK(sv), 0, \
                         (uint64_t)(int64_t)(meth)))
#define AMG_CALLun(sv, meth) AMG_CALLunary((sv), meth##_amg)
/* load_module is perl's `use NAME VERSION LIST` made at once (`no` with
 * PERL_LOADMOD_DENY, `use NAME ()` with PERL_LOADMOD_NOIMPORT): the module
 * is required, asked for the version, and its import (unimport) is called
 * with the arguments, which end at a NULL. Called while code is compiled
 * (Quote::Code loads charnames as it reads "\N{...}"), the import changes
 * the hints of that code, as a BEGIN block's would. The name, the version
 * and the arguments are the call's: their references are released. An op
 * tree for the import (PERL_LOADMOD_IMPORT_OPS) is not taken: the call
 * dies. */
static void goperl_load_module(goperl_frame_t *f, U32 flags, SV *name,
                               SV *ver, va_list *args) {
    uint64_t len = 0;
    const char *p = goperl_direct_sv_pv(f, GOPERL_TOK(name), &len);
    char buf[512];
    SV *list[64];
    int n = 0;
    if (flags & PERL_LOADMOD_IMPORT_OPS)
        goperl_croakf(f, "load_module with an op tree for the import "
                         "(PERL_LOADMOD_IMPORT_OPS) is not supported");
    if (!(flags & PERL_LOADMOD_NOIMPORT)) {
        SV *sv;
        while ((sv = va_arg(*args, SV *)) != NULL) {
            if (n == (int)(sizeof list / sizeof list[0]))
                goperl_croakf(f, "load_module: more than %d import arguments",
                              (int)(sizeof list / sizeof list[0]));
            list[n++] = sv;
        }
    }
    if (len > sizeof buf - 16)
        goperl_croakf(f, "load_module: a module name of %llu bytes",
                      (unsigned long long)len);
    snprintf(buf, sizeof buf, "require %.*s", (int)len, p ? p : "");
    goperl_do_op(f, GOPERL_OP_EVAL_PV, 1, 0, buf, __builtin_strlen(buf));
    if (ver) {
        dSP;
        ENTER;
        SAVETMPS;
        PUSHMARK(SP);
        XPUSHs(name);
        XPUSHs(ver);
        PUTBACK;
        call_method("VERSION", G_VOID | G_DISCARD);
        FREETMPS;
        LEAVE;
    }
    if (!(flags & PERL_LOADMOD_NOIMPORT)) {
        const char *const method = (flags & PERL_LOADMOD_DENY) ? "unimport" : "import";
        dSP;
        ENTER;
        SAVETMPS;
        PUSHMARK(SP);
        XPUSHs(name);
        for (int i = 0; i < n; i++) XPUSHs(list[i]);
        PUTBACK;
        /* perl's method lookup makes a missing import (unimport) one that
         * does nothing, as `use` has it */
        call_method(method, G_VOID | G_DISCARD);
        FREETMPS;
        LEAVE;
    }
    for (int i = 0; i < n; i++) SvREFCNT_dec(list[i]);
    if (ver) SvREFCNT_dec(ver);
    SvREFCNT_dec(name);
}
/* Functions with perl's prototypes: the module names are variadic. */
static void Perl_load_module(pTHX_ U32 flags, SV *name, SV *ver, ...)
    __attribute__((unused));
static void Perl_load_module(pTHX_ U32 flags, SV *name, SV *ver, ...) {
    va_list args;
    va_start(args, ver);
    goperl_load_module(goperl_cur_frame_v, flags, name, ver, &args);
    va_end(args);
}
static void load_module(U32 flags, SV *name, SV *ver, ...)
    __attribute__((unused));
static void load_module(U32 flags, SV *name, SV *ver, ...) {
    va_list args;
    va_start(args, ver);
    goperl_load_module(goperl_cur_frame_v, flags, name, ver, &args);
    va_end(args);
}

/* The SV arenas, which XS walks to list every SV (Devel::Gladiator,
 * Devel::Leak, Devel::FindBlessedRefs): go-perl does not give native code
 * the interpreter's arenas, and a walk dies saying so when it starts. It
 * is not refused where it is compiled: Moose compiles such a walk (its
 * fix-up of overloading after a role is applied to an object) that it runs
 * only on perls before 5.18, where overloading was a flag of every
 * reference. */
static SV *goperl_sv_arenaroot(goperl_frame_t *f) __attribute__((unused));
static SV *goperl_sv_arenaroot(goperl_frame_t *f) {
    goperl_croakf(f, "go-perl gives native code no access to this interpreter internal: "
                     "PL_sv_arenaroot");
}
#define PL_sv_arenaroot (goperl_sv_arenaroot(_gof))
#define SvANY(sv) ((void *)0)
#define SVTYPEMASK 0xff

/* ---- memory helpers ------------------------------------------------------ */

/* A module's own allocations are the C library's. A block of the
 * interpreter's memory is a different thing: the string buffer of an SV,
 * which a module may hold (SvPVX), free (Safefree(SvPVX(sv)) before
 * SvPV_set(sv, NULL)) or resize (Renew). Such a block belongs to the
 * interpreter's allocator, and freeing or resizing it is the interpreter's
 * to do. Which of the two a pointer is, its address says: the interpreter's
 * memory is one reserved range of this process. */
static inline __attribute__((unused)) uint64_t goperl_guest_addr(
    const goperl_frame_t *f, const void *p) {
    const uintptr_t d = (uintptr_t)p - (uintptr_t)f->membase;
    return (p && d < *f->memsize) ? (uint64_t)d : 0;
}
/* The blocks this module's allocators gave (Newx, Newxz, Newxc, Renew,
 * safemalloc, savepv, ...), with their sizes. On perl they are blocks of
 * the interpreter's allocator, and a module hands one to an SV as the SV's
 * buffer (SvPV_set and SvLEN_set, the sv_usepvn idiom: CBOR::Free's
 * encoder, JSON::XS::ByteString's strings): the SV owns it from then on.
 * Here they are the C library's, which the interpreter cannot reach, so a
 * block handed to an SV becomes a block of the interpreter's of the same
 * size and contents (goperl_sv_pvx_store). Only the address says that a
 * pointer is such a block, and how big it is: this table. It is the
 * module's own (every module has its copy of these functions), guarded for
 * the threads native code may allocate on. */
typedef struct goperl_block {
    void *p;
    size_t n;
} goperl_block_t;
#define GOPERL_BLOCK_GONE ((void *)1)
/* one table for the module's every file (a weak definition each file has,
 * which the link makes one; hidden: another module has its own) */
GOPERL_C __attribute__((weak, visibility("hidden"))) goperl_block_t *goperl_blocks_v = 0;
GOPERL_C __attribute__((weak, visibility("hidden"))) size_t goperl_blocks_cap = 0;
GOPERL_C __attribute__((weak, visibility("hidden"))) size_t goperl_blocks_fill = 0; /* live and gone */
GOPERL_C __attribute__((weak, visibility("hidden"))) int goperl_blocks_lock_v = 0;
static void goperl_blocks_lock(void) {
    while (__atomic_exchange_n(&goperl_blocks_lock_v, 1, __ATOMIC_ACQUIRE)) {
    }
}
static void goperl_blocks_unlock(void) { __atomic_store_n(&goperl_blocks_lock_v, 0, __ATOMIC_RELEASE); }
static size_t goperl_block_hash(const void *p, size_t cap) {
    uint64_t x = (uint64_t)(uintptr_t)p >> 4;
    x ^= x >> 29;
    x *= 0xbf58476d1ce4e5b9ULL;
    x ^= x >> 32;
    return (size_t)x & (cap - 1);
}
/* The slot of p, or the empty one it would go in (under the lock). */
static goperl_block_t *goperl_block_slot(const void *p) {
    size_t i = goperl_block_hash(p, goperl_blocks_cap);
    goperl_block_t *gone = 0;
    for (;;) {
        goperl_block_t *b = &goperl_blocks_v[i];
        if (b->p == p) return b;
        if (!b->p) return gone ? gone : b;
        if (b->p == GOPERL_BLOCK_GONE && !gone) gone = b;
        i = (i + 1) & (goperl_blocks_cap - 1);
    }
}
static void goperl_block_note(void *p, size_t n) {
    if (!p) return;
    goperl_blocks_lock();
    if ((goperl_blocks_fill + 1) * 2 > goperl_blocks_cap) {
        /* a table twice as big as the live blocks need, gone slots dropped */
        size_t live = 0, cap = 64;
        for (size_t i = 0; i < goperl_blocks_cap; i++)
            if (goperl_blocks_v[i].p && goperl_blocks_v[i].p != GOPERL_BLOCK_GONE) live++;
        while (cap < (live + 1) * 4) cap *= 2;
        goperl_block_t *old = goperl_blocks_v;
        const size_t oldcap = goperl_blocks_cap;
        goperl_block_t *tab = (goperl_block_t *)calloc(cap, sizeof(goperl_block_t));
        if (!tab) { /* the block stays unnoted: a buffer handed over is refused */
            goperl_blocks_unlock();
            return;
        }
        goperl_blocks_v = tab;
        goperl_blocks_cap = cap;
        goperl_blocks_fill = 0;
        for (size_t i = 0; i < oldcap; i++)
            if (old[i].p && old[i].p != GOPERL_BLOCK_GONE) {
                *goperl_block_slot(old[i].p) = old[i];
                goperl_blocks_fill++;
            }
        free(old);
    }
    goperl_block_t *b = goperl_block_slot(p);
    if (b->p != p) {
        if (!b->p) goperl_blocks_fill++;
        b->p = p;
    }
    b->n = n;
    goperl_blocks_unlock();
}
/* Forgets p; with size, says whether it was a block and how big. */
static int goperl_block_forget(const void *p, size_t *size) {
    int was = 0;
    if (!p) return 0;
    goperl_blocks_lock();
    if (goperl_blocks_cap) {
        goperl_block_t *b = goperl_block_slot(p);
        if (b->p == p) {
            if (size) *size = b->n;
            b->p = GOPERL_BLOCK_GONE;
            was = 1;
        }
    }
    goperl_blocks_unlock();
    return was;
}
/* Whether p is a block, and how big. */
static int goperl_block_size(const void *p, size_t *size) {
    int is = 0;
    if (!p) return 0;
    goperl_blocks_lock();
    if (goperl_blocks_cap) {
        goperl_block_t *b = goperl_block_slot(p);
        if (b->p == p) {
            *size = b->n;
            is = 1;
        }
    }
    goperl_blocks_unlock();
    return is;
}
static void *goperl_block_alloc(size_t n, int zero) {
    void *p = zero ? calloc(1, n ? n : 1) : malloc(n ? n : 1);
    goperl_block_note(p, n);
    return p;
}
static void goperl_safefree(void *p) __attribute__((unused));
static void goperl_safefree(void *p) {
    goperl_frame_t *f = goperl_cur_frame_v;
    /* a copy SvMAGIC showed of the interpreter's magic is the SDK's, and
     * the magic stays the interpreter's (a module that frees the MAGIC it
     * unlinked meant to take it off the value) */
    if (f && p && goperl_magic_is_interp_copy(f, (const MAGIC *)p))
        goperl_magic_unlinked(f, ((const MAGIC *)p)->mg_type);
    const uint64_t g = f ? goperl_guest_addr(f, p) : 0;
    if (g) {
        goperl_do_op(f, GOPERL_OP_MEM_FREE, g, 0, 0, 0);
    } else {
        goperl_block_forget(p, 0);
        free(p);
    }
}
static void *goperl_saferealloc(void *p, size_t n) __attribute__((unused));
static void *goperl_saferealloc(void *p, size_t n) {
    goperl_frame_t *f = goperl_cur_frame_v;
    const uint64_t g = f ? goperl_guest_addr(f, p) : 0;
    if (!g) {
        void *q = realloc(p, n ? n : 1);
        if (q) {
            goperl_block_forget(p, 0);
            goperl_block_note(q, n);
        }
        return q;
    }
    const uint64_t moved = goperl_do_op(f, GOPERL_OP_MEM_REALLOC, g, n, 0, 0);
    return moved ? f->membase + moved : 0;
}
#define Newx(ptr, n, t) ((ptr) = (t *)goperl_block_alloc((size_t)(n) * sizeof(t), 0))
#define Newxz(ptr, n, t) ((ptr) = (t *)goperl_block_alloc((size_t)(n) * sizeof(t), 1))
#define Newxc(ptr, n, t, c) ((ptr) = (c *)goperl_block_alloc((size_t)(n) * sizeof(t), 0))
#define safemalloc(n) goperl_block_alloc((size_t)(n), 0)
#define safefree(p) goperl_safefree((void *)(p))
#define SETERRNO(e, x) (errno = (e))
#define Renew(ptr, n, t) \
    ({ __typeof__(ptr) *const goperl_Renew_ptr_ = &(ptr); \
       *goperl_Renew_ptr_ = (t *)goperl_saferealloc((void *)*goperl_Renew_ptr_, (size_t)(n) * sizeof(t)); })
#define Safefree(p) goperl_safefree((void *)(p))
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
#define sv_free(sv) SvREFCNT_dec(sv)
#define strnEQ(a, b, n) (strncmp((a), (b), (n)) == 0)
#define strnNE(a, b, n) (strncmp((a), (b), (n)) != 0)
typedef U16 OPCODE;
#define GvEGV(gv) \
    ((GV *)GOPERL_SV(goperl_op0(GOPERL_OP_GV_PTR, GOPERL_TOK(gv), 4)))
#define GvNAMELEN(gv) ((U32)goperl_op_second(_gof, GOPERL_OP_GV_NAME, GOPERL_TOK(gv), 0, 0))
#define hv_clear(hv) ((void)goperl_op0(GOPERL_OP_HV_CLEAR, GOPERL_TOK(hv), 0))
#define hv_delete(hv, key, klen, flags)                                    \
    ({ const I32 goperl_hv_delete_klen_ = (I32)(klen);                     \
       GOPERL_SV(goperl_do_op3(_gof, GOPERL_OP_HV_DELETE, GOPERL_TOK(hv),  \
                               (uint64_t)(uint32_t)(flags),                \
                               (uint64_t)(int64_t)goperl_hv_delete_klen_, (key), \
                               (uint64_t)(goperl_hv_delete_klen_ < 0 ? -goperl_hv_delete_klen_ \
                                                                     : goperl_hv_delete_klen_))); })
#define Perl_debug_log PerlIO_stderr()
#define do_sv_dump(level, fp, sv, nest, maxnest, dumpops, pvlim)           \
    goperl_perlio_printf((fp),                     \
                         "<sv dump unavailable in the go-perl XS SDK>\n")

/* grok_number (numeric.c): classifies a decimal number, with the integer
 * limits of the interpreter (UV_MAX above). The radix is '.': the
 * interpreter runs in the C locale. */
#define IS_NUMBER_IN_UV 0x01
#define IS_NUMBER_GREATER_THAN_UV_MAX 0x02
#define IS_NUMBER_NOT_INT 0x04
#define IS_NUMBER_NEG 0x08
#define IS_NUMBER_INFINITY 0x10
#define IS_NUMBER_NAN 0x20
/* grok_infnan: "Inf", "Infinity", "NaN" (with the q/s suffixes) and the
 * "1.#INF" / "1.#IND" / "1.#QNAN" spellings, any letter case, optionally
 * followed by spaces. Payload forms ("nan(123)") are not recognized. */
static int goperl_grok_infnan(const char **sp, const char *send)
    __attribute__((unused));
static int goperl_grok_infnan(const char **sp, const char *send) {
    const char *s = *sp;
    int kind = 0;
    size_t n = (size_t)(send - s);
    if (n >= 3 && s[0] == '1' && s[1] == '.' && s[2] == '#') {
        s += 3;
        n -= 3;
        if (n >= 3 && strncasecmp(s, "INF", 3) == 0) {
            kind = 0x10;
            s += 3;
        } else if (n >= 3 && strncasecmp(s, "IND", 3) == 0) {
            kind = 0x20;
            s += 3;
        } else if (n >= 4 && (strncasecmp(s, "QNAN", 4) == 0 ||
                              strncasecmp(s, "SNAN", 4) == 0)) {
            kind = 0x20;
            s += 4;
        } else if (n >= 3 && strncasecmp(s, "NAN", 3) == 0) {
            kind = 0x20;
            s += 3;
        }
    } else if (n >= 8 && strncasecmp(s, "infinity", 8) == 0) {
        kind = 0x10;
        s += 8;
    } else if (n >= 3 && strncasecmp(s, "inf", 3) == 0) {
        kind = 0x10;
        s += 3;
    } else if (n >= 3 && strncasecmp(s, "nan", 3) == 0) {
        kind = 0x20;
        s += 3;
        if (s < send && (*s == 'q' || *s == 'Q' || *s == 's' || *s == 'S')) s++;
    }
    if (!kind) return 0;
    while (s < send && isspace((unsigned char)*s)) s++;
    if (s < send) return 0;
    *sp = s;
    return kind | 0x04; /* IS_NUMBER_NOT_INT */
}
static int goperl_grok_number_flags(const char *pv, STRLEN len, UV *valuep,
                                    U32 flags) __attribute__((unused));
static int goperl_grok_number_flags(const char *pv, STRLEN len, UV *valuep,
                                    U32 flags) {
    const char *s = pv;
    const char *const send = pv + len;
    const char *d;
    int numtype = 0;
    const UV max_div_10 = UV_MAX / 10;
    const U8 max_mod_10 = (U8)(UV_MAX % 10);
    while (s < send && isspace((unsigned char)*s)) s++;
    if (s == send) return 0;
    if (*s == '-') {
        s++;
        numtype = IS_NUMBER_NEG;
    } else if (*s == '+') {
        s++;
    }
    if (s == send) return 0;
    d = s;
    if (*s >= '0' && *s <= '9') {
        UV value = 0;
        int overflow = 0;
        for (; s < send && *s >= '0' && *s <= '9'; s++) {
            const U8 digit = (U8)(*s - '0');
            if (overflow) continue;
            if (value > max_div_10 ||
                (value == max_div_10 && digit > max_mod_10))
                overflow = 1;
            else
                value = value * 10 + digit;
        }
        if (overflow) {
            numtype |= IS_NUMBER_GREATER_THAN_UV_MAX;
        } else {
            numtype |= IS_NUMBER_IN_UV;
            if (valuep) *valuep = value;
        }
        if (s < send && *s == '.') {
            s++;
            numtype |= IS_NUMBER_NOT_INT;
            while (s < send && *s >= '0' && *s <= '9') s++;
        }
    } else if (*s == '.') {
        s++;
        numtype |= IS_NUMBER_NOT_INT | IS_NUMBER_IN_UV;
        /* no digits before the radix: there must be some after it */
        if (s < send && *s >= '0' && *s <= '9') {
            do {
                s++;
            } while (s < send && *s >= '0' && *s <= '9');
            if (valuep) *valuep = 0;
        } else {
            return 0;
        }
    }
    if (s > d && s < send && (*s == 'e' || *s == 'E')) {
        s++;
        if (s < send && (*s == '-' || *s == '+')) s++;
        if (s < send && *s >= '0' && *s <= '9') {
            do {
                s++;
            } while (s < send && *s >= '0' && *s <= '9');
        } else if (flags & 0x10 /* PERL_SCAN_TRAILING */) {
            return numtype | 0x40 /* IS_NUMBER_TRAILING */;
        } else {
            return 0;
        }
        /* an exponent leaves only the sign of what was learnt */
        numtype &= IS_NUMBER_NEG;
        numtype |= IS_NUMBER_NOT_INT;
    }
    while (s < send && isspace((unsigned char)*s)) s++;
    if (s >= send) return numtype;
    if (len == 10 && memcmp(pv, "0 but true", 10) == 0) {
        if (valuep) *valuep = 0;
        return IS_NUMBER_IN_UV;
    }
    if (s + 2 < send && strchr("inqsINQS#", *s)) {
        /* from the first digit: "1.#INF" has been read as a number so far */
        const int infnan = goperl_grok_infnan(&d, send);
        if (infnan & IS_NUMBER_INFINITY) return numtype | infnan;
        if (infnan & IS_NUMBER_NAN) return (numtype | infnan) & ~IS_NUMBER_NEG;
    } else if (flags & 0x10 /* PERL_SCAN_TRAILING */) {
        return numtype | 0x40 /* IS_NUMBER_TRAILING */;
    }
    return 0;
}
#define grok_number_flags(pv, len, vp, flags) \
    goperl_grok_number_flags((pv), (len), (vp), (U32)(flags))
#define grok_number(pv, len, vp) goperl_grok_number_flags((pv), (len), (vp), 0)
#define grok_infnan(sp, send) goperl_grok_infnan((sp), (send))

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

/* hv_iternextsv (hv.c): the next entry's key, where the entry has it
 * (hv_iterkey), and its value (hv_iterval); NULL at the end. */
static SV *goperl_hv_iternextsv(goperl_frame_t *f, HV *hv, char **keyp,
                                I32 *retlen) __attribute__((unused));
static SV *goperl_hv_iternextsv(goperl_frame_t *f, HV *hv, char **keyp,
                                I32 *retlen) {
    const HE *const he =
        goperl_he_at(f, goperl_do_op(f, GOPERL_OP_HV_ITERNEXT, GOPERL_TOK(hv), 0, 0, 0));
    if (!he) return 0;
    I32 len = 0;
    char *const key = goperl_hv_iterkey(f, he, &len);
    if (keyp) *keyp = key;
    if (retlen) *retlen = len;
    return goperl_hv_iterval(f, hv, he);
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
#define SvREADONLY_on(sv) GOPERL_FLAGS_CHANGE((sv), SVf_READONLY, 0)
#define sv_setuv_mg(sv, v) \
    ({ GOPERL_ONCE(goperl_sv_setuv_mg_sv_, sv); \
       (sv_setuv((goperl_sv_setuv_mg_sv_), (v)), (void)mg_set(goperl_sv_setuv_mg_sv_)); })
#define C_ARRAY_LENGTH(a) (sizeof(a) / sizeof((a)[0]))
#define save_scalar(gv) \
    GOPERL_SV(goperl_op0(GOPERL_OP_SAVE_SCALAR, GOPERL_TOK(gv), 0))
#define eval_pv(code, croak_on_err) \
    ({ GOPERL_ONCE(goperl_eval_pv_code_, code); \
       goperl_optree_flush(_gof); /* the code run reads the ops as native code left them */ \
       GOPERL_SV(goperl_ops(GOPERL_OP_EVAL_PV, (uint64_t)(int64_t)(croak_on_err), 0, (goperl_eval_pv_code_), __builtin_strlen(goperl_eval_pv_code_))); })
#define newCONSTSUB(stash, name, sv) \
    ({ GOPERL_ONCE(goperl_newCONSTSUB_name_, name); \
       ((CV *)GOPERL_SV(goperl_ops(GOPERL_OP_NEW_CONSTSUB, GOPERL_TOK(stash), GOPERL_TOK(sv), (goperl_newCONSTSUB_name_), __builtin_strlen(goperl_newCONSTSUB_name_)))); })
static SV *goperl_newSVpvn_flags(goperl_frame_t *f, const char *p,
                                 STRLEN len, U32 flags)
    __attribute__((unused));
static SV *goperl_newSVpvn_flags(goperl_frame_t *f, const char *p,
                                 STRLEN len, U32 flags) {
    uint64_t tok;
    if (p) {
        goperl_flush(f);
        tok = goperl_direct_new_pvn(f, p, (uint64_t)len);
        if (flags & SVf_UTF8)
            goperl_do_op(f, GOPERL_OP_SV_UTF8_ON, tok, 0, 0, 0);
    } else {
        tok = goperl_do_op(f, GOPERL_OP_NEW_SV, 0, 0, 0, 0);
    }
    if (flags & SVs_TEMP) tok = goperl_direct_sv_mortal(f, tok);
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

#define saferealloc(p, n) goperl_saferealloc((void *)(p), (size_t)(n))
/* The allocation functions as functions, for code that hands one on (the
 * allocator hooks of a library: `strm.bzalloc = ...`, `db_malloc =
 * safemalloc`): the macros above are function-like, so the bare name is
 * the function. */
static __attribute__((unused)) void *(safemalloc)(size_t n) { return malloc(n); }
static __attribute__((unused)) void *(safecalloc)(size_t n, size_t sz) { return goperl_block_alloc(n * sz, 1); }
static __attribute__((unused)) void *(saferealloc)(void *p, size_t n) { return goperl_saferealloc(p, n); }
static __attribute__((unused)) void(safefree)(void *p) { goperl_safefree(p); }
#ifndef NV_DIG
#define NV_DIG DBL_DIG
#endif
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
    goperl_block_forget(p, 0);
    free(p);
}
#define sv_usepvn(sv, p, len) goperl_sv_usepvn(_gof, (SV *)(sv), (p), (len))

/* sv_inc and sv_dec are the interpreter's: the magic increment of a
 * string ("aa" to "ab"), an integer that overflows into a number, the
 * overloaded ++ and --. */
#define sv_inc(sv) ((void)GOPERL_SVCALL(63 /* GOPERL_SVCALL_INC */, sv, 0))
#define sv_inc_nomg(sv) ((void)GOPERL_SVCALL(63 /* GOPERL_SVCALL_INC */, sv, 1))

#define gv_fetchfile_flags(name, namelen, flags) \
    ({ GOPERL_ONCE(goperl_gv_fetchfile_flags_namelen_, namelen); \
       ((GV *)GOPERL_SV(goperl_ops(GOPERL_OP_GV_FETCHFILE, 0, (uint64_t)(goperl_gv_fetchfile_flags_namelen_), (name), (uint64_t)(goperl_gv_fetchfile_flags_namelen_)))); })
#define gv_fetchfile(name) \
    ({ GOPERL_ONCE(goperl_gv_fetchfile_name_, name); \
       gv_fetchfile_flags((goperl_gv_fetchfile_name_), __builtin_strlen(goperl_gv_fetchfile_name_), 0); })

/* my_strlcat (perl's bundled strlcat) */
static Size_t my_strlcat(char *dst, const char *src, Size_t size)
    __attribute__((unused));
static Size_t my_strlcat(char *dst, const char *src, Size_t size) {
    Size_t used = __builtin_strlen(dst);
    Size_t length = __builtin_strlen(src);
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

/* tryAMAGICunDEREF, as pp.h has it: the overloaded dereference of *sp goes
 * to the variable sv of the code that uses it. */
#define tryAMAGICunDEREF(meth)                                 \
    STMT_START {                                               \
        sv = amagic_deref_call(*sp, CAT2(meth, _amg));         \
        SPAGAIN;                                               \
    }                                                          \
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
#define PerlLIO_open3(path, flags, perm) open((path), (flags), (perm))
#define PerlLIO_close(fd) close(fd)
#define PerlLIO_read(fd, buf, count) read((fd), (buf), (count))
#define PerlLIO_write(fd, buf, count) write((fd), (buf), (count))
#define PerlLIO_fstat(fd, buf) fstat((fd), (buf))
#define PerlLIO_lseek(fd, offset, mode) lseek((fd), (offset), (mode))
#define PerlLIO_dup(fd) dup(fd)
#define PerlLIO_dup2(fd1, fd2) dup2((fd1), (fd2))
#define PerlLIO_access(path, mode) access((path), (mode))
#define PerlLIO_chmod(path, mode) chmod((path), (mode))
#define PerlLIO_rename(a, b) rename((a), (b))
#define PerlLIO_isatty(fd) isatty(fd)
/* select(2) of this process, as the other system calls here are (a module
 * that waits a while: Devel::NYTProf's ticks_for_usleep). */
#define PerlSock_select(n, r, w, e, t) select((n), (r), (w), (e), (t))
#ifndef O_BINARY
#define O_BINARY 0
#endif
#ifndef O_TEXT
#define O_TEXT 0
#endif

static char *goperl_savepvn(const char *p, size_t n) {
    char *out = (char *)goperl_block_alloc(n + 1, 0);
    if (out) {
        memcpy(out, p, n);
        out[n] = '\0';
    }
    return out;
}
#define savepvn(p, n) goperl_savepvn((p), (size_t)(n))
#define savepv(p) \
    ({ GOPERL_ONCE(goperl_savepv_p_, p); \
       goperl_savepvn((goperl_savepv_p_), __builtin_strlen(goperl_savepv_p_)); })
#define memEQ(a, b, n) (memcmp((a), (b), (n)) == 0)
#define strEQ(a, b) (strcmp((a), (b)) == 0)

/* perl's extended UTF-8, as utf8.h has it. A character is well-formed when
 * its start byte says how many bytes it has (2 to 7, or 13 for 0xFF) and
 * that many are there, the rest of them continuation bytes, and it is not
 * longer than its code point needs (an overlong); its code point may be a
 * surrogate, a non-character or above Unicode, up to UV_MAX. Returns the
 * length of the character at s (whose end is e), or 0 when the bytes there
 * are not a whole well-formed character. The code point goes to *cp. */
static STRLEN goperl_utf8_char(const U8 *s, const U8 *e, UV *cp) {
    if (s >= e) return 0;
    const U8 c = s[0];
    STRLEN n;
    if (c < 0x80) {
        *cp = c;
        return 1;
    }
    if (c < 0xC2) return 0; /* a continuation byte, or an overlong of two */
    if (c < 0xE0) n = 2;
    else if (c < 0xF0) n = 3;
    else if (c < 0xF8) n = 4;
    else if (c < 0xFC) n = 5;
    else if (c < 0xFE) n = 6;
    else if (c == 0xFE) n = 7;
    else n = 13;
    if ((STRLEN)(e - s) < n) return 0;
    for (STRLEN k = 1; k < n; k++)
        if ((s[k] & 0xC0) != 0x80) return 0;
    /* an overlong: the code point would fit the form one byte shorter */
    if ((c == 0xE0 && s[1] < 0xA0) || (c == 0xF0 && s[1] < 0x90) ||
        (c == 0xF8 && s[1] < 0x88) || (c == 0xFC && s[1] < 0x84) ||
        (c == 0xFE && s[1] < 0x82))
        return 0;
    if (c == 0xFF) {
        /* 72 bits: the top 36 zero is an overlong, any of the top 8 set is
         * above UV_MAX */
        int top = 0;
        for (STRLEN k = 1; k <= 6; k++) top |= s[k] & 0x3F;
        if (!top || (s[1] & 0x3F) || (s[2] & 0x30)) return 0;
    }
    UV v = n == 13 ? 0 : (UV)(c & (0x7F >> n));
    for (STRLEN k = 1; k < n; k++) v = (v << 6) | (UV)(s[k] & 0x3F);
    *cp = v;
    return n;
}
static STRLEN goperl_isutf8_char(const U8 *s, const U8 *e) __attribute__((unused));
static STRLEN goperl_isutf8_char(const U8 *s, const U8 *e) {
    UV cp;
    return goperl_utf8_char(s, e, &cp);
}
/* Unicode's own: no surrogate, no code point above Unicode, and (strict)
 * no non-character. */
static STRLEN goperl_isutf8_char_unicode(const U8 *s, const U8 *e, int strict)
    __attribute__((unused));
static STRLEN goperl_isutf8_char_unicode(const U8 *s, const U8 *e, int strict) {
    UV cp;
    const STRLEN n = goperl_utf8_char(s, e, &cp);
    if (!n || cp > 0x10FFFF || (cp >= 0xD800 && cp <= 0xDFFF)) return 0;
    if (strict && ((cp >= 0xFDD0 && cp <= 0xFDEF) || (cp & 0xFFFE) == 0xFFFE))
        return 0;
    return n;
}
#define isUTF8_CHAR(s, e) goperl_isutf8_char((const U8 *)(s), (const U8 *)(e))
#define is_utf8_char_buf(s, e) isUTF8_CHAR((s), (e))
#define isSTRICT_UTF8_CHAR(s, e) \
    goperl_isutf8_char_unicode((const U8 *)(s), (const U8 *)(e), 1)
#define isC9_STRICT_UTF8_CHAR(s, e) \
    goperl_isutf8_char_unicode((const U8 *)(s), (const U8 *)(e), 0)

/* Whether the bytes are well-formed perl extended UTF-8 throughout. */
static int goperl_is_utf8_string(const U8 *s, STRLEN len)
    __attribute__((unused));
static int goperl_is_utf8_string(const U8 *s, STRLEN len) {
    const U8 *const e = s + len;
    while (s < e) {
        const STRLEN n = goperl_isutf8_char(s, e);
        if (!n) return 0;
        s += n;
    }
    return 1;
}
#define is_utf8_string(s, len) goperl_is_utf8_string((const U8 *)(s), (len))

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
 *     baseline copy (the registry entry's base) and dirty fields are written back to the
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
    ({ GOPERL_ONCE(goperl_OP_GIMME_op_, op); \
       (goperl_OP_GIMME_op_->op_flags & OPf_WANT) ? (goperl_OP_GIMME_op_->op_flags & OPf_WANT) : (dfl); })

/* op class + trait bits for PL_opargs: the interpreter's op.h, as the
 * table itself is its opcode.h. */
#define OA_MARK 1
#define OA_FOLDCONST 2
#define OA_RETSCALAR 4
#define OA_TARGET 8
#define OA_TARGLEX 16
#define OA_OTHERINT 32
#define OA_DANGEROUS 64
#define OA_DEFGV 128
#define OCSHIFT 8
#define OA_CLASS_SHIFT OCSHIFT
#define OA_CLASS_MASK (15 << OCSHIFT)
#define OA_BASEOP (0 << OCSHIFT)
#define OA_UNOP (1 << OCSHIFT)
#define OA_BINOP (2 << OCSHIFT)
#define OA_LOGOP (3 << OCSHIFT)
#define OA_LISTOP (4 << OCSHIFT)
#define OA_PMOP (5 << OCSHIFT)
#define OA_SVOP (6 << OCSHIFT)
#define OA_PADOP (7 << OCSHIFT)
#define OA_PVOP_OR_SVOP (8 << OCSHIFT)
#define OA_LOOP (9 << OCSHIFT)
#define OA_COP (10 << OCSHIFT)
#define OA_BASEOP_OR_UNOP (11 << OCSHIFT)
#define OA_FILESTATOP (12 << OCSHIFT)
#define OA_LOOPEXOP (13 << OCSHIFT)
#define OA_METHOP (14 << OCSHIFT)
#define OA_UNOP_AUX (15 << OCSHIFT)
#define OASHIFT 12
#define OA_SCALAR 1
#define OA_AVREF 3
#define OA_HVREF 4
#define OA_CVREF 5
#define OA_FILEREF 6
#define OA_SCALARREF 7
#define OA_OPTIONAL 8

#define CVf_ANON 0x0080

/* ---- PL_parser shadow --------------------------------------------------- */

/* The parser, as the interpreter has it when a parse function or a hook is
 * entered and after every operation that moves it: a copy. What a module
 * writes into bufptr, error_count, in_my, expect, copline and in_my_stash
 * goes back before the next such operation. */
typedef struct yy_parser {
    SV *linestr;   /* guest SV handle */
    char *bufptr;  /* host views into the guest linestr PV */
    char *oldbufptr;
    char *oldoldbufptr;
    char *bufend;
    char *linestart;
    char *last_uni;
    char *last_lop;
    I32 error_count;
    U16 in_my;
    U16 lex_inwhat;
    U32 lex_flags;
    line_t preambling;
    line_t copline;
    U8 expect;
    U8 lex_state;
    I32 lex_brackets;
    I32 lex_formbrack;
    I32 multi_start;
    I32 multi_end;
    UV multi_open;
    UV multi_close;
    HV *in_my_stash;
    PerlIO *rsfp;
    AV *rsfp_filters;
    SV *lex_stuff;        /* the text of a quoted construct being scanned */
    SV *lex_repl;         /* the replacement of an s/// or a tr/// */
    char *lex_brackstack; /* a host view into the interpreter's memory */
    U8 filtered;          /* source filters are applied to the text */
    U16 last_lop_op;
    char tokenbuf[256];
} yy_parser;
/* How the next token is to be read (parser->expect). */
typedef enum {
    XOPERATOR,
    XTERM,
    XREF,
    XSTATE,
    XBLOCK,
    XATTRBLOCK,
    XATTRTERM,
    XTERMBLOCK,
    XBLOCKTERM,
    XPOSTDEREF,
    XTERMORDORDOR
} expectation;

#define LEX_STUFF_UTF8 0x00000001
#define LEX_KEEP_PREVIOUS 0x00000002
#define LEX_IGNORE_UTF8_HINTS 0x00000002 /* lex_flags bit, parser.h */
#define LEX_START_SAME_FILTER 0x00000001
#define LEX_START_FLAGS 0x00000003
#define PARSE_OPTIONAL 0x00000001

/* the parser's fields, as the interpreter answers CALL 15 */
enum {
    GOPERL_PARSERW_LINESTR = 0,
    GOPERL_PARSERW_PV = 1,
    GOPERL_PARSERW_BUFPTR = 2,
    GOPERL_PARSERW_BUFEND = 3,
    GOPERL_PARSERW_OLDBUFPTR = 4,
    GOPERL_PARSERW_OLDOLDBUFPTR = 5,
    GOPERL_PARSERW_LINESTART = 6,
    GOPERL_PARSERW_ERROR_COUNT = 7,
    GOPERL_PARSERW_IN_MY = 8,
    GOPERL_PARSERW_LEX_FLAGS = 9,
    GOPERL_PARSERW_PREAMBLING = 10,
    GOPERL_PARSERW_EXPECT = 11,
    GOPERL_PARSERW_LEX_STATE = 12,
    GOPERL_PARSERW_LEX_BRACKETS = 13,
    GOPERL_PARSERW_LEX_INWHAT = 14,
    GOPERL_PARSERW_MULTI_START = 15,
    GOPERL_PARSERW_MULTI_END = 16,
    GOPERL_PARSERW_MULTI_OPEN = 17,
    GOPERL_PARSERW_MULTI_CLOSE = 18,
    GOPERL_PARSERW_COPLINE = 19,
    GOPERL_PARSERW_IN_MY_STASH = 20,
    GOPERL_PARSERW_LAST_UNI = 21,
    GOPERL_PARSERW_LAST_LOP = 22,
    GOPERL_PARSERW_RSFP = 23,
    GOPERL_PARSERW_RSFP_FILTERS = 24,
    GOPERL_PARSERW_LEX_FORMBRACK = 25,
    GOPERL_PARSERW_LEX_STUFF = 26,
    GOPERL_PARSERW_LEX_REPL = 27,
    GOPERL_PARSERW_LEX_BRACKSTACK = 28,
    GOPERL_PARSERW_TOKENBUF = 29,
    GOPERL_PARSERW_FILTERED = 30,
    GOPERL_PARSERW_LAST_LOP_OP = 31,
    GOPERL_PARSERW_WORDS = 32
};
/* parser field ids of PARSER_SET */
enum {
    GOPERL_PARSER_BUFPTR = 2,
    GOPERL_PARSER_ERROR_COUNT = 6,
    GOPERL_PARSER_IN_MY = 7
};
/* The copy, and what the interpreter held in the fields that are written
 * back (the storage is the loader's, per interpreter). */
typedef struct goperl_parser_state {
    yy_parser parser;
    yy_parser seen;
    uint64_t tokenbuf; /* the address of the interpreter's parser's tokenbuf */
} goperl_parser_state_t;
#define goperl_parser_state_v (*(goperl_parser_state_t **)&GOPERL_SHARED->parser_shadow)
#define goperl_parser_ptr_v (*(yy_parser **)&GOPERL_SHARED->parser_ptr)
#define goperl_parser_pvg_v (GOPERL_SHARED->parser_pvg)

#define goperl_parser_gen_v (GOPERL_SHARED->parser_gen)
/* The interpreter variable of the parser: saved, and cleared (the parser
 * itself is CALL 15's). */
#define GOPERL_PL_PARSER 94
static void goperl_parser_pull(goperl_frame_t *f) {
    /* CALL 15: the parser's fields, or nothing when there is no parser */
    const goperl_gword_t *w = (const goperl_gword_t *)goperl_guest_mem(
        f, goperl_do_op3(f, GOPERL_OP_CALL, 0, 0, (uint64_t)(15), 0, 0));
    goperl_parser_gen_v = GOPERL_SHARED->gen + 1;
    if (!w) {
        goperl_parser_ptr_v = 0;
        /* the copy stands for no parser of the interpreter's (see
         * goperl_parser_push) */
        if (goperl_parser_state_v) goperl_parser_state_v->seen.linestr = 0;
        return;
    }
    goperl_parser_state_t *st = goperl_parser_state_v;
    if (!st) {
        st = (goperl_parser_state_t *)calloc(1, sizeof(goperl_parser_state_t));
        if (!st) goperl_croakf(f, "Out of memory!");
        goperl_parser_state_v = st;
    }
    yy_parser *p = &st->parser;
    goperl_parser_pvg_v = w[GOPERL_PARSERW_PV];
    char *hbase = w[GOPERL_PARSERW_PV]
                      ? (char *)goperl_guest_mem(f, w[GOPERL_PARSERW_PV])
                      : 0;
#define GOPERL_PARSER_PTR(i) (w[i] && hbase ? hbase + (w[i] - w[GOPERL_PARSERW_PV]) : 0)
    p->linestr = GOPERL_SV(w[GOPERL_PARSERW_LINESTR]);
    p->bufptr = GOPERL_PARSER_PTR(GOPERL_PARSERW_BUFPTR);
    p->bufend = GOPERL_PARSER_PTR(GOPERL_PARSERW_BUFEND);
    p->oldbufptr = GOPERL_PARSER_PTR(GOPERL_PARSERW_OLDBUFPTR);
    p->oldoldbufptr = GOPERL_PARSER_PTR(GOPERL_PARSERW_OLDOLDBUFPTR);
    p->linestart = GOPERL_PARSER_PTR(GOPERL_PARSERW_LINESTART);
    p->last_uni = GOPERL_PARSER_PTR(GOPERL_PARSERW_LAST_UNI);
    p->last_lop = GOPERL_PARSER_PTR(GOPERL_PARSERW_LAST_LOP);
#undef GOPERL_PARSER_PTR
    p->error_count = (I32)w[GOPERL_PARSERW_ERROR_COUNT];
    p->in_my = (U16)w[GOPERL_PARSERW_IN_MY];
    p->lex_inwhat = (U16)w[GOPERL_PARSERW_LEX_INWHAT];
    p->lex_flags = (U32)w[GOPERL_PARSERW_LEX_FLAGS];
    p->preambling = (line_t)w[GOPERL_PARSERW_PREAMBLING];
    p->copline = (line_t)w[GOPERL_PARSERW_COPLINE];
    p->expect = (U8)w[GOPERL_PARSERW_EXPECT];
    p->lex_state = (U8)w[GOPERL_PARSERW_LEX_STATE];
    p->lex_brackets = (I32)w[GOPERL_PARSERW_LEX_BRACKETS];
    p->lex_formbrack = (I32)w[GOPERL_PARSERW_LEX_FORMBRACK];
    p->multi_start = (I32)w[GOPERL_PARSERW_MULTI_START];
    p->multi_end = (I32)w[GOPERL_PARSERW_MULTI_END];
    p->multi_open = (UV)w[GOPERL_PARSERW_MULTI_OPEN];
    p->multi_close = (UV)w[GOPERL_PARSERW_MULTI_CLOSE];
    p->in_my_stash = (HV *)GOPERL_SV(w[GOPERL_PARSERW_IN_MY_STASH]);
    p->rsfp = (PerlIO *)(uintptr_t)w[GOPERL_PARSERW_RSFP];
    p->rsfp_filters = (AV *)GOPERL_SV(w[GOPERL_PARSERW_RSFP_FILTERS]);
    p->lex_stuff = GOPERL_SV(w[GOPERL_PARSERW_LEX_STUFF]);
    p->lex_repl = GOPERL_SV(w[GOPERL_PARSERW_LEX_REPL]);
    p->lex_brackstack = (char *)goperl_guest_mem(f, w[GOPERL_PARSERW_LEX_BRACKSTACK]);
    p->filtered = (U8)w[GOPERL_PARSERW_FILTERED];
    p->last_lop_op = (U16)w[GOPERL_PARSERW_LAST_LOP_OP];
    st->tokenbuf = w[GOPERL_PARSERW_TOKENBUF];
    if (st->tokenbuf)
        memcpy(p->tokenbuf, goperl_guest_mem(f, st->tokenbuf), sizeof p->tokenbuf);
    st->seen = *p;
    goperl_parser_ptr_v = p;
}

/* A parser native code made where there was none (Newxz(PL_parser, 1,
 * yy_parser), with copline and preambling set: what newSTATEOP needs at
 * run time) becomes one of the interpreter's, made as that: zeroed, with
 * those two fields, freed with the scope; the copy is made of it again. */
static void goperl_parser_adopt(goperl_frame_t *f, const yy_parser *own) {
    const goperl_gword_t w[2] = {(goperl_gword_t)own->copline, (goperl_gword_t)own->preambling};
    goperl_xs_opx(f, GOPERL_OP_INTERNAL, 0, sizeof w, 19 /* PARSER_NEW */, (const char *)w,
                  sizeof w, 0);
    goperl_parser_ptr_v = 0;
    goperl_parser_gen_v = 0;
}
static void goperl_parser_push(goperl_frame_t *f) {
    yy_parser *p = goperl_parser_ptr_v;
    goperl_parser_state_t *st = goperl_parser_state_v;
    if (p && (!st || p != &st->parser)) {
        goperl_parser_adopt(f, p);
        return;
    }
    /* The module took the parser away (`SAVEVPTR(PL_parser); PL_parser =
     * NULL;`, which Devel::CallParser does around building an op, so that
     * the hooks the op runs do not take it for the lexer's): the
     * interpreter's goes too. The copy was made since the interpreter last
     * ran, from a parser it had (the linestr of what was seen is that
     * parser's); once the interpreter has run, a null copy says nothing of
     * its parser any more (the scope that saved it may have put it back).
     * The save itself is the interpreter's (goperl_save_routed). */
    if (!p && st && st->seen.linestr &&
        goperl_parser_gen_v == GOPERL_SHARED->gen + 1) {
        st->seen.linestr = 0;
        goperl_xs_op(f, GOPERL_OP_PLVAR_SET, GOPERL_PL_PARSER, 0, 0, 0);
        return;
    }
    if (!p || !st) return;
    /* Only what the module assigned goes to the interpreter: a field the
     * module left alone may have moved on there since it was read (a parse
     * that failed counted an error), and must not be put back.
     *
     * This runs before every operation (goperl_flush): what the
     * interpreter does next may depend on any of the fields (the name a
     * new pad entry is declared under says "my" or "state" by in_my). A
     * field is marked as written before the operation that writes it,
     * which flushes in turn. */
    if (p->bufptr != st->seen.bufptr && p->bufptr && goperl_parser_pvg_v) {
        char *hbase = (char *)goperl_guest_mem(f, goperl_parser_pvg_v);
        st->seen.bufptr = p->bufptr;
        goperl_xs_op(f, GOPERL_OP_PARSER_SET, GOPERL_PARSER_BUFPTR,
                     goperl_parser_pvg_v + (uint64_t)(p->bufptr - hbase), 0,
                     0);
    }
    st->seen.bufptr = p->bufptr;
    if (p->error_count != st->seen.error_count) {
        st->seen.error_count = p->error_count;
        goperl_xs_op(f, GOPERL_OP_PARSER_SET, GOPERL_PARSER_ERROR_COUNT,
                     (uint64_t)(uint32_t)p->error_count, 0, 0);
    }
    if (p->in_my != st->seen.in_my) {
        st->seen.in_my = p->in_my;
        goperl_xs_op(f, GOPERL_OP_PARSER_SET, GOPERL_PARSER_IN_MY,
                     (uint64_t)p->in_my, 0, 0);
    }
    /* 100 + the word of the field: the ones a module may also assign */
    if (p->expect != st->seen.expect) {
        st->seen.expect = p->expect;
        goperl_xs_op(f, GOPERL_OP_PARSER_SET, 100 + GOPERL_PARSERW_EXPECT,
                     p->expect, 0, 0);
    }
    if (p->copline != st->seen.copline) {
        st->seen.copline = p->copline;
        goperl_xs_op(f, GOPERL_OP_PARSER_SET, 100 + GOPERL_PARSERW_COPLINE,
                     p->copline, 0, 0);
    }
    if (p->preambling != st->seen.preambling) {
        st->seen.preambling = p->preambling;
        goperl_xs_op(f, GOPERL_OP_PARSER_SET, 100 + GOPERL_PARSERW_PREAMBLING,
                     p->preambling, 0, 0);
    }
    if (p->in_my_stash != st->seen.in_my_stash) {
        st->seen.in_my_stash = p->in_my_stash;
        goperl_xs_op(f, GOPERL_OP_PARSER_SET, 100 + GOPERL_PARSERW_IN_MY_STASH,
                     GOPERL_TOK(p->in_my_stash), 0, 0);
    }
    /* what native code that scans the buffer by itself stores (a lexer of
     * its own, as Devel::Declare has): a pointer is into the interpreter's
     * memory, the buffer's or the bracket stack's */
#define GOPERL_PARSER_PUSH_PTR(word, field)                                      \
    if (p->field != st->seen.field) {                                            \
        const uint64_t g = goperl_guest_addr(f, p->field);                       \
        if (p->field && !g)                                                      \
            goperl_croakf(f, "PL_parser->" #field " is not in the parser's memory"); \
        st->seen.field = p->field;                                               \
        goperl_xs_op(f, GOPERL_OP_PARSER_SET, 100 + (word), g, 0, 0);            \
    }
#define GOPERL_PARSER_PUSH_VAL(word, field, value)                               \
    if (p->field != st->seen.field) {                                            \
        st->seen.field = p->field;                                               \
        goperl_xs_op(f, GOPERL_OP_PARSER_SET, 100 + (word), (uint64_t)(value), 0, 0); \
    }
    GOPERL_PARSER_PUSH_PTR(GOPERL_PARSERW_BUFEND, bufend)
    GOPERL_PARSER_PUSH_PTR(GOPERL_PARSERW_OLDBUFPTR, oldbufptr)
    GOPERL_PARSER_PUSH_PTR(GOPERL_PARSERW_OLDOLDBUFPTR, oldoldbufptr)
    GOPERL_PARSER_PUSH_PTR(GOPERL_PARSERW_LINESTART, linestart)
    GOPERL_PARSER_PUSH_PTR(GOPERL_PARSERW_LAST_UNI, last_uni)
    GOPERL_PARSER_PUSH_PTR(GOPERL_PARSERW_LAST_LOP, last_lop)
    GOPERL_PARSER_PUSH_PTR(GOPERL_PARSERW_LEX_BRACKSTACK, lex_brackstack)
    GOPERL_PARSER_PUSH_VAL(GOPERL_PARSERW_LEX_STUFF, lex_stuff, GOPERL_TOK(p->lex_stuff))
    GOPERL_PARSER_PUSH_VAL(GOPERL_PARSERW_LEX_REPL, lex_repl, GOPERL_TOK(p->lex_repl))
    GOPERL_PARSER_PUSH_VAL(GOPERL_PARSERW_RSFP, rsfp, (uint64_t)(uintptr_t)p->rsfp)
    GOPERL_PARSER_PUSH_VAL(GOPERL_PARSERW_RSFP_FILTERS, rsfp_filters,
                           GOPERL_TOK(p->rsfp_filters))
    GOPERL_PARSER_PUSH_VAL(GOPERL_PARSERW_LEX_STATE, lex_state, p->lex_state)
    GOPERL_PARSER_PUSH_VAL(GOPERL_PARSERW_LEX_BRACKETS, lex_brackets, (uint32_t)p->lex_brackets)
    GOPERL_PARSER_PUSH_VAL(GOPERL_PARSERW_LEX_FORMBRACK, lex_formbrack, (uint32_t)p->lex_formbrack)
    GOPERL_PARSER_PUSH_VAL(GOPERL_PARSERW_LEX_INWHAT, lex_inwhat, p->lex_inwhat)
    GOPERL_PARSER_PUSH_VAL(GOPERL_PARSERW_MULTI_START, multi_start, (uint32_t)p->multi_start)
    GOPERL_PARSER_PUSH_VAL(GOPERL_PARSERW_MULTI_END, multi_end, (uint32_t)p->multi_end)
    GOPERL_PARSER_PUSH_VAL(GOPERL_PARSERW_MULTI_OPEN, multi_open, p->multi_open)
    GOPERL_PARSER_PUSH_VAL(GOPERL_PARSERW_MULTI_CLOSE, multi_close, p->multi_close)
    GOPERL_PARSER_PUSH_VAL(GOPERL_PARSERW_LAST_LOP_OP, last_lop_op, p->last_lop_op)
#undef GOPERL_PARSER_PUSH_PTR
#undef GOPERL_PARSER_PUSH_VAL
    /* the token buffer is the parser's own memory: written in place */
    if (st->tokenbuf && memcmp(p->tokenbuf, st->seen.tokenbuf, sizeof p->tokenbuf)) {
        memcpy(st->seen.tokenbuf, p->tokenbuf, sizeof p->tokenbuf);
        memcpy(goperl_guest_mem(f, st->tokenbuf), p->tokenbuf, sizeof p->tokenbuf);
    }
}

/* PL_parser is the copy of the interpreter's parser, made again when the
 * interpreter has run since it was made (an operation, or native code was
 * entered anew): an op checker or an XSUB called while a file is compiled
 * finds the parser as it is then, as a keyword plugin does. What the
 * module assigned to the copy has reached the interpreter by then
 * (goperl_parser_push runs before every operation). NULL when nothing is
 * being parsed. */
static yy_parser *goperl_parser_now(goperl_frame_t *f) __attribute__((unused));
static yy_parser *goperl_parser_now(goperl_frame_t *f) {
    if (goperl_parser_gen_v != GOPERL_SHARED->gen + 1) goperl_parser_pull(f);
    return goperl_parser_ptr_v;
}
/* An lvalue: a module that finds no parser may make one of its own (see
 * goperl_parser_adopt). */
static yy_parser **goperl_parser_ref(goperl_frame_t *f) __attribute__((unused));
static yy_parser **goperl_parser_ref(goperl_frame_t *f) {
    (void)goperl_parser_now(f);
    return &goperl_parser_ptr_v;
}
#define PL_parser (*goperl_parser_ref(_gof))

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
    U32 cop_line;
    U32 cop_hints;
    char *cop_warnings;
    U32 cop_features;
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
    GOPERL_OPF_MORESIB = 10,
    GOPERL_OPF_PPADDR_OWN = 11 /* the pp function of the op's type again */
};

static void goperl_op_base_capture(OP *o) {
    goperl_opreg_ent_t *const e = GOPERL_OPENT(o);
    if (!e->base)
        e->base = (struct goperl_op_base *)calloc(1, sizeof(struct goperl_op_base));
    struct goperl_op_base *b = e->base;
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
    b->cop_line = o->cop_line;
    b->cop_hints = o->cop_hints;
    b->cop_warnings = o->cop_warnings;
    b->cop_features = o->cop_features.bits[0];
}

static void goperl_perop_pp_register(goperl_frame_t *f, OP *o);
static uint32_t goperl_xop_number(Perl_ppaddr_t pp);

static void goperl_op_setfield(goperl_frame_t *f, OP *o, int sel,
                               uint64_t val) {
    goperl_do_op3(f, GOPERL_OP_OP_SET, o->gop, val, (uint64_t)(uint32_t)sel, 0, 0);
}

/* The interpreter's op a link of a written-back op leads to. */
static uint64_t goperl_op_link_tok(goperl_frame_t *f, OP *kid) {
    if (!kid) return 0;
    if (!kid->gop)
        goperl_croakf(f, "optree write-back: scratch op (no guest identity) "
                         "linked into a guest tree");
    return kid->gop;
}

static void goperl_op_flush_one(goperl_frame_t *f, OP *o) {
    struct goperl_op_base *b = GOPERL_OPENT(o)->base;
    if (!o->gop || !b) return;
    /* in a module's own op struct only the links it has are links */
    const int links = GOPERL_OPENT(o)->persist ? goperl_raw_links(o) : 2;
    if (b->op_next != o->op_next)
        goperl_op_setfield(f, o, GOPERL_OPF_NEXT,
                           goperl_op_link_tok(f, o->op_next));
    if (b->op_sibparent != o->op_sibparent || b->op_moresib != o->op_moresib)
        goperl_op_setfield(f, o,
                           GOPERL_OPF_SIBPARENT |
                               ((int)(o->op_moresib ? 1 : 0) << 16),
                           goperl_op_link_tok(f, o->op_sibparent));
    if (links >= 1 && b->op_first != o->op_first)
        goperl_op_setfield(f, o, GOPERL_OPF_FIRST,
                           goperl_op_link_tok(f, o->op_first));
    if (links >= 2 && b->op_last != o->op_last)
        goperl_op_setfield(f, o, GOPERL_OPF_LAST,
                           goperl_op_link_tok(f, o->op_last));
    if (links >= 2 && b->op_other != o->op_other)
        goperl_op_setfield(f, o, GOPERL_OPF_OTHER,
                           goperl_op_link_tok(f, o->op_other));
    if (b->op_targ != o->op_targ)
        goperl_op_setfield(f, o, GOPERL_OPF_TARG, (uint64_t)o->op_targ);
    if (b->op_type != o->op_type)
        goperl_op_setfield(f, o, GOPERL_OPF_TYPE, o->op_type);
    if (b->op_flags != o->op_flags)
        goperl_op_setfield(f, o, GOPERL_OPF_FLAGS, o->op_flags);
    if (b->op_private != o->op_private)
        goperl_op_setfield(f, o, GOPERL_OPF_PRIVATE, o->op_private);
    if (b->op_ppaddr != o->op_ppaddr) goperl_perop_pp_register(f, o);
    if (o->gop_iscop) {
        /* COP_SET: 0 the line, 1 the warnings, 2 the hints */
        if (b->cop_line != o->cop_line)
            goperl_do_op3(f, GOPERL_OP_CALL, o->gop, o->cop_line, 6 | ((uint64_t)0 << 32), 0, 0);
        if (b->cop_warnings != o->cop_warnings) {
            /* a warnings buffer is the interpreter's memory, or one of its
             * two markers, or NULL */
            const uintptr_t d = (uintptr_t)o->cop_warnings - (uintptr_t)f->membase;
            if (o->cop_warnings && d >= *f->memsize)
                goperl_croakf(f, "cop_warnings was given a buffer that is not "
                                 "the interpreter's");
            goperl_do_op3(f, GOPERL_OP_CALL, o->gop, (uint64_t)(o->cop_warnings ? d : 0),
                          6 | ((uint64_t)1 << 32), 0, 0);
        }
        if (b->cop_hints != o->cop_hints)
            goperl_do_op3(f, GOPERL_OP_CALL, o->gop, o->cop_hints, 6 | ((uint64_t)2 << 32), 0, 0);
        if (b->cop_features != o->cop_features.bits[0]) /* 5: the feature bits */
            goperl_do_op3(f, GOPERL_OP_CALL, o->gop, o->cop_features.bits[0],
                          6 | ((uint64_t)5 << 32), 0, 0);
    }
    goperl_op_base_capture(o);
}

static void goperl_optree_flush(goperl_frame_t *f) {
    goperl_opreg_t *r = goperl_opreg_v;
    if (!r) return;
    for (goperl_opreg_ent_t *e = r->active; e; e = e->active_next)
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
    if (o->op_type < OP_max && o->op_ppaddr == goperl_ppaddr_get(f)[o->op_type]) {
        /* the pp function of its type again */
        goperl_api_perop_pp_set(f, o->gop, 0);
        goperl_op_setfield(f, o, GOPERL_OPF_PPADDR_OWN, 0);
        return;
    }
    goperl_api_perop_pp_set(f, o->gop, (void *)o->op_ppaddr);
    /* the number of the registered custom op, when the function is one */
    goperl_op_setfield(f, o, GOPERL_OPF_PPADDR_HOOK,
                       goperl_xop_number(o->op_ppaddr));
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
    GOPERL_OPC_SLICEOP = 11,
    /* 12 is GOPERL_OPC_RAW, below */
    GOPERL_OPC_ASSIGNOP = 13
};

/* Which operands of a constructor are ops. */
#define GOPERL_OPERAND_OP1 1
#define GOPERL_OPERAND_OP2 2
static OP *goperl_newop_of(goperl_frame_t *f, int cls, I32 type, I32 flags,
                           uint64_t k1, uint64_t k2, int operand_ops) {
    char kids[16];
    memcpy(kids, &k1, 8);
    memcpy(kids + 8, &k2, 8);
    goperl_optree_flush(f);
    uint64_t tok = goperl_do_op(
        f, GOPERL_OP_OPTREE_NEW,
        ((uint64_t)(uint32_t)cls << 32) | (uint32_t)type,
        (uint64_t)(uint32_t)flags, kids, 16);
    /* the constructor relinked its kids (sibling/parent) guest-side */
    if (operand_ops & GOPERL_OPERAND_OP1) goperl_op_refresh_tok(f, k1);
    if (operand_ops & GOPERL_OPERAND_OP2) goperl_op_refresh_tok(f, k2);
    return goperl_op_shadow_full(f, tok);
}
/* A constructor whose operands are ops. */
static inline __attribute__((unused)) OP *goperl_newop(
    goperl_frame_t *f, int cls, I32 type, I32 flags, uint64_t k1,
    uint64_t k2) {
    return goperl_newop_of(f, cls, type, flags, k1, k2,
                           GOPERL_OPERAND_OP1 | GOPERL_OPERAND_OP2);
}

#define newOP(type, flags) goperl_newop(_gof, GOPERL_OPC_BASEOP, (type), (flags), 0, 0)
#define newUNOP(type, flags, first) \
    ({ GOPERL_ONCE(goperl_newUNOP_first_, first); \
       goperl_newop(_gof, GOPERL_OPC_UNOP, (type), (flags), (goperl_newUNOP_first_) ? ((OP *)(goperl_newUNOP_first_))->gop : 0, 0); })
#define newBINOP(type, flags, first, last) \
    ({ GOPERL_ONCE(goperl_newBINOP_first_, first); GOPERL_ONCE(goperl_newBINOP_last_, last); \
       goperl_newop(_gof, GOPERL_OPC_BINOP, (type), (flags), (goperl_newBINOP_first_) ? ((OP *)(goperl_newBINOP_first_))->gop : 0, (goperl_newBINOP_last_) ? ((OP *)(goperl_newBINOP_last_))->gop : 0); })
#define newLISTOP(type, flags, first, last) \
    ({ GOPERL_ONCE(goperl_newLISTOP_first_, first); GOPERL_ONCE(goperl_newLISTOP_last_, last); \
       goperl_newop(_gof, GOPERL_OPC_LISTOP, (type), (flags), (goperl_newLISTOP_first_) ? ((OP *)(goperl_newLISTOP_first_))->gop : 0, (goperl_newLISTOP_last_) ? ((OP *)(goperl_newLISTOP_last_))->gop : 0); })
#define newLOGOP(type, flags, first, other) \
    ({ GOPERL_ONCE(goperl_newLOGOP_first_, first); GOPERL_ONCE(goperl_newLOGOP_other_, other); \
       goperl_newop(_gof, GOPERL_OPC_LOGOP, (type), (flags), (goperl_newLOGOP_first_) ? ((OP *)(goperl_newLOGOP_first_))->gop : 0, (goperl_newLOGOP_other_) ? ((OP *)(goperl_newLOGOP_other_))->gop : 0); })
/* An assignment: optype is the operator of an assignment operator (OP_ADD
 * for +=), 0 for plain assignment. */
#define newASSIGNOP(flags, left, optype, right) \
    ({ GOPERL_ONCE(goperl_newASSIGNOP_left_, left); GOPERL_ONCE(goperl_newASSIGNOP_right_, right); \
       goperl_newop(_gof, GOPERL_OPC_ASSIGNOP, (optype), (flags), (goperl_newASSIGNOP_left_) ? ((OP *)(goperl_newASSIGNOP_left_))->gop : 0, (goperl_newASSIGNOP_right_) ? ((OP *)(goperl_newASSIGNOP_right_))->gop : 0); })
#define newSVOP(type, flags, sv) \
    goperl_newop_of(_gof, GOPERL_OPC_SVOP, (type), (flags), GOPERL_TOK(sv), 0, 0)
#define newGVOP(type, flags, gv) \
    goperl_newop_of(_gof, GOPERL_OPC_GVOP, (type), (flags), GOPERL_TOK(gv), 0, 0)
#define newMETHOP_named(type, flags, name_sv) \
    goperl_newop_of(_gof, GOPERL_OPC_METHOP_NAMED, (type), (flags), \
                    GOPERL_TOK(name_sv), 0, 0)
#define newSTATEOP(flags, label, o) \
    ({ GOPERL_ONCE(goperl_newSTATEOP_o_, o); \
       goperl_newop_of(_gof, GOPERL_OPC_STATEOP, 0, (flags), (uint64_t)(uintptr_t)(label), (goperl_newSTATEOP_o_) ? ((OP *)(goperl_newSTATEOP_o_))->gop : 0, GOPERL_OPERAND_OP2); })
#define newCONDOP(flags, cond, tru, fls) \
    goperl_newcondop(_gof, (flags), (OP *)(cond), (OP *)(tru), (OP *)(fls))
#define newSLICEOP(flags, subscript, listval) \
    ({ GOPERL_ONCE(goperl_newSLICEOP_subscript_, subscript); GOPERL_ONCE(goperl_newSLICEOP_listval_, listval); \
       goperl_newop(_gof, GOPERL_OPC_SLICEOP, 0, (flags), (goperl_newSLICEOP_subscript_) ? ((OP *)(goperl_newSLICEOP_subscript_))->gop : 0, (goperl_newSLICEOP_listval_) ? ((OP *)(goperl_newSLICEOP_listval_))->gop : 0); })

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
/* newUNOP_AUX is defined with the functions that build op trees (see
 * goperl_newunop_aux). */

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
    if ((sel == GOPERL_OPM_LINKLIST || sel == GOPERL_OPM_CONTEXTUALIZE) && a) {
        /* the interpreter set the links of execution (op_next) or the
         * context of every op below a, which code reads next: `start =
         * LINKLIST(o); dup->op_next = kid->op_next` */
        goperl_op_describe(f, a->gop, GOPERL_OPDUMP_SUBTREE);
    } else {
        goperl_op_refresh(f, a);
    }
    goperl_op_refresh(f, b);
    if (tok && (sel == GOPERL_OPM_APPEND_ELEM || sel == GOPERL_OPM_APPEND_LIST ||
                sel == GOPERL_OPM_PREPEND_ELEM || sel == GOPERL_OPM_CONVERT_LIST ||
                sel == GOPERL_OPM_SCOPE))
        /* the list the result is was relinked below it as well (the kid
         * that was last has a sibling now): every op below it is described
         * again, as code walks the siblings next */
        return goperl_op_describe(f, tok, GOPERL_OPDUMP_SUBTREE);
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
    /* the kids of the parent were relinked, the ones around the splice
     * point and the ones that went in and came out: the parent is described
     * again with everything below it, and so is each op that came out */
    if (parent) goperl_op_describe(f, parent->gop, GOPERL_OPDUMP_SUBTREE);
    goperl_op_refresh(f, start);
    goperl_op_refresh(f, insert);
    OP *const out = tok ? goperl_op_describe(f, tok, GOPERL_OPDUMP_SUBTREE) : 0;
    for (OP *o = out; o && o->op_moresib && o->op_sibparent; o = o->op_sibparent)
        goperl_op_describe(f, o->op_sibparent->gop, GOPERL_OPDUMP_SUBTREE);
    return out;
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
    ({ GOPERL_ONCE(goperl_lex_stuff_pvn_len_, len); \
       ((void)goperl_lex(_gof, GOPERL_LEX_STUFF_PVN, ((uint64_t)(uint32_t)(flags) << 32) | (uint32_t)(goperl_lex_stuff_pvn_len_), (pv), (goperl_lex_stuff_pvn_len_))); })
#define lex_stuff_pvs(pv, flags) lex_stuff_pvn("" pv "", sizeof(pv) - 1, (flags))

static void goperl_lex_read_to(goperl_frame_t *f, char *ptr) {
    yy_parser *p = goperl_parser_ptr_v;
    if (!p || !p->bufptr || ptr < p->bufptr || ptr > p->bufend)
        goperl_croakf(f, "lex_read_to: pointer outside the lexer buffer");
    char *hbase = (char *)goperl_guest_mem(f, goperl_parser_pvg_v);
    goperl_lex(f, GOPERL_LEX_READ_TO,
               goperl_parser_pvg_v + (uint64_t)(ptr - hbase), 0, 0);
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
#define Perl_block_start(...) goperl_block_start(_gof, __VA_ARGS__)
#define Perl_block_end(...) goperl_block_end(_gof, __VA_ARGS__)

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
#define Perl_pad_alloc(...) goperl_pad_alloc(_gof, __VA_ARGS__)
#define pad_add_name_pvn(name, len, flags, typestash, ourstash) \
    goperl_pad_add_name_pvn(_gof, (name), (len), (flags), (typestash), (ourstash))
#define pad_findmy_pvn(name, len, flags) \
    goperl_pad_findmy_pvn(_gof, (name), (len), (flags))
#define intro_my() \
    ((U32)goperl_do_op(_gof, GOPERL_OP_PAD, GOPERL_PAD_INTRO_MY, 0, 0, 0))
static U32 goperl_intro_my(goperl_frame_t *f) {
    return (U32)goperl_do_op(f, GOPERL_OP_PAD, GOPERL_PAD_INTRO_MY, 0, 0, 0);
}
#define Perl_intro_my() goperl_intro_my(_gof)
#define padadd_NO_DUP_CHECK 0x04
#define padadd_STATE 0x02

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
/* PL_keyword_plugin is the head of the chain: perl's own plugin, which
 * declines every word, until a module puts one of its own there, keeping
 * the one it found to call for the words it declines (the way of modules
 * older than wrap_keyword_plugin). The interpreter hands its words to the
 * chain from the moment a module looks at it. */
static Perl_keyword_plugin_t *goperl_kw_plugin_ref(goperl_frame_t *f) {
    if (!goperl_kw_head_v) goperl_kw_head_v = goperl_kw_decline;
    goperl_do_op(f, GOPERL_OP_KEYWORD_ENABLE, 0, 0, 0, 0);
    return &goperl_kw_head_v;
}
#define PL_keyword_plugin (*goperl_kw_plugin_ref(_gof))

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

/* Infix plugins (5.38+ core surface): the interpreter's lexer offers each
 * candidate operator to the chain; an operator a module claims is given to
 * the interpreter with its precedence, and the interpreter has the host run
 * the module's parse and build_op for it (GOPERL_HOST_INFIX). */
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

/* Loader entry: run the host keyword-plugin chain for one candidate word
 * (kind 0 a keyword, 1 an infix operator). Returns 1 when a plugin took it,
 * with its result in *ret_out (the length of an infix operator) and the op
 * (the module's struct of an infix operator) in *op_out, 0 when none did,
 * -1 on a croak (the frame says how). */
GOPERL_C __attribute__((weak, visibility("default"), used)) int32_t
__goperl_keyword_invoke(goperl_frame_t *f, const char *word, uint32_t wordlen,
                        uint32_t kind, uint32_t *ret_out, uint64_t *op_out) {
    jmp_buf jb;
    void *prev_jb = f->jb;
    goperl_frame_t *prev_frame = goperl_cur_frame_v;
    f->jb = (void *)&jb;
    f->prev_frame = (void *)prev_frame;
    goperl_frame_enter(f);
    goperl_xs_depth_v++;
    f->hostsave_base = goperl_hostsave_n;
    goperl_opreg_enter();
    f->native_call = 1;
    if (GOPERL_SETJMP(jb)) {
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
        int ret = goperl_kw_head_v(wbuf, wordlen, &op);
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
        const STRLEN n = goperl_infix_head_v(wbuf, wordlen, &inf);
        if (n != 0) {
            if (!inf || !inf->build_op)
                goperl_croakf(f, "an infix plugin claimed '%s' without a build_op", wbuf);
            /* the loader reads the precedence and whether there is a parse
             * from the struct */
            *ret_out = (uint32_t)n;
            *op_out = (uint64_t)(uintptr_t)inf;
            rc = 1;
        }
    }
    goperl_optree_flush(f);
    goperl_hostsave_settle(f, f->hostsave_base, __builtin_frame_address(0));
    goperl_flush(f);
    goperl_xs_depth_v--;
    goperl_cur_frame_v = prev_frame;
    f->jb = prev_jb;
    return rc;
}

/* ---- op checkers, peephole hooks, call_atexit ----------------------------
 * PL_check is a table of the functions the compiler hands every new op of
 * a type to, and PL_peepp and PL_rpeepp the functions it hands a finished
 * sub to; a module puts its own there and calls the one it found. Here
 * each is host storage that starts out with stubs: a stub runs the
 * interpreter's function (its checker of the type, its peephole
 * optimizer), and an entry that is not the stub is a module's function,
 * which the interpreter is then told to call. It does that through the
 * host-call channel, as it runs a function given to call_atexit when the
 * interpreter is destroyed. */
typedef OP *(*Perl_check_t)(pTHX_ OP *);
typedef U16 Optype;
typedef void (*peep_t)(pTHX_ OP *o);
typedef void (*ATEXIT_t)(pTHX_ void *);
enum {
    GOPERL_HOOK_NEWOP_SET = 0,
    GOPERL_HOOK_NEWOP_ORIG = 1,
    GOPERL_HOOK_PEEP_SET = 2,
    GOPERL_HOOK_PEEP_ORIG = 3,
    GOPERL_HOOK_ATEXIT = 4
};
enum {
    GOPERL_HOST_NEWOP = 0,
    GOPERL_HOST_PEEP = 1,
    GOPERL_HOST_ATEXIT = 2,
    GOPERL_HOST_CALLCK = 3,
    GOPERL_HOST_MAGIC = 4,
    GOPERL_HOST_UVAR = 5,
    GOPERL_HOST_INFIX = 9
};
/* The hooks of GOPERL_HOST_MAGIC (free and local come another way). */
enum {
    GOPERL_MGHOOK_GET = 0,
    GOPERL_MGHOOK_SET = 1,
    GOPERL_MGHOOK_LEN = 2,
    GOPERL_MGHOOK_CLEAR = 3,
    GOPERL_MGHOOK_COPY = 4
};
typedef OP *(*Perl_call_checker)(pTHX_ OP *o, GV *namegv, SV *ckobj);
/* The call checkers modules installed, by the id the interpreter keeps
 * with the sub (index + 1). */
typedef struct goperl_callcks {
    Perl_call_checker *fns;
    int32_t n;
    int32_t cap;
} goperl_callcks_t;
#define goperl_callcks_v (*(goperl_callcks_t **)&GOPERL_SHARED->callcks)

typedef struct goperl_atexit {
    ATEXIT_t fn; /* NULL once it ran */
    void *arg;
    /* NULL, or how many instances hold a copy of the entry: the copies of
     * an instance share it (see __goperl_shared_clone) */
    int32_t *holders;
} goperl_atexit_t;
/* The description of a custom op, laid out as perl's (a module may
 * initialize one statically, flags first): a field counts when its flag is
 * set, which XopENTRY_set does. */
typedef struct xop {
    U32 xop_flags;
    const char *xop_name;
    const char *xop_desc;
    U32 xop_class;
    void (*xop_peep)(pTHX_ OP *o, OP *oldop);
} XOP;
#define GOPERL_XOP_MAX 256
typedef struct goperl_xop_ent {
    Perl_ppaddr_t pp;
    const XOP *xop;
    uint32_t number; /* the interpreter's number of the registered op */
} goperl_xop_ent_t;
typedef struct goperl_hooks {
    goperl_xop_ent_t xops[GOPERL_XOP_MAX]; /* the custom ops modules registered */
    int32_t xop_n;
    int32_t pad;
    Perl_check_t newop[OP_max];      /* PL_check */
    Perl_check_t newop_stub[OP_max]; /* what it started out with */
    Perl_check_t newop_told[OP_max]; /* what the interpreter was last told of */
    peep_t peep[2];                  /* PL_peepp, PL_rpeepp */
    peep_t peep_stub[2];
    peep_t peep_told[2];
    int32_t touched; /* a module has looked at one of them in this native call */
    int32_t atexit_n;
    int32_t atexit_cap;
    goperl_atexit_t *atexit;
} goperl_hooks_t;
#define goperl_hooks_v (*(goperl_hooks_t **)&GOPERL_SHARED->hooks)

/* The interpreter's number of the custom op registered for a pp function,
 * 0 when none was. */
static uint32_t goperl_xop_number(Perl_ppaddr_t pp) {
    const goperl_hooks_t *h = goperl_hooks_v;
    for (int32_t i = 0; h && i < h->xop_n; i++)
        if (h->xops[i].pp == pp) return h->xops[i].number;
    return 0;
}

static void goperl_peep_stub(goperl_frame_t *f, int which, OP *o) {
    if (o && !o->gop)
        goperl_croakf(f, "the peephole optimizer was handed an op that is not "
                         "the interpreter's");
    goperl_optree_flush(f);
    goperl_do_op3(f, GOPERL_OP_HOOK, GOPERL_HOOK_PEEP_ORIG, o ? o->gop : 0,
                  (uint64_t)(uint32_t)which, 0, 0);
    /* the optimizer rewrites what it is handed */
    if (o) goperl_op_shadow_reach(f, o->gop);
}
static void goperl_peep_stub0(pTHX_ OP *o) { goperl_peep_stub(_gof, 0, o); }
static void goperl_peep_stub1(pTHX_ OP *o) { goperl_peep_stub(_gof, 1, o); }

/* The stub of an entry of PL_check: the checker the interpreter has for
 * the type. It may hand back another op than it was given. */
static OP *goperl_newop_dispatch(pTHX_ int type, OP *o) {
    if (o && !o->gop)
        goperl_croakf(_gof, "an op checker was handed an op that is not the "
                            "interpreter's");
    goperl_optree_flush(_gof);
    uint64_t tok = goperl_do_op3(_gof, GOPERL_OP_HOOK, GOPERL_HOOK_NEWOP_ORIG, o ? o->gop : 0,
                                 (uint64_t)(uint32_t)type, 0, 0);
    return tok ? goperl_op_describe(_gof, tok, GOPERL_OPDUMP_SUBTREE) : 0;
}

static goperl_hooks_t *goperl_hooks_get(goperl_frame_t *f) {
    goperl_hooks_t *h = goperl_hooks_v;
    if (!h) {
        h = (goperl_hooks_t *)calloc(1, sizeof(goperl_hooks_t));
        if (!h) goperl_croakf(f, "Out of memory!");
        GOPERL_NEWOP_SLOT_INIT(h->newop);
        memcpy(h->newop_stub, h->newop, sizeof(h->newop));
        memcpy(h->newop_told, h->newop, sizeof(h->newop));
        h->peep[0] = h->peep_stub[0] = h->peep_told[0] = goperl_peep_stub0;
        h->peep[1] = h->peep_stub[1] = h->peep_told[1] = goperl_peep_stub1;
        goperl_hooks_v = h;
    }
    /* an entry may be assigned any time from here on, also after guest
     * operations of the same expression */
    h->touched = 1;
    return h;
}
#define PL_check (goperl_hooks_get(_gof)->newop)
#define PL_peepp (goperl_hooks_get(_gof)->peep[0])
#define PL_rpeepp (goperl_hooks_get(_gof)->peep[1])

/* Tells the interpreter which entries are a module's now. */
static void goperl_hooks_sync(goperl_frame_t *f) {
    goperl_hooks_t *h = goperl_hooks_v;
    if (!h || !h->touched) return;
    for (int i = 0; i < OP_max; i++) {
        if (h->newop[i] == h->newop_told[i]) continue;
        h->newop_told[i] = h->newop[i];
        goperl_xs_op(f, GOPERL_OP_HOOK, GOPERL_HOOK_NEWOP_SET,
                     ((uint64_t)i << 1) | (h->newop[i] != h->newop_stub[i] ? 1 : 0),
                     0, 0);
    }
    for (int i = 0; i < 2; i++) {
        if (h->peep[i] == h->peep_told[i]) continue;
        h->peep_told[i] = h->peep[i];
        goperl_xs_op(f, GOPERL_OP_HOOK, GOPERL_HOOK_PEEP_SET,
                     ((uint64_t)i << 1) | (h->peep[i] != h->peep_stub[i] ? 1 : 0),
                     0, 0);
    }
}
/* The outermost native call has returned: no expression is left that could
 * still assign an entry. */
static void goperl_hooks_rest(goperl_frame_t *f) {
    goperl_hooks_t *h = goperl_hooks_v;
    if (!h) return;
    goperl_hooks_sync(f);
    h->touched = 0;
}

static void goperl_wrap_newop_hook(goperl_frame_t *f, Optype opcode,
                                   Perl_check_t hook,
                                   Perl_check_t *previous_p)
    __attribute__((unused));
static void goperl_wrap_newop_hook(goperl_frame_t *f, Optype opcode,
                                   Perl_check_t hook,
                                   Perl_check_t *previous_p) {
    if (*previous_p) return;
    goperl_hooks_t *h = goperl_hooks_get(f);
    *previous_p = h->newop[opcode];
    h->newop[opcode] = hook;
}
#define wrap_op_checker(opcode, hook, previous_p) \
    goperl_wrap_newop_hook(_gof, (opcode), (hook), (previous_p))
#define Perl_wrap_op_checker(...) goperl_wrap_newop_hook(_gof, __VA_ARGS__)

static void goperl_call_atexit(goperl_frame_t *f, ATEXIT_t fn, void *arg)
    __attribute__((unused));
static void goperl_call_atexit(goperl_frame_t *f, ATEXIT_t fn, void *arg) {
    goperl_hooks_t *h = goperl_hooks_get(f);
    if (h->atexit_n == h->atexit_cap) {
        const int32_t ncap = h->atexit_cap ? h->atexit_cap * 2 : 8;
        goperl_atexit_t *na = (goperl_atexit_t *)realloc(
            h->atexit, (size_t)ncap * sizeof(goperl_atexit_t));
        if (!na) goperl_croakf(f, "Out of memory!");
        h->atexit = na;
        h->atexit_cap = ncap;
    }
    h->atexit[h->atexit_n].fn = fn;
    h->atexit[h->atexit_n].arg = arg;
    h->atexit[h->atexit_n].holders = 0;
    h->atexit_n++;
    goperl_do_op(f, GOPERL_OP_HOOK, GOPERL_HOOK_ATEXIT, (uint64_t)h->atexit_n,
                 0, 0);
}
#define call_atexit(fn, arg) goperl_call_atexit(_gof, (fn), (arg))
#define Perl_call_atexit(...) goperl_call_atexit(_gof, __VA_ARGS__)
#define perl_atexit(fn, arg) call_atexit((fn), (arg))

static int32_t goperl_host_invoke_fn(goperl_frame_t *f, const goperl_gword_t *req,
                                     uint32_t reqlen, goperl_gword_t *resp);
/* Loader entry: one call of native code through the host-call channel.
 * req is [kind] and the kind's words; the answer goes to resp, in words,
 * and its length in bytes is returned (reqlen and respcap are in bytes
 * too). On a croak the frame carries the message. */
GOPERL_C __attribute__((weak, visibility("default"), used)) int32_t
__goperl_host_invoke(goperl_frame_t *f, const goperl_gword_t *req, uint32_t reqlen,
                     goperl_gword_t *resp, uint32_t respcap) {
    jmp_buf jb;
    int32_t rlen = 0;
    f->jb = (void *)&jb;
    f->prev_frame = (void *)goperl_cur_frame_v;
    goperl_frame_enter(f);
    goperl_xs_depth_v++;
    goperl_opreg_enter();
    f->native_call = 1;
    f->hostsave_base = goperl_hostsave_n;
    if (GOPERL_SETJMP(jb)) {
        goperl_hostsave_unwind_to(f, f->hostsave_base);
        goperl_xs_leave(f);
        return 0;
    }
    if (reqlen < 8 || respcap < 16) goperl_croakf(f, "host call: short request");
    goperl_hooks_t *h = goperl_hooks_v;
    switch (req[0]) {
    case GOPERL_HOST_NEWOP: {
        if (reqlen < 24 || req[1] >= OP_max || !h)
            goperl_croakf(f, "host call: no op checker for the request");
        OP *o = goperl_op_describe(f, req[2], GOPERL_OPDUMP_SUBTREE);
        OP *r = h->newop[req[1]](o);
        if (r && !r->gop)
            goperl_croakf(f, "an op checker returned an op that is not the "
                             "interpreter's");
        goperl_optree_flush(f);
        resp[0] = r ? r->gop : 0;
        rlen = 8;
        break;
    }
    case GOPERL_HOST_PEEP: {
        if (reqlen < 24 || req[1] > 1 || !h)
            goperl_croakf(f, "host call: no peephole hook for the request");
        OP *o = req[2] ? goperl_op_shadow_reach(f, req[2]) : 0;
        h->peep[req[1]](o);
        goperl_optree_flush(f);
        break;
    }
    case GOPERL_HOST_ATEXIT: {
        if (reqlen < 16 || !h || req[1] == 0 || (int64_t)req[1] > h->atexit_n)
            goperl_croakf(f, "host call: no function for the request");
        goperl_atexit_t *const slot = &h->atexit[req[1] - 1];
        const goperl_atexit_t a = *slot;
        slot->fn = 0;
        slot->holders = 0;
        /* one the copies of the instance share runs for the last of them:
         * it releases what the module keeps for all of them */
        if (a.holders) {
            if (__atomic_sub_fetch(a.holders, 1, __ATOMIC_ACQ_REL) > 0) break;
            free(a.holders);
        }
        if (a.fn) a.fn(a.arg);
        break;
    }
    case GOPERL_HOST_CALLCK: {
        const goperl_callcks_t *c = goperl_callcks_v;
        if (reqlen < 40 || !c || req[1] == 0 || (int64_t)req[1] > c->n)
            goperl_croakf(f, "host call: no call checker for the request");
        OP *o = goperl_op_describe(f, req[2], GOPERL_OPDUMP_SUBTREE);
        OP *r = c->fns[req[1] - 1](o, (GV *)GOPERL_SV(req[3]), GOPERL_SV(req[4]));
        if (r && !r->gop)
            goperl_croakf(f, "a call checker returned an op that is not the "
                             "interpreter's");
        goperl_optree_flush(f);
        resp[0] = r ? r->gop : 0;
        rlen = 8;
        break;
    }
    case GOPERL_HOST_MAGIC: {
        /* [hook][sv][nsv][name][namlen]: the hook of every MAGIC of the
         * chain that has it; a hook may take its MAGIC off the chain */
        if (reqlen < 48) goperl_croakf(f, "host call: short magic request");
        SV *sv = GOPERL_SV(req[2]);
        goperl_gword_t result = 0;
        goperl_flush(f);
        for (MAGIC *mg = goperl_api_magic_chain(f, req[2]), *next; mg; mg = next) {
            next = mg->mg_moremagic;
            const MGVTBL *v = mg->mg_virtual;
            if (!v) continue;
            switch (req[1]) {
            case GOPERL_MGHOOK_GET:
                if (v->svt_get) result = (uint32_t)v->svt_get(sv, mg);
                break;
            case GOPERL_MGHOOK_SET:
                if (v->svt_set) result = (uint32_t)v->svt_set(sv, mg);
                break;
            case GOPERL_MGHOOK_LEN:
                if (v->svt_len) result = v->svt_len(sv, mg);
                break;
            case GOPERL_MGHOOK_CLEAR:
                if (v->svt_clear) result = (uint32_t)v->svt_clear(sv, mg);
                break;
            case GOPERL_MGHOOK_COPY:
                if (v->svt_copy && (mg->mg_flags & MGf_COPY)) {
                    const I32 namlen = (I32)req[5];
                    /* the name is an SV, or bytes of the interpreter's */
                    const char *name =
                        namlen == HEf_SVKEY
                            ? (const char *)GOPERL_SV(req[4])
                            : (req[4] ? (const char *)goperl_guest_mem(f, req[4]) : 0);
                    result = (uint32_t)v->svt_copy(sv, mg, GOPERL_SV(req[3]), name, namlen);
                }
                break;
            default:
                goperl_croakf(f, "host call: unknown magic hook %u", (unsigned)req[1]);
            }
        }
        resp[0] = result;
        rlen = 8;
        break;
    }
    case GOPERL_HOST_INFIX: {
        /* [0][def][opdata] -> [opdata]: the operator's parse, with the
         * lexer after the operator; [1][def][opdata][lhs][rhs] -> [opdata]
         * [op]: its build_op */
        if (reqlen < 32) goperl_croakf(f, "host call: short infix request");
        struct Perl_custom_infix *def = (struct Perl_custom_infix *)(uintptr_t)req[2];
        SV *opdata = GOPERL_SV(req[3]);
        OP *o = 0;
        if (req[1] == 0) {
            if (!def->parse) goperl_croakf(f, "host call: the infix operator has no parse");
            goperl_parser_pull(f);
            def->parse(&opdata, def);
            goperl_optree_flush(f);
            goperl_parser_push(f);
        } else {
            if (reqlen < 48) goperl_croakf(f, "host call: short infix request");
            OP *lhs = req[4] ? goperl_op_describe(f, req[4], GOPERL_OPDUMP_SUBTREE) : 0;
            OP *rhs = req[5] ? goperl_op_describe(f, req[5], GOPERL_OPDUMP_SUBTREE) : 0;
            o = def->build_op(&opdata, lhs, rhs, def);
            if (o && !o->gop)
                goperl_croakf(f, "an infix operator built an op that is not the "
                                 "interpreter's");
            goperl_optree_flush(f);
        }
        resp[0] = GOPERL_TOK(opdata);
        resp[1] = o ? o->gop : 0;
        rlen = 16;
        break;
    }
    case GOPERL_HOST_UVAR: {
        /* [is_set][arg][sv][key] -> [key][result]. The interpreter has one
         * uvar magic for the value, which stands for every one native code
         * attached (its sv_magic adds no second): perl calls each of them,
         * in the order of the chain, and so does this. One native code took
         * off since is not called (Tk untraces a variable so); nor is one
         * whose ufuncs it has not put in yet. */
        if (reqlen < 40) goperl_croakf(f, "host call: short uvar request");
        SV *sv = GOPERL_SV(req[3]);
        int32_t r = 0;
        goperl_gword_t key = req[4];
        goperl_flush(f);
        for (MAGIC *mg = goperl_api_magic_chain(f, req[3]), *next; mg; mg = next) {
            /* a function may take its own magic off the chain */
            next = mg->mg_moremagic;
            if (mg->mg_type != PERL_MAGIC_uvar || !mg->mg_ptr) continue;
            const struct ufuncs *uf = (const struct ufuncs *)mg->mg_ptr;
            if (SvTYPE(sv) == SVt_PVHV) {
                /* perl's hash access calls uf_val itself, with the kind of
                 * access, of the first uvar magic; the key is in mg_obj,
                 * where the function may leave another */
                SV *const kept = mg->mg_obj;
                mg->mg_obj = GOPERL_SV(req[4]);
                if (uf->uf_val) r = uf->uf_val((IV)(int64_t)req[2], sv);
                key = GOPERL_TOK(mg->mg_obj);
                mg->mg_obj = kept;
                break;
            }
            if (mg->mg_virtual) {
                /* a scalar's magic runs through the MAGIC's vtable, as with
                 * perl's mg_get and mg_set: a module may have put its own
                 * there (SWIG's variables) */
                const MGVTBL *v = mg->mg_virtual;
                if (req[1]) {
                    if (v->svt_set) r = v->svt_set(sv, mg);
                } else if (v->svt_get) {
                    r = v->svt_get(sv, mg);
                }
            } else if (req[1]) {
                /* the mirror has no vtable of its own: perl's PL_vtbl_uvar,
                 * which calls the ufuncs with their index */
                if (uf->uf_set) r = uf->uf_set(uf->uf_index, sv);
            } else {
                if (uf->uf_val) r = uf->uf_val(uf->uf_index, sv);
            }
        }
        resp[0] = key;
        resp[1] = (goperl_gword_t)(int64_t)r;
        rlen = 16;
        break;
    }
    default:
        /* the functions of native code the interpreter runs as its own
         * (source filters, a run loop, block hooks: goperl_compiler.h) */
        rlen = goperl_host_invoke_fn(f, req, reqlen, resp);
        if (rlen < 0) goperl_croakf(f, "host call: unknown kind %u", (unsigned)req[0]);
    }
    goperl_hostsave_settle(f, f->hostsave_base, __builtin_frame_address(0));
    goperl_xs_leave(f);
    return rlen;
}

/* ---- custom op registration --------------------------------------------- */

#define XOPf_xop_name 0x01
#define XOPf_xop_desc 0x02
#define XOPf_xop_class 0x04
#define XOPf_xop_peep 0x08
#define XopFLAGS(xop) ((xop)->xop_flags)
#define XopENTRY_set(xop, which, to) \
    ((xop)->which = (to), (xop)->xop_flags |= XOPf_##which)
/* a field that is not set has perl's default (XOPd_xop_*) */
#define XOPd_xop_name "custom"
#define XOPd_xop_desc "unknown custom operator"
#define XOPd_xop_class OA_BASEOP
#define XOPd_xop_peep ((void (*)(pTHX_ OP *, OP *))0)
#define XopENTRY(xop, which) \
    (((xop)->xop_flags & XOPf_##which) ? (xop)->which : XOPd_##which)
#define XopDISABLE(xop, which) ((xop)->xop_flags &= ~XOPf_##which)
#define XopENABLE(xop, which) ((xop)->xop_flags |= XOPf_##which)

/* How many links a module's own op struct (NewOpSz) has, as perl reads
 * them: none without kids, op_first with them (OPf_KIDS), and op_other or
 * op_last as well unless the op is registered as a custom op of a class
 * that has only op_first. What follows the links is the module's. */
static int goperl_raw_links_by(const goperl_hooks_t *h, const OP *o) {
    if (!(o->op_flags & OPf_KIDS)) return 0;
    for (int32_t i = 0; h && i < h->xop_n; i++) {
        if (h->xops[i].pp != o->op_ppaddr) continue;
        const U32 cls = XopENTRY(h->xops[i].xop, xop_class);
        return cls == OA_BASEOP || cls == OA_UNOP || cls == OA_BASEOP_OR_UNOP ? 1 : 2;
    }
    return 2;
}
static int goperl_raw_links(const OP *o) {
    return goperl_raw_links_by(goperl_cur_frame_v ? goperl_hooks_v : 0, o);
}
/* A registered custom op is the interpreter's too: it is told the name,
 * the description and the class, and gives the op a number that the ops
 * running this pp function carry (see goperl_perop_pp_register). */
static void goperl_custom_op_register(goperl_frame_t *f, Perl_ppaddr_t pp,
                                      const XOP *xop) {
    goperl_hooks_t *h = goperl_hooks_get(f);
    for (int32_t i = 0; i < h->xop_n; i++)
        if (h->xops[i].pp == pp) return;
    if (h->xop_n >= GOPERL_XOP_MAX)
        goperl_croakf(f, "too many custom ops registered");
    const char *name = XopENTRY(xop, xop_name);
    const char *desc = XopENTRY(xop, xop_desc);
    const size_t nl = __builtin_strlen(name) + 1, dl = __builtin_strlen(desc) + 1;
    char *req = (char *)malloc(nl + dl);
    if (!req) goperl_croakf(f, "Out of memory!");
    memcpy(req, name, nl);
    memcpy(req + nl, desc, dl);
    /* CALL 12: XOP; the class as the interpreter's OA_ value */
    const uint32_t number = (uint32_t)goperl_do_op3(f, GOPERL_OP_CALL, 0, XopENTRY(xop, xop_class), (uint64_t)(12), req, nl + dl);
    free(req);
    h->xops[h->xop_n].pp = pp;
    h->xops[h->xop_n].xop = xop;
    h->xops[h->xop_n].number = number;
    h->xop_n++;
}
#define Perl_custom_op_register(...) goperl_custom_op_register(_gof, __VA_ARGS__)
#define custom_op_register(pp, xop) \
    goperl_custom_op_register(_gof, (pp), (xop))
/* The description an op of type OP_CUSTOM was registered with, found by the
 * pp function it runs; perl's empty description, whose fields all have
 * their defaults, for a pp function nothing registered. */
static const XOP goperl_xop_null = {0, 0, 0, 0, 0};
static const XOP *goperl_custom_op_xop(goperl_frame_t *f, const OP *o)
    __attribute__((unused));
static const XOP *goperl_custom_op_xop(goperl_frame_t *f, const OP *o) {
    (void)f;
    const goperl_hooks_t *h = goperl_hooks_v;
    for (int32_t i = 0; h && i < h->xop_n; i++)
        if (h->xops[i].pp == o->op_ppaddr) return h->xops[i].xop;
    return &goperl_xop_null;
}
#define Perl_custom_op_xop(...) goperl_custom_op_xop(_gof, __VA_ARGS__)
#define custom_op_xop(o) goperl_custom_op_xop(_gof, (o))

/* ---- character classification via the guest ----------------------------- */

enum {
    GOPERL_CC_IDFIRST = 0,
    GOPERL_CC_IDCONT = 1,
    GOPERL_CC_WORDCHAR = 2,
    GOPERL_CC_SPACE = 3,
    GOPERL_CC_DIGIT = 4,
    GOPERL_CC_ALPHA = 5,
    GOPERL_CC_UPPER = 6
};
static int goperl_charclass_uv(goperl_frame_t *f, int cls, UV cp) {
    return (int)goperl_do_op(f, GOPERL_OP_CHARCLASS,
                             (uint64_t)(uint32_t)cls, (uint64_t)cp, 0, 0);
}
/* The classes of a code point, as the interpreter's Unicode tables have
 * them. */
#define isALPHA_uvchr(c) goperl_charclass_uv(_gof, GOPERL_CC_ALPHA, (UV)(c))
#define isDIGIT_uvchr(c) goperl_charclass_uv(_gof, GOPERL_CC_DIGIT, (UV)(c))
#define isSPACE_uvchr(c) goperl_charclass_uv(_gof, GOPERL_CC_SPACE, (UV)(c))
#define isUPPER_uvchr(c) goperl_charclass_uv(_gof, GOPERL_CC_UPPER, (UV)(c))
#define isWORDCHAR_uvchr(c) goperl_charclass_uv(_gof, GOPERL_CC_WORDCHAR, (UV)(c))
#define isALNUM_uvchr(c) isWORDCHAR_uvchr(c)
#define isIDFIRST_uvchr(c) goperl_charclass_uv(_gof, GOPERL_CC_IDFIRST, (UV)(c))
#define isIDCONT_uvchr(c) goperl_charclass_uv(_gof, GOPERL_CC_IDCONT, (UV)(c))
#define isALPHA_uni(c) isALPHA_uvchr(c)
#define isDIGIT_uni(c) isDIGIT_uvchr(c)
#define isUPPER_uni(c) isUPPER_uvchr(c)
#define isWORDCHAR_uni(c) isWORDCHAR_uvchr(c)
/* The length of the character at s, but not past a NUL that ends the
 * string inside it. */
#define UTF8_CHK_SKIP(s)                                                   \
    ({ const U8 *const goperl_ucs_ = (const U8 *)(s);                      \
       goperl_ucs_[0] == 0 ? (STRLEN)1                                     \
                           : (STRLEN)strnlen((const char *)goperl_ucs_,    \
                                             (size_t)UTF8SKIP(goperl_ucs_)); })
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

#define NUM2PTR(type, num) INT2PTR(type, num)
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
    if (mode == 0 && (flags & SVf_UTF8)) {
        /* a name of characters: the 24 bits of flags there hold no
         * SVf_UTF8, an SV of the name does (sv_derived_from_pvn and _sv
         * are one function in perl) */
        SV *const n = GOPERL_SV(goperl_do_op(f, GOPERL_OP_NEW_PVN, 0, (uint64_t)len, name, (uint64_t)len));
        goperl_do_op(f, GOPERL_OP_SV_MORTAL, GOPERL_TOK(n), 0, 0, 0);
        goperl_do_op(f, GOPERL_OP_SV_UTF8_ON, GOPERL_TOK(n), 0, 0, 0);
        return goperl_sv_classify(f, sv, 1, flags & ~SVf_UTF8, 0, 0, n);
    }
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

/* The parameters of a signature (the aux of OP_MULTIPARAM), as perl
 * declares them. */
struct op_multiparam_named_aux {
    PADOFFSET padix;
    const char *namepv;
    STRLEN namelen;
    U32 namehash;
    bool is_required : 1;
};
struct op_multiparam_aux {
    size_t min_args;
    size_t n_positional;
    size_t n_named;
    char slurpy;
    PADOFFSET slurpy_padix;
    PADOFFSET *param_padix;
    struct op_multiparam_named_aux *named;
};
struct op_argcheck_aux {
    UV params;
    UV opt_params;
    char slurpy;
};

#define GOPERL_PAD_SETSV 4
#define PAD_SETSV(ix, sv) \
    ((void)goperl_do_op3(_gof, GOPERL_OP_PAD, GOPERL_PAD_SETSV, GOPERL_TOK(sv), \
                         (uint64_t)(ix), 0, 0))

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

/* scan_version reads a version out of a text into an SV and returns where
 * the version ends (CALL 17). The text is the interpreter's when it is in
 * its memory (the line being parsed), and is handed over otherwise. */
static const char *goperl_scan_version(goperl_frame_t *f, const char *s, SV *rv,
                                       bool qv) __attribute__((unused));
static const char *goperl_scan_version(goperl_frame_t *f, const char *s, SV *rv,
                                       bool qv) {
    const uint64_t g = goperl_guest_addr(f, s);
    /* CALL 17: b = the address of the text, or 0 for the text in s; qv in
     * c's high half */
    const uint64_t sel = 17 | ((uint64_t)(qv ? 1 : 0) << 32);
    const uint64_t taken = g ? goperl_do_op3(f, GOPERL_OP_CALL, GOPERL_TOK(rv), g, sel, 0, 0)
                             : goperl_do_op3(f, GOPERL_OP_CALL, GOPERL_TOK(rv), 0, sel, s,
                                             __builtin_strlen(s));
    return s + taken;
}
#define scan_version(s, rv, qv) \
    goperl_scan_version(_gof, (const char *)(s), (SV *)(rv), (bool)(qv))
#define Perl_scan_version(ctx, s, rv, qv) scan_version((s), (rv), (qv))

#define op_dump(o) goperl_do_op_dump(0, PerlIO_stderr(), (const OP *)(o))

/* call-checker surface (registered but not bridged; croaks on use) */
/* ---- call checkers ----------------------------------------------------------
 * cv_set_call_checker gives a sub a function that is handed every call of
 * the sub the compiler builds; the interpreter calls it through the
 * host-call channel. The checker a sub has when it was given none is the
 * interpreter's ck_entersub_args_proto_or_list, which is a function here
 * too: code compares with it and calls it. */
enum {
    GOPERL_TREE_NEWREF = 0,
    GOPERL_TREE_NEWFOROP = 1,
    GOPERL_TREE_NEWNULLLIST = 2,
    GOPERL_TREE_NEWDEFSVOP = 3,
    GOPERL_TREE_NEWRANGE = 4,
    GOPERL_TREE_NEWWHILEOP = 5,
    GOPERL_TREE_NEWANON = 6,
    GOPERL_TREE_OP_LVALUE = 7,
    GOPERL_TREE_START_SUBPARSE = 8,
    GOPERL_TREE_NEWATTRSUB = 9,
    GOPERL_TREE_NEWANONATTRSUB = 10,
    GOPERL_TREE_PAD_ADD_NAME_SV = 11,
    GOPERL_TREE_FIND_RUNCV = 12,
    GOPERL_TREE_SAVE_CLEARSV = 13,
    GOPERL_TREE_LEX_STUFF_SV = 14,
    GOPERL_TREE_RV2CV_OP_CV = 15,
    GOPERL_TREE_CK_ARGS = 16,
    GOPERL_TREE_CALLCK_SET = 17,
    GOPERL_TREE_CALLCK_GET = 18,
    GOPERL_TREE_CXINC = 19,
    GOPERL_TREE_NEWPVOP = 20,
    GOPERL_TREE_NEWLOOPEX = 21,
    GOPERL_TREE_NEWLOOPOP = 22
};
/* One call of a function that builds op trees: n words go out (the token of
 * an op, 0 for none), what modules wrote into shadows first. */
static uint64_t goperl_tree(goperl_frame_t *f, int sel, int n, const goperl_gword_t *w) {
    goperl_optree_flush(f);
    return goperl_do_op(f, GOPERL_OP_TREE, (uint64_t)sel, (uint64_t)n * 8,
                        (const char *)w, (uint64_t)n * 8);
}
static uint64_t goperl_optok(goperl_frame_t *f, const OP *o) {
    if (!o) return 0;
    if (!o->gop)
        goperl_croakf(f, "an op that is not the interpreter's was handed to "
                         "a function that builds op trees");
    return o->gop;
}
/* The op a builder returned, and the ops it was given described again (it
 * linked them into the new one). */
static OP *goperl_tree_op(goperl_frame_t *f, uint64_t tok) {
    return tok ? goperl_op_describe(f, tok, GOPERL_OPDUMP_SUBTREE) : 0;
}

#define RV2CVOPCV_MARK_EARLY 0x00000001
#define RV2CVOPCV_RETURN_NAME_GV 0x00000002
#define RV2CVOPCV_RETURN_STUB 0x00000004
#define RV2CVOPCV_FLAG_MASK 0x0000000f
#define CALL_CHECKER_REQUIRE_GV MGf_REQUIRE_GV
#ifndef MGf_REQUIRE_GV
#define MGf_REQUIRE_GV 1
#endif
static CV *goperl_rv2cv_op_cv(goperl_frame_t *f, OP *cvop, U32 flags)
    __attribute__((unused));
static CV *goperl_rv2cv_op_cv(goperl_frame_t *f, OP *cvop, U32 flags) {
    const goperl_gword_t w[2] = {goperl_optok(f, cvop), flags};
    return (CV *)GOPERL_SV(goperl_tree(f, GOPERL_TREE_RV2CV_OP_CV, 2, w));
}
#define rv2cv_op_cv(cvop, flags) goperl_rv2cv_op_cv(_gof, (OP *)(cvop), (U32)(flags))

static OP *goperl_ck_args(goperl_frame_t *f, int which, OP *o, GV *namegv,
                          SV *protosv) {
    const goperl_gword_t w[4] = {(uint32_t)which, goperl_optok(f, o),
                           GOPERL_TOK(namegv),
                           GOPERL_TOK(protosv)};
    return goperl_tree_op(f, goperl_tree(f, GOPERL_TREE_CK_ARGS, 4, w));
}
static OP *goperl_ck_entersub_args_proto_or_list(pTHX_ OP *o, GV *namegv,
                                                 SV *protosv)
    __attribute__((unused));
static OP *goperl_ck_entersub_args_proto_or_list(pTHX_ OP *o, GV *namegv,
                                                 SV *protosv) {
    return goperl_ck_args(_gof, 0, o, namegv, protosv);
}
#define Perl_ck_entersub_args_proto_or_list goperl_ck_entersub_args_proto_or_list
static OP *goperl_ck_entersub_args_proto(pTHX_ OP *o, GV *namegv, SV *protosv)
    __attribute__((unused));
static OP *goperl_ck_entersub_args_proto(pTHX_ OP *o, GV *namegv, SV *protosv) {
    return goperl_ck_args(_gof, 1, o, namegv, protosv);
}
#define Perl_ck_entersub_args_proto goperl_ck_entersub_args_proto
static OP *goperl_ck_entersub_args_list(pTHX_ OP *o) __attribute__((unused));
static OP *goperl_ck_entersub_args_list(pTHX_ OP *o) {
    return goperl_ck_args(_gof, 2, o, 0, 0);
}
#define Perl_ck_entersub_args_list goperl_ck_entersub_args_list
#define ck_entersub_args_proto_or_list(o, namegv, protosv) \
    goperl_ck_args(_gof, 0, (OP *)(o), (GV *)(namegv), (SV *)(protosv))
#define ck_entersub_args_proto(o, namegv, protosv) \
    goperl_ck_args(_gof, 1, (OP *)(o), (GV *)(namegv), (SV *)(protosv))
#define ck_entersub_args_list(o) goperl_ck_args(_gof, 2, (OP *)(o), 0, 0)


static void goperl_cv_set_call_compiler(goperl_frame_t *f, CV *cv,
                                       Perl_call_checker ckfun, SV *ckobj,
                                       U32 ckflags) __attribute__((unused));
static void goperl_cv_set_call_compiler(goperl_frame_t *f, CV *cv,
                                       Perl_call_checker ckfun, SV *ckobj,
                                       U32 ckflags) {
    uint32_t id = 0;
    if (ckfun != goperl_ck_entersub_args_proto_or_list) {
        goperl_callcks_t *c = goperl_callcks_v;
        if (!c) {
            c = (goperl_callcks_t *)calloc(1, sizeof(goperl_callcks_t));
            if (!c) goperl_croakf(f, "Out of memory!");
            goperl_callcks_v = c;
        }
        for (int32_t i = 0; i < c->n && !id; i++)
            if (c->fns[i] == ckfun) id = (uint32_t)i + 1;
        if (!id) {
            if (c->n == c->cap) {
                const int32_t ncap = c->cap ? c->cap * 2 : 8;
                Perl_call_checker *nf = (Perl_call_checker *)realloc(
                    c->fns, (size_t)ncap * sizeof(Perl_call_checker));
                if (!nf) goperl_croakf(f, "Out of memory!");
                c->fns = nf;
                c->cap = ncap;
            }
            c->fns[c->n++] = ckfun;
            id = (uint32_t)c->n;
        }
    }
    const goperl_gword_t w[4] = {GOPERL_TOK(cv), id,
                           GOPERL_TOK(ckobj), ckflags};
    goperl_tree(f, GOPERL_TREE_CALLCK_SET, 4, w);
}
#define cv_set_call_checker_flags(cv, ckfun, ckobj, ckflags)          \
    goperl_cv_set_call_compiler(_gof, (CV *)(cv), (ckfun), (SV *)(ckobj), \
                               (U32)(ckflags))
#define cv_set_call_checker(cv, ckfun, ckobj) \
    cv_set_call_checker_flags((cv), (ckfun), (ckobj), CALL_CHECKER_REQUIRE_GV)

static void goperl_cv_get_call_compiler(goperl_frame_t *f, CV *cv, U32 gflags,
                                       Perl_call_checker *ckfun_p,
                                       SV **ckobj_p, U32 *ckflags_p)
    __attribute__((unused));
static void goperl_cv_get_call_compiler(goperl_frame_t *f, CV *cv, U32 gflags,
                                       Perl_call_checker *ckfun_p,
                                       SV **ckobj_p, U32 *ckflags_p) {
    const goperl_gword_t w[2] = {GOPERL_TOK(cv), gflags};
    const goperl_gword_t *r = (const goperl_gword_t *)goperl_guest_mem(
        f, goperl_tree(f, GOPERL_TREE_CALLCK_GET, 2, w));
    if (!r) goperl_croakf(f, "cv_get_call_checker: no answer");
    const goperl_callcks_t *c = goperl_callcks_v;
    if (r[0] == 1 && c && r[1] >= 1 && (int64_t)r[1] <= c->n)
        *ckfun_p = c->fns[r[1] - 1];
    else if (r[0] == 2)
        *ckfun_p = goperl_ck_entersub_args_proto_or_list;
    else
        goperl_croakf(f, "cv_get_call_checker: the sub has a call checker of "
                         "the interpreter's that native code cannot call");
    *ckobj_p = GOPERL_SV(r[2]);
    if (ckflags_p) *ckflags_p = r[3];
}
#define cv_get_call_checker_flags(cv, gflags, ckfun_p, ckobj_p, ckflags_p) \
    goperl_cv_get_call_compiler(_gof, (CV *)(cv), (U32)(gflags), (ckfun_p),  \
                               (ckobj_p), (ckflags_p))
#define cv_get_call_checker(cv, ckfun_p, ckobj_p) \
    goperl_cv_get_call_compiler(_gof, (CV *)(cv), 0, (ckfun_p), (ckobj_p), 0)

/* ---- more functions that build op trees -------------------------------------- */

static OP *goperl_newref(goperl_frame_t *f, int kind, I32 arg, OP *o) {
    const goperl_gword_t w[3] = {(uint32_t)kind, (uint32_t)arg, goperl_optok(f, o)};
    return goperl_tree_op(f, goperl_tree(f, GOPERL_TREE_NEWREF, 3, w));
}
#define newCVREF(flags, o) goperl_newref(_gof, 0, (flags), (OP *)(o))
#define newAVREF(o) goperl_newref(_gof, 1, 0, (OP *)(o))
#define newHVREF(o) goperl_newref(_gof, 2, 0, (OP *)(o))
#define newSVREF(o) goperl_newref(_gof, 3, 0, (OP *)(o))
#define newGVREF(type, o) goperl_newref(_gof, 4, (type), (OP *)(o))
static OP *goperl_tree_ops(goperl_frame_t *f, int sel, I32 a, I32 b, OP *o1,
                           OP *o2, OP *o3, OP *o4, I32 last)
    __attribute__((unused));
static OP *goperl_tree_ops(goperl_frame_t *f, int sel, I32 a, I32 b, OP *o1,
                           OP *o2, OP *o3, OP *o4, I32 last) {
    goperl_gword_t w[7];
    int n = 0;
    switch (sel) {
    case GOPERL_TREE_NEWFOROP:
        w[n++] = (uint32_t)a;
        w[n++] = goperl_optok(f, o1);
        w[n++] = goperl_optok(f, o2);
        w[n++] = goperl_optok(f, o3);
        w[n++] = goperl_optok(f, o4);
        break;
    case GOPERL_TREE_NEWWHILEOP:
        w[n++] = (uint32_t)a;
        w[n++] = (uint32_t)b;
        w[n++] = goperl_optok(f, o1);
        w[n++] = goperl_optok(f, o2);
        w[n++] = goperl_optok(f, o3);
        w[n++] = goperl_optok(f, o4);
        w[n++] = (uint32_t)last;
        break;
    case GOPERL_TREE_NEWLOOPOP:
        w[n++] = (uint32_t)a;
        w[n++] = (uint32_t)b;
        w[n++] = goperl_optok(f, o1);
        w[n++] = goperl_optok(f, o2);
        break;
    case GOPERL_TREE_NEWRANGE:
        w[n++] = (uint32_t)a;
        w[n++] = goperl_optok(f, o1);
        w[n++] = goperl_optok(f, o2);
        break;
    case GOPERL_TREE_NEWANON:
    case GOPERL_TREE_NEWLOOPEX:
        w[n++] = (uint32_t)a;
        w[n++] = goperl_optok(f, o1);
        break;
    case GOPERL_TREE_OP_LVALUE:
        w[n++] = goperl_optok(f, o1);
        w[n++] = (uint32_t)a;
        break;
    case GOPERL_TREE_NEWANONATTRSUB:
        w[n++] = (uint32_t)a;
        w[n++] = goperl_optok(f, o1);
        w[n++] = goperl_optok(f, o2);
        w[n++] = goperl_optok(f, o3);
        break;
    default: break;
    }
    return goperl_tree_op(f, goperl_tree(f, sel, n, w));
}
#define newFOROP(flags, sv, expr, block, cont)                              \
    goperl_tree_ops(_gof, GOPERL_TREE_NEWFOROP, (flags), 0, (OP *)(sv),     \
                    (OP *)(expr), (OP *)(block), (OP *)(cont), 0)
#define newWHILEOP(flags, debuggable, loop, expr, block, cont, has_my)       \
    goperl_tree_ops(_gof, GOPERL_TREE_NEWWHILEOP, (flags), (debuggable),     \
                    (OP *)(loop), (OP *)(expr), (OP *)(block), (OP *)(cont), \
                    (has_my))
#define newLOOPOP(flags, debuggable, expr, block)                        \
    goperl_tree_ops(_gof, GOPERL_TREE_NEWLOOPOP, (flags), (debuggable), \
                    (OP *)(expr), (OP *)(block), 0, 0, 0)
#define newLOOPEX(type, label) \
    goperl_tree_ops(_gof, GOPERL_TREE_NEWLOOPEX, (type), 0, (OP *)(label), 0, 0, 0, 0)
#define newRANGE(flags, left, right)                                     \
    goperl_tree_ops(_gof, GOPERL_TREE_NEWRANGE, (flags), 0, (OP *)(left), \
                    (OP *)(right), 0, 0, 0)
#define newANONLIST(o) \
    goperl_tree_ops(_gof, GOPERL_TREE_NEWANON, 0, 0, (OP *)(o), 0, 0, 0, 0)
#define newANONHASH(o) \
    goperl_tree_ops(_gof, GOPERL_TREE_NEWANON, 1, 0, (OP *)(o), 0, 0, 0, 0)
#define newNULLLIST() goperl_tree_op(_gof, goperl_tree(_gof, GOPERL_TREE_NEWNULLLIST, 0, 0))
#define newDEFSVOP() goperl_tree_op(_gof, goperl_tree(_gof, GOPERL_TREE_NEWDEFSVOP, 0, 0))
#define op_lvalue(o, type) \
    goperl_tree_ops(_gof, GOPERL_TREE_OP_LVALUE, (type), 0, (OP *)(o), 0, 0, 0, 0)
#define newANONATTRSUB(floor, proto, attrs, block)                        \
    goperl_tree_ops(_gof, GOPERL_TREE_NEWANONATTRSUB, (floor), 0,         \
                    (OP *)(proto), (OP *)(attrs), (OP *)(block), 0, 0)
#define newANONSUB(floor, proto, block) newANONATTRSUB((floor), (proto), 0, (block))
static I32 goperl_start_subparse_(goperl_frame_t *f, I32 is_format, U32 flags)
    __attribute__((unused));
static I32 goperl_start_subparse_(goperl_frame_t *f, I32 is_format, U32 flags) {
    const goperl_gword_t w[2] = {(uint32_t)is_format, flags};
    return (I32)(int64_t)goperl_tree(f, GOPERL_TREE_START_SUBPARSE, 2, w);
}
static CV *goperl_newattrsub(goperl_frame_t *f, I32 floor, OP *o, OP *proto,
                             OP *attrs, OP *block) __attribute__((unused));
static CV *goperl_newattrsub(goperl_frame_t *f, I32 floor, OP *o, OP *proto,
                             OP *attrs, OP *block) {
    const goperl_gword_t w[5] = {(uint32_t)floor, goperl_optok(f, o),
                           goperl_optok(f, proto), goperl_optok(f, attrs),
                           goperl_optok(f, block)};
    return (CV *)GOPERL_SV(goperl_tree(f, GOPERL_TREE_NEWATTRSUB, 5, w));
}
static OP *goperl_newpvop(goperl_frame_t *f, I32 type, I32 flags, char *pv)
    __attribute__((unused));
static OP *goperl_newpvop(goperl_frame_t *f, I32 type, I32 flags, char *pv) {
    const size_t len = pv ? __builtin_strlen(pv) : 0;
    /* the two words, then the string */
    char *req = (char *)malloc(16 + len + 1);
    if (!req) goperl_croakf(f, "Out of memory!");
    const goperl_gword_t w[2] = {(goperl_gword_t)(int64_t)type, (goperl_gword_t)(int64_t)flags};
    memcpy(req, w, 16);
    memcpy(req + 16, pv ? pv : "", len + 1);
    goperl_optree_flush(f);
    const uint64_t tok = goperl_do_op(f, GOPERL_OP_TREE, GOPERL_TREE_NEWPVOP,
                                      16 + len + 1, req, 16 + len + 1);
    free(req);
    Safefree(pv); /* the op has its own copy; newPVOP takes the string over */
    return goperl_tree_op(f, tok);
}
#define newPVOP(type, flags, pv) goperl_newpvop(_gof, (type), (flags), (pv))

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
    if (size - sizeof(OP) > 0xFFFF)
        goperl_croakf(f, "NewOpSz: an op struct of %lu bytes is larger than "
                         "this SDK keeps", (unsigned long)size);
    uint64_t tok = goperl_do_op(f, GOPERL_OP_OPTREE_NEW,
                                ((uint64_t)GOPERL_OPC_RAW << 32), 0, 0, 0);
    goperl_opreg_t *r = goperl_opreg_get(f);
    goperl_opreg_sync(f, r);
    goperl_opreg_ent_t *e = (goperl_opreg_ent_t *)calloc(
        1, offsetof(goperl_opreg_ent_t, op) + size);
    if (!e) goperl_croakf(f, "NewOpSz: out of memory");
    e->persist = 1;
    e->ext = (U16)(size - sizeof(OP));
    e->op.gop = tok;
    goperl_opreg_insert(f, r, e);
    return goperl_op_describe(f, tok, GOPERL_OPDUMP_ONE);
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
/* The value in a pad slot, for code that names the slot as an lvalue to
 * reach its value (sv_setpvs(PAD_SVl(ix), ...)). The slot itself is the
 * interpreter's: putting another value in it is PAD_SETSV. */
#define PAD_SVl(ix) PAD_SV(ix)
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
#define SAVEFREEPV(p) \
    goperl_save_destructor(_gof, goperl_savefreepv_cb, (void *)(p), GOPERL_DTOR_CALL_OWN)

/* op_private flag values (from the guest perl's generated opcode.h) */
#define OPpCONST_STRICT 0x08
#define OPpCONST_BARE 0x20
#define OPpLVAL_INTRO 0x80
#define HINT_BLOCK_SCOPE 0x00000100

#include "goperl_perlapi.h"
#include "goperl_handy.h"

/* Case-insensitive comparison of bytes (inline.h, util.h): ASCII, Latin-1,
 * or the C library's locale; the ibcmp names are their negations. */
static inline __attribute__((unused)) I32 goperl_fold_eq(const char *s1,
                                                         const char *s2, I32 len,
                                                         const unsigned char *fold) {
    const U8 *a = (const U8 *)s1, *b = (const U8 *)s2;
    while (len-- > 0) {
        if (*a != *b && *a != fold[*b]) return 0;
        a++, b++;
    }
    return 1;
}
static inline __attribute__((unused)) I32 goperl_fold_eq_locale(const char *s1,
                                                                const char *s2,
                                                                I32 len) {
    const U8 *a = (const U8 *)s1, *b = (const U8 *)s2;
    while (len-- > 0) {
        if (*a != *b && *a != (isupper(*b) ? tolower(*b) : toupper(*b))) return 0;
        a++, b++;
    }
    return 1;
}
#define foldEQ(s1, s2, len) goperl_fold_eq((s1), (s2), (I32)(len), PL_fold)
#define foldEQ_latin1(s1, s2, len) \
    goperl_fold_eq((s1), (s2), (I32)(len), PL_fold_latin1)
#define foldEQ_locale(s1, s2, len) goperl_fold_eq_locale((s1), (s2), (I32)(len))
#define ibcmp(s1, s2, len) cBOOL(!foldEQ(s1, s2, len))
#define ibcmp_locale(s1, s2, len) cBOOL(!foldEQ_locale(s1, s2, len))

/* ---- a copy of the instance -----------------------------------------------
 * Loader entry: what this SDK keeps for an instance, for a copy of the
 * instance (a Clone, the child of a fork) whose interpreter memory is a
 * copy of the original's. dst is the copy's block, zeroed; src is the
 * original's.
 *
 * What modules registered with the interpreter goes with the copy, which
 * has its own tables of it from here on, as a forked process has its own
 * copy of perl's: the op checkers, peephole hooks and call_atexit
 * functions, the custom ops, the call checkers, the keyword and infix
 * plugins, the source filters, run loops and block hooks, the destructors
 * the interpreter's save stack names, and the ops of a module's own struct
 * (NewOpSz), whose fields past the op are nowhere else. The strings
 * interned so far are shared: they are never changed or freed.
 *
 * The rest is the state of the native calls in progress, which exist only
 * in the original, or what was read from the interpreter and is read again
 * when needed. Memory of this process is shared by the copies, as a
 * module's own variables are; the destructors that release such memory
 * (SAVEFREEPV, SSNEW, deferred saves) stay with the original, so that it is
 * released once.
 *
 * Returns 0, or -1 when memory runs out. */
GOPERL_C __attribute__((weak, visibility("default"), used)) int32_t
__goperl_shared_clone(void *dst_raw, const void *src_raw) {
    goperl_shared_t *const dst = (goperl_shared_t *)dst_raw;
    const goperl_shared_t *const src = (const goperl_shared_t *)src_raw;
    dst->modglobal_seeded = src->modglobal_seeded;
    dst->kw_head = src->kw_head;
    dst->infix_head = src->infix_head;
    memcpy(dst->intern, src->intern, sizeof dst->intern);
    dst->na = src->na;

    const goperl_dtor_t *const sd = (const goperl_dtor_t *)src->dtors;
    goperl_dtor_t *const dd = (goperl_dtor_t *)dst->dtors;
    for (int32_t i = 0; i < GOPERL_DTOR_MAX; i++)
        if (sd[i].used == GOPERL_DTOR_X || sd[i].used == GOPERL_DTOR_PLAIN)
            dd[i] = sd[i];
    dst->dtor_hint = src->dtor_hint;

    goperl_hooks_t *const sh = (goperl_hooks_t *)src->hooks;
    if (sh) {
        goperl_hooks_t *const h = (goperl_hooks_t *)malloc(sizeof *h);
        if (!h) return -1;
        *h = *sh;
        h->atexit = 0;
        h->atexit_cap = 0;
        if (sh->atexit_n > 0) {
            h->atexit = (goperl_atexit_t *)malloc((size_t)sh->atexit_n *
                                                  sizeof(goperl_atexit_t));
            if (!h->atexit) {
                free(h);
                return -1;
            }
            /* a function given to call_atexit releases what the module
             * keeps, which the copies share: it runs once, for the last
             * instance that holds it */
            for (int32_t i = 0; i < sh->atexit_n; i++) {
                goperl_atexit_t *const a = &sh->atexit[i];
                if (a->fn && !a->holders) {
                    a->holders = (int32_t *)malloc(sizeof(int32_t));
                    if (!a->holders) {
                        free(h->atexit);
                        free(h);
                        return -1;
                    }
                    *a->holders = 1;
                }
                if (a->holders) __atomic_add_fetch(a->holders, 1, __ATOMIC_ACQ_REL);
                h->atexit[i] = *a;
            }
            h->atexit_cap = sh->atexit_n;
        }
        dst->hooks = h;
    }

    const goperl_callcks_t *const sc = (const goperl_callcks_t *)src->callcks;
    if (sc) {
        goperl_callcks_t *const c = (goperl_callcks_t *)calloc(1, sizeof *c);
        if (!c) return -1;
        if (sc->n > 0) {
            c->fns = (Perl_call_checker *)malloc((size_t)sc->n * sizeof(Perl_call_checker));
            if (!c->fns) {
                free(c);
                return -1;
            }
            memcpy(c->fns, sc->fns, (size_t)sc->n * sizeof(Perl_call_checker));
            c->n = c->cap = sc->n;
        }
        dst->callcks = c;
    }

    const goperl_fntab_t *const st = (const goperl_fntab_t *)src->fntab;
    if (st) {
        goperl_fntab_t *const t = (goperl_fntab_t *)malloc(sizeof *t);
        if (!t) return -1;
        *t = *st;
        dst->fntab = t;
    }

    /* the ops of a module's own struct: a copy of each, whose links are
     * the interpreter's to describe again in the copy */
    const goperl_opreg_t *const sr = (const goperl_opreg_t *)src->opreg[0];
    if (sr) {
        goperl_opreg_t *const r = (goperl_opreg_t *)calloc(1, sizeof *r);
        if (!r) return -1;
        r->nbuckets = sr->nbuckets;
        r->buckets = (goperl_opreg_ent_t **)calloc(r->nbuckets, sizeof(void *));
        if (!r->buckets) {
            free(r);
            return -1;
        }
        r->gen = 1;
        for (uint32_t b = 0; b < sr->nbuckets; b++)
            for (const goperl_opreg_ent_t *e = sr->buckets[b]; e; e = e->next) {
                if (!e->persist || !e->op.gop) continue;
                const size_t size = offsetof(goperl_opreg_ent_t, op) + sizeof(OP) + e->ext;
                goperl_opreg_ent_t *const c = (goperl_opreg_ent_t *)malloc(size);
                if (!c) return -1;
                memcpy(c, e, size);
                c->active_next = 0;
                c->active = 0;
                c->gen = 0;
                c->aux_own = 0;
                c->base = 0;
                c->op.gop_full = 0;
                c->op.op_next = c->op.op_sibparent = 0;
                const int links = goperl_raw_links_by(sh, &c->op);
                if (links >= 1) c->op.op_first = 0;
                if (links >= 2) c->op.op_other = 0;
                const uint32_t h = goperl_opreg_hash(r, c->op.gop);
                c->next = r->buckets[h];
                r->buckets[h] = c;
                r->count++;
            }
        dst->opreg[0] = r;
    }
    return 0;
}

/* Loader entry: the instance of the block is gone. The functions given to
 * call_atexit it shares with copies and did not run are left to them: the
 * last instance that runs one runs it. */
GOPERL_C __attribute__((weak, visibility("default"), used)) void
__goperl_shared_release(void *raw) {
    goperl_hooks_t *const h = (goperl_hooks_t *)((goperl_shared_t *)raw)->hooks;
    for (int32_t i = 0; h && i < h->atexit_n; i++) {
        goperl_atexit_t *const a = &h->atexit[i];
        if (!a->fn || !a->holders) continue;
        if (__atomic_sub_fetch(a->holders, 1, __ATOMIC_ACQ_REL) == 0) free(a->holders);
        a->fn = 0;
        a->holders = 0;
    }
}

/* perl's get_context: the interpreter of the current native call. */
#ifndef Perl_get_context
#define Perl_get_context() ((void *)goperl_interp_cur())
#endif
#include "goperl_longnames.h"
#include "goperl_argsassert.h"
/* last: the internals go-perl does not give native code (their names stay
 * out of every header above) */
#include "goperl_internals.h"

#endif /* GOPERL_XS_SDK_PERL_H */
