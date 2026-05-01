# S9-05-01 Dev Deploy Evidence Script

Project: Web ERP for cosmetics operations
Sprint: Sprint 9 - System hardening / production readiness core
Task: S9-05-01 Dev deploy evidence script
Date: 2026-05-01
Status: Evidence script defined

---

## 1. Purpose

This task adds a repeatable way to capture shared dev deployment evidence for release changelogs.

The evidence should be compact enough to paste into a changelog while still showing what was actually verified.

---

## 2. Command

Run after a dev deploy:

```text
make dev-deploy-evidence
```

Equivalent direct script:

```text
./infra/scripts/dev-deploy-evidence.sh dev
```

---

## 3. Evidence Included

The script prints:

```text
- Generated timestamp.
- Current git branch and commit.
- Full commit SHA.
- Dev base URL.
- HTTP status for /healthz, /api/v1/health, and web root. The web root may be `307` when it redirects an authenticated session toward the dashboard.
- Docker Compose service status.
- Full dev smoke output, including persisted stock movement evidence.
- Final smoke result.
```

---

## 4. Release Gate Use

Use this output in Sprint changelogs after the deployment command succeeds.

This script records dev deployment evidence only. It does not replace cloud CI evidence or production tag approval.

---

## 5. Verification Notes

Expected dev-server verification:

```text
sh -n infra/scripts/dev-deploy-evidence.sh
./infra/scripts/dev-deploy-evidence.sh dev
```

If a host lacks `make`, the direct script is the source of truth.
