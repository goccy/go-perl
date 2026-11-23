package mod

import (
	"maps"
	"path/filepath"
	"testing"
)

func TestLibFilesMergesTrees(t *testing.T) {
	a, b := t.TempDir(), t.TempDir()
	writeFile(t, filepath.Join(a, "Plack.pm"), "package Plack;")
	writeFile(t, filepath.Join(a, "Plack", "Util.pm"), "from a")
	writeFile(t, filepath.Join(b, "Plack", "Util.pm"), "from b")
	writeFile(t, filepath.Join(b, "Plack", "Request.pm"), "package Plack::Request;")
	writeFile(t, filepath.Join(b, "auto", "share", "dist", "X", "data.txt"), "shared")
	files, err := LibFiles([]string{a, filepath.Join(t.TempDir(), "missing"), b})
	if err != nil {
		t.Fatal(err)
	}
	want := map[string]string{
		"Plack.pm":                   filepath.Join(a, "Plack.pm"),
		"Plack/Util.pm":              filepath.Join(a, "Plack", "Util.pm"), // the earlier tree wins
		"Plack/Request.pm":           filepath.Join(b, "Plack", "Request.pm"),
		"auto/share/dist/X/data.txt": filepath.Join(b, "auto", "share", "dist", "X", "data.txt"),
	}
	if !maps.Equal(files, want) {
		t.Fatalf("LibFiles = %v, want %v", files, want)
	}
}
