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

func TestSupplierPayablePaymentFinanceE2ESmoke(t *testing.T) {
	authConfig := smokeAuthConfig()
	services := newSupplierPayablePaymentFinanceE2EServices()

	createPayableReq := smokeRequestAsRole(
		httptest.NewRequest(http.MethodPost, "/api/v1/supplier-payables", bytes.NewBufferString(`{
			"id": "ap-e2e-supplier-payment",
			"payable_no": "AP-E2E-SUPPLIER-PAYMENT",
			"supplier_id": "supplier-hcm-001",
			"supplier_code": "SUP-HCM-001",
			"supplier_name": "Nguyen Lieu HCM",
			"source_document": {"type": "qc_inspection", "id": "iqc-e2e-supplier-pass", "no": "IQC-E2E-SUPPLIER-PASS"},
			"total_amount": "4250000.00",
			"currency_code": "VND",
			"due_date": "2026-05-07",
			"lines": [
				{
					"id": "ap-e2e-supplier-payment-line-1",
					"description": "Accepted raw material after inbound QC",
					"source_document": {"type": "warehouse_receipt", "id": "grn-e2e-supplier-pass", "no": "GRN-E2E-SUPPLIER-PASS"},
					"amount": "3000000.00"
				},
				{
					"id": "ap-e2e-supplier-payment-line-2",
					"description": "Accepted packaging after inbound QC",
					"source_document": {"type": "purchase_order", "id": "po-e2e-supplier-pass", "no": "PO-E2E-SUPPLIER-PASS"},
					"amount": "1250000.00"
				}
			]
		}`)),
		authConfig,
		auth.RoleFinanceOps,
	)
	createPayableReq.Header.Set(response.HeaderRequestID, "req-e2e-supplier-ap-create")
	createPayableRec := httptest.NewRecorder()

	supplierPayablesHandler(services.payables).ServeHTTP(createPayableRec, createPayableReq)

	if createPayableRec.Code != http.StatusCreated {
		t.Fatalf("create payable status = %d, want %d: %s", createPayableRec.Code, http.StatusCreated, createPayableRec.Body.String())
	}
	createdPayable := decodeSmokeSuccess[supplierPayableResponse](t, createPayableRec).Data
	if createdPayable.Status != string(financedomain.PayableStatusOpen) ||
		createdPayable.OutstandingAmount != "4250000.00" ||
		createdPayable.AuditLogID == "" {
		t.Fatalf("created payable = %+v, want open AP with audit", createdPayable)
	}

	for _, step := range []struct {
		name       string
		handler    http.HandlerFunc
		body       string
		wantStatus financedomain.PayableStatus
	}{
		{name: "request-payment", handler: supplierPayableRequestPaymentHandler(services.payables), body: `{}`, wantStatus: financedomain.PayableStatusPaymentRequested},
		{name: "approve-payment", handler: supplierPayableApprovePaymentHandler(services.payables), body: `{}`, wantStatus: financedomain.PayableStatusPaymentApproved},
	} {
		req := smokeRequestAsRole(
			httptest.NewRequest(http.MethodPost, "/api/v1/supplier-payables/ap-e2e-supplier-payment/"+step.name, bytes.NewBufferString(step.body)),
			authConfig,
			auth.RoleFinanceOps,
		)
		req.SetPathValue("supplier_payable_id", "ap-e2e-supplier-payment")
		req.Header.Set(response.HeaderRequestID, "req-e2e-supplier-ap-"+step.name)
		rec := httptest.NewRecorder()

		step.handler.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("%s payable status = %d, want %d: %s", step.name, rec.Code, http.StatusOK, rec.Body.String())
		}
		result := decodeSmokeSuccess[supplierPayableActionResultResponse](t, rec).Data
		if result.CurrentStatus != string(step.wantStatus) || result.AuditLogID == "" {
			t.Fatalf("%s payable result = %+v, want %s with audit", step.name, result, step.wantStatus)
		}
	}

	createCashReq := smokeRequestAsRole(
		httptest.NewRequest(http.MethodPost, "/api/v1/cash-transactions", bytes.NewBufferString(`{
			"id": "cash-e2e-supplier-payment",
			"transaction_no": "CASH-OUT-E2E-SUPPLIER-PAYMENT",
			"direction": "cash_out",
			"business_date": "2026-04-30",
			"counterparty_id": "supplier-hcm-001",
			"counterparty_name": "Nguyen Lieu HCM",
			"payment_method": "bank_transfer",
			"reference_no": "BANK-E2E-SUPPLIER-PAYMENT",
			"total_amount": "4250000.00",
			"currency_code": "VND",
			"allocations": [
				{
					"id": "cash-e2e-supplier-payment-line-1",
					"target_type": "supplier_payable",
					"target_id": "ap-e2e-supplier-payment",
					"target_no": "AP-E2E-SUPPLIER-PAYMENT",
					"amount": "4250000.00"
				}
			]
		}`)),
		authConfig,
		auth.RoleFinanceOps,
	)
	createCashReq.Header.Set(response.HeaderRequestID, "req-e2e-supplier-cash-out")
	createCashRec := httptest.NewRecorder()

	cashTransactionsHandler(services.cashTransactions).ServeHTTP(createCashRec, createCashReq)

	if createCashRec.Code != http.StatusCreated {
		t.Fatalf("create cash-out status = %d, want %d: %s", createCashRec.Code, http.StatusCreated, createCashRec.Body.String())
	}
	cashPayment := decodeSmokeSuccess[cashTransactionResponse](t, createCashRec).Data
	if cashPayment.Status != string(financedomain.CashTransactionStatusPosted) ||
		cashPayment.Direction != string(financedomain.CashTransactionDirectionOut) ||
		cashPayment.TotalAmount != "4250000.00" ||
		cashPayment.AuditLogID == "" ||
		len(cashPayment.Allocations) != 1 ||
		cashPayment.Allocations[0].TargetType != string(financedomain.CashAllocationTargetSupplierPayable) {
		t.Fatalf("cash payment = %+v, want posted supplier payable allocation with audit", cashPayment)
	}

	recordPaymentReq := smokeRequestAsRole(
		httptest.NewRequest(http.MethodPost, "/api/v1/supplier-payables/ap-e2e-supplier-payment/record-payment", bytes.NewBufferString(`{"amount":"4250000.00"}`)),
		authConfig,
		auth.RoleFinanceOps,
	)
	recordPaymentReq.SetPathValue("supplier_payable_id", "ap-e2e-supplier-payment")
	recordPaymentReq.Header.Set(response.HeaderRequestID, "req-e2e-supplier-ap-record-payment")
	recordPaymentRec := httptest.NewRecorder()

	supplierPayableRecordPaymentHandler(services.payables).ServeHTTP(recordPaymentRec, recordPaymentReq)

	if recordPaymentRec.Code != http.StatusOK {
		t.Fatalf("record payable payment status = %d, want %d: %s", recordPaymentRec.Code, http.StatusOK, recordPaymentRec.Body.String())
	}
	paidPayable := decodeSmokeSuccess[supplierPayableActionResultResponse](t, recordPaymentRec).Data
	if paidPayable.CurrentStatus != string(financedomain.PayableStatusPaid) ||
		paidPayable.SupplierPayable.PaidAmount != "4250000.00" ||
		paidPayable.SupplierPayable.OutstandingAmount != "0.00" ||
		paidPayable.AuditLogID == "" {
		t.Fatalf("paid payable = %+v, want paid and closed out with audit", paidPayable)
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

type supplierPayablePaymentFinanceE2EServices struct {
	payables         financeapp.SupplierPayableService
	cashTransactions financeapp.CashTransactionService
	auditStore       *audit.InMemoryLogStore
}

func newSupplierPayablePaymentFinanceE2EServices() supplierPayablePaymentFinanceE2EServices {
	auditStore := audit.NewInMemoryLogStore()

	return supplierPayablePaymentFinanceE2EServices{
		payables:         financeapp.NewSupplierPayableService(financeapp.NewPrototypeSupplierPayableStore(), auditStore),
		cashTransactions: financeapp.NewCashTransactionService(financeapp.NewPrototypeCashTransactionStore(), auditStore),
		auditStore:       auditStore,
	}
}

func assertSupplierPayablePaymentE2EAuditAction(t *testing.T, auditStore audit.LogStore, action financedomain.FinanceAuditAction) {
	t.Helper()

	logs, err := auditStore.List(context.Background(), audit.Query{Action: string(action)})
	if err != nil {
		t.Fatalf("list audit logs for %s: %v", action, err)
	}
	if len(logs) != 1 {
		t.Fatalf("audit action %s count = %d, want 1", action, len(logs))
	}
}
