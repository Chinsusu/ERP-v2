package config

import (
	"os"
	"strings"
)

type Config struct {
	AppEnv              string
	AppPort             string
	DatabaseURL         string
	AuthMockEmail       string
	AuthMockPassword    string
	AuthMockAccessToken string
	S3Endpoint          string
	S3Bucket            string
	S3AccessKey         string
	S3SecretKey         string
	S3UseSSL            bool
	S3UsePathStyle      bool
}

func FromEnv() Config {
	return Config{
		AppEnv:              envOrDefault("APP_ENV", "local"),
		AppPort:             envOrDefault("APP_PORT", "8080"),
		DatabaseURL:         envOrDefault("DATABASE_URL", ""),
		AuthMockEmail:       envOrDefault("AUTH_MOCK_EMAIL", "admin@example.local"),
		AuthMockPassword:    envOrDefault("AUTH_MOCK_PASSWORD", "local-only-mock-password"),
		AuthMockAccessToken: envOrDefault("AUTH_MOCK_ACCESS_TOKEN", "local-dev-access-token"),
		S3Endpoint:          envOrDefault("S3_ENDPOINT", "http://minio:9000"),
		S3Bucket:            envOrDefault("S3_BUCKET", "erp-dev"),
		S3AccessKey:         envOrDefault("S3_ACCESS_KEY", "minio"),
		S3SecretKey:         envOrDefault("S3_SECRET_KEY", "minio123"),
		S3UseSSL:            envBoolOrDefault("S3_USE_SSL", false),
		S3UsePathStyle:      envBoolOrDefault("S3_USE_PATH_STYLE", true),
	}
}

func (c Config) StaticAuthAccessToken() string {
	if !AllowsStaticAuthAccessToken(c.AppEnv) {
		return ""
	}

	return c.AuthMockAccessToken
}

func AllowsStaticAuthAccessToken(appEnv string) bool {
	switch strings.ToLower(strings.TrimSpace(appEnv)) {
	case "", "local", "dev", "development", "test":
		return true
	default:
		return false
	}
}

func envOrDefault(key string, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	return value
}

func envBoolOrDefault(key string, fallback bool) bool {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	switch value {
	case "1", "true", "TRUE", "True", "yes", "YES", "Yes":
		return true
	case "0", "false", "FALSE", "False", "no", "NO", "No":
		return false
	default:
		return fallback
	}
}
