package application

import (
	"context"
	"errors"
	"strings"
	"sync"
	"time"

	"github.com/Chinsusu/ERP-v2/apps/api/internal/modules/shipping/domain"
	"github.com/Chinsusu/ERP-v2/apps/api/internal/shared/audit"
)

var ErrCarrierManifestNotFound = errors.New("carrier manifest not found")
var ErrPackedShipmentNotFound = errors.New("packed shipment not found")

type CarrierManifestStore interface {
	List(ctx context.Context, filter domain.CarrierManifestFilter) ([]domain.CarrierManifest, error)
	Get(ctx context.Context, id string) (domain.CarrierManifest, error)
	Save(ctx context.Context, manifest domain.CarrierManifest) error
	GetPackedShipment(ctx context.Context, id string) (domain.PackedShipment, error)
}

type ListCarrierManifests struct {
	store CarrierManifestStore
}

type CreateCarrierManifest struct {
	store    CarrierManifestStore
	auditLog audit.LogStore
	clock    func() time.Time
}

type AddShipmentToCarrierManifest struct {
	store    CarrierManifestStore
	auditLog audit.LogStore
	clock    func() time.Time
}

type CreateCarrierManifestInput struct {
	ID            string
	CarrierCode   string
	CarrierName   string
	WarehouseID   string
	WarehouseCode string
	Date          string
	HandoverBatch string
	StagingZone   string
	Owner         string
	ActorID       string
	RequestID     string
}

type AddShipmentToCarrierManifestInput struct {
	ManifestID string
	ShipmentID string
	ActorID    string
	RequestID  string
}

type CarrierManifestResult struct {
	Manifest   domain.CarrierManifest
	AuditLogID string
}

func NewListCarrierManifests(store CarrierManifestStore) ListCarrierManifests {
	return ListCarrierManifests{store: store}
}

func NewCreateCarrierManifest(store CarrierManifestStore, auditLog audit.LogStore) CreateCarrierManifest {
	return CreateCarrierManifest{
		store:    store,
		auditLog: auditLog,
		clock:    func() time.Time { return time.Now().UTC() },
	}
}

func NewAddShipmentToCarrierManifest(
	store CarrierManifestStore,
	auditLog audit.LogStore,
) AddShipmentToCarrierManifest {
	return AddShipmentToCarrierManifest{
		store:    store,
		auditLog: auditLog,
		clock:    func() time.Time { return time.Now().UTC() },
	}
}

func (uc ListCarrierManifests) Execute(
	ctx context.Context,
	filter domain.CarrierManifestFilter,
) ([]domain.CarrierManifest, error) {
	if uc.store == nil {
		return nil, errors.New("carrier manifest store is required")
	}

	return uc.store.List(ctx, filter)
}

func (uc CreateCarrierManifest) Execute(
	ctx context.Context,
	input CreateCarrierManifestInput,
) (CarrierManifestResult, error) {
	if uc.store == nil {
		return CarrierManifestResult{}, errors.New("carrier manifest store is required")
	}
	if uc.auditLog == nil {
		return CarrierManifestResult{}, errors.New("audit log store is required")
	}

	manifest, err := domain.NewCarrierManifest(domain.NewCarrierManifestInput{
		ID:            input.ID,
		CarrierCode:   input.CarrierCode,
		CarrierName:   input.CarrierName,
		WarehouseID:   input.WarehouseID,
		WarehouseCode: input.WarehouseCode,
		Date:          input.Date,
		HandoverBatch: input.HandoverBatch,
		StagingZone:   input.StagingZone,
		Owner:         input.Owner,
		CreatedAt:     uc.clock(),
	})
	if err != nil {
		return CarrierManifestResult{}, err
	}
	if err := uc.store.Save(ctx, manifest); err != nil {
		return CarrierManifestResult{}, err
	}

	log, err := newManifestAuditLog(
		input.ActorID,
		input.RequestID,
		"shipping.manifest.created",
		manifest,
		map[string]any{"source": "carrier manifest create"},
		manifest.CreatedAt,
	)
	if err != nil {
		return CarrierManifestResult{}, err
	}
	if err := uc.auditLog.Record(ctx, log); err != nil {
		return CarrierManifestResult{}, err
	}

	return CarrierManifestResult{Manifest: manifest, AuditLogID: log.ID}, nil
}

func (uc AddShipmentToCarrierManifest) Execute(
	ctx context.Context,
	input AddShipmentToCarrierManifestInput,
) (CarrierManifestResult, error) {
	if uc.store == nil {
		return CarrierManifestResult{}, errors.New("carrier manifest store is required")
	}
	if uc.auditLog == nil {
		return CarrierManifestResult{}, errors.New("audit log store is required")
	}

	manifest, err := uc.store.Get(ctx, input.ManifestID)
	if err != nil {
		return CarrierManifestResult{}, err
	}
	shipment, err := uc.store.GetPackedShipment(ctx, input.ShipmentID)
	if err != nil {
		return CarrierManifestResult{}, err
	}
	updated, err := manifest.AddShipment(shipment)
	if err != nil {
		return CarrierManifestResult{}, err
	}
	if err := uc.store.Save(ctx, updated); err != nil {
		return CarrierManifestResult{}, err
	}

	log, err := newManifestAuditLog(
		input.ActorID,
		input.RequestID,
		"shipping.manifest.shipment_added",
		updated,
		map[string]any{
			"source":      "carrier manifest add shipment",
			"shipment_id": strings.TrimSpace(input.ShipmentID),
		},
		uc.clock(),
	)
	if err != nil {
		return CarrierManifestResult{}, err
	}
	if err := uc.auditLog.Record(ctx, log); err != nil {
		return CarrierManifestResult{}, err
	}

	return CarrierManifestResult{Manifest: updated, AuditLogID: log.ID}, nil
}

type PrototypeCarrierManifestStore struct {
	mu              sync.RWMutex
	records         map[string]domain.CarrierManifest
	packedShipments map[string]domain.PackedShipment
}

func NewPrototypeCarrierManifestStore() *PrototypeCarrierManifestStore {
	store := &PrototypeCarrierManifestStore{
		records:         make(map[string]domain.CarrierManifest),
		packedShipments: make(map[string]domain.PackedShipment),
	}
	for _, record := range prototypeCarrierManifests() {
		store.records[record.ID] = record.Clone()
	}
	for _, shipment := range prototypePackedShipments() {
		store.packedShipments[shipment.ID] = shipment
	}

	return store
}

func (s *PrototypeCarrierManifestStore) List(
	_ context.Context,
	filter domain.CarrierManifestFilter,
) ([]domain.CarrierManifest, error) {
	if s == nil {
		return nil, errors.New("carrier manifest store is required")
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	rows := make([]domain.CarrierManifest, 0, len(s.records))
	for _, record := range s.records {
		if filter.WarehouseID != "" && record.WarehouseID != filter.WarehouseID {
			continue
		}
		if filter.Date != "" && record.Date != filter.Date {
			continue
		}
		if filter.CarrierCode != "" && record.CarrierCode != filter.CarrierCode {
			continue
		}
		if filter.Status != "" && record.Status != filter.Status {
			continue
		}

		rows = append(rows, record.Clone())
	}
	domain.SortCarrierManifests(rows)

	return rows, nil
}

func (s *PrototypeCarrierManifestStore) Get(_ context.Context, id string) (domain.CarrierManifest, error) {
	if s == nil {
		return domain.CarrierManifest{}, errors.New("carrier manifest store is required")
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	record, ok := s.records[strings.TrimSpace(id)]
	if !ok {
		return domain.CarrierManifest{}, ErrCarrierManifestNotFound
	}

	return record.Clone(), nil
}

func (s *PrototypeCarrierManifestStore) Save(_ context.Context, manifest domain.CarrierManifest) error {
	if s == nil {
		return errors.New("carrier manifest store is required")
	}
	if strings.TrimSpace(manifest.ID) == "" {
		return errors.New("carrier manifest id is required")
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	s.records[manifest.ID] = manifest.Clone()

	return nil
}

func (s *PrototypeCarrierManifestStore) GetPackedShipment(
	_ context.Context,
	id string,
) (domain.PackedShipment, error) {
	if s == nil {
		return domain.PackedShipment{}, errors.New("carrier manifest store is required")
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	shipment, ok := s.packedShipments[strings.TrimSpace(id)]
	if !ok {
		return domain.PackedShipment{}, ErrPackedShipmentNotFound
	}

	return shipment, nil
}

func newManifestAuditLog(
	actorID string,
	requestID string,
	action string,
	manifest domain.CarrierManifest,
	metadata map[string]any,
	createdAt time.Time,
) (audit.Log, error) {
	summary := manifest.Summary()
	return audit.NewLog(audit.NewLogInput{
		OrgID:      "org-my-pham",
		ActorID:    strings.TrimSpace(actorID),
		Action:     action,
		EntityType: "shipping.carrier_manifest",
		EntityID:   manifest.ID,
		RequestID:  strings.TrimSpace(requestID),
		AfterData: map[string]any{
			"status":         string(manifest.Status),
			"carrier_code":   manifest.CarrierCode,
			"warehouse_id":   manifest.WarehouseID,
			"date":           manifest.Date,
			"expected_count": summary.ExpectedCount,
			"scanned_count":  summary.ScannedCount,
			"missing_count":  summary.MissingCount,
		},
		Metadata:  metadata,
		CreatedAt: createdAt,
	})
}

func prototypeCarrierManifests() []domain.CarrierManifest {
	return []domain.CarrierManifest{
		{
			ID:            "manifest-hcm-ghn-morning",
			CarrierCode:   "GHN",
			CarrierName:   "GHN Express",
			WarehouseID:   "wh-hcm",
			WarehouseCode: "HCM",
			Date:          "2026-04-26",
			HandoverBatch: "morning",
			StagingZone:   "handover-a",
			Status:        domain.ManifestStatusScanning,
			Owner:         "Handover Operator",
			CreatedAt:     time.Date(2026, 4, 26, 7, 45, 0, 0, time.UTC),
			Lines: []domain.CarrierManifestLine{
				{ID: "line-ship-hcm-001", ShipmentID: "ship-hcm-260426-001", OrderNo: "SO-260426-001", TrackingNo: "GHN260426001", PackageCode: "TOTE-A01", StagingZone: "handover-a", Scanned: true},
				{ID: "line-ship-hcm-002", ShipmentID: "ship-hcm-260426-002", OrderNo: "SO-260426-002", TrackingNo: "GHN260426002", PackageCode: "TOTE-A01", StagingZone: "handover-a", Scanned: true},
				{ID: "line-ship-hcm-003", ShipmentID: "ship-hcm-260426-003", OrderNo: "SO-260426-003", TrackingNo: "GHN260426003", PackageCode: "TOTE-A02", StagingZone: "handover-a"},
			},
		},
		{
			ID:            "manifest-hcm-vtp-noon",
			CarrierCode:   "VTP",
			CarrierName:   "Viettel Post",
			WarehouseID:   "wh-hcm",
			WarehouseCode: "HCM",
			Date:          "2026-04-26",
			HandoverBatch: "noon",
			StagingZone:   "handover-b",
			Status:        domain.ManifestStatusReady,
			Owner:         "Warehouse Lead",
			CreatedAt:     time.Date(2026, 4, 26, 9, 0, 0, 0, time.UTC),
			Lines: []domain.CarrierManifestLine{
				{ID: "line-ship-hcm-vtp-001", ShipmentID: "ship-hcm-vtp-260426-001", OrderNo: "SO-260426-011", TrackingNo: "VTP260426011", PackageCode: "TOTE-B01", StagingZone: "handover-b"},
			},
		},
		{
			ID:            "manifest-hn-ghn-day",
			CarrierCode:   "GHN",
			CarrierName:   "GHN Express",
			WarehouseID:   "wh-hn",
			WarehouseCode: "HN",
			Date:          "2026-04-26",
			HandoverBatch: "day",
			StagingZone:   "hn-handover",
			Status:        domain.ManifestStatusCompleted,
			Owner:         "HN Lead",
			CreatedAt:     time.Date(2026, 4, 26, 8, 20, 0, 0, time.UTC),
			Lines: []domain.CarrierManifestLine{
				{ID: "line-ship-hn-001", ShipmentID: "ship-hn-260426-001", OrderNo: "SO-260426-HN-011", TrackingNo: "GHNHN260426001", PackageCode: "HN-TOTE-01", StagingZone: "hn-handover", Scanned: true},
			},
		},
	}
}

func prototypePackedShipments() []domain.PackedShipment {
	return []domain.PackedShipment{
		{ID: "ship-hcm-260426-001", OrderNo: "SO-260426-001", TrackingNo: "GHN260426001", CarrierCode: "GHN", CarrierName: "GHN Express", WarehouseID: "wh-hcm", WarehouseCode: "HCM", PackageCode: "TOTE-A01", StagingZone: "handover-a", Packed: true},
		{ID: "ship-hcm-260426-002", OrderNo: "SO-260426-002", TrackingNo: "GHN260426002", CarrierCode: "GHN", CarrierName: "GHN Express", WarehouseID: "wh-hcm", WarehouseCode: "HCM", PackageCode: "TOTE-A01", StagingZone: "handover-a", Packed: true},
		{ID: "ship-hcm-260426-003", OrderNo: "SO-260426-003", TrackingNo: "GHN260426003", CarrierCode: "GHN", CarrierName: "GHN Express", WarehouseID: "wh-hcm", WarehouseCode: "HCM", PackageCode: "TOTE-A02", StagingZone: "handover-a", Packed: true},
		{ID: "ship-hcm-260426-004", OrderNo: "SO-260426-004", TrackingNo: "GHN260426004", CarrierCode: "GHN", CarrierName: "GHN Express", WarehouseID: "wh-hcm", WarehouseCode: "HCM", PackageCode: "TOTE-A03", StagingZone: "handover-a", Packed: true},
		{ID: "ship-hcm-vtp-260426-001", OrderNo: "SO-260426-011", TrackingNo: "VTP260426011", CarrierCode: "VTP", CarrierName: "Viettel Post", WarehouseID: "wh-hcm", WarehouseCode: "HCM", PackageCode: "TOTE-B01", StagingZone: "handover-b", Packed: true},
		{ID: "ship-hn-260426-001", OrderNo: "SO-260426-HN-011", TrackingNo: "GHNHN260426001", CarrierCode: "GHN", CarrierName: "GHN Express", WarehouseID: "wh-hn", WarehouseCode: "HN", PackageCode: "HN-TOTE-01", StagingZone: "hn-handover", Packed: true},
	}
}
