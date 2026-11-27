/* go-perl native XS SDK: the long names of perl's public functions
 * (GENERATED from perl v5.44.0's embed.fnc by gen_longnames2.py; do not edit
 * by hand).
 *
 * Perl names every public function of its API Perl_<name> as well; code
 * calls it so, Perl_av_push(aTHX_ av, sv), where aTHX_ is nothing in a perl
 * without MULTIPLICITY, as this one is. A long name is a function with
 * perl's prototype over the short name the SDK provides, inline only: a
 * module that has a copy of the function of its own (a copy of perl's
 * source) defines it in the same unit, and its definition is the one
 * compiled. Where no such function can be written (a variable argument
 * list), the long name is a macro over the short name. */
#ifndef GOPERL_XS_SDK_LONGNAMES_H
#define GOPERL_XS_SDK_LONGNAMES_H
/* the inline functions call the SDK's static ones */
#pragma GCC system_header

#define GOPERL_THX_DROP(...) __VA_ARGS__
/* the arguments are expanded before the short name's macro takes them */
#define GOPERL_LONG_CALL(f, ...) f(__VA_ARGS__)

static inline __attribute__((always_inline, unused)) SV *
goperl_long_amagic_call(SV * goperl_a0, SV * goperl_a1, int goperl_a2, int goperl_a3) { return amagic_call(goperl_a0, goperl_a1, goperl_a2, goperl_a3); }
extern inline __attribute__((gnu_inline, always_inline, unused)) SV *
Perl_amagic_call(pTHX_ SV * goperl_a0, SV * goperl_a1, int goperl_a2, int goperl_a3) { return goperl_long_amagic_call(goperl_a0, goperl_a1, goperl_a2, goperl_a3); }
static inline __attribute__((always_inline, unused)) SV *
goperl_long_amagic_deref_call(SV * goperl_a0, int goperl_a1) { return amagic_deref_call(goperl_a0, goperl_a1); }
extern inline __attribute__((gnu_inline, always_inline, unused)) SV *
Perl_amagic_deref_call(pTHX_ SV * goperl_a0, int goperl_a1) { return goperl_long_amagic_deref_call(goperl_a0, goperl_a1); }
static inline __attribute__((always_inline, unused)) void
goperl_long_apply_attrs_string(const char * goperl_a0, CV * goperl_a1, const char * goperl_a2, STRLEN goperl_a3) { apply_attrs_string(goperl_a0, goperl_a1, goperl_a2, goperl_a3); }
extern inline __attribute__((gnu_inline, always_inline, unused)) void
Perl_apply_attrs_string(pTHX_ const char * goperl_a0, CV * goperl_a1, const char * goperl_a2, STRLEN goperl_a3) { goperl_long_apply_attrs_string(goperl_a0, goperl_a1, goperl_a2, goperl_a3); }
#if defined(apply_builtin_cv_attributes) && !defined(Perl_apply_builtin_cv_attributes)
#define Perl_apply_builtin_cv_attributes(...) GOPERL_LONG_CALL(apply_builtin_cv_attributes, GOPERL_THX_DROP(__VA_ARGS__))
#endif
static inline __attribute__((always_inline, unused)) void
goperl_long_av_clear(AV * goperl_a0) { av_clear(goperl_a0); }
extern inline __attribute__((gnu_inline, always_inline, unused)) void
Perl_av_clear(pTHX_ AV * goperl_a0) { goperl_long_av_clear(goperl_a0); }
static inline __attribute__((always_inline, unused)) Size_t
goperl_long_av_count(AV * goperl_a0) { return av_count(goperl_a0); }
extern inline __attribute__((gnu_inline, always_inline, unused)) Size_t
Perl_av_count(pTHX_ AV * goperl_a0) { return goperl_long_av_count(goperl_a0); }
#if defined(av_create_and_push) && !defined(Perl_av_create_and_push)
#define Perl_av_create_and_push(...) GOPERL_LONG_CALL(av_create_and_push, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(av_create_and_unshift_one) && !defined(Perl_av_create_and_unshift_one)
#define Perl_av_create_and_unshift_one(...) GOPERL_LONG_CALL(av_create_and_unshift_one, GOPERL_THX_DROP(__VA_ARGS__))
#endif
static inline __attribute__((always_inline, unused)) SV *
goperl_long_av_delete(AV * goperl_a0, SSize_t goperl_a1, I32 goperl_a2) { return av_delete(goperl_a0, goperl_a1, goperl_a2); }
extern inline __attribute__((gnu_inline, always_inline, unused)) SV *
Perl_av_delete(pTHX_ AV * goperl_a0, SSize_t goperl_a1, I32 goperl_a2) { return goperl_long_av_delete(goperl_a0, goperl_a1, goperl_a2); }
#if defined(av_dump) && !defined(Perl_av_dump)
#define Perl_av_dump(...) GOPERL_LONG_CALL(av_dump, GOPERL_THX_DROP(__VA_ARGS__))
#endif
static inline __attribute__((always_inline, unused)) bool
goperl_long_av_exists(AV * goperl_a0, SSize_t goperl_a1) { return av_exists(goperl_a0, goperl_a1); }
extern inline __attribute__((gnu_inline, always_inline, unused)) bool
Perl_av_exists(pTHX_ AV * goperl_a0, SSize_t goperl_a1) { return goperl_long_av_exists(goperl_a0, goperl_a1); }
static inline __attribute__((always_inline, unused)) void
goperl_long_av_extend(AV * goperl_a0, SSize_t goperl_a1) { av_extend(goperl_a0, goperl_a1); }
extern inline __attribute__((gnu_inline, always_inline, unused)) void
Perl_av_extend(pTHX_ AV * goperl_a0, SSize_t goperl_a1) { goperl_long_av_extend(goperl_a0, goperl_a1); }
static inline __attribute__((always_inline, unused)) SV **
goperl_long_av_fetch(AV * goperl_a0, SSize_t goperl_a1, I32 goperl_a2) { return av_fetch(goperl_a0, goperl_a1, goperl_a2); }
extern inline __attribute__((gnu_inline, always_inline, unused)) SV **
Perl_av_fetch(pTHX_ AV * goperl_a0, SSize_t goperl_a1, I32 goperl_a2) { return goperl_long_av_fetch(goperl_a0, goperl_a1, goperl_a2); }
static inline __attribute__((always_inline, unused)) void
goperl_long_av_fill(AV * goperl_a0, SSize_t goperl_a1) { av_fill(goperl_a0, goperl_a1); }
extern inline __attribute__((gnu_inline, always_inline, unused)) void
Perl_av_fill(pTHX_ AV * goperl_a0, SSize_t goperl_a1) { goperl_long_av_fill(goperl_a0, goperl_a1); }
static inline __attribute__((always_inline, unused)) SSize_t
goperl_long_av_len(AV * goperl_a0) { return av_len(goperl_a0); }
extern inline __attribute__((gnu_inline, always_inline, unused)) SSize_t
Perl_av_len(pTHX_ AV * goperl_a0) { return goperl_long_av_len(goperl_a0); }
static inline __attribute__((always_inline, unused)) AV *
goperl_long_av_make(SSize_t goperl_a0, SV ** goperl_a1) { return av_make(goperl_a0, goperl_a1); }
extern inline __attribute__((gnu_inline, always_inline, unused)) AV *
Perl_av_make(pTHX_ SSize_t goperl_a0, SV ** goperl_a1) { return goperl_long_av_make(goperl_a0, goperl_a1); }
static inline __attribute__((always_inline, unused)) SV *
goperl_long_av_pop(AV * goperl_a0) { return av_pop(goperl_a0); }
extern inline __attribute__((gnu_inline, always_inline, unused)) SV *
Perl_av_pop(pTHX_ AV * goperl_a0) { return goperl_long_av_pop(goperl_a0); }
static inline __attribute__((always_inline, unused)) void
goperl_long_av_push(AV * goperl_a0, SV * goperl_a1) { av_push(goperl_a0, goperl_a1); }
extern inline __attribute__((gnu_inline, always_inline, unused)) void
Perl_av_push(pTHX_ AV * goperl_a0, SV * goperl_a1) { goperl_long_av_push(goperl_a0, goperl_a1); }
#if defined(av_push_simple) && !defined(Perl_av_push_simple)
#define Perl_av_push_simple(...) GOPERL_LONG_CALL(av_push_simple, GOPERL_THX_DROP(__VA_ARGS__))
#endif
static inline __attribute__((always_inline, unused)) SV *
goperl_long_av_shift(AV * goperl_a0) { return av_shift(goperl_a0); }
extern inline __attribute__((gnu_inline, always_inline, unused)) SV *
Perl_av_shift(pTHX_ AV * goperl_a0) { return goperl_long_av_shift(goperl_a0); }
static inline __attribute__((always_inline, unused)) SV **
goperl_long_av_store(AV * goperl_a0, SSize_t goperl_a1, SV * goperl_a2) { return av_store(goperl_a0, goperl_a1, goperl_a2); }
extern inline __attribute__((gnu_inline, always_inline, unused)) SV **
Perl_av_store(pTHX_ AV * goperl_a0, SSize_t goperl_a1, SV * goperl_a2) { return goperl_long_av_store(goperl_a0, goperl_a1, goperl_a2); }
static inline __attribute__((always_inline, unused)) void
goperl_long_av_undef(AV * goperl_a0) { av_undef(goperl_a0); }
extern inline __attribute__((gnu_inline, always_inline, unused)) void
Perl_av_undef(pTHX_ AV * goperl_a0) { goperl_long_av_undef(goperl_a0); }
static inline __attribute__((always_inline, unused)) void
goperl_long_av_unshift(AV * goperl_a0, SSize_t goperl_a1) { av_unshift(goperl_a0, goperl_a1); }
extern inline __attribute__((gnu_inline, always_inline, unused)) void
Perl_av_unshift(pTHX_ AV * goperl_a0, SSize_t goperl_a1) { goperl_long_av_unshift(goperl_a0, goperl_a1); }
static inline __attribute__((always_inline, unused)) int
goperl_long_bytes_cmp_utf8(const U8 * goperl_a0, STRLEN goperl_a1, const U8 * goperl_a2, STRLEN goperl_a3) { return bytes_cmp_utf8(goperl_a0, goperl_a1, goperl_a2, goperl_a3); }
extern inline __attribute__((gnu_inline, always_inline, unused)) int
Perl_bytes_cmp_utf8(pTHX_ const U8 * goperl_a0, STRLEN goperl_a1, const U8 * goperl_a2, STRLEN goperl_a3) { return goperl_long_bytes_cmp_utf8(goperl_a0, goperl_a1, goperl_a2, goperl_a3); }
static inline __attribute__((always_inline, unused)) U8 *
goperl_long_bytes_from_utf8(const U8 * goperl_a0, STRLEN * goperl_a1, bool * goperl_a2) { return bytes_from_utf8(goperl_a0, goperl_a1, goperl_a2); }
extern inline __attribute__((gnu_inline, always_inline, unused)) U8 *
Perl_bytes_from_utf8(pTHX_ const U8 * goperl_a0, STRLEN * goperl_a1, bool * goperl_a2) { return goperl_long_bytes_from_utf8(goperl_a0, goperl_a1, goperl_a2); }
static inline __attribute__((always_inline, unused)) U8 *
goperl_long_bytes_to_utf8(const U8 * goperl_a0, STRLEN * goperl_a1) { return bytes_to_utf8(goperl_a0, goperl_a1); }
extern inline __attribute__((gnu_inline, always_inline, unused)) U8 *
Perl_bytes_to_utf8(pTHX_ const U8 * goperl_a0, STRLEN * goperl_a1) { return goperl_long_bytes_to_utf8(goperl_a0, goperl_a1); }
#if defined(bytes_to_utf8_free_me) && !defined(Perl_bytes_to_utf8_free_me)
#define Perl_bytes_to_utf8_free_me(...) GOPERL_LONG_CALL(bytes_to_utf8_free_me, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(bytes_to_utf8_temp_pv) && !defined(Perl_bytes_to_utf8_temp_pv)
#define Perl_bytes_to_utf8_temp_pv(...) GOPERL_LONG_CALL(bytes_to_utf8_temp_pv, GOPERL_THX_DROP(__VA_ARGS__))
#endif
static inline __attribute__((always_inline, unused)) SSize_t
goperl_long_call_argv(const char * goperl_a0, I32 goperl_a1, char ** goperl_a2) { return call_argv(goperl_a0, goperl_a1, goperl_a2); }
extern inline __attribute__((gnu_inline, always_inline, unused)) SSize_t
Perl_call_argv(pTHX_ const char * goperl_a0, I32 goperl_a1, char ** goperl_a2) { return goperl_long_call_argv(goperl_a0, goperl_a1, goperl_a2); }
static inline __attribute__((always_inline, unused)) const PERL_CONTEXT *
goperl_long_caller_cx(I32 goperl_a0, const PERL_CONTEXT ** goperl_a1) { return caller_cx(goperl_a0, goperl_a1); }
extern inline __attribute__((gnu_inline, always_inline, unused)) const PERL_CONTEXT *
Perl_caller_cx(pTHX_ I32 goperl_a0, const PERL_CONTEXT ** goperl_a1) { return goperl_long_caller_cx(goperl_a0, goperl_a1); }
static inline __attribute__((always_inline, unused)) SSize_t
goperl_long_call_method(const char * goperl_a0, I32 goperl_a1) { return call_method(goperl_a0, goperl_a1); }
extern inline __attribute__((gnu_inline, always_inline, unused)) SSize_t
Perl_call_method(pTHX_ const char * goperl_a0, I32 goperl_a1) { return goperl_long_call_method(goperl_a0, goperl_a1); }
static inline __attribute__((always_inline, unused)) SSize_t
goperl_long_call_pv(const char * goperl_a0, I32 goperl_a1) { return call_pv(goperl_a0, goperl_a1); }
extern inline __attribute__((gnu_inline, always_inline, unused)) SSize_t
Perl_call_pv(pTHX_ const char * goperl_a0, I32 goperl_a1) { return goperl_long_call_pv(goperl_a0, goperl_a1); }
#if defined(cop_disable_warning) && !defined(Perl_cop_disable_warning)
#define Perl_cop_disable_warning(...) GOPERL_LONG_CALL(cop_disable_warning, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(cop_enable_warning) && !defined(Perl_cop_enable_warning)
#define Perl_cop_enable_warning(...) GOPERL_LONG_CALL(cop_enable_warning, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(cop_fetch_label) && !defined(Perl_cop_fetch_label)
#define Perl_cop_fetch_label(...) GOPERL_LONG_CALL(cop_fetch_label, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(cop_has_warning) && !defined(Perl_cop_has_warning)
#define Perl_cop_has_warning(...) GOPERL_LONG_CALL(cop_has_warning, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(cop_store_label) && !defined(Perl_cop_store_label)
#define Perl_cop_store_label(...) GOPERL_LONG_CALL(cop_store_label, GOPERL_THX_DROP(__VA_ARGS__))
#endif
static inline __attribute__((always_inline, unused)) void
goperl_long_croak_sv(SV * goperl_a0) { croak_sv(goperl_a0); }
extern inline __attribute__((gnu_inline, always_inline, unused)) void
Perl_croak_sv(pTHX_ SV * goperl_a0) { goperl_long_croak_sv(goperl_a0); }
static inline __attribute__((always_inline, unused)) void
goperl_long_croak_xs_usage(const CV * const goperl_a0, const char * const goperl_a1) { croak_xs_usage(goperl_a0, goperl_a1); }
extern inline __attribute__((gnu_inline, always_inline, unused)) void
Perl_croak_xs_usage(const CV * const goperl_a0, const char * const goperl_a1) { goperl_long_croak_xs_usage(goperl_a0, goperl_a1); }
#if defined(c9strict_utf8_to_uv) && !defined(Perl_c9strict_utf8_to_uv)
#define Perl_c9strict_utf8_to_uv(...) GOPERL_LONG_CALL(c9strict_utf8_to_uv, __VA_ARGS__)
#endif
static inline __attribute__((always_inline, unused)) SV *
goperl_long_cv_const_sv(const CV * const goperl_a0) { return cv_const_sv(goperl_a0); }
extern inline __attribute__((gnu_inline, always_inline, unused)) SV *
Perl_cv_const_sv(const CV * const goperl_a0) { return goperl_long_cv_const_sv(goperl_a0); }
#if defined(CvDEPTH) && !defined(Perl_CvDEPTH)
#define Perl_CvDEPTH(...) GOPERL_LONG_CALL(CvDEPTH, __VA_ARGS__)
#endif
static inline __attribute__((always_inline, unused)) void
goperl_long_cv_get_call_checker(CV * goperl_a0, Perl_call_checker * goperl_a1, SV ** goperl_a2) { cv_get_call_checker(goperl_a0, goperl_a1, goperl_a2); }
extern inline __attribute__((gnu_inline, always_inline, unused)) void
Perl_cv_get_call_checker(pTHX_ CV * goperl_a0, Perl_call_checker * goperl_a1, SV ** goperl_a2) { goperl_long_cv_get_call_checker(goperl_a0, goperl_a1, goperl_a2); }
static inline __attribute__((always_inline, unused)) void
goperl_long_cv_get_call_checker_flags(CV * goperl_a0, U32 goperl_a1, Perl_call_checker * goperl_a2, SV ** goperl_a3, U32 * goperl_a4) { cv_get_call_checker_flags(goperl_a0, goperl_a1, goperl_a2, goperl_a3, goperl_a4); }
extern inline __attribute__((gnu_inline, always_inline, unused)) void
Perl_cv_get_call_checker_flags(pTHX_ CV * goperl_a0, U32 goperl_a1, Perl_call_checker * goperl_a2, SV ** goperl_a3, U32 * goperl_a4) { goperl_long_cv_get_call_checker_flags(goperl_a0, goperl_a1, goperl_a2, goperl_a3, goperl_a4); }
static inline __attribute__((always_inline, unused)) GV *
goperl_long_CvGV(CV * goperl_a0) { return CvGV(goperl_a0); }
extern inline __attribute__((gnu_inline, always_inline, unused)) GV *
Perl_CvGV(pTHX_ CV * goperl_a0) { return goperl_long_CvGV(goperl_a0); }
#if defined(cv_name) && !defined(Perl_cv_name)
#define Perl_cv_name(...) GOPERL_LONG_CALL(cv_name, GOPERL_THX_DROP(__VA_ARGS__))
#endif
static inline __attribute__((always_inline, unused)) void
goperl_long_cv_set_call_checker(CV * goperl_a0, Perl_call_checker goperl_a1, SV * goperl_a2) { cv_set_call_checker(goperl_a0, goperl_a1, goperl_a2); }
extern inline __attribute__((gnu_inline, always_inline, unused)) void
Perl_cv_set_call_checker(pTHX_ CV * goperl_a0, Perl_call_checker goperl_a1, SV * goperl_a2) { goperl_long_cv_set_call_checker(goperl_a0, goperl_a1, goperl_a2); }
static inline __attribute__((always_inline, unused)) void
goperl_long_cv_set_call_checker_flags(CV * goperl_a0, Perl_call_checker goperl_a1, SV * goperl_a2, U32 goperl_a3) { cv_set_call_checker_flags(goperl_a0, goperl_a1, goperl_a2, goperl_a3); }
extern inline __attribute__((gnu_inline, always_inline, unused)) void
Perl_cv_set_call_checker_flags(pTHX_ CV * goperl_a0, Perl_call_checker goperl_a1, SV * goperl_a2, U32 goperl_a3) { goperl_long_cv_set_call_checker_flags(goperl_a0, goperl_a1, goperl_a2, goperl_a3); }
static inline __attribute__((always_inline, unused)) void
goperl_long_cv_undef(CV * goperl_a0) { cv_undef(goperl_a0); }
extern inline __attribute__((gnu_inline, always_inline, unused)) void
Perl_cv_undef(pTHX_ CV * goperl_a0) { goperl_long_cv_undef(goperl_a0); }
#if defined(debstack) && !defined(Perl_debstack)
#define Perl_debstack(...) GOPERL_LONG_CALL(debstack, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(delimcpy) && !defined(Perl_delimcpy)
#define Perl_delimcpy(...) GOPERL_LONG_CALL(delimcpy, __VA_ARGS__)
#endif
#if defined(die_sv) && !defined(Perl_die_sv)
#define Perl_die_sv(...) GOPERL_LONG_CALL(die_sv, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(do_close) && !defined(Perl_do_close)
#define Perl_do_close(...) GOPERL_LONG_CALL(do_close, GOPERL_THX_DROP(__VA_ARGS__))
#endif
static inline __attribute__((always_inline, unused)) void
goperl_long_do_join(SV * goperl_a0, SV * goperl_a1, SV ** goperl_a2, SV ** goperl_a3) { do_join(goperl_a0, goperl_a1, goperl_a2, goperl_a3); }
extern inline __attribute__((gnu_inline, always_inline, unused)) void
Perl_do_join(pTHX_ SV * goperl_a0, SV * goperl_a1, SV ** goperl_a2, SV ** goperl_a3) { goperl_long_do_join(goperl_a0, goperl_a1, goperl_a2, goperl_a3); }
static inline __attribute__((always_inline, unused)) bool
goperl_long_do_open(GV * goperl_a0, const char * goperl_a1, I32 goperl_a2, int goperl_a3, int goperl_a4, int goperl_a5, PerlIO * goperl_a6) { return do_open(goperl_a0, goperl_a1, goperl_a2, goperl_a3, goperl_a4, goperl_a5, goperl_a6); }
extern inline __attribute__((gnu_inline, always_inline, unused)) bool
Perl_do_open(pTHX_ GV * goperl_a0, const char * goperl_a1, I32 goperl_a2, int goperl_a3, int goperl_a4, int goperl_a5, PerlIO * goperl_a6) { return goperl_long_do_open(goperl_a0, goperl_a1, goperl_a2, goperl_a3, goperl_a4, goperl_a5, goperl_a6); }
static inline __attribute__((always_inline, unused)) bool
goperl_long_do_openn(GV * goperl_a0, const char * goperl_a1, I32 goperl_a2, int goperl_a3, int goperl_a4, int goperl_a5, PerlIO * goperl_a6, SV ** goperl_a7, I32 goperl_a8) { return do_openn(goperl_a0, goperl_a1, goperl_a2, goperl_a3, goperl_a4, goperl_a5, goperl_a6, goperl_a7, goperl_a8); }
extern inline __attribute__((gnu_inline, always_inline, unused)) bool
Perl_do_openn(pTHX_ GV * goperl_a0, const char * goperl_a1, I32 goperl_a2, int goperl_a3, int goperl_a4, int goperl_a5, PerlIO * goperl_a6, SV ** goperl_a7, I32 goperl_a8) { return goperl_long_do_openn(goperl_a0, goperl_a1, goperl_a2, goperl_a3, goperl_a4, goperl_a5, goperl_a6, goperl_a7, goperl_a8); }
#if defined(do_sprintf) && !defined(Perl_do_sprintf)
#define Perl_do_sprintf(...) GOPERL_LONG_CALL(do_sprintf, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(dump_all) && !defined(Perl_dump_all)
#define Perl_dump_all(...) GOPERL_LONG_CALL(dump_all, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(dump_eval) && !defined(Perl_dump_eval)
#define Perl_dump_eval(...) GOPERL_LONG_CALL(dump_eval, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(dump_form) && !defined(Perl_dump_form)
#define Perl_dump_form(...) GOPERL_LONG_CALL(dump_form, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(dump_packsubs) && !defined(Perl_dump_packsubs)
#define Perl_dump_packsubs(...) GOPERL_LONG_CALL(dump_packsubs, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(dump_sub) && !defined(Perl_dump_sub)
#define Perl_dump_sub(...) GOPERL_LONG_CALL(dump_sub, GOPERL_THX_DROP(__VA_ARGS__))
#endif
static inline __attribute__((always_inline, unused)) SV *
goperl_long_eval_pv(const char * goperl_a0, I32 goperl_a1) { return eval_pv(goperl_a0, goperl_a1); }
extern inline __attribute__((gnu_inline, always_inline, unused)) SV *
Perl_eval_pv(pTHX_ const char * goperl_a0, I32 goperl_a1) { return goperl_long_eval_pv(goperl_a0, goperl_a1); }
static inline __attribute__((always_inline, unused)) SSize_t
goperl_long_eval_sv(SV * goperl_a0, I32 goperl_a1) { return eval_sv(goperl_a0, goperl_a1); }
extern inline __attribute__((gnu_inline, always_inline, unused)) SSize_t
Perl_eval_sv(pTHX_ SV * goperl_a0, I32 goperl_a1) { return goperl_long_eval_sv(goperl_a0, goperl_a1); }
#if defined(extended_utf8_to_uv) && !defined(Perl_extended_utf8_to_uv)
#define Perl_extended_utf8_to_uv(...) GOPERL_LONG_CALL(extended_utf8_to_uv, __VA_ARGS__)
#endif
#if defined(fatal_warner) && !defined(Perl_fatal_warner)
#define Perl_fatal_warner(...) GOPERL_LONG_CALL(fatal_warner, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(fbm_compile) && !defined(Perl_fbm_compile)
#define Perl_fbm_compile(...) GOPERL_LONG_CALL(fbm_compile, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(fbm_instr) && !defined(Perl_fbm_instr)
#define Perl_fbm_instr(...) GOPERL_LONG_CALL(fbm_instr, GOPERL_THX_DROP(__VA_ARGS__))
#endif
static inline __attribute__((always_inline, unused)) I32
goperl_long_filter_read(int goperl_a0, SV * goperl_a1, int goperl_a2) { return filter_read(goperl_a0, goperl_a1, goperl_a2); }
extern inline __attribute__((gnu_inline, always_inline, unused)) I32
Perl_filter_read(pTHX_ int goperl_a0, SV * goperl_a1, int goperl_a2) { return goperl_long_filter_read(goperl_a0, goperl_a1, goperl_a2); }
static inline __attribute__((always_inline, unused)) CV *
goperl_long_find_runcv(U32 * goperl_a0) { return find_runcv(goperl_a0); }
extern inline __attribute__((gnu_inline, always_inline, unused)) CV *
Perl_find_runcv(pTHX_ U32 * goperl_a0) { return goperl_long_find_runcv(goperl_a0); }
#if defined(find_rundefsv) && !defined(Perl_find_rundefsv)
#define Perl_find_rundefsv(...) GOPERL_LONG_CALL(find_rundefsv, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(foldEQ) && !defined(Perl_foldEQ)
#define Perl_foldEQ(...) GOPERL_LONG_CALL(foldEQ, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(foldEQ_locale) && !defined(Perl_foldEQ_locale)
#define Perl_foldEQ_locale(...) GOPERL_LONG_CALL(foldEQ_locale, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(foldEQ_utf8) && !defined(Perl_foldEQ_utf8)
#define Perl_foldEQ_utf8(...) GOPERL_LONG_CALL(foldEQ_utf8, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(forbid_outofblock_ops) && !defined(Perl_forbid_outofblock_ops)
#define Perl_forbid_outofblock_ops(...) GOPERL_LONG_CALL(forbid_outofblock_ops, GOPERL_THX_DROP(__VA_ARGS__))
#endif
static inline __attribute__((always_inline, unused)) AV *
goperl_long_get_av(const char * goperl_a0, I32 goperl_a1) { return get_av(goperl_a0, goperl_a1); }
extern inline __attribute__((gnu_inline, always_inline, unused)) AV *
Perl_get_av(pTHX_ const char * goperl_a0, I32 goperl_a1) { return goperl_long_get_av(goperl_a0, goperl_a1); }
static inline __attribute__((always_inline, unused)) CV *
goperl_long_get_cv(const char * goperl_a0, I32 goperl_a1) { return get_cv(goperl_a0, goperl_a1); }
extern inline __attribute__((gnu_inline, always_inline, unused)) CV *
Perl_get_cv(pTHX_ const char * goperl_a0, I32 goperl_a1) { return goperl_long_get_cv(goperl_a0, goperl_a1); }
static inline __attribute__((always_inline, unused)) CV *
goperl_long_get_cvn_flags(const char * goperl_a0, STRLEN goperl_a1, I32 goperl_a2) { return get_cvn_flags(goperl_a0, goperl_a1, goperl_a2); }
extern inline __attribute__((gnu_inline, always_inline, unused)) CV *
Perl_get_cvn_flags(pTHX_ const char * goperl_a0, STRLEN goperl_a1, I32 goperl_a2) { return goperl_long_get_cvn_flags(goperl_a0, goperl_a1, goperl_a2); }
static inline __attribute__((always_inline, unused)) int
goperl_long_getcwd_sv(SV * goperl_a0) { return getcwd_sv(goperl_a0); }
extern inline __attribute__((gnu_inline, always_inline, unused)) int
Perl_getcwd_sv(pTHX_ SV * goperl_a0) { return goperl_long_getcwd_sv(goperl_a0); }
static inline __attribute__((always_inline, unused)) HV *
goperl_long_get_hv(const char * goperl_a0, I32 goperl_a1) { return get_hv(goperl_a0, goperl_a1); }
extern inline __attribute__((gnu_inline, always_inline, unused)) HV *
Perl_get_hv(pTHX_ const char * goperl_a0, I32 goperl_a1) { return goperl_long_get_hv(goperl_a0, goperl_a1); }
#if defined(get_op_descs) && !defined(Perl_get_op_descs)
#define Perl_get_op_descs(...) GOPERL_LONG_CALL(get_op_descs, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(get_op_names) && !defined(Perl_get_op_names)
#define Perl_get_op_names(...) GOPERL_LONG_CALL(get_op_names, GOPERL_THX_DROP(__VA_ARGS__))
#endif
static inline __attribute__((always_inline, unused)) SV *
goperl_long_get_sv(const char * goperl_a0, I32 goperl_a1) { return get_sv(goperl_a0, goperl_a1); }
extern inline __attribute__((gnu_inline, always_inline, unused)) SV *
Perl_get_sv(pTHX_ const char * goperl_a0, I32 goperl_a1) { return goperl_long_get_sv(goperl_a0, goperl_a1); }
#if defined(grok_atoUV) && !defined(Perl_grok_atoUV)
#define Perl_grok_atoUV(...) GOPERL_LONG_CALL(grok_atoUV, __VA_ARGS__)
#endif
static inline __attribute__((always_inline, unused)) UV
goperl_long_grok_bin(const char * goperl_a0, STRLEN * goperl_a1, I32 * goperl_a2, NV * goperl_a3) { return grok_bin(goperl_a0, goperl_a1, goperl_a2, goperl_a3); }
extern inline __attribute__((gnu_inline, always_inline, unused)) UV
Perl_grok_bin(pTHX_ const char * goperl_a0, STRLEN * goperl_a1, I32 * goperl_a2, NV * goperl_a3) { return goperl_long_grok_bin(goperl_a0, goperl_a1, goperl_a2, goperl_a3); }
static inline __attribute__((always_inline, unused)) UV
goperl_long_grok_hex(const char * goperl_a0, STRLEN * goperl_a1, I32 * goperl_a2, NV * goperl_a3) { return grok_hex(goperl_a0, goperl_a1, goperl_a2, goperl_a3); }
extern inline __attribute__((gnu_inline, always_inline, unused)) UV
Perl_grok_hex(pTHX_ const char * goperl_a0, STRLEN * goperl_a1, I32 * goperl_a2, NV * goperl_a3) { return goperl_long_grok_hex(goperl_a0, goperl_a1, goperl_a2, goperl_a3); }
#if defined(grok_infnan) && !defined(Perl_grok_infnan)
#define Perl_grok_infnan(...) GOPERL_LONG_CALL(grok_infnan, GOPERL_THX_DROP(__VA_ARGS__))
#endif
static inline __attribute__((always_inline, unused)) int
goperl_long_grok_number(const char * goperl_a0, STRLEN goperl_a1, UV * goperl_a2) { return grok_number(goperl_a0, goperl_a1, goperl_a2); }
extern inline __attribute__((gnu_inline, always_inline, unused)) int
Perl_grok_number(pTHX_ const char * goperl_a0, STRLEN goperl_a1, UV * goperl_a2) { return goperl_long_grok_number(goperl_a0, goperl_a1, goperl_a2); }
static inline __attribute__((always_inline, unused)) int
goperl_long_grok_number_flags(const char * goperl_a0, STRLEN goperl_a1, UV * goperl_a2, U32 goperl_a3) { return grok_number_flags(goperl_a0, goperl_a1, goperl_a2, goperl_a3); }
extern inline __attribute__((gnu_inline, always_inline, unused)) int
Perl_grok_number_flags(pTHX_ const char * goperl_a0, STRLEN goperl_a1, UV * goperl_a2, U32 goperl_a3) { return goperl_long_grok_number_flags(goperl_a0, goperl_a1, goperl_a2, goperl_a3); }
#if defined(grok_numeric_radix) && !defined(Perl_grok_numeric_radix)
#define Perl_grok_numeric_radix(...) GOPERL_LONG_CALL(grok_numeric_radix, GOPERL_THX_DROP(__VA_ARGS__))
#endif
static inline __attribute__((always_inline, unused)) UV
goperl_long_grok_oct(const char * goperl_a0, STRLEN * goperl_a1, I32 * goperl_a2, NV * goperl_a3) { return grok_oct(goperl_a0, goperl_a1, goperl_a2, goperl_a3); }
extern inline __attribute__((gnu_inline, always_inline, unused)) UV
Perl_grok_oct(pTHX_ const char * goperl_a0, STRLEN * goperl_a1, I32 * goperl_a2, NV * goperl_a3) { return goperl_long_grok_oct(goperl_a0, goperl_a1, goperl_a2, goperl_a3); }
#if defined(gv_add_by_type) && !defined(Perl_gv_add_by_type)
#define Perl_gv_add_by_type(...) GOPERL_LONG_CALL(gv_add_by_type, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(Gv_AMupdate) && !defined(Perl_Gv_AMupdate)
#define Perl_Gv_AMupdate(...) GOPERL_LONG_CALL(Gv_AMupdate, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(gv_autoload_pv) && !defined(Perl_gv_autoload_pv)
#define Perl_gv_autoload_pv(...) GOPERL_LONG_CALL(gv_autoload_pv, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(gv_autoload_pvn) && !defined(Perl_gv_autoload_pvn)
#define Perl_gv_autoload_pvn(...) GOPERL_LONG_CALL(gv_autoload_pvn, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(gv_autoload_sv) && !defined(Perl_gv_autoload_sv)
#define Perl_gv_autoload_sv(...) GOPERL_LONG_CALL(gv_autoload_sv, GOPERL_THX_DROP(__VA_ARGS__))
#endif
static inline __attribute__((always_inline, unused)) GV *
goperl_long_gv_AVadd(GV * goperl_a0) { return gv_AVadd(goperl_a0); }
extern inline __attribute__((gnu_inline, always_inline, unused)) GV *
Perl_gv_AVadd(pTHX_ GV * goperl_a0) { return goperl_long_gv_AVadd(goperl_a0); }
#if defined(gv_const_sv) && !defined(Perl_gv_const_sv)
#define Perl_gv_const_sv(...) GOPERL_LONG_CALL(gv_const_sv, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(gv_dump) && !defined(Perl_gv_dump)
#define Perl_gv_dump(...) GOPERL_LONG_CALL(gv_dump, GOPERL_THX_DROP(__VA_ARGS__))
#endif
static inline __attribute__((always_inline, unused)) void
goperl_long_gv_efullname3(SV * goperl_a0, const GV * goperl_a1, const char * goperl_a2) { gv_efullname3(goperl_a0, goperl_a1, goperl_a2); }
extern inline __attribute__((gnu_inline, always_inline, unused)) void
Perl_gv_efullname3(pTHX_ SV * goperl_a0, const GV * goperl_a1, const char * goperl_a2) { goperl_long_gv_efullname3(goperl_a0, goperl_a1, goperl_a2); }
static inline __attribute__((always_inline, unused)) void
goperl_long_gv_efullname4(SV * goperl_a0, const GV * goperl_a1, const char * goperl_a2, bool goperl_a3) { gv_efullname4(goperl_a0, goperl_a1, goperl_a2, goperl_a3); }
extern inline __attribute__((gnu_inline, always_inline, unused)) void
Perl_gv_efullname4(pTHX_ SV * goperl_a0, const GV * goperl_a1, const char * goperl_a2, bool goperl_a3) { goperl_long_gv_efullname4(goperl_a0, goperl_a1, goperl_a2, goperl_a3); }
static inline __attribute__((always_inline, unused)) GV *
goperl_long_gv_fetchfile(const char * goperl_a0) { return gv_fetchfile(goperl_a0); }
extern inline __attribute__((gnu_inline, always_inline, unused)) GV *
Perl_gv_fetchfile(pTHX_ const char * goperl_a0) { return goperl_long_gv_fetchfile(goperl_a0); }
static inline __attribute__((always_inline, unused)) GV *
goperl_long_gv_fetchfile_flags(const char * const goperl_a0, const STRLEN goperl_a1, const U32 goperl_a2) { return gv_fetchfile_flags(goperl_a0, goperl_a1, goperl_a2); }
extern inline __attribute__((gnu_inline, always_inline, unused)) GV *
Perl_gv_fetchfile_flags(pTHX_ const char * const goperl_a0, const STRLEN goperl_a1, const U32 goperl_a2) { return goperl_long_gv_fetchfile_flags(goperl_a0, goperl_a1, goperl_a2); }
static inline __attribute__((always_inline, unused)) GV *
goperl_long_gv_fetchmeth(HV * goperl_a0, const char * goperl_a1, STRLEN goperl_a2, I32 goperl_a3) { return gv_fetchmeth(goperl_a0, goperl_a1, goperl_a2, goperl_a3); }
extern inline __attribute__((gnu_inline, always_inline, unused)) GV *
Perl_gv_fetchmeth(pTHX_ HV * goperl_a0, const char * goperl_a1, STRLEN goperl_a2, I32 goperl_a3) { return goperl_long_gv_fetchmeth(goperl_a0, goperl_a1, goperl_a2, goperl_a3); }
static inline __attribute__((always_inline, unused)) GV *
goperl_long_gv_fetchmeth_autoload(HV * goperl_a0, const char * goperl_a1, STRLEN goperl_a2, I32 goperl_a3) { return gv_fetchmeth_autoload(goperl_a0, goperl_a1, goperl_a2, goperl_a3); }
extern inline __attribute__((gnu_inline, always_inline, unused)) GV *
Perl_gv_fetchmeth_autoload(pTHX_ HV * goperl_a0, const char * goperl_a1, STRLEN goperl_a2, I32 goperl_a3) { return goperl_long_gv_fetchmeth_autoload(goperl_a0, goperl_a1, goperl_a2, goperl_a3); }
static inline __attribute__((always_inline, unused)) GV *
goperl_long_gv_fetchmethod(HV * goperl_a0, const char * goperl_a1) { return gv_fetchmethod(goperl_a0, goperl_a1); }
extern inline __attribute__((gnu_inline, always_inline, unused)) GV *
Perl_gv_fetchmethod(pTHX_ HV * goperl_a0, const char * goperl_a1) { return goperl_long_gv_fetchmethod(goperl_a0, goperl_a1); }
static inline __attribute__((always_inline, unused)) GV *
goperl_long_gv_fetchmethod_autoload(HV * goperl_a0, const char * goperl_a1, I32 goperl_a2) { return gv_fetchmethod_autoload(goperl_a0, goperl_a1, goperl_a2); }
extern inline __attribute__((gnu_inline, always_inline, unused)) GV *
Perl_gv_fetchmethod_autoload(pTHX_ HV * goperl_a0, const char * goperl_a1, I32 goperl_a2) { return goperl_long_gv_fetchmethod_autoload(goperl_a0, goperl_a1, goperl_a2); }
static inline __attribute__((always_inline, unused)) GV *
goperl_long_gv_fetchmethod_pv_flags(HV * goperl_a0, const char * goperl_a1, U32 goperl_a2) { return gv_fetchmethod_pv_flags(goperl_a0, goperl_a1, goperl_a2); }
extern inline __attribute__((gnu_inline, always_inline, unused)) GV *
Perl_gv_fetchmethod_pv_flags(pTHX_ HV * goperl_a0, const char * goperl_a1, U32 goperl_a2) { return goperl_long_gv_fetchmethod_pv_flags(goperl_a0, goperl_a1, goperl_a2); }
static inline __attribute__((always_inline, unused)) GV *
goperl_long_gv_fetchmethod_pvn_flags(HV * goperl_a0, const char * goperl_a1, const STRLEN goperl_a2, U32 goperl_a3) { return gv_fetchmethod_pvn_flags(goperl_a0, goperl_a1, goperl_a2, goperl_a3); }
extern inline __attribute__((gnu_inline, always_inline, unused)) GV *
Perl_gv_fetchmethod_pvn_flags(pTHX_ HV * goperl_a0, const char * goperl_a1, const STRLEN goperl_a2, U32 goperl_a3) { return goperl_long_gv_fetchmethod_pvn_flags(goperl_a0, goperl_a1, goperl_a2, goperl_a3); }
#if defined(gv_fetchmethod_sv_flags) && !defined(Perl_gv_fetchmethod_sv_flags)
#define Perl_gv_fetchmethod_sv_flags(...) GOPERL_LONG_CALL(gv_fetchmethod_sv_flags, GOPERL_THX_DROP(__VA_ARGS__))
#endif
static inline __attribute__((always_inline, unused)) GV *
goperl_long_gv_fetchmeth_pv(HV * goperl_a0, const char * goperl_a1, I32 goperl_a2, U32 goperl_a3) { return gv_fetchmeth_pv(goperl_a0, goperl_a1, goperl_a2, goperl_a3); }
extern inline __attribute__((gnu_inline, always_inline, unused)) GV *
Perl_gv_fetchmeth_pv(pTHX_ HV * goperl_a0, const char * goperl_a1, I32 goperl_a2, U32 goperl_a3) { return goperl_long_gv_fetchmeth_pv(goperl_a0, goperl_a1, goperl_a2, goperl_a3); }
#if defined(gv_fetchmeth_pv_autoload) && !defined(Perl_gv_fetchmeth_pv_autoload)
#define Perl_gv_fetchmeth_pv_autoload(...) GOPERL_LONG_CALL(gv_fetchmeth_pv_autoload, GOPERL_THX_DROP(__VA_ARGS__))
#endif
static inline __attribute__((always_inline, unused)) GV *
goperl_long_gv_fetchmeth_pvn(HV * goperl_a0, const char * goperl_a1, STRLEN goperl_a2, I32 goperl_a3, U32 goperl_a4) { return gv_fetchmeth_pvn(goperl_a0, goperl_a1, goperl_a2, goperl_a3, goperl_a4); }
extern inline __attribute__((gnu_inline, always_inline, unused)) GV *
Perl_gv_fetchmeth_pvn(pTHX_ HV * goperl_a0, const char * goperl_a1, STRLEN goperl_a2, I32 goperl_a3, U32 goperl_a4) { return goperl_long_gv_fetchmeth_pvn(goperl_a0, goperl_a1, goperl_a2, goperl_a3, goperl_a4); }
static inline __attribute__((always_inline, unused)) GV *
goperl_long_gv_fetchmeth_pvn_autoload(HV * goperl_a0, const char * goperl_a1, STRLEN goperl_a2, I32 goperl_a3, U32 goperl_a4) { return gv_fetchmeth_pvn_autoload(goperl_a0, goperl_a1, goperl_a2, goperl_a3, goperl_a4); }
extern inline __attribute__((gnu_inline, always_inline, unused)) GV *
Perl_gv_fetchmeth_pvn_autoload(pTHX_ HV * goperl_a0, const char * goperl_a1, STRLEN goperl_a2, I32 goperl_a3, U32 goperl_a4) { return goperl_long_gv_fetchmeth_pvn_autoload(goperl_a0, goperl_a1, goperl_a2, goperl_a3, goperl_a4); }
static inline __attribute__((always_inline, unused)) GV *
goperl_long_gv_fetchmeth_sv(HV * goperl_a0, SV * goperl_a1, I32 goperl_a2, U32 goperl_a3) { return gv_fetchmeth_sv(goperl_a0, goperl_a1, goperl_a2, goperl_a3); }
extern inline __attribute__((gnu_inline, always_inline, unused)) GV *
Perl_gv_fetchmeth_sv(pTHX_ HV * goperl_a0, SV * goperl_a1, I32 goperl_a2, U32 goperl_a3) { return goperl_long_gv_fetchmeth_sv(goperl_a0, goperl_a1, goperl_a2, goperl_a3); }
#if defined(gv_fetchmeth_sv_autoload) && !defined(Perl_gv_fetchmeth_sv_autoload)
#define Perl_gv_fetchmeth_sv_autoload(...) GOPERL_LONG_CALL(gv_fetchmeth_sv_autoload, GOPERL_THX_DROP(__VA_ARGS__))
#endif
static inline __attribute__((always_inline, unused)) GV *
goperl_long_gv_fetchpv(const char * goperl_a0, I32 goperl_a1, const svtype goperl_a2) { return gv_fetchpv(goperl_a0, goperl_a1, goperl_a2); }
extern inline __attribute__((gnu_inline, always_inline, unused)) GV *
Perl_gv_fetchpv(pTHX_ const char * goperl_a0, I32 goperl_a1, const svtype goperl_a2) { return goperl_long_gv_fetchpv(goperl_a0, goperl_a1, goperl_a2); }
static inline __attribute__((always_inline, unused)) GV *
goperl_long_gv_fetchpvn_flags(const char * goperl_a0, STRLEN goperl_a1, I32 goperl_a2, const svtype goperl_a3) { return gv_fetchpvn_flags(goperl_a0, goperl_a1, goperl_a2, goperl_a3); }
extern inline __attribute__((gnu_inline, always_inline, unused)) GV *
Perl_gv_fetchpvn_flags(pTHX_ const char * goperl_a0, STRLEN goperl_a1, I32 goperl_a2, const svtype goperl_a3) { return goperl_long_gv_fetchpvn_flags(goperl_a0, goperl_a1, goperl_a2, goperl_a3); }
static inline __attribute__((always_inline, unused)) GV *
goperl_long_gv_fetchsv(SV * goperl_a0, I32 goperl_a1, const svtype goperl_a2) { return gv_fetchsv(goperl_a0, goperl_a1, goperl_a2); }
extern inline __attribute__((gnu_inline, always_inline, unused)) GV *
Perl_gv_fetchsv(pTHX_ SV * goperl_a0, I32 goperl_a1, const svtype goperl_a2) { return goperl_long_gv_fetchsv(goperl_a0, goperl_a1, goperl_a2); }
static inline __attribute__((always_inline, unused)) void
goperl_long_gv_fullname3(SV * goperl_a0, const GV * goperl_a1, const char * goperl_a2) { gv_fullname3(goperl_a0, goperl_a1, goperl_a2); }
extern inline __attribute__((gnu_inline, always_inline, unused)) void
Perl_gv_fullname3(pTHX_ SV * goperl_a0, const GV * goperl_a1, const char * goperl_a2) { goperl_long_gv_fullname3(goperl_a0, goperl_a1, goperl_a2); }
static inline __attribute__((always_inline, unused)) void
goperl_long_gv_fullname4(SV * goperl_a0, const GV * goperl_a1, const char * goperl_a2, bool goperl_a3) { gv_fullname4(goperl_a0, goperl_a1, goperl_a2, goperl_a3); }
extern inline __attribute__((gnu_inline, always_inline, unused)) void
Perl_gv_fullname4(pTHX_ SV * goperl_a0, const GV * goperl_a1, const char * goperl_a2, bool goperl_a3) { goperl_long_gv_fullname4(goperl_a0, goperl_a1, goperl_a2, goperl_a3); }
static inline __attribute__((always_inline, unused)) GV *
goperl_long_gv_HVadd(GV * goperl_a0) { return gv_HVadd(goperl_a0); }
extern inline __attribute__((gnu_inline, always_inline, unused)) GV *
Perl_gv_HVadd(pTHX_ GV * goperl_a0) { return goperl_long_gv_HVadd(goperl_a0); }
static inline __attribute__((always_inline, unused)) void
goperl_long_gv_init(GV * goperl_a0, HV * goperl_a1, const char * goperl_a2, STRLEN goperl_a3, int goperl_a4) { gv_init(goperl_a0, goperl_a1, goperl_a2, goperl_a3, goperl_a4); }
extern inline __attribute__((gnu_inline, always_inline, unused)) void
Perl_gv_init(pTHX_ GV * goperl_a0, HV * goperl_a1, const char * goperl_a2, STRLEN goperl_a3, int goperl_a4) { goperl_long_gv_init(goperl_a0, goperl_a1, goperl_a2, goperl_a3, goperl_a4); }
#if defined(gv_init_pv) && !defined(Perl_gv_init_pv)
#define Perl_gv_init_pv(...) GOPERL_LONG_CALL(gv_init_pv, GOPERL_THX_DROP(__VA_ARGS__))
#endif
static inline __attribute__((always_inline, unused)) void
goperl_long_gv_init_pvn(GV * goperl_a0, HV * goperl_a1, const char * goperl_a2, STRLEN goperl_a3, U32 goperl_a4) { gv_init_pvn(goperl_a0, goperl_a1, goperl_a2, goperl_a3, goperl_a4); }
extern inline __attribute__((gnu_inline, always_inline, unused)) void
Perl_gv_init_pvn(pTHX_ GV * goperl_a0, HV * goperl_a1, const char * goperl_a2, STRLEN goperl_a3, U32 goperl_a4) { goperl_long_gv_init_pvn(goperl_a0, goperl_a1, goperl_a2, goperl_a3, goperl_a4); }
static inline __attribute__((always_inline, unused)) void
goperl_long_gv_init_sv(GV * goperl_a0, HV * goperl_a1, SV * goperl_a2, U32 goperl_a3) { gv_init_sv(goperl_a0, goperl_a1, goperl_a2, goperl_a3); }
extern inline __attribute__((gnu_inline, always_inline, unused)) void
Perl_gv_init_sv(pTHX_ GV * goperl_a0, HV * goperl_a1, SV * goperl_a2, U32 goperl_a3) { goperl_long_gv_init_sv(goperl_a0, goperl_a1, goperl_a2, goperl_a3); }
static inline __attribute__((always_inline, unused)) GV *
goperl_long_gv_IOadd(GV * goperl_a0) { return gv_IOadd(goperl_a0); }
extern inline __attribute__((gnu_inline, always_inline, unused)) GV *
Perl_gv_IOadd(pTHX_ GV * goperl_a0) { return goperl_long_gv_IOadd(goperl_a0); }
static inline __attribute__((always_inline, unused)) void
goperl_long_gv_name_set(GV * goperl_a0, const char * goperl_a1, U32 goperl_a2, U32 goperl_a3) { gv_name_set(goperl_a0, goperl_a1, goperl_a2, goperl_a3); }
extern inline __attribute__((gnu_inline, always_inline, unused)) void
Perl_gv_name_set(pTHX_ GV * goperl_a0, const char * goperl_a1, U32 goperl_a2, U32 goperl_a3) { goperl_long_gv_name_set(goperl_a0, goperl_a1, goperl_a2, goperl_a3); }
static inline __attribute__((always_inline, unused)) HV *
goperl_long_gv_stashpv(const char * goperl_a0, I32 goperl_a1) { return gv_stashpv(goperl_a0, goperl_a1); }
extern inline __attribute__((gnu_inline, always_inline, unused)) HV *
Perl_gv_stashpv(pTHX_ const char * goperl_a0, I32 goperl_a1) { return goperl_long_gv_stashpv(goperl_a0, goperl_a1); }
static inline __attribute__((always_inline, unused)) HV *
goperl_long_gv_stashpvn(const char * goperl_a0, U32 goperl_a1, I32 goperl_a2) { return gv_stashpvn(goperl_a0, goperl_a1, goperl_a2); }
extern inline __attribute__((gnu_inline, always_inline, unused)) HV *
Perl_gv_stashpvn(pTHX_ const char * goperl_a0, U32 goperl_a1, I32 goperl_a2) { return goperl_long_gv_stashpvn(goperl_a0, goperl_a1, goperl_a2); }
static inline __attribute__((always_inline, unused)) HV *
goperl_long_gv_stashsv(SV * goperl_a0, I32 goperl_a1) { return gv_stashsv(goperl_a0, goperl_a1); }
extern inline __attribute__((gnu_inline, always_inline, unused)) HV *
Perl_gv_stashsv(pTHX_ SV * goperl_a0, I32 goperl_a1) { return goperl_long_gv_stashsv(goperl_a0, goperl_a1); }
static inline __attribute__((always_inline, unused)) GV *
goperl_long_gv_SVadd(GV * goperl_a0) { return gv_SVadd(goperl_a0); }
extern inline __attribute__((gnu_inline, always_inline, unused)) GV *
Perl_gv_SVadd(pTHX_ GV * goperl_a0) { return goperl_long_gv_SVadd(goperl_a0); }
#if defined(hv_bucket_ratio) && !defined(Perl_hv_bucket_ratio)
#define Perl_hv_bucket_ratio(...) GOPERL_LONG_CALL(hv_bucket_ratio, GOPERL_THX_DROP(__VA_ARGS__))
#endif
static inline __attribute__((always_inline, unused)) void
goperl_long_hv_clear(HV * goperl_a0) { hv_clear(goperl_a0); }
extern inline __attribute__((gnu_inline, always_inline, unused)) void
Perl_hv_clear(pTHX_ HV * goperl_a0) { goperl_long_hv_clear(goperl_a0); }
#if defined(hv_clear_placeholders) && !defined(Perl_hv_clear_placeholders)
#define Perl_hv_clear_placeholders(...) GOPERL_LONG_CALL(hv_clear_placeholders, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(hv_copy_hints_hv) && !defined(Perl_hv_copy_hints_hv)
#define Perl_hv_copy_hints_hv(...) GOPERL_LONG_CALL(hv_copy_hints_hv, GOPERL_THX_DROP(__VA_ARGS__))
#endif
static inline __attribute__((always_inline, unused)) SV *
goperl_long_hv_delete(HV * goperl_a0, const char * goperl_a1, I32 goperl_a2, I32 goperl_a3) { return hv_delete(goperl_a0, goperl_a1, goperl_a2, goperl_a3); }
extern inline __attribute__((gnu_inline, always_inline, unused)) SV *
Perl_hv_delete(pTHX_ HV * goperl_a0, const char * goperl_a1, I32 goperl_a2, I32 goperl_a3) { return goperl_long_hv_delete(goperl_a0, goperl_a1, goperl_a2, goperl_a3); }
static inline __attribute__((always_inline, unused)) SV *
goperl_long_hv_delete_ent(HV * goperl_a0, SV * goperl_a1, I32 goperl_a2, U32 goperl_a3) { return hv_delete_ent(goperl_a0, goperl_a1, goperl_a2, goperl_a3); }
extern inline __attribute__((gnu_inline, always_inline, unused)) SV *
Perl_hv_delete_ent(pTHX_ HV * goperl_a0, SV * goperl_a1, I32 goperl_a2, U32 goperl_a3) { return goperl_long_hv_delete_ent(goperl_a0, goperl_a1, goperl_a2, goperl_a3); }
#if defined(hv_dump) && !defined(Perl_hv_dump)
#define Perl_hv_dump(...) GOPERL_LONG_CALL(hv_dump, GOPERL_THX_DROP(__VA_ARGS__))
#endif
static inline __attribute__((always_inline, unused)) bool
goperl_long_hv_exists(HV * goperl_a0, const char * goperl_a1, I32 goperl_a2) { return hv_exists(goperl_a0, goperl_a1, goperl_a2); }
extern inline __attribute__((gnu_inline, always_inline, unused)) bool
Perl_hv_exists(pTHX_ HV * goperl_a0, const char * goperl_a1, I32 goperl_a2) { return goperl_long_hv_exists(goperl_a0, goperl_a1, goperl_a2); }
static inline __attribute__((always_inline, unused)) bool
goperl_long_hv_exists_ent(HV * goperl_a0, SV * goperl_a1, U32 goperl_a2) { return hv_exists_ent(goperl_a0, goperl_a1, goperl_a2); }
extern inline __attribute__((gnu_inline, always_inline, unused)) bool
Perl_hv_exists_ent(pTHX_ HV * goperl_a0, SV * goperl_a1, U32 goperl_a2) { return goperl_long_hv_exists_ent(goperl_a0, goperl_a1, goperl_a2); }
static inline __attribute__((always_inline, unused)) SV **
goperl_long_hv_fetch(HV * goperl_a0, const char * goperl_a1, I32 goperl_a2, I32 goperl_a3) { return hv_fetch(goperl_a0, goperl_a1, goperl_a2, goperl_a3); }
extern inline __attribute__((gnu_inline, always_inline, unused)) SV **
Perl_hv_fetch(pTHX_ HV * goperl_a0, const char * goperl_a1, I32 goperl_a2, I32 goperl_a3) { return goperl_long_hv_fetch(goperl_a0, goperl_a1, goperl_a2, goperl_a3); }
static inline __attribute__((always_inline, unused)) HE *
goperl_long_hv_fetch_ent(HV * goperl_a0, SV * goperl_a1, I32 goperl_a2, U32 goperl_a3) { return hv_fetch_ent(goperl_a0, goperl_a1, goperl_a2, goperl_a3); }
extern inline __attribute__((gnu_inline, always_inline, unused)) HE *
Perl_hv_fetch_ent(pTHX_ HV * goperl_a0, SV * goperl_a1, I32 goperl_a2, U32 goperl_a3) { return goperl_long_hv_fetch_ent(goperl_a0, goperl_a1, goperl_a2, goperl_a3); }
static inline __attribute__((always_inline, unused)) I32
goperl_long_hv_iterinit(HV * goperl_a0) { return hv_iterinit(goperl_a0); }
extern inline __attribute__((gnu_inline, always_inline, unused)) I32
Perl_hv_iterinit(pTHX_ HV * goperl_a0) { return goperl_long_hv_iterinit(goperl_a0); }
static inline __attribute__((always_inline, unused)) char *
goperl_long_hv_iterkey(HE * goperl_a0, I32 * goperl_a1) { return hv_iterkey(goperl_a0, goperl_a1); }
extern inline __attribute__((gnu_inline, always_inline, unused)) char *
Perl_hv_iterkey(pTHX_ HE * goperl_a0, I32 * goperl_a1) { return goperl_long_hv_iterkey(goperl_a0, goperl_a1); }
static inline __attribute__((always_inline, unused)) SV *
goperl_long_hv_iterkeysv(HE * goperl_a0) { return hv_iterkeysv(goperl_a0); }
extern inline __attribute__((gnu_inline, always_inline, unused)) SV *
Perl_hv_iterkeysv(pTHX_ HE * goperl_a0) { return goperl_long_hv_iterkeysv(goperl_a0); }
static inline __attribute__((always_inline, unused)) HE *
goperl_long_hv_iternext(HV * goperl_a0) { return hv_iternext(goperl_a0); }
extern inline __attribute__((gnu_inline, always_inline, unused)) HE *
Perl_hv_iternext(pTHX_ HV * goperl_a0) { return goperl_long_hv_iternext(goperl_a0); }
static inline __attribute__((always_inline, unused)) HE *
goperl_long_hv_iternext_flags(HV * goperl_a0, I32 goperl_a1) { return hv_iternext_flags(goperl_a0, goperl_a1); }
extern inline __attribute__((gnu_inline, always_inline, unused)) HE *
Perl_hv_iternext_flags(pTHX_ HV * goperl_a0, I32 goperl_a1) { return goperl_long_hv_iternext_flags(goperl_a0, goperl_a1); }
static inline __attribute__((always_inline, unused)) SV *
goperl_long_hv_iternextsv(HV * goperl_a0, char ** goperl_a1, I32 * goperl_a2) { return hv_iternextsv(goperl_a0, goperl_a1, goperl_a2); }
extern inline __attribute__((gnu_inline, always_inline, unused)) SV *
Perl_hv_iternextsv(pTHX_ HV * goperl_a0, char ** goperl_a1, I32 * goperl_a2) { return goperl_long_hv_iternextsv(goperl_a0, goperl_a1, goperl_a2); }
static inline __attribute__((always_inline, unused)) SV *
goperl_long_hv_iterval(HV * goperl_a0, HE * goperl_a1) { return hv_iterval(goperl_a0, goperl_a1); }
extern inline __attribute__((gnu_inline, always_inline, unused)) SV *
Perl_hv_iterval(pTHX_ HV * goperl_a0, HE * goperl_a1) { return goperl_long_hv_iterval(goperl_a0, goperl_a1); }
static inline __attribute__((always_inline, unused)) void
goperl_long_hv_ksplit(HV * goperl_a0, IV goperl_a1) { hv_ksplit(goperl_a0, goperl_a1); }
extern inline __attribute__((gnu_inline, always_inline, unused)) void
Perl_hv_ksplit(pTHX_ HV * goperl_a0, IV goperl_a1) { goperl_long_hv_ksplit(goperl_a0, goperl_a1); }
static inline __attribute__((always_inline, unused)) void
goperl_long_hv_magic(HV * goperl_a0, GV * goperl_a1, int goperl_a2) { hv_magic(goperl_a0, goperl_a1, goperl_a2); }
extern inline __attribute__((gnu_inline, always_inline, unused)) void
Perl_hv_magic(pTHX_ HV * goperl_a0, GV * goperl_a1, int goperl_a2) { goperl_long_hv_magic(goperl_a0, goperl_a1, goperl_a2); }
#if defined(hv_name_set) && !defined(Perl_hv_name_set)
#define Perl_hv_name_set(...) GOPERL_LONG_CALL(hv_name_set, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(hv_scalar) && !defined(Perl_hv_scalar)
#define Perl_hv_scalar(...) GOPERL_LONG_CALL(hv_scalar, GOPERL_THX_DROP(__VA_ARGS__))
#endif
static inline __attribute__((always_inline, unused)) SV **
goperl_long_hv_store(HV * goperl_a0, const char * goperl_a1, I32 goperl_a2, SV * goperl_a3, U32 goperl_a4) { return hv_store(goperl_a0, goperl_a1, goperl_a2, goperl_a3, goperl_a4); }
extern inline __attribute__((gnu_inline, always_inline, unused)) SV **
Perl_hv_store(pTHX_ HV * goperl_a0, const char * goperl_a1, I32 goperl_a2, SV * goperl_a3, U32 goperl_a4) { return goperl_long_hv_store(goperl_a0, goperl_a1, goperl_a2, goperl_a3, goperl_a4); }
static inline __attribute__((always_inline, unused)) HE *
goperl_long_hv_store_ent(HV * goperl_a0, SV * goperl_a1, SV * goperl_a2, U32 goperl_a3) { return hv_store_ent(goperl_a0, goperl_a1, goperl_a2, goperl_a3); }
extern inline __attribute__((gnu_inline, always_inline, unused)) HE *
Perl_hv_store_ent(pTHX_ HV * goperl_a0, SV * goperl_a1, SV * goperl_a2, U32 goperl_a3) { return goperl_long_hv_store_ent(goperl_a0, goperl_a1, goperl_a2, goperl_a3); }
#if defined(hv_store_flags) && !defined(Perl_hv_store_flags)
#define Perl_hv_store_flags(...) GOPERL_LONG_CALL(hv_store_flags, GOPERL_THX_DROP(__VA_ARGS__))
#endif
static inline __attribute__((always_inline, unused)) void
goperl_long_hv_undef(HV * goperl_a0) { hv_undef(goperl_a0); }
extern inline __attribute__((gnu_inline, always_inline, unused)) void
Perl_hv_undef(pTHX_ HV * goperl_a0) { goperl_long_hv_undef(goperl_a0); }
#if defined(ibcmp) && !defined(Perl_ibcmp)
#define Perl_ibcmp(...) GOPERL_LONG_CALL(ibcmp, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(ibcmp_locale) && !defined(Perl_ibcmp_locale)
#define Perl_ibcmp_locale(...) GOPERL_LONG_CALL(ibcmp_locale, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(ibcmp_utf8) && !defined(Perl_ibcmp_utf8)
#define Perl_ibcmp_utf8(...) GOPERL_LONG_CALL(ibcmp_utf8, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(instr) && !defined(Perl_instr)
#define Perl_instr(...) GOPERL_LONG_CALL(instr, __VA_ARGS__)
#endif
#if defined(isC9_STRICT_UTF8_CHAR) && !defined(Perl_isC9_STRICT_UTF8_CHAR)
#define Perl_isC9_STRICT_UTF8_CHAR(...) GOPERL_LONG_CALL(isC9_STRICT_UTF8_CHAR, __VA_ARGS__)
#endif
#if defined(is_c9strict_utf8_string) && !defined(Perl_is_c9strict_utf8_string)
#define Perl_is_c9strict_utf8_string(...) GOPERL_LONG_CALL(is_c9strict_utf8_string, __VA_ARGS__)
#endif
#if defined(is_c9strict_utf8_string_loc) && !defined(Perl_is_c9strict_utf8_string_loc)
#define Perl_is_c9strict_utf8_string_loc(...) GOPERL_LONG_CALL(is_c9strict_utf8_string_loc, __VA_ARGS__)
#endif
#if defined(is_c9strict_utf8_string_loclen) && !defined(Perl_is_c9strict_utf8_string_loclen)
#define Perl_is_c9strict_utf8_string_loclen(...) GOPERL_LONG_CALL(is_c9strict_utf8_string_loclen, __VA_ARGS__)
#endif
#if defined(isinfnan) && !defined(Perl_isinfnan)
#define Perl_isinfnan(...) GOPERL_LONG_CALL(isinfnan, __VA_ARGS__)
#endif
#if defined(is_lvalue_sub) && !defined(Perl_is_lvalue_sub)
#define Perl_is_lvalue_sub(...) GOPERL_LONG_CALL(is_lvalue_sub, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(is_safe_syscall) && !defined(Perl_is_safe_syscall)
#define Perl_is_safe_syscall(...) GOPERL_LONG_CALL(is_safe_syscall, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(isSTRICT_UTF8_CHAR) && !defined(Perl_isSTRICT_UTF8_CHAR)
#define Perl_isSTRICT_UTF8_CHAR(...) GOPERL_LONG_CALL(isSTRICT_UTF8_CHAR, __VA_ARGS__)
#endif
#if defined(is_strict_utf8_string) && !defined(Perl_is_strict_utf8_string)
#define Perl_is_strict_utf8_string(...) GOPERL_LONG_CALL(is_strict_utf8_string, __VA_ARGS__)
#endif
#if defined(is_strict_utf8_string_loc) && !defined(Perl_is_strict_utf8_string_loc)
#define Perl_is_strict_utf8_string_loc(...) GOPERL_LONG_CALL(is_strict_utf8_string_loc, __VA_ARGS__)
#endif
#if defined(is_strict_utf8_string_loclen) && !defined(Perl_is_strict_utf8_string_loclen)
#define Perl_is_strict_utf8_string_loclen(...) GOPERL_LONG_CALL(is_strict_utf8_string_loclen, __VA_ARGS__)
#endif
#if defined(isUTF8_CHAR) && !defined(Perl_isUTF8_CHAR)
#define Perl_isUTF8_CHAR(...) GOPERL_LONG_CALL(isUTF8_CHAR, __VA_ARGS__)
#endif
#if defined(is_utf8_char_buf) && !defined(Perl_is_utf8_char_buf)
#define Perl_is_utf8_char_buf(...) GOPERL_LONG_CALL(is_utf8_char_buf, __VA_ARGS__)
#endif
#if defined(isUTF8_CHAR_flags) && !defined(Perl_isUTF8_CHAR_flags)
#define Perl_isUTF8_CHAR_flags(...) GOPERL_LONG_CALL(isUTF8_CHAR_flags, __VA_ARGS__)
#endif
#if defined(is_utf8_fixed_width_buf_flags) && !defined(Perl_is_utf8_fixed_width_buf_flags)
#define Perl_is_utf8_fixed_width_buf_flags(...) GOPERL_LONG_CALL(is_utf8_fixed_width_buf_flags, __VA_ARGS__)
#endif
#if defined(is_utf8_fixed_width_buf_loc_flags) && !defined(Perl_is_utf8_fixed_width_buf_loc_flags)
#define Perl_is_utf8_fixed_width_buf_loc_flags(...) GOPERL_LONG_CALL(is_utf8_fixed_width_buf_loc_flags, __VA_ARGS__)
#endif
#if defined(is_utf8_fixed_width_buf_loclen_flags) && !defined(Perl_is_utf8_fixed_width_buf_loclen_flags)
#define Perl_is_utf8_fixed_width_buf_loclen_flags(...) GOPERL_LONG_CALL(is_utf8_fixed_width_buf_loclen_flags, __VA_ARGS__)
#endif
#if defined(is_utf8_invariant_string_loc) && !defined(Perl_is_utf8_invariant_string_loc)
#define Perl_is_utf8_invariant_string_loc(...) GOPERL_LONG_CALL(is_utf8_invariant_string_loc, __VA_ARGS__)
#endif
static inline __attribute__((always_inline, unused)) bool
goperl_long_is_utf8_string(const U8 * goperl_a0, STRLEN goperl_a1) { return is_utf8_string(goperl_a0, goperl_a1); }
extern inline __attribute__((gnu_inline, always_inline, unused)) bool
Perl_is_utf8_string(const U8 * goperl_a0, STRLEN goperl_a1) { return goperl_long_is_utf8_string(goperl_a0, goperl_a1); }
#if defined(is_utf8_string_flags) && !defined(Perl_is_utf8_string_flags)
#define Perl_is_utf8_string_flags(...) GOPERL_LONG_CALL(is_utf8_string_flags, __VA_ARGS__)
#endif
static inline __attribute__((always_inline, unused)) bool
goperl_long_is_utf8_string_loc(const U8 * goperl_a0, const STRLEN goperl_a1, const U8 ** goperl_a2) { return is_utf8_string_loc(goperl_a0, goperl_a1, goperl_a2); }
extern inline __attribute__((gnu_inline, always_inline, unused)) bool
Perl_is_utf8_string_loc(const U8 * goperl_a0, const STRLEN goperl_a1, const U8 ** goperl_a2) { return goperl_long_is_utf8_string_loc(goperl_a0, goperl_a1, goperl_a2); }
#if defined(is_utf8_string_loc_flags) && !defined(Perl_is_utf8_string_loc_flags)
#define Perl_is_utf8_string_loc_flags(...) GOPERL_LONG_CALL(is_utf8_string_loc_flags, __VA_ARGS__)
#endif
static inline __attribute__((always_inline, unused)) bool
goperl_long_is_utf8_string_loclen(const U8 * goperl_a0, STRLEN goperl_a1, const U8 ** goperl_a2, STRLEN * goperl_a3) { return is_utf8_string_loclen(goperl_a0, goperl_a1, goperl_a2, goperl_a3); }
extern inline __attribute__((gnu_inline, always_inline, unused)) bool
Perl_is_utf8_string_loclen(const U8 * goperl_a0, STRLEN goperl_a1, const U8 ** goperl_a2, STRLEN * goperl_a3) { return goperl_long_is_utf8_string_loclen(goperl_a0, goperl_a1, goperl_a2, goperl_a3); }
#if defined(is_utf8_string_loclen_flags) && !defined(Perl_is_utf8_string_loclen_flags)
#define Perl_is_utf8_string_loclen_flags(...) GOPERL_LONG_CALL(is_utf8_string_loclen_flags, __VA_ARGS__)
#endif
#if defined(is_utf8_valid_partial_char) && !defined(Perl_is_utf8_valid_partial_char)
#define Perl_is_utf8_valid_partial_char(...) GOPERL_LONG_CALL(is_utf8_valid_partial_char, __VA_ARGS__)
#endif
#if defined(is_utf8_valid_partial_char_flags) && !defined(Perl_is_utf8_valid_partial_char_flags)
#define Perl_is_utf8_valid_partial_char_flags(...) GOPERL_LONG_CALL(is_utf8_valid_partial_char_flags, __VA_ARGS__)
#endif
#if defined(leave_adjust_stacks) && !defined(Perl_leave_adjust_stacks)
#define Perl_leave_adjust_stacks(...) GOPERL_LONG_CALL(leave_adjust_stacks, GOPERL_THX_DROP(__VA_ARGS__))
#endif
static inline __attribute__((always_inline, unused)) bool
goperl_long_lex_bufutf8(void) { return lex_bufutf8(); }
extern inline __attribute__((gnu_inline, always_inline, unused)) bool
Perl_lex_bufutf8(pTHX) { return goperl_long_lex_bufutf8(); }
static inline __attribute__((always_inline, unused)) void
goperl_long_lex_discard_to(char * goperl_a0) { lex_discard_to(goperl_a0); }
extern inline __attribute__((gnu_inline, always_inline, unused)) void
Perl_lex_discard_to(pTHX_ char * goperl_a0) { goperl_long_lex_discard_to(goperl_a0); }
static inline __attribute__((always_inline, unused)) char *
goperl_long_lex_grow_linestr(STRLEN goperl_a0) { return lex_grow_linestr(goperl_a0); }
extern inline __attribute__((gnu_inline, always_inline, unused)) char *
Perl_lex_grow_linestr(pTHX_ STRLEN goperl_a0) { return goperl_long_lex_grow_linestr(goperl_a0); }
static inline __attribute__((always_inline, unused)) bool
goperl_long_lex_next_chunk(U32 goperl_a0) { return lex_next_chunk(goperl_a0); }
extern inline __attribute__((gnu_inline, always_inline, unused)) bool
Perl_lex_next_chunk(pTHX_ U32 goperl_a0) { return goperl_long_lex_next_chunk(goperl_a0); }
static inline __attribute__((always_inline, unused)) I32
goperl_long_lex_peek_unichar(U32 goperl_a0) { return lex_peek_unichar(goperl_a0); }
extern inline __attribute__((gnu_inline, always_inline, unused)) I32
Perl_lex_peek_unichar(pTHX_ U32 goperl_a0) { return goperl_long_lex_peek_unichar(goperl_a0); }
static inline __attribute__((always_inline, unused)) void
goperl_long_lex_read_space(U32 goperl_a0) { lex_read_space(goperl_a0); }
extern inline __attribute__((gnu_inline, always_inline, unused)) void
Perl_lex_read_space(pTHX_ U32 goperl_a0) { goperl_long_lex_read_space(goperl_a0); }
static inline __attribute__((always_inline, unused)) void
goperl_long_lex_read_to(char * goperl_a0) { lex_read_to(goperl_a0); }
extern inline __attribute__((gnu_inline, always_inline, unused)) void
Perl_lex_read_to(pTHX_ char * goperl_a0) { goperl_long_lex_read_to(goperl_a0); }
static inline __attribute__((always_inline, unused)) I32
goperl_long_lex_read_unichar(U32 goperl_a0) { return lex_read_unichar(goperl_a0); }
extern inline __attribute__((gnu_inline, always_inline, unused)) I32
Perl_lex_read_unichar(pTHX_ U32 goperl_a0) { return goperl_long_lex_read_unichar(goperl_a0); }
static inline __attribute__((always_inline, unused)) void
goperl_long_lex_start(SV * goperl_a0, PerlIO * goperl_a1, U32 goperl_a2) { lex_start(goperl_a0, goperl_a1, goperl_a2); }
extern inline __attribute__((gnu_inline, always_inline, unused)) void
Perl_lex_start(pTHX_ SV * goperl_a0, PerlIO * goperl_a1, U32 goperl_a2) { goperl_long_lex_start(goperl_a0, goperl_a1, goperl_a2); }
static inline __attribute__((always_inline, unused)) void
goperl_long_lex_stuff_pv(const char * goperl_a0, U32 goperl_a1) { lex_stuff_pv(goperl_a0, goperl_a1); }
extern inline __attribute__((gnu_inline, always_inline, unused)) void
Perl_lex_stuff_pv(pTHX_ const char * goperl_a0, U32 goperl_a1) { goperl_long_lex_stuff_pv(goperl_a0, goperl_a1); }
static inline __attribute__((always_inline, unused)) void
goperl_long_lex_stuff_pvn(const char * goperl_a0, STRLEN goperl_a1, U32 goperl_a2) { lex_stuff_pvn(goperl_a0, goperl_a1, goperl_a2); }
extern inline __attribute__((gnu_inline, always_inline, unused)) void
Perl_lex_stuff_pvn(pTHX_ const char * goperl_a0, STRLEN goperl_a1, U32 goperl_a2) { goperl_long_lex_stuff_pvn(goperl_a0, goperl_a1, goperl_a2); }
static inline __attribute__((always_inline, unused)) void
goperl_long_lex_stuff_sv(SV * goperl_a0, U32 goperl_a1) { lex_stuff_sv(goperl_a0, goperl_a1); }
extern inline __attribute__((gnu_inline, always_inline, unused)) void
Perl_lex_stuff_sv(pTHX_ SV * goperl_a0, U32 goperl_a1) { goperl_long_lex_stuff_sv(goperl_a0, goperl_a1); }
static inline __attribute__((always_inline, unused)) I32
goperl_long_looks_like_number(SV * const goperl_a0) { return looks_like_number(goperl_a0); }
extern inline __attribute__((gnu_inline, always_inline, unused)) I32
Perl_looks_like_number(pTHX_ SV * const goperl_a0) { return goperl_long_looks_like_number(goperl_a0); }
#if defined(magic_dump) && !defined(Perl_magic_dump)
#define Perl_magic_dump(...) GOPERL_LONG_CALL(magic_dump, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(mess) && !defined(Perl_mess)
#define Perl_mess(...) GOPERL_LONG_CALL(mess, GOPERL_THX_DROP(__VA_ARGS__))
#endif
static inline __attribute__((always_inline, unused)) SV *
goperl_long_mess_sv(SV * goperl_a0, bool goperl_a1) { return mess_sv(goperl_a0, goperl_a1); }
extern inline __attribute__((gnu_inline, always_inline, unused)) SV *
Perl_mess_sv(pTHX_ SV * goperl_a0, bool goperl_a1) { return goperl_long_mess_sv(goperl_a0, goperl_a1); }
static inline __attribute__((always_inline, unused)) int
goperl_long_mg_clear(SV * goperl_a0) { return mg_clear(goperl_a0); }
extern inline __attribute__((gnu_inline, always_inline, unused)) int
Perl_mg_clear(pTHX_ SV * goperl_a0) { return goperl_long_mg_clear(goperl_a0); }
#if defined(mg_copy) && !defined(Perl_mg_copy)
#define Perl_mg_copy(...) GOPERL_LONG_CALL(mg_copy, GOPERL_THX_DROP(__VA_ARGS__))
#endif
static inline __attribute__((always_inline, unused)) MAGIC *
goperl_long_mg_find(const SV * goperl_a0, int goperl_a1) { return mg_find(goperl_a0, goperl_a1); }
extern inline __attribute__((gnu_inline, always_inline, unused)) MAGIC *
Perl_mg_find(const SV * goperl_a0, int goperl_a1) { return goperl_long_mg_find(goperl_a0, goperl_a1); }
static inline __attribute__((always_inline, unused)) MAGIC *
goperl_long_mg_findext(const SV * goperl_a0, int goperl_a1, const MGVTBL * goperl_a2) { return mg_findext(goperl_a0, goperl_a1, goperl_a2); }
extern inline __attribute__((gnu_inline, always_inline, unused)) MAGIC *
Perl_mg_findext(const SV * goperl_a0, int goperl_a1, const MGVTBL * goperl_a2) { return goperl_long_mg_findext(goperl_a0, goperl_a1, goperl_a2); }
#if defined(mg_free) && !defined(Perl_mg_free)
#define Perl_mg_free(...) GOPERL_LONG_CALL(mg_free, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(mg_freeext) && !defined(Perl_mg_freeext)
#define Perl_mg_freeext(...) GOPERL_LONG_CALL(mg_freeext, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(mg_free_type) && !defined(Perl_mg_free_type)
#define Perl_mg_free_type(...) GOPERL_LONG_CALL(mg_free_type, GOPERL_THX_DROP(__VA_ARGS__))
#endif
static inline __attribute__((always_inline, unused)) int
goperl_long_mg_get(SV * goperl_a0) { return mg_get(goperl_a0); }
extern inline __attribute__((gnu_inline, always_inline, unused)) int
Perl_mg_get(pTHX_ SV * goperl_a0) { return goperl_long_mg_get(goperl_a0); }
static inline __attribute__((always_inline, unused)) void
goperl_long_mg_magical(SV * goperl_a0) { mg_magical(goperl_a0); }
extern inline __attribute__((gnu_inline, always_inline, unused)) void
Perl_mg_magical(SV * goperl_a0) { goperl_long_mg_magical(goperl_a0); }
static inline __attribute__((always_inline, unused)) int
goperl_long_mg_set(SV * goperl_a0) { return mg_set(goperl_a0); }
extern inline __attribute__((gnu_inline, always_inline, unused)) int
Perl_mg_set(pTHX_ SV * goperl_a0) { return goperl_long_mg_set(goperl_a0); }
#if defined(mortal_destructor_sv) && !defined(Perl_mortal_destructor_sv)
#define Perl_mortal_destructor_sv(...) GOPERL_LONG_CALL(mortal_destructor_sv, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(mro_get_from_name) && !defined(Perl_mro_get_from_name)
#define Perl_mro_get_from_name(...) GOPERL_LONG_CALL(mro_get_from_name, GOPERL_THX_DROP(__VA_ARGS__))
#endif
static inline __attribute__((always_inline, unused)) AV *
goperl_long_mro_get_linear_isa(HV * goperl_a0) { return mro_get_linear_isa(goperl_a0); }
extern inline __attribute__((gnu_inline, always_inline, unused)) AV *
Perl_mro_get_linear_isa(pTHX_ HV * goperl_a0) { return goperl_long_mro_get_linear_isa(goperl_a0); }
static inline __attribute__((always_inline, unused)) void
goperl_long_mro_method_changed_in(HV * goperl_a0) { mro_method_changed_in(goperl_a0); }
extern inline __attribute__((gnu_inline, always_inline, unused)) void
Perl_mro_method_changed_in(pTHX_ HV * goperl_a0) { goperl_long_mro_method_changed_in(goperl_a0); }
#if defined(mro_register) && !defined(Perl_mro_register)
#define Perl_mro_register(...) GOPERL_LONG_CALL(mro_register, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(mro_set_mro) && !defined(Perl_mro_set_mro)
#define Perl_mro_set_mro(...) GOPERL_LONG_CALL(mro_set_mro, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(mro_set_private_data) && !defined(Perl_mro_set_private_data)
#define Perl_mro_set_private_data(...) GOPERL_LONG_CALL(mro_set_private_data, GOPERL_THX_DROP(__VA_ARGS__))
#endif
static inline __attribute__((always_inline, unused)) NV
goperl_long_my_atof(const char * goperl_a0) { return my_atof(goperl_a0); }
extern inline __attribute__((gnu_inline, always_inline, unused)) NV
Perl_my_atof(pTHX_ const char * goperl_a0) { return goperl_long_my_atof(goperl_a0); }
#if defined(my_dirfd) && !defined(Perl_my_dirfd)
#define Perl_my_dirfd(...) GOPERL_LONG_CALL(my_dirfd, __VA_ARGS__)
#endif
#if defined(my_exit) && !defined(Perl_my_exit)
#define Perl_my_exit(...) GOPERL_LONG_CALL(my_exit, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(my_failure_exit) && !defined(Perl_my_failure_exit)
#define Perl_my_failure_exit(...) GOPERL_LONG_CALL(my_failure_exit, GOPERL_THX_DROP(__VA_ARGS__))
#endif
static inline __attribute__((always_inline, unused)) void
goperl_long_my_setenv(const char * goperl_a0, const char * goperl_a1) { my_setenv(goperl_a0, goperl_a1); }
extern inline __attribute__((gnu_inline, always_inline, unused)) void
Perl_my_setenv(pTHX_ const char * goperl_a0, const char * goperl_a1) { goperl_long_my_setenv(goperl_a0, goperl_a1); }
#if defined(my_snprintf) && !defined(Perl_my_snprintf)
#define Perl_my_snprintf(...) GOPERL_LONG_CALL(my_snprintf, __VA_ARGS__)
#endif
#if defined(my_strftime) && !defined(Perl_my_strftime)
#define Perl_my_strftime(...) GOPERL_LONG_CALL(my_strftime, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(my_strtod) && !defined(Perl_my_strtod)
#define Perl_my_strtod(...) GOPERL_LONG_CALL(my_strtod, __VA_ARGS__)
#endif
static inline __attribute__((always_inline, unused)) int
goperl_long_my_vsnprintf(char * goperl_a0, const Size_t goperl_a1, const char * goperl_a2, va_list goperl_a3) { return my_vsnprintf(goperl_a0, goperl_a1, goperl_a2, goperl_a3); }
extern inline __attribute__((gnu_inline, always_inline, unused)) int
Perl_my_vsnprintf(char * goperl_a0, const Size_t goperl_a1, const char * goperl_a2, va_list goperl_a3) { return goperl_long_my_vsnprintf(goperl_a0, goperl_a1, goperl_a2, goperl_a3); }
static inline __attribute__((always_inline, unused)) OP *
goperl_long_newANONATTRSUB(I32 goperl_a0, OP * goperl_a1, OP * goperl_a2, OP * goperl_a3) { return newANONATTRSUB(goperl_a0, goperl_a1, goperl_a2, goperl_a3); }
extern inline __attribute__((gnu_inline, always_inline, unused)) OP *
Perl_newANONATTRSUB(pTHX_ I32 goperl_a0, OP * goperl_a1, OP * goperl_a2, OP * goperl_a3) { return goperl_long_newANONATTRSUB(goperl_a0, goperl_a1, goperl_a2, goperl_a3); }
static inline __attribute__((always_inline, unused)) OP *
goperl_long_newANONHASH(OP * goperl_a0) { return newANONHASH(goperl_a0); }
extern inline __attribute__((gnu_inline, always_inline, unused)) OP *
Perl_newANONHASH(pTHX_ OP * goperl_a0) { return goperl_long_newANONHASH(goperl_a0); }
static inline __attribute__((always_inline, unused)) OP *
goperl_long_newANONLIST(OP * goperl_a0) { return newANONLIST(goperl_a0); }
extern inline __attribute__((gnu_inline, always_inline, unused)) OP *
Perl_newANONLIST(pTHX_ OP * goperl_a0) { return goperl_long_newANONLIST(goperl_a0); }
static inline __attribute__((always_inline, unused)) OP *
goperl_long_newANONSUB(I32 goperl_a0, OP * goperl_a1, OP * goperl_a2) { return newANONSUB(goperl_a0, goperl_a1, goperl_a2); }
extern inline __attribute__((gnu_inline, always_inline, unused)) OP *
Perl_newANONSUB(pTHX_ I32 goperl_a0, OP * goperl_a1, OP * goperl_a2) { return goperl_long_newANONSUB(goperl_a0, goperl_a1, goperl_a2); }
#if defined(newARGDEFELEMOP) && !defined(Perl_newARGDEFELEMOP)
#define Perl_newARGDEFELEMOP(...) GOPERL_LONG_CALL(newARGDEFELEMOP, GOPERL_THX_DROP(__VA_ARGS__))
#endif
static inline __attribute__((always_inline, unused)) OP *
goperl_long_newASSIGNOP(I32 goperl_a0, OP * goperl_a1, I32 goperl_a2, OP * goperl_a3) { return newASSIGNOP(goperl_a0, goperl_a1, goperl_a2, goperl_a3); }
extern inline __attribute__((gnu_inline, always_inline, unused)) OP *
Perl_newASSIGNOP(pTHX_ I32 goperl_a0, OP * goperl_a1, I32 goperl_a2, OP * goperl_a3) { return goperl_long_newASSIGNOP(goperl_a0, goperl_a1, goperl_a2, goperl_a3); }
static inline __attribute__((always_inline, unused)) CV *
goperl_long_newATTRSUB(I32 goperl_a0, OP * goperl_a1, OP * goperl_a2, OP * goperl_a3, OP * goperl_a4) { return newATTRSUB(goperl_a0, goperl_a1, goperl_a2, goperl_a3, goperl_a4); }
extern inline __attribute__((gnu_inline, always_inline, unused)) CV *
Perl_newATTRSUB(pTHX_ I32 goperl_a0, OP * goperl_a1, OP * goperl_a2, OP * goperl_a3, OP * goperl_a4) { return goperl_long_newATTRSUB(goperl_a0, goperl_a1, goperl_a2, goperl_a3, goperl_a4); }
static inline __attribute__((always_inline, unused)) AV *
goperl_long_newAV(void) { return newAV(); }
extern inline __attribute__((gnu_inline, always_inline, unused)) AV *
Perl_newAV(pTHX) { return goperl_long_newAV(); }
static inline __attribute__((always_inline, unused)) AV *
goperl_long_newAV_alloc_x(SSize_t goperl_a0) { return newAV_alloc_x(goperl_a0); }
extern inline __attribute__((gnu_inline, always_inline, unused)) AV *
Perl_newAV_alloc_x(pTHX_ SSize_t goperl_a0) { return goperl_long_newAV_alloc_x(goperl_a0); }
static inline __attribute__((always_inline, unused)) AV *
goperl_long_newAV_alloc_xz(SSize_t goperl_a0) { return newAV_alloc_xz(goperl_a0); }
extern inline __attribute__((gnu_inline, always_inline, unused)) AV *
Perl_newAV_alloc_xz(pTHX_ SSize_t goperl_a0) { return goperl_long_newAV_alloc_xz(goperl_a0); }
static inline __attribute__((always_inline, unused)) AV *
goperl_long_newAVav(AV * goperl_a0) { return newAVav(goperl_a0); }
extern inline __attribute__((gnu_inline, always_inline, unused)) AV *
Perl_newAVav(pTHX_ AV * goperl_a0) { return goperl_long_newAVav(goperl_a0); }
static inline __attribute__((always_inline, unused)) AV *
goperl_long_newAVhv(HV * goperl_a0) { return newAVhv(goperl_a0); }
extern inline __attribute__((gnu_inline, always_inline, unused)) AV *
Perl_newAVhv(pTHX_ HV * goperl_a0) { return goperl_long_newAVhv(goperl_a0); }
#if defined(newAV_mortal) && !defined(Perl_newAV_mortal)
#define Perl_newAV_mortal(...) GOPERL_LONG_CALL(newAV_mortal, GOPERL_THX_DROP(__VA_ARGS__))
#endif
static inline __attribute__((always_inline, unused)) OP *
goperl_long_newAVREF(OP * goperl_a0) { return newAVREF(goperl_a0); }
extern inline __attribute__((gnu_inline, always_inline, unused)) OP *
Perl_newAVREF(pTHX_ OP * goperl_a0) { return goperl_long_newAVREF(goperl_a0); }
static inline __attribute__((always_inline, unused)) OP *
goperl_long_newBINOP(I32 goperl_a0, I32 goperl_a1, OP * goperl_a2, OP * goperl_a3) { return newBINOP(goperl_a0, goperl_a1, goperl_a2, goperl_a3); }
extern inline __attribute__((gnu_inline, always_inline, unused)) OP *
Perl_newBINOP(pTHX_ I32 goperl_a0, I32 goperl_a1, OP * goperl_a2, OP * goperl_a3) { return goperl_long_newBINOP(goperl_a0, goperl_a1, goperl_a2, goperl_a3); }
static inline __attribute__((always_inline, unused)) OP *
goperl_long_newCONDOP(I32 goperl_a0, OP * goperl_a1, OP * goperl_a2, OP * goperl_a3) { return newCONDOP(goperl_a0, goperl_a1, goperl_a2, goperl_a3); }
extern inline __attribute__((gnu_inline, always_inline, unused)) OP *
Perl_newCONDOP(pTHX_ I32 goperl_a0, OP * goperl_a1, OP * goperl_a2, OP * goperl_a3) { return goperl_long_newCONDOP(goperl_a0, goperl_a1, goperl_a2, goperl_a3); }
static inline __attribute__((always_inline, unused)) CV *
goperl_long_newCONSTSUB(HV * goperl_a0, const char * goperl_a1, SV * goperl_a2) { return newCONSTSUB(goperl_a0, goperl_a1, goperl_a2); }
extern inline __attribute__((gnu_inline, always_inline, unused)) CV *
Perl_newCONSTSUB(pTHX_ HV * goperl_a0, const char * goperl_a1, SV * goperl_a2) { return goperl_long_newCONSTSUB(goperl_a0, goperl_a1, goperl_a2); }
#if defined(newCONSTSUB_flags) && !defined(Perl_newCONSTSUB_flags)
#define Perl_newCONSTSUB_flags(...) GOPERL_LONG_CALL(newCONSTSUB_flags, GOPERL_THX_DROP(__VA_ARGS__))
#endif
static inline __attribute__((always_inline, unused)) OP *
goperl_long_newCVREF(I32 goperl_a0, OP * goperl_a1) { return newCVREF(goperl_a0, goperl_a1); }
extern inline __attribute__((gnu_inline, always_inline, unused)) OP *
Perl_newCVREF(pTHX_ I32 goperl_a0, OP * goperl_a1) { return goperl_long_newCVREF(goperl_a0, goperl_a1); }
#if defined(newDEFEROP) && !defined(Perl_newDEFEROP)
#define Perl_newDEFEROP(...) GOPERL_LONG_CALL(newDEFEROP, GOPERL_THX_DROP(__VA_ARGS__))
#endif
static inline __attribute__((always_inline, unused)) OP *
goperl_long_newDEFSVOP(void) { return newDEFSVOP(); }
extern inline __attribute__((gnu_inline, always_inline, unused)) OP *
Perl_newDEFSVOP(pTHX) { return goperl_long_newDEFSVOP(); }
static inline __attribute__((always_inline, unused)) OP *
goperl_long_newFOROP(I32 goperl_a0, OP * goperl_a1, OP * goperl_a2, OP * goperl_a3, OP * goperl_a4) { return newFOROP(goperl_a0, goperl_a1, goperl_a2, goperl_a3, goperl_a4); }
extern inline __attribute__((gnu_inline, always_inline, unused)) OP *
Perl_newFOROP(pTHX_ I32 goperl_a0, OP * goperl_a1, OP * goperl_a2, OP * goperl_a3, OP * goperl_a4) { return goperl_long_newFOROP(goperl_a0, goperl_a1, goperl_a2, goperl_a3, goperl_a4); }
#if defined(newGIVENOP) && !defined(Perl_newGIVENOP)
#define Perl_newGIVENOP(...) GOPERL_LONG_CALL(newGIVENOP, GOPERL_THX_DROP(__VA_ARGS__))
#endif
static inline __attribute__((always_inline, unused)) GV *
goperl_long_newGVgen(const char * goperl_a0) { return newGVgen(goperl_a0); }
extern inline __attribute__((gnu_inline, always_inline, unused)) GV *
Perl_newGVgen(pTHX_ const char * goperl_a0) { return goperl_long_newGVgen(goperl_a0); }
static inline __attribute__((always_inline, unused)) GV *
goperl_long_newGVgen_flags(const char * goperl_a0, U32 goperl_a1) { return newGVgen_flags(goperl_a0, goperl_a1); }
extern inline __attribute__((gnu_inline, always_inline, unused)) GV *
Perl_newGVgen_flags(pTHX_ const char * goperl_a0, U32 goperl_a1) { return goperl_long_newGVgen_flags(goperl_a0, goperl_a1); }
static inline __attribute__((always_inline, unused)) OP *
goperl_long_newGVOP(I32 goperl_a0, I32 goperl_a1, GV * goperl_a2) { return newGVOP(goperl_a0, goperl_a1, goperl_a2); }
extern inline __attribute__((gnu_inline, always_inline, unused)) OP *
Perl_newGVOP(pTHX_ I32 goperl_a0, I32 goperl_a1, GV * goperl_a2) { return goperl_long_newGVOP(goperl_a0, goperl_a1, goperl_a2); }
static inline __attribute__((always_inline, unused)) OP *
goperl_long_newGVREF(I32 goperl_a0, OP * goperl_a1) { return newGVREF(goperl_a0, goperl_a1); }
extern inline __attribute__((gnu_inline, always_inline, unused)) OP *
Perl_newGVREF(pTHX_ I32 goperl_a0, OP * goperl_a1) { return goperl_long_newGVREF(goperl_a0, goperl_a1); }
static inline __attribute__((always_inline, unused)) HV *
goperl_long_newHV(void) { return newHV(); }
extern inline __attribute__((gnu_inline, always_inline, unused)) HV *
Perl_newHV(pTHX) { return goperl_long_newHV(); }
static inline __attribute__((always_inline, unused)) HV *
goperl_long_newHVhv(HV * goperl_a0) { return newHVhv(goperl_a0); }
extern inline __attribute__((gnu_inline, always_inline, unused)) HV *
Perl_newHVhv(pTHX_ HV * goperl_a0) { return goperl_long_newHVhv(goperl_a0); }
static inline __attribute__((always_inline, unused)) OP *
goperl_long_newHVREF(OP * goperl_a0) { return newHVREF(goperl_a0); }
extern inline __attribute__((gnu_inline, always_inline, unused)) OP *
Perl_newHVREF(pTHX_ OP * goperl_a0) { return goperl_long_newHVREF(goperl_a0); }
#if defined(newIO) && !defined(Perl_newIO)
#define Perl_newIO(...) GOPERL_LONG_CALL(newIO, GOPERL_THX_DROP(__VA_ARGS__))
#endif
static inline __attribute__((always_inline, unused)) OP *
goperl_long_newLISTOP(I32 goperl_a0, I32 goperl_a1, OP * goperl_a2, OP * goperl_a3) { return newLISTOP(goperl_a0, goperl_a1, goperl_a2, goperl_a3); }
extern inline __attribute__((gnu_inline, always_inline, unused)) OP *
Perl_newLISTOP(pTHX_ I32 goperl_a0, I32 goperl_a1, OP * goperl_a2, OP * goperl_a3) { return goperl_long_newLISTOP(goperl_a0, goperl_a1, goperl_a2, goperl_a3); }
#if defined(newLISTOPn) && !defined(Perl_newLISTOPn)
#define Perl_newLISTOPn(...) GOPERL_LONG_CALL(newLISTOPn, GOPERL_THX_DROP(__VA_ARGS__))
#endif
static inline __attribute__((always_inline, unused)) OP *
goperl_long_newLOGOP(I32 goperl_a0, I32 goperl_a1, OP * goperl_a2, OP * goperl_a3) { return newLOGOP(goperl_a0, goperl_a1, goperl_a2, goperl_a3); }
extern inline __attribute__((gnu_inline, always_inline, unused)) OP *
Perl_newLOGOP(pTHX_ I32 goperl_a0, I32 goperl_a1, OP * goperl_a2, OP * goperl_a3) { return goperl_long_newLOGOP(goperl_a0, goperl_a1, goperl_a2, goperl_a3); }
static inline __attribute__((always_inline, unused)) OP *
goperl_long_newLOOPEX(I32 goperl_a0, OP * goperl_a1) { return newLOOPEX(goperl_a0, goperl_a1); }
extern inline __attribute__((gnu_inline, always_inline, unused)) OP *
Perl_newLOOPEX(pTHX_ I32 goperl_a0, OP * goperl_a1) { return goperl_long_newLOOPEX(goperl_a0, goperl_a1); }
static inline __attribute__((always_inline, unused)) OP *
goperl_long_newLOOPOP(I32 goperl_a0, I32 goperl_a1, OP * goperl_a2, OP * goperl_a3) { return newLOOPOP(goperl_a0, goperl_a1, goperl_a2, goperl_a3); }
extern inline __attribute__((gnu_inline, always_inline, unused)) OP *
Perl_newLOOPOP(pTHX_ I32 goperl_a0, I32 goperl_a1, OP * goperl_a2, OP * goperl_a3) { return goperl_long_newLOOPOP(goperl_a0, goperl_a1, goperl_a2, goperl_a3); }
#if defined(newMETHOP) && !defined(Perl_newMETHOP)
#define Perl_newMETHOP(...) GOPERL_LONG_CALL(newMETHOP, GOPERL_THX_DROP(__VA_ARGS__))
#endif
static inline __attribute__((always_inline, unused)) OP *
goperl_long_newMETHOP_named(I32 goperl_a0, I32 goperl_a1, SV * goperl_a2) { return newMETHOP_named(goperl_a0, goperl_a1, goperl_a2); }
extern inline __attribute__((gnu_inline, always_inline, unused)) OP *
Perl_newMETHOP_named(pTHX_ I32 goperl_a0, I32 goperl_a1, SV * goperl_a2) { return goperl_long_newMETHOP_named(goperl_a0, goperl_a1, goperl_a2); }
static inline __attribute__((always_inline, unused)) OP *
goperl_long_newNULLLIST(void) { return newNULLLIST(); }
extern inline __attribute__((gnu_inline, always_inline, unused)) OP *
Perl_newNULLLIST(pTHX) { return goperl_long_newNULLLIST(); }
static inline __attribute__((always_inline, unused)) OP *
goperl_long_newOP(I32 goperl_a0, I32 goperl_a1) { return newOP(goperl_a0, goperl_a1); }
extern inline __attribute__((gnu_inline, always_inline, unused)) OP *
Perl_newOP(pTHX_ I32 goperl_a0, I32 goperl_a1) { return goperl_long_newOP(goperl_a0, goperl_a1); }
#if defined(newPADNAMELIST) && !defined(Perl_newPADNAMELIST)
#define Perl_newPADNAMELIST(...) GOPERL_LONG_CALL(newPADNAMELIST, __VA_ARGS__)
#endif
#if defined(newPADNAMEouter) && !defined(Perl_newPADNAMEouter)
#define Perl_newPADNAMEouter(...) GOPERL_LONG_CALL(newPADNAMEouter, __VA_ARGS__)
#endif
static inline __attribute__((always_inline, unused)) PADNAME *
goperl_long_newPADNAMEpvn(const char * goperl_a0, STRLEN goperl_a1) { return newPADNAMEpvn(goperl_a0, goperl_a1); }
extern inline __attribute__((gnu_inline, always_inline, unused)) PADNAME *
Perl_newPADNAMEpvn(const char * goperl_a0, STRLEN goperl_a1) { return goperl_long_newPADNAMEpvn(goperl_a0, goperl_a1); }
#if defined(newPADxVOP) && !defined(Perl_newPADxVOP)
#define Perl_newPADxVOP(...) GOPERL_LONG_CALL(newPADxVOP, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(newPMOP) && !defined(Perl_newPMOP)
#define Perl_newPMOP(...) GOPERL_LONG_CALL(newPMOP, GOPERL_THX_DROP(__VA_ARGS__))
#endif
static inline __attribute__((always_inline, unused)) OP *
goperl_long_newPVOP(I32 goperl_a0, I32 goperl_a1, char * goperl_a2) { return newPVOP(goperl_a0, goperl_a1, goperl_a2); }
extern inline __attribute__((gnu_inline, always_inline, unused)) OP *
Perl_newPVOP(pTHX_ I32 goperl_a0, I32 goperl_a1, char * goperl_a2) { return goperl_long_newPVOP(goperl_a0, goperl_a1, goperl_a2); }
static inline __attribute__((always_inline, unused)) OP *
goperl_long_newRANGE(I32 goperl_a0, OP * goperl_a1, OP * goperl_a2) { return newRANGE(goperl_a0, goperl_a1, goperl_a2); }
extern inline __attribute__((gnu_inline, always_inline, unused)) OP *
Perl_newRANGE(pTHX_ I32 goperl_a0, OP * goperl_a1, OP * goperl_a2) { return goperl_long_newRANGE(goperl_a0, goperl_a1, goperl_a2); }
static inline __attribute__((always_inline, unused)) SV *
goperl_long_newRV(SV * const goperl_a0) { return newRV(goperl_a0); }
extern inline __attribute__((gnu_inline, always_inline, unused)) SV *
Perl_newRV(pTHX_ SV * const goperl_a0) { return goperl_long_newRV(goperl_a0); }
static inline __attribute__((always_inline, unused)) SV *
goperl_long_newRV_noinc(SV * const goperl_a0) { return newRV_noinc(goperl_a0); }
extern inline __attribute__((gnu_inline, always_inline, unused)) SV *
Perl_newRV_noinc(pTHX_ SV * const goperl_a0) { return goperl_long_newRV_noinc(goperl_a0); }
static inline __attribute__((always_inline, unused)) OP *
goperl_long_newSLICEOP(I32 goperl_a0, OP * goperl_a1, OP * goperl_a2) { return newSLICEOP(goperl_a0, goperl_a1, goperl_a2); }
extern inline __attribute__((gnu_inline, always_inline, unused)) OP *
Perl_newSLICEOP(pTHX_ I32 goperl_a0, OP * goperl_a1, OP * goperl_a2) { return goperl_long_newSLICEOP(goperl_a0, goperl_a1, goperl_a2); }
static inline __attribute__((always_inline, unused)) OP *
goperl_long_newSTATEOP(I32 goperl_a0, char * goperl_a1, OP * goperl_a2) { return newSTATEOP(goperl_a0, goperl_a1, goperl_a2); }
extern inline __attribute__((gnu_inline, always_inline, unused)) OP *
Perl_newSTATEOP(pTHX_ I32 goperl_a0, char * goperl_a1, OP * goperl_a2) { return goperl_long_newSTATEOP(goperl_a0, goperl_a1, goperl_a2); }
static inline __attribute__((always_inline, unused)) CV *
goperl_long_newSUB(I32 goperl_a0, OP * goperl_a1, OP * goperl_a2, OP * goperl_a3) { return newSUB(goperl_a0, goperl_a1, goperl_a2, goperl_a3); }
extern inline __attribute__((gnu_inline, always_inline, unused)) CV *
Perl_newSUB(pTHX_ I32 goperl_a0, OP * goperl_a1, OP * goperl_a2, OP * goperl_a3) { return goperl_long_newSUB(goperl_a0, goperl_a1, goperl_a2, goperl_a3); }
static inline __attribute__((always_inline, unused)) SV *
goperl_long_newSV(const STRLEN goperl_a0) { return newSV(goperl_a0); }
extern inline __attribute__((gnu_inline, always_inline, unused)) SV *
Perl_newSV(pTHX_ const STRLEN goperl_a0) { return goperl_long_newSV(goperl_a0); }
static inline __attribute__((always_inline, unused)) SV *
goperl_long_newSVbool(const bool goperl_a0) { return newSVbool(goperl_a0); }
extern inline __attribute__((gnu_inline, always_inline, unused)) SV *
Perl_newSVbool(pTHX_ const bool goperl_a0) { return goperl_long_newSVbool(goperl_a0); }
static inline __attribute__((always_inline, unused)) SV *
goperl_long_newSV_false(void) { return newSV_false(); }
extern inline __attribute__((gnu_inline, always_inline, unused)) SV *
Perl_newSV_false(pTHX) { return goperl_long_newSV_false(); }
static inline __attribute__((always_inline, unused)) SV *
goperl_long_newSVhek(const HEK * const goperl_a0) { return newSVhek(goperl_a0); }
extern inline __attribute__((gnu_inline, always_inline, unused)) SV *
Perl_newSVhek(pTHX_ const HEK * const goperl_a0) { return goperl_long_newSVhek(goperl_a0); }
static inline __attribute__((always_inline, unused)) SV *
goperl_long_newSVhek_mortal(const HEK * const goperl_a0) { return newSVhek_mortal(goperl_a0); }
extern inline __attribute__((gnu_inline, always_inline, unused)) SV *
Perl_newSVhek_mortal(pTHX_ const HEK * const goperl_a0) { return goperl_long_newSVhek_mortal(goperl_a0); }
static inline __attribute__((always_inline, unused)) SV *
goperl_long_newSViv(const IV goperl_a0) { return newSViv(goperl_a0); }
extern inline __attribute__((gnu_inline, always_inline, unused)) SV *
Perl_newSViv(pTHX_ const IV goperl_a0) { return goperl_long_newSViv(goperl_a0); }
static inline __attribute__((always_inline, unused)) SV *
goperl_long_newSVnv(const NV goperl_a0) { return newSVnv(goperl_a0); }
extern inline __attribute__((gnu_inline, always_inline, unused)) SV *
Perl_newSVnv(pTHX_ const NV goperl_a0) { return goperl_long_newSVnv(goperl_a0); }
static inline __attribute__((always_inline, unused)) OP *
goperl_long_newSVOP(I32 goperl_a0, I32 goperl_a1, SV * goperl_a2) { return newSVOP(goperl_a0, goperl_a1, goperl_a2); }
extern inline __attribute__((gnu_inline, always_inline, unused)) OP *
Perl_newSVOP(pTHX_ I32 goperl_a0, I32 goperl_a1, SV * goperl_a2) { return goperl_long_newSVOP(goperl_a0, goperl_a1, goperl_a2); }
static inline __attribute__((always_inline, unused)) SV *
goperl_long_newSVpv(const char * const goperl_a0, const STRLEN goperl_a1) { return newSVpv(goperl_a0, goperl_a1); }
extern inline __attribute__((gnu_inline, always_inline, unused)) SV *
Perl_newSVpv(pTHX_ const char * const goperl_a0, const STRLEN goperl_a1) { return goperl_long_newSVpv(goperl_a0, goperl_a1); }
static inline __attribute__((always_inline, unused)) SV *
goperl_long_newSVpvn(const char * const goperl_a0, const STRLEN goperl_a1) { return newSVpvn(goperl_a0, goperl_a1); }
extern inline __attribute__((gnu_inline, always_inline, unused)) SV *
Perl_newSVpvn(pTHX_ const char * const goperl_a0, const STRLEN goperl_a1) { return goperl_long_newSVpvn(goperl_a0, goperl_a1); }
static inline __attribute__((always_inline, unused)) SV *
goperl_long_newSVpvn_flags(const char * const goperl_a0, const STRLEN goperl_a1, const U32 goperl_a2) { return newSVpvn_flags(goperl_a0, goperl_a1, goperl_a2); }
extern inline __attribute__((gnu_inline, always_inline, unused)) SV *
Perl_newSVpvn_flags(pTHX_ const char * const goperl_a0, const STRLEN goperl_a1, const U32 goperl_a2) { return goperl_long_newSVpvn_flags(goperl_a0, goperl_a1, goperl_a2); }
static inline __attribute__((always_inline, unused)) SV *
goperl_long_newSVpvn_share(const char * goperl_a0, I32 goperl_a1, U32 goperl_a2) { return newSVpvn_share(goperl_a0, goperl_a1, goperl_a2); }
extern inline __attribute__((gnu_inline, always_inline, unused)) SV *
Perl_newSVpvn_share(pTHX_ const char * goperl_a0, I32 goperl_a1, U32 goperl_a2) { return goperl_long_newSVpvn_share(goperl_a0, goperl_a1, goperl_a2); }
#if defined(newSVpv_share) && !defined(Perl_newSVpv_share)
#define Perl_newSVpv_share(...) GOPERL_LONG_CALL(newSVpv_share, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(newSVpvz) && !defined(Perl_newSVpvz)
#define Perl_newSVpvz(...) GOPERL_LONG_CALL(newSVpvz, GOPERL_THX_DROP(__VA_ARGS__))
#endif
static inline __attribute__((always_inline, unused)) OP *
goperl_long_newSVREF(OP * goperl_a0) { return newSVREF(goperl_a0); }
extern inline __attribute__((gnu_inline, always_inline, unused)) OP *
Perl_newSVREF(pTHX_ OP * goperl_a0) { return goperl_long_newSVREF(goperl_a0); }
static inline __attribute__((always_inline, unused)) SV *
goperl_long_newSVrv(SV * const goperl_a0, const char * const goperl_a1) { return newSVrv(goperl_a0, goperl_a1); }
extern inline __attribute__((gnu_inline, always_inline, unused)) SV *
Perl_newSVrv(pTHX_ SV * const goperl_a0, const char * const goperl_a1) { return goperl_long_newSVrv(goperl_a0, goperl_a1); }
static inline __attribute__((always_inline, unused)) SV *
goperl_long_newSVsv(SV * const goperl_a0) { return newSVsv(goperl_a0); }
extern inline __attribute__((gnu_inline, always_inline, unused)) SV *
Perl_newSVsv(pTHX_ SV * const goperl_a0) { return goperl_long_newSVsv(goperl_a0); }
static inline __attribute__((always_inline, unused)) SV *
goperl_long_newSVsv_flags(SV * const goperl_a0, I32 goperl_a1) { return newSVsv_flags(goperl_a0, goperl_a1); }
extern inline __attribute__((gnu_inline, always_inline, unused)) SV *
Perl_newSVsv_flags(pTHX_ SV * const goperl_a0, I32 goperl_a1) { return goperl_long_newSVsv_flags(goperl_a0, goperl_a1); }
#if defined(newSVsv_flags_NN) && !defined(Perl_newSVsv_flags_NN)
#define Perl_newSVsv_flags_NN(...) GOPERL_LONG_CALL(newSVsv_flags_NN, GOPERL_THX_DROP(__VA_ARGS__))
#endif
static inline __attribute__((always_inline, unused)) SV *
goperl_long_newSVsv_nomg(SV * const goperl_a0) { return newSVsv_nomg(goperl_a0); }
extern inline __attribute__((gnu_inline, always_inline, unused)) SV *
Perl_newSVsv_nomg(pTHX_ SV * const goperl_a0) { return goperl_long_newSVsv_nomg(goperl_a0); }
static inline __attribute__((always_inline, unused)) SV *
goperl_long_newSV_true(void) { return newSV_true(); }
extern inline __attribute__((gnu_inline, always_inline, unused)) SV *
Perl_newSV_true(pTHX) { return goperl_long_newSV_true(); }
static inline __attribute__((always_inline, unused)) SV *
goperl_long_newSV_type(const svtype goperl_a0) { return newSV_type(goperl_a0); }
extern inline __attribute__((gnu_inline, always_inline, unused)) SV *
Perl_newSV_type(pTHX_ const svtype goperl_a0) { return goperl_long_newSV_type(goperl_a0); }
static inline __attribute__((always_inline, unused)) SV *
goperl_long_newSV_type_mortal(const svtype goperl_a0) { return newSV_type_mortal(goperl_a0); }
extern inline __attribute__((gnu_inline, always_inline, unused)) SV *
Perl_newSV_type_mortal(pTHX_ const svtype goperl_a0) { return goperl_long_newSV_type_mortal(goperl_a0); }
static inline __attribute__((always_inline, unused)) SV *
goperl_long_newSVuv(const UV goperl_a0) { return newSVuv(goperl_a0); }
extern inline __attribute__((gnu_inline, always_inline, unused)) SV *
Perl_newSVuv(pTHX_ const UV goperl_a0) { return goperl_long_newSVuv(goperl_a0); }
#if defined(newTRYCATCHOP) && !defined(Perl_newTRYCATCHOP)
#define Perl_newTRYCATCHOP(...) GOPERL_LONG_CALL(newTRYCATCHOP, GOPERL_THX_DROP(__VA_ARGS__))
#endif
static inline __attribute__((always_inline, unused)) OP *
goperl_long_newUNOP(I32 goperl_a0, I32 goperl_a1, OP * goperl_a2) { return newUNOP(goperl_a0, goperl_a1, goperl_a2); }
extern inline __attribute__((gnu_inline, always_inline, unused)) OP *
Perl_newUNOP(pTHX_ I32 goperl_a0, I32 goperl_a1, OP * goperl_a2) { return goperl_long_newUNOP(goperl_a0, goperl_a1, goperl_a2); }
static inline __attribute__((always_inline, unused)) OP *
goperl_long_newUNOP_AUX(I32 goperl_a0, I32 goperl_a1, OP * goperl_a2, UNOP_AUX_item * goperl_a3) { return newUNOP_AUX(goperl_a0, goperl_a1, goperl_a2, goperl_a3); }
extern inline __attribute__((gnu_inline, always_inline, unused)) OP *
Perl_newUNOP_AUX(pTHX_ I32 goperl_a0, I32 goperl_a1, OP * goperl_a2, UNOP_AUX_item * goperl_a3) { return goperl_long_newUNOP_AUX(goperl_a0, goperl_a1, goperl_a2, goperl_a3); }
#if defined(new_version) && !defined(Perl_new_version)
#define Perl_new_version(...) GOPERL_LONG_CALL(new_version, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(newWHENOP) && !defined(Perl_newWHENOP)
#define Perl_newWHENOP(...) GOPERL_LONG_CALL(newWHENOP, GOPERL_THX_DROP(__VA_ARGS__))
#endif
static inline __attribute__((always_inline, unused)) OP *
goperl_long_newWHILEOP(I32 goperl_a0, I32 goperl_a1, LOOP * goperl_a2, OP * goperl_a3, OP * goperl_a4, OP * goperl_a5, I32 goperl_a6) { return newWHILEOP(goperl_a0, goperl_a1, goperl_a2, goperl_a3, goperl_a4, goperl_a5, goperl_a6); }
extern inline __attribute__((gnu_inline, always_inline, unused)) OP *
Perl_newWHILEOP(pTHX_ I32 goperl_a0, I32 goperl_a1, LOOP * goperl_a2, OP * goperl_a3, OP * goperl_a4, OP * goperl_a5, I32 goperl_a6) { return goperl_long_newWHILEOP(goperl_a0, goperl_a1, goperl_a2, goperl_a3, goperl_a4, goperl_a5, goperl_a6); }
static inline __attribute__((always_inline, unused)) CV *
goperl_long_newXS(const char * goperl_a0, XSUBADDR_t goperl_a1, const char * goperl_a2) { return newXS(goperl_a0, goperl_a1, goperl_a2); }
extern inline __attribute__((gnu_inline, always_inline, unused)) CV *
Perl_newXS(pTHX_ const char * goperl_a0, XSUBADDR_t goperl_a1, const char * goperl_a2) { return goperl_long_newXS(goperl_a0, goperl_a1, goperl_a2); }
static inline __attribute__((always_inline, unused)) CV *
goperl_long_newXS_flags(const char * goperl_a0, XSUBADDR_t goperl_a1, const char * const goperl_a2, const char * const goperl_a3, U32 goperl_a4) { return newXS_flags(goperl_a0, goperl_a1, goperl_a2, goperl_a3, goperl_a4); }
extern inline __attribute__((gnu_inline, always_inline, unused)) CV *
Perl_newXS_flags(pTHX_ const char * goperl_a0, XSUBADDR_t goperl_a1, const char * const goperl_a2, const char * const goperl_a3, U32 goperl_a4) { return goperl_long_newXS_flags(goperl_a0, goperl_a1, goperl_a2, goperl_a3, goperl_a4); }
#if defined(ninstr) && !defined(Perl_ninstr)
#define Perl_ninstr(...) GOPERL_LONG_CALL(ninstr, __VA_ARGS__)
#endif
#if defined(nothreadhook) && !defined(Perl_nothreadhook)
#define Perl_nothreadhook(...) GOPERL_LONG_CALL(nothreadhook, GOPERL_THX_DROP(__VA_ARGS__))
#endif
static inline __attribute__((always_inline, unused)) OP *
goperl_long_op_append_elem(I32 goperl_a0, OP * goperl_a1, OP * goperl_a2) { return op_append_elem(goperl_a0, goperl_a1, goperl_a2); }
extern inline __attribute__((gnu_inline, always_inline, unused)) OP *
Perl_op_append_elem(pTHX_ I32 goperl_a0, OP * goperl_a1, OP * goperl_a2) { return goperl_long_op_append_elem(goperl_a0, goperl_a1, goperl_a2); }
static inline __attribute__((always_inline, unused)) OP *
goperl_long_op_append_list(I32 goperl_a0, OP * goperl_a1, OP * goperl_a2) { return op_append_list(goperl_a0, goperl_a1, goperl_a2); }
extern inline __attribute__((gnu_inline, always_inline, unused)) OP *
Perl_op_append_list(pTHX_ I32 goperl_a0, OP * goperl_a1, OP * goperl_a2) { return goperl_long_op_append_list(goperl_a0, goperl_a1, goperl_a2); }
static inline __attribute__((always_inline, unused)) OPclass
goperl_long_op_class(const OP * goperl_a0) { return op_class(goperl_a0); }
extern inline __attribute__((gnu_inline, always_inline, unused)) OPclass
Perl_op_class(pTHX_ const OP * goperl_a0) { return goperl_long_op_class(goperl_a0); }
#if defined(op_clear) && !defined(Perl_op_clear)
#define Perl_op_clear(...) GOPERL_LONG_CALL(op_clear, GOPERL_THX_DROP(__VA_ARGS__))
#endif
static inline __attribute__((always_inline, unused)) OP *
goperl_long_op_contextualize(OP * goperl_a0, I32 goperl_a1) { return op_contextualize(goperl_a0, goperl_a1); }
extern inline __attribute__((gnu_inline, always_inline, unused)) OP *
Perl_op_contextualize(pTHX_ OP * goperl_a0, I32 goperl_a1) { return goperl_long_op_contextualize(goperl_a0, goperl_a1); }
static inline __attribute__((always_inline, unused)) OP *
goperl_long_op_convert_list(I32 goperl_a0, I32 goperl_a1, OP * goperl_a2) { return op_convert_list(goperl_a0, goperl_a1, goperl_a2); }
extern inline __attribute__((gnu_inline, always_inline, unused)) OP *
Perl_op_convert_list(pTHX_ I32 goperl_a0, I32 goperl_a1, OP * goperl_a2) { return goperl_long_op_convert_list(goperl_a0, goperl_a1, goperl_a2); }
static inline __attribute__((always_inline, unused)) void
goperl_long_op_dump(const OP * goperl_a0) { op_dump(goperl_a0); }
extern inline __attribute__((gnu_inline, always_inline, unused)) void
Perl_op_dump(pTHX_ const OP * goperl_a0) { goperl_long_op_dump(goperl_a0); }
#if defined(opdump_printf) && !defined(Perl_opdump_printf)
#define Perl_opdump_printf(...) GOPERL_LONG_CALL(opdump_printf, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(op_force_list) && !defined(Perl_op_force_list)
#define Perl_op_force_list(...) GOPERL_LONG_CALL(op_force_list, GOPERL_THX_DROP(__VA_ARGS__))
#endif
static inline __attribute__((always_inline, unused)) void
goperl_long_op_free(OP * goperl_a0) { op_free(goperl_a0); }
extern inline __attribute__((gnu_inline, always_inline, unused)) void
Perl_op_free(pTHX_ OP * goperl_a0) { goperl_long_op_free(goperl_a0); }
static inline __attribute__((always_inline, unused)) OP *
goperl_long_op_linklist(OP * goperl_a0) { return op_linklist(goperl_a0); }
extern inline __attribute__((gnu_inline, always_inline, unused)) OP *
Perl_op_linklist(pTHX_ OP * goperl_a0) { return goperl_long_op_linklist(goperl_a0); }
static inline __attribute__((always_inline, unused)) OP *
goperl_long_op_lvalue(OP * goperl_a0, I32 goperl_a1) { return op_lvalue(goperl_a0, goperl_a1); }
extern inline __attribute__((gnu_inline, always_inline, unused)) OP *
Perl_op_lvalue(pTHX_ OP * goperl_a0, I32 goperl_a1) { return goperl_long_op_lvalue(goperl_a0, goperl_a1); }
static inline __attribute__((always_inline, unused)) void
goperl_long_op_null(OP * goperl_a0) { op_null(goperl_a0); }
extern inline __attribute__((gnu_inline, always_inline, unused)) void
Perl_op_null(pTHX_ OP * goperl_a0) { goperl_long_op_null(goperl_a0); }
#if defined(op_parent) && !defined(Perl_op_parent)
#define Perl_op_parent(...) GOPERL_LONG_CALL(op_parent, __VA_ARGS__)
#endif
static inline __attribute__((always_inline, unused)) OP *
goperl_long_op_prepend_elem(I32 goperl_a0, OP * goperl_a1, OP * goperl_a2) { return op_prepend_elem(goperl_a0, goperl_a1, goperl_a2); }
extern inline __attribute__((gnu_inline, always_inline, unused)) OP *
Perl_op_prepend_elem(pTHX_ I32 goperl_a0, OP * goperl_a1, OP * goperl_a2) { return goperl_long_op_prepend_elem(goperl_a0, goperl_a1, goperl_a2); }
static inline __attribute__((always_inline, unused)) OP *
goperl_long_op_scope(OP * goperl_a0) { return op_scope(goperl_a0); }
extern inline __attribute__((gnu_inline, always_inline, unused)) OP *
Perl_op_scope(pTHX_ OP * goperl_a0) { return goperl_long_op_scope(goperl_a0); }
static inline __attribute__((always_inline, unused)) OP *
goperl_long_op_sibling_splice(OP * goperl_a0, OP * goperl_a1, int goperl_a2, OP * goperl_a3) { return op_sibling_splice(goperl_a0, goperl_a1, goperl_a2, goperl_a3); }
extern inline __attribute__((gnu_inline, always_inline, unused)) OP *
Perl_op_sibling_splice(pTHX_ OP * goperl_a0, OP * goperl_a1, int goperl_a2, OP * goperl_a3) { return goperl_long_op_sibling_splice(goperl_a0, goperl_a1, goperl_a2, goperl_a3); }
#if defined(op_wrap_finally) && !defined(Perl_op_wrap_finally)
#define Perl_op_wrap_finally(...) GOPERL_LONG_CALL(op_wrap_finally, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(packlist) && !defined(Perl_packlist)
#define Perl_packlist(...) GOPERL_LONG_CALL(packlist, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(pad_add_anon) && !defined(Perl_pad_add_anon)
#define Perl_pad_add_anon(...) GOPERL_LONG_CALL(pad_add_anon, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(pad_add_name_pv) && !defined(Perl_pad_add_name_pv)
#define Perl_pad_add_name_pv(...) GOPERL_LONG_CALL(pad_add_name_pv, GOPERL_THX_DROP(__VA_ARGS__))
#endif
static inline __attribute__((always_inline, unused)) PADOFFSET
goperl_long_pad_add_name_pvn(const char * goperl_a0, STRLEN goperl_a1, U32 goperl_a2, HV * goperl_a3, HV * goperl_a4) { return pad_add_name_pvn(goperl_a0, goperl_a1, goperl_a2, goperl_a3, goperl_a4); }
extern inline __attribute__((gnu_inline, always_inline, unused)) PADOFFSET
Perl_pad_add_name_pvn(pTHX_ const char * goperl_a0, STRLEN goperl_a1, U32 goperl_a2, HV * goperl_a3, HV * goperl_a4) { return goperl_long_pad_add_name_pvn(goperl_a0, goperl_a1, goperl_a2, goperl_a3, goperl_a4); }
static inline __attribute__((always_inline, unused)) PADOFFSET
goperl_long_pad_add_name_sv(SV * goperl_a0, U32 goperl_a1, HV * goperl_a2, HV * goperl_a3) { return pad_add_name_sv(goperl_a0, goperl_a1, goperl_a2, goperl_a3); }
extern inline __attribute__((gnu_inline, always_inline, unused)) PADOFFSET
Perl_pad_add_name_sv(pTHX_ SV * goperl_a0, U32 goperl_a1, HV * goperl_a2, HV * goperl_a3) { return goperl_long_pad_add_name_sv(goperl_a0, goperl_a1, goperl_a2, goperl_a3); }
static inline __attribute__((always_inline, unused)) PADOFFSET
goperl_long_pad_findmy_pv(const char * goperl_a0, U32 goperl_a1) { return pad_findmy_pv(goperl_a0, goperl_a1); }
extern inline __attribute__((gnu_inline, always_inline, unused)) PADOFFSET
Perl_pad_findmy_pv(pTHX_ const char * goperl_a0, U32 goperl_a1) { return goperl_long_pad_findmy_pv(goperl_a0, goperl_a1); }
static inline __attribute__((always_inline, unused)) PADOFFSET
goperl_long_pad_findmy_pvn(const char * goperl_a0, STRLEN goperl_a1, U32 goperl_a2) { return pad_findmy_pvn(goperl_a0, goperl_a1, goperl_a2); }
extern inline __attribute__((gnu_inline, always_inline, unused)) PADOFFSET
Perl_pad_findmy_pvn(pTHX_ const char * goperl_a0, STRLEN goperl_a1, U32 goperl_a2) { return goperl_long_pad_findmy_pvn(goperl_a0, goperl_a1, goperl_a2); }
#if defined(pad_findmy_sv) && !defined(Perl_pad_findmy_sv)
#define Perl_pad_findmy_sv(...) GOPERL_LONG_CALL(pad_findmy_sv, GOPERL_THX_DROP(__VA_ARGS__))
#endif
static inline __attribute__((always_inline, unused)) PADNAME *
goperl_long_padnamelist_fetch(PADNAMELIST * goperl_a0, SSize_t goperl_a1) { return padnamelist_fetch(goperl_a0, goperl_a1); }
extern inline __attribute__((gnu_inline, always_inline, unused)) PADNAME *
Perl_padnamelist_fetch(PADNAMELIST * goperl_a0, SSize_t goperl_a1) { return goperl_long_padnamelist_fetch(goperl_a0, goperl_a1); }
#if defined(padnamelist_store) && !defined(Perl_padnamelist_store)
#define Perl_padnamelist_store(...) GOPERL_LONG_CALL(padnamelist_store, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(pad_tidy) && !defined(Perl_pad_tidy)
#define Perl_pad_tidy(...) GOPERL_LONG_CALL(pad_tidy, GOPERL_THX_DROP(__VA_ARGS__))
#endif
static inline __attribute__((always_inline, unused)) OP *
goperl_long_parse_arithexpr(U32 goperl_a0) { return parse_arithexpr(goperl_a0); }
extern inline __attribute__((gnu_inline, always_inline, unused)) OP *
Perl_parse_arithexpr(pTHX_ U32 goperl_a0) { return goperl_long_parse_arithexpr(goperl_a0); }
static inline __attribute__((always_inline, unused)) OP *
goperl_long_parse_barestmt(U32 goperl_a0) { return parse_barestmt(goperl_a0); }
extern inline __attribute__((gnu_inline, always_inline, unused)) OP *
Perl_parse_barestmt(pTHX_ U32 goperl_a0) { return goperl_long_parse_barestmt(goperl_a0); }
static inline __attribute__((always_inline, unused)) OP *
goperl_long_parse_block(U32 goperl_a0) { return parse_block(goperl_a0); }
extern inline __attribute__((gnu_inline, always_inline, unused)) OP *
Perl_parse_block(pTHX_ U32 goperl_a0) { return goperl_long_parse_block(goperl_a0); }
static inline __attribute__((always_inline, unused)) OP *
goperl_long_parse_fullexpr(U32 goperl_a0) { return parse_fullexpr(goperl_a0); }
extern inline __attribute__((gnu_inline, always_inline, unused)) OP *
Perl_parse_fullexpr(pTHX_ U32 goperl_a0) { return goperl_long_parse_fullexpr(goperl_a0); }
static inline __attribute__((always_inline, unused)) OP *
goperl_long_parse_fullstmt(U32 goperl_a0) { return parse_fullstmt(goperl_a0); }
extern inline __attribute__((gnu_inline, always_inline, unused)) OP *
Perl_parse_fullstmt(pTHX_ U32 goperl_a0) { return goperl_long_parse_fullstmt(goperl_a0); }
#if defined(parse_label) && !defined(Perl_parse_label)
#define Perl_parse_label(...) GOPERL_LONG_CALL(parse_label, GOPERL_THX_DROP(__VA_ARGS__))
#endif
static inline __attribute__((always_inline, unused)) OP *
goperl_long_parse_listexpr(U32 goperl_a0) { return parse_listexpr(goperl_a0); }
extern inline __attribute__((gnu_inline, always_inline, unused)) OP *
Perl_parse_listexpr(pTHX_ U32 goperl_a0) { return goperl_long_parse_listexpr(goperl_a0); }
static inline __attribute__((always_inline, unused)) OP *
goperl_long_parse_stmtseq(U32 goperl_a0) { return parse_stmtseq(goperl_a0); }
extern inline __attribute__((gnu_inline, always_inline, unused)) OP *
Perl_parse_stmtseq(pTHX_ U32 goperl_a0) { return goperl_long_parse_stmtseq(goperl_a0); }
static inline __attribute__((always_inline, unused)) OP *
goperl_long_parse_subsignature(U32 goperl_a0) { return parse_subsignature(goperl_a0); }
extern inline __attribute__((gnu_inline, always_inline, unused)) OP *
Perl_parse_subsignature(pTHX_ U32 goperl_a0) { return goperl_long_parse_subsignature(goperl_a0); }
static inline __attribute__((always_inline, unused)) OP *
goperl_long_parse_termexpr(U32 goperl_a0) { return parse_termexpr(goperl_a0); }
extern inline __attribute__((gnu_inline, always_inline, unused)) OP *
Perl_parse_termexpr(pTHX_ U32 goperl_a0) { return goperl_long_parse_termexpr(goperl_a0); }
#if defined(phase_name) && !defined(Perl_phase_name)
#define Perl_phase_name(...) GOPERL_LONG_CALL(phase_name, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(pmop_dump) && !defined(Perl_pmop_dump)
#define Perl_pmop_dump(...) GOPERL_LONG_CALL(pmop_dump, GOPERL_THX_DROP(__VA_ARGS__))
#endif
static inline __attribute__((always_inline, unused)) REGEXP *
goperl_long_pregcomp(SV * const goperl_a0, const U32 goperl_a1) { return pregcomp(goperl_a0, goperl_a1); }
extern inline __attribute__((gnu_inline, always_inline, unused)) REGEXP *
Perl_pregcomp(pTHX_ SV * const goperl_a0, const U32 goperl_a1) { return goperl_long_pregcomp(goperl_a0, goperl_a1); }
#if defined(pregexec) && !defined(Perl_pregexec)
#define Perl_pregexec(...) GOPERL_LONG_CALL(pregexec, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(prescan_version) && !defined(Perl_prescan_version)
#define Perl_prescan_version(...) GOPERL_LONG_CALL(prescan_version, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(ptr_table_fetch) && !defined(Perl_ptr_table_fetch)
#define Perl_ptr_table_fetch(...) GOPERL_LONG_CALL(ptr_table_fetch, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(ptr_table_free) && !defined(Perl_ptr_table_free)
#define Perl_ptr_table_free(...) GOPERL_LONG_CALL(ptr_table_free, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(ptr_table_new) && !defined(Perl_ptr_table_new)
#define Perl_ptr_table_new(...) GOPERL_LONG_CALL(ptr_table_new, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(ptr_table_split) && !defined(Perl_ptr_table_split)
#define Perl_ptr_table_split(...) GOPERL_LONG_CALL(ptr_table_split, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(ptr_table_store) && !defined(Perl_ptr_table_store)
#define Perl_ptr_table_store(...) GOPERL_LONG_CALL(ptr_table_store, GOPERL_THX_DROP(__VA_ARGS__))
#endif
static inline __attribute__((always_inline, unused)) char *
goperl_long_pv_display(SV * goperl_a0, const char * goperl_a1, STRLEN goperl_a2, STRLEN goperl_a3, STRLEN goperl_a4) { return pv_display(goperl_a0, goperl_a1, goperl_a2, goperl_a3, goperl_a4); }
extern inline __attribute__((gnu_inline, always_inline, unused)) char *
Perl_pv_display(pTHX_ SV * goperl_a0, const char * goperl_a1, STRLEN goperl_a2, STRLEN goperl_a3, STRLEN goperl_a4) { return goperl_long_pv_display(goperl_a0, goperl_a1, goperl_a2, goperl_a3, goperl_a4); }
static inline __attribute__((always_inline, unused)) char *
goperl_long_pv_escape(SV * goperl_a0, char const * const goperl_a1, const STRLEN goperl_a2, STRLEN goperl_a3, STRLEN * const goperl_a4, U32 goperl_a5) { return pv_escape(goperl_a0, goperl_a1, goperl_a2, goperl_a3, goperl_a4, goperl_a5); }
extern inline __attribute__((gnu_inline, always_inline, unused)) char *
Perl_pv_escape(pTHX_ SV * goperl_a0, char const * const goperl_a1, const STRLEN goperl_a2, STRLEN goperl_a3, STRLEN * const goperl_a4, U32 goperl_a5) { return goperl_long_pv_escape(goperl_a0, goperl_a1, goperl_a2, goperl_a3, goperl_a4, goperl_a5); }
static inline __attribute__((always_inline, unused)) char *
goperl_long_pv_pretty(SV * goperl_a0, char const * const goperl_a1, const STRLEN goperl_a2, const STRLEN goperl_a3, char const * const goperl_a4, char const * const goperl_a5, const U32 goperl_a6) { return pv_pretty(goperl_a0, goperl_a1, goperl_a2, goperl_a3, goperl_a4, goperl_a5, goperl_a6); }
extern inline __attribute__((gnu_inline, always_inline, unused)) char *
Perl_pv_pretty(pTHX_ SV * goperl_a0, char const * const goperl_a1, const STRLEN goperl_a2, const STRLEN goperl_a3, char const * const goperl_a4, char const * const goperl_a5, const U32 goperl_a6) { return goperl_long_pv_pretty(goperl_a0, goperl_a1, goperl_a2, goperl_a3, goperl_a4, goperl_a5, goperl_a6); }
static inline __attribute__((always_inline, unused)) char *
goperl_long_pv_uni_display(SV * goperl_a0, const U8 * goperl_a1, STRLEN goperl_a2, STRLEN goperl_a3, UV goperl_a4) { return pv_uni_display(goperl_a0, goperl_a1, goperl_a2, goperl_a3, goperl_a4); }
extern inline __attribute__((gnu_inline, always_inline, unused)) char *
Perl_pv_uni_display(pTHX_ SV * goperl_a0, const U8 * goperl_a1, STRLEN goperl_a2, STRLEN goperl_a3, UV goperl_a4) { return goperl_long_pv_uni_display(goperl_a0, goperl_a1, goperl_a2, goperl_a3, goperl_a4); }
#if defined(rcpv_copy) && !defined(Perl_rcpv_copy)
#define Perl_rcpv_copy(...) GOPERL_LONG_CALL(rcpv_copy, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(rcpv_free) && !defined(Perl_rcpv_free)
#define Perl_rcpv_free(...) GOPERL_LONG_CALL(rcpv_free, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(rcpv_new) && !defined(Perl_rcpv_new)
#define Perl_rcpv_new(...) GOPERL_LONG_CALL(rcpv_new, GOPERL_THX_DROP(__VA_ARGS__))
#endif
static inline __attribute__((always_inline, unused)) REGEXP *
goperl_long_re_compile(SV * const goperl_a0, U32 goperl_a1) { return re_compile(goperl_a0, goperl_a1); }
extern inline __attribute__((gnu_inline, always_inline, unused)) REGEXP *
Perl_re_compile(pTHX_ SV * const goperl_a0, U32 goperl_a1) { return goperl_long_re_compile(goperl_a0, goperl_a1); }
#if defined(repeatcpy) && !defined(Perl_repeatcpy)
#define Perl_repeatcpy(...) GOPERL_LONG_CALL(repeatcpy, __VA_ARGS__)
#endif
static inline __attribute__((always_inline, unused)) void
goperl_long_require_pv(const char * goperl_a0) { require_pv(goperl_a0); }
extern inline __attribute__((gnu_inline, always_inline, unused)) void
Perl_require_pv(pTHX_ const char * goperl_a0) { goperl_long_require_pv(goperl_a0); }
#if defined(resume_compcv_and_save) && !defined(Perl_resume_compcv_and_save)
#define Perl_resume_compcv_and_save(...) GOPERL_LONG_CALL(resume_compcv_and_save, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(resume_compcv_final) && !defined(Perl_resume_compcv_final)
#define Perl_resume_compcv_final(...) GOPERL_LONG_CALL(resume_compcv_final, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(rninstr) && !defined(Perl_rninstr)
#define Perl_rninstr(...) GOPERL_LONG_CALL(rninstr, __VA_ARGS__)
#endif
#if defined(rpp_context) && !defined(Perl_rpp_context)
#define Perl_rpp_context(...) GOPERL_LONG_CALL(rpp_context, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(rpp_extend) && !defined(Perl_rpp_extend)
#define Perl_rpp_extend(...) GOPERL_LONG_CALL(rpp_extend, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(rpp_invoke_xs) && !defined(Perl_rpp_invoke_xs)
#define Perl_rpp_invoke_xs(...) GOPERL_LONG_CALL(rpp_invoke_xs, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(rpp_is_lone) && !defined(Perl_rpp_is_lone)
#define Perl_rpp_is_lone(...) GOPERL_LONG_CALL(rpp_is_lone, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(rpp_popfree_1) && !defined(Perl_rpp_popfree_1)
#define Perl_rpp_popfree_1(...) GOPERL_LONG_CALL(rpp_popfree_1, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(rpp_popfree_2) && !defined(Perl_rpp_popfree_2)
#define Perl_rpp_popfree_2(...) GOPERL_LONG_CALL(rpp_popfree_2, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(rpp_popfree_1_NN) && !defined(Perl_rpp_popfree_1_NN)
#define Perl_rpp_popfree_1_NN(...) GOPERL_LONG_CALL(rpp_popfree_1_NN, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(rpp_popfree_2_NN) && !defined(Perl_rpp_popfree_2_NN)
#define Perl_rpp_popfree_2_NN(...) GOPERL_LONG_CALL(rpp_popfree_2_NN, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(rpp_popfree_to) && !defined(Perl_rpp_popfree_to)
#define Perl_rpp_popfree_to(...) GOPERL_LONG_CALL(rpp_popfree_to, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(rpp_popfree_to_NN) && !defined(Perl_rpp_popfree_to_NN)
#define Perl_rpp_popfree_to_NN(...) GOPERL_LONG_CALL(rpp_popfree_to_NN, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(rpp_pop_1_norc) && !defined(Perl_rpp_pop_1_norc)
#define Perl_rpp_pop_1_norc(...) GOPERL_LONG_CALL(rpp_pop_1_norc, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(rpp_push_1) && !defined(Perl_rpp_push_1)
#define Perl_rpp_push_1(...) GOPERL_LONG_CALL(rpp_push_1, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(rpp_push_2) && !defined(Perl_rpp_push_2)
#define Perl_rpp_push_2(...) GOPERL_LONG_CALL(rpp_push_2, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(rpp_push_IMM) && !defined(Perl_rpp_push_IMM)
#define Perl_rpp_push_IMM(...) GOPERL_LONG_CALL(rpp_push_IMM, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(rpp_push_1_norc) && !defined(Perl_rpp_push_1_norc)
#define Perl_rpp_push_1_norc(...) GOPERL_LONG_CALL(rpp_push_1_norc, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(rpp_replace_1_1) && !defined(Perl_rpp_replace_1_1)
#define Perl_rpp_replace_1_1(...) GOPERL_LONG_CALL(rpp_replace_1_1, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(rpp_replace_2_1) && !defined(Perl_rpp_replace_2_1)
#define Perl_rpp_replace_2_1(...) GOPERL_LONG_CALL(rpp_replace_2_1, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(rpp_replace_at) && !defined(Perl_rpp_replace_at)
#define Perl_rpp_replace_at(...) GOPERL_LONG_CALL(rpp_replace_at, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(rpp_replace_at_NN) && !defined(Perl_rpp_replace_at_NN)
#define Perl_rpp_replace_at_NN(...) GOPERL_LONG_CALL(rpp_replace_at_NN, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(rpp_replace_at_norc) && !defined(Perl_rpp_replace_at_norc)
#define Perl_rpp_replace_at_norc(...) GOPERL_LONG_CALL(rpp_replace_at_norc, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(rpp_replace_at_norc_NN) && !defined(Perl_rpp_replace_at_norc_NN)
#define Perl_rpp_replace_at_norc_NN(...) GOPERL_LONG_CALL(rpp_replace_at_norc_NN, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(rpp_replace_1_IMM_NN) && !defined(Perl_rpp_replace_1_IMM_NN)
#define Perl_rpp_replace_1_IMM_NN(...) GOPERL_LONG_CALL(rpp_replace_1_IMM_NN, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(rpp_replace_2_IMM_NN) && !defined(Perl_rpp_replace_2_IMM_NN)
#define Perl_rpp_replace_2_IMM_NN(...) GOPERL_LONG_CALL(rpp_replace_2_IMM_NN, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(rpp_replace_1_1_NN) && !defined(Perl_rpp_replace_1_1_NN)
#define Perl_rpp_replace_1_1_NN(...) GOPERL_LONG_CALL(rpp_replace_1_1_NN, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(rpp_replace_2_1_NN) && !defined(Perl_rpp_replace_2_1_NN)
#define Perl_rpp_replace_2_1_NN(...) GOPERL_LONG_CALL(rpp_replace_2_1_NN, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(rpp_stack_is_rc) && !defined(Perl_rpp_stack_is_rc)
#define Perl_rpp_stack_is_rc(...) GOPERL_LONG_CALL(rpp_stack_is_rc, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(rpp_try_AMAGIC_1) && !defined(Perl_rpp_try_AMAGIC_1)
#define Perl_rpp_try_AMAGIC_1(...) GOPERL_LONG_CALL(rpp_try_AMAGIC_1, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(rpp_try_AMAGIC_2) && !defined(Perl_rpp_try_AMAGIC_2)
#define Perl_rpp_try_AMAGIC_2(...) GOPERL_LONG_CALL(rpp_try_AMAGIC_2, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(rpp_xpush_1) && !defined(Perl_rpp_xpush_1)
#define Perl_rpp_xpush_1(...) GOPERL_LONG_CALL(rpp_xpush_1, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(rpp_xpush_2) && !defined(Perl_rpp_xpush_2)
#define Perl_rpp_xpush_2(...) GOPERL_LONG_CALL(rpp_xpush_2, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(rpp_xpush_IMM) && !defined(Perl_rpp_xpush_IMM)
#define Perl_rpp_xpush_IMM(...) GOPERL_LONG_CALL(rpp_xpush_IMM, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(rsignal) && !defined(Perl_rsignal)
#define Perl_rsignal(...) GOPERL_LONG_CALL(rsignal, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(rsignal_state) && !defined(Perl_rsignal_state)
#define Perl_rsignal_state(...) GOPERL_LONG_CALL(rsignal_state, GOPERL_THX_DROP(__VA_ARGS__))
#endif
static inline __attribute__((always_inline, unused)) CV *
goperl_long_rv2cv_op_cv(OP * goperl_a0, U32 goperl_a1) { return rv2cv_op_cv(goperl_a0, goperl_a1); }
extern inline __attribute__((gnu_inline, always_inline, unused)) CV *
Perl_rv2cv_op_cv(pTHX_ OP * goperl_a0, U32 goperl_a1) { return goperl_long_rv2cv_op_cv(goperl_a0, goperl_a1); }
static inline __attribute__((always_inline, unused)) Malloc_t
goperl_long_safesyscalloc(MEM_SIZE goperl_a0, MEM_SIZE goperl_a1) { return safesyscalloc(goperl_a0, goperl_a1); }
extern inline __attribute__((gnu_inline, always_inline, unused)) Malloc_t
Perl_safesyscalloc(MEM_SIZE goperl_a0, MEM_SIZE goperl_a1) { return goperl_long_safesyscalloc(goperl_a0, goperl_a1); }
static inline __attribute__((always_inline, unused)) Free_t
goperl_long_safesysfree(Malloc_t goperl_a0) { return safesysfree(goperl_a0); }
extern inline __attribute__((gnu_inline, always_inline, unused)) Free_t
Perl_safesysfree(Malloc_t goperl_a0) { return goperl_long_safesysfree(goperl_a0); }
static inline __attribute__((always_inline, unused)) Malloc_t
goperl_long_safesysmalloc(MEM_SIZE goperl_a0) { return safesysmalloc(goperl_a0); }
extern inline __attribute__((gnu_inline, always_inline, unused)) Malloc_t
Perl_safesysmalloc(MEM_SIZE goperl_a0) { return goperl_long_safesysmalloc(goperl_a0); }
static inline __attribute__((always_inline, unused)) Malloc_t
goperl_long_safesysrealloc(Malloc_t goperl_a0, MEM_SIZE goperl_a1) { return safesysrealloc(goperl_a0, goperl_a1); }
extern inline __attribute__((gnu_inline, always_inline, unused)) Malloc_t
Perl_safesysrealloc(Malloc_t goperl_a0, MEM_SIZE goperl_a1) { return goperl_long_safesysrealloc(goperl_a0, goperl_a1); }
#if defined(save_aelem) && !defined(Perl_save_aelem)
#define Perl_save_aelem(...) GOPERL_LONG_CALL(save_aelem, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(save_aelem_flags) && !defined(Perl_save_aelem_flags)
#define Perl_save_aelem_flags(...) GOPERL_LONG_CALL(save_aelem_flags, GOPERL_THX_DROP(__VA_ARGS__))
#endif
static inline __attribute__((always_inline, unused)) void
goperl_long_save_aptr(AV ** goperl_a0) { save_aptr(goperl_a0); }
extern inline __attribute__((gnu_inline, always_inline, unused)) void
Perl_save_aptr(pTHX_ AV ** goperl_a0) { goperl_long_save_aptr(goperl_a0); }
static inline __attribute__((always_inline, unused)) AV *
goperl_long_save_ary(GV * goperl_a0) { return save_ary(goperl_a0); }
extern inline __attribute__((gnu_inline, always_inline, unused)) AV *
Perl_save_ary(pTHX_ GV * goperl_a0) { return goperl_long_save_ary(goperl_a0); }
static inline __attribute__((always_inline, unused)) void
goperl_long_save_gp(GV * goperl_a0, I32 goperl_a1) { save_gp(goperl_a0, goperl_a1); }
extern inline __attribute__((gnu_inline, always_inline, unused)) void
Perl_save_gp(pTHX_ GV * goperl_a0, I32 goperl_a1) { goperl_long_save_gp(goperl_a0, goperl_a1); }
static inline __attribute__((always_inline, unused)) HV *
goperl_long_save_hash(GV * goperl_a0) { return save_hash(goperl_a0); }
extern inline __attribute__((gnu_inline, always_inline, unused)) HV *
Perl_save_hash(pTHX_ GV * goperl_a0) { return goperl_long_save_hash(goperl_a0); }
static inline __attribute__((always_inline, unused)) void
goperl_long_save_helem(HV * goperl_a0, SV * goperl_a1, SV ** goperl_a2) { save_helem(goperl_a0, goperl_a1, goperl_a2); }
extern inline __attribute__((gnu_inline, always_inline, unused)) void
Perl_save_helem(pTHX_ HV * goperl_a0, SV * goperl_a1, SV ** goperl_a2) { goperl_long_save_helem(goperl_a0, goperl_a1, goperl_a2); }
static inline __attribute__((always_inline, unused)) void
goperl_long_save_helem_flags(HV * goperl_a0, SV * goperl_a1, SV ** goperl_a2, const U32 goperl_a3) { save_helem_flags(goperl_a0, goperl_a1, goperl_a2, goperl_a3); }
extern inline __attribute__((gnu_inline, always_inline, unused)) void
Perl_save_helem_flags(pTHX_ HV * goperl_a0, SV * goperl_a1, SV ** goperl_a2, const U32 goperl_a3) { goperl_long_save_helem_flags(goperl_a0, goperl_a1, goperl_a2, goperl_a3); }
static inline __attribute__((always_inline, unused)) void
goperl_long_save_hptr(HV ** goperl_a0) { save_hptr(goperl_a0); }
extern inline __attribute__((gnu_inline, always_inline, unused)) void
Perl_save_hptr(pTHX_ HV ** goperl_a0) { goperl_long_save_hptr(goperl_a0); }
static inline __attribute__((always_inline, unused)) void
goperl_long_save_item(SV * goperl_a0) { save_item(goperl_a0); }
extern inline __attribute__((gnu_inline, always_inline, unused)) void
Perl_save_item(pTHX_ SV * goperl_a0) { goperl_long_save_item(goperl_a0); }
static inline __attribute__((always_inline, unused)) char *
goperl_long_savepv(const char * goperl_a0) { return savepv(goperl_a0); }
extern inline __attribute__((gnu_inline, always_inline, unused)) char *
Perl_savepv(pTHX_ const char * goperl_a0) { return goperl_long_savepv(goperl_a0); }
static inline __attribute__((always_inline, unused)) char *
goperl_long_savepvn(const char * goperl_a0, Size_t goperl_a1) { return savepvn(goperl_a0, goperl_a1); }
extern inline __attribute__((gnu_inline, always_inline, unused)) char *
Perl_savepvn(pTHX_ const char * goperl_a0, Size_t goperl_a1) { return goperl_long_savepvn(goperl_a0, goperl_a1); }
static inline __attribute__((always_inline, unused)) SV *
goperl_long_save_scalar(GV * goperl_a0) { return save_scalar(goperl_a0); }
extern inline __attribute__((gnu_inline, always_inline, unused)) SV *
Perl_save_scalar(pTHX_ GV * goperl_a0) { return goperl_long_save_scalar(goperl_a0); }
static inline __attribute__((always_inline, unused)) char *
goperl_long_savesharedpv(const char * goperl_a0) { return savesharedpv(goperl_a0); }
extern inline __attribute__((gnu_inline, always_inline, unused)) char *
Perl_savesharedpv(pTHX_ const char * goperl_a0) { return goperl_long_savesharedpv(goperl_a0); }
static inline __attribute__((always_inline, unused)) char *
goperl_long_savesharedpvn(const char * const goperl_a0, const STRLEN goperl_a1) { return savesharedpvn(goperl_a0, goperl_a1); }
extern inline __attribute__((gnu_inline, always_inline, unused)) char *
Perl_savesharedpvn(pTHX_ const char * const goperl_a0, const STRLEN goperl_a1) { return goperl_long_savesharedpvn(goperl_a0, goperl_a1); }
#if defined(savesharedsvpv) && !defined(Perl_savesharedsvpv)
#define Perl_savesharedsvpv(...) GOPERL_LONG_CALL(savesharedsvpv, GOPERL_THX_DROP(__VA_ARGS__))
#endif
static inline __attribute__((always_inline, unused)) char *
goperl_long_savesvpv(SV * goperl_a0) { return savesvpv(goperl_a0); }
extern inline __attribute__((gnu_inline, always_inline, unused)) char *
Perl_savesvpv(pTHX_ SV * goperl_a0) { return goperl_long_savesvpv(goperl_a0); }
#if defined(save_svref) && !defined(Perl_save_svref)
#define Perl_save_svref(...) GOPERL_LONG_CALL(save_svref, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(savetmps) && !defined(Perl_savetmps)
#define Perl_savetmps(...) GOPERL_LONG_CALL(savetmps, GOPERL_THX_DROP(__VA_ARGS__))
#endif
static inline __attribute__((always_inline, unused)) NV
goperl_long_scan_bin(const char * goperl_a0, STRLEN goperl_a1, STRLEN * goperl_a2) { return scan_bin(goperl_a0, goperl_a1, goperl_a2); }
extern inline __attribute__((gnu_inline, always_inline, unused)) NV
Perl_scan_bin(pTHX_ const char * goperl_a0, STRLEN goperl_a1, STRLEN * goperl_a2) { return goperl_long_scan_bin(goperl_a0, goperl_a1, goperl_a2); }
static inline __attribute__((always_inline, unused)) NV
goperl_long_scan_hex(const char * goperl_a0, STRLEN goperl_a1, STRLEN * goperl_a2) { return scan_hex(goperl_a0, goperl_a1, goperl_a2); }
extern inline __attribute__((gnu_inline, always_inline, unused)) NV
Perl_scan_hex(pTHX_ const char * goperl_a0, STRLEN goperl_a1, STRLEN * goperl_a2) { return goperl_long_scan_hex(goperl_a0, goperl_a1, goperl_a2); }
static inline __attribute__((always_inline, unused)) NV
goperl_long_scan_oct(const char * goperl_a0, STRLEN goperl_a1, STRLEN * goperl_a2) { return scan_oct(goperl_a0, goperl_a1, goperl_a2); }
extern inline __attribute__((gnu_inline, always_inline, unused)) NV
Perl_scan_oct(pTHX_ const char * goperl_a0, STRLEN goperl_a1, STRLEN * goperl_a2) { return goperl_long_scan_oct(goperl_a0, goperl_a1, goperl_a2); }
#if defined(scan_vstring) && !defined(Perl_scan_vstring)
#define Perl_scan_vstring(...) GOPERL_LONG_CALL(scan_vstring, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(setdefout) && !defined(Perl_setdefout)
#define Perl_setdefout(...) GOPERL_LONG_CALL(setdefout, GOPERL_THX_DROP(__VA_ARGS__))
#endif
static inline __attribute__((always_inline, unused)) void
goperl_long_sortsv(SV ** goperl_a0, size_t goperl_a1, SVCOMPARE_t goperl_a2) { sortsv(goperl_a0, goperl_a1, goperl_a2); }
extern inline __attribute__((gnu_inline, always_inline, unused)) void
Perl_sortsv(pTHX_ SV ** goperl_a0, size_t goperl_a1, SVCOMPARE_t goperl_a2) { goperl_long_sortsv(goperl_a0, goperl_a1, goperl_a2); }
static inline __attribute__((always_inline, unused)) void
goperl_long_sortsv_flags(SV ** goperl_a0, size_t goperl_a1, SVCOMPARE_t goperl_a2, U32 goperl_a3) { sortsv_flags(goperl_a0, goperl_a1, goperl_a2, goperl_a3); }
extern inline __attribute__((gnu_inline, always_inline, unused)) void
Perl_sortsv_flags(pTHX_ SV ** goperl_a0, size_t goperl_a1, SVCOMPARE_t goperl_a2, U32 goperl_a3) { goperl_long_sortsv_flags(goperl_a0, goperl_a1, goperl_a2, goperl_a3); }
static inline __attribute__((always_inline, unused)) I32
goperl_long_start_subparse(I32 goperl_a0, U32 goperl_a1) { return start_subparse(goperl_a0, goperl_a1); }
extern inline __attribute__((gnu_inline, always_inline, unused)) I32
Perl_start_subparse(pTHX_ I32 goperl_a0, U32 goperl_a1) { return goperl_long_start_subparse(goperl_a0, goperl_a1); }
#if defined(strict_utf8_to_uv) && !defined(Perl_strict_utf8_to_uv)
#define Perl_strict_utf8_to_uv(...) GOPERL_LONG_CALL(strict_utf8_to_uv, __VA_ARGS__)
#endif
#if defined(suspend_compcv) && !defined(Perl_suspend_compcv)
#define Perl_suspend_compcv(...) GOPERL_LONG_CALL(suspend_compcv, GOPERL_THX_DROP(__VA_ARGS__))
#endif
static inline __attribute__((always_inline, unused)) void
goperl_long_SvAMAGIC_off(SV * goperl_a0) { SvAMAGIC_off(goperl_a0); }
extern inline __attribute__((gnu_inline, always_inline, unused)) void
Perl_SvAMAGIC_off(SV * goperl_a0) { goperl_long_SvAMAGIC_off(goperl_a0); }
static inline __attribute__((always_inline, unused)) void
goperl_long_SvAMAGIC_on(SV * goperl_a0) { SvAMAGIC_on(goperl_a0); }
extern inline __attribute__((gnu_inline, always_inline, unused)) void
Perl_SvAMAGIC_on(SV * goperl_a0) { goperl_long_SvAMAGIC_on(goperl_a0); }
static inline __attribute__((always_inline, unused)) void
goperl_long_sv_backoff(SV * const goperl_a0) { sv_backoff(goperl_a0); }
extern inline __attribute__((gnu_inline, always_inline, unused)) void
Perl_sv_backoff(SV * const goperl_a0) { goperl_long_sv_backoff(goperl_a0); }
static inline __attribute__((always_inline, unused)) SV *
goperl_long_sv_bless(SV * const goperl_a0, HV * const goperl_a1) { return sv_bless(goperl_a0, goperl_a1); }
extern inline __attribute__((gnu_inline, always_inline, unused)) SV *
Perl_sv_bless(pTHX_ SV * const goperl_a0, HV * const goperl_a1) { return goperl_long_sv_bless(goperl_a0, goperl_a1); }
static inline __attribute__((always_inline, unused)) void
goperl_long_sv_catpv(SV * const goperl_a0, const char * goperl_a1) { sv_catpv(goperl_a0, goperl_a1); }
extern inline __attribute__((gnu_inline, always_inline, unused)) void
Perl_sv_catpv(pTHX_ SV * const goperl_a0, const char * goperl_a1) { goperl_long_sv_catpv(goperl_a0, goperl_a1); }
static inline __attribute__((always_inline, unused)) void
goperl_long_sv_catpv_flags(SV * const goperl_a0, const char * goperl_a1, const I32 goperl_a2) { sv_catpv_flags(goperl_a0, goperl_a1, goperl_a2); }
extern inline __attribute__((gnu_inline, always_inline, unused)) void
Perl_sv_catpv_flags(pTHX_ SV * const goperl_a0, const char * goperl_a1, const I32 goperl_a2) { goperl_long_sv_catpv_flags(goperl_a0, goperl_a1, goperl_a2); }
#if defined(sv_catpvf_mg) && !defined(Perl_sv_catpvf_mg)
#define Perl_sv_catpvf_mg(...) GOPERL_LONG_CALL(sv_catpvf_mg, GOPERL_THX_DROP(__VA_ARGS__))
#endif
static inline __attribute__((always_inline, unused)) void
goperl_long_sv_catpv_mg(SV * const goperl_a0, const char * const goperl_a1) { sv_catpv_mg(goperl_a0, goperl_a1); }
extern inline __attribute__((gnu_inline, always_inline, unused)) void
Perl_sv_catpv_mg(pTHX_ SV * const goperl_a0, const char * const goperl_a1) { goperl_long_sv_catpv_mg(goperl_a0, goperl_a1); }
static inline __attribute__((always_inline, unused)) void
goperl_long_sv_catpvn(SV * const goperl_a0, const char * goperl_a1, STRLEN goperl_a2) { sv_catpvn(goperl_a0, goperl_a1, goperl_a2); }
extern inline __attribute__((gnu_inline, always_inline, unused)) void
Perl_sv_catpvn(pTHX_ SV * const goperl_a0, const char * goperl_a1, STRLEN goperl_a2) { goperl_long_sv_catpvn(goperl_a0, goperl_a1, goperl_a2); }
static inline __attribute__((always_inline, unused)) void
goperl_long_sv_catpvn_flags(SV * const goperl_a0, const char * goperl_a1, const STRLEN goperl_a2, const I32 goperl_a3) { sv_catpvn_flags(goperl_a0, goperl_a1, goperl_a2, goperl_a3); }
extern inline __attribute__((gnu_inline, always_inline, unused)) void
Perl_sv_catpvn_flags(pTHX_ SV * const goperl_a0, const char * goperl_a1, const STRLEN goperl_a2, const I32 goperl_a3) { goperl_long_sv_catpvn_flags(goperl_a0, goperl_a1, goperl_a2, goperl_a3); }
static inline __attribute__((always_inline, unused)) void
goperl_long_sv_catpvn_mg(SV * const goperl_a0, const char * goperl_a1, STRLEN goperl_a2) { sv_catpvn_mg(goperl_a0, goperl_a1, goperl_a2); }
extern inline __attribute__((gnu_inline, always_inline, unused)) void
Perl_sv_catpvn_mg(pTHX_ SV * const goperl_a0, const char * goperl_a1, STRLEN goperl_a2) { goperl_long_sv_catpvn_mg(goperl_a0, goperl_a1, goperl_a2); }
static inline __attribute__((always_inline, unused)) void
goperl_long_sv_catsv(SV * const goperl_a0, SV * const goperl_a1) { sv_catsv(goperl_a0, goperl_a1); }
extern inline __attribute__((gnu_inline, always_inline, unused)) void
Perl_sv_catsv(pTHX_ SV * const goperl_a0, SV * const goperl_a1) { goperl_long_sv_catsv(goperl_a0, goperl_a1); }
static inline __attribute__((always_inline, unused)) void
goperl_long_sv_catsv_flags(SV * const goperl_a0, SV * const goperl_a1, const I32 goperl_a2) { sv_catsv_flags(goperl_a0, goperl_a1, goperl_a2); }
extern inline __attribute__((gnu_inline, always_inline, unused)) void
Perl_sv_catsv_flags(pTHX_ SV * const goperl_a0, SV * const goperl_a1, const I32 goperl_a2) { goperl_long_sv_catsv_flags(goperl_a0, goperl_a1, goperl_a2); }
static inline __attribute__((always_inline, unused)) void
goperl_long_sv_catsv_mg(SV * const goperl_a0, SV * const goperl_a1) { sv_catsv_mg(goperl_a0, goperl_a1); }
extern inline __attribute__((gnu_inline, always_inline, unused)) void
Perl_sv_catsv_mg(pTHX_ SV * const goperl_a0, SV * const goperl_a1) { goperl_long_sv_catsv_mg(goperl_a0, goperl_a1); }
static inline __attribute__((always_inline, unused)) void
goperl_long_sv_chop(SV * const goperl_a0, const char * const goperl_a1) { sv_chop(goperl_a0, goperl_a1); }
extern inline __attribute__((gnu_inline, always_inline, unused)) void
Perl_sv_chop(pTHX_ SV * const goperl_a0, const char * const goperl_a1) { goperl_long_sv_chop(goperl_a0, goperl_a1); }
#if defined(sv_clear) && !defined(Perl_sv_clear)
#define Perl_sv_clear(...) GOPERL_LONG_CALL(sv_clear, GOPERL_THX_DROP(__VA_ARGS__))
#endif
static inline __attribute__((always_inline, unused)) I32
goperl_long_sv_cmp_flags(SV * const goperl_a0, SV * const goperl_a1, const U32 goperl_a2) { return sv_cmp_flags(goperl_a0, goperl_a1, goperl_a2); }
extern inline __attribute__((gnu_inline, always_inline, unused)) I32
Perl_sv_cmp_flags(pTHX_ SV * const goperl_a0, SV * const goperl_a1, const U32 goperl_a2) { return goperl_long_sv_cmp_flags(goperl_a0, goperl_a1, goperl_a2); }
static inline __attribute__((always_inline, unused)) I32
goperl_long_sv_cmp_locale_flags(SV * const goperl_a0, SV * const goperl_a1, const U32 goperl_a2) { return sv_cmp_locale_flags(goperl_a0, goperl_a1, goperl_a2); }
extern inline __attribute__((gnu_inline, always_inline, unused)) I32
Perl_sv_cmp_locale_flags(pTHX_ SV * const goperl_a0, SV * const goperl_a1, const U32 goperl_a2) { return goperl_long_sv_cmp_locale_flags(goperl_a0, goperl_a1, goperl_a2); }
static inline __attribute__((always_inline, unused)) void
goperl_long_sv_copypv(SV * const goperl_a0, SV * const goperl_a1) { sv_copypv(goperl_a0, goperl_a1); }
extern inline __attribute__((gnu_inline, always_inline, unused)) void
Perl_sv_copypv(pTHX_ SV * const goperl_a0, SV * const goperl_a1) { goperl_long_sv_copypv(goperl_a0, goperl_a1); }
#if defined(sv_copypv_flags) && !defined(Perl_sv_copypv_flags)
#define Perl_sv_copypv_flags(...) GOPERL_LONG_CALL(sv_copypv_flags, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(sv_copypv_nomg) && !defined(Perl_sv_copypv_nomg)
#define Perl_sv_copypv_nomg(...) GOPERL_LONG_CALL(sv_copypv_nomg, GOPERL_THX_DROP(__VA_ARGS__))
#endif
static inline __attribute__((always_inline, unused)) CV *
goperl_long_sv_2cv(SV * goperl_a0, HV ** const goperl_a1, GV ** const goperl_a2, const I32 goperl_a3) { return sv_2cv(goperl_a0, goperl_a1, goperl_a2, goperl_a3); }
extern inline __attribute__((gnu_inline, always_inline, unused)) CV *
Perl_sv_2cv(pTHX_ SV * goperl_a0, HV ** const goperl_a1, GV ** const goperl_a2, const I32 goperl_a3) { return goperl_long_sv_2cv(goperl_a0, goperl_a1, goperl_a2, goperl_a3); }
static inline __attribute__((always_inline, unused)) void
goperl_long_sv_dec(SV * const goperl_a0) { sv_dec(goperl_a0); }
extern inline __attribute__((gnu_inline, always_inline, unused)) void
Perl_sv_dec(pTHX_ SV * const goperl_a0) { goperl_long_sv_dec(goperl_a0); }
#if defined(sv_dec_nomg) && !defined(Perl_sv_dec_nomg)
#define Perl_sv_dec_nomg(...) GOPERL_LONG_CALL(sv_dec_nomg, GOPERL_THX_DROP(__VA_ARGS__))
#endif
static inline __attribute__((always_inline, unused)) bool
goperl_long_sv_derived_from(SV * goperl_a0, const char * const goperl_a1) { return sv_derived_from(goperl_a0, goperl_a1); }
extern inline __attribute__((gnu_inline, always_inline, unused)) bool
Perl_sv_derived_from(pTHX_ SV * goperl_a0, const char * const goperl_a1) { return goperl_long_sv_derived_from(goperl_a0, goperl_a1); }
#if defined(sv_derived_from_hv) && !defined(Perl_sv_derived_from_hv)
#define Perl_sv_derived_from_hv(...) GOPERL_LONG_CALL(sv_derived_from_hv, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(sv_derived_from_pv) && !defined(Perl_sv_derived_from_pv)
#define Perl_sv_derived_from_pv(...) GOPERL_LONG_CALL(sv_derived_from_pv, GOPERL_THX_DROP(__VA_ARGS__))
#endif
static inline __attribute__((always_inline, unused)) bool
goperl_long_sv_derived_from_pvn(SV * goperl_a0, const char * const goperl_a1, const STRLEN goperl_a2, U32 goperl_a3) { return sv_derived_from_pvn(goperl_a0, goperl_a1, goperl_a2, goperl_a3); }
extern inline __attribute__((gnu_inline, always_inline, unused)) bool
Perl_sv_derived_from_pvn(pTHX_ SV * goperl_a0, const char * const goperl_a1, const STRLEN goperl_a2, U32 goperl_a3) { return goperl_long_sv_derived_from_pvn(goperl_a0, goperl_a1, goperl_a2, goperl_a3); }
static inline __attribute__((always_inline, unused)) bool
goperl_long_sv_derived_from_sv(SV * goperl_a0, SV * goperl_a1, U32 goperl_a2) { return sv_derived_from_sv(goperl_a0, goperl_a1, goperl_a2); }
extern inline __attribute__((gnu_inline, always_inline, unused)) bool
Perl_sv_derived_from_sv(pTHX_ SV * goperl_a0, SV * goperl_a1, U32 goperl_a2) { return goperl_long_sv_derived_from_sv(goperl_a0, goperl_a1, goperl_a2); }
#if defined(sv_destroyable) && !defined(Perl_sv_destroyable)
#define Perl_sv_destroyable(...) GOPERL_LONG_CALL(sv_destroyable, GOPERL_THX_DROP(__VA_ARGS__))
#endif
static inline __attribute__((always_inline, unused)) bool
goperl_long_sv_does(SV * goperl_a0, const char * const goperl_a1) { return sv_does(goperl_a0, goperl_a1); }
extern inline __attribute__((gnu_inline, always_inline, unused)) bool
Perl_sv_does(pTHX_ SV * goperl_a0, const char * const goperl_a1) { return goperl_long_sv_does(goperl_a0, goperl_a1); }
static inline __attribute__((always_inline, unused)) bool
goperl_long_sv_does_pv(SV * goperl_a0, const char * const goperl_a1, U32 goperl_a2) { return sv_does_pv(goperl_a0, goperl_a1, goperl_a2); }
extern inline __attribute__((gnu_inline, always_inline, unused)) bool
Perl_sv_does_pv(pTHX_ SV * goperl_a0, const char * const goperl_a1, U32 goperl_a2) { return goperl_long_sv_does_pv(goperl_a0, goperl_a1, goperl_a2); }
static inline __attribute__((always_inline, unused)) bool
goperl_long_sv_does_pvn(SV * goperl_a0, const char * const goperl_a1, const STRLEN goperl_a2, U32 goperl_a3) { return sv_does_pvn(goperl_a0, goperl_a1, goperl_a2, goperl_a3); }
extern inline __attribute__((gnu_inline, always_inline, unused)) bool
Perl_sv_does_pvn(pTHX_ SV * goperl_a0, const char * const goperl_a1, const STRLEN goperl_a2, U32 goperl_a3) { return goperl_long_sv_does_pvn(goperl_a0, goperl_a1, goperl_a2, goperl_a3); }
#if defined(sv_does_sv) && !defined(Perl_sv_does_sv)
#define Perl_sv_does_sv(...) GOPERL_LONG_CALL(sv_does_sv, GOPERL_THX_DROP(__VA_ARGS__))
#endif
static inline __attribute__((always_inline, unused)) void
goperl_long_sv_dump(SV * goperl_a0) { sv_dump(goperl_a0); }
extern inline __attribute__((gnu_inline, always_inline, unused)) void
Perl_sv_dump(pTHX_ SV * goperl_a0) { goperl_long_sv_dump(goperl_a0); }
#if defined(sv_dump_depth) && !defined(Perl_sv_dump_depth)
#define Perl_sv_dump_depth(...) GOPERL_LONG_CALL(sv_dump_depth, GOPERL_THX_DROP(__VA_ARGS__))
#endif
static inline __attribute__((always_inline, unused)) I32
goperl_long_sv_eq(SV * goperl_a0, SV * goperl_a1) { return sv_eq(goperl_a0, goperl_a1); }
extern inline __attribute__((gnu_inline, always_inline, unused)) I32
Perl_sv_eq(pTHX_ SV * goperl_a0, SV * goperl_a1) { return goperl_long_sv_eq(goperl_a0, goperl_a1); }
static inline __attribute__((always_inline, unused)) I32
goperl_long_sv_eq_flags(SV * goperl_a0, SV * goperl_a1, const U32 goperl_a2) { return sv_eq_flags(goperl_a0, goperl_a1, goperl_a2); }
extern inline __attribute__((gnu_inline, always_inline, unused)) I32
Perl_sv_eq_flags(pTHX_ SV * goperl_a0, SV * goperl_a1, const U32 goperl_a2) { return goperl_long_sv_eq_flags(goperl_a0, goperl_a1, goperl_a2); }
static inline __attribute__((always_inline, unused)) void
goperl_long_sv_force_normal(SV * goperl_a0) { sv_force_normal(goperl_a0); }
extern inline __attribute__((gnu_inline, always_inline, unused)) void
Perl_sv_force_normal(pTHX_ SV * goperl_a0) { goperl_long_sv_force_normal(goperl_a0); }
static inline __attribute__((always_inline, unused)) void
goperl_long_sv_force_normal_flags(SV * const goperl_a0, const U32 goperl_a1) { sv_force_normal_flags(goperl_a0, goperl_a1); }
extern inline __attribute__((gnu_inline, always_inline, unused)) void
Perl_sv_force_normal_flags(pTHX_ SV * const goperl_a0, const U32 goperl_a1) { goperl_long_sv_force_normal_flags(goperl_a0, goperl_a1); }
static inline __attribute__((always_inline, unused)) void
goperl_long_sv_free(SV * const goperl_a0) { sv_free(goperl_a0); }
extern inline __attribute__((gnu_inline, always_inline, unused)) void
Perl_sv_free(pTHX_ SV * const goperl_a0) { goperl_long_sv_free(goperl_a0); }
#if defined(sv_get_backrefs) && !defined(Perl_sv_get_backrefs)
#define Perl_sv_get_backrefs(...) GOPERL_LONG_CALL(sv_get_backrefs, __VA_ARGS__)
#endif
static inline __attribute__((always_inline, unused)) void
goperl_long_SvGETMAGIC(SV * goperl_a0) { SvGETMAGIC(goperl_a0); }
extern inline __attribute__((gnu_inline, always_inline, unused)) void
Perl_SvGETMAGIC(pTHX_ SV * goperl_a0) { goperl_long_SvGETMAGIC(goperl_a0); }
static inline __attribute__((always_inline, unused)) void
goperl_long_sv_inc(SV * const goperl_a0) { sv_inc(goperl_a0); }
extern inline __attribute__((gnu_inline, always_inline, unused)) void
Perl_sv_inc(pTHX_ SV * const goperl_a0) { goperl_long_sv_inc(goperl_a0); }
#if defined(sv_inc_nomg) && !defined(Perl_sv_inc_nomg)
#define Perl_sv_inc_nomg(...) GOPERL_LONG_CALL(sv_inc_nomg, GOPERL_THX_DROP(__VA_ARGS__))
#endif
static inline __attribute__((always_inline, unused)) void
goperl_long_sv_insert(SV * const goperl_a0, const STRLEN goperl_a1, const STRLEN goperl_a2, const char * const goperl_a3, const STRLEN goperl_a4) { sv_insert(goperl_a0, goperl_a1, goperl_a2, goperl_a3, goperl_a4); }
extern inline __attribute__((gnu_inline, always_inline, unused)) void
Perl_sv_insert(pTHX_ SV * const goperl_a0, const STRLEN goperl_a1, const STRLEN goperl_a2, const char * const goperl_a3, const STRLEN goperl_a4) { goperl_long_sv_insert(goperl_a0, goperl_a1, goperl_a2, goperl_a3, goperl_a4); }
#if defined(sv_insert_flags) && !defined(Perl_sv_insert_flags)
#define Perl_sv_insert_flags(...) GOPERL_LONG_CALL(sv_insert_flags, GOPERL_THX_DROP(__VA_ARGS__))
#endif
static inline __attribute__((always_inline, unused)) IO *
goperl_long_sv_2io(SV * const goperl_a0) { return sv_2io(goperl_a0); }
extern inline __attribute__((gnu_inline, always_inline, unused)) IO *
Perl_sv_2io(pTHX_ SV * const goperl_a0) { return goperl_long_sv_2io(goperl_a0); }
static inline __attribute__((always_inline, unused)) int
goperl_long_sv_isa(SV * goperl_a0, const char * const goperl_a1) { return sv_isa(goperl_a0, goperl_a1); }
extern inline __attribute__((gnu_inline, always_inline, unused)) int
Perl_sv_isa(pTHX_ SV * goperl_a0, const char * const goperl_a1) { return goperl_long_sv_isa(goperl_a0, goperl_a1); }
static inline __attribute__((always_inline, unused)) bool
goperl_long_sv_isa_sv(SV * goperl_a0, SV * goperl_a1) { return sv_isa_sv(goperl_a0, goperl_a1); }
extern inline __attribute__((gnu_inline, always_inline, unused)) bool
Perl_sv_isa_sv(pTHX_ SV * goperl_a0, SV * goperl_a1) { return goperl_long_sv_isa_sv(goperl_a0, goperl_a1); }
static inline __attribute__((always_inline, unused)) int
goperl_long_sv_isobject(SV * goperl_a0) { return sv_isobject(goperl_a0); }
extern inline __attribute__((gnu_inline, always_inline, unused)) int
Perl_sv_isobject(pTHX_ SV * goperl_a0) { return goperl_long_sv_isobject(goperl_a0); }
static inline __attribute__((always_inline, unused)) IV
goperl_long_SvIV(SV * goperl_a0) { return SvIV(goperl_a0); }
extern inline __attribute__((gnu_inline, always_inline, unused)) IV
Perl_SvIV(pTHX_ SV * goperl_a0) { return goperl_long_SvIV(goperl_a0); }
static inline __attribute__((always_inline, unused)) IV
goperl_long_sv_2iv_flags(SV * const goperl_a0, const I32 goperl_a1) { return sv_2iv_flags(goperl_a0, goperl_a1); }
extern inline __attribute__((gnu_inline, always_inline, unused)) IV
Perl_sv_2iv_flags(pTHX_ SV * const goperl_a0, const I32 goperl_a1) { return goperl_long_sv_2iv_flags(goperl_a0, goperl_a1); }
static inline __attribute__((always_inline, unused)) IV
goperl_long_SvIV_nomg(SV * goperl_a0) { return SvIV_nomg(goperl_a0); }
extern inline __attribute__((gnu_inline, always_inline, unused)) IV
Perl_SvIV_nomg(pTHX_ SV * goperl_a0) { return goperl_long_SvIV_nomg(goperl_a0); }
#if defined(sv_langinfo) && !defined(Perl_sv_langinfo)
#define Perl_sv_langinfo(...) GOPERL_LONG_CALL(sv_langinfo, GOPERL_THX_DROP(__VA_ARGS__))
#endif
static inline __attribute__((always_inline, unused)) STRLEN
goperl_long_sv_len(SV * const goperl_a0) { return sv_len(goperl_a0); }
extern inline __attribute__((gnu_inline, always_inline, unused)) STRLEN
Perl_sv_len(pTHX_ SV * const goperl_a0) { return goperl_long_sv_len(goperl_a0); }
static inline __attribute__((always_inline, unused)) STRLEN
goperl_long_sv_len_utf8(SV * const goperl_a0) { return sv_len_utf8(goperl_a0); }
extern inline __attribute__((gnu_inline, always_inline, unused)) STRLEN
Perl_sv_len_utf8(pTHX_ SV * const goperl_a0) { return goperl_long_sv_len_utf8(goperl_a0); }
#if defined(sv_len_utf8_nomg) && !defined(Perl_sv_len_utf8_nomg)
#define Perl_sv_len_utf8_nomg(...) GOPERL_LONG_CALL(sv_len_utf8_nomg, GOPERL_THX_DROP(__VA_ARGS__))
#endif
static inline __attribute__((always_inline, unused)) void
goperl_long_sv_magic(SV * const goperl_a0, SV * const goperl_a1, const int goperl_a2, const char * const goperl_a3, const I32 goperl_a4) { sv_magic(goperl_a0, goperl_a1, goperl_a2, goperl_a3, goperl_a4); }
extern inline __attribute__((gnu_inline, always_inline, unused)) void
Perl_sv_magic(pTHX_ SV * const goperl_a0, SV * const goperl_a1, const int goperl_a2, const char * const goperl_a3, const I32 goperl_a4) { goperl_long_sv_magic(goperl_a0, goperl_a1, goperl_a2, goperl_a3, goperl_a4); }
static inline __attribute__((always_inline, unused)) MAGIC *
goperl_long_sv_magicext(SV * const goperl_a0, SV * const goperl_a1, const int goperl_a2, const MGVTBL * const goperl_a3, const char * const goperl_a4, const I32 goperl_a5) { return sv_magicext(goperl_a0, goperl_a1, goperl_a2, goperl_a3, goperl_a4, goperl_a5); }
extern inline __attribute__((gnu_inline, always_inline, unused)) MAGIC *
Perl_sv_magicext(pTHX_ SV * const goperl_a0, SV * const goperl_a1, const int goperl_a2, const MGVTBL * const goperl_a3, const char * const goperl_a4, const I32 goperl_a5) { return goperl_long_sv_magicext(goperl_a0, goperl_a1, goperl_a2, goperl_a3, goperl_a4, goperl_a5); }
static inline __attribute__((always_inline, unused)) SV *
goperl_long_sv_2mortal(SV * const goperl_a0) { return sv_2mortal(goperl_a0); }
extern inline __attribute__((gnu_inline, always_inline, unused)) SV *
Perl_sv_2mortal(pTHX_ SV * const goperl_a0) { return goperl_long_sv_2mortal(goperl_a0); }
static inline __attribute__((always_inline, unused)) SV *
goperl_long_sv_mortalcopy(SV * const goperl_a0) { return sv_mortalcopy(goperl_a0); }
extern inline __attribute__((gnu_inline, always_inline, unused)) SV *
Perl_sv_mortalcopy(pTHX_ SV * const goperl_a0) { return goperl_long_sv_mortalcopy(goperl_a0); }
#if defined(sv_mortalcopy_flags) && !defined(Perl_sv_mortalcopy_flags)
#define Perl_sv_mortalcopy_flags(...) GOPERL_LONG_CALL(sv_mortalcopy_flags, GOPERL_THX_DROP(__VA_ARGS__))
#endif
static inline __attribute__((always_inline, unused)) SV *
goperl_long_sv_newmortal(void) { return sv_newmortal(); }
extern inline __attribute__((gnu_inline, always_inline, unused)) SV *
Perl_sv_newmortal(pTHX) { return goperl_long_sv_newmortal(); }
#if defined(sv_nosharing) && !defined(Perl_sv_nosharing)
#define Perl_sv_nosharing(...) GOPERL_LONG_CALL(sv_nosharing, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(sv_numcmp) && !defined(Perl_sv_numcmp)
#define Perl_sv_numcmp(...) GOPERL_LONG_CALL(sv_numcmp, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(sv_numcmp_flags) && !defined(Perl_sv_numcmp_flags)
#define Perl_sv_numcmp_flags(...) GOPERL_LONG_CALL(sv_numcmp_flags, GOPERL_THX_DROP(__VA_ARGS__))
#endif
static inline __attribute__((always_inline, unused)) bool
goperl_long_sv_numeq(SV * goperl_a0, SV * goperl_a1) { return sv_numeq(goperl_a0, goperl_a1); }
extern inline __attribute__((gnu_inline, always_inline, unused)) bool
Perl_sv_numeq(pTHX_ SV * goperl_a0, SV * goperl_a1) { return goperl_long_sv_numeq(goperl_a0, goperl_a1); }
static inline __attribute__((always_inline, unused)) bool
goperl_long_sv_numeq_flags(SV * goperl_a0, SV * goperl_a1, const U32 goperl_a2) { return sv_numeq_flags(goperl_a0, goperl_a1, goperl_a2); }
extern inline __attribute__((gnu_inline, always_inline, unused)) bool
Perl_sv_numeq_flags(pTHX_ SV * goperl_a0, SV * goperl_a1, const U32 goperl_a2) { return goperl_long_sv_numeq_flags(goperl_a0, goperl_a1, goperl_a2); }
#if defined(sv_numge) && !defined(Perl_sv_numge)
#define Perl_sv_numge(...) GOPERL_LONG_CALL(sv_numge, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(sv_numge_flags) && !defined(Perl_sv_numge_flags)
#define Perl_sv_numge_flags(...) GOPERL_LONG_CALL(sv_numge_flags, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(sv_numgt) && !defined(Perl_sv_numgt)
#define Perl_sv_numgt(...) GOPERL_LONG_CALL(sv_numgt, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(sv_numgt_flags) && !defined(Perl_sv_numgt_flags)
#define Perl_sv_numgt_flags(...) GOPERL_LONG_CALL(sv_numgt_flags, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(sv_numle) && !defined(Perl_sv_numle)
#define Perl_sv_numle(...) GOPERL_LONG_CALL(sv_numle, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(sv_numle_flags) && !defined(Perl_sv_numle_flags)
#define Perl_sv_numle_flags(...) GOPERL_LONG_CALL(sv_numle_flags, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(sv_numlt) && !defined(Perl_sv_numlt)
#define Perl_sv_numlt(...) GOPERL_LONG_CALL(sv_numlt, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(sv_numlt_flags) && !defined(Perl_sv_numlt_flags)
#define Perl_sv_numlt_flags(...) GOPERL_LONG_CALL(sv_numlt_flags, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(sv_numne) && !defined(Perl_sv_numne)
#define Perl_sv_numne(...) GOPERL_LONG_CALL(sv_numne, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(sv_numne_flags) && !defined(Perl_sv_numne_flags)
#define Perl_sv_numne_flags(...) GOPERL_LONG_CALL(sv_numne_flags, GOPERL_THX_DROP(__VA_ARGS__))
#endif
static inline __attribute__((always_inline, unused)) NV
goperl_long_SvNV(SV * goperl_a0) { return SvNV(goperl_a0); }
extern inline __attribute__((gnu_inline, always_inline, unused)) NV
Perl_SvNV(pTHX_ SV * goperl_a0) { return goperl_long_SvNV(goperl_a0); }
static inline __attribute__((always_inline, unused)) NV
goperl_long_sv_2nv_flags(SV * const goperl_a0, const I32 goperl_a1) { return sv_2nv_flags(goperl_a0, goperl_a1); }
extern inline __attribute__((gnu_inline, always_inline, unused)) NV
Perl_sv_2nv_flags(pTHX_ SV * const goperl_a0, const I32 goperl_a1) { return goperl_long_sv_2nv_flags(goperl_a0, goperl_a1); }
static inline __attribute__((always_inline, unused)) NV
goperl_long_SvNV_nomg(SV * goperl_a0) { return SvNV_nomg(goperl_a0); }
extern inline __attribute__((gnu_inline, always_inline, unused)) NV
Perl_SvNV_nomg(pTHX_ SV * goperl_a0) { return goperl_long_SvNV_nomg(goperl_a0); }
static inline __attribute__((always_inline, unused)) void
goperl_long_sv_pos_b2u(SV * const goperl_a0, I32 * const goperl_a1) { sv_pos_b2u(goperl_a0, goperl_a1); }
extern inline __attribute__((gnu_inline, always_inline, unused)) void
Perl_sv_pos_b2u(pTHX_ SV * const goperl_a0, I32 * const goperl_a1) { goperl_long_sv_pos_b2u(goperl_a0, goperl_a1); }
static inline __attribute__((always_inline, unused)) STRLEN
goperl_long_sv_pos_b2u_flags(SV * const goperl_a0, STRLEN const goperl_a1, U32 goperl_a2) { return sv_pos_b2u_flags(goperl_a0, goperl_a1, goperl_a2); }
extern inline __attribute__((gnu_inline, always_inline, unused)) STRLEN
Perl_sv_pos_b2u_flags(pTHX_ SV * const goperl_a0, STRLEN const goperl_a1, U32 goperl_a2) { return goperl_long_sv_pos_b2u_flags(goperl_a0, goperl_a1, goperl_a2); }
static inline __attribute__((always_inline, unused)) void
goperl_long_sv_pos_u2b(SV * const goperl_a0, I32 * const goperl_a1, I32 * const goperl_a2) { sv_pos_u2b(goperl_a0, goperl_a1, goperl_a2); }
extern inline __attribute__((gnu_inline, always_inline, unused)) void
Perl_sv_pos_u2b(pTHX_ SV * const goperl_a0, I32 * const goperl_a1, I32 * const goperl_a2) { goperl_long_sv_pos_u2b(goperl_a0, goperl_a1, goperl_a2); }
static inline __attribute__((always_inline, unused)) STRLEN
goperl_long_sv_pos_u2b_flags(SV * const goperl_a0, STRLEN goperl_a1, STRLEN * const goperl_a2, U32 goperl_a3) { return sv_pos_u2b_flags(goperl_a0, goperl_a1, goperl_a2, goperl_a3); }
extern inline __attribute__((gnu_inline, always_inline, unused)) STRLEN
Perl_sv_pos_u2b_flags(pTHX_ SV * const goperl_a0, STRLEN goperl_a1, STRLEN * const goperl_a2, U32 goperl_a3) { return goperl_long_sv_pos_u2b_flags(goperl_a0, goperl_a1, goperl_a2, goperl_a3); }
static inline __attribute__((always_inline, unused)) char *
goperl_long_sv_2pv(SV * goperl_a0, STRLEN * goperl_a1) { return sv_2pv(goperl_a0, goperl_a1); }
extern inline __attribute__((gnu_inline, always_inline, unused)) char *
Perl_sv_2pv(pTHX_ SV * goperl_a0, STRLEN * goperl_a1) { return goperl_long_sv_2pv(goperl_a0, goperl_a1); }
static inline __attribute__((always_inline, unused)) char *
goperl_long_sv_2pvbyte(SV * goperl_a0, STRLEN * const goperl_a1) { return sv_2pvbyte(goperl_a0, goperl_a1); }
extern inline __attribute__((gnu_inline, always_inline, unused)) char *
Perl_sv_2pvbyte(pTHX_ SV * goperl_a0, STRLEN * const goperl_a1) { return goperl_long_sv_2pvbyte(goperl_a0, goperl_a1); }
static inline __attribute__((always_inline, unused)) char *
goperl_long_sv_2pvbyte_flags(SV * goperl_a0, STRLEN * const goperl_a1, const U32 goperl_a2) { return sv_2pvbyte_flags(goperl_a0, goperl_a1, goperl_a2); }
extern inline __attribute__((gnu_inline, always_inline, unused)) char *
Perl_sv_2pvbyte_flags(pTHX_ SV * goperl_a0, STRLEN * const goperl_a1, const U32 goperl_a2) { return goperl_long_sv_2pvbyte_flags(goperl_a0, goperl_a1, goperl_a2); }
static inline __attribute__((always_inline, unused)) char *
goperl_long_sv_2pv_flags(SV * const goperl_a0, STRLEN * const goperl_a1, const U32 goperl_a2) { return sv_2pv_flags(goperl_a0, goperl_a1, goperl_a2); }
extern inline __attribute__((gnu_inline, always_inline, unused)) char *
Perl_sv_2pv_flags(pTHX_ SV * const goperl_a0, STRLEN * const goperl_a1, const U32 goperl_a2) { return goperl_long_sv_2pv_flags(goperl_a0, goperl_a1, goperl_a2); }
static inline __attribute__((always_inline, unused)) char *
goperl_long_sv_pvn_force_flags(SV * const goperl_a0, STRLEN * const goperl_a1, const U32 goperl_a2) { return sv_pvn_force_flags(goperl_a0, goperl_a1, goperl_a2); }
extern inline __attribute__((gnu_inline, always_inline, unused)) char *
Perl_sv_pvn_force_flags(pTHX_ SV * const goperl_a0, STRLEN * const goperl_a1, const U32 goperl_a2) { return goperl_long_sv_pvn_force_flags(goperl_a0, goperl_a1, goperl_a2); }
static inline __attribute__((always_inline, unused)) char *
goperl_long_sv_2pvutf8(SV * goperl_a0, STRLEN * const goperl_a1) { return sv_2pvutf8(goperl_a0, goperl_a1); }
extern inline __attribute__((gnu_inline, always_inline, unused)) char *
Perl_sv_2pvutf8(pTHX_ SV * goperl_a0, STRLEN * const goperl_a1) { return goperl_long_sv_2pvutf8(goperl_a0, goperl_a1); }
static inline __attribute__((always_inline, unused)) char *
goperl_long_sv_2pvutf8_flags(SV * goperl_a0, STRLEN * const goperl_a1, const U32 goperl_a2) { return sv_2pvutf8_flags(goperl_a0, goperl_a1, goperl_a2); }
extern inline __attribute__((gnu_inline, always_inline, unused)) char *
Perl_sv_2pvutf8_flags(pTHX_ SV * goperl_a0, STRLEN * const goperl_a1, const U32 goperl_a2) { return goperl_long_sv_2pvutf8_flags(goperl_a0, goperl_a1, goperl_a2); }
#if defined(SvPVXtrue) && !defined(Perl_SvPVXtrue)
#define Perl_SvPVXtrue(...) GOPERL_LONG_CALL(SvPVXtrue, GOPERL_THX_DROP(__VA_ARGS__))
#endif
static inline __attribute__((always_inline, unused)) char *
goperl_long_sv_recode_to_utf8(SV * goperl_a0, SV * goperl_a1) { return sv_recode_to_utf8(goperl_a0, goperl_a1); }
extern inline __attribute__((gnu_inline, always_inline, unused)) char *
Perl_sv_recode_to_utf8(pTHX_ SV * goperl_a0, SV * goperl_a1) { return goperl_long_sv_recode_to_utf8(goperl_a0, goperl_a1); }
static inline __attribute__((always_inline, unused)) SV *
goperl_long_sv_ref(SV * goperl_a0, const SV * const goperl_a1, const int goperl_a2) { return sv_ref(goperl_a0, goperl_a1, goperl_a2); }
extern inline __attribute__((gnu_inline, always_inline, unused)) SV *
Perl_sv_ref(pTHX_ SV * goperl_a0, const SV * const goperl_a1, const int goperl_a2) { return goperl_long_sv_ref(goperl_a0, goperl_a1, goperl_a2); }
static inline __attribute__((always_inline, unused)) void
goperl_long_SvREFCNT_dec(SV * goperl_a0) { SvREFCNT_dec(goperl_a0); }
extern inline __attribute__((gnu_inline, always_inline, unused)) void
Perl_SvREFCNT_dec(pTHX_ SV * goperl_a0) { goperl_long_SvREFCNT_dec(goperl_a0); }
static inline __attribute__((always_inline, unused)) void
goperl_long_SvREFCNT_dec_NN(SV * goperl_a0) { SvREFCNT_dec_NN(goperl_a0); }
extern inline __attribute__((gnu_inline, always_inline, unused)) void
Perl_SvREFCNT_dec_NN(pTHX_ SV * goperl_a0) { goperl_long_SvREFCNT_dec_NN(goperl_a0); }
#if defined(SvREFCNT_dec_ret_NULL) && !defined(Perl_SvREFCNT_dec_ret_NULL)
#define Perl_SvREFCNT_dec_ret_NULL(...) GOPERL_LONG_CALL(SvREFCNT_dec_ret_NULL, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(SvREFCNT_dec_set_NULL) && !defined(Perl_SvREFCNT_dec_set_NULL)
#define Perl_SvREFCNT_dec_set_NULL(...) GOPERL_LONG_CALL(SvREFCNT_dec_set_NULL, GOPERL_THX_DROP(__VA_ARGS__))
#endif
static inline __attribute__((always_inline, unused)) SV *
goperl_long_SvREFCNT_inc(SV * goperl_a0) { return SvREFCNT_inc(goperl_a0); }
extern inline __attribute__((gnu_inline, always_inline, unused)) SV *
Perl_SvREFCNT_inc(SV * goperl_a0) { return goperl_long_SvREFCNT_inc(goperl_a0); }
static inline __attribute__((always_inline, unused)) SV *
goperl_long_SvREFCNT_inc_NN(SV * goperl_a0) { return SvREFCNT_inc_NN(goperl_a0); }
extern inline __attribute__((gnu_inline, always_inline, unused)) SV *
Perl_SvREFCNT_inc_NN(SV * goperl_a0) { return goperl_long_SvREFCNT_inc_NN(goperl_a0); }
static inline __attribute__((always_inline, unused)) void
goperl_long_SvREFCNT_inc_void(SV * goperl_a0) { SvREFCNT_inc_void(goperl_a0); }
extern inline __attribute__((gnu_inline, always_inline, unused)) void
Perl_SvREFCNT_inc_void(SV * goperl_a0) { goperl_long_SvREFCNT_inc_void(goperl_a0); }
static inline __attribute__((always_inline, unused)) const char *
goperl_long_sv_reftype(const SV * const goperl_a0, const int goperl_a1) { return sv_reftype(goperl_a0, goperl_a1); }
extern inline __attribute__((gnu_inline, always_inline, unused)) const char *
Perl_sv_reftype(pTHX_ const SV * const goperl_a0, const int goperl_a1) { return goperl_long_sv_reftype(goperl_a0, goperl_a1); }
#if defined(sv_regex_global_pos_clear) && !defined(Perl_sv_regex_global_pos_clear)
#define Perl_sv_regex_global_pos_clear(...) GOPERL_LONG_CALL(sv_regex_global_pos_clear, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(sv_regex_global_pos_get) && !defined(Perl_sv_regex_global_pos_get)
#define Perl_sv_regex_global_pos_get(...) GOPERL_LONG_CALL(sv_regex_global_pos_get, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(sv_regex_global_pos_set) && !defined(Perl_sv_regex_global_pos_set)
#define Perl_sv_regex_global_pos_set(...) GOPERL_LONG_CALL(sv_regex_global_pos_set, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(sv_replace) && !defined(Perl_sv_replace)
#define Perl_sv_replace(...) GOPERL_LONG_CALL(sv_replace, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(sv_report_used) && !defined(Perl_sv_report_used)
#define Perl_sv_report_used(...) GOPERL_LONG_CALL(sv_report_used, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(sv_reset) && !defined(Perl_sv_reset)
#define Perl_sv_reset(...) GOPERL_LONG_CALL(sv_reset, GOPERL_THX_DROP(__VA_ARGS__))
#endif
static inline __attribute__((always_inline, unused)) SV *
goperl_long_sv_rvunweaken(SV * const goperl_a0) { return sv_rvunweaken(goperl_a0); }
extern inline __attribute__((gnu_inline, always_inline, unused)) SV *
Perl_sv_rvunweaken(pTHX_ SV * const goperl_a0) { return goperl_long_sv_rvunweaken(goperl_a0); }
static inline __attribute__((always_inline, unused)) SV *
goperl_long_sv_rvweaken(SV * const goperl_a0) { return sv_rvweaken(goperl_a0); }
extern inline __attribute__((gnu_inline, always_inline, unused)) SV *
Perl_sv_rvweaken(pTHX_ SV * const goperl_a0) { return goperl_long_sv_rvweaken(goperl_a0); }
#if defined(sv_set_bool) && !defined(Perl_sv_set_bool)
#define Perl_sv_set_bool(...) GOPERL_LONG_CALL(sv_set_bool, GOPERL_THX_DROP(__VA_ARGS__))
#endif
static inline __attribute__((always_inline, unused)) void
goperl_long_sv_set_false(SV * goperl_a0) { sv_set_false(goperl_a0); }
extern inline __attribute__((gnu_inline, always_inline, unused)) void
Perl_sv_set_false(pTHX_ SV * goperl_a0) { goperl_long_sv_set_false(goperl_a0); }
static inline __attribute__((always_inline, unused)) void
goperl_long_sv_setiv(SV * const goperl_a0, const IV goperl_a1) { sv_setiv(goperl_a0, goperl_a1); }
extern inline __attribute__((gnu_inline, always_inline, unused)) void
Perl_sv_setiv(pTHX_ SV * const goperl_a0, const IV goperl_a1) { goperl_long_sv_setiv(goperl_a0, goperl_a1); }
static inline __attribute__((always_inline, unused)) void
goperl_long_sv_setiv_mg(SV * const goperl_a0, const IV goperl_a1) { sv_setiv_mg(goperl_a0, goperl_a1); }
extern inline __attribute__((gnu_inline, always_inline, unused)) void
Perl_sv_setiv_mg(pTHX_ SV * const goperl_a0, const IV goperl_a1) { goperl_long_sv_setiv_mg(goperl_a0, goperl_a1); }
static inline __attribute__((always_inline, unused)) void
goperl_long_sv_setnv(SV * const goperl_a0, const NV goperl_a1) { sv_setnv(goperl_a0, goperl_a1); }
extern inline __attribute__((gnu_inline, always_inline, unused)) void
Perl_sv_setnv(pTHX_ SV * const goperl_a0, const NV goperl_a1) { goperl_long_sv_setnv(goperl_a0, goperl_a1); }
static inline __attribute__((always_inline, unused)) void
goperl_long_sv_setnv_mg(SV * const goperl_a0, const NV goperl_a1) { sv_setnv_mg(goperl_a0, goperl_a1); }
extern inline __attribute__((gnu_inline, always_inline, unused)) void
Perl_sv_setnv_mg(pTHX_ SV * const goperl_a0, const NV goperl_a1) { goperl_long_sv_setnv_mg(goperl_a0, goperl_a1); }
static inline __attribute__((always_inline, unused)) void
goperl_long_sv_setpv(SV * const goperl_a0, const char * const goperl_a1) { sv_setpv(goperl_a0, goperl_a1); }
extern inline __attribute__((gnu_inline, always_inline, unused)) void
Perl_sv_setpv(pTHX_ SV * const goperl_a0, const char * const goperl_a1) { goperl_long_sv_setpv(goperl_a0, goperl_a1); }
static inline __attribute__((always_inline, unused)) char  *
goperl_long_sv_setpv_bufsize(SV * const goperl_a0, const STRLEN goperl_a1, const STRLEN goperl_a2) { return sv_setpv_bufsize(goperl_a0, goperl_a1, goperl_a2); }
extern inline __attribute__((gnu_inline, always_inline, unused)) char  *
Perl_sv_setpv_bufsize(pTHX_ SV * const goperl_a0, const STRLEN goperl_a1, const STRLEN goperl_a2) { return goperl_long_sv_setpv_bufsize(goperl_a0, goperl_a1, goperl_a2); }
#if defined(sv_setpvf_mg) && !defined(Perl_sv_setpvf_mg)
#define Perl_sv_setpvf_mg(...) GOPERL_LONG_CALL(sv_setpvf_mg, GOPERL_THX_DROP(__VA_ARGS__))
#endif
static inline __attribute__((always_inline, unused)) void
goperl_long_sv_setpv_mg(SV * const goperl_a0, const char * const goperl_a1) { sv_setpv_mg(goperl_a0, goperl_a1); }
extern inline __attribute__((gnu_inline, always_inline, unused)) void
Perl_sv_setpv_mg(pTHX_ SV * const goperl_a0, const char * const goperl_a1) { goperl_long_sv_setpv_mg(goperl_a0, goperl_a1); }
static inline __attribute__((always_inline, unused)) void
goperl_long_sv_setpvn(SV * const goperl_a0, const char * const goperl_a1, const STRLEN goperl_a2) { sv_setpvn(goperl_a0, goperl_a1, goperl_a2); }
extern inline __attribute__((gnu_inline, always_inline, unused)) void
Perl_sv_setpvn(pTHX_ SV * const goperl_a0, const char * const goperl_a1, const STRLEN goperl_a2) { goperl_long_sv_setpvn(goperl_a0, goperl_a1, goperl_a2); }
static inline __attribute__((always_inline, unused)) void
goperl_long_sv_setpvn_fresh(SV * const goperl_a0, const char * const goperl_a1, const STRLEN goperl_a2) { sv_setpvn_fresh(goperl_a0, goperl_a1, goperl_a2); }
extern inline __attribute__((gnu_inline, always_inline, unused)) void
Perl_sv_setpvn_fresh(pTHX_ SV * const goperl_a0, const char * const goperl_a1, const STRLEN goperl_a2) { goperl_long_sv_setpvn_fresh(goperl_a0, goperl_a1, goperl_a2); }
static inline __attribute__((always_inline, unused)) void
goperl_long_sv_setpvn_mg(SV * const goperl_a0, const char * const goperl_a1, const STRLEN goperl_a2) { sv_setpvn_mg(goperl_a0, goperl_a1, goperl_a2); }
extern inline __attribute__((gnu_inline, always_inline, unused)) void
Perl_sv_setpvn_mg(pTHX_ SV * const goperl_a0, const char * const goperl_a1, const STRLEN goperl_a2) { goperl_long_sv_setpvn_mg(goperl_a0, goperl_a1, goperl_a2); }
static inline __attribute__((always_inline, unused)) SV *
goperl_long_sv_setref_iv(SV * const goperl_a0, const char * const goperl_a1, const IV goperl_a2) { return sv_setref_iv(goperl_a0, goperl_a1, goperl_a2); }
extern inline __attribute__((gnu_inline, always_inline, unused)) SV *
Perl_sv_setref_iv(pTHX_ SV * const goperl_a0, const char * const goperl_a1, const IV goperl_a2) { return goperl_long_sv_setref_iv(goperl_a0, goperl_a1, goperl_a2); }
static inline __attribute__((always_inline, unused)) SV *
goperl_long_sv_setref_nv(SV * const goperl_a0, const char * const goperl_a1, const NV goperl_a2) { return sv_setref_nv(goperl_a0, goperl_a1, goperl_a2); }
extern inline __attribute__((gnu_inline, always_inline, unused)) SV *
Perl_sv_setref_nv(pTHX_ SV * const goperl_a0, const char * const goperl_a1, const NV goperl_a2) { return goperl_long_sv_setref_nv(goperl_a0, goperl_a1, goperl_a2); }
static inline __attribute__((always_inline, unused)) SV *
goperl_long_sv_setref_pv(SV * const goperl_a0, const char * const goperl_a1, void * const goperl_a2) { return sv_setref_pv(goperl_a0, goperl_a1, goperl_a2); }
extern inline __attribute__((gnu_inline, always_inline, unused)) SV *
Perl_sv_setref_pv(pTHX_ SV * const goperl_a0, const char * const goperl_a1, void * const goperl_a2) { return goperl_long_sv_setref_pv(goperl_a0, goperl_a1, goperl_a2); }
static inline __attribute__((always_inline, unused)) SV *
goperl_long_sv_setref_pvn(SV * const goperl_a0, const char * const goperl_a1, const char * const goperl_a2, const STRLEN goperl_a3) { return sv_setref_pvn(goperl_a0, goperl_a1, goperl_a2, goperl_a3); }
extern inline __attribute__((gnu_inline, always_inline, unused)) SV *
Perl_sv_setref_pvn(pTHX_ SV * const goperl_a0, const char * const goperl_a1, const char * const goperl_a2, const STRLEN goperl_a3) { return goperl_long_sv_setref_pvn(goperl_a0, goperl_a1, goperl_a2, goperl_a3); }
static inline __attribute__((always_inline, unused)) SV *
goperl_long_sv_setref_uv(SV * const goperl_a0, const char * const goperl_a1, const UV goperl_a2) { return sv_setref_uv(goperl_a0, goperl_a1, goperl_a2); }
extern inline __attribute__((gnu_inline, always_inline, unused)) SV *
Perl_sv_setref_uv(pTHX_ SV * const goperl_a0, const char * const goperl_a1, const UV goperl_a2) { return goperl_long_sv_setref_uv(goperl_a0, goperl_a1, goperl_a2); }
static inline __attribute__((always_inline, unused)) void
goperl_long_sv_setrv_inc(SV * const goperl_a0, SV * const goperl_a1) { sv_setrv_inc(goperl_a0, goperl_a1); }
extern inline __attribute__((gnu_inline, always_inline, unused)) void
Perl_sv_setrv_inc(pTHX_ SV * const goperl_a0, SV * const goperl_a1) { goperl_long_sv_setrv_inc(goperl_a0, goperl_a1); }
static inline __attribute__((always_inline, unused)) void
goperl_long_sv_setrv_inc_mg(SV * const goperl_a0, SV * const goperl_a1) { sv_setrv_inc_mg(goperl_a0, goperl_a1); }
extern inline __attribute__((gnu_inline, always_inline, unused)) void
Perl_sv_setrv_inc_mg(pTHX_ SV * const goperl_a0, SV * const goperl_a1) { goperl_long_sv_setrv_inc_mg(goperl_a0, goperl_a1); }
static inline __attribute__((always_inline, unused)) void
goperl_long_sv_setrv_noinc(SV * const goperl_a0, SV * const goperl_a1) { sv_setrv_noinc(goperl_a0, goperl_a1); }
extern inline __attribute__((gnu_inline, always_inline, unused)) void
Perl_sv_setrv_noinc(pTHX_ SV * const goperl_a0, SV * const goperl_a1) { goperl_long_sv_setrv_noinc(goperl_a0, goperl_a1); }
static inline __attribute__((always_inline, unused)) void
goperl_long_sv_setrv_noinc_mg(SV * const goperl_a0, SV * const goperl_a1) { sv_setrv_noinc_mg(goperl_a0, goperl_a1); }
extern inline __attribute__((gnu_inline, always_inline, unused)) void
Perl_sv_setrv_noinc_mg(pTHX_ SV * const goperl_a0, SV * const goperl_a1) { goperl_long_sv_setrv_noinc_mg(goperl_a0, goperl_a1); }
static inline __attribute__((always_inline, unused)) void
goperl_long_sv_setsv(SV * goperl_a0, SV * goperl_a1) { sv_setsv(goperl_a0, goperl_a1); }
extern inline __attribute__((gnu_inline, always_inline, unused)) void
Perl_sv_setsv(pTHX_ SV * goperl_a0, SV * goperl_a1) { goperl_long_sv_setsv(goperl_a0, goperl_a1); }
static inline __attribute__((always_inline, unused)) void
goperl_long_sv_setsv_flags(SV * goperl_a0, SV * goperl_a1, const I32 goperl_a2) { sv_setsv_flags(goperl_a0, goperl_a1, goperl_a2); }
extern inline __attribute__((gnu_inline, always_inline, unused)) void
Perl_sv_setsv_flags(pTHX_ SV * goperl_a0, SV * goperl_a1, const I32 goperl_a2) { goperl_long_sv_setsv_flags(goperl_a0, goperl_a1, goperl_a2); }
static inline __attribute__((always_inline, unused)) void
goperl_long_sv_setsv_mg(SV * const goperl_a0, SV * const goperl_a1) { sv_setsv_mg(goperl_a0, goperl_a1); }
extern inline __attribute__((gnu_inline, always_inline, unused)) void
Perl_sv_setsv_mg(pTHX_ SV * const goperl_a0, SV * const goperl_a1) { goperl_long_sv_setsv_mg(goperl_a0, goperl_a1); }
static inline __attribute__((always_inline, unused)) void
goperl_long_sv_set_true(SV * goperl_a0) { sv_set_true(goperl_a0); }
extern inline __attribute__((gnu_inline, always_inline, unused)) void
Perl_sv_set_true(pTHX_ SV * goperl_a0) { goperl_long_sv_set_true(goperl_a0); }
static inline __attribute__((always_inline, unused)) void
goperl_long_sv_set_undef(SV * goperl_a0) { sv_set_undef(goperl_a0); }
extern inline __attribute__((gnu_inline, always_inline, unused)) void
Perl_sv_set_undef(pTHX_ SV * goperl_a0) { goperl_long_sv_set_undef(goperl_a0); }
static inline __attribute__((always_inline, unused)) void
goperl_long_sv_setuv(SV * const goperl_a0, const UV goperl_a1) { sv_setuv(goperl_a0, goperl_a1); }
extern inline __attribute__((gnu_inline, always_inline, unused)) void
Perl_sv_setuv(pTHX_ SV * const goperl_a0, const UV goperl_a1) { goperl_long_sv_setuv(goperl_a0, goperl_a1); }
static inline __attribute__((always_inline, unused)) void
goperl_long_sv_setuv_mg(SV * const goperl_a0, const UV goperl_a1) { sv_setuv_mg(goperl_a0, goperl_a1); }
extern inline __attribute__((gnu_inline, always_inline, unused)) void
Perl_sv_setuv_mg(pTHX_ SV * const goperl_a0, const UV goperl_a1) { goperl_long_sv_setuv_mg(goperl_a0, goperl_a1); }
static inline __attribute__((always_inline, unused)) bool
goperl_long_sv_streq(SV * goperl_a0, SV * goperl_a1) { return sv_streq(goperl_a0, goperl_a1); }
extern inline __attribute__((gnu_inline, always_inline, unused)) bool
Perl_sv_streq(pTHX_ SV * goperl_a0, SV * goperl_a1) { return goperl_long_sv_streq(goperl_a0, goperl_a1); }
static inline __attribute__((always_inline, unused)) bool
goperl_long_sv_streq_flags(SV * goperl_a0, SV * goperl_a1, const U32 goperl_a2) { return sv_streq_flags(goperl_a0, goperl_a1, goperl_a2); }
extern inline __attribute__((gnu_inline, always_inline, unused)) bool
Perl_sv_streq_flags(pTHX_ SV * goperl_a0, SV * goperl_a1, const U32 goperl_a2) { return goperl_long_sv_streq_flags(goperl_a0, goperl_a1, goperl_a2); }
#if defined(sv_strftime_ints) && !defined(Perl_sv_strftime_ints)
#define Perl_sv_strftime_ints(...) GOPERL_LONG_CALL(sv_strftime_ints, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(sv_strftime_tm) && !defined(Perl_sv_strftime_tm)
#define Perl_sv_strftime_tm(...) GOPERL_LONG_CALL(sv_strftime_tm, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(sv_string_from_errnum) && !defined(Perl_sv_string_from_errnum)
#define Perl_sv_string_from_errnum(...) GOPERL_LONG_CALL(sv_string_from_errnum, GOPERL_THX_DROP(__VA_ARGS__))
#endif
static inline __attribute__((always_inline, unused)) bool
goperl_long_SvTRUE(SV * goperl_a0) { return SvTRUE(goperl_a0); }
extern inline __attribute__((gnu_inline, always_inline, unused)) bool
Perl_SvTRUE(pTHX_ SV * goperl_a0) { return goperl_long_SvTRUE(goperl_a0); }
static inline __attribute__((always_inline, unused)) bool
goperl_long_SvTRUE_NN(SV * goperl_a0) { return SvTRUE_NN(goperl_a0); }
extern inline __attribute__((gnu_inline, always_inline, unused)) bool
Perl_SvTRUE_NN(pTHX_ SV * goperl_a0) { return goperl_long_SvTRUE_NN(goperl_a0); }
static inline __attribute__((always_inline, unused)) bool
goperl_long_SvTRUE_nomg(SV * goperl_a0) { return SvTRUE_nomg(goperl_a0); }
extern inline __attribute__((gnu_inline, always_inline, unused)) bool
Perl_SvTRUE_nomg(pTHX_ SV * goperl_a0) { return goperl_long_SvTRUE_nomg(goperl_a0); }
static inline __attribute__((always_inline, unused)) char *
goperl_long_sv_uni_display(SV * goperl_a0, SV * goperl_a1, STRLEN goperl_a2, UV goperl_a3) { return sv_uni_display(goperl_a0, goperl_a1, goperl_a2, goperl_a3); }
extern inline __attribute__((gnu_inline, always_inline, unused)) char *
Perl_sv_uni_display(pTHX_ SV * goperl_a0, SV * goperl_a1, STRLEN goperl_a2, UV goperl_a3) { return goperl_long_sv_uni_display(goperl_a0, goperl_a1, goperl_a2, goperl_a3); }
static inline __attribute__((always_inline, unused)) int
goperl_long_sv_unmagic(SV * const goperl_a0, const int goperl_a1) { return sv_unmagic(goperl_a0, goperl_a1); }
extern inline __attribute__((gnu_inline, always_inline, unused)) int
Perl_sv_unmagic(pTHX_ SV * const goperl_a0, const int goperl_a1) { return goperl_long_sv_unmagic(goperl_a0, goperl_a1); }
static inline __attribute__((always_inline, unused)) int
goperl_long_sv_unmagicext(SV * const goperl_a0, const int goperl_a1, const MGVTBL * goperl_a2) { return sv_unmagicext(goperl_a0, goperl_a1, goperl_a2); }
extern inline __attribute__((gnu_inline, always_inline, unused)) int
Perl_sv_unmagicext(pTHX_ SV * const goperl_a0, const int goperl_a1, const MGVTBL * goperl_a2) { return goperl_long_sv_unmagicext(goperl_a0, goperl_a1, goperl_a2); }
static inline __attribute__((always_inline, unused)) void
goperl_long_sv_unref(SV * goperl_a0) { sv_unref(goperl_a0); }
extern inline __attribute__((gnu_inline, always_inline, unused)) void
Perl_sv_unref(pTHX_ SV * goperl_a0) { goperl_long_sv_unref(goperl_a0); }
static inline __attribute__((always_inline, unused)) void
goperl_long_sv_unref_flags(SV * const goperl_a0, const U32 goperl_a1) { sv_unref_flags(goperl_a0, goperl_a1); }
extern inline __attribute__((gnu_inline, always_inline, unused)) void
Perl_sv_unref_flags(pTHX_ SV * const goperl_a0, const U32 goperl_a1) { goperl_long_sv_unref_flags(goperl_a0, goperl_a1); }
static inline __attribute__((always_inline, unused)) void
goperl_long_sv_upgrade(SV * const goperl_a0, svtype goperl_a1) { sv_upgrade(goperl_a0, goperl_a1); }
extern inline __attribute__((gnu_inline, always_inline, unused)) void
Perl_sv_upgrade(pTHX_ SV * const goperl_a0, svtype goperl_a1) { goperl_long_sv_upgrade(goperl_a0, goperl_a1); }
static inline __attribute__((always_inline, unused)) void
goperl_long_sv_usepvn(SV * goperl_a0, char * goperl_a1, STRLEN goperl_a2) { sv_usepvn(goperl_a0, goperl_a1, goperl_a2); }
extern inline __attribute__((gnu_inline, always_inline, unused)) void
Perl_sv_usepvn(pTHX_ SV * goperl_a0, char * goperl_a1, STRLEN goperl_a2) { goperl_long_sv_usepvn(goperl_a0, goperl_a1, goperl_a2); }
static inline __attribute__((always_inline, unused)) void
goperl_long_sv_usepvn_flags(SV * const goperl_a0, char * goperl_a1, const STRLEN goperl_a2, const U32 goperl_a3) { sv_usepvn_flags(goperl_a0, goperl_a1, goperl_a2, goperl_a3); }
extern inline __attribute__((gnu_inline, always_inline, unused)) void
Perl_sv_usepvn_flags(pTHX_ SV * const goperl_a0, char * goperl_a1, const STRLEN goperl_a2, const U32 goperl_a3) { goperl_long_sv_usepvn_flags(goperl_a0, goperl_a1, goperl_a2, goperl_a3); }
#if defined(sv_usepvn_mg) && !defined(Perl_sv_usepvn_mg)
#define Perl_sv_usepvn_mg(...) GOPERL_LONG_CALL(sv_usepvn_mg, GOPERL_THX_DROP(__VA_ARGS__))
#endif
static inline __attribute__((always_inline, unused)) bool
goperl_long_sv_utf8_decode(SV * const goperl_a0) { return sv_utf8_decode(goperl_a0); }
extern inline __attribute__((gnu_inline, always_inline, unused)) bool
Perl_sv_utf8_decode(pTHX_ SV * const goperl_a0) { return goperl_long_sv_utf8_decode(goperl_a0); }
static inline __attribute__((always_inline, unused)) bool
goperl_long_sv_utf8_downgrade(SV * const goperl_a0, const bool goperl_a1) { return sv_utf8_downgrade(goperl_a0, goperl_a1); }
extern inline __attribute__((gnu_inline, always_inline, unused)) bool
Perl_sv_utf8_downgrade(pTHX_ SV * const goperl_a0, const bool goperl_a1) { return goperl_long_sv_utf8_downgrade(goperl_a0, goperl_a1); }
#if defined(sv_utf8_downgrade_flags) && !defined(Perl_sv_utf8_downgrade_flags)
#define Perl_sv_utf8_downgrade_flags(...) GOPERL_LONG_CALL(sv_utf8_downgrade_flags, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(sv_utf8_downgrade_nomg) && !defined(Perl_sv_utf8_downgrade_nomg)
#define Perl_sv_utf8_downgrade_nomg(...) GOPERL_LONG_CALL(sv_utf8_downgrade_nomg, GOPERL_THX_DROP(__VA_ARGS__))
#endif
static inline __attribute__((always_inline, unused)) void
goperl_long_sv_utf8_encode(SV * const goperl_a0) { sv_utf8_encode(goperl_a0); }
extern inline __attribute__((gnu_inline, always_inline, unused)) void
Perl_sv_utf8_encode(pTHX_ SV * const goperl_a0) { goperl_long_sv_utf8_encode(goperl_a0); }
#if defined(sv_utf8_upgrade) && !defined(Perl_sv_utf8_upgrade)
#define Perl_sv_utf8_upgrade(...) GOPERL_LONG_CALL(sv_utf8_upgrade, GOPERL_THX_DROP(__VA_ARGS__))
#endif
static inline __attribute__((always_inline, unused)) STRLEN
goperl_long_sv_utf8_upgrade_flags(SV * const goperl_a0, const I32 goperl_a1) { return sv_utf8_upgrade_flags(goperl_a0, goperl_a1); }
extern inline __attribute__((gnu_inline, always_inline, unused)) STRLEN
Perl_sv_utf8_upgrade_flags(pTHX_ SV * const goperl_a0, const I32 goperl_a1) { return goperl_long_sv_utf8_upgrade_flags(goperl_a0, goperl_a1); }
#if defined(sv_utf8_upgrade_flags_grow) && !defined(Perl_sv_utf8_upgrade_flags_grow)
#define Perl_sv_utf8_upgrade_flags_grow(...) GOPERL_LONG_CALL(sv_utf8_upgrade_flags_grow, GOPERL_THX_DROP(__VA_ARGS__))
#endif
static inline __attribute__((always_inline, unused)) STRLEN
goperl_long_sv_utf8_upgrade_nomg(SV * goperl_a0) { return sv_utf8_upgrade_nomg(goperl_a0); }
extern inline __attribute__((gnu_inline, always_inline, unused)) STRLEN
Perl_sv_utf8_upgrade_nomg(pTHX_ SV * goperl_a0) { return goperl_long_sv_utf8_upgrade_nomg(goperl_a0); }
static inline __attribute__((always_inline, unused)) UV
goperl_long_SvUV(SV * goperl_a0) { return SvUV(goperl_a0); }
extern inline __attribute__((gnu_inline, always_inline, unused)) UV
Perl_SvUV(pTHX_ SV * goperl_a0) { return goperl_long_SvUV(goperl_a0); }
static inline __attribute__((always_inline, unused)) UV
goperl_long_sv_2uv_flags(SV * const goperl_a0, const I32 goperl_a1) { return sv_2uv_flags(goperl_a0, goperl_a1); }
extern inline __attribute__((gnu_inline, always_inline, unused)) UV
Perl_sv_2uv_flags(pTHX_ SV * const goperl_a0, const I32 goperl_a1) { return goperl_long_sv_2uv_flags(goperl_a0, goperl_a1); }
static inline __attribute__((always_inline, unused)) UV
goperl_long_SvUV_nomg(SV * goperl_a0) { return SvUV_nomg(goperl_a0); }
extern inline __attribute__((gnu_inline, always_inline, unused)) UV
Perl_SvUV_nomg(pTHX_ SV * goperl_a0) { return goperl_long_SvUV_nomg(goperl_a0); }
static inline __attribute__((always_inline, unused)) void
goperl_long_sv_vcatpvf(SV * const goperl_a0, const char * const goperl_a1, va_list * const goperl_a2) { sv_vcatpvf(goperl_a0, goperl_a1, goperl_a2); }
extern inline __attribute__((gnu_inline, always_inline, unused)) void
Perl_sv_vcatpvf(pTHX_ SV * const goperl_a0, const char * const goperl_a1, va_list * const goperl_a2) { goperl_long_sv_vcatpvf(goperl_a0, goperl_a1, goperl_a2); }
static inline __attribute__((always_inline, unused)) void
goperl_long_sv_vcatpvf_mg(SV * const goperl_a0, const char * const goperl_a1, va_list * const goperl_a2) { sv_vcatpvf_mg(goperl_a0, goperl_a1, goperl_a2); }
extern inline __attribute__((gnu_inline, always_inline, unused)) void
Perl_sv_vcatpvf_mg(pTHX_ SV * const goperl_a0, const char * const goperl_a1, va_list * const goperl_a2) { goperl_long_sv_vcatpvf_mg(goperl_a0, goperl_a1, goperl_a2); }
static inline __attribute__((always_inline, unused)) void
goperl_long_sv_vcatpvfn(SV * const goperl_a0, const char * const goperl_a1, const STRLEN goperl_a2, va_list * const goperl_a3, SV ** const goperl_a4, const Size_t goperl_a5, bool * const goperl_a6) { sv_vcatpvfn(goperl_a0, goperl_a1, goperl_a2, goperl_a3, goperl_a4, goperl_a5, goperl_a6); }
extern inline __attribute__((gnu_inline, always_inline, unused)) void
Perl_sv_vcatpvfn(pTHX_ SV * const goperl_a0, const char * const goperl_a1, const STRLEN goperl_a2, va_list * const goperl_a3, SV ** const goperl_a4, const Size_t goperl_a5, bool * const goperl_a6) { goperl_long_sv_vcatpvfn(goperl_a0, goperl_a1, goperl_a2, goperl_a3, goperl_a4, goperl_a5, goperl_a6); }
#if defined(sv_vcatpvfn_flags) && !defined(Perl_sv_vcatpvfn_flags)
#define Perl_sv_vcatpvfn_flags(...) GOPERL_LONG_CALL(sv_vcatpvfn_flags, GOPERL_THX_DROP(__VA_ARGS__))
#endif
static inline __attribute__((always_inline, unused)) void
goperl_long_sv_vsetpvf(SV * const goperl_a0, const char * const goperl_a1, va_list * const goperl_a2) { sv_vsetpvf(goperl_a0, goperl_a1, goperl_a2); }
extern inline __attribute__((gnu_inline, always_inline, unused)) void
Perl_sv_vsetpvf(pTHX_ SV * const goperl_a0, const char * const goperl_a1, va_list * const goperl_a2) { goperl_long_sv_vsetpvf(goperl_a0, goperl_a1, goperl_a2); }
static inline __attribute__((always_inline, unused)) void
goperl_long_sv_vsetpvf_mg(SV * const goperl_a0, const char * const goperl_a1, va_list * const goperl_a2) { sv_vsetpvf_mg(goperl_a0, goperl_a1, goperl_a2); }
extern inline __attribute__((gnu_inline, always_inline, unused)) void
Perl_sv_vsetpvf_mg(pTHX_ SV * const goperl_a0, const char * const goperl_a1, va_list * const goperl_a2) { goperl_long_sv_vsetpvf_mg(goperl_a0, goperl_a1, goperl_a2); }
static inline __attribute__((always_inline, unused)) void
goperl_long_sv_vsetpvfn(SV * const goperl_a0, const char * const goperl_a1, const STRLEN goperl_a2, va_list * const goperl_a3, SV ** const goperl_a4, const Size_t goperl_a5, bool * const goperl_a6) { sv_vsetpvfn(goperl_a0, goperl_a1, goperl_a2, goperl_a3, goperl_a4, goperl_a5, goperl_a6); }
extern inline __attribute__((gnu_inline, always_inline, unused)) void
Perl_sv_vsetpvfn(pTHX_ SV * const goperl_a0, const char * const goperl_a1, const STRLEN goperl_a2, va_list * const goperl_a3, SV ** const goperl_a4, const Size_t goperl_a5, bool * const goperl_a6) { goperl_long_sv_vsetpvfn(goperl_a0, goperl_a1, goperl_a2, goperl_a3, goperl_a4, goperl_a5, goperl_a6); }
#if defined(sv_vstring_get) && !defined(Perl_sv_vstring_get)
#define Perl_sv_vstring_get(...) GOPERL_LONG_CALL(sv_vstring_get, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(switch_to_global_locale) && !defined(Perl_switch_to_global_locale)
#define Perl_switch_to_global_locale(...) GOPERL_LONG_CALL(switch_to_global_locale, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(sync_locale) && !defined(Perl_sync_locale)
#define Perl_sync_locale(...) GOPERL_LONG_CALL(sync_locale, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(uiv_2buf) && !defined(Perl_uiv_2buf)
#define Perl_uiv_2buf(...) GOPERL_LONG_CALL(uiv_2buf, __VA_ARGS__)
#endif
#if defined(unpackstring) && !defined(Perl_unpackstring)
#define Perl_unpackstring(...) GOPERL_LONG_CALL(unpackstring, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(upg_version) && !defined(Perl_upg_version)
#define Perl_upg_version(...) GOPERL_LONG_CALL(upg_version, GOPERL_THX_DROP(__VA_ARGS__))
#endif
static inline __attribute__((always_inline, unused)) IV
goperl_long_utf8_distance(const U8 * goperl_a0, const U8 * goperl_a1) { return utf8_distance(goperl_a0, goperl_a1); }
extern inline __attribute__((gnu_inline, always_inline, unused)) IV
Perl_utf8_distance(pTHX_ const U8 * goperl_a0, const U8 * goperl_a1) { return goperl_long_utf8_distance(goperl_a0, goperl_a1); }
static inline __attribute__((always_inline, unused)) U8 *
goperl_long_utf8_hop(const U8 * goperl_a0, SSize_t goperl_a1) { return utf8_hop(goperl_a0, goperl_a1); }
extern inline __attribute__((gnu_inline, always_inline, unused)) U8 *
Perl_utf8_hop(const U8 * goperl_a0, SSize_t goperl_a1) { return goperl_long_utf8_hop(goperl_a0, goperl_a1); }
#if defined(utf8_hop_back) && !defined(Perl_utf8_hop_back)
#define Perl_utf8_hop_back(...) GOPERL_LONG_CALL(utf8_hop_back, __VA_ARGS__)
#endif
#if defined(utf8_hop_back_overshoot) && !defined(Perl_utf8_hop_back_overshoot)
#define Perl_utf8_hop_back_overshoot(...) GOPERL_LONG_CALL(utf8_hop_back_overshoot, __VA_ARGS__)
#endif
#if defined(utf8_hop_forward) && !defined(Perl_utf8_hop_forward)
#define Perl_utf8_hop_forward(...) GOPERL_LONG_CALL(utf8_hop_forward, __VA_ARGS__)
#endif
#if defined(utf8_hop_forward_overshoot) && !defined(Perl_utf8_hop_forward_overshoot)
#define Perl_utf8_hop_forward_overshoot(...) GOPERL_LONG_CALL(utf8_hop_forward_overshoot, __VA_ARGS__)
#endif
#if defined(utf8_hop_overshoot) && !defined(Perl_utf8_hop_overshoot)
#define Perl_utf8_hop_overshoot(...) GOPERL_LONG_CALL(utf8_hop_overshoot, __VA_ARGS__)
#endif
#if defined(utf8_hop_safe) && !defined(Perl_utf8_hop_safe)
#define Perl_utf8_hop_safe(...) GOPERL_LONG_CALL(utf8_hop_safe, __VA_ARGS__)
#endif
#if defined(utf8_length) && !defined(Perl_utf8_length)
#define Perl_utf8_length(...) GOPERL_LONG_CALL(utf8_length, GOPERL_THX_DROP(__VA_ARGS__))
#endif
static inline __attribute__((always_inline, unused)) UV
goperl_long_utf8n_to_uvchr(const U8 * goperl_a0, STRLEN goperl_a1, STRLEN * goperl_a2, const U32 goperl_a3) { return utf8n_to_uvchr(goperl_a0, goperl_a1, goperl_a2, goperl_a3); }
extern inline __attribute__((gnu_inline, always_inline, unused)) UV
Perl_utf8n_to_uvchr(const U8 * goperl_a0, STRLEN goperl_a1, STRLEN * goperl_a2, const U32 goperl_a3) { return goperl_long_utf8n_to_uvchr(goperl_a0, goperl_a1, goperl_a2, goperl_a3); }
#if defined(utf8n_to_uvchr_error) && !defined(Perl_utf8n_to_uvchr_error)
#define Perl_utf8n_to_uvchr_error(...) GOPERL_LONG_CALL(utf8n_to_uvchr_error, __VA_ARGS__)
#endif
#if defined(utf8n_to_uvchr_msgs) && !defined(Perl_utf8n_to_uvchr_msgs)
#define Perl_utf8n_to_uvchr_msgs(...) GOPERL_LONG_CALL(utf8n_to_uvchr_msgs, __VA_ARGS__)
#endif
static inline __attribute__((always_inline, unused)) U8 *
goperl_long_utf8_to_bytes(U8 * goperl_a0, STRLEN * goperl_a1) { return utf8_to_bytes(goperl_a0, goperl_a1); }
extern inline __attribute__((gnu_inline, always_inline, unused)) U8 *
Perl_utf8_to_bytes(pTHX_ U8 * goperl_a0, STRLEN * goperl_a1) { return goperl_long_utf8_to_bytes(goperl_a0, goperl_a1); }
#if defined(utf8_to_bytes_new_pv) && !defined(Perl_utf8_to_bytes_new_pv)
#define Perl_utf8_to_bytes_new_pv(...) GOPERL_LONG_CALL(utf8_to_bytes_new_pv, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(utf8_to_bytes_overwrite) && !defined(Perl_utf8_to_bytes_overwrite)
#define Perl_utf8_to_bytes_overwrite(...) GOPERL_LONG_CALL(utf8_to_bytes_overwrite, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(utf8_to_bytes_temp_pv) && !defined(Perl_utf8_to_bytes_temp_pv)
#define Perl_utf8_to_bytes_temp_pv(...) GOPERL_LONG_CALL(utf8_to_bytes_temp_pv, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(utf8_to_uv) && !defined(Perl_utf8_to_uv)
#define Perl_utf8_to_uv(...) GOPERL_LONG_CALL(utf8_to_uv, __VA_ARGS__)
#endif
static inline __attribute__((always_inline, unused)) UV
goperl_long_utf8_to_uvchr(const U8 * goperl_a0, STRLEN * goperl_a1) { return utf8_to_uvchr(goperl_a0, goperl_a1); }
extern inline __attribute__((gnu_inline, always_inline, unused)) UV
Perl_utf8_to_uvchr(pTHX_ const U8 * goperl_a0, STRLEN * goperl_a1) { return goperl_long_utf8_to_uvchr(goperl_a0, goperl_a1); }
#if defined(utf8_to_uvchr_buf) && !defined(Perl_utf8_to_uvchr_buf)
#define Perl_utf8_to_uvchr_buf(...) GOPERL_LONG_CALL(utf8_to_uvchr_buf, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(utf8_to_uv_errors) && !defined(Perl_utf8_to_uv_errors)
#define Perl_utf8_to_uv_errors(...) GOPERL_LONG_CALL(utf8_to_uv_errors, __VA_ARGS__)
#endif
#if defined(utf8_to_uv_flags) && !defined(Perl_utf8_to_uv_flags)
#define Perl_utf8_to_uv_flags(...) GOPERL_LONG_CALL(utf8_to_uv_flags, __VA_ARGS__)
#endif
#if defined(utf8_to_uv_msgs) && !defined(Perl_utf8_to_uv_msgs)
#define Perl_utf8_to_uv_msgs(...) GOPERL_LONG_CALL(utf8_to_uv_msgs, __VA_ARGS__)
#endif
#if defined(utf8_to_uv_or_die) && !defined(Perl_utf8_to_uv_or_die)
#define Perl_utf8_to_uv_or_die(...) GOPERL_LONG_CALL(utf8_to_uv_or_die, __VA_ARGS__)
#endif
static inline __attribute__((always_inline, unused)) U8 *
goperl_long_uvchr_to_utf8(U8 * goperl_a0, UV goperl_a1) { return uvchr_to_utf8(goperl_a0, goperl_a1); }
extern inline __attribute__((gnu_inline, always_inline, unused)) U8 *
Perl_uvchr_to_utf8(pTHX_ U8 * goperl_a0, UV goperl_a1) { return goperl_long_uvchr_to_utf8(goperl_a0, goperl_a1); }
static inline __attribute__((always_inline, unused)) U8 *
goperl_long_uvchr_to_utf8_flags(U8 * goperl_a0, UV goperl_a1, UV goperl_a2) { return uvchr_to_utf8_flags(goperl_a0, goperl_a1, goperl_a2); }
extern inline __attribute__((gnu_inline, always_inline, unused)) U8 *
Perl_uvchr_to_utf8_flags(pTHX_ U8 * goperl_a0, UV goperl_a1, UV goperl_a2) { return goperl_long_uvchr_to_utf8_flags(goperl_a0, goperl_a1, goperl_a2); }
#if defined(uvchr_to_utf8_flags_msgs) && !defined(Perl_uvchr_to_utf8_flags_msgs)
#define Perl_uvchr_to_utf8_flags_msgs(...) GOPERL_LONG_CALL(uvchr_to_utf8_flags_msgs, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(uv_to_utf8) && !defined(Perl_uv_to_utf8)
#define Perl_uv_to_utf8(...) GOPERL_LONG_CALL(uv_to_utf8, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(uv_to_utf8_flags) && !defined(Perl_uv_to_utf8_flags)
#define Perl_uv_to_utf8_flags(...) GOPERL_LONG_CALL(uv_to_utf8_flags, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(uv_to_utf8_msgs) && !defined(Perl_uv_to_utf8_msgs)
#define Perl_uv_to_utf8_msgs(...) GOPERL_LONG_CALL(uv_to_utf8_msgs, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(valid_identifier_pve) && !defined(Perl_valid_identifier_pve)
#define Perl_valid_identifier_pve(...) GOPERL_LONG_CALL(valid_identifier_pve, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(valid_identifier_pvn) && !defined(Perl_valid_identifier_pvn)
#define Perl_valid_identifier_pvn(...) GOPERL_LONG_CALL(valid_identifier_pvn, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(valid_identifier_sv) && !defined(Perl_valid_identifier_sv)
#define Perl_valid_identifier_sv(...) GOPERL_LONG_CALL(valid_identifier_sv, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(vcmp) && !defined(Perl_vcmp)
#define Perl_vcmp(...) GOPERL_LONG_CALL(vcmp, GOPERL_THX_DROP(__VA_ARGS__))
#endif
static inline __attribute__((always_inline, unused)) void
goperl_long_vcroak(const char * goperl_a0, va_list * goperl_a1) { vcroak(goperl_a0, goperl_a1); }
extern inline __attribute__((gnu_inline, always_inline, unused)) void
Perl_vcroak(pTHX_ const char * goperl_a0, va_list * goperl_a1) { goperl_long_vcroak(goperl_a0, goperl_a1); }
#if defined(vdeb) && !defined(Perl_vdeb)
#define Perl_vdeb(...) GOPERL_LONG_CALL(vdeb, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(vfatal_warner) && !defined(Perl_vfatal_warner)
#define Perl_vfatal_warner(...) GOPERL_LONG_CALL(vfatal_warner, GOPERL_THX_DROP(__VA_ARGS__))
#endif
static inline __attribute__((always_inline, unused)) char *
goperl_long_vform(const char * goperl_a0, va_list * goperl_a1) { return vform(goperl_a0, goperl_a1); }
extern inline __attribute__((gnu_inline, always_inline, unused)) char *
Perl_vform(pTHX_ const char * goperl_a0, va_list * goperl_a1) { return goperl_long_vform(goperl_a0, goperl_a1); }
#if defined(vload_module) && !defined(Perl_vload_module)
#define Perl_vload_module(...) GOPERL_LONG_CALL(vload_module, GOPERL_THX_DROP(__VA_ARGS__))
#endif
static inline __attribute__((always_inline, unused)) SV *
goperl_long_vmess(const char * goperl_a0, va_list * goperl_a1) { return vmess(goperl_a0, goperl_a1); }
extern inline __attribute__((gnu_inline, always_inline, unused)) SV *
Perl_vmess(pTHX_ const char * goperl_a0, va_list * goperl_a1) { return goperl_long_vmess(goperl_a0, goperl_a1); }
static inline __attribute__((always_inline, unused)) SV *
goperl_long_vnewSVpvf(const char * const goperl_a0, va_list * const goperl_a1) { return vnewSVpvf(goperl_a0, goperl_a1); }
extern inline __attribute__((gnu_inline, always_inline, unused)) SV *
Perl_vnewSVpvf(pTHX_ const char * const goperl_a0, va_list * const goperl_a1) { return goperl_long_vnewSVpvf(goperl_a0, goperl_a1); }
#if defined(vnormal) && !defined(Perl_vnormal)
#define Perl_vnormal(...) GOPERL_LONG_CALL(vnormal, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(vnumify) && !defined(Perl_vnumify)
#define Perl_vnumify(...) GOPERL_LONG_CALL(vnumify, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(vstringify) && !defined(Perl_vstringify)
#define Perl_vstringify(...) GOPERL_LONG_CALL(vstringify, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(vverify) && !defined(Perl_vverify)
#define Perl_vverify(...) GOPERL_LONG_CALL(vverify, GOPERL_THX_DROP(__VA_ARGS__))
#endif
static inline __attribute__((always_inline, unused)) void
goperl_long_vwarn(const char * goperl_a0, va_list * goperl_a1) { vwarn(goperl_a0, goperl_a1); }
extern inline __attribute__((gnu_inline, always_inline, unused)) void
Perl_vwarn(pTHX_ const char * goperl_a0, va_list * goperl_a1) { goperl_long_vwarn(goperl_a0, goperl_a1); }
#if defined(vwarner) && !defined(Perl_vwarner)
#define Perl_vwarner(...) GOPERL_LONG_CALL(vwarner, GOPERL_THX_DROP(__VA_ARGS__))
#endif
static inline __attribute__((always_inline, unused)) void
goperl_long_warn_sv(SV * goperl_a0) { warn_sv(goperl_a0); }
extern inline __attribute__((gnu_inline, always_inline, unused)) void
Perl_warn_sv(pTHX_ SV * goperl_a0) { goperl_long_warn_sv(goperl_a0); }
static inline __attribute__((always_inline, unused)) I32
goperl_long_whichsig(const char * goperl_a0) { return whichsig(goperl_a0); }
extern inline __attribute__((gnu_inline, always_inline, unused)) I32
Perl_whichsig(pTHX_ const char * goperl_a0) { return goperl_long_whichsig(goperl_a0); }
static inline __attribute__((always_inline, unused)) I32
goperl_long_whichsig_pv(const char * goperl_a0) { return whichsig_pv(goperl_a0); }
extern inline __attribute__((gnu_inline, always_inline, unused)) I32
Perl_whichsig_pv(pTHX_ const char * goperl_a0) { return goperl_long_whichsig_pv(goperl_a0); }
static inline __attribute__((always_inline, unused)) I32
goperl_long_whichsig_pvn(const char * goperl_a0, STRLEN goperl_a1) { return whichsig_pvn(goperl_a0, goperl_a1); }
extern inline __attribute__((gnu_inline, always_inline, unused)) I32
Perl_whichsig_pvn(pTHX_ const char * goperl_a0, STRLEN goperl_a1) { return goperl_long_whichsig_pvn(goperl_a0, goperl_a1); }
#if defined(whichsig_sv) && !defined(Perl_whichsig_sv)
#define Perl_whichsig_sv(...) GOPERL_LONG_CALL(whichsig_sv, GOPERL_THX_DROP(__VA_ARGS__))
#endif
static inline __attribute__((always_inline, unused)) void
goperl_long_wrap_infix_plugin(Perl_infix_plugin_t goperl_a0, Perl_infix_plugin_t * goperl_a1) { wrap_infix_plugin(goperl_a0, goperl_a1); }
extern inline __attribute__((gnu_inline, always_inline, unused)) void
Perl_wrap_infix_plugin(pTHX_ Perl_infix_plugin_t goperl_a0, Perl_infix_plugin_t * goperl_a1) { goperl_long_wrap_infix_plugin(goperl_a0, goperl_a1); }
static inline __attribute__((always_inline, unused)) void
goperl_long_wrap_keyword_plugin(Perl_keyword_plugin_t goperl_a0, Perl_keyword_plugin_t * goperl_a1) { wrap_keyword_plugin(goperl_a0, goperl_a1); }
extern inline __attribute__((gnu_inline, always_inline, unused)) void
Perl_wrap_keyword_plugin(pTHX_ Perl_keyword_plugin_t goperl_a0, Perl_keyword_plugin_t * goperl_a1) { goperl_long_wrap_keyword_plugin(goperl_a0, goperl_a1); }
#if defined(hv_assert) && !defined(Perl_hv_assert)
#define Perl_hv_assert(...) GOPERL_LONG_CALL(hv_assert, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(my_chsize) && !defined(Perl_my_chsize)
#define Perl_my_chsize(...) GOPERL_LONG_CALL(my_chsize, GOPERL_THX_DROP(__VA_ARGS__))
#endif
static inline __attribute__((always_inline, unused)) Size_t
goperl_long_my_strlcat(char * goperl_a0, const char * goperl_a1, Size_t goperl_a2) { return my_strlcat(goperl_a0, goperl_a1, goperl_a2); }
extern inline __attribute__((gnu_inline, always_inline, unused)) Size_t
Perl_my_strlcat(char * goperl_a0, const char * goperl_a1, Size_t goperl_a2) { return goperl_long_my_strlcat(goperl_a0, goperl_a1, goperl_a2); }
#if defined(my_strnlen) && !defined(Perl_my_strnlen)
#define Perl_my_strnlen(...) GOPERL_LONG_CALL(my_strnlen, __VA_ARGS__)
#endif
#if defined(deb_nocontext) && !defined(Perl_deb_nocontext)
#define Perl_deb_nocontext(...) GOPERL_LONG_CALL(deb_nocontext, __VA_ARGS__)
#endif
#if defined(die_nocontext) && !defined(Perl_die_nocontext)
#define Perl_die_nocontext(...) GOPERL_LONG_CALL(die_nocontext, __VA_ARGS__)
#endif
#if defined(load_module_nocontext) && !defined(Perl_load_module_nocontext)
#define Perl_load_module_nocontext(...) GOPERL_LONG_CALL(load_module_nocontext, __VA_ARGS__)
#endif
#if defined(mess_nocontext) && !defined(Perl_mess_nocontext)
#define Perl_mess_nocontext(...) GOPERL_LONG_CALL(mess_nocontext, __VA_ARGS__)
#endif
#if defined(newSVpvf_nocontext) && !defined(Perl_newSVpvf_nocontext)
#define Perl_newSVpvf_nocontext(...) GOPERL_LONG_CALL(newSVpvf_nocontext, __VA_ARGS__)
#endif
#if defined(sv_catpvf_mg_nocontext) && !defined(Perl_sv_catpvf_mg_nocontext)
#define Perl_sv_catpvf_mg_nocontext(...) GOPERL_LONG_CALL(sv_catpvf_mg_nocontext, __VA_ARGS__)
#endif
#if defined(sv_catpvf_nocontext) && !defined(Perl_sv_catpvf_nocontext)
#define Perl_sv_catpvf_nocontext(...) GOPERL_LONG_CALL(sv_catpvf_nocontext, __VA_ARGS__)
#endif
#if defined(sv_setpvf_mg_nocontext) && !defined(Perl_sv_setpvf_mg_nocontext)
#define Perl_sv_setpvf_mg_nocontext(...) GOPERL_LONG_CALL(sv_setpvf_mg_nocontext, __VA_ARGS__)
#endif
#if defined(sv_setpvf_nocontext) && !defined(Perl_sv_setpvf_nocontext)
#define Perl_sv_setpvf_nocontext(...) GOPERL_LONG_CALL(sv_setpvf_nocontext, __VA_ARGS__)
#endif
#if defined(warner_nocontext) && !defined(Perl_warner_nocontext)
#define Perl_warner_nocontext(...) GOPERL_LONG_CALL(warner_nocontext, __VA_ARGS__)
#endif
#if defined(dump_mstats) && !defined(Perl_dump_mstats)
#define Perl_dump_mstats(...) GOPERL_LONG_CALL(dump_mstats, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(finalize_optree) && !defined(Perl_finalize_optree)
#define Perl_finalize_optree(...) GOPERL_LONG_CALL(finalize_optree, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(optimize_optree) && !defined(Perl_optimize_optree)
#define Perl_optimize_optree(...) GOPERL_LONG_CALL(optimize_optree, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(my_pclose) && !defined(Perl_my_pclose)
#define Perl_my_pclose(...) GOPERL_LONG_CALL(my_pclose, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(my_popen) && !defined(Perl_my_popen)
#define Perl_my_popen(...) GOPERL_LONG_CALL(my_popen, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(dump_c_backtrace) && !defined(Perl_dump_c_backtrace)
#define Perl_dump_c_backtrace(...) GOPERL_LONG_CALL(dump_c_backtrace, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(get_c_backtrace_dump) && !defined(Perl_get_c_backtrace_dump)
#define Perl_get_c_backtrace_dump(...) GOPERL_LONG_CALL(get_c_backtrace_dump, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(alloccopstash) && !defined(Perl_alloccopstash)
#define Perl_alloccopstash(...) GOPERL_LONG_CALL(alloccopstash, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(clone_params_del) && !defined(Perl_clone_params_del)
#define Perl_clone_params_del(...) GOPERL_LONG_CALL(clone_params_del, __VA_ARGS__)
#endif
#if defined(clone_params_new) && !defined(Perl_clone_params_new)
#define Perl_clone_params_new(...) GOPERL_LONG_CALL(clone_params_new, __VA_ARGS__)
#endif
static inline __attribute__((always_inline, unused)) OP *
goperl_long_newPADOP(I32 goperl_a0, I32 goperl_a1, SV * goperl_a2) { return newPADOP(goperl_a0, goperl_a1, goperl_a2); }
extern inline __attribute__((gnu_inline, always_inline, unused)) OP *
Perl_newPADOP(pTHX_ I32 goperl_a0, I32 goperl_a1, SV * goperl_a2) { return goperl_long_newPADOP(goperl_a0, goperl_a1, goperl_a2); }
#if defined(re_dup_guts) && !defined(Perl_re_dup_guts)
#define Perl_re_dup_guts(...) GOPERL_LONG_CALL(re_dup_guts, GOPERL_THX_DROP(__VA_ARGS__))
#endif
static inline __attribute__((always_inline, unused)) SV *
goperl_long_sv_dup(const SV * const goperl_a0, CLONE_PARAMS * const goperl_a1) { return sv_dup(goperl_a0, goperl_a1); }
extern inline __attribute__((gnu_inline, always_inline, unused)) SV *
Perl_sv_dup(pTHX_ const SV * const goperl_a0, CLONE_PARAMS * const goperl_a1) { return goperl_long_sv_dup(goperl_a0, goperl_a1); }
static inline __attribute__((always_inline, unused)) SV *
goperl_long_sv_dup_inc(const SV * const goperl_a0, CLONE_PARAMS * const goperl_a1) { return sv_dup_inc(goperl_a0, goperl_a1); }
extern inline __attribute__((gnu_inline, always_inline, unused)) SV *
Perl_sv_dup_inc(pTHX_ const SV * const goperl_a0, CLONE_PARAMS * const goperl_a1) { return goperl_long_sv_dup_inc(goperl_a0, goperl_a1); }
#if defined(CopFILEGV_set) && !defined(Perl_CopFILEGV_set)
#define Perl_CopFILEGV_set(...) GOPERL_LONG_CALL(CopFILEGV_set, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(sv_collxfrm) && !defined(Perl_sv_collxfrm)
#define Perl_sv_collxfrm(...) GOPERL_LONG_CALL(sv_collxfrm, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(sv_collxfrm_flags) && !defined(Perl_sv_collxfrm_flags)
#define Perl_sv_collxfrm_flags(...) GOPERL_LONG_CALL(sv_collxfrm_flags, GOPERL_THX_DROP(__VA_ARGS__))
#endif
static inline __attribute__((always_inline, unused)) void
goperl_long_PerlIO_clearerr(PerlIO * goperl_a0) { PerlIO_clearerr(goperl_a0); }
extern inline __attribute__((gnu_inline, always_inline, unused)) void
Perl_PerlIO_clearerr(pTHX_ PerlIO * goperl_a0) { goperl_long_PerlIO_clearerr(goperl_a0); }
static inline __attribute__((always_inline, unused)) int
goperl_long_PerlIO_close(PerlIO * goperl_a0) { return PerlIO_close(goperl_a0); }
extern inline __attribute__((gnu_inline, always_inline, unused)) int
Perl_PerlIO_close(pTHX_ PerlIO * goperl_a0) { return goperl_long_PerlIO_close(goperl_a0); }
static inline __attribute__((always_inline, unused)) int
goperl_long_PerlIO_eof(PerlIO * goperl_a0) { return PerlIO_eof(goperl_a0); }
extern inline __attribute__((gnu_inline, always_inline, unused)) int
Perl_PerlIO_eof(pTHX_ PerlIO * goperl_a0) { return goperl_long_PerlIO_eof(goperl_a0); }
static inline __attribute__((always_inline, unused)) int
goperl_long_PerlIO_error(PerlIO * goperl_a0) { return PerlIO_error(goperl_a0); }
extern inline __attribute__((gnu_inline, always_inline, unused)) int
Perl_PerlIO_error(pTHX_ PerlIO * goperl_a0) { return goperl_long_PerlIO_error(goperl_a0); }
static inline __attribute__((always_inline, unused)) int
goperl_long_PerlIO_fileno(PerlIO * goperl_a0) { return PerlIO_fileno(goperl_a0); }
extern inline __attribute__((gnu_inline, always_inline, unused)) int
Perl_PerlIO_fileno(pTHX_ PerlIO * goperl_a0) { return goperl_long_PerlIO_fileno(goperl_a0); }
static inline __attribute__((always_inline, unused)) int
goperl_long_PerlIO_fill(PerlIO * goperl_a0) { return PerlIO_fill(goperl_a0); }
extern inline __attribute__((gnu_inline, always_inline, unused)) int
Perl_PerlIO_fill(pTHX_ PerlIO * goperl_a0) { return goperl_long_PerlIO_fill(goperl_a0); }
static inline __attribute__((always_inline, unused)) int
goperl_long_PerlIO_flush(PerlIO * goperl_a0) { return PerlIO_flush(goperl_a0); }
extern inline __attribute__((gnu_inline, always_inline, unused)) int
Perl_PerlIO_flush(pTHX_ PerlIO * goperl_a0) { return goperl_long_PerlIO_flush(goperl_a0); }
static inline __attribute__((always_inline, unused)) STDCHAR *
goperl_long_PerlIO_get_base(PerlIO * goperl_a0) { return PerlIO_get_base(goperl_a0); }
extern inline __attribute__((gnu_inline, always_inline, unused)) STDCHAR *
Perl_PerlIO_get_base(pTHX_ PerlIO * goperl_a0) { return goperl_long_PerlIO_get_base(goperl_a0); }
static inline __attribute__((always_inline, unused)) SSize_t
goperl_long_PerlIO_get_bufsiz(PerlIO * goperl_a0) { return PerlIO_get_bufsiz(goperl_a0); }
extern inline __attribute__((gnu_inline, always_inline, unused)) SSize_t
Perl_PerlIO_get_bufsiz(pTHX_ PerlIO * goperl_a0) { return goperl_long_PerlIO_get_bufsiz(goperl_a0); }
static inline __attribute__((always_inline, unused)) SSize_t
goperl_long_PerlIO_get_cnt(PerlIO * goperl_a0) { return PerlIO_get_cnt(goperl_a0); }
extern inline __attribute__((gnu_inline, always_inline, unused)) SSize_t
Perl_PerlIO_get_cnt(pTHX_ PerlIO * goperl_a0) { return goperl_long_PerlIO_get_cnt(goperl_a0); }
static inline __attribute__((always_inline, unused)) STDCHAR *
goperl_long_PerlIO_get_ptr(PerlIO * goperl_a0) { return PerlIO_get_ptr(goperl_a0); }
extern inline __attribute__((gnu_inline, always_inline, unused)) STDCHAR *
Perl_PerlIO_get_ptr(pTHX_ PerlIO * goperl_a0) { return goperl_long_PerlIO_get_ptr(goperl_a0); }
static inline __attribute__((always_inline, unused)) SSize_t
goperl_long_PerlIO_read(PerlIO * goperl_a0, void * goperl_a1, Size_t goperl_a2) { return PerlIO_read(goperl_a0, goperl_a1, goperl_a2); }
extern inline __attribute__((gnu_inline, always_inline, unused)) SSize_t
Perl_PerlIO_read(pTHX_ PerlIO * goperl_a0, void * goperl_a1, Size_t goperl_a2) { return goperl_long_PerlIO_read(goperl_a0, goperl_a1, goperl_a2); }
static inline __attribute__((always_inline, unused)) int
goperl_long_PerlIO_seek(PerlIO * goperl_a0, Off_t goperl_a1, int goperl_a2) { return PerlIO_seek(goperl_a0, goperl_a1, goperl_a2); }
extern inline __attribute__((gnu_inline, always_inline, unused)) int
Perl_PerlIO_seek(pTHX_ PerlIO * goperl_a0, Off_t goperl_a1, int goperl_a2) { return goperl_long_PerlIO_seek(goperl_a0, goperl_a1, goperl_a2); }
#if defined(PerlIO_set_cnt) && !defined(Perl_PerlIO_set_cnt)
#define Perl_PerlIO_set_cnt(...) GOPERL_LONG_CALL(PerlIO_set_cnt, GOPERL_THX_DROP(__VA_ARGS__))
#endif
static inline __attribute__((always_inline, unused)) void
goperl_long_PerlIO_setlinebuf(PerlIO * goperl_a0) { PerlIO_setlinebuf(goperl_a0); }
extern inline __attribute__((gnu_inline, always_inline, unused)) void
Perl_PerlIO_setlinebuf(pTHX_ PerlIO * goperl_a0) { goperl_long_PerlIO_setlinebuf(goperl_a0); }
static inline __attribute__((always_inline, unused)) void
goperl_long_PerlIO_set_ptrcnt(PerlIO * goperl_a0, STDCHAR * goperl_a1, SSize_t goperl_a2) { PerlIO_set_ptrcnt(goperl_a0, goperl_a1, goperl_a2); }
extern inline __attribute__((gnu_inline, always_inline, unused)) void
Perl_PerlIO_set_ptrcnt(pTHX_ PerlIO * goperl_a0, STDCHAR * goperl_a1, SSize_t goperl_a2) { goperl_long_PerlIO_set_ptrcnt(goperl_a0, goperl_a1, goperl_a2); }
static inline __attribute__((always_inline, unused)) PerlIO *
goperl_long_PerlIO_stderr(void) { return PerlIO_stderr(); }
extern inline __attribute__((gnu_inline, always_inline, unused)) PerlIO *
Perl_PerlIO_stderr(pTHX) { return goperl_long_PerlIO_stderr(); }
static inline __attribute__((always_inline, unused)) PerlIO *
goperl_long_PerlIO_stdin(void) { return PerlIO_stdin(); }
extern inline __attribute__((gnu_inline, always_inline, unused)) PerlIO *
Perl_PerlIO_stdin(pTHX) { return goperl_long_PerlIO_stdin(); }
static inline __attribute__((always_inline, unused)) PerlIO *
goperl_long_PerlIO_stdout(void) { return PerlIO_stdout(); }
extern inline __attribute__((gnu_inline, always_inline, unused)) PerlIO *
Perl_PerlIO_stdout(pTHX) { return goperl_long_PerlIO_stdout(); }
static inline __attribute__((always_inline, unused)) Off_t
goperl_long_PerlIO_tell(PerlIO * goperl_a0) { return PerlIO_tell(goperl_a0); }
extern inline __attribute__((gnu_inline, always_inline, unused)) Off_t
Perl_PerlIO_tell(pTHX_ PerlIO * goperl_a0) { return goperl_long_PerlIO_tell(goperl_a0); }
static inline __attribute__((always_inline, unused)) SSize_t
goperl_long_PerlIO_unread(PerlIO * goperl_a0, const void * goperl_a1, Size_t goperl_a2) { return PerlIO_unread(goperl_a0, goperl_a1, goperl_a2); }
extern inline __attribute__((gnu_inline, always_inline, unused)) SSize_t
Perl_PerlIO_unread(pTHX_ PerlIO * goperl_a0, const void * goperl_a1, Size_t goperl_a2) { return goperl_long_PerlIO_unread(goperl_a0, goperl_a1, goperl_a2); }
static inline __attribute__((always_inline, unused)) SSize_t
goperl_long_PerlIO_write(PerlIO * goperl_a0, const void * goperl_a1, Size_t goperl_a2) { return PerlIO_write(goperl_a0, goperl_a1, goperl_a2); }
extern inline __attribute__((gnu_inline, always_inline, unused)) SSize_t
Perl_PerlIO_write(pTHX_ PerlIO * goperl_a0, const void * goperl_a1, Size_t goperl_a2) { return goperl_long_PerlIO_write(goperl_a0, goperl_a1, goperl_a2); }
#if defined(thread_locale_init) && !defined(Perl_thread_locale_init)
#define Perl_thread_locale_init(...) GOPERL_LONG_CALL(thread_locale_init, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(thread_locale_term) && !defined(Perl_thread_locale_term)
#define Perl_thread_locale_term(...) GOPERL_LONG_CALL(thread_locale_term, GOPERL_THX_DROP(__VA_ARGS__))
#endif

#endif /* GOPERL_XS_SDK_LONGNAMES_H */
