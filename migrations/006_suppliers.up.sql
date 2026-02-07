-- ============================================
-- LOOPI - Suppliers Table
-- ============================================

-- Create suppliers table
CREATE TABLE `suppliers` (
  `id` SMALLINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `business_name` VARCHAR(100) NOT NULL,
  `tax_id` VARCHAR(20) NOT NULL,
  `contact_name` VARCHAR(100) NOT NULL DEFAULT '',
  `contact_phone` VARCHAR(20) NOT NULL DEFAULT '',
  `contact_email` VARCHAR(100) NOT NULL DEFAULT '',
  `active` TINYINT(1) NOT NULL DEFAULT 1,
  `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `suppliers_tax_id_unique` (`tax_id`),
  KEY `idx_suppliers_active` (`active`),
  KEY `idx_suppliers_business_name` (`business_name`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;
