package main

import (
	"encoding/json"
	"net/http"
	"strings"

	financeapp "github.com/Chinsusu/ERP-v2/apps/api/internal/modules/finance/application"
	financedomain "github.com/Chinsusu/ERP-v2/apps/api/internal/modules/finance/domain"
	"github.com/Chinsusu/ERP-v2/apps/api/internal/shared/auth"
	apperrors "github.com/Chinsusu/ERP-v2/apps/api/internal/shared/errors"
	"github.com/Chinsusu/ERP-v2/apps/api/internal/shared/response"
)

type cashTransactionAllocationRequest struct {
	ID         string `json:"id"`
	TargetType string `json:"target_type"`
	TargetID   string `json:"target_id"`
	TargetNo   string `json:"target_no"`
	Amount     string `json:"amount"`
}

type createCashTransactionRequest struct {
	ID               string                             `json:"id"`
	TransactionNo    string                             `json:"transaction_no"`
	Direction        string                             `json:"direction"`
	BusinessDate     string                             `json:"business_date"`
	CounterpartyID   string                             `json:"counterparty_id"`
	CounterpartyName string                             `json:"counterparty_name"`
	PaymentMethod    string                             `json:"payment_method"`
	ReferenceNo      string                             `json:"reference_no"`
	Allocations      []cashTransactionAllocationRequest `json:"allocations"`
	TotalAmount      string                             `json:"total_amount"`
	CurrencyCode     string                             `json:"currency_code"`
	Memo             string                             `json:"memo"`
}

type cashTransactionAllocationResponse struct {
	ID         string `json:"id"`
	TargetType string `json:"target_type"`
	TargetID   string `json:"target_id"`
	TargetNo   string `json:"target_no"`
	Amount     string `json:"amount"`
}

type cashTransactionListItemResponse struct {
	ID               string `json:"id"`
	TransactionNo    string `json:"transaction_no"`
	Direction        string `json:"direction"`
	Status           string `json:"status"`
	BusinessDate     string `json:"business_date"`
	CounterpartyID   string `json:"counterparty_id,omitempty"`
	CounterpartyName string `json:"counterparty_name"`
	PaymentMethod    string `json:"payment_method"`
	ReferenceNo      string `json:"reference_no,omitempty"`
	TotalAmount      string `json:"total_amount"`
	CurrencyCode     string `json:"currency_code"`
	PostedBy         string `json:"posted_by,omitempty"`
	PostedAt         string `json:"posted_at,omitempty"`
	CreatedAt        string `json:"created_at"`
	UpdatedAt        string `json:"updated_at"`
	Version          int    `json:"version"`
}

type cashTransactionResponse struct {
	ID               string                              `json:"id"`
	OrgID            string                              `json:"org_id"`
	TransactionNo    string                              `json:"transaction_no"`
	Direction        string                              `json:"direction"`
	Status           string                              `json:"status"`
	BusinessDate     string                              `json:"business_date"`
	CounterpartyID   string                              `json:"counterparty_id,omitempty"`
	CounterpartyName string                              `json:"counterparty_name"`
	PaymentMethod    string                              `json:"payment_method"`
	ReferenceNo      string                              `json:"reference_no,omitempty"`
	Allocations      []cashTransactionAllocationResponse `json:"allocations"`
	TotalAmount      string                              `json:"total_amount"`
	CurrencyCode     string                              `json:"currency_code"`
	Memo             string                              `json:"memo,omitempty"`
	PostedBy         string                              `json:"posted_by,omitempty"`
	PostedAt         string                              `json:"posted_at,omitempty"`
	VoidReason       string                              `json:"void_reason,omitempty"`
	VoidedBy         string                              `json:"voided_by,omitempty"`
	VoidedAt         string                              `json:"voided_at,omitempty"`
	AuditLogID       string                              `json:"audit_log_id,omitempty"`
	CreatedAt        string                              `json:"created_at"`
	UpdatedAt        string                              `json:"updated_at"`
	Version          int                                 `json:"version"`
}

func cashTransactionsHandler(service financeapp.CashTransactionService) http.HandlerFunc {
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
			transactions, err := service.ListCashTransactions(r.Context(), cashTransactionFilterFromRequest(r))
			if err != nil {
				writeCashTransactionError(w, r, err)
				return
			}
			payload := make([]cashTransactionListItemResponse, 0, len(transactions))
			for _, transaction := range transactions {
				payload = append(payload, newCashTransactionListItemResponse(transaction))
			}
			response.WriteSuccess(w, r, http.StatusOK, payload)
		case http.MethodPost:
			if !auth.HasPermission(principal, auth.PermissionFinanceManage) {
				writePermissionDenied(w, r, auth.PermissionFinanceManage)
				return
			}
			r = requestWithStableID(r)
			var payload createCashTransactionRequest
			if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
				response.WriteError(w, r, http.StatusBadRequest, response.ErrorCodeValidation, "Invalid cash transaction payload", nil)
				return
			}

			result, err := service.CreateCashTransaction(r.Context(), financeapp.CreateCashTransactionInput{
				ID:               payload.ID,
				TransactionNo:    payload.TransactionNo,
				Direction:        payload.Direction,
				BusinessDate:     payload.BusinessDate,
				CounterpartyID:   payload.CounterpartyID,
				CounterpartyName: payload.CounterpartyName,
				PaymentMethod:    payload.PaymentMethod,
				ReferenceNo:      payload.ReferenceNo,
				Allocations:      cashTransactionAllocationInputs(payload.Allocations),
				TotalAmount:      payload.TotalAmount,
				CurrencyCode:     payload.CurrencyCode,
				Memo:             payload.Memo,
				ActorID:          principal.UserID,
				RequestID:        response.RequestID(r),
			})
			if err != nil {
				writeCashTransactionError(w, r, err)
				return
			}

			response.WriteSuccess(w, r, http.StatusCreated, newCashTransactionResponse(result.CashTransaction, result.AuditLogID))
		default:
			response.WriteError(w, r, http.StatusMethodNotAllowed, response.ErrorCodeNotFound, "Route not found", nil)
		}
	}
}

func cashTransactionDetailHandler(service financeapp.CashTransactionService) http.HandlerFunc {
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

		transaction, err := service.GetCashTransaction(r.Context(), r.PathValue("cash_transaction_id"))
		if err != nil {
			writeCashTransactionError(w, r, err)
			return
		}
		response.WriteSuccess(w, r, http.StatusOK, newCashTransactionResponse(transaction, ""))
	}
}

func cashTransactionFilterFromRequest(r *http.Request) financeapp.CashTransactionFilter {
	query := r.URL.Query()
	search := strings.TrimSpace(query.Get("search"))
	if search == "" {
		search = strings.TrimSpace(query.Get("q"))
	}
	directions := []financedomain.CashTransactionDirection{}
	for _, rawDirection := range strings.Split(query.Get("direction"), ",") {
		direction := financedomain.NormalizeCashTransactionDirection(financedomain.CashTransactionDirection(rawDirection))
		if direction != "" {
			directions = append(directions, direction)
		}
	}
	statuses := []financedomain.CashTransactionStatus{}
	for _, rawStatus := range strings.Split(query.Get("status"), ",") {
		status := financedomain.NormalizeCashTransactionStatus(financedomain.CashTransactionStatus(rawStatus))
		if status != "" {
			statuses = append(statuses, status)
		}
	}

	return financeapp.CashTransactionFilter{
		Search:         search,
		Directions:     directions,
		Statuses:       statuses,
		CounterpartyID: query.Get("counterparty_id"),
	}
}

func cashTransactionAllocationInputs(
	inputs []cashTransactionAllocationRequest,
) []financeapp.CashTransactionAllocationInput {
	allocations := make([]financeapp.CashTransactionAllocationInput, 0, len(inputs))
	for _, input := range inputs {
		allocations = append(allocations, financeapp.CashTransactionAllocationInput{
			ID:         input.ID,
			TargetType: input.TargetType,
			TargetID:   input.TargetID,
			TargetNo:   input.TargetNo,
			Amount:     input.Amount,
		})
	}

	return allocations
}

func newCashTransactionListItemResponse(
	transaction financedomain.CashTransaction,
) cashTransactionListItemResponse {
	return cashTransactionListItemResponse{
		ID:               transaction.ID,
		TransactionNo:    transaction.TransactionNo,
		Direction:        string(transaction.Direction),
		Status:           string(transaction.Status),
		BusinessDate:     dateString(transaction.BusinessDate),
		CounterpartyID:   transaction.CounterpartyID,
		CounterpartyName: transaction.CounterpartyName,
		PaymentMethod:    transaction.PaymentMethod,
		ReferenceNo:      transaction.ReferenceNo,
		TotalAmount:      transaction.TotalAmount.String(),
		CurrencyCode:     transaction.CurrencyCode.String(),
		PostedBy:         transaction.PostedBy,
		PostedAt:         timeString(transaction.PostedAt),
		CreatedAt:        timeString(transaction.CreatedAt),
		UpdatedAt:        timeString(transaction.UpdatedAt),
		Version:          transaction.Version,
	}
}

func newCashTransactionResponse(
	transaction financedomain.CashTransaction,
	auditLogID string,
) cashTransactionResponse {
	payload := cashTransactionResponse{
		ID:               transaction.ID,
		OrgID:            transaction.OrgID,
		TransactionNo:    transaction.TransactionNo,
		Direction:        string(transaction.Direction),
		Status:           string(transaction.Status),
		BusinessDate:     dateString(transaction.BusinessDate),
		CounterpartyID:   transaction.CounterpartyID,
		CounterpartyName: transaction.CounterpartyName,
		PaymentMethod:    transaction.PaymentMethod,
		ReferenceNo:      transaction.ReferenceNo,
		Allocations:      make([]cashTransactionAllocationResponse, 0, len(transaction.Allocations)),
		TotalAmount:      transaction.TotalAmount.String(),
		CurrencyCode:     transaction.CurrencyCode.String(),
		Memo:             transaction.Memo,
		PostedBy:         transaction.PostedBy,
		PostedAt:         timeString(transaction.PostedAt),
		VoidReason:       transaction.VoidReason,
		VoidedBy:         transaction.VoidedBy,
		VoidedAt:         timeString(transaction.VoidedAt),
		AuditLogID:       auditLogID,
		CreatedAt:        timeString(transaction.CreatedAt),
		UpdatedAt:        timeString(transaction.UpdatedAt),
		Version:          transaction.Version,
	}
	for _, allocation := range transaction.Allocations {
		payload.Allocations = append(payload.Allocations, cashTransactionAllocationResponse{
			ID:         allocation.ID,
			TargetType: string(allocation.TargetType),
			TargetID:   allocation.TargetID,
			TargetNo:   allocation.TargetNo,
			Amount:     allocation.Amount.String(),
		})
	}

	return payload
}

func writeCashTransactionError(w http.ResponseWriter, r *http.Request, err error) {
	if appErr, ok := apperrors.As(err); ok {
		response.WriteError(w, r, appErr.HTTPStatus, appErr.Code, appErr.Message, appErr.Details)
		return
	}

	response.WriteError(w, r, http.StatusConflict, response.ErrorCodeConflict, "Cash transaction request could not be processed", nil)
}
