//go:build !linux && !darwin

package internal

// threadIDs reports that threadID tells the threads of the process apart:
// not here, where no native code runs either.
const threadIDs = false

func threadID() uint64 { return 0 }
