package main

import (
	"errors"
	"net/http"

	financeapp "github.com/Chinsusu/ERP-v2/apps/api/internal/modules/finance/application"
	reportingdomain "github.com/Chinsusu/ERP-v2/apps/api/internal/modules/reporting/domain"
	"github.com/Chinsusu/ERP-v2/apps/api/internal/shared/auth"
	"github.com/Chinsusu/ERP-v2/apps/api/internal/shared/response"
)

type financeSummaryReportResponse struct {
	Metadata     reportMetadataResponse     `json:"metadata"`
	CurrencyCode string                     `json:"currency_code"`
	AR           financeSummaryARResponse   `json:"ar"`
	AP           financeSummaryAPResponse   `json:"ap"`
	COD          financeSummaryCODResponse  `json:"cod"`
	Cash         financeSummaryCashResponse `json:"cash"`
}

type financeSummaryARResponse struct {
	OpenCount         int                                 `json:"open_count"`
	OverdueCount      int                                 `json:"overdue_count"`
	DisputedCount     int                                 `json:"disputed_count"`
	OpenAmount        string                              `json:"open_amount"`
	OverdueAmount     string                              `json:"overdue_amount"`
	OutstandingAmount string                              `json:"outstanding_amount"`
	AgingBuckets      []financeSummaryAgingBucketResponse `json:"aging_buckets"`
}

type financeSummaryAPResponse struct {
	OpenCount             int                                 `json:"open_count"`
	DueCount              int                                 `json:"due_count"`
	PaymentRequestedCount int                                 `json:"payment_requested_count"`
	PaymentApprovedCount  int                                 `json:"payment_approved_count"`
	OpenAmount            string                              `json:"open_amount"`
	DueAmount             string                              `json:"due_amount"`
	OutstandingAmount     string                              `json:"outstanding_amount"`
	AgingBuckets          []financeSummaryAgingBucketResponse `json:"aging_buckets"`
}

type financeSummaryCODResponse struct {
	PendingCount       int                                       `json:"pending_count"`
	DiscrepancyCount   int                                       `json:"discrepancy_count"`
	PendingAmount      string                                    `json:"pending_amount"`
	DiscrepancyAmount  string                                    `json:"discrepancy_amount"`
	DiscrepancyBuckets []financeSummaryDiscrepancyBucketResponse `json:"discrepancy_buckets"`
}

type financeSummaryCashResponse struct {
	TransactionCount int    `json:"transaction_count"`
	CashInAmount     string `json:"cash_in_amount"`
	CashOutAmount    string `json:"cash_out_amount"`
	NetCashAmount    string `json:"net_cash_amount"`
}

type financeSummaryAgingBucketResponse struct {
	Bucket string `json:"bucket"`
	Count  int    `json:"count"`
	Amount string `json:"amount"`
}

type financeSummaryDiscrepancyBucketResponse struct {
	Type   string `json:"type"`
	Status string `json:"status"`
	Count  int    `json:"count"`
	Amount string `json:"amount"`
}

func financeSummaryReportHandler(
	receivables financeapp.CustomerReceivableStore,
	payables financeapp.SupplierPayableStore,
	remittances financeapp.CODRemittanceStore,
	cashTransactions financeapp.CashTransactionStore,
) http.HandlerFunc {
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
		if !auth.HasPermission(principal, auth.PermissionReportsView) {
			writePermissionDenied(w, r, auth.PermissionReportsView)
			return
		}
		if !auth.HasPermission(principal, auth.PermissionFinanceReportsView) {
			writePermissionDenied(w, r, auth.PermissionFinanceReportsView)
			return
		}

		filters, err := reportingdomain.NewReportFilters(reportFilterInputFromRequest(r))
		if err != nil {
			writeFinanceSummaryValidationError(w, r, "Invalid finance summary filters", "date")
			return
		}

		receivableRows, err := receivables.List(r.Context(), financeapp.CustomerReceivableFilter{})
		if err != nil {
			writeFinanceSummaryStoreError(w, r)
			return
		}
		payableRows, err := payables.List(r.Context(), financeapp.SupplierPayableFilter{})
		if err != nil {
			writeFinanceSummaryStoreError(w, r)
			return
		}
		remittanceRows, err := remittances.List(r.Context(), financeapp.CODRemittanceFilter{})
		if err != nil {
			writeFinanceSummaryStoreError(w, r)
			return
		}
		cashRows, err := cashTransactions.List(r.Context(), financeapp.CashTransactionFilter{})
		if err != nil {
			writeFinanceSummaryStoreError(w, r)
			return
		}

		report, err := reportingdomain.NewFinanceSummaryReport(
			filters,
			receivableRows,
			payableRows,
			remittanceRows,
			cashRows,
			reportingdomain.FinanceSummaryOptions{},
		)
		if err != nil {
			writeFinanceSummaryReportError(w, r, err)
			return
		}

		response.WriteSuccess(w, r, http.StatusOK, newFinanceSummaryReportResponse(report))
	}
}

func newFinanceSummaryReportResponse(report reportingdomain.FinanceSummaryReport) financeSummaryReportResponse {
	return financeSummaryReportResponse{
		Metadata:     newReportMetadataResponse(report.Metadata),
		CurrencyCode: report.CurrencyCode,
		AR: financeSummaryARResponse{
			OpenCount:         report.AR.OpenCount,
			OverdueCount:      report.AR.OverdueCount,
			DisputedCount:     report.AR.DisputedCount,
			OpenAmount:        report.AR.OpenAmount,
			OverdueAmount:     report.AR.OverdueAmount,
			OutstandingAmount: report.AR.OutstandingAmount,
			AgingBuckets:      newFinanceSummaryAgingBucketResponses(report.AR.AgingBuckets),
		},
		AP: financeSummaryAPResponse{
			OpenCount:             report.AP.OpenCount,
			DueCount:              report.AP.DueCount,
			PaymentRequestedCount: report.AP.PaymentRequestedCount,
			PaymentApprovedCount:  report.AP.PaymentApprovedCount,
			OpenAmount:            report.AP.OpenAmount,
			DueAmount:             report.AP.DueAmount,
			OutstandingAmount:     report.AP.OutstandingAmount,
			AgingBuckets:          newFinanceSummaryAgingBucketResponses(report.AP.AgingBuckets),
		},
		COD: financeSummaryCODResponse{
			PendingCount:       report.COD.PendingCount,
			DiscrepancyCount:   report.COD.DiscrepancyCount,
			PendingAmount:      report.COD.PendingAmount,
			DiscrepancyAmount:  report.COD.DiscrepancyAmount,
			DiscrepancyBuckets: newFinanceSummaryDiscrepancyBucketResponses(report.COD.DiscrepancyBuckets),
		},
		Cash: financeSummaryCashResponse{
			TransactionCount: report.Cash.TransactionCount,
			CashInAmount:     report.Cash.CashInAmount,
			CashOutAmount:    report.Cash.CashOutAmount,
			NetCashAmount:    report.Cash.NetCashAmount,
		},
	}
}

func newFinanceSummaryAgingBucketResponses(
	buckets []reportingdomain.FinanceSummaryAgingBucket,
) []financeSummaryAgingBucketResponse {
	rows := make([]financeSummaryAgingBucketResponse, 0, len(buckets))
	for _, bucket := range buckets {
		rows = append(rows, financeSummaryAgingBucketResponse{
			Bucket: bucket.Bucket,
			Count:  bucket.Count,
			Amount: bucket.Amount,
		})
	}

	return rows
}

func newFinanceSummaryDiscrepancyBucketResponses(
	buckets []reportingdomain.FinanceSummaryDiscrepancyBucket,
) []financeSummaryDiscrepancyBucketResponse {
	rows := make([]financeSummaryDiscrepancyBucketResponse, 0, len(buckets))
	for _, bucket := range buckets {
		rows = append(rows, financeSummaryDiscrepancyBucketResponse{
			Type:   bucket.Type,
			Status: bucket.Status,
			Count:  bucket.Count,
			Amount: bucket.Amount,
		})
	}

	return rows
}

func writeFinanceSummaryReportError(w http.ResponseWriter, r *http.Request, err error) {
	if errors.Is(err, reportingdomain.ErrInvalidFinanceSummaryReport) {
		writeFinanceSummaryValidationError(w, r, "Finance summary report is invalid", "filter")
		return
	}

	writeFinanceSummaryStoreError(w, r)
}

func writeFinanceSummaryStoreError(w http.ResponseWriter, r *http.Request) {
	response.WriteError(
		w,
		r,
		http.StatusConflict,
		response.ErrorCodeConflict,
		"Finance summary report could not be calculated",
		nil,
	)
}

func writeFinanceSummaryValidationError(w http.ResponseWriter, r *http.Request, message string, field string) {
	response.WriteError(
		w,
		r,
		http.StatusBadRequest,
		response.ErrorCodeValidation,
		message,
		map[string]any{"field": field},
	)
}
