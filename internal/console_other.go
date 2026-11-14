//go:build !unix

package internal

// openNoControllingTerminal: a system without controlling terminals has no
// flag to keep a device from becoming one.
const openNoControllingTerminal = 0
