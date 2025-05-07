create table if not exists system_properties (
    name varchar(50) primary key,
    value varchar(100),
    version int
);

create table if not exists clusters (
    id serial primary key,
    nodeGroup varchar(100) not null,
    name varchar(500) not null,
    host varchar(500) not null,
    port varchar(5),
    version int
);

create table if not exists listeners (
    id serial primary key,
    nodeGroup varchar(100) not null,
    bindHost varchar(100) not null,
    bindPort varchar(5) not null,
    name varchar(100) not null,
    routeConfigName varchar(100) not null,
    version int
);

create table if not exists route_configurations (
    id serial primary key,
    nodeGroup varchar(100) not null,
    name varchar(100) not null,
    version int
);

create table if not exists virtual_hosts (
    id serial primary key,
    routeConfigId integer references route_configurations(id),
    name varchar(100) not null,
    version int
);

create table if not exists virtual_host_domains (
    virtualHostId integer references virtual_hosts(id),
    domain varchar(100) not null,
    version int
);

create table if not exists routes (
    id serial primary key,
    virtualHostId integer references virtual_hosts(id),
    routeKey varchar(500) not null,
    rm_prefix varchar(500),
    rm_regExp varchar(500),
    rm_path varchar(100),
    ra_clusterName varchar(500),
    ra_hostRewrite text,
    ra_hostAutoRewrite boolean,
    ra_prefixRewrite varchar(500),
    ra_pathRewrite varchar(100),
    directResponse_status smallint,
    version int
);

create table if not exists header_matchers (
    id serial primary key,
    routeId integer references routes(id),
    name varchar(100),
    exactMatch varchar(100),
    version int
);

ALTER TABLE system_properties ADD COLUMN IF NOT EXISTS version integer;
ALTER TABLE listeners ADD COLUMN IF NOT EXISTS version integer;
ALTER TABLE route_configurations ADD COLUMN IF NOT EXISTS version integer;
ALTER TABLE virtual_hosts ADD COLUMN IF NOT EXISTS version integer;
ALTER TABLE routes ADD COLUMN IF NOT EXISTS version integer;
ALTER TABLE header_matchers ADD COLUMN IF NOT EXISTS version integer;
ALTER TABLE virtual_host_domains ADD COLUMN IF NOT EXISTS version integer;
ALTER TABLE routes ADD COLUMN IF NOT EXISTS timeout bigint;
ALTER TABLE clusters ADD COLUMN IF NOT EXISTS enableh2 boolean;

ALTER TABLE routes ALTER COLUMN routeKey TYPE varchar(500);

DELETE FROM routes WHERE rm_prefix like '%/' AND trim(trailing '/' from rm_prefix) IN (SELECT rm_prefix FROM routes);
DELETE FROM clusters
WHERE id IN
      (SELECT id
       FROM (SELECT id,
                    ROW_NUMBER() OVER ( PARTITION BY name
                        ORDER BY id ) AS row_num
             FROM clusters) t
       WHERE t.row_num > 1);