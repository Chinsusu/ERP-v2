package application

import (
	"context"
	"database/sql"
	"strings"
	"testing"
	"time"

	salesdomain "github.com/Chinsusu/ERP-v2/apps/api/internal/modules/sales/domain"
)

func TestPostgresSalesOrderFindQueryPlacesWhereBeforeLimit(t *testing.T) {
	if !strings.Contains(findSalesOrderHeaderSQL, "FROM sales.sales_orders\nWHERE order_ref = $1") {
		t.Fatalf("find query does not place WHERE immediately after FROM:\n%s", findSalesOrderHeaderSQL)
	}
	if strings.Contains(findSalesOrderHeaderSQL, "ORDER BY order_date DESC, order_no ASC\nWHERE") {
		t.Fatalf("find query places WHERE after ORDER BY:\n%s", findSalesOrderHeaderSQL)
	}
}

func TestBuildPostgresSalesOrderMapsHeaderAndLines(t *testing.T) {
	createdAt := time.Date(2026, 5, 1, 8, 0, 0, 0, time.UTC)
	confirmedAt := createdAt.Add(10 * time.Minute)
	header := postgresSalesOrderHeader{
		PersistedID:       "00000000-0000-4000-8000-000000000901",
		ID:                "so-s11-test-001",
		OrgID:             "org-my-pham",
		OrderNo:           "SO-S11-TEST-001",
		CustomerID:        "cus-dl-minh-anh",
		CustomerCode:      "CUS-DL-MINHANH",
		CustomerName:      "Minh Anh Distributor",
		Channel:           "B2B",
		WarehouseID:       "wh-hcm-fg",
		WarehouseCode:     "WH-HCM-FG",
		OrderDate:         "2026-05-01",
		Status:            "confirmed",
		CurrencyCode:      "VND",
		SubtotalAmount:    "125000.00",
		DiscountAmount:    "0.00",
		TaxAmount:         "0.00",
		ShippingFeeAmount: "0.00",
		NetAmount:         "125000.00",
		TotalAmount:       "125000.00",
		CreatedAt:         createdAt,
		CreatedBy:         "user-sales",
		UpdatedAt:         confirmedAt,
		UpdatedBy:         "user-sales",
		Version:           2,
		ConfirmedAt:       sql.NullTime{Time: confirmedAt, Valid: true},
		ConfirmedBy:       "user-sales",
	}
	line, err := scanPostgresSalesOrderLine(fakeSalesOrderScanner{values: []any{
		"line-s11-test-001",
		1,
		"item-serum-30ml",
		"SERUM-30ML",
		"Hydrating Serum 30ml",
		"1.000000",
		"EA",
		"1.000000",
		"EA",
		"1.000000",
		"125000.0000",
		"VND",
		"0.00",
		"125000.00",
		"0.000000",
		"0.000000",
		"",
		"",
	}})
	if err != nil {
		t.Fatalf("scanPostgresSalesOrderLine() error = %v", err)
	}

	order, err := buildPostgresSalesOrder(header, []salesdomain.SalesOrderLine{line})
	if err != nil {
		t.Fatalf("buildPostgresSalesOrder() error = %v", err)
	}

	if order.ID != "so-s11-test-001" ||
		order.Status != "confirmed" ||
		order.Version != 2 ||
		order.TotalAmount.String() != "125000.00" ||
		len(order.Lines) != 1 ||
		order.Lines[0].ItemID != "item-serum-30ml" ||
		!order.ConfirmedAt.Equal(confirmedAt) ||
		order.ConfirmedBy != "user-sales" {
		t.Fatalf("order = %+v, want mapped sales order", order)
	}
}

func TestPostgresSalesOrderStoreRequiresDatabase(t *testing.T) {
	store := NewPostgresSalesOrderStore(nil, PostgresSalesOrderStoreConfig{})

	if _, err := store.List(nil, SalesOrderFilter{}); err == nil {
		t.Fatal("List() error = nil, want database required error")
	}
	if _, err := store.Get(nil, "so-missing"); err == nil {
		t.Fatal("Get() error = nil, want database required error")
	}
	if err := store.WithinTx(nil, func(context.Context, SalesOrderTx) error { return nil }); err == nil {
		t.Fatal("WithinTx() error = nil, want database required error")
	}
}

type fakeSalesOrderScanner struct {
	values []any
}

func (s fakeSalesOrderScanner) Scan(dest ...any) error {
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
