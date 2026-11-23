//go:build unix

package internal

import "syscall"

// openNoControllingTerminal opens a terminal device without making it the
// controlling terminal of the process (O_NOCTTY).
const openNoControllingTerminal = syscall.O_NOCTTY
