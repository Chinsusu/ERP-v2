package domain

import (
	"errors"
	"testing"
	"time"

	"github.com/Chinsusu/ERP-v2/apps/api/internal/shared/decimal"
)

func TestNewSubcontractMaterialTransferCreatesSentDocument(t *testing.T) {
	transfer, err := NewSubcontractMaterialTransfer(validSubcontractMaterialTransferInput())
	if err != nil {
		t.Fatalf("new subcontract material transfer: %v", err)
	}

	if transfer.Status != SubcontractMaterialTransferStatusSentToFactory {
		t.Fatalf("status = %q, want sent to factory", transfer.Status)
	}
	if transfer.TransferNo != "SMT-20260429-001" || transfer.FactoryCode != "FAC-HCM-01" {
		t.Fatalf("normalized transfer header = %+v, want uppercase codes", transfer)
	}
	if len(transfer.Lines) != 2 {
		t.Fatalf("line count = %d, want 2", len(transfer.Lines))
	}
	if got, want := transfer.Lines[0].BaseIssueQty.String(), "10000.000000"; got != want {
		t.Fatalf("first base issue qty = %s, want %s", got, want)
	}
	if len(transfer.Evidence) != 1 || transfer.Evidence[0].EvidenceType != "handover" {
		t.Fatalf("evidence = %+v, want handover evidence", transfer.Evidence)
	}
}

func TestSubcontractMaterialTransferRequiresBatchForLotTrace(t *testing.T) {
	input := validSubcontractMaterialTransferInput()
	input.Lines[0].BatchID = ""
	input.Lines[0].BatchNo = ""
	input.Lines[0].LotNo = ""

	_, err := NewSubcontractMaterialTransfer(input)
	if !errors.Is(err, ErrSubcontractMaterialTransferBatchRequired) {
		t.Fatalf("error = %v, want batch required", err)
	}
}

func TestSubcontractMaterialTransferRejectsInvalidEvidence(t *testing.T) {
	input := validSubcontractMaterialTransferInput()
	input.Evidence[0].ObjectKey = ""
	input.Evidence[0].ExternalURL = ""

	_, err := NewSubcontractMaterialTransfer(input)
	if !errors.Is(err, ErrSubcontractMaterialTransferRequiredField) {
		t.Fatalf("error = %v, want required field", err)
	}
}

func validSubcontractMaterialTransferInput() NewSubcontractMaterialTransferInput {
	return NewSubcontractMaterialTransferInput{
		ID:                  "smt_001",
		OrgID:               "org-my-pham",
		TransferNo:          " smt-20260429-001 ",
		SubcontractOrderID:  "sco_001",
		SubcontractOrderNo:  "sco-20260429-001",
		FactoryID:           "fac_001",
		FactoryCode:         " fac-hcm-01 ",
		FactoryName:         "HCM Cosmetics Factory",
		SourceWarehouseID:   "wh_main",
		SourceWarehouseCode: "wh-hcm",
		Status:              SubcontractMaterialTransferStatusSentToFactory,
		HandoverBy:          "warehouse-user",
		HandoverAt:          time.Date(2026, 4, 29, 9, 30, 0, 0, time.UTC),
		ReceivedBy:          "factory-receiver",
		ReceiverContact:     "0988000111",
		VehicleNo:           "51A-12345",
		CreatedAt:           time.Date(2026, 4, 29, 9, 0, 0, 0, time.UTC),
		CreatedBy:           "warehouse-user",
		Lines: []NewSubcontractMaterialTransferLineInput{
			{
				ID:                  "smt_line_001",
				LineNo:              1,
				OrderMaterialLineID: "sco_mat_001",
				ItemID:              "item_base",
				SKUCode:             "rm-base-001",
				ItemName:            "Serum Base",
				IssueQty:            decimal.MustQuantity("10"),
				UOMCode:             "kg",
				BaseIssueQty:        decimal.MustQuantity("10000"),
				BaseUOMCode:         "g",
				ConversionFactor:    decimal.MustQuantity("1000"),
				BatchID:             "batch_base_001",
				BatchNo:             "base-lot-001",
				LotTraceRequired:    true,
			},
			{
				ID:                  "smt_line_002",
				LineNo:              2,
				OrderMaterialLineID: "sco_mat_002",
				ItemID:              "item_box",
				SKUCode:             "pk-box-030",
				ItemName:            "Printed Box 30ml",
				IssueQty:            decimal.MustQuantity("1000"),
				UOMCode:             "pcs",
				BaseIssueQty:        decimal.MustQuantity("1000"),
				BaseUOMCode:         "pcs",
				ConversionFactor:    decimal.MustQuantity("1"),
			},
		},
		Evidence: []NewSubcontractMaterialTransferEvidenceInput{
			{
				ID:           "evidence_001",
				EvidenceType: "HANDOVER",
				FileName:     "handover.pdf",
				ObjectKey:    "subcontract/smt_001/handover.pdf",
			},
		},
	}
}
