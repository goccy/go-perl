/* Exercises the perl API surface of the SDK (goperl_perlapi.h) the way XS
 * distributions use it. Written directly in the shape xsubpp generates. */
#include "EXTERN.h"
#include "perl.h"
#include "XSUB.h"
#include "patchlevel.h"

/* The names of the value flags set on an SV, from the real flag word. */
XS(XS_PerlAPI_flags) {
    dXSARGS;
    if (items != 1) croak_xs_usage(cv, "sv");
    {
        U32 f = SvFLAGS(ST(0));
        SV *out = sv_2mortal(newSVpvs(""));
        if (f & SVf_IOK) sv_catpvs(out, "IOK,");
        if (f & SVf_NOK) sv_catpvs(out, "NOK,");
        if (f & SVf_POK) sv_catpvs(out, "POK,");
        if (f & SVf_ROK) sv_catpvs(out, "ROK,");
        if (f & SVf_UTF8) sv_catpvs(out, "UTF8,");
        if (f & SVf_READONLY) sv_catpvs(out, "READONLY,");
        if (f & SVs_OBJECT) sv_catpvs(out, "OBJECT,");
        if (f & SVs_GMG) sv_catpvs(out, "GMG,");
        ST(0) = out;
        XSRETURN(1);
    }
}

/* A defined non-reference: the mask idiom over SVf_OK. */
XS(XS_PerlAPI_is_plain_value) {
    dXSARGS;
    if (items != 1) croak_xs_usage(cv, "sv");
    if (SvMAGICAL(ST(0))) mg_get(ST(0));
    ST(0) = boolSV(SvFLAGS(ST(0)) & (SVf_OK & ~SVf_ROK));
    XSRETURN(1);
}

/* Drops the numeric value of a dual-valued scalar, keeping the string. */
XS(XS_PerlAPI_string_only) {
    dXSARGS;
    if (items != 1) croak_xs_usage(cv, "sv");
    SvPOK_only(ST(0));
    XSRETURN(1);
}

/* Makes a scalar undefined without assigning to it. */
XS(XS_PerlAPI_ok_off) {
    dXSARGS;
    if (items != 1) croak_xs_usage(cv, "sv");
    SvOK_off(ST(0));
    XSRETURN(1);
}

/* One get-magic call, then two reads that must not call it again. */
XS(XS_PerlAPI_read_twice_nomg) {
    dXSARGS;
    if (items != 1) croak_xs_usage(cv, "sv");
    {
        IV a, b;
        SvGETMAGIC(ST(0));
        a = SvIV_nomg(ST(0));
        b = SvIV_nomg(ST(0));
        XSRETURN_IV(a + b);
    }
}

XS(XS_PerlAPI_string_nomg) {
    dXSARGS;
    if (items != 1) croak_xs_usage(cv, "sv");
    {
        STRLEN len;
        const char *p;
        SvGETMAGIC(ST(0));
        p = SvPV_nomg_const(ST(0), len);
        ST(0) = sv_2mortal(newSVpvn(p, len));
        XSRETURN(1);
    }
}

XS(XS_PerlAPI_new_object) {
    dXSARGS;
    if (items != 2) croak_xs_usage(cv, "class, payload");
    {
        STRLEN len;
        const char *p = SvPV(ST(1), len);
        SV *rv = sv_newmortal();
        sv_setref_pvn(rv, SvPV_nolen(ST(0)), p, len);
        ST(0) = rv;
        XSRETURN(1);
    }
}

/* grok_hex: the value, how many bytes were consumed, the result flags. */
XS(XS_PerlAPI_hex) {
    dXSARGS;
    if (items != 1) croak_xs_usage(cv, "string");
    {
        STRLEN len;
        const char *p = SvPV(ST(0), len);
        I32 flags = PERL_SCAN_ALLOW_UNDERSCORES | PERL_SCAN_SILENT_ILLDIGIT;
        NV nv = 0;
        UV uv = grok_hex(p, &len, &flags, &nv);
        SP -= items;
        EXTEND(SP, 3);
        if (flags & PERL_SCAN_GREATER_THAN_UV_MAX)
            mPUSHn(nv);
        else
            mPUSHu(uv);
        mPUSHu(len);
        mPUSHi(flags);
        PUTBACK;
        return;
    }
}

XS(XS_PerlAPI_errsv_via_errgv) {
    dXSARGS;
    PERL_UNUSED_VAR(items);
    ST(0) = sv_2mortal(newSVsv(GvSVn(PL_errgv)));
    XSRETURN(1);
}

XS(XS_PerlAPI_call_with_argv) {
    dXSARGS;
    if (items != 1) croak_xs_usage(cv, "name");
    {
        char a[] = "left", b[] = "right";
        char *argv[] = {a, b, NULL};
        I32 n = call_argv(SvPV_nolen(ST(0)), G_SCALAR, argv);
        SPAGAIN;
        if (n != 1) croak("expected one value, got %d", (int)n);
        ST(0) = POPs;
        XSRETURN(1);
    }
}

XS(XS_PerlAPI_require) {
    dXSARGS;
    if (items != 1) croak_xs_usage(cv, "file");
    require_pv(SvPV_nolen(ST(0)));
    ST(0) = boolSV(!SvTRUE(ERRSV));
    XSRETURN(1);
}

XS(XS_PerlAPI_glob_name) {
    dXSARGS;
    if (items != 2) croak_xs_usage(cv, "globref, keepmain");
    {
        SV *out = sv_newmortal();
        gv_fullname4(out, (GV *)SvRV(ST(0)), "*", SvTRUE(ST(1)));
        ST(0) = out;
        XSRETURN(1);
    }
}

XS(XS_PerlAPI_copy_nosteal) {
    dXSARGS;
    if (items != 1) croak_xs_usage(cv, "sv");
    {
        SV *dst = sv_newmortal();
        SvSetSV_nosteal(dst, ST(0));
        ST(0) = dst;
        XSRETURN(1);
    }
}

XS(XS_PerlAPI_big) {
    dXSARGS;
    PERL_UNUSED_VAR(items);
    XSRETURN_UV(4000000000u);
}

XS(XS_PerlAPI_sub_kind) {
    dXSARGS;
    if (items != 1) croak_xs_usage(cv, "coderef");
    {
        CV *code = (CV *)SvRV(ST(0));
        SV *out = sv_2mortal(newSVpvs(""));
        if (CvANON(code)) sv_catpvs(out, "anon,");
        if (CvCONST(code)) sv_catpvs(out, "const,");
        if (CvISXSUB(code)) sv_catpvs(out, "xsub,");
        if (CvLVALUE(code)) sv_catpvs(out, "lvalue,");
        ST(0) = out;
        XSRETURN(1);
    }
}

/* A callback run "on its own stack", as tie and sort helpers do. */
XS(XS_PerlAPI_call_on_new_stack) {
    dXSARGS;
    if (items != 2) croak_xs_usage(cv, "code, arg");
    {
        SV *code = ST(0), *arg = ST(1), *ret;
        ENTER;
        SAVETMPS;
        PUSHSTACKi(PERLSI_MAGIC);
        PUSHMARK(SP);
        XPUSHs(arg);
        PUTBACK;
        call_sv(code, G_SCALAR);
        SPAGAIN;
        ret = newSVsv(POPs);
        PUTBACK;
        POPSTACK;
        SPAGAIN;
        FREETMPS;
        LEAVE;
        ST(0) = sv_2mortal(ret);
        XSRETURN(1);
    }
}

XS(XS_PerlAPI_utf8_shape) {
    dXSARGS;
    if (items != 1) croak_xs_usage(cv, "string");
    {
        STRLEN len, chars = 0;
        const U8 *p = (const U8 *)SvPV(ST(0), len);
        const U8 *end = 0;
        bool ok = is_utf8_string_loclen(p, len, &end, &chars);
        STRLEN ulen = sv_len_utf8(ST(0));
        SP -= items;
        EXTEND(SP, 4);
        PUSHs(boolSV(ok));
        mPUSHu(chars);
        mPUSHu((UV)(end - p));
        mPUSHu(ulen);
        PUTBACK;
        return;
    }
}

XS(XS_PerlAPI_container_ops) {
    dXSARGS;
    if (items != 2) croak_xs_usage(cv, "arrayref, hashref");
    {
        AV *av = (AV *)SvRV(ST(0));
        HV *hv = (HV *)SvRV(ST(1));
        SV *gone = av_delete(av, 1, 0);
        SP -= items;
        EXTEND(SP, 3);
        PUSHs(gone ? gone : &PL_sv_undef);
        mPUSHu(HvUSEDKEYS(hv));
        PUSHs(boolSV(hv_existss(hv, "k")));
        PUTBACK;
        return;
    }
}

XS(XS_PerlAPI_length_via_na) {
    dXSARGS;
    if (items != 1) croak_xs_usage(cv, "sv");
    (void)SvPV(ST(0), PL_na);
    XSRETURN_IV((IV)PL_na);
}

XS(XS_PerlAPI_version) {
    dXSARGS;
    PERL_UNUSED_VAR(items);
    ST(0) = sv_2mortal(newSVpvf("%d.%d.%d", PERL_REVISION, PATCHLEVEL, SUBVERSION));
    XSRETURN(1);
}

XS(XS_PerlAPI_phase) {
    dXSARGS;
    PERL_UNUSED_VAR(items);
    ST(0) = sv_2mortal(newSVpv(PL_phase_names[PL_phase], 0));
    XSRETURN(1);
}

XS(XS_PerlAPI_localize) {
    dXSARGS;
    if (items != 2) croak_xs_usage(cv, "globref, code");
    {
        GV *gv = (GV *)SvRV(ST(0));
        SV *code = ST(1);
        ENTER;
        save_item(GvSVn(gv));
        sv_setpvs(GvSVn(gv), "inside");
        PUSHMARK(SP);
        PUTBACK;
        call_sv(code, G_DISCARD);
        LEAVE;
        XSRETURN_EMPTY;
    }
}

XS(XS_PerlAPI_length_after_upgrade) {
    dXSARGS;
    if (items != 1) croak_xs_usage(cv, "sv");
    SvUPGRADE(ST(0), SVt_PV);
    XSRETURN_IV((IV)SvCUR(ST(0)));
}

/* sv_upgrade refuses to lower a type by dying; nothing after it may run. */
XS(XS_PerlAPI_lower_type) {
    dXSARGS;
    if (items != 1) croak_xs_usage(cv, "sv");
    sv_upgrade(ST(0), SVt_PV);
    croak("the XSUB went on after an API function died");
}

/* grok_number: the classification bits and, for an integer, the value. */
XS(XS_PerlAPI_number_kind) {
    dXSARGS;
    if (items != 1) croak_xs_usage(cv, "string");
    {
        STRLEN len;
        const char *p = SvPV(ST(0), len);
        UV value = 0;
        int kind = grok_number(p, len, &value);
        SV *out = sv_2mortal(newSVpvs(""));
        if (kind & IS_NUMBER_IN_UV) sv_catpvf(out, "uv=%lu,", (unsigned long)value);
        if (kind & IS_NUMBER_GREATER_THAN_UV_MAX) sv_catpvs(out, "big,");
        if (kind & IS_NUMBER_NOT_INT) sv_catpvs(out, "notint,");
        if (kind & IS_NUMBER_NEG) sv_catpvs(out, "neg,");
        if (kind & IS_NUMBER_INFINITY) sv_catpvs(out, "inf,");
        if (kind & IS_NUMBER_NAN) sv_catpvs(out, "nan,");
        ST(0) = out;
        XSRETURN(1);
    }
}

XS(XS_PerlAPI_throw) {
    dXSARGS;
    if (items != 1) croak_xs_usage(cv, "exception");
    croak_sv(ST(0));
}

/* die(NULL) / croak(NULL): the exception is the current value of $@. */
XS(XS_PerlAPI_rethrow_errsv) {
    dXSARGS;
    if (items != 1) croak_xs_usage(cv, "exception");
    sv_setsv(ERRSV, ST(0));
    croak(NULL);
}

XS_EXTERNAL(boot_PerlAPI) {
    dXSBOOTARGSXSAPIVERCHK;
    PERL_UNUSED_VAR(items);
    newXS_deffile("PerlAPI::flags", XS_PerlAPI_flags);
    newXSproto_portable("PerlAPI::is_plain_value", XS_PerlAPI_is_plain_value, __FILE__, "$");
    newXS_deffile("PerlAPI::string_only", XS_PerlAPI_string_only);
    newXS_deffile("PerlAPI::ok_off", XS_PerlAPI_ok_off);
    newXS_deffile("PerlAPI::read_twice_nomg", XS_PerlAPI_read_twice_nomg);
    newXS_deffile("PerlAPI::string_nomg", XS_PerlAPI_string_nomg);
    newXS_deffile("PerlAPI::new_object", XS_PerlAPI_new_object);
    newXS_deffile("PerlAPI::hex", XS_PerlAPI_hex);
    newXS_deffile("PerlAPI::errsv_via_errgv", XS_PerlAPI_errsv_via_errgv);
    newXS_deffile("PerlAPI::call_with_argv", XS_PerlAPI_call_with_argv);
    newXS_deffile("PerlAPI::require", XS_PerlAPI_require);
    newXS_deffile("PerlAPI::glob_name", XS_PerlAPI_glob_name);
    newXS_deffile("PerlAPI::copy_nosteal", XS_PerlAPI_copy_nosteal);
    newXS_deffile("PerlAPI::big", XS_PerlAPI_big);
    newXS_deffile("PerlAPI::sub_kind", XS_PerlAPI_sub_kind);
    newXS_deffile("PerlAPI::call_on_new_stack", XS_PerlAPI_call_on_new_stack);
    newXS_deffile("PerlAPI::utf8_shape", XS_PerlAPI_utf8_shape);
    newXS_deffile("PerlAPI::container_ops", XS_PerlAPI_container_ops);
    newXS_deffile("PerlAPI::length_via_na", XS_PerlAPI_length_via_na);
    newXS_deffile("PerlAPI::version", XS_PerlAPI_version);
    newXS_deffile("PerlAPI::phase", XS_PerlAPI_phase);
    newXS_deffile("PerlAPI::localize", XS_PerlAPI_localize);
    newXS_deffile("PerlAPI::length_after_upgrade", XS_PerlAPI_length_after_upgrade);
    newXS_deffile("PerlAPI::lower_type", XS_PerlAPI_lower_type);
    newXS_deffile("PerlAPI::throw", XS_PerlAPI_throw);
    newXS_deffile("PerlAPI::number_kind", XS_PerlAPI_number_kind);
    newXS_deffile("PerlAPI::rethrow_errsv", XS_PerlAPI_rethrow_errsv);
    Perl_xs_boot_epilog(aTHX_ ax);
}
