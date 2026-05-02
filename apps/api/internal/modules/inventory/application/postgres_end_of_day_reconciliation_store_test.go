package application

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/Chinsusu/ERP-v2/apps/api/internal/modules/inventory/domain"
	"github.com/Chinsusu/ERP-v2/apps/api/internal/shared/audit"

	_ "github.com/jackc/pgx/v5/stdlib"
)

const (
	testEndOfDayReconciliationOrgID       = "00000000-0000-4000-8000-000000130104"
	testEndOfDayReconciliationWarehouseID = "00000000-0000-4000-8000-000000130105"
	testEndOfDayReconciliationWarehouse   = "WH-S13-EOD"
)

func TestPostgresEndOfDayReconciliationStoreRequiresDatabase(t *testing.T) {
	store := NewPostgresEndOfDayReconciliationStore(nil, PostgresEndOfDayReconciliationStoreConfig{})

	if _, err := store.List(context.Background(), domain.EndOfDayReconciliationFilter{}); err == nil {
		t.Fatal("List() error = nil, want database required error")
	}
	if _, err := store.Get(context.Background(), "rec-s13"); err == nil {
		t.Fatal("Get() error = nil, want database required error")
	}
	if err := store.Save(context.Background(), domain.EndOfDayReconciliation{ID: "rec-s13"}); err == nil {
		t.Fatal("Save() error = nil, want database required error")
	}
}

func TestPostgresEndOfDayReconciliationStorePersistsCloseEvidence(t *testing.T) {
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

	if err := seedEndOfDayReconciliationFixture(ctx, db); err != nil {
		t.Fatalf("seed end-of-day reconciliation fixture: %v", err)
	}

	suffix := fmt.Sprintf("%d", time.Now().UTC().UnixNano())
	reconciliationID := "rec-s13-0104-" + suffix
	record := domain.EndOfDayReconciliation{
		ID:            reconciliationID,
		WarehouseID:   testEndOfDayReconciliationWarehouseID,
		WarehouseCode: testEndOfDayReconciliationWarehouse,
		Date:          "2026-05-02",
		ShiftCode:     "day",
		Status:        domain.ReconciliationStatusInReview,
		Owner:         "user-warehouse-lead",
		Operations: domain.ReconciliationOperations{
			OrderCount:             21,
			HandoverOrderCount:     18,
			ReturnOrderCount:       2,
			StockMovementCount:     7,
			StockCountSessionCount: 1,
			PendingIssueCount:      0,
		},
		Checklist: []domain.ReconciliationChecklistItem{
			{Key: "shipments", Label: "Shipments reconciled", Complete: true, Blocking: true},
			{Key: "variance", Label: "Stock variance reviewed", Complete: false, Blocking: true, Note: "accepted by lead"},
		},
		Lines: []domain.ReconciliationLine{{
			ID:              "line-s13-0104-" + suffix,
			SKU:             "SERUM-30ML",
			BatchNo:         "LOT-S13-0104",
			BinCode:         "PICK-A-01",
			SystemQuantity:  20,
			CountedQuantity: 18,
			Reason:          "cycle count variance",
			Owner:           "Warehouse Lead",
		}},
	}

	store := NewPostgresEndOfDayReconciliationStore(
		db,
		PostgresEndOfDayReconciliationStoreConfig{DefaultOrgID: testEndOfDayReconciliationOrgID},
	)
	if err := store.Save(ctx, record); err != nil {
		t.Fatalf("save initial reconciliation: %v", err)
	}

	auditStore := audit.NewPostgresLogStore(
		db,
		audit.PostgresLogStoreConfig{DefaultOrgID: testEndOfDayReconciliationOrgID},
	)
	closeUsecase := NewCloseEndOfDayReconciliation(store, auditStore)
	closedAt := time.Date(2026, 5, 2, 17, 30, 0, 0, time.UTC)
	closeUsecase.clock = func() time.Time { return closedAt }

	result, err := closeUsecase.Execute(ctx, CloseEndOfDayReconciliationInput{
		ID:            reconciliationID,
		ActorID:       "user-warehouse-lead",
		RequestID:     "req-s13-0104-" + suffix,
		ExceptionNote: "Variance accepted by warehouse lead",
	})
	if err != nil {
		t.Fatalf("close reconciliation: %v", err)
	}
	if result.Reconciliation.Status != domain.ReconciliationStatusClosed ||
		result.Reconciliation.ClosedBy != "user-warehouse-lead" ||
		result.Reconciliation.ExceptionNote != "Variance accepted by warehouse lead" ||
		result.AuditLogID == "" {
		t.Fatalf("close result = %+v, audit %q", result.Reconciliation, result.AuditLogID)
	}

	reloadedStore := NewPostgresEndOfDayReconciliationStore(
		db,
		PostgresEndOfDayReconciliationStoreConfig{DefaultOrgID: testEndOfDayReconciliationOrgID},
	)
	reloaded, err := reloadedStore.Get(ctx, reconciliationID)
	if err != nil {
		t.Fatalf("reload reconciliation: %v", err)
	}
	if reloaded.Status != domain.ReconciliationStatusClosed ||
		reloaded.ClosedBy != "user-warehouse-lead" ||
		!reloaded.ClosedAt.Equal(closedAt) ||
		reloaded.ExceptionNote != "Variance accepted by warehouse lead" ||
		len(reloaded.Checklist) != 2 ||
		len(reloaded.Lines) != 1 ||
		reloaded.Lines[0].VarianceQuantity() != -2 {
		t.Fatalf("reloaded reconciliation = %+v, want persisted close evidence", reloaded)
	}

	rows, err := reloadedStore.List(ctx, domain.NewEndOfDayReconciliationFilter(
		testEndOfDayReconciliationWarehouseID,
		"2026-05-02",
		"day",
		domain.ReconciliationStatusClosed,
	))
	if err != nil {
		t.Fatalf("list reconciliations: %v", err)
	}
	if !containsEndOfDayReconciliation(rows, reconciliationID) {
		t.Fatalf("filtered list missing %s: %+v", reconciliationID, rows)
	}

	logs, err := auditStore.List(ctx, audit.Query{
		Action:     "warehouse.shift.closed",
		EntityType: "inventory.warehouse_daily_closing",
		EntityID:   reconciliationID,
		Limit:      10,
	})
	if err != nil {
		t.Fatalf("list close audit logs: %v", err)
	}
	if len(logs) == 0 {
		t.Fatal("missing warehouse.shift.closed audit log")
	}

	var persistedStatus string
	var persistedExceptionNote string
	var persistedClosedBy string
	if err := db.QueryRowContext(
		ctx,
		`SELECT status, COALESCE(exception_note, ''), COALESCE(closed_by_ref, '')
FROM inventory.warehouse_daily_closings
WHERE closing_ref = $1`,
		reconciliationID,
	).Scan(&persistedStatus, &persistedExceptionNote, &persistedClosedBy); err != nil {
		t.Fatalf("query persisted closing: %v", err)
	}
	if persistedStatus != "closed" ||
		persistedExceptionNote != "Variance accepted by warehouse lead" ||
		persistedClosedBy != "user-warehouse-lead" {
		t.Fatalf(
			"persisted closing = status %q exception %q closed_by %q",
			persistedStatus,
			persistedExceptionNote,
			persistedClosedBy,
		)
	}
}

func TestPostgresEndOfDayReconciliationCloseBlocksUnresolvedIssue(t *testing.T) {
	store := NewPrototypeEndOfDayReconciliationStore()
	usecase := NewCloseEndOfDayReconciliation(store, audit.NewInMemoryLogStore())

	_, err := usecase.Execute(context.Background(), CloseEndOfDayReconciliationInput{
		ID:            "rec-hcm-260426-day",
		ActorID:       "user-warehouse-lead",
		RequestID:     "req-s13-blocked",
		ExceptionNote: "accepted",
	})
	if !errors.Is(err, domain.ErrReconciliationUnresolvedIssue) {
		t.Fatalf("close err = %v, want ErrReconciliationUnresolvedIssue", err)
	}
}

func containsEndOfDayReconciliation(rows []domain.EndOfDayReconciliation, id string) bool {
	for _, row := range rows {
		if row.ID == id {
			return true
		}
	}

	return false
}

func seedEndOfDayReconciliationFixture(ctx context.Context, db *sql.DB) error {
	if _, err := db.ExecContext(
		ctx,
		`INSERT INTO core.organizations (id, code, name, status)
VALUES ($1, 'ERP_S13_EOD', 'ERP S13 EOD Test', 'active')
ON CONFLICT (code) DO UPDATE
SET name = EXCLUDED.name,
    status = EXCLUDED.status,
    updated_at = now()`,
		testEndOfDayReconciliationOrgID,
	); err != nil {
		return err
	}

	_, err := db.ExecContext(
		ctx,
		`INSERT INTO mdm.warehouses (id, org_id, code, name, status)
VALUES ($1, $2, $3, 'Sprint 13 EOD Warehouse', 'active')
ON CONFLICT (org_id, code) DO UPDATE
SET name = EXCLUDED.name,
    status = EXCLUDED.status,
    updated_at = now()`,
		testEndOfDayReconciliationWarehouseID,
		testEndOfDayReconciliationOrgID,
		testEndOfDayReconciliationWarehouse,
	)

	return err
}
