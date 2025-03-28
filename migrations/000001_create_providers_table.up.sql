-- Create providers table
CREATE TABLE IF NOT EXISTS providers (
  id SERIAL PRIMARY KEY,
  name VARCHAR(255) NOT NULL,
  address VARCHAR(255),
  latitude DECIMAL(9, 6),
  longitude DECIMAL(9, 6),
  version INTEGER NOT NULL DEFAULT 1,
  created_at timestamp(0) with time zone NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMP(0) WITH TIME ZONE DEFAULT NULL
);

-- Create function to update timestamp and version
CREATE OR REPLACE FUNCTION update_timestamp()
RETURNS TRIGGER AS $$
BEGIN
  NEW.updated_at = NOW();
  NEW.version = COALESCE(OLD.version, 0) + 1;
  RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Create trigger to automatically update timestamp and version
CREATE TRIGGER set_updated_at
BEFORE UPDATE ON providers
FOR EACH ROW
EXECUTE FUNCTION update_timestamp();