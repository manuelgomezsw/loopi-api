-- ============================================
-- LOOPI - Add inventory frequency and type
-- ============================================

-- Add inventory_frequency to items table
ALTER TABLE `items` 
ADD COLUMN `inventory_frequency` ENUM('daily', 'weekly', 'monthly') NOT NULL DEFAULT 'monthly' AFTER `active`;

-- Add index for filtering by frequency
ALTER TABLE `items`
ADD KEY `idx_items_frequency_active` (`inventory_frequency`, `active`);

-- Add inventory_type to inventories table
ALTER TABLE `inventories`
ADD COLUMN `inventory_type` ENUM('daily', 'weekly', 'monthly') NOT NULL DEFAULT 'daily' AFTER `inventory_date`;

-- Modify schedule to be nullable and remove weekly/monthly values
-- First, update any existing weekly/monthly schedules
UPDATE `inventories` SET `inventory_type` = 'weekly', `schedule` = NULL WHERE `schedule` = 'weekly';
UPDATE `inventories` SET `inventory_type` = 'monthly', `schedule` = NULL WHERE `schedule` = 'monthly';

-- Drop the unique key that includes schedule (we'll recreate it)
ALTER TABLE `inventories` DROP KEY `uk_inventories_date_schedule`;

-- Modify schedule column to only allow opening/noon/closing and be nullable
ALTER TABLE `inventories` 
MODIFY COLUMN `schedule` ENUM('opening', 'noon', 'closing') NULL;

-- Create new unique key that considers both type and schedule
-- For daily inventories: unique by date + type + schedule
-- For weekly/monthly: unique by date + type (schedule is NULL)
ALTER TABLE `inventories`
ADD UNIQUE KEY `uk_inventories_date_type_schedule` (`inventory_date`, `inventory_type`, `schedule`);

-- Add index for filtering by type
ALTER TABLE `inventories`
ADD KEY `idx_inventories_type` (`inventory_type`);
