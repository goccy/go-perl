//go:build darwin || linux

package gperl_test

import (
	"fmt"
	"runtime"
	"strconv"
	"strings"
	"testing"
)

// A constant of an extension has one number for the program, the system's,
// whatever way the program reaches it: a call compiled before or after the
// module is loaded, a reference taken before it is loaded, ->can, a
// symbolic name, an import at compile time or at run time, in a fork child,
// and whichever module was loaded first.

// numberFamilies are the families of numbers the interpreter keeps a table
// of the system's for (Internals::host_numbers).
var numberFamilies = []string{
	"errno", "signal", "open", "poll", "af", "socktype", "sockflag", "sol", "so", "ip", "ipv6", "tcp",
	"msg", "ai", "ni", "eai", "lc", "sighow", "sa", "fcntl", "ioctl", "langinfo",
}

// constantPackage is a package that holds constants and the module that
// loads it.
type constantPackage struct {
	pkg    string
	module string
}

var constantPackages = []constantPackage{
	{"Fcntl", "Fcntl"},
	{"POSIX", "POSIX"},
	{"Socket", "Socket"},
	{"IO::Poll", "IO::Poll"},
	{"IO::Handle", "IO::Handle"},
	{"IO::Seekable", "IO::Seekable"},
	{"IO::Socket", "IO::Socket"},
	{"Errno", "Errno"},
	{"I18N::Langinfo", "I18N::Langinfo"},
	{"Sys::Syslog", "Sys::Syslog"},
	{"Time::HiRes", "Time::HiRes"},
	{"File::Glob", "File::Glob"},
}

// systemNumbers returns the system's number of every name of every family,
// as the interpreter's tables have them, and the names by family. A family
// the system has no numbers of has no names (the flags or-ed into the type
// of a socket, on darwin).
func systemNumbers(t *testing.T) (numbers map[string]int64, families map[string][]string) {
	t.Helper()
	out, status := hostPerl(t, append([]string{"-e", `for my $family (@ARGV) {
	my @table = Internals::host_numbers($family);
	while (my ($name, $number) = splice(@table, 0, 2)) { print "$family $name $number\n" }
}`}, numberFamilies...)...)
	if status != 0 {
		t.Fatalf("status %d", status)
	}
	numbers = map[string]int64{}
	families = map[string][]string{}
	for _, line := range strings.Split(strings.TrimSpace(out), "\n") {
		fields := strings.Fields(line)
		if len(fields) != 3 {
			t.Fatalf("line %q of the tables", line)
		}
		number, err := strconv.ParseInt(fields[2], 10, 64)
		if err != nil {
			t.Fatalf("line %q of the tables", line)
		}
		numbers[fields[1]] = number
		families[fields[0]] = append(families[fields[0]], fields[1])
	}
	if len(numbers) < 300 || len(families) < len(numberFamilies)-1 {
		t.Fatalf("the tables have %d names in %d families", len(numbers), len(families))
	}
	return numbers, families
}

// printedNumbers runs a program that prints "name number" lines, as
// perlNumbers does, for a program that may have nothing to print.
func printedNumbers(t *testing.T, program string) map[string]int64 {
	t.Helper()
	return perlNumbers(t, program+"\n;print \"- 0\\n\";")
}

// constantsOf returns the constants with an integer value a package has
// once its module is loaded, by name.
func constantsOf(t *testing.T, p constantPackage) map[string]int64 {
	t.Helper()
	constants := printedNumbers(t, `require `+p.module+`; require B;
my $stash = \%{"`+p.pkg+`::"};
for my $name (sort keys %$stash) {
	next unless $name =~ /\A[A-Za-z_][A-Za-z0-9_]*\z/;
	my $constant = ref(\$stash->{$name}) eq 'GLOB'
		? (defined *{$stash->{$name}}{CODE} && (B::svref_2object(*{$stash->{$name}}{CODE})->CvFLAGS & B::CVf_CONST()))
		: ref($stash->{$name}) eq 'SCALAR';
	next unless $constant;
	my $value = `+p.pkg+`->can($name)->();
	print "$name $value\n" if defined $value && !ref $value && $value =~ /\A-?[0-9]+\z/;
}`)
	delete(constants, "-")
	return constants
}

// accessPath is one way a program reaches the constants of a package: the
// program prints "name number" lines for the names it could reach that way.
type accessPath struct {
	name string
	// every says that each constant of the package is reachable this way
	every   bool
	program func(p constantPackage, names []string) string
}

// each returns the statement of the template for every name, joined.
func each(names []string, template string, p constantPackage) string {
	var b strings.Builder
	for _, name := range names {
		b.WriteString(strings.NewReplacer("NAME", name, "PKG", p.pkg).Replace(template))
		b.WriteString("\n")
	}
	return b.String()
}

func otherModules(p constantPackage) (others []string) {
	for _, other := range constantPackages {
		if other.module != p.module {
			others = append(others, "require "+other.module+";")
		}
	}
	return others
}

var accessPaths = []accessPath{
	{"a call compiled after the module was loaded at compile time", true, func(p constantPackage, names []string) string {
		return "use " + p.module + " ();\n" + each(names, `print "NAME ", PKG::NAME(), "\n";`, p)
	}},
	{"a call compiled before the module is loaded", true, func(p constantPackage, names []string) string {
		return "require " + p.module + ";\n" + each(names, `print "NAME ", PKG::NAME(), "\n";`, p)
	}},
	{"a reference taken before the module is loaded", true, func(p constantPackage, names []string) string {
		return "my %taken = (\n" + each(names, `NAME => \&PKG::NAME,`, p) + ");\nrequire " + p.module +
			";\nprint \"$_ \", $taken{$_}->(), \"\\n\" for sort keys %taken;\n"
	}},
	{"can", true, func(p constantPackage, names []string) string {
		return "require " + p.module + ";\n" + each(names, `print "NAME ", PKG->can("NAME")->(), "\n";`, p)
	}},
	{"a symbolic name", true, func(p constantPackage, names []string) string {
		return "require " + p.module + ";\n" + each(names, `print "NAME ", &{"PKG::NAME"}(), "\n";`, p)
	}},
	{"code compiled at run time", true, func(p constantPackage, names []string) string {
		return "require " + p.module + ";\n" + each(names, `print "NAME ", eval('PKG::NAME()'), "\n";`, p)
	}},
	{"the value the package holds before the name is used", true, func(p constantPackage, names []string) string {
		// the entry of the package is the constant's value behind a
		// reference, or the sub
		return "require " + p.module + ";\n" + each(names,
			`{ my $entry = $PKG::{NAME}; print "NAME ", (ref($entry) eq 'SCALAR' ? $$entry : *{$entry}{CODE}->()), "\n"; }`, p)
	}},
	{"names imported at compile time", false, func(p constantPackage, names []string) string {
		return "use " + p.module + ";\n" +
			each(names, `{ my $v = eval('no strict; NAME()'); print "NAME $v\n" if defined $v; }`, p)
	}},
	{"names imported at run time", false, func(p constantPackage, names []string) string {
		// the names the package exports: the import of another name fails
		return "require " + p.module + ";\n" +
			each(names, `if (eval { PKG->import("NAME"); 1 } && main->can("NAME")) { print "NAME ", main->can("NAME")->(), "\n"; }`, p)
	}},
	{"a call compiled before an import at run time", true, func(p constantPackage, names []string) string {
		return "require " + p.module + ";\n" + p.module + "->import;\n" +
			each(names, `print "NAME ", PKG::NAME(), "\n";`, p)
	}},
	{"the module loaded after the others", true, func(p constantPackage, names []string) string {
		return strings.Join(otherModules(p), " ") + "\nrequire " + p.module + ";\n" +
			each(names, `print "NAME ", PKG::NAME(), "\n";`, p)
	}},
	{"the module loaded before the others", true, func(p constantPackage, names []string) string {
		return "require " + p.module + ";\n" + strings.Join(otherModules(p), " ") + "\n" +
			each(names, `print "NAME ", PKG::NAME(), "\n";`, p)
	}},
	{"a fork child, the module loaded by the parent", true, func(p constantPackage, names []string) string {
		return "$| = 1; require " + p.module + ";\nmy $pid = fork // die \"fork: $!\";\nif (!$pid) {\n" +
			each(names, `print "NAME ", PKG::NAME(), "\n";`, p) + "exit 0;\n}\nwaitpid($pid, 0); exit($? >> 8);\n"
	}},
	{"a fork child that loads the module", true, func(p constantPackage, names []string) string {
		return "$| = 1; my $pid = fork // die \"fork: $!\";\nif (!$pid) {\nrequire " + p.module + ";\n" +
			each(names, `print "NAME ", PKG::NAME(), "\n";`, p) + "exit 0;\n}\nwaitpid($pid, 0); exit($? >> 8);\n"
	}},
}

func TestConstantsAreTheSystemsOnEveryPath(t *testing.T) {
	system, families := systemNumbers(t)
	covered := map[string]bool{}
	for _, p := range constantPackages {
		constants := constantsOf(t, p)
		names := sortedNames(constants)
		// The number a constant must have: the system's for a name the
		// tables have, and for any other name one number on every path.
		want := map[string]int64{}
		for _, name := range names {
			want[name] = constants[name]
			if number, ok := system[name]; ok {
				want[name] = number
				covered[name] = true
			}
		}
		for _, path := range accessPaths {
			t.Run(p.pkg+"/"+path.name, func(t *testing.T) {
				if len(names) == 0 {
					t.Skip("the package has no integer constants")
				}
				got := printedNumbers(t, path.program(p, names))
				delete(got, "-")
				if path.every && len(got) != len(names) {
					t.Errorf("%d of the %d constants were printed", len(got), len(names))
				}
				var wrong []string
				for _, name := range sortedNames(got) {
					if got[name] != want[name] {
						wrong = append(wrong, fmt.Sprintf("%s = %d, want %d", name, got[name], want[name]))
					}
				}
				if len(wrong) > 0 {
					shown := wrong
					if len(shown) > 12 {
						shown = shown[:12]
					}
					t.Errorf("%d of %d constants have another number:\n\t%s", len(wrong), len(got), strings.Join(shown, "\n\t"))
				}
			})
		}
	}
	// The families the constants were of. No package has a constant for
	// the requests of ioctl: a program writes those as numbers.
	t.Run("families", func(t *testing.T) {
		for _, family := range numberFamilies {
			table := families[family]
			found := 0
			for _, name := range table {
				if covered[name] {
					found++
				}
			}
			if found == 0 && len(table) > 0 && family != "ioctl" {
				t.Errorf("no constant of a package has a name of the family %s (%d names)", family, len(table))
			}
			t.Logf("%s: %d of the system's %d names are constants of a package", family, found, len(table))
		}
	})
}

// errorsHeldAsAnother are the errors $! cannot hold under their own number:
// the interpreter's C library has one number for ENOTSUP and EOPNOTSUPP,
// darwin has two, and $! reads back the first.
var errorsHeldAsAnother = map[string]map[string]bool{
	"darwin": {"EOPNOTSUPP": true},
}

// The constants of Errno and of POSIX are the system's numbers of the
// errors, and $! holds an error under that number: %! knows it by the name,
// and POSIX::errno reports it.
func TestErrorNamesOnEveryPath(t *testing.T) {
	system := perlNumbers(t, `my @table = Internals::host_numbers("errno"); while (my ($name, $number) = splice(@table, 0, 2)) { print "$name $number\n" }`)
	names := sortedNames(system)
	var program strings.Builder
	program.WriteString("require Errno; require POSIX;\n")
	for _, name := range names {
		number := strconv.FormatInt(system[name], 10)
		program.WriteString(strings.NewReplacer("NAME", name, "NUMBER", number).Replace(
			`{ if (Errno->can("NAME")) { print "Errno:NAME ", Errno::NAME(), "\n"; }
  if (defined(my $posix = eval { POSIX::NAME() })) { print "POSIX:NAME $posix\n"; }
  print "exists:NAME ", (exists $!{NAME} ? NUMBER : -1), "\n";
  local $! = NUMBER;
  print "held:NAME ", $! + 0, "\n";
  print "hash:NAME ", ($!{NAME} ? NUMBER : -1), "\n";
  print "errno:NAME ", POSIX::errno(), "\n"; }
`))
	}
	got := perlNumbers(t, program.String())
	seen := map[string]int{}
	for _, key := range sortedNames(got) {
		path, name, _ := strings.Cut(key, ":")
		seen[path]++
		heldAsAnother := errorsHeldAsAnother[runtime.GOOS][name] && (path == "held" || path == "hash" || path == "errno")
		if got[key] != system[name] && !heldAsAnother {
			t.Errorf("%s of %s = %d, want %d", path, name, got[key], system[name])
		}
	}
	for _, path := range []string{"Errno", "exists", "held", "hash", "errno"} {
		if seen[path] != len(names) {
			t.Errorf("%s: %d of the %d names", path, seen[path], len(names))
		}
	}
	if seen["POSIX"] < 60 {
		t.Errorf("POSIX has %d of the names", seen["POSIX"])
	}
}
