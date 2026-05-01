# S9-02-01 Auth / Session Policy Inventory

Project: Web ERP for cosmetics operations
Sprint: Sprint 9 - System hardening / production readiness core
Task: S9-02-01 Auth/session policy inventory
Date: 2026-05-01
Status: Inventory complete; hardening targets identified

---

## 1. Purpose

This inventory records the current auth/session implementation before Sprint 9 changes login/session behavior.

The goal is to avoid hardening by guesswork. S9-02-02 should change the highest-risk behavior with a small, testable diff.

---

## 2. Current Backend Behavior

Source files:

```text
apps/api/internal/shared/auth/auth.go
apps/api/internal/shared/auth/session.go
apps/api/internal/shared/auth/rbac.go
apps/api/internal/shared/config/config.go
apps/api/cmd/api/main.go
```

Current API endpoints:

```text
POST /api/v1/auth/login
POST /api/v1/auth/mock-login
POST /api/v1/auth/refresh
GET  /api/v1/auth/policy
GET  /api/v1/me
```

Current session model:

```text
- Login validates one configured local mock account.
- Access tokens are random local-at-* strings with an 8 hour TTL.
- Refresh tokens are random local-rt-* strings with a 7 day TTL.
- Refresh rotates both access and refresh tokens.
- Old access and refresh tokens are removed during refresh.
- Failed login state is in memory.
- Session state is in memory.
- A configured static access token is seeded at startup for local/dev automation.
```

Current password and lockout policy:

```text
- Minimum password length: 10.
- Must include at least one letter.
- Must include at least one number or symbol.
- Common weak passwords are blocked.
- Lock after 5 failed attempts.
- Lockout window: 15 minutes.
- Lockout duration: 15 minutes.
```

Current audit behavior:

```text
- auth.login_succeeded is recorded on successful login.
- auth.login_failed is recorded on failed login.
- auth.refresh_succeeded is recorded on successful refresh.
- auth.refresh_failed is recorded on failed refresh.
```

Current route protection:

```text
- Protected API routes use RequireSessionToken.
- Permission-sensitive routes wrap RequireSessionToken with RBAC checks.
- /api/v1/health, /healthz, /api/v1/auth/login, /api/v1/auth/mock-login, /api/v1/auth/refresh, and /api/v1/auth/policy are public endpoints.
```

---

## 3. Current Frontend Behavior

Source files:

```text
apps/web/src/shared/auth/mockSession.ts
apps/web/src/shared/auth/sessionPolicy.ts
apps/web/src/app/(auth)/login/actions.ts
apps/web/src/app/(erp)/layout.tsx
apps/web/src/app/(erp)/[module]/page.tsx
apps/web/src/shared/api/client.ts
```

Current UI session model:

```text
- Login server action validates credentials locally in the Next.js app.
- The web session is stored in an httpOnly cookie named erp_mock_session.
- Cookie value contains an access token and expiresAt timestamp.
- Cookie max age is 8 hours.
- Cookie uses sameSite=lax.
- Cookie secure flag depends on AUTH_COOKIE_SECURE=true.
- ERP layout requires the mock session cookie before rendering app pages.
- Module pages also check menu permissions from the mock user role.
```

Current frontend API behavior:

```text
- Most frontend service modules use a hardcoded default local dev access token.
- API client sends Authorization: Bearer <token> when an access token is passed.
- Frontend service calls do not consistently read the server-side cookie token.
- The UI session and service-layer API token are therefore not fully unified.
```

---

## 4. Strengths Already Present

```text
1. Backend has real random issued access/refresh tokens for login.
2. Refresh token rotation is implemented.
3. Expired access tokens are rejected.
4. Password policy and lockout policy exist on backend and frontend.
5. Login and refresh outcomes write audit events.
6. Protected backend routes use session token auth, not only UI gating.
7. Permission wrappers exist for finance/reporting/export and operational actions.
8. Frontend ERP pages are denied when no mock session cookie is present.
```

---

## 5. Gaps / Risks

| ID | Risk | Impact | Evidence | Priority |
| --- | --- | --- | --- | --- |
| AUTH-01 | Static local access token is seeded at API startup | Anyone with the token has ERP admin-equivalent API access in dev-like environments | `SessionManager.seedStaticAccessToken` and `AUTH_MOCK_ACCESS_TOKEN` default | High |
| AUTH-02 | Frontend services use hardcoded default token instead of session token | API calls can succeed even if UI session handling is not representative of production auth | `defaultAccessToken = "local-dev-access-token"` across web service modules | High |
| AUTH-03 | Backend sessions and failed login state are in memory | Restart clears issued sessions and lockout state | `SessionManager` maps in memory | Medium |
| AUTH-04 | Refresh tokens are bearer strings returned to clients, not cookie-bound | Token storage policy is not hardened for browser production use | `/api/v1/auth/refresh` accepts refresh token in JSON body | Medium |
| AUTH-05 | `/api/v1/auth/mock-login` is wired to the same handler as login | Mock endpoint should stay explicitly dev-only or be removed before production | `main.go` route registration | Medium |
| AUTH-06 | Frontend cookie stores a token payload without signing/encryption | httpOnly helps, but tamper resistance depends on parse/expiry only | `mockSession.ts` parses cookie JSON | Medium |
| AUTH-07 | Cookie secure flag is environment-driven and can be false | Correct for local HTTP, unsafe if production deploy forgets secure flag | `AUTH_COOKIE_SECURE` controls cookie secure setting | Medium |
| AUTH-08 | No logout/revoke endpoint is documented in current API inventory | User cannot explicitly revoke access/refresh tokens from API | No logout route found in current auth route list | Medium |

---

## 6. Recommended S9-02-02 Scope

Highest-value next task:

```text
Tighten local static token behavior without breaking dev login.
```

Recommended small implementation:

```text
1. Keep issued login/refresh tokens working.
2. Make static AUTH_MOCK_ACCESS_TOKEN seeding explicitly dev/local only.
3. Fail startup or skip static token seeding when APP_ENV is production-like.
4. Add tests for static token seeding allowed in local/dev and rejected/skipped in production-like env.
5. Keep full dev smoke working through local dev configuration.
```

Why this is first:

```text
- It reduces the biggest backend auth shortcut.
- It is smaller than rewriting all frontend services to consume session tokens.
- It has a clean verification path.
- It does not remove current dev login.
```

Follow-up after S9-02-02:

```text
S9-02-03 should add route permission regression for pages/report tabs.
Later hardening should unify frontend service tokens with the cookie/session token instead of hardcoded defaults.
```

---

## 7. Verification Notes

Inventory checks performed:

```text
- Inspected backend auth/session/config route code.
- Inspected frontend mock session, local policy, and API token usage.
- Checked security standard file 19 for target expectations.
- Confirmed current full dev smoke includes login and protected dashboard/report endpoints.
```

No runtime behavior changed in this task.
