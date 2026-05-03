package main

import (
	"context"
	"testing"

	inventoryapp "github.com/Chinsusu/ERP-v2/apps/api/internal/modules/inventory/application"
	inventorydomain "github.com/Chinsusu/ERP-v2/apps/api/internal/modules/inventory/domain"
	masterdataapp "github.com/Chinsusu/ERP-v2/apps/api/internal/modules/masterdata/application"
	masterdatadomain "github.com/Chinsusu/ERP-v2/apps/api/internal/modules/masterdata/domain"
	salesapp "github.com/Chinsusu/ERP-v2/apps/api/internal/modules/sales/application"
	salesdomain "github.com/Chinsusu/ERP-v2/apps/api/internal/modules/sales/domain"
	shippingapp "github.com/Chinsusu/ERP-v2/apps/api/internal/modules/shipping/application"
	shippingdomain "github.com/Chinsusu/ERP-v2/apps/api/internal/modules/shipping/domain"
	"github.com/Chinsusu/ERP-v2/apps/api/internal/shared/audit"
)

func TestSprint2PrototypeTestDataSeed(t *testing.T) {
	ctx := context.Background()
	auditStore := audit.NewInMemoryLogStore()

	customers, _, err := masterdataapp.NewPrototypePartyCatalog(auditStore).ListCustomers(
		ctx,
		masterdatadomain.NewCustomerFilter("", "", "", 1, 100),
	)
	if err != nil {
		t.Fatalf("list customers: %v", err)
	}
	requireSeedCodes(t, "customers", customerCodes(customers), "CUS-DL-MINHANH", "CUS-DL-LINHCHI", "CUS-MP-SHOPEE", "CUS-MP-TIKTOK")

	requirePrototypeItemSKUs(t, ctx, masterdataapp.NewPrototypeItemCatalog(auditStore), "SERUM-30ML", "CREAM-50G", "TONER-100ML")

	batches, err := inventoryapp.NewPrototypeBatchCatalog(auditStore).ListBatches(
		ctx,
		inventorydomain.NewBatchFilter("", "", ""),
	)
	if err != nil {
		t.Fatalf("list batches: %v", err)
	}
	requireSeedCodes(t, "batches", batchNumbers(batches), "LOT-2604A", "LOT-2603B", "LOT-2604C")

	stockRows, err := inventoryapp.NewListAvailableStock(inventoryapp.NewPrototypeStockAvailabilityStore()).Execute(
		ctx,
		inventorydomain.NewAvailableStockFilter("", "", "", ""),
	)
	if err != nil {
		t.Fatalf("list available stock: %v", err)
	}
	requireSeedCodes(t, "stock skus", availableStockSKUs(stockRows), "SERUM-30ML", "CREAM-50G", "TONER-100ML")

	salesService, _ := newTestSalesOrderAPIService()
	orders, err := salesService.ListSalesOrders(ctx, salesapp.SalesOrderFilter{})
	if err != nil {
		t.Fatalf("list sales orders: %v", err)
	}
	if len(orders) != 20 {
		t.Fatalf("sales orders = %d, want 20", len(orders))
	}
	requireSeedStatuses(t, orderStatuses(orders),
		salesdomain.SalesOrderStatusDraft,
		salesdomain.SalesOrderStatusConfirmed,
		salesdomain.SalesOrderStatusReserved,
		salesdomain.SalesOrderStatusPicking,
		salesdomain.SalesOrderStatusPicked,
		salesdomain.SalesOrderStatusPacking,
		salesdomain.SalesOrderStatusPacked,
		salesdomain.SalesOrderStatusWaitingHandover,
		salesdomain.SalesOrderStatusHandedOver,
		salesdomain.SalesOrderStatusCancelled,
		salesdomain.SalesOrderStatusReservationFailed,
		salesdomain.SalesOrderStatusPickException,
		salesdomain.SalesOrderStatusPackException,
		salesdomain.SalesOrderStatusHandoverException,
	)

	activeCarriers, err := shippingapp.NewListCarriers(shippingapp.NewPrototypeCarrierCatalog()).Execute(
		ctx,
		shippingdomain.NewCarrierFilter("", shippingdomain.CarrierStatusActive),
	)
	if err != nil {
		t.Fatalf("list carriers: %v", err)
	}
	if len(activeCarriers) != 3 {
		t.Fatalf("active carriers = %d, want 3", len(activeCarriers))
	}
	requireSeedCodes(t, "active carriers", carrierCodes(activeCarriers), "GHN", "NJV", "VTP")

	manifests, err := shippingapp.NewListCarrierManifests(shippingapp.NewPrototypeCarrierManifestStore()).Execute(
		ctx,
		shippingdomain.NewCarrierManifestFilter("wh-hcm", "2026-04-26", "", ""),
	)
	if err != nil {
		t.Fatalf("list manifests: %v", err)
	}
	if len(manifests) != 2 {
		t.Fatalf("HCM manifests = %d, want 2", len(manifests))
	}
	requireSeedCodes(t, "HCM manifests", manifestIDs(manifests), "manifest-hcm-ghn-morning", "manifest-hcm-vtp-noon")
}

func customerCodes(customers []masterdatadomain.Customer) map[string]struct{} {
	codes := make(map[string]struct{}, len(customers))
	for _, customer := range customers {
		codes[customer.Code] = struct{}{}
	}

	return codes
}

func itemSKUCodes(items []masterdatadomain.Item) map[string]struct{} {
	codes := make(map[string]struct{}, len(items))
	for _, item := range items {
		codes[item.SKUCode] = struct{}{}
	}

	return codes
}

func requirePrototypeItemSKUs(t *testing.T, ctx context.Context, catalog *masterdataapp.ItemCatalog, expected ...string) {
	t.Helper()

	for _, sku := range expected {
		items, _, err := catalog.List(ctx, masterdatadomain.NewItemFilter(sku, "", "", 1, 20))
		if err != nil {
			t.Fatalf("list item %s: %v", sku, err)
		}
		if _, ok := itemSKUCodes(items)[sku]; !ok {
			t.Fatalf("items missing %s; got %v", sku, itemSKUCodes(items))
		}
	}
}

func batchNumbers(batches []inventorydomain.Batch) map[string]struct{} {
	codes := make(map[string]struct{}, len(batches))
	for _, batch := range batches {
		codes[batch.BatchNo] = struct{}{}
	}

	return codes
}

func availableStockSKUs(stockRows []inventorydomain.AvailableStockSnapshot) map[string]struct{} {
	codes := make(map[string]struct{}, len(stockRows))
	for _, row := range stockRows {
		codes[row.SKU] = struct{}{}
	}

	return codes
}

func orderStatuses(orders []salesdomain.SalesOrder) map[salesdomain.SalesOrderStatus]struct{} {
	statuses := make(map[salesdomain.SalesOrderStatus]struct{}, len(orders))
	for _, order := range orders {
		statuses[order.Status] = struct{}{}
	}

	return statuses
}

func carrierCodes(carriers []shippingdomain.Carrier) map[string]struct{} {
	codes := make(map[string]struct{}, len(carriers))
	for _, carrier := range carriers {
		codes[carrier.Code] = struct{}{}
	}

	return codes
}

func manifestIDs(manifests []shippingdomain.CarrierManifest) map[string]struct{} {
	ids := make(map[string]struct{}, len(manifests))
	for _, manifest := range manifests {
		ids[manifest.ID] = struct{}{}
	}

	return ids
}

func requireSeedCodes(t *testing.T, label string, actual map[string]struct{}, expected ...string) {
	t.Helper()
	for _, code := range expected {
		if _, ok := actual[code]; !ok {
			t.Fatalf("%s missing %s; got %+v", label, code, actual)
		}
	}
}

func requireSeedStatuses(
	t *testing.T,
	actual map[salesdomain.SalesOrderStatus]struct{},
	expected ...salesdomain.SalesOrderStatus,
) {
	t.Helper()
	for _, status := range expected {
		if _, ok := actual[status]; !ok {
			t.Fatalf("sales order status seed missing %s; got %+v", status, actual)
		}
	}
}
