# 83_ERP_Sprint21_Changelog_Auth_UI_Backend_Integration_Production_Runtime_Smoke_MyPham_v1

Project: Web ERP for cosmetics operations
Phase: Phase 1
Sprint: Sprint 21 - Auth UI Backend Integration + Production Runtime Smoke
Document role: Sprint changelog and release evidence
Version: v1.0
Date: 2026-05-03
Status: PR CI passed; manual merge and post-merge dev smoke pending

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
required-ci run 25270108529: success
required-api job 74090950319: success
required-web job 74090950311: success
required-openapi job 74090950316: success
required-migration job 74090950320: success
api-ci run 25270108518: success
web-ci run 25270108527: success
openapi-ci run 25270108541: success
e2e-ci run 25270108517: success
```

Post-merge evidence still required by the release flow:

```text
Manual merge: pending
Dev deploy: pending
Auth UI browser smoke: pending
```

---

## 4. Known Limits

Sprint 21 makes the existing email/password web auth surface backend-backed. It does not claim enterprise identity completeness.

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
