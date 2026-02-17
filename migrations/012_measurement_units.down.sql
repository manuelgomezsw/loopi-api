-- ============================================
-- LOOPI - Rollback measurement units from items
-- ============================================

ALTER TABLE `items` DROP FOREIGN KEY `fk_items_measurement_unit`;
ALTER TABLE `items` DROP INDEX `idx_items_measurement_unit`;
ALTER TABLE `items` DROP COLUMN `measurement_unit_id`;

DROP TABLE `measurement_units`;
