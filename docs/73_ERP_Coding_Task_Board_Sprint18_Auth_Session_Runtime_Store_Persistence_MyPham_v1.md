# 73_ERP_Coding_Task_Board_Sprint18_Auth_Session_Runtime_Store_Persistence_MyPham_v1

Project: Web ERP for cosmetics operations
Phase: Phase 1
Sprint: Sprint 18 - Auth/session runtime store persistence v1
Document role: Coding task board for Sprint 18 implementation
Version: v1.0
Primary stack: Go backend, PostgreSQL 16, OpenAPI, Docker dev deploy
Status: Planned; cloud CI and production tag remain on hold while GitHub Actions minutes are exhausted

---

## 1. Sprint 18 Context

Sprint 17 persisted master data runtime stores and updated the remaining prototype store ledger.

The highest remaining backend restart risk is now auth/session runtime state:

```text
Access token sessions      -> prototype memory
Refresh token sessions     -> prototype memory
Failed login lockout state -> prototype memory
```

Current mock/dev auth is acceptable for local development, but it is not production release evidence. In DB mode, authenticated runtime state must survive API restarts and must not store raw tokens.

---

## 2. Sprint 18 Theme

```text
Auth and Session Runtime Store Persistence
```

Business reason:

```text
The ERP cannot be production-traceable if active sessions, refresh rotation evidence, and lockout state disappear on deploy. Persisting auth/session state closes the last P1 backend runtime reset risk before frontend fallback cleanup and broader production hardening.
```

---

## 3. Sprint 18 Goals

By the end of Sprint 18, DB-mode runtime must support:

```text
1. Access session state persisted with token hash, principal snapshot, expiry, revocation, and audit-ready timestamps.
2. Refresh session state persisted with token hash, rotation/revocation state, expiry, and owner relation.
3. Failed login attempt and lockout state persisted by normalized email and organization.
4. Raw access or refresh tokens are never stored in PostgreSQL.
5. Runtime selectors use PostgreSQL-backed auth/session stores when DATABASE_URL exists and existing in-memory behavior only for no-DB/local mode.
6. Existing /login, /refresh, /logout, and authenticated API behavior remains compatible.
7. Static local dev access token remains local/dev/test only and is not treated as production persistence evidence.
8. Dev smoke proves login session and lockout state survive API restart.
9. Remaining prototype ledger and Sprint 18 release evidence are updated after auth/session persistence.
```

---

## 4. Sprint 18 Non-Goals

Sprint 18 does not include:

```text
- Full user administration UI.
- OIDC/SAML/SSO integration.
- Multi-factor authentication.
- Password reset email flows.
- Production secret rotation automation.
- RBAC policy redesign.
- Public API response shape changes unless required by the auth persistence work.
- Frontend fallback cleanup; that remains the next sprint candidate after auth/session persistence.
```

---

## 5. Branch / PR / Release Rules

Current repo workflow remains:

```text
task branch
-> build/test on dev server when runtime changes require it
-> PR
-> manual self-review comment
-> manual merge into main
-> sync/deploy dev server when runtime changes require it
```

Do not use GitHub auto review or auto merge.
Do not create a long-lived sprint branch unless the team explicitly changes that policy.

Default task branch pattern:

```text
codex/feature-S18-xx-yy-short-task-name
```

Recommended Sprint 18 release tag after completion:

```text
v0.18.0-auth-session-runtime-store-persistence
```

Create the production tag only after:

```text
1. Main required-ci is green, or the team explicitly accepts a manual-only release gate while GitHub Actions quota is exhausted.
2. Dev release gate is green.
3. PostgreSQL migration apply/rollback evidence is green for new migrations.
4. Sprint 18 changelog records persisted scope, remaining prototype stores, dev deploy, CI, and tag status.
```

Current CI note:

```text
GitHub Actions cloud CI is blocked because the monthly included minutes are exhausted: 2,000 min used / 2,000 min included.
Until that changes, do not claim cloud CI verification and do not tag production without explicit manual-only release approval.
```

---

## 6. Sprint 18 Demo Script

### Case 1: Access session survives restart

```text
1. Login with the dev admin credential.
2. Call an authenticated endpoint with the returned access token.
3. Restart/redeploy API.
4. Call the same authenticated endpoint with the same unexpired access token.
5. Confirm the request still succeeds and audit evidence remains available.
```

### Case 2: Refresh rotation survives restart

```text
1. Login and capture access and refresh tokens.
2. Restart/redeploy API.
3. Refresh the session with the original refresh token.
4. Confirm a new access/refresh pair is issued.
5. Confirm the old refresh token cannot be reused.
```

### Case 3: Lockout survives restart

```text
1. Submit invalid login attempts until the configured lockout threshold is reached.
2. Restart/redeploy API.
3. Submit another login attempt during the lockout window.
4. Confirm the lockout is still enforced.
```

---

## 7. Sprint 18 Guardrails

These rules are non-negotiable:

```text
1. Do not store raw access tokens or raw refresh tokens.
2. Token lookup must use deterministic hashes.
3. Refresh token rotation must invalidate the previous refresh token.
4. Expired, revoked, or rotated tokens must not authenticate.
5. Lockout state must use normalized email and organization scope.
6. Static local dev tokens remain disabled for prod-like environments.
7. Do not loosen password policy, lockout thresholds, RBAC, or audit behavior to pass tests.
8. Do not change public API response shapes unless OpenAPI and web clients are updated in the same task.
9. Prototype fallback remains no-DB/local only.
10. Verification must distinguish local/dev checks from cloud CI, which is currently quota-blocked.
```

---

## 8. Dependency Map

```text
S18-00-00 Sprint 18 task board
  -> S18-01-01 auth/session persistence design

S18-01-01 auth/session persistence design
  -> S18-01-02 auth/session migration foundation
  -> S18-02-01 auth session store interface
  -> S18-03-01 failed login lockout store interface

S18-01-02 auth/session migration foundation
  -> S18-02-02 PostgreSQL access/refresh session store
  -> S18-03-02 PostgreSQL failed login lockout store

S18-02-01 auth session store interface
  -> S18-02-02 PostgreSQL access/refresh session store
  -> S18-02-03 auth session persistence tests

S18-03-01 failed login lockout store interface
  -> S18-03-02 PostgreSQL failed login lockout store
  -> S18-03-03 lockout persistence tests

S18-02-03 auth session persistence tests
  -> S18-04-01 runtime selector wiring

S18-03-03 lockout persistence tests
  -> S18-04-01 runtime selector wiring

S18-04-01 runtime selector wiring
  -> S18-04-02 auth restart smoke
  -> S18-05-01 auth audit and permission regression

S18-04-02 auth restart smoke
  -> S18-06-01 remaining prototype ledger update
  -> S18-07-01 Sprint 18 release evidence
```

---

## 9. Task Board

| Task ID | Task | Output / Acceptance | Primary Ref |
| --- | --- | --- | --- |
| S18-00-00 | Sprint 18 task board | File 73 created with scope, guardrails, sequencing, verification gates, and task list | `docs/72_ERP_Sprint17_Changelog_Master_Data_Runtime_Store_Persistence_MyPham_v1.md` |
| S18-01-01 | Auth/session persistence design | Map current session manager, token lifecycle, lockout state, PostgreSQL tables, fallback behavior, tests, smoke, and rollback | `docs/qa/S18-01-01_auth_session_persistence_design.md` |
| S18-01-02 | Auth/session migration foundation | Migration creates auth session and failed-login lockout tables, indexes, constraints, and rollback | `apps/api/migrations/000035_persist_auth_session_runtime_foundation.up.sql` |
| S18-02-01 | Auth session store interface | Existing in-memory session behavior is preserved behind an interface without changing API handlers | `apps/api/internal/shared/auth/session.go` |
| S18-02-02 | PostgreSQL access/refresh session store | Access lookup, refresh lookup, issue, revoke, expiry, and rotation persist by token hash | `apps/api/internal/shared/auth/postgres_session_store.go` |
| S18-02-03 | Auth session persistence tests | Fresh store reload proves access auth survives restart and refresh rotation invalidates the previous refresh token | `apps/api/internal/shared/auth/postgres_session_store_test.go` |
| S18-03-01 | Failed login lockout store interface | Failed-attempt and lockout state are isolated behind a store interface while preserving policy behavior | `apps/api/internal/shared/auth/session.go` |
| S18-03-02 | PostgreSQL failed login lockout store | Failed attempts, first-failed timestamp, and lockout-until persist by org and normalized email | `apps/api/internal/shared/auth/postgres_login_failure_store.go` |
| S18-03-03 | Lockout persistence tests | Fresh store reload proves lockout survives API/session manager restart | `apps/api/internal/shared/auth/postgres_login_failure_store_test.go` |
| S18-04-01 | Runtime selector wiring | DB mode wires PostgreSQL auth/session stores; no-DB/local mode keeps existing in-memory behavior | `apps/api/cmd/api/auth_session_store_selection.go` |
| S18-04-02 | Auth restart smoke | Full dev smoke proves login access, refresh rotation, and lockout state survive API restart | `infra/scripts/smoke-dev-full.sh` |
| S18-05-01 | Auth audit and permission regression | Login, refresh, logout, RBAC-protected routes, and audit behavior remain compatible after persistence wiring | `apps/api/cmd/api/main_test.go` |
| S18-06-01 | Remaining prototype ledger update | Remaining prototype ledger supersedes Sprint 17 and removes auth/session runtime state from P1 backend persistence gaps | `docs/qa/S14-04-01_remaining_prototype_store_ledger.md` |
| S18-07-01 | Sprint 18 release evidence | Changelog records PRs, migrations, verification, persisted stores, remaining prototype stores, dev deploy, CI, and tag status | `docs/74_ERP_Sprint18_Changelog_Auth_Session_Runtime_Store_Persistence_MyPham_v1.md` |

---

## 10. Verification Gates

Backend checks:

```text
go test ./...
go vet ./...
```

Focused auth checks:

```text
go test ./internal/shared/auth ./cmd/api -run "Test(Login|Refresh|Logout|Session|Auth|RBAC|Permission|Postgres|Lockout)" -count=1
```

Migration checks when migrations change:

```text
Apply all migrations on PostgreSQL 16.
Roll back all migrations on PostgreSQL 16.
Apply again after rollback when practical.
```

OpenAPI checks when API contracts change:

```text
pnpm openapi:validate
pnpm openapi:contract
pnpm openapi:generate
git diff --exit-code apps/web/src/shared/api/generated/schema.ts
```

Dev release gate:

```text
Dev deploy or repo sync evidence for merged main.
Dev release gate smoke.
GitHub required checks green only when Actions minutes are available.
```

---

## 11. Definition Of Done

For each code task:

```text
1. Code is scoped to the task.
2. Raw tokens are never persisted or logged.
3. Backend tests pass for touched services/stores/handlers.
4. OpenAPI validate/contract/generate pass when API contracts change.
5. Runtime changes include dev server deploy/smoke evidence after merge.
6. PR includes manual self-review comment.
7. Any remaining unverified release gate is called out explicitly.
```

Sprint 18 completion requires:

```text
1. All S18 tasks merged to main.
2. DB-mode runtime selection wires auth/session stores to PostgreSQL.
3. Dev server smoke proves access session, refresh rotation, and lockout state survive API restart.
4. PostgreSQL migration apply/rollback is verified on PostgreSQL 16.
5. Remaining prototype ledger is updated.
6. Sprint 18 changelog records that GitHub Actions cloud CI is blocked if quota is still exhausted.
7. Tag v0.18.0-auth-session-runtime-store-persistence is created only after release gates are green or explicit manual-only release approval is given.
```
