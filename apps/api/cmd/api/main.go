package main

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	inventoryapp "github.com/Chinsusu/ERP-v2/apps/api/internal/modules/inventory/application"
	"github.com/Chinsusu/ERP-v2/apps/api/internal/modules/inventory/domain"
	"github.com/Chinsusu/ERP-v2/apps/api/internal/shared/audit"
	"github.com/Chinsusu/ERP-v2/apps/api/internal/shared/auth"
	"github.com/Chinsusu/ERP-v2/apps/api/internal/shared/config"
	"github.com/Chinsusu/ERP-v2/apps/api/internal/shared/response"
)

type healthResponse struct {
	Status    string `json:"status"`
	Service   string `json:"service"`
	Timestamp string `json:"timestamp"`
}

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type loginResponse struct {
	AccessToken string       `json:"access_token"`
	TokenType   string       `json:"token_type"`
	ExpiresIn   int          `json:"expires_in"`
	User        userResponse `json:"user"`
}

type userResponse struct {
	ID          string   `json:"id"`
	Email       string   `json:"email"`
	Name        string   `json:"name"`
	Role        string   `json:"role"`
	Permissions []string `json:"permissions"`
}

type roleResponse struct {
	Key         string   `json:"key"`
	Name        string   `json:"name"`
	Permissions []string `json:"permissions"`
}

type availableStockResponse struct {
	WarehouseID    string `json:"warehouse_id"`
	WarehouseCode  string `json:"warehouse_code"`
	SKU            string `json:"sku"`
	BatchID        string `json:"batch_id,omitempty"`
	BatchNo        string `json:"batch_no,omitempty"`
	PhysicalStock  int64  `json:"physical_stock"`
	ReservedStock  int64  `json:"reserved_stock"`
	HoldStock      int64  `json:"hold_stock"`
	AvailableStock int64  `json:"available_stock"`
}

type stockMovementRequest struct {
	MovementID   string  `json:"movementId"`
	SKU          string  `json:"sku"`
	WarehouseID  string  `json:"warehouseId"`
	MovementType string  `json:"movementType"`
	Quantity     float64 `json:"quantity"`
	Reason       string  `json:"reason"`
}

type stockMovementResponse struct {
	MovementID string `json:"movement_id"`
	Status     string `json:"status"`
}

type auditLogResponse struct {
	ID         string         `json:"id"`
	ActorID    string         `json:"actor_id"`
	Action     string         `json:"action"`
	EntityType string         `json:"entity_type"`
	EntityID   string         `json:"entity_id"`
	RequestID  string         `json:"request_id,omitempty"`
	BeforeData map[string]any `json:"before_data,omitempty"`
	AfterData  map[string]any `json:"after_data,omitempty"`
	Metadata   map[string]any `json:"metadata"`
	CreatedAt  string         `json:"created_at"`
}

func main() {
	cfg := config.FromEnv()
	authConfig := auth.MockConfig{
		Email:       cfg.AuthMockEmail,
		Password:    cfg.AuthMockPassword,
		AccessToken: cfg.AuthMockAccessToken,
	}
	availableStockService := inventoryapp.NewListAvailableStock(inventoryapp.NewPrototypeStockAvailabilityStore())
	auditLogStore := audit.NewPrototypeLogStore()

	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", healthHandler)
	mux.HandleFunc("/api/v1/health", healthHandler)
	mux.HandleFunc("/api/v1/auth/mock-login", mockLoginHandler(authConfig))
	mux.Handle("/api/v1/me", auth.RequireBearerToken(authConfig, http.HandlerFunc(meHandler)))
	mux.Handle(
		"/api/v1/rbac/roles",
		auth.RequireBearerPermission(authConfig, auth.PermissionSettingsView, http.HandlerFunc(rbacRolesHandler)),
	)
	mux.Handle(
		"/api/v1/audit-logs",
		auth.RequireBearerPermission(
			authConfig,
			auth.PermissionAuditLogView,
			http.HandlerFunc(auditLogsHandler(auditLogStore)),
		),
	)
	mux.Handle(
		"/api/v1/inventory/stock-movements",
		auth.RequireBearerPermission(
			authConfig,
			auth.PermissionRecordCreate,
			http.HandlerFunc(stockMovementHandler(auditLogStore)),
		),
	)
	mux.Handle(
		"/api/v1/inventory/available-stock",
		auth.RequireBearerPermission(
			authConfig,
			auth.PermissionInventoryView,
			http.HandlerFunc(availableStockHandler(availableStockService)),
		),
	)

	server := &http.Server{
		Addr:              ":" + cfg.AppPort,
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
	}

	log.Printf("api listening on :%s", cfg.AppPort)
	log.Fatal(server.ListenAndServe())
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	response.WriteSuccess(w, r, http.StatusOK, healthResponse{
		Status:    "ok",
		Service:   "api",
		Timestamp: time.Now().UTC().Format(time.RFC3339),
	})
}

func mockLoginHandler(authConfig auth.MockConfig) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			response.WriteError(w, r, http.StatusMethodNotAllowed, response.ErrorCodeNotFound, "Route not found", nil)
			return
		}

		var payload loginRequest
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			response.WriteError(
				w,
				r,
				http.StatusBadRequest,
				response.ErrorCodeValidation,
				"Invalid login payload",
				nil,
			)
			return
		}

		principal, ok := auth.ValidateMockLogin(authConfig, payload.Email, payload.Password)
		if !ok {
			response.WriteError(
				w,
				r,
				http.StatusUnauthorized,
				response.ErrorCodeUnauthorized,
				"Invalid email or password",
				nil,
			)
			return
		}

		response.WriteSuccess(w, r, http.StatusOK, loginResponse{
			AccessToken: authConfig.AccessToken,
			TokenType:   "Bearer",
			ExpiresIn:   28800,
			User:        newUserResponse(principal),
		})
	}
}

func meHandler(w http.ResponseWriter, r *http.Request) {
	principal, ok := auth.PrincipalFromContext(r.Context())
	if !ok {
		response.WriteError(w, r, http.StatusUnauthorized, response.ErrorCodeUnauthorized, "Authentication required", nil)
		return
	}

	response.WriteSuccess(w, r, http.StatusOK, newUserResponse(principal))
}

func newUserResponse(principal auth.Principal) userResponse {
	return userResponse{
		ID:          principal.UserID,
		Email:       principal.Email,
		Name:        principal.Name,
		Role:        string(principal.Role),
		Permissions: permissionStrings(principal.Permissions),
	}
}

func rbacRolesHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		response.WriteError(w, r, http.StatusMethodNotAllowed, response.ErrorCodeNotFound, "Route not found", nil)
		return
	}

	roles := auth.RoleCatalog()
	payload := make([]roleResponse, 0, len(roles))
	for _, role := range roles {
		payload = append(payload, roleResponse{
			Key:         string(role.Key),
			Name:        role.Name,
			Permissions: permissionStrings(role.Permissions),
		})
	}

	response.WriteSuccess(w, r, http.StatusOK, payload)
}

func permissionStrings(permissions []auth.PermissionKey) []string {
	values := make([]string, 0, len(permissions))
	for _, permission := range permissions {
		values = append(values, string(permission))
	}

	return values
}

func auditLogsHandler(store audit.LogStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			response.WriteError(w, r, http.StatusMethodNotAllowed, response.ErrorCodeNotFound, "Route not found", nil)
			return
		}

		logs, err := store.List(r.Context(), audit.Query{
			ActorID:    r.URL.Query().Get("actor_id"),
			Action:     r.URL.Query().Get("action"),
			EntityType: r.URL.Query().Get("entity_type"),
			EntityID:   r.URL.Query().Get("entity_id"),
			Limit:      queryInt(r, "limit"),
		})
		if err != nil {
			response.WriteError(
				w,
				r,
				http.StatusConflict,
				response.ErrorCodeConflict,
				"Audit logs could not be loaded",
				nil,
			)
			return
		}

		payload := make([]auditLogResponse, 0, len(logs))
		for _, log := range logs {
			payload = append(payload, newAuditLogResponse(log))
		}

		response.WriteSuccess(w, r, http.StatusOK, payload)
	}
}

func stockMovementHandler(store audit.LogStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			response.WriteError(w, r, http.StatusMethodNotAllowed, response.ErrorCodeNotFound, "Route not found", nil)
			return
		}

		r = requestWithStableID(r)
		var payload stockMovementRequest
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			response.WriteError(
				w,
				r,
				http.StatusBadRequest,
				response.ErrorCodeValidation,
				"Invalid stock movement payload",
				nil,
			)
			return
		}
		if details := validateStockMovementPayload(payload); len(details) > 0 {
			response.WriteError(
				w,
				r,
				http.StatusBadRequest,
				response.ErrorCodeValidation,
				"Invalid stock movement payload",
				details,
			)
			return
		}

		if strings.EqualFold(strings.TrimSpace(payload.MovementType), "ADJUST") {
			if err := recordStockAdjustmentAudit(r, store, payload); err != nil {
				response.WriteError(
					w,
					r,
					http.StatusConflict,
					response.ErrorCodeConflict,
					"Audit log could not be recorded",
					nil,
				)
				return
			}
		}

		response.WriteSuccess(w, r, http.StatusCreated, stockMovementResponse{
			MovementID: strings.TrimSpace(payload.MovementID),
			Status:     "recorded",
		})
	}
}

func availableStockHandler(service inventoryapp.ListAvailableStock) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			response.WriteError(w, r, http.StatusMethodNotAllowed, response.ErrorCodeNotFound, "Route not found", nil)
			return
		}

		filter := domain.NewAvailableStockFilter(
			r.URL.Query().Get("warehouse_id"),
			r.URL.Query().Get("sku"),
			r.URL.Query().Get("batch_id"),
		)
		snapshots, err := service.Execute(r.Context(), filter)
		if err != nil {
			response.WriteError(
				w,
				r,
				http.StatusConflict,
				response.ErrorCodeConflict,
				"Available stock could not be calculated",
				nil,
			)
			return
		}

		payload := make([]availableStockResponse, 0, len(snapshots))
		for _, snapshot := range snapshots {
			payload = append(payload, newAvailableStockResponse(snapshot))
		}

		response.WriteSuccess(w, r, http.StatusOK, payload)
	}
}

func newAvailableStockResponse(snapshot domain.AvailableStockSnapshot) availableStockResponse {
	return availableStockResponse{
		WarehouseID:    snapshot.WarehouseID,
		WarehouseCode:  snapshot.WarehouseCode,
		SKU:            snapshot.SKU,
		BatchID:        snapshot.BatchID,
		BatchNo:        snapshot.BatchNo,
		PhysicalStock:  snapshot.PhysicalStock,
		ReservedStock:  snapshot.ReservedStock,
		HoldStock:      snapshot.HoldStock,
		AvailableStock: snapshot.AvailableStock,
	}
}

func newAuditLogResponse(log audit.Log) auditLogResponse {
	metadata := log.Metadata
	if metadata == nil {
		metadata = map[string]any{}
	}

	return auditLogResponse{
		ID:         log.ID,
		ActorID:    log.ActorID,
		Action:     log.Action,
		EntityType: log.EntityType,
		EntityID:   log.EntityID,
		RequestID:  log.RequestID,
		BeforeData: log.BeforeData,
		AfterData:  log.AfterData,
		Metadata:   metadata,
		CreatedAt:  log.CreatedAt.UTC().Format(time.RFC3339),
	}
}

func validateStockMovementPayload(payload stockMovementRequest) map[string]any {
	details := make(map[string]any)
	if strings.TrimSpace(payload.MovementID) == "" {
		details["movementId"] = "required"
	}
	if strings.TrimSpace(payload.SKU) == "" {
		details["sku"] = "required"
	}
	if strings.TrimSpace(payload.WarehouseID) == "" {
		details["warehouseId"] = "required"
	}
	if strings.TrimSpace(payload.Reason) == "" {
		details["reason"] = "required"
	}
	if payload.Quantity <= 0 {
		details["quantity"] = "must be positive"
	}
	switch strings.ToUpper(strings.TrimSpace(payload.MovementType)) {
	case "RECEIVE", "ISSUE", "TRANSFER_IN", "ADJUST":
	default:
		details["movementType"] = "unsupported"
	}

	return details
}

func recordStockAdjustmentAudit(r *http.Request, store audit.LogStore, payload stockMovementRequest) error {
	principal, ok := auth.PrincipalFromContext(r.Context())
	if !ok {
		return http.ErrNoCookie
	}

	log, err := audit.NewLog(audit.NewLogInput{
		ActorID:    principal.UserID,
		Action:     "inventory.stock_movement.adjusted",
		EntityType: "inventory.stock_movement",
		EntityID:   strings.TrimSpace(payload.MovementID),
		RequestID:  response.RequestID(r),
		AfterData: map[string]any{
			"movement_type": strings.ToUpper(strings.TrimSpace(payload.MovementType)),
			"quantity":      payload.Quantity,
			"warehouse_id":  strings.TrimSpace(payload.WarehouseID),
			"sku":           strings.ToUpper(strings.TrimSpace(payload.SKU)),
		},
		Metadata: map[string]any{
			"reason": strings.TrimSpace(payload.Reason),
			"source": "inventory stock movement",
		},
	})
	if err != nil {
		return err
	}

	return store.Record(r.Context(), log)
}

func queryInt(r *http.Request, key string) int {
	value := strings.TrimSpace(r.URL.Query().Get(key))
	if value == "" {
		return 0
	}
	parsed, err := strconv.Atoi(value)
	if err != nil {
		return 0
	}

	return parsed
}

func requestWithStableID(r *http.Request) *http.Request {
	if strings.TrimSpace(r.Header.Get(response.HeaderRequestID)) != "" {
		return r
	}

	clone := r.Clone(r.Context())
	clone.Header.Set(response.HeaderRequestID, response.RequestID(r))
	return clone
}
