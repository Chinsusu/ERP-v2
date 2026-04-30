package domain

import (
	"errors"
	"testing"
)

func TestFinanceAuditActionsNormalizeAndValidate(t *testing.T) {
	got := NormalizeFinanceAuditAction(" Finance.Customer_Receivable.Created ")
	if got != FinanceAuditActionReceivableCreated {
		t.Fatalf("action = %q, want %q", got, FinanceAuditActionReceivableCreated)
	}

	validActions := []FinanceAuditAction{
		FinanceAuditActionReceivableReceiptRecorded,
		FinanceAuditActionPayablePaymentApproved,
		FinanceAuditActionCODRemittanceDiscrepancyRecorded,
		FinanceAuditActionCashTransactionPosted,
	}
	for _, action := range validActions {
		if !IsValidFinanceAuditAction(action) {
			t.Fatalf("action %q should be valid", action)
		}
	}
	if IsValidFinanceAuditAction("finance.customer_receivable.deleted") {
		t.Fatal("deleted action should be invalid")
	}
}

func TestFinanceEntityTypesNormalizeAndValidate(t *testing.T) {
	got := NormalizeFinanceEntityType(" Finance.COD_Remittance ")
	if got != FinanceEntityTypeCODRemittance {
		t.Fatalf("entity type = %q, want %q", got, FinanceEntityTypeCODRemittance)
	}

	validTypes := []FinanceEntityType{
		FinanceEntityTypeCustomerReceivable,
		FinanceEntityTypeSupplierPayable,
		FinanceEntityTypeCashTransaction,
		FinanceEntityTypePaymentRequest,
	}
	for _, entityType := range validTypes {
		if !IsValidFinanceEntityType(entityType) {
			t.Fatalf("entity type %q should be valid", entityType)
		}
	}
	if IsValidFinanceEntityType("finance.general_ledger") {
		t.Fatal("general ledger entity type should be invalid in Sprint 6")
	}
}

func TestSourceDocumentRefNormalizesAndBuildsMetadata(t *testing.T) {
	ref, err := NewSourceDocumentRef(" Sales_Order ", " so-260430-0001 ", " SO-260430-0001 ")
	if err != nil {
		t.Fatalf("source ref: %v", err)
	}
	if ref.Type != SourceDocumentTypeSalesOrder {
		t.Fatalf("source type = %q, want %q", ref.Type, SourceDocumentTypeSalesOrder)
	}
	if ref.ID != "so-260430-0001" || ref.No != "SO-260430-0001" {
		t.Fatalf("source ref = %+v, want trimmed id/no", ref)
	}

	metadata := ref.Metadata()
	if metadata["source_type"] != "sales_order" ||
		metadata["source_id"] != "so-260430-0001" ||
		metadata["source_no"] != "SO-260430-0001" {
		t.Fatalf("metadata = %+v, want source fields", metadata)
	}
}

func TestSourceDocumentRefRequiresKnownTypeAndIdentifier(t *testing.T) {
	tests := []struct {
		name       string
		sourceType SourceDocumentType
		id         string
		no         string
	}{
		{name: "unknown type", sourceType: "general_ledger", id: "gl-1"},
		{name: "missing identifiers", sourceType: SourceDocumentTypePurchaseOrder},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewSourceDocumentRef(tt.sourceType, tt.id, tt.no)
			if !errors.Is(err, ErrFinanceInvalidSourceDocument) {
				t.Fatalf("error = %v, want %v", err, ErrFinanceInvalidSourceDocument)
			}
		})
	}
}
