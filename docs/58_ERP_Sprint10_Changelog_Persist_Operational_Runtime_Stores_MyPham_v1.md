# 58_ERP_Sprint10_Changelog_Persist_Operational_Runtime_Stores_MyPham_v1

Project: Web ERP for cosmetics operations
Phase: Phase 1
Sprint: Sprint 10 - Persist operational runtime stores v1
Date: 2026-05-01
Status: Dev/main merged; main CI green; dev release gate green; production tag not created in this task

---

## 1. Summary

Sprint 10 reduced the highest operational restart risk by moving runtime evidence and document stores from prototype memory stores to PostgreSQL-backed runtime paths when database configuration is present.

Persistence work completed:

```text
audit log runtime evidence
-> sales order reservation state
-> stock adjustment document lifecycle
-> stock count session lifecycle
-> warehouse receiving document lifecycle
-> inbound QC inspection/checklist lifecycle
-> dev smoke checks for persisted runtime paths
-> remaining prototype store ledger
```

Guardrails kept:

```text
- No direct stock balance writes.
- Stock changes continue through the stock movement service.
- Prototype fallback remains explicit for no-DB/local runtime paths.
- API response envelopes and decimal-string boundaries remain unchanged.
- Migration up/down coverage is required for each new persistent store.
- Frontend fallback state is not counted as backend persistence evidence.
```

## 2. Merged PRs

Planning and persistence design:

```text
#390 docs(S10-00-00): add sprint 10 task board
#391 docs(S10-01-01): design audit persistence
#394 docs(S10-02-01): design reservation persistence
```

Audit and reservation persistence:

```text
#392 feat(S10-01-02): persist audit log store
#393 chore(S10-01-03): add audit persistence smoke
#395 feat(S10-02-02): persist sales order reservations
#396 chore(S10-02-03): add reservation persistence smoke
```

Inventory document persistence:

```text
#397 feat(S10-03-01): persist stock adjustments
#398 feat(S10-03-02): persist stock counts
#399 chore(S10-03-03): add inventory document persistence smoke
```

Inbound persistence:

```text
#400 feat(S10-04-01): persist warehouse receivings
#401 feat(S10-04-02): persist inbound QC inspections
#402 test(S10-04-03): add inbound QC persistence smoke
```

Release ledger:

```text
#403 docs(S10-05-01): rank remaining prototype stores
```

## 3. Persisted Stores

PostgreSQL-backed when DB config exists:

```text
newRuntimeAuditLogStore
newRuntimeSalesOrderReservationStore
newRuntimeStockAdjustmentStore
newRuntimeStockCountStore
newRuntimeWarehouseReceivingStore
newRuntimeInboundQCInspectionStore
newRuntimeStockMovementStore
```

Prototype/no-DB fallback remains intentional for local or non-DB environments.

## 4. Migration Evidence

Sprint 10 migrations:

```text
000014_persist_audit_runtime_refs
000015_persist_sales_order_reservation_refs
000016_persist_stock_adjustments
000017_persist_stock_count_sessions
000018_persist_warehouse_receivings
000019_persist_inbound_qc_inspections
```

Latest main migration CI:

```text
commit: a4d05ae71e8f555fe7fb2e9cd11afbb747057104
workflow: required-ci
run id: 25206499089
result: success
jobs: required-migration applied migrations and rolled them back successfully
```

## 5. Verification

Latest `main` cloud CI observed after S10-05-01:

```text
commit: a4d05ae71e8f555fe7fb2e9cd11afbb747057104
workflow: required-ci
run id: 25206499089
result: success
```

Required CI jobs passed:

```text
required-api
required-web
required-openapi
required-migration
```

Latest S10 PR E2E check passed on #403:

```text
e2e
```

S10-06-01 full dev release gate run from `/opt/ERP-v2` on `main`:

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

## 6. Dev Deployment Status

Latest deployed `main` commit:

```text
a4d05ae71e8f555fe7fb2e9cd11afbb747057104
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

Full dev smoke persistence signals:

```text
persisted_audit_login        ok
persisted_sales_reservation  ok
persisted_stock_adjustment   ok
persisted_stock_movement     ok
persisted_stock_count        ok
persisted_inbound_qc         ok
Full ERP dev smoke passed
```

Latest captured smoke IDs:

```text
sales reservation: SO-S10-02-03-SMOKE-0015
stock adjustment:  ADJ-S9-03-03-SMOKE-0027
stock count:       CNT-S10-03-03-SMOKE-0011
inbound QC:        GRN-S10-04-03-SMOKE-0007
```

## 7. Remaining Prototype Stores

S10-05-01 re-ranked the remaining prototype stores in:

```text
docs/qa/S10-05-01_remaining_prototype_store_ledger.md
```

Next persistence order:

```text
1. Available stock read model from PostgreSQL stock_balances.
2. Sales order document store.
3. Purchase order document store.
4. Return receipt + supplier rejection stores.
5. Batch/QC status persistence or adapter to inventory.batches.
6. Finance AR/AP/COD/cash stores.
7. End-of-day reconciliation.
8. Shipping manifest/pick/pack stores.
9. Subcontract runtime stores.
10. Master data catalogs and auth/session hardening.
```

## 8. Release Gate Status

Green:

```text
- All Sprint 10 task PRs through S10-05-01 are merged to main.
- Latest main required-ci is green.
- Backend gofmt/vet/test/build pass in the dev release gate.
- OpenAPI lint, route/envelope contract, and generated-client dry run pass in the dev release gate.
- Frontend typecheck/test/build pass in the dev release gate.
- Dev deployment and full smoke pass on main.
- Migration apply/rollback passes on PostgreSQL 16 in CI.
- Persisted runtime smoke checks cover audit, reservation, stock adjustment, stock movement, stock count, and inbound QC.
- Remaining prototype stores are re-ranked.
```

Hold / not done in this task:

```text
- Production tag v0.10.0-persist-operational-runtime-stores is not created.
- Earlier held Sprint 5, Sprint 6, Sprint 7, Sprint 8, and Sprint 9 tags are still absent.
- Historical held tags still need explicit tag target selection before backfill.
```

Candidate Sprint 10 tag command, only after the team accepts this release evidence:

```bash
git checkout main
git pull --ff-only origin main
git tag v0.10.0-persist-operational-runtime-stores a4d05ae71e8f555fe7fb2e9cd11afbb747057104
git push origin v0.10.0-persist-operational-runtime-stores
```

## 9. Known Notes

```text
- Redocly still reports the existing proprietary-license warning; it is non-fatal.
- GitHub Actions reports a Node.js 20 actions deprecation warning for actions/checkout, actions/setup-node, and pnpm/action-setup. Checks still pass.
- GHCR dev image pull is still denied in this environment, then dev deploy builds images from source successfully.
- Available stock read model is the next highest persistence target because stock movement writes are persisted but the runtime query path still uses the prototype availability store.
```
