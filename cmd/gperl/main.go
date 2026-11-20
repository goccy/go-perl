// gperl is the perl-meets-go command for go-perl: it runs Perl programs on
// the embedded (wasm2go-transpiled) interpreter with go-style tooling.
//
//	gperl run script.pl [args...]   sync dependencies, then run the script
//	gperl run [switches] [script]   with any perl(1) switch (-e/-I/-M/-n/...),
//	                                or a program on standard input, run
//	                                exactly like the perl command instead
//	gperl build [-o out] script.pl  produce a self-contained Go binary that
//	                                embeds the script, its CPAN dependencies
//	                                (native XS modules linked in), and the
//	                                interpreter
//	gperl test [files]              run a project's test files
//	gperl psgi run app.psgi         serve a PSGI application
//	gperl psgi build app.psgi       build its server binary
//	gperl mod tidy                  resolve the cpanfile into
//	                                cpanfile.snapshot, then fetch and build
//	gperl mod download              fetch and build what the snapshot names
//	gperl get module[@version]...   add modules to the cpanfile and resolve
//	gperl clean -modcache           remove the module directory
//	gperl install toolchain         install the Go and C toolchains gperl
//	                                builds with
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
  gperl test [options] [files or directories]
  gperl psgi run [options] app.psgi
  gperl psgi build [-o out] [options] app.psgi
  gperl mod download
  gperl mod tidy
  gperl get module[@version]...
  gperl clean -modcache
  gperl install toolchain

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
	case "test":
		status, err := gperl.Test(os.Args[2:])
		exitOn(err)
		os.Exit(status)
	case "psgi":
		if len(os.Args) < 3 {
			usage()
		}
		switch os.Args[2] {
		case "run":
			exitOn(gperl.PSGIRun(os.Args[3:]))
		case "build":
			exitOn(gperl.PSGIBuild(os.Args[3:]))
		default:
			usage()
		}
	case "install":
		if len(os.Args) != 3 || os.Args[2] != "toolchain" {
			usage()
		}
		exitOn(gperl.InstallToolchain())
	case "mod":
		if len(os.Args) != 3 {
			usage()
		}
		switch os.Args[2] {
		case "tidy":
			exitOn(gperl.ModTidy("."))
		case "download":
			exitOn(gperl.ModDownload("."))
		default:
			usage()
		}
	case "get":
		if len(os.Args) < 3 {
			usage()
		}
		exitOn(gperl.Get(".", os.Args[2:]...))
	case "clean":
		if len(os.Args) != 3 || os.Args[2] != "-modcache" {
			usage()
		}
		exitOn(gperl.CleanModCache())
	default:
		usage()
	}
}

func exitOn(err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "gperl: %v\n", err)
		os.Exit(1)
	}
}
