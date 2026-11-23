package perl_test

import (
	"testing"

	perl "github.com/goccy/go-perl"
)

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
