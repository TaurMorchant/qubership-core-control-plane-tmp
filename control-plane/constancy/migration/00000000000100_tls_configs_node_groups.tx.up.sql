CREATE TABLE IF NOT EXISTS tls_configs_node_groups
(
    tls_config_id     int8         references tls_configs (id) on delete cascade not null,
    node_group_name   VARCHAR(255) references node_groups (name) not null,
    PRIMARY KEY (tls_config_id, node_group_name)
);