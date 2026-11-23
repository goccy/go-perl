//go:build darwin || linux

package xs

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"unsafe"
)

// The loader's frame (cFrame) is the SDK's goperl_frame_t: every field the
// two share is at the same offset, which the C compiler says of the SDK's
// struct and Go of the loader's.
func TestFrameLayoutIsTheSDKs(t *testing.T) {
	cc, err := exec.LookPath("cc")
	if err != nil {
		t.Skip("no host C compiler available")
	}
	var f cFrame
	fields := []struct {
		c  string
		at uintptr
	}{
		{"api", unsafe.Offsetof(f.api)},
		{"subname", unsafe.Offsetof(f.subname)},
		{"jb", unsafe.Offsetof(f.jb)},
		{"failed", unsafe.Offsetof(f.failed)},
		{"psp", unsafe.Offsetof(f.psp)},
		{"markidx", unsafe.Offsetof(f.markidx)},
		{"hostsave_base", unsafe.Offsetof(f.hostsaveBase)},
		{"hook_dirty", unsafe.Offsetof(f.hookDirty)},
		{"tmpidx", unsafe.Offsetof(f.tmpidx)},
		{"hook_val", unsafe.Offsetof(f.hookVal)},
		{"imm", unsafe.Offsetof(f.imm)},
		{"plop", unsafe.Offsetof(f.plop)},
		{"hook_op_tok", unsafe.Offsetof(f.hookOpTok)},
		{"prev_frame", unsafe.Offsetof(f.prevFrame)},
		{"inst", unsafe.Offsetof(f.instShared)},
		{"stbase", unsafe.Offsetof(f.stBase)},
		{"stcap", unsafe.Offsetof(f.stCap)},
		{"tmp", unsafe.Offsetof(f.tmp)},
		{"err", unsafe.Offsetof(f.err)},
		{"xerrno", unsafe.Offsetof(f.xerrno)},
		{"membase", unsafe.Offsetof(f.memBase)},
		{"memsize", unsafe.Offsetof(f.memSize)},
		{"pp_frame", unsafe.Offsetof(f.ppFrame)},
		{"pp_vars", unsafe.Offsetof(f.ppVars)},
		{"native_call", unsafe.Offsetof(f.nativeCall)},
		{"entry_jb", unsafe.Offsetof(f.entryJB)},
		{"opbox", unsafe.Offsetof(f.opbox)},
		{"tmp_more", unsafe.Offsetof(f.tmpMore)},
		{"markcap", unsafe.Offsetof(f.markcap)},
		{"markv", unsafe.Offsetof(f.markv)},
		{"plop_base", unsafe.Offsetof(f.plopBase)},
	}
	var src strings.Builder
	src.WriteString("#include \"EXTERN.h\"\n#include \"perl.h\"\n#include <stdio.h>\n#include <stddef.h>\nint main(void) {\n")
	for _, fl := range fields {
		fmt.Fprintf(&src, "    printf(\"%s %%zu\\n\", offsetof(goperl_frame_t, %s));\n", fl.c, fl.c)
	}
	src.WriteString("    return 0;\n}\n")
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "layout.c"), []byte(src.String()), 0o644); err != nil {
		t.Fatal(err)
	}
	sdk, err := filepath.Abs("sdk/include")
	if err != nil {
		t.Fatal(err)
	}
	args := []string{"-std=gnu99", "-w", "-I", sdk, "-o", filepath.Join(dir, "layout"), filepath.Join(dir, "layout.c")}
	if runtime.GOOS == "darwin" {
		switch runtime.GOARCH {
		case "arm64":
			args = append([]string{"-arch", "arm64"}, args...)
		case "amd64":
			args = append([]string{"-arch", "x86_64"}, args...)
		}
	} else {
		args = append([]string{"-D_GNU_SOURCE"}, args...)
	}
	if out, err := exec.Command(cc, args...).CombinedOutput(); err != nil {
		t.Fatalf("compile: %v\n%s", err, out)
	}
	out, err := exec.Command(filepath.Join(dir, "layout")).Output()
	if err != nil {
		t.Fatalf("run: %v", err)
	}
	var want strings.Builder
	for _, fl := range fields {
		fmt.Fprintf(&want, "%s %d\n", fl.c, fl.at)
	}
	if string(out) != want.String() {
		t.Errorf("the SDK's offsets:\n%s\nthe loader's:\n%s", out, want.String())
	}
}
