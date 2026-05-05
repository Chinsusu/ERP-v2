package domain

import (
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/Chinsusu/ERP-v2/apps/api/internal/shared/decimal"
)

type StockTransferStatus string

const StockTransferStatusDraft StockTransferStatus = "draft"
const StockTransferStatusSubmitted StockTransferStatus = "submitted"
const StockTransferStatusApproved StockTransferStatus = "approved"
const StockTransferStatusPosted StockTransferStatus = "posted"
const StockTransferStatusCancelled StockTransferStatus = "cancelled"

type WarehouseIssueStatus string

const WarehouseIssueStatusDraft WarehouseIssueStatus = "draft"
const WarehouseIssueStatusSubmitted WarehouseIssueStatus = "submitted"
const WarehouseIssueStatusApproved WarehouseIssueStatus = "approved"
const WarehouseIssueStatusPosted WarehouseIssueStatus = "posted"
const WarehouseIssueStatusCancelled WarehouseIssueStatus = "cancelled"

var ErrStockTransferRequiredField = errors.New("stock transfer required field is missing")
var ErrStockTransferInvalidQuantity = errors.New("stock transfer quantity is invalid")
var ErrStockTransferInvalidStatus = errors.New("stock transfer status transition is invalid")
var ErrStockTransferSameWarehouse = errors.New("stock transfer source and destination warehouse must differ")

var ErrWarehouseIssueRequiredField = errors.New("warehouse issue required field is missing")
var ErrWarehouseIssueInvalidQuantity = errors.New("warehouse issue quantity is invalid")
var ErrWarehouseIssueInvalidStatus = errors.New("warehouse issue status transition is invalid")

type StockTransfer struct {
	ID                       string
	TransferNo               string
	OrgID                    string
	SourceWarehouseID        string
	SourceWarehouseCode      string
	DestinationWarehouseID   string
	DestinationWarehouseCode string
	ReasonCode               string
	Status                   StockTransferStatus
	RequestedBy              string
	SubmittedBy              string
	ApprovedBy               string
	PostedBy                 string
	Lines                    []StockTransferLine
	CreatedAt                time.Time
	UpdatedAt                time.Time
	SubmittedAt              time.Time
	ApprovedAt               time.Time
	PostedAt                 time.Time
}

type StockTransferLine struct {
	ID                      string
	ItemID                  string
	SKU                     string
	BatchID                 string
	BatchNo                 string
	SourceLocationID        string
	SourceLocationCode      string
	DestinationLocationID   string
	DestinationLocationCode string
	Quantity                decimal.Decimal
	BaseUOMCode             decimal.UOMCode
	Note                    string
}

type NewStockTransferInput struct {
	ID                       string
	TransferNo               string
	OrgID                    string
	SourceWarehouseID        string
	SourceWarehouseCode      string
	DestinationWarehouseID   string
	DestinationWarehouseCode string
	ReasonCode               string
	RequestedBy              string
	Lines                    []NewStockTransferLineInput
	CreatedAt                time.Time
}

type NewStockTransferLineInput struct {
	ID                      string
	ItemID                  string
	SKU                     string
	BatchID                 string
	BatchNo                 string
	SourceLocationID        string
	SourceLocationCode      string
	DestinationLocationID   string
	DestinationLocationCode string
	Quantity                decimal.Decimal
	BaseUOMCode             string
	Note                    string
}

type WarehouseIssue struct {
	ID              string
	IssueNo         string
	OrgID           string
	WarehouseID     string
	WarehouseCode   string
	DestinationType string
	DestinationName string
	ReasonCode      string
	Status          WarehouseIssueStatus
	RequestedBy     string
	SubmittedBy     string
	ApprovedBy      string
	PostedBy        string
	Lines           []WarehouseIssueLine
	CreatedAt       time.Time
	UpdatedAt       time.Time
	SubmittedAt     time.Time
	ApprovedAt      time.Time
	PostedAt        time.Time
}

type WarehouseIssueLine struct {
	ID                 string
	ItemID             string
	SKU                string
	ItemName           string
	Category           string
	BatchID            string
	BatchNo            string
	LocationID         string
	LocationCode       string
	Quantity           decimal.Decimal
	BaseUOMCode        decimal.UOMCode
	Specification      string
	SourceDocumentType string
	SourceDocumentID   string
	Note               string
}

type NewWarehouseIssueInput struct {
	ID              string
	IssueNo         string
	OrgID           string
	WarehouseID     string
	WarehouseCode   string
	DestinationType string
	DestinationName string
	ReasonCode      string
	RequestedBy     string
	Lines           []NewWarehouseIssueLineInput
	CreatedAt       time.Time
}

type NewWarehouseIssueLineInput struct {
	ID                 string
	ItemID             string
	SKU                string
	ItemName           string
	Category           string
	BatchID            string
	BatchNo            string
	LocationID         string
	LocationCode       string
	Quantity           decimal.Decimal
	BaseUOMCode        string
	Specification      string
	SourceDocumentType string
	SourceDocumentID   string
	Note               string
}

func NewStockTransfer(input NewStockTransferInput) (StockTransfer, error) {
	createdAt := normalizedCreatedAt(input.CreatedAt)
	transfer := StockTransfer{
		ID:                       strings.TrimSpace(input.ID),
		TransferNo:               strings.TrimSpace(input.TransferNo),
		OrgID:                    strings.TrimSpace(input.OrgID),
		SourceWarehouseID:        strings.TrimSpace(input.SourceWarehouseID),
		SourceWarehouseCode:      strings.ToUpper(strings.TrimSpace(input.SourceWarehouseCode)),
		DestinationWarehouseID:   strings.TrimSpace(input.DestinationWarehouseID),
		DestinationWarehouseCode: strings.ToUpper(strings.TrimSpace(input.DestinationWarehouseCode)),
		ReasonCode:               strings.TrimSpace(input.ReasonCode),
		Status:                   StockTransferStatusDraft,
		RequestedBy:              strings.TrimSpace(input.RequestedBy),
		CreatedAt:                createdAt,
		UpdatedAt:                createdAt,
	}
	if transfer.OrgID == "" {
		transfer.OrgID = "org-my-pham"
	}
	if transfer.ID == "" {
		transfer.ID = fmt.Sprintf("stock-transfer-%d", createdAt.UnixNano())
	}
	if transfer.TransferNo == "" {
		transfer.TransferNo = fmt.Sprintf("ST-%s-%06d", createdAt.Format("060102"), createdAt.UnixNano()%1000000)
	}
	for index, lineInput := range input.Lines {
		line, err := NewStockTransferLine(lineInput)
		if err != nil {
			return StockTransfer{}, err
		}
		if line.ID == "" {
			line.ID = fmt.Sprintf("transfer-line-%03d", index+1)
		}
		transfer.Lines = append(transfer.Lines, line)
	}
	if err := transfer.Validate(); err != nil {
		return StockTransfer{}, err
	}

	return transfer, nil
}

func NewStockTransferLine(input NewStockTransferLineInput) (StockTransferLine, error) {
	quantity, err := decimal.ParseQuantity(input.Quantity.String())
	if err != nil || quantity.IsZero() || quantity.IsNegative() {
		return StockTransferLine{}, ErrStockTransferInvalidQuantity
	}
	uom, err := decimal.NormalizeUOMCode(input.BaseUOMCode)
	if err != nil {
		return StockTransferLine{}, err
	}
	line := StockTransferLine{
		ID:                      strings.TrimSpace(input.ID),
		ItemID:                  strings.TrimSpace(input.ItemID),
		SKU:                     strings.ToUpper(strings.TrimSpace(input.SKU)),
		BatchID:                 strings.TrimSpace(input.BatchID),
		BatchNo:                 strings.TrimSpace(input.BatchNo),
		SourceLocationID:        strings.TrimSpace(input.SourceLocationID),
		SourceLocationCode:      strings.TrimSpace(input.SourceLocationCode),
		DestinationLocationID:   strings.TrimSpace(input.DestinationLocationID),
		DestinationLocationCode: strings.TrimSpace(input.DestinationLocationCode),
		Quantity:                quantity,
		BaseUOMCode:             uom,
		Note:                    strings.TrimSpace(input.Note),
	}
	if line.SKU == "" || line.BaseUOMCode == "" {
		return StockTransferLine{}, ErrStockTransferRequiredField
	}

	return line, nil
}

func (t StockTransfer) Validate() error {
	if strings.TrimSpace(t.ID) == "" ||
		strings.TrimSpace(t.TransferNo) == "" ||
		strings.TrimSpace(t.OrgID) == "" ||
		strings.TrimSpace(t.SourceWarehouseID) == "" ||
		strings.TrimSpace(t.DestinationWarehouseID) == "" ||
		strings.TrimSpace(t.ReasonCode) == "" ||
		strings.TrimSpace(t.RequestedBy) == "" ||
		len(t.Lines) == 0 {
		return ErrStockTransferRequiredField
	}
	if t.SourceWarehouseID == t.DestinationWarehouseID {
		return ErrStockTransferSameWarehouse
	}

	return nil
}

func (t StockTransfer) Submit(actorID string, at time.Time) (StockTransfer, error) {
	if t.Status != StockTransferStatusDraft {
		return StockTransfer{}, ErrStockTransferInvalidStatus
	}
	return t.transition(StockTransferStatusSubmitted, actorID, at, func(updated *StockTransfer, actor string, when time.Time) {
		updated.SubmittedBy = actor
		updated.SubmittedAt = when
	})
}

func (t StockTransfer) Approve(actorID string, at time.Time) (StockTransfer, error) {
	if t.Status != StockTransferStatusSubmitted {
		return StockTransfer{}, ErrStockTransferInvalidStatus
	}
	return t.transition(StockTransferStatusApproved, actorID, at, func(updated *StockTransfer, actor string, when time.Time) {
		updated.ApprovedBy = actor
		updated.ApprovedAt = when
	})
}

func (t StockTransfer) MarkPosted(actorID string, at time.Time) (StockTransfer, error) {
	if t.Status != StockTransferStatusApproved {
		return StockTransfer{}, ErrStockTransferInvalidStatus
	}
	return t.transition(StockTransferStatusPosted, actorID, at, func(updated *StockTransfer, actor string, when time.Time) {
		updated.PostedBy = actor
		updated.PostedAt = when
	})
}

func (t StockTransfer) Clone() StockTransfer {
	clone := t
	clone.Lines = append([]StockTransferLine(nil), t.Lines...)
	return clone
}

func (t StockTransfer) transition(
	status StockTransferStatus,
	actorID string,
	at time.Time,
	apply func(updated *StockTransfer, actor string, when time.Time),
) (StockTransfer, error) {
	actorID = strings.TrimSpace(actorID)
	if actorID == "" {
		return StockTransfer{}, ErrStockTransferRequiredField
	}
	when := normalizedCreatedAt(at)
	updated := t.Clone()
	updated.Status = status
	updated.UpdatedAt = when
	apply(&updated, actorID, when)

	return updated, nil
}

func NewWarehouseIssue(input NewWarehouseIssueInput) (WarehouseIssue, error) {
	createdAt := normalizedCreatedAt(input.CreatedAt)
	issue := WarehouseIssue{
		ID:              strings.TrimSpace(input.ID),
		IssueNo:         strings.TrimSpace(input.IssueNo),
		OrgID:           strings.TrimSpace(input.OrgID),
		WarehouseID:     strings.TrimSpace(input.WarehouseID),
		WarehouseCode:   strings.ToUpper(strings.TrimSpace(input.WarehouseCode)),
		DestinationType: strings.TrimSpace(input.DestinationType),
		DestinationName: strings.TrimSpace(input.DestinationName),
		ReasonCode:      strings.TrimSpace(input.ReasonCode),
		Status:          WarehouseIssueStatusDraft,
		RequestedBy:     strings.TrimSpace(input.RequestedBy),
		CreatedAt:       createdAt,
		UpdatedAt:       createdAt,
	}
	if issue.OrgID == "" {
		issue.OrgID = "org-my-pham"
	}
	if issue.ID == "" {
		issue.ID = fmt.Sprintf("warehouse-issue-%d", createdAt.UnixNano())
	}
	if issue.IssueNo == "" {
		issue.IssueNo = fmt.Sprintf("WI-%s-%06d", createdAt.Format("060102"), createdAt.UnixNano()%1000000)
	}
	for index, lineInput := range input.Lines {
		line, err := NewWarehouseIssueLine(lineInput)
		if err != nil {
			return WarehouseIssue{}, err
		}
		if line.ID == "" {
			line.ID = fmt.Sprintf("issue-line-%03d", index+1)
		}
		issue.Lines = append(issue.Lines, line)
	}
	if err := issue.Validate(); err != nil {
		return WarehouseIssue{}, err
	}

	return issue, nil
}

func NewWarehouseIssueLine(input NewWarehouseIssueLineInput) (WarehouseIssueLine, error) {
	quantity, err := decimal.ParseQuantity(input.Quantity.String())
	if err != nil || quantity.IsZero() || quantity.IsNegative() {
		return WarehouseIssueLine{}, ErrWarehouseIssueInvalidQuantity
	}
	uom, err := decimal.NormalizeUOMCode(input.BaseUOMCode)
	if err != nil {
		return WarehouseIssueLine{}, err
	}
	line := WarehouseIssueLine{
		ID:                 strings.TrimSpace(input.ID),
		ItemID:             strings.TrimSpace(input.ItemID),
		SKU:                strings.ToUpper(strings.TrimSpace(input.SKU)),
		ItemName:           strings.TrimSpace(input.ItemName),
		Category:           strings.TrimSpace(input.Category),
		BatchID:            strings.TrimSpace(input.BatchID),
		BatchNo:            strings.TrimSpace(input.BatchNo),
		LocationID:         strings.TrimSpace(input.LocationID),
		LocationCode:       strings.TrimSpace(input.LocationCode),
		Quantity:           quantity,
		BaseUOMCode:        uom,
		Specification:      strings.TrimSpace(input.Specification),
		SourceDocumentType: strings.TrimSpace(input.SourceDocumentType),
		SourceDocumentID:   strings.TrimSpace(input.SourceDocumentID),
		Note:               strings.TrimSpace(input.Note),
	}
	if line.SKU == "" || line.BaseUOMCode == "" {
		return WarehouseIssueLine{}, ErrWarehouseIssueRequiredField
	}

	return line, nil
}

func (i WarehouseIssue) Validate() error {
	if strings.TrimSpace(i.ID) == "" ||
		strings.TrimSpace(i.IssueNo) == "" ||
		strings.TrimSpace(i.OrgID) == "" ||
		strings.TrimSpace(i.WarehouseID) == "" ||
		strings.TrimSpace(i.DestinationType) == "" ||
		strings.TrimSpace(i.DestinationName) == "" ||
		strings.TrimSpace(i.ReasonCode) == "" ||
		strings.TrimSpace(i.RequestedBy) == "" ||
		len(i.Lines) == 0 {
		return ErrWarehouseIssueRequiredField
	}

	return nil
}

func (i WarehouseIssue) Submit(actorID string, at time.Time) (WarehouseIssue, error) {
	if i.Status != WarehouseIssueStatusDraft {
		return WarehouseIssue{}, ErrWarehouseIssueInvalidStatus
	}
	return i.transition(WarehouseIssueStatusSubmitted, actorID, at, func(updated *WarehouseIssue, actor string, when time.Time) {
		updated.SubmittedBy = actor
		updated.SubmittedAt = when
	})
}

func (i WarehouseIssue) Approve(actorID string, at time.Time) (WarehouseIssue, error) {
	if i.Status != WarehouseIssueStatusSubmitted {
		return WarehouseIssue{}, ErrWarehouseIssueInvalidStatus
	}
	return i.transition(WarehouseIssueStatusApproved, actorID, at, func(updated *WarehouseIssue, actor string, when time.Time) {
		updated.ApprovedBy = actor
		updated.ApprovedAt = when
	})
}

func (i WarehouseIssue) MarkPosted(actorID string, at time.Time) (WarehouseIssue, error) {
	if i.Status != WarehouseIssueStatusApproved {
		return WarehouseIssue{}, ErrWarehouseIssueInvalidStatus
	}
	return i.transition(WarehouseIssueStatusPosted, actorID, at, func(updated *WarehouseIssue, actor string, when time.Time) {
		updated.PostedBy = actor
		updated.PostedAt = when
	})
}

func (i WarehouseIssue) Clone() WarehouseIssue {
	clone := i
	clone.Lines = append([]WarehouseIssueLine(nil), i.Lines...)
	return clone
}

func (i WarehouseIssue) transition(
	status WarehouseIssueStatus,
	actorID string,
	at time.Time,
	apply func(updated *WarehouseIssue, actor string, when time.Time),
) (WarehouseIssue, error) {
	actorID = strings.TrimSpace(actorID)
	if actorID == "" {
		return WarehouseIssue{}, ErrWarehouseIssueRequiredField
	}
	when := normalizedCreatedAt(at)
	updated := i.Clone()
	updated.Status = status
	updated.UpdatedAt = when
	apply(&updated, actorID, when)

	return updated, nil
}

func SortStockTransfers(rows []StockTransfer) {
	sort.SliceStable(rows, func(i int, j int) bool {
		left := rows[i]
		right := rows[j]
		if !left.CreatedAt.Equal(right.CreatedAt) {
			return left.CreatedAt.After(right.CreatedAt)
		}

		return left.TransferNo > right.TransferNo
	})
}

func SortWarehouseIssues(rows []WarehouseIssue) {
	sort.SliceStable(rows, func(i int, j int) bool {
		left := rows[i]
		right := rows[j]
		if !left.CreatedAt.Equal(right.CreatedAt) {
			return left.CreatedAt.After(right.CreatedAt)
		}

		return left.IssueNo > right.IssueNo
	})
}

func normalizedCreatedAt(value time.Time) time.Time {
	if value.IsZero() {
		return time.Now().UTC()
	}

	return value.UTC()
}
