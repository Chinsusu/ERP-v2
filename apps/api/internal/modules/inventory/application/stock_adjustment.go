package application

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/Chinsusu/ERP-v2/apps/api/internal/modules/inventory/domain"
	"github.com/Chinsusu/ERP-v2/apps/api/internal/shared/audit"
	"github.com/Chinsusu/ERP-v2/apps/api/internal/shared/decimal"
)

var ErrStockAdjustmentNotFound = errors.New("stock adjustment not found")

const (
	stockAdjustmentCreatedAction   = "inventory.stock_adjustment.created"
	stockAdjustmentSubmittedAction = "inventory.stock_adjustment.submitted"
	stockAdjustmentApprovedAction  = "inventory.stock_adjustment.approved"
	stockAdjustmentRejectedAction  = "inventory.stock_adjustment.rejected"
	stockAdjustmentPostedAction    = "inventory.stock_adjustment.posted"
	stockAdjustmentEntityType      = "inventory.stock_adjustment"
)

type StockAdjustmentStore interface {
	ListStockAdjustments(ctx context.Context) ([]domain.StockAdjustment, error)
	FindStockAdjustmentByID(ctx context.Context, id string) (domain.StockAdjustment, error)
	SaveStockAdjustment(ctx context.Context, adjustment domain.StockAdjustment) error
}

type PrototypeStockAdjustmentStore struct {
	mu      sync.RWMutex
	records map[string]domain.StockAdjustment
}

type ListStockAdjustments struct {
	store StockAdjustmentStore
}

type CreateStockAdjustment struct {
	store    StockAdjustmentStore
	auditLog audit.LogStore
	clock    func() time.Time
}

type TransitionStockAdjustment struct {
	store    StockAdjustmentStore
	movement StockMovementStore
	auditLog audit.LogStore
	clock    func() time.Time
}

type CreateStockAdjustmentInput struct {
	ID            string
	AdjustmentNo  string
	OrgID         string
	WarehouseID   string
	WarehouseCode string
	SourceType    string
	SourceID      string
	Reason        string
	RequestedBy   string
	RequestID     string
	Lines         []CreateStockAdjustmentLineInput
}

type CreateStockAdjustmentLineInput struct {
	ID           string
	ItemID       string
	SKU          string
	BatchID      string
	BatchNo      string
	LocationID   string
	LocationCode string
	ExpectedQty  string
	CountedQty   string
	BaseUOMCode  string
	Reason       string
}

type StockAdjustmentResult struct {
	Adjustment domain.StockAdjustment
	AuditLogID string
}

func NewPrototypeStockAdjustmentStore() *PrototypeStockAdjustmentStore {
	return &PrototypeStockAdjustmentStore{records: make(map[string]domain.StockAdjustment)}
}

func NewListStockAdjustments(store StockAdjustmentStore) ListStockAdjustments {
	return ListStockAdjustments{store: store}
}

func NewCreateStockAdjustment(store StockAdjustmentStore, auditLog audit.LogStore) CreateStockAdjustment {
	return CreateStockAdjustment{
		store:    store,
		auditLog: auditLog,
		clock:    func() time.Time { return time.Now().UTC() },
	}
}

func NewTransitionStockAdjustment(
	store StockAdjustmentStore,
	movement StockMovementStore,
	auditLog audit.LogStore,
) TransitionStockAdjustment {
	return TransitionStockAdjustment{
		store:    store,
		movement: movement,
		auditLog: auditLog,
		clock:    func() time.Time { return time.Now().UTC() },
	}
}

func (uc ListStockAdjustments) Execute(ctx context.Context) ([]domain.StockAdjustment, error) {
	if uc.store == nil {
		return nil, errors.New("stock adjustment store is required")
	}

	return uc.store.ListStockAdjustments(ctx)
}

func (uc CreateStockAdjustment) Execute(
	ctx context.Context,
	input CreateStockAdjustmentInput,
) (StockAdjustmentResult, error) {
	if uc.store == nil {
		return StockAdjustmentResult{}, errors.New("stock adjustment store is required")
	}
	if uc.auditLog == nil {
		return StockAdjustmentResult{}, errors.New("audit log store is required")
	}

	lines, err := newStockAdjustmentLines(input.Lines)
	if err != nil {
		return StockAdjustmentResult{}, err
	}
	adjustment, err := domain.NewStockAdjustment(domain.NewStockAdjustmentInput{
		ID:            input.ID,
		AdjustmentNo:  input.AdjustmentNo,
		OrgID:         input.OrgID,
		WarehouseID:   input.WarehouseID,
		WarehouseCode: input.WarehouseCode,
		SourceType:    input.SourceType,
		SourceID:      input.SourceID,
		Reason:        input.Reason,
		RequestedBy:   input.RequestedBy,
		Lines:         lines,
		CreatedAt:     uc.clock(),
	})
	if err != nil {
		return StockAdjustmentResult{}, err
	}
	if err := uc.store.SaveStockAdjustment(ctx, adjustment); err != nil {
		return StockAdjustmentResult{}, err
	}
	log, err := newStockAdjustmentAuditLog(input.RequestedBy, input.RequestID, stockAdjustmentCreatedAction, adjustment)
	if err != nil {
		return StockAdjustmentResult{}, err
	}
	if err := uc.auditLog.Record(ctx, log); err != nil {
		return StockAdjustmentResult{}, err
	}

	return StockAdjustmentResult{Adjustment: adjustment, AuditLogID: log.ID}, nil
}

func (uc TransitionStockAdjustment) Submit(
	ctx context.Context,
	id string,
	actorID string,
	requestID string,
) (StockAdjustmentResult, error) {
	return uc.transition(ctx, id, actorID, requestID, stockAdjustmentSubmittedAction, func(adjustment domain.StockAdjustment, actorID string, at time.Time) (domain.StockAdjustment, error) {
		return adjustment.Submit(actorID, at)
	})
}

func (uc TransitionStockAdjustment) Approve(
	ctx context.Context,
	id string,
	actorID string,
	requestID string,
) (StockAdjustmentResult, error) {
	return uc.transition(ctx, id, actorID, requestID, stockAdjustmentApprovedAction, func(adjustment domain.StockAdjustment, actorID string, at time.Time) (domain.StockAdjustment, error) {
		return adjustment.Approve(actorID, at)
	})
}

func (uc TransitionStockAdjustment) Reject(
	ctx context.Context,
	id string,
	actorID string,
	requestID string,
) (StockAdjustmentResult, error) {
	return uc.transition(ctx, id, actorID, requestID, stockAdjustmentRejectedAction, func(adjustment domain.StockAdjustment, actorID string, at time.Time) (domain.StockAdjustment, error) {
		return adjustment.Reject(actorID, at)
	})
}

func (uc TransitionStockAdjustment) Post(
	ctx context.Context,
	id string,
	actorID string,
	requestID string,
) (StockAdjustmentResult, error) {
	if uc.movement == nil {
		return StockAdjustmentResult{}, errors.New("stock movement store is required")
	}
	return uc.transition(ctx, id, actorID, requestID, stockAdjustmentPostedAction, func(adjustment domain.StockAdjustment, actorID string, at time.Time) (domain.StockAdjustment, error) {
		posted, err := adjustment.MarkPosted(actorID, at)
		if err != nil {
			return domain.StockAdjustment{}, err
		}
		movements, err := newStockAdjustmentMovements(posted, actorID, at)
		if err != nil {
			return domain.StockAdjustment{}, err
		}
		for _, movement := range movements {
			if err := uc.movement.Record(ctx, movement); err != nil {
				return domain.StockAdjustment{}, err
			}
		}

		return posted, nil
	})
}

func (uc TransitionStockAdjustment) transition(
	ctx context.Context,
	id string,
	actorID string,
	requestID string,
	action string,
	apply func(adjustment domain.StockAdjustment, actorID string, at time.Time) (domain.StockAdjustment, error),
) (StockAdjustmentResult, error) {
	if uc.store == nil {
		return StockAdjustmentResult{}, errors.New("stock adjustment store is required")
	}
	if uc.auditLog == nil {
		return StockAdjustmentResult{}, errors.New("audit log store is required")
	}
	current, err := uc.store.FindStockAdjustmentByID(ctx, id)
	if err != nil {
		return StockAdjustmentResult{}, err
	}
	updated, err := apply(current, actorID, uc.clock())
	if err != nil {
		return StockAdjustmentResult{}, err
	}
	if err := uc.store.SaveStockAdjustment(ctx, updated); err != nil {
		return StockAdjustmentResult{}, err
	}
	log, err := newStockAdjustmentAuditLog(actorID, requestID, action, updated)
	if err != nil {
		return StockAdjustmentResult{}, err
	}
	if err := uc.auditLog.Record(ctx, log); err != nil {
		return StockAdjustmentResult{}, err
	}

	return StockAdjustmentResult{Adjustment: updated, AuditLogID: log.ID}, nil
}

func newStockAdjustmentLines(
	inputs []CreateStockAdjustmentLineInput,
) ([]domain.NewStockAdjustmentLineInput, error) {
	lines := make([]domain.NewStockAdjustmentLineInput, 0, len(inputs))
	for _, input := range inputs {
		expectedQty, err := decimal.ParseQuantity(input.ExpectedQty)
		if err != nil {
			return nil, domain.ErrStockAdjustmentInvalidQuantity
		}
		countedQty, err := decimal.ParseQuantity(input.CountedQty)
		if err != nil {
			return nil, domain.ErrStockAdjustmentInvalidQuantity
		}
		lines = append(lines, domain.NewStockAdjustmentLineInput{
			ID:           input.ID,
			ItemID:       input.ItemID,
			SKU:          input.SKU,
			BatchID:      input.BatchID,
			BatchNo:      input.BatchNo,
			LocationID:   input.LocationID,
			LocationCode: input.LocationCode,
			ExpectedQty:  expectedQty,
			CountedQty:   countedQty,
			BaseUOMCode:  input.BaseUOMCode,
			Reason:       input.Reason,
		})
	}

	return lines, nil
}

func newStockAdjustmentMovements(
	adjustment domain.StockAdjustment,
	actorID string,
	movementAt time.Time,
) ([]domain.StockMovement, error) {
	movements := make([]domain.StockMovement, 0, len(adjustment.Lines))
	for index, line := range adjustment.Lines {
		if line.DeltaQty.IsZero() {
			continue
		}
		movementType := domain.MovementAdjustmentIn
		quantity := line.DeltaQty
		if line.DeltaQty.IsNegative() {
			movementType = domain.MovementAdjustmentOut
			quantity = decimal.Decimal(strings.TrimPrefix(line.DeltaQty.String(), "-"))
		}
		movement, err := domain.NewStockMovement(domain.NewStockMovementInput{
			MovementNo:       fmt.Sprintf("%s-MV-%03d", adjustment.AdjustmentNo, index+1),
			MovementType:     movementType,
			OrgID:            adjustment.OrgID,
			ItemID:           stockAdjustmentItemID(line),
			BatchID:          line.BatchID,
			WarehouseID:      adjustment.WarehouseID,
			BinID:            line.LocationID,
			Quantity:         quantity,
			BaseUOMCode:      line.BaseUOMCode.String(),
			SourceQuantity:   quantity,
			SourceUOMCode:    line.BaseUOMCode.String(),
			ConversionFactor: decimal.MustQuantity("1"),
			StockStatus:      domain.StockStatusAvailable,
			SourceDocType:    "stock_adjustment",
			SourceDocID:      adjustment.ID,
			SourceDocLineID:  line.ID,
			Reason:           adjustment.Reason,
			CreatedBy:        actorID,
			MovementAt:       movementAt,
		})
		if err != nil {
			return nil, err
		}
		movements = append(movements, movement)
	}

	return movements, nil
}

func stockAdjustmentItemID(line domain.StockAdjustmentLine) string {
	if strings.TrimSpace(line.ItemID) != "" {
		return strings.TrimSpace(line.ItemID)
	}
	sku := strings.ToLower(strings.TrimSpace(line.SKU))
	if sku == "" {
		return "item-unknown-sku"
	}

	return "item-" + sku
}

func (s *PrototypeStockAdjustmentStore) ListStockAdjustments(_ context.Context) ([]domain.StockAdjustment, error) {
	if s == nil {
		return nil, errors.New("stock adjustment store is required")
	}

	s.mu.RLock()
	defer s.mu.RUnlock()
	rows := make([]domain.StockAdjustment, 0, len(s.records))
	for _, adjustment := range s.records {
		rows = append(rows, adjustment.Clone())
	}
	domain.SortStockAdjustments(rows)

	return rows, nil
}

func (s *PrototypeStockAdjustmentStore) FindStockAdjustmentByID(_ context.Context, id string) (domain.StockAdjustment, error) {
	if s == nil {
		return domain.StockAdjustment{}, errors.New("stock adjustment store is required")
	}

	s.mu.RLock()
	defer s.mu.RUnlock()
	adjustment, ok := s.records[strings.TrimSpace(id)]
	if !ok {
		return domain.StockAdjustment{}, ErrStockAdjustmentNotFound
	}

	return adjustment.Clone(), nil
}

func (s *PrototypeStockAdjustmentStore) SaveStockAdjustment(_ context.Context, adjustment domain.StockAdjustment) error {
	if s == nil {
		return errors.New("stock adjustment store is required")
	}
	if strings.TrimSpace(adjustment.ID) == "" {
		return domain.ErrStockAdjustmentRequiredField
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	s.records[adjustment.ID] = adjustment.Clone()

	return nil
}

func newStockAdjustmentAuditLog(actorID string, requestID string, action string, adjustment domain.StockAdjustment) (audit.Log, error) {
	createdAt := adjustment.CreatedAt
	if action != stockAdjustmentCreatedAction {
		createdAt = adjustment.UpdatedAt
	}

	return audit.NewLog(audit.NewLogInput{
		OrgID:      adjustment.OrgID,
		ActorID:    strings.TrimSpace(actorID),
		Action:     action,
		EntityType: stockAdjustmentEntityType,
		EntityID:   adjustment.ID,
		RequestID:  strings.TrimSpace(requestID),
		AfterData: map[string]any{
			"adjustment_no": adjustment.AdjustmentNo,
			"status":        string(adjustment.Status),
			"warehouse_id":  adjustment.WarehouseID,
			"source_type":   adjustment.SourceType,
			"source_id":     adjustment.SourceID,
			"line_count":    len(adjustment.Lines),
			"reason":        adjustment.Reason,
			"submitted_by":  adjustment.SubmittedBy,
			"approved_by":   adjustment.ApprovedBy,
			"rejected_by":   adjustment.RejectedBy,
			"posted_by":     adjustment.PostedBy,
		},
		Metadata: map[string]any{
			"source": "inventory adjustment request",
		},
		CreatedAt: createdAt,
	})
}
