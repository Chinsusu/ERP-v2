package main

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	inventoryapp "github.com/Chinsusu/ERP-v2/apps/api/internal/modules/inventory/application"
	masterdataapp "github.com/Chinsusu/ERP-v2/apps/api/internal/modules/masterdata/application"
	returnsapp "github.com/Chinsusu/ERP-v2/apps/api/internal/modules/returns/application"
	salesapp "github.com/Chinsusu/ERP-v2/apps/api/internal/modules/sales/application"
	salesdomain "github.com/Chinsusu/ERP-v2/apps/api/internal/modules/sales/domain"
	shippingapp "github.com/Chinsusu/ERP-v2/apps/api/internal/modules/shipping/application"
	shippingdomain "github.com/Chinsusu/ERP-v2/apps/api/internal/modules/shipping/domain"
	"github.com/Chinsusu/ERP-v2/apps/api/internal/shared/audit"
	"github.com/Chinsusu/ERP-v2/apps/api/internal/shared/auth"
	"github.com/Chinsusu/ERP-v2/apps/api/internal/shared/response"
)

func TestOrderFulfillmentPermissionRegressionSmoke(t *testing.T) {
	authConfig := smokeAuthConfig()
	harness := newPermissionAuditHarness()

	warehouseCreateReq := smokeRequestAsRole(
		httptest.NewRequest(http.MethodPost, "/api/v1/sales-orders", bytes.NewBufferString(`{
			"id": "so-permission-denied",
			"order_no": "SO-PERMISSION-DENIED",
			"customer_id": "cus-dl-minh-anh",
			"channel": "B2B",
			"warehouse_id": "wh-hcm-fg",
			"order_date": "2026-04-28",
			"currency_code": "VND",
			"lines": [{"item_id":"item-serum-30ml","ordered_qty":"1","uom_code":"EA","unit_price":"125000"}]
		}`)),
		authConfig,
		auth.RoleWarehouseStaff,
	)
	warehouseCreateRec := httptest.NewRecorder()

	salesOrdersHandler(harness.salesService).ServeHTTP(warehouseCreateRec, warehouseCreateReq)

	assertForbidden(t, warehouseCreateRec)

	salesCreateReq := smokeRequestAsRole(
		httptest.NewRequest(http.MethodPost, "/api/v1/sales-orders", bytes.NewBufferString(`{
			"id": "so-permission-sales",
			"order_no": "SO-PERMISSION-SALES",
			"customer_id": "cus-dl-minh-anh",
			"channel": "B2B",
			"warehouse_id": "wh-hcm-fg",
			"order_date": "2026-04-28",
			"currency_code": "VND",
			"lines": [{"item_id":"item-serum-30ml","ordered_qty":"1","uom_code":"EA","unit_price":"125000"}]
		}`)),
		authConfig,
		auth.RoleSalesOps,
	)
	salesCreateReq.Header.Set(response.HeaderRequestID, "req-permission-sales-create")
	salesCreateRec := httptest.NewRecorder()

	salesOrdersHandler(harness.salesService).ServeHTTP(salesCreateRec, salesCreateReq)

	if salesCreateRec.Code != http.StatusCreated {
		t.Fatalf("sales create status = %d, want %d: %s", salesCreateRec.Code, http.StatusCreated, salesCreateRec.Body.String())
	}
	salesConfirmReq := smokeRequestAsRole(
		httptest.NewRequest(http.MethodPost, "/api/v1/sales-orders/so-permission-sales/confirm", bytes.NewBufferString(`{"expected_version":1}`)),
		authConfig,
		auth.RoleSalesOps,
	)
	salesConfirmReq.SetPathValue("sales_order_id", "so-permission-sales")
	salesConfirmReq.Header.Set(response.HeaderRequestID, "req-permission-sales-confirm")
	salesConfirmRec := httptest.NewRecorder()

	salesOrderConfirmHandler(harness.salesService).ServeHTTP(salesConfirmRec, salesConfirmReq)

	if salesConfirmRec.Code != http.StatusOK {
		t.Fatalf("sales confirm status = %d, want %d: %s", salesConfirmRec.Code, http.StatusOK, salesConfirmRec.Body.String())
	}

	pickStore := shippingapp.NewPrototypePickTaskStore(mustPrototypePickTask())
	pickAuditStore := audit.NewInMemoryLogStore()
	salesPickReq := smokeRequestAsRole(
		httptest.NewRequest(http.MethodPost, "/api/v1/pick-tasks/pick-so-260428-0001/start", nil),
		authConfig,
		auth.RoleSalesOps,
	)
	salesPickReq.SetPathValue("pick_task_id", "pick-so-260428-0001")
	salesPickRec := httptest.NewRecorder()

	startPickTaskHandler(shippingapp.NewStartPickTask(pickStore, pickAuditStore)).ServeHTTP(salesPickRec, salesPickReq)

	assertForbidden(t, salesPickRec)

	warehousePickReq := smokeRequestAsRole(
		httptest.NewRequest(http.MethodPost, "/api/v1/pick-tasks/pick-so-260428-0001/start", nil),
		authConfig,
		auth.RoleWarehouseLead,
	)
	warehousePickReq.SetPathValue("pick_task_id", "pick-so-260428-0001")
	warehousePickReq.Header.Set(response.HeaderRequestID, "req-permission-warehouse-pick")
	warehousePickRec := httptest.NewRecorder()

	startPickTaskHandler(shippingapp.NewStartPickTask(pickStore, pickAuditStore)).ServeHTTP(warehousePickRec, warehousePickReq)

	if warehousePickRec.Code != http.StatusOK {
		t.Fatalf("warehouse pick status = %d, want %d: %s", warehousePickRec.Code, http.StatusOK, warehousePickRec.Body.String())
	}

	salesManifestReq := smokeRequestAsRole(
		httptest.NewRequest(http.MethodPost, "/api/v1/shipping/manifests", bytes.NewBufferString(`{
			"id": "manifest-permission-sales-denied",
			"carrier_code": "GHN",
			"warehouse_id": "wh-hcm-fg",
			"warehouse_code": "WH-HCM-FG",
			"date": "2026-04-28"
		}`)),
		authConfig,
		auth.RoleSalesOps,
	)
	salesManifestRec := httptest.NewRecorder()

	carrierManifestsHandler(
		shippingapp.NewListCarrierManifests(harness.manifestStore),
		shippingapp.NewCreateCarrierManifest(harness.manifestStore, harness.auditStore),
	).ServeHTTP(salesManifestRec, salesManifestReq)

	assertForbidden(t, salesManifestRec)

	adminManifestReq := smokeRequestAsRole(
		httptest.NewRequest(http.MethodPost, "/api/v1/shipping/manifests", bytes.NewBufferString(`{
			"id": "manifest-permission-admin",
			"carrier_code": "GHN",
			"warehouse_id": "wh-hcm-fg",
			"warehouse_code": "WH-HCM-FG",
			"date": "2026-04-28"
		}`)),
		authConfig,
		auth.RoleERPAdmin,
	)
	adminManifestReq.Header.Set(response.HeaderRequestID, "req-permission-admin-manifest")
	adminManifestRec := httptest.NewRecorder()

	carrierManifestsHandler(
		shippingapp.NewListCarrierManifests(harness.manifestStore),
		shippingapp.NewCreateCarrierManifest(harness.manifestStore, harness.auditStore),
	).ServeHTTP(adminManifestRec, adminManifestReq)

	if adminManifestRec.Code != http.StatusCreated {
		t.Fatalf("admin manifest status = %d, want %d: %s", adminManifestRec.Code, http.StatusCreated, adminManifestRec.Body.String())
	}
}

func TestReturnsAndShiftClosingPermissionAuditRegressionSmoke(t *testing.T) {
	authConfig := smokeAuthConfig()
	auditStore := audit.NewInMemoryLogStore()
	returnStore := returnsapp.NewPrototypeReturnReceiptStore()
	movementStore := inventoryapp.NewInMemoryStockMovementStore()
	receiveService := returnsapp.NewReceiveReturn(returnStore, auditStore)
	inspectService := returnsapp.NewInspectReturn(returnStore, auditStore)
	dispositionService := returnsapp.NewApplyReturnDisposition(returnStore, movementStore, auditStore)

	staffScanReq := smokeRequestAsRole(
		httptest.NewRequest(http.MethodPost, "/api/v1/returns/scan", bytes.NewBufferString(`{
			"warehouse_id": "wh-hcm",
			"warehouse_code": "HCM",
			"source": "CARRIER",
			"code": "GHN260426001",
			"package_condition": "sealed bag"
		}`)),
		authConfig,
		auth.RoleWarehouseStaff,
	)
	staffScanRec := httptest.NewRecorder()

	returnScanHandler(receiveService).ServeHTTP(staffScanRec, staffScanReq)

	assertForbidden(t, staffScanRec)
	assertNoAuditAction(t, auditStore, "returns.receipt.created")

	leadScanReq := smokeRequestAsRole(
		httptest.NewRequest(http.MethodPost, "/api/v1/returns/scan", bytes.NewBufferString(`{
			"warehouse_id": "wh-hcm",
			"warehouse_code": "HCM",
			"source": "CARRIER",
			"code": "GHN260426001",
			"package_condition": "sealed bag"
		}`)),
		authConfig,
		auth.RoleWarehouseLead,
	)
	leadScanReq.Header.Set(response.HeaderRequestID, "req-regression-return-scan")
	leadScanRec := httptest.NewRecorder()

	returnScanHandler(receiveService).ServeHTTP(leadScanRec, leadScanReq)

	if leadScanRec.Code != http.StatusCreated {
		t.Fatalf("lead scan status = %d, want %d: %s", leadScanRec.Code, http.StatusCreated, leadScanRec.Body.String())
	}
	received := decodeSmokeSuccess[returnReceiptResponse](t, leadScanRec).Data
	if received.ID == "" || received.AuditLogID == "" || received.Status != "pending_inspection" {
		t.Fatalf("received = %+v, want pending inspection receipt with audit", received)
	}

	qaInspectReq := smokeRequestAsRole(
		httptest.NewRequest(
			http.MethodPost,
			"/api/v1/returns/"+received.ID+"/inspect",
			bytes.NewBufferString(`{
				"condition": "intact",
				"disposition": "reusable",
				"note": "qa released for putaway",
				"evidence_label": "photo-regression-001"
			}`),
		),
		authConfig,
		auth.RoleQA,
	)
	qaInspectReq.SetPathValue("return_receipt_id", received.ID)
	qaInspectReq.Header.Set(response.HeaderRequestID, "req-regression-return-inspect")
	qaInspectRec := httptest.NewRecorder()

	returnInspectionHandler(inspectService).ServeHTTP(qaInspectRec, qaInspectReq)

	if qaInspectRec.Code != http.StatusOK {
		t.Fatalf("qa inspect status = %d, want %d: %s", qaInspectRec.Code, http.StatusOK, qaInspectRec.Body.String())
	}
	inspection := decodeSmokeSuccess[returnInspectionResponse](t, qaInspectRec).Data
	if inspection.ReceiptID != received.ID || inspection.AuditLogID == "" {
		t.Fatalf("inspection = %+v, want inspected receipt with audit", inspection)
	}

	adminDispositionReq := smokeRequestAsRole(
		httptest.NewRequest(
			http.MethodPost,
			"/api/v1/returns/"+received.ID+"/disposition",
			bytes.NewBufferString(`{"disposition":"reusable","note":"admin approved putaway"}`),
		),
		authConfig,
		auth.RoleERPAdmin,
	)
	adminDispositionReq.SetPathValue("return_receipt_id", received.ID)
	adminDispositionReq.Header.Set(response.HeaderRequestID, "req-regression-return-disposition")
	adminDispositionRec := httptest.NewRecorder()

	returnDispositionHandler(dispositionService).ServeHTTP(adminDispositionRec, adminDispositionReq)

	if adminDispositionRec.Code != http.StatusOK {
		t.Fatalf(
			"admin disposition status = %d, want %d: %s",
			adminDispositionRec.Code,
			http.StatusOK,
			adminDispositionRec.Body.String(),
		)
	}
	disposition := decodeSmokeSuccess[returnDispositionActionResponse](t, adminDispositionRec).Data
	if disposition.ReceiptID != received.ID || disposition.AuditLogID == "" {
		t.Fatalf("disposition = %+v, want disposition with audit", disposition)
	}
	if movementStore.Count() != 1 {
		t.Fatalf("movement count = %d, want 1 return disposition movement", movementStore.Count())
	}

	reconciliationStore := inventoryapp.NewPrototypeEndOfDayReconciliationStore()
	closeService := inventoryapp.NewCloseEndOfDayReconciliation(reconciliationStore, auditStore)
	staffCloseReq := smokeRequestAsRole(
		httptest.NewRequest(
			http.MethodPost,
			"/api/v1/warehouse/end-of-day-reconciliations/rec-hn-260426-day/close",
			bytes.NewBufferString(`{"exception_note":""}`),
		),
		authConfig,
		auth.RoleWarehouseStaff,
	)
	staffCloseReq.SetPathValue("reconciliation_id", "rec-hn-260426-day")
	staffCloseRec := httptest.NewRecorder()

	closeEndOfDayReconciliationHandler(closeService).ServeHTTP(staffCloseRec, staffCloseReq)

	assertForbidden(t, staffCloseRec)
	assertNoAuditAction(t, auditStore, "warehouse.shift.closed")

	adminCloseReq := smokeRequestAsRole(
		httptest.NewRequest(
			http.MethodPost,
			"/api/v1/warehouse/end-of-day-reconciliations/rec-hn-260426-day/close",
			bytes.NewBufferString(`{"exception_note":""}`),
		),
		authConfig,
		auth.RoleERPAdmin,
	)
	adminCloseReq.SetPathValue("reconciliation_id", "rec-hn-260426-day")
	adminCloseReq.Header.Set(response.HeaderRequestID, "req-regression-shift-close")
	adminCloseRec := httptest.NewRecorder()

	closeEndOfDayReconciliationHandler(closeService).ServeHTTP(adminCloseRec, adminCloseReq)

	if adminCloseRec.Code != http.StatusOK {
		t.Fatalf("admin close status = %d, want %d: %s", adminCloseRec.Code, http.StatusOK, adminCloseRec.Body.String())
	}
	closed := decodeSmokeSuccess[endOfDayReconciliationResponse](t, adminCloseRec).Data
	if closed.Status != "closed" || closed.AuditLogID == "" {
		t.Fatalf("closed = %+v, want closed shift with audit", closed)
	}

	assertAuditAction(t, auditStore, "returns.receipt.created")
	assertAuditAction(t, auditStore, "returns.receipt.inspected")
	assertAuditAction(t, auditStore, "returns.inspection.disposition")
	assertAuditAction(t, auditStore, "returns.stock_movement.recorded")
	assertAuditAction(t, auditStore, "warehouse.shift.closed")
}

func TestOrderFulfillmentAuditRegressionSmoke(t *testing.T) {
	harness := newPermissionAuditHarness()
	ctx := context.Background()

	created, err := harness.salesService.CreateSalesOrder(ctx, salesapp.CreateSalesOrderInput{
		ID:           "so-audit-260428-0001",
		OrderNo:      "SO-AUDIT-260428-0001",
		CustomerID:   "cus-dl-minh-anh",
		Channel:      "B2B",
		WarehouseID:  "wh-hcm-fg",
		OrderDate:    "2026-04-28",
		CurrencyCode: "VND",
		Lines: []salesapp.SalesOrderLineInput{
			{ItemID: "item-serum-30ml", OrderedQty: "2", UOMCode: "EA", UnitPrice: "125000"},
		},
		ActorID:   "user-sales-ops",
		RequestID: "req-audit-create",
	})
	if err != nil {
		t.Fatalf("create sales order: %v", err)
	}
	reserved, err := harness.salesService.ConfirmSalesOrder(ctx, salesapp.SalesOrderActionInput{
		ID:              created.SalesOrder.ID,
		ExpectedVersion: created.SalesOrder.Version,
		ActorID:         "user-sales-ops",
		RequestID:       "req-audit-confirm-reserve",
	})
	if err != nil {
		t.Fatalf("confirm sales order: %v", err)
	}
	if reserved.CurrentStatus != salesdomain.SalesOrderStatusReserved {
		t.Fatalf("reserved status = %s, want reserved", reserved.CurrentStatus)
	}

	generatedPick, err := shippingapp.NewGeneratePickTaskFromReservedOrder(harness.pickStore, harness.auditStore).Execute(ctx, shippingapp.GeneratePickTaskFromReservedOrderInput{
		SalesOrder:   reserved.SalesOrder,
		Reservations: harness.reservationStore.Reservations(),
		AssignedTo:   "user-picker",
		ActorID:      "user-warehouse-lead",
		RequestID:    "req-audit-generate-pick",
	})
	if err != nil {
		t.Fatalf("generate pick task: %v", err)
	}
	startedPick, err := shippingapp.NewStartPickTask(harness.pickStore, harness.auditStore).Execute(ctx, shippingapp.PickTaskActionInput{
		PickTaskID: generatedPick.PickTask.ID,
		ActorID:    "user-warehouse-lead",
		RequestID:  "req-audit-start-pick",
	})
	if err != nil {
		t.Fatalf("start pick task: %v", err)
	}
	confirmedPick, err := shippingapp.NewConfirmPickTaskLine(harness.pickStore, harness.auditStore).Execute(ctx, shippingapp.ConfirmPickTaskLineInput{
		PickTaskID: startedPick.PickTask.ID,
		LineID:     startedPick.PickTask.Lines[0].ID,
		PickedQty:  startedPick.PickTask.Lines[0].QtyToPick.String(),
		ActorID:    "user-warehouse-lead",
		RequestID:  "req-audit-confirm-pick",
	})
	if err != nil {
		t.Fatalf("confirm pick line: %v", err)
	}
	completedPick, err := shippingapp.NewCompletePickTask(harness.pickStore, harness.auditStore).Execute(ctx, shippingapp.PickTaskActionInput{
		PickTaskID: confirmedPick.PickTask.ID,
		ActorID:    "user-warehouse-lead",
		RequestID:  "req-audit-complete-pick",
	})
	if err != nil {
		t.Fatalf("complete pick task: %v", err)
	}

	pickedOrder := markSalesOrderPickedForE2E(t, ctx, harness.salesStore, created.SalesOrder.ID)
	generatedPack, err := shippingapp.NewGeneratePackTaskAfterPick(harness.packStore, harness.auditStore).Execute(ctx, shippingapp.GeneratePackTaskAfterPickInput{
		SalesOrder: pickedOrder,
		PickTask:   completedPick.PickTask,
		AssignedTo: "user-packer",
		ActorID:    "user-warehouse-lead",
		RequestID:  "req-audit-generate-pack",
	})
	if err != nil {
		t.Fatalf("generate pack task: %v", err)
	}
	saveSalesOrderSnapshotForE2E(t, ctx, harness.salesStore, generatedPack.SalesOrder)
	startedPack, err := shippingapp.NewStartPackTask(harness.packStore, harness.auditStore).Execute(ctx, shippingapp.PackTaskActionInput{
		PackTaskID: generatedPack.PackTask.ID,
		ActorID:    "user-warehouse-lead",
		RequestID:  "req-audit-start-pack",
	})
	if err != nil {
		t.Fatalf("start pack task: %v", err)
	}
	confirmedPack, err := shippingapp.NewConfirmPackTask(
		harness.packStore,
		harness.auditStore,
		salesOrderPackerAdapter{service: harness.salesService},
	).Execute(ctx, shippingapp.ConfirmPackTaskInput{
		PackTaskID: startedPack.PackTask.ID,
		ActorID:    "user-warehouse-lead",
		RequestID:  "req-audit-confirm-pack",
	})
	if err != nil {
		t.Fatalf("confirm pack task: %v", err)
	}
	if confirmedPack.SalesOrder.Status != salesdomain.SalesOrderStatusPacked {
		t.Fatalf("packed order status = %s, want packed", confirmedPack.SalesOrder.Status)
	}

	if err := harness.manifestStore.SavePackedShipment(shippingdomain.PackedShipment{
		ID:               "ship-audit-260428-0001",
		OrderNo:          confirmedPack.SalesOrder.OrderNo,
		TrackingNo:       "GHNAUDIT260428001",
		CarrierCode:      "GHN",
		CarrierName:      "GHN Express",
		WarehouseID:      confirmedPack.SalesOrder.WarehouseID,
		WarehouseCode:    confirmedPack.SalesOrder.WarehouseCode,
		PackageCode:      "PKG-AUDIT-001",
		StagingZone:      "handover-a",
		HandoverZoneCode: "HANDOVER-A",
		HandoverBinCode:  "TOTE-AUDIT-01",
		Packed:           true,
	}); err != nil {
		t.Fatalf("save packed shipment: %v", err)
	}
	createdManifest, err := shippingapp.NewCreateCarrierManifest(harness.manifestStore, harness.auditStore).Execute(ctx, shippingapp.CreateCarrierManifestInput{
		ID:               "manifest-audit-ghn-260428",
		CarrierCode:      "GHN",
		WarehouseID:      "wh-hcm-fg",
		WarehouseCode:    "WH-HCM-FG",
		Date:             "2026-04-28",
		HandoverBatch:    "day",
		StagingZone:      "handover-a",
		HandoverZoneCode: "HANDOVER-A",
		HandoverBinCode:  "TOTE-AUDIT-01",
		ActorID:          "user-warehouse-lead",
		RequestID:        "req-audit-create-manifest",
	})
	if err != nil {
		t.Fatalf("create manifest: %v", err)
	}
	addedShipment, err := shippingapp.NewAddShipmentToCarrierManifest(harness.manifestStore, harness.auditStore).Execute(ctx, shippingapp.AddShipmentToCarrierManifestInput{
		ManifestID: createdManifest.Manifest.ID,
		ShipmentID: "ship-audit-260428-0001",
		ActorID:    "user-warehouse-lead",
		RequestID:  "req-audit-add-shipment",
	})
	if err != nil {
		t.Fatalf("add shipment: %v", err)
	}
	readyManifest, err := shippingapp.NewMarkCarrierManifestReadyToScan(harness.manifestStore, harness.auditStore).Execute(ctx, shippingapp.CarrierManifestActionInput{
		ManifestID: addedShipment.Manifest.ID,
		ActorID:    "user-warehouse-lead",
		RequestID:  "req-audit-ready-manifest",
	})
	if err != nil {
		t.Fatalf("ready manifest: %v", err)
	}
	scanned, err := shippingapp.NewVerifyCarrierManifestScan(harness.manifestStore, harness.auditStore).Execute(ctx, shippingapp.VerifyCarrierManifestScanInput{
		ManifestID: readyManifest.Manifest.ID,
		Code:       "GHNAUDIT260428001",
		StationID:  "dock-audit",
		DeviceID:   "scanner-audit",
		Source:     "qa_permission_audit",
		ActorID:    "user-warehouse-staff",
		RequestID:  "req-audit-scan",
	})
	if err != nil {
		t.Fatalf("scan manifest: %v", err)
	}
	if scanned.Code != shippingdomain.ScanResultMatched {
		t.Fatalf("scan result = %s, want matched", scanned.Code)
	}
	handedOver, err := shippingapp.NewConfirmCarrierManifestHandover(
		harness.manifestStore,
		harness.auditStore,
		salesOrderHandoverAdapter{service: harness.salesService},
	).Execute(ctx, shippingapp.CarrierManifestActionInput{
		ManifestID: scanned.Manifest.ID,
		ActorID:    "user-warehouse-lead",
		RequestID:  "req-audit-handover",
	})
	if err != nil {
		t.Fatalf("confirm handover: %v", err)
	}
	if handedOver.Manifest.Status != shippingdomain.ManifestStatusHandedOver || len(handedOver.SalesOrders) != 1 {
		t.Fatalf("handover result = %+v, want manifest and one order handed over", handedOver)
	}

	assertAuditAction(t, harness.auditStore, "sales.order.reserved")
	assertAuditAction(t, harness.auditStore, "inventory.stock_reservation.reserved")
	assertAuditAction(t, harness.auditStore, "shipping.pick_task.created")
	assertAuditAction(t, harness.auditStore, "shipping.pick_task.line_confirmed")
	assertAuditAction(t, harness.auditStore, "shipping.pick_task.completed")
	assertAuditAction(t, harness.auditStore, "shipping.pack_task.created")
	assertAuditAction(t, harness.auditStore, "shipping.pack_task.confirmed")
	assertAuditAction(t, harness.auditStore, "sales.order.packed")
	assertAuditAction(t, harness.auditStore, "shipping.manifest.scan_recorded")
	assertAuditAction(t, harness.auditStore, "shipping.manifest.handed_over")
	assertAuditAction(t, harness.auditStore, "sales.order.handed_over")
}

type permissionAuditHarness struct {
	auditStore       *audit.InMemoryLogStore
	salesStore       *salesapp.PrototypeSalesOrderStore
	reservationStore *inventoryapp.PrototypeSalesOrderReservationStore
	salesService     salesapp.SalesOrderService
	pickStore        *shippingapp.PrototypePickTaskStore
	packStore        *shippingapp.PrototypePackTaskStore
	manifestStore    *shippingapp.PrototypeCarrierManifestStore
}

func newPermissionAuditHarness() permissionAuditHarness {
	auditStore := audit.NewInMemoryLogStore()
	salesStore := salesapp.NewPrototypeSalesOrderStore(auditStore)
	reservationStore := inventoryapp.NewPrototypeSalesOrderReservationStore(auditStore)
	salesService := salesapp.NewSalesOrderService(
		salesStore,
		masterdataapp.NewPrototypePartyCatalog(auditStore),
		masterdataapp.NewPrototypeItemCatalog(auditStore),
		masterdataapp.NewPrototypeWarehouseLocationCatalog(auditStore),
	).WithStockReserver(reservationStore)

	return permissionAuditHarness{
		auditStore:       auditStore,
		salesStore:       salesStore,
		reservationStore: reservationStore,
		salesService:     salesService,
		pickStore:        shippingapp.NewPrototypePickTaskStore(),
		packStore:        shippingapp.NewPrototypePackTaskStore(),
		manifestStore:    shippingapp.NewPrototypeCarrierManifestStore(),
	}
}

func assertForbidden(t *testing.T, rec *httptest.ResponseRecorder) {
	t.Helper()

	if rec.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want %d: %s", rec.Code, http.StatusForbidden, rec.Body.String())
	}
	payload := decodeSmokeError(t, rec)
	if payload.Error.Code != response.ErrorCodeForbidden {
		t.Fatalf("code = %s, want %s", payload.Error.Code, response.ErrorCodeForbidden)
	}
}

func assertAuditAction(t *testing.T, store audit.LogStore, action string) {
	t.Helper()

	logs, err := store.List(context.Background(), audit.Query{Action: action})
	if err != nil {
		t.Fatalf("list audit logs for %s: %v", action, err)
	}
	if len(logs) == 0 {
		t.Fatalf("audit action %s missing", action)
	}
}

func assertNoAuditAction(t *testing.T, store audit.LogStore, action string) {
	t.Helper()

	logs, err := store.List(context.Background(), audit.Query{Action: action})
	if err != nil {
		t.Fatalf("list audit logs for %s: %v", action, err)
	}
	if len(logs) != 0 {
		t.Fatalf("audit action %s count = %d, want 0", action, len(logs))
	}
}
