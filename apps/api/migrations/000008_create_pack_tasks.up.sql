BEGIN;

CREATE TABLE shipping.pack_tasks (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  org_id uuid NOT NULL REFERENCES core.organizations(id),
  pack_task_no text NOT NULL,
  source_doc_type text NOT NULL DEFAULT 'sales_order',
  sales_order_id uuid NOT NULL REFERENCES sales.sales_orders(id),
  pick_task_id uuid NOT NULL REFERENCES shipping.pick_tasks(id),
  warehouse_id uuid NOT NULL REFERENCES mdm.warehouses(id),
  status text NOT NULL DEFAULT 'created',
  assigned_to uuid REFERENCES core.users(id),
  assigned_at timestamptz,
  started_at timestamptz,
  started_by uuid REFERENCES core.users(id),
  packed_at timestamptz,
  packed_by uuid REFERENCES core.users(id),
  exception_code text,
  note text,
  created_at timestamptz NOT NULL DEFAULT now(),
  created_by uuid REFERENCES core.users(id),
  updated_at timestamptz NOT NULL DEFAULT now(),
  updated_by uuid REFERENCES core.users(id),
  version integer NOT NULL DEFAULT 1,
  CONSTRAINT uq_pack_tasks_org_no UNIQUE (org_id, pack_task_no),
  CONSTRAINT uq_pack_tasks_sales_order UNIQUE (org_id, sales_order_id),
  CONSTRAINT uq_pack_tasks_pick_task UNIQUE (org_id, pick_task_id),
  CONSTRAINT ck_pack_tasks_source CHECK (source_doc_type = 'sales_order'),
  CONSTRAINT ck_pack_tasks_status CHECK (
    status IN (
      'created',
      'in_progress',
      'packed',
      'pack_exception',
      'cancelled'
    )
  ),
  CONSTRAINT ck_pack_tasks_lifecycle CHECK (
    (
      status = 'created'
      AND started_at IS NULL
      AND packed_at IS NULL
    )
    OR (
      status = 'in_progress'
      AND started_at IS NOT NULL
      AND started_by IS NOT NULL
      AND packed_at IS NULL
    )
    OR (
      status = 'packed'
      AND started_at IS NOT NULL
      AND started_by IS NOT NULL
      AND packed_at IS NOT NULL
      AND packed_by IS NOT NULL
    )
    OR status IN ('pack_exception', 'cancelled')
  )
);

CREATE INDEX ix_pack_tasks_status_assignee
  ON shipping.pack_tasks(org_id, status, assigned_to, created_at DESC);

CREATE INDEX ix_pack_tasks_sales_order
  ON shipping.pack_tasks(sales_order_id);

CREATE INDEX ix_pack_tasks_pick_task
  ON shipping.pack_tasks(pick_task_id);

CREATE TABLE shipping.pack_task_lines (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  org_id uuid NOT NULL REFERENCES core.organizations(id),
  pack_task_id uuid NOT NULL REFERENCES shipping.pack_tasks(id) ON DELETE RESTRICT,
  line_no integer NOT NULL,
  pick_task_line_id uuid NOT NULL REFERENCES shipping.pick_task_lines(id),
  sales_order_line_id uuid NOT NULL REFERENCES sales.sales_order_lines(id),
  item_id uuid NOT NULL REFERENCES mdm.items(id),
  sku_code text NOT NULL,
  batch_id uuid NOT NULL REFERENCES inventory.batches(id),
  warehouse_id uuid NOT NULL REFERENCES mdm.warehouses(id),
  base_uom_code varchar(20) NOT NULL REFERENCES mdm.uoms(uom_code),
  qty_to_pack numeric(18, 6) NOT NULL,
  qty_packed numeric(18, 6) NOT NULL DEFAULT 0,
  status text NOT NULL DEFAULT 'pending',
  packed_at timestamptz,
  packed_by uuid REFERENCES core.users(id),
  exception_code text,
  note text,
  created_at timestamptz NOT NULL DEFAULT now(),
  created_by uuid REFERENCES core.users(id),
  updated_at timestamptz NOT NULL DEFAULT now(),
  updated_by uuid REFERENCES core.users(id),
  CONSTRAINT uq_pack_task_lines_line UNIQUE (pack_task_id, line_no),
  CONSTRAINT uq_pack_task_lines_pick_line UNIQUE (pack_task_id, pick_task_line_id),
  CONSTRAINT ck_pack_task_lines_qty CHECK (
    qty_to_pack > 0
    AND qty_packed >= 0
    AND qty_packed <= qty_to_pack
  ),
  CONSTRAINT ck_pack_task_lines_status CHECK (
    status IN (
      'pending',
      'packed',
      'pack_exception',
      'cancelled'
    )
  ),
  CONSTRAINT ck_pack_task_lines_packed_metadata CHECK (
    (status = 'packed' AND qty_packed = qty_to_pack AND packed_at IS NOT NULL AND packed_by IS NOT NULL)
    OR (status <> 'packed')
  )
);

CREATE INDEX ix_pack_task_lines_task
  ON shipping.pack_task_lines(pack_task_id, line_no);

CREATE INDEX ix_pack_task_lines_sales_order_line
  ON shipping.pack_task_lines(sales_order_line_id);

CREATE INDEX ix_pack_task_lines_pick_line
  ON shipping.pack_task_lines(pick_task_line_id);

COMMIT;
