package main

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	inventoryapp "github.com/Chinsusu/ERP-v2/apps/api/internal/modules/inventory/application"
	"github.com/Chinsusu/ERP-v2/apps/api/internal/modules/inventory/domain"
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

func main() {
	cfg := config.FromEnv()
	authConfig := auth.MockConfig{
		Email:       cfg.AuthMockEmail,
		Password:    cfg.AuthMockPassword,
		AccessToken: cfg.AuthMockAccessToken,
	}
	availableStockService := inventoryapp.NewListAvailableStock(inventoryapp.NewPrototypeStockAvailabilityStore())

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
