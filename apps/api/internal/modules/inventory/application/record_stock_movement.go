package application

import (
	"context"
	"sync"

	"github.com/Chinsusu/ERP-v2/apps/api/internal/modules/inventory/domain"
)

type StockMovementStore interface {
	Append(ctx context.Context, movement domain.StockMovement) error
}

type InMemoryStockMovementStore struct {
	mu        sync.Mutex
	movements []domain.StockMovement
}

func NewInMemoryStockMovementStore() *InMemoryStockMovementStore {
	return &InMemoryStockMovementStore{}
}

func (s *InMemoryStockMovementStore) Append(_ context.Context, movement domain.StockMovement) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.movements = append(s.movements, movement)
	return nil
}

func (s *InMemoryStockMovementStore) Count() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return len(s.movements)
}

type RecordStockMovement struct {
	store StockMovementStore
}

func NewRecordStockMovement(store StockMovementStore) RecordStockMovement {
	return RecordStockMovement{store: store}
}

func (uc RecordStockMovement) Execute(ctx context.Context, movement domain.StockMovement) error {
	return uc.store.Append(ctx, movement)
}
