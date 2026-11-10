package main

import "testing"

func TestSatisfies(t *testing.T) {
	for _, tc := range []struct {
		have, want    string
		satisfied, ok bool
	}{
		{"1.302219", "0.88", true, true},
		{"1.302219", "1.302220", false, true},
		{"1.3", "1.29", true, true},    // 1.300 vs 1.290
		{"1.10", "1.9", false, true},   // 1.100 vs 1.900
		{"v6.0.2", "4.14", true, true}, // dotted vs decimal: 6.0.2 vs 4.140
		{"0.25", "v0.250.0", true, true},
		{"0.25", "v0.251.0", false, true},
		{"1.70", ">= 1.45, < 2.0", true, true},
		{"2.01", ">= 1.45, < 2.0", false, true},
		{"1.50", "!= 1.50", false, true},
		{"1.23_01", "1.23", true, true},
		{"", "", true, true},
		{"", "0", true, true},
		{"", "1.0", false, true}, // no declared version is version 0
		{"1.0", "abc", false, false},
		{"1.0-TRIAL", "0.5", false, false},
	} {
		satisfied, ok := satisfies(tc.have, tc.want)
		if satisfied != tc.satisfied || ok != tc.ok {
			t.Errorf("satisfies(%q, %q) = %v, %v; want %v, %v",
				tc.have, tc.want, satisfied, ok, tc.satisfied, tc.ok)
		}
	}
}

func TestStronglyConnected(t *testing.T) {
	// 0 -> 1 -> 2 -> 1, 3 -> 0: {1,2} is a cycle below 0, which is below 3.
	adj := [][]int{{1}, {2}, {1}, {0}}
	component, components := stronglyConnected(adj)
	if len(components) != 3 {
		t.Fatalf("got %d components, want 3", len(components))
	}
	if component[1] != component[2] {
		t.Errorf("1 and 2 should share a component")
	}
	if !(component[1] < component[0] && component[0] < component[3]) {
		t.Errorf("components are not dependencies-first: %v", component)
	}
}
