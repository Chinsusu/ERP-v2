package application

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/Chinsusu/ERP-v2/apps/api/internal/modules/masterdata/domain"
	"github.com/Chinsusu/ERP-v2/apps/api/internal/shared/audit"
	"github.com/Chinsusu/ERP-v2/apps/api/internal/shared/decimal"
	"github.com/Chinsusu/ERP-v2/apps/api/internal/shared/response"
)

type PostgresPartyCatalogConfig struct {
	DefaultOrgID string
	Clock        func() time.Time
}

type PostgresPartyCatalog struct {
	db           *sql.DB
	auditLog     audit.LogStore
	defaultOrgID string
	clock        func() time.Time
}

func NewPostgresPartyCatalog(db *sql.DB, auditLog audit.LogStore, cfg PostgresPartyCatalogConfig) *PostgresPartyCatalog {
	clock := cfg.Clock
	if clock == nil {
		clock = func() time.Time { return time.Now().UTC() }
	}

	return &PostgresPartyCatalog{
		db:           db,
		auditLog:     auditLog,
		defaultOrgID: strings.TrimSpace(cfg.DefaultOrgID),
		clock:        clock,
	}
}

const selectPostgresSuppliersSQL = `
SELECT
  COALESCE(supplier.supplier_ref, supplier.id::text),
  supplier.code,
  supplier.name,
  COALESCE(supplier.supplier_group, 'service'),
  COALESCE(supplier.contact_name, ''),
  COALESCE(supplier.phone, ''),
  COALESCE(supplier.email, ''),
  COALESCE(supplier.tax_code, ''),
  COALESCE(supplier.address, ''),
  COALESCE(supplier.payment_terms, ''),
  supplier.lead_time_days,
  supplier.moq::text,
  supplier.quality_score::text,
  supplier.delivery_score::text,
  CASE WHEN supplier.status = 'blocked' THEN 'blacklisted' ELSE supplier.status END,
  supplier.created_at,
  supplier.updated_at
FROM mdm.suppliers AS supplier
ORDER BY supplier.status, supplier.supplier_group, supplier.code`

const selectPostgresSupplierSQL = `
SELECT
  COALESCE(supplier.supplier_ref, supplier.id::text),
  supplier.code,
  supplier.name,
  COALESCE(supplier.supplier_group, 'service'),
  COALESCE(supplier.contact_name, ''),
  COALESCE(supplier.phone, ''),
  COALESCE(supplier.email, ''),
  COALESCE(supplier.tax_code, ''),
  COALESCE(supplier.address, ''),
  COALESCE(supplier.payment_terms, ''),
  supplier.lead_time_days,
  supplier.moq::text,
  supplier.quality_score::text,
  supplier.delivery_score::text,
  CASE WHEN supplier.status = 'blocked' THEN 'blacklisted' ELSE supplier.status END,
  supplier.created_at,
  supplier.updated_at
FROM mdm.suppliers AS supplier
WHERE lower(COALESCE(supplier.supplier_ref, supplier.id::text)) = lower($1)
   OR supplier.id::text = $1
   OR lower(supplier.code) = lower($1)
LIMIT 1`

const insertPostgresSupplierSQL = `
INSERT INTO mdm.suppliers (
  id,
  org_id,
  supplier_ref,
  code,
  name,
  supplier_type,
  supplier_group,
  contact_name,
  phone,
  email,
  tax_code,
  address,
  payment_terms,
  lead_time_days,
  moq,
  quality_score,
  delivery_score,
  status,
  created_at,
  updated_at
) VALUES (
  COALESCE($1::uuid, gen_random_uuid()),
  $2::uuid,
  $3,
  $4,
  $5,
  $6,
  $7,
  $8,
  $9,
  $10,
  $11,
  $12,
  $13,
  $14,
  $15,
  $16,
  $17,
  $18,
  $19,
  $20
)`

const updatePostgresSupplierSQL = `
UPDATE mdm.suppliers
SET code = $3,
    name = $4,
    supplier_type = $5,
    supplier_group = $6,
    contact_name = $7,
    phone = $8,
    email = $9,
    tax_code = $10,
    address = $11,
    payment_terms = $12,
    lead_time_days = $13,
    moq = $14,
    quality_score = $15,
    delivery_score = $16,
    status = $17,
    updated_at = $18,
    version = version + 1
WHERE org_id = $1::uuid
  AND lower(COALESCE(supplier_ref, id::text)) = lower($2)`

const updatePostgresSupplierStatusSQL = `
UPDATE mdm.suppliers
SET status = $3,
    updated_at = $4,
    version = version + 1
WHERE org_id = $1::uuid
  AND lower(COALESCE(supplier_ref, id::text)) = lower($2)`

const selectPostgresSupplierDuplicateCodeSQL = `
SELECT COALESCE(supplier_ref, id::text)
FROM mdm.suppliers
WHERE org_id = $1::uuid
  AND lower(code) = lower($2)
  AND lower(COALESCE(supplier_ref, id::text)) <> lower($3)
LIMIT 1`

const selectPostgresCustomersSQL = `
SELECT
  COALESCE(customer.customer_ref, customer.id::text),
  customer.code,
  customer.name,
  COALESCE(customer.customer_type, 'distributor'),
  COALESCE(customer.channel_code, ''),
  COALESCE(customer.price_list_code, ''),
  COALESCE(customer.discount_group, ''),
  customer.credit_limit::text,
  COALESCE(customer.payment_terms, ''),
  COALESCE(customer.contact_name, ''),
  COALESCE(customer.phone, ''),
  COALESCE(customer.email, ''),
  COALESCE(customer.tax_code, ''),
  COALESCE(customer.address, ''),
  customer.status,
  customer.created_at,
  customer.updated_at
FROM mdm.customers AS customer
ORDER BY customer.status, customer.customer_type, customer.code`

const selectPostgresCustomerSQL = `
SELECT
  COALESCE(customer.customer_ref, customer.id::text),
  customer.code,
  customer.name,
  COALESCE(customer.customer_type, 'distributor'),
  COALESCE(customer.channel_code, ''),
  COALESCE(customer.price_list_code, ''),
  COALESCE(customer.discount_group, ''),
  customer.credit_limit::text,
  COALESCE(customer.payment_terms, ''),
  COALESCE(customer.contact_name, ''),
  COALESCE(customer.phone, ''),
  COALESCE(customer.email, ''),
  COALESCE(customer.tax_code, ''),
  COALESCE(customer.address, ''),
  customer.status,
  customer.created_at,
  customer.updated_at
FROM mdm.customers AS customer
WHERE lower(COALESCE(customer.customer_ref, customer.id::text)) = lower($1)
   OR customer.id::text = $1
   OR lower(customer.code) = lower($1)
LIMIT 1`

const insertPostgresCustomerSQL = `
INSERT INTO mdm.customers (
  id,
  org_id,
  customer_ref,
  code,
  name,
  customer_type,
  channel_code,
  price_list_code,
  discount_group,
  credit_limit,
  payment_terms,
  contact_name,
  phone,
  email,
  tax_code,
  address,
  status,
  created_at,
  updated_at
) VALUES (
  COALESCE($1::uuid, gen_random_uuid()),
  $2::uuid,
  $3,
  $4,
  $5,
  $6,
  $7,
  $8,
  $9,
  $10,
  $11,
  $12,
  $13,
  $14,
  $15,
  $16,
  $17,
  $18,
  $19
)`

const updatePostgresCustomerSQL = `
UPDATE mdm.customers
SET code = $3,
    name = $4,
    customer_type = $5,
    channel_code = $6,
    price_list_code = $7,
    discount_group = $8,
    credit_limit = $9,
    payment_terms = $10,
    contact_name = $11,
    phone = $12,
    email = $13,
    tax_code = $14,
    address = $15,
    status = $16,
    updated_at = $17,
    version = version + 1
WHERE org_id = $1::uuid
  AND lower(COALESCE(customer_ref, id::text)) = lower($2)`

const updatePostgresCustomerStatusSQL = `
UPDATE mdm.customers
SET status = $3,
    updated_at = $4,
    version = version + 1
WHERE org_id = $1::uuid
  AND lower(COALESCE(customer_ref, id::text)) = lower($2)`

const selectPostgresCustomerDuplicateCodeSQL = `
SELECT COALESCE(customer_ref, id::text)
FROM mdm.customers
WHERE org_id = $1::uuid
  AND lower(code) = lower($2)
  AND lower(COALESCE(customer_ref, id::text)) <> lower($3)
LIMIT 1`

func (s *PostgresPartyCatalog) ListSuppliers(ctx context.Context, filter domain.SupplierFilter) ([]domain.Supplier, response.Pagination, error) {
	if s == nil || s.db == nil {
		return nil, response.Pagination{}, errors.New("database connection is required")
	}
	if filter.Status != "" && !domain.IsValidSupplierStatus(filter.Status) {
		return nil, response.Pagination{}, domain.ErrSupplierInvalidStatus
	}
	if filter.Group != "" && !domain.IsValidSupplierGroup(filter.Group) {
		return nil, response.Pagination{}, domain.ErrSupplierInvalidGroup
	}

	rows, err := s.db.QueryContext(ctx, selectPostgresSuppliersSQL)
	if err != nil {
		return nil, response.Pagination{}, err
	}
	defer rows.Close()

	suppliers := make([]domain.Supplier, 0)
	for rows.Next() {
		supplier, err := scanPostgresSupplier(rows)
		if err != nil {
			return nil, response.Pagination{}, err
		}
		if filter.Matches(supplier) {
			suppliers = append(suppliers, supplier)
		}
	}
	if err := rows.Err(); err != nil {
		return nil, response.Pagination{}, err
	}
	domain.SortSuppliers(suppliers)
	pageRows, pagination := paginateSuppliers(suppliers, filter.Page, filter.PageSize)

	return pageRows, pagination, nil
}

func (s *PostgresPartyCatalog) GetSupplier(ctx context.Context, id string) (domain.Supplier, error) {
	if s == nil || s.db == nil {
		return domain.Supplier{}, errors.New("database connection is required")
	}
	supplier, err := scanPostgresSupplier(s.db.QueryRowContext(ctx, selectPostgresSupplierSQL, strings.TrimSpace(id)))
	if errors.Is(err, sql.ErrNoRows) {
		return domain.Supplier{}, ErrSupplierNotFound
	}
	if err != nil {
		return domain.Supplier{}, err
	}

	return supplier, nil
}

func (s *PostgresPartyCatalog) CreateSupplier(ctx context.Context, input CreateSupplierInput) (SupplierResult, error) {
	if s == nil || s.db == nil {
		return SupplierResult{}, errors.New("database connection is required")
	}
	if s.auditLog == nil {
		return SupplierResult{}, errors.New("audit log store is required")
	}
	now := s.clock().UTC()
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
	if err := s.insertSupplier(ctx, supplier); err != nil {
		return SupplierResult{}, err
	}
	log, err := newSupplierAuditLog(input.ActorID, input.RequestID, "masterdata.supplier.created", supplier, nil, supplierToAuditMap(supplier), now)
	if err != nil {
		return SupplierResult{}, err
	}
	if err := s.auditLog.Record(ctx, log); err != nil {
		return SupplierResult{}, err
	}

	return SupplierResult{Supplier: supplier, AuditLogID: log.ID}, nil
}

func (s *PostgresPartyCatalog) UpdateSupplier(ctx context.Context, input UpdateSupplierInput) (SupplierResult, error) {
	if s == nil || s.db == nil {
		return SupplierResult{}, errors.New("database connection is required")
	}
	if s.auditLog == nil {
		return SupplierResult{}, errors.New("audit log store is required")
	}
	current, err := s.GetSupplier(ctx, input.ID)
	if err != nil {
		return SupplierResult{}, err
	}
	now := s.clock().UTC()
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
		return SupplierResult{}, err
	}
	if err := s.updateSupplier(ctx, updated); err != nil {
		return SupplierResult{}, err
	}
	log, err := newSupplierAuditLog(input.ActorID, input.RequestID, "masterdata.supplier.updated", updated, supplierToAuditMap(current), supplierToAuditMap(updated), now)
	if err != nil {
		return SupplierResult{}, err
	}
	if err := s.auditLog.Record(ctx, log); err != nil {
		return SupplierResult{}, err
	}

	return SupplierResult{Supplier: updated, AuditLogID: log.ID}, nil
}

func (s *PostgresPartyCatalog) ChangeSupplierStatus(ctx context.Context, input ChangeSupplierStatusInput) (SupplierResult, error) {
	if s == nil || s.db == nil {
		return SupplierResult{}, errors.New("database connection is required")
	}
	if s.auditLog == nil {
		return SupplierResult{}, errors.New("audit log store is required")
	}
	current, err := s.GetSupplier(ctx, input.ID)
	if err != nil {
		return SupplierResult{}, err
	}
	now := s.clock().UTC()
	updated, err := current.ChangeStatus(domain.SupplierStatus(input.Status), now)
	if err != nil {
		return SupplierResult{}, err
	}
	orgID, err := s.resolveOrgID()
	if err != nil {
		return SupplierResult{}, err
	}
	result, err := s.db.ExecContext(ctx, updatePostgresSupplierStatusSQL, orgID, updated.ID, string(updated.Status), updated.UpdatedAt)
	if err != nil {
		return SupplierResult{}, fmt.Errorf("update supplier status %q: %w", updated.ID, err)
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return SupplierResult{}, err
	}
	if affected == 0 {
		return SupplierResult{}, ErrSupplierNotFound
	}
	log, err := newSupplierAuditLog(input.ActorID, input.RequestID, "masterdata.supplier.status_changed", updated, map[string]any{"status": string(current.Status)}, map[string]any{"status": string(updated.Status)}, now)
	if err != nil {
		return SupplierResult{}, err
	}
	if err := s.auditLog.Record(ctx, log); err != nil {
		return SupplierResult{}, err
	}

	return SupplierResult{Supplier: updated, AuditLogID: log.ID}, nil
}

func (s *PostgresPartyCatalog) ListCustomers(ctx context.Context, filter domain.CustomerFilter) ([]domain.Customer, response.Pagination, error) {
	if s == nil || s.db == nil {
		return nil, response.Pagination{}, errors.New("database connection is required")
	}
	if filter.Status != "" && !domain.IsValidCustomerStatus(filter.Status) {
		return nil, response.Pagination{}, domain.ErrCustomerInvalidStatus
	}
	if filter.Type != "" && !domain.IsValidCustomerType(filter.Type) {
		return nil, response.Pagination{}, domain.ErrCustomerInvalidType
	}

	rows, err := s.db.QueryContext(ctx, selectPostgresCustomersSQL)
	if err != nil {
		return nil, response.Pagination{}, err
	}
	defer rows.Close()

	customers := make([]domain.Customer, 0)
	for rows.Next() {
		customer, err := scanPostgresCustomer(rows)
		if err != nil {
			return nil, response.Pagination{}, err
		}
		if filter.Matches(customer) {
			customers = append(customers, customer)
		}
	}
	if err := rows.Err(); err != nil {
		return nil, response.Pagination{}, err
	}
	domain.SortCustomers(customers)
	pageRows, pagination := paginateCustomers(customers, filter.Page, filter.PageSize)

	return pageRows, pagination, nil
}

func (s *PostgresPartyCatalog) GetCustomer(ctx context.Context, id string) (domain.Customer, error) {
	if s == nil || s.db == nil {
		return domain.Customer{}, errors.New("database connection is required")
	}
	customer, err := scanPostgresCustomer(s.db.QueryRowContext(ctx, selectPostgresCustomerSQL, strings.TrimSpace(id)))
	if errors.Is(err, sql.ErrNoRows) {
		return domain.Customer{}, ErrCustomerNotFound
	}
	if err != nil {
		return domain.Customer{}, err
	}

	return customer, nil
}

func (s *PostgresPartyCatalog) CreateCustomer(ctx context.Context, input CreateCustomerInput) (CustomerResult, error) {
	if s == nil || s.db == nil {
		return CustomerResult{}, errors.New("database connection is required")
	}
	if s.auditLog == nil {
		return CustomerResult{}, errors.New("audit log store is required")
	}
	now := s.clock().UTC()
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
	if err := s.insertCustomer(ctx, customer); err != nil {
		return CustomerResult{}, err
	}
	log, err := newCustomerAuditLog(input.ActorID, input.RequestID, "masterdata.customer.created", customer, nil, customerToAuditMap(customer), now)
	if err != nil {
		return CustomerResult{}, err
	}
	if err := s.auditLog.Record(ctx, log); err != nil {
		return CustomerResult{}, err
	}

	return CustomerResult{Customer: customer, AuditLogID: log.ID}, nil
}

func (s *PostgresPartyCatalog) UpdateCustomer(ctx context.Context, input UpdateCustomerInput) (CustomerResult, error) {
	if s == nil || s.db == nil {
		return CustomerResult{}, errors.New("database connection is required")
	}
	if s.auditLog == nil {
		return CustomerResult{}, errors.New("audit log store is required")
	}
	current, err := s.GetCustomer(ctx, input.ID)
	if err != nil {
		return CustomerResult{}, err
	}
	now := s.clock().UTC()
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
		return CustomerResult{}, err
	}
	if err := s.updateCustomer(ctx, updated); err != nil {
		return CustomerResult{}, err
	}
	log, err := newCustomerAuditLog(input.ActorID, input.RequestID, "masterdata.customer.updated", updated, customerToAuditMap(current), customerToAuditMap(updated), now)
	if err != nil {
		return CustomerResult{}, err
	}
	if err := s.auditLog.Record(ctx, log); err != nil {
		return CustomerResult{}, err
	}

	return CustomerResult{Customer: updated, AuditLogID: log.ID}, nil
}

func (s *PostgresPartyCatalog) ChangeCustomerStatus(ctx context.Context, input ChangeCustomerStatusInput) (CustomerResult, error) {
	if s == nil || s.db == nil {
		return CustomerResult{}, errors.New("database connection is required")
	}
	if s.auditLog == nil {
		return CustomerResult{}, errors.New("audit log store is required")
	}
	current, err := s.GetCustomer(ctx, input.ID)
	if err != nil {
		return CustomerResult{}, err
	}
	now := s.clock().UTC()
	updated, err := current.ChangeStatus(domain.CustomerStatus(input.Status), now)
	if err != nil {
		return CustomerResult{}, err
	}
	orgID, err := s.resolveOrgID()
	if err != nil {
		return CustomerResult{}, err
	}
	result, err := s.db.ExecContext(ctx, updatePostgresCustomerStatusSQL, orgID, updated.ID, string(updated.Status), updated.UpdatedAt)
	if err != nil {
		return CustomerResult{}, fmt.Errorf("update customer status %q: %w", updated.ID, err)
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return CustomerResult{}, err
	}
	if affected == 0 {
		return CustomerResult{}, ErrCustomerNotFound
	}
	log, err := newCustomerAuditLog(input.ActorID, input.RequestID, "masterdata.customer.status_changed", updated, map[string]any{"status": string(current.Status)}, map[string]any{"status": string(updated.Status)}, now)
	if err != nil {
		return CustomerResult{}, err
	}
	if err := s.auditLog.Record(ctx, log); err != nil {
		return CustomerResult{}, err
	}

	return CustomerResult{Customer: updated, AuditLogID: log.ID}, nil
}

func (s *PostgresPartyCatalog) insertSupplier(ctx context.Context, supplier domain.Supplier) error {
	orgID, err := s.resolveOrgID()
	if err != nil {
		return err
	}
	if err := s.ensureUniqueSupplier(ctx, orgID, supplier, ""); err != nil {
		return err
	}
	_, err = s.db.ExecContext(ctx, insertPostgresSupplierSQL,
		nullablePostgresPartyUUID(supplier.ID),
		orgID,
		supplier.ID,
		supplier.Code,
		supplier.Name,
		postgresSupplierType(supplier.Group),
		string(supplier.Group),
		nullablePostgresPartyText(supplier.ContactName),
		nullablePostgresPartyText(supplier.Phone),
		nullablePostgresPartyText(supplier.Email),
		nullablePostgresPartyText(supplier.TaxCode),
		nullablePostgresPartyText(supplier.Address),
		nullablePostgresPartyText(supplier.PaymentTerms),
		supplier.LeadTimeDays,
		supplier.MOQ.String(),
		supplier.QualityScore.String(),
		supplier.DeliveryScore.String(),
		string(supplier.Status),
		supplier.CreatedAt,
		supplier.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("insert supplier %q: %w", supplier.ID, err)
	}

	return nil
}

func (s *PostgresPartyCatalog) updateSupplier(ctx context.Context, supplier domain.Supplier) error {
	orgID, err := s.resolveOrgID()
	if err != nil {
		return err
	}
	if err := s.ensureUniqueSupplier(ctx, orgID, supplier, supplier.ID); err != nil {
		return err
	}
	result, err := s.db.ExecContext(ctx, updatePostgresSupplierSQL,
		orgID,
		supplier.ID,
		supplier.Code,
		supplier.Name,
		postgresSupplierType(supplier.Group),
		string(supplier.Group),
		nullablePostgresPartyText(supplier.ContactName),
		nullablePostgresPartyText(supplier.Phone),
		nullablePostgresPartyText(supplier.Email),
		nullablePostgresPartyText(supplier.TaxCode),
		nullablePostgresPartyText(supplier.Address),
		nullablePostgresPartyText(supplier.PaymentTerms),
		supplier.LeadTimeDays,
		supplier.MOQ.String(),
		supplier.QualityScore.String(),
		supplier.DeliveryScore.String(),
		string(supplier.Status),
		supplier.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("update supplier %q: %w", supplier.ID, err)
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return ErrSupplierNotFound
	}

	return nil
}

func (s *PostgresPartyCatalog) insertCustomer(ctx context.Context, customer domain.Customer) error {
	orgID, err := s.resolveOrgID()
	if err != nil {
		return err
	}
	if err := s.ensureUniqueCustomer(ctx, orgID, customer, ""); err != nil {
		return err
	}
	_, err = s.db.ExecContext(ctx, insertPostgresCustomerSQL,
		nullablePostgresPartyUUID(customer.ID),
		orgID,
		customer.ID,
		customer.Code,
		customer.Name,
		string(customer.Type),
		nullablePostgresPartyText(customer.ChannelCode),
		nullablePostgresPartyText(customer.PriceListCode),
		nullablePostgresPartyText(customer.DiscountGroup),
		customer.CreditLimit.String(),
		nullablePostgresPartyText(customer.PaymentTerms),
		nullablePostgresPartyText(customer.ContactName),
		nullablePostgresPartyText(customer.Phone),
		nullablePostgresPartyText(customer.Email),
		nullablePostgresPartyText(customer.TaxCode),
		nullablePostgresPartyText(customer.Address),
		string(customer.Status),
		customer.CreatedAt,
		customer.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("insert customer %q: %w", customer.ID, err)
	}

	return nil
}

func (s *PostgresPartyCatalog) updateCustomer(ctx context.Context, customer domain.Customer) error {
	orgID, err := s.resolveOrgID()
	if err != nil {
		return err
	}
	if err := s.ensureUniqueCustomer(ctx, orgID, customer, customer.ID); err != nil {
		return err
	}
	result, err := s.db.ExecContext(ctx, updatePostgresCustomerSQL,
		orgID,
		customer.ID,
		customer.Code,
		customer.Name,
		string(customer.Type),
		nullablePostgresPartyText(customer.ChannelCode),
		nullablePostgresPartyText(customer.PriceListCode),
		nullablePostgresPartyText(customer.DiscountGroup),
		customer.CreditLimit.String(),
		nullablePostgresPartyText(customer.PaymentTerms),
		nullablePostgresPartyText(customer.ContactName),
		nullablePostgresPartyText(customer.Phone),
		nullablePostgresPartyText(customer.Email),
		nullablePostgresPartyText(customer.TaxCode),
		nullablePostgresPartyText(customer.Address),
		string(customer.Status),
		customer.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("update customer %q: %w", customer.ID, err)
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return ErrCustomerNotFound
	}

	return nil
}

func (s *PostgresPartyCatalog) ensureUniqueSupplier(ctx context.Context, orgID string, supplier domain.Supplier, currentID string) error {
	var duplicate string
	err := s.db.QueryRowContext(ctx, selectPostgresSupplierDuplicateCodeSQL, orgID, supplier.Code, currentID).Scan(&duplicate)
	if err == nil {
		return ErrDuplicateSupplierCode
	}
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return fmt.Errorf("check duplicate supplier code: %w", err)
	}

	return nil
}

func (s *PostgresPartyCatalog) ensureUniqueCustomer(ctx context.Context, orgID string, customer domain.Customer, currentID string) error {
	var duplicate string
	err := s.db.QueryRowContext(ctx, selectPostgresCustomerDuplicateCodeSQL, orgID, customer.Code, currentID).Scan(&duplicate)
	if err == nil {
		return ErrDuplicateCustomerCode
	}
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return fmt.Errorf("check duplicate customer code: %w", err)
	}

	return nil
}

func (s *PostgresPartyCatalog) resolveOrgID() (string, error) {
	if isPostgresPartyUUIDText(s.defaultOrgID) {
		return s.defaultOrgID, nil
	}

	return "", errors.New("party catalog default org id is required")
}

func scanPostgresSupplier(scanner interface{ Scan(dest ...any) error }) (domain.Supplier, error) {
	var (
		supplier      domain.Supplier
		group         string
		status        string
		moq           string
		qualityScore  string
		deliveryScore string
		err           error
	)
	if err := scanner.Scan(
		&supplier.ID,
		&supplier.Code,
		&supplier.Name,
		&group,
		&supplier.ContactName,
		&supplier.Phone,
		&supplier.Email,
		&supplier.TaxCode,
		&supplier.Address,
		&supplier.PaymentTerms,
		&supplier.LeadTimeDays,
		&moq,
		&qualityScore,
		&deliveryScore,
		&status,
		&supplier.CreatedAt,
		&supplier.UpdatedAt,
	); err != nil {
		return domain.Supplier{}, err
	}
	supplier.MOQ, err = decimal.ParseQuantity(moq)
	if err != nil {
		return domain.Supplier{}, err
	}
	supplier.QualityScore, err = decimal.ParseRate(qualityScore)
	if err != nil {
		return domain.Supplier{}, err
	}
	supplier.DeliveryScore, err = decimal.ParseRate(deliveryScore)
	if err != nil {
		return domain.Supplier{}, err
	}

	return domain.NewSupplier(domain.NewSupplierInput{
		ID:            supplier.ID,
		Code:          supplier.Code,
		Name:          supplier.Name,
		Group:         domain.SupplierGroup(group),
		ContactName:   supplier.ContactName,
		Phone:         supplier.Phone,
		Email:         supplier.Email,
		TaxCode:       supplier.TaxCode,
		Address:       supplier.Address,
		PaymentTerms:  supplier.PaymentTerms,
		LeadTimeDays:  supplier.LeadTimeDays,
		MOQ:           supplier.MOQ,
		QualityScore:  supplier.QualityScore,
		DeliveryScore: supplier.DeliveryScore,
		Status:        domain.SupplierStatus(status),
		CreatedAt:     supplier.CreatedAt,
		UpdatedAt:     supplier.UpdatedAt,
	})
}

func scanPostgresCustomer(scanner interface{ Scan(dest ...any) error }) (domain.Customer, error) {
	var (
		customer    domain.Customer
		customerTyp string
		status      string
		creditLimit string
		err         error
	)
	if err := scanner.Scan(
		&customer.ID,
		&customer.Code,
		&customer.Name,
		&customerTyp,
		&customer.ChannelCode,
		&customer.PriceListCode,
		&customer.DiscountGroup,
		&creditLimit,
		&customer.PaymentTerms,
		&customer.ContactName,
		&customer.Phone,
		&customer.Email,
		&customer.TaxCode,
		&customer.Address,
		&status,
		&customer.CreatedAt,
		&customer.UpdatedAt,
	); err != nil {
		return domain.Customer{}, err
	}
	customer.CreditLimit, err = decimal.ParseMoneyAmount(creditLimit)
	if err != nil {
		return domain.Customer{}, err
	}

	return domain.NewCustomer(domain.NewCustomerInput{
		ID:            customer.ID,
		Code:          customer.Code,
		Name:          customer.Name,
		Type:          domain.CustomerType(customerTyp),
		ChannelCode:   customer.ChannelCode,
		PriceListCode: customer.PriceListCode,
		DiscountGroup: customer.DiscountGroup,
		CreditLimit:   customer.CreditLimit,
		PaymentTerms:  customer.PaymentTerms,
		ContactName:   customer.ContactName,
		Phone:         customer.Phone,
		Email:         customer.Email,
		TaxCode:       customer.TaxCode,
		Address:       customer.Address,
		Status:        domain.CustomerStatus(status),
		CreatedAt:     customer.CreatedAt,
		UpdatedAt:     customer.UpdatedAt,
	})
}

func postgresSupplierType(group domain.SupplierGroup) string {
	switch group {
	case domain.SupplierGroupOutsource:
		return "factory"
	case domain.SupplierGroupLogistics:
		return "carrier_partner"
	default:
		return "supplier"
	}
}

func nullablePostgresPartyText(value string) any {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil
	}

	return value
}

func nullablePostgresPartyUUID(value string) any {
	value = strings.TrimSpace(value)
	if !isPostgresPartyUUIDText(value) {
		return nil
	}

	return value
}

func isPostgresPartyUUIDText(value string) bool {
	value = strings.TrimSpace(value)
	if len(value) != 36 {
		return false
	}
	for index, char := range value {
		switch index {
		case 8, 13, 18, 23:
			if char != '-' {
				return false
			}
		default:
			if !((char >= '0' && char <= '9') || (char >= 'a' && char <= 'f') || (char >= 'A' && char <= 'F')) {
				return false
			}
		}
	}

	return true
}
