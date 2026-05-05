# 105_ERP_Coding_Task_Board_Sprint26_Production_IA_External_Factory_Order_Detail_MyPham_v1

Project: Web ERP for cosmetics operations
Phase: Phase 1
Sprint: Sprint 26 - Production IA cleanup and external factory order detail
Document role: Coding task board
Version: v1
Date: 2026-05-06
Status: Active implementation

---

## 1. Goal

Sprint 26 aligns the UI information architecture with the real business model:

```text
Production = user-facing module
External factory / subcontract = current production execution method
```

The company does not operate an internal factory in Phase 1. All finished goods production is sent to external factories. Therefore users should enter through **Production**, while the technical subcontract runtime can remain as an internal execution surface.

---

## 2. Scope

Implement:

```text
1. Sidebar shows only Production for the production area.
2. Subcontract remains a hidden technical/legacy route, not a primary sidebar entrypoint.
3. Production Plan worklist points users back to Production instead of /subcontract.
4. Production Plan detail related factory orders open a production-facing detail route.
5. Add /production/factory-orders/:orderId detail page.
6. Factory order detail shows summary, source Production Plan link, timeline, closeout state, and material lines.
7. Timeline covers factory confirmation, deposit, material issue, sample approval, mass production, finished goods receipt, QC, final payment, and close.
8. Document the updated production IA and evidence.
```

Do not implement:

```text
1. Internal MES/work-center production.
2. Full rewrite of existing subcontract operational forms.
3. Factory dispatch/export channel.
4. Costing/GL posting.
5. Release tag v0.26.
```

---

## 3. Acceptance Criteria

Frontend:

```text
- PRODUCTION_OPS sees Production in the sidebar and does not see Subcontract as a sibling entrypoint.
- /subcontract still resolves for hidden operational execution and backward compatibility.
- Production Plan worklist no longer deep-links users to /subcontract as the primary next step.
- Production Plan detail related factory orders open /production/factory-orders/:orderId.
- Factory order detail shows source plan, factory, product, planned/received/accepted quantities, deposit, final payment, timeline, and material lines.
```

Verification:

```text
- Menu permission regression test covers hidden Subcontract entrypoint.
- Production worklist regression test covers Production-facing next action.
- Factory order timeline test covers current/completed gates and links.
- Web typecheck, targeted tests, full web tests, and production build pass before PR.
- Dev deploy and browser smoke cover sidebar Production-only visibility and factory order detail page.
```

---

## 4. Status Log

```text
2026-05-06:
- Sprint opened on branch codex/s26-production-subcontract-ia.
- Business decision locked: Production is the user-facing module; external factory/subcontract is the current execution method.
- Initial implementation scope set to IA cleanup, hidden Subcontract sidebar entrypoint, Production-facing factory order detail page, and timeline.
```

---

## 5. Tag Status

```text
No v0.26 tag should be created unless a separate release checkpoint is requested.
Tag status: hold.
```
