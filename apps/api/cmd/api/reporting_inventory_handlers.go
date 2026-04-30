package main

import (
	"errors"
	"net/http"
	"strconv"
	"strings"

	inventoryapp "github.com/Chinsusu/ERP-v2/apps/api/internal/modules/inventory/application"
	inventorydomain "github.com/Chinsusu/ERP-v2/apps/api/internal/modules/inventory/domain"
	reportingdomain "github.com/Chinsusu/ERP-v2/apps/api/internal/modules/reporting/domain"
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
	WarehouseID      string `json:"warehouse_id"`
	WarehouseCode    string `json:"warehouse_code,omitempty"`
	LocationID       string `json:"location_id,omitempty"`
	LocationCode     string `json:"location_code,omitempty"`
	ItemID           string `json:"item_id,omitempty"`
	SKU              string `json:"sku"`
	BatchID          string `json:"batch_id,omitempty"`
	BatchNo          string `json:"batch_no,omitempty"`
	BatchExpiry      string `json:"batch_expiry,omitempty"`
	BaseUOMCode      string `json:"base_uom_code"`
	PhysicalQty      string `json:"physical_qty"`
	ReservedQty      string `json:"reserved_qty"`
	QuarantineQty    string `json:"quarantine_qty"`
	BlockedQty       string `json:"blocked_qty"`
	AvailableQty     string `json:"available_qty"`
	LowStock         bool   `json:"low_stock"`
	ExpiryWarning    bool   `json:"expiry_warning"`
	Expired          bool   `json:"expired"`
	BatchQCStatus    string `json:"batch_qc_status,omitempty"`
	BatchStatus      string `json:"batch_status,omitempty"`
	SourceStockState string `json:"source_stock_state"`
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

		filters, err := reportingdomain.NewReportFilters(reportFilterInputFromRequest(r))
		if err != nil {
			writeInventorySnapshotValidationError(w, r, "Invalid inventory snapshot filters", "date")
			return
		}
		expiryWarningDays, err := parseInventorySnapshotPositiveInt(r.URL.Query().Get("expiry_warning_days"))
		if err != nil {
			writeInventorySnapshotValidationError(w, r, "Invalid inventory snapshot filters", "expiry_warning_days")
			return
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
			return
		}

		snapshots, ok = filterInventorySnapshotStatus(snapshots, filters.Status)
		if !ok {
			writeInventorySnapshotValidationError(w, r, "Invalid inventory snapshot filters", "status")
			return
		}

		report, err := reportingdomain.NewInventorySnapshotReport(filters, snapshots, reportingdomain.InventorySnapshotOptions{
			LowStockThreshold: r.URL.Query().Get("low_stock_threshold"),
			ExpiryWarningDays: expiryWarningDays,
		})
		if err != nil {
			writeInventorySnapshotReportError(w, r, err)
			return
		}

		response.WriteSuccess(w, r, http.StatusOK, newInventorySnapshotReportResponse(report))
	}
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
	}
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
