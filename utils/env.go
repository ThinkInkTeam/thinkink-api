package utils

import "os"

// GetEnvWithDefault returns the environment variable value or a default if not set
func GetEnvWithDefault(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}
