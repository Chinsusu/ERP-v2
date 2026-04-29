package application

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	inventorydomain "github.com/Chinsusu/ERP-v2/apps/api/internal/modules/inventory/domain"
	productiondomain "github.com/Chinsusu/ERP-v2/apps/api/internal/modules/production/domain"
	"github.com/Chinsusu/ERP-v2/apps/api/internal/shared/decimal"
)

const subcontractFinishedGoodsReceiptSourceDoc = "subcontract_finished_goods_receipt"

type SubcontractFinishedGoodsReceiptStore interface {
	Save(ctx context.Context, receipt productiondomain.SubcontractFinishedGoodsReceipt) error
	ListBySubcontractOrder(ctx context.Context, subcontractOrderID string) ([]productiondomain.SubcontractFinishedGoodsReceipt, error)
}

type PrototypeSubcontractFinishedGoodsReceiptStore struct {
	mu      sync.RWMutex
	records map[string]productiondomain.SubcontractFinishedGoodsReceipt
}

type SubcontractFinishedGoodsReceiptService struct {
	clock func() time.Time
}

type BuildSubcontractFinishedGoodsReceiptInput struct {
	ID             string
	ReceiptNo      string
	Order          productiondomain.SubcontractOrder
	WarehouseID    string
	WarehouseCode  string
	LocationID     string
	LocationCode   string
	DeliveryNoteNo string
	Lines          []BuildSubcontractFinishedGoodsReceiptLineInput
	Evidence       []BuildSubcontractFinishedGoodsReceiptEvidenceInput
	ReceivedBy     string
	ReceivedAt     time.Time
	Note           string
	ActorID        string
}

type BuildSubcontractFinishedGoodsReceiptLineInput struct {
	ID               string
	LineNo           int
	ItemID           string
	SKUCode          string
	ItemName         string
	BatchID          string
	BatchNo          string
	LotNo            string
	ExpiryDate       string
	ReceiveQty       string
	UOMCode          string
	BaseReceiveQty   string
	BaseUOMCode      string
	ConversionFactor string
	PackagingStatus  string
	Note             string
}

type BuildSubcontractFinishedGoodsReceiptEvidenceInput struct {
	ID           string
	EvidenceType string
	FileName     string
	ObjectKey    string
	ExternalURL  string
	Note         string
}

type SubcontractFinishedGoodsReceiptBuildResult struct {
	Receipt        productiondomain.SubcontractFinishedGoodsReceipt
	UpdatedOrder   productiondomain.SubcontractOrder
	StockMovements []inventorydomain.StockMovement
}

func NewSubcontractFinishedGoodsReceiptService() SubcontractFinishedGoodsReceiptService {
	return SubcontractFinishedGoodsReceiptService{clock: func() time.Time { return time.Now().UTC() }}
}

func NewPrototypeSubcontractFinishedGoodsReceiptStore() *PrototypeSubcontractFinishedGoodsReceiptStore {
	return &PrototypeSubcontractFinishedGoodsReceiptStore{records: make(map[string]productiondomain.SubcontractFinishedGoodsReceipt)}
}

func (s SubcontractFinishedGoodsReceiptService) BuildReceipt(
	ctx context.Context,
	input BuildSubcontractFinishedGoodsReceiptInput,
) (SubcontractFinishedGoodsReceiptBuildResult, error) {
	_ = ctx
	actorID := strings.TrimSpace(input.ActorID)
	if actorID == "" {
		return SubcontractFinishedGoodsReceiptBuildResult{}, productiondomain.ErrSubcontractOrderTransitionActorRequired
	}
	if strings.TrimSpace(input.WarehouseID) == "" || strings.TrimSpace(input.LocationID) == "" || len(input.Lines) == 0 {
		return SubcontractFinishedGoodsReceiptBuildResult{}, productiondomain.ErrSubcontractFinishedGoodsReceiptRequiredField
	}

	now := s.now()
	id := strings.TrimSpace(input.ID)
	if id == "" {
		id = newSubcontractFinishedGoodsReceiptID(now)
	}
	receiptNo := strings.TrimSpace(input.ReceiptNo)
	if receiptNo == "" {
		receiptNo = newSubcontractFinishedGoodsReceiptNo(now)
	}
	receivedAt := input.ReceivedAt
	if receivedAt.IsZero() {
		receivedAt = now
	}

	receipt, err := productiondomain.NewSubcontractFinishedGoodsReceipt(productiondomain.NewSubcontractFinishedGoodsReceiptInput{
		ID:                 id,
		OrgID:              input.Order.OrgID,
		ReceiptNo:          receiptNo,
		SubcontractOrderID: input.Order.ID,
		SubcontractOrderNo: input.Order.OrderNo,
		FactoryID:          input.Order.FactoryID,
		FactoryCode:        input.Order.FactoryCode,
		FactoryName:        input.Order.FactoryName,
		WarehouseID:        input.WarehouseID,
		WarehouseCode:      input.WarehouseCode,
		LocationID:         input.LocationID,
		LocationCode:       input.LocationCode,
		DeliveryNoteNo:     input.DeliveryNoteNo,
		Status:             productiondomain.SubcontractFinishedGoodsReceiptStatusQCHold,
		Lines:              subcontractFinishedGoodsReceiptLineInputs(id, input.Order, input.Lines),
		Evidence:           subcontractFinishedGoodsReceiptEvidenceInputs(id, input.Evidence),
		ReceivedBy:         firstNonBlankSubcontractOrder(input.ReceivedBy, actorID),
		ReceivedAt:         receivedAt,
		Note:               input.Note,
		CreatedAt:          now,
		CreatedBy:          actorID,
		UpdatedAt:          now,
		UpdatedBy:          actorID,
	})
	if err != nil {
		return SubcontractFinishedGoodsReceiptBuildResult{}, err
	}
	totalReceiptQty, totalBaseReceiptQty, err := validateSubcontractFinishedGoodsReceiptLines(receipt, input.Order)
	if err != nil {
		return SubcontractFinishedGoodsReceiptBuildResult{}, err
	}
	updatedOrder, err := input.Order.ReceiveFinishedGoods(productiondomain.ReceiveSubcontractFinishedGoodsInput{
		ReceiptQty:       totalReceiptQty,
		UOMCode:          input.Order.UOMCode.String(),
		BaseReceiptQty:   totalBaseReceiptQty,
		BaseUOMCode:      input.Order.BaseUOMCode.String(),
		ConversionFactor: input.Order.ConversionFactor,
		ActorID:          actorID,
		ChangedAt:        receivedAt,
	})
	if err != nil {
		return SubcontractFinishedGoodsReceiptBuildResult{}, err
	}
	movements, err := buildSubcontractFinishedGoodsReceiptMovements(receipt, actorID)
	if err != nil {
		return SubcontractFinishedGoodsReceiptBuildResult{}, err
	}

	return SubcontractFinishedGoodsReceiptBuildResult{
		Receipt:        receipt,
		UpdatedOrder:   updatedOrder,
		StockMovements: movements,
	}, nil
}

func (s SubcontractFinishedGoodsReceiptService) now() time.Time {
	if s.clock == nil {
		return time.Now().UTC()
	}

	return s.clock().UTC()
}

func subcontractFinishedGoodsReceiptLineInputs(
	receiptID string,
	order productiondomain.SubcontractOrder,
	inputs []BuildSubcontractFinishedGoodsReceiptLineInput,
) []productiondomain.NewSubcontractFinishedGoodsReceiptLineInput {
	lines := make([]productiondomain.NewSubcontractFinishedGoodsReceiptLineInput, 0, len(inputs))
	for index, input := range inputs {
		lineNo := input.LineNo
		if lineNo == 0 {
			lineNo = index + 1
		}
		lineID := strings.TrimSpace(input.ID)
		if lineID == "" {
			lineID = fmt.Sprintf("%s-line-%02d", strings.TrimSpace(receiptID), lineNo)
		}
		lines = append(lines, productiondomain.NewSubcontractFinishedGoodsReceiptLineInput{
			ID:               lineID,
			LineNo:           lineNo,
			ItemID:           firstNonBlankSubcontractOrder(input.ItemID, order.FinishedItemID),
			SKUCode:          firstNonBlankSubcontractOrder(input.SKUCode, order.FinishedSKUCode),
			ItemName:         firstNonBlankSubcontractOrder(input.ItemName, order.FinishedItemName),
			BatchID:          input.BatchID,
			BatchNo:          input.BatchNo,
			LotNo:            input.LotNo,
			ExpiryDate:       parseSubcontractFinishedGoodsReceiptDate(input.ExpiryDate),
			ReceiveQty:       decimal.Decimal(strings.TrimSpace(input.ReceiveQty)),
			UOMCode:          firstNonBlankSubcontractOrder(input.UOMCode, order.UOMCode.String()),
			BaseReceiveQty:   decimal.Decimal(strings.TrimSpace(input.BaseReceiveQty)),
			BaseUOMCode:      firstNonBlankSubcontractOrder(input.BaseUOMCode, order.BaseUOMCode.String()),
			ConversionFactor: decimal.Decimal(strings.TrimSpace(firstNonBlankSubcontractOrder(input.ConversionFactor, order.ConversionFactor.String()))),
			PackagingStatus:  input.PackagingStatus,
			Note:             input.Note,
		})
	}

	return lines
}

func subcontractFinishedGoodsReceiptEvidenceInputs(
	receiptID string,
	inputs []BuildSubcontractFinishedGoodsReceiptEvidenceInput,
) []productiondomain.NewSubcontractFinishedGoodsReceiptEvidenceInput {
	evidence := make([]productiondomain.NewSubcontractFinishedGoodsReceiptEvidenceInput, 0, len(inputs))
	for index, input := range inputs {
		id := strings.TrimSpace(input.ID)
		if id == "" {
			id = fmt.Sprintf("%s-evidence-%02d", strings.TrimSpace(receiptID), index+1)
		}
		evidence = append(evidence, productiondomain.NewSubcontractFinishedGoodsReceiptEvidenceInput{
			ID:           id,
			EvidenceType: input.EvidenceType,
			FileName:     input.FileName,
			ObjectKey:    input.ObjectKey,
			ExternalURL:  input.ExternalURL,
			Note:         input.Note,
		})
	}

	return evidence
}

func validateSubcontractFinishedGoodsReceiptLines(
	receipt productiondomain.SubcontractFinishedGoodsReceipt,
	order productiondomain.SubcontractOrder,
) (decimal.Decimal, decimal.Decimal, error) {
	totalReceiptQty := decimal.MustQuantity("0")
	totalBaseReceiptQty := decimal.MustQuantity("0")
	for _, line := range receipt.Lines {
		if line.ItemID != order.FinishedItemID ||
			line.SKUCode != order.FinishedSKUCode ||
			line.UOMCode != order.UOMCode ||
			line.BaseUOMCode != order.BaseUOMCode ||
			line.ConversionFactor != order.ConversionFactor {
			return "", "", productiondomain.ErrSubcontractFinishedGoodsReceiptInvalidQuantity
		}
		var err error
		totalReceiptQty, err = decimal.AddQuantity(totalReceiptQty, line.ReceiveQty)
		if err != nil {
			return "", "", productiondomain.ErrSubcontractFinishedGoodsReceiptInvalidQuantity
		}
		totalBaseReceiptQty, err = decimal.AddQuantity(totalBaseReceiptQty, line.BaseReceiveQty)
		if err != nil {
			return "", "", productiondomain.ErrSubcontractFinishedGoodsReceiptInvalidQuantity
		}
	}

	return totalReceiptQty, totalBaseReceiptQty, nil
}

func buildSubcontractFinishedGoodsReceiptMovements(
	receipt productiondomain.SubcontractFinishedGoodsReceipt,
	actorID string,
) ([]inventorydomain.StockMovement, error) {
	movements := make([]inventorydomain.StockMovement, 0, len(receipt.Lines))
	for _, line := range receipt.Lines {
		movement, err := inventorydomain.NewStockMovement(inventorydomain.NewStockMovementInput{
			MovementNo:       fmt.Sprintf("%s-MOV-%02d", receipt.ReceiptNo, line.LineNo),
			MovementType:     inventorydomain.MovementSubcontractReceipt,
			OrgID:            receipt.OrgID,
			ItemID:           line.ItemID,
			BatchID:          line.BatchID,
			WarehouseID:      receipt.WarehouseID,
			BinID:            receipt.LocationID,
			Quantity:         line.BaseReceiveQty,
			BaseUOMCode:      line.BaseUOMCode.String(),
			SourceQuantity:   line.ReceiveQty,
			SourceUOMCode:    line.UOMCode.String(),
			ConversionFactor: line.ConversionFactor,
			StockStatus:      inventorydomain.StockStatusQCHold,
			SourceDocType:    subcontractFinishedGoodsReceiptSourceDoc,
			SourceDocID:      receipt.ID,
			SourceDocLineID:  line.ID,
			Reason:           "subcontract finished goods received into qc hold",
			CreatedBy:        actorID,
			MovementAt:       receipt.ReceivedAt,
		})
		if err != nil {
			return nil, err
		}
		movements = append(movements, movement)
	}

	return movements, nil
}

func buildSubcontractFinishedGoodsAcceptanceMovements(
	receipts []productiondomain.SubcontractFinishedGoodsReceipt,
	actorID string,
	acceptedAt time.Time,
) ([]inventorydomain.StockMovement, error) {
	lineCount := 0
	for _, receipt := range receipts {
		lineCount += len(receipt.Lines)
	}
	movements := make([]inventorydomain.StockMovement, 0, lineCount)
	for _, receipt := range receipts {
		for _, line := range receipt.Lines {
			movement, err := newSubcontractFinishedGoodsAcceptanceMovement(receipt, line, line.ReceiveQty, line.BaseReceiveQty, actorID, acceptedAt)
			if err != nil {
				return nil, err
			}
			movements = append(movements, movement)
		}
	}

	return movements, nil
}

func buildSubcontractFinishedGoodsAcceptanceMovementsForQuantity(
	receipts []productiondomain.SubcontractFinishedGoodsReceipt,
	acceptedQty decimal.Decimal,
	baseAcceptedQty decimal.Decimal,
	actorID string,
	acceptedAt time.Time,
) ([]inventorydomain.StockMovement, error) {
	remainingQty := acceptedQty
	remainingBaseQty := baseAcceptedQty
	movements := make([]inventorydomain.StockMovement, 0, len(receipts))
	for _, receipt := range receipts {
		for _, line := range receipt.Lines {
			if remainingBaseQty.IsZero() {
				break
			}
			releaseQty := line.ReceiveQty
			releaseBaseQty := line.BaseReceiveQty
			baseDiff, err := decimal.SubtractQuantity(remainingBaseQty, line.BaseReceiveQty)
			if err != nil {
				return nil, productiondomain.ErrSubcontractFinishedGoodsReceiptInvalidQuantity
			}
			if baseDiff.IsNegative() {
				releaseQty = remainingQty
				releaseBaseQty = remainingBaseQty
			}
			movement, err := newSubcontractFinishedGoodsAcceptanceMovement(receipt, line, releaseQty, releaseBaseQty, actorID, acceptedAt)
			if err != nil {
				return nil, err
			}
			movements = append(movements, movement)
			remainingQty, err = decimal.SubtractQuantity(remainingQty, releaseQty)
			if err != nil || remainingQty.IsNegative() {
				return nil, productiondomain.ErrSubcontractFinishedGoodsReceiptInvalidQuantity
			}
			remainingBaseQty, err = decimal.SubtractQuantity(remainingBaseQty, releaseBaseQty)
			if err != nil || remainingBaseQty.IsNegative() {
				return nil, productiondomain.ErrSubcontractFinishedGoodsReceiptInvalidQuantity
			}
		}
	}
	if !remainingQty.IsZero() || !remainingBaseQty.IsZero() {
		return nil, productiondomain.ErrSubcontractFinishedGoodsReceiptInvalidQuantity
	}

	return movements, nil
}

func newSubcontractFinishedGoodsAcceptanceMovement(
	receipt productiondomain.SubcontractFinishedGoodsReceipt,
	line productiondomain.SubcontractFinishedGoodsReceiptLine,
	releaseQty decimal.Decimal,
	releaseBaseQty decimal.Decimal,
	actorID string,
	acceptedAt time.Time,
) (inventorydomain.StockMovement, error) {
	return inventorydomain.NewStockMovement(inventorydomain.NewStockMovementInput{
		MovementNo:       fmt.Sprintf("%s-ACCEPT-MOV-%02d", receipt.ReceiptNo, line.LineNo),
		MovementType:     inventorydomain.MovementQCRelease,
		OrgID:            receipt.OrgID,
		ItemID:           line.ItemID,
		BatchID:          line.BatchID,
		WarehouseID:      receipt.WarehouseID,
		BinID:            receipt.LocationID,
		Quantity:         releaseBaseQty,
		BaseUOMCode:      line.BaseUOMCode.String(),
		SourceQuantity:   releaseQty,
		SourceUOMCode:    line.UOMCode.String(),
		ConversionFactor: line.ConversionFactor,
		StockStatus:      inventorydomain.StockStatusAvailable,
		SourceDocType:    subcontractFinishedGoodsReceiptSourceDoc,
		SourceDocID:      receipt.ID,
		SourceDocLineID:  line.ID,
		Reason:           "subcontract finished goods accepted from qc hold",
		CreatedBy:        actorID,
		MovementAt:       acceptedAt,
	})
}

func parseSubcontractFinishedGoodsReceiptDate(value string) time.Time {
	value = strings.TrimSpace(value)
	if value == "" {
		return time.Time{}
	}
	parsed, err := time.Parse("2006-01-02", value)
	if err != nil {
		return time.Time{}
	}

	return parsed
}

func newSubcontractFinishedGoodsReceiptID(now time.Time) string {
	return fmt.Sprintf("sfgr-%s-%06d", now.UTC().Format("060102"), now.UTC().UnixNano()%1000000)
}

func newSubcontractFinishedGoodsReceiptNo(now time.Time) string {
	return fmt.Sprintf("SFGR-%s-%06d", now.UTC().Format("060102"), now.UTC().UnixNano()%1000000)
}

func (s *PrototypeSubcontractFinishedGoodsReceiptStore) Save(
	_ context.Context,
	receipt productiondomain.SubcontractFinishedGoodsReceipt,
) error {
	if s == nil {
		return fmt.Errorf("subcontract finished goods receipt store is required")
	}
	if err := receipt.Validate(); err != nil {
		return err
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	s.records[receipt.ID] = receipt.Clone()

	return nil
}

func (s *PrototypeSubcontractFinishedGoodsReceiptStore) ListBySubcontractOrder(
	_ context.Context,
	subcontractOrderID string,
) ([]productiondomain.SubcontractFinishedGoodsReceipt, error) {
	if s == nil {
		return nil, fmt.Errorf("subcontract finished goods receipt store is required")
	}

	s.mu.RLock()
	defer s.mu.RUnlock()
	receipts := make([]productiondomain.SubcontractFinishedGoodsReceipt, 0)
	for _, receipt := range s.records {
		if receipt.SubcontractOrderID == strings.TrimSpace(subcontractOrderID) {
			receipts = append(receipts, receipt.Clone())
		}
	}

	return receipts, nil
}

func (s *PrototypeSubcontractFinishedGoodsReceiptStore) Count() int {
	if s == nil {
		return 0
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	return len(s.records)
}
