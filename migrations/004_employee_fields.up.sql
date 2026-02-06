-- ============================================
-- Add new fields to employees table
-- ============================================

ALTER TABLE employees 
  ADD COLUMN document_type VARCHAR(10) AFTER last_name,
  ADD COLUMN document_number VARCHAR(20) AFTER document_type,
  ADD COLUMN phone VARCHAR(20) AFTER document_number,
  ADD COLUMN email VARCHAR(100) AFTER phone,
  ADD COLUMN birth_date DATE AFTER email;

-- Add index for document lookup
CREATE INDEX idx_employees_document ON employees (document_type, document_number);
