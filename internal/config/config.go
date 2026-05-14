package config

import (
	"os"
	"time"

	"github.com/JustARandomBadDev/captive-portal-admin/internal/adminauth"
)

type Config struct {
	AppAddr           string
	DatabaseURL       string
	RadiusDatabaseURL string
	SessionSecret     string
	AdminSessionTTL   time.Duration
	AdminCookieSecure bool
}

func Load() Config {
	return Config{
		AppAddr:           envOrDefault("APP_ADDR", ":8080"),
		DatabaseURL:       os.Getenv("DATABASE_URL"),
		RadiusDatabaseURL: os.Getenv("RADIUS_DATABASE_URL"),
		SessionSecret:     os.Getenv("SESSION_SECRET"),
		AdminSessionTTL:   envDuration("ADMIN_SESSION_TTL", adminauth.DefaultSessionTTL),
		AdminCookieSecure: os.Getenv("ADMIN_COOKIE_SECURE") == "true",
	}
}

func envOrDefault(key, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	return value
}

func envDuration(key string, fallback time.Duration) time.Duration {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}

	duration, err := time.ParseDuration(value)
	if err != nil || duration <= 0 {
		return fallback
	}
	return duration
}
