# 125_ERP_Sprint32_Changelog_Factory_Finished_Goods_QC_Closeout_MyPham_v1

Project: Web ERP for cosmetics operations
Phase: Phase 1
Sprint: 32
Change type: Runtime UI bridge and QC closeout helper
Version: v1
Date: 2026-05-06
Status: PR CI passed; manual merge and post-merge dev smoke pending

---

## 1. Summary

Sprint 32 adds the production-facing QC closeout step for external-factory finished goods after receipt into QC hold.

The target user-facing route is:

```text
/production/factory-orders/:orderId#factory-finished-goods-qc-closeout
```

---

## 2. Changed Runtime Surface

Implemented branch scope:

```text
1. QC closeout helper for full pass, partial pass, and full fail decisions.
2. Web service bridge for existing accept, partial-accept, and report-factory-defect APIs.
3. Prototype fallback for local no-backend mode: receipt-backed QC release movement, partial factory claim, and full factory defect claim.
4. Factory order detail section for QC closeout on /production/factory-orders/:orderId#factory-finished-goods-qc-closeout.
5. Tracker and timeline QC links to the production-facing QC section.
6. README and master index updated to include Sprint 32 docs and current scope.
```

---

## 3. Guardrails

```text
Receipt to QC hold remains separate from QC pass.
Only accepted quantity can become available stock.
Rejected quantity opens factory claim.
Final payment readiness remains a later separate action.
No email/Zalo/factory portal/API delivery is included.
No internal MES/work-center production is included.
```

---

## 4. Verification Evidence

Runtime PR:

```text
PR: #604 Add factory finished goods QC closeout
Merge commit: pending
```

Local verification:

```text
git diff --check: pass
Targeted Vitest: blocked locally; pnpm is not installed and direct node/vitest execution returns Windows "Access is denied"
make web-test / web-lint / api-test: blocked locally because make is not installed
go test ./...: blocked locally because go is not installed
Full web/API/OpenAPI verification: completed through GitHub CI for PR #604
```

GitHub CI:

```text
PR #604 required checks passed:
- required-api
- required-web
- required-openapi
- required-migration
- web
- e2e
```

Dev deploy and smoke:

```text
pending
```

---

## 5. Release Tag

No `v0.32` tag is planned by this sprint branch.

Sprint 32 is a Phase 1 runtime milestone, but release tagging remains held until target staging/pilot readiness evidence is explicitly approved.
