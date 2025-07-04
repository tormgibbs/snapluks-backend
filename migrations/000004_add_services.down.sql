DROP TRIGGER IF EXISTS service_primary_image_trigger ON service_images;
DROP FUNCTION IF EXISTS update_service_primary_image;

DROP INDEX IF EXISTS one_primary_image_per_service;

DROP TABLE IF EXISTS service_images;

DROP TABLE IF EXISTS service_categories;
DROP TABLE IF EXISTS staff_services;

DROP TABLE IF EXISTS services;
DROP TABLE IF EXISTS service_types;
DROP TABLE IF EXISTS categories;
