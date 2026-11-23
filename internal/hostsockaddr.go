package internal

import (
	"encoding/binary"
	"fmt"
)

// The guest's question about a socket address (perl_numbers.cc): words of
// 64 bits, the direction first. To the host: the kind, the port, the IPv6
// flow information and scope, then the address bytes (4 or 16 bytes of an
// IP address, the name of a Unix-domain socket). From the host: the bytes
// of the host's struct sockaddr; the answer is the words and the bytes.
const (
	sockaddrToHost   = 0
	sockaddrFromHost = 1

	sockaddrIPv4 = 1
	sockaddrIPv6 = 2
	sockaddrUnix = 3
)

// guestSockaddr is a socket address as the question carries it.
type guestSockaddr struct {
	kind, port, flowinfo, scope uint64
	data                        []byte
}

// hostSockaddrAnswer answers the guest's question: [1] and the address on
// the other side, [1] alone when the other side has no form of it, [0] and
// the message for a malformed question.
func hostSockaddrAnswer(req []byte) []byte {
	if len(req) < 8 {
		return append([]byte{0}, "a socket address question without its direction"...)
	}
	switch binary.LittleEndian.Uint64(req) {
	case sockaddrToHost:
		if len(req) < 8*5 {
			return append([]byte{0}, fmt.Sprintf("a guest socket address of %d bytes", len(req))...)
		}
		a := guestSockaddr{
			kind:     binary.LittleEndian.Uint64(req[8:]),
			port:     binary.LittleEndian.Uint64(req[16:]),
			flowinfo: binary.LittleEndian.Uint64(req[24:]),
			scope:    binary.LittleEndian.Uint64(req[32:]),
			data:     req[40:],
		}
		return append([]byte{1}, hostSockaddrBytes(a)...)
	case sockaddrFromHost:
		a, ok := guestSockaddrOfHost(req[8:])
		if !ok {
			return []byte{1}
		}
		out := make([]byte, 1+8*4, 1+8*4+len(a.data))
		out[0] = 1
		binary.LittleEndian.PutUint64(out[1:], a.kind)
		binary.LittleEndian.PutUint64(out[9:], a.port)
		binary.LittleEndian.PutUint64(out[17:], a.flowinfo)
		binary.LittleEndian.PutUint64(out[25:], a.scope)
		return append(out, a.data...)
	}
	return append([]byte{0}, "a socket address question of an unknown direction"...)
}
