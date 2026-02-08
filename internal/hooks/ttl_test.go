package hooks

import (
	"testing"
	"time"
)

func TestParseTTL(t *testing.T) {
	tests := []struct {
		tag  string
		want time.Duration
	}{
		{"5m", 5 * time.Minute},
		{"30m", 30 * time.Minute},
		{"1h", time.Hour},
		{"6h", 6 * time.Hour},
		{"24h", 24 * time.Hour},
		{"1d", 24 * time.Hour},
		{"1w", 7 * 24 * time.Hour},
		{"1h30m", time.Hour + 30*time.Minute},
		{"2d12h", 2*24*time.Hour + 12*time.Hour},
		{"30s", 30 * time.Second},
		{"1d12h30m", 24*time.Hour + 12*time.Hour + 30*time.Minute},
		// Invalid tags
		{"latest", -1},
		{"v1.0.0", -1},
		{"", -1},
		{"abc", -1},
		{"sha-abc123", -1},
	}

	for _, tt := range tests {
		t.Run(tt.tag, func(t *testing.T) {
			got := ParseTTL(tt.tag)
			if got != tt.want {
				t.Errorf("ParseTTL(%q) = %v, want %v", tt.tag, got, tt.want)
			}
		})
	}
}

func TestClampTTL(t *testing.T) {
	defaultTTL := time.Hour
	maxTTL := 24 * time.Hour

	tests := []struct {
		name string
		d    time.Duration
		want time.Duration
	}{
		{"negative uses default", -1, defaultTTL},
		{"zero uses default", 0, defaultTTL},
		{"within range", 6 * time.Hour, 6 * time.Hour},
		{"exceeds max is clamped", 48 * time.Hour, maxTTL},
		{"exactly max", maxTTL, maxTTL},
		{"exactly default", defaultTTL, defaultTTL},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ClampTTL(tt.d, defaultTTL, maxTTL)
			if got != tt.want {
				t.Errorf("ClampTTL(%v) = %v, want %v", tt.d, got, tt.want)
			}
		})
	}
}
