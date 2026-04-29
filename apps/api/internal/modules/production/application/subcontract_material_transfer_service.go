package application

import (
	"context"
	"fmt"
	"strings"
	"time"

	inventorydomain "github.com/Chinsusu/ERP-v2/apps/api/internal/modules/inventory/domain"
	productiondomain "github.com/Chinsusu/ERP-v2/apps/api/internal/modules/production/domain"
	"github.com/Chinsusu/ERP-v2/apps/api/internal/shared/decimal"
)

type SubcontractMaterialTransferService struct {
	clock func() time.Time
}

type BuildSubcontractMaterialTransferInput struct {
	ID                  string
	TransferNo          string
	Order               productiondomain.SubcontractOrder
	SourceWarehouseID   string
	SourceWarehouseCode string
	Lines               []BuildSubcontractMaterialTransferLineInput
	Evidence            []BuildSubcontractMaterialTransferEvidenceInput
	HandoverBy          string
	HandoverAt          time.Time
	ReceivedBy          string
	ReceiverContact     string
	VehicleNo           string
	Note                string
	ActorID             string
}

type BuildSubcontractMaterialTransferLineInput struct {
	ID                  string
	LineNo              int
	OrderMaterialLineID string
	IssueQty            string
	UOMCode             string
	BaseIssueQty        string
	BaseUOMCode         string
	ConversionFactor    string
	BatchID             string
	BatchNo             string
	LotNo               string
	SourceBinID         string
	Note                string
}

type BuildSubcontractMaterialTransferEvidenceInput struct {
	ID           string
	EvidenceType string
	FileName     string
	ObjectKey    string
	ExternalURL  string
	Note         string
}

type SubcontractMaterialTransferBuildResult struct {
	Transfer       productiondomain.SubcontractMaterialTransfer
	UpdatedOrder   productiondomain.SubcontractOrder
	StockMovements []inventorydomain.StockMovement
}

func NewSubcontractMaterialTransferService() SubcontractMaterialTransferService {
	return SubcontractMaterialTransferService{clock: func() time.Time { return time.Now().UTC() }}
}

func (s SubcontractMaterialTransferService) BuildIssue(
	ctx context.Context,
	input BuildSubcontractMaterialTransferInput,
) (SubcontractMaterialTransferBuildResult, error) {
	_ = ctx
	if strings.TrimSpace(input.ActorID) == "" {
		return SubcontractMaterialTransferBuildResult{}, productiondomain.ErrSubcontractOrderTransitionActorRequired
	}
	if strings.TrimSpace(input.SourceWarehouseID) == "" {
		return SubcontractMaterialTransferBuildResult{}, productiondomain.ErrSubcontractMaterialTransferRequiredField
	}
	if len(input.Lines) == 0 {
		return SubcontractMaterialTransferBuildResult{}, productiondomain.ErrSubcontractMaterialTransferRequiredField
	}

	now := s.now()
	id := strings.TrimSpace(input.ID)
	if id == "" {
		id = newSubcontractMaterialTransferID(now)
	}
	transferNo := strings.TrimSpace(input.TransferNo)
	if transferNo == "" {
		transferNo = newSubcontractMaterialTransferNo(now)
	}
	handoverAt := input.HandoverAt
	if handoverAt.IsZero() {
		handoverAt = now
	}

	transferLines, issueLines, err := buildSubcontractMaterialTransferLines(id, input.Order, input.Lines)
	if err != nil {
		return SubcontractMaterialTransferBuildResult{}, err
	}
	updatedOrder, err := input.Order.IssueMaterials(productiondomain.IssueSubcontractMaterialsInput{
		Lines:     issueLines,
		ActorID:   input.ActorID,
		ChangedAt: handoverAt,
	})
	if err != nil {
		return SubcontractMaterialTransferBuildResult{}, err
	}

	status := productiondomain.SubcontractMaterialTransferStatusPartiallySent
	if updatedOrder.Status == productiondomain.SubcontractOrderStatusMaterialsIssued {
		status = productiondomain.SubcontractMaterialTransferStatusSentToFactory
	}
	transfer, err := productiondomain.NewSubcontractMaterialTransfer(productiondomain.NewSubcontractMaterialTransferInput{
		ID:                  id,
		OrgID:               input.Order.OrgID,
		TransferNo:          transferNo,
		SubcontractOrderID:  input.Order.ID,
		SubcontractOrderNo:  input.Order.OrderNo,
		FactoryID:           input.Order.FactoryID,
		FactoryCode:         input.Order.FactoryCode,
		FactoryName:         input.Order.FactoryName,
		SourceWarehouseID:   input.SourceWarehouseID,
		SourceWarehouseCode: input.SourceWarehouseCode,
		Status:              status,
		Lines:               transferLines,
		Evidence:            subcontractMaterialTransferEvidenceInputs(input.Evidence),
		HandoverBy:          firstNonBlankSubcontractOrder(input.HandoverBy, input.ActorID),
		HandoverAt:          handoverAt,
		ReceivedBy:          input.ReceivedBy,
		ReceiverContact:     input.ReceiverContact,
		VehicleNo:           input.VehicleNo,
		Note:                input.Note,
		CreatedAt:           now,
		CreatedBy:           input.ActorID,
		UpdatedAt:           now,
		UpdatedBy:           input.ActorID,
	})
	if err != nil {
		return SubcontractMaterialTransferBuildResult{}, err
	}

	movements, err := buildSubcontractIssueMovements(transfer, input.ActorID)
	if err != nil {
		return SubcontractMaterialTransferBuildResult{}, err
	}

	return SubcontractMaterialTransferBuildResult{
		Transfer:       transfer,
		UpdatedOrder:   updatedOrder,
		StockMovements: movements,
	}, nil
}

func (s SubcontractMaterialTransferService) now() time.Time {
	if s.clock == nil {
		return time.Now().UTC()
	}

	return s.clock().UTC()
}

func buildSubcontractMaterialTransferLines(
	transferID string,
	order productiondomain.SubcontractOrder,
	inputs []BuildSubcontractMaterialTransferLineInput,
) (
	[]productiondomain.NewSubcontractMaterialTransferLineInput,
	[]productiondomain.IssueSubcontractMaterialLineInput,
	error,
) {
	transferLines := make([]productiondomain.NewSubcontractMaterialTransferLineInput, 0, len(inputs))
	issueLines := make([]productiondomain.IssueSubcontractMaterialLineInput, 0, len(inputs))
	for index, input := range inputs {
		orderLine, ok := subcontractMaterialLineByID(order, input.OrderMaterialLineID)
		if !ok {
			return nil, nil, productiondomain.ErrSubcontractOrderMaterialLineNotFound
		}
		lineNo := input.LineNo
		if lineNo == 0 {
			lineNo = index + 1
		}
		lineID := strings.TrimSpace(input.ID)
		if lineID == "" {
			lineID = fmt.Sprintf("%s-line-%02d", strings.TrimSpace(transferID), lineNo)
		}
		remainingQty, err := decimal.SubtractQuantity(orderLine.PlannedQty, orderLine.IssuedQty)
		if err != nil {
			return nil, nil, productiondomain.ErrSubcontractMaterialTransferInvalidQuantity
		}
		issueQty, err := decimal.ParseQuantity(firstNonBlankSubcontractOrder(input.IssueQty, remainingQty.String()))
		if err != nil || issueQty.IsNegative() || issueQty.IsZero() {
			return nil, nil, productiondomain.ErrSubcontractMaterialTransferInvalidQuantity
		}
		uomCode := firstNonBlankSubcontractOrder(input.UOMCode, orderLine.UOMCode.String())
		baseUOMCode := firstNonBlankSubcontractOrder(input.BaseUOMCode, orderLine.BaseUOMCode.String())
		conversionFactor, err := decimal.ParseQuantity(firstNonBlankSubcontractOrder(input.ConversionFactor, orderLine.ConversionFactor.String()))
		if err != nil || conversionFactor.IsNegative() || conversionFactor.IsZero() {
			return nil, nil, productiondomain.ErrSubcontractMaterialTransferInvalidQuantity
		}
		baseIssueQty := decimal.Decimal("")
		if strings.TrimSpace(input.BaseIssueQty) != "" {
			baseIssueQty, err = decimal.ParseQuantity(input.BaseIssueQty)
			if err != nil || baseIssueQty.IsNegative() || baseIssueQty.IsZero() {
				return nil, nil, productiondomain.ErrSubcontractMaterialTransferInvalidQuantity
			}
		} else if strings.EqualFold(strings.TrimSpace(uomCode), strings.TrimSpace(baseUOMCode)) {
			baseIssueQty = issueQty
		} else {
			baseIssueQty, err = decimal.MultiplyQuantityByFactor(issueQty, conversionFactor)
			if err != nil {
				return nil, nil, productiondomain.ErrSubcontractMaterialTransferInvalidQuantity
			}
		}

		transferLines = append(transferLines, productiondomain.NewSubcontractMaterialTransferLineInput{
			ID:                  lineID,
			LineNo:              lineNo,
			OrderMaterialLineID: orderLine.ID,
			ItemID:              orderLine.ItemID,
			SKUCode:             orderLine.SKUCode,
			ItemName:            orderLine.ItemName,
			IssueQty:            issueQty,
			UOMCode:             uomCode,
			BaseIssueQty:        baseIssueQty,
			BaseUOMCode:         baseUOMCode,
			ConversionFactor:    conversionFactor,
			BatchID:             input.BatchID,
			BatchNo:             input.BatchNo,
			LotNo:               input.LotNo,
			SourceBinID:         input.SourceBinID,
			LotTraceRequired:    orderLine.LotTraceRequired,
			Note:                input.Note,
		})
		issueLines = append(issueLines, productiondomain.IssueSubcontractMaterialLineInput{
			OrderMaterialLineID: orderLine.ID,
			IssueQty:            issueQty,
			UOMCode:             uomCode,
			BaseIssueQty:        baseIssueQty,
			BaseUOMCode:         baseUOMCode,
			ConversionFactor:    conversionFactor,
		})
	}

	return transferLines, issueLines, nil
}

func buildSubcontractIssueMovements(
	transfer productiondomain.SubcontractMaterialTransfer,
	actorID string,
) ([]inventorydomain.StockMovement, error) {
	movements := make([]inventorydomain.StockMovement, 0, len(transfer.Lines))
	for _, line := range transfer.Lines {
		movement, err := inventorydomain.NewStockMovement(inventorydomain.NewStockMovementInput{
			MovementNo:       fmt.Sprintf("%s-MOV-%02d", transfer.TransferNo, line.LineNo),
			MovementType:     inventorydomain.MovementSubcontractIssue,
			OrgID:            transfer.OrgID,
			ItemID:           line.ItemID,
			BatchID:          firstNonBlankSubcontractOrder(line.BatchID, line.BatchNo, line.LotNo),
			WarehouseID:      transfer.SourceWarehouseID,
			BinID:            line.SourceBinID,
			Quantity:         line.BaseIssueQty,
			BaseUOMCode:      line.BaseUOMCode.String(),
			SourceQuantity:   line.IssueQty,
			SourceUOMCode:    line.UOMCode.String(),
			ConversionFactor: line.ConversionFactor,
			StockStatus:      inventorydomain.StockStatusSubcontractIssued,
			SourceDocType:    "subcontract_material_transfer",
			SourceDocID:      transfer.ID,
			SourceDocLineID:  line.ID,
			Reason:           "subcontract material issued to factory",
			CreatedBy:        actorID,
			MovementAt:       transfer.HandoverAt,
		})
		if err != nil {
			return nil, err
		}
		movements = append(movements, movement)
	}

	return movements, nil
}

func subcontractMaterialLineByID(order productiondomain.SubcontractOrder, lineID string) (productiondomain.SubcontractMaterialLine, bool) {
	lineID = strings.TrimSpace(lineID)
	if lineID == "" {
		return productiondomain.SubcontractMaterialLine{}, false
	}
	for _, line := range order.MaterialLines {
		if line.ID == lineID {
			return line, true
		}
	}

	return productiondomain.SubcontractMaterialLine{}, false
}

func subcontractMaterialTransferEvidenceInputs(
	inputs []BuildSubcontractMaterialTransferEvidenceInput,
) []productiondomain.NewSubcontractMaterialTransferEvidenceInput {
	evidence := make([]productiondomain.NewSubcontractMaterialTransferEvidenceInput, 0, len(inputs))
	for _, input := range inputs {
		evidence = append(evidence, productiondomain.NewSubcontractMaterialTransferEvidenceInput{
			ID:           input.ID,
			EvidenceType: input.EvidenceType,
			FileName:     input.FileName,
			ObjectKey:    input.ObjectKey,
			ExternalURL:  input.ExternalURL,
			Note:         input.Note,
		})
	}

	return evidence
}

func newSubcontractMaterialTransferID(now time.Time) string {
	return fmt.Sprintf("smt-%s-%06d", now.UTC().Format("060102"), now.UTC().UnixNano()%1000000)
}

func newSubcontractMaterialTransferNo(now time.Time) string {
	return fmt.Sprintf("SMT-%s-%06d", now.UTC().Format("060102"), now.UTC().UnixNano()%1000000)
}
