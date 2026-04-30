package main

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"

	financeapp "github.com/Chinsusu/ERP-v2/apps/api/internal/modules/finance/application"
	financedomain "github.com/Chinsusu/ERP-v2/apps/api/internal/modules/finance/domain"
	"github.com/Chinsusu/ERP-v2/apps/api/internal/shared/auth"
	apperrors "github.com/Chinsusu/ERP-v2/apps/api/internal/shared/errors"
	"github.com/Chinsusu/ERP-v2/apps/api/internal/shared/response"
)

type supplierPayableSourceDocumentRequest struct {
	Type string `json:"type"`
	ID   string `json:"id"`
	No   string `json:"no"`
}

type supplierPayableLineRequest struct {
	ID             string                               `json:"id"`
	Description    string                               `json:"description"`
	SourceDocument supplierPayableSourceDocumentRequest `json:"source_document"`
	Amount         string                               `json:"amount"`
}

type createSupplierPayableRequest struct {
	ID             string                               `json:"id"`
	PayableNo      string                               `json:"payable_no"`
	SupplierID     string                               `json:"supplier_id"`
	SupplierCode   string                               `json:"supplier_code"`
	SupplierName   string                               `json:"supplier_name"`
	Status         string                               `json:"status"`
	SourceDocument supplierPayableSourceDocumentRequest `json:"source_document"`
	Lines          []supplierPayableLineRequest         `json:"lines"`
	TotalAmount    string                               `json:"total_amount"`
	CurrencyCode   string                               `json:"currency_code"`
	DueDate        string                               `json:"due_date"`
}

type supplierPayableActionRequest struct {
	Amount string `json:"amount"`
	Reason string `json:"reason"`
}

type supplierPayableSourceDocumentResponse struct {
	Type string `json:"type"`
	ID   string `json:"id,omitempty"`
	No   string `json:"no,omitempty"`
}

type supplierPayableLineResponse struct {
	ID             string                                `json:"id"`
	Description    string                                `json:"description"`
	SourceDocument supplierPayableSourceDocumentResponse `json:"source_document"`
	Amount         string                                `json:"amount"`
}

type supplierPayableListItemResponse struct {
	ID                string `json:"id"`
	PayableNo         string `json:"payable_no"`
	SupplierID        string `json:"supplier_id"`
	SupplierCode      string `json:"supplier_code,omitempty"`
	SupplierName      string `json:"supplier_name"`
	Status            string `json:"status"`
	TotalAmount       string `json:"total_amount"`
	PaidAmount        string `json:"paid_amount"`
	OutstandingAmount string `json:"outstanding_amount"`
	CurrencyCode      string `json:"currency_code"`
	DueDate           string `json:"due_date,omitempty"`
	CreatedAt         string `json:"created_at"`
	UpdatedAt         string `json:"updated_at"`
	Version           int    `json:"version"`
}

type supplierPayableResponse struct {
	ID                  string                                `json:"id"`
	OrgID               string                                `json:"org_id"`
	PayableNo           string                                `json:"payable_no"`
	SupplierID          string                                `json:"supplier_id"`
	SupplierCode        string                                `json:"supplier_code,omitempty"`
	SupplierName        string                                `json:"supplier_name"`
	Status              string                                `json:"status"`
	SourceDocument      supplierPayableSourceDocumentResponse `json:"source_document"`
	Lines               []supplierPayableLineResponse         `json:"lines"`
	TotalAmount         string                                `json:"total_amount"`
	PaidAmount          string                                `json:"paid_amount"`
	OutstandingAmount   string                                `json:"outstanding_amount"`
	CurrencyCode        string                                `json:"currency_code"`
	DueDate             string                                `json:"due_date,omitempty"`
	PaymentRequestedBy  string                                `json:"payment_requested_by,omitempty"`
	PaymentRequestedAt  string                                `json:"payment_requested_at,omitempty"`
	PaymentApprovedBy   string                                `json:"payment_approved_by,omitempty"`
	PaymentApprovedAt   string                                `json:"payment_approved_at,omitempty"`
	PaymentRejectedBy   string                                `json:"payment_rejected_by,omitempty"`
	PaymentRejectedAt   string                                `json:"payment_rejected_at,omitempty"`
	PaymentRejectReason string                                `json:"payment_reject_reason,omitempty"`
	DisputeReason       string                                `json:"dispute_reason,omitempty"`
	VoidReason          string                                `json:"void_reason,omitempty"`
	AuditLogID          string                                `json:"audit_log_id,omitempty"`
	CreatedAt           string                                `json:"created_at"`
	UpdatedAt           string                                `json:"updated_at"`
	Version             int                                   `json:"version"`
}

type supplierPayableActionResultResponse struct {
	SupplierPayable supplierPayableResponse `json:"supplier_payable"`
	PreviousStatus  string                  `json:"previous_status"`
	CurrentStatus   string                  `json:"current_status"`
	AuditLogID      string                  `json:"audit_log_id,omitempty"`
}

func supplierPayablesHandler(service financeapp.SupplierPayableService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		principal, ok := auth.PrincipalFromContext(r.Context())
		if !ok {
			response.WriteError(w, r, http.StatusUnauthorized, response.ErrorCodeUnauthorized, "Authentication required", nil)
			return
		}
		if !auth.HasPermission(principal, auth.PermissionFinanceView) {
			writePermissionDenied(w, r, auth.PermissionFinanceView)
			return
		}

		switch r.Method {
		case http.MethodGet:
			payables, err := service.ListSupplierPayables(r.Context(), supplierPayableFilterFromRequest(r))
			if err != nil {
				writeSupplierPayableError(w, r, err)
				return
			}
			payload := make([]supplierPayableListItemResponse, 0, len(payables))
			for _, payable := range payables {
				payload = append(payload, newSupplierPayableListItemResponse(payable))
			}
			response.WriteSuccess(w, r, http.StatusOK, payload)
		case http.MethodPost:
			if !auth.HasPermission(principal, auth.PermissionFinanceManage) {
				writePermissionDenied(w, r, auth.PermissionFinanceManage)
				return
			}
			r = requestWithStableID(r)
			var payload createSupplierPayableRequest
			if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
				response.WriteError(w, r, http.StatusBadRequest, response.ErrorCodeValidation, "Invalid supplier payable payload", nil)
				return
			}

			result, err := service.CreateSupplierPayable(r.Context(), financeapp.CreateSupplierPayableInput{
				ID:             payload.ID,
				PayableNo:      payload.PayableNo,
				SupplierID:     payload.SupplierID,
				SupplierCode:   payload.SupplierCode,
				SupplierName:   payload.SupplierName,
				Status:         payload.Status,
				SourceDocument: supplierPayableSourceDocumentInput(payload.SourceDocument),
				Lines:          supplierPayableLineInputs(payload.Lines),
				TotalAmount:    payload.TotalAmount,
				CurrencyCode:   payload.CurrencyCode,
				DueDate:        payload.DueDate,
				ActorID:        principal.UserID,
				RequestID:      response.RequestID(r),
			})
			if err != nil {
				writeSupplierPayableError(w, r, err)
				return
			}

			response.WriteSuccess(w, r, http.StatusCreated, newSupplierPayableResponse(result.SupplierPayable, result.AuditLogID))
		default:
			response.WriteError(w, r, http.StatusMethodNotAllowed, response.ErrorCodeNotFound, "Route not found", nil)
		}
	}
}

func supplierPayableDetailHandler(service financeapp.SupplierPayableService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		principal, ok := auth.PrincipalFromContext(r.Context())
		if !ok {
			response.WriteError(w, r, http.StatusUnauthorized, response.ErrorCodeUnauthorized, "Authentication required", nil)
			return
		}
		if !auth.HasPermission(principal, auth.PermissionFinanceView) {
			writePermissionDenied(w, r, auth.PermissionFinanceView)
			return
		}
		if r.Method != http.MethodGet {
			response.WriteError(w, r, http.StatusMethodNotAllowed, response.ErrorCodeNotFound, "Route not found", nil)
			return
		}

		payable, err := service.GetSupplierPayable(r.Context(), r.PathValue("supplier_payable_id"))
		if err != nil {
			writeSupplierPayableError(w, r, err)
			return
		}
		response.WriteSuccess(w, r, http.StatusOK, newSupplierPayableResponse(payable, ""))
	}
}

func supplierPayableApprovePaymentHandler(service financeapp.SupplierPayableService) http.HandlerFunc {
	return supplierPayableActionHandler(service, "approve-payment")
}

func supplierPayableRequestPaymentHandler(service financeapp.SupplierPayableService) http.HandlerFunc {
	return supplierPayableActionHandler(service, "request-payment")
}

func supplierPayableRejectPaymentHandler(service financeapp.SupplierPayableService) http.HandlerFunc {
	return supplierPayableActionHandler(service, "reject-payment")
}

func supplierPayableRecordPaymentHandler(service financeapp.SupplierPayableService) http.HandlerFunc {
	return supplierPayableActionHandler(service, "record-payment")
}

func supplierPayableVoidHandler(service financeapp.SupplierPayableService) http.HandlerFunc {
	return supplierPayableActionHandler(service, "void")
}

func supplierPayableActionHandler(
	service financeapp.SupplierPayableService,
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
		if !auth.HasPermission(principal, auth.PermissionFinanceView) {
			writePermissionDenied(w, r, auth.PermissionFinanceView)
			return
		}
		if action == "approve-payment" || action == "reject-payment" {
			if !auth.HasPermission(principal, auth.PermissionPaymentApprove) {
				writePermissionDenied(w, r, auth.PermissionPaymentApprove)
				return
			}
		} else if !auth.HasPermission(principal, auth.PermissionFinanceManage) {
			writePermissionDenied(w, r, auth.PermissionFinanceManage)
			return
		}

		r = requestWithStableID(r)
		var payload supplierPayableActionRequest
		if r.Body != nil {
			if err := json.NewDecoder(r.Body).Decode(&payload); err != nil && !errors.Is(err, io.EOF) {
				response.WriteError(w, r, http.StatusBadRequest, response.ErrorCodeValidation, "Invalid supplier payable action payload", nil)
				return
			}
		}

		input := financeapp.SupplierPayableActionInput{
			ID:        r.PathValue("supplier_payable_id"),
			Amount:    payload.Amount,
			Reason:    payload.Reason,
			ActorID:   principal.UserID,
			RequestID: response.RequestID(r),
		}
		var (
			result financeapp.SupplierPayableActionResult
			err    error
		)
		switch action {
		case "request-payment":
			result, err = service.RequestSupplierPayablePayment(r.Context(), input)
		case "approve-payment":
			result, err = service.ApproveSupplierPayablePayment(r.Context(), input)
		case "reject-payment":
			result, err = service.RejectSupplierPayablePayment(r.Context(), input)
		case "record-payment":
			result, err = service.RecordSupplierPayablePayment(r.Context(), input)
		case "void":
			result, err = service.VoidSupplierPayable(r.Context(), input)
		default:
			response.WriteError(w, r, http.StatusNotFound, response.ErrorCodeNotFound, "Route not found", nil)
			return
		}
		if err != nil {
			writeSupplierPayableError(w, r, err)
			return
		}

		response.WriteSuccess(w, r, http.StatusOK, supplierPayableActionResultResponse{
			SupplierPayable: newSupplierPayableResponse(result.SupplierPayable, ""),
			PreviousStatus:  string(result.PreviousStatus),
			CurrentStatus:   string(result.CurrentStatus),
			AuditLogID:      result.AuditLogID,
		})
	}
}

func supplierPayableFilterFromRequest(r *http.Request) financeapp.SupplierPayableFilter {
	query := r.URL.Query()
	search := strings.TrimSpace(query.Get("search"))
	if search == "" {
		search = strings.TrimSpace(query.Get("q"))
	}
	statuses := []financedomain.PayableStatus{}
	for _, rawStatus := range strings.Split(query.Get("status"), ",") {
		status := financedomain.NormalizePayableStatus(financedomain.PayableStatus(rawStatus))
		if status != "" {
			statuses = append(statuses, status)
		}
	}

	return financeapp.SupplierPayableFilter{
		Search:     search,
		Statuses:   statuses,
		SupplierID: query.Get("supplier_id"),
	}
}

func supplierPayableSourceDocumentInput(
	input supplierPayableSourceDocumentRequest,
) financeapp.SourceDocumentInput {
	return financeapp.SourceDocumentInput{
		Type: input.Type,
		ID:   input.ID,
		No:   input.No,
	}
}

func supplierPayableLineInputs(inputs []supplierPayableLineRequest) []financeapp.SupplierPayableLineInput {
	lines := make([]financeapp.SupplierPayableLineInput, 0, len(inputs))
	for _, input := range inputs {
		lines = append(lines, financeapp.SupplierPayableLineInput{
			ID:             input.ID,
			Description:    input.Description,
			SourceDocument: supplierPayableSourceDocumentInput(input.SourceDocument),
			Amount:         input.Amount,
		})
	}

	return lines
}

func newSupplierPayableListItemResponse(
	payable financedomain.SupplierPayable,
) supplierPayableListItemResponse {
	return supplierPayableListItemResponse{
		ID:                payable.ID,
		PayableNo:         payable.PayableNo,
		SupplierID:        payable.SupplierID,
		SupplierCode:      payable.SupplierCode,
		SupplierName:      payable.SupplierName,
		Status:            string(payable.Status),
		TotalAmount:       payable.TotalAmount.String(),
		PaidAmount:        payable.PaidAmount.String(),
		OutstandingAmount: payable.OutstandingAmount.String(),
		CurrencyCode:      payable.CurrencyCode.String(),
		DueDate:           dateString(payable.DueDate),
		CreatedAt:         timeString(payable.CreatedAt),
		UpdatedAt:         timeString(payable.UpdatedAt),
		Version:           payable.Version,
	}
}

func newSupplierPayableResponse(
	payable financedomain.SupplierPayable,
	auditLogID string,
) supplierPayableResponse {
	payload := supplierPayableResponse{
		ID:                  payable.ID,
		OrgID:               payable.OrgID,
		PayableNo:           payable.PayableNo,
		SupplierID:          payable.SupplierID,
		SupplierCode:        payable.SupplierCode,
		SupplierName:        payable.SupplierName,
		Status:              string(payable.Status),
		SourceDocument:      newSupplierPayableSourceDocumentResponse(payable.SourceDocument),
		Lines:               make([]supplierPayableLineResponse, 0, len(payable.Lines)),
		TotalAmount:         payable.TotalAmount.String(),
		PaidAmount:          payable.PaidAmount.String(),
		OutstandingAmount:   payable.OutstandingAmount.String(),
		CurrencyCode:        payable.CurrencyCode.String(),
		DueDate:             dateString(payable.DueDate),
		PaymentRequestedBy:  payable.PaymentRequestedBy,
		PaymentRequestedAt:  timeString(payable.PaymentRequestedAt),
		PaymentApprovedBy:   payable.PaymentApprovedBy,
		PaymentApprovedAt:   timeString(payable.PaymentApprovedAt),
		PaymentRejectedBy:   payable.PaymentRejectedBy,
		PaymentRejectedAt:   timeString(payable.PaymentRejectedAt),
		PaymentRejectReason: payable.PaymentRejectReason,
		DisputeReason:       payable.DisputeReason,
		VoidReason:          payable.VoidReason,
		AuditLogID:          auditLogID,
		CreatedAt:           timeString(payable.CreatedAt),
		UpdatedAt:           timeString(payable.UpdatedAt),
		Version:             payable.Version,
	}
	for _, line := range payable.Lines {
		payload.Lines = append(payload.Lines, supplierPayableLineResponse{
			ID:             line.ID,
			Description:    line.Description,
			SourceDocument: newSupplierPayableSourceDocumentResponse(line.SourceDocument),
			Amount:         line.Amount.String(),
		})
	}

	return payload
}

func newSupplierPayableSourceDocumentResponse(
	source financedomain.SourceDocumentRef,
) supplierPayableSourceDocumentResponse {
	return supplierPayableSourceDocumentResponse{
		Type: string(source.Type),
		ID:   source.ID,
		No:   source.No,
	}
}

func writeSupplierPayableError(w http.ResponseWriter, r *http.Request, err error) {
	if appErr, ok := apperrors.As(err); ok {
		response.WriteError(w, r, appErr.HTTPStatus, appErr.Code, appErr.Message, appErr.Details)
		return
	}

	response.WriteError(w, r, http.StatusConflict, response.ErrorCodeConflict, "Supplier payable request could not be processed", nil)
}
