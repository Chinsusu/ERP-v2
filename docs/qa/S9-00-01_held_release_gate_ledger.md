# S9-00-01 Held Release Gate Ledger

Project: Web ERP for cosmetics operations
Sprint: Sprint 9 - System hardening / production readiness core
Task: S9-00-01 Held release gate ledger
Date: 2026-05-01
Status: Active ledger; production tags are still not created

---

## 1. Purpose

This ledger tracks the production tags carried forward into Sprint 9.

It separates three things that were previously easy to mix together:

```text
1. Dev/main merge status.
2. Latest observed GitHub Actions status.
3. Production tag creation status.
```

A sprint can be merged and dev-verified without being production-tagged.

---

## 2. Current CI Observation

Latest observed `main` checks on 2026-05-01:

```text
main commit: dc76ba4ab387b1353bcbda693e2f58a1afa0249e
workflow: required-ci
run id: 25199914071
result: success
```

The earlier GitHub Actions billing/spending-limit blocker no longer appears to be blocking the latest Sprint 9 PRs or the latest `main` required-ci run.

This does not automatically create or validate older production release tags. Each held tag still needs an explicit tag target, release gate confirmation, and push.

---

## 3. Existing Tags

Observed local tags:

```text
v0.1.0-foundation
v0.2.0-order-fulfillment-core
```

No Sprint 5, Sprint 6, Sprint 7, or Sprint 8 production tag is currently present in the local tag list.

Older Sprint 3 and Sprint 4 tag holds are documented in their changelogs and are outside this S9-00-01 carry-forward list unless the team decides to backfill all historical tags.

---

## 4. Held Production Tags

| Sprint | Intended tag | Source changelog | Current state | Release gate to clear |
| --- | --- | --- | --- | --- |
| Sprint 5 | `v0.5.0-subcontract-manufacturing-core` | `docs/48_ERP_Sprint5_Changelog_Subcontract_Manufacturing_Core_MyPham_v1.md` | Hold; tag not created | Confirm tag target, rerun/confirm green CI on `main`, then push tag |
| Sprint 6 | `v0.6.0-finance-lite-cod-ar-ap-core` | `docs/50_ERP_Sprint6_Changelog_Finance_Lite_COD_AR_AP_Core_MyPham_v1.md` | Hold; tag not created | Confirm tag target, rerun/confirm green CI on `main`, then push tag |
| Sprint 7 | `v0.7.0-reporting-inventory-operations-dashboard` | `docs/52_ERP_Sprint7_Changelog_Reporting_Inventory_Operations_Dashboard_MyPham_v1.md` | Hold; tag not created | Confirm tag target, rerun/confirm green CI on `main`, then push tag |
| Sprint 8 | `v0.8.0-reporting-hardening-dashboard-drilldowns` | `docs/54_ERP_Sprint8_Changelog_Reporting_Hardening_Dashboard_Drilldowns_MyPham_v1.md` | Hold; tag not created | Confirm tag target, rerun/confirm green CI on `main`, then push tag |

---

## 5. Tag Target Rule

Do not create a held release tag by blindly tagging the latest `main`.

Before creating each tag:

```text
1. Identify the exact merge commit that represents that sprint's release close-out.
2. Confirm later commits do not need to be included in that sprint tag.
3. Confirm GitHub Actions is green for the selected target or for the accepted release baseline.
4. Confirm migration apply/rollback requirements for that sprint.
5. Create and push the tag only after the above is recorded.
```

If the team chooses to tag all held releases at the current `main` tip instead of historical close-out commits, that must be stated explicitly in the tag evidence. Otherwise, tags should point to their sprint close-out commits.

---

## 6. Current Blockers / Open Decisions

```text
1. Production tags are still absent.
2. Exact historical tag target commits for Sprint 5, Sprint 6, Sprint 7, and Sprint 8 must be selected before tagging.
3. Some historical changelog entries mention GitHub Actions as blocked; latest main checks now show success, so those notes are historical rather than current.
4. Migration/runtime apply-rollback evidence must be checked per sprint before tagging if that sprint changed migrations.
```

---

## 7. Next Actions

Recommended order:

```text
1. Keep this ledger updated during Sprint 9.
2. Run or confirm latest full GitHub required-ci on main.
3. Select exact target commits for v0.5.0, v0.6.0, v0.7.0, and v0.8.0.
4. Confirm migration requirements for each selected release.
5. Create tags only after evidence is recorded.
```
