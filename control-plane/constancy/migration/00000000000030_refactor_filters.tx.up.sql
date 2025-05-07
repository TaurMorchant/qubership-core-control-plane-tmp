ALTER TABLE listeners
    DROP COLUMN IF EXISTS wasm_filters;

CREATE TABLE IF NOT EXISTS wasm_filters
(
    id              serial primary key,
    name            varchar(100) not null unique,
    url             text,
    sha256          varchar(64),
    tls_config_name varchar(100),
    timeout         int,
    params          jsonb
);

CREATE TABLE IF NOT EXISTS listeners_wasm_filters
(
    listener_id    int not null,
    wasm_filter_id int not null,
    PRIMARY KEY (listener_id, wasm_filter_id)
);
