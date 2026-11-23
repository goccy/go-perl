package internal

import (
	"encoding/binary"
	"errors"
	"testing"
	"testing/fstest"
)

func sysctlOf(values map[string]uint64) func(string) ([]byte, error) {
	return func(name string) ([]byte, error) {
		v, ok := values[name]
		if !ok {
			return nil, errors.New("no such sysctl")
		}
		var b [8]byte
		binary.LittleEndian.PutUint64(b[:], v)
		// syscall.Sysctl drops a trailing zero byte
		return b[:7], nil
	}
}

func file(s string) *fstest.MapFile { return &fstest.MapFile{Data: []byte(s)} }

const meminfo16G = "MemTotal:       16777216 kB\nMemFree:         1048576 kB\n"

// The host's memory is the machine's, as each system reports it, or the
// limit of the control groups the process is in on Linux when that is
// lower.
func TestHostMemoryOf(t *testing.T) {
	noFiles := fstest.MapFS{}
	noSysctl := sysctlOf(nil)
	for _, c := range []struct {
		name   string
		goos   string
		root   fstest.MapFS
		sysctl func(string) ([]byte, error)
		want   uint64
		ok     bool
	}{
		{"darwin", "darwin", noFiles, sysctlOf(map[string]uint64{"hw.memsize": 32 << 30}), 32 << 30, true},
		{"freebsd", "freebsd", noFiles, sysctlOf(map[string]uint64{"hw.physmem": 8 << 30}), 8 << 30, true},
		{"openbsd", "openbsd", noFiles, sysctlOf(map[string]uint64{"hw.physmem64": 4 << 30}), 4 << 30, true},
		{"darwin without the sysctl", "darwin", noFiles, noSysctl, 0, false},
		{"linux without a control group", "linux", fstest.MapFS{"proc/meminfo": file(meminfo16G)}, noSysctl, 16 << 30, true},
		{"linux, cgroup v2, limited above the group", "linux", fstest.MapFS{
			"proc/meminfo":     file(meminfo16G),
			"proc/self/cgroup": file("0::/system.slice/job.scope\n"),
			"sys/fs/cgroup/system.slice/job.scope/memory.max": file("max\n"),
			"sys/fs/cgroup/system.slice/memory.max":           file("4294967296\n"),
		}, noSysctl, 4 << 30, true},
		{"linux, cgroup v2 in a container's namespace", "linux", fstest.MapFS{
			"proc/meminfo":             file(meminfo16G),
			"proc/self/cgroup":         file("0::/\n"),
			"sys/fs/cgroup/memory.max": file("2726297600\n"),
		}, noSysctl, 2726297600, true},
		{"linux, cgroup v2 without a limit", "linux", fstest.MapFS{
			"proc/meminfo":             file(meminfo16G),
			"proc/self/cgroup":         file("0::/\n"),
			"sys/fs/cgroup/memory.max": file("max\n"),
		}, noSysctl, 16 << 30, true},
		{"linux, cgroup v1 memory controller", "linux", fstest.MapFS{
			"proc/meminfo":     file(meminfo16G),
			"proc/self/cgroup": file("12:cpu,cpuacct:/docker/abc\n4:memory:/docker/abc\n"),
			"sys/fs/cgroup/memory/docker/abc/memory.limit_in_bytes": file("2147483648\n"),
			"sys/fs/cgroup/memory/memory.limit_in_bytes":            file("9223372036854771712\n"),
		}, noSysctl, 2 << 30, true},
		{"linux, cgroup v1 without a limit", "linux", fstest.MapFS{
			"proc/meminfo":     file(meminfo16G),
			"proc/self/cgroup": file("4:memory:/\n"),
			"sys/fs/cgroup/memory/memory.limit_in_bytes": file("9223372036854771712\n"),
		}, noSysctl, 16 << 30, true},
		{"linux, a limit above the machine's memory", "linux", fstest.MapFS{
			"proc/meminfo":             file(meminfo16G),
			"proc/self/cgroup":         file("0::/\n"),
			"sys/fs/cgroup/memory.max": file("68719476736\n"),
		}, noSysctl, 16 << 30, true},
		{"linux without /proc", "linux", noFiles, noSysctl, 0, false},
		{"windows", "windows", noFiles, noSysctl, 0, false},
	} {
		t.Run(c.name, func(t *testing.T) {
			got, ok := hostMemoryOf(c.goos, c.root, c.sysctl)
			if got != c.want || ok != c.ok {
				t.Errorf("hostMemoryOf = %d, %v, want %d, %v", got, ok, c.want, c.ok)
			}
		})
	}
}

// An instance's memory limit: 8 GiB without a setting, exactly the limit
// set (to a page), and the host's memory when asked for.
func TestMemoryCeiling(t *testing.T) {
	const least = 16 << 20
	if got := memoryCeiling(InstanceOptions{}, least); got != 8<<30 {
		t.Errorf("no limit set: %d, want 8 GiB", got)
	}
	if got := memoryCeiling(InstanceOptions{MaxMemoryBytes: 10<<30 + 12345}, least); got != 10<<30 {
		t.Errorf("10 GiB and a bit: %d, want 10 GiB", got)
	}
	if got := memoryCeiling(InstanceOptions{MaxMemoryBytes: 1 << 40}, least); got != 1<<40 {
		t.Errorf("1 TiB: %d, want 1 TiB", got)
	}
	if got := memoryCeiling(InstanceOptions{MaxMemoryBytes: 64 << 20}, least); got != 64<<20 {
		t.Errorf("64 MiB: %d, want 64 MiB", got)
	}
	if got := memoryCeiling(InstanceOptions{MaxMemoryBytes: 1 << 20}, least); got != 8<<30 {
		t.Errorf("below the memory the instance starts with: %d, want the default", got)
	}
	want := uint64(8 << 30)
	if size, ok := hostMemory(); ok {
		want = size / wasmPage * wasmPage
	}
	if got := memoryCeiling(InstanceOptions{MaxMemoryBytes: HostMemoryLimit}, least); got != want {
		t.Errorf("the host's memory: %d, want %d", got, want)
	}
}
