package application

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/Chinsusu/ERP-v2/apps/api/internal/modules/inventory/domain"
	"github.com/Chinsusu/ERP-v2/apps/api/internal/shared/decimal"
)

type PostgresStockAvailabilityStore struct {
	db *sql.DB
}

func NewPostgresStockAvailabilityStore(db *sql.DB) PostgresStockAvailabilityStore {
	return PostgresStockAvailabilityStore{db: db}
}

const selectStockAvailabilityBaseSQL = `
SELECT
  balance.warehouse_id::text,
  warehouse.code,
  COALESCE(balance.bin_id::text, ''),
  COALESCE(bin.code, ''),
  balance.item_id::text,
  item.sku,
  COALESCE(balance.batch_id::text, ''),
  COALESCE(batch.batch_no, ''),
  COALESCE(batch.qc_status, ''),
  COALESCE(batch.status, ''),
  batch.expiry_date,
  balance.base_uom_code,
  balance.stock_status,
  balance.qty_on_hand::text,
  balance.qty_reserved::text
FROM inventory.stock_balances AS balance
JOIN mdm.items AS item ON item.id = balance.item_id
JOIN mdm.warehouses AS warehouse ON warehouse.id = balance.warehouse_id
LEFT JOIN mdm.warehouse_bins AS bin ON bin.id = balance.bin_id
LEFT JOIN inventory.batches AS batch ON batch.id = balance.batch_id`

const selectStockAvailabilityOrderSQL = `
ORDER BY warehouse.code, COALESCE(bin.code, ''), item.sku, COALESCE(batch.batch_no, ''), balance.stock_status`

func (s PostgresStockAvailabilityStore) ListBalances(
	ctx context.Context,
	filter domain.AvailableStockFilter,
) ([]domain.StockBalanceSnapshot, error) {
	if s.db == nil {
		return nil, errors.New("database connection is required")
	}

	query, args := buildStockAvailabilityQuery(filter)
	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	snapshots := make([]domain.StockBalanceSnapshot, 0)
	for rows.Next() {
		snapshot, err := scanPostgresStockBalance(rows)
		if err != nil {
			return nil, err
		}
		snapshots = append(snapshots, snapshot)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return snapshots, nil
}

func buildStockAvailabilityQuery(filter domain.AvailableStockFilter) (string, []any) {
	clauses := make([]string, 0, 5)
	args := make([]any, 0, 5)

	addClause := func(clause string, value string) {
		trimmed := strings.TrimSpace(value)
		if trimmed == "" {
			return
		}
		args = append(args, trimmed)
		clauses = append(clauses, fmt.Sprintf(clause, len(args)))
	}

	addClause("balance.warehouse_id::text = $%d", filter.WarehouseID)
	addClause("balance.bin_id::text = $%d", filter.LocationID)
	addClause("balance.item_id::text = $%d", filter.ItemID)
	if sku := strings.ToUpper(strings.TrimSpace(filter.SKU)); sku != "" {
		args = append(args, sku)
		clauses = append(clauses, fmt.Sprintf("upper(item.sku) = $%d", len(args)))
	}
	addClause("balance.batch_id::text = $%d", filter.BatchID)

	query := selectStockAvailabilityBaseSQL
	if len(clauses) > 0 {
		query += "\nWHERE " + strings.Join(clauses, "\n  AND ")
	}
	query += selectStockAvailabilityOrderSQL

	return query, args
}

type stockBalanceRowScanner interface {
	Scan(dest ...any) error
}

func scanPostgresStockBalance(scanner stockBalanceRowScanner) (domain.StockBalanceSnapshot, error) {
	var (
		warehouseID   string
		warehouseCode string
		locationID    string
		locationCode  string
		itemID        string
		sku           string
		batchID       string
		batchNo       string
		batchQCStatus string
		batchStatus   string
		batchExpiry   sql.NullTime
		baseUOMCode   string
		stockStatus   string
		qtyOnHand     string
		qtyReserved   string
	)

	if err := scanner.Scan(
		&warehouseID,
		&warehouseCode,
		&locationID,
		&locationCode,
		&itemID,
		&sku,
		&batchID,
		&batchNo,
		&batchQCStatus,
		&batchStatus,
		&batchExpiry,
		&baseUOMCode,
		&stockStatus,
		&qtyOnHand,
		&qtyReserved,
	); err != nil {
		return domain.StockBalanceSnapshot{}, err
	}

	uomCode, err := decimal.NormalizeUOMCode(baseUOMCode)
	if err != nil {
		return domain.StockBalanceSnapshot{}, err
	}
	onHand, err := decimal.ParseQuantity(qtyOnHand)
	if err != nil {
		return domain.StockBalanceSnapshot{}, err
	}
	reserved, err := decimal.ParseQuantity(qtyReserved)
	if err != nil {
		return domain.StockBalanceSnapshot{}, err
	}

	var expiry time.Time
	if batchExpiry.Valid {
		expiry = batchExpiry.Time
	}

	return domain.StockBalanceSnapshot{
		WarehouseID:   strings.TrimSpace(warehouseID),
		WarehouseCode: strings.TrimSpace(warehouseCode),
		LocationID:    strings.TrimSpace(locationID),
		LocationCode:  strings.TrimSpace(locationCode),
		ItemID:        strings.TrimSpace(itemID),
		SKU:           strings.ToUpper(strings.TrimSpace(sku)),
		BatchID:       strings.TrimSpace(batchID),
		BatchNo:       strings.TrimSpace(batchNo),
		BatchQCStatus: domain.QCStatus(strings.TrimSpace(batchQCStatus)),
		BatchStatus:   domain.BatchStatus(strings.TrimSpace(batchStatus)),
		BatchExpiry:   expiry,
		BaseUOMCode:   uomCode,
		StockStatus:   domain.StockStatus(strings.TrimSpace(stockStatus)),
		QtyOnHand:     onHand,
		QtyReserved:   reserved,
	}, nil
}
