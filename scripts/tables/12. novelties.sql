CREATE TABLE novelties
(
  id          INT AUTO_INCREMENT PRIMARY KEY,
  employee_id INT                           NOT NULL,
  date        DATE                          NOT NULL,
  hours       DECIMAL(5, 2)                 NOT NULL,
  type        ENUM ('positive', 'negative') NOT NULL,
  comment     VARCHAR(255),
  created_at  DATETIME DEFAULT CURRENT_TIMESTAMP,
  updated_at  DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,

  CONSTRAINT fk_novelty_employee FOREIGN KEY (employee_id) REFERENCES users (id)
);
