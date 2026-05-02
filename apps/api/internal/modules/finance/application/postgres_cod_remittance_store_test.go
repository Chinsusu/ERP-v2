package application

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	financedomain "github.com/Chinsusu/ERP-v2/apps/api/internal/modules/finance/domain"

	_ "github.com/jackc/pgx/v5/stdlib"
)

const testCODRemittanceOrgID = "00000000-0000-4000-8000-000000000001"

func TestPostgresCODRemittanceStorePersistsDiscrepancyLifecycle(t *testing.T) {
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

	if err := seedCODRemittanceSmokeOrg(ctx, db); err != nil {
		t.Fatalf("seed org: %v", err)
	}

	suffix := fmt.Sprintf("%d", time.Now().UTC().UnixNano())
	store := NewPostgresCODRemittanceStore(
		db,
		PostgresCODRemittanceStoreConfig{DefaultOrgID: testCODRemittanceOrgID},
	)
	lineID := "cod-line-s15-04-01-" + suffix
	remittance, err := financedomain.NewCODRemittance(financedomain.NewCODRemittanceInput{
		ID:             "cod-remit-s15-04-01-" + suffix,
		OrgID:          defaultFinanceOrgID,
		RemittanceNo:   "COD-S15-04-01-" + suffix,
		CarrierID:      "carrier-s15-04-01",
		CarrierCode:    "GHN",
		CarrierName:    "GHN Express",
		BusinessDate:   time.Date(2026, 5, 2, 0, 0, 0, 0, time.UTC),
		ExpectedAmount: "1250000.00",
		RemittedAmount: "1200000.00",
		CurrencyCode:   "VND",
		CreatedAt:      time.Date(2026, 5, 2, 11, 0, 0, 0, time.UTC),
		CreatedBy:      "finance-user",
		Lines: []financedomain.NewCODRemittanceLineInput{{
			ID:             lineID,
			ReceivableID:   "ar-cod-s15-04-01-" + suffix,
			ReceivableNo:   "AR-COD-S15-04-01-" + suffix,
			ShipmentID:     "shipment-s15-04-01-" + suffix,
			TrackingNo:     "GHN-S15-04-01-" + suffix,
			CustomerName:   "Sprint 15 COD Customer",
			ExpectedAmount: "1250000.00",
			RemittedAmount: "1200000.00",
		}},
	})
	if err != nil {
		t.Fatalf("new cod remittance: %v", err)
	}

	if err := store.Save(ctx, remittance); err != nil {
		t.Fatalf("save remittance: %v", err)
	}
	loaded, err := store.Get(ctx, remittance.ID)
	if err != nil {
		t.Fatalf("get remittance: %v", err)
	}
	if loaded.ID != remittance.ID ||
		loaded.DiscrepancyAmount.String() != "-50000.00" ||
		len(loaded.Lines) != 1 ||
		loaded.Lines[0].ID != lineID ||
		loaded.Lines[0].MatchStatus != financedomain.CODLineMatchStatusShortPaid {
		t.Fatalf("loaded remittance = %+v, want persisted short-paid line", loaded)
	}

	discrepant, err := loaded.RecordDiscrepancy(financedomain.RecordCODDiscrepancyInput{
		ID:         "cod-disc-s15-04-01-" + suffix,
		LineID:     lineID,
		Reason:     "carrier remitted short",
		OwnerID:    "finance-user",
		RecordedBy: "finance-user",
		RecordedAt: time.Date(2026, 5, 2, 11, 10, 0, 0, time.UTC),
	})
	if err != nil {
		t.Fatalf("record discrepancy: %v", err)
	}
	if err := store.Save(ctx, discrepant); err != nil {
		t.Fatalf("save discrepancy update: %v", err)
	}
	submitted, err := discrepant.Submit("finance-lead", time.Date(2026, 5, 2, 11, 20, 0, 0, time.UTC))
	if err != nil {
		t.Fatalf("submit remittance: %v", err)
	}
	if err := store.Save(ctx, submitted); err != nil {
		t.Fatalf("save submitted update: %v", err)
	}
	approved, err := submitted.Approve("finance-lead", time.Date(2026, 5, 2, 11, 30, 0, 0, time.UTC))
	if err != nil {
		t.Fatalf("approve remittance: %v", err)
	}
	if err := store.Save(ctx, approved); err != nil {
		t.Fatalf("save approved update: %v", err)
	}
	closed, err := approved.Close("finance-lead", time.Date(2026, 5, 2, 11, 40, 0, 0, time.UTC))
	if err != nil {
		t.Fatalf("close remittance: %v", err)
	}
	if err := store.Save(ctx, closed); err != nil {
		t.Fatalf("save closed update: %v", err)
	}

	loaded, err = store.Get(ctx, remittance.RemittanceNo)
	if err != nil {
		t.Fatalf("get closed remittance by no: %v", err)
	}
	if loaded.Status != financedomain.CODRemittanceStatusClosed ||
		loaded.SubmittedBy != "finance-lead" ||
		loaded.ApprovedBy != "finance-lead" ||
		loaded.ClosedBy != "finance-lead" ||
		loaded.Version != 5 ||
		len(loaded.Discrepancies) != 1 ||
		loaded.Discrepancies[0].ID != "cod-disc-s15-04-01-"+suffix ||
		loaded.Discrepancies[0].Amount.String() != "-50000.00" {
		t.Fatalf("loaded closed remittance = %+v, want persisted discrepancy lifecycle", loaded)
	}
}

func TestPostgresCODRemittanceFindQueryPlacesWhereBeforeLimit(t *testing.T) {
	if !strings.Contains(findCODRemittanceHeaderSQL, "FROM finance.cod_remittances AS remittance\nWHERE") {
		t.Fatalf("find query does not place WHERE immediately after FROM:\n%s", findCODRemittanceHeaderSQL)
	}
	if strings.Contains(findCODRemittanceHeaderSQL, "ORDER BY remittance.business_date DESC") {
		t.Fatalf("find query unexpectedly includes list ORDER BY:\n%s", findCODRemittanceHeaderSQL)
	}
}

func TestScanPostgresCODRemittanceLineMapsRow(t *testing.T) {
	line, err := scanPostgresCODRemittanceLine(fakeCODRemittanceScanner{values: []any{
		"cod-line-s15-test-001",
		"ar-cod-s15-test-001",
		"AR-COD-S15-TEST-001",
		"shipment-s15-test-001",
		"GHN-S15-TEST-001",
		"COD Customer",
		"1250000.00",
		"1200000.00",
		"-50000.00",
		"short_paid",
	}})
	if err != nil {
		t.Fatalf("scanPostgresCODRemittanceLine() error = %v", err)
	}

	if line.ID != "cod-line-s15-test-001" ||
		line.MatchStatus != financedomain.CODLineMatchStatusShortPaid ||
		line.DiscrepancyAmount.String() != "-50000.00" {
		t.Fatalf("line = %+v, want mapped short-paid COD line", line)
	}
}

func TestPostgresCODRemittanceStoreRequiresDatabase(t *testing.T) {
	store := NewPostgresCODRemittanceStore(nil, PostgresCODRemittanceStoreConfig{})

	if _, err := store.List(context.Background(), CODRemittanceFilter{}); err == nil {
		t.Fatal("List() error = nil, want database required error")
	}
	if _, err := store.Get(context.Background(), "cod-missing"); err == nil {
		t.Fatal("Get() error = nil, want database required error")
	}
	if err := store.Save(context.Background(), financedomain.CODRemittance{ID: "cod-missing"}); err == nil {
		t.Fatal("Save() error = nil, want database required error")
	}
}

type fakeCODRemittanceScanner struct {
	values []any
}

func (s fakeCODRemittanceScanner) Scan(dest ...any) error {
	for i, target := range dest {
		switch typed := target.(type) {
		case *string:
			*typed = s.values[i].(string)
		default:
			panic("unsupported scan destination")
		}
	}

	return nil
}

func seedCODRemittanceSmokeOrg(ctx context.Context, db *sql.DB) error {
	_, err := db.ExecContext(
		ctx,
		`INSERT INTO core.organizations (id, code, name, status)
VALUES ($1, 'ERP_LOCAL', 'ERP Local Demo', 'active')
ON CONFLICT (code) DO UPDATE
SET name = EXCLUDED.name,
    status = EXCLUDED.status,
    updated_at = now()`,
		testCODRemittanceOrgID,
	)

	return err
}
