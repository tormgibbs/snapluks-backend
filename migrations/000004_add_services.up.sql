CREATE TABLE IF NOT EXISTS categories (
  id INT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  provider_id INTEGER NOT NULL REFERENCES providers(id) ON DELETE CASCADE,
  name TEXT NOT NULL,
  UNIQUE (provider_id, name),
  UNIQUE (provider_id, id)
);

CREATE TABLE IF NOT EXISTS service_types (
  id INT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  name TEXT NOT NULL UNIQUE
);

INSERT INTO service_types (name) VALUES
  ('Haircut'),
  ('Beard Trim'),
  ('Shave'),
  ('Hair Wash'),
  ('Hair Coloring'),
  ('Fade'),
  ('Line Up'),
  ('Hot Towel Shave'),
  ('Scalp Massage'),
  ('Kids Haircut'),
  ('Senior Haircut'),
  ('Buzz Cut'),
  ('Neck Trim'),
  ('Eyebrow Trim'),
  ('Hair Treatment');


CREATE TABLE IF NOT EXISTS services (
  id INT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  name TEXT NOT NULL,
  description TEXT,
  duration INTERVAL NOT NULL CHECK (duration > INTERVAL '0'),
  price NUMERIC(10, 2) NOT NULL CHECK (price > 0),
  type_id INTEGER NOT NULL REFERENCES service_types(id) ON DELETE CASCADE,
  provider_id INTEGER NOT NULL REFERENCES providers(id) ON DELETE CASCADE,
  primary_image_url TEXT,
  UNIQUE (provider_id, name),
  UNIQUE (provider_id, id)
);

CREATE TABLE IF NOT EXISTS staff_services (
  staff_id INTEGER NOT NULL,
  service_id INTEGER NOT NULL,
  provider_id INTEGER NOT NULL,
  PRIMARY KEY (staff_id, service_id),
  FOREIGN KEY (staff_id, provider_id) REFERENCES staff(id, provider_id),
  FOREIGN KEY (service_id, provider_id) REFERENCES services(id, provider_id)
);

CREATE TABLE IF NOT EXISTS service_categories (
  service_id INTEGER NOT NULL,
  category_id INTEGER NOT NULL,
  provider_id INTEGER NOT NULL,
  PRIMARY KEY (service_id, category_id),
  FOREIGN KEY (category_id, provider_id) REFERENCES categories(id, provider_id),
  FOREIGN KEY (service_id, provider_id) REFERENCES services(id, provider_id)
);

CREATE TABLE IF NOT EXISTS service_images (
  id INT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  service_id INTEGER NOT NULL,
  provider_id INTEGER NOT NULL,
  image_url TEXT NOT NULL,
  is_primary BOOLEAN DEFAULT FALSE,
  created_at TIMESTAMPTZ(0) DEFAULT NOW(),
  FOREIGN KEY (service_id, provider_id) REFERENCES services(id, provider_id) ON DELETE CASCADE
);

CREATE UNIQUE INDEX IF NOT EXISTS one_primary_image_per_service
ON service_images(service_id, provider_id)
WHERE is_primary = TRUE;

CREATE OR REPLACE FUNCTION update_service_primary_image()
RETURNS TRIGGER AS $$
DECLARE
  target_service_id INTEGER := COALESCE(NEW.service_id, OLD.service_id);
  target_provider_id INTEGER := COALESCE(NEW.provider_id, OLD.provider_id);
BEGIN
  IF (TG_OP = 'INSERT' AND NEW.is_primary)
     OR (TG_OP = 'UPDATE' AND NEW.is_primary AND NOT OLD.is_primary) THEN

    UPDATE services
    SET primary_image_url = NEW.image_url
    WHERE id = target_service_id AND provider_id = target_provider_id;

  ELSIF (TG_OP = 'UPDATE' AND NOT NEW.is_primary AND OLD.is_primary)
     OR (TG_OP = 'DELETE' AND OLD.is_primary) THEN

    UPDATE services
    SET primary_image_url = NULL
    WHERE id = target_service_id AND provider_id = target_provider_id;
  END IF;

  RETURN COALESCE(NEW, OLD);
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER service_primary_image_trigger
  AFTER INSERT OR UPDATE OR DELETE ON service_images
  FOR EACH ROW
  EXECUTE FUNCTION update_service_primary_image();