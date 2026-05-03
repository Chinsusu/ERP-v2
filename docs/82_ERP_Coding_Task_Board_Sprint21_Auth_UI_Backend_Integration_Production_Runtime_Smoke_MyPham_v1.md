# 82_ERP_Coding_Task_Board_Sprint21_Auth_UI_Backend_Integration_Production_Runtime_Smoke_MyPham_v1

Project: Web ERP for cosmetics operations
Phase: Phase 1
Sprint: Sprint 21 - Auth UI Backend Integration + Production Runtime Smoke
Document role: Coding task board for integrating the Vietnamese frontend auth surface with backend auth/session APIs and proving production-like runtime readiness.

---

## 1. Sprint Goal

Sprint 21 turns the current Vietnamese-first ERP frontend from a mock/staging auth surface into an explicit backend-backed auth flow suitable for production-like environments.

The sprint goal is:

```text
Frontend login UI
-> backend login API
-> backend session/me check
-> refresh rotation
-> logout
-> failed login / lockout policy
-> route/RBAC guard
-> no silent mock auth in production-like runtime
-> production runtime smoke pack
```

This sprint does **not** add a new business workflow. It closes the production-readiness caveat recorded after Sprint 20:

```text
Web auth UI is still mock/staging-only until wired to the backend auth/session API.
```

---

## 2. Current Context

The current documentation source-of-truth says:

```text
Current main: Sprint 20 hardening completed after Sprint 19 Vietnamese UI localization.
Latest release tag: v0.19.0-vietnamese-ui-localization.
Business display: Vietnamese-first.
Technical contract: English.
Routes: English.
Locale: vi-VN.
Currency: VND.
Timezone: Asia/Ho_Chi_Minh.
```

Sprint 20 hardened release evidence, migration gates, API route modularization, Node.js 24 compatibility, and production fallback blocking.

Remaining production-readiness caveat:

```text
Backend auth/session persistence exists from Sprint 18.
Frontend auth UI is still mock/staging-only until backend auth/session API wiring is explicit.
```

Sprint 21 exists to remove that caveat.

---

## 3. Primary Design Decision

### 3.1 Keep technical contracts in English

Do not rename:

```text
API routes
OpenAPI schemas
DB enum values
Permission keys
Audit event codes
Backend error codes
Workflow event names
```

Examples that remain technical English:

```text
/auth/login
/auth/me
/auth/refresh
/auth/logout
INVALID_CREDENTIALS
AUTH_SESSION_EXPIRED
AUTH_ACCOUNT_LOCKED
permission:inventory:view
```

### 3.2 Keep frontend business display Vietnamese-first

User-facing copy must use Vietnamese operational UI language:

```text
Đăng nhập
Đăng xuất
Phiên đăng nhập đã hết hạn
Tài khoản tạm khóa do đăng nhập sai nhiều lần
Bạn không có quyền truy cập màn hình này
```

### 3.3 Do not localize routes

Keep existing technical routes:

```text
/dashboard
/inventory
/sales/orders
/shipping/manifests
/returns
/purchase
/qc
/finance
/settings
```

Do not change to Vietnamese routes.

### 3.4 Production-like mode must not use mock auth

Local/no-DB development may keep mock auth for developer velocity.

Production-like runtime must fail loudly if backend auth is unavailable.

```text
APP_ENV=staging/prod/production
NODE_ENV=production
NEXT_PUBLIC_MOCK_AUTH_STATE must not silently authenticate users
Static auth tokens are not allowed outside local/dev/test
```

---

## 4. Source-of-Truth References

| Area | Primary source |
|---|---|
| Current document index and source-of-truth rules | `80_ERP_Master_Document_Index_Current_Status_MyPham_v2.md` |
| Production runtime caveat and smoke checklist | `78_ERP_Production_Runtime_Mode_Checklist_Sprint20_MyPham_v1.md` |
| Sprint 20 hardening evidence | `79_ERP_Sprint20_Changelog_Release_Hygiene_API_Modularization_Fallback_Cleanup_MyPham_v1.md` |
| Vietnamese UI glossary | `81_ERP_Vietnamese_UI_Glossary_Operational_Copy_MyPham_v1.md` |
| Sprint 19 localization board | `75_ERP_Coding_Task_Board_Sprint19_Vietnamese_UI_Localization_MyPham_v1.md` |
| Frontend architecture | `15_ERP_Frontend_Architecture_React_NextJS_Phase1_MyPham_v1.md` |
| API contract standard | `16_ERP_API_Contract_OpenAPI_Standards_Phase1_MyPham_v1.md` |
| Security/RBAC/audit standard | `19_ERP_Security_RBAC_Audit_Compliance_Standards_Phase1_MyPham_v1.md` |
| QA/test strategy | `24_ERP_QA_Test_Strategy_Automation_Phase1_MyPham_v1.md` |
| Workspace structure | `38_ERP_Workspace_Repository_Structure_Standards_Phase1_MyPham_v1.md` |

---

## 5. Sprint Scope

### In scope

```text
Backend auth API contract review
Frontend auth API client wiring
Login form wired to backend
Session bootstrap through /auth/me or equivalent
Refresh rotation integration
Logout integration
Failed login and lockout UI
Route guard and RBAC guard hardening
No mock auth in production-like runtime
Vietnamese auth messages and error mapping
Production runtime smoke pack
README/runtime caveat update after completion
Sprint 21 changelog creation
```

### Out of scope

```text
New business workflows
Route localization
Backend enum renaming
Database value renaming
Permission key renaming
Audit event renaming
Full enterprise SSO/SAML/OIDC
2FA/MFA unless already available in backend
Password reset by email unless already available in backend
Marketing website login
Role redesign beyond existing RBAC contract
```

---

## 6. Branch and Release Naming

Suggested branch:

```bash
git checkout main
git pull origin main
git checkout -b sprint/21-auth-ui-backend-integration-runtime-smoke
```

Suggested tag after acceptance:

```bash
git tag v0.21.0-auth-ui-backend-integration-runtime-smoke
git push origin v0.21.0-auth-ui-backend-integration-runtime-smoke
```

If Sprint 20 remains intentionally untagged as hardening, Sprint 21 may be the next explicit release tag after `v0.19.0-vietnamese-ui-localization`.

---

## 7. Runtime Auth Principles

### 7.1 Auth/session flow

Target flow:

```text
User opens web app
-> app checks existing session
-> if no valid session, redirect to login
-> user submits credentials
-> backend validates credentials
-> backend creates session/refresh state
-> frontend stores/uses session according to approved transport
-> app loads /auth/me
-> RBAC menus/routes render by permission
-> refresh rotates when needed
-> logout invalidates session and clears frontend state
```

### 7.2 Token/session transport

Confirm current backend behavior before implementation:

```text
Option A: secure httpOnly cookie for refresh/session, short-lived access token in memory
Option B: bearer access token returned by API plus backend refresh endpoint
Option C: existing backend session cookie contract
```

Guardrail:

```text
Do not store long-lived secrets in localStorage for production-like runtime.
Do not log tokens, passwords, refresh tokens, cookie values, or auth headers.
```

### 7.3 Vietnamese UX rules

Use short, operational Vietnamese copy:

```text
Đăng nhập
Email
Mật khẩu
Đăng xuất
Phiên đăng nhập đã hết hạn. Vui lòng đăng nhập lại.
Thông tin đăng nhập không đúng.
Tài khoản tạm khóa do đăng nhập sai nhiều lần.
Bạn không có quyền truy cập màn hình này.
```

Do not show raw backend codes to normal users.

---

## 8. Sprint 21 Backlog

| Task ID | Priority | Owner | Task | Output / Acceptance | Primary Ref |
|---|---:|---|---|---|---|
| S21-00-01 | P0 | PM/Tech Lead | Sprint 21 kickoff and scope lock | Branch created, current main clean, Sprint 20 status documented, production auth caveat confirmed as Sprint 21 target | `80_ERP_Master_Document_Index_Current_Status_MyPham_v2.md` |
| S21-00-02 | P0 | PM/Tech Lead | Update release/backlog note for Sprint 21 | README or sprint note states Sprint 21 target: wire web auth UI to backend auth/session API | `79_ERP_Sprint20_Changelog_Release_Hygiene_API_Modularization_Fallback_Cleanup_MyPham_v1.md` |
| S21-01-01 | P0 | BE/FE Lead | Auth API contract inventory | List existing auth endpoints, payloads, cookies/tokens, error codes, and session behavior; no code guesswork | `16_ERP_API_Contract_OpenAPI_Standards_Phase1_MyPham_v1.md` |
| S21-01-02 | P0 | BE | Auth OpenAPI alignment | OpenAPI documents login/me/refresh/logout/policy endpoints, request/response, error codes, auth headers/cookies | `16_ERP_API_Contract_OpenAPI_Standards_Phase1_MyPham_v1.md` |
| S21-01-03 | P0 | FE | Generated auth client refresh | Frontend generated client includes auth endpoints; no hand-written duplicate API shape when OpenAPI coverage exists | `15_ERP_Frontend_Architecture_React_NextJS_Phase1_MyPham_v1.md` |
| S21-02-01 | P0 | FE | Login UI wired to backend | Login form calls backend login endpoint, handles success/failure, redirects to intended route | `15_ERP_Frontend_Architecture_React_NextJS_Phase1_MyPham_v1.md` |
| S21-02-02 | P0 | FE | Vietnamese login copy | Login page, errors, loading, disabled states, and empty states use Vietnamese dictionary labels | `81_ERP_Vietnamese_UI_Glossary_Operational_Copy_MyPham_v1.md` |
| S21-02-03 | P0 | FE | Auth error mapping | Backend error codes map to Vietnamese UI messages; raw technical codes hidden outside debug/dev mode | `75_ERP_Coding_Task_Board_Sprint19_Vietnamese_UI_Localization_MyPham_v1.md` |
| S21-03-01 | P0 | FE | Session bootstrap | App bootstraps current user through `/auth/me` or equivalent backend session endpoint | `78_ERP_Production_Runtime_Mode_Checklist_Sprint20_MyPham_v1.md` |
| S21-03-02 | P0 | FE | Route guard integration | Protected routes require valid backend-backed session; unauthenticated users redirect to login | `19_ERP_Security_RBAC_Audit_Compliance_Standards_Phase1_MyPham_v1.md` |
| S21-03-03 | P0 | FE | RBAC menu from authenticated user | Sidebar/menu visibility uses backend-backed user roles/permissions, not hardcoded mock auth | `19_ERP_Security_RBAC_Audit_Compliance_Standards_Phase1_MyPham_v1.md` |
| S21-04-01 | P0 | FE/BE | Refresh rotation integration | Expired/near-expired session refreshes through backend refresh endpoint; failure redirects to login | `78_ERP_Production_Runtime_Mode_Checklist_Sprint20_MyPham_v1.md` |
| S21-04-02 | P0 | FE | Logout integration | Logout calls backend, clears frontend session state, and redirects to login | `78_ERP_Production_Runtime_Mode_Checklist_Sprint20_MyPham_v1.md` |
| S21-04-03 | P1 | FE | Multi-tab logout/session sync | Logout or session expiry in one tab is reflected in other tabs without stale privileged UI | `19_ERP_Security_RBAC_Audit_Compliance_Standards_Phase1_MyPham_v1.md` |
| S21-05-01 | P0 | FE/BE | Failed login and lockout policy UI | Failed login and lockout responses are surfaced safely in Vietnamese without leaking sensitive policy internals | `19_ERP_Security_RBAC_Audit_Compliance_Standards_Phase1_MyPham_v1.md` |
| S21-05-02 | P1 | FE | Unauthorized/forbidden screens | 401/403 screens are Vietnamese, minimal, and route-safe; include return-to-login or return-to-dashboard action | `39_ERP_UI_Template_Hetzner_Minimal_Style_Phase1_MyPham_v1.md` |
| S21-06-01 | P0 | FE | Production mock auth block | Production-like runtime cannot silently use mock auth or static signed-in state | `78_ERP_Production_Runtime_Mode_Checklist_Sprint20_MyPham_v1.md` |
| S21-06-02 | P0 | FE | Prototype fallback guard for auth surfaces | Auth-backed UI fails loudly when backend auth is unavailable in production-like runtime | `78_ERP_Production_Runtime_Mode_Checklist_Sprint20_MyPham_v1.md` |
| S21-06-03 | P1 | DevOps/FE | Env validation for auth runtime | Required auth env vars are validated for staging/prod; unsafe defaults fail before deploy/smoke | `18_ERP_DevOps_CICD_Environment_Standards_Phase1_MyPham_v1.md` |
| S21-07-01 | P0 | QA/FE | Auth UI smoke tests | Tests cover login success, invalid credentials, session bootstrap, logout, expired session redirect | `24_ERP_QA_Test_Strategy_Automation_Phase1_MyPham_v1.md` |
| S21-07-02 | P0 | QA/FE | RBAC route/menu regression | Tests verify role-based menu visibility and forbidden route handling | `19_ERP_Security_RBAC_Audit_Compliance_Standards_Phase1_MyPham_v1.md` |
| S21-07-03 | P0 | QA/FE | Production fallback regression | Production-mode test proves mock auth/fallback cannot mask backend auth failure | `78_ERP_Production_Runtime_Mode_Checklist_Sprint20_MyPham_v1.md` |
| S21-07-04 | P1 | QA | Vietnamese auth copy regression | Tests or snapshots verify key auth surfaces display Vietnamese copy | `81_ERP_Vietnamese_UI_Glossary_Operational_Copy_MyPham_v1.md` |
| S21-08-01 | P0 | DevOps/QA | Production-like smoke pack | Smoke script covers health/readiness, login, me, refresh, logout, RBAC route, warehouse page access | `78_ERP_Production_Runtime_Mode_Checklist_Sprint20_MyPham_v1.md` |
| S21-08-02 | P1 | DevOps | Dev/staging deploy smoke update | `make smoke-dev` / `make smoke-staging` include auth session checks | `18_ERP_DevOps_CICD_Environment_Standards_Phase1_MyPham_v1.md` |
| S21-09-01 | P0 | PM/Tech Lead | Remove production auth caveat from current docs only after proof | Update README/Index/Checklist to say web auth UI is backend-wired, with evidence; do not overclaim production readiness | `80_ERP_Master_Document_Index_Current_Status_MyPham_v2.md` |
| S21-09-02 | P0 | PM/Tech Lead | Sprint 21 changelog | Create Sprint 21 changelog with PR list, CI runs, auth smoke evidence, known limitations, next recommendation | `79_ERP_Sprint20_Changelog_Release_Hygiene_API_Modularization_Fallback_Cleanup_MyPham_v1.md` |
| S21-09-03 | P1 | PM/Tech Lead | Release tag decision | If required-ci and auth smoke pass, tag `v0.21.0-auth-ui-backend-integration-runtime-smoke` or explicitly hold tag with reason | `78_ERP_Production_Runtime_Mode_Checklist_Sprint20_MyPham_v1.md` |

---

## 9. Recommended Build Order

Do not start with visual polish. Build the auth contract and runtime safety first.

```text
1. Auth API contract inventory
2. OpenAPI/generated client alignment
3. Login UI -> backend login
4. /auth/me session bootstrap
5. Route/RBAC guard integration
6. Refresh/logout integration
7. Failed login/lockout UI
8. Production mock/fallback blocking
9. Auth smoke and regression tests
10. README/index/checklist update
11. Sprint 21 changelog and tag decision
```

---

## 10. Demo Script

End-of-sprint demo must show:

```text
1. Open ERP web without session.
2. User is redirected to login.
3. Login page displays Vietnamese copy.
4. Invalid credentials show safe Vietnamese error.
5. Valid login succeeds through backend API.
6. App loads current user/session from backend.
7. Sidebar/menu reflects backend-backed RBAC.
8. User without permission cannot access restricted route.
9. Refresh flow keeps session alive when appropriate.
10. Logout invalidates backend session and clears UI state.
11. Production-like mode cannot use mock auth silently.
12. Production runtime smoke pack passes.
```

---

## 11. Acceptance Criteria

Sprint 21 is done only when:

```text
- Login UI calls backend auth API.
- Frontend session state comes from backend session/me endpoint.
- Refresh rotation is integrated or explicitly handled according to backend contract.
- Logout invalidates backend session and clears frontend session state.
- RBAC menu/route behavior uses backend-backed user/permissions.
- Invalid login, expired session, locked account, and forbidden access show Vietnamese copy.
- Production-like runtime cannot silently use mock auth.
- Auth smoke tests pass in CI or documented production-like environment.
- required-ci is green.
- Production runtime checklist auth section is updated with evidence.
- README and Master Index no longer carry the auth mock caveat unless a known limitation remains.
- Sprint 21 changelog is created.
```

---

## 12. Guardrails

```text
1. Do not rename backend/API/DB auth codes to Vietnamese.
2. Do not change route paths to Vietnamese.
3. Do not log passwords, tokens, refresh tokens, cookies, or auth headers.
4. Do not store long-lived secrets in localStorage for production-like runtime.
5. Do not let production-like runtime silently sign in a mock user.
6. Do not let API failure render authenticated UI through prototype fallback.
7. Do not overclaim production readiness if SSO/MFA/password reset is not in scope.
8. Keep Vietnamese copy short and operational.
9. Keep all auth-sensitive actions auditable where backend supports it.
10. Manual PR review and merge only; no GitHub auto-review/auto-merge.
```

---

## 13. Known Non-Goals

These may be future work, not Sprint 21 blockers unless business requires them now:

```text
SSO/SAML/OIDC
MFA/2FA
Password reset email flow
User invitation flow
Role redesign
Advanced password policy admin UI
Device/session management dashboard
Audit export for auth events
Full English fallback completion for every deep label
```

---

## 14. Suggested Next Sprint After Sprint 21

After backend-backed auth UI is proven, the next sprint should return to operations or production hardening depending on remaining evidence:

```text
Option A: Sprint 22 - Attachment Storage + Evidence Lifecycle Hardening
Option B: Sprint 22 - English Fallback Cleanup + Deep UI Localization QA
Option C: Sprint 22 - UAT Pilot Pack for Warehouse + Sales + QC
```

Recommended default:

```text
Sprint 22 - UAT Pilot Pack for Warehouse + Sales + QC
```

Reason: after auth UI is backend-wired, the system becomes safer to put in front of real internal users for controlled UAT.

---

## 15. Definition of Done

A Sprint 21 PR is done only if:

```text
- Primary Ref and Task Ref are included in the PR.
- Auth/session behavior is tested or honestly documented.
- Production-like runtime does not use mock auth.
- Vietnamese UI copy follows file 81 glossary.
- Routes and technical codes remain English.
- No sensitive auth data is logged.
- required-ci passes.
- Manual review is completed.
```
