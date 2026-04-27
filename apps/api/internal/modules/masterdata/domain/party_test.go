package domain

import (
	"errors"
	"testing"
)

func TestNewSupplierNormalizesAndFiltersFields(t *testing.T) {
	supplier, err := NewSupplier(NewSupplierInput{
		ID:            "sup-test",
		Code:          " sup-rm-bio ",
		Name:          "BioActive Raw Materials",
		Group:         SupplierGroupRawMaterial,
		ContactName:   " Nguyen Van An ",
		Email:         " Purchasing@BioActive.Example ",
		TaxCode:       " 0312345001 ",
		Address:       "Binh Duong",
		PaymentTerms:  " net30 ",
		LeadTimeDays:  12,
		MOQ:           50,
		QualityScore:  94,
		DeliveryScore: 91,
		Status:        SupplierStatusActive,
	})
	if err != nil {
		t.Fatalf("new supplier: %v", err)
	}

	if supplier.Code != "SUP-RM-BIO" || supplier.Email != "purchasing@bioactive.example" || supplier.PaymentTerms != "NET30" {
		t.Fatalf("supplier = %+v, want normalized code email and payment terms", supplier)
	}

	filter := NewSupplierFilter("bioactive", SupplierStatusActive, SupplierGroupRawMaterial, 1, 20)
	if !filter.Matches(supplier) {
		t.Fatal("filter did not match supplier status/group/search")
	}
}

func TestNewSupplierRejectsInvalidGroupAndBlockedTransition(t *testing.T) {
	_, err := NewSupplier(NewSupplierInput{
		ID:     "sup-test",
		Code:   "SUP-UNKNOWN",
		Name:   "Unknown Supplier",
		Group:  SupplierGroup("unknown"),
		Status: SupplierStatusActive,
	})
	if !errors.Is(err, ErrSupplierInvalidGroup) {
		t.Fatalf("error = %v, want invalid supplier group", err)
	}

	supplier, err := NewSupplier(NewSupplierInput{
		ID:     "sup-test",
		Code:   "SUP-BLOCKED",
		Name:   "Blocked Supplier",
		Group:  SupplierGroupService,
		Status: SupplierStatusBlacklisted,
	})
	if err != nil {
		t.Fatalf("new blacklisted supplier: %v", err)
	}
	_, err = supplier.ChangeStatus(SupplierStatusActive, supplier.UpdatedAt)
	if !errors.Is(err, ErrSupplierInvalidStatusTransition) {
		t.Fatalf("error = %v, want invalid supplier status transition", err)
	}
}

func TestNewCustomerNormalizesAndFiltersFields(t *testing.T) {
	customer, err := NewCustomer(NewCustomerInput{
		ID:            "cus-test",
		Code:          " cus-dl-minhanh ",
		Name:          "Minh Anh Distributor",
		Type:          CustomerTypeDistributor,
		ChannelCode:   " b2b ",
		PriceListCode: " pl-b2b-2026 ",
		DiscountGroup: " tier_1 ",
		CreditLimit:   500000000,
		PaymentTerms:  " net30 ",
		Email:         " Orders@MinhAnh.Example ",
		TaxCode:       " 0315678001 ",
		Address:       "District 7",
		Status:        CustomerStatusActive,
	})
	if err != nil {
		t.Fatalf("new customer: %v", err)
	}

	if customer.Code != "CUS-DL-MINHANH" || customer.ChannelCode != "B2B" || customer.PriceListCode != "PL-B2B-2026" {
		t.Fatalf("customer = %+v, want normalized codes", customer)
	}

	filter := NewCustomerFilter("minh", CustomerStatusActive, CustomerTypeDistributor, 1, 20)
	if !filter.Matches(customer) {
		t.Fatal("filter did not match customer status/type/search")
	}
}

func TestNewCustomerRejectsInvalidTypeAndBlockedTransition(t *testing.T) {
	_, err := NewCustomer(NewCustomerInput{
		ID:     "cus-test",
		Code:   "CUS-UNKNOWN",
		Name:   "Unknown Customer",
		Type:   CustomerType("unknown"),
		Status: CustomerStatusActive,
	})
	if !errors.Is(err, ErrCustomerInvalidType) {
		t.Fatalf("error = %v, want invalid customer type", err)
	}

	customer, err := NewCustomer(NewCustomerInput{
		ID:     "cus-test",
		Code:   "CUS-BLOCKED",
		Name:   "Blocked Customer",
		Type:   CustomerTypeDealer,
		Status: CustomerStatusBlocked,
	})
	if err != nil {
		t.Fatalf("new blocked customer: %v", err)
	}
	_, err = customer.ChangeStatus(CustomerStatusActive, customer.UpdatedAt)
	if !errors.Is(err, ErrCustomerInvalidStatusTransition) {
		t.Fatalf("error = %v, want invalid customer status transition", err)
	}
}
