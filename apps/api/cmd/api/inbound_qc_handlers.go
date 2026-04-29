package main

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"

	qcapp "github.com/Chinsusu/ERP-v2/apps/api/internal/modules/qc/application"
	qcdomain "github.com/Chinsusu/ERP-v2/apps/api/internal/modules/qc/domain"
	"github.com/Chinsusu/ERP-v2/apps/api/internal/shared/auth"
	"github.com/Chinsusu/ERP-v2/apps/api/internal/shared/decimal"
	"github.com/Chinsusu/ERP-v2/apps/api/internal/shared/response"
)

type inboundQCChecklistItemRequest struct {
	ID       string `json:"id"`
	Code     string `json:"code"`
	Label    string `json:"label"`
	Required bool   `json:"required"`
	Status   string `json:"status"`
	Note     string `json:"note"`
}

type createInboundQCInspectionRequest struct {
	ID                 string                          `json:"id"`
	OrgID              string                          `json:"org_id"`
	GoodsReceiptID     string                          `json:"goods_receipt_id"`
	GoodsReceiptLineID string                          `json:"goods_receipt_line_id"`
	InspectorID        string                          `json:"inspector_id"`
	Checklist          []inboundQCChecklistItemRequest `json:"checklist"`
	Note               string                          `json:"note"`
}

type inboundQCDecisionRequest struct {
	PassedQuantity string                          `json:"passed_qty"`
	FailedQuantity string                          `json:"failed_qty"`
	HoldQuantity   string                          `json:"hold_qty"`
	Checklist      []inboundQCChecklistItemRequest `json:"checklist"`
	Reason         string                          `json:"reason"`
	Note           string                          `json:"note"`
}

type inboundQCChecklistItemResponse struct {
	ID       string `json:"id"`
	Code     string `json:"code"`
	Label    string `json:"label"`
	Required bool   `json:"required"`
	Status   string `json:"status"`
	Note     string `json:"note,omitempty"`
}

type inboundQCInspectionResponse struct {
	ID                  string                           `json:"id"`
	OrgID               string                           `json:"org_id"`
	GoodsReceiptID      string                           `json:"goods_receipt_id"`
	GoodsReceiptNo      string                           `json:"goods_receipt_no"`
	GoodsReceiptLineID  string                           `json:"goods_receipt_line_id"`
	PurchaseOrderID     string                           `json:"purchase_order_id"`
	PurchaseOrderLineID string                           `json:"purchase_order_line_id"`
	ItemID              string                           `json:"item_id"`
	SKU                 string                           `json:"sku"`
	ItemName            string                           `json:"item_name"`
	BatchID             string                           `json:"batch_id"`
	BatchNo             string                           `json:"batch_no"`
	LotNo               string                           `json:"lot_no"`
	ExpiryDate          string                           `json:"expiry_date"`
	WarehouseID         string                           `json:"warehouse_id"`
	LocationID          string                           `json:"location_id"`
	Quantity            string                           `json:"quantity"`
	UOMCode             string                           `json:"uom_code"`
	InspectorID         string                           `json:"inspector_id"`
	Status              string                           `json:"status"`
	Result              string                           `json:"result,omitempty"`
	PassedQuantity      string                           `json:"passed_qty"`
	FailedQuantity      string                           `json:"failed_qty"`
	HoldQuantity        string                           `json:"hold_qty"`
	Checklist           []inboundQCChecklistItemResponse `json:"checklist"`
	Reason              string                           `json:"reason,omitempty"`
	Note                string                           `json:"note,omitempty"`
	AuditLogID          string                           `json:"audit_log_id,omitempty"`
	CreatedAt           string                           `json:"created_at"`
	CreatedBy           string                           `json:"created_by"`
	UpdatedAt           string                           `json:"updated_at"`
	UpdatedBy           string                           `json:"updated_by"`
	StartedAt           string                           `json:"started_at,omitempty"`
	StartedBy           string                           `json:"started_by,omitempty"`
	DecidedAt           string                           `json:"decided_at,omitempty"`
	DecidedBy           string                           `json:"decided_by,omitempty"`
}

type inboundQCActionResultResponse struct {
	Inspection     inboundQCInspectionResponse `json:"inspection"`
	PreviousStatus string                      `json:"previous_status,omitempty"`
	CurrentStatus  string                      `json:"current_status"`
	PreviousResult string                      `json:"previous_result,omitempty"`
	CurrentResult  string                      `json:"current_result,omitempty"`
	AuditLogID     string                      `json:"audit_log_id,omitempty"`
}

func inboundQCInspectionsHandler(service qcapp.InboundQCInspectionService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		principal, ok := auth.PrincipalFromContext(r.Context())
		if !ok {
			response.WriteError(w, r, http.StatusUnauthorized, response.ErrorCodeUnauthorized, "Authentication required", nil)
			return
		}
		if !auth.HasPermission(principal, auth.PermissionQCView) {
			writePermissionDenied(w, r, auth.PermissionQCView)
			return
		}

		switch r.Method {
		case http.MethodGet:
			rows, err := service.ListInboundQCInspections(r.Context(), inboundQCInspectionFilterFromRequest(r))
			if err != nil {
				writeInboundQCInspectionError(w, r, err)
				return
			}
			payload := make([]inboundQCInspectionResponse, 0, len(rows))
			for _, row := range rows {
				payload = append(payload, newInboundQCInspectionResponse(row, ""))
			}
			response.WriteSuccess(w, r, http.StatusOK, payload)
		case http.MethodPost:
			if !auth.HasPermission(principal, auth.PermissionQCDecision) {
				writePermissionDenied(w, r, auth.PermissionQCDecision)
				return
			}
			r = requestWithStableID(r)
			var payload createInboundQCInspectionRequest
			if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
				response.WriteError(w, r, http.StatusBadRequest, response.ErrorCodeValidation, "Invalid inbound QC inspection payload", nil)
				return
			}

			result, err := service.CreateInboundQCInspection(r.Context(), qcapp.CreateInboundQCInspectionInput{
				ID:                 payload.ID,
				OrgID:              payload.OrgID,
				GoodsReceiptID:     payload.GoodsReceiptID,
				GoodsReceiptLineID: payload.GoodsReceiptLineID,
				InspectorID:        payload.InspectorID,
				Checklist:          inboundQCChecklistInputs(payload.Checklist),
				Note:               payload.Note,
				ActorID:            principal.UserID,
				RequestID:          response.RequestID(r),
			})
			if err != nil {
				writeInboundQCInspectionError(w, r, err)
				return
			}
			response.WriteSuccess(w, r, http.StatusCreated, newInboundQCActionResultResponse(result))
		default:
			response.WriteError(w, r, http.StatusMethodNotAllowed, response.ErrorCodeNotFound, "Route not found", nil)
		}
	}
}

func inboundQCInspectionDetailHandler(service qcapp.InboundQCInspectionService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			response.WriteError(w, r, http.StatusMethodNotAllowed, response.ErrorCodeNotFound, "Route not found", nil)
			return
		}
		principal, ok := auth.PrincipalFromContext(r.Context())
		if !ok {
			response.WriteError(w, r, http.StatusUnauthorized, response.ErrorCodeUnauthorized, "Authentication required", nil)
			return
		}
		if !auth.HasPermission(principal, auth.PermissionQCView) {
			writePermissionDenied(w, r, auth.PermissionQCView)
			return
		}

		inspection, err := service.GetInboundQCInspection(r.Context(), r.PathValue("inspection_id"))
		if err != nil {
			writeInboundQCInspectionError(w, r, err)
			return
		}
		response.WriteSuccess(w, r, http.StatusOK, newInboundQCInspectionResponse(inspection, ""))
	}
}

func inboundQCInspectionStartHandler(service qcapp.InboundQCInspectionService) http.HandlerFunc {
	return inboundQCInspectionActionHandler(service, "start")
}

func inboundQCInspectionPassHandler(service qcapp.InboundQCInspectionService) http.HandlerFunc {
	return inboundQCInspectionActionHandler(service, "pass")
}

func inboundQCInspectionFailHandler(service qcapp.InboundQCInspectionService) http.HandlerFunc {
	return inboundQCInspectionActionHandler(service, "fail")
}

func inboundQCInspectionHoldHandler(service qcapp.InboundQCInspectionService) http.HandlerFunc {
	return inboundQCInspectionActionHandler(service, "hold")
}

func inboundQCInspectionPartialHandler(service qcapp.InboundQCInspectionService) http.HandlerFunc {
	return inboundQCInspectionActionHandler(service, "partial")
}

func inboundQCInspectionActionHandler(service qcapp.InboundQCInspectionService, action string) http.HandlerFunc {
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
		if !auth.HasPermission(principal, auth.PermissionQCDecision) {
			writePermissionDenied(w, r, auth.PermissionQCDecision)
			return
		}

		r = requestWithStableID(r)
		input := qcapp.InboundQCActionInput{
			ID:        r.PathValue("inspection_id"),
			ActorID:   principal.UserID,
			RequestID: response.RequestID(r),
		}
		if action != "start" {
			var payload inboundQCDecisionRequest
			if r.Body != nil {
				if err := json.NewDecoder(r.Body).Decode(&payload); err != nil && !errors.Is(err, io.EOF) {
					response.WriteError(w, r, http.StatusBadRequest, response.ErrorCodeValidation, "Invalid inbound QC decision payload", nil)
					return
				}
			}
			input.PassedQuantity = payload.PassedQuantity
			input.FailedQuantity = payload.FailedQuantity
			input.HoldQuantity = payload.HoldQuantity
			input.Checklist = inboundQCChecklistInputs(payload.Checklist)
			input.Reason = payload.Reason
			input.Note = payload.Note
		}

		result, err := applyInboundQCAction(r.Context(), service, action, input)
		if err != nil {
			writeInboundQCInspectionError(w, r, err)
			return
		}
		response.WriteSuccess(w, r, http.StatusOK, newInboundQCActionResultResponse(result))
	}
}

func applyInboundQCAction(
	ctx context.Context,
	service qcapp.InboundQCInspectionService,
	action string,
	input qcapp.InboundQCActionInput,
) (qcapp.InboundQCInspectionResult, error) {
	switch action {
	case "start":
		return service.StartInboundQCInspection(ctx, input)
	case "pass":
		return service.PassInboundQCInspection(ctx, input)
	case "fail":
		return service.FailInboundQCInspection(ctx, input)
	case "hold":
		return service.HoldInboundQCInspection(ctx, input)
	case "partial":
		return service.PartialInboundQCInspection(ctx, input)
	default:
		return qcapp.InboundQCInspectionResult{}, qcapp.ErrInboundQCInspectionNotFound
	}
}

func inboundQCInspectionFilterFromRequest(r *http.Request) qcapp.InboundQCInspectionFilter {
	query := r.URL.Query()
	return qcapp.NewInboundQCInspectionFilter(
		qcdomain.InboundQCInspectionStatus(query.Get("status")),
		query.Get("goods_receipt_id"),
		query.Get("goods_receipt_line_id"),
		query.Get("warehouse_id"),
	)
}

func inboundQCChecklistInputs(inputs []inboundQCChecklistItemRequest) []qcapp.InboundQCChecklistInput {
	if inputs == nil {
		return nil
	}
	items := make([]qcapp.InboundQCChecklistInput, 0, len(inputs))
	for _, input := range inputs {
		items = append(items, qcapp.InboundQCChecklistInput{
			ID:       input.ID,
			Code:     input.Code,
			Label:    input.Label,
			Required: input.Required,
			Status:   input.Status,
			Note:     input.Note,
		})
	}

	return items
}

func newInboundQCActionResultResponse(result qcapp.InboundQCInspectionResult) inboundQCActionResultResponse {
	return inboundQCActionResultResponse{
		Inspection:     newInboundQCInspectionResponse(result.Inspection, result.AuditLogID),
		PreviousStatus: string(result.PreviousStatus),
		CurrentStatus:  string(result.CurrentStatus),
		PreviousResult: string(result.PreviousResult),
		CurrentResult:  string(result.CurrentResult),
		AuditLogID:     result.AuditLogID,
	}
}

func newInboundQCInspectionResponse(
	inspection qcdomain.InboundQCInspection,
	auditLogID string,
) inboundQCInspectionResponse {
	payload := inboundQCInspectionResponse{
		ID:                  inspection.ID,
		OrgID:               inspection.OrgID,
		GoodsReceiptID:      inspection.GoodsReceiptID,
		GoodsReceiptNo:      inspection.GoodsReceiptNo,
		GoodsReceiptLineID:  inspection.GoodsReceiptLineID,
		PurchaseOrderID:     inspection.PurchaseOrderID,
		PurchaseOrderLineID: inspection.PurchaseOrderLineID,
		ItemID:              inspection.ItemID,
		SKU:                 inspection.SKU,
		ItemName:            inspection.ItemName,
		BatchID:             inspection.BatchID,
		BatchNo:             inspection.BatchNo,
		LotNo:               inspection.LotNo,
		ExpiryDate:          dateString(inspection.ExpiryDate),
		WarehouseID:         inspection.WarehouseID,
		LocationID:          inspection.LocationID,
		Quantity:            inspection.Quantity.String(),
		UOMCode:             inspection.UOMCode.String(),
		InspectorID:         inspection.InspectorID,
		Status:              string(inspection.Status),
		Result:              string(inspection.Result),
		PassedQuantity:      inboundQCResponseQuantityString(inspection.PassedQuantity),
		FailedQuantity:      inboundQCResponseQuantityString(inspection.FailedQuantity),
		HoldQuantity:        inboundQCResponseQuantityString(inspection.HoldQuantity),
		Checklist:           make([]inboundQCChecklistItemResponse, 0, len(inspection.Checklist)),
		Reason:              inspection.Reason,
		Note:                inspection.Note,
		AuditLogID:          auditLogID,
		CreatedAt:           timeString(inspection.CreatedAt),
		CreatedBy:           inspection.CreatedBy,
		UpdatedAt:           timeString(inspection.UpdatedAt),
		UpdatedBy:           inspection.UpdatedBy,
		StartedAt:           timeString(inspection.StartedAt),
		StartedBy:           inspection.StartedBy,
		DecidedAt:           timeString(inspection.DecidedAt),
		DecidedBy:           inspection.DecidedBy,
	}
	for _, item := range inspection.Checklist {
		payload.Checklist = append(payload.Checklist, inboundQCChecklistItemResponse{
			ID:       item.ID,
			Code:     item.Code,
			Label:    item.Label,
			Required: item.Required,
			Status:   string(item.Status),
			Note:     item.Note,
		})
	}

	return payload
}

func inboundQCResponseQuantityString(value decimal.Decimal) string {
	if value.String() == "" {
		return decimal.MustQuantity("0").String()
	}

	return value.String()
}

func writeInboundQCInspectionError(w http.ResponseWriter, r *http.Request, err error) {
	switch {
	case errors.Is(err, qcapp.ErrInboundQCInspectionNotFound):
		response.WriteError(w, r, http.StatusNotFound, response.ErrorCodeNotFound, "Inbound QC inspection not found", nil)
	case errors.Is(err, qcapp.ErrInboundQCReceivingNotFound):
		response.WriteError(w, r, http.StatusNotFound, response.ErrorCodeNotFound, "Goods receipt not found", nil)
	case errors.Is(err, qcapp.ErrInboundQCReceivingLineNotFound):
		response.WriteError(w, r, http.StatusNotFound, response.ErrorCodeNotFound, "Goods receipt line not found", nil)
	case errors.Is(err, qcapp.ErrInboundQCReceivingInvalidState):
		response.WriteError(w, r, http.StatusConflict, response.ErrorCodeConflict, "Goods receipt must be inspection-ready before inbound QC", nil)
	case errors.Is(err, qcapp.ErrInboundQCDuplicateReceivingLine):
		response.WriteError(w, r, http.StatusConflict, response.ErrorCodeConflict, "Inbound QC inspection already exists for this goods receipt line", nil)
	case errors.Is(err, qcdomain.ErrInboundQCInspectionInvalidTransition):
		response.WriteError(w, r, http.StatusConflict, response.ErrorCodeConflict, "Inbound QC inspection status transition is not allowed", nil)
	case errors.Is(err, qcdomain.ErrInboundQCInspectionRequiredField),
		errors.Is(err, qcdomain.ErrInboundQCInspectionInvalidStatus),
		errors.Is(err, qcdomain.ErrInboundQCInspectionInvalidResult),
		errors.Is(err, qcdomain.ErrInboundQCInspectionInvalidQuantity),
		errors.Is(err, qcdomain.ErrInboundQCChecklistInvalidStatus),
		errors.Is(err, qcdomain.ErrInboundQCChecklistIncomplete),
		errors.Is(err, decimal.ErrInvalidDecimal),
		errors.Is(err, decimal.ErrDecimalOutOfRange),
		errors.Is(err, decimal.ErrInvalidUOMCode):
		response.WriteError(
			w,
			r,
			http.StatusBadRequest,
			response.ErrorCodeValidation,
			"Invalid inbound QC inspection payload",
			map[string]any{"required": "goods_receipt_id, goods_receipt_line_id, inspector, checklist, valid quantities, and reason for fail/hold/partial"},
		)
	default:
		response.WriteError(w, r, http.StatusConflict, response.ErrorCodeConflict, "Inbound QC inspection request could not be processed", nil)
	}
}
