/* go-perl native XS SDK: the long names of the functions of perl's
 * embed.fnc that extensions call although they are outside the public API
 * (GENERATED from perl v5.44.0's embed.fnc by gen_longnames_core.py; do
 * not edit by hand).
 *
 * They are flagged C ("core, but usable by extensions") or X (exported):
 * perl gives each its Perl_<name> as well, which a module calls with the
 * interpreter as the first argument, Perl_sv_grow(aTHX_ sv, len). Each is
 * a macro over the short name, where the SDK provides it, with the
 * interpreter argument left out (aTHX_ is one argument of the call before
 * it is expanded; the interpreter is the SDK's current frame). */
#ifndef GOPERL_XS_SDK_LONGNAMES_CORE_H
#define GOPERL_XS_SDK_LONGNAMES_CORE_H

#if defined(amagic_applies) && !defined(Perl_amagic_applies)
#define Perl_amagic_applies(...) GOPERL_LONG_CALL(amagic_applies, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(api_version_assert) && !defined(Perl_api_version_assert)
#define Perl_api_version_assert(...) GOPERL_LONG_CALL(api_version_assert, __VA_ARGS__)
#endif
#if defined(append_utf8_from_native_byte) && !defined(Perl_append_utf8_from_native_byte)
#define Perl_append_utf8_from_native_byte(...) GOPERL_LONG_CALL(append_utf8_from_native_byte, __VA_ARGS__)
#endif
#if defined(atfork_child) && !defined(Perl_atfork_child)
#define Perl_atfork_child(...) GOPERL_LONG_CALL(atfork_child, __VA_ARGS__)
#endif
#if defined(atfork_lock) && !defined(Perl_atfork_lock)
#define Perl_atfork_lock(...) GOPERL_LONG_CALL(atfork_lock, __VA_ARGS__)
#endif
#if defined(atfork_unlock) && !defined(Perl_atfork_unlock)
#define Perl_atfork_unlock(...) GOPERL_LONG_CALL(atfork_unlock, __VA_ARGS__)
#endif
#if defined(av_arylen_p) && !defined(Perl_av_arylen_p)
#define Perl_av_arylen_p(...) GOPERL_LONG_CALL(av_arylen_p, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(av_fetch_simple) && !defined(Perl_av_fetch_simple)
#define Perl_av_fetch_simple(...) GOPERL_LONG_CALL(av_fetch_simple, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(AvFILL_) && !defined(Perl_AvFILL_)
#define Perl_AvFILL_(...) GOPERL_LONG_CALL(AvFILL_, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(av_iter_p) && !defined(Perl_av_iter_p)
#define Perl_av_iter_p(...) GOPERL_LONG_CALL(av_iter_p, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(av_new_alloc) && !defined(Perl_av_new_alloc)
#define Perl_av_new_alloc(...) GOPERL_LONG_CALL(av_new_alloc, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(av_reify) && !defined(Perl_av_reify)
#define Perl_av_reify(...) GOPERL_LONG_CALL(av_reify, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(av_store_simple) && !defined(Perl_av_store_simple)
#define Perl_av_store_simple(...) GOPERL_LONG_CALL(av_store_simple, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(block_gimme) && !defined(Perl_block_gimme)
#define Perl_block_gimme(...) GOPERL_LONG_CALL(block_gimme, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(byte_dump_string_) && !defined(Perl_byte_dump_string_)
#define Perl_byte_dump_string_(...) GOPERL_LONG_CALL(byte_dump_string_, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(call_list) && !defined(Perl_call_list)
#define Perl_call_list(...) GOPERL_LONG_CALL(call_list, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(cast_i32) && !defined(Perl_cast_i32)
#define Perl_cast_i32(...) GOPERL_LONG_CALL(cast_i32, __VA_ARGS__)
#endif
#if defined(cast_ulong) && !defined(Perl_cast_ulong)
#define Perl_cast_ulong(...) GOPERL_LONG_CALL(cast_ulong, __VA_ARGS__)
#endif
#if defined(clear_defarray) && !defined(Perl_clear_defarray)
#define Perl_clear_defarray(...) GOPERL_LONG_CALL(clear_defarray, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(clear_defarray_simple) && !defined(Perl_clear_defarray_simple)
#define Perl_clear_defarray_simple(...) GOPERL_LONG_CALL(clear_defarray_simple, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(cntrl_to_mnemonic) && !defined(Perl_cntrl_to_mnemonic)
#define Perl_cntrl_to_mnemonic(...) GOPERL_LONG_CALL(cntrl_to_mnemonic, __VA_ARGS__)
#endif
#if defined(croak_popstack) && !defined(Perl_croak_popstack)
#define Perl_croak_popstack(...) GOPERL_LONG_CALL(croak_popstack, __VA_ARGS__)
#endif
#if defined(csighandler1) && !defined(Perl_csighandler1)
#define Perl_csighandler1(...) GOPERL_LONG_CALL(csighandler1, __VA_ARGS__)
#endif
#if defined(csighandler3) && !defined(Perl_csighandler3)
#define Perl_csighandler3(...) GOPERL_LONG_CALL(csighandler3, __VA_ARGS__)
#endif
#if defined(current_re_engine) && !defined(Perl_current_re_engine)
#define Perl_current_re_engine(...) GOPERL_LONG_CALL(current_re_engine, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(custom_op_get_field) && !defined(Perl_custom_op_get_field)
#define Perl_custom_op_get_field(...) GOPERL_LONG_CALL(custom_op_get_field, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(cv_ckproto_len_flags) && !defined(Perl_cv_ckproto_len_flags)
#define Perl_cv_ckproto_len_flags(...) GOPERL_LONG_CALL(cv_ckproto_len_flags, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(cvgv_from_hek) && !defined(Perl_cvgv_from_hek)
#define Perl_cvgv_from_hek(...) GOPERL_LONG_CALL(cvgv_from_hek, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(cvgv_set) && !defined(Perl_cvgv_set)
#define Perl_cvgv_set(...) GOPERL_LONG_CALL(cvgv_set, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(cvstash_set) && !defined(Perl_cvstash_set)
#define Perl_cvstash_set(...) GOPERL_LONG_CALL(cvstash_set, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(cx_dump) && !defined(Perl_cx_dump)
#define Perl_cx_dump(...) GOPERL_LONG_CALL(cx_dump, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(debop) && !defined(Perl_debop)
#define Perl_debop(...) GOPERL_LONG_CALL(debop, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(debprofdump) && !defined(Perl_debprofdump)
#define Perl_debprofdump(...) GOPERL_LONG_CALL(debprofdump, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(debstackptrs) && !defined(Perl_debstackptrs)
#define Perl_debstackptrs(...) GOPERL_LONG_CALL(debstackptrs, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(delimcpy_no_escape) && !defined(Perl_delimcpy_no_escape)
#define Perl_delimcpy_no_escape(...) GOPERL_LONG_CALL(delimcpy_no_escape, __VA_ARGS__)
#endif
#if defined(do_gv_dump) && !defined(Perl_do_gv_dump)
#define Perl_do_gv_dump(...) GOPERL_LONG_CALL(do_gv_dump, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(do_gvgv_dump) && !defined(Perl_do_gvgv_dump)
#define Perl_do_gvgv_dump(...) GOPERL_LONG_CALL(do_gvgv_dump, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(do_hv_dump) && !defined(Perl_do_hv_dump)
#define Perl_do_hv_dump(...) GOPERL_LONG_CALL(do_hv_dump, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(doing_taint) && !defined(Perl_doing_taint)
#define Perl_doing_taint(...) GOPERL_LONG_CALL(doing_taint, __VA_ARGS__)
#endif
#if defined(do_magic_dump) && !defined(Perl_do_magic_dump)
#define Perl_do_magic_dump(...) GOPERL_LONG_CALL(do_magic_dump, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(do_op_dump) && !defined(Perl_do_op_dump)
#define Perl_do_op_dump(...) GOPERL_LONG_CALL(do_op_dump, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(do_pmop_dump) && !defined(Perl_do_pmop_dump)
#define Perl_do_pmop_dump(...) GOPERL_LONG_CALL(do_pmop_dump, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(doref) && !defined(Perl_doref)
#define Perl_doref(...) GOPERL_LONG_CALL(doref, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(do_sv_dump) && !defined(Perl_do_sv_dump)
#define Perl_do_sv_dump(...) GOPERL_LONG_CALL(do_sv_dump, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(do_uniprop_match) && !defined(Perl_do_uniprop_match)
#define Perl_do_uniprop_match(...) GOPERL_LONG_CALL(do_uniprop_match, __VA_ARGS__)
#endif
#if defined(dowantarray) && !defined(Perl_dowantarray)
#define Perl_dowantarray(...) GOPERL_LONG_CALL(dowantarray, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(drand48_init_r) && !defined(Perl_drand48_init_r)
#define Perl_drand48_init_r(...) GOPERL_LONG_CALL(drand48_init_r, __VA_ARGS__)
#endif
#if defined(drand48_r) && !defined(Perl_drand48_r)
#define Perl_drand48_r(...) GOPERL_LONG_CALL(drand48_r, __VA_ARGS__)
#endif
#if defined(dump_indent) && !defined(Perl_dump_indent)
#define Perl_dump_indent(...) GOPERL_LONG_CALL(dump_indent, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(dump_vindent) && !defined(Perl_dump_vindent)
#define Perl_dump_vindent(...) GOPERL_LONG_CALL(dump_vindent, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(dup_warnings) && !defined(Perl_dup_warnings)
#define Perl_dup_warnings(...) GOPERL_LONG_CALL(dup_warnings, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(emulate_cop_io) && !defined(Perl_emulate_cop_io)
#define Perl_emulate_cop_io(...) GOPERL_LONG_CALL(emulate_cop_io, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(foldEQ_latin1) && !defined(Perl_foldEQ_latin1)
#define Perl_foldEQ_latin1(...) GOPERL_LONG_CALL(foldEQ_latin1, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(foldEQ_utf8_flags) && !defined(Perl_foldEQ_utf8_flags)
#define Perl_foldEQ_utf8_flags(...) GOPERL_LONG_CALL(foldEQ_utf8_flags, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(force_out_malformed_utf8_message_) && !defined(Perl_force_out_malformed_utf8_message_)
#define Perl_force_out_malformed_utf8_message_(...) GOPERL_LONG_CALL(force_out_malformed_utf8_message_, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(free_tmps) && !defined(Perl_free_tmps)
#define Perl_free_tmps(...) GOPERL_LONG_CALL(free_tmps, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(get_and_check_backslash_N_name) && !defined(Perl_get_and_check_backslash_N_name)
#define Perl_get_and_check_backslash_N_name(...) GOPERL_LONG_CALL(get_and_check_backslash_N_name, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(get_deprecated_property_msg) && !defined(Perl_get_deprecated_property_msg)
#define Perl_get_deprecated_property_msg(...) GOPERL_LONG_CALL(get_deprecated_property_msg, __VA_ARGS__)
#endif
#if defined(get_ppaddr) && !defined(Perl_get_ppaddr)
#define Perl_get_ppaddr(...) GOPERL_LONG_CALL(get_ppaddr, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(get_prop_definition) && !defined(Perl_get_prop_definition)
#define Perl_get_prop_definition(...) GOPERL_LONG_CALL(get_prop_definition, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(get_prop_values) && !defined(Perl_get_prop_values)
#define Perl_get_prop_values(...) GOPERL_LONG_CALL(get_prop_values, __VA_ARGS__)
#endif
#if defined(get_re_arg) && !defined(Perl_get_re_arg)
#define Perl_get_re_arg(...) GOPERL_LONG_CALL(get_re_arg, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(get_vtbl) && !defined(Perl_get_vtbl)
#define Perl_get_vtbl(...) GOPERL_LONG_CALL(get_vtbl, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(gp_free) && !defined(Perl_gp_free)
#define Perl_gp_free(...) GOPERL_LONG_CALL(gp_free, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(gp_ref) && !defined(Perl_gp_ref)
#define Perl_gp_ref(...) GOPERL_LONG_CALL(gp_ref, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(grok_bin_hex) && !defined(Perl_grok_bin_hex)
#define Perl_grok_bin_hex(...) GOPERL_LONG_CALL(grok_bin_hex, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(grok_uint_by_base) && !defined(Perl_grok_uint_by_base)
#define Perl_grok_uint_by_base(...) GOPERL_LONG_CALL(grok_uint_by_base, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(gv_check) && !defined(Perl_gv_check)
#define Perl_gv_check(...) GOPERL_LONG_CALL(gv_check, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(gv_handler) && !defined(Perl_gv_handler)
#define Perl_gv_handler(...) GOPERL_LONG_CALL(gv_handler, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(gv_try_downgrade) && !defined(Perl_gv_try_downgrade)
#define Perl_gv_try_downgrade(...) GOPERL_LONG_CALL(gv_try_downgrade, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(hv_delayfree_ent) && !defined(Perl_hv_delayfree_ent)
#define Perl_hv_delayfree_ent(...) GOPERL_LONG_CALL(hv_delayfree_ent, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(hv_eiter_p) && !defined(Perl_hv_eiter_p)
#define Perl_hv_eiter_p(...) GOPERL_LONG_CALL(hv_eiter_p, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(hv_eiter_set) && !defined(Perl_hv_eiter_set)
#define Perl_hv_eiter_set(...) GOPERL_LONG_CALL(hv_eiter_set, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(hv_fill) && !defined(Perl_hv_fill)
#define Perl_hv_fill(...) GOPERL_LONG_CALL(hv_fill, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(hv_free_ent) && !defined(Perl_hv_free_ent)
#define Perl_hv_free_ent(...) GOPERL_LONG_CALL(hv_free_ent, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(hv_placeholders_get) && !defined(Perl_hv_placeholders_get)
#define Perl_hv_placeholders_get(...) GOPERL_LONG_CALL(hv_placeholders_get, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(hv_placeholders_p) && !defined(Perl_hv_placeholders_p)
#define Perl_hv_placeholders_p(...) GOPERL_LONG_CALL(hv_placeholders_p, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(hv_placeholders_set) && !defined(Perl_hv_placeholders_set)
#define Perl_hv_placeholders_set(...) GOPERL_LONG_CALL(hv_placeholders_set, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(hv_rand_set) && !defined(Perl_hv_rand_set)
#define Perl_hv_rand_set(...) GOPERL_LONG_CALL(hv_rand_set, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(hv_riter_p) && !defined(Perl_hv_riter_p)
#define Perl_hv_riter_p(...) GOPERL_LONG_CALL(hv_riter_p, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(hv_riter_set) && !defined(Perl_hv_riter_set)
#define Perl_hv_riter_set(...) GOPERL_LONG_CALL(hv_riter_set, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(hv_undef_flags) && !defined(Perl_hv_undef_flags)
#define Perl_hv_undef_flags(...) GOPERL_LONG_CALL(hv_undef_flags, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(init_i18nl10n) && !defined(Perl_init_i18nl10n)
#define Perl_init_i18nl10n(...) GOPERL_LONG_CALL(init_i18nl10n, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(init_named_cv) && !defined(Perl_init_named_cv)
#define Perl_init_named_cv(...) GOPERL_LONG_CALL(init_named_cv, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(init_stacks) && !defined(Perl_init_stacks)
#define Perl_init_stacks(...) GOPERL_LONG_CALL(init_stacks, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(init_tm) && !defined(Perl_init_tm)
#define Perl_init_tm(...) GOPERL_LONG_CALL(init_tm, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(inverse_folds_) && !defined(Perl_inverse_folds_)
#define Perl_inverse_folds_(...) GOPERL_LONG_CALL(inverse_folds_, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(is_in_locale_category_) && !defined(Perl_is_in_locale_category_)
#define Perl_is_in_locale_category_(...) GOPERL_LONG_CALL(is_in_locale_category_, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(is_uni_FOO_) && !defined(Perl_is_uni_FOO_)
#define Perl_is_uni_FOO_(...) GOPERL_LONG_CALL(is_uni_FOO_, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(is_uni_perl_idcont_) && !defined(Perl_is_uni_perl_idcont_)
#define Perl_is_uni_perl_idcont_(...) GOPERL_LONG_CALL(is_uni_perl_idcont_, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(is_uni_perl_idstart_) && !defined(Perl_is_uni_perl_idstart_)
#define Perl_is_uni_perl_idstart_(...) GOPERL_LONG_CALL(is_uni_perl_idstart_, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(is_utf8_char_helper_) && !defined(Perl_is_utf8_char_helper_)
#define Perl_is_utf8_char_helper_(...) GOPERL_LONG_CALL(is_utf8_char_helper_, __VA_ARGS__)
#endif
#if defined(is_utf8_FF_helper_) && !defined(Perl_is_utf8_FF_helper_)
#define Perl_is_utf8_FF_helper_(...) GOPERL_LONG_CALL(is_utf8_FF_helper_, __VA_ARGS__)
#endif
#if defined(is_utf8_FOO_) && !defined(Perl_is_utf8_FOO_)
#define Perl_is_utf8_FOO_(...) GOPERL_LONG_CALL(is_utf8_FOO_, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(is_utf8_perl_idcont_) && !defined(Perl_is_utf8_perl_idcont_)
#define Perl_is_utf8_perl_idcont_(...) GOPERL_LONG_CALL(is_utf8_perl_idcont_, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(is_utf8_perl_idstart_) && !defined(Perl_is_utf8_perl_idstart_)
#define Perl_is_utf8_perl_idstart_(...) GOPERL_LONG_CALL(is_utf8_perl_idstart_, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(load_charnames) && !defined(Perl_load_charnames)
#define Perl_load_charnames(...) GOPERL_LONG_CALL(load_charnames, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(locale_panic) && !defined(Perl_locale_panic)
#define Perl_locale_panic(...) GOPERL_LONG_CALL(locale_panic, __VA_ARGS__)
#endif
#if defined(long_valid_utf8_to_uv) && !defined(Perl_long_valid_utf8_to_uv)
#define Perl_long_valid_utf8_to_uv(...) GOPERL_LONG_CALL(long_valid_utf8_to_uv, __VA_ARGS__)
#endif
#if defined(lsbit_pos32) && !defined(Perl_lsbit_pos32)
#define Perl_lsbit_pos32(...) GOPERL_LONG_CALL(lsbit_pos32, __VA_ARGS__)
#endif
#if defined(markstack_grow) && !defined(Perl_markstack_grow)
#define Perl_markstack_grow(...) GOPERL_LONG_CALL(markstack_grow, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(mbtowc_) && !defined(Perl_mbtowc_)
#define Perl_mbtowc_(...) GOPERL_LONG_CALL(mbtowc_, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(mg_find_mglob) && !defined(Perl_mg_find_mglob)
#define Perl_mg_find_mglob(...) GOPERL_LONG_CALL(mg_find_mglob, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(mg_size) && !defined(Perl_mg_size)
#define Perl_mg_size(...) GOPERL_LONG_CALL(mg_size, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(more_bodies) && !defined(Perl_more_bodies)
#define Perl_more_bodies(...) GOPERL_LONG_CALL(more_bodies, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(more_sv) && !defined(Perl_more_sv)
#define Perl_more_sv(...) GOPERL_LONG_CALL(more_sv, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(moreswitches) && !defined(Perl_moreswitches)
#define Perl_moreswitches(...) GOPERL_LONG_CALL(moreswitches, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(mortal_getenv) && !defined(Perl_mortal_getenv)
#define Perl_mortal_getenv(...) GOPERL_LONG_CALL(mortal_getenv, __VA_ARGS__)
#endif
#if defined(mortal_svfunc_x) && !defined(Perl_mortal_svfunc_x)
#define Perl_mortal_svfunc_x(...) GOPERL_LONG_CALL(mortal_svfunc_x, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(mro_get_private_data) && !defined(Perl_mro_get_private_data)
#define Perl_mro_get_private_data(...) GOPERL_LONG_CALL(mro_get_private_data, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(mro_meta_init) && !defined(Perl_mro_meta_init)
#define Perl_mro_meta_init(...) GOPERL_LONG_CALL(mro_meta_init, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(msbit_pos32) && !defined(Perl_msbit_pos32)
#define Perl_msbit_pos32(...) GOPERL_LONG_CALL(msbit_pos32, __VA_ARGS__)
#endif
#if defined(multiconcat_stringify) && !defined(Perl_multiconcat_stringify)
#define Perl_multiconcat_stringify(...) GOPERL_LONG_CALL(multiconcat_stringify, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(multideref_stringify) && !defined(Perl_multideref_stringify)
#define Perl_multideref_stringify(...) GOPERL_LONG_CALL(multideref_stringify, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(my_atof2) && !defined(Perl_my_atof2)
#define Perl_my_atof2(...) GOPERL_LONG_CALL(my_atof2, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(my_atof3) && !defined(Perl_my_atof3)
#define Perl_my_atof3(...) GOPERL_LONG_CALL(my_atof3, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(my_fflush_all) && !defined(Perl_my_fflush_all)
#define Perl_my_fflush_all(...) GOPERL_LONG_CALL(my_fflush_all, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(my_fork) && !defined(Perl_my_fork)
#define Perl_my_fork(...) GOPERL_LONG_CALL(my_fork, __VA_ARGS__)
#endif
#if defined(my_lstat_flags) && !defined(Perl_my_lstat_flags)
#define Perl_my_lstat_flags(...) GOPERL_LONG_CALL(my_lstat_flags, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(my_popen_list) && !defined(Perl_my_popen_list)
#define Perl_my_popen_list(...) GOPERL_LONG_CALL(my_popen_list, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(my_socketpair) && !defined(Perl_my_socketpair)
#define Perl_my_socketpair(...) GOPERL_LONG_CALL(my_socketpair, __VA_ARGS__)
#endif
#if defined(my_stat_flags) && !defined(Perl_my_stat_flags)
#define Perl_my_stat_flags(...) GOPERL_LONG_CALL(my_stat_flags, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(newATTRSUB_x) && !defined(Perl_newATTRSUB_x)
#define Perl_newATTRSUB_x(...) GOPERL_LONG_CALL(newATTRSUB_x, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(newFORM) && !defined(Perl_newFORM)
#define Perl_newFORM(...) GOPERL_LONG_CALL(newFORM, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(newMYSUB) && !defined(Perl_newMYSUB)
#define Perl_newMYSUB(...) GOPERL_LONG_CALL(newMYSUB, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(newPROG) && !defined(Perl_newPROG)
#define Perl_newPROG(...) GOPERL_LONG_CALL(newPROG, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(new_sv) && !defined(Perl_new_sv)
#define Perl_new_sv(...) GOPERL_LONG_CALL(new_sv, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(noshutdownhook) && !defined(Perl_noshutdownhook)
#define Perl_noshutdownhook(...) GOPERL_LONG_CALL(noshutdownhook, __VA_ARGS__)
#endif
#if defined(op_lvalue_flags) && !defined(Perl_op_lvalue_flags)
#define Perl_op_lvalue_flags(...) GOPERL_LONG_CALL(op_lvalue_flags, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(op_refcnt_lock) && !defined(Perl_op_refcnt_lock)
#define Perl_op_refcnt_lock(...) GOPERL_LONG_CALL(op_refcnt_lock, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(op_refcnt_unlock) && !defined(Perl_op_refcnt_unlock)
#define Perl_op_refcnt_unlock(...) GOPERL_LONG_CALL(op_refcnt_unlock, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(padname_free) && !defined(Perl_padname_free)
#define Perl_padname_free(...) GOPERL_LONG_CALL(padname_free, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(padnamelist_free) && !defined(Perl_padnamelist_free)
#define Perl_padnamelist_free(...) GOPERL_LONG_CALL(padnamelist_free, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(parse_ident_msg) && !defined(Perl_parse_ident_msg)
#define Perl_parse_ident_msg(...) GOPERL_LONG_CALL(parse_ident_msg, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(PerlIO_context_layers) && !defined(Perl_PerlIO_context_layers)
#define Perl_PerlIO_context_layers(...) GOPERL_LONG_CALL(PerlIO_context_layers, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(perly_sighandler) && !defined(Perl_perly_sighandler)
#define Perl_perly_sighandler(...) GOPERL_LONG_CALL(perly_sighandler, __VA_ARGS__)
#endif
#if defined(POPMARK) && !defined(Perl_POPMARK)
#define Perl_POPMARK(...) GOPERL_LONG_CALL(POPMARK, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(pop_scope) && !defined(Perl_pop_scope)
#define Perl_pop_scope(...) GOPERL_LONG_CALL(pop_scope, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(pop_stackinfo) && !defined(Perl_pop_stackinfo)
#define Perl_pop_stackinfo(...) GOPERL_LONG_CALL(pop_stackinfo, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(pregfree) && !defined(Perl_pregfree)
#define Perl_pregfree(...) GOPERL_LONG_CALL(pregfree, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(pregfree2) && !defined(Perl_pregfree2)
#define Perl_pregfree2(...) GOPERL_LONG_CALL(pregfree2, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(push_scope) && !defined(Perl_push_scope)
#define Perl_push_scope(...) GOPERL_LONG_CALL(push_scope, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(push_stackinfo) && !defined(Perl_push_stackinfo)
#define Perl_push_stackinfo(...) GOPERL_LONG_CALL(push_stackinfo, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(ReANY) && !defined(Perl_ReANY)
#define Perl_ReANY(...) GOPERL_LONG_CALL(ReANY, __VA_ARGS__)
#endif
#if defined(reentrant_free) && !defined(Perl_reentrant_free)
#define Perl_reentrant_free(...) GOPERL_LONG_CALL(reentrant_free, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(reentrant_init) && !defined(Perl_reentrant_init)
#define Perl_reentrant_init(...) GOPERL_LONG_CALL(reentrant_init, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(reentrant_retry) && !defined(Perl_reentrant_retry)
#define Perl_reentrant_retry(...) GOPERL_LONG_CALL(reentrant_retry, __VA_ARGS__)
#endif
#if defined(reentrant_size) && !defined(Perl_reentrant_size)
#define Perl_reentrant_size(...) GOPERL_LONG_CALL(reentrant_size, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(refcounted_he_chain_2hv) && !defined(Perl_refcounted_he_chain_2hv)
#define Perl_refcounted_he_chain_2hv(...) GOPERL_LONG_CALL(refcounted_he_chain_2hv, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(refcounted_he_fetch_pv) && !defined(Perl_refcounted_he_fetch_pv)
#define Perl_refcounted_he_fetch_pv(...) GOPERL_LONG_CALL(refcounted_he_fetch_pv, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(refcounted_he_fetch_pvn) && !defined(Perl_refcounted_he_fetch_pvn)
#define Perl_refcounted_he_fetch_pvn(...) GOPERL_LONG_CALL(refcounted_he_fetch_pvn, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(refcounted_he_fetch_sv) && !defined(Perl_refcounted_he_fetch_sv)
#define Perl_refcounted_he_fetch_sv(...) GOPERL_LONG_CALL(refcounted_he_fetch_sv, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(refcounted_he_free) && !defined(Perl_refcounted_he_free)
#define Perl_refcounted_he_free(...) GOPERL_LONG_CALL(refcounted_he_free, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(refcounted_he_inc) && !defined(Perl_refcounted_he_inc)
#define Perl_refcounted_he_inc(...) GOPERL_LONG_CALL(refcounted_he_inc, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(refcounted_he_new_pv) && !defined(Perl_refcounted_he_new_pv)
#define Perl_refcounted_he_new_pv(...) GOPERL_LONG_CALL(refcounted_he_new_pv, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(refcounted_he_new_pvn) && !defined(Perl_refcounted_he_new_pvn)
#define Perl_refcounted_he_new_pvn(...) GOPERL_LONG_CALL(refcounted_he_new_pvn, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(refcounted_he_new_sv) && !defined(Perl_refcounted_he_new_sv)
#define Perl_refcounted_he_new_sv(...) GOPERL_LONG_CALL(refcounted_he_new_sv, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(regdump) && !defined(Perl_regdump)
#define Perl_regdump(...) GOPERL_LONG_CALL(regdump, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(regexec_flags) && !defined(Perl_regexec_flags)
#define Perl_regexec_flags(...) GOPERL_LONG_CALL(regexec_flags, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(regfree_internal) && !defined(Perl_regfree_internal)
#define Perl_regfree_internal(...) GOPERL_LONG_CALL(regfree_internal, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(reginitcolors) && !defined(Perl_reginitcolors)
#define Perl_reginitcolors(...) GOPERL_LONG_CALL(reginitcolors, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(reg_named_buff) && !defined(Perl_reg_named_buff)
#define Perl_reg_named_buff(...) GOPERL_LONG_CALL(reg_named_buff, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(reg_named_buff_all) && !defined(Perl_reg_named_buff_all)
#define Perl_reg_named_buff_all(...) GOPERL_LONG_CALL(reg_named_buff_all, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(reg_named_buff_exists) && !defined(Perl_reg_named_buff_exists)
#define Perl_reg_named_buff_exists(...) GOPERL_LONG_CALL(reg_named_buff_exists, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(reg_named_buff_fetch) && !defined(Perl_reg_named_buff_fetch)
#define Perl_reg_named_buff_fetch(...) GOPERL_LONG_CALL(reg_named_buff_fetch, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(reg_named_buff_firstkey) && !defined(Perl_reg_named_buff_firstkey)
#define Perl_reg_named_buff_firstkey(...) GOPERL_LONG_CALL(reg_named_buff_firstkey, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(reg_named_buff_iter) && !defined(Perl_reg_named_buff_iter)
#define Perl_reg_named_buff_iter(...) GOPERL_LONG_CALL(reg_named_buff_iter, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(reg_named_buff_nextkey) && !defined(Perl_reg_named_buff_nextkey)
#define Perl_reg_named_buff_nextkey(...) GOPERL_LONG_CALL(reg_named_buff_nextkey, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(reg_named_buff_scalar) && !defined(Perl_reg_named_buff_scalar)
#define Perl_reg_named_buff_scalar(...) GOPERL_LONG_CALL(reg_named_buff_scalar, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(reg_numbered_buff_fetch) && !defined(Perl_reg_numbered_buff_fetch)
#define Perl_reg_numbered_buff_fetch(...) GOPERL_LONG_CALL(reg_numbered_buff_fetch, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(reg_numbered_buff_fetch_flags) && !defined(Perl_reg_numbered_buff_fetch_flags)
#define Perl_reg_numbered_buff_fetch_flags(...) GOPERL_LONG_CALL(reg_numbered_buff_fetch_flags, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(reg_numbered_buff_length) && !defined(Perl_reg_numbered_buff_length)
#define Perl_reg_numbered_buff_length(...) GOPERL_LONG_CALL(reg_numbered_buff_length, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(reg_numbered_buff_store) && !defined(Perl_reg_numbered_buff_store)
#define Perl_reg_numbered_buff_store(...) GOPERL_LONG_CALL(reg_numbered_buff_store, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(reg_qr_package) && !defined(Perl_reg_qr_package)
#define Perl_reg_qr_package(...) GOPERL_LONG_CALL(reg_qr_package, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(reg_temp_copy) && !defined(Perl_reg_temp_copy)
#define Perl_reg_temp_copy(...) GOPERL_LONG_CALL(reg_temp_copy, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(re_intuit_start) && !defined(Perl_re_intuit_start)
#define Perl_re_intuit_start(...) GOPERL_LONG_CALL(re_intuit_start, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(re_intuit_string) && !defined(Perl_re_intuit_string)
#define Perl_re_intuit_string(...) GOPERL_LONG_CALL(re_intuit_string, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(re_op_compile) && !defined(Perl_re_op_compile)
#define Perl_re_op_compile(...) GOPERL_LONG_CALL(re_op_compile, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(report_uninit) && !defined(Perl_report_uninit)
#define Perl_report_uninit(...) GOPERL_LONG_CALL(report_uninit, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(resume_compcv) && !defined(Perl_resume_compcv)
#define Perl_resume_compcv(...) GOPERL_LONG_CALL(resume_compcv, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(rpp_free_2_) && !defined(Perl_rpp_free_2_)
#define Perl_rpp_free_2_(...) GOPERL_LONG_CALL(rpp_free_2_, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(rpp_obliterate_stack_to) && !defined(Perl_rpp_obliterate_stack_to)
#define Perl_rpp_obliterate_stack_to(...) GOPERL_LONG_CALL(rpp_obliterate_stack_to, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(rpp_replace_2_1_COMMON) && !defined(Perl_rpp_replace_2_1_COMMON)
#define Perl_rpp_replace_2_1_COMMON(...) GOPERL_LONG_CALL(rpp_replace_2_1_COMMON, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(runops_debug) && !defined(Perl_runops_debug)
#define Perl_runops_debug(...) GOPERL_LONG_CALL(runops_debug, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(runops_standard) && !defined(Perl_runops_standard)
#define Perl_runops_standard(...) GOPERL_LONG_CALL(runops_standard, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(save_adelete) && !defined(Perl_save_adelete)
#define Perl_save_adelete(...) GOPERL_LONG_CALL(save_adelete, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(save_bool) && !defined(Perl_save_bool)
#define Perl_save_bool(...) GOPERL_LONG_CALL(save_bool, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(save_clearsv) && !defined(Perl_save_clearsv)
#define Perl_save_clearsv(...) GOPERL_LONG_CALL(save_clearsv, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(save_delete) && !defined(Perl_save_delete)
#define Perl_save_delete(...) GOPERL_LONG_CALL(save_delete, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(save_destructor) && !defined(Perl_save_destructor)
#define Perl_save_destructor(...) GOPERL_LONG_CALL(save_destructor, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(save_destructor_x) && !defined(Perl_save_destructor_x)
#define Perl_save_destructor_x(...) GOPERL_LONG_CALL(save_destructor_x, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(save_freercpv) && !defined(Perl_save_freercpv)
#define Perl_save_freercpv(...) GOPERL_LONG_CALL(save_freercpv, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(save_generic_pvref) && !defined(Perl_save_generic_pvref)
#define Perl_save_generic_pvref(...) GOPERL_LONG_CALL(save_generic_pvref, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(save_generic_svref) && !defined(Perl_save_generic_svref)
#define Perl_save_generic_svref(...) GOPERL_LONG_CALL(save_generic_svref, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(save_hdelete) && !defined(Perl_save_hdelete)
#define Perl_save_hdelete(...) GOPERL_LONG_CALL(save_hdelete, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(save_hints) && !defined(Perl_save_hints)
#define Perl_save_hints(...) GOPERL_LONG_CALL(save_hints, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(save_I16) && !defined(Perl_save_I16)
#define Perl_save_I16(...) GOPERL_LONG_CALL(save_I16, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(save_I32) && !defined(Perl_save_I32)
#define Perl_save_I32(...) GOPERL_LONG_CALL(save_I32, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(save_I8) && !defined(Perl_save_I8)
#define Perl_save_I8(...) GOPERL_LONG_CALL(save_I8, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(save_int) && !defined(Perl_save_int)
#define Perl_save_int(...) GOPERL_LONG_CALL(save_int, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(save_iv) && !defined(Perl_save_iv)
#define Perl_save_iv(...) GOPERL_LONG_CALL(save_iv, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(save_padsv) && !defined(Perl_save_padsv)
#define Perl_save_padsv(...) GOPERL_LONG_CALL(save_padsv, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(save_padsv_and_mortalize) && !defined(Perl_save_padsv_and_mortalize)
#define Perl_save_padsv_and_mortalize(...) GOPERL_LONG_CALL(save_padsv_and_mortalize, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(save_pptr) && !defined(Perl_save_pptr)
#define Perl_save_pptr(...) GOPERL_LONG_CALL(save_pptr, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(save_rcpv) && !defined(Perl_save_rcpv)
#define Perl_save_rcpv(...) GOPERL_LONG_CALL(save_rcpv, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(save_re_context) && !defined(Perl_save_re_context)
#define Perl_save_re_context(...) GOPERL_LONG_CALL(save_re_context, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(save_set_svflags) && !defined(Perl_save_set_svflags)
#define Perl_save_set_svflags(...) GOPERL_LONG_CALL(save_set_svflags, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(save_shared_pvref) && !defined(Perl_save_shared_pvref)
#define Perl_save_shared_pvref(...) GOPERL_LONG_CALL(save_shared_pvref, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(save_sptr) && !defined(Perl_save_sptr)
#define Perl_save_sptr(...) GOPERL_LONG_CALL(save_sptr, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(save_strlen) && !defined(Perl_save_strlen)
#define Perl_save_strlen(...) GOPERL_LONG_CALL(save_strlen, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(save_vptr) && !defined(Perl_save_vptr)
#define Perl_save_vptr(...) GOPERL_LONG_CALL(save_vptr, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(scan_num) && !defined(Perl_scan_num)
#define Perl_scan_num(...) GOPERL_LONG_CALL(scan_num, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(scan_str) && !defined(Perl_scan_str)
#define Perl_scan_str(...) GOPERL_LONG_CALL(scan_str, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(scan_word) && !defined(Perl_scan_word)
#define Perl_scan_word(...) GOPERL_LONG_CALL(scan_word, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(set_context) && !defined(Perl_set_context)
#define Perl_set_context(...) GOPERL_LONG_CALL(set_context, __VA_ARGS__)
#endif
#if defined(set_numeric_standard) && !defined(Perl_set_numeric_standard)
#define Perl_set_numeric_standard(...) GOPERL_LONG_CALL(set_numeric_standard, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(set_numeric_underlying) && !defined(Perl_set_numeric_underlying)
#define Perl_set_numeric_underlying(...) GOPERL_LONG_CALL(set_numeric_underlying, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(share_hek) && !defined(Perl_share_hek)
#define Perl_share_hek(...) GOPERL_LONG_CALL(share_hek, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(single_1bit_pos32) && !defined(Perl_single_1bit_pos32)
#define Perl_single_1bit_pos32(...) GOPERL_LONG_CALL(single_1bit_pos32, __VA_ARGS__)
#endif
#if defined(skipspace_flags) && !defined(Perl_skipspace_flags)
#define Perl_skipspace_flags(...) GOPERL_LONG_CALL(skipspace_flags, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(Slab_Alloc) && !defined(Perl_Slab_Alloc)
#define Perl_Slab_Alloc(...) GOPERL_LONG_CALL(Slab_Alloc, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(Slab_Free) && !defined(Perl_Slab_Free)
#define Perl_Slab_Free(...) GOPERL_LONG_CALL(Slab_Free, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(stack_grow) && !defined(Perl_stack_grow)
#define Perl_stack_grow(...) GOPERL_LONG_CALL(stack_grow, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(str_to_version) && !defined(Perl_str_to_version)
#define Perl_str_to_version(...) GOPERL_LONG_CALL(str_to_version, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(sv_2bool_flags) && !defined(Perl_sv_2bool_flags)
#define Perl_sv_2bool_flags(...) GOPERL_LONG_CALL(sv_2bool_flags, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(sv_can_swipe_pv_buf) && !defined(Perl_sv_can_swipe_pv_buf)
#define Perl_sv_can_swipe_pv_buf(...) GOPERL_LONG_CALL(sv_can_swipe_pv_buf, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(sv_del_backref) && !defined(Perl_sv_del_backref)
#define Perl_sv_del_backref(...) GOPERL_LONG_CALL(sv_del_backref, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(sv_free2) && !defined(Perl_sv_free2)
#define Perl_sv_free2(...) GOPERL_LONG_CALL(sv_free2, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(sv_grow) && !defined(Perl_sv_grow)
#define Perl_sv_grow(...) GOPERL_LONG_CALL(sv_grow, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(sv_grow_fresh) && !defined(Perl_sv_grow_fresh)
#define Perl_sv_grow_fresh(...) GOPERL_LONG_CALL(sv_grow_fresh, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(sv_magicext_mglob) && !defined(Perl_sv_magicext_mglob)
#define Perl_sv_magicext_mglob(...) GOPERL_LONG_CALL(sv_magicext_mglob, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(sv_newref) && !defined(Perl_sv_newref)
#define Perl_sv_newref(...) GOPERL_LONG_CALL(sv_newref, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(sv_pvbyten_force) && !defined(Perl_sv_pvbyten_force)
#define Perl_sv_pvbyten_force(...) GOPERL_LONG_CALL(sv_pvbyten_force, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(SvPV_helper) && !defined(Perl_SvPV_helper)
#define Perl_SvPV_helper(...) GOPERL_LONG_CALL(SvPV_helper, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(sv_pvutf8n_force) && !defined(Perl_sv_pvutf8n_force)
#define Perl_sv_pvutf8n_force(...) GOPERL_LONG_CALL(sv_pvutf8n_force, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(sv_sethek) && !defined(Perl_sv_sethek)
#define Perl_sv_sethek(...) GOPERL_LONG_CALL(sv_sethek, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(sv_setpv_freshbuf) && !defined(Perl_sv_setpv_freshbuf)
#define Perl_sv_setpv_freshbuf(...) GOPERL_LONG_CALL(sv_setpv_freshbuf, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(sv_tainted) && !defined(Perl_sv_tainted)
#define Perl_sv_tainted(...) GOPERL_LONG_CALL(sv_tainted, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(sv_true) && !defined(Perl_sv_true)
#define Perl_sv_true(...) GOPERL_LONG_CALL(sv_true, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(SvTRUE_common) && !defined(Perl_SvTRUE_common)
#define Perl_SvTRUE_common(...) GOPERL_LONG_CALL(SvTRUE_common, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(sv_untaint) && !defined(Perl_sv_untaint)
#define Perl_sv_untaint(...) GOPERL_LONG_CALL(sv_untaint, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(switch_argstack) && !defined(Perl_switch_argstack)
#define Perl_switch_argstack(...) GOPERL_LONG_CALL(switch_argstack, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(sys_init) && !defined(Perl_sys_init)
#define Perl_sys_init(...) GOPERL_LONG_CALL(sys_init, __VA_ARGS__)
#endif
#if defined(sys_init3) && !defined(Perl_sys_init3)
#define Perl_sys_init3(...) GOPERL_LONG_CALL(sys_init3, __VA_ARGS__)
#endif
#if defined(sys_term) && !defined(Perl_sys_term)
#define Perl_sys_term(...) GOPERL_LONG_CALL(sys_term, __VA_ARGS__)
#endif
#if defined(taint_env) && !defined(Perl_taint_env)
#define Perl_taint_env(...) GOPERL_LONG_CALL(taint_env, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(taint_proper) && !defined(Perl_taint_proper)
#define Perl_taint_proper(...) GOPERL_LONG_CALL(taint_proper, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(tmps_grow_p) && !defined(Perl_tmps_grow_p)
#define Perl_tmps_grow_p(...) GOPERL_LONG_CALL(tmps_grow_p, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(TOPMARK) && !defined(Perl_TOPMARK)
#define Perl_TOPMARK(...) GOPERL_LONG_CALL(TOPMARK, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(to_uni_fold_flags_) && !defined(Perl_to_uni_fold_flags_)
#define Perl_to_uni_fold_flags_(...) GOPERL_LONG_CALL(to_uni_fold_flags_, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(to_uni_lower) && !defined(Perl_to_uni_lower)
#define Perl_to_uni_lower(...) GOPERL_LONG_CALL(to_uni_lower, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(to_uni_title) && !defined(Perl_to_uni_title)
#define Perl_to_uni_title(...) GOPERL_LONG_CALL(to_uni_title, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(to_uni_upper) && !defined(Perl_to_uni_upper)
#define Perl_to_uni_upper(...) GOPERL_LONG_CALL(to_uni_upper, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(to_utf8_fold_flags_) && !defined(Perl_to_utf8_fold_flags_)
#define Perl_to_utf8_fold_flags_(...) GOPERL_LONG_CALL(to_utf8_fold_flags_, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(to_utf8_lower_flags_) && !defined(Perl_to_utf8_lower_flags_)
#define Perl_to_utf8_lower_flags_(...) GOPERL_LONG_CALL(to_utf8_lower_flags_, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(to_utf8_title_flags_) && !defined(Perl_to_utf8_title_flags_)
#define Perl_to_utf8_title_flags_(...) GOPERL_LONG_CALL(to_utf8_title_flags_, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(to_utf8_upper_flags_) && !defined(Perl_to_utf8_upper_flags_)
#define Perl_to_utf8_upper_flags_(...) GOPERL_LONG_CALL(to_utf8_upper_flags_, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(try_amagic_bin) && !defined(Perl_try_amagic_bin)
#define Perl_try_amagic_bin(...) GOPERL_LONG_CALL(try_amagic_bin, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(try_amagic_un) && !defined(Perl_try_amagic_un)
#define Perl_try_amagic_un(...) GOPERL_LONG_CALL(try_amagic_un, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(unshare_hek) && !defined(Perl_unshare_hek)
#define Perl_unshare_hek(...) GOPERL_LONG_CALL(unshare_hek, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(unsharepvn) && !defined(Perl_unsharepvn)
#define Perl_unsharepvn(...) GOPERL_LONG_CALL(unsharepvn, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(utf8n_to_uvuni) && !defined(Perl_utf8n_to_uvuni)
#define Perl_utf8n_to_uvuni(...) GOPERL_LONG_CALL(utf8n_to_uvuni, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(utf8_to_bytes_) && !defined(Perl_utf8_to_bytes_)
#define Perl_utf8_to_bytes_(...) GOPERL_LONG_CALL(utf8_to_bytes_, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(utf16_to_utf8_base) && !defined(Perl_utf16_to_utf8_base)
#define Perl_utf16_to_utf8_base(...) GOPERL_LONG_CALL(utf16_to_utf8_base, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(utf8_to_utf16_base) && !defined(Perl_utf8_to_utf16_base)
#define Perl_utf8_to_utf16_base(...) GOPERL_LONG_CALL(utf8_to_utf16_base, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(utf8_to_uv_msgs_helper_) && !defined(Perl_utf8_to_uv_msgs_helper_)
#define Perl_utf8_to_uv_msgs_helper_(...) GOPERL_LONG_CALL(utf8_to_uv_msgs_helper_, __VA_ARGS__)
#endif
#if defined(utf8_to_uvuni) && !defined(Perl_utf8_to_uvuni)
#define Perl_utf8_to_uvuni(...) GOPERL_LONG_CALL(utf8_to_uvuni, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(uvoffuni_to_utf8_flags_msgs) && !defined(Perl_uvoffuni_to_utf8_flags_msgs)
#define Perl_uvoffuni_to_utf8_flags_msgs(...) GOPERL_LONG_CALL(uvoffuni_to_utf8_flags_msgs, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(uvuni_to_utf8) && !defined(Perl_uvuni_to_utf8)
#define Perl_uvuni_to_utf8(...) GOPERL_LONG_CALL(uvuni_to_utf8, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(validate_proto) && !defined(Perl_validate_proto)
#define Perl_validate_proto(...) GOPERL_LONG_CALL(validate_proto, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(valid_utf8_to_uv) && !defined(Perl_valid_utf8_to_uv)
#define Perl_valid_utf8_to_uv(...) GOPERL_LONG_CALL(valid_utf8_to_uv, __VA_ARGS__)
#endif
#if defined(variant_byte_number) && !defined(Perl_variant_byte_number)
#define Perl_variant_byte_number(...) GOPERL_LONG_CALL(variant_byte_number, __VA_ARGS__)
#endif
#if defined(vivify_defelem) && !defined(Perl_vivify_defelem)
#define Perl_vivify_defelem(...) GOPERL_LONG_CALL(vivify_defelem, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(was_lvalue_sub) && !defined(Perl_was_lvalue_sub)
#define Perl_was_lvalue_sub(...) GOPERL_LONG_CALL(was_lvalue_sub, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(xs_handshake) && !defined(Perl_xs_handshake)
#define Perl_xs_handshake(...) GOPERL_LONG_CALL(xs_handshake, __VA_ARGS__)
#endif
#if defined(yylex) && !defined(Perl_yylex)
#define Perl_yylex(...) GOPERL_LONG_CALL(yylex, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(pad_setsv) && !defined(Perl_pad_setsv)
#define Perl_pad_setsv(...) GOPERL_LONG_CALL(pad_setsv, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(pad_sv) && !defined(Perl_pad_sv)
#define Perl_pad_sv(...) GOPERL_LONG_CALL(pad_sv, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(set_padlist) && !defined(Perl_set_padlist)
#define Perl_set_padlist(...) GOPERL_LONG_CALL(set_padlist, __VA_ARGS__)
#endif
#if defined(sys_intern_clear) && !defined(Perl_sys_intern_clear)
#define Perl_sys_intern_clear(...) GOPERL_LONG_CALL(sys_intern_clear, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(sys_intern_init) && !defined(Perl_sys_intern_init)
#define Perl_sys_intern_init(...) GOPERL_LONG_CALL(sys_intern_init, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(sys_intern_dup) && !defined(Perl_sys_intern_dup)
#define Perl_sys_intern_dup(...) GOPERL_LONG_CALL(sys_intern_dup, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(my_cxt_init) && !defined(Perl_my_cxt_init)
#define Perl_my_cxt_init(...) GOPERL_LONG_CALL(my_cxt_init, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(get_mstats) && !defined(Perl_get_mstats)
#define Perl_get_mstats(...) GOPERL_LONG_CALL(get_mstats, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(sv_setsv_cow) && !defined(Perl_sv_setsv_cow)
#define Perl_sv_setsv_cow(...) GOPERL_LONG_CALL(sv_setsv_cow, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(isSCRIPT_RUN) && !defined(Perl_isSCRIPT_RUN)
#define Perl_isSCRIPT_RUN(...) GOPERL_LONG_CALL(isSCRIPT_RUN, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(is_utf8_non_invariant_string) && !defined(Perl_is_utf8_non_invariant_string)
#define Perl_is_utf8_non_invariant_string(...) GOPERL_LONG_CALL(is_utf8_non_invariant_string, __VA_ARGS__)
#endif
#if defined(class_add_ADJUST) && !defined(Perl_class_add_ADJUST)
#define Perl_class_add_ADJUST(...) GOPERL_LONG_CALL(class_add_ADJUST, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(class_add_field) && !defined(Perl_class_add_field)
#define Perl_class_add_field(...) GOPERL_LONG_CALL(class_add_field, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(class_apply_attributes) && !defined(Perl_class_apply_attributes)
#define Perl_class_apply_attributes(...) GOPERL_LONG_CALL(class_apply_attributes, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(class_apply_field_attributes) && !defined(Perl_class_apply_field_attributes)
#define Perl_class_apply_field_attributes(...) GOPERL_LONG_CALL(class_apply_field_attributes, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(class_prepare_initfield_parse) && !defined(Perl_class_prepare_initfield_parse)
#define Perl_class_prepare_initfield_parse(...) GOPERL_LONG_CALL(class_prepare_initfield_parse, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(class_prepare_method_parse) && !defined(Perl_class_prepare_method_parse)
#define Perl_class_prepare_method_parse(...) GOPERL_LONG_CALL(class_prepare_method_parse, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(class_seal_stash) && !defined(Perl_class_seal_stash)
#define Perl_class_seal_stash(...) GOPERL_LONG_CALL(class_seal_stash, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(class_set_field_defop) && !defined(Perl_class_set_field_defop)
#define Perl_class_set_field_defop(...) GOPERL_LONG_CALL(class_set_field_defop, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(class_setup_stash) && !defined(Perl_class_setup_stash)
#define Perl_class_setup_stash(...) GOPERL_LONG_CALL(class_setup_stash, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(class_wrap_method_body) && !defined(Perl_class_wrap_method_body)
#define Perl_class_wrap_method_body(...) GOPERL_LONG_CALL(class_wrap_method_body, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(croak_kw_unless_class) && !defined(Perl_croak_kw_unless_class)
#define Perl_croak_kw_unless_class(...) GOPERL_LONG_CALL(croak_kw_unless_class, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(invlist_search_) && !defined(Perl_invlist_search_)
#define Perl_invlist_search_(...) GOPERL_LONG_CALL(invlist_search_, __VA_ARGS__)
#endif
#if defined(add_range_to_invlist_) && !defined(Perl_add_range_to_invlist_)
#define Perl_add_range_to_invlist_(...) GOPERL_LONG_CALL(add_range_to_invlist_, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(invlist_intersection_maybe_complement_2nd_) && !defined(Perl_invlist_intersection_maybe_complement_2nd_)
#define Perl_invlist_intersection_maybe_complement_2nd_(...) GOPERL_LONG_CALL(invlist_intersection_maybe_complement_2nd_, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(invlist_invert_) && !defined(Perl_invlist_invert_)
#define Perl_invlist_invert_(...) GOPERL_LONG_CALL(invlist_invert_, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(invlist_union_maybe_complement_2nd_) && !defined(Perl_invlist_union_maybe_complement_2nd_)
#define Perl_invlist_union_maybe_complement_2nd_(...) GOPERL_LONG_CALL(invlist_union_maybe_complement_2nd_, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(new_invlist_) && !defined(Perl_new_invlist_)
#define Perl_new_invlist_(...) GOPERL_LONG_CALL(new_invlist_, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(setup_canned_invlist_) && !defined(Perl_setup_canned_invlist_)
#define Perl_setup_canned_invlist_(...) GOPERL_LONG_CALL(setup_canned_invlist_, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(form_alien_digit_msg) && !defined(Perl_form_alien_digit_msg)
#define Perl_form_alien_digit_msg(...) GOPERL_LONG_CALL(form_alien_digit_msg, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(grok_bslash_c) && !defined(Perl_grok_bslash_c)
#define Perl_grok_bslash_c(...) GOPERL_LONG_CALL(grok_bslash_c, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(grok_bslash_o) && !defined(Perl_grok_bslash_o)
#define Perl_grok_bslash_o(...) GOPERL_LONG_CALL(grok_bslash_o, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(grok_bslash_x) && !defined(Perl_grok_bslash_x)
#define Perl_grok_bslash_x(...) GOPERL_LONG_CALL(grok_bslash_x, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(form_cp_too_large_msg) && !defined(Perl_form_cp_too_large_msg)
#define Perl_form_cp_too_large_msg(...) GOPERL_LONG_CALL(form_cp_too_large_msg, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(invlist_dump_) && !defined(Perl_invlist_dump_)
#define Perl_invlist_dump_(...) GOPERL_LONG_CALL(invlist_dump_, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(invlistEQ_) && !defined(Perl_invlistEQ_)
#define Perl_invlistEQ_(...) GOPERL_LONG_CALL(invlistEQ_, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(new_invlist_C_array_) && !defined(Perl_new_invlist_C_array_)
#define Perl_new_invlist_C_array_(...) GOPERL_LONG_CALL(new_invlist_C_array_, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(invlist_clone) && !defined(Perl_invlist_clone)
#define Perl_invlist_clone(...) GOPERL_LONG_CALL(invlist_clone, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(to_fold_latin1_) && !defined(Perl_to_fold_latin1_)
#define Perl_to_fold_latin1_(...) GOPERL_LONG_CALL(to_fold_latin1_, __VA_ARGS__)
#endif
#if defined(regcurly) && !defined(Perl_regcurly)
#define Perl_regcurly(...) GOPERL_LONG_CALL(regcurly, __VA_ARGS__)
#endif
#if defined(check_regnode_after) && !defined(Perl_check_regnode_after)
#define Perl_check_regnode_after(...) GOPERL_LONG_CALL(check_regnode_after, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(regnext) && !defined(Perl_regnext)
#define Perl_regnext(...) GOPERL_LONG_CALL(regnext, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(regnode_after) && !defined(Perl_regnode_after)
#define Perl_regnode_after(...) GOPERL_LONG_CALL(regnode_after, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(mem_log_alloc) && !defined(Perl_mem_log_alloc)
#define Perl_mem_log_alloc(...) GOPERL_LONG_CALL(mem_log_alloc, __VA_ARGS__)
#endif
#if defined(mem_log_del_sv) && !defined(Perl_mem_log_del_sv)
#define Perl_mem_log_del_sv(...) GOPERL_LONG_CALL(mem_log_del_sv, __VA_ARGS__)
#endif
#if defined(mem_log_free) && !defined(Perl_mem_log_free)
#define Perl_mem_log_free(...) GOPERL_LONG_CALL(mem_log_free, __VA_ARGS__)
#endif
#if defined(mem_log_new_sv) && !defined(Perl_mem_log_new_sv)
#define Perl_mem_log_new_sv(...) GOPERL_LONG_CALL(mem_log_new_sv, __VA_ARGS__)
#endif
#if defined(mem_log_realloc) && !defined(Perl_mem_log_realloc)
#define Perl_mem_log_realloc(...) GOPERL_LONG_CALL(mem_log_realloc, __VA_ARGS__)
#endif
#if defined(cx_popblock) && !defined(Perl_cx_popblock)
#define Perl_cx_popblock(...) GOPERL_LONG_CALL(cx_popblock, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(cx_popeval) && !defined(Perl_cx_popeval)
#define Perl_cx_popeval(...) GOPERL_LONG_CALL(cx_popeval, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(cx_popformat) && !defined(Perl_cx_popformat)
#define Perl_cx_popformat(...) GOPERL_LONG_CALL(cx_popformat, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(cx_popgiven) && !defined(Perl_cx_popgiven)
#define Perl_cx_popgiven(...) GOPERL_LONG_CALL(cx_popgiven, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(cx_poploop) && !defined(Perl_cx_poploop)
#define Perl_cx_poploop(...) GOPERL_LONG_CALL(cx_poploop, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(cx_popsub) && !defined(Perl_cx_popsub)
#define Perl_cx_popsub(...) GOPERL_LONG_CALL(cx_popsub, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(cx_popsub_args) && !defined(Perl_cx_popsub_args)
#define Perl_cx_popsub_args(...) GOPERL_LONG_CALL(cx_popsub_args, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(cx_popsub_common) && !defined(Perl_cx_popsub_common)
#define Perl_cx_popsub_common(...) GOPERL_LONG_CALL(cx_popsub_common, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(cx_popwhen) && !defined(Perl_cx_popwhen)
#define Perl_cx_popwhen(...) GOPERL_LONG_CALL(cx_popwhen, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(cx_pushblock) && !defined(Perl_cx_pushblock)
#define Perl_cx_pushblock(...) GOPERL_LONG_CALL(cx_pushblock, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(cx_pusheval) && !defined(Perl_cx_pusheval)
#define Perl_cx_pusheval(...) GOPERL_LONG_CALL(cx_pusheval, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(cx_pushformat) && !defined(Perl_cx_pushformat)
#define Perl_cx_pushformat(...) GOPERL_LONG_CALL(cx_pushformat, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(cx_pushgiven) && !defined(Perl_cx_pushgiven)
#define Perl_cx_pushgiven(...) GOPERL_LONG_CALL(cx_pushgiven, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(cx_pushloop_for) && !defined(Perl_cx_pushloop_for)
#define Perl_cx_pushloop_for(...) GOPERL_LONG_CALL(cx_pushloop_for, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(cx_pushloop_plain) && !defined(Perl_cx_pushloop_plain)
#define Perl_cx_pushloop_plain(...) GOPERL_LONG_CALL(cx_pushloop_plain, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(cx_pushsub) && !defined(Perl_cx_pushsub)
#define Perl_cx_pushsub(...) GOPERL_LONG_CALL(cx_pushsub, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(cx_pushtry) && !defined(Perl_cx_pushtry)
#define Perl_cx_pushtry(...) GOPERL_LONG_CALL(cx_pushtry, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(cx_pushwhen) && !defined(Perl_cx_pushwhen)
#define Perl_cx_pushwhen(...) GOPERL_LONG_CALL(cx_pushwhen, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(cx_topblock) && !defined(Perl_cx_topblock)
#define Perl_cx_topblock(...) GOPERL_LONG_CALL(cx_topblock, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(gimme_V) && !defined(Perl_gimme_V)
#define Perl_gimme_V(...) GOPERL_LONG_CALL(gimme_V, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(pp_wrap) && !defined(Perl_pp_wrap)
#define Perl_pp_wrap(...) GOPERL_LONG_CALL(pp_wrap, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(runops_wrap) && !defined(Perl_runops_wrap)
#define Perl_runops_wrap(...) GOPERL_LONG_CALL(runops_wrap, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(xs_wrap) && !defined(Perl_xs_wrap)
#define Perl_xs_wrap(...) GOPERL_LONG_CALL(xs_wrap, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(csighandler) && !defined(Perl_csighandler)
#define Perl_csighandler(...) GOPERL_LONG_CALL(csighandler, __VA_ARGS__)
#endif
#if defined(lsbit_pos64) && !defined(Perl_lsbit_pos64)
#define Perl_lsbit_pos64(...) GOPERL_LONG_CALL(lsbit_pos64, __VA_ARGS__)
#endif
#if defined(msbit_pos64) && !defined(Perl_msbit_pos64)
#define Perl_msbit_pos64(...) GOPERL_LONG_CALL(msbit_pos64, __VA_ARGS__)
#endif
#if defined(single_1bit_pos64) && !defined(Perl_single_1bit_pos64)
#define Perl_single_1bit_pos64(...) GOPERL_LONG_CALL(single_1bit_pos64, __VA_ARGS__)
#endif
#if defined(unlnk) && !defined(Perl_unlnk)
#define Perl_unlnk(...) GOPERL_LONG_CALL(unlnk, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(dtrace_probe_call) && !defined(Perl_dtrace_probe_call)
#define Perl_dtrace_probe_call(...) GOPERL_LONG_CALL(dtrace_probe_call, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(dtrace_probe_load) && !defined(Perl_dtrace_probe_load)
#define Perl_dtrace_probe_load(...) GOPERL_LONG_CALL(dtrace_probe_load, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(dtrace_probe_op) && !defined(Perl_dtrace_probe_op)
#define Perl_dtrace_probe_op(...) GOPERL_LONG_CALL(dtrace_probe_op, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(dtrace_probe_phase) && !defined(Perl_dtrace_probe_phase)
#define Perl_dtrace_probe_phase(...) GOPERL_LONG_CALL(dtrace_probe_phase, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(any_dup) && !defined(Perl_any_dup)
#define Perl_any_dup(...) GOPERL_LONG_CALL(any_dup, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(cop_file_avn) && !defined(Perl_cop_file_avn)
#define Perl_cop_file_avn(...) GOPERL_LONG_CALL(cop_file_avn, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(cx_dup) && !defined(Perl_cx_dup)
#define Perl_cx_dup(...) GOPERL_LONG_CALL(cx_dup, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(dirp_dup) && !defined(Perl_dirp_dup)
#define Perl_dirp_dup(...) GOPERL_LONG_CALL(dirp_dup, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(fp_dup) && !defined(Perl_fp_dup)
#define Perl_fp_dup(...) GOPERL_LONG_CALL(fp_dup, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(gp_dup) && !defined(Perl_gp_dup)
#define Perl_gp_dup(...) GOPERL_LONG_CALL(gp_dup, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(he_dup) && !defined(Perl_he_dup)
#define Perl_he_dup(...) GOPERL_LONG_CALL(he_dup, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(hek_dup) && !defined(Perl_hek_dup)
#define Perl_hek_dup(...) GOPERL_LONG_CALL(hek_dup, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(mg_dup) && !defined(Perl_mg_dup)
#define Perl_mg_dup(...) GOPERL_LONG_CALL(mg_dup, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(parser_dup) && !defined(Perl_parser_dup)
#define Perl_parser_dup(...) GOPERL_LONG_CALL(parser_dup, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(regdupe_internal) && !defined(Perl_regdupe_internal)
#define Perl_regdupe_internal(...) GOPERL_LONG_CALL(regdupe_internal, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(rvpv_dup) && !defined(Perl_rvpv_dup)
#define Perl_rvpv_dup(...) GOPERL_LONG_CALL(rvpv_dup, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(si_dup) && !defined(Perl_si_dup)
#define Perl_si_dup(...) GOPERL_LONG_CALL(si_dup, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(ss_dup) && !defined(Perl_ss_dup)
#define Perl_ss_dup(...) GOPERL_LONG_CALL(ss_dup, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(strxfrm) && !defined(Perl_strxfrm)
#define Perl_strxfrm(...) GOPERL_LONG_CALL(strxfrm, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(warn_problematic_locale) && !defined(Perl_warn_problematic_locale)
#define Perl_warn_problematic_locale(...) GOPERL_LONG_CALL(warn_problematic_locale, __VA_ARGS__)
#endif
#if defined(PerlIO_restore_errno) && !defined(Perl_PerlIO_restore_errno)
#define Perl_PerlIO_restore_errno(...) GOPERL_LONG_CALL(PerlIO_restore_errno, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(PerlIO_save_errno) && !defined(Perl_PerlIO_save_errno)
#define Perl_PerlIO_save_errno(...) GOPERL_LONG_CALL(PerlIO_save_errno, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(switch_locale_context) && !defined(Perl_switch_locale_context)
#define Perl_switch_locale_context(...) GOPERL_LONG_CALL(switch_locale_context, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(do_aspawn) && !defined(Perl_do_aspawn)
#define Perl_do_aspawn(...) GOPERL_LONG_CALL(do_aspawn, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(do_spawn) && !defined(Perl_do_spawn)
#define Perl_do_spawn(...) GOPERL_LONG_CALL(do_spawn, GOPERL_THX_DROP(__VA_ARGS__))
#endif
#if defined(do_spawn_nowait) && !defined(Perl_do_spawn_nowait)
#define Perl_do_spawn_nowait(...) GOPERL_LONG_CALL(do_spawn_nowait, GOPERL_THX_DROP(__VA_ARGS__))
#endif

#endif /* GOPERL_XS_SDK_LONGNAMES_CORE_H */
