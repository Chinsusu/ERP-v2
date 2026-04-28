package errors

import (
	stderrors "errors"
	"net/http"
	"strings"

	"github.com/Chinsusu/ERP-v2/apps/api/internal/shared/response"
)

type AppError struct {
	Code       response.ErrorCode
	Message    string
	Details    map[string]any
	HTTPStatus int
	Cause      error
}

func New(
	code response.ErrorCode,
	message string,
	httpStatus int,
	cause error,
	details map[string]any,
) AppError {
	return AppError{
		Code:       code,
		Message:    strings.TrimSpace(message),
		Details:    cloneDetails(details),
		HTTPStatus: httpStatus,
		Cause:      cause,
	}
}

func BadRequest(code response.ErrorCode, message string, cause error, details map[string]any) AppError {
	return New(code, message, http.StatusBadRequest, cause, details)
}

func NotFound(code response.ErrorCode, message string, cause error, details map[string]any) AppError {
	return New(code, message, http.StatusNotFound, cause, details)
}

func Conflict(code response.ErrorCode, message string, cause error, details map[string]any) AppError {
	return New(code, message, http.StatusConflict, cause, details)
}

func Unprocessable(code response.ErrorCode, message string, cause error, details map[string]any) AppError {
	return New(code, message, http.StatusUnprocessableEntity, cause, details)
}

func (e AppError) Error() string {
	if e.Message != "" {
		return e.Message
	}
	if e.Code != "" {
		return string(e.Code)
	}
	if e.Cause != nil {
		return e.Cause.Error()
	}

	return "application error"
}

func (e AppError) Unwrap() error {
	return e.Cause
}

func As(err error) (AppError, bool) {
	var appErr AppError
	if stderrors.As(err, &appErr) {
		return appErr, true
	}

	return AppError{}, false
}

func cloneDetails(details map[string]any) map[string]any {
	if details == nil {
		return nil
	}

	clone := make(map[string]any, len(details))
	for key, value := range details {
		clone[key] = value
	}

	return clone
}
