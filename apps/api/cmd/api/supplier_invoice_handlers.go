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

type supplierInvoiceLineRequest struct {
	ID             string                               `json:"id"`
	Description    string                               `json:"description"`
	SourceDocument supplierPayableSourceDocumentRequest `json:"source_document"`
	Amount         string                               `json:"amount"`
}

type createSupplierInvoiceRequest struct {
	ID            string                       `json:"id"`
	InvoiceNo     string                       `json:"invoice_no"`
	SupplierID    string                       `json:"supplier_id"`
	SupplierCode  string                       `json:"supplier_code"`
	SupplierName  string                       `json:"supplier_name"`
	PayableID     string                       `json:"payable_id"`
	InvoiceDate   string                       `json:"invoice_date"`
	InvoiceAmount string                       `json:"invoice_amount"`
	CurrencyCode  string                       `json:"currency_code"`
	Lines         []supplierInvoiceLineRequest `json:"lines"`
}

type supplierInvoiceActionRequest struct {
	Reason string `json:"reason"`
}

type supplierInvoiceLineResponse struct {
	ID             string                                `json:"id"`
	Description    string                                `json:"description"`
	SourceDocument supplierPayableSourceDocumentResponse `json:"source_document"`
	Amount         string                                `json:"amount"`
}

type supplierInvoiceListItemResponse struct {
	ID             string                                `json:"id"`
	InvoiceNo      string                                `json:"invoice_no"`
	SupplierID     string                                `json:"supplier_id"`
	SupplierCode   string                                `json:"supplier_code,omitempty"`
	SupplierName   string                                `json:"supplier_name"`
	PayableID      string                                `json:"payable_id"`
	PayableNo      string                                `json:"payable_no"`
	Status         string                                `json:"status"`
	MatchStatus    string                                `json:"match_status"`
	SourceDocument supplierPayableSourceDocumentResponse `json:"source_document"`
	InvoiceAmount  string                                `json:"invoice_amount"`
	ExpectedAmount string                                `json:"expected_amount"`
	VarianceAmount string                                `json:"variance_amount"`
	CurrencyCode   string                                `json:"currency_code"`
	InvoiceDate    string                                `json:"invoice_date"`
	CreatedAt      string                                `json:"created_at"`
	UpdatedAt      string                                `json:"updated_at"`
	Version        int                                   `json:"version"`
}

type supplierInvoiceResponse struct {
	ID             string                                `json:"id"`
	OrgID          string                                `json:"org_id"`
	InvoiceNo      string                                `json:"invoice_no"`
	SupplierID     string                                `json:"supplier_id"`
	SupplierCode   string                                `json:"supplier_code,omitempty"`
	SupplierName   string                                `json:"supplier_name"`
	PayableID      string                                `json:"payable_id"`
	PayableNo      string                                `json:"payable_no"`
	Status         string                                `json:"status"`
	MatchStatus    string                                `json:"match_status"`
	SourceDocument supplierPayableSourceDocumentResponse `json:"source_document"`
	Lines          []supplierInvoiceLineResponse         `json:"lines"`
	InvoiceAmount  string                                `json:"invoice_amount"`
	ExpectedAmount string                                `json:"expected_amount"`
	VarianceAmount string                                `json:"variance_amount"`
	CurrencyCode   string                                `json:"currency_code"`
	InvoiceDate    string                                `json:"invoice_date"`
	VoidReason     string                                `json:"void_reason,omitempty"`
	AuditLogID     string                                `json:"audit_log_id,omitempty"`
	CreatedAt      string                                `json:"created_at"`
	UpdatedAt      string                                `json:"updated_at"`
	Version        int                                   `json:"version"`
}

type supplierInvoiceActionResultResponse struct {
	SupplierInvoice supplierInvoiceResponse `json:"supplier_invoice"`
	PreviousStatus  string                  `json:"previous_status"`
	CurrentStatus   string                  `json:"current_status"`
	AuditLogID      string                  `json:"audit_log_id,omitempty"`
}

func supplierInvoicesHandler(service financeapp.SupplierInvoiceService) http.HandlerFunc {
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
			invoices, err := service.ListSupplierInvoices(r.Context(), supplierInvoiceFilterFromRequest(r))
			if err != nil {
				writeSupplierInvoiceError(w, r, err)
				return
			}
			payload := make([]supplierInvoiceListItemResponse, 0, len(invoices))
			for _, invoice := range invoices {
				payload = append(payload, newSupplierInvoiceListItemResponse(invoice))
			}
			response.WriteSuccess(w, r, http.StatusOK, payload)
		case http.MethodPost:
			if !auth.HasPermission(principal, auth.PermissionFinanceManage) {
				writePermissionDenied(w, r, auth.PermissionFinanceManage)
				return
			}
			r = requestWithStableID(r)
			var payload createSupplierInvoiceRequest
			if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
				response.WriteError(w, r, http.StatusBadRequest, response.ErrorCodeValidation, "Invalid supplier invoice payload", nil)
				return
			}

			result, err := service.CreateSupplierInvoice(r.Context(), financeapp.CreateSupplierInvoiceInput{
				ID:            payload.ID,
				InvoiceNo:     payload.InvoiceNo,
				SupplierID:    payload.SupplierID,
				SupplierCode:  payload.SupplierCode,
				SupplierName:  payload.SupplierName,
				PayableID:     payload.PayableID,
				InvoiceDate:   payload.InvoiceDate,
				InvoiceAmount: payload.InvoiceAmount,
				CurrencyCode:  payload.CurrencyCode,
				Lines:         supplierInvoiceLineInputs(payload.Lines),
				ActorID:       principal.UserID,
				RequestID:     response.RequestID(r),
			})
			if err != nil {
				writeSupplierInvoiceError(w, r, err)
				return
			}

			response.WriteSuccess(w, r, http.StatusCreated, newSupplierInvoiceResponse(result.SupplierInvoice, result.AuditLogID))
		default:
			response.WriteError(w, r, http.StatusMethodNotAllowed, response.ErrorCodeNotFound, "Route not found", nil)
		}
	}
}

func supplierInvoiceDetailHandler(service financeapp.SupplierInvoiceService) http.HandlerFunc {
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

		invoice, err := service.GetSupplierInvoice(r.Context(), r.PathValue("supplier_invoice_id"))
		if err != nil {
			writeSupplierInvoiceError(w, r, err)
			return
		}
		response.WriteSuccess(w, r, http.StatusOK, newSupplierInvoiceResponse(invoice, ""))
	}
}

func supplierInvoiceVoidHandler(service financeapp.SupplierInvoiceService) http.HandlerFunc {
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
		if !auth.HasPermission(principal, auth.PermissionFinanceManage) {
			writePermissionDenied(w, r, auth.PermissionFinanceManage)
			return
		}

		r = requestWithStableID(r)
		var payload supplierInvoiceActionRequest
		if r.Body != nil {
			if err := json.NewDecoder(r.Body).Decode(&payload); err != nil && !errors.Is(err, io.EOF) {
				response.WriteError(w, r, http.StatusBadRequest, response.ErrorCodeValidation, "Invalid supplier invoice action payload", nil)
				return
			}
		}

		result, err := service.VoidSupplierInvoice(r.Context(), financeapp.SupplierInvoiceActionInput{
			ID:        r.PathValue("supplier_invoice_id"),
			Reason:    payload.Reason,
			ActorID:   principal.UserID,
			RequestID: response.RequestID(r),
		})
		if err != nil {
			writeSupplierInvoiceError(w, r, err)
			return
		}

		response.WriteSuccess(w, r, http.StatusOK, supplierInvoiceActionResultResponse{
			SupplierInvoice: newSupplierInvoiceResponse(result.SupplierInvoice, result.AuditLogID),
			PreviousStatus:  string(result.PreviousStatus),
			CurrentStatus:   string(result.CurrentStatus),
			AuditLogID:      result.AuditLogID,
		})
	}
}

func supplierInvoiceFilterFromRequest(r *http.Request) financeapp.SupplierInvoiceFilter {
	query := r.URL.Query()
	search := strings.TrimSpace(query.Get("search"))
	if search == "" {
		search = strings.TrimSpace(query.Get("q"))
	}
	statuses := []financedomain.SupplierInvoiceStatus{}
	for _, rawStatus := range strings.Split(query.Get("status"), ",") {
		status := financedomain.NormalizeSupplierInvoiceStatus(financedomain.SupplierInvoiceStatus(rawStatus))
		if status != "" {
			statuses = append(statuses, status)
		}
	}

	return financeapp.SupplierInvoiceFilter{
		Search:     search,
		Statuses:   statuses,
		SupplierID: strings.TrimSpace(query.Get("supplier_id")),
		PayableID:  strings.TrimSpace(query.Get("payable_id")),
	}
}

func supplierInvoiceLineInputs(inputs []supplierInvoiceLineRequest) []financeapp.SupplierInvoiceLineInput {
	lines := make([]financeapp.SupplierInvoiceLineInput, 0, len(inputs))
	for _, input := range inputs {
		lines = append(lines, financeapp.SupplierInvoiceLineInput{
			ID:             input.ID,
			Description:    input.Description,
			SourceDocument: supplierPayableSourceDocumentInput(input.SourceDocument),
			Amount:         input.Amount,
		})
	}

	return lines
}

func newSupplierInvoiceListItemResponse(invoice financedomain.SupplierInvoice) supplierInvoiceListItemResponse {
	return supplierInvoiceListItemResponse{
		ID:             invoice.ID,
		InvoiceNo:      invoice.InvoiceNo,
		SupplierID:     invoice.SupplierID,
		SupplierCode:   invoice.SupplierCode,
		SupplierName:   invoice.SupplierName,
		PayableID:      invoice.PayableID,
		PayableNo:      invoice.PayableNo,
		Status:         string(invoice.Status),
		MatchStatus:    string(invoice.MatchStatus),
		SourceDocument: newSupplierPayableSourceDocumentResponse(invoice.SourceDocument),
		InvoiceAmount:  invoice.InvoiceAmount.String(),
		ExpectedAmount: invoice.ExpectedAmount.String(),
		VarianceAmount: invoice.VarianceAmount.String(),
		CurrencyCode:   invoice.CurrencyCode.String(),
		InvoiceDate:    dateString(invoice.InvoiceDate),
		CreatedAt:      timeString(invoice.CreatedAt),
		UpdatedAt:      timeString(invoice.UpdatedAt),
		Version:        invoice.Version,
	}
}

func newSupplierInvoiceResponse(invoice financedomain.SupplierInvoice, auditLogID string) supplierInvoiceResponse {
	lines := make([]supplierInvoiceLineResponse, 0, len(invoice.Lines))
	for _, line := range invoice.Lines {
		lines = append(lines, supplierInvoiceLineResponse{
			ID:             line.ID,
			Description:    line.Description,
			SourceDocument: newSupplierPayableSourceDocumentResponse(line.SourceDocument),
			Amount:         line.Amount.String(),
		})
	}

	return supplierInvoiceResponse{
		ID:             invoice.ID,
		OrgID:          invoice.OrgID,
		InvoiceNo:      invoice.InvoiceNo,
		SupplierID:     invoice.SupplierID,
		SupplierCode:   invoice.SupplierCode,
		SupplierName:   invoice.SupplierName,
		PayableID:      invoice.PayableID,
		PayableNo:      invoice.PayableNo,
		Status:         string(invoice.Status),
		MatchStatus:    string(invoice.MatchStatus),
		SourceDocument: newSupplierPayableSourceDocumentResponse(invoice.SourceDocument),
		Lines:          lines,
		InvoiceAmount:  invoice.InvoiceAmount.String(),
		ExpectedAmount: invoice.ExpectedAmount.String(),
		VarianceAmount: invoice.VarianceAmount.String(),
		CurrencyCode:   invoice.CurrencyCode.String(),
		InvoiceDate:    dateString(invoice.InvoiceDate),
		VoidReason:     invoice.VoidReason,
		AuditLogID:     auditLogID,
		CreatedAt:      timeString(invoice.CreatedAt),
		UpdatedAt:      timeString(invoice.UpdatedAt),
		Version:        invoice.Version,
	}
}

func writeSupplierInvoiceError(w http.ResponseWriter, r *http.Request, err error) {
	if appErr, ok := apperrors.As(err); ok {
		response.WriteError(w, r, appErr.HTTPStatus, appErr.Code, appErr.Message, appErr.Details)
		return
	}

	response.WriteError(w, r, http.StatusConflict, response.ErrorCodeConflict, "Supplier invoice request could not be processed", nil)
}
