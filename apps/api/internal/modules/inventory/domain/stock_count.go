package domain

import (
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/Chinsusu/ERP-v2/apps/api/internal/shared/decimal"
)

type StockCountStatus string

const StockCountStatusOpen StockCountStatus = "open"
const StockCountStatusSubmitted StockCountStatus = "submitted"
const StockCountStatusVarianceReview StockCountStatus = "variance_review"

var ErrStockCountRequiredField = errors.New("stock count required field is missing")
var ErrStockCountInvalidQuantity = errors.New("stock count quantity is invalid")
var ErrStockCountInvalidStatus = errors.New("stock count status is invalid")
var ErrStockCountLineNotFound = errors.New("stock count line not found")

type StockCountSession struct {
	ID            string
	CountNo       string
	OrgID         string
	WarehouseID   string
	WarehouseCode string
	Scope         string
	Status        StockCountStatus
	CreatedBy     string
	SubmittedBy   string
	Lines         []StockCountLine
	CreatedAt     time.Time
	UpdatedAt     time.Time
	SubmittedAt   time.Time
}

type StockCountLine struct {
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
	Counted      bool
	Note         string
}

type NewStockCountSessionInput struct {
	ID            string
	CountNo       string
	OrgID         string
	WarehouseID   string
	WarehouseCode string
	Scope         string
	CreatedBy     string
	Lines         []NewStockCountLineInput
	CreatedAt     time.Time
}

type NewStockCountLineInput struct {
	ID           string
	ItemID       string
	SKU          string
	BatchID      string
	BatchNo      string
	LocationID   string
	LocationCode string
	ExpectedQty  decimal.Decimal
	BaseUOMCode  string
}

type SubmitStockCountLineInput struct {
	ID         string
	SKU        string
	CountedQty decimal.Decimal
	Note       string
}

func NewStockCountSession(input NewStockCountSessionInput) (StockCountSession, error) {
	createdAt := input.CreatedAt
	if createdAt.IsZero() {
		createdAt = time.Now().UTC()
	}
	session := StockCountSession{
		ID:            strings.TrimSpace(input.ID),
		CountNo:       strings.TrimSpace(input.CountNo),
		OrgID:         strings.TrimSpace(input.OrgID),
		WarehouseID:   strings.TrimSpace(input.WarehouseID),
		WarehouseCode: strings.ToUpper(strings.TrimSpace(input.WarehouseCode)),
		Scope:         strings.TrimSpace(input.Scope),
		Status:        StockCountStatusOpen,
		CreatedBy:     strings.TrimSpace(input.CreatedBy),
		CreatedAt:     createdAt.UTC(),
		UpdatedAt:     createdAt.UTC(),
	}
	if session.OrgID == "" {
		session.OrgID = "org-my-pham"
	}
	if session.WarehouseCode == "" {
		session.WarehouseCode = strings.ToUpper(session.WarehouseID)
	}
	if session.Scope == "" {
		session.Scope = "warehouse"
	}
	if session.ID == "" {
		session.ID = fmt.Sprintf("count_%d", createdAt.UTC().UnixNano())
	}
	if session.CountNo == "" {
		session.CountNo = fmt.Sprintf("CNT-%s-%06d", createdAt.UTC().Format("060102"), createdAt.UTC().UnixNano()%1000000)
	}

	lines := make([]StockCountLine, 0, len(input.Lines))
	for index, lineInput := range input.Lines {
		line, err := NewStockCountLine(lineInput)
		if err != nil {
			return StockCountSession{}, err
		}
		if line.ID == "" {
			line.ID = fmt.Sprintf("count-line-%03d", index+1)
		}
		lines = append(lines, line)
	}
	session.Lines = lines

	if err := session.Validate(); err != nil {
		return StockCountSession{}, err
	}

	return session, nil
}

func NewStockCountLine(input NewStockCountLineInput) (StockCountLine, error) {
	expectedQty, err := decimal.ParseQuantity(input.ExpectedQty.String())
	if err != nil || expectedQty.IsNegative() {
		return StockCountLine{}, ErrStockCountInvalidQuantity
	}
	baseUOMCode, err := decimal.NormalizeUOMCode(input.BaseUOMCode)
	if err != nil {
		return StockCountLine{}, err
	}
	line := StockCountLine{
		ID:           strings.TrimSpace(input.ID),
		ItemID:       strings.TrimSpace(input.ItemID),
		SKU:          strings.ToUpper(strings.TrimSpace(input.SKU)),
		BatchID:      strings.TrimSpace(input.BatchID),
		BatchNo:      strings.TrimSpace(input.BatchNo),
		LocationID:   strings.TrimSpace(input.LocationID),
		LocationCode: strings.TrimSpace(input.LocationCode),
		ExpectedQty:  expectedQty,
		CountedQty:   decimal.MustQuantity("0"),
		DeltaQty:     decimal.MustQuantity("0"),
		BaseUOMCode:  baseUOMCode,
	}
	if line.SKU == "" || line.BaseUOMCode == "" {
		return StockCountLine{}, ErrStockCountRequiredField
	}

	return line, nil
}

func (s StockCountSession) Submit(
	lines []SubmitStockCountLineInput,
	actorID string,
	submittedAt time.Time,
) (StockCountSession, error) {
	if s.Status != StockCountStatusOpen {
		return StockCountSession{}, ErrStockCountInvalidStatus
	}
	if strings.TrimSpace(actorID) == "" || len(lines) == 0 {
		return StockCountSession{}, ErrStockCountRequiredField
	}
	if submittedAt.IsZero() {
		submittedAt = time.Now().UTC()
	}

	updated := s.Clone()
	for _, input := range lines {
		index := updated.findLineIndex(input)
		if index < 0 {
			return StockCountSession{}, ErrStockCountLineNotFound
		}
		countedQty, err := decimal.ParseQuantity(input.CountedQty.String())
		if err != nil || countedQty.IsNegative() {
			return StockCountSession{}, ErrStockCountInvalidQuantity
		}
		deltaQty, err := decimal.SubtractQuantity(countedQty, updated.Lines[index].ExpectedQty)
		if err != nil {
			return StockCountSession{}, ErrStockCountInvalidQuantity
		}
		updated.Lines[index].CountedQty = countedQty
		updated.Lines[index].DeltaQty = deltaQty
		updated.Lines[index].Counted = true
		updated.Lines[index].Note = strings.TrimSpace(input.Note)
	}
	for _, line := range updated.Lines {
		if !line.Counted {
			return StockCountSession{}, ErrStockCountRequiredField
		}
	}

	updated.Status = StockCountStatusSubmitted
	if updated.HasVariance() {
		updated.Status = StockCountStatusVarianceReview
	}
	updated.SubmittedBy = strings.TrimSpace(actorID)
	updated.SubmittedAt = submittedAt.UTC()
	updated.UpdatedAt = submittedAt.UTC()

	return updated, nil
}

func (s StockCountSession) Validate() error {
	if strings.TrimSpace(s.ID) == "" ||
		strings.TrimSpace(s.CountNo) == "" ||
		strings.TrimSpace(s.OrgID) == "" ||
		strings.TrimSpace(s.WarehouseID) == "" ||
		strings.TrimSpace(s.CreatedBy) == "" ||
		len(s.Lines) == 0 {
		return ErrStockCountRequiredField
	}

	return nil
}

func (s StockCountSession) HasVariance() bool {
	for _, line := range s.Lines {
		if !line.DeltaQty.IsZero() {
			return true
		}
	}

	return false
}

func (s StockCountSession) Clone() StockCountSession {
	clone := s
	clone.Lines = append([]StockCountLine(nil), s.Lines...)
	return clone
}

func (s StockCountSession) findLineIndex(input SubmitStockCountLineInput) int {
	id := strings.TrimSpace(input.ID)
	sku := strings.ToUpper(strings.TrimSpace(input.SKU))
	for index, line := range s.Lines {
		if id != "" && strings.TrimSpace(line.ID) == id {
			return index
		}
		if id == "" && sku != "" && strings.ToUpper(strings.TrimSpace(line.SKU)) == sku {
			return index
		}
	}

	return -1
}

func SortStockCountSessions(rows []StockCountSession) {
	sort.SliceStable(rows, func(i int, j int) bool {
		left := rows[i]
		right := rows[j]
		if !left.CreatedAt.Equal(right.CreatedAt) {
			return left.CreatedAt.After(right.CreatedAt)
		}

		return left.CountNo > right.CountNo
	})
}
