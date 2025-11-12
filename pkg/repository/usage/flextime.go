package usage

import (
	"fmt"
	"time"
)

// FlexTime is a custom time type that can unmarshal from multiple time formats
// including ISO format with timezone (RFC3339) and custom formats without timezone.
type FlexTime struct {
	time.Time
}

// UnmarshalJSON implements the json.Unmarshaler interface.
// It attempts to parse the time string using multiple formats.
func (ft *FlexTime) UnmarshalJSON(data []byte) error {
	// Remove quotes from the JSON string
	if len(data) < 2 || data[0] != '"' || data[len(data)-1] != '"' {
		return fmt.Errorf("invalid time format: expected quoted string")
	}
	str := string(data[1 : len(data)-1])

	// Try multiple time formats
	formats := []string{
		time.RFC3339Nano,             // ISO format with timezone and nanoseconds (e.g., "2025-11-07T21:54:44.669032Z")
		time.RFC3339,                 // ISO format with timezone (e.g., "2025-11-07T21:54:44Z")
		"2006-01-02T15:04:05.999999", // Custom format without timezone
		"2006-01-02T15:04:05",        // Basic format without timezone
	}

	var err error
	for _, format := range formats {
		ft.Time, err = time.Parse(format, str)
		if err == nil {
			return nil
		}
	}

	// If none of the formats worked, return the last error
	return fmt.Errorf("unable to parse time from '%s': %w", str, err)
}

// MarshalJSON implements the json.Marshaler interface.
// It marshals the time in RFC3339Nano format (ISO format with timezone).
func (ft FlexTime) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf(`"%s"`, ft.Time.Format(time.RFC3339Nano))), nil
}
