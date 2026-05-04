# Esquema de Base de Datos — Loopi

**Motor**: MySQL 8.0 (Cloud SQL)
**Fuente**: Análisis de migraciones SQL (001-012) + entidades Go

## Tablas Activas (6)

### employees
```sql
CREATE TABLE employees (
  id               SMALLINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
  username         VARCHAR(50) NOT NULL UNIQUE,
  password_hash    VARCHAR(255) NOT NULL,
  name             VARCHAR(50) NOT NULL,
  last_name        VARCHAR(50) NOT NULL,
  document_type    VARCHAR(10),
  document_number  VARCHAR(20),
  phone            VARCHAR(20),
  email            VARCHAR(100),
  birth_date       DATE,
  role             ENUM('employee','admin') NOT NULL DEFAULT 'employee',
  active           TINYINT(1) NOT NULL DEFAULT 1,
  created_at       DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at       DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  INDEX idx_employees_document (document_type, document_number)
);
```

### items
```sql
CREATE TABLE items (
  id                    SMALLINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
  type                  ENUM('product','supply') NOT NULL,
  name                  VARCHAR(70) NOT NULL UNIQUE,
  active                TINYINT(1) NOT NULL DEFAULT 1,
  inventory_frequency   ENUM('daily','weekly','monthly') NOT NULL DEFAULT 'monthly',
  category_id           SMALLINT UNSIGNED NOT NULL,
  supplier_id           SMALLINT UNSIGNED,
  cost                  INT UNSIGNED NOT NULL DEFAULT 0,
  measurement_unit_id   SMALLINT UNSIGNED NOT NULL,
  created_at            DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at            DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  FOREIGN KEY (category_id) REFERENCES categories(id),
  FOREIGN KEY (supplier_id) REFERENCES suppliers(id),
  FOREIGN KEY (measurement_unit_id) REFERENCES measurement_units(id),
  INDEX idx_items_type_active (type, active),
  INDEX idx_items_frequency_active (inventory_frequency, active),
  INDEX idx_items_category (category_id),
  INDEX idx_items_supplier (supplier_id),
  INDEX idx_items_measurement_unit (measurement_unit_id)
);
```

### inventories
```sql
CREATE TABLE inventories (
  id               INT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
  inventory_date   DATE NOT NULL,
  inventory_type   ENUM('daily','weekly','monthly','initial') NOT NULL,
  schedule         ENUM('opening','noon','closing'),   -- NULL para weekly/monthly/initial
  status           ENUM('in_progress','completed') NOT NULL DEFAULT 'in_progress',
  responsible_id   SMALLINT UNSIGNED NOT NULL,
  started_at       DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  completed_at     DATETIME,
  created_at       DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  FOREIGN KEY (responsible_id) REFERENCES employees(id),
  UNIQUE KEY uk_inventories_date_type_schedule (inventory_date, inventory_type, schedule),
  INDEX idx_inventories_date_schedule (inventory_date, schedule),
  INDEX idx_inventories_type (inventory_type)
);
```

### inventory_details
```sql
CREATE TABLE inventory_details (
  id               INT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
  inventory_id     INT UNSIGNED NOT NULL,
  item_id          SMALLINT UNSIGNED NOT NULL,
  suggested_value  SMALLINT UNSIGNED,
  real_value       SMALLINT UNSIGNED,
  stock_received   SMALLINT UNSIGNED,
  units_sold       SMALLINT UNSIGNED,
  shrinkage        SMALLINT UNSIGNED,
  created_at       DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at       DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  FOREIGN KEY (inventory_id) REFERENCES inventories(id) ON DELETE CASCADE,
  FOREIGN KEY (item_id) REFERENCES items(id),
  UNIQUE KEY uk_inventory_item (inventory_id, item_id)
);
```

### categories
```sql
CREATE TABLE categories (
  id             SMALLINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
  name           VARCHAR(50) NOT NULL UNIQUE,
  display_order  INT NOT NULL DEFAULT 0,
  active         TINYINT(1) NOT NULL DEFAULT 1,
  created_at     DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at     DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  INDEX idx_categories_order (display_order),
  INDEX idx_categories_active (active)
);
```

### suppliers
```sql
CREATE TABLE suppliers (
  id              SMALLINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
  business_name   VARCHAR(100) NOT NULL,
  tax_id          VARCHAR(20) NOT NULL UNIQUE,
  contact_name    VARCHAR(100) NOT NULL DEFAULT '',
  contact_phone   VARCHAR(20) NOT NULL DEFAULT '',
  contact_email   VARCHAR(100) NOT NULL DEFAULT '',
  active          TINYINT(1) NOT NULL DEFAULT 1,
  created_at      DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at      DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  INDEX idx_suppliers_active (active),
  INDEX idx_suppliers_business_name (business_name)
);
```

### measurement_units (catálogo estático)
```sql
CREATE TABLE measurement_units (
  id    SMALLINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
  code  VARCHAR(20) NOT NULL UNIQUE,
  name  VARCHAR(50) NOT NULL
);
-- Datos semilla: unit/Unidad, grams/Gramos, liters/Litros, meters/Metros, milliliters/Mililitros, ounces/Onzas
```

## Tablas Eliminadas

| Tabla | Migración | Motivo |
|-------|-----------|--------|
| `inventory_issues` | 011 | Discrepancias ahora son calculadas en el servicio, no persistidas |

## Relaciones

```
employees
  └─── inventories (responsible_id → employees.id)
         └─── inventory_details (inventory_id → inventories.id CASCADE DELETE)
                └─── items (item_id → items.id)
                       ├─── categories (category_id → categories.id)
                       ├─── suppliers (supplier_id → suppliers.id)
                       └─── measurement_units (measurement_unit_id → measurement_units.id)
```
