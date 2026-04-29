package main

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	inventoryapp "github.com/Chinsusu/ERP-v2/apps/api/internal/modules/inventory/application"
	inventorydomain "github.com/Chinsusu/ERP-v2/apps/api/internal/modules/inventory/domain"
	returnsapp "github.com/Chinsusu/ERP-v2/apps/api/internal/modules/returns/application"
	returnsdomain "github.com/Chinsusu/ERP-v2/apps/api/internal/modules/returns/domain"
	"github.com/Chinsusu/ERP-v2/apps/api/internal/shared/audit"
	"github.com/Chinsusu/ERP-v2/apps/api/internal/shared/auth"
	"github.com/Chinsusu/ERP-v2/apps/api/internal/shared/response"
)

func TestReturnReceiptReusableHappyPathSmoke(t *testing.T) {
	authConfig := smokeAuthConfig()
	ctx := context.Background()
	auditStore := audit.NewInMemoryLogStore()
	returnStore := returnsapp.NewPrototypeReturnReceiptStore()
	movementStore := inventoryapp.NewInMemoryStockMovementStore()
	receiveService := returnsapp.NewReceiveReturn(returnStore, auditStore)
	inspectService := returnsapp.NewInspectReturn(returnStore, auditStore)
	dispositionService := returnsapp.NewApplyReturnDisposition(returnStore, movementStore, auditStore)

	scanReq := smokeRequestAsRole(
		httptest.NewRequest(http.MethodPost, "/api/v1/returns/scan", bytes.NewBufferString(`{
			"warehouse_id": "wh-hcm",
			"warehouse_code": "HCM",
			"source": "CARRIER",
			"code": "GHN260426001",
			"package_condition": "sealed bag"
		}`)),
		authConfig,
		auth.RoleWarehouseLead,
	)
	scanReq.Header.Set(response.HeaderRequestID, "req-e2e-return-scan")
	scanRec := httptest.NewRecorder()

	returnScanHandler(receiveService).ServeHTTP(scanRec, scanReq)

	if scanRec.Code != http.StatusCreated {
		t.Fatalf("scan status = %d, want %d: %s", scanRec.Code, http.StatusCreated, scanRec.Body.String())
	}
	received := decodeSmokeSuccess[returnReceiptResponse](t, scanRec).Data
	if received.OriginalOrderNo != "SO-260426-001" ||
		received.TrackingNo != "GHN260426001" ||
		received.Status != string(returnsdomain.ReturnStatusPendingInspection) ||
		received.UnknownCase {
		t.Fatalf("received = %+v, want known handed-over order pending inspection", received)
	}
	if received.AuditLogID == "" {
		t.Fatal("return receipt audit log id is empty")
	}

	inspectReq := smokeRequestAsRole(
		httptest.NewRequest(
			http.MethodPost,
			"/api/v1/returns/"+received.ID+"/inspect",
			bytes.NewBufferString(`{
				"condition": "intact",
				"disposition": "reusable",
				"note": "seal intact",
				"evidence_label": "photo-return-e2e-001"
			}`),
		),
		authConfig,
		auth.RoleWarehouseLead,
	)
	inspectReq.SetPathValue("return_receipt_id", received.ID)
	inspectReq.Header.Set(response.HeaderRequestID, "req-e2e-return-inspect")
	inspectRec := httptest.NewRecorder()

	returnInspectionHandler(inspectService).ServeHTTP(inspectRec, inspectReq)

	if inspectRec.Code != http.StatusOK {
		t.Fatalf("inspect status = %d, want %d: %s", inspectRec.Code, http.StatusOK, inspectRec.Body.String())
	}
	inspection := decodeSmokeSuccess[returnInspectionResponse](t, inspectRec).Data
	if inspection.ReceiptID != received.ID ||
		inspection.Status != "inspection_recorded" ||
		inspection.TargetLocation != "return-area-qc-release" ||
		inspection.AuditLogID == "" {
		t.Fatalf("inspection = %+v, want reusable inspection with audit", inspection)
	}

	dispositionReq := smokeRequestAsRole(
		httptest.NewRequest(
			http.MethodPost,
			"/api/v1/returns/"+received.ID+"/disposition",
			bytes.NewBufferString(`{
				"disposition": "reusable",
				"note": "ready for putaway"
			}`),
		),
		authConfig,
		auth.RoleWarehouseLead,
	)
	dispositionReq.SetPathValue("return_receipt_id", received.ID)
	dispositionReq.Header.Set(response.HeaderRequestID, "req-e2e-return-disposition")
	dispositionRec := httptest.NewRecorder()

	returnDispositionHandler(dispositionService).ServeHTTP(dispositionRec, dispositionReq)

	if dispositionRec.Code != http.StatusOK {
		t.Fatalf("disposition status = %d, want %d: %s", dispositionRec.Code, http.StatusOK, dispositionRec.Body.String())
	}
	disposition := decodeSmokeSuccess[returnDispositionActionResponse](t, dispositionRec).Data
	if disposition.ReceiptID != received.ID ||
		disposition.ActionCode != "route_to_putaway" ||
		disposition.TargetStockStatus != "return_pending" ||
		disposition.AuditLogID == "" {
		t.Fatalf("disposition = %+v, want reusable putaway action with audit", disposition)
	}

	rows, err := returnsapp.NewListReturnReceipts(returnStore).Execute(
		ctx,
		returnsdomain.NewReturnReceiptFilter("wh-hcm", returnsdomain.ReturnStatusDispositioned),
	)
	if err != nil {
		t.Fatalf("list dispositioned returns: %v", err)
	}
	var finalReceipt returnsdomain.ReturnReceipt
	for _, row := range rows {
		if row.ID == received.ID {
			finalReceipt = row
			break
		}
	}
	if finalReceipt.ID == "" ||
		finalReceipt.StockMovement == nil ||
		finalReceipt.StockMovement.MovementType != string(inventorydomain.MovementReturnRestock) ||
		finalReceipt.StockMovement.TargetStockStatus != string(inventorydomain.StockStatusAvailable) {
		t.Fatalf("final receipt = %+v, want dispositioned receipt with reusable stock movement", finalReceipt)
	}
	if movementStore.Count() != 1 {
		t.Fatalf("stock movement count = %d, want 1", movementStore.Count())
	}
	movement := movementStore.Movements()[0]
	if movement.MovementType != inventorydomain.MovementReturnRestock ||
		movement.StockStatus != inventorydomain.StockStatusAvailable ||
		movement.SourceDocID != received.ID {
		t.Fatalf("movement = %+v, want available return restock for receipt %s", movement, received.ID)
	}

	assertReturnE2EAuditAction(t, auditStore, "returns.receipt.created")
	assertReturnE2EAuditAction(t, auditStore, "returns.receipt.inspected")
	assertReturnE2EAuditAction(t, auditStore, "returns.inspection.disposition")
	assertReturnE2EAuditAction(t, auditStore, "returns.stock_movement.recorded")
}

func assertReturnE2EAuditAction(t *testing.T, auditStore audit.LogStore, action string) {
	t.Helper()

	logs, err := auditStore.List(context.Background(), audit.Query{Action: action})
	if err != nil {
		t.Fatalf("list audit logs for %s: %v", action, err)
	}
	if len(logs) != 1 {
		t.Fatalf("audit action %s count = %d, want 1", action, len(logs))
	}
}
