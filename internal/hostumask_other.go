//go:build !unix

package internal

import "os"

// startUmask is the umask a program that is the host process starts with on
// a system without umasks: POSIX's usual one, which the runtime takes out of
// the modes of the files the program creates.
const startUmask os.FileMode = 0o022
