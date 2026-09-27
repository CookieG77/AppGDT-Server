// Small helper to parse more easily the environnement values while logging encountered problems while parsing

package config

import (
	"fmt"
	"log/slog"
	"os"
	"strconv"
	"time"
)

// getEnvOrDefault returns the requested env variable if present and valid.
// Otherwise, returns the given default value.
func getEnvOrDefault(key string, defaultVal string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	} else if defaultVal != "" {
		slog.Warn("environment variable not set", "key", key, "defaultVal", defaultVal)
	}
	return defaultVal
}

// requireEnv returns the requested env variable if present otherwise, returns an error
func requireEnv(key string) (string, error) {
	if value, exists := os.LookupEnv(key); exists && value != "" {
		return value, nil
	}
	return "", fmt.Errorf("environment variable '%s' not set", key)
}

// getEnvOrDefaultInt returns the requested int env variable if present and valid.
// Otherwise, returns the given int default value.
func getEnvOrDefaultInt(key string, defaultVal int) int {
	if value, exists := os.LookupEnv(key); exists {
		if i, err := strconv.Atoi(value); value == "" {
			slog.Warn("environment variable not set", "key", key, "defaultVal", defaultVal)
			return defaultVal
		} else if err != nil {
			slog.Warn("environment variable is not a valid integer", "key", key, "defaultVal", defaultVal)
		} else {
			return i
		}
	}
	return defaultVal
}

// getEnvOrDefaultBool returns the requested boolean env variable if present and valid.
// Otherwise, returns the given boolean default value.
func getEnvOrDefaultBool(key string, defaultVal bool) bool {
	if value, exists := os.LookupEnv(key); exists {
		if b, err := strconv.ParseBool(value); value == "" {
			slog.Warn("environment variable not set", "key", key, "defaultVal", defaultVal)
		} else if err != nil {
			slog.Warn("environment variable is not a valid boolean", "key", key, "defaultVal", defaultVal)
		} else {
			return b
		}
	}
	return defaultVal
}

// inRange returns true if 'n' is in [a, b], false otherwise.
func inRange(a int, b int, n int) bool {
	return n >= a && n <= b
}

// getEnvOrDefaultPort returns the requested port env variable if present and valid.
// Otherwise, returns the given port default value.
func getEnvOrDefaultPort(key string, defaultVal int) int {
	if value := getEnvOrDefaultInt(key, defaultVal); inRange(1, 65535, value) {
		return value
	}
	slog.Warn("environment variable is not a valid port", "key", key, "defaultVal", defaultVal)
	return defaultVal
}

// getEnvOrDefaultDuration returns the requested duration env variable if present and valid.
// Otherwise, returns the given duration default value.
func getEnvOrDefaultDuration(key string, defaultVal time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		duration, err := time.ParseDuration(value)
		if err == nil {
			return duration
		}
		slog.Warn("environment variable is not a valid time", "key", key, "defaultVal", defaultVal)
	}
	return defaultVal
}
