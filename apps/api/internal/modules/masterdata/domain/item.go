package domain

import (
	"errors"
	"sort"
	"strings"
	"time"

	"github.com/Chinsusu/ERP-v2/apps/api/internal/shared/decimal"
)

var ErrItemRequiredField = errors.New("item required field is missing")
var ErrItemInvalidType = errors.New("item type is invalid")
var ErrItemInvalidStatus = errors.New("item status is invalid")
var ErrItemInvalidShelfLife = errors.New("item shelf life is invalid")
var ErrItemInvalidCost = errors.New("item standard cost is invalid")

type ItemType string

const ItemTypeRawMaterial ItemType = "raw_material"
const ItemTypePackaging ItemType = "packaging"
const ItemTypeSemiFinished ItemType = "semi_finished"
const ItemTypeFinishedGood ItemType = "finished_good"
const ItemTypeService ItemType = "service"

type ItemStatus string

const ItemStatusDraft ItemStatus = "draft"
const ItemStatusActive ItemStatus = "active"
const ItemStatusInactive ItemStatus = "inactive"
const ItemStatusObsolete ItemStatus = "obsolete"

type Item struct {
	ID               string
	ItemCode         string
	SKUCode          string
	Name             string
	Type             ItemType
	Group            string
	BrandCode        string
	UOMBase          string
	UOMPurchase      string
	UOMIssue         string
	LotControlled    bool
	ExpiryControlled bool
	ShelfLifeDays    int
	QCRequired       bool
	Status           ItemStatus
	StandardCost     decimal.Decimal
	IsSellable       bool
	IsPurchasable    bool
	IsProducible     bool
	SpecVersion      string
	CreatedAt        time.Time
	UpdatedAt        time.Time
}

type NewItemInput struct {
	ID               string
	ItemCode         string
	SKUCode          string
	Name             string
	Type             ItemType
	Group            string
	BrandCode        string
	UOMBase          string
	UOMPurchase      string
	UOMIssue         string
	LotControlled    bool
	ExpiryControlled bool
	ShelfLifeDays    int
	QCRequired       bool
	Status           ItemStatus
	StandardCost     decimal.Decimal
	IsSellable       bool
	IsPurchasable    bool
	IsProducible     bool
	SpecVersion      string
	CreatedAt        time.Time
	UpdatedAt        time.Time
}

type UpdateItemInput struct {
	ItemCode         string
	SKUCode          string
	Name             string
	Type             ItemType
	Group            string
	BrandCode        string
	UOMBase          string
	UOMPurchase      string
	UOMIssue         string
	LotControlled    bool
	ExpiryControlled bool
	ShelfLifeDays    int
	QCRequired       bool
	Status           ItemStatus
	StandardCost     decimal.Decimal
	IsSellable       bool
	IsPurchasable    bool
	IsProducible     bool
	SpecVersion      string
	UpdatedAt        time.Time
}

type ItemFilter struct {
	Search   string
	Status   ItemStatus
	Type     ItemType
	Page     int
	PageSize int
}

func NewItem(input NewItemInput) (Item, error) {
	createdAt := input.CreatedAt
	if createdAt.IsZero() {
		createdAt = time.Now().UTC()
	}
	updatedAt := input.UpdatedAt
	if updatedAt.IsZero() {
		updatedAt = createdAt
	}
	status := NormalizeItemStatus(input.Status)
	if status == "" {
		status = ItemStatusDraft
	}

	standardCost, err := decimal.ParseUnitCost(input.StandardCost.String())
	if err != nil {
		return Item{}, ErrItemInvalidCost
	}

	item := Item{
		ID:               strings.TrimSpace(input.ID),
		ItemCode:         NormalizeItemCode(input.ItemCode),
		SKUCode:          NormalizeSKUCode(input.SKUCode),
		Name:             strings.TrimSpace(input.Name),
		Type:             NormalizeItemType(input.Type),
		Group:            strings.TrimSpace(input.Group),
		BrandCode:        strings.ToUpper(strings.TrimSpace(input.BrandCode)),
		UOMBase:          NormalizeUOM(input.UOMBase),
		UOMPurchase:      NormalizeUOM(input.UOMPurchase),
		UOMIssue:         NormalizeUOM(input.UOMIssue),
		LotControlled:    input.LotControlled,
		ExpiryControlled: input.ExpiryControlled,
		ShelfLifeDays:    input.ShelfLifeDays,
		QCRequired:       input.QCRequired,
		Status:           status,
		StandardCost:     standardCost,
		IsSellable:       input.IsSellable,
		IsPurchasable:    input.IsPurchasable,
		IsProducible:     input.IsProducible,
		SpecVersion:      strings.TrimSpace(input.SpecVersion),
		CreatedAt:        createdAt.UTC(),
		UpdatedAt:        updatedAt.UTC(),
	}
	if item.UOMPurchase == "" {
		item.UOMPurchase = item.UOMBase
	}
	if item.UOMIssue == "" {
		item.UOMIssue = item.UOMBase
	}

	if err := item.Validate(); err != nil {
		return Item{}, err
	}

	return item, nil
}

func (i Item) Update(input UpdateItemInput) (Item, error) {
	updatedAt := input.UpdatedAt
	if updatedAt.IsZero() {
		updatedAt = time.Now().UTC()
	}
	status := NormalizeItemStatus(input.Status)
	if status == "" {
		status = i.Status
	}

	standardCost, err := decimal.ParseUnitCost(input.StandardCost.String())
	if err != nil {
		return Item{}, ErrItemInvalidCost
	}

	updated := i.Clone()
	updated.ItemCode = NormalizeItemCode(input.ItemCode)
	updated.SKUCode = NormalizeSKUCode(input.SKUCode)
	updated.Name = strings.TrimSpace(input.Name)
	updated.Type = NormalizeItemType(input.Type)
	updated.Group = strings.TrimSpace(input.Group)
	updated.BrandCode = strings.ToUpper(strings.TrimSpace(input.BrandCode))
	updated.UOMBase = NormalizeUOM(input.UOMBase)
	updated.UOMPurchase = NormalizeUOM(input.UOMPurchase)
	updated.UOMIssue = NormalizeUOM(input.UOMIssue)
	updated.LotControlled = input.LotControlled
	updated.ExpiryControlled = input.ExpiryControlled
	updated.ShelfLifeDays = input.ShelfLifeDays
	updated.QCRequired = input.QCRequired
	updated.Status = status
	updated.StandardCost = standardCost
	updated.IsSellable = input.IsSellable
	updated.IsPurchasable = input.IsPurchasable
	updated.IsProducible = input.IsProducible
	updated.SpecVersion = strings.TrimSpace(input.SpecVersion)
	updated.UpdatedAt = updatedAt.UTC()
	if updated.UOMPurchase == "" {
		updated.UOMPurchase = updated.UOMBase
	}
	if updated.UOMIssue == "" {
		updated.UOMIssue = updated.UOMBase
	}

	if err := updated.Validate(); err != nil {
		return Item{}, err
	}

	return updated, nil
}

func (i Item) ChangeStatus(status ItemStatus, updatedAt time.Time) (Item, error) {
	status = NormalizeItemStatus(status)
	if !IsValidItemStatus(status) {
		return Item{}, ErrItemInvalidStatus
	}
	if updatedAt.IsZero() {
		updatedAt = time.Now().UTC()
	}

	updated := i.Clone()
	updated.Status = status
	updated.UpdatedAt = updatedAt.UTC()

	return updated, nil
}

func (i Item) Validate() error {
	if strings.TrimSpace(i.ID) == "" ||
		strings.TrimSpace(i.ItemCode) == "" ||
		strings.TrimSpace(i.SKUCode) == "" ||
		strings.TrimSpace(i.Name) == "" ||
		strings.TrimSpace(i.UOMBase) == "" {
		return ErrItemRequiredField
	}
	if !IsValidItemType(i.Type) {
		return ErrItemInvalidType
	}
	if !IsValidItemStatus(i.Status) {
		return ErrItemInvalidStatus
	}
	if i.ExpiryControlled && i.ShelfLifeDays <= 0 {
		return ErrItemInvalidShelfLife
	}
	if i.ShelfLifeDays < 0 {
		return ErrItemInvalidShelfLife
	}
	if i.StandardCost.IsNegative() {
		return ErrItemInvalidCost
	}

	return nil
}

func (i Item) Clone() Item {
	return i
}

func NewItemFilter(search string, status ItemStatus, itemType ItemType, page int, pageSize int) ItemFilter {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}
	if pageSize > 100 {
		pageSize = 100
	}

	return ItemFilter{
		Search:   strings.ToLower(strings.TrimSpace(search)),
		Status:   NormalizeItemStatus(status),
		Type:     NormalizeItemType(itemType),
		Page:     page,
		PageSize: pageSize,
	}
}

func (f ItemFilter) Matches(item Item) bool {
	if f.Status != "" && item.Status != f.Status {
		return false
	}
	if f.Type != "" && item.Type != f.Type {
		return false
	}
	if f.Search == "" {
		return true
	}

	candidates := []string{
		item.ItemCode,
		item.SKUCode,
		item.Name,
		item.Group,
		item.BrandCode,
	}
	for _, candidate := range candidates {
		if strings.Contains(strings.ToLower(candidate), f.Search) {
			return true
		}
	}

	return false
}

func SortItems(items []Item) {
	sort.Slice(items, func(i int, j int) bool {
		left := items[i]
		right := items[j]
		if left.Status != right.Status {
			return itemStatusRank(left.Status) < itemStatusRank(right.Status)
		}
		if left.SKUCode != right.SKUCode {
			return left.SKUCode < right.SKUCode
		}

		return left.ItemCode < right.ItemCode
	})
}

func NormalizeItemCode(value string) string {
	return strings.ToUpper(strings.TrimSpace(value))
}

func NormalizeSKUCode(value string) string {
	return strings.ToUpper(strings.TrimSpace(value))
}

func NormalizeUOM(value string) string {
	return strings.ToUpper(strings.TrimSpace(value))
}

func NormalizeItemType(value ItemType) ItemType {
	return ItemType(strings.ToLower(strings.TrimSpace(string(value))))
}

func NormalizeItemStatus(value ItemStatus) ItemStatus {
	return ItemStatus(strings.ToLower(strings.TrimSpace(string(value))))
}

func IsValidItemType(value ItemType) bool {
	switch NormalizeItemType(value) {
	case ItemTypeRawMaterial, ItemTypePackaging, ItemTypeSemiFinished, ItemTypeFinishedGood, ItemTypeService:
		return true
	default:
		return false
	}
}

func IsValidItemStatus(value ItemStatus) bool {
	switch NormalizeItemStatus(value) {
	case ItemStatusDraft, ItemStatusActive, ItemStatusInactive, ItemStatusObsolete:
		return true
	default:
		return false
	}
}

func itemStatusRank(status ItemStatus) int {
	switch status {
	case ItemStatusActive:
		return 0
	case ItemStatusDraft:
		return 1
	case ItemStatusInactive:
		return 2
	case ItemStatusObsolete:
		return 3
	default:
		return 4
	}
}
