# 78_ERP_Production_Runtime_Mode_Checklist_Sprint20_MyPham_v1

Project: Web ERP for cosmetics operations
Phase: Phase 1
Sprint: Sprint 20 - Release hygiene, API modularization, and fallback cleanup
Document role: Production runtime mode checklist
Version: v1.0
Date: 2026-05-02
Status: Active checklist for production-like deployments

---

## 1. Goal

This checklist defines the minimum runtime checks before treating a deployment as production-like.

Production-like means staging, pilot, production rehearsal, or production. Local/no-DB development remains allowed to use mock data and prototype fallbacks, but production-like runtime must fail loudly when required backend, persistence, auth, attachment, or reporting coverage is unavailable.

---

## 2. Runtime Modes

| Mode | Intended use | Required behavior |
| --- | --- | --- |
| Local/no-DB | Developer workflow, UI smoke, disconnected demo | Prototype fallbacks and mock auth may be enabled |
| Shared dev | Team integration environment | Backend services should be used; fallback is allowed only for known prototype gaps |
| Staging | Production rehearsal | Backend services, persistence, attachments, and auth/session stores must be live |
| Production | Real operations | No silent prototype fallback, no default secrets, no static auth token, all smoke checks green |

Runtime mode must be explicit. Do not infer production readiness from a successful web build alone.

---

## 3. Environment Checklist

API/worker:

```text
APP_ENV is staging, prod, or production for production-like runtime.
APP_PORT is set and exposed only through the intended proxy.
DATABASE_URL points to the production-like PostgreSQL database, not localhost defaults.
REDIS_URL and QUEUE_URL point to the production-like Redis instance.
JWT_SECRET is rotated away from every example/default value.
S3_ENDPOINT, S3_BUCKET, S3_ACCESS_KEY, S3_SECRET_KEY, S3_USE_SSL, and S3_USE_PATH_STYLE match the attachment store.
SMTP_HOST, SMTP_PORT, and SMTP_FROM match the notification environment when email is enabled.
LOG_LEVEL is appropriate for operations and does not log credentials or private payloads.
```

Web:

```text
NODE_ENV=production for production-like builds/runtime.
NEXT_PUBLIC_API_BASE_URL points to the deployed API root ending in /api/v1.
NEXT_PUBLIC_MOCK_AUTH_STATE is unset or signed-out unless the environment is explicitly a local/dev mock.
AUTH_COOKIE_SECURE=true for HTTPS deployments.
```

Auth:

```text
APP_ENV must not allow static auth tokens outside local/dev/test.
Auth/session persistence must use PostgreSQL-backed stores from Sprint 18.
Login, refresh rotation, failed login attempts, lockout, and logout must be smoke-checked.
```

---

## 4. Persistence Checklist

```text
PostgreSQL 16 migration gate passes apply -> rollback -> reapply.
Inventory, stock ledger, allocation, shipping, returns, purchase, receiving, QC, subcontract, finance, master data, auth, and session runtime stores are PostgreSQL-backed where backend coverage exists.
Redis is reachable for worker and queue paths.
S3/MinIO-compatible storage is reachable for attachments, delivery notes, COA/MSDS, QC evidence, return evidence, and finance evidence.
Runtime services must not write operational data only to process memory in production-like mode.
```

Any remaining prototype or in-memory store must be listed as a known non-production gap before release.

---

## 5. Fallback Checklist

Sprint 20 introduced a shared web fallback guard:

```text
apps/web/src/shared/api/prototypeFallback.ts
```

Required checks:

```text
NODE_ENV=production disables shouldUsePrototypeFallback().
Backend API errors are not hidden by prototype fallback data.
Generic API request failures are not hidden by prototype fallback data.
Local/no-DB development still keeps fallback support for developer velocity.
Production-like release notes call out any frontend page still backed only by prototype data.
```

Hard failure:

```text
If a production-like page claims backend coverage but silently renders prototype data after an API failure, the release is blocked.
```

---

## 6. Smoke Checklist

Run these after deploy and before tagging a production-like release:

```text
Health/readiness:
- API health and readiness endpoints return success.
- Web root renders and can reach NEXT_PUBLIC_API_BASE_URL.

Auth/session:
- Login succeeds with intended credentials.
- /auth/me or equivalent session check returns the expected user.
- Refresh rotation works.
- Logout clears the session.
- Invalid login increments policy state without exposing sensitive details.

Inbound:
- Purchase order -> receiving -> inbound QC PASS -> stock available.
- Purchase order -> receiving -> inbound QC FAIL -> no available stock and supplier rejection path recorded.

Outbound:
- Sales order -> reserve -> pick -> pack -> manifest -> handover scan.
- Unavailable stock cannot be reserved or picked.

Returns/reconciliation:
- Return scan -> inspect -> disposition.
- End-of-day reconciliation records mismatches and closing status.

Subcontract:
- Material issue -> factory receiving -> final receiving.
- PASS/FAIL/PARTIAL receiving updates stock and audit records correctly.

Finance:
- COD collection/reconciliation, AR receipt, AP payment, and cash dashboard paths render from backend data.

Reporting/audit/attachments:
- Inventory and operations reports load.
- CSV/export paths return files.
- Audit log records privileged actions.
- Attachment upload/download works through the configured object store.
```

---

## 7. CI And Release Checklist

Before merge:

```text
Task branch uses the codex/ prefix and maps to the Sprint task ID.
Pull request includes Primary Ref, Task Ref, verification, and rollback notes.
Relevant local checks are run and reported honestly.
Cloud required-ci is green.
Manual self-review is complete.
GitHub auto-review and auto-merge are not used.
```

Before tag:

```text
main is up to date.
required-api is green.
required-web is green.
required-openapi is green.
required-migration is green with PostgreSQL 16 apply -> rollback -> reapply.
E2E PR workflow is green for user-facing workflow changes.
Release docs and README match the actual commit/tag/CI state.
Production-like smoke checklist is green, or the release is explicitly marked not production-ready.
```

Do not tag if any required CI gate is red or skipped.

---

## 8. Sprint 20 Evidence Snapshot

Sprint 20 closed the release hygiene and hardening items needed before returning to feature work:

```text
S20-01 README current status updated.
S20-02 Sprint 19 changelog updated after CI rerun.
S20-03 v0.19.0-vietnamese-ui-localization pushed after main CI was green.
S20-04 required-migration now includes migration reapply.
S20-05 GitHub Actions now force Node 24 runtime compatibility.
S20-06 API route registration split by module while preserving public paths.
S20-07 Production web runtime blocks prototype fallback masking for backend-backed services.
S20-08 This production runtime checklist captures env, persistence, fallback, smoke, and release gates.
```

Known remaining production-readiness risk:

```text
Frontend auth screens still use the local mock session surface. Treat backend auth/session persistence as available, but do not call the web auth surface production-ready until it is explicitly wired to the backend auth API or accepted as a controlled staging-only mock.
```
