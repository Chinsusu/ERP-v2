package domain

import (
	"errors"
	"strings"
	"time"

	"github.com/Chinsusu/ERP-v2/apps/api/internal/shared/decimal"
)

var ErrSubcontractFinishedGoodsReceiptRequiredField = errors.New("subcontract finished goods receipt required field is missing")
var ErrSubcontractFinishedGoodsReceiptInvalidStatus = errors.New("subcontract finished goods receipt status is invalid")
var ErrSubcontractFinishedGoodsReceiptInvalidQuantity = errors.New("subcontract finished goods receipt quantity is invalid")

type SubcontractFinishedGoodsReceiptStatus string

const (
	SubcontractFinishedGoodsReceiptStatusQCHold SubcontractFinishedGoodsReceiptStatus = "qc_hold"
)

type SubcontractFinishedGoodsReceipt struct {
	ID                 string
	OrgID              string
	ReceiptNo          string
	SubcontractOrderID string
	SubcontractOrderNo string
	FactoryID          string
	FactoryCode        string
	FactoryName        string
	WarehouseID        string
	WarehouseCode      string
	LocationID         string
	LocationCode       string
	DeliveryNoteNo     string
	Status             SubcontractFinishedGoodsReceiptStatus
	Lines              []SubcontractFinishedGoodsReceiptLine
	Evidence           []SubcontractFinishedGoodsReceiptEvidence
	ReceivedBy         string
	ReceivedAt         time.Time
	Note               string
	CreatedAt          time.Time
	CreatedBy          string
	UpdatedAt          time.Time
	UpdatedBy          string
	Version            int
}

type SubcontractFinishedGoodsReceiptLine struct {
	ID               string
	LineNo           int
	ItemID           string
	SKUCode          string
	ItemName         string
	BatchID          string
	BatchNo          string
	LotNo            string
	ExpiryDate       time.Time
	ReceiveQty       decimal.Decimal
	UOMCode          decimal.UOMCode
	BaseReceiveQty   decimal.Decimal
	BaseUOMCode      decimal.UOMCode
	ConversionFactor decimal.Decimal
	PackagingStatus  string
	Note             string
}

type SubcontractFinishedGoodsReceiptEvidence struct {
	ID           string
	EvidenceType string
	FileName     string
	ObjectKey    string
	ExternalURL  string
	Note         string
}

type NewSubcontractFinishedGoodsReceiptInput struct {
	ID                 string
	OrgID              string
	ReceiptNo          string
	SubcontractOrderID string
	SubcontractOrderNo string
	FactoryID          string
	FactoryCode        string
	FactoryName        string
	WarehouseID        string
	WarehouseCode      string
	LocationID         string
	LocationCode       string
	DeliveryNoteNo     string
	Status             SubcontractFinishedGoodsReceiptStatus
	Lines              []NewSubcontractFinishedGoodsReceiptLineInput
	Evidence           []NewSubcontractFinishedGoodsReceiptEvidenceInput
	ReceivedBy         string
	ReceivedAt         time.Time
	Note               string
	CreatedAt          time.Time
	CreatedBy          string
	UpdatedAt          time.Time
	UpdatedBy          string
}

type NewSubcontractFinishedGoodsReceiptLineInput struct {
	ID               string
	LineNo           int
	ItemID           string
	SKUCode          string
	ItemName         string
	BatchID          string
	BatchNo          string
	LotNo            string
	ExpiryDate       time.Time
	ReceiveQty       decimal.Decimal
	UOMCode          string
	BaseReceiveQty   decimal.Decimal
	BaseUOMCode      string
	ConversionFactor decimal.Decimal
	PackagingStatus  string
	Note             string
}

type NewSubcontractFinishedGoodsReceiptEvidenceInput struct {
	ID           string
	EvidenceType string
	FileName     string
	ObjectKey    string
	ExternalURL  string
	Note         string
}

func NewSubcontractFinishedGoodsReceipt(input NewSubcontractFinishedGoodsReceiptInput) (SubcontractFinishedGoodsReceipt, error) {
	createdAt := input.CreatedAt
	if createdAt.IsZero() {
		createdAt = time.Now().UTC()
	}
	receivedAt := input.ReceivedAt
	if receivedAt.IsZero() {
		receivedAt = createdAt
	}
	updatedAt := input.UpdatedAt
	if updatedAt.IsZero() {
		updatedAt = createdAt
	}
	updatedBy := strings.TrimSpace(input.UpdatedBy)
	if updatedBy == "" {
		updatedBy = strings.TrimSpace(input.CreatedBy)
	}
	status := NormalizeSubcontractFinishedGoodsReceiptStatus(input.Status)
	if status == "" {
		status = SubcontractFinishedGoodsReceiptStatusQCHold
	}

	receipt := SubcontractFinishedGoodsReceipt{
		ID:                 strings.TrimSpace(input.ID),
		OrgID:              strings.TrimSpace(input.OrgID),
		ReceiptNo:          strings.ToUpper(strings.TrimSpace(input.ReceiptNo)),
		SubcontractOrderID: strings.TrimSpace(input.SubcontractOrderID),
		SubcontractOrderNo: strings.ToUpper(strings.TrimSpace(input.SubcontractOrderNo)),
		FactoryID:          strings.TrimSpace(input.FactoryID),
		FactoryCode:        strings.ToUpper(strings.TrimSpace(input.FactoryCode)),
		FactoryName:        strings.TrimSpace(input.FactoryName),
		WarehouseID:        strings.TrimSpace(input.WarehouseID),
		WarehouseCode:      strings.ToUpper(strings.TrimSpace(input.WarehouseCode)),
		LocationID:         strings.TrimSpace(input.LocationID),
		LocationCode:       strings.ToUpper(strings.TrimSpace(input.LocationCode)),
		DeliveryNoteNo:     strings.ToUpper(strings.TrimSpace(input.DeliveryNoteNo)),
		Status:             status,
		Lines:              make([]SubcontractFinishedGoodsReceiptLine, 0, len(input.Lines)),
		Evidence:           make([]SubcontractFinishedGoodsReceiptEvidence, 0, len(input.Evidence)),
		ReceivedBy:         strings.TrimSpace(input.ReceivedBy),
		ReceivedAt:         receivedAt.UTC(),
		Note:               strings.TrimSpace(input.Note),
		CreatedAt:          createdAt.UTC(),
		CreatedBy:          strings.TrimSpace(input.CreatedBy),
		UpdatedAt:          updatedAt.UTC(),
		UpdatedBy:          updatedBy,
		Version:            1,
	}
	for index, lineInput := range input.Lines {
		if lineInput.LineNo == 0 {
			lineInput.LineNo = index + 1
		}
		line, err := NewSubcontractFinishedGoodsReceiptLine(lineInput)
		if err != nil {
			return SubcontractFinishedGoodsReceipt{}, err
		}
		receipt.Lines = append(receipt.Lines, line)
	}
	for _, evidenceInput := range input.Evidence {
		evidence, err := NewSubcontractFinishedGoodsReceiptEvidence(evidenceInput)
		if err != nil {
			return SubcontractFinishedGoodsReceipt{}, err
		}
		receipt.Evidence = append(receipt.Evidence, evidence)
	}
	if err := receipt.Validate(); err != nil {
		return SubcontractFinishedGoodsReceipt{}, err
	}

	return receipt, nil
}

func NewSubcontractFinishedGoodsReceiptLine(input NewSubcontractFinishedGoodsReceiptLineInput) (SubcontractFinishedGoodsReceiptLine, error) {
	receiveQty, baseReceiveQty, uomCode, baseUOMCode, conversionFactor, err := normalizeSubcontractOrderQuantitySet(
		input.ReceiveQty,
		input.BaseReceiveQty,
		input.UOMCode,
		input.BaseUOMCode,
		input.ConversionFactor,
		true,
	)
	if err != nil {
		return SubcontractFinishedGoodsReceiptLine{}, ErrSubcontractFinishedGoodsReceiptInvalidQuantity
	}
	line := SubcontractFinishedGoodsReceiptLine{
		ID:               strings.TrimSpace(input.ID),
		LineNo:           input.LineNo,
		ItemID:           strings.TrimSpace(input.ItemID),
		SKUCode:          strings.ToUpper(strings.TrimSpace(input.SKUCode)),
		ItemName:         strings.TrimSpace(input.ItemName),
		BatchID:          strings.TrimSpace(input.BatchID),
		BatchNo:          strings.ToUpper(strings.TrimSpace(input.BatchNo)),
		LotNo:            strings.ToUpper(strings.TrimSpace(input.LotNo)),
		ExpiryDate:       subcontractFinishedGoodsReceiptDateOnly(input.ExpiryDate),
		ReceiveQty:       receiveQty,
		UOMCode:          uomCode,
		BaseReceiveQty:   baseReceiveQty,
		BaseUOMCode:      baseUOMCode,
		ConversionFactor: conversionFactor,
		PackagingStatus:  strings.ToLower(strings.TrimSpace(input.PackagingStatus)),
		Note:             strings.TrimSpace(input.Note),
	}
	if line.LotNo == "" {
		line.LotNo = line.BatchNo
	}
	if err := line.Validate(); err != nil {
		return SubcontractFinishedGoodsReceiptLine{}, err
	}

	return line, nil
}

func NewSubcontractFinishedGoodsReceiptEvidence(
	input NewSubcontractFinishedGoodsReceiptEvidenceInput,
) (SubcontractFinishedGoodsReceiptEvidence, error) {
	evidence := SubcontractFinishedGoodsReceiptEvidence{
		ID:           strings.TrimSpace(input.ID),
		EvidenceType: strings.ToLower(strings.TrimSpace(input.EvidenceType)),
		FileName:     strings.TrimSpace(input.FileName),
		ObjectKey:    strings.TrimSpace(input.ObjectKey),
		ExternalURL:  strings.TrimSpace(input.ExternalURL),
		Note:         strings.TrimSpace(input.Note),
	}
	if err := evidence.Validate(); err != nil {
		return SubcontractFinishedGoodsReceiptEvidence{}, err
	}

	return evidence, nil
}

func (r SubcontractFinishedGoodsReceipt) Validate() error {
	if strings.TrimSpace(r.ID) == "" ||
		strings.TrimSpace(r.OrgID) == "" ||
		strings.TrimSpace(r.ReceiptNo) == "" ||
		strings.TrimSpace(r.SubcontractOrderID) == "" ||
		strings.TrimSpace(r.SubcontractOrderNo) == "" ||
		strings.TrimSpace(r.FactoryID) == "" ||
		strings.TrimSpace(r.FactoryName) == "" ||
		strings.TrimSpace(r.WarehouseID) == "" ||
		strings.TrimSpace(r.LocationID) == "" ||
		strings.TrimSpace(r.DeliveryNoteNo) == "" ||
		strings.TrimSpace(r.ReceivedBy) == "" ||
		strings.TrimSpace(r.CreatedBy) == "" ||
		len(r.Lines) == 0 {
		return ErrSubcontractFinishedGoodsReceiptRequiredField
	}
	if !IsValidSubcontractFinishedGoodsReceiptStatus(r.Status) {
		return ErrSubcontractFinishedGoodsReceiptInvalidStatus
	}
	seenLineNo := map[int]struct{}{}
	for _, line := range r.Lines {
		if _, exists := seenLineNo[line.LineNo]; exists {
			return ErrSubcontractFinishedGoodsReceiptRequiredField
		}
		seenLineNo[line.LineNo] = struct{}{}
		if err := line.Validate(); err != nil {
			return err
		}
	}
	for _, evidence := range r.Evidence {
		if err := evidence.Validate(); err != nil {
			return err
		}
	}

	return nil
}

func (l SubcontractFinishedGoodsReceiptLine) Validate() error {
	if strings.TrimSpace(l.ID) == "" ||
		l.LineNo <= 0 ||
		strings.TrimSpace(l.ItemID) == "" ||
		strings.TrimSpace(l.SKUCode) == "" ||
		strings.TrimSpace(l.ItemName) == "" ||
		strings.TrimSpace(l.BatchID) == "" ||
		strings.TrimSpace(l.BatchNo) == "" ||
		strings.TrimSpace(l.LotNo) == "" ||
		l.ExpiryDate.IsZero() {
		return ErrSubcontractFinishedGoodsReceiptRequiredField
	}
	for _, quantity := range []decimal.Decimal{l.ReceiveQty, l.BaseReceiveQty, l.ConversionFactor} {
		value, err := decimal.ParseQuantity(quantity.String())
		if err != nil || value.IsNegative() || value.IsZero() {
			return ErrSubcontractFinishedGoodsReceiptInvalidQuantity
		}
	}
	if _, err := decimal.NormalizeUOMCode(l.UOMCode.String()); err != nil {
		return ErrSubcontractFinishedGoodsReceiptInvalidQuantity
	}
	if _, err := decimal.NormalizeUOMCode(l.BaseUOMCode.String()); err != nil {
		return ErrSubcontractFinishedGoodsReceiptInvalidQuantity
	}

	return nil
}

func (e SubcontractFinishedGoodsReceiptEvidence) Validate() error {
	if strings.TrimSpace(e.ID) == "" ||
		strings.TrimSpace(e.EvidenceType) == "" ||
		(strings.TrimSpace(e.ObjectKey) == "" && strings.TrimSpace(e.ExternalURL) == "") {
		return ErrSubcontractFinishedGoodsReceiptRequiredField
	}

	return nil
}

func (r SubcontractFinishedGoodsReceipt) Clone() SubcontractFinishedGoodsReceipt {
	clone := r
	clone.Lines = append([]SubcontractFinishedGoodsReceiptLine(nil), r.Lines...)
	clone.Evidence = append([]SubcontractFinishedGoodsReceiptEvidence(nil), r.Evidence...)

	return clone
}

func NormalizeSubcontractFinishedGoodsReceiptStatus(status SubcontractFinishedGoodsReceiptStatus) SubcontractFinishedGoodsReceiptStatus {
	return SubcontractFinishedGoodsReceiptStatus(strings.ToLower(strings.TrimSpace(string(status))))
}

func IsValidSubcontractFinishedGoodsReceiptStatus(status SubcontractFinishedGoodsReceiptStatus) bool {
	switch NormalizeSubcontractFinishedGoodsReceiptStatus(status) {
	case SubcontractFinishedGoodsReceiptStatusQCHold:
		return true
	default:
		return false
	}
}

func subcontractFinishedGoodsReceiptDateOnly(value time.Time) time.Time {
	if value.IsZero() {
		return time.Time{}
	}

	return time.Date(value.UTC().Year(), value.UTC().Month(), value.UTC().Day(), 0, 0, 0, 0, time.UTC)
}
