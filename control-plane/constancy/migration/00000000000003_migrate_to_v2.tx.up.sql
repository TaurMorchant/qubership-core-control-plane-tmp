CREATE TABLE IF NOT EXISTS node_groups
(
    name VARCHAR(100) PRIMARY KEY
);

INSERT INTO node_groups(name)
VALUES ('public-gateway-service'),
       ('private-gateway-service'),
       ('internal-gateway-service') ON CONFLICT (name) DO NOTHING;

CREATE TABLE IF NOT EXISTS clusters_node_groups
(
    clusters_id     int8         not null,
    nodeGroups_name VARCHAR(255) not null,
    PRIMARY KEY (clusters_id, nodeGroups_name)
);

CREATE TABLE IF NOT EXISTS deployment_versions
(
    version VARCHAR(50) PRIMARY KEY,
    stage   VARCHAR(50) NOT NULL
);

INSERT INTO deployment_versions(version, stage)
VALUES ('v1', 'ACTIVE') ON CONFLICT (version) DO NOTHING;

CREATE TABLE IF NOT EXISTS endpoints
(
    id                 SERIAL PRIMARY KEY,
    clusterId          INT,
    address            VARCHAR(500) NOT NULL,
    port               SMALLINT     NOT NULL,
    deployment_version VARCHAR(50) REFERENCES deployment_versions (version)
);

ALTER TABLE routes
    ADD COLUMN IF NOT EXISTS deployment_version varchar(50) DEFAULT 'v1';

ALTER TABLE routes DROP CONSTRAINT IF EXISTS deployment_version_fk;
ALTER TABLE routes
    ADD CONSTRAINT deployment_version_fk FOREIGN KEY (deployment_version) REFERENCES deployment_versions (version);

ALTER TABLE listeners DROP CONSTRAINT IF EXISTS listeners_nodeGroup_fk;
ALTER TABLE listeners
    ADD CONSTRAINT listeners_nodeGroup_fk FOREIGN KEY (nodeGroup) REFERENCES node_groups (name);

ALTER TABLE route_configurations DROP CONSTRAINT IF EXISTS route_configurations_nodeGroup_fk;
ALTER TABLE route_configurations
    ADD CONSTRAINT route_configurations_nodeGroup_fk FOREIGN KEY (nodeGroup) REFERENCES node_groups (name);

UPDATE routes SET version = 0 WHERE version is null;
UPDATE clusters SET version = 0 WHERE version is null;
UPDATE route_configurations SET version = 0 WHERE version is null;
UPDATE listeners SET version = 0 WHERE version is null;
UPDATE virtual_hosts SET version = 0 WHERE version is null;
UPDATE virtual_host_domains SET version = 0 WHERE version is null;