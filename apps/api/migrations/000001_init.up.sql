CREATE SCHEMA IF NOT EXISTS audit;
CREATE SCHEMA IF NOT EXISTS inventory;

CREATE TABLE IF NOT EXISTS audit.audit_logs (
  id bigserial PRIMARY KEY,
  actor_id text NOT NULL,
  action text NOT NULL,
  resource text NOT NULL,
  metadata jsonb NOT NULL DEFAULT '{}'::jsonb,
  created_at timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS inventory.stock_ledger (
  id bigserial PRIMARY KEY,
  movement_id text NOT NULL UNIQUE,
  sku text NOT NULL,
  warehouse_id text NOT NULL,
  movement_type text NOT NULL,
  quantity numeric(18, 4) NOT NULL CHECK (quantity > 0),
  reason text NOT NULL,
  created_at timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_stock_ledger_sku_warehouse_created
  ON inventory.stock_ledger (sku, warehouse_id, created_at);
