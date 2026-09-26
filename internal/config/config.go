// Handle the loading of the '.env' variables through custom struct to prevent changes during run

package config

import (
	"errors"
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	Port    int
	Address string
}

func LoadConfig() (*Config, error) {

	if err := godotenv.Load(); err != nil && !errors.Is(err, os.ErrNotExist) {
		return nil, fmt.Errorf("loading .env file failed: %w", err)
	}

	cfg := &Config{
		Port:    getEnvOrDefaultPort("PORT", 8080),
		Address: getEnvOrDefault("ADDRESS", "localhost"),
	}
	return cfg, nil
}
