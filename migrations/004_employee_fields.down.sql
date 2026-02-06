-- ============================================
-- Rollback: Remove new fields from employees table
-- ============================================

DROP INDEX idx_employees_document ON employees;

ALTER TABLE employees 
  DROP COLUMN document_type,
  DROP COLUMN document_number,
  DROP COLUMN phone,
  DROP COLUMN email,
  DROP COLUMN birth_date;
