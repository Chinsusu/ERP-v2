# 75_ERP_Coding_Task_Board_Sprint19_Release_Hygiene_API_Modularization_Fallback_Cleanup_MyPham_v1

Project: Web ERP for cosmetics operations
Phase: Phase 1
Sprint: Sprint 19 - Release hygiene, API modularization, and fallback cleanup
Document role: Coding task board
Version: v1.0
Date: 2026-05-02
Status: Approved for sequential implementation after Sprint 18 CI rerun green

---

## 1. Goal

Sprint 19 closes release hygiene gaps before new business features:

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
1. Do not add new product workflows in Sprint 19.
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
| S19-01 | Update README current status | README reflects Sprint 18 completion, latest release candidate, current CI gate, and hardening focus |
| S19-02 | Update Sprint 18 changelog after CI rerun | Sprint 18 changelog no longer says cloud CI blocked or production tag HOLD |
| S19-03 | Push Sprint 18 release tag | `v0.18.0-auth-session-runtime-store-persistence` exists on the Sprint 18 release commit after required-ci is green |
| S19-04 | Add migration reapply step to required-ci | `required-migration` performs apply, rollback, and reapply against PostgreSQL 16 |
| S19-05 | Fix Node.js 24 GitHub Actions warning | Required workflow opts into Node.js 24 action runtime compatibility and reruns green |
| S19-06 | Refactor API route registration by module | `cmd/api/main.go` route registration is split into module-owned registration helpers without changing paths |
| S19-07 | Gate frontend fallback services in production mode | Production builds/runtime do not silently use prototype fallback services where backend coverage exists |
| S19-08 | Add production runtime mode checklist | Docs capture required env, persistence, fallback, smoke, and release checks for production-like deployments |

---

## 4. Execution Order

```text
S19-01 + S19-02 + S19-04 + S19-05 first, because they clean release evidence and CI.
S19-03 after the release hygiene PR is merged and required-ci is green on main.
S19-06 next, because cmd/api/main.go is now a high-risk maintenance file.
S19-07 after backend coverage and runtime mode assumptions are clear.
S19-08 last, using evidence gathered from the previous tasks.
```

---

## 5. Verification

Minimum checks:

```text
required-ci on pull request
required-ci on main after merge
manual check that README and Sprint 18 changelog match current release status
release tag points to the intended main commit
```

For S19-06 and S19-07, add targeted backend/web tests as needed and run the relevant local/dev checks before PR.
