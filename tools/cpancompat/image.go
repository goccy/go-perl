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

// buildImage builds the image from the embedded context and the given
// gperl binary (built for the architecture and operating system of the
// containers, i.e. linux).
func buildImage(dataDir, gperl, tag string) error {
	if gperl == "" {
		return fmt.Errorf("image: no gperl binary (-gperl)")
	}
	ctx := filepath.Join(dataDir, "image-context")
	if err := os.RemoveAll(ctx); err != nil {
		return err
	}
	if err := os.MkdirAll(ctx, 0o755); err != nil {
		return err
	}
	entries, err := imageFiles.ReadDir("image")
	if err != nil {
		return err
	}
	for _, e := range entries {
		data, err := imageFiles.ReadFile("image/" + e.Name())
		if err != nil {
			return err
		}
		if err := os.WriteFile(filepath.Join(ctx, e.Name()), data, 0o644); err != nil {
			return err
		}
	}
	if err := copyFile(gperl, filepath.Join(ctx, "gperl"), 0o755); err != nil {
		return err
	}
	cmd := exec.Command("docker", "build", "-t", tag, ctx)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
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
