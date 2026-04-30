package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	inventoryapp "github.com/Chinsusu/ERP-v2/apps/api/internal/modules/inventory/application"
	"github.com/Chinsusu/ERP-v2/apps/api/internal/shared/auth"
	"github.com/Chinsusu/ERP-v2/apps/api/internal/shared/response"
)

func TestReportingPermissionRegression(t *testing.T) {
	inventoryService := inventoryapp.NewListAvailableStock(inventoryapp.NewPrototypeStockAvailabilityStore())

	tests := []struct {
		name           string
		handler        http.HandlerFunc
		target         string
		principal      auth.Principal
		wantPermission auth.PermissionKey
	}{
		{
			name:           "inventory report requires reports view",
			handler:        inventorySnapshotReportHandler(inventoryService),
			target:         "/api/v1/reports/inventory-snapshot?business_date=2026-04-30",
			principal:      reportingPermissionPrincipal(),
			wantPermission: auth.PermissionReportsView,
		},
		{
			name:           "inventory CSV requires reports view before export",
			handler:        inventorySnapshotCSVExportHandler(inventoryService),
			target:         "/api/v1/reports/inventory-snapshot/export.csv?business_date=2026-04-30",
			principal:      reportingPermissionPrincipal(auth.PermissionReportsExport),
			wantPermission: auth.PermissionReportsView,
		},
		{
			name:           "inventory CSV requires export after reports view",
			handler:        inventorySnapshotCSVExportHandler(inventoryService),
			target:         "/api/v1/reports/inventory-snapshot/export.csv?business_date=2026-04-30",
			principal:      reportingPermissionPrincipal(auth.PermissionReportsView),
			wantPermission: auth.PermissionReportsExport,
		},
		{
			name:           "operations report requires reports view",
			handler:        operationsDailyReportHandler(prototypeOperationsDailySignalSource{}),
			target:         "/api/v1/reports/operations-daily?business_date=2026-04-30",
			principal:      reportingPermissionPrincipal(),
			wantPermission: auth.PermissionReportsView,
		},
		{
			name:           "operations CSV requires reports view before export",
			handler:        operationsDailyCSVExportHandler(prototypeOperationsDailySignalSource{}),
			target:         "/api/v1/reports/operations-daily/export.csv?business_date=2026-04-30",
			principal:      reportingPermissionPrincipal(auth.PermissionReportsExport),
			wantPermission: auth.PermissionReportsView,
		},
		{
			name:           "operations CSV requires export after reports view",
			handler:        operationsDailyCSVExportHandler(prototypeOperationsDailySignalSource{}),
			target:         "/api/v1/reports/operations-daily/export.csv?business_date=2026-04-30",
			principal:      reportingPermissionPrincipal(auth.PermissionReportsView),
			wantPermission: auth.PermissionReportsExport,
		},
		{
			name:           "finance report requires reports view before finance view",
			handler:        newTestFinanceSummaryReportHandler(),
			target:         "/api/v1/reports/finance-summary?business_date=2026-04-30",
			principal:      reportingPermissionPrincipal(auth.PermissionFinanceReportsView),
			wantPermission: auth.PermissionReportsView,
		},
		{
			name:           "finance report requires finance view after reports view",
			handler:        newTestFinanceSummaryReportHandler(),
			target:         "/api/v1/reports/finance-summary?business_date=2026-04-30",
			principal:      reportingPermissionPrincipal(auth.PermissionReportsView),
			wantPermission: auth.PermissionFinanceReportsView,
		},
		{
			name:           "finance CSV requires reports view before finance and export",
			handler:        newTestFinanceSummaryCSVExportHandler(),
			target:         "/api/v1/reports/finance-summary/export.csv?business_date=2026-04-30",
			principal:      reportingPermissionPrincipal(auth.PermissionFinanceReportsView, auth.PermissionReportsExport),
			wantPermission: auth.PermissionReportsView,
		},
		{
			name:           "finance CSV requires finance view before export",
			handler:        newTestFinanceSummaryCSVExportHandler(),
			target:         "/api/v1/reports/finance-summary/export.csv?business_date=2026-04-30",
			principal:      reportingPermissionPrincipal(auth.PermissionReportsView, auth.PermissionReportsExport),
			wantPermission: auth.PermissionFinanceReportsView,
		},
		{
			name:           "finance CSV requires export after reports and finance view",
			handler:        newTestFinanceSummaryCSVExportHandler(),
			target:         "/api/v1/reports/finance-summary/export.csv?business_date=2026-04-30",
			principal:      reportingPermissionPrincipal(auth.PermissionReportsView, auth.PermissionFinanceReportsView),
			wantPermission: auth.PermissionReportsExport,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			req := reportingPermissionRequest(tc.target, tc.principal)
			rec := httptest.NewRecorder()

			tc.handler.ServeHTTP(rec, req)

			assertReportingPermissionDenied(t, rec, tc.wantPermission)
		})
	}
}

func reportingPermissionRequest(target string, principal auth.Principal) *http.Request {
	req := httptest.NewRequest(http.MethodGet, target, nil)
	req.Header.Set(response.HeaderRequestID, "req-reporting-permission-regression")
	return req.WithContext(auth.WithPrincipal(req.Context(), principal))
}

func reportingPermissionPrincipal(permissions ...auth.PermissionKey) auth.Principal {
	return auth.Principal{
		UserID:      "reporting-permission-regression",
		Email:       "reporting-permission-regression@example.local",
		Name:        "Reporting Permission Regression",
		Role:        auth.RoleFinanceOps,
		Permissions: permissions,
	}
}

func assertReportingPermissionDenied(t *testing.T, rec *httptest.ResponseRecorder, permission auth.PermissionKey) {
	t.Helper()

	if rec.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want %d; body = %s", rec.Code, http.StatusForbidden, rec.Body.String())
	}

	var payload response.ErrorEnvelope
	if err := json.NewDecoder(rec.Body).Decode(&payload); err != nil {
		t.Fatalf("decode error envelope: %v", err)
	}
	if payload.Error.Code != response.ErrorCodeForbidden {
		t.Fatalf("error code = %q, want %q", payload.Error.Code, response.ErrorCodeForbidden)
	}
	if got := payload.Error.Details["permission"]; got != string(permission) {
		t.Fatalf("permission detail = %v, want %s", got, permission)
	}
}
