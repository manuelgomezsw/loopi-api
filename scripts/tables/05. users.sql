-- Usuarios y empleados unificados
CREATE TABLE users
(
  id              INT AUTO_INCREMENT PRIMARY KEY,
  first_name      VARCHAR(100)        NOT NULL,
  last_name       VARCHAR(100)        NOT NULL,
  document_type   VARCHAR(20)         NOT NULL,
  document_number VARCHAR(50)         NOT NULL,
  birthdate       DATE                NOT NULL,
  phone           VARCHAR(50)         NOT NULL,
  email           VARCHAR(100) UNIQUE NOT NULL,
  password_hash   VARCHAR(255)        NOT NULL,
  position        VARCHAR(100)        NOT NULL,
  salary          DECIMAL(10, 2)      NOT NULL,
  is_active       BOOLEAN  DEFAULT TRUE,
  created_at      DATETIME DEFAULT CURRENT_TIMESTAMP,
  updated_at      DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
);
