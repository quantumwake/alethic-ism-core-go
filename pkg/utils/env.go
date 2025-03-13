package utils

import (
	"os"
	"strconv"
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
