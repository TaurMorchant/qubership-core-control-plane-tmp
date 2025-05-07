ALTER TABLE tls_configs ADD COLUMN IF NOT EXISTS client_cert text;
ALTER TABLE tls_configs ADD COLUMN IF NOT EXISTS private_key text;