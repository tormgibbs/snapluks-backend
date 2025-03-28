-- Remove the trigger
DROP TRIGGER IF EXISTS set_updated_at ON providers;

-- Remove the update timestamp function
DROP FUNCTION IF EXISTS update_timestamp();

-- Drop the providers table
DROP TABLE IF EXISTS providers;