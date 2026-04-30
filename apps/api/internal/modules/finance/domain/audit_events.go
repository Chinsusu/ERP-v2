package domain

import (
	"errors"
	"strings"
)

var ErrFinanceInvalidAuditAction = errors.New("finance audit action is invalid")
var ErrFinanceInvalidEntityType = errors.New("finance entity type is invalid")
var ErrFinanceInvalidSourceDocument = errors.New("finance source document is invalid")

type FinanceAuditAction string

const (
	FinanceAuditActionReceivableCreated         FinanceAuditAction = "finance.customer_receivable.created"
	FinanceAuditActionReceivableReceiptRecorded FinanceAuditAction = "finance.customer_receivable.receipt_recorded"
	FinanceAuditActionReceivableDisputed        FinanceAuditAction = "finance.customer_receivable.disputed"
	FinanceAuditActionReceivableVoided          FinanceAuditAction = "finance.customer_receivable.voided"

	FinanceAuditActionPayableCreated          FinanceAuditAction = "finance.supplier_payable.created"
	FinanceAuditActionPayablePaymentRequested FinanceAuditAction = "finance.supplier_payable.payment_requested"
	FinanceAuditActionPayablePaymentApproved  FinanceAuditAction = "finance.supplier_payable.payment_approved"
	FinanceAuditActionPayablePaymentRejected  FinanceAuditAction = "finance.supplier_payable.payment_rejected"
	FinanceAuditActionPayablePaymentRecorded  FinanceAuditAction = "finance.supplier_payable.payment_recorded"
	FinanceAuditActionPayableVoided           FinanceAuditAction = "finance.supplier_payable.voided"

	FinanceAuditActionCODRemittanceCreated             FinanceAuditAction = "finance.cod_remittance.created"
	FinanceAuditActionCODRemittanceMatched             FinanceAuditAction = "finance.cod_remittance.matched"
	FinanceAuditActionCODRemittanceSubmitted           FinanceAuditAction = "finance.cod_remittance.submitted"
	FinanceAuditActionCODRemittanceApproved            FinanceAuditAction = "finance.cod_remittance.approved"
	FinanceAuditActionCODRemittanceDiscrepancyRecorded FinanceAuditAction = "finance.cod_remittance.discrepancy_recorded"
	FinanceAuditActionCODRemittanceClosed              FinanceAuditAction = "finance.cod_remittance.closed"

	FinanceAuditActionCashTransactionRecorded FinanceAuditAction = "finance.cash_transaction.recorded"
	FinanceAuditActionCashTransactionPosted   FinanceAuditAction = "finance.cash_transaction.posted"
	FinanceAuditActionCashTransactionVoided   FinanceAuditAction = "finance.cash_transaction.voided"
)

type FinanceEntityType string

const (
	FinanceEntityTypeCustomerReceivable FinanceEntityType = "finance.customer_receivable"
	FinanceEntityTypeSupplierPayable    FinanceEntityType = "finance.supplier_payable"
	FinanceEntityTypeCODRemittance      FinanceEntityType = "finance.cod_remittance"
	FinanceEntityTypeCashTransaction    FinanceEntityType = "finance.cash_transaction"
	FinanceEntityTypePaymentRequest     FinanceEntityType = "finance.payment_request"
)

type SourceDocumentType string

const (
	SourceDocumentTypeSalesOrder                  SourceDocumentType = "sales_order"
	SourceDocumentTypeShipment                    SourceDocumentType = "shipment"
	SourceDocumentTypeReturnOrder                 SourceDocumentType = "return_order"
	SourceDocumentTypePurchaseOrder               SourceDocumentType = "purchase_order"
	SourceDocumentTypeWarehouseReceipt            SourceDocumentType = "warehouse_receipt"
	SourceDocumentTypeQCInspection                SourceDocumentType = "qc_inspection"
	SourceDocumentTypeSubcontractOrder            SourceDocumentType = "subcontract_order"
	SourceDocumentTypeSubcontractPaymentMilestone SourceDocumentType = "subcontract_payment_milestone"
	SourceDocumentTypeCODRemittance               SourceDocumentType = "cod_remittance"
	SourceDocumentTypeManualAdjustment            SourceDocumentType = "manual_adjustment"
)

type SourceDocumentRef struct {
	Type SourceDocumentType
	ID   string
	No   string
}

func NewSourceDocumentRef(sourceType SourceDocumentType, id string, no string) (SourceDocumentRef, error) {
	ref := SourceDocumentRef{
		Type: NormalizeSourceDocumentType(sourceType),
		ID:   strings.TrimSpace(id),
		No:   strings.TrimSpace(no),
	}
	if !IsValidSourceDocumentType(ref.Type) || (ref.ID == "" && ref.No == "") {
		return SourceDocumentRef{}, ErrFinanceInvalidSourceDocument
	}

	return ref, nil
}

func (r SourceDocumentRef) Metadata() map[string]any {
	metadata := map[string]any{
		"source_type": string(r.Type),
	}
	if r.ID != "" {
		metadata["source_id"] = r.ID
	}
	if r.No != "" {
		metadata["source_no"] = r.No
	}

	return metadata
}

func NormalizeFinanceAuditAction(action FinanceAuditAction) FinanceAuditAction {
	return FinanceAuditAction(normalizeStatus(string(action)))
}

func IsValidFinanceAuditAction(action FinanceAuditAction) bool {
	switch NormalizeFinanceAuditAction(action) {
	case FinanceAuditActionReceivableCreated,
		FinanceAuditActionReceivableReceiptRecorded,
		FinanceAuditActionReceivableDisputed,
		FinanceAuditActionReceivableVoided,
		FinanceAuditActionPayableCreated,
		FinanceAuditActionPayablePaymentRequested,
		FinanceAuditActionPayablePaymentApproved,
		FinanceAuditActionPayablePaymentRejected,
		FinanceAuditActionPayablePaymentRecorded,
		FinanceAuditActionPayableVoided,
		FinanceAuditActionCODRemittanceCreated,
		FinanceAuditActionCODRemittanceMatched,
		FinanceAuditActionCODRemittanceSubmitted,
		FinanceAuditActionCODRemittanceApproved,
		FinanceAuditActionCODRemittanceDiscrepancyRecorded,
		FinanceAuditActionCODRemittanceClosed,
		FinanceAuditActionCashTransactionRecorded,
		FinanceAuditActionCashTransactionPosted,
		FinanceAuditActionCashTransactionVoided:
		return true
	default:
		return false
	}
}

func NormalizeFinanceEntityType(entityType FinanceEntityType) FinanceEntityType {
	return FinanceEntityType(normalizeStatus(string(entityType)))
}

func IsValidFinanceEntityType(entityType FinanceEntityType) bool {
	switch NormalizeFinanceEntityType(entityType) {
	case FinanceEntityTypeCustomerReceivable,
		FinanceEntityTypeSupplierPayable,
		FinanceEntityTypeCODRemittance,
		FinanceEntityTypeCashTransaction,
		FinanceEntityTypePaymentRequest:
		return true
	default:
		return false
	}
}

func NormalizeSourceDocumentType(sourceType SourceDocumentType) SourceDocumentType {
	return SourceDocumentType(normalizeStatus(string(sourceType)))
}

func IsValidSourceDocumentType(sourceType SourceDocumentType) bool {
	switch NormalizeSourceDocumentType(sourceType) {
	case SourceDocumentTypeSalesOrder,
		SourceDocumentTypeShipment,
		SourceDocumentTypeReturnOrder,
		SourceDocumentTypePurchaseOrder,
		SourceDocumentTypeWarehouseReceipt,
		SourceDocumentTypeQCInspection,
		SourceDocumentTypeSubcontractOrder,
		SourceDocumentTypeSubcontractPaymentMilestone,
		SourceDocumentTypeCODRemittance,
		SourceDocumentTypeManualAdjustment:
		return true
	default:
		return false
	}
}
