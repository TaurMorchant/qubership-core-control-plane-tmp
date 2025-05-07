ALTER TABLE clusters ADD COLUMN IF NOT EXISTS enableh2 boolean not null default false;

UPDATE clusters SET enableh2 = true WHERE http_version = 2;