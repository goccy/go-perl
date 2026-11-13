//go:build darwin || linux

package xs_test

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"testing"
)

// buildPerlAPI compiles testdata/perlapi/PerlAPI.c against the SDK headers.
// Undefined symbols are refused at link time, so a function the headers
// declare without providing fails here and not when the module is loaded.
func buildPerlAPI(t testing.TB) string {
	t.Helper()
	cc, err := exec.LookPath("cc")
	if err != nil {
		t.Skip("no host C compiler available")
	}
	so := filepath.Join(t.TempDir(), "PerlAPI.so")
	args := []string{"-shared", "-fPIC", "-Wall", "-Werror=implicit-function-declaration",
		"-I", "sdk/include", "-o", so, "testdata/perlapi/PerlAPI.c", "-lm"}
	if runtime.GOOS == "darwin" {
		switch runtime.GOARCH {
		case "arm64":
			args = append([]string{"-arch", "arm64"}, args...)
		case "amd64":
			args = append([]string{"-arch", "x86_64"}, args...)
		}
	} else {
		// perl's ccflags on Linux carry -D_GNU_SOURCE (gperl's do too): a
		// source file may include a system header before perl.h.
		args = append([]string{"-Wl,-z,defs", "-D_GNU_SOURCE"}, args...)
	}
	cmd := exec.Command(cc, args...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("compile PerlAPI.so: %v\n%s", err, out)
	}
	return so
}

// TestPerlAPISurface runs XSUBs written against the perl API the SDK
// provides and compares what they observe with what perl itself documents
// for each function or macro.
func TestPerlAPISurface(t *testing.T) {
	so := buildPerlAPI(t)
	p := newHostPerl(t)
	ctx := context.Background()
	loadXSModule(t, p, "PerlAPI", so)

	libdir := t.TempDir()
	if err := os.WriteFile(filepath.Join(libdir, "PerlAPI.pm"),
		[]byte("package PerlAPI;\nuse XSLoader;\nXSLoader::load(__PACKAGE__);\n1;\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	setup := `unshift @INC, '` + libdir + `'; require PerlAPI;
package Counter; sub TIESCALAR { my $n = 0; bless \$n } sub FETCH { ${$_[0]}++; 21 }
package main; sub joiner { join "+", @_ } our ($foo, $loc); 1;`
	if r, err := p.Eval(ctx, setup); err != nil || r.Error != nil {
		t.Fatalf("setup: err=%v error=%v", err, r.Error)
	}

	for _, tc := range []struct{ name, code, want string }{
		{"flag word of an integer", `my $x = 42; PerlAPI::flags($x)`, "IOK,"},
		{"flag word of a string", `my $x = "a"; PerlAPI::flags($x)`, "POK,"},
		{"flag word of a number", `my $x = 1.5; PerlAPI::flags($x)`, "NOK,"},
		{"flag word of a reference", `my $x = []; PerlAPI::flags($x)`, "ROK,"},
		{"flag word of a wide string", `my $x = "\x{263a}"; PerlAPI::flags($x)`, "POK,UTF8,"},
		{"flag word of a constant", `PerlAPI::flags("lit")`, "POK,READONLY,"},
		{"flag word of a tied scalar", `tie my $t, 'Counter'; PerlAPI::flags($t) =~ /GMG/ ? "magical" : "plain"`, "magical"},
		{"SVf_OK mask without SVf_ROK", `join ",", map { PerlAPI::is_plain_value($_) ? 1 : 0 } 1, "a", [], undef, \"x"`, "1,1,0,0,0"},
		{"prototype given to newXSproto", `prototype(\&PerlAPI::is_plain_value)`, "$"},
		{"prototype enforced at compile time", `eval q{ PerlAPI::is_plain_value() }; $@ =~ /^Not enough arguments for PerlAPI::is_plain_value/ ? "refused" : "accepted: $@"`, "refused"},
		{"SvPOK_only", `my $d = "12abc"; { no warnings; my $n = $d + 0 } PerlAPI::string_only($d); PerlAPI::flags($d) . $d`, "POK,12abc"},
		{"SvOK_off", `my $v = "abc"; PerlAPI::ok_off($v); defined($v) ? "defined" : "undef"`, "undef"},
		{"one FETCH for SvGETMAGIC and two nomg reads", `tie my $t, 'Counter'; my $sum = PerlAPI::read_twice_nomg($t); $sum . "/" . ${tied $t}`, "42/1"},
		{"SvPV_nomg", `PerlAPI::string_nomg("plain text")`, "plain text"},
		{"sv_setref_pvn", `my $o = PerlAPI::new_object("My::Class", "payload"); ref($o) . ":" . $$o`, "My::Class:payload"},
		{"grok_hex", `join ",", PerlAPI::hex("ff")`, "255,2,0"},
		{"grok_hex prefix and underscore", `join ",", PerlAPI::hex("0x1_0")`, "16,5,0"},
		{"grok_hex stops at a bad digit", `join ",", PerlAPI::hex("1fz9")`, "31,2,0"},
		{"grok_hex beyond 32 bits", `join ",", PerlAPI::hex("1ffffffff")`, "8589934591,9,0"},
		{"grok_hex of 16 digits", `join ",", PerlAPI::hex("ffffffffffffffff")`, "18446744073709551615,16,0"},
		{"grok_hex past the integer range", `no warnings 'overflow'; join ",", PerlAPI::hex("1ffffffffffffffff")`, "3.68934881474191e+19,17,4"},
		{"PL_errgv", `eval { die "boom\n" }; PerlAPI::errsv_via_errgv()`, "boom\n"},
		{"call_argv", `PerlAPI::call_with_argv("main::joiner")`, "left+right"},
		{"require_pv", `PerlAPI::require("File/Basename.pm") && exists $INC{"File/Basename.pm"} ? "loaded" : "missing"`, "loaded"},
		{"require_pv failure stays in $@", `my $ok = PerlAPI::require("No/Such/Module.pm"); ($ok ? "loaded" : "failed") . ($@ =~ /^Can't locate No\/Such\/Module.pm/ ? "+message" : "")`, "failed+message"},
		{"gv_fullname4", `PerlAPI::glob_name(\*foo, 1) . "|" . PerlAPI::glob_name(\*foo, 0)`, "*main::foo|*foo"},
		{"SvSetSV_nosteal", `my $src = "keep"; PerlAPI::copy_nosteal($src) . $src`, "keepkeep"},
		{"XSRETURN_UV", `"" . PerlAPI::big()`, "4000000000"},
		{"CvANON", `my $y = 5; PerlAPI::sub_kind(sub { $y })`, "anon,"},
		{"CvCONST", `use constant PI => 3; PerlAPI::sub_kind(\&PI)`, "const,xsub,"},
		{"CvISXSUB", `PerlAPI::sub_kind(\&PerlAPI::big)`, "xsub,"},
		{"PUSHSTACKi and POPSTACK around a callback", `PerlAPI::call_on_new_stack(sub { $_[0] * 2 }, 21)`, "42"},
		{"UTF-8 inspection", `join ",", map { $_ || 0 } PerlAPI::utf8_shape("h\x{e9}llo\x{263a}")`, "1,6,9,6"},
		{"UTF-8 inspection of malformed bytes", `join ",", map { $_ || 0 } PerlAPI::utf8_shape("ab\xff")`, "0,2,2,3"},
		{"av_delete, HvUSEDKEYS, hv_existss", `my @a = (10, 20, 30); my %h = (k => 1, j => 2); join ",", PerlAPI::container_ops(\@a, \%h), scalar(@a), defined $a[1] ? "defined" : "undef"`, "20,2,1,3,undef"},
		{"PL_na", `PerlAPI::length_via_na("hello")`, "5"},
		{"version macros", `PerlAPI::version() eq sprintf("%vd", $^V) ? "same" : PerlAPI::version()`, "same"},
		{"PL_phase", `PerlAPI::phase()`, "RUN"},
		{"SvUPGRADE of an SV of a higher type", `my $s = "12"; bless \$s, "Marked"; PerlAPI::length_after_upgrade($s)`, "2"},
		{"an API function that dies ends the XSUB", `my $s = "12"; bless \$s, "Marked"; eval { PerlAPI::lower_type($s) }; $@ =~ /^sv_upgrade from type 7 down to type 3/ ? "died in the API" : "other: $@"`, "died in the API"},
		{"croak_sv with an object", `my $e = bless { code => 7 }, "My::Err"; eval { PerlAPI::throw($e) }; ref($@) . ":" . $@->{code} . ":" . ($@ == $e ? "same" : "copy")`, "My::Err:7:same"},
		{"croak_sv with a string", `eval { PerlAPI::throw("plain\n") }; $@`, "plain\n"},
		{"croak_sv with a long string", `my $m = ("x" x 600) . "\n"; eval { PerlAPI::throw($m) }; length($@)`, "601"},
		{"croak_sv adds the location to an unterminated message", `eval { PerlAPI::throw("no newline") }; $@ =~ /^no newline at .* line \d+\.$/ ? "located" : "bare: $@"`, "located"},
		{"croak(NULL) dies with $@", `eval { PerlAPI::rethrow_errsv(bless [], "My::Other") }; ref $@`, "My::Other"},
		{"grok_number", `join "|", map { PerlAPI::number_kind($_) } "42", "-1", "0.00", " 12 ", "1e3", "-.5", ".", "12abc", "", "4294967296", "18446744073709551615", "18446744073709551616", "0 but true", "Inf", "-Infinity", "NaN", "1.#INF"`,
			"uv=42,|uv=1,neg,|uv=0,notint,|uv=12,|notint,|uv=0,notint,neg,||||uv=4294967296,|uv=18446744073709551615,|big,|uv=0,|notint,inf,|notint,neg,inf,|notint,nan,|uv=1,notint,inf,"},
		{"a nested XSUB that croaks leaves its caller running", `PerlAPI::call_on_new_stack(sub { eval { PerlAPI::throw("inner\n") }; "survived:$@" }, 1)`, "survived:inner\n"},
		{"av_store reports that it took the value", `my $x = "v"; my @a; my $ok = PerlAPI::store_first(\@a, $x); join ",", ($ok ? 1 : 0), Internals::SvREFCNT($x), $a[0]`, "1,2,v"},
		{"an exception passes through call_sv without G_EVAL", `my $e = bless [], "My::Err"; eval { PerlAPI::call_plain(sub { die $e }) }; ref($@) . ($@ == $e ? ":same" : ":copy")`, "My::Err:same"},
		{"GIMME_V", `my $s = PerlAPI::context(); my ($l) = PerlAPI::context(); PerlAPI::context(); "$s,$l,$PerlAPI::last_context"`, "scalar,list,void"},
		{"utf8n_to_uvchr", `join ",", map { PerlAPI::first_char($_) } "A", "\xc3\xa9", "\xe2\x98\xba", "\xa0", "\xc3", "\xc0\x80"`, "65/1,233/2,9786/3,malformed,malformed,malformed"},
		{"croak after a directly called XSUB that returned bare", `PerlAPI::bare_return(1, 2); eval { PerlAPI::direct_then_croak() }; $@ =~ /^after the direct call at / ? "caught" : "other: $@"`, "caught"},
		{"save_item", `$loc = "outside"; my $seen; PerlAPI::localize(\*loc, sub { $seen = $loc }); "$seen/$loc"`, "inside/outside"},
	} {
		r, err := p.Eval(ctx, tc.code)
		if err != nil {
			t.Errorf("%s: Eval: %v", tc.name, err)
			continue
		}
		if r.Error != nil {
			t.Errorf("%s: died: %v", tc.name, r.Error)
			continue
		}
		if got := resultStr(r); got != tc.want {
			t.Errorf("%s: got %q, want %q", tc.name, got, tc.want)
		}
	}
}
