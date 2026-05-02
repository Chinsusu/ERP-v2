package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"math/big"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	financeapp "github.com/Chinsusu/ERP-v2/apps/api/internal/modules/finance/application"
	financedomain "github.com/Chinsusu/ERP-v2/apps/api/internal/modules/finance/domain"
	"github.com/Chinsusu/ERP-v2/apps/api/internal/shared/audit"
	"github.com/Chinsusu/ERP-v2/apps/api/internal/shared/auth"
	"github.com/Chinsusu/ERP-v2/apps/api/internal/shared/config"
	"github.com/Chinsusu/ERP-v2/apps/api/internal/shared/response"
)

func TestFinanceDashboardAndSummaryReportUseRuntimeFinanceStores(t *testing.T) {
	databaseURL := os.Getenv("ERP_TEST_DATABASE_URL")
	if databaseURL == "" {
		t.Skip("ERP_TEST_DATABASE_URL is not set")
	}

	ctx := context.Background()
	db, err := sql.Open("pgx", databaseURL)
	if err != nil {
		t.Fatalf("open database: %v", err)
	}
	defer db.Close()
	seedFinanceRuntimeIntegrationOrg(t, ctx, db)

	stores, closeStores, err := newRuntimeFinanceStores(config.Config{
		AppEnv:      "dev",
		DatabaseURL: databaseURL,
	})
	if err != nil {
		t.Fatalf("new runtime finance stores: %v", err)
	}
	defer func() {
		if closeStores != nil {
			if err := closeStores(); err != nil {
				t.Fatalf("close finance stores: %v", err)
			}
		}
	}()

	auditStore := audit.NewPostgresLogStore(db, audit.PostgresLogStoreConfig{DefaultOrgID: localAuditOrgID})
	dashboardService := financeapp.NewFinanceDashboardService(
		stores.customerReceivables,
		stores.supplierPayables,
		stores.codRemittances,
		stores.cashTransactions,
	).WithClock(func() time.Time {
		return time.Date(2026, 5, 2, 8, 0, 0, 0, time.UTC)
	})
	reportHandler := financeSummaryReportHandler(
		stores.customerReceivables,
		stores.supplierPayables,
		stores.codRemittances,
		stores.cashTransactions,
	)

	dashboardBefore := financeRuntimeDashboardPayload(t, dashboardService)
	reportBefore := financeRuntimeReportPayload(t, reportHandler)

	suffix := fmt.Sprintf("%d", time.Now().UTC().UnixNano())
	createFinanceRuntimeIntegrationRows(t, ctx, stores, auditStore, suffix)

	dashboardAfter := financeRuntimeDashboardPayload(t, dashboardService)
	reportAfter := financeRuntimeReportPayload(t, reportHandler)

	assertIntDelta(t, "dashboard AR open count", dashboardBefore.AR.OpenCount, dashboardAfter.AR.OpenCount, 1)
	assertMoneyDelta(t, "dashboard AR open amount", dashboardBefore.AR.OpenAmount, dashboardAfter.AR.OpenAmount, "1250000.00")
	assertIntDelta(t, "dashboard AP open count", dashboardBefore.AP.OpenCount, dashboardAfter.AP.OpenCount, 1)
	assertMoneyDelta(t, "dashboard AP open amount", dashboardBefore.AP.OpenAmount, dashboardAfter.AP.OpenAmount, "4250000.00")
	assertIntDelta(t, "dashboard COD pending count", dashboardBefore.COD.PendingCount, dashboardAfter.COD.PendingCount, 1)
	assertMoneyDelta(t, "dashboard COD pending amount", dashboardBefore.COD.PendingAmount, dashboardAfter.COD.PendingAmount, "1250000.00")
	assertIntDelta(t, "dashboard cash transaction count", dashboardBefore.Cash.TransactionCount, dashboardAfter.Cash.TransactionCount, 1)
	assertMoneyDelta(t, "dashboard cash in", dashboardBefore.Cash.CashInToday, dashboardAfter.Cash.CashInToday, "1250000.00")

	assertIntDelta(t, "report AR open count", reportBefore.AR.OpenCount, reportAfter.AR.OpenCount, 1)
	assertMoneyDelta(t, "report AR outstanding amount", reportBefore.AR.OutstandingAmount, reportAfter.AR.OutstandingAmount, "1250000.00")
	assertIntDelta(t, "report AP open count", reportBefore.AP.OpenCount, reportAfter.AP.OpenCount, 1)
	assertMoneyDelta(t, "report AP outstanding amount", reportBefore.AP.OutstandingAmount, reportAfter.AP.OutstandingAmount, "4250000.00")
	assertIntDelta(t, "report COD pending count", reportBefore.COD.PendingCount, reportAfter.COD.PendingCount, 1)
	assertMoneyDelta(t, "report COD pending amount", reportBefore.COD.PendingAmount, reportAfter.COD.PendingAmount, "1250000.00")
	assertIntDelta(t, "report cash transaction count", reportBefore.Cash.TransactionCount, reportAfter.Cash.TransactionCount, 1)
	assertMoneyDelta(t, "report cash net amount", reportBefore.Cash.NetCashAmount, reportAfter.Cash.NetCashAmount, "1250000.00")
	if reportAfter.CurrencyCode != "VND" || dashboardAfter.CurrencyCode != "VND" {
		t.Fatalf("currency codes = dashboard %q report %q, want VND", dashboardAfter.CurrencyCode, reportAfter.CurrencyCode)
	}
}

func createFinanceRuntimeIntegrationRows(
	t *testing.T,
	ctx context.Context,
	stores financeRuntimeStores,
	auditStore audit.LogStore,
	suffix string,
) {
	t.Helper()

	receivableService := financeapp.NewCustomerReceivableService(stores.customerReceivables, auditStore).
		WithClock(func() time.Time { return time.Date(2026, 5, 2, 9, 0, 0, 0, time.UTC) })
	receivableSource := financeapp.SourceDocumentInput{
		Type: string(financedomain.SourceDocumentTypeShipment),
		ID:   "shipment-s15-06-02-" + suffix,
		No:   "SHP-S15-06-02-" + suffix,
	}
	if _, err := receivableService.CreateCustomerReceivable(ctx, financeapp.CreateCustomerReceivableInput{
		ID:             "ar-s15-06-02-" + suffix,
		ReceivableNo:   "AR-S15-06-02-" + suffix,
		CustomerID:     "customer-s15-06-02",
		CustomerCode:   "CUS-S15-06-02",
		CustomerName:   "Sprint 15 Finance Runtime Customer",
		SourceDocument: receivableSource,
		TotalAmount:    "1250000.00",
		CurrencyCode:   "VND",
		DueDate:        "2026-05-01",
		ActorID:        "finance-user",
		RequestID:      "req-ar-s15-06-02-" + suffix,
		Lines: []financeapp.CustomerReceivableLineInput{{
			ID:             "ar-line-s15-06-02-" + suffix,
			Description:    "Runtime finance dashboard/report AR",
			SourceDocument: receivableSource,
			Amount:         "1250000.00",
		}},
	}); err != nil {
		t.Fatalf("create receivable: %v", err)
	}

	payableService := financeapp.NewSupplierPayableService(stores.supplierPayables, auditStore).
		WithClock(func() time.Time { return time.Date(2026, 5, 2, 9, 5, 0, 0, time.UTC) })
	payableSource := financeapp.SourceDocumentInput{
		Type: string(financedomain.SourceDocumentTypeQCInspection),
		ID:   "qc-s15-06-02-" + suffix,
		No:   "QC-S15-06-02-" + suffix,
	}
	if _, err := payableService.CreateSupplierPayable(ctx, financeapp.CreateSupplierPayableInput{
		ID:             "ap-s15-06-02-" + suffix,
		PayableNo:      "AP-S15-06-02-" + suffix,
		SupplierID:     "supplier-s15-06-02",
		SupplierCode:   "SUP-S15-06-02",
		SupplierName:   "Sprint 15 Finance Runtime Supplier",
		SourceDocument: payableSource,
		TotalAmount:    "4250000.00",
		CurrencyCode:   "VND",
		DueDate:        "2026-05-02",
		ActorID:        "finance-user",
		RequestID:      "req-ap-s15-06-02-" + suffix,
		Lines: []financeapp.SupplierPayableLineInput{{
			ID:          "ap-line-s15-06-02-" + suffix,
			Description: "Runtime finance dashboard/report AP",
			SourceDocument: financeapp.SourceDocumentInput{
				Type: string(financedomain.SourceDocumentTypeWarehouseReceipt),
				ID:   "gr-s15-06-02-" + suffix,
				No:   "GR-S15-06-02-" + suffix,
			},
			Amount: "4250000.00",
		}},
	}); err != nil {
		t.Fatalf("create payable: %v", err)
	}

	codService := financeapp.NewCODRemittanceService(stores.codRemittances, auditStore).
		WithClock(func() time.Time { return time.Date(2026, 5, 2, 9, 10, 0, 0, time.UTC) })
	if _, err := codService.CreateCODRemittance(ctx, financeapp.CreateCODRemittanceInput{
		ID:             "cod-s15-06-02-" + suffix,
		RemittanceNo:   "COD-S15-06-02-" + suffix,
		CarrierID:      "carrier-s15-06-02",
		CarrierCode:    "GHN",
		CarrierName:    "GHN Express",
		BusinessDate:   "2026-05-02",
		ExpectedAmount: "1250000.00",
		RemittedAmount: "1200000.00",
		CurrencyCode:   "VND",
		ActorID:        "finance-user",
		RequestID:      "req-cod-s15-06-02-" + suffix,
		Lines: []financeapp.CODRemittanceLineInput{{
			ID:             "cod-line-s15-06-02-" + suffix,
			ReceivableID:   "ar-s15-06-02-" + suffix,
			ReceivableNo:   "AR-S15-06-02-" + suffix,
			ShipmentID:     "shipment-s15-06-02-" + suffix,
			TrackingNo:     "GHN-S15-06-02-" + suffix,
			CustomerName:   "Sprint 15 Finance Runtime Customer",
			ExpectedAmount: "1250000.00",
			RemittedAmount: "1200000.00",
		}},
	}); err != nil {
		t.Fatalf("create cod remittance: %v", err)
	}

	cashService := financeapp.NewCashTransactionService(stores.cashTransactions, auditStore).
		WithClock(func() time.Time { return time.Date(2026, 5, 2, 9, 15, 0, 0, time.UTC) })
	if _, err := cashService.CreateCashTransaction(ctx, financeapp.CreateCashTransactionInput{
		ID:               "cash-s15-06-02-" + suffix,
		TransactionNo:    "CASH-IN-S15-06-02-" + suffix,
		Direction:        string(financedomain.CashTransactionDirectionIn),
		BusinessDate:     "2026-05-02",
		CounterpartyID:   "carrier-s15-06-02",
		CounterpartyName: "GHN Express",
		PaymentMethod:    "bank_transfer",
		ReferenceNo:      "BANK-S15-06-02-" + suffix,
		TotalAmount:      "1250000.00",
		CurrencyCode:     "VND",
		ActorID:          "finance-user",
		RequestID:        "req-cash-s15-06-02-" + suffix,
		Allocations: []financeapp.CashTransactionAllocationInput{{
			ID:         "cash-alloc-s15-06-02-" + suffix,
			TargetType: string(financedomain.CashAllocationTargetCustomerReceivable),
			TargetID:   "ar-s15-06-02-" + suffix,
			TargetNo:   "AR-S15-06-02-" + suffix,
			Amount:     "1250000.00",
		}},
	}); err != nil {
		t.Fatalf("create cash transaction: %v", err)
	}
}

func financeRuntimeDashboardPayload(
	t *testing.T,
	service financeapp.FinanceDashboardService,
) financeDashboardResponse {
	t.Helper()

	req := cashTransactionRequest(http.MethodGet, "/api/v1/finance/dashboard?business_date=2026-05-02", nil, auth.RoleFinanceOps)
	rec := httptest.NewRecorder()
	financeDashboardHandler(service).ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("dashboard status = %d, body = %s", rec.Code, rec.Body.String())
	}

	var payload response.SuccessEnvelope[financeDashboardResponse]
	if err := json.NewDecoder(rec.Body).Decode(&payload); err != nil {
		t.Fatalf("decode dashboard: %v", err)
	}

	return payload.Data
}

func financeRuntimeReportPayload(t *testing.T, handler http.HandlerFunc) financeSummaryReportResponse {
	t.Helper()

	req := cashTransactionRequest(
		http.MethodGet,
		"/api/v1/reports/finance-summary?from_date=2026-05-02&to_date=2026-05-02&business_date=2026-05-02",
		nil,
		auth.RoleFinanceOps,
	)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("report status = %d, body = %s", rec.Code, rec.Body.String())
	}

	var payload response.SuccessEnvelope[financeSummaryReportResponse]
	if err := json.NewDecoder(rec.Body).Decode(&payload); err != nil {
		t.Fatalf("decode report: %v", err)
	}

	return payload.Data
}

func seedFinanceRuntimeIntegrationOrg(t *testing.T, ctx context.Context, db *sql.DB) {
	t.Helper()

	_, err := db.ExecContext(
		ctx,
		`INSERT INTO core.organizations (id, code, name, status)
VALUES ($1, 'ERP_LOCAL', 'ERP Local Demo', 'active')
ON CONFLICT (code) DO UPDATE
SET name = EXCLUDED.name,
    status = EXCLUDED.status,
    updated_at = now()`,
		localAuditOrgID,
	)
	if err != nil {
		t.Fatalf("seed org: %v", err)
	}
}

func assertIntDelta(t *testing.T, label string, before int, after int, want int) {
	t.Helper()

	if got := after - before; got != want {
		t.Fatalf("%s delta = %d, want %d (before=%d after=%d)", label, got, want, before, after)
	}
}

func assertMoneyDelta(t *testing.T, label string, before string, after string, want string) {
	t.Helper()

	gotCents := new(big.Int).Sub(moneyCents(t, after), moneyCents(t, before))
	wantCents := moneyCents(t, want)
	if gotCents.Cmp(wantCents) != 0 {
		t.Fatalf("%s delta = %s, want %s (before=%s after=%s)", label, formatMoneyCents(gotCents), want, before, after)
	}
}

func moneyCents(t *testing.T, value string) *big.Int {
	t.Helper()

	value = strings.TrimSpace(value)
	negative := strings.HasPrefix(value, "-")
	value = strings.TrimPrefix(value, "-")
	parts := strings.Split(value, ".")
	if len(parts) == 1 {
		parts = append(parts, "00")
	}
	if len(parts) != 2 || len(parts[1]) > 2 {
		t.Fatalf("invalid money value %q", value)
	}
	fraction := parts[1] + strings.Repeat("0", 2-len(parts[1]))
	cents, ok := new(big.Int).SetString(parts[0]+fraction, 10)
	if !ok {
		t.Fatalf("invalid money value %q", value)
	}
	if negative {
		cents.Neg(cents)
	}

	return cents
}

func formatMoneyCents(cents *big.Int) string {
	value := new(big.Int).Set(cents)
	negative := value.Sign() < 0
	if negative {
		value.Abs(value)
	}
	digits := value.String()
	if len(digits) <= 2 {
		digits = strings.Repeat("0", 3-len(digits)) + digits
	}
	formatted := digits[:len(digits)-2] + "." + digits[len(digits)-2:]
	if negative && value.Sign() != 0 {
		return "-" + formatted
	}

	return formatted
}
