/* go-perl native XS SDK: the long names of perl's public functions
 * (GENERATED from perl v5.44.0's embed.fnc; do not edit by hand).
 *
 * Perl names every public function of its API Perl_<name> as well; code
 * calls it so with the interpreter as the first argument, Perl_av_push(aTHX_
 * av, sv), or none for a function that takes no interpreter. Each is the
 * short name the SDK provides, where it provides one (as a macro), with
 * the interpreter left out: aTHX_ is one argument of the call before it
 * is expanded, and the interpreter is the SDK's current frame anyway. */
#ifndef GOPERL_XS_SDK_LONGNAMES_H
#define GOPERL_XS_SDK_LONGNAMES_H

#define GOPERL_THX_DROP(ctx, ...) __VA_ARGS__
/* the arguments are expanded before the short name's macro takes them */
#define GOPERL_LONG_CALL(f, ...) f(__VA_ARGS__)

#if defined(CopFILEGV_set) && !defined(Perl_CopFILEGV_set)
#define Perl_CopFILEGV_set(...) GOPERL_LONG_CALL(CopFILEGV_set, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(CvDEPTH) && !defined(Perl_CvDEPTH)
#define Perl_CvDEPTH(...) GOPERL_LONG_CALL(CvDEPTH, __VA_ARGS__)
#endif
#if defined(CvGV) && !defined(Perl_CvGV)
#define Perl_CvGV(...) GOPERL_LONG_CALL(CvGV, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(Gv_AMupdate) && !defined(Perl_Gv_AMupdate)
#define Perl_Gv_AMupdate(...) GOPERL_LONG_CALL(Gv_AMupdate, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(PerlIO_clearerr) && !defined(Perl_PerlIO_clearerr)
#define Perl_PerlIO_clearerr(...) GOPERL_LONG_CALL(PerlIO_clearerr, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(PerlIO_close) && !defined(Perl_PerlIO_close)
#define Perl_PerlIO_close(...) GOPERL_LONG_CALL(PerlIO_close, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(PerlIO_eof) && !defined(Perl_PerlIO_eof)
#define Perl_PerlIO_eof(...) GOPERL_LONG_CALL(PerlIO_eof, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(PerlIO_error) && !defined(Perl_PerlIO_error)
#define Perl_PerlIO_error(...) GOPERL_LONG_CALL(PerlIO_error, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(PerlIO_fileno) && !defined(Perl_PerlIO_fileno)
#define Perl_PerlIO_fileno(...) GOPERL_LONG_CALL(PerlIO_fileno, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(PerlIO_fill) && !defined(Perl_PerlIO_fill)
#define Perl_PerlIO_fill(...) GOPERL_LONG_CALL(PerlIO_fill, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(PerlIO_flush) && !defined(Perl_PerlIO_flush)
#define Perl_PerlIO_flush(...) GOPERL_LONG_CALL(PerlIO_flush, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(PerlIO_get_base) && !defined(Perl_PerlIO_get_base)
#define Perl_PerlIO_get_base(...) GOPERL_LONG_CALL(PerlIO_get_base, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(PerlIO_get_bufsiz) && !defined(Perl_PerlIO_get_bufsiz)
#define Perl_PerlIO_get_bufsiz(...) GOPERL_LONG_CALL(PerlIO_get_bufsiz, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(PerlIO_get_cnt) && !defined(Perl_PerlIO_get_cnt)
#define Perl_PerlIO_get_cnt(...) GOPERL_LONG_CALL(PerlIO_get_cnt, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(PerlIO_get_ptr) && !defined(Perl_PerlIO_get_ptr)
#define Perl_PerlIO_get_ptr(...) GOPERL_LONG_CALL(PerlIO_get_ptr, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(PerlIO_read) && !defined(Perl_PerlIO_read)
#define Perl_PerlIO_read(...) GOPERL_LONG_CALL(PerlIO_read, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(PerlIO_seek) && !defined(Perl_PerlIO_seek)
#define Perl_PerlIO_seek(...) GOPERL_LONG_CALL(PerlIO_seek, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(PerlIO_set_cnt) && !defined(Perl_PerlIO_set_cnt)
#define Perl_PerlIO_set_cnt(...) GOPERL_LONG_CALL(PerlIO_set_cnt, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(PerlIO_set_ptrcnt) && !defined(Perl_PerlIO_set_ptrcnt)
#define Perl_PerlIO_set_ptrcnt(...) GOPERL_LONG_CALL(PerlIO_set_ptrcnt, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(PerlIO_setlinebuf) && !defined(Perl_PerlIO_setlinebuf)
#define Perl_PerlIO_setlinebuf(...) GOPERL_LONG_CALL(PerlIO_setlinebuf, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(PerlIO_stderr) && !defined(Perl_PerlIO_stderr)
#define Perl_PerlIO_stderr(...) GOPERL_LONG_CALL(PerlIO_stderr, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(PerlIO_stdin) && !defined(Perl_PerlIO_stdin)
#define Perl_PerlIO_stdin(...) GOPERL_LONG_CALL(PerlIO_stdin, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(PerlIO_stdout) && !defined(Perl_PerlIO_stdout)
#define Perl_PerlIO_stdout(...) GOPERL_LONG_CALL(PerlIO_stdout, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(PerlIO_tell) && !defined(Perl_PerlIO_tell)
#define Perl_PerlIO_tell(...) GOPERL_LONG_CALL(PerlIO_tell, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(PerlIO_unread) && !defined(Perl_PerlIO_unread)
#define Perl_PerlIO_unread(...) GOPERL_LONG_CALL(PerlIO_unread, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(PerlIO_write) && !defined(Perl_PerlIO_write)
#define Perl_PerlIO_write(...) GOPERL_LONG_CALL(PerlIO_write, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(SvAMAGIC_off) && !defined(Perl_SvAMAGIC_off)
#define Perl_SvAMAGIC_off(...) GOPERL_LONG_CALL(SvAMAGIC_off, __VA_ARGS__)
#endif
#if defined(SvAMAGIC_on) && !defined(Perl_SvAMAGIC_on)
#define Perl_SvAMAGIC_on(...) GOPERL_LONG_CALL(SvAMAGIC_on, __VA_ARGS__)
#endif
#if defined(SvGETMAGIC) && !defined(Perl_SvGETMAGIC)
#define Perl_SvGETMAGIC(...) GOPERL_LONG_CALL(SvGETMAGIC, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(SvIV) && !defined(Perl_SvIV)
#define Perl_SvIV(...) GOPERL_LONG_CALL(SvIV, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(SvIV_nomg) && !defined(Perl_SvIV_nomg)
#define Perl_SvIV_nomg(...) GOPERL_LONG_CALL(SvIV_nomg, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(SvNV) && !defined(Perl_SvNV)
#define Perl_SvNV(...) GOPERL_LONG_CALL(SvNV, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(SvNV_nomg) && !defined(Perl_SvNV_nomg)
#define Perl_SvNV_nomg(...) GOPERL_LONG_CALL(SvNV_nomg, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(SvPVXtrue) && !defined(Perl_SvPVXtrue)
#define Perl_SvPVXtrue(...) GOPERL_LONG_CALL(SvPVXtrue, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(SvREFCNT_dec) && !defined(Perl_SvREFCNT_dec)
#define Perl_SvREFCNT_dec(...) GOPERL_LONG_CALL(SvREFCNT_dec, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(SvREFCNT_dec_NN) && !defined(Perl_SvREFCNT_dec_NN)
#define Perl_SvREFCNT_dec_NN(...) GOPERL_LONG_CALL(SvREFCNT_dec_NN, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(SvREFCNT_dec_ret_NULL) && !defined(Perl_SvREFCNT_dec_ret_NULL)
#define Perl_SvREFCNT_dec_ret_NULL(...) GOPERL_LONG_CALL(SvREFCNT_dec_ret_NULL, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(SvREFCNT_dec_set_NULL) && !defined(Perl_SvREFCNT_dec_set_NULL)
#define Perl_SvREFCNT_dec_set_NULL(...) GOPERL_LONG_CALL(SvREFCNT_dec_set_NULL, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(SvREFCNT_inc) && !defined(Perl_SvREFCNT_inc)
#define Perl_SvREFCNT_inc(...) GOPERL_LONG_CALL(SvREFCNT_inc, __VA_ARGS__)
#endif
#if defined(SvREFCNT_inc_NN) && !defined(Perl_SvREFCNT_inc_NN)
#define Perl_SvREFCNT_inc_NN(...) GOPERL_LONG_CALL(SvREFCNT_inc_NN, __VA_ARGS__)
#endif
#if defined(SvREFCNT_inc_void) && !defined(Perl_SvREFCNT_inc_void)
#define Perl_SvREFCNT_inc_void(...) GOPERL_LONG_CALL(SvREFCNT_inc_void, __VA_ARGS__)
#endif
#if defined(SvTRUE) && !defined(Perl_SvTRUE)
#define Perl_SvTRUE(...) GOPERL_LONG_CALL(SvTRUE, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(SvTRUE_NN) && !defined(Perl_SvTRUE_NN)
#define Perl_SvTRUE_NN(...) GOPERL_LONG_CALL(SvTRUE_NN, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(SvTRUE_nomg) && !defined(Perl_SvTRUE_nomg)
#define Perl_SvTRUE_nomg(...) GOPERL_LONG_CALL(SvTRUE_nomg, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(SvUV) && !defined(Perl_SvUV)
#define Perl_SvUV(...) GOPERL_LONG_CALL(SvUV, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(SvUV_nomg) && !defined(Perl_SvUV_nomg)
#define Perl_SvUV_nomg(...) GOPERL_LONG_CALL(SvUV_nomg, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(alloccopstash) && !defined(Perl_alloccopstash)
#define Perl_alloccopstash(...) GOPERL_LONG_CALL(alloccopstash, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(amagic_call) && !defined(Perl_amagic_call)
#define Perl_amagic_call(...) GOPERL_LONG_CALL(amagic_call, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(amagic_deref_call) && !defined(Perl_amagic_deref_call)
#define Perl_amagic_deref_call(...) GOPERL_LONG_CALL(amagic_deref_call, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(apply_attrs_string) && !defined(Perl_apply_attrs_string)
#define Perl_apply_attrs_string(...) GOPERL_LONG_CALL(apply_attrs_string, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(apply_builtin_cv_attributes) && !defined(Perl_apply_builtin_cv_attributes)
#define Perl_apply_builtin_cv_attributes(...) GOPERL_LONG_CALL(apply_builtin_cv_attributes, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(av_clear) && !defined(Perl_av_clear)
#define Perl_av_clear(...) GOPERL_LONG_CALL(av_clear, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(av_count) && !defined(Perl_av_count)
#define Perl_av_count(...) GOPERL_LONG_CALL(av_count, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(av_create_and_push) && !defined(Perl_av_create_and_push)
#define Perl_av_create_and_push(...) GOPERL_LONG_CALL(av_create_and_push, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(av_create_and_unshift_one) && !defined(Perl_av_create_and_unshift_one)
#define Perl_av_create_and_unshift_one(...) GOPERL_LONG_CALL(av_create_and_unshift_one, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(av_delete) && !defined(Perl_av_delete)
#define Perl_av_delete(...) GOPERL_LONG_CALL(av_delete, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(av_dump) && !defined(Perl_av_dump)
#define Perl_av_dump(...) GOPERL_LONG_CALL(av_dump, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(av_exists) && !defined(Perl_av_exists)
#define Perl_av_exists(...) GOPERL_LONG_CALL(av_exists, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(av_extend) && !defined(Perl_av_extend)
#define Perl_av_extend(...) GOPERL_LONG_CALL(av_extend, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(av_fetch) && !defined(Perl_av_fetch)
#define Perl_av_fetch(...) GOPERL_LONG_CALL(av_fetch, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(av_fill) && !defined(Perl_av_fill)
#define Perl_av_fill(...) GOPERL_LONG_CALL(av_fill, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(av_len) && !defined(Perl_av_len)
#define Perl_av_len(...) GOPERL_LONG_CALL(av_len, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(av_make) && !defined(Perl_av_make)
#define Perl_av_make(...) GOPERL_LONG_CALL(av_make, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(av_pop) && !defined(Perl_av_pop)
#define Perl_av_pop(...) GOPERL_LONG_CALL(av_pop, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(av_push) && !defined(Perl_av_push)
#define Perl_av_push(...) GOPERL_LONG_CALL(av_push, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(av_push_simple) && !defined(Perl_av_push_simple)
#define Perl_av_push_simple(...) GOPERL_LONG_CALL(av_push_simple, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(av_shift) && !defined(Perl_av_shift)
#define Perl_av_shift(...) GOPERL_LONG_CALL(av_shift, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(av_store) && !defined(Perl_av_store)
#define Perl_av_store(...) GOPERL_LONG_CALL(av_store, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(av_undef) && !defined(Perl_av_undef)
#define Perl_av_undef(...) GOPERL_LONG_CALL(av_undef, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(av_unshift) && !defined(Perl_av_unshift)
#define Perl_av_unshift(...) GOPERL_LONG_CALL(av_unshift, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(bytes_cmp_utf8) && !defined(Perl_bytes_cmp_utf8)
#define Perl_bytes_cmp_utf8(...) GOPERL_LONG_CALL(bytes_cmp_utf8, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(bytes_from_utf8) && !defined(Perl_bytes_from_utf8)
#define Perl_bytes_from_utf8(...) GOPERL_LONG_CALL(bytes_from_utf8, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(bytes_to_utf8) && !defined(Perl_bytes_to_utf8)
#define Perl_bytes_to_utf8(...) GOPERL_LONG_CALL(bytes_to_utf8, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(bytes_to_utf8_free_me) && !defined(Perl_bytes_to_utf8_free_me)
#define Perl_bytes_to_utf8_free_me(...) GOPERL_LONG_CALL(bytes_to_utf8_free_me, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(bytes_to_utf8_temp_pv) && !defined(Perl_bytes_to_utf8_temp_pv)
#define Perl_bytes_to_utf8_temp_pv(...) GOPERL_LONG_CALL(bytes_to_utf8_temp_pv, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(c9strict_utf8_to_uv) && !defined(Perl_c9strict_utf8_to_uv)
#define Perl_c9strict_utf8_to_uv(...) GOPERL_LONG_CALL(c9strict_utf8_to_uv, __VA_ARGS__)
#endif
#if defined(call_argv) && !defined(Perl_call_argv)
#define Perl_call_argv(...) GOPERL_LONG_CALL(call_argv, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(call_method) && !defined(Perl_call_method)
#define Perl_call_method(...) GOPERL_LONG_CALL(call_method, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(call_pv) && !defined(Perl_call_pv)
#define Perl_call_pv(...) GOPERL_LONG_CALL(call_pv, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(caller_cx) && !defined(Perl_caller_cx)
#define Perl_caller_cx(...) GOPERL_LONG_CALL(caller_cx, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(clone_params_del) && !defined(Perl_clone_params_del)
#define Perl_clone_params_del(...) GOPERL_LONG_CALL(clone_params_del, __VA_ARGS__)
#endif
#if defined(clone_params_new) && !defined(Perl_clone_params_new)
#define Perl_clone_params_new(...) GOPERL_LONG_CALL(clone_params_new, __VA_ARGS__)
#endif
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
#if defined(croak_sv) && !defined(Perl_croak_sv)
#define Perl_croak_sv(...) GOPERL_LONG_CALL(croak_sv, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(croak_xs_usage) && !defined(Perl_croak_xs_usage)
#define Perl_croak_xs_usage(...) GOPERL_LONG_CALL(croak_xs_usage, __VA_ARGS__)
#endif
#if defined(cv_const_sv) && !defined(Perl_cv_const_sv)
#define Perl_cv_const_sv(...) GOPERL_LONG_CALL(cv_const_sv, __VA_ARGS__)
#endif
#if defined(cv_get_call_checker) && !defined(Perl_cv_get_call_checker)
#define Perl_cv_get_call_checker(...) GOPERL_LONG_CALL(cv_get_call_checker, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(cv_get_call_checker_flags) && !defined(Perl_cv_get_call_checker_flags)
#define Perl_cv_get_call_checker_flags(...) GOPERL_LONG_CALL(cv_get_call_checker_flags, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(cv_name) && !defined(Perl_cv_name)
#define Perl_cv_name(...) GOPERL_LONG_CALL(cv_name, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(cv_set_call_checker) && !defined(Perl_cv_set_call_checker)
#define Perl_cv_set_call_checker(...) GOPERL_LONG_CALL(cv_set_call_checker, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(cv_set_call_checker_flags) && !defined(Perl_cv_set_call_checker_flags)
#define Perl_cv_set_call_checker_flags(...) GOPERL_LONG_CALL(cv_set_call_checker_flags, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(cv_undef) && !defined(Perl_cv_undef)
#define Perl_cv_undef(...) GOPERL_LONG_CALL(cv_undef, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(deb_nocontext) && !defined(Perl_deb_nocontext)
#define Perl_deb_nocontext(...) GOPERL_LONG_CALL(deb_nocontext, __VA_ARGS__)
#endif
#if defined(debstack) && !defined(Perl_debstack)
#define Perl_debstack(...) GOPERL_LONG_CALL(debstack, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(delimcpy) && !defined(Perl_delimcpy)
#define Perl_delimcpy(...) GOPERL_LONG_CALL(delimcpy, __VA_ARGS__)
#endif
#if defined(die_nocontext) && !defined(Perl_die_nocontext)
#define Perl_die_nocontext(...) GOPERL_LONG_CALL(die_nocontext, __VA_ARGS__)
#endif
#if defined(die_sv) && !defined(Perl_die_sv)
#define Perl_die_sv(...) GOPERL_LONG_CALL(die_sv, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(do_close) && !defined(Perl_do_close)
#define Perl_do_close(...) GOPERL_LONG_CALL(do_close, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(do_join) && !defined(Perl_do_join)
#define Perl_do_join(...) GOPERL_LONG_CALL(do_join, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(do_open) && !defined(Perl_do_open)
#define Perl_do_open(...) GOPERL_LONG_CALL(do_open, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(do_openn) && !defined(Perl_do_openn)
#define Perl_do_openn(...) GOPERL_LONG_CALL(do_openn, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(do_sprintf) && !defined(Perl_do_sprintf)
#define Perl_do_sprintf(...) GOPERL_LONG_CALL(do_sprintf, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(dump_all) && !defined(Perl_dump_all)
#define Perl_dump_all(...) GOPERL_LONG_CALL(dump_all, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(dump_c_backtrace) && !defined(Perl_dump_c_backtrace)
#define Perl_dump_c_backtrace(...) GOPERL_LONG_CALL(dump_c_backtrace, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(dump_eval) && !defined(Perl_dump_eval)
#define Perl_dump_eval(...) GOPERL_LONG_CALL(dump_eval, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(dump_form) && !defined(Perl_dump_form)
#define Perl_dump_form(...) GOPERL_LONG_CALL(dump_form, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(dump_mstats) && !defined(Perl_dump_mstats)
#define Perl_dump_mstats(...) GOPERL_LONG_CALL(dump_mstats, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(dump_packsubs) && !defined(Perl_dump_packsubs)
#define Perl_dump_packsubs(...) GOPERL_LONG_CALL(dump_packsubs, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(dump_sub) && !defined(Perl_dump_sub)
#define Perl_dump_sub(...) GOPERL_LONG_CALL(dump_sub, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(eval_pv) && !defined(Perl_eval_pv)
#define Perl_eval_pv(...) GOPERL_LONG_CALL(eval_pv, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(eval_sv) && !defined(Perl_eval_sv)
#define Perl_eval_sv(...) GOPERL_LONG_CALL(eval_sv, GOPERL_THX_DROP(__VA_ARGS__))
#endif
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
#if defined(filter_read) && !defined(Perl_filter_read)
#define Perl_filter_read(...) GOPERL_LONG_CALL(filter_read, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(finalize_optree) && !defined(Perl_finalize_optree)
#define Perl_finalize_optree(...) GOPERL_LONG_CALL(finalize_optree, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(find_runcv) && !defined(Perl_find_runcv)
#define Perl_find_runcv(...) GOPERL_LONG_CALL(find_runcv, GOPERL_THX_DROP(__VA_ARGS__))
#endif
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
#if defined(get_av) && !defined(Perl_get_av)
#define Perl_get_av(...) GOPERL_LONG_CALL(get_av, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(get_c_backtrace_dump) && !defined(Perl_get_c_backtrace_dump)
#define Perl_get_c_backtrace_dump(...) GOPERL_LONG_CALL(get_c_backtrace_dump, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(get_cv) && !defined(Perl_get_cv)
#define Perl_get_cv(...) GOPERL_LONG_CALL(get_cv, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(get_cvn_flags) && !defined(Perl_get_cvn_flags)
#define Perl_get_cvn_flags(...) GOPERL_LONG_CALL(get_cvn_flags, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(get_hv) && !defined(Perl_get_hv)
#define Perl_get_hv(...) GOPERL_LONG_CALL(get_hv, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(get_op_descs) && !defined(Perl_get_op_descs)
#define Perl_get_op_descs(...) GOPERL_LONG_CALL(get_op_descs, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(get_op_names) && !defined(Perl_get_op_names)
#define Perl_get_op_names(...) GOPERL_LONG_CALL(get_op_names, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(get_sv) && !defined(Perl_get_sv)
#define Perl_get_sv(...) GOPERL_LONG_CALL(get_sv, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(getcwd_sv) && !defined(Perl_getcwd_sv)
#define Perl_getcwd_sv(...) GOPERL_LONG_CALL(getcwd_sv, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(grok_atoUV) && !defined(Perl_grok_atoUV)
#define Perl_grok_atoUV(...) GOPERL_LONG_CALL(grok_atoUV, __VA_ARGS__)
#endif
#if defined(grok_bin) && !defined(Perl_grok_bin)
#define Perl_grok_bin(...) GOPERL_LONG_CALL(grok_bin, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(grok_hex) && !defined(Perl_grok_hex)
#define Perl_grok_hex(...) GOPERL_LONG_CALL(grok_hex, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(grok_infnan) && !defined(Perl_grok_infnan)
#define Perl_grok_infnan(...) GOPERL_LONG_CALL(grok_infnan, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(grok_number) && !defined(Perl_grok_number)
#define Perl_grok_number(...) GOPERL_LONG_CALL(grok_number, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(grok_number_flags) && !defined(Perl_grok_number_flags)
#define Perl_grok_number_flags(...) GOPERL_LONG_CALL(grok_number_flags, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(grok_numeric_radix) && !defined(Perl_grok_numeric_radix)
#define Perl_grok_numeric_radix(...) GOPERL_LONG_CALL(grok_numeric_radix, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(grok_oct) && !defined(Perl_grok_oct)
#define Perl_grok_oct(...) GOPERL_LONG_CALL(grok_oct, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(gv_AVadd) && !defined(Perl_gv_AVadd)
#define Perl_gv_AVadd(...) GOPERL_LONG_CALL(gv_AVadd, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(gv_HVadd) && !defined(Perl_gv_HVadd)
#define Perl_gv_HVadd(...) GOPERL_LONG_CALL(gv_HVadd, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(gv_IOadd) && !defined(Perl_gv_IOadd)
#define Perl_gv_IOadd(...) GOPERL_LONG_CALL(gv_IOadd, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(gv_SVadd) && !defined(Perl_gv_SVadd)
#define Perl_gv_SVadd(...) GOPERL_LONG_CALL(gv_SVadd, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(gv_add_by_type) && !defined(Perl_gv_add_by_type)
#define Perl_gv_add_by_type(...) GOPERL_LONG_CALL(gv_add_by_type, GOPERL_THX_DROP(__VA_ARGS__))
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
#if defined(gv_const_sv) && !defined(Perl_gv_const_sv)
#define Perl_gv_const_sv(...) GOPERL_LONG_CALL(gv_const_sv, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(gv_dump) && !defined(Perl_gv_dump)
#define Perl_gv_dump(...) GOPERL_LONG_CALL(gv_dump, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(gv_efullname3) && !defined(Perl_gv_efullname3)
#define Perl_gv_efullname3(...) GOPERL_LONG_CALL(gv_efullname3, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(gv_efullname4) && !defined(Perl_gv_efullname4)
#define Perl_gv_efullname4(...) GOPERL_LONG_CALL(gv_efullname4, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(gv_fetchfile) && !defined(Perl_gv_fetchfile)
#define Perl_gv_fetchfile(...) GOPERL_LONG_CALL(gv_fetchfile, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(gv_fetchfile_flags) && !defined(Perl_gv_fetchfile_flags)
#define Perl_gv_fetchfile_flags(...) GOPERL_LONG_CALL(gv_fetchfile_flags, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(gv_fetchmeth) && !defined(Perl_gv_fetchmeth)
#define Perl_gv_fetchmeth(...) GOPERL_LONG_CALL(gv_fetchmeth, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(gv_fetchmeth_autoload) && !defined(Perl_gv_fetchmeth_autoload)
#define Perl_gv_fetchmeth_autoload(...) GOPERL_LONG_CALL(gv_fetchmeth_autoload, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(gv_fetchmeth_pv) && !defined(Perl_gv_fetchmeth_pv)
#define Perl_gv_fetchmeth_pv(...) GOPERL_LONG_CALL(gv_fetchmeth_pv, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(gv_fetchmeth_pv_autoload) && !defined(Perl_gv_fetchmeth_pv_autoload)
#define Perl_gv_fetchmeth_pv_autoload(...) GOPERL_LONG_CALL(gv_fetchmeth_pv_autoload, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(gv_fetchmeth_pvn) && !defined(Perl_gv_fetchmeth_pvn)
#define Perl_gv_fetchmeth_pvn(...) GOPERL_LONG_CALL(gv_fetchmeth_pvn, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(gv_fetchmeth_pvn_autoload) && !defined(Perl_gv_fetchmeth_pvn_autoload)
#define Perl_gv_fetchmeth_pvn_autoload(...) GOPERL_LONG_CALL(gv_fetchmeth_pvn_autoload, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(gv_fetchmeth_sv) && !defined(Perl_gv_fetchmeth_sv)
#define Perl_gv_fetchmeth_sv(...) GOPERL_LONG_CALL(gv_fetchmeth_sv, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(gv_fetchmeth_sv_autoload) && !defined(Perl_gv_fetchmeth_sv_autoload)
#define Perl_gv_fetchmeth_sv_autoload(...) GOPERL_LONG_CALL(gv_fetchmeth_sv_autoload, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(gv_fetchmethod) && !defined(Perl_gv_fetchmethod)
#define Perl_gv_fetchmethod(...) GOPERL_LONG_CALL(gv_fetchmethod, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(gv_fetchmethod_autoload) && !defined(Perl_gv_fetchmethod_autoload)
#define Perl_gv_fetchmethod_autoload(...) GOPERL_LONG_CALL(gv_fetchmethod_autoload, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(gv_fetchmethod_pv_flags) && !defined(Perl_gv_fetchmethod_pv_flags)
#define Perl_gv_fetchmethod_pv_flags(...) GOPERL_LONG_CALL(gv_fetchmethod_pv_flags, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(gv_fetchmethod_pvn_flags) && !defined(Perl_gv_fetchmethod_pvn_flags)
#define Perl_gv_fetchmethod_pvn_flags(...) GOPERL_LONG_CALL(gv_fetchmethod_pvn_flags, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(gv_fetchmethod_sv_flags) && !defined(Perl_gv_fetchmethod_sv_flags)
#define Perl_gv_fetchmethod_sv_flags(...) GOPERL_LONG_CALL(gv_fetchmethod_sv_flags, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(gv_fetchpv) && !defined(Perl_gv_fetchpv)
#define Perl_gv_fetchpv(...) GOPERL_LONG_CALL(gv_fetchpv, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(gv_fetchpvn_flags) && !defined(Perl_gv_fetchpvn_flags)
#define Perl_gv_fetchpvn_flags(...) GOPERL_LONG_CALL(gv_fetchpvn_flags, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(gv_fetchsv) && !defined(Perl_gv_fetchsv)
#define Perl_gv_fetchsv(...) GOPERL_LONG_CALL(gv_fetchsv, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(gv_fullname3) && !defined(Perl_gv_fullname3)
#define Perl_gv_fullname3(...) GOPERL_LONG_CALL(gv_fullname3, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(gv_fullname4) && !defined(Perl_gv_fullname4)
#define Perl_gv_fullname4(...) GOPERL_LONG_CALL(gv_fullname4, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(gv_init) && !defined(Perl_gv_init)
#define Perl_gv_init(...) GOPERL_LONG_CALL(gv_init, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(gv_init_pv) && !defined(Perl_gv_init_pv)
#define Perl_gv_init_pv(...) GOPERL_LONG_CALL(gv_init_pv, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(gv_init_pvn) && !defined(Perl_gv_init_pvn)
#define Perl_gv_init_pvn(...) GOPERL_LONG_CALL(gv_init_pvn, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(gv_init_sv) && !defined(Perl_gv_init_sv)
#define Perl_gv_init_sv(...) GOPERL_LONG_CALL(gv_init_sv, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(gv_name_set) && !defined(Perl_gv_name_set)
#define Perl_gv_name_set(...) GOPERL_LONG_CALL(gv_name_set, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(gv_stashpv) && !defined(Perl_gv_stashpv)
#define Perl_gv_stashpv(...) GOPERL_LONG_CALL(gv_stashpv, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(gv_stashpvn) && !defined(Perl_gv_stashpvn)
#define Perl_gv_stashpvn(...) GOPERL_LONG_CALL(gv_stashpvn, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(gv_stashsv) && !defined(Perl_gv_stashsv)
#define Perl_gv_stashsv(...) GOPERL_LONG_CALL(gv_stashsv, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(hv_assert) && !defined(Perl_hv_assert)
#define Perl_hv_assert(...) GOPERL_LONG_CALL(hv_assert, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(hv_bucket_ratio) && !defined(Perl_hv_bucket_ratio)
#define Perl_hv_bucket_ratio(...) GOPERL_LONG_CALL(hv_bucket_ratio, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(hv_clear) && !defined(Perl_hv_clear)
#define Perl_hv_clear(...) GOPERL_LONG_CALL(hv_clear, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(hv_clear_placeholders) && !defined(Perl_hv_clear_placeholders)
#define Perl_hv_clear_placeholders(...) GOPERL_LONG_CALL(hv_clear_placeholders, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(hv_copy_hints_hv) && !defined(Perl_hv_copy_hints_hv)
#define Perl_hv_copy_hints_hv(...) GOPERL_LONG_CALL(hv_copy_hints_hv, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(hv_delete) && !defined(Perl_hv_delete)
#define Perl_hv_delete(...) GOPERL_LONG_CALL(hv_delete, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(hv_delete_ent) && !defined(Perl_hv_delete_ent)
#define Perl_hv_delete_ent(...) GOPERL_LONG_CALL(hv_delete_ent, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(hv_dump) && !defined(Perl_hv_dump)
#define Perl_hv_dump(...) GOPERL_LONG_CALL(hv_dump, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(hv_exists) && !defined(Perl_hv_exists)
#define Perl_hv_exists(...) GOPERL_LONG_CALL(hv_exists, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(hv_exists_ent) && !defined(Perl_hv_exists_ent)
#define Perl_hv_exists_ent(...) GOPERL_LONG_CALL(hv_exists_ent, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(hv_fetch) && !defined(Perl_hv_fetch)
#define Perl_hv_fetch(...) GOPERL_LONG_CALL(hv_fetch, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(hv_fetch_ent) && !defined(Perl_hv_fetch_ent)
#define Perl_hv_fetch_ent(...) GOPERL_LONG_CALL(hv_fetch_ent, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(hv_iterinit) && !defined(Perl_hv_iterinit)
#define Perl_hv_iterinit(...) GOPERL_LONG_CALL(hv_iterinit, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(hv_iterkey) && !defined(Perl_hv_iterkey)
#define Perl_hv_iterkey(...) GOPERL_LONG_CALL(hv_iterkey, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(hv_iterkeysv) && !defined(Perl_hv_iterkeysv)
#define Perl_hv_iterkeysv(...) GOPERL_LONG_CALL(hv_iterkeysv, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(hv_iternext) && !defined(Perl_hv_iternext)
#define Perl_hv_iternext(...) GOPERL_LONG_CALL(hv_iternext, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(hv_iternext_flags) && !defined(Perl_hv_iternext_flags)
#define Perl_hv_iternext_flags(...) GOPERL_LONG_CALL(hv_iternext_flags, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(hv_iternextsv) && !defined(Perl_hv_iternextsv)
#define Perl_hv_iternextsv(...) GOPERL_LONG_CALL(hv_iternextsv, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(hv_iterval) && !defined(Perl_hv_iterval)
#define Perl_hv_iterval(...) GOPERL_LONG_CALL(hv_iterval, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(hv_ksplit) && !defined(Perl_hv_ksplit)
#define Perl_hv_ksplit(...) GOPERL_LONG_CALL(hv_ksplit, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(hv_magic) && !defined(Perl_hv_magic)
#define Perl_hv_magic(...) GOPERL_LONG_CALL(hv_magic, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(hv_name_set) && !defined(Perl_hv_name_set)
#define Perl_hv_name_set(...) GOPERL_LONG_CALL(hv_name_set, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(hv_scalar) && !defined(Perl_hv_scalar)
#define Perl_hv_scalar(...) GOPERL_LONG_CALL(hv_scalar, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(hv_store) && !defined(Perl_hv_store)
#define Perl_hv_store(...) GOPERL_LONG_CALL(hv_store, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(hv_store_ent) && !defined(Perl_hv_store_ent)
#define Perl_hv_store_ent(...) GOPERL_LONG_CALL(hv_store_ent, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(hv_store_flags) && !defined(Perl_hv_store_flags)
#define Perl_hv_store_flags(...) GOPERL_LONG_CALL(hv_store_flags, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(hv_undef) && !defined(Perl_hv_undef)
#define Perl_hv_undef(...) GOPERL_LONG_CALL(hv_undef, GOPERL_THX_DROP(__VA_ARGS__))
#endif
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
#if defined(isSTRICT_UTF8_CHAR) && !defined(Perl_isSTRICT_UTF8_CHAR)
#define Perl_isSTRICT_UTF8_CHAR(...) GOPERL_LONG_CALL(isSTRICT_UTF8_CHAR, __VA_ARGS__)
#endif
#if defined(isUTF8_CHAR) && !defined(Perl_isUTF8_CHAR)
#define Perl_isUTF8_CHAR(...) GOPERL_LONG_CALL(isUTF8_CHAR, __VA_ARGS__)
#endif
#if defined(isUTF8_CHAR_flags) && !defined(Perl_isUTF8_CHAR_flags)
#define Perl_isUTF8_CHAR_flags(...) GOPERL_LONG_CALL(isUTF8_CHAR_flags, __VA_ARGS__)
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
#if defined(is_lvalue_sub) && !defined(Perl_is_lvalue_sub)
#define Perl_is_lvalue_sub(...) GOPERL_LONG_CALL(is_lvalue_sub, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(is_safe_syscall) && !defined(Perl_is_safe_syscall)
#define Perl_is_safe_syscall(...) GOPERL_LONG_CALL(is_safe_syscall, GOPERL_THX_DROP(__VA_ARGS__))
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
#if defined(is_utf8_char_buf) && !defined(Perl_is_utf8_char_buf)
#define Perl_is_utf8_char_buf(...) GOPERL_LONG_CALL(is_utf8_char_buf, __VA_ARGS__)
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
#if defined(is_utf8_string) && !defined(Perl_is_utf8_string)
#define Perl_is_utf8_string(...) GOPERL_LONG_CALL(is_utf8_string, __VA_ARGS__)
#endif
#if defined(is_utf8_string_flags) && !defined(Perl_is_utf8_string_flags)
#define Perl_is_utf8_string_flags(...) GOPERL_LONG_CALL(is_utf8_string_flags, __VA_ARGS__)
#endif
#if defined(is_utf8_string_loc) && !defined(Perl_is_utf8_string_loc)
#define Perl_is_utf8_string_loc(...) GOPERL_LONG_CALL(is_utf8_string_loc, __VA_ARGS__)
#endif
#if defined(is_utf8_string_loc_flags) && !defined(Perl_is_utf8_string_loc_flags)
#define Perl_is_utf8_string_loc_flags(...) GOPERL_LONG_CALL(is_utf8_string_loc_flags, __VA_ARGS__)
#endif
#if defined(is_utf8_string_loclen) && !defined(Perl_is_utf8_string_loclen)
#define Perl_is_utf8_string_loclen(...) GOPERL_LONG_CALL(is_utf8_string_loclen, __VA_ARGS__)
#endif
#if defined(is_utf8_string_loclen_flags) && !defined(Perl_is_utf8_string_loclen_flags)
#define Perl_is_utf8_string_loclen_flags(...) GOPERL_LONG_CALL(is_utf8_string_loclen_flags, __VA_ARGS__)
#endif
#if defined(is_utf8_valid_partial_char) && !defined(Perl_is_utf8_valid_partial_char)
#define Perl_is_utf8_valid_partial_char(...) GOPERL_LONG_CALL(is_utf8_valid_partial_char, __VA_ARGS__)
#endif
#if defined(is_utf8_valid_partial_char_flags) && !defined(Perl_is_utf8_valid_partial_char_flags)
#define Perl_is_utf8_valid_partial_char_flags(...) GOPERL_LONG_CALL(is_utf8_valid_partial_char_flags, __VA_ARGS__)
#endif
#if defined(isinfnan) && !defined(Perl_isinfnan)
#define Perl_isinfnan(...) GOPERL_LONG_CALL(isinfnan, __VA_ARGS__)
#endif
#if defined(leave_adjust_stacks) && !defined(Perl_leave_adjust_stacks)
#define Perl_leave_adjust_stacks(...) GOPERL_LONG_CALL(leave_adjust_stacks, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(lex_bufutf8) && !defined(Perl_lex_bufutf8)
#define Perl_lex_bufutf8(...) GOPERL_LONG_CALL(lex_bufutf8, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(lex_discard_to) && !defined(Perl_lex_discard_to)
#define Perl_lex_discard_to(...) GOPERL_LONG_CALL(lex_discard_to, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(lex_grow_linestr) && !defined(Perl_lex_grow_linestr)
#define Perl_lex_grow_linestr(...) GOPERL_LONG_CALL(lex_grow_linestr, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(lex_next_chunk) && !defined(Perl_lex_next_chunk)
#define Perl_lex_next_chunk(...) GOPERL_LONG_CALL(lex_next_chunk, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(lex_peek_unichar) && !defined(Perl_lex_peek_unichar)
#define Perl_lex_peek_unichar(...) GOPERL_LONG_CALL(lex_peek_unichar, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(lex_read_space) && !defined(Perl_lex_read_space)
#define Perl_lex_read_space(...) GOPERL_LONG_CALL(lex_read_space, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(lex_read_to) && !defined(Perl_lex_read_to)
#define Perl_lex_read_to(...) GOPERL_LONG_CALL(lex_read_to, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(lex_read_unichar) && !defined(Perl_lex_read_unichar)
#define Perl_lex_read_unichar(...) GOPERL_LONG_CALL(lex_read_unichar, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(lex_start) && !defined(Perl_lex_start)
#define Perl_lex_start(...) GOPERL_LONG_CALL(lex_start, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(lex_stuff_pv) && !defined(Perl_lex_stuff_pv)
#define Perl_lex_stuff_pv(...) GOPERL_LONG_CALL(lex_stuff_pv, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(lex_stuff_pvn) && !defined(Perl_lex_stuff_pvn)
#define Perl_lex_stuff_pvn(...) GOPERL_LONG_CALL(lex_stuff_pvn, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(lex_stuff_sv) && !defined(Perl_lex_stuff_sv)
#define Perl_lex_stuff_sv(...) GOPERL_LONG_CALL(lex_stuff_sv, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(load_module_nocontext) && !defined(Perl_load_module_nocontext)
#define Perl_load_module_nocontext(...) GOPERL_LONG_CALL(load_module_nocontext, __VA_ARGS__)
#endif
#if defined(looks_like_number) && !defined(Perl_looks_like_number)
#define Perl_looks_like_number(...) GOPERL_LONG_CALL(looks_like_number, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(magic_dump) && !defined(Perl_magic_dump)
#define Perl_magic_dump(...) GOPERL_LONG_CALL(magic_dump, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(mess) && !defined(Perl_mess)
#define Perl_mess(...) GOPERL_LONG_CALL(mess, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(mess_nocontext) && !defined(Perl_mess_nocontext)
#define Perl_mess_nocontext(...) GOPERL_LONG_CALL(mess_nocontext, __VA_ARGS__)
#endif
#if defined(mess_sv) && !defined(Perl_mess_sv)
#define Perl_mess_sv(...) GOPERL_LONG_CALL(mess_sv, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(mg_clear) && !defined(Perl_mg_clear)
#define Perl_mg_clear(...) GOPERL_LONG_CALL(mg_clear, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(mg_copy) && !defined(Perl_mg_copy)
#define Perl_mg_copy(...) GOPERL_LONG_CALL(mg_copy, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(mg_find) && !defined(Perl_mg_find)
#define Perl_mg_find(...) GOPERL_LONG_CALL(mg_find, __VA_ARGS__)
#endif
#if defined(mg_findext) && !defined(Perl_mg_findext)
#define Perl_mg_findext(...) GOPERL_LONG_CALL(mg_findext, __VA_ARGS__)
#endif
#if defined(mg_free) && !defined(Perl_mg_free)
#define Perl_mg_free(...) GOPERL_LONG_CALL(mg_free, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(mg_free_type) && !defined(Perl_mg_free_type)
#define Perl_mg_free_type(...) GOPERL_LONG_CALL(mg_free_type, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(mg_freeext) && !defined(Perl_mg_freeext)
#define Perl_mg_freeext(...) GOPERL_LONG_CALL(mg_freeext, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(mg_get) && !defined(Perl_mg_get)
#define Perl_mg_get(...) GOPERL_LONG_CALL(mg_get, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(mg_magical) && !defined(Perl_mg_magical)
#define Perl_mg_magical(...) GOPERL_LONG_CALL(mg_magical, __VA_ARGS__)
#endif
#if defined(mg_set) && !defined(Perl_mg_set)
#define Perl_mg_set(...) GOPERL_LONG_CALL(mg_set, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(mini_mktime) && !defined(Perl_mini_mktime)
#define Perl_mini_mktime(...) GOPERL_LONG_CALL(mini_mktime, __VA_ARGS__)
#endif
#if defined(mortal_destructor_sv) && !defined(Perl_mortal_destructor_sv)
#define Perl_mortal_destructor_sv(...) GOPERL_LONG_CALL(mortal_destructor_sv, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(mro_get_from_name) && !defined(Perl_mro_get_from_name)
#define Perl_mro_get_from_name(...) GOPERL_LONG_CALL(mro_get_from_name, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(mro_get_linear_isa) && !defined(Perl_mro_get_linear_isa)
#define Perl_mro_get_linear_isa(...) GOPERL_LONG_CALL(mro_get_linear_isa, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(mro_method_changed_in) && !defined(Perl_mro_method_changed_in)
#define Perl_mro_method_changed_in(...) GOPERL_LONG_CALL(mro_method_changed_in, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(mro_register) && !defined(Perl_mro_register)
#define Perl_mro_register(...) GOPERL_LONG_CALL(mro_register, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(mro_set_mro) && !defined(Perl_mro_set_mro)
#define Perl_mro_set_mro(...) GOPERL_LONG_CALL(mro_set_mro, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(mro_set_private_data) && !defined(Perl_mro_set_private_data)
#define Perl_mro_set_private_data(...) GOPERL_LONG_CALL(mro_set_private_data, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(my_atof) && !defined(Perl_my_atof)
#define Perl_my_atof(...) GOPERL_LONG_CALL(my_atof, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(my_chsize) && !defined(Perl_my_chsize)
#define Perl_my_chsize(...) GOPERL_LONG_CALL(my_chsize, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(my_dirfd) && !defined(Perl_my_dirfd)
#define Perl_my_dirfd(...) GOPERL_LONG_CALL(my_dirfd, __VA_ARGS__)
#endif
#if defined(my_exit) && !defined(Perl_my_exit)
#define Perl_my_exit(...) GOPERL_LONG_CALL(my_exit, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(my_failure_exit) && !defined(Perl_my_failure_exit)
#define Perl_my_failure_exit(...) GOPERL_LONG_CALL(my_failure_exit, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(my_pclose) && !defined(Perl_my_pclose)
#define Perl_my_pclose(...) GOPERL_LONG_CALL(my_pclose, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(my_popen) && !defined(Perl_my_popen)
#define Perl_my_popen(...) GOPERL_LONG_CALL(my_popen, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(my_setenv) && !defined(Perl_my_setenv)
#define Perl_my_setenv(...) GOPERL_LONG_CALL(my_setenv, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(my_snprintf) && !defined(Perl_my_snprintf)
#define Perl_my_snprintf(...) GOPERL_LONG_CALL(my_snprintf, __VA_ARGS__)
#endif
#if defined(my_strftime) && !defined(Perl_my_strftime)
#define Perl_my_strftime(...) GOPERL_LONG_CALL(my_strftime, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(my_strlcat) && !defined(Perl_my_strlcat)
#define Perl_my_strlcat(...) GOPERL_LONG_CALL(my_strlcat, __VA_ARGS__)
#endif
#if defined(my_strnlen) && !defined(Perl_my_strnlen)
#define Perl_my_strnlen(...) GOPERL_LONG_CALL(my_strnlen, __VA_ARGS__)
#endif
#if defined(my_strtod) && !defined(Perl_my_strtod)
#define Perl_my_strtod(...) GOPERL_LONG_CALL(my_strtod, __VA_ARGS__)
#endif
#if defined(my_vsnprintf) && !defined(Perl_my_vsnprintf)
#define Perl_my_vsnprintf(...) GOPERL_LONG_CALL(my_vsnprintf, __VA_ARGS__)
#endif
#if defined(newANONATTRSUB) && !defined(Perl_newANONATTRSUB)
#define Perl_newANONATTRSUB(...) GOPERL_LONG_CALL(newANONATTRSUB, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(newANONHASH) && !defined(Perl_newANONHASH)
#define Perl_newANONHASH(...) GOPERL_LONG_CALL(newANONHASH, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(newANONLIST) && !defined(Perl_newANONLIST)
#define Perl_newANONLIST(...) GOPERL_LONG_CALL(newANONLIST, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(newANONSUB) && !defined(Perl_newANONSUB)
#define Perl_newANONSUB(...) GOPERL_LONG_CALL(newANONSUB, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(newARGDEFELEMOP) && !defined(Perl_newARGDEFELEMOP)
#define Perl_newARGDEFELEMOP(...) GOPERL_LONG_CALL(newARGDEFELEMOP, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(newASSIGNOP) && !defined(Perl_newASSIGNOP)
#define Perl_newASSIGNOP(...) GOPERL_LONG_CALL(newASSIGNOP, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(newATTRSUB) && !defined(Perl_newATTRSUB)
#define Perl_newATTRSUB(...) GOPERL_LONG_CALL(newATTRSUB, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(newAV) && !defined(Perl_newAV)
#define Perl_newAV(...) GOPERL_LONG_CALL(newAV, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(newAVREF) && !defined(Perl_newAVREF)
#define Perl_newAVREF(...) GOPERL_LONG_CALL(newAVREF, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(newAV_alloc_x) && !defined(Perl_newAV_alloc_x)
#define Perl_newAV_alloc_x(...) GOPERL_LONG_CALL(newAV_alloc_x, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(newAV_alloc_xz) && !defined(Perl_newAV_alloc_xz)
#define Perl_newAV_alloc_xz(...) GOPERL_LONG_CALL(newAV_alloc_xz, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(newAV_mortal) && !defined(Perl_newAV_mortal)
#define Perl_newAV_mortal(...) GOPERL_LONG_CALL(newAV_mortal, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(newAVav) && !defined(Perl_newAVav)
#define Perl_newAVav(...) GOPERL_LONG_CALL(newAVav, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(newAVhv) && !defined(Perl_newAVhv)
#define Perl_newAVhv(...) GOPERL_LONG_CALL(newAVhv, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(newBINOP) && !defined(Perl_newBINOP)
#define Perl_newBINOP(...) GOPERL_LONG_CALL(newBINOP, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(newCONDOP) && !defined(Perl_newCONDOP)
#define Perl_newCONDOP(...) GOPERL_LONG_CALL(newCONDOP, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(newCONSTSUB) && !defined(Perl_newCONSTSUB)
#define Perl_newCONSTSUB(...) GOPERL_LONG_CALL(newCONSTSUB, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(newCONSTSUB_flags) && !defined(Perl_newCONSTSUB_flags)
#define Perl_newCONSTSUB_flags(...) GOPERL_LONG_CALL(newCONSTSUB_flags, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(newCVREF) && !defined(Perl_newCVREF)
#define Perl_newCVREF(...) GOPERL_LONG_CALL(newCVREF, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(newDEFEROP) && !defined(Perl_newDEFEROP)
#define Perl_newDEFEROP(...) GOPERL_LONG_CALL(newDEFEROP, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(newDEFSVOP) && !defined(Perl_newDEFSVOP)
#define Perl_newDEFSVOP(...) GOPERL_LONG_CALL(newDEFSVOP, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(newFOROP) && !defined(Perl_newFOROP)
#define Perl_newFOROP(...) GOPERL_LONG_CALL(newFOROP, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(newGIVENOP) && !defined(Perl_newGIVENOP)
#define Perl_newGIVENOP(...) GOPERL_LONG_CALL(newGIVENOP, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(newGVOP) && !defined(Perl_newGVOP)
#define Perl_newGVOP(...) GOPERL_LONG_CALL(newGVOP, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(newGVREF) && !defined(Perl_newGVREF)
#define Perl_newGVREF(...) GOPERL_LONG_CALL(newGVREF, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(newGVgen) && !defined(Perl_newGVgen)
#define Perl_newGVgen(...) GOPERL_LONG_CALL(newGVgen, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(newGVgen_flags) && !defined(Perl_newGVgen_flags)
#define Perl_newGVgen_flags(...) GOPERL_LONG_CALL(newGVgen_flags, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(newHV) && !defined(Perl_newHV)
#define Perl_newHV(...) GOPERL_LONG_CALL(newHV, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(newHVREF) && !defined(Perl_newHVREF)
#define Perl_newHVREF(...) GOPERL_LONG_CALL(newHVREF, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(newHVhv) && !defined(Perl_newHVhv)
#define Perl_newHVhv(...) GOPERL_LONG_CALL(newHVhv, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(newIO) && !defined(Perl_newIO)
#define Perl_newIO(...) GOPERL_LONG_CALL(newIO, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(newLISTOP) && !defined(Perl_newLISTOP)
#define Perl_newLISTOP(...) GOPERL_LONG_CALL(newLISTOP, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(newLISTOPn) && !defined(Perl_newLISTOPn)
#define Perl_newLISTOPn(...) GOPERL_LONG_CALL(newLISTOPn, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(newLOGOP) && !defined(Perl_newLOGOP)
#define Perl_newLOGOP(...) GOPERL_LONG_CALL(newLOGOP, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(newLOOPEX) && !defined(Perl_newLOOPEX)
#define Perl_newLOOPEX(...) GOPERL_LONG_CALL(newLOOPEX, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(newLOOPOP) && !defined(Perl_newLOOPOP)
#define Perl_newLOOPOP(...) GOPERL_LONG_CALL(newLOOPOP, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(newMETHOP) && !defined(Perl_newMETHOP)
#define Perl_newMETHOP(...) GOPERL_LONG_CALL(newMETHOP, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(newMETHOP_named) && !defined(Perl_newMETHOP_named)
#define Perl_newMETHOP_named(...) GOPERL_LONG_CALL(newMETHOP_named, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(newNULLLIST) && !defined(Perl_newNULLLIST)
#define Perl_newNULLLIST(...) GOPERL_LONG_CALL(newNULLLIST, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(newOP) && !defined(Perl_newOP)
#define Perl_newOP(...) GOPERL_LONG_CALL(newOP, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(newPADNAMELIST) && !defined(Perl_newPADNAMELIST)
#define Perl_newPADNAMELIST(...) GOPERL_LONG_CALL(newPADNAMELIST, __VA_ARGS__)
#endif
#if defined(newPADNAMEouter) && !defined(Perl_newPADNAMEouter)
#define Perl_newPADNAMEouter(...) GOPERL_LONG_CALL(newPADNAMEouter, __VA_ARGS__)
#endif
#if defined(newPADNAMEpvn) && !defined(Perl_newPADNAMEpvn)
#define Perl_newPADNAMEpvn(...) GOPERL_LONG_CALL(newPADNAMEpvn, __VA_ARGS__)
#endif
#if defined(newPADOP) && !defined(Perl_newPADOP)
#define Perl_newPADOP(...) GOPERL_LONG_CALL(newPADOP, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(newPADxVOP) && !defined(Perl_newPADxVOP)
#define Perl_newPADxVOP(...) GOPERL_LONG_CALL(newPADxVOP, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(newPMOP) && !defined(Perl_newPMOP)
#define Perl_newPMOP(...) GOPERL_LONG_CALL(newPMOP, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(newPVOP) && !defined(Perl_newPVOP)
#define Perl_newPVOP(...) GOPERL_LONG_CALL(newPVOP, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(newRANGE) && !defined(Perl_newRANGE)
#define Perl_newRANGE(...) GOPERL_LONG_CALL(newRANGE, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(newRV) && !defined(Perl_newRV)
#define Perl_newRV(...) GOPERL_LONG_CALL(newRV, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(newRV_noinc) && !defined(Perl_newRV_noinc)
#define Perl_newRV_noinc(...) GOPERL_LONG_CALL(newRV_noinc, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(newSLICEOP) && !defined(Perl_newSLICEOP)
#define Perl_newSLICEOP(...) GOPERL_LONG_CALL(newSLICEOP, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(newSTATEOP) && !defined(Perl_newSTATEOP)
#define Perl_newSTATEOP(...) GOPERL_LONG_CALL(newSTATEOP, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(newSUB) && !defined(Perl_newSUB)
#define Perl_newSUB(...) GOPERL_LONG_CALL(newSUB, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(newSV) && !defined(Perl_newSV)
#define Perl_newSV(...) GOPERL_LONG_CALL(newSV, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(newSVOP) && !defined(Perl_newSVOP)
#define Perl_newSVOP(...) GOPERL_LONG_CALL(newSVOP, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(newSVREF) && !defined(Perl_newSVREF)
#define Perl_newSVREF(...) GOPERL_LONG_CALL(newSVREF, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(newSV_false) && !defined(Perl_newSV_false)
#define Perl_newSV_false(...) GOPERL_LONG_CALL(newSV_false, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(newSV_true) && !defined(Perl_newSV_true)
#define Perl_newSV_true(...) GOPERL_LONG_CALL(newSV_true, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(newSV_type) && !defined(Perl_newSV_type)
#define Perl_newSV_type(...) GOPERL_LONG_CALL(newSV_type, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(newSV_type_mortal) && !defined(Perl_newSV_type_mortal)
#define Perl_newSV_type_mortal(...) GOPERL_LONG_CALL(newSV_type_mortal, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(newSVbool) && !defined(Perl_newSVbool)
#define Perl_newSVbool(...) GOPERL_LONG_CALL(newSVbool, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(newSVhek) && !defined(Perl_newSVhek)
#define Perl_newSVhek(...) GOPERL_LONG_CALL(newSVhek, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(newSVhek_mortal) && !defined(Perl_newSVhek_mortal)
#define Perl_newSVhek_mortal(...) GOPERL_LONG_CALL(newSVhek_mortal, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(newSViv) && !defined(Perl_newSViv)
#define Perl_newSViv(...) GOPERL_LONG_CALL(newSViv, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(newSVnv) && !defined(Perl_newSVnv)
#define Perl_newSVnv(...) GOPERL_LONG_CALL(newSVnv, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(newSVpv) && !defined(Perl_newSVpv)
#define Perl_newSVpv(...) GOPERL_LONG_CALL(newSVpv, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(newSVpv_share) && !defined(Perl_newSVpv_share)
#define Perl_newSVpv_share(...) GOPERL_LONG_CALL(newSVpv_share, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(newSVpvf_nocontext) && !defined(Perl_newSVpvf_nocontext)
#define Perl_newSVpvf_nocontext(...) GOPERL_LONG_CALL(newSVpvf_nocontext, __VA_ARGS__)
#endif
#if defined(newSVpvn) && !defined(Perl_newSVpvn)
#define Perl_newSVpvn(...) GOPERL_LONG_CALL(newSVpvn, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(newSVpvn_flags) && !defined(Perl_newSVpvn_flags)
#define Perl_newSVpvn_flags(...) GOPERL_LONG_CALL(newSVpvn_flags, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(newSVpvn_share) && !defined(Perl_newSVpvn_share)
#define Perl_newSVpvn_share(...) GOPERL_LONG_CALL(newSVpvn_share, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(newSVpvz) && !defined(Perl_newSVpvz)
#define Perl_newSVpvz(...) GOPERL_LONG_CALL(newSVpvz, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(newSVrv) && !defined(Perl_newSVrv)
#define Perl_newSVrv(...) GOPERL_LONG_CALL(newSVrv, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(newSVsv) && !defined(Perl_newSVsv)
#define Perl_newSVsv(...) GOPERL_LONG_CALL(newSVsv, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(newSVsv_flags) && !defined(Perl_newSVsv_flags)
#define Perl_newSVsv_flags(...) GOPERL_LONG_CALL(newSVsv_flags, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(newSVsv_flags_NN) && !defined(Perl_newSVsv_flags_NN)
#define Perl_newSVsv_flags_NN(...) GOPERL_LONG_CALL(newSVsv_flags_NN, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(newSVsv_nomg) && !defined(Perl_newSVsv_nomg)
#define Perl_newSVsv_nomg(...) GOPERL_LONG_CALL(newSVsv_nomg, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(newSVuv) && !defined(Perl_newSVuv)
#define Perl_newSVuv(...) GOPERL_LONG_CALL(newSVuv, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(newTRYCATCHOP) && !defined(Perl_newTRYCATCHOP)
#define Perl_newTRYCATCHOP(...) GOPERL_LONG_CALL(newTRYCATCHOP, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(newUNOP) && !defined(Perl_newUNOP)
#define Perl_newUNOP(...) GOPERL_LONG_CALL(newUNOP, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(newUNOP_AUX) && !defined(Perl_newUNOP_AUX)
#define Perl_newUNOP_AUX(...) GOPERL_LONG_CALL(newUNOP_AUX, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(newWHENOP) && !defined(Perl_newWHENOP)
#define Perl_newWHENOP(...) GOPERL_LONG_CALL(newWHENOP, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(newWHILEOP) && !defined(Perl_newWHILEOP)
#define Perl_newWHILEOP(...) GOPERL_LONG_CALL(newWHILEOP, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(newXS) && !defined(Perl_newXS)
#define Perl_newXS(...) GOPERL_LONG_CALL(newXS, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(newXS_flags) && !defined(Perl_newXS_flags)
#define Perl_newXS_flags(...) GOPERL_LONG_CALL(newXS_flags, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(new_version) && !defined(Perl_new_version)
#define Perl_new_version(...) GOPERL_LONG_CALL(new_version, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(ninstr) && !defined(Perl_ninstr)
#define Perl_ninstr(...) GOPERL_LONG_CALL(ninstr, __VA_ARGS__)
#endif
#if defined(nothreadhook) && !defined(Perl_nothreadhook)
#define Perl_nothreadhook(...) GOPERL_LONG_CALL(nothreadhook, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(op_append_elem) && !defined(Perl_op_append_elem)
#define Perl_op_append_elem(...) GOPERL_LONG_CALL(op_append_elem, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(op_append_list) && !defined(Perl_op_append_list)
#define Perl_op_append_list(...) GOPERL_LONG_CALL(op_append_list, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(op_class) && !defined(Perl_op_class)
#define Perl_op_class(...) GOPERL_LONG_CALL(op_class, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(op_clear) && !defined(Perl_op_clear)
#define Perl_op_clear(...) GOPERL_LONG_CALL(op_clear, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(op_contextualize) && !defined(Perl_op_contextualize)
#define Perl_op_contextualize(...) GOPERL_LONG_CALL(op_contextualize, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(op_convert_list) && !defined(Perl_op_convert_list)
#define Perl_op_convert_list(...) GOPERL_LONG_CALL(op_convert_list, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(op_dump) && !defined(Perl_op_dump)
#define Perl_op_dump(...) GOPERL_LONG_CALL(op_dump, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(op_force_list) && !defined(Perl_op_force_list)
#define Perl_op_force_list(...) GOPERL_LONG_CALL(op_force_list, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(op_free) && !defined(Perl_op_free)
#define Perl_op_free(...) GOPERL_LONG_CALL(op_free, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(op_linklist) && !defined(Perl_op_linklist)
#define Perl_op_linklist(...) GOPERL_LONG_CALL(op_linklist, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(op_lvalue) && !defined(Perl_op_lvalue)
#define Perl_op_lvalue(...) GOPERL_LONG_CALL(op_lvalue, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(op_null) && !defined(Perl_op_null)
#define Perl_op_null(...) GOPERL_LONG_CALL(op_null, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(op_parent) && !defined(Perl_op_parent)
#define Perl_op_parent(...) GOPERL_LONG_CALL(op_parent, __VA_ARGS__)
#endif
#if defined(op_prepend_elem) && !defined(Perl_op_prepend_elem)
#define Perl_op_prepend_elem(...) GOPERL_LONG_CALL(op_prepend_elem, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(op_scope) && !defined(Perl_op_scope)
#define Perl_op_scope(...) GOPERL_LONG_CALL(op_scope, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(op_sibling_splice) && !defined(Perl_op_sibling_splice)
#define Perl_op_sibling_splice(...) GOPERL_LONG_CALL(op_sibling_splice, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(op_wrap_finally) && !defined(Perl_op_wrap_finally)
#define Perl_op_wrap_finally(...) GOPERL_LONG_CALL(op_wrap_finally, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(opdump_printf) && !defined(Perl_opdump_printf)
#define Perl_opdump_printf(...) GOPERL_LONG_CALL(opdump_printf, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(optimize_optree) && !defined(Perl_optimize_optree)
#define Perl_optimize_optree(...) GOPERL_LONG_CALL(optimize_optree, GOPERL_THX_DROP(__VA_ARGS__))
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
#if defined(pad_add_name_pvn) && !defined(Perl_pad_add_name_pvn)
#define Perl_pad_add_name_pvn(...) GOPERL_LONG_CALL(pad_add_name_pvn, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(pad_add_name_sv) && !defined(Perl_pad_add_name_sv)
#define Perl_pad_add_name_sv(...) GOPERL_LONG_CALL(pad_add_name_sv, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(pad_findmy_pv) && !defined(Perl_pad_findmy_pv)
#define Perl_pad_findmy_pv(...) GOPERL_LONG_CALL(pad_findmy_pv, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(pad_findmy_pvn) && !defined(Perl_pad_findmy_pvn)
#define Perl_pad_findmy_pvn(...) GOPERL_LONG_CALL(pad_findmy_pvn, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(pad_findmy_sv) && !defined(Perl_pad_findmy_sv)
#define Perl_pad_findmy_sv(...) GOPERL_LONG_CALL(pad_findmy_sv, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(pad_tidy) && !defined(Perl_pad_tidy)
#define Perl_pad_tidy(...) GOPERL_LONG_CALL(pad_tidy, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(padnamelist_fetch) && !defined(Perl_padnamelist_fetch)
#define Perl_padnamelist_fetch(...) GOPERL_LONG_CALL(padnamelist_fetch, __VA_ARGS__)
#endif
#if defined(padnamelist_store) && !defined(Perl_padnamelist_store)
#define Perl_padnamelist_store(...) GOPERL_LONG_CALL(padnamelist_store, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(parse_arithexpr) && !defined(Perl_parse_arithexpr)
#define Perl_parse_arithexpr(...) GOPERL_LONG_CALL(parse_arithexpr, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(parse_barestmt) && !defined(Perl_parse_barestmt)
#define Perl_parse_barestmt(...) GOPERL_LONG_CALL(parse_barestmt, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(parse_block) && !defined(Perl_parse_block)
#define Perl_parse_block(...) GOPERL_LONG_CALL(parse_block, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(parse_fullexpr) && !defined(Perl_parse_fullexpr)
#define Perl_parse_fullexpr(...) GOPERL_LONG_CALL(parse_fullexpr, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(parse_fullstmt) && !defined(Perl_parse_fullstmt)
#define Perl_parse_fullstmt(...) GOPERL_LONG_CALL(parse_fullstmt, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(parse_label) && !defined(Perl_parse_label)
#define Perl_parse_label(...) GOPERL_LONG_CALL(parse_label, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(parse_listexpr) && !defined(Perl_parse_listexpr)
#define Perl_parse_listexpr(...) GOPERL_LONG_CALL(parse_listexpr, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(parse_stmtseq) && !defined(Perl_parse_stmtseq)
#define Perl_parse_stmtseq(...) GOPERL_LONG_CALL(parse_stmtseq, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(parse_subsignature) && !defined(Perl_parse_subsignature)
#define Perl_parse_subsignature(...) GOPERL_LONG_CALL(parse_subsignature, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(parse_termexpr) && !defined(Perl_parse_termexpr)
#define Perl_parse_termexpr(...) GOPERL_LONG_CALL(parse_termexpr, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(phase_name) && !defined(Perl_phase_name)
#define Perl_phase_name(...) GOPERL_LONG_CALL(phase_name, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(pmop_dump) && !defined(Perl_pmop_dump)
#define Perl_pmop_dump(...) GOPERL_LONG_CALL(pmop_dump, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(pregcomp) && !defined(Perl_pregcomp)
#define Perl_pregcomp(...) GOPERL_LONG_CALL(pregcomp, GOPERL_THX_DROP(__VA_ARGS__))
#endif
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
#if defined(pv_display) && !defined(Perl_pv_display)
#define Perl_pv_display(...) GOPERL_LONG_CALL(pv_display, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(pv_escape) && !defined(Perl_pv_escape)
#define Perl_pv_escape(...) GOPERL_LONG_CALL(pv_escape, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(pv_pretty) && !defined(Perl_pv_pretty)
#define Perl_pv_pretty(...) GOPERL_LONG_CALL(pv_pretty, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(pv_uni_display) && !defined(Perl_pv_uni_display)
#define Perl_pv_uni_display(...) GOPERL_LONG_CALL(pv_uni_display, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(rcpv_copy) && !defined(Perl_rcpv_copy)
#define Perl_rcpv_copy(...) GOPERL_LONG_CALL(rcpv_copy, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(rcpv_free) && !defined(Perl_rcpv_free)
#define Perl_rcpv_free(...) GOPERL_LONG_CALL(rcpv_free, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(rcpv_new) && !defined(Perl_rcpv_new)
#define Perl_rcpv_new(...) GOPERL_LONG_CALL(rcpv_new, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(re_compile) && !defined(Perl_re_compile)
#define Perl_re_compile(...) GOPERL_LONG_CALL(re_compile, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(re_dup_guts) && !defined(Perl_re_dup_guts)
#define Perl_re_dup_guts(...) GOPERL_LONG_CALL(re_dup_guts, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(repeatcpy) && !defined(Perl_repeatcpy)
#define Perl_repeatcpy(...) GOPERL_LONG_CALL(repeatcpy, __VA_ARGS__)
#endif
#if defined(require_pv) && !defined(Perl_require_pv)
#define Perl_require_pv(...) GOPERL_LONG_CALL(require_pv, GOPERL_THX_DROP(__VA_ARGS__))
#endif
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
#if defined(rpp_pop_1_norc) && !defined(Perl_rpp_pop_1_norc)
#define Perl_rpp_pop_1_norc(...) GOPERL_LONG_CALL(rpp_pop_1_norc, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(rpp_popfree_1) && !defined(Perl_rpp_popfree_1)
#define Perl_rpp_popfree_1(...) GOPERL_LONG_CALL(rpp_popfree_1, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(rpp_popfree_1_NN) && !defined(Perl_rpp_popfree_1_NN)
#define Perl_rpp_popfree_1_NN(...) GOPERL_LONG_CALL(rpp_popfree_1_NN, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(rpp_popfree_2) && !defined(Perl_rpp_popfree_2)
#define Perl_rpp_popfree_2(...) GOPERL_LONG_CALL(rpp_popfree_2, GOPERL_THX_DROP(__VA_ARGS__))
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
#if defined(rpp_push_1) && !defined(Perl_rpp_push_1)
#define Perl_rpp_push_1(...) GOPERL_LONG_CALL(rpp_push_1, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(rpp_push_1_norc) && !defined(Perl_rpp_push_1_norc)
#define Perl_rpp_push_1_norc(...) GOPERL_LONG_CALL(rpp_push_1_norc, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(rpp_push_2) && !defined(Perl_rpp_push_2)
#define Perl_rpp_push_2(...) GOPERL_LONG_CALL(rpp_push_2, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(rpp_push_IMM) && !defined(Perl_rpp_push_IMM)
#define Perl_rpp_push_IMM(...) GOPERL_LONG_CALL(rpp_push_IMM, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(rpp_replace_1_1) && !defined(Perl_rpp_replace_1_1)
#define Perl_rpp_replace_1_1(...) GOPERL_LONG_CALL(rpp_replace_1_1, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(rpp_replace_1_1_NN) && !defined(Perl_rpp_replace_1_1_NN)
#define Perl_rpp_replace_1_1_NN(...) GOPERL_LONG_CALL(rpp_replace_1_1_NN, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(rpp_replace_1_IMM_NN) && !defined(Perl_rpp_replace_1_IMM_NN)
#define Perl_rpp_replace_1_IMM_NN(...) GOPERL_LONG_CALL(rpp_replace_1_IMM_NN, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(rpp_replace_2_1) && !defined(Perl_rpp_replace_2_1)
#define Perl_rpp_replace_2_1(...) GOPERL_LONG_CALL(rpp_replace_2_1, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(rpp_replace_2_1_NN) && !defined(Perl_rpp_replace_2_1_NN)
#define Perl_rpp_replace_2_1_NN(...) GOPERL_LONG_CALL(rpp_replace_2_1_NN, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(rpp_replace_2_IMM_NN) && !defined(Perl_rpp_replace_2_IMM_NN)
#define Perl_rpp_replace_2_IMM_NN(...) GOPERL_LONG_CALL(rpp_replace_2_IMM_NN, GOPERL_THX_DROP(__VA_ARGS__))
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
#if defined(rv2cv_op_cv) && !defined(Perl_rv2cv_op_cv)
#define Perl_rv2cv_op_cv(...) GOPERL_LONG_CALL(rv2cv_op_cv, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(safesyscalloc) && !defined(Perl_safesyscalloc)
#define Perl_safesyscalloc(...) GOPERL_LONG_CALL(safesyscalloc, __VA_ARGS__)
#endif
#if defined(safesysfree) && !defined(Perl_safesysfree)
#define Perl_safesysfree(...) GOPERL_LONG_CALL(safesysfree, __VA_ARGS__)
#endif
#if defined(safesysmalloc) && !defined(Perl_safesysmalloc)
#define Perl_safesysmalloc(...) GOPERL_LONG_CALL(safesysmalloc, __VA_ARGS__)
#endif
#if defined(safesysrealloc) && !defined(Perl_safesysrealloc)
#define Perl_safesysrealloc(...) GOPERL_LONG_CALL(safesysrealloc, __VA_ARGS__)
#endif
#if defined(save_aelem) && !defined(Perl_save_aelem)
#define Perl_save_aelem(...) GOPERL_LONG_CALL(save_aelem, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(save_aelem_flags) && !defined(Perl_save_aelem_flags)
#define Perl_save_aelem_flags(...) GOPERL_LONG_CALL(save_aelem_flags, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(save_aptr) && !defined(Perl_save_aptr)
#define Perl_save_aptr(...) GOPERL_LONG_CALL(save_aptr, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(save_ary) && !defined(Perl_save_ary)
#define Perl_save_ary(...) GOPERL_LONG_CALL(save_ary, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(save_gp) && !defined(Perl_save_gp)
#define Perl_save_gp(...) GOPERL_LONG_CALL(save_gp, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(save_hash) && !defined(Perl_save_hash)
#define Perl_save_hash(...) GOPERL_LONG_CALL(save_hash, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(save_helem) && !defined(Perl_save_helem)
#define Perl_save_helem(...) GOPERL_LONG_CALL(save_helem, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(save_helem_flags) && !defined(Perl_save_helem_flags)
#define Perl_save_helem_flags(...) GOPERL_LONG_CALL(save_helem_flags, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(save_hptr) && !defined(Perl_save_hptr)
#define Perl_save_hptr(...) GOPERL_LONG_CALL(save_hptr, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(save_item) && !defined(Perl_save_item)
#define Perl_save_item(...) GOPERL_LONG_CALL(save_item, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(save_scalar) && !defined(Perl_save_scalar)
#define Perl_save_scalar(...) GOPERL_LONG_CALL(save_scalar, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(save_svref) && !defined(Perl_save_svref)
#define Perl_save_svref(...) GOPERL_LONG_CALL(save_svref, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(savepv) && !defined(Perl_savepv)
#define Perl_savepv(...) GOPERL_LONG_CALL(savepv, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(savepvn) && !defined(Perl_savepvn)
#define Perl_savepvn(...) GOPERL_LONG_CALL(savepvn, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(savesharedpv) && !defined(Perl_savesharedpv)
#define Perl_savesharedpv(...) GOPERL_LONG_CALL(savesharedpv, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(savesharedpvn) && !defined(Perl_savesharedpvn)
#define Perl_savesharedpvn(...) GOPERL_LONG_CALL(savesharedpvn, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(savesharedsvpv) && !defined(Perl_savesharedsvpv)
#define Perl_savesharedsvpv(...) GOPERL_LONG_CALL(savesharedsvpv, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(savesvpv) && !defined(Perl_savesvpv)
#define Perl_savesvpv(...) GOPERL_LONG_CALL(savesvpv, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(savetmps) && !defined(Perl_savetmps)
#define Perl_savetmps(...) GOPERL_LONG_CALL(savetmps, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(scan_bin) && !defined(Perl_scan_bin)
#define Perl_scan_bin(...) GOPERL_LONG_CALL(scan_bin, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(scan_hex) && !defined(Perl_scan_hex)
#define Perl_scan_hex(...) GOPERL_LONG_CALL(scan_hex, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(scan_oct) && !defined(Perl_scan_oct)
#define Perl_scan_oct(...) GOPERL_LONG_CALL(scan_oct, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(scan_vstring) && !defined(Perl_scan_vstring)
#define Perl_scan_vstring(...) GOPERL_LONG_CALL(scan_vstring, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(setdefout) && !defined(Perl_setdefout)
#define Perl_setdefout(...) GOPERL_LONG_CALL(setdefout, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(sortsv) && !defined(Perl_sortsv)
#define Perl_sortsv(...) GOPERL_LONG_CALL(sortsv, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(sortsv_flags) && !defined(Perl_sortsv_flags)
#define Perl_sortsv_flags(...) GOPERL_LONG_CALL(sortsv_flags, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(start_subparse) && !defined(Perl_start_subparse)
#define Perl_start_subparse(...) GOPERL_LONG_CALL(start_subparse, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(strict_utf8_to_uv) && !defined(Perl_strict_utf8_to_uv)
#define Perl_strict_utf8_to_uv(...) GOPERL_LONG_CALL(strict_utf8_to_uv, __VA_ARGS__)
#endif
#if defined(suspend_compcv) && !defined(Perl_suspend_compcv)
#define Perl_suspend_compcv(...) GOPERL_LONG_CALL(suspend_compcv, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(sv_2cv) && !defined(Perl_sv_2cv)
#define Perl_sv_2cv(...) GOPERL_LONG_CALL(sv_2cv, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(sv_2io) && !defined(Perl_sv_2io)
#define Perl_sv_2io(...) GOPERL_LONG_CALL(sv_2io, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(sv_2iv_flags) && !defined(Perl_sv_2iv_flags)
#define Perl_sv_2iv_flags(...) GOPERL_LONG_CALL(sv_2iv_flags, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(sv_2mortal) && !defined(Perl_sv_2mortal)
#define Perl_sv_2mortal(...) GOPERL_LONG_CALL(sv_2mortal, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(sv_2nv_flags) && !defined(Perl_sv_2nv_flags)
#define Perl_sv_2nv_flags(...) GOPERL_LONG_CALL(sv_2nv_flags, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(sv_2pv) && !defined(Perl_sv_2pv)
#define Perl_sv_2pv(...) GOPERL_LONG_CALL(sv_2pv, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(sv_2pv_flags) && !defined(Perl_sv_2pv_flags)
#define Perl_sv_2pv_flags(...) GOPERL_LONG_CALL(sv_2pv_flags, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(sv_2pvbyte) && !defined(Perl_sv_2pvbyte)
#define Perl_sv_2pvbyte(...) GOPERL_LONG_CALL(sv_2pvbyte, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(sv_2pvbyte_flags) && !defined(Perl_sv_2pvbyte_flags)
#define Perl_sv_2pvbyte_flags(...) GOPERL_LONG_CALL(sv_2pvbyte_flags, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(sv_2pvutf8) && !defined(Perl_sv_2pvutf8)
#define Perl_sv_2pvutf8(...) GOPERL_LONG_CALL(sv_2pvutf8, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(sv_2pvutf8_flags) && !defined(Perl_sv_2pvutf8_flags)
#define Perl_sv_2pvutf8_flags(...) GOPERL_LONG_CALL(sv_2pvutf8_flags, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(sv_2uv_flags) && !defined(Perl_sv_2uv_flags)
#define Perl_sv_2uv_flags(...) GOPERL_LONG_CALL(sv_2uv_flags, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(sv_backoff) && !defined(Perl_sv_backoff)
#define Perl_sv_backoff(...) GOPERL_LONG_CALL(sv_backoff, __VA_ARGS__)
#endif
#if defined(sv_bless) && !defined(Perl_sv_bless)
#define Perl_sv_bless(...) GOPERL_LONG_CALL(sv_bless, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(sv_catpv) && !defined(Perl_sv_catpv)
#define Perl_sv_catpv(...) GOPERL_LONG_CALL(sv_catpv, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(sv_catpv_flags) && !defined(Perl_sv_catpv_flags)
#define Perl_sv_catpv_flags(...) GOPERL_LONG_CALL(sv_catpv_flags, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(sv_catpv_mg) && !defined(Perl_sv_catpv_mg)
#define Perl_sv_catpv_mg(...) GOPERL_LONG_CALL(sv_catpv_mg, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(sv_catpvf_mg) && !defined(Perl_sv_catpvf_mg)
#define Perl_sv_catpvf_mg(...) GOPERL_LONG_CALL(sv_catpvf_mg, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(sv_catpvf_mg_nocontext) && !defined(Perl_sv_catpvf_mg_nocontext)
#define Perl_sv_catpvf_mg_nocontext(...) GOPERL_LONG_CALL(sv_catpvf_mg_nocontext, __VA_ARGS__)
#endif
#if defined(sv_catpvf_nocontext) && !defined(Perl_sv_catpvf_nocontext)
#define Perl_sv_catpvf_nocontext(...) GOPERL_LONG_CALL(sv_catpvf_nocontext, __VA_ARGS__)
#endif
#if defined(sv_catpvn) && !defined(Perl_sv_catpvn)
#define Perl_sv_catpvn(...) GOPERL_LONG_CALL(sv_catpvn, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(sv_catpvn_flags) && !defined(Perl_sv_catpvn_flags)
#define Perl_sv_catpvn_flags(...) GOPERL_LONG_CALL(sv_catpvn_flags, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(sv_catpvn_mg) && !defined(Perl_sv_catpvn_mg)
#define Perl_sv_catpvn_mg(...) GOPERL_LONG_CALL(sv_catpvn_mg, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(sv_catsv) && !defined(Perl_sv_catsv)
#define Perl_sv_catsv(...) GOPERL_LONG_CALL(sv_catsv, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(sv_catsv_flags) && !defined(Perl_sv_catsv_flags)
#define Perl_sv_catsv_flags(...) GOPERL_LONG_CALL(sv_catsv_flags, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(sv_catsv_mg) && !defined(Perl_sv_catsv_mg)
#define Perl_sv_catsv_mg(...) GOPERL_LONG_CALL(sv_catsv_mg, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(sv_chop) && !defined(Perl_sv_chop)
#define Perl_sv_chop(...) GOPERL_LONG_CALL(sv_chop, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(sv_clear) && !defined(Perl_sv_clear)
#define Perl_sv_clear(...) GOPERL_LONG_CALL(sv_clear, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(sv_cmp_flags) && !defined(Perl_sv_cmp_flags)
#define Perl_sv_cmp_flags(...) GOPERL_LONG_CALL(sv_cmp_flags, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(sv_cmp_locale_flags) && !defined(Perl_sv_cmp_locale_flags)
#define Perl_sv_cmp_locale_flags(...) GOPERL_LONG_CALL(sv_cmp_locale_flags, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(sv_collxfrm) && !defined(Perl_sv_collxfrm)
#define Perl_sv_collxfrm(...) GOPERL_LONG_CALL(sv_collxfrm, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(sv_collxfrm_flags) && !defined(Perl_sv_collxfrm_flags)
#define Perl_sv_collxfrm_flags(...) GOPERL_LONG_CALL(sv_collxfrm_flags, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(sv_copypv) && !defined(Perl_sv_copypv)
#define Perl_sv_copypv(...) GOPERL_LONG_CALL(sv_copypv, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(sv_copypv_flags) && !defined(Perl_sv_copypv_flags)
#define Perl_sv_copypv_flags(...) GOPERL_LONG_CALL(sv_copypv_flags, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(sv_copypv_nomg) && !defined(Perl_sv_copypv_nomg)
#define Perl_sv_copypv_nomg(...) GOPERL_LONG_CALL(sv_copypv_nomg, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(sv_dec) && !defined(Perl_sv_dec)
#define Perl_sv_dec(...) GOPERL_LONG_CALL(sv_dec, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(sv_dec_nomg) && !defined(Perl_sv_dec_nomg)
#define Perl_sv_dec_nomg(...) GOPERL_LONG_CALL(sv_dec_nomg, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(sv_derived_from) && !defined(Perl_sv_derived_from)
#define Perl_sv_derived_from(...) GOPERL_LONG_CALL(sv_derived_from, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(sv_derived_from_hv) && !defined(Perl_sv_derived_from_hv)
#define Perl_sv_derived_from_hv(...) GOPERL_LONG_CALL(sv_derived_from_hv, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(sv_derived_from_pv) && !defined(Perl_sv_derived_from_pv)
#define Perl_sv_derived_from_pv(...) GOPERL_LONG_CALL(sv_derived_from_pv, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(sv_derived_from_pvn) && !defined(Perl_sv_derived_from_pvn)
#define Perl_sv_derived_from_pvn(...) GOPERL_LONG_CALL(sv_derived_from_pvn, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(sv_derived_from_sv) && !defined(Perl_sv_derived_from_sv)
#define Perl_sv_derived_from_sv(...) GOPERL_LONG_CALL(sv_derived_from_sv, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(sv_destroyable) && !defined(Perl_sv_destroyable)
#define Perl_sv_destroyable(...) GOPERL_LONG_CALL(sv_destroyable, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(sv_does) && !defined(Perl_sv_does)
#define Perl_sv_does(...) GOPERL_LONG_CALL(sv_does, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(sv_does_pv) && !defined(Perl_sv_does_pv)
#define Perl_sv_does_pv(...) GOPERL_LONG_CALL(sv_does_pv, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(sv_does_pvn) && !defined(Perl_sv_does_pvn)
#define Perl_sv_does_pvn(...) GOPERL_LONG_CALL(sv_does_pvn, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(sv_does_sv) && !defined(Perl_sv_does_sv)
#define Perl_sv_does_sv(...) GOPERL_LONG_CALL(sv_does_sv, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(sv_dump) && !defined(Perl_sv_dump)
#define Perl_sv_dump(...) GOPERL_LONG_CALL(sv_dump, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(sv_dump_depth) && !defined(Perl_sv_dump_depth)
#define Perl_sv_dump_depth(...) GOPERL_LONG_CALL(sv_dump_depth, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(sv_dup) && !defined(Perl_sv_dup)
#define Perl_sv_dup(...) GOPERL_LONG_CALL(sv_dup, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(sv_dup_inc) && !defined(Perl_sv_dup_inc)
#define Perl_sv_dup_inc(...) GOPERL_LONG_CALL(sv_dup_inc, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(sv_eq) && !defined(Perl_sv_eq)
#define Perl_sv_eq(...) GOPERL_LONG_CALL(sv_eq, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(sv_eq_flags) && !defined(Perl_sv_eq_flags)
#define Perl_sv_eq_flags(...) GOPERL_LONG_CALL(sv_eq_flags, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(sv_force_normal) && !defined(Perl_sv_force_normal)
#define Perl_sv_force_normal(...) GOPERL_LONG_CALL(sv_force_normal, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(sv_force_normal_flags) && !defined(Perl_sv_force_normal_flags)
#define Perl_sv_force_normal_flags(...) GOPERL_LONG_CALL(sv_force_normal_flags, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(sv_free) && !defined(Perl_sv_free)
#define Perl_sv_free(...) GOPERL_LONG_CALL(sv_free, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(sv_get_backrefs) && !defined(Perl_sv_get_backrefs)
#define Perl_sv_get_backrefs(...) GOPERL_LONG_CALL(sv_get_backrefs, __VA_ARGS__)
#endif
#if defined(sv_inc) && !defined(Perl_sv_inc)
#define Perl_sv_inc(...) GOPERL_LONG_CALL(sv_inc, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(sv_inc_nomg) && !defined(Perl_sv_inc_nomg)
#define Perl_sv_inc_nomg(...) GOPERL_LONG_CALL(sv_inc_nomg, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(sv_insert) && !defined(Perl_sv_insert)
#define Perl_sv_insert(...) GOPERL_LONG_CALL(sv_insert, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(sv_insert_flags) && !defined(Perl_sv_insert_flags)
#define Perl_sv_insert_flags(...) GOPERL_LONG_CALL(sv_insert_flags, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(sv_isa) && !defined(Perl_sv_isa)
#define Perl_sv_isa(...) GOPERL_LONG_CALL(sv_isa, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(sv_isa_sv) && !defined(Perl_sv_isa_sv)
#define Perl_sv_isa_sv(...) GOPERL_LONG_CALL(sv_isa_sv, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(sv_isobject) && !defined(Perl_sv_isobject)
#define Perl_sv_isobject(...) GOPERL_LONG_CALL(sv_isobject, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(sv_langinfo) && !defined(Perl_sv_langinfo)
#define Perl_sv_langinfo(...) GOPERL_LONG_CALL(sv_langinfo, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(sv_len) && !defined(Perl_sv_len)
#define Perl_sv_len(...) GOPERL_LONG_CALL(sv_len, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(sv_len_utf8) && !defined(Perl_sv_len_utf8)
#define Perl_sv_len_utf8(...) GOPERL_LONG_CALL(sv_len_utf8, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(sv_len_utf8_nomg) && !defined(Perl_sv_len_utf8_nomg)
#define Perl_sv_len_utf8_nomg(...) GOPERL_LONG_CALL(sv_len_utf8_nomg, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(sv_magic) && !defined(Perl_sv_magic)
#define Perl_sv_magic(...) GOPERL_LONG_CALL(sv_magic, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(sv_magicext) && !defined(Perl_sv_magicext)
#define Perl_sv_magicext(...) GOPERL_LONG_CALL(sv_magicext, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(sv_mortalcopy) && !defined(Perl_sv_mortalcopy)
#define Perl_sv_mortalcopy(...) GOPERL_LONG_CALL(sv_mortalcopy, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(sv_mortalcopy_flags) && !defined(Perl_sv_mortalcopy_flags)
#define Perl_sv_mortalcopy_flags(...) GOPERL_LONG_CALL(sv_mortalcopy_flags, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(sv_newmortal) && !defined(Perl_sv_newmortal)
#define Perl_sv_newmortal(...) GOPERL_LONG_CALL(sv_newmortal, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(sv_nosharing) && !defined(Perl_sv_nosharing)
#define Perl_sv_nosharing(...) GOPERL_LONG_CALL(sv_nosharing, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(sv_numcmp) && !defined(Perl_sv_numcmp)
#define Perl_sv_numcmp(...) GOPERL_LONG_CALL(sv_numcmp, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(sv_numcmp_flags) && !defined(Perl_sv_numcmp_flags)
#define Perl_sv_numcmp_flags(...) GOPERL_LONG_CALL(sv_numcmp_flags, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(sv_numeq) && !defined(Perl_sv_numeq)
#define Perl_sv_numeq(...) GOPERL_LONG_CALL(sv_numeq, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(sv_numeq_flags) && !defined(Perl_sv_numeq_flags)
#define Perl_sv_numeq_flags(...) GOPERL_LONG_CALL(sv_numeq_flags, GOPERL_THX_DROP(__VA_ARGS__))
#endif
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
#if defined(sv_pos_b2u) && !defined(Perl_sv_pos_b2u)
#define Perl_sv_pos_b2u(...) GOPERL_LONG_CALL(sv_pos_b2u, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(sv_pos_b2u_flags) && !defined(Perl_sv_pos_b2u_flags)
#define Perl_sv_pos_b2u_flags(...) GOPERL_LONG_CALL(sv_pos_b2u_flags, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(sv_pos_u2b) && !defined(Perl_sv_pos_u2b)
#define Perl_sv_pos_u2b(...) GOPERL_LONG_CALL(sv_pos_u2b, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(sv_pos_u2b_flags) && !defined(Perl_sv_pos_u2b_flags)
#define Perl_sv_pos_u2b_flags(...) GOPERL_LONG_CALL(sv_pos_u2b_flags, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(sv_pvn_force_flags) && !defined(Perl_sv_pvn_force_flags)
#define Perl_sv_pvn_force_flags(...) GOPERL_LONG_CALL(sv_pvn_force_flags, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(sv_recode_to_utf8) && !defined(Perl_sv_recode_to_utf8)
#define Perl_sv_recode_to_utf8(...) GOPERL_LONG_CALL(sv_recode_to_utf8, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(sv_ref) && !defined(Perl_sv_ref)
#define Perl_sv_ref(...) GOPERL_LONG_CALL(sv_ref, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(sv_reftype) && !defined(Perl_sv_reftype)
#define Perl_sv_reftype(...) GOPERL_LONG_CALL(sv_reftype, GOPERL_THX_DROP(__VA_ARGS__))
#endif
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
#if defined(sv_rvunweaken) && !defined(Perl_sv_rvunweaken)
#define Perl_sv_rvunweaken(...) GOPERL_LONG_CALL(sv_rvunweaken, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(sv_rvweaken) && !defined(Perl_sv_rvweaken)
#define Perl_sv_rvweaken(...) GOPERL_LONG_CALL(sv_rvweaken, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(sv_set_bool) && !defined(Perl_sv_set_bool)
#define Perl_sv_set_bool(...) GOPERL_LONG_CALL(sv_set_bool, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(sv_set_false) && !defined(Perl_sv_set_false)
#define Perl_sv_set_false(...) GOPERL_LONG_CALL(sv_set_false, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(sv_set_true) && !defined(Perl_sv_set_true)
#define Perl_sv_set_true(...) GOPERL_LONG_CALL(sv_set_true, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(sv_set_undef) && !defined(Perl_sv_set_undef)
#define Perl_sv_set_undef(...) GOPERL_LONG_CALL(sv_set_undef, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(sv_setiv) && !defined(Perl_sv_setiv)
#define Perl_sv_setiv(...) GOPERL_LONG_CALL(sv_setiv, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(sv_setiv_mg) && !defined(Perl_sv_setiv_mg)
#define Perl_sv_setiv_mg(...) GOPERL_LONG_CALL(sv_setiv_mg, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(sv_setnv) && !defined(Perl_sv_setnv)
#define Perl_sv_setnv(...) GOPERL_LONG_CALL(sv_setnv, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(sv_setnv_mg) && !defined(Perl_sv_setnv_mg)
#define Perl_sv_setnv_mg(...) GOPERL_LONG_CALL(sv_setnv_mg, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(sv_setpv) && !defined(Perl_sv_setpv)
#define Perl_sv_setpv(...) GOPERL_LONG_CALL(sv_setpv, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(sv_setpv_bufsize) && !defined(Perl_sv_setpv_bufsize)
#define Perl_sv_setpv_bufsize(...) GOPERL_LONG_CALL(sv_setpv_bufsize, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(sv_setpv_mg) && !defined(Perl_sv_setpv_mg)
#define Perl_sv_setpv_mg(...) GOPERL_LONG_CALL(sv_setpv_mg, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(sv_setpvf_mg) && !defined(Perl_sv_setpvf_mg)
#define Perl_sv_setpvf_mg(...) GOPERL_LONG_CALL(sv_setpvf_mg, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(sv_setpvf_mg_nocontext) && !defined(Perl_sv_setpvf_mg_nocontext)
#define Perl_sv_setpvf_mg_nocontext(...) GOPERL_LONG_CALL(sv_setpvf_mg_nocontext, __VA_ARGS__)
#endif
#if defined(sv_setpvf_nocontext) && !defined(Perl_sv_setpvf_nocontext)
#define Perl_sv_setpvf_nocontext(...) GOPERL_LONG_CALL(sv_setpvf_nocontext, __VA_ARGS__)
#endif
#if defined(sv_setpvn) && !defined(Perl_sv_setpvn)
#define Perl_sv_setpvn(...) GOPERL_LONG_CALL(sv_setpvn, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(sv_setpvn_fresh) && !defined(Perl_sv_setpvn_fresh)
#define Perl_sv_setpvn_fresh(...) GOPERL_LONG_CALL(sv_setpvn_fresh, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(sv_setpvn_mg) && !defined(Perl_sv_setpvn_mg)
#define Perl_sv_setpvn_mg(...) GOPERL_LONG_CALL(sv_setpvn_mg, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(sv_setref_iv) && !defined(Perl_sv_setref_iv)
#define Perl_sv_setref_iv(...) GOPERL_LONG_CALL(sv_setref_iv, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(sv_setref_nv) && !defined(Perl_sv_setref_nv)
#define Perl_sv_setref_nv(...) GOPERL_LONG_CALL(sv_setref_nv, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(sv_setref_pv) && !defined(Perl_sv_setref_pv)
#define Perl_sv_setref_pv(...) GOPERL_LONG_CALL(sv_setref_pv, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(sv_setref_pvn) && !defined(Perl_sv_setref_pvn)
#define Perl_sv_setref_pvn(...) GOPERL_LONG_CALL(sv_setref_pvn, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(sv_setref_uv) && !defined(Perl_sv_setref_uv)
#define Perl_sv_setref_uv(...) GOPERL_LONG_CALL(sv_setref_uv, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(sv_setrv_inc) && !defined(Perl_sv_setrv_inc)
#define Perl_sv_setrv_inc(...) GOPERL_LONG_CALL(sv_setrv_inc, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(sv_setrv_inc_mg) && !defined(Perl_sv_setrv_inc_mg)
#define Perl_sv_setrv_inc_mg(...) GOPERL_LONG_CALL(sv_setrv_inc_mg, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(sv_setrv_noinc) && !defined(Perl_sv_setrv_noinc)
#define Perl_sv_setrv_noinc(...) GOPERL_LONG_CALL(sv_setrv_noinc, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(sv_setrv_noinc_mg) && !defined(Perl_sv_setrv_noinc_mg)
#define Perl_sv_setrv_noinc_mg(...) GOPERL_LONG_CALL(sv_setrv_noinc_mg, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(sv_setsv) && !defined(Perl_sv_setsv)
#define Perl_sv_setsv(...) GOPERL_LONG_CALL(sv_setsv, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(sv_setsv_flags) && !defined(Perl_sv_setsv_flags)
#define Perl_sv_setsv_flags(...) GOPERL_LONG_CALL(sv_setsv_flags, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(sv_setsv_mg) && !defined(Perl_sv_setsv_mg)
#define Perl_sv_setsv_mg(...) GOPERL_LONG_CALL(sv_setsv_mg, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(sv_setuv) && !defined(Perl_sv_setuv)
#define Perl_sv_setuv(...) GOPERL_LONG_CALL(sv_setuv, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(sv_setuv_mg) && !defined(Perl_sv_setuv_mg)
#define Perl_sv_setuv_mg(...) GOPERL_LONG_CALL(sv_setuv_mg, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(sv_streq) && !defined(Perl_sv_streq)
#define Perl_sv_streq(...) GOPERL_LONG_CALL(sv_streq, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(sv_streq_flags) && !defined(Perl_sv_streq_flags)
#define Perl_sv_streq_flags(...) GOPERL_LONG_CALL(sv_streq_flags, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(sv_strftime_ints) && !defined(Perl_sv_strftime_ints)
#define Perl_sv_strftime_ints(...) GOPERL_LONG_CALL(sv_strftime_ints, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(sv_strftime_tm) && !defined(Perl_sv_strftime_tm)
#define Perl_sv_strftime_tm(...) GOPERL_LONG_CALL(sv_strftime_tm, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(sv_string_from_errnum) && !defined(Perl_sv_string_from_errnum)
#define Perl_sv_string_from_errnum(...) GOPERL_LONG_CALL(sv_string_from_errnum, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(sv_uni_display) && !defined(Perl_sv_uni_display)
#define Perl_sv_uni_display(...) GOPERL_LONG_CALL(sv_uni_display, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(sv_unmagic) && !defined(Perl_sv_unmagic)
#define Perl_sv_unmagic(...) GOPERL_LONG_CALL(sv_unmagic, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(sv_unmagicext) && !defined(Perl_sv_unmagicext)
#define Perl_sv_unmagicext(...) GOPERL_LONG_CALL(sv_unmagicext, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(sv_unref) && !defined(Perl_sv_unref)
#define Perl_sv_unref(...) GOPERL_LONG_CALL(sv_unref, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(sv_unref_flags) && !defined(Perl_sv_unref_flags)
#define Perl_sv_unref_flags(...) GOPERL_LONG_CALL(sv_unref_flags, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(sv_upgrade) && !defined(Perl_sv_upgrade)
#define Perl_sv_upgrade(...) GOPERL_LONG_CALL(sv_upgrade, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(sv_usepvn) && !defined(Perl_sv_usepvn)
#define Perl_sv_usepvn(...) GOPERL_LONG_CALL(sv_usepvn, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(sv_usepvn_flags) && !defined(Perl_sv_usepvn_flags)
#define Perl_sv_usepvn_flags(...) GOPERL_LONG_CALL(sv_usepvn_flags, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(sv_usepvn_mg) && !defined(Perl_sv_usepvn_mg)
#define Perl_sv_usepvn_mg(...) GOPERL_LONG_CALL(sv_usepvn_mg, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(sv_utf8_decode) && !defined(Perl_sv_utf8_decode)
#define Perl_sv_utf8_decode(...) GOPERL_LONG_CALL(sv_utf8_decode, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(sv_utf8_downgrade) && !defined(Perl_sv_utf8_downgrade)
#define Perl_sv_utf8_downgrade(...) GOPERL_LONG_CALL(sv_utf8_downgrade, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(sv_utf8_downgrade_flags) && !defined(Perl_sv_utf8_downgrade_flags)
#define Perl_sv_utf8_downgrade_flags(...) GOPERL_LONG_CALL(sv_utf8_downgrade_flags, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(sv_utf8_downgrade_nomg) && !defined(Perl_sv_utf8_downgrade_nomg)
#define Perl_sv_utf8_downgrade_nomg(...) GOPERL_LONG_CALL(sv_utf8_downgrade_nomg, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(sv_utf8_encode) && !defined(Perl_sv_utf8_encode)
#define Perl_sv_utf8_encode(...) GOPERL_LONG_CALL(sv_utf8_encode, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(sv_utf8_upgrade) && !defined(Perl_sv_utf8_upgrade)
#define Perl_sv_utf8_upgrade(...) GOPERL_LONG_CALL(sv_utf8_upgrade, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(sv_utf8_upgrade_flags) && !defined(Perl_sv_utf8_upgrade_flags)
#define Perl_sv_utf8_upgrade_flags(...) GOPERL_LONG_CALL(sv_utf8_upgrade_flags, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(sv_utf8_upgrade_flags_grow) && !defined(Perl_sv_utf8_upgrade_flags_grow)
#define Perl_sv_utf8_upgrade_flags_grow(...) GOPERL_LONG_CALL(sv_utf8_upgrade_flags_grow, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(sv_utf8_upgrade_nomg) && !defined(Perl_sv_utf8_upgrade_nomg)
#define Perl_sv_utf8_upgrade_nomg(...) GOPERL_LONG_CALL(sv_utf8_upgrade_nomg, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(sv_vcatpvf) && !defined(Perl_sv_vcatpvf)
#define Perl_sv_vcatpvf(...) GOPERL_LONG_CALL(sv_vcatpvf, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(sv_vcatpvf_mg) && !defined(Perl_sv_vcatpvf_mg)
#define Perl_sv_vcatpvf_mg(...) GOPERL_LONG_CALL(sv_vcatpvf_mg, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(sv_vcatpvfn) && !defined(Perl_sv_vcatpvfn)
#define Perl_sv_vcatpvfn(...) GOPERL_LONG_CALL(sv_vcatpvfn, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(sv_vcatpvfn_flags) && !defined(Perl_sv_vcatpvfn_flags)
#define Perl_sv_vcatpvfn_flags(...) GOPERL_LONG_CALL(sv_vcatpvfn_flags, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(sv_vsetpvf) && !defined(Perl_sv_vsetpvf)
#define Perl_sv_vsetpvf(...) GOPERL_LONG_CALL(sv_vsetpvf, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(sv_vsetpvf_mg) && !defined(Perl_sv_vsetpvf_mg)
#define Perl_sv_vsetpvf_mg(...) GOPERL_LONG_CALL(sv_vsetpvf_mg, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(sv_vsetpvfn) && !defined(Perl_sv_vsetpvfn)
#define Perl_sv_vsetpvfn(...) GOPERL_LONG_CALL(sv_vsetpvfn, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(sv_vstring_get) && !defined(Perl_sv_vstring_get)
#define Perl_sv_vstring_get(...) GOPERL_LONG_CALL(sv_vstring_get, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(switch_to_global_locale) && !defined(Perl_switch_to_global_locale)
#define Perl_switch_to_global_locale(...) GOPERL_LONG_CALL(switch_to_global_locale, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(sync_locale) && !defined(Perl_sync_locale)
#define Perl_sync_locale(...) GOPERL_LONG_CALL(sync_locale, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(thread_locale_init) && !defined(Perl_thread_locale_init)
#define Perl_thread_locale_init(...) GOPERL_LONG_CALL(thread_locale_init, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(thread_locale_term) && !defined(Perl_thread_locale_term)
#define Perl_thread_locale_term(...) GOPERL_LONG_CALL(thread_locale_term, GOPERL_THX_DROP(__VA_ARGS__))
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
#if defined(utf8_distance) && !defined(Perl_utf8_distance)
#define Perl_utf8_distance(...) GOPERL_LONG_CALL(utf8_distance, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(utf8_hop) && !defined(Perl_utf8_hop)
#define Perl_utf8_hop(...) GOPERL_LONG_CALL(utf8_hop, __VA_ARGS__)
#endif
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
#if defined(utf8_to_bytes) && !defined(Perl_utf8_to_bytes)
#define Perl_utf8_to_bytes(...) GOPERL_LONG_CALL(utf8_to_bytes, GOPERL_THX_DROP(__VA_ARGS__))
#endif
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
#if defined(utf8_to_uvchr) && !defined(Perl_utf8_to_uvchr)
#define Perl_utf8_to_uvchr(...) GOPERL_LONG_CALL(utf8_to_uvchr, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(utf8_to_uvchr_buf) && !defined(Perl_utf8_to_uvchr_buf)
#define Perl_utf8_to_uvchr_buf(...) GOPERL_LONG_CALL(utf8_to_uvchr_buf, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(utf8n_to_uvchr) && !defined(Perl_utf8n_to_uvchr)
#define Perl_utf8n_to_uvchr(...) GOPERL_LONG_CALL(utf8n_to_uvchr, __VA_ARGS__)
#endif
#if defined(utf8n_to_uvchr_error) && !defined(Perl_utf8n_to_uvchr_error)
#define Perl_utf8n_to_uvchr_error(...) GOPERL_LONG_CALL(utf8n_to_uvchr_error, __VA_ARGS__)
#endif
#if defined(utf8n_to_uvchr_msgs) && !defined(Perl_utf8n_to_uvchr_msgs)
#define Perl_utf8n_to_uvchr_msgs(...) GOPERL_LONG_CALL(utf8n_to_uvchr_msgs, __VA_ARGS__)
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
#if defined(uvchr_to_utf8) && !defined(Perl_uvchr_to_utf8)
#define Perl_uvchr_to_utf8(...) GOPERL_LONG_CALL(uvchr_to_utf8, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(uvchr_to_utf8_flags) && !defined(Perl_uvchr_to_utf8_flags)
#define Perl_uvchr_to_utf8_flags(...) GOPERL_LONG_CALL(uvchr_to_utf8_flags, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(uvchr_to_utf8_flags_msgs) && !defined(Perl_uvchr_to_utf8_flags_msgs)
#define Perl_uvchr_to_utf8_flags_msgs(...) GOPERL_LONG_CALL(uvchr_to_utf8_flags_msgs, GOPERL_THX_DROP(__VA_ARGS__))
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
#if defined(vcroak) && !defined(Perl_vcroak)
#define Perl_vcroak(...) GOPERL_LONG_CALL(vcroak, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(vdeb) && !defined(Perl_vdeb)
#define Perl_vdeb(...) GOPERL_LONG_CALL(vdeb, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(vfatal_warner) && !defined(Perl_vfatal_warner)
#define Perl_vfatal_warner(...) GOPERL_LONG_CALL(vfatal_warner, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(vform) && !defined(Perl_vform)
#define Perl_vform(...) GOPERL_LONG_CALL(vform, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(vload_module) && !defined(Perl_vload_module)
#define Perl_vload_module(...) GOPERL_LONG_CALL(vload_module, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(vmess) && !defined(Perl_vmess)
#define Perl_vmess(...) GOPERL_LONG_CALL(vmess, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(vnewSVpvf) && !defined(Perl_vnewSVpvf)
#define Perl_vnewSVpvf(...) GOPERL_LONG_CALL(vnewSVpvf, GOPERL_THX_DROP(__VA_ARGS__))
#endif
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
#if defined(vwarn) && !defined(Perl_vwarn)
#define Perl_vwarn(...) GOPERL_LONG_CALL(vwarn, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(vwarner) && !defined(Perl_vwarner)
#define Perl_vwarner(...) GOPERL_LONG_CALL(vwarner, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(warn_sv) && !defined(Perl_warn_sv)
#define Perl_warn_sv(...) GOPERL_LONG_CALL(warn_sv, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(warner_nocontext) && !defined(Perl_warner_nocontext)
#define Perl_warner_nocontext(...) GOPERL_LONG_CALL(warner_nocontext, __VA_ARGS__)
#endif
#if defined(whichsig) && !defined(Perl_whichsig)
#define Perl_whichsig(...) GOPERL_LONG_CALL(whichsig, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(whichsig_pv) && !defined(Perl_whichsig_pv)
#define Perl_whichsig_pv(...) GOPERL_LONG_CALL(whichsig_pv, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(whichsig_pvn) && !defined(Perl_whichsig_pvn)
#define Perl_whichsig_pvn(...) GOPERL_LONG_CALL(whichsig_pvn, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(whichsig_sv) && !defined(Perl_whichsig_sv)
#define Perl_whichsig_sv(...) GOPERL_LONG_CALL(whichsig_sv, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(wrap_infix_plugin) && !defined(Perl_wrap_infix_plugin)
#define Perl_wrap_infix_plugin(...) GOPERL_LONG_CALL(wrap_infix_plugin, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(wrap_keyword_plugin) && !defined(Perl_wrap_keyword_plugin)
#define Perl_wrap_keyword_plugin(...) GOPERL_LONG_CALL(wrap_keyword_plugin, GOPERL_THX_DROP(__VA_ARGS__))
#endif

#endif /* GOPERL_XS_SDK_LONGNAMES_H */
