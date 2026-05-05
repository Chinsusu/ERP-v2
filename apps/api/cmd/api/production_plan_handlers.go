package main

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	productionapp "github.com/Chinsusu/ERP-v2/apps/api/internal/modules/production/application"
	productiondomain "github.com/Chinsusu/ERP-v2/apps/api/internal/modules/production/domain"
	"github.com/Chinsusu/ERP-v2/apps/api/internal/shared/auth"
	"github.com/Chinsusu/ERP-v2/apps/api/internal/shared/response"
)

type createProductionPlanRequest struct {
	ID               string `json:"id"`
	PlanNo           string `json:"plan_no"`
	OutputItemID     string `json:"output_item_id"`
	FormulaID        string `json:"formula_id"`
	PlannedQty       string `json:"planned_qty"`
	UOMCode          string `json:"uom_code"`
	PlannedStartDate string `json:"planned_start_date"`
	PlannedEndDate   string `json:"planned_end_date"`
}

type productionPlanLineResponse struct {
	ID                   string `json:"id"`
	FormulaLineID        string `json:"formula_line_id"`
	LineNo               int    `json:"line_no"`
	ComponentItemID      string `json:"component_item_id,omitempty"`
	ComponentSKU         string `json:"component_sku"`
	ComponentName        string `json:"component_name"`
	ComponentType        string `json:"component_type"`
	FormulaQty           string `json:"formula_qty"`
	FormulaUOMCode       string `json:"formula_uom_code"`
	RequiredQty          string `json:"required_qty"`
	RequiredUOMCode      string `json:"required_uom_code"`
	RequiredStockBaseQty string `json:"required_stock_base_qty"`
	StockBaseUOMCode     string `json:"stock_base_uom_code"`
	AvailableQty         string `json:"available_qty"`
	ShortageQty          string `json:"shortage_qty"`
	PurchaseDraftQty     string `json:"purchase_draft_qty"`
	PurchaseDraftUOMCode string `json:"purchase_draft_uom_code"`
	IsStockManaged       bool   `json:"is_stock_managed"`
	NeedsPurchase        bool   `json:"needs_purchase"`
	Note                 string `json:"note,omitempty"`
}

type purchaseRequestDraftLineResponse struct {
	ID                       string `json:"id"`
	LineNo                   int    `json:"line_no"`
	SourceProductionPlanLine string `json:"source_production_plan_line_id"`
	ItemID                   string `json:"item_id,omitempty"`
	SKU                      string `json:"sku"`
	ItemName                 string `json:"item_name"`
	RequestedQty             string `json:"requested_qty"`
	UOMCode                  string `json:"uom_code"`
	Note                     string `json:"note,omitempty"`
}

type purchaseRequestDraftResponse struct {
	ID                       string                             `json:"id,omitempty"`
	RequestNo                string                             `json:"request_no,omitempty"`
	SourceProductionPlanID   string                             `json:"source_production_plan_id,omitempty"`
	SourceProductionPlanNo   string                             `json:"source_production_plan_no,omitempty"`
	Status                   string                             `json:"status,omitempty"`
	Lines                    []purchaseRequestDraftLineResponse `json:"lines"`
	CreatedAt                string                             `json:"created_at,omitempty"`
	CreatedBy                string                             `json:"created_by,omitempty"`
	SubmittedAt              string                             `json:"submitted_at,omitempty"`
	SubmittedBy              string                             `json:"submitted_by,omitempty"`
	ApprovedAt               string                             `json:"approved_at,omitempty"`
	ApprovedBy               string                             `json:"approved_by,omitempty"`
	ConvertedAt              string                             `json:"converted_at,omitempty"`
	ConvertedBy              string                             `json:"converted_by,omitempty"`
	ConvertedPurchaseOrderID string                             `json:"converted_purchase_order_id,omitempty"`
	ConvertedPurchaseOrderNo string                             `json:"converted_purchase_order_no,omitempty"`
	CancelledAt              string                             `json:"cancelled_at,omitempty"`
	CancelledBy              string                             `json:"cancelled_by,omitempty"`
	RejectedAt               string                             `json:"rejected_at,omitempty"`
	RejectedBy               string                             `json:"rejected_by,omitempty"`
	RejectReason             string                             `json:"reject_reason,omitempty"`
}

type productionPlanResponse struct {
	ID                   string                       `json:"id"`
	OrgID                string                       `json:"org_id"`
	PlanNo               string                       `json:"plan_no"`
	OutputItemID         string                       `json:"output_item_id"`
	OutputSKU            string                       `json:"output_sku"`
	OutputItemName       string                       `json:"output_item_name"`
	OutputItemType       string                       `json:"output_item_type"`
	PlannedQty           string                       `json:"planned_qty"`
	UOMCode              string                       `json:"uom_code"`
	FormulaID            string                       `json:"formula_id"`
	FormulaCode          string                       `json:"formula_code"`
	FormulaVersion       string                       `json:"formula_version"`
	FormulaBatchQty      string                       `json:"formula_batch_qty"`
	FormulaBatchUOMCode  string                       `json:"formula_batch_uom_code"`
	PlannedStartDate     string                       `json:"planned_start_date,omitempty"`
	PlannedEndDate       string                       `json:"planned_end_date,omitempty"`
	Status               string                       `json:"status"`
	Lines                []productionPlanLineResponse `json:"lines"`
	PurchaseRequestDraft purchaseRequestDraftResponse `json:"purchase_request_draft"`
	AuditLogID           string                       `json:"audit_log_id,omitempty"`
	CreatedAt            string                       `json:"created_at"`
	CreatedBy            string                       `json:"created_by"`
	UpdatedAt            string                       `json:"updated_at"`
	UpdatedBy            string                       `json:"updated_by"`
	Version              int                          `json:"version"`
}

func productionPlansHandler(service productionapp.ProductionPlanService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		principal, ok := auth.PrincipalFromContext(r.Context())
		if !ok {
			response.WriteError(w, r, http.StatusUnauthorized, response.ErrorCodeUnauthorized, "Authentication required", nil)
			return
		}
		if !auth.HasPermission(principal, auth.PermissionProductionView) {
			writePermissionDenied(w, r, auth.PermissionProductionView)
			return
		}

		switch r.Method {
		case http.MethodGet:
			plans, err := service.ListProductionPlans(r.Context(), productionPlanFilterFromRequest(r))
			if err != nil {
				writeProductionPlanError(w, r, err)
				return
			}
			payload := make([]productionPlanResponse, 0, len(plans))
			for _, plan := range plans {
				payload = append(payload, newProductionPlanResponse(plan, ""))
			}
			response.WriteSuccess(w, r, http.StatusOK, payload)
		case http.MethodPost:
			if !auth.HasPermission(principal, auth.PermissionRecordCreate) {
				writePermissionDenied(w, r, auth.PermissionRecordCreate)
				return
			}
			r = requestWithStableID(r)
			var payload createProductionPlanRequest
			if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
				response.WriteError(w, r, http.StatusBadRequest, response.ErrorCodeValidation, "Invalid production plan payload", nil)
				return
			}
			result, err := service.CreateProductionPlan(r.Context(), productionapp.CreateProductionPlanInput{
				ID:               payload.ID,
				PlanNo:           payload.PlanNo,
				OutputItemID:     payload.OutputItemID,
				FormulaID:        payload.FormulaID,
				PlannedQty:       payload.PlannedQty,
				UOMCode:          payload.UOMCode,
				PlannedStartDate: payload.PlannedStartDate,
				PlannedEndDate:   payload.PlannedEndDate,
				ActorID:          principal.UserID,
				RequestID:        response.RequestID(r),
			})
			if err != nil {
				writeProductionPlanError(w, r, err)
				return
			}
			response.WriteSuccess(w, r, http.StatusCreated, newProductionPlanResponse(result.ProductionPlan, result.AuditLogID))
		default:
			response.WriteError(w, r, http.StatusMethodNotAllowed, response.ErrorCodeNotFound, "Route not found", nil)
		}
	}
}

func productionPlanDetailHandler(service productionapp.ProductionPlanService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		principal, ok := auth.PrincipalFromContext(r.Context())
		if !ok {
			response.WriteError(w, r, http.StatusUnauthorized, response.ErrorCodeUnauthorized, "Authentication required", nil)
			return
		}
		if !auth.HasPermission(principal, auth.PermissionProductionView) {
			writePermissionDenied(w, r, auth.PermissionProductionView)
			return
		}
		if r.Method != http.MethodGet {
			response.WriteError(w, r, http.StatusMethodNotAllowed, response.ErrorCodeNotFound, "Route not found", nil)
			return
		}

		plan, err := service.GetProductionPlan(r.Context(), r.PathValue("production_plan_id"))
		if err != nil {
			writeProductionPlanError(w, r, err)
			return
		}
		response.WriteSuccess(w, r, http.StatusOK, newProductionPlanResponse(plan, ""))
	}
}

func productionPlanFilterFromRequest(r *http.Request) productionapp.ProductionPlanFilter {
	query := r.URL.Query()
	statuses := make([]productiondomain.ProductionPlanStatus, 0)
	for _, value := range query["status"] {
		for _, part := range strings.Split(value, ",") {
			if trimmed := strings.TrimSpace(part); trimmed != "" {
				statuses = append(statuses, productiondomain.ProductionPlanStatus(trimmed))
			}
		}
	}

	return productionapp.ProductionPlanFilter{
		Search:       query.Get("q"),
		Statuses:     statuses,
		OutputItemID: query.Get("output_item_id"),
	}
}

func newProductionPlanResponse(plan productiondomain.ProductionPlan, auditLogID string) productionPlanResponse {
	lines := make([]productionPlanLineResponse, 0, len(plan.Lines))
	for _, line := range plan.Lines {
		lines = append(lines, productionPlanLineResponse{
			ID:                   line.ID,
			FormulaLineID:        line.FormulaLineID,
			LineNo:               line.LineNo,
			ComponentItemID:      line.ComponentItemID,
			ComponentSKU:         line.ComponentSKU,
			ComponentName:        line.ComponentName,
			ComponentType:        line.ComponentType,
			FormulaQty:           line.FormulaQty.String(),
			FormulaUOMCode:       line.FormulaUOMCode.String(),
			RequiredQty:          line.RequiredQty.String(),
			RequiredUOMCode:      line.RequiredUOMCode.String(),
			RequiredStockBaseQty: line.RequiredStockBaseQty.String(),
			StockBaseUOMCode:     line.StockBaseUOMCode.String(),
			AvailableQty:         line.AvailableQty.String(),
			ShortageQty:          line.ShortageQty.String(),
			PurchaseDraftQty:     line.PurchaseDraftQty.String(),
			PurchaseDraftUOMCode: line.PurchaseDraftUOMCode.String(),
			IsStockManaged:       line.IsStockManaged,
			NeedsPurchase:        line.NeedsPurchase,
			Note:                 line.Note,
		})
	}

	return productionPlanResponse{
		ID:                   plan.ID,
		OrgID:                plan.OrgID,
		PlanNo:               plan.PlanNo,
		OutputItemID:         plan.OutputItemID,
		OutputSKU:            plan.OutputSKU,
		OutputItemName:       plan.OutputItemName,
		OutputItemType:       plan.OutputItemType,
		PlannedQty:           plan.PlannedQty.String(),
		UOMCode:              plan.UOMCode.String(),
		FormulaID:            plan.FormulaID,
		FormulaCode:          plan.FormulaCode,
		FormulaVersion:       plan.FormulaVersion,
		FormulaBatchQty:      plan.FormulaBatchQty.String(),
		FormulaBatchUOMCode:  plan.FormulaBatchUOMCode.String(),
		PlannedStartDate:     plan.PlannedStartDate,
		PlannedEndDate:       plan.PlannedEndDate,
		Status:               string(plan.Status),
		Lines:                lines,
		PurchaseRequestDraft: newPurchaseRequestDraftResponse(plan.PurchaseDraft),
		AuditLogID:           auditLogID,
		CreatedAt:            timeString(plan.CreatedAt),
		CreatedBy:            plan.CreatedBy,
		UpdatedAt:            timeString(plan.UpdatedAt),
		UpdatedBy:            plan.UpdatedBy,
		Version:              plan.Version,
	}
}

func newPurchaseRequestDraftResponse(draft productiondomain.PurchaseRequestDraft) purchaseRequestDraftResponse {
	lines := make([]purchaseRequestDraftLineResponse, 0, len(draft.Lines))
	for _, line := range draft.Lines {
		lines = append(lines, purchaseRequestDraftLineResponse{
			ID:                       line.ID,
			LineNo:                   line.LineNo,
			SourceProductionPlanLine: line.SourceProductionPlanLine,
			ItemID:                   line.ItemID,
			SKU:                      line.SKU,
			ItemName:                 line.ItemName,
			RequestedQty:             line.RequestedQty.String(),
			UOMCode:                  line.UOMCode.String(),
			Note:                     line.Note,
		})
	}

	return purchaseRequestDraftResponse{
		ID:                       draft.ID,
		RequestNo:                draft.RequestNo,
		SourceProductionPlanID:   draft.SourceProductionPlanID,
		SourceProductionPlanNo:   draft.SourceProductionPlanNo,
		Status:                   string(draft.Status),
		Lines:                    lines,
		CreatedAt:                timeString(draft.CreatedAt),
		CreatedBy:                draft.CreatedBy,
		SubmittedAt:              timeString(draft.SubmittedAt),
		SubmittedBy:              draft.SubmittedBy,
		ApprovedAt:               timeString(draft.ApprovedAt),
		ApprovedBy:               draft.ApprovedBy,
		ConvertedAt:              timeString(draft.ConvertedAt),
		ConvertedBy:              draft.ConvertedBy,
		ConvertedPurchaseOrderID: draft.ConvertedPurchaseOrderID,
		ConvertedPurchaseOrderNo: draft.ConvertedPurchaseOrderNo,
		CancelledAt:              timeString(draft.CancelledAt),
		CancelledBy:              draft.CancelledBy,
		RejectedAt:               timeString(draft.RejectedAt),
		RejectedBy:               draft.RejectedBy,
		RejectReason:             draft.RejectReason,
	}
}

func writeProductionPlanError(w http.ResponseWriter, r *http.Request, err error) {
	switch {
	case errors.Is(err, productionapp.ErrProductionPlanNotFound):
		response.WriteError(w, r, http.StatusNotFound, productionapp.ErrorCodeProductionPlanNotFound, "Production plan was not found", nil)
	case errors.Is(err, productionapp.ErrPurchaseRequestDraftNotFound):
		response.WriteError(w, r, http.StatusNotFound, productionapp.ErrorCodePurchaseRequestNotFound, "Purchase request was not found", nil)
	case errors.Is(err, productionapp.ErrProductionPlanFormulaNotFound):
		response.WriteError(w, r, http.StatusBadRequest, productionapp.ErrorCodeProductionPlanValidation, "Active formula was not found for the output item", nil)
	case errors.Is(err, productionapp.ErrProductionPlanFormulaInactive):
		response.WriteError(w, r, http.StatusBadRequest, productionapp.ErrorCodeProductionPlanValidation, "Formula must be active before production planning", nil)
	case errors.Is(err, productiondomain.ErrProductionPlanRequiredField),
		errors.Is(err, productiondomain.ErrProductionPlanInvalidOutputType),
		errors.Is(err, productiondomain.ErrProductionPlanInvalidQuantity),
		errors.Is(err, productiondomain.ErrProductionPlanInvalidUOM),
		errors.Is(err, productiondomain.ErrProductionPlanInvalidComponentType),
		errors.Is(err, productiondomain.ErrProductionPlanInvalidShortage):
		response.WriteError(w, r, http.StatusBadRequest, productionapp.ErrorCodeProductionPlanValidation, "Production plan payload is invalid", nil)
	case errors.Is(err, productiondomain.ErrProductionPlanInvalidPurchaseRequestTransition):
		response.WriteError(w, r, http.StatusConflict, productionapp.ErrorCodePurchaseRequestInvalid, "Purchase request status transition is invalid", nil)
	default:
		response.WriteError(w, r, http.StatusInternalServerError, response.ErrorCodeInvalidState, "Production plan could not be processed", nil)
	}
}
