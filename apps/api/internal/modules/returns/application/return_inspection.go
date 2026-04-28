package application

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/Chinsusu/ERP-v2/apps/api/internal/modules/returns/domain"
	"github.com/Chinsusu/ERP-v2/apps/api/internal/shared/audit"
)

var ErrReturnReceiptNotInspectable = errors.New("return receipt is not inspectable")

type ReturnInspectionStore interface {
	FindReceiptByID(ctx context.Context, id string) (domain.ReturnReceipt, error)
	SaveInspection(ctx context.Context, receipt domain.ReturnReceipt, inspection domain.ReturnInspection) error
}

type InspectReturn struct {
	store    ReturnInspectionStore
	auditLog audit.LogStore
	clock    func() time.Time
}

type InspectReturnInput struct {
	ReceiptID     string
	Condition     string
	Disposition   string
	Note          string
	EvidenceLabel string
	ActorID       string
	RequestID     string
}

type ReturnInspectionResult struct {
	Inspection domain.ReturnInspection
	Receipt    domain.ReturnReceipt
	AuditLogID string
}

func NewInspectReturn(store ReturnInspectionStore, auditLog audit.LogStore) InspectReturn {
	return InspectReturn{
		store:    store,
		auditLog: auditLog,
		clock:    func() time.Time { return time.Now().UTC() },
	}
}

func (uc InspectReturn) Execute(ctx context.Context, input InspectReturnInput) (ReturnInspectionResult, error) {
	if uc.store == nil {
		return ReturnInspectionResult{}, errors.New("return inspection store is required")
	}
	if uc.auditLog == nil {
		return ReturnInspectionResult{}, errors.New("audit log store is required")
	}

	receipt, err := uc.store.FindReceiptByID(ctx, input.ReceiptID)
	if err != nil {
		return ReturnInspectionResult{}, err
	}
	if receipt.Status != domain.ReturnStatusPendingInspection {
		return ReturnInspectionResult{}, ErrReturnReceiptNotInspectable
	}

	inspection, err := domain.NewReturnInspection(domain.NewReturnInspectionInput{
		ReceiptID:     receipt.ID,
		ReceiptNo:     receipt.ReceiptNo,
		Condition:     domain.ReturnInspectionCondition(input.Condition),
		Disposition:   domain.ReturnDisposition(input.Disposition),
		InspectorID:   input.ActorID,
		Note:          input.Note,
		EvidenceLabel: input.EvidenceLabel,
		InspectedAt:   uc.clock(),
	})
	if err != nil {
		return ReturnInspectionResult{}, err
	}

	updatedReceipt := receipt.ApplyInspection(inspection)
	if err := uc.store.SaveInspection(ctx, updatedReceipt, inspection); err != nil {
		return ReturnInspectionResult{}, err
	}

	log, err := newReturnInspectionAuditLog(input.ActorID, input.RequestID, receipt, updatedReceipt, inspection)
	if err != nil {
		return ReturnInspectionResult{}, err
	}
	if err := uc.auditLog.Record(ctx, log); err != nil {
		return ReturnInspectionResult{}, err
	}

	return ReturnInspectionResult{
		Inspection: inspection,
		Receipt:    updatedReceipt,
		AuditLogID: log.ID,
	}, nil
}

func newReturnInspectionAuditLog(
	actorID string,
	requestID string,
	before domain.ReturnReceipt,
	after domain.ReturnReceipt,
	inspection domain.ReturnInspection,
) (audit.Log, error) {
	return audit.NewLog(audit.NewLogInput{
		OrgID:      "org-my-pham",
		ActorID:    strings.TrimSpace(actorID),
		Action:     "returns.receipt.inspected",
		EntityType: "returns.return_receipt",
		EntityID:   after.ID,
		RequestID:  strings.TrimSpace(requestID),
		BeforeData: map[string]any{
			"status":          string(before.Status),
			"disposition":     string(before.Disposition),
			"target_location": before.TargetLocation,
		},
		AfterData: map[string]any{
			"inspection_id":   inspection.ID,
			"status":          string(after.Status),
			"condition":       string(inspection.Condition),
			"disposition":     string(inspection.Disposition),
			"target_location": inspection.TargetLocation,
			"risk_level":      inspection.RiskLevel,
		},
		Metadata: map[string]any{
			"source": "return inspection",
		},
		CreatedAt: inspection.InspectedAt,
	})
}

func NewPrototypeInspectReturnAt(store ReturnInspectionStore, auditLog audit.LogStore, now time.Time) InspectReturn {
	service := NewInspectReturn(store, auditLog)
	service.clock = func() time.Time { return now.UTC() }

	return service
}
