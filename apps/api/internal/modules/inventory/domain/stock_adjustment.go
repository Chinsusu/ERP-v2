package domain

import (
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/Chinsusu/ERP-v2/apps/api/internal/shared/decimal"
)

type StockAdjustmentStatus string

const StockAdjustmentStatusDraft StockAdjustmentStatus = "draft"
const StockAdjustmentStatusSubmitted StockAdjustmentStatus = "submitted"
const StockAdjustmentStatusApproved StockAdjustmentStatus = "approved"
const StockAdjustmentStatusRejected StockAdjustmentStatus = "rejected"
const StockAdjustmentStatusPosted StockAdjustmentStatus = "posted"
const StockAdjustmentStatusCancelled StockAdjustmentStatus = "cancelled"

var ErrStockAdjustmentRequiredField = errors.New("stock adjustment required field is missing")
var ErrStockAdjustmentInvalidQuantity = errors.New("stock adjustment quantity is invalid")
var ErrStockAdjustmentNoVariance = errors.New("stock adjustment requires at least one variance")

type StockAdjustment struct {
	ID            string
	AdjustmentNo  string
	OrgID         string
	WarehouseID   string
	WarehouseCode string
	SourceType    string
	SourceID      string
	Reason        string
	Status        StockAdjustmentStatus
	RequestedBy   string
	Lines         []StockAdjustmentLine
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

type StockAdjustmentLine struct {
	ID           string
	ItemID       string
	SKU          string
	BatchID      string
	BatchNo      string
	LocationID   string
	LocationCode string
	ExpectedQty  decimal.Decimal
	CountedQty   decimal.Decimal
	DeltaQty     decimal.Decimal
	BaseUOMCode  decimal.UOMCode
	Reason       string
}

type NewStockAdjustmentInput struct {
	ID            string
	AdjustmentNo  string
	OrgID         string
	WarehouseID   string
	WarehouseCode string
	SourceType    string
	SourceID      string
	Reason        string
	RequestedBy   string
	Lines         []NewStockAdjustmentLineInput
	CreatedAt     time.Time
}

type NewStockAdjustmentLineInput struct {
	ID           string
	ItemID       string
	SKU          string
	BatchID      string
	BatchNo      string
	LocationID   string
	LocationCode string
	ExpectedQty  decimal.Decimal
	CountedQty   decimal.Decimal
	BaseUOMCode  string
	Reason       string
}

func NewStockAdjustment(input NewStockAdjustmentInput) (StockAdjustment, error) {
	createdAt := input.CreatedAt
	if createdAt.IsZero() {
		createdAt = time.Now().UTC()
	}
	adjustment := StockAdjustment{
		ID:            strings.TrimSpace(input.ID),
		AdjustmentNo:  strings.TrimSpace(input.AdjustmentNo),
		OrgID:         strings.TrimSpace(input.OrgID),
		WarehouseID:   strings.TrimSpace(input.WarehouseID),
		WarehouseCode: strings.ToUpper(strings.TrimSpace(input.WarehouseCode)),
		SourceType:    strings.TrimSpace(input.SourceType),
		SourceID:      strings.TrimSpace(input.SourceID),
		Reason:        strings.TrimSpace(input.Reason),
		Status:        StockAdjustmentStatusDraft,
		RequestedBy:   strings.TrimSpace(input.RequestedBy),
		CreatedAt:     createdAt.UTC(),
		UpdatedAt:     createdAt.UTC(),
	}
	if adjustment.OrgID == "" {
		adjustment.OrgID = "org-my-pham"
	}
	if adjustment.WarehouseCode == "" {
		adjustment.WarehouseCode = strings.ToUpper(adjustment.WarehouseID)
	}
	if adjustment.ID == "" {
		adjustment.ID = fmt.Sprintf("adj_%d", createdAt.UTC().UnixNano())
	}
	if adjustment.AdjustmentNo == "" {
		adjustment.AdjustmentNo = fmt.Sprintf("ADJ-%s-%06d", createdAt.UTC().Format("060102"), createdAt.UTC().UnixNano()%1000000)
	}

	lines := make([]StockAdjustmentLine, 0, len(input.Lines))
	for index, lineInput := range input.Lines {
		line, err := NewStockAdjustmentLine(lineInput)
		if err != nil {
			return StockAdjustment{}, err
		}
		if line.ID == "" {
			line.ID = fmt.Sprintf("adj-line-%03d", index+1)
		}
		lines = append(lines, line)
	}
	adjustment.Lines = lines

	if err := adjustment.Validate(); err != nil {
		return StockAdjustment{}, err
	}

	return adjustment, nil
}

func NewStockAdjustmentLine(input NewStockAdjustmentLineInput) (StockAdjustmentLine, error) {
	expectedQty, err := decimal.ParseQuantity(input.ExpectedQty.String())
	if err != nil || expectedQty.IsNegative() {
		return StockAdjustmentLine{}, ErrStockAdjustmentInvalidQuantity
	}
	countedQty, err := decimal.ParseQuantity(input.CountedQty.String())
	if err != nil || countedQty.IsNegative() {
		return StockAdjustmentLine{}, ErrStockAdjustmentInvalidQuantity
	}
	deltaQty, err := decimal.SubtractQuantity(countedQty, expectedQty)
	if err != nil {
		return StockAdjustmentLine{}, ErrStockAdjustmentInvalidQuantity
	}
	baseUOMCode, err := decimal.NormalizeUOMCode(input.BaseUOMCode)
	if err != nil {
		return StockAdjustmentLine{}, err
	}

	line := StockAdjustmentLine{
		ID:           strings.TrimSpace(input.ID),
		ItemID:       strings.TrimSpace(input.ItemID),
		SKU:          strings.ToUpper(strings.TrimSpace(input.SKU)),
		BatchID:      strings.TrimSpace(input.BatchID),
		BatchNo:      strings.TrimSpace(input.BatchNo),
		LocationID:   strings.TrimSpace(input.LocationID),
		LocationCode: strings.TrimSpace(input.LocationCode),
		ExpectedQty:  expectedQty,
		CountedQty:   countedQty,
		DeltaQty:     deltaQty,
		BaseUOMCode:  baseUOMCode,
		Reason:       strings.TrimSpace(input.Reason),
	}
	if line.SKU == "" || line.ExpectedQty == "" || line.CountedQty == "" || line.BaseUOMCode == "" {
		return StockAdjustmentLine{}, ErrStockAdjustmentRequiredField
	}

	return line, nil
}

func (a StockAdjustment) Validate() error {
	if strings.TrimSpace(a.ID) == "" ||
		strings.TrimSpace(a.AdjustmentNo) == "" ||
		strings.TrimSpace(a.OrgID) == "" ||
		strings.TrimSpace(a.WarehouseID) == "" ||
		strings.TrimSpace(a.Reason) == "" ||
		strings.TrimSpace(a.RequestedBy) == "" ||
		len(a.Lines) == 0 {
		return ErrStockAdjustmentRequiredField
	}
	if !a.HasVariance() {
		return ErrStockAdjustmentNoVariance
	}

	return nil
}

func (a StockAdjustment) HasVariance() bool {
	for _, line := range a.Lines {
		if !line.DeltaQty.IsZero() {
			return true
		}
	}

	return false
}

func (a StockAdjustment) Clone() StockAdjustment {
	clone := a
	clone.Lines = append([]StockAdjustmentLine(nil), a.Lines...)
	return clone
}

func SortStockAdjustments(rows []StockAdjustment) {
	sort.SliceStable(rows, func(i int, j int) bool {
		left := rows[i]
		right := rows[j]
		if !left.CreatedAt.Equal(right.CreatedAt) {
			return left.CreatedAt.After(right.CreatedAt)
		}

		return left.AdjustmentNo > right.AdjustmentNo
	})
}
