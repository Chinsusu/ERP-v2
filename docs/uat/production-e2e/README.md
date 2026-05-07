# Production External-Factory E2E UAT Evidence Folder

This folder supports:

```text
docs/138_ERP_UAT_Pilot_Pack_Production_External_Factory_E2E_MyPham_v1.md
docs/139_ERP_Production_E2E_Discovery_Mode_S36_Blocker_MyPham_v1.md
```

The current templates are prepared for Production E2E Discovery / Controlled Walkthrough mode. They do not claim Business UAT pass.

`PFX-UAT-013` is marked `BLOCKED_BY_S36` until Sprint 36 payment voucher / cash-out evidence runtime is merged and smoke-tested.

Commit templates and sanitized evidence only. Do not commit raw screenshots, exports, logs, passwords, tokens, customer PII, private addresses, phone numbers, bank details, supplier pricing secrets, or commercial secrets.

## Structure

```text
templates/      CSV and report templates used during Production E2E UAT
evidence/       Sanitized screenshots, logs, exports, files, and session notes
```

## Required Templates

```text
templates/uat_users_roles.csv
templates/seed_data_plan.csv
templates/scenario_results.csv
templates/observation_log.csv
templates/issue_triage_board.csv
templates/session_schedule.csv
templates/go_no_go_report.md
```

## Evidence Handling

Raw UAT evidence must be reviewed and redacted before it is committed.

Use these subfolders for sanitized evidence:

```text
evidence/screenshots
evidence/logs
evidence/exports
evidence/session-notes
evidence/sanitized-files
```
