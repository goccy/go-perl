package internal

import "unsafe"

// childSiginfo is Linux's siginfo_t of a child (waitid(2)): si_signo,
// si_errno and si_code, then, in the union, aligned as a pointer is,
// si_pid, si_uid and si_status, and room for the rest of its 128 bytes.
type childSiginfo struct {
	signo, errno, code int32
	_                  [unsafe.Sizeof(uintptr(0))/4 - 1]int32
	pid, uid, status   int32
	_                  [128 - 24]byte
}
