package domain

import (
	"errors"
	"testing"
	"time"

	"github.com/Chinsusu/ERP-v2/apps/api/internal/shared/decimal"
)

func TestNewSubcontractOrderDocumentCreatesDraftWithMaterials(t *testing.T) {
	order, err := NewSubcontractOrderDocument(validSubcontractOrderInput())
	if err != nil {
		t.Fatalf("new subcontract order: %v", err)
	}

	if order.Status != SubcontractOrderStatusDraft {
		t.Fatalf("status = %q, want draft", order.Status)
	}
	if order.OrderNo != "SCO-20260429-001" || order.FactoryCode != "FAC-HCM-01" || order.FinishedSKUCode != "FG-SERUM-001" {
		t.Fatalf("normalized header = %+v, want uppercase order/factory/sku codes", order)
	}
	if order.CurrencyCode != decimal.CurrencyVND {
		t.Fatalf("currency = %q, want VND", order.CurrencyCode)
	}
	if !order.SampleRequired {
		t.Fatal("sample required = false, want true")
	}
	if got, want := order.BasePlannedQty.String(), "1000.000000"; got != want {
		t.Fatalf("base planned qty = %s, want %s", got, want)
	}
	if got, want := order.EstimatedCostAmount.String(), "2200000.00"; got != want {
		t.Fatalf("estimated cost = %s, want %s", got, want)
	}
	if order.ClaimWindowDays != 7 {
		t.Fatalf("claim window = %d, want default 7", order.ClaimWindowDays)
	}
	if len(order.MaterialLines) != 2 {
		t.Fatalf("material line count = %d, want 2", len(order.MaterialLines))
	}
	first := order.MaterialLines[0]
	if first.LineNo != 1 || first.SKUCode != "RM-BASE-001" || first.UOMCode != "KG" || first.BaseUOMCode != "G" {
		t.Fatalf("first material line = %+v, want normalized fields", first)
	}
	if first.BasePlannedQty.String() != "10000.000000" || first.BaseIssuedQty.String() != "0.000000" {
		t.Fatalf("first material base qty = %s/%s, want planned default and zero issued", first.BasePlannedQty, first.BaseIssuedQty)
	}
}

func TestSubcontractOrderHappyPathWithSampleGate(t *testing.T) {
	changedAt := time.Date(2026, 4, 29, 9, 0, 0, 0, time.UTC)
	order, err := NewSubcontractOrderDocument(validSubcontractOrderInput())
	if err != nil {
		t.Fatalf("new subcontract order: %v", err)
	}

	steps := []struct {
		name string
		run  func(SubcontractOrder) (SubcontractOrder, error)
		want SubcontractOrderStatus
	}{
		{"submit", func(o SubcontractOrder) (SubcontractOrder, error) {
			return o.Submit("subcontract-user", changedAt)
		}, SubcontractOrderStatusSubmitted},
		{"approve", func(o SubcontractOrder) (SubcontractOrder, error) {
			return o.Approve("operations-lead", changedAt.Add(time.Minute))
		}, SubcontractOrderStatusApproved},
		{"confirm factory", func(o SubcontractOrder) (SubcontractOrder, error) {
			return o.ConfirmFactory("factory-coordinator", changedAt.Add(2*time.Minute))
		}, SubcontractOrderStatusFactoryConfirmed},
		{"record deposit", func(o SubcontractOrder) (SubcontractOrder, error) {
			return o.RecordDeposit("finance-user", decimal.MustMoneyAmount("1000000"), changedAt.Add(3*time.Minute))
		}, SubcontractOrderStatusDepositRecorded},
		{"issue materials", func(o SubcontractOrder) (SubcontractOrder, error) {
			return o.MarkMaterialsIssued("warehouse-user", changedAt.Add(4*time.Minute))
		}, SubcontractOrderStatusMaterialsIssued},
		{"submit sample", func(o SubcontractOrder) (SubcontractOrder, error) {
			return o.SubmitSample("qa-user", changedAt.Add(5*time.Minute))
		}, SubcontractOrderStatusSampleSubmitted},
		{"approve sample", func(o SubcontractOrder) (SubcontractOrder, error) {
			return o.ApproveSample("qa-lead", changedAt.Add(6*time.Minute))
		}, SubcontractOrderStatusSampleApproved},
		{"start mass production", func(o SubcontractOrder) (SubcontractOrder, error) {
			return o.StartMassProduction("operations-lead", changedAt.Add(7*time.Minute))
		}, SubcontractOrderStatusMassProductionStarted},
		{"receive finished goods", func(o SubcontractOrder) (SubcontractOrder, error) {
			return o.MarkFinishedGoodsReceived("warehouse-user", changedAt.Add(8*time.Minute))
		}, SubcontractOrderStatusFinishedGoodsReceived},
		{"start qc", func(o SubcontractOrder) (SubcontractOrder, error) {
			return o.StartQC("qa-user", changedAt.Add(9*time.Minute))
		}, SubcontractOrderStatusQCInProgress},
		{"accept", func(o SubcontractOrder) (SubcontractOrder, error) {
			return o.Accept("qa-lead", changedAt.Add(10*time.Minute))
		}, SubcontractOrderStatusAccepted},
		{"final payment ready", func(o SubcontractOrder) (SubcontractOrder, error) {
			return o.MarkFinalPaymentReady("finance-user", changedAt.Add(11*time.Minute))
		}, SubcontractOrderStatusFinalPaymentReady},
		{"close", func(o SubcontractOrder) (SubcontractOrder, error) {
			return o.Close("subcontract-user", changedAt.Add(12*time.Minute))
		}, SubcontractOrderStatusClosed},
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
	if order.Version != 14 {
		t.Fatalf("version = %d, want 14", order.Version)
	}
	if order.DepositAmount.String() != "1000000.00" || order.SampleApprovedBy == "" || order.FinalPaymentReadyBy == "" {
		t.Fatalf("transition metadata = %+v, want deposit/sample/final payment metadata", order)
	}
}

func TestSubcontractOrderBlocksMassProductionUntilSampleApproved(t *testing.T) {
	changedAt := time.Date(2026, 4, 29, 9, 0, 0, 0, time.UTC)
	order := subcontractOrderReadyForMaterials(t, true)

	_, err := order.StartMassProduction("operations-lead", changedAt)
	if !errors.Is(err, ErrSubcontractOrderSampleApprovalRequired) {
		t.Fatalf("error = %v, want sample approval required", err)
	}

	submitted, err := order.SubmitSample("qa-user", changedAt)
	if err != nil {
		t.Fatalf("submit sample: %v", err)
	}
	rejected, err := submitted.RejectSample("qa-lead", "wrong shade", changedAt.Add(time.Minute))
	if err != nil {
		t.Fatalf("reject sample: %v", err)
	}
	if rejected.SampleRejectReason != "wrong shade" {
		t.Fatalf("sample reject reason = %q, want trimmed reason", rejected.SampleRejectReason)
	}
	_, err = rejected.StartMassProduction("operations-lead", changedAt.Add(2*time.Minute))
	if !errors.Is(err, ErrSubcontractOrderInvalidTransition) {
		t.Fatalf("error after sample rejection = %v, want invalid transition", err)
	}
}

func TestSubcontractOrderAllowsMassProductionWithoutSampleRequirement(t *testing.T) {
	order := subcontractOrderReadyForMaterials(t, false)

	started, err := order.StartMassProduction("operations-lead", time.Now())
	if err != nil {
		t.Fatalf("start mass production: %v", err)
	}
	if started.Status != SubcontractOrderStatusMassProductionStarted {
		t.Fatalf("status = %q, want mass production started", started.Status)
	}
}

func TestSubcontractOrderRejectsInvalidInputs(t *testing.T) {
	input := validSubcontractOrderInput()
	input.CurrencyCode = "USD"
	_, err := NewSubcontractOrderDocument(input)
	if !errors.Is(err, ErrSubcontractOrderInvalidCurrency) {
		t.Fatalf("error = %v, want invalid currency", err)
	}

	input = validSubcontractOrderInput()
	input.FactoryID = " "
	_, err = NewSubcontractOrderDocument(input)
	if !errors.Is(err, ErrSubcontractOrderRequiredField) {
		t.Fatalf("error = %v, want required field", err)
	}

	input = validSubcontractOrderInput()
	input.PlannedQty = decimal.MustQuantity("0")
	_, err = NewSubcontractOrderDocument(input)
	if !errors.Is(err, ErrSubcontractOrderInvalidQuantity) {
		t.Fatalf("error = %v, want invalid quantity", err)
	}

	input = validSubcontractOrderInput()
	input.ClaimWindowDays = 10
	_, err = NewSubcontractOrderDocument(input)
	if !errors.Is(err, ErrSubcontractOrderRequiredField) {
		t.Fatalf("error = %v, want invalid claim window", err)
	}

	input = validSubcontractOrderInput()
	input.MaterialLines[0].UnitCost = decimal.Decimal("-1.000000")
	_, err = NewSubcontractOrderDocument(input)
	if !errors.Is(err, ErrSubcontractOrderInvalidAmount) {
		t.Fatalf("error = %v, want invalid amount", err)
	}
}

func TestSubcontractMaterialIssuedQuantityCannotExceedPlanned(t *testing.T) {
	input := validSubcontractOrderInput()
	input.MaterialLines[0].IssuedQty = decimal.MustQuantity("11")
	input.MaterialLines[0].BaseIssuedQty = decimal.MustQuantity("11000")
	_, err := NewSubcontractOrderDocument(input)
	if !errors.Is(err, ErrSubcontractOrderInvalidQuantity) {
		t.Fatalf("error = %v, want invalid issued quantity", err)
	}

	input = validSubcontractOrderInput()
	input.ReceivedQty = decimal.MustQuantity("1001")
	input.BaseReceivedQty = decimal.MustQuantity("1001")
	_, err = NewSubcontractOrderDocument(input)
	if !errors.Is(err, ErrSubcontractOrderInvalidQuantity) {
		t.Fatalf("error = %v, want invalid received quantity", err)
	}

	input = validSubcontractOrderInput()
	input.ReceivedQty = decimal.MustQuantity("10")
	input.BaseReceivedQty = decimal.MustQuantity("10")
	input.AcceptedQty = decimal.MustQuantity("7")
	input.BaseAcceptedQty = decimal.MustQuantity("7")
	input.RejectedQty = decimal.MustQuantity("4")
	input.BaseRejectedQty = decimal.MustQuantity("4")
	_, err = NewSubcontractOrderDocument(input)
	if !errors.Is(err, ErrSubcontractOrderInvalidQuantity) {
		t.Fatalf("error = %v, want invalid accepted plus rejected quantity", err)
	}
}

func TestIssueMaterialsUpdatesProgressAndTransitionsWhenComplete(t *testing.T) {
	changedAt := time.Date(2026, 4, 29, 10, 0, 0, 0, time.UTC)
	order, err := subcontractOrderReadyForFactoryConfirmation(t)
	if err != nil {
		t.Fatalf("factory confirmed order: %v", err)
	}

	updated, err := order.IssueMaterials(IssueSubcontractMaterialsInput{
		ActorID:   "warehouse-user",
		ChangedAt: changedAt,
		Lines: []IssueSubcontractMaterialLineInput{
			{
				OrderMaterialLineID: "sco_mat_001",
				IssueQty:            decimal.MustQuantity("10"),
				UOMCode:             "KG",
				BaseIssueQty:        decimal.MustQuantity("10000"),
				BaseUOMCode:         "G",
				ConversionFactor:    decimal.MustQuantity("1000"),
			},
			{
				OrderMaterialLineID: "sco_mat_002",
				IssueQty:            decimal.MustQuantity("1000"),
				UOMCode:             "PCS",
				BaseIssueQty:        decimal.MustQuantity("1000"),
				BaseUOMCode:         "PCS",
				ConversionFactor:    decimal.MustQuantity("1"),
			},
		},
	})
	if err != nil {
		t.Fatalf("issue materials: %v", err)
	}

	if updated.Status != SubcontractOrderStatusMaterialsIssued {
		t.Fatalf("status = %q, want materials issued", updated.Status)
	}
	if updated.MaterialsIssuedBy != "warehouse-user" || !updated.MaterialsIssuedAt.Equal(changedAt) {
		t.Fatalf("materials issued metadata = %s/%s, want actor and timestamp", updated.MaterialsIssuedBy, updated.MaterialsIssuedAt)
	}
	if got, want := updated.MaterialLines[0].BaseIssuedQty.String(), "10000.000000"; got != want {
		t.Fatalf("first base issued qty = %s, want %s", got, want)
	}
	if got, want := updated.MaterialLines[1].IssuedQty.String(), "1000.000000"; got != want {
		t.Fatalf("second issued qty = %s, want %s", got, want)
	}
}

func TestIssueMaterialsAllowsPartialIssueWithoutFinalTransition(t *testing.T) {
	changedAt := time.Date(2026, 4, 29, 10, 0, 0, 0, time.UTC)
	order, err := subcontractOrderReadyForFactoryConfirmation(t)
	if err != nil {
		t.Fatalf("factory confirmed order: %v", err)
	}

	updated, err := order.IssueMaterials(IssueSubcontractMaterialsInput{
		ActorID:   "warehouse-user",
		ChangedAt: changedAt,
		Lines: []IssueSubcontractMaterialLineInput{
			{
				OrderMaterialLineID: "sco_mat_001",
				IssueQty:            decimal.MustQuantity("5"),
				UOMCode:             "KG",
				BaseIssueQty:        decimal.MustQuantity("5000"),
				BaseUOMCode:         "G",
				ConversionFactor:    decimal.MustQuantity("1000"),
			},
		},
	})
	if err != nil {
		t.Fatalf("issue partial materials: %v", err)
	}

	if updated.Status != SubcontractOrderStatusFactoryConfirmed {
		t.Fatalf("status = %q, want factory confirmed until all materials are issued", updated.Status)
	}
	if updated.MaterialsIssuedBy != "" || !updated.MaterialsIssuedAt.IsZero() {
		t.Fatalf("materials issued metadata should be empty for partial issue: %+v", updated)
	}
	if got, want := updated.MaterialLines[0].BaseIssuedQty.String(), "5000.000000"; got != want {
		t.Fatalf("first base issued qty = %s, want %s", got, want)
	}
}

func TestIssueMaterialsRejectsOverIssue(t *testing.T) {
	order, err := subcontractOrderReadyForFactoryConfirmation(t)
	if err != nil {
		t.Fatalf("factory confirmed order: %v", err)
	}

	_, err = order.IssueMaterials(IssueSubcontractMaterialsInput{
		ActorID:   "warehouse-user",
		ChangedAt: time.Now(),
		Lines: []IssueSubcontractMaterialLineInput{
			{
				OrderMaterialLineID: "sco_mat_001",
				IssueQty:            decimal.MustQuantity("11"),
				UOMCode:             "KG",
				BaseIssueQty:        decimal.MustQuantity("11000"),
				BaseUOMCode:         "G",
				ConversionFactor:    decimal.MustQuantity("1000"),
			},
		},
	})
	if !errors.Is(err, ErrSubcontractOrderInvalidQuantity) {
		t.Fatalf("error = %v, want invalid quantity", err)
	}
}

func TestSubcontractOrderRejectsInvalidTransitionsAndMissingActor(t *testing.T) {
	order, err := NewSubcontractOrder(SubcontractOrderStatusDraft)
	if err != nil {
		t.Fatalf("new subcontract order: %v", err)
	}

	_, err = order.Approve("operations-lead", time.Now())
	if !errors.Is(err, ErrSubcontractOrderInvalidTransition) {
		t.Fatalf("error = %v, want invalid transition", err)
	}
	_, err = order.Submit(" ", time.Now())
	if !errors.Is(err, ErrSubcontractOrderTransitionActorRequired) {
		t.Fatalf("error = %v, want actor required", err)
	}

	cancelled, err := order.Cancel("operations-lead", time.Now())
	if err != nil {
		t.Fatalf("cancel: %v", err)
	}
	_, err = cancelled.Submit("operations-lead", time.Now())
	if !errors.Is(err, ErrSubcontractOrderInvalidTransition) {
		t.Fatalf("error = %v, want terminal cancelled status", err)
	}
}

func TestCanTransitionSubcontractOrderStatus(t *testing.T) {
	if !CanTransitionSubcontractOrderStatus(SubcontractOrderStatusDraft, SubcontractOrderStatusSubmitted) {
		t.Fatal("draft should transition to submitted")
	}
	if CanTransitionSubcontractOrderStatus(SubcontractOrderStatusDraft, SubcontractOrderStatusApproved) {
		t.Fatal("draft should not skip approval")
	}
	if CanTransitionSubcontractOrderStatus(SubcontractOrderStatusClosed, SubcontractOrderStatusCancelled) {
		t.Fatal("closed should be terminal")
	}
	if CanTransitionSubcontractOrderStatus(SubcontractOrderStatusDraft, SubcontractOrderStatusDraft) {
		t.Fatal("same-status transition should not be accepted")
	}
}

func subcontractOrderReadyForMaterials(t *testing.T, sampleRequired bool) SubcontractOrder {
	t.Helper()

	order, err := subcontractOrderReadyForFactoryConfirmation(t)
	if err != nil {
		t.Fatalf("factory confirmed order: %v", err)
	}
	order.SampleRequired = sampleRequired
	order, err = order.MarkMaterialsIssued("warehouse-user", time.Now())
	if err != nil {
		t.Fatalf("mark materials issued: %v", err)
	}

	return order
}

func subcontractOrderReadyForFactoryConfirmation(t *testing.T) (SubcontractOrder, error) {
	t.Helper()

	order, err := NewSubcontractOrderDocument(validSubcontractOrderInput())
	if err != nil {
		return SubcontractOrder{}, err
	}
	order, err = order.Submit("subcontract-user", time.Now())
	if err != nil {
		return SubcontractOrder{}, err
	}
	order, err = order.Approve("operations-lead", time.Now())
	if err != nil {
		return SubcontractOrder{}, err
	}

	return order.ConfirmFactory("factory-coordinator", time.Now())
}

func validSubcontractOrderInput() NewSubcontractOrderDocumentInput {
	return NewSubcontractOrderDocumentInput{
		ID:                  "sco_001",
		OrgID:               "org-my-pham",
		OrderNo:             " sco-20260429-001 ",
		FactoryID:           "fac_001",
		FactoryCode:         " fac-hcm-01 ",
		FactoryName:         "HCM Cosmetics Factory",
		FinishedItemID:      "item_serum",
		FinishedSKUCode:     " fg-serum-001 ",
		FinishedItemName:    "Brightening Serum",
		PlannedQty:          decimal.MustQuantity("1000"),
		UOMCode:             " pcs ",
		BaseUOMCode:         " pcs ",
		ConversionFactor:    decimal.MustQuantity("1"),
		CurrencyCode:        "VND",
		SpecSummary:         "30ml bottle, printed box, approved label",
		SampleRequired:      true,
		TargetStartDate:     "2026-05-02",
		ExpectedReceiptDate: "2026-05-12",
		CreatedAt:           time.Date(2026, 4, 29, 8, 0, 0, 0, time.UTC),
		CreatedBy:           "subcontract-user",
		MaterialLines: []NewSubcontractMaterialLineInput{
			{
				ID:               "sco_mat_001",
				LineNo:           1,
				ItemID:           "item_base",
				SKUCode:          " rm-base-001 ",
				ItemName:         "Serum Base",
				PlannedQty:       decimal.MustQuantity("10"),
				UOMCode:          " kg ",
				BaseUOMCode:      " g ",
				ConversionFactor: decimal.MustQuantity("1000"),
				UnitCost:         decimal.MustUnitCost("150000"),
				LotTraceRequired: true,
			},
			{
				ID:               "sco_mat_002",
				ItemID:           "item_box",
				SKUCode:          " pk-box-030 ",
				ItemName:         "Printed Box 30ml",
				PlannedQty:       decimal.MustQuantity("1000"),
				UOMCode:          " pcs ",
				BaseUOMCode:      " pcs ",
				ConversionFactor: decimal.MustQuantity("1"),
				UnitCost:         decimal.MustUnitCost("700"),
				LotTraceRequired: false,
			},
		},
	}
}
