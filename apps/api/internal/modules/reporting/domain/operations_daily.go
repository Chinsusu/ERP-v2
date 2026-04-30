package domain

import (
	"errors"
	"sort"
	"strings"
	"time"

	"github.com/Chinsusu/ERP-v2/apps/api/internal/shared/decimal"
)

type OperationsDailyArea string
type OperationsDailyStatus string
type OperationsDailySeverity string

const (
	OperationsDailyAreaInbound     OperationsDailyArea = "inbound"
	OperationsDailyAreaQC          OperationsDailyArea = "qc"
	OperationsDailyAreaOutbound    OperationsDailyArea = "outbound"
	OperationsDailyAreaReturns     OperationsDailyArea = "returns"
	OperationsDailyAreaStock       OperationsDailyArea = "stock_count"
	OperationsDailyAreaSubcontract OperationsDailyArea = "subcontract"
)

const (
	OperationsDailyStatusPending    OperationsDailyStatus = "pending"
	OperationsDailyStatusInProgress OperationsDailyStatus = "in_progress"
	OperationsDailyStatusCompleted  OperationsDailyStatus = "completed"
	OperationsDailyStatusBlocked    OperationsDailyStatus = "blocked"
	OperationsDailyStatusException  OperationsDailyStatus = "exception"
)

const (
	OperationsDailySeverityNormal  OperationsDailySeverity = "normal"
	OperationsDailySeverityWarning OperationsDailySeverity = "warning"
	OperationsDailySeverityDanger  OperationsDailySeverity = "danger"
)

var ErrInvalidOperationsDailyReport = errors.New("operations daily report is invalid")

type OperationsDailySignal struct {
	ID            string
	Area          OperationsDailyArea
	SourceType    string
	SourceID      string
	RefNo         string
	Title         string
	WarehouseID   string
	WarehouseCode string
	BusinessDate  time.Time
	Status        OperationsDailyStatus
	Severity      OperationsDailySeverity
	Quantity      string
	UOMCode       string
	ExceptionCode string
	Owner         string
}

type OperationsDailyOptions struct {
	GeneratedAt time.Time
}

type OperationsDailyReport struct {
	Metadata ReportMetadata
	Summary  OperationsDailySummary
	Areas    []OperationsDailyAreaSummary
	Rows     []OperationsDailyRow
}

type OperationsDailySummary struct {
	SignalCount     int
	PendingCount    int
	InProgressCount int
	CompletedCount  int
	BlockedCount    int
	ExceptionCount  int
}

type OperationsDailyAreaSummary struct {
	Area            string
	SignalCount     int
	PendingCount    int
	InProgressCount int
	CompletedCount  int
	BlockedCount    int
	ExceptionCount  int
}

type OperationsDailyRow struct {
	ID            string
	Area          string
	SourceType    string
	SourceID      string
	RefNo         string
	Title         string
	WarehouseID   string
	WarehouseCode string
	BusinessDate  string
	Status        string
	Severity      string
	Quantity      string
	UOMCode       string
	ExceptionCode string
	Owner         string
}

func NewOperationsDailyReport(
	filters ReportFilters,
	signals []OperationsDailySignal,
	options OperationsDailyOptions,
) (OperationsDailyReport, error) {
	generatedAt := options.GeneratedAt
	if generatedAt.IsZero() {
		generatedAt = time.Now().UTC()
	}

	rows := make([]OperationsDailyRow, 0, len(signals))
	summary := OperationsDailySummary{}
	areasByKey := make(map[OperationsDailyArea]*OperationsDailyAreaSummary)

	for _, signal := range signals {
		normalized, err := normalizeOperationsDailySignal(signal)
		if err != nil {
			return OperationsDailyReport{}, err
		}
		if !matchesOperationsDailyFilters(filters, normalized) {
			continue
		}

		row := newOperationsDailyRow(normalized)
		rows = append(rows, row)
		addOperationsDailyCounts(&summary, normalized.Status)

		areaSummary := areasByKey[normalized.Area]
		if areaSummary == nil {
			areaSummary = &OperationsDailyAreaSummary{Area: string(normalized.Area)}
			areasByKey[normalized.Area] = areaSummary
		}
		addOperationsDailyAreaCounts(areaSummary, normalized.Status)
	}

	sort.Slice(rows, func(i int, j int) bool {
		left := rows[i]
		right := rows[j]
		if left.BusinessDate != right.BusinessDate {
			return left.BusinessDate < right.BusinessDate
		}
		if operationsDailyAreaOrder(left.Area) != operationsDailyAreaOrder(right.Area) {
			return operationsDailyAreaOrder(left.Area) < operationsDailyAreaOrder(right.Area)
		}
		if operationsDailyStatusOrder(left.Status) != operationsDailyStatusOrder(right.Status) {
			return operationsDailyStatusOrder(left.Status) < operationsDailyStatusOrder(right.Status)
		}

		return left.RefNo < right.RefNo
	})

	areas := make([]OperationsDailyAreaSummary, 0, len(areasByKey))
	for _, area := range areasByKey {
		areas = append(areas, *area)
	}
	sort.Slice(areas, func(i int, j int) bool {
		return operationsDailyAreaOrder(areas[i].Area) < operationsDailyAreaOrder(areas[j].Area)
	})

	return OperationsDailyReport{
		Metadata: NewReportMetadata(filters, generatedAt),
		Summary:  summary,
		Areas:    areas,
		Rows:     rows,
	}, nil
}

func normalizeOperationsDailySignal(signal OperationsDailySignal) (OperationsDailySignal, error) {
	signal.ID = strings.TrimSpace(signal.ID)
	signal.SourceType = strings.TrimSpace(signal.SourceType)
	signal.SourceID = strings.TrimSpace(signal.SourceID)
	signal.RefNo = strings.TrimSpace(signal.RefNo)
	signal.Title = strings.TrimSpace(signal.Title)
	signal.WarehouseID = strings.TrimSpace(signal.WarehouseID)
	signal.WarehouseCode = strings.TrimSpace(signal.WarehouseCode)
	signal.Area = normalizeOperationsDailyArea(signal.Area)
	signal.Status = normalizeOperationsDailyStatus(signal.Status)
	signal.Severity = normalizeOperationsDailySeverity(signal.Severity)
	signal.Quantity = strings.TrimSpace(signal.Quantity)
	signal.UOMCode = strings.TrimSpace(signal.UOMCode)
	signal.ExceptionCode = strings.TrimSpace(signal.ExceptionCode)
	signal.Owner = strings.TrimSpace(signal.Owner)

	if signal.ID == "" ||
		signal.Area == "" ||
		signal.SourceType == "" ||
		signal.SourceID == "" ||
		signal.RefNo == "" ||
		signal.Title == "" ||
		signal.WarehouseID == "" ||
		signal.BusinessDate.IsZero() ||
		signal.Status == "" ||
		signal.Severity == "" {
		return OperationsDailySignal{}, ErrInvalidOperationsDailyReport
	}
	if signal.Quantity != "" {
		if _, err := decimal.ParseQuantity(signal.Quantity); err != nil {
			return OperationsDailySignal{}, ErrInvalidOperationsDailyReport
		}
	}
	if signal.UOMCode != "" {
		code, err := decimal.NormalizeUOMCode(signal.UOMCode)
		if err != nil {
			return OperationsDailySignal{}, ErrInvalidOperationsDailyReport
		}
		signal.UOMCode = code.String()
	}

	return signal, nil
}

func matchesOperationsDailyFilters(filters ReportFilters, signal OperationsDailySignal) bool {
	if !filters.IncludesBusinessDate(signal.BusinessDate) {
		return false
	}
	if filters.WarehouseID != "" && signal.WarehouseID != filters.WarehouseID {
		return false
	}
	if filters.Status != "" && string(signal.Status) != strings.TrimSpace(filters.Status) {
		return false
	}

	return true
}

func newOperationsDailyRow(signal OperationsDailySignal) OperationsDailyRow {
	return OperationsDailyRow{
		ID:            signal.ID,
		Area:          string(signal.Area),
		SourceType:    signal.SourceType,
		SourceID:      signal.SourceID,
		RefNo:         signal.RefNo,
		Title:         signal.Title,
		WarehouseID:   signal.WarehouseID,
		WarehouseCode: signal.WarehouseCode,
		BusinessDate:  formatReportDate(signal.BusinessDate),
		Status:        string(signal.Status),
		Severity:      string(signal.Severity),
		Quantity:      signal.Quantity,
		UOMCode:       signal.UOMCode,
		ExceptionCode: signal.ExceptionCode,
		Owner:         signal.Owner,
	}
}

func addOperationsDailyCounts(summary *OperationsDailySummary, status OperationsDailyStatus) {
	summary.SignalCount++
	switch status {
	case OperationsDailyStatusPending:
		summary.PendingCount++
	case OperationsDailyStatusInProgress:
		summary.InProgressCount++
	case OperationsDailyStatusCompleted:
		summary.CompletedCount++
	case OperationsDailyStatusBlocked:
		summary.BlockedCount++
	case OperationsDailyStatusException:
		summary.ExceptionCount++
	}
}

func addOperationsDailyAreaCounts(summary *OperationsDailyAreaSummary, status OperationsDailyStatus) {
	summary.SignalCount++
	switch status {
	case OperationsDailyStatusPending:
		summary.PendingCount++
	case OperationsDailyStatusInProgress:
		summary.InProgressCount++
	case OperationsDailyStatusCompleted:
		summary.CompletedCount++
	case OperationsDailyStatusBlocked:
		summary.BlockedCount++
	case OperationsDailyStatusException:
		summary.ExceptionCount++
	}
}

func normalizeOperationsDailyArea(area OperationsDailyArea) OperationsDailyArea {
	switch OperationsDailyArea(strings.ToLower(strings.TrimSpace(string(area)))) {
	case OperationsDailyAreaInbound:
		return OperationsDailyAreaInbound
	case OperationsDailyAreaQC:
		return OperationsDailyAreaQC
	case OperationsDailyAreaOutbound:
		return OperationsDailyAreaOutbound
	case OperationsDailyAreaReturns:
		return OperationsDailyAreaReturns
	case OperationsDailyAreaStock:
		return OperationsDailyAreaStock
	case OperationsDailyAreaSubcontract:
		return OperationsDailyAreaSubcontract
	default:
		return ""
	}
}

func normalizeOperationsDailyStatus(status OperationsDailyStatus) OperationsDailyStatus {
	switch OperationsDailyStatus(strings.ToLower(strings.TrimSpace(string(status)))) {
	case OperationsDailyStatusPending:
		return OperationsDailyStatusPending
	case OperationsDailyStatusInProgress:
		return OperationsDailyStatusInProgress
	case OperationsDailyStatusCompleted:
		return OperationsDailyStatusCompleted
	case OperationsDailyStatusBlocked:
		return OperationsDailyStatusBlocked
	case OperationsDailyStatusException:
		return OperationsDailyStatusException
	default:
		return ""
	}
}

func normalizeOperationsDailySeverity(severity OperationsDailySeverity) OperationsDailySeverity {
	switch OperationsDailySeverity(strings.ToLower(strings.TrimSpace(string(severity)))) {
	case "":
		return OperationsDailySeverityNormal
	case OperationsDailySeverityNormal:
		return OperationsDailySeverityNormal
	case OperationsDailySeverityWarning:
		return OperationsDailySeverityWarning
	case OperationsDailySeverityDanger:
		return OperationsDailySeverityDanger
	default:
		return ""
	}
}

func operationsDailyAreaOrder(area string) int {
	switch area {
	case string(OperationsDailyAreaInbound):
		return 1
	case string(OperationsDailyAreaQC):
		return 2
	case string(OperationsDailyAreaOutbound):
		return 3
	case string(OperationsDailyAreaReturns):
		return 4
	case string(OperationsDailyAreaStock):
		return 5
	case string(OperationsDailyAreaSubcontract):
		return 6
	default:
		return 99
	}
}

func operationsDailyStatusOrder(status string) int {
	switch status {
	case string(OperationsDailyStatusBlocked):
		return 1
	case string(OperationsDailyStatusException):
		return 2
	case string(OperationsDailyStatusPending):
		return 3
	case string(OperationsDailyStatusInProgress):
		return 4
	case string(OperationsDailyStatusCompleted):
		return 5
	default:
		return 99
	}
}
