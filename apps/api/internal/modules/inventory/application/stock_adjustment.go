package application

import (
	"context"
	"errors"
	"strings"
	"sync"
	"time"

	"github.com/Chinsusu/ERP-v2/apps/api/internal/modules/inventory/domain"
	"github.com/Chinsusu/ERP-v2/apps/api/internal/shared/audit"
	"github.com/Chinsusu/ERP-v2/apps/api/internal/shared/decimal"
)

const (
	stockAdjustmentCreatedAction = "inventory.stock_adjustment.created"
	stockAdjustmentEntityType    = "inventory.stock_adjustment"
)

type StockAdjustmentStore interface {
	ListStockAdjustments(ctx context.Context) ([]domain.StockAdjustment, error)
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
	log, err := newStockAdjustmentAuditLog(input.RequestedBy, input.RequestID, adjustment)
	if err != nil {
		return StockAdjustmentResult{}, err
	}
	if err := uc.auditLog.Record(ctx, log); err != nil {
		return StockAdjustmentResult{}, err
	}

	return StockAdjustmentResult{Adjustment: adjustment, AuditLogID: log.ID}, nil
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

func newStockAdjustmentAuditLog(actorID string, requestID string, adjustment domain.StockAdjustment) (audit.Log, error) {
	return audit.NewLog(audit.NewLogInput{
		OrgID:      adjustment.OrgID,
		ActorID:    strings.TrimSpace(actorID),
		Action:     stockAdjustmentCreatedAction,
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
		},
		Metadata: map[string]any{
			"source": "inventory adjustment request",
		},
		CreatedAt: adjustment.CreatedAt,
	})
}
