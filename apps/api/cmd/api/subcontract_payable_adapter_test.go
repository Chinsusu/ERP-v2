package main

import (
	"context"
	"testing"

	financeapp "github.com/Chinsusu/ERP-v2/apps/api/internal/modules/finance/application"
	productionapp "github.com/Chinsusu/ERP-v2/apps/api/internal/modules/production/application"
	productiondomain "github.com/Chinsusu/ERP-v2/apps/api/internal/modules/production/domain"
	"github.com/Chinsusu/ERP-v2/apps/api/internal/shared/audit"
	"github.com/Chinsusu/ERP-v2/apps/api/internal/shared/decimal"
)

func TestSubcontractSupplierPayableAdapterCreatesFinancePayable(t *testing.T) {
	ctx := context.Background()
	auditStore := audit.NewInMemoryLogStore()
	store := financeapp.NewPrototypeSupplierPayableStore()
	service := financeapp.NewSupplierPayableService(store, auditStore)
	adapter := subcontractSupplierPayableAdapter{service: service}

	result, err := adapter.CreateSubcontractPayable(ctx, productionapp.CreateSubcontractPayableInput{
		SubcontractOrder: productiondomain.SubcontractOrder{
			ID:           "sco-adapter-0001",
			OrderNo:      "SCO-ADAPTER-0001",
			FactoryID:    "factory-bd-002",
			FactoryCode:  "FACT-BD-002",
			FactoryName:  "Binh Duong Gia Cong",
			CurrencyCode: decimal.MustCurrencyCode("VND"),
		},
		Milestone: productiondomain.SubcontractPaymentMilestone{
			ID:                 "spm-adapter-0001",
			MilestoneNo:        "SPM-ADAPTER-0001",
			Amount:             decimal.MustMoneyAmount("6800000.00"),
			CurrencyCode:       decimal.MustCurrencyCode("VND"),
			SubcontractOrderID: "sco-adapter-0001",
		},
		ActorID:   "finance-user",
		RequestID: "req-subcontract-payable-adapter",
	})
	if err != nil {
		t.Fatalf("create subcontract payable: %v", err)
	}

	if result.PayableID != "ap-spm-adapter-0001" ||
		result.PayableNo != "AP-SPM-ADAPTER-0001" ||
		result.AuditLogID == "" {
		t.Fatalf("result = %+v, want created AP identifiers and audit", result)
	}
	payable, err := service.GetSupplierPayable(ctx, result.PayableID)
	if err != nil {
		t.Fatalf("get payable: %v", err)
	}
	if payable.SupplierID != "factory-bd-002" ||
		payable.SourceDocument.Type != "subcontract_payment_milestone" ||
		payable.SourceDocument.ID != "spm-adapter-0001" ||
		payable.TotalAmount.String() != "6800000.00" ||
		len(payable.Lines) != 1 ||
		payable.Lines[0].SourceDocument.Type != "subcontract_order" ||
		payable.Lines[0].Amount.String() != "6800000.00" {
		t.Fatalf("payable = %+v, want AP sourced from subcontract milestone/order", payable)
	}
	logs, err := auditStore.List(ctx, audit.Query{Action: "finance.supplier_payable.created"})
	if err != nil {
		t.Fatalf("list payable audit logs: %v", err)
	}
	if len(logs) != 1 || logs[0].EntityID != result.PayableID {
		t.Fatalf("audit logs = %+v, want payable creation audit", logs)
	}
}
