package perl

import (
	"archive/zip"
	"bytes"
	"testing"
)

// zipOf builds an archive of the given entries, in order; a name ending in
// "/" is a directory.
func zipOf(t *testing.T, entries ...[2]string) []byte {
	t.Helper()
	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)
	for _, e := range entries {
		w, err := zw.Create(e[0])
		if err != nil {
			t.Fatal(err)
		}
		if _, err := w.Write([]byte(e[1])); err != nil {
			t.Fatal(err)
		}
	}
	if err := zw.Close(); err != nil {
		t.Fatal(err)
	}
	return buf.Bytes()
}

// The cache directory of the library is named by what an unpacked copy is
// made of: archives of the same entries share it, and a change to the
// content of one entry (even one that keeps its size), to a name, or to
// the set of entries gives another.
func TestStdlibDigestNamesTheContent(t *testing.T) {
	base := [][2]string{{"strict.pm", "package strict; 1;\n"}, {"File/", ""}, {"File/Spec.pm", "package File::Spec; 1;\n"}}
	digest := func(entries ...[2]string) [32]byte {
		t.Helper()
		sum, err := stdlibDigest(zipOf(t, entries...))
		if err != nil {
			t.Fatal(err)
		}
		return sum
	}
	want := digest(base...)
	if got := digest(base...); got != want {
		t.Fatalf("two archives of the same entries: %x and %x", got, want)
	}
	changed := map[string][][2]string{
		"content of the same size": {base[0], base[1], {"File/Spec.pm", "package File::Spek; 1;\n"}},
		"a name":                   {base[0], base[1], {"File/Spec2.pm", base[2][1]}},
		"an entry more":            {base[0], base[1], base[2], {"warnings.pm", "1;\n"}},
		"an entry less":            {base[0], base[2]},
	}
	for what, entries := range changed {
		if digest(entries...) == want {
			t.Errorf("%s: the digest did not change", what)
		}
	}
	if _, err := stdlibDigest(stdlibZip); err != nil {
		t.Fatalf("the embedded library: %v", err)
	}
	if _, err := stdlibDigest([]byte("not an archive")); err == nil {
		t.Fatal("no error for something that is no archive")
	}
}
