package main

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	inventoryapp "github.com/Chinsusu/ERP-v2/apps/api/internal/modules/inventory/application"
	inventorydomain "github.com/Chinsusu/ERP-v2/apps/api/internal/modules/inventory/domain"
	masterdataapp "github.com/Chinsusu/ERP-v2/apps/api/internal/modules/masterdata/application"
	purchasedomain "github.com/Chinsusu/ERP-v2/apps/api/internal/modules/purchase/domain"
	qcapp "github.com/Chinsusu/ERP-v2/apps/api/internal/modules/qc/application"
	"github.com/Chinsusu/ERP-v2/apps/api/internal/shared/audit"
	"github.com/Chinsusu/ERP-v2/apps/api/internal/shared/auth"
	"github.com/Chinsusu/ERP-v2/apps/api/internal/shared/decimal"
	"github.com/Chinsusu/ERP-v2/apps/api/internal/shared/response"
)

func TestInboundPassReceivingE2ESmoke(t *testing.T) {
	authConfig := smokeAuthConfig()
	services := newInboundQCE2EServices(approvedInboundPassE2EPurchaseOrder())

	createReceiptReq := smokeRequestAsRole(
		httptest.NewRequest(http.MethodPost, "/api/v1/goods-receipts", bytes.NewBufferString(`{
			"id": "grn-e2e-inbound-pass",
			"receipt_no": "GRN-260429-PASS-E2E",
			"warehouse_id": "wh-hcm-fg",
			"location_id": "loc-hcm-fg-recv-01",
			"reference_doc_type": "purchase_order",
			"reference_doc_id": "PO-260429-PASS-E2E",
			"supplier_id": "supplier-local",
			"delivery_note_no": "DN-260429-PASS-E2E",
			"lines": [
				{
					"id": "line-e2e-inbound-pass",
					"purchase_order_line_id": "po-line-260429-pass-e2e-001",
					"batch_id": "batch-serum-2604a",
					"quantity": "8",
					"uom_code": "EA",
					"base_uom_code": "EA",
					"packaging_status": "intact"
				}
			]
		}`)),
		authConfig,
		auth.RoleWarehouseLead,
	)
	createReceiptReq.Header.Set(response.HeaderRequestID, "req-e2e-pass-receiving-create")
	createReceiptRec := httptest.NewRecorder()

	goodsReceiptsHandler(services.receiving).ServeHTTP(createReceiptRec, createReceiptReq)

	if createReceiptRec.Code != http.StatusCreated {
		t.Fatalf("create receipt status = %d, want %d: %s", createReceiptRec.Code, http.StatusCreated, createReceiptRec.Body.String())
	}
	receipt := decodeSmokeSuccess[warehouseReceivingResponse](t, createReceiptRec).Data
	if receipt.Status != "draft" ||
		receipt.Lines[0].ID != "line-e2e-inbound-pass" ||
		receipt.Lines[0].PackagingStatus != "intact" ||
		receipt.AuditLogID == "" {
		t.Fatalf("receipt = %+v, want draft intact inbound line with audit", receipt)
	}

	for _, step := range []struct {
		name    string
		handler http.HandlerFunc
	}{
		{name: "submit", handler: submitGoodsReceiptHandler(services.receiving)},
		{name: "inspect", handler: markGoodsReceiptInspectReadyHandler(services.receiving)},
	} {
		req := smokeRequestAsRole(
			httptest.NewRequest(http.MethodPost, "/api/v1/goods-receipts/"+receipt.ID+"/"+step.name, nil),
			authConfig,
			auth.RoleWarehouseLead,
		)
		req.SetPathValue("receipt_id", receipt.ID)
		req.Header.Set(response.HeaderRequestID, "req-e2e-pass-receiving-"+step.name)
		rec := httptest.NewRecorder()

		step.handler.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("%s receipt status = %d, want %d: %s", step.name, rec.Code, http.StatusOK, rec.Body.String())
		}
	}
	if services.movementStore.Count() != 0 {
		t.Fatalf("stock movements before QC pass = %d, want 0", services.movementStore.Count())
	}

	createQCReq := smokeRequestAsRole(
		httptest.NewRequest(http.MethodPost, "/api/v1/inbound-qc-inspections", bytes.NewBufferString(`{
			"id": "iqc-e2e-inbound-pass",
			"goods_receipt_id": "grn-e2e-inbound-pass",
			"goods_receipt_line_id": "line-e2e-inbound-pass",
			"note": "inbound goods matched PO, lot, expiry, and packaging"
		}`)),
		authConfig,
		auth.RoleQA,
	)
	createQCReq.Header.Set(response.HeaderRequestID, "req-e2e-pass-qc-create")
	createQCRec := httptest.NewRecorder()

	inboundQCInspectionsHandler(services.inboundQC).ServeHTTP(createQCRec, createQCReq)

	if createQCRec.Code != http.StatusCreated {
		t.Fatalf("create qc status = %d, want %d: %s", createQCRec.Code, http.StatusCreated, createQCRec.Body.String())
	}
	createdQC := decodeSmokeSuccess[inboundQCActionResultResponse](t, createQCRec).Data
	if createdQC.Inspection.Status != "pending" ||
		createdQC.Inspection.GoodsReceiptID != receipt.ID ||
		createdQC.Inspection.GoodsReceiptLineID != receipt.Lines[0].ID {
		t.Fatalf("created qc = %+v, want pending inspection linked to receipt line", createdQC)
	}

	startQCReq := smokeRequestAsRole(
		httptest.NewRequest(http.MethodPost, "/api/v1/inbound-qc-inspections/iqc-e2e-inbound-pass/start", nil),
		authConfig,
		auth.RoleQA,
	)
	startQCReq.SetPathValue("inspection_id", createdQC.Inspection.ID)
	startQCReq.Header.Set(response.HeaderRequestID, "req-e2e-pass-qc-start")
	startQCRec := httptest.NewRecorder()

	inboundQCInspectionStartHandler(services.inboundQC).ServeHTTP(startQCRec, startQCReq)

	if startQCRec.Code != http.StatusOK {
		t.Fatalf("start qc status = %d, want %d: %s", startQCRec.Code, http.StatusOK, startQCRec.Body.String())
	}

	passQCReq := smokeRequestAsRole(
		httptest.NewRequest(http.MethodPost, "/api/v1/inbound-qc-inspections/iqc-e2e-inbound-pass/pass", bytes.NewBufferString(`{
			"checklist": [
				{"id": "check-packaging", "code": "PACKAGING", "label": "Packaging condition", "required": true, "status": "pass"},
				{"id": "check-lot-expiry", "code": "LOT_EXPIRY", "label": "Lot and expiry match delivery", "required": true, "status": "pass"},
				{"id": "check-sample", "code": "SAMPLE", "label": "Sample retained when required", "status": "not_applicable"}
			]
		}`)),
		authConfig,
		auth.RoleQA,
	)
	passQCReq.SetPathValue("inspection_id", createdQC.Inspection.ID)
	passQCReq.Header.Set(response.HeaderRequestID, "req-e2e-pass-qc-pass")
	passQCRec := httptest.NewRecorder()

	inboundQCInspectionPassHandler(services.inboundQC).ServeHTTP(passQCRec, passQCReq)

	if passQCRec.Code != http.StatusOK {
		t.Fatalf("pass qc status = %d, want %d: %s", passQCRec.Code, http.StatusOK, passQCRec.Body.String())
	}
	passedQC := decodeSmokeSuccess[inboundQCActionResultResponse](t, passQCRec).Data
	if passedQC.PreviousStatus != "in_progress" ||
		passedQC.CurrentStatus != "completed" ||
		passedQC.CurrentResult != "pass" ||
		passedQC.Inspection.PassedQuantity != passedQC.Inspection.Quantity ||
		passedQC.Inspection.FailedQuantity != "0.000000" ||
		passedQC.Inspection.HoldQuantity != "0.000000" ||
		passedQC.AuditLogID == "" {
		t.Fatalf("passed qc = %+v, want full pass with audit", passedQC)
	}
	assertInboundPassE2EAvailableStock(t, services.movementStore.Movements(), passedQC.Inspection)

	assertInboundQCE2EAuditAction(t, services.auditStore, "inventory.receiving.created")
	assertInboundQCE2EAuditAction(t, services.auditStore, "inventory.receiving.submitted")
	assertInboundQCE2EAuditAction(t, services.auditStore, "inventory.receiving.inspect_ready")
	assertInboundQCE2EAuditAction(t, services.auditStore, "qc.inbound_inspection.created")
	assertInboundQCE2EAuditAction(t, services.auditStore, "qc.inbound_inspection.started")
	assertInboundQCE2EAuditAction(t, services.auditStore, "qc.inbound_inspection.passed")
	assertInboundQCE2EAuditAction(t, services.auditStore, "qc.inbound_inspection.stock_movement.recorded")
}

func TestInboundPartialReceivingQCE2ESmoke(t *testing.T) {
	authConfig := smokeAuthConfig()
	services := newInboundQCE2EServices(approvedInboundPartialE2EPurchaseOrder())

	createReceiptReq := smokeRequestAsRole(
		httptest.NewRequest(http.MethodPost, "/api/v1/goods-receipts", bytes.NewBufferString(`{
			"id": "grn-e2e-inbound-partial",
			"receipt_no": "GRN-260429-PARTIAL-E2E",
			"warehouse_id": "wh-hcm-fg",
			"location_id": "loc-hcm-fg-recv-01",
			"reference_doc_type": "purchase_order",
			"reference_doc_id": "PO-260429-PARTIAL-E2E",
			"supplier_id": "supplier-local",
			"delivery_note_no": "DN-260429-PARTIAL-E2E",
			"lines": [
				{
					"id": "line-e2e-inbound-partial",
					"purchase_order_line_id": "po-line-260429-partial-e2e-001",
					"batch_id": "batch-serum-2604a",
					"quantity": "10",
					"uom_code": "EA",
					"base_uom_code": "EA",
					"packaging_status": "intact"
				}
			]
		}`)),
		authConfig,
		auth.RoleWarehouseLead,
	)
	createReceiptReq.Header.Set(response.HeaderRequestID, "req-e2e-partial-receiving-create")
	createReceiptRec := httptest.NewRecorder()

	goodsReceiptsHandler(services.receiving).ServeHTTP(createReceiptRec, createReceiptReq)

	if createReceiptRec.Code != http.StatusCreated {
		t.Fatalf("create receipt status = %d, want %d: %s", createReceiptRec.Code, http.StatusCreated, createReceiptRec.Body.String())
	}
	receipt := decodeSmokeSuccess[warehouseReceivingResponse](t, createReceiptRec).Data
	if receipt.Status != "draft" ||
		receipt.Lines[0].ID != "line-e2e-inbound-partial" ||
		receipt.Lines[0].PackagingStatus != "intact" ||
		receipt.AuditLogID == "" {
		t.Fatalf("receipt = %+v, want draft intact inbound line with audit", receipt)
	}

	for _, step := range []struct {
		name    string
		handler http.HandlerFunc
	}{
		{name: "submit", handler: submitGoodsReceiptHandler(services.receiving)},
		{name: "inspect", handler: markGoodsReceiptInspectReadyHandler(services.receiving)},
	} {
		req := smokeRequestAsRole(
			httptest.NewRequest(http.MethodPost, "/api/v1/goods-receipts/"+receipt.ID+"/"+step.name, nil),
			authConfig,
			auth.RoleWarehouseLead,
		)
		req.SetPathValue("receipt_id", receipt.ID)
		req.Header.Set(response.HeaderRequestID, "req-e2e-partial-receiving-"+step.name)
		rec := httptest.NewRecorder()

		step.handler.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("%s receipt status = %d, want %d: %s", step.name, rec.Code, http.StatusOK, rec.Body.String())
		}
	}
	if services.movementStore.Count() != 0 {
		t.Fatalf("stock movements before partial QC decision = %d, want 0", services.movementStore.Count())
	}

	createQCReq := smokeRequestAsRole(
		httptest.NewRequest(http.MethodPost, "/api/v1/inbound-qc-inspections", bytes.NewBufferString(`{
			"id": "iqc-e2e-inbound-partial",
			"goods_receipt_id": "grn-e2e-inbound-partial",
			"goods_receipt_line_id": "line-e2e-inbound-partial",
			"note": "supplier delivered partial quantity, QA holds label sample"
		}`)),
		authConfig,
		auth.RoleQA,
	)
	createQCReq.Header.Set(response.HeaderRequestID, "req-e2e-partial-qc-create")
	createQCRec := httptest.NewRecorder()

	inboundQCInspectionsHandler(services.inboundQC).ServeHTTP(createQCRec, createQCReq)

	if createQCRec.Code != http.StatusCreated {
		t.Fatalf("create qc status = %d, want %d: %s", createQCRec.Code, http.StatusCreated, createQCRec.Body.String())
	}
	createdQC := decodeSmokeSuccess[inboundQCActionResultResponse](t, createQCRec).Data
	if createdQC.Inspection.Status != "pending" ||
		createdQC.Inspection.Quantity != "10.000000" ||
		createdQC.Inspection.GoodsReceiptLineID != receipt.Lines[0].ID {
		t.Fatalf("created qc = %+v, want pending partial inspection linked to receipt line", createdQC)
	}

	startQCReq := smokeRequestAsRole(
		httptest.NewRequest(http.MethodPost, "/api/v1/inbound-qc-inspections/iqc-e2e-inbound-partial/start", nil),
		authConfig,
		auth.RoleQA,
	)
	startQCReq.SetPathValue("inspection_id", createdQC.Inspection.ID)
	startQCReq.Header.Set(response.HeaderRequestID, "req-e2e-partial-qc-start")
	startQCRec := httptest.NewRecorder()

	inboundQCInspectionStartHandler(services.inboundQC).ServeHTTP(startQCRec, startQCReq)

	if startQCRec.Code != http.StatusOK {
		t.Fatalf("start qc status = %d, want %d: %s", startQCRec.Code, http.StatusOK, startQCRec.Body.String())
	}

	partialQCReq := smokeRequestAsRole(
		httptest.NewRequest(http.MethodPost, "/api/v1/inbound-qc-inspections/iqc-e2e-inbound-partial/partial", bytes.NewBufferString(`{
			"passed_qty": "7",
			"failed_qty": "0",
			"hold_qty": "3",
			"reason": "hold label sample for QA retention",
			"checklist": [
				{"id": "check-packaging", "code": "PACKAGING", "label": "Packaging condition", "required": true, "status": "pass"},
				{"id": "check-lot-expiry", "code": "LOT_EXPIRY", "label": "Lot and expiry match delivery", "required": true, "status": "pass"},
				{"id": "check-sample", "code": "SAMPLE", "label": "Sample retained when required", "status": "pass"}
			]
		}`)),
		authConfig,
		auth.RoleQA,
	)
	partialQCReq.SetPathValue("inspection_id", createdQC.Inspection.ID)
	partialQCReq.Header.Set(response.HeaderRequestID, "req-e2e-partial-qc-decision")
	partialQCRec := httptest.NewRecorder()

	inboundQCInspectionPartialHandler(services.inboundQC).ServeHTTP(partialQCRec, partialQCReq)

	if partialQCRec.Code != http.StatusOK {
		t.Fatalf("partial qc status = %d, want %d: %s", partialQCRec.Code, http.StatusOK, partialQCRec.Body.String())
	}
	partialQC := decodeSmokeSuccess[inboundQCActionResultResponse](t, partialQCRec).Data
	if partialQC.PreviousStatus != "in_progress" ||
		partialQC.CurrentStatus != "completed" ||
		partialQC.CurrentResult != "partial" ||
		partialQC.Inspection.PassedQuantity != "7.000000" ||
		partialQC.Inspection.FailedQuantity != "0.000000" ||
		partialQC.Inspection.HoldQuantity != "3.000000" ||
		partialQC.AuditLogID == "" {
		t.Fatalf("partial qc = %+v, want 7 pass and 3 hold with audit", partialQC)
	}
	assertInboundPartialE2EStockSplit(t, services.movementStore.Movements(), partialQC.Inspection)

	assertInboundQCE2EAuditAction(t, services.auditStore, "inventory.receiving.created")
	assertInboundQCE2EAuditAction(t, services.auditStore, "inventory.receiving.submitted")
	assertInboundQCE2EAuditAction(t, services.auditStore, "inventory.receiving.inspect_ready")
	assertInboundQCE2EAuditAction(t, services.auditStore, "qc.inbound_inspection.created")
	assertInboundQCE2EAuditAction(t, services.auditStore, "qc.inbound_inspection.started")
	assertInboundQCE2EAuditAction(t, services.auditStore, "qc.inbound_inspection.partial")
	assertInboundQCE2EAuditActionCount(t, services.auditStore, "qc.inbound_inspection.stock_movement.recorded", 2)
}

type inboundQCE2EServices struct {
	receiving     inventoryapp.WarehouseReceivingService
	inboundQC     qcapp.InboundQCInspectionService
	movementStore *inventoryapp.InMemoryStockMovementStore
	auditStore    *audit.InMemoryLogStore
}

func newInboundQCE2EServices(order purchasedomain.PurchaseOrder) inboundQCE2EServices {
	auditStore := audit.NewInMemoryLogStore()
	receivingStore := inventoryapp.NewPrototypeWarehouseReceivingStore()
	movementStore := inventoryapp.NewInMemoryStockMovementStore()
	batchCatalog := inventoryapp.NewPrototypeBatchCatalog(auditStore)
	receiving := inventoryapp.NewWarehouseReceivingService(
		receivingStore,
		masterdataapp.NewPrototypeWarehouseLocationCatalog(auditStore),
		batchCatalog,
		movementStore,
		auditStore,
	).WithPurchaseOrderReader(testGoodsReceiptPurchaseOrderReader{order: order})
	inboundQC := qcapp.NewInboundQCInspectionService(
		qcapp.NewPrototypeInboundQCInspectionStore(),
		receivingStore,
		auditStore,
	).WithStockMovementRecorder(movementStore).
		WithBatchQCStatusUpdater(inboundQCBatchQCStatusAdapter{catalog: batchCatalog})

	return inboundQCE2EServices{
		receiving:     receiving,
		inboundQC:     inboundQC,
		movementStore: movementStore,
		auditStore:    auditStore,
	}
}

func approvedInboundPassE2EPurchaseOrder() purchasedomain.PurchaseOrder {
	order, err := purchasedomain.NewPurchaseOrderDocument(purchasedomain.NewPurchaseOrderDocumentInput{
		ID:            "PO-260429-PASS-E2E",
		OrgID:         "org-my-pham",
		PONo:          "PO-260429-PASS-E2E",
		SupplierID:    "supplier-local",
		SupplierCode:  "SUP-LOCAL",
		SupplierName:  "Local Supplier",
		WarehouseID:   "wh-hcm-fg",
		WarehouseCode: "WH-HCM-FG",
		ExpectedDate:  "2026-04-29",
		CurrencyCode:  "VND",
		Lines: []purchasedomain.NewPurchaseOrderLineInput{
			{
				ID:           "po-line-260429-pass-e2e-001",
				LineNo:       1,
				ItemID:       "item-serum-30ml",
				SKUCode:      "SERUM-30ML",
				ItemName:     "Vitamin C Serum",
				OrderedQty:   decimal.MustQuantity("12"),
				UOMCode:      "EA",
				BaseUOMCode:  "EA",
				UnitPrice:    decimal.MustUnitPrice("1"),
				CurrencyCode: "VND",
			},
		},
		CreatedAt: time.Date(2026, 4, 29, 8, 0, 0, 0, time.UTC),
		CreatedBy: "user-purchase-ops",
	})
	if err != nil {
		panic(err)
	}
	submitted, err := order.Submit("user-purchase-ops", time.Date(2026, 4, 29, 8, 30, 0, 0, time.UTC))
	if err != nil {
		panic(err)
	}
	approved, err := submitted.Approve("user-purchase-ops", time.Date(2026, 4, 29, 9, 0, 0, 0, time.UTC))
	if err != nil {
		panic(err)
	}

	return approved
}

func approvedInboundPartialE2EPurchaseOrder() purchasedomain.PurchaseOrder {
	order, err := purchasedomain.NewPurchaseOrderDocument(purchasedomain.NewPurchaseOrderDocumentInput{
		ID:            "PO-260429-PARTIAL-E2E",
		OrgID:         "org-my-pham",
		PONo:          "PO-260429-PARTIAL-E2E",
		SupplierID:    "supplier-local",
		SupplierCode:  "SUP-LOCAL",
		SupplierName:  "Local Supplier",
		WarehouseID:   "wh-hcm-fg",
		WarehouseCode: "WH-HCM-FG",
		ExpectedDate:  "2026-04-29",
		CurrencyCode:  "VND",
		Lines: []purchasedomain.NewPurchaseOrderLineInput{
			{
				ID:           "po-line-260429-partial-e2e-001",
				LineNo:       1,
				ItemID:       "item-serum-30ml",
				SKUCode:      "SERUM-30ML",
				ItemName:     "Vitamin C Serum",
				OrderedQty:   decimal.MustQuantity("12"),
				UOMCode:      "EA",
				BaseUOMCode:  "EA",
				UnitPrice:    decimal.MustUnitPrice("1"),
				CurrencyCode: "VND",
			},
		},
		CreatedAt: time.Date(2026, 4, 29, 8, 0, 0, 0, time.UTC),
		CreatedBy: "user-purchase-ops",
	})
	if err != nil {
		panic(err)
	}
	submitted, err := order.Submit("user-purchase-ops", time.Date(2026, 4, 29, 8, 30, 0, 0, time.UTC))
	if err != nil {
		panic(err)
	}
	approved, err := submitted.Approve("user-purchase-ops", time.Date(2026, 4, 29, 9, 0, 0, 0, time.UTC))
	if err != nil {
		panic(err)
	}

	return approved
}

func assertInboundPassE2EAvailableStock(
	t *testing.T,
	movements []inventorydomain.StockMovement,
	inspection inboundQCInspectionResponse,
) {
	t.Helper()

	if len(movements) != 1 {
		t.Fatalf("stock movements = %+v, want one available movement from QC pass", movements)
	}
	movement := movements[0]
	if movement.MovementType != inventorydomain.MovementPurchaseReceipt ||
		movement.StockStatus != inventorydomain.StockStatusAvailable ||
		movement.Quantity.String() != inspection.PassedQuantity ||
		movement.SourceDocType != "inbound_qc_inspection" ||
		movement.SourceDocID != inspection.ID ||
		movement.SourceDocLineID != inspection.GoodsReceiptLineID {
		t.Fatalf("stock movement = %+v, want inbound QC PASS available qty %s", movement, inspection.PassedQuantity)
	}

	expiryDate, err := time.Parse("2006-01-02", inspection.ExpiryDate)
	if err != nil {
		t.Fatalf("parse inspection expiry date %q: %v", inspection.ExpiryDate, err)
	}
	snapshots := inventorydomain.CalculateAvailableStockAt(
		[]inventorydomain.StockBalanceSnapshot{
			{
				WarehouseID:   movement.WarehouseID,
				LocationID:    movement.BinID,
				ItemID:        movement.ItemID,
				SKU:           inspection.SKU,
				BatchID:       movement.BatchID,
				BatchNo:       inspection.BatchNo,
				BatchQCStatus: inventorydomain.QCStatusPass,
				BatchStatus:   inventorydomain.BatchStatusActive,
				BatchExpiry:   expiryDate,
				BaseUOMCode:   decimal.MustUOMCode(inspection.UOMCode),
				StockStatus:   movement.StockStatus,
				QtyOnHand:     movement.Quantity,
				QtyReserved:   decimal.MustQuantity("0"),
			},
		},
		time.Date(2026, 4, 29, 0, 0, 0, 0, time.UTC),
	)
	if len(snapshots) != 1 {
		t.Fatalf("available snapshots = %d, want 1", len(snapshots))
	}
	if snapshots[0].PhysicalQty.String() != inspection.PassedQuantity ||
		snapshots[0].AvailableQty.String() != inspection.PassedQuantity ||
		snapshots[0].HoldQty.String() != "0.000000" {
		t.Fatalf("available snapshot = %+v, want available qty %s with no hold", snapshots[0], inspection.PassedQuantity)
	}
}

func assertInboundPartialE2EStockSplit(
	t *testing.T,
	movements []inventorydomain.StockMovement,
	inspection inboundQCInspectionResponse,
) {
	t.Helper()

	if len(movements) != 2 {
		t.Fatalf("stock movements = %+v, want available and qc hold movements from partial QC", movements)
	}
	available, ok := inboundQCE2EMovementByStatus(movements, inventorydomain.StockStatusAvailable)
	if !ok ||
		available.MovementType != inventorydomain.MovementPurchaseReceipt ||
		available.Quantity.String() != inspection.PassedQuantity ||
		available.SourceDocType != "inbound_qc_inspection" ||
		available.SourceDocID != inspection.ID ||
		available.SourceDocLineID != inspection.GoodsReceiptLineID {
		t.Fatalf("available movement = %+v, want passed qty %s from inbound QC partial", available, inspection.PassedQuantity)
	}
	hold, ok := inboundQCE2EMovementByStatus(movements, inventorydomain.StockStatusQCHold)
	if !ok ||
		hold.MovementType != inventorydomain.MovementPurchaseReceipt ||
		hold.Quantity.String() != inspection.HoldQuantity ||
		hold.SourceDocType != "inbound_qc_inspection" ||
		hold.SourceDocID != inspection.ID ||
		hold.SourceDocLineID != inspection.GoodsReceiptLineID {
		t.Fatalf("hold movement = %+v, want hold qty %s from inbound QC partial", hold, inspection.HoldQuantity)
	}

	expiryDate, err := time.Parse("2006-01-02", inspection.ExpiryDate)
	if err != nil {
		t.Fatalf("parse inspection expiry date %q: %v", inspection.ExpiryDate, err)
	}
	snapshots := inventorydomain.CalculateAvailableStockAt(
		[]inventorydomain.StockBalanceSnapshot{
			inboundQCE2EStockBalanceSnapshot(available, inspection, inventorydomain.QCStatusPass, expiryDate),
			inboundQCE2EStockBalanceSnapshot(hold, inspection, inventorydomain.QCStatusQuarantine, expiryDate),
		},
		time.Date(2026, 4, 29, 0, 0, 0, 0, time.UTC),
	)
	if len(snapshots) != 1 {
		t.Fatalf("available snapshots = %d, want 1", len(snapshots))
	}
	if snapshots[0].PhysicalQty.String() != inspection.Quantity ||
		snapshots[0].AvailableQty.String() != inspection.PassedQuantity ||
		snapshots[0].QCHoldQty.String() != inspection.HoldQuantity ||
		snapshots[0].DamagedQty.String() != "0.000000" {
		t.Fatalf(
			"available snapshot = %+v, want physical %s, available %s, qc hold %s",
			snapshots[0],
			inspection.Quantity,
			inspection.PassedQuantity,
			inspection.HoldQuantity,
		)
	}
}

func inboundQCE2EStockBalanceSnapshot(
	movement inventorydomain.StockMovement,
	inspection inboundQCInspectionResponse,
	qcStatus inventorydomain.QCStatus,
	expiryDate time.Time,
) inventorydomain.StockBalanceSnapshot {
	return inventorydomain.StockBalanceSnapshot{
		WarehouseID:   movement.WarehouseID,
		LocationID:    movement.BinID,
		ItemID:        movement.ItemID,
		SKU:           inspection.SKU,
		BatchID:       movement.BatchID,
		BatchNo:       inspection.BatchNo,
		BatchQCStatus: qcStatus,
		BatchStatus:   inventorydomain.BatchStatusActive,
		BatchExpiry:   expiryDate,
		BaseUOMCode:   decimal.MustUOMCode(inspection.UOMCode),
		StockStatus:   movement.StockStatus,
		QtyOnHand:     movement.Quantity,
		QtyReserved:   decimal.MustQuantity("0"),
	}
}

func inboundQCE2EMovementByStatus(
	movements []inventorydomain.StockMovement,
	status inventorydomain.StockStatus,
) (inventorydomain.StockMovement, bool) {
	for _, movement := range movements {
		if movement.StockStatus == status {
			return movement, true
		}
	}

	return inventorydomain.StockMovement{}, false
}

func assertInboundQCE2EAuditAction(t *testing.T, auditStore audit.LogStore, action string) {
	t.Helper()

	assertInboundQCE2EAuditActionCount(t, auditStore, action, 1)
}

func assertInboundQCE2EAuditActionCount(t *testing.T, auditStore audit.LogStore, action string, want int) {
	t.Helper()

	logs, err := auditStore.List(context.Background(), audit.Query{Action: action})
	if err != nil {
		t.Fatalf("list audit logs for %s: %v", action, err)
	}
	if len(logs) != want {
		t.Fatalf("audit action %s count = %d, want %d", action, len(logs), want)
	}
}
