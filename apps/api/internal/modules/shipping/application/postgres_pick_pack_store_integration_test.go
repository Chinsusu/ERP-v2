package application

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"testing"
	"time"

	salesdomain "github.com/Chinsusu/ERP-v2/apps/api/internal/modules/sales/domain"
	"github.com/Chinsusu/ERP-v2/apps/api/internal/modules/shipping/domain"
	"github.com/Chinsusu/ERP-v2/apps/api/internal/shared/audit"

	_ "github.com/jackc/pgx/v5/stdlib"
)

type pickPackPersistenceFixture struct {
	OrgID             string
	UserID            string
	UnitID            string
	CustomerID        string
	ItemID            string
	ItemSKU           string
	WarehouseID       string
	WarehouseCode     string
	ZoneID            string
	BinID             string
	BinCode           string
	BatchID           string
	BatchRef          string
	BatchNo           string
	SalesOrderID      string
	SalesOrderRef     string
	OrderNo           string
	SalesOrderLineID  string
	SalesOrderLineRef string
	ReservationID     string
	ReservationRef    string
	PickRef           string
	PickNo            string
	PickLineRef       string
	PackRef           string
	PackNo            string
	PackLineRef       string
}

func TestPostgresPickAndPackStoresPersistLifecycle(t *testing.T) {
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

	fixture := newPickPackPersistenceFixture()
	if err := seedPickPackPersistenceFixture(ctx, db, fixture); err != nil {
		t.Fatalf("seed pick/pack persistence fixture: %v", err)
	}

	auditStore := audit.NewInMemoryLogStore()
	pickStore := NewPostgresPickTaskStore(
		db,
		PostgresPickTaskStoreConfig{DefaultOrgID: fixture.OrgID},
	)
	pickTask := newTestPostgresPickTask(t, fixture)
	if err := pickStore.SavePickTask(ctx, pickTask); err != nil {
		t.Fatalf("SavePickTask(created): %v", err)
	}
	if _, err := NewStartPickTask(pickStore, auditStore).Execute(ctx, PickTaskActionInput{
		PickTaskID: fixture.PickRef,
		ActorID:    fixture.UserID,
		RequestID:  "req-s14-pick-start",
	}); err != nil {
		t.Fatalf("start pick task: %v", err)
	}
	if _, err := NewConfirmPickTaskLine(pickStore, auditStore).Execute(ctx, ConfirmPickTaskLineInput{
		PickTaskID: fixture.PickRef,
		LineID:     fixture.PickLineRef,
		PickedQty:  "3.000000",
		ActorID:    fixture.UserID,
		RequestID:  "req-s14-pick-line",
	}); err != nil {
		t.Fatalf("confirm pick line: %v", err)
	}
	if _, err := NewCompletePickTask(pickStore, auditStore).Execute(ctx, PickTaskActionInput{
		PickTaskID: fixture.PickRef,
		ActorID:    fixture.UserID,
		RequestID:  "req-s14-pick-complete",
	}); err != nil {
		t.Fatalf("complete pick task: %v", err)
	}

	reloadedPickStore := NewPostgresPickTaskStore(
		db,
		PostgresPickTaskStoreConfig{DefaultOrgID: fixture.OrgID},
	)
	reloadedPick, err := reloadedPickStore.GetPickTask(ctx, fixture.PickRef)
	if err != nil {
		t.Fatalf("GetPickTask(%q): %v", fixture.PickRef, err)
	}
	if reloadedPick.Status != domain.PickTaskStatusCompleted ||
		reloadedPick.AssignedTo != fixture.UserID ||
		reloadedPick.CompletedBy != fixture.UserID ||
		len(reloadedPick.Lines) != 1 ||
		reloadedPick.Lines[0].Status != domain.PickTaskLineStatusPicked ||
		reloadedPick.Lines[0].QtyPicked.String() != "3.000000" {
		t.Fatalf("reloaded pick task = %+v, want completed picked line", reloadedPick)
	}
	if bySalesOrder, err := reloadedPickStore.GetPickTaskBySalesOrder(ctx, fixture.SalesOrderRef); err != nil || bySalesOrder.ID != fixture.PickRef {
		t.Fatalf("GetPickTaskBySalesOrder(%q) = %+v, %v", fixture.SalesOrderRef, bySalesOrder, err)
	}
	filteredPickTasks, err := NewListPickTasks(reloadedPickStore).Execute(ctx, PickTaskFilter{
		WarehouseID: fixture.WarehouseCode,
		Status:      domain.PickTaskStatusCompleted,
		AssignedTo:  fixture.UserID,
	})
	if err != nil {
		t.Fatalf("filtered pick list: %v", err)
	}
	if !containsPickTask(filteredPickTasks, fixture.PickRef) {
		t.Fatalf("filtered pick list missing %s: %+v", fixture.PickRef, filteredPickTasks)
	}

	packStore := NewPostgresPackTaskStore(
		db,
		PostgresPackTaskStoreConfig{DefaultOrgID: fixture.OrgID},
	)
	packTask := newTestPostgresPackTask(t, fixture)
	if err := packStore.SavePackTask(ctx, packTask); err != nil {
		t.Fatalf("SavePackTask(created): %v", err)
	}
	if _, err := NewStartPackTask(packStore, auditStore).Execute(ctx, PackTaskActionInput{
		PackTaskID: fixture.PackRef,
		ActorID:    fixture.UserID,
		RequestID:  "req-s14-pack-start",
	}); err != nil {
		t.Fatalf("start pack task: %v", err)
	}
	if _, err := NewConfirmPackTask(packStore, auditStore, testPackTaskPacker{}).Execute(ctx, ConfirmPackTaskInput{
		PackTaskID: fixture.PackRef,
		Lines: []ConfirmPackTaskLineInput{{
			LineID:    fixture.PackLineRef,
			PackedQty: "3.000000",
		}},
		ActorID:   fixture.UserID,
		RequestID: "req-s14-pack-confirm",
	}); err != nil {
		t.Fatalf("confirm pack task: %v", err)
	}

	reloadedPackStore := NewPostgresPackTaskStore(
		db,
		PostgresPackTaskStoreConfig{DefaultOrgID: fixture.OrgID},
	)
	reloadedPack, err := reloadedPackStore.GetPackTask(ctx, fixture.PackRef)
	if err != nil {
		t.Fatalf("GetPackTask(%q): %v", fixture.PackRef, err)
	}
	if reloadedPack.Status != domain.PackTaskStatusPacked ||
		reloadedPack.PackedBy != fixture.UserID ||
		len(reloadedPack.Lines) != 1 ||
		reloadedPack.Lines[0].Status != domain.PackTaskLineStatusPacked ||
		reloadedPack.Lines[0].QtyPacked.String() != "3.000000" {
		t.Fatalf("reloaded pack task = %+v, want packed line", reloadedPack)
	}
	if bySalesOrder, err := reloadedPackStore.GetPackTaskBySalesOrder(ctx, fixture.SalesOrderRef); err != nil || bySalesOrder.ID != fixture.PackRef {
		t.Fatalf("GetPackTaskBySalesOrder(%q) = %+v, %v", fixture.SalesOrderRef, bySalesOrder, err)
	}
	if byPickTask, err := reloadedPackStore.GetPackTaskByPickTask(ctx, fixture.PickRef); err != nil || byPickTask.ID != fixture.PackRef {
		t.Fatalf("GetPackTaskByPickTask(%q) = %+v, %v", fixture.PickRef, byPickTask, err)
	}
	filteredPackTasks, err := NewListPackTasks(reloadedPackStore).Execute(ctx, PackTaskFilter{
		WarehouseID: fixture.WarehouseCode,
		Status:      domain.PackTaskStatusPacked,
		AssignedTo:  fixture.UserID,
	})
	if err != nil {
		t.Fatalf("filtered pack list: %v", err)
	}
	if !containsPackTask(filteredPackTasks, fixture.PackRef) {
		t.Fatalf("filtered pack list missing %s: %+v", fixture.PackRef, filteredPackTasks)
	}

	assertAuditRecorded(t, auditStore, "shipping.pick_task.started")
	assertAuditRecorded(t, auditStore, "shipping.pick_task.line_confirmed")
	assertAuditRecorded(t, auditStore, "shipping.pick_task.completed")
	assertAuditRecorded(t, auditStore, "shipping.pack_task.started")
	assertAuditRecorded(t, auditStore, "shipping.pack_task.confirmed")
}

func newPickPackPersistenceFixture() pickPackPersistenceFixture {
	seed := time.Now().UTC().UnixNano() % 0x1000000000000
	suffix := fmt.Sprintf("%012x", seed)

	return pickPackPersistenceFixture{
		OrgID:             pickPackPersistenceUUID(1, seed),
		UserID:            pickPackPersistenceUUID(2, seed),
		UnitID:            pickPackPersistenceUUID(3, seed),
		CustomerID:        pickPackPersistenceUUID(4, seed),
		ItemID:            pickPackPersistenceUUID(5, seed),
		ItemSKU:           "S14-PP-" + suffix,
		WarehouseID:       pickPackPersistenceUUID(6, seed),
		WarehouseCode:     "WH-S14-PP-" + suffix,
		ZoneID:            pickPackPersistenceUUID(7, seed),
		BinID:             pickPackPersistenceUUID(8, seed),
		BinCode:           "BIN-S14-PP-" + suffix,
		BatchID:           pickPackPersistenceUUID(9, seed),
		BatchRef:          "batch-s14-pp-" + suffix,
		BatchNo:           "BATCH-S14-PP-" + suffix,
		SalesOrderID:      pickPackPersistenceUUID(10, seed),
		SalesOrderRef:     "so-s14-pp-" + suffix,
		OrderNo:           "SO-S14-PP-" + suffix,
		SalesOrderLineID:  pickPackPersistenceUUID(11, seed),
		SalesOrderLineRef: "so-s14-pp-" + suffix + "-line-01",
		ReservationID:     pickPackPersistenceUUID(12, seed),
		ReservationRef:    "res-s14-pp-" + suffix,
		PickRef:           "pick-s14-pp-" + suffix,
		PickNo:            "PICK-S14-PP-" + suffix,
		PickLineRef:       "pick-s14-pp-" + suffix + "-line-01",
		PackRef:           "pack-s14-pp-" + suffix,
		PackNo:            "PACK-S14-PP-" + suffix,
		PackLineRef:       "pack-s14-pp-" + suffix + "-line-01",
	}
}

func pickPackPersistenceUUID(offset int, seed int64) string {
	return fmt.Sprintf("00000000-0000-4000-%04x-%012x", 0x8000+offset, seed)
}

func newTestPostgresPickTask(t *testing.T, fixture pickPackPersistenceFixture) domain.PickTask {
	t.Helper()

	task, err := domain.NewPickTask(domain.NewPickTaskInput{
		ID:            fixture.PickRef,
		OrgID:         fixture.OrgID,
		PickTaskNo:    fixture.PickNo,
		SalesOrderID:  fixture.SalesOrderRef,
		OrderNo:       fixture.OrderNo,
		WarehouseID:   fixture.WarehouseCode,
		WarehouseCode: fixture.WarehouseCode,
		CreatedAt:     time.Date(2026, 5, 2, 8, 0, 0, 0, time.UTC),
		Lines: []domain.NewPickTaskLineInput{{
			ID:                 fixture.PickLineRef,
			LineNo:             1,
			SalesOrderLineID:   fixture.SalesOrderLineRef,
			StockReservationID: fixture.ReservationRef,
			ItemID:             fixture.ItemSKU,
			SKUCode:            fixture.ItemSKU,
			BatchID:            fixture.BatchRef,
			BatchNo:            fixture.BatchNo,
			WarehouseID:        fixture.WarehouseCode,
			BinID:              fixture.BinCode,
			BinCode:            fixture.BinCode,
			QtyToPick:          "3.000000",
			BaseUOMCode:        "EA",
		}},
	})
	if err != nil {
		t.Fatalf("new pick task: %v", err)
	}

	return task
}

func newTestPostgresPackTask(t *testing.T, fixture pickPackPersistenceFixture) domain.PackTask {
	t.Helper()

	task, err := domain.NewPackTask(domain.NewPackTaskInput{
		ID:            fixture.PackRef,
		OrgID:         fixture.OrgID,
		PackTaskNo:    fixture.PackNo,
		SalesOrderID:  fixture.SalesOrderRef,
		OrderNo:       fixture.OrderNo,
		PickTaskID:    fixture.PickRef,
		PickTaskNo:    fixture.PickNo,
		WarehouseID:   fixture.WarehouseCode,
		WarehouseCode: fixture.WarehouseCode,
		AssignedTo:    fixture.UserID,
		AssignedAt:    time.Date(2026, 5, 2, 8, 30, 0, 0, time.UTC),
		CreatedAt:     time.Date(2026, 5, 2, 8, 30, 0, 0, time.UTC),
		Lines: []domain.NewPackTaskLineInput{{
			ID:               fixture.PackLineRef,
			LineNo:           1,
			PickTaskLineID:   fixture.PickLineRef,
			SalesOrderLineID: fixture.SalesOrderLineRef,
			ItemID:           fixture.ItemSKU,
			SKUCode:          fixture.ItemSKU,
			BatchID:          fixture.BatchRef,
			BatchNo:          fixture.BatchNo,
			WarehouseID:      fixture.WarehouseCode,
			QtyToPack:        "3.000000",
			BaseUOMCode:      "EA",
		}},
	})
	if err != nil {
		t.Fatalf("new pack task: %v", err)
	}

	return task
}

func seedPickPackPersistenceFixture(ctx context.Context, db *sql.DB, fixture pickPackPersistenceFixture) error {
	statements := []struct {
		query string
		args  []any
	}{
		{
			query: `INSERT INTO core.organizations (id, code, name, status)
VALUES ($1, $2, 'ERP S14 Pick Pack Test', 'active')
ON CONFLICT (code) DO UPDATE
SET name = EXCLUDED.name,
    status = EXCLUDED.status,
    updated_at = now()`,
			args: []any{fixture.OrgID, "ERP_S14_PP_" + fixture.SalesOrderRef},
		},
		{
			query: `INSERT INTO core.users (id, org_id, email, username, display_name, status)
SELECT $1::uuid, $2::uuid, $3, $4, 'Sprint 14 Warehouse User', 'active'
WHERE NOT EXISTS (
  SELECT 1
  FROM core.users existing
  WHERE existing.org_id = $2::uuid
    AND lower(existing.username) = lower($4)
)`,
			args: []any{fixture.UserID, fixture.OrgID, fixture.SalesOrderRef + "@example.local", "user_" + fixture.SalesOrderRef},
		},
		{
			query: `INSERT INTO mdm.units (id, org_id, code, name, precision_scale, status)
VALUES ($1, $2, 'EA', 'Each', 0, 'active')
ON CONFLICT (org_id, code) DO UPDATE
SET name = EXCLUDED.name,
    status = EXCLUDED.status,
    updated_at = now()`,
			args: []any{fixture.UnitID, fixture.OrgID},
		},
		{
			query: `INSERT INTO mdm.customers (id, org_id, code, name, status)
VALUES ($1, $2, $3, 'Sprint 14 Customer', 'active')
ON CONFLICT (org_id, code) DO UPDATE
SET name = EXCLUDED.name,
    status = EXCLUDED.status,
    updated_at = now()`,
			args: []any{fixture.CustomerID, fixture.OrgID, "CUST-" + fixture.SalesOrderRef},
		},
		{
			query: `INSERT INTO mdm.items (id, org_id, sku, name, item_type, base_unit_id, requires_batch, requires_expiry, status)
VALUES ($1, $2, $3, 'Sprint 14 Pick Pack Item', 'finished_good', $4, true, true, 'active')
ON CONFLICT (org_id, sku) DO UPDATE
SET name = EXCLUDED.name,
    base_unit_id = EXCLUDED.base_unit_id,
    status = EXCLUDED.status,
    updated_at = now()`,
			args: []any{fixture.ItemID, fixture.OrgID, fixture.ItemSKU, fixture.UnitID},
		},
		{
			query: `INSERT INTO mdm.warehouses (id, org_id, code, name, status)
VALUES ($1, $2, $3, 'Sprint 14 Pick Pack Warehouse', 'active')
ON CONFLICT (org_id, code) DO UPDATE
SET name = EXCLUDED.name,
    status = EXCLUDED.status,
    updated_at = now()`,
			args: []any{fixture.WarehouseID, fixture.OrgID, fixture.WarehouseCode},
		},
		{
			query: `INSERT INTO mdm.warehouse_zones (id, org_id, warehouse_id, code, name, zone_type, status)
VALUES ($1, $2, $3, 'PICK', 'Sprint 14 Pick Zone', 'pick', 'active')
ON CONFLICT (warehouse_id, code) DO UPDATE
SET name = EXCLUDED.name,
    status = EXCLUDED.status,
    updated_at = now()`,
			args: []any{fixture.ZoneID, fixture.OrgID, fixture.WarehouseID},
		},
		{
			query: `INSERT INTO mdm.warehouse_bins (id, org_id, warehouse_id, zone_id, code, name, bin_type, status)
VALUES ($1, $2, $3, $4, $5, 'Sprint 14 Pick Bin', 'pick', 'active')
ON CONFLICT (warehouse_id, code) DO UPDATE
SET name = EXCLUDED.name,
    status = EXCLUDED.status,
    updated_at = now()`,
			args: []any{fixture.BinID, fixture.OrgID, fixture.WarehouseID, fixture.ZoneID, fixture.BinCode},
		},
		{
			query: `INSERT INTO inventory.batches (
  id,
  org_id,
  item_id,
  batch_no,
  expiry_date,
  qc_status,
  status,
  batch_ref,
  org_ref,
  item_ref
) VALUES (
  $1,
  $2::uuid,
  $3,
  $4,
  '2027-05-02',
  'pass',
  'active',
  $5,
  $2::text,
  $6
)
ON CONFLICT (item_id, batch_no) DO UPDATE
SET qc_status = EXCLUDED.qc_status,
    status = EXCLUDED.status,
    batch_ref = EXCLUDED.batch_ref,
    updated_at = now()`,
			args: []any{fixture.BatchID, fixture.OrgID, fixture.ItemID, fixture.BatchNo, fixture.BatchRef, fixture.ItemSKU},
		},
		{
			query: `INSERT INTO sales.sales_orders (
  id,
  org_id,
  order_no,
  customer_id,
  order_date,
  channel,
  status,
  currency_code,
  total_amount,
  subtotal_amount,
  net_amount,
  order_ref,
  org_ref,
  customer_ref,
  customer_code,
  customer_name,
  warehouse_id,
  warehouse_ref,
  warehouse_code,
  created_at,
  updated_at
) VALUES (
  $1,
  $2::uuid,
  $3,
  $4::uuid,
  '2026-05-02',
  'online',
  'packing',
  'VND',
  300000,
  300000,
  300000,
  $5,
  $2::text,
  $4::text,
  $6,
  'Sprint 14 Customer',
  $7::uuid,
  $8,
  $8,
  now(),
  now()
)
ON CONFLICT (org_id, order_no) DO UPDATE
SET status = EXCLUDED.status,
    order_ref = EXCLUDED.order_ref,
    warehouse_id = EXCLUDED.warehouse_id,
    warehouse_ref = EXCLUDED.warehouse_ref,
    warehouse_code = EXCLUDED.warehouse_code,
    updated_at = now()`,
			args: []any{fixture.SalesOrderID, fixture.OrgID, fixture.OrderNo, fixture.CustomerID, fixture.SalesOrderRef, "CUST-" + fixture.SalesOrderRef, fixture.WarehouseID, fixture.WarehouseCode},
		},
		{
			query: `INSERT INTO sales.sales_order_lines (
  id,
  org_id,
  sales_order_id,
  line_no,
  item_id,
  unit_id,
  ordered_qty,
  reserved_qty,
  unit_price,
  uom_code,
  base_ordered_qty,
  base_uom_code,
  conversion_factor,
  currency_code,
  line_amount,
  line_ref,
  item_ref,
  sku_code,
  item_name,
  batch_id,
  batch_ref,
  batch_no,
  created_at,
  updated_at
) VALUES (
  $1,
  $2::uuid,
  $3,
  1,
  $4,
  $5,
  3.000000,
  3.000000,
  100000,
  'EA',
  3.000000,
  'EA',
  1.000000,
  'VND',
  300000,
  $6,
  $7,
  $7,
  'Sprint 14 Pick Pack Item',
  $8,
  $9,
  $10,
  now(),
  now()
)
ON CONFLICT (sales_order_id, line_no) DO UPDATE
SET reserved_qty = EXCLUDED.reserved_qty,
    line_ref = EXCLUDED.line_ref,
    item_ref = EXCLUDED.item_ref,
    batch_id = EXCLUDED.batch_id,
    batch_ref = EXCLUDED.batch_ref,
    batch_no = EXCLUDED.batch_no,
    updated_at = now()`,
			args: []any{fixture.SalesOrderLineID, fixture.OrgID, fixture.SalesOrderID, fixture.ItemID, fixture.UnitID, fixture.SalesOrderLineRef, fixture.ItemSKU, fixture.BatchID, fixture.BatchRef, fixture.BatchNo},
		},
		{
			query: `INSERT INTO inventory.stock_reservations (
  id,
  org_id,
  reservation_no,
  item_id,
  batch_id,
  warehouse_id,
  reserved_qty,
  source_doc_type,
  source_doc_id,
  status,
  created_at,
  created_by,
  sales_order_id,
  sales_order_line_id,
  source_doc_line_id,
  bin_id,
  base_uom_code,
  stock_status,
  updated_at,
  reservation_ref,
  org_ref,
  sales_order_ref,
  sales_order_line_ref,
  source_doc_ref,
  source_doc_line_ref,
  item_ref,
  sku_code,
  batch_ref,
  batch_no,
  warehouse_ref,
  warehouse_code,
  bin_ref,
  bin_code,
  created_by_ref
) VALUES (
  $1,
  $2::uuid,
  $3,
  $4,
  $5,
  $6,
  3.000000,
  'sales_order',
  $7,
  'active',
  now(),
  $8::uuid,
  $7,
  $9,
  $9,
  $10,
  'EA',
  'available',
  now(),
  $11,
  $2::text,
  $12,
  $13,
  $12,
  $13,
  $14,
  $14,
  $15,
  $16,
  $17,
  $17,
  $18,
  $18,
  $8::text
)
ON CONFLICT (org_id, reservation_no) DO UPDATE
SET status = EXCLUDED.status,
    reservation_ref = EXCLUDED.reservation_ref,
    sales_order_ref = EXCLUDED.sales_order_ref,
    sales_order_line_ref = EXCLUDED.sales_order_line_ref,
    item_ref = EXCLUDED.item_ref,
    batch_ref = EXCLUDED.batch_ref,
    warehouse_ref = EXCLUDED.warehouse_ref,
    bin_ref = EXCLUDED.bin_ref,
    updated_at = now()`,
			args: []any{fixture.ReservationID, fixture.OrgID, "RES-" + fixture.SalesOrderRef, fixture.ItemID, fixture.BatchID, fixture.WarehouseID, fixture.SalesOrderID, fixture.UserID, fixture.SalesOrderLineID, fixture.BinID, fixture.ReservationRef, fixture.SalesOrderRef, fixture.SalesOrderLineRef, fixture.ItemSKU, fixture.BatchRef, fixture.BatchNo, fixture.WarehouseCode, fixture.BinCode},
		},
	}
	for _, statement := range statements {
		if _, err := db.ExecContext(ctx, statement.query, statement.args...); err != nil {
			return err
		}
	}

	return nil
}

func containsPickTask(tasks []domain.PickTask, id string) bool {
	for _, task := range tasks {
		if task.ID == id {
			return true
		}
	}

	return false
}

func containsPackTask(tasks []domain.PackTask, id string) bool {
	for _, task := range tasks {
		if task.ID == id {
			return true
		}
	}

	return false
}

func assertAuditRecorded(t *testing.T, store audit.LogStore, action string) {
	t.Helper()

	logs, err := store.List(context.Background(), audit.Query{Action: action})
	if err != nil {
		t.Fatalf("list audit %s: %v", action, err)
	}
	if len(logs) == 0 {
		t.Fatalf("audit action %s was not recorded", action)
	}
}

type testPackTaskPacker struct{}

func (testPackTaskPacker) MarkSalesOrderPacked(
	_ context.Context,
	input PackTaskSalesOrderPackedInput,
) (salesdomain.SalesOrder, error) {
	return salesdomain.SalesOrder{
		ID:     input.SalesOrderID,
		Status: salesdomain.SalesOrderStatusPacked,
	}, nil
}
