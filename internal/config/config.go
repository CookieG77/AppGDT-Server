// Handle the loading of the '.env' variables through custom struct to prevent changes during run

package config

import (
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	Port        int
	Address     string
	DatabaseCfg *DatabaseConfig
	HashingCfg  *HashingConfig
	TokenCfg    *TokenConfig
	LoginCfg    *LoginLimitConfig
	TLSCfg      *TLSConfig
}

type DatabaseConfig struct {
	User     string
	Password string
	Database string
	Host     string
	Port     int
}

type HashingConfig struct {
	Cost int
}

type TokenConfig struct {
	Secret string
	TTL    time.Duration
}

// TLSConfig enables HTTPS when a certificate and its private key are given.
type TLSConfig struct {
	CertFile string
	KeyFile  string
}

// Enabled reports whether the API must be served over HTTPS.
func (t *TLSConfig) Enabled() bool {
	return t.CertFile != "" && t.KeyFile != ""
}

type LoginLimitConfig struct {
	MaxPerEmail int
	MaxPerIP    int
	Window      time.Duration
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

	bcryptCost := getEnvOrDefaultInt("BCRYPT_COST", 12)
	if !inRange(10, 14, bcryptCost) { // Enforced minimal hash security and prevent absurd level of hashing
		return nil, fmt.Errorf("BCRYPT_COST must be between %d and %d", 10, 14)
	}

	hashingCfg := &HashingConfig{
		Cost: bcryptCost,
	}

	secretJWT, err := requireEnv("JWT_SECRET")
	if err != nil {
		return nil, err
	}
	if !inRange(32, 256, len(secretJWT)) {
		return nil, fmt.Errorf("JWT_SECRET length must be between 32 and 256 inclusive")
	}

	ttl := getEnvOrDefaultDuration("JWT_TTL", time.Hour)
	if ttl < time.Minute*5 || ttl > time.Hour*24 { // Prevent absurdly low or high jwt validity duration
		return nil, fmt.Errorf("JWT_TTL must be between 5 minute and 24 hours")
	}

	tokenCfg := &TokenConfig{
		Secret: secretJWT,
		TTL:    ttl,
	}

	maxPerEmail := getEnvOrDefaultInt("LOGIN_MAX_FAILURES_PER_EMAIL", 5)
	if !inRange(3, 20, maxPerEmail) { // Too low locks out users after a few typos, too high allows guessing
		return nil, fmt.Errorf("LOGIN_MAX_FAILURES_PER_EMAIL must be between %d and %d", 3, 20)
	}

	maxPerIP := getEnvOrDefaultInt("LOGIN_MAX_FAILURES_PER_IP", 50)
	if !inRange(10, 1000, maxPerIP) { // Several users may share the same IP (company, school network)
		return nil, fmt.Errorf("LOGIN_MAX_FAILURES_PER_IP must be between %d and %d", 10, 1000)
	}

	loginWindow := getEnvOrDefaultDuration("LOGIN_FAILURE_WINDOW", 15*time.Minute)
	if loginWindow < time.Minute || loginWindow > time.Hour*24 {
		return nil, fmt.Errorf("LOGIN_FAILURE_WINDOW must be between 1 minute and 24 hours")
	}

	loginCfg := &LoginLimitConfig{
		MaxPerEmail: maxPerEmail,
		MaxPerIP:    maxPerIP,
		Window:      loginWindow,
	}

	// HTTPS is optional: both files must be given together, or none
	tlsCfg := &TLSConfig{
		CertFile: getEnvOrDefault("TLS_CERT_FILE", ""),
		KeyFile:  getEnvOrDefault("TLS_KEY_FILE", ""),
	}
	if (tlsCfg.CertFile == "") != (tlsCfg.KeyFile == "") {
		return nil, fmt.Errorf("TLS_CERT_FILE and TLS_KEY_FILE must be set together")
	}

	cfg := &Config{
		Port:        getEnvOrDefaultPort("PORT", 8080),
		Address:     getEnvOrDefault("ADDRESS", "localhost"),
		DatabaseCfg: dbCfg,
		HashingCfg:  hashingCfg,
		TokenCfg:    tokenCfg,
		LoginCfg:    loginCfg,
		TLSCfg:      tlsCfg,
	}

	return cfg, nil
}
