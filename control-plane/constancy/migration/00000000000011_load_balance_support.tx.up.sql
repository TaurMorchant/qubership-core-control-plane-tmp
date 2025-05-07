ALTER TABLE clusters
    ADD COLUMN IF NOT EXISTS lbPolicy VARCHAR(50) NOT NULL DEFAULT 'LEAST_REQUEST';
ALTER TABLE clusters
    ADD COLUMN IF NOT EXISTS type VARCHAR(50) NOT NULL DEFAULT 'STRICT_DNS';

CREATE TABLE IF NOT EXISTS hash_policy
(
    id           serial PRIMARY KEY,
    h_headerName varchar(255),
    c_name       varchar(255),
    c_ttl        int,
    c_path       varchar(255),
    qp_sourceIp  varchar(255),
    qp_name      varchar(255),
    terminal     boolean,
    routeId      int,
    endpointId   int,
    CONSTRAINT hashpolicy_routeid_fkey FOREIGN KEY (routeId) REFERENCES routes (id),
    CONSTRAINT hashpolicy_endpointid_fkey FOREIGN KEY (endpointId) REFERENCES endpoints (id)
);