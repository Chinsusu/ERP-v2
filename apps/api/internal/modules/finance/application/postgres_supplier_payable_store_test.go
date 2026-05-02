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

const testSupplierPayableOrgID = "00000000-0000-4000-8000-000000000001"

func TestPostgresSupplierPayableStorePersistsSupplierPayable(t *testing.T) {
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

	if err := seedSupplierPayableSmokeOrg(ctx, db); err != nil {
		t.Fatalf("seed org: %v", err)
	}

	suffix := fmt.Sprintf("%d", time.Now().UTC().UnixNano())
	store := NewPostgresSupplierPayableStore(
		db,
		PostgresSupplierPayableStoreConfig{DefaultOrgID: testSupplierPayableOrgID},
	)
	headerSource, err := financedomain.NewSourceDocumentRef(
		financedomain.SourceDocumentTypeQCInspection,
		"qc-s15-03-01-"+suffix,
		"QC-S15-03-01-"+suffix,
	)
	if err != nil {
		t.Fatalf("new header source: %v", err)
	}
	lineSource, err := financedomain.NewSourceDocumentRef(
		financedomain.SourceDocumentTypeWarehouseReceipt,
		"gr-s15-03-01-"+suffix,
		"GR-S15-03-01-"+suffix,
	)
	if err != nil {
		t.Fatalf("new line source: %v", err)
	}
	payable, err := financedomain.NewSupplierPayable(financedomain.NewSupplierPayableInput{
		ID:             "ap-s15-03-01-" + suffix,
		OrgID:          defaultFinanceOrgID,
		PayableNo:      "AP-S15-03-01-" + suffix,
		SupplierID:     "supplier-s15-03-01",
		SupplierCode:   "SUP-S15",
		SupplierName:   "Sprint 15 Supplier",
		SourceDocument: headerSource,
		TotalAmount:    "4250000.00",
		CurrencyCode:   "VND",
		DueDate:        time.Date(2026, 5, 8, 0, 0, 0, 0, time.UTC),
		CreatedAt:      time.Date(2026, 5, 2, 10, 0, 0, 0, time.UTC),
		CreatedBy:      "finance-user",
		Lines: []financedomain.NewSupplierPayableLineInput{{
			ID:             "ap-line-s15-03-01-" + suffix,
			Description:    "Accepted raw material after inbound QC",
			SourceDocument: lineSource,
			Amount:         "4250000.00",
		}},
	})
	if err != nil {
		t.Fatalf("new supplier payable: %v", err)
	}

	if err := store.Save(ctx, payable); err != nil {
		t.Fatalf("save payable: %v", err)
	}
	loaded, err := store.Get(ctx, payable.ID)
	if err != nil {
		t.Fatalf("get payable: %v", err)
	}
	if loaded.ID != payable.ID ||
		loaded.PayableNo != payable.PayableNo ||
		loaded.TotalAmount.String() != "4250000.00" ||
		loaded.OutstandingAmount.String() != "4250000.00" ||
		len(loaded.Lines) != 1 ||
		loaded.Lines[0].ID != "ap-line-s15-03-01-"+suffix {
		t.Fatalf("loaded payable = %+v, want persisted payable with line refs", loaded)
	}

	requested, err := loaded.RequestPayment("finance-user", time.Date(2026, 5, 2, 10, 10, 0, 0, time.UTC))
	if err != nil {
		t.Fatalf("request payment: %v", err)
	}
	if err := store.Save(ctx, requested); err != nil {
		t.Fatalf("save requested payable: %v", err)
	}
	approved, err := requested.ApprovePayment("finance-lead", time.Date(2026, 5, 2, 10, 20, 0, 0, time.UTC))
	if err != nil {
		t.Fatalf("approve payment: %v", err)
	}
	if err := store.Save(ctx, approved); err != nil {
		t.Fatalf("save approved payable: %v", err)
	}
	paid, err := approved.RecordPayment("1250000.00", "cashier", time.Date(2026, 5, 2, 10, 30, 0, 0, time.UTC))
	if err != nil {
		t.Fatalf("record payment: %v", err)
	}
	if err := store.Save(ctx, paid); err != nil {
		t.Fatalf("save payment update: %v", err)
	}

	loaded, err = store.Get(ctx, payable.PayableNo)
	if err != nil {
		t.Fatalf("get updated payable by no: %v", err)
	}
	if loaded.Status != financedomain.PayableStatusPartiallyPaid ||
		loaded.PaymentRequestedBy != "finance-user" ||
		loaded.PaymentApprovedBy != "finance-lead" ||
		loaded.LastPaymentBy != "cashier" ||
		loaded.PaidAmount.String() != "1250000.00" ||
		loaded.OutstandingAmount.String() != "3000000.00" ||
		loaded.Version != 4 {
		t.Fatalf("loaded updated payable = %+v, want payment state persisted", loaded)
	}
}

func TestPostgresSupplierPayableFindQueryPlacesWhereBeforeLimit(t *testing.T) {
	if !strings.Contains(findSupplierPayableHeaderSQL, "FROM finance.supplier_payables AS payable\nWHERE") {
		t.Fatalf("find query does not place WHERE immediately after FROM:\n%s", findSupplierPayableHeaderSQL)
	}
	if strings.Contains(findSupplierPayableHeaderSQL, "ORDER BY payable.created_at DESC") {
		t.Fatalf("find query unexpectedly includes list ORDER BY:\n%s", findSupplierPayableHeaderSQL)
	}
}

func TestScanPostgresSupplierPayableLineMapsRow(t *testing.T) {
	line, err := scanPostgresSupplierPayableLine(fakeSupplierPayableScanner{values: []any{
		"ap-line-s15-test-001",
		"Accepted raw material after inbound QC",
		"warehouse_receipt",
		"gr-s15-test-001",
		"GR-S15-TEST-001",
		"4250000.00",
	}})
	if err != nil {
		t.Fatalf("scanPostgresSupplierPayableLine() error = %v", err)
	}

	if line.ID != "ap-line-s15-test-001" ||
		line.SourceDocument.Type != financedomain.SourceDocumentTypeWarehouseReceipt ||
		line.Amount.String() != "4250000.00" {
		t.Fatalf("line = %+v, want mapped supplier payable line", line)
	}
}

func TestPostgresSupplierPayableStoreRequiresDatabase(t *testing.T) {
	store := NewPostgresSupplierPayableStore(nil, PostgresSupplierPayableStoreConfig{})

	if _, err := store.List(context.Background(), SupplierPayableFilter{}); err == nil {
		t.Fatal("List() error = nil, want database required error")
	}
	if _, err := store.Get(context.Background(), "ap-missing"); err == nil {
		t.Fatal("Get() error = nil, want database required error")
	}
	if err := store.Save(context.Background(), financedomain.SupplierPayable{ID: "ap-missing"}); err == nil {
		t.Fatal("Save() error = nil, want database required error")
	}
}

type fakeSupplierPayableScanner struct {
	values []any
}

func (s fakeSupplierPayableScanner) Scan(dest ...any) error {
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

func seedSupplierPayableSmokeOrg(ctx context.Context, db *sql.DB) error {
	_, err := db.ExecContext(
		ctx,
		`INSERT INTO core.organizations (id, code, name, status)
VALUES ($1, 'ERP_LOCAL', 'ERP Local Demo', 'active')
ON CONFLICT (code) DO UPDATE
SET name = EXCLUDED.name,
    status = EXCLUDED.status,
    updated_at = now()`,
		testSupplierPayableOrgID,
	)

	return err
}
