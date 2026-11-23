//go:build darwin || linux

package xs

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

// The loader writes serving_in_init of the SDK's process block at
// processServingInInitOffset, which is where the C compiler puts the field
// of the SDK's goperl_process_t.
func TestProcessBlockServingInInitIsTheSDKs(t *testing.T) {
	cc, err := exec.LookPath("cc")
	if err != nil {
		t.Skip("no host C compiler available")
	}
	dir := t.TempDir()
	src := filepath.Join(dir, "layout.c")
	code := fmt.Sprintf(`#include "EXTERN.h"
#include "perl.h"
#include "XSUB.h"
_Static_assert(__builtin_offsetof(goperl_process_t, serving_in_init) == %d, "serving_in_init");
`, processServingInInitOffset)
	if err := os.WriteFile(src, []byte(code), 0o644); err != nil {
		t.Fatal(err)
	}
	sdk, err := filepath.Abs("sdk/include")
	if err != nil {
		t.Fatal(err)
	}
	if out, err := exec.Command(cc, "-fsyntax-only", "-std=gnu99", "-D_GNU_SOURCE", "-I", sdk, src).CombinedOutput(); err != nil {
		t.Fatalf("the SDK's serving_in_init is not at %d: %v\n%s", processServingInInitOffset, err, out)
	}
}
