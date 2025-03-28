ALTER TABLE providers 
ADD CONSTRAINT latitude_range_check CHECK (latitude BETWEEN -90 AND 90),
ADD CONSTRAINT longitude_range_check CHECK (longitude BETWEEN -180 AND 180),
ADD CONSTRAINT name_not_empty_check CHECK (char_length(name) > 0);

