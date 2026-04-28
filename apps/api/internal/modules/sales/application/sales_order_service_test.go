package application

import (
	"context"
	"errors"
	"testing"
	"time"

	masterdataapp "github.com/Chinsusu/ERP-v2/apps/api/internal/modules/masterdata/application"
	salesdomain "github.com/Chinsusu/ERP-v2/apps/api/internal/modules/sales/domain"
	"github.com/Chinsusu/ERP-v2/apps/api/internal/shared/audit"
	apperrors "github.com/Chinsusu/ERP-v2/apps/api/internal/shared/errors"
	"github.com/Chinsusu/ERP-v2/apps/api/internal/shared/response"
)

func TestCreateSalesOrderCalculatesAmountsAuditsAndUsesTransaction(t *testing.T) {
	service, store, auditStore := newTestSalesOrderService()

	result, err := service.CreateSalesOrder(context.Background(), CreateSalesOrderInput{
		ID:           "so-test-create",
		OrderNo:      "SO-TEST-CREATE",
		CustomerID:   "cus-dl-minh-anh",
		Channel:      "B2B",
		WarehouseID:  "wh-hcm-fg",
		OrderDate:    "2026-04-28",
		CurrencyCode: "VND",
		Lines: []SalesOrderLineInput{
			{
				ItemID:             "item-serum-30ml",
				OrderedQty:         "2",
				UOMCode:            "EA",
				UnitPrice:          "125000.5000",
				CurrencyCode:       "VND",
				LineDiscountAmount: "1000.00",
			},
		},
		ActorID:   "user-sales",
		RequestID: "req-sales-create",
	})
	if err != nil {
		t.Fatalf("create sales order: %v", err)
	}
	order := result.SalesOrder
	if order.Status != salesdomain.SalesOrderStatusDraft || order.Version != 1 {
		t.Fatalf("order status/version = %s/%d, want draft/1", order.Status, order.Version)
	}
	if order.CustomerCode != "CUS-DL-MINHANH" || order.Lines[0].SKUCode != "SERUM-30ML" {
		t.Fatalf("order enrichment = %+v line %+v, want customer and item data", order, order.Lines[0])
	}
	if order.SubtotalAmount != "250001.00" || order.DiscountAmount != "1000.00" || order.TotalAmount != "249001.00" {
		t.Fatalf("amounts = subtotal %s discount %s total %s, want calculated amounts", order.SubtotalAmount, order.DiscountAmount, order.TotalAmount)
	}
	if result.AuditLogID == "" {
		t.Fatal("audit log id is empty")
	}
	if store.TransactionCount() != 1 {
		t.Fatalf("transaction count = %d, want 1", store.TransactionCount())
	}
	logs, err := auditStore.List(context.Background(), audit.Query{Action: "sales.order.created"})
	if err != nil {
		t.Fatalf("list audit logs: %v", err)
	}
	if len(logs) != 1 || logs[0].AfterData["total"] != "249001.00" {
		t.Fatalf("audit logs = %+v, want created total", logs)
	}
}

func TestUpdateSalesOrderReplacesDraftLinesAndChecksVersion(t *testing.T) {
	service, _, _ := newTestSalesOrderService()
	ctx := context.Background()
	created, err := service.CreateSalesOrder(ctx, validCreateSalesOrderInput("so-test-update"))
	if err != nil {
		t.Fatalf("create sales order: %v", err)
	}

	updated, err := service.UpdateSalesOrder(ctx, UpdateSalesOrderInput{
		ID:              created.SalesOrder.ID,
		ExpectedVersion: 1,
		Lines: []SalesOrderLineInput{
			{
				ItemID:     "item-cream-50g",
				OrderedQty: "3",
				UOMCode:    "EA",
				UnitPrice:  "95000",
			},
		},
		ActorID:   "user-sales",
		RequestID: "req-sales-update",
	})
	if err != nil {
		t.Fatalf("update sales order: %v", err)
	}
	if updated.SalesOrder.Version != 2 {
		t.Fatalf("version = %d, want 2", updated.SalesOrder.Version)
	}
	if len(updated.SalesOrder.Lines) != 1 || updated.SalesOrder.Lines[0].SKUCode != "CREAM-50G" {
		t.Fatalf("lines = %+v, want replacement cream line", updated.SalesOrder.Lines)
	}
	if updated.SalesOrder.TotalAmount != "285000.00" {
		t.Fatalf("total = %s, want 285000.00", updated.SalesOrder.TotalAmount)
	}

	_, err = service.UpdateSalesOrder(ctx, UpdateSalesOrderInput{
		ID:              created.SalesOrder.ID,
		ExpectedVersion: 1,
		ActorID:         "user-sales",
	})
	if !errors.Is(err, ErrSalesOrderVersionConflict) {
		t.Fatalf("err = %v, want version conflict", err)
	}
	var appErr apperrors.AppError
	if !errors.As(err, &appErr) || appErr.Code != ErrorCodeSalesOrderVersionConflict {
		t.Fatalf("app error = %+v, want version conflict code", appErr)
	}
}

func TestConfirmAndCancelSalesOrderUseStateMachineAndAudit(t *testing.T) {
	service, _, auditStore := newTestSalesOrderService()
	ctx := context.Background()
	created, err := service.CreateSalesOrder(ctx, validCreateSalesOrderInput("so-test-confirm"))
	if err != nil {
		t.Fatalf("create sales order: %v", err)
	}

	confirmed, err := service.ConfirmSalesOrder(ctx, SalesOrderActionInput{
		ID:              created.SalesOrder.ID,
		ExpectedVersion: 1,
		ActorID:         "user-sales",
		RequestID:       "req-sales-confirm",
	})
	if err != nil {
		t.Fatalf("confirm sales order: %v", err)
	}
	if confirmed.PreviousStatus != salesdomain.SalesOrderStatusDraft ||
		confirmed.CurrentStatus != salesdomain.SalesOrderStatusConfirmed ||
		confirmed.SalesOrder.Version != 2 {
		t.Fatalf("confirm result = %+v, want draft -> confirmed version 2", confirmed)
	}

	_, err = service.CancelSalesOrder(ctx, SalesOrderActionInput{
		ID:              created.SalesOrder.ID,
		ExpectedVersion: 2,
		Reason:          "customer changed order",
		ActorID:         "user-sales",
		RequestID:       "req-sales-cancel",
	})
	if err != nil {
		t.Fatalf("cancel sales order: %v", err)
	}
	logs, err := auditStore.List(ctx, audit.Query{EntityID: created.SalesOrder.ID})
	if err != nil {
		t.Fatalf("list audit logs: %v", err)
	}
	if len(logs) != 3 {
		t.Fatalf("audit log count = %d, want create/confirm/cancel", len(logs))
	}
}

func TestSalesOrderRejectsUpdateAfterConfirm(t *testing.T) {
	service, _, _ := newTestSalesOrderService()
	ctx := context.Background()
	created, err := service.CreateSalesOrder(ctx, validCreateSalesOrderInput("so-test-confirmed-update"))
	if err != nil {
		t.Fatalf("create sales order: %v", err)
	}
	if _, err := service.ConfirmSalesOrder(ctx, SalesOrderActionInput{
		ID:              created.SalesOrder.ID,
		ExpectedVersion: 1,
		ActorID:         "user-sales",
	}); err != nil {
		t.Fatalf("confirm sales order: %v", err)
	}

	_, err = service.UpdateSalesOrder(ctx, UpdateSalesOrderInput{
		ID:              created.SalesOrder.ID,
		ExpectedVersion: 2,
		ActorID:         "user-sales",
		Lines: []SalesOrderLineInput{
			{ItemID: "item-serum-30ml", OrderedQty: "1", UOMCode: "EA", UnitPrice: "125000"},
		},
	})
	if !errors.Is(err, salesdomain.ErrSalesOrderInvalidTransition) {
		t.Fatalf("err = %v, want invalid transition", err)
	}
	var appErr apperrors.AppError
	if !errors.As(err, &appErr) || appErr.Code != ErrorCodeSalesOrderInvalidState {
		t.Fatalf("app error = %+v, want invalid state code", appErr)
	}
}

func TestSalesOrderRejectsNonBaseUOMWithStandardCode(t *testing.T) {
	service, _, _ := newTestSalesOrderService()

	_, err := service.CreateSalesOrder(context.Background(), CreateSalesOrderInput{
		ID:           "so-test-uom",
		OrderNo:      "SO-TEST-UOM",
		CustomerID:   "cus-dl-minh-anh",
		Channel:      "B2B",
		WarehouseID:  "wh-hcm-fg",
		OrderDate:    "2026-04-28",
		CurrencyCode: "VND",
		Lines: []SalesOrderLineInput{
			{ItemID: "item-serum-30ml", OrderedQty: "1", UOMCode: "CARTON", UnitPrice: "125000"},
		},
		ActorID: "user-sales",
	})
	var appErr apperrors.AppError
	if !errors.As(err, &appErr) || appErr.Code != response.ErrorCodeUOMConversionNotFound {
		t.Fatalf("err = %v app = %+v, want uom conversion code", err, appErr)
	}
}

func TestSalesOrderTransactionRollsBackWhenAuditFails(t *testing.T) {
	auditErr := errors.New("audit failed")
	auditStore := failingAuditStore{err: auditErr}
	store := NewPrototypeSalesOrderStore(auditStore)
	service := newTestSalesOrderServiceWithStore(store, auditStore)

	_, err := service.CreateSalesOrder(context.Background(), validCreateSalesOrderInput("so-test-rollback"))
	if !errors.Is(err, auditErr) {
		t.Fatalf("err = %v, want audit failure", err)
	}
	_, err = store.Get(context.Background(), "so-test-rollback")
	if !errors.Is(err, ErrSalesOrderNotFound) {
		t.Fatalf("stored order err = %v, want not found after rollback", err)
	}
}

func newTestSalesOrderService() (SalesOrderService, *PrototypeSalesOrderStore, *audit.InMemoryLogStore) {
	auditStore := audit.NewInMemoryLogStore()
	store := NewPrototypeSalesOrderStore(auditStore)
	service := newTestSalesOrderServiceWithStore(store, auditStore)

	return service, store, auditStore
}

func newTestSalesOrderServiceWithStore(
	store *PrototypeSalesOrderStore,
	auditStore audit.LogStore,
) SalesOrderService {
	service := NewSalesOrderService(
		store,
		masterdataapp.NewPrototypePartyCatalog(auditStore),
		masterdataapp.NewPrototypeItemCatalog(auditStore),
		masterdataapp.NewPrototypeWarehouseLocationCatalog(auditStore),
	)
	service.clock = func() time.Time {
		return time.Date(2026, 4, 28, 12, 0, 0, 0, time.UTC)
	}

	return service
}

func validCreateSalesOrderInput(id string) CreateSalesOrderInput {
	return CreateSalesOrderInput{
		ID:           id,
		OrderNo:      "SO-" + id,
		CustomerID:   "cus-dl-minh-anh",
		Channel:      "B2B",
		WarehouseID:  "wh-hcm-fg",
		OrderDate:    "2026-04-28",
		CurrencyCode: "VND",
		Lines: []SalesOrderLineInput{
			{
				ItemID:     "item-serum-30ml",
				OrderedQty: "2",
				UOMCode:    "EA",
				UnitPrice:  "125000",
			},
		},
		ActorID:   "user-sales",
		RequestID: "req-" + id,
	}
}

type failingAuditStore struct {
	err error
}

func (s failingAuditStore) Record(context.Context, audit.Log) error {
	return s.err
}

func (s failingAuditStore) List(context.Context, audit.Query) ([]audit.Log, error) {
	return nil, s.err
}
