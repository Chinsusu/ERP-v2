package domain

import (
	"errors"
	"testing"
	"time"
)

func TestNewSubcontractSampleApprovalCreatesSubmittedRecordWithEvidence(t *testing.T) {
	submittedAt := time.Date(2026, 4, 29, 11, 0, 0, 0, time.UTC)

	approval, err := NewSubcontractSampleApproval(validSubcontractSampleApprovalInput(submittedAt))
	if err != nil {
		t.Fatalf("new sample approval: %v", err)
	}

	if approval.Status != SubcontractSampleApprovalStatusSubmitted ||
		approval.SampleCode != "SCO-260429-001-SAMPLE-A" ||
		approval.Evidence[0].EvidenceType != "photo" ||
		approval.SubmittedBy != "factory-user" {
		t.Fatalf("approval = %+v, want normalized submitted sample record", approval)
	}
}

func TestSubcontractSampleApprovalApproveRequiresStorageStatus(t *testing.T) {
	approval, err := NewSubcontractSampleApproval(validSubcontractSampleApprovalInput(time.Now()))
	if err != nil {
		t.Fatalf("new sample approval: %v", err)
	}

	_, err = approval.Approve(ApproveSubcontractSampleApprovalInput{
		ActorID:    "qa-lead",
		DecisionAt: time.Now(),
	})
	if !errors.Is(err, ErrSubcontractSampleApprovalRequiredField) {
		t.Fatalf("error = %v, want required storage status", err)
	}

	approved, err := approval.Approve(ApproveSubcontractSampleApprovalInput{
		ActorID:       "qa-lead",
		DecisionAt:    time.Date(2026, 4, 29, 12, 0, 0, 0, time.UTC),
		Reason:        "shade and fill level approved",
		StorageStatus: "retained_in_qa_cabinet",
	})
	if err != nil {
		t.Fatalf("approve sample: %v", err)
	}
	if approved.Status != SubcontractSampleApprovalStatusApproved ||
		approved.DecisionBy != "qa-lead" ||
		approved.StorageStatus != "retained_in_qa_cabinet" ||
		approved.Version != 2 {
		t.Fatalf("approved = %+v, want approved decision metadata", approved)
	}
}

func TestSubcontractSampleApprovalRejectRequiresReason(t *testing.T) {
	approval, err := NewSubcontractSampleApproval(validSubcontractSampleApprovalInput(time.Now()))
	if err != nil {
		t.Fatalf("new sample approval: %v", err)
	}

	_, err = approval.Reject(RejectSubcontractSampleApprovalInput{
		ActorID:    "qa-lead",
		DecisionAt: time.Now(),
		Reason:     " ",
	})
	if !errors.Is(err, ErrSubcontractSampleApprovalRequiredField) {
		t.Fatalf("error = %v, want required rejection reason", err)
	}

	rejected, err := approval.Reject(RejectSubcontractSampleApprovalInput{
		ActorID:    "qa-lead",
		DecisionAt: time.Date(2026, 4, 29, 12, 0, 0, 0, time.UTC),
		Reason:     "label color is wrong",
	})
	if err != nil {
		t.Fatalf("reject sample: %v", err)
	}
	if rejected.Status != SubcontractSampleApprovalStatusRejected ||
		rejected.DecisionReason != "label color is wrong" ||
		rejected.Version != 2 {
		t.Fatalf("rejected = %+v, want rejected decision metadata", rejected)
	}
}

func TestSubcontractSampleApprovalRejectsMissingEvidence(t *testing.T) {
	input := validSubcontractSampleApprovalInput(time.Now())
	input.Evidence = nil

	_, err := NewSubcontractSampleApproval(input)
	if !errors.Is(err, ErrSubcontractSampleApprovalRequiredField) {
		t.Fatalf("error = %v, want required evidence", err)
	}
}

func validSubcontractSampleApprovalInput(submittedAt time.Time) NewSubcontractSampleApprovalInput {
	return NewSubcontractSampleApprovalInput{
		ID:                 "sample-001",
		OrgID:              "org-my-pham",
		SubcontractOrderID: "sco-001",
		SubcontractOrderNo: "sco-260429-001",
		SampleCode:         " sco-260429-001-sample-a ",
		FormulaVersion:     "formula-2026.04",
		SpecVersion:        "spec-2026.04",
		Evidence: []NewSubcontractSampleApprovalEvidenceInput{
			{
				ID:           "sample-evidence-001",
				EvidenceType: "PHOTO",
				FileName:     "sample-front.jpg",
				ObjectKey:    "subcontract/sco-001/sample-front.jpg",
			},
		},
		SubmittedBy: "factory-user",
		SubmittedAt: submittedAt,
		CreatedAt:   submittedAt,
		CreatedBy:   "subcontract-user",
	}
}
