package application

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/Chinsusu/ERP-v2/apps/api/internal/modules/returns/domain"
	"github.com/Chinsusu/ERP-v2/apps/api/internal/shared/audit"
)

var ErrReturnReceiptDispositionNotAllowed = errors.New("return receipt disposition action is not allowed")

type ReturnDispositionStore interface {
	FindReceiptByID(ctx context.Context, id string) (domain.ReturnReceipt, error)
	SaveDisposition(ctx context.Context, receipt domain.ReturnReceipt, action domain.ReturnDispositionAction) error
}

type ApplyReturnDisposition struct {
	store    ReturnDispositionStore
	auditLog audit.LogStore
	clock    func() time.Time
}

type ApplyReturnDispositionInput struct {
	ReceiptID   string
	Disposition string
	Note        string
	ActorID     string
	RequestID   string
}

type ReturnDispositionResult struct {
	Action     domain.ReturnDispositionAction
	Receipt    domain.ReturnReceipt
	AuditLogID string
}

func NewApplyReturnDisposition(store ReturnDispositionStore, auditLog audit.LogStore) ApplyReturnDisposition {
	return ApplyReturnDisposition{
		store:    store,
		auditLog: auditLog,
		clock:    func() time.Time { return time.Now().UTC() },
	}
}

func (uc ApplyReturnDisposition) Execute(
	ctx context.Context,
	input ApplyReturnDispositionInput,
) (ReturnDispositionResult, error) {
	if uc.store == nil {
		return ReturnDispositionResult{}, errors.New("return disposition store is required")
	}
	if uc.auditLog == nil {
		return ReturnDispositionResult{}, errors.New("audit log store is required")
	}

	receipt, err := uc.store.FindReceiptByID(ctx, input.ReceiptID)
	if err != nil {
		return ReturnDispositionResult{}, err
	}
	if receipt.Status != domain.ReturnStatusInspected {
		return ReturnDispositionResult{}, ErrReturnReceiptDispositionNotAllowed
	}

	action, err := domain.NewReturnDispositionAction(domain.NewReturnDispositionActionInput{
		ReceiptID:   receipt.ID,
		ReceiptNo:   receipt.ReceiptNo,
		Disposition: domain.ReturnDisposition(input.Disposition),
		ActorID:     input.ActorID,
		Note:        input.Note,
		DecidedAt:   uc.clock(),
	})
	if err != nil {
		return ReturnDispositionResult{}, err
	}

	updatedReceipt := receipt.ApplyDisposition(action)
	if err := uc.store.SaveDisposition(ctx, updatedReceipt, action); err != nil {
		return ReturnDispositionResult{}, err
	}

	log, err := newReturnDispositionAuditLog(input.ActorID, input.RequestID, receipt, updatedReceipt, action)
	if err != nil {
		return ReturnDispositionResult{}, err
	}
	if err := uc.auditLog.Record(ctx, log); err != nil {
		return ReturnDispositionResult{}, err
	}

	return ReturnDispositionResult{
		Action:     action,
		Receipt:    updatedReceipt,
		AuditLogID: log.ID,
	}, nil
}

func newReturnDispositionAuditLog(
	actorID string,
	requestID string,
	before domain.ReturnReceipt,
	after domain.ReturnReceipt,
	action domain.ReturnDispositionAction,
) (audit.Log, error) {
	return audit.NewLog(audit.NewLogInput{
		OrgID:      "org-my-pham",
		ActorID:    strings.TrimSpace(actorID),
		Action:     "returns.inspection.disposition",
		EntityType: "returns.return_receipt",
		EntityID:   after.ID,
		RequestID:  strings.TrimSpace(requestID),
		BeforeData: map[string]any{
			"status":          string(before.Status),
			"disposition":     string(before.Disposition),
			"target_location": before.TargetLocation,
		},
		AfterData: map[string]any{
			"action_id":           action.ID,
			"status":              string(after.Status),
			"disposition":         string(action.Disposition),
			"target_location":     action.TargetLocation,
			"target_stock_status": action.TargetStockStatus,
			"action_code":         action.ActionCode,
		},
		Metadata: map[string]any{
			"source": "return disposition",
		},
		CreatedAt: action.DecidedAt,
	})
}

func NewPrototypeApplyReturnDispositionAt(
	store ReturnDispositionStore,
	auditLog audit.LogStore,
	now time.Time,
) ApplyReturnDisposition {
	service := NewApplyReturnDisposition(store, auditLog)
	service.clock = func() time.Time { return now.UTC() }

	return service
}
