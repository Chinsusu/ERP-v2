package main

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"

	masterdataapp "github.com/Chinsusu/ERP-v2/apps/api/internal/modules/masterdata/application"
	purchaseapp "github.com/Chinsusu/ERP-v2/apps/api/internal/modules/purchase/application"
	purchasedomain "github.com/Chinsusu/ERP-v2/apps/api/internal/modules/purchase/domain"
	"github.com/Chinsusu/ERP-v2/apps/api/internal/shared/auth"
	apperrors "github.com/Chinsusu/ERP-v2/apps/api/internal/shared/errors"
	"github.com/Chinsusu/ERP-v2/apps/api/internal/shared/response"
)

type purchaseOrderUOMConverterAdapter struct {
	catalog uomCatalog
}

type purchaseOrderLineRequest struct {
	ID           string `json:"id"`
	LineNo       int    `json:"line_no"`
	ItemID       string `json:"item_id"`
	OrderedQty   string `json:"ordered_qty"`
	UOMCode      string `json:"uom_code"`
	UnitPrice    string `json:"unit_price"`
	CurrencyCode string `json:"currency_code"`
	ExpectedDate string `json:"expected_date"`
	Note         string `json:"note"`
}

type createPurchaseOrderRequest struct {
	ID           string                     `json:"id"`
	PONo         string                     `json:"po_no"`
	SupplierID   string                     `json:"supplier_id"`
	WarehouseID  string                     `json:"warehouse_id"`
	ExpectedDate string                     `json:"expected_date"`
	CurrencyCode string                     `json:"currency_code"`
	Note         string                     `json:"note"`
	Lines        []purchaseOrderLineRequest `json:"lines"`
}

type updatePurchaseOrderRequest struct {
	SupplierID      string                     `json:"supplier_id"`
	WarehouseID     string                     `json:"warehouse_id"`
	ExpectedDate    string                     `json:"expected_date"`
	Note            string                     `json:"note"`
	ExpectedVersion int                        `json:"expected_version"`
	Lines           []purchaseOrderLineRequest `json:"lines"`
}

type purchaseOrderActionRequest struct {
	ExpectedVersion int    `json:"expected_version"`
	Reason          string `json:"reason"`
	Note            string `json:"note"`
}

type purchaseOrderLineResponse struct {
	ID               string `json:"id"`
	LineNo           int    `json:"line_no"`
	ItemID           string `json:"item_id"`
	SKUCode          string `json:"sku_code"`
	ItemName         string `json:"item_name"`
	OrderedQty       string `json:"ordered_qty"`
	ReceivedQty      string `json:"received_qty"`
	UOMCode          string `json:"uom_code"`
	BaseOrderedQty   string `json:"base_ordered_qty"`
	BaseReceivedQty  string `json:"base_received_qty"`
	BaseUOMCode      string `json:"base_uom_code"`
	ConversionFactor string `json:"conversion_factor"`
	UnitPrice        string `json:"unit_price"`
	CurrencyCode     string `json:"currency_code"`
	LineAmount       string `json:"line_amount"`
	ExpectedDate     string `json:"expected_date"`
	Note             string `json:"note,omitempty"`
}

type purchaseOrderListItemResponse struct {
	ID            string `json:"id"`
	PONo          string `json:"po_no"`
	SupplierID    string `json:"supplier_id"`
	SupplierCode  string `json:"supplier_code,omitempty"`
	SupplierName  string `json:"supplier_name"`
	WarehouseID   string `json:"warehouse_id"`
	WarehouseCode string `json:"warehouse_code,omitempty"`
	ExpectedDate  string `json:"expected_date"`
	Status        string `json:"status"`
	CurrencyCode  string `json:"currency_code"`
	TotalAmount   string `json:"total_amount"`
	Note          string `json:"note,omitempty"`
	LineCount     int    `json:"line_count"`
	ReceivedLines int    `json:"received_line_count"`
	CreatedAt     string `json:"created_at"`
	UpdatedAt     string `json:"updated_at"`
	Version       int    `json:"version"`
}

type purchaseOrderResponse struct {
	ID             string                      `json:"id"`
	PONo           string                      `json:"po_no"`
	SupplierID     string                      `json:"supplier_id"`
	SupplierCode   string                      `json:"supplier_code,omitempty"`
	SupplierName   string                      `json:"supplier_name"`
	WarehouseID    string                      `json:"warehouse_id"`
	WarehouseCode  string                      `json:"warehouse_code,omitempty"`
	ExpectedDate   string                      `json:"expected_date"`
	Status         string                      `json:"status"`
	CurrencyCode   string                      `json:"currency_code"`
	SubtotalAmount string                      `json:"subtotal_amount"`
	TotalAmount    string                      `json:"total_amount"`
	Note           string                      `json:"note,omitempty"`
	Lines          []purchaseOrderLineResponse `json:"lines"`
	AuditLogID     string                      `json:"audit_log_id,omitempty"`
	CreatedAt      string                      `json:"created_at"`
	UpdatedAt      string                      `json:"updated_at"`
	SubmittedAt    string                      `json:"submitted_at,omitempty"`
	ApprovedAt     string                      `json:"approved_at,omitempty"`
	ClosedAt       string                      `json:"closed_at,omitempty"`
	CancelledAt    string                      `json:"cancelled_at,omitempty"`
	RejectedAt     string                      `json:"rejected_at,omitempty"`
	CancelReason   string                      `json:"cancel_reason,omitempty"`
	RejectReason   string                      `json:"reject_reason,omitempty"`
	Version        int                         `json:"version"`
}

type purchaseOrderActionResultResponse struct {
	PurchaseOrder  purchaseOrderResponse `json:"purchase_order"`
	PreviousStatus string                `json:"previous_status"`
	CurrentStatus  string                `json:"current_status"`
	AuditLogID     string                `json:"audit_log_id,omitempty"`
}

func (a purchaseOrderUOMConverterAdapter) ConvertToBase(
	ctx context.Context,
	input purchaseapp.ConvertPurchaseOrderLineToBaseInput,
) (purchaseapp.ConvertPurchaseOrderLineToBaseResult, error) {
	result, err := a.catalog.ConvertToBase(ctx, masterdataapp.ConvertToBaseInput{
		ItemID:      input.ItemID,
		SKU:         input.SKU,
		Quantity:    input.Quantity,
		FromUOMCode: input.FromUOMCode,
		BaseUOMCode: input.BaseUOMCode,
	})
	if err != nil {
		return purchaseapp.ConvertPurchaseOrderLineToBaseResult{}, err
	}

	return purchaseapp.ConvertPurchaseOrderLineToBaseResult{
		Quantity:         result.Quantity,
		SourceUOMCode:    result.SourceUOMCode,
		BaseQuantity:     result.BaseQuantity,
		BaseUOMCode:      result.BaseUOMCode,
		ConversionFactor: result.ConversionFactor,
	}, nil
}

func purchaseOrdersHandler(service purchaseapp.PurchaseOrderService) http.HandlerFunc {
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

		switch r.Method {
		case http.MethodGet:
			orders, err := service.ListPurchaseOrders(r.Context(), purchaseOrderFilterFromRequest(r))
			if err != nil {
				writePurchaseOrderError(w, r, err)
				return
			}

			payload := make([]purchaseOrderListItemResponse, 0, len(orders))
			for _, order := range orders {
				payload = append(payload, newPurchaseOrderListItemResponse(order))
			}
			response.WriteSuccess(w, r, http.StatusOK, payload)
		case http.MethodPost:
			if !auth.HasPermission(principal, auth.PermissionRecordCreate) {
				writePermissionDenied(w, r, auth.PermissionRecordCreate)
				return
			}
			r = requestWithStableID(r)
			var payload createPurchaseOrderRequest
			if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
				response.WriteError(w, r, http.StatusBadRequest, response.ErrorCodeValidation, "Invalid purchase order payload", nil)
				return
			}

			result, err := service.CreatePurchaseOrder(r.Context(), purchaseapp.CreatePurchaseOrderInput{
				ID:           payload.ID,
				PONo:         payload.PONo,
				SupplierID:   payload.SupplierID,
				WarehouseID:  payload.WarehouseID,
				ExpectedDate: payload.ExpectedDate,
				CurrencyCode: payload.CurrencyCode,
				Note:         payload.Note,
				Lines:        purchaseOrderLineInputs(payload.Lines),
				ActorID:      principal.UserID,
				RequestID:    response.RequestID(r),
			})
			if err != nil {
				writePurchaseOrderError(w, r, err)
				return
			}

			response.WriteSuccess(w, r, http.StatusCreated, newPurchaseOrderResponse(result.PurchaseOrder, result.AuditLogID))
		default:
			response.WriteError(w, r, http.StatusMethodNotAllowed, response.ErrorCodeNotFound, "Route not found", nil)
		}
	}
}

func purchaseOrderDetailHandler(service purchaseapp.PurchaseOrderService) http.HandlerFunc {
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

		switch r.Method {
		case http.MethodGet:
			order, err := service.GetPurchaseOrder(r.Context(), r.PathValue("purchase_order_id"))
			if err != nil {
				writePurchaseOrderError(w, r, err)
				return
			}
			response.WriteSuccess(w, r, http.StatusOK, newPurchaseOrderResponse(order, ""))
		case http.MethodPatch:
			if !auth.HasPermission(principal, auth.PermissionRecordCreate) {
				writePermissionDenied(w, r, auth.PermissionRecordCreate)
				return
			}
			r = requestWithStableID(r)
			var payload updatePurchaseOrderRequest
			if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
				response.WriteError(w, r, http.StatusBadRequest, response.ErrorCodeValidation, "Invalid purchase order payload", nil)
				return
			}

			result, err := service.UpdatePurchaseOrder(r.Context(), purchaseapp.UpdatePurchaseOrderInput{
				ID:              r.PathValue("purchase_order_id"),
				SupplierID:      payload.SupplierID,
				WarehouseID:     payload.WarehouseID,
				ExpectedDate:    payload.ExpectedDate,
				Note:            payload.Note,
				Lines:           purchaseOrderLineInputs(payload.Lines),
				ExpectedVersion: payload.ExpectedVersion,
				ActorID:         principal.UserID,
				RequestID:       response.RequestID(r),
			})
			if err != nil {
				writePurchaseOrderError(w, r, err)
				return
			}
			response.WriteSuccess(w, r, http.StatusOK, newPurchaseOrderResponse(result.PurchaseOrder, result.AuditLogID))
		default:
			response.WriteError(w, r, http.StatusMethodNotAllowed, response.ErrorCodeNotFound, "Route not found", nil)
		}
	}
}

func purchaseOrderSubmitHandler(service purchaseapp.PurchaseOrderService) http.HandlerFunc {
	return purchaseOrderActionHandler(service, "submit")
}

func purchaseOrderApproveHandler(service purchaseapp.PurchaseOrderService) http.HandlerFunc {
	return purchaseOrderActionHandler(service, "approve")
}

func purchaseOrderCancelHandler(service purchaseapp.PurchaseOrderService) http.HandlerFunc {
	return purchaseOrderActionHandler(service, "cancel")
}

func purchaseOrderCloseHandler(service purchaseapp.PurchaseOrderService) http.HandlerFunc {
	return purchaseOrderActionHandler(service, "close")
}

func purchaseOrderActionHandler(service purchaseapp.PurchaseOrderService, action string) http.HandlerFunc {
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
		if !auth.HasPermission(principal, auth.PermissionPurchaseView) {
			writePermissionDenied(w, r, auth.PermissionPurchaseView)
			return
		}
		if !auth.HasPermission(principal, auth.PermissionRecordCreate) {
			writePermissionDenied(w, r, auth.PermissionRecordCreate)
			return
		}

		r = requestWithStableID(r)
		var payload purchaseOrderActionRequest
		if r.Body != nil {
			if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
				if !errors.Is(err, io.EOF) {
					response.WriteError(w, r, http.StatusBadRequest, response.ErrorCodeValidation, "Invalid purchase order action payload", nil)
					return
				}
			}
		}

		input := purchaseapp.PurchaseOrderActionInput{
			ID:              r.PathValue("purchase_order_id"),
			ExpectedVersion: payload.ExpectedVersion,
			Reason:          payload.Reason,
			Note:            payload.Note,
			ActorID:         principal.UserID,
			RequestID:       response.RequestID(r),
		}
		var (
			result purchaseapp.PurchaseOrderActionResult
			err    error
		)
		switch action {
		case "submit":
			result, err = service.SubmitPurchaseOrder(r.Context(), input)
		case "approve":
			result, err = service.ApprovePurchaseOrder(r.Context(), input)
		case "cancel":
			result, err = service.CancelPurchaseOrder(r.Context(), input)
		case "close":
			result, err = service.ClosePurchaseOrder(r.Context(), input)
		default:
			response.WriteError(w, r, http.StatusNotFound, response.ErrorCodeNotFound, "Route not found", nil)
			return
		}
		if err != nil {
			writePurchaseOrderError(w, r, err)
			return
		}

		response.WriteSuccess(w, r, http.StatusOK, purchaseOrderActionResultResponse{
			PurchaseOrder:  newPurchaseOrderResponse(result.PurchaseOrder, ""),
			PreviousStatus: string(result.PreviousStatus),
			CurrentStatus:  string(result.CurrentStatus),
			AuditLogID:     result.AuditLogID,
		})
	}
}

func purchaseOrderFilterFromRequest(r *http.Request) purchaseapp.PurchaseOrderFilter {
	query := r.URL.Query()
	statuses := []purchasedomain.PurchaseOrderStatus{}
	for _, rawStatus := range strings.Split(query.Get("status"), ",") {
		status := purchasedomain.NormalizePurchaseOrderStatus(purchasedomain.PurchaseOrderStatus(rawStatus))
		if status != "" {
			statuses = append(statuses, status)
		}
	}

	return purchaseapp.PurchaseOrderFilter{
		Search:       query.Get("search"),
		Statuses:     statuses,
		SupplierID:   query.Get("supplier_id"),
		WarehouseID:  query.Get("warehouse_id"),
		ExpectedFrom: query.Get("expected_from"),
		ExpectedTo:   query.Get("expected_to"),
	}
}

func purchaseOrderLineInputs(inputs []purchaseOrderLineRequest) []purchaseapp.PurchaseOrderLineInput {
	if inputs == nil {
		return nil
	}
	lines := make([]purchaseapp.PurchaseOrderLineInput, 0, len(inputs))
	for _, input := range inputs {
		lines = append(lines, purchaseapp.PurchaseOrderLineInput{
			ID:           input.ID,
			LineNo:       input.LineNo,
			ItemID:       input.ItemID,
			OrderedQty:   input.OrderedQty,
			UOMCode:      input.UOMCode,
			UnitPrice:    input.UnitPrice,
			CurrencyCode: input.CurrencyCode,
			ExpectedDate: input.ExpectedDate,
			Note:         input.Note,
		})
	}

	return lines
}

func newPurchaseOrderListItemResponse(order purchasedomain.PurchaseOrder) purchaseOrderListItemResponse {
	receivedLines := 0
	for _, line := range order.Lines {
		if !line.ReceivedQty.IsZero() {
			receivedLines++
		}
	}

	return purchaseOrderListItemResponse{
		ID:            order.ID,
		PONo:          order.PONo,
		SupplierID:    order.SupplierID,
		SupplierCode:  order.SupplierCode,
		SupplierName:  order.SupplierName,
		WarehouseID:   order.WarehouseID,
		WarehouseCode: order.WarehouseCode,
		ExpectedDate:  order.ExpectedDate,
		Status:        string(order.Status),
		CurrencyCode:  order.CurrencyCode.String(),
		TotalAmount:   order.TotalAmount.String(),
		Note:          order.Note,
		LineCount:     len(order.Lines),
		ReceivedLines: receivedLines,
		CreatedAt:     timeString(order.CreatedAt),
		UpdatedAt:     timeString(order.UpdatedAt),
		Version:       order.Version,
	}
}

func newPurchaseOrderResponse(order purchasedomain.PurchaseOrder, auditLogID string) purchaseOrderResponse {
	payload := purchaseOrderResponse{
		ID:             order.ID,
		PONo:           order.PONo,
		SupplierID:     order.SupplierID,
		SupplierCode:   order.SupplierCode,
		SupplierName:   order.SupplierName,
		WarehouseID:    order.WarehouseID,
		WarehouseCode:  order.WarehouseCode,
		ExpectedDate:   order.ExpectedDate,
		Status:         string(order.Status),
		CurrencyCode:   order.CurrencyCode.String(),
		SubtotalAmount: order.SubtotalAmount.String(),
		TotalAmount:    order.TotalAmount.String(),
		Note:           order.Note,
		Lines:          make([]purchaseOrderLineResponse, 0, len(order.Lines)),
		AuditLogID:     auditLogID,
		CreatedAt:      timeString(order.CreatedAt),
		UpdatedAt:      timeString(order.UpdatedAt),
		SubmittedAt:    timeString(order.SubmittedAt),
		ApprovedAt:     timeString(order.ApprovedAt),
		ClosedAt:       timeString(order.ClosedAt),
		CancelledAt:    timeString(order.CancelledAt),
		RejectedAt:     timeString(order.RejectedAt),
		CancelReason:   order.CancelReason,
		RejectReason:   order.RejectReason,
		Version:        order.Version,
	}
	for _, line := range order.Lines {
		payload.Lines = append(payload.Lines, purchaseOrderLineResponse{
			ID:               line.ID,
			LineNo:           line.LineNo,
			ItemID:           line.ItemID,
			SKUCode:          line.SKUCode,
			ItemName:         line.ItemName,
			OrderedQty:       line.OrderedQty.String(),
			ReceivedQty:      line.ReceivedQty.String(),
			UOMCode:          line.UOMCode.String(),
			BaseOrderedQty:   line.BaseOrderedQty.String(),
			BaseReceivedQty:  line.BaseReceivedQty.String(),
			BaseUOMCode:      line.BaseUOMCode.String(),
			ConversionFactor: line.ConversionFactor.String(),
			UnitPrice:        line.UnitPrice.String(),
			CurrencyCode:     line.CurrencyCode.String(),
			LineAmount:       line.LineAmount.String(),
			ExpectedDate:     line.ExpectedDate,
			Note:             line.Note,
		})
	}

	return payload
}

func writePurchaseOrderError(w http.ResponseWriter, r *http.Request, err error) {
	if appErr, ok := apperrors.As(err); ok {
		response.WriteError(w, r, appErr.HTTPStatus, appErr.Code, appErr.Message, appErr.Details)
		return
	}

	response.WriteError(w, r, http.StatusConflict, response.ErrorCodeConflict, "Purchase order request could not be processed", nil)
}
