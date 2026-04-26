package main

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	inventoryapp "github.com/Chinsusu/ERP-v2/apps/api/internal/modules/inventory/application"
	"github.com/Chinsusu/ERP-v2/apps/api/internal/modules/inventory/domain"
	returnsapp "github.com/Chinsusu/ERP-v2/apps/api/internal/modules/returns/application"
	returnsdomain "github.com/Chinsusu/ERP-v2/apps/api/internal/modules/returns/domain"
	shippingapp "github.com/Chinsusu/ERP-v2/apps/api/internal/modules/shipping/application"
	shippingdomain "github.com/Chinsusu/ERP-v2/apps/api/internal/modules/shipping/domain"
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

type endOfDayReconciliationSummaryResponse struct {
	SystemQuantity     int64 `json:"system_quantity"`
	CountedQuantity    int64 `json:"counted_quantity"`
	VarianceQuantity   int64 `json:"variance_quantity"`
	VarianceCount      int   `json:"variance_count"`
	ChecklistTotal     int   `json:"checklist_total"`
	ChecklistCompleted int   `json:"checklist_completed"`
	ReadyToClose       bool  `json:"ready_to_close"`
}

type endOfDayReconciliationChecklistResponse struct {
	Key      string `json:"key"`
	Label    string `json:"label"`
	Complete bool   `json:"complete"`
	Blocking bool   `json:"blocking"`
	Note     string `json:"note,omitempty"`
}

type endOfDayReconciliationLineResponse struct {
	ID               string `json:"id"`
	SKU              string `json:"sku"`
	BatchNo          string `json:"batch_no"`
	BinCode          string `json:"bin_code"`
	SystemQuantity   int64  `json:"system_quantity"`
	CountedQuantity  int64  `json:"counted_quantity"`
	VarianceQuantity int64  `json:"variance_quantity"`
	Reason           string `json:"reason,omitempty"`
	Owner            string `json:"owner"`
}

type endOfDayReconciliationResponse struct {
	ID            string                                    `json:"id"`
	WarehouseID   string                                    `json:"warehouse_id"`
	WarehouseCode string                                    `json:"warehouse_code"`
	Date          string                                    `json:"date"`
	ShiftCode     string                                    `json:"shift_code"`
	Status        string                                    `json:"status"`
	Owner         string                                    `json:"owner"`
	ClosedAt      string                                    `json:"closed_at,omitempty"`
	ClosedBy      string                                    `json:"closed_by,omitempty"`
	AuditLogID    string                                    `json:"audit_log_id,omitempty"`
	Summary       endOfDayReconciliationSummaryResponse     `json:"summary"`
	Checklist     []endOfDayReconciliationChecklistResponse `json:"checklist"`
	Lines         []endOfDayReconciliationLineResponse      `json:"lines"`
}

type closeReconciliationRequest struct {
	ExceptionNote string `json:"exception_note"`
}

type carrierManifestSummaryResponse struct {
	ExpectedCount int `json:"expected_count"`
	ScannedCount  int `json:"scanned_count"`
	MissingCount  int `json:"missing_count"`
}

type carrierManifestLineResponse struct {
	ID          string `json:"id"`
	ShipmentID  string `json:"shipment_id"`
	OrderNo     string `json:"order_no"`
	TrackingNo  string `json:"tracking_no"`
	PackageCode string `json:"package_code"`
	StagingZone string `json:"staging_zone"`
	Scanned     bool   `json:"scanned"`
}

type carrierManifestResponse struct {
	ID            string                         `json:"id"`
	CarrierCode   string                         `json:"carrier_code"`
	CarrierName   string                         `json:"carrier_name"`
	WarehouseID   string                         `json:"warehouse_id"`
	WarehouseCode string                         `json:"warehouse_code"`
	Date          string                         `json:"date"`
	HandoverBatch string                         `json:"handover_batch"`
	StagingZone   string                         `json:"staging_zone"`
	Status        string                         `json:"status"`
	Owner         string                         `json:"owner"`
	AuditLogID    string                         `json:"audit_log_id,omitempty"`
	Summary       carrierManifestSummaryResponse `json:"summary"`
	Lines         []carrierManifestLineResponse  `json:"lines"`
	CreatedAt     string                         `json:"created_at,omitempty"`
}

type createCarrierManifestRequest struct {
	ID            string `json:"id"`
	CarrierCode   string `json:"carrier_code"`
	CarrierName   string `json:"carrier_name"`
	WarehouseID   string `json:"warehouse_id"`
	WarehouseCode string `json:"warehouse_code"`
	Date          string `json:"date"`
	HandoverBatch string `json:"handover_batch"`
	StagingZone   string `json:"staging_zone"`
	Owner         string `json:"owner"`
}

type addShipmentToManifestRequest struct {
	ShipmentID string `json:"shipment_id"`
}

type verifyCarrierManifestScanRequest struct {
	Code      string `json:"code"`
	StationID string `json:"station_id"`
}

type carrierManifestScanEventResponse struct {
	ID                 string `json:"id"`
	ManifestID         string `json:"manifest_id"`
	ExpectedManifestID string `json:"expected_manifest_id,omitempty"`
	Code               string `json:"code"`
	ResultCode         string `json:"result_code"`
	Severity           string `json:"severity"`
	Message            string `json:"message"`
	ShipmentID         string `json:"shipment_id,omitempty"`
	OrderNo            string `json:"order_no,omitempty"`
	TrackingNo         string `json:"tracking_no,omitempty"`
	ActorID            string `json:"actor_id"`
	StationID          string `json:"station_id"`
	WarehouseID        string `json:"warehouse_id"`
	CarrierCode        string `json:"carrier_code"`
	CreatedAt          string `json:"created_at"`
}

type carrierManifestScanResponse struct {
	ResultCode         string                           `json:"result_code"`
	Severity           string                           `json:"severity"`
	Message            string                           `json:"message"`
	ExpectedManifestID string                           `json:"expected_manifest_id,omitempty"`
	Line               *carrierManifestLineResponse     `json:"line,omitempty"`
	ScanEvent          carrierManifestScanEventResponse `json:"scan_event"`
	Manifest           carrierManifestResponse          `json:"manifest"`
	AuditLogID         string                           `json:"audit_log_id,omitempty"`
}

type returnReceiptLineResponse struct {
	ID          string `json:"id"`
	SKU         string `json:"sku"`
	ProductName string `json:"product_name"`
	Quantity    int    `json:"quantity"`
	Condition   string `json:"condition"`
}

type returnStockMovementResponse struct {
	ID                string `json:"id"`
	MovementType      string `json:"movement_type"`
	SKU               string `json:"sku"`
	WarehouseID       string `json:"warehouse_id"`
	Quantity          int    `json:"quantity"`
	TargetStockStatus string `json:"target_stock_status"`
	SourceDocID       string `json:"source_doc_id"`
}

type returnReceiptResponse struct {
	ID                string                       `json:"id"`
	ReceiptNo         string                       `json:"receipt_no"`
	WarehouseID       string                       `json:"warehouse_id"`
	WarehouseCode     string                       `json:"warehouse_code"`
	Source            string                       `json:"source"`
	ReceivedBy        string                       `json:"received_by"`
	ReceivedAt        string                       `json:"received_at"`
	PackageCondition  string                       `json:"package_condition"`
	Status            string                       `json:"status"`
	Disposition       string                       `json:"disposition"`
	TargetLocation    string                       `json:"target_location"`
	OriginalOrderNo   string                       `json:"original_order_no,omitempty"`
	TrackingNo        string                       `json:"tracking_no,omitempty"`
	ReturnCode        string                       `json:"return_code,omitempty"`
	ScanCode          string                       `json:"scan_code"`
	CustomerName      string                       `json:"customer_name"`
	UnknownCase       bool                         `json:"unknown_case"`
	Lines             []returnReceiptLineResponse  `json:"lines"`
	StockMovement     *returnStockMovementResponse `json:"stock_movement,omitempty"`
	InvestigationNote string                       `json:"investigation_note,omitempty"`
	AuditLogID        string                       `json:"audit_log_id,omitempty"`
	CreatedAt         string                       `json:"created_at"`
}

type receiveReturnRequest struct {
	WarehouseID       string `json:"warehouse_id"`
	WarehouseCode     string `json:"warehouse_code"`
	Source            string `json:"source"`
	Code              string `json:"code"`
	PackageCondition  string `json:"package_condition"`
	Disposition       string `json:"disposition"`
	InvestigationNote string `json:"investigation_note"`
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
	reconciliationStore := inventoryapp.NewPrototypeEndOfDayReconciliationStore()
	listEndOfDayReconciliations := inventoryapp.NewListEndOfDayReconciliations(reconciliationStore)
	closeEndOfDayReconciliation := inventoryapp.NewCloseEndOfDayReconciliation(reconciliationStore, auditLogStore)
	carrierManifestStore := shippingapp.NewPrototypeCarrierManifestStore()
	listCarrierManifests := shippingapp.NewListCarrierManifests(carrierManifestStore)
	createCarrierManifest := shippingapp.NewCreateCarrierManifest(carrierManifestStore, auditLogStore)
	addShipmentToCarrierManifest := shippingapp.NewAddShipmentToCarrierManifest(carrierManifestStore, auditLogStore)
	verifyCarrierManifestScan := shippingapp.NewVerifyCarrierManifestScan(carrierManifestStore, auditLogStore)
	returnReceiptStore := returnsapp.NewPrototypeReturnReceiptStore()
	listReturnReceipts := returnsapp.NewListReturnReceipts(returnReceiptStore)
	receiveReturn := returnsapp.NewReceiveReturn(returnReceiptStore, auditLogStore)

	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", healthHandler)
	mux.HandleFunc("/readyz", readinessHandler)
	mux.HandleFunc("/api/v1/health", healthHandler)
	mux.HandleFunc("/api/v1/ready", readinessHandler)
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
	mux.Handle(
		"/api/v1/warehouse/end-of-day-reconciliations",
		auth.RequireBearerPermission(
			authConfig,
			auth.PermissionWarehouseView,
			http.HandlerFunc(endOfDayReconciliationsHandler(listEndOfDayReconciliations)),
		),
	)
	mux.Handle(
		"/api/v1/warehouse/end-of-day-reconciliations/{reconciliation_id}/close",
		auth.RequireBearerPermission(
			authConfig,
			auth.PermissionRecordCreate,
			http.HandlerFunc(closeEndOfDayReconciliationHandler(closeEndOfDayReconciliation)),
		),
	)
	mux.Handle(
		"/api/v1/shipping/manifests",
		auth.RequireBearerToken(
			authConfig,
			http.HandlerFunc(carrierManifestsHandler(listCarrierManifests, createCarrierManifest)),
		),
	)
	mux.Handle(
		"/api/v1/shipping/manifests/{manifest_id}/shipments",
		auth.RequireBearerPermission(
			authConfig,
			auth.PermissionRecordCreate,
			http.HandlerFunc(addShipmentToCarrierManifestHandler(addShipmentToCarrierManifest)),
		),
	)
	mux.Handle(
		"/api/v1/shipping/manifests/{manifest_id}/scan",
		auth.RequireBearerPermission(
			authConfig,
			auth.PermissionShippingView,
			http.HandlerFunc(verifyCarrierManifestScanHandler(verifyCarrierManifestScan)),
		),
	)
	mux.Handle(
		"/api/v1/returns/receipts",
		auth.RequireBearerToken(
			authConfig,
			http.HandlerFunc(returnReceiptsHandler(listReturnReceipts, receiveReturn)),
		),
	)

	server := &http.Server{
		Addr:              ":" + cfg.AppPort,
		Handler:           accessLogMiddleware(mux, log.Default()),
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

func readinessHandler(w http.ResponseWriter, r *http.Request) {
	response.WriteSuccess(w, r, http.StatusOK, healthResponse{
		Status:    "ready",
		Service:   "api",
		Timestamp: time.Now().UTC().Format(time.RFC3339),
	})
}

type statusRecordingResponseWriter struct {
	http.ResponseWriter
	statusCode int
	bytes      int
}

func (w *statusRecordingResponseWriter) WriteHeader(statusCode int) {
	if w.statusCode != 0 {
		return
	}
	w.statusCode = statusCode
	w.ResponseWriter.WriteHeader(statusCode)
}

func (w *statusRecordingResponseWriter) Write(body []byte) (int, error) {
	if w.statusCode == 0 {
		w.statusCode = http.StatusOK
	}
	written, err := w.ResponseWriter.Write(body)
	w.bytes += written
	return written, err
}

func accessLogMiddleware(next http.Handler, logger *log.Logger) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		startedAt := time.Now()
		recorder := &statusRecordingResponseWriter{ResponseWriter: w}

		next.ServeHTTP(recorder, r)

		statusCode := recorder.statusCode
		if statusCode == 0 {
			statusCode = http.StatusOK
		}

		logger.Printf(
			"access method=%s path=%s status=%d bytes=%d duration_ms=%d remote=%s request_id=%s",
			r.Method,
			r.URL.Path,
			statusCode,
			recorder.bytes,
			time.Since(startedAt).Milliseconds(),
			r.RemoteAddr,
			r.Header.Get(response.HeaderRequestID),
		)
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

func endOfDayReconciliationsHandler(service inventoryapp.ListEndOfDayReconciliations) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			response.WriteError(w, r, http.StatusMethodNotAllowed, response.ErrorCodeNotFound, "Route not found", nil)
			return
		}

		filter := domain.NewEndOfDayReconciliationFilter(
			r.URL.Query().Get("warehouse_id"),
			r.URL.Query().Get("date"),
			domain.EndOfDayReconciliationStatus(r.URL.Query().Get("status")),
		)
		reconciliations, err := service.Execute(r.Context(), filter)
		if err != nil {
			response.WriteError(
				w,
				r,
				http.StatusConflict,
				response.ErrorCodeConflict,
				"End-of-day reconciliations could not be loaded",
				nil,
			)
			return
		}

		payload := make([]endOfDayReconciliationResponse, 0, len(reconciliations))
		for _, reconciliation := range reconciliations {
			payload = append(payload, newEndOfDayReconciliationResponse(reconciliation, ""))
		}

		response.WriteSuccess(w, r, http.StatusOK, payload)
	}
}

func closeEndOfDayReconciliationHandler(service inventoryapp.CloseEndOfDayReconciliation) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			response.WriteError(w, r, http.StatusMethodNotAllowed, response.ErrorCodeNotFound, "Route not found", nil)
			return
		}

		var payload closeReconciliationRequest
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			response.WriteError(
				w,
				r,
				http.StatusBadRequest,
				response.ErrorCodeValidation,
				"Invalid close reconciliation payload",
				nil,
			)
			return
		}

		principal, ok := auth.PrincipalFromContext(r.Context())
		if !ok {
			response.WriteError(w, r, http.StatusUnauthorized, response.ErrorCodeUnauthorized, "Authentication required", nil)
			return
		}

		result, err := service.Execute(r.Context(), inventoryapp.CloseEndOfDayReconciliationInput{
			ID:            r.PathValue("reconciliation_id"),
			ActorID:       principal.UserID,
			RequestID:     response.RequestID(r),
			ExceptionNote: payload.ExceptionNote,
		})
		if err != nil {
			writeCloseReconciliationError(w, r, err)
			return
		}

		response.WriteSuccess(
			w,
			r,
			http.StatusOK,
			newEndOfDayReconciliationResponse(result.Reconciliation, result.AuditLogID),
		)
	}
}

func carrierManifestsHandler(
	listService shippingapp.ListCarrierManifests,
	createService shippingapp.CreateCarrierManifest,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		principal, ok := auth.PrincipalFromContext(r.Context())
		if !ok {
			response.WriteError(w, r, http.StatusUnauthorized, response.ErrorCodeUnauthorized, "Authentication required", nil)
			return
		}

		switch r.Method {
		case http.MethodGet:
			if !auth.HasPermission(principal, auth.PermissionShippingView) {
				writePermissionDenied(w, r, auth.PermissionShippingView)
				return
			}
			filter := shippingdomain.NewCarrierManifestFilter(
				r.URL.Query().Get("warehouse_id"),
				r.URL.Query().Get("date"),
				r.URL.Query().Get("carrier_code"),
				shippingdomain.CarrierManifestStatus(r.URL.Query().Get("status")),
			)
			manifests, err := listService.Execute(r.Context(), filter)
			if err != nil {
				response.WriteError(
					w,
					r,
					http.StatusConflict,
					response.ErrorCodeConflict,
					"Carrier manifests could not be loaded",
					nil,
				)
				return
			}

			payload := make([]carrierManifestResponse, 0, len(manifests))
			for _, manifest := range manifests {
				payload = append(payload, newCarrierManifestResponse(manifest, ""))
			}
			response.WriteSuccess(w, r, http.StatusOK, payload)
		case http.MethodPost:
			if !auth.HasPermission(principal, auth.PermissionRecordCreate) {
				writePermissionDenied(w, r, auth.PermissionRecordCreate)
				return
			}
			var payload createCarrierManifestRequest
			if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
				response.WriteError(
					w,
					r,
					http.StatusBadRequest,
					response.ErrorCodeValidation,
					"Invalid carrier manifest payload",
					nil,
				)
				return
			}

			result, err := createService.Execute(r.Context(), shippingapp.CreateCarrierManifestInput{
				ID:            payload.ID,
				CarrierCode:   payload.CarrierCode,
				CarrierName:   payload.CarrierName,
				WarehouseID:   payload.WarehouseID,
				WarehouseCode: payload.WarehouseCode,
				Date:          payload.Date,
				HandoverBatch: payload.HandoverBatch,
				StagingZone:   payload.StagingZone,
				Owner:         payload.Owner,
				ActorID:       principal.UserID,
				RequestID:     response.RequestID(r),
			})
			if err != nil {
				writeCarrierManifestError(w, r, err)
				return
			}

			response.WriteSuccess(w, r, http.StatusCreated, newCarrierManifestResponse(result.Manifest, result.AuditLogID))
		default:
			response.WriteError(w, r, http.StatusMethodNotAllowed, response.ErrorCodeNotFound, "Route not found", nil)
		}
	}
}

func addShipmentToCarrierManifestHandler(service shippingapp.AddShipmentToCarrierManifest) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			response.WriteError(w, r, http.StatusMethodNotAllowed, response.ErrorCodeNotFound, "Route not found", nil)
			return
		}

		var payload addShipmentToManifestRequest
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			response.WriteError(
				w,
				r,
				http.StatusBadRequest,
				response.ErrorCodeValidation,
				"Invalid add shipment payload",
				nil,
			)
			return
		}

		principal, ok := auth.PrincipalFromContext(r.Context())
		if !ok {
			response.WriteError(w, r, http.StatusUnauthorized, response.ErrorCodeUnauthorized, "Authentication required", nil)
			return
		}
		result, err := service.Execute(r.Context(), shippingapp.AddShipmentToCarrierManifestInput{
			ManifestID: r.PathValue("manifest_id"),
			ShipmentID: payload.ShipmentID,
			ActorID:    principal.UserID,
			RequestID:  response.RequestID(r),
		})
		if err != nil {
			writeCarrierManifestError(w, r, err)
			return
		}

		response.WriteSuccess(w, r, http.StatusOK, newCarrierManifestResponse(result.Manifest, result.AuditLogID))
	}
}

func verifyCarrierManifestScanHandler(service shippingapp.VerifyCarrierManifestScan) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			response.WriteError(w, r, http.StatusMethodNotAllowed, response.ErrorCodeNotFound, "Route not found", nil)
			return
		}

		var payload verifyCarrierManifestScanRequest
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			response.WriteError(
				w,
				r,
				http.StatusBadRequest,
				response.ErrorCodeValidation,
				"Invalid carrier manifest scan payload",
				nil,
			)
			return
		}

		principal, ok := auth.PrincipalFromContext(r.Context())
		if !ok {
			response.WriteError(w, r, http.StatusUnauthorized, response.ErrorCodeUnauthorized, "Authentication required", nil)
			return
		}
		result, err := service.Execute(r.Context(), shippingapp.VerifyCarrierManifestScanInput{
			ManifestID: r.PathValue("manifest_id"),
			Code:       payload.Code,
			StationID:  payload.StationID,
			ActorID:    principal.UserID,
			RequestID:  response.RequestID(r),
		})
		if err != nil {
			writeCarrierManifestError(w, r, err)
			return
		}

		response.WriteSuccess(w, r, http.StatusOK, newCarrierManifestScanResponse(result))
	}
}

func returnReceiptsHandler(
	listService returnsapp.ListReturnReceipts,
	receiveService returnsapp.ReceiveReturn,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		principal, ok := auth.PrincipalFromContext(r.Context())
		if !ok {
			response.WriteError(w, r, http.StatusUnauthorized, response.ErrorCodeUnauthorized, "Authentication required", nil)
			return
		}

		switch r.Method {
		case http.MethodGet:
			if !auth.HasPermission(principal, auth.PermissionReturnsView) {
				writePermissionDenied(w, r, auth.PermissionReturnsView)
				return
			}
			filter := returnsdomain.NewReturnReceiptFilter(
				r.URL.Query().Get("warehouse_id"),
				returnsdomain.ReturnReceiptStatus(r.URL.Query().Get("status")),
			)
			receipts, err := listService.Execute(r.Context(), filter)
			if err != nil {
				response.WriteError(
					w,
					r,
					http.StatusConflict,
					response.ErrorCodeConflict,
					"Return receipts could not be loaded",
					nil,
				)
				return
			}

			payload := make([]returnReceiptResponse, 0, len(receipts))
			for _, receipt := range receipts {
				payload = append(payload, newReturnReceiptResponse(receipt, ""))
			}
			response.WriteSuccess(w, r, http.StatusOK, payload)
		case http.MethodPost:
			if !auth.HasPermission(principal, auth.PermissionRecordCreate) {
				writePermissionDenied(w, r, auth.PermissionRecordCreate)
				return
			}
			var payload receiveReturnRequest
			if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
				response.WriteError(
					w,
					r,
					http.StatusBadRequest,
					response.ErrorCodeValidation,
					"Invalid return receiving payload",
					nil,
				)
				return
			}

			result, err := receiveService.Execute(r.Context(), returnsapp.ReceiveReturnInput{
				WarehouseID:       payload.WarehouseID,
				WarehouseCode:     payload.WarehouseCode,
				Source:            payload.Source,
				ScanCode:          payload.Code,
				PackageCondition:  payload.PackageCondition,
				Disposition:       payload.Disposition,
				InvestigationNote: payload.InvestigationNote,
				ActorID:           principal.UserID,
				RequestID:         response.RequestID(r),
			})
			if err != nil {
				writeReturnReceiptError(w, r, err)
				return
			}

			response.WriteSuccess(w, r, http.StatusCreated, newReturnReceiptResponse(result.Receipt, result.AuditLogID))
		default:
			response.WriteError(w, r, http.StatusMethodNotAllowed, response.ErrorCodeNotFound, "Route not found", nil)
		}
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

func newEndOfDayReconciliationResponse(
	reconciliation domain.EndOfDayReconciliation,
	auditLogID string,
) endOfDayReconciliationResponse {
	summary := reconciliation.Summary("")
	payload := endOfDayReconciliationResponse{
		ID:            reconciliation.ID,
		WarehouseID:   reconciliation.WarehouseID,
		WarehouseCode: reconciliation.WarehouseCode,
		Date:          reconciliation.Date,
		ShiftCode:     reconciliation.ShiftCode,
		Status:        string(reconciliation.Status),
		Owner:         reconciliation.Owner,
		AuditLogID:    auditLogID,
		Summary: endOfDayReconciliationSummaryResponse{
			SystemQuantity:     summary.SystemQuantity,
			CountedQuantity:    summary.CountedQuantity,
			VarianceQuantity:   summary.VarianceQuantity,
			VarianceCount:      summary.VarianceCount,
			ChecklistTotal:     summary.ChecklistTotal,
			ChecklistCompleted: summary.ChecklistCompleted,
			ReadyToClose:       summary.ReadyToClose,
		},
		Checklist: make([]endOfDayReconciliationChecklistResponse, 0, len(reconciliation.Checklist)),
		Lines:     make([]endOfDayReconciliationLineResponse, 0, len(reconciliation.Lines)),
	}
	if !reconciliation.ClosedAt.IsZero() {
		payload.ClosedAt = reconciliation.ClosedAt.UTC().Format(time.RFC3339)
	}
	if strings.TrimSpace(reconciliation.ClosedBy) != "" {
		payload.ClosedBy = strings.TrimSpace(reconciliation.ClosedBy)
	}
	for _, item := range reconciliation.Checklist {
		payload.Checklist = append(payload.Checklist, endOfDayReconciliationChecklistResponse{
			Key:      item.Key,
			Label:    item.Label,
			Complete: item.Complete,
			Blocking: item.Blocking,
			Note:     item.Note,
		})
	}
	for _, line := range reconciliation.Lines {
		payload.Lines = append(payload.Lines, endOfDayReconciliationLineResponse{
			ID:               line.ID,
			SKU:              line.SKU,
			BatchNo:          line.BatchNo,
			BinCode:          line.BinCode,
			SystemQuantity:   line.SystemQuantity,
			CountedQuantity:  line.CountedQuantity,
			VarianceQuantity: line.VarianceQuantity(),
			Reason:           line.Reason,
			Owner:            line.Owner,
		})
	}

	return payload
}

func newCarrierManifestResponse(manifest shippingdomain.CarrierManifest, auditLogID string) carrierManifestResponse {
	summary := manifest.Summary()
	payload := carrierManifestResponse{
		ID:            manifest.ID,
		CarrierCode:   manifest.CarrierCode,
		CarrierName:   manifest.CarrierName,
		WarehouseID:   manifest.WarehouseID,
		WarehouseCode: manifest.WarehouseCode,
		Date:          manifest.Date,
		HandoverBatch: manifest.HandoverBatch,
		StagingZone:   manifest.StagingZone,
		Status:        string(manifest.Status),
		Owner:         manifest.Owner,
		AuditLogID:    auditLogID,
		Summary: carrierManifestSummaryResponse{
			ExpectedCount: summary.ExpectedCount,
			ScannedCount:  summary.ScannedCount,
			MissingCount:  summary.MissingCount,
		},
		Lines: make([]carrierManifestLineResponse, 0, len(manifest.Lines)),
	}
	if !manifest.CreatedAt.IsZero() {
		payload.CreatedAt = manifest.CreatedAt.UTC().Format(time.RFC3339)
	}
	for _, line := range manifest.Lines {
		payload.Lines = append(payload.Lines, carrierManifestLineResponse{
			ID:          line.ID,
			ShipmentID:  line.ShipmentID,
			OrderNo:     line.OrderNo,
			TrackingNo:  line.TrackingNo,
			PackageCode: line.PackageCode,
			StagingZone: line.StagingZone,
			Scanned:     line.Scanned,
		})
	}

	return payload
}

func newCarrierManifestScanResponse(result shippingapp.CarrierManifestScanResult) carrierManifestScanResponse {
	payload := carrierManifestScanResponse{
		ResultCode:         string(result.Code),
		Severity:           result.Severity,
		Message:            result.Message,
		ExpectedManifestID: result.ExpectedManifestID,
		ScanEvent: carrierManifestScanEventResponse{
			ID:                 result.Event.ID,
			ManifestID:         result.Event.ManifestID,
			ExpectedManifestID: result.Event.ExpectedManifestID,
			Code:               result.Event.Code,
			ResultCode:         string(result.Event.ResultCode),
			Severity:           result.Event.Severity,
			Message:            result.Event.Message,
			ShipmentID:         result.Event.ShipmentID,
			OrderNo:            result.Event.OrderNo,
			TrackingNo:         result.Event.TrackingNo,
			ActorID:            result.Event.ActorID,
			StationID:          result.Event.StationID,
			WarehouseID:        result.Event.WarehouseID,
			CarrierCode:        result.Event.CarrierCode,
			CreatedAt:          result.Event.CreatedAt.UTC().Format(time.RFC3339),
		},
		Manifest:   newCarrierManifestResponse(result.Manifest, ""),
		AuditLogID: result.AuditLogID,
	}
	if result.Line != nil {
		payload.Line = &carrierManifestLineResponse{
			ID:          result.Line.ID,
			ShipmentID:  result.Line.ShipmentID,
			OrderNo:     result.Line.OrderNo,
			TrackingNo:  result.Line.TrackingNo,
			PackageCode: result.Line.PackageCode,
			StagingZone: result.Line.StagingZone,
			Scanned:     result.Line.Scanned,
		}
	}

	return payload
}

func newReturnReceiptResponse(receipt returnsdomain.ReturnReceipt, auditLogID string) returnReceiptResponse {
	payload := returnReceiptResponse{
		ID:                receipt.ID,
		ReceiptNo:         receipt.ReceiptNo,
		WarehouseID:       receipt.WarehouseID,
		WarehouseCode:     receipt.WarehouseCode,
		Source:            string(receipt.Source),
		ReceivedBy:        receipt.ReceivedBy,
		ReceivedAt:        receipt.ReceivedAt.UTC().Format(time.RFC3339),
		PackageCondition:  receipt.PackageCondition,
		Status:            string(receipt.Status),
		Disposition:       string(receipt.Disposition),
		TargetLocation:    receipt.TargetLocation,
		OriginalOrderNo:   receipt.OriginalOrderNo,
		TrackingNo:        receipt.TrackingNo,
		ReturnCode:        receipt.ReturnCode,
		ScanCode:          receipt.ScanCode,
		CustomerName:      receipt.CustomerName,
		UnknownCase:       receipt.UnknownCase,
		Lines:             make([]returnReceiptLineResponse, 0, len(receipt.Lines)),
		InvestigationNote: receipt.InvestigationNote,
		AuditLogID:        auditLogID,
		CreatedAt:         receipt.CreatedAt.UTC().Format(time.RFC3339),
	}
	for _, line := range receipt.Lines {
		payload.Lines = append(payload.Lines, returnReceiptLineResponse{
			ID:          line.ID,
			SKU:         line.SKU,
			ProductName: line.ProductName,
			Quantity:    line.Quantity,
			Condition:   line.Condition,
		})
	}
	if receipt.StockMovement != nil {
		payload.StockMovement = &returnStockMovementResponse{
			ID:                receipt.StockMovement.ID,
			MovementType:      receipt.StockMovement.MovementType,
			SKU:               receipt.StockMovement.SKU,
			WarehouseID:       receipt.StockMovement.WarehouseID,
			Quantity:          receipt.StockMovement.Quantity,
			TargetStockStatus: receipt.StockMovement.TargetStockStatus,
			SourceDocID:       receipt.StockMovement.SourceDocID,
		}
	}

	return payload
}

func writeCloseReconciliationError(w http.ResponseWriter, r *http.Request, err error) {
	switch {
	case errors.Is(err, inventoryapp.ErrEndOfDayReconciliationNotFound):
		response.WriteError(
			w,
			r,
			http.StatusNotFound,
			response.ErrorCodeNotFound,
			"End-of-day reconciliation not found",
			nil,
		)
	case errors.Is(err, domain.ErrReconciliationAlreadyClosed):
		response.WriteError(
			w,
			r,
			http.StatusConflict,
			response.ErrorCodeConflict,
			"End-of-day reconciliation is already closed",
			nil,
		)
	case errors.Is(err, domain.ErrReconciliationNeedsExceptionNote):
		response.WriteError(
			w,
			r,
			http.StatusConflict,
			response.ErrorCodeConflict,
			"Exception note is required before closing this shift",
			map[string]any{"exception_note": "required"},
		)
	default:
		response.WriteError(
			w,
			r,
			http.StatusConflict,
			response.ErrorCodeConflict,
			"End-of-day reconciliation could not be closed",
			nil,
		)
	}
}

func writeCarrierManifestError(w http.ResponseWriter, r *http.Request, err error) {
	switch {
	case errors.Is(err, shippingapp.ErrCarrierManifestNotFound), errors.Is(err, shippingapp.ErrPackedShipmentNotFound):
		response.WriteError(w, r, http.StatusNotFound, response.ErrorCodeNotFound, "Carrier manifest resource not found", nil)
	case errors.Is(err, shippingdomain.ErrManifestRequiredField):
		response.WriteError(
			w,
			r,
			http.StatusBadRequest,
			response.ErrorCodeValidation,
			"Invalid carrier manifest payload",
			map[string]any{"required": "carrier_code, warehouse_id, and date"},
		)
	case errors.Is(err, shippingdomain.ErrManifestScanCodeRequired):
		response.WriteError(
			w,
			r,
			http.StatusBadRequest,
			response.ErrorCodeValidation,
			"Scan code is required",
			map[string]any{"required": "code"},
		)
	case errors.Is(err, shippingdomain.ErrManifestShipmentNotPacked):
		response.WriteError(w, r, http.StatusConflict, response.ErrorCodeConflict, "Shipment must be packed before adding to manifest", nil)
	case errors.Is(err, shippingdomain.ErrManifestDuplicateShipment):
		response.WriteError(w, r, http.StatusConflict, response.ErrorCodeConflict, "Shipment already exists in carrier manifest", nil)
	case errors.Is(err, shippingdomain.ErrManifestAlreadyCompleted):
		response.WriteError(w, r, http.StatusConflict, response.ErrorCodeConflict, "Carrier manifest is already completed", nil)
	default:
		response.WriteError(w, r, http.StatusConflict, response.ErrorCodeConflict, "Carrier manifest request could not be processed", nil)
	}
}

func writeReturnReceiptError(w http.ResponseWriter, r *http.Request, err error) {
	switch {
	case errors.Is(err, returnsdomain.ErrReturnReceiptScanCodeRequired):
		response.WriteError(
			w,
			r,
			http.StatusBadRequest,
			response.ErrorCodeValidation,
			"Return scan code is required",
			map[string]any{"required": "code"},
		)
	case errors.Is(err, returnsdomain.ErrReturnReceiptInvalidDisposition):
		response.WriteError(
			w,
			r,
			http.StatusBadRequest,
			response.ErrorCodeValidation,
			"Return disposition is invalid",
			map[string]any{"allowed": "reusable, not_reusable, needs_inspection"},
		)
	case errors.Is(err, returnsdomain.ErrReturnReceiptRequiredField):
		response.WriteError(
			w,
			r,
			http.StatusBadRequest,
			response.ErrorCodeValidation,
			"Invalid return receiving payload",
			map[string]any{"required": "warehouse_id"},
		)
	default:
		response.WriteError(w, r, http.StatusConflict, response.ErrorCodeConflict, "Return receipt could not be processed", nil)
	}
}

func writePermissionDenied(w http.ResponseWriter, r *http.Request, permission auth.PermissionKey) {
	response.WriteError(
		w,
		r,
		http.StatusForbidden,
		response.ErrorCodeForbidden,
		"Permission denied",
		map[string]any{"permission": string(permission)},
	)
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
