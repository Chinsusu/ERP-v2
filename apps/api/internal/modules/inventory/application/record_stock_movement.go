package application

import (
	"context"
	"errors"
	"sync"

	"github.com/Chinsusu/ERP-v2/apps/api/internal/modules/inventory/domain"
)

type StockMovementStore interface {
	Record(ctx context.Context, movement domain.StockMovement) error
}

type InMemoryStockMovementStore struct {
	mu        sync.Mutex
	movements []domain.StockMovement
}

func NewInMemoryStockMovementStore() *InMemoryStockMovementStore {
	return &InMemoryStockMovementStore{}
}

func (s *InMemoryStockMovementStore) Record(_ context.Context, movement domain.StockMovement) error {
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
	if uc.store == nil {
		return errors.New("stock movement store is required")
	}
	if err := movement.Validate(); err != nil {
		return err
	}

	return uc.store.Record(ctx, movement)
}
