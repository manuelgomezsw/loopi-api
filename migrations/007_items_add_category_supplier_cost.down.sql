-- ============================================
-- LOOPI - Rollback category, supplier and cost from items
-- ============================================

ALTER TABLE `items` DROP FOREIGN KEY `fk_items_supplier`;
ALTER TABLE `items` DROP FOREIGN KEY `fk_items_category`;
ALTER TABLE `items` DROP INDEX `idx_items_supplier`;
ALTER TABLE `items` DROP INDEX `idx_items_category`;
ALTER TABLE `items` DROP COLUMN `cost`;
ALTER TABLE `items` DROP COLUMN `supplier_id`;
ALTER TABLE `items` DROP COLUMN `category_id`;
