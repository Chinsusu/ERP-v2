package domain

import (
	"errors"
	"math/big"
	"net/url"
	"sort"
	"strings"
	"time"

	inventorydomain "github.com/Chinsusu/ERP-v2/apps/api/internal/modules/inventory/domain"
	"github.com/Chinsusu/ERP-v2/apps/api/internal/shared/decimal"
)

const defaultLowStockThreshold = "10"
const defaultExpiryWarningDays = 30

var ErrInvalidInventorySnapshotReport = errors.New("inventory snapshot report is invalid")

type InventorySnapshotOptions struct {
	LowStockThreshold string
	ExpiryWarningDays int
	GeneratedAt       time.Time
}

type InventorySnapshotReport struct {
	Metadata ReportMetadata
	Summary  InventorySnapshotSummary
	Rows     []InventorySnapshotRow
}

type InventorySnapshotSummary struct {
	RowCount          int
	LowStockRowCount  int
	ExpiryWarningRows int
	ExpiredRows       int
	TotalsByUOM       []InventorySnapshotUOMTotal
}

type InventorySnapshotUOMTotal struct {
	BaseUOMCode   string
	PhysicalQty   string
	ReservedQty   string
	QuarantineQty string
	BlockedQty    string
	AvailableQty  string
}

type InventorySnapshotRow struct {
	WarehouseID      string
	WarehouseCode    string
	LocationID       string
	LocationCode     string
	ItemID           string
	SKU              string
	BatchID          string
	BatchNo          string
	BatchExpiry      string
	BaseUOMCode      string
	PhysicalQty      string
	ReservedQty      string
	QuarantineQty    string
	BlockedQty       string
	AvailableQty     string
	LowStock         bool
	ExpiryWarning    bool
	Expired          bool
	BatchQCStatus    string
	BatchStatus      string
	SourceStockState string
	SourceReferences []ReportSourceReference
}

func NewInventorySnapshotReport(
	filters ReportFilters,
	snapshots []inventorydomain.AvailableStockSnapshot,
	options InventorySnapshotOptions,
) (InventorySnapshotReport, error) {
	threshold, err := normalizeLowStockThreshold(options.LowStockThreshold)
	if err != nil {
		return InventorySnapshotReport{}, err
	}
	warningDays := options.ExpiryWarningDays
	if warningDays <= 0 {
		warningDays = defaultExpiryWarningDays
	}
	generatedAt := options.GeneratedAt
	if generatedAt.IsZero() {
		generatedAt = time.Now().UTC()
	}
	asOf := filters.BusinessDate
	if asOf.IsZero() {
		asOf = generatedAt
	}
	asOf = dateOnly(asOf)

	rows := make([]InventorySnapshotRow, 0, len(snapshots))
	totalByUOM := make(map[string]*inventorySnapshotQuantityTotal)
	summary := InventorySnapshotSummary{}

	for _, snapshot := range snapshots {
		if strings.TrimSpace(snapshot.WarehouseID) == "" ||
			strings.TrimSpace(snapshot.SKU) == "" ||
			snapshot.BaseUOMCode == "" {
			return InventorySnapshotReport{}, ErrInvalidInventorySnapshotReport
		}

		row := inventorySnapshotRow(snapshot, threshold, asOf, warningDays)
		rows = append(rows, row)
		summary.RowCount++
		if row.LowStock {
			summary.LowStockRowCount++
		}
		if row.ExpiryWarning {
			summary.ExpiryWarningRows++
		}
		if row.Expired {
			summary.ExpiredRows++
		}

		uom := row.BaseUOMCode
		total := totalByUOM[uom]
		if total == nil {
			total = newInventorySnapshotQuantityTotal(uom)
			totalByUOM[uom] = total
		}
		total.add(snapshot)
	}

	summary.TotalsByUOM = inventorySnapshotTotals(totalByUOM)

	return InventorySnapshotReport{
		Metadata: NewReportMetadata(filters, generatedAt),
		Summary:  summary,
		Rows:     rows,
	}, nil
}

func inventorySnapshotRow(
	snapshot inventorydomain.AvailableStockSnapshot,
	threshold decimal.Decimal,
	asOf time.Time,
	warningDays int,
) InventorySnapshotRow {
	expired := isExpired(snapshot.BatchExpiry, asOf)
	expiryWarning := isExpiryWarning(snapshot.BatchExpiry, asOf, warningDays)
	lowStock := compareQuantity(snapshot.AvailableQty, threshold) <= 0
	stockState := sourceStockState(snapshot)

	return InventorySnapshotRow{
		WarehouseID:      strings.TrimSpace(snapshot.WarehouseID),
		WarehouseCode:    strings.TrimSpace(snapshot.WarehouseCode),
		LocationID:       strings.TrimSpace(snapshot.LocationID),
		LocationCode:     strings.TrimSpace(snapshot.LocationCode),
		ItemID:           strings.TrimSpace(snapshot.ItemID),
		SKU:              strings.ToUpper(strings.TrimSpace(snapshot.SKU)),
		BatchID:          strings.TrimSpace(snapshot.BatchID),
		BatchNo:          strings.TrimSpace(snapshot.BatchNo),
		BatchExpiry:      formatOptionalReportDate(snapshot.BatchExpiry),
		BaseUOMCode:      snapshot.BaseUOMCode.String(),
		PhysicalQty:      snapshot.PhysicalQty.String(),
		ReservedQty:      snapshot.ReservedQty.String(),
		QuarantineQty:    snapshot.QCHoldQty.String(),
		BlockedQty:       snapshot.BlockedQty.String(),
		AvailableQty:     snapshot.AvailableQty.String(),
		LowStock:         lowStock,
		ExpiryWarning:    expiryWarning,
		Expired:          expired,
		BatchQCStatus:    string(snapshot.BatchQCStatus),
		BatchStatus:      string(snapshot.BatchStatus),
		SourceStockState: stockState,
		SourceReferences: inventorySnapshotSourceReferences(snapshot, stockState, lowStock, expiryWarning, expired),
	}
}

func inventorySnapshotSourceReferences(
	snapshot inventorydomain.AvailableStockSnapshot,
	stockState string,
	lowStock bool,
	expiryWarning bool,
	expired bool,
) []ReportSourceReference {
	refs := make([]ReportSourceReference, 0, 6)
	warehouseID := strings.TrimSpace(snapshot.WarehouseID)
	if warehouseID != "" {
		refs = appendInventorySnapshotReference(refs, ReportSourceReferenceInput{
			EntityType: "warehouse",
			ID:         warehouseID,
			Label:      firstNonEmpty(snapshot.WarehouseCode, warehouseID),
			Href:       inventorySnapshotEntityHref("master-data", "warehouse", warehouseID),
		})
	}

	itemID := strings.TrimSpace(snapshot.ItemID)
	if itemID != "" {
		refs = appendInventorySnapshotReference(refs, ReportSourceReferenceInput{
			EntityType: "item",
			ID:         itemID,
			Label:      strings.ToUpper(strings.TrimSpace(snapshot.SKU)),
			Href:       inventorySnapshotEntityHref("master-data", "item", itemID),
		})
	}

	batchID := strings.TrimSpace(snapshot.BatchID)
	if batchID != "" {
		refs = appendInventorySnapshotReference(refs, ReportSourceReferenceInput{
			EntityType: "inventory_batch",
			ID:         batchID,
			Label:      firstNonEmpty(snapshot.BatchNo, batchID),
			Href:       inventorySnapshotEntityHref("inventory", "inventory_batch", batchID),
		})
	}

	if warehouseID != "" && strings.TrimSpace(snapshot.SKU) != "" {
		refs = appendInventorySnapshotReference(refs, ReportSourceReferenceInput{
			EntityType: "stock_state",
			ID:         inventorySnapshotStockContextID(snapshot, stockState),
			Label:      stockState,
			Href:       inventorySnapshotStockContextHref(snapshot, stockState),
		})
	}

	for _, warning := range inventorySnapshotWarnings(lowStock, expiryWarning, expired) {
		refs = appendInventorySnapshotReference(refs, ReportSourceReferenceInput{
			EntityType: "inventory_warning",
			ID:         inventorySnapshotStockContextID(snapshot, stockState) + ":" + warning,
			Label:      warning,
			Href:       inventorySnapshotWarningHref(snapshot, stockState, warning),
		})
	}

	return refs
}

func appendInventorySnapshotReference(
	refs []ReportSourceReference,
	input ReportSourceReferenceInput,
) []ReportSourceReference {
	reference, err := NewReportSourceReference(input)
	if err != nil {
		panic(ErrInvalidInventorySnapshotReport)
	}

	return append(refs, reference)
}

func inventorySnapshotEntityHref(module string, entityType string, id string) string {
	params := url.Values{}
	params.Set("source_type", entityType)
	params.Set("source_id", id)

	return "/" + module + "?" + params.Encode()
}

func inventorySnapshotStockContextHref(snapshot inventorydomain.AvailableStockSnapshot, stockState string) string {
	params := inventorySnapshotStockContextParams(snapshot, stockState)

	return "/inventory?" + params.Encode()
}

func inventorySnapshotWarningHref(
	snapshot inventorydomain.AvailableStockSnapshot,
	stockState string,
	warning string,
) string {
	params := inventorySnapshotStockContextParams(snapshot, stockState)
	params.Set("warning", warning)

	return "/reports?report=inventory-snapshot&" + params.Encode()
}

func inventorySnapshotStockContextParams(
	snapshot inventorydomain.AvailableStockSnapshot,
	stockState string,
) url.Values {
	params := url.Values{}
	params.Set("warehouse_id", strings.TrimSpace(snapshot.WarehouseID))
	params.Set("sku", strings.ToUpper(strings.TrimSpace(snapshot.SKU)))
	if strings.TrimSpace(snapshot.ItemID) != "" {
		params.Set("item_id", strings.TrimSpace(snapshot.ItemID))
	}
	if strings.TrimSpace(snapshot.BatchID) != "" {
		params.Set("batch_id", strings.TrimSpace(snapshot.BatchID))
	}
	if stockState != "" {
		params.Set("status", stockState)
	}

	return params
}

func inventorySnapshotStockContextID(snapshot inventorydomain.AvailableStockSnapshot, stockState string) string {
	parts := []string{
		strings.TrimSpace(snapshot.WarehouseID),
		strings.TrimSpace(snapshot.LocationID),
		strings.ToUpper(strings.TrimSpace(snapshot.SKU)),
		strings.TrimSpace(snapshot.BatchID),
		stockState,
	}
	for index, part := range parts {
		if part == "" {
			parts[index] = "-"
		}
	}

	return strings.Join(parts, ":")
}

func inventorySnapshotWarnings(lowStock bool, expiryWarning bool, expired bool) []string {
	warnings := make([]string, 0, 3)
	if lowStock {
		warnings = append(warnings, "low_stock")
	}
	if expiryWarning {
		warnings = append(warnings, "expiry_warning")
	}
	if expired {
		warnings = append(warnings, "expired")
	}

	return warnings
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		normalized := strings.TrimSpace(value)
		if normalized != "" {
			return normalized
		}
	}

	return ""
}

type inventorySnapshotQuantityTotal struct {
	uom           string
	physicalQty   decimal.Decimal
	reservedQty   decimal.Decimal
	quarantineQty decimal.Decimal
	blockedQty    decimal.Decimal
	availableQty  decimal.Decimal
}

func newInventorySnapshotQuantityTotal(uom string) *inventorySnapshotQuantityTotal {
	zero := decimal.MustQuantity("0")

	return &inventorySnapshotQuantityTotal{
		uom:           uom,
		physicalQty:   zero,
		reservedQty:   zero,
		quarantineQty: zero,
		blockedQty:    zero,
		availableQty:  zero,
	}
}

func (t *inventorySnapshotQuantityTotal) add(snapshot inventorydomain.AvailableStockSnapshot) {
	t.physicalQty = mustReportQuantityAdd(t.physicalQty, snapshot.PhysicalQty)
	t.reservedQty = mustReportQuantityAdd(t.reservedQty, snapshot.ReservedQty)
	t.quarantineQty = mustReportQuantityAdd(t.quarantineQty, snapshot.QCHoldQty)
	t.blockedQty = mustReportQuantityAdd(t.blockedQty, snapshot.BlockedQty)
	t.availableQty = mustReportQuantityAdd(t.availableQty, snapshot.AvailableQty)
}

func inventorySnapshotTotals(totals map[string]*inventorySnapshotQuantityTotal) []InventorySnapshotUOMTotal {
	result := make([]InventorySnapshotUOMTotal, 0, len(totals))
	for _, total := range totals {
		result = append(result, InventorySnapshotUOMTotal{
			BaseUOMCode:   total.uom,
			PhysicalQty:   total.physicalQty.String(),
			ReservedQty:   total.reservedQty.String(),
			QuarantineQty: total.quarantineQty.String(),
			BlockedQty:    total.blockedQty.String(),
			AvailableQty:  total.availableQty.String(),
		})
	}
	sort.Slice(result, func(i int, j int) bool {
		return result[i].BaseUOMCode < result[j].BaseUOMCode
	})

	return result
}

func normalizeLowStockThreshold(value string) (decimal.Decimal, error) {
	raw := strings.TrimSpace(value)
	if raw == "" {
		raw = defaultLowStockThreshold
	}
	threshold, err := decimal.ParseQuantity(raw)
	if err != nil || threshold.IsNegative() {
		return "", ErrInvalidInventorySnapshotReport
	}

	return threshold, nil
}

func sourceStockState(snapshot inventorydomain.AvailableStockSnapshot) string {
	if !snapshot.QCHoldQty.IsZero() {
		return "quarantine"
	}
	if !snapshot.BlockedQty.IsZero() {
		return "blocked"
	}
	if !snapshot.ReservedQty.IsZero() {
		return "reserved"
	}

	return "available"
}

func isExpired(expiry time.Time, asOf time.Time) bool {
	if expiry.IsZero() {
		return false
	}

	return dateOnly(expiry).Before(dateOnly(asOf))
}

func isExpiryWarning(expiry time.Time, asOf time.Time, warningDays int) bool {
	if expiry.IsZero() || isExpired(expiry, asOf) {
		return false
	}
	warningUntil := dateOnly(asOf).AddDate(0, 0, warningDays)

	return !dateOnly(expiry).After(warningUntil)
}

func formatOptionalReportDate(value time.Time) string {
	if value.IsZero() {
		return ""
	}

	return formatReportDate(value)
}

func mustReportQuantityAdd(left decimal.Decimal, right decimal.Decimal) decimal.Decimal {
	result, err := decimal.AddQuantity(left, right)
	if err != nil {
		panic(err)
	}

	return result
}

func compareQuantity(left decimal.Decimal, right decimal.Decimal) int {
	leftValue, ok := new(big.Rat).SetString(left.String())
	if !ok {
		panic(ErrInvalidInventorySnapshotReport)
	}
	rightValue, ok := new(big.Rat).SetString(right.String())
	if !ok {
		panic(ErrInvalidInventorySnapshotReport)
	}

	return leftValue.Cmp(rightValue)
}
