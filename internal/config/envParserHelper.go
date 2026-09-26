// Small helper to parse more easily the environnement values while logging encountered problems while parsing

package config

import (
	"log/slog"
	"os"
	"strconv"
)

func getEnvOrDefault(key string, defaultVal string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	} else if defaultVal != "" {
		slog.Warn("environment variable not set", "key", key, "defaultVal", defaultVal)
	}
	return defaultVal
}

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

func inRange(a int, b int, n int) bool {
	return n >= a && n <= b
}

func getEnvOrDefaultPort(key string, defaultVal int) int {
	if value := getEnvOrDefaultInt(key, defaultVal); inRange(1, 65535, value) {
		return value
	}
	slog.Warn("environment variable is not a valid port", "key", key, "defaultVal", defaultVal)
	return defaultVal
}
