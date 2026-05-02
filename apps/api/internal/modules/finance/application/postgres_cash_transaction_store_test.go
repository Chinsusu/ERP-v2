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

const testCashTransactionOrgID = "00000000-0000-4000-8000-000000000001"

func TestPostgresCashTransactionStorePersistsCashTransaction(t *testing.T) {
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

	if err := seedCashTransactionSmokeOrg(ctx, db); err != nil {
		t.Fatalf("seed org: %v", err)
	}

	suffix := fmt.Sprintf("%d", time.Now().UTC().UnixNano())
	store := NewPostgresCashTransactionStore(
		db,
		PostgresCashTransactionStoreConfig{DefaultOrgID: testCashTransactionOrgID},
	)
	transaction, err := financedomain.NewCashTransaction(financedomain.NewCashTransactionInput{
		ID:               "cash-s15-05-01-" + suffix,
		OrgID:            defaultFinanceOrgID,
		TransactionNo:    "CASH-IN-S15-05-01-" + suffix,
		Direction:        financedomain.CashTransactionDirectionIn,
		BusinessDate:     time.Date(2026, 5, 2, 0, 0, 0, 0, time.UTC),
		CounterpartyID:   "carrier-s15-05-01",
		CounterpartyName: "GHN COD",
		PaymentMethod:    "bank_transfer",
		ReferenceNo:      "BANK-COD-S15-05-01-" + suffix,
		TotalAmount:      "1250000.00",
		CurrencyCode:     "VND",
		Memo:             "COD receipt smoke",
		CreatedAt:        time.Date(2026, 5, 2, 12, 0, 0, 0, time.UTC),
		CreatedBy:        "finance-user",
		Allocations: []financedomain.NewCashTransactionAllocationInput{
			{
				ID:         "cash-alloc-ar-s15-05-01-" + suffix,
				TargetType: financedomain.CashAllocationTargetCustomerReceivable,
				TargetID:   "ar-cod-s15-05-01-" + suffix,
				TargetNo:   "AR-COD-S15-05-01-" + suffix,
				Amount:     "1000000.00",
			},
			{
				ID:         "cash-alloc-cod-s15-05-01-" + suffix,
				TargetType: financedomain.CashAllocationTargetCODRemittance,
				TargetID:   "cod-remit-s15-05-01-" + suffix,
				TargetNo:   "COD-S15-05-01-" + suffix,
				Amount:     "250000.00",
			},
		},
	})
	if err != nil {
		t.Fatalf("new cash transaction: %v", err)
	}
	transaction, err = transaction.Post("finance-user", time.Date(2026, 5, 2, 12, 5, 0, 0, time.UTC))
	if err != nil {
		t.Fatalf("post cash transaction: %v", err)
	}

	if err := store.Save(ctx, transaction); err != nil {
		t.Fatalf("save transaction: %v", err)
	}
	loaded, err := store.Get(ctx, transaction.ID)
	if err != nil {
		t.Fatalf("get transaction: %v", err)
	}
	if loaded.ID != transaction.ID ||
		loaded.Status != financedomain.CashTransactionStatusPosted ||
		loaded.Direction != financedomain.CashTransactionDirectionIn ||
		loaded.TotalAmount.String() != "1250000.00" ||
		len(loaded.Allocations) != 2 ||
		loaded.Allocations[0].ID != "cash-alloc-ar-s15-05-01-"+suffix {
		t.Fatalf("loaded transaction = %+v, want persisted posted cash transaction", loaded)
	}

	voided, err := loaded.Void("finance-lead", "duplicate cash receipt", time.Date(2026, 5, 2, 12, 15, 0, 0, time.UTC))
	if err != nil {
		t.Fatalf("void cash transaction: %v", err)
	}
	if err := store.Save(ctx, voided); err != nil {
		t.Fatalf("save voided transaction: %v", err)
	}
	loaded, err = store.Get(ctx, transaction.TransactionNo)
	if err != nil {
		t.Fatalf("get voided transaction by no: %v", err)
	}
	if loaded.Status != financedomain.CashTransactionStatusVoid ||
		loaded.PostedBy != "finance-user" ||
		loaded.VoidReason != "duplicate cash receipt" ||
		loaded.VoidedBy != "finance-lead" ||
		loaded.Version != 3 ||
		len(loaded.Allocations) != 2 {
		t.Fatalf("loaded voided transaction = %+v, want persisted void state and allocations", loaded)
	}
}

func TestPostgresCashTransactionFindQueryPlacesWhereBeforeLimit(t *testing.T) {
	if !strings.Contains(findCashTransactionHeaderSQL, "FROM finance.cash_transactions AS transaction\nWHERE") {
		t.Fatalf("find query does not place WHERE immediately after FROM:\n%s", findCashTransactionHeaderSQL)
	}
	if strings.Contains(findCashTransactionHeaderSQL, "ORDER BY transaction.business_date DESC") {
		t.Fatalf("find query unexpectedly includes list ORDER BY:\n%s", findCashTransactionHeaderSQL)
	}
}

func TestScanPostgresCashTransactionAllocationMapsRow(t *testing.T) {
	allocation, err := scanPostgresCashTransactionAllocation(fakeCashTransactionScanner{values: []any{
		"cash-alloc-s15-test-001",
		"customer_receivable",
		"ar-cod-s15-test-001",
		"AR-COD-S15-TEST-001",
		"1250000.00",
	}}, financedomain.CashTransactionDirectionIn)
	if err != nil {
		t.Fatalf("scanPostgresCashTransactionAllocation() error = %v", err)
	}

	if allocation.ID != "cash-alloc-s15-test-001" ||
		allocation.TargetType != financedomain.CashAllocationTargetCustomerReceivable ||
		allocation.Amount.String() != "1250000.00" {
		t.Fatalf("allocation = %+v, want mapped cash allocation", allocation)
	}
}

func TestPostgresCashTransactionStoreRequiresDatabase(t *testing.T) {
	store := NewPostgresCashTransactionStore(nil, PostgresCashTransactionStoreConfig{})

	if _, err := store.List(context.Background(), CashTransactionFilter{}); err == nil {
		t.Fatal("List() error = nil, want database required error")
	}
	if _, err := store.Get(context.Background(), "cash-missing"); err == nil {
		t.Fatal("Get() error = nil, want database required error")
	}
	if err := store.Save(context.Background(), financedomain.CashTransaction{ID: "cash-missing"}); err == nil {
		t.Fatal("Save() error = nil, want database required error")
	}
}

type fakeCashTransactionScanner struct {
	values []any
}

func (s fakeCashTransactionScanner) Scan(dest ...any) error {
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

func seedCashTransactionSmokeOrg(ctx context.Context, db *sql.DB) error {
	_, err := db.ExecContext(
		ctx,
		`INSERT INTO core.organizations (id, code, name, status)
VALUES ($1, 'ERP_LOCAL', 'ERP Local Demo', 'active')
ON CONFLICT (code) DO UPDATE
SET name = EXCLUDED.name,
    status = EXCLUDED.status,
    updated_at = now()`,
		testCashTransactionOrgID,
	)

	return err
}
