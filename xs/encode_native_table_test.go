//go:build darwin || linux

package xs_test

import "testing"

// An encoding of native code as enc2xs makes one (Encode::EUCJPASCII,
// Encode::HanExtra, Encode::JISX0213): the module's tables, in the host
// process, made an Encode::XS object. Encode works on a copy of the tables
// it makes on first use: decoding one and two bytes per character,
// encoding back, the replacement character for one the encoding lacks, a
// byte sequence it does not map under FB_CROAK, and the encoding's name
// (each died with "go-perl gives native code no access to this interpreter
// internal: an Encode::XS table of native code"). What is expected is what
// perl 5.36 on Linux gives with the same module.
func TestNativeEncodeTableOfAnEnc2xsModule(t *testing.T) {
	so := buildModuleSource(t, "Enctab", `#include "EXTERN.h"
#include "perl.h"
#include "XSUB.h"
/* An encoding as enc2xs makes a module of one (Encode::EUCJPASCII,
 * Encode::HanExtra): Encode's tables (encode.h), the tables enc2xs made of
 * a small character map (toy.ucm: one byte for ASCII and e-acute, two for
 * two kana), and enc2xs's Encode_XSEncoding, which makes the table an
 * Encode::XS object and defines the encoding with it. */
typedef struct encpage_s encpage_t;
struct encpage_s {
    const U8 *const seq;
    const encpage_t *const next;
    const U8 min;
    const U8 max;
    const U8 dlen;
    const U8 slen;
};
typedef struct encode_s encode_t;
struct encode_s {
    const encpage_t *const t_utf8;
    const encpage_t *const f_utf8;
    const U8 *const rep;
    int replen;
    U8 min_el;
    U8 max_el;
    const char *const name[2];
};
static const encpage_t toy_enc_utf8[];
static const encpage_t utf8_toy_enc[];

static const U8 enctable[18] = { 227,129,130,227,129,132,65,66,67,129,65,129,
66,195,169,0,63,128 };
static const encpage_t toy_enc_utf8_81[2] = {
{enctable + 0,toy_enc_utf8,0x41,0x42,3,1},
{0,toy_enc_utf8_81,0x43,0xff,0,0},
};
static const encpage_t toy_enc_utf8[6] = {
{enctable + 15,toy_enc_utf8,0x00,0x00,1,1},
{enctable + 16,toy_enc_utf8,0x3f,0x3f,1,1},
{enctable + 6,toy_enc_utf8,0x41,0x43,1,1},
{enctable + 13,toy_enc_utf8,0x80,0x80,2,1},
{0,toy_enc_utf8_81,0x81,0x81,0,2},
{0,toy_enc_utf8,0x82,0xff,0,0},
};
static const encpage_t utf8_toy_enc_c3[2] = {
{enctable + 17,utf8_toy_enc,0xa9,0xa9,1,1},
{0,utf8_toy_enc_c3,0xaa,0xff,0,0},
};
static const encpage_t utf8_toy_enc_e3_81[3] = {
{enctable + 9,utf8_toy_enc,0x82,0x82,2,1},
{enctable + 11,utf8_toy_enc,0x84,0x84,2,1},
{0,utf8_toy_enc_e3_81,0x85,0xff,0,0},
};
static const encpage_t utf8_toy_enc_e3[2] = {
{0,utf8_toy_enc_e3_81,0x81,0x81,0,2},
{0,utf8_toy_enc_e3,0x82,0xff,0,0},
};
static const encpage_t utf8_toy_enc[6] = {
{enctable + 15,utf8_toy_enc,0x00,0x00,1,1},
{enctable + 16,utf8_toy_enc,0x3f,0x3f,1,1},
{enctable + 6,utf8_toy_enc,0x41,0x43,1,1},
{0,utf8_toy_enc_c3,0xc3,0xc3,0,2},
{0,utf8_toy_enc_e3,0xe3,0xe3,0,3},
{0,utf8_toy_enc,0xe4,0xff,0,0},
};
static const U8 toy_enc_encoding_rep_character[] = "\x3F";
static const char toy_enc_encoding_enc_name[] = "toy-enc";

 const encode_t toy_enc_encoding = 
 {toy_enc_utf8,utf8_toy_enc,toy_enc_encoding_rep_character,1,1,2,{toy_enc_encoding_enc_name,(const char *)0}};

static void Encode_XSEncoding(pTHX_ encode_t *enc) {
    dSP;
    HV *stash = gv_stashpv("Encode::XS", TRUE);
    SV *iv = newSViv(PTR2IV(enc));
    SV *sv = sv_bless(newRV_noinc(iv), stash);
    int i = 0;
    PUSHMARK(sp);
    XPUSHs(sv);
    while (enc->name[i]) {
        const char *name = enc->name[i++];
        XPUSHs(sv_2mortal(newSVpvn(name, strlen(name))));
    }
    PUTBACK;
    call_pv("Encode::define_encoding", G_DISCARD);
    SvREFCNT_dec(sv);
}
XS_EXTERNAL(boot_Enctab) {
    dXSBOOTARGSXSAPIVERCHK;
    PERL_UNUSED_VAR(items);
    /* the module's .pm uses Encode before it loads the module */
    load_module(PERL_LOADMOD_NOIMPORT, newSVpvs("Encode"), NULL);
    Encode_XSEncoding(aTHX_ (encode_t *)&toy_enc_encoding);
    Perl_xs_boot_epilog(aTHX_ ax);
}
`)
	program := `use Encode ();
my $bytes = "AB\x80\x81\x41C\x81\x42?";
my $s = Encode::decode("toy-enc", $bytes);
my $back = Encode::encode("toy-enc", $s);
my $sub = Encode::encode("toy-enc", "A\x{263a}B");
my $in = "A\x81\x43"; my $bad = eval { Encode::decode("toy-enc", $in, Encode::FB_CROAK); 1 } ? "no error" : ($@ =~ /does not map/ ? "does not map" : "other: $@");
my $enc = Encode::find_encoding("toy-enc");
join ",", sprintf("%vx", $s), unpack("H*", $back), unpack("H*", $sub), $bad, $enc->name, ref($enc), $enc->mime_name // "-"`
	if got, want := requireModule(t, "Enctab", so, program), "41.42.e9.3042.43.3044.3f,41428081414381423f,413f42,does not map,toy-enc,Encode::XS,-"; got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}
