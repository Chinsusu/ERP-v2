package domain

import (
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"
)

type ReturnReceiptStatus string

const ReturnStatusPendingInspection ReturnReceiptStatus = "pending_inspection"

type ReturnSource string

const ReturnSourceShipper ReturnSource = "SHIPPER"
const ReturnSourceCarrier ReturnSource = "CARRIER"
const ReturnSourceCustomer ReturnSource = "CUSTOMER"
const ReturnSourceMarketplace ReturnSource = "MARKETPLACE"
const ReturnSourceUnknown ReturnSource = "UNKNOWN"

type ReturnDisposition string

const ReturnDispositionReusable ReturnDisposition = "reusable"
const ReturnDispositionNotReusable ReturnDisposition = "not_reusable"
const ReturnDispositionNeedsInspection ReturnDisposition = "needs_inspection"

const ReturnReceiptMovementType = "RETURN_RECEIPT"

var ErrReturnReceiptRequiredField = errors.New("return receipt required field is missing")
var ErrReturnReceiptScanCodeRequired = errors.New("return receipt scan code is required")
var ErrReturnReceiptInvalidDisposition = errors.New("return receipt disposition is invalid")

type ExpectedReturn struct {
	OrderNo       string
	OrderStatus   string
	TrackingNo    string
	ReturnCode    string
	ShipmentID    string
	CustomerName  string
	SKU           string
	ProductName   string
	Quantity      int
	WarehouseID   string
	WarehouseCode string
	Source        ReturnSource
}

type ReturnReceipt struct {
	ID                string
	ReceiptNo         string
	WarehouseID       string
	WarehouseCode     string
	Source            ReturnSource
	ReceivedBy        string
	ReceivedAt        time.Time
	PackageCondition  string
	Status            ReturnReceiptStatus
	Disposition       ReturnDisposition
	TargetLocation    string
	OriginalOrderNo   string
	TrackingNo        string
	ReturnCode        string
	ScanCode          string
	CustomerName      string
	UnknownCase       bool
	Lines             []ReturnReceiptLine
	StockMovement     *ReturnStockMovement
	InvestigationNote string
	CreatedAt         time.Time
}

type ReturnReceiptLine struct {
	ID          string
	SKU         string
	ProductName string
	Quantity    int
	Condition   string
}

type ReturnStockMovement struct {
	ID                string
	MovementType      string
	SKU               string
	WarehouseID       string
	Quantity          int
	TargetStockStatus string
	SourceDocID       string
}

type ReturnReceiptFilter struct {
	WarehouseID string
	Status      ReturnReceiptStatus
}

type NewReturnReceiptInput struct {
	ID                string
	ReceiptNo         string
	WarehouseID       string
	WarehouseCode     string
	Source            ReturnSource
	ReceivedBy        string
	ScanCode          string
	PackageCondition  string
	Disposition       ReturnDisposition
	ExpectedReturn    *ExpectedReturn
	InvestigationNote string
	CreatedAt         time.Time
}

func NewReturnReceipt(input NewReturnReceiptInput) (ReturnReceipt, error) {
	scanCode := NormalizeReturnScanCode(input.ScanCode)
	if scanCode == "" {
		return ReturnReceipt{}, ErrReturnReceiptScanCodeRequired
	}

	disposition := NormalizeReturnDisposition(input.Disposition)
	if disposition == "" {
		return ReturnReceipt{}, ErrReturnReceiptInvalidDisposition
	}

	warehouseID := strings.TrimSpace(input.WarehouseID)
	if warehouseID == "" && input.ExpectedReturn != nil {
		warehouseID = strings.TrimSpace(input.ExpectedReturn.WarehouseID)
	}
	if warehouseID == "" {
		return ReturnReceipt{}, ErrReturnReceiptRequiredField
	}

	createdAt := input.CreatedAt
	if createdAt.IsZero() {
		createdAt = time.Now().UTC()
	}

	receipt := ReturnReceipt{
		ID:                strings.TrimSpace(input.ID),
		ReceiptNo:         strings.TrimSpace(input.ReceiptNo),
		WarehouseID:       warehouseID,
		WarehouseCode:     strings.ToUpper(strings.TrimSpace(input.WarehouseCode)),
		Source:            NormalizeReturnSource(input.Source),
		ReceivedBy:        strings.TrimSpace(input.ReceivedBy),
		ReceivedAt:        createdAt.UTC(),
		PackageCondition:  strings.TrimSpace(input.PackageCondition),
		Status:            ReturnStatusPendingInspection,
		Disposition:       disposition,
		ScanCode:          scanCode,
		InvestigationNote: strings.TrimSpace(input.InvestigationNote),
		CreatedAt:         createdAt.UTC(),
	}
	if receipt.WarehouseCode == "" {
		receipt.WarehouseCode = strings.ToUpper(receipt.WarehouseID)
	}
	if receipt.Source == "" {
		receipt.Source = ReturnSourceUnknown
	}
	if receipt.ReceivedBy == "" {
		receipt.ReceivedBy = "return-receiver"
	}
	if receipt.PackageCondition == "" {
		receipt.PackageCondition = "pending inspection"
	}

	if input.ExpectedReturn != nil {
		applyExpectedReturn(&receipt, *input.ExpectedReturn)
	} else {
		applyUnknownReturn(&receipt)
	}

	if receipt.ReceiptNo == "" {
		receipt.ReceiptNo = defaultReturnReceiptNo(receipt)
	}
	if receipt.ID == "" {
		receipt.ID = strings.ToLower(receipt.ReceiptNo)
	}

	receipt.TargetLocation = targetLocationForDisposition(disposition)
	if disposition == ReturnDispositionReusable {
		receipt.StockMovement = newReturnReceiptMovement(receipt)
	}

	return receipt, nil
}

func NewReturnReceiptFilter(warehouseID string, status ReturnReceiptStatus) ReturnReceiptFilter {
	return ReturnReceiptFilter{
		WarehouseID: strings.TrimSpace(warehouseID),
		Status:      NormalizeReturnReceiptStatus(status),
	}
}

func NormalizeReturnReceiptStatus(status ReturnReceiptStatus) ReturnReceiptStatus {
	switch status {
	case ReturnStatusPendingInspection:
		return status
	default:
		return ""
	}
}

func NormalizeReturnSource(source ReturnSource) ReturnSource {
	switch ReturnSource(strings.ToUpper(strings.TrimSpace(string(source)))) {
	case ReturnSourceShipper:
		return ReturnSourceShipper
	case ReturnSourceCarrier:
		return ReturnSourceCarrier
	case ReturnSourceCustomer:
		return ReturnSourceCustomer
	case ReturnSourceMarketplace:
		return ReturnSourceMarketplace
	case ReturnSourceUnknown:
		return ReturnSourceUnknown
	default:
		return ReturnSourceUnknown
	}
}

func NormalizeReturnDisposition(disposition ReturnDisposition) ReturnDisposition {
	switch ReturnDisposition(strings.ToLower(strings.TrimSpace(string(disposition)))) {
	case ReturnDispositionReusable:
		return ReturnDispositionReusable
	case ReturnDispositionNotReusable:
		return ReturnDispositionNotReusable
	case ReturnDispositionNeedsInspection:
		return ReturnDispositionNeedsInspection
	default:
		return ""
	}
}

func NormalizeReturnScanCode(code string) string {
	return strings.ToUpper(strings.TrimSpace(code))
}

func IsReturnReceivableOrderStatus(status string) bool {
	switch strings.ToLower(strings.TrimSpace(status)) {
	case "handed_over", "delivered":
		return true
	default:
		return false
	}
}

func (expected ExpectedReturn) MatchesScanCode(code string) bool {
	normalizedCode := NormalizeReturnScanCode(code)
	if normalizedCode == "" {
		return false
	}

	for _, candidate := range []string{expected.OrderNo, expected.TrackingNo, expected.ReturnCode, expected.ShipmentID} {
		if NormalizeReturnScanCode(candidate) == normalizedCode {
			return true
		}
	}

	return false
}

func (r ReturnReceipt) Clone() ReturnReceipt {
	clone := r
	clone.Lines = append([]ReturnReceiptLine(nil), r.Lines...)
	if r.StockMovement != nil {
		movement := *r.StockMovement
		clone.StockMovement = &movement
	}

	return clone
}

func SortReturnReceipts(rows []ReturnReceipt) {
	sort.SliceStable(rows, func(i int, j int) bool {
		left := rows[i]
		right := rows[j]
		if !left.CreatedAt.Equal(right.CreatedAt) {
			return left.CreatedAt.After(right.CreatedAt)
		}

		return left.ReceiptNo > right.ReceiptNo
	})
}

func applyExpectedReturn(receipt *ReturnReceipt, expected ExpectedReturn) {
	receipt.OriginalOrderNo = strings.TrimSpace(expected.OrderNo)
	receipt.TrackingNo = strings.TrimSpace(expected.TrackingNo)
	receipt.ReturnCode = strings.TrimSpace(expected.ReturnCode)
	receipt.CustomerName = strings.TrimSpace(expected.CustomerName)
	receipt.UnknownCase = false
	if receipt.WarehouseCode == "" {
		receipt.WarehouseCode = strings.ToUpper(strings.TrimSpace(expected.WarehouseCode))
	}
	if receipt.Source == ReturnSourceUnknown && expected.Source != "" {
		receipt.Source = NormalizeReturnSource(expected.Source)
	}
	if receipt.CustomerName == "" {
		receipt.CustomerName = "Known customer"
	}

	quantity := expected.Quantity
	if quantity <= 0 {
		quantity = 1
	}
	receipt.Lines = []ReturnReceiptLine{
		{
			ID:          fmt.Sprintf("line-%s", strings.ToLower(strings.TrimSpace(expected.SKU))),
			SKU:         strings.ToUpper(strings.TrimSpace(expected.SKU)),
			ProductName: strings.TrimSpace(expected.ProductName),
			Quantity:    quantity,
			Condition:   receipt.PackageCondition,
		},
	}
	if receipt.Lines[0].SKU == "" {
		receipt.Lines[0].SKU = "UNKNOWN-SKU"
	}
	if receipt.Lines[0].ProductName == "" {
		receipt.Lines[0].ProductName = receipt.Lines[0].SKU
	}
}

func applyUnknownReturn(receipt *ReturnReceipt) {
	receipt.TrackingNo = receipt.ScanCode
	receipt.CustomerName = "Unknown customer"
	receipt.UnknownCase = true
	if receipt.InvestigationNote == "" {
		receipt.InvestigationNote = "Unknown return case created from receiving scan"
	}
	receipt.Lines = []ReturnReceiptLine{
		{
			ID:          "line-unknown-return",
			SKU:         "UNKNOWN-SKU",
			ProductName: "Unknown return item",
			Quantity:    1,
			Condition:   receipt.PackageCondition,
		},
	}
}

func defaultReturnReceiptNo(receipt ReturnReceipt) string {
	if receipt.OriginalOrderNo != "" {
		return fmt.Sprintf("RR-%s", strings.ReplaceAll(receipt.OriginalOrderNo, "SO-", ""))
	}

	code := strings.NewReplacer(" ", "-", "/", "-", "_", "-").Replace(strings.ToLower(receipt.ScanCode))
	return fmt.Sprintf("RR-UNKNOWN-%s", strings.ToUpper(code))
}

func targetLocationForDisposition(disposition ReturnDisposition) string {
	switch disposition {
	case ReturnDispositionReusable:
		return "return-area-pending-inspection"
	case ReturnDispositionNotReusable:
		return "lab-damaged-placeholder"
	case ReturnDispositionNeedsInspection:
		return "return-inspection-queue"
	default:
		return "return-inspection-queue"
	}
}

func newReturnReceiptMovement(receipt ReturnReceipt) *ReturnStockMovement {
	line := ReturnReceiptLine{SKU: "UNKNOWN-SKU", Quantity: 1}
	if len(receipt.Lines) > 0 {
		line = receipt.Lines[0]
	}
	if line.Quantity <= 0 {
		line.Quantity = 1
	}

	return &ReturnStockMovement{
		ID:                fmt.Sprintf("mov-%s", strings.ToLower(receipt.ReceiptNo)),
		MovementType:      ReturnReceiptMovementType,
		SKU:               line.SKU,
		WarehouseID:       receipt.WarehouseID,
		Quantity:          line.Quantity,
		TargetStockStatus: "return_pending",
		SourceDocID:       receipt.ID,
	}
}
