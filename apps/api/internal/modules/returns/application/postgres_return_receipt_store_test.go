package application

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/Chinsusu/ERP-v2/apps/api/internal/modules/returns/domain"
	"github.com/Chinsusu/ERP-v2/apps/api/internal/shared/audit"

	_ "github.com/jackc/pgx/v5/stdlib"
)

const testReturnReceiptOrgID = "00000000-0000-4000-8000-000000000001"

func TestPostgresReturnReceiptStorePersistsReturnLifecycle(t *testing.T) {
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

	if err := seedReturnReceiptSmokeOrg(ctx, db); err != nil {
		t.Fatalf("seed org: %v", err)
	}

	suffix := fmt.Sprintf("%d", time.Now().UTC().UnixNano())
	store := NewPostgresReturnReceiptStore(
		db,
		PostgresReturnReceiptStoreConfig{DefaultOrgID: testReturnReceiptOrgID},
	)
	auditStore := audit.NewPostgresLogStore(
		db,
		audit.PostgresLogStoreConfig{DefaultOrgID: testReturnReceiptOrgID},
	)
	receivedAt := time.Date(2026, 5, 1, 8, 0, 0, 0, time.UTC)
	receiveReturn := NewPrototypeReceiveReturnAt(store, auditStore, receivedAt)

	received, err := receiveReturn.Execute(ctx, ReceiveReturnInput{
		WarehouseID:       "wh-hcm-return",
		WarehouseCode:     "WH-HCM-RETURN",
		Source:            "CARRIER",
		ScanCode:          "UNKNOWN-S11-04-02-" + suffix,
		PackageCondition:  "sealed bag",
		Disposition:       "needs_inspection",
		InvestigationNote: "postgres return receipt smoke",
		ActorID:           "user-return-ops",
		RequestID:         "req-return-receive-" + suffix,
	})
	if err != nil {
		t.Fatalf("receive return: %v", err)
	}

	loaded, err := store.FindReceiptByID(ctx, received.Receipt.ID)
	if err != nil {
		t.Fatalf("find received return: %v", err)
	}
	if loaded.ID != received.Receipt.ID ||
		loaded.Status != domain.ReturnStatusPendingInspection ||
		!loaded.UnknownCase ||
		len(loaded.Lines) != 1 ||
		loaded.Lines[0].Quantity != 1 {
		t.Fatalf("loaded receipt = %+v, want persisted pending unknown receipt", loaded)
	}

	inspectReturn := NewPrototypeInspectReturnAt(store, auditStore, receivedAt.Add(15*time.Minute))
	inspected, err := inspectReturn.Execute(ctx, InspectReturnInput{
		ReceiptID:     received.Receipt.ID,
		Condition:     "intact",
		Disposition:   "reusable",
		EvidenceLabel: "photo-s11-04-02",
		ActorID:       "user-return-inspector",
		RequestID:     "req-return-inspect-" + suffix,
	})
	if err != nil {
		t.Fatalf("inspect return: %v", err)
	}

	loadedInspection, err := store.FindInspectionByID(ctx, inspected.Inspection.ID)
	if err != nil {
		t.Fatalf("find inspection: %v", err)
	}
	if loadedInspection.ID != inspected.Inspection.ID ||
		loadedInspection.ReceiptID != received.Receipt.ID ||
		loadedInspection.Condition != domain.ReturnInspectionConditionIntact {
		t.Fatalf("loaded inspection = %+v, want persisted inspection", loadedInspection)
	}

	uploadAttachment := NewPrototypeUploadReturnAttachmentAt(
		store,
		auditStore,
		receivedAt.Add(20*time.Minute),
	).WithObjectStore(NewInMemoryReturnAttachmentObjectStore())
	attachment, err := uploadAttachment.Execute(ctx, UploadReturnAttachmentInput{
		ReceiptID:     received.Receipt.ID,
		InspectionID:  inspected.Inspection.ID,
		FileName:      "return-smoke.png",
		MIMEType:      "image/png",
		FileSizeBytes: 16,
		Content:       strings.NewReader("0123456789abcdef"),
		ActorID:       "user-return-inspector",
		RequestID:     "req-return-attachment-" + suffix,
	})
	if err != nil {
		t.Fatalf("upload attachment: %v", err)
	}

	applyDisposition := NewPrototypeApplyReturnDispositionAt(store, auditStore, receivedAt.Add(30*time.Minute))
	dispositioned, err := applyDisposition.Execute(ctx, ApplyReturnDispositionInput{
		ReceiptID:   received.Receipt.ID,
		Disposition: "reusable",
		Note:        "ready for putaway",
		ActorID:     "user-return-inspector",
		RequestID:   "req-return-disposition-" + suffix,
	})
	if err != nil {
		t.Fatalf("apply disposition: %v", err)
	}

	loaded, err = store.FindReceiptByID(ctx, received.Receipt.ID)
	if err != nil {
		t.Fatalf("find dispositioned return: %v", err)
	}
	if loaded.Status != domain.ReturnStatusDispositioned ||
		loaded.Disposition != domain.ReturnDispositionReusable ||
		loaded.StockMovement == nil ||
		loaded.StockMovement.ID != dispositioned.Receipt.StockMovement.ID {
		t.Fatalf("loaded dispositioned receipt = %+v, want movement summary", loaded)
	}

	var actionCount int
	if err := db.QueryRowContext(
		ctx,
		`SELECT count(*) FROM returns.return_disposition_actions WHERE action_ref = $1`,
		dispositioned.Action.ID,
	).Scan(&actionCount); err != nil {
		t.Fatalf("count disposition actions: %v", err)
	}
	var attachmentCount int
	if err := db.QueryRowContext(
		ctx,
		`SELECT count(*) FROM returns.return_attachments WHERE attachment_ref = $1`,
		attachment.Attachment.ID,
	).Scan(&attachmentCount); err != nil {
		t.Fatalf("count attachments: %v", err)
	}
	if actionCount != 1 || attachmentCount != 1 {
		t.Fatalf("actionCount=%d attachmentCount=%d, want persisted action and attachment", actionCount, attachmentCount)
	}
}

func TestPostgresReturnReceiptFindQueryPlacesWhereBeforeLimit(t *testing.T) {
	if !strings.Contains(findReturnReceiptHeaderSQL, "FROM returns.return_orders\nWHERE return_ref = $1") {
		t.Fatalf("find query does not place WHERE immediately after FROM:\n%s", findReturnReceiptHeaderSQL)
	}
	if strings.Contains(findReturnReceiptHeaderSQL, "ORDER BY created_at DESC, return_no DESC\nWHERE") {
		t.Fatalf("find query places WHERE after ORDER BY:\n%s", findReturnReceiptHeaderSQL)
	}
}

func TestBuildPostgresReturnReceiptMapsHeaderLinesAndMovement(t *testing.T) {
	createdAt := time.Date(2026, 5, 1, 8, 0, 0, 0, time.UTC)
	header := postgresReturnReceiptHeader{
		PersistedID:       "00000000-0000-4000-8000-000000000923",
		ID:                "rr-s11-test-001",
		OrgID:             "org-my-pham",
		ReceiptNo:         "RR-S11-TEST-001",
		WarehouseID:       "wh-hcm-return",
		WarehouseCode:     "WH-HCM-RETURN",
		Source:            "CARRIER",
		ReceivedBy:        "user-return-ops",
		ReceivedAt:        createdAt,
		PackageCondition:  "intact",
		Status:            "dispositioned",
		Disposition:       "reusable",
		TargetLocation:    "return-putaway-ready",
		OriginalOrderNo:   "SO-S11-TEST-001",
		TrackingNo:        "GHN-S11-TEST-001",
		ReturnCode:        "RET-S11-TEST-001",
		ScanCode:          "RET-S11-TEST-001",
		CustomerName:      "Nguyen Test",
		StockMovementRef:  "RR-S11-TEST-001-RESTOCK-001",
		StockMovementType: "return_restock",
		TargetStockStatus: "available",
		CreatedAt:         createdAt,
	}
	lines := []domain.ReturnReceiptLine{{
		ID:          "line-s11-test-001",
		SKU:         "SERUM-30ML",
		ProductName: "Hydrating Serum 30ml",
		Quantity:    2,
		Condition:   "intact",
	}}

	receipt, err := buildPostgresReturnReceipt(header, lines)
	if err != nil {
		t.Fatalf("buildPostgresReturnReceipt() error = %v", err)
	}

	if receipt.ID != "rr-s11-test-001" ||
		receipt.Status != domain.ReturnStatusDispositioned ||
		receipt.Disposition != domain.ReturnDispositionReusable ||
		len(receipt.Lines) != 1 ||
		receipt.Lines[0].Quantity != 2 ||
		receipt.StockMovement == nil ||
		receipt.StockMovement.TargetStockStatus != "available" {
		t.Fatalf("receipt = %+v, want mapped persisted return receipt", receipt)
	}
}

func TestPostgresReturnReceiptStoreRequiresDatabase(t *testing.T) {
	store := NewPostgresReturnReceiptStore(nil, PostgresReturnReceiptStoreConfig{})

	if _, err := store.List(nil, domain.ReturnReceiptFilter{}); err == nil {
		t.Fatal("List() error = nil, want database required error")
	}
	if err := store.Save(nil, domain.ReturnReceipt{ID: "rr-missing"}); err == nil {
		t.Fatal("Save() error = nil, want database required error")
	}
	if _, err := store.FindReceiptByID(nil, "rr-missing"); err == nil {
		t.Fatal("FindReceiptByID() error = nil, want database required error")
	}
	if err := store.SaveInspection(nil, domain.ReturnReceipt{}, domain.ReturnInspection{}); err == nil {
		t.Fatal("SaveInspection() error = nil, want database required error")
	}
	if _, err := store.FindInspectionByID(nil, "inspect-missing"); err == nil {
		t.Fatal("FindInspectionByID() error = nil, want database required error")
	}
	if err := store.SaveDisposition(nil, domain.ReturnReceipt{}, domain.ReturnDispositionAction{}); err == nil {
		t.Fatal("SaveDisposition() error = nil, want database required error")
	}
	if err := store.SaveAttachment(nil, domain.ReturnAttachment{}); err == nil {
		t.Fatal("SaveAttachment() error = nil, want database required error")
	}
}

func seedReturnReceiptSmokeOrg(ctx context.Context, db *sql.DB) error {
	_, err := db.ExecContext(
		ctx,
		`INSERT INTO core.organizations (id, code, name, status)
VALUES ($1, 'ERP_LOCAL', 'ERP Local Demo', 'active')
ON CONFLICT (code) DO UPDATE
SET name = EXCLUDED.name,
    status = EXCLUDED.status,
    updated_at = now()`,
		testReturnReceiptOrgID,
	)

	return err
}
