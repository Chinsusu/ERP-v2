package main

import (
	"testing"

	productionapp "github.com/Chinsusu/ERP-v2/apps/api/internal/modules/production/application"
	"github.com/Chinsusu/ERP-v2/apps/api/internal/shared/audit"
	"github.com/Chinsusu/ERP-v2/apps/api/internal/shared/config"
)

func TestNewRuntimeSubcontractStoresFallsBackToPrototypeWithoutDatabaseURL(t *testing.T) {
	stores, closeStores, err := newRuntimeSubcontractStores(config.Config{}, audit.NewInMemoryLogStore())
	if err != nil {
		t.Fatalf("newRuntimeSubcontractStores() error = %v", err)
	}
	if closeStores != nil {
		t.Fatal("closeStores is not nil, want nil for prototype stores")
	}

	if _, ok := stores.orders.(*productionapp.PrototypeSubcontractOrderStore); !ok {
		t.Fatalf("orders store type = %T, want *PrototypeSubcontractOrderStore", stores.orders)
	}
	if _, ok := stores.materialTransfers.(*productionapp.PrototypeSubcontractMaterialTransferStore); !ok {
		t.Fatalf("material transfers store type = %T, want *PrototypeSubcontractMaterialTransferStore", stores.materialTransfers)
	}
	if _, ok := stores.sampleApprovals.(*productionapp.PrototypeSubcontractSampleApprovalStore); !ok {
		t.Fatalf("sample approvals store type = %T, want *PrototypeSubcontractSampleApprovalStore", stores.sampleApprovals)
	}
	if _, ok := stores.finishedGoodsReceipts.(*productionapp.PrototypeSubcontractFinishedGoodsReceiptStore); !ok {
		t.Fatalf("finished goods receipts store type = %T, want *PrototypeSubcontractFinishedGoodsReceiptStore", stores.finishedGoodsReceipts)
	}
	if _, ok := stores.factoryClaims.(*productionapp.PrototypeSubcontractFactoryClaimStore); !ok {
		t.Fatalf("factory claims store type = %T, want *PrototypeSubcontractFactoryClaimStore", stores.factoryClaims)
	}
	if _, ok := stores.paymentMilestones.(*productionapp.PrototypeSubcontractPaymentMilestoneStore); !ok {
		t.Fatalf("payment milestones store type = %T, want *PrototypeSubcontractPaymentMilestoneStore", stores.paymentMilestones)
	}
}

func TestNewRuntimeSubcontractStoresUsesPostgresAsPackageWhenDatabaseURLConfigured(t *testing.T) {
	stores, closeStores, err := newRuntimeSubcontractStores(
		config.Config{
			AppEnv:      "dev",
			DatabaseURL: "postgres://erp_dev:erp_dev@postgres:5432/erp_dev?sslmode=disable",
		},
		audit.NewInMemoryLogStore(),
	)
	if err != nil {
		t.Fatalf("newRuntimeSubcontractStores() error = %v", err)
	}
	if closeStores == nil {
		t.Fatal("closeStores = nil, want database close function")
	}
	defer func() {
		if err := closeStores(); err != nil {
			t.Fatalf("closeStores() error = %v", err)
		}
	}()

	if _, ok := stores.orders.(productionapp.PostgresSubcontractOrderStore); !ok {
		t.Fatalf("orders store type = %T, want PostgresSubcontractOrderStore", stores.orders)
	}
	if _, ok := stores.materialTransfers.(productionapp.PostgresSubcontractMaterialTransferStore); !ok {
		t.Fatalf("material transfers store type = %T, want PostgresSubcontractMaterialTransferStore", stores.materialTransfers)
	}
	if _, ok := stores.sampleApprovals.(productionapp.PostgresSubcontractSampleApprovalStore); !ok {
		t.Fatalf("sample approvals store type = %T, want PostgresSubcontractSampleApprovalStore", stores.sampleApprovals)
	}
	if _, ok := stores.finishedGoodsReceipts.(productionapp.PostgresSubcontractFinishedGoodsReceiptStore); !ok {
		t.Fatalf("finished goods receipts store type = %T, want PostgresSubcontractFinishedGoodsReceiptStore", stores.finishedGoodsReceipts)
	}
	if _, ok := stores.factoryClaims.(productionapp.PostgresSubcontractFactoryClaimStore); !ok {
		t.Fatalf("factory claims store type = %T, want PostgresSubcontractFactoryClaimStore", stores.factoryClaims)
	}
	if _, ok := stores.paymentMilestones.(productionapp.PostgresSubcontractPaymentMilestoneStore); !ok {
		t.Fatalf("payment milestones store type = %T, want PostgresSubcontractPaymentMilestoneStore", stores.paymentMilestones)
	}
}
