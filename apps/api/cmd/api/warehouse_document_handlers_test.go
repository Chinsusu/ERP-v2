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

func TestStockTransferHandlersCreateAndPost(t *testing.T) {
	store := inventoryapp.NewPrototypeStockTransferStore()
	movementStore := inventoryapp.NewInMemoryStockMovementStore()
	auditStore := audit.NewInMemoryLogStore()
	service := inventoryapp.NewStockTransferService(store, movementStore, auditStore)
	ctx := auth.WithPrincipal(context.Background(), auth.MockPrincipalForRole(auth.MockConfig{
		Email:       "warehouse@example.local",
		Password:    "local-only-mock-password",
		AccessToken: "local-dev-access-token",
	}, auth.RoleWarehouseLead))

	createReq := httptest.NewRequest(http.MethodPost, "/api/v1/stock-transfers", bytes.NewBufferString(`{
		"id": "transfer-api-0001",
		"transfer_no": "ST-API-0001",
		"source_warehouse_id": "wh-main",
		"source_warehouse_code": "MAIN",
		"destination_warehouse_id": "wh-stage",
		"destination_warehouse_code": "STAGE",
		"reason_code": "staging",
		"lines": [
			{"id": "transfer-api-line", "sku": "SERUM-30ML", "item_id": "item-serum-30ml", "quantity": "5", "base_uom_code": "PCS"}
		]
	}`)).WithContext(ctx)
	createRec := httptest.NewRecorder()
	stockTransfersHandler(service).ServeHTTP(createRec, createReq)
	if createRec.Code != http.StatusCreated {
		t.Fatalf("create status = %d, want %d: %s", createRec.Code, http.StatusCreated, createRec.Body.String())
	}
	var createPayload response.SuccessEnvelope[stockTransferResponse]
	if err := json.NewDecoder(createRec.Body).Decode(&createPayload); err != nil {
		t.Fatalf("decode create response: %v", err)
	}
	if createPayload.Data.Status != "draft" || createPayload.Data.AuditLogID == "" {
		t.Fatalf("create payload = %+v, want draft transfer with audit", createPayload.Data)
	}

	for _, action := range []string{"submit", "approve", "post"} {
		req := httptest.NewRequest(http.MethodPost, "/api/v1/stock-transfers/transfer-api-0001/"+action, nil).WithContext(ctx)
		req.SetPathValue("stock_transfer_id", "transfer-api-0001")
		rec := httptest.NewRecorder()
		stockTransferActionHandler(service, action).ServeHTTP(rec, req)
		if rec.Code != http.StatusOK {
			t.Fatalf("%s status = %d, want %d: %s", action, rec.Code, http.StatusOK, rec.Body.String())
		}
	}
	if movementStore.Count() != 2 {
		t.Fatalf("movement count = %d, want 2", movementStore.Count())
	}
}

func TestWarehouseIssueHandlersCreateAndPost(t *testing.T) {
	store := inventoryapp.NewPrototypeWarehouseIssueStore()
	movementStore := inventoryapp.NewInMemoryStockMovementStore()
	auditStore := audit.NewInMemoryLogStore()
	service := inventoryapp.NewWarehouseIssueService(store, movementStore, auditStore)
	ctx := auth.WithPrincipal(context.Background(), auth.MockPrincipalForRole(auth.MockConfig{
		Email:       "warehouse@example.local",
		Password:    "local-only-mock-password",
		AccessToken: "local-dev-access-token",
	}, auth.RoleWarehouseLead))

	createReq := httptest.NewRequest(http.MethodPost, "/api/v1/warehouse-issues", bytes.NewBufferString(`{
		"id": "issue-api-0001",
		"issue_no": "WI-API-0001",
		"warehouse_id": "wh-main",
		"warehouse_code": "MAIN",
		"destination_type": "factory",
		"destination_name": "Factory A",
		"reason_code": "production_plan_issue",
		"lines": [
			{"id": "issue-api-line", "sku": "ACI_BHA", "item_id": "item-aci-bha", "item_name": "ACID SALICYLIC", "quantity": "0.125", "base_uom_code": "KG", "source_document_type": "production_plan", "source_document_id": "plan-0001"}
		]
	}`)).WithContext(ctx)
	createRec := httptest.NewRecorder()
	warehouseIssuesHandler(service).ServeHTTP(createRec, createReq)
	if createRec.Code != http.StatusCreated {
		t.Fatalf("create status = %d, want %d: %s", createRec.Code, http.StatusCreated, createRec.Body.String())
	}
	var createPayload response.SuccessEnvelope[warehouseIssueResponse]
	if err := json.NewDecoder(createRec.Body).Decode(&createPayload); err != nil {
		t.Fatalf("decode create response: %v", err)
	}
	if createPayload.Data.Status != "draft" || createPayload.Data.AuditLogID == "" {
		t.Fatalf("create payload = %+v, want draft issue with audit", createPayload.Data)
	}

	for _, action := range []string{"submit", "approve", "post"} {
		req := httptest.NewRequest(http.MethodPost, "/api/v1/warehouse-issues/issue-api-0001/"+action, nil).WithContext(ctx)
		req.SetPathValue("warehouse_issue_id", "issue-api-0001")
		rec := httptest.NewRecorder()
		warehouseIssueActionHandler(service, action).ServeHTTP(rec, req)
		if rec.Code != http.StatusOK {
			t.Fatalf("%s status = %d, want %d: %s", action, rec.Code, http.StatusOK, rec.Body.String())
		}
	}
	if movementStore.Count() != 1 {
		t.Fatalf("movement count = %d, want 1", movementStore.Count())
	}
}
