package domain

import (
	"errors"
	"sort"
	"strings"
	"time"
)

var ErrWarehouseRequiredField = errors.New("warehouse required field is missing")
var ErrWarehouseInvalidType = errors.New("warehouse type is invalid")
var ErrWarehouseInvalidStatus = errors.New("warehouse status is invalid")
var ErrLocationRequiredField = errors.New("warehouse location required field is missing")
var ErrLocationInvalidType = errors.New("warehouse location type is invalid")
var ErrLocationInvalidStatus = errors.New("warehouse location status is invalid")

type WarehouseType string

const WarehouseTypeRawMaterial WarehouseType = "raw_material"
const WarehouseTypePackaging WarehouseType = "packaging"
const WarehouseTypeSemiFinished WarehouseType = "semi_finished"
const WarehouseTypeFinishedGood WarehouseType = "finished_good"
const WarehouseTypeQuarantine WarehouseType = "quarantine"
const WarehouseTypeSample WarehouseType = "sample"
const WarehouseTypeDefect WarehouseType = "defect"
const WarehouseTypeRetailStore WarehouseType = "retail_store"

type WarehouseStatus string

const WarehouseStatusActive WarehouseStatus = "active"
const WarehouseStatusInactive WarehouseStatus = "inactive"

type LocationType string

const LocationTypeReceiving LocationType = "receiving"
const LocationTypeQCHold LocationType = "qc_hold"
const LocationTypeStorage LocationType = "storage"
const LocationTypePick LocationType = "pick"
const LocationTypePack LocationType = "pack"
const LocationTypeHandover LocationType = "handover"
const LocationTypeReturn LocationType = "return"
const LocationTypeLab LocationType = "lab"
const LocationTypeScrap LocationType = "scrap"

type LocationStatus string

const LocationStatusActive LocationStatus = "active"
const LocationStatusInactive LocationStatus = "inactive"

type Warehouse struct {
	ID              string
	Code            string
	Name            string
	Type            WarehouseType
	SiteCode        string
	Address         string
	AllowSaleIssue  bool
	AllowProdIssue  bool
	AllowQuarantine bool
	Status          WarehouseStatus
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

type NewWarehouseInput struct {
	ID              string
	Code            string
	Name            string
	Type            WarehouseType
	SiteCode        string
	Address         string
	AllowSaleIssue  bool
	AllowProdIssue  bool
	AllowQuarantine bool
	Status          WarehouseStatus
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

type UpdateWarehouseInput struct {
	Code            string
	Name            string
	Type            WarehouseType
	SiteCode        string
	Address         string
	AllowSaleIssue  bool
	AllowProdIssue  bool
	AllowQuarantine bool
	Status          WarehouseStatus
	UpdatedAt       time.Time
}

type WarehouseFilter struct {
	Search   string
	Status   WarehouseStatus
	Type     WarehouseType
	Page     int
	PageSize int
}

type Location struct {
	ID            string
	WarehouseID   string
	WarehouseCode string
	Code          string
	Name          string
	Type          LocationType
	ZoneCode      string
	AllowReceive  bool
	AllowPick     bool
	AllowStore    bool
	IsDefault     bool
	Status        LocationStatus
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

type NewLocationInput struct {
	ID            string
	WarehouseID   string
	WarehouseCode string
	Code          string
	Name          string
	Type          LocationType
	ZoneCode      string
	AllowReceive  bool
	AllowPick     bool
	AllowStore    bool
	IsDefault     bool
	Status        LocationStatus
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

type UpdateLocationInput struct {
	WarehouseID   string
	WarehouseCode string
	Code          string
	Name          string
	Type          LocationType
	ZoneCode      string
	AllowReceive  bool
	AllowPick     bool
	AllowStore    bool
	IsDefault     bool
	Status        LocationStatus
	UpdatedAt     time.Time
}

type LocationFilter struct {
	Search      string
	WarehouseID string
	Status      LocationStatus
	Type        LocationType
	Page        int
	PageSize    int
}

func NewWarehouse(input NewWarehouseInput) (Warehouse, error) {
	createdAt := input.CreatedAt
	if createdAt.IsZero() {
		createdAt = time.Now().UTC()
	}
	updatedAt := input.UpdatedAt
	if updatedAt.IsZero() {
		updatedAt = createdAt
	}
	status := NormalizeWarehouseStatus(input.Status)
	if status == "" {
		status = WarehouseStatusActive
	}

	warehouse := Warehouse{
		ID:              strings.TrimSpace(input.ID),
		Code:            NormalizeWarehouseCode(input.Code),
		Name:            strings.TrimSpace(input.Name),
		Type:            NormalizeWarehouseType(input.Type),
		SiteCode:        NormalizeWarehouseCode(input.SiteCode),
		Address:         strings.TrimSpace(input.Address),
		AllowSaleIssue:  input.AllowSaleIssue,
		AllowProdIssue:  input.AllowProdIssue,
		AllowQuarantine: input.AllowQuarantine,
		Status:          status,
		CreatedAt:       createdAt.UTC(),
		UpdatedAt:       updatedAt.UTC(),
	}
	if err := warehouse.Validate(); err != nil {
		return Warehouse{}, err
	}

	return warehouse, nil
}

func (w Warehouse) Update(input UpdateWarehouseInput) (Warehouse, error) {
	updatedAt := input.UpdatedAt
	if updatedAt.IsZero() {
		updatedAt = time.Now().UTC()
	}
	status := NormalizeWarehouseStatus(input.Status)
	if status == "" {
		status = w.Status
	}

	updated := w.Clone()
	updated.Code = NormalizeWarehouseCode(input.Code)
	updated.Name = strings.TrimSpace(input.Name)
	updated.Type = NormalizeWarehouseType(input.Type)
	updated.SiteCode = NormalizeWarehouseCode(input.SiteCode)
	updated.Address = strings.TrimSpace(input.Address)
	updated.AllowSaleIssue = input.AllowSaleIssue
	updated.AllowProdIssue = input.AllowProdIssue
	updated.AllowQuarantine = input.AllowQuarantine
	updated.Status = status
	updated.UpdatedAt = updatedAt.UTC()
	if err := updated.Validate(); err != nil {
		return Warehouse{}, err
	}

	return updated, nil
}

func (w Warehouse) ChangeStatus(status WarehouseStatus, updatedAt time.Time) (Warehouse, error) {
	status = NormalizeWarehouseStatus(status)
	if !IsValidWarehouseStatus(status) {
		return Warehouse{}, ErrWarehouseInvalidStatus
	}
	if updatedAt.IsZero() {
		updatedAt = time.Now().UTC()
	}

	updated := w.Clone()
	updated.Status = status
	updated.UpdatedAt = updatedAt.UTC()

	return updated, nil
}

func (w Warehouse) Validate() error {
	if strings.TrimSpace(w.ID) == "" ||
		strings.TrimSpace(w.Code) == "" ||
		strings.TrimSpace(w.Name) == "" ||
		strings.TrimSpace(w.SiteCode) == "" {
		return ErrWarehouseRequiredField
	}
	if !IsValidWarehouseType(w.Type) {
		return ErrWarehouseInvalidType
	}
	if !IsValidWarehouseStatus(w.Status) {
		return ErrWarehouseInvalidStatus
	}

	return nil
}

func (w Warehouse) Clone() Warehouse {
	return w
}

func NewLocation(input NewLocationInput) (Location, error) {
	createdAt := input.CreatedAt
	if createdAt.IsZero() {
		createdAt = time.Now().UTC()
	}
	updatedAt := input.UpdatedAt
	if updatedAt.IsZero() {
		updatedAt = createdAt
	}
	status := NormalizeLocationStatus(input.Status)
	if status == "" {
		status = LocationStatusActive
	}

	location := Location{
		ID:            strings.TrimSpace(input.ID),
		WarehouseID:   strings.TrimSpace(input.WarehouseID),
		WarehouseCode: NormalizeWarehouseCode(input.WarehouseCode),
		Code:          NormalizeLocationCode(input.Code),
		Name:          strings.TrimSpace(input.Name),
		Type:          NormalizeLocationType(input.Type),
		ZoneCode:      NormalizeLocationCode(input.ZoneCode),
		AllowReceive:  input.AllowReceive,
		AllowPick:     input.AllowPick,
		AllowStore:    input.AllowStore,
		IsDefault:     input.IsDefault,
		Status:        status,
		CreatedAt:     createdAt.UTC(),
		UpdatedAt:     updatedAt.UTC(),
	}
	if err := location.Validate(); err != nil {
		return Location{}, err
	}

	return location, nil
}

func (l Location) Update(input UpdateLocationInput) (Location, error) {
	updatedAt := input.UpdatedAt
	if updatedAt.IsZero() {
		updatedAt = time.Now().UTC()
	}
	status := NormalizeLocationStatus(input.Status)
	if status == "" {
		status = l.Status
	}

	updated := l.Clone()
	updated.WarehouseID = strings.TrimSpace(input.WarehouseID)
	updated.WarehouseCode = NormalizeWarehouseCode(input.WarehouseCode)
	updated.Code = NormalizeLocationCode(input.Code)
	updated.Name = strings.TrimSpace(input.Name)
	updated.Type = NormalizeLocationType(input.Type)
	updated.ZoneCode = NormalizeLocationCode(input.ZoneCode)
	updated.AllowReceive = input.AllowReceive
	updated.AllowPick = input.AllowPick
	updated.AllowStore = input.AllowStore
	updated.IsDefault = input.IsDefault
	updated.Status = status
	updated.UpdatedAt = updatedAt.UTC()
	if err := updated.Validate(); err != nil {
		return Location{}, err
	}

	return updated, nil
}

func (l Location) ChangeStatus(status LocationStatus, updatedAt time.Time) (Location, error) {
	status = NormalizeLocationStatus(status)
	if !IsValidLocationStatus(status) {
		return Location{}, ErrLocationInvalidStatus
	}
	if updatedAt.IsZero() {
		updatedAt = time.Now().UTC()
	}

	updated := l.Clone()
	updated.Status = status
	updated.UpdatedAt = updatedAt.UTC()

	return updated, nil
}

func (l Location) Validate() error {
	if strings.TrimSpace(l.ID) == "" ||
		strings.TrimSpace(l.WarehouseID) == "" ||
		strings.TrimSpace(l.WarehouseCode) == "" ||
		strings.TrimSpace(l.Code) == "" ||
		strings.TrimSpace(l.Name) == "" {
		return ErrLocationRequiredField
	}
	if !IsValidLocationType(l.Type) {
		return ErrLocationInvalidType
	}
	if !IsValidLocationStatus(l.Status) {
		return ErrLocationInvalidStatus
	}

	return nil
}

func (l Location) Clone() Location {
	return l
}

func NewWarehouseFilter(search string, status WarehouseStatus, warehouseType WarehouseType, page int, pageSize int) WarehouseFilter {
	page, pageSize = normalizePage(page, pageSize)

	return WarehouseFilter{
		Search:   strings.ToLower(strings.TrimSpace(search)),
		Status:   NormalizeWarehouseStatus(status),
		Type:     NormalizeWarehouseType(warehouseType),
		Page:     page,
		PageSize: pageSize,
	}
}

func (f WarehouseFilter) Matches(warehouse Warehouse) bool {
	if f.Status != "" && warehouse.Status != f.Status {
		return false
	}
	if f.Type != "" && warehouse.Type != f.Type {
		return false
	}
	if f.Search == "" {
		return true
	}

	candidates := []string{
		warehouse.Code,
		warehouse.Name,
		warehouse.SiteCode,
		warehouse.Address,
	}
	for _, candidate := range candidates {
		if strings.Contains(strings.ToLower(candidate), f.Search) {
			return true
		}
	}

	return false
}

func NewLocationFilter(search string, warehouseID string, status LocationStatus, locationType LocationType, page int, pageSize int) LocationFilter {
	page, pageSize = normalizePage(page, pageSize)

	return LocationFilter{
		Search:      strings.ToLower(strings.TrimSpace(search)),
		WarehouseID: strings.TrimSpace(warehouseID),
		Status:      NormalizeLocationStatus(status),
		Type:        NormalizeLocationType(locationType),
		Page:        page,
		PageSize:    pageSize,
	}
}

func (f LocationFilter) Matches(location Location) bool {
	if f.WarehouseID != "" && location.WarehouseID != f.WarehouseID {
		return false
	}
	if f.Status != "" && location.Status != f.Status {
		return false
	}
	if f.Type != "" && location.Type != f.Type {
		return false
	}
	if f.Search == "" {
		return true
	}

	candidates := []string{
		location.WarehouseCode,
		location.Code,
		location.Name,
		location.ZoneCode,
	}
	for _, candidate := range candidates {
		if strings.Contains(strings.ToLower(candidate), f.Search) {
			return true
		}
	}

	return false
}

func SortWarehouses(warehouses []Warehouse) {
	sort.Slice(warehouses, func(i int, j int) bool {
		left := warehouses[i]
		right := warehouses[j]
		if left.Status != right.Status {
			return warehouseStatusRank(left.Status) < warehouseStatusRank(right.Status)
		}
		if left.SiteCode != right.SiteCode {
			return left.SiteCode < right.SiteCode
		}

		return left.Code < right.Code
	})
}

func SortLocations(locations []Location) {
	sort.Slice(locations, func(i int, j int) bool {
		left := locations[i]
		right := locations[j]
		if left.Status != right.Status {
			return locationStatusRank(left.Status) < locationStatusRank(right.Status)
		}
		if left.WarehouseCode != right.WarehouseCode {
			return left.WarehouseCode < right.WarehouseCode
		}
		if left.ZoneCode != right.ZoneCode {
			return left.ZoneCode < right.ZoneCode
		}

		return left.Code < right.Code
	})
}

func NormalizeWarehouseCode(value string) string {
	return strings.ToUpper(strings.TrimSpace(value))
}

func NormalizeLocationCode(value string) string {
	return strings.ToUpper(strings.TrimSpace(value))
}

func NormalizeWarehouseType(value WarehouseType) WarehouseType {
	return WarehouseType(strings.ToLower(strings.TrimSpace(string(value))))
}

func NormalizeWarehouseStatus(value WarehouseStatus) WarehouseStatus {
	return WarehouseStatus(strings.ToLower(strings.TrimSpace(string(value))))
}

func NormalizeLocationType(value LocationType) LocationType {
	return LocationType(strings.ToLower(strings.TrimSpace(string(value))))
}

func NormalizeLocationStatus(value LocationStatus) LocationStatus {
	return LocationStatus(strings.ToLower(strings.TrimSpace(string(value))))
}

func IsValidWarehouseType(value WarehouseType) bool {
	switch NormalizeWarehouseType(value) {
	case WarehouseTypeRawMaterial,
		WarehouseTypePackaging,
		WarehouseTypeSemiFinished,
		WarehouseTypeFinishedGood,
		WarehouseTypeQuarantine,
		WarehouseTypeSample,
		WarehouseTypeDefect,
		WarehouseTypeRetailStore:
		return true
	default:
		return false
	}
}

func IsValidWarehouseStatus(value WarehouseStatus) bool {
	switch NormalizeWarehouseStatus(value) {
	case WarehouseStatusActive, WarehouseStatusInactive:
		return true
	default:
		return false
	}
}

func IsValidLocationType(value LocationType) bool {
	switch NormalizeLocationType(value) {
	case LocationTypeReceiving,
		LocationTypeQCHold,
		LocationTypeStorage,
		LocationTypePick,
		LocationTypePack,
		LocationTypeHandover,
		LocationTypeReturn,
		LocationTypeLab,
		LocationTypeScrap:
		return true
	default:
		return false
	}
}

func IsValidLocationStatus(value LocationStatus) bool {
	switch NormalizeLocationStatus(value) {
	case LocationStatusActive, LocationStatusInactive:
		return true
	default:
		return false
	}
}

func normalizePage(page int, pageSize int) (int, int) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}
	if pageSize > 100 {
		pageSize = 100
	}

	return page, pageSize
}

func warehouseStatusRank(status WarehouseStatus) int {
	switch status {
	case WarehouseStatusActive:
		return 0
	case WarehouseStatusInactive:
		return 1
	default:
		return 2
	}
}

func locationStatusRank(status LocationStatus) int {
	switch status {
	case LocationStatusActive:
		return 0
	case LocationStatusInactive:
		return 1
	default:
		return 2
	}
}
