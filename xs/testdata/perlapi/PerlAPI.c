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

/* The store idiom of perlapi: the caller keeps its reference only when
 * av_store did not take the value. */
XS(XS_PerlAPI_store_first) {
    dXSARGS;
    if (items != 2) croak_xs_usage(cv, "arrayref, sv");
    {
        AV *av = (AV *)SvRV(ST(0));
        SV *sv = ST(1);
        SV **slot = av_store(av, 0, SvREFCNT_inc(sv));
        if (!slot) SvREFCNT_dec(sv);
        ST(0) = boolSV(slot && *slot == sv);
        XSRETURN(1);
    }
}

/* A callback called without G_EVAL: its exception passes through. */
XS(XS_PerlAPI_call_plain) {
    dXSARGS;
    if (items != 1) croak_xs_usage(cv, "code");
    PUSHMARK(SP);
    PUTBACK;
    call_sv(ST(0), G_DISCARD);
    XSRETURN_EMPTY;
}

/* The context of the call, returned and left in $PerlAPI::last_context. */
XS(XS_PerlAPI_context) {
    dXSARGS;
    PERL_UNUSED_VAR(items);
    {
        const U8 gimme = GIMME_V;
        const char *name = gimme == G_VOID ? "void" : gimme == G_SCALAR ? "scalar" : "list";
        sv_setpv(get_sv("PerlAPI::last_context", GV_ADD), name);
        ST(0) = sv_2mortal(newSVpv(name, 0));
        XSRETURN(1);
    }
}

/* The first character of a byte string: "codepoint/length", or
 * "malformed" when the bytes do not start with a well-formed character. */
XS(XS_PerlAPI_first_char) {
    dXSARGS;
    if (items != 1) croak_xs_usage(cv, "bytes");
    {
        STRLEN len, used = 0;
        const U8 *p = (const U8 *)SvPV(ST(0), len);
        UV cp = utf8n_to_uvchr(p, len, &used, UTF8_CHECK_ONLY);
        if (used == (STRLEN)-1)
            ST(0) = sv_2mortal(newSVpvs("malformed"));
        else
            ST(0) = sv_2mortal(newSVpvf("%lu/%lu", (unsigned long)cp, (unsigned long)used));
        XSRETURN(1);
    }
}

/* Ends the way xsubpp's PPCODE sections do: no XSRETURN. */
XS(XS_PerlAPI_bare_return) {
    dXSARGS;
    SP -= items;
    PUTBACK;
    return;
}

/* Calls an XSUB by its C function, then croaks: the croak has to unwind to
 * this function, not to the one that already returned. */
XS(XS_PerlAPI_direct_then_croak) {
    dXSARGS;
    PERL_UNUSED_VAR(items);
    PUSHMARK(SP);
    PUTBACK;
    XS_PerlAPI_bare_return(aTHX_ cv);
    SPAGAIN;
    croak("after the direct call");
}

/* ---- integers and pointers at the native width ---------------------------- */

/* A heap pointer kept in an SV by plain casts, as XS code written for a perl
 * whose IV is as wide as a pointer does, and read back each way. */
XS(XS_PerlAPI_pointer_roundtrip) {
    dXSARGS;
    PERL_UNUSED_VAR(items);
    {
        char *block = (char *)malloc(64);
        SV *plain = sv_2mortal(newSViv((IV)block));
        SV *set = sv_newmortal();
        SV *obj = sv_newmortal();
        const char *bad = 0;
        sv_setiv(set, (IV)block);
        sv_setref_pv(obj, "PerlAPI::Ptr", (void *)block);
        if ((char *)SvIV(plain) != block) bad = "newSViv/SvIV";
        else if ((char *)SvIV(set) != block) bad = "sv_setiv/SvIV";
        else if ((char *)SvUV(set) != block) bad = "sv_setiv/SvUV";
        else if (INT2PTR(char *, SvIV(SvRV(obj))) != block) bad = "sv_setref_pv/INT2PTR";
        else if ((char *)SvIVX(SvRV(obj)) != block) bad = "sv_setref_pv/SvIVX";
        else if (INT2PTR(char *, PTR2IV(block)) != block) bad = "PTR2IV/INT2PTR";
        else if (!sv_isa(obj, "PerlAPI::Ptr")) bad = "class";
        else if (!SvIOK(plain) || !SvIOK(SvRV(obj))) bad = "SvIOK";
        else if (!(SvFLAGS(set) & SVf_IOK)) bad = "SVf_IOK";
        free(block);
        ST(0) = sv_2mortal(newSVpv(bad ? bad : "same", 0));
        XSRETURN(1);
    }
}

/* SvIV and SvUV of the argument, printed by C at their full width. */
XS(XS_PerlAPI_wide_values) {
    dXSARGS;
    if (items != 1) croak_xs_usage(cv, "sv");
    {
        char buf[96];
        snprintf(buf, sizeof buf, "%lld/%llu", (long long)SvIV(ST(0)),
                 (unsigned long long)SvUV(ST(0)));
        ST(0) = sv_2mortal(newSVpv(buf, 0));
        XSRETURN(1);
    }
}

/* Integers beyond 32 bits made by each constructor. */
XS(XS_PerlAPI_wide_made) {
    dXSARGS;
    SP -= items;
    {
        SV *set = sv_newmortal();
        sv_setuv(set, (UV)1 << 41);
        XPUSHs(sv_2mortal(newSViv((IV)1 << 40)));
        XPUSHs(sv_2mortal(newSViv(-((IV)1 << 40))));
        XPUSHs(sv_2mortal(newSVuv((UV)3000000000u)));
        XPUSHs(set);
        XPUSHs(sv_2mortal(newSViv(-5)));
    }
    PUTBACK;
    return;
}

/* ---- an interpreter death inside a direct accessor ------------------------- */

/* Reads the argument as an integer (or, for how == 1, as a string) and then
 * records that it went on: when reading dies, it must not. */
XS(XS_PerlAPI_read_then_mark) {
    dXSARGS;
    if (items != 2) croak_xs_usage(cv, "sv, how");
    {
        SV *mark = get_sv("PerlAPI::went_on", GV_ADD);
        sv_setiv(mark, 0);
        if (SvIV(ST(1)) == 1) {
            STRLEN len;
            (void)SvPV(ST(0), len);
        } else {
            (void)SvIV(ST(0));
        }
        sv_setiv(mark, 1);
        XSRETURN_EMPTY;
    }
}

/* ---- body slots and interpreter variables as lvalues ----------------------- */

/* SvIVX(sv)++ on every element of an array, however many. */
XS(XS_PerlAPI_increment_all) {
    dXSARGS;
    if (items != 1) croak_xs_usage(cv, "aref");
    {
        AV *av = (AV *)SvRV(ST(0));
        SSize_t i, n = av_len(av) + 1;
        for (i = 0; i < n; i++) {
            SV **svp = av_fetch(av, i, 0);
            if (svp) SvIVX(*svp)++;
        }
        XSRETURN_IV((IV)n);
    }
}

/* A store through the address of a slot taken before other operations ran:
 * the order a compiler is free to give `SvIVX(sv) = f()`. */
XS(XS_PerlAPI_late_store) {
    dXSARGS;
    if (items != 1) croak_xs_usage(cv, "sv");
    {
        IV *slot = &SvIVX(ST(0));
        SV *other = sv_2mortal(newSViv(1)); /* a guest operation in between */
        *slot = 7 + SvIV(other);
        XSRETURN_IV(SvIV(ST(0)));
    }
}

/* Eight raw bytes in the NV slot, the way a module keeps a 64-bit integer
 * there on a perl whose IV has 32 bits. */
XS(XS_PerlAPI_nv_slot_bits) {
    dXSARGS;
    if (items != 1) croak_xs_usage(cv, "sv");
    {
        SV *sv = ST(0);
        char buf[40];
        SvUPGRADE(sv, SVt_NV);
        *(int64_t *)(&(SvNVX(sv))) = INT64_C(0x0123456789abcdef);
        SvNOK_on(sv);
        snprintf(buf, sizeof buf, "%llx",
                 (unsigned long long)*(uint64_t *)(&(SvNVX(sv))));
        ST(0) = sv_2mortal(newSVpv(buf, 0));
        XSRETURN(1);
    }
}

/* Assigns PL_tainted through its address after an operation, then reads it
 * back after another one; restores the old value. */
XS(XS_PerlAPI_late_variable_store) {
    dXSARGS;
    PERL_UNUSED_VAR(items);
    {
        IV *slot = (IV *)&PL_tainted;
        IV before = *slot, seen;
        SV *other = sv_2mortal(newSViv(1));
        *slot = 1;
        other = sv_2mortal(newSViv(2));
        PERL_UNUSED_VAR(other);
        seen = PL_tainted;
        PL_tainted = before;
        XSRETURN_IV(seen);
    }
}

/* PL_tainting assigned, as Taint::Runtime does. */
XS(XS_PerlAPI_set_tainting) {
    dXSARGS;
    if (items != 1) croak_xs_usage(cv, "on");
    PL_tainting = SvTRUE(ST(0)) ? 1 : 0;
    XSRETURN_IV(PL_tainting ? 1 : 0);
}

/* ---- caller_cx ---------------------------------------------------------------- */

XS(XS_PerlAPI_caller_frame) {
    dXSARGS;
    if (items != 1) croak_xs_usage(cv, "level");
    {
        const PERL_CONTEXT *cx = caller_cx((I32)SvIV(ST(0)), NULL);
        SV *out = sv_newmortal();
        if (!cx) {
            sv_setpvs(out, "none");
        } else if (CxTYPE(cx) == CXt_SUB) {
            GV *gv = CvGV(cx->blk_sub.cv);
            sv_setpvs(out, "sub ");
            if (gv && isGV(gv)) {
                SV *name = sv_newmortal();
                gv_efullname4(name, gv, NULL, 1);
                sv_catsv(out, name);
            }
            if (CxHASARGS(cx)) sv_catpvs(out, " with args");
        } else if (CxTYPE(cx) == CXt_EVAL) {
            sv_setpv(out, CxREALEVAL(cx) ? "eval string" : "eval block");
        } else {
            sv_setpvs(out, "other");
        }
        ST(0) = out;
        XSRETURN(1);
    }
}

/* ---- PerlIO handles -------------------------------------------------------------- */

/* Everything left on the input handle, read in small pieces. */
XS(XS_PerlAPI_pio_slurp) {
    dXSARGS;
    if (items != 1) croak_xs_usage(cv, "fh");
    {
        PerlIO *in = IoIFP(sv_2io(ST(0)));
        SV *out = sv_2mortal(newSVpvs(""));
        char buf[7];
        SSize_t n;
        while ((n = PerlIO_read(in, buf, sizeof buf)) > 0) sv_catpvn(out, buf, n);
        if (!PerlIO_eof(in)) sv_catpvs(out, "<not at eof>");
        if (PerlIO_error(in)) sv_catpvs(out, "<error>");
        ST(0) = out;
        XSRETURN(1);
    }
}

XS(XS_PerlAPI_pio_emit) {
    dXSARGS;
    if (items != 2) croak_xs_usage(cv, "fh, text");
    {
        PerlIO *out = IoOFP(sv_2io(ST(0)));
        STRLEN len;
        const char *text = SvPV(ST(1), len);
        SSize_t n = PerlIO_write(out, text, len);
        n += PerlIO_printf(out, "[%d:%s]", 42, "fmt");
        n += PerlIO_puts(out, "!");
        PerlIO_putc(out, '\n');
        XSRETURN_IV((IV)n);
    }
}

/* Seeks, reports the position, and reads the next n bytes. */
XS(XS_PerlAPI_pio_at) {
    dXSARGS;
    if (items != 4) croak_xs_usage(cv, "fh, offset, whence, n");
    {
        PerlIO *in = IoIFP(sv_2io(ST(0)));
        char buf[32], pos[32];
        IV n = SvIV(ST(3));
        SV *out = sv_newmortal();
        SSize_t got;
        if (n > (IV)sizeof buf) n = sizeof buf;
        if (PerlIO_seek(in, (Off_t)SvIV(ST(1)), (int)SvIV(ST(2))) != 0)
            XSRETURN_UNDEF;
        snprintf(pos, sizeof pos, "%ld:", (long)PerlIO_tell(in));
        sv_setpv(out, pos);
        got = PerlIO_read(in, buf, (Size_t)n);
        if (got > 0) sv_catpvn(out, buf, got);
        ST(0) = out;
        XSRETURN(1);
    }
}

/* The size fstat reports for the host descriptor behind the handle, or -1
 * when the handle has none. */
XS(XS_PerlAPI_pio_host_size) {
    dXSARGS;
    if (items != 1) croak_xs_usage(cv, "fh");
    {
        int fd = PerlIO_fileno(IoIFP(sv_2io(ST(0))));
        struct stat st;
        if (fd < 0) XSRETURN_IV(-1);
        if (fstat(fd, &st) != 0) XSRETURN_IV(-2);
        XSRETURN_IV((IV)st.st_size);
    }
}

/* One line through the stdio stream of the handle; the stream is the same
 * one every time it is asked for. */
XS(XS_PerlAPI_pio_stdio_line) {
    dXSARGS;
    if (items != 1) croak_xs_usage(cv, "fh");
    {
        PerlIO *in = IoIFP(sv_2io(ST(0)));
        FILE *fp = PerlIO_findFILE(in);
        char line[128];
        if (!fp || fp != PerlIO_findFILE(in)) XSRETURN_UNDEF;
        if (!fgets(line, sizeof line, fp)) XSRETURN_UNDEF;
        ST(0) = sv_2mortal(newSVpv(line, 0));
        XSRETURN(1);
    }
}

/* Writes through a stdio stream exported from the handle, then gives the
 * stream back. */
XS(XS_PerlAPI_pio_stdio_write) {
    dXSARGS;
    if (items != 2) croak_xs_usage(cv, "fh, text");
    {
        PerlIO *out = IoOFP(sv_2io(ST(0)));
        FILE *fp = PerlIO_exportFILE(out, "w");
        int n;
        if (!fp) XSRETURN_UNDEF;
        n = fprintf(fp, "<%s>", SvPV_nolen(ST(1)));
        fflush(fp);
        PerlIO_releaseFILE(out, fp);
        XSRETURN_IV(n);
    }
}

XS(XS_PerlAPI_pio_std) {
    dXSARGS;
    if (items != 1) croak_xs_usage(cv, "text");
    PerlIO_printf(PerlIO_stderr(), "err:%s\n", SvPV_nolen(ST(0)));
    PerlIO_printf(PerlIO_stdout(), "out:%s\n", SvPV_nolen(ST(0)));
    PerlIO_flush(PerlIO_stdout());
    XSRETURN_EMPTY;
}

/* The bytes SvPVbyte gives, in hex. */
XS(XS_PerlAPI_bytes_hex) {
    dXSARGS;
    if (items != 1) croak_xs_usage(cv, "sv");
    {
        STRLEN len, i;
        const unsigned char *p = (const unsigned char *)SvPVbyte(ST(0), len);
        SV *out = sv_2mortal(newSVpvs(""));
        for (i = 0; i < len; i++) {
            char hex[3];
            snprintf(hex, sizeof hex, "%02x", p[i]);
            sv_catpvn(out, hex, 2);
        }
        ST(0) = out;
        XSRETURN(1);
    }
}

/* my_setenv, then what the C library of the host process has for the name. */
XS(XS_PerlAPI_setenv_then_getenv) {
    dXSARGS;
    if (items != 2) croak_xs_usage(cv, "name, value");
    {
        const char *name = SvPV_nolen(ST(0));
        const char *seen;
        my_setenv(name, SvOK(ST(1)) ? SvPV_nolen(ST(1)) : NULL);
        seen = getenv(name);
        ST(0) = seen ? sv_2mortal(newSVpv(seen, 0)) : &PL_sv_undef;
        XSRETURN(1);
    }
}

/* Runs code with $@ localized the way XS code does it (SAVESPTR and a fresh
 * SV); returns what the code left in the inner $@. */
XS(XS_PerlAPI_with_local_errsv) {
    dXSARGS;
    if (items != 1) croak_xs_usage(cv, "code");
    {
        SV *seen;
        ENTER;
        SAVETMPS;
        SAVESPTR(ERRSV);
        ERRSV = sv_newmortal();
        PUSHMARK(SP);
        PUTBACK;
        call_sv(ST(0), G_DISCARD | G_EVAL);
        SPAGAIN;
        seen = newSVsv(ERRSV);
        FREETMPS;
        LEAVE;
        ST(0) = sv_2mortal(seen);
        XSRETURN(1);
    }
}

/* Looks into the read buffer of a handle: fills it, reports how much is
 * buffered and the first two bytes, and consumes those two. */
XS(XS_PerlAPI_pio_buffer) {
    dXSARGS;
    if (items != 1) croak_xs_usage(cv, "fh");
    {
        PerlIO *in = IoIFP(sv_2io(ST(0)));
        SV *out = sv_newmortal();
        char text[64];
        SSize_t cnt;
        STDCHAR *ptr;
        if (!PerlIO_fast_gets(in) || !PerlIO_has_cntptr(in)) XSRETURN_UNDEF;
        if (PerlIO_get_cnt(in) <= 0 && PerlIO_fill(in) != 0) XSRETURN_UNDEF;
        cnt = PerlIO_get_cnt(in);
        ptr = PerlIO_get_ptr(in);
        if (cnt < 2 || !ptr) XSRETURN_UNDEF;
        snprintf(text, sizeof text, "%ld:%c%c", (long)cnt, ptr[0], ptr[1]);
        PerlIO_set_ptrcnt(in, ptr + 2, cnt - 2);
        sv_setpv(out, text);
        ST(0) = out;
        XSRETURN(1);
    }
}

XS(XS_PerlAPI_presize) {
    dXSARGS;
    if (items != 2) croak_xs_usage(cv, "href, keys");
    {
        HV *hv = (HV *)SvRV(ST(0));
        hv_ksplit(hv, SvIV(ST(1)));
        if (ckWARN4_d(WARN_MISC, WARN_VOID, WARN_NUMERIC, WARN_UTF8)) {
            /* the four-category form compiles and answers */
        }
        XSRETURN_IV((IV)HvMAX(hv));
    }
}

/* ---- long lists on the stack --------------------------------------------- */

/* The sum of all arguments, however many. */
XS(XS_PerlAPI_sum_args) {
    dXSARGS;
    {
        IV sum = 0;
        I32 i;
        for (i = 0; i < items; i++) sum += SvIV(ST(i));
        XSRETURN_IV(sum);
    }
}

/* Returns 1..n: with one EXTEND for all of them, or with XPUSHs each. */
XS(XS_PerlAPI_count_up) {
    dXSARGS;
    if (items != 2) croak_xs_usage(cv, "n, one_by_one");
    {
        IV n = SvIV(ST(0)), i;
        int one_by_one = SvTRUE(ST(1));
        SP -= items;
        if (!one_by_one) EXTEND(SP, n);
        for (i = 1; i <= n; i++) {
            if (one_by_one)
                XPUSHs(sv_2mortal(newSViv(i)));
            else
                PUSHs(sv_2mortal(newSViv(i)));
        }
        PUTBACK;
        return;
    }
}

/* Calls code with n arguments in list context and returns how many values
 * came back and the last of them. */
XS(XS_PerlAPI_relay_list) {
    dXSARGS;
    if (items != 2) croak_xs_usage(cv, "code, n");
    {
        SV *code = ST(0);
        IV n = SvIV(ST(1)), i, last = 0;
        I32 count;
        ENTER;
        SAVETMPS;
        PUSHMARK(SP);
        EXTEND(SP, n);
        for (i = 1; i <= n; i++) PUSHs(sv_2mortal(newSViv(i)));
        PUTBACK;
        count = call_sv(code, G_LIST);
        SPAGAIN;
        if (count > 0) last = SvIV(TOPs);
        SP -= count;
        PUTBACK;
        FREETMPS;
        LEAVE;
        EXTEND(SP, 2);
        ST(0) = sv_2mortal(newSViv(count));
        ST(1) = sv_2mortal(newSViv(last));
        XSRETURN(2);
    }
}

/* The argument's characters escaped for display, and whether its first
 * character is a vowel. */
XS(XS_PerlAPI_display) {
    dXSARGS;
    if (items != 1) croak_xs_usage(cv, "sv");
    {
        SV *out = sv_newmortal();
        STRLEN len;
        const char *p = SvPV(ST(0), len);
        int vowel = len > 0 && memCHRs("aeiou", p[0]) != NULL;
        sv_uni_display(out, ST(0), 100, UNI_DISPLAY_QQ);
        sv_catpv(out, vowel ? " (vowel)" : " (other)");
        ST(0) = out;
        XSRETURN(1);
    }
}

/* What a module that keeps a looked-up method compares to know whether it
 * is still current (DBI's dispatch does this). */
XS(XS_PerlAPI_method_cache_gen) {
    dXSARGS;
    if (items != 1) croak_xs_usage(cv, "class");
    {
        HV *stash = gv_stashsv(ST(0), GV_ADD);
        XSRETURN_IV((IV)(PL_sub_generation + HvMROMETA(stash)->cache_gen));
    }
}

/* The sub a method lookup finds, taken with GvCV as method-dispatching XS
 * code does; looked up twice, so that the second answer comes from the
 * class's method cache. */
XS(XS_PerlAPI_method_twice) {
    dXSARGS;
    if (items != 2) croak_xs_usage(cv, "class, method");
    {
        HV *stash = gv_stashsv(ST(0), GV_ADD);
        const char *name = SvPV_nolen(ST(1));
        GV *first = gv_fetchmethod_autoload(stash, name, FALSE);
        GV *second = gv_fetchmethod_autoload(stash, name, FALSE);
        CV *a = first ? GvCV(first) : NULL;
        CV *b = second ? GvCV(second) : NULL;
        if (!a || !b) XSRETURN_PV(!a ? "first lookup: no sub" : "second lookup: no sub");
        XSRETURN_PV(a == b ? "same sub" : "different subs");
    }
}

/* Calls code with $_ set to a name, the way DBI runs a callback: the SV in
 * $_ is swapped for another and put back, never written to. Returns what
 * the code left in its $_. */
XS(XS_PerlAPI_with_topic) {
    dXSARGS;
    if (items != 2) croak_xs_usage(cv, "code, name");
    {
        SV *code = ST(0);
        SV *orig = DEFSV;
        SV *topic, *seen;
        SAVE_DEFSV;
        topic = sv_2mortal(newSVsv(ST(1)));
        DEFSV_set(topic);
        PUSHMARK(SP);
        PUTBACK;
        call_sv(code, G_DISCARD);
        SPAGAIN;
        seen = sv_2mortal(newSVsv(DEFSV));
        DEFSV_set(orig);
        ST(0) = seen;
        XSRETURN(1);
    }
}

/* Leaves errno as asked: "ENOENT", "EINVAL", "ERANGE" set it (then other
 * API calls follow, which must not disturb it), "zero" clears it, anything
 * else does not touch it. */
XS(XS_PerlAPI_leave_errno) {
    dXSARGS;
    if (items != 1) croak_xs_usage(cv, "how");
    {
        const char *how = SvPV_nolen(ST(0));
        if (strEQ(how, "ENOENT")) errno = ENOENT;
        else if (strEQ(how, "EINVAL")) errno = EINVAL;
        else if (strEQ(how, "ERANGE")) errno = ERANGE;
        else if (strEQ(how, "zero")) errno = 0;
        {
            SV *noise = sv_2mortal(newSVpvs("calls into the interpreter"));
            sv_catpvs(noise, " in between");
            ST(0) = sv_2mortal(newSViv((IV)SvCUR(noise)));
        }
        XSRETURN(1);
    }
}

static int perlapi_string_calls;
static const char *perlapi_next_string(int null) {
    perlapi_string_calls++;
    return null ? NULL : "text";
}

/* A `char *` result set the way xsubpp does it; a null pointer is undef.
 * Returns the value and how often the argument expression was evaluated. */
XS(XS_PerlAPI_set_from_pointer) {
    dXSARGS;
    if (items != 1) croak_xs_usage(cv, "null");
    {
        SV *out = sv_2mortal(newSVpvs("before"));
        SV *more = sv_2mortal(newSVpvs("kept"));
        int null = SvTRUE(ST(0));
        perlapi_string_calls = 0;
        sv_setpv(out, perlapi_next_string(null));
        sv_catpv(more, perlapi_next_string(null));
        SP -= items;
        EXTEND(SP, 3);
        PUSHs(out);
        PUSHs(more);
        PUSHs(sv_2mortal(newSViv(perlapi_string_calls)));
        PUTBACK;
        return;
    }
}

/* Passes all its arguments on to code (the first argument), then reads its
 * own second argument through MARK: the way a dispatching XSUB goes on
 * with its arguments after a callback, when the stack has moved meanwhile. */
XS(XS_PerlAPI_callback_then_mark) {
    dXSARGS;
    dORIGMARK;
    if (items < 2) croak_xs_usage(cv, "code, first, ...");
    {
        I32 i;
        EXTEND(SP, items + 1);
        PUSHMARK(SP);
        for (i = 1; i < items; i++) PUSHs(ST(i));
        PUTBACK;
        call_sv(ST(0), G_DISCARD);
        MSPAGAIN;
        XSRETURN_IV(SvIV(mark[2]) + (IV)(sp - mark));
    }
}

static int perlapi_name_calls;
static const char *perlapi_class_name(void) {
    perlapi_name_calls++;
    return "PerlAPI::Weak";
}

/* Keeps a weak reference to what the argument refers to, made the way DBI
 * records a child handle (the new reference is the argument of
 * sv_rvweaken), and reports the strong reference count of the referent
 * and how often an argument expression of sv_isa was evaluated. */
XS(XS_PerlAPI_weak_count) {
    dXSARGS;
    if (items != 2) croak_xs_usage(cv, "keep, ref");
    {
        AV *keep = (AV *)SvRV(ST(0));
        SV *target = SvRV(ST(1));
        U32 count;
        av_push(keep, (SV *)sv_rvweaken(newRV_inc(target)));
        count = SvREFCNT(target);
        perlapi_name_calls = 0;
        (void)sv_isa(ST(1), perlapi_class_name());
        SP -= items;
        EXTEND(SP, 2);
        mPUSHi((IV)count);
        mPUSHi((IV)perlapi_name_calls);
        PUTBACK;
        return;
    }
}

/* The file of the calling statement, read through its file glob the way
 * DBI's profiler names a caller. */
XS(XS_PerlAPI_caller_file) {
    dXSARGS;
    PERL_UNUSED_VAR(items);
    {
        GV *gv = CopFILEGV(PL_curcop);
        const char *via_gv = gv ? SvPV_nolen(GvSV(gv)) : "";
        const char *direct = CopFILE(PL_curcop);
        if (!*via_gv || !direct || strNE(via_gv, direct)) XSRETURN_UNDEF;
        ST(0) = sv_2mortal(newSVpv(via_gv, 0));
        XSRETURN(1);
    }
}

/* Native state kept on a variable: a counter in C memory reached through
 * ext magic, with a named copy of a label, released by svt_free. */
static int perlapi_tally_freed;
static int perlapi_tally_free(pTHX_ SV *sv, MAGIC *mg) {
    PERL_UNUSED_VAR(sv);
    perlapi_tally_freed++;
    free(mg->mg_ptr);
    mg->mg_ptr = NULL;
    return 0;
}
static MGVTBL perlapi_tally_vtbl = {0, 0, 0, 0, perlapi_tally_free, 0, 0, 0};
static MGVTBL perlapi_label_vtbl = {0, 0, 0, 0, 0, 0, 0, 0};

XS(XS_PerlAPI_tally_attach) {
    dXSARGS;
    if (items != 3) croak_xs_usage(cv, "sv, start, label");
    {
        STRLEN len;
        const char *label = SvPV(ST(2), len);
        IV *count = (IV *)malloc(sizeof(IV));
        *count = SvIV(ST(1));
        sv_magicext(ST(0), NULL, PERL_MAGIC_ext, &perlapi_tally_vtbl, (const char *)count, 0);
        sv_magicext(ST(0), NULL, PERL_MAGIC_ext, &perlapi_label_vtbl, label, (I32)len);
        XSRETURN(0);
    }
}

/* Adds to the counter of a variable and returns "label=count". */
XS(XS_PerlAPI_tally_add) {
    dXSARGS;
    if (items != 2) croak_xs_usage(cv, "sv, n");
    {
        MAGIC *tally = mg_findext(ST(0), PERL_MAGIC_ext, &perlapi_tally_vtbl);
        MAGIC *label = mg_findext(ST(0), PERL_MAGIC_ext, &perlapi_label_vtbl);
        if (!tally || !label) XSRETURN_UNDEF;
        *(IV *)tally->mg_ptr += SvIV(ST(1));
        ST(0) = sv_2mortal(newSVpvf("%s=%ld", label->mg_ptr, (long)*(IV *)tally->mg_ptr));
        XSRETURN(1);
    }
}

XS(XS_PerlAPI_tally_freed) {
    dXSARGS;
    PERL_UNUSED_VAR(items);
    ST(0) = sv_2mortal(newSViv(perlapi_tally_freed));
    XSRETURN(1);
}

/* A value kept in the XSUB's own slot (CvXSUBANY): set by a call with an
 * argument, returned by every call. */
XS(XS_PerlAPI_sub_slot) {
    dXSARGS;
    if (items > 0) CvXSUBANY(cv).any_i32 = (I32)SvIV(ST(0));
    ST(0) = sv_2mortal(newSViv(CvXSUBANY(cv).any_i32));
    XSRETURN(1);
}

/* Formats its argument into text by each of the formatting functions:
 * how = "new" (newSVpvf), "set" (sv_setpvf), "cat" (sv_catpvf onto the
 * third argument), "mixed" (bytes above 0x7f beside the argument), "long"
 * (a %s argument of the given length), "warn" and "croak". */
XS(XS_PerlAPI_formatted) {
    dXSARGS;
    if (items < 2) croak_xs_usage(cv, "how, sv, ...");
    {
        const char *how = SvPV_nolen(ST(0));
        SV *arg = ST(1);
        SV *out;
        if (strEQ(how, "new")) {
            out = sv_2mortal(newSVpvf("<%" SVf ">", SVfARG(arg)));
        } else if (strEQ(how, "set")) {
            out = sv_newmortal();
            sv_setpvf(out, "%d:%" SVf, 7, SVfARG(arg));
        } else if (strEQ(how, "cat")) {
            out = sv_mortalcopy(ST(2));
            sv_catpvf(out, "+%" SVf, SVfARG(arg));
        } else if (strEQ(how, "mixed")) {
            out = sv_2mortal(newSVpvf("\xe9%s|%" SVf "|\xe8", "\xea", SVfARG(arg)));
        } else if (strEQ(how, "long")) {
            STRLEN n = (STRLEN)SvIV(arg);
            char *text = (char *)malloc(n + 1);
            memset(text, 'x', n);
            text[n] = '\0';
            out = sv_2mortal(newSVpvf("[%s]", text));
            free(text);
        } else if (strEQ(how, "warn")) {
            warn("warned: %" SVf "\n", SVfARG(arg));
            XSRETURN(0);
        } else if (strEQ(how, "croak")) {
            croak("%" SVf, SVfARG(arg));
        } else {
            croak("formatted: unknown way %s", how);
        }
        ST(0) = out;
        XSRETURN(1);
    }
}

/* Magic whose mg_ptr points at the string buffer of another SV, kept
 * alive as mg_obj: the way DBI reaches the structure of a handle. */
static MGVTBL perlapi_view_vtbl = {0, 0, 0, 0, 0, 0, 0, 0};

XS(XS_PerlAPI_view_attach) {
    dXSARGS;
    if (items != 2) croak_xs_usage(cv, "holder, buffer");
    sv_magicext(ST(0), ST(1), PERL_MAGIC_ext, &perlapi_view_vtbl, SvPV_nolen(ST(1)), 0);
    XSRETURN(0);
}

/* The first n bytes at the place the magic points to. */
XS(XS_PerlAPI_view_read) {
    dXSARGS;
    if (items != 2) croak_xs_usage(cv, "holder, n");
    {
        MAGIC *mg = mg_findext(ST(0), PERL_MAGIC_ext, &perlapi_view_vtbl);
        if (!mg) XSRETURN_UNDEF;
        ST(0) = sv_2mortal(newSVpvn(mg->mg_ptr, (STRLEN)SvIV(ST(1))));
        XSRETURN(1);
    }
}

/* The process id as C code asks for it. */
XS(XS_PerlAPI_process_id) {
    dXSARGS;
    PERL_UNUSED_VAR(items);
    EXTEND(SP, 2);
    ST(0) = sv_2mortal(newSViv((IV)getpid()));
    ST(1) = sv_2mortal(newSViv((IV)PerlProc_getpid()));
    XSRETURN(2);
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
    newXS_deffile("PerlAPI::bare_return", XS_PerlAPI_bare_return);
    newXS_deffile("PerlAPI::direct_then_croak", XS_PerlAPI_direct_then_croak);
    newXS_deffile("PerlAPI::first_char", XS_PerlAPI_first_char);
    newXS_deffile("PerlAPI::store_first", XS_PerlAPI_store_first);
    newXS_deffile("PerlAPI::call_plain", XS_PerlAPI_call_plain);
    newXS_deffile("PerlAPI::context", XS_PerlAPI_context);
    newXS_deffile("PerlAPI::number_kind", XS_PerlAPI_number_kind);
    newXS_deffile("PerlAPI::rethrow_errsv", XS_PerlAPI_rethrow_errsv);
    newXS_deffile("PerlAPI::pointer_roundtrip", XS_PerlAPI_pointer_roundtrip);
    newXS_deffile("PerlAPI::wide_values", XS_PerlAPI_wide_values);
    newXS_deffile("PerlAPI::wide_made", XS_PerlAPI_wide_made);
    newXS_deffile("PerlAPI::read_then_mark", XS_PerlAPI_read_then_mark);
    newXS_deffile("PerlAPI::increment_all", XS_PerlAPI_increment_all);
    newXS_deffile("PerlAPI::late_store", XS_PerlAPI_late_store);
    newXS_deffile("PerlAPI::nv_slot_bits", XS_PerlAPI_nv_slot_bits);
    newXS_deffile("PerlAPI::late_variable_store", XS_PerlAPI_late_variable_store);
    newXS_deffile("PerlAPI::set_tainting", XS_PerlAPI_set_tainting);
    newXS_deffile("PerlAPI::caller_frame", XS_PerlAPI_caller_frame);
    newXS_deffile("PerlAPI::pio_slurp", XS_PerlAPI_pio_slurp);
    newXS_deffile("PerlAPI::pio_emit", XS_PerlAPI_pio_emit);
    newXS_deffile("PerlAPI::pio_at", XS_PerlAPI_pio_at);
    newXS_deffile("PerlAPI::pio_host_size", XS_PerlAPI_pio_host_size);
    newXS_deffile("PerlAPI::pio_stdio_line", XS_PerlAPI_pio_stdio_line);
    newXS_deffile("PerlAPI::pio_stdio_write", XS_PerlAPI_pio_stdio_write);
    newXS_deffile("PerlAPI::pio_std", XS_PerlAPI_pio_std);
    newXS_deffile("PerlAPI::bytes_hex", XS_PerlAPI_bytes_hex);
    newXS_deffile("PerlAPI::setenv_then_getenv", XS_PerlAPI_setenv_then_getenv);
    newXS_deffile("PerlAPI::with_local_errsv", XS_PerlAPI_with_local_errsv);
    newXS_deffile("PerlAPI::pio_buffer", XS_PerlAPI_pio_buffer);
    newXS_deffile("PerlAPI::presize", XS_PerlAPI_presize);
    newXS_deffile("PerlAPI::sum_args", XS_PerlAPI_sum_args);
    newXS_deffile("PerlAPI::count_up", XS_PerlAPI_count_up);
    newXS_deffile("PerlAPI::relay_list", XS_PerlAPI_relay_list);
    newXS_deffile("PerlAPI::display", XS_PerlAPI_display);
    newXS_deffile("PerlAPI::method_cache_gen", XS_PerlAPI_method_cache_gen);
    newXS_deffile("PerlAPI::method_twice", XS_PerlAPI_method_twice);
    newXS_deffile("PerlAPI::with_topic", XS_PerlAPI_with_topic);
    newXS_deffile("PerlAPI::leave_errno", XS_PerlAPI_leave_errno);
    newXS_deffile("PerlAPI::set_from_pointer", XS_PerlAPI_set_from_pointer);
    newXS_deffile("PerlAPI::callback_then_mark", XS_PerlAPI_callback_then_mark);
    newXS_deffile("PerlAPI::weak_count", XS_PerlAPI_weak_count);
    newXS_deffile("PerlAPI::caller_file", XS_PerlAPI_caller_file);
    newXS_deffile("PerlAPI::tally_attach", XS_PerlAPI_tally_attach);
    newXS_deffile("PerlAPI::tally_add", XS_PerlAPI_tally_add);
    newXS_deffile("PerlAPI::tally_freed", XS_PerlAPI_tally_freed);
    newXS_deffile("PerlAPI::sub_slot", XS_PerlAPI_sub_slot);
    newXS_deffile("PerlAPI::formatted", XS_PerlAPI_formatted);
    newXS_deffile("PerlAPI::view_attach", XS_PerlAPI_view_attach);
    newXS_deffile("PerlAPI::view_read", XS_PerlAPI_view_read);
    newXS_deffile("PerlAPI::process_id", XS_PerlAPI_process_id);
    Perl_xs_boot_epilog(aTHX_ ax);
}
