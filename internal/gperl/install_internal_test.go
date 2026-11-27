package gperl

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

// fakeTool puts an executable where a managed tool of this version lives.
func fakeTool(t *testing.T, dir, name string) string {
	t.Helper()
	for _, tool := range managedTools {
		if tool.name == name {
			bin := filepath.Join(dir, tool.name+"-"+tool.version, tool.bin)
			if err := os.MkdirAll(filepath.Dir(bin), 0o755); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(bin, []byte("#!/bin/sh\n"), 0o755); err != nil {
				t.Fatal(err)
			}
			return bin
		}
	}
	t.Fatalf("no managed tool %s", name)
	return ""
}

// An explicit CC/AR wins, then the managed Zig, then PATH.
func TestCToolchainPrecedence(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("GPERLTOOLCHAIN", dir)
	t.Setenv("CC", "")
	t.Setenv("AR", "")
	host := hostPlatform()

	// Nothing managed: the compiler on PATH for this platform, and an
	// explanation for another.
	binDir := t.TempDir()
	if err := os.WriteFile(filepath.Join(binDir, "cc"), []byte("#!/bin/sh\n"), 0o755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", binDir)
	cc, ar, err := cToolchain(host)
	if err != nil || !reflect.DeepEqual(cc, []string{filepath.Join(binDir, "cc")}) || !reflect.DeepEqual(ar, []string{"ar"}) {
		t.Fatalf("PATH compiler: %v %v %v", cc, ar, err)
	}
	// This platform's musl variant is still this platform: the compiler on
	// PATH builds for it.
	if cc, _, err := cToolchain(host + muslSuffix); err != nil || !reflect.DeepEqual(cc, []string{filepath.Join(binDir, "cc")}) {
		t.Fatalf("PATH compiler for musl: %v %v", cc, err)
	}
	if _, _, err := cToolchain("linux_riscv64"); err == nil || !strings.Contains(err.Error(), "gperl install toolchain") {
		t.Fatalf("cross target without a cross compiler: %v", err)
	}
	t.Setenv("PATH", t.TempDir())
	if _, _, err := cToolchain(host); err == nil || !strings.Contains(err.Error(), "gperl install toolchain") {
		t.Fatalf("no compiler anywhere: %v", err)
	}

	// The managed Zig takes over, compiling for any CPU of this platform,
	// and is told a foreign target.
	zig := fakeTool(t, dir, "zig")
	cc, ar, err = cToolchain(host)
	if err != nil || !reflect.DeepEqual(cc, []string{zig, "cc", "-mcpu=baseline"}) || !reflect.DeepEqual(ar, []string{zig, "ar"}) {
		t.Fatalf("managed compiler: %v %v %v", cc, ar, err)
	}
	other := "linux_amd64"
	want := []string{zig, "cc", "-target", "x86_64-linux-gnu"}
	if host == other {
		other, want = "linux_arm64", []string{zig, "cc", "-target", "aarch64-linux-gnu"}
	}
	if cc, _, err = cToolchain(other); err != nil || !reflect.DeepEqual(cc, want) {
		t.Fatalf("managed cross compiler: %v %v", cc, err)
	}
	// and told musl for a musl target.
	want[len(want)-1] = strings.TrimSuffix(want[len(want)-1], "-gnu") + "-musl"
	if cc, _, err = cToolchain(other + muslSuffix); err != nil || !reflect.DeepEqual(cc, want) {
		t.Fatalf("managed cross compiler for musl: %v %v", cc, err)
	}

	// What the environment names wins over both.
	t.Setenv("CC", "zig cc -target aarch64-linux-musl")
	t.Setenv("AR", "zig ar")
	cc, ar, err = cToolchain(other)
	if err != nil || strings.Join(cc, " ") != "zig cc -target aarch64-linux-musl" || strings.Join(ar, " ") != "zig ar" {
		t.Fatalf("explicit compiler: %v %v %v", cc, ar, err)
	}
}

func TestGoCommandPrecedence(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("GPERLTOOLCHAIN", dir)
	t.Setenv("PATH", t.TempDir())
	if _, err := goCommand(); err == nil || !strings.Contains(err.Error(), "gperl install toolchain") {
		t.Fatalf("no go anywhere: %v", err)
	}
	binDir := t.TempDir()
	onPath := filepath.Join(binDir, "go")
	if err := os.WriteFile(onPath, []byte("#!/bin/sh\n"), 0o755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", binDir)
	if got, err := goCommand(); err != nil || got != onPath {
		t.Fatalf("go on PATH: %q, %v", got, err)
	}
	managed := fakeTool(t, dir, "go")
	if got, err := goCommand(); err != nil || got != managed {
		t.Fatalf("managed go: %q, %v", got, err)
	}
}

func TestZigTripleAndPins(t *testing.T) {
	for platform, want := range map[string]string{
		"darwin_arm64": "aarch64-macos", "darwin_amd64": "x86_64-macos",
		"linux_arm64": "aarch64-linux", "linux_amd64": "x86_64-linux",
	} {
		if got := zigTriple(platform); got != want {
			t.Errorf("zigTriple(%s) = %q", platform, got)
		}
		for _, tool := range managedTools {
			if len(tool.sha256[platform]) != 64 {
				t.Errorf("%s has no pinned hash for %s", tool.name, platform)
			}
			if u := tool.url(platform); !strings.HasPrefix(u, "https://") || !strings.Contains(u, tool.version) {
				t.Errorf("%s url for %s = %q", tool.name, platform, u)
			}
		}
	}
	t.Setenv("GPERLTOOLCHAIN", "")
	t.Setenv("HOME", "/home/gopher")
	if dir, err := toolchainDir(); err != nil || dir != filepath.Join("/home/gopher", "gperl", "toolchain") {
		t.Fatalf("default toolchain directory = %q, %v", dir, err)
	}
}

// An archive is installed only when its hash matches the pinned one.
func TestInstallToolVerifiesTheDownload(t *testing.T) {
	var buf bytes.Buffer
	zw := gzip.NewWriter(&buf)
	tw := tar.NewWriter(zw)
	for _, hdr := range []*tar.Header{
		{Name: "tool-1.0/", Typeflag: tar.TypeDir, Mode: 0o755},
		{Name: "tool-1.0/bin/tool", Typeflag: tar.TypeReg, Mode: 0o755, Size: 2},
		{Name: "tool-1.0/bin/alias", Typeflag: tar.TypeSymlink, Linkname: "tool"},
	} {
		if err := tw.WriteHeader(hdr); err != nil {
			t.Fatal(err)
		}
		if hdr.Typeflag == tar.TypeReg {
			if _, err := tw.Write([]byte("ok")); err != nil {
				t.Fatal(err)
			}
		}
	}
	if err := tw.Close(); err != nil {
		t.Fatal(err)
	}
	if err := zw.Close(); err != nil {
		t.Fatal(err)
	}
	archive := buf.Bytes()
	sum := sha256.Sum256(archive)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/tool.tar.gz" {
			http.NotFound(w, r)
			return
		}
		if _, err := w.Write(archive); err != nil {
			t.Error(err)
		}
	}))
	defer srv.Close()

	dir := t.TempDir()
	dst := filepath.Join(dir, "tool-1.0")
	if err := installTool(dir, dst, srv.URL+"/tool.tar.gz", strings.Repeat("0", 64)); err == nil || !strings.Contains(err.Error(), "does not match") {
		t.Fatalf("a download with the wrong hash was installed: %v", err)
	}
	if err := installTool(dir, dst, srv.URL+"/missing.tar.gz", hex.EncodeToString(sum[:])); err == nil {
		t.Fatal("a missing download was installed")
	}
	if entries, err := os.ReadDir(dir); err != nil || len(entries) != 0 {
		t.Fatalf("failed installs left files behind: %v, %v", entries, err)
	}
	if err := installTool(dir, dst, srv.URL+"/tool.tar.gz", hex.EncodeToString(sum[:])); err != nil {
		t.Fatal(err)
	}
	if data, err := os.ReadFile(filepath.Join(dst, "bin", "alias")); err != nil || string(data) != "ok" {
		t.Fatalf("installed tool: %q, %v", data, err)
	}
	if entries, err := os.ReadDir(dir); err != nil || len(entries) != 1 {
		t.Fatalf("install left temporary files: %v, %v", entries, err)
	}
}
