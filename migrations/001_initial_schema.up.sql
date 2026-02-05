-- ============================================
-- LOOPI - Initial Schema
-- ============================================

-- Catálogo de items (productos e insumos)
CREATE TABLE `items` (
  `id` SMALLINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `type` ENUM('product', 'supply') NOT NULL,
  `name` VARCHAR(70) NOT NULL,
  `active` TINYINT(1) NOT NULL DEFAULT 1,
  `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `items_name_unique` (`name`),
  KEY `idx_items_type_active` (`type`, `active`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

-- Empleados del punto de venta
CREATE TABLE `employees` (
  `id` SMALLINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `username` VARCHAR(50) NOT NULL,
  `password_hash` VARCHAR(255) NOT NULL,
  `name` VARCHAR(50) NOT NULL,
  `last_name` VARCHAR(50) NOT NULL,
  `role` ENUM('employee', 'admin') NOT NULL DEFAULT 'employee',
  `active` TINYINT(1) NOT NULL DEFAULT 1,
  `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `employees_username_unique` (`username`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

-- Inventarios (cabecera)
CREATE TABLE `inventories` (
  `id` INT UNSIGNED NOT NULL AUTO_INCREMENT,
  `inventory_date` DATE NOT NULL,
  `schedule` ENUM('opening', 'noon', 'closing', 'weekly', 'monthly') NOT NULL,
  `status` ENUM('in_progress', 'completed') NOT NULL DEFAULT 'in_progress',
  `responsible_id` SMALLINT UNSIGNED NOT NULL,
  `started_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `completed_at` DATETIME DEFAULT NULL,
  `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `inventories_employees_FK` (`responsible_id`),
  KEY `idx_inventories_date_schedule` (`inventory_date`, `schedule`),
  UNIQUE KEY `uk_inventories_date_schedule` (`inventory_date`, `schedule`),
  CONSTRAINT `inventories_employees_FK` FOREIGN KEY (`responsible_id`) REFERENCES `employees` (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

-- Detalle de inventario por item
CREATE TABLE `inventory_details` (
  `id` INT UNSIGNED NOT NULL AUTO_INCREMENT,
  `inventory_id` INT UNSIGNED NOT NULL,
  `item_id` SMALLINT UNSIGNED NOT NULL,
  `suggested_value` SMALLINT UNSIGNED DEFAULT NULL,
  `real_value` SMALLINT UNSIGNED DEFAULT NULL,
  `stock_received` SMALLINT UNSIGNED DEFAULT NULL,
  `units_sold` SMALLINT UNSIGNED DEFAULT NULL,
  `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_inventory_item` (`inventory_id`, `item_id`),
  KEY `inventory_details_items_FK` (`item_id`),
  CONSTRAINT `inventory_details_inventories_FK` FOREIGN KEY (`inventory_id`) REFERENCES `inventories` (`id`) ON DELETE CASCADE,
  CONSTRAINT `inventory_details_items_FK` FOREIGN KEY (`item_id`) REFERENCES `items` (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

-- Novedades de inventario (discrepancias y anomalías)
CREATE TABLE `inventory_issues` (
  `id` INT UNSIGNED NOT NULL AUTO_INCREMENT,
  `inventory_detail_id` INT UNSIGNED NOT NULL,
  `type` ENUM('discrepancy', 'skipped_schedule') NOT NULL,
  `expected_value` SMALLINT UNSIGNED DEFAULT NULL,
  `actual_value` SMALLINT UNSIGNED DEFAULT NULL,
  `difference` SMALLINT DEFAULT NULL,
  `status` ENUM('open', 'resolved') NOT NULL DEFAULT 'open',
  `resolution_notes` TEXT DEFAULT NULL,
  `resolved_by` SMALLINT UNSIGNED DEFAULT NULL,
  `resolved_at` DATETIME DEFAULT NULL,
  `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `inventory_issues_detail_FK` (`inventory_detail_id`),
  KEY `inventory_issues_resolved_by_FK` (`resolved_by`),
  KEY `idx_inventory_issues_status` (`status`),
  CONSTRAINT `inventory_issues_detail_FK` FOREIGN KEY (`inventory_detail_id`) REFERENCES `inventory_details` (`id`) ON DELETE CASCADE,
  CONSTRAINT `inventory_issues_resolved_by_FK` FOREIGN KEY (`resolved_by`) REFERENCES `employees` (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;
