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
	"github.com/Chinsusu/ERP-v2/apps/api/internal/shared/decimal"
	"github.com/Chinsusu/ERP-v2/apps/api/internal/shared/response"
)

var ErrSupplierNotFound = errors.New("supplier not found")
var ErrCustomerNotFound = errors.New("customer not found")
var ErrDuplicateSupplierCode = errors.New("supplier code already exists")
var ErrDuplicateCustomerCode = errors.New("customer code already exists")

type PartyCatalog struct {
	mu        sync.RWMutex
	suppliers map[string]domain.Supplier
	customers map[string]domain.Customer
	auditLog  audit.LogStore
	clock     func() time.Time
}

type CreateSupplierInput struct {
	Code          string
	Name          string
	Group         string
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
	Status        string
	ActorID       string
	RequestID     string
}

type UpdateSupplierInput struct {
	ID            string
	Code          string
	Name          string
	Group         string
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
	Status        string
	ActorID       string
	RequestID     string
}

type ChangeSupplierStatusInput struct {
	ID        string
	Status    string
	ActorID   string
	RequestID string
}

type CreateCustomerInput struct {
	Code          string
	Name          string
	Type          string
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
	Status        string
	ActorID       string
	RequestID     string
}

type UpdateCustomerInput struct {
	ID            string
	Code          string
	Name          string
	Type          string
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
	Status        string
	ActorID       string
	RequestID     string
}

type ChangeCustomerStatusInput struct {
	ID        string
	Status    string
	ActorID   string
	RequestID string
}

type SupplierResult struct {
	Supplier   domain.Supplier
	AuditLogID string
}

type CustomerResult struct {
	Customer   domain.Customer
	AuditLogID string
}

func NewPrototypePartyCatalog(auditLog audit.LogStore) *PartyCatalog {
	store := &PartyCatalog{
		suppliers: make(map[string]domain.Supplier),
		customers: make(map[string]domain.Customer),
		auditLog:  auditLog,
		clock:     func() time.Time { return time.Now().UTC() },
	}
	for _, supplier := range prototypeSuppliers() {
		store.suppliers[supplier.ID] = supplier.Clone()
	}
	for _, customer := range prototypeCustomers() {
		store.customers[customer.ID] = customer.Clone()
	}

	return store
}

func NewPrototypePartyCatalogAt(auditLog audit.LogStore, now time.Time) *PartyCatalog {
	store := NewPrototypePartyCatalog(auditLog)
	store.clock = func() time.Time { return now.UTC() }

	return store
}

func (s *PartyCatalog) ListSuppliers(_ context.Context, filter domain.SupplierFilter) ([]domain.Supplier, response.Pagination, error) {
	if s == nil {
		return nil, response.Pagination{}, errors.New("party catalog is required")
	}
	if filter.Status != "" && !domain.IsValidSupplierStatus(filter.Status) {
		return nil, response.Pagination{}, domain.ErrSupplierInvalidStatus
	}
	if filter.Group != "" && !domain.IsValidSupplierGroup(filter.Group) {
		return nil, response.Pagination{}, domain.ErrSupplierInvalidGroup
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	rows := make([]domain.Supplier, 0, len(s.suppliers))
	for _, supplier := range s.suppliers {
		if filter.Matches(supplier) {
			rows = append(rows, supplier.Clone())
		}
	}
	domain.SortSuppliers(rows)
	pageRows, pagination := paginateSuppliers(rows, filter.Page, filter.PageSize)

	return pageRows, pagination, nil
}

func (s *PartyCatalog) GetSupplier(_ context.Context, id string) (domain.Supplier, error) {
	if s == nil {
		return domain.Supplier{}, errors.New("party catalog is required")
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	supplier, ok := s.suppliers[strings.TrimSpace(id)]
	if !ok {
		return domain.Supplier{}, ErrSupplierNotFound
	}

	return supplier.Clone(), nil
}

func (s *PartyCatalog) CreateSupplier(ctx context.Context, input CreateSupplierInput) (SupplierResult, error) {
	if s == nil {
		return SupplierResult{}, errors.New("party catalog is required")
	}
	if s.auditLog == nil {
		return SupplierResult{}, errors.New("audit log store is required")
	}

	now := s.clock()
	supplier, err := domain.NewSupplier(domain.NewSupplierInput{
		ID:            newSupplierID(input.Code, now),
		Code:          input.Code,
		Name:          input.Name,
		Group:         domain.SupplierGroup(input.Group),
		ContactName:   input.ContactName,
		Phone:         input.Phone,
		Email:         input.Email,
		TaxCode:       input.TaxCode,
		Address:       input.Address,
		PaymentTerms:  input.PaymentTerms,
		LeadTimeDays:  input.LeadTimeDays,
		MOQ:           input.MOQ,
		QualityScore:  input.QualityScore,
		DeliveryScore: input.DeliveryScore,
		Status:        domain.SupplierStatus(input.Status),
		CreatedAt:     now,
		UpdatedAt:     now,
	})
	if err != nil {
		return SupplierResult{}, err
	}

	s.mu.Lock()
	if err := s.ensureUniqueSupplierLocked(supplier, ""); err != nil {
		s.mu.Unlock()
		return SupplierResult{}, err
	}
	s.suppliers[supplier.ID] = supplier.Clone()
	s.mu.Unlock()

	log, err := newSupplierAuditLog(input.ActorID, input.RequestID, "masterdata.supplier.created", supplier, nil, supplierToAuditMap(supplier), now)
	if err != nil {
		return SupplierResult{}, err
	}
	if err := s.auditLog.Record(ctx, log); err != nil {
		return SupplierResult{}, err
	}

	return SupplierResult{Supplier: supplier, AuditLogID: log.ID}, nil
}

func (s *PartyCatalog) UpdateSupplier(ctx context.Context, input UpdateSupplierInput) (SupplierResult, error) {
	if s == nil {
		return SupplierResult{}, errors.New("party catalog is required")
	}
	if s.auditLog == nil {
		return SupplierResult{}, errors.New("audit log store is required")
	}

	now := s.clock()
	s.mu.Lock()
	current, ok := s.suppliers[strings.TrimSpace(input.ID)]
	if !ok {
		s.mu.Unlock()
		return SupplierResult{}, ErrSupplierNotFound
	}
	updated, err := current.Update(domain.UpdateSupplierInput{
		Code:          input.Code,
		Name:          input.Name,
		Group:         domain.SupplierGroup(input.Group),
		ContactName:   input.ContactName,
		Phone:         input.Phone,
		Email:         input.Email,
		TaxCode:       input.TaxCode,
		Address:       input.Address,
		PaymentTerms:  input.PaymentTerms,
		LeadTimeDays:  input.LeadTimeDays,
		MOQ:           input.MOQ,
		QualityScore:  input.QualityScore,
		DeliveryScore: input.DeliveryScore,
		Status:        domain.SupplierStatus(input.Status),
		UpdatedAt:     now,
	})
	if err != nil {
		s.mu.Unlock()
		return SupplierResult{}, err
	}
	if err := s.ensureUniqueSupplierLocked(updated, current.ID); err != nil {
		s.mu.Unlock()
		return SupplierResult{}, err
	}
	s.suppliers[current.ID] = updated.Clone()
	s.mu.Unlock()

	log, err := newSupplierAuditLog(input.ActorID, input.RequestID, "masterdata.supplier.updated", updated, supplierToAuditMap(current), supplierToAuditMap(updated), now)
	if err != nil {
		return SupplierResult{}, err
	}
	if err := s.auditLog.Record(ctx, log); err != nil {
		return SupplierResult{}, err
	}

	return SupplierResult{Supplier: updated, AuditLogID: log.ID}, nil
}

func (s *PartyCatalog) ChangeSupplierStatus(ctx context.Context, input ChangeSupplierStatusInput) (SupplierResult, error) {
	if s == nil {
		return SupplierResult{}, errors.New("party catalog is required")
	}
	if s.auditLog == nil {
		return SupplierResult{}, errors.New("audit log store is required")
	}

	now := s.clock()
	s.mu.Lock()
	current, ok := s.suppliers[strings.TrimSpace(input.ID)]
	if !ok {
		s.mu.Unlock()
		return SupplierResult{}, ErrSupplierNotFound
	}
	updated, err := current.ChangeStatus(domain.SupplierStatus(input.Status), now)
	if err != nil {
		s.mu.Unlock()
		return SupplierResult{}, err
	}
	s.suppliers[current.ID] = updated.Clone()
	s.mu.Unlock()

	log, err := newSupplierAuditLog(
		input.ActorID,
		input.RequestID,
		"masterdata.supplier.status_changed",
		updated,
		map[string]any{"status": string(current.Status)},
		map[string]any{"status": string(updated.Status)},
		now,
	)
	if err != nil {
		return SupplierResult{}, err
	}
	if err := s.auditLog.Record(ctx, log); err != nil {
		return SupplierResult{}, err
	}

	return SupplierResult{Supplier: updated, AuditLogID: log.ID}, nil
}

func (s *PartyCatalog) ListCustomers(_ context.Context, filter domain.CustomerFilter) ([]domain.Customer, response.Pagination, error) {
	if s == nil {
		return nil, response.Pagination{}, errors.New("party catalog is required")
	}
	if filter.Status != "" && !domain.IsValidCustomerStatus(filter.Status) {
		return nil, response.Pagination{}, domain.ErrCustomerInvalidStatus
	}
	if filter.Type != "" && !domain.IsValidCustomerType(filter.Type) {
		return nil, response.Pagination{}, domain.ErrCustomerInvalidType
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	rows := make([]domain.Customer, 0, len(s.customers))
	for _, customer := range s.customers {
		if filter.Matches(customer) {
			rows = append(rows, customer.Clone())
		}
	}
	domain.SortCustomers(rows)
	pageRows, pagination := paginateCustomers(rows, filter.Page, filter.PageSize)

	return pageRows, pagination, nil
}

func (s *PartyCatalog) GetCustomer(_ context.Context, id string) (domain.Customer, error) {
	if s == nil {
		return domain.Customer{}, errors.New("party catalog is required")
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	customer, ok := s.customers[strings.TrimSpace(id)]
	if !ok {
		return domain.Customer{}, ErrCustomerNotFound
	}

	return customer.Clone(), nil
}

func (s *PartyCatalog) CreateCustomer(ctx context.Context, input CreateCustomerInput) (CustomerResult, error) {
	if s == nil {
		return CustomerResult{}, errors.New("party catalog is required")
	}
	if s.auditLog == nil {
		return CustomerResult{}, errors.New("audit log store is required")
	}

	now := s.clock()
	customer, err := domain.NewCustomer(domain.NewCustomerInput{
		ID:            newCustomerID(input.Code, now),
		Code:          input.Code,
		Name:          input.Name,
		Type:          domain.CustomerType(input.Type),
		ChannelCode:   input.ChannelCode,
		PriceListCode: input.PriceListCode,
		DiscountGroup: input.DiscountGroup,
		CreditLimit:   input.CreditLimit,
		PaymentTerms:  input.PaymentTerms,
		ContactName:   input.ContactName,
		Phone:         input.Phone,
		Email:         input.Email,
		TaxCode:       input.TaxCode,
		Address:       input.Address,
		Status:        domain.CustomerStatus(input.Status),
		CreatedAt:     now,
		UpdatedAt:     now,
	})
	if err != nil {
		return CustomerResult{}, err
	}

	s.mu.Lock()
	if err := s.ensureUniqueCustomerLocked(customer, ""); err != nil {
		s.mu.Unlock()
		return CustomerResult{}, err
	}
	s.customers[customer.ID] = customer.Clone()
	s.mu.Unlock()

	log, err := newCustomerAuditLog(input.ActorID, input.RequestID, "masterdata.customer.created", customer, nil, customerToAuditMap(customer), now)
	if err != nil {
		return CustomerResult{}, err
	}
	if err := s.auditLog.Record(ctx, log); err != nil {
		return CustomerResult{}, err
	}

	return CustomerResult{Customer: customer, AuditLogID: log.ID}, nil
}

func (s *PartyCatalog) UpdateCustomer(ctx context.Context, input UpdateCustomerInput) (CustomerResult, error) {
	if s == nil {
		return CustomerResult{}, errors.New("party catalog is required")
	}
	if s.auditLog == nil {
		return CustomerResult{}, errors.New("audit log store is required")
	}

	now := s.clock()
	s.mu.Lock()
	current, ok := s.customers[strings.TrimSpace(input.ID)]
	if !ok {
		s.mu.Unlock()
		return CustomerResult{}, ErrCustomerNotFound
	}
	updated, err := current.Update(domain.UpdateCustomerInput{
		Code:          input.Code,
		Name:          input.Name,
		Type:          domain.CustomerType(input.Type),
		ChannelCode:   input.ChannelCode,
		PriceListCode: input.PriceListCode,
		DiscountGroup: input.DiscountGroup,
		CreditLimit:   input.CreditLimit,
		PaymentTerms:  input.PaymentTerms,
		ContactName:   input.ContactName,
		Phone:         input.Phone,
		Email:         input.Email,
		TaxCode:       input.TaxCode,
		Address:       input.Address,
		Status:        domain.CustomerStatus(input.Status),
		UpdatedAt:     now,
	})
	if err != nil {
		s.mu.Unlock()
		return CustomerResult{}, err
	}
	if err := s.ensureUniqueCustomerLocked(updated, current.ID); err != nil {
		s.mu.Unlock()
		return CustomerResult{}, err
	}
	s.customers[current.ID] = updated.Clone()
	s.mu.Unlock()

	log, err := newCustomerAuditLog(input.ActorID, input.RequestID, "masterdata.customer.updated", updated, customerToAuditMap(current), customerToAuditMap(updated), now)
	if err != nil {
		return CustomerResult{}, err
	}
	if err := s.auditLog.Record(ctx, log); err != nil {
		return CustomerResult{}, err
	}

	return CustomerResult{Customer: updated, AuditLogID: log.ID}, nil
}

func (s *PartyCatalog) ChangeCustomerStatus(ctx context.Context, input ChangeCustomerStatusInput) (CustomerResult, error) {
	if s == nil {
		return CustomerResult{}, errors.New("party catalog is required")
	}
	if s.auditLog == nil {
		return CustomerResult{}, errors.New("audit log store is required")
	}

	now := s.clock()
	s.mu.Lock()
	current, ok := s.customers[strings.TrimSpace(input.ID)]
	if !ok {
		s.mu.Unlock()
		return CustomerResult{}, ErrCustomerNotFound
	}
	updated, err := current.ChangeStatus(domain.CustomerStatus(input.Status), now)
	if err != nil {
		s.mu.Unlock()
		return CustomerResult{}, err
	}
	s.customers[current.ID] = updated.Clone()
	s.mu.Unlock()

	log, err := newCustomerAuditLog(
		input.ActorID,
		input.RequestID,
		"masterdata.customer.status_changed",
		updated,
		map[string]any{"status": string(current.Status)},
		map[string]any{"status": string(updated.Status)},
		now,
	)
	if err != nil {
		return CustomerResult{}, err
	}
	if err := s.auditLog.Record(ctx, log); err != nil {
		return CustomerResult{}, err
	}

	return CustomerResult{Customer: updated, AuditLogID: log.ID}, nil
}

func (s *PartyCatalog) ensureUniqueSupplierLocked(supplier domain.Supplier, currentID string) error {
	for _, existing := range s.suppliers {
		if strings.TrimSpace(currentID) != "" && existing.ID == currentID {
			continue
		}
		if existing.Code == supplier.Code {
			return ErrDuplicateSupplierCode
		}
	}

	return nil
}

func (s *PartyCatalog) ensureUniqueCustomerLocked(customer domain.Customer, currentID string) error {
	for _, existing := range s.customers {
		if strings.TrimSpace(currentID) != "" && existing.ID == currentID {
			continue
		}
		if existing.Code == customer.Code {
			return ErrDuplicateCustomerCode
		}
	}

	return nil
}

func paginateSuppliers(suppliers []domain.Supplier, page int, pageSize int) ([]domain.Supplier, response.Pagination) {
	totalItems := len(suppliers)
	totalPages := 0
	if totalItems > 0 {
		totalPages = (totalItems + pageSize - 1) / pageSize
	}
	start := (page - 1) * pageSize
	if start >= totalItems {
		return []domain.Supplier{}, response.Pagination{
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

	return append([]domain.Supplier(nil), suppliers[start:end]...), response.Pagination{
		Page:       page,
		PageSize:   pageSize,
		TotalItems: totalItems,
		TotalPages: totalPages,
	}
}

func paginateCustomers(customers []domain.Customer, page int, pageSize int) ([]domain.Customer, response.Pagination) {
	totalItems := len(customers)
	totalPages := 0
	if totalItems > 0 {
		totalPages = (totalItems + pageSize - 1) / pageSize
	}
	start := (page - 1) * pageSize
	if start >= totalItems {
		return []domain.Customer{}, response.Pagination{
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

	return append([]domain.Customer(nil), customers[start:end]...), response.Pagination{
		Page:       page,
		PageSize:   pageSize,
		TotalItems: totalItems,
		TotalPages: totalPages,
	}
}

func newSupplierAuditLog(
	actorID string,
	requestID string,
	action string,
	supplier domain.Supplier,
	beforeData map[string]any,
	afterData map[string]any,
	createdAt time.Time,
) (audit.Log, error) {
	return audit.NewLog(audit.NewLogInput{
		OrgID:      "org-my-pham",
		ActorID:    strings.TrimSpace(actorID),
		Action:     strings.TrimSpace(action),
		EntityType: "mdm.supplier",
		EntityID:   supplier.ID,
		RequestID:  strings.TrimSpace(requestID),
		BeforeData: beforeData,
		AfterData:  afterData,
		Metadata: map[string]any{
			"source":         "supplier master data",
			"supplier_code":  supplier.Code,
			"supplier_group": string(supplier.Group),
		},
		CreatedAt: createdAt,
	})
}

func newCustomerAuditLog(
	actorID string,
	requestID string,
	action string,
	customer domain.Customer,
	beforeData map[string]any,
	afterData map[string]any,
	createdAt time.Time,
) (audit.Log, error) {
	return audit.NewLog(audit.NewLogInput{
		OrgID:      "org-my-pham",
		ActorID:    strings.TrimSpace(actorID),
		Action:     strings.TrimSpace(action),
		EntityType: "mdm.customer",
		EntityID:   customer.ID,
		RequestID:  strings.TrimSpace(requestID),
		BeforeData: beforeData,
		AfterData:  afterData,
		Metadata: map[string]any{
			"source":        "customer master data",
			"customer_code": customer.Code,
			"customer_type": string(customer.Type),
		},
		CreatedAt: createdAt,
	})
}

func supplierToAuditMap(supplier domain.Supplier) map[string]any {
	return map[string]any{
		"supplier_code":  supplier.Code,
		"supplier_name":  supplier.Name,
		"supplier_group": string(supplier.Group),
		"contact_name":   supplier.ContactName,
		"phone":          supplier.Phone,
		"email":          supplier.Email,
		"tax_code":       supplier.TaxCode,
		"address":        supplier.Address,
		"payment_terms":  supplier.PaymentTerms,
		"lead_time_days": supplier.LeadTimeDays,
		"moq":            supplier.MOQ,
		"quality_score":  supplier.QualityScore,
		"delivery_score": supplier.DeliveryScore,
		"status":         string(supplier.Status),
	}
}

func customerToAuditMap(customer domain.Customer) map[string]any {
	return map[string]any{
		"customer_code":   customer.Code,
		"customer_name":   customer.Name,
		"customer_type":   string(customer.Type),
		"channel_code":    customer.ChannelCode,
		"price_list_code": customer.PriceListCode,
		"discount_group":  customer.DiscountGroup,
		"credit_limit":    customer.CreditLimit,
		"payment_terms":   customer.PaymentTerms,
		"contact_name":    customer.ContactName,
		"phone":           customer.Phone,
		"email":           customer.Email,
		"tax_code":        customer.TaxCode,
		"address":         customer.Address,
		"status":          string(customer.Status),
	}
}

func newSupplierID(code string, now time.Time) string {
	value := strings.ToLower(domain.NormalizePartyCode(code))
	value = strings.ReplaceAll(value, "-", "_")
	if value == "" {
		value = "supplier"
	}

	return fmt.Sprintf("sup_%s_%d", value, now.UnixNano())
}

func newCustomerID(code string, now time.Time) string {
	value := strings.ToLower(domain.NormalizePartyCode(code))
	value = strings.ReplaceAll(value, "-", "_")
	if value == "" {
		value = "customer"
	}

	return fmt.Sprintf("cus_%s_%d", value, now.UnixNano())
}

func prototypeSuppliers() []domain.Supplier {
	baseTime := time.Date(2026, 4, 26, 11, 0, 0, 0, time.UTC)
	suppliers := make([]domain.Supplier, 0, 4)
	for _, input := range []domain.NewSupplierInput{
		{
			ID:            "sup-rm-bioactive",
			Code:          "SUP-RM-BIO",
			Name:          "BioActive Raw Materials",
			Group:         domain.SupplierGroupRawMaterial,
			ContactName:   "Nguyen Van An",
			Phone:         "+84901234501",
			Email:         "purchasing@bioactive.example",
			TaxCode:       "0312345001",
			Address:       "Binh Duong raw material hub",
			PaymentTerms:  "NET30",
			LeadTimeDays:  12,
			MOQ:           decimal.MustQuantity("50"),
			QualityScore:  decimal.MustRate("94"),
			DeliveryScore: decimal.MustRate("91"),
			Status:        domain.SupplierStatusActive,
			CreatedAt:     baseTime,
			UpdatedAt:     baseTime,
		},
		{
			ID:            "sup-pkg-vina",
			Code:          "SUP-PKG-VINA",
			Name:          "Vina Packaging Solutions",
			Group:         domain.SupplierGroupPackaging,
			ContactName:   "Tran Thi Binh",
			Phone:         "+84901234502",
			Email:         "sales@vinapack.example",
			TaxCode:       "0312345002",
			Address:       "Long An packaging park",
			PaymentTerms:  "NET45",
			LeadTimeDays:  8,
			MOQ:           decimal.MustQuantity("1000"),
			QualityScore:  decimal.MustRate("89"),
			DeliveryScore: decimal.MustRate("88"),
			Status:        domain.SupplierStatusActive,
			CreatedAt:     baseTime.Add(10 * time.Minute),
			UpdatedAt:     baseTime.Add(10 * time.Minute),
		},
		{
			ID:            "sup-log-fastgo",
			Code:          "SUP-LOG-FASTGO",
			Name:          "FastGo Logistics",
			Group:         domain.SupplierGroupLogistics,
			ContactName:   "Le Minh Chau",
			Phone:         "+84901234503",
			Email:         "ops@fastgo.example",
			TaxCode:       "0312345003",
			Address:       "Ho Chi Minh logistics center",
			PaymentTerms:  "NET15",
			LeadTimeDays:  2,
			QualityScore:  decimal.MustRate("87"),
			DeliveryScore: decimal.MustRate("93"),
			Status:        domain.SupplierStatusActive,
			CreatedAt:     baseTime.Add(20 * time.Minute),
			UpdatedAt:     baseTime.Add(20 * time.Minute),
		},
		{
			ID:            "sup-out-lotus",
			Code:          "SUP-OUT-LOTUS",
			Name:          "Lotus Filling Partner",
			Group:         domain.SupplierGroupOutsource,
			ContactName:   "Pham Quoc Duy",
			Phone:         "+84901234504",
			Email:         "qa@lotusfill.example",
			TaxCode:       "0312345004",
			Address:       "Dong Nai outsource site",
			PaymentTerms:  "NET30",
			LeadTimeDays:  15,
			MOQ:           decimal.MustQuantity("500"),
			QualityScore:  decimal.MustRate("82"),
			DeliveryScore: decimal.MustRate("80"),
			Status:        domain.SupplierStatusInactive,
			CreatedAt:     baseTime.Add(30 * time.Minute),
			UpdatedAt:     baseTime.Add(30 * time.Minute),
		},
	} {
		supplier, err := domain.NewSupplier(input)
		if err == nil {
			suppliers = append(suppliers, supplier)
		}
	}

	return suppliers
}

func prototypeCustomers() []domain.Customer {
	baseTime := time.Date(2026, 4, 26, 12, 0, 0, 0, time.UTC)
	customers := make([]domain.Customer, 0, 4)
	for _, input := range []domain.NewCustomerInput{
		{
			ID:            "cus-dl-minh-anh",
			Code:          "CUS-DL-MINHANH",
			Name:          "Minh Anh Distributor",
			Type:          domain.CustomerTypeDistributor,
			ChannelCode:   "B2B",
			PriceListCode: "PL-B2B-2026",
			DiscountGroup: "tier_1",
			CreditLimit:   decimal.MustMoneyAmount("500000000"),
			PaymentTerms:  "NET30",
			ContactName:   "Do Minh Anh",
			Phone:         "+84909888111",
			Email:         "orders@minhanh.example",
			TaxCode:       "0315678001",
			Address:       "District 7, Ho Chi Minh City",
			Status:        domain.CustomerStatusActive,
			CreatedAt:     baseTime,
			UpdatedAt:     baseTime,
		},
		{
			ID:            "cus-dealer-linh-chi",
			Code:          "CUS-DL-LINHCHI",
			Name:          "Linh Chi Dealer",
			Type:          domain.CustomerTypeDealer,
			ChannelCode:   "DEALER",
			PriceListCode: "PL-DEALER-2026",
			DiscountGroup: "tier_2",
			CreditLimit:   decimal.MustMoneyAmount("150000000"),
			PaymentTerms:  "NET15",
			ContactName:   "Nguyen Linh Chi",
			Phone:         "+84909888222",
			Email:         "buyer@linhchi.example",
			TaxCode:       "0315678002",
			Address:       "Thu Duc City",
			Status:        domain.CustomerStatusActive,
			CreatedAt:     baseTime.Add(10 * time.Minute),
			UpdatedAt:     baseTime.Add(10 * time.Minute),
		},
		{
			ID:            "cus-mp-shopee",
			Code:          "CUS-MP-SHOPEE",
			Name:          "Shopee Marketplace",
			Type:          domain.CustomerTypeMarketplace,
			ChannelCode:   "MP",
			PriceListCode: "PL-MP-2026",
			DiscountGroup: "marketplace",
			CreditLimit:   decimal.MustMoneyAmount("0"),
			PaymentTerms:  "PREPAID",
			ContactName:   "Marketplace Ops",
			Phone:         "+84909888333",
			Email:         "ops@marketplace.example",
			TaxCode:       "0315678003",
			Address:       "Marketplace fulfillment channel",
			Status:        domain.CustomerStatusActive,
			CreatedAt:     baseTime.Add(20 * time.Minute),
			UpdatedAt:     baseTime.Add(20 * time.Minute),
		},
		{
			ID:            "cus-internal-hcm-store",
			Code:          "CUS-INT-HCMSTORE",
			Name:          "HCM Internal Store",
			Type:          domain.CustomerTypeInternalStore,
			ChannelCode:   "INT",
			PriceListCode: "PL-INT-2026",
			DiscountGroup: "internal",
			CreditLimit:   decimal.MustMoneyAmount("0"),
			PaymentTerms:  "INTERNAL",
			ContactName:   "Store Lead",
			Phone:         "+84909888444",
			Email:         "store-hcm@example.local",
			TaxCode:       "0315678004",
			Address:       "Ho Chi Minh flagship store",
			Status:        domain.CustomerStatusDraft,
			CreatedAt:     baseTime.Add(30 * time.Minute),
			UpdatedAt:     baseTime.Add(30 * time.Minute),
		},
	} {
		customer, err := domain.NewCustomer(input)
		if err == nil {
			customers = append(customers, customer)
		}
	}

	return customers
}
