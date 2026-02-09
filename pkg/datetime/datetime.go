// Package datetime provides centralized date and time functions
// with timezone support for the application.
// This ensures all dates are handled consistently using the configured timezone.
package datetime

import (
	"sync"
	"time"
)

const (
	// DefaultTimezone is the default timezone for Colombia (GMT-5).
	DefaultTimezone = "America/Bogota"
)

var (
	location     *time.Location
	locationOnce sync.Once
	timezone     string = DefaultTimezone
)

// SetTimezone sets the timezone to use for all date/time operations.
// This should be called during application initialization before any
// date operations are performed. If not called, DefaultTimezone is used.
func SetTimezone(tz string) {
	timezone = tz
	// Reset the location so it gets reloaded with new timezone
	locationOnce = sync.Once{}
	location = nil
}

// GetTimezone returns the currently configured timezone.
func GetTimezone() string {
	return timezone
}

// Location returns the configured timezone location.
// It loads the location lazily and caches it for performance.
func Location() *time.Location {
	locationOnce.Do(func() {
		loc, err := time.LoadLocation(timezone)
		if err != nil {
			// Fallback to UTC if location cannot be loaded
			loc = time.UTC
		}
		location = loc
	})
	return location
}

// Now returns the current time in the configured timezone.
// Use this instead of time.Now() throughout the application.
func Now() time.Time {
	return time.Now().In(Location())
}

// Today returns today's date at midnight in the configured timezone.
func Today() time.Time {
	now := Now()
	return time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, Location())
}

// StartOfDay returns the given time set to the start of the day (00:00:00)
// in the configured timezone.
func StartOfDay(t time.Time) time.Time {
	t = t.In(Location())
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, Location())
}

// EndOfDay returns the given time set to the end of the day (23:59:59.999999999)
// in the configured timezone.
func EndOfDay(t time.Time) time.Time {
	t = t.In(Location())
	return time.Date(t.Year(), t.Month(), t.Day(), 23, 59, 59, 999999999, Location())
}

// InLocation converts a time to the configured timezone.
func InLocation(t time.Time) time.Time {
	return t.In(Location())
}

// Parse parses a time string using the given layout in the configured timezone.
func Parse(layout, value string) (time.Time, error) {
	return time.ParseInLocation(layout, value, Location())
}

// Format formats a time in the configured timezone using the given layout.
func Format(t time.Time, layout string) string {
	return t.In(Location()).Format(layout)
}

// DateOnly extracts only the date part (year, month, day) from a time
// and returns it at midnight in the configured timezone.
func DateOnly(t time.Time) time.Time {
	t = t.In(Location())
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, Location())
}

// IsSameDay checks if two times represent the same calendar day
// in the configured timezone.
func IsSameDay(t1, t2 time.Time) bool {
	t1 = t1.In(Location())
	t2 = t2.In(Location())
	return t1.Year() == t2.Year() && t1.Month() == t2.Month() && t1.Day() == t2.Day()
}
