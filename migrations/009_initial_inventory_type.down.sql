-- ============================================
-- LOOPI - Revert 'initial' inventory type
-- ============================================

-- Remove any initial inventories first
DELETE FROM `inventories` WHERE `inventory_type` = 'initial';

ALTER TABLE `inventories`
MODIFY COLUMN `inventory_type` ENUM('daily', 'weekly', 'monthly') NOT NULL;
