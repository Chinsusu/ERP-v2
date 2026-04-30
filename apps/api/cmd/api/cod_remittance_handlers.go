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

type codRemittanceLineRequest struct {
	ID             string `json:"id"`
	ReceivableID   string `json:"receivable_id"`
	ReceivableNo   string `json:"receivable_no"`
	ShipmentID     string `json:"shipment_id"`
	TrackingNo     string `json:"tracking_no"`
	CustomerName   string `json:"customer_name"`
	ExpectedAmount string `json:"expected_amount"`
	RemittedAmount string `json:"remitted_amount"`
}

type createCODRemittanceRequest struct {
	ID             string                     `json:"id"`
	RemittanceNo   string                     `json:"remittance_no"`
	CarrierID      string                     `json:"carrier_id"`
	CarrierCode    string                     `json:"carrier_code"`
	CarrierName    string                     `json:"carrier_name"`
	BusinessDate   string                     `json:"business_date"`
	ExpectedAmount string                     `json:"expected_amount"`
	RemittedAmount string                     `json:"remitted_amount"`
	CurrencyCode   string                     `json:"currency_code"`
	Lines          []codRemittanceLineRequest `json:"lines"`
}

type codRemittanceDiscrepancyRequest struct {
	ID      string `json:"id"`
	LineID  string `json:"line_id"`
	Type    string `json:"type"`
	Status  string `json:"status"`
	Reason  string `json:"reason"`
	OwnerID string `json:"owner_id"`
}

type codRemittanceListItemResponse struct {
	ID                string `json:"id"`
	RemittanceNo      string `json:"remittance_no"`
	CarrierID         string `json:"carrier_id"`
	CarrierCode       string `json:"carrier_code,omitempty"`
	CarrierName       string `json:"carrier_name"`
	Status            string `json:"status"`
	BusinessDate      string `json:"business_date"`
	ExpectedAmount    string `json:"expected_amount"`
	RemittedAmount    string `json:"remitted_amount"`
	DiscrepancyAmount string `json:"discrepancy_amount"`
	CurrencyCode      string `json:"currency_code"`
	LineCount         int    `json:"line_count"`
	DiscrepancyCount  int    `json:"discrepancy_count"`
	CreatedAt         string `json:"created_at"`
	UpdatedAt         string `json:"updated_at"`
	Version           int    `json:"version"`
}

type codRemittanceLineResponse struct {
	ID                string `json:"id"`
	ReceivableID      string `json:"receivable_id"`
	ReceivableNo      string `json:"receivable_no"`
	ShipmentID        string `json:"shipment_id,omitempty"`
	TrackingNo        string `json:"tracking_no"`
	CustomerName      string `json:"customer_name,omitempty"`
	ExpectedAmount    string `json:"expected_amount"`
	RemittedAmount    string `json:"remitted_amount"`
	DiscrepancyAmount string `json:"discrepancy_amount"`
	MatchStatus       string `json:"match_status"`
}

type codDiscrepancyResponse struct {
	ID           string `json:"id"`
	LineID       string `json:"line_id"`
	ReceivableID string `json:"receivable_id"`
	Type         string `json:"type"`
	Status       string `json:"status"`
	Amount       string `json:"amount"`
	Reason       string `json:"reason"`
	OwnerID      string `json:"owner_id"`
	RecordedBy   string `json:"recorded_by"`
	RecordedAt   string `json:"recorded_at"`
	ResolvedBy   string `json:"resolved_by,omitempty"`
	ResolvedAt   string `json:"resolved_at,omitempty"`
	Resolution   string `json:"resolution,omitempty"`
}

type codRemittanceResponse struct {
	ID                string                      `json:"id"`
	OrgID             string                      `json:"org_id"`
	RemittanceNo      string                      `json:"remittance_no"`
	CarrierID         string                      `json:"carrier_id"`
	CarrierCode       string                      `json:"carrier_code,omitempty"`
	CarrierName       string                      `json:"carrier_name"`
	Status            string                      `json:"status"`
	BusinessDate      string                      `json:"business_date"`
	ExpectedAmount    string                      `json:"expected_amount"`
	RemittedAmount    string                      `json:"remitted_amount"`
	DiscrepancyAmount string                      `json:"discrepancy_amount"`
	CurrencyCode      string                      `json:"currency_code"`
	Lines             []codRemittanceLineResponse `json:"lines"`
	Discrepancies     []codDiscrepancyResponse    `json:"discrepancies"`
	AuditLogID        string                      `json:"audit_log_id,omitempty"`
	SubmittedBy       string                      `json:"submitted_by,omitempty"`
	SubmittedAt       string                      `json:"submitted_at,omitempty"`
	ApprovedBy        string                      `json:"approved_by,omitempty"`
	ApprovedAt        string                      `json:"approved_at,omitempty"`
	ClosedBy          string                      `json:"closed_by,omitempty"`
	ClosedAt          string                      `json:"closed_at,omitempty"`
	CreatedAt         string                      `json:"created_at"`
	UpdatedAt         string                      `json:"updated_at"`
	Version           int                         `json:"version"`
}

type codRemittanceActionResultResponse struct {
	CODRemittance  codRemittanceResponse `json:"cod_remittance"`
	PreviousStatus string                `json:"previous_status"`
	CurrentStatus  string                `json:"current_status"`
	AuditLogID     string                `json:"audit_log_id,omitempty"`
}

func codRemittancesHandler(service financeapp.CODRemittanceService) http.HandlerFunc {
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
			remittances, err := service.ListCODRemittances(r.Context(), codRemittanceFilterFromRequest(r))
			if err != nil {
				writeCODRemittanceError(w, r, err)
				return
			}
			payload := make([]codRemittanceListItemResponse, 0, len(remittances))
			for _, remittance := range remittances {
				payload = append(payload, newCODRemittanceListItemResponse(remittance))
			}
			response.WriteSuccess(w, r, http.StatusOK, payload)
		case http.MethodPost:
			if !auth.HasPermission(principal, auth.PermissionCODReconcile) {
				writePermissionDenied(w, r, auth.PermissionCODReconcile)
				return
			}
			r = requestWithStableID(r)
			var payload createCODRemittanceRequest
			if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
				response.WriteError(w, r, http.StatusBadRequest, response.ErrorCodeValidation, "Invalid COD remittance payload", nil)
				return
			}
			result, err := service.CreateCODRemittance(r.Context(), financeapp.CreateCODRemittanceInput{
				ID:             payload.ID,
				RemittanceNo:   payload.RemittanceNo,
				CarrierID:      payload.CarrierID,
				CarrierCode:    payload.CarrierCode,
				CarrierName:    payload.CarrierName,
				BusinessDate:   payload.BusinessDate,
				ExpectedAmount: payload.ExpectedAmount,
				RemittedAmount: payload.RemittedAmount,
				CurrencyCode:   payload.CurrencyCode,
				Lines:          codRemittanceLineInputs(payload.Lines),
				ActorID:        principal.UserID,
				RequestID:      response.RequestID(r),
			})
			if err != nil {
				writeCODRemittanceError(w, r, err)
				return
			}
			response.WriteSuccess(w, r, http.StatusCreated, newCODRemittanceResponse(result.CODRemittance, result.AuditLogID))
		default:
			response.WriteError(w, r, http.StatusMethodNotAllowed, response.ErrorCodeNotFound, "Route not found", nil)
		}
	}
}

func codRemittanceDetailHandler(service financeapp.CODRemittanceService) http.HandlerFunc {
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
		remittance, err := service.GetCODRemittance(r.Context(), r.PathValue("cod_remittance_id"))
		if err != nil {
			writeCODRemittanceError(w, r, err)
			return
		}
		response.WriteSuccess(w, r, http.StatusOK, newCODRemittanceResponse(remittance, ""))
	}
}

func codRemittanceMatchHandler(service financeapp.CODRemittanceService) http.HandlerFunc {
	return codRemittanceActionHandler(service, "match")
}

func codRemittanceSubmitHandler(service financeapp.CODRemittanceService) http.HandlerFunc {
	return codRemittanceActionHandler(service, "submit")
}

func codRemittanceApproveHandler(service financeapp.CODRemittanceService) http.HandlerFunc {
	return codRemittanceActionHandler(service, "approve")
}

func codRemittanceCloseHandler(service financeapp.CODRemittanceService) http.HandlerFunc {
	return codRemittanceActionHandler(service, "close")
}

func codRemittanceActionHandler(service financeapp.CODRemittanceService, action string) http.HandlerFunc {
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
		if !auth.HasPermission(principal, auth.PermissionCODReconcile) {
			writePermissionDenied(w, r, auth.PermissionCODReconcile)
			return
		}
		if r.Body != nil {
			var discard map[string]any
			if err := json.NewDecoder(r.Body).Decode(&discard); err != nil && !errors.Is(err, io.EOF) {
				response.WriteError(w, r, http.StatusBadRequest, response.ErrorCodeValidation, "Invalid COD remittance action payload", nil)
				return
			}
		}
		r = requestWithStableID(r)
		input := financeapp.CODRemittanceActionInput{
			ID:        r.PathValue("cod_remittance_id"),
			ActorID:   principal.UserID,
			RequestID: response.RequestID(r),
		}
		var (
			result financeapp.CODRemittanceActionResult
			err    error
		)
		switch action {
		case "match":
			result, err = service.MarkCODRemittanceMatching(r.Context(), input)
		case "submit":
			result, err = service.SubmitCODRemittance(r.Context(), input)
		case "approve":
			result, err = service.ApproveCODRemittance(r.Context(), input)
		case "close":
			result, err = service.CloseCODRemittance(r.Context(), input)
		default:
			response.WriteError(w, r, http.StatusNotFound, response.ErrorCodeNotFound, "Route not found", nil)
			return
		}
		if err != nil {
			writeCODRemittanceError(w, r, err)
			return
		}
		response.WriteSuccess(w, r, http.StatusOK, newCODRemittanceActionResultResponse(result))
	}
}

func codRemittanceDiscrepancyHandler(service financeapp.CODRemittanceService) http.HandlerFunc {
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
		if !auth.HasPermission(principal, auth.PermissionCODReconcile) {
			writePermissionDenied(w, r, auth.PermissionCODReconcile)
			return
		}
		r = requestWithStableID(r)
		var payload codRemittanceDiscrepancyRequest
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			response.WriteError(w, r, http.StatusBadRequest, response.ErrorCodeValidation, "Invalid COD discrepancy payload", nil)
			return
		}
		result, err := service.RecordCODRemittanceDiscrepancy(r.Context(), financeapp.CODRemittanceDiscrepancyInput{
			RemittanceID:  r.PathValue("cod_remittance_id"),
			DiscrepancyID: payload.ID,
			LineID:        payload.LineID,
			Type:          payload.Type,
			Status:        payload.Status,
			Reason:        payload.Reason,
			OwnerID:       payload.OwnerID,
			ActorID:       principal.UserID,
			RequestID:     response.RequestID(r),
		})
		if err != nil {
			writeCODRemittanceError(w, r, err)
			return
		}
		response.WriteSuccess(w, r, http.StatusOK, newCODRemittanceActionResultResponse(result))
	}
}

func codRemittanceFilterFromRequest(r *http.Request) financeapp.CODRemittanceFilter {
	query := r.URL.Query()
	search := strings.TrimSpace(query.Get("search"))
	if search == "" {
		search = strings.TrimSpace(query.Get("q"))
	}
	statuses := []financedomain.CODRemittanceStatus{}
	for _, rawStatus := range strings.Split(query.Get("status"), ",") {
		status := financedomain.NormalizeCODRemittanceStatus(financedomain.CODRemittanceStatus(rawStatus))
		if status != "" {
			statuses = append(statuses, status)
		}
	}

	return financeapp.CODRemittanceFilter{
		Search:    search,
		Statuses:  statuses,
		CarrierID: query.Get("carrier_id"),
	}
}

func codRemittanceLineInputs(inputs []codRemittanceLineRequest) []financeapp.CODRemittanceLineInput {
	lines := make([]financeapp.CODRemittanceLineInput, 0, len(inputs))
	for _, input := range inputs {
		lines = append(lines, financeapp.CODRemittanceLineInput{
			ID:             input.ID,
			ReceivableID:   input.ReceivableID,
			ReceivableNo:   input.ReceivableNo,
			ShipmentID:     input.ShipmentID,
			TrackingNo:     input.TrackingNo,
			CustomerName:   input.CustomerName,
			ExpectedAmount: input.ExpectedAmount,
			RemittedAmount: input.RemittedAmount,
		})
	}

	return lines
}

func newCODRemittanceListItemResponse(remittance financedomain.CODRemittance) codRemittanceListItemResponse {
	return codRemittanceListItemResponse{
		ID:                remittance.ID,
		RemittanceNo:      remittance.RemittanceNo,
		CarrierID:         remittance.CarrierID,
		CarrierCode:       remittance.CarrierCode,
		CarrierName:       remittance.CarrierName,
		Status:            string(remittance.Status),
		BusinessDate:      dateString(remittance.BusinessDate),
		ExpectedAmount:    remittance.ExpectedAmount.String(),
		RemittedAmount:    remittance.RemittedAmount.String(),
		DiscrepancyAmount: remittance.DiscrepancyAmount.String(),
		CurrencyCode:      remittance.CurrencyCode.String(),
		LineCount:         len(remittance.Lines),
		DiscrepancyCount:  len(remittance.Discrepancies),
		CreatedAt:         timeString(remittance.CreatedAt),
		UpdatedAt:         timeString(remittance.UpdatedAt),
		Version:           remittance.Version,
	}
}

func newCODRemittanceResponse(remittance financedomain.CODRemittance, auditLogID string) codRemittanceResponse {
	payload := codRemittanceResponse{
		ID:                remittance.ID,
		OrgID:             remittance.OrgID,
		RemittanceNo:      remittance.RemittanceNo,
		CarrierID:         remittance.CarrierID,
		CarrierCode:       remittance.CarrierCode,
		CarrierName:       remittance.CarrierName,
		Status:            string(remittance.Status),
		BusinessDate:      dateString(remittance.BusinessDate),
		ExpectedAmount:    remittance.ExpectedAmount.String(),
		RemittedAmount:    remittance.RemittedAmount.String(),
		DiscrepancyAmount: remittance.DiscrepancyAmount.String(),
		CurrencyCode:      remittance.CurrencyCode.String(),
		Lines:             make([]codRemittanceLineResponse, 0, len(remittance.Lines)),
		Discrepancies:     make([]codDiscrepancyResponse, 0, len(remittance.Discrepancies)),
		AuditLogID:        auditLogID,
		SubmittedBy:       remittance.SubmittedBy,
		SubmittedAt:       timeString(remittance.SubmittedAt),
		ApprovedBy:        remittance.ApprovedBy,
		ApprovedAt:        timeString(remittance.ApprovedAt),
		ClosedBy:          remittance.ClosedBy,
		ClosedAt:          timeString(remittance.ClosedAt),
		CreatedAt:         timeString(remittance.CreatedAt),
		UpdatedAt:         timeString(remittance.UpdatedAt),
		Version:           remittance.Version,
	}
	for _, line := range remittance.Lines {
		payload.Lines = append(payload.Lines, codRemittanceLineResponse{
			ID:                line.ID,
			ReceivableID:      line.ReceivableID,
			ReceivableNo:      line.ReceivableNo,
			ShipmentID:        line.ShipmentID,
			TrackingNo:        line.TrackingNo,
			CustomerName:      line.CustomerName,
			ExpectedAmount:    line.ExpectedAmount.String(),
			RemittedAmount:    line.RemittedAmount.String(),
			DiscrepancyAmount: line.DiscrepancyAmount.String(),
			MatchStatus:       string(line.MatchStatus),
		})
	}
	for _, discrepancy := range remittance.Discrepancies {
		payload.Discrepancies = append(payload.Discrepancies, codDiscrepancyResponse{
			ID:           discrepancy.ID,
			LineID:       discrepancy.LineID,
			ReceivableID: discrepancy.ReceivableID,
			Type:         string(discrepancy.Type),
			Status:       string(discrepancy.Status),
			Amount:       discrepancy.Amount.String(),
			Reason:       discrepancy.Reason,
			OwnerID:      discrepancy.OwnerID,
			RecordedBy:   discrepancy.RecordedBy,
			RecordedAt:   timeString(discrepancy.RecordedAt),
			ResolvedBy:   discrepancy.ResolvedBy,
			ResolvedAt:   timeString(discrepancy.ResolvedAt),
			Resolution:   discrepancy.Resolution,
		})
	}

	return payload
}

func newCODRemittanceActionResultResponse(result financeapp.CODRemittanceActionResult) codRemittanceActionResultResponse {
	return codRemittanceActionResultResponse{
		CODRemittance:  newCODRemittanceResponse(result.CODRemittance, ""),
		PreviousStatus: string(result.PreviousStatus),
		CurrentStatus:  string(result.CurrentStatus),
		AuditLogID:     result.AuditLogID,
	}
}

func writeCODRemittanceError(w http.ResponseWriter, r *http.Request, err error) {
	if appErr, ok := apperrors.As(err); ok {
		response.WriteError(w, r, appErr.HTTPStatus, appErr.Code, appErr.Message, appErr.Details)
		return
	}

	response.WriteError(w, r, http.StatusConflict, response.ErrorCodeConflict, "COD remittance request could not be processed", nil)
}
