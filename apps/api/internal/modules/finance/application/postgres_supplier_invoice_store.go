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

type PostgresSupplierInvoiceStoreConfig struct {
	DefaultOrgID string
}

type PostgresSupplierInvoiceStore struct {
	db           *sql.DB
	defaultOrgID string
}

type postgresSupplierInvoiceQueryer interface {
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
}

func NewPostgresSupplierInvoiceStore(
	db *sql.DB,
	cfg PostgresSupplierInvoiceStoreConfig,
) PostgresSupplierInvoiceStore {
	return PostgresSupplierInvoiceStore{
		db:           db,
		defaultOrgID: strings.TrimSpace(cfg.DefaultOrgID),
	}
}

const selectSupplierInvoiceOrgIDSQL = `
SELECT id::text
FROM core.organizations
WHERE code = $1
LIMIT 1`

const selectSupplierInvoiceHeadersBaseSQL = `
SELECT
  invoice.id::text,
  COALESCE(invoice.invoice_ref, invoice.id::text),
  invoice.org_ref,
  invoice.invoice_no,
  invoice.supplier_ref,
  COALESCE(invoice.supplier_code, ''),
  invoice.supplier_name,
  invoice.payable_ref,
  invoice.payable_no,
  invoice.status,
  invoice.match_status,
  invoice.source_document_type,
  COALESCE(invoice.source_document_ref, ''),
  COALESCE(invoice.source_document_no, ''),
  invoice.invoice_amount::text,
  invoice.expected_amount::text,
  invoice.variance_amount::text,
  invoice.currency_code,
  invoice.invoice_date::text,
  COALESCE(invoice.void_reason, ''),
  COALESCE(invoice.voided_by_ref, ''),
  invoice.voided_at,
  invoice.created_at,
  invoice.created_by_ref,
  invoice.updated_at,
  invoice.updated_by_ref,
  invoice.version
FROM finance.supplier_invoices AS invoice`

const selectSupplierInvoiceHeadersSQL = selectSupplierInvoiceHeadersBaseSQL + `
ORDER BY invoice.created_at DESC, invoice.invoice_no DESC`

const findSupplierInvoiceHeaderSQL = selectSupplierInvoiceHeadersBaseSQL + `
WHERE lower(COALESCE(invoice.invoice_ref, invoice.id::text)) = lower($1)
   OR invoice.id::text = $1
   OR lower(invoice.invoice_no) = lower($1)
LIMIT 1`

const findSupplierInvoicePersistedSQL = `
SELECT id::text, org_id::text
FROM finance.supplier_invoices
WHERE lower(COALESCE(invoice_ref, id::text)) = lower($1)
   OR id::text = $1
   OR lower(invoice_no) = lower($1)
LIMIT 1
FOR UPDATE`

const selectSupplierInvoiceLinesSQL = `
SELECT
  COALESCE(line.line_ref, line.id::text),
  line.description,
  line.source_document_type,
  COALESCE(line.source_document_ref, ''),
  COALESCE(line.source_document_no, ''),
  line.amount::text
FROM finance.supplier_invoice_lines AS line
WHERE line.supplier_invoice_id = $1::uuid
ORDER BY line.created_at, COALESCE(line.line_ref, line.id::text)`

const upsertSupplierInvoiceSQL = `
INSERT INTO finance.supplier_invoices (
  id,
  org_id,
  org_ref,
  invoice_ref,
  invoice_no,
  supplier_ref,
  supplier_code,
  supplier_name,
  payable_ref,
  payable_no,
  status,
  match_status,
  source_document_type,
  source_document_ref,
  source_document_no,
  invoice_amount,
  expected_amount,
  variance_amount,
  currency_code,
  invoice_date,
  void_reason,
  voided_by_ref,
  voided_at,
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
  $28
)
ON CONFLICT (org_id, lower(invoice_ref))
DO UPDATE SET
  org_ref = EXCLUDED.org_ref,
  invoice_no = EXCLUDED.invoice_no,
  supplier_ref = EXCLUDED.supplier_ref,
  supplier_code = EXCLUDED.supplier_code,
  supplier_name = EXCLUDED.supplier_name,
  payable_ref = EXCLUDED.payable_ref,
  payable_no = EXCLUDED.payable_no,
  status = EXCLUDED.status,
  match_status = EXCLUDED.match_status,
  source_document_type = EXCLUDED.source_document_type,
  source_document_ref = EXCLUDED.source_document_ref,
  source_document_no = EXCLUDED.source_document_no,
  invoice_amount = EXCLUDED.invoice_amount,
  expected_amount = EXCLUDED.expected_amount,
  variance_amount = EXCLUDED.variance_amount,
  currency_code = EXCLUDED.currency_code,
  invoice_date = EXCLUDED.invoice_date,
  void_reason = EXCLUDED.void_reason,
  voided_by_ref = EXCLUDED.voided_by_ref,
  voided_at = EXCLUDED.voided_at,
  updated_at = EXCLUDED.updated_at,
  updated_by_ref = EXCLUDED.updated_by_ref,
  version = EXCLUDED.version
RETURNING id::text`

const upsertSupplierInvoiceLineSQL = `
INSERT INTO finance.supplier_invoice_lines (
  id,
  org_id,
  supplier_invoice_id,
  line_ref,
  invoice_ref,
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
ON CONFLICT (org_id, supplier_invoice_id, lower(line_ref))
DO UPDATE SET
  invoice_ref = EXCLUDED.invoice_ref,
  description = EXCLUDED.description,
  source_document_type = EXCLUDED.source_document_type,
  source_document_ref = EXCLUDED.source_document_ref,
  source_document_no = EXCLUDED.source_document_no,
  amount = EXCLUDED.amount,
  updated_at = EXCLUDED.updated_at`

const selectSupplierInvoiceLineRefsSQL = `
SELECT line_ref
FROM finance.supplier_invoice_lines
WHERE supplier_invoice_id = $1::uuid`

const deleteSupplierInvoiceStaleLineSQL = `
DELETE FROM finance.supplier_invoice_lines
WHERE supplier_invoice_id = $1::uuid
  AND lower(line_ref) = lower($2)`

func (s PostgresSupplierInvoiceStore) List(
	ctx context.Context,
	filter SupplierInvoiceFilter,
) ([]financedomain.SupplierInvoice, error) {
	if s.db == nil {
		return nil, errors.New("database connection is required")
	}
	rows, err := s.db.QueryContext(ctx, selectSupplierInvoiceHeadersSQL)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	filter = normalizeSupplierInvoiceFilter(filter)
	invoices := make([]financedomain.SupplierInvoice, 0)
	for rows.Next() {
		invoice, err := scanPostgresSupplierInvoice(ctx, s.db, rows)
		if err != nil {
			return nil, err
		}
		if !matchesSupplierInvoiceFilter(invoice, filter) {
			continue
		}
		invoices = append(invoices, invoice)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return invoices, nil
}

func (s PostgresSupplierInvoiceStore) Get(
	ctx context.Context,
	id string,
) (financedomain.SupplierInvoice, error) {
	if s.db == nil {
		return financedomain.SupplierInvoice{}, errors.New("database connection is required")
	}
	row := s.db.QueryRowContext(ctx, findSupplierInvoiceHeaderSQL, strings.TrimSpace(id))
	invoice, err := scanPostgresSupplierInvoice(ctx, s.db, row)
	if errors.Is(err, sql.ErrNoRows) {
		return financedomain.SupplierInvoice{}, ErrSupplierInvoiceNotFound
	}
	if err != nil {
		return financedomain.SupplierInvoice{}, err
	}

	return invoice, nil
}

func (s PostgresSupplierInvoiceStore) Save(
	ctx context.Context,
	invoice financedomain.SupplierInvoice,
) error {
	if s.db == nil {
		return errors.New("database connection is required")
	}
	if err := invoice.Validate(); err != nil {
		return err
	}
	tx, err := s.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
	if err != nil {
		return fmt.Errorf("begin supplier invoice transaction: %w", err)
	}
	committed := false
	defer func() {
		if !committed {
			_ = tx.Rollback()
		}
	}()

	persistedID, orgID, err := findPostgresSupplierInvoice(ctx, tx, invoice.ID)
	if err != nil {
		return err
	}
	if orgID == "" {
		orgID, err = s.resolveOrgID(ctx, tx, invoice.OrgID)
		if err != nil {
			return err
		}
	}
	persistedID, err = upsertPostgresSupplierInvoice(ctx, tx, persistedID, orgID, invoice)
	if err != nil {
		return err
	}
	if err := upsertPostgresSupplierInvoiceLines(ctx, tx, orgID, persistedID, invoice); err != nil {
		return err
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit supplier invoice transaction: %w", err)
	}
	committed = true

	return nil
}

func (s PostgresSupplierInvoiceStore) resolveOrgID(
	ctx context.Context,
	queryer postgresSupplierInvoiceQueryer,
	orgRef string,
) (string, error) {
	orgRef = strings.TrimSpace(orgRef)
	if isPostgresSupplierInvoiceUUIDText(orgRef) {
		return orgRef, nil
	}
	if orgRef != "" {
		var orgID string
		err := queryer.QueryRowContext(ctx, selectSupplierInvoiceOrgIDSQL, orgRef).Scan(&orgID)
		if err == nil && isPostgresSupplierInvoiceUUIDText(orgID) {
			return orgID, nil
		}
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			return "", fmt.Errorf("resolve supplier invoice org %q: %w", orgRef, err)
		}
	}
	if isPostgresSupplierInvoiceUUIDText(s.defaultOrgID) {
		return s.defaultOrgID, nil
	}

	return "", fmt.Errorf("supplier invoice org %q cannot be resolved", orgRef)
}

func scanPostgresSupplierInvoice(
	ctx context.Context,
	queryer postgresSupplierInvoiceQueryer,
	row interface{ Scan(dest ...any) error },
) (financedomain.SupplierInvoice, error) {
	var (
		persistedID    string
		status         string
		matchStatus    string
		sourceType     string
		sourceRef      string
		sourceNo       string
		invoiceAmount  string
		expectedAmount string
		varianceAmount string
		currencyCode   string
		invoiceDate    string
		voidedAt       sql.NullTime
		invoice        financedomain.SupplierInvoice
		err            error
	)
	if err := row.Scan(
		&persistedID,
		&invoice.ID,
		&invoice.OrgID,
		&invoice.InvoiceNo,
		&invoice.SupplierID,
		&invoice.SupplierCode,
		&invoice.SupplierName,
		&invoice.PayableID,
		&invoice.PayableNo,
		&status,
		&matchStatus,
		&sourceType,
		&sourceRef,
		&sourceNo,
		&invoiceAmount,
		&expectedAmount,
		&varianceAmount,
		&currencyCode,
		&invoiceDate,
		&invoice.VoidReason,
		&invoice.VoidedBy,
		&voidedAt,
		&invoice.CreatedAt,
		&invoice.CreatedBy,
		&invoice.UpdatedAt,
		&invoice.UpdatedBy,
		&invoice.Version,
	); err != nil {
		return financedomain.SupplierInvoice{}, err
	}
	invoice.Status = financedomain.NormalizeSupplierInvoiceStatus(financedomain.SupplierInvoiceStatus(status))
	invoice.MatchStatus = financedomain.NormalizeSupplierInvoiceMatchStatus(financedomain.SupplierInvoiceMatchStatus(matchStatus))
	invoice.SourceDocument, err = financedomain.NewSourceDocumentRef(
		financedomain.SourceDocumentType(sourceType),
		sourceRef,
		sourceNo,
	)
	if err != nil {
		return financedomain.SupplierInvoice{}, err
	}
	invoice.InvoiceAmount, err = decimal.ParseMoneyAmount(invoiceAmount)
	if err != nil {
		return financedomain.SupplierInvoice{}, err
	}
	invoice.ExpectedAmount, err = decimal.ParseMoneyAmount(expectedAmount)
	if err != nil {
		return financedomain.SupplierInvoice{}, err
	}
	invoice.VarianceAmount, err = decimal.ParseMoneyAmount(varianceAmount)
	if err != nil {
		return financedomain.SupplierInvoice{}, err
	}
	invoice.CurrencyCode, err = decimal.NormalizeCurrencyCode(currencyCode)
	if err != nil {
		return financedomain.SupplierInvoice{}, err
	}
	invoice.InvoiceDate, err = time.Parse(time.DateOnly, invoiceDate)
	if err != nil {
		return financedomain.SupplierInvoice{}, err
	}
	if voidedAt.Valid {
		invoice.VoidedAt = voidedAt.Time.UTC()
	}
	invoice.CreatedAt = invoice.CreatedAt.UTC()
	invoice.UpdatedAt = invoice.UpdatedAt.UTC()
	invoice.Lines, err = listPostgresSupplierInvoiceLines(ctx, queryer, persistedID)
	if err != nil {
		return financedomain.SupplierInvoice{}, err
	}
	if err := invoice.Validate(); err != nil {
		return financedomain.SupplierInvoice{}, err
	}

	return invoice, nil
}

func listPostgresSupplierInvoiceLines(
	ctx context.Context,
	queryer postgresSupplierInvoiceQueryer,
	persistedID string,
) ([]financedomain.SupplierInvoiceLine, error) {
	rows, err := queryer.QueryContext(ctx, selectSupplierInvoiceLinesSQL, persistedID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	lines := make([]financedomain.SupplierInvoiceLine, 0)
	for rows.Next() {
		line, err := scanPostgresSupplierInvoiceLine(rows)
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

func scanPostgresSupplierInvoiceLine(
	row interface{ Scan(dest ...any) error },
) (financedomain.SupplierInvoiceLine, error) {
	var (
		line       financedomain.SupplierInvoiceLine
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
		return financedomain.SupplierInvoiceLine{}, err
	}
	line.SourceDocument, err = financedomain.NewSourceDocumentRef(
		financedomain.SourceDocumentType(sourceType),
		sourceRef,
		sourceNo,
	)
	if err != nil {
		return financedomain.SupplierInvoiceLine{}, err
	}
	line.Amount, err = decimal.ParseMoneyAmount(amount)
	if err != nil {
		return financedomain.SupplierInvoiceLine{}, err
	}
	if err := line.Validate(); err != nil {
		return financedomain.SupplierInvoiceLine{}, err
	}

	return line, nil
}

func findPostgresSupplierInvoice(
	ctx context.Context,
	queryer postgresSupplierInvoiceQueryer,
	id string,
) (string, string, error) {
	var persistedID string
	var orgID string
	err := queryer.QueryRowContext(ctx, findSupplierInvoicePersistedSQL, strings.TrimSpace(id)).
		Scan(&persistedID, &orgID)
	if errors.Is(err, sql.ErrNoRows) {
		return "", "", nil
	}
	if err != nil {
		return "", "", fmt.Errorf("find supplier invoice %q: %w", id, err)
	}

	return persistedID, orgID, nil
}

func upsertPostgresSupplierInvoice(
	ctx context.Context,
	queryer postgresSupplierInvoiceQueryer,
	persistedID string,
	orgID string,
	invoice financedomain.SupplierInvoice,
) (string, error) {
	var savedID string
	err := queryer.QueryRowContext(
		ctx,
		upsertSupplierInvoiceSQL,
		nullablePostgresSupplierInvoiceUUID(firstNonBlankPostgresSupplierInvoice(persistedID, invoice.ID)),
		orgID,
		invoice.OrgID,
		invoice.ID,
		invoice.InvoiceNo,
		invoice.SupplierID,
		nullablePostgresSupplierInvoiceText(invoice.SupplierCode),
		invoice.SupplierName,
		invoice.PayableID,
		invoice.PayableNo,
		string(invoice.Status),
		string(invoice.MatchStatus),
		string(invoice.SourceDocument.Type),
		nullablePostgresSupplierInvoiceText(invoice.SourceDocument.ID),
		nullablePostgresSupplierInvoiceText(invoice.SourceDocument.No),
		invoice.InvoiceAmount.String(),
		invoice.ExpectedAmount.String(),
		invoice.VarianceAmount.String(),
		invoice.CurrencyCode.String(),
		postgresSupplierInvoiceDate(invoice.InvoiceDate),
		nullablePostgresSupplierInvoiceText(invoice.VoidReason),
		nullablePostgresSupplierInvoiceText(invoice.VoidedBy),
		nullablePostgresSupplierInvoiceTime(invoice.VoidedAt),
		postgresSupplierInvoiceTime(invoice.CreatedAt),
		invoice.CreatedBy,
		postgresSupplierInvoiceTime(invoice.UpdatedAt),
		invoice.UpdatedBy,
		invoice.Version,
	).Scan(&savedID)
	if err != nil {
		return "", fmt.Errorf("upsert supplier invoice %q: %w", invoice.ID, err)
	}

	return savedID, nil
}

func upsertPostgresSupplierInvoiceLines(
	ctx context.Context,
	queryer postgresSupplierInvoiceQueryer,
	orgID string,
	persistedID string,
	invoice financedomain.SupplierInvoice,
) error {
	keptLineRefs := make(map[string]struct{}, len(invoice.Lines))
	for _, line := range invoice.Lines {
		if _, err := queryer.ExecContext(
			ctx,
			upsertSupplierInvoiceLineSQL,
			nullablePostgresSupplierInvoiceUUID(line.ID),
			orgID,
			persistedID,
			line.ID,
			invoice.ID,
			line.Description,
			string(line.SourceDocument.Type),
			nullablePostgresSupplierInvoiceText(line.SourceDocument.ID),
			nullablePostgresSupplierInvoiceText(line.SourceDocument.No),
			line.Amount.String(),
			postgresSupplierInvoiceTime(invoice.CreatedAt),
			postgresSupplierInvoiceTime(invoice.UpdatedAt),
		); err != nil {
			return fmt.Errorf("upsert supplier invoice line %q: %w", line.ID, err)
		}
		keptLineRefs[strings.ToLower(strings.TrimSpace(line.ID))] = struct{}{}
	}

	rows, err := queryer.QueryContext(ctx, selectSupplierInvoiceLineRefsSQL, persistedID)
	if err != nil {
		return fmt.Errorf("list supplier invoice line refs: %w", err)
	}
	defer rows.Close()

	lineRefs := make([]string, 0)
	for rows.Next() {
		var lineRef string
		if err := rows.Scan(&lineRef); err != nil {
			return fmt.Errorf("scan supplier invoice line ref: %w", err)
		}
		lineRefs = append(lineRefs, lineRef)
	}
	if err := rows.Err(); err != nil {
		return fmt.Errorf("scan supplier invoice line refs: %w", err)
	}
	for _, lineRef := range lineRefs {
		if _, ok := keptLineRefs[strings.ToLower(strings.TrimSpace(lineRef))]; ok {
			continue
		}
		if _, err := queryer.ExecContext(ctx, deleteSupplierInvoiceStaleLineSQL, persistedID, lineRef); err != nil {
			return fmt.Errorf("delete stale supplier invoice line %q: %w", lineRef, err)
		}
	}

	return nil
}

func nullablePostgresSupplierInvoiceText(value string) any {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil
	}

	return value
}

func nullablePostgresSupplierInvoiceUUID(value string) any {
	value = strings.TrimSpace(value)
	if !isPostgresSupplierInvoiceUUIDText(value) {
		return nil
	}

	return value
}

func nullablePostgresSupplierInvoiceTime(value time.Time) any {
	if value.IsZero() {
		return nil
	}

	return value.UTC()
}

func postgresSupplierInvoiceTime(value time.Time) time.Time {
	if value.IsZero() {
		return time.Now().UTC()
	}

	return value.UTC()
}

func postgresSupplierInvoiceDate(value time.Time) time.Time {
	if value.IsZero() {
		return time.Date(1, time.January, 1, 0, 0, 0, 0, time.UTC)
	}

	return value.UTC()
}

func firstNonBlankPostgresSupplierInvoice(values ...string) string {
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value != "" {
			return value
		}
	}

	return ""
}

func isPostgresSupplierInvoiceUUIDText(value string) bool {
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
			if !isPostgresSupplierInvoiceHexText(char) {
				return false
			}
		}
	}

	return true
}

func isPostgresSupplierInvoiceHexText(char rune) bool {
	return (char >= '0' && char <= '9') ||
		(char >= 'a' && char <= 'f') ||
		(char >= 'A' && char <= 'F')
}

var _ SupplierInvoiceStore = PostgresSupplierInvoiceStore{}
