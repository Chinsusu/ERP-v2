package main

import (
	"errors"
	"net/http"
	"strings"
	"time"

	reportingdomain "github.com/Chinsusu/ERP-v2/apps/api/internal/modules/reporting/domain"
	reportinghandler "github.com/Chinsusu/ERP-v2/apps/api/internal/modules/reporting/handler"
	"github.com/Chinsusu/ERP-v2/apps/api/internal/shared/auth"
	"github.com/Chinsusu/ERP-v2/apps/api/internal/shared/response"
)

type operationsDailyReportResponse struct {
	Metadata reportMetadataResponse               `json:"metadata"`
	Summary  operationsDailySummaryResponse       `json:"summary"`
	Areas    []operationsDailyAreaSummaryResponse `json:"areas"`
	Rows     []operationsDailyReportRowResponse   `json:"rows"`
}

type operationsDailySummaryResponse struct {
	SignalCount     int `json:"signal_count"`
	PendingCount    int `json:"pending_count"`
	InProgressCount int `json:"in_progress_count"`
	CompletedCount  int `json:"completed_count"`
	BlockedCount    int `json:"blocked_count"`
	ExceptionCount  int `json:"exception_count"`
}

type operationsDailyAreaSummaryResponse struct {
	Area            string `json:"area"`
	SignalCount     int    `json:"signal_count"`
	PendingCount    int    `json:"pending_count"`
	InProgressCount int    `json:"in_progress_count"`
	CompletedCount  int    `json:"completed_count"`
	BlockedCount    int    `json:"blocked_count"`
	ExceptionCount  int    `json:"exception_count"`
}

type operationsDailyReportRowResponse struct {
	ID              string                        `json:"id"`
	Area            string                        `json:"area"`
	SourceType      string                        `json:"source_type"`
	SourceID        string                        `json:"source_id"`
	SourceReference reportSourceReferenceResponse `json:"source_reference"`
	RefNo           string                        `json:"ref_no"`
	Title           string                        `json:"title"`
	WarehouseID     string                        `json:"warehouse_id"`
	WarehouseCode   string                        `json:"warehouse_code,omitempty"`
	BusinessDate    string                        `json:"business_date"`
	Status          string                        `json:"status"`
	Severity        string                        `json:"severity"`
	Quantity        string                        `json:"quantity,omitempty"`
	UOMCode         string                        `json:"uom_code,omitempty"`
	ExceptionCode   string                        `json:"exception_code,omitempty"`
	Owner           string                        `json:"owner,omitempty"`
}

var operationsDailyCSVHeaders = []string{
	"id",
	"area",
	"source_type",
	"source_id",
	"ref_no",
	"title",
	"warehouse_id",
	"warehouse_code",
	"business_date",
	"status",
	"severity",
	"quantity",
	"uom_code",
	"exception_code",
	"owner",
}

func operationsDailyReportHandler(source operationsDailySignalSource) http.HandlerFunc {
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

		report, ok := operationsDailyReportFromRequest(w, r, source)
		if !ok {
			return
		}

		response.WriteSuccess(w, r, http.StatusOK, newOperationsDailyReportResponse(report))
	}
}

func operationsDailyCSVExportHandler(source operationsDailySignalSource) http.HandlerFunc {
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
		if !auth.HasPermission(principal, auth.PermissionReportsExport) {
			writePermissionDenied(w, r, auth.PermissionReportsExport)
			return
		}

		report, ok := operationsDailyReportFromRequest(w, r, source)
		if !ok {
			return
		}

		err := reportinghandler.WriteCSV(w, r, reportinghandler.CSVExport{
			Filename: operationsDailyCSVFilename(report),
			Headers:  operationsDailyCSVHeaders,
			Rows:     newOperationsDailyCSVRows(report),
		})
		if err != nil {
			response.WriteError(
				w,
				r,
				http.StatusConflict,
				response.ErrorCodeConflict,
				"Operations daily CSV could not be exported",
				nil,
			)
		}
	}
}

func operationsDailyReportFromRequest(
	w http.ResponseWriter,
	r *http.Request,
	source operationsDailySignalSource,
) (reportingdomain.OperationsDailyReport, bool) {
	filters, err := reportingdomain.NewReportFilters(reportFilterInputFromRequest(r))
	if err != nil {
		writeOperationsDailyValidationError(w, r, "Invalid operations daily filters", "date")
		return reportingdomain.OperationsDailyReport{}, false
	}
	status, ok := normalizeOperationsDailyStatusFilter(filters.Status)
	if !ok {
		writeOperationsDailyValidationError(w, r, "Invalid operations daily filters", "status")
		return reportingdomain.OperationsDailyReport{}, false
	}
	filters.Status = status

	if source == nil {
		source = prototypeOperationsDailySignalSource{}
	}
	signals, err := source.ListOperationsDailySignals(r.Context(), filters)
	if err != nil {
		writeOperationsDailyReportError(w, r, err)
		return reportingdomain.OperationsDailyReport{}, false
	}

	report, err := reportingdomain.NewOperationsDailyReport(
		filters,
		signals,
		reportingdomain.OperationsDailyOptions{},
	)
	if err != nil {
		writeOperationsDailyReportError(w, r, err)
		return reportingdomain.OperationsDailyReport{}, false
	}

	return report, true
}

func normalizeOperationsDailyStatusFilter(status string) (string, bool) {
	normalized := strings.ToLower(strings.TrimSpace(status))
	switch normalized {
	case "", "all":
		return "", true
	case "pending", "in_progress", "completed", "blocked", "exception":
		return normalized, true
	default:
		return "", false
	}
}

func prototypeOperationsDailySignals() []reportingdomain.OperationsDailySignal {
	businessDate := time.Date(2026, 4, 30, 0, 0, 0, 0, reportingdomain.HoChiMinhLocation())
	hanoiDate := time.Date(2026, 4, 30, 0, 0, 0, 0, reportingdomain.HoChiMinhLocation())

	return []reportingdomain.OperationsDailySignal{
		{
			ID:            "ops-inbound-hcm-260430-0001",
			Area:          reportingdomain.OperationsDailyAreaInbound,
			SourceType:    "goods_receipt",
			SourceID:      "gr-260430-0001",
			RefNo:         "GR-260430-0001",
			Title:         "Supplier delivery awaiting receiving check",
			WarehouseID:   "wh-hcm",
			WarehouseCode: "HCM",
			BusinessDate:  businessDate,
			Status:        reportingdomain.OperationsDailyStatusPending,
			Severity:      reportingdomain.OperationsDailySeverityWarning,
			Quantity:      "12.000000",
			UOMCode:       "PCS",
			Owner:         "warehouse",
		},
		{
			ID:            "ops-qc-hcm-260430-0001",
			Area:          reportingdomain.OperationsDailyAreaQC,
			SourceType:    "inbound_qc",
			SourceID:      "iqc-260430-fail",
			RefNo:         "IQC-260430-FAIL",
			Title:         "Inbound QC failed for damaged packaging",
			WarehouseID:   "wh-hcm",
			WarehouseCode: "HCM",
			BusinessDate:  businessDate,
			Status:        reportingdomain.OperationsDailyStatusException,
			Severity:      reportingdomain.OperationsDailySeverityDanger,
			ExceptionCode: "QC_FAIL",
			Owner:         "qa",
		},
		{
			ID:            "ops-outbound-hcm-260430-0001",
			Area:          reportingdomain.OperationsDailyAreaOutbound,
			SourceType:    "pick_task",
			SourceID:      "pick-260430-0001",
			RefNo:         "PICK-260430-0001",
			Title:         "Pick wave in progress for ecommerce orders",
			WarehouseID:   "wh-hcm",
			WarehouseCode: "HCM",
			BusinessDate:  businessDate,
			Status:        reportingdomain.OperationsDailyStatusInProgress,
			Severity:      reportingdomain.OperationsDailySeverityNormal,
			Quantity:      "24.000000",
			UOMCode:       "PCS",
			Owner:         "warehouse",
		},
		{
			ID:            "ops-outbound-hcm-260430-0002",
			Area:          reportingdomain.OperationsDailyAreaOutbound,
			SourceType:    "carrier_manifest",
			SourceID:      "manifest-260430-ghn",
			RefNo:         "MAN-260430-GHN",
			Title:         "Carrier handover missing scan",
			WarehouseID:   "wh-hcm",
			WarehouseCode: "HCM",
			BusinessDate:  businessDate,
			Status:        reportingdomain.OperationsDailyStatusBlocked,
			Severity:      reportingdomain.OperationsDailySeverityDanger,
			ExceptionCode: "MISSING_HANDOVER_SCAN",
			Owner:         "shipping",
		},
		{
			ID:            "ops-returns-hcm-260430-0001",
			Area:          reportingdomain.OperationsDailyAreaReturns,
			SourceType:    "return_receipt",
			SourceID:      "ret-260430-0001",
			RefNo:         "RET-260430-0001",
			Title:         "Return receipt awaiting inspection",
			WarehouseID:   "wh-hcm",
			WarehouseCode: "HCM",
			BusinessDate:  businessDate,
			Status:        reportingdomain.OperationsDailyStatusPending,
			Severity:      reportingdomain.OperationsDailySeverityWarning,
			Quantity:      "3.000000",
			UOMCode:       "PCS",
			Owner:         "returns",
		},
		{
			ID:            "ops-stock-hcm-260430-0001",
			Area:          reportingdomain.OperationsDailyAreaStock,
			SourceType:    "stock_count",
			SourceID:      "count-260430-0001",
			RefNo:         "CNT-260430-0001",
			Title:         "Cycle count variance needs review",
			WarehouseID:   "wh-hcm",
			WarehouseCode: "HCM",
			BusinessDate:  businessDate,
			Status:        reportingdomain.OperationsDailyStatusBlocked,
			Severity:      reportingdomain.OperationsDailySeverityDanger,
			ExceptionCode: "VARIANCE_REVIEW",
			Owner:         "warehouse_lead",
		},
		{
			ID:            "ops-subcontract-hcm-260430-0001",
			Area:          reportingdomain.OperationsDailyAreaSubcontract,
			SourceType:    "subcontract_order",
			SourceID:      "sco-260430-0001",
			RefNo:         "SCO-260430-0001",
			Title:         "Material issue in progress for subcontract factory",
			WarehouseID:   "wh-hcm",
			WarehouseCode: "HCM",
			BusinessDate:  businessDate,
			Status:        reportingdomain.OperationsDailyStatusInProgress,
			Severity:      reportingdomain.OperationsDailySeverityNormal,
			Quantity:      "80.000000",
			UOMCode:       "PCS",
			Owner:         "production",
		},
		{
			ID:            "ops-outbound-hn-260430-0001",
			Area:          reportingdomain.OperationsDailyAreaOutbound,
			SourceType:    "carrier_manifest",
			SourceID:      "manifest-260430-hn",
			RefNo:         "MAN-260430-HN",
			Title:         "Hanoi carrier handover completed",
			WarehouseID:   "wh-hn",
			WarehouseCode: "HN",
			BusinessDate:  hanoiDate,
			Status:        reportingdomain.OperationsDailyStatusCompleted,
			Severity:      reportingdomain.OperationsDailySeverityNormal,
			Owner:         "shipping",
		},
	}
}

func newOperationsDailyReportResponse(report reportingdomain.OperationsDailyReport) operationsDailyReportResponse {
	areas := make([]operationsDailyAreaSummaryResponse, 0, len(report.Areas))
	for _, area := range report.Areas {
		areas = append(areas, operationsDailyAreaSummaryResponse{
			Area:            area.Area,
			SignalCount:     area.SignalCount,
			PendingCount:    area.PendingCount,
			InProgressCount: area.InProgressCount,
			CompletedCount:  area.CompletedCount,
			BlockedCount:    area.BlockedCount,
			ExceptionCount:  area.ExceptionCount,
		})
	}

	rows := make([]operationsDailyReportRowResponse, 0, len(report.Rows))
	for _, row := range report.Rows {
		rows = append(rows, operationsDailyReportRowResponse{
			ID:              row.ID,
			Area:            row.Area,
			SourceType:      row.SourceType,
			SourceID:        row.SourceID,
			SourceReference: newReportSourceReferenceResponse(row.SourceReference),
			RefNo:           row.RefNo,
			Title:           row.Title,
			WarehouseID:     row.WarehouseID,
			WarehouseCode:   row.WarehouseCode,
			BusinessDate:    row.BusinessDate,
			Status:          row.Status,
			Severity:        row.Severity,
			Quantity:        row.Quantity,
			UOMCode:         row.UOMCode,
			ExceptionCode:   row.ExceptionCode,
			Owner:           row.Owner,
		})
	}

	return operationsDailyReportResponse{
		Metadata: newReportMetadataResponse(report.Metadata),
		Summary: operationsDailySummaryResponse{
			SignalCount:     report.Summary.SignalCount,
			PendingCount:    report.Summary.PendingCount,
			InProgressCount: report.Summary.InProgressCount,
			CompletedCount:  report.Summary.CompletedCount,
			BlockedCount:    report.Summary.BlockedCount,
			ExceptionCount:  report.Summary.ExceptionCount,
		},
		Areas: areas,
		Rows:  rows,
	}
}

func operationsDailyCSVFilename(report reportingdomain.OperationsDailyReport) string {
	filters := report.Metadata.Filters
	return "operations-daily-" + filters.FromDateString() + "-to-" + filters.ToDateString() + ".csv"
}

func newOperationsDailyCSVRows(report reportingdomain.OperationsDailyReport) [][]string {
	rows := make([][]string, 0, len(report.Rows))
	for _, row := range report.Rows {
		rows = append(rows, []string{
			row.ID,
			row.Area,
			row.SourceType,
			row.SourceID,
			row.RefNo,
			row.Title,
			row.WarehouseID,
			row.WarehouseCode,
			row.BusinessDate,
			row.Status,
			row.Severity,
			row.Quantity,
			row.UOMCode,
			row.ExceptionCode,
			row.Owner,
		})
	}

	return rows
}

func writeOperationsDailyReportError(w http.ResponseWriter, r *http.Request, err error) {
	if errors.Is(err, reportingdomain.ErrInvalidOperationsDailyReport) {
		writeOperationsDailyValidationError(w, r, "Operations daily report is invalid", "filter")
		return
	}

	response.WriteError(
		w,
		r,
		http.StatusConflict,
		response.ErrorCodeConflict,
		"Operations daily report could not be calculated",
		nil,
	)
}

func writeOperationsDailyValidationError(w http.ResponseWriter, r *http.Request, message string, field string) {
	response.WriteError(
		w,
		r,
		http.StatusBadRequest,
		response.ErrorCodeValidation,
		message,
		map[string]any{"field": field},
	)
}
