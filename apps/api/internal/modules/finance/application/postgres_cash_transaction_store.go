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

type PostgresCashTransactionStoreConfig struct {
	DefaultOrgID string
}

type PostgresCashTransactionStore struct {
	db           *sql.DB
	defaultOrgID string
}

type postgresCashTransactionQueryer interface {
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
}

func NewPostgresCashTransactionStore(
	db *sql.DB,
	cfg PostgresCashTransactionStoreConfig,
) PostgresCashTransactionStore {
	return PostgresCashTransactionStore{
		db:           db,
		defaultOrgID: strings.TrimSpace(cfg.DefaultOrgID),
	}
}

const selectCashTransactionOrgIDSQL = `
SELECT id::text
FROM core.organizations
WHERE code = $1
LIMIT 1`

const selectCashTransactionHeadersBaseSQL = `
SELECT
  transaction.id::text,
  COALESCE(transaction.transaction_ref, transaction.id::text),
  transaction.org_ref,
  transaction.transaction_no,
  transaction.direction,
  transaction.status,
  transaction.business_date::text,
  COALESCE(transaction.counterparty_ref, ''),
  transaction.counterparty_name,
  transaction.payment_method,
  COALESCE(transaction.reference_no, ''),
  transaction.total_amount::text,
  transaction.currency_code,
  COALESCE(transaction.memo, ''),
  COALESCE(transaction.posted_by_ref, ''),
  transaction.posted_at,
  COALESCE(transaction.void_reason, ''),
  COALESCE(transaction.voided_by_ref, ''),
  transaction.voided_at,
  transaction.created_at,
  transaction.created_by_ref,
  transaction.updated_at,
  transaction.updated_by_ref,
  transaction.version
FROM finance.cash_transactions AS transaction`

const selectCashTransactionHeadersSQL = selectCashTransactionHeadersBaseSQL + `
ORDER BY transaction.business_date DESC, transaction.transaction_no DESC`

const findCashTransactionHeaderSQL = selectCashTransactionHeadersBaseSQL + `
WHERE lower(COALESCE(transaction.transaction_ref, transaction.id::text)) = lower($1)
   OR transaction.id::text = $1
   OR lower(transaction.transaction_no) = lower($1)
LIMIT 1`

const findCashTransactionPersistedSQL = `
SELECT id::text, org_id::text
FROM finance.cash_transactions
WHERE lower(COALESCE(transaction_ref, id::text)) = lower($1)
   OR id::text = $1
   OR lower(transaction_no) = lower($1)
LIMIT 1
FOR UPDATE`

const selectCashTransactionAllocationsSQL = `
SELECT
  COALESCE(allocation.allocation_ref, allocation.id::text),
  allocation.target_type,
  allocation.target_ref,
  allocation.target_no,
  allocation.amount::text
FROM finance.cash_transaction_allocations AS allocation
WHERE allocation.cash_transaction_id = $1::uuid
ORDER BY allocation.created_at, COALESCE(allocation.allocation_ref, allocation.id::text)`

const upsertCashTransactionSQL = `
INSERT INTO finance.cash_transactions (
  id,
  org_id,
  org_ref,
  transaction_ref,
  transaction_no,
  direction,
  status,
  business_date,
  counterparty_ref,
  counterparty_name,
  payment_method,
  reference_no,
  total_amount,
  currency_code,
  memo,
  posted_by_ref,
  posted_at,
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
  $25
)
ON CONFLICT (org_id, lower(transaction_ref))
DO UPDATE SET
  org_ref = EXCLUDED.org_ref,
  transaction_no = EXCLUDED.transaction_no,
  direction = EXCLUDED.direction,
  status = EXCLUDED.status,
  business_date = EXCLUDED.business_date,
  counterparty_ref = EXCLUDED.counterparty_ref,
  counterparty_name = EXCLUDED.counterparty_name,
  payment_method = EXCLUDED.payment_method,
  reference_no = EXCLUDED.reference_no,
  total_amount = EXCLUDED.total_amount,
  currency_code = EXCLUDED.currency_code,
  memo = EXCLUDED.memo,
  posted_by_ref = EXCLUDED.posted_by_ref,
  posted_at = EXCLUDED.posted_at,
  void_reason = EXCLUDED.void_reason,
  voided_by_ref = EXCLUDED.voided_by_ref,
  voided_at = EXCLUDED.voided_at,
  updated_at = EXCLUDED.updated_at,
  updated_by_ref = EXCLUDED.updated_by_ref,
  version = EXCLUDED.version
RETURNING id::text`

const upsertCashTransactionAllocationSQL = `
INSERT INTO finance.cash_transaction_allocations (
  id,
  org_id,
  cash_transaction_id,
  allocation_ref,
  transaction_ref,
  target_type,
  target_ref,
  target_no,
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
  $11
)
ON CONFLICT (org_id, cash_transaction_id, lower(allocation_ref))
DO UPDATE SET
  transaction_ref = EXCLUDED.transaction_ref,
  target_type = EXCLUDED.target_type,
  target_ref = EXCLUDED.target_ref,
  target_no = EXCLUDED.target_no,
  amount = EXCLUDED.amount,
  updated_at = EXCLUDED.updated_at`

const selectCashTransactionAllocationRefsSQL = `
SELECT allocation_ref
FROM finance.cash_transaction_allocations
WHERE cash_transaction_id = $1::uuid`

const deleteCashTransactionStaleAllocationSQL = `
DELETE FROM finance.cash_transaction_allocations
WHERE cash_transaction_id = $1::uuid
  AND lower(allocation_ref) = lower($2)`

func (s PostgresCashTransactionStore) List(
	ctx context.Context,
	filter CashTransactionFilter,
) ([]financedomain.CashTransaction, error) {
	if s.db == nil {
		return nil, errors.New("database connection is required")
	}
	rows, err := s.db.QueryContext(ctx, selectCashTransactionHeadersSQL)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	filter = normalizeCashTransactionFilter(filter)
	transactions := make([]financedomain.CashTransaction, 0)
	for rows.Next() {
		transaction, err := scanPostgresCashTransaction(ctx, s.db, rows)
		if err != nil {
			return nil, err
		}
		if !matchesCashTransactionFilter(transaction, filter) {
			continue
		}
		transactions = append(transactions, transaction)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return transactions, nil
}

func (s PostgresCashTransactionStore) Get(
	ctx context.Context,
	id string,
) (financedomain.CashTransaction, error) {
	if s.db == nil {
		return financedomain.CashTransaction{}, errors.New("database connection is required")
	}
	row := s.db.QueryRowContext(ctx, findCashTransactionHeaderSQL, strings.TrimSpace(id))
	transaction, err := scanPostgresCashTransaction(ctx, s.db, row)
	if errors.Is(err, sql.ErrNoRows) {
		return financedomain.CashTransaction{}, ErrCashTransactionNotFound
	}
	if err != nil {
		return financedomain.CashTransaction{}, err
	}

	return transaction, nil
}

func (s PostgresCashTransactionStore) Save(
	ctx context.Context,
	transaction financedomain.CashTransaction,
) error {
	if s.db == nil {
		return errors.New("database connection is required")
	}
	if err := transaction.Validate(); err != nil {
		return err
	}
	tx, err := s.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
	if err != nil {
		return fmt.Errorf("begin cash transaction transaction: %w", err)
	}
	committed := false
	defer func() {
		if !committed {
			_ = tx.Rollback()
		}
	}()

	persistedID, orgID, err := findPostgresCashTransaction(ctx, tx, transaction.ID)
	if err != nil {
		return err
	}
	if orgID == "" {
		orgID, err = s.resolveOrgID(ctx, tx, transaction.OrgID)
		if err != nil {
			return err
		}
	}
	persistedID, err = upsertPostgresCashTransaction(ctx, tx, persistedID, orgID, transaction)
	if err != nil {
		return err
	}
	if err := upsertPostgresCashTransactionAllocations(ctx, tx, orgID, persistedID, transaction); err != nil {
		return err
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit cash transaction transaction: %w", err)
	}
	committed = true

	return nil
}

func (s PostgresCashTransactionStore) resolveOrgID(
	ctx context.Context,
	queryer postgresCashTransactionQueryer,
	orgRef string,
) (string, error) {
	orgRef = strings.TrimSpace(orgRef)
	if isPostgresCashTransactionUUIDText(orgRef) {
		return orgRef, nil
	}
	if orgRef != "" {
		var orgID string
		err := queryer.QueryRowContext(ctx, selectCashTransactionOrgIDSQL, orgRef).Scan(&orgID)
		if err == nil && isPostgresCashTransactionUUIDText(orgID) {
			return orgID, nil
		}
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			return "", fmt.Errorf("resolve cash transaction org %q: %w", orgRef, err)
		}
	}
	if isPostgresCashTransactionUUIDText(s.defaultOrgID) {
		return s.defaultOrgID, nil
	}

	return "", fmt.Errorf("cash transaction org %q cannot be resolved", orgRef)
}

func scanPostgresCashTransaction(
	ctx context.Context,
	queryer postgresCashTransactionQueryer,
	row interface{ Scan(dest ...any) error },
) (financedomain.CashTransaction, error) {
	var (
		persistedID  string
		direction    string
		status       string
		businessDate string
		totalAmount  string
		currencyCode string
		postedAt     sql.NullTime
		voidedAt     sql.NullTime
		transaction  financedomain.CashTransaction
		err          error
	)
	if err := row.Scan(
		&persistedID,
		&transaction.ID,
		&transaction.OrgID,
		&transaction.TransactionNo,
		&direction,
		&status,
		&businessDate,
		&transaction.CounterpartyID,
		&transaction.CounterpartyName,
		&transaction.PaymentMethod,
		&transaction.ReferenceNo,
		&totalAmount,
		&currencyCode,
		&transaction.Memo,
		&transaction.PostedBy,
		&postedAt,
		&transaction.VoidReason,
		&transaction.VoidedBy,
		&voidedAt,
		&transaction.CreatedAt,
		&transaction.CreatedBy,
		&transaction.UpdatedAt,
		&transaction.UpdatedBy,
		&transaction.Version,
	); err != nil {
		return financedomain.CashTransaction{}, err
	}
	transaction.Direction = financedomain.NormalizeCashTransactionDirection(financedomain.CashTransactionDirection(direction))
	transaction.Status = financedomain.NormalizeCashTransactionStatus(financedomain.CashTransactionStatus(status))
	transaction.BusinessDate, err = time.Parse(time.DateOnly, businessDate)
	if err != nil {
		return financedomain.CashTransaction{}, err
	}
	transaction.TotalAmount, err = decimal.ParseMoneyAmount(totalAmount)
	if err != nil {
		return financedomain.CashTransaction{}, err
	}
	transaction.CurrencyCode, err = decimal.NormalizeCurrencyCode(currencyCode)
	if err != nil {
		return financedomain.CashTransaction{}, err
	}
	if postedAt.Valid {
		transaction.PostedAt = postedAt.Time.UTC()
	}
	if voidedAt.Valid {
		transaction.VoidedAt = voidedAt.Time.UTC()
	}
	transaction.CreatedAt = transaction.CreatedAt.UTC()
	transaction.UpdatedAt = transaction.UpdatedAt.UTC()
	transaction.Allocations, err = listPostgresCashTransactionAllocations(ctx, queryer, persistedID, transaction.Direction)
	if err != nil {
		return financedomain.CashTransaction{}, err
	}
	if err := transaction.Validate(); err != nil {
		return financedomain.CashTransaction{}, err
	}

	return transaction, nil
}

func listPostgresCashTransactionAllocations(
	ctx context.Context,
	queryer postgresCashTransactionQueryer,
	persistedID string,
	direction financedomain.CashTransactionDirection,
) ([]financedomain.CashTransactionAllocation, error) {
	rows, err := queryer.QueryContext(ctx, selectCashTransactionAllocationsSQL, persistedID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	allocations := make([]financedomain.CashTransactionAllocation, 0)
	for rows.Next() {
		allocation, err := scanPostgresCashTransactionAllocation(rows, direction)
		if err != nil {
			return nil, err
		}
		allocations = append(allocations, allocation)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return allocations, nil
}

func scanPostgresCashTransactionAllocation(
	row interface{ Scan(dest ...any) error },
	direction financedomain.CashTransactionDirection,
) (financedomain.CashTransactionAllocation, error) {
	var (
		allocation financedomain.CashTransactionAllocation
		targetType string
		amount     string
		err        error
	)
	if err := row.Scan(
		&allocation.ID,
		&targetType,
		&allocation.TargetID,
		&allocation.TargetNo,
		&amount,
	); err != nil {
		return financedomain.CashTransactionAllocation{}, err
	}
	allocation.TargetType = financedomain.NormalizeCashAllocationTargetType(financedomain.CashAllocationTargetType(targetType))
	allocation.Amount, err = decimal.ParseMoneyAmount(amount)
	if err != nil {
		return financedomain.CashTransactionAllocation{}, err
	}
	if err := allocation.Validate(direction); err != nil {
		return financedomain.CashTransactionAllocation{}, err
	}

	return allocation, nil
}

func findPostgresCashTransaction(
	ctx context.Context,
	queryer postgresCashTransactionQueryer,
	id string,
) (string, string, error) {
	var persistedID string
	var orgID string
	err := queryer.QueryRowContext(ctx, findCashTransactionPersistedSQL, strings.TrimSpace(id)).
		Scan(&persistedID, &orgID)
	if errors.Is(err, sql.ErrNoRows) {
		return "", "", nil
	}
	if err != nil {
		return "", "", fmt.Errorf("find cash transaction %q: %w", id, err)
	}

	return persistedID, orgID, nil
}

func upsertPostgresCashTransaction(
	ctx context.Context,
	queryer postgresCashTransactionQueryer,
	persistedID string,
	orgID string,
	transaction financedomain.CashTransaction,
) (string, error) {
	var savedID string
	err := queryer.QueryRowContext(
		ctx,
		upsertCashTransactionSQL,
		nullablePostgresCashTransactionUUID(firstNonBlankPostgresCashTransaction(persistedID, transaction.ID)),
		orgID,
		transaction.OrgID,
		transaction.ID,
		transaction.TransactionNo,
		string(transaction.Direction),
		string(transaction.Status),
		postgresCashTransactionDate(transaction.BusinessDate),
		nullablePostgresCashTransactionText(transaction.CounterpartyID),
		transaction.CounterpartyName,
		transaction.PaymentMethod,
		nullablePostgresCashTransactionText(transaction.ReferenceNo),
		transaction.TotalAmount.String(),
		transaction.CurrencyCode.String(),
		nullablePostgresCashTransactionText(transaction.Memo),
		nullablePostgresCashTransactionText(transaction.PostedBy),
		nullablePostgresCashTransactionTime(transaction.PostedAt),
		nullablePostgresCashTransactionText(transaction.VoidReason),
		nullablePostgresCashTransactionText(transaction.VoidedBy),
		nullablePostgresCashTransactionTime(transaction.VoidedAt),
		postgresCashTransactionTime(transaction.CreatedAt),
		transaction.CreatedBy,
		postgresCashTransactionTime(transaction.UpdatedAt),
		transaction.UpdatedBy,
		transaction.Version,
	).Scan(&savedID)
	if err != nil {
		return "", fmt.Errorf("upsert cash transaction %q: %w", transaction.ID, err)
	}

	return savedID, nil
}

func upsertPostgresCashTransactionAllocations(
	ctx context.Context,
	queryer postgresCashTransactionQueryer,
	orgID string,
	persistedID string,
	transaction financedomain.CashTransaction,
) error {
	keptAllocationRefs := make(map[string]struct{}, len(transaction.Allocations))
	for _, allocation := range transaction.Allocations {
		if _, err := queryer.ExecContext(
			ctx,
			upsertCashTransactionAllocationSQL,
			nullablePostgresCashTransactionUUID(allocation.ID),
			orgID,
			persistedID,
			allocation.ID,
			transaction.ID,
			string(allocation.TargetType),
			allocation.TargetID,
			allocation.TargetNo,
			allocation.Amount.String(),
			postgresCashTransactionTime(transaction.CreatedAt),
			postgresCashTransactionTime(transaction.UpdatedAt),
		); err != nil {
			return fmt.Errorf("upsert cash transaction allocation %q: %w", allocation.ID, err)
		}
		keptAllocationRefs[strings.ToLower(strings.TrimSpace(allocation.ID))] = struct{}{}
	}

	rows, err := queryer.QueryContext(ctx, selectCashTransactionAllocationRefsSQL, persistedID)
	if err != nil {
		return fmt.Errorf("list cash transaction allocation refs: %w", err)
	}
	defer rows.Close()

	allocationRefs := make([]string, 0)
	for rows.Next() {
		var allocationRef string
		if err := rows.Scan(&allocationRef); err != nil {
			return fmt.Errorf("scan cash transaction allocation ref: %w", err)
		}
		allocationRefs = append(allocationRefs, allocationRef)
	}
	if err := rows.Err(); err != nil {
		return fmt.Errorf("scan cash transaction allocation refs: %w", err)
	}
	for _, allocationRef := range allocationRefs {
		if _, ok := keptAllocationRefs[strings.ToLower(strings.TrimSpace(allocationRef))]; ok {
			continue
		}
		if _, err := queryer.ExecContext(ctx, deleteCashTransactionStaleAllocationSQL, persistedID, allocationRef); err != nil {
			return fmt.Errorf("delete stale cash transaction allocation %q: %w", allocationRef, err)
		}
	}

	return nil
}

func nullablePostgresCashTransactionText(value string) any {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil
	}

	return value
}

func nullablePostgresCashTransactionUUID(value string) any {
	value = strings.TrimSpace(value)
	if !isPostgresCashTransactionUUIDText(value) {
		return nil
	}

	return value
}

func nullablePostgresCashTransactionTime(value time.Time) any {
	if value.IsZero() {
		return nil
	}

	return value.UTC()
}

func postgresCashTransactionTime(value time.Time) time.Time {
	if value.IsZero() {
		return time.Now().UTC()
	}

	return value.UTC()
}

func postgresCashTransactionDate(value time.Time) time.Time {
	if value.IsZero() {
		return time.Date(1, time.January, 1, 0, 0, 0, 0, time.UTC)
	}

	return value.UTC()
}

func firstNonBlankPostgresCashTransaction(values ...string) string {
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value != "" {
			return value
		}
	}

	return ""
}

func isPostgresCashTransactionUUIDText(value string) bool {
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
			if !isPostgresCashTransactionHexText(char) {
				return false
			}
		}
	}

	return true
}

func isPostgresCashTransactionHexText(char rune) bool {
	return (char >= '0' && char <= '9') ||
		(char >= 'a' && char <= 'f') ||
		(char >= 'A' && char <= 'F')
}

var _ CashTransactionStore = PostgresCashTransactionStore{}
