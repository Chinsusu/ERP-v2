package audit

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
)

type PostgresLogStoreConfig struct {
	DefaultOrgID string
}

type PostgresLogStore struct {
	executor     postgresLogStoreExecutor
	defaultOrgID string
}

type postgresLogStoreExecutor interface {
	Exec(ctx context.Context, query string, args ...any) error
	Query(ctx context.Context, query string, args ...any) (postgresAuditRows, error)
	QueryString(ctx context.Context, query string, args ...any) (string, error)
}

type postgresAuditRows interface {
	Next() bool
	Scan(dest ...any) error
	Close() error
	Err() error
}

func NewPostgresLogStore(db *sql.DB, cfg PostgresLogStoreConfig) PostgresLogStore {
	return PostgresLogStore{
		executor:     sqlAuditExecutor{db: db},
		defaultOrgID: strings.TrimSpace(cfg.DefaultOrgID),
	}
}

func newPostgresLogStoreWithExecutor(executor postgresLogStoreExecutor, cfg PostgresLogStoreConfig) PostgresLogStore {
	return PostgresLogStore{
		executor:     executor,
		defaultOrgID: strings.TrimSpace(cfg.DefaultOrgID),
	}
}

const insertAuditLogSQL = `
INSERT INTO audit.audit_logs (
  id,
  org_id,
  actor_id,
  action,
  entity_type,
  entity_id,
  request_id,
  before_data,
  after_data,
  metadata,
  created_at,
  log_ref,
  org_ref,
  actor_ref,
  entity_ref
) VALUES (
  COALESCE($1::uuid, gen_random_uuid()),
  $2::uuid,
  $3::uuid,
  $4,
  $5,
  $6::uuid,
  $7,
  $8::jsonb,
  $9::jsonb,
  $10::jsonb,
  $11,
  $12,
  $13,
  $14,
  $15
)`

const resolveAuditOrgIDSQL = `
SELECT id::text
FROM core.organizations
WHERE code = $1
LIMIT 1`

func (s PostgresLogStore) Record(ctx context.Context, log Log) error {
	if s.executor == nil {
		return errors.New("audit postgres executor is required")
	}

	normalizedLog, err := NewLog(NewLogInput{
		ID:         log.ID,
		OrgID:      log.OrgID,
		ActorID:    log.ActorID,
		Action:     log.Action,
		EntityType: log.EntityType,
		EntityID:   log.EntityID,
		RequestID:  log.RequestID,
		BeforeData: log.BeforeData,
		AfterData:  log.AfterData,
		Metadata:   log.Metadata,
		CreatedAt:  log.CreatedAt,
	})
	if err != nil {
		return err
	}

	orgID, err := s.resolveOrgID(ctx, normalizedLog.OrgID)
	if err != nil {
		return err
	}
	beforeData, err := jsonMap(normalizedLog.BeforeData)
	if err != nil {
		return fmt.Errorf("encode audit before_data: %w", err)
	}
	afterData, err := jsonMap(normalizedLog.AfterData)
	if err != nil {
		return fmt.Errorf("encode audit after_data: %w", err)
	}
	metadata, err := requiredJSONMap(normalizedLog.Metadata)
	if err != nil {
		return fmt.Errorf("encode audit metadata: %w", err)
	}

	return s.executor.Exec(
		ctx,
		insertAuditLogSQL,
		nullableUUIDText(normalizedLog.ID),
		orgID,
		nullableUUIDText(normalizedLog.ActorID),
		normalizedLog.Action,
		normalizedLog.EntityType,
		nullableUUIDText(normalizedLog.EntityID),
		nullableString(normalizedLog.RequestID),
		beforeData,
		afterData,
		metadata,
		normalizedLog.CreatedAt.UTC(),
		nullableString(normalizedLog.ID),
		nullableString(normalizedLog.OrgID),
		nullableString(normalizedLog.ActorID),
		nullableString(normalizedLog.EntityID),
	)
}

func (s PostgresLogStore) List(ctx context.Context, query Query) ([]Log, error) {
	if s.executor == nil {
		return nil, errors.New("audit postgres executor is required")
	}

	query = normalizeQuery(query)
	listQuery, args := buildAuditLogListQuery(query)
	rows, err := s.executor.Query(ctx, listQuery, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	logs := make([]Log, 0, query.Limit)
	for rows.Next() {
		log, err := scanAuditLog(rows)
		if err != nil {
			return nil, err
		}
		logs = append(logs, log)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return logs, nil
}

func (s PostgresLogStore) resolveOrgID(ctx context.Context, orgRef string) (string, error) {
	orgRef = strings.TrimSpace(orgRef)
	if isUUIDText(orgRef) {
		return orgRef, nil
	}
	if orgRef != "" {
		orgID, err := s.executor.QueryString(ctx, resolveAuditOrgIDSQL, orgRef)
		if err == nil && isUUIDText(orgID) {
			return orgID, nil
		}
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			return "", fmt.Errorf("resolve audit org %q: %w", orgRef, err)
		}
	}
	if isUUIDText(s.defaultOrgID) {
		return s.defaultOrgID, nil
	}

	return "", fmt.Errorf("audit org %q cannot be resolved", orgRef)
}

func buildAuditLogListQuery(query Query) (string, []any) {
	clauses := make([]string, 0, 4)
	args := make([]any, 0, 5)
	addClause := func(clause string, value string) {
		args = append(args, value)
		clauses = append(clauses, fmt.Sprintf(clause, len(args)))
	}

	if query.ActorID != "" {
		addClause("(lower(COALESCE(actor_ref, actor_id::text, '')) = lower($%d))", query.ActorID)
	}
	if query.Action != "" {
		addClause("lower(action) = lower($%d)", query.Action)
	}
	if query.EntityType != "" {
		addClause("lower(entity_type) = lower($%d)", query.EntityType)
	}
	if query.EntityID != "" {
		addClause("(lower(COALESCE(entity_ref, entity_id::text, '')) = lower($%d))", query.EntityID)
	}

	builder := strings.Builder{}
	builder.WriteString(`
SELECT
  COALESCE(log_ref, id::text),
  COALESCE(org_ref, org_id::text),
  COALESCE(actor_ref, actor_id::text, ''),
  action,
  entity_type,
  COALESCE(entity_ref, entity_id::text, ''),
  COALESCE(request_id, ''),
  before_data,
  after_data,
  COALESCE(metadata, '{}'::jsonb),
  created_at
FROM audit.audit_logs`)
	if len(clauses) > 0 {
		builder.WriteString("\nWHERE ")
		builder.WriteString(strings.Join(clauses, "\n  AND "))
	}
	args = append(args, query.Limit)
	builder.WriteString(fmt.Sprintf("\nORDER BY created_at DESC, COALESCE(log_ref, id::text) DESC\nLIMIT $%d", len(args)))

	return builder.String(), args
}

func scanAuditLog(rows postgresAuditRows) (Log, error) {
	var (
		log        Log
		beforeData sql.NullString
		afterData  sql.NullString
		metadata   sql.NullString
	)
	if err := rows.Scan(
		&log.ID,
		&log.OrgID,
		&log.ActorID,
		&log.Action,
		&log.EntityType,
		&log.EntityID,
		&log.RequestID,
		&beforeData,
		&afterData,
		&metadata,
		&log.CreatedAt,
	); err != nil {
		return Log{}, err
	}

	var err error
	log.BeforeData, err = mapFromNullableJSON(beforeData)
	if err != nil {
		return Log{}, fmt.Errorf("decode audit before_data: %w", err)
	}
	log.AfterData, err = mapFromNullableJSON(afterData)
	if err != nil {
		return Log{}, fmt.Errorf("decode audit after_data: %w", err)
	}
	log.Metadata, err = mapFromNullableJSON(metadata)
	if err != nil {
		return Log{}, fmt.Errorf("decode audit metadata: %w", err)
	}
	if log.Metadata == nil {
		log.Metadata = map[string]any{}
	}
	log.CreatedAt = log.CreatedAt.UTC()

	return log, nil
}

func jsonMap(value map[string]any) (any, error) {
	if value == nil {
		return nil, nil
	}
	return requiredJSONMap(value)
}

func requiredJSONMap(value map[string]any) (string, error) {
	if value == nil {
		value = map[string]any{}
	}
	data, err := json.Marshal(value)
	if err != nil {
		return "", err
	}

	return string(data), nil
}

func mapFromNullableJSON(value sql.NullString) (map[string]any, error) {
	if !value.Valid || strings.TrimSpace(value.String) == "" {
		return nil, nil
	}

	decoded := map[string]any{}
	if err := json.Unmarshal([]byte(value.String), &decoded); err != nil {
		return nil, err
	}

	return decoded, nil
}

func nullableString(value string) any {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil
	}

	return value
}

func nullableUUIDText(value string) any {
	value = strings.TrimSpace(value)
	if !isUUIDText(value) {
		return nil
	}

	return value
}

func isUUIDText(value string) bool {
	value = strings.TrimSpace(value)
	if len(value) != 36 {
		return false
	}
	for index, char := range value {
		switch index {
		case 8, 13, 18, 23:
			if char != '-' {
				return false
			}
		default:
			if !isHexText(char) {
				return false
			}
		}
	}

	return true
}

func isHexText(char rune) bool {
	return (char >= '0' && char <= '9') ||
		(char >= 'a' && char <= 'f') ||
		(char >= 'A' && char <= 'F')
}

type sqlAuditExecutor struct {
	db *sql.DB
}

func (e sqlAuditExecutor) Exec(ctx context.Context, query string, args ...any) error {
	if e.db == nil {
		return errors.New("database connection is required")
	}
	if _, err := e.db.ExecContext(ctx, query, args...); err != nil {
		return err
	}

	return nil
}

func (e sqlAuditExecutor) Query(ctx context.Context, query string, args ...any) (postgresAuditRows, error) {
	if e.db == nil {
		return nil, errors.New("database connection is required")
	}

	return e.db.QueryContext(ctx, query, args...)
}

func (e sqlAuditExecutor) QueryString(ctx context.Context, query string, args ...any) (string, error) {
	if e.db == nil {
		return "", errors.New("database connection is required")
	}

	var value string
	if err := e.db.QueryRowContext(ctx, query, args...).Scan(&value); err != nil {
		return "", err
	}

	return value, nil
}

var _ LogStore = PostgresLogStore{}
var _ postgresAuditRows = (*sql.Rows)(nil)
