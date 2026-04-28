package application

import (
	"context"
	"errors"
	"strings"
	"testing"

	salesdomain "github.com/Chinsusu/ERP-v2/apps/api/internal/modules/sales/domain"
	"github.com/Chinsusu/ERP-v2/apps/api/internal/modules/shipping/domain"
	"github.com/Chinsusu/ERP-v2/apps/api/internal/shared/audit"
)

func TestListCarrierManifestsFiltersByWarehouseDateCarrierAndStatus(t *testing.T) {
	service := NewListCarrierManifests(NewPrototypeCarrierManifestStore())
	rows, err := service.Execute(context.Background(), domain.NewCarrierManifestFilter(
		"wh-hcm",
		"2026-04-26",
		"ghn",
		domain.ManifestStatusScanning,
	))
	if err != nil {
		t.Fatalf("list manifests: %v", err)
	}

	if len(rows) != 1 {
		t.Fatalf("rows = %d, want 1", len(rows))
	}
	if rows[0].ID != "manifest-hcm-ghn-morning" {
		t.Fatalf("manifest id = %q, want manifest-hcm-ghn-morning", rows[0].ID)
	}
	if got := rows[0].Summary(); got.ExpectedCount != 3 || got.ScannedCount != 2 || got.MissingCount != 1 {
		t.Fatalf("summary = %+v, want 3 expected, 2 scanned, 1 missing", got)
	}
}

func TestCreateCarrierManifestWritesAudit(t *testing.T) {
	store := NewPrototypeCarrierManifestStore()
	auditStore := audit.NewInMemoryLogStore()
	service := NewCreateCarrierManifest(store, auditStore)

	result, err := service.Execute(context.Background(), CreateCarrierManifestInput{
		CarrierCode:   "NJV",
		WarehouseID:   "wh-hcm",
		WarehouseCode: "HCM",
		Date:          "2026-04-26",
		ActorID:       "user-warehouse-lead",
		RequestID:     "req-create-manifest",
	})
	if err != nil {
		t.Fatalf("create manifest: %v", err)
	}
	if result.Manifest.Status != domain.ManifestStatusDraft {
		t.Fatalf("status = %q, want draft", result.Manifest.Status)
	}
	if result.Manifest.CarrierName != "Ninja Van" ||
		result.Manifest.StagingZone != "handover-c" ||
		result.Manifest.HandoverZoneCode != "HANDOVER-C" {
		t.Fatalf("manifest = %+v, want carrier name and handover zone from carrier master", result.Manifest)
	}
	if result.AuditLogID == "" {
		t.Fatal("audit log id is empty")
	}

	logs, err := auditStore.List(context.Background(), audit.Query{Action: "shipping.manifest.created"})
	if err != nil {
		t.Fatalf("list audit logs: %v", err)
	}
	if len(logs) != 1 || logs[0].Metadata["carrier_sla_profile"] != "standard" {
		t.Fatalf("audit logs = %+v, want one log with carrier SLA profile", logs)
	}
}

func TestCreateCarrierManifestRejectsInactiveCarrier(t *testing.T) {
	service := NewCreateCarrierManifest(NewPrototypeCarrierManifestStore(), audit.NewInMemoryLogStore())

	_, err := service.Execute(context.Background(), CreateCarrierManifestInput{
		CarrierCode:   "GHTK",
		WarehouseID:   "wh-hcm",
		WarehouseCode: "HCM",
		Date:          "2026-04-26",
		ActorID:       "user-warehouse-lead",
		RequestID:     "req-create-manifest-inactive",
	})
	if !errors.Is(err, ErrCarrierInactive) {
		t.Fatalf("err = %v, want inactive carrier", err)
	}
}

func TestCreateCarrierManifestRejectsUnknownCarrier(t *testing.T) {
	service := NewCreateCarrierManifest(NewPrototypeCarrierManifestStore(), audit.NewInMemoryLogStore())

	_, err := service.Execute(context.Background(), CreateCarrierManifestInput{
		CarrierCode:   "UNKNOWN",
		WarehouseID:   "wh-hcm",
		WarehouseCode: "HCM",
		Date:          "2026-04-26",
		ActorID:       "user-warehouse-lead",
		RequestID:     "req-create-manifest-unknown",
	})
	if !errors.Is(err, ErrCarrierNotFound) {
		t.Fatalf("err = %v, want carrier not found", err)
	}
}

func TestAddShipmentToCarrierManifestUpdatesCountsAndAudit(t *testing.T) {
	store := NewPrototypeCarrierManifestStore()
	auditStore := audit.NewInMemoryLogStore()
	manifest := draftCarrierManifestForActionTest(t)
	if err := store.Save(context.Background(), manifest); err != nil {
		t.Fatalf("save manifest: %v", err)
	}
	service := NewAddShipmentToCarrierManifest(store, auditStore)

	result, err := service.Execute(context.Background(), AddShipmentToCarrierManifestInput{
		ManifestID: manifest.ID,
		ShipmentID: "ship-hcm-260426-004",
		ActorID:    "user-warehouse-lead",
		RequestID:  "req-add-shipment",
	})
	if err != nil {
		t.Fatalf("add shipment: %v", err)
	}
	if got := result.Manifest.Summary(); got.ExpectedCount != 1 || got.ScannedCount != 0 || got.MissingCount != 1 {
		t.Fatalf("summary = %+v, want 1 expected, 0 scanned, 1 missing", got)
	}
	if result.Manifest.Lines[0].HandoverZoneCode != "HANDOVER-A" ||
		result.Manifest.Lines[0].HandoverBinCode != "TOTE-A03" {
		t.Fatalf("line = %+v, want shipment handover zone/bin copied", result.Manifest.Lines[0])
	}

	logs, err := auditStore.List(context.Background(), audit.Query{Action: "shipping.manifest.shipment_added"})
	if err != nil {
		t.Fatalf("list audit logs: %v", err)
	}
	if len(logs) != 1 {
		t.Fatalf("audit logs = %d, want 1", len(logs))
	}
}

func TestAddShipmentToCarrierManifestRejectsUnpackedAndWrongCarrier(t *testing.T) {
	ctx := context.Background()
	store := NewPrototypeCarrierManifestStore()
	auditStore := audit.NewInMemoryLogStore()
	manifest := draftCarrierManifestForActionTest(t)
	if err := store.Save(ctx, manifest); err != nil {
		t.Fatalf("save manifest: %v", err)
	}
	service := NewAddShipmentToCarrierManifest(store, auditStore)

	cases := []struct {
		name       string
		shipmentID string
		want       error
	}{
		{name: "unpacked", shipmentID: "ship-hcm-260426-099", want: domain.ErrManifestShipmentNotPacked},
		{name: "wrong carrier", shipmentID: "ship-hcm-vtp-260426-001", want: domain.ErrManifestCarrierMismatch},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := service.Execute(ctx, AddShipmentToCarrierManifestInput{
				ManifestID: manifest.ID,
				ShipmentID: tc.shipmentID,
				ActorID:    "user-warehouse-lead",
				RequestID:  "req-add-shipment-reject",
			})
			if !errors.Is(err, tc.want) {
				t.Fatalf("err = %v, want %v", err, tc.want)
			}
		})
	}
}

func TestCarrierManifestReadyRemoveAndCancelActionsWriteAudit(t *testing.T) {
	ctx := context.Background()
	store := NewPrototypeCarrierManifestStore()
	auditStore := audit.NewInMemoryLogStore()
	manifest := draftCarrierManifestForActionTest(t)
	if err := store.Save(ctx, manifest); err != nil {
		t.Fatalf("save manifest: %v", err)
	}
	added, err := NewAddShipmentToCarrierManifest(store, auditStore).Execute(ctx, AddShipmentToCarrierManifestInput{
		ManifestID: manifest.ID,
		ShipmentID: "ship-hcm-260426-004",
		ActorID:    "user-warehouse-lead",
		RequestID:  "req-add-shipment-for-action",
	})
	if err != nil {
		t.Fatalf("add shipment: %v", err)
	}
	if added.Manifest.Status != domain.ManifestStatusDraft {
		t.Fatalf("added status = %q, want draft before ready action", added.Manifest.Status)
	}

	ready, err := NewMarkCarrierManifestReadyToScan(store, auditStore).Execute(ctx, CarrierManifestActionInput{
		ManifestID: manifest.ID,
		ActorID:    "user-warehouse-lead",
		RequestID:  "req-ready-manifest",
	})
	if err != nil {
		t.Fatalf("ready manifest: %v", err)
	}
	if ready.Manifest.Status != domain.ManifestStatusReady || ready.AuditLogID == "" {
		t.Fatalf("ready = %+v, want ready manifest with audit", ready)
	}

	removed, err := NewRemoveShipmentFromCarrierManifest(store, auditStore).Execute(ctx, RemoveShipmentFromCarrierManifestInput{
		ManifestID: manifest.ID,
		ShipmentID: "ship-hcm-260426-004",
		ActorID:    "user-warehouse-lead",
		RequestID:  "req-remove-shipment",
	})
	if err != nil {
		t.Fatalf("remove shipment: %v", err)
	}
	if removed.Manifest.Status != domain.ManifestStatusDraft || len(removed.Manifest.Lines) != 0 || removed.AuditLogID == "" {
		t.Fatalf("removed = %+v, want empty draft manifest with audit", removed)
	}

	cancelled, err := NewCancelCarrierManifest(store, auditStore).Execute(ctx, CarrierManifestActionInput{
		ManifestID: manifest.ID,
		ActorID:    "user-warehouse-lead",
		RequestID:  "req-cancel-manifest",
		Reason:     "carrier pickup moved",
	})
	if err != nil {
		t.Fatalf("cancel manifest: %v", err)
	}
	if cancelled.Manifest.Status != domain.ManifestStatusCancelled || cancelled.AuditLogID == "" {
		t.Fatalf("cancelled = %+v, want cancelled manifest with audit", cancelled)
	}

	for _, action := range []string{
		"shipping.manifest.ready_to_scan",
		"shipping.manifest.shipment_removed",
		"shipping.manifest.cancelled",
	} {
		logs, err := auditStore.List(ctx, audit.Query{Action: action})
		if err != nil {
			t.Fatalf("list audit logs for %s: %v", action, err)
		}
		if len(logs) != 1 {
			t.Fatalf("%s logs = %d, want 1", action, len(logs))
		}
	}
}

func TestReportCarrierManifestMissingOrdersMarksExceptionAndAuditsMissingLines(t *testing.T) {
	ctx := context.Background()
	store := NewPrototypeCarrierManifestStore()
	auditStore := audit.NewInMemoryLogStore()
	service := NewReportCarrierManifestMissingOrders(store, auditStore)

	result, err := service.Execute(ctx, CarrierManifestActionInput{
		ManifestID: "manifest-hcm-ghn-morning",
		ActorID:    "user-handover-operator",
		RequestID:  "req-missing-manifest",
		Reason:     "physical tote missing one order",
	})
	if err != nil {
		t.Fatalf("report missing orders: %v", err)
	}
	if result.Manifest.Status != domain.ManifestStatusException || result.AuditLogID == "" {
		t.Fatalf("result = %+v, want exception manifest with audit", result)
	}
	missingLines := result.Manifest.MissingLines()
	if len(missingLines) != 1 || missingLines[0].OrderNo != "SO-260426-003" || missingLines[0].TrackingNo != "GHN260426003" {
		t.Fatalf("missing lines = %+v, want SO-260426-003/GHN260426003", missingLines)
	}

	stored, err := store.Get(ctx, "manifest-hcm-ghn-morning")
	if err != nil {
		t.Fatalf("get stored manifest: %v", err)
	}
	if stored.Status != domain.ManifestStatusException {
		t.Fatalf("stored status = %q, want exception", stored.Status)
	}

	logs, err := auditStore.List(ctx, audit.Query{Action: "shipping.manifest.missing_exception_reported"})
	if err != nil {
		t.Fatalf("list audit logs: %v", err)
	}
	if len(logs) != 1 || logs[0].Metadata["missing_count"] != 1 {
		t.Fatalf("audit logs = %+v, want missing exception count", logs)
	}
}

func TestReportCarrierManifestMissingOrdersRejectsCleanManifest(t *testing.T) {
	ctx := context.Background()
	store := NewPrototypeCarrierManifestStore()
	auditStore := audit.NewInMemoryLogStore()
	if _, err := NewVerifyCarrierManifestScan(store, auditStore).Execute(ctx, VerifyCarrierManifestScanInput{
		ManifestID: "manifest-hcm-ghn-morning",
		Code:       "GHN260426003",
		ActorID:    "user-handover-operator",
		RequestID:  "req-clean-manifest-scan",
	}); err != nil {
		t.Fatalf("scan missing line: %v", err)
	}

	_, err := NewReportCarrierManifestMissingOrders(store, auditStore).Execute(ctx, CarrierManifestActionInput{
		ManifestID: "manifest-hcm-ghn-morning",
		ActorID:    "user-handover-operator",
		RequestID:  "req-clean-missing-manifest",
	})
	if !errors.Is(err, domain.ErrManifestNoMissingOrders) {
		t.Fatalf("err = %v, want no missing orders", err)
	}
}

func TestConfirmCarrierManifestHandoverRequiresAllScansAndMarksOrders(t *testing.T) {
	ctx := context.Background()
	store := NewPrototypeCarrierManifestStore()
	auditStore := audit.NewInMemoryLogStore()
	handover := &recordingSalesOrderHandover{}
	service := NewConfirmCarrierManifestHandover(store, auditStore, handover)

	_, err := service.Execute(ctx, CarrierManifestActionInput{
		ManifestID: "manifest-hcm-ghn-morning",
		ActorID:    "user-handover-operator",
		RequestID:  "req-confirm-missing",
	})
	if !errors.Is(err, domain.ErrManifestMissingOrders) {
		t.Fatalf("err = %v, want missing orders", err)
	}
	if len(handover.orderNos) != 0 {
		t.Fatalf("handover calls = %+v, want none when missing", handover.orderNos)
	}

	if _, err := NewVerifyCarrierManifestScan(store, auditStore).Execute(ctx, VerifyCarrierManifestScanInput{
		ManifestID: "manifest-hcm-ghn-morning",
		Code:       "GHN260426003",
		ActorID:    "user-handover-operator",
		RequestID:  "req-scan-before-confirm",
	}); err != nil {
		t.Fatalf("scan missing line: %v", err)
	}
	result, err := service.Execute(ctx, CarrierManifestActionInput{
		ManifestID: "manifest-hcm-ghn-morning",
		ActorID:    "user-handover-operator",
		RequestID:  "req-confirm-handover",
	})
	if err != nil {
		t.Fatalf("confirm handover: %v", err)
	}
	if result.Manifest.Status != domain.ManifestStatusHandedOver || result.AuditLogID == "" {
		t.Fatalf("result = %+v, want handed_over manifest with audit", result)
	}
	if len(handover.orderNos) != 3 || handover.orderNos[2] != "SO-260426-003" {
		t.Fatalf("handover order calls = %+v, want all manifest orders", handover.orderNos)
	}
	if len(result.SalesOrders) != 3 || result.SalesOrders[0].Status != salesdomain.SalesOrderStatusHandedOver {
		t.Fatalf("sales orders = %+v, want handed over orders", result.SalesOrders)
	}

	logs, err := auditStore.List(ctx, audit.Query{Action: "shipping.manifest.handed_over"})
	if err != nil {
		t.Fatalf("list audit logs: %v", err)
	}
	if len(logs) != 1 || logs[0].Metadata["handed_over_order_count"] != 3 {
		t.Fatalf("audit logs = %+v, want handed over count", logs)
	}
}

func TestVerifyCarrierManifestScanMarksExpectedLineAndRecordsEvent(t *testing.T) {
	store := NewPrototypeCarrierManifestStore()
	auditStore := audit.NewInMemoryLogStore()
	service := NewVerifyCarrierManifestScan(store, auditStore)

	result, err := service.Execute(context.Background(), VerifyCarrierManifestScanInput{
		ManifestID: "manifest-hcm-ghn-morning",
		Code:       "GHN260426003",
		StationID:  "dock-a",
		DeviceID:   "scanner-01",
		Source:     "handheld_scanner",
		ActorID:    "user-handover-operator",
		RequestID:  "req-scan-match",
	})
	if err != nil {
		t.Fatalf("verify scan: %v", err)
	}
	if result.Code != domain.ScanResultMatched {
		t.Fatalf("result code = %q, want matched", result.Code)
	}
	if got := result.Manifest.Summary(); got.ExpectedCount != 3 || got.ScannedCount != 3 || got.MissingCount != 0 {
		t.Fatalf("summary = %+v, want 3 expected, 3 scanned, 0 missing", got)
	}

	events, err := store.ListScanEvents(context.Background(), "manifest-hcm-ghn-morning")
	if err != nil {
		t.Fatalf("list events: %v", err)
	}
	if len(events) != 1 || events[0].ResultCode != domain.ScanResultMatched {
		t.Fatalf("scan events = %+v, want one matched event", events)
	}
	event := events[0]
	if event.ManifestID != "manifest-hcm-ghn-morning" || event.OrderNo != "SO-260426-003" || event.TrackingNo != "GHN260426003" {
		t.Fatalf("scan event = %+v, want manifest/order/tracking retained", event)
	}
	if event.ActorID != "user-handover-operator" || event.DeviceID != "scanner-01" || event.Source != "handheld_scanner" {
		t.Fatalf("scan event = %+v, want actor/device/source retained", event)
	}

	logs, err := auditStore.List(context.Background(), audit.Query{Action: "shipping.manifest.scan_recorded"})
	if err != nil {
		t.Fatalf("list audit logs: %v", err)
	}
	if len(logs) != 1 {
		t.Fatalf("audit logs = %d, want 1", len(logs))
	}
	if logs[0].AfterData["device_id"] != "scanner-01" || logs[0].AfterData["source"] != "handheld_scanner" {
		t.Fatalf("audit after data = %+v, want device/source retained", logs[0].AfterData)
	}
	if logs[0].Metadata["scan_source"] != "handheld_scanner" {
		t.Fatalf("audit metadata = %+v, want scan source retained", logs[0].Metadata)
	}
}

type recordingSalesOrderHandover struct {
	orderNos []string
}

func (r *recordingSalesOrderHandover) MarkSalesOrderHandedOver(
	_ context.Context,
	input CarrierManifestSalesOrderHandoverInput,
) (salesdomain.SalesOrder, error) {
	r.orderNos = append(r.orderNos, input.OrderNo)
	return salesdomain.SalesOrder{
		OrderNo: input.OrderNo,
		Status:  salesdomain.SalesOrderStatusHandedOver,
	}, nil
}

func TestVerifyCarrierManifestScanReturnsClearWarningCodes(t *testing.T) {
	store := NewPrototypeCarrierManifestStore()
	service := NewVerifyCarrierManifestScan(store, audit.NewInMemoryLogStore())

	cases := []struct {
		name string
		code string
		want domain.CarrierManifestScanResultCode
	}{
		{name: "duplicate", code: "GHN260426001", want: domain.ScanResultDuplicate},
		{name: "wrong manifest", code: "VTP260426011", want: domain.ScanResultManifestMismatch},
		{name: "unpacked", code: "GHN260426099", want: domain.ScanResultInvalidState},
		{name: "unknown", code: "UNKNOWN-CODE", want: domain.ScanResultNotFound},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := service.Execute(context.Background(), VerifyCarrierManifestScanInput{
				ManifestID: "manifest-hcm-ghn-morning",
				Code:       tc.code,
				StationID:  "dock-a",
				ActorID:    "user-handover-operator",
				RequestID:  "req-scan-warning",
			})
			if err != nil {
				t.Fatalf("verify scan: %v", err)
			}
			if result.Code != tc.want {
				t.Fatalf("result code = %q, want %q", result.Code, tc.want)
			}
		})
	}
}

func TestVerifyCarrierManifestScanRejectsPackedShipmentWithWrongCarrier(t *testing.T) {
	store := NewPrototypeCarrierManifestStore()
	service := NewVerifyCarrierManifestScan(store, audit.NewInMemoryLogStore())

	result, err := service.Execute(context.Background(), VerifyCarrierManifestScanInput{
		ManifestID: "manifest-hcm-ghn-morning",
		Code:       "VTP260426012",
		StationID:  "dock-a",
		ActorID:    "user-handover-operator",
		RequestID:  "req-scan-wrong-carrier",
	})
	if err != nil {
		t.Fatalf("verify scan: %v", err)
	}
	if result.Code != domain.ScanResultManifestMismatch ||
		result.Severity != "danger" ||
		!strings.Contains(result.Message, "carrier") {
		t.Fatalf("result = %+v, want carrier mismatch", result)
	}
	if result.Line == nil || result.Line.TrackingNo != "VTP260426012" {
		t.Fatalf("line = %+v, want wrong-carrier shipment context", result.Line)
	}
}

func draftCarrierManifestForActionTest(t *testing.T) domain.CarrierManifest {
	t.Helper()

	manifest, err := domain.NewCarrierManifest(domain.NewCarrierManifestInput{
		ID:               "manifest-hcm-ghn-action-test",
		CarrierCode:      "GHN",
		CarrierName:      "GHN Express",
		WarehouseID:      "wh-hcm",
		WarehouseCode:    "HCM",
		Date:             "2026-04-28",
		HandoverBatch:    "afternoon",
		StagingZone:      "handover-a",
		HandoverZoneCode: "HANDOVER-A",
		HandoverBinCode:  "TOTE-A01",
		Owner:            "Warehouse Lead",
	})
	if err != nil {
		t.Fatalf("new carrier manifest: %v", err)
	}

	return manifest
}
