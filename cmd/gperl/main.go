// gperl is the perl-meets-go command for go-perl: it runs Perl programs on
// the embedded (wasm2go-transpiled) interpreter with go-style tooling.
//
//	gperl run script.pl [args...]   resolve dependencies, then run the script
//	gperl run [switches] [script]   with any perl(1) switch (-e/-I/-M/-n/...),
//	                                or a program on standard input, run
//	                                exactly like the perl command instead
//	gperl build [-o out] script.pl  produce a self-contained Go binary that
//	                                embeds the script, its vendored modules,
//	                                and the interpreter
//	gperl xs build dist...          compile XS distributions (source dir or
//	                                .tar.gz) against the native XS SDK into
//	                                ./local/xs; gperl run and built binaries
//	                                load them automatically
//
// GPERL_MAX_MEMORY limits the interpreter's memory, for gperl run and for
// the binaries gperl build makes: a size in bytes (with K, M, G or T), or
// "host" for the memory the host gives the process; 8 GiB when unset.
//
// Every subcommand is a thin wrapper over the github.com/goccy/go-perl/gperl
// library, so applications can drive the same functionality programmatically.

// A program gperl runs is the process and has its descriptor numbers, the
// lowest free one first: its first file is descriptor 3. On Linux the Go
// runtime otherwise keeps the file of the process's CPU limit open from its
// own start, under that number, to follow the limit while the process runs.
// With this setting it reads the limit of the control group at no time and
// sizes itself by the CPUs of the machine, as it did before Go 1.25.
//
//go:debug containermaxprocs=0
package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/goccy/go-perl/gperl"
)

func usage() {
	fmt.Fprintf(os.Stderr, `usage:
  gperl run script.pl [args...]
  gperl run [perl switches] [script] [args...]
  gperl build [-o out] script.pl
  gperl xs build [-C dir] dist [dist...]

environment:
  GPERL_MAX_MEMORY   the interpreter's memory limit: bytes (with K, M, G or T),
                     or host for the host's memory; 8G when unset
`)
	os.Exit(2)
}

func main() {
	if len(os.Args) < 2 {
		usage()
	}
	switch os.Args[1] {
	case "run":
		status, err := gperl.RunCLI(os.Args[2:])
		if err != nil {
			gperl.PrintRunError(err)
		}
		os.Exit(status)
	case "build":
		fs := flag.NewFlagSet("build", flag.ExitOnError)
		out := fs.String("o", "", "output binary path")
		_ = fs.Parse(os.Args[2:])
		if fs.NArg() != 1 {
			usage()
		}
		if err := gperl.Build(fs.Arg(0), *out); err != nil {
			fmt.Fprintf(os.Stderr, "gperl: %v\n", err)
			os.Exit(1)
		}
	case "xs":
		if len(os.Args) < 4 || os.Args[2] != "build" {
			usage()
		}
		fs := flag.NewFlagSet("xs build", flag.ExitOnError)
		dir := fs.String("C", ".", "project directory (module output goes to its local/xs)")
		_ = fs.Parse(os.Args[3:])
		if fs.NArg() == 0 {
			usage()
		}
		modules, err := gperl.XSBuild(*dir, fs.Args())
		for _, m := range modules {
			fmt.Fprintf(os.Stderr, "gperl: built native module %s\n", m)
		}
		if err != nil {
			fmt.Fprintf(os.Stderr, "gperl: %v\n", err)
			os.Exit(1)
		}
	default:
		usage()
	}
}
