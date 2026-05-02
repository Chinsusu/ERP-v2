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

type PostgresSubcontractMaterialTransferStoreConfig struct {
	DefaultOrgID string
}

type PostgresSubcontractMaterialTransferStore struct {
	db           *sql.DB
	defaultOrgID string
}

type postgresSubcontractMaterialTransferQueryer interface {
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
}

type postgresSubcontractMaterialTransferHeader struct {
	PersistedID         string
	ID                  string
	OrgID               string
	TransferNo          string
	SubcontractOrderID  string
	SubcontractOrderNo  string
	FactoryID           string
	FactoryCode         string
	FactoryName         string
	SourceWarehouseID   string
	SourceWarehouseCode string
	Status              string
	HandoverBy          string
	HandoverAt          time.Time
	ReceivedBy          string
	ReceiverContact     string
	VehicleNo           string
	Note                string
	CreatedAt           time.Time
	CreatedBy           string
	UpdatedAt           time.Time
	UpdatedBy           string
	Version             int
}

func NewPostgresSubcontractMaterialTransferStore(
	db *sql.DB,
	cfg PostgresSubcontractMaterialTransferStoreConfig,
) PostgresSubcontractMaterialTransferStore {
	return PostgresSubcontractMaterialTransferStore{
		db:           db,
		defaultOrgID: strings.TrimSpace(cfg.DefaultOrgID),
	}
}

const selectSubcontractMaterialTransferOrgIDSQL = `
SELECT id::text
FROM core.organizations
WHERE code = $1
LIMIT 1`

const selectSubcontractMaterialTransferHeadersBaseSQL = `
SELECT
  transfer.id::text,
  transfer.transfer_ref,
  transfer.org_ref,
  transfer.transfer_no,
  transfer.subcontract_order_ref,
  transfer.subcontract_order_no,
  transfer.factory_ref,
  COALESCE(transfer.factory_code, ''),
  transfer.factory_name,
  transfer.source_warehouse_ref,
  COALESCE(transfer.source_warehouse_code, ''),
  transfer.status,
  transfer.handover_by_ref,
  transfer.handover_at,
  transfer.received_by_ref,
  COALESCE(transfer.receiver_contact, ''),
  COALESCE(transfer.vehicle_no, ''),
  COALESCE(transfer.note, ''),
  transfer.created_at,
  transfer.created_by_ref,
  transfer.updated_at,
  transfer.updated_by_ref,
  transfer.version
FROM subcontract.subcontract_material_transfers AS transfer`

const selectSubcontractMaterialTransfersByOrderSQL = selectSubcontractMaterialTransferHeadersBaseSQL + `
WHERE lower(transfer.subcontract_order_ref) = lower($1)
   OR transfer.subcontract_order_id::text = $1
   OR lower(transfer.subcontract_order_no) = lower($1)
ORDER BY transfer.handover_at DESC, transfer.transfer_no DESC`

const selectSubcontractMaterialTransferLinesSQL = `
SELECT
  line.line_ref,
  line.line_no,
  line.order_material_line_ref,
  line.item_ref,
  line.sku_code,
  line.item_name,
  line.issue_qty::text,
  line.uom_code,
  line.base_issue_qty::text,
  line.base_uom_code,
  line.conversion_factor::text,
  COALESCE(line.batch_ref, ''),
  COALESCE(line.batch_no, ''),
  COALESCE(line.lot_no, ''),
  COALESCE(line.source_bin_ref, ''),
  line.lot_trace_required,
  COALESCE(line.note, '')
FROM subcontract.subcontract_material_transfer_lines AS line
WHERE line.material_transfer_id = $1::uuid
ORDER BY line.line_no, line.line_ref`

const selectSubcontractMaterialTransferEvidenceSQL = `
SELECT
  evidence.evidence_ref,
  evidence.evidence_type,
  COALESCE(evidence.file_name, ''),
  COALESCE(evidence.object_key, ''),
  COALESCE(evidence.external_url, ''),
  COALESCE(evidence.note, '')
FROM subcontract.subcontract_material_transfer_evidence AS evidence
WHERE evidence.material_transfer_id = $1::uuid
ORDER BY evidence.evidence_type, evidence.evidence_ref`

const upsertSubcontractMaterialTransferSQL = `
INSERT INTO subcontract.subcontract_material_transfers (
  id,
  org_id,
  org_ref,
  transfer_ref,
  transfer_no,
  subcontract_order_id,
  subcontract_order_ref,
  subcontract_order_no,
  factory_ref,
  factory_code,
  factory_name,
  source_warehouse_ref,
  source_warehouse_code,
  status,
  handover_by_ref,
  handover_at,
  received_by_ref,
  receiver_contact,
  vehicle_no,
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
  $23
)
ON CONFLICT (org_id, transfer_ref)
DO UPDATE SET
  org_ref = EXCLUDED.org_ref,
  transfer_no = EXCLUDED.transfer_no,
  subcontract_order_id = EXCLUDED.subcontract_order_id,
  subcontract_order_ref = EXCLUDED.subcontract_order_ref,
  subcontract_order_no = EXCLUDED.subcontract_order_no,
  factory_ref = EXCLUDED.factory_ref,
  factory_code = EXCLUDED.factory_code,
  factory_name = EXCLUDED.factory_name,
  source_warehouse_ref = EXCLUDED.source_warehouse_ref,
  source_warehouse_code = EXCLUDED.source_warehouse_code,
  status = EXCLUDED.status,
  handover_by_ref = EXCLUDED.handover_by_ref,
  handover_at = EXCLUDED.handover_at,
  received_by_ref = EXCLUDED.received_by_ref,
  receiver_contact = EXCLUDED.receiver_contact,
  vehicle_no = EXCLUDED.vehicle_no,
  note = EXCLUDED.note,
  created_at = EXCLUDED.created_at,
  created_by_ref = EXCLUDED.created_by_ref,
  updated_at = EXCLUDED.updated_at,
  updated_by_ref = EXCLUDED.updated_by_ref,
  version = EXCLUDED.version
RETURNING id::text`

const deleteSubcontractMaterialTransferLinesSQL = `
DELETE FROM subcontract.subcontract_material_transfer_lines
WHERE material_transfer_id = $1::uuid`

const insertSubcontractMaterialTransferLineSQL = `
INSERT INTO subcontract.subcontract_material_transfer_lines (
  id,
  org_id,
  material_transfer_id,
  line_ref,
  line_no,
  order_material_line_ref,
  item_ref,
  sku_code,
  item_name,
  issue_qty,
  uom_code,
  base_issue_qty,
  base_uom_code,
  conversion_factor,
  batch_ref,
  batch_no,
  lot_no,
  source_bin_ref,
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
  $12,
  $13,
  $14,
  $15,
  $16,
  $17,
  $18,
  $19
)`

const deleteSubcontractMaterialTransferEvidenceSQL = `
DELETE FROM subcontract.subcontract_material_transfer_evidence
WHERE material_transfer_id = $1::uuid`

const insertSubcontractMaterialTransferEvidenceSQL = `
INSERT INTO subcontract.subcontract_material_transfer_evidence (
  id,
  org_id,
  material_transfer_id,
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

func (s PostgresSubcontractMaterialTransferStore) Save(
	ctx context.Context,
	transfer productiondomain.SubcontractMaterialTransfer,
) error {
	if s.db == nil {
		return errors.New("database connection is required")
	}
	if err := transfer.Validate(); err != nil {
		return err
	}
	return s.withTx(ctx, func(tx *sql.Tx) error {
		orgID, err := s.resolveOrgID(ctx, tx, transfer.OrgID)
		if err != nil {
			return err
		}
		persistedID, err := upsertPostgresSubcontractMaterialTransfer(ctx, tx, orgID, transfer)
		if err != nil {
			return err
		}
		if err := replacePostgresSubcontractMaterialTransferLines(ctx, tx, orgID, persistedID, transfer); err != nil {
			return err
		}
		if err := replacePostgresSubcontractMaterialTransferEvidence(ctx, tx, orgID, persistedID, transfer); err != nil {
			return err
		}

		return nil
	})
}

func (s PostgresSubcontractMaterialTransferStore) ListBySubcontractOrder(
	ctx context.Context,
	subcontractOrderID string,
) ([]productiondomain.SubcontractMaterialTransfer, error) {
	if s.db == nil {
		return nil, errors.New("database connection is required")
	}
	rows, err := s.db.QueryContext(ctx, selectSubcontractMaterialTransfersByOrderSQL, strings.TrimSpace(subcontractOrderID))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	transfers := make([]productiondomain.SubcontractMaterialTransfer, 0)
	for rows.Next() {
		transfer, err := scanPostgresSubcontractMaterialTransfer(ctx, s.db, rows)
		if err != nil {
			return nil, err
		}
		transfers = append(transfers, transfer)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return transfers, nil
}

func (s PostgresSubcontractMaterialTransferStore) withTx(ctx context.Context, fn func(*sql.Tx) error) error {
	tx, err := s.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
	if err != nil {
		return fmt.Errorf("begin subcontract material transfer transaction: %w", err)
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
		return fmt.Errorf("commit subcontract material transfer transaction: %w", err)
	}
	committed = true

	return nil
}

func (s PostgresSubcontractMaterialTransferStore) resolveOrgID(
	ctx context.Context,
	queryer postgresSubcontractMaterialTransferQueryer,
	orgRef string,
) (string, error) {
	orgRef = strings.TrimSpace(orgRef)
	if isPostgresSubcontractMaterialTransferUUIDText(orgRef) {
		return orgRef, nil
	}
	if orgRef != "" {
		var orgID string
		err := queryer.QueryRowContext(ctx, selectSubcontractMaterialTransferOrgIDSQL, orgRef).Scan(&orgID)
		if err == nil && isPostgresSubcontractMaterialTransferUUIDText(orgID) {
			return orgID, nil
		}
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			return "", fmt.Errorf("resolve subcontract material transfer org %q: %w", orgRef, err)
		}
	}
	if isPostgresSubcontractMaterialTransferUUIDText(s.defaultOrgID) {
		return s.defaultOrgID, nil
	}

	return "", fmt.Errorf("subcontract material transfer org %q cannot be resolved", orgRef)
}

func scanPostgresSubcontractMaterialTransfer(
	ctx context.Context,
	queryer postgresSubcontractMaterialTransferQueryer,
	row interface{ Scan(dest ...any) error },
) (productiondomain.SubcontractMaterialTransfer, error) {
	header, err := scanPostgresSubcontractMaterialTransferHeader(row)
	if err != nil {
		return productiondomain.SubcontractMaterialTransfer{}, err
	}
	lines, err := listPostgresSubcontractMaterialTransferLines(ctx, queryer, header.PersistedID)
	if err != nil {
		return productiondomain.SubcontractMaterialTransfer{}, err
	}
	evidence, err := listPostgresSubcontractMaterialTransferEvidence(ctx, queryer, header.PersistedID)
	if err != nil {
		return productiondomain.SubcontractMaterialTransfer{}, err
	}

	return buildPostgresSubcontractMaterialTransfer(header, lines, evidence)
}

func scanPostgresSubcontractMaterialTransferHeader(
	row interface{ Scan(dest ...any) error },
) (postgresSubcontractMaterialTransferHeader, error) {
	var header postgresSubcontractMaterialTransferHeader
	err := row.Scan(
		&header.PersistedID,
		&header.ID,
		&header.OrgID,
		&header.TransferNo,
		&header.SubcontractOrderID,
		&header.SubcontractOrderNo,
		&header.FactoryID,
		&header.FactoryCode,
		&header.FactoryName,
		&header.SourceWarehouseID,
		&header.SourceWarehouseCode,
		&header.Status,
		&header.HandoverBy,
		&header.HandoverAt,
		&header.ReceivedBy,
		&header.ReceiverContact,
		&header.VehicleNo,
		&header.Note,
		&header.CreatedAt,
		&header.CreatedBy,
		&header.UpdatedAt,
		&header.UpdatedBy,
		&header.Version,
	)
	if err != nil {
		return postgresSubcontractMaterialTransferHeader{}, err
	}

	return header, nil
}

func listPostgresSubcontractMaterialTransferLines(
	ctx context.Context,
	queryer postgresSubcontractMaterialTransferQueryer,
	persistedID string,
) ([]productiondomain.SubcontractMaterialTransferLine, error) {
	rows, err := queryer.QueryContext(ctx, selectSubcontractMaterialTransferLinesSQL, persistedID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	lines := make([]productiondomain.SubcontractMaterialTransferLine, 0)
	for rows.Next() {
		line, err := scanPostgresSubcontractMaterialTransferLine(rows)
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

func scanPostgresSubcontractMaterialTransferLine(
	row interface{ Scan(dest ...any) error },
) (productiondomain.SubcontractMaterialTransferLine, error) {
	var (
		id                  string
		lineNo              int
		orderMaterialLineID string
		itemID              string
		skuCode             string
		itemName            string
		issueQtyText        string
		uomCode             string
		baseIssueQtyText    string
		baseUOMCode         string
		conversionText      string
		batchID             string
		batchNo             string
		lotNo               string
		sourceBinID         string
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
		&issueQtyText,
		&uomCode,
		&baseIssueQtyText,
		&baseUOMCode,
		&conversionText,
		&batchID,
		&batchNo,
		&lotNo,
		&sourceBinID,
		&lotTraceRequired,
		&note,
	); err != nil {
		return productiondomain.SubcontractMaterialTransferLine{}, err
	}
	issueQty, err := decimal.ParseQuantity(issueQtyText)
	if err != nil {
		return productiondomain.SubcontractMaterialTransferLine{}, err
	}
	baseIssueQty, err := decimal.ParseQuantity(baseIssueQtyText)
	if err != nil {
		return productiondomain.SubcontractMaterialTransferLine{}, err
	}
	conversion, err := decimal.ParseQuantity(conversionText)
	if err != nil {
		return productiondomain.SubcontractMaterialTransferLine{}, err
	}

	return productiondomain.NewSubcontractMaterialTransferLine(productiondomain.NewSubcontractMaterialTransferLineInput{
		ID:                  id,
		LineNo:              lineNo,
		OrderMaterialLineID: orderMaterialLineID,
		ItemID:              itemID,
		SKUCode:             skuCode,
		ItemName:            itemName,
		IssueQty:            issueQty,
		UOMCode:             uomCode,
		BaseIssueQty:        baseIssueQty,
		BaseUOMCode:         baseUOMCode,
		ConversionFactor:    conversion,
		BatchID:             batchID,
		BatchNo:             batchNo,
		LotNo:               lotNo,
		SourceBinID:         sourceBinID,
		LotTraceRequired:    lotTraceRequired,
		Note:                note,
	})
}

func listPostgresSubcontractMaterialTransferEvidence(
	ctx context.Context,
	queryer postgresSubcontractMaterialTransferQueryer,
	persistedID string,
) ([]productiondomain.SubcontractMaterialTransferEvidence, error) {
	rows, err := queryer.QueryContext(ctx, selectSubcontractMaterialTransferEvidenceSQL, persistedID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	evidence := make([]productiondomain.SubcontractMaterialTransferEvidence, 0)
	for rows.Next() {
		item, err := scanPostgresSubcontractMaterialTransferEvidence(rows)
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

func scanPostgresSubcontractMaterialTransferEvidence(
	row interface{ Scan(dest ...any) error },
) (productiondomain.SubcontractMaterialTransferEvidence, error) {
	var (
		id           string
		evidenceType string
		fileName     string
		objectKey    string
		externalURL  string
		note         string
	)
	if err := row.Scan(&id, &evidenceType, &fileName, &objectKey, &externalURL, &note); err != nil {
		return productiondomain.SubcontractMaterialTransferEvidence{}, err
	}

	return productiondomain.NewSubcontractMaterialTransferEvidence(productiondomain.NewSubcontractMaterialTransferEvidenceInput{
		ID:           id,
		EvidenceType: evidenceType,
		FileName:     fileName,
		ObjectKey:    objectKey,
		ExternalURL:  externalURL,
		Note:         note,
	})
}

func buildPostgresSubcontractMaterialTransfer(
	header postgresSubcontractMaterialTransferHeader,
	lines []productiondomain.SubcontractMaterialTransferLine,
	evidence []productiondomain.SubcontractMaterialTransferEvidence,
) (productiondomain.SubcontractMaterialTransfer, error) {
	lineInputs := make([]productiondomain.NewSubcontractMaterialTransferLineInput, 0, len(lines))
	for _, line := range lines {
		lineInputs = append(lineInputs, productiondomain.NewSubcontractMaterialTransferLineInput{
			ID:                  line.ID,
			LineNo:              line.LineNo,
			OrderMaterialLineID: line.OrderMaterialLineID,
			ItemID:              line.ItemID,
			SKUCode:             line.SKUCode,
			ItemName:            line.ItemName,
			IssueQty:            line.IssueQty,
			UOMCode:             line.UOMCode.String(),
			BaseIssueQty:        line.BaseIssueQty,
			BaseUOMCode:         line.BaseUOMCode.String(),
			ConversionFactor:    line.ConversionFactor,
			BatchID:             line.BatchID,
			BatchNo:             line.BatchNo,
			LotNo:               line.LotNo,
			SourceBinID:         line.SourceBinID,
			LotTraceRequired:    line.LotTraceRequired,
			Note:                line.Note,
		})
	}
	evidenceInputs := make([]productiondomain.NewSubcontractMaterialTransferEvidenceInput, 0, len(evidence))
	for _, item := range evidence {
		evidenceInputs = append(evidenceInputs, productiondomain.NewSubcontractMaterialTransferEvidenceInput{
			ID:           item.ID,
			EvidenceType: item.EvidenceType,
			FileName:     item.FileName,
			ObjectKey:    item.ObjectKey,
			ExternalURL:  item.ExternalURL,
			Note:         item.Note,
		})
	}

	transfer, err := productiondomain.NewSubcontractMaterialTransfer(productiondomain.NewSubcontractMaterialTransferInput{
		ID:                  header.ID,
		OrgID:               header.OrgID,
		TransferNo:          header.TransferNo,
		SubcontractOrderID:  header.SubcontractOrderID,
		SubcontractOrderNo:  header.SubcontractOrderNo,
		FactoryID:           header.FactoryID,
		FactoryCode:         header.FactoryCode,
		FactoryName:         header.FactoryName,
		SourceWarehouseID:   header.SourceWarehouseID,
		SourceWarehouseCode: header.SourceWarehouseCode,
		Status:              productiondomain.SubcontractMaterialTransferStatus(header.Status),
		Lines:               lineInputs,
		Evidence:            evidenceInputs,
		HandoverBy:          header.HandoverBy,
		HandoverAt:          header.HandoverAt,
		ReceivedBy:          header.ReceivedBy,
		ReceiverContact:     header.ReceiverContact,
		VehicleNo:           header.VehicleNo,
		Note:                header.Note,
		CreatedAt:           header.CreatedAt,
		CreatedBy:           header.CreatedBy,
		UpdatedAt:           header.UpdatedAt,
		UpdatedBy:           header.UpdatedBy,
	})
	if err != nil {
		return productiondomain.SubcontractMaterialTransfer{}, err
	}
	transfer.Version = header.Version
	if err := transfer.Validate(); err != nil {
		return productiondomain.SubcontractMaterialTransfer{}, err
	}

	return transfer, nil
}

func upsertPostgresSubcontractMaterialTransfer(
	ctx context.Context,
	queryer postgresSubcontractMaterialTransferQueryer,
	orgID string,
	transfer productiondomain.SubcontractMaterialTransfer,
) (string, error) {
	var persistedID string
	err := queryer.QueryRowContext(
		ctx,
		upsertSubcontractMaterialTransferSQL,
		orgID,
		transfer.ID,
		transfer.OrgID,
		transfer.TransferNo,
		transfer.SubcontractOrderID,
		transfer.SubcontractOrderNo,
		transfer.FactoryID,
		nullablePostgresSubcontractMaterialTransferText(transfer.FactoryCode),
		transfer.FactoryName,
		transfer.SourceWarehouseID,
		nullablePostgresSubcontractMaterialTransferText(transfer.SourceWarehouseCode),
		string(transfer.Status),
		transfer.HandoverBy,
		transfer.HandoverAt.UTC(),
		transfer.ReceivedBy,
		nullablePostgresSubcontractMaterialTransferText(transfer.ReceiverContact),
		nullablePostgresSubcontractMaterialTransferText(transfer.VehicleNo),
		nullablePostgresSubcontractMaterialTransferText(transfer.Note),
		transfer.CreatedAt.UTC(),
		transfer.CreatedBy,
		transfer.UpdatedAt.UTC(),
		transfer.UpdatedBy,
		transfer.Version,
	).Scan(&persistedID)
	if err != nil {
		return "", fmt.Errorf("upsert subcontract material transfer %q: %w", transfer.ID, err)
	}

	return persistedID, nil
}

func replacePostgresSubcontractMaterialTransferLines(
	ctx context.Context,
	queryer postgresSubcontractMaterialTransferQueryer,
	orgID string,
	persistedID string,
	transfer productiondomain.SubcontractMaterialTransfer,
) error {
	if _, err := queryer.ExecContext(ctx, deleteSubcontractMaterialTransferLinesSQL, persistedID); err != nil {
		return fmt.Errorf("delete subcontract material transfer lines: %w", err)
	}
	for _, line := range transfer.Lines {
		if _, err := queryer.ExecContext(
			ctx,
			insertSubcontractMaterialTransferLineSQL,
			line.ID,
			orgID,
			persistedID,
			line.LineNo,
			line.OrderMaterialLineID,
			line.ItemID,
			line.SKUCode,
			line.ItemName,
			line.IssueQty.String(),
			line.UOMCode.String(),
			line.BaseIssueQty.String(),
			line.BaseUOMCode.String(),
			line.ConversionFactor.String(),
			nullablePostgresSubcontractMaterialTransferText(line.BatchID),
			nullablePostgresSubcontractMaterialTransferText(line.BatchNo),
			nullablePostgresSubcontractMaterialTransferText(line.LotNo),
			nullablePostgresSubcontractMaterialTransferText(line.SourceBinID),
			line.LotTraceRequired,
			nullablePostgresSubcontractMaterialTransferText(line.Note),
		); err != nil {
			return fmt.Errorf("insert subcontract material transfer line %q: %w", line.ID, err)
		}
	}

	return nil
}

func replacePostgresSubcontractMaterialTransferEvidence(
	ctx context.Context,
	queryer postgresSubcontractMaterialTransferQueryer,
	orgID string,
	persistedID string,
	transfer productiondomain.SubcontractMaterialTransfer,
) error {
	if _, err := queryer.ExecContext(ctx, deleteSubcontractMaterialTransferEvidenceSQL, persistedID); err != nil {
		return fmt.Errorf("delete subcontract material transfer evidence: %w", err)
	}
	for _, evidence := range transfer.Evidence {
		if _, err := queryer.ExecContext(
			ctx,
			insertSubcontractMaterialTransferEvidenceSQL,
			evidence.ID,
			orgID,
			persistedID,
			evidence.EvidenceType,
			nullablePostgresSubcontractMaterialTransferText(evidence.FileName),
			nullablePostgresSubcontractMaterialTransferText(evidence.ObjectKey),
			nullablePostgresSubcontractMaterialTransferText(evidence.ExternalURL),
			nullablePostgresSubcontractMaterialTransferText(evidence.Note),
		); err != nil {
			return fmt.Errorf("insert subcontract material transfer evidence %q: %w", evidence.ID, err)
		}
	}

	return nil
}

func nullablePostgresSubcontractMaterialTransferText(value string) any {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil
	}

	return value
}

func isPostgresSubcontractMaterialTransferUUIDText(value string) bool {
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
			if !isPostgresSubcontractMaterialTransferHexText(char) {
				return false
			}
		}
	}

	return true
}

func isPostgresSubcontractMaterialTransferHexText(char rune) bool {
	return (char >= '0' && char <= '9') ||
		(char >= 'a' && char <= 'f') ||
		(char >= 'A' && char <= 'F')
}

var _ SubcontractMaterialTransferStore = PostgresSubcontractMaterialTransferStore{}
