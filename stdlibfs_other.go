//go:build !darwin && !linux

package perl

// Sys gives no metadata beyond the FileInfo on a host the runtime reads
// none from (syscall.Stat_t is read on darwin and linux): every file of
// the mounts then has device and inode 0.
func (fi roInfo) Sys() any { return nil }
