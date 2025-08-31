CREATE TABLE absences
(
  id          INT AUTO_INCREMENT PRIMARY KEY,
  employee_id INT           NOT NULL,
  date        DATE          NOT NULL,
  hours       DECIMAL(5, 2) NOT NULL,
  reason      VARCHAR(255),
  created_at  DATETIME DEFAULT CURRENT_TIMESTAMP,
  updated_at  DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,

  CONSTRAINT fk_absence_employee FOREIGN KEY (employee_id) REFERENCES users (id)
);
