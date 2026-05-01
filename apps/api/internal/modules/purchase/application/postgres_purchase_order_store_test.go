package application

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	purchasedomain "github.com/Chinsusu/ERP-v2/apps/api/internal/modules/purchase/domain"
	"github.com/Chinsusu/ERP-v2/apps/api/internal/shared/audit"
	"github.com/Chinsusu/ERP-v2/apps/api/internal/shared/decimal"

	_ "github.com/jackc/pgx/v5/stdlib"
)

const testPurchaseOrderOrgID = "00000000-0000-4000-8000-000000000001"

func TestPostgresPurchaseOrderStorePersistsPurchaseOrderDocument(t *testing.T) {
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

	if err := seedPurchaseOrderSmokeOrg(ctx, db); err != nil {
		t.Fatalf("seed org: %v", err)
	}

	suffix := fmt.Sprintf("%d", time.Now().UTC().UnixNano())
	store := NewPostgresPurchaseOrderStore(
		db,
		PostgresPurchaseOrderStoreConfig{DefaultOrgID: testPurchaseOrderOrgID},
	)
	order, err := purchasedomain.NewPurchaseOrderDocument(purchasedomain.NewPurchaseOrderDocumentInput{
		ID:            "po-s11-03-02-" + suffix,
		OrgID:         "org-my-pham",
		PONo:          "PO-S11-03-02-" + suffix,
		SupplierID:    "sup-rm-bioactive",
		SupplierCode:  "SUP-RM-BIO",
		SupplierName:  "BioActive Raw Materials",
		WarehouseID:   "wh-hcm-rm",
		WarehouseCode: "WH-HCM-RM",
		ExpectedDate:  "2026-05-08",
		CurrencyCode:  "VND",
		Note:          "postgres purchase order smoke",
		Lines: []purchasedomain.NewPurchaseOrderLineInput{{
			ID:               "po-line-s11-03-02-" + suffix,
			LineNo:           1,
			ItemID:           "item-serum-30ml",
			SKUCode:          "SERUM-30ML",
			ItemName:         "Hydrating Serum 30ml",
			OrderedQty:       decimal.MustQuantity("2"),
			ReceivedQty:      decimal.MustQuantity("0"),
			UOMCode:          "PCS",
			BaseOrderedQty:   decimal.MustQuantity("2"),
			BaseReceivedQty:  decimal.MustQuantity("0"),
			BaseUOMCode:      "PCS",
			ConversionFactor: decimal.MustQuantity("1"),
			UnitPrice:        decimal.MustUnitPrice("125000"),
			CurrencyCode:     "VND",
			ExpectedDate:     "2026-05-08",
			Note:             "first inbound lot",
		}},
		CreatedAt: time.Date(2026, 5, 1, 8, 0, 0, 0, time.UTC),
		CreatedBy: "user-purchase-ops",
	})
	if err != nil {
		t.Fatalf("new purchase order: %v", err)
	}
	log, err := audit.NewLog(audit.NewLogInput{
		OrgID:      order.OrgID,
		ActorID:    order.CreatedBy,
		Action:     "purchase.order.created",
		EntityType: purchaseOrderEntityType,
		EntityID:   order.ID,
		RequestID:  "req-s11-03-02-" + suffix,
		AfterData: map[string]any{
			"po_no":  order.PONo,
			"status": string(order.Status),
		},
		Metadata:  map[string]any{"source": "postgres_purchase_order_store_test"},
		CreatedAt: order.CreatedAt,
	})
	if err != nil {
		t.Fatalf("new audit log: %v", err)
	}

	if err := store.WithinTx(ctx, func(txCtx context.Context, tx PurchaseOrderTx) error {
		if err := tx.Save(txCtx, order); err != nil {
			return err
		}
		return tx.RecordAudit(txCtx, log)
	}); err != nil {
		t.Fatalf("persist purchase order: %v", err)
	}

	loaded, err := store.Get(ctx, order.ID)
	if err != nil {
		t.Fatalf("get persisted purchase order: %v", err)
	}
	if loaded.ID != order.ID ||
		loaded.SupplierID != "sup-rm-bioactive" ||
		loaded.WarehouseID != "wh-hcm-rm" ||
		loaded.TotalAmount.String() != "250000.00" ||
		len(loaded.Lines) != 1 ||
		loaded.Lines[0].ID != "po-line-s11-03-02-"+suffix ||
		loaded.Lines[0].BaseOrderedQty.String() != "2.000000" {
		t.Fatalf("loaded purchase order = %+v, want persisted document with line refs", loaded)
	}

	auditStore := audit.NewPostgresLogStore(
		db,
		audit.PostgresLogStoreConfig{DefaultOrgID: testPurchaseOrderOrgID},
	)
	logs, err := auditStore.List(ctx, audit.Query{
		Action:     "purchase.order.created",
		EntityType: purchaseOrderEntityType,
		EntityID:   order.ID,
	})
	if err != nil {
		t.Fatalf("list purchase order audit logs: %v", err)
	}
	if len(logs) != 1 || logs[0].ActorID != "user-purchase-ops" {
		t.Fatalf("audit logs = %+v, want persisted purchase order audit", logs)
	}
}

func TestPostgresPurchaseOrderFindQueryPlacesWhereBeforeLimit(t *testing.T) {
	if !strings.Contains(findPurchaseOrderHeaderSQL, "FROM purchase.purchase_orders\nWHERE po_ref = $1") {
		t.Fatalf("find query does not place WHERE immediately after FROM:\n%s", findPurchaseOrderHeaderSQL)
	}
	if strings.Contains(findPurchaseOrderHeaderSQL, "ORDER BY expected_date DESC, po_no ASC\nWHERE") {
		t.Fatalf("find query places WHERE after ORDER BY:\n%s", findPurchaseOrderHeaderSQL)
	}
}

func TestBuildPostgresPurchaseOrderMapsHeaderAndLines(t *testing.T) {
	createdAt := time.Date(2026, 5, 1, 8, 0, 0, 0, time.UTC)
	approvedAt := createdAt.Add(15 * time.Minute)
	header := postgresPurchaseOrderHeader{
		PersistedID:    "00000000-0000-4000-8000-000000000911",
		ID:             "po-s11-test-001",
		OrgID:          "org-my-pham",
		PONo:           "PO-S11-TEST-001",
		SupplierID:     "sup-herbal-lab",
		SupplierCode:   "SUP-HERBAL-LAB",
		SupplierName:   "Herbal Lab Supplier",
		WarehouseID:    "wh-hcm-fg",
		WarehouseCode:  "WH-HCM-FG",
		ExpectedDate:   "2026-05-06",
		Status:         "approved",
		CurrencyCode:   "VND",
		SubtotalAmount: "250000.00",
		TotalAmount:    "250000.00",
		CreatedAt:      createdAt,
		CreatedBy:      "user-purchase",
		UpdatedAt:      approvedAt,
		UpdatedBy:      "user-purchase",
		Version:        3,
		ApprovedAt:     sql.NullTime{Time: approvedAt, Valid: true},
		ApprovedBy:     "user-purchase",
	}
	line, err := scanPostgresPurchaseOrderLine(fakePurchaseOrderScanner{values: []any{
		"line-s11-test-001",
		1,
		"item-serum-30ml",
		"SERUM-30ML",
		"Hydrating Serum 30ml",
		"2.000000",
		"0.000000",
		"EA",
		"2.000000",
		"0.000000",
		"EA",
		"1.000000",
		"125000.0000",
		"VND",
		"250000.00",
		"2026-05-06",
		"first inbound lot",
	}})
	if err != nil {
		t.Fatalf("scanPostgresPurchaseOrderLine() error = %v", err)
	}

	order, err := buildPostgresPurchaseOrder(header, []purchasedomain.PurchaseOrderLine{line})
	if err != nil {
		t.Fatalf("buildPostgresPurchaseOrder() error = %v", err)
	}

	if order.ID != "po-s11-test-001" ||
		order.Status != "approved" ||
		order.Version != 3 ||
		order.TotalAmount.String() != "250000.00" ||
		len(order.Lines) != 1 ||
		order.Lines[0].Note != "first inbound lot" ||
		!order.ApprovedAt.Equal(approvedAt) ||
		order.ApprovedBy != "user-purchase" {
		t.Fatalf("order = %+v, want mapped purchase order", order)
	}
}

func TestPostgresPurchaseOrderStoreRequiresDatabase(t *testing.T) {
	store := NewPostgresPurchaseOrderStore(nil, PostgresPurchaseOrderStoreConfig{})

	if _, err := store.List(nil, PurchaseOrderFilter{}); err == nil {
		t.Fatal("List() error = nil, want database required error")
	}
	if _, err := store.Get(nil, "po-missing"); err == nil {
		t.Fatal("Get() error = nil, want database required error")
	}
	if err := store.WithinTx(nil, func(context.Context, PurchaseOrderTx) error { return nil }); err == nil {
		t.Fatal("WithinTx() error = nil, want database required error")
	}
}

type fakePurchaseOrderScanner struct {
	values []any
}

func (s fakePurchaseOrderScanner) Scan(dest ...any) error {
	for i, target := range dest {
		switch typed := target.(type) {
		case *string:
			*typed = s.values[i].(string)
		case *int:
			*typed = s.values[i].(int)
		case *time.Time:
			*typed = s.values[i].(time.Time)
		case *sql.NullTime:
			*typed = s.values[i].(sql.NullTime)
		default:
			panic("unsupported scan destination")
		}
	}

	return nil
}

func seedPurchaseOrderSmokeOrg(ctx context.Context, db *sql.DB) error {
	_, err := db.ExecContext(
		ctx,
		`INSERT INTO core.organizations (id, code, name, status)
VALUES ($1, 'ERP_LOCAL', 'ERP Local Demo', 'active')
ON CONFLICT (code) DO UPDATE
SET name = EXCLUDED.name,
    status = EXCLUDED.status,
    updated_at = now()`,
		testPurchaseOrderOrgID,
	)

	return err
}
