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
