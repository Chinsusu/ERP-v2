# S18-01-01 Auth Session Persistence Design

Project: Web ERP for cosmetics operations
Sprint: Sprint 18 - Auth/session runtime store persistence
Task: S18-01-01 Auth/session persistence design
Date: 2026-05-02
Status: Design ready for migration and implementation tasks

---

## 1. Purpose

Sprint 18 closes the remaining P1 backend runtime reset risk from the Sprint 17 prototype ledger:

```text
Access/refresh sessions reset after API restart.
Failed login attempts and lockout state reset after API restart.
```

The goal is to persist runtime auth/session state in PostgreSQL when `DATABASE_URL` is configured, while preserving the existing mock/dev login surface and in-memory fallback for no-DB/local development.

---

## 2. Current Runtime Behavior

Current constructor:

```text
auth.NewSessionManager(authConfig, time.Now)
```

Current storage:

```text
accessTokens  map[string]Session
refreshTokens map[string]Session
failedLogins  map[string]failedLoginState
```

Current public methods:

```text
Login(email, password) -> Session, LoginFailure, bool
Refresh(refreshToken) -> Session, bool
AuthenticateAccessToken(accessToken) -> Principal, bool
PasswordPolicy()
LockoutPolicy()
```

Current handlers using the session manager:

```text
POST /api/v1/auth/login
POST /api/v1/auth/mock-login
POST /api/v1/auth/refresh
GET  /api/v1/auth/policy
GET  /api/v1/me
RBAC-protected API routes through RequireSessionToken / RequireSessionPermission
```

Current policies:

```text
Access token TTL: 8 hours
Refresh token TTL: 7 days
Lockout window: 15 minutes
Lockout duration: 15 minutes
Max failed attempts: 5
Minimum password length: 10
```

Current static token rule:

```text
AUTH_MOCK_ACCESS_TOKEN is only exposed for local/dev/development/test APP_ENV values.
Production-like APP_ENV values return an empty static token.
```

---

## 3. Design Decisions

### D1 - Keep Handler Contracts Stable

Do not change route paths, request payloads, response payloads, or existing handler call sites in the first persistence tasks.

Rationale:

```text
Auth persistence can be implemented behind SessionManager without forcing a frontend or OpenAPI churn task.
```

Implementation direction:

```text
Add store abstractions behind SessionManager.
Keep Login, Refresh, AuthenticateAccessToken, PasswordPolicy, and LockoutPolicy stable.
```

### D2 - Never Persist Raw Tokens

Persist deterministic token hashes only.

Recommended hash:

```text
sha256(token) as lowercase hex
```

Rationale:

```text
Access and refresh tokens are bearer secrets. Database reads, dumps, logs, or admin queries must not reveal reusable credentials.
```

### D3 - Store Principal Snapshot With Session

Persist the principal snapshot issued at login time:

```text
user_ref
email
display_name
role_code
permissions jsonb
```

Rationale:

```text
The current auth model is mock/dev principal based and does not yet read real users/roles from core.users. Persisting the issued principal preserves current behavior without inventing full user administration in Sprint 18.
```

Future upgrade:

```text
When real user/password auth is introduced, sessions can additionally link to core.users.id and refresh principal claims from RBAC tables.
```

### D4 - Rotate Refresh Tokens Atomically

Refresh must:

```text
1. Find an active, unexpired refresh token by hash.
2. Mark the old session rotated/revoked.
3. Insert a new access/refresh session.
4. Ensure the old access token no longer authenticates.
5. Ensure the old refresh token cannot be reused.
```

Rationale:

```text
This preserves current in-memory behavior and prevents replay of refresh tokens after rotation.
```

### D5 - Persist Lockout By Org And Normalized Email

Failed login state uses:

```text
org_id
email_normalized
attempts
first_failed_at
locked_until
updated_at
```

Rationale:

```text
The current lockout key is normalized email. Adding org_id matches the rest of Phase 1 tables and avoids cross-tenant lockout collisions.
```

Local/dev org:

```text
Use localAuditOrgID (00000000-0000-4000-8000-000000000001) when static dev auth is allowed.
```

### D6 - Preserve No-DB Fallback

Runtime selector:

```text
if DATABASE_URL is empty:
  SessionManager uses current in-memory session and lockout stores
else:
  SessionManager uses PostgreSQL-backed session and lockout stores
```

Rationale:

```text
No-DB/local mode is still useful for fast local tests and does not count as production persistence evidence.
```

---

## 4. Proposed Database Shape

Migration:

```text
000035_persist_auth_session_runtime_foundation
```

### core.auth_sessions

Purpose:

```text
Persist issued access/refresh session records and rotation/revocation state.
```

Columns:

```text
id uuid primary key default gen_random_uuid()
org_id uuid not null references core.organizations(id)
session_ref text not null unique
user_ref text not null
email text not null
display_name text not null
role_code text not null
permissions jsonb not null default '[]'::jsonb
access_token_hash text not null unique
refresh_token_hash text not null unique
access_expires_at timestamptz not null
refresh_expires_at timestamptz not null
revoked_at timestamptz
rotated_at timestamptz
last_seen_at timestamptz
created_at timestamptz not null default now()
updated_at timestamptz not null default now()
version integer not null default 1
```

Indexes:

```text
unique access_token_hash
unique refresh_token_hash
org_id, email
access_expires_at
refresh_expires_at
revoked_at where revoked_at is null
```

Constraints:

```text
jsonb_typeof(permissions) = 'array'
access_expires_at < refresh_expires_at
```

Security note:

```text
No column may store raw access token or raw refresh token.
```

### core.auth_login_failures

Purpose:

```text
Persist failed login attempts and lockout window state.
```

Columns:

```text
id uuid primary key default gen_random_uuid()
org_id uuid not null references core.organizations(id)
email_normalized text not null
attempts integer not null default 0
first_failed_at timestamptz
locked_until timestamptz
created_at timestamptz not null default now()
updated_at timestamptz not null default now()
version integer not null default 1
```

Indexes and constraints:

```text
unique(org_id, email_normalized)
attempts >= 0
index locked_until
```

Cleanup behavior:

```text
Successful login clears the row for the normalized email.
Expired lockout is cleared lazily on the next lockout check or failed-login write.
```

---

## 5. Store Abstractions

Keep this narrow and close to current behavior.

Proposed interfaces:

```text
type SessionStore interface {
  StoreSession(Session, time.Time) error
  FindByAccessToken(accessToken string, now time.Time) (Session, bool, error)
  RotateRefreshToken(refreshToken string, next Session, now time.Time) (bool, error)
}

type LoginFailureStore interface {
  LockedUntil(email string, now time.Time) (time.Time, bool, error)
  RecordFailure(email string, now time.Time, policy LockoutPolicy) (time.Time, error)
  Clear(email string) error
}
```

Current in-memory maps become the no-DB implementation of these interfaces.

PostgreSQL implementations:

```text
PostgresSessionStore
PostgresLoginFailureStore
```

SessionManager remains the auth policy owner:

```text
password validation
mock credential validation
token generation
TTL calculation
refresh rotation rule
principal creation
```

Stores own persistence only:

```text
hash token before write/read
enforce active/unexpired query filters
persist rotation/revocation fields
serialize/deserialize principal permissions
persist lockout counters
```

---

## 6. Runtime Selector

New selector:

```text
newRuntimeSessionManager(cfg config.Config, now func() time.Time) (*auth.SessionManager, func() error, error)
```

Expected behavior:

```text
Without DATABASE_URL:
  return auth.NewSessionManager(authConfig, now), nil, nil

With DATABASE_URL:
  open pgx DB
  configure default org for local/dev/test
  create PostgresSessionStore and PostgresLoginFailureStore
  return auth.NewSessionManagerWithStores(authConfig, now, sessionStore, failureStore), db.Close, nil
```

Main wiring:

```text
replace direct auth.NewSessionManager(authConfig, time.Now)
with newRuntimeSessionManager(cfg, time.Now)
```

Close ordering:

```text
Close the auth DB connection with the other runtime store close functions during shutdown/error cleanup.
```

Static token seeding:

```text
No-DB/local: keep current static token seed.
DB-mode local/dev/test: static token may be seeded into the selected runtime manager for dev compatibility, but it remains explicitly local/dev/test only and is not production auth evidence.
Production-like APP_ENV: static token remains disabled because config.StaticAuthAccessToken() returns empty.
```

---

## 7. Test Plan

### Unit Tests

```text
Existing session tests continue to pass.
New in-memory store tests cover the same behavior through the store abstraction.
```

### PostgreSQL Store Tests

Session persistence:

```text
1. Login through manager backed by PostgreSQL store.
2. Create a second manager with the same database.
3. Authenticate the original access token through the second manager.
4. Refresh the original refresh token through the second manager.
5. Confirm old access token no longer authenticates.
6. Confirm old refresh token cannot be reused.
```

Lockout persistence:

```text
1. Submit invalid logins until locked using manager A.
2. Create manager B with the same PostgreSQL store.
3. Submit valid login during lockout.
4. Confirm login is rejected with LoginFailureLocked.
```

Selector tests:

```text
DATABASE_URL empty -> in-memory stores and nil close function.
DATABASE_URL configured -> PostgreSQL stores and close function.
```

### Migration Gate

```text
Apply all migrations through 000035 on PostgreSQL 16.
Roll back all migrations from 000035 down to baseline in reverse order.
Apply all migrations again when practical.
```

### Dev Smoke

```text
1. Login with admin@example.local / local-only-mock-password.
2. Call /api/v1/me with returned access token.
3. Restart API.
4. Call /api/v1/me again with the same access token.
5. Refresh with the original refresh token.
6. Confirm old refresh token cannot be reused.
7. Trigger lockout and confirm it survives API restart.
```

Cloud CI:

```text
GitHub Actions checks remain quota-blocked until Actions minutes are available. Do not report cloud CI as verified while checks remain queued/skipped/blocked.
```

---

## 8. Rollback Plan

Code rollback:

```text
Revert runtime selector to auth.NewSessionManager(authConfig, time.Now).
PostgreSQL tables can remain unused until the migration rollback is applied.
```

Migration rollback:

```text
DROP TABLE IF EXISTS core.auth_login_failures;
DROP TABLE IF EXISTS core.auth_sessions;
```

Operational rollback:

```text
Restart API after rollback.
Existing persisted auth sessions are invalidated if the tables are dropped.
Users can login again through the existing mock/dev login path.
```

---

## 9. Risks And Mitigations

| Risk | Mitigation |
| --- | --- |
| Raw token accidentally persisted | Use hash helper inside the PostgreSQL store; schema contains only hash columns |
| Refresh token replay | Rotate in a transaction and require `revoked_at IS NULL AND rotated_at IS NULL` on refresh lookup |
| Old access token remains valid after refresh | Rotation updates old session revoked/rotated fields and access lookup filters them out |
| Lockout row races under repeated login attempts | Use PostgreSQL upsert with row-level update semantics |
| Mock/dev auth mistaken for production auth | Keep docs and changelog explicit that S18 persists runtime sessions but does not implement full user admin/OIDC/MFA |
| CI status misreported | Keep release evidence explicit that GitHub Actions is quota-blocked |

---

## 10. Acceptance Criteria

S18-01-01 is complete when:

```text
1. Current auth/session behavior and code paths are documented.
2. PostgreSQL schema shape is specified before migration work.
3. Store boundaries preserve SessionManager public methods.
4. Token hashing, refresh rotation, lockout persistence, and static-token constraints are documented.
5. Test, smoke, migration, and rollback plans are documented.
```
