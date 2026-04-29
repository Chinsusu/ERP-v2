package main

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	inventoryapp "github.com/Chinsusu/ERP-v2/apps/api/internal/modules/inventory/application"
	inventorydomain "github.com/Chinsusu/ERP-v2/apps/api/internal/modules/inventory/domain"
	"github.com/Chinsusu/ERP-v2/apps/api/internal/shared/audit"
	"github.com/Chinsusu/ERP-v2/apps/api/internal/shared/auth"
	"github.com/Chinsusu/ERP-v2/apps/api/internal/shared/response"
)

func TestShiftClosingCleanStockCountHappyPathSmoke(t *testing.T) {
	authConfig := smokeAuthConfig()
	auditStore := audit.NewInMemoryLogStore()
	stockCountStore := inventoryapp.NewPrototypeStockCountStore()
	reconciliationStore := inventoryapp.NewPrototypeEndOfDayReconciliationStore()
	listStockCounts := inventoryapp.NewListStockCounts(stockCountStore)
	createStockCount := inventoryapp.NewCreateStockCount(stockCountStore, auditStore)
	submitStockCount := inventoryapp.NewSubmitStockCount(stockCountStore, auditStore)
	listReconciliations := inventoryapp.NewListEndOfDayReconciliations(reconciliationStore)
	closeReconciliation := inventoryapp.NewCloseEndOfDayReconciliation(reconciliationStore, auditStore)

	boardReq := smokeRequestAsRole(
		httptest.NewRequest(
			http.MethodGet,
			"/api/v1/warehouse/end-of-day-reconciliations?warehouse_id=wh-hn&date=2026-04-26&shift_code=day&status=open",
			nil,
		),
		authConfig,
		auth.RoleWarehouseLead,
	)
	boardReq.Header.Set(response.HeaderRequestID, "req-e2e-shift-board")
	boardRec := httptest.NewRecorder()

	endOfDayReconciliationsHandler(listReconciliations).ServeHTTP(boardRec, boardReq)

	if boardRec.Code != http.StatusOK {
		t.Fatalf("board status = %d, want %d: %s", boardRec.Code, http.StatusOK, boardRec.Body.String())
	}
	board := decodeSmokeSuccess[[]endOfDayReconciliationResponse](t, boardRec).Data
	if len(board) != 1 ||
		board[0].ID != "rec-hn-260426-day" ||
		!board[0].Summary.ReadyToClose ||
		board[0].Operations.StockCountSessionCount != 1 {
		t.Fatalf("board = %+v, want one ready HN reconciliation with stock count source", board)
	}

	createReq := smokeRequestAsRole(
		httptest.NewRequest(http.MethodPost, "/api/v1/stock-counts", bytes.NewBufferString(`{
			"id": "count-e2e-hn-260426-clean",
			"count_no": "CNT-E2E-HN-260426-CLEAN",
			"warehouse_id": "wh-hn",
			"warehouse_code": "HN",
			"scope": "shift_closing",
			"lines": [
				{
					"id": "line-hn-toner-clean",
					"item_id": "item-toner-100ml",
					"sku": "TONER-100ML",
					"batch_id": "batch-toner-2604c",
					"batch_no": "LOT-2604C",
					"location_id": "bin-hn-b04",
					"location_code": "HN-B-04",
					"expected_qty": "85.000000",
					"base_uom_code": "EA"
				}
			]
		}`)),
		authConfig,
		auth.RoleWarehouseLead,
	)
	createReq.Header.Set(response.HeaderRequestID, "req-e2e-stock-count-create")
	createRec := httptest.NewRecorder()

	stockCountsHandler(listStockCounts, createStockCount).ServeHTTP(createRec, createReq)

	if createRec.Code != http.StatusCreated {
		t.Fatalf("create count status = %d, want %d: %s", createRec.Code, http.StatusCreated, createRec.Body.String())
	}
	createdCount := decodeSmokeSuccess[stockCountResponse](t, createRec).Data
	if createdCount.Status != "open" || createdCount.AuditLogID == "" {
		t.Fatalf("created count = %+v, want open stock count with audit", createdCount)
	}

	submitReq := smokeRequestAsRole(
		httptest.NewRequest(
			http.MethodPost,
			"/api/v1/stock-counts/count-e2e-hn-260426-clean/submit",
			bytes.NewBufferString(`{
				"lines": [
					{"id": "line-hn-toner-clean", "counted_qty": "85.000000", "note": "matched shift close count"}
				]
			}`),
		),
		authConfig,
		auth.RoleWarehouseLead,
	)
	submitReq.SetPathValue("stock_count_id", "count-e2e-hn-260426-clean")
	submitReq.Header.Set(response.HeaderRequestID, "req-e2e-stock-count-submit")
	submitRec := httptest.NewRecorder()

	stockCountSubmitHandler(submitStockCount).ServeHTTP(submitRec, submitReq)

	if submitRec.Code != http.StatusOK {
		t.Fatalf("submit count status = %d, want %d: %s", submitRec.Code, http.StatusOK, submitRec.Body.String())
	}
	submittedCount := decodeSmokeSuccess[stockCountResponse](t, submitRec).Data
	if submittedCount.Status != "submitted" ||
		submittedCount.Lines[0].DeltaQty != "0.000000" ||
		submittedCount.AuditLogID == "" {
		t.Fatalf("submitted count = %+v, want clean submitted stock count", submittedCount)
	}

	reviewReq := smokeRequestAsRole(
		httptest.NewRequest(
			http.MethodGet,
			"/api/v1/warehouse/end-of-day-reconciliations?warehouse_id=wh-hn&date=2026-04-26&shift_code=day&status=open",
			nil,
		),
		authConfig,
		auth.RoleWarehouseLead,
	)
	reviewReq.Header.Set(response.HeaderRequestID, "req-e2e-shift-reconciliation")
	reviewRec := httptest.NewRecorder()

	endOfDayReconciliationsHandler(listReconciliations).ServeHTTP(reviewRec, reviewReq)

	if reviewRec.Code != http.StatusOK {
		t.Fatalf("review status = %d, want %d: %s", reviewRec.Code, http.StatusOK, reviewRec.Body.String())
	}
	review := decodeSmokeSuccess[[]endOfDayReconciliationResponse](t, reviewRec).Data
	if len(review) != 1 ||
		!review[0].Summary.ReadyToClose ||
		review[0].Summary.VarianceCount != 0 ||
		review[0].Operations.PendingIssueCount != 0 {
		t.Fatalf("review = %+v, want clean reconciliation ready to close", review)
	}

	closeReq := smokeRequestAsRole(
		httptest.NewRequest(
			http.MethodPost,
			"/api/v1/warehouse/end-of-day-reconciliations/rec-hn-260426-day/close",
			bytes.NewBufferString(`{"exception_note":""}`),
		),
		authConfig,
		auth.RoleWarehouseLead,
	)
	closeReq.SetPathValue("reconciliation_id", "rec-hn-260426-day")
	closeReq.Header.Set(response.HeaderRequestID, "req-e2e-shift-close")
	closeRec := httptest.NewRecorder()

	closeEndOfDayReconciliationHandler(closeReconciliation).ServeHTTP(closeRec, closeReq)

	if closeRec.Code != http.StatusOK {
		t.Fatalf("close status = %d, want %d: %s", closeRec.Code, http.StatusOK, closeRec.Body.String())
	}
	closed := decodeSmokeSuccess[endOfDayReconciliationResponse](t, closeRec).Data
	if closed.Status != string(inventorydomain.ReconciliationStatusClosed) ||
		closed.ClosedBy == "" ||
		closed.AuditLogID == "" {
		t.Fatalf("closed = %+v, want closed shift with audit", closed)
	}

	assertShiftClosingE2EAuditAction(t, auditStore, "inventory.stock_count.created")
	assertShiftClosingE2EAuditAction(t, auditStore, "inventory.stock_count.submitted")
	assertShiftClosingE2EAuditAction(t, auditStore, "warehouse.shift.closed")
}

func assertShiftClosingE2EAuditAction(t *testing.T, auditStore audit.LogStore, action string) {
	t.Helper()

	logs, err := auditStore.List(context.Background(), audit.Query{Action: action})
	if err != nil {
		t.Fatalf("list audit logs for %s: %v", action, err)
	}
	if len(logs) != 1 {
		t.Fatalf("audit action %s count = %d, want 1", action, len(logs))
	}
}
