# 74_ERP_Sprint18_Changelog_Auth_Session_Runtime_Store_Persistence_MyPham_v1

Project: Web ERP for cosmetics operations
Phase: Phase 1
Sprint: Sprint 18 - Auth/session runtime store persistence v1
Document role: Sprint changelog and release evidence
Version: v1.0
Date: 2026-05-02
Status: Release evidence complete except cloud CI; production tag is on hold while GitHub Actions minutes are exhausted

---

## 1. Sprint 18 Scope

Sprint 18 closed the highest remaining P1 backend runtime reset risk after Sprint 17:

```text
access token sessions were in-memory only
refresh token sessions and rotation evidence were in-memory only
failed login attempt and lockout state were in-memory only
runtime API selected in-memory auth/session state even when DB-mode stores were used elsewhere
full dev smoke did not prove auth/session state survived API restart
```

Promoted scope:

```text
Sprint 18 task board
auth/session persistence design
auth/session migration foundation
auth SessionStore interface and in-memory implementation
PostgreSQL-backed access/refresh session store
LoginFailureStore interface and in-memory implementation
PostgreSQL-backed login failure lockout store
DB-mode runtime auth/session selector
full dev auth restart smoke
auth audit and permission regression verification
remaining prototype store ledger update
Sprint 18 release evidence
```

No full user administration UI, OIDC/SAML/SSO integration, MFA, password reset email flow, production identity-provider integration, RBAC policy redesign, or frontend fallback cleanup was introduced. Sprint 18 persisted runtime auth/session state behind the existing mock/dev auth surface.

---

## 2. Merged PRs

| Task | PR | Result |
| --- | --- | --- |
| S18-00-00 Sprint 18 task board | #495 | Created Sprint 18 task board |
| S18-01-01 Auth/session persistence design | #496 | Documented token hashing, refresh rotation, lockout persistence, selector, tests, smoke, and rollback |
| S18-01-02 Auth/session migration foundation | #497 | Added migration `000035_persist_auth_session_runtime_foundation` |
| S18-02-01 Auth session store interface | #498 | Moved access/refresh session state behind `SessionStore` with in-memory fallback |
| S18-02-02 / S18-02-03 PostgreSQL access/refresh session store and tests | #499 | Added token-hash-only PostgreSQL session store and persistence tests |
| S18-03-01 Failed login lockout store interface | #500 | Moved failed login and lockout state behind `LoginFailureStore` with in-memory fallback |
| S18-03-02 / S18-03-03 PostgreSQL failed login lockout store and tests | #501 | Added PostgreSQL lockout store and persistence tests |
| S18-04-01 Runtime selector wiring | #502 | Wired DB-mode auth/session runtime to PostgreSQL stores |
| S18-04-02 Auth restart smoke | #503 | Added auth session restart/refresh/lockout checks to full dev smoke |
| S18-05-01 Auth audit and permission regression | verification task | Backend auth/audit and web permission/audit regression tests passed without code changes |
| S18-06-01 Remaining prototype ledger update | #504 | Removed auth/session runtime state from remaining P1 backend persistence gaps |
| S18-07-01 Sprint 18 release evidence | #505 | Records release evidence, CI quota blocker, remaining gaps, and production tag hold |

All PRs used the manual review and merge flow. GitHub auto review and auto merge were not used.

---

## 3. Persistence Changes

### Runtime Selector

| Runtime path | DB mode | No-DB/local fallback |
| --- | --- | --- |
| `newRuntimeSessionManager` session store | `auth.PostgresSessionStore` | `auth.InMemorySessionStore` |
| `newRuntimeSessionManager` login failure store | `auth.PostgresLoginFailureStore` | `auth.InMemoryLoginFailureStore` |

DB mode selects auth session and login failure stores as one package. In-memory fallback remains intentional for no-DB/local mode and is not production persistence evidence.

### PostgreSQL Persistence

| Migration | Purpose |
| --- | --- |
| `000035_persist_auth_session_runtime_foundation` | Adds `core.auth_sessions` and `core.auth_login_failures` with indexes, constraints, and rollback |

Persisted behavior:

```text
POST /api/v1/auth/login
POST /api/v1/auth/mock-login
POST /api/v1/auth/refresh
GET  /api/v1/me
RBAC-protected routes using RequireSessionToken / RequireSessionPermission
failed login lockout checks before credential validation
```

Persisted evidence:

```text
core.auth_sessions
core.auth_login_failures
audit.audit_logs auth.login_succeeded/auth.login_failed/auth.refresh_succeeded/auth.refresh_failed
```

Security behavior:

```text
Raw access tokens are not stored in PostgreSQL.
Raw refresh tokens are not stored in PostgreSQL.
Token lookup uses deterministic SHA-256 hex hashes.
Refresh rotation revokes/rotates the previous refresh row.
Expired, revoked, or rotated sessions do not authenticate.
```

---

## 4. Dev Release Evidence

Dev server:

```text
Host: 10.1.1.120
Repo: /opt/ERP-v2
Runtime dev URL: http://10.1.1.120:8088
```

Runtime deploy evidence:

```text
After PR #497, main was synced to dev and migration `000035` was applied to the dev PostgreSQL database.
After PR #498, main was synced/deployed to dev and API was rebuilt/restarted.
After PR #500, main was synced/deployed to dev and API was rebuilt/restarted.
After PR #502, main was synced/deployed to dev, migration up was confirmed no-change, API was rebuilt/restarted, and manual auth restart smoke passed.
After PR #503, main was synced to dev and full dev smoke passed using the updated smoke script.
PR #504 was docs-only and does not require a runtime rebuild.
This S18-07 task is documentation-only and does not require a runtime rebuild.
```

Latest Sprint 18 smoke evidence on dev:

```text
auth_session_login           200
auth_me_before_restart       200
api_restart                  ok
auth_me_after_restart        200
auth_refresh_rotate          200
auth_old_refresh_reject      401
auth_lockout_attempt_1       401
auth_lockout_attempt_2       401
auth_lockout_attempt_3       401
auth_lockout_attempt_4       401
auth_lockout_attempt_5       401
api_restart                  ok
auth_lockout_after_restart   401
persisted_auth_session       ok access/refresh/lockout
Full ERP dev smoke passed
```

The same full smoke also passed the previously persisted master data, finance, sales reservation/order, stock adjustment/movement/count, purchase order, inbound QC, carrier manifest, pick task, pack task, return receipt, supplier rejection, and subcontract checks.

---

## 5. CI And Migration Evidence

GitHub Actions status:

```text
Cloud CI is blocked for Sprint 18 PRs because the GitHub Actions plan has used 100% of the included monthly minutes.
Quota message: 2,000 min used / 2,000 min included.
Do not treat Sprint 18 as production-tagged while this CI gate is blocked.
```

Local/dev verification highlights:

```text
S18-01-02: PostgreSQL 16 isolated migration apply/down/reapply passed through migration 000035.
S18-02-01: focused auth and auth-handler tests passed on dev server Docker.
S18-02-02: PostgreSQL 16 isolated auth session store tests passed.
S18-03-01: focused auth and auth-handler tests passed on dev server Docker.
S18-03-02: PostgreSQL 16 isolated login failure store tests passed.
S18-04-01: PostgreSQL 16 isolated runtime selector/auth tests passed.
S18-04-01: dev deploy passed; access token survived API restart; refresh rotation and old refresh rejection passed; lockout survived API restart.
S18-04-02: smoke script syntax check passed; merged full dev smoke passed.
S18-05-01: backend auth/audit regression tests passed.
S18-05-01: web permission/audit regression tests passed, 19 tests.
S18-06-01: remaining prototype ledger updated; documentation-only change.
S18-07-01: changelog created with CI blocked and tag hold recorded.
```

Migration runtime gate:

```text
PostgreSQL 16 isolated container
Action: apply every *.up.sql in order through 000035, apply 35 down migrations, then apply every *.up.sql again
Result: passed
Applied migrations: 35
Rolled back migrations: 35
Reapplied migrations: 35
```

---

## 6. Remaining Prototype / Fallback Areas

Current remaining-store ledger:

```text
docs/qa/S14-04-01_remaining_prototype_store_ledger.md
```

Highest remaining persistence-adjacent candidate after Sprint 18:

```text
1. Remove or gate frontend fallback services where backend coverage is now available.
```

Auth/session runtime state is no longer listed as a P1 backend persistence gap when DB config exists.

Future auth/security product work remains separate:

```text
full user administration
password reset email flow
OIDC/SAML/SSO
MFA
production identity-provider integration
```

---

## 7. Release Status

Sprint 18 release gate status:

```text
Task PRs: merged through S18-06-01 PR #504; this changelog is S18-07-01 PR #505
Main cloud CI: blocked by exhausted GitHub Actions minutes
Dev runtime smoke: green at runtime commit be417d23
Docs-only main sync: pending after this changelog merge
Migration apply/down/reapply: green on PostgreSQL 16 isolated instance
Production tag: HOLD
```

Recommended production tag after CI is available and green:

```text
v0.18.0-auth-session-runtime-store-persistence
```

Do not create the tag until either:

```text
1. GitHub Actions minutes reset or billing is fixed, required checks run, and the checks are green; or
2. the team explicitly accepts a manual-only release gate for this sprint.
```
