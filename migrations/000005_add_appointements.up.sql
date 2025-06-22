CREATE TYPE appointment_status AS ENUM (
  'confirmed', 
  'completed',
  'cancelled',
  'no_show'
);

CREATE TABLE IF NOT EXISTS appointments (
  id INT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  service_id INTEGER NOT NULL REFERENCES services(id) ON DELETE CASCADE,
  staff_id INTEGER NOT NULL REFERENCES staff(id) ON DELETE CASCADE,
  client_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  date TIMESTAMP NOT NULL,
  status appointment_status NOT NULL DEFAULT 'confirmed',
  created_at TIMESTAMP NOT NULL DEFAULT NOW()
);