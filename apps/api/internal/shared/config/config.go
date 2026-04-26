package config

import "os"

type Config struct {
	AppEnv              string
	AppPort             string
	AuthMockEmail       string
	AuthMockPassword    string
	AuthMockAccessToken string
}

func FromEnv() Config {
	return Config{
		AppEnv:              envOrDefault("APP_ENV", "local"),
		AppPort:             envOrDefault("APP_PORT", "8080"),
		AuthMockEmail:       envOrDefault("AUTH_MOCK_EMAIL", "admin@example.local"),
		AuthMockPassword:    envOrDefault("AUTH_MOCK_PASSWORD", "local-only-mock-password"),
		AuthMockAccessToken: envOrDefault("AUTH_MOCK_ACCESS_TOKEN", "local-dev-access-token"),
	}
}

func envOrDefault(key string, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	return value
}
