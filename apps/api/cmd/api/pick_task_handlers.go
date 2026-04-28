package main

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"

	shippingapp "github.com/Chinsusu/ERP-v2/apps/api/internal/modules/shipping/application"
	shippingdomain "github.com/Chinsusu/ERP-v2/apps/api/internal/modules/shipping/domain"
	"github.com/Chinsusu/ERP-v2/apps/api/internal/shared/auth"
	"github.com/Chinsusu/ERP-v2/apps/api/internal/shared/response"
)

type pickTaskLineResponse struct {
	ID                 string `json:"id"`
	LineNo             int    `json:"line_no"`
	SalesOrderLineID   string `json:"sales_order_line_id"`
	StockReservationID string `json:"stock_reservation_id"`
	ItemID             string `json:"item_id"`
	SKUCode            string `json:"sku_code"`
	BatchID            string `json:"batch_id"`
	BatchNo            string `json:"batch_no"`
	WarehouseID        string `json:"warehouse_id"`
	BinID              string `json:"bin_id"`
	BinCode            string `json:"bin_code"`
	QtyToPick          string `json:"qty_to_pick"`
	QtyPicked          string `json:"qty_picked"`
	BaseUOMCode        string `json:"base_uom_code"`
	Status             string `json:"status"`
	PickedAt           string `json:"picked_at,omitempty"`
	PickedBy           string `json:"picked_by,omitempty"`
	CreatedAt          string `json:"created_at"`
	UpdatedAt          string `json:"updated_at"`
}

type pickTaskResponse struct {
	ID            string                 `json:"id"`
	OrgID         string                 `json:"org_id"`
	PickTaskNo    string                 `json:"pick_task_no"`
	SalesOrderID  string                 `json:"sales_order_id"`
	OrderNo       string                 `json:"order_no"`
	WarehouseID   string                 `json:"warehouse_id"`
	WarehouseCode string                 `json:"warehouse_code"`
	Status        string                 `json:"status"`
	AssignedTo    string                 `json:"assigned_to,omitempty"`
	AssignedAt    string                 `json:"assigned_at,omitempty"`
	StartedAt     string                 `json:"started_at,omitempty"`
	StartedBy     string                 `json:"started_by,omitempty"`
	CompletedAt   string                 `json:"completed_at,omitempty"`
	CompletedBy   string                 `json:"completed_by,omitempty"`
	AuditLogID    string                 `json:"audit_log_id,omitempty"`
	Lines         []pickTaskLineResponse `json:"lines"`
	CreatedAt     string                 `json:"created_at"`
	UpdatedAt     string                 `json:"updated_at"`
}

type confirmPickTaskLineRequest struct {
	LineID    string `json:"line_id"`
	PickedQty string `json:"picked_qty"`
}

type pickTaskExceptionRequest struct {
	LineID        string `json:"line_id"`
	ExceptionCode string `json:"exception_code"`
	Investigation string `json:"investigation"`
}

func pickTasksHandler(listService shippingapp.ListPickTasks) http.HandlerFunc {
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

		tasks, err := listService.Execute(r.Context(), pickTaskFilterFromRequest(r))
		if err != nil {
			writePickTaskError(w, r, err)
			return
		}

		payload := make([]pickTaskResponse, 0, len(tasks))
		for _, task := range tasks {
			payload = append(payload, newPickTaskResponse(task, ""))
		}
		response.WriteSuccess(w, r, http.StatusOK, payload)
	}
}

func pickTaskDetailHandler(getService shippingapp.GetPickTask) http.HandlerFunc {
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

		task, err := getService.Execute(r.Context(), r.PathValue("pick_task_id"))
		if err != nil {
			writePickTaskError(w, r, err)
			return
		}
		response.WriteSuccess(w, r, http.StatusOK, newPickTaskResponse(task, ""))
	}
}

func startPickTaskHandler(service shippingapp.StartPickTask) http.HandlerFunc {
	return pickTaskActionHandler(func(_ http.ResponseWriter, r *http.Request, actorID string) (shippingapp.PickTaskResult, bool, error) {
		result, err := service.Execute(r.Context(), shippingapp.PickTaskActionInput{
			PickTaskID: r.PathValue("pick_task_id"),
			ActorID:    actorID,
			RequestID:  response.RequestID(r),
		})
		return result, false, err
	})
}

func completePickTaskHandler(service shippingapp.CompletePickTask) http.HandlerFunc {
	return pickTaskActionHandler(func(_ http.ResponseWriter, r *http.Request, actorID string) (shippingapp.PickTaskResult, bool, error) {
		result, err := service.Execute(r.Context(), shippingapp.PickTaskActionInput{
			PickTaskID: r.PathValue("pick_task_id"),
			ActorID:    actorID,
			RequestID:  response.RequestID(r),
		})
		return result, false, err
	})
}

func confirmPickTaskLineHandler(service shippingapp.ConfirmPickTaskLine) http.HandlerFunc {
	return pickTaskActionHandler(func(w http.ResponseWriter, r *http.Request, actorID string) (shippingapp.PickTaskResult, bool, error) {
		var payload confirmPickTaskLineRequest
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			response.WriteError(w, r, http.StatusBadRequest, response.ErrorCodeValidation, "Invalid pick task line payload", nil)
			return shippingapp.PickTaskResult{}, true, nil
		}

		result, err := service.Execute(r.Context(), shippingapp.ConfirmPickTaskLineInput{
			PickTaskID: r.PathValue("pick_task_id"),
			LineID:     payload.LineID,
			PickedQty:  payload.PickedQty,
			ActorID:    actorID,
			RequestID:  response.RequestID(r),
		})
		return result, false, err
	})
}

func reportPickTaskExceptionHandler(service shippingapp.ReportPickTaskException) http.HandlerFunc {
	return pickTaskActionHandler(func(w http.ResponseWriter, r *http.Request, actorID string) (shippingapp.PickTaskResult, bool, error) {
		var payload pickTaskExceptionRequest
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			response.WriteError(w, r, http.StatusBadRequest, response.ErrorCodeValidation, "Invalid pick task exception payload", nil)
			return shippingapp.PickTaskResult{}, true, nil
		}

		result, err := service.Execute(r.Context(), shippingapp.ReportPickTaskExceptionInput{
			PickTaskID:    r.PathValue("pick_task_id"),
			LineID:        payload.LineID,
			ExceptionCode: payload.ExceptionCode,
			ActorID:       actorID,
			RequestID:     response.RequestID(r),
			Investigation: payload.Investigation,
		})
		return result, false, err
	})
}

type pickTaskAction func(w http.ResponseWriter, r *http.Request, actorID string) (shippingapp.PickTaskResult, bool, error)

func pickTaskActionHandler(action pickTaskAction) http.HandlerFunc {
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
			writePickTaskError(w, r, err)
			return
		}

		response.WriteSuccess(w, r, http.StatusOK, newPickTaskResponse(result.PickTask, result.AuditLogID))
	}
}

func pickTaskFilterFromRequest(r *http.Request) shippingapp.PickTaskFilter {
	return shippingapp.PickTaskFilter{
		WarehouseID: r.URL.Query().Get("warehouse_id"),
		Status:      shippingdomain.PickTaskStatus(r.URL.Query().Get("status")),
		AssignedTo:  r.URL.Query().Get("assigned_to"),
	}
}

func newPickTaskResponse(task shippingdomain.PickTask, auditLogID string) pickTaskResponse {
	payload := pickTaskResponse{
		ID:            task.ID,
		OrgID:         task.OrgID,
		PickTaskNo:    task.PickTaskNo,
		SalesOrderID:  task.SalesOrderID,
		OrderNo:       task.OrderNo,
		WarehouseID:   task.WarehouseID,
		WarehouseCode: task.WarehouseCode,
		Status:        string(task.Status),
		AssignedTo:    task.AssignedTo,
		AssignedAt:    timeString(task.AssignedAt),
		StartedAt:     timeString(task.StartedAt),
		StartedBy:     task.StartedBy,
		CompletedAt:   timeString(task.CompletedAt),
		CompletedBy:   task.CompletedBy,
		AuditLogID:    auditLogID,
		Lines:         make([]pickTaskLineResponse, 0, len(task.Lines)),
		CreatedAt:     task.CreatedAt.UTC().Format(time.RFC3339),
		UpdatedAt:     task.UpdatedAt.UTC().Format(time.RFC3339),
	}
	for _, line := range task.Lines {
		payload.Lines = append(payload.Lines, pickTaskLineResponse{
			ID:                 line.ID,
			LineNo:             line.LineNo,
			SalesOrderLineID:   line.SalesOrderLineID,
			StockReservationID: line.StockReservationID,
			ItemID:             line.ItemID,
			SKUCode:            line.SKUCode,
			BatchID:            line.BatchID,
			BatchNo:            line.BatchNo,
			WarehouseID:        line.WarehouseID,
			BinID:              line.BinID,
			BinCode:            line.BinCode,
			QtyToPick:          line.QtyToPick.String(),
			QtyPicked:          line.QtyPicked.String(),
			BaseUOMCode:        line.BaseUOMCode.String(),
			Status:             string(line.Status),
			PickedAt:           timeString(line.PickedAt),
			PickedBy:           line.PickedBy,
			CreatedAt:          line.CreatedAt.UTC().Format(time.RFC3339),
			UpdatedAt:          line.UpdatedAt.UTC().Format(time.RFC3339),
		})
	}

	return payload
}

func writePickTaskError(w http.ResponseWriter, r *http.Request, err error) {
	switch {
	case errors.Is(err, shippingapp.ErrPickTaskNotFound):
		response.WriteError(w, r, http.StatusNotFound, response.ErrorCodeNotFound, "Pick task was not found", nil)
	case errors.Is(err, shippingapp.ErrPickTaskDuplicate):
		response.WriteError(w, r, http.StatusConflict, response.ErrorCodeConflict, "Pick task already exists", nil)
	case errors.Is(err, shippingapp.ErrPickTaskSalesOrderNotReserved):
		response.WriteError(w, r, http.StatusConflict, response.ErrorCodeConflict, "Sales order must be reserved before pick task action", nil)
	case errors.Is(err, shippingapp.ErrPickTaskReservationMissing):
		response.WriteError(w, r, http.StatusConflict, response.ErrorCodeConflict, "Active stock reservation is required for pick task", nil)
	case errors.Is(err, shippingdomain.ErrPickTaskInvalidTransition):
		response.WriteError(w, r, http.StatusConflict, response.ErrorCodeConflict, "Pick task status transition is invalid", nil)
	case errors.Is(err, shippingdomain.ErrPickTaskRequiredField):
		response.WriteError(w, r, http.StatusBadRequest, response.ErrorCodeValidation, "Invalid pick task payload", nil)
	case errors.Is(err, shippingdomain.ErrPickTaskInvalidQuantity):
		response.WriteError(w, r, http.StatusBadRequest, response.ErrorCodeValidation, "Pick task quantity is invalid", nil)
	case errors.Is(err, shippingdomain.ErrPickTaskInvalidStatus):
		response.WriteError(w, r, http.StatusBadRequest, response.ErrorCodeValidation, "Pick task status is invalid", nil)
	case errors.Is(err, shippingdomain.ErrPickTaskActorRequired):
		response.WriteError(w, r, http.StatusBadRequest, response.ErrorCodeValidation, "Pick task actor is required", nil)
	default:
		response.WriteError(w, r, http.StatusConflict, response.ErrorCodeConflict, "Pick task request could not be processed", nil)
	}
}

func mustPrototypePickTask() shippingdomain.PickTask {
	createdAt := time.Date(2026, 4, 28, 9, 30, 0, 0, time.UTC)
	task, err := shippingdomain.NewPickTask(shippingdomain.NewPickTaskInput{
		ID:            "pick-so-260428-0001",
		OrgID:         "org-my-pham",
		PickTaskNo:    "PICK-SO-260428-0001",
		SalesOrderID:  "so-260428-0001",
		OrderNo:       "SO-260428-0001",
		WarehouseID:   "wh-hcm-fg",
		WarehouseCode: "WH-HCM-FG",
		CreatedAt:     createdAt,
		Lines: []shippingdomain.NewPickTaskLineInput{
			{
				ID:                 "pick-so-260428-0001-line-01",
				LineNo:             1,
				SalesOrderLineID:   "so-260428-0001-line-01",
				StockReservationID: "rsv-so-260428-0001-line-01",
				ItemID:             "item-serum-30ml",
				SKUCode:            "SERUM-30ML",
				BatchID:            "batch-serum-2604a",
				BatchNo:            "LOT-2604A",
				WarehouseID:        "wh-hcm-fg",
				BinID:              "bin-hcm-pick-a01",
				BinCode:            "PICK-A-01",
				QtyToPick:          "3.000000",
				BaseUOMCode:        "EA",
			},
		},
	})
	if err != nil {
		panic(err)
	}

	return task
}
