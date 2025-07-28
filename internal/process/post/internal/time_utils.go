package internal

import (
	"fmt"
	"time"
)

// ParseScheduleTime attempts to parse schedule time in various formats
func ParseScheduleTime(scheduleAt string) (*time.Time, error) {
	if scheduleAt == "now" {
		now := time.Now()
		return &now, nil
	}

	// Try different time formats
	formats := []string{
		time.RFC3339,                // "2006-01-02T15:04:05Z07:00"
		"2006-01-02T15:04:05-07:00", // RFC3339 without Z
		"2006-01-02T15:04:05",       // Without timezone
		"2006-01-02T15:04",          // Without seconds and timezone
		"2006-01-02 15:04:05",       // Space separator
		"2006-01-02 15:04",          // Space separator without seconds
	}

	for _, format := range formats {
		if parsedTime, err := time.Parse(format, scheduleAt); err == nil {
			// If the format doesn't include timezone, assume local timezone
			if format == "2006-01-02T15:04:05" || format == "2006-01-02T15:04" ||
				format == "2006-01-02 15:04:05" || format == "2006-01-02 15:04" {
				// Add local timezone
				loc, _ := time.LoadLocation("Local")
				parsedTime = time.Date(
					parsedTime.Year(), parsedTime.Month(), parsedTime.Day(),
					parsedTime.Hour(), parsedTime.Minute(), parsedTime.Second(),
					0, loc,
				)
			}
			return &parsedTime, nil
		}
	}

	return nil, fmt.Errorf("invalid schedule_at format, use ISO8601 or 'now': %s", scheduleAt)
}
