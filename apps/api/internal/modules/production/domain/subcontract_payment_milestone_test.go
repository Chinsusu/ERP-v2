package domain

import (
	"errors"
	"testing"
	"time"

	"github.com/Chinsusu/ERP-v2/apps/api/internal/shared/decimal"
)

func TestNewSubcontractPaymentMilestoneCreatesPendingDeposit(t *testing.T) {
	milestone, err := NewSubcontractPaymentMilestone(validSubcontractPaymentMilestoneInput(SubcontractPaymentMilestoneKindDeposit))
	if err != nil {
		t.Fatalf("new payment milestone: %v", err)
	}

	if milestone.Status != SubcontractPaymentMilestoneStatusPending ||
		milestone.Kind != SubcontractPaymentMilestoneKindDeposit ||
		milestone.CurrencyCode != decimal.CurrencyVND ||
		milestone.MilestoneNo != "SPM-20260429-001" ||
		milestone.FactoryCode != "FAC-HCM-01" {
		t.Fatalf("milestone = %+v, want normalized pending VND deposit", milestone)
	}
	if milestone.BlocksFinalPayment() {
		t.Fatalf("deposit milestone blocks final payment")
	}
}

func TestSubcontractPaymentMilestoneRecordsDeposit(t *testing.T) {
	recordedAt := time.Date(2026, 4, 29, 10, 0, 0, 0, time.UTC)
	milestone, err := NewSubcontractPaymentMilestone(validSubcontractPaymentMilestoneInput(SubcontractPaymentMilestoneKindDeposit))
	if err != nil {
		t.Fatalf("new payment milestone: %v", err)
	}

	recorded, err := milestone.Record("finance-user", recordedAt)
	if err != nil {
		t.Fatalf("record deposit: %v", err)
	}

	if recorded.Status != SubcontractPaymentMilestoneStatusRecorded ||
		recorded.RecordedBy != "finance-user" ||
		!recorded.RecordedAt.Equal(recordedAt) ||
		recorded.Version != milestone.Version+1 {
		t.Fatalf("recorded = %+v, want recorded metadata and version bump", recorded)
	}
	_, err = recorded.Record("finance-user", recordedAt.Add(time.Minute))
	if !errors.Is(err, ErrSubcontractPaymentMilestoneInvalidTransition) {
		t.Fatalf("record twice error = %v, want invalid transition", err)
	}
}

func TestSubcontractPaymentMilestoneFinalPaymentBlockRequiresExceptionBeforeReady(t *testing.T) {
	blockedAt := time.Date(2026, 4, 29, 11, 0, 0, 0, time.UTC)
	milestone, err := NewSubcontractPaymentMilestone(validSubcontractPaymentMilestoneInput(SubcontractPaymentMilestoneKindFinalPayment))
	if err != nil {
		t.Fatalf("new payment milestone: %v", err)
	}

	blocked, err := milestone.Block("qa-lead", "open factory claim", blockedAt)
	if err != nil {
		t.Fatalf("block final payment: %v", err)
	}
	if !blocked.BlocksFinalPayment() ||
		blocked.Status != SubcontractPaymentMilestoneStatusBlocked ||
		blocked.BlockedBy != "qa-lead" ||
		blocked.BlockReason != "open factory claim" {
		t.Fatalf("blocked = %+v, want blocking final payment milestone", blocked)
	}

	_, err = blocked.MarkReady("finance-user", blockedAt.Add(time.Hour))
	if !errors.Is(err, ErrSubcontractPaymentMilestoneBlocked) {
		t.Fatalf("ready without exception error = %v, want blocked", err)
	}

	blocked.ApprovedExceptionID = "exception_001"
	ready, err := blocked.MarkReady("finance-user", blockedAt.Add(time.Hour))
	if err != nil {
		t.Fatalf("ready with approved exception: %v", err)
	}
	if ready.BlocksFinalPayment() ||
		ready.Status != SubcontractPaymentMilestoneStatusReady ||
		ready.ReadyBy != "finance-user" ||
		ready.BlockReason != "" ||
		ready.ApprovedExceptionID != "exception_001" {
		t.Fatalf("ready = %+v, want unblocked ready final payment milestone with exception trace", ready)
	}
}

func TestSubcontractPaymentMilestoneRejectsInvalidInputs(t *testing.T) {
	input := validSubcontractPaymentMilestoneInput(SubcontractPaymentMilestoneKindDeposit)
	input.CurrencyCode = "USD"
	_, err := NewSubcontractPaymentMilestone(input)
	if !errors.Is(err, ErrSubcontractPaymentMilestoneInvalidCurrency) {
		t.Fatalf("currency error = %v, want invalid currency", err)
	}

	input = validSubcontractPaymentMilestoneInput(SubcontractPaymentMilestoneKindDeposit)
	input.Amount = decimal.MustMoneyAmount("0")
	_, err = NewSubcontractPaymentMilestone(input)
	if !errors.Is(err, ErrSubcontractPaymentMilestoneInvalidAmount) {
		t.Fatalf("amount error = %v, want invalid amount", err)
	}

	input = validSubcontractPaymentMilestoneInput(SubcontractPaymentMilestoneKindDeposit)
	input.Kind = "other"
	_, err = NewSubcontractPaymentMilestone(input)
	if !errors.Is(err, ErrSubcontractPaymentMilestoneInvalidKind) {
		t.Fatalf("kind error = %v, want invalid kind", err)
	}
}

func validSubcontractPaymentMilestoneInput(kind SubcontractPaymentMilestoneKind) NewSubcontractPaymentMilestoneInput {
	return NewSubcontractPaymentMilestoneInput{
		ID:                 "spm_001",
		OrgID:              "org-my-pham",
		MilestoneNo:        " spm-20260429-001 ",
		SubcontractOrderID: "sco_001",
		SubcontractOrderNo: " sco-20260429-001 ",
		FactoryID:          "fac_001",
		FactoryCode:        " fac-hcm-01 ",
		FactoryName:        "HCM Cosmetics Factory",
		Kind:               kind,
		Amount:             decimal.MustMoneyAmount("1000000"),
		CurrencyCode:       "VND",
		CreatedAt:          time.Date(2026, 4, 29, 9, 0, 0, 0, time.UTC),
		CreatedBy:          "finance-user",
		UpdatedBy:          "finance-user",
	}
}
