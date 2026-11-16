package internal

import "syscall"

// threadIDs reports that threadID tells the threads of the process apart.
const threadIDs = true

// threadID is the id of the thread the calling goroutine runs on.
func threadID() uint64 { return uint64(syscall.Gettid()) }
