// gperl is the perl-meets-go command for go-perl: it runs Perl programs on
// the embedded (wasm2go-transpiled) interpreter with go-style tooling.
//
//	gperl run script.pl [args...]   resolve dependencies, then run the script
//	gperl build [-o out] script.pl  produce a self-contained Go binary that
//	                                embeds the script, its vendored modules,
//	                                and the interpreter
//	gperl get Module...             vendor CPAN modules into ./local (cpanm)
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
  gperl build [-o out] script.pl
  gperl get Module [Module...]
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
		status, err := gperl.Run(os.Args[2], os.Args[3:])
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
	case "get":
		if len(os.Args) < 3 {
			usage()
		}
		wd, err := os.Getwd()
		if err == nil {
			err = gperl.Get(wd, os.Args[2:])
		}
		if err != nil {
			fmt.Fprintf(os.Stderr, "gperl: %v\n", err)
			os.Exit(1)
		}
	default:
		usage()
	}
}
