package main

import (
	"embed"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
)

// imageFiles is the build context of the container image every
// distribution is tested in, minus the gperl binary.
//
//go:embed image
var imageFiles embed.FS

// buildImage builds the image from the embedded context, around either
// the given gperl binary (built for the architecture and operating system
// of the containers, i.e. linux) or, for the reference image, the given
// official perl image (Dockerfile.reference): the same environment with
// perl itself as the perl command.
func buildImage(dataDir, gperl, perlImage, tag string) error {
	if (gperl == "") == (perlImage == "") {
		return fmt.Errorf("image: give either a gperl binary (-gperl) or a perl image (-perl-image)")
	}
	ctx := filepath.Join(dataDir, "image-context")
	entries, err := imageFiles.ReadDir("image")
	if err != nil {
		return err
	}
	var names []string
	for _, e := range entries {
		names = append(names, e.Name())
	}
	if err := writeContext(ctx, names...); err != nil {
		return err
	}
	args := []string{"build", "-t", tag}
	if perlImage != "" {
		args = append(args, "-f", filepath.Join(ctx, "Dockerfile.reference"), "--build-arg", "PERL_IMAGE="+perlImage)
	} else if err := copyFile(gperl, filepath.Join(ctx, "gperl"), 0o755); err != nil {
		return err
	}
	cmd := exec.Command("docker", append(args, ctx)...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// writeContext makes dir a build context holding the named files of the
// embedded image directory.
func writeContext(dir string, names ...string) error {
	if err := os.RemoveAll(dir); err != nil {
		return err
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	for _, name := range names {
		data, err := imageFiles.ReadFile("image/" + name)
		if err != nil {
			return err
		}
		if err := os.WriteFile(filepath.Join(dir, name), data, 0o644); err != nil {
			return err
		}
	}
	return nil
}

func copyFile(from, to string, mode os.FileMode) error {
	in, err := os.Open(from)
	if err != nil {
		return err
	}
	defer in.Close()
	out, err := os.OpenFile(to, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, mode)
	if err != nil {
		return err
	}
	if _, err := io.Copy(out, in); err != nil {
		out.Close()
		return err
	}
	return out.Close()
}
