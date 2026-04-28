package application

import (
	"context"
	"errors"
	"strings"
	"sync"
	"time"

	"github.com/Chinsusu/ERP-v2/apps/api/internal/modules/shipping/domain"
)

var ErrCarrierNotFound = errors.New("carrier not found")
var ErrCarrierInactive = errors.New("carrier is inactive")

type CarrierMasterStore interface {
	ListCarriers(ctx context.Context, filter domain.CarrierFilter) ([]domain.Carrier, error)
	GetCarrierByCode(ctx context.Context, code string) (domain.Carrier, error)
}

type ListCarriers struct {
	store CarrierMasterStore
}

type GetCarrier struct {
	store CarrierMasterStore
}

type PrototypeCarrierCatalog struct {
	mu      sync.RWMutex
	records map[string]domain.Carrier
}

func NewListCarriers(store CarrierMasterStore) ListCarriers {
	return ListCarriers{store: store}
}

func NewGetCarrier(store CarrierMasterStore) GetCarrier {
	return GetCarrier{store: store}
}

func (uc ListCarriers) Execute(ctx context.Context, filter domain.CarrierFilter) ([]domain.Carrier, error) {
	if uc.store == nil {
		return nil, errors.New("carrier catalog is required")
	}

	return uc.store.ListCarriers(ctx, filter)
}

func (uc GetCarrier) Execute(ctx context.Context, code string) (domain.Carrier, error) {
	if uc.store == nil {
		return domain.Carrier{}, errors.New("carrier catalog is required")
	}

	return uc.store.GetCarrierByCode(ctx, code)
}

func NewPrototypeCarrierCatalog(carriers ...domain.Carrier) *PrototypeCarrierCatalog {
	store := &PrototypeCarrierCatalog{records: make(map[string]domain.Carrier)}
	if len(carriers) == 0 {
		carriers = prototypeCarriers()
	}
	for _, carrier := range carriers {
		if err := carrier.Validate(); err != nil {
			continue
		}
		store.records[carrier.Code] = carrier.Clone()
	}

	return store
}

func (s *PrototypeCarrierCatalog) ListCarriers(
	_ context.Context,
	filter domain.CarrierFilter,
) ([]domain.Carrier, error) {
	if s == nil {
		return nil, errors.New("carrier catalog is required")
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	rows := make([]domain.Carrier, 0, len(s.records))
	for _, record := range s.records {
		if filter.Status != "" && record.Status != filter.Status {
			continue
		}
		if filter.Search != "" && !carrierMatchesSearch(record, filter.Search) {
			continue
		}
		rows = append(rows, record.Clone())
	}
	domain.SortCarriers(rows)

	return rows, nil
}

func (s *PrototypeCarrierCatalog) GetCarrierByCode(_ context.Context, code string) (domain.Carrier, error) {
	if s == nil {
		return domain.Carrier{}, errors.New("carrier catalog is required")
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	carrier, ok := s.records[strings.ToUpper(strings.TrimSpace(code))]
	if !ok {
		return domain.Carrier{}, ErrCarrierNotFound
	}

	return carrier.Clone(), nil
}

func carrierMatchesSearch(carrier domain.Carrier, search string) bool {
	needle := strings.ToLower(strings.TrimSpace(search))
	if needle == "" {
		return true
	}

	return strings.Contains(strings.ToLower(carrier.Code), needle) ||
		strings.Contains(strings.ToLower(carrier.Name), needle) ||
		strings.Contains(strings.ToLower(carrier.HandoverZone), needle)
}

func prototypeCarriers() []domain.Carrier {
	baseTime := time.Date(2026, 4, 28, 8, 0, 0, 0, time.UTC)
	inputs := []domain.NewCarrierInput{
		{
			Code:         "GHN",
			Name:         "GHN Express",
			HandoverZone: "handover-a",
			Status:       domain.CarrierStatusActive,
			SLAProfile:   "standard",
			CreatedAt:    baseTime,
		},
		{
			Code:         "VTP",
			Name:         "Viettel Post",
			HandoverZone: "handover-b",
			Status:       domain.CarrierStatusActive,
			SLAProfile:   "standard",
			CreatedAt:    baseTime,
		},
		{
			Code:         "NJV",
			Name:         "Ninja Van",
			HandoverZone: "handover-c",
			Status:       domain.CarrierStatusActive,
			SLAProfile:   "standard",
			CreatedAt:    baseTime,
		},
		{
			Code:         "GHTK",
			Name:         "Giao Hang Tiet Kiem",
			HandoverZone: "handover-d",
			Status:       domain.CarrierStatusInactive,
			SLAProfile:   "paused",
			CreatedAt:    baseTime,
		},
	}

	carriers := make([]domain.Carrier, 0, len(inputs))
	for _, input := range inputs {
		carrier, err := domain.NewCarrier(input)
		if err != nil {
			continue
		}
		carriers = append(carriers, carrier)
	}

	return carriers
}
