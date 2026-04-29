package main

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"

	inventoryapp "github.com/Chinsusu/ERP-v2/apps/api/internal/modules/inventory/application"
	inventorydomain "github.com/Chinsusu/ERP-v2/apps/api/internal/modules/inventory/domain"
	"github.com/Chinsusu/ERP-v2/apps/api/internal/shared/auth"
	"github.com/Chinsusu/ERP-v2/apps/api/internal/shared/response"
)

type supplierRejectionRequest struct {
	ID                    string                               `json:"id"`
	OrgID                 string                               `json:"org_id"`
	RejectionNo           string                               `json:"rejection_no"`
	SupplierID            string                               `json:"supplier_id"`
	SupplierCode          string                               `json:"supplier_code"`
	SupplierName          string                               `json:"supplier_name"`
	PurchaseOrderID       string                               `json:"purchase_order_id"`
	PurchaseOrderNo       string                               `json:"purchase_order_no"`
	GoodsReceiptID        string                               `json:"goods_receipt_id"`
	GoodsReceiptNo        string                               `json:"goods_receipt_no"`
	InboundQCInspectionID string                               `json:"inbound_qc_inspection_id"`
	WarehouseID           string                               `json:"warehouse_id"`
	WarehouseCode         string                               `json:"warehouse_code"`
	Reason                string                               `json:"reason"`
	Lines                 []supplierRejectionLineRequest       `json:"lines"`
	Attachments           []supplierRejectionAttachmentRequest `json:"attachments"`
}

type supplierRejectionLineRequest struct {
	ID                    string `json:"id"`
	PurchaseOrderLineID   string `json:"purchase_order_line_id"`
	GoodsReceiptLineID    string `json:"goods_receipt_line_id"`
	InboundQCInspectionID string `json:"inbound_qc_inspection_id"`
	ItemID                string `json:"item_id"`
	SKU                   string `json:"sku"`
	ItemName              string `json:"item_name"`
	BatchID               string `json:"batch_id"`
	BatchNo               string `json:"batch_no"`
	LotNo                 string `json:"lot_no"`
	ExpiryDate            string `json:"expiry_date"`
	RejectedQuantity      string `json:"rejected_qty"`
	UOMCode               string `json:"uom_code"`
	BaseUOMCode           string `json:"base_uom_code"`
	Reason                string `json:"reason"`
}

type supplierRejectionAttachmentRequest struct {
	ID          string `json:"id"`
	LineID      string `json:"line_id"`
	FileName    string `json:"file_name"`
	ObjectKey   string `json:"object_key"`
	ContentType string `json:"content_type"`
	Source      string `json:"source"`
}

type supplierRejectionResponse struct {
	ID                    string                                `json:"id"`
	OrgID                 string                                `json:"org_id"`
	RejectionNo           string                                `json:"rejection_no"`
	SupplierID            string                                `json:"supplier_id"`
	SupplierCode          string                                `json:"supplier_code,omitempty"`
	SupplierName          string                                `json:"supplier_name,omitempty"`
	PurchaseOrderID       string                                `json:"purchase_order_id,omitempty"`
	PurchaseOrderNo       string                                `json:"purchase_order_no,omitempty"`
	GoodsReceiptID        string                                `json:"goods_receipt_id"`
	GoodsReceiptNo        string                                `json:"goods_receipt_no,omitempty"`
	InboundQCInspectionID string                                `json:"inbound_qc_inspection_id"`
	WarehouseID           string                                `json:"warehouse_id"`
	WarehouseCode         string                                `json:"warehouse_code,omitempty"`
	Status                string                                `json:"status"`
	Reason                string                                `json:"reason"`
	Lines                 []supplierRejectionLineResponse       `json:"lines"`
	Attachments           []supplierRejectionAttachmentResponse `json:"attachments"`
	AuditLogID            string                                `json:"audit_log_id,omitempty"`
	CreatedAt             string                                `json:"created_at"`
	CreatedBy             string                                `json:"created_by"`
	UpdatedAt             string                                `json:"updated_at"`
	UpdatedBy             string                                `json:"updated_by"`
	SubmittedAt           string                                `json:"submitted_at,omitempty"`
	SubmittedBy           string                                `json:"submitted_by,omitempty"`
	ConfirmedAt           string                                `json:"confirmed_at,omitempty"`
	ConfirmedBy           string                                `json:"confirmed_by,omitempty"`
}

type supplierRejectionLineResponse struct {
	ID                    string `json:"id"`
	PurchaseOrderLineID   string `json:"purchase_order_line_id,omitempty"`
	GoodsReceiptLineID    string `json:"goods_receipt_line_id"`
	InboundQCInspectionID string `json:"inbound_qc_inspection_id"`
	ItemID                string `json:"item_id"`
	SKU                   string `json:"sku"`
	ItemName              string `json:"item_name,omitempty"`
	BatchID               string `json:"batch_id"`
	BatchNo               string `json:"batch_no"`
	LotNo                 string `json:"lot_no"`
	ExpiryDate            string `json:"expiry_date"`
	RejectedQuantity      string `json:"rejected_qty"`
	UOMCode               string `json:"uom_code"`
	BaseUOMCode           string `json:"base_uom_code"`
	Reason                string `json:"reason"`
}

type supplierRejectionAttachmentResponse struct {
	ID          string `json:"id"`
	LineID      string `json:"line_id,omitempty"`
	FileName    string `json:"file_name"`
	ObjectKey   string `json:"object_key"`
	ContentType string `json:"content_type,omitempty"`
	UploadedAt  string `json:"uploaded_at"`
	UploadedBy  string `json:"uploaded_by"`
	Source      string `json:"source,omitempty"`
}

type supplierRejectionActionResultResponse struct {
	Rejection      supplierRejectionResponse `json:"rejection"`
	PreviousStatus string                    `json:"previous_status,omitempty"`
	CurrentStatus  string                    `json:"current_status"`
	AuditLogID     string                    `json:"audit_log_id,omitempty"`
}

func supplierRejectionsHandler(
	list inventoryapp.ListSupplierRejections,
	create inventoryapp.CreateSupplierRejection,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		principal, ok := auth.PrincipalFromContext(r.Context())
		if !ok {
			response.WriteError(w, r, http.StatusUnauthorized, response.ErrorCodeUnauthorized, "Authentication required", nil)
			return
		}

		switch r.Method {
		case http.MethodGet:
			if !auth.HasPermission(principal, auth.PermissionWarehouseView) {
				writePermissionDenied(w, r, auth.PermissionWarehouseView)
				return
			}
			rows, err := list.Execute(
				r.Context(),
				inventorydomain.NewSupplierRejectionFilter(
					r.URL.Query().Get("supplier_id"),
					r.URL.Query().Get("warehouse_id"),
					inventorydomain.SupplierRejectionStatus(r.URL.Query().Get("status")),
				),
			)
			if err != nil {
				writeSupplierRejectionError(w, r, err)
				return
			}
			payload := make([]supplierRejectionResponse, 0, len(rows))
			for _, row := range rows {
				payload = append(payload, newSupplierRejectionResponse(row, ""))
			}
			response.WriteSuccess(w, r, http.StatusOK, payload)
		case http.MethodPost:
			if !auth.HasPermission(principal, auth.PermissionRecordCreate) {
				writePermissionDenied(w, r, auth.PermissionRecordCreate)
				return
			}
			var payload supplierRejectionRequest
			if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
				response.WriteError(w, r, http.StatusBadRequest, response.ErrorCodeValidation, "Invalid supplier rejection payload", nil)
				return
			}
			result, err := create.Execute(r.Context(), inventoryapp.CreateSupplierRejectionInput{
				ID:                    payload.ID,
				OrgID:                 payload.OrgID,
				RejectionNo:           payload.RejectionNo,
				SupplierID:            payload.SupplierID,
				SupplierCode:          payload.SupplierCode,
				SupplierName:          payload.SupplierName,
				PurchaseOrderID:       payload.PurchaseOrderID,
				PurchaseOrderNo:       payload.PurchaseOrderNo,
				GoodsReceiptID:        payload.GoodsReceiptID,
				GoodsReceiptNo:        payload.GoodsReceiptNo,
				InboundQCInspectionID: payload.InboundQCInspectionID,
				WarehouseID:           payload.WarehouseID,
				WarehouseCode:         payload.WarehouseCode,
				Reason:                payload.Reason,
				Lines:                 newCreateSupplierRejectionLineInputs(payload.Lines),
				Attachments:           newCreateSupplierRejectionAttachmentInputs(payload.Attachments),
				ActorID:               principal.UserID,
				RequestID:             response.RequestID(r),
			})
			if err != nil {
				writeSupplierRejectionError(w, r, err)
				return
			}
			response.WriteSuccess(w, r, http.StatusCreated, newSupplierRejectionResponse(result.Rejection, result.AuditLogID))
		default:
			response.WriteError(w, r, http.StatusMethodNotAllowed, response.ErrorCodeNotFound, "Route not found", nil)
		}
	}
}

func supplierRejectionDetailHandler(store inventoryapp.SupplierRejectionStore) http.HandlerFunc {
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
		if !auth.HasPermission(principal, auth.PermissionWarehouseView) {
			writePermissionDenied(w, r, auth.PermissionWarehouseView)
			return
		}

		rejection, err := store.Get(r.Context(), r.PathValue("supplier_rejection_id"))
		if err != nil {
			writeSupplierRejectionError(w, r, err)
			return
		}
		response.WriteSuccess(w, r, http.StatusOK, newSupplierRejectionResponse(rejection, ""))
	}
}

func supplierRejectionActionHandler(
	transition inventoryapp.TransitionSupplierRejection,
	action string,
) http.HandlerFunc {
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
		if !auth.HasPermission(principal, auth.PermissionRecordCreate) {
			writePermissionDenied(w, r, auth.PermissionRecordCreate)
			return
		}

		id := r.PathValue("supplier_rejection_id")
		var result inventoryapp.SupplierRejectionTransitionResult
		var err error
		switch action {
		case "submit":
			result, err = transition.Submit(r.Context(), id, principal.UserID, response.RequestID(r))
		case "confirm":
			result, err = transition.Confirm(r.Context(), id, principal.UserID, response.RequestID(r))
		default:
			response.WriteError(w, r, http.StatusNotFound, response.ErrorCodeNotFound, "Route not found", nil)
			return
		}
		if err != nil {
			writeSupplierRejectionError(w, r, err)
			return
		}
		response.WriteSuccess(w, r, http.StatusOK, supplierRejectionActionResultResponse{
			Rejection:      newSupplierRejectionResponse(result.Rejection, result.AuditLogID),
			PreviousStatus: string(result.PreviousStatus),
			CurrentStatus:  string(result.CurrentStatus),
			AuditLogID:     result.AuditLogID,
		})
	}
}

func newCreateSupplierRejectionLineInputs(
	rows []supplierRejectionLineRequest,
) []inventoryapp.CreateSupplierRejectionLineInput {
	inputs := make([]inventoryapp.CreateSupplierRejectionLineInput, 0, len(rows))
	for _, row := range rows {
		inputs = append(inputs, inventoryapp.CreateSupplierRejectionLineInput{
			ID:                    row.ID,
			PurchaseOrderLineID:   row.PurchaseOrderLineID,
			GoodsReceiptLineID:    row.GoodsReceiptLineID,
			InboundQCInspectionID: row.InboundQCInspectionID,
			ItemID:                row.ItemID,
			SKU:                   row.SKU,
			ItemName:              row.ItemName,
			BatchID:               row.BatchID,
			BatchNo:               row.BatchNo,
			LotNo:                 row.LotNo,
			ExpiryDate:            parseSupplierRejectionDate(row.ExpiryDate),
			RejectedQuantity:      row.RejectedQuantity,
			UOMCode:               row.UOMCode,
			BaseUOMCode:           row.BaseUOMCode,
			Reason:                row.Reason,
		})
	}

	return inputs
}

func newCreateSupplierRejectionAttachmentInputs(
	rows []supplierRejectionAttachmentRequest,
) []inventoryapp.CreateSupplierRejectionAttachmentInput {
	inputs := make([]inventoryapp.CreateSupplierRejectionAttachmentInput, 0, len(rows))
	for _, row := range rows {
		inputs = append(inputs, inventoryapp.CreateSupplierRejectionAttachmentInput{
			ID:          row.ID,
			LineID:      row.LineID,
			FileName:    row.FileName,
			ObjectKey:   row.ObjectKey,
			ContentType: row.ContentType,
			Source:      row.Source,
		})
	}

	return inputs
}

func newSupplierRejectionResponse(
	rejection inventorydomain.SupplierRejection,
	auditLogID string,
) supplierRejectionResponse {
	lines := make([]supplierRejectionLineResponse, 0, len(rejection.Lines))
	for _, line := range rejection.Lines {
		lines = append(lines, supplierRejectionLineResponse{
			ID:                    line.ID,
			PurchaseOrderLineID:   line.PurchaseOrderLineID,
			GoodsReceiptLineID:    line.GoodsReceiptLineID,
			InboundQCInspectionID: line.InboundQCInspectionID,
			ItemID:                line.ItemID,
			SKU:                   line.SKU,
			ItemName:              line.ItemName,
			BatchID:               line.BatchID,
			BatchNo:               line.BatchNo,
			LotNo:                 line.LotNo,
			ExpiryDate:            dateString(line.ExpiryDate),
			RejectedQuantity:      line.RejectedQuantity.String(),
			UOMCode:               line.UOMCode.String(),
			BaseUOMCode:           line.BaseUOMCode.String(),
			Reason:                line.Reason,
		})
	}
	attachments := make([]supplierRejectionAttachmentResponse, 0, len(rejection.Attachments))
	for _, attachment := range rejection.Attachments {
		attachments = append(attachments, supplierRejectionAttachmentResponse{
			ID:          attachment.ID,
			LineID:      attachment.LineID,
			FileName:    attachment.FileName,
			ObjectKey:   attachment.ObjectKey,
			ContentType: attachment.ContentType,
			UploadedAt:  attachment.UploadedAt.Format(time.RFC3339),
			UploadedBy:  attachment.UploadedBy,
			Source:      attachment.Source,
		})
	}

	return supplierRejectionResponse{
		ID:                    rejection.ID,
		OrgID:                 rejection.OrgID,
		RejectionNo:           rejection.RejectionNo,
		SupplierID:            rejection.SupplierID,
		SupplierCode:          rejection.SupplierCode,
		SupplierName:          rejection.SupplierName,
		PurchaseOrderID:       rejection.PurchaseOrderID,
		PurchaseOrderNo:       rejection.PurchaseOrderNo,
		GoodsReceiptID:        rejection.GoodsReceiptID,
		GoodsReceiptNo:        rejection.GoodsReceiptNo,
		InboundQCInspectionID: rejection.InboundQCInspectionID,
		WarehouseID:           rejection.WarehouseID,
		WarehouseCode:         rejection.WarehouseCode,
		Status:                string(rejection.Status),
		Reason:                rejection.Reason,
		Lines:                 lines,
		Attachments:           attachments,
		AuditLogID:            auditLogID,
		CreatedAt:             rejection.CreatedAt.Format(time.RFC3339),
		CreatedBy:             rejection.CreatedBy,
		UpdatedAt:             rejection.UpdatedAt.Format(time.RFC3339),
		UpdatedBy:             rejection.UpdatedBy,
		SubmittedAt:           timeString(rejection.SubmittedAt),
		SubmittedBy:           rejection.SubmittedBy,
		ConfirmedAt:           timeString(rejection.ConfirmedAt),
		ConfirmedBy:           rejection.ConfirmedBy,
	}
}

func parseSupplierRejectionDate(value string) time.Time {
	if value == "" {
		return time.Time{}
	}
	parsed, err := time.Parse("2006-01-02", value)
	if err != nil {
		return time.Time{}
	}

	return parsed
}

func writeSupplierRejectionError(w http.ResponseWriter, r *http.Request, err error) {
	switch {
	case errors.Is(err, inventoryapp.ErrSupplierRejectionNotFound):
		response.WriteError(w, r, http.StatusNotFound, response.ErrorCodeNotFound, "Supplier rejection not found", nil)
	case errors.Is(err, inventoryapp.ErrSupplierRejectionDuplicate):
		response.WriteError(w, r, http.StatusConflict, response.ErrorCodeConflict, "Supplier rejection already exists", nil)
	case errors.Is(err, inventorydomain.ErrSupplierRejectionRequiredField),
		errors.Is(err, inventorydomain.ErrSupplierRejectionInvalidQuantity),
		errors.Is(err, inventorydomain.ErrSupplierRejectionInvalidStatus),
		errors.Is(err, inventorydomain.ErrSupplierRejectionInvalidTransition):
		response.WriteError(w, r, http.StatusBadRequest, response.ErrorCodeValidation, "Invalid supplier rejection payload", nil)
	default:
		response.WriteError(w, r, http.StatusConflict, response.ErrorCodeConflict, "Supplier rejection could not be processed", nil)
	}
}
