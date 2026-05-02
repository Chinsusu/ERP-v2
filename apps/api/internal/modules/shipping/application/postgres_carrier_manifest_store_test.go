package application

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/Chinsusu/ERP-v2/apps/api/internal/modules/shipping/domain"

	_ "github.com/jackc/pgx/v5/stdlib"
)

const (
	testCarrierManifestOrgID     = "00000000-0000-4000-8000-000000140103"
	testCarrierManifestWarehouse = "00000000-0000-4000-8000-000000140104"
	testCarrierManifestCarrier   = "00000000-0000-4000-8000-000000140105"
	testCarrierManifestShipment  = "00000000-0000-4000-8000-000000140106"
)

func TestPostgresCarrierManifestStoreRequiresDatabase(t *testing.T) {
	store := NewPostgresCarrierManifestStore(nil, PostgresCarrierManifestStoreConfig{})

	if _, err := store.List(context.Background(), domain.CarrierManifestFilter{}); err == nil {
		t.Fatal("List() error = nil, want database required error")
	}
	if _, err := store.Get(context.Background(), "manifest-s14"); err == nil {
		t.Fatal("Get() error = nil, want database required error")
	}
	if err := store.Save(context.Background(), domain.CarrierManifest{ID: "manifest-s14"}); err == nil {
		t.Fatal("Save() error = nil, want database required error")
	}
	if _, err := store.GetPackedShipment(context.Background(), "ship-s14"); err == nil {
		t.Fatal("GetPackedShipment() error = nil, want database required error")
	}
}

func TestPostgresCarrierManifestStorePersistsManifestScanAndHandover(t *testing.T) {
	databaseURL := os.Getenv("ERP_TEST_DATABASE_URL")
	if databaseURL == "" {
		t.Skip("ERP_TEST_DATABASE_URL is not set")
	}

	ctx := context.Background()
	db, err := sql.Open("pgx", databaseURL)
	if err != nil {
		t.Fatalf("open database: %v", err)
	}
	defer db.Close()

	suffix := fmt.Sprintf("%d", time.Now().UTC().UnixNano())
	shipmentNo := "ship-s14-cm-" + suffix
	trackingNo := "S14TRACK" + suffix
	if err := seedCarrierManifestFixture(ctx, db, shipmentNo, trackingNo); err != nil {
		t.Fatalf("seed carrier manifest fixture: %v", err)
	}

	store := NewPostgresCarrierManifestStore(
		db,
		PostgresCarrierManifestStoreConfig{DefaultOrgID: testCarrierManifestOrgID},
	)
	shipment, err := store.GetPackedShipment(ctx, shipmentNo)
	if err != nil {
		t.Fatalf("GetPackedShipment(%q): %v", shipmentNo, err)
	}
	if !shipment.Packed || shipment.ID != shipmentNo || shipment.TrackingNo != trackingNo {
		t.Fatalf("shipment = %+v, want packed %s/%s", shipment, shipmentNo, trackingNo)
	}
	if found, err := store.FindPackedShipmentByCode(ctx, trackingNo); err != nil || found.ID != shipmentNo {
		t.Fatalf("FindPackedShipmentByCode(%q) = %+v, %v", trackingNo, found, err)
	}

	createdAt := time.Date(2026, 5, 2, 8, 30, 0, 0, time.UTC)
	manifest, err := domain.NewCarrierManifest(domain.NewCarrierManifestInput{
		ID:            "manifest-s14-cm-" + suffix,
		CarrierCode:   "GHN-S14",
		CarrierName:   "GHN S14 Express",
		WarehouseID:   testCarrierManifestWarehouse,
		WarehouseCode: "WH-S14-CM",
		Date:          "2026-05-02",
		HandoverBatch: "morning",
		StagingZone:   "handover",
		Owner:         "user-warehouse-lead",
		CreatedAt:     createdAt,
	})
	if err != nil {
		t.Fatalf("new manifest: %v", err)
	}
	manifest, err = manifest.AddShipment(shipment)
	if err != nil {
		t.Fatalf("add shipment: %v", err)
	}
	manifest, err = manifest.MarkReadyToScan()
	if err != nil {
		t.Fatalf("mark ready: %v", err)
	}
	manifest, _, err = manifest.MarkLineScanned(trackingNo)
	if err != nil {
		t.Fatalf("mark line scanned: %v", err)
	}
	manifest, err = manifest.ConfirmHandover()
	if err != nil {
		t.Fatalf("confirm handover: %v", err)
	}
	if err := store.Save(ctx, manifest); err != nil {
		t.Fatalf("Save(): %v", err)
	}
	if err := store.RecordScanEvent(ctx, CarrierManifestScanEvent{
		ID:          "scan-s14-cm-" + suffix,
		ManifestID:  manifest.ID,
		Code:        trackingNo,
		ResultCode:  domain.ScanResultMatched,
		Severity:    "success",
		Message:     "Scan matched manifest line",
		ShipmentID:  shipmentNo,
		OrderNo:     shipment.OrderNo,
		TrackingNo:  trackingNo,
		ActorID:     "user-warehouse-lead",
		StationID:   "shipping-handover",
		DeviceID:    "scanner-s14",
		Source:      "shipping_handover",
		WarehouseID: testCarrierManifestWarehouse,
		CarrierCode: "GHN-S14",
		CreatedAt:   createdAt.Add(10 * time.Minute),
	}); err != nil {
		t.Fatalf("RecordScanEvent(): %v", err)
	}

	reloadedStore := NewPostgresCarrierManifestStore(
		db,
		PostgresCarrierManifestStoreConfig{DefaultOrgID: testCarrierManifestOrgID},
	)
	reloaded, err := reloadedStore.Get(ctx, manifest.ID)
	if err != nil {
		t.Fatalf("Get(%q): %v", manifest.ID, err)
	}
	if reloaded.Status != domain.ManifestStatusHandedOver ||
		len(reloaded.Lines) != 1 ||
		!reloaded.Lines[0].Scanned ||
		reloaded.Lines[0].TrackingNo != trackingNo {
		t.Fatalf("reloaded manifest = %+v, want handed over with scanned line", reloaded)
	}
	_, foundLine, err := reloadedStore.FindCarrierManifestLineByCode(ctx, trackingNo)
	if err != nil {
		t.Fatalf("FindCarrierManifestLineByCode(%q): %v", trackingNo, err)
	}
	if foundLine.ShipmentID != shipmentNo {
		t.Fatalf("found line shipment = %q, want %q", foundLine.ShipmentID, shipmentNo)
	}
	events, err := reloadedStore.ListScanEvents(ctx, manifest.ID)
	if err != nil {
		t.Fatalf("ListScanEvents(%q): %v", manifest.ID, err)
	}
	if len(events) == 0 || events[0].ResultCode != domain.ScanResultMatched || events[0].Code != trackingNo {
		t.Fatalf("scan events = %+v, want matched %q", events, trackingNo)
	}
	rows, err := reloadedStore.List(ctx, domain.NewCarrierManifestFilter(
		testCarrierManifestWarehouse,
		"2026-05-02",
		"GHN-S14",
		domain.ManifestStatusHandedOver,
	))
	if err != nil {
		t.Fatalf("List(): %v", err)
	}
	if !containsCarrierManifest(rows, manifest.ID) {
		t.Fatalf("filtered list missing %s: %+v", manifest.ID, rows)
	}
}

func containsCarrierManifest(rows []domain.CarrierManifest, id string) bool {
	for _, row := range rows {
		if row.ID == id {
			return true
		}
	}

	return false
}

func seedCarrierManifestFixture(ctx context.Context, db *sql.DB, shipmentNo string, trackingNo string) error {
	if _, err := db.ExecContext(
		ctx,
		`INSERT INTO core.organizations (id, code, name, status)
VALUES ($1, 'ERP_S14_CM', 'ERP S14 Carrier Manifest Test', 'active')
ON CONFLICT (code) DO UPDATE
SET name = EXCLUDED.name,
    status = EXCLUDED.status,
    updated_at = now()`,
		testCarrierManifestOrgID,
	); err != nil {
		return err
	}
	if _, err := db.ExecContext(
		ctx,
		`INSERT INTO mdm.warehouses (id, org_id, code, name, status)
VALUES ($1, $2, 'WH-S14-CM', 'Sprint 14 Carrier Manifest Warehouse', 'active')
ON CONFLICT (org_id, code) DO UPDATE
SET name = EXCLUDED.name,
    status = EXCLUDED.status,
    updated_at = now()`,
		testCarrierManifestWarehouse,
		testCarrierManifestOrgID,
	); err != nil {
		return err
	}
	if _, err := db.ExecContext(
		ctx,
		`INSERT INTO mdm.carriers (id, org_id, code, name, status)
VALUES ($1, $2, 'GHN-S14', 'GHN S14 Express', 'active')
ON CONFLICT (org_id, code) DO UPDATE
SET name = EXCLUDED.name,
    status = EXCLUDED.status,
    updated_at = now()`,
		testCarrierManifestCarrier,
		testCarrierManifestOrgID,
	); err != nil {
		return err
	}
	_, err := db.ExecContext(
		ctx,
		`INSERT INTO shipping.shipments (
  id,
  org_id,
  shipment_no,
  warehouse_id,
  carrier_id,
  tracking_no,
  status,
  packed_at,
  created_at,
  updated_at
) VALUES (
  $1,
  $2,
  $3,
  $4,
  $5,
  $6,
  'packed',
  now(),
  now(),
  now()
)
ON CONFLICT (id) DO UPDATE
SET shipment_no = EXCLUDED.shipment_no,
    tracking_no = EXCLUDED.tracking_no,
    status = EXCLUDED.status,
    packed_at = EXCLUDED.packed_at,
    updated_at = now()`,
		testCarrierManifestShipment,
		testCarrierManifestOrgID,
		shipmentNo,
		testCarrierManifestWarehouse,
		testCarrierManifestCarrier,
		trackingNo,
	)

	return err
}
