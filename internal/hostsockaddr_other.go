//go:build !darwin && !linux

package internal

func hostSockaddrBytes(guestSockaddr) []byte { return nil }

func guestSockaddrOfHost([]byte) (guestSockaddr, bool) { return guestSockaddr{}, false }
