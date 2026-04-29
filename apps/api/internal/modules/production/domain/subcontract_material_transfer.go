package domain

import (
	"errors"
	"strings"
	"time"

	"github.com/Chinsusu/ERP-v2/apps/api/internal/shared/decimal"
)

var ErrSubcontractMaterialTransferRequiredField = errors.New("subcontract material transfer required field is missing")
var ErrSubcontractMaterialTransferInvalidStatus = errors.New("subcontract material transfer status is invalid")
var ErrSubcontractMaterialTransferInvalidQuantity = errors.New("subcontract material transfer quantity is invalid")
var ErrSubcontractMaterialTransferBatchRequired = errors.New("subcontract material transfer batch is required")

type SubcontractMaterialTransferStatus string

const (
	SubcontractMaterialTransferStatusSentToFactory SubcontractMaterialTransferStatus = "sent_to_factory"
	SubcontractMaterialTransferStatusPartiallySent SubcontractMaterialTransferStatus = "partially_sent"
)

type SubcontractMaterialTransfer struct {
	ID                  string
	OrgID               string
	TransferNo          string
	SubcontractOrderID  string
	SubcontractOrderNo  string
	FactoryID           string
	FactoryCode         string
	FactoryName         string
	SourceWarehouseID   string
	SourceWarehouseCode string
	Status              SubcontractMaterialTransferStatus
	Lines               []SubcontractMaterialTransferLine
	Evidence            []SubcontractMaterialTransferEvidence
	HandoverBy          string
	HandoverAt          time.Time
	ReceivedBy          string
	ReceiverContact     string
	VehicleNo           string
	Note                string
	CreatedAt           time.Time
	CreatedBy           string
	UpdatedAt           time.Time
	UpdatedBy           string
	Version             int
}

type SubcontractMaterialTransferLine struct {
	ID                  string
	LineNo              int
	OrderMaterialLineID string
	ItemID              string
	SKUCode             string
	ItemName            string
	IssueQty            decimal.Decimal
	UOMCode             decimal.UOMCode
	BaseIssueQty        decimal.Decimal
	BaseUOMCode         decimal.UOMCode
	ConversionFactor    decimal.Decimal
	BatchID             string
	BatchNo             string
	LotNo               string
	SourceBinID         string
	LotTraceRequired    bool
	Note                string
}

type SubcontractMaterialTransferEvidence struct {
	ID           string
	EvidenceType string
	FileName     string
	ObjectKey    string
	ExternalURL  string
	Note         string
}

type NewSubcontractMaterialTransferInput struct {
	ID                  string
	OrgID               string
	TransferNo          string
	SubcontractOrderID  string
	SubcontractOrderNo  string
	FactoryID           string
	FactoryCode         string
	FactoryName         string
	SourceWarehouseID   string
	SourceWarehouseCode string
	Status              SubcontractMaterialTransferStatus
	Lines               []NewSubcontractMaterialTransferLineInput
	Evidence            []NewSubcontractMaterialTransferEvidenceInput
	HandoverBy          string
	HandoverAt          time.Time
	ReceivedBy          string
	ReceiverContact     string
	VehicleNo           string
	Note                string
	CreatedAt           time.Time
	CreatedBy           string
	UpdatedAt           time.Time
	UpdatedBy           string
}

type NewSubcontractMaterialTransferLineInput struct {
	ID                  string
	LineNo              int
	OrderMaterialLineID string
	ItemID              string
	SKUCode             string
	ItemName            string
	IssueQty            decimal.Decimal
	UOMCode             string
	BaseIssueQty        decimal.Decimal
	BaseUOMCode         string
	ConversionFactor    decimal.Decimal
	BatchID             string
	BatchNo             string
	LotNo               string
	SourceBinID         string
	LotTraceRequired    bool
	Note                string
}

type NewSubcontractMaterialTransferEvidenceInput struct {
	ID           string
	EvidenceType string
	FileName     string
	ObjectKey    string
	ExternalURL  string
	Note         string
}

func NewSubcontractMaterialTransfer(input NewSubcontractMaterialTransferInput) (SubcontractMaterialTransfer, error) {
	createdAt := input.CreatedAt
	if createdAt.IsZero() {
		createdAt = time.Now().UTC()
	}
	handoverAt := input.HandoverAt
	if handoverAt.IsZero() {
		handoverAt = createdAt
	}
	updatedAt := input.UpdatedAt
	if updatedAt.IsZero() {
		updatedAt = createdAt
	}
	updatedBy := strings.TrimSpace(input.UpdatedBy)
	if updatedBy == "" {
		updatedBy = strings.TrimSpace(input.CreatedBy)
	}
	status := NormalizeSubcontractMaterialTransferStatus(input.Status)
	if status == "" {
		status = SubcontractMaterialTransferStatusSentToFactory
	}

	transfer := SubcontractMaterialTransfer{
		ID:                  strings.TrimSpace(input.ID),
		OrgID:               strings.TrimSpace(input.OrgID),
		TransferNo:          strings.ToUpper(strings.TrimSpace(input.TransferNo)),
		SubcontractOrderID:  strings.TrimSpace(input.SubcontractOrderID),
		SubcontractOrderNo:  strings.ToUpper(strings.TrimSpace(input.SubcontractOrderNo)),
		FactoryID:           strings.TrimSpace(input.FactoryID),
		FactoryCode:         strings.ToUpper(strings.TrimSpace(input.FactoryCode)),
		FactoryName:         strings.TrimSpace(input.FactoryName),
		SourceWarehouseID:   strings.TrimSpace(input.SourceWarehouseID),
		SourceWarehouseCode: strings.ToUpper(strings.TrimSpace(input.SourceWarehouseCode)),
		Status:              status,
		Lines:               make([]SubcontractMaterialTransferLine, 0, len(input.Lines)),
		Evidence:            make([]SubcontractMaterialTransferEvidence, 0, len(input.Evidence)),
		HandoverBy:          strings.TrimSpace(input.HandoverBy),
		HandoverAt:          handoverAt.UTC(),
		ReceivedBy:          strings.TrimSpace(input.ReceivedBy),
		ReceiverContact:     strings.TrimSpace(input.ReceiverContact),
		VehicleNo:           strings.ToUpper(strings.TrimSpace(input.VehicleNo)),
		Note:                strings.TrimSpace(input.Note),
		CreatedAt:           createdAt.UTC(),
		CreatedBy:           strings.TrimSpace(input.CreatedBy),
		UpdatedAt:           updatedAt.UTC(),
		UpdatedBy:           updatedBy,
		Version:             1,
	}
	for index, lineInput := range input.Lines {
		if lineInput.LineNo == 0 {
			lineInput.LineNo = index + 1
		}
		line, err := NewSubcontractMaterialTransferLine(lineInput)
		if err != nil {
			return SubcontractMaterialTransfer{}, err
		}
		transfer.Lines = append(transfer.Lines, line)
	}
	for _, evidenceInput := range input.Evidence {
		evidence, err := NewSubcontractMaterialTransferEvidence(evidenceInput)
		if err != nil {
			return SubcontractMaterialTransfer{}, err
		}
		transfer.Evidence = append(transfer.Evidence, evidence)
	}
	if err := transfer.Validate(); err != nil {
		return SubcontractMaterialTransfer{}, err
	}

	return transfer, nil
}

func NewSubcontractMaterialTransferLine(input NewSubcontractMaterialTransferLineInput) (SubcontractMaterialTransferLine, error) {
	issueQty, baseIssueQty, uomCode, baseUOMCode, conversionFactor, err := normalizeSubcontractOrderQuantitySet(
		input.IssueQty,
		input.BaseIssueQty,
		input.UOMCode,
		input.BaseUOMCode,
		input.ConversionFactor,
		true,
	)
	if err != nil {
		return SubcontractMaterialTransferLine{}, ErrSubcontractMaterialTransferInvalidQuantity
	}

	line := SubcontractMaterialTransferLine{
		ID:                  strings.TrimSpace(input.ID),
		LineNo:              input.LineNo,
		OrderMaterialLineID: strings.TrimSpace(input.OrderMaterialLineID),
		ItemID:              strings.TrimSpace(input.ItemID),
		SKUCode:             strings.ToUpper(strings.TrimSpace(input.SKUCode)),
		ItemName:            strings.TrimSpace(input.ItemName),
		IssueQty:            issueQty,
		UOMCode:             uomCode,
		BaseIssueQty:        baseIssueQty,
		BaseUOMCode:         baseUOMCode,
		ConversionFactor:    conversionFactor,
		BatchID:             strings.TrimSpace(input.BatchID),
		BatchNo:             strings.ToUpper(strings.TrimSpace(input.BatchNo)),
		LotNo:               strings.ToUpper(strings.TrimSpace(input.LotNo)),
		SourceBinID:         strings.TrimSpace(input.SourceBinID),
		LotTraceRequired:    input.LotTraceRequired,
		Note:                strings.TrimSpace(input.Note),
	}
	if err := line.Validate(); err != nil {
		return SubcontractMaterialTransferLine{}, err
	}

	return line, nil
}

func NewSubcontractMaterialTransferEvidence(
	input NewSubcontractMaterialTransferEvidenceInput,
) (SubcontractMaterialTransferEvidence, error) {
	evidence := SubcontractMaterialTransferEvidence{
		ID:           strings.TrimSpace(input.ID),
		EvidenceType: strings.ToLower(strings.TrimSpace(input.EvidenceType)),
		FileName:     strings.TrimSpace(input.FileName),
		ObjectKey:    strings.TrimSpace(input.ObjectKey),
		ExternalURL:  strings.TrimSpace(input.ExternalURL),
		Note:         strings.TrimSpace(input.Note),
	}
	if err := evidence.Validate(); err != nil {
		return SubcontractMaterialTransferEvidence{}, err
	}

	return evidence, nil
}

func (t SubcontractMaterialTransfer) Validate() error {
	if strings.TrimSpace(t.ID) == "" ||
		strings.TrimSpace(t.OrgID) == "" ||
		strings.TrimSpace(t.TransferNo) == "" ||
		strings.TrimSpace(t.SubcontractOrderID) == "" ||
		strings.TrimSpace(t.SubcontractOrderNo) == "" ||
		strings.TrimSpace(t.FactoryID) == "" ||
		strings.TrimSpace(t.FactoryName) == "" ||
		strings.TrimSpace(t.SourceWarehouseID) == "" ||
		strings.TrimSpace(t.HandoverBy) == "" ||
		strings.TrimSpace(t.ReceivedBy) == "" ||
		strings.TrimSpace(t.CreatedBy) == "" ||
		len(t.Lines) == 0 {
		return ErrSubcontractMaterialTransferRequiredField
	}
	if !IsValidSubcontractMaterialTransferStatus(t.Status) {
		return ErrSubcontractMaterialTransferInvalidStatus
	}
	seenLineNo := map[int]struct{}{}
	for _, line := range t.Lines {
		if _, exists := seenLineNo[line.LineNo]; exists {
			return ErrSubcontractMaterialTransferRequiredField
		}
		seenLineNo[line.LineNo] = struct{}{}
		if err := line.Validate(); err != nil {
			return err
		}
	}
	for _, evidence := range t.Evidence {
		if err := evidence.Validate(); err != nil {
			return err
		}
	}

	return nil
}

func (l SubcontractMaterialTransferLine) Validate() error {
	if strings.TrimSpace(l.ID) == "" ||
		l.LineNo <= 0 ||
		strings.TrimSpace(l.OrderMaterialLineID) == "" ||
		strings.TrimSpace(l.ItemID) == "" ||
		strings.TrimSpace(l.SKUCode) == "" ||
		strings.TrimSpace(l.ItemName) == "" {
		return ErrSubcontractMaterialTransferRequiredField
	}
	for _, quantity := range []decimal.Decimal{l.IssueQty, l.BaseIssueQty, l.ConversionFactor} {
		value, err := decimal.ParseQuantity(quantity.String())
		if err != nil || value.IsNegative() || value.IsZero() {
			return ErrSubcontractMaterialTransferInvalidQuantity
		}
	}
	if _, err := decimal.NormalizeUOMCode(l.UOMCode.String()); err != nil {
		return ErrSubcontractMaterialTransferInvalidQuantity
	}
	if _, err := decimal.NormalizeUOMCode(l.BaseUOMCode.String()); err != nil {
		return ErrSubcontractMaterialTransferInvalidQuantity
	}
	if l.LotTraceRequired && strings.TrimSpace(l.BatchID) == "" && strings.TrimSpace(l.BatchNo) == "" && strings.TrimSpace(l.LotNo) == "" {
		return ErrSubcontractMaterialTransferBatchRequired
	}

	return nil
}

func (e SubcontractMaterialTransferEvidence) Validate() error {
	if strings.TrimSpace(e.ID) == "" ||
		strings.TrimSpace(e.EvidenceType) == "" ||
		(strings.TrimSpace(e.ObjectKey) == "" && strings.TrimSpace(e.ExternalURL) == "") {
		return ErrSubcontractMaterialTransferRequiredField
	}

	return nil
}

func (t SubcontractMaterialTransfer) Clone() SubcontractMaterialTransfer {
	clone := t
	clone.Lines = append([]SubcontractMaterialTransferLine(nil), t.Lines...)
	clone.Evidence = append([]SubcontractMaterialTransferEvidence(nil), t.Evidence...)

	return clone
}

func NormalizeSubcontractMaterialTransferStatus(status SubcontractMaterialTransferStatus) SubcontractMaterialTransferStatus {
	return SubcontractMaterialTransferStatus(strings.ToLower(strings.TrimSpace(string(status))))
}

func IsValidSubcontractMaterialTransferStatus(status SubcontractMaterialTransferStatus) bool {
	switch NormalizeSubcontractMaterialTransferStatus(status) {
	case SubcontractMaterialTransferStatusSentToFactory,
		SubcontractMaterialTransferStatusPartiallySent:
		return true
	default:
		return false
	}
}
