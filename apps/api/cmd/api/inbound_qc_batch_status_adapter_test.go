package main

import (
	"context"
	"database/sql"
	"os"
	"testing"
	"time"

	inventoryapp "github.com/Chinsusu/ERP-v2/apps/api/internal/modules/inventory/application"
	inventorydomain "github.com/Chinsusu/ERP-v2/apps/api/internal/modules/inventory/domain"
	qcapp "github.com/Chinsusu/ERP-v2/apps/api/internal/modules/qc/application"
	"github.com/Chinsusu/ERP-v2/apps/api/internal/shared/audit"

	_ "github.com/jackc/pgx/v5/stdlib"
)

const (
	testInboundQCBatchStatusOrgID      = "00000000-0000-4000-8000-000000120301"
	testInboundQCBatchStatusUnitID     = "00000000-0000-4000-8000-000000120302"
	testInboundQCBatchStatusSupplierID = "00000000-0000-4000-8000-000000120303"
	testInboundQCBatchStatusItemID     = "00000000-0000-4000-8000-000000120304"
	testInboundQCBatchStatusBatchID    = "00000000-0000-4000-8000-000000120305"
	testInboundQCBatchStatusBatchRef   = "batch-s12-0301"
)

func TestInboundQCBatchQCStatusAdapterMapsInput(t *testing.T) {
	catalog := &recordingBatchCatalogStore{}
	adapter := inboundQCBatchQCStatusAdapter{catalog: catalog}
	changedAt := time.Date(2026, 5, 1, 11, 0, 0, 0, time.UTC)

	if err := adapter.ChangeInboundQCBatchQCStatus(context.Background(), qcapp.InboundQCBatchQCStatusInput{
		BatchID:     "batch-adapter",
		NextStatus:  inventorydomain.QCStatusPass,
		ActorID:     "user-qa",
		Reason:      "inbound qc pass",
		BusinessRef: "iqc-adapter",
		RequestID:   "req-adapter",
		ChangedAt:   changedAt,
	}); err != nil {
		t.Fatalf("ChangeInboundQCBatchQCStatus() error = %v", err)
	}

	got := catalog.input
	if got.BatchID != "batch-adapter" ||
		got.NextStatus != inventorydomain.QCStatusPass ||
		got.ActorID != "user-qa" ||
		got.Reason != "inbound qc pass" ||
		got.BusinessRef != "iqc-adapter" ||
		got.RequestID != "req-adapter" ||
		!got.ChangedAt.Equal(changedAt) {
		t.Fatalf("adapter input = %+v, want inbound QC values mapped to batch catalog input", got)
	}
}

func TestInboundQCBatchQCStatusAdapterPersistsWithPostgresCatalog(t *testing.T) {
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

	if err := seedInboundQCBatchStatusFixture(ctx, db); err != nil {
		t.Fatalf("seed inbound QC batch fixture: %v", err)
	}

	catalog := inventoryapp.NewPostgresBatchCatalogStore(db, audit.NewPostgresLogStore(db, audit.PostgresLogStoreConfig{}))
	adapter := inboundQCBatchQCStatusAdapter{catalog: catalog}
	changedAt := time.Date(2026, 5, 1, 11, 30, 0, 0, time.UTC)

	if err := adapter.ChangeInboundQCBatchQCStatus(ctx, qcapp.InboundQCBatchQCStatusInput{
		BatchID:     testInboundQCBatchStatusBatchRef,
		NextStatus:  inventorydomain.QCStatusPass,
		ActorID:     "user-qa",
		Reason:      "S12-03-01 inbound QC batch persistence integration",
		BusinessRef: "IQC-S12-03-01",
		RequestID:   "req-s12-03-01",
		ChangedAt:   changedAt,
	}); err != nil {
		t.Fatalf("ChangeInboundQCBatchQCStatus() error = %v", err)
	}

	var persistedQCStatus, updatedByRef string
	if err := db.QueryRowContext(
		ctx,
		`SELECT qc_status, COALESCE(updated_by_ref, '')
FROM inventory.batches
WHERE batch_ref = $1`,
		testInboundQCBatchStatusBatchRef,
	).Scan(&persistedQCStatus, &updatedByRef); err != nil {
		t.Fatalf("query persisted batch: %v", err)
	}
	if persistedQCStatus != "pass" || updatedByRef != "user-qa" {
		t.Fatalf("persisted batch = qc %q updated_by_ref %q, want pass/user-qa", persistedQCStatus, updatedByRef)
	}

	transitions, err := catalog.ListQCTransitions(ctx, testInboundQCBatchStatusBatchRef)
	if err != nil {
		t.Fatalf("ListQCTransitions() error = %v", err)
	}
	if !containsInboundQCBatchStatusTransition(transitions, inventorydomain.QCStatusHold, inventorydomain.QCStatusPass) {
		t.Fatalf("transitions = %+v, missing inbound QC hold -> pass audit transition", transitions)
	}
}

type recordingBatchCatalogStore struct {
	input inventoryapp.ChangeBatchQCStatusInput
}

func (s *recordingBatchCatalogStore) ListBatches(
	context.Context,
	inventorydomain.BatchFilter,
) ([]inventorydomain.Batch, error) {
	return nil, nil
}

func (s *recordingBatchCatalogStore) GetBatch(context.Context, string) (inventorydomain.Batch, error) {
	return inventorydomain.Batch{}, nil
}

func (s *recordingBatchCatalogStore) ChangeQCStatus(
	_ context.Context,
	input inventoryapp.ChangeBatchQCStatusInput,
) (inventoryapp.ChangeBatchQCStatusResult, error) {
	s.input = input

	return inventoryapp.ChangeBatchQCStatusResult{}, nil
}

func (s *recordingBatchCatalogStore) ListQCTransitions(
	context.Context,
	string,
) ([]inventorydomain.BatchQCTransition, error) {
	return nil, nil
}

func seedInboundQCBatchStatusFixture(ctx context.Context, db *sql.DB) error {
	if _, err := db.ExecContext(ctx, `
INSERT INTO core.organizations (id, code, name, status)
VALUES ($1::uuid, 'S12_0301_ORG', 'S12 Inbound QC Batch Test Org', 'active')
ON CONFLICT (code) DO UPDATE
SET name = EXCLUDED.name,
    status = EXCLUDED.status,
    updated_at = now()`,
		testInboundQCBatchStatusOrgID,
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
		testInboundQCBatchStatusUnitID,
		testInboundQCBatchStatusOrgID,
	); err != nil {
		return err
	}
	if _, err := db.ExecContext(ctx, `
INSERT INTO mdm.suppliers (id, org_id, code, name, supplier_type, status)
VALUES ($1::uuid, $2::uuid, 'SUP-S12-0301', 'S12 Inbound QC Supplier', 'supplier', 'active')
ON CONFLICT (org_id, code) DO UPDATE
SET name = EXCLUDED.name,
    supplier_type = EXCLUDED.supplier_type,
    status = EXCLUDED.status,
    updated_at = now()`,
		testInboundQCBatchStatusSupplierID,
		testInboundQCBatchStatusOrgID,
	); err != nil {
		return err
	}
	if _, err := db.ExecContext(ctx, `
INSERT INTO mdm.items (id, org_id, sku, name, item_type, base_unit_id, requires_batch, requires_expiry, status)
VALUES ($1::uuid, $2::uuid, 'S12-IQC-SERUM', 'S12 Inbound QC Serum', 'finished_good', $3::uuid, true, true, 'active')
ON CONFLICT (org_id, sku) DO UPDATE
SET name = EXCLUDED.name,
    item_type = EXCLUDED.item_type,
    base_unit_id = EXCLUDED.base_unit_id,
    requires_batch = EXCLUDED.requires_batch,
    requires_expiry = EXCLUDED.requires_expiry,
    status = EXCLUDED.status,
    updated_at = now()`,
		testInboundQCBatchStatusItemID,
		testInboundQCBatchStatusOrgID,
		testInboundQCBatchStatusUnitID,
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
  'LOT-S12-0301',
  $4::uuid,
  '2026-05-01',
  '2027-05-01',
  'hold',
  'active',
  $5,
  'org-s12-0301',
  'item-s12-iqc-serum',
  'supplier-s12-0301',
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
		testInboundQCBatchStatusBatchID,
		testInboundQCBatchStatusOrgID,
		testInboundQCBatchStatusItemID,
		testInboundQCBatchStatusSupplierID,
		testInboundQCBatchStatusBatchRef,
	)

	return err
}

func containsInboundQCBatchStatusTransition(
	transitions []inventorydomain.BatchQCTransition,
	from inventorydomain.QCStatus,
	to inventorydomain.QCStatus,
) bool {
	for _, transition := range transitions {
		if transition.FromQCStatus == from &&
			transition.ToQCStatus == to &&
			transition.BusinessRef == "IQC-S12-03-01" {
			return true
		}
	}

	return false
}
