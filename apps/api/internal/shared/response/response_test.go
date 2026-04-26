package response

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestWriteSuccessUsesEnvelope(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/api/v1/health", nil)
	req.Header.Set(HeaderRequestID, "req-test")
	rec := httptest.NewRecorder()

	WriteSuccess(rec, req, http.StatusOK, map[string]string{"status": "ok"})

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	if got := rec.Header().Get(HeaderRequestID); got != "req-test" {
		t.Fatalf("request header = %q, want req-test", got)
	}

	var payload SuccessEnvelope[map[string]string]
	if err := json.NewDecoder(rec.Body).Decode(&payload); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if !payload.Success {
		t.Fatal("success = false, want true")
	}
	if payload.RequestID != "req-test" {
		t.Fatalf("request id = %q, want req-test", payload.RequestID)
	}
	if payload.Data["status"] != "ok" {
		t.Fatalf("data.status = %q, want ok", payload.Data["status"])
	}
}

func TestWriteErrorUsesStandardEnvelope(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/api/v1/inventory/stock-movements", nil)
	req.Header.Set(HeaderRequestID, "req-error")
	rec := httptest.NewRecorder()

	WriteError(rec, req, http.StatusBadRequest, ErrorCodeValidation, "Invalid request", map[string]any{
		"field": "quantity",
	})

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusBadRequest)
	}

	var payload ErrorEnvelope
	if err := json.NewDecoder(rec.Body).Decode(&payload); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if payload.Error.Code != ErrorCodeValidation {
		t.Fatalf("code = %q, want %q", payload.Error.Code, ErrorCodeValidation)
	}
	if payload.Error.Message != "Invalid request" {
		t.Fatalf("message = %q, want Invalid request", payload.Error.Message)
	}
	if payload.Error.RequestID != "req-error" {
		t.Fatalf("request id = %q, want req-error", payload.Error.RequestID)
	}
	if payload.Error.Details["field"] != "quantity" {
		t.Fatalf("details.field = %v, want quantity", payload.Error.Details["field"])
	}
}
