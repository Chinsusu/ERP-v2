package auth

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"
)

type PostgresSessionStoreConfig struct {
	DefaultOrgID string
}

type PostgresSessionStore struct {
	db           *sql.DB
	defaultOrgID string
}

type postgresSessionExecutor interface {
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
}

type postgresSessionRow interface {
	Scan(dest ...any) error
}

func NewPostgresSessionStore(db *sql.DB, cfg PostgresSessionStoreConfig) *PostgresSessionStore {
	return &PostgresSessionStore{
		db:           db,
		defaultOrgID: strings.TrimSpace(cfg.DefaultOrgID),
	}
}

func (s *PostgresSessionStore) StoreSession(session Session, now time.Time) error {
	if s == nil || s.db == nil {
		return errors.New("auth postgres session store database is required")
	}

	return s.storeSession(context.Background(), s.db, session, now)
}

func (s *PostgresSessionStore) FindByAccessToken(accessToken string, now time.Time) (Session, bool, error) {
	if s == nil || s.db == nil {
		return Session{}, false, errors.New("auth postgres session store database is required")
	}

	accessToken = strings.TrimSpace(accessToken)
	if accessToken == "" {
		return Session{}, false, nil
	}

	row := s.db.QueryRowContext(
		context.Background(),
		postgresFindAccessSessionSQL,
		tokenHash(accessToken),
		now.UTC(),
	)
	session, ok, err := scanPostgresSession(row, accessToken, "")
	if err != nil || !ok {
		return Session{}, ok, err
	}

	return session, true, nil
}

func (s *PostgresSessionStore) RotateRefreshToken(
	refreshToken string,
	now time.Time,
	buildNext func(Session) Session,
) (Session, bool, error) {
	if s == nil || s.db == nil {
		return Session{}, false, errors.New("auth postgres session store database is required")
	}
	if buildNext == nil {
		return Session{}, false, errors.New("auth session rotation builder is required")
	}

	refreshToken = strings.TrimSpace(refreshToken)
	if refreshToken == "" {
		return Session{}, false, nil
	}

	ctx := context.Background()
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return Session{}, false, err
	}
	committed := false
	defer func() {
		if !committed {
			_ = tx.Rollback()
		}
	}()

	row := tx.QueryRowContext(ctx, postgresFindRefreshSessionForUpdateSQL, tokenHash(refreshToken), now.UTC())
	existing, ok, err := scanPostgresSession(row, "", refreshToken)
	if err != nil || !ok {
		return Session{}, ok, err
	}

	next := buildNext(existing)
	if _, err := tx.ExecContext(ctx, postgresRevokeRefreshSessionSQL, tokenHash(refreshToken), now.UTC()); err != nil {
		return Session{}, false, err
	}
	if err := s.storeSession(ctx, tx, next, now); err != nil {
		return Session{}, false, err
	}
	if err := tx.Commit(); err != nil {
		return Session{}, false, err
	}
	committed = true

	return next, true, nil
}

func (s *PostgresSessionStore) RevokeRefreshToken(refreshToken string, now time.Time) (bool, error) {
	if s == nil || s.db == nil {
		return false, errors.New("auth postgres session store database is required")
	}

	refreshToken = strings.TrimSpace(refreshToken)
	if refreshToken == "" {
		return false, nil
	}

	result, err := s.db.ExecContext(
		context.Background(),
		postgresRevokeRefreshSessionSQL,
		tokenHash(refreshToken),
		now.UTC(),
	)
	if err != nil {
		return false, err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return false, err
	}
	return rows > 0, nil
}

func (s *PostgresSessionStore) storeSession(
	ctx context.Context,
	executor postgresSessionExecutor,
	session Session,
	now time.Time,
) error {
	orgID, err := s.resolveOrgID()
	if err != nil {
		return err
	}
	if strings.TrimSpace(session.AccessToken) == "" || strings.TrimSpace(session.RefreshToken) == "" {
		return errors.New("auth session tokens are required")
	}
	if strings.TrimSpace(session.Principal.UserID) == "" ||
		strings.TrimSpace(session.Principal.Email) == "" ||
		strings.TrimSpace(session.Principal.Name) == "" ||
		strings.TrimSpace(string(session.Principal.Role)) == "" {
		return errors.New("auth session principal is required")
	}

	permissionsJSON, err := postgresSessionPermissionsJSON(session.Principal.Permissions)
	if err != nil {
		return err
	}

	_, err = executor.ExecContext(
		ctx,
		postgresStoreSessionSQL,
		orgID,
		"session-"+randomToken(),
		strings.TrimSpace(session.Principal.UserID),
		strings.TrimSpace(session.Principal.Email),
		strings.TrimSpace(session.Principal.Name),
		strings.TrimSpace(string(session.Principal.Role)),
		permissionsJSON,
		tokenHash(session.AccessToken),
		tokenHash(session.RefreshToken),
		session.AccessExpiresAt.UTC(),
		session.RefreshExpiresAt.UTC(),
		now.UTC(),
	)
	if err != nil {
		return fmt.Errorf("store auth session: %w", err)
	}

	return nil
}

func (s *PostgresSessionStore) resolveOrgID() (string, error) {
	if strings.TrimSpace(s.defaultOrgID) == "" {
		return "", errors.New("auth postgres session default org id is required")
	}

	return strings.TrimSpace(s.defaultOrgID), nil
}

func scanPostgresSession(row postgresSessionRow, accessToken string, refreshToken string) (Session, bool, error) {
	var (
		userRef          string
		email            string
		displayName      string
		roleCode         string
		permissionsBytes []byte
		accessExpiresAt  time.Time
		refreshExpiresAt time.Time
	)
	if err := row.Scan(
		&userRef,
		&email,
		&displayName,
		&roleCode,
		&permissionsBytes,
		&accessExpiresAt,
		&refreshExpiresAt,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Session{}, false, nil
		}
		return Session{}, false, err
	}

	permissions, err := parsePostgresSessionPermissions(permissionsBytes)
	if err != nil {
		return Session{}, false, err
	}

	return Session{
		AccessToken:      accessToken,
		RefreshToken:     refreshToken,
		AccessExpiresAt:  accessExpiresAt.UTC(),
		RefreshExpiresAt: refreshExpiresAt.UTC(),
		Principal: Principal{
			UserID:      userRef,
			Email:       email,
			Name:        displayName,
			Role:        RoleKey(roleCode),
			Permissions: permissions,
		},
	}, true, nil
}

func postgresSessionPermissionsJSON(permissions []PermissionKey) ([]byte, error) {
	values := make([]string, 0, len(permissions))
	for _, permission := range permissions {
		values = append(values, string(permission))
	}

	encoded, err := json.Marshal(values)
	if err != nil {
		return nil, fmt.Errorf("encode auth session permissions: %w", err)
	}
	return encoded, nil
}

func parsePostgresSessionPermissions(raw []byte) ([]PermissionKey, error) {
	var values []string
	if err := json.Unmarshal(raw, &values); err != nil {
		return nil, fmt.Errorf("decode auth session permissions: %w", err)
	}

	permissions := make([]PermissionKey, 0, len(values))
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value == "" {
			continue
		}
		permissions = append(permissions, PermissionKey(value))
	}
	return permissions, nil
}

func tokenHash(token string) string {
	sum := sha256.Sum256([]byte(strings.TrimSpace(token)))
	return hex.EncodeToString(sum[:])
}

const postgresStoreSessionSQL = `
INSERT INTO core.auth_sessions (
  org_id,
  session_ref,
  user_ref,
  email,
  display_name,
  role_code,
  permissions,
  access_token_hash,
  refresh_token_hash,
  access_expires_at,
  refresh_expires_at,
  revoked_at,
  rotated_at,
  last_seen_at,
  created_at,
  updated_at
) VALUES (
  $1::uuid,
  $2,
  $3,
  $4,
  $5,
  $6,
  $7::jsonb,
  $8,
  $9,
  $10,
  $11,
  NULL,
  NULL,
  NULL,
  $12,
  $12
)
ON CONFLICT (access_token_hash) DO UPDATE SET
  session_ref = EXCLUDED.session_ref,
  user_ref = EXCLUDED.user_ref,
  email = EXCLUDED.email,
  display_name = EXCLUDED.display_name,
  role_code = EXCLUDED.role_code,
  permissions = EXCLUDED.permissions,
  refresh_token_hash = EXCLUDED.refresh_token_hash,
  access_expires_at = EXCLUDED.access_expires_at,
  refresh_expires_at = EXCLUDED.refresh_expires_at,
  revoked_at = NULL,
  rotated_at = NULL,
  updated_at = EXCLUDED.updated_at,
  version = core.auth_sessions.version + 1`

const postgresFindAccessSessionSQL = `
UPDATE core.auth_sessions
SET last_seen_at = $2,
    updated_at = $2,
    version = version + 1
WHERE access_token_hash = $1
  AND revoked_at IS NULL
  AND rotated_at IS NULL
  AND access_expires_at > $2
RETURNING user_ref, email, display_name, role_code, permissions, access_expires_at, refresh_expires_at`

const postgresFindRefreshSessionForUpdateSQL = `
SELECT user_ref, email, display_name, role_code, permissions, access_expires_at, refresh_expires_at
FROM core.auth_sessions
WHERE refresh_token_hash = $1
  AND revoked_at IS NULL
  AND rotated_at IS NULL
  AND refresh_expires_at > $2
FOR UPDATE`

const postgresRevokeRefreshSessionSQL = `
UPDATE core.auth_sessions
SET revoked_at = $2,
    rotated_at = $2,
    updated_at = $2,
    version = version + 1
WHERE refresh_token_hash = $1
  AND revoked_at IS NULL
  AND rotated_at IS NULL`
