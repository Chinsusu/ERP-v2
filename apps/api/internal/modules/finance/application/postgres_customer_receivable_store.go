package application

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	financedomain "github.com/Chinsusu/ERP-v2/apps/api/internal/modules/finance/domain"
	"github.com/Chinsusu/ERP-v2/apps/api/internal/shared/decimal"
)

type PostgresCustomerReceivableStoreConfig struct {
	DefaultOrgID string
}

type PostgresCustomerReceivableStore struct {
	db           *sql.DB
	defaultOrgID string
}

type postgresCustomerReceivableQueryer interface {
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
}

func NewPostgresCustomerReceivableStore(
	db *sql.DB,
	cfg PostgresCustomerReceivableStoreConfig,
) PostgresCustomerReceivableStore {
	return PostgresCustomerReceivableStore{
		db:           db,
		defaultOrgID: strings.TrimSpace(cfg.DefaultOrgID),
	}
}

const selectCustomerReceivableOrgIDSQL = `
SELECT id::text
FROM core.organizations
WHERE code = $1
LIMIT 1`

const selectCustomerReceivableHeadersBaseSQL = `
SELECT
  receivable.id::text,
  COALESCE(receivable.receivable_ref, receivable.id::text),
  receivable.org_ref,
  receivable.receivable_no,
  receivable.customer_ref,
  COALESCE(receivable.customer_code, ''),
  receivable.customer_name,
  receivable.status,
  receivable.source_document_type,
  COALESCE(receivable.source_document_ref, ''),
  COALESCE(receivable.source_document_no, ''),
  receivable.total_amount::text,
  receivable.paid_amount::text,
  receivable.outstanding_amount::text,
  receivable.currency_code,
  receivable.due_date::text,
  COALESCE(receivable.dispute_reason, ''),
  COALESCE(receivable.disputed_by_ref, ''),
  receivable.disputed_at,
  COALESCE(receivable.void_reason, ''),
  COALESCE(receivable.voided_by_ref, ''),
  receivable.voided_at,
  COALESCE(receivable.last_receipt_by_ref, ''),
  receivable.last_receipt_at,
  receivable.created_at,
  receivable.created_by_ref,
  receivable.updated_at,
  receivable.updated_by_ref,
  receivable.version
FROM finance.customer_receivables AS receivable`

const selectCustomerReceivableHeadersSQL = selectCustomerReceivableHeadersBaseSQL + `
ORDER BY receivable.created_at DESC, receivable.receivable_no DESC`

const findCustomerReceivableHeaderSQL = selectCustomerReceivableHeadersBaseSQL + `
WHERE lower(COALESCE(receivable.receivable_ref, receivable.id::text)) = lower($1)
   OR receivable.id::text = $1
   OR lower(receivable.receivable_no) = lower($1)
LIMIT 1`

const findCustomerReceivablePersistedSQL = `
SELECT id::text, org_id::text
FROM finance.customer_receivables
WHERE lower(COALESCE(receivable_ref, id::text)) = lower($1)
   OR id::text = $1
   OR lower(receivable_no) = lower($1)
LIMIT 1
FOR UPDATE`

const selectCustomerReceivableLinesSQL = `
SELECT
  COALESCE(line.line_ref, line.id::text),
  line.description,
  line.source_document_type,
  COALESCE(line.source_document_ref, ''),
  COALESCE(line.source_document_no, ''),
  line.amount::text
FROM finance.customer_receivable_lines AS line
WHERE line.customer_receivable_id = $1::uuid
ORDER BY line.created_at, COALESCE(line.line_ref, line.id::text)`

const upsertCustomerReceivableSQL = `
INSERT INTO finance.customer_receivables (
  id,
  org_id,
  org_ref,
  receivable_ref,
  receivable_no,
  customer_ref,
  customer_code,
  customer_name,
  status,
  source_document_type,
  source_document_ref,
  source_document_no,
  total_amount,
  paid_amount,
  outstanding_amount,
  currency_code,
  due_date,
  dispute_reason,
  disputed_by_ref,
  disputed_at,
  void_reason,
  voided_by_ref,
  voided_at,
  last_receipt_by_ref,
  last_receipt_at,
  created_at,
  created_by_ref,
  updated_at,
  updated_by_ref,
  version
) VALUES (
  COALESCE($1::uuid, gen_random_uuid()),
  $2::uuid,
  $3,
  $4,
  $5,
  $6,
  $7,
  $8,
  $9,
  $10,
  $11,
  $12,
  $13,
  $14,
  $15,
  $16,
  $17,
  $18,
  $19,
  $20,
  $21,
  $22,
  $23,
  $24,
  $25,
  $26,
  $27,
  $28,
  $29,
  $30
)
ON CONFLICT (org_id, lower(receivable_ref))
DO UPDATE SET
  org_ref = EXCLUDED.org_ref,
  receivable_no = EXCLUDED.receivable_no,
  customer_ref = EXCLUDED.customer_ref,
  customer_code = EXCLUDED.customer_code,
  customer_name = EXCLUDED.customer_name,
  status = EXCLUDED.status,
  source_document_type = EXCLUDED.source_document_type,
  source_document_ref = EXCLUDED.source_document_ref,
  source_document_no = EXCLUDED.source_document_no,
  total_amount = EXCLUDED.total_amount,
  paid_amount = EXCLUDED.paid_amount,
  outstanding_amount = EXCLUDED.outstanding_amount,
  currency_code = EXCLUDED.currency_code,
  due_date = EXCLUDED.due_date,
  dispute_reason = EXCLUDED.dispute_reason,
  disputed_by_ref = EXCLUDED.disputed_by_ref,
  disputed_at = EXCLUDED.disputed_at,
  void_reason = EXCLUDED.void_reason,
  voided_by_ref = EXCLUDED.voided_by_ref,
  voided_at = EXCLUDED.voided_at,
  last_receipt_by_ref = EXCLUDED.last_receipt_by_ref,
  last_receipt_at = EXCLUDED.last_receipt_at,
  updated_at = EXCLUDED.updated_at,
  updated_by_ref = EXCLUDED.updated_by_ref,
  version = EXCLUDED.version
RETURNING id::text`

const upsertCustomerReceivableLineSQL = `
INSERT INTO finance.customer_receivable_lines (
  id,
  org_id,
  customer_receivable_id,
  line_ref,
  receivable_ref,
  description,
  source_document_type,
  source_document_ref,
  source_document_no,
  amount,
  created_at,
  updated_at
) VALUES (
  COALESCE($1::uuid, gen_random_uuid()),
  $2::uuid,
  $3::uuid,
  $4,
  $5,
  $6,
  $7,
  $8,
  $9,
  $10,
  $11,
  $12
)
ON CONFLICT (org_id, customer_receivable_id, lower(line_ref))
DO UPDATE SET
  receivable_ref = EXCLUDED.receivable_ref,
  description = EXCLUDED.description,
  source_document_type = EXCLUDED.source_document_type,
  source_document_ref = EXCLUDED.source_document_ref,
  source_document_no = EXCLUDED.source_document_no,
  amount = EXCLUDED.amount,
  updated_at = EXCLUDED.updated_at`

const selectCustomerReceivableLineRefsSQL = `
SELECT line_ref
FROM finance.customer_receivable_lines
WHERE customer_receivable_id = $1::uuid`

const deleteCustomerReceivableStaleLineSQL = `
DELETE FROM finance.customer_receivable_lines
WHERE customer_receivable_id = $1::uuid
  AND lower(line_ref) = lower($2)`

func (s PostgresCustomerReceivableStore) List(
	ctx context.Context,
	filter CustomerReceivableFilter,
) ([]financedomain.CustomerReceivable, error) {
	if s.db == nil {
		return nil, errors.New("database connection is required")
	}
	rows, err := s.db.QueryContext(ctx, selectCustomerReceivableHeadersSQL)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	filter = normalizeCustomerReceivableFilter(filter)
	receivables := make([]financedomain.CustomerReceivable, 0)
	for rows.Next() {
		receivable, err := scanPostgresCustomerReceivable(ctx, s.db, rows)
		if err != nil {
			return nil, err
		}
		if !matchesCustomerReceivableFilter(receivable, filter) {
			continue
		}
		receivables = append(receivables, receivable)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return receivables, nil
}

func (s PostgresCustomerReceivableStore) Get(
	ctx context.Context,
	id string,
) (financedomain.CustomerReceivable, error) {
	if s.db == nil {
		return financedomain.CustomerReceivable{}, errors.New("database connection is required")
	}
	row := s.db.QueryRowContext(ctx, findCustomerReceivableHeaderSQL, strings.TrimSpace(id))
	receivable, err := scanPostgresCustomerReceivable(ctx, s.db, row)
	if errors.Is(err, sql.ErrNoRows) {
		return financedomain.CustomerReceivable{}, ErrCustomerReceivableNotFound
	}
	if err != nil {
		return financedomain.CustomerReceivable{}, err
	}

	return receivable, nil
}

func (s PostgresCustomerReceivableStore) Save(
	ctx context.Context,
	receivable financedomain.CustomerReceivable,
) error {
	if s.db == nil {
		return errors.New("database connection is required")
	}
	if err := receivable.Validate(); err != nil {
		return err
	}
	tx, err := s.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
	if err != nil {
		return fmt.Errorf("begin customer receivable transaction: %w", err)
	}
	committed := false
	defer func() {
		if !committed {
			_ = tx.Rollback()
		}
	}()

	persistedID, orgID, err := findPostgresCustomerReceivable(ctx, tx, receivable.ID)
	if err != nil {
		return err
	}
	if orgID == "" {
		orgID, err = s.resolveOrgID(ctx, tx, receivable.OrgID)
		if err != nil {
			return err
		}
	}
	persistedID, err = upsertPostgresCustomerReceivable(ctx, tx, persistedID, orgID, receivable)
	if err != nil {
		return err
	}
	if err := upsertPostgresCustomerReceivableLines(ctx, tx, orgID, persistedID, receivable); err != nil {
		return err
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit customer receivable transaction: %w", err)
	}
	committed = true

	return nil
}

func (s PostgresCustomerReceivableStore) resolveOrgID(
	ctx context.Context,
	queryer postgresCustomerReceivableQueryer,
	orgRef string,
) (string, error) {
	orgRef = strings.TrimSpace(orgRef)
	if isPostgresCustomerReceivableUUIDText(orgRef) {
		return orgRef, nil
	}
	if orgRef != "" {
		var orgID string
		err := queryer.QueryRowContext(ctx, selectCustomerReceivableOrgIDSQL, orgRef).Scan(&orgID)
		if err == nil && isPostgresCustomerReceivableUUIDText(orgID) {
			return orgID, nil
		}
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			return "", fmt.Errorf("resolve customer receivable org %q: %w", orgRef, err)
		}
	}
	if isPostgresCustomerReceivableUUIDText(s.defaultOrgID) {
		return s.defaultOrgID, nil
	}

	return "", fmt.Errorf("customer receivable org %q cannot be resolved", orgRef)
}

func scanPostgresCustomerReceivable(
	ctx context.Context,
	queryer postgresCustomerReceivableQueryer,
	row interface{ Scan(dest ...any) error },
) (financedomain.CustomerReceivable, error) {
	var (
		persistedID       string
		status            string
		sourceType        string
		sourceRef         string
		sourceNo          string
		totalAmount       string
		paidAmount        string
		outstandingAmount string
		currencyCode      string
		dueDate           string
		disputedAt        sql.NullTime
		voidedAt          sql.NullTime
		lastReceiptAt     sql.NullTime
		receivable        financedomain.CustomerReceivable
		err               error
	)
	if err := row.Scan(
		&persistedID,
		&receivable.ID,
		&receivable.OrgID,
		&receivable.ReceivableNo,
		&receivable.CustomerID,
		&receivable.CustomerCode,
		&receivable.CustomerName,
		&status,
		&sourceType,
		&sourceRef,
		&sourceNo,
		&totalAmount,
		&paidAmount,
		&outstandingAmount,
		&currencyCode,
		&dueDate,
		&receivable.DisputeReason,
		&receivable.DisputedBy,
		&disputedAt,
		&receivable.VoidReason,
		&receivable.VoidedBy,
		&voidedAt,
		&receivable.LastReceiptBy,
		&lastReceiptAt,
		&receivable.CreatedAt,
		&receivable.CreatedBy,
		&receivable.UpdatedAt,
		&receivable.UpdatedBy,
		&receivable.Version,
	); err != nil {
		return financedomain.CustomerReceivable{}, err
	}
	receivable.Status = financedomain.NormalizeReceivableStatus(financedomain.ReceivableStatus(status))
	receivable.SourceDocument, err = financedomain.NewSourceDocumentRef(
		financedomain.SourceDocumentType(sourceType),
		sourceRef,
		sourceNo,
	)
	if err != nil {
		return financedomain.CustomerReceivable{}, err
	}
	receivable.TotalAmount, err = decimal.ParseMoneyAmount(totalAmount)
	if err != nil {
		return financedomain.CustomerReceivable{}, err
	}
	receivable.PaidAmount, err = decimal.ParseMoneyAmount(paidAmount)
	if err != nil {
		return financedomain.CustomerReceivable{}, err
	}
	receivable.OutstandingAmount, err = decimal.ParseMoneyAmount(outstandingAmount)
	if err != nil {
		return financedomain.CustomerReceivable{}, err
	}
	receivable.CurrencyCode, err = decimal.NormalizeCurrencyCode(currencyCode)
	if err != nil {
		return financedomain.CustomerReceivable{}, err
	}
	receivable.DueDate, err = time.Parse(time.DateOnly, dueDate)
	if err != nil {
		return financedomain.CustomerReceivable{}, err
	}
	if disputedAt.Valid {
		receivable.DisputedAt = disputedAt.Time.UTC()
	}
	if voidedAt.Valid {
		receivable.VoidedAt = voidedAt.Time.UTC()
	}
	if lastReceiptAt.Valid {
		receivable.LastReceiptAt = lastReceiptAt.Time.UTC()
	}
	receivable.CreatedAt = receivable.CreatedAt.UTC()
	receivable.UpdatedAt = receivable.UpdatedAt.UTC()
	receivable.Lines, err = listPostgresCustomerReceivableLines(ctx, queryer, persistedID)
	if err != nil {
		return financedomain.CustomerReceivable{}, err
	}
	if err := receivable.Validate(); err != nil {
		return financedomain.CustomerReceivable{}, err
	}

	return receivable, nil
}

func listPostgresCustomerReceivableLines(
	ctx context.Context,
	queryer postgresCustomerReceivableQueryer,
	persistedID string,
) ([]financedomain.CustomerReceivableLine, error) {
	rows, err := queryer.QueryContext(ctx, selectCustomerReceivableLinesSQL, persistedID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	lines := make([]financedomain.CustomerReceivableLine, 0)
	for rows.Next() {
		line, err := scanPostgresCustomerReceivableLine(rows)
		if err != nil {
			return nil, err
		}
		lines = append(lines, line)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return lines, nil
}

func scanPostgresCustomerReceivableLine(
	row interface{ Scan(dest ...any) error },
) (financedomain.CustomerReceivableLine, error) {
	var (
		line       financedomain.CustomerReceivableLine
		sourceType string
		sourceRef  string
		sourceNo   string
		amount     string
		err        error
	)
	if err := row.Scan(
		&line.ID,
		&line.Description,
		&sourceType,
		&sourceRef,
		&sourceNo,
		&amount,
	); err != nil {
		return financedomain.CustomerReceivableLine{}, err
	}
	line.SourceDocument, err = financedomain.NewSourceDocumentRef(
		financedomain.SourceDocumentType(sourceType),
		sourceRef,
		sourceNo,
	)
	if err != nil {
		return financedomain.CustomerReceivableLine{}, err
	}
	line.Amount, err = decimal.ParseMoneyAmount(amount)
	if err != nil {
		return financedomain.CustomerReceivableLine{}, err
	}
	if err := line.Validate(); err != nil {
		return financedomain.CustomerReceivableLine{}, err
	}

	return line, nil
}

func findPostgresCustomerReceivable(
	ctx context.Context,
	queryer postgresCustomerReceivableQueryer,
	id string,
) (string, string, error) {
	var persistedID string
	var orgID string
	err := queryer.QueryRowContext(ctx, findCustomerReceivablePersistedSQL, strings.TrimSpace(id)).
		Scan(&persistedID, &orgID)
	if errors.Is(err, sql.ErrNoRows) {
		return "", "", nil
	}
	if err != nil {
		return "", "", fmt.Errorf("find customer receivable %q: %w", id, err)
	}

	return persistedID, orgID, nil
}

func upsertPostgresCustomerReceivable(
	ctx context.Context,
	queryer postgresCustomerReceivableQueryer,
	persistedID string,
	orgID string,
	receivable financedomain.CustomerReceivable,
) (string, error) {
	var savedID string
	err := queryer.QueryRowContext(
		ctx,
		upsertCustomerReceivableSQL,
		nullablePostgresCustomerReceivableUUID(firstNonBlankPostgresCustomerReceivable(persistedID, receivable.ID)),
		orgID,
		receivable.OrgID,
		receivable.ID,
		receivable.ReceivableNo,
		receivable.CustomerID,
		nullablePostgresCustomerReceivableText(receivable.CustomerCode),
		receivable.CustomerName,
		string(receivable.Status),
		string(receivable.SourceDocument.Type),
		nullablePostgresCustomerReceivableText(receivable.SourceDocument.ID),
		nullablePostgresCustomerReceivableText(receivable.SourceDocument.No),
		receivable.TotalAmount.String(),
		receivable.PaidAmount.String(),
		receivable.OutstandingAmount.String(),
		receivable.CurrencyCode.String(),
		postgresCustomerReceivableDate(receivable.DueDate),
		nullablePostgresCustomerReceivableText(receivable.DisputeReason),
		nullablePostgresCustomerReceivableText(receivable.DisputedBy),
		nullablePostgresCustomerReceivableTime(receivable.DisputedAt),
		nullablePostgresCustomerReceivableText(receivable.VoidReason),
		nullablePostgresCustomerReceivableText(receivable.VoidedBy),
		nullablePostgresCustomerReceivableTime(receivable.VoidedAt),
		nullablePostgresCustomerReceivableText(receivable.LastReceiptBy),
		nullablePostgresCustomerReceivableTime(receivable.LastReceiptAt),
		postgresCustomerReceivableTime(receivable.CreatedAt),
		receivable.CreatedBy,
		postgresCustomerReceivableTime(receivable.UpdatedAt),
		receivable.UpdatedBy,
		receivable.Version,
	).Scan(&savedID)
	if err != nil {
		return "", fmt.Errorf("upsert customer receivable %q: %w", receivable.ID, err)
	}

	return savedID, nil
}

func upsertPostgresCustomerReceivableLines(
	ctx context.Context,
	queryer postgresCustomerReceivableQueryer,
	orgID string,
	persistedID string,
	receivable financedomain.CustomerReceivable,
) error {
	keptLineRefs := make(map[string]struct{}, len(receivable.Lines))
	for _, line := range receivable.Lines {
		if _, err := queryer.ExecContext(
			ctx,
			upsertCustomerReceivableLineSQL,
			nullablePostgresCustomerReceivableUUID(line.ID),
			orgID,
			persistedID,
			line.ID,
			receivable.ID,
			line.Description,
			string(line.SourceDocument.Type),
			nullablePostgresCustomerReceivableText(line.SourceDocument.ID),
			nullablePostgresCustomerReceivableText(line.SourceDocument.No),
			line.Amount.String(),
			postgresCustomerReceivableTime(receivable.CreatedAt),
			postgresCustomerReceivableTime(receivable.UpdatedAt),
		); err != nil {
			return fmt.Errorf("upsert customer receivable line %q: %w", line.ID, err)
		}
		keptLineRefs[strings.ToLower(strings.TrimSpace(line.ID))] = struct{}{}
	}

	rows, err := queryer.QueryContext(ctx, selectCustomerReceivableLineRefsSQL, persistedID)
	if err != nil {
		return fmt.Errorf("list customer receivable line refs: %w", err)
	}
	defer rows.Close()

	lineRefs := make([]string, 0)
	for rows.Next() {
		var lineRef string
		if err := rows.Scan(&lineRef); err != nil {
			return fmt.Errorf("scan customer receivable line ref: %w", err)
		}
		lineRefs = append(lineRefs, lineRef)
	}
	if err := rows.Err(); err != nil {
		return fmt.Errorf("scan customer receivable line refs: %w", err)
	}
	for _, lineRef := range lineRefs {
		if _, ok := keptLineRefs[strings.ToLower(strings.TrimSpace(lineRef))]; ok {
			continue
		}
		if _, err := queryer.ExecContext(ctx, deleteCustomerReceivableStaleLineSQL, persistedID, lineRef); err != nil {
			return fmt.Errorf("delete stale customer receivable line %q: %w", lineRef, err)
		}
	}

	return nil
}

func nullablePostgresCustomerReceivableText(value string) any {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil
	}

	return value
}

func nullablePostgresCustomerReceivableUUID(value string) any {
	value = strings.TrimSpace(value)
	if !isPostgresCustomerReceivableUUIDText(value) {
		return nil
	}

	return value
}

func nullablePostgresCustomerReceivableTime(value time.Time) any {
	if value.IsZero() {
		return nil
	}

	return value.UTC()
}

func postgresCustomerReceivableTime(value time.Time) time.Time {
	if value.IsZero() {
		return time.Now().UTC()
	}

	return value.UTC()
}

func postgresCustomerReceivableDate(value time.Time) time.Time {
	if value.IsZero() {
		return time.Date(1, time.January, 1, 0, 0, 0, 0, time.UTC)
	}

	return value.UTC()
}

func firstNonBlankPostgresCustomerReceivable(values ...string) string {
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value != "" {
			return value
		}
	}

	return ""
}

func isPostgresCustomerReceivableUUIDText(value string) bool {
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
			if !isPostgresCustomerReceivableHexText(char) {
				return false
			}
		}
	}

	return true
}

func isPostgresCustomerReceivableHexText(char rune) bool {
	return (char >= '0' && char <= '9') ||
		(char >= 'a' && char <= 'f') ||
		(char >= 'A' && char <= 'F')
}

var _ CustomerReceivableStore = PostgresCustomerReceivableStore{}
