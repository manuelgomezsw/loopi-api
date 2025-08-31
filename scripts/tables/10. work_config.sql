CREATE TABLE work_config
(
  id            INT AUTO_INCREMENT PRIMARY KEY,
  diurnal_start TIME NOT NULL,
  diurnal_end   TIME NOT NULL,
  is_active     BOOLEAN DEFAULT TRUE
);
