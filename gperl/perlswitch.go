package gperl

// The perl command line.
//
// `gperl run` takes the command line of perl(1) — perlrun's switches, the
// program file, its arguments — and this file is its reader: a rendering
// of the switch loop in perl.c (S_parse_body and Perl_moreswitches), switch
// for switch, including how switches cluster and which ones take the rest
// of their argument. Its result is a perl.Program, the description the
// interpreter runs; nothing here executes Perl.

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"unicode"

	perl "github.com/goccy/go-perl"
)

// cliAction is what a command line asks for besides running a program.
type cliAction int

const (
	actionRun     cliAction = iota
	actionUsage             // -h, -?, --help
	actionVersion           // -v, --version
)

// perlCommand is one parsed perl command line.
type perlCommand struct {
	prog   perl.Program
	action cliAction
	// switches reports whether the command line itself carried a switch.
	switches bool
	// search looks the program file up on PATH (-S).
	search bool
	// env holds the environment variables the switches set (-d).
	env []string
	// notes are the diagnostics perl prints while reading the switches.
	notes []string

	// The current input record separator, which -l copies: rs when rsSet,
	// "\n" otherwise; rsUndef is -g and -0777.
	rsSet   bool
	rsUndef bool
	rs      string
	// bitBucket: a switch said to read no program at all (-V).
	bitBucket bool
	// defines: the program's leading arguments are -name[=value]
	// switches (-s).
	defines bool
}

// bitBucketPath is the empty program perl reads when a switch said not to
// look for one.
const bitBucketPath = "/dev/null"

// parsePerlCommand reads a perl command line: argv without the command
// name, and the environment it is given in (for PERL5OPT and, under -S,
// PATH).
func parsePerlCommand(argv, env []string) (*perlCommand, error) {
	c := &perlCommand{}
	i := 0
args:
	for ; i < len(argv); i++ {
		a := argv[i]
		if len(a) < 2 || a[0] != '-' {
			break
		}
		c.switches = true
		s := a[1:]
		for s != "" {
			switch ch := s[0]; ch {
			case 'C', ' ', '0', 'F', 'a', 'c', 'd', 'D', 'g', '?', 'h', 'i', 'l',
				'n', 'p', 's', 'u', 'U', 'v', 'W', 'X', 'w', '\r':
				rest, more, err := c.moreSwitches(s)
				if err != nil {
					return nil, err
				}
				if c.action != actionRun {
					return c, nil
				}
				if !more {
					rest = ""
				}
				s = rest
			case 'M', 'm':
				arg := s[1:]
				if arg == "" {
					if i+1 >= len(argv) {
						return nil, fmt.Errorf("Missing argument to -%c", ch)
					}
					i++
					arg = argv[i]
				}
				if err := c.useModule(ch, arg); err != nil {
					return nil, err
				}
				s = ""
			case 't':
				if c.prog.Taint == perl.TaintOff {
					c.prog.Taint = perl.TaintWarn
				}
				s = s[1:]
			case 'T':
				c.prog.Taint = perl.TaintOn
				s = s[1:]
			case 'E', 'e':
				if ch == 'E' {
					c.prog.Features = true
				}
				code := s[1:]
				if code == "" {
					if i+1 >= len(argv) {
						return nil, fmt.Errorf("No code specified for -%c", ch)
					}
					i++
					code = argv[i]
				}
				c.prog.Code = append(c.prog.Code, code)
				s = ""
			case 'f':
				// No sitecustomize.pl: the embedded perl never loads one.
				s = s[1:]
			case 'I':
				dir := s[1:]
				if dir == "" && i+1 < len(argv) {
					i++
					dir = argv[i]
				}
				if dir == "" {
					return nil, fmt.Errorf("No directory specified for -I")
				}
				c.prog.Inc = append(c.prog.Inc, dir)
				s = ""
			case 'S':
				c.search = true
				s = s[1:]
			case 'V':
				c.bitBucket = true
				s = s[1:]
				if strings.HasPrefix(s, ":") {
					c.prog.Preamble = append(c.prog.Preamble,
						"use Config; Config::config_vars(qw\x00"+s[1:]+"\x00)")
					s = ""
				} else {
					c.prog.Preamble = append(c.prog.Preamble, "use Config; Config::_V()")
				}
			case 'x':
				c.prog.Extract = true
				c.prog.ExtractDir = s[1:]
				s = ""
			case '-':
				long := s[1:]
				if long == "" || isPerlSpace(long[0]) {
					i++
					break args
				}
				switch long {
				case "version":
					c.action = actionVersion
					return c, nil
				case "help":
					c.action = actionUsage
					return c, nil
				}
				return nil, unrecognizedSwitch(s)
			default:
				return nil, unrecognizedSwitch(s)
			}
		}
	}

	if err := c.applyPERL5OPT(envValue(env, "PERL5OPT")); err != nil {
		return nil, err
	}
	if c.action != actionRun {
		return c, nil
	}

	rest := argv[i:]
	switch {
	case c.prog.Code != nil:
	case c.bitBucket:
		c.prog.File = bitBucketPath
		if len(rest) > 0 {
			rest = rest[1:] // perl skips it as the program's name
		}
	case len(rest) == 0:
		c.prog.Stdin = true
	case rest[0] == "-":
		c.prog.Stdin = true
		rest = rest[1:]
	default:
		c.prog.File = rest[0]
		rest = rest[1:]
		if c.search {
			found, err := searchProgram(c.prog.File, envValue(env, "PATH"))
			if err != nil {
				return nil, err
			}
			c.prog.File = found
		}
	}
	if c.defines {
		rest = c.takeDefines(rest)
	}
	c.prog.Args = append([]string(nil), rest...)
	return c, nil
}

func unrecognizedSwitch(s string) error {
	return fmt.Errorf("Unrecognized switch: -%s  (-h will show valid options)", s)
}

// isPerlSpace is perl's isSPACE for the ASCII range.
func isPerlSpace(b byte) bool {
	switch b {
	case ' ', '\t', '\n', '\r', '\f', '\v':
		return true
	}
	return false
}

// isWordChar is perl's isWORDCHAR for the ASCII range.
func isWordChar(b byte) bool {
	return b == '_' || '0' <= b && b <= '9' || 'a' <= b && b <= 'z' || 'A' <= b && b <= 'Z'
}

// untilSpace splits s at its first whitespace character.
func untilSpace(s string) (word, rest string) {
	for i := 0; i < len(s); i++ {
		if isPerlSpace(s[i]) {
			return s[:i], s[i:]
		}
	}
	return s, ""
}

// moreSwitches applies the switch at s[0] — one of those that may also
// appear in PERL5OPT — and returns what follows it. more is false when the
// switch ends its argument.
func (c *perlCommand) moreSwitches(s string) (rest string, more bool, err error) {
	switch ch := s[0]; ch {
	case '0':
		return c.recordSeparator(s)
	case 'C':
		flags, rest, err := parseUnicodeOptions(s[1:])
		if err != nil {
			return "", false, err
		}
		c.prog.Unicode = &flags
		return rest, true, nil
	case 'F':
		pattern, rest := untilSpace(s[1:])
		c.prog.SplitPattern = &pattern
		c.prog.AutoSplit = true
		c.loop(perl.LoopRead)
		return rest, true, nil
	case 'a':
		c.prog.AutoSplit = true
		c.loop(perl.LoopRead)
		return s[1:], true, nil
	case 'c':
		c.prog.SyntaxOnly = true
		return s[1:], true, nil
	case 'd':
		return c.debugger(s[1:])
	case 'D':
		c.notes = append(c.notes, "Recompile perl with -DDEBUGGING to use -D switch (did you mean -d ?)")
		s = s[1:]
		for s != "" && isWordChar(s[0]) {
			s = s[1:]
		}
		return s, true, nil
	case 'g':
		c.rsSet, c.rsUndef, c.rs = true, true, ""
		c.prog.Slurp = true
		c.prog.RecordSeparator = nil
		return s[1:], true, nil
	case '?', 'h':
		c.action = actionUsage
		return "", false, nil
	case 'i':
		suffix, rest := untilSpace(s[1:])
		c.prog.InPlace = &suffix
		return rest, true, nil
	case 'I':
		// The PERL5OPT form: the directory follows in the same word, and
		// goes in front of the ones named so far.
		dir := strings.TrimLeftFunc(s[1:], unicode.IsSpace)
		dir, _ = untilSpace(dir)
		if dir == "" {
			return "", false, fmt.Errorf("No directory specified for -I")
		}
		c.prog.Inc = append([]string{dir}, c.prog.Inc...)
		return "", false, nil
	case 'l':
		return c.lineEndings(s[1:])
	case 'M', 'm':
		if s[1:] == "" {
			return "", false, fmt.Errorf("Missing argument to -%c", ch)
		}
		return "", false, c.useModule(ch, s[1:])
	case 'n':
		c.loop(perl.LoopRead)
		return s[1:], true, nil
	case 'p':
		c.loop(perl.LoopPrint)
		return s[1:], true, nil
	case 's':
		c.defines = true
		return s[1:], true, nil
	case 'u':
		return "", false, fmt.Errorf("-u (dump core after parsing the program) is not available")
	case 'U':
		c.prog.Unsafe = true
		return s[1:], true, nil
	case 'v':
		c.action = actionVersion
		return "", false, nil
	case 'w':
		// -w does not undo an earlier -W or -X.
		if c.prog.Warnings == perl.WarningsDefault {
			c.prog.Warnings = perl.WarningsOn
		}
		return s[1:], true, nil
	case 'W':
		c.prog.Warnings = perl.WarningsAll
		return s[1:], true, nil
	case 'X':
		c.prog.Warnings = perl.WarningsNone
		return s[1:], true, nil
	case ' ':
		s = strings.TrimLeft(s, " ")
		if strings.HasPrefix(s, "-") {
			return s[1:], true, nil
		}
		return "", false, nil
	case '-', '\r', '\n', '\t':
		return "", false, nil
	}
	return "", false, fmt.Errorf("Unrecognized switch: -%.1s  (-h will show valid options)", s)
}

// loop turns on the input loop; -p wins over -n.
func (c *perlCommand) loop(mode perl.LoopMode) {
	if mode > c.prog.Loop {
		c.prog.Loop = mode
	}
}

// recordSeparator is -0: s starts at the '0'.
func (c *perlCommand) recordSeparator(s string) (rest string, more bool, err error) {
	set := func(rs string) {
		c.rsSet, c.rsUndef, c.rs = true, false, rs
		c.prog.Slurp = false
		c.prog.RecordSeparator = &rs
	}
	if len(s) > 2 && s[1] == 'x' {
		digits := s[2:]
		n := 0
		for n < len(digits) && isHexDigit(digits[n]) {
			n++
		}
		if n < len(digits) {
			// Not a hexadecimal number after all: -0 followed by -x with
			// a directory.
			set("\x00")
			return s[1:], true, nil
		}
		v, err := strconv.ParseUint(digits, 16, 32)
		if err != nil || v > unicode.MaxRune || !validRune(rune(v)) {
			return "", false, fmt.Errorf("-0x%s is not a character", digits)
		}
		set(string(rune(v)))
		return "", true, nil
	}
	// Up to four octal digits, the leading 0 included.
	n, v := 0, 0
	for n < len(s) && n < 4 && '0' <= s[n] && s[n] <= '7' {
		v = v*8 + int(s[n]-'0')
		n++
	}
	switch {
	case v > 0xff:
		c.rsSet, c.rsUndef, c.rs = true, true, ""
		c.prog.Slurp = true
		c.prog.RecordSeparator = nil
	case v == 0 && n >= 2:
		set("") // paragraph mode
	default:
		set(string([]byte{byte(v)}))
	}
	return s[n:], true, nil
}

func isHexDigit(b byte) bool {
	return '0' <= b && b <= '9' || 'a' <= b && b <= 'f' || 'A' <= b && b <= 'F'
}

func validRune(r rune) bool {
	return !(0xD800 <= r && r <= 0xDFFF)
}

// lineEndings is -l: s is what follows the 'l'.
func (c *perlCommand) lineEndings(s string) (rest string, more bool, err error) {
	if s != "" && '0' <= s[0] && s[0] <= '9' {
		// Three octal digits, four when the first is 0.
		limit := 3
		if s[0] == '0' {
			limit = 4
		}
		n, v := 0, 0
		for n < len(s) && n < limit && '0' <= s[n] && s[n] <= '7' {
			v = v*8 + int(s[n]-'0')
			n++
		}
		end := string([]byte{byte(v)})
		c.prog.LineEnd = &end
		return s[n:], true, nil
	}
	// No number: the output separator is the input separator as it stands
	// now.
	end := "\n"
	switch {
	case c.rsUndef:
		end = ""
	case c.rsSet && c.rs == "":
		end = "\n\n"
	case c.rsSet:
		end = c.rs
	}
	c.prog.LineEnd = &end
	return s, true, nil
}

// debugger is -d: s is what follows the 'd'.
func (c *perlCommand) debugger(s string) (rest string, more bool, err error) {
	if strings.HasPrefix(s, "t") && (len(s) == 1 || !isWordChar(s[1])) {
		s = s[1:]
		c.env = append(c.env, "PERL5DB_THREADED=1")
	}
	if s != "" && (s[0] == ':' || s[0] == '=') {
		s = s[1:]
		stmt := "use Devel::"
		if strings.HasPrefix(s, "-") {
			s = s[1:]
			stmt = "no Devel::"
		}
		n := 0
		for n < len(s) && (isWordChar(s[n]) || s[n] == ':') {
			n++
		}
		if n < len(s) && s[n] == '=' {
			stmt += s[:n] + " split(/,/,q{" + s[n+1:] + "});"
		} else {
			stmt += s
		}
		c.env = append(c.env, "PERL5DB="+stmt)
		s = ""
	}
	c.prog.Debug = true
	return s, true, nil
}

// useModule is -M and -m: the statement that loads the module is compiled
// ahead of the program.
func (c *perlCommand) useModule(option byte, s string) error {
	use := "use "
	if strings.HasPrefix(s, "-") {
		use = " no "
		s = s[1:]
	}
	n := 0
	colon := false
	for n < len(s) && (isWordChar(s[n]) || s[n] == ':') {
		if s[n] == ':' {
			if n+1 < len(s) && s[n+1] == ':' {
				n++
			} else {
				colon = true
			}
		}
		n++
	}
	if n == 0 {
		return fmt.Errorf("Module name required with -%c option", option)
	}
	if colon {
		return fmt.Errorf("Invalid module name %s with -%c option: contains single ':'", s[:n], option)
	}
	var stmt string
	switch {
	case n < len(s) && s[n] == '=':
		// -MModule=a,b: the import list, split on commas. NUL delimits
		// the list so that it may contain any printable character.
		stmt = use + s[:n] + " split(/,/,q\x00" + s[n+1:] + "\x00)"
	case option == 'm':
		if n < len(s) {
			return fmt.Errorf("Can't use '%c' after -mname", s[n])
		}
		stmt = use + s + " ()"
	default:
		// -M'Module qw(a b)': the text is the rest of the statement.
		stmt = use + s
	}
	c.prog.Preamble = append(c.prog.Preamble, stmt)
	return nil
}

// The -C letters.
var unicodeLetters = map[byte]uint32{
	'I': perl.UnicodeStdin,
	'O': perl.UnicodeStdout,
	'E': perl.UnicodeStderr,
	'S': perl.UnicodeStdin | perl.UnicodeStdout | perl.UnicodeStderr,
	'i': perl.UnicodeIn,
	'o': perl.UnicodeOut,
	'D': perl.UnicodeIn | perl.UnicodeOut,
	'A': perl.UnicodeArgv,
	'L': perl.UnicodeLocale,
	'a': perl.UnicodeUTF8CacheAssert,
}

const (
	// -C without a value.
	unicodeDefault = perl.UnicodeStdin | perl.UnicodeStdout | perl.UnicodeStderr |
		perl.UnicodeIn | perl.UnicodeOut | perl.UnicodeLocale
	unicodeAll = 0x1ff
)

// parseUnicodeOptions reads the value of -C: a number or a list of
// letters, up to the end of s or the first whitespace.
func parseUnicodeOptions(s string) (flags uint32, rest string, err error) {
	switch {
	case s == "":
		return unicodeDefault, "", nil
	case '0' <= s[0] && s[0] <= '9':
		n := 0
		for n < len(s) && '0' <= s[n] && s[n] <= '9' {
			n++
		}
		v, perr := strconv.ParseUint(s[:n], 10, 32)
		if perr != nil || n > 1 && s[0] == '0' {
			return 0, "", fmt.Errorf("Invalid number '%s' for -C option.", s)
		}
		flags = uint32(v)
		rest = s[n:]
		if rest != "" && rest[0] != '\n' && rest[0] != '\r' && !isPerlSpace(rest[0]) {
			return 0, "", fmt.Errorf("Unknown Unicode option letter '%c'", rest[0])
		}
	default:
		i := 0
	letters:
		for ; i < len(s); i++ {
			bit, ok := unicodeLetters[s[i]]
			switch {
			case ok:
				flags |= bit
			case s[i] == '\n' || s[i] == '\r':
			case isPerlSpace(s[i]):
				break letters
			default:
				return 0, "", fmt.Errorf("Unknown Unicode option letter '%c'", s[i])
			}
		}
		rest = s[i:]
	}
	if extra := flags &^ unicodeAll; extra != 0 {
		return 0, "", fmt.Errorf("Unknown Unicode option value %d", extra)
	}
	return flags, rest, nil
}

// applyPERL5OPT applies the switches of the PERL5OPT environment variable,
// which perl reads after its command line. Each word holds one switch.
func (c *perlCommand) applyPERL5OPT(opt string) error {
	if c.prog.Taint == perl.TaintOn {
		return nil // ignored under taint checks
	}
	s := strings.TrimLeftFunc(opt, func(r rune) bool { return r < 0x80 && isPerlSpace(byte(r)) })
	if strings.HasPrefix(s, "-T") {
		c.prog.Taint = perl.TaintOn
		return nil
	}
	for s != "" {
		for s != "" && isPerlSpace(s[0]) {
			s = s[1:]
		}
		if strings.HasPrefix(s, "-") {
			s = s[1:]
			if s != "" && isPerlSpace(s[0]) {
				continue
			}
		}
		if s == "" {
			break
		}
		if !strings.ContainsRune("CDIMUdmtwW", rune(s[0])) {
			return fmt.Errorf("Illegal switch in PERL5OPT: -%c", s[0])
		}
		var word string
		word, s = untilSpace(s)
		if word[0] == 't' {
			if c.prog.Taint == perl.TaintOff {
				c.prog.Taint = perl.TaintWarn
			}
			continue
		}
		if _, _, err := c.moreSwitches(word); err != nil {
			return err
		}
	}
	return nil
}

// takeDefines strips the -name and -name=value arguments that lead args
// (-s) and records them.
func (c *perlCommand) takeDefines(args []string) []string {
	for len(args) > 0 && strings.HasPrefix(args[0], "-") {
		a := args[0]
		if a == "-" {
			break
		}
		args = args[1:]
		if a == "--" {
			break
		}
		if name, value, ok := strings.Cut(a[1:], "="); ok {
			c.prog.Defines = append(c.prog.Defines, perl.Define{Name: name, Value: &value})
		} else {
			c.prog.Defines = append(c.prog.Defines, perl.Define{Name: name})
		}
	}
	return args
}

// searchProgram is -S: a program name without a directory is looked up in
// the directories of PATH.
func searchProgram(name, path string) (string, error) {
	if strings.ContainsRune(name, '/') {
		return name, nil
	}
	for _, dir := range filepath.SplitList(path) {
		if dir == "" {
			dir = "."
		}
		candidate := filepath.Join(dir, name)
		if st, err := os.Stat(candidate); err == nil && st.Mode().IsRegular() {
			return candidate, nil
		}
	}
	return "", fmt.Errorf("Can't find %s on PATH", name)
}

// perlUsage is perl -h's text.
const perlUsage = `
Usage: %s [switches] [--] [programfile] [arguments]
  -0[octal/hexadecimal] specify record separator (\0, if no argument)
  -a                    autosplit mode with -n or -p (splits $_ into @F)
  -C[number/list]       enables the listed Unicode features
  -c                    check syntax only (runs BEGIN and CHECK blocks)
  -d[t][:MOD]           run program under debugger or module Devel::MOD
  -D[number/letters]    set debugging flags (argument is a bit mask or alphabets)
  -e commandline        one line of program (several -e's allowed, omit programfile)
  -E commandline        like -e, but enables all optional features
  -f                    don't do $sitelib/sitecustomize.pl at startup
  -F/pattern/           split() pattern for -a switch (//'s are optional)
  -g                    read all input in one go (slurp), rather than line-by-line (alias for -0777)
  -i[extension]         edit <> files in place (makes backup if extension supplied)
  -Idirectory           specify @INC/#include directory (several -I's allowed)
  -l[octnum]            enable line ending processing, specifies line terminator
  -[mM][-]module        execute "use/no module..." before executing program
  -n                    assume "while (<>) { ... }" loop around program
  -p                    assume loop like -n but print line also, like sed
  -s                    enable rudimentary parsing for switches after programfile
  -S                    look for programfile using PATH environment variable
  -t                    enable tainting warnings
  -T                    enable tainting checks
  -U                    allow unsafe operations
  -v                    print version, patchlevel and license
  -V[:configvar]        print configuration summary (or a single Config.pm variable)
  -w                    enable many useful warnings
  -W                    enable all warnings
  -x[directory]         ignore text before #!perl line (optionally cd to directory)
  -X                    disable all warnings

Run 'perldoc perl' for more help with Perl.

`

// perlVersionProgram prints perl -v's text for the embedded interpreter.
var perlVersionProgram = []string{
	`use Config;`,
	`my ($revision, $version, $subversion) = (sprintf('%vd', $^V) =~ /(\d+)\.(\d+)\.(\d+)/);`,
	`print "\nThis is perl $revision, version $version, subversion $subversion ($^V) built for $Config{archname}\n\n";`,
	`print "Copyright 1987-2026, Larry Wall\n\n";`,
	`print "Perl may be copied only under the terms of either the Artistic License or the\n";`,
	`print "GNU General Public License, which may be found in the Perl 5 source kit.\n\n";`,
	`print "Complete documentation for Perl, including FAQ lists, should be found on\n";`,
	`print "this system using \"man perl\" or \"perldoc perl\".  If you have access to the\n";`,
	`print "Internet, point your browser at https://www.perl.org/, the Perl Home Page.\n\n";`,
}
