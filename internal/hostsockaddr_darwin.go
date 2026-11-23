package internal

// darwin's struct sockaddr starts with a length byte, then the family byte.
const unixPathOffset = 2

func putHostFamily(raw []byte, family int) { raw[1] = byte(family) }

func hostFamily(raw []byte) int { return int(raw[1]) }

func setHostSockaddrLen(raw []byte) { raw[0] = byte(len(raw)) }
