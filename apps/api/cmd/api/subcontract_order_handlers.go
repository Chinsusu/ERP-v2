package main

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"
	"time"

	inventorydomain "github.com/Chinsusu/ERP-v2/apps/api/internal/modules/inventory/domain"
	masterdataapp "github.com/Chinsusu/ERP-v2/apps/api/internal/modules/masterdata/application"
	productionapp "github.com/Chinsusu/ERP-v2/apps/api/internal/modules/production/application"
	productiondomain "github.com/Chinsusu/ERP-v2/apps/api/internal/modules/production/domain"
	"github.com/Chinsusu/ERP-v2/apps/api/internal/shared/auth"
	apperrors "github.com/Chinsusu/ERP-v2/apps/api/internal/shared/errors"
	"github.com/Chinsusu/ERP-v2/apps/api/internal/shared/response"
)

type subcontractOrderUOMConverterAdapter struct {
	catalog uomCatalog
}

type subcontractOrderMaterialLineRequest struct {
	ID               string `json:"id"`
	LineNo           int    `json:"line_no"`
	ItemID           string `json:"item_id"`
	PlannedQty       string `json:"planned_qty"`
	UOMCode          string `json:"uom_code"`
	UnitCost         string `json:"unit_cost"`
	CurrencyCode     string `json:"currency_code"`
	LotTraceRequired bool   `json:"lot_trace_required"`
	Note             string `json:"note"`
}

type createSubcontractOrderRequest struct {
	ID                     string                                `json:"id"`
	OrderNo                string                                `json:"order_no"`
	FactoryID              string                                `json:"factory_id"`
	FinishedItemID         string                                `json:"finished_item_id"`
	PlannedQty             string                                `json:"planned_qty"`
	UOMCode                string                                `json:"uom_code"`
	CurrencyCode           string                                `json:"currency_code"`
	SpecSummary            string                                `json:"spec_summary"`
	SourceProductionPlanID string                                `json:"source_production_plan_id"`
	SourceProductionPlanNo string                                `json:"source_production_plan_no"`
	SampleRequired         bool                                  `json:"sample_required"`
	ClaimWindowDays        int                                   `json:"claim_window_days"`
	TargetStartDate        string                                `json:"target_start_date"`
	ExpectedReceiptDate    string                                `json:"expected_receipt_date"`
	MaterialLines          []subcontractOrderMaterialLineRequest `json:"material_lines"`
}

type updateSubcontractOrderRequest struct {
	FactoryID              string                                `json:"factory_id"`
	FinishedItemID         string                                `json:"finished_item_id"`
	PlannedQty             string                                `json:"planned_qty"`
	UOMCode                string                                `json:"uom_code"`
	SpecSummary            string                                `json:"spec_summary"`
	SourceProductionPlanID string                                `json:"source_production_plan_id"`
	SourceProductionPlanNo string                                `json:"source_production_plan_no"`
	SampleRequired         *bool                                 `json:"sample_required"`
	ClaimWindowDays        int                                   `json:"claim_window_days"`
	TargetStartDate        string                                `json:"target_start_date"`
	ExpectedReceiptDate    string                                `json:"expected_receipt_date"`
	ExpectedVersion        int                                   `json:"expected_version"`
	MaterialLines          []subcontractOrderMaterialLineRequest `json:"material_lines"`
}

type subcontractOrderActionRequest struct {
	ExpectedVersion int    `json:"expected_version"`
	Reason          string `json:"reason"`
}

type recordSubcontractDepositRequest struct {
	ExpectedVersion int    `json:"expected_version"`
	MilestoneID     string `json:"milestone_id"`
	MilestoneNo     string `json:"milestone_no"`
	Amount          string `json:"amount"`
	RecordedBy      string `json:"recorded_by"`
	RecordedAt      string `json:"recorded_at"`
	Note            string `json:"note"`
}

type markSubcontractFinalPaymentReadyRequest struct {
	ExpectedVersion     int    `json:"expected_version"`
	MilestoneID         string `json:"milestone_id"`
	MilestoneNo         string `json:"milestone_no"`
	Amount              string `json:"amount"`
	ReadyBy             string `json:"ready_by"`
	ReadyAt             string `json:"ready_at"`
	ApprovedExceptionID string `json:"approved_exception_id"`
	Note                string `json:"note"`
}

type issueSubcontractMaterialsLineRequest struct {
	ID                  string `json:"id"`
	LineNo              int    `json:"line_no"`
	OrderMaterialLineID string `json:"order_material_line_id"`
	IssueQty            string `json:"issue_qty"`
	UOMCode             string `json:"uom_code"`
	BaseIssueQty        string `json:"base_issue_qty"`
	BaseUOMCode         string `json:"base_uom_code"`
	ConversionFactor    string `json:"conversion_factor"`
	BatchID             string `json:"batch_id"`
	BatchNo             string `json:"batch_no"`
	LotNo               string `json:"lot_no"`
	SourceBinID         string `json:"source_bin_id"`
	Note                string `json:"note"`
}

type issueSubcontractMaterialsEvidenceRequest struct {
	ID           string `json:"id"`
	EvidenceType string `json:"evidence_type"`
	FileName     string `json:"file_name"`
	ObjectKey    string `json:"object_key"`
	ExternalURL  string `json:"external_url"`
	Note         string `json:"note"`
}

type issueSubcontractMaterialsRequest struct {
	ExpectedVersion     int                                        `json:"expected_version"`
	TransferID          string                                     `json:"transfer_id"`
	TransferNo          string                                     `json:"transfer_no"`
	SourceWarehouseID   string                                     `json:"source_warehouse_id"`
	SourceWarehouseCode string                                     `json:"source_warehouse_code"`
	HandoverBy          string                                     `json:"handover_by"`
	HandoverAt          string                                     `json:"handover_at"`
	ReceivedBy          string                                     `json:"received_by"`
	ReceiverContact     string                                     `json:"receiver_contact"`
	VehicleNo           string                                     `json:"vehicle_no"`
	Note                string                                     `json:"note"`
	Lines               []issueSubcontractMaterialsLineRequest     `json:"lines"`
	Evidence            []issueSubcontractMaterialsEvidenceRequest `json:"evidence"`
}

type subcontractSampleEvidenceRequest struct {
	ID           string `json:"id"`
	EvidenceType string `json:"evidence_type"`
	FileName     string `json:"file_name"`
	ObjectKey    string `json:"object_key"`
	ExternalURL  string `json:"external_url"`
	Note         string `json:"note"`
}

type submitSubcontractSampleRequest struct {
	ExpectedVersion  int                                `json:"expected_version"`
	SampleApprovalID string                             `json:"sample_approval_id"`
	SampleCode       string                             `json:"sample_code"`
	FormulaVersion   string                             `json:"formula_version"`
	SpecVersion      string                             `json:"spec_version"`
	SubmittedBy      string                             `json:"submitted_by"`
	SubmittedAt      string                             `json:"submitted_at"`
	Note             string                             `json:"note"`
	Evidence         []subcontractSampleEvidenceRequest `json:"evidence"`
}

type decideSubcontractSampleRequest struct {
	ExpectedVersion  int    `json:"expected_version"`
	SampleApprovalID string `json:"sample_approval_id"`
	DecisionAt       string `json:"decision_at"`
	Reason           string `json:"reason"`
	StorageStatus    string `json:"storage_status"`
}

type createSubcontractFactoryDispatchRequest struct {
	ExpectedVersion int    `json:"expected_version"`
	DispatchID      string `json:"dispatch_id"`
	DispatchNo      string `json:"dispatch_no"`
	Note            string `json:"note"`
}

type subcontractFactoryDispatchEvidenceRequest struct {
	ID           string `json:"id"`
	EvidenceType string `json:"evidence_type"`
	FileName     string `json:"file_name"`
	ObjectKey    string `json:"object_key"`
	ExternalURL  string `json:"external_url"`
	Note         string `json:"note"`
}

type markSubcontractFactoryDispatchSentRequest struct {
	ExpectedVersion int                                         `json:"expected_version"`
	SentBy          string                                      `json:"sent_by"`
	SentAt          string                                      `json:"sent_at"`
	Note            string                                      `json:"note"`
	Evidence        []subcontractFactoryDispatchEvidenceRequest `json:"evidence"`
}

type recordSubcontractFactoryDispatchResponseRequest struct {
	ExpectedVersion int    `json:"expected_version"`
	ResponseStatus  string `json:"response_status"`
	ResponseBy      string `json:"response_by"`
	RespondedAt     string `json:"responded_at"`
	ResponseNote    string `json:"response_note"`
}

type receiveSubcontractFinishedGoodsLineRequest struct {
	ID               string `json:"id"`
	LineNo           int    `json:"line_no"`
	ItemID           string `json:"item_id"`
	SKUCode          string `json:"sku_code"`
	ItemName         string `json:"item_name"`
	BatchID          string `json:"batch_id"`
	BatchNo          string `json:"batch_no"`
	LotNo            string `json:"lot_no"`
	ExpiryDate       string `json:"expiry_date"`
	ReceiveQty       string `json:"receive_qty"`
	UOMCode          string `json:"uom_code"`
	BaseReceiveQty   string `json:"base_receive_qty"`
	BaseUOMCode      string `json:"base_uom_code"`
	ConversionFactor string `json:"conversion_factor"`
	PackagingStatus  string `json:"packaging_status"`
	Note             string `json:"note"`
}

type receiveSubcontractFinishedGoodsEvidenceRequest struct {
	ID           string `json:"id"`
	EvidenceType string `json:"evidence_type"`
	FileName     string `json:"file_name"`
	ObjectKey    string `json:"object_key"`
	ExternalURL  string `json:"external_url"`
	Note         string `json:"note"`
}

type receiveSubcontractFinishedGoodsRequest struct {
	ExpectedVersion int                                              `json:"expected_version"`
	ReceiptID       string                                           `json:"receipt_id"`
	ReceiptNo       string                                           `json:"receipt_no"`
	WarehouseID     string                                           `json:"warehouse_id"`
	WarehouseCode   string                                           `json:"warehouse_code"`
	LocationID      string                                           `json:"location_id"`
	LocationCode    string                                           `json:"location_code"`
	DeliveryNoteNo  string                                           `json:"delivery_note_no"`
	ReceivedBy      string                                           `json:"received_by"`
	ReceivedAt      string                                           `json:"received_at"`
	Note            string                                           `json:"note"`
	Lines           []receiveSubcontractFinishedGoodsLineRequest     `json:"lines"`
	Evidence        []receiveSubcontractFinishedGoodsEvidenceRequest `json:"evidence"`
}

type acceptSubcontractFinishedGoodsRequest struct {
	ExpectedVersion int    `json:"expected_version"`
	AcceptedBy      string `json:"accepted_by"`
	AcceptedAt      string `json:"accepted_at"`
	Note            string `json:"note"`
}

type partialAcceptSubcontractFinishedGoodsRequest struct {
	ExpectedVersion int                                             `json:"expected_version"`
	AcceptedQty     string                                          `json:"accepted_qty"`
	UOMCode         string                                          `json:"uom_code"`
	BaseAcceptedQty string                                          `json:"base_accepted_qty"`
	BaseUOMCode     string                                          `json:"base_uom_code"`
	RejectedQty     string                                          `json:"rejected_qty"`
	BaseRejectedQty string                                          `json:"base_rejected_qty"`
	ClaimID         string                                          `json:"claim_id"`
	ClaimNo         string                                          `json:"claim_no"`
	ReceiptID       string                                          `json:"receipt_id"`
	ReceiptNo       string                                          `json:"receipt_no"`
	ReasonCode      string                                          `json:"reason_code"`
	Reason          string                                          `json:"reason"`
	Severity        string                                          `json:"severity"`
	Evidence        []reportSubcontractFactoryDefectEvidenceRequest `json:"evidence"`
	OwnerID         string                                          `json:"owner_id"`
	AcceptedBy      string                                          `json:"accepted_by"`
	AcceptedAt      string                                          `json:"accepted_at"`
	OpenedBy        string                                          `json:"opened_by"`
	OpenedAt        string                                          `json:"opened_at"`
	Note            string                                          `json:"note"`
}

type reportSubcontractFactoryDefectEvidenceRequest struct {
	ID           string `json:"id"`
	EvidenceType string `json:"evidence_type"`
	FileName     string `json:"file_name"`
	ObjectKey    string `json:"object_key"`
	ExternalURL  string `json:"external_url"`
	Note         string `json:"note"`
}

type reportSubcontractFactoryDefectRequest struct {
	ExpectedVersion int                                             `json:"expected_version"`
	ClaimID         string                                          `json:"claim_id"`
	ClaimNo         string                                          `json:"claim_no"`
	ReceiptID       string                                          `json:"receipt_id"`
	ReceiptNo       string                                          `json:"receipt_no"`
	ReasonCode      string                                          `json:"reason_code"`
	Reason          string                                          `json:"reason"`
	Severity        string                                          `json:"severity"`
	AffectedQty     string                                          `json:"affected_qty"`
	UOMCode         string                                          `json:"uom_code"`
	BaseAffectedQty string                                          `json:"base_affected_qty"`
	BaseUOMCode     string                                          `json:"base_uom_code"`
	Evidence        []reportSubcontractFactoryDefectEvidenceRequest `json:"evidence"`
	OwnerID         string                                          `json:"owner_id"`
	OpenedBy        string                                          `json:"opened_by"`
	OpenedAt        string                                          `json:"opened_at"`
}

type subcontractFactoryClaimDecisionRequest struct {
	ExpectedVersion int    `json:"expected_version"`
	AcknowledgedBy  string `json:"acknowledged_by"`
	AcknowledgedAt  string `json:"acknowledged_at"`
	ResolvedBy      string `json:"resolved_by"`
	ResolvedAt      string `json:"resolved_at"`
	ResolutionNote  string `json:"resolution_note"`
}

type subcontractOrderMaterialLineResponse struct {
	ID               string `json:"id"`
	LineNo           int    `json:"line_no"`
	ItemID           string `json:"item_id"`
	SKUCode          string `json:"sku_code"`
	ItemName         string `json:"item_name"`
	PlannedQty       string `json:"planned_qty"`
	IssuedQty        string `json:"issued_qty"`
	UOMCode          string `json:"uom_code"`
	BasePlannedQty   string `json:"base_planned_qty"`
	BaseIssuedQty    string `json:"base_issued_qty"`
	BaseUOMCode      string `json:"base_uom_code"`
	ConversionFactor string `json:"conversion_factor"`
	UnitCost         string `json:"unit_cost"`
	CurrencyCode     string `json:"currency_code"`
	LineCostAmount   string `json:"line_cost_amount"`
	LotTraceRequired bool   `json:"lot_trace_required"`
	Note             string `json:"note,omitempty"`
}

type subcontractOrderListItemResponse struct {
	ID                     string `json:"id"`
	OrderNo                string `json:"order_no"`
	FactoryID              string `json:"factory_id"`
	FactoryCode            string `json:"factory_code,omitempty"`
	FactoryName            string `json:"factory_name"`
	FinishedItemID         string `json:"finished_item_id"`
	FinishedSKUCode        string `json:"finished_sku_code"`
	FinishedItemName       string `json:"finished_item_name"`
	PlannedQty             string `json:"planned_qty"`
	UOMCode                string `json:"uom_code"`
	ExpectedReceiptDate    string `json:"expected_receipt_date"`
	Status                 string `json:"status"`
	CurrencyCode           string `json:"currency_code"`
	EstimatedCostAmount    string `json:"estimated_cost_amount"`
	ReceivedQty            string `json:"received_qty"`
	AcceptedQty            string `json:"accepted_qty"`
	RejectedQty            string `json:"rejected_qty"`
	SourceProductionPlanID string `json:"source_production_plan_id,omitempty"`
	SourceProductionPlanNo string `json:"source_production_plan_no,omitempty"`
	MaterialLineCount      int    `json:"material_line_count"`
	SampleRequired         bool   `json:"sample_required"`
	CreatedAt              string `json:"created_at"`
	UpdatedAt              string `json:"updated_at"`
	Version                int    `json:"version"`
}

type subcontractOrderResponse struct {
	ID                      string                                 `json:"id"`
	OrgID                   string                                 `json:"org_id"`
	OrderNo                 string                                 `json:"order_no"`
	FactoryID               string                                 `json:"factory_id"`
	FactoryCode             string                                 `json:"factory_code,omitempty"`
	FactoryName             string                                 `json:"factory_name"`
	FinishedItemID          string                                 `json:"finished_item_id"`
	FinishedSKUCode         string                                 `json:"finished_sku_code"`
	FinishedItemName        string                                 `json:"finished_item_name"`
	PlannedQty              string                                 `json:"planned_qty"`
	ReceivedQty             string                                 `json:"received_qty"`
	AcceptedQty             string                                 `json:"accepted_qty"`
	RejectedQty             string                                 `json:"rejected_qty"`
	UOMCode                 string                                 `json:"uom_code"`
	BasePlannedQty          string                                 `json:"base_planned_qty"`
	BaseReceivedQty         string                                 `json:"base_received_qty"`
	BaseAcceptedQty         string                                 `json:"base_accepted_qty"`
	BaseRejectedQty         string                                 `json:"base_rejected_qty"`
	BaseUOMCode             string                                 `json:"base_uom_code"`
	ConversionFactor        string                                 `json:"conversion_factor"`
	CurrencyCode            string                                 `json:"currency_code"`
	EstimatedCostAmount     string                                 `json:"estimated_cost_amount"`
	DepositAmount           string                                 `json:"deposit_amount"`
	SpecSummary             string                                 `json:"spec_summary,omitempty"`
	SourceProductionPlanID  string                                 `json:"source_production_plan_id,omitempty"`
	SourceProductionPlanNo  string                                 `json:"source_production_plan_no,omitempty"`
	SampleRequired          bool                                   `json:"sample_required"`
	ClaimWindowDays         int                                    `json:"claim_window_days"`
	TargetStartDate         string                                 `json:"target_start_date,omitempty"`
	ExpectedReceiptDate     string                                 `json:"expected_receipt_date"`
	Status                  string                                 `json:"status"`
	MaterialLines           []subcontractOrderMaterialLineResponse `json:"material_lines"`
	AuditLogID              string                                 `json:"audit_log_id,omitempty"`
	CreatedAt               string                                 `json:"created_at"`
	UpdatedAt               string                                 `json:"updated_at"`
	SubmittedAt             string                                 `json:"submitted_at,omitempty"`
	ApprovedAt              string                                 `json:"approved_at,omitempty"`
	FactoryConfirmedAt      string                                 `json:"factory_confirmed_at,omitempty"`
	DepositRecordedAt       string                                 `json:"deposit_recorded_at,omitempty"`
	MaterialsIssuedAt       string                                 `json:"materials_issued_at,omitempty"`
	SampleSubmittedAt       string                                 `json:"sample_submitted_at,omitempty"`
	SampleApprovedAt        string                                 `json:"sample_approved_at,omitempty"`
	SampleRejectedAt        string                                 `json:"sample_rejected_at,omitempty"`
	MassProductionStartedAt string                                 `json:"mass_production_started_at,omitempty"`
	FinishedGoodsReceivedAt string                                 `json:"finished_goods_received_at,omitempty"`
	QCStartedAt             string                                 `json:"qc_started_at,omitempty"`
	AcceptedAt              string                                 `json:"accepted_at,omitempty"`
	RejectedFactoryIssueAt  string                                 `json:"rejected_factory_issue_at,omitempty"`
	FinalPaymentReadyAt     string                                 `json:"final_payment_ready_at,omitempty"`
	ClosedAt                string                                 `json:"closed_at,omitempty"`
	CancelledAt             string                                 `json:"cancelled_at,omitempty"`
	CancelReason            string                                 `json:"cancel_reason,omitempty"`
	SampleRejectReason      string                                 `json:"sample_reject_reason,omitempty"`
	FactoryIssueReason      string                                 `json:"factory_issue_reason,omitempty"`
	Version                 int                                    `json:"version"`
}

type subcontractOrderActionResultResponse struct {
	SubcontractOrder subcontractOrderResponse `json:"subcontract_order"`
	PreviousStatus   string                   `json:"previous_status"`
	CurrentStatus    string                   `json:"current_status"`
	AuditLogID       string                   `json:"audit_log_id,omitempty"`
}

type subcontractPaymentMilestoneResponse struct {
	ID                  string `json:"id"`
	MilestoneNo         string `json:"milestone_no"`
	SubcontractOrderID  string `json:"subcontract_order_id"`
	SubcontractOrderNo  string `json:"subcontract_order_no"`
	FactoryID           string `json:"factory_id"`
	FactoryCode         string `json:"factory_code,omitempty"`
	FactoryName         string `json:"factory_name"`
	Kind                string `json:"kind"`
	Status              string `json:"status"`
	Amount              string `json:"amount"`
	CurrencyCode        string `json:"currency_code"`
	Note                string `json:"note,omitempty"`
	ApprovedExceptionID string `json:"approved_exception_id,omitempty"`
	RecordedBy          string `json:"recorded_by,omitempty"`
	RecordedAt          string `json:"recorded_at,omitempty"`
	ReadyBy             string `json:"ready_by,omitempty"`
	ReadyAt             string `json:"ready_at,omitempty"`
	BlockedBy           string `json:"blocked_by,omitempty"`
	BlockedAt           string `json:"blocked_at,omitempty"`
	BlockReason         string `json:"block_reason,omitempty"`
	CreatedAt           string `json:"created_at"`
	UpdatedAt           string `json:"updated_at"`
	Version             int    `json:"version"`
}

type subcontractPaymentMilestoneResultResponse struct {
	SubcontractOrder subcontractOrderResponse            `json:"subcontract_order"`
	Milestone        subcontractPaymentMilestoneResponse `json:"milestone"`
	PreviousStatus   string                              `json:"previous_status"`
	CurrentStatus    string                              `json:"current_status"`
	AuditLogID       string                              `json:"audit_log_id,omitempty"`
}

type subcontractMaterialTransferLineResponse struct {
	ID                  string `json:"id"`
	LineNo              int    `json:"line_no"`
	OrderMaterialLineID string `json:"order_material_line_id"`
	ItemID              string `json:"item_id"`
	SKUCode             string `json:"sku_code"`
	ItemName            string `json:"item_name"`
	IssueQty            string `json:"issue_qty"`
	UOMCode             string `json:"uom_code"`
	BaseIssueQty        string `json:"base_issue_qty"`
	BaseUOMCode         string `json:"base_uom_code"`
	ConversionFactor    string `json:"conversion_factor"`
	BatchID             string `json:"batch_id,omitempty"`
	BatchNo             string `json:"batch_no,omitempty"`
	LotNo               string `json:"lot_no,omitempty"`
	SourceBinID         string `json:"source_bin_id,omitempty"`
	LotTraceRequired    bool   `json:"lot_trace_required"`
	Note                string `json:"note,omitempty"`
}

type subcontractMaterialTransferEvidenceResponse struct {
	ID           string `json:"id"`
	EvidenceType string `json:"evidence_type"`
	FileName     string `json:"file_name,omitempty"`
	ObjectKey    string `json:"object_key,omitempty"`
	ExternalURL  string `json:"external_url,omitempty"`
	Note         string `json:"note,omitempty"`
}

type subcontractMaterialTransferResponse struct {
	ID                  string                                        `json:"id"`
	OrgID               string                                        `json:"org_id"`
	TransferNo          string                                        `json:"transfer_no"`
	SubcontractOrderID  string                                        `json:"subcontract_order_id"`
	SubcontractOrderNo  string                                        `json:"subcontract_order_no"`
	FactoryID           string                                        `json:"factory_id"`
	FactoryCode         string                                        `json:"factory_code,omitempty"`
	FactoryName         string                                        `json:"factory_name"`
	SourceWarehouseID   string                                        `json:"source_warehouse_id"`
	SourceWarehouseCode string                                        `json:"source_warehouse_code,omitempty"`
	Status              string                                        `json:"status"`
	Lines               []subcontractMaterialTransferLineResponse     `json:"lines"`
	Evidence            []subcontractMaterialTransferEvidenceResponse `json:"evidence,omitempty"`
	HandoverBy          string                                        `json:"handover_by"`
	HandoverAt          string                                        `json:"handover_at"`
	ReceivedBy          string                                        `json:"received_by"`
	ReceiverContact     string                                        `json:"receiver_contact,omitempty"`
	VehicleNo           string                                        `json:"vehicle_no,omitempty"`
	Note                string                                        `json:"note,omitempty"`
	CreatedAt           string                                        `json:"created_at"`
	UpdatedAt           string                                        `json:"updated_at"`
	Version             int                                           `json:"version"`
}

type subcontractMaterialIssueMovementResponse struct {
	MovementNo       string `json:"movement_no"`
	MovementType     string `json:"movement_type"`
	ItemID           string `json:"item_id"`
	BatchID          string `json:"batch_id,omitempty"`
	WarehouseID      string `json:"warehouse_id"`
	BinID            string `json:"bin_id,omitempty"`
	Quantity         string `json:"quantity"`
	BaseUOMCode      string `json:"base_uom_code"`
	SourceQuantity   string `json:"source_quantity"`
	SourceUOMCode    string `json:"source_uom_code"`
	ConversionFactor string `json:"conversion_factor"`
	StockStatus      string `json:"stock_status"`
	SourceDocType    string `json:"source_doc_type"`
	SourceDocID      string `json:"source_doc_id"`
	SourceDocLineID  string `json:"source_doc_line_id"`
	Reason           string `json:"reason"`
}

type issueSubcontractMaterialsResponse struct {
	SubcontractOrder subcontractOrderResponse                   `json:"subcontract_order"`
	Transfer         subcontractMaterialTransferResponse        `json:"transfer"`
	StockMovements   []subcontractMaterialIssueMovementResponse `json:"stock_movements"`
	AuditLogID       string                                     `json:"audit_log_id,omitempty"`
}

type subcontractFinishedGoodsReceiptLineResponse struct {
	ID               string `json:"id"`
	LineNo           int    `json:"line_no"`
	ItemID           string `json:"item_id"`
	SKUCode          string `json:"sku_code"`
	ItemName         string `json:"item_name"`
	BatchID          string `json:"batch_id"`
	BatchNo          string `json:"batch_no"`
	LotNo            string `json:"lot_no"`
	ExpiryDate       string `json:"expiry_date"`
	ReceiveQty       string `json:"receive_qty"`
	UOMCode          string `json:"uom_code"`
	BaseReceiveQty   string `json:"base_receive_qty"`
	BaseUOMCode      string `json:"base_uom_code"`
	ConversionFactor string `json:"conversion_factor"`
	PackagingStatus  string `json:"packaging_status,omitempty"`
	Note             string `json:"note,omitempty"`
}

type subcontractFinishedGoodsReceiptEvidenceResponse struct {
	ID           string `json:"id"`
	EvidenceType string `json:"evidence_type"`
	FileName     string `json:"file_name,omitempty"`
	ObjectKey    string `json:"object_key,omitempty"`
	ExternalURL  string `json:"external_url,omitempty"`
	Note         string `json:"note,omitempty"`
}

type subcontractFinishedGoodsReceiptResponse struct {
	ID                 string                                            `json:"id"`
	OrgID              string                                            `json:"org_id"`
	ReceiptNo          string                                            `json:"receipt_no"`
	SubcontractOrderID string                                            `json:"subcontract_order_id"`
	SubcontractOrderNo string                                            `json:"subcontract_order_no"`
	FactoryID          string                                            `json:"factory_id"`
	FactoryCode        string                                            `json:"factory_code,omitempty"`
	FactoryName        string                                            `json:"factory_name"`
	WarehouseID        string                                            `json:"warehouse_id"`
	WarehouseCode      string                                            `json:"warehouse_code,omitempty"`
	LocationID         string                                            `json:"location_id"`
	LocationCode       string                                            `json:"location_code,omitempty"`
	DeliveryNoteNo     string                                            `json:"delivery_note_no"`
	Status             string                                            `json:"status"`
	Lines              []subcontractFinishedGoodsReceiptLineResponse     `json:"lines"`
	Evidence           []subcontractFinishedGoodsReceiptEvidenceResponse `json:"evidence,omitempty"`
	ReceivedBy         string                                            `json:"received_by"`
	ReceivedAt         string                                            `json:"received_at"`
	Note               string                                            `json:"note,omitempty"`
	CreatedAt          string                                            `json:"created_at"`
	UpdatedAt          string                                            `json:"updated_at"`
	Version            int                                               `json:"version"`
}

type receiveSubcontractFinishedGoodsResponse struct {
	SubcontractOrder subcontractOrderResponse                   `json:"subcontract_order"`
	Receipt          subcontractFinishedGoodsReceiptResponse    `json:"receipt"`
	StockMovements   []subcontractMaterialIssueMovementResponse `json:"stock_movements"`
	AuditLogID       string                                     `json:"audit_log_id,omitempty"`
}

type acceptSubcontractFinishedGoodsResponse struct {
	SubcontractOrder subcontractOrderResponse                   `json:"subcontract_order"`
	StockMovements   []subcontractMaterialIssueMovementResponse `json:"stock_movements"`
	PreviousStatus   string                                     `json:"previous_status"`
	CurrentStatus    string                                     `json:"current_status"`
	AuditLogID       string                                     `json:"audit_log_id,omitempty"`
}

type partialAcceptSubcontractFinishedGoodsResponse struct {
	SubcontractOrder subcontractOrderResponse                   `json:"subcontract_order"`
	Claim            subcontractFactoryClaimResponse            `json:"claim"`
	StockMovements   []subcontractMaterialIssueMovementResponse `json:"stock_movements"`
	PreviousStatus   string                                     `json:"previous_status"`
	CurrentStatus    string                                     `json:"current_status"`
	AcceptAuditLogID string                                     `json:"accept_audit_log_id,omitempty"`
	ClaimAuditLogID  string                                     `json:"claim_audit_log_id,omitempty"`
}

type subcontractFactoryClaimEvidenceResponse struct {
	ID           string `json:"id"`
	EvidenceType string `json:"evidence_type"`
	FileName     string `json:"file_name,omitempty"`
	ObjectKey    string `json:"object_key,omitempty"`
	ExternalURL  string `json:"external_url,omitempty"`
	Note         string `json:"note,omitempty"`
}

type subcontractFactoryClaimResponse struct {
	ID                 string                                    `json:"id"`
	OrgID              string                                    `json:"org_id"`
	ClaimNo            string                                    `json:"claim_no"`
	SubcontractOrderID string                                    `json:"subcontract_order_id"`
	SubcontractOrderNo string                                    `json:"subcontract_order_no"`
	FactoryID          string                                    `json:"factory_id"`
	FactoryCode        string                                    `json:"factory_code,omitempty"`
	FactoryName        string                                    `json:"factory_name"`
	ReceiptID          string                                    `json:"receipt_id"`
	ReceiptNo          string                                    `json:"receipt_no"`
	ReasonCode         string                                    `json:"reason_code"`
	Reason             string                                    `json:"reason"`
	Severity           string                                    `json:"severity"`
	Status             string                                    `json:"status"`
	AffectedQty        string                                    `json:"affected_qty"`
	UOMCode            string                                    `json:"uom_code"`
	BaseAffectedQty    string                                    `json:"base_affected_qty"`
	BaseUOMCode        string                                    `json:"base_uom_code"`
	Evidence           []subcontractFactoryClaimEvidenceResponse `json:"evidence,omitempty"`
	OwnerID            string                                    `json:"owner_id,omitempty"`
	OpenedBy           string                                    `json:"opened_by"`
	OpenedAt           string                                    `json:"opened_at"`
	DueAt              string                                    `json:"due_at"`
	AcknowledgedBy     string                                    `json:"acknowledged_by,omitempty"`
	AcknowledgedAt     string                                    `json:"acknowledged_at,omitempty"`
	ResolvedBy         string                                    `json:"resolved_by,omitempty"`
	ResolvedAt         string                                    `json:"resolved_at,omitempty"`
	ResolutionNote     string                                    `json:"resolution_note,omitempty"`
	BlocksFinalPayment bool                                      `json:"blocks_final_payment"`
	CreatedAt          string                                    `json:"created_at"`
	UpdatedAt          string                                    `json:"updated_at"`
	Version            int                                       `json:"version"`
}

type reportSubcontractFactoryDefectResponse struct {
	SubcontractOrder subcontractOrderResponse        `json:"subcontract_order"`
	Claim            subcontractFactoryClaimResponse `json:"claim"`
	PreviousStatus   string                          `json:"previous_status"`
	CurrentStatus    string                          `json:"current_status"`
	AuditLogID       string                          `json:"audit_log_id,omitempty"`
}

type subcontractFactoryClaimDecisionResponse struct {
	Claim      subcontractFactoryClaimResponse `json:"claim"`
	AuditLogID string                          `json:"audit_log_id,omitempty"`
}

type subcontractSampleEvidenceResponse struct {
	ID           string `json:"id"`
	EvidenceType string `json:"evidence_type"`
	FileName     string `json:"file_name,omitempty"`
	ObjectKey    string `json:"object_key,omitempty"`
	ExternalURL  string `json:"external_url,omitempty"`
	Note         string `json:"note,omitempty"`
	CreatedAt    string `json:"created_at"`
	CreatedBy    string `json:"created_by"`
}

type subcontractSampleApprovalResponse struct {
	ID                 string                              `json:"id"`
	OrgID              string                              `json:"org_id"`
	SubcontractOrderID string                              `json:"subcontract_order_id"`
	SubcontractOrderNo string                              `json:"subcontract_order_no"`
	SampleCode         string                              `json:"sample_code"`
	FormulaVersion     string                              `json:"formula_version,omitempty"`
	SpecVersion        string                              `json:"spec_version,omitempty"`
	Status             string                              `json:"status"`
	Evidence           []subcontractSampleEvidenceResponse `json:"evidence"`
	SubmittedBy        string                              `json:"submitted_by"`
	SubmittedAt        string                              `json:"submitted_at"`
	DecisionBy         string                              `json:"decision_by,omitempty"`
	DecisionAt         string                              `json:"decision_at,omitempty"`
	DecisionReason     string                              `json:"decision_reason,omitempty"`
	StorageStatus      string                              `json:"storage_status,omitempty"`
	Note               string                              `json:"note,omitempty"`
	CreatedAt          string                              `json:"created_at"`
	UpdatedAt          string                              `json:"updated_at"`
	Version            int                                 `json:"version"`
}

type subcontractSampleApprovalResultResponse struct {
	SubcontractOrder subcontractOrderResponse          `json:"subcontract_order"`
	SampleApproval   subcontractSampleApprovalResponse `json:"sample_approval"`
	PreviousStatus   string                            `json:"previous_status"`
	CurrentStatus    string                            `json:"current_status"`
	AuditLogID       string                            `json:"audit_log_id,omitempty"`
}

type subcontractFactoryDispatchLineResponse struct {
	ID                  string `json:"id"`
	LineNo              int    `json:"line_no"`
	OrderMaterialLineID string `json:"order_material_line_id"`
	ItemID              string `json:"item_id"`
	SKUCode             string `json:"sku_code"`
	ItemName            string `json:"item_name"`
	PlannedQty          string `json:"planned_qty"`
	UOMCode             string `json:"uom_code"`
	LotTraceRequired    bool   `json:"lot_trace_required"`
	Note                string `json:"note,omitempty"`
}

type subcontractFactoryDispatchEvidenceResponse struct {
	ID           string `json:"id"`
	EvidenceType string `json:"evidence_type"`
	FileName     string `json:"file_name,omitempty"`
	ObjectKey    string `json:"object_key,omitempty"`
	ExternalURL  string `json:"external_url,omitempty"`
	Note         string `json:"note,omitempty"`
}

type subcontractFactoryDispatchResponse struct {
	ID                     string                                       `json:"id"`
	OrgID                  string                                       `json:"org_id"`
	DispatchNo             string                                       `json:"dispatch_no"`
	SubcontractOrderID     string                                       `json:"subcontract_order_id"`
	SubcontractOrderNo     string                                       `json:"subcontract_order_no"`
	SourceProductionPlanID string                                       `json:"source_production_plan_id,omitempty"`
	SourceProductionPlanNo string                                       `json:"source_production_plan_no,omitempty"`
	FactoryID              string                                       `json:"factory_id"`
	FactoryCode            string                                       `json:"factory_code,omitempty"`
	FactoryName            string                                       `json:"factory_name"`
	FinishedItemID         string                                       `json:"finished_item_id"`
	FinishedSKUCode        string                                       `json:"finished_sku_code"`
	FinishedItemName       string                                       `json:"finished_item_name"`
	PlannedQty             string                                       `json:"planned_qty"`
	UOMCode                string                                       `json:"uom_code"`
	SpecSummary            string                                       `json:"spec_summary,omitempty"`
	SampleRequired         bool                                         `json:"sample_required"`
	TargetStartDate        string                                       `json:"target_start_date,omitempty"`
	ExpectedReceiptDate    string                                       `json:"expected_receipt_date,omitempty"`
	Status                 string                                       `json:"status"`
	Lines                  []subcontractFactoryDispatchLineResponse     `json:"lines"`
	Evidence               []subcontractFactoryDispatchEvidenceResponse `json:"evidence,omitempty"`
	ReadyAt                string                                       `json:"ready_at,omitempty"`
	ReadyBy                string                                       `json:"ready_by,omitempty"`
	SentAt                 string                                       `json:"sent_at,omitempty"`
	SentBy                 string                                       `json:"sent_by,omitempty"`
	RespondedAt            string                                       `json:"responded_at,omitempty"`
	ResponseBy             string                                       `json:"response_by,omitempty"`
	FactoryResponseNote    string                                       `json:"factory_response_note,omitempty"`
	Note                   string                                       `json:"note,omitempty"`
	CreatedAt              string                                       `json:"created_at"`
	UpdatedAt              string                                       `json:"updated_at"`
	Version                int                                          `json:"version"`
}

type subcontractFactoryDispatchResultResponse struct {
	SubcontractOrder subcontractOrderResponse           `json:"subcontract_order"`
	Dispatch         subcontractFactoryDispatchResponse `json:"factory_dispatch"`
	AuditLogID       string                             `json:"audit_log_id,omitempty"`
}

func (a subcontractOrderUOMConverterAdapter) ConvertToBase(
	ctx context.Context,
	input productionapp.ConvertSubcontractOrderLineToBaseInput,
) (productionapp.ConvertSubcontractOrderLineToBaseResult, error) {
	result, err := a.catalog.ConvertToBase(ctx, masterdataapp.ConvertToBaseInput{
		ItemID:      input.ItemID,
		SKU:         input.SKU,
		Quantity:    input.Quantity,
		FromUOMCode: input.FromUOMCode,
		BaseUOMCode: input.BaseUOMCode,
	})
	if err != nil {
		return productionapp.ConvertSubcontractOrderLineToBaseResult{}, err
	}

	return productionapp.ConvertSubcontractOrderLineToBaseResult{
		Quantity:         result.Quantity,
		SourceUOMCode:    result.SourceUOMCode,
		BaseQuantity:     result.BaseQuantity,
		BaseUOMCode:      result.BaseUOMCode,
		ConversionFactor: result.ConversionFactor,
	}, nil
}

func subcontractOrdersHandler(service productionapp.SubcontractOrderService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		principal, ok := auth.PrincipalFromContext(r.Context())
		if !ok {
			response.WriteError(w, r, http.StatusUnauthorized, response.ErrorCodeUnauthorized, "Authentication required", nil)
			return
		}
		if !auth.HasPermission(principal, auth.PermissionSubcontractView) {
			writePermissionDenied(w, r, auth.PermissionSubcontractView)
			return
		}

		switch r.Method {
		case http.MethodGet:
			orders, err := service.ListSubcontractOrders(r.Context(), subcontractOrderFilterFromRequest(r))
			if err != nil {
				writeSubcontractOrderError(w, r, err)
				return
			}

			payload := make([]subcontractOrderListItemResponse, 0, len(orders))
			for _, order := range orders {
				payload = append(payload, newSubcontractOrderListItemResponse(order))
			}
			response.WriteSuccess(w, r, http.StatusOK, payload)
		case http.MethodPost:
			if !auth.HasPermission(principal, auth.PermissionRecordCreate) {
				writePermissionDenied(w, r, auth.PermissionRecordCreate)
				return
			}
			r = requestWithStableID(r)
			var payload createSubcontractOrderRequest
			if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
				response.WriteError(w, r, http.StatusBadRequest, response.ErrorCodeValidation, "Invalid subcontract order payload", nil)
				return
			}

			result, err := service.CreateSubcontractOrder(r.Context(), productionapp.CreateSubcontractOrderInput{
				ID:                     payload.ID,
				OrderNo:                payload.OrderNo,
				FactoryID:              payload.FactoryID,
				FinishedItemID:         payload.FinishedItemID,
				PlannedQty:             payload.PlannedQty,
				UOMCode:                payload.UOMCode,
				CurrencyCode:           payload.CurrencyCode,
				SpecSummary:            payload.SpecSummary,
				SourceProductionPlanID: payload.SourceProductionPlanID,
				SourceProductionPlanNo: payload.SourceProductionPlanNo,
				SampleRequired:         payload.SampleRequired,
				ClaimWindowDays:        payload.ClaimWindowDays,
				TargetStartDate:        payload.TargetStartDate,
				ExpectedReceiptDate:    payload.ExpectedReceiptDate,
				MaterialLines:          subcontractOrderMaterialLineInputs(payload.MaterialLines),
				ActorID:                principal.UserID,
				RequestID:              response.RequestID(r),
			})
			if err != nil {
				writeSubcontractOrderError(w, r, err)
				return
			}

			response.WriteSuccess(w, r, http.StatusCreated, newSubcontractOrderResponse(result.SubcontractOrder, result.AuditLogID))
		default:
			response.WriteError(w, r, http.StatusMethodNotAllowed, response.ErrorCodeNotFound, "Route not found", nil)
		}
	}
}

func subcontractOrderDetailHandler(service productionapp.SubcontractOrderService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		principal, ok := auth.PrincipalFromContext(r.Context())
		if !ok {
			response.WriteError(w, r, http.StatusUnauthorized, response.ErrorCodeUnauthorized, "Authentication required", nil)
			return
		}
		if !auth.HasPermission(principal, auth.PermissionSubcontractView) {
			writePermissionDenied(w, r, auth.PermissionSubcontractView)
			return
		}

		switch r.Method {
		case http.MethodGet:
			order, err := service.GetSubcontractOrder(r.Context(), r.PathValue("subcontract_order_id"))
			if err != nil {
				writeSubcontractOrderError(w, r, err)
				return
			}
			response.WriteSuccess(w, r, http.StatusOK, newSubcontractOrderResponse(order, ""))
		case http.MethodPatch:
			if !auth.HasPermission(principal, auth.PermissionRecordCreate) {
				writePermissionDenied(w, r, auth.PermissionRecordCreate)
				return
			}
			r = requestWithStableID(r)
			var payload updateSubcontractOrderRequest
			if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
				response.WriteError(w, r, http.StatusBadRequest, response.ErrorCodeValidation, "Invalid subcontract order payload", nil)
				return
			}

			result, err := service.UpdateSubcontractOrder(r.Context(), productionapp.UpdateSubcontractOrderInput{
				ID:                     r.PathValue("subcontract_order_id"),
				FactoryID:              payload.FactoryID,
				FinishedItemID:         payload.FinishedItemID,
				PlannedQty:             payload.PlannedQty,
				UOMCode:                payload.UOMCode,
				SpecSummary:            payload.SpecSummary,
				SourceProductionPlanID: payload.SourceProductionPlanID,
				SourceProductionPlanNo: payload.SourceProductionPlanNo,
				SampleRequired:         payload.SampleRequired,
				ClaimWindowDays:        payload.ClaimWindowDays,
				TargetStartDate:        payload.TargetStartDate,
				ExpectedReceiptDate:    payload.ExpectedReceiptDate,
				MaterialLines:          subcontractOrderMaterialLineInputs(payload.MaterialLines),
				ExpectedVersion:        payload.ExpectedVersion,
				ActorID:                principal.UserID,
				RequestID:              response.RequestID(r),
			})
			if err != nil {
				writeSubcontractOrderError(w, r, err)
				return
			}
			response.WriteSuccess(w, r, http.StatusOK, newSubcontractOrderResponse(result.SubcontractOrder, result.AuditLogID))
		default:
			response.WriteError(w, r, http.StatusMethodNotAllowed, response.ErrorCodeNotFound, "Route not found", nil)
		}
	}
}

func subcontractOrderSubmitHandler(service productionapp.SubcontractOrderService) http.HandlerFunc {
	return subcontractOrderActionHandler(service, "submit")
}

func subcontractOrderApproveHandler(service productionapp.SubcontractOrderService) http.HandlerFunc {
	return subcontractOrderActionHandler(service, "approve")
}

func subcontractOrderConfirmFactoryHandler(service productionapp.SubcontractOrderService) http.HandlerFunc {
	return subcontractOrderActionHandler(service, "confirm-factory")
}

func subcontractOrderStartMassProductionHandler(service productionapp.SubcontractOrderService) http.HandlerFunc {
	return subcontractOrderActionHandler(service, "start-mass-production")
}

func subcontractOrderCancelHandler(service productionapp.SubcontractOrderService) http.HandlerFunc {
	return subcontractOrderActionHandler(service, "cancel")
}

func subcontractOrderCloseHandler(service productionapp.SubcontractOrderService) http.HandlerFunc {
	return subcontractOrderActionHandler(service, "close")
}

func subcontractOrderRecordDepositHandler(service productionapp.SubcontractOrderService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			response.WriteError(w, r, http.StatusMethodNotAllowed, response.ErrorCodeNotFound, "Route not found", nil)
			return
		}
		principal, ok := auth.PrincipalFromContext(r.Context())
		if !ok {
			response.WriteError(w, r, http.StatusUnauthorized, response.ErrorCodeUnauthorized, "Authentication required", nil)
			return
		}
		if !auth.HasPermission(principal, auth.PermissionSubcontractView) {
			writePermissionDenied(w, r, auth.PermissionSubcontractView)
			return
		}
		if !auth.HasPermission(principal, auth.PermissionRecordCreate) {
			writePermissionDenied(w, r, auth.PermissionRecordCreate)
			return
		}

		r = requestWithStableID(r)
		var payload recordSubcontractDepositRequest
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			response.WriteError(w, r, http.StatusBadRequest, response.ErrorCodeValidation, "Invalid subcontract deposit payload", nil)
			return
		}
		recordedAt, err := parseSubcontractOptionalTime(payload.RecordedAt)
		if err != nil {
			response.WriteError(w, r, http.StatusBadRequest, response.ErrorCodeValidation, "Invalid subcontract deposit payload", map[string]any{"field": "recorded_at"})
			return
		}

		result, err := service.RecordSubcontractDeposit(r.Context(), productionapp.RecordSubcontractDepositInput{
			ID:              r.PathValue("subcontract_order_id"),
			ExpectedVersion: payload.ExpectedVersion,
			MilestoneID:     payload.MilestoneID,
			MilestoneNo:     payload.MilestoneNo,
			Amount:          payload.Amount,
			RecordedBy:      payload.RecordedBy,
			RecordedAt:      recordedAt,
			Note:            payload.Note,
			ActorID:         principal.UserID,
			RequestID:       response.RequestID(r),
		})
		if err != nil {
			writeSubcontractOrderError(w, r, err)
			return
		}

		response.WriteSuccess(w, r, http.StatusOK, newSubcontractPaymentMilestoneResultResponse(result))
	}
}

func subcontractOrderMarkFinalPaymentReadyHandler(service productionapp.SubcontractOrderService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			response.WriteError(w, r, http.StatusMethodNotAllowed, response.ErrorCodeNotFound, "Route not found", nil)
			return
		}
		principal, ok := auth.PrincipalFromContext(r.Context())
		if !ok {
			response.WriteError(w, r, http.StatusUnauthorized, response.ErrorCodeUnauthorized, "Authentication required", nil)
			return
		}
		if !auth.HasPermission(principal, auth.PermissionSubcontractView) {
			writePermissionDenied(w, r, auth.PermissionSubcontractView)
			return
		}
		if !auth.HasPermission(principal, auth.PermissionRecordCreate) {
			writePermissionDenied(w, r, auth.PermissionRecordCreate)
			return
		}

		r = requestWithStableID(r)
		var payload markSubcontractFinalPaymentReadyRequest
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			response.WriteError(w, r, http.StatusBadRequest, response.ErrorCodeValidation, "Invalid subcontract final payment payload", nil)
			return
		}
		readyAt, err := parseSubcontractOptionalTime(payload.ReadyAt)
		if err != nil {
			response.WriteError(w, r, http.StatusBadRequest, response.ErrorCodeValidation, "Invalid subcontract final payment payload", map[string]any{"field": "ready_at"})
			return
		}

		result, err := service.MarkSubcontractFinalPaymentReady(r.Context(), productionapp.MarkSubcontractFinalPaymentReadyInput{
			ID:                  r.PathValue("subcontract_order_id"),
			ExpectedVersion:     payload.ExpectedVersion,
			MilestoneID:         payload.MilestoneID,
			MilestoneNo:         payload.MilestoneNo,
			Amount:              payload.Amount,
			ReadyBy:             payload.ReadyBy,
			ReadyAt:             readyAt,
			ApprovedExceptionID: payload.ApprovedExceptionID,
			Note:                payload.Note,
			ActorID:             principal.UserID,
			RequestID:           response.RequestID(r),
		})
		if err != nil {
			writeSubcontractOrderError(w, r, err)
			return
		}

		response.WriteSuccess(w, r, http.StatusOK, newSubcontractPaymentMilestoneResultResponse(result))
	}
}

func subcontractOrderIssueMaterialsHandler(service productionapp.SubcontractOrderService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			response.WriteError(w, r, http.StatusMethodNotAllowed, response.ErrorCodeNotFound, "Route not found", nil)
			return
		}
		principal, ok := auth.PrincipalFromContext(r.Context())
		if !ok {
			response.WriteError(w, r, http.StatusUnauthorized, response.ErrorCodeUnauthorized, "Authentication required", nil)
			return
		}
		if !auth.HasPermission(principal, auth.PermissionSubcontractView) {
			writePermissionDenied(w, r, auth.PermissionSubcontractView)
			return
		}
		if !auth.HasPermission(principal, auth.PermissionRecordCreate) {
			writePermissionDenied(w, r, auth.PermissionRecordCreate)
			return
		}

		r = requestWithStableID(r)
		var payload issueSubcontractMaterialsRequest
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			response.WriteError(w, r, http.StatusBadRequest, response.ErrorCodeValidation, "Invalid subcontract material issue payload", nil)
			return
		}
		handoverAt, err := parseSubcontractOptionalTime(payload.HandoverAt)
		if err != nil {
			response.WriteError(w, r, http.StatusBadRequest, response.ErrorCodeValidation, "Invalid subcontract material issue payload", map[string]any{"field": "handover_at"})
			return
		}

		result, err := service.IssueSubcontractMaterials(r.Context(), productionapp.IssueSubcontractMaterialsInput{
			ID:                  r.PathValue("subcontract_order_id"),
			ExpectedVersion:     payload.ExpectedVersion,
			TransferID:          payload.TransferID,
			TransferNo:          payload.TransferNo,
			SourceWarehouseID:   payload.SourceWarehouseID,
			SourceWarehouseCode: payload.SourceWarehouseCode,
			Lines:               issueSubcontractMaterialLineInputs(payload.Lines),
			Evidence:            issueSubcontractMaterialEvidenceInputs(payload.Evidence),
			HandoverBy:          payload.HandoverBy,
			HandoverAt:          handoverAt,
			ReceivedBy:          payload.ReceivedBy,
			ReceiverContact:     payload.ReceiverContact,
			VehicleNo:           payload.VehicleNo,
			Note:                payload.Note,
			ActorID:             principal.UserID,
			RequestID:           response.RequestID(r),
		})
		if err != nil {
			writeSubcontractOrderError(w, r, err)
			return
		}

		response.WriteSuccess(w, r, http.StatusOK, issueSubcontractMaterialsResponse{
			SubcontractOrder: newSubcontractOrderResponse(result.SubcontractOrder, ""),
			Transfer:         newSubcontractMaterialTransferResponse(result.Transfer),
			StockMovements:   newSubcontractMaterialIssueMovementResponses(result.StockMovements),
			AuditLogID:       result.AuditLogID,
		})
	}
}

func subcontractOrderReceiveFinishedGoodsHandler(service productionapp.SubcontractOrderService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			response.WriteError(w, r, http.StatusMethodNotAllowed, response.ErrorCodeNotFound, "Route not found", nil)
			return
		}
		principal, ok := auth.PrincipalFromContext(r.Context())
		if !ok {
			response.WriteError(w, r, http.StatusUnauthorized, response.ErrorCodeUnauthorized, "Authentication required", nil)
			return
		}
		if !auth.HasPermission(principal, auth.PermissionSubcontractView) {
			writePermissionDenied(w, r, auth.PermissionSubcontractView)
			return
		}
		if !auth.HasPermission(principal, auth.PermissionRecordCreate) {
			writePermissionDenied(w, r, auth.PermissionRecordCreate)
			return
		}

		r = requestWithStableID(r)
		var payload receiveSubcontractFinishedGoodsRequest
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			response.WriteError(w, r, http.StatusBadRequest, response.ErrorCodeValidation, "Invalid subcontract finished goods receipt payload", nil)
			return
		}
		receivedAt, err := parseSubcontractOptionalTime(payload.ReceivedAt)
		if err != nil {
			response.WriteError(w, r, http.StatusBadRequest, response.ErrorCodeValidation, "Invalid subcontract finished goods receipt payload", map[string]any{"field": "received_at"})
			return
		}

		result, err := service.ReceiveSubcontractFinishedGoods(r.Context(), productionapp.ReceiveSubcontractFinishedGoodsInput{
			ID:              r.PathValue("subcontract_order_id"),
			ExpectedVersion: payload.ExpectedVersion,
			ReceiptID:       payload.ReceiptID,
			ReceiptNo:       payload.ReceiptNo,
			WarehouseID:     payload.WarehouseID,
			WarehouseCode:   payload.WarehouseCode,
			LocationID:      payload.LocationID,
			LocationCode:    payload.LocationCode,
			DeliveryNoteNo:  payload.DeliveryNoteNo,
			Lines:           receiveSubcontractFinishedGoodsLineInputs(payload.Lines),
			Evidence:        receiveSubcontractFinishedGoodsEvidenceInputs(payload.Evidence),
			ReceivedBy:      payload.ReceivedBy,
			ReceivedAt:      receivedAt,
			Note:            payload.Note,
			ActorID:         principal.UserID,
			RequestID:       response.RequestID(r),
		})
		if err != nil {
			writeSubcontractOrderError(w, r, err)
			return
		}

		response.WriteSuccess(w, r, http.StatusOK, receiveSubcontractFinishedGoodsResponse{
			SubcontractOrder: newSubcontractOrderResponse(result.SubcontractOrder, ""),
			Receipt:          newSubcontractFinishedGoodsReceiptResponse(result.Receipt),
			StockMovements:   newSubcontractMaterialIssueMovementResponses(result.StockMovements),
			AuditLogID:       result.AuditLogID,
		})
	}
}

func subcontractOrderAcceptFinishedGoodsHandler(service productionapp.SubcontractOrderService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			response.WriteError(w, r, http.StatusMethodNotAllowed, response.ErrorCodeNotFound, "Route not found", nil)
			return
		}
		principal, ok := auth.PrincipalFromContext(r.Context())
		if !ok {
			response.WriteError(w, r, http.StatusUnauthorized, response.ErrorCodeUnauthorized, "Authentication required", nil)
			return
		}
		if !auth.HasPermission(principal, auth.PermissionSubcontractView) {
			writePermissionDenied(w, r, auth.PermissionSubcontractView)
			return
		}
		if !auth.HasPermission(principal, auth.PermissionRecordCreate) {
			writePermissionDenied(w, r, auth.PermissionRecordCreate)
			return
		}

		r = requestWithStableID(r)
		var payload acceptSubcontractFinishedGoodsRequest
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			response.WriteError(w, r, http.StatusBadRequest, response.ErrorCodeValidation, "Invalid subcontract finished goods accept payload", nil)
			return
		}
		acceptedAt, err := parseSubcontractOptionalTime(payload.AcceptedAt)
		if err != nil {
			response.WriteError(w, r, http.StatusBadRequest, response.ErrorCodeValidation, "Invalid subcontract finished goods accept payload", map[string]any{"field": "accepted_at"})
			return
		}

		result, err := service.AcceptSubcontractFinishedGoods(r.Context(), productionapp.AcceptSubcontractFinishedGoodsInput{
			ID:              r.PathValue("subcontract_order_id"),
			ExpectedVersion: payload.ExpectedVersion,
			AcceptedBy:      payload.AcceptedBy,
			AcceptedAt:      acceptedAt,
			Note:            payload.Note,
			ActorID:         principal.UserID,
			RequestID:       response.RequestID(r),
		})
		if err != nil {
			writeSubcontractOrderError(w, r, err)
			return
		}

		response.WriteSuccess(w, r, http.StatusOK, acceptSubcontractFinishedGoodsResponse{
			SubcontractOrder: newSubcontractOrderResponse(result.SubcontractOrder, ""),
			StockMovements:   newSubcontractMaterialIssueMovementResponses(result.StockMovements),
			PreviousStatus:   string(result.PreviousStatus),
			CurrentStatus:    string(result.CurrentStatus),
			AuditLogID:       result.AuditLogID,
		})
	}
}

func subcontractOrderPartialAcceptFinishedGoodsHandler(service productionapp.SubcontractOrderService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			response.WriteError(w, r, http.StatusMethodNotAllowed, response.ErrorCodeNotFound, "Route not found", nil)
			return
		}
		principal, ok := auth.PrincipalFromContext(r.Context())
		if !ok {
			response.WriteError(w, r, http.StatusUnauthorized, response.ErrorCodeUnauthorized, "Authentication required", nil)
			return
		}
		if !auth.HasPermission(principal, auth.PermissionSubcontractView) {
			writePermissionDenied(w, r, auth.PermissionSubcontractView)
			return
		}
		if !auth.HasPermission(principal, auth.PermissionRecordCreate) {
			writePermissionDenied(w, r, auth.PermissionRecordCreate)
			return
		}

		r = requestWithStableID(r)
		var payload partialAcceptSubcontractFinishedGoodsRequest
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			response.WriteError(w, r, http.StatusBadRequest, response.ErrorCodeValidation, "Invalid subcontract finished goods partial accept payload", nil)
			return
		}
		acceptedAt, err := parseSubcontractOptionalTime(payload.AcceptedAt)
		if err != nil {
			response.WriteError(w, r, http.StatusBadRequest, response.ErrorCodeValidation, "Invalid subcontract finished goods partial accept payload", map[string]any{"field": "accepted_at"})
			return
		}
		openedAt, err := parseSubcontractOptionalTime(payload.OpenedAt)
		if err != nil {
			response.WriteError(w, r, http.StatusBadRequest, response.ErrorCodeValidation, "Invalid subcontract finished goods partial accept payload", map[string]any{"field": "opened_at"})
			return
		}

		result, err := service.PartialAcceptSubcontractFinishedGoods(r.Context(), productionapp.PartialAcceptSubcontractFinishedGoodsInput{
			ID:              r.PathValue("subcontract_order_id"),
			ExpectedVersion: payload.ExpectedVersion,
			AcceptedQty:     payload.AcceptedQty,
			UOMCode:         payload.UOMCode,
			BaseAcceptedQty: payload.BaseAcceptedQty,
			BaseUOMCode:     payload.BaseUOMCode,
			RejectedQty:     payload.RejectedQty,
			BaseRejectedQty: payload.BaseRejectedQty,
			ClaimID:         payload.ClaimID,
			ClaimNo:         payload.ClaimNo,
			ReceiptID:       payload.ReceiptID,
			ReceiptNo:       payload.ReceiptNo,
			ReasonCode:      payload.ReasonCode,
			Reason:          payload.Reason,
			Severity:        payload.Severity,
			Evidence:        reportSubcontractFactoryDefectEvidenceInputs(payload.Evidence),
			OwnerID:         payload.OwnerID,
			AcceptedBy:      payload.AcceptedBy,
			AcceptedAt:      acceptedAt,
			OpenedBy:        payload.OpenedBy,
			OpenedAt:        openedAt,
			Note:            payload.Note,
			ActorID:         principal.UserID,
			RequestID:       response.RequestID(r),
		})
		if err != nil {
			writeSubcontractOrderError(w, r, err)
			return
		}

		response.WriteSuccess(w, r, http.StatusOK, partialAcceptSubcontractFinishedGoodsResponse{
			SubcontractOrder: newSubcontractOrderResponse(result.SubcontractOrder, ""),
			Claim:            newSubcontractFactoryClaimResponse(result.Claim),
			StockMovements:   newSubcontractMaterialIssueMovementResponses(result.StockMovements),
			PreviousStatus:   string(result.PreviousStatus),
			CurrentStatus:    string(result.CurrentStatus),
			AcceptAuditLogID: result.AcceptAuditLogID,
			ClaimAuditLogID:  result.ClaimAuditLogID,
		})
	}
}

func subcontractOrderReportFactoryDefectHandler(service productionapp.SubcontractOrderService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			response.WriteError(w, r, http.StatusMethodNotAllowed, response.ErrorCodeNotFound, "Route not found", nil)
			return
		}
		principal, ok := auth.PrincipalFromContext(r.Context())
		if !ok {
			response.WriteError(w, r, http.StatusUnauthorized, response.ErrorCodeUnauthorized, "Authentication required", nil)
			return
		}
		if !auth.HasPermission(principal, auth.PermissionSubcontractView) {
			writePermissionDenied(w, r, auth.PermissionSubcontractView)
			return
		}
		if !auth.HasPermission(principal, auth.PermissionRecordCreate) {
			writePermissionDenied(w, r, auth.PermissionRecordCreate)
			return
		}

		r = requestWithStableID(r)
		var payload reportSubcontractFactoryDefectRequest
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			response.WriteError(w, r, http.StatusBadRequest, response.ErrorCodeValidation, "Invalid subcontract factory defect payload", nil)
			return
		}
		openedAt, err := parseSubcontractOptionalTime(payload.OpenedAt)
		if err != nil {
			response.WriteError(w, r, http.StatusBadRequest, response.ErrorCodeValidation, "Invalid subcontract factory defect payload", map[string]any{"field": "opened_at"})
			return
		}

		result, err := service.CreateSubcontractFactoryClaim(r.Context(), productionapp.CreateSubcontractFactoryClaimInput{
			ID:              r.PathValue("subcontract_order_id"),
			ExpectedVersion: payload.ExpectedVersion,
			ClaimID:         payload.ClaimID,
			ClaimNo:         payload.ClaimNo,
			ReceiptID:       payload.ReceiptID,
			ReceiptNo:       payload.ReceiptNo,
			ReasonCode:      payload.ReasonCode,
			Reason:          payload.Reason,
			Severity:        payload.Severity,
			AffectedQty:     payload.AffectedQty,
			UOMCode:         payload.UOMCode,
			BaseAffectedQty: payload.BaseAffectedQty,
			BaseUOMCode:     payload.BaseUOMCode,
			Evidence:        reportSubcontractFactoryDefectEvidenceInputs(payload.Evidence),
			OwnerID:         payload.OwnerID,
			OpenedBy:        payload.OpenedBy,
			OpenedAt:        openedAt,
			ActorID:         principal.UserID,
			RequestID:       response.RequestID(r),
		})
		if err != nil {
			writeSubcontractOrderError(w, r, err)
			return
		}

		response.WriteSuccess(w, r, http.StatusOK, reportSubcontractFactoryDefectResponse{
			SubcontractOrder: newSubcontractOrderResponse(result.SubcontractOrder, ""),
			Claim:            newSubcontractFactoryClaimResponse(result.Claim),
			PreviousStatus:   string(result.PreviousStatus),
			CurrentStatus:    string(result.CurrentStatus),
			AuditLogID:       result.AuditLogID,
		})
	}
}

func subcontractFactoryClaimsHandler(service productionapp.SubcontractOrderService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		principal, ok := auth.PrincipalFromContext(r.Context())
		if !ok {
			response.WriteError(w, r, http.StatusUnauthorized, response.ErrorCodeUnauthorized, "Authentication required", nil)
			return
		}
		if !auth.HasPermission(principal, auth.PermissionSubcontractView) {
			writePermissionDenied(w, r, auth.PermissionSubcontractView)
			return
		}

		switch r.Method {
		case http.MethodGet:
			claims, err := service.ListSubcontractFactoryClaims(r.Context(), r.PathValue("subcontract_order_id"))
			if err != nil {
				writeSubcontractOrderError(w, r, err)
				return
			}
			payload := make([]subcontractFactoryClaimResponse, 0, len(claims))
			for _, claim := range claims {
				payload = append(payload, newSubcontractFactoryClaimResponse(claim))
			}
			response.WriteSuccess(w, r, http.StatusOK, payload)
		default:
			response.WriteError(w, r, http.StatusMethodNotAllowed, response.ErrorCodeNotFound, "Route not found", nil)
		}
	}
}

func subcontractFactoryClaimAcknowledgeHandler(service productionapp.SubcontractOrderService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			response.WriteError(w, r, http.StatusMethodNotAllowed, response.ErrorCodeNotFound, "Route not found", nil)
			return
		}
		principal, ok := auth.PrincipalFromContext(r.Context())
		if !ok {
			response.WriteError(w, r, http.StatusUnauthorized, response.ErrorCodeUnauthorized, "Authentication required", nil)
			return
		}
		if !auth.HasPermission(principal, auth.PermissionSubcontractView) {
			writePermissionDenied(w, r, auth.PermissionSubcontractView)
			return
		}
		if !auth.HasPermission(principal, auth.PermissionRecordCreate) {
			writePermissionDenied(w, r, auth.PermissionRecordCreate)
			return
		}

		r = requestWithStableID(r)
		var payload subcontractFactoryClaimDecisionRequest
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			response.WriteError(w, r, http.StatusBadRequest, response.ErrorCodeValidation, "Invalid subcontract factory claim payload", nil)
			return
		}
		acknowledgedAt, err := parseSubcontractOptionalTime(payload.AcknowledgedAt)
		if err != nil {
			response.WriteError(w, r, http.StatusBadRequest, response.ErrorCodeValidation, "Invalid subcontract factory claim payload", map[string]any{"field": "acknowledged_at"})
			return
		}

		result, err := service.AcknowledgeSubcontractFactoryClaim(r.Context(), productionapp.AcknowledgeSubcontractFactoryClaimInput{
			ID:                 r.PathValue("factory_claim_id"),
			SubcontractOrderID: r.PathValue("subcontract_order_id"),
			ExpectedVersion:    payload.ExpectedVersion,
			AcknowledgedBy:     payload.AcknowledgedBy,
			AcknowledgedAt:     acknowledgedAt,
			ActorID:            principal.UserID,
			RequestID:          response.RequestID(r),
		})
		if err != nil {
			writeSubcontractOrderError(w, r, err)
			return
		}
		response.WriteSuccess(w, r, http.StatusOK, subcontractFactoryClaimDecisionResponse{
			Claim:      newSubcontractFactoryClaimResponse(result.Claim),
			AuditLogID: result.AuditLogID,
		})
	}
}

func subcontractFactoryClaimResolveHandler(service productionapp.SubcontractOrderService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			response.WriteError(w, r, http.StatusMethodNotAllowed, response.ErrorCodeNotFound, "Route not found", nil)
			return
		}
		principal, ok := auth.PrincipalFromContext(r.Context())
		if !ok {
			response.WriteError(w, r, http.StatusUnauthorized, response.ErrorCodeUnauthorized, "Authentication required", nil)
			return
		}
		if !auth.HasPermission(principal, auth.PermissionSubcontractView) {
			writePermissionDenied(w, r, auth.PermissionSubcontractView)
			return
		}
		if !auth.HasPermission(principal, auth.PermissionRecordCreate) {
			writePermissionDenied(w, r, auth.PermissionRecordCreate)
			return
		}

		r = requestWithStableID(r)
		var payload subcontractFactoryClaimDecisionRequest
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			response.WriteError(w, r, http.StatusBadRequest, response.ErrorCodeValidation, "Invalid subcontract factory claim payload", nil)
			return
		}
		resolvedAt, err := parseSubcontractOptionalTime(payload.ResolvedAt)
		if err != nil {
			response.WriteError(w, r, http.StatusBadRequest, response.ErrorCodeValidation, "Invalid subcontract factory claim payload", map[string]any{"field": "resolved_at"})
			return
		}

		result, err := service.ResolveSubcontractFactoryClaim(r.Context(), productionapp.ResolveSubcontractFactoryClaimInput{
			ID:                 r.PathValue("factory_claim_id"),
			SubcontractOrderID: r.PathValue("subcontract_order_id"),
			ExpectedVersion:    payload.ExpectedVersion,
			ResolvedBy:         payload.ResolvedBy,
			ResolvedAt:         resolvedAt,
			ResolutionNote:     payload.ResolutionNote,
			ActorID:            principal.UserID,
			RequestID:          response.RequestID(r),
		})
		if err != nil {
			writeSubcontractOrderError(w, r, err)
			return
		}
		response.WriteSuccess(w, r, http.StatusOK, subcontractFactoryClaimDecisionResponse{
			Claim:      newSubcontractFactoryClaimResponse(result.Claim),
			AuditLogID: result.AuditLogID,
		})
	}
}

func subcontractOrderSubmitSampleHandler(service productionapp.SubcontractOrderService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			response.WriteError(w, r, http.StatusMethodNotAllowed, response.ErrorCodeNotFound, "Route not found", nil)
			return
		}
		principal, ok := auth.PrincipalFromContext(r.Context())
		if !ok {
			response.WriteError(w, r, http.StatusUnauthorized, response.ErrorCodeUnauthorized, "Authentication required", nil)
			return
		}
		if !auth.HasPermission(principal, auth.PermissionSubcontractView) {
			writePermissionDenied(w, r, auth.PermissionSubcontractView)
			return
		}
		if !auth.HasPermission(principal, auth.PermissionRecordCreate) {
			writePermissionDenied(w, r, auth.PermissionRecordCreate)
			return
		}

		r = requestWithStableID(r)
		var payload submitSubcontractSampleRequest
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			response.WriteError(w, r, http.StatusBadRequest, response.ErrorCodeValidation, "Invalid subcontract sample payload", nil)
			return
		}
		submittedAt, err := parseSubcontractOptionalTime(payload.SubmittedAt)
		if err != nil {
			response.WriteError(w, r, http.StatusBadRequest, response.ErrorCodeValidation, "Invalid subcontract sample payload", map[string]any{"field": "submitted_at"})
			return
		}

		result, err := service.SubmitSubcontractSample(r.Context(), productionapp.SubmitSubcontractSampleInput{
			ID:               r.PathValue("subcontract_order_id"),
			ExpectedVersion:  payload.ExpectedVersion,
			SampleApprovalID: payload.SampleApprovalID,
			SampleCode:       payload.SampleCode,
			FormulaVersion:   payload.FormulaVersion,
			SpecVersion:      payload.SpecVersion,
			Evidence:         subcontractSampleEvidenceInputs(payload.Evidence),
			SubmittedBy:      payload.SubmittedBy,
			SubmittedAt:      submittedAt,
			Note:             payload.Note,
			ActorID:          principal.UserID,
			RequestID:        response.RequestID(r),
		})
		if err != nil {
			writeSubcontractOrderError(w, r, err)
			return
		}

		response.WriteSuccess(w, r, http.StatusOK, newSubcontractSampleApprovalResultResponse(result))
	}
}

func subcontractOrderApproveSampleHandler(service productionapp.SubcontractOrderService) http.HandlerFunc {
	return subcontractOrderDecideSampleHandler(service, "approve")
}

func subcontractOrderRejectSampleHandler(service productionapp.SubcontractOrderService) http.HandlerFunc {
	return subcontractOrderDecideSampleHandler(service, "reject")
}

func subcontractFactoryDispatchesHandler(service productionapp.SubcontractOrderService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		principal, ok := auth.PrincipalFromContext(r.Context())
		if !ok {
			response.WriteError(w, r, http.StatusUnauthorized, response.ErrorCodeUnauthorized, "Authentication required", nil)
			return
		}
		if !auth.HasPermission(principal, auth.PermissionSubcontractView) {
			writePermissionDenied(w, r, auth.PermissionSubcontractView)
			return
		}

		switch r.Method {
		case http.MethodGet:
			dispatches, err := service.ListFactoryDispatches(r.Context(), r.PathValue("subcontract_order_id"))
			if err != nil {
				writeSubcontractOrderError(w, r, err)
				return
			}
			payload := make([]subcontractFactoryDispatchResponse, 0, len(dispatches))
			for _, dispatch := range dispatches {
				payload = append(payload, newSubcontractFactoryDispatchResponse(dispatch))
			}
			response.WriteSuccess(w, r, http.StatusOK, payload)
		case http.MethodPost:
			if !auth.HasPermission(principal, auth.PermissionRecordCreate) {
				writePermissionDenied(w, r, auth.PermissionRecordCreate)
				return
			}
			r = requestWithStableID(r)
			var payload createSubcontractFactoryDispatchRequest
			if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
				response.WriteError(w, r, http.StatusBadRequest, response.ErrorCodeValidation, "Invalid subcontract factory dispatch payload", nil)
				return
			}
			result, err := service.CreateFactoryDispatch(r.Context(), productionapp.CreateFactoryDispatchInput{
				ID:                 payload.DispatchID,
				DispatchNo:         payload.DispatchNo,
				SubcontractOrderID: r.PathValue("subcontract_order_id"),
				ExpectedVersion:    payload.ExpectedVersion,
				ActorID:            principal.UserID,
				RequestID:          response.RequestID(r),
				Note:               payload.Note,
			})
			if err != nil {
				writeSubcontractOrderError(w, r, err)
				return
			}
			response.WriteSuccess(w, r, http.StatusCreated, newSubcontractFactoryDispatchResultResponse(result))
		default:
			response.WriteError(w, r, http.StatusMethodNotAllowed, response.ErrorCodeNotFound, "Route not found", nil)
		}
	}
}

func subcontractFactoryDispatchReadyHandler(service productionapp.SubcontractOrderService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			response.WriteError(w, r, http.StatusMethodNotAllowed, response.ErrorCodeNotFound, "Route not found", nil)
			return
		}
		principal, ok := auth.PrincipalFromContext(r.Context())
		if !ok {
			response.WriteError(w, r, http.StatusUnauthorized, response.ErrorCodeUnauthorized, "Authentication required", nil)
			return
		}
		if !auth.HasPermission(principal, auth.PermissionSubcontractView) {
			writePermissionDenied(w, r, auth.PermissionSubcontractView)
			return
		}
		if !auth.HasPermission(principal, auth.PermissionRecordCreate) {
			writePermissionDenied(w, r, auth.PermissionRecordCreate)
			return
		}

		r = requestWithStableID(r)
		var payload subcontractOrderActionRequest
		if r.Body != nil {
			if err := json.NewDecoder(r.Body).Decode(&payload); err != nil && !errors.Is(err, io.EOF) {
				response.WriteError(w, r, http.StatusBadRequest, response.ErrorCodeValidation, "Invalid subcontract factory dispatch payload", nil)
				return
			}
		}
		result, err := service.MarkFactoryDispatchReady(r.Context(), productionapp.FactoryDispatchActionInput{
			SubcontractOrderID: r.PathValue("subcontract_order_id"),
			DispatchID:         r.PathValue("factory_dispatch_id"),
			ExpectedVersion:    payload.ExpectedVersion,
			ActorID:            principal.UserID,
			RequestID:          response.RequestID(r),
		})
		if err != nil {
			writeSubcontractOrderError(w, r, err)
			return
		}
		response.WriteSuccess(w, r, http.StatusOK, newSubcontractFactoryDispatchResultResponse(result))
	}
}

func subcontractFactoryDispatchSentHandler(service productionapp.SubcontractOrderService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			response.WriteError(w, r, http.StatusMethodNotAllowed, response.ErrorCodeNotFound, "Route not found", nil)
			return
		}
		principal, ok := auth.PrincipalFromContext(r.Context())
		if !ok {
			response.WriteError(w, r, http.StatusUnauthorized, response.ErrorCodeUnauthorized, "Authentication required", nil)
			return
		}
		if !auth.HasPermission(principal, auth.PermissionSubcontractView) {
			writePermissionDenied(w, r, auth.PermissionSubcontractView)
			return
		}
		if !auth.HasPermission(principal, auth.PermissionRecordCreate) {
			writePermissionDenied(w, r, auth.PermissionRecordCreate)
			return
		}

		r = requestWithStableID(r)
		var payload markSubcontractFactoryDispatchSentRequest
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			response.WriteError(w, r, http.StatusBadRequest, response.ErrorCodeValidation, "Invalid subcontract factory dispatch payload", nil)
			return
		}
		sentAt, err := parseSubcontractOptionalTime(payload.SentAt)
		if err != nil {
			response.WriteError(w, r, http.StatusBadRequest, response.ErrorCodeValidation, "Invalid subcontract factory dispatch payload", map[string]any{"field": "sent_at"})
			return
		}
		result, err := service.MarkFactoryDispatchSent(r.Context(), productionapp.MarkFactoryDispatchSentInput{
			SubcontractOrderID: r.PathValue("subcontract_order_id"),
			DispatchID:         r.PathValue("factory_dispatch_id"),
			ExpectedVersion:    payload.ExpectedVersion,
			SentBy:             payload.SentBy,
			SentAt:             sentAt,
			ActorID:            principal.UserID,
			RequestID:          response.RequestID(r),
			Note:               payload.Note,
			Evidence:           subcontractFactoryDispatchEvidenceInputs(payload.Evidence),
		})
		if err != nil {
			writeSubcontractOrderError(w, r, err)
			return
		}
		response.WriteSuccess(w, r, http.StatusOK, newSubcontractFactoryDispatchResultResponse(result))
	}
}

func subcontractFactoryDispatchRespondHandler(service productionapp.SubcontractOrderService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			response.WriteError(w, r, http.StatusMethodNotAllowed, response.ErrorCodeNotFound, "Route not found", nil)
			return
		}
		principal, ok := auth.PrincipalFromContext(r.Context())
		if !ok {
			response.WriteError(w, r, http.StatusUnauthorized, response.ErrorCodeUnauthorized, "Authentication required", nil)
			return
		}
		if !auth.HasPermission(principal, auth.PermissionSubcontractView) {
			writePermissionDenied(w, r, auth.PermissionSubcontractView)
			return
		}
		if !auth.HasPermission(principal, auth.PermissionRecordCreate) {
			writePermissionDenied(w, r, auth.PermissionRecordCreate)
			return
		}

		r = requestWithStableID(r)
		var payload recordSubcontractFactoryDispatchResponseRequest
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			response.WriteError(w, r, http.StatusBadRequest, response.ErrorCodeValidation, "Invalid subcontract factory dispatch payload", nil)
			return
		}
		respondedAt, err := parseSubcontractOptionalTime(payload.RespondedAt)
		if err != nil {
			response.WriteError(w, r, http.StatusBadRequest, response.ErrorCodeValidation, "Invalid subcontract factory dispatch payload", map[string]any{"field": "responded_at"})
			return
		}
		result, err := service.RecordFactoryDispatchResponse(r.Context(), productionapp.RecordFactoryDispatchResponseInput{
			SubcontractOrderID: r.PathValue("subcontract_order_id"),
			DispatchID:         r.PathValue("factory_dispatch_id"),
			ExpectedVersion:    payload.ExpectedVersion,
			ResponseStatus:     payload.ResponseStatus,
			ResponseBy:         payload.ResponseBy,
			RespondedAt:        respondedAt,
			ResponseNote:       payload.ResponseNote,
			ActorID:            principal.UserID,
			RequestID:          response.RequestID(r),
		})
		if err != nil {
			writeSubcontractOrderError(w, r, err)
			return
		}
		response.WriteSuccess(w, r, http.StatusOK, newSubcontractFactoryDispatchResultResponse(result))
	}
}

func subcontractOrderDecideSampleHandler(service productionapp.SubcontractOrderService, decision string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			response.WriteError(w, r, http.StatusMethodNotAllowed, response.ErrorCodeNotFound, "Route not found", nil)
			return
		}
		principal, ok := auth.PrincipalFromContext(r.Context())
		if !ok {
			response.WriteError(w, r, http.StatusUnauthorized, response.ErrorCodeUnauthorized, "Authentication required", nil)
			return
		}
		if !auth.HasPermission(principal, auth.PermissionSubcontractView) {
			writePermissionDenied(w, r, auth.PermissionSubcontractView)
			return
		}
		if !auth.HasPermission(principal, auth.PermissionRecordCreate) {
			writePermissionDenied(w, r, auth.PermissionRecordCreate)
			return
		}

		r = requestWithStableID(r)
		var payload decideSubcontractSampleRequest
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			response.WriteError(w, r, http.StatusBadRequest, response.ErrorCodeValidation, "Invalid subcontract sample decision payload", nil)
			return
		}
		decisionAt, err := parseSubcontractOptionalTime(payload.DecisionAt)
		if err != nil {
			response.WriteError(w, r, http.StatusBadRequest, response.ErrorCodeValidation, "Invalid subcontract sample decision payload", map[string]any{"field": "decision_at"})
			return
		}

		input := productionapp.DecideSubcontractSampleInput{
			ID:               r.PathValue("subcontract_order_id"),
			ExpectedVersion:  payload.ExpectedVersion,
			SampleApprovalID: payload.SampleApprovalID,
			DecisionAt:       decisionAt,
			Reason:           payload.Reason,
			StorageStatus:    payload.StorageStatus,
			ActorID:          principal.UserID,
			RequestID:        response.RequestID(r),
		}
		var result productionapp.SubcontractSampleApprovalResult
		switch decision {
		case "approve":
			result, err = service.ApproveSubcontractSample(r.Context(), input)
		case "reject":
			result, err = service.RejectSubcontractSample(r.Context(), input)
		default:
			response.WriteError(w, r, http.StatusNotFound, response.ErrorCodeNotFound, "Route not found", nil)
			return
		}
		if err != nil {
			writeSubcontractOrderError(w, r, err)
			return
		}

		response.WriteSuccess(w, r, http.StatusOK, newSubcontractSampleApprovalResultResponse(result))
	}
}

func subcontractOrderActionHandler(service productionapp.SubcontractOrderService, action string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			response.WriteError(w, r, http.StatusMethodNotAllowed, response.ErrorCodeNotFound, "Route not found", nil)
			return
		}
		principal, ok := auth.PrincipalFromContext(r.Context())
		if !ok {
			response.WriteError(w, r, http.StatusUnauthorized, response.ErrorCodeUnauthorized, "Authentication required", nil)
			return
		}
		if !auth.HasPermission(principal, auth.PermissionSubcontractView) {
			writePermissionDenied(w, r, auth.PermissionSubcontractView)
			return
		}
		if !auth.HasPermission(principal, auth.PermissionRecordCreate) {
			writePermissionDenied(w, r, auth.PermissionRecordCreate)
			return
		}

		r = requestWithStableID(r)
		var payload subcontractOrderActionRequest
		if r.Body != nil {
			if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
				if !errors.Is(err, io.EOF) {
					response.WriteError(w, r, http.StatusBadRequest, response.ErrorCodeValidation, "Invalid subcontract order action payload", nil)
					return
				}
			}
		}

		input := productionapp.SubcontractOrderActionInput{
			ID:              r.PathValue("subcontract_order_id"),
			ExpectedVersion: payload.ExpectedVersion,
			Reason:          payload.Reason,
			ActorID:         principal.UserID,
			RequestID:       response.RequestID(r),
		}
		var (
			result productionapp.SubcontractOrderActionResult
			err    error
		)
		switch action {
		case "submit":
			result, err = service.SubmitSubcontractOrder(r.Context(), input)
		case "approve":
			result, err = service.ApproveSubcontractOrder(r.Context(), input)
		case "confirm-factory":
			result, err = service.ConfirmFactorySubcontractOrder(r.Context(), input)
		case "start-mass-production":
			result, err = service.StartMassProductionSubcontractOrder(r.Context(), input)
		case "cancel":
			result, err = service.CancelSubcontractOrder(r.Context(), input)
		case "close":
			result, err = service.CloseSubcontractOrder(r.Context(), input)
		default:
			response.WriteError(w, r, http.StatusNotFound, response.ErrorCodeNotFound, "Route not found", nil)
			return
		}
		if err != nil {
			writeSubcontractOrderError(w, r, err)
			return
		}

		response.WriteSuccess(w, r, http.StatusOK, subcontractOrderActionResultResponse{
			SubcontractOrder: newSubcontractOrderResponse(result.SubcontractOrder, ""),
			PreviousStatus:   string(result.PreviousStatus),
			CurrentStatus:    string(result.CurrentStatus),
			AuditLogID:       result.AuditLogID,
		})
	}
}

func subcontractOrderFilterFromRequest(r *http.Request) productionapp.SubcontractOrderFilter {
	query := r.URL.Query()
	statuses := []productiondomain.SubcontractOrderStatus{}
	for _, rawStatus := range strings.Split(query.Get("status"), ",") {
		status := productiondomain.NormalizeSubcontractOrderStatus(productiondomain.SubcontractOrderStatus(rawStatus))
		if status != "" {
			statuses = append(statuses, status)
		}
	}

	return productionapp.SubcontractOrderFilter{
		Search:                 query.Get("search"),
		Statuses:               statuses,
		FactoryID:              query.Get("factory_id"),
		FinishedItemID:         query.Get("finished_item_id"),
		SourceProductionPlanID: query.Get("source_production_plan_id"),
		ExpectedReceiptFrom:    query.Get("expected_receipt_from"),
		ExpectedReceiptTo:      query.Get("expected_receipt_to"),
	}
}

func subcontractOrderMaterialLineInputs(
	inputs []subcontractOrderMaterialLineRequest,
) []productionapp.SubcontractOrderMaterialLineInput {
	if inputs == nil {
		return nil
	}
	lines := make([]productionapp.SubcontractOrderMaterialLineInput, 0, len(inputs))
	for _, input := range inputs {
		lines = append(lines, productionapp.SubcontractOrderMaterialLineInput{
			ID:               input.ID,
			LineNo:           input.LineNo,
			ItemID:           input.ItemID,
			PlannedQty:       input.PlannedQty,
			UOMCode:          input.UOMCode,
			UnitCost:         input.UnitCost,
			CurrencyCode:     input.CurrencyCode,
			LotTraceRequired: input.LotTraceRequired,
			Note:             input.Note,
		})
	}

	return lines
}

func issueSubcontractMaterialLineInputs(
	inputs []issueSubcontractMaterialsLineRequest,
) []productionapp.IssueSubcontractMaterialsLineInput {
	lines := make([]productionapp.IssueSubcontractMaterialsLineInput, 0, len(inputs))
	for _, input := range inputs {
		lines = append(lines, productionapp.IssueSubcontractMaterialsLineInput{
			ID:                  input.ID,
			LineNo:              input.LineNo,
			OrderMaterialLineID: input.OrderMaterialLineID,
			IssueQty:            input.IssueQty,
			UOMCode:             input.UOMCode,
			BaseIssueQty:        input.BaseIssueQty,
			BaseUOMCode:         input.BaseUOMCode,
			ConversionFactor:    input.ConversionFactor,
			BatchID:             input.BatchID,
			BatchNo:             input.BatchNo,
			LotNo:               input.LotNo,
			SourceBinID:         input.SourceBinID,
			Note:                input.Note,
		})
	}

	return lines
}

func issueSubcontractMaterialEvidenceInputs(
	inputs []issueSubcontractMaterialsEvidenceRequest,
) []productionapp.IssueSubcontractMaterialsEvidenceInput {
	evidence := make([]productionapp.IssueSubcontractMaterialsEvidenceInput, 0, len(inputs))
	for _, input := range inputs {
		evidence = append(evidence, productionapp.IssueSubcontractMaterialsEvidenceInput{
			ID:           input.ID,
			EvidenceType: input.EvidenceType,
			FileName:     input.FileName,
			ObjectKey:    input.ObjectKey,
			ExternalURL:  input.ExternalURL,
			Note:         input.Note,
		})
	}

	return evidence
}

func subcontractSampleEvidenceInputs(
	inputs []subcontractSampleEvidenceRequest,
) []productionapp.SubcontractSampleEvidenceInput {
	evidence := make([]productionapp.SubcontractSampleEvidenceInput, 0, len(inputs))
	for _, input := range inputs {
		evidence = append(evidence, productionapp.SubcontractSampleEvidenceInput{
			ID:           input.ID,
			EvidenceType: input.EvidenceType,
			FileName:     input.FileName,
			ObjectKey:    input.ObjectKey,
			ExternalURL:  input.ExternalURL,
			Note:         input.Note,
		})
	}

	return evidence
}

func receiveSubcontractFinishedGoodsLineInputs(
	inputs []receiveSubcontractFinishedGoodsLineRequest,
) []productionapp.ReceiveSubcontractFinishedGoodsLineInput {
	lines := make([]productionapp.ReceiveSubcontractFinishedGoodsLineInput, 0, len(inputs))
	for _, input := range inputs {
		lines = append(lines, productionapp.ReceiveSubcontractFinishedGoodsLineInput{
			ID:               input.ID,
			LineNo:           input.LineNo,
			ItemID:           input.ItemID,
			SKUCode:          input.SKUCode,
			ItemName:         input.ItemName,
			BatchID:          input.BatchID,
			BatchNo:          input.BatchNo,
			LotNo:            input.LotNo,
			ExpiryDate:       input.ExpiryDate,
			ReceiveQty:       input.ReceiveQty,
			UOMCode:          input.UOMCode,
			BaseReceiveQty:   input.BaseReceiveQty,
			BaseUOMCode:      input.BaseUOMCode,
			ConversionFactor: input.ConversionFactor,
			PackagingStatus:  input.PackagingStatus,
			Note:             input.Note,
		})
	}

	return lines
}

func receiveSubcontractFinishedGoodsEvidenceInputs(
	inputs []receiveSubcontractFinishedGoodsEvidenceRequest,
) []productionapp.ReceiveSubcontractFinishedGoodsEvidenceInput {
	evidence := make([]productionapp.ReceiveSubcontractFinishedGoodsEvidenceInput, 0, len(inputs))
	for _, input := range inputs {
		evidence = append(evidence, productionapp.ReceiveSubcontractFinishedGoodsEvidenceInput{
			ID:           input.ID,
			EvidenceType: input.EvidenceType,
			FileName:     input.FileName,
			ObjectKey:    input.ObjectKey,
			ExternalURL:  input.ExternalURL,
			Note:         input.Note,
		})
	}

	return evidence
}

func reportSubcontractFactoryDefectEvidenceInputs(
	inputs []reportSubcontractFactoryDefectEvidenceRequest,
) []productionapp.CreateSubcontractFactoryClaimEvidenceInput {
	evidence := make([]productionapp.CreateSubcontractFactoryClaimEvidenceInput, 0, len(inputs))
	for _, input := range inputs {
		evidence = append(evidence, productionapp.CreateSubcontractFactoryClaimEvidenceInput{
			ID:           input.ID,
			EvidenceType: input.EvidenceType,
			FileName:     input.FileName,
			ObjectKey:    input.ObjectKey,
			ExternalURL:  input.ExternalURL,
			Note:         input.Note,
		})
	}

	return evidence
}

func subcontractFactoryDispatchEvidenceInputs(
	inputs []subcontractFactoryDispatchEvidenceRequest,
) []productionapp.FactoryDispatchEvidenceInput {
	evidence := make([]productionapp.FactoryDispatchEvidenceInput, 0, len(inputs))
	for _, input := range inputs {
		evidence = append(evidence, productionapp.FactoryDispatchEvidenceInput{
			ID:           input.ID,
			EvidenceType: input.EvidenceType,
			FileName:     input.FileName,
			ObjectKey:    input.ObjectKey,
			ExternalURL:  input.ExternalURL,
			Note:         input.Note,
		})
	}

	return evidence
}

func parseSubcontractOptionalTime(value string) (time.Time, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return time.Time{}, nil
	}

	return time.Parse(time.RFC3339, value)
}

func newSubcontractOrderListItemResponse(order productiondomain.SubcontractOrder) subcontractOrderListItemResponse {
	return subcontractOrderListItemResponse{
		ID:                     order.ID,
		OrderNo:                order.OrderNo,
		FactoryID:              order.FactoryID,
		FactoryCode:            order.FactoryCode,
		FactoryName:            order.FactoryName,
		FinishedItemID:         order.FinishedItemID,
		FinishedSKUCode:        order.FinishedSKUCode,
		FinishedItemName:       order.FinishedItemName,
		PlannedQty:             order.PlannedQty.String(),
		UOMCode:                order.UOMCode.String(),
		ExpectedReceiptDate:    order.ExpectedReceiptDate,
		Status:                 string(order.Status),
		CurrencyCode:           order.CurrencyCode.String(),
		EstimatedCostAmount:    order.EstimatedCostAmount.String(),
		ReceivedQty:            order.ReceivedQty.String(),
		AcceptedQty:            order.AcceptedQty.String(),
		RejectedQty:            order.RejectedQty.String(),
		SourceProductionPlanID: order.SourceProductionPlanID,
		SourceProductionPlanNo: order.SourceProductionPlanNo,
		MaterialLineCount:      len(order.MaterialLines),
		SampleRequired:         order.SampleRequired,
		CreatedAt:              timeString(order.CreatedAt),
		UpdatedAt:              timeString(order.UpdatedAt),
		Version:                order.Version,
	}
}

func newSubcontractOrderResponse(order productiondomain.SubcontractOrder, auditLogID string) subcontractOrderResponse {
	payload := subcontractOrderResponse{
		ID:                      order.ID,
		OrgID:                   order.OrgID,
		OrderNo:                 order.OrderNo,
		FactoryID:               order.FactoryID,
		FactoryCode:             order.FactoryCode,
		FactoryName:             order.FactoryName,
		FinishedItemID:          order.FinishedItemID,
		FinishedSKUCode:         order.FinishedSKUCode,
		FinishedItemName:        order.FinishedItemName,
		PlannedQty:              order.PlannedQty.String(),
		ReceivedQty:             order.ReceivedQty.String(),
		AcceptedQty:             order.AcceptedQty.String(),
		RejectedQty:             order.RejectedQty.String(),
		UOMCode:                 order.UOMCode.String(),
		BasePlannedQty:          order.BasePlannedQty.String(),
		BaseReceivedQty:         order.BaseReceivedQty.String(),
		BaseAcceptedQty:         order.BaseAcceptedQty.String(),
		BaseRejectedQty:         order.BaseRejectedQty.String(),
		BaseUOMCode:             order.BaseUOMCode.String(),
		ConversionFactor:        order.ConversionFactor.String(),
		CurrencyCode:            order.CurrencyCode.String(),
		EstimatedCostAmount:     order.EstimatedCostAmount.String(),
		DepositAmount:           order.DepositAmount.String(),
		SpecSummary:             order.SpecSummary,
		SourceProductionPlanID:  order.SourceProductionPlanID,
		SourceProductionPlanNo:  order.SourceProductionPlanNo,
		SampleRequired:          order.SampleRequired,
		ClaimWindowDays:         order.ClaimWindowDays,
		TargetStartDate:         order.TargetStartDate,
		ExpectedReceiptDate:     order.ExpectedReceiptDate,
		Status:                  string(order.Status),
		MaterialLines:           make([]subcontractOrderMaterialLineResponse, 0, len(order.MaterialLines)),
		AuditLogID:              auditLogID,
		CreatedAt:               timeString(order.CreatedAt),
		UpdatedAt:               timeString(order.UpdatedAt),
		SubmittedAt:             timeString(order.SubmittedAt),
		ApprovedAt:              timeString(order.ApprovedAt),
		FactoryConfirmedAt:      timeString(order.FactoryConfirmedAt),
		DepositRecordedAt:       timeString(order.DepositRecordedAt),
		MaterialsIssuedAt:       timeString(order.MaterialsIssuedAt),
		SampleSubmittedAt:       timeString(order.SampleSubmittedAt),
		SampleApprovedAt:        timeString(order.SampleApprovedAt),
		SampleRejectedAt:        timeString(order.SampleRejectedAt),
		MassProductionStartedAt: timeString(order.MassProductionStartedAt),
		FinishedGoodsReceivedAt: timeString(order.FinishedGoodsReceivedAt),
		QCStartedAt:             timeString(order.QCStartedAt),
		AcceptedAt:              timeString(order.AcceptedAt),
		RejectedFactoryIssueAt:  timeString(order.RejectedFactoryIssueAt),
		FinalPaymentReadyAt:     timeString(order.FinalPaymentReadyAt),
		ClosedAt:                timeString(order.ClosedAt),
		CancelledAt:             timeString(order.CancelledAt),
		CancelReason:            order.CancelReason,
		SampleRejectReason:      order.SampleRejectReason,
		FactoryIssueReason:      order.FactoryIssueReason,
		Version:                 order.Version,
	}
	for _, line := range order.MaterialLines {
		payload.MaterialLines = append(payload.MaterialLines, subcontractOrderMaterialLineResponse{
			ID:               line.ID,
			LineNo:           line.LineNo,
			ItemID:           line.ItemID,
			SKUCode:          line.SKUCode,
			ItemName:         line.ItemName,
			PlannedQty:       line.PlannedQty.String(),
			IssuedQty:        line.IssuedQty.String(),
			UOMCode:          line.UOMCode.String(),
			BasePlannedQty:   line.BasePlannedQty.String(),
			BaseIssuedQty:    line.BaseIssuedQty.String(),
			BaseUOMCode:      line.BaseUOMCode.String(),
			ConversionFactor: line.ConversionFactor.String(),
			UnitCost:         line.UnitCost.String(),
			CurrencyCode:     line.CurrencyCode.String(),
			LineCostAmount:   line.LineCostAmount.String(),
			LotTraceRequired: line.LotTraceRequired,
			Note:             line.Note,
		})
	}

	return payload
}

func newSubcontractPaymentMilestoneResultResponse(
	result productionapp.SubcontractPaymentMilestoneResult,
) subcontractPaymentMilestoneResultResponse {
	return subcontractPaymentMilestoneResultResponse{
		SubcontractOrder: newSubcontractOrderResponse(result.SubcontractOrder, ""),
		Milestone:        newSubcontractPaymentMilestoneResponse(result.Milestone),
		PreviousStatus:   string(result.PreviousStatus),
		CurrentStatus:    string(result.CurrentStatus),
		AuditLogID:       result.AuditLogID,
	}
}

func newSubcontractPaymentMilestoneResponse(
	milestone productiondomain.SubcontractPaymentMilestone,
) subcontractPaymentMilestoneResponse {
	return subcontractPaymentMilestoneResponse{
		ID:                  milestone.ID,
		MilestoneNo:         milestone.MilestoneNo,
		SubcontractOrderID:  milestone.SubcontractOrderID,
		SubcontractOrderNo:  milestone.SubcontractOrderNo,
		FactoryID:           milestone.FactoryID,
		FactoryCode:         milestone.FactoryCode,
		FactoryName:         milestone.FactoryName,
		Kind:                string(milestone.Kind),
		Status:              string(milestone.Status),
		Amount:              milestone.Amount.String(),
		CurrencyCode:        milestone.CurrencyCode.String(),
		Note:                milestone.Note,
		ApprovedExceptionID: milestone.ApprovedExceptionID,
		RecordedBy:          milestone.RecordedBy,
		RecordedAt:          timeString(milestone.RecordedAt),
		ReadyBy:             milestone.ReadyBy,
		ReadyAt:             timeString(milestone.ReadyAt),
		BlockedBy:           milestone.BlockedBy,
		BlockedAt:           timeString(milestone.BlockedAt),
		BlockReason:         milestone.BlockReason,
		CreatedAt:           timeString(milestone.CreatedAt),
		UpdatedAt:           timeString(milestone.UpdatedAt),
		Version:             milestone.Version,
	}
}

func newSubcontractMaterialTransferResponse(
	transfer productiondomain.SubcontractMaterialTransfer,
) subcontractMaterialTransferResponse {
	payload := subcontractMaterialTransferResponse{
		ID:                  transfer.ID,
		OrgID:               transfer.OrgID,
		TransferNo:          transfer.TransferNo,
		SubcontractOrderID:  transfer.SubcontractOrderID,
		SubcontractOrderNo:  transfer.SubcontractOrderNo,
		FactoryID:           transfer.FactoryID,
		FactoryCode:         transfer.FactoryCode,
		FactoryName:         transfer.FactoryName,
		SourceWarehouseID:   transfer.SourceWarehouseID,
		SourceWarehouseCode: transfer.SourceWarehouseCode,
		Status:              string(transfer.Status),
		Lines:               make([]subcontractMaterialTransferLineResponse, 0, len(transfer.Lines)),
		Evidence:            make([]subcontractMaterialTransferEvidenceResponse, 0, len(transfer.Evidence)),
		HandoverBy:          transfer.HandoverBy,
		HandoverAt:          timeString(transfer.HandoverAt),
		ReceivedBy:          transfer.ReceivedBy,
		ReceiverContact:     transfer.ReceiverContact,
		VehicleNo:           transfer.VehicleNo,
		Note:                transfer.Note,
		CreatedAt:           timeString(transfer.CreatedAt),
		UpdatedAt:           timeString(transfer.UpdatedAt),
		Version:             transfer.Version,
	}
	for _, line := range transfer.Lines {
		payload.Lines = append(payload.Lines, subcontractMaterialTransferLineResponse{
			ID:                  line.ID,
			LineNo:              line.LineNo,
			OrderMaterialLineID: line.OrderMaterialLineID,
			ItemID:              line.ItemID,
			SKUCode:             line.SKUCode,
			ItemName:            line.ItemName,
			IssueQty:            line.IssueQty.String(),
			UOMCode:             line.UOMCode.String(),
			BaseIssueQty:        line.BaseIssueQty.String(),
			BaseUOMCode:         line.BaseUOMCode.String(),
			ConversionFactor:    line.ConversionFactor.String(),
			BatchID:             line.BatchID,
			BatchNo:             line.BatchNo,
			LotNo:               line.LotNo,
			SourceBinID:         line.SourceBinID,
			LotTraceRequired:    line.LotTraceRequired,
			Note:                line.Note,
		})
	}
	for _, evidence := range transfer.Evidence {
		payload.Evidence = append(payload.Evidence, subcontractMaterialTransferEvidenceResponse{
			ID:           evidence.ID,
			EvidenceType: evidence.EvidenceType,
			FileName:     evidence.FileName,
			ObjectKey:    evidence.ObjectKey,
			ExternalURL:  evidence.ExternalURL,
			Note:         evidence.Note,
		})
	}

	return payload
}

func newSubcontractFinishedGoodsReceiptResponse(
	receipt productiondomain.SubcontractFinishedGoodsReceipt,
) subcontractFinishedGoodsReceiptResponse {
	payload := subcontractFinishedGoodsReceiptResponse{
		ID:                 receipt.ID,
		OrgID:              receipt.OrgID,
		ReceiptNo:          receipt.ReceiptNo,
		SubcontractOrderID: receipt.SubcontractOrderID,
		SubcontractOrderNo: receipt.SubcontractOrderNo,
		FactoryID:          receipt.FactoryID,
		FactoryCode:        receipt.FactoryCode,
		FactoryName:        receipt.FactoryName,
		WarehouseID:        receipt.WarehouseID,
		WarehouseCode:      receipt.WarehouseCode,
		LocationID:         receipt.LocationID,
		LocationCode:       receipt.LocationCode,
		DeliveryNoteNo:     receipt.DeliveryNoteNo,
		Status:             string(receipt.Status),
		Lines:              make([]subcontractFinishedGoodsReceiptLineResponse, 0, len(receipt.Lines)),
		Evidence:           make([]subcontractFinishedGoodsReceiptEvidenceResponse, 0, len(receipt.Evidence)),
		ReceivedBy:         receipt.ReceivedBy,
		ReceivedAt:         timeString(receipt.ReceivedAt),
		Note:               receipt.Note,
		CreatedAt:          timeString(receipt.CreatedAt),
		UpdatedAt:          timeString(receipt.UpdatedAt),
		Version:            receipt.Version,
	}
	for _, line := range receipt.Lines {
		payload.Lines = append(payload.Lines, subcontractFinishedGoodsReceiptLineResponse{
			ID:               line.ID,
			LineNo:           line.LineNo,
			ItemID:           line.ItemID,
			SKUCode:          line.SKUCode,
			ItemName:         line.ItemName,
			BatchID:          line.BatchID,
			BatchNo:          line.BatchNo,
			LotNo:            line.LotNo,
			ExpiryDate:       dateString(line.ExpiryDate),
			ReceiveQty:       line.ReceiveQty.String(),
			UOMCode:          line.UOMCode.String(),
			BaseReceiveQty:   line.BaseReceiveQty.String(),
			BaseUOMCode:      line.BaseUOMCode.String(),
			ConversionFactor: line.ConversionFactor.String(),
			PackagingStatus:  line.PackagingStatus,
			Note:             line.Note,
		})
	}
	for _, evidence := range receipt.Evidence {
		payload.Evidence = append(payload.Evidence, subcontractFinishedGoodsReceiptEvidenceResponse{
			ID:           evidence.ID,
			EvidenceType: evidence.EvidenceType,
			FileName:     evidence.FileName,
			ObjectKey:    evidence.ObjectKey,
			ExternalURL:  evidence.ExternalURL,
			Note:         evidence.Note,
		})
	}

	return payload
}

func newSubcontractFactoryClaimResponse(
	claim productiondomain.SubcontractFactoryClaim,
) subcontractFactoryClaimResponse {
	payload := subcontractFactoryClaimResponse{
		ID:                 claim.ID,
		OrgID:              claim.OrgID,
		ClaimNo:            claim.ClaimNo,
		SubcontractOrderID: claim.SubcontractOrderID,
		SubcontractOrderNo: claim.SubcontractOrderNo,
		FactoryID:          claim.FactoryID,
		FactoryCode:        claim.FactoryCode,
		FactoryName:        claim.FactoryName,
		ReceiptID:          claim.ReceiptID,
		ReceiptNo:          claim.ReceiptNo,
		ReasonCode:         claim.ReasonCode,
		Reason:             claim.Reason,
		Severity:           claim.Severity,
		Status:             string(claim.Status),
		AffectedQty:        claim.AffectedQty.String(),
		UOMCode:            claim.UOMCode.String(),
		BaseAffectedQty:    claim.BaseAffectedQty.String(),
		BaseUOMCode:        claim.BaseUOMCode.String(),
		Evidence:           make([]subcontractFactoryClaimEvidenceResponse, 0, len(claim.Evidence)),
		OwnerID:            claim.OwnerID,
		OpenedBy:           claim.OpenedBy,
		OpenedAt:           timeString(claim.OpenedAt),
		DueAt:              timeString(claim.DueAt),
		AcknowledgedBy:     claim.AcknowledgedBy,
		AcknowledgedAt:     timeString(claim.AcknowledgedAt),
		ResolvedBy:         claim.ResolvedBy,
		ResolvedAt:         timeString(claim.ResolvedAt),
		ResolutionNote:     claim.ResolutionNote,
		BlocksFinalPayment: claim.BlocksFinalPayment(),
		CreatedAt:          timeString(claim.CreatedAt),
		UpdatedAt:          timeString(claim.UpdatedAt),
		Version:            claim.Version,
	}
	for _, evidence := range claim.Evidence {
		payload.Evidence = append(payload.Evidence, subcontractFactoryClaimEvidenceResponse{
			ID:           evidence.ID,
			EvidenceType: evidence.EvidenceType,
			FileName:     evidence.FileName,
			ObjectKey:    evidence.ObjectKey,
			ExternalURL:  evidence.ExternalURL,
			Note:         evidence.Note,
		})
	}

	return payload
}

func newSubcontractSampleApprovalResultResponse(
	result productionapp.SubcontractSampleApprovalResult,
) subcontractSampleApprovalResultResponse {
	return subcontractSampleApprovalResultResponse{
		SubcontractOrder: newSubcontractOrderResponse(result.SubcontractOrder, ""),
		SampleApproval:   newSubcontractSampleApprovalResponse(result.SampleApproval),
		PreviousStatus:   string(result.PreviousStatus),
		CurrentStatus:    string(result.CurrentStatus),
		AuditLogID:       result.AuditLogID,
	}
}

func newSubcontractSampleApprovalResponse(
	sampleApproval productiondomain.SubcontractSampleApproval,
) subcontractSampleApprovalResponse {
	payload := subcontractSampleApprovalResponse{
		ID:                 sampleApproval.ID,
		OrgID:              sampleApproval.OrgID,
		SubcontractOrderID: sampleApproval.SubcontractOrderID,
		SubcontractOrderNo: sampleApproval.SubcontractOrderNo,
		SampleCode:         sampleApproval.SampleCode,
		FormulaVersion:     sampleApproval.FormulaVersion,
		SpecVersion:        sampleApproval.SpecVersion,
		Status:             string(sampleApproval.Status),
		Evidence:           make([]subcontractSampleEvidenceResponse, 0, len(sampleApproval.Evidence)),
		SubmittedBy:        sampleApproval.SubmittedBy,
		SubmittedAt:        timeString(sampleApproval.SubmittedAt),
		DecisionBy:         sampleApproval.DecisionBy,
		DecisionAt:         timeString(sampleApproval.DecisionAt),
		DecisionReason:     sampleApproval.DecisionReason,
		StorageStatus:      sampleApproval.StorageStatus,
		Note:               sampleApproval.Note,
		CreatedAt:          timeString(sampleApproval.CreatedAt),
		UpdatedAt:          timeString(sampleApproval.UpdatedAt),
		Version:            sampleApproval.Version,
	}
	for _, evidence := range sampleApproval.Evidence {
		payload.Evidence = append(payload.Evidence, subcontractSampleEvidenceResponse{
			ID:           evidence.ID,
			EvidenceType: evidence.EvidenceType,
			FileName:     evidence.FileName,
			ObjectKey:    evidence.ObjectKey,
			ExternalURL:  evidence.ExternalURL,
			Note:         evidence.Note,
			CreatedAt:    timeString(evidence.CreatedAt),
			CreatedBy:    evidence.CreatedBy,
		})
	}

	return payload
}

func newSubcontractFactoryDispatchResultResponse(
	result productionapp.FactoryDispatchResult,
) subcontractFactoryDispatchResultResponse {
	return subcontractFactoryDispatchResultResponse{
		SubcontractOrder: newSubcontractOrderResponse(result.SubcontractOrder, ""),
		Dispatch:         newSubcontractFactoryDispatchResponse(result.Dispatch),
		AuditLogID:       result.AuditLogID,
	}
}

func newSubcontractFactoryDispatchResponse(
	dispatch productiondomain.SubcontractFactoryDispatch,
) subcontractFactoryDispatchResponse {
	payload := subcontractFactoryDispatchResponse{
		ID:                     dispatch.ID,
		OrgID:                  dispatch.OrgID,
		DispatchNo:             dispatch.DispatchNo,
		SubcontractOrderID:     dispatch.SubcontractOrderID,
		SubcontractOrderNo:     dispatch.SubcontractOrderNo,
		SourceProductionPlanID: dispatch.SourceProductionPlanID,
		SourceProductionPlanNo: dispatch.SourceProductionPlanNo,
		FactoryID:              dispatch.FactoryID,
		FactoryCode:            dispatch.FactoryCode,
		FactoryName:            dispatch.FactoryName,
		FinishedItemID:         dispatch.FinishedItemID,
		FinishedSKUCode:        dispatch.FinishedSKUCode,
		FinishedItemName:       dispatch.FinishedItemName,
		PlannedQty:             dispatch.PlannedQty.String(),
		UOMCode:                dispatch.UOMCode.String(),
		SpecSummary:            dispatch.SpecSummary,
		SampleRequired:         dispatch.SampleRequired,
		TargetStartDate:        dispatch.TargetStartDate,
		ExpectedReceiptDate:    dispatch.ExpectedReceiptDate,
		Status:                 string(dispatch.Status),
		Lines:                  make([]subcontractFactoryDispatchLineResponse, 0, len(dispatch.Lines)),
		Evidence:               make([]subcontractFactoryDispatchEvidenceResponse, 0, len(dispatch.Evidence)),
		ReadyAt:                timeString(dispatch.ReadyAt),
		ReadyBy:                dispatch.ReadyBy,
		SentAt:                 timeString(dispatch.SentAt),
		SentBy:                 dispatch.SentBy,
		RespondedAt:            timeString(dispatch.RespondedAt),
		ResponseBy:             dispatch.ResponseBy,
		FactoryResponseNote:    dispatch.FactoryResponseNote,
		Note:                   dispatch.Note,
		CreatedAt:              timeString(dispatch.CreatedAt),
		UpdatedAt:              timeString(dispatch.UpdatedAt),
		Version:                dispatch.Version,
	}
	for _, line := range dispatch.Lines {
		payload.Lines = append(payload.Lines, subcontractFactoryDispatchLineResponse{
			ID:                  line.ID,
			LineNo:              line.LineNo,
			OrderMaterialLineID: line.OrderMaterialLineID,
			ItemID:              line.ItemID,
			SKUCode:             line.SKUCode,
			ItemName:            line.ItemName,
			PlannedQty:          line.PlannedQty.String(),
			UOMCode:             line.UOMCode.String(),
			LotTraceRequired:    line.LotTraceRequired,
			Note:                line.Note,
		})
	}
	for _, evidence := range dispatch.Evidence {
		payload.Evidence = append(payload.Evidence, subcontractFactoryDispatchEvidenceResponse{
			ID:           evidence.ID,
			EvidenceType: evidence.EvidenceType,
			FileName:     evidence.FileName,
			ObjectKey:    evidence.ObjectKey,
			ExternalURL:  evidence.ExternalURL,
			Note:         evidence.Note,
		})
	}

	return payload
}

func newSubcontractMaterialIssueMovementResponses(
	movements []inventorydomain.StockMovement,
) []subcontractMaterialIssueMovementResponse {
	payload := make([]subcontractMaterialIssueMovementResponse, 0, len(movements))
	for _, movement := range movements {
		payload = append(payload, subcontractMaterialIssueMovementResponse{
			MovementNo:       movement.MovementNo,
			MovementType:     string(movement.MovementType),
			ItemID:           movement.ItemID,
			BatchID:          movement.BatchID,
			WarehouseID:      movement.WarehouseID,
			BinID:            movement.BinID,
			Quantity:         movement.Quantity.String(),
			BaseUOMCode:      movement.BaseUOMCode.String(),
			SourceQuantity:   movement.SourceQuantity.String(),
			SourceUOMCode:    movement.SourceUOMCode.String(),
			ConversionFactor: movement.ConversionFactor.String(),
			StockStatus:      string(movement.StockStatus),
			SourceDocType:    movement.SourceDocType,
			SourceDocID:      movement.SourceDocID,
			SourceDocLineID:  movement.SourceDocLineID,
			Reason:           movement.Reason,
		})
	}

	return payload
}

func writeSubcontractOrderError(w http.ResponseWriter, r *http.Request, err error) {
	if appErr, ok := apperrors.As(err); ok {
		response.WriteError(w, r, appErr.HTTPStatus, appErr.Code, appErr.Message, appErr.Details)
		return
	}

	response.WriteError(w, r, http.StatusConflict, response.ErrorCodeConflict, "Subcontract order request could not be processed", nil)
}
