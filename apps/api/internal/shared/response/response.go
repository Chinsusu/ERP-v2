package response

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"
)

const HeaderRequestID = "X-Request-ID"

type ErrorCode string

const (
	ErrorCodeValidation        ErrorCode = "VALIDATION_ERROR"
	ErrorCodeUnauthorized      ErrorCode = "UNAUTHORIZED"
	ErrorCodeForbidden         ErrorCode = "FORBIDDEN"
	ErrorCodeNotFound          ErrorCode = "NOT_FOUND"
	ErrorCodeConflict          ErrorCode = "CONFLICT"
	ErrorCodeInsufficientStock ErrorCode = "INSUFFICIENT_STOCK"
	ErrorCodeInvalidState      ErrorCode = "INVALID_STATE"
)

type SuccessEnvelope[T any] struct {
	Success   bool   `json:"success"`
	Data      T      `json:"data"`
	RequestID string `json:"request_id"`
}

type Pagination struct {
	Page       int `json:"page"`
	PageSize   int `json:"page_size"`
	TotalItems int `json:"total_items"`
	TotalPages int `json:"total_pages"`
}

type PaginatedSuccessEnvelope[T any] struct {
	Success    bool       `json:"success"`
	Data       T          `json:"data"`
	Pagination Pagination `json:"pagination"`
	RequestID  string     `json:"request_id"`
}

type ErrorEnvelope struct {
	Error APIError `json:"error"`
}

type APIError struct {
	Code      ErrorCode      `json:"code"`
	Message   string         `json:"message"`
	Details   map[string]any `json:"details,omitempty"`
	RequestID string         `json:"request_id"`
}

func WriteSuccess[T any](w http.ResponseWriter, r *http.Request, status int, data T) {
	requestID := RequestID(r)
	writeJSON(w, status, requestID, SuccessEnvelope[T]{
		Success:   true,
		Data:      data,
		RequestID: requestID,
	})
}

func WritePaginatedSuccess[T any](w http.ResponseWriter, r *http.Request, status int, data T, pagination Pagination) {
	requestID := RequestID(r)
	writeJSON(w, status, requestID, PaginatedSuccessEnvelope[T]{
		Success:    true,
		Data:       data,
		Pagination: pagination,
		RequestID:  requestID,
	})
}

func WriteError(w http.ResponseWriter, r *http.Request, status int, code ErrorCode, message string, details map[string]any) {
	requestID := RequestID(r)
	writeJSON(w, status, requestID, ErrorEnvelope{
		Error: APIError{
			Code:      code,
			Message:   message,
			Details:   details,
			RequestID: requestID,
		},
	})
}

func RequestID(r *http.Request) string {
	if r == nil {
		return newRequestID()
	}

	requestID := strings.TrimSpace(r.Header.Get(HeaderRequestID))
	if requestID != "" {
		return requestID
	}

	return newRequestID()
}

func writeJSON(w http.ResponseWriter, status int, requestID string, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set(HeaderRequestID, requestID)
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

func newRequestID() string {
	return "req_" + strconv.FormatInt(time.Now().UTC().UnixNano(), 36)
}
