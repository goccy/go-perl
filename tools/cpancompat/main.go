// cpancompat measures how much of CPAN runs unmodified on go-perl.
//
// The work is split into steps that each read and write plain files in a
// data directory (outside the repository), so every step can be re-run or
// resumed on its own:
//
//	cpancompat fetch   download the CPAN package index and the MetaCPAN
//	                   metadata of every latest release
//	cpancompat graph   join them into one record per distribution and
//	                   derive the dependency graph
//
// The data directory is given with -data or the CPANCOMPAT_DATA
// environment variable.
package main

import (
	"flag"
	"fmt"
	"os"
)

func usage() {
	fmt.Fprintf(os.Stderr, `usage:
  cpancompat fetch [-data dir]
  cpancompat graph [-data dir]
`)
	os.Exit(2)
}

func main() {
	if len(os.Args) < 2 {
		usage()
	}
	fs := flag.NewFlagSet(os.Args[1], flag.ExitOnError)
	data := fs.String("data", os.Getenv("CPANCOMPAT_DATA"), "data directory")
	_ = fs.Parse(os.Args[2:])
	if *data == "" {
		fmt.Fprintln(os.Stderr, "cpancompat: no data directory (-data or CPANCOMPAT_DATA)")
		os.Exit(2)
	}
	var err error
	switch os.Args[1] {
	case "fetch":
		err = fetch(*data)
	case "graph":
		err = buildGraph(*data)
	default:
		usage()
	}
	if err != nil {
		fmt.Fprintf(os.Stderr, "cpancompat: %v\n", err)
		os.Exit(1)
	}
}
