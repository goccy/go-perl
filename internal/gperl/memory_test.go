package gperl

import (
	"testing"

	perl "github.com/goccy/go-perl"
)

// GPERL_MAX_MEMORY: unset is the default, "host" the host's memory, a size
// exactly that size; anything else is refused.
func TestParseMemoryLimit(t *testing.T) {
	for _, c := range []struct {
		value string
		want  int
	}{
		{"", 0},
		{"host", perl.HostMemoryLimit},
		{"1073741824", 1 << 30},
		{"512M", 512 << 20},
		{"16g", 16 << 30},
		{"2T", 2 << 40},
		{"64k", 64 << 10},
	} {
		got, err := parseMemoryLimit(c.value)
		if err != nil || got != c.want {
			t.Errorf("parseMemoryLimit(%q) = %d, %v, want %d", c.value, got, err, c.want)
		}
	}
	for _, value := range []string{"0", "-1", "+5", "1.5G", "G", "lots", "Host", "16GB", "99999999999T"} {
		if got, err := parseMemoryLimit(value); err == nil {
			t.Errorf("parseMemoryLimit(%q) = %d, want an error", value, got)
		}
	}
}

// A run takes its memory limit from GPERL_MAX_MEMORY.
func TestHostConfigTakesTheMemoryLimit(t *testing.T) {
	t.Setenv(memoryLimitEnv, "host")
	cfg, err := HostConfig()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.MaxMemoryBytes != perl.HostMemoryLimit {
		t.Errorf("MaxMemoryBytes = %d, want HostMemoryLimit", cfg.MaxMemoryBytes)
	}
	t.Setenv(memoryLimitEnv, "lots")
	if _, err := HostConfig(); err == nil {
		t.Error("HostConfig accepted GPERL_MAX_MEMORY=lots")
	}
	t.Setenv(memoryLimitEnv, "")
	cfg, err = HostConfig()
	if err != nil || cfg.MaxMemoryBytes != 0 {
		t.Errorf("unset: MaxMemoryBytes = %d, %v, want 0 (the default)", cfg.MaxMemoryBytes, err)
	}
}
