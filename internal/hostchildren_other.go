//go:build !darwin && !linux

package internal

import "github.com/goccy/perlwasm2go/base"

// hostChildrenConfig returns the Children of a program that is the host
// process: none on this platform, whose runtime gives every command a pid
// token and cannot wait for a child without reaping it.
func hostChildrenConfig() base.Children { return nil }
