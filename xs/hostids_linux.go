//go:build linux

package xs

// The user and group ids of the host process, changed through the C
// library.
//
// To the Linux kernel the ids belong to a thread. A process changes them as
// a whole because its C library changes them on every thread, and the Go
// runtime does the same for a pure Go program — but refuses to, by
// panicking, in a program that is linked with a C runtime it does not
// control, which a program that loads native modules is. There the C
// library is the one to ask: its setuid family keeps the threads of the
// process in step.

import (
	"errors"
	"fmt"
	"runtime"
	"syscall"
	"unsafe"

	"github.com/goccy/go-perl/internal/ccall"

	"github.com/goccy/go-perl/internal"
)

func init() {
	internal.RegisterHostIDs(libcIDs{})
}

// libcIDs changes the ids of the process with the C library's calls.
type libcIDs struct{}

// libcCall calls the C library function name, whose result is 0 or -1 with
// errno. An id of -1 (leave alone) is the id type's all-ones value.
//
// errno is the calling thread's, and a call into C does not collect it:
// the call and the reading of errno, at the address
// __errno_location gives, run on one locked OS thread, and the C library's
// memcpy reads it (Go code does not dereference the C library's memory).
//
// The functions are the C library's the native modules are linked with
// (libcSymbol): in a binary that links the modules in, the one linked into
// it, not one the dynamic loader would load beside it.
func libcCall(name string, args ...uintptr) error {
	fn, err := libcSymbol(name)
	if err != nil {
		return errors.ErrUnsupported
	}
	errnoLocation, err := libcSymbol("__errno_location")
	if err != nil {
		return errors.ErrUnsupported
	}
	memcpy, err := libcSymbol("memcpy")
	if err != nil {
		return errors.ErrUnsupported
	}
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()
	result := ccall.Invoke(fn, args...)
	if int32(result) == 0 {
		return nil
	}
	at := ccall.Invoke(errnoLocation)
	var errno int32
	ccall.Invoke(memcpy, uintptr(unsafe.Pointer(&errno)), at, unsafe.Sizeof(errno))
	if errno == 0 {
		return fmt.Errorf("%s failed without an errno", name)
	}
	return syscall.Errno(errno)
}

func id(n int) uintptr { return uintptr(uint32(int32(n))) }

func (libcIDs) Setuid(uid int) error          { return libcCall("setuid", id(uid)) }
func (libcIDs) Seteuid(euid int) error        { return libcCall("seteuid", id(euid)) }
func (libcIDs) Setreuid(ruid, euid int) error { return libcCall("setreuid", id(ruid), id(euid)) }
func (libcIDs) Setresuid(ruid, euid, suid int) error {
	return libcCall("setresuid", id(ruid), id(euid), id(suid))
}
func (libcIDs) Setgid(gid int) error          { return libcCall("setgid", id(gid)) }
func (libcIDs) Setegid(egid int) error        { return libcCall("setegid", id(egid)) }
func (libcIDs) Setregid(rgid, egid int) error { return libcCall("setregid", id(rgid), id(egid)) }
func (libcIDs) Setresgid(rgid, egid, sgid int) error {
	return libcCall("setresgid", id(rgid), id(egid), id(sgid))
}

func (libcIDs) Setgroups(gids []int) error {
	list := make([]uint32, len(gids))
	for i, gid := range gids {
		list[i] = uint32(gid)
	}
	var first unsafe.Pointer
	if len(list) > 0 {
		first = unsafe.Pointer(&list[0])
	}
	err := libcCall("setgroups", uintptr(len(list)), uintptr(first))
	// list is only an address to libcCall: it stays alive until the call
	// has returned.
	runtime.KeepAlive(list)
	return err
}
