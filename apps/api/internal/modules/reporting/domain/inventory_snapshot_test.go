package domain

import (
	"errors"
	"testing"
	"time"

	inventorydomain "github.com/Chinsusu/ERP-v2/apps/api/internal/modules/inventory/domain"
	"github.com/Chinsusu/ERP-v2/apps/api/internal/shared/decimal"
)

func TestNewInventorySnapshotReportBuildsRowsAndUOMTotals(t *testing.T) {
	filters := mustReportFilters(t, "2026-04-30")
	report, err := NewInventorySnapshotReport(
		filters,
		[]inventorydomain.AvailableStockSnapshot{
			{
				WarehouseID:   "wh-hcm",
				WarehouseCode: "HCM",
				LocationID:    "bin-a",
				LocationCode:  "A-01",
				ItemID:        "item-serum-30ml",
				SKU:           "serum-30ml",
				BatchID:       "batch-a",
				BatchNo:       "LOT-A",
				BatchQCStatus: inventorydomain.QCStatusPass,
				BatchStatus:   inventorydomain.BatchStatusActive,
				BatchExpiry:   time.Date(2026, 5, 20, 0, 0, 0, 0, time.UTC),
				BaseUOMCode:   decimal.MustUOMCode("PCS"),
				PhysicalQty:   decimal.MustQuantity("120"),
				ReservedQty:   decimal.MustQuantity("10"),
				QCHoldQty:     decimal.MustQuantity("8"),
				BlockedQty:    decimal.MustQuantity("2"),
				AvailableQty:  decimal.MustQuantity("100"),
			},
			{
				WarehouseID:   "wh-hcm",
				WarehouseCode: "HCM",
				LocationID:    "bin-b",
				LocationCode:  "B-01",
				SKU:           "cream-50g",
				BatchID:       "batch-b",
				BatchNo:       "LOT-B",
				BatchQCStatus: inventorydomain.QCStatusHold,
				BatchStatus:   inventorydomain.BatchStatusActive,
				BaseUOMCode:   decimal.MustUOMCode("PCS"),
				PhysicalQty:   decimal.MustQuantity("12"),
				ReservedQty:   decimal.MustQuantity("1"),
				QCHoldQty:     decimal.MustQuantity("6"),
				BlockedQty:    decimal.MustQuantity("0"),
				AvailableQty:  decimal.MustQuantity("5"),
			},
		},
		InventorySnapshotOptions{
			LowStockThreshold: "10",
			ExpiryWarningDays: 30,
			GeneratedAt:       time.Date(2026, 4, 30, 8, 0, 0, 0, time.UTC),
		},
	)
	if err != nil {
		t.Fatalf("NewInventorySnapshotReport returned error: %v", err)
	}

	if report.Metadata.SourceVersion != ReportingSourceVersion {
		t.Fatalf("source version = %q, want %q", report.Metadata.SourceVersion, ReportingSourceVersion)
	}
	if report.Summary.RowCount != 2 || report.Summary.LowStockRowCount != 1 || report.Summary.ExpiryWarningRows != 1 {
		t.Fatalf("summary = %+v, want row/low stock/expiry counts", report.Summary)
	}
	if len(report.Summary.TotalsByUOM) != 1 {
		t.Fatalf("totals length = %d, want 1", len(report.Summary.TotalsByUOM))
	}
	total := report.Summary.TotalsByUOM[0]
	if total.BaseUOMCode != "PCS" ||
		total.PhysicalQty != "132.000000" ||
		total.ReservedQty != "11.000000" ||
		total.QuarantineQty != "14.000000" ||
		total.BlockedQty != "2.000000" ||
		total.AvailableQty != "105.000000" {
		t.Fatalf("total = %+v, want aggregated PCS quantities", total)
	}

	first := report.Rows[0]
	if first.SKU != "SERUM-30ML" ||
		first.BatchExpiry != "2026-05-20" ||
		!first.ExpiryWarning ||
		first.LowStock ||
		first.QuarantineQty != "8.000000" ||
		first.SourceStockState != "quarantine" {
		t.Fatalf("first row = %+v", first)
	}
	assertInventorySourceReference(t, first.SourceReferences, "warehouse", "wh-hcm", "HCM")
	assertInventorySourceReference(t, first.SourceReferences, "item", "item-serum-30ml", "SERUM-30ML")
	assertInventorySourceReference(t, first.SourceReferences, "inventory_batch", "batch-a", "LOT-A")
	assertInventorySourceReference(t, first.SourceReferences, "stock_state", "wh-hcm:bin-a:SERUM-30ML:batch-a:quarantine", "quarantine")
	assertInventorySourceReference(t, first.SourceReferences, "inventory_warning", "wh-hcm:bin-a:SERUM-30ML:batch-a:quarantine:expiry_warning", "expiry_warning")
	second := report.Rows[1]
	if second.SKU != "CREAM-50G" || !second.LowStock {
		t.Fatalf("second row = %+v, want low-stock row", second)
	}
}

func TestNewInventorySnapshotReportLinksEveryQuantityBucketContext(t *testing.T) {
	report, err := NewInventorySnapshotReport(
		mustReportFilters(t, "2026-04-30"),
		[]inventorydomain.AvailableStockSnapshot{
			{
				WarehouseID:   "wh-hcm",
				WarehouseCode: "HCM",
				LocationID:    "bin-a",
				LocationCode:  "A-01",
				ItemID:        "item-mixed",
				SKU:           "mixed-sku",
				BatchID:       "batch-mixed",
				BatchNo:       "LOT-MIX",
				BatchQCStatus: inventorydomain.QCStatusPass,
				BatchStatus:   inventorydomain.BatchStatusActive,
				BatchExpiry:   time.Date(2026, 5, 10, 0, 0, 0, 0, time.UTC),
				BaseUOMCode:   decimal.MustUOMCode("PCS"),
				PhysicalQty:   decimal.MustQuantity("14"),
				ReservedQty:   decimal.MustQuantity("2"),
				QCHoldQty:     decimal.MustQuantity("3"),
				BlockedQty:    decimal.MustQuantity("4"),
				AvailableQty:  decimal.MustQuantity("5"),
			},
		},
		InventorySnapshotOptions{
			LowStockThreshold: "10",
			ExpiryWarningDays: 30,
			GeneratedAt:       time.Date(2026, 4, 30, 8, 0, 0, 0, time.UTC),
		},
	)
	if err != nil {
		t.Fatalf("NewInventorySnapshotReport returned error: %v", err)
	}

	row := report.Rows[0]
	if row.SourceStockState != "quarantine" || !row.LowStock || !row.ExpiryWarning {
		t.Fatalf("row = %+v, want quarantine primary state with low stock and expiry warning", row)
	}
	assertInventorySourceReference(t, row.SourceReferences, "stock_state", "wh-hcm:bin-a:MIXED-SKU:batch-mixed:quarantine", "quarantine")
	assertInventorySourceReference(t, row.SourceReferences, "stock_state", "wh-hcm:bin-a:MIXED-SKU:batch-mixed:available", "available")
	assertInventorySourceReference(t, row.SourceReferences, "stock_state", "wh-hcm:bin-a:MIXED-SKU:batch-mixed:reserved", "reserved")
	assertInventorySourceReference(t, row.SourceReferences, "stock_state", "wh-hcm:bin-a:MIXED-SKU:batch-mixed:blocked", "blocked")
	assertInventorySourceReference(t, row.SourceReferences, "inventory_warning", "wh-hcm:bin-a:MIXED-SKU:batch-mixed:available:low_stock", "low_stock")
	assertInventorySourceReference(t, row.SourceReferences, "inventory_warning", "wh-hcm:bin-a:MIXED-SKU:batch-mixed:quarantine:expiry_warning", "expiry_warning")
}

func assertInventorySourceReference(
	t *testing.T,
	references []ReportSourceReference,
	entityType string,
	id string,
	label string,
) {
	t.Helper()
	for _, reference := range references {
		if reference.EntityType == entityType && reference.ID == id {
			if reference.Label != label || reference.Href == "" || reference.Unavailable {
				t.Fatalf("reference = %+v, want label %q and available href", reference, label)
			}
			return
		}
	}

	t.Fatalf("references = %+v, missing %s %s", references, entityType, id)
}

func TestNewInventorySnapshotReportKeepsTotalsSeparatedByUOM(t *testing.T) {
	report, err := NewInventorySnapshotReport(
		mustReportFilters(t, "2026-04-30"),
		[]inventorydomain.AvailableStockSnapshot{
			{
				WarehouseID:  "wh-hcm",
				SKU:          "BOX-SET",
				BaseUOMCode:  decimal.MustUOMCode("SET"),
				PhysicalQty:  decimal.MustQuantity("3"),
				ReservedQty:  decimal.MustQuantity("1"),
				AvailableQty: decimal.MustQuantity("2"),
			},
			{
				WarehouseID:  "wh-hcm",
				SKU:          "BULK-BASE",
				BaseUOMCode:  decimal.MustUOMCode("KG"),
				PhysicalQty:  decimal.MustQuantity("4.5"),
				AvailableQty: decimal.MustQuantity("4.5"),
			},
		},
		InventorySnapshotOptions{},
	)
	if err != nil {
		t.Fatalf("NewInventorySnapshotReport returned error: %v", err)
	}

	if len(report.Summary.TotalsByUOM) != 2 {
		t.Fatalf("totals length = %d, want 2", len(report.Summary.TotalsByUOM))
	}
	if report.Summary.TotalsByUOM[0].BaseUOMCode != "KG" ||
		report.Summary.TotalsByUOM[1].BaseUOMCode != "SET" {
		t.Fatalf("totals = %+v, want sorted by UOM", report.Summary.TotalsByUOM)
	}
}

func TestNewInventorySnapshotReportFlagsExpiredRows(t *testing.T) {
	report, err := NewInventorySnapshotReport(
		mustReportFilters(t, "2026-04-30"),
		[]inventorydomain.AvailableStockSnapshot{
			{
				WarehouseID:  "wh-hcm",
				SKU:          "EXPIRED-SKU",
				BatchExpiry:  time.Date(2026, 4, 29, 0, 0, 0, 0, time.UTC),
				BaseUOMCode:  decimal.MustUOMCode("PCS"),
				PhysicalQty:  decimal.MustQuantity("2"),
				AvailableQty: decimal.MustQuantity("0"),
			},
		},
		InventorySnapshotOptions{},
	)
	if err != nil {
		t.Fatalf("NewInventorySnapshotReport returned error: %v", err)
	}

	if report.Summary.ExpiredRows != 1 || report.Summary.ExpiryWarningRows != 0 {
		t.Fatalf("summary = %+v, want expired without warning", report.Summary)
	}
	if !report.Rows[0].Expired || report.Rows[0].ExpiryWarning {
		t.Fatalf("row = %+v, want expired only", report.Rows[0])
	}
}

func TestNewInventorySnapshotReportRejectsInvalidInput(t *testing.T) {
	_, err := NewInventorySnapshotReport(
		mustReportFilters(t, "2026-04-30"),
		[]inventorydomain.AvailableStockSnapshot{{WarehouseID: "wh-hcm", SKU: "SKU-1"}},
		InventorySnapshotOptions{},
	)
	if !errors.Is(err, ErrInvalidInventorySnapshotReport) {
		t.Fatalf("error = %v, want ErrInvalidInventorySnapshotReport", err)
	}

	_, err = NewInventorySnapshotReport(
		mustReportFilters(t, "2026-04-30"),
		nil,
		InventorySnapshotOptions{LowStockThreshold: "-1"},
	)
	if !errors.Is(err, ErrInvalidInventorySnapshotReport) {
		t.Fatalf("threshold error = %v, want ErrInvalidInventorySnapshotReport", err)
	}
}

func mustReportFilters(t *testing.T, businessDate string) ReportFilters {
	t.Helper()
	filters, err := NewReportFilters(ReportFilterInput{BusinessDate: businessDate})
	if err != nil {
		t.Fatalf("NewReportFilters returned error: %v", err)
	}

	return filters
}
