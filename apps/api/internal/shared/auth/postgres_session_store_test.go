package auth

import (
	"context"
	"database/sql"
	"os"
	"testing"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
)

const testPostgresAuthSessionOrgID = "00000000-0000-4000-8000-000000180202"

func TestPostgresSessionStoreRequiresDatabase(t *testing.T) {
	store := NewPostgresSessionStore(nil, PostgresSessionStoreConfig{DefaultOrgID: testPostgresAuthSessionOrgID})
	now := time.Date(2026, 5, 2, 10, 0, 0, 0, time.UTC)

	if err := store.StoreSession(Session{}, now); err == nil {
		t.Fatal("StoreSession() error = nil, want database required error")
	}
	if _, _, err := store.FindByAccessToken("token", now); err == nil {
		t.Fatal("FindByAccessToken() error = nil, want database required error")
	}
	if _, _, err := store.RotateRefreshToken("token", now, func(session Session) Session { return session }); err == nil {
		t.Fatal("RotateRefreshToken() error = nil, want database required error")
	}
}

func TestPostgresSessionStorePersistsSessionAndRotatesRefresh(t *testing.T) {
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

	if err := seedPostgresAuthSessionFixture(ctx, db); err != nil {
		t.Fatalf("seed fixture: %v", err)
	}

	now := time.Date(2026, 5, 2, 10, 0, 0, 0, time.UTC)
	store := NewPostgresSessionStore(db, PostgresSessionStoreConfig{DefaultOrgID: testPostgresAuthSessionOrgID})
	managerA, err := NewSessionManagerWithSessionStore(testConfig, func() time.Time { return now }, store)
	if err != nil {
		t.Fatalf("NewSessionManagerWithSessionStore(managerA) error = %v", err)
	}

	session, failure, ok := managerA.Login("admin@example.local", "local-only-mock-password")
	if !ok {
		t.Fatalf("login rejected: %+v", failure)
	}

	managerB, err := NewSessionManagerWithSessionStore(testConfig, func() time.Time { return now }, store)
	if err != nil {
		t.Fatalf("NewSessionManagerWithSessionStore(managerB) error = %v", err)
	}

	principal, ok := managerB.AuthenticateAccessToken(session.AccessToken)
	if !ok || principal.Email != "admin@example.local" {
		t.Fatalf("principal = %+v, authenticated = %v", principal, ok)
	}

	next, ok := managerB.Refresh(session.RefreshToken)
	if !ok {
		t.Fatal("refresh rejected after manager restart")
	}
	if next.AccessToken == session.AccessToken || next.RefreshToken == session.RefreshToken {
		t.Fatalf("tokens were not rotated: old=%+v new=%+v", session, next)
	}
	if _, ok := managerA.AuthenticateAccessToken(session.AccessToken); ok {
		t.Fatal("old access token still authenticates after refresh")
	}
	if _, ok := managerA.Refresh(session.RefreshToken); ok {
		t.Fatal("old refresh token still refreshes after rotation")
	}

	assertPostgresAuthSessionStoresOnlyHashes(t, ctx, db, session, next)
}

func seedPostgresAuthSessionFixture(ctx context.Context, db *sql.DB) error {
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

	_, err := db.ExecContext(ctx, `DELETE FROM core.auth_sessions WHERE org_id = $1::uuid`, testPostgresAuthSessionOrgID)
	return err
}

func assertPostgresAuthSessionStoresOnlyHashes(
	t *testing.T,
	ctx context.Context,
	db *sql.DB,
	oldSession Session,
	newSession Session,
) {
	t.Helper()

	var rawTokenMatches int
	if err := db.QueryRowContext(ctx, `
SELECT count(*)
FROM core.auth_sessions
WHERE access_token_hash IN ($1, $2)
   OR refresh_token_hash IN ($3, $4)`,
		oldSession.AccessToken,
		newSession.AccessToken,
		oldSession.RefreshToken,
		newSession.RefreshToken,
	).Scan(&rawTokenMatches); err != nil {
		t.Fatalf("count raw token matches: %v", err)
	}
	if rawTokenMatches != 0 {
		t.Fatalf("raw tokens were stored in hash columns, count = %d", rawTokenMatches)
	}

	var shortHashes int
	if err := db.QueryRowContext(ctx, `
SELECT count(*)
FROM core.auth_sessions
WHERE org_id = $1::uuid
  AND (length(access_token_hash) <> 64 OR length(refresh_token_hash) <> 64)`,
		testPostgresAuthSessionOrgID,
	).Scan(&shortHashes); err != nil {
		t.Fatalf("count short hashes: %v", err)
	}
	if shortHashes != 0 {
		t.Fatalf("token hash length mismatch count = %d", shortHashes)
	}
}
