package perl

import "strconv"

// Value is the result of evaluating Perl code — a Perl scalar as seen from
// Go. Perl scalars are stringly at heart, so the accessors coerce the way
// Perl itself would: String is the scalar's string form, Float and Int
// numify it (a leading numeric prefix counts, anything else is 0 — Perl's
// numeric conversion), and Bool applies Perl truth ("" and "0" are false,
// everything else is true). Export returns the default Go representation.
//
// The interface is sealed: only this package's types are Values. A scalar
// that was undef surfaces with an empty String (the evaluation envelope
// carries the stringified value).
type Value interface {
	String() string
	Float() float64
	Int() int
	Bool() bool
	Export() any

	// isValue seals the interface: only this package's types are Values.
	isValue()
}

// scalar implements Value over the stringified form of a Perl scalar.
type scalar struct{ s string }

func (scalar) isValue()         {}
func (v scalar) String() string { return v.s }
func (v scalar) Export() any    { return v.s }
func (v scalar) Float() float64 { return numify(v.s) }
func (v scalar) Int() int       { return int(numify(v.s)) }
func (v scalar) Bool() bool     { return v.s != "" && v.s != "0" }

// numify converts a string to a number the way Perl does: skip leading
// whitespace, then read the longest prefix that parses as a decimal number
// (sign, digits, fraction, exponent); anything else contributes 0. This is
// a total scan — every input maps to a defined value.
func numify(s string) float64 {
	i := 0
	for i < len(s) && (s[i] == ' ' || s[i] == '\t' || s[i] == '\n' || s[i] == '\r' || s[i] == '\f') {
		i++
	}
	start := i
	if i < len(s) && (s[i] == '+' || s[i] == '-') {
		i++
	}
	digits := false
	for i < len(s) && s[i] >= '0' && s[i] <= '9' {
		i++
		digits = true
	}
	if i < len(s) && s[i] == '.' {
		j := i + 1
		for j < len(s) && s[j] >= '0' && s[j] <= '9' {
			j++
			digits = true
		}
		if j > i+1 || digits {
			i = j
		}
	}
	if !digits {
		return 0
	}
	if i < len(s) && (s[i] == 'e' || s[i] == 'E') {
		j := i + 1
		if j < len(s) && (s[j] == '+' || s[j] == '-') {
			j++
		}
		expDigits := false
		for j < len(s) && s[j] >= '0' && s[j] <= '9' {
			j++
			expDigits = true
		}
		if expDigits {
			i = j
		}
	}
	f, err := strconv.ParseFloat(s[start:i], 64)
	if err != nil {
		return 0
	}
	return f
}
