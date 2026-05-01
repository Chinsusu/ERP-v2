package application

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/Chinsusu/ERP-v2/apps/api/internal/modules/inventory/domain"
	salesapp "github.com/Chinsusu/ERP-v2/apps/api/internal/modules/sales/application"
	"github.com/Chinsusu/ERP-v2/apps/api/internal/shared/decimal"

	_ "github.com/jackc/pgx/v5/stdlib"
)

const testReservationOrgID = "00000000-0000-4000-8000-000000000001"

func TestPostgresSalesOrderReservationStoreRowsWithActiveReservations(t *testing.T) {
	store := NewPostgresSalesOrderReservationStore(nil, PostgresSalesOrderReservationStoreConfig{
		BaselineRows: []domain.StockBalanceSnapshot{{
			WarehouseID:   "wh-hcm-fg",
			WarehouseCode: "WH-HCM-FG",
			LocationID:    "bin-hcm-pick-a01",
			LocationCode:  "PICK-A-01",
			ItemID:        "item-serum-30ml",
			SKU:           "SERUM-30ML",
			BatchID:       "batch-serum-2604a",
			BatchNo:       "LOT-2604A",
			BaseUOMCode:   decimal.MustUOMCode("EA"),
			StockStatus:   domain.StockStatusAvailable,
			QtyOnHand:     decimal.MustQuantity("120"),
			QtyReserved:   decimal.MustQuantity("10"),
		}},
	})
	reservation := mustTestStockReservation(t, func(input *domain.NewStockReservationInput) {
		input.ReservedQty = decimal.MustQuantity("5")
	})

	rows, err := store.rowsWithActiveReservations([]domain.StockReservation{reservation})
	if err != nil {
		t.Fatalf("rowsWithActiveReservations() error = %v", err)
	}
	if got := rows[0].QtyReserved.String(); got != "15.000000" {
		t.Fatalf("QtyReserved = %s, want 15.000000", got)
	}
}

func TestScanStockReservationKeepsRuntimeReferenceFields(t *testing.T) {
	reservedAt := time.Date(2026, 5, 1, 8, 0, 0, 0, time.UTC)
	row := fakeStockReservationScanner{values: []any{
		"rsv-so-reserve-line-serum",
		"org-my-pham",
		"RSV-SO-RESERVE-01",
		"so-reserve",
		"line-serum",
		"item-serum-30ml",
		"SERUM-30ML",
		"batch-serum-2604a",
		"LOT-2604A",
		"wh-hcm-fg",
		"WH-HCM-FG",
		"bin-hcm-pick-a01",
		"PICK-A-01",
		"available",
		"7.000000",
		"EA",
		"active",
		reservedAt,
		"user-sales",
		sql.NullTime{},
		"",
		sql.NullTime{},
		"",
		reservedAt,
	}}

	reservation, err := scanStockReservation(row)
	if err != nil {
		t.Fatalf("scanStockReservation() error = %v", err)
	}
	if reservation.ID != "rsv-so-reserve-line-serum" ||
		reservation.OrgID != "org-my-pham" ||
		reservation.SalesOrderID != "so-reserve" ||
		reservation.ReservedBy != "user-sales" {
		t.Fatalf("reservation refs = %+v, want runtime text refs preserved", reservation)
	}
	if got := reservation.ReservedQty.String(); got != "7.000000" {
		t.Fatalf("reserved qty = %s, want 7.000000", got)
	}
}

func TestPostgresSalesOrderReservationStorePersistsAndReleasesRuntimeRefs(t *testing.T) {
	databaseURL := os.Getenv("ERP_TEST_DATABASE_URL")
	if databaseURL == "" {
		t.Skip("ERP_TEST_DATABASE_URL is not set")
	}

	ctx := context.Background()
	db, err := sql.Open("pgx", databaseURL)
	if err != nil {
		t.Fatalf("open database: %v", err)
	}
	defer db.Close()

	if err := seedReservationSmokeOrg(ctx, db); err != nil {
		t.Fatalf("seed org: %v", err)
	}

	now := time.Now().UTC()
	suffix := fmt.Sprintf("%d", now.UnixNano())
	orderID := "so-s10-0202-" + suffix
	lineID := "line-s10-0202-" + suffix
	reservationID := fmt.Sprintf("rsv-%s-%s", orderID, lineID)
	store := NewPostgresSalesOrderReservationStore(
		db,
		PostgresSalesOrderReservationStoreConfig{DefaultOrgID: testReservationOrgID},
	)

	reserved, err := store.ReserveSalesOrder(ctx, salesapp.SalesOrderStockReservationInput{
		OrgID:        "org-my-pham",
		SalesOrderID: orderID,
		OrderNo:      "SO-S10-0202-" + suffix,
		WarehouseID:  "wh-hcm-fg",
		ActorID:      "user-sales",
		RequestID:    "req-s10-0202-reserve-" + suffix,
		ReservedAt:   now,
		Lines: []salesapp.SalesOrderStockReservationLineInput{{
			SalesOrderLineID: lineID,
			LineNo:           1,
			ItemID:           "item-serum-30ml",
			SKUCode:          "SERUM-30ML",
			OrderedQty:       decimal.MustQuantity("2"),
			BaseOrderedQty:   decimal.MustQuantity("2"),
			BaseUOMCode:      decimal.MustUOMCode("EA"),
		}},
	})
	if err != nil {
		t.Fatalf("ReserveSalesOrder() error = %v", err)
	}
	if len(reserved.Lines) != 1 || reserved.Lines[0].ReservedQty.String() != "2.000000" {
		t.Fatalf("reserved lines = %+v, want one 2 EA line", reserved.Lines)
	}

	var status, salesOrderRef, baseUOMCode, createdByRef string
	if err := db.QueryRowContext(
		ctx,
		`SELECT status, sales_order_ref, base_uom_code, created_by_ref
FROM inventory.stock_reservations
WHERE org_id = $1 AND reservation_ref = $2`,
		testReservationOrgID,
		reservationID,
	).Scan(&status, &salesOrderRef, &baseUOMCode, &createdByRef); err != nil {
		t.Fatalf("query reserved row: %v", err)
	}
	if status != "active" || salesOrderRef != orderID || baseUOMCode != "EA" || createdByRef != "user-sales" {
		t.Fatalf("reserved row = status %q order %q uom %q actor %q", status, salesOrderRef, baseUOMCode, createdByRef)
	}

	released, err := store.ReleaseSalesOrder(ctx, salesapp.SalesOrderStockReleaseInput{
		OrgID:        "org-my-pham",
		SalesOrderID: orderID,
		OrderNo:      "SO-S10-0202-" + suffix,
		ActorID:      "user-sales",
		RequestID:    "req-s10-0202-release-" + suffix,
		ReleasedAt:   now.Add(time.Minute),
	})
	if err != nil {
		t.Fatalf("ReleaseSalesOrder() error = %v", err)
	}
	if released.ReleasedReservationCount != 1 {
		t.Fatalf("ReleasedReservationCount = %d, want 1", released.ReleasedReservationCount)
	}
	if err := db.QueryRowContext(
		ctx,
		`SELECT status, released_by_ref
FROM inventory.stock_reservations
WHERE org_id = $1 AND reservation_ref = $2`,
		testReservationOrgID,
		reservationID,
	).Scan(&status, &createdByRef); err != nil {
		t.Fatalf("query released row: %v", err)
	}
	if status != "released" || createdByRef != "user-sales" {
		t.Fatalf("released row = status %q actor %q, want released by user-sales", status, createdByRef)
	}
}

type fakeStockReservationScanner struct {
	values []any
}

func (s fakeStockReservationScanner) Scan(dest ...any) error {
	for index := range dest {
		switch target := dest[index].(type) {
		case *string:
			*target = s.values[index].(string)
		case *time.Time:
			*target = s.values[index].(time.Time)
		case *sql.NullTime:
			*target = s.values[index].(sql.NullTime)
		}
	}

	return nil
}

func mustTestStockReservation(
	t *testing.T,
	mutate func(input *domain.NewStockReservationInput),
) domain.StockReservation {
	t.Helper()

	input := domain.NewStockReservationInput{
		ID:               "rsv-so-reserve-line-serum",
		OrgID:            "org-my-pham",
		ReservationNo:    "RSV-SO-RESERVE-01",
		SalesOrderID:     "so-reserve",
		SalesOrderLineID: "line-serum",
		ItemID:           "item-serum-30ml",
		SKUCode:          "SERUM-30ML",
		BatchID:          "batch-serum-2604a",
		BatchNo:          "LOT-2604A",
		WarehouseID:      "wh-hcm-fg",
		WarehouseCode:    "WH-HCM-FG",
		BinID:            "bin-hcm-pick-a01",
		BinCode:          "PICK-A-01",
		StockStatus:      domain.StockStatusAvailable,
		ReservedQty:      decimal.MustQuantity("1"),
		BaseUOMCode:      "EA",
		ReservedAt:       time.Date(2026, 5, 1, 8, 0, 0, 0, time.UTC),
		ReservedBy:       "user-sales",
	}
	if mutate != nil {
		mutate(&input)
	}
	reservation, err := domain.NewStockReservation(input)
	if err != nil {
		t.Fatalf("NewStockReservation() error = %v", err)
	}

	return reservation
}

func seedReservationSmokeOrg(ctx context.Context, db *sql.DB) error {
	_, err := db.ExecContext(
		ctx,
		`INSERT INTO core.organizations (id, code, name, status)
VALUES ($1, 'ERP_LOCAL', 'ERP Local Demo', 'active')
ON CONFLICT (code) DO UPDATE
SET name = EXCLUDED.name,
    status = EXCLUDED.status,
    updated_at = now()`,
		testReservationOrgID,
	)

	return err
}
