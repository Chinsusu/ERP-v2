package auth

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Chinsusu/ERP-v2/apps/api/internal/shared/response"
)

var testConfig = MockConfig{
	Email:       "admin@example.local",
	Password:    "local-only-mock-password",
	AccessToken: "local-dev-access-token",
}

func TestValidateMockLoginAcceptsSeedAccount(t *testing.T) {
	principal, ok := ValidateMockLogin(testConfig, "ADMIN@example.local", "local-only-mock-password")
	if !ok {
		t.Fatal("login rejected, want accepted")
	}
	if principal.UserID != "user-erp-admin" {
		t.Fatalf("user id = %q, want user-erp-admin", principal.UserID)
	}
	if principal.Role != RoleERPAdmin {
		t.Fatalf("role = %q, want %q", principal.Role, RoleERPAdmin)
	}
	if !HasPermission(principal, PermissionSettingsView) {
		t.Fatal("principal missing settings view permission")
	}
}

func TestValidateMockLoginRejectsWrongPassword(t *testing.T) {
	_, ok := ValidateMockLogin(testConfig, "admin@example.local", "wrong")
	if ok {
		t.Fatal("login accepted, want rejected")
	}
}

func TestRequireBearerTokenRejectsMissingToken(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/api/v1/me", nil)
	rec := httptest.NewRecorder()
	handler := RequireBearerToken(testConfig, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("handler called without token")
	}))

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusUnauthorized)
	}

	var payload response.ErrorEnvelope
	if err := json.NewDecoder(rec.Body).Decode(&payload); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if payload.Error.Code != response.ErrorCodeUnauthorized {
		t.Fatalf("code = %q, want %q", payload.Error.Code, response.ErrorCodeUnauthorized)
	}
}

func TestRequireBearerTokenAddsPrincipal(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/api/v1/me", nil)
	req.Header.Set("Authorization", "Bearer local-dev-access-token")
	rec := httptest.NewRecorder()
	handler := RequireBearerToken(testConfig, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		principal, ok := PrincipalFromContext(r.Context())
		if !ok {
			t.Fatal("principal missing from context")
		}
		if principal.Email != "admin@example.local" {
			t.Fatalf("email = %q, want admin@example.local", principal.Email)
		}
		w.WriteHeader(http.StatusNoContent)
	}))

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusNoContent)
	}
}

func TestRoleCatalogIncludesPhaseOneRoles(t *testing.T) {
	roles := RoleCatalog()
	got := make(map[RoleKey]bool, len(roles))
	for _, role := range roles {
		got[role.Key] = true
	}

	for _, role := range []RoleKey{
		RoleCEO,
		RoleERPAdmin,
		RoleWarehouseStaff,
		RoleWarehouseLead,
		RoleQA,
		RoleSalesOps,
		RoleProductionOps,
	} {
		if !got[role] {
			t.Fatalf("role %q missing from catalog", role)
		}
	}
}

func TestRequirePermissionRejectsMissingPermission(t *testing.T) {
	principal := MockPrincipalForRole(testConfig, RoleWarehouseStaff)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/rbac/roles", nil)
	req = req.WithContext(WithPrincipal(req.Context(), principal))
	rec := httptest.NewRecorder()
	handler := RequirePermission(PermissionSettingsView, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("handler called without permission")
	}))

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusForbidden)
	}

	var payload response.ErrorEnvelope
	if err := json.NewDecoder(rec.Body).Decode(&payload); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if payload.Error.Code != response.ErrorCodeForbidden {
		t.Fatalf("code = %q, want %q", payload.Error.Code, response.ErrorCodeForbidden)
	}
}

func TestRequirePermissionAllowsGrantedPermission(t *testing.T) {
	principal := MockPrincipalForRole(testConfig, RoleERPAdmin)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/rbac/roles", nil)
	req = req.WithContext(WithPrincipal(req.Context(), principal))
	rec := httptest.NewRecorder()
	handler := RequirePermission(PermissionSettingsView, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusNoContent)
	}
}
