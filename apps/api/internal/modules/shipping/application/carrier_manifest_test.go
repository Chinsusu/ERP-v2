package application

import (
	"context"
	"testing"

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
		CarrierName:   "Ninja Van",
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
	if result.AuditLogID == "" {
		t.Fatal("audit log id is empty")
	}

	logs, err := auditStore.List(context.Background(), audit.Query{Action: "shipping.manifest.created"})
	if err != nil {
		t.Fatalf("list audit logs: %v", err)
	}
	if len(logs) != 1 {
		t.Fatalf("audit logs = %d, want 1", len(logs))
	}
}

func TestAddShipmentToCarrierManifestUpdatesCountsAndAudit(t *testing.T) {
	store := NewPrototypeCarrierManifestStore()
	auditStore := audit.NewInMemoryLogStore()
	service := NewAddShipmentToCarrierManifest(store, auditStore)

	result, err := service.Execute(context.Background(), AddShipmentToCarrierManifestInput{
		ManifestID: "manifest-hcm-ghn-morning",
		ShipmentID: "ship-hcm-260426-004",
		ActorID:    "user-warehouse-lead",
		RequestID:  "req-add-shipment",
	})
	if err != nil {
		t.Fatalf("add shipment: %v", err)
	}
	if got := result.Manifest.Summary(); got.ExpectedCount != 4 || got.ScannedCount != 2 || got.MissingCount != 2 {
		t.Fatalf("summary = %+v, want 4 expected, 2 scanned, 2 missing", got)
	}

	logs, err := auditStore.List(context.Background(), audit.Query{Action: "shipping.manifest.shipment_added"})
	if err != nil {
		t.Fatalf("list audit logs: %v", err)
	}
	if len(logs) != 1 {
		t.Fatalf("audit logs = %d, want 1", len(logs))
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

	logs, err := auditStore.List(context.Background(), audit.Query{Action: "shipping.manifest.scan_recorded"})
	if err != nil {
		t.Fatalf("list audit logs: %v", err)
	}
	if len(logs) != 1 {
		t.Fatalf("audit logs = %d, want 1", len(logs))
	}
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
