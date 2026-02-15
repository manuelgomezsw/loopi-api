-- ============================================
-- LOOPI - Add 'initial' inventory type
-- Allows creating a baseline inventory (moment zero)
-- ============================================

ALTER TABLE `inventories`
MODIFY COLUMN `inventory_type` ENUM('daily', 'weekly', 'monthly', 'initial') NOT NULL;
