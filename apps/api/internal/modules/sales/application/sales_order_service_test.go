package application

import (
	"context"
	"errors"
	"testing"
	"time"

	masterdataapp "github.com/Chinsusu/ERP-v2/apps/api/internal/modules/masterdata/application"
	salesdomain "github.com/Chinsusu/ERP-v2/apps/api/internal/modules/sales/domain"
	"github.com/Chinsusu/ERP-v2/apps/api/internal/shared/audit"
	"github.com/Chinsusu/ERP-v2/apps/api/internal/shared/decimal"
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

func TestMarkSalesOrderHandedOverAdvancesWaitingOrder(t *testing.T) {
	service, store, auditStore := newTestSalesOrderService()
	ctx := context.Background()
	order, err := store.Get(ctx, "so-260426-003")
	if err != nil {
		t.Fatalf("get prototype handover order: %v", err)
	}
	if order.Status != salesdomain.SalesOrderStatusWaitingHandover {
		t.Fatalf("prototype status = %s, want waiting_handover", order.Status)
	}

	result, err := service.MarkSalesOrderHandedOver(ctx, SalesOrderActionInput{
		ID:        "so-260426-003",
		ActorID:   "user-handover-operator",
		RequestID: "req-sales-handover",
	})
	if err != nil {
		t.Fatalf("mark handed over: %v", err)
	}
	if result.SalesOrder.Status != salesdomain.SalesOrderStatusHandedOver {
		t.Fatalf("status = %s, want handed_over", result.SalesOrder.Status)
	}

	logs, err := auditStore.List(ctx, audit.Query{Action: "sales.order.handed_over"})
	if err != nil {
		t.Fatalf("list audit logs: %v", err)
	}
	if len(logs) != 1 || logs[0].AfterData["status"] != string(salesdomain.SalesOrderStatusHandedOver) {
		t.Fatalf("audit logs = %+v, want handed over status", logs)
	}
}

func TestConfirmSalesOrderReservesStockWhenConfigured(t *testing.T) {
	service, _, auditStore := newTestSalesOrderService()
	reserver := &recordingSalesOrderStockReserver{}
	service = service.WithStockReserver(reserver)
	ctx := context.Background()
	created, err := service.CreateSalesOrder(ctx, validCreateSalesOrderInput("so-test-reserve"))
	if err != nil {
		t.Fatalf("create sales order: %v", err)
	}

	reserved, err := service.ConfirmSalesOrder(ctx, SalesOrderActionInput{
		ID:              created.SalesOrder.ID,
		ExpectedVersion: 1,
		ActorID:         "user-sales",
		RequestID:       "req-sales-reserve",
	})
	if err != nil {
		t.Fatalf("confirm and reserve sales order: %v", err)
	}

	if reserved.PreviousStatus != salesdomain.SalesOrderStatusDraft ||
		reserved.CurrentStatus != salesdomain.SalesOrderStatusReserved ||
		reserved.SalesOrder.Version != 3 {
		t.Fatalf("reserve result = %+v, want draft -> reserved version 3", reserved)
	}
	if reserved.SalesOrder.Lines[0].ReservedQty != "2.000000" ||
		reserved.SalesOrder.Lines[0].BatchID != "batch-reserved" {
		t.Fatalf("reserved line = %+v, want full reserved qty and batch", reserved.SalesOrder.Lines[0])
	}
	if reserver.calls != 1 || reserver.inputs[0].Lines[0].BaseOrderedQty != "2.000000" {
		t.Fatalf("reserver calls = %d inputs = %+v, want one base qty reservation", reserver.calls, reserver.inputs)
	}
	logs, err := auditStore.List(ctx, audit.Query{EntityID: created.SalesOrder.ID})
	if err != nil {
		t.Fatalf("list audit logs: %v", err)
	}
	if len(logs) != 2 || logs[1].Action != "sales.order.reserved" || logs[1].AfterData["status"] != "reserved" {
		t.Fatalf("audit logs = %+v, want reserved audit", logs)
	}
}

func TestConfirmSalesOrderRollsBackWhenReservationFails(t *testing.T) {
	service, store, auditStore := newTestSalesOrderService()
	service = service.WithStockReserver(&recordingSalesOrderStockReserver{
		err: apperrors.Conflict(
			response.ErrorCodeInsufficientStock,
			"Insufficient stock for sales order reservation",
			errors.New("insufficient stock"),
			nil,
		),
	})
	ctx := context.Background()
	created, err := service.CreateSalesOrder(ctx, validCreateSalesOrderInput("so-test-reserve-fail"))
	if err != nil {
		t.Fatalf("create sales order: %v", err)
	}

	_, err = service.ConfirmSalesOrder(ctx, SalesOrderActionInput{
		ID:              created.SalesOrder.ID,
		ExpectedVersion: 1,
		ActorID:         "user-sales",
		RequestID:       "req-sales-reserve-fail",
	})
	var appErr apperrors.AppError
	if !errors.As(err, &appErr) || appErr.Code != response.ErrorCodeInsufficientStock {
		t.Fatalf("err = %v app = %+v, want insufficient stock code", err, appErr)
	}
	stored, err := store.Get(ctx, created.SalesOrder.ID)
	if err != nil {
		t.Fatalf("get stored order: %v", err)
	}
	if stored.Status != salesdomain.SalesOrderStatusDraft || stored.Version != 1 {
		t.Fatalf("stored order status/version = %s/%d, want rollback to draft/1", stored.Status, stored.Version)
	}
	logs, err := auditStore.List(ctx, audit.Query{EntityID: created.SalesOrder.ID})
	if err != nil {
		t.Fatalf("list audit logs: %v", err)
	}
	if len(logs) != 1 {
		t.Fatalf("audit log count = %d, want only create audit after reservation failure", len(logs))
	}
}

func TestCancelReservedSalesOrderReleasesStockWhenConfigured(t *testing.T) {
	service, _, auditStore := newTestSalesOrderService()
	reserver := &recordingSalesOrderStockReserver{}
	service = service.WithStockReserver(reserver)
	ctx := context.Background()
	created, err := service.CreateSalesOrder(ctx, validCreateSalesOrderInput("so-test-cancel-release"))
	if err != nil {
		t.Fatalf("create sales order: %v", err)
	}
	confirmed, err := service.ConfirmSalesOrder(ctx, SalesOrderActionInput{
		ID:              created.SalesOrder.ID,
		ExpectedVersion: 1,
		ActorID:         "user-sales",
	})
	if err != nil {
		t.Fatalf("confirm sales order: %v", err)
	}

	cancelled, err := service.CancelSalesOrder(ctx, SalesOrderActionInput{
		ID:              created.SalesOrder.ID,
		ExpectedVersion: confirmed.SalesOrder.Version,
		Reason:          "customer changed order",
		ActorID:         "user-sales",
		RequestID:       "req-sales-cancel-release",
	})
	if err != nil {
		t.Fatalf("cancel reserved sales order: %v", err)
	}

	if cancelled.PreviousStatus != salesdomain.SalesOrderStatusReserved ||
		cancelled.CurrentStatus != salesdomain.SalesOrderStatusCancelled ||
		cancelled.SalesOrder.Version != 4 {
		t.Fatalf("cancel result = %+v, want reserved -> cancelled version 4", cancelled)
	}
	if reserver.releaseCalls != 1 || reserver.releaseInputs[0].Reason != "customer changed order" {
		t.Fatalf("release calls = %d inputs = %+v, want one release with reason", reserver.releaseCalls, reserver.releaseInputs)
	}
	logs, err := auditStore.List(ctx, audit.Query{EntityID: created.SalesOrder.ID})
	if err != nil {
		t.Fatalf("list audit logs: %v", err)
	}
	if len(logs) != 3 || logs[2].Action != "sales.order.cancelled" {
		t.Fatalf("audit logs = %+v, want cancelled audit after release", logs)
	}
}

func TestCancelReservedSalesOrderRollsBackWhenReleaseFails(t *testing.T) {
	service, store, auditStore := newTestSalesOrderService()
	reserver := &recordingSalesOrderStockReserver{}
	service = service.WithStockReserver(reserver)
	ctx := context.Background()
	created, err := service.CreateSalesOrder(ctx, validCreateSalesOrderInput("so-test-cancel-release-fail"))
	if err != nil {
		t.Fatalf("create sales order: %v", err)
	}
	confirmed, err := service.ConfirmSalesOrder(ctx, SalesOrderActionInput{
		ID:              created.SalesOrder.ID,
		ExpectedVersion: 1,
		ActorID:         "user-sales",
	})
	if err != nil {
		t.Fatalf("confirm sales order: %v", err)
	}
	reserver.releaseErr = apperrors.Conflict(
		response.ErrorCodeConflict,
		"Reserved stock could not be released",
		errors.New("release failed"),
		nil,
	)

	_, err = service.CancelSalesOrder(ctx, SalesOrderActionInput{
		ID:              created.SalesOrder.ID,
		ExpectedVersion: confirmed.SalesOrder.Version,
		Reason:          "customer changed order",
		ActorID:         "user-sales",
		RequestID:       "req-sales-cancel-release-fail",
	})
	var appErr apperrors.AppError
	if !errors.As(err, &appErr) || appErr.Code != response.ErrorCodeConflict {
		t.Fatalf("err = %v app = %+v, want release conflict", err, appErr)
	}
	stored, err := store.Get(ctx, created.SalesOrder.ID)
	if err != nil {
		t.Fatalf("get stored order: %v", err)
	}
	if stored.Status != salesdomain.SalesOrderStatusReserved || stored.Version != 3 {
		t.Fatalf("stored order status/version = %s/%d, want rollback to reserved/3", stored.Status, stored.Version)
	}
	logs, err := auditStore.List(ctx, audit.Query{EntityID: created.SalesOrder.ID})
	if err != nil {
		t.Fatalf("list audit logs: %v", err)
	}
	if len(logs) != 2 {
		t.Fatalf("audit log count = %d, want create/reserved after release failure", len(logs))
	}
}

func TestCancelReservedSalesOrderRejectsMissingActiveReservation(t *testing.T) {
	service, store, auditStore := newTestSalesOrderService()
	service = service.WithStockReserver(&missingActiveReservationStockReserver{})
	ctx := context.Background()
	created, err := service.CreateSalesOrder(ctx, validCreateSalesOrderInput("so-test-cancel-missing-reservation"))
	if err != nil {
		t.Fatalf("create sales order: %v", err)
	}
	confirmed, err := service.ConfirmSalesOrder(ctx, SalesOrderActionInput{
		ID:              created.SalesOrder.ID,
		ExpectedVersion: 1,
		ActorID:         "user-sales",
	})
	if err != nil {
		t.Fatalf("confirm sales order: %v", err)
	}

	_, err = service.CancelSalesOrder(ctx, SalesOrderActionInput{
		ID:              created.SalesOrder.ID,
		ExpectedVersion: confirmed.SalesOrder.Version,
		Reason:          "customer changed order",
		ActorID:         "user-sales",
	})
	var appErr apperrors.AppError
	if !errors.As(err, &appErr) || appErr.Code != response.ErrorCodeConflict {
		t.Fatalf("err = %v app = %+v, want missing reservation conflict", err, appErr)
	}
	stored, err := store.Get(ctx, created.SalesOrder.ID)
	if err != nil {
		t.Fatalf("get stored order: %v", err)
	}
	if stored.Status != salesdomain.SalesOrderStatusReserved || stored.Version != 3 {
		t.Fatalf("stored order status/version = %s/%d, want reserved/3", stored.Status, stored.Version)
	}
	logs, err := auditStore.List(ctx, audit.Query{EntityID: created.SalesOrder.ID})
	if err != nil {
		t.Fatalf("list audit logs: %v", err)
	}
	if len(logs) != 2 {
		t.Fatalf("audit log count = %d, want no cancel audit", len(logs))
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

type recordingSalesOrderStockReserver struct {
	calls         int
	inputs        []SalesOrderStockReservationInput
	err           error
	releaseCalls  int
	releaseInputs []SalesOrderStockReleaseInput
	releaseErr    error
}

func (r *recordingSalesOrderStockReserver) ReserveSalesOrder(
	_ context.Context,
	input SalesOrderStockReservationInput,
) (SalesOrderStockReservationResult, error) {
	r.calls++
	r.inputs = append(r.inputs, input)
	if r.err != nil {
		return SalesOrderStockReservationResult{}, r.err
	}

	lines := make([]SalesOrderReservedLine, 0, len(input.Lines))
	for _, line := range input.Lines {
		lines = append(lines, SalesOrderReservedLine{
			SalesOrderLineID: line.SalesOrderLineID,
			ReservedQty:      decimal.MustQuantity(line.BaseOrderedQty.String()),
			BatchID:          "batch-reserved",
			BatchNo:          "LOT-RESERVED",
			BinID:            "bin-reserved",
			BinCode:          "PICK-01",
		})
	}

	return SalesOrderStockReservationResult{Lines: lines}, nil
}

func (r *recordingSalesOrderStockReserver) ReleaseSalesOrder(
	_ context.Context,
	input SalesOrderStockReleaseInput,
) (SalesOrderStockReleaseResult, error) {
	r.releaseCalls++
	r.releaseInputs = append(r.releaseInputs, input)
	if r.releaseErr != nil {
		return SalesOrderStockReleaseResult{}, r.releaseErr
	}

	return SalesOrderStockReleaseResult{ReleasedReservationCount: 1}, nil
}

type missingActiveReservationStockReserver struct {
	recordingSalesOrderStockReserver
}

func (r *missingActiveReservationStockReserver) ReleaseSalesOrder(
	_ context.Context,
	input SalesOrderStockReleaseInput,
) (SalesOrderStockReleaseResult, error) {
	r.releaseCalls++
	r.releaseInputs = append(r.releaseInputs, input)

	return SalesOrderStockReleaseResult{}, nil
}
