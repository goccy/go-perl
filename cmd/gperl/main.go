// gperl is the perl-meets-go command for go-perl: it runs Perl programs on
// the embedded (wasm2go-transpiled) interpreter with go-style tooling.
//
//	gperl run script.pl [args...]   resolve dependencies, then run the script
//	gperl run [-e code|switches]    with any perl(1) switch (-e/-I/-M/-l/...),
//	                                run exactly like the perl command instead
//	gperl build [-o out] script.pl  produce a self-contained Go binary that
//	                                embeds the script, its vendored modules,
//	                                and the interpreter
//	gperl xs build dist...          compile XS distributions (source dir or
//	                                .tar.gz) against the native XS SDK into
//	                                ./local/xs; gperl run and built binaries
//	                                load them automatically
//
// Every subcommand is a thin wrapper over the github.com/goccy/go-perl/gperl
// library, so applications can drive the same functionality programmatically.
package main

import (
	"errors"
	"flag"
	"fmt"
	"os"

	perl "github.com/goccy/go-perl"
	"github.com/goccy/go-perl/gperl"
)

func usage() {
	fmt.Fprintf(os.Stderr, `usage:
  gperl run script.pl [args...]
  gperl run [perl switches] [script] [args...]
  gperl build [-o out] script.pl
  gperl xs build [-C dir] dist [dist...]
`)
	os.Exit(2)
}

func main() {
	if len(os.Args) < 2 {
		usage()
	}
	switch os.Args[1] {
	case "run":
		if len(os.Args) < 3 {
			usage()
		}
		status, err := gperl.RunCLI(os.Args[2:])
		var pe *perl.PerlError
		if errors.As(err, &pe) {
			msg := pe.Message
			if len(msg) == 0 || msg[len(msg)-1] != '\n' {
				msg += "\n"
			}
			fmt.Fprint(os.Stderr, msg)
		} else if err != nil {
			fmt.Fprintf(os.Stderr, "gperl: %v\n", err)
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
