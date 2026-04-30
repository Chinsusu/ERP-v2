package main

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	financedomain "github.com/Chinsusu/ERP-v2/apps/api/internal/modules/finance/domain"
	"github.com/Chinsusu/ERP-v2/apps/api/internal/shared/auth"
	"github.com/Chinsusu/ERP-v2/apps/api/internal/shared/response"
)

func TestCODDiscrepancyFinanceE2ESmoke(t *testing.T) {
	authConfig := smokeAuthConfig()
	services := newCODHappyPathFinanceE2EServices()

	createReceivableReq := smokeRequestAsRole(
		httptest.NewRequest(http.MethodPost, "/api/v1/customer-receivables", bytes.NewBufferString(`{
			"id": "ar-e2e-cod-discrepancy",
			"receivable_no": "AR-E2E-COD-DISCREPANCY",
			"customer_id": "customer-hcm-001",
			"customer_code": "KH-HCM-001",
			"customer_name": "My Pham HCM Retail",
			"source_document": {"type": "shipment", "id": "shipment-e2e-cod-discrepancy", "no": "SHP-E2E-COD-DISCREPANCY"},
			"total_amount": "1250000.00",
			"currency_code": "VND",
			"due_date": "2026-05-03",
			"lines": [
				{
					"id": "ar-e2e-cod-discrepancy-line-1",
					"description": "COD delivered shipment with carrier short remittance",
					"source_document": {"type": "shipment", "id": "shipment-e2e-cod-discrepancy", "no": "SHP-E2E-COD-DISCREPANCY"},
					"amount": "1250000.00"
				}
			]
		}`)),
		authConfig,
		auth.RoleFinanceOps,
	)
	createReceivableReq.Header.Set(response.HeaderRequestID, "req-e2e-cod-discrepancy-ar-create")
	createReceivableRec := httptest.NewRecorder()

	customerReceivablesHandler(services.receivables).ServeHTTP(createReceivableRec, createReceivableReq)

	if createReceivableRec.Code != http.StatusCreated {
		t.Fatalf("create receivable status = %d, want %d: %s", createReceivableRec.Code, http.StatusCreated, createReceivableRec.Body.String())
	}
	createdReceivable := decodeSmokeSuccess[customerReceivableResponse](t, createReceivableRec).Data
	if createdReceivable.Status != string(financedomain.ReceivableStatusOpen) ||
		createdReceivable.OutstandingAmount != "1250000.00" ||
		createdReceivable.AuditLogID == "" {
		t.Fatalf("created receivable = %+v, want open receivable with audit", createdReceivable)
	}

	createRemittanceReq := smokeRequestAsRole(
		httptest.NewRequest(http.MethodPost, "/api/v1/cod-remittances", bytes.NewBufferString(`{
			"id": "cod-e2e-discrepancy",
			"remittance_no": "COD-E2E-DISCREPANCY",
			"carrier_id": "carrier-ghn",
			"carrier_code": "GHN",
			"carrier_name": "GHN Express",
			"business_date": "2026-04-30",
			"expected_amount": "1250000.00",
			"remitted_amount": "1200000.00",
			"currency_code": "VND",
			"lines": [
				{
					"id": "cod-e2e-discrepancy-line-1",
					"receivable_id": "ar-e2e-cod-discrepancy",
					"receivable_no": "AR-E2E-COD-DISCREPANCY",
					"shipment_id": "shipment-e2e-cod-discrepancy",
					"tracking_no": "GHN-E2E-COD-DISCREPANCY",
					"customer_name": "My Pham HCM Retail",
					"expected_amount": "1250000.00",
					"remitted_amount": "1200000.00"
				}
			]
		}`)),
		authConfig,
		auth.RoleFinanceOps,
	)
	createRemittanceReq.Header.Set(response.HeaderRequestID, "req-e2e-cod-discrepancy-remit-create")
	createRemittanceRec := httptest.NewRecorder()

	codRemittancesHandler(services.codRemittances).ServeHTTP(createRemittanceRec, createRemittanceReq)

	if createRemittanceRec.Code != http.StatusCreated {
		t.Fatalf("create COD remittance status = %d, want %d: %s", createRemittanceRec.Code, http.StatusCreated, createRemittanceRec.Body.String())
	}
	createdRemittance := decodeSmokeSuccess[codRemittanceResponse](t, createRemittanceRec).Data
	if createdRemittance.Status != string(financedomain.CODRemittanceStatusDraft) ||
		createdRemittance.DiscrepancyAmount != "-50000.00" ||
		createdRemittance.AuditLogID == "" {
		t.Fatalf("created COD remittance = %+v, want draft short remittance with audit", createdRemittance)
	}

	recordDiscrepancyReq := smokeRequestAsRole(
		httptest.NewRequest(http.MethodPost, "/api/v1/cod-remittances/cod-e2e-discrepancy/record-discrepancy", bytes.NewBufferString(`{
			"id": "disc-e2e-cod-short",
			"line_id": "cod-e2e-discrepancy-line-1",
			"reason": "carrier remitted short by 50.000 VND",
			"owner_id": "finance-ops"
		}`)),
		authConfig,
		auth.RoleFinanceOps,
	)
	recordDiscrepancyReq.SetPathValue("cod_remittance_id", "cod-e2e-discrepancy")
	recordDiscrepancyReq.Header.Set(response.HeaderRequestID, "req-e2e-cod-discrepancy-record")
	recordDiscrepancyRec := httptest.NewRecorder()

	codRemittanceDiscrepancyHandler(services.codRemittances).ServeHTTP(recordDiscrepancyRec, recordDiscrepancyReq)

	if recordDiscrepancyRec.Code != http.StatusOK {
		t.Fatalf("record COD discrepancy status = %d, want %d: %s", recordDiscrepancyRec.Code, http.StatusOK, recordDiscrepancyRec.Body.String())
	}
	recordedDiscrepancy := decodeSmokeSuccess[codRemittanceActionResultResponse](t, recordDiscrepancyRec).Data
	if recordedDiscrepancy.CurrentStatus != string(financedomain.CODRemittanceStatusDiscrepancy) ||
		recordedDiscrepancy.CODRemittance.DiscrepancyAmount != "-50000.00" ||
		len(recordedDiscrepancy.CODRemittance.Discrepancies) != 1 ||
		recordedDiscrepancy.AuditLogID == "" {
		t.Fatalf("recorded COD discrepancy = %+v, want traced short remittance", recordedDiscrepancy)
	}

	for _, step := range []struct {
		name       string
		handler    http.HandlerFunc
		wantStatus financedomain.CODRemittanceStatus
	}{
		{name: "submit", handler: codRemittanceSubmitHandler(services.codRemittances), wantStatus: financedomain.CODRemittanceStatusSubmitted},
		{name: "approve", handler: codRemittanceApproveHandler(services.codRemittances), wantStatus: financedomain.CODRemittanceStatusApproved},
	} {
		req := smokeRequestAsRole(
			httptest.NewRequest(http.MethodPost, "/api/v1/cod-remittances/cod-e2e-discrepancy/"+step.name, nil),
			authConfig,
			auth.RoleFinanceOps,
		)
		req.SetPathValue("cod_remittance_id", "cod-e2e-discrepancy")
		req.Header.Set(response.HeaderRequestID, "req-e2e-cod-discrepancy-"+step.name)
		rec := httptest.NewRecorder()

		step.handler.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("%s COD remittance status = %d, want %d: %s", step.name, rec.Code, http.StatusOK, rec.Body.String())
		}
		result := decodeSmokeSuccess[codRemittanceActionResultResponse](t, rec).Data
		if result.CurrentStatus != string(step.wantStatus) ||
			result.CODRemittance.DiscrepancyAmount != "-50000.00" ||
			len(result.CODRemittance.Discrepancies) != 1 ||
			result.AuditLogID == "" {
			t.Fatalf("%s COD result = %+v, want %s with discrepancy audit", step.name, result, step.wantStatus)
		}
	}

	createCashReq := smokeRequestAsRole(
		httptest.NewRequest(http.MethodPost, "/api/v1/cash-transactions", bytes.NewBufferString(`{
			"id": "cash-e2e-cod-discrepancy",
			"transaction_no": "CASH-IN-E2E-COD-DISCREPANCY",
			"direction": "cash_in",
			"business_date": "2026-04-30",
			"counterparty_id": "carrier-ghn",
			"counterparty_name": "GHN Express",
			"payment_method": "bank_transfer",
			"reference_no": "BANK-E2E-COD-DISCREPANCY",
			"total_amount": "1200000.00",
			"currency_code": "VND",
			"allocations": [
				{
					"id": "cash-e2e-cod-discrepancy-line-1",
					"target_type": "cod_remittance",
					"target_id": "cod-e2e-discrepancy",
					"target_no": "COD-E2E-DISCREPANCY",
					"amount": "1200000.00"
				}
			]
		}`)),
		authConfig,
		auth.RoleFinanceOps,
	)
	createCashReq.Header.Set(response.HeaderRequestID, "req-e2e-cod-discrepancy-cash")
	createCashRec := httptest.NewRecorder()

	cashTransactionsHandler(services.cashTransactions).ServeHTTP(createCashRec, createCashReq)

	if createCashRec.Code != http.StatusCreated {
		t.Fatalf("create cash status = %d, want %d: %s", createCashRec.Code, http.StatusCreated, createCashRec.Body.String())
	}
	cashReceipt := decodeSmokeSuccess[cashTransactionResponse](t, createCashRec).Data
	if cashReceipt.Status != string(financedomain.CashTransactionStatusPosted) ||
		cashReceipt.TotalAmount != "1200000.00" ||
		cashReceipt.AuditLogID == "" {
		t.Fatalf("cash receipt = %+v, want posted partial carrier remittance", cashReceipt)
	}

	recordReceiptReq := smokeRequestAsRole(
		httptest.NewRequest(http.MethodPost, "/api/v1/customer-receivables/ar-e2e-cod-discrepancy/record-receipt", bytes.NewBufferString(`{"amount":"1200000.00"}`)),
		authConfig,
		auth.RoleFinanceOps,
	)
	recordReceiptReq.SetPathValue("customer_receivable_id", "ar-e2e-cod-discrepancy")
	recordReceiptReq.Header.Set(response.HeaderRequestID, "req-e2e-cod-discrepancy-ar-receipt")
	recordReceiptRec := httptest.NewRecorder()

	customerReceivableRecordReceiptHandler(services.receivables).ServeHTTP(recordReceiptRec, recordReceiptReq)

	if recordReceiptRec.Code != http.StatusOK {
		t.Fatalf("record receivable receipt status = %d, want %d: %s", recordReceiptRec.Code, http.StatusOK, recordReceiptRec.Body.String())
	}
	partiallyPaidReceivable := decodeSmokeSuccess[customerReceivableActionResultResponse](t, recordReceiptRec).Data
	if partiallyPaidReceivable.CurrentStatus != string(financedomain.ReceivableStatusPartiallyPaid) ||
		partiallyPaidReceivable.CustomerReceivable.PaidAmount != "1200000.00" ||
		partiallyPaidReceivable.CustomerReceivable.OutstandingAmount != "50000.00" ||
		partiallyPaidReceivable.AuditLogID == "" {
		t.Fatalf("partially paid receivable = %+v, want 50.000 VND outstanding", partiallyPaidReceivable)
	}

	disputeReceivableReq := smokeRequestAsRole(
		httptest.NewRequest(http.MethodPost, "/api/v1/customer-receivables/ar-e2e-cod-discrepancy/mark-disputed", bytes.NewBufferString(`{"reason":"carrier COD short remittance under investigation"}`)),
		authConfig,
		auth.RoleFinanceOps,
	)
	disputeReceivableReq.SetPathValue("customer_receivable_id", "ar-e2e-cod-discrepancy")
	disputeReceivableReq.Header.Set(response.HeaderRequestID, "req-e2e-cod-discrepancy-ar-dispute")
	disputeReceivableRec := httptest.NewRecorder()

	customerReceivableMarkDisputedHandler(services.receivables).ServeHTTP(disputeReceivableRec, disputeReceivableReq)

	if disputeReceivableRec.Code != http.StatusOK {
		t.Fatalf("mark receivable disputed status = %d, want %d: %s", disputeReceivableRec.Code, http.StatusOK, disputeReceivableRec.Body.String())
	}
	disputedReceivable := decodeSmokeSuccess[customerReceivableActionResultResponse](t, disputeReceivableRec).Data
	if disputedReceivable.CurrentStatus != string(financedomain.ReceivableStatusDisputed) ||
		disputedReceivable.CustomerReceivable.OutstandingAmount != "50000.00" ||
		disputedReceivable.CustomerReceivable.DisputeReason == "" ||
		disputedReceivable.AuditLogID == "" {
		t.Fatalf("disputed receivable = %+v, want outstanding balance held in dispute", disputedReceivable)
	}

	for _, action := range []financedomain.FinanceAuditAction{
		financedomain.FinanceAuditActionReceivableCreated,
		financedomain.FinanceAuditActionCODRemittanceCreated,
		financedomain.FinanceAuditActionCODRemittanceDiscrepancyRecorded,
		financedomain.FinanceAuditActionCODRemittanceSubmitted,
		financedomain.FinanceAuditActionCODRemittanceApproved,
		financedomain.FinanceAuditActionCashTransactionRecorded,
		financedomain.FinanceAuditActionReceivableReceiptRecorded,
		financedomain.FinanceAuditActionReceivableDisputed,
	} {
		assertCODHappyPathE2EAuditAction(t, services.auditStore, action)
	}
}
