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

var ErrStockCountNotFound = errors.New("stock count not found")

const (
	stockCountCreatedAction   = "inventory.stock_count.created"
	stockCountSubmittedAction = "inventory.stock_count.submitted"
	stockCountEntityType      = "inventory.stock_count"
)

type StockCountStore interface {
	ListStockCounts(ctx context.Context) ([]domain.StockCountSession, error)
	FindStockCountByID(ctx context.Context, id string) (domain.StockCountSession, error)
	SaveStockCount(ctx context.Context, session domain.StockCountSession) error
}

type PrototypeStockCountStore struct {
	mu      sync.RWMutex
	records map[string]domain.StockCountSession
}

type ListStockCounts struct {
	store StockCountStore
}

type CreateStockCount struct {
	store    StockCountStore
	auditLog audit.LogStore
	clock    func() time.Time
}

type SubmitStockCount struct {
	store    StockCountStore
	auditLog audit.LogStore
	clock    func() time.Time
}

type CreateStockCountInput struct {
	ID            string
	CountNo       string
	OrgID         string
	WarehouseID   string
	WarehouseCode string
	Scope         string
	CreatedBy     string
	RequestID     string
	Lines         []CreateStockCountLineInput
}

type CreateStockCountLineInput struct {
	ID           string
	ItemID       string
	SKU          string
	BatchID      string
	BatchNo      string
	LocationID   string
	LocationCode string
	ExpectedQty  string
	BaseUOMCode  string
}

type SubmitStockCountInput struct {
	ID          string
	SubmittedBy string
	RequestID   string
	Lines       []SubmitStockCountLineInput
}

type SubmitStockCountLineInput struct {
	ID         string
	SKU        string
	CountedQty string
	Note       string
}

type StockCountResult struct {
	Session    domain.StockCountSession
	AuditLogID string
}

func NewPrototypeStockCountStore() *PrototypeStockCountStore {
	return &PrototypeStockCountStore{records: make(map[string]domain.StockCountSession)}
}

func NewListStockCounts(store StockCountStore) ListStockCounts {
	return ListStockCounts{store: store}
}

func NewCreateStockCount(store StockCountStore, auditLog audit.LogStore) CreateStockCount {
	return CreateStockCount{
		store:    store,
		auditLog: auditLog,
		clock:    func() time.Time { return time.Now().UTC() },
	}
}

func NewSubmitStockCount(store StockCountStore, auditLog audit.LogStore) SubmitStockCount {
	return SubmitStockCount{
		store:    store,
		auditLog: auditLog,
		clock:    func() time.Time { return time.Now().UTC() },
	}
}

func (uc ListStockCounts) Execute(ctx context.Context) ([]domain.StockCountSession, error) {
	if uc.store == nil {
		return nil, errors.New("stock count store is required")
	}

	return uc.store.ListStockCounts(ctx)
}

func (uc CreateStockCount) Execute(ctx context.Context, input CreateStockCountInput) (StockCountResult, error) {
	if uc.store == nil {
		return StockCountResult{}, errors.New("stock count store is required")
	}
	if uc.auditLog == nil {
		return StockCountResult{}, errors.New("audit log store is required")
	}
	lines, err := newStockCountLines(input.Lines)
	if err != nil {
		return StockCountResult{}, err
	}
	session, err := domain.NewStockCountSession(domain.NewStockCountSessionInput{
		ID:            input.ID,
		CountNo:       input.CountNo,
		OrgID:         input.OrgID,
		WarehouseID:   input.WarehouseID,
		WarehouseCode: input.WarehouseCode,
		Scope:         input.Scope,
		CreatedBy:     input.CreatedBy,
		Lines:         lines,
		CreatedAt:     uc.clock(),
	})
	if err != nil {
		return StockCountResult{}, err
	}
	if err := uc.store.SaveStockCount(ctx, session); err != nil {
		return StockCountResult{}, err
	}
	log, err := newStockCountAuditLog(input.CreatedBy, input.RequestID, stockCountCreatedAction, session)
	if err != nil {
		return StockCountResult{}, err
	}
	if err := uc.auditLog.Record(ctx, log); err != nil {
		return StockCountResult{}, err
	}

	return StockCountResult{Session: session, AuditLogID: log.ID}, nil
}

func (uc SubmitStockCount) Execute(ctx context.Context, input SubmitStockCountInput) (StockCountResult, error) {
	if uc.store == nil {
		return StockCountResult{}, errors.New("stock count store is required")
	}
	if uc.auditLog == nil {
		return StockCountResult{}, errors.New("audit log store is required")
	}
	current, err := uc.store.FindStockCountByID(ctx, input.ID)
	if err != nil {
		return StockCountResult{}, err
	}
	lines, err := newSubmitStockCountLines(input.Lines)
	if err != nil {
		return StockCountResult{}, err
	}
	submitted, err := current.Submit(lines, input.SubmittedBy, uc.clock())
	if err != nil {
		return StockCountResult{}, err
	}
	if err := uc.store.SaveStockCount(ctx, submitted); err != nil {
		return StockCountResult{}, err
	}
	log, err := newStockCountAuditLog(input.SubmittedBy, input.RequestID, stockCountSubmittedAction, submitted)
	if err != nil {
		return StockCountResult{}, err
	}
	if err := uc.auditLog.Record(ctx, log); err != nil {
		return StockCountResult{}, err
	}

	return StockCountResult{Session: submitted, AuditLogID: log.ID}, nil
}

func newStockCountLines(inputs []CreateStockCountLineInput) ([]domain.NewStockCountLineInput, error) {
	lines := make([]domain.NewStockCountLineInput, 0, len(inputs))
	for _, input := range inputs {
		expectedQty, err := decimal.ParseQuantity(input.ExpectedQty)
		if err != nil {
			return nil, domain.ErrStockCountInvalidQuantity
		}
		lines = append(lines, domain.NewStockCountLineInput{
			ID:           input.ID,
			ItemID:       input.ItemID,
			SKU:          input.SKU,
			BatchID:      input.BatchID,
			BatchNo:      input.BatchNo,
			LocationID:   input.LocationID,
			LocationCode: input.LocationCode,
			ExpectedQty:  expectedQty,
			BaseUOMCode:  input.BaseUOMCode,
		})
	}

	return lines, nil
}

func newSubmitStockCountLines(inputs []SubmitStockCountLineInput) ([]domain.SubmitStockCountLineInput, error) {
	lines := make([]domain.SubmitStockCountLineInput, 0, len(inputs))
	for _, input := range inputs {
		countedQty, err := decimal.ParseQuantity(input.CountedQty)
		if err != nil {
			return nil, domain.ErrStockCountInvalidQuantity
		}
		lines = append(lines, domain.SubmitStockCountLineInput{
			ID:         input.ID,
			SKU:        input.SKU,
			CountedQty: countedQty,
			Note:       input.Note,
		})
	}

	return lines, nil
}

func (s *PrototypeStockCountStore) ListStockCounts(_ context.Context) ([]domain.StockCountSession, error) {
	if s == nil {
		return nil, errors.New("stock count store is required")
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	rows := make([]domain.StockCountSession, 0, len(s.records))
	for _, session := range s.records {
		rows = append(rows, session.Clone())
	}
	domain.SortStockCountSessions(rows)

	return rows, nil
}

func (s *PrototypeStockCountStore) FindStockCountByID(_ context.Context, id string) (domain.StockCountSession, error) {
	if s == nil {
		return domain.StockCountSession{}, errors.New("stock count store is required")
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	session, ok := s.records[strings.TrimSpace(id)]
	if !ok {
		return domain.StockCountSession{}, ErrStockCountNotFound
	}

	return session.Clone(), nil
}

func (s *PrototypeStockCountStore) SaveStockCount(_ context.Context, session domain.StockCountSession) error {
	if s == nil {
		return errors.New("stock count store is required")
	}
	if strings.TrimSpace(session.ID) == "" {
		return domain.ErrStockCountRequiredField
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.records[session.ID] = session.Clone()

	return nil
}

func newStockCountAuditLog(actorID string, requestID string, action string, session domain.StockCountSession) (audit.Log, error) {
	return audit.NewLog(audit.NewLogInput{
		OrgID:      session.OrgID,
		ActorID:    strings.TrimSpace(actorID),
		Action:     action,
		EntityType: stockCountEntityType,
		EntityID:   session.ID,
		RequestID:  strings.TrimSpace(requestID),
		AfterData: map[string]any{
			"count_no":     session.CountNo,
			"status":       string(session.Status),
			"warehouse_id": session.WarehouseID,
			"scope":        session.Scope,
			"line_count":   len(session.Lines),
			"has_variance": session.HasVariance(),
			"submitted_by": session.SubmittedBy,
			"submitted_at": timeString(session.SubmittedAt),
		},
		Metadata: map[string]any{
			"source": "stock count session",
		},
		CreatedAt: session.UpdatedAt,
	})
}

func timeString(value time.Time) string {
	if value.IsZero() {
		return ""
	}

	return value.UTC().Format(time.RFC3339)
}
