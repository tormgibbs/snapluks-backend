DROP TYPE IF EXISTS role;
DROP TABLE IF EXISTS users;
DROP TABLE IF EXISTS clients;
DROP TABLE IF EXISTS provider_types;
DROP TABLE IF EXISTS providers;
DROP TABLE IF EXISTS email_verifications;

DELETE FROM provider_types WHERE name IN (
  'Barber',
  'Hair Stylist',
  'Makeup Artist',
  'Nail Technician',
  'Spa Therapist'
);

