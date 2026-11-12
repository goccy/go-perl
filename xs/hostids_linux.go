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
	"syscall"
	"unsafe"

	"github.com/ebitengine/purego"

	"github.com/goccy/go-perl/internal"
)

func init() {
	internal.RegisterHostIDs(libcIDs{})
}

// libcIDs changes the ids of the process with the C library's calls.
type libcIDs struct{}

// libcCall calls the C library function name, whose result is 0 or -1 with
// errno. An id of -1 (leave alone) is the id type's all-ones value.
func libcCall(name string, args ...uintptr) error {
	fn, err := purego.Dlsym(purego.RTLD_DEFAULT, name)
	if err != nil {
		return errors.ErrUnsupported
	}
	result, _, errno := purego.SyscallN(fn, args...)
	if int32(result) != 0 {
		return syscall.Errno(errno)
	}
	return nil
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
	return libcCall("setgroups", uintptr(len(list)), uintptr(first))
}
