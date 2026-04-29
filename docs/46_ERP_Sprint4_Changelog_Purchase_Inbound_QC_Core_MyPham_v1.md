# 46_ERP_Sprint4_Changelog_Purchase_Inbound_QC_Core_MyPham_v1

Project: Web ERP for cosmetics operations
Phase: Phase 1
Sprint: Sprint 4 - Purchase Order + Inbound QC Full Flow
Date: 2026-04-29
Status: Dev/main merged; production tag hold until GitHub Actions billing blocker is cleared

---

## 1. Summary

Sprint 4 completed the inbound operations core:

```text
Purchase Order
-> supplier delivery / goods receipt
-> batch, lot, expiry, packaging capture
-> inbound QC PASS / FAIL / HOLD / PARTIAL
-> controlled stock movement
-> supplier rejection / return-to-supplier
-> Warehouse Daily Board inbound signals
```

The implementation keeps the Sprint 4 guardrail that receiving goods does not automatically make stock available. Available stock is created only through QC-controlled stock movements.

## 2. Merged PRs

Core planning and release-gate work:

```text
#228 docs(S4-00-00): finalize sprint 4 task board
#229 docs(S4-00-02): record postgres migration verification
#230 docs(S4-00-03): map sprint 3 runtime store persistence
#231 docs(S4-00-04): record sprint 4 kickoff checklist
#232 feat(S4-00-05): add sprint 4 purchase finance roles
```

Purchase order and receiving:

```text
#233 feat(S4-01-01): add purchase order domain model
#234 feat(S4-01-02): add purchase order migration
#235 feat(S4-01-03): add purchase order API
#236 fix(S4-01-03): restore purchase order build
#237 feat(S4-01-04): add purchase order UI
#238 test(S4-01-05): cover purchase order permission audit
#239 feat(S4-02-01): harden goods receiving model
#240 feat(S4-02-02): validate PO-linked goods receiving
#241 fix(S4-02-02): avoid prototype batch supplier lock
#242 feat(S4-02-03): add PO-linked goods receiving UI
```

Inbound QC, stock movement, and supplier rejection:

```text
#243 feat(S4-03-01): add inbound QC inspection model
#244 feat(S4-03-02): add inbound QC inspection API
#245 feat(S4-03-03): add inbound QC UI
#246 feat(S4-04-01): record stock movement on inbound QC pass
#247 feat(S4-04-02): quarantine failed inbound QC stock
#248 feat(S4-04-03): sync inbound QC with batch status
#249 test(S4-04-04): cover inbound QC stock availability
#250 feat(S4-05-01): add supplier rejection model
#251 feat(S4-05-02): add supplier rejection API
#252 feat(S4-05-03): add supplier rejection UI
#253 test(S4-05-04): cover supplier rejection e2e
```

Attachments, daily board, OpenAPI, and release evidence:

```text
#254 feat(S4-06-01): store return attachments in object storage
#255 feat(S4-06-02): reuse attachment panel in inbound UI
#256 test(S4-06-03): cover attachment upload security
#257 feat(S4-07-01): expose inbound daily board metrics
#258 feat(S4-07-02): show inbound metrics on daily board
#259 test(S4-07-03): regress inbound daily board source counts
#260 docs(S4-08-01): document sprint 4 api endpoints
#261 feat(S4-08-02): regenerate sprint 4 frontend api client
#262 test(S4-08-03): add sprint 4 openapi contract check
#263 test(S4-09-01): add inbound pass E2E smoke
#264 test(S4-09-03): add partial inbound QC E2E
#265 test(S4-09-04): assert denied actions skip audit
```

S4-09-02 is covered by the existing supplier rejection E2E, which verifies receiving -> QC FAIL -> no available stock -> supplier rejection -> audit.

## 3. Verification

Backend verification on dev server:

```text
go test ./cmd/api -run TestInboundPassReceivingE2ESmoke -count=1
go test ./cmd/api -run TestSupplierRejectionFailReceivingE2ESmoke -count=1
go test ./cmd/api -run TestInbound.*Receiving -count=1
go test ./cmd/api -run Requires -count=1
go test ./... && go vet ./...
```

Frontend and OpenAPI verification run during S4-08:

```text
pnpm --filter web test
pnpm --filter web typecheck
pnpm --filter web build
pnpm --package=@redocly/cli dlx redocly lint packages/openapi/openapi.yaml
pnpm openapi:contract
```

Runtime migration verification on isolated PostgreSQL 16:

```text
12 up migrations applied successfully.
12 down migrations rolled back successfully.
Evidence: docs/qa/S4-00-02_PostgreSQL16_Migration_Runtime_Verification_2026-04-29.md
```

Dev deployment status:

```text
Dev server main synced to latest merge.
Runtime deploy was previously run after runtime changes and smoke passed.
Later S4-09 changes are backend test/docs only, so no additional runtime deploy was required.
```

## 4. Release Gate Status

Green:

```text
Local/dev backend checks
Frontend tests, typecheck, and build during S4-08
OpenAPI lint and contract check during S4-08
PostgreSQL 16 migration apply/rollback
Task branch -> PR -> manual self-review -> manual merge workflow
```

Blocked:

```text
GitHub Actions cloud CI is blocked by account billing/spending-limit.
Production tag v0.4.0-purchase-inbound-qc-core is on hold until CI can be rerun green.
```

## 5. Known Carry-Forward

```text
S4-00-01: Repo owner must clear GitHub Actions billing/spending-limit.
After billing is fixed: rerun full CI on main.
After CI is green: tag v0.4.0-purchase-inbound-qc-core.
```
