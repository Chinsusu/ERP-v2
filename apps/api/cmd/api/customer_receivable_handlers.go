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

type customerReceivableSourceDocumentRequest struct {
	Type string `json:"type"`
	ID   string `json:"id"`
	No   string `json:"no"`
}

type customerReceivableLineRequest struct {
	ID             string                                  `json:"id"`
	Description    string                                  `json:"description"`
	SourceDocument customerReceivableSourceDocumentRequest `json:"source_document"`
	Amount         string                                  `json:"amount"`
}

type createCustomerReceivableRequest struct {
	ID             string                                  `json:"id"`
	ReceivableNo   string                                  `json:"receivable_no"`
	CustomerID     string                                  `json:"customer_id"`
	CustomerCode   string                                  `json:"customer_code"`
	CustomerName   string                                  `json:"customer_name"`
	Status         string                                  `json:"status"`
	SourceDocument customerReceivableSourceDocumentRequest `json:"source_document"`
	Lines          []customerReceivableLineRequest         `json:"lines"`
	TotalAmount    string                                  `json:"total_amount"`
	CurrencyCode   string                                  `json:"currency_code"`
	DueDate        string                                  `json:"due_date"`
}

type customerReceivableActionRequest struct {
	Amount string `json:"amount"`
	Reason string `json:"reason"`
}

type customerReceivableSourceDocumentResponse struct {
	Type string `json:"type"`
	ID   string `json:"id,omitempty"`
	No   string `json:"no,omitempty"`
}

type customerReceivableLineResponse struct {
	ID             string                                   `json:"id"`
	Description    string                                   `json:"description"`
	SourceDocument customerReceivableSourceDocumentResponse `json:"source_document"`
	Amount         string                                   `json:"amount"`
}

type customerReceivableListItemResponse struct {
	ID                string `json:"id"`
	ReceivableNo      string `json:"receivable_no"`
	CustomerID        string `json:"customer_id"`
	CustomerCode      string `json:"customer_code,omitempty"`
	CustomerName      string `json:"customer_name"`
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

type customerReceivableResponse struct {
	ID                string                                   `json:"id"`
	OrgID             string                                   `json:"org_id"`
	ReceivableNo      string                                   `json:"receivable_no"`
	CustomerID        string                                   `json:"customer_id"`
	CustomerCode      string                                   `json:"customer_code,omitempty"`
	CustomerName      string                                   `json:"customer_name"`
	Status            string                                   `json:"status"`
	SourceDocument    customerReceivableSourceDocumentResponse `json:"source_document"`
	Lines             []customerReceivableLineResponse         `json:"lines"`
	TotalAmount       string                                   `json:"total_amount"`
	PaidAmount        string                                   `json:"paid_amount"`
	OutstandingAmount string                                   `json:"outstanding_amount"`
	CurrencyCode      string                                   `json:"currency_code"`
	DueDate           string                                   `json:"due_date,omitempty"`
	DisputeReason     string                                   `json:"dispute_reason,omitempty"`
	VoidReason        string                                   `json:"void_reason,omitempty"`
	AuditLogID        string                                   `json:"audit_log_id,omitempty"`
	CreatedAt         string                                   `json:"created_at"`
	UpdatedAt         string                                   `json:"updated_at"`
	Version           int                                      `json:"version"`
}

type customerReceivableActionResultResponse struct {
	CustomerReceivable customerReceivableResponse `json:"customer_receivable"`
	PreviousStatus     string                     `json:"previous_status"`
	CurrentStatus      string                     `json:"current_status"`
	AuditLogID         string                     `json:"audit_log_id,omitempty"`
}

func customerReceivablesHandler(service financeapp.CustomerReceivableService) http.HandlerFunc {
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
			receivables, err := service.ListCustomerReceivables(r.Context(), customerReceivableFilterFromRequest(r))
			if err != nil {
				writeCustomerReceivableError(w, r, err)
				return
			}
			payload := make([]customerReceivableListItemResponse, 0, len(receivables))
			for _, receivable := range receivables {
				payload = append(payload, newCustomerReceivableListItemResponse(receivable))
			}
			response.WriteSuccess(w, r, http.StatusOK, payload)
		case http.MethodPost:
			if !auth.HasPermission(principal, auth.PermissionFinanceManage) {
				writePermissionDenied(w, r, auth.PermissionFinanceManage)
				return
			}
			r = requestWithStableID(r)
			var payload createCustomerReceivableRequest
			if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
				response.WriteError(w, r, http.StatusBadRequest, response.ErrorCodeValidation, "Invalid customer receivable payload", nil)
				return
			}

			result, err := service.CreateCustomerReceivable(r.Context(), financeapp.CreateCustomerReceivableInput{
				ID:             payload.ID,
				ReceivableNo:   payload.ReceivableNo,
				CustomerID:     payload.CustomerID,
				CustomerCode:   payload.CustomerCode,
				CustomerName:   payload.CustomerName,
				Status:         payload.Status,
				SourceDocument: customerReceivableSourceDocumentInput(payload.SourceDocument),
				Lines:          customerReceivableLineInputs(payload.Lines),
				TotalAmount:    payload.TotalAmount,
				CurrencyCode:   payload.CurrencyCode,
				DueDate:        payload.DueDate,
				ActorID:        principal.UserID,
				RequestID:      response.RequestID(r),
			})
			if err != nil {
				writeCustomerReceivableError(w, r, err)
				return
			}

			response.WriteSuccess(w, r, http.StatusCreated, newCustomerReceivableResponse(result.CustomerReceivable, result.AuditLogID))
		default:
			response.WriteError(w, r, http.StatusMethodNotAllowed, response.ErrorCodeNotFound, "Route not found", nil)
		}
	}
}

func customerReceivableDetailHandler(service financeapp.CustomerReceivableService) http.HandlerFunc {
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

		receivable, err := service.GetCustomerReceivable(r.Context(), r.PathValue("customer_receivable_id"))
		if err != nil {
			writeCustomerReceivableError(w, r, err)
			return
		}
		response.WriteSuccess(w, r, http.StatusOK, newCustomerReceivableResponse(receivable, ""))
	}
}

func customerReceivableRecordReceiptHandler(service financeapp.CustomerReceivableService) http.HandlerFunc {
	return customerReceivableActionHandler(service, "record-receipt")
}

func customerReceivableMarkDisputedHandler(service financeapp.CustomerReceivableService) http.HandlerFunc {
	return customerReceivableActionHandler(service, "mark-disputed")
}

func customerReceivableVoidHandler(service financeapp.CustomerReceivableService) http.HandlerFunc {
	return customerReceivableActionHandler(service, "void")
}

func customerReceivableActionHandler(
	service financeapp.CustomerReceivableService,
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
		if !auth.HasPermission(principal, auth.PermissionFinanceManage) {
			writePermissionDenied(w, r, auth.PermissionFinanceManage)
			return
		}

		r = requestWithStableID(r)
		var payload customerReceivableActionRequest
		if r.Body != nil {
			if err := json.NewDecoder(r.Body).Decode(&payload); err != nil && !errors.Is(err, io.EOF) {
				response.WriteError(w, r, http.StatusBadRequest, response.ErrorCodeValidation, "Invalid customer receivable action payload", nil)
				return
			}
		}

		input := financeapp.CustomerReceivableActionInput{
			ID:        r.PathValue("customer_receivable_id"),
			Amount:    payload.Amount,
			Reason:    payload.Reason,
			ActorID:   principal.UserID,
			RequestID: response.RequestID(r),
		}
		var (
			result financeapp.CustomerReceivableActionResult
			err    error
		)
		switch action {
		case "record-receipt":
			result, err = service.RecordCustomerReceivableReceipt(r.Context(), input)
		case "mark-disputed":
			result, err = service.MarkCustomerReceivableDisputed(r.Context(), input)
		case "void":
			result, err = service.VoidCustomerReceivable(r.Context(), input)
		default:
			response.WriteError(w, r, http.StatusNotFound, response.ErrorCodeNotFound, "Route not found", nil)
			return
		}
		if err != nil {
			writeCustomerReceivableError(w, r, err)
			return
		}

		response.WriteSuccess(w, r, http.StatusOK, customerReceivableActionResultResponse{
			CustomerReceivable: newCustomerReceivableResponse(result.CustomerReceivable, ""),
			PreviousStatus:     string(result.PreviousStatus),
			CurrentStatus:      string(result.CurrentStatus),
			AuditLogID:         result.AuditLogID,
		})
	}
}

func customerReceivableFilterFromRequest(r *http.Request) financeapp.CustomerReceivableFilter {
	query := r.URL.Query()
	statuses := []financedomain.ReceivableStatus{}
	for _, rawStatus := range strings.Split(query.Get("status"), ",") {
		status := financedomain.NormalizeReceivableStatus(financedomain.ReceivableStatus(rawStatus))
		if status != "" {
			statuses = append(statuses, status)
		}
	}

	return financeapp.CustomerReceivableFilter{
		Search:     query.Get("search"),
		Statuses:   statuses,
		CustomerID: query.Get("customer_id"),
	}
}

func customerReceivableSourceDocumentInput(
	input customerReceivableSourceDocumentRequest,
) financeapp.SourceDocumentInput {
	return financeapp.SourceDocumentInput{
		Type: input.Type,
		ID:   input.ID,
		No:   input.No,
	}
}

func customerReceivableLineInputs(inputs []customerReceivableLineRequest) []financeapp.CustomerReceivableLineInput {
	lines := make([]financeapp.CustomerReceivableLineInput, 0, len(inputs))
	for _, input := range inputs {
		lines = append(lines, financeapp.CustomerReceivableLineInput{
			ID:             input.ID,
			Description:    input.Description,
			SourceDocument: customerReceivableSourceDocumentInput(input.SourceDocument),
			Amount:         input.Amount,
		})
	}

	return lines
}

func newCustomerReceivableListItemResponse(
	receivable financedomain.CustomerReceivable,
) customerReceivableListItemResponse {
	return customerReceivableListItemResponse{
		ID:                receivable.ID,
		ReceivableNo:      receivable.ReceivableNo,
		CustomerID:        receivable.CustomerID,
		CustomerCode:      receivable.CustomerCode,
		CustomerName:      receivable.CustomerName,
		Status:            string(receivable.Status),
		TotalAmount:       receivable.TotalAmount.String(),
		PaidAmount:        receivable.PaidAmount.String(),
		OutstandingAmount: receivable.OutstandingAmount.String(),
		CurrencyCode:      receivable.CurrencyCode.String(),
		DueDate:           dateString(receivable.DueDate),
		CreatedAt:         timeString(receivable.CreatedAt),
		UpdatedAt:         timeString(receivable.UpdatedAt),
		Version:           receivable.Version,
	}
}

func newCustomerReceivableResponse(
	receivable financedomain.CustomerReceivable,
	auditLogID string,
) customerReceivableResponse {
	payload := customerReceivableResponse{
		ID:                receivable.ID,
		OrgID:             receivable.OrgID,
		ReceivableNo:      receivable.ReceivableNo,
		CustomerID:        receivable.CustomerID,
		CustomerCode:      receivable.CustomerCode,
		CustomerName:      receivable.CustomerName,
		Status:            string(receivable.Status),
		SourceDocument:    newCustomerReceivableSourceDocumentResponse(receivable.SourceDocument),
		Lines:             make([]customerReceivableLineResponse, 0, len(receivable.Lines)),
		TotalAmount:       receivable.TotalAmount.String(),
		PaidAmount:        receivable.PaidAmount.String(),
		OutstandingAmount: receivable.OutstandingAmount.String(),
		CurrencyCode:      receivable.CurrencyCode.String(),
		DueDate:           dateString(receivable.DueDate),
		DisputeReason:     receivable.DisputeReason,
		VoidReason:        receivable.VoidReason,
		AuditLogID:        auditLogID,
		CreatedAt:         timeString(receivable.CreatedAt),
		UpdatedAt:         timeString(receivable.UpdatedAt),
		Version:           receivable.Version,
	}
	for _, line := range receivable.Lines {
		payload.Lines = append(payload.Lines, customerReceivableLineResponse{
			ID:             line.ID,
			Description:    line.Description,
			SourceDocument: newCustomerReceivableSourceDocumentResponse(line.SourceDocument),
			Amount:         line.Amount.String(),
		})
	}

	return payload
}

func newCustomerReceivableSourceDocumentResponse(
	source financedomain.SourceDocumentRef,
) customerReceivableSourceDocumentResponse {
	return customerReceivableSourceDocumentResponse{
		Type: string(source.Type),
		ID:   source.ID,
		No:   source.No,
	}
}

func writeCustomerReceivableError(w http.ResponseWriter, r *http.Request, err error) {
	if appErr, ok := apperrors.As(err); ok {
		response.WriteError(w, r, appErr.HTTPStatus, appErr.Code, appErr.Message, appErr.Details)
		return
	}

	response.WriteError(w, r, http.StatusConflict, response.ErrorCodeConflict, "Customer receivable request could not be processed", nil)
}
