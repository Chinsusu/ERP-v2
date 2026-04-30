package main

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	financeapp "github.com/Chinsusu/ERP-v2/apps/api/internal/modules/finance/application"
	financedomain "github.com/Chinsusu/ERP-v2/apps/api/internal/modules/finance/domain"
	"github.com/Chinsusu/ERP-v2/apps/api/internal/shared/audit"
	"github.com/Chinsusu/ERP-v2/apps/api/internal/shared/auth"
	"github.com/Chinsusu/ERP-v2/apps/api/internal/shared/response"
)

func TestCODHappyPathFinanceE2ESmoke(t *testing.T) {
	authConfig := smokeAuthConfig()
	services := newCODHappyPathFinanceE2EServices()

	createReceivableReq := smokeRequestAsRole(
		httptest.NewRequest(http.MethodPost, "/api/v1/customer-receivables", bytes.NewBufferString(`{
			"id": "ar-e2e-cod-happy",
			"receivable_no": "AR-E2E-COD-HAPPY",
			"customer_id": "customer-hcm-001",
			"customer_code": "KH-HCM-001",
			"customer_name": "My Pham HCM Retail",
			"source_document": {"type": "shipment", "id": "shipment-e2e-cod-happy", "no": "SHP-E2E-COD-HAPPY"},
			"total_amount": "1250000.00",
			"currency_code": "VND",
			"due_date": "2026-05-03",
			"lines": [
				{
					"id": "ar-e2e-cod-happy-line-1",
					"description": "COD delivered shipment",
					"source_document": {"type": "shipment", "id": "shipment-e2e-cod-happy", "no": "SHP-E2E-COD-HAPPY"},
					"amount": "1250000.00"
				}
			]
		}`)),
		authConfig,
		auth.RoleFinanceOps,
	)
	createReceivableReq.Header.Set(response.HeaderRequestID, "req-e2e-cod-happy-ar-create")
	createReceivableRec := httptest.NewRecorder()

	customerReceivablesHandler(services.receivables).ServeHTTP(createReceivableRec, createReceivableReq)

	if createReceivableRec.Code != http.StatusCreated {
		t.Fatalf("create receivable status = %d, want %d: %s", createReceivableRec.Code, http.StatusCreated, createReceivableRec.Body.String())
	}
	createdReceivable := decodeSmokeSuccess[customerReceivableResponse](t, createReceivableRec).Data
	if createdReceivable.Status != string(financedomain.ReceivableStatusOpen) ||
		createdReceivable.OutstandingAmount != "1250000.00" ||
		createdReceivable.AuditLogID == "" {
		t.Fatalf("created receivable = %+v, want open COD receivable with audit", createdReceivable)
	}

	createRemittanceReq := smokeRequestAsRole(
		httptest.NewRequest(http.MethodPost, "/api/v1/cod-remittances", bytes.NewBufferString(`{
			"id": "cod-e2e-happy",
			"remittance_no": "COD-E2E-HAPPY",
			"carrier_id": "carrier-ghn",
			"carrier_code": "GHN",
			"carrier_name": "GHN Express",
			"business_date": "2026-04-30",
			"expected_amount": "1250000.00",
			"remitted_amount": "1250000.00",
			"currency_code": "VND",
			"lines": [
				{
					"id": "cod-e2e-happy-line-1",
					"receivable_id": "ar-e2e-cod-happy",
					"receivable_no": "AR-E2E-COD-HAPPY",
					"shipment_id": "shipment-e2e-cod-happy",
					"tracking_no": "GHN-E2E-COD-HAPPY",
					"customer_name": "My Pham HCM Retail",
					"expected_amount": "1250000.00",
					"remitted_amount": "1250000.00"
				}
			]
		}`)),
		authConfig,
		auth.RoleFinanceOps,
	)
	createRemittanceReq.Header.Set(response.HeaderRequestID, "req-e2e-cod-happy-remit-create")
	createRemittanceRec := httptest.NewRecorder()

	codRemittancesHandler(services.codRemittances).ServeHTTP(createRemittanceRec, createRemittanceReq)

	if createRemittanceRec.Code != http.StatusCreated {
		t.Fatalf("create COD remittance status = %d, want %d: %s", createRemittanceRec.Code, http.StatusCreated, createRemittanceRec.Body.String())
	}
	createdRemittance := decodeSmokeSuccess[codRemittanceResponse](t, createRemittanceRec).Data
	if createdRemittance.Status != string(financedomain.CODRemittanceStatusDraft) ||
		createdRemittance.DiscrepancyAmount != "0.00" ||
		createdRemittance.AuditLogID == "" {
		t.Fatalf("created COD remittance = %+v, want draft matched amount with audit", createdRemittance)
	}

	for _, step := range []struct {
		name       string
		handler    http.HandlerFunc
		wantStatus financedomain.CODRemittanceStatus
	}{
		{name: "match", handler: codRemittanceMatchHandler(services.codRemittances), wantStatus: financedomain.CODRemittanceStatusMatching},
		{name: "submit", handler: codRemittanceSubmitHandler(services.codRemittances), wantStatus: financedomain.CODRemittanceStatusSubmitted},
		{name: "approve", handler: codRemittanceApproveHandler(services.codRemittances), wantStatus: financedomain.CODRemittanceStatusApproved},
		{name: "close", handler: codRemittanceCloseHandler(services.codRemittances), wantStatus: financedomain.CODRemittanceStatusClosed},
	} {
		req := smokeRequestAsRole(
			httptest.NewRequest(http.MethodPost, "/api/v1/cod-remittances/cod-e2e-happy/"+step.name, nil),
			authConfig,
			auth.RoleFinanceOps,
		)
		req.SetPathValue("cod_remittance_id", "cod-e2e-happy")
		req.Header.Set(response.HeaderRequestID, "req-e2e-cod-happy-"+step.name)
		rec := httptest.NewRecorder()

		step.handler.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("%s COD remittance status = %d, want %d: %s", step.name, rec.Code, http.StatusOK, rec.Body.String())
		}
		result := decodeSmokeSuccess[codRemittanceActionResultResponse](t, rec).Data
		if result.CurrentStatus != string(step.wantStatus) || result.AuditLogID == "" {
			t.Fatalf("%s COD result = %+v, want %s with audit", step.name, result, step.wantStatus)
		}
	}

	createCashReq := smokeRequestAsRole(
		httptest.NewRequest(http.MethodPost, "/api/v1/cash-transactions", bytes.NewBufferString(`{
			"id": "cash-e2e-cod-happy",
			"transaction_no": "CASH-IN-E2E-COD-HAPPY",
			"direction": "cash_in",
			"business_date": "2026-04-30",
			"counterparty_id": "carrier-ghn",
			"counterparty_name": "GHN Express",
			"payment_method": "bank_transfer",
			"reference_no": "BANK-E2E-COD-HAPPY",
			"total_amount": "1250000.00",
			"currency_code": "VND",
			"allocations": [
				{
					"id": "cash-e2e-cod-happy-line-1",
					"target_type": "cod_remittance",
					"target_id": "cod-e2e-happy",
					"target_no": "COD-E2E-HAPPY",
					"amount": "1250000.00"
				}
			]
		}`)),
		authConfig,
		auth.RoleFinanceOps,
	)
	createCashReq.Header.Set(response.HeaderRequestID, "req-e2e-cod-happy-cash")
	createCashRec := httptest.NewRecorder()

	cashTransactionsHandler(services.cashTransactions).ServeHTTP(createCashRec, createCashReq)

	if createCashRec.Code != http.StatusCreated {
		t.Fatalf("create cash status = %d, want %d: %s", createCashRec.Code, http.StatusCreated, createCashRec.Body.String())
	}
	cashReceipt := decodeSmokeSuccess[cashTransactionResponse](t, createCashRec).Data
	if cashReceipt.Status != string(financedomain.CashTransactionStatusPosted) ||
		cashReceipt.Direction != string(financedomain.CashTransactionDirectionIn) ||
		cashReceipt.TotalAmount != "1250000.00" ||
		cashReceipt.AuditLogID == "" ||
		len(cashReceipt.Allocations) != 1 ||
		cashReceipt.Allocations[0].TargetType != string(financedomain.CashAllocationTargetCODRemittance) {
		t.Fatalf("cash receipt = %+v, want posted COD remittance allocation with audit", cashReceipt)
	}

	recordReceiptReq := smokeRequestAsRole(
		httptest.NewRequest(http.MethodPost, "/api/v1/customer-receivables/ar-e2e-cod-happy/record-receipt", bytes.NewBufferString(`{"amount":"1250000.00"}`)),
		authConfig,
		auth.RoleFinanceOps,
	)
	recordReceiptReq.SetPathValue("customer_receivable_id", "ar-e2e-cod-happy")
	recordReceiptReq.Header.Set(response.HeaderRequestID, "req-e2e-cod-happy-ar-receipt")
	recordReceiptRec := httptest.NewRecorder()

	customerReceivableRecordReceiptHandler(services.receivables).ServeHTTP(recordReceiptRec, recordReceiptReq)

	if recordReceiptRec.Code != http.StatusOK {
		t.Fatalf("record receivable receipt status = %d, want %d: %s", recordReceiptRec.Code, http.StatusOK, recordReceiptRec.Body.String())
	}
	paidReceivable := decodeSmokeSuccess[customerReceivableActionResultResponse](t, recordReceiptRec).Data
	if paidReceivable.CurrentStatus != string(financedomain.ReceivableStatusPaid) ||
		paidReceivable.CustomerReceivable.PaidAmount != "1250000.00" ||
		paidReceivable.CustomerReceivable.OutstandingAmount != "0.00" ||
		paidReceivable.AuditLogID == "" {
		t.Fatalf("paid receivable = %+v, want paid and closed out with audit", paidReceivable)
	}

	for _, action := range []financedomain.FinanceAuditAction{
		financedomain.FinanceAuditActionReceivableCreated,
		financedomain.FinanceAuditActionCODRemittanceCreated,
		financedomain.FinanceAuditActionCODRemittanceMatched,
		financedomain.FinanceAuditActionCODRemittanceSubmitted,
		financedomain.FinanceAuditActionCODRemittanceApproved,
		financedomain.FinanceAuditActionCODRemittanceClosed,
		financedomain.FinanceAuditActionCashTransactionRecorded,
		financedomain.FinanceAuditActionReceivableReceiptRecorded,
	} {
		assertCODHappyPathE2EAuditAction(t, services.auditStore, action)
	}
}

type codHappyPathFinanceE2EServices struct {
	receivables      financeapp.CustomerReceivableService
	codRemittances   financeapp.CODRemittanceService
	cashTransactions financeapp.CashTransactionService
	auditStore       *audit.InMemoryLogStore
}

func newCODHappyPathFinanceE2EServices() codHappyPathFinanceE2EServices {
	auditStore := audit.NewInMemoryLogStore()

	return codHappyPathFinanceE2EServices{
		receivables:      financeapp.NewCustomerReceivableService(financeapp.NewPrototypeCustomerReceivableStore(), auditStore),
		codRemittances:   financeapp.NewCODRemittanceService(financeapp.NewPrototypeCODRemittanceStore(), auditStore),
		cashTransactions: financeapp.NewCashTransactionService(financeapp.NewPrototypeCashTransactionStore(), auditStore),
		auditStore:       auditStore,
	}
}

func assertCODHappyPathE2EAuditAction(t *testing.T, auditStore audit.LogStore, action financedomain.FinanceAuditAction) {
	t.Helper()

	logs, err := auditStore.List(context.Background(), audit.Query{Action: string(action)})
	if err != nil {
		t.Fatalf("list audit logs for %s: %v", action, err)
	}
	if len(logs) != 1 {
		t.Fatalf("audit action %s count = %d, want 1", action, len(logs))
	}
}
