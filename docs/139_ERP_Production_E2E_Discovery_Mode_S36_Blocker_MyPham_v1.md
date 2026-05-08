# 139_ERP_Production_E2E_Discovery_Mode_S36_Blocker_MyPham_v1

Project: Web ERP for cosmetics operations
Area: Production / External Factory / Finance handoff
Document role: Current docs-only guide for Production E2E Discovery Mode and Sprint 36 blocker handling.
Status: Current discovery-mode guidance
Release tag: Not applicable
Business UAT decision: Not claimed

---

## 1. Decision summary

The project will proceed with **Production E2E Discovery / Controlled Walkthrough** before claiming full Production Business UAT pass.

This means:

```text
Production E2E Discovery: allowed
Business UAT Pass: not claimed
Release tag: not created
PFX-UAT-013: READY_TO_RUN after Sprint 36 PR #616 merge and dev/browser smoke
Sprint 22 Warehouse/Sales/QC Go-No-Go: still separate and pending
```

This decision is intentionally conservative. The goal is to expose workflow gaps early without pretending the flow is already business accepted.

---

## 2. Why this document exists

The current Production External-Factory E2E UAT pack defines a full end-to-end path:

```text
Production Plan
-> Material Demand
-> Purchase Request / PO / Receiving / QC
-> Warehouse Issue to Factory
-> Factory Dispatch
-> Material Handover
-> Sample Approval
-> Mass Production
-> Finished Goods Receipt to QC Hold
-> QC Closeout
-> Factory Claim / Final Payment Readiness
-> AP Handoff
-> Finance Closeout
-> Cash-out Voucher Evidence
```

The payment voucher / cash-out evidence scenario depended on Sprint 36 runtime implementation.

Sprint 36 is now merged and smoke-tested, so `PFX-UAT-013` can be treated as executable for controlled discovery mode. This still does not claim business UAT pass.

---

## 3. Business context

The production workflow for the cosmetics company is not internal MES-first. It is currently an **external factory / subcontract manufacturing workflow**.

The real-world flow includes:

```text
Order with factory
-> confirm quantity / specification / sample / production-receiving timing
-> deposit
-> transfer raw materials and packaging to factory
-> factory sample / approved sample
-> mass production
-> finished goods delivered back to warehouse
-> quantity and quality check
-> accepted goods go forward
-> rejected goods are reported back to factory within 3-7 days
-> final payment
```

ERP Phase 1 should model this as external-factory production, not as internal machine routing, work centers, labor time, or MES shopfloor execution.

---

## 4. Mode definitions

### 4.1. Mode A — Business UAT

Use this only when all required runtime dependencies are ready and business users execute the scenarios.

Mode A may result in:

```text
Go
Conditional Go
No-Go
```

Mode A requires:

```text
- Business users or delegated business owners execute scenarios.
- Required seed data is loaded and approved.
- All required scenarios are run or explicitly waived.
- P0/P1 issues are triaged.
- Go/Conditional Go/No-Go report is signed.
- Evidence pack is complete.
```

### 4.2. Mode B — Production E2E Discovery / Controlled Walkthrough

Use this now.

Mode B is allowed to run the production flow early to discover issues, but it must not be called business UAT pass.

Mode B rules:

```text
- Run PFX-UAT-001 through PFX-UAT-012 and PFX-UAT-014 if executable.
- Mark PFX-UAT-013 as READY_TO_RUN after Sprint 36 PR #616 smoke evidence.
- Do not create a release tag from this discovery session.
- Do not mark Production E2E as business accepted.
- Do not use Go/Conditional Go/No-Go as the final decision.
- Use "Discovery only — no business UAT decision" in the report.
```

---

## 5. Scenario execution plan for Mode B

Run these scenarios:

```text
PFX-UAT-001 — Production plan + material demand
PFX-UAT-002 — Shortage -> Purchase Request traceability
PFX-UAT-003 — PO -> Receiving -> Inbound QC
PFX-UAT-004 — Warehouse issue NVL/bao bi to factory
PFX-UAT-005 — Factory dispatch pack and confirmation
PFX-UAT-006 — Factory material handover evidence
PFX-UAT-007 — Sample approval / rework
PFX-UAT-008 — Mass production progress
PFX-UAT-009 — Finished goods receipt to QC hold
PFX-UAT-010 — Finished goods QC closeout full/partial/fail
PFX-UAT-011 — Factory claim within 3-7 days
PFX-UAT-012 — Final payment readiness
PFX-UAT-014 — Negative controls
```

Now ready to run in controlled discovery mode:

```text
PFX-UAT-013 — AP handoff + payment voucher/cash-out evidence
```

Mark `PFX-UAT-013` as:

```text
Status: NOT_RUN / READY_TO_RUN
Issue ID: PFX-BLOCKER-001
Decision note: S36_RUNTIME_READY_PR616
```

---

## 6. Alignment with docs/138

`PFX-UAT-013` in `docs/138_ERP_UAT_Pilot_Pack_Production_External_Factory_E2E_MyPham_v1.md` must carry this dependency note.

```text
Dependency:
This scenario depended on Sprint 36 runtime implementation for payment voucher / cash-out evidence.
Sprint 36 PR #616 is merged and smoke-tested, so PFX-UAT-013 is ready to run in controlled discovery mode.
Do not claim Business UAT Go until the scenario is actually executed or explicitly waived by the Business Owner with a documented reason.
No release tag may be created from discovery preparation alone.
```

The UAT mode section in file 138 must also carry this discovery rule:

```text
Mode B / Discovery rule:
Production E2E Discovery may proceed without PFX-UAT-013 if Sprint 36 runtime is not ready.
In older snapshots where Sprint 36 runtime was not ready, PFX-UAT-013 was marked BLOCKED_BY_S36.
The session result must be recorded as Discovery only — no business UAT decision.
```

---

## 7. Sprint 36 dependency statement

Sprint 36 must be treated as a dependency for full Production E2E Business UAT because it owns the runtime behavior for:

```text
- payment voucher / cash-out evidence
- supplier payable payment evidence
- payment amount, method, business date, reference, memo
- source evidence / attachment path
- allocation or reconciliation with supplier payable
```

Guardrail:

```text
Supplier payable payment status must be derived from or reconciled with posted cash_out allocations.
Do not maintain two independent payment truths.
```

---

## 8. Discovery session schedule

Recommended schedule:

### Session 1 — Planning + Material Demand

```text
PFX-UAT-001
PFX-UAT-002
```

### Session 2 — Purchase / Receiving / QC + Issue to Factory

```text
PFX-UAT-003
PFX-UAT-004
```

### Session 3 — Factory Execution

```text
PFX-UAT-005
PFX-UAT-006
PFX-UAT-007
PFX-UAT-008
```

### Session 4 — Finished Goods + Claim + Controls

```text
PFX-UAT-009
PFX-UAT-010
PFX-UAT-011
PFX-UAT-012
PFX-UAT-014
```

`PFX-UAT-013` was blocked until Sprint 36 runtime was ready. After PR #616 and dev/browser smoke, it is ready to run in controlled discovery.

---

## 9. Evidence rules

Evidence must be sanitized before commit or sharing.

Evidence folders:

```text
docs/uat/production-e2e/evidence/screenshots/
docs/uat/production-e2e/evidence/logs/
docs/uat/production-e2e/evidence/exports/
docs/uat/production-e2e/evidence/session-notes/
```

Do not store:

```text
- real customer personal data
- real bank account numbers
- real supplier confidential prices unless sanitized
- raw access tokens
- passwords
- session cookies
- unredacted production logs
```

---

## 10. Status policy

Use these statuses in `scenario_results.csv`:

| Status | Meaning |
|---|---|
| PASS | Scenario passed as written. |
| PASS_WITH_OBSERVATION | Scenario passed but produced non-blocking observations. |
| BLOCKED | Scenario could not execute because of dependency/environment/missing data. |
| FAIL | Scenario executed and failed expected behavior. |
| WAIVED | Scenario explicitly waived by owner with reason. |
| NOT_RUN | Scenario has not been executed. |

For `PFX-UAT-013`, use:

```text
Status: NOT_RUN / READY_TO_RUN
Issue IDs: PFX-BLOCKER-001 closed history
Decision note: S36_RUNTIME_READY_PR616
```

---

## 11. Go/No-Go rule for Mode B

Do not select Go, Conditional Go, or No-Go for Mode B.

Use:

```text
Discovery only — no business UAT decision
```

This protects the project from false confidence.

---

## 12. README / Master Index update wording

Optional wording for README or Master Index:

```text
Production External-Factory E2E Discovery can proceed in controlled walkthrough mode.
PFX-UAT-013 AP handoff/payment voucher scenario is ready to run after Sprint 36 PR #616 merge and smoke evidence.
No Business UAT Go or release tag is claimed from the discovery session.
Sprint 22 Warehouse/Sales/QC Go/No-Go remains a separate pending business-validation gate.
```

---

## 13. Documentation integration scope

```text
1. File 138 carries PFX-UAT-013 dependency on Sprint 36.
2. File 138 carries Mode B discovery rule: no Go/No-Go, no release tag, no business UAT claim.
3. Production E2E templates are updated:
   - scenario_results.csv
   - issue_triage_board.csv
   - go_no_go_report.md
4. README/docs/80 record discovery status wording.
5. No runtime code change.
6. No API/schema/migration change.
```

---

## 14. Acceptance criteria

This docs update is complete when:

```text
- PFX-UAT-013 clearly states Sprint 36 dependency and current ready-to-run status.
- Discovery mode explicitly says no business UAT decision.
- Scenario template marks PFX-UAT-013 as NOT_RUN / S36_RUNTIME_READY_PR616.
- Issue triage template contains resolved PFX-BLOCKER-001 history.
- Go/No-Go template includes Discovery only option.
- README or Master Index does not imply Production E2E UAT has passed.
- No v0.xx release tag is created from docs-only discovery prep.
```

---

## 15. Next action after this docs update

After merging the docs update:

```text
1. Run Production E2E Discovery sessions 1-4.
2. Record results in scenario_results.csv.
3. Record observations and issues.
4. Run PFX-UAT-013 in controlled discovery after Sprint 36 runtime smoke evidence.
5. Decide whether discovery issues require fixes before business UAT or Go/No-Go.
```
