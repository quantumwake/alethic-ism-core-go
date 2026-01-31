package utils

import (
	"os"
	"strconv"
	"time"
)

func StringFromEnvWithDefault(key string, defaultValue string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return defaultValue
}

func Int64FromEnvWithDefault(key string, defaultValue int64) int64 {
	if value, ok := os.LookupEnv(key); ok {
		if intValue, err := strconv.ParseInt(value, 10, 64); err != nil {
			return defaultValue
		} else {
			return intValue
		}
	}
	return defaultValue
}

func DurationFromEnvWithDefault(key string, defaultValue time.Duration) time.Duration {
	if value, ok := os.LookupEnv(key); ok {
		if durationValue, err := time.ParseDuration(value); err == nil {
			return durationValue
		}
	}

	return defaultValue
}
