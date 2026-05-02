package application

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	productiondomain "github.com/Chinsusu/ERP-v2/apps/api/internal/modules/production/domain"
	"github.com/Chinsusu/ERP-v2/apps/api/internal/shared/decimal"
)

type PostgresSubcontractPaymentMilestoneStoreConfig struct {
	DefaultOrgID string
}

type PostgresSubcontractPaymentMilestoneStore struct {
	db           *sql.DB
	defaultOrgID string
}

type postgresSubcontractPaymentMilestoneQueryer interface {
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
}

type postgresSubcontractPaymentMilestoneHeader struct {
	PersistedID         string
	ID                  string
	OrgID               string
	MilestoneNo         string
	SubcontractOrderID  string
	SubcontractOrderNo  string
	FactoryID           string
	FactoryCode         string
	FactoryName         string
	Kind                string
	Status              string
	Amount              string
	CurrencyCode        string
	Note                string
	BlockReason         string
	ApprovedExceptionID string
	RecordedBy          string
	RecordedAt          sql.NullTime
	ReadyBy             string
	ReadyAt             sql.NullTime
	BlockedBy           string
	BlockedAt           sql.NullTime
	CreatedAt           time.Time
	CreatedBy           string
	UpdatedAt           time.Time
	UpdatedBy           string
	Version             int
}

func NewPostgresSubcontractPaymentMilestoneStore(
	db *sql.DB,
	cfg PostgresSubcontractPaymentMilestoneStoreConfig,
) PostgresSubcontractPaymentMilestoneStore {
	return PostgresSubcontractPaymentMilestoneStore{
		db:           db,
		defaultOrgID: strings.TrimSpace(cfg.DefaultOrgID),
	}
}

const selectSubcontractPaymentMilestoneOrgIDSQL = `
SELECT id::text
FROM core.organizations
WHERE code = $1
LIMIT 1`

const selectSubcontractPaymentMilestoneHeadersBaseSQL = `
SELECT
  milestone.id::text,
  milestone.milestone_ref,
  milestone.org_ref,
  milestone.milestone_no,
  milestone.subcontract_order_ref,
  milestone.subcontract_order_no,
  milestone.factory_ref,
  COALESCE(milestone.factory_code, ''),
  milestone.factory_name,
  milestone.kind,
  milestone.status,
  milestone.amount::text,
  milestone.currency_code,
  COALESCE(milestone.note, ''),
  COALESCE(milestone.block_reason, ''),
  COALESCE(milestone.approved_exception_ref, ''),
  COALESCE(milestone.recorded_by_ref, ''),
  milestone.recorded_at,
  COALESCE(milestone.ready_by_ref, ''),
  milestone.ready_at,
  COALESCE(milestone.blocked_by_ref, ''),
  milestone.blocked_at,
  milestone.created_at,
  milestone.created_by_ref,
  milestone.updated_at,
  milestone.updated_by_ref,
  milestone.version
FROM subcontract.subcontract_payment_milestones AS milestone`

const findSubcontractPaymentMilestoneHeaderSQL = selectSubcontractPaymentMilestoneHeadersBaseSQL + `
WHERE lower(milestone.milestone_ref) = lower($1)
   OR milestone.id::text = $1
   OR lower(milestone.milestone_no) = lower($1)
LIMIT 1`

const selectSubcontractPaymentMilestonesByOrderSQL = selectSubcontractPaymentMilestoneHeadersBaseSQL + `
WHERE lower(milestone.subcontract_order_ref) = lower($1)
   OR milestone.subcontract_order_id::text = $1
   OR lower(milestone.subcontract_order_no) = lower($1)
ORDER BY milestone.created_at DESC, milestone.milestone_no DESC`

const upsertSubcontractPaymentMilestoneSQL = `
INSERT INTO subcontract.subcontract_payment_milestones (
  id,
  org_id,
  org_ref,
  milestone_ref,
  milestone_no,
  subcontract_order_id,
  subcontract_order_ref,
  subcontract_order_no,
  factory_ref,
  factory_code,
  factory_name,
  kind,
  status,
  amount,
  currency_code,
  note,
  block_reason,
  approved_exception_ref,
  recorded_by_ref,
  recorded_at,
  ready_by_ref,
  ready_at,
  blocked_by_ref,
  blocked_at,
  created_at,
  created_by_ref,
  updated_at,
  updated_by_ref,
  version
) VALUES (
  COALESCE(CASE WHEN NULLIF($2::text, '') ~* '^[0-9a-f]{8}-[0-9a-f]{4}-[1-5][0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$' THEN $2::uuid END, gen_random_uuid()),
  $1::uuid,
  $3,
  $2,
  $4,
  (
    SELECT subcontract_order.id
    FROM subcontract.subcontract_orders AS subcontract_order
    WHERE subcontract_order.org_id = $1::uuid
      AND (
        subcontract_order.id::text = $5
        OR lower(subcontract_order.order_ref) = lower($5)
      )
    LIMIT 1
  ),
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
  $27
)
ON CONFLICT (org_id, milestone_ref)
DO UPDATE SET
  org_ref = EXCLUDED.org_ref,
  milestone_no = EXCLUDED.milestone_no,
  subcontract_order_id = EXCLUDED.subcontract_order_id,
  subcontract_order_ref = EXCLUDED.subcontract_order_ref,
  subcontract_order_no = EXCLUDED.subcontract_order_no,
  factory_ref = EXCLUDED.factory_ref,
  factory_code = EXCLUDED.factory_code,
  factory_name = EXCLUDED.factory_name,
  kind = EXCLUDED.kind,
  status = EXCLUDED.status,
  amount = EXCLUDED.amount,
  currency_code = EXCLUDED.currency_code,
  note = EXCLUDED.note,
  block_reason = EXCLUDED.block_reason,
  approved_exception_ref = EXCLUDED.approved_exception_ref,
  recorded_by_ref = EXCLUDED.recorded_by_ref,
  recorded_at = EXCLUDED.recorded_at,
  ready_by_ref = EXCLUDED.ready_by_ref,
  ready_at = EXCLUDED.ready_at,
  blocked_by_ref = EXCLUDED.blocked_by_ref,
  blocked_at = EXCLUDED.blocked_at,
  created_at = EXCLUDED.created_at,
  created_by_ref = EXCLUDED.created_by_ref,
  updated_at = EXCLUDED.updated_at,
  updated_by_ref = EXCLUDED.updated_by_ref,
  version = EXCLUDED.version
RETURNING id::text`

func (s PostgresSubcontractPaymentMilestoneStore) Save(
	ctx context.Context,
	milestone productiondomain.SubcontractPaymentMilestone,
) error {
	if s.db == nil {
		return errors.New("database connection is required")
	}
	if err := milestone.Validate(); err != nil {
		return err
	}
	return s.withTx(ctx, func(tx *sql.Tx) error {
		orgID, err := s.resolveOrgID(ctx, tx, milestone.OrgID)
		if err != nil {
			return err
		}
		_, err = upsertPostgresSubcontractPaymentMilestone(ctx, tx, orgID, milestone)
		return err
	})
}

func (s PostgresSubcontractPaymentMilestoneStore) Get(
	ctx context.Context,
	id string,
) (productiondomain.SubcontractPaymentMilestone, error) {
	if s.db == nil {
		return productiondomain.SubcontractPaymentMilestone{}, errors.New("database connection is required")
	}
	row := s.db.QueryRowContext(ctx, findSubcontractPaymentMilestoneHeaderSQL, strings.TrimSpace(id))
	milestone, err := scanPostgresSubcontractPaymentMilestone(row)
	if errors.Is(err, sql.ErrNoRows) {
		return productiondomain.SubcontractPaymentMilestone{}, ErrSubcontractPaymentMilestoneNotFound
	}
	if err != nil {
		return productiondomain.SubcontractPaymentMilestone{}, err
	}

	return milestone, nil
}

func (s PostgresSubcontractPaymentMilestoneStore) ListBySubcontractOrder(
	ctx context.Context,
	subcontractOrderID string,
) ([]productiondomain.SubcontractPaymentMilestone, error) {
	if s.db == nil {
		return nil, errors.New("database connection is required")
	}
	rows, err := s.db.QueryContext(ctx, selectSubcontractPaymentMilestonesByOrderSQL, strings.TrimSpace(subcontractOrderID))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	milestones := make([]productiondomain.SubcontractPaymentMilestone, 0)
	for rows.Next() {
		milestone, err := scanPostgresSubcontractPaymentMilestone(rows)
		if err != nil {
			return nil, err
		}
		milestones = append(milestones, milestone)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return milestones, nil
}

func (s PostgresSubcontractPaymentMilestoneStore) withTx(ctx context.Context, fn func(*sql.Tx) error) error {
	if tx, ok := postgresSubcontractTxFromContext(ctx); ok {
		return fn(tx)
	}

	tx, err := s.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
	if err != nil {
		return fmt.Errorf("begin subcontract payment milestone transaction: %w", err)
	}
	committed := false
	defer func() {
		if !committed {
			_ = tx.Rollback()
		}
	}()

	if err := fn(tx); err != nil {
		return err
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit subcontract payment milestone transaction: %w", err)
	}
	committed = true

	return nil
}

func (s PostgresSubcontractPaymentMilestoneStore) resolveOrgID(
	ctx context.Context,
	queryer postgresSubcontractPaymentMilestoneQueryer,
	orgRef string,
) (string, error) {
	orgRef = strings.TrimSpace(orgRef)
	if isPostgresSubcontractPaymentMilestoneUUIDText(orgRef) {
		return orgRef, nil
	}
	if orgRef != "" {
		var orgID string
		err := queryer.QueryRowContext(ctx, selectSubcontractPaymentMilestoneOrgIDSQL, orgRef).Scan(&orgID)
		if err == nil && isPostgresSubcontractPaymentMilestoneUUIDText(orgID) {
			return orgID, nil
		}
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			return "", fmt.Errorf("resolve subcontract payment milestone org %q: %w", orgRef, err)
		}
	}
	if isPostgresSubcontractPaymentMilestoneUUIDText(s.defaultOrgID) {
		return s.defaultOrgID, nil
	}

	return "", fmt.Errorf("subcontract payment milestone org %q cannot be resolved", orgRef)
}

func scanPostgresSubcontractPaymentMilestone(
	row interface{ Scan(dest ...any) error },
) (productiondomain.SubcontractPaymentMilestone, error) {
	header, err := scanPostgresSubcontractPaymentMilestoneHeader(row)
	if err != nil {
		return productiondomain.SubcontractPaymentMilestone{}, err
	}

	return buildPostgresSubcontractPaymentMilestone(header)
}

func scanPostgresSubcontractPaymentMilestoneHeader(
	row interface{ Scan(dest ...any) error },
) (postgresSubcontractPaymentMilestoneHeader, error) {
	var header postgresSubcontractPaymentMilestoneHeader
	err := row.Scan(
		&header.PersistedID,
		&header.ID,
		&header.OrgID,
		&header.MilestoneNo,
		&header.SubcontractOrderID,
		&header.SubcontractOrderNo,
		&header.FactoryID,
		&header.FactoryCode,
		&header.FactoryName,
		&header.Kind,
		&header.Status,
		&header.Amount,
		&header.CurrencyCode,
		&header.Note,
		&header.BlockReason,
		&header.ApprovedExceptionID,
		&header.RecordedBy,
		&header.RecordedAt,
		&header.ReadyBy,
		&header.ReadyAt,
		&header.BlockedBy,
		&header.BlockedAt,
		&header.CreatedAt,
		&header.CreatedBy,
		&header.UpdatedAt,
		&header.UpdatedBy,
		&header.Version,
	)
	if err != nil {
		return postgresSubcontractPaymentMilestoneHeader{}, err
	}

	return header, nil
}

func buildPostgresSubcontractPaymentMilestone(
	header postgresSubcontractPaymentMilestoneHeader,
) (productiondomain.SubcontractPaymentMilestone, error) {
	amount, err := decimal.ParseMoneyAmount(header.Amount)
	if err != nil {
		return productiondomain.SubcontractPaymentMilestone{}, err
	}
	milestone, err := productiondomain.NewSubcontractPaymentMilestone(productiondomain.NewSubcontractPaymentMilestoneInput{
		ID:                  header.ID,
		OrgID:               header.OrgID,
		MilestoneNo:         header.MilestoneNo,
		SubcontractOrderID:  header.SubcontractOrderID,
		SubcontractOrderNo:  header.SubcontractOrderNo,
		FactoryID:           header.FactoryID,
		FactoryCode:         header.FactoryCode,
		FactoryName:         header.FactoryName,
		Kind:                productiondomain.SubcontractPaymentMilestoneKind(header.Kind),
		Status:              productiondomain.SubcontractPaymentMilestoneStatus(header.Status),
		Amount:              amount,
		CurrencyCode:        header.CurrencyCode,
		Note:                header.Note,
		BlockReason:         header.BlockReason,
		ApprovedExceptionID: header.ApprovedExceptionID,
		RecordedBy:          header.RecordedBy,
		RecordedAt:          nullablePostgresSubcontractPaymentMilestoneTimeValue(header.RecordedAt),
		ReadyBy:             header.ReadyBy,
		ReadyAt:             nullablePostgresSubcontractPaymentMilestoneTimeValue(header.ReadyAt),
		BlockedBy:           header.BlockedBy,
		BlockedAt:           nullablePostgresSubcontractPaymentMilestoneTimeValue(header.BlockedAt),
		CreatedAt:           header.CreatedAt,
		CreatedBy:           header.CreatedBy,
		UpdatedAt:           header.UpdatedAt,
		UpdatedBy:           header.UpdatedBy,
	})
	if err != nil {
		return productiondomain.SubcontractPaymentMilestone{}, err
	}
	milestone.Version = header.Version
	if err := milestone.Validate(); err != nil {
		return productiondomain.SubcontractPaymentMilestone{}, err
	}

	return milestone, nil
}

func upsertPostgresSubcontractPaymentMilestone(
	ctx context.Context,
	queryer postgresSubcontractPaymentMilestoneQueryer,
	orgID string,
	milestone productiondomain.SubcontractPaymentMilestone,
) (string, error) {
	var persistedID string
	err := queryer.QueryRowContext(
		ctx,
		upsertSubcontractPaymentMilestoneSQL,
		orgID,
		milestone.ID,
		milestone.OrgID,
		milestone.MilestoneNo,
		milestone.SubcontractOrderID,
		milestone.SubcontractOrderNo,
		milestone.FactoryID,
		nullablePostgresSubcontractPaymentMilestoneText(milestone.FactoryCode),
		milestone.FactoryName,
		string(milestone.Kind),
		string(milestone.Status),
		milestone.Amount.String(),
		milestone.CurrencyCode.String(),
		nullablePostgresSubcontractPaymentMilestoneText(milestone.Note),
		nullablePostgresSubcontractPaymentMilestoneText(milestone.BlockReason),
		nullablePostgresSubcontractPaymentMilestoneText(milestone.ApprovedExceptionID),
		nullablePostgresSubcontractPaymentMilestoneText(milestone.RecordedBy),
		nullablePostgresSubcontractPaymentMilestoneTime(milestone.RecordedAt),
		nullablePostgresSubcontractPaymentMilestoneText(milestone.ReadyBy),
		nullablePostgresSubcontractPaymentMilestoneTime(milestone.ReadyAt),
		nullablePostgresSubcontractPaymentMilestoneText(milestone.BlockedBy),
		nullablePostgresSubcontractPaymentMilestoneTime(milestone.BlockedAt),
		milestone.CreatedAt.UTC(),
		milestone.CreatedBy,
		milestone.UpdatedAt.UTC(),
		milestone.UpdatedBy,
		milestone.Version,
	).Scan(&persistedID)
	if err != nil {
		return "", fmt.Errorf("upsert subcontract payment milestone %q: %w", milestone.ID, err)
	}

	return persistedID, nil
}

func nullablePostgresSubcontractPaymentMilestoneText(value string) any {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil
	}

	return value
}

func nullablePostgresSubcontractPaymentMilestoneTime(value time.Time) any {
	if value.IsZero() {
		return nil
	}

	return value.UTC()
}

func nullablePostgresSubcontractPaymentMilestoneTimeValue(value sql.NullTime) time.Time {
	if !value.Valid {
		return time.Time{}
	}

	return value.Time.UTC()
}

func isPostgresSubcontractPaymentMilestoneUUIDText(value string) bool {
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
			if !isPostgresSubcontractPaymentMilestoneHexText(char) {
				return false
			}
		}
	}

	return true
}

func isPostgresSubcontractPaymentMilestoneHexText(char rune) bool {
	return (char >= '0' && char <= '9') ||
		(char >= 'a' && char <= 'f') ||
		(char >= 'A' && char <= 'F')
}

var _ SubcontractPaymentMilestoneStore = PostgresSubcontractPaymentMilestoneStore{}
