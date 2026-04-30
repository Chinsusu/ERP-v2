package main

import (
	"net/http"

	financeapp "github.com/Chinsusu/ERP-v2/apps/api/internal/modules/finance/application"
	"github.com/Chinsusu/ERP-v2/apps/api/internal/shared/auth"
	apperrors "github.com/Chinsusu/ERP-v2/apps/api/internal/shared/errors"
	"github.com/Chinsusu/ERP-v2/apps/api/internal/shared/response"
)

type financeDashboardResponse struct {
	BusinessDate string                                    `json:"business_date"`
	GeneratedAt  string                                    `json:"generated_at"`
	CurrencyCode string                                    `json:"currency_code"`
	AR           financeDashboardReceivableMetricsResponse `json:"ar"`
	AP           financeDashboardPayableMetricsResponse    `json:"ap"`
	COD          financeDashboardCODMetricsResponse        `json:"cod"`
	Cash         financeDashboardCashMetricsResponse       `json:"cash"`
}

type financeDashboardReceivableMetricsResponse struct {
	OpenCount         int    `json:"open_count"`
	OverdueCount      int    `json:"overdue_count"`
	DisputedCount     int    `json:"disputed_count"`
	OpenAmount        string `json:"open_amount"`
	OverdueAmount     string `json:"overdue_amount"`
	OutstandingAmount string `json:"outstanding_amount"`
}

type financeDashboardPayableMetricsResponse struct {
	OpenCount             int    `json:"open_count"`
	DueCount              int    `json:"due_count"`
	PaymentRequestedCount int    `json:"payment_requested_count"`
	PaymentApprovedCount  int    `json:"payment_approved_count"`
	OpenAmount            string `json:"open_amount"`
	DueAmount             string `json:"due_amount"`
	OutstandingAmount     string `json:"outstanding_amount"`
}

type financeDashboardCODMetricsResponse struct {
	PendingCount      int    `json:"pending_count"`
	DiscrepancyCount  int    `json:"discrepancy_count"`
	PendingAmount     string `json:"pending_amount"`
	DiscrepancyAmount string `json:"discrepancy_amount"`
}

type financeDashboardCashMetricsResponse struct {
	TransactionCount int    `json:"transaction_count"`
	CashInToday      string `json:"cash_in_today"`
	CashOutToday     string `json:"cash_out_today"`
	NetCashToday     string `json:"net_cash_today"`
}

func financeDashboardHandler(service financeapp.FinanceDashboardService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			response.WriteError(w, r, http.StatusMethodNotAllowed, response.ErrorCodeNotFound, "Route not found", nil)
			return
		}
		principal, ok := auth.PrincipalFromContext(r.Context())
		if !ok {
			response.WriteError(w, r, http.StatusUnauthorized, response.ErrorCodeUnauthorized, "Authentication required", nil)
			return
		}
		if !auth.HasPermission(principal, auth.PermissionFinanceView) {
			writePermissionDenied(w, r, auth.PermissionFinanceView)
			return
		}

		metrics, err := service.GetFinanceDashboardMetrics(r.Context(), financeapp.FinanceDashboardFilter{
			BusinessDate: r.URL.Query().Get("business_date"),
		})
		if err != nil {
			writeFinanceDashboardError(w, r, err)
			return
		}

		response.WriteSuccess(w, r, http.StatusOK, newFinanceDashboardResponse(metrics))
	}
}

func newFinanceDashboardResponse(metrics financeapp.FinanceDashboardMetrics) financeDashboardResponse {
	return financeDashboardResponse{
		BusinessDate: metrics.BusinessDate,
		GeneratedAt:  timeString(metrics.GeneratedAt),
		CurrencyCode: metrics.CurrencyCode,
		AR: financeDashboardReceivableMetricsResponse{
			OpenCount:         metrics.AR.OpenCount,
			OverdueCount:      metrics.AR.OverdueCount,
			DisputedCount:     metrics.AR.DisputedCount,
			OpenAmount:        metrics.AR.OpenAmount,
			OverdueAmount:     metrics.AR.OverdueAmount,
			OutstandingAmount: metrics.AR.OutstandingAmount,
		},
		AP: financeDashboardPayableMetricsResponse{
			OpenCount:             metrics.AP.OpenCount,
			DueCount:              metrics.AP.DueCount,
			PaymentRequestedCount: metrics.AP.PaymentRequestedCount,
			PaymentApprovedCount:  metrics.AP.PaymentApprovedCount,
			OpenAmount:            metrics.AP.OpenAmount,
			DueAmount:             metrics.AP.DueAmount,
			OutstandingAmount:     metrics.AP.OutstandingAmount,
		},
		COD: financeDashboardCODMetricsResponse{
			PendingCount:      metrics.COD.PendingCount,
			DiscrepancyCount:  metrics.COD.DiscrepancyCount,
			PendingAmount:     metrics.COD.PendingAmount,
			DiscrepancyAmount: metrics.COD.DiscrepancyAmount,
		},
		Cash: financeDashboardCashMetricsResponse{
			TransactionCount: metrics.Cash.TransactionCount,
			CashInToday:      metrics.Cash.CashInToday,
			CashOutToday:     metrics.Cash.CashOutToday,
			NetCashToday:     metrics.Cash.NetCashToday,
		},
	}
}

func writeFinanceDashboardError(w http.ResponseWriter, r *http.Request, err error) {
	if appErr, ok := apperrors.As(err); ok {
		response.WriteError(w, r, appErr.HTTPStatus, appErr.Code, appErr.Message, appErr.Details)
		return
	}

	response.WriteError(w, r, http.StatusConflict, response.ErrorCodeConflict, "Finance dashboard request could not be processed", nil)
}
