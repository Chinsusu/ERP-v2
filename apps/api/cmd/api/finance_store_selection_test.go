package main

import (
	"testing"

	financeapp "github.com/Chinsusu/ERP-v2/apps/api/internal/modules/finance/application"
	"github.com/Chinsusu/ERP-v2/apps/api/internal/shared/config"
)

func TestNewRuntimeFinanceStoresFallsBackToPrototypeWithoutDatabaseURL(t *testing.T) {
	stores, closeStores, err := newRuntimeFinanceStores(config.Config{})
	if err != nil {
		t.Fatalf("newRuntimeFinanceStores() error = %v", err)
	}
	if closeStores != nil {
		t.Fatal("closeStores is not nil, want nil for prototype stores")
	}

	if _, ok := stores.customerReceivables.(*financeapp.PrototypeCustomerReceivableStore); !ok {
		t.Fatalf("customer receivable store type = %T, want *PrototypeCustomerReceivableStore", stores.customerReceivables)
	}
	if _, ok := stores.supplierPayables.(*financeapp.PrototypeSupplierPayableStore); !ok {
		t.Fatalf("supplier payable store type = %T, want *PrototypeSupplierPayableStore", stores.supplierPayables)
	}
	if _, ok := stores.codRemittances.(*financeapp.PrototypeCODRemittanceStore); !ok {
		t.Fatalf("cod remittance store type = %T, want *PrototypeCODRemittanceStore", stores.codRemittances)
	}
	if _, ok := stores.cashTransactions.(*financeapp.PrototypeCashTransactionStore); !ok {
		t.Fatalf("cash transaction store type = %T, want *PrototypeCashTransactionStore", stores.cashTransactions)
	}
}

func TestNewRuntimeFinanceStoresUsesPostgresAsPackageWhenDatabaseURLConfigured(t *testing.T) {
	stores, closeStores, err := newRuntimeFinanceStores(config.Config{
		AppEnv:      "dev",
		DatabaseURL: "postgres://erp_dev:erp_dev@postgres:5432/erp_dev?sslmode=disable",
	})
	if err != nil {
		t.Fatalf("newRuntimeFinanceStores() error = %v", err)
	}
	if closeStores == nil {
		t.Fatal("closeStores = nil, want database close function")
	}
	defer func() {
		if err := closeStores(); err != nil {
			t.Fatalf("closeStores() error = %v", err)
		}
	}()

	if _, ok := stores.customerReceivables.(financeapp.PostgresCustomerReceivableStore); !ok {
		t.Fatalf("customer receivable store type = %T, want PostgresCustomerReceivableStore", stores.customerReceivables)
	}
	if _, ok := stores.supplierPayables.(financeapp.PostgresSupplierPayableStore); !ok {
		t.Fatalf("supplier payable store type = %T, want PostgresSupplierPayableStore", stores.supplierPayables)
	}
	if _, ok := stores.codRemittances.(financeapp.PostgresCODRemittanceStore); !ok {
		t.Fatalf("cod remittance store type = %T, want PostgresCODRemittanceStore", stores.codRemittances)
	}
	if _, ok := stores.cashTransactions.(financeapp.PostgresCashTransactionStore); !ok {
		t.Fatalf("cash transaction store type = %T, want PostgresCashTransactionStore", stores.cashTransactions)
	}
}
