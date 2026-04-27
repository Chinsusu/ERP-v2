package application

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/Chinsusu/ERP-v2/apps/api/internal/modules/masterdata/domain"
	"github.com/Chinsusu/ERP-v2/apps/api/internal/shared/audit"
)

func TestPartyCatalogListsFilteredPrototypeSuppliers(t *testing.T) {
	store := NewPrototypePartyCatalog(audit.NewInMemoryLogStore())

	suppliers, pagination, err := store.ListSuppliers(context.Background(), domain.NewSupplierFilter("bio", domain.SupplierStatusActive, domain.SupplierGroupRawMaterial, 1, 20))
	if err != nil {
		t.Fatalf("list suppliers: %v", err)
	}

	if len(suppliers) != 1 {
		t.Fatalf("suppliers = %d, want 1", len(suppliers))
	}
	if suppliers[0].Code != "SUP-RM-BIO" {
		t.Fatalf("supplier = %q, want SUP-RM-BIO", suppliers[0].Code)
	}
	if pagination.TotalItems != 1 || pagination.Page != 1 {
		t.Fatalf("pagination = %+v, want one supplier on page 1", pagination)
	}
}

func TestPartyCatalogCreatesUpdatesSupplierStatusAndWritesAudit(t *testing.T) {
	auditStore := audit.NewInMemoryLogStore()
	store := NewPrototypePartyCatalogAt(auditStore, time.Date(2026, 4, 27, 9, 0, 0, 0, time.UTC))
	ctx := context.Background()

	created, err := store.CreateSupplier(ctx, CreateSupplierInput{
		Code:          "SUP-SVC-LAB",
		Name:          "Lab Services Partner",
		Group:         "service",
		ContactName:   "Nguyen Lab",
		Email:         "Lab@Partner.Example",
		TaxCode:       "0319999001",
		Address:       "Ho Chi Minh lab site",
		PaymentTerms:  "NET15",
		LeadTimeDays:  5,
		QualityScore:  "90",
		DeliveryScore: "92",
		Status:        "draft",
		ActorID:       "user-erp-admin",
		RequestID:     "req-supplier-create",
	})
	if err != nil {
		t.Fatalf("create supplier: %v", err)
	}
	if created.Supplier.Code != "SUP-SVC-LAB" || created.Supplier.Email != "lab@partner.example" || created.AuditLogID == "" {
		t.Fatalf("supplier = %+v, want normalized supplier with audit", created)
	}

	updated, err := store.UpdateSupplier(ctx, UpdateSupplierInput{
		ID:            created.Supplier.ID,
		Code:          "SUP-SVC-LAB",
		Name:          "Lab Services Partner v2",
		Group:         "service",
		ContactName:   "Nguyen Lab",
		Email:         "lab@partner.example",
		TaxCode:       "0319999001",
		Address:       "Ho Chi Minh lab site",
		PaymentTerms:  "NET30",
		LeadTimeDays:  4,
		QualityScore:  "91",
		DeliveryScore: "93",
		Status:        "active",
		ActorID:       "user-erp-admin",
		RequestID:     "req-supplier-update",
	})
	if err != nil {
		t.Fatalf("update supplier: %v", err)
	}
	if updated.Supplier.Name != "Lab Services Partner v2" || updated.Supplier.Status != domain.SupplierStatusActive {
		t.Fatalf("updated supplier = %+v, want active updated name", updated.Supplier)
	}

	statusChanged, err := store.ChangeSupplierStatus(ctx, ChangeSupplierStatusInput{
		ID:        created.Supplier.ID,
		Status:    "inactive",
		ActorID:   "user-erp-admin",
		RequestID: "req-supplier-status",
	})
	if err != nil {
		t.Fatalf("change supplier status: %v", err)
	}
	if statusChanged.Supplier.Status != domain.SupplierStatusInactive {
		t.Fatalf("status = %q, want inactive", statusChanged.Supplier.Status)
	}

	logs, err := auditStore.List(ctx, audit.Query{EntityID: created.Supplier.ID})
	if err != nil {
		t.Fatalf("list audit logs: %v", err)
	}
	if len(logs) != 3 {
		t.Fatalf("audit logs = %d, want 3", len(logs))
	}
}

func TestPartyCatalogBlocksDuplicateSupplierAndInvalidTransition(t *testing.T) {
	store := NewPrototypePartyCatalog(audit.NewInMemoryLogStore())

	_, err := store.CreateSupplier(context.Background(), CreateSupplierInput{
		Code:   "SUP-RM-BIO",
		Name:   "Duplicate Supplier",
		Group:  "raw_material",
		Status: "active",
	})
	if !errors.Is(err, ErrDuplicateSupplierCode) {
		t.Fatalf("error = %v, want duplicate supplier code", err)
	}

	_, err = store.ChangeSupplierStatus(context.Background(), ChangeSupplierStatusInput{
		ID:      "sup-out-lotus",
		Status:  "blacklisted",
		ActorID: "user-erp-admin",
	})
	if err != nil {
		t.Fatalf("blacklist supplier: %v", err)
	}
	_, err = store.ChangeSupplierStatus(context.Background(), ChangeSupplierStatusInput{
		ID:     "sup-out-lotus",
		Status: "active",
	})
	if !errors.Is(err, domain.ErrSupplierInvalidStatusTransition) {
		t.Fatalf("error = %v, want invalid supplier transition", err)
	}
}

func TestPartyCatalogCreatesUpdatesCustomerStatusAndWritesAudit(t *testing.T) {
	auditStore := audit.NewInMemoryLogStore()
	store := NewPrototypePartyCatalogAt(auditStore, time.Date(2026, 4, 27, 9, 0, 0, 0, time.UTC))
	ctx := context.Background()

	created, err := store.CreateCustomer(ctx, CreateCustomerInput{
		Code:          "CUS-DL-HANOI",
		Name:          "Ha Noi Dealer",
		Type:          "dealer",
		ChannelCode:   "dealer",
		PriceListCode: "pl-dealer-2026",
		DiscountGroup: "tier_2",
		CreditLimit:   "200000000",
		PaymentTerms:  "NET15",
		ContactName:   "Tran Ha Noi",
		Email:         "Buyer@HaNoiDealer.Example",
		TaxCode:       "0319999002",
		Address:       "Ha Noi",
		Status:        "draft",
		ActorID:       "user-erp-admin",
		RequestID:     "req-customer-create",
	})
	if err != nil {
		t.Fatalf("create customer: %v", err)
	}
	if created.Customer.Code != "CUS-DL-HANOI" || created.Customer.ChannelCode != "DEALER" || created.AuditLogID == "" {
		t.Fatalf("customer = %+v, want normalized customer with audit", created)
	}

	updated, err := store.UpdateCustomer(ctx, UpdateCustomerInput{
		ID:            created.Customer.ID,
		Code:          "CUS-DL-HANOI",
		Name:          "Ha Noi Dealer v2",
		Type:          "dealer",
		ChannelCode:   "DEALER",
		PriceListCode: "PL-DEALER-2026",
		DiscountGroup: "tier_1",
		CreditLimit:   "250000000",
		PaymentTerms:  "NET30",
		ContactName:   "Tran Ha Noi",
		Email:         "buyer@hanoidealer.example",
		TaxCode:       "0319999002",
		Address:       "Ha Noi",
		Status:        "active",
		ActorID:       "user-erp-admin",
		RequestID:     "req-customer-update",
	})
	if err != nil {
		t.Fatalf("update customer: %v", err)
	}
	if updated.Customer.Name != "Ha Noi Dealer v2" || updated.Customer.Status != domain.CustomerStatusActive {
		t.Fatalf("updated customer = %+v, want active updated name", updated.Customer)
	}

	statusChanged, err := store.ChangeCustomerStatus(ctx, ChangeCustomerStatusInput{
		ID:        created.Customer.ID,
		Status:    "inactive",
		ActorID:   "user-erp-admin",
		RequestID: "req-customer-status",
	})
	if err != nil {
		t.Fatalf("change customer status: %v", err)
	}
	if statusChanged.Customer.Status != domain.CustomerStatusInactive {
		t.Fatalf("status = %q, want inactive", statusChanged.Customer.Status)
	}

	logs, err := auditStore.List(ctx, audit.Query{EntityID: created.Customer.ID})
	if err != nil {
		t.Fatalf("list audit logs: %v", err)
	}
	if len(logs) != 3 {
		t.Fatalf("audit logs = %d, want 3", len(logs))
	}
}

func TestPartyCatalogBlocksDuplicateCustomerAndInvalidTransition(t *testing.T) {
	store := NewPrototypePartyCatalog(audit.NewInMemoryLogStore())

	_, err := store.CreateCustomer(context.Background(), CreateCustomerInput{
		Code:   "CUS-DL-MINHANH",
		Name:   "Duplicate Customer",
		Type:   "distributor",
		Status: "active",
	})
	if !errors.Is(err, ErrDuplicateCustomerCode) {
		t.Fatalf("error = %v, want duplicate customer code", err)
	}

	_, err = store.ChangeCustomerStatus(context.Background(), ChangeCustomerStatusInput{
		ID:      "cus-internal-hcm-store",
		Status:  "blocked",
		ActorID: "user-erp-admin",
	})
	if err != nil {
		t.Fatalf("block customer: %v", err)
	}
	_, err = store.ChangeCustomerStatus(context.Background(), ChangeCustomerStatusInput{
		ID:     "cus-internal-hcm-store",
		Status: "active",
	})
	if !errors.Is(err, domain.ErrCustomerInvalidStatusTransition) {
		t.Fatalf("error = %v, want invalid customer transition", err)
	}
}
