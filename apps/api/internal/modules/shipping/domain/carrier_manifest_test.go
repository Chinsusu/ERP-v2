package domain

import (
	"errors"
	"testing"
	"time"
)

func TestNewCarrierManifestDefaultsToDraft(t *testing.T) {
	manifest, err := NewCarrierManifest(NewCarrierManifestInput{
		CarrierCode:   "ghn",
		WarehouseID:   "wh-hcm",
		WarehouseCode: "HCM",
		Date:          "2026-04-26",
		CreatedAt:     time.Date(2026, 4, 26, 8, 0, 0, 0, time.UTC),
	})
	if err != nil {
		t.Fatalf("new manifest: %v", err)
	}

	if manifest.Status != ManifestStatusDraft {
		t.Fatalf("status = %q, want draft", manifest.Status)
	}
	if manifest.CarrierCode != "GHN" || manifest.HandoverBatch != "day" || manifest.StagingZone != "handover" {
		t.Fatalf("manifest defaults = %+v", manifest)
	}
}

func TestCarrierManifestAddShipmentUpdatesExpectedAndMissingCounts(t *testing.T) {
	manifest, err := NewCarrierManifest(NewCarrierManifestInput{
		CarrierCode: "VTP",
		WarehouseID: "wh-hcm",
		Date:        "2026-04-26",
	})
	if err != nil {
		t.Fatalf("new manifest: %v", err)
	}

	updated, err := manifest.AddShipment(PackedShipment{
		ID:          "ship-001",
		OrderNo:     "SO-001",
		TrackingNo:  "TRK-001",
		PackageCode: "BOX-1",
		Packed:      true,
	})
	if err != nil {
		t.Fatalf("add shipment: %v", err)
	}

	if updated.Status != ManifestStatusReady {
		t.Fatalf("status = %q, want ready", updated.Status)
	}
	if got := updated.Summary(); got.ExpectedCount != 1 || got.ScannedCount != 0 || got.MissingCount != 1 {
		t.Fatalf("summary = %+v, want 1 expected, 0 scanned, 1 missing", got)
	}
}

func TestCarrierManifestRejectsDuplicateAndUnpackedShipments(t *testing.T) {
	manifest, err := NewCarrierManifest(NewCarrierManifestInput{
		CarrierCode: "GHN",
		WarehouseID: "wh-hcm",
		Date:        "2026-04-26",
	})
	if err != nil {
		t.Fatalf("new manifest: %v", err)
	}

	if _, err := manifest.AddShipment(PackedShipment{ID: "ship-unpacked"}); !errors.Is(err, ErrManifestShipmentNotPacked) {
		t.Fatalf("err = %v, want unpacked error", err)
	}

	updated, err := manifest.AddShipment(PackedShipment{ID: "ship-001", Packed: true})
	if err != nil {
		t.Fatalf("add shipment: %v", err)
	}
	if _, err := updated.AddShipment(PackedShipment{ID: "ship-001", Packed: true}); !errors.Is(err, ErrManifestDuplicateShipment) {
		t.Fatalf("err = %v, want duplicate error", err)
	}
}
