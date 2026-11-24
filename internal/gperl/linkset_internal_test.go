package gperl

import (
	"testing"

	"github.com/goccy/go-perl/internal/mod"
)

// What gperl generates from the dependencies — the merged module tree and
// the link set — is generated again by another gperl.
func TestGeneratedDirectoriesAreKeyedByGperl(t *testing.T) {
	m := &mod.Manager{Dir: t.TempDir()}
	saved := mod.GperlVersion
	t.Cleanup(func() { mod.GperlVersion = saved })
	type dirs struct{ lib, linkSet string }
	generate := func(version string) dirs {
		t.Helper()
		mod.GperlVersion = func() string { return version }
		lib, err := mergedLib(m, nil, map[string]string{})
		if err != nil {
			t.Fatal(err)
		}
		links, err := generateLinkSet(m, &projectDeps{lib: map[string]string{}})
		if err != nil {
			t.Fatal(err)
		}
		return dirs{lib: lib, linkSet: links.cpanDir}
	}
	a, b := generate("gperl a"), generate("gperl b")
	if a.lib == b.lib || a.linkSet == b.linkSet {
		t.Fatalf("two gperls share %+v and %+v", a, b)
	}
	if again := generate("gperl a"); again != a {
		t.Fatalf("one gperl generated %+v, then %+v", a, again)
	}
}
