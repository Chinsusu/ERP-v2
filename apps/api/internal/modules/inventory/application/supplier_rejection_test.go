package application

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/Chinsusu/ERP-v2/apps/api/internal/modules/inventory/domain"
	"github.com/Chinsusu/ERP-v2/apps/api/internal/shared/audit"
)

func TestCreateSupplierRejectionPersistsTraceableDraftAndAudit(t *testing.T) {
	store := NewPrototypeSupplierRejectionStore()
	auditStore := audit.NewInMemoryLogStore()
	service := NewCreateSupplierRejection(store, auditStore)
	service.clock = func() time.Time {
		return time.Date(2026, 4, 29, 9, 0, 0, 0, time.UTC)
	}

	result, err := service.Execute(context.Background(), validCreateSupplierRejectionInput())
	if err != nil {
		t.Fatalf("create supplier rejection: %v", err)
	}
	if result.Rejection.Status != domain.SupplierRejectionStatusDraft ||
		result.Rejection.RejectionNo != "SRJ-260429-0001" ||
		result.Rejection.SupplierID != "supplier-local" ||
		result.Rejection.Lines[0].RejectedQuantity.String() != "6.000000" ||
		result.AuditLogID == "" {
		t.Fatalf("result = %+v, want traceable draft rejection", result)
	}

	saved, err := store.Get(context.Background(), result.Rejection.ID)
	if err != nil {
		t.Fatalf("get supplier rejection: %v", err)
	}
	if saved.InboundQCInspectionID != "iqc-fail-001" || len(saved.Attachments) != 1 {
		t.Fatalf("saved rejection = %+v, want QC link and evidence", saved)
	}
	logs, err := auditStore.List(context.Background(), audit.Query{Action: supplierRejectionCreatedAction})
	if err != nil {
		t.Fatalf("list audit logs: %v", err)
	}
	if len(logs) != 1 ||
		logs[0].EntityID != result.Rejection.ID ||
		logs[0].AfterData["total_rejected_base_qty"] != "6.000000" ||
		logs[0].Metadata["reason"] != "damaged packaging" {
		t.Fatalf("audit logs = %+v, want supplier rejection create audit", logs)
	}
}

func TestSupplierRejectionStorePreventsDuplicateRejectionNo(t *testing.T) {
	store := NewPrototypeSupplierRejectionStore()
	service := NewCreateSupplierRejection(store, audit.NewInMemoryLogStore())
	service.clock = func() time.Time {
		return time.Date(2026, 4, 29, 9, 0, 0, 0, time.UTC)
	}

	if _, err := service.Execute(context.Background(), validCreateSupplierRejectionInput()); err != nil {
		t.Fatalf("create first supplier rejection: %v", err)
	}
	duplicate := validCreateSupplierRejectionInput()
	duplicate.ID = "srj-260429-duplicate"

	_, err := service.Execute(context.Background(), duplicate)
	if !errors.Is(err, ErrSupplierRejectionDuplicate) {
		t.Fatalf("err = %v, want duplicate rejection no", err)
	}
}

func TestSupplierRejectionStoreAllowsUpdatingSameRecord(t *testing.T) {
	store := NewPrototypeSupplierRejectionStore()
	service := NewCreateSupplierRejection(store, audit.NewInMemoryLogStore())
	service.clock = func() time.Time {
		return time.Date(2026, 4, 29, 9, 0, 0, 0, time.UTC)
	}
	result, err := service.Execute(context.Background(), validCreateSupplierRejectionInput())
	if err != nil {
		t.Fatalf("create supplier rejection: %v", err)
	}

	submitted, err := result.Rejection.Submit("user-warehouse-lead", time.Date(2026, 4, 29, 10, 0, 0, 0, time.UTC))
	if err != nil {
		t.Fatalf("submit supplier rejection: %v", err)
	}
	if err := store.Save(context.Background(), submitted); err != nil {
		t.Fatalf("save updated supplier rejection: %v", err)
	}
	saved, err := store.Get(context.Background(), result.Rejection.ID)
	if err != nil {
		t.Fatalf("get supplier rejection: %v", err)
	}
	if saved.Status != domain.SupplierRejectionStatusSubmitted {
		t.Fatalf("saved status = %q, want submitted", saved.Status)
	}
}

func TestTransitionSupplierRejectionSubmitsAndConfirmsWithAudit(t *testing.T) {
	store := NewPrototypeSupplierRejectionStore()
	auditStore := audit.NewInMemoryLogStore()
	createService := NewCreateSupplierRejection(store, auditStore)
	createService.clock = func() time.Time {
		return time.Date(2026, 4, 29, 9, 0, 0, 0, time.UTC)
	}
	created, err := createService.Execute(context.Background(), validCreateSupplierRejectionInput())
	if err != nil {
		t.Fatalf("create supplier rejection: %v", err)
	}
	transition := NewTransitionSupplierRejection(store, auditStore)
	transition.clock = func() time.Time {
		return time.Date(2026, 4, 29, 10, 0, 0, 0, time.UTC)
	}

	submitted, err := transition.Submit(
		context.Background(),
		created.Rejection.ID,
		"user-warehouse-lead",
		"req-supplier-rejection-submit",
	)
	if err != nil {
		t.Fatalf("submit supplier rejection: %v", err)
	}
	transition.clock = func() time.Time {
		return time.Date(2026, 4, 29, 11, 0, 0, 0, time.UTC)
	}
	confirmed, err := transition.Confirm(
		context.Background(),
		created.Rejection.ID,
		"user-warehouse-lead",
		"req-supplier-rejection-confirm",
	)
	if err != nil {
		t.Fatalf("confirm supplier rejection: %v", err)
	}

	if submitted.PreviousStatus != domain.SupplierRejectionStatusDraft ||
		submitted.CurrentStatus != domain.SupplierRejectionStatusSubmitted ||
		confirmed.PreviousStatus != domain.SupplierRejectionStatusSubmitted ||
		confirmed.CurrentStatus != domain.SupplierRejectionStatusConfirmed {
		t.Fatalf("submitted = %+v confirmed = %+v, want draft -> submitted -> confirmed", submitted, confirmed)
	}
	logs, err := auditStore.List(context.Background(), audit.Query{EntityID: created.Rejection.ID})
	if err != nil {
		t.Fatalf("list audit logs: %v", err)
	}
	if len(logs) != 3 {
		t.Fatalf("audit logs = %d, want create, submit, confirm", len(logs))
	}
}

func TestListSupplierRejectionsFiltersRows(t *testing.T) {
	store := NewPrototypeSupplierRejectionStore()
	service := NewCreateSupplierRejection(store, audit.NewInMemoryLogStore())
	service.clock = func() time.Time {
		return time.Date(2026, 4, 29, 9, 0, 0, 0, time.UTC)
	}
	if _, err := service.Execute(context.Background(), validCreateSupplierRejectionInput()); err != nil {
		t.Fatalf("create supplier rejection: %v", err)
	}

	rows, err := NewListSupplierRejections(store).Execute(
		context.Background(),
		domain.NewSupplierRejectionFilter("supplier-local", "wh-hcm-fg", domain.SupplierRejectionStatusDraft),
	)
	if err != nil {
		t.Fatalf("list supplier rejections: %v", err)
	}
	if len(rows) != 1 || rows[0].RejectionNo != "SRJ-260429-0001" {
		t.Fatalf("rows = %+v, want one matching supplier rejection", rows)
	}
}

func validCreateSupplierRejectionInput() CreateSupplierRejectionInput {
	return CreateSupplierRejectionInput{
		ID:                    "srj-260429-0001",
		OrgID:                 "org-my-pham",
		RejectionNo:           "SRJ-260429-0001",
		SupplierID:            "supplier-local",
		SupplierCode:          "SUP-LOCAL",
		SupplierName:          "Local Supplier",
		PurchaseOrderID:       "po-260427-0003",
		PurchaseOrderNo:       "PO-260427-0003",
		GoodsReceiptID:        "grn-hcm-260427-inspect",
		GoodsReceiptNo:        "GRN-260427-0003",
		InboundQCInspectionID: "iqc-fail-001",
		WarehouseID:           "wh-hcm-fg",
		WarehouseCode:         "WH-HCM-FG",
		Reason:                "damaged packaging",
		ActorID:               "user-qa",
		RequestID:             "req-supplier-rejection-create",
		Lines: []CreateSupplierRejectionLineInput{
			{
				ID:                    "srj-line-001",
				PurchaseOrderLineID:   "po-line-260427-0003-001",
				GoodsReceiptLineID:    "grn-line-draft-001",
				InboundQCInspectionID: "iqc-fail-001",
				ItemID:                "item-serum-30ml",
				SKU:                   "serum-30ml",
				ItemName:              "Vitamin C Serum",
				BatchID:               "batch-serum-2604a",
				BatchNo:               "LOT-2604A",
				LotNo:                 "LOT-2604A",
				ExpiryDate:            time.Date(2027, 4, 1, 0, 0, 0, 0, time.UTC),
				RejectedQuantity:      "6.000000",
				UOMCode:               "EA",
				BaseUOMCode:           "EA",
				Reason:                "damaged packaging",
			},
		},
		Attachments: []CreateSupplierRejectionAttachmentInput{
			{
				ID:          "srj-att-001",
				LineID:      "srj-line-001",
				FileName:    "damage-photo.jpg",
				ObjectKey:   "supplier-rejections/srj-260429-0001/damage-photo.jpg",
				ContentType: "image/jpeg",
				Source:      "inbound_qc",
			},
		},
	}
}
