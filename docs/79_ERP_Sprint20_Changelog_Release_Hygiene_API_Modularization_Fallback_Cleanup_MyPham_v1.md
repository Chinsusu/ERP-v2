# 79_ERP_Sprint20_Changelog_Release_Hygiene_API_Modularization_Fallback_Cleanup_MyPham_v1

Project: Web ERP for cosmetics operations
Phase: Phase 1
Sprint: Sprint 20 - Release hygiene, API modularization, and fallback cleanup
Document role: Sprint changelog and release evidence
Version: v1.0
Date: 2026-05-03
Status: Completed on main after PR #537, with post-S20 main hygiene through PR #540

---

## 1. Sprint 20 Scope

Sprint 20 closed hardening gaps after the Sprint 19 Vietnamese UI localization release.

```text
Sprint 19: Vietnamese-first ERP UI localization release
Sprint 20: release evidence, migration gate, CI runtime compatibility, API route maintainability, and production fallback safety
```

In scope:

```text
README and release evidence follow-through
Sprint 19 release tag evidence
Migration apply -> rollback -> reapply gate
GitHub Actions Node.js 24 compatibility
API route registration modularization
Production-mode prototype fallback blocking
Production runtime mode checklist
```

Out of scope:

```text
New business workflows
Route localization
Backend enum renaming
Database value renaming
Permission key renaming
Frontend auth production integration
```

---

## 2. Merged PRs

| Task | PR | Result |
| --- | --- | --- |
| S20-01/S20-02 release docs follow-through | #532 | Recorded Sprint 19 release tag and current release status after CI rerun |
| S20-03 push Sprint 19 release tag | #532 evidence plus Git tag | `v0.19.0-vietnamese-ui-localization` points to release commit `df9b9567` |
| S20-04 migration reapply gate | #533 | `required-migration` now validates PostgreSQL 16 apply -> rollback -> reapply |
| S20-05 Node.js 24 GitHub Actions compatibility | #534 | Required workflows were updated for Node 24 runtime compatibility |
| S20-06 API route modularization | #535 | API route registration was split by module while preserving public paths |
| S20-07 production fallback gating | #536 | Web prototype fallback guard blocks backend-backed fallback masking in production mode |
| S20-08 production runtime checklist | #537 | Added production runtime checklist covering env, persistence, fallback, smoke, and release gates |

All Sprint 20 PRs used manual review and merge. GitHub auto-review and auto-merge were not used.

---

## 3. Verification Evidence

Cloud checks recorded during Sprint 20:

```text
#532 required-ci run 25260025598: success
#533 required-ci run 25260113666: success
#533 migration-ci run 25260113663: success
#534 required-ci run 25260193486: success
#534 api-ci run 25260193480: success
#534 openapi-ci run 25260193481: success
#534 web-ci run 25260193491: success
#534 migration-ci run 25260193493: success
#535 required-ci run 25260607147: success
#535 api-ci run 25260607161: success
#535 openapi-ci run 25260607157: success
#535 web-ci run 25260607155: success
#536 required-ci run 25260846812: success
#536 web-ci run 25260846805: success
#537 required-ci run 25261048405: success
```

Post-S20 main evidence before this docs traceability cleanup:

```text
main commit d455aa16
PR #538 language switch merged: required-ci run 25267277657 success, web-ci run 25267277649 success
PR #539 dev seed runtime constraints merged: required-ci run 25267741375 success, api-ci run 25267741372 success
PR #540 dev deploy proxy recreate merged: required-ci run 25267823659 success
```

Migration gate state:

```text
Release tag v0.19.0 gate: PostgreSQL 16 apply + rollback passed.
Current main after Sprint 20: PostgreSQL 16 apply -> rollback -> reapply passed.
```

---

## 4. Runtime And Production Readiness Notes

Sprint 20 makes production-like runtime stricter, but it does not make every frontend surface production-ready.

Required caveat:

```text
Web auth UI is still mock/staging-only until wired to the backend auth/session API.
Backend auth/session persistence exists from Sprint 18, but the frontend login surface must not be called production-ready until the integration is explicit or accepted as a controlled staging-only mock.
```

Fallback rule:

```text
Production-like runtime must not silently render prototype fallback data after backend API failures for pages that claim backend coverage.
```

---

## 5. Release Traceability

Latest release tag:

```text
v0.19.0-vietnamese-ui-localization -> df9b9567
```

Current main after Sprint 20 and post-hardening dev fixes:

```text
d455aa16 Merge pull request #540 from Chinsusu/codex/fix-dev-deploy-recreate-proxy
```

Release numbering note:

```text
Sprint 16-17 runtime persistence work was merged to main and consolidated under the v0.18.0 auth/session runtime store release evidence.
Sprint 19 was released as v0.19.0-vietnamese-ui-localization.
Sprint 20 is hardening after v0.19.0 and does not create a separate product feature tag in this document.
```

---

## 6. Source Documents

| Area | Document |
| --- | --- |
| Sprint 19 localization task board | `75_ERP_Coding_Task_Board_Sprint19_Vietnamese_UI_Localization_MyPham_v1.md` |
| Sprint 20 hardening task board | `76_ERP_Coding_Task_Board_Sprint20_Release_Hygiene_API_Modularization_Fallback_Cleanup_MyPham_v1.md` |
| Sprint 19 release changelog | `77_ERP_Sprint19_Changelog_Vietnamese_UI_Localization_MyPham_v1.md` |
| Production runtime checklist | `78_ERP_Production_Runtime_Mode_Checklist_Sprint20_MyPham_v1.md` |
| Current master document index | `80_ERP_Master_Document_Index_Current_Status_MyPham_v2.md` |
| Vietnamese operational glossary | `81_ERP_Vietnamese_UI_Glossary_Operational_Copy_MyPham_v1.md` |
