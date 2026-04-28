package application

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	inventoryapp "github.com/Chinsusu/ERP-v2/apps/api/internal/modules/inventory/application"
	inventorydomain "github.com/Chinsusu/ERP-v2/apps/api/internal/modules/inventory/domain"
	"github.com/Chinsusu/ERP-v2/apps/api/internal/modules/returns/domain"
	"github.com/Chinsusu/ERP-v2/apps/api/internal/shared/audit"
	"github.com/Chinsusu/ERP-v2/apps/api/internal/shared/decimal"
)

var ErrReturnReceiptDispositionNotAllowed = errors.New("return receipt disposition action is not allowed")

const returnStockSourceDocType = "return_receipt"

type ReturnDispositionStore interface {
	FindReceiptByID(ctx context.Context, id string) (domain.ReturnReceipt, error)
	SaveDisposition(ctx context.Context, receipt domain.ReturnReceipt, action domain.ReturnDispositionAction) error
}

type ApplyReturnDisposition struct {
	store         ReturnDispositionStore
	stockMovement inventoryapp.StockMovementStore
	auditLog      audit.LogStore
	clock         func() time.Time
}

type ApplyReturnDispositionInput struct {
	ReceiptID   string
	Disposition string
	Note        string
	ActorID     string
	RequestID   string
}

type ReturnDispositionResult struct {
	Action     domain.ReturnDispositionAction
	Receipt    domain.ReturnReceipt
	AuditLogID string
}

func NewApplyReturnDisposition(
	store ReturnDispositionStore,
	stockMovement inventoryapp.StockMovementStore,
	auditLog audit.LogStore,
) ApplyReturnDisposition {
	return ApplyReturnDisposition{
		store:         store,
		stockMovement: stockMovement,
		auditLog:      auditLog,
		clock:         func() time.Time { return time.Now().UTC() },
	}
}

func (uc ApplyReturnDisposition) Execute(
	ctx context.Context,
	input ApplyReturnDispositionInput,
) (ReturnDispositionResult, error) {
	if uc.store == nil {
		return ReturnDispositionResult{}, errors.New("return disposition store is required")
	}
	if uc.auditLog == nil {
		return ReturnDispositionResult{}, errors.New("audit log store is required")
	}

	receipt, err := uc.store.FindReceiptByID(ctx, input.ReceiptID)
	if err != nil {
		return ReturnDispositionResult{}, err
	}
	if receipt.Status != domain.ReturnStatusInspected {
		return ReturnDispositionResult{}, ErrReturnReceiptDispositionNotAllowed
	}

	action, err := domain.NewReturnDispositionAction(domain.NewReturnDispositionActionInput{
		ReceiptID:   receipt.ID,
		ReceiptNo:   receipt.ReceiptNo,
		Disposition: domain.ReturnDisposition(input.Disposition),
		ActorID:     input.ActorID,
		Note:        input.Note,
		DecidedAt:   uc.clock(),
	})
	if err != nil {
		return ReturnDispositionResult{}, err
	}

	updatedReceipt := receipt.ApplyDisposition(action)
	movements, err := uc.newReturnStockMovements(updatedReceipt, action, input.ActorID)
	if err != nil {
		return ReturnDispositionResult{}, err
	}
	for _, movement := range movements {
		if err := uc.stockMovement.Record(ctx, movement); err != nil {
			return ReturnDispositionResult{}, err
		}
	}
	if len(movements) > 0 {
		updatedReceipt.StockMovement = newReturnStockMovementSummary(updatedReceipt, movements[0])
	}
	if err := uc.store.SaveDisposition(ctx, updatedReceipt, action); err != nil {
		return ReturnDispositionResult{}, err
	}

	log, err := newReturnDispositionAuditLog(input.ActorID, input.RequestID, receipt, updatedReceipt, action)
	if err != nil {
		return ReturnDispositionResult{}, err
	}
	if err := uc.auditLog.Record(ctx, log); err != nil {
		return ReturnDispositionResult{}, err
	}

	return ReturnDispositionResult{
		Action:     action,
		Receipt:    updatedReceipt,
		AuditLogID: log.ID,
	}, nil
}

type returnStockMovementSpec struct {
	movementNoSuffix string
	movementType     inventorydomain.MovementType
	stockStatus      inventorydomain.StockStatus
	reason           string
}

func (uc ApplyReturnDisposition) newReturnStockMovements(
	receipt domain.ReturnReceipt,
	action domain.ReturnDispositionAction,
	actorID string,
) ([]inventorydomain.StockMovement, error) {
	spec, ok := returnStockMovementSpecForDisposition(action.Disposition)
	if !ok {
		return nil, nil
	}
	if uc.stockMovement == nil {
		return nil, errors.New("stock movement store is required")
	}

	movements := make([]inventorydomain.StockMovement, 0, len(receipt.Lines))
	for index, line := range receipt.Lines {
		quantity := line.Quantity
		if quantity <= 0 {
			quantity = 1
		}
		baseQuantity := decimal.MustQuantity(strconv.Itoa(quantity))
		movement, err := inventorydomain.NewStockMovement(inventorydomain.NewStockMovementInput{
			MovementNo:       fmt.Sprintf("%s-%s-%03d", receipt.ReceiptNo, spec.movementNoSuffix, index+1),
			MovementType:     spec.movementType,
			OrgID:            "org-my-pham",
			ItemID:           returnStockItemID(line),
			WarehouseID:      receipt.WarehouseID,
			Quantity:         baseQuantity,
			BaseUOMCode:      "EA",
			SourceQuantity:   baseQuantity,
			SourceUOMCode:    "EA",
			ConversionFactor: decimal.MustQuantity("1"),
			StockStatus:      spec.stockStatus,
			SourceDocType:    returnStockSourceDocType,
			SourceDocID:      receipt.ID,
			SourceDocLineID:  line.ID,
			Reason:           spec.reason,
			CreatedBy:        actorID,
			MovementAt:       action.DecidedAt,
		})
		if err != nil {
			return nil, err
		}
		movements = append(movements, movement)
	}

	return movements, nil
}

func returnStockMovementSpecForDisposition(
	disposition domain.ReturnDisposition,
) (returnStockMovementSpec, bool) {
	switch disposition {
	case domain.ReturnDispositionReusable:
		return returnStockMovementSpec{
			movementNoSuffix: "RESTOCK",
			movementType:     inventorydomain.MovementReturnRestock,
			stockStatus:      inventorydomain.StockStatusAvailable,
			reason:           "return reusable restock",
		}, true
	case domain.ReturnDispositionNeedsInspection:
		return returnStockMovementSpec{
			movementNoSuffix: "QCHOLD",
			movementType:     inventorydomain.MovementReturnReceipt,
			stockStatus:      inventorydomain.StockStatusQCHold,
			reason:           "return quarantine hold",
		}, true
	default:
		return returnStockMovementSpec{}, false
	}
}

func returnStockItemID(line domain.ReturnReceiptLine) string {
	sku := strings.ToLower(strings.TrimSpace(line.SKU))
	if sku == "" {
		return "item-unknown-sku"
	}

	return "item-" + sku
}

func newReturnStockMovementSummary(
	receipt domain.ReturnReceipt,
	movement inventorydomain.StockMovement,
) *domain.ReturnStockMovement {
	line := domain.ReturnReceiptLine{SKU: "UNKNOWN-SKU", Quantity: 1}
	if len(receipt.Lines) > 0 {
		line = receipt.Lines[0]
	}
	if line.Quantity <= 0 {
		line.Quantity = 1
	}

	return &domain.ReturnStockMovement{
		ID:                movement.MovementNo,
		MovementType:      string(movement.MovementType),
		SKU:               line.SKU,
		WarehouseID:       movement.WarehouseID,
		Quantity:          line.Quantity,
		TargetStockStatus: string(movement.StockStatus),
		SourceDocID:       movement.SourceDocID,
	}
}

func newReturnDispositionAuditLog(
	actorID string,
	requestID string,
	before domain.ReturnReceipt,
	after domain.ReturnReceipt,
	action domain.ReturnDispositionAction,
) (audit.Log, error) {
	afterData := map[string]any{
		"action_id":           action.ID,
		"status":              string(after.Status),
		"disposition":         string(action.Disposition),
		"target_location":     action.TargetLocation,
		"target_stock_status": action.TargetStockStatus,
		"action_code":         action.ActionCode,
	}
	if after.StockMovement != nil {
		afterData["stock_movement_id"] = after.StockMovement.ID
		afterData["stock_movement_type"] = after.StockMovement.MovementType
		afterData["stock_movement_status"] = after.StockMovement.TargetStockStatus
	}

	return audit.NewLog(audit.NewLogInput{
		OrgID:      "org-my-pham",
		ActorID:    strings.TrimSpace(actorID),
		Action:     "returns.inspection.disposition",
		EntityType: "returns.return_receipt",
		EntityID:   after.ID,
		RequestID:  strings.TrimSpace(requestID),
		BeforeData: map[string]any{
			"status":          string(before.Status),
			"disposition":     string(before.Disposition),
			"target_location": before.TargetLocation,
		},
		AfterData: afterData,
		Metadata: map[string]any{
			"source": "return disposition",
		},
		CreatedAt: action.DecidedAt,
	})
}

func NewPrototypeApplyReturnDispositionAt(
	store ReturnDispositionStore,
	auditLog audit.LogStore,
	now time.Time,
) ApplyReturnDisposition {
	service := NewApplyReturnDisposition(store, inventoryapp.NewInMemoryStockMovementStore(), auditLog)
	service.clock = func() time.Time { return now.UTC() }

	return service
}
