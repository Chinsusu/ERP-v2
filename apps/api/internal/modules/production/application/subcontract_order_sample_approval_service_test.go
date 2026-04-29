package application

import (
	"context"
	"errors"
	"testing"
	"time"

	productiondomain "github.com/Chinsusu/ERP-v2/apps/api/internal/modules/production/domain"
	"github.com/Chinsusu/ERP-v2/apps/api/internal/shared/audit"
)

func TestSubcontractOrderServiceSampleSubmitApproveWithAudit(t *testing.T) {
	ctx := context.Background()
	auditStore := audit.NewInMemoryLogStore()
	orderStore := NewPrototypeSubcontractOrderStore(auditStore)
	sampleStore := NewPrototypeSubcontractSampleApprovalStore()
	service := SubcontractOrderService{
		store:               orderStore,
		sampleApprovalStore: sampleStore,
		sampleApprovalBuild: NewSubcontractSampleApprovalService(),
	}
	order := subcontractSampleOrderReadyForSample(t)
	if err := orderStore.WithinTx(ctx, func(txCtx context.Context, tx SubcontractOrderTx) error {
		return tx.Save(txCtx, order)
	}); err != nil {
		t.Fatalf("seed order: %v", err)
	}

	submitted, err := service.SubmitSubcontractSample(ctx, SubmitSubcontractSampleInput{
		ID:               order.ID,
		ExpectedVersion:  order.Version,
		SampleApprovalID: "sample-001",
		SampleCode:       "SCO-260429-001-SAMPLE-A",
		SubmittedBy:      "factory-user",
		SubmittedAt:      time.Date(2026, 4, 29, 11, 0, 0, 0, time.UTC),
		ActorID:          "qa-user",
		RequestID:        "req-sample-submit",
		Evidence: []SubcontractSampleEvidenceInput{
			{
				EvidenceType: "photo",
				FileName:     "sample-front.jpg",
				ObjectKey:    "subcontract/sco-001/sample-front.jpg",
			},
		},
	})
	if err != nil {
		t.Fatalf("submit sample: %v", err)
	}
	approved, err := service.ApproveSubcontractSample(ctx, DecideSubcontractSampleInput{
		ID:               order.ID,
		ExpectedVersion:  submitted.SubcontractOrder.Version,
		SampleApprovalID: submitted.SampleApproval.ID,
		Reason:           "approved against spec",
		StorageStatus:    "retained_in_qa_cabinet",
		ActorID:          "qa-lead",
		RequestID:        "req-sample-approve",
	})
	if err != nil {
		t.Fatalf("approve sample: %v", err)
	}

	if submitted.CurrentStatus != productiondomain.SubcontractOrderStatusSampleSubmitted ||
		approved.CurrentStatus != productiondomain.SubcontractOrderStatusSampleApproved ||
		approved.SampleApproval.Status != productiondomain.SubcontractSampleApprovalStatusApproved ||
		sampleStore.Count() != 1 {
		t.Fatalf("sample approval result = submitted %+v approved %+v, want submitted then approved", submitted, approved)
	}
	logs, err := auditStore.List(ctx, audit.Query{EntityID: order.ID})
	if err != nil {
		t.Fatalf("list audit logs: %v", err)
	}
	if !subcontractAuditActionsContain(logs, subcontractSampleSubmittedAction) ||
		!subcontractAuditActionsContain(logs, subcontractSampleApprovedAction) {
		t.Fatalf("audit logs = %+v, want sample submit and approve actions", logs)
	}
}

func TestSubcontractOrderServiceSampleRejectionRequiresReason(t *testing.T) {
	ctx := context.Background()
	auditStore := audit.NewInMemoryLogStore()
	orderStore := NewPrototypeSubcontractOrderStore(auditStore)
	sampleStore := NewPrototypeSubcontractSampleApprovalStore()
	service := SubcontractOrderService{
		store:               orderStore,
		sampleApprovalStore: sampleStore,
		sampleApprovalBuild: NewSubcontractSampleApprovalService(),
	}
	order := subcontractSampleOrderReadyForSample(t)
	if err := orderStore.WithinTx(ctx, func(txCtx context.Context, tx SubcontractOrderTx) error {
		return tx.Save(txCtx, order)
	}); err != nil {
		t.Fatalf("seed order: %v", err)
	}
	submitted, err := service.SubmitSubcontractSample(ctx, SubmitSubcontractSampleInput{
		ID:               order.ID,
		ExpectedVersion:  order.Version,
		SampleApprovalID: "sample-001",
		SampleCode:       "SCO-260429-001-SAMPLE-A",
		SubmittedBy:      "factory-user",
		ActorID:          "qa-user",
		RequestID:        "req-sample-submit",
		Evidence: []SubcontractSampleEvidenceInput{
			{
				EvidenceType: "photo",
				ObjectKey:    "subcontract/sco-001/sample-front.jpg",
			},
		},
	})
	if err != nil {
		t.Fatalf("submit sample: %v", err)
	}

	_, err = service.RejectSubcontractSample(ctx, DecideSubcontractSampleInput{
		ID:               order.ID,
		ExpectedVersion:  submitted.SubcontractOrder.Version,
		SampleApprovalID: submitted.SampleApproval.ID,
		Reason:           " ",
		ActorID:          "qa-lead",
		RequestID:        "req-sample-reject",
	})
	if !errors.Is(err, productiondomain.ErrSubcontractSampleApprovalRequiredField) {
		t.Fatalf("error = %v, want required sample rejection reason", err)
	}
	if sampleStore.Count() != 1 {
		t.Fatalf("sample count = %d, want unchanged submitted sample", sampleStore.Count())
	}
	logs, err := auditStore.List(ctx, audit.Query{Action: subcontractSampleRejectedAction})
	if err != nil {
		t.Fatalf("list audit logs: %v", err)
	}
	if len(logs) != 0 {
		t.Fatalf("rejected audit logs = %+v, want none after validation failure", logs)
	}
}

func subcontractAuditActionsContain(logs []audit.Log, action string) bool {
	for _, log := range logs {
		if log.Action == action {
			return true
		}
	}

	return false
}
