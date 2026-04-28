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

func TestNormalizeCarrierManifestStatusIncludesHandoverStates(t *testing.T) {
	cases := []CarrierManifestStatus{
		ManifestStatusDraft,
		ManifestStatusReady,
		ManifestStatusScanning,
		ManifestStatusCompleted,
		ManifestStatusHandedOver,
		ManifestStatusException,
		ManifestStatusCancelled,
	}

	for _, status := range cases {
		if NormalizeManifestStatus(status) != status {
			t.Fatalf("normalize %q = %q, want same status", status, NormalizeManifestStatus(status))
		}
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

	if updated.Status != ManifestStatusDraft {
		t.Fatalf("status = %q, want draft until ready-to-scan action", updated.Status)
	}
	if got := updated.Summary(); got.ExpectedCount != 1 || got.ScannedCount != 0 || got.MissingCount != 1 {
		t.Fatalf("summary = %+v, want 1 expected, 0 scanned, 1 missing", got)
	}
}

func TestCarrierManifestReadyRemoveAndCancel(t *testing.T) {
	manifest, err := NewCarrierManifest(NewCarrierManifestInput{
		CarrierCode: "GHN",
		WarehouseID: "wh-hcm",
		Date:        "2026-04-26",
	})
	if err != nil {
		t.Fatalf("new manifest: %v", err)
	}

	if _, err := manifest.MarkReadyToScan(); !errors.Is(err, ErrManifestRequiredField) {
		t.Fatalf("ready empty manifest err = %v, want required field", err)
	}

	withShipment, err := manifest.AddShipment(PackedShipment{ID: "ship-001", OrderNo: "SO-001", Packed: true})
	if err != nil {
		t.Fatalf("add shipment: %v", err)
	}
	ready, err := withShipment.MarkReadyToScan()
	if err != nil {
		t.Fatalf("mark ready: %v", err)
	}
	if ready.Status != ManifestStatusReady {
		t.Fatalf("ready status = %q, want ready", ready.Status)
	}

	removed, err := ready.RemoveShipment("ship-001")
	if err != nil {
		t.Fatalf("remove shipment: %v", err)
	}
	if removed.Status != ManifestStatusDraft || len(removed.Lines) != 0 {
		t.Fatalf("removed manifest = %+v, want empty draft manifest", removed)
	}

	cancelled, err := removed.Cancel()
	if err != nil {
		t.Fatalf("cancel manifest: %v", err)
	}
	if cancelled.Status != ManifestStatusCancelled {
		t.Fatalf("cancelled status = %q, want cancelled", cancelled.Status)
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

func TestCarrierManifestScansByOrderOrTrackingCode(t *testing.T) {
	manifest := CarrierManifest{
		ID:     "manifest-hcm-ghn-morning",
		Status: ManifestStatusReady,
		Lines: []CarrierManifestLine{
			{
				ID:         "line-001",
				ShipmentID: "ship-001",
				OrderNo:    "SO-001",
				TrackingNo: "GHN001",
			},
		},
	}

	updated, line, err := manifest.MarkLineScanned(" ghn001 ")
	if err != nil {
		t.Fatalf("scan line: %v", err)
	}
	if line.ID != "line-001" || !updated.Lines[0].Scanned {
		t.Fatalf("scan result = %+v, manifest = %+v", line, updated)
	}
	if updated.Status != ManifestStatusScanning {
		t.Fatalf("status = %q, want scanning", updated.Status)
	}
}

func TestCarrierManifestScanRejectsDuplicateUnknownAndInvalidState(t *testing.T) {
	manifest := CarrierManifest{
		ID:     "manifest-hcm-ghn-morning",
		Status: ManifestStatusScanning,
		Lines: []CarrierManifestLine{
			{
				ID:         "line-001",
				ShipmentID: "ship-001",
				OrderNo:    "SO-001",
				TrackingNo: "GHN001",
				Scanned:    true,
			},
		},
	}

	if _, _, err := manifest.MarkLineScanned("GHN001"); !errors.Is(err, ErrManifestScanDuplicate) {
		t.Fatalf("err = %v, want duplicate scan", err)
	}
	if _, _, err := manifest.MarkLineScanned("GHN999"); !errors.Is(err, ErrManifestScanNotFound) {
		t.Fatalf("err = %v, want not found", err)
	}

	manifest.Status = ManifestStatusCompleted
	if _, _, err := manifest.MarkLineScanned("SO-001"); !errors.Is(err, ErrManifestScanInvalidState) {
		t.Fatalf("err = %v, want invalid state", err)
	}
}
