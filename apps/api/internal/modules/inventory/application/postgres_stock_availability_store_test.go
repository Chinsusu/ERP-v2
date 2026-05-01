package application

import (
	"database/sql"
	"strings"
	"testing"
	"time"

	"github.com/Chinsusu/ERP-v2/apps/api/internal/modules/inventory/domain"
)

func TestBuildStockAvailabilityQueryAppliesFilters(t *testing.T) {
	query, args := buildStockAvailabilityQuery(domain.AvailableStockFilter{
		WarehouseID: "warehouse-1",
		LocationID:  "bin-1",
		ItemID:      "item-1",
		SKU:         "serum-30ml",
		BatchID:     "batch-1",
	})

	for _, want := range []string{
		"balance.warehouse_id::text = $1",
		"balance.bin_id::text = $2",
		"balance.item_id::text = $3",
		"upper(item.sku) = $4",
		"balance.batch_id::text = $5",
		"ORDER BY warehouse.code",
	} {
		if !strings.Contains(query, want) {
			t.Fatalf("query missing %q:\n%s", want, query)
		}
	}
	if got, want := len(args), 5; got != want {
		t.Fatalf("len(args) = %d, want %d", got, want)
	}
	if args[3] != "SERUM-30ML" {
		t.Fatalf("sku arg = %v, want SERUM-30ML", args[3])
	}
}

func TestBuildStockAvailabilityQueryWithoutFiltersHasNoWhereClause(t *testing.T) {
	query, args := buildStockAvailabilityQuery(domain.AvailableStockFilter{})

	if strings.Contains(query, "\nWHERE ") {
		t.Fatalf("query has WHERE clause without filters:\n%s", query)
	}
	if len(args) != 0 {
		t.Fatalf("args = %v, want empty", args)
	}
}

func TestScanPostgresStockBalanceMapsRow(t *testing.T) {
	expiry := time.Date(2027, 4, 1, 0, 0, 0, 0, time.UTC)
	row, err := scanPostgresStockBalance(fakeStockBalanceScanner{values: []any{
		"00000000-0000-4000-8000-000000000801",
		"warehouse_main",
		"00000000-0000-4000-8000-000000001001",
		"A-01",
		"00000000-0000-4000-8000-000000001101",
		"fg-lip-001",
		"00000000-0000-4000-8000-000000001301",
		"LOT-S11",
		"pass",
		"active",
		sql.NullTime{Time: expiry, Valid: true},
		"pcs",
		"available",
		"21.000000",
		"3.000000",
	}})
	if err != nil {
		t.Fatalf("scanPostgresStockBalance() error = %v", err)
	}

	if row.WarehouseCode != "warehouse_main" ||
		row.LocationCode != "A-01" ||
		row.SKU != "FG-LIP-001" ||
		row.BatchNo != "LOT-S11" ||
		row.BatchQCStatus != domain.QCStatusPass ||
		row.BatchStatus != domain.BatchStatusActive ||
		row.BaseUOMCode.String() != "PCS" ||
		row.StockStatus != domain.StockStatusAvailable ||
		row.QtyOnHand.String() != "21.000000" ||
		row.QtyReserved.String() != "3.000000" ||
		!row.BatchExpiry.Equal(expiry) {
		t.Fatalf("row = %+v, want mapped stock balance snapshot", row)
	}
}

func TestPostgresStockAvailabilityStoreRequiresDatabase(t *testing.T) {
	store := NewPostgresStockAvailabilityStore(nil)

	if _, err := store.ListBalances(nil, domain.AvailableStockFilter{}); err == nil {
		t.Fatal("ListBalances() error = nil, want database required error")
	}
}

type fakeStockBalanceScanner struct {
	values []any
}

func (s fakeStockBalanceScanner) Scan(dest ...any) error {
	for i, target := range dest {
		switch typed := target.(type) {
		case *string:
			*typed = s.values[i].(string)
		case *sql.NullTime:
			*typed = s.values[i].(sql.NullTime)
		default:
			panic("unsupported scan destination")
		}
	}

	return nil
}
