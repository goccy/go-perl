//go:build darwin || linux

package xs_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	perl "github.com/goccy/go-perl"
)

// resultStr renders an Eval result's scalar as its Perl string form.
func resultStr(r perl.Result) string {
	if s, err := perl.As[perl.ScalarValue](r.Value); err == nil {
		return s.String()
	}
	return r.Value.Kind().String()
}

// loadXSModule registers one prebuilt native module with the instance
// through the public path: it stages the .so under a temp directory with
// the conventional <Module-Name>.so spelling ("::" as "-") and hands the
// directory to AddXSDir, exactly as an application would.
func loadXSModule(t *testing.T, p *perl.Perl, module, so string) {
	t.Helper()
	dir := t.TempDir()
	name := strings.ReplaceAll(module, "::", "-") + ".so"
	if err := os.Symlink(so, filepath.Join(dir, name)); err != nil {
		t.Fatalf("stage %s: %v", name, err)
	}
	if err := p.AddXSDir(dir); err != nil {
		t.Fatalf("AddXSDir(%s): %v", dir, err)
	}
}
