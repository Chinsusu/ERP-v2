package application

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	productiondomain "github.com/Chinsusu/ERP-v2/apps/api/internal/modules/production/domain"
	"github.com/Chinsusu/ERP-v2/apps/api/internal/shared/audit"
	"github.com/Chinsusu/ERP-v2/apps/api/internal/shared/decimal"

	_ "github.com/jackc/pgx/v5/stdlib"
)

const testSubcontractOrderOrgID = "00000000-0000-4000-8000-000000000001"

func TestPostgresSubcontractOrderStorePersistsOrderDocument(t *testing.T) {
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

	if err := seedSubcontractOrderSmokeOrg(ctx, db); err != nil {
		t.Fatalf("seed org: %v", err)
	}

	suffix := fmt.Sprintf("%d", time.Now().UTC().UnixNano())
	store := NewPostgresSubcontractOrderStore(
		db,
		PostgresSubcontractOrderStoreConfig{DefaultOrgID: testSubcontractOrderOrgID},
	)
	order := subcontractMaterialTransferTestOrder(t)
	order.ID = "sco-s16-02-01-" + suffix
	order.OrderNo = "SCO-S16-02-01-" + suffix
	order.MaterialLines[0].ID = "sco-line-s16-02-01-a-" + suffix
	order.MaterialLines[1].ID = "sco-line-s16-02-01-b-" + suffix
	if err := order.Validate(); err != nil {
		t.Fatalf("validate test order: %v", err)
	}
	log, err := audit.NewLog(audit.NewLogInput{
		OrgID:      order.OrgID,
		ActorID:    order.UpdatedBy,
		Action:     "subcontract.order.persisted",
		EntityType: subcontractOrderEntityType,
		EntityID:   order.ID,
		RequestID:  "req-s16-02-01-" + suffix,
		AfterData: map[string]any{
			"order_no": order.OrderNo,
			"status":   string(order.Status),
		},
		Metadata:  map[string]any{"source": "postgres_subcontract_order_store_test"},
		CreatedAt: order.UpdatedAt,
	})
	if err != nil {
		t.Fatalf("new audit log: %v", err)
	}

	if err := store.WithinTx(ctx, func(txCtx context.Context, tx SubcontractOrderTx) error {
		if err := tx.Save(txCtx, order); err != nil {
			return err
		}
		return tx.RecordAudit(txCtx, log)
	}); err != nil {
		t.Fatalf("persist subcontract order: %v", err)
	}

	loaded, err := store.Get(ctx, order.ID)
	if err != nil {
		t.Fatalf("get persisted subcontract order: %v", err)
	}
	if loaded.ID != order.ID ||
		loaded.OrderNo != order.OrderNo ||
		loaded.Status != productiondomain.SubcontractOrderStatusFactoryConfirmed ||
		loaded.FactoryID != "fac_001" ||
		loaded.FinishedSKUCode != "FG-SERUM-001" ||
		loaded.EstimatedCostAmount.String() != "2200000.00" ||
		len(loaded.MaterialLines) != 2 ||
		loaded.MaterialLines[0].ID != "sco-line-s16-02-01-a-"+suffix {
		t.Fatalf("loaded subcontract order = %+v, want persisted document with material lines", loaded)
	}
	orders, err := store.List(ctx, SubcontractOrderFilter{
		FactoryID: "fac_001",
		Statuses:  []productiondomain.SubcontractOrderStatus{productiondomain.SubcontractOrderStatusFactoryConfirmed},
	})
	if err != nil {
		t.Fatalf("list persisted subcontract orders: %v", err)
	}
	if len(orders) != 1 || orders[0].ID != order.ID {
		t.Fatalf("listed subcontract orders = %+v, want persisted order", orders)
	}

	depositAt := time.Date(2026, 5, 2, 9, 30, 0, 0, time.UTC)
	if err := store.WithinTx(ctx, func(txCtx context.Context, tx SubcontractOrderTx) error {
		current, err := tx.GetForUpdate(txCtx, order.OrderNo)
		if err != nil {
			return err
		}
		updated, err := current.RecordDeposit("finance-user", decimal.MustMoneyAmount("500000"), depositAt)
		if err != nil {
			return err
		}
		return tx.Save(txCtx, updated)
	}); err != nil {
		t.Fatalf("persist subcontract deposit transition: %v", err)
	}
	loaded, err = store.Get(ctx, order.OrderNo)
	if err != nil {
		t.Fatalf("get updated subcontract order: %v", err)
	}
	if loaded.Status != productiondomain.SubcontractOrderStatusDepositRecorded ||
		loaded.DepositAmount.String() != "500000.00" ||
		loaded.DepositRecordedBy != "finance-user" ||
		!loaded.DepositRecordedAt.Equal(depositAt) {
		t.Fatalf("loaded updated subcontract order = %+v, want deposit transition persisted", loaded)
	}

	auditStore := audit.NewPostgresLogStore(
		db,
		audit.PostgresLogStoreConfig{DefaultOrgID: testSubcontractOrderOrgID},
	)
	logs, err := auditStore.List(ctx, audit.Query{
		Action:     "subcontract.order.persisted",
		EntityType: subcontractOrderEntityType,
		EntityID:   order.ID,
	})
	if err != nil {
		t.Fatalf("list subcontract order audit logs: %v", err)
	}
	if len(logs) != 1 || logs[0].ActorID != order.UpdatedBy {
		t.Fatalf("audit logs = %+v, want persisted subcontract order audit", logs)
	}
}

func TestPostgresSubcontractOrderFindQueryPlacesWhereBeforeLimit(t *testing.T) {
	if !strings.Contains(findSubcontractOrderHeaderSQL, "FROM subcontract.subcontract_orders AS subcontract_order\nLEFT JOIN") {
		t.Fatalf("find query does not include subcontract order FROM:\n%s", findSubcontractOrderHeaderSQL)
	}
	if strings.Contains(findSubcontractOrderHeaderSQL, "ORDER BY subcontract_order.expected_receipt_date DESC") {
		t.Fatalf("find query unexpectedly includes list ORDER BY:\n%s", findSubcontractOrderHeaderSQL)
	}
	if strings.Contains(findSubcontractOrderHeaderSQL, "LIMIT 1\nWHERE") {
		t.Fatalf("find query places WHERE after LIMIT:\n%s", findSubcontractOrderHeaderSQL)
	}
}

func TestBuildPostgresSubcontractOrderMapsHeaderAndLines(t *testing.T) {
	createdAt := time.Date(2026, 5, 2, 8, 0, 0, 0, time.UTC)
	depositAt := createdAt.Add(30 * time.Minute)
	line, err := scanPostgresSubcontractOrderLine(fakeSubcontractOrderScanner{values: []any{
		"sco-line-s16-test-001",
		1,
		"item-base",
		"RM-BASE-001",
		"Serum Base",
		"10.000000",
		"2.000000",
		"KG",
		"10000.000000",
		"2000.000000",
		"G",
		"1000.000000",
		"150000.000000",
		"VND",
		"1500000.00",
		true,
		"first material line",
	}})
	if err != nil {
		t.Fatalf("scanPostgresSubcontractOrderLine() error = %v", err)
	}

	order, err := buildPostgresSubcontractOrder(postgresSubcontractOrderHeader{
		PersistedID:         "00000000-0000-4000-8000-000000000916",
		ID:                  "sco-s16-test-001",
		OrgID:               defaultSubcontractOrderOrgID,
		OrderNo:             "SCO-S16-TEST-001",
		FactoryID:           "fac-001",
		FactoryCode:         "FAC-HCM-01",
		FactoryName:         "HCM Cosmetics Factory",
		FinishedItemID:      "item-serum",
		FinishedSKUCode:     "FG-SERUM-001",
		FinishedItemName:    "Brightening Serum",
		PlannedQty:          "1000.000000",
		ReceivedQty:         "0.000000",
		AcceptedQty:         "0.000000",
		RejectedQty:         "0.000000",
		UOMCode:             "PCS",
		BasePlannedQty:      "1000.000000",
		BaseReceivedQty:     "0.000000",
		BaseAcceptedQty:     "0.000000",
		BaseRejectedQty:     "0.000000",
		BaseUOMCode:         "PCS",
		ConversionFactor:    "1.000000",
		CurrencyCode:        "VND",
		EstimatedCostAmount: "1500000.00",
		DepositAmount:       "500000.00",
		SpecSummary:         "30ml bottle, printed box",
		SampleRequired:      true,
		ClaimWindowDays:     7,
		TargetStartDate:     "2026-05-02",
		ExpectedReceiptDate: "2026-05-12",
		Status:              "deposit_recorded",
		CreatedAt:           createdAt,
		CreatedBy:           "subcontract-user",
		UpdatedAt:           depositAt,
		UpdatedBy:           "finance-user",
		Version:             4,
		DepositRecordedAt:   sql.NullTime{Time: depositAt, Valid: true},
		DepositRecordedBy:   "finance-user",
	}, []productiondomain.SubcontractMaterialLine{line})
	if err != nil {
		t.Fatalf("buildPostgresSubcontractOrder() error = %v", err)
	}
	if order.ID != "sco-s16-test-001" ||
		order.Status != productiondomain.SubcontractOrderStatusDepositRecorded ||
		order.Version != 4 ||
		order.DepositAmount.String() != "500000.00" ||
		order.MaterialLines[0].IssuedQty.String() != "2.000000" ||
		order.DepositRecordedBy != "finance-user" ||
		!order.DepositRecordedAt.Equal(depositAt) {
		t.Fatalf("order = %+v, want mapped subcontract order", order)
	}
}

func TestPostgresSubcontractOrderStoreRequiresDatabase(t *testing.T) {
	store := NewPostgresSubcontractOrderStore(nil, PostgresSubcontractOrderStoreConfig{})

	if _, err := store.List(context.Background(), SubcontractOrderFilter{}); err == nil {
		t.Fatal("List() error = nil, want database required error")
	}
	if _, err := store.Get(context.Background(), "sco-missing"); err == nil {
		t.Fatal("Get() error = nil, want database required error")
	}
	if err := store.WithinTx(context.Background(), func(context.Context, SubcontractOrderTx) error { return nil }); err == nil {
		t.Fatal("WithinTx() error = nil, want database required error")
	}
}

type fakeSubcontractOrderScanner struct {
	values []any
}

func (s fakeSubcontractOrderScanner) Scan(dest ...any) error {
	for i, target := range dest {
		switch typed := target.(type) {
		case *string:
			*typed = s.values[i].(string)
		case *int:
			*typed = s.values[i].(int)
		case *bool:
			*typed = s.values[i].(bool)
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

func seedSubcontractOrderSmokeOrg(ctx context.Context, db *sql.DB) error {
	_, err := db.ExecContext(
		ctx,
		`INSERT INTO core.organizations (id, code, name, status)
VALUES ($1, 'ERP_LOCAL', 'ERP Local Demo', 'active')
ON CONFLICT (code) DO UPDATE
SET name = EXCLUDED.name,
    status = EXCLUDED.status,
    updated_at = now()`,
		testSubcontractOrderOrgID,
	)

	return err
}
