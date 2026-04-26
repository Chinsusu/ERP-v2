package config

import "os"

type Config struct {
	AppEnv  string
	AppPort string
}

func FromEnv() Config {
	return Config{
		AppEnv:  envOrDefault("APP_ENV", "local"),
		AppPort: envOrDefault("APP_PORT", "8080"),
	}
}

func envOrDefault(key string, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	return value
}
