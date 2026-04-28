package domain

import (
	"errors"
	"fmt"
	"strings"
	"time"
)

var ErrReturnDispositionRequiredField = errors.New("return disposition required field is missing")

type ReturnDispositionAction struct {
	ID                string
	ReceiptID         string
	ReceiptNo         string
	Disposition       ReturnDisposition
	TargetLocation    string
	TargetStockStatus string
	ActionCode        string
	ActorID           string
	Note              string
	DecidedAt         time.Time
}

type NewReturnDispositionActionInput struct {
	ID          string
	ReceiptID   string
	ReceiptNo   string
	Disposition ReturnDisposition
	ActorID     string
	Note        string
	DecidedAt   time.Time
}

func NewReturnDispositionAction(input NewReturnDispositionActionInput) (ReturnDispositionAction, error) {
	receiptID := strings.TrimSpace(input.ReceiptID)
	if receiptID == "" || strings.TrimSpace(input.ActorID) == "" {
		return ReturnDispositionAction{}, ErrReturnDispositionRequiredField
	}

	disposition := NormalizeReturnDisposition(input.Disposition)
	if disposition == "" {
		return ReturnDispositionAction{}, ErrReturnReceiptInvalidDisposition
	}

	decidedAt := input.DecidedAt
	if decidedAt.IsZero() {
		decidedAt = time.Now().UTC()
	}

	action := ReturnDispositionAction{
		ID:                strings.TrimSpace(input.ID),
		ReceiptID:         receiptID,
		ReceiptNo:         strings.TrimSpace(input.ReceiptNo),
		Disposition:       disposition,
		TargetLocation:    returnDispositionTargetLocation(disposition),
		TargetStockStatus: returnDispositionTargetStockStatus(disposition),
		ActionCode:        returnDispositionActionCode(disposition),
		ActorID:           strings.TrimSpace(input.ActorID),
		Note:              strings.TrimSpace(input.Note),
		DecidedAt:         decidedAt.UTC(),
	}
	if action.ID == "" {
		action.ID = fmt.Sprintf("dispose-%s-%s", strings.ToLower(receiptID), disposition)
	}

	return action, nil
}

func (action ReturnDispositionAction) Clone() ReturnDispositionAction {
	return action
}

func (r ReturnReceipt) ApplyDisposition(action ReturnDispositionAction) ReturnReceipt {
	receipt := r.Clone()
	receipt.Status = ReturnStatusDispositioned
	receipt.Disposition = action.Disposition
	receipt.TargetLocation = action.TargetLocation
	receipt.StockMovement = nil

	return receipt
}

func returnDispositionTargetLocation(disposition ReturnDisposition) string {
	switch disposition {
	case ReturnDispositionReusable:
		return "return-putaway-ready"
	case ReturnDispositionNotReusable:
		return "lab-damaged-placeholder"
	case ReturnDispositionNeedsInspection:
		return "return-quarantine-hold"
	default:
		return "return-quarantine-hold"
	}
}

func returnDispositionTargetStockStatus(disposition ReturnDisposition) string {
	switch disposition {
	case ReturnDispositionReusable:
		return "return_pending"
	case ReturnDispositionNotReusable:
		return "damaged"
	case ReturnDispositionNeedsInspection:
		return "qc_hold"
	default:
		return "qc_hold"
	}
}

func returnDispositionActionCode(disposition ReturnDisposition) string {
	switch disposition {
	case ReturnDispositionReusable:
		return "route_to_putaway"
	case ReturnDispositionNotReusable:
		return "route_to_lab_or_damaged"
	case ReturnDispositionNeedsInspection:
		return "route_to_quarantine_hold"
	default:
		return "route_to_quarantine_hold"
	}
}
