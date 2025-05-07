ALTER TABLE clusters ADD COLUMN IF NOT EXISTS http_version int not null default 1;

UPDATE clusters SET http_version = 2 WHERE enableh2 = true;

ALTER TABLE clusters DROP COLUMN IF EXISTS enableh2;