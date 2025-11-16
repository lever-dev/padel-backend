package httputil

import (
	"fmt"
	"time"
)

func ParseTime(s string) (time.Time, error) {
	layouts := []string{
		"2006-01-02T15:04",    // 2025-11-04T18:30
		time.RFC3339,          // 2025-11-04T19:45:00Z
		"2006-01-02T15:04:05", // 2025-11-04T19:45:00
	}

	var lastErr error
	for _, layout := range layouts {
		if t, err := time.Parse(layout, s); err == nil {
			return t.UTC(), nil
		} else {
			lastErr = err
		}
	}

	return time.Time{}, fmt.Errorf(
		"invalid time format %q (supported: 2006-01-02T15:04, RFC3339): %w",
		s,
		lastErr,
	)
}
