package domain

import (
	"errors"
	"testing"
	"time"

	"github.com/Chinsusu/ERP-v2/apps/api/internal/shared/decimal"
)

func TestNewReportFiltersDefaultsToBusinessDate(t *testing.T) {
	filters, err := NewReportFilters(ReportFilterInput{
		BusinessDate: "2026-04-30",
		WarehouseID:  " wh-hcm ",
		Status:       " open ",
		ItemID:       " sku-001 ",
		Category:     " finished_goods ",
	})
	if err != nil {
		t.Fatalf("NewReportFilters returned error: %v", err)
	}

	if filters.FromDateString() != "2026-04-30" ||
		filters.ToDateString() != "2026-04-30" ||
		filters.BusinessDateString() != "2026-04-30" {
		t.Fatalf("filters dates = %s/%s/%s, want business date defaults", filters.FromDateString(), filters.ToDateString(), filters.BusinessDateString())
	}
	if filters.WarehouseID != "wh-hcm" ||
		filters.Status != "open" ||
		filters.ItemID != "sku-001" ||
		filters.Category != "finished_goods" {
		t.Fatalf("filters = %+v, want trimmed optional filters", filters)
	}
	if filters.Timezone != decimal.TimezoneHoChiMinh {
		t.Fatalf("timezone = %q, want %q", filters.Timezone, decimal.TimezoneHoChiMinh)
	}
}

func TestNewReportFiltersUsesInclusiveDateRange(t *testing.T) {
	filters, err := NewReportFilters(ReportFilterInput{
		FromDate: "2026-04-01",
		ToDate:   "2026-04-30",
		Now:      time.Date(2026, 4, 30, 15, 30, 0, 0, time.UTC),
	})
	if err != nil {
		t.Fatalf("NewReportFilters returned error: %v", err)
	}

	for _, day := range []time.Time{
		time.Date(2026, 4, 1, 0, 0, 0, 0, time.UTC),
		time.Date(2026, 4, 15, 12, 0, 0, 0, time.UTC),
		time.Date(2026, 4, 30, 16, 59, 0, 0, time.UTC),
	} {
		if !filters.IncludesBusinessDate(day) {
			t.Fatalf("expected %s to be included in %+v", day, filters)
		}
	}

	if filters.IncludesBusinessDate(time.Date(2026, 5, 1, 0, 0, 0, 0, time.UTC)) {
		t.Fatal("expected May 1 to be excluded")
	}
}

func TestNewReportFiltersRejectsInvalidDateRange(t *testing.T) {
	_, err := NewReportFilters(ReportFilterInput{
		FromDate: "2026-05-01",
		ToDate:   "2026-04-30",
	})
	if !errors.Is(err, ErrInvalidReportFilter) {
		t.Fatalf("error = %v, want ErrInvalidReportFilter", err)
	}
}

func TestNewReportMetadataUsesUTCGeneratedAt(t *testing.T) {
	filters, err := NewReportFilters(ReportFilterInput{BusinessDate: "2026-04-30"})
	if err != nil {
		t.Fatalf("NewReportFilters returned error: %v", err)
	}
	metadata := NewReportMetadata(filters, time.Date(2026, 4, 30, 12, 0, 0, 0, HoChiMinhLocation()))

	if metadata.Timezone != decimal.TimezoneHoChiMinh {
		t.Fatalf("timezone = %q, want %q", metadata.Timezone, decimal.TimezoneHoChiMinh)
	}
	if metadata.SourceVersion != ReportingSourceVersion {
		t.Fatalf("source version = %q, want %q", metadata.SourceVersion, ReportingSourceVersion)
	}
	if metadata.GeneratedAt.Location() != time.UTC {
		t.Fatalf("generated location = %s, want UTC", metadata.GeneratedAt.Location())
	}
}
