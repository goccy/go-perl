package main

import (
	"strconv"
	"strings"
)

// perlVersion is a Perl module version in the normalized form version.pm
// compares: a sequence of integer components.
//
// Perl versions are text with a grammar of their own (perldoc version):
// a dotted-decimal version ("v1.2.3", or any version with two or more
// dots) lists its components directly, while a decimal version ("1.302")
// stands for the components obtained by cutting its fraction into groups
// of three digits ("1.302" is 1.302.0, "1.3" is 1.300). Underscores mark
// developer releases and do not affect the value.
type perlVersion []int

func parsePerlVersion(s string) (perlVersion, bool) {
	s = strings.ReplaceAll(strings.TrimSpace(s), "_", "")
	if s == "" {
		return nil, false
	}
	dotted := strings.HasPrefix(s, "v")
	s = strings.TrimPrefix(s, "v")
	parts := strings.Split(s, ".")
	if len(parts) > 2 {
		dotted = true
	}
	var v perlVersion
	if dotted {
		for _, p := range parts {
			n, ok := digits(p)
			if !ok {
				return nil, false
			}
			v = append(v, n)
		}
		return v, true
	}
	whole := parts[0]
	if whole == "" {
		whole = "0" // ".5" is a valid decimal version
	}
	n, ok := digits(whole)
	if !ok {
		return nil, false
	}
	v = append(v, n)
	if len(parts) == 2 {
		frac := parts[1]
		for len(frac)%3 != 0 {
			frac += "0"
		}
		for i := 0; i < len(frac); i += 3 {
			n, ok := digits(frac[i : i+3])
			if !ok {
				return nil, false
			}
			v = append(v, n)
		}
	}
	return v, true
}

func digits(s string) (int, bool) {
	if s == "" {
		return 0, false
	}
	for _, c := range s {
		if c < '0' || c > '9' {
			return 0, false
		}
	}
	n, err := strconv.Atoi(s)
	return n, err == nil
}

// compare returns -1, 0 or 1; missing trailing components count as zero.
func (v perlVersion) compare(w perlVersion) int {
	for i := 0; i < len(v) || i < len(w); i++ {
		var a, b int
		if i < len(v) {
			a = v[i]
		}
		if i < len(w) {
			b = w[i]
		}
		if a != b {
			if a < b {
				return -1
			}
			return 1
		}
	}
	return 0
}

// satisfies reports whether a module at version have meets the version
// requirement want, written in the CPAN::Meta::Spec form: either a bare
// version (meaning "at least") or a comma-separated list of ranges such
// as ">= 1.2, != 1.5, < 2.0". An empty requirement accepts anything. ok
// is false when either side is not a version this grammar describes.
func satisfies(have, want string) (satisfied, ok bool) {
	want = strings.TrimSpace(want)
	if want == "" || want == "0" {
		return true, true
	}
	// A module that declares no version counts as version 0.
	hv := perlVersion{0}
	if have != "" {
		if hv, ok = parsePerlVersion(have); !ok {
			return false, false
		}
	}
	for _, clause := range strings.Split(want, ",") {
		clause = strings.TrimSpace(clause)
		op := ">="
		for _, candidate := range []string{">=", "<=", "==", "!=", ">", "<"} {
			if rest, found := strings.CutPrefix(clause, candidate); found {
				op, clause = candidate, strings.TrimSpace(rest)
				break
			}
		}
		wv, ok := parsePerlVersion(clause)
		if !ok {
			return false, false
		}
		c := hv.compare(wv)
		var holds bool
		switch op {
		case ">=":
			holds = c >= 0
		case "<=":
			holds = c <= 0
		case "==":
			holds = c == 0
		case "!=":
			holds = c != 0
		case ">":
			holds = c > 0
		case "<":
			holds = c < 0
		}
		if !holds {
			return false, true
		}
	}
	return true, true
}
