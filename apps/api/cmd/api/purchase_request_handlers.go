package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

	productionapp "github.com/Chinsusu/ERP-v2/apps/api/internal/modules/production/application"
	productiondomain "github.com/Chinsusu/ERP-v2/apps/api/internal/modules/production/domain"
	purchaseapp "github.com/Chinsusu/ERP-v2/apps/api/internal/modules/purchase/application"
	"github.com/Chinsusu/ERP-v2/apps/api/internal/shared/auth"
	"github.com/Chinsusu/ERP-v2/apps/api/internal/shared/response"
)

type purchaseRequestActionRequest struct {
	ExpectedStatus string `json:"expected_status,omitempty"`
}

type convertPurchaseRequestToPurchaseOrderRequest struct {
	SupplierID   string `json:"supplier_id"`
	WarehouseID  string `json:"warehouse_id"`
	ExpectedDate string `json:"expected_date"`
	CurrencyCode string `json:"currency_code,omitempty"`
	UnitPrice    string `json:"unit_price,omitempty"`
}

type purchaseRequestActionResponse struct {
	PurchaseRequest purchaseRequestDraftResponse `json:"purchase_request"`
	PreviousStatus  string                       `json:"previous_status"`
	CurrentStatus   string                       `json:"current_status"`
	AuditLogID      string                       `json:"audit_log_id,omitempty"`
}

type convertPurchaseRequestToPurchaseOrderResponse struct {
	PurchaseRequest purchaseRequestDraftResponse `json:"purchase_request"`
	PurchaseOrder   purchaseOrderResponse        `json:"purchase_order"`
	AuditLogID      string                       `json:"audit_log_id,omitempty"`
}

func purchaseRequestsHandler(service productionapp.ProductionPlanService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		principal, ok := auth.PrincipalFromContext(r.Context())
		if !ok {
			response.WriteError(w, r, http.StatusUnauthorized, response.ErrorCodeUnauthorized, "Authentication required", nil)
			return
		}
		if !auth.HasPermission(principal, auth.PermissionPurchaseView) {
			writePermissionDenied(w, r, auth.PermissionPurchaseView)
			return
		}
		if r.Method != http.MethodGet {
			response.WriteError(w, r, http.StatusMethodNotAllowed, response.ErrorCodeNotFound, "Route not found", nil)
			return
		}

		drafts, err := service.ListPurchaseRequestDrafts(r.Context(), purchaseRequestDraftFilterFromRequest(r))
		if err != nil {
			writePurchaseRequestDraftError(w, r, err)
			return
		}
		payload := make([]purchaseRequestDraftResponse, 0, len(drafts))
		for _, draft := range drafts {
			payload = append(payload, newPurchaseRequestDraftResponse(draft))
		}
		response.WriteSuccess(w, r, http.StatusOK, payload)
	}
}

func purchaseRequestDetailHandler(service productionapp.ProductionPlanService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		principal, ok := auth.PrincipalFromContext(r.Context())
		if !ok {
			response.WriteError(w, r, http.StatusUnauthorized, response.ErrorCodeUnauthorized, "Authentication required", nil)
			return
		}
		if !auth.HasPermission(principal, auth.PermissionPurchaseView) {
			writePermissionDenied(w, r, auth.PermissionPurchaseView)
			return
		}
		if r.Method != http.MethodGet {
			response.WriteError(w, r, http.StatusMethodNotAllowed, response.ErrorCodeNotFound, "Route not found", nil)
			return
		}

		draft, err := service.GetPurchaseRequestDraft(r.Context(), r.PathValue("purchase_request_id"))
		if err != nil {
			writePurchaseRequestDraftError(w, r, err)
			return
		}
		response.WriteSuccess(w, r, http.StatusOK, newPurchaseRequestDraftResponse(draft))
	}
}

func purchaseRequestSubmitHandler(service productionapp.ProductionPlanService) http.HandlerFunc {
	return purchaseRequestActionHandler(service, "submit")
}

func purchaseRequestApproveHandler(service productionapp.ProductionPlanService) http.HandlerFunc {
	return purchaseRequestActionHandler(service, "approve")
}

func purchaseRequestActionHandler(service productionapp.ProductionPlanService, action string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		principal, ok := auth.PrincipalFromContext(r.Context())
		if !ok {
			response.WriteError(w, r, http.StatusUnauthorized, response.ErrorCodeUnauthorized, "Authentication required", nil)
			return
		}
		if !auth.HasPermission(principal, auth.PermissionPurchaseView) {
			writePermissionDenied(w, r, auth.PermissionPurchaseView)
			return
		}
		if !auth.HasPermission(principal, auth.PermissionRecordCreate) {
			writePermissionDenied(w, r, auth.PermissionRecordCreate)
			return
		}
		if r.Method != http.MethodPost {
			response.WriteError(w, r, http.StatusMethodNotAllowed, response.ErrorCodeNotFound, "Route not found", nil)
			return
		}

		r = requestWithStableID(r)
		var payload purchaseRequestActionRequest
		if r.Body != nil {
			if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
				if !errors.Is(err, io.EOF) {
					response.WriteError(w, r, http.StatusBadRequest, response.ErrorCodeValidation, "Invalid purchase request action payload", nil)
					return
				}
			}
		}
		input := productionapp.PurchaseRequestDraftActionInput{
			ID:             r.PathValue("purchase_request_id"),
			ExpectedStatus: productiondomain.PurchaseRequestDraftStatus(payload.ExpectedStatus),
			ActorID:        principal.UserID,
			RequestID:      response.RequestID(r),
		}
		var (
			result productionapp.PurchaseRequestDraftResult
			err    error
		)
		switch action {
		case "submit":
			result, err = service.SubmitPurchaseRequestDraft(r.Context(), input)
		case "approve":
			result, err = service.ApprovePurchaseRequestDraft(r.Context(), input)
		default:
			response.WriteError(w, r, http.StatusMethodNotAllowed, response.ErrorCodeNotFound, "Route not found", nil)
			return
		}
		if err != nil {
			writePurchaseRequestDraftError(w, r, err)
			return
		}
		response.WriteSuccess(w, r, http.StatusOK, purchaseRequestActionResponse{
			PurchaseRequest: newPurchaseRequestDraftResponse(result.PurchaseRequestDraft),
			PreviousStatus:  string(result.PreviousStatus),
			CurrentStatus:   string(result.CurrentStatus),
			AuditLogID:      result.AuditLogID,
		})
	}
}

func purchaseRequestConvertToPOHandler(
	productionPlans productionapp.ProductionPlanService,
	purchaseOrders purchaseapp.PurchaseOrderService,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		principal, ok := auth.PrincipalFromContext(r.Context())
		if !ok {
			response.WriteError(w, r, http.StatusUnauthorized, response.ErrorCodeUnauthorized, "Authentication required", nil)
			return
		}
		if !auth.HasPermission(principal, auth.PermissionPurchaseView) {
			writePermissionDenied(w, r, auth.PermissionPurchaseView)
			return
		}
		if !auth.HasPermission(principal, auth.PermissionRecordCreate) {
			writePermissionDenied(w, r, auth.PermissionRecordCreate)
			return
		}
		if r.Method != http.MethodPost {
			response.WriteError(w, r, http.StatusMethodNotAllowed, response.ErrorCodeNotFound, "Route not found", nil)
			return
		}

		r = requestWithStableID(r)
		var payload convertPurchaseRequestToPurchaseOrderRequest
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			response.WriteError(w, r, http.StatusBadRequest, response.ErrorCodeValidation, "Invalid purchase request conversion payload", nil)
			return
		}
		draft, err := productionPlans.GetPurchaseRequestDraft(r.Context(), r.PathValue("purchase_request_id"))
		if err != nil {
			writePurchaseRequestDraftError(w, r, err)
			return
		}
		if draft.Status != productiondomain.PurchaseRequestDraftStatusApproved {
			writePurchaseRequestDraftError(w, r, productiondomain.ErrProductionPlanInvalidPurchaseRequestTransition)
			return
		}

		orderResult, err := purchaseOrders.CreatePurchaseOrder(r.Context(), purchaseapp.CreatePurchaseOrderInput{
			ID:           purchaseOrderIDForPurchaseRequest(draft),
			PONo:         purchaseOrderNoForPurchaseRequest(draft),
			SupplierID:   payload.SupplierID,
			WarehouseID:  payload.WarehouseID,
			ExpectedDate: payload.ExpectedDate,
			CurrencyCode: firstNonBlankPurchaseRequest(payload.CurrencyCode, "VND"),
			Note:         purchaseOrderNoteForPurchaseRequest(draft),
			Lines:        purchaseOrderLinesForPurchaseRequest(draft, payload),
			ActorID:      principal.UserID,
			RequestID:    response.RequestID(r),
		})
		if err != nil {
			writePurchaseOrderError(w, r, err)
			return
		}
		converted, err := productionPlans.MarkPurchaseRequestDraftConverted(r.Context(), productionapp.ConvertPurchaseRequestDraftInput{
			ID:              draft.ID,
			PurchaseOrderID: orderResult.PurchaseOrder.ID,
			PurchaseOrderNo: orderResult.PurchaseOrder.PONo,
			ActorID:         principal.UserID,
			RequestID:       response.RequestID(r),
		})
		if err != nil {
			writePurchaseRequestDraftError(w, r, err)
			return
		}

		response.WriteSuccess(w, r, http.StatusCreated, convertPurchaseRequestToPurchaseOrderResponse{
			PurchaseRequest: newPurchaseRequestDraftResponse(converted.PurchaseRequestDraft),
			PurchaseOrder:   newPurchaseOrderResponse(orderResult.PurchaseOrder, orderResult.AuditLogID),
			AuditLogID:      converted.AuditLogID,
		})
	}
}

func purchaseRequestDraftFilterFromRequest(r *http.Request) productionapp.PurchaseRequestDraftFilter {
	query := r.URL.Query()
	statuses := make([]productiondomain.PurchaseRequestDraftStatus, 0)
	for _, value := range query["status"] {
		for _, part := range strings.Split(value, ",") {
			if trimmed := strings.TrimSpace(part); trimmed != "" {
				statuses = append(statuses, productiondomain.PurchaseRequestDraftStatus(trimmed))
			}
		}
	}

	return productionapp.PurchaseRequestDraftFilter{
		Search:                 query.Get("q"),
		Statuses:               statuses,
		SourceProductionPlanID: query.Get("source_production_plan_id"),
	}
}

func purchaseOrderLinesForPurchaseRequest(
	draft productiondomain.PurchaseRequestDraft,
	payload convertPurchaseRequestToPurchaseOrderRequest,
) []purchaseapp.PurchaseOrderLineInput {
	lines := make([]purchaseapp.PurchaseOrderLineInput, 0, len(draft.Lines))
	for _, line := range draft.Lines {
		lines = append(lines, purchaseapp.PurchaseOrderLineInput{
			ID:           purchaseOrderLineIDForPurchaseRequest(draft, line.LineNo),
			LineNo:       line.LineNo,
			ItemID:       line.ItemID,
			OrderedQty:   line.RequestedQty.String(),
			UOMCode:      line.UOMCode.String(),
			UnitPrice:    firstNonBlankPurchaseRequest(payload.UnitPrice, "0"),
			CurrencyCode: firstNonBlankPurchaseRequest(payload.CurrencyCode, "VND"),
			ExpectedDate: payload.ExpectedDate,
			Note:         "Tu " + draft.RequestNo + " dong " + strings.TrimSpace(line.SKU),
		})
	}

	return lines
}

func purchaseOrderIDForPurchaseRequest(draft productiondomain.PurchaseRequestDraft) string {
	return "po-" + purchaseRequestRefSuffix(draft)
}

func purchaseOrderLineIDForPurchaseRequest(draft productiondomain.PurchaseRequestDraft, lineNo int) string {
	return fmt.Sprintf("%s-line-%02d", purchaseOrderIDForPurchaseRequest(draft), lineNo)
}

func purchaseOrderNoForPurchaseRequest(draft productiondomain.PurchaseRequestDraft) string {
	return "PO-" + strings.ToUpper(purchaseRequestRefSuffix(draft))
}

func purchaseRequestRefSuffix(draft productiondomain.PurchaseRequestDraft) string {
	value := strings.ToLower(strings.TrimSpace(draft.ID))
	value = strings.TrimPrefix(value, "pr-draft-")
	value = strings.TrimPrefix(value, "pr-")
	if value == "" {
		value = strings.ToLower(strings.TrimSpace(draft.RequestNo))
		value = strings.TrimPrefix(value, "pr-draft-")
		value = strings.TrimPrefix(value, "pr-")
	}
	return value
}

func purchaseOrderNoteForPurchaseRequest(draft productiondomain.PurchaseRequestDraft) string {
	return "Tao tu de nghi mua " + draft.RequestNo + " / ke hoach san xuat " + draft.SourceProductionPlanNo
}

func firstNonBlankPurchaseRequest(values ...string) string {
	for _, value := range values {
		if trimmed := strings.TrimSpace(value); trimmed != "" {
			return trimmed
		}
	}

	return ""
}

func writePurchaseRequestDraftError(w http.ResponseWriter, r *http.Request, err error) {
	switch {
	case errors.Is(err, productionapp.ErrPurchaseRequestDraftNotFound):
		response.WriteError(w, r, http.StatusNotFound, productionapp.ErrorCodePurchaseRequestNotFound, "Purchase request was not found", nil)
	case errors.Is(err, productiondomain.ErrProductionPlanInvalidPurchaseRequestTransition):
		response.WriteError(w, r, http.StatusConflict, productionapp.ErrorCodePurchaseRequestInvalid, "Purchase request status transition is invalid", nil)
	case errors.Is(err, productiondomain.ErrProductionPlanRequiredField):
		response.WriteError(w, r, http.StatusBadRequest, productionapp.ErrorCodeProductionPlanValidation, "Purchase request payload is invalid", nil)
	default:
		writeProductionPlanError(w, r, err)
	}
}
