package gperl_test

import (
	"os"
	"os/exec"
	"testing"
)

// %Config names the C toolchain of the machine the program runs on, as a
// perl's does: a program that runs the preprocessor or the compiler
// (Filter::cpp runs $Config{cppstdin}, Inline::C and Affix::Build
// $Config{cc}) finds them. They named the compiler the interpreter was
// built with, on the machine it was built on (wasi-sdk's clang, with
// --target=wasm64-wasip1). What is expected is what perl gives on the same
// machine.
func TestConfigNamesTheHostsToolchain(t *testing.T) {
	if _, err := exec.LookPath("cc"); err != nil {
		t.Skip("no host C compiler available")
	}
	dir := t.TempDir()
	program := `use Config;
my $dir = shift;
my $out = qx{echo ANSWER | $Config{cppstdin} -DANSWER=42 $Config{cppminus} 2>/dev/null};
print $out =~ /^42$/m ? "cpp 42\n" : "cpp: $out\n";
open my $c, '>', "$dir/t.c" or die "t.c: $!";
print $c qq{#include <stdio.h>\nint main(void) { puts("hello from cc"); return 0; }\n};
close $c;
system("$Config{cc} $Config{ccflags} $Config{optimize} -o $dir/t $dir/t.c") == 0 or die "cc: $?";
print qx{$dir/t};`
	cmd := exec.Command(os.Args[0], "-e", program, dir)
	cmd.Env = append(os.Environ(), asPerlEnv+"=1")
	out, err := cmd.CombinedOutput()
	if want := "cpp 42\nhello from cc\n"; err != nil || string(out) != want {
		t.Fatalf("err %v, output:\n%s\nwant:\n%s", err, out, want)
	}
}
