package main

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	inventoryapp "github.com/Chinsusu/ERP-v2/apps/api/internal/modules/inventory/application"
	"github.com/Chinsusu/ERP-v2/apps/api/internal/shared/audit"
	"github.com/Chinsusu/ERP-v2/apps/api/internal/shared/auth"
	"github.com/Chinsusu/ERP-v2/apps/api/internal/shared/response"
)

func TestSupplierRejectionHandlersCreateSubmitConfirm(t *testing.T) {
	store := inventoryapp.NewPrototypeSupplierRejectionStore()
	auditStore := audit.NewInMemoryLogStore()
	listSupplierRejections := inventoryapp.NewListSupplierRejections(store)
	createSupplierRejection := inventoryapp.NewCreateSupplierRejection(store, auditStore)
	transitionSupplierRejection := inventoryapp.NewTransitionSupplierRejection(store, auditStore)
	principalContext := auth.WithPrincipal(context.Background(), auth.MockPrincipalForRole(auth.MockConfig{
		Email:       "lead@example.local",
		Password:    "local-only-mock-password",
		AccessToken: "local-dev-access-token",
	}, auth.RoleWarehouseLead))

	createReq := httptest.NewRequest(
		http.MethodPost,
		"/api/v1/supplier-rejections",
		bytes.NewBufferString(`{
			"id": "srj-handler-flow",
			"org_id": "org-my-pham",
			"rejection_no": "SRJ-260429-HANDLER",
			"supplier_id": "supplier-local",
			"supplier_code": "SUP-LOCAL",
			"supplier_name": "Local Supplier",
			"purchase_order_id": "po-260427-0003",
			"purchase_order_no": "PO-260427-0003",
			"goods_receipt_id": "grn-hcm-260427-inspect",
			"goods_receipt_no": "GRN-260427-0003",
			"inbound_qc_inspection_id": "iqc-fail-handler",
			"warehouse_id": "wh-hcm-fg",
			"warehouse_code": "WH-HCM-FG",
			"reason": "damaged packaging",
			"lines": [
				{
					"id": "srj-handler-line-001",
					"purchase_order_line_id": "po-line-260427-0003-001",
					"goods_receipt_line_id": "grn-line-draft-001",
					"inbound_qc_inspection_id": "iqc-fail-handler",
					"item_id": "item-serum-30ml",
					"sku": "serum-30ml",
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
					"id": "srj-handler-att-001",
					"line_id": "srj-handler-line-001",
					"file_name": "damage-photo.jpg",
					"object_key": "supplier-rejections/srj-handler-flow/damage-photo.jpg",
					"content_type": "image/jpeg",
					"source": "inbound_qc"
				}
			]
		}`),
	).WithContext(principalContext)
	createReq.Header.Set(response.HeaderRequestID, "req-srj-create")
	createRec := httptest.NewRecorder()

	supplierRejectionsHandler(listSupplierRejections, createSupplierRejection).ServeHTTP(createRec, createReq)

	if createRec.Code != http.StatusCreated {
		t.Fatalf("create status = %d, want %d: %s", createRec.Code, http.StatusCreated, createRec.Body.String())
	}
	var createPayload response.SuccessEnvelope[supplierRejectionResponse]
	if err := json.NewDecoder(createRec.Body).Decode(&createPayload); err != nil {
		t.Fatalf("decode create response: %v", err)
	}
	if createPayload.Data.Status != "draft" ||
		createPayload.Data.RejectionNo != "SRJ-260429-HANDLER" ||
		createPayload.Data.Lines[0].RejectedQuantity != "6.000000" ||
		createPayload.Data.AuditLogID == "" {
		t.Fatalf("create payload = %+v, want audited draft supplier rejection", createPayload.Data)
	}

	submitReq := httptest.NewRequest(http.MethodPost, "/api/v1/supplier-rejections/srj-handler-flow/submit", nil).
		WithContext(principalContext)
	submitReq.SetPathValue("supplier_rejection_id", "srj-handler-flow")
	submitReq.Header.Set(response.HeaderRequestID, "req-srj-submit")
	submitRec := httptest.NewRecorder()
	supplierRejectionActionHandler(transitionSupplierRejection, "submit").ServeHTTP(submitRec, submitReq)
	if submitRec.Code != http.StatusOK {
		t.Fatalf("submit status = %d, want %d: %s", submitRec.Code, http.StatusOK, submitRec.Body.String())
	}
	var submitPayload response.SuccessEnvelope[supplierRejectionActionResultResponse]
	if err := json.NewDecoder(submitRec.Body).Decode(&submitPayload); err != nil {
		t.Fatalf("decode submit response: %v", err)
	}
	if submitPayload.Data.PreviousStatus != "draft" ||
		submitPayload.Data.CurrentStatus != "submitted" ||
		submitPayload.Data.AuditLogID == "" {
		t.Fatalf("submit payload = %+v, want audited draft -> submitted", submitPayload.Data)
	}

	confirmReq := httptest.NewRequest(http.MethodPost, "/api/v1/supplier-rejections/srj-handler-flow/confirm", nil).
		WithContext(principalContext)
	confirmReq.SetPathValue("supplier_rejection_id", "srj-handler-flow")
	confirmReq.Header.Set(response.HeaderRequestID, "req-srj-confirm")
	confirmRec := httptest.NewRecorder()
	supplierRejectionActionHandler(transitionSupplierRejection, "confirm").ServeHTTP(confirmRec, confirmReq)
	if confirmRec.Code != http.StatusOK {
		t.Fatalf("confirm status = %d, want %d: %s", confirmRec.Code, http.StatusOK, confirmRec.Body.String())
	}
	var confirmPayload response.SuccessEnvelope[supplierRejectionActionResultResponse]
	if err := json.NewDecoder(confirmRec.Body).Decode(&confirmPayload); err != nil {
		t.Fatalf("decode confirm response: %v", err)
	}
	if confirmPayload.Data.PreviousStatus != "submitted" ||
		confirmPayload.Data.CurrentStatus != "confirmed" ||
		confirmPayload.Data.Rejection.Status != "confirmed" ||
		confirmPayload.Data.AuditLogID == "" {
		t.Fatalf("confirm payload = %+v, want audited submitted -> confirmed", confirmPayload.Data)
	}

	listReq := httptest.NewRequest(http.MethodGet, "/api/v1/supplier-rejections?status=confirmed", nil).
		WithContext(principalContext)
	listRec := httptest.NewRecorder()
	supplierRejectionsHandler(listSupplierRejections, createSupplierRejection).ServeHTTP(listRec, listReq)
	if listRec.Code != http.StatusOK {
		t.Fatalf("list status = %d, want %d: %s", listRec.Code, http.StatusOK, listRec.Body.String())
	}
	var listPayload response.SuccessEnvelope[[]supplierRejectionResponse]
	if err := json.NewDecoder(listRec.Body).Decode(&listPayload); err != nil {
		t.Fatalf("decode list response: %v", err)
	}
	if len(listPayload.Data) != 1 || listPayload.Data[0].ID != "srj-handler-flow" {
		t.Fatalf("list payload = %+v, want confirmed supplier rejection", listPayload.Data)
	}

	logs, err := auditStore.List(context.Background(), audit.Query{EntityID: "srj-handler-flow"})
	if err != nil {
		t.Fatalf("list audit logs: %v", err)
	}
	if len(logs) != 3 {
		t.Fatalf("audit logs = %d, want create, submit, confirm", len(logs))
	}
}

func TestSupplierRejectionActionRequiresRecordCreatePermission(t *testing.T) {
	store := inventoryapp.NewPrototypeSupplierRejectionStore()
	transitionSupplierRejection := inventoryapp.NewTransitionSupplierRejection(store, audit.NewInMemoryLogStore())
	req := httptest.NewRequest(http.MethodPost, "/api/v1/supplier-rejections/srj-handler-flow/confirm", nil)
	req.SetPathValue("supplier_rejection_id", "srj-handler-flow")
	req = req.WithContext(auth.WithPrincipal(req.Context(), auth.MockPrincipalForRole(auth.MockConfig{
		Email:       "staff@example.local",
		Password:    "local-only-mock-password",
		AccessToken: "local-dev-access-token",
	}, auth.RoleWarehouseStaff)))
	rec := httptest.NewRecorder()

	supplierRejectionActionHandler(transitionSupplierRejection, "confirm").ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusForbidden)
	}
}
