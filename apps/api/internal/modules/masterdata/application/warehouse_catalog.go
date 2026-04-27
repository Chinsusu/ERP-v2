package application

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/Chinsusu/ERP-v2/apps/api/internal/modules/masterdata/domain"
	"github.com/Chinsusu/ERP-v2/apps/api/internal/shared/audit"
	"github.com/Chinsusu/ERP-v2/apps/api/internal/shared/response"
)

var ErrWarehouseNotFound = errors.New("warehouse not found")
var ErrLocationNotFound = errors.New("warehouse location not found")
var ErrDuplicateWarehouseCode = errors.New("warehouse code already exists")
var ErrDuplicateLocationCode = errors.New("warehouse location code already exists")
var ErrInvalidLocationWarehouse = errors.New("warehouse location references an invalid warehouse")
var ErrInactiveLocation = errors.New("warehouse location is inactive")

type WarehouseLocationCatalog struct {
	mu         sync.RWMutex
	warehouses map[string]domain.Warehouse
	locations  map[string]domain.Location
	auditLog   audit.LogStore
	clock      func() time.Time
}

type CreateWarehouseInput struct {
	Code            string
	Name            string
	Type            string
	SiteCode        string
	Address         string
	AllowSaleIssue  bool
	AllowProdIssue  bool
	AllowQuarantine bool
	Status          string
	ActorID         string
	RequestID       string
}

type UpdateWarehouseInput struct {
	ID              string
	Code            string
	Name            string
	Type            string
	SiteCode        string
	Address         string
	AllowSaleIssue  bool
	AllowProdIssue  bool
	AllowQuarantine bool
	Status          string
	ActorID         string
	RequestID       string
}

type ChangeWarehouseStatusInput struct {
	ID        string
	Status    string
	ActorID   string
	RequestID string
}

type CreateLocationInput struct {
	WarehouseID  string
	Code         string
	Name         string
	Type         string
	ZoneCode     string
	AllowReceive bool
	AllowPick    bool
	AllowStore   bool
	IsDefault    bool
	Status       string
	ActorID      string
	RequestID    string
}

type UpdateLocationInput struct {
	ID           string
	WarehouseID  string
	Code         string
	Name         string
	Type         string
	ZoneCode     string
	AllowReceive bool
	AllowPick    bool
	AllowStore   bool
	IsDefault    bool
	Status       string
	ActorID      string
	RequestID    string
}

type ChangeLocationStatusInput struct {
	ID        string
	Status    string
	ActorID   string
	RequestID string
}

type WarehouseResult struct {
	Warehouse  domain.Warehouse
	AuditLogID string
}

type LocationResult struct {
	Location   domain.Location
	AuditLogID string
}

func NewPrototypeWarehouseLocationCatalog(auditLog audit.LogStore) *WarehouseLocationCatalog {
	store := &WarehouseLocationCatalog{
		warehouses: make(map[string]domain.Warehouse),
		locations:  make(map[string]domain.Location),
		auditLog:   auditLog,
		clock:      func() time.Time { return time.Now().UTC() },
	}
	for _, warehouse := range prototypeWarehouses() {
		store.warehouses[warehouse.ID] = warehouse.Clone()
	}
	for _, location := range prototypeLocations() {
		store.locations[location.ID] = location.Clone()
	}

	return store
}

func NewPrototypeWarehouseLocationCatalogAt(auditLog audit.LogStore, now time.Time) *WarehouseLocationCatalog {
	store := NewPrototypeWarehouseLocationCatalog(auditLog)
	store.clock = func() time.Time { return now.UTC() }

	return store
}

func (s *WarehouseLocationCatalog) ListWarehouses(_ context.Context, filter domain.WarehouseFilter) ([]domain.Warehouse, response.Pagination, error) {
	if s == nil {
		return nil, response.Pagination{}, errors.New("warehouse catalog is required")
	}
	if filter.Status != "" && !domain.IsValidWarehouseStatus(filter.Status) {
		return nil, response.Pagination{}, domain.ErrWarehouseInvalidStatus
	}
	if filter.Type != "" && !domain.IsValidWarehouseType(filter.Type) {
		return nil, response.Pagination{}, domain.ErrWarehouseInvalidType
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	rows := make([]domain.Warehouse, 0, len(s.warehouses))
	for _, warehouse := range s.warehouses {
		if filter.Matches(warehouse) {
			rows = append(rows, warehouse.Clone())
		}
	}
	domain.SortWarehouses(rows)
	pageRows, pagination := paginateWarehouses(rows, filter.Page, filter.PageSize)

	return pageRows, pagination, nil
}

func (s *WarehouseLocationCatalog) GetWarehouse(_ context.Context, id string) (domain.Warehouse, error) {
	if s == nil {
		return domain.Warehouse{}, errors.New("warehouse catalog is required")
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	warehouse, ok := s.warehouses[strings.TrimSpace(id)]
	if !ok {
		return domain.Warehouse{}, ErrWarehouseNotFound
	}

	return warehouse.Clone(), nil
}

func (s *WarehouseLocationCatalog) CreateWarehouse(ctx context.Context, input CreateWarehouseInput) (WarehouseResult, error) {
	if s == nil {
		return WarehouseResult{}, errors.New("warehouse catalog is required")
	}
	if s.auditLog == nil {
		return WarehouseResult{}, errors.New("audit log store is required")
	}

	now := s.clock()
	warehouse, err := domain.NewWarehouse(domain.NewWarehouseInput{
		ID:              newWarehouseID(input.Code, now),
		Code:            input.Code,
		Name:            input.Name,
		Type:            domain.WarehouseType(input.Type),
		SiteCode:        input.SiteCode,
		Address:         input.Address,
		AllowSaleIssue:  input.AllowSaleIssue,
		AllowProdIssue:  input.AllowProdIssue,
		AllowQuarantine: input.AllowQuarantine,
		Status:          domain.WarehouseStatus(input.Status),
		CreatedAt:       now,
		UpdatedAt:       now,
	})
	if err != nil {
		return WarehouseResult{}, err
	}

	s.mu.Lock()
	if err := s.ensureUniqueWarehouseLocked(warehouse, ""); err != nil {
		s.mu.Unlock()
		return WarehouseResult{}, err
	}
	s.warehouses[warehouse.ID] = warehouse.Clone()
	s.mu.Unlock()

	log, err := newWarehouseAuditLog(input.ActorID, input.RequestID, "masterdata.warehouse.created", warehouse, nil, warehouseToAuditMap(warehouse), now)
	if err != nil {
		return WarehouseResult{}, err
	}
	if err := s.auditLog.Record(ctx, log); err != nil {
		return WarehouseResult{}, err
	}

	return WarehouseResult{Warehouse: warehouse, AuditLogID: log.ID}, nil
}

func (s *WarehouseLocationCatalog) UpdateWarehouse(ctx context.Context, input UpdateWarehouseInput) (WarehouseResult, error) {
	if s == nil {
		return WarehouseResult{}, errors.New("warehouse catalog is required")
	}
	if s.auditLog == nil {
		return WarehouseResult{}, errors.New("audit log store is required")
	}

	now := s.clock()
	s.mu.Lock()
	current, ok := s.warehouses[strings.TrimSpace(input.ID)]
	if !ok {
		s.mu.Unlock()
		return WarehouseResult{}, ErrWarehouseNotFound
	}
	updated, err := current.Update(domain.UpdateWarehouseInput{
		Code:            input.Code,
		Name:            input.Name,
		Type:            domain.WarehouseType(input.Type),
		SiteCode:        input.SiteCode,
		Address:         input.Address,
		AllowSaleIssue:  input.AllowSaleIssue,
		AllowProdIssue:  input.AllowProdIssue,
		AllowQuarantine: input.AllowQuarantine,
		Status:          domain.WarehouseStatus(input.Status),
		UpdatedAt:       now,
	})
	if err != nil {
		s.mu.Unlock()
		return WarehouseResult{}, err
	}
	if err := s.ensureUniqueWarehouseLocked(updated, current.ID); err != nil {
		s.mu.Unlock()
		return WarehouseResult{}, err
	}
	s.warehouses[current.ID] = updated.Clone()
	for id, location := range s.locations {
		if location.WarehouseID == current.ID {
			location.WarehouseCode = updated.Code
			location.UpdatedAt = now
			s.locations[id] = location.Clone()
		}
	}
	s.mu.Unlock()

	log, err := newWarehouseAuditLog(input.ActorID, input.RequestID, "masterdata.warehouse.updated", updated, warehouseToAuditMap(current), warehouseToAuditMap(updated), now)
	if err != nil {
		return WarehouseResult{}, err
	}
	if err := s.auditLog.Record(ctx, log); err != nil {
		return WarehouseResult{}, err
	}

	return WarehouseResult{Warehouse: updated, AuditLogID: log.ID}, nil
}

func (s *WarehouseLocationCatalog) ChangeWarehouseStatus(ctx context.Context, input ChangeWarehouseStatusInput) (WarehouseResult, error) {
	if s == nil {
		return WarehouseResult{}, errors.New("warehouse catalog is required")
	}
	if s.auditLog == nil {
		return WarehouseResult{}, errors.New("audit log store is required")
	}

	now := s.clock()
	s.mu.Lock()
	current, ok := s.warehouses[strings.TrimSpace(input.ID)]
	if !ok {
		s.mu.Unlock()
		return WarehouseResult{}, ErrWarehouseNotFound
	}
	updated, err := current.ChangeStatus(domain.WarehouseStatus(input.Status), now)
	if err != nil {
		s.mu.Unlock()
		return WarehouseResult{}, err
	}
	s.warehouses[current.ID] = updated.Clone()
	s.mu.Unlock()

	log, err := newWarehouseAuditLog(
		input.ActorID,
		input.RequestID,
		"masterdata.warehouse.status_changed",
		updated,
		map[string]any{"status": string(current.Status)},
		map[string]any{"status": string(updated.Status)},
		now,
	)
	if err != nil {
		return WarehouseResult{}, err
	}
	if err := s.auditLog.Record(ctx, log); err != nil {
		return WarehouseResult{}, err
	}

	return WarehouseResult{Warehouse: updated, AuditLogID: log.ID}, nil
}

func (s *WarehouseLocationCatalog) ListLocations(_ context.Context, filter domain.LocationFilter) ([]domain.Location, response.Pagination, error) {
	if s == nil {
		return nil, response.Pagination{}, errors.New("warehouse catalog is required")
	}
	if filter.Status != "" && !domain.IsValidLocationStatus(filter.Status) {
		return nil, response.Pagination{}, domain.ErrLocationInvalidStatus
	}
	if filter.Type != "" && !domain.IsValidLocationType(filter.Type) {
		return nil, response.Pagination{}, domain.ErrLocationInvalidType
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	rows := make([]domain.Location, 0, len(s.locations))
	for _, location := range s.locations {
		if filter.Matches(location) {
			rows = append(rows, location.Clone())
		}
	}
	domain.SortLocations(rows)
	pageRows, pagination := paginateLocations(rows, filter.Page, filter.PageSize)

	return pageRows, pagination, nil
}

func (s *WarehouseLocationCatalog) GetLocation(_ context.Context, id string) (domain.Location, error) {
	if s == nil {
		return domain.Location{}, errors.New("warehouse catalog is required")
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	location, ok := s.locations[strings.TrimSpace(id)]
	if !ok {
		return domain.Location{}, ErrLocationNotFound
	}

	return location.Clone(), nil
}

func (s *WarehouseLocationCatalog) CreateLocation(ctx context.Context, input CreateLocationInput) (LocationResult, error) {
	if s == nil {
		return LocationResult{}, errors.New("warehouse catalog is required")
	}
	if s.auditLog == nil {
		return LocationResult{}, errors.New("audit log store is required")
	}

	now := s.clock()
	s.mu.Lock()
	warehouse, ok := s.warehouses[strings.TrimSpace(input.WarehouseID)]
	if !ok {
		s.mu.Unlock()
		return LocationResult{}, ErrInvalidLocationWarehouse
	}
	location, err := domain.NewLocation(domain.NewLocationInput{
		ID:            newLocationID(warehouse.Code, input.Code, now),
		WarehouseID:   warehouse.ID,
		WarehouseCode: warehouse.Code,
		Code:          input.Code,
		Name:          input.Name,
		Type:          domain.LocationType(input.Type),
		ZoneCode:      input.ZoneCode,
		AllowReceive:  input.AllowReceive,
		AllowPick:     input.AllowPick,
		AllowStore:    input.AllowStore,
		IsDefault:     input.IsDefault,
		Status:        domain.LocationStatus(input.Status),
		CreatedAt:     now,
		UpdatedAt:     now,
	})
	if err != nil {
		s.mu.Unlock()
		return LocationResult{}, err
	}
	if err := s.ensureUniqueLocationLocked(location, ""); err != nil {
		s.mu.Unlock()
		return LocationResult{}, err
	}
	s.locations[location.ID] = location.Clone()
	s.mu.Unlock()

	log, err := newLocationAuditLog(input.ActorID, input.RequestID, "masterdata.location.created", location, nil, locationToAuditMap(location), now)
	if err != nil {
		return LocationResult{}, err
	}
	if err := s.auditLog.Record(ctx, log); err != nil {
		return LocationResult{}, err
	}

	return LocationResult{Location: location, AuditLogID: log.ID}, nil
}

func (s *WarehouseLocationCatalog) UpdateLocation(ctx context.Context, input UpdateLocationInput) (LocationResult, error) {
	if s == nil {
		return LocationResult{}, errors.New("warehouse catalog is required")
	}
	if s.auditLog == nil {
		return LocationResult{}, errors.New("audit log store is required")
	}

	now := s.clock()
	s.mu.Lock()
	current, ok := s.locations[strings.TrimSpace(input.ID)]
	if !ok {
		s.mu.Unlock()
		return LocationResult{}, ErrLocationNotFound
	}
	if current.Status != domain.LocationStatusActive && domain.NormalizeLocationStatus(domain.LocationStatus(input.Status)) != domain.LocationStatusActive {
		s.mu.Unlock()
		return LocationResult{}, ErrInactiveLocation
	}
	warehouse, ok := s.warehouses[strings.TrimSpace(input.WarehouseID)]
	if !ok {
		s.mu.Unlock()
		return LocationResult{}, ErrInvalidLocationWarehouse
	}
	updated, err := current.Update(domain.UpdateLocationInput{
		WarehouseID:   warehouse.ID,
		WarehouseCode: warehouse.Code,
		Code:          input.Code,
		Name:          input.Name,
		Type:          domain.LocationType(input.Type),
		ZoneCode:      input.ZoneCode,
		AllowReceive:  input.AllowReceive,
		AllowPick:     input.AllowPick,
		AllowStore:    input.AllowStore,
		IsDefault:     input.IsDefault,
		Status:        domain.LocationStatus(input.Status),
		UpdatedAt:     now,
	})
	if err != nil {
		s.mu.Unlock()
		return LocationResult{}, err
	}
	if err := s.ensureUniqueLocationLocked(updated, current.ID); err != nil {
		s.mu.Unlock()
		return LocationResult{}, err
	}
	s.locations[current.ID] = updated.Clone()
	s.mu.Unlock()

	log, err := newLocationAuditLog(input.ActorID, input.RequestID, "masterdata.location.updated", updated, locationToAuditMap(current), locationToAuditMap(updated), now)
	if err != nil {
		return LocationResult{}, err
	}
	if err := s.auditLog.Record(ctx, log); err != nil {
		return LocationResult{}, err
	}

	return LocationResult{Location: updated, AuditLogID: log.ID}, nil
}

func (s *WarehouseLocationCatalog) ChangeLocationStatus(ctx context.Context, input ChangeLocationStatusInput) (LocationResult, error) {
	if s == nil {
		return LocationResult{}, errors.New("warehouse catalog is required")
	}
	if s.auditLog == nil {
		return LocationResult{}, errors.New("audit log store is required")
	}

	now := s.clock()
	s.mu.Lock()
	current, ok := s.locations[strings.TrimSpace(input.ID)]
	if !ok {
		s.mu.Unlock()
		return LocationResult{}, ErrLocationNotFound
	}
	updated, err := current.ChangeStatus(domain.LocationStatus(input.Status), now)
	if err != nil {
		s.mu.Unlock()
		return LocationResult{}, err
	}
	s.locations[current.ID] = updated.Clone()
	s.mu.Unlock()

	log, err := newLocationAuditLog(
		input.ActorID,
		input.RequestID,
		"masterdata.location.status_changed",
		updated,
		map[string]any{"status": string(current.Status)},
		map[string]any{"status": string(updated.Status)},
		now,
	)
	if err != nil {
		return LocationResult{}, err
	}
	if err := s.auditLog.Record(ctx, log); err != nil {
		return LocationResult{}, err
	}

	return LocationResult{Location: updated, AuditLogID: log.ID}, nil
}

func (s *WarehouseLocationCatalog) ensureUniqueWarehouseLocked(warehouse domain.Warehouse, currentID string) error {
	for _, existing := range s.warehouses {
		if strings.TrimSpace(currentID) != "" && existing.ID == currentID {
			continue
		}
		if existing.Code == warehouse.Code {
			return ErrDuplicateWarehouseCode
		}
	}

	return nil
}

func (s *WarehouseLocationCatalog) ensureUniqueLocationLocked(location domain.Location, currentID string) error {
	for _, existing := range s.locations {
		if strings.TrimSpace(currentID) != "" && existing.ID == currentID {
			continue
		}
		if existing.WarehouseID == location.WarehouseID && existing.Code == location.Code {
			return ErrDuplicateLocationCode
		}
	}

	return nil
}

func paginateWarehouses(warehouses []domain.Warehouse, page int, pageSize int) ([]domain.Warehouse, response.Pagination) {
	totalItems := len(warehouses)
	totalPages := 0
	if totalItems > 0 {
		totalPages = (totalItems + pageSize - 1) / pageSize
	}
	start := (page - 1) * pageSize
	if start >= totalItems {
		return []domain.Warehouse{}, response.Pagination{
			Page:       page,
			PageSize:   pageSize,
			TotalItems: totalItems,
			TotalPages: totalPages,
		}
	}
	end := start + pageSize
	if end > totalItems {
		end = totalItems
	}

	return append([]domain.Warehouse(nil), warehouses[start:end]...), response.Pagination{
		Page:       page,
		PageSize:   pageSize,
		TotalItems: totalItems,
		TotalPages: totalPages,
	}
}

func paginateLocations(locations []domain.Location, page int, pageSize int) ([]domain.Location, response.Pagination) {
	totalItems := len(locations)
	totalPages := 0
	if totalItems > 0 {
		totalPages = (totalItems + pageSize - 1) / pageSize
	}
	start := (page - 1) * pageSize
	if start >= totalItems {
		return []domain.Location{}, response.Pagination{
			Page:       page,
			PageSize:   pageSize,
			TotalItems: totalItems,
			TotalPages: totalPages,
		}
	}
	end := start + pageSize
	if end > totalItems {
		end = totalItems
	}

	return append([]domain.Location(nil), locations[start:end]...), response.Pagination{
		Page:       page,
		PageSize:   pageSize,
		TotalItems: totalItems,
		TotalPages: totalPages,
	}
}

func newWarehouseAuditLog(actorID string, requestID string, action string, warehouse domain.Warehouse, beforeData map[string]any, afterData map[string]any, createdAt time.Time) (audit.Log, error) {
	return audit.NewLog(audit.NewLogInput{
		OrgID:      "org-my-pham",
		ActorID:    strings.TrimSpace(actorID),
		Action:     strings.TrimSpace(action),
		EntityType: "mdm.warehouse",
		EntityID:   warehouse.ID,
		RequestID:  strings.TrimSpace(requestID),
		BeforeData: beforeData,
		AfterData:  afterData,
		Metadata: map[string]any{
			"source":         "warehouse master data",
			"warehouse_code": warehouse.Code,
			"site_code":      warehouse.SiteCode,
		},
		CreatedAt: createdAt,
	})
}

func newLocationAuditLog(actorID string, requestID string, action string, location domain.Location, beforeData map[string]any, afterData map[string]any, createdAt time.Time) (audit.Log, error) {
	return audit.NewLog(audit.NewLogInput{
		OrgID:      "org-my-pham",
		ActorID:    strings.TrimSpace(actorID),
		Action:     strings.TrimSpace(action),
		EntityType: "mdm.warehouse_location",
		EntityID:   location.ID,
		RequestID:  strings.TrimSpace(requestID),
		BeforeData: beforeData,
		AfterData:  afterData,
		Metadata: map[string]any{
			"source":         "warehouse location master data",
			"warehouse_code": location.WarehouseCode,
			"location_code":  location.Code,
		},
		CreatedAt: createdAt,
	})
}

func warehouseToAuditMap(warehouse domain.Warehouse) map[string]any {
	return map[string]any{
		"warehouse_code":   warehouse.Code,
		"warehouse_name":   warehouse.Name,
		"warehouse_type":   string(warehouse.Type),
		"site_code":        warehouse.SiteCode,
		"address":          warehouse.Address,
		"allow_sale_issue": warehouse.AllowSaleIssue,
		"allow_prod_issue": warehouse.AllowProdIssue,
		"allow_quarantine": warehouse.AllowQuarantine,
		"status":           string(warehouse.Status),
	}
}

func locationToAuditMap(location domain.Location) map[string]any {
	return map[string]any{
		"warehouse_id":   location.WarehouseID,
		"warehouse_code": location.WarehouseCode,
		"location_code":  location.Code,
		"location_name":  location.Name,
		"location_type":  string(location.Type),
		"zone_code":      location.ZoneCode,
		"allow_receive":  location.AllowReceive,
		"allow_pick":     location.AllowPick,
		"allow_store":    location.AllowStore,
		"is_default":     location.IsDefault,
		"status":         string(location.Status),
	}
}

func newWarehouseID(code string, now time.Time) string {
	value := strings.ToLower(domain.NormalizeWarehouseCode(code))
	value = strings.ReplaceAll(value, "-", "_")
	if value == "" {
		value = "warehouse"
	}

	return fmt.Sprintf("wh_%s_%d", value, now.UnixNano())
}

func newLocationID(warehouseCode string, code string, now time.Time) string {
	value := strings.ToLower(domain.NormalizeWarehouseCode(warehouseCode) + "_" + domain.NormalizeLocationCode(code))
	value = strings.ReplaceAll(value, "-", "_")
	if strings.Trim(value, "_") == "" {
		value = "warehouse_location"
	}

	return fmt.Sprintf("loc_%s_%d", value, now.UnixNano())
}

func prototypeWarehouses() []domain.Warehouse {
	baseTime := time.Date(2026, 4, 26, 9, 0, 0, 0, time.UTC)
	warehouses := make([]domain.Warehouse, 0, 4)
	for _, input := range []domain.NewWarehouseInput{
		{
			ID:              "wh-hcm-fg",
			Code:            "WH-HCM-FG",
			Name:            "Finished Goods Warehouse HCM",
			Type:            domain.WarehouseTypeFinishedGood,
			SiteCode:        "HCM",
			Address:         "Ho Chi Minh distribution center",
			AllowSaleIssue:  true,
			AllowProdIssue:  false,
			AllowQuarantine: false,
			Status:          domain.WarehouseStatusActive,
			CreatedAt:       baseTime,
			UpdatedAt:       baseTime,
		},
		{
			ID:              "wh-hcm-rm",
			Code:            "WH-HCM-RM",
			Name:            "Raw Material Warehouse HCM",
			Type:            domain.WarehouseTypeRawMaterial,
			SiteCode:        "HCM",
			Address:         "Ho Chi Minh production site",
			AllowSaleIssue:  false,
			AllowProdIssue:  true,
			AllowQuarantine: false,
			Status:          domain.WarehouseStatusActive,
			CreatedAt:       baseTime.Add(10 * time.Minute),
			UpdatedAt:       baseTime.Add(10 * time.Minute),
		},
		{
			ID:              "wh-hcm-qh",
			Code:            "WH-HCM-QH",
			Name:            "QC Hold Warehouse HCM",
			Type:            domain.WarehouseTypeQuarantine,
			SiteCode:        "HCM",
			Address:         "QC quarantine area",
			AllowSaleIssue:  false,
			AllowProdIssue:  false,
			AllowQuarantine: true,
			Status:          domain.WarehouseStatusActive,
			CreatedAt:       baseTime.Add(20 * time.Minute),
			UpdatedAt:       baseTime.Add(20 * time.Minute),
		},
		{
			ID:              "wh-hcm-def",
			Code:            "WH-HCM-DEF",
			Name:            "Defect Warehouse HCM",
			Type:            domain.WarehouseTypeDefect,
			SiteCode:        "HCM",
			Address:         "Defect and scrap area",
			AllowSaleIssue:  false,
			AllowProdIssue:  false,
			AllowQuarantine: false,
			Status:          domain.WarehouseStatusInactive,
			CreatedAt:       baseTime.Add(30 * time.Minute),
			UpdatedAt:       baseTime.Add(30 * time.Minute),
		},
	} {
		warehouse, err := domain.NewWarehouse(input)
		if err == nil {
			warehouses = append(warehouses, warehouse)
		}
	}

	return warehouses
}

func prototypeLocations() []domain.Location {
	baseTime := time.Date(2026, 4, 26, 10, 0, 0, 0, time.UTC)
	locations := make([]domain.Location, 0, 5)
	for _, input := range []domain.NewLocationInput{
		{
			ID:            "loc-hcm-fg-recv-01",
			WarehouseID:   "wh-hcm-fg",
			WarehouseCode: "WH-HCM-FG",
			Code:          "FG-RECV-01",
			Name:          "Finished Goods Receiving Dock",
			Type:          domain.LocationTypeReceiving,
			ZoneCode:      "RECV",
			AllowReceive:  true,
			AllowPick:     false,
			AllowStore:    true,
			IsDefault:     true,
			Status:        domain.LocationStatusActive,
			CreatedAt:     baseTime,
			UpdatedAt:     baseTime,
		},
		{
			ID:            "loc-hcm-fg-pick-a01",
			WarehouseID:   "wh-hcm-fg",
			WarehouseCode: "WH-HCM-FG",
			Code:          "FG-PICK-A01",
			Name:          "Finished Goods Pick A01",
			Type:          domain.LocationTypePick,
			ZoneCode:      "PICK",
			AllowReceive:  false,
			AllowPick:     true,
			AllowStore:    true,
			Status:        domain.LocationStatusActive,
			CreatedAt:     baseTime.Add(10 * time.Minute),
			UpdatedAt:     baseTime.Add(10 * time.Minute),
		},
		{
			ID:            "loc-hcm-rm-recv-01",
			WarehouseID:   "wh-hcm-rm",
			WarehouseCode: "WH-HCM-RM",
			Code:          "RM-RECV-01",
			Name:          "Raw Material Receiving Dock",
			Type:          domain.LocationTypeReceiving,
			ZoneCode:      "RECV",
			AllowReceive:  true,
			AllowPick:     false,
			AllowStore:    true,
			IsDefault:     true,
			Status:        domain.LocationStatusActive,
			CreatedAt:     baseTime.Add(20 * time.Minute),
			UpdatedAt:     baseTime.Add(20 * time.Minute),
		},
		{
			ID:            "loc-hcm-qh-hold-01",
			WarehouseID:   "wh-hcm-qh",
			WarehouseCode: "WH-HCM-QH",
			Code:          "QH-HOLD-01",
			Name:          "QC Hold Bay 01",
			Type:          domain.LocationTypeQCHold,
			ZoneCode:      "QC",
			AllowReceive:  true,
			AllowPick:     false,
			AllowStore:    true,
			IsDefault:     true,
			Status:        domain.LocationStatusActive,
			CreatedAt:     baseTime.Add(30 * time.Minute),
			UpdatedAt:     baseTime.Add(30 * time.Minute),
		},
		{
			ID:            "loc-hcm-def-scrap-01",
			WarehouseID:   "wh-hcm-def",
			WarehouseCode: "WH-HCM-DEF",
			Code:          "DEF-SCRAP-01",
			Name:          "Defect Scrap Bay",
			Type:          domain.LocationTypeScrap,
			ZoneCode:      "SCRAP",
			AllowReceive:  true,
			AllowPick:     false,
			AllowStore:    true,
			Status:        domain.LocationStatusInactive,
			CreatedAt:     baseTime.Add(40 * time.Minute),
			UpdatedAt:     baseTime.Add(40 * time.Minute),
		},
	} {
		location, err := domain.NewLocation(input)
		if err == nil {
			locations = append(locations, location)
		}
	}

	return locations
}
