-- ============================================
-- LOOPI - Rollback inventory frequency and type
-- ============================================

-- Remove indexes
ALTER TABLE `inventories` DROP KEY `idx_inventories_type`;
ALTER TABLE `inventories` DROP KEY `uk_inventories_date_type_schedule`;

-- Restore schedule column with weekly/monthly values
ALTER TABLE `inventories` 
MODIFY COLUMN `schedule` ENUM('opening', 'noon', 'closing', 'weekly', 'monthly') NOT NULL;

-- Migrate data back
UPDATE `inventories` SET `schedule` = 'weekly' WHERE `inventory_type` = 'weekly';
UPDATE `inventories` SET `schedule` = 'monthly' WHERE `inventory_type` = 'monthly';

-- Restore original unique key
ALTER TABLE `inventories`
ADD UNIQUE KEY `uk_inventories_date_schedule` (`inventory_date`, `schedule`);

-- Remove inventory_type column
ALTER TABLE `inventories` DROP COLUMN `inventory_type`;

-- Remove inventory_frequency from items
ALTER TABLE `items` DROP KEY `idx_items_frequency_active`;
ALTER TABLE `items` DROP COLUMN `inventory_frequency`;
