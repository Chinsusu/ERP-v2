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

type PostgresSupplierPayableStoreConfig struct {
	DefaultOrgID string
}

type PostgresSupplierPayableStore struct {
	db           *sql.DB
	defaultOrgID string
}

type postgresSupplierPayableQueryer interface {
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
}

func NewPostgresSupplierPayableStore(
	db *sql.DB,
	cfg PostgresSupplierPayableStoreConfig,
) PostgresSupplierPayableStore {
	return PostgresSupplierPayableStore{
		db:           db,
		defaultOrgID: strings.TrimSpace(cfg.DefaultOrgID),
	}
}

const selectSupplierPayableOrgIDSQL = `
SELECT id::text
FROM core.organizations
WHERE code = $1
LIMIT 1`

const selectSupplierPayableHeadersBaseSQL = `
SELECT
  payable.id::text,
  COALESCE(payable.payable_ref, payable.id::text),
  payable.org_ref,
  payable.payable_no,
  payable.supplier_ref,
  COALESCE(payable.supplier_code, ''),
  payable.supplier_name,
  payable.status,
  payable.source_document_type,
  COALESCE(payable.source_document_ref, ''),
  COALESCE(payable.source_document_no, ''),
  payable.total_amount::text,
  payable.paid_amount::text,
  payable.outstanding_amount::text,
  payable.currency_code,
  payable.due_date::text,
  COALESCE(payable.payment_requested_by_ref, ''),
  payable.payment_requested_at,
  COALESCE(payable.payment_approved_by_ref, ''),
  payable.payment_approved_at,
  COALESCE(payable.payment_rejected_by_ref, ''),
  payable.payment_rejected_at,
  COALESCE(payable.payment_reject_reason, ''),
  COALESCE(payable.dispute_reason, ''),
  COALESCE(payable.disputed_by_ref, ''),
  payable.disputed_at,
  COALESCE(payable.void_reason, ''),
  COALESCE(payable.voided_by_ref, ''),
  payable.voided_at,
  COALESCE(payable.last_payment_by_ref, ''),
  payable.last_payment_at,
  payable.created_at,
  payable.created_by_ref,
  payable.updated_at,
  payable.updated_by_ref,
  payable.version
FROM finance.supplier_payables AS payable`

const selectSupplierPayableHeadersSQL = selectSupplierPayableHeadersBaseSQL + `
ORDER BY payable.created_at DESC, payable.payable_no DESC`

const findSupplierPayableHeaderSQL = selectSupplierPayableHeadersBaseSQL + `
WHERE lower(COALESCE(payable.payable_ref, payable.id::text)) = lower($1)
   OR payable.id::text = $1
   OR lower(payable.payable_no) = lower($1)
LIMIT 1`

const findSupplierPayablePersistedSQL = `
SELECT id::text, org_id::text
FROM finance.supplier_payables
WHERE lower(COALESCE(payable_ref, id::text)) = lower($1)
   OR id::text = $1
   OR lower(payable_no) = lower($1)
LIMIT 1
FOR UPDATE`

const selectSupplierPayableLinesSQL = `
SELECT
  COALESCE(line.line_ref, line.id::text),
  line.description,
  line.source_document_type,
  COALESCE(line.source_document_ref, ''),
  COALESCE(line.source_document_no, ''),
  line.amount::text
FROM finance.supplier_payable_lines AS line
WHERE line.supplier_payable_id = $1::uuid
ORDER BY line.created_at, COALESCE(line.line_ref, line.id::text)`

const upsertSupplierPayableSQL = `
INSERT INTO finance.supplier_payables (
  id,
  org_id,
  org_ref,
  payable_ref,
  payable_no,
  supplier_ref,
  supplier_code,
  supplier_name,
  status,
  source_document_type,
  source_document_ref,
  source_document_no,
  total_amount,
  paid_amount,
  outstanding_amount,
  currency_code,
  due_date,
  payment_requested_by_ref,
  payment_requested_at,
  payment_approved_by_ref,
  payment_approved_at,
  payment_rejected_by_ref,
  payment_rejected_at,
  payment_reject_reason,
  dispute_reason,
  disputed_by_ref,
  disputed_at,
  void_reason,
  voided_by_ref,
  voided_at,
  last_payment_by_ref,
  last_payment_at,
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
  $30,
  $31,
  $32,
  $33,
  $34,
  $35,
  $36,
  $37
)
ON CONFLICT (org_id, lower(payable_ref))
DO UPDATE SET
  org_ref = EXCLUDED.org_ref,
  payable_no = EXCLUDED.payable_no,
  supplier_ref = EXCLUDED.supplier_ref,
  supplier_code = EXCLUDED.supplier_code,
  supplier_name = EXCLUDED.supplier_name,
  status = EXCLUDED.status,
  source_document_type = EXCLUDED.source_document_type,
  source_document_ref = EXCLUDED.source_document_ref,
  source_document_no = EXCLUDED.source_document_no,
  total_amount = EXCLUDED.total_amount,
  paid_amount = EXCLUDED.paid_amount,
  outstanding_amount = EXCLUDED.outstanding_amount,
  currency_code = EXCLUDED.currency_code,
  due_date = EXCLUDED.due_date,
  payment_requested_by_ref = EXCLUDED.payment_requested_by_ref,
  payment_requested_at = EXCLUDED.payment_requested_at,
  payment_approved_by_ref = EXCLUDED.payment_approved_by_ref,
  payment_approved_at = EXCLUDED.payment_approved_at,
  payment_rejected_by_ref = EXCLUDED.payment_rejected_by_ref,
  payment_rejected_at = EXCLUDED.payment_rejected_at,
  payment_reject_reason = EXCLUDED.payment_reject_reason,
  dispute_reason = EXCLUDED.dispute_reason,
  disputed_by_ref = EXCLUDED.disputed_by_ref,
  disputed_at = EXCLUDED.disputed_at,
  void_reason = EXCLUDED.void_reason,
  voided_by_ref = EXCLUDED.voided_by_ref,
  voided_at = EXCLUDED.voided_at,
  last_payment_by_ref = EXCLUDED.last_payment_by_ref,
  last_payment_at = EXCLUDED.last_payment_at,
  updated_at = EXCLUDED.updated_at,
  updated_by_ref = EXCLUDED.updated_by_ref,
  version = EXCLUDED.version
RETURNING id::text`

const upsertSupplierPayableLineSQL = `
INSERT INTO finance.supplier_payable_lines (
  id,
  org_id,
  supplier_payable_id,
  line_ref,
  payable_ref,
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
ON CONFLICT (org_id, supplier_payable_id, lower(line_ref))
DO UPDATE SET
  payable_ref = EXCLUDED.payable_ref,
  description = EXCLUDED.description,
  source_document_type = EXCLUDED.source_document_type,
  source_document_ref = EXCLUDED.source_document_ref,
  source_document_no = EXCLUDED.source_document_no,
  amount = EXCLUDED.amount,
  updated_at = EXCLUDED.updated_at`

const selectSupplierPayableLineRefsSQL = `
SELECT line_ref
FROM finance.supplier_payable_lines
WHERE supplier_payable_id = $1::uuid`

const deleteSupplierPayableStaleLineSQL = `
DELETE FROM finance.supplier_payable_lines
WHERE supplier_payable_id = $1::uuid
  AND lower(line_ref) = lower($2)`

func (s PostgresSupplierPayableStore) List(
	ctx context.Context,
	filter SupplierPayableFilter,
) ([]financedomain.SupplierPayable, error) {
	if s.db == nil {
		return nil, errors.New("database connection is required")
	}
	rows, err := s.db.QueryContext(ctx, selectSupplierPayableHeadersSQL)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	filter = normalizeSupplierPayableFilter(filter)
	payables := make([]financedomain.SupplierPayable, 0)
	for rows.Next() {
		payable, err := scanPostgresSupplierPayable(ctx, s.db, rows)
		if err != nil {
			return nil, err
		}
		if !matchesSupplierPayableFilter(payable, filter) {
			continue
		}
		payables = append(payables, payable)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return payables, nil
}

func (s PostgresSupplierPayableStore) Get(
	ctx context.Context,
	id string,
) (financedomain.SupplierPayable, error) {
	if s.db == nil {
		return financedomain.SupplierPayable{}, errors.New("database connection is required")
	}
	row := s.db.QueryRowContext(ctx, findSupplierPayableHeaderSQL, strings.TrimSpace(id))
	payable, err := scanPostgresSupplierPayable(ctx, s.db, row)
	if errors.Is(err, sql.ErrNoRows) {
		return financedomain.SupplierPayable{}, ErrSupplierPayableNotFound
	}
	if err != nil {
		return financedomain.SupplierPayable{}, err
	}

	return payable, nil
}

func (s PostgresSupplierPayableStore) Save(
	ctx context.Context,
	payable financedomain.SupplierPayable,
) error {
	if s.db == nil {
		return errors.New("database connection is required")
	}
	if err := payable.Validate(); err != nil {
		return err
	}
	tx, err := s.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
	if err != nil {
		return fmt.Errorf("begin supplier payable transaction: %w", err)
	}
	committed := false
	defer func() {
		if !committed {
			_ = tx.Rollback()
		}
	}()

	persistedID, orgID, err := findPostgresSupplierPayable(ctx, tx, payable.ID)
	if err != nil {
		return err
	}
	if orgID == "" {
		orgID, err = s.resolveOrgID(ctx, tx, payable.OrgID)
		if err != nil {
			return err
		}
	}
	persistedID, err = upsertPostgresSupplierPayable(ctx, tx, persistedID, orgID, payable)
	if err != nil {
		return err
	}
	if err := upsertPostgresSupplierPayableLines(ctx, tx, orgID, persistedID, payable); err != nil {
		return err
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit supplier payable transaction: %w", err)
	}
	committed = true

	return nil
}

func (s PostgresSupplierPayableStore) resolveOrgID(
	ctx context.Context,
	queryer postgresSupplierPayableQueryer,
	orgRef string,
) (string, error) {
	orgRef = strings.TrimSpace(orgRef)
	if isPostgresSupplierPayableUUIDText(orgRef) {
		return orgRef, nil
	}
	if orgRef != "" {
		var orgID string
		err := queryer.QueryRowContext(ctx, selectSupplierPayableOrgIDSQL, orgRef).Scan(&orgID)
		if err == nil && isPostgresSupplierPayableUUIDText(orgID) {
			return orgID, nil
		}
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			return "", fmt.Errorf("resolve supplier payable org %q: %w", orgRef, err)
		}
	}
	if isPostgresSupplierPayableUUIDText(s.defaultOrgID) {
		return s.defaultOrgID, nil
	}

	return "", fmt.Errorf("supplier payable org %q cannot be resolved", orgRef)
}

func scanPostgresSupplierPayable(
	ctx context.Context,
	queryer postgresSupplierPayableQueryer,
	row interface{ Scan(dest ...any) error },
) (financedomain.SupplierPayable, error) {
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
		requestedAt       sql.NullTime
		approvedAt        sql.NullTime
		rejectedAt        sql.NullTime
		disputedAt        sql.NullTime
		voidedAt          sql.NullTime
		lastPaymentAt     sql.NullTime
		payable           financedomain.SupplierPayable
		err               error
	)
	if err := row.Scan(
		&persistedID,
		&payable.ID,
		&payable.OrgID,
		&payable.PayableNo,
		&payable.SupplierID,
		&payable.SupplierCode,
		&payable.SupplierName,
		&status,
		&sourceType,
		&sourceRef,
		&sourceNo,
		&totalAmount,
		&paidAmount,
		&outstandingAmount,
		&currencyCode,
		&dueDate,
		&payable.PaymentRequestedBy,
		&requestedAt,
		&payable.PaymentApprovedBy,
		&approvedAt,
		&payable.PaymentRejectedBy,
		&rejectedAt,
		&payable.PaymentRejectReason,
		&payable.DisputeReason,
		&payable.DisputedBy,
		&disputedAt,
		&payable.VoidReason,
		&payable.VoidedBy,
		&voidedAt,
		&payable.LastPaymentBy,
		&lastPaymentAt,
		&payable.CreatedAt,
		&payable.CreatedBy,
		&payable.UpdatedAt,
		&payable.UpdatedBy,
		&payable.Version,
	); err != nil {
		return financedomain.SupplierPayable{}, err
	}
	payable.Status = financedomain.NormalizePayableStatus(financedomain.PayableStatus(status))
	payable.SourceDocument, err = financedomain.NewSourceDocumentRef(
		financedomain.SourceDocumentType(sourceType),
		sourceRef,
		sourceNo,
	)
	if err != nil {
		return financedomain.SupplierPayable{}, err
	}
	payable.TotalAmount, err = decimal.ParseMoneyAmount(totalAmount)
	if err != nil {
		return financedomain.SupplierPayable{}, err
	}
	payable.PaidAmount, err = decimal.ParseMoneyAmount(paidAmount)
	if err != nil {
		return financedomain.SupplierPayable{}, err
	}
	payable.OutstandingAmount, err = decimal.ParseMoneyAmount(outstandingAmount)
	if err != nil {
		return financedomain.SupplierPayable{}, err
	}
	payable.CurrencyCode, err = decimal.NormalizeCurrencyCode(currencyCode)
	if err != nil {
		return financedomain.SupplierPayable{}, err
	}
	payable.DueDate, err = time.Parse(time.DateOnly, dueDate)
	if err != nil {
		return financedomain.SupplierPayable{}, err
	}
	if requestedAt.Valid {
		payable.PaymentRequestedAt = requestedAt.Time.UTC()
	}
	if approvedAt.Valid {
		payable.PaymentApprovedAt = approvedAt.Time.UTC()
	}
	if rejectedAt.Valid {
		payable.PaymentRejectedAt = rejectedAt.Time.UTC()
	}
	if disputedAt.Valid {
		payable.DisputedAt = disputedAt.Time.UTC()
	}
	if voidedAt.Valid {
		payable.VoidedAt = voidedAt.Time.UTC()
	}
	if lastPaymentAt.Valid {
		payable.LastPaymentAt = lastPaymentAt.Time.UTC()
	}
	payable.CreatedAt = payable.CreatedAt.UTC()
	payable.UpdatedAt = payable.UpdatedAt.UTC()
	payable.Lines, err = listPostgresSupplierPayableLines(ctx, queryer, persistedID)
	if err != nil {
		return financedomain.SupplierPayable{}, err
	}
	if err := payable.Validate(); err != nil {
		return financedomain.SupplierPayable{}, err
	}

	return payable, nil
}

func listPostgresSupplierPayableLines(
	ctx context.Context,
	queryer postgresSupplierPayableQueryer,
	persistedID string,
) ([]financedomain.SupplierPayableLine, error) {
	rows, err := queryer.QueryContext(ctx, selectSupplierPayableLinesSQL, persistedID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	lines := make([]financedomain.SupplierPayableLine, 0)
	for rows.Next() {
		line, err := scanPostgresSupplierPayableLine(rows)
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

func scanPostgresSupplierPayableLine(
	row interface{ Scan(dest ...any) error },
) (financedomain.SupplierPayableLine, error) {
	var (
		line       financedomain.SupplierPayableLine
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
		return financedomain.SupplierPayableLine{}, err
	}
	line.SourceDocument, err = financedomain.NewSourceDocumentRef(
		financedomain.SourceDocumentType(sourceType),
		sourceRef,
		sourceNo,
	)
	if err != nil {
		return financedomain.SupplierPayableLine{}, err
	}
	line.Amount, err = decimal.ParseMoneyAmount(amount)
	if err != nil {
		return financedomain.SupplierPayableLine{}, err
	}
	if err := line.Validate(); err != nil {
		return financedomain.SupplierPayableLine{}, err
	}

	return line, nil
}

func findPostgresSupplierPayable(
	ctx context.Context,
	queryer postgresSupplierPayableQueryer,
	id string,
) (string, string, error) {
	var persistedID string
	var orgID string
	err := queryer.QueryRowContext(ctx, findSupplierPayablePersistedSQL, strings.TrimSpace(id)).
		Scan(&persistedID, &orgID)
	if errors.Is(err, sql.ErrNoRows) {
		return "", "", nil
	}
	if err != nil {
		return "", "", fmt.Errorf("find supplier payable %q: %w", id, err)
	}

	return persistedID, orgID, nil
}

func upsertPostgresSupplierPayable(
	ctx context.Context,
	queryer postgresSupplierPayableQueryer,
	persistedID string,
	orgID string,
	payable financedomain.SupplierPayable,
) (string, error) {
	var savedID string
	err := queryer.QueryRowContext(
		ctx,
		upsertSupplierPayableSQL,
		nullablePostgresSupplierPayableUUID(firstNonBlankPostgresSupplierPayable(persistedID, payable.ID)),
		orgID,
		payable.OrgID,
		payable.ID,
		payable.PayableNo,
		payable.SupplierID,
		nullablePostgresSupplierPayableText(payable.SupplierCode),
		payable.SupplierName,
		string(payable.Status),
		string(payable.SourceDocument.Type),
		nullablePostgresSupplierPayableText(payable.SourceDocument.ID),
		nullablePostgresSupplierPayableText(payable.SourceDocument.No),
		payable.TotalAmount.String(),
		payable.PaidAmount.String(),
		payable.OutstandingAmount.String(),
		payable.CurrencyCode.String(),
		postgresSupplierPayableDate(payable.DueDate),
		nullablePostgresSupplierPayableText(payable.PaymentRequestedBy),
		nullablePostgresSupplierPayableTime(payable.PaymentRequestedAt),
		nullablePostgresSupplierPayableText(payable.PaymentApprovedBy),
		nullablePostgresSupplierPayableTime(payable.PaymentApprovedAt),
		nullablePostgresSupplierPayableText(payable.PaymentRejectedBy),
		nullablePostgresSupplierPayableTime(payable.PaymentRejectedAt),
		nullablePostgresSupplierPayableText(payable.PaymentRejectReason),
		nullablePostgresSupplierPayableText(payable.DisputeReason),
		nullablePostgresSupplierPayableText(payable.DisputedBy),
		nullablePostgresSupplierPayableTime(payable.DisputedAt),
		nullablePostgresSupplierPayableText(payable.VoidReason),
		nullablePostgresSupplierPayableText(payable.VoidedBy),
		nullablePostgresSupplierPayableTime(payable.VoidedAt),
		nullablePostgresSupplierPayableText(payable.LastPaymentBy),
		nullablePostgresSupplierPayableTime(payable.LastPaymentAt),
		postgresSupplierPayableTime(payable.CreatedAt),
		payable.CreatedBy,
		postgresSupplierPayableTime(payable.UpdatedAt),
		payable.UpdatedBy,
		payable.Version,
	).Scan(&savedID)
	if err != nil {
		return "", fmt.Errorf("upsert supplier payable %q: %w", payable.ID, err)
	}

	return savedID, nil
}

func upsertPostgresSupplierPayableLines(
	ctx context.Context,
	queryer postgresSupplierPayableQueryer,
	orgID string,
	persistedID string,
	payable financedomain.SupplierPayable,
) error {
	keptLineRefs := make(map[string]struct{}, len(payable.Lines))
	for _, line := range payable.Lines {
		if _, err := queryer.ExecContext(
			ctx,
			upsertSupplierPayableLineSQL,
			nullablePostgresSupplierPayableUUID(line.ID),
			orgID,
			persistedID,
			line.ID,
			payable.ID,
			line.Description,
			string(line.SourceDocument.Type),
			nullablePostgresSupplierPayableText(line.SourceDocument.ID),
			nullablePostgresSupplierPayableText(line.SourceDocument.No),
			line.Amount.String(),
			postgresSupplierPayableTime(payable.CreatedAt),
			postgresSupplierPayableTime(payable.UpdatedAt),
		); err != nil {
			return fmt.Errorf("upsert supplier payable line %q: %w", line.ID, err)
		}
		keptLineRefs[strings.ToLower(strings.TrimSpace(line.ID))] = struct{}{}
	}

	rows, err := queryer.QueryContext(ctx, selectSupplierPayableLineRefsSQL, persistedID)
	if err != nil {
		return fmt.Errorf("list supplier payable line refs: %w", err)
	}
	defer rows.Close()

	lineRefs := make([]string, 0)
	for rows.Next() {
		var lineRef string
		if err := rows.Scan(&lineRef); err != nil {
			return fmt.Errorf("scan supplier payable line ref: %w", err)
		}
		lineRefs = append(lineRefs, lineRef)
	}
	if err := rows.Err(); err != nil {
		return fmt.Errorf("scan supplier payable line refs: %w", err)
	}
	for _, lineRef := range lineRefs {
		if _, ok := keptLineRefs[strings.ToLower(strings.TrimSpace(lineRef))]; ok {
			continue
		}
		if _, err := queryer.ExecContext(ctx, deleteSupplierPayableStaleLineSQL, persistedID, lineRef); err != nil {
			return fmt.Errorf("delete stale supplier payable line %q: %w", lineRef, err)
		}
	}

	return nil
}

func nullablePostgresSupplierPayableText(value string) any {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil
	}

	return value
}

func nullablePostgresSupplierPayableUUID(value string) any {
	value = strings.TrimSpace(value)
	if !isPostgresSupplierPayableUUIDText(value) {
		return nil
	}

	return value
}

func nullablePostgresSupplierPayableTime(value time.Time) any {
	if value.IsZero() {
		return nil
	}

	return value.UTC()
}

func postgresSupplierPayableTime(value time.Time) time.Time {
	if value.IsZero() {
		return time.Now().UTC()
	}

	return value.UTC()
}

func postgresSupplierPayableDate(value time.Time) time.Time {
	if value.IsZero() {
		return time.Date(1, time.January, 1, 0, 0, 0, 0, time.UTC)
	}

	return value.UTC()
}

func firstNonBlankPostgresSupplierPayable(values ...string) string {
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value != "" {
			return value
		}
	}

	return ""
}

func isPostgresSupplierPayableUUIDText(value string) bool {
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
			if !isPostgresSupplierPayableHexText(char) {
				return false
			}
		}
	}

	return true
}

func isPostgresSupplierPayableHexText(char rune) bool {
	return (char >= '0' && char <= '9') ||
		(char >= 'a' && char <= 'f') ||
		(char >= 'A' && char <= 'F')
}

var _ SupplierPayableStore = PostgresSupplierPayableStore{}
