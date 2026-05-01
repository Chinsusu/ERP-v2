package application

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/Chinsusu/ERP-v2/apps/api/internal/modules/inventory/domain"
	"github.com/Chinsusu/ERP-v2/apps/api/internal/shared/audit"
	"github.com/Chinsusu/ERP-v2/apps/api/internal/shared/decimal"

	_ "github.com/jackc/pgx/v5/stdlib"
)

const testSupplierRejectionOrgID = "00000000-0000-4000-8000-000000000001"

func TestPostgresSupplierRejectionStorePersistsLifecycle(t *testing.T) {
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

	if err := seedSupplierRejectionSmokeOrg(ctx, db); err != nil {
		t.Fatalf("seed org: %v", err)
	}

	suffix := fmt.Sprintf("%d", time.Now().UTC().UnixNano())
	store := NewPostgresSupplierRejectionStore(
		db,
		PostgresSupplierRejectionStoreConfig{DefaultOrgID: testSupplierRejectionOrgID},
	)
	auditStore := audit.NewPostgresLogStore(
		db,
		audit.PostgresLogStoreConfig{DefaultOrgID: testSupplierRejectionOrgID},
	)
	createService := NewCreateSupplierRejection(store, auditStore)
	createService.clock = func() time.Time {
		return time.Date(2026, 5, 1, 9, 0, 0, 0, time.UTC)
	}
	input := validCreateSupplierRejectionInput()
	input.ID = "srj-s11-04-03-" + suffix
	input.RejectionNo = "SRJ-S11-04-03-" + suffix
	input.GoodsReceiptID = "grn-s11-04-03-" + suffix
	input.GoodsReceiptNo = "GRN-S11-04-03-" + suffix
	input.InboundQCInspectionID = "iqc-s11-04-03-" + suffix
	input.RequestID = "req-supplier-rejection-create-" + suffix
	input.Lines[0].ID = "srj-line-s11-04-03-" + suffix
	input.Lines[0].GoodsReceiptLineID = "grn-line-s11-04-03-" + suffix
	input.Lines[0].InboundQCInspectionID = input.InboundQCInspectionID
	input.Attachments[0].ID = "srj-att-s11-04-03-" + suffix
	input.Attachments[0].LineID = input.Lines[0].ID
	input.Attachments[0].ObjectKey = "supplier-rejections/" + input.ID + "/damage-photo.jpg"

	created, err := createService.Execute(ctx, input)
	if err != nil {
		t.Fatalf("create supplier rejection: %v", err)
	}
	loaded, err := store.Get(ctx, created.Rejection.ID)
	if err != nil {
		t.Fatalf("get persisted supplier rejection: %v", err)
	}
	if loaded.ID != created.Rejection.ID ||
		loaded.Status != domain.SupplierRejectionStatusDraft ||
		len(loaded.Lines) != 1 ||
		len(loaded.Attachments) != 1 ||
		loaded.Lines[0].RejectedQuantity.String() != "6.000000" {
		t.Fatalf("loaded rejection = %+v, want persisted draft with line and attachment", loaded)
	}

	transition := NewTransitionSupplierRejection(store, auditStore)
	transition.clock = func() time.Time {
		return time.Date(2026, 5, 1, 10, 0, 0, 0, time.UTC)
	}
	if _, err := transition.Submit(ctx, created.Rejection.ID, "user-warehouse-lead", "req-sr-submit-"+suffix); err != nil {
		t.Fatalf("submit supplier rejection: %v", err)
	}
	transition.clock = func() time.Time {
		return time.Date(2026, 5, 1, 11, 0, 0, 0, time.UTC)
	}
	confirmed, err := transition.Confirm(ctx, created.Rejection.ID, "user-warehouse-lead", "req-sr-confirm-"+suffix)
	if err != nil {
		t.Fatalf("confirm supplier rejection: %v", err)
	}

	loaded, err = store.Get(ctx, created.Rejection.ID)
	if err != nil {
		t.Fatalf("get confirmed supplier rejection: %v", err)
	}
	if loaded.Status != domain.SupplierRejectionStatusConfirmed ||
		loaded.SubmittedBy != "user-warehouse-lead" ||
		loaded.ConfirmedBy != "user-warehouse-lead" ||
		loaded.ConfirmedAt.IsZero() ||
		confirmed.CurrentStatus != domain.SupplierRejectionStatusConfirmed {
		t.Fatalf("loaded confirmed rejection = %+v, want persisted lifecycle", loaded)
	}

	rows, err := store.List(ctx, domain.NewSupplierRejectionFilter("supplier-local", "wh-hcm-fg", domain.SupplierRejectionStatusConfirmed))
	if err != nil {
		t.Fatalf("list supplier rejections: %v", err)
	}
	if len(rows) != 1 || rows[0].ID != created.Rejection.ID {
		t.Fatalf("rows = %+v, want confirmed supplier rejection", rows)
	}

	var lineCount int
	if err := db.QueryRowContext(
		ctx,
		`SELECT count(*) FROM inventory.supplier_rejection_lines WHERE line_ref = $1`,
		input.Lines[0].ID,
	).Scan(&lineCount); err != nil {
		t.Fatalf("count supplier rejection lines: %v", err)
	}
	var attachmentCount int
	if err := db.QueryRowContext(
		ctx,
		`SELECT count(*) FROM inventory.supplier_rejection_attachments WHERE attachment_ref = $1`,
		input.Attachments[0].ID,
	).Scan(&attachmentCount); err != nil {
		t.Fatalf("count supplier rejection attachments: %v", err)
	}
	if lineCount != 1 || attachmentCount != 1 {
		t.Fatalf("lineCount=%d attachmentCount=%d, want persisted line and attachment", lineCount, attachmentCount)
	}
}

func TestPostgresSupplierRejectionFindQueryPlacesWhereBeforeLimit(t *testing.T) {
	if !strings.Contains(findSupplierRejectionHeaderSQL, "FROM inventory.supplier_rejections\nWHERE rejection_ref = $1") {
		t.Fatalf("find query does not place WHERE immediately after FROM:\n%s", findSupplierRejectionHeaderSQL)
	}
	if strings.Contains(findSupplierRejectionHeaderSQL, "ORDER BY created_at DESC, rejection_no DESC\nWHERE") {
		t.Fatalf("find query places WHERE after ORDER BY:\n%s", findSupplierRejectionHeaderSQL)
	}
}

func TestBuildPostgresSupplierRejectionMapsLifecycle(t *testing.T) {
	createdAt := time.Date(2026, 5, 1, 9, 0, 0, 0, time.UTC)
	submittedAt := createdAt.Add(time.Hour)
	confirmedAt := createdAt.Add(2 * time.Hour)
	header := postgresSupplierRejectionHeader{
		PersistedID:           "00000000-0000-4000-8000-000000000924",
		ID:                    "srj-s11-test-001",
		OrgID:                 "org-my-pham",
		RejectionNo:           "SRJ-S11-TEST-001",
		SupplierID:            "supplier-local",
		SupplierCode:          "SUP-LOCAL",
		SupplierName:          "Local Supplier",
		PurchaseOrderID:       "po-s11-test-001",
		PurchaseOrderNo:       "PO-S11-TEST-001",
		GoodsReceiptID:        "grn-s11-test-001",
		GoodsReceiptNo:        "GRN-S11-TEST-001",
		InboundQCInspectionID: "iqc-s11-test-001",
		WarehouseID:           "wh-hcm-fg",
		WarehouseCode:         "WH-HCM-FG",
		Status:                "confirmed",
		Reason:                "damaged packaging",
		CreatedAt:             createdAt,
		CreatedBy:             "user-qa",
		UpdatedAt:             confirmedAt,
		UpdatedBy:             "user-warehouse-lead",
		SubmittedAt:           sql.NullTime{Time: submittedAt, Valid: true},
		SubmittedBy:           "user-warehouse-lead",
		ConfirmedAt:           sql.NullTime{Time: confirmedAt, Valid: true},
		ConfirmedBy:           "user-warehouse-lead",
	}
	lines := []domain.NewSupplierRejectionLineInput{{
		ID:                    "srj-line-s11-test-001",
		PurchaseOrderLineID:   "po-line-s11-test-001",
		GoodsReceiptLineID:    "grn-line-s11-test-001",
		InboundQCInspectionID: "iqc-s11-test-001",
		ItemID:                "item-serum-30ml",
		SKU:                   "SERUM-30ML",
		ItemName:              "Vitamin C Serum",
		BatchID:               "batch-serum-2604a",
		BatchNo:               "LOT-2604A",
		LotNo:                 "LOT-2604A",
		ExpiryDate:            time.Date(2027, 4, 1, 0, 0, 0, 0, time.UTC),
		RejectedQuantity:      decimal.MustQuantity("6"),
		UOMCode:               "EA",
		BaseUOMCode:           "EA",
		Reason:                "damaged packaging",
	}}
	attachments := []domain.NewSupplierRejectionAttachmentInput{{
		ID:          "srj-att-s11-test-001",
		LineID:      "srj-line-s11-test-001",
		FileName:    "damage-photo.jpg",
		ObjectKey:   "supplier-rejections/srj-s11-test-001/damage-photo.jpg",
		ContentType: "image/jpeg",
		UploadedAt:  createdAt,
		UploadedBy:  "user-qa",
		Source:      "inbound_qc",
	}}

	rejection, err := buildPostgresSupplierRejection(header, lines, attachments)
	if err != nil {
		t.Fatalf("buildPostgresSupplierRejection() error = %v", err)
	}
	if rejection.Status != domain.SupplierRejectionStatusConfirmed ||
		rejection.ConfirmedBy != "user-warehouse-lead" ||
		len(rejection.Lines) != 1 ||
		rejection.Lines[0].RejectedQuantity.String() != "6.000000" ||
		len(rejection.Attachments) != 1 {
		t.Fatalf("rejection = %+v, want confirmed rejection with line and attachment", rejection)
	}
}

func TestPostgresSupplierRejectionStoreRequiresDatabase(t *testing.T) {
	store := NewPostgresSupplierRejectionStore(nil, PostgresSupplierRejectionStoreConfig{})

	if _, err := store.List(nil, domain.SupplierRejectionFilter{}); err == nil {
		t.Fatal("List() error = nil, want database required error")
	}
	if _, err := store.Get(nil, "srj-missing"); err == nil {
		t.Fatal("Get() error = nil, want database required error")
	}
	if err := store.Save(nil, domain.SupplierRejection{}); err == nil {
		t.Fatal("Save() error = nil, want database required error")
	}
}

func seedSupplierRejectionSmokeOrg(ctx context.Context, db *sql.DB) error {
	_, err := db.ExecContext(
		ctx,
		`INSERT INTO core.organizations (id, code, name, status)
VALUES ($1, 'ERP_LOCAL', 'ERP Local Demo', 'active')
ON CONFLICT (code) DO UPDATE
SET name = EXCLUDED.name,
    status = EXCLUDED.status,
    updated_at = now()`,
		testSupplierRejectionOrgID,
	)

	return err
}
