package application

import (
	"context"
	"errors"
	"testing"
	"time"

	apperrors "github.com/Chinsusu/ERP-v2/apps/api/internal/shared/errors"
)

func TestFinanceDashboardServiceSummarizesPrototypeStores(t *testing.T) {
	service := newTestFinanceDashboardService().WithClock(func() time.Time {
		return time.Date(2026, 5, 8, 10, 0, 0, 0, time.UTC)
	})

	metrics, err := service.GetFinanceDashboardMetrics(context.Background(), FinanceDashboardFilter{
		BusinessDate: "2026-05-08",
	})
	if err != nil {
		t.Fatalf("metrics: %v", err)
	}
	if metrics.BusinessDate != "2026-05-08" || metrics.CurrencyCode != "VND" {
		t.Fatalf("metadata = %+v", metrics)
	}
	if metrics.AR.OpenCount != 1 ||
		metrics.AR.OverdueCount != 1 ||
		metrics.AR.OpenAmount != "1250000.00" ||
		metrics.AR.OverdueAmount != "1250000.00" {
		t.Fatalf("ar metrics = %+v", metrics.AR)
	}
	if metrics.AP.OpenCount != 1 ||
		metrics.AP.DueCount != 1 ||
		metrics.AP.OpenAmount != "4250000.00" ||
		metrics.AP.DueAmount != "4250000.00" {
		t.Fatalf("ap metrics = %+v", metrics.AP)
	}
	if metrics.COD.PendingCount != 1 ||
		metrics.COD.DiscrepancyCount != 1 ||
		metrics.COD.PendingAmount != "2000000.00" ||
		metrics.COD.DiscrepancyAmount != "-50000.00" {
		t.Fatalf("cod metrics = %+v", metrics.COD)
	}
}

func TestFinanceDashboardServiceSummarizesCashToday(t *testing.T) {
	service := newTestFinanceDashboardService()

	metrics, err := service.GetFinanceDashboardMetrics(context.Background(), FinanceDashboardFilter{
		BusinessDate: "2026-04-30",
	})
	if err != nil {
		t.Fatalf("metrics: %v", err)
	}
	if metrics.Cash.TransactionCount != 2 ||
		metrics.Cash.CashInToday != "1250000.00" ||
		metrics.Cash.CashOutToday != "4250000.00" ||
		metrics.Cash.NetCashToday != "-3000000.00" {
		t.Fatalf("cash metrics = %+v", metrics.Cash)
	}
}

func TestFinanceDashboardServiceDefaultsBusinessDateToHoChiMinh(t *testing.T) {
	service := newTestFinanceDashboardService().WithClock(func() time.Time {
		return time.Date(2026, 4, 30, 17, 30, 0, 0, time.UTC)
	})

	metrics, err := service.GetFinanceDashboardMetrics(context.Background(), FinanceDashboardFilter{})
	if err != nil {
		t.Fatalf("metrics: %v", err)
	}
	if metrics.BusinessDate != "2026-05-01" {
		t.Fatalf("business date = %q", metrics.BusinessDate)
	}
}

func TestFinanceDashboardServiceRejectsInvalidBusinessDate(t *testing.T) {
	service := newTestFinanceDashboardService()

	_, err := service.GetFinanceDashboardMetrics(context.Background(), FinanceDashboardFilter{
		BusinessDate: "30/04/2026",
	})
	var appErr apperrors.AppError
	if !errors.As(err, &appErr) {
		t.Fatalf("error = %T, want app error", err)
	}
	if appErr.Code != ErrorCodeFinanceDashboardValidation {
		t.Fatalf("code = %q, want dashboard validation", appErr.Code)
	}
}

func newTestFinanceDashboardService() FinanceDashboardService {
	return NewFinanceDashboardService(
		NewPrototypeCustomerReceivableStore(),
		NewPrototypeSupplierPayableStore(),
		NewPrototypeCODRemittanceStore(),
		NewPrototypeCashTransactionStore(),
	).WithClock(func() time.Time {
		return time.Date(2026, 4, 30, 10, 0, 0, 0, time.UTC)
	})
}
