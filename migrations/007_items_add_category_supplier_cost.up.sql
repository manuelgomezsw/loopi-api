-- ============================================
-- LOOPI - Add category, supplier and cost to items
-- ============================================

-- Add category_id column (required, default to 1 which is "Sin categorizar")
ALTER TABLE `items` ADD COLUMN `category_id` SMALLINT UNSIGNED NOT NULL DEFAULT 1 AFTER `inventory_frequency`;
ALTER TABLE `items` ADD CONSTRAINT `fk_items_category` FOREIGN KEY (`category_id`) REFERENCES `categories`(`id`);

-- Add supplier_id column (optional)
ALTER TABLE `items` ADD COLUMN `supplier_id` SMALLINT UNSIGNED NULL AFTER `category_id`;
ALTER TABLE `items` ADD CONSTRAINT `fk_items_supplier` FOREIGN KEY (`supplier_id`) REFERENCES `suppliers`(`id`);

-- Add cost column (COP without decimals)
ALTER TABLE `items` ADD COLUMN `cost` INT UNSIGNED NOT NULL DEFAULT 0 AFTER `supplier_id`;

-- Add indexes
ALTER TABLE `items` ADD INDEX `idx_items_category` (`category_id`);
ALTER TABLE `items` ADD INDEX `idx_items_supplier` (`supplier_id`);
