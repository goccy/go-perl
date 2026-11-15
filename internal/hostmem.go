package internal

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"io/fs"
	"os"
	"path"
	"runtime"
	"strconv"
	"strings"
)

// HostMemoryLimit, as InstanceOptions.MaxMemoryBytes, limits an instance's
// linear memory to the memory the host gives this process (hostMemory).
const HostMemoryLimit = -1

// hostMemory is the memory the host gives this process, in bytes: the
// machine's memory, or the memory limit of the control group (container)
// the process runs in when that is lower. ok is false where it cannot be
// told (a system with neither sysctl nor /proc: an instance then gets the
// default limit).
func hostMemory() (size uint64, ok bool) {
	return hostMemoryOf(runtime.GOOS, os.DirFS("/"), sysctlValue)
}

// HostMemory is hostMemory, for the tests of the packages that embed the
// interpreter: a test that writes gigabytes runs where the host has them.
func HostMemory() (size uint64, ok bool) {
	return hostMemory()
}

// hostMemoryOf is hostMemory for the system goos, with its files under
// root and its sysctl values from sysctl: the machine's memory is
// hw.memsize on darwin, hw.physmem on FreeBSD and DragonFly, hw.physmem64
// on OpenBSD and NetBSD (each a 64-bit integer in the machine's byte
// order, little-endian on the hosts go-perl runs on), MemTotal in
// /proc/meminfo on Linux, where the control groups of /proc/self/cgroup
// may limit it further (cgroup v2 memory.max, v1 memory.limit_in_bytes, in
// the group and every group above it).
func hostMemoryOf(goos string, root fs.FS, sysctl func(name string) ([]byte, error)) (uint64, bool) {
	switch goos {
	case "darwin":
		return sysctlUint64(sysctl, "hw.memsize")
	case "freebsd", "dragonfly":
		return sysctlUint64(sysctl, "hw.physmem")
	case "openbsd", "netbsd":
		return sysctlUint64(sysctl, "hw.physmem64")
	case "linux":
		total, ok := linuxMemTotal(root)
		if !ok {
			return 0, false
		}
		if limit, ok := linuxCgroupLimit(root); ok && limit < total {
			return limit, true
		}
		return total, true
	}
	return 0, false
}

func sysctlUint64(sysctl func(string) ([]byte, error), name string) (uint64, bool) {
	b, err := sysctl(name)
	if err != nil || len(b) == 0 || len(b) > 8 {
		return 0, false
	}
	var v [8]byte
	copy(v[:], b) // a trailing zero byte may be left out
	n := binary.LittleEndian.Uint64(v[:])
	return n, n > 0
}

// linuxMemTotal is MemTotal of /proc/meminfo, in bytes.
func linuxMemTotal(root fs.FS) (uint64, bool) {
	b, err := fs.ReadFile(root, "proc/meminfo")
	if err != nil {
		return 0, false
	}
	s := bufio.NewScanner(bytes.NewReader(b))
	for s.Scan() {
		fields := strings.Fields(s.Text())
		if len(fields) == 3 && fields[0] == "MemTotal:" && fields[2] == "kB" {
			kb, err := strconv.ParseUint(fields[1], 10, 64)
			return kb << 10, err == nil && kb > 0
		}
	}
	return 0, false
}

// linuxCgroupLimit is the lowest memory limit of the control groups the
// process is in and of the groups above them, in bytes; ok is false
// without one. /proc/self/cgroup names the group of each hierarchy: "0::"
// the cgroup v2 one, a hierarchy with the memory controller among its
// controllers the v1 one.
func linuxCgroupLimit(root fs.FS) (uint64, bool) {
	b, err := fs.ReadFile(root, "proc/self/cgroup")
	if err != nil {
		return 0, false
	}
	var limit uint64
	found := false
	take := func(v uint64, ok bool) {
		if ok && (!found || v < limit) {
			limit, found = v, true
		}
	}
	s := bufio.NewScanner(bytes.NewReader(b))
	for s.Scan() {
		parts := strings.SplitN(s.Text(), ":", 3)
		if len(parts) != 3 || !strings.HasPrefix(parts[2], "/") {
			continue
		}
		group := parts[2]
		switch {
		case parts[0] == "0" && parts[1] == "":
			take(cgroupLimitUp(root, "sys/fs/cgroup", group, "memory.max"))
		case hasController(parts[1], "memory"):
			take(cgroupLimitUp(root, "sys/fs/cgroup/memory", group, "memory.limit_in_bytes"))
		}
	}
	return limit, found
}

func hasController(list, name string) bool {
	for _, c := range strings.Split(list, ",") {
		if c == name {
			return true
		}
	}
	return false
}

// cgroupLimitUp is the lowest limit in file of the group under mount and
// of every group above it up to the mount; "max" (v2) and the huge number
// v1 reports for no limit are no limit.
func cgroupLimitUp(root fs.FS, mount, group, file string) (uint64, bool) {
	var limit uint64
	found := false
	for dir := path.Clean(group); ; dir = path.Dir(dir) {
		b, err := fs.ReadFile(root, path.Join(mount, dir, file))
		if err == nil {
			value := strings.TrimSpace(string(b))
			if v, err := strconv.ParseUint(value, 10, 64); err == nil && v > 0 && v < 1<<62 && (!found || v < limit) {
				limit, found = v, true
			}
		}
		if dir == "/" || dir == "." {
			break
		}
	}
	return limit, found
}
