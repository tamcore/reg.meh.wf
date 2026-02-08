package hooks

import (
	"regexp"
	"strconv"
	"time"
)

// durationPattern matches tags like "5m", "1h", "24h", "1d", "1w", "30m", "1h30m".
var durationPattern = regexp.MustCompile(
	`^(?:(\d+)w)?(?:(\d+)d)?(?:(\d+)h)?(?:(\d+)m)?(?:(\d+)s)?$`,
)

// ParseTTL parses a duration string from an image tag.
// Returns -1 if the tag is not a valid duration.
func ParseTTL(tag string) time.Duration {
	if tag == "" {
		return -1
	}

	matches := durationPattern.FindStringSubmatch(tag)
	if matches == nil {
		return -1
	}

	// At least one group must be non-empty.
	hasValue := false
	for _, m := range matches[1:] {
		if m != "" {
			hasValue = true
			break
		}
	}
	if !hasValue {
		return -1
	}

	var d time.Duration
	if matches[1] != "" {
		n, _ := strconv.Atoi(matches[1])
		d += time.Duration(n) * 7 * 24 * time.Hour
	}
	if matches[2] != "" {
		n, _ := strconv.Atoi(matches[2])
		d += time.Duration(n) * 24 * time.Hour
	}
	if matches[3] != "" {
		n, _ := strconv.Atoi(matches[3])
		d += time.Duration(n) * time.Hour
	}
	if matches[4] != "" {
		n, _ := strconv.Atoi(matches[4])
		d += time.Duration(n) * time.Minute
	}
	if matches[5] != "" {
		n, _ := strconv.Atoi(matches[5])
		d += time.Duration(n) * time.Second
	}

	return d
}

// ClampTTL applies default and max TTL limits.
func ClampTTL(d, defaultTTL, maxTTL time.Duration) time.Duration {
	if d <= 0 {
		return defaultTTL
	}
	if d > maxTTL {
		return maxTTL
	}
	return d
}
