# S4-00-04 Sprint 4 Kickoff Checklist

Task: S4-00-04
Date: 2026-04-29
Owner: PM / QA
Verifier: Codex

## Kickoff Decision

Sprint 4 is cleared to start implementation work after the Sprint 3 release-gate hardening tasks below:

```text
S4-00-01 GitHub Actions billing/spending-limit blocker: blocked by account owner action
S4-00-02 PostgreSQL 16 migration apply/rollback verification: done
S4-00-03 Sprint 3 runtime store persistence map and risk-owner documentation: done
```

Production release tagging for Sprint 3 remains on hold until GitHub Actions can run in the cloud.

## Workflow Confirmed

The current repo workflow remains:

```text
task branch
-> local/dev-server verification
-> commit
-> push
-> PR
-> manual self-review
-> admin/manual merge into main
-> delete feature branch
```

Rules:

```text
Branch pattern: feature/S4-xx-yy-short-task-name
Merge target: main
GitHub auto-review: off
GitHub auto-merge: off
Cloud CI failure caused by billing/spending-limit must be commented on the PR with local verification evidence.
```

## Sprint 4 Scope Confirmed

Sprint 4 builds:

```text
Purchase Order
-> Supplier delivery
-> Goods receiving
-> Batch/lot/expiry/package checks
-> Inbound QC
-> QC PASS / FAIL / HOLD / PARTIAL
-> Controlled stock movement
-> Available stock / quarantine / return-to-supplier
-> Warehouse Daily Board inbound signals
```

Sprint 4 starts with backend state and API contracts before frontend screens where the backend state is not stable yet.

## Non-Goals Confirmed

Sprint 4 does not include:

```text
Full AP accounting
Supplier portal
Automated supplier EDI/API integration
Advanced landed cost allocation
Full subcontract manufacturing flow
Production planning/MRP
Multi-currency purchasing
Full tax accounting
```

## Known Blockers / Risks

| Blocker / Risk | Status | Owner | Required action |
|---|---|---|---|
| GitHub Actions billing/spending-limit | Blocked | Repo owner | Fix account billing/spending-limit and rerun required-ci/e2e |
| Sprint 3 production tag | Hold | PM / Tech Lead | Tag only after cloud CI is green |
| Sprint 3 runtime stores prototype-only | Documented | BE / Tech Lead | Implement DB-backed stores before production use |
| API DB connection not wired | Open | BE | Add `DATABASE_URL` connection path before PostgreSQL stores can be used |
| GHCR image pull denied on dev deploy | Known | DevOps | Continue source build fallback or fix GHCR package access |
| OpenAPI proprietary license warning | Known warning | Tech Lead | Add license URL/identifier if required by API governance |
| Purchase/QC role split not finalized | Open | BE / Tech Lead | Complete S4-00-05 RBAC mapping before PO approval/QC implementation |

## Demo Script Ready For Acceptance

Demo cases for Sprint 4:

```text
Case 1: Goods pass
Create PO -> receive supplier delivery -> QC PASS -> inbound movement -> available stock increases -> Daily Board updates

Case 2: Goods fail
Create PO -> receive supplier delivery -> QC FAIL -> no available stock -> return-to-supplier record -> audit exists

Case 3: Partial receive/QC
Create PO 100 -> receive 80 -> pass 70 / hold 10 -> available 70 / quarantine 10 / PO pending 20
```

Business-owner acceptance signal:

```text
The business owner supplied the Sprint 4 backlog and instructed the agent to continue tasks sequentially on the dev server.
No product decision is currently blocking kickoff.
```

This is build-start acceptance, not final UAT acceptance.

## Verification Performed

Dev server baseline after deploy:

```text
main is synced
dev stack deployed on http://10.1.1.120:8088
reverse proxy health passed
API health passed
web /login returns HTTP 200
```

Build/test already executed on dev server before Sprint 4 kickoff:

```text
go test ./...
Redocly OpenAPI validation
pnpm --filter web lint
pnpm --filter web test
pnpm --filter web build
docker compose build api worker web
dev deploy smoke
```

## Next Task

Proceed to:

```text
S4-00-05 Sprint 4 RBAC role mapping
```
