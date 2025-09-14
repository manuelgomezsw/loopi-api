-- Asignación de usuarios a tiendas
CREATE TABLE store_users
(
  id       INT AUTO_INCREMENT PRIMARY KEY,
  store_id INT NOT NULL,
  user_id  INT NOT NULL,
  FOREIGN KEY (store_id) REFERENCES stores (id) ON DELETE CASCADE,
  FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE CASCADE
);
