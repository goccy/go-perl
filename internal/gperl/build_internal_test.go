package gperl

import (
	"runtime/debug"
	"testing"
)

func TestGoModSrcRequiresGperlsGoPerlVersion(t *testing.T) {
	const header = "module " + generatedModulePath + "\n\ngo 1.25.0\n\n"
	main := func(version string) *debug.BuildInfo {
		return &debug.BuildInfo{Main: debug.Module{Path: goPerlModule, Version: version}}
	}
	for _, tc := range []struct {
		name       string
		info       *debug.BuildInfo
		devReplace string
		want       string // "" when goModSrc fails
	}{
		{name: "tagged", info: main("v0.3.1"), want: header + "require github.com/goccy/go-perl v0.3.1\n"},
		{name: "pseudo-version", info: main("v0.3.1-0.20261124000000-0123456789ab"), want: header + "require github.com/goccy/go-perl v0.3.1-0.20261124000000-0123456789ab\n"},
		{name: "devel", info: main("(devel)")},
		{name: "no version", info: main("")},
		{name: "uncommitted changes", info: main("v0.3.1-0.20261124000000-0123456789ab+dirty")},
		{name: "replaced", info: &debug.BuildInfo{Main: debug.Module{Path: goPerlModule, Version: "v0.3.1", Replace: &debug.Module{Path: "../go-perl"}}}},
		{name: "no build information"},
		{name: "devel replaced", info: main("(devel)"), devReplace: "github.com/goccy/go-perl=/src/go-perl, github.com/goccy/perlwasm2go=/src/perlwasm2go",
			want: header + "require github.com/goccy/go-perl v0.0.0\n\nreplace github.com/goccy/go-perl => /src/go-perl\nreplace github.com/goccy/perlwasm2go => /src/perlwasm2go\n"},
		{name: "no build information replaced", devReplace: "github.com/goccy/go-perl =/src/go-perl",
			want: header + "require github.com/goccy/go-perl v0.0.0\n\nreplace github.com/goccy/go-perl  => /src/go-perl\n"},
		{name: "tagged replaced", info: main("v0.3.1"), devReplace: "github.com/goccy/go-perl=/src/go-perl",
			want: header + "require github.com/goccy/go-perl v0.3.1\n\nreplace github.com/goccy/go-perl => /src/go-perl\n"},
		{name: "devel with go-perl not replaced", info: main("(devel)"), devReplace: "github.com/goccy/perlwasm2go=/src/perlwasm2go"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got, err := goModSrc(tc.info, tc.devReplace)
			if tc.want == "" {
				if err == nil {
					t.Fatalf("goModSrc succeeded with\n%s", got)
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if string(got) != tc.want {
				t.Fatalf("go.mod:\n%s\nwant:\n%s", got, tc.want)
			}
		})
	}
}
