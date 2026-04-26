package auth

import (
	"context"
	"net/http"
	"strings"

	"github.com/Chinsusu/ERP-v2/apps/api/internal/shared/response"
)

const authorizationHeader = "Authorization"

type MockConfig struct {
	Email       string
	Password    string
	AccessToken string
}

type Principal struct {
	UserID      string
	Email       string
	Name        string
	Role        RoleKey
	Permissions []PermissionKey
}

type principalContextKey struct{}

func ValidateMockLogin(cfg MockConfig, email string, password string) (Principal, bool) {
	if !strings.EqualFold(strings.TrimSpace(email), strings.TrimSpace(cfg.Email)) {
		return Principal{}, false
	}
	if password != cfg.Password {
		return Principal{}, false
	}

	return MockPrincipal(cfg), true
}

func MockPrincipal(cfg MockConfig) Principal {
	return MockPrincipalForRole(cfg, RoleERPAdmin)
}

func MockPrincipalForRole(cfg MockConfig, role RoleKey) Principal {
	return Principal{
		UserID:      "user-" + strings.ToLower(strings.ReplaceAll(string(role), "_", "-")),
		Email:       cfg.Email,
		Name:        RoleDisplayName(role),
		Role:        role,
		Permissions: PermissionsForRole(role),
	}
}

func RequireBearerToken(cfg MockConfig, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := bearerToken(r.Header.Get(authorizationHeader))
		if token == "" || token != cfg.AccessToken {
			response.WriteError(
				w,
				r,
				http.StatusUnauthorized,
				response.ErrorCodeUnauthorized,
				"Authentication required",
				nil,
			)
			return
		}

		ctx := WithPrincipal(r.Context(), MockPrincipal(cfg))
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func WithPrincipal(ctx context.Context, principal Principal) context.Context {
	return context.WithValue(ctx, principalContextKey{}, principal)
}

func PrincipalFromContext(ctx context.Context) (Principal, bool) {
	principal, ok := ctx.Value(principalContextKey{}).(Principal)
	return principal, ok
}

func bearerToken(headerValue string) string {
	const prefix = "Bearer "

	if !strings.HasPrefix(headerValue, prefix) {
		return ""
	}

	return strings.TrimSpace(strings.TrimPrefix(headerValue, prefix))
}
