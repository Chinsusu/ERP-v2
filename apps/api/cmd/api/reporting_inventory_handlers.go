package main

import (
	"errors"
	"net/http"
	"strconv"
	"strings"

	inventoryapp "github.com/Chinsusu/ERP-v2/apps/api/internal/modules/inventory/application"
	inventorydomain "github.com/Chinsusu/ERP-v2/apps/api/internal/modules/inventory/domain"
	reportingdomain "github.com/Chinsusu/ERP-v2/apps/api/internal/modules/reporting/domain"
	reportinghandler "github.com/Chinsusu/ERP-v2/apps/api/internal/modules/reporting/handler"
	"github.com/Chinsusu/ERP-v2/apps/api/internal/shared/auth"
	"github.com/Chinsusu/ERP-v2/apps/api/internal/shared/response"
)

type reportMetadataResponse struct {
	GeneratedAt   string                `json:"generated_at"`
	Timezone      string                `json:"timezone"`
	SourceVersion string                `json:"source_version"`
	Filters       reportFiltersResponse `json:"filters"`
}

type reportFiltersResponse struct {
	FromDate     string `json:"from_date"`
	ToDate       string `json:"to_date"`
	BusinessDate string `json:"business_date"`
	WarehouseID  string `json:"warehouse_id,omitempty"`
	Status       string `json:"status,omitempty"`
	ItemID       string `json:"item_id,omitempty"`
	Category     string `json:"category,omitempty"`
}

type reportSourceReferenceResponse struct {
	EntityType  string `json:"entity_type"`
	ID          string `json:"id"`
	Label       string `json:"label"`
	Href        string `json:"href,omitempty"`
	Unavailable bool   `json:"unavailable"`
}

type inventorySnapshotReportResponse struct {
	Metadata reportMetadataResponse           `json:"metadata"`
	Summary  inventorySnapshotSummaryResponse `json:"summary"`
	Rows     []inventorySnapshotRowResponse   `json:"rows"`
}

type inventorySnapshotSummaryResponse struct {
	RowCount          int                                 `json:"row_count"`
	LowStockRowCount  int                                 `json:"low_stock_row_count"`
	ExpiryWarningRows int                                 `json:"expiry_warning_rows"`
	ExpiredRows       int                                 `json:"expired_rows"`
	TotalsByUOM       []inventorySnapshotUOMTotalResponse `json:"totals_by_uom"`
}

type inventorySnapshotUOMTotalResponse struct {
	BaseUOMCode   string `json:"base_uom_code"`
	PhysicalQty   string `json:"physical_qty"`
	ReservedQty   string `json:"reserved_qty"`
	QuarantineQty string `json:"quarantine_qty"`
	BlockedQty    string `json:"blocked_qty"`
	AvailableQty  string `json:"available_qty"`
}

type inventorySnapshotRowResponse struct {
	WarehouseID      string                          `json:"warehouse_id"`
	WarehouseCode    string                          `json:"warehouse_code,omitempty"`
	LocationID       string                          `json:"location_id,omitempty"`
	LocationCode     string                          `json:"location_code,omitempty"`
	ItemID           string                          `json:"item_id,omitempty"`
	SKU              string                          `json:"sku"`
	BatchID          string                          `json:"batch_id,omitempty"`
	BatchNo          string                          `json:"batch_no,omitempty"`
	BatchExpiry      string                          `json:"batch_expiry,omitempty"`
	BaseUOMCode      string                          `json:"base_uom_code"`
	PhysicalQty      string                          `json:"physical_qty"`
	ReservedQty      string                          `json:"reserved_qty"`
	QuarantineQty    string                          `json:"quarantine_qty"`
	BlockedQty       string                          `json:"blocked_qty"`
	AvailableQty     string                          `json:"available_qty"`
	LowStock         bool                            `json:"low_stock"`
	ExpiryWarning    bool                            `json:"expiry_warning"`
	Expired          bool                            `json:"expired"`
	BatchQCStatus    string                          `json:"batch_qc_status,omitempty"`
	BatchStatus      string                          `json:"batch_status,omitempty"`
	SourceStockState string                          `json:"source_stock_state"`
	SourceReferences []reportSourceReferenceResponse `json:"source_references"`
}

var inventorySnapshotCSVHeaders = []string{
	"warehouse_id",
	"warehouse_code",
	"location_id",
	"location_code",
	"item_id",
	"sku",
	"batch_id",
	"batch_no",
	"batch_expiry",
	"base_uom_code",
	"physical_qty",
	"reserved_qty",
	"quarantine_qty",
	"blocked_qty",
	"available_qty",
	"low_stock",
	"expiry_warning",
	"expired",
	"batch_qc_status",
	"batch_status",
	"source_stock_state",
}

func inventorySnapshotReportHandler(service inventoryapp.ListAvailableStock) http.HandlerFunc {
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

		report, ok := inventorySnapshotReportFromRequest(w, r, service)
		if !ok {
			return
		}

		response.WriteSuccess(w, r, http.StatusOK, newInventorySnapshotReportResponse(report))
	}
}

func inventorySnapshotCSVExportHandler(service inventoryapp.ListAvailableStock) http.HandlerFunc {
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

		report, ok := inventorySnapshotReportFromRequest(w, r, service)
		if !ok {
			return
		}

		err := reportinghandler.WriteCSV(w, r, reportinghandler.CSVExport{
			Filename: "inventory-snapshot-" + report.Metadata.Filters.BusinessDateString() + ".csv",
			Headers:  inventorySnapshotCSVHeaders,
			Rows:     newInventorySnapshotCSVRows(report),
		})
		if err != nil {
			response.WriteError(
				w,
				r,
				http.StatusConflict,
				response.ErrorCodeConflict,
				"Inventory snapshot CSV could not be exported",
				nil,
			)
			return
		}
	}
}

func inventorySnapshotReportFromRequest(
	w http.ResponseWriter,
	r *http.Request,
	service inventoryapp.ListAvailableStock,
) (reportingdomain.InventorySnapshotReport, bool) {
	filters, err := reportingdomain.NewReportFilters(reportFilterInputFromRequest(r))
	if err != nil {
		writeInventorySnapshotValidationError(w, r, "Invalid inventory snapshot filters", "date")
		return reportingdomain.InventorySnapshotReport{}, false
	}
	expiryWarningDays, err := parseInventorySnapshotPositiveInt(r.URL.Query().Get("expiry_warning_days"))
	if err != nil {
		writeInventorySnapshotValidationError(w, r, "Invalid inventory snapshot filters", "expiry_warning_days")
		return reportingdomain.InventorySnapshotReport{}, false
	}

	stockFilter := inventorydomain.NewAvailableStockFilter(
		filters.WarehouseID,
		r.URL.Query().Get("location_id"),
		r.URL.Query().Get("sku"),
		r.URL.Query().Get("batch_id"),
	)
	stockFilter.ItemID = filters.ItemID

	snapshots, err := service.Execute(r.Context(), stockFilter)
	if err != nil {
		response.WriteError(
			w,
			r,
			http.StatusConflict,
			response.ErrorCodeConflict,
			"Inventory snapshot report could not be calculated",
			nil,
		)
		return reportingdomain.InventorySnapshotReport{}, false
	}

	snapshots, ok := filterInventorySnapshotStatus(snapshots, filters.Status)
	if !ok {
		writeInventorySnapshotValidationError(w, r, "Invalid inventory snapshot filters", "status")
		return reportingdomain.InventorySnapshotReport{}, false
	}

	report, err := reportingdomain.NewInventorySnapshotReport(filters, snapshots, reportingdomain.InventorySnapshotOptions{
		LowStockThreshold: r.URL.Query().Get("low_stock_threshold"),
		ExpiryWarningDays: expiryWarningDays,
	})
	if err != nil {
		writeInventorySnapshotReportError(w, r, err)
		return reportingdomain.InventorySnapshotReport{}, false
	}

	return report, true
}

func reportFilterInputFromRequest(r *http.Request) reportingdomain.ReportFilterInput {
	return reportingdomain.ReportFilterInput{
		FromDate:     r.URL.Query().Get("from_date"),
		ToDate:       r.URL.Query().Get("to_date"),
		BusinessDate: r.URL.Query().Get("business_date"),
		WarehouseID:  r.URL.Query().Get("warehouse_id"),
		Status:       r.URL.Query().Get("status"),
		ItemID:       r.URL.Query().Get("item_id"),
		Category:     r.URL.Query().Get("category"),
	}
}

func filterInventorySnapshotStatus(
	snapshots []inventorydomain.AvailableStockSnapshot,
	status string,
) ([]inventorydomain.AvailableStockSnapshot, bool) {
	normalized := strings.ToLower(strings.TrimSpace(status))
	if normalized == "" || normalized == "all" {
		return snapshots, true
	}
	if normalized != "available" &&
		normalized != "reserved" &&
		normalized != "quarantine" &&
		normalized != "qc_hold" &&
		normalized != "blocked" {
		return nil, false
	}

	filtered := make([]inventorydomain.AvailableStockSnapshot, 0, len(snapshots))
	for _, snapshot := range snapshots {
		if inventorySnapshotStatusMatches(snapshot, normalized) {
			filtered = append(filtered, snapshot)
		}
	}

	return filtered, true
}

func inventorySnapshotStatusMatches(snapshot inventorydomain.AvailableStockSnapshot, status string) bool {
	switch status {
	case "available":
		return snapshot.QCHoldQty.IsZero() && snapshot.BlockedQty.IsZero() && snapshot.ReservedQty.IsZero()
	case "reserved":
		return snapshot.QCHoldQty.IsZero() && snapshot.BlockedQty.IsZero() && !snapshot.ReservedQty.IsZero()
	case "quarantine", "qc_hold":
		return !snapshot.QCHoldQty.IsZero()
	case "blocked":
		return snapshot.QCHoldQty.IsZero() && !snapshot.BlockedQty.IsZero()
	default:
		return false
	}
}

func parseInventorySnapshotPositiveInt(value string) (int, error) {
	raw := strings.TrimSpace(value)
	if raw == "" {
		return 0, nil
	}
	parsed, err := strconv.Atoi(raw)
	if err != nil || parsed <= 0 {
		return 0, reportingdomain.ErrInvalidInventorySnapshotReport
	}

	return parsed, nil
}

func newInventorySnapshotReportResponse(report reportingdomain.InventorySnapshotReport) inventorySnapshotReportResponse {
	totals := make([]inventorySnapshotUOMTotalResponse, 0, len(report.Summary.TotalsByUOM))
	for _, total := range report.Summary.TotalsByUOM {
		totals = append(totals, inventorySnapshotUOMTotalResponse{
			BaseUOMCode:   total.BaseUOMCode,
			PhysicalQty:   total.PhysicalQty,
			ReservedQty:   total.ReservedQty,
			QuarantineQty: total.QuarantineQty,
			BlockedQty:    total.BlockedQty,
			AvailableQty:  total.AvailableQty,
		})
	}

	rows := make([]inventorySnapshotRowResponse, 0, len(report.Rows))
	for _, row := range report.Rows {
		rows = append(rows, newInventorySnapshotRowResponse(row))
	}

	return inventorySnapshotReportResponse{
		Metadata: newReportMetadataResponse(report.Metadata),
		Summary: inventorySnapshotSummaryResponse{
			RowCount:          report.Summary.RowCount,
			LowStockRowCount:  report.Summary.LowStockRowCount,
			ExpiryWarningRows: report.Summary.ExpiryWarningRows,
			ExpiredRows:       report.Summary.ExpiredRows,
			TotalsByUOM:       totals,
		},
		Rows: rows,
	}
}

func newReportMetadataResponse(metadata reportingdomain.ReportMetadata) reportMetadataResponse {
	return reportMetadataResponse{
		GeneratedAt:   timeString(metadata.GeneratedAt),
		Timezone:      metadata.Timezone,
		SourceVersion: metadata.SourceVersion,
		Filters: reportFiltersResponse{
			FromDate:     metadata.Filters.FromDateString(),
			ToDate:       metadata.Filters.ToDateString(),
			BusinessDate: metadata.Filters.BusinessDateString(),
			WarehouseID:  metadata.Filters.WarehouseID,
			Status:       metadata.Filters.Status,
			ItemID:       metadata.Filters.ItemID,
			Category:     metadata.Filters.Category,
		},
	}
}

func newReportSourceReferenceResponse(reference reportingdomain.ReportSourceReference) reportSourceReferenceResponse {
	return reportSourceReferenceResponse{
		EntityType:  reference.EntityType,
		ID:          reference.ID,
		Label:       reference.Label,
		Href:        reference.Href,
		Unavailable: reference.Unavailable,
	}
}

func newInventorySnapshotRowResponse(row reportingdomain.InventorySnapshotRow) inventorySnapshotRowResponse {
	return inventorySnapshotRowResponse{
		WarehouseID:      row.WarehouseID,
		WarehouseCode:    row.WarehouseCode,
		LocationID:       row.LocationID,
		LocationCode:     row.LocationCode,
		ItemID:           row.ItemID,
		SKU:              row.SKU,
		BatchID:          row.BatchID,
		BatchNo:          row.BatchNo,
		BatchExpiry:      row.BatchExpiry,
		BaseUOMCode:      row.BaseUOMCode,
		PhysicalQty:      row.PhysicalQty,
		ReservedQty:      row.ReservedQty,
		QuarantineQty:    row.QuarantineQty,
		BlockedQty:       row.BlockedQty,
		AvailableQty:     row.AvailableQty,
		LowStock:         row.LowStock,
		ExpiryWarning:    row.ExpiryWarning,
		Expired:          row.Expired,
		BatchQCStatus:    row.BatchQCStatus,
		BatchStatus:      row.BatchStatus,
		SourceStockState: row.SourceStockState,
		SourceReferences: newReportSourceReferenceResponses(row.SourceReferences),
	}
}

func newReportSourceReferenceResponses(
	references []reportingdomain.ReportSourceReference,
) []reportSourceReferenceResponse {
	rows := make([]reportSourceReferenceResponse, 0, len(references))
	for _, reference := range references {
		rows = append(rows, newReportSourceReferenceResponse(reference))
	}

	return rows
}

func newInventorySnapshotCSVRows(report reportingdomain.InventorySnapshotReport) [][]string {
	rows := make([][]string, 0, len(report.Rows))
	for _, row := range report.Rows {
		rows = append(rows, []string{
			row.WarehouseID,
			row.WarehouseCode,
			row.LocationID,
			row.LocationCode,
			row.ItemID,
			row.SKU,
			row.BatchID,
			row.BatchNo,
			row.BatchExpiry,
			row.BaseUOMCode,
			row.PhysicalQty,
			row.ReservedQty,
			row.QuarantineQty,
			row.BlockedQty,
			row.AvailableQty,
			strconv.FormatBool(row.LowStock),
			strconv.FormatBool(row.ExpiryWarning),
			strconv.FormatBool(row.Expired),
			row.BatchQCStatus,
			row.BatchStatus,
			row.SourceStockState,
		})
	}

	return rows
}

func writeInventorySnapshotReportError(w http.ResponseWriter, r *http.Request, err error) {
	if errors.Is(err, reportingdomain.ErrInvalidInventorySnapshotReport) {
		writeInventorySnapshotValidationError(w, r, "Inventory snapshot report is invalid", "filter")
		return
	}

	response.WriteError(
		w,
		r,
		http.StatusConflict,
		response.ErrorCodeConflict,
		"Inventory snapshot report could not be calculated",
		nil,
	)
}

func writeInventorySnapshotValidationError(w http.ResponseWriter, r *http.Request, message string, field string) {
	response.WriteError(
		w,
		r,
		http.StatusBadRequest,
		response.ErrorCodeValidation,
		message,
		map[string]any{"field": field},
	)
}
