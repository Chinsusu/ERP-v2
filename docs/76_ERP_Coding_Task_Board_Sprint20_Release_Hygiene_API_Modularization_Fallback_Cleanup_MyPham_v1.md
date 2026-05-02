# 76_ERP_Coding_Task_Board_Sprint20_Release_Hygiene_API_Modularization_Fallback_Cleanup_MyPham_v1

Project: Web ERP for cosmetics operations
Phase: Phase 1
Sprint: Sprint 20 - Release hygiene, API modularization, and fallback cleanup
Document role: Coding task board
Version: v1.1
Date: 2026-05-02
Status: Implemented on main through S20-08

---

## 1. Goal

Sprint 20 closes release hygiene gaps after Vietnamese UI localization:

```text
release docs and tags must match the real GitHub Actions status
migration CI must prove apply -> rollback -> reapply
GitHub Actions must be tested against Node.js 24 runtime behavior
cmd/api/main.go route registration must be reduced by module
frontend fallback services must not mask missing backend coverage in production mode
```

Correctness and verifiable release evidence take priority over feature velocity.

---

## 2. Guardrails

```text
1. Do not add new product workflows in Sprint 20.
2. Do not tag release candidates before main required-ci is green.
3. Do not weaken migration checks to make CI pass.
4. Do not remove fallback behavior needed for local/no-DB development.
5. Production mode must fail loudly when required backend-backed services are unavailable.
6. Route modularization must preserve public API paths and auth behavior.
7. Keep behavioral and cosmetic changes separated by task.
8. Record exact verification run names and results in each PR.
```

---

## 3. Task List

| Task ID | Task | Output / Acceptance |
| --- | --- | --- |
| S20-01 | Update README current status | README reflects Sprint 19 completion, latest release candidate, current CI gate, and hardening focus |
| S20-02 | Update Sprint 19 changelog after CI rerun | Sprint 19 changelog records cloud CI, release tag, and known hardening gaps |
| S20-03 | Push Sprint 19 release tag | `v0.19.0-vietnamese-ui-localization` exists on the Sprint 19 release commit after required-ci is green |
| S20-04 | Keep migration reapply step green in required-ci | `required-migration` performs apply, rollback, and reapply against PostgreSQL 16 |
| S20-05 | Keep GitHub Actions Node.js 24 clean | Required workflows use Node 24-native actions and rerun green |
| S20-06 | Refactor API route registration by module | `cmd/api/main.go` route registration is split into module-owned registration helpers without changing paths |
| S20-07 | Gate frontend fallback services in production mode | Production builds/runtime do not silently use prototype fallback services where backend coverage exists |
| S20-08 | Add production runtime mode checklist | Docs capture required env, persistence, fallback, smoke, and release checks for production-like deployments |

---

## 4. Execution Order

```text
S20-01 + S20-02 first, because release docs must reflect the Sprint 19 state.
S20-03 after required-ci is green on the Sprint 19 release commit.
S20-04 + S20-05 keep the release gate clean while CI evolves.
S20-06 next, because cmd/api/main.go is now a high-risk maintenance file.
S20-07 after backend coverage and runtime mode assumptions are clear.
S20-08 last, using evidence gathered from the previous tasks.
```

---

## 5. Verification

Minimum checks:

```text
required-ci on pull request
required-ci on main after merge
manual check that README and Sprint 19 changelog match current release status
release tag points to the intended main commit
```

For S20-06 and S20-07, add targeted backend/web tests as needed and run the relevant local/dev checks before PR.

---

## 6. Completion Notes

Sprint 20 follows Sprint 19 in the release order. It is a hardening sprint, not a product-workflow sprint.

Completed outputs:

```text
S20-01 README current status updated.
S20-02 Sprint 19 changelog updated after CI rerun.
S20-03 Sprint 19 release tag v0.19.0-vietnamese-ui-localization pushed.
S20-04 Migration CI now validates apply -> rollback -> reapply.
S20-05 GitHub Actions run with Node 24 compatibility enabled.
S20-06 API route registration is split by module.
S20-07 Frontend prototype fallback is blocked in production mode.
S20-08 Production runtime mode checklist added in docs/78_ERP_Production_Runtime_Mode_Checklist_Sprint20_MyPham_v1.md.
```

Next sprint should not start until the latest `main` required-ci run is green after S20-08 merge.
