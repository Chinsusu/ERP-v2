package auth

import (
	"context"
	"database/sql"
	"os"
	"testing"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
)

func TestPostgresLoginFailureStoreRequiresDatabase(t *testing.T) {
	store := NewPostgresLoginFailureStore(nil, PostgresLoginFailureStoreConfig{DefaultOrgID: testPostgresAuthSessionOrgID})
	now := time.Date(2026, 5, 2, 11, 0, 0, 0, time.UTC)

	if _, _, err := store.LockedUntil("admin@example.local", now); err == nil {
		t.Fatal("LockedUntil() error = nil, want database required error")
	}
	if _, err := store.RecordFailure("admin@example.local", now, LockoutPolicy{}); err == nil {
		t.Fatal("RecordFailure() error = nil, want database required error")
	}
	if err := store.Clear("admin@example.local"); err == nil {
		t.Fatal("Clear() error = nil, want database required error")
	}
}

func TestPostgresLoginFailureStorePersistsLockoutAcrossManagers(t *testing.T) {
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

	if err := seedPostgresLoginFailureFixture(ctx, db); err != nil {
		t.Fatalf("seed fixture: %v", err)
	}

	now := time.Date(2026, 5, 2, 11, 0, 0, 0, time.UTC)
	failureStore := NewPostgresLoginFailureStore(
		db,
		PostgresLoginFailureStoreConfig{DefaultOrgID: testPostgresAuthSessionOrgID},
	)
	managerA, err := NewSessionManagerWithStores(
		testConfig,
		func() time.Time { return now },
		NewInMemorySessionStore(),
		failureStore,
	)
	if err != nil {
		t.Fatalf("NewSessionManagerWithStores(managerA) error = %v", err)
	}

	for range defaultMaxFailedLogins {
		_, _, _ = managerA.Login("admin@example.local", "wrong-password!")
	}

	managerB, err := NewSessionManagerWithStores(
		testConfig,
		func() time.Time { return now },
		NewInMemorySessionStore(),
		failureStore,
	)
	if err != nil {
		t.Fatalf("NewSessionManagerWithStores(managerB) error = %v", err)
	}

	_, failure, ok := managerB.Login("admin@example.local", "local-only-mock-password")
	if ok {
		t.Fatal("login accepted while persisted failure store is locked")
	}
	if failure.Code != LoginFailureLocked {
		t.Fatalf("failure code = %q, want locked", failure.Code)
	}
	if failure.LockedUntil.IsZero() {
		t.Fatal("locked_until is empty")
	}

	if err := failureStore.Clear("admin@example.local"); err != nil {
		t.Fatalf("clear lockout: %v", err)
	}
	_, failure, ok = managerB.Login("admin@example.local", "local-only-mock-password")
	if !ok {
		t.Fatalf("login rejected after clear: %+v", failure)
	}
}

func seedPostgresLoginFailureFixture(ctx context.Context, db *sql.DB) error {
	if _, err := db.ExecContext(ctx, `
INSERT INTO core.organizations (id, code, name, status)
VALUES ($1::uuid, 'S18_AUTH_TEST', 'S18 Auth Test Org', 'active')
ON CONFLICT (code) DO UPDATE
SET name = EXCLUDED.name,
    status = EXCLUDED.status,
    updated_at = now()`,
		testPostgresAuthSessionOrgID,
	); err != nil {
		return err
	}

	_, err := db.ExecContext(ctx, `DELETE FROM core.auth_login_failures WHERE org_id = $1::uuid`, testPostgresAuthSessionOrgID)
	return err
}
