ALTER TABLE clusters
    DROP COLUMN IF EXISTS tls;

CREATE TABLE IF NOT EXISTS tls_configs
(
    id         serial primary key,
    name       varchar(100) not null unique,
    enabled    boolean,
    insecure   boolean,
    trusted_ca text,
    sni        varchar(200)
);

ALTER TABLE clusters
    ADD COLUMN IF NOT EXISTS tls_id integer references tls_configs (id);

