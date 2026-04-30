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

func TestSubcontractPayablePaymentFinanceE2ESmoke(t *testing.T) {
	authConfig := smokeAuthConfig()
	services := newSupplierPayablePaymentFinanceE2EServices()

	createPayableReq := smokeRequestAsRole(
		httptest.NewRequest(http.MethodPost, "/api/v1/supplier-payables", bytes.NewBufferString(`{
			"id": "ap-e2e-subcontract-payment",
			"payable_no": "AP-E2E-SUBCONTRACT-PAYMENT",
			"supplier_id": "factory-hcm-001",
			"supplier_code": "FACT-HCM-001",
			"supplier_name": "Nha May Gia Cong HCM",
			"source_document": {"type": "subcontract_payment_milestone", "id": "spm-e2e-subcontract-final", "no": "SPM-E2E-SUBCONTRACT-FINAL"},
			"total_amount": "6800000.00",
			"currency_code": "VND",
			"due_date": "2026-05-10",
			"lines": [
				{
					"id": "ap-e2e-subcontract-payment-line-1",
					"description": "Accepted finished goods final payment milestone",
					"source_document": {"type": "subcontract_order", "id": "sco-e2e-subcontract-accepted", "no": "SCO-E2E-SUBCONTRACT-ACCEPTED"},
					"amount": "6800000.00"
				}
			]
		}`)),
		authConfig,
		auth.RoleFinanceOps,
	)
	createPayableReq.Header.Set(response.HeaderRequestID, "req-e2e-subcontract-ap-create")
	createPayableRec := httptest.NewRecorder()

	supplierPayablesHandler(services.payables).ServeHTTP(createPayableRec, createPayableReq)

	if createPayableRec.Code != http.StatusCreated {
		t.Fatalf("create subcontract payable status = %d, want %d: %s", createPayableRec.Code, http.StatusCreated, createPayableRec.Body.String())
	}
	createdPayable := decodeSmokeSuccess[supplierPayableResponse](t, createPayableRec).Data
	if createdPayable.Status != string(financedomain.PayableStatusOpen) ||
		createdPayable.SourceDocument.Type != string(financedomain.SourceDocumentTypeSubcontractPaymentMilestone) ||
		createdPayable.OutstandingAmount != "6800000.00" ||
		createdPayable.AuditLogID == "" {
		t.Fatalf("created subcontract payable = %+v, want open factory AP with milestone source and audit", createdPayable)
	}

	for _, step := range []struct {
		name       string
		handler    http.HandlerFunc
		wantStatus financedomain.PayableStatus
	}{
		{name: "request-payment", handler: supplierPayableRequestPaymentHandler(services.payables), wantStatus: financedomain.PayableStatusPaymentRequested},
		{name: "approve-payment", handler: supplierPayableApprovePaymentHandler(services.payables), wantStatus: financedomain.PayableStatusPaymentApproved},
	} {
		req := smokeRequestAsRole(
			httptest.NewRequest(http.MethodPost, "/api/v1/supplier-payables/ap-e2e-subcontract-payment/"+step.name, bytes.NewBufferString(`{}`)),
			authConfig,
			auth.RoleFinanceOps,
		)
		req.SetPathValue("supplier_payable_id", "ap-e2e-subcontract-payment")
		req.Header.Set(response.HeaderRequestID, "req-e2e-subcontract-ap-"+step.name)
		rec := httptest.NewRecorder()

		step.handler.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("%s subcontract payable status = %d, want %d: %s", step.name, rec.Code, http.StatusOK, rec.Body.String())
		}
		result := decodeSmokeSuccess[supplierPayableActionResultResponse](t, rec).Data
		if result.CurrentStatus != string(step.wantStatus) || result.AuditLogID == "" {
			t.Fatalf("%s subcontract payable result = %+v, want %s with audit", step.name, result, step.wantStatus)
		}
	}

	createCashReq := smokeRequestAsRole(
		httptest.NewRequest(http.MethodPost, "/api/v1/cash-transactions", bytes.NewBufferString(`{
			"id": "cash-e2e-subcontract-payment",
			"transaction_no": "CASH-OUT-E2E-SUBCONTRACT-PAYMENT",
			"direction": "cash_out",
			"business_date": "2026-04-30",
			"counterparty_id": "factory-hcm-001",
			"counterparty_name": "Nha May Gia Cong HCM",
			"payment_method": "bank_transfer",
			"reference_no": "BANK-E2E-SUBCONTRACT-PAYMENT",
			"total_amount": "6800000.00",
			"currency_code": "VND",
			"allocations": [
				{
					"id": "cash-e2e-subcontract-payment-line-1",
					"target_type": "supplier_payable",
					"target_id": "ap-e2e-subcontract-payment",
					"target_no": "AP-E2E-SUBCONTRACT-PAYMENT",
					"amount": "6800000.00"
				}
			]
		}`)),
		authConfig,
		auth.RoleFinanceOps,
	)
	createCashReq.Header.Set(response.HeaderRequestID, "req-e2e-subcontract-cash-out")
	createCashRec := httptest.NewRecorder()

	cashTransactionsHandler(services.cashTransactions).ServeHTTP(createCashRec, createCashReq)

	if createCashRec.Code != http.StatusCreated {
		t.Fatalf("create subcontract cash-out status = %d, want %d: %s", createCashRec.Code, http.StatusCreated, createCashRec.Body.String())
	}
	cashPayment := decodeSmokeSuccess[cashTransactionResponse](t, createCashRec).Data
	if cashPayment.Status != string(financedomain.CashTransactionStatusPosted) ||
		cashPayment.Direction != string(financedomain.CashTransactionDirectionOut) ||
		cashPayment.TotalAmount != "6800000.00" ||
		cashPayment.AuditLogID == "" ||
		len(cashPayment.Allocations) != 1 ||
		cashPayment.Allocations[0].TargetType != string(financedomain.CashAllocationTargetSupplierPayable) {
		t.Fatalf("subcontract cash payment = %+v, want posted factory payable allocation with audit", cashPayment)
	}

	recordPaymentReq := smokeRequestAsRole(
		httptest.NewRequest(http.MethodPost, "/api/v1/supplier-payables/ap-e2e-subcontract-payment/record-payment", bytes.NewBufferString(`{"amount":"6800000.00"}`)),
		authConfig,
		auth.RoleFinanceOps,
	)
	recordPaymentReq.SetPathValue("supplier_payable_id", "ap-e2e-subcontract-payment")
	recordPaymentReq.Header.Set(response.HeaderRequestID, "req-e2e-subcontract-ap-record-payment")
	recordPaymentRec := httptest.NewRecorder()

	supplierPayableRecordPaymentHandler(services.payables).ServeHTTP(recordPaymentRec, recordPaymentReq)

	if recordPaymentRec.Code != http.StatusOK {
		t.Fatalf("record subcontract payable payment status = %d, want %d: %s", recordPaymentRec.Code, http.StatusOK, recordPaymentRec.Body.String())
	}
	paidPayable := decodeSmokeSuccess[supplierPayableActionResultResponse](t, recordPaymentRec).Data
	if paidPayable.CurrentStatus != string(financedomain.PayableStatusPaid) ||
		paidPayable.SupplierPayable.PaidAmount != "6800000.00" ||
		paidPayable.SupplierPayable.OutstandingAmount != "0.00" ||
		paidPayable.AuditLogID == "" {
		t.Fatalf("paid subcontract payable = %+v, want factory AP paid and closed out with audit", paidPayable)
	}

	for _, action := range []financedomain.FinanceAuditAction{
		financedomain.FinanceAuditActionPayableCreated,
		financedomain.FinanceAuditActionPayablePaymentRequested,
		financedomain.FinanceAuditActionPayablePaymentApproved,
		financedomain.FinanceAuditActionCashTransactionRecorded,
		financedomain.FinanceAuditActionPayablePaymentRecorded,
	} {
		assertSupplierPayablePaymentE2EAuditAction(t, services.auditStore, action)
	}
}
