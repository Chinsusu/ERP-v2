package auth

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"
)

type PostgresLoginFailureStoreConfig struct {
	DefaultOrgID string
}

type PostgresLoginFailureStore struct {
	db           *sql.DB
	defaultOrgID string
}

func NewPostgresLoginFailureStore(
	db *sql.DB,
	cfg PostgresLoginFailureStoreConfig,
) *PostgresLoginFailureStore {
	return &PostgresLoginFailureStore{
		db:           db,
		defaultOrgID: strings.TrimSpace(cfg.DefaultOrgID),
	}
}

func (s *PostgresLoginFailureStore) LockedUntil(email string, now time.Time) (time.Time, bool, error) {
	if s == nil || s.db == nil {
		return time.Time{}, false, errors.New("auth postgres login failure store database is required")
	}

	email = normalizeEmail(email)
	if email == "" {
		return time.Time{}, false, nil
	}

	orgID, err := s.resolveOrgID()
	if err != nil {
		return time.Time{}, false, err
	}

	var lockedUntil sql.NullTime
	if err := s.db.QueryRowContext(
		context.Background(),
		postgresLoginFailureLockedUntilSQL,
		orgID,
		email,
	).Scan(&lockedUntil); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return time.Time{}, false, nil
		}
		return time.Time{}, false, err
	}

	if !lockedUntil.Valid {
		return time.Time{}, false, nil
	}
	if lockedUntil.Time.After(now.UTC()) {
		return lockedUntil.Time.UTC(), true, nil
	}

	if err := s.Clear(email); err != nil {
		return time.Time{}, false, err
	}
	return time.Time{}, false, nil
}

func (s *PostgresLoginFailureStore) RecordFailure(
	email string,
	now time.Time,
	policy LockoutPolicy,
) (time.Time, error) {
	if s == nil || s.db == nil {
		return time.Time{}, errors.New("auth postgres login failure store database is required")
	}

	email = normalizeEmail(email)
	if email == "" {
		return time.Time{}, nil
	}

	orgID, err := s.resolveOrgID()
	if err != nil {
		return time.Time{}, err
	}

	ctx := context.Background()
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return time.Time{}, err
	}
	committed := false
	defer func() {
		if !committed {
			_ = tx.Rollback()
		}
	}()

	state, found, err := selectPostgresLoginFailureForUpdate(ctx, tx, orgID, email)
	if err != nil {
		return time.Time{}, err
	}

	now = now.UTC()
	if !found || state.FirstFailed.IsZero() || now.Sub(state.FirstFailed) > policy.Window {
		state = failedLoginState{FirstFailed: now}
	}

	state.Attempts++
	if state.Attempts >= policy.MaxFailedAttempts {
		state.LockedUntil = now.Add(policy.Duration)
	}

	if _, err := tx.ExecContext(
		ctx,
		postgresUpsertLoginFailureSQL,
		orgID,
		email,
		state.Attempts,
		nullableTime(state.FirstFailed),
		nullableTime(state.LockedUntil),
		now,
	); err != nil {
		return time.Time{}, fmt.Errorf("store auth login failure: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return time.Time{}, err
	}
	committed = true

	return state.LockedUntil, nil
}

func (s *PostgresLoginFailureStore) Clear(email string) error {
	if s == nil || s.db == nil {
		return errors.New("auth postgres login failure store database is required")
	}

	email = normalizeEmail(email)
	if email == "" {
		return nil
	}

	orgID, err := s.resolveOrgID()
	if err != nil {
		return err
	}

	_, err = s.db.ExecContext(context.Background(), postgresDeleteLoginFailureSQL, orgID, email)
	return err
}

func (s *PostgresLoginFailureStore) resolveOrgID() (string, error) {
	if strings.TrimSpace(s.defaultOrgID) == "" {
		return "", errors.New("auth postgres login failure default org id is required")
	}

	return strings.TrimSpace(s.defaultOrgID), nil
}

func selectPostgresLoginFailureForUpdate(
	ctx context.Context,
	tx *sql.Tx,
	orgID string,
	email string,
) (failedLoginState, bool, error) {
	var (
		state       failedLoginState
		firstFailed sql.NullTime
		lockedUntil sql.NullTime
	)
	if err := tx.QueryRowContext(
		ctx,
		postgresSelectLoginFailureForUpdateSQL,
		orgID,
		email,
	).Scan(&state.Attempts, &firstFailed, &lockedUntil); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return failedLoginState{}, false, nil
		}
		return failedLoginState{}, false, err
	}
	if firstFailed.Valid {
		state.FirstFailed = firstFailed.Time.UTC()
	}
	if lockedUntil.Valid {
		state.LockedUntil = lockedUntil.Time.UTC()
	}
	return state, true, nil
}

func nullableTime(value time.Time) any {
	if value.IsZero() {
		return nil
	}
	return value.UTC()
}

const postgresLoginFailureLockedUntilSQL = `
SELECT locked_until
FROM core.auth_login_failures
WHERE org_id = $1::uuid
  AND email_normalized = $2`

const postgresSelectLoginFailureForUpdateSQL = `
SELECT attempts, first_failed_at, locked_until
FROM core.auth_login_failures
WHERE org_id = $1::uuid
  AND email_normalized = $2
FOR UPDATE`

const postgresUpsertLoginFailureSQL = `
INSERT INTO core.auth_login_failures (
  org_id,
  email_normalized,
  attempts,
  first_failed_at,
  locked_until,
  created_at,
  updated_at
) VALUES (
  $1::uuid,
  $2,
  $3,
  $4,
  $5,
  $6,
  $6
)
ON CONFLICT (org_id, email_normalized) DO UPDATE SET
  attempts = EXCLUDED.attempts,
  first_failed_at = EXCLUDED.first_failed_at,
  locked_until = EXCLUDED.locked_until,
  updated_at = EXCLUDED.updated_at,
  version = core.auth_login_failures.version + 1`

const postgresDeleteLoginFailureSQL = `
DELETE FROM core.auth_login_failures
WHERE org_id = $1::uuid
  AND email_normalized = $2`
