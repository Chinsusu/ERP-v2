package application

import (
	"context"
	"fmt"
	"strings"
	"time"

	productiondomain "github.com/Chinsusu/ERP-v2/apps/api/internal/modules/production/domain"
)

type SubcontractSampleApprovalService struct {
	clock func() time.Time
}

type BuildSubcontractSampleSubmissionInput struct {
	ID             string
	Order          productiondomain.SubcontractOrder
	SampleCode     string
	FormulaVersion string
	SpecVersion    string
	Evidence       []BuildSubcontractSampleEvidenceInput
	SubmittedBy    string
	SubmittedAt    time.Time
	Note           string
	ActorID        string
}

type BuildSubcontractSampleDecisionInput struct {
	Order          productiondomain.SubcontractOrder
	SampleApproval productiondomain.SubcontractSampleApproval
	DecisionBy     string
	DecisionAt     time.Time
	Reason         string
	StorageStatus  string
	ActorID        string
}

type BuildSubcontractSampleEvidenceInput struct {
	ID           string
	EvidenceType string
	FileName     string
	ObjectKey    string
	ExternalURL  string
	Note         string
}

type SubcontractSampleApprovalBuildResult struct {
	SampleApproval      productiondomain.SubcontractSampleApproval
	UpdatedOrder        productiondomain.SubcontractOrder
	PreviousOrderStatus productiondomain.SubcontractOrderStatus
	CurrentOrderStatus  productiondomain.SubcontractOrderStatus
}

func NewSubcontractSampleApprovalService() SubcontractSampleApprovalService {
	return SubcontractSampleApprovalService{clock: func() time.Time { return time.Now().UTC() }}
}

func (s SubcontractSampleApprovalService) BuildSubmission(
	ctx context.Context,
	input BuildSubcontractSampleSubmissionInput,
) (SubcontractSampleApprovalBuildResult, error) {
	_ = ctx
	actorID := strings.TrimSpace(input.ActorID)
	if actorID == "" {
		return SubcontractSampleApprovalBuildResult{}, productiondomain.ErrSubcontractOrderTransitionActorRequired
	}

	submittedAt := input.SubmittedAt
	if submittedAt.IsZero() {
		submittedAt = s.now()
	}
	updatedOrder, err := input.Order.SubmitSample(actorID, submittedAt)
	if err != nil {
		return SubcontractSampleApprovalBuildResult{}, err
	}

	id := strings.TrimSpace(input.ID)
	if id == "" {
		id = newSubcontractSampleApprovalID(input.Order.ID, submittedAt)
	}
	sampleCode := firstNonBlankSubcontractOrder(input.SampleCode, newSubcontractSampleCode(input.Order.OrderNo, submittedAt))
	sampleApproval, err := productiondomain.NewSubcontractSampleApproval(productiondomain.NewSubcontractSampleApprovalInput{
		ID:                 id,
		OrgID:              input.Order.OrgID,
		SubcontractOrderID: input.Order.ID,
		SubcontractOrderNo: input.Order.OrderNo,
		SampleCode:         sampleCode,
		FormulaVersion:     input.FormulaVersion,
		SpecVersion:        firstNonBlankSubcontractOrder(input.SpecVersion, input.Order.SpecSummary),
		Evidence:           subcontractSampleEvidenceInputs(id, input.Evidence),
		SubmittedBy:        firstNonBlankSubcontractOrder(input.SubmittedBy, actorID),
		SubmittedAt:        submittedAt,
		Note:               input.Note,
		CreatedAt:          submittedAt,
		CreatedBy:          actorID,
		UpdatedAt:          submittedAt,
		UpdatedBy:          actorID,
	})
	if err != nil {
		return SubcontractSampleApprovalBuildResult{}, err
	}

	return SubcontractSampleApprovalBuildResult{
		SampleApproval:      sampleApproval,
		UpdatedOrder:        updatedOrder,
		PreviousOrderStatus: input.Order.Status,
		CurrentOrderStatus:  updatedOrder.Status,
	}, nil
}

func (s SubcontractSampleApprovalService) BuildApproval(
	ctx context.Context,
	input BuildSubcontractSampleDecisionInput,
) (SubcontractSampleApprovalBuildResult, error) {
	_ = ctx
	actorID := firstNonBlankSubcontractOrder(input.ActorID, input.DecisionBy)
	if strings.TrimSpace(actorID) == "" {
		return SubcontractSampleApprovalBuildResult{}, productiondomain.ErrSubcontractOrderTransitionActorRequired
	}
	decisionAt := input.DecisionAt
	if decisionAt.IsZero() {
		decisionAt = s.now()
	}
	approvedSample, err := input.SampleApproval.Approve(productiondomain.ApproveSubcontractSampleApprovalInput{
		ActorID:       actorID,
		DecisionAt:    decisionAt,
		Reason:        input.Reason,
		StorageStatus: input.StorageStatus,
	})
	if err != nil {
		return SubcontractSampleApprovalBuildResult{}, err
	}
	updatedOrder, err := input.Order.ApproveSample(actorID, decisionAt)
	if err != nil {
		return SubcontractSampleApprovalBuildResult{}, err
	}

	return SubcontractSampleApprovalBuildResult{
		SampleApproval:      approvedSample,
		UpdatedOrder:        updatedOrder,
		PreviousOrderStatus: input.Order.Status,
		CurrentOrderStatus:  updatedOrder.Status,
	}, nil
}

func (s SubcontractSampleApprovalService) BuildRejection(
	ctx context.Context,
	input BuildSubcontractSampleDecisionInput,
) (SubcontractSampleApprovalBuildResult, error) {
	_ = ctx
	actorID := firstNonBlankSubcontractOrder(input.ActorID, input.DecisionBy)
	if strings.TrimSpace(actorID) == "" {
		return SubcontractSampleApprovalBuildResult{}, productiondomain.ErrSubcontractOrderTransitionActorRequired
	}
	decisionAt := input.DecisionAt
	if decisionAt.IsZero() {
		decisionAt = s.now()
	}
	rejectedSample, err := input.SampleApproval.Reject(productiondomain.RejectSubcontractSampleApprovalInput{
		ActorID:    actorID,
		DecisionAt: decisionAt,
		Reason:     input.Reason,
	})
	if err != nil {
		return SubcontractSampleApprovalBuildResult{}, err
	}
	updatedOrder, err := input.Order.RejectSample(actorID, input.Reason, decisionAt)
	if err != nil {
		return SubcontractSampleApprovalBuildResult{}, err
	}

	return SubcontractSampleApprovalBuildResult{
		SampleApproval:      rejectedSample,
		UpdatedOrder:        updatedOrder,
		PreviousOrderStatus: input.Order.Status,
		CurrentOrderStatus:  updatedOrder.Status,
	}, nil
}

func (s SubcontractSampleApprovalService) now() time.Time {
	if s.clock == nil {
		return time.Now().UTC()
	}

	return s.clock().UTC()
}

func subcontractSampleEvidenceInputs(
	sampleApprovalID string,
	inputs []BuildSubcontractSampleEvidenceInput,
) []productiondomain.NewSubcontractSampleApprovalEvidenceInput {
	evidence := make([]productiondomain.NewSubcontractSampleApprovalEvidenceInput, 0, len(inputs))
	for index, input := range inputs {
		id := strings.TrimSpace(input.ID)
		if id == "" {
			id = fmt.Sprintf("%s-evidence-%02d", strings.TrimSpace(sampleApprovalID), index+1)
		}
		evidence = append(evidence, productiondomain.NewSubcontractSampleApprovalEvidenceInput{
			ID:           id,
			EvidenceType: input.EvidenceType,
			FileName:     input.FileName,
			ObjectKey:    input.ObjectKey,
			ExternalURL:  input.ExternalURL,
			Note:         input.Note,
		})
	}

	return evidence
}

func newSubcontractSampleApprovalID(orderID string, now time.Time) string {
	return fmt.Sprintf("%s-sample-%s-%06d", strings.TrimSpace(orderID), now.UTC().Format("060102"), now.UTC().UnixNano()%1000000)
}

func newSubcontractSampleCode(orderNo string, now time.Time) string {
	return fmt.Sprintf("%s-SAMPLE-%s", strings.ToUpper(strings.TrimSpace(orderNo)), now.UTC().Format("060102-150405"))
}
