-- ============================================
-- LOOPI - Replace TI with NUIP document type
-- ============================================

-- Update existing records that have TI to NUIP
UPDATE `employees` 
SET `document_type` = 'NUIP' 
WHERE `document_type` = 'TI';
