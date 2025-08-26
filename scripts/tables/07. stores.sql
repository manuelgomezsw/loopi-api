CREATE TABLE stores
(
  id           INT AUTO_INCREMENT PRIMARY KEY,
  franchise_id INT          NOT NULL,
  code         CHAR(3) UNIQUE,
  name         VARCHAR(100) NOT NULL,
  location     VARCHAR(255),
  address      VARCHAR(255),
  is_active    BOOLEAN  DEFAULT TRUE,
  created_at   DATETIME DEFAULT CURRENT_TIMESTAMP,
  updated_at   DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  FOREIGN KEY (franchise_id) REFERENCES franchises (id) ON DELETE CASCADE
);
