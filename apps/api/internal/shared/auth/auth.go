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
	Users       []LoginUser
}

type LoginUser struct {
	Email    string
	Password string
	Name     string
	Role     RoleKey
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
	admin := LoginUser{
		Email:    cfg.Email,
		Password: cfg.Password,
		Name:     RoleDisplayName(RoleERPAdmin),
		Role:     RoleERPAdmin,
	}
	if principal, ok := validateLoginUser(admin, email, password, cfg.Password); ok {
		return principal, true
	}

	for _, user := range cfg.Users {
		if principal, ok := validateLoginUser(user, email, password, cfg.Password); ok {
			return principal, true
		}
	}

	return Principal{}, false
}

func MockPrincipal(cfg MockConfig) Principal {
	return MockPrincipalForRole(cfg, RoleERPAdmin)
}

func MockPrincipalForRole(cfg MockConfig, role RoleKey) Principal {
	return principalForLoginUser(LoginUser{
		Email: cfg.Email,
		Name:  RoleDisplayName(role),
		Role:  role,
	})
}

func validateLoginUser(user LoginUser, email string, password string, fallbackPassword string) (Principal, bool) {
	if !strings.EqualFold(strings.TrimSpace(email), strings.TrimSpace(user.Email)) {
		return Principal{}, false
	}
	expectedPassword := user.Password
	if expectedPassword == "" {
		expectedPassword = fallbackPassword
	}
	if password != expectedPassword {
		return Principal{}, false
	}

	return principalForLoginUser(user), true
}

func principalForLoginUser(user LoginUser) Principal {
	role := user.Role
	if role == "" {
		role = RoleERPAdmin
	}
	name := strings.TrimSpace(user.Name)
	if name == "" {
		name = RoleDisplayName(role)
	}

	return Principal{
		UserID:      "user-" + strings.ToLower(strings.ReplaceAll(string(role), "_", "-")),
		Email:       strings.TrimSpace(user.Email),
		Name:        name,
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

func RequireSessionToken(sessions *SessionManager, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := bearerToken(r.Header.Get(authorizationHeader))
		principal, ok := sessions.AuthenticateAccessToken(token)
		if token == "" || !ok {
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

		ctx := WithPrincipal(r.Context(), principal)
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
