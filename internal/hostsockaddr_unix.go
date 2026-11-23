//go:build darwin || linux

package internal

import (
	"bytes"
	"encoding/binary"
	"syscall"
)

// hostSockaddrBytes renders a as the host C library's struct sockaddr_in,
// sockaddr_in6 or sockaddr_un, nil for one the host has no form of.
func hostSockaddrBytes(a guestSockaddr) []byte {
	var raw []byte
	switch a.kind {
	case sockaddrIPv4:
		if len(a.data) != 4 {
			return nil
		}
		raw = make([]byte, syscall.SizeofSockaddrInet4)
		putHostFamily(raw, syscall.AF_INET)
		binary.BigEndian.PutUint16(raw[2:], uint16(a.port))
		copy(raw[4:8], a.data)
	case sockaddrIPv6:
		if len(a.data) != 16 {
			return nil
		}
		raw = make([]byte, syscall.SizeofSockaddrInet6)
		putHostFamily(raw, syscall.AF_INET6)
		binary.BigEndian.PutUint16(raw[2:], uint16(a.port))
		binary.BigEndian.PutUint32(raw[4:], uint32(a.flowinfo))
		copy(raw[8:24], a.data)
		binary.NativeEndian.PutUint32(raw[24:], uint32(a.scope))
	case sockaddrUnix:
		if len(a.data) > syscall.SizeofSockaddrUnix-unixPathOffset {
			return nil
		}
		raw = make([]byte, syscall.SizeofSockaddrUnix)
		putHostFamily(raw, syscall.AF_UNIX)
		copy(raw[unixPathOffset:], a.data)
		if len(a.data) == 0 || a.data[0] == 0 {
			raw = raw[:unixPathOffset+len(a.data)]
		}
	default:
		return nil
	}
	setHostSockaddrLen(raw)
	return raw
}

// guestSockaddrOfHost reads a struct sockaddr of the host's C library.
func guestSockaddrOfHost(raw []byte) (guestSockaddr, bool) {
	if len(raw) < 2 {
		return guestSockaddr{}, false
	}
	switch hostFamily(raw) {
	case syscall.AF_INET:
		if len(raw) < syscall.SizeofSockaddrInet4 {
			return guestSockaddr{}, false
		}
		return guestSockaddr{kind: sockaddrIPv4, port: uint64(binary.BigEndian.Uint16(raw[2:])),
			data: append([]byte(nil), raw[4:8]...)}, true
	case syscall.AF_INET6:
		if len(raw) < syscall.SizeofSockaddrInet6 {
			return guestSockaddr{}, false
		}
		return guestSockaddr{kind: sockaddrIPv6, port: uint64(binary.BigEndian.Uint16(raw[2:])),
			flowinfo: uint64(binary.BigEndian.Uint32(raw[4:])),
			data:     append([]byte(nil), raw[8:24]...),
			scope:    uint64(binary.NativeEndian.Uint32(raw[24:]))}, true
	case syscall.AF_UNIX:
		name := raw[unixPathOffset:]
		if len(name) > syscall.SizeofSockaddrUnix-unixPathOffset {
			name = name[:syscall.SizeofSockaddrUnix-unixPathOffset]
		}
		if len(name) > 0 && name[0] != 0 {
			if end := bytes.IndexByte(name, 0); end >= 0 {
				name = name[:end]
			}
		}
		return guestSockaddr{kind: sockaddrUnix, data: append([]byte(nil), name...)}, true
	}
	return guestSockaddr{}, false
}
