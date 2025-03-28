ALTER TABLE providers 
DROP CONSTRAINT IF EXISTS latitude_range_check,
DROP CONSTRAINT IF EXISTS longitude_range_check,
DROP CONSTRAINT IF EXISTS name_not_empty_check;
