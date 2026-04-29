package application

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/Chinsusu/ERP-v2/apps/api/internal/modules/inventory/domain"
	masterdataapp "github.com/Chinsusu/ERP-v2/apps/api/internal/modules/masterdata/application"
	purchasedomain "github.com/Chinsusu/ERP-v2/apps/api/internal/modules/purchase/domain"
	"github.com/Chinsusu/ERP-v2/apps/api/internal/shared/audit"
	"github.com/Chinsusu/ERP-v2/apps/api/internal/shared/decimal"
)

func TestWarehouseReceivingPostCreatesStockMovementAndAudit(t *testing.T) {
	service, movementStore, auditStore := newTestWarehouseReceivingService()

	result, err := service.PostWarehouseReceiving(context.Background(), WarehouseReceivingTransitionInput{
		ID:        "grn-hcm-260427-inspect",
		ActorID:   "user-warehouse-lead",
		RequestID: "req-receive-post",
	})
	if err != nil {
		t.Fatalf("post receiving: %v", err)
	}
	if result.Receipt.Status != domain.WarehouseReceivingStatusPosted {
		t.Fatalf("status = %q, want posted", result.Receipt.Status)
	}
	if movementStore.Count() != 1 || len(result.Receipt.StockMovements) != 1 {
		t.Fatalf("stock movements = store %d response %d, want 1", movementStore.Count(), len(result.Receipt.StockMovements))
	}
	movement := result.Receipt.StockMovements[0]
	if movement.SourceDocType != receivingSourceDocType ||
		movement.SourceDocID != "grn-hcm-260427-inspect" ||
		movement.StockStatus != domain.StockStatusAvailable {
		t.Fatalf("movement = %+v, want available purchase receipt linked to receiving", movement)
	}
	if result.AuditLogID == "" {
		t.Fatal("audit log id is empty")
	}
	logs, err := auditStore.List(context.Background(), audit.Query{Action: "inventory.receiving.posted"})
	if err != nil {
		t.Fatalf("list audit logs: %v", err)
	}
	if len(logs) != 1 || logs[0].AfterData["stock_movement_count"] != 1 {
		t.Fatalf("audit logs = %+v, want posted movement count", logs)
	}
}

func TestWarehouseReceivingRejectsDuplicatePost(t *testing.T) {
	service, movementStore, _ := newTestWarehouseReceivingService()

	_, err := service.PostWarehouseReceiving(context.Background(), WarehouseReceivingTransitionInput{
		ID:      "grn-hcm-260427-inspect",
		ActorID: "user-warehouse-lead",
	})
	if err != nil {
		t.Fatalf("first post receiving: %v", err)
	}
	_, err = service.PostWarehouseReceiving(context.Background(), WarehouseReceivingTransitionInput{
		ID:      "grn-hcm-260427-inspect",
		ActorID: "user-warehouse-lead",
	})
	if !errors.Is(err, domain.ErrReceivingAlreadyPosted) {
		t.Fatalf("err = %v, want already posted", err)
	}
	if movementStore.Count() != 1 {
		t.Fatalf("movement count = %d, want unchanged 1", movementStore.Count())
	}
}

func TestWarehouseReceivingRejectsInvalidLocation(t *testing.T) {
	service, _, _ := newTestWarehouseReceivingService()

	_, err := service.CreateWarehouseReceiving(context.Background(), CreateWarehouseReceivingInput{
		WarehouseID:      "wh-hcm-fg",
		LocationID:       "loc-hcm-rm-recv-01",
		ReferenceDocType: "purchase_order",
		ReferenceDocID:   "PO-260427-0100",
		SupplierID:       "supplier-local",
		DeliveryNoteNo:   "DN-260427-0100",
		Lines: []CreateWarehouseReceivingLineInput{
			{
				PurchaseOrderLineID: "po-line-260427-0100-001",
				ItemID:              "item-cream-50g",
				SKU:                 "CREAM-50G",
				BatchID:             "batch-cream-2603b",
				Quantity:            "12",
				BaseUOMCode:         "EA",
				PackagingStatus:     "intact",
			},
		},
		ActorID: "user-warehouse-lead",
	})
	if !errors.Is(err, ErrReceivingInvalidLocation) {
		t.Fatalf("err = %v, want invalid location", err)
	}
}

func TestWarehouseReceivingPostRequiresBatchAndQCData(t *testing.T) {
	service, _, _ := newTestWarehouseReceivingService()
	ctx := context.Background()

	created, err := service.CreateWarehouseReceiving(ctx, CreateWarehouseReceivingInput{
		ID:               "grn-missing-batch-qc",
		ReceiptNo:        "GRN-260427-0999",
		WarehouseID:      "wh-hcm-fg",
		LocationID:       "loc-hcm-fg-recv-01",
		ReferenceDocType: "manual_receiving",
		ReferenceDocID:   "MANUAL-260427-0999",
		SupplierID:       "supplier-local",
		DeliveryNoteNo:   "DN-260427-0999",
		Lines: []CreateWarehouseReceivingLineInput{
			{
				ID:                  "line-missing-batch-qc",
				PurchaseOrderLineID: "po-line-260427-0999-001",
				ItemID:              "item-cream-50g",
				SKU:                 "CREAM-50G",
				Quantity:            "12",
				BaseUOMCode:         "EA",
				PackagingStatus:     "intact",
			},
		},
		ActorID: "user-warehouse-lead",
	})
	if err != nil {
		t.Fatalf("create receiving: %v", err)
	}
	submitted, err := service.SubmitWarehouseReceiving(ctx, WarehouseReceivingTransitionInput{
		ID:      created.Receipt.ID,
		ActorID: "user-warehouse-lead",
	})
	if err != nil {
		t.Fatalf("submit receiving: %v", err)
	}
	_, err = service.MarkWarehouseReceivingInspectReady(ctx, WarehouseReceivingTransitionInput{
		ID:      submitted.Receipt.ID,
		ActorID: "user-qa",
	})
	if err != nil {
		t.Fatalf("mark inspect ready: %v", err)
	}
	_, err = service.PostWarehouseReceiving(ctx, WarehouseReceivingTransitionInput{
		ID:      created.Receipt.ID,
		ActorID: "user-warehouse-lead",
	})
	if !errors.Is(err, domain.ErrReceivingMissingBatchQCData) {
		t.Fatalf("err = %v, want missing batch/qc data", err)
	}
}

func TestWarehouseReceivingRejectsInvalidPurchaseOrderLinkedReceipt(t *testing.T) {
	service, _, _ := newTestWarehouseReceivingService()
	ctx := context.Background()

	tests := []struct {
		name    string
		mutate  func(*CreateWarehouseReceivingInput)
		wantErr error
	}{
		{
			name: "supplier mismatch",
			mutate: func(input *CreateWarehouseReceivingInput) {
				input.SupplierID = "supplier-other"
			},
			wantErr: ErrReceivingPurchaseOrderMismatch,
		},
		{
			name: "wrong po line",
			mutate: func(input *CreateWarehouseReceivingInput) {
				input.Lines[0].PurchaseOrderLineID = "po-line-missing"
			},
			wantErr: ErrReceivingPurchaseOrderMismatch,
		},
		{
			name: "uom mismatch",
			mutate: func(input *CreateWarehouseReceivingInput) {
				input.Lines[0].UOMCode = "BOX"
			},
			wantErr: ErrReceivingPurchaseOrderMismatch,
		},
		{
			name: "over receive",
			mutate: func(input *CreateWarehouseReceivingInput) {
				input.Lines[0].Quantity = "999"
			},
			wantErr: ErrReceivingQuantityExceedsPurchaseOrder,
		},
		{
			name: "missing batch",
			mutate: func(input *CreateWarehouseReceivingInput) {
				input.Lines[0].BatchID = ""
			},
			wantErr: domain.ErrReceivingRequiredField,
		},
		{
			name: "draft purchase order",
			mutate: func(input *CreateWarehouseReceivingInput) {
				input.ReferenceDocID = "po-draft-receiving-test"
			},
			wantErr: ErrReceivingPurchaseOrderInvalidState,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			input := validPurchaseOrderReceivingInputForTest()
			tt.mutate(&input)

			_, err := service.CreateWarehouseReceiving(ctx, input)
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("err = %v, want %v", err, tt.wantErr)
			}
		})
	}
}

func TestWarehouseReceivingCreateHydratesInboundFieldsFromBatch(t *testing.T) {
	service, _, auditStore := newTestWarehouseReceivingService()

	result, err := service.CreateWarehouseReceiving(context.Background(), CreateWarehouseReceivingInput{
		ID:               "grn-hydrate-inbound-fields",
		ReceiptNo:        "GRN-260427-HYDRATE",
		WarehouseID:      "wh-hcm-fg",
		LocationID:       "loc-hcm-fg-recv-01",
		ReferenceDocType: "purchase_order",
		ReferenceDocID:   "PO-260427-HYDRATE",
		SupplierID:       "supplier-local",
		DeliveryNoteNo:   "dn-260427-hydrate",
		Lines: []CreateWarehouseReceivingLineInput{
			{
				PurchaseOrderLineID: "po-line-260427-hydrate-001",
				BatchID:             "batch-cream-2603b",
				Quantity:            "12",
				UOMCode:             "ea",
				BaseUOMCode:         "ea",
				PackagingStatus:     "intact",
			},
		},
		ActorID:   "user-warehouse-lead",
		RequestID: "req-receive-hydrate",
	})
	if err != nil {
		t.Fatalf("create receiving: %v", err)
	}
	if result.Receipt.DeliveryNoteNo != "DN-260427-HYDRATE" {
		t.Fatalf("delivery note = %q, want normalized", result.Receipt.DeliveryNoteNo)
	}
	line := result.Receipt.Lines[0]
	if line.ID != "line-001" ||
		line.ItemID != "item-cream-50g" ||
		line.SKU != "CREAM-50G" ||
		line.BatchNo != "LOT-2603B" ||
		line.LotNo != "LOT-2603B" ||
		line.ExpiryDate.Format("2006-01-02") != "2028-03-01" ||
		line.UOMCode.String() != "EA" ||
		line.PackagingStatus != domain.ReceivingPackagingStatusIntact ||
		line.QCStatus != domain.QCStatusPass {
		t.Fatalf("line = %+v, want hydrated batch, expiry, UOM, packaging, QC", line)
	}

	logs, err := auditStore.List(context.Background(), audit.Query{Action: "inventory.receiving.created"})
	if err != nil {
		t.Fatalf("list audit logs: %v", err)
	}
	if len(logs) != 1 || logs[0].AfterData["delivery_note_no"] != "DN-260427-HYDRATE" {
		t.Fatalf("audit logs = %+v, want delivery note audit data", logs)
	}
}

func newTestWarehouseReceivingService() (WarehouseReceivingService, *InMemoryStockMovementStore, *audit.InMemoryLogStore) {
	auditStore := audit.NewInMemoryLogStore()
	movementStore := NewInMemoryStockMovementStore()
	service := NewWarehouseReceivingService(
		NewPrototypeWarehouseReceivingStore(),
		masterdataapp.NewPrototypeWarehouseLocationCatalog(auditStore),
		NewPrototypeBatchCatalog(auditStore),
		movementStore,
		auditStore,
	).WithPurchaseOrderReader(staticWarehouseReceivingPurchaseOrderReader{
		orders: testWarehouseReceivingPurchaseOrders(),
	})
	service.clock = func() time.Time {
		return time.Date(2026, 4, 27, 11, 0, 0, 0, time.UTC)
	}

	return service, movementStore, auditStore
}

func validPurchaseOrderReceivingInputForTest() CreateWarehouseReceivingInput {
	return CreateWarehouseReceivingInput{
		ID:               "grn-po-validation",
		ReceiptNo:        "GRN-260427-VALIDATE",
		WarehouseID:      "wh-hcm-fg",
		LocationID:       "loc-hcm-fg-recv-01",
		ReferenceDocType: "purchase_order",
		ReferenceDocID:   "PO-260427-HYDRATE",
		SupplierID:       "supplier-local",
		DeliveryNoteNo:   "DN-260427-VALIDATE",
		Lines: []CreateWarehouseReceivingLineInput{
			{
				PurchaseOrderLineID: "po-line-260427-hydrate-001",
				BatchID:             "batch-cream-2603b",
				Quantity:            "12",
				UOMCode:             "EA",
				BaseUOMCode:         "EA",
				PackagingStatus:     "intact",
			},
		},
		ActorID: "user-warehouse-lead",
	}
}

type staticWarehouseReceivingPurchaseOrderReader struct {
	orders map[string]purchasedomain.PurchaseOrder
}

func (r staticWarehouseReceivingPurchaseOrderReader) GetPurchaseOrder(
	_ context.Context,
	id string,
) (purchasedomain.PurchaseOrder, error) {
	order, ok := r.orders[id]
	if !ok {
		return purchasedomain.PurchaseOrder{}, errors.New("purchase order not found")
	}

	return order.Clone(), nil
}

func testWarehouseReceivingPurchaseOrders() map[string]purchasedomain.PurchaseOrder {
	return map[string]purchasedomain.PurchaseOrder{
		"PO-260427-HYDRATE":       approvedPurchaseOrderForReceiving("PO-260427-HYDRATE", "po-line-260427-hydrate-001", "item-cream-50g", "CREAM-50G", "Moisturizing Cream", "100"),
		"po-draft-receiving-test": draftPurchaseOrderForReceiving("po-draft-receiving-test", "po-line-260427-hydrate-001", "item-cream-50g", "CREAM-50G", "Moisturizing Cream", "100"),
		"PO-260427-0001":          approvedPurchaseOrderForReceiving("PO-260427-0001", "po-line-260427-0001-001", "item-serum-30ml", "SERUM-30ML", "Vitamin C Serum", "100"),
		"PO-260427-0002":          approvedPurchaseOrderForReceiving("PO-260427-0002", "po-line-260427-0002-001", "item-serum-30ml", "SERUM-30ML", "Vitamin C Serum", "100"),
		"PO-260427-0003":          approvedPurchaseOrderForReceiving("PO-260427-0003", "po-line-260427-0003-001", "item-cream-50g", "CREAM-50G", "Moisturizing Cream", "100"),
	}
}

func approvedPurchaseOrderForReceiving(
	id string,
	lineID string,
	itemID string,
	sku string,
	itemName string,
	orderedQty string,
) purchasedomain.PurchaseOrder {
	order := draftPurchaseOrderForReceiving(id, lineID, itemID, sku, itemName, orderedQty)
	submitted, err := order.Submit("user-purchase-ops", time.Date(2026, 4, 27, 9, 30, 0, 0, time.UTC))
	if err != nil {
		panic(err)
	}
	approved, err := submitted.Approve("user-purchase-ops", time.Date(2026, 4, 27, 10, 0, 0, 0, time.UTC))
	if err != nil {
		panic(err)
	}

	return approved
}

func draftPurchaseOrderForReceiving(
	id string,
	lineID string,
	itemID string,
	sku string,
	itemName string,
	orderedQty string,
) purchasedomain.PurchaseOrder {
	order, err := purchasedomain.NewPurchaseOrderDocument(purchasedomain.NewPurchaseOrderDocumentInput{
		ID:            id,
		OrgID:         "org-my-pham",
		PONo:          id,
		SupplierID:    "supplier-local",
		SupplierCode:  "SUP-LOCAL",
		SupplierName:  "Local Supplier",
		WarehouseID:   "wh-hcm-fg",
		WarehouseCode: "WH-HCM-FG",
		ExpectedDate:  "2026-04-29",
		CurrencyCode:  "VND",
		Lines: []purchasedomain.NewPurchaseOrderLineInput{
			{
				ID:           lineID,
				LineNo:       1,
				ItemID:       itemID,
				SKUCode:      sku,
				ItemName:     itemName,
				OrderedQty:   decimal.MustQuantity(orderedQty),
				UOMCode:      "EA",
				BaseUOMCode:  "EA",
				UnitPrice:    decimal.MustUnitPrice("1"),
				CurrencyCode: "VND",
			},
		},
		CreatedAt: time.Date(2026, 4, 27, 9, 0, 0, 0, time.UTC),
		CreatedBy: "user-purchase-ops",
	})
	if err != nil {
		panic(err)
	}

	return order
}
