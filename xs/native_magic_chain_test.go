//go:build darwin || linux

package xs_test

import "testing"

// magicWalkModule is a module that copies a value the way Clone does: it
// walks the value's chain of magic (SvMAGIC) and attaches what it finds to
// the copy with sv_magic, the tie object copied too; copies the elements of
// arrays and hashes that are not tied, and the referent of a reference. It
// also lists the types on a chain, links a MAGIC of its own onto the end
// of a chain, as Event does, or before its head, and unlinks the head.
const magicWalkModule = `#include "EXTERN.h"
#include "perl.h"
#include "XSUB.h"
static SV *walk_clone(SV *ref) {
    SV *clone;
    MAGIC *mg;
    int tied = 0;
    switch (SvTYPE(ref)) {
    case SVt_PVAV: clone = (SV *)newAV(); break;
    case SVt_PVHV: clone = (SV *)newHV(); break;
    default: clone = newSVsv(ref); break;
    }
    if (SvMAGICAL(ref)) {
        for (mg = SvMAGIC(ref); mg; mg = mg->mg_moremagic) {
            SV *obj = NULL;
            char *ptr = mg->mg_ptr;
            if (mg->mg_type == 'P' || mg->mg_type == 'q') {
                obj = walk_clone(mg->mg_obj);
                tied = 1;
            } else if (mg->mg_type == 'w' && mg->mg_len == -1) {
                STRLEN *cache;
                if (!ptr) continue;
                /* the magic owns the cache it is given */
                Newxz(cache, PERL_MAGIC_UTF8_CACHESIZE * 2, STRLEN);
                Copy(mg->mg_ptr, cache, PERL_MAGIC_UTF8_CACHESIZE * 2, STRLEN);
                ptr = (char *)cache;
            } else {
                continue;
            }
            sv_magic(clone, obj, mg->mg_type, ptr, mg->mg_len);
            if (obj) SvREFCNT_dec(obj);
        }
    }
    if (tied) return clone;
    if (SvTYPE(ref) == SVt_PVAV) {
        I32 n = av_len((AV *)ref);
        for (I32 i = 0; i <= n; i++) {
            SV **svp = av_fetch((AV *)ref, i, 0);
            if (svp) av_store((AV *)clone, i, walk_clone(*svp));
        }
    } else if (SvTYPE(ref) == SVt_PVHV) {
        HE *he;
        hv_iterinit((HV *)ref);
        while ((he = hv_iternext((HV *)ref)))
            hv_store_ent((HV *)clone, hv_iterkeysv(he), walk_clone(hv_iterval((HV *)ref, he)), 0);
    } else if (SvROK(ref)) {
        SvREFCNT_dec(SvRV(clone));
        SvRV(clone) = walk_clone(SvRV(ref));
        if (SvOBJECT(SvRV(ref))) sv_bless(clone, SvSTASH(SvRV(ref)));
    }
    return clone;
}
XS(XS_MagicWalk_clone) {
    dXSARGS;
    if (items != 1) croak_xs_usage(cv, "value");
    ST(0) = sv_2mortal(walk_clone(ST(0)));
    XSRETURN(1);
}
XS(XS_MagicWalk_types) {
    dXSARGS;
    if (items != 1) croak_xs_usage(cv, "ref");
    {
        SV *out = sv_2mortal(newSVpvs(""));
        SV *sv = SvRV(ST(0));
        MAGIC *mg;
        for (mg = SvMAGICAL(sv) ? SvMAGIC(sv) : NULL; mg; mg = mg->mg_moremagic)
            sv_catpvf(out, "%c", mg->mg_type);
        ST(0) = out;
        XSRETURN(1);
    }
}
static int walk_mark = 42;
XS(XS_MagicWalk_mark) {
    dXSARGS;
    if (items != 1) croak_xs_usage(cv, "ref");
    {
        SV *sv = SvRV(ST(0));
        MAGIC **mgp, *mg;
        mgp = &SvMAGIC(sv);
        while ((mg = *mgp)) mgp = &mg->mg_moremagic;
        Newxz(mg, 1, MAGIC);
        mg->mg_type = '~';
        mg->mg_ptr = (char *)&walk_mark;
        mg->mg_private = 0x6d6b;
        *mgp = mg;
        XSRETURN_EMPTY;
    }
}
XS(XS_MagicWalk_mark_head) {
    dXSARGS;
    if (items != 1) croak_xs_usage(cv, "ref");
    {
        SV *sv = SvRV(ST(0));
        MAGIC *mg;
        Newxz(mg, 1, MAGIC);
        mg->mg_type = '~';
        mg->mg_ptr = (char *)&walk_mark;
        mg->mg_private = 0x6d6b;
        mg->mg_moremagic = SvMAGIC(sv);
        SvMAGIC(sv) = mg;
        XSRETURN_EMPTY;
    }
}
XS(XS_MagicWalk_unlink_head) {
    dXSARGS;
    if (items != 1) croak_xs_usage(cv, "ref");
    {
        SV *sv = SvRV(ST(0));
        MAGIC *mg = SvMAGIC(sv);
        if (mg) SvMAGIC(sv) = mg->mg_moremagic;
        XSRETURN_EMPTY;
    }
}
XS(XS_MagicWalk_utf8_cache) {
    dXSARGS;
    if (items != 1) croak_xs_usage(cv, "ref");
    {
        MAGIC *mg = mg_find(SvRV(ST(0)), PERL_MAGIC_utf8);
        const STRLEN *c = mg ? (const STRLEN *)mg->mg_ptr : NULL;
        ST(0) = !c ? sv_2mortal(newSVpvs("none"))
                   : sv_2mortal(newSVpvf("%d:%d,%d,%d,%d", (int)mg->mg_len, (int)c[0],
                                         (int)c[1], (int)c[2], (int)c[3]));
        XSRETURN(1);
    }
}
XS(XS_MagicWalk_name_sv) {
    dXSARGS;
    if (items != 2) croak_xs_usage(cv, "ref, name");
    sv_magic(SvRV(ST(0)), NULL, PERL_MAGIC_ext, (const char *)ST(1), HEf_SVKEY);
    XSRETURN_EMPTY;
}
XS(XS_MagicWalk_name_of) {
    dXSARGS;
    if (items != 1) croak_xs_usage(cv, "ref");
    {
        MAGIC *mg = mg_find(SvRV(ST(0)), PERL_MAGIC_ext);
        ST(0) = mg && mg->mg_len == HEf_SVKEY ? (SV *)mg->mg_ptr : &PL_sv_undef;
        XSRETURN(1);
    }
}
XS(XS_MagicWalk_marked) {
    dXSARGS;
    if (items != 1) croak_xs_usage(cv, "ref");
    {
        MAGIC *mg = mg_find(SvRV(ST(0)), '~');
        ST(0) = sv_2mortal(newSViv(mg && mg->mg_private == 0x6d6b ? *(int *)mg->mg_ptr : -1));
        XSRETURN(1);
    }
}
XS_EXTERNAL(boot_MagicWalk) {
    dXSBOOTARGSXSAPIVERCHK;
    newXS_deffile("MagicWalk::clone", XS_MagicWalk_clone);
    newXS_deffile("MagicWalk::types", XS_MagicWalk_types);
    newXS_deffile("MagicWalk::mark", XS_MagicWalk_mark);
    newXS_deffile("MagicWalk::marked", XS_MagicWalk_marked);
    newXS_deffile("MagicWalk::mark_head", XS_MagicWalk_mark_head);
    newXS_deffile("MagicWalk::unlink_head", XS_MagicWalk_unlink_head);
    newXS_deffile("MagicWalk::utf8_cache", XS_MagicWalk_utf8_cache);
    newXS_deffile("MagicWalk::name_sv", XS_MagicWalk_name_sv);
    newXS_deffile("MagicWalk::name_of", XS_MagicWalk_name_of);
    Perl_xs_boot_epilog(aTHX_ ax);
}
`

// The chain of magic native code finds with SvMAGIC is the value's whole
// chain: the magic the interpreter keeps there (a tie, the UTF-8 cache of a
// string) as well as what native code attached. A module that copies a
// value by walking that chain, as Clone does, copies a tie (its object
// too) and the UTF-8 cache, and a MAGIC of the module's own linked onto
// the end of such a chain, or before its head, is the value's, while
// unlinking the interpreter's magic by hand, which gperl cannot do, fails
// saying so. Before, the chain showed only
// what native code had attached: the copy of a tied hash, array or scalar
// was an untied empty one, and its UTF-8 cache was lost (Clone's t/04-tie.t,
// t/07-magic.t, t/32-tied-mgobj-leak.t).
func TestNativeMagicChainShowsTheInterpretersMagic(t *testing.T) {
	so := buildModuleSource(t, "MagicWalk", magicWalkModule)
	ties := `require Tie::Hash; require Tie::Array; require Tie::Scalar;
		tie my %h, "Tie::StdHash"; %h = (a => 1, b => 2);
		tie my @a, "Tie::StdArray"; @a = (3, 4);
		tie my $s, "Tie::StdScalar"; $s = 5;`
	for _, tc := range []struct{ name, code, want string }{
		{"a tie on the chain", ties + ` join ",", MagicWalk::types(\%h), MagicWalk::types(\@a), MagicWalk::types(\$s)`, "P,P,q"},
		{"the copy of a tied hash", ties + ` my $c = MagicWalk::clone(\%h); join ",", ref($c), ref(tied(%$c)), map { "$_=$c->{$_}" } sort keys %$c`, "HASH,Tie::StdHash,a=1,b=2"},
		{"the copy is a tie of its own", ties + ` my $c = MagicWalk::clone(\%h); $c->{a} = 9; "$h{a} $c->{a}"`, "1 9"},
		{"tied values in an array", ties + ` my $c = MagicWalk::clone([\%h, \@a, \$s]); join ",", map(ref, @$c), $c->[0]{b}, $c->[1][1], ${$c->[2]}, ref(tied(@{$c->[1]})), ref(tied(${$c->[2]}))`, "HASH,ARRAY,SCALAR,2,4,5,Tie::StdArray,Tie::StdScalar"},
		{"the UTF-8 cache", `require B; my $v = "a\r\n"; utf8::upgrade($v); index($v, "\n"); my $c = MagicWalk::clone(\$v); my $m = B::svref_2object($c)->MAGIC; ($m ? $m->TYPE : "none") . " " . $$c`, "w a\r\n"},
		{"the positions in the UTF-8 cache", `my $v = "\x{263a}b\x{263a}cd" x 3; index($v, "c"); my $c = MagicWalk::clone(\$v); my ($o, $k) = (MagicWalk::utf8_cache(\$v), MagicWalk::utf8_cache($c)); ($o eq $k ? "same" : "$o vs $k") . " " . ($o =~ /^-1:[1-9]/ ? "cached" : $o) . " " . length($$c) . " " . ord(substr($$c, 7, 1))`, "same cached 15 9786"},
		{"a name that is an SV", `my $v = 1; MagicWalk::name_sv(\$v, "the name"); MagicWalk::name_of(\$v) // "none"`, "the name"},
		{"a MAGIC of the module's after the interpreter's", ties + ` MagicWalk::mark(\%h); MagicWalk::marked(\%h) . " " . (tied(%h) ? "still tied" : "untied") . " " . $h{b}`, "42 still tied 2"},
		{"a MAGIC of the module's before the interpreter's", ties + ` MagicWalk::mark_head(\%h); join ",", MagicWalk::marked(\%h), MagicWalk::types(\%h), (tied(%h) ? "still tied" : "untied"), $h{b}`, "42,P~,still tied,2"},
		{"the module's MAGIC on an untied value", `my %h = (a => 1); MagicWalk::mark_head(\%h); MagicWalk::mark(\%h); join ",", MagicWalk::marked(\%h), MagicWalk::types(\%h)`, "42,~~"},
		{"unlinking the module's own MAGIC", `my %h = (a => 1); MagicWalk::mark(\%h); MagicWalk::unlink_head(\%h); join ",", MagicWalk::marked(\%h), "[" . MagicWalk::types(\%h) . "]"`, "-1,[]"},
		{"unlinking the interpreter's magic", ties + ` eval { MagicWalk::unlink_head(\%h); 1 } ? "unlinked" : ($@ =~ /unlinked magic 'P' that the interpreter keeps/ ? "refused" : $@) . " " . (tied(%h) ? "still tied" : "untied")`, "refused still tied"},
	} {
		if got := requireModule(t, "MagicWalk", so, tc.code); got != tc.want {
			t.Errorf("%s: got %q, want %q", tc.name, got, tc.want)
		}
	}
}
