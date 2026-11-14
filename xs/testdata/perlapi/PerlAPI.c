/* Exercises the perl API surface of the SDK (goperl_perlapi.h) the way XS
 * distributions use it. Written directly in the shape xsubpp generates. */
#include <time.h>
#include "EXTERN.h"
#include "perl.h"
#include "XSUB.h"
#include "patchlevel.h"
#include <perlio.h>

/* The sizes config.h gives are numbers, as perl's are: a module compares
 * them in the preprocessor (Compress::LZF: #if Size_t_size > 4). */
#if !(Size_t_size >= 4 && SIZE_T_SIZE == Size_t_size && SHORTSIZE == 2 && \
      INTSIZE == 4 && LONGSIZE >= 4 && LONGLONGSIZE == 8 && PTRSIZE >= 4 && \
      DOUBLESIZE == 8 && IVSIZE == 8 && UVSIZE == 8 && NVSIZE == 8 && \
      U8SIZE == 1 && U16SIZE == 2 && U32SIZE == 4 && I32SIZE == 4 && \
      U64SIZE == 8 && I64SIZE == 8 && CHARBITS == 8 && BYTEORDER > 0 && \
      MEM_ALIGNBYTES >= 4)
#error "a size of config.h is not what it should be"
#endif

#include <sys/socket.h>
#include <sys/mman.h>
#include <unistd.h>

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

/* live_values(): how many SVs the interpreter has. */
XS(XS_PerlAPI_live_values) {
    dXSARGS;
    PERL_UNUSED_VAR(items);
    ST(0) = sv_2mortal(newSViv(PL_sv_count));
    XSRETURN(1);
}

/* die_in_operation(value, how): an operation on the value that dies when
 * the value's stringification does, after the XSUB has made values of its
 * own the way an XSUB does before it works: none (how 0), mortal ones (1),
 * and a mortal copy of the value that is then used as a hash key, which is
 * what List::MoreUtils' uniq does (2). */
XS(XS_PerlAPI_die_in_operation) {
    dXSARGS;
    if (items != 2) croak_xs_usage(cv, "value, how");
    {
        const IV how = SvIV(ST(1));
        if (how == 0) {
            (void)SvPV_nolen(ST(0));
        } else {
            HV *hv = newHV();
            SV *tmp = sv_newmortal();
            sv_2mortal(newRV_noinc((SV *)hv));
            if (how == 1) {
                (void)SvPV_nolen(ST(0));
            } else {
                SvSetSV_nosteal(tmp, ST(0));
                (void)hv_exists_ent(hv, tmp, 0);
            }
        }
        croak("the XSUB went on after an operation died");
    }
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

/* ---- names of the API that stand for other names --------------------------------- */

static int plain_destructor_runs;
static void plain_destructor(void *arg) { plain_destructor_runs += *(int *)arg; }

/* other_names(hv, globname): a destructor without an interpreter argument,
 * the hash calls by key and length, the slots of a glob made on demand,
 * what an array and a pad value are, and a buffer handed to an SV. The
 * answers are joined with commas. */
XS(XS_PerlAPI_other_names) {
    dXSARGS;
    if (items != 2) croak_xs_usage(cv, "hv, globname");
    {
        HV *const hv = (HV *)SvRV(ST(0));
        GV *const gv = gv_fetchpv(SvPV_nolen(ST(1)), GV_ADD, SVt_PVGV);
        SV *const out = sv_2mortal(newSVpvs(""));
        AV *const av = (AV *)sv_2mortal((SV *)newAV());
        SV *const constant = sv_2mortal(newSViv(1));
        SV *const taken = sv_newmortal();
        SV **svp;
        char *buf;
        int by = 3;

        plain_destructor_runs = 0;
        ENTER;
        SAVEDESTRUCTOR(plain_destructor, &by);
        LEAVE;
        sv_catpvf(out, "destructor %d", plain_destructor_runs);

        (void)hv_common_key_len(hv, "stored", 6, HV_FETCH_ISSTORE | HV_FETCH_JUST_SV, newSVpvs("value"), 0);
        svp = (SV **)hv_common_key_len(hv, "stored", 6, HV_FETCH_JUST_SV, NULL, 0);
        sv_catpvf(out, ",fetched %s", svp ? SvPV_nolen(*svp) : "nothing");
        {
            /* the other shapes of hv_common: the entry by a scalar key,
             * the place of a value stored, a value made on demand */
            SV *key = sv_2mortal(newSVpvs("stored"));
            HE *he = (HE *)hv_common(hv, key, NULL, 0, 0, 0, NULL, 0);
            SV **made = (SV **)hv_common_key_len(hv, "made", 4, HV_FETCH_JUST_SV | HV_FETCH_LVALUE, NULL, 0);
            SV **put = (SV **)hv_common(hv, NULL, "put", 3, 0, HV_FETCH_ISSTORE | HV_FETCH_JUST_SV,
                                        newSVpvs("there"), 0);
            sv_catpvf(out, ",entry %s,made %d,put %s,absent %d", he ? SvPV_nolen(HeVAL(he)) : "nothing",
                      made && *made && !SvOK(*made) && hv_exists(hv, "made", 4) ? 1 : 0,
                      put ? SvPV_nolen(*put) : "nothing",
                      hv_common_key_len(hv, "absent", 6, HV_FETCH_JUST_SV, NULL, 0) ? 1 : 0);
        }
        sv_catpvf(out, ",exists %d%d",
                  hv_common_key_len(hv, "stored", 6, HV_FETCH_ISEXISTS, NULL, 0) ? 1 : 0,
                  hv_common_key_len(hv, "absent", 6, HV_FETCH_ISEXISTS, NULL, 0) ? 1 : 0);
        (void)hv_common_key_len(hv, "stored", 6, HV_DELETE | G_DISCARD, NULL, 0);
        sv_catpvf(out, ",deleted %d", hv_exists(hv, "stored", 6) ? 0 : 1);

        sv_catpvf(out, ",glob %d%d", gv_HVadd(gv) == gv && GvHV(gv) ? 1 : 0,
                  gv_AVadd(gv) == gv && GvAV(gv) ? 1 : 0);

        SvREADONLY_on(constant);
        sv_catpvf(out, ",array %d%d,pad %d%d", AvREAL(av) ? 1 : 0, AvREIFY(av) ? 1 : 0,
                  IS_PADCONST(constant) ? 1 : 0, IS_PADGV((SV *)gv) ? 1 : 0);

        Newx(buf, 16, char);
        strcpy(buf, "handed over");
        sv_usepvn_flags(taken, buf, 11, SV_SMAGIC);
        sv_catpvf(out, ",string %" SVf, SVfARG(taken));

        ST(0) = out;
        XSRETURN(1);
    }
}

/* ---- versions, globs without a name, random numbers ------------------------------ */

/* scan_a_version(text): what scan_version makes of the text: the version as
 * it was written, and what follows it in the text. */
XS(XS_PerlAPI_scan_a_version) {
    dXSARGS;
    if (items != 1) croak_xs_usage(cv, "text");
    {
        const char *const text = SvPV_nolen(ST(0));
        char own[64];
        SV *const from_sv = sv_newmortal();
        SV *const from_own = sv_newmortal();
        const char *end = scan_version(text, from_sv, FALSE);
        SV **original = SvROK(from_sv) ? hv_fetchs((HV *)SvRV(from_sv), "original", 0) : NULL;
        SV *const out = sv_2mortal(newSVpvf("%s|%s", original ? SvPV_nolen(*original) : "no version", end));
        /* and from a buffer of the module's own */
        strncpy(own, text, sizeof own - 1);
        own[sizeof own - 1] = 0;
        end = scan_version(own, from_own, FALSE);
        sv_catpvf(out, "|%d", (int)(end - own));
        ST(0) = out;
        XSRETURN(1);
    }
}

/* generated_globs(package): two globs made by newGVgen: whether each is a
 * glob of the package under a generated name, and whether they differ. */
XS(XS_PerlAPI_generated_globs) {
    dXSARGS;
    if (items != 1) croak_xs_usage(cv, "package");
    {
        GV *const a = newGVgen(SvPV_nolen(ST(0)));
        GV *const b = newGVgen(SvPV_nolen(ST(0)));
        ST(0) = sv_2mortal(newSVpvf("%d%d%d,%s", isGV(a) ? 1 : 0, strnEQ(GvNAME(a), "_GEN_", 5) ? 1 : 0,
                                    a != b && strNE(GvNAME(a), GvNAME(b)) ? 1 : 0, HvNAME(GvSTASH(a))));
        XSRETURN(1);
    }
}

/* random_numbers(n): the next n numbers of the generator rand() uses, after
 * seeding it the way pp_rand does when nobody has; then whether it counts
 * as seeded, and a comparison of two strings in the locale. */
XS(XS_PerlAPI_random_numbers) {
    dXSARGS;
    if (items != 1) croak_xs_usage(cv, "n");
    {
        const IV n = SvIV(ST(0));
        IV i;
        SP -= items;
        if (!PL_srand_called) {
            (void)seedDrand01((Rand_seed_t)seed());
            PL_srand_called = TRUE;
        }
        EXTEND(SP, n + 3);
        for (i = 0; i < n; i++) mPUSHn(Drand01());
        mPUSHi(PL_srand_called ? 1 : 0);
        mPUSHi(sv_cmp_locale(sv_2mortal(newSVpvs("a")), sv_2mortal(newSVpvs("b"))));
        mPUSHi(PL_sv_count > 0 ? 1 : 0);
        PUTBACK;
        return;
    }
}

/* ---- a pp function run for an op of the module's own ---------------------------- */

/* pp_on_own_op(target, value, how): does `target += value` (how 0),
 * `target .= value` (1) or `target ||= value` (2) by running the pp function
 * of the operator for an op it builds on its own stack, the way a module
 * gets the interpreter's own semantics for an operator (overloading
 * included) without having an op tree. */
XS(XS_PerlAPI_pp_on_own_op) {
    dXSARGS;
    if (items != 3) croak_xs_usage(cv, "target, value, how");
    {
        static const int types[] = {OP_ADD, OP_CONCAT, OP_OR};
        SV *const target = ST(0);
        SV *const value = ST(1);
        const int how = (int)SvIV(ST(2));
        OP *oldop;
        BINOP myop;
        if (how < 0 || how > 2) croak("pp_on_own_op: unknown operator %d", how);
        Zero(&myop, 1, BINOP);
        myop.op_flags = OPf_STACKED;
        myop.op_type = types[how];
        ENTER;
        SAVETMPS;
        PUSHMARK(SP);
        if (how == 2) {
            XPUSHs(value);
            XPUSHs(target);
        } else {
            XPUSHs(target);
            XPUSHs(value);
        }
        PUTBACK;
        oldop = PL_op;
        PL_op = (OP *)&myop;
        PL_ppaddr[PL_op->op_type](aTHX);
        PL_op = oldop;
        if (how == 2) {
            SPAGAIN;
            sv_setsv(target, TOPs);
            SvSETMAGIC(target);
        }
        POPMARK;
        FREETMPS;
        LEAVE;
        XSRETURN_EMPTY;
    }
}

/* ---- the op that called an XSUB ---------------------------------------------------- */

/* calling_op(): the name of PL_op as the XSUB finds it, and whether its pp
 * function is the interpreter's for its type. */
XS(XS_PerlAPI_calling_op) {
    dXSARGS;
    PERL_UNUSED_VAR(items);
    ST(0) = sv_2mortal(newSVpvf("%s,%s", PL_op ? OP_NAME(PL_op) : "none",
                                PL_op && PL_op->op_ppaddr == PL_ppaddr[PL_op->op_type] ? "its own pp" : "another pp"));
    XSRETURN(1);
}

/* ---- an array filled by its slots, an lvalue scalar built by hand ---------------- */

/* filled_by_slots(n): an array of n numbers made the way code makes one it
 * knows the size of: room for n, every slot stored through AvARRAY, then
 * AvFILLp. With n < 0, an array of -n numbers is made that way and then
 * emptied by releasing the elements and setting AvFILLp to -1. */
XS(XS_PerlAPI_filled_by_slots) {
    dXSARGS;
    if (items != 1) croak_xs_usage(cv, "n");
    {
        const IV want = SvIV(ST(0));
        const SSize_t n = (SSize_t)(want < 0 ? -want : want);
        AV *const av = newAV();
        SSize_t i;
        av_extend(av, n - 1);
        for (i = 0; i < n; i++) AvARRAY(av)[i] = newSViv((IV)i * 10);
        AvFILLp(av) = n - 1;
        if (want < 0) {
            SV **const body = AvARRAY(av);
            for (i = 0; i < n; i++) {
                SvREFCNT_dec(body[i]);
                body[i] = NULL;
            }
            AvFILLp(av) = -1;
        }
        ST(0) = sv_2mortal(newRV_noinc((SV *)av));
        XSRETURN(1);
    }
}

/* lvalue_fields(): an lvalue scalar (PVLV) given its type and its target
 * by assignment, as code that builds one for its own magic does; what the
 * two fields hold afterwards. */
XS(XS_PerlAPI_lvalue_fields) {
    dXSARGS;
    PERL_UNUSED_VAR(items);
    {
        SV *const sv = sv_newmortal();
        SV *const target = sv_2mortal(newSVpvs("target"));
        sv_upgrade(sv, SVt_PVLV);
        LvTYPE(sv) = '~';
        LvTARG(sv) = SvREFCNT_inc(target);
        ST(0) = sv_2mortal(newSVpvf("%c,%s", LvTYPE(sv), LvTARG(sv) == target ? "the target" : "something else"));
        XSRETURN(1);
    }
}

/* ---- what the accessors cost -------------------------------------------------- */

/* accessor_loop(sv, n, how): evaluates one accessor of sv n times and
 * returns a sum of what it read (BenchmarkNativeAccessors). */
XS(XS_PerlAPI_accessor_loop) {
    dXSARGS;
    if (items != 3) croak_xs_usage(cv, "sv, n, how");
    {
        SV *const sv = ST(0);
        const IV n = SvIV(ST(1));
        const int how = (int)SvIV(ST(2));
        UV sum = 0;
        IV i;
        switch (how) {
        case 0:
            for (i = 0; i < n; i++) sum += (unsigned char)SvPVX_const(sv)[0];
            break;
        case 1:
            for (i = 0; i < n; i++) sum += (unsigned char)SvPVX(sv)[0];
            break;
        case 2:
            for (i = 0; i < n; i++) sum += (UV)SvLEN(sv);
            break;
        case 3:
            for (i = 0; i < n; i++) sum += (UV)SvCUR(sv);
            break;
        case 4:
            for (i = 0; i < n; i++) sum += (UV)SvIVX(sv);
            break;
        case 5:
            for (i = 0; i < n; i++) sum += (UV)SvIV(sv);
            break;
        default:
            croak("accessor_loop: unknown accessor %d", how);
        }
        ST(0) = sv_2mortal(newSVuv(sum));
        XSRETURN(1);
    }
}

/* array_loop(av, rounds, how): looks at every element of the array, rounds
 * times over, and returns how many slots were filled: with av_fetch (how
 * 0), with AvARRAY(av)[i] written out for every element (how 1), and with
 * the body taken once per round, `SV **body = AvARRAY(av)` (how 2). */
XS(XS_PerlAPI_array_loop) {
    dXSARGS;
    if (items != 3) croak_xs_usage(cv, "av, rounds, how");
    {
        AV *const av = (AV *)SvRV(ST(0));
        const IV rounds = SvIV(ST(1));
        const int how = (int)SvIV(ST(2));
        const SSize_t len = av_len(av) + 1;
        UV filled = 0;
        IV r;
        SSize_t i;
        for (r = 0; r < rounds; r++) {
            if (how == 0) {
                for (i = 0; i < len; i++) {
                    SV **const svp = av_fetch(av, i, 0);
                    if (svp && *svp) filled++;
                }
            } else if (how == 1) {
                for (i = 0; i < len; i++)
                    if (AvARRAY(av)[i]) filled++;
            } else {
                SV **const body = AvARRAY(av);
                for (i = 0; i < len; i++)
                    if (body[i]) filled++;
            }
        }
        ST(0) = sv_2mortal(newSVuv(filled));
        XSRETURN(1);
    }
}

/* ---- a structure of the module's own behind an SV --------------------------- */

struct kept_pair {
    IV first;
    IV second;
};

/* keep_pair(n, how): an SV whose buffer slot holds the address of a
 * structure the module allocated, stored with SvPVX as an lvalue (how 0) or
 * with SvPV_set (how 1): the SV is no string and owns no buffer. With how 2
 * the SV is also given the size of the structure as the room of its
 * buffer, which hands the memory over to it. */
XS(XS_PerlAPI_keep_pair) {
    dXSARGS;
    if (items != 2) croak_xs_usage(cv, "n, how");
    {
        struct kept_pair *pair;
        SV *const sv = newSV(0);
        Newx(pair, 1, struct kept_pair);
        pair->first = SvIV(ST(0));
        pair->second = 2 * pair->first;
        sv_upgrade(sv, SVt_PV);
        if (SvIV(ST(1)) == 1) {
            SvPV_set(sv, (char *)pair);
        } else {
            SvPVX(sv) = (char *)pair;
            if (SvIV(ST(1)) == 2) SvLEN(sv) = sizeof(struct kept_pair);
        }
        ST(0) = sv_2mortal(newRV_noinc(sv));
        XSRETURN(1);
    }
}

/* kept_pair(ref): what the structure behind the SV holds, whether the SV
 * counts as a string, and the room it says its buffer has. */
XS(XS_PerlAPI_kept_pair) {
    dXSARGS;
    if (items != 1) croak_xs_usage(cv, "ref");
    {
        SV *const sv = SvRV(ST(0));
        const struct kept_pair *const pair = (const struct kept_pair *)SvPVX(sv);
        const struct kept_pair *const again = (const struct kept_pair *)SvPVX_const(sv);
        if (!pair || pair != again)
            ST(0) = sv_2mortal(newSVpvs("lost"));
        else
            ST(0) = sv_2mortal(newSVpvf("%" IVdf ",%" IVdf ",%s,%d", pair->first, pair->second,
                                        SvPOK(sv) ? "string" : "no string", (int)SvLEN(sv)));
        XSRETURN(1);
    }
}

/* release_pair(ref): frees the structure and empties the slot, as a module
 * does before the SV goes; says what the slot holds afterwards. */
XS(XS_PerlAPI_release_pair) {
    dXSARGS;
    if (items != 1) croak_xs_usage(cv, "ref");
    {
        SV *const sv = SvRV(ST(0));
        Safefree(SvPVX(sv));
        SvPVX(sv) = NULL;
        ST(0) = sv_2mortal(newSVpv(SvPVX(sv) ? "still set" : "empty", 0));
        XSRETURN(1);
    }
}

/* ---- structures of the module's own kept as values ------------------------ */

/* A structure of the module's own, in memory mapped for it, so that its
 * address is as wide as the addresses of this process are. */
typedef struct {
    U32 magic;
    IV value;
} kept_struct;
#define KEPT_MAGIC 0x4b505453u
static kept_struct *kept_structs(pTHX_ IV n) {
    static kept_struct *block;
    static IV used;
    const IV room = 4096;
    kept_struct *s;
    if (!block) {
        block = (kept_struct *)mmap(NULL, sizeof(kept_struct) * (size_t)room,
                                    PROT_READ | PROT_WRITE, MAP_PRIVATE | MAP_ANON, -1, 0);
        if (block == (kept_struct *)MAP_FAILED) croak("mmap failed");
    }
    if (used + n > room) croak("out of kept structures");
    s = block + used;
    used += n;
    return s;
}
static IV kept_value(pTHX_ SV *sv) {
    const kept_struct *const s = (const kept_struct *)sv;
    if (!s || s->magic != KEPT_MAGIC) croak("not a kept structure");
    return s->value;
}

/* kept_array(n): an array of n structures of the module's own, the first
 * half pushed, the rest stored at their index (structures, not values:
 * the array is kept as long as the program runs, as Object::Pad keeps its
 * arrays of method descriptions). */
XS(XS_PerlAPI_kept_array) {
    dXSARGS;
    if (items != 1) croak_xs_usage(cv, "n");
    {
        const IV n = SvIV(ST(0));
        kept_struct *const s = kept_structs(aTHX_ n);
        AV *const av = newAV();
        IV i;
        for (i = 0; i < n; i++) {
            s[i].magic = KEPT_MAGIC;
            s[i].value = i;
            if (i < n / 2)
                av_push(av, (SV *)&s[i]);
            else
                av_store(av, i, (SV *)&s[i]);
        }
        ST(0) = sv_2mortal(newRV_noinc((SV *)av));
        XSRETURN(1);
    }
}

/* kept_array_sums(ref): the sum of the values of the structures, read
 * through AvARRAY, through av_fetch and through a copy that AvARRAY
 * stores (a slot it is given back to). */
XS(XS_PerlAPI_kept_array_sums) {
    dXSARGS;
    if (items != 1) croak_xs_usage(cv, "ref");
    {
        AV *const av = (AV *)SvRV(ST(0));
        const SSize_t n = av_count(av);
        IV a = 0, b = 0, c = 0;
        SSize_t i;
        AV *const copy = newAV();
        SV **arr;
        for (i = 0; i < n; i++) a += kept_value(aTHX_ AvARRAY(av)[i]);
        for (i = 0; i < n; i++) b += kept_value(aTHX_ *av_fetch(av, i, 0));
        av_extend(copy, n);
        arr = AvARRAY(av);
        for (i = 0; i < n; i++) AvARRAY(copy)[i] = arr[i];
        AvFILLp(copy) = n - 1;
        for (i = 0; i < n; i++) c += kept_value(aTHX_ AvARRAY(copy)[i]);
        /* the structures are no values for the copy to release */
        for (i = 0; i < n; i++) AvARRAY(copy)[i] = &PL_sv_undef;
        AvFILLp(copy) = -1;
        SvREFCNT_dec((SV *)copy);
        ST(0) = sv_2mortal(newSVpvf("%" IVdf ",%" IVdf ",%" IVdf, a, b, c));
        XSRETURN(1);
    }
}

/* kept_hash(n): a hash of n structures of the module's own under the keys
 * k0 .. k<n-1>, half stored by hv_store, the rest by hv_store_ent. */
XS(XS_PerlAPI_kept_hash) {
    dXSARGS;
    if (items != 1) croak_xs_usage(cv, "n");
    {
        const IV n = SvIV(ST(0));
        kept_struct *const s = kept_structs(aTHX_ n);
        HV *const hv = newHV();
        IV i;
        for (i = 0; i < n; i++) {
            char key[32];
            const int klen = snprintf(key, sizeof key, "k%" IVdf, i);
            s[i].magic = KEPT_MAGIC;
            s[i].value = i;
            if (i < n / 2) {
                if (!hv_store(hv, key, klen, (SV *)&s[i], 0)) croak("hv_store failed");
            } else {
                SV *const keysv = sv_2mortal(newSVpvn(key, klen));
                if (!hv_store_ent(hv, keysv, (SV *)&s[i], 0)) croak("hv_store_ent failed");
            }
        }
        ST(0) = sv_2mortal(newRV_noinc((SV *)hv));
        XSRETURN(1);
    }
}

/* kept_hash_sums(ref, n): the sum of the values of the structures, read
 * through hv_fetch, hv_fetch_ent and an iteration of the hash. */
XS(XS_PerlAPI_kept_hash_sums) {
    dXSARGS;
    if (items != 2) croak_xs_usage(cv, "ref, n");
    {
        HV *const hv = (HV *)SvRV(ST(0));
        const IV n = SvIV(ST(1));
        IV a = 0, b = 0, c = 0, i;
        HE *he;
        for (i = 0; i < n; i++) {
            char key[32];
            const int klen = snprintf(key, sizeof key, "k%" IVdf, i);
            SV **svp = hv_fetch(hv, key, klen, 0);
            SV *const keysv = sv_2mortal(newSVpvn(key, klen));
            HE *const ent = hv_fetch_ent(hv, keysv, 0, 0);
            if (!svp || !ent) croak("no entry for %s", key);
            a += kept_value(aTHX_ *svp);
            b += kept_value(aTHX_ HeVAL(ent));
        }
        hv_iterinit(hv);
        while ((he = hv_iternext(hv))) c += kept_value(aTHX_ HeVAL(he));
        ST(0) = sv_2mortal(newSVpvf("%" IVdf ",%" IVdf ",%" IVdf, a, b, c));
        XSRETURN(1);
    }
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

/* The limits of the integers, made by each constructor and through the
 * lvalue of the IV slot, and the size the SDK says an IV has. */
XS(XS_PerlAPI_integer_limits) {
    dXSARGS;
    SP -= items;
    {
        SV *set_iv = sv_newmortal();
        SV *set_uv = sv_newmortal();
        SV *slot = sv_2mortal(newSViv(0));
        sv_setiv(set_iv, IV_MIN);
        sv_setuv(set_uv, UV_MAX);
        SvIVX(slot) = IV_MAX;
        EXTEND(SP, 7);
        PUSHs(sv_2mortal(newSViv(IV_MAX)));
        PUSHs(sv_2mortal(newSViv(IV_MIN)));
        PUSHs(sv_2mortal(newSVuv(UV_MAX)));
        PUSHs(set_iv);
        PUSHs(set_uv);
        PUSHs(slot);
        PUSHs(sv_2mortal(newSViv((IV)IVSIZE)));
    }
    PUTBACK;
    return;
}

/* A number a module keeps in the IV slot of a scalar that is read-only for
 * the program, the way Math::Int64 keeps its 64-bit values on a perl whose
 * IV has 64 bits: sealed_new makes the object, sealed_add updates the slot
 * in place through its address and returns what the slot holds. */
XS(XS_PerlAPI_sealed_new) {
    dXSARGS;
    if (items != 1) croak_xs_usage(cv, "value");
    {
        IV value = SvIV(ST(0));
        SV *inner = newSV(0);
        SV *ref;
        SvUPGRADE(inner, SVt_IV);
        SvIOK_on(inner);
        ref = newRV_noinc(inner);
        sv_bless(ref, gv_stashpvs("PerlAPI::Sealed", TRUE));
        *(int64_t *)(&(SvIVX(inner))) = (int64_t)value;
        SvREADONLY_on(inner);
        ST(0) = sv_2mortal(ref);
        XSRETURN(1);
    }
}

XS(XS_PerlAPI_sealed_add) {
    dXSARGS;
    if (items != 2) croak_xs_usage(cv, "object, n");
    {
        SV *inner = SvRV(ST(0));
        IV n = SvIV(ST(1));
        *(int64_t *)(&(SvIVX(inner))) += (int64_t)n;
        SvIVX(inner)++;
        SvIVX(inner)--;
        ST(0) = sv_2mortal(newSViv(SvIVX(inner)));
        XSRETURN(1);
    }
}

/* A new scalar with the integer the argument holds, read from its slot:
 * undef for an argument that is not an integer. */
XS(XS_PerlAPI_copy_integer) {
    dXSARGS;
    if (items != 1) croak_xs_usage(cv, "sv");
    {
        SV *sv = ST(0);
        if (!SvIOK(sv))
            ST(0) = &PL_sv_undef;
        else if (SvIsUV(sv))
            ST(0) = sv_2mortal(newSVuv(SvUVX(sv)));
        else
            ST(0) = sv_2mortal(newSViv(SvIVX(sv)));
        XSRETURN(1);
    }
}

/* ---- native code that blocks ------------------------------------------------ */

/* Reads one byte from a descriptor of the host process with the C library,
 * for as long as that takes, and returns what read returned. */
XS(XS_PerlAPI_read_host_descriptor) {
    dXSARGS;
    if (items != 1) croak_xs_usage(cv, "fd");
    {
        char byte;
        ssize_t n = read((int)SvIV(ST(0)), &byte, 1);
        ST(0) = sv_2mortal(newSViv((IV)n));
        XSRETURN(1);
    }
}

/* ---- a message that names the caller's line ---------------------------------- */

/* croak_at_caller(msg): dies with msg as code of the sub that called this
 * one would: PL_curcop is pointed at the statement that called that sub,
 * and the interpreter appends its file and line. */
XS(XS_PerlAPI_croak_at_caller) {
    dXSARGS;
    if (items != 1) croak_xs_usage(cv, "msg");
    {
        const PERL_CONTEXT *const cx = caller_cx(0, NULL);
        if (cx) PL_curcop = cx->blk_oldcop;
        croak_sv(ST(0));
    }
}

/* ---- an exception caught and thrown again ------------------------------------ */

static int native_went_on_count; /* defined with read_then_mark, below */

/* try_then_rethrow(code): calls code inside XCPT_TRY, and throws what it
 * caught again after counting that it did (native_went_on). */
XS(XS_PerlAPI_try_then_rethrow) {
    dXSARGS;
    dXCPT;
    if (items != 1) croak_xs_usage(cv, "code");
    XCPT_TRY_START {
        PUSHMARK(SP);
        PUTBACK;
        call_sv(ST(0), G_DISCARD | G_NOARGS);
    } XCPT_TRY_END
    XCPT_CATCH {
        native_went_on_count++;
        XCPT_RETHROW;
    }
    XSRETURN_EMPTY;
}

/* ---- an interpreter death inside a direct accessor ------------------------- */

/* How often native code went on after a call that can die (read_then_mark
 * after reading, the counted_block keyword after parsing), counted in this
 * module's own memory: what the native code did, whatever the interpreter
 * made of the operations it asked for afterwards. */
static int native_went_on_count;

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
        native_went_on_count++;
        sv_setiv(mark, 1);
        XSRETURN_EMPTY;
    }
}

/* native_went_on(): how often native code went on since the last call. */
XS(XS_PerlAPI_native_went_on) {
    dXSARGS;
    PERL_UNUSED_VAR(items);
    {
        const int n = native_went_on_count;
        native_went_on_count = 0;
        ST(0) = sv_2mortal(newSViv(n));
        XSRETURN(1);
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

/* ---- descriptors, as the host's system calls see them ------------------------ */

/* The descriptor of a handle, as PerlIO_fileno gives it to native code. */
XS(XS_PerlAPI_fd_of) {
    dXSARGS;
    if (items != 1) croak_xs_usage(cv, "fh");
    XSRETURN_IV((IV)PerlIO_fileno(IoIFP(sv_2io(ST(0)))));
}

/* What the host says about a descriptor: "closed", or its kind and mode,
 * as in "pipe blocking" or "socket nonblocking". */
XS(XS_PerlAPI_fd_state) {
    dXSARGS;
    if (items != 1) croak_xs_usage(cv, "fd");
    {
        const int fd = (int)SvIV(ST(0));
        struct stat st;
        int flags;
        const char *kind;
        char out[64];
        if (fstat(fd, &st) != 0 || (flags = fcntl(fd, F_GETFL)) < 0) {
            ST(0) = sv_2mortal(newSVpvs("closed"));
            XSRETURN(1);
        }
        kind = S_ISREG(st.st_mode)    ? "file"
               : S_ISDIR(st.st_mode)  ? "directory"
               : S_ISFIFO(st.st_mode) ? "pipe"
               : S_ISSOCK(st.st_mode) ? "socket"
               : S_ISCHR(st.st_mode)  ? "device"
                                      : "other";
        snprintf(out, sizeof out, "%s %s", kind,
                 (flags & O_NONBLOCK) ? "nonblocking" : "blocking");
        ST(0) = sv_2mortal(newSVpv(out, 0));
        XSRETURN(1);
    }
}

/* write(2) on a descriptor; the count, or undef with errno left for $!. */
XS(XS_PerlAPI_fd_write) {
    dXSARGS;
    if (items != 2) croak_xs_usage(cv, "fd, text");
    {
        STRLEN len;
        const char *text = SvPV(ST(1), len);
        const ssize_t n = write((int)SvIV(ST(0)), text, len);
        if (n < 0) XSRETURN_UNDEF;
        XSRETURN_IV((IV)n);
    }
}

/* read(2) on a descriptor, once, of at most max bytes; what was read, or
 * undef with errno left for $!. */
XS(XS_PerlAPI_fd_read) {
    dXSARGS;
    if (items != 2) croak_xs_usage(cv, "fd, max");
    {
        char buf[256];
        size_t max = (size_t)SvUV(ST(1));
        ssize_t n;
        if (max > sizeof buf) max = sizeof buf;
        n = read((int)SvIV(ST(0)), buf, max);
        if (n < 0) XSRETURN_UNDEF;
        ST(0) = sv_2mortal(newSVpvn(buf, (STRLEN)n));
        XSRETURN(1);
    }
}

/* Descriptors made by native code: the two ends of a pipe, or of a pair of
 * connected sockets. */
XS(XS_PerlAPI_fd_pair) {
    dXSARGS;
    if (items != 1) croak_xs_usage(cv, "kind");
    {
        int fds[2];
        const char *kind = SvPV_nolen(ST(0));
        const int rc = strEQ(kind, "pipe") ? pipe(fds)
                                           : socketpair(AF_UNIX, SOCK_STREAM, 0, fds);
        if (rc != 0) XSRETURN_EMPTY;
        SP -= items;
        EXTEND(SP, 2);
        PUSHs(sv_2mortal(newSViv(fds[0])));
        PUSHs(sv_2mortal(newSViv(fds[1])));
        PUTBACK;
        return;
    }
}

/* close(2) on a descriptor. */
XS(XS_PerlAPI_fd_close) {
    dXSARGS;
    if (items != 1) croak_xs_usage(cv, "fd");
    XSRETURN_IV((IV)close((int)SvIV(ST(0))));
}

/* A line read through a handle native code opens on a descriptor
 * (PerlIO_fdopen); the handle is closed again, and the descriptor with it. */
XS(XS_PerlAPI_fd_line) {
    dXSARGS;
    if (items != 1) croak_xs_usage(cv, "fd");
    {
        PerlIO *in = PerlIO_fdopen((int)SvIV(ST(0)), "r");
        SV *line;
        int c;
        if (!in) XSRETURN_UNDEF;
        line = sv_2mortal(newSVpvs(""));
        while ((c = PerlIO_getc(in)) != EOF && c != '\n') sv_catpvf(line, "%c", c);
        PerlIO_close(in);
        ST(0) = line;
        XSRETURN(1);
    }
}

/* Writes through a handle imported from a stdio stream that native code
 * opened on a descriptor of its own (PerlIO_importFILE), and closes it. */
XS(XS_PerlAPI_fd_import_write) {
    dXSARGS;
    if (items != 2) croak_xs_usage(cv, "fd, text");
    {
        FILE *fp = fdopen((int)SvIV(ST(0)), "w");
        PerlIO *out;
        STRLEN len;
        const char *text = SvPV(ST(1), len);
        SSize_t n;
        if (!fp) XSRETURN_UNDEF;
        out = PerlIO_importFILE(fp, 0);
        if (!out) XSRETURN_UNDEF;
        n = PerlIO_write(out, text, len);
        PerlIO_close(out);
        XSRETURN_IV((IV)n);
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

/* Whether an object derives from a class, asked by the name of its stash,
 * as Object::Pad asks it of the class of a method (sv_derived_from_hv). */
XS(XS_PerlAPI_derived_from_stash) {
    dXSARGS;
    if (items != 2) croak_xs_usage(cv, "object, class");
    {
        HV *const stash = gv_stashsv(ST(1), 0);
        if (!stash) XSRETURN_PV("no stash");
        XSRETURN_IV(sv_derived_from_pvn(ST(0), HvNAME(stash), HvNAMELEN(stash),
                                        HvNAMEUTF8(stash) ? SVf_UTF8 : 0) ? 1 : 0);
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
#ifdef EHOSTDOWN
        else if (strEQ(how, "EHOSTDOWN")) errno = EHOSTDOWN;
#endif
        else if (strEQ(how, "zero")) errno = 0;
        {
            SV *noise = sv_2mortal(newSVpvs("calls into the interpreter"));
            sv_catpvs(noise, " in between");
            ST(0) = sv_2mortal(newSViv((IV)SvCUR(noise)));
        }
        XSRETURN(1);
    }
}

/* What native code finds in $! when the program assigned an error number to
 * it: the name the number has in this system's <errno.h>, and the number.
 * Then it assigns ERANGE to $!, as a number of that header. */
XS(XS_PerlAPI_errno_of_the_program) {
    dXSARGS;
    if (items != 0) croak_xs_usage(cv, "");
    {
        SV *bang = get_sv("!", 0);
        IV seen = SvIV(bang);
        SP -= items;
        EXTEND(SP, 2);
        mPUSHs(newSVpv(seen == EACCES ? "EACCES" : seen == ENOENT ? "ENOENT" : "another", 0));
        mPUSHi(seen);
        sv_setiv_mg(bang, ERANGE);
        PUTBACK;
        return;
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

/* Magic whose mg_obj the module sets itself after sv_magicext and gives up
 * in its own svt_free, the way Data::Buffer::Shared pins the handle a view
 * aliases: sv_magicext took no reference, so the magic is not
 * MGf_REFCOUNTED and mg_free must not give one up. */
static int perlapi_pin_free(pTHX_ SV *sv, MAGIC *mg) {
    PERL_UNUSED_ARG(sv);
    if (mg->mg_obj) SvREFCNT_dec(mg->mg_obj);
    return 0;
}
static MGVTBL perlapi_pin_vtbl = {0, 0, 0, 0, perlapi_pin_free, 0, 0, 0};

XS(XS_PerlAPI_pin_attach) {
    dXSARGS;
    if (items != 2) croak_xs_usage(cv, "holder, obj");
    {
        MAGIC *mg = sv_magicext(ST(0), NULL, PERL_MAGIC_ext, &perlapi_pin_vtbl, NULL, 0);
        mg->mg_obj = SvREFCNT_inc_simple_NN(ST(1));
        XSRETURN(0);
    }
}

/* Whether the magic of a vtable on a holder holds a reference on its
 * mg_obj that mg_free gives up: "pin" or "view". */
XS(XS_PerlAPI_magic_refcounted) {
    dXSARGS;
    if (items != 2) croak_xs_usage(cv, "holder, which");
    {
        MAGIC *mg = mg_findext(ST(0), PERL_MAGIC_ext,
                               strEQ(SvPV_nolen(ST(1)), "pin") ? &perlapi_pin_vtbl : &perlapi_view_vtbl);
        if (!mg) XSRETURN_UNDEF;
        ST(0) = boolSV(mg->mg_flags & MGf_REFCOUNTED);
        XSRETURN(1);
    }
}

/* SvRV reads the slot, whatever the flags say: a reference whose ROK flag
 * is turned off still has its target there, as in perl. */
XS(XS_PerlAPI_rv_slot_after_rok_off) {
    dXSARGS;
    if (items != 1) croak_xs_usage(cv, "target");
    {
        SV *rv = sv_2mortal(newRV_inc(ST(0)));
        int same;
        SvROK_off(rv);
        same = SvRV(rv) == ST(0);
        SvROK_on(rv); /* the mortal gives its reference back */
        ST(0) = boolSV(same);
        XSRETURN(1);
    }
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

/* Makes the argument a number without a string, the way DBI's
 * sql_type_cast discards the string of a value it converted: the buffer,
 * which is the interpreter's, is freed by the module. */
XS(XS_PerlAPI_discard_string) {
    dXSARGS;
    if (items != 1) croak_xs_usage(cv, "sv");
    {
        SV *sv = ST(0);
        sv_2iv(sv);
        if (SvNIOK(sv) && SvPVX(sv)) {
            SvOOK_off(sv);
            sv_force_normal(sv);
            if (SvLEN(sv)) Safefree(SvPVX(sv));
            SvPOK_off(sv);
            SvPV_set(sv, NULL);
            SvLEN_set(sv, 0);
            SvCUR_set(sv, 0);
        }
        XSRETURN(0);
    }
}

/* Gives the string of the argument a buffer of n bytes by resizing the one
 * it has, and fills what was added with dots. */
XS(XS_PerlAPI_resize_buffer) {
    dXSARGS;
    if (items != 2) croak_xs_usage(cv, "sv, n");
    {
        SV *sv = ST(0);
        STRLEN n = (STRLEN)SvIV(ST(1));
        STRLEN cur;
        char *buf = SvPV_force(sv, cur);
        Renew(buf, n + 1, char);
        memset(buf + cur, '.', n - cur);
        buf[n] = '\0';
        SvPV_set(sv, buf);
        SvLEN_set(sv, n + 1);
        SvCUR_set(sv, n);
        XSRETURN(0);
    }
}

/* What a module does with an argument it must not take from tainted data. */
XS(XS_PerlAPI_insecure) {
    dXSARGS;
    if (items != 2) croak_xs_usage(cv, "sv, what");
    TAINT_IF(SvTAINTED(ST(0)));
    TAINT_PROPER(SvPV_nolen(ST(1)));
    XSRETURN_YES;
}

/* Calls the code with $? saved and zero, as an object system does around
 * its destructors (local $?). */
XS(XS_PerlAPI_with_local_status) {
    dXSARGS;
    if (items != 1) croak_xs_usage(cv, "code");
    SAVEI32(PL_statusvalue);
    PL_statusvalue = 0;
    PUSHMARK(SP);
    PUTBACK;
    call_sv(ST(0), G_DISCARD);
    XSRETURN_EMPTY;
}

/* Saves one int of a pair in a scope, changes both, leaves the scope:
 * returns the pair, of which only the saved one is restored. */
XS(XS_PerlAPI_save_one_int) {
    dXSARGS;
    static struct {
        int saved;
        int beside;
    } pair;
    PERL_UNUSED_VAR(items);
    pair.saved = 1;
    pair.beside = 2;
    ENTER;
    SAVEI32(pair.saved);
    pair.saved = 5;
    pair.beside = 7;
    LEAVE;
    EXTEND(SP, 2);
    ST(0) = sv_2mortal(newSViv(pair.saved));
    ST(1) = sv_2mortal(newSViv(pair.beside));
    XSRETURN(2);
}

/* Opens a file the way open() does, through do_openn, and returns its
 * first bytes read from the new handle. */
XS(XS_PerlAPI_open_and_read) {
    dXSARGS;
    if (items != 3) croak_xs_usage(cv, "mode, name, n");
    {
        STRLEN mode_len;
        const char *mode = SvPV(ST(0), mode_len);
        SV *name = ST(1);
        char buf[64];
        SSize_t got;
        GV *gv = MUTABLE_GV(sv_2mortal(newSV_type(SVt_NULL)));
        gv_init_pvn(gv, CopSTASH(PL_curcop), "__ANONIO__", 10, GV_ADDMULTI);
        if (!do_openn(gv, mode, mode_len, 0, 0, 0, NULL, &name, 1)) XSRETURN_UNDEF;
        got = PerlIO_read(IoIFP(GvIO(gv)), buf, (Size_t)SvIV(ST(2)) < sizeof buf ? (Size_t)SvIV(ST(2)) : sizeof buf);
        if (got < 0) XSRETURN_UNDEF;
        ST(0) = sv_2mortal(newSVpvn(buf, (STRLEN)got));
        XSRETURN(1);
    }
}

/* The size of a file and its first byte, by the low-level I/O functions. */
XS(XS_PerlAPI_lio_peek) {
    dXSARGS;
    if (items != 1) croak_xs_usage(cv, "path");
    {
        Stat_t st;
        char first = 0;
        int fd = PerlLIO_open3(SvPV_nolen(ST(0)), O_RDONLY | O_BINARY, 0);
        if (fd < 0) XSRETURN_UNDEF;
        if (PerlLIO_fstat(fd, &st) != 0 || PerlLIO_read(fd, &first, 1) != 1) {
            PerlLIO_close(fd);
            XSRETURN_UNDEF;
        }
        PerlLIO_close(fd);
        ST(0) = sv_2mortal(newSVpvf("%ld:%c", (long)st.st_size, first));
        XSRETURN(1);
    }
}

/* Counts references to a structure of the module's own with SvREFCNT_inc
 * and SvREFCNT_dec, as Marpa::R2 does: on perl the count lands in the
 * structure, in the second word. Returns that word's change after the
 * increment and after the decrement, and the reference counts the two
 * values the structure points at had before. */
XS(XS_PerlAPI_count_own_struct) {
    dXSARGS;
    if (items != 2) croak_xs_usage(cv, "a, b");
    {
        struct perlapi_owner {
            SV *first;
            UV second;
            SV *third;
        } *owner = (struct perlapi_owner *)malloc(sizeof(struct perlapi_owner));
        IV after_inc, after_dec;
        const IV count_a = (IV)SvREFCNT(ST(0)), count_b = (IV)SvREFCNT(ST(1));
        owner->first = ST(0);
        owner->second = 1000;
        owner->third = ST(1);
        SvREFCNT_inc(owner);
        after_inc = (IV)owner->second - 1000;
        SvREFCNT_dec(owner);
        after_dec = (IV)owner->second - 1000;
        free(owner);
        SP -= items;
        EXTEND(SP, 4);
        mPUSHi(after_inc);
        mPUSHi(after_dec);
        mPUSHi(count_a);
        mPUSHi(count_b);
        PUTBACK;
        return;
    }
}

/* The bytes, in hex, of the string of a value in one of the forms the
 * SvPVutf8 and SvPVbyte families have, or "NULL" where the form has no
 * string. */
XS(XS_PerlAPI_string_form) {
    dXSARGS;
    if (items != 2) croak_xs_usage(cv, "form, sv");
    {
        const char *const form = SvPV_nolen(ST(0));
        SV *const sv = ST(1);
        STRLEN len = 0, i;
        const char *p = NULL;
        SV *out;
        if (strEQ(form, "utf8")) p = SvPVutf8(sv, len);
        else if (strEQ(form, "utf8_nomg")) p = SvPVutf8_nomg(sv, len);
        else if (strEQ(form, "utf8x")) p = SvPVutf8x(sv, len);
        else if (strEQ(form, "utf8_or_null")) p = SvPVutf8_or_null(sv, len);
        else if (strEQ(form, "utf8_or_null_nomg")) p = SvPVutf8_or_null_nomg(sv, len);
        else if (strEQ(form, "byte_or_null")) p = SvPVbyte_or_null(sv, len);
        else if (strEQ(form, "byte_or_null_nomg")) p = SvPVbyte_or_null_nomg(sv, len);
        else if (strEQ(form, "2pvutf8_flags")) p = sv_2pvutf8_flags(sv, &len, 0);
        else if (strEQ(form, "2pvbyte_flags")) p = sv_2pvbyte_flags(sv, &len, SV_GMAGIC);
        else if (strEQ(form, "utf8_nolen")) { p = SvPVutf8_nolen(sv); len = p ? strlen(p) : 0; }
        else if (strEQ(form, "utf8_force")) p = SvPVutf8_force(sv, len);
        else if (strEQ(form, "utf8x_force")) p = SvPVutf8x_force(sv, len);
        else if (strEQ(form, "byte_force")) p = SvPVbyte_force(sv, len);
        else if (strEQ(form, "bytex_force")) p = SvPVbytex_force(sv, len);
        else if (strEQ(form, "force_nomg")) p = SvPV_force_nomg(sv, len);
        else if (strEQ(form, "force_flags")) p = SvPV_force_flags(sv, len, SV_GMAGIC);
        else if (strEQ(form, "pvn_force_nomg")) p = sv_pvn_force_nomg(sv, &len);
        else croak("unknown form %s", form);
        if (!p) {
            ST(0) = sv_2mortal(newSVpvs("NULL"));
            XSRETURN(1);
        }
        out = sv_2mortal(newSVpvs(""));
        for (i = 0; i < len; i++) sv_catpvf(out, "%02x", (unsigned)(unsigned char)p[i]);
        ST(0) = out;
        XSRETURN(1);
    }
}

/* A number formatted by a call whose result is dropped with
 * PERL_UNUSED_RESULT: the call is made. */
XS(XS_PerlAPI_format_dropping_result) {
    dXSARGS;
    if (items != 1) croak_xs_usage(cv, "nv");
    {
        char buf[64];
        buf[0] = '\0';
        PERL_UNUSED_RESULT(Gconvert(SvNV(ST(0)), NV_DIG, 0, buf));
        ST(0) = sv_2mortal(newSVpv(buf, 0));
        XSRETURN(1);
    }
}

/* Which of the first eight slots of an array's body (AvARRAY) are filled,
 * for an array that held eight elements and was then emptied, extended to
 * room for eight and given one element at index 2: the slots the array has
 * room for and no element in are NULL. */
XS(XS_PerlAPI_array_slots) {
    dXSARGS;
    if (items != 1 || !SvROK(ST(0)) || SvTYPE(SvRV(ST(0))) != SVt_PVAV)
        croak_xs_usage(cv, "arrayref");
    {
        AV *const av = (AV *)SvRV(ST(0));
        SV *const out = sv_2mortal(newSVpvs(""));
        IV i, filled = 0;
        av_clear(av);
        for (i = 0; i < 8; i++) av_push(av, newSViv(i));
        for (i = 0; i < 8; i++)
            if (AvARRAY(av)[i]) filled++;
        sv_catpvf(out, "%" IVdf " filled;", filled);
        av_clear(av);
        av_extend(av, 7);
        av_store(av, 2, newSViv(42));
        for (i = 0; i < 8; i++)
            sv_catpv(out, AvARRAY(av)[i] ? " set" : " null");
        ST(0) = out;
        XSRETURN(1);
    }
}

/* The body of the subs anonymous_subs makes: which one it is. */
XS(XS_PerlAPI_anonymous_body) {
    dXSARGS;
    dXSI32;
    PERL_UNUSED_VAR(items);
    ST(0) = sv_2mortal(newSViv(ix));
    XSRETURN(1);
}

/* Makes n subs without a name (newXS(NULL, ...)), the way a module builds
 * closures over native code, and returns references to them. The caller of
 * newXS owns the one reference to such a sub: it is handed to the mortals
 * here, and the reference returned keeps the sub. */
XS(XS_PerlAPI_anonymous_subs) {
    dXSARGS;
    if (items != 1) croak_xs_usage(cv, "n");
    {
        const IV n = SvIV(ST(0));
        IV i;
        SP -= items;
        EXTEND(SP, n);
        for (i = 0; i < n; i++) {
            CV *const sub = newXS(NULL, XS_PerlAPI_anonymous_body, __FILE__);
            CvXSUBANY(sub).any_i32 = (I32)i;
            sv_2mortal((SV *)sub);
            PUSHs(sv_2mortal(newRV_inc((SV *)sub)));
        }
        PUTBACK;
        return;
    }
}

/* The SVs the string constructors make without a string, and one with an
 * empty string: which of them are defined. */
XS(XS_PerlAPI_null_strings) {
    dXSARGS;
    PERL_UNUSED_VAR(items);
    {
        SV *made[4];
        int i;
        made[0] = sv_2mortal(newSVpvn(NULL, 0));
        made[1] = sv_2mortal(newSVpv(NULL, 0));
        made[2] = newSVpvn_flags(NULL, 0, SVs_TEMP);
        made[3] = sv_2mortal(newSVpvn("", 0));
        SP -= items;
        EXTEND(SP, 4);
        for (i = 0; i < 4; i++) PUSHs(boolSV(SvOK(made[i])));
        PUTBACK;
        return;
    }
}

/* Set magic that records every value its variable is given. */
static char perlapi_watch_log[256];
static int perlapi_watch_set(pTHX_ SV *sv, MAGIC *mg) {
    size_t used = strlen(perlapi_watch_log);
    PERL_UNUSED_VAR(mg);
    snprintf(perlapi_watch_log + used, sizeof perlapi_watch_log - used, "|%s",
             SvOK(sv) ? SvPV_nolen(sv) : "undef");
    return 0;
}
static MGVTBL perlapi_watch_vtbl = {0, perlapi_watch_set, 0, 0, 0, 0, 0, 0};

XS(XS_PerlAPI_watch) {
    dXSARGS;
    if (items != 1) croak_xs_usage(cv, "sv");
    perlapi_watch_log[0] = '\0';
    sv_magicext(ST(0), NULL, PERL_MAGIC_ext, &perlapi_watch_vtbl, "watched", 0);
    XSRETURN(0);
}

XS(XS_PerlAPI_watched) {
    dXSARGS;
    PERL_UNUSED_VAR(items);
    ST(0) = sv_2mortal(newSVpv(perlapi_watch_log, 0));
    XSRETURN(1);
}

/* Magic whose name is an SV (HEf_SVKEY), as an accessor generator keeps
 * the name of its slot: attach returns the reference count of the key
 * before and after, read returns the key through the magic. */
static MGVTBL perlapi_keyed_vtbl = {0, 0, 0, 0, 0, 0, 0, 0};

XS(XS_PerlAPI_keyed_attach) {
    dXSARGS;
    if (items != 2) croak_xs_usage(cv, "holder, key");
    {
        const IV before = (IV)SvREFCNT(ST(1));
        MAGIC *mg = sv_magicext(ST(0), NULL, PERL_MAGIC_ext, &perlapi_keyed_vtbl,
                                (const char *)ST(1), HEf_SVKEY);
        const IV after = (IV)SvREFCNT(ST(1));
        const IV len = (IV)mg->mg_len;
        SP -= items;
        EXTEND(SP, 2);
        mPUSHi(after - before);
        mPUSHi(len);
        PUTBACK;
        return;
    }
}

XS(XS_PerlAPI_keyed_read) {
    dXSARGS;
    if (items != 1) croak_xs_usage(cv, "holder");
    {
        MAGIC *mg = mg_findext(ST(0), PERL_MAGIC_ext, &perlapi_keyed_vtbl);
        if (!mg || mg->mg_len != HEf_SVKEY) XSRETURN_UNDEF;
        ST(0) = (SV *)mg->mg_ptr;
        XSRETURN(1);
    }
}

/* What a hash or array is tied to: the class of the object, or "not tied". */
XS(XS_PerlAPI_tied_to) {
    dXSARGS;
    if (items != 1 || !SvROK(ST(0))) croak_xs_usage(cv, "ref");
    {
        SV *target = SvRV(ST(0));
        MAGIC *mg = SvTIED_mg(target, PERL_MAGIC_tied);
        if (!mg) {
            ST(0) = sv_2mortal(newSVpvs("not tied"));
        } else {
            ST(0) = sv_2mortal(newSVpv(sv_reftype(SvRV(SvTIED_obj(target, mg)), 1), 0));
        }
        XSRETURN(1);
    }
}

/* Decodes the first character of a byte string: "code point:length", or
 * "malformed". how: 0 strict, 1 overlong allowed, 2 overlong with its
 * value. */
XS(XS_PerlAPI_decode_first) {
    dXSARGS;
    if (items != 2) croak_xs_usage(cv, "bytes, how");
    {
        STRLEN len, used = 0;
        const U8 *bytes = (const U8 *)SvPV(ST(0), len);
        const IV how = SvIV(ST(1));
        const U32 flags = UTF8_CHECK_ONLY | (how == 1 ? UTF8_ALLOW_LONG : how == 2 ? UTF8_ALLOW_LONG_AND_ITS_VALUE : 0);
        const UV uv = utf8n_to_uvchr(bytes, len, &used, flags);
        if (used == (STRLEN)-1) {
            ST(0) = sv_2mortal(newSVpvs("malformed"));
        } else {
            ST(0) = sv_2mortal(newSVpvf("%lx:%lu", (unsigned long)uv, (unsigned long)used));
        }
        XSRETURN(1);
    }
}

/* Puts layers on a handle with PerlIO_binmode (how 0) or
 * PerlIO_apply_layers (how 1); returns what the function returned. */
XS(XS_PerlAPI_pio_layers) {
    dXSARGS;
    if (items != 3) croak_xs_usage(cv, "handle, layers, how");
    {
        PerlIO *io = IoIFP(sv_2io(ST(0)));
        const char *layers = SvPV_nolen(ST(1));
        int r;
        if (SvIV(ST(2)) == 0)
            r = PerlIO_binmode(aTHX_ io, '<', O_BINARY, layers);
        else
            r = PerlIO_apply_layers(aTHX_ io, "r", layers);
        XSRETURN_IV(r);
    }
}

/* Opens a file that may not exist and, with a handle, seeks on it: for
 * each of the two whether it failed and the errno native code finds
 * afterwards. */
XS(XS_PerlAPI_pio_failures) {
    dXSARGS;
    if (items != 2) croak_xs_usage(cv, "path, handle");
    {
        PerlIO *opened, *io = IoIFP(sv_2io(ST(1)));
        int open_failed, open_errno, seek_failed, seek_errno;
        errno = 0;
        opened = PerlIO_open(SvPV_nolen(ST(0)), "r");
        open_failed = opened == NULL;
        open_errno = errno;
        if (opened) PerlIO_close(opened);
        errno = 0;
        seek_failed = PerlIO_seek(io, 0, SEEK_SET) != 0;
        seek_errno = errno;
        errno = 0; /* nothing of this is the XSUB's own failure */
        SP -= items;
        EXTEND(SP, 4);
        mPUSHi(open_failed);
        mPUSHs(newSVpv(open_errno == ENOENT ? "ENOENT" : open_errno == 0 ? "none" : strerror(open_errno), 0));
        mPUSHi(seek_failed);
        mPUSHs(newSVpv(seek_errno == ESPIPE ? "ESPIPE" : seek_errno == EINVAL ? "EINVAL" : seek_errno == 0 ? "none" : strerror(seek_errno), 0));
        PUTBACK;
        return;
    }
}

/* Two functions for one sub, PerlAPI::impl, and an XSUB that exchanges
 * them by assigning to CvXSUB: it returns whether the second is installed
 * afterwards. */
XS(XS_PerlAPI_impl_one) {
    dXSARGS;
    PERL_UNUSED_VAR(items);
    ST(0) = sv_2mortal(newSVpvs("one"));
    XSRETURN(1);
}
XS(XS_PerlAPI_impl_two) {
    dXSARGS;
    PERL_UNUSED_VAR(items);
    ST(0) = sv_2mortal(newSVpvs("two"));
    XSRETURN(1);
}
XS(XS_PerlAPI_exchange) {
    dXSARGS;
    PERL_UNUSED_VAR(items);
    {
        CV *target = get_cv("PerlAPI::impl", 0);
        if (CvXSUB(target) == XS_PerlAPI_impl_one)
            CvXSUB(target) = XS_PerlAPI_impl_two;
        else
            CvXSUB(target) = XS_PerlAPI_impl_one;
        ST(0) = boolSV(CvXSUB(target) == XS_PerlAPI_impl_two);
        XSRETURN(1);
    }
}

/* Assigns to the method-cache generation of a glob and to its stash:
 * returns the generation after `= 7` and after `= 0`, and the name of the
 * stash read back after `GvSTASH(gv) = ...`. */
XS(XS_PerlAPI_glob_fields) {
    dXSARGS;
    if (items != 2) croak_xs_usage(cv, "name, package");
    {
        GV *gv = gv_fetchpv(SvPV_nolen(ST(0)), GV_ADD, SVt_PVCV);
        U32 set, cleared;
        const char *stash_name;
        GvCVGEN(gv) = 7;
        set = GvCVGEN(gv);
        GvCVGEN(gv) = 0;
        cleared = GvCVGEN(gv);
        GvSTASH(gv) = gv_stashpv(SvPV_nolen(ST(1)), GV_ADD);
        stash_name = HvNAME(GvSTASH(gv));
        SP -= items;
        EXTEND(SP, 3);
        mPUSHu(set);
        mPUSHu(cleared);
        mPUSHs(newSVpv(stash_name, 0));
        PUTBACK;
        return;
    }
}

/* A copy of a chain of references, made the way Clone does it: the
 * referent of the new reference is assigned through SvRV, and the right
 * side of that assignment comes back to the same line for every level. */
static SV *PerlAPI_clone_chain(pTHX_ SV *sv) {
    SV *clone;
    if (!SvROK(sv)) return newSVsv(sv);
    clone = newSV(0);
    sv_upgrade(clone, SVt_IV);
    SvRV(clone) = PerlAPI_clone_chain(aTHX_ SvRV(sv));
    SvROK_on(clone);
    return clone;
}
XS(XS_PerlAPI_clone_chain) {
    dXSARGS;
    if (items != 1) croak_xs_usage(cv, "ref");
    ST(0) = sv_2mortal(PerlAPI_clone_chain(aTHX_ ST(0)));
    XSRETURN(1);
}

/* Makes the first reference point at what the second points at, by
 * assigning to SvRV; returns what SvRV says afterwards, as a reference. */
XS(XS_PerlAPI_repoint) {
    dXSARGS;
    if (items != 2) croak_xs_usage(cv, "ref, to");
    {
        SV *old = SvRV(ST(0));
        SvRV(ST(0)) = SvREFCNT_inc(SvRV(ST(1)));
        SvREFCNT_dec(old);
        ST(0) = sv_2mortal(newRV_inc(SvRV(ST(0))));
        XSRETURN(1);
    }
}

/* The address of the integer of a scalar, from a function of its own. */
static IV *PerlAPI_integer_of(pTHX_ SV *sv) { return &SvIVX(sv); }

/* Adds one to the integer of every argument through addresses that were
 * all taken first, by one accessor expression in a loop and by a function
 * that returns the address; then reads every referent of an array of
 * references and returns how many are defined. */
XS(XS_PerlAPI_kept_addresses) {
    dXSARGS;
    {
        IV *at[8];
        IV *first, *second;
        I32 i, n = items - 1 > 8 ? 8 : items - 1;
        IV defined_referents = 0;
        AV *refs;
        if (items < 3) croak_xs_usage(cv, "refs, a, b, ...");
        refs = (AV *)SvRV(ST(0));
        for (i = 0; i < n; i++) at[i] = &SvIVX(ST(i + 1));
        first = PerlAPI_integer_of(aTHX_ ST(1));
        second = PerlAPI_integer_of(aTHX_ ST(2));
        for (i = 0; i < n; i++) *at[i] += 1;
        (void)SvIV(ST(1));
        *first += 10;
        *second += 20;
        for (i = 0; i <= av_len(refs); i++) {
            SV **e = av_fetch(refs, i, 0);
            if (e && SvROK(*e) && SvOK(SvRV(*e))) defined_referents++;
        }
        ST(0) = sv_2mortal(newSViv(defined_referents));
        XSRETURN(1);
    }
}

/* Replaces the scalar, the array and the hash of a glob by assigning to
 * GvSV, GvAV and GvHV, as code that installs variables does; the values
 * that were there are released. */
XS(XS_PerlAPI_glob_slots) {
    dXSARGS;
    if (items != 1) croak_xs_usage(cv, "name");
    {
        GV *gv = gv_fetchpv(SvPV_nolen(ST(0)), GV_ADD, SVt_PV);
        SV *old_sv = GvSV(gv);
        AV *old_av = GvAV(gv);
        HV *old_hv = GvHV(gv);
        AV *av = newAV();
        HV *hv = newHV();
        av_push(av, newSViv(1));
        av_push(av, newSViv(2));
        (void)hv_store(hv, "key", 3, newSVpvs("value"), 0);
        GvSV(gv) = newSVpvs("scalar");
        GvAV(gv) = av;
        GvHV(gv) = hv;
        SvREFCNT_dec(old_sv);
        SvREFCNT_dec((SV *)old_av);
        SvREFCNT_dec((SV *)old_hv);
        ST(0) = boolSV(GvSV(gv) && GvAV(gv) == av && GvHV(gv) == hv);
        XSRETURN(1);
    }
}

/* An op checker for additions, installed with wrap_op_checker and taken
 * out by putting back the checker it found. It calls that checker, counts
 * the ops it sees, and gives each addition a pp function of this module:
 * the sum and a thousand, which it leaves on the stack in place of the two
 * operands. */
static Perl_check_t PerlAPI_add_hook_before;
static IV PerlAPI_adds_seen;
/* How often the pp function below found the mark pointer of the SDK naming
 * the marks of another frame when it was entered. (This looks into the SDK:
 * PL_markstack_ptr itself brings the pointer in line when it is read. A
 * pointer left by a frame that is gone is taken for this frame's when this
 * frame lies where that one was.) */
static IV PerlAPI_adds_with_foreign_marks;
static OP *PerlAPI_pp_add_more(pTHX) {
    if (goperl_markstack_owner_v != aTHX ||
        goperl_markstack_ptr_v != aTHX->marks + aTHX->markidx)
        PerlAPI_adds_with_foreign_marks++;
    dSP;
    SV *right = POPs;
    SV *left = TOPs;
    SETs(sv_2mortal(newSViv(SvIV(left) + SvIV(right) + 1000)));
    RETURN;
}
static OP *PerlAPI_on_new_add(pTHX_ OP *o) {
    o = PerlAPI_add_hook_before(aTHX_ o);
    PerlAPI_adds_seen++;
    /* two operands that are not constants: the op is still an addition,
     * and its operands are below it */
    if (o->op_type == OP_ADD && (o->op_flags & OPf_KIDS) &&
        cBINOPo->op_first && OpSIBLING(cBINOPo->op_first) == cBINOPo->op_last &&
        o->op_ppaddr == PL_ppaddr[OP_ADD])
        o->op_ppaddr = PerlAPI_pp_add_more;
    return o;
}

/* An op checker for join that is assigned straight into PL_check. Its pp
 * function looks at the list through the mark, replaces the separator,
 * and runs the interpreter's join on what it left; how many values it saw
 * is kept for the test. */
static Perl_check_t PerlAPI_join_hook_before;
static IV PerlAPI_join_values;
static OP *PerlAPI_pp_join_plus(pTHX) {
    dSP;
    SV **mark = PL_stack_base + TOPMARK;
    PerlAPI_join_values = (IV)(SP - mark) - 1;
    mark[1] = sv_2mortal(newSVpvs("+"));
    return PL_ppaddr[OP_JOIN](aTHX);
}
static OP *PerlAPI_on_new_join(pTHX_ OP *o) {
    o = PerlAPI_join_hook_before(aTHX_ o);
    if (o->op_type == OP_JOIN) o->op_ppaddr = PerlAPI_pp_join_plus;
    return o;
}

/* A peephole hook: it runs the optimizer it replaced, then follows the
 * execution order of what it was handed and remembers every join it has
 * not seen before. */
static peep_t PerlAPI_rpeep_before;
static OP *PerlAPI_joins_seen[64];
static IV PerlAPI_njoins_seen;
static void PerlAPI_rpeep(pTHX_ OP *o) {
    IV steps = 0;
    PerlAPI_rpeep_before(aTHX_ o);
    for (; o && steps < 100000; o = o->op_next, steps++) {
        IV i;
        if (o->op_type != OP_JOIN) continue;
        for (i = 0; i < PerlAPI_njoins_seen; i++)
            if (PerlAPI_joins_seen[i] == o) break;
        if (i == PerlAPI_njoins_seen && i < 64) PerlAPI_joins_seen[PerlAPI_njoins_seen++] = o;
    }
}

/* An op checker for sprintf that links the op's execution order itself
 * (LINKLIST, as Syntax::Keyword::Inplace does before it threads an op of
 * its own in) and then follows the links from where the order starts: how
 * many ops come before the op itself, or -1 when the links do not lead to
 * it. */
static Perl_check_t PerlAPI_sprintf_hook_before;
static IV PerlAPI_linklist_seen;
static OP *PerlAPI_on_new_sprintf(pTHX_ OP *o) {
    o = PerlAPI_sprintf_hook_before(aTHX_ o);
    if (o->op_type == OP_SPRINTF && (o->op_flags & OPf_KIDS)) {
        OP *kid = LINKLIST(o);
        IV n = 0;
        for (; kid && kid != o && n < 100; kid = kid->op_next) n++;
        PerlAPI_linklist_seen = kid == o ? n : -1;
    }
    return o;
}

/* An op checker for negation whose pp function sets $? to 99, saved: the
 * save is the interpreter's, undone when the scope the op runs in ends (the
 * sub's), not when the pp function returns. */
static Perl_check_t PerlAPI_negate_hook_before;
static IV PerlAPI_negates_seen;
static OP *PerlAPI_pp_negate_status(pTHX) {
    SAVEI32(PL_statusvalue);
    PL_statusvalue = 99;
    return PL_ppaddr[OP_NEGATE](aTHX);
}
static OP *PerlAPI_on_new_negate(pTHX_ OP *o) {
    o = PerlAPI_negate_hook_before(aTHX_ o);
    if (o->op_type == OP_NEGATE) {
        o->op_ppaddr = PerlAPI_pp_negate_status;
        PerlAPI_negates_seen++;
    }
    return o;
}

/* hooks(what, on): installs or removes one of the hooks above and returns
 * its counter. */
XS(XS_PerlAPI_hooks) {
    dXSARGS;
    if (items != 2) croak_xs_usage(cv, "what, on");
    {
        const char *what = SvPV_nolen(ST(0));
        const int on = SvTRUE(ST(1));
        IV counter = 0;
        if (strEQ(what, "add")) {
            if (on) {
                wrap_op_checker(OP_ADD, PerlAPI_on_new_add, &PerlAPI_add_hook_before);
            } else if (PerlAPI_add_hook_before) {
                PL_check[OP_ADD] = PerlAPI_add_hook_before;
                PerlAPI_add_hook_before = NULL;
            }
            counter = PerlAPI_adds_seen;
        } else if (strEQ(what, "add with foreign marks")) {
            counter = PerlAPI_adds_with_foreign_marks;
        } else if (strEQ(what, "join")) {
            if (on) {
                PerlAPI_join_hook_before = PL_check[OP_JOIN];
                PL_check[OP_JOIN] = PerlAPI_on_new_join;
            } else if (PerlAPI_join_hook_before) {
                PL_check[OP_JOIN] = PerlAPI_join_hook_before;
                PerlAPI_join_hook_before = NULL;
            }
            counter = PerlAPI_join_values;
        } else if (strEQ(what, "peep")) {
            if (on) {
                PerlAPI_njoins_seen = 0;
                PerlAPI_rpeep_before = PL_rpeepp;
                PL_rpeepp = PerlAPI_rpeep;
            } else if (PerlAPI_rpeep_before) {
                PL_rpeepp = PerlAPI_rpeep_before;
                PerlAPI_rpeep_before = NULL;
            }
            counter = PerlAPI_njoins_seen;
        } else if (strEQ(what, "status")) {
            if (on) {
                wrap_op_checker(OP_NEGATE, PerlAPI_on_new_negate, &PerlAPI_negate_hook_before);
            } else if (PerlAPI_negate_hook_before) {
                PL_check[OP_NEGATE] = PerlAPI_negate_hook_before;
                PerlAPI_negate_hook_before = NULL;
            }
            counter = PerlAPI_negates_seen;
        } else if (strEQ(what, "linklist")) {
            if (on) {
                PerlAPI_linklist_seen = 0;
                wrap_op_checker(OP_SPRINTF, PerlAPI_on_new_sprintf, &PerlAPI_sprintf_hook_before);
            } else if (PerlAPI_sprintf_hook_before) {
                PL_check[OP_SPRINTF] = PerlAPI_sprintf_hook_before;
                PerlAPI_sprintf_hook_before = NULL;
            }
            counter = PerlAPI_linklist_seen;
        } else {
            croak("unknown hook %s", what);
        }
        ST(0) = sv_2mortal(newSViv(counter));
        XSRETURN(1);
    }
}

/* A function for call_atexit: it writes a line to the file it was given
 * the name of, when the interpreter is destroyed. */
static void PerlAPI_write_at_exit(pTHX_ void *arg) {
    char *path = (char *)arg;
    FILE *out = fopen(path, "w");
    if (out) {
        fputs("the interpreter is gone\n", out);
        fclose(out);
    }
    free(path);
}
XS(XS_PerlAPI_at_exit_write) {
    dXSARGS;
    if (items != 1) croak_xs_usage(cv, "path");
    call_atexit(PerlAPI_write_at_exit, strdup(SvPV_nolen(ST(0))));
    XSRETURN_EMPTY;
}

/* rx_info(qr, text): the pattern of a compiled regexp, whether it folds
 * case, and where it and its first group match in the text (through
 * pregexec on the text's own buffer and the offsets of the regexp). */
XS(XS_PerlAPI_rx_info) {
    dXSARGS;
    if (items != 2) croak_xs_usage(cv, "qr, text");
    {
        REGEXP *rx = SvRX(ST(0));
        SV *text = ST(1);
        STRLEN len;
        char *p = SvPV(text, len);
        SV *out;
        if (!rx) croak("not a regexp");
        out = newSVpvn(RX_PRECOMP(rx), RX_PRELEN(rx));
        sv_catpvf(out, ":%s", (RX_EXTFLAGS(rx) & RXf_PMf_FOLD) ? "fold" : "exact");
        if (pregexec(rx, p, p + len, p, 0, text, 1)) {
            regexp *r = ReANY(rx);
            sv_catpvf(out, ":%d-%d", (int)r->offs[0].start, (int)r->offs[0].end);
            if (RX_NPARENS(rx) >= 1)
                sv_catpvf(out, ":%d-%d", (int)RX_OFFS(rx)[1].start, (int)RX_OFFS(rx)[1].end);
            sv_catpvf(out, ":%d", (int)((regexp *)SvANY(rx))->nparens);
        } else {
            sv_catpvs(out, ":no match");
        }
        ST(0) = sv_2mortal(out);
        XSRETURN(1);
    }
}

/* rx_compile(pattern, text): compiles the pattern with pregcomp and matches
 * a copy of the text that is this module's own memory; returns where. */
XS(XS_PerlAPI_rx_compile) {
    dXSARGS;
    if (items != 2) croak_xs_usage(cv, "pattern, text");
    {
        REGEXP *rx = pregcomp(ST(0), 0);
        STRLEN len;
        const char *p = SvPV(ST(1), len);
        char *own = (char *)malloc(len + 1);
        SV *out = newSVpvs("");
        memcpy(own, p, len + 1);
        if (pregexec(rx, own, own + len, own, 0, ST(1), 1))
            sv_catpvf(out, "%d-%d", (int)RX_OFFS(rx)[0].start, (int)RX_OFFS(rx)[0].end);
        else
            sv_catpvs(out, "no match");
        free(own);
        ReREFCNT_dec(rx);
        ST(0) = sv_2mortal(out);
        XSRETURN(1);
    }
}

/* own_magic(ref): what mg_find says of the magic the interpreter itself
 * keeps on the referent: "backref" for one with weak references to it,
 * "amt:<n>" for a stash with an overload table, n being 1 when the table
 * has an array dereference. */
XS(XS_PerlAPI_own_magic) {
    dXSARGS;
    if (items != 1) croak_xs_usage(cv, "ref");
    {
        SV *sv = SvRV(ST(0));
        SV *out = newSVpvs("");
        MAGIC *mg;
        if (SvRMAGICAL(sv) && mg_find(sv, PERL_MAGIC_backref)) sv_catpvs(out, "backref");
        if (SvOBJECT(sv) && Gv_AMG(SvSTASH(sv)) &&
            (mg = mg_find((SV *)SvSTASH(sv), PERL_MAGIC_overload_table)) != NULL) {
            const AMT *amt = (AMT *)mg->mg_ptr;
            sv_catpvf(out, "amt:%d", (AMT_AMAGIC(amt) && amt->table[to_av_amg]) ? 1 : 0);
        }
        ST(0) = sv_2mortal(out);
        XSRETURN(1);
    }
}

/* A call checker: every call of the sub it is given to compiles into the
 * value it was given with. */
static OP *PerlAPI_compile_to_constant(pTHX_ OP *entersubop, GV *namegv, SV *ckobj) {
    PERL_UNUSED_ARG(namegv);
    op_free(entersubop);
    return newSVOP(OP_CONST, 0, newSVsv(ckobj));
}
/* call_checker(code, value): gives the sub the checker above; returns
 * whether the sub had the interpreter's own checker before and has this
 * one after. */
XS(XS_PerlAPI_constant_calls) {
    dXSARGS;
    if (items != 2) croak_xs_usage(cv, "code, value");
    {
        CV *target = (CV *)SvRV(ST(0));
        Perl_call_checker before, after;
        SV *obj_before, *obj_after;
        cv_get_call_checker(target, &before, &obj_before);
        cv_set_call_checker(target, PerlAPI_compile_to_constant, ST(1));
        cv_get_call_checker(target, &after, &obj_after);
        ST(0) = sv_2mortal(newSVpvf(
            "%s,%s,%s", before == Perl_ck_entersub_args_proto_or_list ? "default" : "other",
            after == PerlAPI_compile_to_constant ? "ours" : "other", SvPV_nolen(obj_after)));
        XSRETURN(1);
    }
}

/* A destructor for a variable of the function that registers it: it notes
 * what it finds there. */
struct PerlAPI_guarded {
    int marker;
    char *buffer;
};
static int PerlAPI_destructor_found;
static void PerlAPI_release_guarded(pTHX_ void *arg) {
    struct PerlAPI_guarded *g = (struct PerlAPI_guarded *)arg;
    PerlAPI_destructor_found = g->marker;
    free(g->buffer);
    g->buffer = NULL;
}
/* croak_guarded(nested): registers the destructor for a local (inside a
 * scope of its own when nested is true) and croaks; destructor_found says
 * what the destructor saw. */
XS(XS_PerlAPI_croak_guarded) {
    dXSARGS;
    if (items != 1) croak_xs_usage(cv, "nested");
    {
        struct PerlAPI_guarded g;
        const int nested = SvTRUE(ST(0));
        g.marker = 4711;
        g.buffer = (char *)malloc(64);
        PerlAPI_destructor_found = 0;
        if (nested) {
            ENTER;
            SAVETMPS;
        }
        SAVEDESTRUCTOR_X(PerlAPI_release_guarded, &g);
        croak("guarded failure");
    }
}
XS(XS_PerlAPI_destructor_found) {
    dXSARGS;
    PERL_UNUSED_VAR(items);
    ST(0) = sv_2mortal(newSViv(PerlAPI_destructor_found));
    XSRETURN(1);
}

/* show(text, max): pv_pretty of the text in quotes, pv_display of it, and
 * how much of it pv_escape took. */
XS(XS_PerlAPI_show) {
    dXSARGS;
    if (items != 2) croak_xs_usage(cv, "text, max");
    {
        STRLEN len, taken = 0;
        const char *p = SvPV(ST(0), len);
        SV *a = sv_2mortal(newSVpvs(""));
        SV *b = sv_2mortal(newSVpvs(""));
        SV *c = sv_2mortal(newSVpvs(""));
        SV *out = newSVpvs("");
        pv_pretty(a, p, len, (STRLEN)SvIV(ST(1)), NULL, NULL, PERL_PV_PRETTY_DUMP);
        pv_display(b, p, len, len + 1, 0);
        pv_escape(c, p, len, (STRLEN)SvIV(ST(1)), &taken, 0);
        sv_catpvf(out, "%s %s %d", SvPV_nolen(a), SvPV_nolen(b), (int)taken);
        ST(0) = sv_2mortal(out);
        XSRETURN(1);
    }
}

/* Magic that watches a variable: its vtable counts the reads, the writes
 * and the clears, and on a read of a scalar it gives the scalar the number
 * of reads so far when the scalar holds "count". */
static IV PerlAPI_observed[3];
static int PerlAPI_observe_get(pTHX_ SV *sv, MAGIC *mg) {
    PERL_UNUSED_ARG(mg);
    PerlAPI_observed[0]++;
    if (SvTYPE(sv) < SVt_PVAV && SvPOK(sv) && strEQ(SvPVX(sv), "count"))
        sv_setiv(sv, PerlAPI_observed[0]);
    return 0;
}
static int PerlAPI_observe_set(pTHX_ SV *sv, MAGIC *mg) {
    PERL_UNUSED_ARG(sv);
    PERL_UNUSED_ARG(mg);
    PerlAPI_observed[1]++;
    return 0;
}
static int PerlAPI_observe_clear(pTHX_ SV *sv, MAGIC *mg) {
    PERL_UNUSED_ARG(sv);
    PERL_UNUSED_ARG(mg);
    PerlAPI_observed[2]++;
    return 0;
}
static MGVTBL PerlAPI_observe_vtbl = {PerlAPI_observe_get, PerlAPI_observe_set, NULL,
                                    PerlAPI_observe_clear, NULL, NULL, NULL, NULL};
/* observe(ref): attaches the magic; observed(): the three counts, which start
 * again from zero. */
XS(XS_PerlAPI_observe) {
    dXSARGS;
    if (items != 1) croak_xs_usage(cv, "ref");
    sv_magicext(SvRV(ST(0)), NULL, PERL_MAGIC_ext, &PerlAPI_observe_vtbl, NULL, 0);
    XSRETURN_EMPTY;
}
XS(XS_PerlAPI_observed) {
    dXSARGS;
    PERL_UNUSED_VAR(items);
    SP -= items;
    EXTEND(SP, 3);
    mPUSHi(PerlAPI_observed[0]);
    mPUSHi(PerlAPI_observed[1]);
    mPUSHi(PerlAPI_observed[2]);
    PerlAPI_observed[0] = PerlAPI_observed[1] = PerlAPI_observed[2] = 0;
    PUTBACK;
    return;
}

/* uvar magic for a hash: the function is called for every access with the
 * key in the MAGIC's mg_obj, counts them, and makes the key "alias" the
 * key "real". */
static IV PerlAPI_uvar_calls;
static struct ufuncs PerlAPI_uvar_funcs;
static I32 PerlAPI_uvar_key(pTHX_ IV action, SV *hv) {
    MAGIC *mg;
    PERL_UNUSED_ARG(action);
    PerlAPI_uvar_calls++;
    for (mg = SvMAGIC(hv); mg; mg = mg->mg_moremagic)
        if (mg->mg_type == PERL_MAGIC_uvar && (struct ufuncs *)mg->mg_ptr == &PerlAPI_uvar_funcs)
            break;
    if (mg && mg->mg_obj && SvPOK(mg->mg_obj) && strEQ(SvPV_nolen(mg->mg_obj), "alias"))
        mg->mg_obj = sv_2mortal(newSVpvs("real"));
    return 0;
}
XS(XS_PerlAPI_uvar_hash) {
    dXSARGS;
    if (items != 1) croak_xs_usage(cv, "hashref");
    PerlAPI_uvar_funcs.uf_val = PerlAPI_uvar_key;
    PerlAPI_uvar_funcs.uf_set = NULL;
    PerlAPI_uvar_funcs.uf_index = 0;
    sv_magic(SvRV(ST(0)), NULL, PERL_MAGIC_uvar, (char *)&PerlAPI_uvar_funcs, 0);
    XSRETURN_EMPTY;
}
XS(XS_PerlAPI_uvar_calls) {
    dXSARGS;
    PERL_UNUSED_VAR(items);
    ST(0) = sv_2mortal(newSViv(PerlAPI_uvar_calls));
    PerlAPI_uvar_calls = 0;
    XSRETURN(1);
}

/* firsts(arrays): the first element of every array of the list, read
 * from the array's body, and the last element of each array replaced by
 * the count of its elements through the body. */
XS(XS_PerlAPI_firsts) {
    dXSARGS;
    if (items != 1) croak_xs_usage(cv, "arrays");
    {
        AV *arrays = (AV *)SvRV(ST(0));
        SSize_t n = av_len(arrays) + 1;
        SV *joined = sv_2mortal(newSVpvs(""));
        SSize_t i;
        for (i = 0; i < n; i++) {
            AV *av = (AV *)SvRV(AvARRAY(arrays)[i]);
            SV **body = AvARRAY(av);
            SSize_t last = av_len(av);
            if (last < 0) continue;
            sv_catsv(joined, body[0]);
            if (last > 0) {
                SvREFCNT_dec(body[last]);
                body[last] = newSViv((IV)(last + 1));
            }
        }
        ST(0) = joined;
        XSRETURN(1);
    }
}

/* take_buffer(text): the text, by way of an SV made with room for it
 * whose buffer this code takes over (the SV is left without one and freed)
 * and frees itself, as code that builds a table in the buffer of an SV and
 * keeps the table does. */
XS(XS_PerlAPI_take_buffer) {
    dXSARGS;
    if (items != 1) croak_xs_usage(cv, "text");
    {
        STRLEN len;
        const char *text = SvPV(ST(0), len);
        SV *sv = newSV(len + 16);
        char *buf;
        STRLEN room;
        int emptied;
        SV *copy;
        sv_setpvn(sv, text, len);
        buf = SvPVX(sv);
        room = SvLEN(sv);
        SvPVX(sv) = NULL;
        SvLEN(sv) = 0;
        emptied = SvPVX(sv) == NULL && SvLEN(sv) == 0;
        SvREFCNT_dec(sv);
        copy = newSVpvf("%d:%d:%.*s", emptied, room > len, (int)len, buf);
        Safefree(buf);
        ST(0) = sv_2mortal(copy);
        XSRETURN(1);
    }
}

/* const_ops(n): builds and frees n const ops, each from an SV of its own;
 * the count of those whose SV the op holds. */
XS(XS_PerlAPI_const_ops) {
    dXSARGS;
    if (items != 1) croak_xs_usage(cv, "n");
    {
        IV n = SvIV(ST(0));
        IV held = 0;
        IV i;
        for (i = 0; i < n; i++) {
            SV *sv = newSVpvf("constant %d", (int)i);
            OP *o = newSVOP(OP_CONST, 0, sv);
            if (cSVOPx(o)->op_sv == sv) held++;
            op_free(o);
        }
        ST(0) = sv_2mortal(newSViv(held));
        XSRETURN(1);
    }
}

/* spin(n): waits a millisecond n times, or until $PerlAPI::stop is true,
 * letting the interpreter deliver what is pending after each wait; the
 * count of waits. */
XS(XS_PerlAPI_spin) {
    dXSARGS;
    if (items != 1) croak_xs_usage(cv, "n");
    {
        IV n = SvIV(ST(0));
        IV i;
        SV *stop = get_sv("PerlAPI::stop", GV_ADD);
        for (i = 0; i < n && !SvTRUE(stop); i++) {
            struct timespec ms = {0, 1000000};
            nanosleep(&ms, NULL);
            PERL_ASYNC_CHECK();
        }
        ST(0) = sv_2mortal(newSViv(i));
        XSRETURN(1);
    }
}

/* by_number(n): keys a hash by the bytes of the integers 0 .. n-1 (bytes
 * that are 0 among them), then looks every one up, deletes the odd ones
 * and counts what is left: "stored:found:exist after:keys left". */
XS(XS_PerlAPI_by_number) {
    dXSARGS;
    if (items != 1) croak_xs_usage(cv, "n");
    {
        int n = (int)SvIV(ST(0));
        HV *hv = (HV *)sv_2mortal((SV *)newHV());
        int stored = 0, found = 0, exist = 0;
        int i;
        for (i = 0; i < n; i++)
            if (hv_store(hv, (char *)&i, sizeof(i), newSViv(i), 0)) stored++;
        for (i = 0; i < n; i++) {
            SV **svp = hv_fetch(hv, (char *)&i, sizeof(i), 0);
            if (svp && SvIV(*svp) == i) found++;
        }
        for (i = 1; i < n; i += 2) hv_delete(hv, (char *)&i, sizeof(i), G_DISCARD);
        for (i = 0; i < n; i++)
            if (hv_exists(hv, (char *)&i, sizeof(i))) exist++;
        ST(0) = sv_2mortal(newSVpvf("%d:%d:%d:%d", stored, found, exist,
                                    (int)HvUSEDKEYS(hv)));
        XSRETURN(1);
    }
}

/* own_name(code [, name]): the name the sub has by itself (a lexical sub
 * has one; a sub in a glob has the glob's), "" when it has none; with a
 * name, the sub is given it first. */
XS(XS_PerlAPI_own_name) {
    dXSARGS;
    if (items < 1 || items > 2) croak_xs_usage(cv, "code [, name]");
    {
        CV *code = (CV *)SvRV(ST(0));
        HEK *hek;
        if (items == 2) {
            STRLEN len;
            const char *name = SvPV(ST(1), len);
            U32 hash;
            PERL_HASH(hash, name, len);
            CvNAME_HEK_set(code, share_hek(name, SvUTF8(ST(1)) ? -(I32)len : (I32)len, hash));
        }
        hek = CvNAME_HEK(code);
        if (!hek) {
            ST(0) = sv_2mortal(newSVpvs(""));
        } else {
            SV *name = newSVpvn(HEK_KEY(hek), HEK_LEN(hek));
            if (HEK_UTF8(hek)) SvUTF8_on(name);
            ST(0) = sv_2mortal(name);
        }
        XSRETURN(1);
    }
}

/* The op of the type below the op, first found. */
static OP *perlapi_find_op(pTHX_ OP *o, OPCODE type) {
    OP *kid;
    if (!o) return NULL;
    if (o->op_type == type) return o;
    if (!(o->op_flags & OPf_KIDS)) return NULL;
    for (kid = cUNOPx(o)->op_first; kid; kid = OpSIBLING(kid)) {
        OP *found = perlapi_find_op(aTHX_ kid, type);
        if (found) return found;
    }
    return NULL;
}

/* signature([code]): what the sub's signature op says of its parameters:
 * "min:positional:named:slurpy:names of the named ones", or "none".
 * Without a sub, the op found by the call before is read again (the sub
 * has to exist still), as code that keeps the op of a signature does. */
static OP *perlapi_signature_op;
XS(XS_PerlAPI_signature) {
    dXSARGS;
    if (items > 1) croak_xs_usage(cv, "[code]");
    {
        OP *o = items ? perlapi_find_op(aTHX_ CvROOT((CV *)SvRV(ST(0))), OP_MULTIPARAM)
                      : perlapi_signature_op;
        SV *out;
        perlapi_signature_op = o;
        if (!o) {
            out = newSVpvs("none");
        } else {
            struct op_multiparam_aux *aux =
                (struct op_multiparam_aux *)cUNOP_AUXx(o)->op_aux;
            size_t i;
            out = newSVpvf("%d:%d:%d:%s", (int)aux->min_args,
                           (int)aux->n_positional, (int)aux->n_named,
                           aux->slurpy == '@' ? "@" : aux->slurpy == '%' ? "%" : "-");
            for (i = 0; i < aux->n_positional; i++)
                sv_catpvf(out, ":%s", aux->param_padix[i] ? "p" : "_");
            for (i = 0; i < aux->n_named; i++)
                sv_catpvf(out, ":%.*s%s", (int)aux->named[i].namelen,
                          aux->named[i].namepv,
                          aux->named[i].is_required ? "!" : "");
        }
        ST(0) = sv_2mortal(out);
        XSRETURN(1);
    }
}

/* The comparison of two values by a block that reads $a and $b, in a
 * function that is handed nothing of the MULTICALL but multicall_cop, as
 * List::MoreUtils' sort hands it on. */
static IV PerlAPI_compare_in_block(pTHX_ SV *left, SV *right, OP *multicall_cop) {
    SV *olda = GvSV(PL_firstgv);
    SV *oldb = GvSV(PL_secondgv);
    GvSV(PL_firstgv) = SvREFCNT_inc_simple_NN(left);
    GvSV(PL_secondgv) = SvREFCNT_inc_simple_NN(right);
    SvREFCNT_dec(olda);
    SvREFCNT_dec(oldb);
    MULTICALL;
    return SvIV(*PL_stack_sp);
}

/* largest(code, values...): the value that the block, comparing $a and $b,
 * puts last; then what the overloaded <=> of the first two values says, or
 * "plain" when neither overloads it. */
XS(XS_PerlAPI_largest) {
    dXSARGS;
    if (items < 3) croak_xs_usage(cv, "code, value, value, ...");
    {
        dMULTICALL;
        I32 gimme = GIMME_V; /* the context largest itself was called in */
        I32 i, best = 1;
        CV *code = (CV *)SvRV(ST(0));
        SV **args = &PL_stack_base[ax];
        SV *out;
        PUSH_MULTICALL(code);
        SAVEGENERICSV(PL_firstgv);
        SAVEGENERICSV(PL_secondgv);
        PL_firstgv = MUTABLE_GV(SvREFCNT_inc(gv_fetchpvs("a", GV_ADD | GV_NOTQUAL, SVt_PV)));
        PL_secondgv = MUTABLE_GV(SvREFCNT_inc(gv_fetchpvs("b", GV_ADD | GV_NOTQUAL, SVt_PV)));
        save_gp(PL_firstgv, 0);
        save_gp(PL_secondgv, 0);
        GvINTRO_off(PL_firstgv);
        GvINTRO_off(PL_secondgv);
        SAVEGENERICSV(GvSV(PL_firstgv));
        SvREFCNT_inc(GvSV(PL_firstgv));
        SAVEGENERICSV(GvSV(PL_secondgv));
        SvREFCNT_inc(GvSV(PL_secondgv));
        for (i = 2; i < items; i++)
            if (PerlAPI_compare_in_block(aTHX_ args[best], args[i], multicall_cop) < 0) best = i;
        POP_MULTICALL;
        out = newSVsv(args[best]);
        if (SvAMAGIC(args[1]) || SvAMAGIC(args[2]))
            sv_catpvf(out, ",overloaded %" IVdf, SvIV(amagic_call(args[1], args[2], ncmp_amg, 0)));
        else
            sv_catpvs(out, ",plain");
        ST(0) = sv_2mortal(out);
        XSRETURN(1);
    }
}

/* each_pair(code, values...): what the block returns for each pair of
 * neighbours ($a, $b), in list context: every value of every call, read
 * off the stack of the calls the way List::MoreUtils' pairwise reads it;
 * then the arguments as they are afterwards. */
XS(XS_PerlAPI_each_pair) {
    dXSARGS;
    if (items < 3) croak_xs_usage(cv, "code, value, value, ...");
    {
        dMULTICALL;
        I32 gimme = G_LIST;
        I32 i;
        CV *code = (CV *)SvRV(ST(0));
        SV **args = &PL_stack_base[ax];
        SV *out = sv_2mortal(newSVpvs(""));
        PUSH_MULTICALL(code);
        SAVEGENERICSV(PL_firstgv);
        SAVEGENERICSV(PL_secondgv);
        PL_firstgv = MUTABLE_GV(SvREFCNT_inc(gv_fetchpvs("a", GV_ADD | GV_NOTQUAL, SVt_PV)));
        PL_secondgv = MUTABLE_GV(SvREFCNT_inc(gv_fetchpvs("b", GV_ADD | GV_NOTQUAL, SVt_PV)));
        save_gp(PL_firstgv, 0);
        save_gp(PL_secondgv, 0);
        GvINTRO_off(PL_firstgv);
        GvINTRO_off(PL_secondgv);
        SAVEGENERICSV(GvSV(PL_firstgv));
        SvREFCNT_inc(GvSV(PL_firstgv));
        SAVEGENERICSV(GvSV(PL_secondgv));
        SvREFCNT_inc(GvSV(PL_secondgv));
        for (i = 1; i + 1 < items; i++) {
            SV **j;
            SV *olda = GvSV(PL_firstgv), *oldb = GvSV(PL_secondgv);
            GvSV(PL_firstgv) = SvREFCNT_inc_simple_NN(args[i]);
            GvSV(PL_secondgv) = SvREFCNT_inc_simple_NN(args[i + 1]);
            SvREFCNT_dec(olda);
            SvREFCNT_dec(oldb);
            MULTICALL;
            for (j = PL_stack_base + 1; j <= PL_stack_sp; j++) sv_catpvf(out, "%" SVf ",", SVfARG(*j));
            sv_catpvs(out, ";");
        }
        POP_MULTICALL;
        for (i = 1; i < items; i++) sv_catpvf(out, "%s%" SVf, i == 1 ? "args " : " ", SVfARG(ST(i)));
        ST(0) = out;
        XSRETURN(1);
    }
}

/* first_index(code, values...): the index of the first value for which
 * the code is true with the value in $_, by MULTICALL, as the list
 * utilities do it. */
XS(XS_PerlAPI_first_index) {
    dXSARGS;
    if (items < 1) croak_xs_usage(cv, "code, ...");
    {
        dMULTICALL;
        I32 gimme = G_SCALAR;
        IV found = -1;
        I32 i;
        CV *code = (CV *)SvRV(ST(0));
        SV **args = &PL_stack_base[ax];
        SAVESPTR(GvSV(PL_defgv));
        PUSH_MULTICALL(code);
        for (i = 1; i < items; i++) {
            GvSV(PL_defgv) = args[i];
            MULTICALL;
            if (SvTRUE(*PL_stack_sp)) {
                found = i - 1;
                break;
            }
        }
        POP_MULTICALL;
        ST(0) = sv_2mortal(newSViv(found));
        XSRETURN(1);
    }
}

/* A warning of the utf8 category, given with Perl_warner (how 0),
 * Perl_ck_warner (1) or Perl_ck_warner_d (2). */
XS(XS_PerlAPI_category_warning) {
    dXSARGS;
    if (items != 2) croak_xs_usage(cv, "how, text");
    {
        const IV how = SvIV(ST(0));
        const char *text = SvPV_nolen(ST(1));
        if (how == 0)
            Perl_warner(aTHX_ packWARN(WARN_UTF8), "%s\n", text);
        else if (how == 1)
            Perl_ck_warner(aTHX_ packWARN(WARN_UTF8), "%s\n", text);
        else
            Perl_ck_warner_d(aTHX_ packWARN(WARN_UTF8), "%s\n", text);
        XSRETURN_YES;
    }
}

/* Arguments that count how often they are evaluated. */
static int perlapi_evaluations;
static int perlapi_counted_char(int c) {
    perlapi_evaluations++;
    return c;
}
static SV *perlapi_counted_sv(SV *sv) {
    perlapi_evaluations++;
    return sv;
}
static char **perlapi_counted_slot(char **slot) {
    perlapi_evaluations++;
    return slot;
}
static STRLEN perlapi_counted_len(STRLEN n) {
    perlapi_evaluations++;
    return n;
}

/* Uses each of a number of API macros once with an argument that counts
 * its evaluations, and returns the names of those that evaluated it more
 * or less than once (nothing when all is well). */
XS(XS_PerlAPI_evaluated_twice) {
    dXSARGS;
    if (items != 1) croak_xs_usage(cv, "sv");
    {
        SV *sv = ST(0);
        SV *report = sv_2mortal(newSVpvs(""));
        char *block = (char *)malloc(8);
        STRLEN len;
        (void)SvPV_force(sv, len);
#define PERLAPI_ONCE(name, use)                                   \
    do {                                                          \
        perlapi_evaluations = 0;                                  \
        use;                                                      \
        if (perlapi_evaluations != 1)                             \
            sv_catpvf(report, "%s:%d ", name, perlapi_evaluations); \
    } while (0)
        PERLAPI_ONCE("toLOWER", (void)toLOWER(perlapi_counted_char('A')));
        PERLAPI_ONCE("toUPPER", (void)toUPPER(perlapi_counted_char('a')));
        PERLAPI_ONCE("isALPHA", (void)isALPHA(perlapi_counted_char('1')));
        PERLAPI_ONCE("isDIGIT", (void)isDIGIT(perlapi_counted_char('1')));
        PERLAPI_ONCE("isALNUM", (void)isALNUM(perlapi_counted_char('-')));
        PERLAPI_ONCE("isSPACE", (void)isSPACE(perlapi_counted_char('x')));
        PERLAPI_ONCE("isPUNCT", (void)isPUNCT(perlapi_counted_char('a')));
        PERLAPI_ONCE("isXDIGIT", (void)isXDIGIT(perlapi_counted_char('g')));
        PERLAPI_ONCE("isIDFIRST", (void)isIDFIRST(perlapi_counted_char('1')));
        PERLAPI_ONCE("UVCHR_SKIP", (void)UVCHR_SKIP(perlapi_counted_len(0x10000)));
        PERLAPI_ONCE("SvEND", (void)SvEND(perlapi_counted_sv(sv)));
        PERLAPI_ONCE("SvUPGRADE", SvUPGRADE(perlapi_counted_sv(sv), SVt_PV));
        PERLAPI_ONCE("SvRXOK", (void)SvRXOK(perlapi_counted_sv(sv)));
        PERLAPI_ONCE("SvTAINTED", (void)SvTAINTED(perlapi_counted_sv(sv)));
        PERLAPI_ONCE("SvGETMAGIC", SvGETMAGIC(perlapi_counted_sv(sv)));
        PERLAPI_ONCE("SvGAMAGIC", (void)SvGAMAGIC(perlapi_counted_sv(sv)));
        PERLAPI_ONCE("SvIMMORTAL", (void)SvIMMORTAL(perlapi_counted_sv(sv)));
        PERLAPI_ONCE("SvOOK_off", SvOOK_off(perlapi_counted_sv(sv)));
        PERLAPI_ONCE("SvIV_please", SvIV_please(perlapi_counted_sv(sv)));
        PERLAPI_ONCE("SV_CHECK_THINKFIRST", SV_CHECK_THINKFIRST(perlapi_counted_sv(sv)));
        PERLAPI_ONCE("SvSetSV_nosteal", SvSetSV_nosteal(perlapi_counted_sv(sv), sv));
        PERLAPI_ONCE("SvREFCNT_inc", (void)SvREFCNT_inc(perlapi_counted_sv(sv)));
        PERLAPI_ONCE("SvREFCNT_dec", SvREFCNT_dec(perlapi_counted_sv(sv)));
        PERLAPI_ONCE("memEQs", (void)memEQs("abc", perlapi_counted_len(3), "abc"));
        PERLAPI_ONCE("Renew", Renew(*perlapi_counted_slot(&block), 16, char));
        PERLAPI_ONCE("Renewc", Renewc(*perlapi_counted_slot(&block), 32, char, char));
        PERLAPI_ONCE("ZeroD", (void)ZeroD(block, perlapi_counted_len(8), char));
#undef PERLAPI_ONCE
        free(block);
        ST(0) = report;
        XSRETURN(1);
    }
}

/* The UTF-8 bytes of a code point, in hex, and the length UVCHR_SKIP and
 * UTF8SKIP give for it. */
XS(XS_PerlAPI_encode_uv) {
    dXSARGS;
    if (items != 1) croak_xs_usage(cv, "uv");
    {
        U8 bytes[UTF8_MAXBYTES + 1];
        char hex[2 * UTF8_MAXBYTES + 1];
        const UV uv = SvUV(ST(0));
        U8 *end = uvchr_to_utf8(bytes, uv);
        STRLEN i, n = (STRLEN)(end - bytes);
        for (i = 0; i < n; i++) snprintf(hex + 2 * i, 3, "%02x", bytes[i]);
        ST(0) = sv_2mortal(newSVpvf("%s:%lu:%lu", hex, (unsigned long)UVCHR_SKIP(uv),
                                    (unsigned long)UTF8SKIP(bytes)));
        XSRETURN(1);
    }
}

/* Gives the glob of a name another name with gv_name_set. */
XS(XS_PerlAPI_rename_glob) {
    dXSARGS;
    if (items != 2) croak_xs_usage(cv, "glob, name");
    {
        STRLEN len;
        const char *name = SvPV(ST(1), len);
        GV *gv = gv_fetchpv(SvPV_nolen(ST(0)), GV_ADD, SVt_PV);
        gv_name_set(gv, name, len, GV_NOTQUAL);
        ST(0) = sv_2mortal(newSVpv(GvNAME(gv), 0));
        XSRETURN(1);
    }
}

/* The argument stack read and written by a plain C function, the way
 * Inline::C's glue does it: the function takes no interpreter context, says
 * dXSARGS itself (Inline_Stack_Vars) and returns through XSRETURN; the XSUB
 * that calls it un-pops the mark first so that the function finds the same
 * arguments, and looks at the mark afterwards to learn whether the function
 * left values. */
static void stack_in_a_plain_function(SV *first, ...) {
    dXSARGS;
    int i;
    IV sum = 0;
    for (i = 0; i < items; i++) sum += SvIV(ST(i));
    sp = mark;
    XPUSHs(sv_2mortal(newSViv(items)));
    XPUSHs(sv_2mortal(newSViv(sum)));
    XPUSHs(sv_2mortal(newSVsv(first)));
    PUTBACK;
    XSRETURN(3);
}

XS(XS_PerlAPI_stack_in_a_plain_function) {
    dXSARGS;
    if (items < 1) croak_xs_usage(cv, "first, ...");
    PERL_UNUSED_VAR(ax);
    SP -= items;
    {
        SV *first = ST(0);
        I32 *temp;
        temp = PL_markstack_ptr++;
        stack_in_a_plain_function(first);
        if (PL_markstack_ptr != temp) {
            /* truly void, because dXSARGS not invoked */
            PL_markstack_ptr = temp;
            XSRETURN_EMPTY; /* return empty stack */
        }
        /* must have used dXSARGS; list context implied */
        return; /* assume stack size is correct */
    }
}

/* An XSUB that reads the stack with dMARK and dSP and never says dXSARGS,
 * as hand-written XSUBs of one or two arguments do (Params::Classify). The
 * number its sub carries in CvXSUBANY is added to the sum. */
static void XS_PerlAPI_sum_without_dxsargs(pTHX_ CV *cv) {
    I32 base = CvXSUBANY(cv).any_i32;
    dMARK; dSP;
    IV sum = base;
    switch (SP - MARK) {
        case 2: sum += SvIV(SP[-1]); /* FALLTHROUGH */
        case 1: sum += SvIV(SP[0]); break;
        default: croak_xs_usage(cv, "a, [b]");
    }
    SP = MARK;
    XPUSHs(sv_2mortal(newSViv(sum)));
    PUTBACK;
}

/* One look at a value that must not run its get magic: what a module does
 * after it has called SvGETMAGIC itself. "get" is the one that does run
 * it. */
XS(XS_PerlAPI_look_without_magic) {
    dXSARGS;
    if (items != 2) croak_xs_usage(cv, "sv, how");
    {
        SV *sv = ST(0);
        const char *how = SvPV_nolen(ST(1));
        STRLEN len = 0;
        IV seen = 0;
        if (strEQ(how, "get")) { SvGETMAGIC(sv); }
        else if (strEQ(how, "ok")) seen = SvOK(sv) ? 1 : 0;
        else if (strEQ(how, "pv")) { const char *s = SvPV_nomg(sv, len); seen = (IV)len + (s ? 0 : 100); }
        else if (strEQ(how, "pv_nolen")) seen = (IV)strlen(SvPV_nomg_nolen(sv));
        else if (strEQ(how, "pvbyte")) { const char *s = SvPVbyte_nomg(sv, len); seen = (IV)len + (s ? 0 : 100); }
        else if (strEQ(how, "iv")) seen = SvIV_nomg(sv);
        else if (strEQ(how, "uv")) seen = (IV)SvUV_nomg(sv);
        else if (strEQ(how, "nv")) seen = (IV)SvNV_nomg(sv);
        else if (strEQ(how, "true")) seen = SvTRUE_nomg(sv) ? 1 : 0;
        else if (strEQ(how, "utf8")) seen = SvUTF8(sv) ? 1 : 0;
        else if (strEQ(how, "tainted")) seen = SvTAINTED(sv) ? 1 : 0;
        else if (strEQ(how, "rok")) seen = SvROK(sv) ? 1 : 0;
        else if (strEQ(how, "gmagical")) seen = SvGMAGICAL(sv) ? 1 : 0;
        else if (strEQ(how, "flags")) seen = (SvFLAGS(sv) & SVs_GMG) ? 1 : 0;
        else if (strEQ(how, "pok")) seen = SvPOK(sv) ? 1 : 0;
        else if (strEQ(how, "iok")) seen = SvIOK(sv) ? 1 : 0;
        else if (strEQ(how, "isobject")) seen = sv_isobject(sv) ? 1 : 0;
        else if (strEQ(how, "derived")) seen = sv_derived_from(sv, "No::Such::Class") ? 1 : 0;
        else if (strEQ(how, "type")) seen = (IV)SvTYPE(sv);
        else if (strEQ(how, "cur")) seen = SvPOK(sv) ? (IV)SvCUR(sv) : -1;
        else if (strEQ(how, "copy")) { SV *copy = sv_2mortal(newSVsv_nomg(sv)); seen = SvOK(copy) ? 1 : 0; }
        else croak("look_without_magic: no such look: %s", how);
        ST(0) = sv_2mortal(newSViv(seen));
        XSRETURN(1);
    }
}

/* Two keywords that declare a lexical, the way a syntax module does it:
 *
 *   one_into_x;    is  my $x = 1;  — the assignment is built with newASSIGNOP
 *   hello_into_y;  declares $y and gives its pad slot the value "Hello" at
 *                  compile time, reached through PAD_SVl
 *
 * and two that take a block:
 *
 *   in_block { ... }       is the block, parsed with parse_block
 *   counted_block { ... }  is the same, and counts that the module went on
 *                          after the block was parsed (native_went_on)
 *
 * and one that declares a variable twice, as the parameters of a signature
 * are declared:
 *
 *   twice_z;  declares $z two times with PL_parser->in_my set to
 *             KEY_sigvar, which is what the interpreter names the
 *             declaration by when it warns that the second masks the first
 */
static Perl_keyword_plugin_t perlapi_next_keyword_plugin;

static int perlapi_keyword_plugin(pTHX_ char *kw, STRLEN kwlen, OP **op_ptr) {
    if (kwlen == 10 && strnEQ(kw, "one_into_x", 10)) {
        PADOFFSET padix = pad_add_name_pvn("$x", 2, 0, NULL, NULL);
        OP *varop, *assignop;
        intro_my();
        varop = newOP(OP_PADSV, OPf_MOD | OPf_REF | (OPpLVAL_INTRO << 8));
        varop->op_targ = padix;
        assignop = newASSIGNOP(OPf_WANT_VOID, varop, 0, newSVOP(OP_CONST, 0, newSViv(1)));
        *op_ptr = assignop;
        return KEYWORD_PLUGIN_STMT;
    }
    if (kwlen == 12 && strnEQ(kw, "hello_into_y", 12)) {
        PADOFFSET padix = pad_add_name_pvn("$y", 2, 0, NULL, NULL);
        intro_my();
        sv_setpvs(PAD_SVl(padix), "Hello");
        *op_ptr = newOP(OP_NULL, 0);
        return KEYWORD_PLUGIN_STMT;
    }
    if (kwlen == 8 && strnEQ(kw, "in_block", 8)) {
        lex_read_space(0);
        *op_ptr = parse_block(0);
        return KEYWORD_PLUGIN_STMT;
    }
    if (kwlen == 13 && strnEQ(kw, "counted_block", 13)) {
        lex_read_space(0);
        *op_ptr = parse_block(0);
        native_went_on_count++;
        return KEYWORD_PLUGIN_STMT;
    }
    if (kwlen == 10 && strnEQ(kw, "throw_hash", 10)) {
        /* a plugin that dies with an object */
        HV *const hv = newHV();
        (void)hv_stores(hv, "code", newSViv(9));
        croak_sv(sv_2mortal(newRV_noinc((SV *)hv)));
    }
    if (kwlen == 10 && strnEQ(kw, "throw_wide", 10)) {
        /* and one that dies with a message that has a name in UTF-8 in it */
        SV *const name = sv_2mortal(newSVpvs("p\xc5\x95\xc3\xb6\xc3\xa7"));
        SvUTF8_on(name);
        croak("In %" SVf ": refused", SVfARG(name));
    }
    if (kwlen == 7 && strnEQ(kw, "twice_z", 7)) {
        ENTER;
        SAVEI16(PL_parser->in_my);
        PL_parser->in_my = KEY_sigvar;
        pad_add_name_pvn("$z", 2, 0, NULL, NULL);
        pad_add_name_pvn("$z", 2, 0, NULL, NULL);
        LEAVE;
        intro_my();
        *op_ptr = newOP(OP_NULL, 0);
        return KEYWORD_PLUGIN_STMT;
    }
    return (*perlapi_next_keyword_plugin)(aTHX_ kw, kwlen, op_ptr);
}

XS(XS_PerlAPI_lexical_keywords) {
    dXSARGS;
    PERL_UNUSED_VAR(items);
    if (!perlapi_next_keyword_plugin)
        wrap_keyword_plugin(perlapi_keyword_plugin, &perlapi_next_keyword_plugin);
    XSRETURN_EMPTY;
}

/* Sets a signal aside the way modules do around a call that must not be
 * disturbed by it (IO::Tty around grantpt): the handler found there — the
 * pointer alone — is kept, the default action is set, and the handler is
 * put back with the flags the module has for handlers of its own. With
 * "sigaction" it is done through sigaction, with "signal" through signal.
 * Returns whether a handler was found. */
XS(XS_PerlAPI_set_signal_aside) {
    dXSARGS;
    if (items != 2) croak_xs_usage(cv, "how, sig");
    {
        const char *const how = SvPV_nolen(ST(0));
        const int sig = (int)SvIV(ST(1));
        void (*found)(int);
        if (strEQ(how, "sigaction")) {
            struct sigaction sa, osa;
            if (sigaction(sig, NULL, &osa) == -1) croak("sigaction: %s", strerror(errno));
            found = osa.sa_handler;
            memset(&sa, 0, sizeof(sa));
            sigemptyset(&sa.sa_mask);
            sa.sa_flags = 0;
            sa.sa_handler = SIG_DFL;
            if (sigaction(sig, &sa, NULL) == -1) croak("sigaction: %s", strerror(errno));
            sa.sa_handler = found;
            if (sigaction(sig, &sa, NULL) == -1) croak("sigaction: %s", strerror(errno));
        } else {
            found = signal(sig, SIG_DFL);
            if (found == SIG_ERR) croak("signal: %s", strerror(errno));
            if (signal(sig, found) == SIG_ERR) croak("signal: %s", strerror(errno));
        }
        ST(0) = boolSV(found != SIG_DFL && found != SIG_IGN);
        XSRETURN(1);
    }
}

/* ---- the compiler's structures (goperl_compiler.h) ---------------------- */

/* status_saved(n): $? is n until the scope that called this ends, as a
 * save of an interpreter variable made by native code is the interpreter's
 * to undo (no ENTER/LEAVE of the XSUB's own). */
XS(XS_PerlAPI_status_saved) {
    dXSARGS;
    if (items != 1) croak_xs_usage(cv, "n");
    SAVEI32(PL_statusvalue);
    PL_statusvalue = (I32)SvIV(ST(0));
    XSRETURN_EMPTY;
}

/* status_left(n): $? is n while a scope of the XSUB's own is open, and what
 * it was afterwards (LEAVE_SCOPE undoes the saves above an index). */
XS(XS_PerlAPI_status_left) {
    dXSARGS;
    if (items != 1) croak_xs_usage(cv, "n");
    {
        const I32 base = PL_savestack_ix;
        IV inside;
        SAVEI32(PL_statusvalue);
        PL_statusvalue = (I32)SvIV(ST(0));
        inside = (IV)PL_statusvalue;
        LEAVE_SCOPE(base);
        ST(0) = sv_2mortal(newSViv(inside * 1000 + (IV)PL_statusvalue));
        XSRETURN(1);
    }
}

/* pad_first(): the value at index 1 of the values of the pad in use. */
XS(XS_PerlAPI_pad_first) {
    dXSARGS;
    if (items != 0) croak_xs_usage(cv, "");
    ST(0) = PAD_SVl(1) ? PAD_SVl(1) : &PL_sv_undef;
    XSRETURN(1);
}

/* own_pad(sub, xsub): what index 1 of PL_curpad holds once it is given
 * values in this module's own memory, as Object::Pad's constructor gives
 * its field hooks a pad of their own: after an operation of the
 * interpreter's, inside a sub the interpreter runs (whose pad is its own),
 * inside an XSUB it calls (which has the pad of its caller), after both,
 * and after the scope ends. */
XS(XS_PerlAPI_own_pad) {
    dXSARGS;
    if (items != 2) croak_xs_usage(cv, "sub, xsub");
    {
        SV *const sub = ST(0), *const xsub = ST(1);
        SV *const out = sv_2mortal(newSVpvs(""));
        SV *const mine = sv_2mortal(newSVpvs("mine"));
        SV *got;
        int n;
        ENTER;
        SAVETMPS; /* the subs it calls free only temporaries made after this */
        SAVEVPTR(PL_curpad);
        Newx(PL_curpad, 3, SV *);
        SAVEFREEPV(PL_curpad);
        PAD_SVl(0) = 0;
        PAD_SVl(1) = mine;
        PAD_SVl(2) = 0;
        sv_catsv(out, PAD_SVl(1)); /* a temporary would give its string away to sv_setsv */
        {
            dSP;
            PUSHMARK(SP);
            PUTBACK;
            n = call_sv(sub, G_SCALAR);
            SPAGAIN;
            got = n ? POPs : &PL_sv_undef;
            sv_catpvf(out, ",%" SVf, SVfARG(got));
            PUSHMARK(SP);
            PUTBACK;
            n = call_sv(xsub, G_SCALAR);
            SPAGAIN;
            got = n ? POPs : &PL_sv_undef;
            sv_catpvf(out, ",%" SVf, SVfARG(got));
            PUTBACK;
        }
        sv_catpvf(out, ",%" SVf, SVfARG(PAD_SVl(1)));
        FREETMPS;
        LEAVE;
        sv_catpvf(out, ",%s", PL_curpad && PAD_SVl(1) == mine ? "mine" : "theirs");
        ST(0) = out;
        XSRETURN(1);
    }
}

/* names_of_a_sub_that_goes(code): the number of named entries of the pad
 * of the sub the code makes, which is freed afterwards, while the module
 * goes on with operations that take the memory the sub had (strings of
 * the size of its structures, holding addresses beyond the interpreter's
 * memory). */
XS(XS_PerlAPI_names_of_a_sub_that_goes) {
    dXSARGS;
    if (items != 1) croak_xs_usage(cv, "code");
    {
        IV named = 0, i;
        static const char fill[] =
            "\002\000\000\000\360\360\360\360\002\000\000\000\002\000\000\000"
            "\360\360\360\360\360\360\360\360\360\360\360\360\360\360\360\360"
            "\360\360\360\360\360\360\360\360";
        ENTER;
        SAVETMPS;
        {
            CV *const sub = (CV *)SvRV(eval_pv(SvPV_nolen(ST(0)), TRUE));
            PADNAMELIST *const pnl = PadlistNAMES(CvPADLIST(sub));
            for (i = 1; i <= PadnamelistMAX(pnl); i++) {
                PADNAME *const pn = PadnamelistARRAY(pnl)[i];
                if (pn && PadnameLEN(pn)) named++;
            }
        }
        FREETMPS;
        LEAVE;
        {
            AV *const keep = (AV *)sv_2mortal((SV *)newAV());
            for (i = 0; i < 4000; i++) {
                SV *const sv = newSVpvn(fill, 20 + (size_t)(i % 5) * 4);
                SvREFCNT_dec(sv);
                if (i % 2) av_push(keep, newSVpvn(fill, 20 + (size_t)(i % 7) * 4));
            }
        }
        ST(0) = sv_2mortal(newSViv(named));
        XSRETURN(1);
    }
}

/* pad_names(code): the names of the sub's pad, in order, with the COP
 * sequence range of each named one; with a second argument, the first
 * name that has that one is renamed (a new pad name of the interpreter's
 * takes its place in the list, as Object::Pad unshares a captured field)
 * and its range is set to [1, 2^31]. */
XS(XS_PerlAPI_pad_names) {
    dXSARGS;
    if (items < 1 || items > 3) croak_xs_usage(cv, "code [, from, to]");
    {
        CV *code = (CV *)SvRV(ST(0));
        PADLIST *pl = CvPADLIST(code);
        PADNAMELIST *pnl = PadlistNAMES(pl);
        PADNAME **names = PadnamelistARRAY(pnl);
        SV *out = sv_2mortal(newSVpvs(""));
        SSize_t i;
        if (items == 3) {
            STRLEN flen, tlen;
            const char *from = SvPV(ST(1), flen);
            const char *to = SvPV(ST(2), tlen);
            for (i = 1; i <= PadnamelistMAX(pnl); i++) {
                PADNAME *pn = names[i];
                if (!pn || PadnameLEN(pn) != flen || memNE(PadnamePV(pn), from, flen))
                    continue;
                {
                    PADNAME *renamed = newPADNAMEpvn(to, tlen);
                    PadnameREFCNT_dec(names[i]);
                    names[i] = renamed;
                    COP_SEQ_RANGE_LOW(renamed) = 1;
                    COP_SEQ_RANGE_HIGH(renamed) = 0x80000000u;
                }
                break;
            }
        }
        for (i = 1; i <= PadnamelistMAX(PadlistNAMES(CvPADLIST(code))); i++) {
            PADNAME *pn = PadnamelistARRAY(PadlistNAMES(CvPADLIST(code)))[i];
            if (!pn || !PadnameLEN(pn)) continue;
            sv_catpvf(out, "%s%.*s[%u,%u]", SvCUR(out) ? " " : "", (int)PadnameLEN(pn),
                      PadnamePV(pn), (unsigned)COP_SEQ_RANGE_LOW(pn),
                      (unsigned)COP_SEQ_RANGE_HIGH(pn));
        }
        ST(0) = out;
        XSRETURN(1);
    }
}

/* sub_fields(code [, file]): the outer sequence number, whether the sub is
 * marked lexical, and its file; with a file the sub is given that file
 * (CvFILE is assigned) and the lexical mark, and its outer sequence number
 * goes up by one. */
XS(XS_PerlAPI_sub_fields) {
    dXSARGS;
    if (items < 1 || items > 2) croak_xs_usage(cv, "code [, file]");
    {
        CV *code = (CV *)SvRV(ST(0));
        if (items == 2) {
            CvFILE(code) = savepv(SvPV_nolen(ST(1)));
            CvFLAGS(code) |= CVf_LEXICAL;
            CvOUTSIDE_SEQ(code) = CvOUTSIDE_SEQ(code) + 1;
        }
        ST(0) = sv_2mortal(newSVpvf("%u:%s:%s", (unsigned)CvOUTSIDE_SEQ(code),
                                    CvLEXICAL(code) ? "lexical" : "-", CvFILE(code)));
        XSRETURN(1);
    }
}

/* copy_sub(code): a new sub with the ops, the pad names and the outer
 * scope of the sub, made the way cv_copy_flags of Object::Pad and
 * Future::AsyncAwait makes one: the name of a sub that has one of its own
 * goes into the body (SvANY's xcv_hek), a padlist of pad_new with PL_compcv
 * saved, its lexicals new, what it captured shared. */
XS(XS_PerlAPI_copy_sub) {
    dXSARGS;
    if (items != 1) croak_xs_usage(cv, "code");
    {
        CV *orig = (CV *)SvRV(ST(0));
        CV *copy = MUTABLE_CV(newSV_type(SVt_PVCV));
        PADNAMELIST *padnames;
        SV **origpad;
        SSize_t fpad, padix;
        CvFLAGS(copy) = CvFLAGS(orig) & ~CVf_CVGV_RC;
        CvFILE(copy) = CvDYNFILE(orig) ? savepv(CvFILE(orig)) : CvFILE(orig);
        if (CvNAMED(orig)) {
            ((XPVCV *)MUTABLE_PTR(SvANY(copy)))->xcv_gv_u.xcv_hek =
                share_hek_hek(CvNAME_HEK(orig));
            CvNAMED_on(copy);
        } else {
            CvGV_set(copy, CvGV(orig));
        }
        CvSTASH_set(copy, CvSTASH(orig));
        OP_REFCNT_LOCK;
        CvROOT(copy) = OpREFCNT_inc(CvROOT(orig));
        OP_REFCNT_UNLOCK;
        CvSTART(copy) = CvSTART(orig);
        CvOUTSIDE(copy) = MUTABLE_CV(SvREFCNT_inc(CvOUTSIDE(orig)));
        CvOUTSIDE_SEQ(copy) = CvOUTSIDE_SEQ(orig);
        ENTER;
        SAVESPTR(PL_compcv);
        PL_compcv = copy;
        SAVESPTR(PL_comppad_name);
        PL_comppad_name = PadlistNAMES(CvPADLIST(orig));
        CvPADLIST_set(copy, pad_new(padnew_CLONE | padnew_SAVE));
        CvPADLIST(copy)->xpadl_id = CvPADLIST(orig)->xpadl_id;
        padnames = PadlistNAMES(CvPADLIST(orig));
        fpad = AvFILLp(PadlistARRAY(CvPADLIST(orig))[1]);
        origpad = AvARRAY(PadlistARRAY(CvPADLIST(orig))[1]);
        av_fill(PL_comppad, fpad);
        PL_curpad = AvARRAY(PL_comppad);
        for (padix = 1; padix <= fpad; padix++) {
            PADNAME *pn = padix <= PadnamelistMAX(padnames) ? PadnamelistARRAY(padnames)[padix] : NULL;
            const bool lexical = pn && PadnameLEN(pn) && !PadnameOUTER(pn) &&
                                 PadnamePV(pn)[0] != '&';
            if (lexical && PadnamePV(pn)[0] == '@')
                PL_curpad[padix] = MUTABLE_SV(newAV());
            else if (lexical && PadnamePV(pn)[0] == '%')
                PL_curpad[padix] = MUTABLE_SV(newHV());
            else if (lexical)
                PL_curpad[padix] = newSV(0);
            else if (!origpad[padix])
                PL_curpad[padix] = NULL;
            else if (SvPADTMP(origpad[padix])) {
                PL_curpad[padix] = newSVsv(origpad[padix]);
                SvPADTMP_on(PL_curpad[padix]);
            } else
                PL_curpad[padix] = SvREFCNT_inc(origpad[padix]);
        }
        LEAVE;
        ST(0) = sv_2mortal(newRV_noinc((SV *)copy));
        XSRETURN(1);
    }
}

/* hint(key): the value of the key in the hints of the calling statement
 * (%^H at its compilation), "absent" when it is not there; and whether
 * that hash says it exists. */
XS(XS_PerlAPI_hint) {
    dXSARGS;
    if (items != 1) croak_xs_usage(cv, "key");
    {
        STRLEN len;
        const char *key = SvPV(ST(0), len);
        COPHH *hh = CopHINTHASH_get(PL_curcop);
        SV *val = cophh_fetch_pvn(hh, key, len, 0, 0);
        const bool exists = cophh_exists_pvn(hh, key, len, 0, 0);
        SV *out;
        if (val == &PL_sv_placeholder) out = newSVpvs("absent");
        else out = newSVsv(val);
        sv_catpv(out, exists ? ":exists" : ":-");
        ST(0) = sv_2mortal(out);
        XSRETURN(1);
    }
}

/* run_list(values...): the values, as a list op made of constants returns
 * them when its ops are run on the stack of this XSUB (exec_optree_list of
 * Object::Pad); then the number of values left above the mark. */
XS(XS_PerlAPI_run_list) {
    dXSARGS;
    {
        OP *list = newLISTOP(OP_LIST, 0, NULL, NULL);
        AV *got;
        I32 nargs, i;
        for (i = 0; i < items; i++)
            list = op_append_elem(OP_LIST, list, newSVOP(OP_CONST, 0, newSVsv(ST(i))));
        list->op_flags |= OPf_WANT_LIST;
        ENTER;
        SAVETMPS;
        SAVEVPTR(PL_op);
        PL_op = LINKLIST(list);
        list->op_next = NULL;
        SP = PL_stack_base + ax - 1;
        PUSHMARK(SP);
        PUTBACK;
        CALLRUNOPS(aTHX);
        SPAGAIN;
        nargs = (I32)(SP - PL_stack_base - TOPMARK);
        got = newAV();
        for (i = 0; i < nargs; i++) av_push(got, newSVsv(SP[i - nargs + 1]));
        POPMARK;
        FREETMPS;
        LEAVE;
        op_free(list);
        SP = PL_stack_base + ax - 1;
        XPUSHs(sv_2mortal(newRV_noinc((SV *)got)));
        XPUSHs(sv_2mortal(newSViv(nargs)));
        PUTBACK;
        return;
    }
}

/* fields_object(class, values...): a new object of the class whose fields
 * are the values, made the way Object::Pad makes one of :repr(pvobj) (an
 * array of fields of native code's own, its number less one, zeroed and
 * filled); fields_of(object): its fields. */
XS(XS_PerlAPI_fields_object) {
    dXSARGS;
    if (items < 1) croak_xs_usage(cv, "class, values...");
    {
        HV *stash = gv_stashsv(ST(0), GV_ADD);
        SV *obj = newSV_type(SVt_PVOBJ);
        const U32 count = (U32)(items - 1);
        SV *self;
        U32 i;
        Newx(ObjectFIELDS(obj), count, SV *);
        ObjectMAXFIELD(obj) = (SSize_t)count - 1;
        Zero(ObjectFIELDS(obj), count, SV *);
        for (i = 0; i < count; i++) ObjectFIELDS(obj)[i] = newSVsv(ST(1 + i));
        self = sv_2mortal(newRV_noinc(obj));
        sv_bless(self, stash);
        ST(0) = self;
        XSRETURN(1);
    }
}
XS(XS_PerlAPI_fields_of) {
    dXSARGS;
    if (items != 1) croak_xs_usage(cv, "object");
    {
        SV *obj = SvRV(ST(0));
        SV *out = sv_2mortal(newSVpvs(""));
        SSize_t i;
        if (SvTYPE(obj) != SVt_PVOBJ) croak("not an object of fields");
        for (i = 0; i <= ObjectMAXFIELD(obj); i++)
            sv_catpvf(out, "%s%" SVf, i ? "," : "", SVfARG(ObjectFIELDS(obj)[i]));
        ST(0) = out;
        XSRETURN(1);
    }
}

/* statement_line(): the line of a statement op made, with a parser made
 * for it the way Object::Pad makes one when there is none (newSTATEOP needs
 * one); whether there was a parser before, and whether there is one after
 * the scope that made it. */
XS(XS_PerlAPI_statement_line) {
    dXSARGS;
    PERL_UNUSED_VAR(items);
    {
        OP *o;
        IV line;
        const bool had = PL_parser != NULL;
        ENTER;
        if (!PL_parser) {
            SAVEVPTR(PL_parser);
            Newxz(PL_parser, 1, yy_parser);
            SAVEFREEPV(PL_parser);
            PL_parser->copline = NOLINE;
            PL_parser->preambling = NOLINE;
        }
        o = newSTATEOP(0, NULL, NULL);
        line = (IV)CopLINE((COP *)o);
        op_free(o);
        LEAVE;
        EXTEND(SP, 3);
        ST(0) = sv_2mortal(newSViv(line));
        ST(1) = had ? &PL_sv_yes : &PL_sv_no;
        ST(2) = PL_parser ? &PL_sv_yes : &PL_sv_no;
        XSRETURN(3);
    }
}

/* stash_hek(object): croaks with the name of the object's class, in a
 * message made with HEKf. */
XS(XS_PerlAPI_stash_hek) {
    dXSARGS;
    if (items != 1) croak_xs_usage(cv, "object");
    croak("class %" HEKf " says hello", HEKfARG(HvNAME_HEK(SvSTASH(SvRV(ST(0))))));
}

/* buckets(hash): "none" for a hash without buckets (pp_defined's test),
 * else the number of entries the chains of its buckets hold and their keys
 * in order; PerlIO_debug is called on the way, which the interpreter's
 * build answers as perl's does without -Di (nothing). */
XS(XS_PerlAPI_buckets) {
    dXSARGS;
    if (items != 1) croak_xs_usage(cv, "hashref");
    {
        HV *hv = (HV *)SvRV(ST(0));
        HE **buckets = HvARRAY(hv);
        AV *keys;
        SV *out;
        STRLEN i;
        I32 n = 0;
        PerlIO_debug("buckets of %p\n", (void *)hv);
        if (!buckets) {
            ST(0) = sv_2mortal(newSVpvs("none"));
            XSRETURN(1);
        }
        keys = (AV *)sv_2mortal((SV *)newAV());
        for (i = 0; i <= HvMAX(hv); i++) {
            HE *he;
            for (he = buckets[i]; he; he = HeNEXT(he)) {
                av_push(keys, newSVsv(HeSVKEY_force(he)));
                n++;
            }
        }
        sortsv(AvARRAY(keys), av_count(keys), Perl_sv_cmp);
        out = newSVpvf("%d", (int)n);
        for (i = 0; i < (STRLEN)av_count(keys); i++)
            sv_catpvf(out, ":%" SVf, SVfARG(*av_fetch(keys, (SSize_t)i, 0)));
        ST(0) = sv_2mortal(out);
        XSRETURN(1);
    }
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
    newXS_deffile("PerlAPI::live_values", XS_PerlAPI_live_values);
    newXS_deffile("PerlAPI::die_in_operation", XS_PerlAPI_die_in_operation);
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
    newXS_deffile("PerlAPI::other_names", XS_PerlAPI_other_names);
    newXS_deffile("PerlAPI::scan_a_version", XS_PerlAPI_scan_a_version);
    newXS_deffile("PerlAPI::generated_globs", XS_PerlAPI_generated_globs);
    newXS_deffile("PerlAPI::random_numbers", XS_PerlAPI_random_numbers);
    newXS_deffile("PerlAPI::pp_on_own_op", XS_PerlAPI_pp_on_own_op);
    newXS_deffile("PerlAPI::calling_op", XS_PerlAPI_calling_op);
    newXS_deffile("PerlAPI::filled_by_slots", XS_PerlAPI_filled_by_slots);
    newXS_deffile("PerlAPI::lvalue_fields", XS_PerlAPI_lvalue_fields);
    newXS_deffile("PerlAPI::accessor_loop", XS_PerlAPI_accessor_loop);
    newXS_deffile("PerlAPI::array_loop", XS_PerlAPI_array_loop);
    newXS_deffile("PerlAPI::keep_pair", XS_PerlAPI_keep_pair);
    newXS_deffile("PerlAPI::kept_pair", XS_PerlAPI_kept_pair);
    newXS_deffile("PerlAPI::release_pair", XS_PerlAPI_release_pair);
    newXS_deffile("PerlAPI::kept_array", XS_PerlAPI_kept_array);
    newXS_deffile("PerlAPI::kept_array_sums", XS_PerlAPI_kept_array_sums);
    newXS_deffile("PerlAPI::kept_hash", XS_PerlAPI_kept_hash);
    newXS_deffile("PerlAPI::kept_hash_sums", XS_PerlAPI_kept_hash_sums);
    newXS_deffile("PerlAPI::wide_values", XS_PerlAPI_wide_values);
    newXS_deffile("PerlAPI::wide_made", XS_PerlAPI_wide_made);
    newXS_deffile("PerlAPI::read_host_descriptor", XS_PerlAPI_read_host_descriptor);
    newXS_deffile("PerlAPI::integer_limits", XS_PerlAPI_integer_limits);
    newXS_deffile("PerlAPI::copy_integer", XS_PerlAPI_copy_integer);
    newXS_deffile("PerlAPI::sealed_new", XS_PerlAPI_sealed_new);
    newXS_deffile("PerlAPI::sealed_add", XS_PerlAPI_sealed_add);
    newXS_deffile("PerlAPI::read_then_mark", XS_PerlAPI_read_then_mark);
    newXS_deffile("PerlAPI::croak_at_caller", XS_PerlAPI_croak_at_caller);
    newXS_deffile("PerlAPI::try_then_rethrow", XS_PerlAPI_try_then_rethrow);
    newXS_deffile("PerlAPI::native_went_on", XS_PerlAPI_native_went_on);
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
    newXS_deffile("PerlAPI::derived_from_stash", XS_PerlAPI_derived_from_stash);
    newXS_deffile("PerlAPI::with_topic", XS_PerlAPI_with_topic);
    newXS_deffile("PerlAPI::leave_errno", XS_PerlAPI_leave_errno);
    newXS_deffile("PerlAPI::errno_of_the_program", XS_PerlAPI_errno_of_the_program);
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
    newXS_deffile("PerlAPI::pin_attach", XS_PerlAPI_pin_attach);
    newXS_deffile("PerlAPI::magic_refcounted", XS_PerlAPI_magic_refcounted);
    newXS_deffile("PerlAPI::rv_slot_after_rok_off", XS_PerlAPI_rv_slot_after_rok_off);
    newXS_deffile("PerlAPI::process_id", XS_PerlAPI_process_id);
    newXS_deffile("PerlAPI::discard_string", XS_PerlAPI_discard_string);
    newXS_deffile("PerlAPI::resize_buffer", XS_PerlAPI_resize_buffer);
    newXS_deffile("PerlAPI::insecure", XS_PerlAPI_insecure);
    newXS_deffile("PerlAPI::with_local_status", XS_PerlAPI_with_local_status);
    newXS_deffile("PerlAPI::save_one_int", XS_PerlAPI_save_one_int);
    newXS_deffile("PerlAPI::open_and_read", XS_PerlAPI_open_and_read);
    newXS_deffile("PerlAPI::lio_peek", XS_PerlAPI_lio_peek);
    newXS_deffile("PerlAPI::count_own_struct", XS_PerlAPI_count_own_struct);
    newXS_deffile("PerlAPI::null_strings", XS_PerlAPI_null_strings);
    newXS_deffile("PerlAPI::anonymous_subs", XS_PerlAPI_anonymous_subs);
    newXS_deffile("PerlAPI::array_slots", XS_PerlAPI_array_slots);
    newXS_deffile("PerlAPI::string_form", XS_PerlAPI_string_form);
    newXS_deffile("PerlAPI::format_dropping_result", XS_PerlAPI_format_dropping_result);
    newXS_deffile("PerlAPI::watch", XS_PerlAPI_watch);
    newXS_deffile("PerlAPI::watched", XS_PerlAPI_watched);
    newXS_deffile("PerlAPI::keyed_attach", XS_PerlAPI_keyed_attach);
    newXS_deffile("PerlAPI::keyed_read", XS_PerlAPI_keyed_read);
    newXS_deffile("PerlAPI::tied_to", XS_PerlAPI_tied_to);
    newXS_deffile("PerlAPI::decode_first", XS_PerlAPI_decode_first);
    newXS_deffile("PerlAPI::pio_layers", XS_PerlAPI_pio_layers);
    newXS_deffile("PerlAPI::pio_failures", XS_PerlAPI_pio_failures);
    newXS_deffile("PerlAPI::impl", XS_PerlAPI_impl_one);
    newXS_deffile("PerlAPI::exchange", XS_PerlAPI_exchange);
    newXS_deffile("PerlAPI::glob_fields", XS_PerlAPI_glob_fields);
    newXS_deffile("PerlAPI::clone_chain", XS_PerlAPI_clone_chain);
    newXS_deffile("PerlAPI::repoint", XS_PerlAPI_repoint);
    newXS_deffile("PerlAPI::kept_addresses", XS_PerlAPI_kept_addresses);
    newXS_deffile("PerlAPI::glob_slots", XS_PerlAPI_glob_slots);
    newXS_deffile("PerlAPI::hooks", XS_PerlAPI_hooks);
    newXS_deffile("PerlAPI::at_exit_write", XS_PerlAPI_at_exit_write);
    newXS_deffile("PerlAPI::rx_info", XS_PerlAPI_rx_info);
    newXS_deffile("PerlAPI::rx_compile", XS_PerlAPI_rx_compile);
    newXS_deffile("PerlAPI::own_magic", XS_PerlAPI_own_magic);
    newXS_deffile("PerlAPI::constant_calls", XS_PerlAPI_constant_calls);
    newXS_deffile("PerlAPI::croak_guarded", XS_PerlAPI_croak_guarded);
    newXS_deffile("PerlAPI::destructor_found", XS_PerlAPI_destructor_found);
    newXS_deffile("PerlAPI::show", XS_PerlAPI_show);
    newXS_deffile("PerlAPI::observe", XS_PerlAPI_observe);
    newXS_deffile("PerlAPI::observed", XS_PerlAPI_observed);
    newXS_deffile("PerlAPI::uvar_hash", XS_PerlAPI_uvar_hash);
    newXS_deffile("PerlAPI::uvar_calls", XS_PerlAPI_uvar_calls);
    newXS_deffile("PerlAPI::firsts", XS_PerlAPI_firsts);
    newXS_deffile("PerlAPI::take_buffer", XS_PerlAPI_take_buffer);
    newXS_deffile("PerlAPI::const_ops", XS_PerlAPI_const_ops);
    newXS_deffile("PerlAPI::spin", XS_PerlAPI_spin);
    newXS_deffile("PerlAPI::by_number", XS_PerlAPI_by_number);
    newXS_deffile("PerlAPI::own_name", XS_PerlAPI_own_name);
    newXS_deffile("PerlAPI::signature", XS_PerlAPI_signature);
    newXS_deffile("PerlAPI::first_index", XS_PerlAPI_first_index);
    newXS_deffile("PerlAPI::largest", XS_PerlAPI_largest);
    newXS_deffile("PerlAPI::each_pair", XS_PerlAPI_each_pair);
    newXS_deffile("PerlAPI::category_warning", XS_PerlAPI_category_warning);
    newXS_deffile("PerlAPI::evaluated_twice", XS_PerlAPI_evaluated_twice);
    newXS_deffile("PerlAPI::encode_uv", XS_PerlAPI_encode_uv);
    newXS_deffile("PerlAPI::rename_glob", XS_PerlAPI_rename_glob);
    newXS_deffile("PerlAPI::stack_in_a_plain_function", XS_PerlAPI_stack_in_a_plain_function);
    newXS_deffile("PerlAPI::look_without_magic", XS_PerlAPI_look_without_magic);
    {
        CV *sum = newXS_deffile("PerlAPI::sum_without_dxsargs", XS_PerlAPI_sum_without_dxsargs);
        CvXSUBANY(sum).any_i32 = 100;
    }
    newXS_deffile("PerlAPI::lexical_keywords", XS_PerlAPI_lexical_keywords);
    newXS_deffile("PerlAPI::set_signal_aside", XS_PerlAPI_set_signal_aside);
    newXS_deffile("PerlAPI::fd_of", XS_PerlAPI_fd_of);
    newXS_deffile("PerlAPI::fd_state", XS_PerlAPI_fd_state);
    newXS_deffile("PerlAPI::fd_write", XS_PerlAPI_fd_write);
    newXS_deffile("PerlAPI::fd_read", XS_PerlAPI_fd_read);
    newXS_deffile("PerlAPI::fd_pair", XS_PerlAPI_fd_pair);
    newXS_deffile("PerlAPI::fd_close", XS_PerlAPI_fd_close);
    newXS_deffile("PerlAPI::fd_line", XS_PerlAPI_fd_line);
    newXS_deffile("PerlAPI::fd_import_write", XS_PerlAPI_fd_import_write);
    newXS_deffile("PerlAPI::status_saved", XS_PerlAPI_status_saved);
    newXS_deffile("PerlAPI::status_left", XS_PerlAPI_status_left);
    newXS_deffile("PerlAPI::pad_names", XS_PerlAPI_pad_names);
    newXS_deffile("PerlAPI::names_of_a_sub_that_goes", XS_PerlAPI_names_of_a_sub_that_goes);
    newXS_deffile("PerlAPI::pad_first", XS_PerlAPI_pad_first);
    newXS_deffile("PerlAPI::own_pad", XS_PerlAPI_own_pad);
    newXS_deffile("PerlAPI::sub_fields", XS_PerlAPI_sub_fields);
    newXS_deffile("PerlAPI::copy_sub", XS_PerlAPI_copy_sub);
    newXS_deffile("PerlAPI::hint", XS_PerlAPI_hint);
    newXS_deffile("PerlAPI::run_list", XS_PerlAPI_run_list);
    newXS_deffile("PerlAPI::fields_object", XS_PerlAPI_fields_object);
    newXS_deffile("PerlAPI::fields_of", XS_PerlAPI_fields_of);
    newXS_deffile("PerlAPI::statement_line", XS_PerlAPI_statement_line);
    newXS_deffile("PerlAPI::stash_hek", XS_PerlAPI_stash_hek);
    newXS_deffile("PerlAPI::buckets", XS_PerlAPI_buckets);
    Perl_xs_boot_epilog(aTHX_ ax);
}
