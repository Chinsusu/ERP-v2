package main

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	inventoryapp "github.com/Chinsusu/ERP-v2/apps/api/internal/modules/inventory/application"
	masterdataapp "github.com/Chinsusu/ERP-v2/apps/api/internal/modules/masterdata/application"
	salesapp "github.com/Chinsusu/ERP-v2/apps/api/internal/modules/sales/application"
	salesdomain "github.com/Chinsusu/ERP-v2/apps/api/internal/modules/sales/domain"
	shippingapp "github.com/Chinsusu/ERP-v2/apps/api/internal/modules/shipping/application"
	shippingdomain "github.com/Chinsusu/ERP-v2/apps/api/internal/modules/shipping/domain"
	"github.com/Chinsusu/ERP-v2/apps/api/internal/shared/audit"
	"github.com/Chinsusu/ERP-v2/apps/api/internal/shared/auth"
	"github.com/Chinsusu/ERP-v2/apps/api/internal/shared/response"
)

func TestOrderToHandoverHappyPathSmoke(t *testing.T) {
	authConfig := smokeAuthConfig()
	ctx := context.Background()
	auditStore := audit.NewInMemoryLogStore()
	salesStore := salesapp.NewPrototypeSalesOrderStore(auditStore)
	reservationStore := inventoryapp.NewPrototypeSalesOrderReservationStore(auditStore)
	salesService := salesapp.NewSalesOrderService(
		salesStore,
		masterdataapp.NewPrototypePartyCatalog(auditStore),
		masterdataapp.NewPrototypeItemCatalog(auditStore),
		masterdataapp.NewPrototypeWarehouseLocationCatalog(auditStore),
	).WithStockReserver(reservationStore)
	pickStore := shippingapp.NewPrototypePickTaskStore()
	packStore := shippingapp.NewPrototypePackTaskStore()
	manifestStore := shippingapp.NewPrototypeCarrierManifestStore()

	const (
		orderID    = "so-e2e-260428-0001"
		orderNo    = "SO-E2E-260428-0001"
		manifestID = "manifest-e2e-ghn-260428"
		shipmentID = "ship-e2e-260428-0001"
		trackingNo = "GHNE2E260428001"
	)

	createBody := bytes.NewBufferString(`{
		"id": "so-e2e-260428-0001",
		"order_no": "SO-E2E-260428-0001",
		"customer_id": "cus-dl-minh-anh",
		"channel": "B2B",
		"warehouse_id": "wh-hcm-fg",
		"order_date": "2026-04-28",
		"currency_code": "VND",
		"lines": [
			{
				"item_id": "item-serum-30ml",
				"ordered_qty": "2",
				"uom_code": "EA",
				"unit_price": "125000"
			}
		]
	}`)
	createReq := smokeRequestAsRole(
		httptest.NewRequest(http.MethodPost, "/api/v1/sales-orders", createBody),
		authConfig,
		auth.RoleSalesOps,
	)
	createReq.Header.Set(response.HeaderRequestID, "req-e2e-sales-create")
	createRec := httptest.NewRecorder()

	salesOrdersHandler(salesService).ServeHTTP(createRec, createReq)

	if createRec.Code != http.StatusCreated {
		t.Fatalf("create status = %d, want %d: %s", createRec.Code, http.StatusCreated, createRec.Body.String())
	}
	created := decodeSmokeSuccess[salesOrderResponse](t, createRec).Data
	if created.ID != orderID || created.OrderNo != orderNo || created.Status != "draft" || created.Version != 1 {
		t.Fatalf("created order = %+v, want draft e2e order", created)
	}

	confirmReq := smokeRequestAsRole(
		httptest.NewRequest(http.MethodPost, "/api/v1/sales-orders/"+orderID+"/confirm", bytes.NewBufferString(`{"expected_version":1}`)),
		authConfig,
		auth.RoleSalesOps,
	)
	confirmReq.SetPathValue("sales_order_id", orderID)
	confirmReq.Header.Set(response.HeaderRequestID, "req-e2e-sales-confirm")
	confirmRec := httptest.NewRecorder()

	salesOrderConfirmHandler(salesService).ServeHTTP(confirmRec, confirmReq)

	if confirmRec.Code != http.StatusOK {
		t.Fatalf("confirm status = %d, want %d: %s", confirmRec.Code, http.StatusOK, confirmRec.Body.String())
	}
	reserved := decodeSmokeSuccess[salesOrderActionResultResponse](t, confirmRec).Data
	if reserved.CurrentStatus != string(salesdomain.SalesOrderStatusReserved) ||
		reserved.SalesOrder.Lines[0].ReservedQty != "2.000000" ||
		reserved.SalesOrder.Lines[0].BatchID == "" {
		t.Fatalf("reserved result = %+v, want reserved order with batch allocation", reserved)
	}

	reservedOrder, err := salesService.GetSalesOrder(ctx, orderID)
	if err != nil {
		t.Fatalf("get reserved order: %v", err)
	}
	generatedPick, err := shippingapp.NewGeneratePickTaskFromReservedOrder(pickStore, auditStore).Execute(ctx, shippingapp.GeneratePickTaskFromReservedOrderInput{
		SalesOrder:   reservedOrder,
		Reservations: reservationStore.Reservations(),
		AssignedTo:   "user-picker",
		ActorID:      "user-warehouse-lead",
		RequestID:    "req-e2e-generate-pick",
	})
	if err != nil {
		t.Fatalf("generate pick task: %v", err)
	}
	if generatedPick.PickTask.SalesOrderID != orderID || len(generatedPick.PickTask.Lines) != 1 || generatedPick.AuditLogID == "" {
		t.Fatalf("generated pick = %+v, want one pick task with audit", generatedPick)
	}

	pickTaskID := generatedPick.PickTask.ID
	startPickReq := smokeRequestAsRole(
		httptest.NewRequest(http.MethodPost, "/api/v1/pick-tasks/"+pickTaskID+"/start", nil),
		authConfig,
		auth.RoleWarehouseLead,
	)
	startPickReq.SetPathValue("pick_task_id", pickTaskID)
	startPickReq.Header.Set(response.HeaderRequestID, "req-e2e-pick-start")
	startPickRec := httptest.NewRecorder()

	startPickTaskHandler(shippingapp.NewStartPickTask(pickStore, auditStore)).ServeHTTP(startPickRec, startPickReq)

	if startPickRec.Code != http.StatusOK {
		t.Fatalf("start pick status = %d, want %d: %s", startPickRec.Code, http.StatusOK, startPickRec.Body.String())
	}
	startedPick := decodeSmokeSuccess[pickTaskResponse](t, startPickRec).Data
	if startedPick.Status != string(shippingdomain.PickTaskStatusInProgress) {
		t.Fatalf("started pick = %+v, want in progress", startedPick)
	}

	pickLine := generatedPick.PickTask.Lines[0]
	confirmPickBody := bytes.NewBufferString(fmt.Sprintf(
		`{"line_id":%q,"picked_qty":%q}`,
		pickLine.ID,
		pickLine.QtyToPick.String(),
	))
	confirmPickReq := smokeRequestAsRole(
		httptest.NewRequest(http.MethodPost, "/api/v1/pick-tasks/"+pickTaskID+"/confirm-line", confirmPickBody),
		authConfig,
		auth.RoleWarehouseLead,
	)
	confirmPickReq.SetPathValue("pick_task_id", pickTaskID)
	confirmPickReq.Header.Set(response.HeaderRequestID, "req-e2e-pick-confirm")
	confirmPickRec := httptest.NewRecorder()

	confirmPickTaskLineHandler(shippingapp.NewConfirmPickTaskLine(pickStore, auditStore)).ServeHTTP(confirmPickRec, confirmPickReq)

	if confirmPickRec.Code != http.StatusOK {
		t.Fatalf("confirm pick status = %d, want %d: %s", confirmPickRec.Code, http.StatusOK, confirmPickRec.Body.String())
	}
	confirmedPick := decodeSmokeSuccess[pickTaskResponse](t, confirmPickRec).Data
	if confirmedPick.Lines[0].Status != string(shippingdomain.PickTaskLineStatusPicked) ||
		confirmedPick.Lines[0].QtyPicked != "2.000000" {
		t.Fatalf("confirmed pick = %+v, want picked line", confirmedPick)
	}

	completePickReq := smokeRequestAsRole(
		httptest.NewRequest(http.MethodPost, "/api/v1/pick-tasks/"+pickTaskID+"/complete", nil),
		authConfig,
		auth.RoleWarehouseLead,
	)
	completePickReq.SetPathValue("pick_task_id", pickTaskID)
	completePickReq.Header.Set(response.HeaderRequestID, "req-e2e-pick-complete")
	completePickRec := httptest.NewRecorder()

	completePickTaskHandler(shippingapp.NewCompletePickTask(pickStore, auditStore)).ServeHTTP(completePickRec, completePickReq)

	if completePickRec.Code != http.StatusOK {
		t.Fatalf("complete pick status = %d, want %d: %s", completePickRec.Code, http.StatusOK, completePickRec.Body.String())
	}
	completedPick := decodeSmokeSuccess[pickTaskResponse](t, completePickRec).Data
	if completedPick.Status != string(shippingdomain.PickTaskStatusCompleted) {
		t.Fatalf("completed pick = %+v, want completed", completedPick)
	}

	pickedOrder := markSalesOrderPickedForE2E(t, ctx, salesStore, orderID)
	completedPickTask, err := pickStore.GetPickTask(ctx, pickTaskID)
	if err != nil {
		t.Fatalf("get completed pick task: %v", err)
	}
	generatedPack, err := shippingapp.NewGeneratePackTaskAfterPick(packStore, auditStore).Execute(ctx, shippingapp.GeneratePackTaskAfterPickInput{
		SalesOrder: pickedOrder,
		PickTask:   completedPickTask,
		AssignedTo: "user-packer",
		ActorID:    "user-warehouse-lead",
		RequestID:  "req-e2e-generate-pack",
	})
	if err != nil {
		t.Fatalf("generate pack task: %v", err)
	}
	if generatedPack.PackTask.SalesOrderID != orderID ||
		generatedPack.SalesOrder.Status != salesdomain.SalesOrderStatusPacking ||
		generatedPack.AuditLogID == "" {
		t.Fatalf("generated pack = %+v, want pack task and packing order", generatedPack)
	}
	saveSalesOrderSnapshotForE2E(t, ctx, salesStore, generatedPack.SalesOrder)

	packTaskID := generatedPack.PackTask.ID
	startPackReq := smokeRequestAsRole(
		httptest.NewRequest(http.MethodPost, "/api/v1/pack-tasks/"+packTaskID+"/start", nil),
		authConfig,
		auth.RoleWarehouseLead,
	)
	startPackReq.SetPathValue("pack_task_id", packTaskID)
	startPackReq.Header.Set(response.HeaderRequestID, "req-e2e-pack-start")
	startPackRec := httptest.NewRecorder()

	startPackTaskHandler(shippingapp.NewStartPackTask(packStore, auditStore)).ServeHTTP(startPackRec, startPackReq)

	if startPackRec.Code != http.StatusOK {
		t.Fatalf("start pack status = %d, want %d: %s", startPackRec.Code, http.StatusOK, startPackRec.Body.String())
	}
	startedPack := decodeSmokeSuccess[packTaskResponse](t, startPackRec).Data
	if startedPack.Status != string(shippingdomain.PackTaskStatusInProgress) {
		t.Fatalf("started pack = %+v, want in progress", startedPack)
	}

	packLine := generatedPack.PackTask.Lines[0]
	confirmPackBody := bytes.NewBufferString(fmt.Sprintf(
		`{"lines":[{"line_id":%q,"packed_qty":%q}]}`,
		packLine.ID,
		packLine.QtyToPack.String(),
	))
	confirmPackReq := smokeRequestAsRole(
		httptest.NewRequest(http.MethodPost, "/api/v1/pack-tasks/"+packTaskID+"/confirm", confirmPackBody),
		authConfig,
		auth.RoleWarehouseLead,
	)
	confirmPackReq.SetPathValue("pack_task_id", packTaskID)
	confirmPackReq.Header.Set(response.HeaderRequestID, "req-e2e-pack-confirm")
	confirmPackRec := httptest.NewRecorder()

	confirmPackTaskHandler(
		shippingapp.NewConfirmPackTask(packStore, auditStore, salesOrderPackerAdapter{service: salesService}),
	).ServeHTTP(confirmPackRec, confirmPackReq)

	if confirmPackRec.Code != http.StatusOK {
		t.Fatalf("confirm pack status = %d, want %d: %s", confirmPackRec.Code, http.StatusOK, confirmPackRec.Body.String())
	}
	confirmedPack := decodeSmokeSuccess[packTaskResponse](t, confirmPackRec).Data
	if confirmedPack.Status != string(shippingdomain.PackTaskStatusPacked) ||
		confirmedPack.Lines[0].Status != string(shippingdomain.PackTaskLineStatusPacked) ||
		confirmedPack.SalesOrderStatus != string(salesdomain.SalesOrderStatusPacked) {
		t.Fatalf("confirmed pack = %+v, want packed task and order", confirmedPack)
	}

	packedOrder, err := salesService.GetSalesOrder(ctx, orderID)
	if err != nil {
		t.Fatalf("get packed order: %v", err)
	}
	if packedOrder.Status != salesdomain.SalesOrderStatusPacked {
		t.Fatalf("packed order status = %s, want packed", packedOrder.Status)
	}
	if err := manifestStore.SavePackedShipment(shippingdomain.PackedShipment{
		ID:               shipmentID,
		OrderNo:          packedOrder.OrderNo,
		TrackingNo:       trackingNo,
		CarrierCode:      "GHN",
		CarrierName:      "GHN Express",
		WarehouseID:      packedOrder.WarehouseID,
		WarehouseCode:    packedOrder.WarehouseCode,
		PackageCode:      "PKG-E2E-001",
		StagingZone:      "handover-a",
		HandoverZoneCode: "HANDOVER-A",
		HandoverBinCode:  "TOTE-E2E-01",
		Packed:           true,
	}); err != nil {
		t.Fatalf("save packed shipment: %v", err)
	}

	createManifestReq := smokeRequestAsRole(
		httptest.NewRequest(http.MethodPost, "/api/v1/shipping/manifests", bytes.NewBufferString(`{
			"id": "manifest-e2e-ghn-260428",
			"carrier_code": "GHN",
			"warehouse_id": "wh-hcm-fg",
			"warehouse_code": "WH-HCM-FG",
			"date": "2026-04-28",
			"handover_batch": "day",
			"staging_zone": "handover-a",
			"handover_zone_code": "HANDOVER-A",
			"handover_bin_code": "TOTE-E2E-01",
			"owner": "Warehouse Lead"
		}`)),
		authConfig,
		auth.RoleWarehouseLead,
	)
	createManifestReq.Header.Set(response.HeaderRequestID, "req-e2e-manifest-create")
	createManifestRec := httptest.NewRecorder()

	carrierManifestsHandler(
		shippingapp.NewListCarrierManifests(manifestStore),
		shippingapp.NewCreateCarrierManifest(manifestStore, auditStore),
	).ServeHTTP(createManifestRec, createManifestReq)

	if createManifestRec.Code != http.StatusCreated {
		t.Fatalf("create manifest status = %d, want %d: %s", createManifestRec.Code, http.StatusCreated, createManifestRec.Body.String())
	}
	createdManifest := decodeSmokeSuccess[carrierManifestResponse](t, createManifestRec).Data
	if createdManifest.ID != manifestID || createdManifest.Status != string(shippingdomain.ManifestStatusDraft) {
		t.Fatalf("created manifest = %+v, want draft manifest", createdManifest)
	}

	addShipmentReq := smokeRequestAsRole(
		httptest.NewRequest(http.MethodPost, "/api/v1/shipping/manifests/"+manifestID+"/shipments", bytes.NewBufferString(`{"shipment_id":"ship-e2e-260428-0001"}`)),
		authConfig,
		auth.RoleWarehouseLead,
	)
	addShipmentReq.SetPathValue("manifest_id", manifestID)
	addShipmentReq.Header.Set(response.HeaderRequestID, "req-e2e-manifest-add-shipment")
	addShipmentRec := httptest.NewRecorder()

	addShipmentToCarrierManifestHandler(shippingapp.NewAddShipmentToCarrierManifest(manifestStore, auditStore)).ServeHTTP(addShipmentRec, addShipmentReq)

	if addShipmentRec.Code != http.StatusOK {
		t.Fatalf("add shipment status = %d, want %d: %s", addShipmentRec.Code, http.StatusOK, addShipmentRec.Body.String())
	}
	addedShipment := decodeSmokeSuccess[carrierManifestResponse](t, addShipmentRec).Data
	if addedShipment.Summary.ExpectedCount != 1 || addedShipment.Summary.MissingCount != 1 {
		t.Fatalf("manifest after add = %+v, want one missing shipment", addedShipment)
	}

	readyReq := smokeRequestAsRole(
		httptest.NewRequest(http.MethodPost, "/api/v1/shipping/manifests/"+manifestID+"/ready-to-scan", nil),
		authConfig,
		auth.RoleWarehouseLead,
	)
	readyReq.SetPathValue("manifest_id", manifestID)
	readyReq.Header.Set(response.HeaderRequestID, "req-e2e-manifest-ready")
	readyRec := httptest.NewRecorder()

	markCarrierManifestReadyToScanHandler(shippingapp.NewMarkCarrierManifestReadyToScan(manifestStore, auditStore)).ServeHTTP(readyRec, readyReq)

	if readyRec.Code != http.StatusOK {
		t.Fatalf("ready manifest status = %d, want %d: %s", readyRec.Code, http.StatusOK, readyRec.Body.String())
	}
	readyManifest := decodeSmokeSuccess[carrierManifestResponse](t, readyRec).Data
	if readyManifest.Status != string(shippingdomain.ManifestStatusReady) {
		t.Fatalf("ready manifest = %+v, want ready", readyManifest)
	}

	scanReq := smokeRequestAsRole(
		httptest.NewRequest(http.MethodPost, "/api/v1/shipping/manifests/"+manifestID+"/scan", bytes.NewBufferString(`{
			"code": "GHNE2E260428001",
			"station_id": "dock-e2e",
			"device_id": "scanner-e2e",
			"source": "qa_e2e_happy_path"
		}`)),
		authConfig,
		auth.RoleWarehouseStaff,
	)
	scanReq.SetPathValue("manifest_id", manifestID)
	scanReq.Header.Set(response.HeaderRequestID, "req-e2e-manifest-scan")
	scanRec := httptest.NewRecorder()

	verifyCarrierManifestScanHandler(shippingapp.NewVerifyCarrierManifestScan(manifestStore, auditStore)).ServeHTTP(scanRec, scanReq)

	if scanRec.Code != http.StatusOK {
		t.Fatalf("scan status = %d, want %d: %s", scanRec.Code, http.StatusOK, scanRec.Body.String())
	}
	scanned := decodeSmokeSuccess[carrierManifestScanResponse](t, scanRec).Data
	if scanned.ResultCode != string(shippingdomain.ScanResultMatched) ||
		scanned.Manifest.Summary.ScannedCount != 1 ||
		scanned.Manifest.Summary.MissingCount != 0 ||
		scanned.ScanEvent.DeviceID != "scanner-e2e" {
		t.Fatalf("scan payload = %+v, want matched full manifest scan", scanned)
	}

	handoverReq := smokeRequestAsRole(
		httptest.NewRequest(http.MethodPost, "/api/v1/shipping/manifests/"+manifestID+"/confirm-handover", nil),
		authConfig,
		auth.RoleWarehouseLead,
	)
	handoverReq.SetPathValue("manifest_id", manifestID)
	handoverReq.Header.Set(response.HeaderRequestID, "req-e2e-manifest-handover")
	handoverRec := httptest.NewRecorder()

	confirmCarrierManifestHandoverHandler(
		shippingapp.NewConfirmCarrierManifestHandover(manifestStore, auditStore, salesOrderHandoverAdapter{service: salesService}),
	).ServeHTTP(handoverRec, handoverReq)

	if handoverRec.Code != http.StatusOK {
		t.Fatalf("handover status = %d, want %d: %s", handoverRec.Code, http.StatusOK, handoverRec.Body.String())
	}
	handedOver := decodeSmokeSuccess[carrierManifestResponse](t, handoverRec).Data
	if handedOver.Status != string(shippingdomain.ManifestStatusHandedOver) ||
		handedOver.Summary.ExpectedCount != 1 ||
		handedOver.Summary.MissingCount != 0 {
		t.Fatalf("handover manifest = %+v, want handed over without missing lines", handedOver)
	}

	finalOrder, err := salesService.GetSalesOrder(ctx, orderID)
	if err != nil {
		t.Fatalf("get final order: %v", err)
	}
	if finalOrder.Status != salesdomain.SalesOrderStatusHandedOver {
		t.Fatalf("final order status = %s, want handed_over", finalOrder.Status)
	}
	scanEvents, err := manifestStore.ListScanEvents(ctx, manifestID)
	if err != nil {
		t.Fatalf("list scan events: %v", err)
	}
	if len(scanEvents) != 1 || scanEvents[0].TrackingNo != trackingNo || scanEvents[0].ResultCode != shippingdomain.ScanResultMatched {
		t.Fatalf("scan events = %+v, want one matched e2e scan event", scanEvents)
	}
	handoverLogs, err := auditStore.List(ctx, audit.Query{Action: "shipping.manifest.handed_over"})
	if err != nil {
		t.Fatalf("list handover audit logs: %v", err)
	}
	if len(handoverLogs) != 1 || handoverLogs[0].AfterData["status"] != string(shippingdomain.ManifestStatusHandedOver) {
		t.Fatalf("handover audit logs = %+v, want one handed-over manifest log", handoverLogs)
	}
}

func markSalesOrderPickedForE2E(
	t *testing.T,
	ctx context.Context,
	store *salesapp.PrototypeSalesOrderStore,
	orderID string,
) salesdomain.SalesOrder {
	t.Helper()

	var picked salesdomain.SalesOrder
	err := store.WithinTx(ctx, func(txCtx context.Context, tx salesapp.SalesOrderTx) error {
		current, err := tx.GetForUpdate(txCtx, orderID)
		if err != nil {
			return err
		}
		pickStartedAt := time.Date(2026, 4, 28, 10, 30, 0, 0, time.UTC)
		picking, err := current.StartPicking("user-picker", pickStartedAt)
		if err != nil {
			return err
		}
		picked, err = picking.MarkPicked("user-picker", pickStartedAt.Add(15*time.Minute))
		if err != nil {
			return err
		}

		return tx.Save(txCtx, picked)
	})
	if err != nil {
		t.Fatalf("mark sales order picked: %v", err)
	}

	return picked
}

func saveSalesOrderSnapshotForE2E(
	t *testing.T,
	ctx context.Context,
	store *salesapp.PrototypeSalesOrderStore,
	order salesdomain.SalesOrder,
) {
	t.Helper()

	err := store.WithinTx(ctx, func(txCtx context.Context, tx salesapp.SalesOrderTx) error {
		return tx.Save(txCtx, order)
	})
	if err != nil {
		t.Fatalf("save sales order snapshot: %v", err)
	}
}
