CREATE TABLE roles
(
  id          INT AUTO_INCREMENT PRIMARY KEY,
  name        VARCHAR(50) UNIQUE NOT NULL,
  description TEXT,
  is_active   BOOLEAN DEFAULT TRUE
);
