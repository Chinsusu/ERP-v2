package application

import (
	"context"
	"database/sql"
	"errors"
	"os"
	"testing"
	"time"

	"github.com/Chinsusu/ERP-v2/apps/api/internal/modules/masterdata/domain"
	"github.com/Chinsusu/ERP-v2/apps/api/internal/shared/audit"
	"github.com/Chinsusu/ERP-v2/apps/api/internal/shared/decimal"

	_ "github.com/jackc/pgx/v5/stdlib"
)

const testPostgresPartyCatalogOrgID = "00000000-0000-4000-8000-000000170401"

func TestPostgresPartyCatalogRequiresDatabase(t *testing.T) {
	store := NewPostgresPartyCatalog(nil, nil, PostgresPartyCatalogConfig{})

	if _, _, err := store.ListSuppliers(context.Background(), domain.SupplierFilter{}); err == nil {
		t.Fatal("ListSuppliers() error = nil, want database required error")
	}
	if _, err := store.GetSupplier(context.Background(), "supplier"); err == nil {
		t.Fatal("GetSupplier() error = nil, want database required error")
	}
	if _, err := store.CreateSupplier(context.Background(), CreateSupplierInput{}); err == nil {
		t.Fatal("CreateSupplier() error = nil, want database required error")
	}
	if _, _, err := store.ListCustomers(context.Background(), domain.CustomerFilter{}); err == nil {
		t.Fatal("ListCustomers() error = nil, want database required error")
	}
	if _, err := store.GetCustomer(context.Background(), "customer"); err == nil {
		t.Fatal("GetCustomer() error = nil, want database required error")
	}
	if _, err := store.CreateCustomer(context.Background(), CreateCustomerInput{}); err == nil {
		t.Fatal("CreateCustomer() error = nil, want database required error")
	}
}

func TestPostgresPartyCatalogPersistsSupplierCustomerLifecycleAndReload(t *testing.T) {
	databaseURL := os.Getenv("ERP_TEST_DATABASE_URL")
	if databaseURL == "" {
		t.Skip("ERP_TEST_DATABASE_URL is not set")
	}

	ctx := context.Background()
	db, err := sql.Open("pgx", databaseURL)
	if err != nil {
		t.Fatalf("open database: %v", err)
	}
	defer db.Close()

	if err := seedPostgresPartyCatalogFixture(ctx, db); err != nil {
		t.Fatalf("seed fixture: %v", err)
	}
	auditStore := audit.NewPostgresLogStore(db, audit.PostgresLogStoreConfig{DefaultOrgID: testPostgresPartyCatalogOrgID})
	store := NewPostgresPartyCatalog(db, auditStore, PostgresPartyCatalogConfig{
		DefaultOrgID: testPostgresPartyCatalogOrgID,
		Clock:        fixedPostgresPartyClock(time.Date(2026, 5, 2, 11, 0, 0, 0, time.UTC)),
	})

	createdSupplier, err := store.CreateSupplier(ctx, CreateSupplierInput{
		Code:          "SUP-S17-LAB",
		Name:          "S17 Lab Partner",
		Group:         "service",
		ContactName:   "Lab Lead",
		Phone:         "+84900000001",
		Email:         "Lab@S17.Example",
		TaxCode:       "0317000001",
		Address:       "HCM Lab",
		PaymentTerms:  "NET15",
		LeadTimeDays:  5,
		MOQ:           decimal.MustQuantity("10"),
		QualityScore:  decimal.MustRate("90"),
		DeliveryScore: decimal.MustRate("92"),
		Status:        "draft",
		ActorID:       "user-erp-admin",
		RequestID:     "req-s17-supplier-create",
	})
	if err != nil {
		t.Fatalf("CreateSupplier() error = %v", err)
	}
	if createdSupplier.Supplier.Email != "lab@s17.example" || createdSupplier.AuditLogID == "" {
		t.Fatalf("created supplier = %+v", createdSupplier)
	}
	if _, err := store.CreateSupplier(ctx, CreateSupplierInput{
		Code:   "SUP-S17-LAB",
		Name:   "Duplicate Supplier",
		Group:  "service",
		Status: "active",
	}); !errors.Is(err, ErrDuplicateSupplierCode) {
		t.Fatalf("duplicate supplier err = %v, want ErrDuplicateSupplierCode", err)
	}

	updatedSupplier, err := store.UpdateSupplier(ctx, UpdateSupplierInput{
		ID:            createdSupplier.Supplier.ID,
		Code:          "SUP-S17-LAB",
		Name:          "S17 Lab Partner Updated",
		Group:         "service",
		ContactName:   "Lab Lead",
		Phone:         "+84900000001",
		Email:         "lab@s17.example",
		TaxCode:       "0317000001",
		Address:       "HCM Lab",
		PaymentTerms:  "NET30",
		LeadTimeDays:  4,
		MOQ:           decimal.MustQuantity("12"),
		QualityScore:  decimal.MustRate("91"),
		DeliveryScore: decimal.MustRate("93"),
		Status:        "active",
		ActorID:       "user-erp-admin",
		RequestID:     "req-s17-supplier-update",
	})
	if err != nil {
		t.Fatalf("UpdateSupplier() error = %v", err)
	}
	if updatedSupplier.Supplier.Name != "S17 Lab Partner Updated" || updatedSupplier.Supplier.Status != domain.SupplierStatusActive {
		t.Fatalf("updated supplier = %+v", updatedSupplier.Supplier)
	}
	blacklisted, err := store.ChangeSupplierStatus(ctx, ChangeSupplierStatusInput{
		ID:        createdSupplier.Supplier.ID,
		Status:    "blacklisted",
		ActorID:   "user-erp-admin",
		RequestID: "req-s17-supplier-blacklist",
	})
	if err != nil {
		t.Fatalf("ChangeSupplierStatus() error = %v", err)
	}
	if blacklisted.Supplier.Status != domain.SupplierStatusBlacklisted {
		t.Fatalf("supplier status = %s, want blacklisted", blacklisted.Supplier.Status)
	}
	if _, err := store.ChangeSupplierStatus(ctx, ChangeSupplierStatusInput{
		ID:     createdSupplier.Supplier.ID,
		Status: "active",
	}); !errors.Is(err, domain.ErrSupplierInvalidStatusTransition) {
		t.Fatalf("supplier invalid transition err = %v, want ErrSupplierInvalidStatusTransition", err)
	}

	createdCustomer, err := store.CreateCustomer(ctx, CreateCustomerInput{
		Code:          "CUS-S17-HN",
		Name:          "S17 Ha Noi Dealer",
		Type:          "dealer",
		ChannelCode:   "dealer",
		PriceListCode: "pl-dealer-2026",
		DiscountGroup: "tier_2",
		CreditLimit:   decimal.MustMoneyAmount("200000000"),
		PaymentTerms:  "NET15",
		ContactName:   "HN Buyer",
		Phone:         "+84900000002",
		Email:         "Buyer@S17Dealer.Example",
		TaxCode:       "0317000002",
		Address:       "Ha Noi",
		Status:        "draft",
		ActorID:       "user-erp-admin",
		RequestID:     "req-s17-customer-create",
	})
	if err != nil {
		t.Fatalf("CreateCustomer() error = %v", err)
	}
	if createdCustomer.Customer.ChannelCode != "DEALER" || createdCustomer.AuditLogID == "" {
		t.Fatalf("created customer = %+v", createdCustomer)
	}
	if _, err := store.CreateCustomer(ctx, CreateCustomerInput{
		Code:   "CUS-S17-HN",
		Name:   "Duplicate Customer",
		Type:   "dealer",
		Status: "active",
	}); !errors.Is(err, ErrDuplicateCustomerCode) {
		t.Fatalf("duplicate customer err = %v, want ErrDuplicateCustomerCode", err)
	}

	updatedCustomer, err := store.UpdateCustomer(ctx, UpdateCustomerInput{
		ID:            createdCustomer.Customer.ID,
		Code:          "CUS-S17-HN",
		Name:          "S17 Ha Noi Dealer Updated",
		Type:          "dealer",
		ChannelCode:   "DEALER",
		PriceListCode: "PL-DEALER-2026",
		DiscountGroup: "tier_1",
		CreditLimit:   decimal.MustMoneyAmount("250000000"),
		PaymentTerms:  "NET30",
		ContactName:   "HN Buyer",
		Phone:         "+84900000002",
		Email:         "buyer@s17dealer.example",
		TaxCode:       "0317000002",
		Address:       "Ha Noi",
		Status:        "active",
		ActorID:       "user-erp-admin",
		RequestID:     "req-s17-customer-update",
	})
	if err != nil {
		t.Fatalf("UpdateCustomer() error = %v", err)
	}
	if updatedCustomer.Customer.Name != "S17 Ha Noi Dealer Updated" || updatedCustomer.Customer.Status != domain.CustomerStatusActive {
		t.Fatalf("updated customer = %+v", updatedCustomer.Customer)
	}
	blocked, err := store.ChangeCustomerStatus(ctx, ChangeCustomerStatusInput{
		ID:        createdCustomer.Customer.ID,
		Status:    "blocked",
		ActorID:   "user-erp-admin",
		RequestID: "req-s17-customer-block",
	})
	if err != nil {
		t.Fatalf("ChangeCustomerStatus() error = %v", err)
	}
	if blocked.Customer.Status != domain.CustomerStatusBlocked {
		t.Fatalf("customer status = %s, want blocked", blocked.Customer.Status)
	}
	if _, err := store.ChangeCustomerStatus(ctx, ChangeCustomerStatusInput{
		ID:     createdCustomer.Customer.ID,
		Status: "active",
	}); !errors.Is(err, domain.ErrCustomerInvalidStatusTransition) {
		t.Fatalf("customer invalid transition err = %v, want ErrCustomerInvalidStatusTransition", err)
	}

	reloadedStore := NewPostgresPartyCatalog(db, auditStore, PostgresPartyCatalogConfig{DefaultOrgID: testPostgresPartyCatalogOrgID})
	reloadedSupplier, err := reloadedStore.GetSupplier(ctx, createdSupplier.Supplier.ID)
	if err != nil {
		t.Fatalf("GetSupplier() reload error = %v", err)
	}
	if reloadedSupplier.Name != "S17 Lab Partner Updated" || reloadedSupplier.Status != domain.SupplierStatusBlacklisted {
		t.Fatalf("reloaded supplier = %+v", reloadedSupplier)
	}
	reloadedCustomer, err := reloadedStore.GetCustomer(ctx, createdCustomer.Customer.ID)
	if err != nil {
		t.Fatalf("GetCustomer() reload error = %v", err)
	}
	if reloadedCustomer.Name != "S17 Ha Noi Dealer Updated" || reloadedCustomer.Status != domain.CustomerStatusBlocked {
		t.Fatalf("reloaded customer = %+v", reloadedCustomer)
	}

	suppliers, supplierPagination, err := reloadedStore.ListSuppliers(ctx, domain.NewSupplierFilter("lab", domain.SupplierStatusBlacklisted, domain.SupplierGroupService, 1, 20))
	if err != nil {
		t.Fatalf("ListSuppliers() error = %v", err)
	}
	if supplierPagination.TotalItems == 0 || !containsPostgresSupplier(suppliers, createdSupplier.Supplier.ID) {
		t.Fatalf("suppliers = %+v pagination = %+v, missing created supplier", suppliers, supplierPagination)
	}
	customers, customerPagination, err := reloadedStore.ListCustomers(ctx, domain.NewCustomerFilter("ha noi", domain.CustomerStatusBlocked, domain.CustomerTypeDealer, 1, 20))
	if err != nil {
		t.Fatalf("ListCustomers() error = %v", err)
	}
	if customerPagination.TotalItems == 0 || !containsPostgresCustomer(customers, createdCustomer.Customer.ID) {
		t.Fatalf("customers = %+v pagination = %+v, missing created customer", customers, customerPagination)
	}

	supplierLogs, err := auditStore.List(ctx, audit.Query{EntityID: createdSupplier.Supplier.ID})
	if err != nil {
		t.Fatalf("list supplier audit logs: %v", err)
	}
	if len(supplierLogs) < 3 {
		t.Fatalf("supplier audit logs = %d, want at least 3", len(supplierLogs))
	}
	customerLogs, err := auditStore.List(ctx, audit.Query{EntityID: createdCustomer.Customer.ID})
	if err != nil {
		t.Fatalf("list customer audit logs: %v", err)
	}
	if len(customerLogs) < 3 {
		t.Fatalf("customer audit logs = %d, want at least 3", len(customerLogs))
	}
}

func seedPostgresPartyCatalogFixture(ctx context.Context, db *sql.DB) error {
	_, err := db.ExecContext(ctx, `
INSERT INTO core.organizations (id, code, name, status)
VALUES ($1::uuid, 'S17_PARTY_ORG', 'S17 Party Catalog Test Org', 'active')
ON CONFLICT (code) DO UPDATE
SET name = EXCLUDED.name,
    status = EXCLUDED.status,
    updated_at = now()`,
		testPostgresPartyCatalogOrgID,
	)

	return err
}

func fixedPostgresPartyClock(base time.Time) func() time.Time {
	current := base

	return func() time.Time {
		current = current.Add(time.Minute)
		return current
	}
}

func containsPostgresSupplier(suppliers []domain.Supplier, id string) bool {
	for _, supplier := range suppliers {
		if supplier.ID == id {
			return true
		}
	}

	return false
}

func containsPostgresCustomer(customers []domain.Customer, id string) bool {
	for _, customer := range customers {
		if customer.ID == id {
			return true
		}
	}

	return false
}
