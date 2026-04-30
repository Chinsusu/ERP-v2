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
)

func TestFinanceDeniedMutationsDoNotWriteAuditLogs(t *testing.T) {
	for _, tc := range []struct {
		name  string
		setup func() (http.Handler, *http.Request, *audit.InMemoryLogStore, financedomain.FinanceAuditAction)
	}{
		{
			name: "AR create requires finance manage",
			setup: func() (http.Handler, *http.Request, *audit.InMemoryLogStore, financedomain.FinanceAuditAction) {
				auditStore := audit.NewInMemoryLogStore()
				service := financeapp.NewCustomerReceivableService(financeapp.NewPrototypeCustomerReceivableStore(), auditStore)
				req := financeRegressionRequest(
					http.MethodPost,
					"/api/v1/customer-receivables",
					[]byte(`{}`),
					financeRegressionPrincipal(auth.PermissionFinanceView),
				)

				return customerReceivablesHandler(service), req, auditStore, financedomain.FinanceAuditActionReceivableCreated
			},
		},
		{
			name: "AP approve requires payment approve",
			setup: func() (http.Handler, *http.Request, *audit.InMemoryLogStore, financedomain.FinanceAuditAction) {
				auditStore := audit.NewInMemoryLogStore()
				service := financeapp.NewSupplierPayableService(financeapp.NewPrototypeSupplierPayableStore(), auditStore)
				req := financeRegressionRequest(
					http.MethodPost,
					"/api/v1/supplier-payables/ap-supplier-260430-0001/approve-payment",
					[]byte(`{}`),
					financeRegressionPrincipal(auth.PermissionFinanceView, auth.PermissionFinanceManage),
				)
				req.SetPathValue("supplier_payable_id", "ap-supplier-260430-0001")

				return supplierPayableApprovePaymentHandler(service), req, auditStore, financedomain.FinanceAuditActionPayablePaymentApproved
			},
		},
		{
			name: "COD create requires COD reconcile",
			setup: func() (http.Handler, *http.Request, *audit.InMemoryLogStore, financedomain.FinanceAuditAction) {
				auditStore := audit.NewInMemoryLogStore()
				service := financeapp.NewCODRemittanceService(financeapp.NewPrototypeCODRemittanceStore(), auditStore)
				req := financeRegressionRequest(
					http.MethodPost,
					"/api/v1/cod-remittances",
					[]byte(`{}`),
					financeRegressionPrincipal(auth.PermissionFinanceView),
				)

				return codRemittancesHandler(service), req, auditStore, financedomain.FinanceAuditActionCODRemittanceCreated
			},
		},
		{
			name: "cash create requires finance manage",
			setup: func() (http.Handler, *http.Request, *audit.InMemoryLogStore, financedomain.FinanceAuditAction) {
				auditStore := audit.NewInMemoryLogStore()
				service := financeapp.NewCashTransactionService(financeapp.NewPrototypeCashTransactionStore(), auditStore)
				req := financeRegressionRequest(
					http.MethodPost,
					"/api/v1/cash-transactions",
					[]byte(`{}`),
					financeRegressionPrincipal(auth.PermissionFinanceView),
				)

				return cashTransactionsHandler(service), req, auditStore, financedomain.FinanceAuditActionCashTransactionRecorded
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			handler, req, auditStore, action := tc.setup()
			rec := httptest.NewRecorder()

			handler.ServeHTTP(rec, req)

			if rec.Code != http.StatusForbidden {
				t.Fatalf("status = %d, want forbidden; body = %s", rec.Code, rec.Body.String())
			}
			assertFinanceRegressionAuditCount(t, auditStore, action, "", 0)
		})
	}
}

func TestFinanceMutatingHandlersWriteExpectedAuditLogs(t *testing.T) {
	for _, tc := range []struct {
		name     string
		setup    func() (http.Handler, *http.Request, *audit.InMemoryLogStore, financedomain.FinanceAuditAction, string)
		wantCode int
	}{
		{
			name: "AR receipt records audit",
			setup: func() (http.Handler, *http.Request, *audit.InMemoryLogStore, financedomain.FinanceAuditAction, string) {
				auditStore := audit.NewInMemoryLogStore()
				service := financeapp.NewCustomerReceivableService(financeapp.NewPrototypeCustomerReceivableStore(), auditStore)
				req := financeRegressionRequest(
					http.MethodPost,
					"/api/v1/customer-receivables/ar-cod-260430-0001/record-receipt",
					[]byte(`{"amount":"125000.00"}`),
					auth.MockPrincipalForRole(financeRegressionMockConfig(), auth.RoleFinanceOps),
				)
				req.SetPathValue("customer_receivable_id", "ar-cod-260430-0001")

				return customerReceivableRecordReceiptHandler(service), req, auditStore, financedomain.FinanceAuditActionReceivableReceiptRecorded, "ar-cod-260430-0001"
			},
			wantCode: http.StatusOK,
		},
		{
			name: "AP approval records audit",
			setup: func() (http.Handler, *http.Request, *audit.InMemoryLogStore, financedomain.FinanceAuditAction, string) {
				auditStore := audit.NewInMemoryLogStore()
				service := financeapp.NewSupplierPayableService(financeapp.NewPrototypeSupplierPayableStore(), auditStore)
				req := financeRegressionRequest(
					http.MethodPost,
					"/api/v1/supplier-payables/ap-supplier-260430-0001/approve-payment",
					[]byte(`{}`),
					auth.MockPrincipalForRole(financeRegressionMockConfig(), auth.RoleFinanceOps),
				)
				req.SetPathValue("supplier_payable_id", "ap-supplier-260430-0001")

				return supplierPayableApprovePaymentHandler(service), req, auditStore, financedomain.FinanceAuditActionPayablePaymentApproved, "ap-supplier-260430-0001"
			},
			wantCode: http.StatusOK,
		},
		{
			name: "COD discrepancy records audit",
			setup: func() (http.Handler, *http.Request, *audit.InMemoryLogStore, financedomain.FinanceAuditAction, string) {
				auditStore := audit.NewInMemoryLogStore()
				service := financeapp.NewCODRemittanceService(financeapp.NewPrototypeCODRemittanceStore(), auditStore)
				req := financeRegressionRequest(
					http.MethodPost,
					"/api/v1/cod-remittances/cod-remit-260430-0001/record-discrepancy",
					[]byte(`{"id":"disc-regression-1","line_id":"cod-remit-260430-0001-line-1","reason":"carrier short","owner_id":"finance-user"}`),
					auth.MockPrincipalForRole(financeRegressionMockConfig(), auth.RoleFinanceOps),
				)
				req.SetPathValue("cod_remittance_id", "cod-remit-260430-0001")

				return codRemittanceDiscrepancyHandler(service), req, auditStore, financedomain.FinanceAuditActionCODRemittanceDiscrepancyRecorded, "cod-remit-260430-0001"
			},
			wantCode: http.StatusOK,
		},
		{
			name: "cash receipt records audit",
			setup: func() (http.Handler, *http.Request, *audit.InMemoryLogStore, financedomain.FinanceAuditAction, string) {
				auditStore := audit.NewInMemoryLogStore()
				service := financeapp.NewCashTransactionService(financeapp.NewPrototypeCashTransactionStore(), auditStore)
				req := financeRegressionRequest(
					http.MethodPost,
					"/api/v1/cash-transactions",
					[]byte(`{
						"id":"cash-regression-0001",
						"transaction_no":"CASH-IN-REGRESSION-0001",
						"direction":"cash_in",
						"business_date":"2026-04-30",
						"counterparty_id":"carrier-ghn",
						"counterparty_name":"GHN COD",
						"payment_method":"bank_transfer",
						"reference_no":"BANK-REGRESSION-0001",
						"total_amount":"125000.00",
						"currency_code":"VND",
						"allocations":[
							{"id":"cash-regression-line-1","target_type":"customer_receivable","target_id":"ar-cod-260430-0001","target_no":"AR-COD-260430-0001","amount":"125000.00"}
						]
					}`),
					auth.MockPrincipalForRole(financeRegressionMockConfig(), auth.RoleFinanceOps),
				)

				return cashTransactionsHandler(service), req, auditStore, financedomain.FinanceAuditActionCashTransactionRecorded, "cash-regression-0001"
			},
			wantCode: http.StatusCreated,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			handler, req, auditStore, action, entityID := tc.setup()
			rec := httptest.NewRecorder()

			handler.ServeHTTP(rec, req)

			if rec.Code != tc.wantCode {
				t.Fatalf("status = %d, want %d; body = %s", rec.Code, tc.wantCode, rec.Body.String())
			}
			assertFinanceRegressionAuditCount(t, auditStore, action, entityID, 1)
		})
	}
}

func financeRegressionRequest(
	method string,
	target string,
	body []byte,
	principal auth.Principal,
) *http.Request {
	req := httptest.NewRequest(method, target, bytes.NewReader(body))
	return req.WithContext(auth.WithPrincipal(req.Context(), principal))
}

func financeRegressionPrincipal(permissions ...auth.PermissionKey) auth.Principal {
	return auth.Principal{
		UserID:      "finance-regression-user",
		Email:       "finance-regression@example.local",
		Name:        "Finance Regression",
		Role:        auth.RoleFinanceOps,
		Permissions: permissions,
	}
}

func financeRegressionMockConfig() auth.MockConfig {
	return auth.MockConfig{
		Email:       "admin@example.local",
		Password:    "local-only-mock-password",
		AccessToken: "local-dev-access-token",
	}
}

func assertFinanceRegressionAuditCount(
	t *testing.T,
	auditStore *audit.InMemoryLogStore,
	action financedomain.FinanceAuditAction,
	entityID string,
	want int,
) {
	t.Helper()

	logs, err := auditStore.List(context.Background(), audit.Query{
		Action:   string(action),
		EntityID: entityID,
	})
	if err != nil {
		t.Fatalf("list audit logs for %s: %v", action, err)
	}
	if len(logs) != want {
		t.Fatalf("audit logs for %s/%s = %d, want %d", action, entityID, len(logs), want)
	}
}
