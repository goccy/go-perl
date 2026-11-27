package perl

import "syscall"

// Sys describes a file of a mounted tree as the runtime reads a host
// file's metadata (syscall.Stat_t): an inode number of its own, on a device
// of the mounts' own, so that programs telling files apart by device and
// inode (File::Find, File::Copy) see distinct files.
func (fi roInfo) Sys() any {
	ts := syscall.NsecToTimespec(fi.modTime.UnixNano())
	return &syscall.Stat_t{Dev: treeDev, Ino: fi.ino, Nlink: 1, Size: fi.size,
		Atim: ts, Mtim: ts, Ctim: ts, Blksize: treeBlockSize, Blocks: (fi.size + 511) / 512}
}
