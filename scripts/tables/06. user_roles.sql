CREATE TABLE user_roles
(
  user_id      INT,
  role_id      INT,
  franchise_id INT,
  PRIMARY KEY (user_id, role_id, franchise_id),
  FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE CASCADE,
  FOREIGN KEY (role_id) REFERENCES roles (id) ON DELETE CASCADE,
  FOREIGN KEY (franchise_id) REFERENCES franchises (id) ON DELETE CASCADE
);
