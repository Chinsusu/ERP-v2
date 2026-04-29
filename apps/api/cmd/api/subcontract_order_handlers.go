package main

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"

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

func writeSubcontractOrderError(w http.ResponseWriter, r *http.Request, err error) {
	if appErr, ok := apperrors.As(err); ok {
		response.WriteError(w, r, appErr.HTTPStatus, appErr.Code, appErr.Message, appErr.Details)
		return
	}

	response.WriteError(w, r, http.StatusConflict, response.ErrorCodeConflict, "Subcontract order request could not be processed", nil)
}
