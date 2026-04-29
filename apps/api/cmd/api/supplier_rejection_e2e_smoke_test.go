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

func TestSupplierRejectionFailReceivingE2ESmoke(t *testing.T) {
	authConfig := smokeAuthConfig()
	services := newSupplierRejectionE2EServices()

	createReceiptReq := smokeRequestAsRole(
		httptest.NewRequest(http.MethodPost, "/api/v1/goods-receipts", bytes.NewBufferString(`{
			"id": "grn-e2e-supplier-reject",
			"receipt_no": "GRN-260429-SRJ-E2E",
			"warehouse_id": "wh-hcm-fg",
			"location_id": "loc-hcm-fg-recv-01",
			"reference_doc_type": "purchase_order",
			"reference_doc_id": "PO-260429-SRJ-E2E",
			"supplier_id": "supplier-local",
			"delivery_note_no": "dn-260429-srj-e2e",
			"lines": [
				{
					"id": "line-e2e-supplier-reject",
					"purchase_order_line_id": "po-line-260429-srj-e2e-001",
					"batch_id": "batch-serum-2604a",
					"quantity": "6",
					"uom_code": "EA",
					"base_uom_code": "EA",
					"packaging_status": "damaged"
				}
			]
		}`)),
		authConfig,
		auth.RoleWarehouseLead,
	)
	createReceiptReq.Header.Set(response.HeaderRequestID, "req-e2e-srj-receiving-create")
	createReceiptRec := httptest.NewRecorder()

	goodsReceiptsHandler(services.receiving).ServeHTTP(createReceiptRec, createReceiptReq)

	if createReceiptRec.Code != http.StatusCreated {
		t.Fatalf("create receipt status = %d, want %d: %s", createReceiptRec.Code, http.StatusCreated, createReceiptRec.Body.String())
	}
	receipt := decodeSmokeSuccess[warehouseReceivingResponse](t, createReceiptRec).Data
	if receipt.Status != "draft" ||
		receipt.Lines[0].ID != "line-e2e-supplier-reject" ||
		receipt.Lines[0].PackagingStatus != "damaged" ||
		receipt.AuditLogID == "" {
		t.Fatalf("receipt = %+v, want draft damaged inbound line with audit", receipt)
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
		req.Header.Set(response.HeaderRequestID, "req-e2e-srj-receiving-"+step.name)
		rec := httptest.NewRecorder()

		step.handler.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("%s receipt status = %d, want %d: %s", step.name, rec.Code, http.StatusOK, rec.Body.String())
		}
	}

	createQCReq := smokeRequestAsRole(
		httptest.NewRequest(http.MethodPost, "/api/v1/inbound-qc-inspections", bytes.NewBufferString(`{
			"id": "iqc-e2e-supplier-reject",
			"goods_receipt_id": "grn-e2e-supplier-reject",
			"goods_receipt_line_id": "line-e2e-supplier-reject",
			"note": "packaging damaged during inbound check"
		}`)),
		authConfig,
		auth.RoleQA,
	)
	createQCReq.Header.Set(response.HeaderRequestID, "req-e2e-srj-qc-create")
	createQCRec := httptest.NewRecorder()

	inboundQCInspectionsHandler(services.inboundQC).ServeHTTP(createQCRec, createQCReq)

	if createQCRec.Code != http.StatusCreated {
		t.Fatalf("create qc status = %d, want %d: %s", createQCRec.Code, http.StatusCreated, createQCRec.Body.String())
	}
	createdQC := decodeSmokeSuccess[inboundQCActionResultResponse](t, createQCRec).Data
	if createdQC.Inspection.Status != "pending" || createdQC.Inspection.GoodsReceiptID != receipt.ID {
		t.Fatalf("created qc = %+v, want pending inspection linked to receipt", createdQC)
	}

	startQCReq := smokeRequestAsRole(
		httptest.NewRequest(http.MethodPost, "/api/v1/inbound-qc-inspections/iqc-e2e-supplier-reject/start", nil),
		authConfig,
		auth.RoleQA,
	)
	startQCReq.SetPathValue("inspection_id", createdQC.Inspection.ID)
	startQCReq.Header.Set(response.HeaderRequestID, "req-e2e-srj-qc-start")
	startQCRec := httptest.NewRecorder()

	inboundQCInspectionStartHandler(services.inboundQC).ServeHTTP(startQCRec, startQCReq)

	if startQCRec.Code != http.StatusOK {
		t.Fatalf("start qc status = %d, want %d: %s", startQCRec.Code, http.StatusOK, startQCRec.Body.String())
	}

	failQCReq := smokeRequestAsRole(
		httptest.NewRequest(http.MethodPost, "/api/v1/inbound-qc-inspections/iqc-e2e-supplier-reject/fail", bytes.NewBufferString(`{
			"reason": "damaged packaging",
			"checklist": [
				{"id": "check-packaging", "code": "PACKAGING", "label": "Packaging condition", "required": true, "status": "fail"},
				{"id": "check-lot-expiry", "code": "LOT_EXPIRY", "label": "Lot and expiry match delivery", "required": true, "status": "pass"},
				{"id": "check-sample", "code": "SAMPLE", "label": "Sample retained when required", "status": "not_applicable"}
			]
		}`)),
		authConfig,
		auth.RoleQA,
	)
	failQCReq.SetPathValue("inspection_id", createdQC.Inspection.ID)
	failQCReq.Header.Set(response.HeaderRequestID, "req-e2e-srj-qc-fail")
	failQCRec := httptest.NewRecorder()

	inboundQCInspectionFailHandler(services.inboundQC).ServeHTTP(failQCRec, failQCReq)

	if failQCRec.Code != http.StatusOK {
		t.Fatalf("fail qc status = %d, want %d: %s", failQCRec.Code, http.StatusOK, failQCRec.Body.String())
	}
	failedQC := decodeSmokeSuccess[inboundQCActionResultResponse](t, failQCRec).Data
	if failedQC.CurrentResult != "fail" ||
		failedQC.Inspection.FailedQuantity != failedQC.Inspection.Quantity ||
		failedQC.AuditLogID == "" {
		t.Fatalf("failed qc = %+v, want full fail with audit", failedQC)
	}
	assertSupplierRejectionE2ENoAvailableMovement(t, services.movementStore.Movements(), failedQC.Inspection.FailedQuantity)

	createRejectionReq := smokeRequestAsRole(
		httptest.NewRequest(http.MethodPost, "/api/v1/supplier-rejections", bytes.NewBufferString(`{
			"id": "srj-e2e-supplier-reject",
			"org_id": "org-my-pham",
			"rejection_no": "SRJ-260429-E2E",
			"supplier_id": "supplier-local",
			"supplier_code": "SUP-LOCAL",
			"supplier_name": "Local Supplier",
			"purchase_order_id": "PO-260429-SRJ-E2E",
			"purchase_order_no": "PO-260429-SRJ-E2E",
			"goods_receipt_id": "grn-e2e-supplier-reject",
			"goods_receipt_no": "GRN-260429-SRJ-E2E",
			"inbound_qc_inspection_id": "iqc-e2e-supplier-reject",
			"warehouse_id": "wh-hcm-fg",
			"warehouse_code": "WH-HCM-FG",
			"reason": "damaged packaging",
			"lines": [
				{
					"id": "srj-e2e-line-001",
					"purchase_order_line_id": "po-line-260429-srj-e2e-001",
					"goods_receipt_line_id": "line-e2e-supplier-reject",
					"inbound_qc_inspection_id": "iqc-e2e-supplier-reject",
					"item_id": "item-serum-30ml",
					"sku": "SERUM-30ML",
					"item_name": "Vitamin C Serum",
					"batch_id": "batch-serum-2604a",
					"batch_no": "LOT-2604A",
					"lot_no": "LOT-2604A",
					"expiry_date": "2027-04-01",
					"rejected_qty": "6.000000",
					"uom_code": "EA",
					"base_uom_code": "EA",
					"reason": "damaged packaging"
				}
			],
			"attachments": [
				{
					"id": "srj-e2e-att-001",
					"line_id": "srj-e2e-line-001",
					"file_name": "damage-photo.jpg",
					"object_key": "supplier-rejections/srj-e2e-supplier-reject/damage-photo.jpg",
					"content_type": "image/jpeg",
					"source": "inbound_qc"
				}
			]
		}`)),
		authConfig,
		auth.RoleWarehouseLead,
	)
	createRejectionReq.Header.Set(response.HeaderRequestID, "req-e2e-srj-create")
	createRejectionRec := httptest.NewRecorder()

	supplierRejectionsHandler(services.listRejections, services.createRejection).ServeHTTP(createRejectionRec, createRejectionReq)

	if createRejectionRec.Code != http.StatusCreated {
		t.Fatalf("create supplier rejection status = %d, want %d: %s", createRejectionRec.Code, http.StatusCreated, createRejectionRec.Body.String())
	}
	rejection := decodeSmokeSuccess[supplierRejectionResponse](t, createRejectionRec).Data
	if rejection.Status != "draft" ||
		rejection.InboundQCInspectionID != failedQC.Inspection.ID ||
		rejection.Lines[0].RejectedQuantity != failedQC.Inspection.FailedQuantity ||
		rejection.AuditLogID == "" {
		t.Fatalf("rejection = %+v, want audited draft rejection linked to failed QC", rejection)
	}

	for _, step := range []struct {
		name       string
		wantStatus string
	}{
		{name: "submit", wantStatus: "submitted"},
		{name: "confirm", wantStatus: "confirmed"},
	} {
		req := smokeRequestAsRole(
			httptest.NewRequest(http.MethodPost, "/api/v1/supplier-rejections/"+rejection.ID+"/"+step.name, nil),
			authConfig,
			auth.RoleWarehouseLead,
		)
		req.SetPathValue("supplier_rejection_id", rejection.ID)
		req.Header.Set(response.HeaderRequestID, "req-e2e-srj-"+step.name)
		rec := httptest.NewRecorder()

		supplierRejectionActionHandler(services.transitionRejection, step.name).ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("%s supplier rejection status = %d, want %d: %s", step.name, rec.Code, http.StatusOK, rec.Body.String())
		}
		result := decodeSmokeSuccess[supplierRejectionActionResultResponse](t, rec).Data
		if result.CurrentStatus != step.wantStatus || result.AuditLogID == "" {
			t.Fatalf("%s result = %+v, want %s with audit", step.name, result, step.wantStatus)
		}
	}

	if services.movementStore.Count() != 1 {
		t.Fatalf("stock movement count = %d, want only QC damaged movement and no supplier-rejection movement", services.movementStore.Count())
	}
	assertSupplierRejectionE2EAuditAction(t, services.auditStore, "inventory.receiving.created")
	assertSupplierRejectionE2EAuditAction(t, services.auditStore, "inventory.receiving.submitted")
	assertSupplierRejectionE2EAuditAction(t, services.auditStore, "inventory.receiving.inspect_ready")
	assertSupplierRejectionE2EAuditAction(t, services.auditStore, "qc.inbound_inspection.failed")
	assertSupplierRejectionE2EAuditAction(t, services.auditStore, "qc.inbound_inspection.stock_movement.recorded")
	assertSupplierRejectionE2EAuditAction(t, services.auditStore, "inventory.supplier_rejection.created")
	assertSupplierRejectionE2EAuditAction(t, services.auditStore, "inventory.supplier_rejection.submitted")
	assertSupplierRejectionE2EAuditAction(t, services.auditStore, "inventory.supplier_rejection.confirmed")
}

type supplierRejectionE2EServices struct {
	receiving           inventoryapp.WarehouseReceivingService
	inboundQC           qcapp.InboundQCInspectionService
	listRejections      inventoryapp.ListSupplierRejections
	createRejection     inventoryapp.CreateSupplierRejection
	transitionRejection inventoryapp.TransitionSupplierRejection
	movementStore       *inventoryapp.InMemoryStockMovementStore
	auditStore          *audit.InMemoryLogStore
}

func newSupplierRejectionE2EServices() supplierRejectionE2EServices {
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
	).WithPurchaseOrderReader(testGoodsReceiptPurchaseOrderReader{order: approvedSupplierRejectionE2EPurchaseOrder()})
	inboundQC := qcapp.NewInboundQCInspectionService(
		qcapp.NewPrototypeInboundQCInspectionStore(),
		receivingStore,
		auditStore,
	).WithStockMovementRecorder(movementStore).
		WithBatchQCStatusUpdater(inboundQCBatchQCStatusAdapter{catalog: batchCatalog})
	rejectionStore := inventoryapp.NewPrototypeSupplierRejectionStore()

	return supplierRejectionE2EServices{
		receiving:           receiving,
		inboundQC:           inboundQC,
		listRejections:      inventoryapp.NewListSupplierRejections(rejectionStore),
		createRejection:     inventoryapp.NewCreateSupplierRejection(rejectionStore, auditStore),
		transitionRejection: inventoryapp.NewTransitionSupplierRejection(rejectionStore, auditStore),
		movementStore:       movementStore,
		auditStore:          auditStore,
	}
}

func approvedSupplierRejectionE2EPurchaseOrder() purchasedomain.PurchaseOrder {
	order, err := purchasedomain.NewPurchaseOrderDocument(purchasedomain.NewPurchaseOrderDocumentInput{
		ID:            "PO-260429-SRJ-E2E",
		OrgID:         "org-my-pham",
		PONo:          "PO-260429-SRJ-E2E",
		SupplierID:    "supplier-local",
		SupplierCode:  "SUP-LOCAL",
		SupplierName:  "Local Supplier",
		WarehouseID:   "wh-hcm-fg",
		WarehouseCode: "WH-HCM-FG",
		ExpectedDate:  "2026-04-29",
		CurrencyCode:  "VND",
		Lines: []purchasedomain.NewPurchaseOrderLineInput{
			{
				ID:           "po-line-260429-srj-e2e-001",
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

func assertSupplierRejectionE2ENoAvailableMovement(
	t *testing.T,
	movements []inventorydomain.StockMovement,
	failedQuantity string,
) {
	t.Helper()

	if len(movements) != 1 {
		t.Fatalf("stock movements = %+v, want one damaged movement from QC fail", movements)
	}
	if movements[0].StockStatus != inventorydomain.StockStatusDamaged ||
		movements[0].Quantity.String() != failedQuantity {
		t.Fatalf("stock movement = %+v, want damaged qty %s and no available stock", movements[0], failedQuantity)
	}
	if movements[0].StockStatus == inventorydomain.StockStatusAvailable {
		t.Fatalf("stock movement = %+v, must not make QC FAIL goods available", movements[0])
	}
}

func assertSupplierRejectionE2EAuditAction(t *testing.T, auditStore audit.LogStore, action string) {
	t.Helper()

	logs, err := auditStore.List(context.Background(), audit.Query{Action: action})
	if err != nil {
		t.Fatalf("list audit logs for %s: %v", action, err)
	}
	if len(logs) != 1 {
		t.Fatalf("audit action %s count = %d, want 1", action, len(logs))
	}
}
