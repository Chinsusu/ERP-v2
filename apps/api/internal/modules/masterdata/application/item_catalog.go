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

var ErrItemNotFound = errors.New("item not found")
var ErrDuplicateItemCode = errors.New("item code already exists")
var ErrDuplicateSKUCode = errors.New("sku code already exists")

type ItemCatalog struct {
	mu       sync.RWMutex
	records  map[string]domain.Item
	auditLog audit.LogStore
	clock    func() time.Time
}

type CreateItemInput struct {
	ItemCode         string
	SKUCode          string
	Name             string
	Type             string
	Group            string
	BrandCode        string
	UOMBase          string
	UOMPurchase      string
	UOMIssue         string
	LotControlled    bool
	ExpiryControlled bool
	ShelfLifeDays    int
	QCRequired       bool
	Status           string
	StandardCost     float64
	IsSellable       bool
	IsPurchasable    bool
	IsProducible     bool
	SpecVersion      string
	ActorID          string
	RequestID        string
}

type UpdateItemInput struct {
	ID               string
	ItemCode         string
	SKUCode          string
	Name             string
	Type             string
	Group            string
	BrandCode        string
	UOMBase          string
	UOMPurchase      string
	UOMIssue         string
	LotControlled    bool
	ExpiryControlled bool
	ShelfLifeDays    int
	QCRequired       bool
	Status           string
	StandardCost     float64
	IsSellable       bool
	IsPurchasable    bool
	IsProducible     bool
	SpecVersion      string
	ActorID          string
	RequestID        string
}

type ChangeItemStatusInput struct {
	ID        string
	Status    string
	ActorID   string
	RequestID string
}

type ItemResult struct {
	Item       domain.Item
	AuditLogID string
}

func NewPrototypeItemCatalog(auditLog audit.LogStore) *ItemCatalog {
	store := &ItemCatalog{
		records:  make(map[string]domain.Item),
		auditLog: auditLog,
		clock:    func() time.Time { return time.Now().UTC() },
	}
	for _, item := range prototypeItems() {
		store.records[item.ID] = item.Clone()
	}

	return store
}

func NewPrototypeItemCatalogAt(auditLog audit.LogStore, now time.Time) *ItemCatalog {
	store := NewPrototypeItemCatalog(auditLog)
	store.clock = func() time.Time { return now.UTC() }

	return store
}

func (s *ItemCatalog) List(_ context.Context, filter domain.ItemFilter) ([]domain.Item, response.Pagination, error) {
	if s == nil {
		return nil, response.Pagination{}, errors.New("item catalog is required")
	}
	if filter.Status != "" && !domain.IsValidItemStatus(filter.Status) {
		return nil, response.Pagination{}, domain.ErrItemInvalidStatus
	}
	if filter.Type != "" && !domain.IsValidItemType(filter.Type) {
		return nil, response.Pagination{}, domain.ErrItemInvalidType
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	rows := make([]domain.Item, 0, len(s.records))
	for _, item := range s.records {
		if filter.Matches(item) {
			rows = append(rows, item.Clone())
		}
	}
	domain.SortItems(rows)
	pageRows, pagination := paginateItems(rows, filter.Page, filter.PageSize)

	return pageRows, pagination, nil
}

func (s *ItemCatalog) Get(_ context.Context, id string) (domain.Item, error) {
	if s == nil {
		return domain.Item{}, errors.New("item catalog is required")
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	item, ok := s.records[strings.TrimSpace(id)]
	if !ok {
		return domain.Item{}, ErrItemNotFound
	}

	return item.Clone(), nil
}

func (s *ItemCatalog) Create(ctx context.Context, input CreateItemInput) (ItemResult, error) {
	if s == nil {
		return ItemResult{}, errors.New("item catalog is required")
	}
	if s.auditLog == nil {
		return ItemResult{}, errors.New("audit log store is required")
	}

	now := s.clock()
	item, err := domain.NewItem(domain.NewItemInput{
		ID:               newItemID(input.SKUCode, now),
		ItemCode:         input.ItemCode,
		SKUCode:          input.SKUCode,
		Name:             input.Name,
		Type:             domain.ItemType(input.Type),
		Group:            input.Group,
		BrandCode:        input.BrandCode,
		UOMBase:          input.UOMBase,
		UOMPurchase:      input.UOMPurchase,
		UOMIssue:         input.UOMIssue,
		LotControlled:    input.LotControlled,
		ExpiryControlled: input.ExpiryControlled,
		ShelfLifeDays:    input.ShelfLifeDays,
		QCRequired:       input.QCRequired,
		Status:           domain.ItemStatus(input.Status),
		StandardCost:     input.StandardCost,
		IsSellable:       input.IsSellable,
		IsPurchasable:    input.IsPurchasable,
		IsProducible:     input.IsProducible,
		SpecVersion:      input.SpecVersion,
		CreatedAt:        now,
		UpdatedAt:        now,
	})
	if err != nil {
		return ItemResult{}, err
	}

	s.mu.Lock()
	if err := s.ensureUniqueLocked(item, ""); err != nil {
		s.mu.Unlock()
		return ItemResult{}, err
	}
	s.records[item.ID] = item.Clone()
	s.mu.Unlock()

	log, err := newItemAuditLog(input.ActorID, input.RequestID, "masterdata.item.created", item, nil, itemToAuditMap(item), now)
	if err != nil {
		return ItemResult{}, err
	}
	if err := s.auditLog.Record(ctx, log); err != nil {
		return ItemResult{}, err
	}

	return ItemResult{Item: item, AuditLogID: log.ID}, nil
}

func (s *ItemCatalog) Update(ctx context.Context, input UpdateItemInput) (ItemResult, error) {
	if s == nil {
		return ItemResult{}, errors.New("item catalog is required")
	}
	if s.auditLog == nil {
		return ItemResult{}, errors.New("audit log store is required")
	}

	now := s.clock()
	s.mu.Lock()
	current, ok := s.records[strings.TrimSpace(input.ID)]
	if !ok {
		s.mu.Unlock()
		return ItemResult{}, ErrItemNotFound
	}
	updated, err := current.Update(domain.UpdateItemInput{
		ItemCode:         input.ItemCode,
		SKUCode:          input.SKUCode,
		Name:             input.Name,
		Type:             domain.ItemType(input.Type),
		Group:            input.Group,
		BrandCode:        input.BrandCode,
		UOMBase:          input.UOMBase,
		UOMPurchase:      input.UOMPurchase,
		UOMIssue:         input.UOMIssue,
		LotControlled:    input.LotControlled,
		ExpiryControlled: input.ExpiryControlled,
		ShelfLifeDays:    input.ShelfLifeDays,
		QCRequired:       input.QCRequired,
		Status:           domain.ItemStatus(input.Status),
		StandardCost:     input.StandardCost,
		IsSellable:       input.IsSellable,
		IsPurchasable:    input.IsPurchasable,
		IsProducible:     input.IsProducible,
		SpecVersion:      input.SpecVersion,
		UpdatedAt:        now,
	})
	if err != nil {
		s.mu.Unlock()
		return ItemResult{}, err
	}
	if err := s.ensureUniqueLocked(updated, current.ID); err != nil {
		s.mu.Unlock()
		return ItemResult{}, err
	}
	s.records[current.ID] = updated.Clone()
	s.mu.Unlock()

	log, err := newItemAuditLog(input.ActorID, input.RequestID, "masterdata.item.updated", updated, itemToAuditMap(current), itemToAuditMap(updated), now)
	if err != nil {
		return ItemResult{}, err
	}
	if err := s.auditLog.Record(ctx, log); err != nil {
		return ItemResult{}, err
	}

	return ItemResult{Item: updated, AuditLogID: log.ID}, nil
}

func (s *ItemCatalog) ChangeStatus(ctx context.Context, input ChangeItemStatusInput) (ItemResult, error) {
	if s == nil {
		return ItemResult{}, errors.New("item catalog is required")
	}
	if s.auditLog == nil {
		return ItemResult{}, errors.New("audit log store is required")
	}

	now := s.clock()
	s.mu.Lock()
	current, ok := s.records[strings.TrimSpace(input.ID)]
	if !ok {
		s.mu.Unlock()
		return ItemResult{}, ErrItemNotFound
	}
	updated, err := current.ChangeStatus(domain.ItemStatus(input.Status), now)
	if err != nil {
		s.mu.Unlock()
		return ItemResult{}, err
	}
	s.records[current.ID] = updated.Clone()
	s.mu.Unlock()

	log, err := newItemAuditLog(
		input.ActorID,
		input.RequestID,
		"masterdata.item.status_changed",
		updated,
		map[string]any{"status": string(current.Status)},
		map[string]any{"status": string(updated.Status)},
		now,
	)
	if err != nil {
		return ItemResult{}, err
	}
	if err := s.auditLog.Record(ctx, log); err != nil {
		return ItemResult{}, err
	}

	return ItemResult{Item: updated, AuditLogID: log.ID}, nil
}

func (s *ItemCatalog) ensureUniqueLocked(item domain.Item, currentID string) error {
	for _, existing := range s.records {
		if strings.TrimSpace(currentID) != "" && existing.ID == currentID {
			continue
		}
		if existing.ItemCode == item.ItemCode {
			return ErrDuplicateItemCode
		}
		if existing.SKUCode == item.SKUCode {
			return ErrDuplicateSKUCode
		}
	}

	return nil
}

func paginateItems(items []domain.Item, page int, pageSize int) ([]domain.Item, response.Pagination) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}
	if pageSize > 100 {
		pageSize = 100
	}

	totalItems := len(items)
	totalPages := 0
	if totalItems > 0 {
		totalPages = (totalItems + pageSize - 1) / pageSize
	}
	start := (page - 1) * pageSize
	if start >= totalItems {
		return []domain.Item{}, response.Pagination{
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

	return append([]domain.Item(nil), items[start:end]...), response.Pagination{
		Page:       page,
		PageSize:   pageSize,
		TotalItems: totalItems,
		TotalPages: totalPages,
	}
}

func newItemAuditLog(
	actorID string,
	requestID string,
	action string,
	item domain.Item,
	beforeData map[string]any,
	afterData map[string]any,
	createdAt time.Time,
) (audit.Log, error) {
	return audit.NewLog(audit.NewLogInput{
		OrgID:      "org-my-pham",
		ActorID:    strings.TrimSpace(actorID),
		Action:     strings.TrimSpace(action),
		EntityType: "mdm.item",
		EntityID:   item.ID,
		RequestID:  strings.TrimSpace(requestID),
		BeforeData: beforeData,
		AfterData:  afterData,
		Metadata: map[string]any{
			"source":    "item sku master data",
			"item_code": item.ItemCode,
			"sku_code":  item.SKUCode,
		},
		CreatedAt: createdAt,
	})
}

func itemToAuditMap(item domain.Item) map[string]any {
	return map[string]any{
		"item_code":         item.ItemCode,
		"sku_code":          item.SKUCode,
		"name":              item.Name,
		"item_type":         string(item.Type),
		"item_group":        item.Group,
		"brand_code":        item.BrandCode,
		"uom_base":          item.UOMBase,
		"uom_purchase":      item.UOMPurchase,
		"uom_issue":         item.UOMIssue,
		"lot_controlled":    item.LotControlled,
		"expiry_controlled": item.ExpiryControlled,
		"shelf_life_days":   item.ShelfLifeDays,
		"qc_required":       item.QCRequired,
		"status":            string(item.Status),
		"standard_cost":     item.StandardCost,
		"is_sellable":       item.IsSellable,
		"is_purchasable":    item.IsPurchasable,
		"is_producible":     item.IsProducible,
		"spec_version":      item.SpecVersion,
	}
}

func newItemID(skuCode string, now time.Time) string {
	sku := strings.ToLower(domain.NormalizeSKUCode(skuCode))
	sku = strings.ReplaceAll(sku, "-", "_")
	if sku == "" {
		sku = "item"
	}

	return fmt.Sprintf("item_%s_%d", sku, now.UnixNano())
}

func prototypeItems() []domain.Item {
	baseTime := time.Date(2026, 4, 26, 8, 0, 0, 0, time.UTC)
	items := make([]domain.Item, 0, 3)
	for _, input := range []domain.NewItemInput{
		{
			ID:               "item-serum-30ml",
			ItemCode:         "ITEM-SERUM-HYDRA",
			SKUCode:          "SERUM-30ML",
			Name:             "Hydrating Serum 30ml",
			Type:             domain.ItemTypeFinishedGood,
			Group:            "serum",
			BrandCode:        "MYH",
			UOMBase:          "EA",
			UOMPurchase:      "EA",
			UOMIssue:         "EA",
			LotControlled:    true,
			ExpiryControlled: true,
			ShelfLifeDays:    730,
			QCRequired:       true,
			Status:           domain.ItemStatusActive,
			StandardCost:     64000,
			IsSellable:       true,
			IsPurchasable:    false,
			IsProducible:     true,
			SpecVersion:      "SPEC-SERUM-2026.04",
			CreatedAt:        baseTime,
			UpdatedAt:        baseTime,
		},
		{
			ID:               "item-cream-50g",
			ItemCode:         "ITEM-CREAM-REPAIR",
			SKUCode:          "CREAM-50G",
			Name:             "Repair Cream 50g",
			Type:             domain.ItemTypeFinishedGood,
			Group:            "cream",
			BrandCode:        "MYH",
			UOMBase:          "EA",
			UOMPurchase:      "EA",
			UOMIssue:         "EA",
			LotControlled:    true,
			ExpiryControlled: true,
			ShelfLifeDays:    540,
			QCRequired:       true,
			Status:           domain.ItemStatusActive,
			StandardCost:     58000,
			IsSellable:       true,
			IsPurchasable:    false,
			IsProducible:     true,
			SpecVersion:      "SPEC-CREAM-2026.03",
			CreatedAt:        baseTime.Add(10 * time.Minute),
			UpdatedAt:        baseTime.Add(10 * time.Minute),
		},
		{
			ID:               "item-toner-100ml",
			ItemCode:         "ITEM-TONER-BALANCE",
			SKUCode:          "TONER-100ML",
			Name:             "Balancing Toner 100ml",
			Type:             domain.ItemTypeFinishedGood,
			Group:            "toner",
			BrandCode:        "MYH",
			UOMBase:          "EA",
			UOMPurchase:      "EA",
			UOMIssue:         "EA",
			LotControlled:    true,
			ExpiryControlled: true,
			ShelfLifeDays:    720,
			QCRequired:       true,
			Status:           domain.ItemStatusDraft,
			StandardCost:     42000,
			IsSellable:       true,
			IsPurchasable:    false,
			IsProducible:     true,
			SpecVersion:      "SPEC-TONER-2026.04-DRAFT",
			CreatedAt:        baseTime.Add(20 * time.Minute),
			UpdatedAt:        baseTime.Add(20 * time.Minute),
		},
	} {
		item, err := domain.NewItem(input)
		if err == nil {
			items = append(items, item)
		}
	}

	return items
}
