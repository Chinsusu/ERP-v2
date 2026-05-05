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

type PostgresSubcontractFactoryDispatchStoreConfig struct {
	DefaultOrgID string
}

type PostgresSubcontractFactoryDispatchStore struct {
	db           *sql.DB
	defaultOrgID string
}

type postgresSubcontractFactoryDispatchQueryer interface {
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
}

type postgresSubcontractFactoryDispatchHeader struct {
	PersistedID            string
	ID                     string
	OrgID                  string
	DispatchNo             string
	SubcontractOrderID     string
	SubcontractOrderNo     string
	SourceProductionPlanID string
	SourceProductionPlanNo string
	FactoryID              string
	FactoryCode            string
	FactoryName            string
	FinishedItemID         string
	FinishedSKUCode        string
	FinishedItemName       string
	PlannedQty             string
	UOMCode                string
	SpecSummary            string
	SampleRequired         bool
	TargetStartDate        string
	ExpectedReceiptDate    string
	Status                 string
	ReadyAt                sql.NullTime
	ReadyBy                string
	SentAt                 sql.NullTime
	SentBy                 string
	RespondedAt            sql.NullTime
	ResponseBy             string
	FactoryResponseNote    string
	Note                   string
	CreatedAt              time.Time
	CreatedBy              string
	UpdatedAt              time.Time
	UpdatedBy              string
	Version                int
}

func NewPostgresSubcontractFactoryDispatchStore(
	db *sql.DB,
	cfg PostgresSubcontractFactoryDispatchStoreConfig,
) PostgresSubcontractFactoryDispatchStore {
	return PostgresSubcontractFactoryDispatchStore{
		db:           db,
		defaultOrgID: strings.TrimSpace(cfg.DefaultOrgID),
	}
}

const selectSubcontractFactoryDispatchOrgIDSQL = `
SELECT id::text
FROM core.organizations
WHERE code = $1
LIMIT 1`

const selectSubcontractFactoryDispatchHeadersBaseSQL = `
SELECT
  dispatch.id::text,
  dispatch.dispatch_ref,
  dispatch.org_ref,
  dispatch.dispatch_no,
  dispatch.subcontract_order_ref,
  dispatch.subcontract_order_no,
  COALESCE(dispatch.source_production_plan_ref, ''),
  COALESCE(dispatch.source_production_plan_no, ''),
  dispatch.factory_ref,
  COALESCE(dispatch.factory_code, ''),
  dispatch.factory_name,
  dispatch.finished_item_ref,
  dispatch.finished_sku_code,
  dispatch.finished_item_name,
  dispatch.planned_qty::text,
  dispatch.uom_code,
  COALESCE(dispatch.spec_summary, ''),
  dispatch.sample_required,
  COALESCE(dispatch.target_start_date::text, ''),
  COALESCE(dispatch.expected_receipt_date::text, ''),
  dispatch.status,
  dispatch.ready_at,
  COALESCE(dispatch.ready_by_ref, ''),
  dispatch.sent_at,
  COALESCE(dispatch.sent_by_ref, ''),
  dispatch.responded_at,
  COALESCE(dispatch.response_by_ref, ''),
  COALESCE(dispatch.factory_response_note, ''),
  COALESCE(dispatch.note, ''),
  dispatch.created_at,
  dispatch.created_by_ref,
  dispatch.updated_at,
  dispatch.updated_by_ref,
  dispatch.version
FROM subcontract.subcontract_factory_dispatches AS dispatch`

const findSubcontractFactoryDispatchHeaderSQL = selectSubcontractFactoryDispatchHeadersBaseSQL + `
WHERE lower(dispatch.dispatch_ref) = lower($1)
   OR dispatch.id::text = $1
   OR lower(dispatch.dispatch_no) = lower($1)
LIMIT 1`

const selectSubcontractFactoryDispatchesByOrderSQL = selectSubcontractFactoryDispatchHeadersBaseSQL + `
WHERE lower(dispatch.subcontract_order_ref) = lower($1)
   OR dispatch.subcontract_order_id::text = $1
   OR lower(dispatch.subcontract_order_no) = lower($1)
ORDER BY dispatch.created_at DESC, dispatch.dispatch_no DESC`

const findLatestSubcontractFactoryDispatchByOrderSQL = selectSubcontractFactoryDispatchHeadersBaseSQL + `
WHERE lower(dispatch.subcontract_order_ref) = lower($1)
   OR dispatch.subcontract_order_id::text = $1
   OR lower(dispatch.subcontract_order_no) = lower($1)
ORDER BY dispatch.created_at DESC, dispatch.dispatch_no DESC
LIMIT 1`

const selectSubcontractFactoryDispatchLinesSQL = `
SELECT
  line.line_ref,
  line.line_no,
  line.order_material_line_ref,
  line.item_ref,
  line.sku_code,
  line.item_name,
  line.planned_qty::text,
  line.uom_code,
  line.lot_trace_required,
  COALESCE(line.note, '')
FROM subcontract.subcontract_factory_dispatch_lines AS line
WHERE line.factory_dispatch_id = $1::uuid
ORDER BY line.line_no, line.line_ref`

const selectSubcontractFactoryDispatchEvidenceSQL = `
SELECT
  evidence.evidence_ref,
  evidence.evidence_type,
  COALESCE(evidence.file_name, ''),
  COALESCE(evidence.object_key, ''),
  COALESCE(evidence.external_url, ''),
  COALESCE(evidence.note, '')
FROM subcontract.subcontract_factory_dispatch_evidence AS evidence
WHERE evidence.factory_dispatch_id = $1::uuid
ORDER BY evidence.evidence_type, evidence.evidence_ref`

const upsertSubcontractFactoryDispatchSQL = `
INSERT INTO subcontract.subcontract_factory_dispatches (
  id,
  org_id,
  org_ref,
  dispatch_ref,
  dispatch_no,
  subcontract_order_id,
  subcontract_order_ref,
  subcontract_order_no,
  source_production_plan_ref,
  source_production_plan_no,
  factory_ref,
  factory_code,
  factory_name,
  finished_item_ref,
  finished_sku_code,
  finished_item_name,
  planned_qty,
  uom_code,
  spec_summary,
  sample_required,
  target_start_date,
  expected_receipt_date,
  status,
  ready_at,
  ready_by_ref,
  sent_at,
  sent_by_ref,
  responded_at,
  response_by_ref,
  factory_response_note,
  note,
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
        OR lower(subcontract_order.order_no) = lower($6)
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
  $19::date,
  $20::date,
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
  $34
)
ON CONFLICT (org_id, dispatch_ref)
DO UPDATE SET
  org_ref = EXCLUDED.org_ref,
  dispatch_no = EXCLUDED.dispatch_no,
  subcontract_order_id = EXCLUDED.subcontract_order_id,
  subcontract_order_ref = EXCLUDED.subcontract_order_ref,
  subcontract_order_no = EXCLUDED.subcontract_order_no,
  source_production_plan_ref = EXCLUDED.source_production_plan_ref,
  source_production_plan_no = EXCLUDED.source_production_plan_no,
  factory_ref = EXCLUDED.factory_ref,
  factory_code = EXCLUDED.factory_code,
  factory_name = EXCLUDED.factory_name,
  finished_item_ref = EXCLUDED.finished_item_ref,
  finished_sku_code = EXCLUDED.finished_sku_code,
  finished_item_name = EXCLUDED.finished_item_name,
  planned_qty = EXCLUDED.planned_qty,
  uom_code = EXCLUDED.uom_code,
  spec_summary = EXCLUDED.spec_summary,
  sample_required = EXCLUDED.sample_required,
  target_start_date = EXCLUDED.target_start_date,
  expected_receipt_date = EXCLUDED.expected_receipt_date,
  status = EXCLUDED.status,
  ready_at = EXCLUDED.ready_at,
  ready_by_ref = EXCLUDED.ready_by_ref,
  sent_at = EXCLUDED.sent_at,
  sent_by_ref = EXCLUDED.sent_by_ref,
  responded_at = EXCLUDED.responded_at,
  response_by_ref = EXCLUDED.response_by_ref,
  factory_response_note = EXCLUDED.factory_response_note,
  note = EXCLUDED.note,
  created_at = EXCLUDED.created_at,
  created_by_ref = EXCLUDED.created_by_ref,
  updated_at = EXCLUDED.updated_at,
  updated_by_ref = EXCLUDED.updated_by_ref,
  version = EXCLUDED.version
RETURNING id::text`

const deleteSubcontractFactoryDispatchLinesSQL = `
DELETE FROM subcontract.subcontract_factory_dispatch_lines
WHERE factory_dispatch_id = $1::uuid`

const insertSubcontractFactoryDispatchLineSQL = `
INSERT INTO subcontract.subcontract_factory_dispatch_lines (
  id,
  org_id,
  factory_dispatch_id,
  line_ref,
  line_no,
  order_material_line_ref,
  item_ref,
  sku_code,
  item_name,
  planned_qty,
  uom_code,
  lot_trace_required,
  note
) VALUES (
  COALESCE(CASE WHEN NULLIF($1::text, '') ~* '^[0-9a-f]{8}-[0-9a-f]{4}-[1-5][0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$' THEN $1::uuid END, gen_random_uuid()),
  $2::uuid,
  $3::uuid,
  $1,
  $4,
  $5,
  $6,
  $7,
  $8,
  $9,
  $10,
  $11,
  $12
)`

const deleteSubcontractFactoryDispatchEvidenceSQL = `
DELETE FROM subcontract.subcontract_factory_dispatch_evidence
WHERE factory_dispatch_id = $1::uuid`

const insertSubcontractFactoryDispatchEvidenceSQL = `
INSERT INTO subcontract.subcontract_factory_dispatch_evidence (
  id,
  org_id,
  factory_dispatch_id,
  evidence_ref,
  evidence_type,
  file_name,
  object_key,
  external_url,
  note
) VALUES (
  COALESCE(CASE WHEN NULLIF($1::text, '') ~* '^[0-9a-f]{8}-[0-9a-f]{4}-[1-5][0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$' THEN $1::uuid END, gen_random_uuid()),
  $2::uuid,
  $3::uuid,
  $1,
  $4,
  $5,
  $6,
  $7,
  $8
)`

func (s PostgresSubcontractFactoryDispatchStore) Save(
	ctx context.Context,
	dispatch productiondomain.SubcontractFactoryDispatch,
) error {
	if s.db == nil {
		return errors.New("database connection is required")
	}
	if err := dispatch.Validate(); err != nil {
		return err
	}
	return s.withTx(ctx, func(tx *sql.Tx) error {
		orgID, err := s.resolveOrgID(ctx, tx, dispatch.OrgID)
		if err != nil {
			return err
		}
		persistedID, err := upsertPostgresSubcontractFactoryDispatch(ctx, tx, orgID, dispatch)
		if err != nil {
			return err
		}
		if err := replacePostgresSubcontractFactoryDispatchLines(ctx, tx, orgID, persistedID, dispatch); err != nil {
			return err
		}
		return replacePostgresSubcontractFactoryDispatchEvidence(ctx, tx, orgID, persistedID, dispatch)
	})
}

func (s PostgresSubcontractFactoryDispatchStore) Get(
	ctx context.Context,
	id string,
) (productiondomain.SubcontractFactoryDispatch, error) {
	if s.db == nil {
		return productiondomain.SubcontractFactoryDispatch{}, errors.New("database connection is required")
	}
	row := s.db.QueryRowContext(ctx, findSubcontractFactoryDispatchHeaderSQL, strings.TrimSpace(id))
	dispatch, err := scanPostgresSubcontractFactoryDispatch(ctx, s.db, row)
	if errors.Is(err, sql.ErrNoRows) {
		return productiondomain.SubcontractFactoryDispatch{}, ErrSubcontractFactoryDispatchNotFound
	}
	if err != nil {
		return productiondomain.SubcontractFactoryDispatch{}, err
	}

	return dispatch, nil
}

func (s PostgresSubcontractFactoryDispatchStore) ListBySubcontractOrder(
	ctx context.Context,
	subcontractOrderID string,
) ([]productiondomain.SubcontractFactoryDispatch, error) {
	if s.db == nil {
		return nil, errors.New("database connection is required")
	}
	rows, err := s.db.QueryContext(ctx, selectSubcontractFactoryDispatchesByOrderSQL, strings.TrimSpace(subcontractOrderID))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	dispatches := make([]productiondomain.SubcontractFactoryDispatch, 0)
	for rows.Next() {
		dispatch, err := scanPostgresSubcontractFactoryDispatch(ctx, s.db, rows)
		if err != nil {
			return nil, err
		}
		dispatches = append(dispatches, dispatch)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return dispatches, nil
}

func (s PostgresSubcontractFactoryDispatchStore) GetLatestBySubcontractOrder(
	ctx context.Context,
	subcontractOrderID string,
) (productiondomain.SubcontractFactoryDispatch, error) {
	if s.db == nil {
		return productiondomain.SubcontractFactoryDispatch{}, errors.New("database connection is required")
	}
	row := s.db.QueryRowContext(ctx, findLatestSubcontractFactoryDispatchByOrderSQL, strings.TrimSpace(subcontractOrderID))
	dispatch, err := scanPostgresSubcontractFactoryDispatch(ctx, s.db, row)
	if errors.Is(err, sql.ErrNoRows) {
		return productiondomain.SubcontractFactoryDispatch{}, ErrSubcontractFactoryDispatchNotFound
	}
	if err != nil {
		return productiondomain.SubcontractFactoryDispatch{}, err
	}

	return dispatch, nil
}

func (s PostgresSubcontractFactoryDispatchStore) withTx(ctx context.Context, fn func(*sql.Tx) error) error {
	if tx, ok := postgresSubcontractTxFromContext(ctx); ok {
		return fn(tx)
	}

	tx, err := s.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
	if err != nil {
		return fmt.Errorf("begin subcontract factory dispatch transaction: %w", err)
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
		return fmt.Errorf("commit subcontract factory dispatch transaction: %w", err)
	}
	committed = true

	return nil
}

func (s PostgresSubcontractFactoryDispatchStore) resolveOrgID(
	ctx context.Context,
	queryer postgresSubcontractFactoryDispatchQueryer,
	orgRef string,
) (string, error) {
	orgRef = strings.TrimSpace(orgRef)
	if isPostgresSubcontractFactoryDispatchUUIDText(orgRef) {
		return orgRef, nil
	}
	if orgRef != "" {
		var orgID string
		err := queryer.QueryRowContext(ctx, selectSubcontractFactoryDispatchOrgIDSQL, orgRef).Scan(&orgID)
		if err == nil && isPostgresSubcontractFactoryDispatchUUIDText(orgID) {
			return orgID, nil
		}
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			return "", fmt.Errorf("resolve subcontract factory dispatch org %q: %w", orgRef, err)
		}
	}
	if isPostgresSubcontractFactoryDispatchUUIDText(s.defaultOrgID) {
		return s.defaultOrgID, nil
	}

	return "", fmt.Errorf("subcontract factory dispatch org %q cannot be resolved", orgRef)
}

func scanPostgresSubcontractFactoryDispatch(
	ctx context.Context,
	queryer postgresSubcontractFactoryDispatchQueryer,
	row interface{ Scan(dest ...any) error },
) (productiondomain.SubcontractFactoryDispatch, error) {
	header, err := scanPostgresSubcontractFactoryDispatchHeader(row)
	if err != nil {
		return productiondomain.SubcontractFactoryDispatch{}, err
	}
	lines, err := listPostgresSubcontractFactoryDispatchLines(ctx, queryer, header.PersistedID)
	if err != nil {
		return productiondomain.SubcontractFactoryDispatch{}, err
	}
	evidence, err := listPostgresSubcontractFactoryDispatchEvidence(ctx, queryer, header.PersistedID)
	if err != nil {
		return productiondomain.SubcontractFactoryDispatch{}, err
	}

	return buildPostgresSubcontractFactoryDispatch(header, lines, evidence)
}

func scanPostgresSubcontractFactoryDispatchHeader(
	row interface{ Scan(dest ...any) error },
) (postgresSubcontractFactoryDispatchHeader, error) {
	var header postgresSubcontractFactoryDispatchHeader
	err := row.Scan(
		&header.PersistedID,
		&header.ID,
		&header.OrgID,
		&header.DispatchNo,
		&header.SubcontractOrderID,
		&header.SubcontractOrderNo,
		&header.SourceProductionPlanID,
		&header.SourceProductionPlanNo,
		&header.FactoryID,
		&header.FactoryCode,
		&header.FactoryName,
		&header.FinishedItemID,
		&header.FinishedSKUCode,
		&header.FinishedItemName,
		&header.PlannedQty,
		&header.UOMCode,
		&header.SpecSummary,
		&header.SampleRequired,
		&header.TargetStartDate,
		&header.ExpectedReceiptDate,
		&header.Status,
		&header.ReadyAt,
		&header.ReadyBy,
		&header.SentAt,
		&header.SentBy,
		&header.RespondedAt,
		&header.ResponseBy,
		&header.FactoryResponseNote,
		&header.Note,
		&header.CreatedAt,
		&header.CreatedBy,
		&header.UpdatedAt,
		&header.UpdatedBy,
		&header.Version,
	)
	if err != nil {
		return postgresSubcontractFactoryDispatchHeader{}, err
	}

	return header, nil
}

func listPostgresSubcontractFactoryDispatchLines(
	ctx context.Context,
	queryer postgresSubcontractFactoryDispatchQueryer,
	persistedID string,
) ([]productiondomain.SubcontractFactoryDispatchLine, error) {
	rows, err := queryer.QueryContext(ctx, selectSubcontractFactoryDispatchLinesSQL, persistedID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	lines := make([]productiondomain.SubcontractFactoryDispatchLine, 0)
	for rows.Next() {
		line, err := scanPostgresSubcontractFactoryDispatchLine(rows)
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

func scanPostgresSubcontractFactoryDispatchLine(
	row interface{ Scan(dest ...any) error },
) (productiondomain.SubcontractFactoryDispatchLine, error) {
	var (
		id                  string
		lineNo              int
		orderMaterialLineID string
		itemID              string
		skuCode             string
		itemName            string
		plannedQtyText      string
		uomCode             string
		lotTraceRequired    bool
		note                string
	)
	if err := row.Scan(
		&id,
		&lineNo,
		&orderMaterialLineID,
		&itemID,
		&skuCode,
		&itemName,
		&plannedQtyText,
		&uomCode,
		&lotTraceRequired,
		&note,
	); err != nil {
		return productiondomain.SubcontractFactoryDispatchLine{}, err
	}
	plannedQty, err := decimal.ParseQuantity(plannedQtyText)
	if err != nil {
		return productiondomain.SubcontractFactoryDispatchLine{}, err
	}

	return productiondomain.NewSubcontractFactoryDispatchLine(productiondomain.NewSubcontractFactoryDispatchLineInput{
		ID:                  id,
		LineNo:              lineNo,
		OrderMaterialLineID: orderMaterialLineID,
		ItemID:              itemID,
		SKUCode:             skuCode,
		ItemName:            itemName,
		PlannedQty:          plannedQty,
		UOMCode:             uomCode,
		LotTraceRequired:    lotTraceRequired,
		Note:                note,
	})
}

func listPostgresSubcontractFactoryDispatchEvidence(
	ctx context.Context,
	queryer postgresSubcontractFactoryDispatchQueryer,
	persistedID string,
) ([]productiondomain.SubcontractFactoryDispatchEvidence, error) {
	rows, err := queryer.QueryContext(ctx, selectSubcontractFactoryDispatchEvidenceSQL, persistedID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	evidence := make([]productiondomain.SubcontractFactoryDispatchEvidence, 0)
	for rows.Next() {
		item, err := scanPostgresSubcontractFactoryDispatchEvidence(rows)
		if err != nil {
			return nil, err
		}
		evidence = append(evidence, item)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return evidence, nil
}

func scanPostgresSubcontractFactoryDispatchEvidence(
	row interface{ Scan(dest ...any) error },
) (productiondomain.SubcontractFactoryDispatchEvidence, error) {
	var (
		id           string
		evidenceType string
		fileName     string
		objectKey    string
		externalURL  string
		note         string
	)
	if err := row.Scan(&id, &evidenceType, &fileName, &objectKey, &externalURL, &note); err != nil {
		return productiondomain.SubcontractFactoryDispatchEvidence{}, err
	}

	return productiondomain.NewSubcontractFactoryDispatchEvidence(productiondomain.NewSubcontractFactoryDispatchEvidenceInput{
		ID:           id,
		EvidenceType: evidenceType,
		FileName:     fileName,
		ObjectKey:    objectKey,
		ExternalURL:  externalURL,
		Note:         note,
	})
}

func buildPostgresSubcontractFactoryDispatch(
	header postgresSubcontractFactoryDispatchHeader,
	lines []productiondomain.SubcontractFactoryDispatchLine,
	evidence []productiondomain.SubcontractFactoryDispatchEvidence,
) (productiondomain.SubcontractFactoryDispatch, error) {
	plannedQty, err := decimal.ParseQuantity(header.PlannedQty)
	if err != nil {
		return productiondomain.SubcontractFactoryDispatch{}, err
	}
	lineInputs := make([]productiondomain.NewSubcontractFactoryDispatchLineInput, 0, len(lines))
	for _, line := range lines {
		lineInputs = append(lineInputs, productiondomain.NewSubcontractFactoryDispatchLineInput{
			ID:                  line.ID,
			LineNo:              line.LineNo,
			OrderMaterialLineID: line.OrderMaterialLineID,
			ItemID:              line.ItemID,
			SKUCode:             line.SKUCode,
			ItemName:            line.ItemName,
			PlannedQty:          line.PlannedQty,
			UOMCode:             line.UOMCode.String(),
			LotTraceRequired:    line.LotTraceRequired,
			Note:                line.Note,
		})
	}
	evidenceInputs := make([]productiondomain.NewSubcontractFactoryDispatchEvidenceInput, 0, len(evidence))
	for _, item := range evidence {
		evidenceInputs = append(evidenceInputs, productiondomain.NewSubcontractFactoryDispatchEvidenceInput{
			ID:           item.ID,
			EvidenceType: item.EvidenceType,
			FileName:     item.FileName,
			ObjectKey:    item.ObjectKey,
			ExternalURL:  item.ExternalURL,
			Note:         item.Note,
		})
	}

	dispatch, err := productiondomain.NewSubcontractFactoryDispatch(productiondomain.NewSubcontractFactoryDispatchInput{
		ID:                     header.ID,
		OrgID:                  header.OrgID,
		DispatchNo:             header.DispatchNo,
		SubcontractOrderID:     header.SubcontractOrderID,
		SubcontractOrderNo:     header.SubcontractOrderNo,
		SourceProductionPlanID: header.SourceProductionPlanID,
		SourceProductionPlanNo: header.SourceProductionPlanNo,
		FactoryID:              header.FactoryID,
		FactoryCode:            header.FactoryCode,
		FactoryName:            header.FactoryName,
		FinishedItemID:         header.FinishedItemID,
		FinishedSKUCode:        header.FinishedSKUCode,
		FinishedItemName:       header.FinishedItemName,
		PlannedQty:             plannedQty,
		UOMCode:                header.UOMCode,
		SpecSummary:            header.SpecSummary,
		SampleRequired:         header.SampleRequired,
		TargetStartDate:        header.TargetStartDate,
		ExpectedReceiptDate:    header.ExpectedReceiptDate,
		Status:                 productiondomain.SubcontractFactoryDispatchStatus(header.Status),
		Lines:                  lineInputs,
		Evidence:               evidenceInputs,
		ReadyAt:                nullablePostgresSubcontractFactoryDispatchTimeValue(header.ReadyAt),
		ReadyBy:                header.ReadyBy,
		SentAt:                 nullablePostgresSubcontractFactoryDispatchTimeValue(header.SentAt),
		SentBy:                 header.SentBy,
		RespondedAt:            nullablePostgresSubcontractFactoryDispatchTimeValue(header.RespondedAt),
		ResponseBy:             header.ResponseBy,
		FactoryResponseNote:    header.FactoryResponseNote,
		Note:                   header.Note,
		CreatedAt:              header.CreatedAt,
		CreatedBy:              header.CreatedBy,
		UpdatedAt:              header.UpdatedAt,
		UpdatedBy:              header.UpdatedBy,
	})
	if err != nil {
		return productiondomain.SubcontractFactoryDispatch{}, err
	}
	dispatch.Version = header.Version
	if err := dispatch.Validate(); err != nil {
		return productiondomain.SubcontractFactoryDispatch{}, err
	}

	return dispatch, nil
}

func upsertPostgresSubcontractFactoryDispatch(
	ctx context.Context,
	queryer postgresSubcontractFactoryDispatchQueryer,
	orgID string,
	dispatch productiondomain.SubcontractFactoryDispatch,
) (string, error) {
	var persistedID string
	err := queryer.QueryRowContext(
		ctx,
		upsertSubcontractFactoryDispatchSQL,
		orgID,
		dispatch.ID,
		dispatch.OrgID,
		dispatch.DispatchNo,
		dispatch.SubcontractOrderID,
		dispatch.SubcontractOrderNo,
		nullablePostgresSubcontractFactoryDispatchText(dispatch.SourceProductionPlanID),
		nullablePostgresSubcontractFactoryDispatchText(dispatch.SourceProductionPlanNo),
		dispatch.FactoryID,
		nullablePostgresSubcontractFactoryDispatchText(dispatch.FactoryCode),
		dispatch.FactoryName,
		dispatch.FinishedItemID,
		dispatch.FinishedSKUCode,
		dispatch.FinishedItemName,
		dispatch.PlannedQty.String(),
		dispatch.UOMCode.String(),
		nullablePostgresSubcontractFactoryDispatchText(dispatch.SpecSummary),
		dispatch.SampleRequired,
		nullablePostgresSubcontractFactoryDispatchText(dispatch.TargetStartDate),
		nullablePostgresSubcontractFactoryDispatchText(dispatch.ExpectedReceiptDate),
		string(dispatch.Status),
		nullablePostgresSubcontractFactoryDispatchTime(dispatch.ReadyAt),
		nullablePostgresSubcontractFactoryDispatchText(dispatch.ReadyBy),
		nullablePostgresSubcontractFactoryDispatchTime(dispatch.SentAt),
		nullablePostgresSubcontractFactoryDispatchText(dispatch.SentBy),
		nullablePostgresSubcontractFactoryDispatchTime(dispatch.RespondedAt),
		nullablePostgresSubcontractFactoryDispatchText(dispatch.ResponseBy),
		nullablePostgresSubcontractFactoryDispatchText(dispatch.FactoryResponseNote),
		nullablePostgresSubcontractFactoryDispatchText(dispatch.Note),
		dispatch.CreatedAt.UTC(),
		dispatch.CreatedBy,
		dispatch.UpdatedAt.UTC(),
		dispatch.UpdatedBy,
		dispatch.Version,
	).Scan(&persistedID)
	if err != nil {
		return "", fmt.Errorf("upsert subcontract factory dispatch %q: %w", dispatch.ID, err)
	}

	return persistedID, nil
}

func replacePostgresSubcontractFactoryDispatchLines(
	ctx context.Context,
	queryer postgresSubcontractFactoryDispatchQueryer,
	orgID string,
	persistedID string,
	dispatch productiondomain.SubcontractFactoryDispatch,
) error {
	if _, err := queryer.ExecContext(ctx, deleteSubcontractFactoryDispatchLinesSQL, persistedID); err != nil {
		return fmt.Errorf("delete subcontract factory dispatch lines: %w", err)
	}
	for _, line := range dispatch.Lines {
		if _, err := queryer.ExecContext(
			ctx,
			insertSubcontractFactoryDispatchLineSQL,
			line.ID,
			orgID,
			persistedID,
			line.LineNo,
			line.OrderMaterialLineID,
			line.ItemID,
			line.SKUCode,
			line.ItemName,
			line.PlannedQty.String(),
			line.UOMCode.String(),
			line.LotTraceRequired,
			nullablePostgresSubcontractFactoryDispatchText(line.Note),
		); err != nil {
			return fmt.Errorf("insert subcontract factory dispatch line %q: %w", line.ID, err)
		}
	}

	return nil
}

func replacePostgresSubcontractFactoryDispatchEvidence(
	ctx context.Context,
	queryer postgresSubcontractFactoryDispatchQueryer,
	orgID string,
	persistedID string,
	dispatch productiondomain.SubcontractFactoryDispatch,
) error {
	if _, err := queryer.ExecContext(ctx, deleteSubcontractFactoryDispatchEvidenceSQL, persistedID); err != nil {
		return fmt.Errorf("delete subcontract factory dispatch evidence: %w", err)
	}
	for _, evidence := range dispatch.Evidence {
		if _, err := queryer.ExecContext(
			ctx,
			insertSubcontractFactoryDispatchEvidenceSQL,
			evidence.ID,
			orgID,
			persistedID,
			evidence.EvidenceType,
			nullablePostgresSubcontractFactoryDispatchText(evidence.FileName),
			nullablePostgresSubcontractFactoryDispatchText(evidence.ObjectKey),
			nullablePostgresSubcontractFactoryDispatchText(evidence.ExternalURL),
			nullablePostgresSubcontractFactoryDispatchText(evidence.Note),
		); err != nil {
			return fmt.Errorf("insert subcontract factory dispatch evidence %q: %w", evidence.ID, err)
		}
	}

	return nil
}

func nullablePostgresSubcontractFactoryDispatchText(value string) any {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil
	}

	return value
}

func nullablePostgresSubcontractFactoryDispatchTime(value time.Time) any {
	if value.IsZero() {
		return nil
	}

	return value.UTC()
}

func nullablePostgresSubcontractFactoryDispatchTimeValue(value sql.NullTime) time.Time {
	if !value.Valid {
		return time.Time{}
	}

	return value.Time.UTC()
}

func isPostgresSubcontractFactoryDispatchUUIDText(value string) bool {
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
			if !isPostgresSubcontractFactoryDispatchHexText(char) {
				return false
			}
		}
	}

	return true
}

func isPostgresSubcontractFactoryDispatchHexText(char rune) bool {
	return (char >= '0' && char <= '9') ||
		(char >= 'a' && char <= 'f') ||
		(char >= 'A' && char <= 'F')
}

var _ SubcontractFactoryDispatchStore = PostgresSubcontractFactoryDispatchStore{}
