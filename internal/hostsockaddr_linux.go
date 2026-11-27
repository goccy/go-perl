package internal

import "encoding/binary"

// Linux's struct sockaddr starts with the family as a native 16-bit word.
const unixPathOffset = 2

func putHostFamily(raw []byte, family int) { binary.NativeEndian.PutUint16(raw, uint16(family)) }

func hostFamily(raw []byte) int { return int(binary.NativeEndian.Uint16(raw)) }

func setHostSockaddrLen([]byte) {}
