// cpancompat measures how much of CPAN runs unmodified on go-perl.
//
// The work is split into steps that each read and write plain files in a
// data directory (outside the repository), so every step can be re-run or
// resumed on its own:
//
//	cpancompat fetch   download the CPAN package index and the MetaCPAN
//	                   metadata of every latest release
//	cpancompat testers download the CPAN Testers reports made on one perl
//	                   version (-perl), the version the interpreter is
//	                   built from
//	cpancompat graph   join them into one record per distribution and
//	                   derive the dependency graph (-perl: with the reports
//	                   each release got on that perl version)
//	cpancompat native  find the distributions with native (XS) code among
//	                   the most depended-upon ones (-top N)
//	cpancompat image   build the container image distributions are tested
//	                   in, around a gperl binary built for linux, or the
//	                   same environment around an official perl image (the
//	                   reference the interpreter is compared with)
//	cpancompat run     configure, build, test and install distributions on
//	                   go-perl, in dependency order, one container each;
//	                   results accumulate in runs/<name>/results.jsonl and
//	                   an interrupted run continues where it stopped
//	cpancompat report  summarize a run: outcomes, and coverage against the
//	                   distributions that pass on perl
//
// The data directory is given with -data or the CPANCOMPAT_DATA
// environment variable.
package main

import (
	"flag"
	"fmt"
	"os"
	"strings"
	"time"
)

func usage() {
	fmt.Fprintf(os.Stderr, `usage:
  cpancompat fetch  [-data dir]
  cpancompat testers [-data dir] -perl version
  cpancompat graph  [-data dir] [-perl version]
  cpancompat native [-data dir] [-top n]
  cpancompat image  [-data dir] (-gperl binary | -perl-image image) [-image tag]
  cpancompat run    [-data dir] -run name [-image tag] [-top n | -dists a,b,...]
                    [-jobs n] [-cpus n] [-memory size] [-wall duration]
                    [-interpreter label] [-redo outcome,...]
  cpancompat report [-data dir] -run name
`)
	os.Exit(2)
}

func main() {
	if len(os.Args) < 2 {
		usage()
	}
	fs := flag.NewFlagSet(os.Args[1], flag.ExitOnError)
	data := fs.String("data", os.Getenv("CPANCOMPAT_DATA"), "data directory")
	top := fs.Int("top", 0, "native, run: only the n most depended-upon distributions (native: default 1000; run: default all)")
	gperl := fs.String("gperl", "", "image: the gperl binary, built for linux")
	perl := fs.String("perl", "", "testers, graph: the perl version the interpreter is built from, as CPAN Testers names it (e.g. 5.44.0)")
	perlImage := fs.String("perl-image", "", "image: build the reference image around this official perl image instead of gperl")
	image := fs.String("image", "cpancompat:latest", "image, run: the container image tag")
	run := fs.String("run", "", "run, report: the name of the run")
	only := fs.String("dists", "", "run: only these distributions (comma separated), with what they need")
	jobs := fs.Int("jobs", 2, "run: how many distributions to run at once")
	cpus := fs.String("cpus", "2", "run: CPUs of each container")
	memory := fs.String("memory", "2g", "run: memory of each container")
	wall := fs.Duration("wall", 2*time.Hour, "run: how long one container may run")
	interpreter := fs.String("interpreter", "", "run: a label for the interpreter build, kept in every result")
	redo := fs.String("redo", "", "run: outcomes to run again (comma separated) instead of keeping their recorded result")
	_ = fs.Parse(os.Args[2:])
	if *data == "" {
		fmt.Fprintln(os.Stderr, "cpancompat: no data directory (-data or CPANCOMPAT_DATA)")
		os.Exit(2)
	}
	var err error
	switch os.Args[1] {
	case "fetch":
		err = fetch(*data)
	case "testers":
		err = fetchTesters(*data, *perl)
	case "graph":
		err = buildGraph(*data, *perl)
	case "native":
		if *top == 0 {
			*top = 1000
		}
		err = scanNative(*data, *top)
	case "image":
		err = buildImage(*data, *gperl, *perlImage, *image)
	case "run":
		if *run == "" {
			usage()
		}
		opts := runOptions{
			dataDir: *data, run: *run, image: *image, interpreter: *interpreter,
			top: *top, jobs: *jobs, cpus: *cpus, memory: *memory, wall: *wall,
			redo: map[string]bool{},
		}
		if *only != "" {
			opts.only = strings.Split(*only, ",")
		}
		if *redo != "" {
			for _, outcome := range strings.Split(*redo, ",") {
				opts.redo[outcome] = true
			}
		}
		err = runSurvey(opts)
	case "report":
		if *run == "" {
			usage()
		}
		err = report(*data, *run, os.Stdout)
	default:
		usage()
	}
	if err != nil {
		fmt.Fprintf(os.Stderr, "cpancompat: %v\n", err)
		os.Exit(1)
	}
}
