CREATE TABLE IF NOT EXISTS composite_satellites (namespace text primary key);

ALTER TABLE routes ADD COLUMN IF NOT EXISTS fallback boolean;