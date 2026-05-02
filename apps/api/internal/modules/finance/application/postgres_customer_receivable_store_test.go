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

const testCustomerReceivableOrgID = "00000000-0000-4000-8000-000000000001"

func TestPostgresCustomerReceivableStorePersistsCustomerReceivable(t *testing.T) {
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

	if err := seedCustomerReceivableSmokeOrg(ctx, db); err != nil {
		t.Fatalf("seed org: %v", err)
	}

	suffix := fmt.Sprintf("%d", time.Now().UTC().UnixNano())
	store := NewPostgresCustomerReceivableStore(
		db,
		PostgresCustomerReceivableStoreConfig{DefaultOrgID: testCustomerReceivableOrgID},
	)
	source, err := financedomain.NewSourceDocumentRef(
		financedomain.SourceDocumentTypeShipment,
		"shipment-s15-02-01-"+suffix,
		"SHP-S15-02-01-"+suffix,
	)
	if err != nil {
		t.Fatalf("new source document: %v", err)
	}
	receivable, err := financedomain.NewCustomerReceivable(financedomain.NewCustomerReceivableInput{
		ID:             "ar-s15-02-01-" + suffix,
		OrgID:          defaultFinanceOrgID,
		ReceivableNo:   "AR-S15-02-01-" + suffix,
		CustomerID:     "customer-s15-02-01",
		CustomerCode:   "KH-S15",
		CustomerName:   "Sprint 15 Customer",
		SourceDocument: source,
		TotalAmount:    "1250000.00",
		CurrencyCode:   "VND",
		DueDate:        time.Date(2026, 5, 4, 0, 0, 0, 0, time.UTC),
		CreatedAt:      time.Date(2026, 5, 2, 9, 0, 0, 0, time.UTC),
		CreatedBy:      "finance-user",
		Lines: []financedomain.NewCustomerReceivableLineInput{{
			ID:             "ar-line-s15-02-01-" + suffix,
			Description:    "COD delivered goods",
			SourceDocument: source,
			Amount:         "1250000.00",
		}},
	})
	if err != nil {
		t.Fatalf("new customer receivable: %v", err)
	}

	if err := store.Save(ctx, receivable); err != nil {
		t.Fatalf("save receivable: %v", err)
	}
	loaded, err := store.Get(ctx, receivable.ID)
	if err != nil {
		t.Fatalf("get receivable: %v", err)
	}
	if loaded.ID != receivable.ID ||
		loaded.ReceivableNo != receivable.ReceivableNo ||
		loaded.TotalAmount.String() != "1250000.00" ||
		loaded.OutstandingAmount.String() != "1250000.00" ||
		len(loaded.Lines) != 1 ||
		loaded.Lines[0].ID != "ar-line-s15-02-01-"+suffix {
		t.Fatalf("loaded receivable = %+v, want persisted receivable with line refs", loaded)
	}

	updated, err := loaded.RecordReceipt("250000.00", "finance-user", time.Date(2026, 5, 2, 10, 0, 0, 0, time.UTC))
	if err != nil {
		t.Fatalf("record receipt: %v", err)
	}
	if err := store.Save(ctx, updated); err != nil {
		t.Fatalf("save receipt update: %v", err)
	}
	loaded, err = store.Get(ctx, receivable.ReceivableNo)
	if err != nil {
		t.Fatalf("get updated receivable by no: %v", err)
	}
	if loaded.Status != financedomain.ReceivableStatusPartiallyPaid ||
		loaded.PaidAmount.String() != "250000.00" ||
		loaded.OutstandingAmount.String() != "1000000.00" ||
		loaded.Version != 2 {
		t.Fatalf("loaded updated receivable = %+v, want receipt state persisted", loaded)
	}
}

func TestPostgresCustomerReceivableFindQueryPlacesWhereBeforeLimit(t *testing.T) {
	if !strings.Contains(findCustomerReceivableHeaderSQL, "FROM finance.customer_receivables AS receivable\nWHERE") {
		t.Fatalf("find query does not place WHERE immediately after FROM:\n%s", findCustomerReceivableHeaderSQL)
	}
	if strings.Contains(findCustomerReceivableHeaderSQL, "ORDER BY receivable.created_at DESC") {
		t.Fatalf("find query unexpectedly includes list ORDER BY:\n%s", findCustomerReceivableHeaderSQL)
	}
}

func TestScanPostgresCustomerReceivableLineMapsRow(t *testing.T) {
	line, err := scanPostgresCustomerReceivableLine(fakeCustomerReceivableScanner{values: []any{
		"ar-line-s15-test-001",
		"COD delivered goods",
		"shipment",
		"shipment-s15-test-001",
		"SHP-S15-TEST-001",
		"1250000.00",
	}})
	if err != nil {
		t.Fatalf("scanPostgresCustomerReceivableLine() error = %v", err)
	}

	if line.ID != "ar-line-s15-test-001" ||
		line.SourceDocument.Type != financedomain.SourceDocumentTypeShipment ||
		line.Amount.String() != "1250000.00" {
		t.Fatalf("line = %+v, want mapped customer receivable line", line)
	}
}

func TestPostgresCustomerReceivableStoreRequiresDatabase(t *testing.T) {
	store := NewPostgresCustomerReceivableStore(nil, PostgresCustomerReceivableStoreConfig{})

	if _, err := store.List(context.Background(), CustomerReceivableFilter{}); err == nil {
		t.Fatal("List() error = nil, want database required error")
	}
	if _, err := store.Get(context.Background(), "ar-missing"); err == nil {
		t.Fatal("Get() error = nil, want database required error")
	}
	if err := store.Save(context.Background(), financedomain.CustomerReceivable{ID: "ar-missing"}); err == nil {
		t.Fatal("Save() error = nil, want database required error")
	}
}

type fakeCustomerReceivableScanner struct {
	values []any
}

func (s fakeCustomerReceivableScanner) Scan(dest ...any) error {
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

func seedCustomerReceivableSmokeOrg(ctx context.Context, db *sql.DB) error {
	_, err := db.ExecContext(
		ctx,
		`INSERT INTO core.organizations (id, code, name, status)
VALUES ($1, 'ERP_LOCAL', 'ERP Local Demo', 'active')
ON CONFLICT (code) DO UPDATE
SET name = EXCLUDED.name,
    status = EXCLUDED.status,
    updated_at = now()`,
		testCustomerReceivableOrgID,
	)

	return err
}
