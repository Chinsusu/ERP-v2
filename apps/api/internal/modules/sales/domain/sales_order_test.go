package domain

import (
	"errors"
	"testing"
	"time"
)

func TestNewSalesOrderDefaultsAndNormalizesStatus(t *testing.T) {
	order, err := NewSalesOrder("")
	if err != nil {
		t.Fatalf("new sales order: %v", err)
	}
	if order.Status != SalesOrderStatusDraft {
		t.Fatalf("status = %q, want draft", order.Status)
	}

	order, err = NewSalesOrder(SalesOrderStatus(" CONFIRMED "))
	if err != nil {
		t.Fatalf("new confirmed sales order: %v", err)
	}
	if order.Status != SalesOrderStatusConfirmed {
		t.Fatalf("status = %q, want confirmed", order.Status)
	}

	_, err = NewSalesOrder(SalesOrderStatus("unknown"))
	if !errors.Is(err, ErrSalesOrderInvalidStatus) {
		t.Fatalf("error = %v, want invalid status", err)
	}
}

func TestSalesOrderHappyPathTransitions(t *testing.T) {
	changedAt := time.Date(2026, 4, 28, 10, 0, 0, 0, time.UTC)
	order, err := NewSalesOrder(SalesOrderStatusDraft)
	if err != nil {
		t.Fatalf("new sales order: %v", err)
	}

	steps := []struct {
		name string
		run  func(SalesOrder) (SalesOrder, error)
		want SalesOrderStatus
	}{
		{"confirm", func(o SalesOrder) (SalesOrder, error) { return o.Confirm("sales-user", changedAt) }, SalesOrderStatusConfirmed},
		{"reserve", func(o SalesOrder) (SalesOrder, error) {
			return o.MarkReserved("inventory-user", changedAt.Add(time.Minute))
		}, SalesOrderStatusReserved},
		{"start picking", func(o SalesOrder) (SalesOrder, error) { return o.StartPicking("picker", changedAt.Add(2*time.Minute)) }, SalesOrderStatusPicking},
		{"mark picked", func(o SalesOrder) (SalesOrder, error) { return o.MarkPicked("picker", changedAt.Add(3*time.Minute)) }, SalesOrderStatusPicked},
		{"start packing", func(o SalesOrder) (SalesOrder, error) { return o.StartPacking("packer", changedAt.Add(4*time.Minute)) }, SalesOrderStatusPacking},
		{"mark packed", func(o SalesOrder) (SalesOrder, error) { return o.MarkPacked("packer", changedAt.Add(5*time.Minute)) }, SalesOrderStatusPacked},
		{"waiting handover", func(o SalesOrder) (SalesOrder, error) {
			return o.MarkWaitingHandover("shipper", changedAt.Add(6*time.Minute))
		}, SalesOrderStatusWaitingHandover},
		{"handed over", func(o SalesOrder) (SalesOrder, error) {
			return o.MarkHandedOver("shipper", changedAt.Add(7*time.Minute))
		}, SalesOrderStatusHandedOver},
		{"close", func(o SalesOrder) (SalesOrder, error) { return o.Close("sales-user", changedAt.Add(8*time.Minute)) }, SalesOrderStatusClosed},
	}
	for _, step := range steps {
		order, err = step.run(order)
		if err != nil {
			t.Fatalf("%s: %v", step.name, err)
		}
		if order.Status != step.want {
			t.Fatalf("%s status = %q, want %q", step.name, order.Status, step.want)
		}
	}
	if order.UpdatedBy == "" || order.ClosedBy == "" || order.UpdatedAt.IsZero() || order.ClosedAt.IsZero() {
		t.Fatalf("order transition metadata = %+v, want actor and timestamp", order)
	}
}

func TestSalesOrderRejectsInvalidTransitions(t *testing.T) {
	order, err := NewSalesOrder(SalesOrderStatusDraft)
	if err != nil {
		t.Fatalf("new sales order: %v", err)
	}

	_, err = order.MarkReserved("inventory-user", time.Now())
	if !errors.Is(err, ErrSalesOrderInvalidTransition) {
		t.Fatalf("error = %v, want invalid transition", err)
	}

	cancelled, err := order.Cancel("sales-user", time.Now())
	if err != nil {
		t.Fatalf("cancel: %v", err)
	}
	_, err = cancelled.Confirm("sales-user", time.Now())
	if !errors.Is(err, ErrSalesOrderInvalidTransition) {
		t.Fatalf("error = %v, want invalid transition from terminal status", err)
	}

	_, err = order.Confirm(" ", time.Now())
	if !errors.Is(err, ErrSalesOrderTransitionActorRequired) {
		t.Fatalf("error = %v, want actor required", err)
	}
}

func TestSalesOrderExceptionTransitions(t *testing.T) {
	changedAt := time.Date(2026, 4, 28, 11, 0, 0, 0, time.UTC)
	confirmed := SalesOrder{Status: SalesOrderStatusConfirmed}
	reservationFailed, err := confirmed.MarkReservationFailed("inventory-user", changedAt)
	if err != nil {
		t.Fatalf("reservation failed: %v", err)
	}
	if reservationFailed.Status != SalesOrderStatusReservationFailed || reservationFailed.ExceptionBy == "" || reservationFailed.ExceptionAt.IsZero() {
		t.Fatalf("reservation failed = %+v, want exception metadata", reservationFailed)
	}

	picking := SalesOrder{Status: SalesOrderStatusPicking}
	pickException, err := picking.MarkPickException("picker", changedAt)
	if err != nil {
		t.Fatalf("pick exception: %v", err)
	}
	if pickException.Status != SalesOrderStatusPickException {
		t.Fatalf("status = %q, want pick exception", pickException.Status)
	}

	packing := SalesOrder{Status: SalesOrderStatusPacking}
	packException, err := packing.MarkPackException("packer", changedAt)
	if err != nil {
		t.Fatalf("pack exception: %v", err)
	}
	if packException.Status != SalesOrderStatusPackException {
		t.Fatalf("status = %q, want pack exception", packException.Status)
	}

	waitingHandover := SalesOrder{Status: SalesOrderStatusWaitingHandover}
	handoverException, err := waitingHandover.MarkHandoverException("shipper", changedAt)
	if err != nil {
		t.Fatalf("handover exception: %v", err)
	}
	if handoverException.Status != SalesOrderStatusHandoverException {
		t.Fatalf("status = %q, want handover exception", handoverException.Status)
	}
}

func TestCanTransitionSalesOrderStatus(t *testing.T) {
	if !CanTransitionSalesOrderStatus(SalesOrderStatusDraft, SalesOrderStatusConfirmed) {
		t.Fatal("draft should transition to confirmed")
	}
	if CanTransitionSalesOrderStatus(SalesOrderStatusDraft, SalesOrderStatusReserved) {
		t.Fatal("draft should not skip to reserved")
	}
	if CanTransitionSalesOrderStatus(SalesOrderStatusClosed, SalesOrderStatusCancelled) {
		t.Fatal("closed should be terminal")
	}
	if CanTransitionSalesOrderStatus(SalesOrderStatusDraft, SalesOrderStatusDraft) {
		t.Fatal("same-status transition should not be accepted")
	}
}
