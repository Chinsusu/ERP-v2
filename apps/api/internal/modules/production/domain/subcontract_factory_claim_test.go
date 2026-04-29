package domain

import (
	"errors"
	"testing"
	"time"

	"github.com/Chinsusu/ERP-v2/apps/api/internal/shared/decimal"
)

func TestNewSubcontractFactoryClaimCreatesOpenClaimWithSLA(t *testing.T) {
	openedAt := time.Date(2026, 4, 29, 15, 0, 0, 0, time.UTC)

	claim, err := NewSubcontractFactoryClaim(validSubcontractFactoryClaimInput(openedAt))
	if err != nil {
		t.Fatalf("new factory claim: %v", err)
	}

	if claim.Status != SubcontractFactoryClaimStatusOpen ||
		claim.ClaimNo != "SFC-20260429-001" ||
		claim.ReasonCode != "PACKAGING_DAMAGED" ||
		claim.Evidence[0].EvidenceType != "qc_photo" {
		t.Fatalf("claim = %+v, want normalized open claim", claim)
	}
	if !claim.BlocksFinalPayment() {
		t.Fatal("open claim should block final payment")
	}
	if !claim.IsOverdue(openedAt.AddDate(0, 0, 8)) {
		t.Fatal("claim should be overdue after due date")
	}
}

func TestSubcontractFactoryClaimRejectsMissingEvidenceInvalidSLAAndQuantity(t *testing.T) {
	openedAt := time.Date(2026, 4, 29, 15, 0, 0, 0, time.UTC)

	input := validSubcontractFactoryClaimInput(openedAt)
	input.Evidence = nil
	_, err := NewSubcontractFactoryClaim(input)
	if !errors.Is(err, ErrSubcontractFactoryClaimRequiredField) {
		t.Fatalf("missing evidence error = %v, want required field", err)
	}

	input = validSubcontractFactoryClaimInput(openedAt)
	input.DueAt = openedAt.AddDate(0, 0, 2)
	_, err = NewSubcontractFactoryClaim(input)
	if !errors.Is(err, ErrSubcontractFactoryClaimInvalidSLA) {
		t.Fatalf("short sla error = %v, want invalid sla", err)
	}

	input = validSubcontractFactoryClaimInput(openedAt)
	input.AffectedQty = decimal.MustQuantity("0")
	_, err = NewSubcontractFactoryClaim(input)
	if !errors.Is(err, ErrSubcontractFactoryClaimInvalidQuantity) {
		t.Fatalf("zero quantity error = %v, want invalid quantity", err)
	}
}

func TestSubcontractFactoryClaimAcknowledgeAndResolve(t *testing.T) {
	openedAt := time.Date(2026, 4, 29, 15, 0, 0, 0, time.UTC)
	claim, err := NewSubcontractFactoryClaim(validSubcontractFactoryClaimInput(openedAt))
	if err != nil {
		t.Fatalf("new factory claim: %v", err)
	}

	acknowledged, err := claim.Acknowledge("factory-owner", openedAt.Add(time.Hour))
	if err != nil {
		t.Fatalf("acknowledge claim: %v", err)
	}
	if acknowledged.Status != SubcontractFactoryClaimStatusAcknowledged ||
		!acknowledged.BlocksFinalPayment() ||
		acknowledged.Version != claim.Version+1 {
		t.Fatalf("acknowledged = %+v, want acknowledged blocking claim", acknowledged)
	}

	resolved, err := acknowledged.Resolve("qa-lead", "factory credit note accepted", openedAt.Add(2*time.Hour))
	if err != nil {
		t.Fatalf("resolve claim: %v", err)
	}
	if resolved.Status != SubcontractFactoryClaimStatusResolved ||
		resolved.BlocksFinalPayment() ||
		resolved.ResolutionNote != "factory credit note accepted" {
		t.Fatalf("resolved = %+v, want non-blocking resolved claim", resolved)
	}
}

func validSubcontractFactoryClaimInput(openedAt time.Time) NewSubcontractFactoryClaimInput {
	return NewSubcontractFactoryClaimInput{
		ID:                 "sfc_001",
		OrgID:              "org-my-pham",
		ClaimNo:            " sfc-20260429-001 ",
		SubcontractOrderID: "sco_001",
		SubcontractOrderNo: "sco-20260429-001",
		FactoryID:          "fac_001",
		FactoryCode:        "fac-hcm-01",
		FactoryName:        "HCM Cosmetics Factory",
		ReceiptID:          "sfgr_001",
		ReceiptNo:          "sfgr-20260429-001",
		ReasonCode:         " packaging_damaged ",
		Reason:             "Outer cartons crushed and bottle caps scratched",
		Severity:           "p1",
		AffectedQty:        decimal.MustQuantity("12"),
		UOMCode:            "pcs",
		BaseAffectedQty:    decimal.MustQuantity("12"),
		BaseUOMCode:        "pcs",
		Evidence: []NewSubcontractFactoryClaimEvidenceInput{
			{
				ID:           "sfc_001_evidence_01",
				EvidenceType: " QC_PHOTO ",
				FileName:     "damaged-cartons.jpg",
				ObjectKey:    "subcontract/sfc_001/damaged-cartons.jpg",
			},
		},
		OwnerID:   "factory-owner",
		OpenedBy:  "qa-user",
		OpenedAt:  openedAt,
		DueAt:     openedAt.AddDate(0, 0, 5),
		CreatedAt: openedAt,
		CreatedBy: "qa-user",
		UpdatedAt: openedAt,
		UpdatedBy: "qa-user",
	}
}
