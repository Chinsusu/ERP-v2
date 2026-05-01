package application

import (
	"context"
	"database/sql"
	"errors"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/Chinsusu/ERP-v2/apps/api/internal/modules/inventory/domain"
	"github.com/Chinsusu/ERP-v2/apps/api/internal/shared/audit"

	_ "github.com/jackc/pgx/v5/stdlib"
)

const (
	testBatchCatalogOrgID      = "00000000-0000-4000-8000-000000120104"
	testBatchCatalogUnitID     = "00000000-0000-4000-8000-000000120105"
	testBatchCatalogSupplierID = "00000000-0000-4000-8000-000000120106"
	testBatchCatalogItemID     = "00000000-0000-4000-8000-000000120107"
	testBatchCatalogBatchID    = "00000000-0000-4000-8000-000000120108"
	testBatchCatalogBatchRef   = "batch-s12-0104"
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

func TestPostgresBatchCatalogStorePersistsQCTransitionWithAudit(t *testing.T) {
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

	if err := seedPostgresBatchCatalogFixture(ctx, db); err != nil {
		t.Fatalf("seed batch catalog fixture: %v", err)
	}

	store := NewPostgresBatchCatalogStore(db, audit.NewPostgresLogStore(db, audit.PostgresLogStoreConfig{}))
	before, err := store.GetBatch(ctx, testBatchCatalogBatchRef)
	if err != nil {
		t.Fatalf("GetBatch() before transition error = %v", err)
	}
	if before.QCStatus != domain.QCStatusHold {
		t.Fatalf("before QCStatus = %s, want hold", before.QCStatus)
	}

	changedAt := time.Date(2026, 5, 1, 10, 0, 0, 0, time.UTC)
	result, err := store.ChangeQCStatus(ctx, ChangeBatchQCStatusInput{
		BatchID:     testBatchCatalogBatchRef,
		NextStatus:  domain.QCStatusPass,
		ActorID:     "user-qa",
		Reason:      "S12-01-04 integration test",
		BusinessRef: "IQC-S12-01-04",
		RequestID:   "req-s12-01-04",
		ChangedAt:   changedAt,
	})
	if err != nil {
		t.Fatalf("ChangeQCStatus() error = %v", err)
	}
	if result.Batch.QCStatus != domain.QCStatusPass ||
		result.Transition.FromQCStatus != domain.QCStatusHold ||
		result.Transition.ToQCStatus != domain.QCStatusPass ||
		result.Transition.ActorID != "user-qa" {
		t.Fatalf("transition result = %+v, want hold -> pass by user-qa", result)
	}

	after, err := store.GetBatch(ctx, testBatchCatalogBatchRef)
	if err != nil {
		t.Fatalf("GetBatch() after transition error = %v", err)
	}
	if after.QCStatus != domain.QCStatusPass {
		t.Fatalf("after QCStatus = %s, want pass", after.QCStatus)
	}

	transitions, err := store.ListQCTransitions(ctx, testBatchCatalogBatchRef)
	if err != nil {
		t.Fatalf("ListQCTransitions() error = %v", err)
	}
	if !containsBatchTransition(transitions, result.AuditLogID, domain.QCStatusHold, domain.QCStatusPass) {
		t.Fatalf("transitions = %+v, missing audit log %s", transitions, result.AuditLogID)
	}

	var persistedQCStatus, updatedByRef string
	if err := db.QueryRowContext(
		ctx,
		`SELECT qc_status, COALESCE(updated_by_ref, '')
FROM inventory.batches
WHERE batch_ref = $1`,
		testBatchCatalogBatchRef,
	).Scan(&persistedQCStatus, &updatedByRef); err != nil {
		t.Fatalf("query persisted batch: %v", err)
	}
	if persistedQCStatus != "pass" || updatedByRef != "user-qa" {
		t.Fatalf("persisted batch = qc %q updated_by_ref %q, want pass/user-qa", persistedQCStatus, updatedByRef)
	}

	if _, err := store.ChangeQCStatus(ctx, ChangeBatchQCStatusInput{
		BatchID:    testBatchCatalogBatchRef,
		NextStatus: domain.QCStatusFail,
		ActorID:    "user-qa",
		Reason:     "invalid transition from pass",
		ChangedAt:  changedAt.Add(time.Minute),
	}); !errors.Is(err, domain.ErrBatchInvalidQCTransition) {
		t.Fatalf("ChangeQCStatus() invalid err = %v, want ErrBatchInvalidQCTransition", err)
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

func seedPostgresBatchCatalogFixture(ctx context.Context, db *sql.DB) error {
	if _, err := db.ExecContext(ctx, `
INSERT INTO core.organizations (id, code, name, status)
VALUES ($1::uuid, 'S12_0104_ORG', 'S12 Batch Catalog Test Org', 'active')
ON CONFLICT (code) DO UPDATE
SET name = EXCLUDED.name,
    status = EXCLUDED.status,
    updated_at = now()`,
		testBatchCatalogOrgID,
	); err != nil {
		return err
	}
	if _, err := db.ExecContext(ctx, `
INSERT INTO mdm.units (id, org_id, code, name, precision_scale, status)
VALUES ($1::uuid, $2::uuid, 'PCS', 'Piece', 6, 'active')
ON CONFLICT (org_id, code) DO UPDATE
SET name = EXCLUDED.name,
    precision_scale = EXCLUDED.precision_scale,
    status = EXCLUDED.status,
    updated_at = now()`,
		testBatchCatalogUnitID,
		testBatchCatalogOrgID,
	); err != nil {
		return err
	}
	if _, err := db.ExecContext(ctx, `
INSERT INTO mdm.suppliers (id, org_id, code, name, supplier_type, status)
VALUES ($1::uuid, $2::uuid, 'SUP-S12-0104', 'S12 Supplier', 'supplier', 'active')
ON CONFLICT (org_id, code) DO UPDATE
SET name = EXCLUDED.name,
    supplier_type = EXCLUDED.supplier_type,
    status = EXCLUDED.status,
    updated_at = now()`,
		testBatchCatalogSupplierID,
		testBatchCatalogOrgID,
	); err != nil {
		return err
	}
	if _, err := db.ExecContext(ctx, `
INSERT INTO mdm.items (id, org_id, sku, name, item_type, base_unit_id, requires_batch, requires_expiry, status)
VALUES ($1::uuid, $2::uuid, 'S12-SERUM', 'S12 Serum', 'finished_good', $3::uuid, true, true, 'active')
ON CONFLICT (org_id, sku) DO UPDATE
SET name = EXCLUDED.name,
    item_type = EXCLUDED.item_type,
    base_unit_id = EXCLUDED.base_unit_id,
    requires_batch = EXCLUDED.requires_batch,
    requires_expiry = EXCLUDED.requires_expiry,
    status = EXCLUDED.status,
    updated_at = now()`,
		testBatchCatalogItemID,
		testBatchCatalogOrgID,
		testBatchCatalogUnitID,
	); err != nil {
		return err
	}
	_, err := db.ExecContext(ctx, `
INSERT INTO inventory.batches (
  id,
  org_id,
  item_id,
  batch_no,
  supplier_id,
  mfg_date,
  expiry_date,
  qc_status,
  status,
  batch_ref,
  org_ref,
  item_ref,
  supplier_ref,
  created_by_ref,
  updated_by_ref
) VALUES (
  $1::uuid,
  $2::uuid,
  $3::uuid,
  'LOT-S12-0104',
  $4::uuid,
  '2026-05-01',
  '2027-05-01',
  'hold',
  'active',
  $5,
  'org-s12-0104',
  'item-s12-serum',
  'supplier-s12',
  'user-qa',
  'user-qa'
)
ON CONFLICT (item_id, batch_no) DO UPDATE
SET qc_status = 'hold',
    status = 'active',
    batch_ref = EXCLUDED.batch_ref,
    org_ref = EXCLUDED.org_ref,
    item_ref = EXCLUDED.item_ref,
    supplier_ref = EXCLUDED.supplier_ref,
    updated_by_ref = EXCLUDED.updated_by_ref,
    updated_at = now()`,
		testBatchCatalogBatchID,
		testBatchCatalogOrgID,
		testBatchCatalogItemID,
		testBatchCatalogSupplierID,
		testBatchCatalogBatchRef,
	)

	return err
}

func containsBatchTransition(
	transitions []domain.BatchQCTransition,
	auditLogID string,
	from domain.QCStatus,
	to domain.QCStatus,
) bool {
	for _, transition := range transitions {
		if transition.AuditLogID == auditLogID &&
			transition.FromQCStatus == from &&
			transition.ToQCStatus == to {
			return true
		}
	}

	return false
}
