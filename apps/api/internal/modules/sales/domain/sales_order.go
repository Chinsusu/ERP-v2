package domain

import (
	"errors"
	"strings"
	"time"
)

var ErrSalesOrderInvalidStatus = errors.New("sales order status is invalid")
var ErrSalesOrderInvalidTransition = errors.New("sales order status transition is invalid")
var ErrSalesOrderTransitionActorRequired = errors.New("sales order status transition actor is required")

type SalesOrderStatus string

const (
	SalesOrderStatusDraft             SalesOrderStatus = "draft"
	SalesOrderStatusConfirmed         SalesOrderStatus = "confirmed"
	SalesOrderStatusReserved          SalesOrderStatus = "reserved"
	SalesOrderStatusPicking           SalesOrderStatus = "picking"
	SalesOrderStatusPicked            SalesOrderStatus = "picked"
	SalesOrderStatusPacking           SalesOrderStatus = "packing"
	SalesOrderStatusPacked            SalesOrderStatus = "packed"
	SalesOrderStatusWaitingHandover   SalesOrderStatus = "waiting_handover"
	SalesOrderStatusHandedOver        SalesOrderStatus = "handed_over"
	SalesOrderStatusDelivered         SalesOrderStatus = "delivered"
	SalesOrderStatusReturned          SalesOrderStatus = "returned"
	SalesOrderStatusClosed            SalesOrderStatus = "closed"
	SalesOrderStatusCancelled         SalesOrderStatus = "cancelled"
	SalesOrderStatusReservationFailed SalesOrderStatus = "reservation_failed"
	SalesOrderStatusPickException     SalesOrderStatus = "pick_exception"
	SalesOrderStatusPackException     SalesOrderStatus = "pack_exception"
	SalesOrderStatusHandoverException SalesOrderStatus = "handover_exception"
)

type SalesOrder struct {
	ID        string
	Status    SalesOrderStatus
	UpdatedAt time.Time
	UpdatedBy string

	ConfirmedAt       time.Time
	ConfirmedBy       string
	ReservedAt        time.Time
	ReservedBy        string
	PickingStartedAt  time.Time
	PickingStartedBy  string
	PickedAt          time.Time
	PickedBy          string
	PackingStartedAt  time.Time
	PackingStartedBy  string
	PackedAt          time.Time
	PackedBy          string
	WaitingHandoverAt time.Time
	WaitingHandoverBy string
	HandedOverAt      time.Time
	HandedOverBy      string
	ClosedAt          time.Time
	ClosedBy          string
	CancelledAt       time.Time
	CancelledBy       string
	ExceptionAt       time.Time
	ExceptionBy       string
}

var salesOrderTransitions = map[SalesOrderStatus][]SalesOrderStatus{
	SalesOrderStatusDraft: {
		SalesOrderStatusConfirmed,
		SalesOrderStatusCancelled,
	},
	SalesOrderStatusConfirmed: {
		SalesOrderStatusReserved,
		SalesOrderStatusReservationFailed,
		SalesOrderStatusCancelled,
	},
	SalesOrderStatusReserved: {
		SalesOrderStatusPicking,
		SalesOrderStatusCancelled,
	},
	SalesOrderStatusPicking: {
		SalesOrderStatusPicked,
		SalesOrderStatusPickException,
		SalesOrderStatusCancelled,
	},
	SalesOrderStatusPicked: {
		SalesOrderStatusPacking,
		SalesOrderStatusPickException,
		SalesOrderStatusCancelled,
	},
	SalesOrderStatusPacking: {
		SalesOrderStatusPacked,
		SalesOrderStatusPackException,
		SalesOrderStatusCancelled,
	},
	SalesOrderStatusPacked: {
		SalesOrderStatusWaitingHandover,
		SalesOrderStatusPackException,
		SalesOrderStatusCancelled,
	},
	SalesOrderStatusWaitingHandover: {
		SalesOrderStatusHandedOver,
		SalesOrderStatusHandoverException,
	},
	SalesOrderStatusHandedOver: {
		SalesOrderStatusClosed,
	},
	SalesOrderStatusDelivered: {
		SalesOrderStatusClosed,
	},
	SalesOrderStatusReturned: {
		SalesOrderStatusClosed,
	},
}

func NewSalesOrder(status SalesOrderStatus) (SalesOrder, error) {
	status = NormalizeSalesOrderStatus(status)
	if status == "" {
		status = SalesOrderStatusDraft
	}
	if !IsValidSalesOrderStatus(status) {
		return SalesOrder{}, ErrSalesOrderInvalidStatus
	}

	return SalesOrder{Status: status}, nil
}

func (o SalesOrder) Confirm(actorID string, changedAt time.Time) (SalesOrder, error) {
	return o.TransitionTo(SalesOrderStatusConfirmed, actorID, changedAt)
}

func (o SalesOrder) MarkReserved(actorID string, changedAt time.Time) (SalesOrder, error) {
	return o.TransitionTo(SalesOrderStatusReserved, actorID, changedAt)
}

func (o SalesOrder) MarkReservationFailed(actorID string, changedAt time.Time) (SalesOrder, error) {
	return o.TransitionTo(SalesOrderStatusReservationFailed, actorID, changedAt)
}

func (o SalesOrder) StartPicking(actorID string, changedAt time.Time) (SalesOrder, error) {
	return o.TransitionTo(SalesOrderStatusPicking, actorID, changedAt)
}

func (o SalesOrder) MarkPicked(actorID string, changedAt time.Time) (SalesOrder, error) {
	return o.TransitionTo(SalesOrderStatusPicked, actorID, changedAt)
}

func (o SalesOrder) MarkPickException(actorID string, changedAt time.Time) (SalesOrder, error) {
	return o.TransitionTo(SalesOrderStatusPickException, actorID, changedAt)
}

func (o SalesOrder) StartPacking(actorID string, changedAt time.Time) (SalesOrder, error) {
	return o.TransitionTo(SalesOrderStatusPacking, actorID, changedAt)
}

func (o SalesOrder) MarkPacked(actorID string, changedAt time.Time) (SalesOrder, error) {
	return o.TransitionTo(SalesOrderStatusPacked, actorID, changedAt)
}

func (o SalesOrder) MarkPackException(actorID string, changedAt time.Time) (SalesOrder, error) {
	return o.TransitionTo(SalesOrderStatusPackException, actorID, changedAt)
}

func (o SalesOrder) MarkWaitingHandover(actorID string, changedAt time.Time) (SalesOrder, error) {
	return o.TransitionTo(SalesOrderStatusWaitingHandover, actorID, changedAt)
}

func (o SalesOrder) MarkHandedOver(actorID string, changedAt time.Time) (SalesOrder, error) {
	return o.TransitionTo(SalesOrderStatusHandedOver, actorID, changedAt)
}

func (o SalesOrder) MarkHandoverException(actorID string, changedAt time.Time) (SalesOrder, error) {
	return o.TransitionTo(SalesOrderStatusHandoverException, actorID, changedAt)
}

func (o SalesOrder) Close(actorID string, changedAt time.Time) (SalesOrder, error) {
	return o.TransitionTo(SalesOrderStatusClosed, actorID, changedAt)
}

func (o SalesOrder) Cancel(actorID string, changedAt time.Time) (SalesOrder, error) {
	return o.TransitionTo(SalesOrderStatusCancelled, actorID, changedAt)
}

func (o SalesOrder) TransitionTo(status SalesOrderStatus, actorID string, changedAt time.Time) (SalesOrder, error) {
	actorID = strings.TrimSpace(actorID)
	if actorID == "" {
		return SalesOrder{}, ErrSalesOrderTransitionActorRequired
	}
	from := NormalizeSalesOrderStatus(o.Status)
	to := NormalizeSalesOrderStatus(status)
	if !IsValidSalesOrderStatus(from) || !IsValidSalesOrderStatus(to) {
		return SalesOrder{}, ErrSalesOrderInvalidStatus
	}
	if !CanTransitionSalesOrderStatus(from, to) {
		return SalesOrder{}, ErrSalesOrderInvalidTransition
	}
	if changedAt.IsZero() {
		changedAt = time.Now().UTC()
	}

	updated := o
	updated.Status = to
	updated.UpdatedAt = changedAt.UTC()
	updated.UpdatedBy = actorID
	updated.markTransition(to, actorID, changedAt.UTC())

	return updated, nil
}

func (o *SalesOrder) markTransition(status SalesOrderStatus, actorID string, changedAt time.Time) {
	switch status {
	case SalesOrderStatusConfirmed:
		o.ConfirmedAt = changedAt
		o.ConfirmedBy = actorID
	case SalesOrderStatusReserved:
		o.ReservedAt = changedAt
		o.ReservedBy = actorID
	case SalesOrderStatusPicking:
		o.PickingStartedAt = changedAt
		o.PickingStartedBy = actorID
	case SalesOrderStatusPicked:
		o.PickedAt = changedAt
		o.PickedBy = actorID
	case SalesOrderStatusPacking:
		o.PackingStartedAt = changedAt
		o.PackingStartedBy = actorID
	case SalesOrderStatusPacked:
		o.PackedAt = changedAt
		o.PackedBy = actorID
	case SalesOrderStatusWaitingHandover:
		o.WaitingHandoverAt = changedAt
		o.WaitingHandoverBy = actorID
	case SalesOrderStatusHandedOver:
		o.HandedOverAt = changedAt
		o.HandedOverBy = actorID
	case SalesOrderStatusClosed:
		o.ClosedAt = changedAt
		o.ClosedBy = actorID
	case SalesOrderStatusCancelled:
		o.CancelledAt = changedAt
		o.CancelledBy = actorID
	case SalesOrderStatusReservationFailed, SalesOrderStatusPickException, SalesOrderStatusPackException, SalesOrderStatusHandoverException:
		o.ExceptionAt = changedAt
		o.ExceptionBy = actorID
	}
}

func NormalizeSalesOrderStatus(status SalesOrderStatus) SalesOrderStatus {
	return SalesOrderStatus(strings.ToLower(strings.TrimSpace(string(status))))
}

func IsValidSalesOrderStatus(status SalesOrderStatus) bool {
	switch NormalizeSalesOrderStatus(status) {
	case SalesOrderStatusDraft,
		SalesOrderStatusConfirmed,
		SalesOrderStatusReserved,
		SalesOrderStatusPicking,
		SalesOrderStatusPicked,
		SalesOrderStatusPacking,
		SalesOrderStatusPacked,
		SalesOrderStatusWaitingHandover,
		SalesOrderStatusHandedOver,
		SalesOrderStatusDelivered,
		SalesOrderStatusReturned,
		SalesOrderStatusClosed,
		SalesOrderStatusCancelled,
		SalesOrderStatusReservationFailed,
		SalesOrderStatusPickException,
		SalesOrderStatusPackException,
		SalesOrderStatusHandoverException:
		return true
	default:
		return false
	}
}

func CanTransitionSalesOrderStatus(from SalesOrderStatus, to SalesOrderStatus) bool {
	from = NormalizeSalesOrderStatus(from)
	to = NormalizeSalesOrderStatus(to)
	if from == to || !IsValidSalesOrderStatus(from) || !IsValidSalesOrderStatus(to) {
		return false
	}
	for _, candidate := range salesOrderTransitions[from] {
		if candidate == to {
			return true
		}
	}

	return false
}
