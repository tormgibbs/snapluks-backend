-- User roles
CREATE TYPE role AS ENUM ('client', 'provider');
-- User table
CREATE TABLE users (
	id INT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
	email citext UNIQUE NOT NULL,
	first_name TEXT,
	last_name TEXT,
	phone_number TEXT,
	password_hash bytea,
	activated BOOLEAN DEFAULT FALSE,
	role role NOT NULL DEFAULT 'client',
    created_at timestamptz(0) DEFAULT NOW()
);
-- Provider types
CREATE TABLE provider_types (
	id INT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
	name TEXT UNIQUE NOT NULL
);
-- Provider table
CREATE TABLE providers (
	id INT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
	user_id INT UNIQUE NOT NULL REFERENCES users(id) ON DELETE CASCADE,
	provider_type_id INT NOT NULL REFERENCES provider_types(id) ON DELETE RESTRICT,
	phone_number TEXT,
	email citext UNIQUE NOT NULL,
	name TEXT NOT NULL,
	description TEXT,
	latitude DOUBLE PRECISION,
	longitude DOUBLE PRECISION,
	address TEXT,
	logo_url TEXT,
	cover_url TEXT
);

CREATE TABLE provider_business_hours (
	id SERIAL PRIMARY KEY,
	provider_id INTEGER NOT NULL REFERENCES providers(id) ON DELETE CASCADE,
	day_of_week SMALLINT NOT NULL CHECK (day_of_week BETWEEN 0 AND 6), -- 0=Sunday, 6=Saturday
	is_closed BOOLEAN NOT NULL DEFAULT FALSE,
	open_time TIME,
	close_time TIME,
	CHECK (
		(is_closed = TRUE AND open_time IS NULL AND close_time IS NULL) OR
		(is_closed = FALSE AND open_time IS NOT NULL AND close_time IS NOT NULL AND open_time < close_time)
	),
	UNIQUE (provider_id, day_of_week)
);


CREATE TABLE provider_images (
  id INT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  provider_id INTEGER NOT NULL REFERENCES providers(id) ON DELETE CASCADE,
  image_url TEXT NOT NULL,
  uploaded_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_provider_images_id ON provider_images(provider_id);

-- Client table
CREATE TABLE clients (
	id INT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
	user_id INT UNIQUE NOT NULL REFERENCES users(id) ON DELETE CASCADE
);

-- Seed the provider types
INSERT INTO provider_types (name) VALUES
  ('Barber'),
  ('Hair Stylist'),
  ('Makeup Artist'),
  ('Nail Technician'),
  ('Spa Therapist');