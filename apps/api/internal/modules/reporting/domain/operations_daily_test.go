package domain

import (
	"errors"
	"testing"
	"time"
)

func TestNewOperationsDailyReportBuildsSummaryAreasAndRows(t *testing.T) {
	report, err := NewOperationsDailyReport(
		mustReportFiltersWithRange(t, "2026-04-30", "2026-04-30", "2026-04-30"),
		[]OperationsDailySignal{
			{
				ID:            "recv-1",
				Area:          OperationsDailyAreaInbound,
				SourceType:    "goods_receipt",
				SourceID:      "gr-1",
				RefNo:         "GR-001",
				Title:         "Receiving pending",
				WarehouseID:   "wh-hcm",
				WarehouseCode: "HCM",
				BusinessDate:  reportDate(t, "2026-04-30"),
				Status:        OperationsDailyStatusPending,
				Severity:      OperationsDailySeverityWarning,
				Quantity:      "12",
				UOMCode:       "pcs",
				Owner:         "warehouse",
			},
			{
				ID:            "qc-1",
				Area:          OperationsDailyAreaQC,
				SourceType:    "inbound_qc",
				SourceID:      "iqc-1",
				RefNo:         "IQC-001",
				Title:         "QC fail",
				WarehouseID:   "wh-hcm",
				WarehouseCode: "HCM",
				BusinessDate:  reportDate(t, "2026-04-30"),
				Status:        OperationsDailyStatusException,
				Severity:      OperationsDailySeverityDanger,
				ExceptionCode: "QC_FAIL",
				Owner:         "qa",
			},
			{
				ID:            "ship-1",
				Area:          OperationsDailyAreaOutbound,
				SourceType:    "carrier_manifest",
				SourceID:      "manifest-1",
				RefNo:         "MAN-001",
				Title:         "Handover completed",
				WarehouseID:   "wh-hcm",
				WarehouseCode: "HCM",
				BusinessDate:  reportDate(t, "2026-04-30"),
				Status:        OperationsDailyStatusCompleted,
				Severity:      OperationsDailySeverityNormal,
				Owner:         "shipping",
			},
			{
				ID:            "return-1",
				Area:          OperationsDailyAreaReturns,
				SourceType:    "return_receipt",
				SourceID:      "ret-1",
				RefNo:         "RET-001",
				Title:         "Return outside filter",
				WarehouseID:   "wh-hn",
				WarehouseCode: "HN",
				BusinessDate:  reportDate(t, "2026-04-30"),
				Status:        OperationsDailyStatusBlocked,
				Severity:      OperationsDailySeverityDanger,
			},
		},
		OperationsDailyOptions{GeneratedAt: time.Date(2026, 4, 30, 8, 0, 0, 0, time.UTC)},
	)
	if err != nil {
		t.Fatalf("NewOperationsDailyReport returned error: %v", err)
	}

	if report.Metadata.SourceVersion != ReportingSourceVersion {
		t.Fatalf("source version = %q, want %q", report.Metadata.SourceVersion, ReportingSourceVersion)
	}
	if report.Summary.SignalCount != 3 ||
		report.Summary.PendingCount != 1 ||
		report.Summary.CompletedCount != 1 ||
		report.Summary.ExceptionCount != 1 {
		t.Fatalf("summary = %+v, want filtered HCM counts", report.Summary)
	}
	if len(report.Areas) != 3 ||
		report.Areas[0].Area != string(OperationsDailyAreaInbound) ||
		report.Areas[1].Area != string(OperationsDailyAreaQC) ||
		report.Areas[2].Area != string(OperationsDailyAreaOutbound) {
		t.Fatalf("areas = %+v, want inbound/qc/outbound sorted", report.Areas)
	}
	if len(report.Rows) != 3 {
		t.Fatalf("rows length = %d, want 3", len(report.Rows))
	}
	first := report.Rows[0]
	if first.Area != string(OperationsDailyAreaInbound) ||
		first.BusinessDate != "2026-04-30" ||
		first.Quantity != "12" ||
		first.UOMCode != "PCS" {
		t.Fatalf("first row = %+v, want normalized inbound row", first)
	}
	if first.SourceReference.EntityType != "goods_receipt" ||
		first.SourceReference.ID != "gr-1" ||
		first.SourceReference.Label != "GR-001" ||
		first.SourceReference.Href != "/receiving?source_id=gr-1&source_type=goods_receipt" ||
		first.SourceReference.Unavailable {
		t.Fatalf("source reference = %+v, want available goods receipt reference", first.SourceReference)
	}
}

func TestNewOperationsDailyReportFiltersByStatusAndDateRange(t *testing.T) {
	report, err := NewOperationsDailyReport(
		mustReportFiltersWithStatus(t, "2026-04-29", "2026-04-30", "2026-04-30", "blocked"),
		[]OperationsDailySignal{
			operationsDailyTestSignal("return-blocked", OperationsDailyAreaReturns, "2026-04-29", OperationsDailyStatusBlocked),
			operationsDailyTestSignal("stock-blocked", OperationsDailyAreaStock, "2026-04-30", OperationsDailyStatusBlocked),
			operationsDailyTestSignal("subcontract-pending", OperationsDailyAreaSubcontract, "2026-04-30", OperationsDailyStatusPending),
			operationsDailyTestSignal("old-blocked", OperationsDailyAreaOutbound, "2026-04-28", OperationsDailyStatusBlocked),
		},
		OperationsDailyOptions{},
	)
	if err != nil {
		t.Fatalf("NewOperationsDailyReport returned error: %v", err)
	}

	if report.Summary.SignalCount != 2 || report.Summary.BlockedCount != 2 {
		t.Fatalf("summary = %+v, want two blocked rows in date range", report.Summary)
	}
	if report.Rows[0].ID != "return-blocked" || report.Rows[1].ID != "stock-blocked" {
		t.Fatalf("rows = %+v, want date and area sorted blocked rows", report.Rows)
	}
}

func TestNewOperationsDailyReportRejectsInvalidSignals(t *testing.T) {
	tests := []OperationsDailySignal{
		operationsDailyTestSignal("", OperationsDailyAreaInbound, "2026-04-30", OperationsDailyStatusPending),
		operationsDailyTestSignal("bad-area", OperationsDailyArea("unknown"), "2026-04-30", OperationsDailyStatusPending),
		operationsDailyTestSignal("bad-status", OperationsDailyAreaInbound, "2026-04-30", OperationsDailyStatus("unknown")),
		operationsDailyTestSignalWithQuantity("bad-qty", "1.1234567", "PCS"),
		operationsDailyTestSignalWithQuantity("bad-uom", "1", "P C S"),
	}

	for _, signal := range tests {
		_, err := NewOperationsDailyReport(
			mustReportFiltersWithRange(t, "2026-04-30", "2026-04-30", "2026-04-30"),
			[]OperationsDailySignal{signal},
			OperationsDailyOptions{},
		)
		if !errors.Is(err, ErrInvalidOperationsDailyReport) {
			t.Fatalf("signal %+v error = %v, want ErrInvalidOperationsDailyReport", signal, err)
		}
	}
}

func TestNewOperationsDailyReportMarksUnknownSourceReferenceUnavailable(t *testing.T) {
	signal := operationsDailyTestSignal(
		"external-1",
		OperationsDailyAreaInbound,
		"2026-04-30",
		OperationsDailyStatusPending,
	)
	signal.SourceType = "external_note"
	signal.SourceID = "note-1"
	signal.RefNo = "NOTE-001"

	report, err := NewOperationsDailyReport(
		mustReportFiltersWithRange(t, "2026-04-30", "2026-04-30", "2026-04-30"),
		[]OperationsDailySignal{signal},
		OperationsDailyOptions{},
	)
	if err != nil {
		t.Fatalf("NewOperationsDailyReport returned error: %v", err)
	}

	reference := report.Rows[0].SourceReference
	if reference.EntityType != "external_note" ||
		reference.ID != "note-1" ||
		reference.Label != "NOTE-001" ||
		reference.Href != "" ||
		!reference.Unavailable {
		t.Fatalf("source reference = %+v, want unavailable external reference", reference)
	}
}

func operationsDailyTestSignal(
	id string,
	area OperationsDailyArea,
	date string,
	status OperationsDailyStatus,
) OperationsDailySignal {
	return OperationsDailySignal{
		ID:            id,
		Area:          area,
		SourceType:    "test_source",
		SourceID:      "source-" + id,
		RefNo:         "REF-" + id,
		Title:         "Test signal " + id,
		WarehouseID:   "wh-hcm",
		WarehouseCode: "HCM",
		BusinessDate:  mustParseReportDate(date),
		Status:        status,
		Severity:      OperationsDailySeverityNormal,
	}
}

func operationsDailyTestSignalWithQuantity(id string, quantity string, uomCode string) OperationsDailySignal {
	signal := operationsDailyTestSignal(id, OperationsDailyAreaInbound, "2026-04-30", OperationsDailyStatusPending)
	signal.Quantity = quantity
	signal.UOMCode = uomCode

	return signal
}

func mustReportFiltersWithRange(t *testing.T, fromDate string, toDate string, businessDate string) ReportFilters {
	t.Helper()
	filters, err := NewReportFilters(ReportFilterInput{
		FromDate:     fromDate,
		ToDate:       toDate,
		BusinessDate: businessDate,
		WarehouseID:  "wh-hcm",
	})
	if err != nil {
		t.Fatalf("NewReportFilters returned error: %v", err)
	}

	return filters
}

func mustReportFiltersWithStatus(
	t *testing.T,
	fromDate string,
	toDate string,
	businessDate string,
	status string,
) ReportFilters {
	t.Helper()
	filters := mustReportFiltersWithRange(t, fromDate, toDate, businessDate)
	filters.Status = status

	return filters
}

func reportDate(t *testing.T, value string) time.Time {
	t.Helper()
	return mustParseReportDate(value)
}

func mustParseReportDate(value string) time.Time {
	parsed, err := time.ParseInLocation(ReportDateLayout, value, HoChiMinhLocation())
	if err != nil {
		panic(err)
	}

	return parsed
}
