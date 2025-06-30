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