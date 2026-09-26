// Handle the loading of the '.env' variables through custom struct to prevent changes during run

package config

import (
	"errors"
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	Port        int
	Address     string
	DatabaseCfg *DatabaseConfig
}

type DatabaseConfig struct {
	User     string
	Password string
	Database string
	Host     string
	Port     int
}

func LoadConfig() (*Config, error) {

	if err := godotenv.Load(); err != nil && !errors.Is(err, os.ErrNotExist) {
		return nil, fmt.Errorf("loading .env file failed: %w", err)
	}

	user, err := requireEnv("POSTGRES_USER")
	if err != nil {
		return nil, err
	}
	password, err := requireEnv("POSTGRES_PASSWORD")
	if err != nil {
		return nil, err
	}

	dbCfg := &DatabaseConfig{
		User:     user,
		Password: password,
		Database: getEnvOrDefault("POSTGRES_DB", "gdt"),
		Host:     getEnvOrDefault("POSTGRES_HOST", "localhost"),
		Port:     getEnvOrDefaultPort("DB_PORT", 5432),
	}

	cfg := &Config{
		Port:        getEnvOrDefaultPort("PORT", 8080),
		Address:     getEnvOrDefault("ADDRESS", "localhost"),
		DatabaseCfg: dbCfg,
	}

	return cfg, nil
}
