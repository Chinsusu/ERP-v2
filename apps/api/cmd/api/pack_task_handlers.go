package main

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"time"

	salesapp "github.com/Chinsusu/ERP-v2/apps/api/internal/modules/sales/application"
	salesdomain "github.com/Chinsusu/ERP-v2/apps/api/internal/modules/sales/domain"
	shippingapp "github.com/Chinsusu/ERP-v2/apps/api/internal/modules/shipping/application"
	shippingdomain "github.com/Chinsusu/ERP-v2/apps/api/internal/modules/shipping/domain"
	"github.com/Chinsusu/ERP-v2/apps/api/internal/shared/auth"
	apperrors "github.com/Chinsusu/ERP-v2/apps/api/internal/shared/errors"
	"github.com/Chinsusu/ERP-v2/apps/api/internal/shared/response"
)

type salesOrderPackerAdapter struct {
	service salesapp.SalesOrderService
}

func (a salesOrderPackerAdapter) MarkSalesOrderPacked(
	ctx context.Context,
	input shippingapp.PackTaskSalesOrderPackedInput,
) (salesdomain.SalesOrder, error) {
	result, err := a.service.MarkSalesOrderPacked(ctx, salesapp.SalesOrderActionInput{
		ID:        input.SalesOrderID,
		ActorID:   input.ActorID,
		RequestID: input.RequestID,
	})
	if err != nil {
		return salesdomain.SalesOrder{}, err
	}

	return result.SalesOrder, nil
}

type packTaskLineResponse struct {
	ID               string `json:"id"`
	LineNo           int    `json:"line_no"`
	PickTaskLineID   string `json:"pick_task_line_id"`
	SalesOrderLineID string `json:"sales_order_line_id"`
	ItemID           string `json:"item_id"`
	SKUCode          string `json:"sku_code"`
	BatchID          string `json:"batch_id"`
	BatchNo          string `json:"batch_no"`
	WarehouseID      string `json:"warehouse_id"`
	QtyToPack        string `json:"qty_to_pack"`
	QtyPacked        string `json:"qty_packed"`
	BaseUOMCode      string `json:"base_uom_code"`
	Status           string `json:"status"`
	PackedAt         string `json:"packed_at,omitempty"`
	PackedBy         string `json:"packed_by,omitempty"`
	CreatedAt        string `json:"created_at"`
	UpdatedAt        string `json:"updated_at"`
}

type packTaskResponse struct {
	ID               string                 `json:"id"`
	OrgID            string                 `json:"org_id"`
	PackTaskNo       string                 `json:"pack_task_no"`
	SalesOrderID     string                 `json:"sales_order_id"`
	SalesOrderStatus string                 `json:"sales_order_status,omitempty"`
	OrderNo          string                 `json:"order_no"`
	PickTaskID       string                 `json:"pick_task_id"`
	PickTaskNo       string                 `json:"pick_task_no"`
	WarehouseID      string                 `json:"warehouse_id"`
	WarehouseCode    string                 `json:"warehouse_code"`
	Status           string                 `json:"status"`
	AssignedTo       string                 `json:"assigned_to,omitempty"`
	AssignedAt       string                 `json:"assigned_at,omitempty"`
	StartedAt        string                 `json:"started_at,omitempty"`
	StartedBy        string                 `json:"started_by,omitempty"`
	PackedAt         string                 `json:"packed_at,omitempty"`
	PackedBy         string                 `json:"packed_by,omitempty"`
	AuditLogID       string                 `json:"audit_log_id,omitempty"`
	Lines            []packTaskLineResponse `json:"lines"`
	CreatedAt        string                 `json:"created_at"`
	UpdatedAt        string                 `json:"updated_at"`
}

type confirmPackTaskRequest struct {
	Lines []confirmPackTaskLineRequest `json:"lines"`
}

type confirmPackTaskLineRequest struct {
	LineID    string `json:"line_id"`
	PackedQty string `json:"packed_qty"`
}

type packTaskExceptionRequest struct {
	LineID        string `json:"line_id"`
	ExceptionCode string `json:"exception_code"`
	Investigation string `json:"investigation"`
}

func packTasksHandler(listService shippingapp.ListPackTasks) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			response.WriteError(w, r, http.StatusMethodNotAllowed, response.ErrorCodeNotFound, "Route not found", nil)
			return
		}
		principal, ok := auth.PrincipalFromContext(r.Context())
		if !ok {
			response.WriteError(w, r, http.StatusUnauthorized, response.ErrorCodeUnauthorized, "Authentication required", nil)
			return
		}
		if !auth.HasPermission(principal, auth.PermissionShippingView) {
			writePermissionDenied(w, r, auth.PermissionShippingView)
			return
		}

		tasks, err := listService.Execute(r.Context(), packTaskFilterFromRequest(r))
		if err != nil {
			writePackTaskError(w, r, err)
			return
		}

		payload := make([]packTaskResponse, 0, len(tasks))
		for _, task := range tasks {
			payload = append(payload, newPackTaskResponse(task, "", ""))
		}
		response.WriteSuccess(w, r, http.StatusOK, payload)
	}
}

func packTaskDetailHandler(getService shippingapp.GetPackTask) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			response.WriteError(w, r, http.StatusMethodNotAllowed, response.ErrorCodeNotFound, "Route not found", nil)
			return
		}
		principal, ok := auth.PrincipalFromContext(r.Context())
		if !ok {
			response.WriteError(w, r, http.StatusUnauthorized, response.ErrorCodeUnauthorized, "Authentication required", nil)
			return
		}
		if !auth.HasPermission(principal, auth.PermissionShippingView) {
			writePermissionDenied(w, r, auth.PermissionShippingView)
			return
		}

		task, err := getService.Execute(r.Context(), r.PathValue("pack_task_id"))
		if err != nil {
			writePackTaskError(w, r, err)
			return
		}
		response.WriteSuccess(w, r, http.StatusOK, newPackTaskResponse(task, "", ""))
	}
}

func startPackTaskHandler(service shippingapp.StartPackTask) http.HandlerFunc {
	return packTaskActionHandler(func(_ http.ResponseWriter, r *http.Request, actorID string) (shippingapp.PackTaskResult, bool, error) {
		result, err := service.Execute(r.Context(), shippingapp.PackTaskActionInput{
			PackTaskID: r.PathValue("pack_task_id"),
			ActorID:    actorID,
			RequestID:  response.RequestID(r),
		})
		return result, false, err
	})
}

func confirmPackTaskHandler(service shippingapp.ConfirmPackTask) http.HandlerFunc {
	return packTaskActionHandler(func(w http.ResponseWriter, r *http.Request, actorID string) (shippingapp.PackTaskResult, bool, error) {
		var payload confirmPackTaskRequest
		if r.Body != nil {
			if err := json.NewDecoder(r.Body).Decode(&payload); err != nil && !errors.Is(err, io.EOF) {
				response.WriteError(w, r, http.StatusBadRequest, response.ErrorCodeValidation, "Invalid pack task confirm payload", nil)
				return shippingapp.PackTaskResult{}, true, nil
			}
		}

		result, err := service.Execute(r.Context(), shippingapp.ConfirmPackTaskInput{
			PackTaskID: r.PathValue("pack_task_id"),
			Lines:      packTaskLineInputs(payload.Lines),
			ActorID:    actorID,
			RequestID:  response.RequestID(r),
		})
		return result, false, err
	})
}

func reportPackTaskExceptionHandler(service shippingapp.ReportPackTaskException) http.HandlerFunc {
	return packTaskActionHandler(func(w http.ResponseWriter, r *http.Request, actorID string) (shippingapp.PackTaskResult, bool, error) {
		var payload packTaskExceptionRequest
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			response.WriteError(w, r, http.StatusBadRequest, response.ErrorCodeValidation, "Invalid pack task exception payload", nil)
			return shippingapp.PackTaskResult{}, true, nil
		}

		result, err := service.Execute(r.Context(), shippingapp.ReportPackTaskExceptionInput{
			PackTaskID:    r.PathValue("pack_task_id"),
			LineID:        payload.LineID,
			ExceptionCode: payload.ExceptionCode,
			ActorID:       actorID,
			RequestID:     response.RequestID(r),
			Investigation: payload.Investigation,
		})
		return result, false, err
	})
}

type packTaskAction func(w http.ResponseWriter, r *http.Request, actorID string) (shippingapp.PackTaskResult, bool, error)

func packTaskActionHandler(action packTaskAction) http.HandlerFunc {
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
		if !auth.HasPermission(principal, auth.PermissionShippingView) {
			writePermissionDenied(w, r, auth.PermissionShippingView)
			return
		}
		if !auth.HasPermission(principal, auth.PermissionRecordCreate) {
			writePermissionDenied(w, r, auth.PermissionRecordCreate)
			return
		}

		r = requestWithStableID(r)
		result, handled, err := action(w, r, principal.UserID)
		if handled {
			return
		}
		if err != nil {
			writePackTaskError(w, r, err)
			return
		}

		response.WriteSuccess(w, r, http.StatusOK, newPackTaskResponse(result.PackTask, result.AuditLogID, string(result.SalesOrder.Status)))
	}
}

func packTaskFilterFromRequest(r *http.Request) shippingapp.PackTaskFilter {
	return shippingapp.PackTaskFilter{
		WarehouseID: r.URL.Query().Get("warehouse_id"),
		Status:      shippingdomain.PackTaskStatus(r.URL.Query().Get("status")),
		AssignedTo:  r.URL.Query().Get("assigned_to"),
	}
}

func packTaskLineInputs(lines []confirmPackTaskLineRequest) []shippingapp.ConfirmPackTaskLineInput {
	if lines == nil {
		return nil
	}
	result := make([]shippingapp.ConfirmPackTaskLineInput, 0, len(lines))
	for _, line := range lines {
		result = append(result, shippingapp.ConfirmPackTaskLineInput{
			LineID:    line.LineID,
			PackedQty: line.PackedQty,
		})
	}

	return result
}

func newPackTaskResponse(task shippingdomain.PackTask, auditLogID string, salesOrderStatus string) packTaskResponse {
	payload := packTaskResponse{
		ID:               task.ID,
		OrgID:            task.OrgID,
		PackTaskNo:       task.PackTaskNo,
		SalesOrderID:     task.SalesOrderID,
		SalesOrderStatus: salesOrderStatus,
		OrderNo:          task.OrderNo,
		PickTaskID:       task.PickTaskID,
		PickTaskNo:       task.PickTaskNo,
		WarehouseID:      task.WarehouseID,
		WarehouseCode:    task.WarehouseCode,
		Status:           string(task.Status),
		AssignedTo:       task.AssignedTo,
		AssignedAt:       timeString(task.AssignedAt),
		StartedAt:        timeString(task.StartedAt),
		StartedBy:        task.StartedBy,
		PackedAt:         timeString(task.PackedAt),
		PackedBy:         task.PackedBy,
		AuditLogID:       auditLogID,
		Lines:            make([]packTaskLineResponse, 0, len(task.Lines)),
		CreatedAt:        task.CreatedAt.UTC().Format(time.RFC3339),
		UpdatedAt:        task.UpdatedAt.UTC().Format(time.RFC3339),
	}
	for _, line := range task.Lines {
		payload.Lines = append(payload.Lines, packTaskLineResponse{
			ID:               line.ID,
			LineNo:           line.LineNo,
			PickTaskLineID:   line.PickTaskLineID,
			SalesOrderLineID: line.SalesOrderLineID,
			ItemID:           line.ItemID,
			SKUCode:          line.SKUCode,
			BatchID:          line.BatchID,
			BatchNo:          line.BatchNo,
			WarehouseID:      line.WarehouseID,
			QtyToPack:        line.QtyToPack.String(),
			QtyPacked:        line.QtyPacked.String(),
			BaseUOMCode:      line.BaseUOMCode.String(),
			Status:           string(line.Status),
			PackedAt:         timeString(line.PackedAt),
			PackedBy:         line.PackedBy,
			CreatedAt:        line.CreatedAt.UTC().Format(time.RFC3339),
			UpdatedAt:        line.UpdatedAt.UTC().Format(time.RFC3339),
		})
	}

	return payload
}

func writePackTaskError(w http.ResponseWriter, r *http.Request, err error) {
	if appErr, ok := apperrors.As(err); ok {
		response.WriteError(w, r, appErr.HTTPStatus, appErr.Code, appErr.Message, appErr.Details)
		return
	}
	switch {
	case errors.Is(err, shippingapp.ErrPackTaskNotFound):
		response.WriteError(w, r, http.StatusNotFound, response.ErrorCodeNotFound, "Pack task was not found", nil)
	case errors.Is(err, shippingapp.ErrPackTaskDuplicate):
		response.WriteError(w, r, http.StatusConflict, response.ErrorCodeConflict, "Pack task already exists", nil)
	case errors.Is(err, shippingapp.ErrPackTaskSalesOrderNotPicked):
		response.WriteError(w, r, http.StatusConflict, response.ErrorCodeConflict, "Sales order must be picked before pack task action", nil)
	case errors.Is(err, shippingapp.ErrPackTaskPickTaskNotCompleted):
		response.WriteError(w, r, http.StatusConflict, response.ErrorCodeConflict, "Pick task must be completed before pack task action", nil)
	case errors.Is(err, shippingdomain.ErrPackTaskInvalidTransition):
		response.WriteError(w, r, http.StatusConflict, response.ErrorCodeConflict, "Pack task status transition is invalid", nil)
	case errors.Is(err, shippingdomain.ErrPackTaskRequiredField):
		response.WriteError(w, r, http.StatusBadRequest, response.ErrorCodeValidation, "Invalid pack task payload", nil)
	case errors.Is(err, shippingdomain.ErrPackTaskInvalidQuantity):
		response.WriteError(w, r, http.StatusBadRequest, response.ErrorCodeValidation, "Pack task quantity is invalid", nil)
	case errors.Is(err, shippingdomain.ErrPackTaskInvalidStatus):
		response.WriteError(w, r, http.StatusBadRequest, response.ErrorCodeValidation, "Pack task status is invalid", nil)
	case errors.Is(err, shippingdomain.ErrPackTaskActorRequired):
		response.WriteError(w, r, http.StatusBadRequest, response.ErrorCodeValidation, "Pack task actor is required", nil)
	default:
		response.WriteError(w, r, http.StatusConflict, response.ErrorCodeConflict, "Pack task request could not be processed", nil)
	}
}

func mustPrototypePackTask() shippingdomain.PackTask {
	createdAt := time.Date(2026, 4, 28, 10, 45, 0, 0, time.UTC)
	task, err := shippingdomain.NewPackTask(shippingdomain.NewPackTaskInput{
		ID:            "pack-so-260428-0003",
		OrgID:         "org-my-pham",
		PackTaskNo:    "PACK-SO-260428-0003",
		SalesOrderID:  "so-260428-0003",
		OrderNo:       "SO-260428-0003",
		PickTaskID:    "pick-so-260428-0003",
		PickTaskNo:    "PICK-SO-260428-0003",
		WarehouseID:   "wh-hcm-fg",
		WarehouseCode: "WH-HCM-FG",
		CreatedAt:     createdAt,
		Lines: []shippingdomain.NewPackTaskLineInput{
			{
				ID:               "pack-so-260428-0003-line-01",
				LineNo:           1,
				PickTaskLineID:   "pick-so-260428-0003-line-01",
				SalesOrderLineID: "so-260428-0003-line-01",
				ItemID:           "item-serum-30ml",
				SKUCode:          "SERUM-30ML",
				BatchID:          "batch-serum-2604a",
				BatchNo:          "LOT-2604A",
				WarehouseID:      "wh-hcm-fg",
				QtyToPack:        "3.000000",
				BaseUOMCode:      "EA",
			},
		},
	})
	if err != nil {
		panic(err)
	}

	return task
}
