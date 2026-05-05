package main

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"

	inventoryapp "github.com/Chinsusu/ERP-v2/apps/api/internal/modules/inventory/application"
	"github.com/Chinsusu/ERP-v2/apps/api/internal/modules/inventory/domain"
	"github.com/Chinsusu/ERP-v2/apps/api/internal/shared/auth"
	"github.com/Chinsusu/ERP-v2/apps/api/internal/shared/decimal"
	"github.com/Chinsusu/ERP-v2/apps/api/internal/shared/response"
)

type stockTransferLineRequest struct {
	ID                      string `json:"id"`
	ItemID                  string `json:"item_id"`
	SKU                     string `json:"sku"`
	BatchID                 string `json:"batch_id"`
	BatchNo                 string `json:"batch_no"`
	SourceLocationID        string `json:"source_location_id"`
	SourceLocationCode      string `json:"source_location_code"`
	DestinationLocationID   string `json:"destination_location_id"`
	DestinationLocationCode string `json:"destination_location_code"`
	Quantity                string `json:"quantity"`
	BaseUOMCode             string `json:"base_uom_code"`
	Note                    string `json:"note"`
}

type createStockTransferRequest struct {
	ID                       string                     `json:"id"`
	TransferNo               string                     `json:"transfer_no"`
	OrgID                    string                     `json:"org_id"`
	SourceWarehouseID        string                     `json:"source_warehouse_id"`
	SourceWarehouseCode      string                     `json:"source_warehouse_code"`
	DestinationWarehouseID   string                     `json:"destination_warehouse_id"`
	DestinationWarehouseCode string                     `json:"destination_warehouse_code"`
	ReasonCode               string                     `json:"reason_code"`
	Lines                    []stockTransferLineRequest `json:"lines"`
}

type stockTransferLineResponse struct {
	ID                      string `json:"id"`
	ItemID                  string `json:"item_id,omitempty"`
	SKU                     string `json:"sku"`
	BatchID                 string `json:"batch_id,omitempty"`
	BatchNo                 string `json:"batch_no,omitempty"`
	SourceLocationID        string `json:"source_location_id,omitempty"`
	SourceLocationCode      string `json:"source_location_code,omitempty"`
	DestinationLocationID   string `json:"destination_location_id,omitempty"`
	DestinationLocationCode string `json:"destination_location_code,omitempty"`
	Quantity                string `json:"quantity"`
	BaseUOMCode             string `json:"base_uom_code"`
	Note                    string `json:"note,omitempty"`
}

type stockTransferResponse struct {
	ID                       string                      `json:"id"`
	TransferNo               string                      `json:"transfer_no"`
	OrgID                    string                      `json:"org_id"`
	SourceWarehouseID        string                      `json:"source_warehouse_id"`
	SourceWarehouseCode      string                      `json:"source_warehouse_code,omitempty"`
	DestinationWarehouseID   string                      `json:"destination_warehouse_id"`
	DestinationWarehouseCode string                      `json:"destination_warehouse_code,omitempty"`
	ReasonCode               string                      `json:"reason_code"`
	Status                   string                      `json:"status"`
	RequestedBy              string                      `json:"requested_by"`
	SubmittedBy              string                      `json:"submitted_by,omitempty"`
	ApprovedBy               string                      `json:"approved_by,omitempty"`
	PostedBy                 string                      `json:"posted_by,omitempty"`
	Lines                    []stockTransferLineResponse `json:"lines"`
	AuditLogID               string                      `json:"audit_log_id,omitempty"`
	CreatedAt                string                      `json:"created_at"`
	UpdatedAt                string                      `json:"updated_at"`
	SubmittedAt              string                      `json:"submitted_at,omitempty"`
	ApprovedAt               string                      `json:"approved_at,omitempty"`
	PostedAt                 string                      `json:"posted_at,omitempty"`
}

type warehouseIssueLineRequest struct {
	ID                   string `json:"id"`
	ItemID               string `json:"item_id"`
	SKU                  string `json:"sku"`
	ItemName             string `json:"item_name"`
	Category             string `json:"category"`
	BatchID              string `json:"batch_id"`
	BatchNo              string `json:"batch_no"`
	LocationID           string `json:"location_id"`
	LocationCode         string `json:"location_code"`
	Quantity             string `json:"quantity"`
	BaseUOMCode          string `json:"base_uom_code"`
	Specification        string `json:"specification"`
	SourceDocumentType   string `json:"source_document_type"`
	SourceDocumentID     string `json:"source_document_id"`
	SourceDocumentLineID string `json:"source_document_line_id"`
	Note                 string `json:"note"`
}

type createWarehouseIssueRequest struct {
	ID              string                      `json:"id"`
	IssueNo         string                      `json:"issue_no"`
	OrgID           string                      `json:"org_id"`
	WarehouseID     string                      `json:"warehouse_id"`
	WarehouseCode   string                      `json:"warehouse_code"`
	DestinationType string                      `json:"destination_type"`
	DestinationName string                      `json:"destination_name"`
	ReasonCode      string                      `json:"reason_code"`
	Lines           []warehouseIssueLineRequest `json:"lines"`
}

type warehouseIssueLineResponse struct {
	ID                   string `json:"id"`
	ItemID               string `json:"item_id,omitempty"`
	SKU                  string `json:"sku"`
	ItemName             string `json:"item_name,omitempty"`
	Category             string `json:"category,omitempty"`
	BatchID              string `json:"batch_id,omitempty"`
	BatchNo              string `json:"batch_no,omitempty"`
	LocationID           string `json:"location_id,omitempty"`
	LocationCode         string `json:"location_code,omitempty"`
	Quantity             string `json:"quantity"`
	BaseUOMCode          string `json:"base_uom_code"`
	Specification        string `json:"specification,omitempty"`
	SourceDocumentType   string `json:"source_document_type,omitempty"`
	SourceDocumentID     string `json:"source_document_id,omitempty"`
	SourceDocumentLineID string `json:"source_document_line_id,omitempty"`
	Note                 string `json:"note,omitempty"`
}

type warehouseIssueResponse struct {
	ID              string                       `json:"id"`
	IssueNo         string                       `json:"issue_no"`
	OrgID           string                       `json:"org_id"`
	WarehouseID     string                       `json:"warehouse_id"`
	WarehouseCode   string                       `json:"warehouse_code,omitempty"`
	DestinationType string                       `json:"destination_type"`
	DestinationName string                       `json:"destination_name"`
	ReasonCode      string                       `json:"reason_code"`
	Status          string                       `json:"status"`
	RequestedBy     string                       `json:"requested_by"`
	SubmittedBy     string                       `json:"submitted_by,omitempty"`
	ApprovedBy      string                       `json:"approved_by,omitempty"`
	PostedBy        string                       `json:"posted_by,omitempty"`
	Lines           []warehouseIssueLineResponse `json:"lines"`
	AuditLogID      string                       `json:"audit_log_id,omitempty"`
	CreatedAt       string                       `json:"created_at"`
	UpdatedAt       string                       `json:"updated_at"`
	SubmittedAt     string                       `json:"submitted_at,omitempty"`
	ApprovedAt      string                       `json:"approved_at,omitempty"`
	PostedAt        string                       `json:"posted_at,omitempty"`
}

func stockTransfersHandler(service inventoryapp.StockTransferService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		principal, ok := auth.PrincipalFromContext(r.Context())
		if !ok {
			response.WriteError(w, r, http.StatusUnauthorized, response.ErrorCodeUnauthorized, "Authentication required", nil)
			return
		}
		switch r.Method {
		case http.MethodGet:
			if !auth.HasPermission(principal, auth.PermissionInventoryView) {
				writePermissionDenied(w, r, auth.PermissionInventoryView)
				return
			}
			transfers, err := service.ListStockTransfers(r.Context())
			if err != nil {
				writeStockTransferError(w, r, err)
				return
			}
			payload := make([]stockTransferResponse, 0, len(transfers))
			for _, transfer := range transfers {
				payload = append(payload, newStockTransferResponse(transfer, ""))
			}
			response.WriteSuccess(w, r, http.StatusOK, payload)
		case http.MethodPost:
			if !auth.HasPermission(principal, auth.PermissionRecordCreate) {
				writePermissionDenied(w, r, auth.PermissionRecordCreate)
				return
			}
			var payload createStockTransferRequest
			if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
				response.WriteError(w, r, http.StatusBadRequest, response.ErrorCodeValidation, "Invalid stock transfer payload", nil)
				return
			}
			result, err := service.CreateStockTransfer(r.Context(), inventoryapp.CreateStockTransferInput{
				ID:                       payload.ID,
				TransferNo:               payload.TransferNo,
				OrgID:                    payload.OrgID,
				SourceWarehouseID:        payload.SourceWarehouseID,
				SourceWarehouseCode:      payload.SourceWarehouseCode,
				DestinationWarehouseID:   payload.DestinationWarehouseID,
				DestinationWarehouseCode: payload.DestinationWarehouseCode,
				ReasonCode:               payload.ReasonCode,
				RequestedBy:              principal.UserID,
				RequestID:                response.RequestID(r),
				Lines:                    newCreateStockTransferLines(payload.Lines),
			})
			if err != nil {
				writeStockTransferError(w, r, err)
				return
			}
			response.WriteSuccess(w, r, http.StatusCreated, newStockTransferResponse(result.StockTransfer, result.AuditLogID))
		default:
			response.WriteError(w, r, http.StatusMethodNotAllowed, response.ErrorCodeNotFound, "Route not found", nil)
		}
	}
}

func stockTransferActionHandler(service inventoryapp.StockTransferService, action string) http.HandlerFunc {
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
		result, err := service.TransitionStockTransfer(r.Context(), inventoryapp.StockTransferTransitionInput{
			ID:        r.PathValue("stock_transfer_id"),
			ActorID:   principal.UserID,
			Action:    action,
			RequestID: response.RequestID(r),
		})
		if err != nil {
			writeStockTransferError(w, r, err)
			return
		}
		response.WriteSuccess(w, r, http.StatusOK, newStockTransferResponse(result.StockTransfer, result.AuditLogID))
	}
}

func warehouseIssuesHandler(service inventoryapp.WarehouseIssueService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		principal, ok := auth.PrincipalFromContext(r.Context())
		if !ok {
			response.WriteError(w, r, http.StatusUnauthorized, response.ErrorCodeUnauthorized, "Authentication required", nil)
			return
		}
		switch r.Method {
		case http.MethodGet:
			if !auth.HasPermission(principal, auth.PermissionInventoryView) {
				writePermissionDenied(w, r, auth.PermissionInventoryView)
				return
			}
			issues, err := service.ListWarehouseIssues(r.Context())
			if err != nil {
				writeWarehouseIssueError(w, r, err)
				return
			}
			payload := make([]warehouseIssueResponse, 0, len(issues))
			for _, issue := range issues {
				payload = append(payload, newWarehouseIssueResponse(issue, ""))
			}
			response.WriteSuccess(w, r, http.StatusOK, payload)
		case http.MethodPost:
			if !auth.HasPermission(principal, auth.PermissionRecordCreate) {
				writePermissionDenied(w, r, auth.PermissionRecordCreate)
				return
			}
			var payload createWarehouseIssueRequest
			if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
				response.WriteError(w, r, http.StatusBadRequest, response.ErrorCodeValidation, "Invalid warehouse issue payload", nil)
				return
			}
			result, err := service.CreateWarehouseIssue(r.Context(), inventoryapp.CreateWarehouseIssueInput{
				ID:              payload.ID,
				IssueNo:         payload.IssueNo,
				OrgID:           payload.OrgID,
				WarehouseID:     payload.WarehouseID,
				WarehouseCode:   payload.WarehouseCode,
				DestinationType: payload.DestinationType,
				DestinationName: payload.DestinationName,
				ReasonCode:      payload.ReasonCode,
				RequestedBy:     principal.UserID,
				RequestID:       response.RequestID(r),
				Lines:           newCreateWarehouseIssueLines(payload.Lines),
			})
			if err != nil {
				writeWarehouseIssueError(w, r, err)
				return
			}
			response.WriteSuccess(w, r, http.StatusCreated, newWarehouseIssueResponse(result.WarehouseIssue, result.AuditLogID))
		default:
			response.WriteError(w, r, http.StatusMethodNotAllowed, response.ErrorCodeNotFound, "Route not found", nil)
		}
	}
}

func warehouseIssueActionHandler(service inventoryapp.WarehouseIssueService, action string) http.HandlerFunc {
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
		result, err := service.TransitionWarehouseIssue(r.Context(), inventoryapp.WarehouseIssueTransitionInput{
			ID:        r.PathValue("warehouse_issue_id"),
			ActorID:   principal.UserID,
			Action:    action,
			RequestID: response.RequestID(r),
		})
		if err != nil {
			writeWarehouseIssueError(w, r, err)
			return
		}
		response.WriteSuccess(w, r, http.StatusOK, newWarehouseIssueResponse(result.WarehouseIssue, result.AuditLogID))
	}
}

func newCreateStockTransferLines(inputs []stockTransferLineRequest) []inventoryapp.CreateStockTransferLineInput {
	lines := make([]inventoryapp.CreateStockTransferLineInput, 0, len(inputs))
	for _, input := range inputs {
		lines = append(lines, inventoryapp.CreateStockTransferLineInput{
			ID:                      input.ID,
			ItemID:                  input.ItemID,
			SKU:                     input.SKU,
			BatchID:                 input.BatchID,
			BatchNo:                 input.BatchNo,
			SourceLocationID:        input.SourceLocationID,
			SourceLocationCode:      input.SourceLocationCode,
			DestinationLocationID:   input.DestinationLocationID,
			DestinationLocationCode: input.DestinationLocationCode,
			Quantity:                input.Quantity,
			BaseUOMCode:             input.BaseUOMCode,
			Note:                    input.Note,
		})
	}

	return lines
}

func newCreateWarehouseIssueLines(inputs []warehouseIssueLineRequest) []inventoryapp.CreateWarehouseIssueLineInput {
	lines := make([]inventoryapp.CreateWarehouseIssueLineInput, 0, len(inputs))
	for _, input := range inputs {
		lines = append(lines, inventoryapp.CreateWarehouseIssueLineInput{
			ID:                   input.ID,
			ItemID:               input.ItemID,
			SKU:                  input.SKU,
			ItemName:             input.ItemName,
			Category:             input.Category,
			BatchID:              input.BatchID,
			BatchNo:              input.BatchNo,
			LocationID:           input.LocationID,
			LocationCode:         input.LocationCode,
			Quantity:             input.Quantity,
			BaseUOMCode:          input.BaseUOMCode,
			Specification:        input.Specification,
			SourceDocumentType:   input.SourceDocumentType,
			SourceDocumentID:     input.SourceDocumentID,
			SourceDocumentLineID: input.SourceDocumentLineID,
			Note:                 input.Note,
		})
	}

	return lines
}

func newStockTransferResponse(transfer domain.StockTransfer, auditLogID string) stockTransferResponse {
	payload := stockTransferResponse{
		ID:                       transfer.ID,
		TransferNo:               transfer.TransferNo,
		OrgID:                    transfer.OrgID,
		SourceWarehouseID:        transfer.SourceWarehouseID,
		SourceWarehouseCode:      transfer.SourceWarehouseCode,
		DestinationWarehouseID:   transfer.DestinationWarehouseID,
		DestinationWarehouseCode: transfer.DestinationWarehouseCode,
		ReasonCode:               transfer.ReasonCode,
		Status:                   string(transfer.Status),
		RequestedBy:              transfer.RequestedBy,
		SubmittedBy:              transfer.SubmittedBy,
		ApprovedBy:               transfer.ApprovedBy,
		PostedBy:                 transfer.PostedBy,
		Lines:                    make([]stockTransferLineResponse, 0, len(transfer.Lines)),
		AuditLogID:               auditLogID,
		CreatedAt:                transfer.CreatedAt.UTC().Format(time.RFC3339),
		UpdatedAt:                transfer.UpdatedAt.UTC().Format(time.RFC3339),
		SubmittedAt:              timeString(transfer.SubmittedAt),
		ApprovedAt:               timeString(transfer.ApprovedAt),
		PostedAt:                 timeString(transfer.PostedAt),
	}
	for _, line := range transfer.Lines {
		payload.Lines = append(payload.Lines, stockTransferLineResponse{
			ID:                      line.ID,
			ItemID:                  line.ItemID,
			SKU:                     line.SKU,
			BatchID:                 line.BatchID,
			BatchNo:                 line.BatchNo,
			SourceLocationID:        line.SourceLocationID,
			SourceLocationCode:      line.SourceLocationCode,
			DestinationLocationID:   line.DestinationLocationID,
			DestinationLocationCode: line.DestinationLocationCode,
			Quantity:                line.Quantity.String(),
			BaseUOMCode:             line.BaseUOMCode.String(),
			Note:                    line.Note,
		})
	}

	return payload
}

func newWarehouseIssueResponse(issue domain.WarehouseIssue, auditLogID string) warehouseIssueResponse {
	payload := warehouseIssueResponse{
		ID:              issue.ID,
		IssueNo:         issue.IssueNo,
		OrgID:           issue.OrgID,
		WarehouseID:     issue.WarehouseID,
		WarehouseCode:   issue.WarehouseCode,
		DestinationType: issue.DestinationType,
		DestinationName: issue.DestinationName,
		ReasonCode:      issue.ReasonCode,
		Status:          string(issue.Status),
		RequestedBy:     issue.RequestedBy,
		SubmittedBy:     issue.SubmittedBy,
		ApprovedBy:      issue.ApprovedBy,
		PostedBy:        issue.PostedBy,
		Lines:           make([]warehouseIssueLineResponse, 0, len(issue.Lines)),
		AuditLogID:      auditLogID,
		CreatedAt:       issue.CreatedAt.UTC().Format(time.RFC3339),
		UpdatedAt:       issue.UpdatedAt.UTC().Format(time.RFC3339),
		SubmittedAt:     timeString(issue.SubmittedAt),
		ApprovedAt:      timeString(issue.ApprovedAt),
		PostedAt:        timeString(issue.PostedAt),
	}
	for _, line := range issue.Lines {
		payload.Lines = append(payload.Lines, warehouseIssueLineResponse{
			ID:                   line.ID,
			ItemID:               line.ItemID,
			SKU:                  line.SKU,
			ItemName:             line.ItemName,
			Category:             line.Category,
			BatchID:              line.BatchID,
			BatchNo:              line.BatchNo,
			LocationID:           line.LocationID,
			LocationCode:         line.LocationCode,
			Quantity:             line.Quantity.String(),
			BaseUOMCode:          line.BaseUOMCode.String(),
			Specification:        line.Specification,
			SourceDocumentType:   line.SourceDocumentType,
			SourceDocumentID:     line.SourceDocumentID,
			SourceDocumentLineID: line.SourceDocumentLineID,
			Note:                 line.Note,
		})
	}

	return payload
}

func writeStockTransferError(w http.ResponseWriter, r *http.Request, err error) {
	switch {
	case errors.Is(err, inventoryapp.ErrStockTransferNotFound):
		response.WriteError(w, r, http.StatusNotFound, response.ErrorCodeNotFound, "Stock transfer not found", nil)
	case errors.Is(err, domain.ErrStockTransferRequiredField):
		response.WriteError(w, r, http.StatusBadRequest, response.ErrorCodeValidation, "Stock transfer required field is missing", nil)
	case errors.Is(err, domain.ErrStockTransferInvalidQuantity),
		errors.Is(err, decimal.ErrInvalidDecimal),
		errors.Is(err, decimal.ErrInvalidUOMCode):
		response.WriteError(w, r, http.StatusBadRequest, response.ErrorCodeValidation, "Stock transfer quantity or UOM is invalid", nil)
	case errors.Is(err, domain.ErrStockTransferSameWarehouse):
		response.WriteError(w, r, http.StatusBadRequest, response.ErrorCodeValidation, "Stock transfer source and destination warehouse must differ", nil)
	case errors.Is(err, domain.ErrStockTransferInvalidStatus):
		response.WriteError(w, r, http.StatusConflict, response.ErrorCodeConflict, "Stock transfer status transition is not allowed", nil)
	default:
		response.WriteError(w, r, http.StatusConflict, response.ErrorCodeConflict, "Stock transfer could not be processed", nil)
	}
}

func writeWarehouseIssueError(w http.ResponseWriter, r *http.Request, err error) {
	switch {
	case errors.Is(err, inventoryapp.ErrWarehouseIssueNotFound):
		response.WriteError(w, r, http.StatusNotFound, response.ErrorCodeNotFound, "Warehouse issue not found", nil)
	case errors.Is(err, domain.ErrWarehouseIssueRequiredField):
		response.WriteError(w, r, http.StatusBadRequest, response.ErrorCodeValidation, "Warehouse issue required field is missing", nil)
	case errors.Is(err, domain.ErrWarehouseIssueInvalidQuantity),
		errors.Is(err, decimal.ErrInvalidDecimal),
		errors.Is(err, decimal.ErrInvalidUOMCode):
		response.WriteError(w, r, http.StatusBadRequest, response.ErrorCodeValidation, "Warehouse issue quantity or UOM is invalid", nil)
	case errors.Is(err, domain.ErrWarehouseIssueInvalidStatus):
		response.WriteError(w, r, http.StatusConflict, response.ErrorCodeConflict, "Warehouse issue status transition is not allowed", nil)
	default:
		response.WriteError(w, r, http.StatusConflict, response.ErrorCodeConflict, "Warehouse issue could not be processed", nil)
	}
}
