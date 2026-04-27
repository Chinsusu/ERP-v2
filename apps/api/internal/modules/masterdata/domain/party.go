package domain

import (
	"errors"
	"sort"
	"strings"
	"time"

	"github.com/Chinsusu/ERP-v2/apps/api/internal/shared/decimal"
)

var ErrSupplierRequiredField = errors.New("supplier required field is missing")
var ErrSupplierInvalidGroup = errors.New("supplier group is invalid")
var ErrSupplierInvalidStatus = errors.New("supplier status is invalid")
var ErrSupplierInvalidMetric = errors.New("supplier metric is invalid")
var ErrSupplierInvalidStatusTransition = errors.New("supplier status transition is invalid")
var ErrCustomerRequiredField = errors.New("customer required field is missing")
var ErrCustomerInvalidType = errors.New("customer type is invalid")
var ErrCustomerInvalidStatus = errors.New("customer status is invalid")
var ErrCustomerInvalidCreditLimit = errors.New("customer credit limit is invalid")
var ErrCustomerInvalidStatusTransition = errors.New("customer status transition is invalid")

type SupplierGroup string

const SupplierGroupRawMaterial SupplierGroup = "raw_material"
const SupplierGroupPackaging SupplierGroup = "packaging"
const SupplierGroupService SupplierGroup = "service"
const SupplierGroupLogistics SupplierGroup = "logistics"
const SupplierGroupOutsource SupplierGroup = "outsource"

type SupplierStatus string

const SupplierStatusDraft SupplierStatus = "draft"
const SupplierStatusActive SupplierStatus = "active"
const SupplierStatusInactive SupplierStatus = "inactive"
const SupplierStatusBlacklisted SupplierStatus = "blacklisted"

type CustomerType string

const CustomerTypeDistributor CustomerType = "distributor"
const CustomerTypeDealer CustomerType = "dealer"
const CustomerTypeRetailCustomer CustomerType = "retail_customer"
const CustomerTypeMarketplace CustomerType = "marketplace"
const CustomerTypeInternalStore CustomerType = "internal_store"

type CustomerStatus string

const CustomerStatusDraft CustomerStatus = "draft"
const CustomerStatusActive CustomerStatus = "active"
const CustomerStatusInactive CustomerStatus = "inactive"
const CustomerStatusBlocked CustomerStatus = "blocked"

type Supplier struct {
	ID            string
	Code          string
	Name          string
	Group         SupplierGroup
	ContactName   string
	Phone         string
	Email         string
	TaxCode       string
	Address       string
	PaymentTerms  string
	LeadTimeDays  int
	MOQ           decimal.Decimal
	QualityScore  decimal.Decimal
	DeliveryScore decimal.Decimal
	Status        SupplierStatus
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

type NewSupplierInput struct {
	ID            string
	Code          string
	Name          string
	Group         SupplierGroup
	ContactName   string
	Phone         string
	Email         string
	TaxCode       string
	Address       string
	PaymentTerms  string
	LeadTimeDays  int
	MOQ           decimal.Decimal
	QualityScore  decimal.Decimal
	DeliveryScore decimal.Decimal
	Status        SupplierStatus
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

type UpdateSupplierInput struct {
	Code          string
	Name          string
	Group         SupplierGroup
	ContactName   string
	Phone         string
	Email         string
	TaxCode       string
	Address       string
	PaymentTerms  string
	LeadTimeDays  int
	MOQ           decimal.Decimal
	QualityScore  decimal.Decimal
	DeliveryScore decimal.Decimal
	Status        SupplierStatus
	UpdatedAt     time.Time
}

type SupplierFilter struct {
	Search   string
	Status   SupplierStatus
	Group    SupplierGroup
	Page     int
	PageSize int
}

type Customer struct {
	ID            string
	Code          string
	Name          string
	Type          CustomerType
	ChannelCode   string
	PriceListCode string
	DiscountGroup string
	CreditLimit   decimal.Decimal
	PaymentTerms  string
	ContactName   string
	Phone         string
	Email         string
	TaxCode       string
	Address       string
	Status        CustomerStatus
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

type NewCustomerInput struct {
	ID            string
	Code          string
	Name          string
	Type          CustomerType
	ChannelCode   string
	PriceListCode string
	DiscountGroup string
	CreditLimit   decimal.Decimal
	PaymentTerms  string
	ContactName   string
	Phone         string
	Email         string
	TaxCode       string
	Address       string
	Status        CustomerStatus
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

type UpdateCustomerInput struct {
	Code          string
	Name          string
	Type          CustomerType
	ChannelCode   string
	PriceListCode string
	DiscountGroup string
	CreditLimit   decimal.Decimal
	PaymentTerms  string
	ContactName   string
	Phone         string
	Email         string
	TaxCode       string
	Address       string
	Status        CustomerStatus
	UpdatedAt     time.Time
}

type CustomerFilter struct {
	Search   string
	Status   CustomerStatus
	Type     CustomerType
	Page     int
	PageSize int
}

func NewSupplier(input NewSupplierInput) (Supplier, error) {
	createdAt := input.CreatedAt
	if createdAt.IsZero() {
		createdAt = time.Now().UTC()
	}
	updatedAt := input.UpdatedAt
	if updatedAt.IsZero() {
		updatedAt = createdAt
	}
	status := NormalizeSupplierStatus(input.Status)
	if status == "" {
		status = SupplierStatusDraft
	}

	moq, err := decimal.ParseQuantity(input.MOQ.String())
	if err != nil {
		return Supplier{}, ErrSupplierInvalidMetric
	}
	qualityScore, err := decimal.ParseRate(input.QualityScore.String())
	if err != nil {
		return Supplier{}, ErrSupplierInvalidMetric
	}
	deliveryScore, err := decimal.ParseRate(input.DeliveryScore.String())
	if err != nil {
		return Supplier{}, ErrSupplierInvalidMetric
	}

	supplier := Supplier{
		ID:            strings.TrimSpace(input.ID),
		Code:          NormalizePartyCode(input.Code),
		Name:          strings.TrimSpace(input.Name),
		Group:         NormalizeSupplierGroup(input.Group),
		ContactName:   strings.TrimSpace(input.ContactName),
		Phone:         strings.TrimSpace(input.Phone),
		Email:         strings.ToLower(strings.TrimSpace(input.Email)),
		TaxCode:       NormalizePartyCode(input.TaxCode),
		Address:       strings.TrimSpace(input.Address),
		PaymentTerms:  NormalizePartyCode(input.PaymentTerms),
		LeadTimeDays:  input.LeadTimeDays,
		MOQ:           moq,
		QualityScore:  qualityScore,
		DeliveryScore: deliveryScore,
		Status:        status,
		CreatedAt:     createdAt.UTC(),
		UpdatedAt:     updatedAt.UTC(),
	}
	if err := supplier.Validate(); err != nil {
		return Supplier{}, err
	}

	return supplier, nil
}

func (s Supplier) Update(input UpdateSupplierInput) (Supplier, error) {
	updatedAt := input.UpdatedAt
	if updatedAt.IsZero() {
		updatedAt = time.Now().UTC()
	}
	status := NormalizeSupplierStatus(input.Status)
	if status == "" {
		status = s.Status
	}
	if !IsValidSupplierStatusTransition(s.Status, status) {
		return Supplier{}, ErrSupplierInvalidStatusTransition
	}

	moq, err := decimal.ParseQuantity(input.MOQ.String())
	if err != nil {
		return Supplier{}, ErrSupplierInvalidMetric
	}
	qualityScore, err := decimal.ParseRate(input.QualityScore.String())
	if err != nil {
		return Supplier{}, ErrSupplierInvalidMetric
	}
	deliveryScore, err := decimal.ParseRate(input.DeliveryScore.String())
	if err != nil {
		return Supplier{}, ErrSupplierInvalidMetric
	}

	updated := s.Clone()
	updated.Code = NormalizePartyCode(input.Code)
	updated.Name = strings.TrimSpace(input.Name)
	updated.Group = NormalizeSupplierGroup(input.Group)
	updated.ContactName = strings.TrimSpace(input.ContactName)
	updated.Phone = strings.TrimSpace(input.Phone)
	updated.Email = strings.ToLower(strings.TrimSpace(input.Email))
	updated.TaxCode = NormalizePartyCode(input.TaxCode)
	updated.Address = strings.TrimSpace(input.Address)
	updated.PaymentTerms = NormalizePartyCode(input.PaymentTerms)
	updated.LeadTimeDays = input.LeadTimeDays
	updated.MOQ = moq
	updated.QualityScore = qualityScore
	updated.DeliveryScore = deliveryScore
	updated.Status = status
	updated.UpdatedAt = updatedAt.UTC()
	if err := updated.Validate(); err != nil {
		return Supplier{}, err
	}

	return updated, nil
}

func (s Supplier) ChangeStatus(status SupplierStatus, updatedAt time.Time) (Supplier, error) {
	status = NormalizeSupplierStatus(status)
	if !IsValidSupplierStatus(status) {
		return Supplier{}, ErrSupplierInvalidStatus
	}
	if !IsValidSupplierStatusTransition(s.Status, status) {
		return Supplier{}, ErrSupplierInvalidStatusTransition
	}
	if updatedAt.IsZero() {
		updatedAt = time.Now().UTC()
	}

	updated := s.Clone()
	updated.Status = status
	updated.UpdatedAt = updatedAt.UTC()

	return updated, nil
}

func (s Supplier) Validate() error {
	if strings.TrimSpace(s.ID) == "" ||
		strings.TrimSpace(s.Code) == "" ||
		strings.TrimSpace(s.Name) == "" {
		return ErrSupplierRequiredField
	}
	if !IsValidSupplierGroup(s.Group) {
		return ErrSupplierInvalidGroup
	}
	if !IsValidSupplierStatus(s.Status) {
		return ErrSupplierInvalidStatus
	}
	if s.LeadTimeDays < 0 || s.MOQ.IsNegative() || s.QualityScore.IsNegative() || s.DeliveryScore.IsNegative() {
		return ErrSupplierInvalidMetric
	}

	return nil
}

func (s Supplier) Clone() Supplier {
	return s
}

func NewCustomer(input NewCustomerInput) (Customer, error) {
	createdAt := input.CreatedAt
	if createdAt.IsZero() {
		createdAt = time.Now().UTC()
	}
	updatedAt := input.UpdatedAt
	if updatedAt.IsZero() {
		updatedAt = createdAt
	}
	status := NormalizeCustomerStatus(input.Status)
	if status == "" {
		status = CustomerStatusDraft
	}

	creditLimit, err := decimal.ParseMoneyAmount(input.CreditLimit.String())
	if err != nil {
		return Customer{}, ErrCustomerInvalidCreditLimit
	}

	customer := Customer{
		ID:            strings.TrimSpace(input.ID),
		Code:          NormalizePartyCode(input.Code),
		Name:          strings.TrimSpace(input.Name),
		Type:          NormalizeCustomerType(input.Type),
		ChannelCode:   NormalizePartyCode(input.ChannelCode),
		PriceListCode: NormalizePartyCode(input.PriceListCode),
		DiscountGroup: strings.TrimSpace(input.DiscountGroup),
		CreditLimit:   creditLimit,
		PaymentTerms:  NormalizePartyCode(input.PaymentTerms),
		ContactName:   strings.TrimSpace(input.ContactName),
		Phone:         strings.TrimSpace(input.Phone),
		Email:         strings.ToLower(strings.TrimSpace(input.Email)),
		TaxCode:       NormalizePartyCode(input.TaxCode),
		Address:       strings.TrimSpace(input.Address),
		Status:        status,
		CreatedAt:     createdAt.UTC(),
		UpdatedAt:     updatedAt.UTC(),
	}
	if err := customer.Validate(); err != nil {
		return Customer{}, err
	}

	return customer, nil
}

func (c Customer) Update(input UpdateCustomerInput) (Customer, error) {
	updatedAt := input.UpdatedAt
	if updatedAt.IsZero() {
		updatedAt = time.Now().UTC()
	}
	status := NormalizeCustomerStatus(input.Status)
	if status == "" {
		status = c.Status
	}
	if !IsValidCustomerStatusTransition(c.Status, status) {
		return Customer{}, ErrCustomerInvalidStatusTransition
	}

	creditLimit, err := decimal.ParseMoneyAmount(input.CreditLimit.String())
	if err != nil {
		return Customer{}, ErrCustomerInvalidCreditLimit
	}

	updated := c.Clone()
	updated.Code = NormalizePartyCode(input.Code)
	updated.Name = strings.TrimSpace(input.Name)
	updated.Type = NormalizeCustomerType(input.Type)
	updated.ChannelCode = NormalizePartyCode(input.ChannelCode)
	updated.PriceListCode = NormalizePartyCode(input.PriceListCode)
	updated.DiscountGroup = strings.TrimSpace(input.DiscountGroup)
	updated.CreditLimit = creditLimit
	updated.PaymentTerms = NormalizePartyCode(input.PaymentTerms)
	updated.ContactName = strings.TrimSpace(input.ContactName)
	updated.Phone = strings.TrimSpace(input.Phone)
	updated.Email = strings.ToLower(strings.TrimSpace(input.Email))
	updated.TaxCode = NormalizePartyCode(input.TaxCode)
	updated.Address = strings.TrimSpace(input.Address)
	updated.Status = status
	updated.UpdatedAt = updatedAt.UTC()
	if err := updated.Validate(); err != nil {
		return Customer{}, err
	}

	return updated, nil
}

func (c Customer) ChangeStatus(status CustomerStatus, updatedAt time.Time) (Customer, error) {
	status = NormalizeCustomerStatus(status)
	if !IsValidCustomerStatus(status) {
		return Customer{}, ErrCustomerInvalidStatus
	}
	if !IsValidCustomerStatusTransition(c.Status, status) {
		return Customer{}, ErrCustomerInvalidStatusTransition
	}
	if updatedAt.IsZero() {
		updatedAt = time.Now().UTC()
	}

	updated := c.Clone()
	updated.Status = status
	updated.UpdatedAt = updatedAt.UTC()

	return updated, nil
}

func (c Customer) Validate() error {
	if strings.TrimSpace(c.ID) == "" ||
		strings.TrimSpace(c.Code) == "" ||
		strings.TrimSpace(c.Name) == "" {
		return ErrCustomerRequiredField
	}
	if !IsValidCustomerType(c.Type) {
		return ErrCustomerInvalidType
	}
	if !IsValidCustomerStatus(c.Status) {
		return ErrCustomerInvalidStatus
	}
	if c.CreditLimit.IsNegative() {
		return ErrCustomerInvalidCreditLimit
	}

	return nil
}

func (c Customer) Clone() Customer {
	return c
}

func NewSupplierFilter(search string, status SupplierStatus, group SupplierGroup, page int, pageSize int) SupplierFilter {
	page, pageSize = normalizePage(page, pageSize)

	return SupplierFilter{
		Search:   strings.ToLower(strings.TrimSpace(search)),
		Status:   NormalizeSupplierStatus(status),
		Group:    NormalizeSupplierGroup(group),
		Page:     page,
		PageSize: pageSize,
	}
}

func (f SupplierFilter) Matches(supplier Supplier) bool {
	if f.Status != "" && supplier.Status != f.Status {
		return false
	}
	if f.Group != "" && supplier.Group != f.Group {
		return false
	}
	if f.Search == "" {
		return true
	}

	candidates := []string{
		supplier.Code,
		supplier.Name,
		supplier.Group.String(),
		supplier.ContactName,
		supplier.Phone,
		supplier.Email,
		supplier.TaxCode,
		supplier.Address,
		supplier.PaymentTerms,
	}
	for _, candidate := range candidates {
		if strings.Contains(strings.ToLower(candidate), f.Search) {
			return true
		}
	}

	return false
}

func NewCustomerFilter(search string, status CustomerStatus, customerType CustomerType, page int, pageSize int) CustomerFilter {
	page, pageSize = normalizePage(page, pageSize)

	return CustomerFilter{
		Search:   strings.ToLower(strings.TrimSpace(search)),
		Status:   NormalizeCustomerStatus(status),
		Type:     NormalizeCustomerType(customerType),
		Page:     page,
		PageSize: pageSize,
	}
}

func (f CustomerFilter) Matches(customer Customer) bool {
	if f.Status != "" && customer.Status != f.Status {
		return false
	}
	if f.Type != "" && customer.Type != f.Type {
		return false
	}
	if f.Search == "" {
		return true
	}

	candidates := []string{
		customer.Code,
		customer.Name,
		customer.Type.String(),
		customer.ChannelCode,
		customer.PriceListCode,
		customer.DiscountGroup,
		customer.PaymentTerms,
		customer.ContactName,
		customer.Phone,
		customer.Email,
		customer.TaxCode,
		customer.Address,
	}
	for _, candidate := range candidates {
		if strings.Contains(strings.ToLower(candidate), f.Search) {
			return true
		}
	}

	return false
}

func SortSuppliers(suppliers []Supplier) {
	sort.Slice(suppliers, func(i int, j int) bool {
		left := suppliers[i]
		right := suppliers[j]
		if left.Status != right.Status {
			return supplierStatusRank(left.Status) < supplierStatusRank(right.Status)
		}
		if left.Group != right.Group {
			return left.Group < right.Group
		}

		return left.Code < right.Code
	})
}

func SortCustomers(customers []Customer) {
	sort.Slice(customers, func(i int, j int) bool {
		left := customers[i]
		right := customers[j]
		if left.Status != right.Status {
			return customerStatusRank(left.Status) < customerStatusRank(right.Status)
		}
		if left.Type != right.Type {
			return left.Type < right.Type
		}

		return left.Code < right.Code
	})
}

func NormalizePartyCode(value string) string {
	return strings.ToUpper(strings.TrimSpace(value))
}

func NormalizeSupplierGroup(value SupplierGroup) SupplierGroup {
	return SupplierGroup(strings.ToLower(strings.TrimSpace(string(value))))
}

func NormalizeSupplierStatus(value SupplierStatus) SupplierStatus {
	return SupplierStatus(strings.ToLower(strings.TrimSpace(string(value))))
}

func NormalizeCustomerType(value CustomerType) CustomerType {
	return CustomerType(strings.ToLower(strings.TrimSpace(string(value))))
}

func NormalizeCustomerStatus(value CustomerStatus) CustomerStatus {
	return CustomerStatus(strings.ToLower(strings.TrimSpace(string(value))))
}

func IsValidSupplierGroup(value SupplierGroup) bool {
	switch NormalizeSupplierGroup(value) {
	case SupplierGroupRawMaterial, SupplierGroupPackaging, SupplierGroupService, SupplierGroupLogistics, SupplierGroupOutsource:
		return true
	default:
		return false
	}
}

func IsValidSupplierStatus(value SupplierStatus) bool {
	switch NormalizeSupplierStatus(value) {
	case SupplierStatusDraft, SupplierStatusActive, SupplierStatusInactive, SupplierStatusBlacklisted:
		return true
	default:
		return false
	}
}

func IsValidSupplierStatusTransition(from SupplierStatus, to SupplierStatus) bool {
	from = NormalizeSupplierStatus(from)
	to = NormalizeSupplierStatus(to)
	if from == to {
		return IsValidSupplierStatus(to)
	}
	if !IsValidSupplierStatus(from) || !IsValidSupplierStatus(to) {
		return false
	}
	if from == SupplierStatusBlacklisted && to == SupplierStatusActive {
		return false
	}

	return true
}

func IsValidCustomerType(value CustomerType) bool {
	switch NormalizeCustomerType(value) {
	case CustomerTypeDistributor, CustomerTypeDealer, CustomerTypeRetailCustomer, CustomerTypeMarketplace, CustomerTypeInternalStore:
		return true
	default:
		return false
	}
}

func IsValidCustomerStatus(value CustomerStatus) bool {
	switch NormalizeCustomerStatus(value) {
	case CustomerStatusDraft, CustomerStatusActive, CustomerStatusInactive, CustomerStatusBlocked:
		return true
	default:
		return false
	}
}

func IsValidCustomerStatusTransition(from CustomerStatus, to CustomerStatus) bool {
	from = NormalizeCustomerStatus(from)
	to = NormalizeCustomerStatus(to)
	if from == to {
		return IsValidCustomerStatus(to)
	}
	if !IsValidCustomerStatus(from) || !IsValidCustomerStatus(to) {
		return false
	}
	if from == CustomerStatusBlocked && to == CustomerStatusActive {
		return false
	}

	return true
}

func (g SupplierGroup) String() string {
	return string(g)
}

func (t CustomerType) String() string {
	return string(t)
}

func supplierStatusRank(status SupplierStatus) int {
	switch status {
	case SupplierStatusActive:
		return 0
	case SupplierStatusDraft:
		return 1
	case SupplierStatusInactive:
		return 2
	case SupplierStatusBlacklisted:
		return 3
	default:
		return 4
	}
}

func customerStatusRank(status CustomerStatus) int {
	switch status {
	case CustomerStatusActive:
		return 0
	case CustomerStatusDraft:
		return 1
	case CustomerStatusInactive:
		return 2
	case CustomerStatusBlocked:
		return 3
	default:
		return 4
	}
}
