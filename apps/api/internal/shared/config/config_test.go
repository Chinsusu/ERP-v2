package config

import "testing"

func TestStaticAuthAccessTokenAllowedOnlyForLocalLikeEnvironments(t *testing.T) {
	for _, appEnv := range []string{"", "local", "dev", "development", "test", " DEV "} {
		cfg := Config{AppEnv: appEnv, AuthMockAccessToken: "local-dev-access-token"}
		if got := cfg.StaticAuthAccessToken(); got != "local-dev-access-token" {
			t.Fatalf("StaticAuthAccessToken(%q) = %q, want local token", appEnv, got)
		}
	}
}

func TestStaticAuthAccessTokenDisabledForProductionLikeEnvironments(t *testing.T) {
	for _, appEnv := range []string{"staging", "stage", "prod", "production", "qa"} {
		cfg := Config{AppEnv: appEnv, AuthMockAccessToken: "local-dev-access-token"}
		if got := cfg.StaticAuthAccessToken(); got != "" {
			t.Fatalf("StaticAuthAccessToken(%q) = %q, want empty", appEnv, got)
		}
	}
}

func TestFromEnvReadsDatabaseURL(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgres://erp_dev:erp_dev@postgres:5432/erp_dev?sslmode=disable")

	cfg := FromEnv()

	if cfg.DatabaseURL != "postgres://erp_dev:erp_dev@postgres:5432/erp_dev?sslmode=disable" {
		t.Fatalf("DatabaseURL = %q, want configured URL", cfg.DatabaseURL)
	}
}
