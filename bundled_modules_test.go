package perl_test

import (
	"testing"

	perl "github.com/goccy/go-perl"
)

// The XS distributions linked into the interpreter load without anything
// installed, in the default sandbox.
func TestBundledXSModulesLoad(t *testing.T) {
	r := runProgram(t, perl.Config{}, perl.Program{Code: code(
		`for my $m (qw(Clone Test::LeakTrace Package::Stash::XS List::MoreUtils::XS`,
		`    Params::Util HTML::Parser HTML::Entities PadWalker Sub::Name`,
		`    Class::XSAccessor Digest::SHA1 Text::CSV_XS YAML::XS Devel::Size Want`,
		`    Sub::Identify Set::Object Cpanel::JSON::XS)) {`,
		`    eval "require $m; 1" or print "$m: $@";`,
		`}`,
		`print join(",", @{ Clone::clone([1, [2, 3]])->[1] }), "\n";`,
		`print Cpanel::JSON::XS->new->canonical->encode({ a => [1, { b => undef }] }), "\n";`,
		`print join(",", List::MoreUtils::XS::uniq(1, 1, 2, 3, 3)), "\n";`,
		`print Digest::SHA1::sha1_hex("abc"), "\n";`,
		`my $csv = Text::CSV_XS->new({ binary => 1 }); $csv->combine("a", "b,c", 3); print $csv->string, "\n";`,
		`print Sub::Identify::sub_name(Sub::Name::subname("Foo::bar", sub { })), "\n";`,
		`print HTML::Entities::decode_entities("&lt;a&gt;"), "\n";`,
		`print YAML::XS::Load("k: [1, 2]")->{k}[1], "\n";`,
	)}, "")
	r.want(t, 0, "2,3\n"+
		`{"a":[1,{"b":null}]}`+"\n"+
		"1,2,3\n"+
		"a9993e364706816aba3e25717850c26c9cd0d89d\n"+
		`a,"b,c",3`+"\n"+
		"bar\n"+
		"<a>\n"+
		"2\n")
}

// Every encoding family of Encode is linked in, not only the ones
// Encode.pm itself implements.
func TestEncodeEncodingFamilies(t *testing.T) {
	r := runProgram(t, perl.Config{}, perl.Program{Code: code(
		`use Encode;`,
		`for my $e (qw(iso-8859-2 shift_jis euc-jp euc-kr big5 gb2312 UTF-16 cp1047 symbol)) {`,
		`    my $o = find_encoding($e); print "$e=", ($o ? $o->name : "undef"), "\n";`,
		`}`,
		`my $s = "\x{65e5}\x{672c}";`,
		`print length(encode("shift_jis", $s)), decode("shift_jis", encode("shift_jis", $s)) eq $s ? " ok" : " bad", "\n";`,
	)}, "")
	r.want(t, 0, "iso-8859-2=iso-8859-2\nshift_jis=shiftjis\neuc-jp=euc-jp\neuc-kr=euc-kr\n"+
		"big5=big5-eten\ngb2312=euc-cn\nUTF-16=UTF-16\ncp1047=cp1047\nsymbol=symbol\n4 ok\n")
}

// Under -T everything that comes from outside the program is tainted:
// its arguments, its name, the interpreter's name and the environment.
func TestRunProgramTaintsOutsideData(t *testing.T) {
	cfg := perl.Config{Env: []string{"PATH=/bin"}}
	r := runProgram(t, cfg, perl.Program{
		Taint: perl.TaintOn,
		Args:  []string{"arg"},
		Code: code(
			`use Scalar::Util "tainted";`,
			`print join(" ", map { tainted($_) ? 1 : 0 } $ARGV[0], $0, $^X, $ENV{PATH}), "\n";`,
			`print tainted("literal") ? 1 : 0, "\n";`,
		),
	}, "")
	r.want(t, 0, "1 1 1 1\n0\n")
}
