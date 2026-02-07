-- ============================================
-- LOOPI - Rollback NUIP to TI document type
-- ============================================

-- Revert NUIP back to TI
UPDATE `employees` 
SET `document_type` = 'TI' 
WHERE `document_type` = 'NUIP';
