CREATE TABLE shifts
(
  id            INT AUTO_INCREMENT PRIMARY KEY,
  store_id      INT          NOT NULL,
  name          VARCHAR(100) NOT NULL,
  period ENUM ('weekly', 'biweekly', 'monthly') NOT NULL,
  start_time    TIME         NOT NULL,
  end_time      TIME         NOT NULL,
  lunch_minutes INT     DEFAULT 0,
  is_active     BOOLEAN DEFAULT TRUE,
  FOREIGN KEY (store_id) REFERENCES stores (id) ON DELETE CASCADE
);
