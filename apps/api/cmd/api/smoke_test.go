package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	shippingapp "github.com/Chinsusu/ERP-v2/apps/api/internal/modules/shipping/application"
	"github.com/Chinsusu/ERP-v2/apps/api/internal/shared/audit"
	"github.com/Chinsusu/ERP-v2/apps/api/internal/shared/auth"
	"github.com/Chinsusu/ERP-v2/apps/api/internal/shared/response"
)

func TestSprint0APISmokePack(t *testing.T) {
	authConfig := smokeAuthConfig()

	t.Run("health and readiness", func(t *testing.T) {
		for _, tc := range []struct {
			name    string
			path    string
			handler http.HandlerFunc
			want    string
		}{
			{name: "health", path: "/healthz", handler: healthHandler, want: "ok"},
			{name: "readiness", path: "/readyz", handler: readinessHandler, want: "ready"},
		} {
			t.Run(tc.name, func(t *testing.T) {
				rec := httptest.NewRecorder()
				req := httptest.NewRequest(http.MethodGet, tc.path, nil)

				tc.handler(rec, req)

				if rec.Code != http.StatusOK {
					t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
				}
				payload := decodeSmokeSuccess[healthResponse](t, rec)
				if payload.Data.Status != tc.want {
					t.Fatalf("status = %q, want %q", payload.Data.Status, tc.want)
				}
			})
		}
	})

	t.Run("login and me", func(t *testing.T) {
		loginBody := bytes.NewBufferString(`{"email":"admin@example.local","password":"local-only-mock-password"}`)
		loginReq := httptest.NewRequest(http.MethodPost, "/api/v1/auth/mock-login", loginBody)
		loginRec := httptest.NewRecorder()

		mockLoginHandler(authConfig).ServeHTTP(loginRec, loginReq)

		if loginRec.Code != http.StatusOK {
			t.Fatalf("login status = %d, want %d: %s", loginRec.Code, http.StatusOK, loginRec.Body.String())
		}
		loginPayload := decodeSmokeSuccess[loginResponse](t, loginRec)
		if loginPayload.Data.AccessToken == "" {
			t.Fatal("access token is empty")
		}

		meReq := smokeRequestAsRole(httptest.NewRequest(http.MethodGet, "/api/v1/me", nil), authConfig, auth.RoleERPAdmin)
		meRec := httptest.NewRecorder()

		meHandler(meRec, meReq)

		if meRec.Code != http.StatusOK {
			t.Fatalf("me status = %d, want %d: %s", meRec.Code, http.StatusOK, meRec.Body.String())
		}
		mePayload := decodeSmokeSuccess[userResponse](t, meRec)
		if mePayload.Data.Email != "admin@example.local" || mePayload.Data.Role != string(auth.RoleERPAdmin) {
			t.Fatalf("me payload = %+v, want ERP admin", mePayload.Data)
		}
	})

	t.Run("master data permission is in role catalog", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/rbac/roles", nil)
		rec := httptest.NewRecorder()

		rbacRolesHandler(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("status = %d, want %d: %s", rec.Code, http.StatusOK, rec.Body.String())
		}
		payload := decodeSmokeSuccess[[]roleResponse](t, rec)
		for _, role := range payload.Data {
			if role.Key == string(auth.RoleERPAdmin) && smokeContains(role.Permissions, string(auth.PermissionMasterDataView)) {
				return
			}
		}
		t.Fatalf("ERP_ADMIN role catalog does not include %s", auth.PermissionMasterDataView)
	})

	t.Run("stock movement adjustment writes audit", func(t *testing.T) {
		auditStore := audit.NewInMemoryLogStore()
		body := bytes.NewBufferString(`{
			"movementId": "mov-smoke-adjust",
			"sku": "serum-30ml",
			"warehouseId": "wh-hcm",
			"movementType": "ADJUST",
			"quantity": "8.000000",
			"baseUomCode": "PCS",
			"reason": "smoke cycle count"
		}`)
		req := smokeRequestAsRole(
			httptest.NewRequest(http.MethodPost, "/api/v1/inventory/stock-movements", body),
			authConfig,
			auth.RoleERPAdmin,
		)
		req.Header.Set(response.HeaderRequestID, "req-smoke-stock")
		rec := httptest.NewRecorder()

		stockMovementHandler(auditStore).ServeHTTP(rec, req)

		if rec.Code != http.StatusCreated {
			t.Fatalf("status = %d, want %d: %s", rec.Code, http.StatusCreated, rec.Body.String())
		}
		logs, err := auditStore.List(req.Context(), audit.Query{EntityID: "mov-smoke-adjust"})
		if err != nil {
			t.Fatalf("list audit logs: %v", err)
		}
		if len(logs) != 1 || logs[0].Action != "inventory.stock_movement.adjusted" {
			t.Fatalf("audit logs = %+v, want one stock movement adjustment", logs)
		}
	})

	t.Run("scan handover marks manifest matched", func(t *testing.T) {
		manifestStore := shippingapp.NewPrototypeCarrierManifestStore()
		auditStore := audit.NewInMemoryLogStore()
		service := shippingapp.NewVerifyCarrierManifestScan(manifestStore, auditStore)
		body := bytes.NewBufferString(`{"code":"GHN260426003","station_id":"dock-a"}`)
		req := smokeRequestAsRole(
			httptest.NewRequest(http.MethodPost, "/api/v1/shipping/manifests/manifest-hcm-ghn-morning/scan", body),
			authConfig,
			auth.RoleWarehouseStaff,
		)
		req.SetPathValue("manifest_id", "manifest-hcm-ghn-morning")
		req.Header.Set(response.HeaderRequestID, "req-smoke-scan")
		rec := httptest.NewRecorder()

		verifyCarrierManifestScanHandler(service).ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("status = %d, want %d: %s", rec.Code, http.StatusOK, rec.Body.String())
		}
		payload := decodeSmokeSuccess[carrierManifestScanResponse](t, rec)
		if payload.Data.ResultCode != "MATCHED" || payload.Data.Manifest.Summary.MissingCount != 0 || payload.Data.AuditLogID == "" {
			t.Fatalf("scan payload = %+v, want matched handover with audit", payload.Data)
		}
	})
}

func smokeAuthConfig() auth.MockConfig {
	return auth.MockConfig{
		Email:       "admin@example.local",
		Password:    "local-only-mock-password",
		AccessToken: "local-dev-access-token",
	}
}

func smokeRequestAsRole(req *http.Request, authConfig auth.MockConfig, role auth.RoleKey) *http.Request {
	return req.WithContext(auth.WithPrincipal(req.Context(), auth.MockPrincipalForRole(authConfig, role)))
}

func decodeSmokeSuccess[T any](t *testing.T, rec *httptest.ResponseRecorder) response.SuccessEnvelope[T] {
	t.Helper()

	var payload response.SuccessEnvelope[T]
	if err := json.NewDecoder(rec.Body).Decode(&payload); err != nil {
		t.Fatalf("decode response: %v", err)
	}

	return payload
}

func smokeContains(values []string, want string) bool {
	for _, value := range values {
		if value == want {
			return true
		}
	}
	return false
}
