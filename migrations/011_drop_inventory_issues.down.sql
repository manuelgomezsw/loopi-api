-- Restore inventory_issues table (rollback).
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
