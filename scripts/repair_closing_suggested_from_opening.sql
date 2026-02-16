-- =============================================================================
-- Repair: suggested_value for CLOSING inventories that were built from initial
-- instead of from OPENING (same day) when noon does not exist.
--
-- Use when: closing inventories show wrong "esperado" (e.g. 0 or values from
-- initial) because FindPreviousInventory used initial instead of same-day opening.
--
-- Logic: For each daily CLOSING inventory that has a same-day OPENING completed,
-- set each detail's suggested_value = opening.real_value (per item).
-- Sugerido del siguiente = solo conteo anterior; mermas no restan del esperado del siguiente.
--
-- Run after backup. Prefer running in a transaction and checking counts/values
-- before COMMIT.
-- =============================================================================

-- 1) Preview: closing inventories that have same-day opening (will be repaired)
SELECT
  inv.id AS closing_id,
  inv.inventory_date,
  open_inv.id AS opening_id,
  (SELECT COUNT(*) FROM inventory_details WHERE inventory_id = inv.id) AS closing_details,
  (SELECT COUNT(*) FROM inventory_details WHERE inventory_id = open_inv.id) AS opening_details
FROM inventories inv
INNER JOIN inventories open_inv
  ON open_inv.inventory_date = inv.inventory_date
  AND open_inv.inventory_type = 'daily'
  AND open_inv.schedule = 'opening'
  AND open_inv.status = 'completed'
WHERE inv.inventory_type = 'daily'
  AND inv.schedule = 'closing'
ORDER BY inv.inventory_date DESC, inv.id;

-- 2) Preview: suggested_value changes (before/after) for one closing — adjust IDs
-- Replace 20 with the closing inventory_id you want to inspect
/*
SELECT
  c.id AS detail_id,
  c.item_id,
  c.suggested_value AS current_suggested,
  COALESCE(o.real_value, 0) AS new_suggested
FROM inventory_details c
INNER JOIN inventories inv ON c.inventory_id = inv.id AND inv.id = 20
INNER JOIN inventories open_inv
  ON open_inv.inventory_date = inv.inventory_date
  AND open_inv.schedule = 'opening'
  AND open_inv.inventory_type = 'daily'
  AND open_inv.status = 'completed'
INNER JOIN inventory_details o ON o.inventory_id = open_inv.id AND o.item_id = c.item_id;
*/

-- 3) Repair: update suggested_value for all closing inventories that have same-day opening
UPDATE inventory_details c
INNER JOIN inventories inv
  ON c.inventory_id = inv.id
  AND inv.inventory_type = 'daily'
  AND inv.schedule = 'closing'
INNER JOIN inventories open_inv
  ON open_inv.inventory_date = inv.inventory_date
  AND open_inv.inventory_type = 'daily'
  AND open_inv.schedule = 'opening'
  AND open_inv.status = 'completed'
INNER JOIN inventory_details o
  ON o.inventory_id = open_inv.id
  AND o.item_id = c.item_id
SET c.suggested_value = COALESCE(o.real_value, 0);

-- 4) Optional: verify row count
-- SELECT ROW_COUNT() AS rows_updated;
