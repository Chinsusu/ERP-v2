# 56_ERP_Sprint9_Changelog_System_Hardening_Production_Readiness_Core_MyPham_v1

Project: Web ERP for cosmetics operations
Phase: Phase 1
Sprint: Sprint 9 - System hardening / production readiness core
Date: 2026-05-01
Status: Dev/main merged; dev release gate green; production tag not created in this task

---

## 1. Summary

Sprint 9 hardened the release and runtime readiness path:

```text
Release process
-> held release tag ledger
-> dev disk preflight
-> consolidated full dev smoke
-> deploy evidence script
-> end-to-end dev release gate

Auth and permission confidence
-> auth/session policy inventory
-> local-only static token hardening
-> reporting route permission regression
-> audit/permission regression rollup

Runtime persistence
-> prototype runtime store inventory
-> stock movement store selection for PostgreSQL runtime
-> safe runtime routing for prototype/text-id movements
-> SQL decimal, UUID, and audit JSON argument fixes
-> deterministic persisted stock movement smoke fixture
```

Sprint 9 kept the hardening guardrails:

```text
- Do not weaken auth, session, permission, or audit behavior to pass smoke.
- Do not write stock balances directly.
- Persisted stock movement writes remain behind the stock movement service.
- API money, quantity, and rate boundaries remain decimal-string based.
- Dev release evidence separates local/dev verification from tag status.
- Production tags are created only after explicit tag target and release evidence are accepted.
```

## 2. Merged PRs

Planning, release ledger, and fixture stabilization:

```text
#370 docs(S9-00-00): add sprint 9 task board
#371 test(S9-00-01): stabilize operations report fixture date
#372 docs(S9-00-01): add held release gate ledger
```

OpenAPI wording, disk preflight, and full dev smoke:

```text
#373 chore(S9-01-01): generalize openapi contract check
#374 chore(S9-01-01): update make openapi contract target
#375 chore(S9-01-02): add dev verification disk preflight
#376 chore(S9-01-03): add full dev smoke script
```

Auth/session and permission hardening:

```text
#377 docs(S9-02-01): inventory auth session policy
#378 feat(S9-02-02): restrict static auth token to local envs
#379 test(S9-02-03): cover report route permissions
```

Runtime store persistence:

```text
#380 docs(S9-03-01): inventory prototype runtime stores
#381 feat(S9-03-02): select persistent stock movement store
#382 fix(S9-03-02): route stock movements safely at runtime
#383 fix(S9-03-02): encode stock movement SQL args safely
#384 fix(S9-03-02): cast stock movement audit JSON args
#385 test(S9-03-03): smoke persisted stock movement path
```

Regression rollup and release evidence:

```text
#386 chore(S9-04-01): add audit permission regression rollup
#387 chore(S9-05-01): add dev deploy evidence script
#388 chore(S9-06-01): add dev release gate script
```

## 3. Verification

Latest `main` cloud CI observed after Sprint 9 merge set:

```text
commit: dc76ba4ab387b1353bcbda693e2f58a1afa0249e
workflow: required-ci
run id: 25199914071
result: success
```

S9-06-01 full dev release gate run from `/opt/ERP-v2` on `main`:

```text
SMOKE_BASE_URL=http://127.0.0.1:8088 ./infra/scripts/dev-release-gate.sh dev
```

Release gate result:

```text
Dev release gate passed
```

Backend checks included:

```text
gofmt check
go vet ./...
go test ./...
go build ./cmd/api
go build ./cmd/worker
```

OpenAPI checks included:

```text
redocly lint packages/openapi/openapi.yaml
pnpm openapi:contract
openapi-typescript dry run to /tmp/erp-openapi-schema.ts
```

Latest observed OpenAPI contract result:

```text
OpenAPI route/envelope contract check passed: 70 routes and 38 envelopes.
```

`redocly lint` passed with the existing proprietary-license warning:

```text
License object should contain one of the fields: url, identifier.
```

Frontend checks included:

```text
pnpm --filter web typecheck
pnpm --filter web test
pnpm --filter web build
```

Latest observed frontend test result:

```text
35 test files passed.
226 tests passed.
```

Full dev smoke output captured by the release gate:

```text
healthz                      200 41
api_health                   200 123
login                        200 855
warehouse_fulfillment        200 276
warehouse_inbound            200 446
warehouse_subcontract        200 287
finance_dashboard            200 688
inventory_report_json        200 2448
inventory_report_csv         200 438
operations_report_json       200 451
operations_report_csv        200 140
finance_report_json          200 5909
finance_report_csv           200 856
stock_adjustment_create      201 885
stock_adjustment_submit      200 959
stock_adjustment_approve     200 1026
stock_adjustment_post        200 1088
persisted_stock_movement     ok ADJ-S9-03-03-SMOKE-0009
Full ERP dev smoke passed
```

Audit/permission rollup verification from S9-04-01:

```text
go test ./cmd/api ./internal/shared/auth ./internal/shared/audit -count=1
pnpm --filter web test -- src/modules/reporting/services/reportAccessRegression.test.ts src/shared/permissions/menu.test.ts src/modules/audit/services/auditLogService.test.ts
```

## 4. Dev Deployment Status

Latest deployed `main` commit:

```text
dc76ba4ab387b1353bcbda693e2f58a1afa0249e
```

Latest dev evidence health result:

```text
healthz          200 http://127.0.0.1:8088/healthz
api_health       200 http://127.0.0.1:8088/api/v1/health
web_root         307 http://127.0.0.1:8088/
```

Latest dev container state:

```text
api             Up, healthy
web             Up, healthy
worker          Up
reverse-proxy   Up, healthy
postgres        Up, healthy
redis           Up, healthy
minio           Up
mailhog         Up
```

The GHCR image pull warnings are still expected in the current dev setup; the deploy script builds service images from source and smoke checks pass.

## 5. Release Gate Status

Green:

```text
- All Sprint 9 task PRs are merged to main.
- Latest main required-ci is green.
- Backend gofmt/vet/test/build pass in the dev release gate.
- OpenAPI lint, route/envelope contract, and generated-client dry run pass in the dev release gate.
- Frontend typecheck/test/build pass in the dev release gate.
- Dev deployment and full smoke pass.
- Persisted stock movement smoke writes and verifies a PostgreSQL stock_ledger row.
- Audit/permission regression rollup is documented and runnable.
```

Hold / not done in this task:

```text
- Production tag v0.9.0-system-hardening-production-readiness-core is not created.
- Held Sprint 5, Sprint 6, Sprint 7, and Sprint 8 tags are still absent.
- Historical held tags still need explicit tag target selection before backfill.
```

Candidate Sprint 9 tag command, only after the team accepts this release evidence:

```bash
git checkout main
git pull --ff-only origin main
git tag v0.9.0-system-hardening-production-readiness-core dc76ba4ab387b1353bcbda693e2f58a1afa0249e
git push origin v0.9.0-system-hardening-production-readiness-core
```

## 6. Known Notes

```text
- Earlier Sprint 5-8 changelogs describe GitHub Actions as blocked; Sprint 9 latest main required-ci is now green.
- The Redocly license warning is pre-existing and non-fatal.
- The highest-risk persisted runtime store completed in Sprint 9 is stock movement recording; many other prototype stores remain documented for future persistence work.
- Runtime stock movement routing persists DB-compatible UUID movement flows and keeps prototype/text-id flows in memory to avoid breaking existing demo/dev paths.
- Full dev smoke intentionally creates deterministic stock adjustment smoke records.
```

## 7. Next Step

```text
Decide whether to create the Sprint 9 production tag.
Select exact target commits before backfilling held Sprint 5-8 tags.
Plan the next sprint from the release-gated main state.
```
