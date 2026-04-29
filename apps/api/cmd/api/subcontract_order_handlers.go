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
	catalog *masterdataapp.UOMCatalog
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
	ID                  string                                `json:"id"`
	OrderNo             string                                `json:"order_no"`
	FactoryID           string                                `json:"factory_id"`
	FinishedItemID      string                                `json:"finished_item_id"`
	PlannedQty          string                                `json:"planned_qty"`
	UOMCode             string                                `json:"uom_code"`
	CurrencyCode        string                                `json:"currency_code"`
	SpecSummary         string                                `json:"spec_summary"`
	SampleRequired      bool                                  `json:"sample_required"`
	ClaimWindowDays     int                                   `json:"claim_window_days"`
	TargetStartDate     string                                `json:"target_start_date"`
	ExpectedReceiptDate string                                `json:"expected_receipt_date"`
	MaterialLines       []subcontractOrderMaterialLineRequest `json:"material_lines"`
}

type updateSubcontractOrderRequest struct {
	FactoryID           string                                `json:"factory_id"`
	FinishedItemID      string                                `json:"finished_item_id"`
	PlannedQty          string                                `json:"planned_qty"`
	UOMCode             string                                `json:"uom_code"`
	SpecSummary         string                                `json:"spec_summary"`
	SampleRequired      *bool                                 `json:"sample_required"`
	ClaimWindowDays     int                                   `json:"claim_window_days"`
	TargetStartDate     string                                `json:"target_start_date"`
	ExpectedReceiptDate string                                `json:"expected_receipt_date"`
	ExpectedVersion     int                                   `json:"expected_version"`
	MaterialLines       []subcontractOrderMaterialLineRequest `json:"material_lines"`
}

type subcontractOrderActionRequest struct {
	ExpectedVersion int    `json:"expected_version"`
	Reason          string `json:"reason"`
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
	ID                  string `json:"id"`
	OrderNo             string `json:"order_no"`
	FactoryID           string `json:"factory_id"`
	FactoryCode         string `json:"factory_code,omitempty"`
	FactoryName         string `json:"factory_name"`
	FinishedItemID      string `json:"finished_item_id"`
	FinishedSKUCode     string `json:"finished_sku_code"`
	FinishedItemName    string `json:"finished_item_name"`
	PlannedQty          string `json:"planned_qty"`
	UOMCode             string `json:"uom_code"`
	ExpectedReceiptDate string `json:"expected_receipt_date"`
	Status              string `json:"status"`
	CurrencyCode        string `json:"currency_code"`
	EstimatedCostAmount string `json:"estimated_cost_amount"`
	MaterialLineCount   int    `json:"material_line_count"`
	SampleRequired      bool   `json:"sample_required"`
	CreatedAt           string `json:"created_at"`
	UpdatedAt           string `json:"updated_at"`
	Version             int    `json:"version"`
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
				ID:                  payload.ID,
				OrderNo:             payload.OrderNo,
				FactoryID:           payload.FactoryID,
				FinishedItemID:      payload.FinishedItemID,
				PlannedQty:          payload.PlannedQty,
				UOMCode:             payload.UOMCode,
				CurrencyCode:        payload.CurrencyCode,
				SpecSummary:         payload.SpecSummary,
				SampleRequired:      payload.SampleRequired,
				ClaimWindowDays:     payload.ClaimWindowDays,
				TargetStartDate:     payload.TargetStartDate,
				ExpectedReceiptDate: payload.ExpectedReceiptDate,
				MaterialLines:       subcontractOrderMaterialLineInputs(payload.MaterialLines),
				ActorID:             principal.UserID,
				RequestID:           response.RequestID(r),
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
				ID:                  r.PathValue("subcontract_order_id"),
				FactoryID:           payload.FactoryID,
				FinishedItemID:      payload.FinishedItemID,
				PlannedQty:          payload.PlannedQty,
				UOMCode:             payload.UOMCode,
				SpecSummary:         payload.SpecSummary,
				SampleRequired:      payload.SampleRequired,
				ClaimWindowDays:     payload.ClaimWindowDays,
				TargetStartDate:     payload.TargetStartDate,
				ExpectedReceiptDate: payload.ExpectedReceiptDate,
				MaterialLines:       subcontractOrderMaterialLineInputs(payload.MaterialLines),
				ExpectedVersion:     payload.ExpectedVersion,
				ActorID:             principal.UserID,
				RequestID:           response.RequestID(r),
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

func subcontractOrderCancelHandler(service productionapp.SubcontractOrderService) http.HandlerFunc {
	return subcontractOrderActionHandler(service, "cancel")
}

func subcontractOrderCloseHandler(service productionapp.SubcontractOrderService) http.HandlerFunc {
	return subcontractOrderActionHandler(service, "close")
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
		Search:              query.Get("search"),
		Statuses:            statuses,
		FactoryID:           query.Get("factory_id"),
		FinishedItemID:      query.Get("finished_item_id"),
		ExpectedReceiptFrom: query.Get("expected_receipt_from"),
		ExpectedReceiptTo:   query.Get("expected_receipt_to"),
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

func parseSubcontractOptionalTime(value string) (time.Time, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return time.Time{}, nil
	}

	return time.Parse(time.RFC3339, value)
}

func newSubcontractOrderListItemResponse(order productiondomain.SubcontractOrder) subcontractOrderListItemResponse {
	return subcontractOrderListItemResponse{
		ID:                  order.ID,
		OrderNo:             order.OrderNo,
		FactoryID:           order.FactoryID,
		FactoryCode:         order.FactoryCode,
		FactoryName:         order.FactoryName,
		FinishedItemID:      order.FinishedItemID,
		FinishedSKUCode:     order.FinishedSKUCode,
		FinishedItemName:    order.FinishedItemName,
		PlannedQty:          order.PlannedQty.String(),
		UOMCode:             order.UOMCode.String(),
		ExpectedReceiptDate: order.ExpectedReceiptDate,
		Status:              string(order.Status),
		CurrencyCode:        order.CurrencyCode.String(),
		EstimatedCostAmount: order.EstimatedCostAmount.String(),
		MaterialLineCount:   len(order.MaterialLines),
		SampleRequired:      order.SampleRequired,
		CreatedAt:           timeString(order.CreatedAt),
		UpdatedAt:           timeString(order.UpdatedAt),
		Version:             order.Version,
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
