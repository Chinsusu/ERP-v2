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

type PostgresCODRemittanceStoreConfig struct {
	DefaultOrgID string
}

type PostgresCODRemittanceStore struct {
	db           *sql.DB
	defaultOrgID string
}

type postgresCODRemittanceQueryer interface {
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
}

func NewPostgresCODRemittanceStore(
	db *sql.DB,
	cfg PostgresCODRemittanceStoreConfig,
) PostgresCODRemittanceStore {
	return PostgresCODRemittanceStore{
		db:           db,
		defaultOrgID: strings.TrimSpace(cfg.DefaultOrgID),
	}
}

const selectCODRemittanceOrgIDSQL = `
SELECT id::text
FROM core.organizations
WHERE code = $1
LIMIT 1`

const selectCODRemittanceHeadersBaseSQL = `
SELECT
  remittance.id::text,
  COALESCE(remittance.remittance_ref, remittance.id::text),
  remittance.org_ref,
  remittance.remittance_no,
  remittance.carrier_ref,
  COALESCE(remittance.carrier_code, ''),
  remittance.carrier_name,
  remittance.status,
  remittance.business_date::text,
  remittance.expected_amount::text,
  remittance.remitted_amount::text,
  remittance.discrepancy_amount::text,
  remittance.currency_code,
  COALESCE(remittance.submitted_by_ref, ''),
  remittance.submitted_at,
  COALESCE(remittance.approved_by_ref, ''),
  remittance.approved_at,
  COALESCE(remittance.closed_by_ref, ''),
  remittance.closed_at,
  COALESCE(remittance.void_reason, ''),
  COALESCE(remittance.voided_by_ref, ''),
  remittance.voided_at,
  remittance.created_at,
  remittance.created_by_ref,
  remittance.updated_at,
  remittance.updated_by_ref,
  remittance.version
FROM finance.cod_remittances AS remittance`

const selectCODRemittanceHeadersSQL = selectCODRemittanceHeadersBaseSQL + `
ORDER BY remittance.business_date DESC, remittance.remittance_no DESC`

const findCODRemittanceHeaderSQL = selectCODRemittanceHeadersBaseSQL + `
WHERE lower(COALESCE(remittance.remittance_ref, remittance.id::text)) = lower($1)
   OR remittance.id::text = $1
   OR lower(remittance.remittance_no) = lower($1)
LIMIT 1`

const findCODRemittancePersistedSQL = `
SELECT id::text, org_id::text
FROM finance.cod_remittances
WHERE lower(COALESCE(remittance_ref, id::text)) = lower($1)
   OR id::text = $1
   OR lower(remittance_no) = lower($1)
LIMIT 1
FOR UPDATE`

const selectCODRemittanceLinesSQL = `
SELECT
  COALESCE(line.line_ref, line.id::text),
  line.receivable_ref,
  line.receivable_no,
  COALESCE(line.shipment_ref, ''),
  line.tracking_no,
  COALESCE(line.customer_name, ''),
  line.expected_amount::text,
  line.remitted_amount::text,
  line.discrepancy_amount::text,
  line.match_status
FROM finance.cod_remittance_lines AS line
WHERE line.cod_remittance_id = $1::uuid
ORDER BY line.created_at, COALESCE(line.line_ref, line.id::text)`

const selectCODRemittanceDiscrepanciesSQL = `
SELECT
  COALESCE(discrepancy.discrepancy_ref, discrepancy.id::text),
  discrepancy.line_ref,
  discrepancy.receivable_ref,
  discrepancy.discrepancy_type,
  discrepancy.status,
  discrepancy.amount::text,
  discrepancy.reason,
  discrepancy.owner_ref,
  discrepancy.recorded_by_ref,
  discrepancy.recorded_at,
  COALESCE(discrepancy.resolved_by_ref, ''),
  discrepancy.resolved_at,
  COALESCE(discrepancy.resolution, '')
FROM finance.cod_discrepancies AS discrepancy
WHERE discrepancy.cod_remittance_id = $1::uuid
ORDER BY discrepancy.recorded_at, COALESCE(discrepancy.discrepancy_ref, discrepancy.id::text)`

const upsertCODRemittanceSQL = `
INSERT INTO finance.cod_remittances (
  id,
  org_id,
  org_ref,
  remittance_ref,
  remittance_no,
  carrier_ref,
  carrier_code,
  carrier_name,
  status,
  business_date,
  expected_amount,
  remitted_amount,
  discrepancy_amount,
  currency_code,
  submitted_by_ref,
  submitted_at,
  approved_by_ref,
  approved_at,
  closed_by_ref,
  closed_at,
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
ON CONFLICT (org_id, lower(remittance_ref))
DO UPDATE SET
  org_ref = EXCLUDED.org_ref,
  remittance_no = EXCLUDED.remittance_no,
  carrier_ref = EXCLUDED.carrier_ref,
  carrier_code = EXCLUDED.carrier_code,
  carrier_name = EXCLUDED.carrier_name,
  status = EXCLUDED.status,
  business_date = EXCLUDED.business_date,
  expected_amount = EXCLUDED.expected_amount,
  remitted_amount = EXCLUDED.remitted_amount,
  discrepancy_amount = EXCLUDED.discrepancy_amount,
  currency_code = EXCLUDED.currency_code,
  submitted_by_ref = EXCLUDED.submitted_by_ref,
  submitted_at = EXCLUDED.submitted_at,
  approved_by_ref = EXCLUDED.approved_by_ref,
  approved_at = EXCLUDED.approved_at,
  closed_by_ref = EXCLUDED.closed_by_ref,
  closed_at = EXCLUDED.closed_at,
  void_reason = EXCLUDED.void_reason,
  voided_by_ref = EXCLUDED.voided_by_ref,
  voided_at = EXCLUDED.voided_at,
  updated_at = EXCLUDED.updated_at,
  updated_by_ref = EXCLUDED.updated_by_ref,
  version = EXCLUDED.version
RETURNING id::text`

const upsertCODRemittanceLineSQL = `
INSERT INTO finance.cod_remittance_lines (
  id,
  org_id,
  cod_remittance_id,
  line_ref,
  remittance_ref,
  receivable_ref,
  receivable_no,
  shipment_ref,
  tracking_no,
  customer_name,
  expected_amount,
  remitted_amount,
  discrepancy_amount,
  match_status,
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
  $12,
  $13,
  $14,
  $15,
  $16
)
ON CONFLICT (org_id, cod_remittance_id, lower(line_ref))
DO UPDATE SET
  remittance_ref = EXCLUDED.remittance_ref,
  receivable_ref = EXCLUDED.receivable_ref,
  receivable_no = EXCLUDED.receivable_no,
  shipment_ref = EXCLUDED.shipment_ref,
  tracking_no = EXCLUDED.tracking_no,
  customer_name = EXCLUDED.customer_name,
  expected_amount = EXCLUDED.expected_amount,
  remitted_amount = EXCLUDED.remitted_amount,
  discrepancy_amount = EXCLUDED.discrepancy_amount,
  match_status = EXCLUDED.match_status,
  updated_at = EXCLUDED.updated_at
RETURNING id::text`

const upsertCODDiscrepancySQL = `
INSERT INTO finance.cod_discrepancies (
  id,
  org_id,
  cod_remittance_id,
  cod_remittance_line_id,
  discrepancy_ref,
  remittance_ref,
  line_ref,
  receivable_ref,
  discrepancy_type,
  status,
  amount,
  reason,
  owner_ref,
  recorded_by_ref,
  recorded_at,
  resolved_by_ref,
  resolved_at,
  resolution,
  created_at,
  updated_at
) VALUES (
  COALESCE($1::uuid, gen_random_uuid()),
  $2::uuid,
  $3::uuid,
  $4::uuid,
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
  $20
)
ON CONFLICT (org_id, cod_remittance_id, lower(discrepancy_ref))
DO UPDATE SET
  cod_remittance_line_id = EXCLUDED.cod_remittance_line_id,
  remittance_ref = EXCLUDED.remittance_ref,
  line_ref = EXCLUDED.line_ref,
  receivable_ref = EXCLUDED.receivable_ref,
  discrepancy_type = EXCLUDED.discrepancy_type,
  status = EXCLUDED.status,
  amount = EXCLUDED.amount,
  reason = EXCLUDED.reason,
  owner_ref = EXCLUDED.owner_ref,
  recorded_by_ref = EXCLUDED.recorded_by_ref,
  recorded_at = EXCLUDED.recorded_at,
  resolved_by_ref = EXCLUDED.resolved_by_ref,
  resolved_at = EXCLUDED.resolved_at,
  resolution = EXCLUDED.resolution,
  updated_at = EXCLUDED.updated_at`

const selectCODRemittanceLineRefsSQL = `
SELECT line_ref
FROM finance.cod_remittance_lines
WHERE cod_remittance_id = $1::uuid`

const selectCODDiscrepancyRefsSQL = `
SELECT discrepancy_ref
FROM finance.cod_discrepancies
WHERE cod_remittance_id = $1::uuid`

const deleteCODRemittanceStaleLineSQL = `
DELETE FROM finance.cod_remittance_lines
WHERE cod_remittance_id = $1::uuid
  AND lower(line_ref) = lower($2)`

const deleteCODRemittanceStaleDiscrepancySQL = `
DELETE FROM finance.cod_discrepancies
WHERE cod_remittance_id = $1::uuid
  AND lower(discrepancy_ref) = lower($2)`

func (s PostgresCODRemittanceStore) List(
	ctx context.Context,
	filter CODRemittanceFilter,
) ([]financedomain.CODRemittance, error) {
	if s.db == nil {
		return nil, errors.New("database connection is required")
	}
	rows, err := s.db.QueryContext(ctx, selectCODRemittanceHeadersSQL)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	filter = normalizeCODRemittanceFilter(filter)
	remittances := make([]financedomain.CODRemittance, 0)
	for rows.Next() {
		remittance, err := scanPostgresCODRemittance(ctx, s.db, rows)
		if err != nil {
			return nil, err
		}
		if !matchesCODRemittanceFilter(remittance, filter) {
			continue
		}
		remittances = append(remittances, remittance)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return remittances, nil
}

func (s PostgresCODRemittanceStore) Get(
	ctx context.Context,
	id string,
) (financedomain.CODRemittance, error) {
	if s.db == nil {
		return financedomain.CODRemittance{}, errors.New("database connection is required")
	}
	row := s.db.QueryRowContext(ctx, findCODRemittanceHeaderSQL, strings.TrimSpace(id))
	remittance, err := scanPostgresCODRemittance(ctx, s.db, row)
	if errors.Is(err, sql.ErrNoRows) {
		return financedomain.CODRemittance{}, ErrCODRemittanceNotFound
	}
	if err != nil {
		return financedomain.CODRemittance{}, err
	}

	return remittance, nil
}

func (s PostgresCODRemittanceStore) Save(
	ctx context.Context,
	remittance financedomain.CODRemittance,
) error {
	if s.db == nil {
		return errors.New("database connection is required")
	}
	if err := remittance.Validate(); err != nil {
		return err
	}
	tx, err := s.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
	if err != nil {
		return fmt.Errorf("begin cod remittance transaction: %w", err)
	}
	committed := false
	defer func() {
		if !committed {
			_ = tx.Rollback()
		}
	}()

	persistedID, orgID, err := findPostgresCODRemittance(ctx, tx, remittance.ID)
	if err != nil {
		return err
	}
	if orgID == "" {
		orgID, err = s.resolveOrgID(ctx, tx, remittance.OrgID)
		if err != nil {
			return err
		}
	}
	persistedID, err = upsertPostgresCODRemittance(ctx, tx, persistedID, orgID, remittance)
	if err != nil {
		return err
	}
	lineIDsByRef, err := upsertPostgresCODRemittanceLines(ctx, tx, orgID, persistedID, remittance)
	if err != nil {
		return err
	}
	if err := upsertPostgresCODDiscrepancies(ctx, tx, orgID, persistedID, lineIDsByRef, remittance); err != nil {
		return err
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit cod remittance transaction: %w", err)
	}
	committed = true

	return nil
}

func (s PostgresCODRemittanceStore) resolveOrgID(
	ctx context.Context,
	queryer postgresCODRemittanceQueryer,
	orgRef string,
) (string, error) {
	orgRef = strings.TrimSpace(orgRef)
	if isPostgresCODRemittanceUUIDText(orgRef) {
		return orgRef, nil
	}
	if orgRef != "" {
		var orgID string
		err := queryer.QueryRowContext(ctx, selectCODRemittanceOrgIDSQL, orgRef).Scan(&orgID)
		if err == nil && isPostgresCODRemittanceUUIDText(orgID) {
			return orgID, nil
		}
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			return "", fmt.Errorf("resolve cod remittance org %q: %w", orgRef, err)
		}
	}
	if isPostgresCODRemittanceUUIDText(s.defaultOrgID) {
		return s.defaultOrgID, nil
	}

	return "", fmt.Errorf("cod remittance org %q cannot be resolved", orgRef)
}

func scanPostgresCODRemittance(
	ctx context.Context,
	queryer postgresCODRemittanceQueryer,
	row interface{ Scan(dest ...any) error },
) (financedomain.CODRemittance, error) {
	var (
		persistedID       string
		status            string
		businessDate      string
		expectedAmount    string
		remittedAmount    string
		discrepancyAmount string
		currencyCode      string
		submittedAt       sql.NullTime
		approvedAt        sql.NullTime
		closedAt          sql.NullTime
		voidedAt          sql.NullTime
		remittance        financedomain.CODRemittance
		err               error
	)
	if err := row.Scan(
		&persistedID,
		&remittance.ID,
		&remittance.OrgID,
		&remittance.RemittanceNo,
		&remittance.CarrierID,
		&remittance.CarrierCode,
		&remittance.CarrierName,
		&status,
		&businessDate,
		&expectedAmount,
		&remittedAmount,
		&discrepancyAmount,
		&currencyCode,
		&remittance.SubmittedBy,
		&submittedAt,
		&remittance.ApprovedBy,
		&approvedAt,
		&remittance.ClosedBy,
		&closedAt,
		&remittance.VoidReason,
		&remittance.VoidedBy,
		&voidedAt,
		&remittance.CreatedAt,
		&remittance.CreatedBy,
		&remittance.UpdatedAt,
		&remittance.UpdatedBy,
		&remittance.Version,
	); err != nil {
		return financedomain.CODRemittance{}, err
	}
	remittance.Status = financedomain.NormalizeCODRemittanceStatus(financedomain.CODRemittanceStatus(status))
	remittance.BusinessDate, err = time.Parse(time.DateOnly, businessDate)
	if err != nil {
		return financedomain.CODRemittance{}, err
	}
	remittance.ExpectedAmount, err = decimal.ParseMoneyAmount(expectedAmount)
	if err != nil {
		return financedomain.CODRemittance{}, err
	}
	remittance.RemittedAmount, err = decimal.ParseMoneyAmount(remittedAmount)
	if err != nil {
		return financedomain.CODRemittance{}, err
	}
	remittance.DiscrepancyAmount, err = decimal.ParseMoneyAmount(discrepancyAmount)
	if err != nil {
		return financedomain.CODRemittance{}, err
	}
	remittance.CurrencyCode, err = decimal.NormalizeCurrencyCode(currencyCode)
	if err != nil {
		return financedomain.CODRemittance{}, err
	}
	if submittedAt.Valid {
		remittance.SubmittedAt = submittedAt.Time.UTC()
	}
	if approvedAt.Valid {
		remittance.ApprovedAt = approvedAt.Time.UTC()
	}
	if closedAt.Valid {
		remittance.ClosedAt = closedAt.Time.UTC()
	}
	if voidedAt.Valid {
		remittance.VoidedAt = voidedAt.Time.UTC()
	}
	remittance.CreatedAt = remittance.CreatedAt.UTC()
	remittance.UpdatedAt = remittance.UpdatedAt.UTC()
	remittance.Lines, err = listPostgresCODRemittanceLines(ctx, queryer, persistedID)
	if err != nil {
		return financedomain.CODRemittance{}, err
	}
	remittance.Discrepancies, err = listPostgresCODRemittanceDiscrepancies(ctx, queryer, persistedID)
	if err != nil {
		return financedomain.CODRemittance{}, err
	}
	if err := remittance.Validate(); err != nil {
		return financedomain.CODRemittance{}, err
	}

	return remittance, nil
}

func listPostgresCODRemittanceLines(
	ctx context.Context,
	queryer postgresCODRemittanceQueryer,
	persistedID string,
) ([]financedomain.CODRemittanceLine, error) {
	rows, err := queryer.QueryContext(ctx, selectCODRemittanceLinesSQL, persistedID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	lines := make([]financedomain.CODRemittanceLine, 0)
	for rows.Next() {
		line, err := scanPostgresCODRemittanceLine(rows)
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

func scanPostgresCODRemittanceLine(
	row interface{ Scan(dest ...any) error },
) (financedomain.CODRemittanceLine, error) {
	var (
		line              financedomain.CODRemittanceLine
		expectedAmount    string
		remittedAmount    string
		discrepancyAmount string
		matchStatus       string
		err               error
	)
	if err := row.Scan(
		&line.ID,
		&line.ReceivableID,
		&line.ReceivableNo,
		&line.ShipmentID,
		&line.TrackingNo,
		&line.CustomerName,
		&expectedAmount,
		&remittedAmount,
		&discrepancyAmount,
		&matchStatus,
	); err != nil {
		return financedomain.CODRemittanceLine{}, err
	}
	line.ExpectedAmount, err = decimal.ParseMoneyAmount(expectedAmount)
	if err != nil {
		return financedomain.CODRemittanceLine{}, err
	}
	line.RemittedAmount, err = decimal.ParseMoneyAmount(remittedAmount)
	if err != nil {
		return financedomain.CODRemittanceLine{}, err
	}
	line.DiscrepancyAmount, err = decimal.ParseMoneyAmount(discrepancyAmount)
	if err != nil {
		return financedomain.CODRemittanceLine{}, err
	}
	line.MatchStatus = financedomain.NormalizeCODLineMatchStatus(financedomain.CODLineMatchStatus(matchStatus))
	if err := line.Validate(); err != nil {
		return financedomain.CODRemittanceLine{}, err
	}

	return line, nil
}

func listPostgresCODRemittanceDiscrepancies(
	ctx context.Context,
	queryer postgresCODRemittanceQueryer,
	persistedID string,
) ([]financedomain.CODDiscrepancy, error) {
	rows, err := queryer.QueryContext(ctx, selectCODRemittanceDiscrepanciesSQL, persistedID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	discrepancies := make([]financedomain.CODDiscrepancy, 0)
	for rows.Next() {
		discrepancy, err := scanPostgresCODDiscrepancy(rows)
		if err != nil {
			return nil, err
		}
		discrepancies = append(discrepancies, discrepancy)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return discrepancies, nil
}

func scanPostgresCODDiscrepancy(
	row interface{ Scan(dest ...any) error },
) (financedomain.CODDiscrepancy, error) {
	var (
		discrepancy financedomain.CODDiscrepancy
		kind        string
		status      string
		amount      string
		resolvedAt  sql.NullTime
		err         error
	)
	if err := row.Scan(
		&discrepancy.ID,
		&discrepancy.LineID,
		&discrepancy.ReceivableID,
		&kind,
		&status,
		&amount,
		&discrepancy.Reason,
		&discrepancy.OwnerID,
		&discrepancy.RecordedBy,
		&discrepancy.RecordedAt,
		&discrepancy.ResolvedBy,
		&resolvedAt,
		&discrepancy.Resolution,
	); err != nil {
		return financedomain.CODDiscrepancy{}, err
	}
	discrepancy.Type = financedomain.NormalizeCODDiscrepancyType(financedomain.CODDiscrepancyType(kind))
	discrepancy.Status = financedomain.NormalizeCODDiscrepancyStatus(financedomain.CODDiscrepancyStatus(status))
	discrepancy.Amount, err = decimal.ParseMoneyAmount(amount)
	if err != nil {
		return financedomain.CODDiscrepancy{}, err
	}
	discrepancy.RecordedAt = discrepancy.RecordedAt.UTC()
	if resolvedAt.Valid {
		discrepancy.ResolvedAt = resolvedAt.Time.UTC()
	}
	if err := discrepancy.Validate(); err != nil {
		return financedomain.CODDiscrepancy{}, err
	}

	return discrepancy, nil
}

func findPostgresCODRemittance(
	ctx context.Context,
	queryer postgresCODRemittanceQueryer,
	id string,
) (string, string, error) {
	var persistedID string
	var orgID string
	err := queryer.QueryRowContext(ctx, findCODRemittancePersistedSQL, strings.TrimSpace(id)).
		Scan(&persistedID, &orgID)
	if errors.Is(err, sql.ErrNoRows) {
		return "", "", nil
	}
	if err != nil {
		return "", "", fmt.Errorf("find cod remittance %q: %w", id, err)
	}

	return persistedID, orgID, nil
}

func upsertPostgresCODRemittance(
	ctx context.Context,
	queryer postgresCODRemittanceQueryer,
	persistedID string,
	orgID string,
	remittance financedomain.CODRemittance,
) (string, error) {
	var savedID string
	err := queryer.QueryRowContext(
		ctx,
		upsertCODRemittanceSQL,
		nullablePostgresCODRemittanceUUID(firstNonBlankPostgresCODRemittance(persistedID, remittance.ID)),
		orgID,
		remittance.OrgID,
		remittance.ID,
		remittance.RemittanceNo,
		remittance.CarrierID,
		nullablePostgresCODRemittanceText(remittance.CarrierCode),
		remittance.CarrierName,
		string(remittance.Status),
		postgresCODRemittanceDate(remittance.BusinessDate),
		remittance.ExpectedAmount.String(),
		remittance.RemittedAmount.String(),
		remittance.DiscrepancyAmount.String(),
		remittance.CurrencyCode.String(),
		nullablePostgresCODRemittanceText(remittance.SubmittedBy),
		nullablePostgresCODRemittanceTime(remittance.SubmittedAt),
		nullablePostgresCODRemittanceText(remittance.ApprovedBy),
		nullablePostgresCODRemittanceTime(remittance.ApprovedAt),
		nullablePostgresCODRemittanceText(remittance.ClosedBy),
		nullablePostgresCODRemittanceTime(remittance.ClosedAt),
		nullablePostgresCODRemittanceText(remittance.VoidReason),
		nullablePostgresCODRemittanceText(remittance.VoidedBy),
		nullablePostgresCODRemittanceTime(remittance.VoidedAt),
		postgresCODRemittanceTime(remittance.CreatedAt),
		remittance.CreatedBy,
		postgresCODRemittanceTime(remittance.UpdatedAt),
		remittance.UpdatedBy,
		remittance.Version,
	).Scan(&savedID)
	if err != nil {
		return "", fmt.Errorf("upsert cod remittance %q: %w", remittance.ID, err)
	}

	return savedID, nil
}

func upsertPostgresCODRemittanceLines(
	ctx context.Context,
	queryer postgresCODRemittanceQueryer,
	orgID string,
	persistedID string,
	remittance financedomain.CODRemittance,
) (map[string]string, error) {
	keptLineRefs := make(map[string]struct{}, len(remittance.Lines))
	lineIDsByRef := make(map[string]string, len(remittance.Lines))
	for _, line := range remittance.Lines {
		var savedLineID string
		err := queryer.QueryRowContext(
			ctx,
			upsertCODRemittanceLineSQL,
			nullablePostgresCODRemittanceUUID(line.ID),
			orgID,
			persistedID,
			line.ID,
			remittance.ID,
			line.ReceivableID,
			line.ReceivableNo,
			nullablePostgresCODRemittanceText(line.ShipmentID),
			line.TrackingNo,
			nullablePostgresCODRemittanceText(line.CustomerName),
			line.ExpectedAmount.String(),
			line.RemittedAmount.String(),
			line.DiscrepancyAmount.String(),
			string(line.MatchStatus),
			postgresCODRemittanceTime(remittance.CreatedAt),
			postgresCODRemittanceTime(remittance.UpdatedAt),
		).Scan(&savedLineID)
		if err != nil {
			return nil, fmt.Errorf("upsert cod remittance line %q: %w", line.ID, err)
		}
		normalizedRef := strings.ToLower(strings.TrimSpace(line.ID))
		keptLineRefs[normalizedRef] = struct{}{}
		lineIDsByRef[normalizedRef] = savedLineID
	}

	rows, err := queryer.QueryContext(ctx, selectCODRemittanceLineRefsSQL, persistedID)
	if err != nil {
		return nil, fmt.Errorf("list cod remittance line refs: %w", err)
	}
	defer rows.Close()

	lineRefs := make([]string, 0)
	for rows.Next() {
		var lineRef string
		if err := rows.Scan(&lineRef); err != nil {
			return nil, fmt.Errorf("scan cod remittance line ref: %w", err)
		}
		lineRefs = append(lineRefs, lineRef)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("scan cod remittance line refs: %w", err)
	}
	for _, lineRef := range lineRefs {
		if _, ok := keptLineRefs[strings.ToLower(strings.TrimSpace(lineRef))]; ok {
			continue
		}
		if _, err := queryer.ExecContext(ctx, deleteCODRemittanceStaleLineSQL, persistedID, lineRef); err != nil {
			return nil, fmt.Errorf("delete stale cod remittance line %q: %w", lineRef, err)
		}
	}

	return lineIDsByRef, nil
}

func upsertPostgresCODDiscrepancies(
	ctx context.Context,
	queryer postgresCODRemittanceQueryer,
	orgID string,
	persistedID string,
	lineIDsByRef map[string]string,
	remittance financedomain.CODRemittance,
) error {
	keptDiscrepancyRefs := make(map[string]struct{}, len(remittance.Discrepancies))
	for _, discrepancy := range remittance.Discrepancies {
		lineID, ok := lineIDsByRef[strings.ToLower(strings.TrimSpace(discrepancy.LineID))]
		if !ok {
			return fmt.Errorf("cod remittance discrepancy line %q cannot be resolved", discrepancy.LineID)
		}
		if _, err := queryer.ExecContext(
			ctx,
			upsertCODDiscrepancySQL,
			nullablePostgresCODRemittanceUUID(discrepancy.ID),
			orgID,
			persistedID,
			lineID,
			discrepancy.ID,
			remittance.ID,
			discrepancy.LineID,
			discrepancy.ReceivableID,
			string(discrepancy.Type),
			string(discrepancy.Status),
			discrepancy.Amount.String(),
			discrepancy.Reason,
			discrepancy.OwnerID,
			discrepancy.RecordedBy,
			postgresCODRemittanceTime(discrepancy.RecordedAt),
			nullablePostgresCODRemittanceText(discrepancy.ResolvedBy),
			nullablePostgresCODRemittanceTime(discrepancy.ResolvedAt),
			nullablePostgresCODRemittanceText(discrepancy.Resolution),
			postgresCODRemittanceTime(discrepancy.RecordedAt),
			postgresCODRemittanceTime(remittance.UpdatedAt),
		); err != nil {
			return fmt.Errorf("upsert cod discrepancy %q: %w", discrepancy.ID, err)
		}
		keptDiscrepancyRefs[strings.ToLower(strings.TrimSpace(discrepancy.ID))] = struct{}{}
	}

	rows, err := queryer.QueryContext(ctx, selectCODDiscrepancyRefsSQL, persistedID)
	if err != nil {
		return fmt.Errorf("list cod discrepancy refs: %w", err)
	}
	defer rows.Close()

	discrepancyRefs := make([]string, 0)
	for rows.Next() {
		var discrepancyRef string
		if err := rows.Scan(&discrepancyRef); err != nil {
			return fmt.Errorf("scan cod discrepancy ref: %w", err)
		}
		discrepancyRefs = append(discrepancyRefs, discrepancyRef)
	}
	if err := rows.Err(); err != nil {
		return fmt.Errorf("scan cod discrepancy refs: %w", err)
	}
	for _, discrepancyRef := range discrepancyRefs {
		if _, ok := keptDiscrepancyRefs[strings.ToLower(strings.TrimSpace(discrepancyRef))]; ok {
			continue
		}
		if _, err := queryer.ExecContext(ctx, deleteCODRemittanceStaleDiscrepancySQL, persistedID, discrepancyRef); err != nil {
			return fmt.Errorf("delete stale cod discrepancy %q: %w", discrepancyRef, err)
		}
	}

	return nil
}

func nullablePostgresCODRemittanceText(value string) any {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil
	}

	return value
}

func nullablePostgresCODRemittanceUUID(value string) any {
	value = strings.TrimSpace(value)
	if !isPostgresCODRemittanceUUIDText(value) {
		return nil
	}

	return value
}

func nullablePostgresCODRemittanceTime(value time.Time) any {
	if value.IsZero() {
		return nil
	}

	return value.UTC()
}

func postgresCODRemittanceTime(value time.Time) time.Time {
	if value.IsZero() {
		return time.Now().UTC()
	}

	return value.UTC()
}

func postgresCODRemittanceDate(value time.Time) time.Time {
	if value.IsZero() {
		return time.Date(1, time.January, 1, 0, 0, 0, 0, time.UTC)
	}

	return value.UTC()
}

func firstNonBlankPostgresCODRemittance(values ...string) string {
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value != "" {
			return value
		}
	}

	return ""
}

func isPostgresCODRemittanceUUIDText(value string) bool {
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
			if !isPostgresCODRemittanceHexText(char) {
				return false
			}
		}
	}

	return true
}

func isPostgresCODRemittanceHexText(char rune) bool {
	return (char >= '0' && char <= '9') ||
		(char >= 'a' && char <= 'f') ||
		(char >= 'A' && char <= 'F')
}

var _ CODRemittanceStore = PostgresCODRemittanceStore{}
