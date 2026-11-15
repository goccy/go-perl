package perl_test

import (
	"context"
	"math"
	"os"
	"path/filepath"
	"strings"
	"testing"

	perl "github.com/goccy/go-perl"
)

// The interpreter's integers are 64 bits wide (IV and UV): a program
// computes with them as on the 64-bit perls it was written on.
func TestIntegersAre64BitsWide(t *testing.T) {
	p := newPerl(t)
	ctx := context.Background()
	for _, c := range []struct{ name, src, want string }{
		{"the largest unsigned integer", `~0`, "18446744073709551615"},
		{"~0 is an integer", `~0 == 18446744073709551615 ? "equal" : "different"`, "equal"},
		{"a shift beyond 32 bits", `1 << 40`, "1099511627776"},
		{"a shift to the top bit", `1 << 63`, "9223372036854775808"},
		{"the largest signed integer", `9223372036854775807`, "9223372036854775807"},
		{"%d of the largest signed integer", `sprintf "%d", 9223372036854775807`, "9223372036854775807"},
		{"%d of the smallest signed integer", `sprintf "%d", -9223372036854775807 - 1`, "-9223372036854775808"},
		{"%u of the largest unsigned integer", `sprintf "%u", 18446744073709551615`, "18446744073709551615"},
		{"%x, %o and %b beyond 32 bits", `sprintf "%x %o %b", 0xdeadbeefcafe, 0xdeadbeefcafe, 1 << 33`, "deadbeefcafe 6752667673745376 1000000000000000000000000000000000"},
		{"%X of all bits", `sprintf "%X", ~0`, "FFFFFFFFFFFFFFFF"},
		{"integer addition stays exact beyond 2**53", `9007199254740993 + 2`, "9007199254740995"},
		{"integer multiplication stays exact", `3037000499 * 3037000499`, "9223372030926249001"},
		{"the largest signed integer plus one is unsigned", `9223372036854775807 + 1`, "9223372036854775808"},
		{"the largest unsigned integer plus one is a float", `18446744073709551615 + 1`, "1.84467440737096e+19"},
		{"2**64 is a float", `2 ** 64`, "1.84467440737096e+19"},
		{"2**63 is exact", `2 ** 63`, "9.22337203685478e+18"},
		{"an integer overflowing by multiplication is a float", `4294967296 * 4294967296 * 2`, "3.68934881474191e+19"},
		{"subtraction below the smallest integer is a float", `-9223372036854775807 - 2`, "-9.22337203685478e+18"},
		{"use integer wraps at 64 bits", `use integer; 9223372036854775807 + 1`, "-9223372036854775808"},
		{"use integer divides toward zero", `use integer; -9223372036854775807 / 2`, "-4611686018427387903"},
		{"use integer of all bits", `use integer; ~0`, "-1"},
		{"a hexadecimal literal beyond 32 bits", `0xFFFFFFFFFFFFFFFF`, "18446744073709551615"},
		{"hex beyond 32 bits", `hex("ffffffffffff")`, "281474976710655"},
		{"hex of 16 digits", `hex("0x7fffffffffffffff")`, "9223372036854775807"},
		{"oct beyond 32 bits", `oct("0777777777777777777777")`, "9223372036854775807"},
		{"oct of a binary string of 64 bits", `oct("0b" . ("1" x 64))`, "18446744073709551615"},
		{"a binary literal beyond 32 bits", `0b1000000000000000000000000000000000000000`, "549755813888"},
		{"pack and unpack Q", `join ",", unpack("Q2", pack("Q2", 18446744073709551615, 1 << 40))`, "18446744073709551615,1099511627776"},
		{"pack and unpack q", `join ",", unpack("q2", pack("q2", -9223372036854775807 - 1, -(1 << 40)))`, "-9223372036854775808,-1099511627776"},
		{"pack Q is eight bytes, little-endian", `unpack("H*", pack("Q", 0x0102030405060708))`, "0807060504030201"},
		{"pack Q> is big-endian", `unpack("H*", pack("Q>", 0x0102030405060708))`, "0102030405060708"},
		{"pack j and J are the interpreter's integers", `length(pack("j", 0)) . "," . length(pack("J", 0)) . "," . join(",", unpack("j J", pack("j J", -9223372036854775807 - 1, ~0)))`, "8,8,-9223372036854775808,18446744073709551615"},
		{"a bitwise and beyond 32 bits", `0xFFFF0000FFFF0000 & 0x0FF000000FF00000`, "1148417905246863360"},
		{"a right shift of all bits", `~0 >> 1`, "9223372036854775807"},
		{"an integer string beyond 2**53 compares exactly", `"9007199254740993" == 9007199254740993 ? "equal" : "different"`, "equal"},
		{"increment beyond 32 bits", `my $n = 4294967295; $n++; $n`, "4294967296"},
		{"increment at the largest signed integer", `my $n = 9223372036854775807; $n++; $n`, "9223372036854775808"},
		{"int of a float beyond 32 bits", `int(1e15 + 0.5)`, "1000000000000000"},
		{"a range beyond 32 bits", `join ",", 4294967295 .. 4294967297`, "4294967295,4294967296,4294967297"},
		{"modulus beyond 32 bits", `18446744073709551615 % 1000000007`, "582344007"},
		{"printf %s of an integer", `sprintf "%s", 1 << 62`, "4611686018427387904"},
		{"%Config says so", `require Config; join ",", map { "$_=$Config::Config{$_}" } qw(ivsize uvsize ivtype uvtype use64bitint d_quad quadtype u64type i64type ivdformat uvuformat uvxformat nv_preserves_uv_bits ptrsize longsize byteorder)`,
			`ivsize=8,uvsize=8,ivtype=long,uvtype=unsigned long,use64bitint=define,d_quad=define,quadtype=long,u64type=unsigned long,i64type=long,ivdformat="ld",uvuformat="lu",uvxformat="lx",nv_preserves_uv_bits=53,ptrsize=8,longsize=8,byteorder=12345678`},
		{"an NV does not hold every integer", `require Config; defined $Config::Config{d_nv_preserves_uv} ? "holds" : "does not"`, "does not"},
	} {
		r, err := p.Eval(ctx, c.src)
		if err != nil || r.Error != nil {
			t.Errorf("%s: %s: err=%v error=%v", c.name, c.src, err, r.Error)
			continue
		}
		if got := resultStr(r); got != c.want {
			t.Errorf("%s: %s = %q, want %q", c.name, c.src, got, c.want)
		}
	}
}

// The interpreter's C data model is LP64, the one of the 64-bit machines the
// native code a program loads is compiled for: a pointer, a long and a
// size_t are 8 bytes and an int is 4. %Config says so, and so do the pack
// letters whose size is the C type's.
func TestTheCDataModelIsLP64(t *testing.T) {
	p := newPerl(t)
	ctx := context.Background()
	for _, c := range []struct{ name, src, want string }{
		{"%Config says so", `require Config; join ",", map { "$_=" . ($Config::Config{$_} // "undef") } qw(ptrsize longsize sizesize intsize longlongsize use64bitall use64bitint i32type u32type)`,
			"ptrsize=8,longsize=8,sizesize=8,intsize=4,longlongsize=8,use64bitall=define,use64bitint=define,i32type=int,u32type=unsigned int"},
		{"pack l! and L! are a long", `length(pack("l!", 0)) . "," . length(pack("L!", 0))`, "8,8"},
		{"pack i and I are an int", `length(pack("i", 0)) . "," . length(pack("I", 0))`, "4,4"},
		{"pack p is a pointer", `length(pack("p", "x"))`, "8"},
		{"pack l! holds a 64-bit long", `join ",", unpack("l! L!", pack("l! L!", -9223372036854775807 - 1, ~0))`, "-9223372036854775808,18446744073709551615"},
		{"pack p and unpack p round-trip a string", `unpack("p", pack("p", "a string"))`, "a string"},
	} {
		r, err := p.Eval(ctx, c.src)
		if err != nil || r.Error != nil {
			t.Errorf("%s: %s: err=%v error=%v", c.name, c.src, err, r.Error)
			continue
		}
		if got := resultStr(r); got != c.want {
			t.Errorf("%s: %s = %q, want %q", c.name, c.src, got, c.want)
		}
	}
}

// A file offset beyond 4 GiB is an integer: the size of a file, a position
// sought and reported.
func TestFileOffsetsBeyondFourGigabytes(t *testing.T) {
	path := filepath.Join(t.TempDir(), "sparse")
	const size = 5<<30 + 123
	f, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	// A file with a hole: no host this runs on stores the five gigabytes.
	if err := f.Truncate(size); err != nil {
		f.Close()
		t.Skipf("no file of %d bytes here: %v", size, err)
	}
	if _, err := f.WriteAt([]byte("tail"), size-4); err != nil {
		f.Close()
		t.Fatal(err)
	}
	f.Close()
	r := runProgram(t, hostConfig(t), perl.Program{
		Code: code(
			`my $path = shift;`,
			`print "size=", -s $path, "\n";`,
			`open(my $fh, "<", $path) or die "open: $!";`,
			`print "sysseek=", sysseek($fh, 5368709120, 0), "\n";`,
			`print "position=", sysseek($fh, 0, 1), "\n";`,
			`print "seek=", (seek($fh, -4, 2) ? 1 : 0), " tell=", tell($fh), "\n";`,
			`read($fh, my $tail, 4); print "tail=$tail\n";`,
			`print "stat=", (stat $path)[7], "\n";`,
		),
		Args: []string{path},
	}, "")
	want := "size=5368709243\nsysseek=5368709120\nposition=5368709120\nseek=1 tell=5368709239\ntail=tail\nstat=5368709243\n"
	r.want(t, 0, want)
}

// Every int64 and uint64 crosses the bridge as the integer it is, in both
// directions.
func TestIntegersCrossTheBridgeAtFullWidth(t *testing.T) {
	p := newPerl(t)
	ctx := context.Background()

	// Perl to Go.
	for _, c := range []struct {
		src      string
		kind     perl.Kind
		signed   int64
		unsigned uint64
		text     string
	}{
		{`9223372036854775807`, perl.KindInt, math.MaxInt64, math.MaxInt64, "9223372036854775807"},
		{`-9223372036854775807 - 1`, perl.KindInt, math.MinInt64, 1 << 63, "-9223372036854775808"},
		{`18446744073709551615`, perl.KindUint, -1, math.MaxUint64, "18446744073709551615"},
		{`~0`, perl.KindUint, -1, math.MaxUint64, "18446744073709551615"},
		{`9223372036854775808`, perl.KindUint, math.MinInt64, 1 << 63, "9223372036854775808"},
		{`9007199254740993`, perl.KindInt, 9007199254740993, 9007199254740993, "9007199254740993"},
		{`-9007199254740993`, perl.KindInt, -9007199254740993, 1<<64 - 9007199254740993, "-9007199254740993"},
		{`1 << 40`, perl.KindInt, 1 << 40, 1 << 40, "1099511627776"},
		{`4294967295`, perl.KindInt, 4294967295, 4294967295, "4294967295"},
		{`"18446744073709551615"`, perl.KindString, -1, math.MaxUint64, "18446744073709551615"},
		{`"-9223372036854775808"`, perl.KindString, math.MinInt64, 1 << 63, "-9223372036854775808"},
		{`" 9007199254740993 apples"`, perl.KindString, 9007199254740993, 9007199254740993, " 9007199254740993 apples"},
		{`2 ** 64`, perl.KindFloat, -1, math.MaxUint64, "1.84467440737096e+19"},
		{`-(2 ** 70)`, perl.KindFloat, math.MinInt64, 1 << 63, "-1.18059162071741e+21"},
		{`1.5e10`, perl.KindFloat, 15000000000, 15000000000, "15000000000"},
		{`-7.9`, perl.KindFloat, -7, 1<<64 - 7, "-7.9"},
	} {
		r, err := p.Eval(ctx, c.src)
		if err != nil || r.Error != nil {
			t.Fatalf("%s: err=%v error=%v", c.src, err, r.Error)
		}
		sv, err := perl.As[perl.ScalarValue](r.Value)
		if err != nil {
			t.Fatalf("%s: %v", c.src, err)
		}
		if sv.Kind() != c.kind || sv.Int() != c.signed || sv.Uint() != c.unsigned || sv.String() != c.text {
			t.Errorf("%s: kind %v, Int() %d, Uint() %d, String() %q; want %v, %d, %d, %q",
				c.src, sv.Kind(), sv.Int(), sv.Uint(), sv.String(), c.kind, c.signed, c.unsigned, c.text)
		}
	}

	// Go to Perl, and back.
	r, err := p.Eval(ctx, `sub { (join(",", map { "$_" } @_), @_) }`)
	if err != nil || r.Error != nil {
		t.Fatalf("define: err=%v error=%v", err, r.Error)
	}
	ref, err := perl.As[perl.RefValue](r.Value)
	if err != nil {
		t.Fatal(err)
	}
	target, err := ref.Deref(ctx)
	if err != nil {
		t.Fatal(err)
	}
	echo, err := perl.As[perl.CodeValue](target)
	if err != nil {
		t.Fatal(err)
	}
	sent := []perl.Value{
		perl.ValueOf(int64(math.MaxInt64)), perl.ValueOf(int64(math.MinInt64)),
		perl.ValueOf(uint64(math.MaxUint64)), perl.ValueOf(uint64(1) << 63), perl.ValueOf(uint64(math.MaxInt64)),
		perl.ValueOf(uint(7)), perl.ValueOf(int64(9007199254740993)), perl.ValueOf(int64(-9007199254740993)),
	}
	out, err := echo.Call(ctx, sent...)
	if err != nil {
		t.Fatal(err)
	}
	if len(out) != len(sent)+1 {
		t.Fatalf("Call returned %d values, want %d", len(out), len(sent)+1)
	}
	texts := []string{
		"9223372036854775807", "-9223372036854775808", "18446744073709551615", "9223372036854775808",
		"9223372036854775807", "7", "9007199254740993", "-9007199254740993",
	}
	seen, err := perl.As[perl.ScalarValue](out[0])
	if err != nil {
		t.Fatal(err)
	}
	if want := strings.Join(texts, ","); seen.String() != want {
		t.Errorf("Go integers in Perl = %q, want %q", seen.String(), want)
	}
	for i, v := range out[1:] {
		back, err := perl.As[perl.ScalarValue](v)
		if err != nil {
			t.Fatal(err)
		}
		want := sent[i].(perl.ScalarValue)
		if back.Kind() != want.Kind() || back.Int() != want.Int() || back.Uint() != want.Uint() {
			t.Errorf("value %d came back as kind %v, Int() %d, Uint() %d; sent kind %v, %d, %d",
				i, back.Kind(), back.Int(), back.Uint(), want.Kind(), want.Int(), want.Uint())
		}
		if back.String() != texts[i] {
			t.Errorf("value %d came back as %q, want %q", i, back.String(), texts[i])
		}
	}

	// A Go function gets and returns them.
	if err := p.Bind("go_plus_one", func(args []perl.Value) ([]perl.Value, error) {
		sv, err := perl.As[perl.ScalarValue](args[0])
		if err != nil {
			return nil, err
		}
		if sv.Kind() == perl.KindUint {
			return []perl.Value{perl.ValueOf(sv.Uint() - 1)}, nil
		}
		return []perl.Value{perl.ValueOf(sv.Int() + 1)}, nil
	}); err != nil {
		t.Fatal(err)
	}
	r, err = p.Eval(ctx, `join ",", go_plus_one(9223372036854775806), go_plus_one(18446744073709551615), go_plus_one(-9223372036854775807 - 1)`)
	if err != nil || r.Error != nil {
		t.Fatalf("go_plus_one: err=%v error=%v", err, r.Error)
	}
	if got, want := resultStr(r), "9223372036854775807,18446744073709551614,-9223372036854775807"; got != want {
		t.Errorf("through a Go function: %q, want %q", got, want)
	}
	if got := perl.KindUint.String(); got != "uint" {
		t.Errorf("KindUint is called %q", got)
	}
}

// The scalars Go makes read back as Perl would read them, as a signed and
// as an unsigned integer.
func TestScalarIntegerAccessors(t *testing.T) {
	for _, c := range []struct {
		name     string
		value    perl.ScalarValue
		kind     perl.Kind
		signed   int64
		unsigned uint64
		text     string
	}{
		{"int64 max", perl.ValueOf(int64(math.MaxInt64)), perl.KindInt, math.MaxInt64, math.MaxInt64, "9223372036854775807"},
		{"int64 min", perl.ValueOf(int64(math.MinInt64)), perl.KindInt, math.MinInt64, 1 << 63, "-9223372036854775808"},
		{"uint64 max", perl.ValueOf(uint64(math.MaxUint64)), perl.KindUint, -1, math.MaxUint64, "18446744073709551615"},
		{"uint64 that an int64 holds", perl.ValueOf(uint64(42)), perl.KindInt, 42, 42, "42"},
		{"uint", perl.ValueOf(uint(1) << 40), perl.KindInt, 1 << 40, 1 << 40, "1099511627776"},
		{"uint32 max", perl.ValueOf(uint32(math.MaxUint32)), perl.KindInt, math.MaxUint32, math.MaxUint32, "4294967295"},
		{"a string of an integer beyond 2**53", perl.ValueOf("-9007199254740993"), perl.KindString, -9007199254740993, 1<<64 - 9007199254740993, "-9007199254740993"},
		{"a string of the largest unsigned integer", perl.ValueOf("18446744073709551615"), perl.KindString, -1, math.MaxUint64, "18446744073709551615"},
		{"a string of an integer beyond the unsigned range", perl.ValueOf("18446744073709551616"), perl.KindString, -1, math.MaxUint64, "18446744073709551616"},
		{"a string of an integer below the signed range", perl.ValueOf("-9223372036854775809"), perl.KindString, math.MinInt64, 1 << 63, "-9223372036854775809"},
		{"a string with a fraction", perl.ValueOf("12.75abc"), perl.KindString, 12, 12, "12.75abc"},
		{"a string with an exponent", perl.ValueOf("1e3"), perl.KindString, 1000, 1000, "1e3"},
		{"a string that is no number", perl.ValueOf("abc"), perl.KindString, 0, 0, "abc"},
		{"a float beyond the signed range", perl.ValueOf(1e19), perl.KindFloat, -8446744073709551616, 10000000000000000000, "1e+19"},
		{"a float beyond every integer", perl.ValueOf(1e30), perl.KindFloat, -1, math.MaxUint64, "1e+30"},
		{"a float below every integer", perl.ValueOf(-1e30), perl.KindFloat, math.MinInt64, 1 << 63, "-1e+30"},
		{"not a number", perl.ValueOf(math.NaN()), perl.KindFloat, 0, 0, "NaN"},
		{"true", perl.ValueOf(true), perl.KindBool, 1, 1, "1"},
		{"undef", perl.Undef(), perl.KindUndef, 0, 0, ""},
	} {
		v := c.value
		if v.Kind() != c.kind || v.Int() != c.signed || v.Uint() != c.unsigned || v.String() != c.text {
			t.Errorf("%s: kind %v, Int() %d, Uint() %d, String() %q; want %v, %d, %d, %q",
				c.name, v.Kind(), v.Int(), v.Uint(), v.String(), c.kind, c.signed, c.unsigned, c.text)
		}
	}
	if !perl.ValueOf(uint64(math.MaxUint64)).Bool() || perl.ValueOf(uint64(math.MaxUint64)).Float() != 18446744073709551615.0 {
		t.Errorf("the largest unsigned integer is not true, or not itself as a float")
	}
}
