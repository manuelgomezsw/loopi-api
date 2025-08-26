CREATE TABLE permissions
(
  id          INT AUTO_INCREMENT PRIMARY KEY,
  name        VARCHAR(100) UNIQUE NOT NULL,
  description TEXT
);
