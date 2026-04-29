package application

import (
	"context"
	"testing"
	"time"

	inventorydomain "github.com/Chinsusu/ERP-v2/apps/api/internal/modules/inventory/domain"
	productiondomain "github.com/Chinsusu/ERP-v2/apps/api/internal/modules/production/domain"
	"github.com/Chinsusu/ERP-v2/apps/api/internal/shared/decimal"
)

func TestSubcontractMaterialTransferServiceBuildIssueCreatesTransferAndMovementRequests(t *testing.T) {
	fixedNow := time.Date(2026, 4, 29, 9, 0, 0, 0, time.UTC)
	service := NewSubcontractMaterialTransferService()
	service.clock = func() time.Time { return fixedNow }
	order := subcontractMaterialTransferTestOrder(t)

	result, err := service.BuildIssue(context.Background(), BuildSubcontractMaterialTransferInput{
		ID:                  "smt_001",
		TransferNo:          "SMT-20260429-001",
		Order:               order,
		SourceWarehouseID:   "wh_main",
		SourceWarehouseCode: "WH-HCM",
		HandoverBy:          "warehouse-user",
		HandoverAt:          fixedNow.Add(30 * time.Minute),
		ReceivedBy:          "factory-receiver",
		ReceiverContact:     "0988000111",
		VehicleNo:           "51A-12345",
		ActorID:             "warehouse-user",
		Lines: []BuildSubcontractMaterialTransferLineInput{
			{
				OrderMaterialLineID: "sco_mat_001",
				IssueQty:            "10",
				UOMCode:             "KG",
				BatchID:             "batch_base_001",
				BatchNo:             "BASE-LOT-001",
				SourceBinID:         "bin_a01",
			},
			{
				OrderMaterialLineID: "sco_mat_002",
				IssueQty:            "1000",
				UOMCode:             "PCS",
				SourceBinID:         "bin_b01",
			},
		},
		Evidence: []BuildSubcontractMaterialTransferEvidenceInput{
			{
				ID:           "evidence_001",
				EvidenceType: "handover",
				FileName:     "handover.pdf",
				ObjectKey:    "subcontract/smt_001/handover.pdf",
			},
		},
	})
	if err != nil {
		t.Fatalf("build material transfer issue: %v", err)
	}

	if result.Transfer.Status != productiondomain.SubcontractMaterialTransferStatusSentToFactory {
		t.Fatalf("transfer status = %q, want sent to factory", result.Transfer.Status)
	}
	if result.UpdatedOrder.Status != productiondomain.SubcontractOrderStatusMaterialsIssued {
		t.Fatalf("order status = %q, want materials issued", result.UpdatedOrder.Status)
	}
	if got, want := result.UpdatedOrder.MaterialLines[0].BaseIssuedQty.String(), "10000.000000"; got != want {
		t.Fatalf("updated base issued qty = %s, want %s", got, want)
	}
	if len(result.StockMovements) != 2 {
		t.Fatalf("stock movement count = %d, want 2", len(result.StockMovements))
	}
	firstMovement := result.StockMovements[0]
	if firstMovement.MovementType != inventorydomain.MovementSubcontractIssue ||
		firstMovement.StockStatus != inventorydomain.StockStatusSubcontractIssued {
		t.Fatalf("movement = %+v, want subcontract issue movement", firstMovement)
	}
	if firstMovement.SourceDocID != result.Transfer.ID || firstMovement.SourceDocLineID != result.Transfer.Lines[0].ID {
		t.Fatalf("movement source = %s/%s, want transfer source", firstMovement.SourceDocID, firstMovement.SourceDocLineID)
	}
	if got, want := firstMovement.Quantity.String(), "10000.000000"; got != want {
		t.Fatalf("movement base qty = %s, want %s", got, want)
	}
	if got, want := firstMovement.SourceQuantity.String(), "10.000000"; got != want {
		t.Fatalf("movement source qty = %s, want %s", got, want)
	}
}

func TestSubcontractMaterialTransferServiceBuildIssueSupportsPartialTransfer(t *testing.T) {
	fixedNow := time.Date(2026, 4, 29, 9, 0, 0, 0, time.UTC)
	service := NewSubcontractMaterialTransferService()
	service.clock = func() time.Time { return fixedNow }
	order := subcontractMaterialTransferTestOrder(t)

	result, err := service.BuildIssue(context.Background(), BuildSubcontractMaterialTransferInput{
		ID:                "smt_partial_001",
		Order:             order,
		SourceWarehouseID: "wh_main",
		HandoverBy:        "warehouse-user",
		ReceivedBy:        "factory-receiver",
		ActorID:           "warehouse-user",
		Lines: []BuildSubcontractMaterialTransferLineInput{
			{
				OrderMaterialLineID: "sco_mat_001",
				IssueQty:            "5",
				UOMCode:             "KG",
				BatchID:             "batch_base_001",
			},
		},
	})
	if err != nil {
		t.Fatalf("build partial material transfer: %v", err)
	}

	if result.Transfer.Status != productiondomain.SubcontractMaterialTransferStatusPartiallySent {
		t.Fatalf("transfer status = %q, want partially sent", result.Transfer.Status)
	}
	if result.UpdatedOrder.Status != productiondomain.SubcontractOrderStatusFactoryConfirmed {
		t.Fatalf("order status = %q, want factory confirmed until fully issued", result.UpdatedOrder.Status)
	}
	if got, want := result.StockMovements[0].Quantity.String(), "5000.000000"; got != want {
		t.Fatalf("partial movement base qty = %s, want %s", got, want)
	}
}

func subcontractMaterialTransferTestOrder(t *testing.T) productiondomain.SubcontractOrder {
	t.Helper()

	order, err := productiondomain.NewSubcontractOrderDocument(productiondomain.NewSubcontractOrderDocumentInput{
		ID:                  "sco_001",
		OrgID:               "org-my-pham",
		OrderNo:             "SCO-20260429-001",
		FactoryID:           "fac_001",
		FactoryCode:         "FAC-HCM-01",
		FactoryName:         "HCM Cosmetics Factory",
		FinishedItemID:      "item_serum",
		FinishedSKUCode:     "FG-SERUM-001",
		FinishedItemName:    "Brightening Serum",
		PlannedQty:          decimal.MustQuantity("1000"),
		UOMCode:             "PCS",
		BasePlannedQty:      decimal.MustQuantity("1000"),
		BaseUOMCode:         "PCS",
		ConversionFactor:    decimal.MustQuantity("1"),
		CurrencyCode:        "VND",
		SpecSummary:         "30ml bottle, printed box, approved label",
		SampleRequired:      true,
		TargetStartDate:     "2026-05-02",
		ExpectedReceiptDate: "2026-05-12",
		CreatedAt:           time.Date(2026, 4, 29, 8, 0, 0, 0, time.UTC),
		CreatedBy:           "subcontract-user",
		MaterialLines: []productiondomain.NewSubcontractMaterialLineInput{
			{
				ID:               "sco_mat_001",
				LineNo:           1,
				ItemID:           "item_base",
				SKUCode:          "RM-BASE-001",
				ItemName:         "Serum Base",
				PlannedQty:       decimal.MustQuantity("10"),
				UOMCode:          "KG",
				BasePlannedQty:   decimal.MustQuantity("10000"),
				BaseUOMCode:      "G",
				ConversionFactor: decimal.MustQuantity("1000"),
				UnitCost:         decimal.MustUnitCost("150000"),
				CurrencyCode:     "VND",
				LotTraceRequired: true,
			},
			{
				ID:               "sco_mat_002",
				LineNo:           2,
				ItemID:           "item_box",
				SKUCode:          "PK-BOX-030",
				ItemName:         "Printed Box 30ml",
				PlannedQty:       decimal.MustQuantity("1000"),
				UOMCode:          "PCS",
				BasePlannedQty:   decimal.MustQuantity("1000"),
				BaseUOMCode:      "PCS",
				ConversionFactor: decimal.MustQuantity("1"),
				UnitCost:         decimal.MustUnitCost("700"),
				CurrencyCode:     "VND",
			},
		},
	})
	if err != nil {
		t.Fatalf("new subcontract order: %v", err)
	}
	order, err = order.Submit("subcontract-user", time.Now())
	if err != nil {
		t.Fatalf("submit: %v", err)
	}
	order, err = order.Approve("operations-lead", time.Now())
	if err != nil {
		t.Fatalf("approve: %v", err)
	}
	order, err = order.ConfirmFactory("factory-coordinator", time.Now())
	if err != nil {
		t.Fatalf("confirm factory: %v", err)
	}

	return order
}
