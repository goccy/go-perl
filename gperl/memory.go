package gperl

import (
	"fmt"
	"strconv"
	"strings"

	perl "github.com/goccy/go-perl"
)

// memoryLimitEnv names the environment variable of gperl's memory limit:
// how much linear memory the interpreter of a run (and of a binary gperl
// built) may use. Unset or empty: the library's default, 8 GiB. "host":
// the memory the host gives the process (perl.HostMemoryLimit). A number
// of bytes, optionally followed by K, M, G or T (KiB, MiB, GiB, TiB): that
// limit.
const memoryLimitEnv = "GPERL_MAX_MEMORY"

// parseMemoryLimit is the Config.MaxMemoryBytes value of GPERL_MAX_MEMORY.
func parseMemoryLimit(value string) (int, error) {
	if value == "" {
		return 0, nil
	}
	if value == "host" {
		return perl.HostMemoryLimit, nil
	}
	digits, shift := value, uint(0)
	switch value[len(value)-1] {
	case 'K', 'k':
		shift = 10
	case 'M', 'm':
		shift = 20
	case 'G', 'g':
		shift = 30
	case 'T', 't':
		shift = 40
	}
	if shift > 0 {
		digits = value[:len(value)-1]
	}
	n, err := strconv.ParseUint(digits, 10, 63)
	if err != nil || n == 0 || strings.HasPrefix(digits, "+") || n > (1<<63-1)>>shift {
		return 0, fmt.Errorf("%s=%q: want a number of bytes (with K, M, G or T), or host", memoryLimitEnv, value)
	}
	return int(n << shift), nil
}
