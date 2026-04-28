package errors

import (
	stderrors "errors"
	"net/http"
	"testing"

	"github.com/Chinsusu/ERP-v2/apps/api/internal/shared/response"
)

func TestAppErrorWrapsCauseAndKeepsStandardFields(t *testing.T) {
	cause := stderrors.New("sales order not found")
	err := NotFound(response.ErrorCode("SALES_ORDER_NOT_FOUND"), "Sales order not found", cause, map[string]any{
		"sales_order_id": "so-1",
	})

	if err.HTTPStatus != http.StatusNotFound {
		t.Fatalf("status = %d, want %d", err.HTTPStatus, http.StatusNotFound)
	}
	if err.Code != "SALES_ORDER_NOT_FOUND" {
		t.Fatalf("code = %q, want SALES_ORDER_NOT_FOUND", err.Code)
	}
	if err.Error() != "Sales order not found" {
		t.Fatalf("message = %q, want Sales order not found", err.Error())
	}
	if !stderrors.Is(err, cause) {
		t.Fatal("errors.Is did not unwrap the cause")
	}

	appErr, ok := As(err)
	if !ok {
		t.Fatal("As did not return AppError")
	}
	if appErr.Details["sales_order_id"] != "so-1" {
		t.Fatalf("details = %+v, want sales order id", appErr.Details)
	}
}
