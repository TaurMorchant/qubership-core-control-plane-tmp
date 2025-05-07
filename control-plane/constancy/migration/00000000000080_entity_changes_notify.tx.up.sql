-- clusters
create or replace function public.notify_clusters_changes() returns trigger as
$BODY$
begin
    if (TG_OP = 'DELETE') then
        perform pg_notify('notify_changes', '{"operation": "DELETE", "entity": "clusters", "name": "' || old.name || '"}');
return old;
end if;
    perform pg_notify('notify_changes', '{"operation": "UPSERT", "entity": "clusters", "name": "' || new.name || '"}');
return new;
end;
$BODY$
language plpgsql;

drop trigger IF exists notify_clusters_changes_trigger on public.clusters;
create trigger notify_clusters_changes_trigger
    after update or insert or delete on public.clusters
   for each row
execute procedure public.notify_clusters_changes();

-- clusters_node_groups
create or replace function public.notify_clusters_node_groups_changes() returns trigger as
$BODY$
begin
    if (TG_OP = 'DELETE') then
        perform pg_notify('notify_changes',
            '{"operation": "DELETE", "entity": "clusters_node_groups", "clusters_id": ' || old.clusters_id::text || ', "nodegroups_name": "' || old.nodegroups_name || '"}');
return old;
end if;
    perform pg_notify('notify_changes',
            '{"operation": "UPSERT", "entity": "clusters_node_groups", "clusters_id": ' || new.clusters_id::text || ', "nodegroups_name": "' || new.nodegroups_name || '"}');
return new;
end;
$BODY$
language plpgsql;

drop trigger IF exists notify_clusters_node_groups_changes_trigger on public.clusters_node_groups;
create trigger notify_clusters_node_groups_changes_trigger
    after update or insert or delete on public.clusters_node_groups
   for each row
execute procedure public.notify_clusters_node_groups_changes();

-- composite_satellites
create or replace function public.notify_composite_satellites_changes() returns trigger as
$BODY$
begin
    if (TG_OP = 'DELETE') then
        perform pg_notify('notify_changes', '{"operation": "DELETE", "entity": "composite_satellites", "namespace": "' || old.namespace || '"}');
return old;
end if;
    perform pg_notify('notify_changes', '{"operation": "UPSERT", "entity": "composite_satellites", "namespace": "' || new.namespace || '"}');
return new;
end;
$BODY$
language plpgsql;

drop trigger IF exists notify_composite_satellites_changes_trigger on public.composite_satellites;
create trigger notify_composite_satellites_changes_trigger
    after update or insert or delete on public.composite_satellites
   for each row
execute procedure public.notify_composite_satellites_changes();

-- deployment_versions
create or replace function public.notify_deployment_versions_changes() returns trigger as
$BODY$
begin
    if (TG_OP = 'DELETE') then
        perform pg_notify('notify_changes', '{"operation": "DELETE", "entity": "deployment_versions", "version": "' || old.version || '"}');
return old;
end if;
    perform pg_notify('notify_changes', '{"operation": "UPSERT", "entity": "deployment_versions", "version": "' || new.version || '"}');
return new;
end;
$BODY$
language plpgsql;

drop trigger IF exists notify_deployment_versions_changes_trigger on public.deployment_versions;
create trigger notify_deployment_versions_changes_trigger
    after update or insert or delete on public.deployment_versions
   for each row
execute procedure public.notify_deployment_versions_changes();

-- endpoints
create or replace function public.notify_endpoints_changes() returns trigger as
$BODY$
begin
    if (TG_OP = 'DELETE') then
        perform pg_notify('notify_changes', '{"operation": "DELETE", "entity": "endpoints", "id": ' || old.id::text || '}');
return old;
end if;
    perform pg_notify('notify_changes', '{"operation": "UPSERT", "entity": "endpoints", "id": ' || new.id::text || '}');
return new;
end;
$BODY$
language plpgsql;

drop trigger IF exists notify_endpoints_changes_trigger on public.endpoints;
create trigger notify_endpoints_changes_trigger
    after update or insert or delete on public.endpoints
   for each row
execute procedure public.notify_endpoints_changes();

-- envoy_config_versions
create or replace function public.notify_envoy_config_versions_changes() returns trigger as
$BODY$
begin
    if (TG_OP = 'DELETE') then
        perform pg_notify('notify_changes',
            '{"operation": "DELETE", "entity": "envoy_config_version", "node_group": "' || old.node_group ||
            '", "entity_type": "' || old.entity_type || '", "version": ' || old.version::text || '}');
return old;
end if;
    perform pg_notify('notify_changes',
            '{"operation": "UPSERT", "entity": "envoy_config_version", "node_group": "' || new.node_group ||
            '", "entity_type": "' || new.entity_type || '", "version": ' || new.version::text || '}');
return new;
end;
$BODY$
language plpgsql;

drop trigger IF exists notify_envoy_config_versions_changes_trigger on public.envoy_config_versions;
create trigger notify_envoy_config_versions_changes_trigger
    after update or insert or delete on public.envoy_config_versions
   for each row
execute procedure public.notify_envoy_config_versions_changes();

-- hash_policy
create or replace function public.notify_hash_policy_changes() returns trigger as
$BODY$
begin
    if (TG_OP = 'DELETE') then
        perform pg_notify('notify_changes', '{"operation": "DELETE", "entity": "hash_policy", "id": ' || old.id::text || '}');
return old;
end if;
    perform pg_notify('notify_changes', '{"operation": "UPSERT", "entity": "hash_policy", "id": ' || new.id::text || '}');
return new;
end;
$BODY$
language plpgsql;

drop trigger IF exists notify_hash_policy_changes_trigger on public.hash_policy;
create trigger notify_hash_policy_changes_trigger
    after update or insert or delete on public.hash_policy
   for each row
execute procedure public.notify_hash_policy_changes();

-- header_matchers
create or replace function public.notify_header_matchers_changes() returns trigger as
$BODY$
begin
    if (TG_OP = 'DELETE') then
        perform pg_notify('notify_changes', '{"operation": "DELETE", "entity": "header_matchers", "id": ' || old.id::text || '}');
return old;
end if;
    perform pg_notify('notify_changes', '{"operation": "UPSERT", "entity": "header_matchers", "id": ' || new.id::text || '}');
return new;
end;
$BODY$
language plpgsql;

drop trigger IF exists notify_header_matchers_changes_trigger on public.header_matchers;
create trigger notify_header_matchers_changes_trigger
    after update or insert or delete on public.header_matchers
   for each row
execute procedure public.notify_header_matchers_changes();

-- health_check
create or replace function public.notify_health_check_changes() returns trigger as
$BODY$
begin
    if (TG_OP = 'DELETE') then
        perform pg_notify('notify_changes', '{"operation": "DELETE", "entity": "health_check", "id": ' || old.id::text || '}');
return old;
end if;
    perform pg_notify('notify_changes', '{"operation": "UPSERT", "entity": "health_check", "id": ' || new.id::text || '}');
return new;
end;
$BODY$
language plpgsql;

drop trigger IF exists notify_health_check_changes_trigger on public.health_check;
create trigger notify_health_check_changes_trigger
    after update or insert or delete on public.health_check
   for each row
execute procedure public.notify_health_check_changes();

-- listeners
create or replace function public.notify_listeners_changes() returns trigger as
$BODY$
begin
    if (TG_OP = 'DELETE') then
        perform pg_notify('notify_changes', '{"operation": "DELETE", "entity": "listeners", "id": ' || old.id::text || '}');
return old;
end if;
    perform pg_notify('notify_changes', '{"operation": "UPSERT", "entity": "listeners", "id": ' || new.id::text || '}');
return new;
end;
$BODY$
language plpgsql;

drop trigger IF exists notify_listeners_changes_trigger on public.listeners;
create trigger notify_listeners_changes_trigger
    after update or insert or delete on public.listeners
   for each row
execute procedure public.notify_listeners_changes();

-- listeners_wasm_filters
create or replace function public.notify_listeners_wasm_filters_changes() returns trigger as
$BODY$
begin
    if (TG_OP = 'DELETE') then
        perform pg_notify('notify_changes', '{"operation": "DELETE", "entity": "listeners_wasm_filters", "listener_id": '
            || old.listener_id::text || ', "wasm_filter_id": ' || old.wasm_filter_id::text || '}');
return old;
end if;
    perform pg_notify('notify_changes', '{"operation": "UPSERT", "entity": "listeners_wasm_filters", "listener_id": '
            || new.listener_id::text || ', "wasm_filter_id": ' || new.wasm_filter_id::text || '}');
return new;
end;
$BODY$
language plpgsql;

drop trigger IF exists notify_listeners_wasm_filters_changes_trigger on public.listeners_wasm_filters;
create trigger notify_listeners_wasm_filters_changes_trigger
    after update or insert or delete on public.listeners_wasm_filters
   for each row
execute procedure public.notify_listeners_wasm_filters_changes();

-- node_groups
create or replace function public.notify_node_groups_changes() returns trigger as
$BODY$
begin
    if (TG_OP = 'DELETE') then
        perform pg_notify('notify_changes', '{"operation": "DELETE", "entity": "node_groups", "name": "' || old.name || '"}');
return old;
end if;
    perform pg_notify('notify_changes', '{"operation": "UPSERT", "entity": "node_groups", "name": "' || new.name || '"}');
return new;
end;
$BODY$
language plpgsql;

drop trigger IF exists notify_node_groups_changes_trigger on public.node_groups;
create trigger notify_node_groups_changes_trigger
    after update or insert or delete on public.node_groups
   for each row
execute procedure public.notify_node_groups_changes();

-- retry_policy
create or replace function public.notify_retry_policy_changes() returns trigger as
$BODY$
begin
    if (TG_OP = 'DELETE') then
        perform pg_notify('notify_changes', '{"operation": "DELETE", "entity": "retry_policy", "id": ' || old.id::text || '}');
return old;
end if;
    perform pg_notify('notify_changes', '{"operation": "UPSERT", "entity": "retry_policy", "id": ' || new.id::text || '}');
return new;
end;
$BODY$
language plpgsql;

drop trigger IF exists notify_retry_policy_changes_trigger on public.retry_policy;
create trigger notify_retry_policy_changes_trigger
    after update or insert or delete on public.retry_policy
   for each row
execute procedure public.notify_retry_policy_changes();

-- route_configurations
create or replace function public.notify_route_configurations_changes() returns trigger as
$BODY$
begin
    if (TG_OP = 'DELETE') then
        perform pg_notify('notify_changes', '{"operation": "DELETE", "entity": "route_configurations", "id": ' || old.id::text || '}');
return old;
end if;
    perform pg_notify('notify_changes', '{"operation": "UPSERT", "entity": "route_configurations", "id": ' || new.id::text || '}');
return new;
end;
$BODY$
language plpgsql;

drop trigger IF exists notify_route_configurations_changes_trigger on public.route_configurations;
create trigger notify_route_configurations_changes_trigger
    after update or insert or delete on public.route_configurations
   for each row
execute procedure public.notify_route_configurations_changes();

-- routes
create or replace function public.notify_routes_changes() returns trigger as
$BODY$
begin
    if (TG_OP = 'DELETE') then
        perform pg_notify('notify_changes', '{"operation": "DELETE", "entity": "routes", "id": ' || old.id::text || '}');
return old;
end if;
    perform pg_notify('notify_changes', '{"operation": "UPSERT", "entity": "routes", "id": ' || new.id::text || '}');
return new;
end;
$BODY$
language plpgsql;

drop trigger IF exists notify_routes_changes_trigger on public.routes;
create trigger notify_routes_changes_trigger
    after update or insert or delete on public.routes
   for each row
execute procedure public.notify_routes_changes();

-- tls_configs
create or replace function public.notify_tls_configs_changes() returns trigger as
$BODY$
begin
    if (TG_OP = 'DELETE') then
        perform pg_notify('notify_changes', '{"operation": "DELETE", "entity": "tls_configs", "id": ' || old.id::text || '}');
return old;
end if;
    perform pg_notify('notify_changes', '{"operation": "UPSERT", "entity": "tls_configs", "id": ' || new.id::text || '}');
return new;
end;
$BODY$
language plpgsql;

drop trigger IF exists notify_tls_configs_changes_trigger on public.tls_configs;
create trigger notify_tls_configs_changes_trigger
    after update or insert or delete on public.tls_configs
   for each row
execute procedure public.notify_tls_configs_changes();

-- virtual_host_domains
create or replace function public.notify_virtual_host_domains_changes() returns trigger as
$BODY$
begin
    if (TG_OP = 'DELETE') then
        perform pg_notify('notify_changes', '{"operation": "DELETE", "entity": "virtual_host_domains", "virtualhostid": '
            || old.virtualhostid::text || ', "domain": "' || old.domain || '"}');
return old;
end if;
    perform pg_notify('notify_changes', '{"operation": "UPSERT", "entity": "virtual_host_domains", "virtualhostid": '
            || new.virtualhostid::text || ', "domain": "' || new.domain || '"}');
return new;
end;
$BODY$
language plpgsql;

drop trigger IF exists notify_virtual_host_domains_changes_trigger on public.virtual_host_domains;
create trigger notify_virtual_host_domains_changes_trigger
    after update or insert or delete on public.virtual_host_domains
   for each row
execute procedure public.notify_virtual_host_domains_changes();

-- virtual_hosts
create or replace function public.notify_virtual_hosts_changes() returns trigger as
$BODY$
begin
    if (TG_OP = 'DELETE') then
        perform pg_notify('notify_changes', '{"operation": "DELETE", "entity": "virtual_hosts", "id": ' || old.id::text || '}');
return old;
end if;
    perform pg_notify('notify_changes', '{"operation": "UPSERT", "entity": "virtual_hosts", "id": ' || new.id::text || '}');
return new;
end;
$BODY$
language plpgsql;

drop trigger IF exists notify_virtual_hosts_changes_trigger on public.virtual_hosts;
create trigger notify_virtual_hosts_changes_trigger
    after update or insert or delete on public.virtual_hosts
   for each row
execute procedure public.notify_virtual_hosts_changes();

-- wasm_filters
create or replace function public.notify_wasm_filters_changes() returns trigger as
$BODY$
begin
    if (TG_OP = 'DELETE') then
        perform pg_notify('notify_changes', '{"operation": "DELETE", "entity": "wasm_filters", "id": ' || old.id::text || '}');
return old;
end if;
    perform pg_notify('notify_changes', '{"operation": "UPSERT", "entity": "wasm_filters", "id": ' || new.id::text || '}');
return new;
end;
$BODY$
language plpgsql;

drop trigger IF exists notify_wasm_filters_changes_trigger on public.wasm_filters;
create trigger notify_wasm_filters_changes_trigger
    after update or insert or delete on public.wasm_filters
   for each row
execute procedure public.notify_wasm_filters_changes();