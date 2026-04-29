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
12 up migrations
12 down migrations
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
Temporary Docker Compose project: erp-migration-verify
Temporary PostgreSQL container: erp-migration-verify-postgres-1
Temporary Docker volume: erp-migration-verify_postgres-data
```

The temporary container, network, and volume were removed after verification.

## Commands Verified

Apply all migrations:

```text
docker compose -p erp-migration-verify -f infra/compose/docker-compose.local.yml --profile tools run --rm migrate
```

Roll back all migrations:

```text
docker compose -p erp-migration-verify -f infra/compose/docker-compose.local.yml --profile tools run --rm migrate -path /migrations -database postgres://erp:***@postgres:5432/erp?sslmode=disable down 12
```

Cleanup:

```text
docker compose -p erp-migration-verify -f infra/compose/docker-compose.local.yml down -v --remove-orphans
```

## Results

Apply result:

```text
1/u init
2/u create_phase1_base_tables
3/u create_uom_foundation
4/u stock_ledger_decimal_base_uom
5/u sales_order_foundation
6/u harden_stock_reservations
7/u create_pick_tasks
8/u create_pack_tasks
9/u harden_carrier_manifest_db
10/u add_manifest_handover_zone_bins
11/u harden_return_receiving_db_model
12/u purchase_order_full_flow
```

Rollback result:

```text
12/d purchase_order_full_flow
11/d harden_return_receiving_db_model
10/d add_manifest_handover_zone_bins
9/d harden_carrier_manifest_db
8/d create_pack_tasks
7/d create_pick_tasks
6/d harden_stock_reservations
5/d sales_order_foundation
4/d stock_ledger_decimal_base_uom
3/d create_uom_foundation
2/d create_phase1_base_tables
1/d init
```

No application migration failed during apply or rollback.

## Remaining Release Gate

S4-00-02 is verified. Production release tagging remains on hold because S4-00-01 is blocked by the GitHub Actions account billing/spending-limit issue.
