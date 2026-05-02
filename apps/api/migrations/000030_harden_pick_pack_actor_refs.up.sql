BEGIN;

ALTER TABLE shipping.pick_tasks
  DROP CONSTRAINT IF EXISTS ck_pick_tasks_assignment,
  ADD CONSTRAINT ck_pick_tasks_assignment CHECK (
    (
      status = 'created'
      AND assigned_to IS NULL
      AND nullif(btrim(coalesce(assigned_to_ref, '')), '') IS NULL
      AND assigned_at IS NULL
      AND started_at IS NULL
      AND completed_at IS NULL
    )
    OR (
      status = 'assigned'
      AND (assigned_to IS NOT NULL OR nullif(btrim(coalesce(assigned_to_ref, '')), '') IS NOT NULL)
      AND assigned_at IS NOT NULL
      AND started_at IS NULL
      AND completed_at IS NULL
    )
    OR (
      status = 'in_progress'
      AND (assigned_to IS NOT NULL OR nullif(btrim(coalesce(assigned_to_ref, '')), '') IS NOT NULL)
      AND assigned_at IS NOT NULL
      AND started_at IS NOT NULL
      AND (started_by IS NOT NULL OR nullif(btrim(coalesce(started_by_ref, '')), '') IS NOT NULL)
      AND completed_at IS NULL
    )
    OR (
      status = 'completed'
      AND (assigned_to IS NOT NULL OR nullif(btrim(coalesce(assigned_to_ref, '')), '') IS NOT NULL)
      AND assigned_at IS NOT NULL
      AND started_at IS NOT NULL
      AND (started_by IS NOT NULL OR nullif(btrim(coalesce(started_by_ref, '')), '') IS NOT NULL)
      AND completed_at IS NOT NULL
      AND (completed_by IS NOT NULL OR nullif(btrim(coalesce(completed_by_ref, '')), '') IS NOT NULL)
    )
    OR status IN ('missing_stock', 'wrong_sku', 'wrong_batch', 'wrong_location', 'cancelled')
  );

ALTER TABLE shipping.pick_task_lines
  DROP CONSTRAINT IF EXISTS ck_pick_task_lines_picked_metadata,
  ADD CONSTRAINT ck_pick_task_lines_picked_metadata CHECK (
    (
      status = 'picked'
      AND qty_picked = qty_to_pick
      AND picked_at IS NOT NULL
      AND (picked_by IS NOT NULL OR nullif(btrim(coalesce(picked_by_ref, '')), '') IS NOT NULL)
    )
    OR (status <> 'picked')
  );

ALTER TABLE shipping.pack_tasks
  DROP CONSTRAINT IF EXISTS ck_pack_tasks_lifecycle,
  ADD CONSTRAINT ck_pack_tasks_lifecycle CHECK (
    (
      status = 'created'
      AND started_at IS NULL
      AND packed_at IS NULL
    )
    OR (
      status = 'in_progress'
      AND started_at IS NOT NULL
      AND (started_by IS NOT NULL OR nullif(btrim(coalesce(started_by_ref, '')), '') IS NOT NULL)
      AND packed_at IS NULL
    )
    OR (
      status = 'packed'
      AND started_at IS NOT NULL
      AND (started_by IS NOT NULL OR nullif(btrim(coalesce(started_by_ref, '')), '') IS NOT NULL)
      AND packed_at IS NOT NULL
      AND (packed_by IS NOT NULL OR nullif(btrim(coalesce(packed_by_ref, '')), '') IS NOT NULL)
    )
    OR status IN ('pack_exception', 'cancelled')
  );

ALTER TABLE shipping.pack_task_lines
  DROP CONSTRAINT IF EXISTS ck_pack_task_lines_packed_metadata,
  ADD CONSTRAINT ck_pack_task_lines_packed_metadata CHECK (
    (
      status = 'packed'
      AND qty_packed = qty_to_pack
      AND packed_at IS NOT NULL
      AND (packed_by IS NOT NULL OR nullif(btrim(coalesce(packed_by_ref, '')), '') IS NOT NULL)
    )
    OR (status <> 'packed')
  );

COMMIT;
