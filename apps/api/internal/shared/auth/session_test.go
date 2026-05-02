package auth

import (
	"testing"
	"time"
)

func TestSessionManagerLoginIssuesExpiringTokens(t *testing.T) {
	now := time.Date(2026, 4, 27, 9, 0, 0, 0, time.UTC)
	manager := NewSessionManager(testConfig, func() time.Time { return now })

	session, failure, ok := manager.Login("admin@example.local", "local-only-mock-password")
	if !ok {
		t.Fatalf("login rejected: %+v", failure)
	}
	if session.AccessToken == "" || session.RefreshToken == "" {
		t.Fatalf("session tokens are empty: %+v", session)
	}
	if !session.AccessExpiresAt.Equal(now.Add(defaultAccessTokenTTL)) {
		t.Fatalf("access expiry = %s, want %s", session.AccessExpiresAt, now.Add(defaultAccessTokenTTL))
	}
	if !session.RefreshExpiresAt.Equal(now.Add(defaultRefreshTokenTTL)) {
		t.Fatalf("refresh expiry = %s, want %s", session.RefreshExpiresAt, now.Add(defaultRefreshTokenTTL))
	}

	principal, ok := manager.AuthenticateAccessToken(session.AccessToken)
	if !ok || principal.Email != "admin@example.local" {
		t.Fatalf("principal = %+v, authenticated = %v", principal, ok)
	}
}

func TestSessionManagerRefreshRotatesTokens(t *testing.T) {
	now := time.Date(2026, 4, 27, 9, 0, 0, 0, time.UTC)
	manager := NewSessionManager(testConfig, func() time.Time { return now })
	session, _, ok := manager.Login("admin@example.local", "local-only-mock-password")
	if !ok {
		t.Fatal("login rejected")
	}

	next, ok := manager.Refresh(session.RefreshToken)
	if !ok {
		t.Fatal("refresh rejected")
	}
	if next.AccessToken == session.AccessToken || next.RefreshToken == session.RefreshToken {
		t.Fatalf("tokens were not rotated: old=%+v new=%+v", session, next)
	}
	if _, ok := manager.AuthenticateAccessToken(session.AccessToken); ok {
		t.Fatal("old access token still authenticates after refresh")
	}
}

func TestSessionManagerWithSharedSessionStoreSurvivesManagerRestart(t *testing.T) {
	now := time.Date(2026, 4, 27, 9, 0, 0, 0, time.UTC)
	store := NewInMemorySessionStore()
	managerA, err := NewSessionManagerWithSessionStore(testConfig, func() time.Time { return now }, store)
	if err != nil {
		t.Fatalf("NewSessionManagerWithSessionStore(managerA) error = %v", err)
	}
	managerB, err := NewSessionManagerWithSessionStore(testConfig, func() time.Time { return now }, store)
	if err != nil {
		t.Fatalf("NewSessionManagerWithSessionStore(managerB) error = %v", err)
	}

	session, failure, ok := managerA.Login("admin@example.local", "local-only-mock-password")
	if !ok {
		t.Fatalf("login rejected: %+v", failure)
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
		t.Fatal("old access token still authenticates after refresh through shared store")
	}
}

func TestSessionManagerWithSharedLoginFailureStoreSurvivesManagerRestart(t *testing.T) {
	now := time.Date(2026, 4, 27, 9, 0, 0, 0, time.UTC)
	failureStore := NewInMemoryLoginFailureStore()
	managerA, err := NewSessionManagerWithStores(
		testConfig,
		func() time.Time { return now },
		NewInMemorySessionStore(),
		failureStore,
	)
	if err != nil {
		t.Fatalf("NewSessionManagerWithStores(managerA) error = %v", err)
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

	for range defaultMaxFailedLogins {
		_, _, _ = managerA.Login("admin@example.local", "wrong-password!")
	}

	_, failure, ok := managerB.Login("admin@example.local", "local-only-mock-password")
	if ok {
		t.Fatal("login accepted while shared failure store is locked")
	}
	if failure.Code != LoginFailureLocked {
		t.Fatalf("failure code = %q, want locked", failure.Code)
	}
}

func TestSessionManagerLocksAfterFailedLogins(t *testing.T) {
	now := time.Date(2026, 4, 27, 9, 0, 0, 0, time.UTC)
	manager := NewSessionManager(testConfig, func() time.Time { return now })

	var failure LoginFailure
	for range defaultMaxFailedLogins {
		_, failure, _ = manager.Login("admin@example.local", "wrong-password!")
	}

	if failure.Code != LoginFailureLocked {
		t.Fatalf("failure code = %q, want %q", failure.Code, LoginFailureLocked)
	}

	_, failure, ok := manager.Login("admin@example.local", "local-only-mock-password")
	if ok {
		t.Fatal("login accepted while locked")
	}
	if failure.Code != LoginFailureLocked {
		t.Fatalf("failure code = %q, want locked", failure.Code)
	}
}

func TestValidatePasswordPolicyRejectsWeakPasswords(t *testing.T) {
	policy := PasswordPolicy{
		MinLength:              defaultMinPasswordLen,
		RequireLetter:          true,
		RequireNumberOrSymbol:  true,
		CommonPasswordsBlocked: true,
	}

	for _, password := range []string{"short-1", "onlyletterslong", "password123"} {
		if got := ValidatePasswordPolicy(password, policy); got == "" {
			t.Fatalf("password %q accepted, want rejected", password)
		}
	}
	if got := ValidatePasswordPolicy("local-only-mock-password", policy); got != "" {
		t.Fatalf("local dev password rejected: %s", got)
	}
}
