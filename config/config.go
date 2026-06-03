package config

import (
	"os"
	"strconv"
	"time"

	_ "github.com/joho/godotenv/autoload"
)

type Config struct {
	DatabasePath  string
	JWTSecret     string
	AdminPassword string
	ServerPort    string
	ServerURL     string
	SessionTTL    time.Duration
	Environment   string
}

func Load() *Config {
	return &Config{
		DatabasePath:  getEnv("DATABASE_PATH", "./vault.db"),
		JWTSecret:     getEnv("JWT_SECRET", "change-me-in-production"),
		AdminPassword: getEnv("ADMIN_PASSWORD", "admin123"),
		ServerPort:    getEnv("PORT", "8080"),
		ServerURL:     getEnv("SERVER_URL", "http://localhost:8080"),
		SessionTTL:    parseDuration(getEnv("SESSION_TTL", "8h")),
		Environment:   getEnv("ENV", "development"),
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func parseDuration(s string) time.Duration {
	d, err := time.ParseDuration(s)
	if err != nil {
		return 8 * time.Hour
	}
	return d
}

func getEnvInt(key string, fallback int) int {
	if v := os.Getenv(key); v != "" {
		if i, err := strconv.Atoi(v); err == nil {
			return i
		}
	}
	return fallback
}
