package domain

import (
	"errors"
	"strings"
	"time"

	"github.com/Chinsusu/ERP-v2/apps/api/internal/shared/decimal"
)

const ReportDateLayout = "2006-01-02"
const ReportingSourceVersion = "reporting-v1"

var ErrInvalidReportFilter = errors.New("report filter is invalid")

type ReportFilterInput struct {
	FromDate     string
	ToDate       string
	BusinessDate string
	WarehouseID  string
	Status       string
	ItemID       string
	Category     string
	Now          time.Time
}

type ReportFilters struct {
	FromDate     time.Time
	ToDate       time.Time
	BusinessDate time.Time
	WarehouseID  string
	Status       string
	ItemID       string
	Category     string
	Timezone     string
}

type ReportMetadata struct {
	GeneratedAt   time.Time
	Timezone      string
	Filters       ReportFilters
	SourceVersion string
}

func NewReportFilters(input ReportFilterInput) (ReportFilters, error) {
	location := HoChiMinhLocation()
	now := input.Now
	if now.IsZero() {
		now = time.Now()
	}
	today := dateOnly(now.In(location))

	businessDate, err := parseReportDate(input.BusinessDate, today)
	if err != nil {
		return ReportFilters{}, err
	}

	fromRaw := strings.TrimSpace(input.FromDate)
	toRaw := strings.TrimSpace(input.ToDate)
	if fromRaw == "" && toRaw == "" {
		fromRaw = formatReportDate(businessDate)
		toRaw = fromRaw
	} else if fromRaw == "" {
		fromRaw = toRaw
	} else if toRaw == "" {
		toRaw = fromRaw
	}

	fromDate, err := parseReportDate(fromRaw, today)
	if err != nil {
		return ReportFilters{}, err
	}
	toDate, err := parseReportDate(toRaw, today)
	if err != nil {
		return ReportFilters{}, err
	}
	if toDate.Before(fromDate) {
		return ReportFilters{}, ErrInvalidReportFilter
	}

	return ReportFilters{
		FromDate:     fromDate,
		ToDate:       toDate,
		BusinessDate: businessDate,
		WarehouseID:  strings.TrimSpace(input.WarehouseID),
		Status:       strings.TrimSpace(input.Status),
		ItemID:       strings.TrimSpace(input.ItemID),
		Category:     strings.TrimSpace(input.Category),
		Timezone:     decimal.TimezoneHoChiMinh,
	}, nil
}

func NewReportMetadata(filters ReportFilters, generatedAt time.Time) ReportMetadata {
	if generatedAt.IsZero() {
		generatedAt = time.Now().UTC()
	}

	return ReportMetadata{
		GeneratedAt:   generatedAt.UTC(),
		Timezone:      decimal.TimezoneHoChiMinh,
		Filters:       filters,
		SourceVersion: ReportingSourceVersion,
	}
}

func HoChiMinhLocation() *time.Location {
	return time.FixedZone(decimal.TimezoneHoChiMinh, 7*60*60)
}

func (f ReportFilters) FromDateString() string {
	return formatReportDate(f.FromDate)
}

func (f ReportFilters) ToDateString() string {
	return formatReportDate(f.ToDate)
}

func (f ReportFilters) BusinessDateString() string {
	return formatReportDate(f.BusinessDate)
}

func (f ReportFilters) IncludesBusinessDate(day time.Time) bool {
	candidate := dateOnly(day.In(HoChiMinhLocation()))

	return !candidate.Before(f.FromDate) && !candidate.After(f.ToDate)
}

func parseReportDate(value string, fallback time.Time) (time.Time, error) {
	raw := strings.TrimSpace(value)
	if raw == "" {
		return dateOnly(fallback.In(HoChiMinhLocation())), nil
	}

	parsed, err := time.ParseInLocation(ReportDateLayout, raw, HoChiMinhLocation())
	if err != nil {
		return time.Time{}, ErrInvalidReportFilter
	}

	return dateOnly(parsed), nil
}

func dateOnly(value time.Time) time.Time {
	location := HoChiMinhLocation()
	local := value.In(location)

	return time.Date(local.Year(), local.Month(), local.Day(), 0, 0, 0, 0, location)
}

func formatReportDate(value time.Time) string {
	return dateOnly(value).Format(ReportDateLayout)
}
