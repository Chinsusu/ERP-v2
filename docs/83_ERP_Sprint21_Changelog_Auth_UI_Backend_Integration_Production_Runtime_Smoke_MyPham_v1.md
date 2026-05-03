# 83_ERP_Sprint21_Changelog_Auth_UI_Backend_Integration_Production_Runtime_Smoke_MyPham_v1

Project: Web ERP for cosmetics operations
Phase: Phase 1
Sprint: Sprint 21 - Auth UI Backend Integration + Production Runtime Smoke
Document role: Sprint changelog and release evidence
Version: v1.0
Date: 2026-05-03
Status: Completed and merged

---

## 1. Sprint 21 Scope

Sprint 21 closes the Sprint 20 production-readiness caveat that the Vietnamese web auth UI was still mock-backed.

In scope:

```text
Backend logout contract and session revocation
OpenAPI auth/logout alignment
Frontend login action wired to backend auth/login
Backend-backed session bootstrap through /me
Access-token use for frontend API services without localStorage token storage
Refresh rotation through backend auth/refresh via a same-origin web route
Logout action that revokes backend refresh session and clears httpOnly cookies
Multi-tab logout/session expiry signal without storing secrets
Vietnamese auth error copy for invalid credentials, lockout, session expiry, and forbidden access
Production-like mock auth blocking and dev/staging mock default cleanup
Dev smoke auth pack extended with login, me, refresh, logout, and lockout checks
```

Out of scope:

```text
SSO/SAML/OIDC
MFA/2FA
Password reset email flow
Role redesign
Device/session management dashboard
Advanced auth audit export
```

---

## 2. Implementation Summary

Backend:

```text
POST /api/v1/auth/logout revokes a refresh session.
Logout invalidates both the refresh token and the paired access token.
Invalid or expired refresh tokens return 401.
PostgreSQL session store supports refresh-token revocation by token hash.
```

Frontend:

```text
Login form now calls backend /auth/login.
ERP layout bootstraps the authenticated user from /me using httpOnly access-token cookies.
Sidebar and route access use backend-backed user roles and permissions.
App shell keeps only the short-lived access token in memory for browser API calls.
Same-origin /api/auth/refresh rotates cookies and returns the next access token to memory.
Logout calls backend /auth/logout, clears cookies, and redirects to /login.
Tabs receive a non-secret logout timestamp signal and leave privileged UI.
```

Runtime:

```text
Dev and staging compose defaults now use NEXT_PUBLIC_MOCK_AUTH_STATE=signed-out.
Production-like runtime rejects forced signed-in mock auth.
Static local-dev access token remains allowed only for explicit local/dev/test app environments.
```

---

## 3. Verification Evidence

Local checks already run on the Sprint 21 branch:

```text
go fmt ./cmd/api ./internal/shared/auth
go test ./internal/shared/auth ./cmd/api -run 'Logout|Refresh|LoginHandler|AuthPolicy' -count=1
go test ./...
go vet ./...
sh -n infra/scripts/smoke-dev-full.sh
sh -n infra/scripts/smoke-dev-staging.sh
git diff --check
```

Local web checks were not runnable in this workstation session:

```text
node.exe failed with Access is denied.
pnpm was not available on PATH.
npm was not available on PATH.
docker was not available on PATH.
make was not available on PATH.
Local tsc/vitest, OpenAPI lint/contract, and make targets are therefore deferred to GitHub Actions.
```

Cloud PR evidence:

```text
PR number: #542
Title: Wire web auth UI to backend sessions
Manual merge: completed
Main merge commit: c07409ccecb513e3c0089311dc303ff4a2a390a2
required-ci run 25270151483: success
required-api job 74091067339: success
required-web job 74091067341: success
required-openapi job 74091067343: success
required-migration job 74091067340: success
api-ci run 25270151488: success
web-ci run 25270151472: success
openapi-ci run 25270151469: success
e2e-ci run 25270151470: success
```

Post-merge dev evidence:

```text
Dev deploy: ./infra/scripts/deploy-dev-staging.sh dev passed
Dev runtime smoke: login, /me before restart, /me after restart, refresh rotate, old refresh reject, /me after refresh, logout, /me after logout, refresh after logout, and lockout passed
Auth UI browser smoke: login dashboard, logout to login, and invalid login Vietnamese error passed
UI smoke method: Chrome headless CDP fallback because local Playwright CLI/npx was unavailable
Screenshot evidence: output/playwright/dev-auth-dashboard.png
Screenshot evidence: output/playwright/dev-auth-logout-login.png
Screenshot evidence: output/playwright/dev-auth-invalid-login.png
Dev server health: http://10.1.1.120:8088/health returned 200
```

Release tag status:

```text
Tag hold.
No v0.21.0-auth-ui-backend-integration-runtime-smoke tag has been created.
Reason: Sprint 21 has merged main, CI, dev deploy, full dev smoke, and auth UI browser smoke evidence, but production-like release tagging still requires target staging/pilot environment smoke evidence.
```

---

## 4. Known Limits

Sprint 21 makes the existing email/password web auth surface backend-backed. It does not claim enterprise identity completeness.

Verification limits:

```text
Local workstation was missing node/pnpm/make/docker/npx.
Web checks, OpenAPI checks, and build confidence were verified through GitHub Actions.
Auth UI browser smoke used Chrome headless CDP instead of Playwright CLI.
```

Known non-goals remain:

```text
No SSO/SAML/OIDC.
No MFA/2FA.
No password reset email flow.
No user invitation flow.
No device/session management dashboard.
```

---

## 5. Source Documents

| Area | Document |
| --- | --- |
| Sprint 21 task board | `82_ERP_Coding_Task_Board_Sprint21_Auth_UI_Backend_Integration_Production_Runtime_Smoke_MyPham_v1.md` |
| Production runtime checklist | `78_ERP_Production_Runtime_Mode_Checklist_Sprint20_MyPham_v1.md` |
| Sprint 20 changelog | `79_ERP_Sprint20_Changelog_Release_Hygiene_API_Modularization_Fallback_Cleanup_MyPham_v1.md` |
| Current master document index | `80_ERP_Master_Document_Index_Current_Status_MyPham_v2.md` |
| Vietnamese operational glossary | `81_ERP_Vietnamese_UI_Glossary_Operational_Copy_MyPham_v1.md` |
