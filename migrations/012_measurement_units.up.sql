-- ============================================
-- LOOPI - Measurement units for items
-- ============================================

-- Create measurement_units catalog (Unidad first so id=1 for default on existing items)
CREATE TABLE `measurement_units` (
  `id` SMALLINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `code` VARCHAR(20) NOT NULL,
  `name` VARCHAR(50) NOT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `measurement_units_code_unique` (`code`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

-- Seed: Unidad first (id=1), then rest alphabetically by name
INSERT INTO `measurement_units` (`code`, `name`) VALUES
  ('unit', 'Unidad'),
  ('grams', 'Gramos'),
  ('liters', 'Litros'),
  ('meters', 'Metros'),
  ('milliliters', 'Mililitros'),
  ('ounces', 'Onzas');

-- Add measurement_unit_id to items (nullable first for backfill)
ALTER TABLE `items` ADD COLUMN `measurement_unit_id` SMALLINT UNSIGNED NULL AFTER `cost`;

-- Default existing items to Unidad (id=1)
UPDATE `items` SET `measurement_unit_id` = 1 WHERE `measurement_unit_id` IS NULL;

-- Enforce NOT NULL and FK
ALTER TABLE `items` MODIFY COLUMN `measurement_unit_id` SMALLINT UNSIGNED NOT NULL;
ALTER TABLE `items` ADD CONSTRAINT `fk_items_measurement_unit` FOREIGN KEY (`measurement_unit_id`) REFERENCES `measurement_units`(`id`);
ALTER TABLE `items` ADD INDEX `idx_items_measurement_unit` (`measurement_unit_id`);
