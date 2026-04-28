package application

import (
	"context"
	"errors"
	"fmt"
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
	FindPackedShipmentByCode(ctx context.Context, code string) (domain.PackedShipment, error)
	FindCarrierManifestLineByCode(
		ctx context.Context,
		code string,
	) (domain.CarrierManifest, domain.CarrierManifestLine, error)
	RecordScanEvent(ctx context.Context, event CarrierManifestScanEvent) error
}

type ListCarrierManifests struct {
	store CarrierManifestStore
}

type CreateCarrierManifest struct {
	store          CarrierManifestStore
	carrierCatalog CarrierMasterStore
	auditLog       audit.LogStore
	clock          func() time.Time
}

type AddShipmentToCarrierManifest struct {
	store    CarrierManifestStore
	auditLog audit.LogStore
	clock    func() time.Time
}

type VerifyCarrierManifestScan struct {
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

type VerifyCarrierManifestScanInput struct {
	ManifestID string
	Code       string
	StationID  string
	ActorID    string
	RequestID  string
}

type CarrierManifestResult struct {
	Manifest   domain.CarrierManifest
	AuditLogID string
}

type CarrierManifestScanEvent struct {
	ID                 string
	ManifestID         string
	ExpectedManifestID string
	Code               string
	ResultCode         domain.CarrierManifestScanResultCode
	Severity           string
	Message            string
	ShipmentID         string
	OrderNo            string
	TrackingNo         string
	ActorID            string
	StationID          string
	WarehouseID        string
	CarrierCode        string
	CreatedAt          time.Time
}

type CarrierManifestScanResult struct {
	Manifest           domain.CarrierManifest
	Line               *domain.CarrierManifestLine
	Event              CarrierManifestScanEvent
	AuditLogID         string
	Code               domain.CarrierManifestScanResultCode
	Severity           string
	Message            string
	ExpectedManifestID string
}

func NewListCarrierManifests(store CarrierManifestStore) ListCarrierManifests {
	return ListCarrierManifests{store: store}
}

func NewCreateCarrierManifest(store CarrierManifestStore, auditLog audit.LogStore) CreateCarrierManifest {
	return NewCreateCarrierManifestWithCarrierCatalog(store, auditLog, NewPrototypeCarrierCatalog())
}

func NewCreateCarrierManifestWithCarrierCatalog(
	store CarrierManifestStore,
	auditLog audit.LogStore,
	carrierCatalog CarrierMasterStore,
) CreateCarrierManifest {
	return CreateCarrierManifest{
		store:          store,
		carrierCatalog: carrierCatalog,
		auditLog:       auditLog,
		clock:          func() time.Time { return time.Now().UTC() },
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

func NewVerifyCarrierManifestScan(
	store CarrierManifestStore,
	auditLog audit.LogStore,
) VerifyCarrierManifestScan {
	return VerifyCarrierManifestScan{
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
	carrier, err := uc.activeCarrier(ctx, input.CarrierCode)
	if err != nil {
		return CarrierManifestResult{}, err
	}

	manifest, err := domain.NewCarrierManifest(domain.NewCarrierManifestInput{
		ID:            input.ID,
		CarrierCode:   carrier.Code,
		CarrierName:   carrier.Name,
		WarehouseID:   input.WarehouseID,
		WarehouseCode: input.WarehouseCode,
		Date:          input.Date,
		HandoverBatch: input.HandoverBatch,
		StagingZone:   carrierHandoverZone(input.StagingZone, carrier),
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
		map[string]any{
			"source":              "carrier manifest create",
			"carrier_sla_profile": carrier.SLAProfile,
			"handover_zone":       manifest.StagingZone,
		},
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

func (uc CreateCarrierManifest) activeCarrier(ctx context.Context, carrierCode string) (domain.Carrier, error) {
	if uc.carrierCatalog == nil {
		return domain.Carrier{}, errors.New("carrier catalog is required")
	}

	carrier, err := uc.carrierCatalog.GetCarrierByCode(ctx, carrierCode)
	if err != nil {
		return domain.Carrier{}, err
	}
	if !carrier.IsActive() {
		return domain.Carrier{}, ErrCarrierInactive
	}

	return carrier, nil
}

func carrierHandoverZone(inputZone string, carrier domain.Carrier) string {
	if strings.TrimSpace(inputZone) != "" {
		return inputZone
	}

	return carrier.HandoverZone
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

func (uc VerifyCarrierManifestScan) Execute(
	ctx context.Context,
	input VerifyCarrierManifestScanInput,
) (CarrierManifestScanResult, error) {
	if uc.store == nil {
		return CarrierManifestScanResult{}, errors.New("carrier manifest store is required")
	}
	if uc.auditLog == nil {
		return CarrierManifestScanResult{}, errors.New("audit log store is required")
	}
	if domain.NormalizeManifestScanCode(input.Code) == "" {
		return CarrierManifestScanResult{}, domain.ErrManifestScanCodeRequired
	}

	manifest, err := uc.store.Get(ctx, input.ManifestID)
	if err != nil {
		return CarrierManifestScanResult{}, err
	}

	result := uc.evaluateScan(ctx, manifest, input.Code)
	if result.Code == domain.ScanResultMatched {
		if err := uc.store.Save(ctx, result.Manifest); err != nil {
			return CarrierManifestScanResult{}, err
		}
	}

	event := newCarrierManifestScanEvent(input, result, uc.clock())
	if err := uc.store.RecordScanEvent(ctx, event); err != nil {
		return CarrierManifestScanResult{}, err
	}

	log, err := newManifestScanAuditLog(input.ActorID, input.RequestID, result, event, event.CreatedAt)
	if err != nil {
		return CarrierManifestScanResult{}, err
	}
	if err := uc.auditLog.Record(ctx, log); err != nil {
		return CarrierManifestScanResult{}, err
	}

	result.Event = event
	result.AuditLogID = log.ID

	return result, nil
}

func (uc VerifyCarrierManifestScan) evaluateScan(
	ctx context.Context,
	manifest domain.CarrierManifest,
	code string,
) CarrierManifestScanResult {
	updated, line, err := manifest.MarkLineScanned(code)
	switch {
	case err == nil:
		return CarrierManifestScanResult{
			Manifest: updated,
			Line:     &line,
			Code:     domain.ScanResultMatched,
			Severity: "success",
			Message:  "Scan matched manifest line",
		}
	case errors.Is(err, domain.ErrManifestScanDuplicate):
		_, currentLine, _ := manifest.FindLineByScanCode(code)
		return CarrierManifestScanResult{
			Manifest: manifest,
			Line:     &currentLine,
			Code:     domain.ScanResultDuplicate,
			Severity: "warning",
			Message:  "Shipment was already scanned for this manifest",
		}
	case errors.Is(err, domain.ErrManifestScanInvalidState):
		result := CarrierManifestScanResult{
			Manifest: manifest,
			Code:     domain.ScanResultInvalidState,
			Severity: "danger",
			Message:  "Manifest cannot accept scans in its current state",
		}
		if _, currentLine, ok := manifest.FindLineByScanCode(code); ok {
			result.Line = &currentLine
		}

		return result
	case errors.Is(err, domain.ErrManifestScanNotFound):
		return uc.evaluateMissingScanCode(ctx, manifest, code)
	default:
		return CarrierManifestScanResult{
			Manifest: manifest,
			Code:     domain.ScanResultNotFound,
			Severity: "danger",
			Message:  "Scan code was not found",
		}
	}
}

func (uc VerifyCarrierManifestScan) evaluateMissingScanCode(
	ctx context.Context,
	manifest domain.CarrierManifest,
	code string,
) CarrierManifestScanResult {
	expectedManifest, expectedLine, err := uc.store.FindCarrierManifestLineByCode(ctx, code)
	if err == nil {
		return CarrierManifestScanResult{
			Manifest:           manifest,
			Line:               &expectedLine,
			Code:               domain.ScanResultManifestMismatch,
			Severity:           "danger",
			Message:            "Scan code belongs to a different manifest",
			ExpectedManifestID: expectedManifest.ID,
		}
	}

	shipment, err := uc.store.FindPackedShipmentByCode(ctx, code)
	if err == nil {
		if !shipment.Packed {
			return CarrierManifestScanResult{
				Manifest: manifest,
				Line: &domain.CarrierManifestLine{
					ShipmentID:  shipment.ID,
					OrderNo:     shipment.OrderNo,
					TrackingNo:  shipment.TrackingNo,
					PackageCode: shipment.PackageCode,
					StagingZone: shipment.StagingZone,
				},
				Code:     domain.ScanResultInvalidState,
				Severity: "danger",
				Message:  "Shipment is not packed and cannot be handed over",
			}
		}

		return CarrierManifestScanResult{
			Manifest: manifest,
			Line: &domain.CarrierManifestLine{
				ShipmentID:  shipment.ID,
				OrderNo:     shipment.OrderNo,
				TrackingNo:  shipment.TrackingNo,
				PackageCode: shipment.PackageCode,
				StagingZone: shipment.StagingZone,
			},
			Code:     domain.ScanResultManifestMismatch,
			Severity: "danger",
			Message:  "Packed shipment is not expected on this manifest",
		}
	}

	return CarrierManifestScanResult{
		Manifest: manifest,
		Code:     domain.ScanResultNotFound,
		Severity: "danger",
		Message:  "Scan code was not found",
	}
}

type PrototypeCarrierManifestStore struct {
	mu              sync.RWMutex
	records         map[string]domain.CarrierManifest
	packedShipments map[string]domain.PackedShipment
	scanEvents      []CarrierManifestScanEvent
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

func (s *PrototypeCarrierManifestStore) FindPackedShipmentByCode(
	_ context.Context,
	code string,
) (domain.PackedShipment, error) {
	if s == nil {
		return domain.PackedShipment{}, errors.New("carrier manifest store is required")
	}

	normalizedCode := domain.NormalizeManifestScanCode(code)
	if normalizedCode == "" {
		return domain.PackedShipment{}, ErrPackedShipmentNotFound
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	for _, shipment := range s.packedShipments {
		for _, candidate := range []string{
			shipment.ID,
			shipment.OrderNo,
			shipment.TrackingNo,
			shipment.PackageCode,
		} {
			if domain.NormalizeManifestScanCode(candidate) == normalizedCode {
				return shipment, nil
			}
		}
	}

	return domain.PackedShipment{}, ErrPackedShipmentNotFound
}

func (s *PrototypeCarrierManifestStore) FindCarrierManifestLineByCode(
	_ context.Context,
	code string,
) (domain.CarrierManifest, domain.CarrierManifestLine, error) {
	if s == nil {
		return domain.CarrierManifest{}, domain.CarrierManifestLine{}, errors.New("carrier manifest store is required")
	}

	normalizedCode := domain.NormalizeManifestScanCode(code)
	if normalizedCode == "" {
		return domain.CarrierManifest{}, domain.CarrierManifestLine{}, ErrPackedShipmentNotFound
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	for _, manifest := range s.records {
		_, line, ok := manifest.FindLineByScanCode(normalizedCode)
		if ok {
			return manifest.Clone(), line, nil
		}
	}

	return domain.CarrierManifest{}, domain.CarrierManifestLine{}, ErrPackedShipmentNotFound
}

func (s *PrototypeCarrierManifestStore) RecordScanEvent(_ context.Context, event CarrierManifestScanEvent) error {
	if s == nil {
		return errors.New("carrier manifest store is required")
	}
	if strings.TrimSpace(event.ID) == "" {
		return errors.New("carrier manifest scan event id is required")
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	s.scanEvents = append(s.scanEvents, event)

	return nil
}

func (s *PrototypeCarrierManifestStore) ListScanEvents(
	_ context.Context,
	manifestID string,
) ([]CarrierManifestScanEvent, error) {
	if s == nil {
		return nil, errors.New("carrier manifest store is required")
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	events := make([]CarrierManifestScanEvent, 0, len(s.scanEvents))
	for _, event := range s.scanEvents {
		if strings.TrimSpace(manifestID) != "" && event.ManifestID != strings.TrimSpace(manifestID) {
			continue
		}
		events = append(events, event)
	}

	return events, nil
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

func newCarrierManifestScanEvent(
	input VerifyCarrierManifestScanInput,
	result CarrierManifestScanResult,
	createdAt time.Time,
) CarrierManifestScanEvent {
	if createdAt.IsZero() {
		createdAt = time.Now().UTC()
	}

	event := CarrierManifestScanEvent{
		ID:                 fmt.Sprintf("scan_%d", createdAt.UnixNano()),
		ManifestID:         result.Manifest.ID,
		ExpectedManifestID: strings.TrimSpace(result.ExpectedManifestID),
		Code:               domain.NormalizeManifestScanCode(input.Code),
		ResultCode:         result.Code,
		Severity:           strings.TrimSpace(result.Severity),
		Message:            strings.TrimSpace(result.Message),
		ActorID:            strings.TrimSpace(input.ActorID),
		StationID:          strings.TrimSpace(input.StationID),
		WarehouseID:        result.Manifest.WarehouseID,
		CarrierCode:        result.Manifest.CarrierCode,
		CreatedAt:          createdAt.UTC(),
	}
	if event.StationID == "" {
		event.StationID = "shipping-handover"
	}
	if result.Line != nil {
		event.ShipmentID = result.Line.ShipmentID
		event.OrderNo = result.Line.OrderNo
		event.TrackingNo = result.Line.TrackingNo
	}

	return event
}

func newManifestScanAuditLog(
	actorID string,
	requestID string,
	result CarrierManifestScanResult,
	event CarrierManifestScanEvent,
	createdAt time.Time,
) (audit.Log, error) {
	return audit.NewLog(audit.NewLogInput{
		OrgID:      "org-my-pham",
		ActorID:    strings.TrimSpace(actorID),
		Action:     "shipping.manifest.scan_recorded",
		EntityType: "shipping.scan_event",
		EntityID:   event.ID,
		RequestID:  strings.TrimSpace(requestID),
		AfterData: map[string]any{
			"manifest_id":          event.ManifestID,
			"expected_manifest_id": event.ExpectedManifestID,
			"code":                 event.Code,
			"result_code":          string(result.Code),
			"severity":             result.Severity,
			"shipment_id":          event.ShipmentID,
			"order_no":             event.OrderNo,
			"tracking_no":          event.TrackingNo,
		},
		Metadata: map[string]any{
			"source":     "carrier manifest scan",
			"station_id": event.StationID,
		},
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
		{ID: "ship-hcm-260426-099", OrderNo: "SO-260426-099", TrackingNo: "GHN260426099", CarrierCode: "GHN", CarrierName: "GHN Express", WarehouseID: "wh-hcm", WarehouseCode: "HCM", PackageCode: "PACKING-LANE-01", StagingZone: "packing", Packed: false},
		{ID: "ship-hcm-vtp-260426-001", OrderNo: "SO-260426-011", TrackingNo: "VTP260426011", CarrierCode: "VTP", CarrierName: "Viettel Post", WarehouseID: "wh-hcm", WarehouseCode: "HCM", PackageCode: "TOTE-B01", StagingZone: "handover-b", Packed: true},
		{ID: "ship-hn-260426-001", OrderNo: "SO-260426-HN-011", TrackingNo: "GHNHN260426001", CarrierCode: "GHN", CarrierName: "GHN Express", WarehouseID: "wh-hn", WarehouseCode: "HN", PackageCode: "HN-TOTE-01", StagingZone: "hn-handover", Packed: true},
	}
}
