-- Add shrinkage (merma) column to inventory_details.
-- Only admin edits it when inventory is completed; it subtracts from expected value.
ALTER TABLE inventory_details
  ADD COLUMN shrinkage SMALLINT UNSIGNED DEFAULT NULL AFTER units_sold;
