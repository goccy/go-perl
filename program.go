package perl

// Whole-program execution: what the perl command does with the program its
// command line names. The command line itself is the caller's to parse —
// Program is what it means.

import (
	"context"
	"encoding/binary"
	"fmt"
	"unicode/utf8"
)

// WarningMode selects the program-wide warning switch.
type WarningMode uint8

const (
	// WarningsDefault leaves warnings to the program (use warnings).
	WarningsDefault WarningMode = iota
	// WarningsOn enables warnings where the program does not say (-w).
	WarningsOn
	// WarningsAll enables all warnings regardless of the program (-W).
	WarningsAll
	// WarningsNone disables all warnings regardless of the program (-X).
	WarningsNone
)

// TaintMode selects taint checking.
type TaintMode uint8

const (
	// TaintOff runs without taint checks.
	TaintOff TaintMode = iota
	// TaintWarn reports taint violations as warnings (-t).
	TaintWarn
	// TaintOn makes taint violations fatal (-T).
	TaintOn
)

// LoopMode wraps the program in an input loop.
type LoopMode uint8

const (
	// LoopNone runs the program once.
	LoopNone LoopMode = iota
	// LoopRead runs the program once per input line: while (<>) { ... } (-n).
	LoopRead
	// LoopPrint is LoopRead that also prints each line after the program
	// ran on it (-p).
	LoopPrint
)

// The bits of Program.Unicode (${^UNICODE}, the -C switch).
const (
	UnicodeStdin  uint32 = 0x01 // STDIN is UTF-8 (I)
	UnicodeStdout uint32 = 0x02 // STDOUT is UTF-8 (O)
	UnicodeStderr uint32 = 0x04 // STDERR is UTF-8 (E)
	UnicodeIn     uint32 = 0x08 // UTF-8 is the default layer for input streams (i)
	UnicodeOut    uint32 = 0x10 // UTF-8 is the default layer for output streams (o)
	UnicodeArgv   uint32 = 0x20 // @ARGV elements are UTF-8 (A)
	UnicodeLocale uint32 = 0x40 // apply the above only under a UTF-8 locale (L)
	// UnicodeUTF8CacheAssert enables the UTF-8 caching code's assertions (a).
	UnicodeUTF8CacheAssert uint32 = 0x100
)

// Define is one main:: scalar set before the program starts (the -s
// switch's -name and -name=value arguments).
type Define struct {
	Name string
	// Value is the scalar's value; nil sets it to 1.
	Value *string
}

// Program describes one whole-program run: the program text's source, its
// arguments, and the interpreter switches it runs under. The zero value of
// every switch field is perl's behaviour without that switch.
//
// Exactly one of File, Code and Stdin selects the source.
type Program struct {
	// File is the path of the program file — a guest path against
	// Config.FS. It becomes $0.
	File string
	// Code is the program text, one element per line (the -e switch, which
	// may be repeated). $0 is "-e".
	Code []string
	// Stdin reads the program text from the instance's standard input. $0
	// is "-".
	Stdin bool

	// Args is @ARGV.
	Args []string
	// Inc is put in front of @INC, in order (-I).
	Inc []string
	// Preamble holds Perl statements compiled ahead of the program text, in
	// order — what -M and -m contribute ("use Module;").
	Preamble []string
	// Defines are set before the program starts (-s).
	Defines []Define

	// Warnings is the program-wide warning switch (-w, -W, -X).
	Warnings WarningMode
	// Taint selects taint checking (-t, -T).
	Taint TaintMode
	// Features enables the interpreter version's feature bundle for Code,
	// as "use feature ':5.xx'" would (-E).
	Features bool
	// SyntaxOnly compiles the program — BEGIN, UNITCHECK and CHECK blocks
	// run — and reports "syntax OK" instead of running it (-c).
	SyntaxOnly bool
	// Unsafe allows unsafe operations (-U).
	Unsafe bool
	// Debug runs the program under the debugger named by the PERL5DB
	// environment variable, perl5db.pl by default (-d).
	Debug bool

	// Loop wraps the program in an input loop (-n, -p).
	Loop LoopMode
	// AutoSplit splits each input line into @F (-a). It has an effect
	// under Loop only.
	AutoSplit bool
	// SplitPattern, when non-nil, is the pattern AutoSplit splits on in
	// place of whitespace (-F).
	SplitPattern *string
	// LineEnd, when non-nil, enables line-ending processing (-l): under
	// Loop each input line is chomped, and $\ is set to *LineEnd.
	LineEnd *string
	// RecordSeparator, when non-nil, is the input record separator $/
	// (-0): "" reads paragraphs.
	RecordSeparator *string
	// Slurp sets $/ to undef, so a read returns the whole input (-g,
	// -0777). It takes precedence over RecordSeparator.
	Slurp bool
	// InPlace, when non-nil, edits the files <> reads in place, keeping a
	// backup under the suffix *InPlace when it is not empty (-i).
	InPlace *string
	// Unicode, when non-nil, is the ${^UNICODE} flag word (-C): the
	// Unicode* bits.
	Unicode *uint32

	// Extract skips the leading text of the program source up to the first
	// #! line that names perl (-x).
	Extract bool
	// ExtractDir, when not empty, is a directory to change to before the
	// program runs. It has an effect under Extract only.
	ExtractDir string
}

// The flag bits and enumerations of the program description perl-wasm's
// perl_run_program takes (see its pl.h).
const (
	progLoop = 1 << iota
	progLoopPrint
	progAutoSplit
	progSplitPattern
	progLineEndings
	progRSSet
	progRSUndef
	progRSUTF8
	progInPlace
	progSyntaxOnly
	progFeatures
	progExtract
	progUnsafe
	progDebugger
	progUnicode
)

const (
	progSourceFile  = 0
	progSourceText  = 1
	progSourceStdin = 2
)

// encode renders the description in perl_run_program's layout.
func (prog *Program) encode() ([]byte, error) {
	sources := 0
	source := byte(progSourceFile)
	if prog.File != "" {
		sources++
	}
	if prog.Code != nil {
		sources++
		source = progSourceText
	}
	if prog.Stdin {
		sources++
		source = progSourceStdin
	}
	if sources != 1 {
		return nil, fmt.Errorf("perl: a Program takes exactly one of File, Code and Stdin")
	}
	if prog.Warnings > WarningsNone {
		return nil, fmt.Errorf("perl: unknown WarningMode %d", prog.Warnings)
	}
	if prog.Taint > TaintOn {
		return nil, fmt.Errorf("perl: unknown TaintMode %d", prog.Taint)
	}

	var flags uint32
	switch prog.Loop {
	case LoopNone:
	case LoopRead:
		flags |= progLoop
	case LoopPrint:
		flags |= progLoop | progLoopPrint
	default:
		return nil, fmt.Errorf("perl: unknown LoopMode %d", prog.Loop)
	}
	var split, lineEnd, rs, inPlace string
	if prog.AutoSplit {
		flags |= progAutoSplit
	}
	if prog.SplitPattern != nil {
		flags |= progSplitPattern
		split = *prog.SplitPattern
	}
	if prog.LineEnd != nil {
		flags |= progLineEndings
		lineEnd = *prog.LineEnd
	}
	switch {
	case prog.Slurp:
		flags |= progRSSet | progRSUndef
	case prog.RecordSeparator != nil:
		flags |= progRSSet
		rs = *prog.RecordSeparator
		if !isASCII(rs) && utf8.ValidString(rs) {
			flags |= progRSUTF8
		}
	}
	if prog.InPlace != nil {
		flags |= progInPlace
		inPlace = *prog.InPlace
	}
	if prog.SyntaxOnly {
		flags |= progSyntaxOnly
	}
	if prog.Features {
		flags |= progFeatures
	}
	if prog.Extract {
		flags |= progExtract
	}
	if prog.Unsafe {
		flags |= progUnsafe
	}
	if prog.Debug {
		flags |= progDebugger
	}
	var unicode uint32
	if prog.Unicode != nil {
		flags |= progUnicode
		unicode = *prog.Unicode
	}

	// perl appends a newline to every -e argument.
	var code []byte
	for _, line := range prog.Code {
		code = append(code, line...)
		code = append(code, '\n')
	}

	le := binary.LittleEndian
	str := func(b []byte, s string) []byte {
		b = le.AppendUint32(b, uint32(len(s)))
		return append(b, s...)
	}
	strs := func(b []byte, list []string) []byte {
		b = le.AppendUint32(b, uint32(len(list)))
		for _, s := range list {
			b = str(b, s)
		}
		return b
	}
	var b []byte
	b = le.AppendUint32(b, flags)
	b = append(b, byte(prog.Warnings), byte(prog.Taint))
	b = le.AppendUint32(b, unicode)
	b = append(b, source)
	b = str(b, prog.File)
	b = str(b, string(code))
	b = str(b, prog.ExtractDir)
	b = str(b, rs)
	b = str(b, lineEnd)
	b = str(b, split)
	b = str(b, inPlace)
	b = strs(b, prog.Preamble)
	b = strs(b, prog.Inc)
	b = le.AppendUint32(b, uint32(len(prog.Defines)))
	for _, d := range prog.Defines {
		b = str(b, d.Name)
		if d.Value != nil {
			b = append(b, 1)
			b = str(b, *d.Value)
		} else {
			b = append(b, 0)
			b = str(b, "")
		}
	}
	b = strs(b, prog.Args)
	return b, nil
}

func isASCII(s string) bool {
	for i := 0; i < len(s); i++ {
		if s[i] >= utf8.RuneSelf {
			return false
		}
	}
	return true
}

// RunProgram runs prog the way the perl command runs the program its
// command line names: the text is compiled as the main program (no caller
// at its top level, __END__/__DATA__ attached, the switches of its #! line
// honoured), BEGIN, CHECK and INIT blocks run in their phases, then the
// program, then its END blocks. Output goes to the instance's configured
// stdio.
//
// status is the process exit status the run maps to: 0 when the program
// ran to its end, the argument of exit(), and perl's die status (errno, or
// 255) for an uncaught die — whose message perl has already written to
// STDERR. err reports a host failure or the cancellation of ctx, not the
// program's own outcome.
//
// The program is the instance's last act: after RunProgram returns, the
// instance must only be Closed, which flushes its output and runs global
// destruction.
func (p *Perl) RunProgram(ctx context.Context, prog Program) (status int, err error) {
	desc, err := prog.encode()
	if err != nil {
		return 0, err
	}
	return p.raw.RunProgram(ctx, desc)
}
