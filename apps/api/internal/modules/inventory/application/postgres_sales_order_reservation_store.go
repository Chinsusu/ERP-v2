package application

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/Chinsusu/ERP-v2/apps/api/internal/modules/inventory/domain"
	salesapp "github.com/Chinsusu/ERP-v2/apps/api/internal/modules/sales/application"
	"github.com/Chinsusu/ERP-v2/apps/api/internal/shared/audit"
	"github.com/Chinsusu/ERP-v2/apps/api/internal/shared/decimal"
)

type PostgresSalesOrderReservationStoreConfig struct {
	DefaultOrgID string
	BaselineRows []domain.StockBalanceSnapshot
}

type PostgresSalesOrderReservationStore struct {
	db           *sql.DB
	defaultOrgID string
	baselineRows []domain.StockBalanceSnapshot
}

func NewPostgresSalesOrderReservationStore(
	db *sql.DB,
	cfg PostgresSalesOrderReservationStoreConfig,
) PostgresSalesOrderReservationStore {
	baselineRows := cloneStockBalanceRows(cfg.BaselineRows)
	if len(baselineRows) == 0 {
		baselineRows = prototypeSalesOrderReservationRows()
	}

	return PostgresSalesOrderReservationStore{
		db:           db,
		defaultOrgID: strings.TrimSpace(cfg.DefaultOrgID),
		baselineRows: baselineRows,
	}
}

const selectReservationOrgIDSQL = `
SELECT id::text
FROM core.organizations
WHERE code = $1
LIMIT 1`

const selectActiveStockReservationsSQL = `
SELECT
  COALESCE(reservation.reservation_ref, reservation.id::text),
  COALESCE(reservation.org_ref, reservation.org_id::text),
  reservation.reservation_no,
  COALESCE(reservation.sales_order_ref, reservation.sales_order_id::text),
  COALESCE(reservation.sales_order_line_ref, reservation.sales_order_line_id::text),
  COALESCE(reservation.item_ref, reservation.item_id::text),
  COALESCE(reservation.sku_code, item.sku, ''),
  COALESCE(reservation.batch_ref, reservation.batch_id::text),
  COALESCE(reservation.batch_no, batch.batch_no, ''),
  COALESCE(reservation.warehouse_ref, reservation.warehouse_id::text),
  COALESCE(reservation.warehouse_code, warehouse.code, ''),
  COALESCE(reservation.bin_ref, reservation.bin_id::text),
  COALESCE(reservation.bin_code, bin.code, ''),
  reservation.stock_status,
  reservation.reserved_qty::text,
  reservation.base_uom_code,
  reservation.status,
  reservation.created_at,
  COALESCE(reservation.created_by_ref, reservation.created_by::text, ''),
  reservation.released_at,
  COALESCE(reservation.released_by_ref, reservation.released_by::text, ''),
  reservation.consumed_at,
  COALESCE(reservation.consumed_by_ref, reservation.consumed_by::text, ''),
  reservation.updated_at
FROM inventory.stock_reservations AS reservation
LEFT JOIN mdm.items AS item ON item.id = reservation.item_id
LEFT JOIN inventory.batches AS batch ON batch.id = reservation.batch_id
LEFT JOIN mdm.warehouses AS warehouse ON warehouse.id = reservation.warehouse_id
LEFT JOIN mdm.warehouse_bins AS bin ON bin.id = reservation.bin_id
WHERE reservation.org_id = $1
  AND reservation.status = 'active'`

const selectActiveStockReservationsByOrderSQL = selectActiveStockReservationsSQL + `
  AND (
    reservation.sales_order_ref = $2
    OR reservation.sales_order_id::text = $2
  )`

const insertStockReservationSQL = `
INSERT INTO inventory.stock_reservations (
  id,
  org_id,
  reservation_no,
  item_id,
  batch_id,
  warehouse_id,
  reserved_qty,
  source_doc_type,
  source_doc_id,
  status,
  created_at,
  created_by,
  released_at,
  released_by,
  sales_order_id,
  sales_order_line_id,
  source_doc_line_id,
  bin_id,
  base_uom_code,
  stock_status,
  consumed_at,
  consumed_by,
  updated_at,
  reservation_ref,
  org_ref,
  sales_order_ref,
  sales_order_line_ref,
  source_doc_ref,
  source_doc_line_ref,
  item_ref,
  sku_code,
  batch_ref,
  batch_no,
  warehouse_ref,
  warehouse_code,
  bin_ref,
  bin_code,
  created_by_ref,
  released_by_ref,
  consumed_by_ref
) VALUES (
  COALESCE($1::uuid, gen_random_uuid()),
  $2::uuid,
  $3,
  $4::uuid,
  $5::uuid,
  $6::uuid,
  $7,
  'sales_order',
  $8::uuid,
  $9,
  $10,
  $11::uuid,
  $12,
  $13::uuid,
  $14::uuid,
  $15::uuid,
  $16::uuid,
  $17::uuid,
  $18,
  $19,
  $20,
  $21::uuid,
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
  $37,
  $38,
  $39
)`

const releaseStockReservationSQL = `
UPDATE inventory.stock_reservations
SET
  status = 'released',
  released_at = $3,
  released_by = $4::uuid,
  released_by_ref = $5,
  updated_at = $6
WHERE org_id = $1::uuid
  AND status = 'active'
  AND (
    reservation_ref = $2
    OR id::text = $2
  )`

const insertStockReservationAuditSQL = `
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

func (s PostgresSalesOrderReservationStore) ReserveSalesOrder(
	ctx context.Context,
	input salesapp.SalesOrderStockReservationInput,
) (salesapp.SalesOrderStockReservationResult, error) {
	if s.db == nil {
		return salesapp.SalesOrderStockReservationResult{}, errors.New("database connection is required")
	}
	if strings.TrimSpace(input.ActorID) == "" {
		return salesapp.SalesOrderStockReservationResult{}, domain.ErrStockReservationActorRequired
	}
	if len(input.Lines) == 0 {
		return salesapp.SalesOrderStockReservationResult{}, domain.ErrStockReservationRequiredField
	}

	orgID, err := s.resolveOrgID(ctx, input.OrgID)
	if err != nil {
		return salesapp.SalesOrderStockReservationResult{}, err
	}
	tx, err := s.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
	if err != nil {
		return salesapp.SalesOrderStockReservationResult{}, fmt.Errorf("begin sales reservation transaction: %w", err)
	}

	committed := false
	defer func() {
		if !committed {
			_ = tx.Rollback()
		}
	}()

	activeReservations, err := s.listActiveReservations(ctx, tx, orgID)
	if err != nil {
		return salesapp.SalesOrderStockReservationResult{}, err
	}
	rows, err := s.rowsWithActiveReservations(activeReservations)
	if err != nil {
		return salesapp.SalesOrderStockReservationResult{}, err
	}

	reservations := make([]domain.StockReservation, 0, len(input.Lines))
	resultLines := make([]salesapp.SalesOrderReservedLine, 0, len(input.Lines))
	for _, line := range input.Lines {
		allocated, err := reserveSalesOrderLine(rows, input, line)
		if err != nil {
			return salesapp.SalesOrderStockReservationResult{}, err
		}
		reservations = append(reservations, allocated.reservation)
		resultLines = append(resultLines, allocated.line)
	}
	for _, reservation := range reservations {
		if err := insertStockReservation(ctx, tx, orgID, reservation); err != nil {
			return salesapp.SalesOrderStockReservationResult{}, err
		}
	}
	for _, reservation := range reservations {
		if err := insertStockReservationAudit(
			ctx,
			tx,
			orgID,
			input.ActorID,
			input.RequestID,
			stockReservationReservedAction,
			reservation,
			nil,
			stockReservationAuditData(reservation),
			stockReservationAuditMetadata(input.OrderNo, input.Reason, "sales order reservation postgres store"),
			reservation.ReservedAt,
		); err != nil {
			return salesapp.SalesOrderStockReservationResult{}, err
		}
	}
	if err := tx.Commit(); err != nil {
		return salesapp.SalesOrderStockReservationResult{}, fmt.Errorf("commit sales reservation transaction: %w", err)
	}
	committed = true

	return salesapp.SalesOrderStockReservationResult{Lines: resultLines}, nil
}

func (s PostgresSalesOrderReservationStore) ReleaseSalesOrder(
	ctx context.Context,
	input salesapp.SalesOrderStockReleaseInput,
) (salesapp.SalesOrderStockReleaseResult, error) {
	if s.db == nil {
		return salesapp.SalesOrderStockReleaseResult{}, errors.New("database connection is required")
	}
	if strings.TrimSpace(input.ActorID) == "" {
		return salesapp.SalesOrderStockReleaseResult{}, domain.ErrStockReservationActorRequired
	}

	orgID, err := s.resolveOrgID(ctx, input.OrgID)
	if err != nil {
		return salesapp.SalesOrderStockReleaseResult{}, err
	}
	tx, err := s.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
	if err != nil {
		return salesapp.SalesOrderStockReleaseResult{}, fmt.Errorf("begin sales reservation release transaction: %w", err)
	}

	committed := false
	defer func() {
		if !committed {
			_ = tx.Rollback()
		}
	}()

	activeReservations, err := s.listActiveReservationsBySalesOrder(ctx, tx, orgID, input.SalesOrderID)
	if err != nil {
		return salesapp.SalesOrderStockReleaseResult{}, err
	}
	changes := make([]stockReservationAuditChange, 0, len(activeReservations))
	for _, reservation := range activeReservations {
		released, err := reservation.Release(input.ActorID, releasedAt(input.ReleasedAt))
		if err != nil {
			return salesapp.SalesOrderStockReleaseResult{}, err
		}
		if err := releaseStockReservation(ctx, tx, orgID, released); err != nil {
			return salesapp.SalesOrderStockReleaseResult{}, err
		}
		changes = append(changes, stockReservationAuditChange{before: reservation, after: released})
	}
	for _, change := range changes {
		if err := insertStockReservationAudit(
			ctx,
			tx,
			orgID,
			input.ActorID,
			input.RequestID,
			stockReservationReleasedAction,
			change.after,
			stockReservationAuditData(change.before),
			stockReservationAuditData(change.after),
			stockReservationAuditMetadata(input.OrderNo, input.Reason, "sales order reservation postgres store"),
			change.after.ReleasedAt,
		); err != nil {
			return salesapp.SalesOrderStockReleaseResult{}, err
		}
	}
	if err := tx.Commit(); err != nil {
		return salesapp.SalesOrderStockReleaseResult{}, fmt.Errorf("commit sales reservation release transaction: %w", err)
	}
	committed = true

	return salesapp.SalesOrderStockReleaseResult{ReleasedReservationCount: len(changes)}, nil
}

func (s PostgresSalesOrderReservationStore) resolveOrgID(ctx context.Context, orgRef string) (string, error) {
	orgRef = strings.TrimSpace(orgRef)
	if isUUIDText(orgRef) {
		return orgRef, nil
	}
	if orgRef != "" {
		var orgID string
		err := s.db.QueryRowContext(ctx, selectReservationOrgIDSQL, orgRef).Scan(&orgID)
		if err == nil && isUUIDText(orgID) {
			return orgID, nil
		}
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			return "", fmt.Errorf("resolve reservation org %q: %w", orgRef, err)
		}
	}
	if isUUIDText(s.defaultOrgID) {
		return s.defaultOrgID, nil
	}

	return "", fmt.Errorf("reservation org %q cannot be resolved", orgRef)
}

func (s PostgresSalesOrderReservationStore) listActiveReservations(
	ctx context.Context,
	tx *sql.Tx,
	orgID string,
) ([]domain.StockReservation, error) {
	rows, err := tx.QueryContext(ctx, selectActiveStockReservationsSQL, orgID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanStockReservations(rows)
}

func (s PostgresSalesOrderReservationStore) listActiveReservationsBySalesOrder(
	ctx context.Context,
	tx *sql.Tx,
	orgID string,
	salesOrderID string,
) ([]domain.StockReservation, error) {
	rows, err := tx.QueryContext(ctx, selectActiveStockReservationsByOrderSQL, orgID, strings.TrimSpace(salesOrderID))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanStockReservations(rows)
}

func (s PostgresSalesOrderReservationStore) rowsWithActiveReservations(
	reservations []domain.StockReservation,
) ([]domain.StockBalanceSnapshot, error) {
	rows := cloneStockBalanceRows(s.baselineRows)
	for _, reservation := range reservations {
		rowIndex := findReservationStockBalanceRow(rows, reservation)
		if rowIndex < 0 {
			continue
		}
		updatedReservedQty, err := decimal.AddQuantity(rows[rowIndex].QtyReserved, reservation.ReservedQty)
		if err != nil {
			return nil, err
		}
		rows[rowIndex].QtyReserved = updatedReservedQty
	}

	return rows, nil
}

func scanStockReservations(rows *sql.Rows) ([]domain.StockReservation, error) {
	reservations := make([]domain.StockReservation, 0)
	for rows.Next() {
		reservation, err := scanStockReservation(rows)
		if err != nil {
			return nil, err
		}
		reservations = append(reservations, reservation)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return reservations, nil
}

func scanStockReservation(rows interface{ Scan(dest ...any) error }) (domain.StockReservation, error) {
	var (
		id               string
		orgID            string
		reservationNo    string
		salesOrderID     string
		salesOrderLineID string
		itemID           string
		skuCode          string
		batchID          string
		batchNo          string
		warehouseID      string
		warehouseCode    string
		binID            string
		binCode          string
		stockStatus      string
		reservedQtyText  string
		baseUOMCode      string
		status           string
		reservedAt       time.Time
		reservedBy       string
		releasedAt       sql.NullTime
		releasedBy       string
		consumedAt       sql.NullTime
		consumedBy       string
		updatedAt        time.Time
	)
	if err := rows.Scan(
		&id,
		&orgID,
		&reservationNo,
		&salesOrderID,
		&salesOrderLineID,
		&itemID,
		&skuCode,
		&batchID,
		&batchNo,
		&warehouseID,
		&warehouseCode,
		&binID,
		&binCode,
		&stockStatus,
		&reservedQtyText,
		&baseUOMCode,
		&status,
		&reservedAt,
		&reservedBy,
		&releasedAt,
		&releasedBy,
		&consumedAt,
		&consumedBy,
		&updatedAt,
	); err != nil {
		return domain.StockReservation{}, err
	}

	reservedQty, err := decimal.ParseQuantity(reservedQtyText)
	if err != nil {
		return domain.StockReservation{}, err
	}
	input := domain.NewStockReservationInput{
		ID:               id,
		OrgID:            orgID,
		ReservationNo:    reservationNo,
		SalesOrderID:     salesOrderID,
		SalesOrderLineID: salesOrderLineID,
		ItemID:           itemID,
		SKUCode:          skuCode,
		BatchID:          batchID,
		BatchNo:          batchNo,
		WarehouseID:      warehouseID,
		WarehouseCode:    warehouseCode,
		BinID:            binID,
		BinCode:          binCode,
		StockStatus:      domain.StockStatus(stockStatus),
		ReservedQty:      reservedQty,
		BaseUOMCode:      baseUOMCode,
		Status:           domain.ReservationStatus(status),
		ReservedAt:       reservedAt,
		ReservedBy:       reservedBy,
		ReleasedBy:       releasedBy,
		ConsumedBy:       consumedBy,
		CreatedAt:        reservedAt,
		UpdatedAt:        updatedAt,
	}
	if releasedAt.Valid {
		input.ReleasedAt = releasedAt.Time
	}
	if consumedAt.Valid {
		input.ConsumedAt = consumedAt.Time
	}

	return domain.NewStockReservation(input)
}

func insertStockReservation(ctx context.Context, tx *sql.Tx, orgID string, reservation domain.StockReservation) error {
	_, err := tx.ExecContext(
		ctx,
		insertStockReservationSQL,
		nullableUUID(reservation.ID),
		orgID,
		reservation.ReservationNo,
		nullableUUID(reservation.ItemID),
		nullableUUID(reservation.BatchID),
		nullableUUID(reservation.WarehouseID),
		reservation.ReservedQty.String(),
		nullableUUID(reservation.SalesOrderID),
		string(reservation.Status),
		reservation.ReservedAt.UTC(),
		nullableUUID(reservation.ReservedBy),
		nullableTime(reservation.ReleasedAt),
		nullableUUID(reservation.ReleasedBy),
		nullableUUID(reservation.SalesOrderID),
		nullableUUID(reservation.SalesOrderLineID),
		nullableUUID(reservation.SalesOrderLineID),
		nullableUUID(reservation.BinID),
		reservation.BaseUOMCode.String(),
		string(reservation.StockStatus),
		nullableTime(reservation.ConsumedAt),
		nullableUUID(reservation.ConsumedBy),
		reservation.UpdatedAt.UTC(),
		nullableText(reservation.ID),
		nullableText(reservation.OrgID),
		nullableText(reservation.SalesOrderID),
		nullableText(reservation.SalesOrderLineID),
		nullableText(reservation.SalesOrderID),
		nullableText(reservation.SalesOrderLineID),
		nullableText(reservation.ItemID),
		nullableText(reservation.SKUCode),
		nullableText(reservation.BatchID),
		nullableText(reservation.BatchNo),
		nullableText(reservation.WarehouseID),
		nullableText(reservation.WarehouseCode),
		nullableText(reservation.BinID),
		nullableText(reservation.BinCode),
		nullableText(reservation.ReservedBy),
		nullableText(reservation.ReleasedBy),
		nullableText(reservation.ConsumedBy),
	)
	if err != nil {
		return fmt.Errorf("insert stock reservation: %w", err)
	}

	return nil
}

func releaseStockReservation(ctx context.Context, tx *sql.Tx, orgID string, reservation domain.StockReservation) error {
	result, err := tx.ExecContext(
		ctx,
		releaseStockReservationSQL,
		orgID,
		reservation.ID,
		reservation.ReleasedAt.UTC(),
		nullableUUID(reservation.ReleasedBy),
		nullableText(reservation.ReleasedBy),
		reservation.UpdatedAt.UTC(),
	)
	if err != nil {
		return fmt.Errorf("release stock reservation: %w", err)
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("release stock reservation affected rows: %w", err)
	}
	if affected != 1 {
		return fmt.Errorf("release stock reservation affected %d rows, want 1", affected)
	}

	return nil
}

func insertStockReservationAudit(
	ctx context.Context,
	tx *sql.Tx,
	orgID string,
	actorID string,
	requestID string,
	action string,
	reservation domain.StockReservation,
	beforeData map[string]any,
	afterData map[string]any,
	metadata map[string]any,
	createdAt time.Time,
) error {
	log, err := audit.NewLog(audit.NewLogInput{
		ID:         newStockReservationAuditID(action, reservation.ID, createdAt),
		OrgID:      reservation.OrgID,
		ActorID:    strings.TrimSpace(actorID),
		Action:     action,
		EntityType: stockReservationEntityType,
		EntityID:   reservation.ID,
		RequestID:  strings.TrimSpace(requestID),
		BeforeData: beforeData,
		AfterData:  afterData,
		Metadata:   metadata,
		CreatedAt:  createdAt,
	})
	if err != nil {
		return err
	}
	beforeJSON, err := stockReservationJSONMap(log.BeforeData)
	if err != nil {
		return fmt.Errorf("encode stock reservation audit before_data: %w", err)
	}
	afterJSON, err := stockReservationJSONMap(log.AfterData)
	if err != nil {
		return fmt.Errorf("encode stock reservation audit after_data: %w", err)
	}
	metadataJSON, err := requiredStockReservationJSONMap(log.Metadata)
	if err != nil {
		return fmt.Errorf("encode stock reservation audit metadata: %w", err)
	}

	_, err = tx.ExecContext(
		ctx,
		insertStockReservationAuditSQL,
		nullableUUID(log.ID),
		orgID,
		nullableUUID(log.ActorID),
		log.Action,
		log.EntityType,
		nullableUUID(log.EntityID),
		nullableText(log.RequestID),
		beforeJSON,
		afterJSON,
		metadataJSON,
		log.CreatedAt.UTC(),
		nullableText(log.ID),
		nullableText(log.OrgID),
		nullableText(log.ActorID),
		nullableText(log.EntityID),
	)
	if err != nil {
		return fmt.Errorf("insert stock reservation audit: %w", err)
	}

	return nil
}

func stockReservationJSONMap(value map[string]any) (any, error) {
	if value == nil {
		return nil, nil
	}

	return requiredStockReservationJSONMap(value)
}

func requiredStockReservationJSONMap(value map[string]any) (string, error) {
	if value == nil {
		value = map[string]any{}
	}
	data, err := json.Marshal(value)
	if err != nil {
		return "", err
	}

	return string(data), nil
}

func nullableText(value string) any {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil
	}

	return value
}

func nullableTime(value time.Time) any {
	if value.IsZero() {
		return nil
	}

	return value.UTC()
}

var _ salesapp.SalesOrderStockReserver = PostgresSalesOrderReservationStore{}
