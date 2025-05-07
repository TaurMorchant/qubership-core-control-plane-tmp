ALTER TABLE virtual_hosts ADD COLUMN IF NOT EXISTS request_header_to_add jsonb;
ALTER TABLE virtual_hosts ADD COLUMN IF NOT EXISTS request_header_to_remove jsonb;
ALTER TABLE routes ADD COLUMN IF NOT EXISTS request_header_to_add jsonb;
ALTER TABLE routes ADD COLUMN IF NOT EXISTS request_header_to_remove jsonb;