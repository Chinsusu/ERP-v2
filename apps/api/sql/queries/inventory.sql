-- name: ListStockMovements :many
SELECT
  movement_id,
  sku,
  warehouse_id,
  movement_type,
  quantity,
  reason,
  created_at
FROM inventory.stock_ledger
ORDER BY created_at DESC
LIMIT $1 OFFSET $2;
