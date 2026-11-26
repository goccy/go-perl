package gperl

import (
	"context"
	"debug/elf"
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/goccy/go-perl/internal/oci"
)

// The base images of an image whose binary is static, and of one whose
// binary needs the dynamic loader and the C library.
const (
	staticBase  = "cgr.dev/chainguard/static:latest"
	dynamicBase = "cgr.dev/chainguard/glibc-dynamic:latest"
)

const imageUsage = "usage: gperl image build [-platform list] [-base image] [-push repository:tag] [-tarball file] (script.pl | -psgi app.psgi [-- gperl psgi build options])"

// ImageBuild implements `gperl image build`: it builds the script (or the
// PSGI application's server) for each platform as gperl build does, puts
// each binary at /app/<name> on the base image as its entrypoint, and
// pushes the image — an index of them for several platforms — or writes
// it as the archive docker load reads.
func ImageBuild(args []string) (err error) {
	fs := flag.NewFlagSet("gperl image build", flag.ContinueOnError)
	platforms := fs.String("platform", "linux/amd64,linux/arm64", "the platforms to build for, comma-separated")
	base := fs.String("base", "", "the base image (default: "+staticBase+" for a static binary, "+dynamicBase+" for one that needs the dynamic loader)")
	push := fs.String("push", "", "push the image to this repository:tag")
	tarball := fs.String("tarball", "", "write the image to this file, which docker load reads")
	psgiApp := fs.String("psgi", "", "build a server image of this PSGI application; the arguments after -- are gperl psgi build options")
	if err := fs.Parse(args); err != nil {
		return err
	}
	script := ""
	var psgiDefaults []string
	if *psgiApp == "" {
		if fs.NArg() != 1 {
			return errors.New(imageUsage)
		}
		script = fs.Arg(0)
	} else {
		app, out, defaults, err := parsePSGIBuild(append(fs.Args(), *psgiApp))
		if err != nil {
			return err
		}
		if out != "" {
			return errors.New("gperl image build: -o is not an option of an image")
		}
		script, psgiDefaults = app, defaults
	}
	if *push == "" && *tarball == "" {
		return errors.New("gperl image build: give -push, -tarball, or both")
	}
	var ref oci.Reference
	if *push != "" {
		if ref, err = oci.ParseReference(*push); err != nil {
			return err
		}
	}
	var plats []oci.Platform
	for _, s := range strings.Split(*platforms, ",") {
		// An architecture with variants (arm's v6 and v7) is not taken: the
		// image would not say which one the binary needs.
		goos, goarch, _ := strings.Cut(s, "/")
		if goos != "linux" || (goarch != "amd64" && goarch != "arm64") {
			return fmt.Errorf("gperl image build: platform %q: want linux/amd64 or linux/arm64", s)
		}
		plats = append(plats, oci.Platform{OS: goos, Architecture: goarch})
	}
	if *tarball != "" && len(plats) > 1 {
		return errors.New("gperl image build: docker load reads an image of one platform: give -tarball with one -platform")
	}

	work, err := os.MkdirTemp("", "gperl-image-")
	if err != nil {
		return err
	}
	defer func() { err = errors.Join(err, os.RemoveAll(work)) }()
	name := strings.TrimSuffix(filepath.Base(script), filepath.Ext(script))
	ctx := context.Background()
	client := &oci.Client{}
	var imgs []*oci.Image
	for _, p := range plats {
		config, err := newBuildConfig(p.OS, p.Architecture)
		if err != nil {
			return err
		}
		out := filepath.Join(work, p.OS+"_"+p.Architecture, name)
		if err := buildBinaryFor(config, script, out, psgiDefaults); err != nil {
			return err
		}
		baseName := *base
		if baseName == "" {
			dynamic, err := needsLoader(out)
			if err != nil {
				return err
			}
			baseName = staticBase
			if dynamic {
				baseName = dynamicBase
			}
		}
		baseRef, err := oci.ParseReference(baseName)
		if err != nil {
			return err
		}
		b, err := client.Pull(ctx, baseRef, p)
		if err != nil {
			return err
		}
		data, err := os.ReadFile(out)
		if err != nil {
			return err
		}
		img, err := oci.Append(b, "/app/"+name, data)
		if err != nil {
			return err
		}
		imgs = append(imgs, img)
		fmt.Fprintf(os.Stderr, "gperl: image for %s on %s\n", p, baseRef)
	}
	if *tarball != "" {
		tag := ""
		if *push != "" {
			tag = ref.String()
		}
		f, err := os.Create(*tarball)
		if err != nil {
			return err
		}
		if err := errors.Join(client.WriteTarball(ctx, f, imgs[0], tag), f.Close()); err != nil {
			return err
		}
		fmt.Fprintf(os.Stderr, "gperl: wrote %s\n", *tarball)
	}
	if *push != "" {
		digest, err := client.Push(ctx, ref, imgs)
		if err != nil {
			return err
		}
		fmt.Println(ref.Name() + "@" + digest)
	}
	return nil
}

// glibcLibraries are the shared libraries of the GNU C library, which
// dynamicBase has.
var glibcLibraries = []string{
	"libc.so.6", "libm.so.6", "libpthread.so.0", "libdl.so.2", "librt.so.1", "libresolv.so.2", "libutil.so.1",
	"ld-linux-x86-64.so.2", "ld-linux-aarch64.so.1",
}

// needsLoader reports whether the binary at path is run by a program
// interpreter (the dynamic loader), and so needs it and the C library in
// its image. A binary that needs a shared library beyond the C library's
// is refused: no default base has it.
func needsLoader(path string) (dynamic bool, err error) {
	f, err := elf.Open(path)
	if err != nil {
		return false, fmt.Errorf("%s: %w", path, err)
	}
	defer func() { err = errors.Join(err, f.Close()) }()
	libs, err := f.ImportedLibraries()
	if err != nil {
		return false, fmt.Errorf("%s: %w", path, err)
	}
	for _, lib := range libs {
		if !slices.Contains(glibcLibraries, lib) {
			return false, fmt.Errorf("the binary needs the shared library %s, which no default base image has: give -base an image that has it", lib)
		}
	}
	for _, p := range f.Progs {
		if p.Type == elf.PT_INTERP {
			return true, nil
		}
	}
	return false, nil
}
