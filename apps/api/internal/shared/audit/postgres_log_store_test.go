package audit

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"
)

func TestPostgresLogStoreRecordPersistsRuntimeRefs(t *testing.T) {
	executor := &recordingAuditExecutor{
		queryStringErr: sql.ErrNoRows,
	}
	store := newPostgresLogStoreWithExecutor(executor, PostgresLogStoreConfig{
		DefaultOrgID: "00000000-0000-4000-8000-000000000001",
	})
	log, err := NewLog(NewLogInput{
		ID:         "audit-text-1",
		OrgID:      "org-my-pham",
		ActorID:    "user-erp-admin",
		Action:     "security.login.succeeded",
		EntityType: "auth.session",
		EntityID:   "session-local-1",
		RequestID:  "req-local-1",
		BeforeData: map[string]any{"status": "anonymous"},
		AfterData:  map[string]any{"status": "authenticated"},
		Metadata:   map[string]any{"source": "login"},
		CreatedAt:  time.Date(2026, 5, 1, 4, 0, 0, 0, time.UTC),
	})
	if err != nil {
		t.Fatalf("new log: %v", err)
	}

	if err := store.Record(context.Background(), log); err != nil {
		t.Fatalf("record: %v", err)
	}

	if len(executor.execCalls) != 1 {
		t.Fatalf("exec calls = %d, want 1", len(executor.execCalls))
	}
	call := executor.execCalls[0]
	assertAuditQueryContains(t, call.query, "INSERT INTO audit.audit_logs")
	assertAuditQueryContains(t, call.query, "log_ref")
	assertAuditQueryContains(t, call.query, "actor_ref")
	if call.args[0] != nil {
		t.Fatalf("id uuid arg = %v, want nil for text audit id", call.args[0])
	}
	if call.args[1] != "00000000-0000-4000-8000-000000000001" {
		t.Fatalf("org uuid arg = %v, want local default org", call.args[1])
	}
	if call.args[2] != nil || call.args[5] != nil {
		t.Fatalf("actor/entity uuid args = %v/%v, want nil for text refs", call.args[2], call.args[5])
	}
	if call.args[11] != "audit-text-1" || call.args[12] != "org-my-pham" || call.args[13] != "user-erp-admin" || call.args[14] != "session-local-1" {
		t.Fatalf("runtime refs args = %v, want text refs", call.args[11:15])
	}
	if !strings.Contains(fmt.Sprint(call.args[7]), "anonymous") || !strings.Contains(fmt.Sprint(call.args[9]), "login") {
		t.Fatalf("json args = before %v metadata %v, want encoded maps", call.args[7], call.args[9])
	}
}

func TestPostgresLogStoreRecordResolvesOrganizationCode(t *testing.T) {
	executor := &recordingAuditExecutor{
		queryStringValue: "00000000-0000-4000-8000-000000000099",
	}
	store := newPostgresLogStoreWithExecutor(executor, PostgresLogStoreConfig{})

	if err := store.Record(context.Background(), mustAuditLog(t, NewLogInput{
		OrgID:      "ERP_LOCAL",
		ActorID:    "00000000-0000-4000-8000-000000000101",
		Action:     "security.login.succeeded",
		EntityType: "auth.session",
		EntityID:   "00000000-0000-4000-8000-000000000201",
	})); err != nil {
		t.Fatalf("record: %v", err)
	}

	if len(executor.queryStringCalls) != 1 {
		t.Fatalf("org lookup calls = %d, want 1", len(executor.queryStringCalls))
	}
	if executor.execCalls[0].args[1] != "00000000-0000-4000-8000-000000000099" {
		t.Fatalf("org uuid arg = %v, want resolved org id", executor.execCalls[0].args[1])
	}
	if executor.execCalls[0].args[2] != "00000000-0000-4000-8000-000000000101" {
		t.Fatalf("actor uuid arg = %v, want uuid actor id", executor.execCalls[0].args[2])
	}
	if executor.execCalls[0].args[5] != "00000000-0000-4000-8000-000000000201" {
		t.Fatalf("entity uuid arg = %v, want uuid entity id", executor.execCalls[0].args[5])
	}
}

func TestPostgresLogStoreListFiltersAndMapsRefs(t *testing.T) {
	createdAt := time.Date(2026, 5, 1, 5, 0, 0, 0, time.UTC)
	executor := &recordingAuditExecutor{
		rows: &fakeAuditRows{
			values: [][]any{
				{
					"audit-login-1",
					"org-my-pham",
					"user-erp-admin",
					"security.login.succeeded",
					"auth.session",
					"session-local-1",
					"req-local-1",
					`{"status":"anonymous"}`,
					`{"status":"authenticated"}`,
					`{"source":"login"}`,
					createdAt,
				},
			},
		},
	}
	store := newPostgresLogStoreWithExecutor(executor, PostgresLogStoreConfig{})

	logs, err := store.List(context.Background(), Query{
		ActorID:    "USER-ERP-ADMIN",
		EntityType: "AUTH.SESSION",
		EntityID:   "SESSION-LOCAL-1",
		Limit:      250,
	})
	if err != nil {
		t.Fatalf("list: %v", err)
	}

	if len(logs) != 1 {
		t.Fatalf("logs = %d, want 1", len(logs))
	}
	if logs[0].ID != "audit-login-1" || logs[0].ActorID != "user-erp-admin" || logs[0].EntityID != "session-local-1" {
		t.Fatalf("mapped log = %+v, want runtime refs", logs[0])
	}
	if logs[0].BeforeData["status"] != "anonymous" || logs[0].AfterData["status"] != "authenticated" || logs[0].Metadata["source"] != "login" {
		t.Fatalf("mapped json = before %v after %v metadata %v", logs[0].BeforeData, logs[0].AfterData, logs[0].Metadata)
	}
	if !logs[0].CreatedAt.Equal(createdAt) {
		t.Fatalf("created_at = %s, want %s", logs[0].CreatedAt, createdAt)
	}
	if len(executor.queryCalls) != 1 {
		t.Fatalf("query calls = %d, want 1", len(executor.queryCalls))
	}
	queryCall := executor.queryCalls[0]
	assertAuditQueryContains(t, queryCall.query, "lower(COALESCE(actor_ref, actor_id::text, '')) = lower($1)")
	assertAuditQueryContains(t, queryCall.query, "lower(entity_type) = lower($2)")
	assertAuditQueryContains(t, queryCall.query, "lower(COALESCE(entity_ref, entity_id::text, '')) = lower($3)")
	if queryCall.args[3] != 100 {
		t.Fatalf("limit arg = %v, want normalized max 100", queryCall.args[3])
	}
}

func TestPostgresLogStoreRequiresResolvableOrg(t *testing.T) {
	executor := &recordingAuditExecutor{
		queryStringErr: sql.ErrNoRows,
	}
	store := newPostgresLogStoreWithExecutor(executor, PostgresLogStoreConfig{})

	err := store.Record(context.Background(), mustAuditLog(t, NewLogInput{
		OrgID:      "org-missing",
		ActorID:    "user-erp-admin",
		Action:     "security.login.succeeded",
		EntityType: "auth.session",
		EntityID:   "session-local-1",
	}))
	if err == nil {
		t.Fatal("record error = nil, want unresolved org error")
	}
	if len(executor.execCalls) != 0 {
		t.Fatalf("exec calls = %d, want no insert after unresolved org", len(executor.execCalls))
	}
}

type auditExecCall struct {
	query string
	args  []any
}

type recordingAuditExecutor struct {
	execCalls        []auditExecCall
	queryCalls       []auditExecCall
	queryStringCalls []auditExecCall
	queryStringValue string
	queryStringErr   error
	rows             *fakeAuditRows
}

func (e *recordingAuditExecutor) Exec(_ context.Context, query string, args ...any) error {
	e.execCalls = append(e.execCalls, auditExecCall{query: query, args: append([]any(nil), args...)})
	return nil
}

func (e *recordingAuditExecutor) Query(_ context.Context, query string, args ...any) (postgresAuditRows, error) {
	e.queryCalls = append(e.queryCalls, auditExecCall{query: query, args: append([]any(nil), args...)})
	if e.rows == nil {
		return nil, errors.New("rows are required")
	}
	return e.rows, nil
}

func (e *recordingAuditExecutor) QueryString(_ context.Context, query string, args ...any) (string, error) {
	e.queryStringCalls = append(e.queryStringCalls, auditExecCall{query: query, args: append([]any(nil), args...)})
	if e.queryStringErr != nil {
		return "", e.queryStringErr
	}
	return e.queryStringValue, nil
}

type fakeAuditRows struct {
	values [][]any
	index  int
	err    error
}

func (r *fakeAuditRows) Next() bool {
	if r.index >= len(r.values) {
		return false
	}
	r.index++
	return true
}

func (r *fakeAuditRows) Scan(dest ...any) error {
	if r.index == 0 || r.index > len(r.values) {
		return errors.New("scan called without current row")
	}
	row := r.values[r.index-1]
	if len(dest) != len(row) {
		return fmt.Errorf("scan destinations = %d, values = %d", len(dest), len(row))
	}
	for index, value := range row {
		switch target := dest[index].(type) {
		case *string:
			*target = fmt.Sprint(value)
		case *sql.NullString:
			if value == nil {
				target.Valid = false
				target.String = ""
				continue
			}
			target.Valid = true
			target.String = fmt.Sprint(value)
		case *time.Time:
			timestamp, ok := value.(time.Time)
			if !ok {
				return fmt.Errorf("value %d is %T, want time.Time", index, value)
			}
			*target = timestamp
		default:
			return fmt.Errorf("unsupported scan destination %T", target)
		}
	}

	return nil
}

func (r *fakeAuditRows) Close() error {
	return nil
}

func (r *fakeAuditRows) Err() error {
	return r.err
}

func mustAuditLog(t *testing.T, input NewLogInput) Log {
	t.Helper()
	log, err := NewLog(input)
	if err != nil {
		t.Fatalf("new audit log: %v", err)
	}

	return log
}

func assertAuditQueryContains(t *testing.T, query string, want string) {
	t.Helper()
	if !strings.Contains(query, want) {
		t.Fatalf("query %q does not contain %q", query, want)
	}
}
