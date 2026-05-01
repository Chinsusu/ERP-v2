# 55_ERP_Coding_Task_Board_Sprint9_System_Hardening_Production_Readiness_Core_MyPham_v1

Project: Web ERP for cosmetics operations
Phase: Phase 1
Sprint: Sprint 9 - System hardening / production readiness core
Document role: Coding task board for Sprint 9 implementation
Version: v1.0
Primary stack: Go backend, PostgreSQL, Next.js frontend, OpenAPI, Docker dev deploy
Status: Ready for implementation; production tags remain on hold until GitHub Actions billing blocker is cleared

---

## 1. Sprint 9 Context

Sprints 5 through 8 delivered the core operational and reporting flows:

```text
Subcontract manufacturing
-> Finance Lite / COD / AR / AP
-> Reporting v1
-> Reporting source drilldowns and dashboard links
```

Those sprints are merged and verified on the dev server, but the production release gate is not fully closed:

```text
- GitHub Actions cloud CI is blocked by the billing/spending-limit issue.
- Production tags for held releases must wait until CI can rerun green.
- Runtime verification is being done on the dev server while cloud CI is blocked.
- Some prototype/runtime-store and deploy-smoke gaps should be reduced before adding another large business module.
```

Sprint 9 exists to make the system easier to verify, deploy, and trust before Sprint 10 adds another major operational area.

---

## 2. Sprint 9 Theme

```text
System hardening + Production readiness core
```

Business reason:

```text
The ERP now has enough warehouse, purchasing, returns, subcontract, finance, and reporting surface area that the next risk is not missing screens.
The next risk is shipping changes that are hard to verify, hard to deploy, or backed by prototype-only runtime state.
```

Sprint 9 should improve confidence in:

```text
- Release gates and held production tags.
- Repeatable dev verification.
- OpenAPI contract checks and generated client stability.
- Auth/session and permission boundaries.
- Persistent runtime state for high-risk prototype stores.
- Deterministic seed data and smoke scenarios.
```

---

## 3. Sprint 9 Goals

By the end of Sprint 9, the system must support:

```text
1. Keep GitHub Actions billing/spending-limit blocker visible and rerun CI when unblocked.
2. Track held production tags for Sprint 5, Sprint 6, Sprint 7, and Sprint 8 until CI is green.
3. Rename or generalize OpenAPI contract check output so it no longer reports stale Sprint 4/5/6/7 wording.
4. Harden dev verification scripts to reduce disk-pressure failures and make smoke evidence repeatable.
5. Add a consolidated dev smoke check for health, auth, dashboard, report JSON, and CSV endpoints.
6. Inventory current auth/session behavior and close the highest-risk local-mock-only gaps without changing production secrets.
7. Strengthen route permission regressions for module pages, report tabs, and finance-only surfaces.
8. Inventory prototype/runtime stores and prioritize which stores must become PostgreSQL-backed first.
9. Persist the highest-risk remaining runtime store that affects operational correctness.
10. Make seed/dev data deterministic enough for repeated smoke tests.
11. Add release evidence that clearly says what was verified locally, what cloud CI could not verify, and why tags remain held.
```

---

## 4. Sprint 9 Non-Goals

Sprint 9 does not include:

```text
- New large business module scope.
- Full production security overhaul.
- SSO, OAuth, or external identity-provider integration.
- Row-level security redesign.
- Full observability platform.
- Full data warehouse or BI system.
- Accounting general ledger.
- Rewriting completed Sprint 5 through Sprint 8 flows.
- Destructive production migration work.
- Creating production tags while GitHub Actions is still blocked.
```

---

## 5. Branch / PR / Release Rules

Current repo workflow remains:

```text
task branch
-> build/test on dev server
-> PR
-> manual self-review comment
-> manual merge into main
-> sync/deploy dev server when runtime changes require it
```

Do not use GitHub auto review or auto merge.
Do not create a long-lived sprint branch unless the team explicitly changes that policy.

Default task branch pattern:

```text
codex/feature-S9-xx-yy-short-task-name
```

Recommended Sprint 9 release tag after completion:

```text
v0.9.0-system-hardening-production-readiness-core
```

Production tags remain on hold while GitHub Actions is blocked by the billing/spending-limit issue.
Held Sprint 5, Sprint 6, Sprint 7, and Sprint 8 production tags remain carry-forward blockers until CI can run green on GitHub and the release tags are created.

---

## 6. Sprint 9 Demo Script

### Case 1: Repeatable dev verification

```text
1. Run the consolidated dev smoke check after deployment.
2. Confirm API health, login/session, dashboard, inventory report, operations report, finance report, and CSV endpoints return expected status codes.
3. Confirm smoke output is compact enough to paste into changelog release evidence.
```

### Case 2: Permission boundary confidence

```text
1. Login with users that have different role scopes.
2. Confirm restricted module pages and finance/reporting surfaces stay blocked for unauthorized users.
3. Confirm authorized users can access only the expected operational and finance/reporting views.
```

### Case 3: Persistent high-risk runtime state

```text
1. Create or update an operational record that uses the selected high-risk runtime store.
2. Restart or redeploy the dev service.
3. Confirm the record is still present and appears in the relevant dashboard/reporting view.
4. Confirm stock, money, quantity, and audit invariants remain unchanged.
```

---

## 7. Sprint 9 Guardrails

These rules are non-negotiable:

```text
1. Do not tag production while GitHub Actions is blocked.
2. Do not hide cloud CI failures behind local verification.
3. Do not commit credentials, tokens, passwords, private keys, or private environment files.
4. Do not weaken auth, session, permission, audit, or report access gates to make smoke checks pass.
5. Do not mutate stock balance directly; stock changes must continue through stock movement services.
6. Money, quantity, and rate values remain decimal strings in APIs.
7. VND, vi-VN, and Asia/Ho_Chi_Minh rules from file 40 remain active.
8. Verification scripts must fail clearly when a required endpoint or permission gate fails.
9. Seed/dev data may be deterministic, but must not depend on production data.
10. Runtime-store hardening must preserve current API contracts unless a task explicitly updates OpenAPI and generated clients.
11. Cosmetic reformatting must stay out of behavioral diffs.
12. Every release evidence file must separate local verification from cloud CI status.
```

---

## 8. Dependency Map

```text
S9-00-00 Sprint 9 task board
  -> S9-00-01 held release gate ledger
  -> S9-01-01 OpenAPI contract check wording

S9-00-01 held release gate ledger
  -> S9-07-01 Sprint 9 release evidence

S9-01-01 OpenAPI contract check wording
  -> S9-01-03 consolidated dev smoke check
  -> S9-06-01 end-to-end release gate smoke

S9-01-02 dev verification disk hardening
  -> S9-01-03 consolidated dev smoke check
  -> S9-06-01 end-to-end release gate smoke

S9-02-01 auth/session policy inventory
  -> S9-02-02 login/session hardening
  -> S9-02-03 route permission regression

S9-03-01 prototype store inventory
  -> S9-03-02 persist highest-risk runtime store
  -> S9-03-03 deterministic seed/smoke fixtures

S9-04-01 audit/permission regression rollup
  -> S9-06-01 end-to-end release gate smoke

S9-05-01 dev deploy evidence script
  -> S9-06-01 end-to-end release gate smoke

S9-06-01 end-to-end release gate smoke
  -> S9-07-01 Sprint 9 release evidence
```

---

## 9. Task Board

| Task ID | Task | Output / Acceptance | Primary Ref |
| --- | --- | --- | --- |
| S9-00-00 | Sprint 9 task board | File 55 created, reviewed, merged to main | `54_ERP_Sprint8_Changelog_Reporting_Hardening_Dashboard_Drilldowns_MyPham_v1.md` |
| S9-00-01 | Held release gate ledger | Docs or release evidence clearly tracks held Sprint 5/6/7/8 tags, CI blocker, and rerun requirements | `54_ERP_Sprint8_Changelog_Reporting_Hardening_Dashboard_Drilldowns_MyPham_v1.md` |
| S9-01-01 | OpenAPI contract check wording | Contract check output no longer references stale Sprint 4/5/6/7 wording and remains route/envelope based | `16_ERP_API_Contract_OpenAPI_Standards_Phase1_MyPham_v1.md` |
| S9-01-02 | Dev verification disk hardening | Verification/deploy guidance cleans task-local temp clones/caches safely and reports disk state before expensive checks | `54_ERP_Sprint8_Changelog_Reporting_Hardening_Dashboard_Drilldowns_MyPham_v1.md` |
| S9-01-03 | Consolidated dev smoke check | Repo has a repeatable smoke command/script for health, login, dashboard, report JSON, and CSV endpoints | `24_ERP_QA_Test_Strategy_Automation_Phase1_MyPham_v1.md` |
| S9-02-01 | Auth/session policy inventory | Current login/session/API token behavior is documented with risks and next hardening targets | `19_ERP_Security_RBAC_Audit_Compliance_Standards_Phase1_MyPham_v1.md` |
| S9-02-02 | Login/session hardening | Highest-risk local-mock-only session behavior is tightened without committing secrets or breaking dev login | `19_ERP_Security_RBAC_Audit_Compliance_Standards_Phase1_MyPham_v1.md` |
| S9-02-03 | Route permission regression | Module pages, reporting tabs, and finance-only routes have focused permission regression coverage | `19_ERP_Security_RBAC_Audit_Compliance_Standards_Phase1_MyPham_v1.md` |
| S9-03-01 | Prototype store inventory | Remaining prototype/runtime stores are listed with risk level and persistence priority | `44_ERP_Sprint3_Changelog_Returns_Reconciliation_Core_MyPham_v1.md` |
| S9-03-02 | Persist highest-risk runtime store | One high-risk runtime store becomes PostgreSQL-backed or is adapted to an existing persistent store | `17_ERP_Database_Schema_PostgreSQL_Standards_Phase1_MyPham_v1.md` |
| S9-03-03 | Deterministic seed/smoke fixtures | Dev seed data supports repeatable smoke checks for login, dashboard, reports, and selected persisted store | `24_ERP_QA_Test_Strategy_Automation_Phase1_MyPham_v1.md` |
| S9-04-01 | Audit/permission regression rollup | Existing audit and permission tests are grouped or documented so release checks can run them deliberately | `19_ERP_Security_RBAC_Audit_Compliance_Standards_Phase1_MyPham_v1.md` |
| S9-05-01 | Dev deploy evidence script | Deploy evidence includes commit, health, smoke status, and compact report output for changelogs | `infra/scripts/deploy-dev-staging.sh` |
| S9-06-01 | End-to-end release gate smoke | Full dev-server gate runs backend, OpenAPI, frontend, deploy, and consolidated smoke checks | `24_ERP_QA_Test_Strategy_Automation_Phase1_MyPham_v1.md` |
| S9-07-01 | Sprint 9 release evidence | Changelog captures PRs, verification, dev deploy status, CI blocker, held tags, and tag/release status | `54_ERP_Sprint8_Changelog_Reporting_Hardening_Dashboard_Drilldowns_MyPham_v1.md` |

---

## 10. Verification Gates

Each implementation PR should run the smallest relevant checks plus broader checks when contracts, auth, scripts, or shared UI change.

Backend checks:

```text
go test ./...
go vet ./...
```

Frontend checks:

```text
pnpm --filter web test
pnpm --filter web typecheck
pnpm --filter web build
```

OpenAPI checks when API shapes or contract scripts change:

```text
pnpm openapi:validate
pnpm openapi:contract
pnpm openapi:generate
git diff --stat -- apps/web/src/shared/api/generated/schema.ts
```

Dev server checks when runtime behavior changes:

```text
deploy dev staging
API health smoke
auth/session smoke
dashboard smoke
inventory / operations / finance report JSON and CSV smoke
```

Release gate checks:

```text
GitHub Actions rerun after billing/spending-limit blocker is cleared
held production tags created only after cloud CI is green
```

---

## 11. Definition of Done

Sprint 9 is complete when:

```text
1. All task PRs are merged to main through manual review/merge.
2. Held release gate status is visible and accurate.
3. OpenAPI contract output uses current generic wording.
4. Dev verification and smoke checks are repeatable and documented.
5. Auth/session policy and route permission risks are reduced or explicitly documented.
6. The highest-risk selected runtime store is persisted or has a documented persistence plan with evidence.
7. Seed/smoke fixtures support repeatable dev checks.
8. Backend, frontend, OpenAPI, deploy, and consolidated smoke checks pass on the dev server.
9. Sprint 9 changelog is created with release evidence.
10. Production tag remains held if GitHub Actions is still blocked.
```
