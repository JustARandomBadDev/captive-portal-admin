package config

import "os"

type Config struct {
	AppAddr       string
	DatabaseURL   string
	SessionSecret string
}

func Load() Config {
	return Config{
		AppAddr:       envOrDefault("APP_ADDR", ":8080"),
		DatabaseURL:   os.Getenv("DATABASE_URL"),
		SessionSecret: os.Getenv("SESSION_SECRET"),
	}
}

func envOrDefault(key, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	return value
}
