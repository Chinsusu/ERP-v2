BEGIN;

CREATE TABLE shipping.pick_tasks (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  org_id uuid NOT NULL REFERENCES core.organizations(id),
  pick_task_no text NOT NULL,
  source_doc_type text NOT NULL DEFAULT 'sales_order',
  sales_order_id uuid NOT NULL REFERENCES sales.sales_orders(id),
  warehouse_id uuid NOT NULL REFERENCES mdm.warehouses(id),
  status text NOT NULL DEFAULT 'created',
  assigned_to uuid REFERENCES core.users(id),
  assigned_at timestamptz,
  started_at timestamptz,
  started_by uuid REFERENCES core.users(id),
  completed_at timestamptz,
  completed_by uuid REFERENCES core.users(id),
  exception_code text,
  note text,
  created_at timestamptz NOT NULL DEFAULT now(),
  created_by uuid REFERENCES core.users(id),
  updated_at timestamptz NOT NULL DEFAULT now(),
  updated_by uuid REFERENCES core.users(id),
  version integer NOT NULL DEFAULT 1,
  CONSTRAINT uq_pick_tasks_org_no UNIQUE (org_id, pick_task_no),
  CONSTRAINT uq_pick_tasks_sales_order UNIQUE (org_id, sales_order_id),
  CONSTRAINT ck_pick_tasks_source CHECK (source_doc_type = 'sales_order'),
  CONSTRAINT ck_pick_tasks_status CHECK (
    status IN (
      'created',
      'assigned',
      'in_progress',
      'completed',
      'missing_stock',
      'wrong_sku',
      'wrong_batch',
      'wrong_location',
      'cancelled'
    )
  ),
  CONSTRAINT ck_pick_tasks_assignment CHECK (
    (status = 'created' AND assigned_to IS NULL AND assigned_at IS NULL AND started_at IS NULL AND completed_at IS NULL)
    OR (status = 'assigned' AND assigned_to IS NOT NULL AND assigned_at IS NOT NULL AND started_at IS NULL AND completed_at IS NULL)
    OR (status = 'in_progress' AND assigned_to IS NOT NULL AND assigned_at IS NOT NULL AND started_at IS NOT NULL AND started_by IS NOT NULL AND completed_at IS NULL)
    OR (status = 'completed' AND assigned_to IS NOT NULL AND assigned_at IS NOT NULL AND started_at IS NOT NULL AND started_by IS NOT NULL AND completed_at IS NOT NULL AND completed_by IS NOT NULL)
    OR status IN ('missing_stock', 'wrong_sku', 'wrong_batch', 'wrong_location', 'cancelled')
  )
);

CREATE INDEX ix_pick_tasks_status_assignee
  ON shipping.pick_tasks(org_id, status, assigned_to, created_at DESC);

CREATE INDEX ix_pick_tasks_sales_order
  ON shipping.pick_tasks(sales_order_id);

CREATE TABLE shipping.pick_task_lines (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  org_id uuid NOT NULL REFERENCES core.organizations(id),
  pick_task_id uuid NOT NULL REFERENCES shipping.pick_tasks(id) ON DELETE RESTRICT,
  line_no integer NOT NULL,
  sales_order_line_id uuid NOT NULL REFERENCES sales.sales_order_lines(id),
  stock_reservation_id uuid NOT NULL REFERENCES inventory.stock_reservations(id),
  item_id uuid NOT NULL REFERENCES mdm.items(id),
  sku_code text NOT NULL,
  batch_id uuid NOT NULL REFERENCES inventory.batches(id),
  warehouse_id uuid NOT NULL REFERENCES mdm.warehouses(id),
  bin_id uuid NOT NULL REFERENCES mdm.warehouse_bins(id),
  base_uom_code varchar(20) NOT NULL REFERENCES mdm.uoms(uom_code),
  qty_to_pick numeric(18, 6) NOT NULL,
  qty_picked numeric(18, 6) NOT NULL DEFAULT 0,
  status text NOT NULL DEFAULT 'pending',
  picked_at timestamptz,
  picked_by uuid REFERENCES core.users(id),
  exception_code text,
  note text,
  created_at timestamptz NOT NULL DEFAULT now(),
  created_by uuid REFERENCES core.users(id),
  updated_at timestamptz NOT NULL DEFAULT now(),
  updated_by uuid REFERENCES core.users(id),
  CONSTRAINT uq_pick_task_lines_line UNIQUE (pick_task_id, line_no),
  CONSTRAINT uq_pick_task_lines_reservation UNIQUE (pick_task_id, stock_reservation_id),
  CONSTRAINT ck_pick_task_lines_qty CHECK (
    qty_to_pick > 0
    AND qty_picked >= 0
    AND qty_picked <= qty_to_pick
  ),
  CONSTRAINT ck_pick_task_lines_status CHECK (
    status IN (
      'pending',
      'picked',
      'missing_stock',
      'wrong_sku',
      'wrong_batch',
      'wrong_location',
      'cancelled'
    )
  ),
  CONSTRAINT ck_pick_task_lines_picked_metadata CHECK (
    (status = 'picked' AND qty_picked = qty_to_pick AND picked_at IS NOT NULL AND picked_by IS NOT NULL)
    OR (status <> 'picked')
  )
);

CREATE INDEX ix_pick_task_lines_task
  ON shipping.pick_task_lines(pick_task_id, line_no);

CREATE INDEX ix_pick_task_lines_sales_order_line
  ON shipping.pick_task_lines(sales_order_line_id);

CREATE INDEX ix_pick_task_lines_reservation
  ON shipping.pick_task_lines(stock_reservation_id);

COMMIT;
