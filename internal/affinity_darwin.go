package internal

import "syscall"

// threadIDs reports that threadID tells the threads of the process apart.
const threadIDs = true

// threadID is the id of the thread the calling goroutine runs on (the
// kernel's thread_selfid, which pthread_threadid_np answers).
func threadID() uint64 {
	id, _, _ := syscall.RawSyscall(syscall.SYS_THREAD_SELFID, 0, 0, 0)
	return uint64(id)
}
