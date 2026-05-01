package application

import (
	"database/sql"
	"strings"
	"testing"
	"time"

	"github.com/Chinsusu/ERP-v2/apps/api/internal/modules/inventory/domain"
)

func TestBuildPostgresBatchListQueryAppliesFilters(t *testing.T) {
	query, args := buildPostgresBatchListQuery(domain.NewBatchFilter(
		"serum-30ml",
		domain.QCStatusHold,
		domain.BatchStatusActive,
	))

	for _, want := range []string{
		"upper(item.sku) = $1",
		"batch.qc_status = $2",
		"batch.status = $3",
		"ORDER BY item.sku",
	} {
		if !strings.Contains(query, want) {
			t.Fatalf("query missing %q:\n%s", want, query)
		}
	}
	if len(args) != 3 {
		t.Fatalf("args = %v, want three filters", args)
	}
	if args[0] != "SERUM-30ML" || args[1] != "hold" || args[2] != "active" {
		t.Fatalf("args = %v, want normalized filters", args)
	}
}

func TestBuildPostgresBatchListQueryWithoutFiltersHasNoWhereClause(t *testing.T) {
	query, args := buildPostgresBatchListQuery(domain.BatchFilter{})

	if strings.Contains(query, "\nWHERE ") {
		t.Fatalf("query has WHERE clause without filters:\n%s", query)
	}
	if len(args) != 0 {
		t.Fatalf("args = %v, want empty", args)
	}
}

func TestScanPostgresBatchRowMapsRuntimeRefs(t *testing.T) {
	mfgDate := time.Date(2026, 4, 1, 0, 0, 0, 0, time.UTC)
	expiryDate := time.Date(2027, 4, 1, 0, 0, 0, 0, time.UTC)
	createdAt := time.Date(2026, 4, 27, 9, 0, 0, 0, time.UTC)
	updatedAt := createdAt.Add(time.Hour)

	row, err := scanPostgresBatchRow(fakePostgresBatchScanner{values: []any{
		"00000000-0000-4000-8000-000000001206",
		"00000000-0000-4000-8000-000000000001",
		"batch-serum-2604a",
		"org-my-pham",
		"item-serum-30ml",
		"serum-30ml",
		"Vitamin C Serum",
		"lot-2604a",
		"sup-rm-bioactive",
		sql.NullTime{Time: mfgDate, Valid: true},
		sql.NullTime{Time: expiryDate, Valid: true},
		"hold",
		"active",
		createdAt,
		updatedAt,
	}})
	if err != nil {
		t.Fatalf("scanPostgresBatchRow() error = %v", err)
	}

	if row.persistedID != "00000000-0000-4000-8000-000000001206" ||
		row.persistedOrgID != "00000000-0000-4000-8000-000000000001" {
		t.Fatalf("persisted refs = %q/%q", row.persistedID, row.persistedOrgID)
	}
	if row.batch.ID != "batch-serum-2604a" ||
		row.batch.OrgID != "org-my-pham" ||
		row.batch.ItemID != "item-serum-30ml" ||
		row.batch.SKU != "SERUM-30ML" ||
		row.batch.BatchNo != "LOT-2604A" ||
		row.batch.QCStatus != domain.QCStatusHold ||
		row.batch.Status != domain.BatchStatusActive ||
		!row.batch.ExpiryDate.Equal(expiryDate) {
		t.Fatalf("batch = %+v, want normalized runtime refs", row.batch)
	}
}

func TestPostgresBatchCatalogStoreRequiresDatabase(t *testing.T) {
	store := NewPostgresBatchCatalogStore(nil, nil)

	if _, err := store.ListBatches(nil, domain.BatchFilter{}); err == nil {
		t.Fatal("ListBatches() error = nil, want database required error")
	}
	if _, err := store.GetBatch(nil, "batch-serum-2604a"); err == nil {
		t.Fatal("GetBatch() error = nil, want database required error")
	}
	if _, err := store.ChangeQCStatus(nil, ChangeBatchQCStatusInput{}); err == nil {
		t.Fatal("ChangeQCStatus() error = nil, want database required error")
	}
}

type fakePostgresBatchScanner struct {
	values []any
}

func (s fakePostgresBatchScanner) Scan(dest ...any) error {
	for index := range dest {
		switch target := dest[index].(type) {
		case *string:
			*target = s.values[index].(string)
		case *sql.NullTime:
			*target = s.values[index].(sql.NullTime)
		case *time.Time:
			*target = s.values[index].(time.Time)
		default:
			panic("unsupported scan destination")
		}
	}

	return nil
}
