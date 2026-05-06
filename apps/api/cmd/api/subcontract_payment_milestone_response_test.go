package main

import (
	"testing"

	productionapp "github.com/Chinsusu/ERP-v2/apps/api/internal/modules/production/application"
	productiondomain "github.com/Chinsusu/ERP-v2/apps/api/internal/modules/production/domain"
	"github.com/Chinsusu/ERP-v2/apps/api/internal/shared/decimal"
)

func TestSubcontractPaymentMilestoneResponseIncludesSupplierPayableHandoff(t *testing.T) {
	payload := newSubcontractPaymentMilestoneResultResponse(productionapp.SubcontractPaymentMilestoneResult{
		SubcontractOrder: productiondomain.SubcontractOrder{
			ID:      "sco-s34-handoff",
			OrderNo: "SCO-S34-HANDOFF",
			Status:  productiondomain.SubcontractOrderStatusFinalPaymentReady,
		},
		Milestone: productiondomain.SubcontractPaymentMilestone{
			ID:                 "spm-s34-handoff",
			MilestoneNo:        "SPM-S34-HANDOFF",
			SubcontractOrderID: "sco-s34-handoff",
			SubcontractOrderNo: "SCO-S34-HANDOFF",
			Kind:               productiondomain.SubcontractPaymentMilestoneKindFinalPayment,
			Status:             productiondomain.SubcontractPaymentMilestoneStatusReady,
			Amount:             decimal.MustMoneyAmount("6800000.00"),
			CurrencyCode:       decimal.MustCurrencyCode("VND"),
		},
		PreviousStatus: productiondomain.SubcontractOrderStatusAccepted,
		CurrentStatus:  productiondomain.SubcontractOrderStatusFinalPaymentReady,
		AuditLogID:     "audit-s34-final-payment",
		SupplierPayable: productionapp.SubcontractPayableCreationResult{
			PayableID:  "ap-spm-s34-handoff",
			PayableNo:  "AP-SPM-S34-HANDOFF",
			AuditLogID: "audit-ap-spm-s34-handoff",
		},
	})

	if payload.SupplierPayable == nil ||
		payload.SupplierPayable.PayableID != "ap-spm-s34-handoff" ||
		payload.SupplierPayable.PayableNo != "AP-SPM-S34-HANDOFF" ||
		payload.SupplierPayable.AuditLogID != "audit-ap-spm-s34-handoff" {
		t.Fatalf("supplier payable handoff = %+v, want AP identifiers exposed to production UI", payload.SupplierPayable)
	}
}
