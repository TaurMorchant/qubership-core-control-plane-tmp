ALTER TABLE clusters ADD COLUMN IF NOT EXISTS discovery_type VARCHAR(50) NOT NULL DEFAULT 'STRICT_DNS';

UPDATE clusters SET discovery_type = type;