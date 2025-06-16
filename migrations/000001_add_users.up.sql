-- User roles
CREATE TYPE role AS ENUM ('client', 'provider');
-- User table
CREATE TABLE users (
	id INT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
	email citext UNIQUE NOT NULL,
	first_name TEXT NOT NULL,
	last_name TEXT NOT NULL,
	phone_number TEXT,
	password_hash bytea NOT NULL,
	activated BOOLEAN DEFAULT FALSE,
	role role NOT NULL DEFAULT 'client'
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
	name TEXT NOT NULL,
	description TEXT,
	phone_number TEXT,
	latitude DOUBLE PRECISION,
	longitude DOUBLE PRECISION,
	address TEXT,
	logo_url TEXT,
	cover_url TEXT
);

-- Client table
CREATE TABLE clients (
	id INT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
	user_id INT UNIQUE NOT NULL REFERENCES users(id) ON DELETE CASCADE
);
CREATE TABLE email_verifications (
	hash BYTEA PRIMARY KEY,
	email TEXT NOT NULL,
	expiry TIMESTAMPTZ(0) NOT NULL
);
-- Seed the provider types
INSERT INTO provider_types (name) VALUES
  ('Barber'),
  ('Hair Stylist'),
  ('Makeup Artist'),
  ('Nail Technician'),
  ('Spa Therapist');