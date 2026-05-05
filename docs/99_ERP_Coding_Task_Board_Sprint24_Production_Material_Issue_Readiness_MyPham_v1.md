# 99_ERP_Coding_Task_Board_Sprint24_Production_Material_Issue_Readiness_MyPham_v1

Project: Web ERP for cosmetics operations
Phase: Phase 1
Sprint: Sprint 24 - Production Material Issue and Subcontract Readiness
Document role: Selected task board for the next implementation sprint after Stock Transfer and Warehouse Issue runtime
Version: v1
Date: 2026-05-05
Status: Runtime implementation merged and deployed to dev; release tag hold
Previous sprint: Sprint 23 - Production planning, purchase traceability, supplier invoice/payment gate, stock transfer, and warehouse issue runtime

---

## 1. Executive Summary

Sprint 24 should connect the production plan material-demand surface to the first-class Warehouse Issue Note runtime added in file 98.

The business flow should become:

```text
Production Plan
-> Material Demand
-> Purchase missing material if needed
-> Receive and QC material
-> Create Warehouse Issue Note from ready demand lines
-> Submit / Approve / Post issue
-> Mark production plan material issue readiness
-> Allow subcontract order creation when required material is issued
```

The sprint must not turn shortage into issue. A material line can create a Warehouse Issue Note only when the required quantity is available for issue.

---

## 2. Sprint Goal

Sprint 24 goal:

```text
Let users issue ready raw material and packaging directly from a production plan, preserve traceability between plan demand lines and warehouse issue lines, and gate subcontract readiness on posted material issue evidence.
```

The sprint succeeds when:

```text
1. A production plan detail page shows material issue readiness per demand line.
2. Ready material lines can create a Warehouse Issue Note without re-keying SKU, batch, UOM, or quantity.
3. Shortage lines cannot create issue lines until available stock is sufficient or the user intentionally issues a partial quantity.
4. Warehouse Issue Note stores source production plan and source material demand line references.
5. Posted issue movements update stock and can be traced back to the production plan.
6. Production plan worklist/status shows material issue progress.
7. Subcontract order creation is gated by posted material issue readiness.
```

---

## 3. Primary References

| Ref | Document / Source | Use |
| --- | --- | --- |
| `92_ERP_Coding_Task_Board_Sprint23_Production_Planning_Material_Demand_MyPham_v1.md` | Production planning/material demand board | Source production plan and demand calculation behavior |
| `94_ERP_Purchase_Request_Workflow_Production_Plan_PO_Traceability_MyPham_v1.md` | Purchase Request workflow | Shortage-to-purchase boundary |
| `95_ERP_PO_Receiving_QC_Supplier_Payable_Flow_MyPham_v1.md` | PO/receiving/QC/payable flow | Material becomes available only after accepted receiving/QC behavior |
| `98_ERP_Stock_Transfer_Warehouse_Issue_Runtime_Flow_MyPham_v1.md` | Warehouse Issue runtime | Issue document lifecycle, API, movement posting, and persistence |
| `100_ERP_Production_Material_Issue_Subcontract_Readiness_Flow_MyPham_v1.md` | Sprint 24 flow design | Detailed source-link, readiness, and gating rules |
| `101_ERP_Sprint24_Changelog_Production_Material_Issue_Readiness_MyPham_v1.md` | Sprint 24 changelog | PR, CI, deploy, smoke, known limits, and tag status |
| `88_ERP_BOM_Formula_Module_Design_MyPham_v1.md` | Formula design | Material demand line and UOM/decimal handling |
| `47_ERP_Coding_Task_Board_Sprint5_Subcontract_Manufacturing_Core_MyPham_v1.md` | Subcontract manufacturing core | External factory execution boundary |
| `81_ERP_Vietnamese_UI_Glossary_Operational_Copy_MyPham_v1.md` | Vietnamese glossary | User-facing copy |

---

## 4. In Scope

```text
- Production plan material issue readiness model
- Source link from production plan demand line to Warehouse Issue Note line
- Create Warehouse Issue Note from ready demand lines
- Partial issue support only if explicitly visible as partial, never hidden as complete
- Issue status rollup on production plan
- Plan detail timeline step for material issue
- Worklist action: open/create related issue note
- Gate subcontract order readiness on posted issue evidence
- OpenAPI/API/client updates if runtime implementation changes endpoints
- Dev smoke for plan -> issue note -> post -> readiness gate
```

---

## 5. Out Of Scope

```text
- Costing / gia thanh
- General ledger posting
- Automatic supplier PO creation
- Automatic warehouse receiving
- Automatic QC pass
- Internal MES / work centers
- Factory PO/email dispatch
- Finished goods receipt and finished goods QC changes
- Two-step in-transit transfer accounting
- Stock reservation rewrite
```

Important boundary:

```text
Sprint 24 issues material to production/subcontract execution.
It does not calculate final product cost and does not create finished goods stock.
```

---

## 6. Guardrails

```text
1. Do not create Warehouse Issue Note from shortage quantity that is not available.
2. Do not mark a production plan ready for subcontract until required issue quantity is posted or explicitly waived.
3. Do not let Warehouse Issue Note line quantities exceed available stock unless an approved negative-stock policy exists; no such policy is approved for Phase 1.
4. Do not let QC_HOLD or QC_FAIL material become issue-ready.
5. Preserve source_document_type = production_plan on issue lines created from production plan.
6. Preserve source_document_id = production plan id.
7. Preserve source_document_line_id = material demand line id.
8. Posted Warehouse Issue Note is immutable.
9. API routes, DB fields, enums, permission keys, and audit actions remain English technical contracts.
10. UI labels and operational copy are Vietnamese-first.
```

---

## 7. Dependency Map

```text
S24-00 Sprint boundary and docs
  -> S24-01 Production plan material issue readiness model
  -> S24-02 Source-linked Warehouse Issue creation
  -> S24-03 Production plan timeline/worklist updates
  -> S24-04 Subcontract readiness gate
  -> S24-05 UI smoke, API smoke, docs, changelog

S24-02 Source-linked Warehouse Issue creation
  -> posted inventory movement traceability
  -> future costing input

S24-04 Subcontract readiness gate
  -> future factory order dispatch
  -> future finished goods receipt/QC
```

---

## 8. Task Board

| Task ID | Priority | Owner | Task | Output / Acceptance | Primary Ref |
| --- | ---: | --- | --- | --- | --- |
| S24-00-01 | P0 | PM/BA | Confirm issue-ready rule | Issue can be created only from available stock; shortage continues to PR/PO/receiving/QC | `100`, section 2 |
| S24-00-02 | P0 | PM/BA | Confirm partial issue behavior | Partial issue must show `partial`, `remaining`, and cannot mark full readiness | `100`, section 5 |
| S24-00-03 | P0 | PM/BA | Confirm subcontract gate | Subcontract creation requires posted material issue or explicit waiver | `100`, section 6 |
| S24-01-01 | P0 | BE | Add material issue readiness calculation | Plan returns required, available, already-issued, remaining-to-issue, shortage, and readiness status per demand line | `92`, `98` |
| S24-01-02 | P0 | BE | Add issued quantity lookup | Posted Warehouse Issue lines linked to production plan roll up to plan demand lines | `98`, section 2.2 |
| S24-01-03 | P0 | BE | Add readiness status tests | Tests cover ready, shortage, partial, issued, and waived states | `100`, section 5 |
| S24-02-01 | P0 | BE | Create issue note from plan demand | API/service creates Warehouse Issue Note lines from selected ready demand lines | `100`, section 4 |
| S24-02-02 | P0 | BE | Preserve source references | Created issue lines store plan id and demand line id and expose them in API response | `100`, section 4 |
| S24-02-03 | P0 | BE | Block issue from shortage | API returns validation error when issue quantity exceeds available issue-ready quantity | `100`, section 5 |
| S24-02-04 | P1 | BE | Support partial issue intentionally | API accepts an issue quantity less than remaining demand and leaves plan line partial | `100`, section 5 |
| S24-03-01 | P1 | FE | Plan detail issue action | Production plan detail shows create/open issue action per ready line and bulk action for selected ready lines | `100`, section 7 |
| S24-03-02 | P1 | FE | Plan timeline issue step | Plan timeline shows issue note status and links to Warehouse Issue Note detail/list | `100`, section 7 |
| S24-03-03 | P1 | FE | Worklist readiness display | Production worklist shows material issue readiness and next action | `92`, section 10 |
| S24-04-01 | P0 | BE | Subcontract readiness gate | Creating subcontract order from a plan is blocked until material issue readiness is complete or waived | `47`, `100` |
| S24-04-02 | P1 | BE/FE | Waiver placeholder | If implemented, waiver requires reason, actor, audit log, and visible status; otherwise waiver remains out of runtime scope | `100`, section 6 |
| S24-05-01 | P0 | QA | API smoke | Create plan, create linked issue, submit/approve/post issue, verify plan readiness updates | This board |
| S24-05-02 | P0 | QA | UI smoke | Browser smoke covers plan detail issue action, issue link, posted status, and readiness gate | This board |
| S24-05-03 | P0 | DevOps | Required checks | api/web/openapi/migration checks green for code-change PRs | `README.md` |
| S24-05-04 | P0 | PM | Sprint 24 changelog | Changelog records PRs, checks, deploy, smoke, known limits, and tag status | `80` |

---

## 9. Demo Script

```text
1. Login as admin or production user.
2. Open Production.
3. Open a production plan with calculated material demand.
4. Review each demand line: required, available, issued, remaining, status.
5. Select one ready raw-material or packaging line.
6. Create Warehouse Issue Note from the selected line.
7. Open the created issue note.
8. Confirm source production plan and demand line link.
9. Submit, approve, and post the issue note.
10. Return to the production plan.
11. Confirm issued quantity and readiness status updated.
12. Try creating subcontract order before all required issue is complete; confirm gate blocks it.
13. Complete remaining issue or record approved waiver if waiver is in scope.
14. Confirm subcontract order action becomes available.
```

Expected:

```text
Material issue is source-linked, posted, and visible from the production plan.
Shortage cannot be issued.
Subcontract readiness depends on posted issue evidence.
```

---

## 10. Permissions

Recommended first-cut permissions using existing practical role pattern:

```text
production_plan:view
production_plan:create_issue
warehouse_issue:view
warehouse_issue:create
warehouse_issue:submit
warehouse_issue:approve
warehouse_issue:post
subcontract_order:create
```

Role direction:

| Role | Access |
| --- | --- |
| ERP Admin | Full |
| Production/Subcontract user | View plan, request issue, open related issue/subcontract |
| Warehouse staff | Create/submit issue |
| Warehouse manager | Approve/post issue |
| Purchase user | View shortage and purchase context |
| QC user | View readiness impact from QC hold/pass/fail |
| Finance observer | View issue evidence for future costing, no posting control |

---

## 11. Verification Gates

For code-change PRs:

```text
make api-test
make api-lint
make web-test
make web-lint
make openapi-validate
make openapi-contract
```

If local workstation lacks tools, run equivalent Docker/dev-host checks and record exact commands.

Required dev smoke after runtime/UI changes:

```text
./infra/scripts/deploy-dev-staging.sh dev
make smoke-dev
browser smoke for /production and related Warehouse Issue surface
```

Minimum Sprint 24 smoke:

```text
production_plan_material_issue_ready
warehouse_issue_created_from_plan
warehouse_issue_submit_approve_post
production_plan_readiness_updates_after_post
subcontract_gate_blocks_before_issue
subcontract_gate_allows_after_issue_or_waiver
```

---

## 12. Definition Of Ready

Sprint 24 is ready to start when:

```text
1. File 99 and file 100 are merged to main.
2. Current main is green and deployable.
3. Business accepts that shortage cannot be issued.
4. Business accepts the partial issue rule.
5. Business accepts the subcontract readiness gate.
6. No v0.24 tag is planned until runtime evidence exists.
```

---

## 13. Definition Of Done

Sprint 24 is done when:

```text
1. Production plan material issue readiness is implemented.
2. Warehouse Issue Note can be created from ready plan demand lines.
3. Issue source links back to production plan and material demand line.
4. Posted issue updates stock movements and plan readiness rollup.
5. Subcontract order creation is gated by material issue readiness.
6. CI/checks are green.
7. Dev deploy passes.
8. API and browser smoke pass.
9. Sprint 24 changelog records evidence and known limits.
10. v0.24 tag is created only if release evidence is complete; otherwise tag hold is documented.
```

---

## 14. Release And Tag Rule

Recommended checkpoint only after completed implementation and verification:

```text
v0.24.0-production-material-issue-readiness
```

Tag hold rule:

```text
Do not create v0.24 while Sprint 24 is only opened as documentation.
Do not create v0.24 until code is merged, CI is green, dev deploy and smoke pass, changelog evidence is complete, and known limits are recorded.
```

---

## 15. One-Line Rule

```text
Sprint 24 turns production material demand into controlled warehouse issue evidence before subcontract execution.
```

---

## 16. Current Evidence Status

```text
Runtime PR: #586 Add production material issue readiness runtime, merged at 9e28c05e.
Dev runtime fix PR: #587 Use same-origin API base for dev web, merged at 114105b2.
Required CI for #586 passed: api, web, openapi, e2e, required-api, required-web, required-openapi, required-migration.
Dev deploy passed: ./infra/scripts/deploy-dev-staging.sh dev.
Full dev smoke passed.
Browser smoke passed for /production and /production/plans/{plan_id}.
Screenshot evidence:
- output/playwright/s24-production-list.png
- output/playwright/s24-production-detail.png
Release tag: hold; no v0.24 tag created.
```
