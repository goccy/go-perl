package main

import (
	"bytes"
	"encoding/json"
)

// versionText is a version as written in distribution metadata. MetaCPAN
// passes the author's JSON through, so the same field arrives as a string
// ("1.02") from one release and as a number (1.02) from another; the
// digits as written are the version either way.
type versionText string

func (v *versionText) UnmarshalJSON(b []byte) error {
	if bytes.Equal(b, []byte("null")) {
		*v = ""
		return nil
	}
	if len(b) > 0 && b[0] == '"' {
		var s string
		if err := json.Unmarshal(b, &s); err != nil {
			return err
		}
		*v = versionText(s)
		return nil
	}
	var n json.Number
	if err := json.Unmarshal(b, &n); err != nil {
		return err
	}
	*v = versionText(n.String())
	return nil
}

// count is a non-negative counter from MetaCPAN, which serves the same
// field as a JSON number in one document and as a string of digits in
// another.
type count int64

func (c *count) UnmarshalJSON(b []byte) error {
	if bytes.Equal(b, []byte("null")) {
		*c = 0
		return nil
	}
	if len(b) > 0 && b[0] == '"' {
		var s string
		if err := json.Unmarshal(b, &s); err != nil {
			return err
		}
		b = []byte(s)
	}
	var n int64
	if err := json.Unmarshal(b, &n); err != nil {
		return err
	}
	*c = count(n)
	return nil
}
