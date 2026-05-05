package application

import (
	"context"
	"testing"
	"time"

	inventorydomain "github.com/Chinsusu/ERP-v2/apps/api/internal/modules/inventory/domain"
	"github.com/Chinsusu/ERP-v2/apps/api/internal/shared/audit"
)

func TestStockTransferServicePostsOutAndInMovements(t *testing.T) {
	transferStore := NewPrototypeStockTransferStore()
	movementStore := NewInMemoryStockMovementStore()
	auditStore := audit.NewInMemoryLogStore()
	service := NewStockTransferService(transferStore, movementStore, auditStore).
		WithClock(func() time.Time {
			return time.Date(2026, 5, 5, 9, 0, 0, 0, time.UTC)
		})

	created, err := service.CreateStockTransfer(context.Background(), CreateStockTransferInput{
		ID:                       "transfer-hcm-stage-0001",
		TransferNo:               "ST-260505-0001",
		SourceWarehouseID:        "wh-main",
		SourceWarehouseCode:      "MAIN",
		DestinationWarehouseID:   "wh-stage",
		DestinationWarehouseCode: "STAGE",
		ReasonCode:               "staging",
		RequestedBy:              "warehouse-lead",
		RequestID:                "req-transfer-create",
		Lines: []CreateStockTransferLineInput{
			{
				ID:                      "transfer-line-serum",
				ItemID:                  "item-serum-30ml",
				SKU:                     "SERUM-30ML",
				SourceLocationID:        "bin-main-a01",
				SourceLocationCode:      "A-01",
				DestinationLocationID:   "bin-stage-a01",
				DestinationLocationCode: "ST-A01",
				Quantity:                "12",
				BaseUOMCode:             "PCS",
			},
		},
	})
	if err != nil {
		t.Fatalf("create transfer: %v", err)
	}

	for _, action := range []string{"submit", "approve", "post"} {
		updated, err := service.TransitionStockTransfer(context.Background(), StockTransferTransitionInput{
			ID:        created.StockTransfer.ID,
			ActorID:   "warehouse-lead",
			Action:    action,
			RequestID: "req-transfer-" + action,
		})
		if err != nil {
			t.Fatalf("%s transfer: %v", action, err)
		}
		created.StockTransfer = updated.StockTransfer
	}

	if created.StockTransfer.Status != inventorydomain.StockTransferStatusPosted {
		t.Fatalf("status = %q, want posted", created.StockTransfer.Status)
	}
	movements := movementStore.Movements()
	if len(movements) != 2 {
		t.Fatalf("movements = %+v, want two transfer movements", movements)
	}
	if movements[0].MovementType != inventorydomain.MovementTransferOut ||
		movements[1].MovementType != inventorydomain.MovementTransferIn {
		t.Fatalf("movement types = %s/%s, want transfer_out/transfer_in", movements[0].MovementType, movements[1].MovementType)
	}
	if movements[0].WarehouseID != "wh-main" || movements[1].WarehouseID != "wh-stage" {
		t.Fatalf("movement warehouses = %s/%s, want source/destination", movements[0].WarehouseID, movements[1].WarehouseID)
	}
	if movements[0].SourceDocType != "stock_transfer" || movements[1].SourceDocType != "stock_transfer" {
		t.Fatalf("source doc types = %s/%s, want stock_transfer", movements[0].SourceDocType, movements[1].SourceDocType)
	}
}

func TestWarehouseIssueServicePostsWarehouseIssueMovement(t *testing.T) {
	issueStore := NewPrototypeWarehouseIssueStore()
	movementStore := NewInMemoryStockMovementStore()
	auditStore := audit.NewInMemoryLogStore()
	service := NewWarehouseIssueService(issueStore, movementStore, auditStore).
		WithClock(func() time.Time {
			return time.Date(2026, 5, 5, 10, 0, 0, 0, time.UTC)
		})

	created, err := service.CreateWarehouseIssue(context.Background(), CreateWarehouseIssueInput{
		ID:              "issue-plan-0001",
		IssueNo:         "WI-260505-0001",
		WarehouseID:     "wh-main",
		WarehouseCode:   "MAIN",
		DestinationType: "factory",
		DestinationName: "Factory A",
		ReasonCode:      "production_plan_issue",
		RequestedBy:     "production-user",
		RequestID:       "req-issue-create",
		Lines: []CreateWarehouseIssueLineInput{
			{
				ID:                 "issue-line-material",
				ItemID:             "item-aci-bha",
				SKU:                "ACI_BHA",
				ItemName:           "ACID SALICYLIC",
				LocationID:         "bin-main-rm01",
				LocationCode:       "RM-01",
				Quantity:           "0.125",
				BaseUOMCode:        "KG",
				SourceDocumentType: "production_plan",
				SourceDocumentID:   "plan-260505-0001",
			},
		},
	})
	if err != nil {
		t.Fatalf("create issue: %v", err)
	}

	for _, action := range []string{"submit", "approve", "post"} {
		updated, err := service.TransitionWarehouseIssue(context.Background(), WarehouseIssueTransitionInput{
			ID:        created.WarehouseIssue.ID,
			ActorID:   "warehouse-lead",
			Action:    action,
			RequestID: "req-issue-" + action,
		})
		if err != nil {
			t.Fatalf("%s issue: %v", action, err)
		}
		created.WarehouseIssue = updated.WarehouseIssue
	}

	if created.WarehouseIssue.Status != inventorydomain.WarehouseIssueStatusPosted {
		t.Fatalf("status = %q, want posted", created.WarehouseIssue.Status)
	}
	movements := movementStore.Movements()
	if len(movements) != 1 {
		t.Fatalf("movements = %+v, want one warehouse issue movement", movements)
	}
	if movements[0].MovementType != inventorydomain.MovementWarehouseIssue {
		t.Fatalf("movement type = %s, want warehouse_issue", movements[0].MovementType)
	}
	if movements[0].SourceDocType != "warehouse_issue" || movements[0].SourceDocID != "issue-plan-0001" {
		t.Fatalf("source = %s/%s, want warehouse_issue/issue-plan-0001", movements[0].SourceDocType, movements[0].SourceDocID)
	}
}
