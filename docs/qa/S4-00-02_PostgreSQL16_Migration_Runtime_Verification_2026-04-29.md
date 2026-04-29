# S4-00-02 PostgreSQL 16 Migration Runtime Verification

Task: S4-00-02
Date: 2026-04-29
Verifier: Codex

## Scope

Verify that all current API migrations can apply and roll back on an isolated PostgreSQL 16 runtime.

Migration source:

```text
apps/api/migrations
```

Migration count:

```text
11 up migrations
11 down migrations
```

## Environment

Local Docker was unavailable on the workstation, so the verification used an isolated Docker runtime on the dev server.

Runtime:

```text
PostgreSQL 16.13 on x86_64-pc-linux-musl
migrate/migrate:v4.17.1
```

Isolation:

```text
Temporary migration copy: /tmp/erp-s4-migration-verify/migrations
Temporary Docker network: erp_s4_migration_verify_net
Temporary PostgreSQL container: erp_s4_migration_verify_pg
```

The temporary container and network were removed after verification.

## Commands Verified

Apply all migrations:

```text
migrate -path /migrations -database postgres://erp:***@erp_s4_migration_verify_pg:5432/erp?sslmode=disable -verbose up
```

Check version after apply:

```text
migrate -path /migrations -database postgres://erp:***@erp_s4_migration_verify_pg:5432/erp?sslmode=disable version
```

Roll back all migrations:

```text
migrate -path /migrations -database postgres://erp:***@erp_s4_migration_verify_pg:5432/erp?sslmode=disable -verbose down -all
```

## Results

Apply result:

```text
000001_init.up.sql through 000011_harden_return_receiving_db_model.up.sql applied successfully.
Version after up: 11
Relation count after up: 55
```

Rollback result:

```text
000011_harden_return_receiving_db_model.down.sql through 000001_init.down.sql applied successfully.
Version after down all: no migration
Relation count after down all: 1
```

The remaining relation after rollback is migration metadata. No application migration failed during apply or rollback.

## Remaining Release Gate

S4-00-02 is verified. The Sprint 3 production tag still remains on hold because S4-00-01 is blocked by the GitHub Actions account billing/spending-limit issue.
