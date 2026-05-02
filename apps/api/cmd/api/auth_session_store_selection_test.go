package main

import (
	"context"
	"database/sql"
	"os"
	"testing"
	"time"

	"github.com/Chinsusu/ERP-v2/apps/api/internal/shared/config"

	_ "github.com/jackc/pgx/v5/stdlib"
)

func TestNewRuntimeSessionManagerFallsBackToMemoryWithoutDatabaseURL(t *testing.T) {
	manager, closeManager, err := newRuntimeSessionManager(config.Config{
		AppEnv:              "dev",
		AuthMockEmail:       "admin@example.local",
		AuthMockPassword:    "local-only-mock-password",
		AuthMockAccessToken: "local-dev-access-token",
	}, func() time.Time { return time.Date(2026, 5, 2, 12, 0, 0, 0, time.UTC) })
	if err != nil {
		t.Fatalf("newRuntimeSessionManager() error = %v", err)
	}
	if closeManager != nil {
		t.Fatal("closeManager is not nil, want nil for in-memory auth manager")
	}

	principal, ok := manager.AuthenticateAccessToken("local-dev-access-token")
	if !ok || principal.Email != "admin@example.local" {
		t.Fatalf("principal = %+v, authenticated = %v", principal, ok)
	}
}

func TestNewRuntimeSessionManagerUsesPostgresWhenDatabaseURLConfigured(t *testing.T) {
	databaseURL := os.Getenv("ERP_TEST_DATABASE_URL")
	if databaseURL == "" {
		t.Skip("ERP_TEST_DATABASE_URL is not set")
	}

	ctx := context.Background()
	db, err := sql.Open("pgx", databaseURL)
	if err != nil {
		t.Fatalf("open database: %v", err)
	}
	defer db.Close()
	if err := seedRuntimeSessionManagerFixture(ctx, db); err != nil {
		t.Fatalf("seed fixture: %v", err)
	}

	now := time.Date(2026, 5, 2, 12, 0, 0, 0, time.UTC)
	cfg := config.Config{
		AppEnv:              "dev",
		DatabaseURL:         databaseURL,
		AuthMockEmail:       "admin@example.local",
		AuthMockPassword:    "local-only-mock-password",
		AuthMockAccessToken: "local-dev-access-token",
	}
	managerA, closeManagerA, err := newRuntimeSessionManager(cfg, func() time.Time { return now })
	if err != nil {
		t.Fatalf("newRuntimeSessionManager(managerA) error = %v", err)
	}
	if closeManagerA == nil {
		t.Fatal("closeManagerA is nil, want database close function")
	}
	session, failure, ok := managerA.Login("admin@example.local", "local-only-mock-password")
	if !ok {
		t.Fatalf("login rejected: %+v", failure)
	}
	if err := closeManagerA(); err != nil {
		t.Fatalf("closeManagerA() error = %v", err)
	}

	managerB, closeManagerB, err := newRuntimeSessionManager(cfg, func() time.Time { return now })
	if err != nil {
		t.Fatalf("newRuntimeSessionManager(managerB) error = %v", err)
	}
	defer func() {
		if err := closeManagerB(); err != nil {
			t.Fatalf("closeManagerB() error = %v", err)
		}
	}()

	principal, ok := managerB.AuthenticateAccessToken(session.AccessToken)
	if !ok || principal.Email != "admin@example.local" {
		t.Fatalf("principal = %+v, authenticated = %v", principal, ok)
	}
}

func seedRuntimeSessionManagerFixture(ctx context.Context, db *sql.DB) error {
	if _, err := db.ExecContext(ctx, `
INSERT INTO core.organizations (id, code, name, status)
VALUES ($1::uuid, 'LOCAL_DEV', 'Local Dev', 'active')
ON CONFLICT (id) DO UPDATE
SET code = EXCLUDED.code,
    name = EXCLUDED.name,
    status = EXCLUDED.status,
    updated_at = now()`,
		localAuditOrgID,
	); err != nil {
		return err
	}
	if _, err := db.ExecContext(ctx, `DELETE FROM core.auth_login_failures WHERE org_id = $1::uuid`, localAuditOrgID); err != nil {
		return err
	}
	_, err := db.ExecContext(ctx, `DELETE FROM core.auth_sessions WHERE org_id = $1::uuid`, localAuditOrgID)
	return err
}
