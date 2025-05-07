ALTER TABLE clusters ADD COLUMN IF NOT EXISTS common_lb_config jsonb;
ALTER TABLE clusters ADD COLUMN IF NOT EXISTS dns_resolvers jsonb;
ALTER TABLE endpoints ADD COLUMN IF NOT EXISTS hostname varchar(500);
ALTER TABLE endpoints ADD COLUMN IF NOT EXISTS order_id integer;