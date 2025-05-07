-- tls_configs_node_groups
create or replace function public.notify_tls_configs_node_groups_changes() returns trigger as
$BODY$
begin
    if (TG_OP = 'DELETE') then
        perform pg_notify('notify_changes',
        '{"operation": "DELETE", "entity": "tls_configs_node_groups", "tls_config_id": ' || old.tls_config_id::text || ', "node_group_name": "' || old.node_group_name || '"}');
return old;
end if;
    perform pg_notify('notify_changes',
            '{"operation": "UPSERT", "entity": "tls_configs_node_groups", "tls_config_id": ' || new.tls_config_id::text || ', "node_group_name": "' || new.node_group_name || '"}');
return new;
end;
$BODY$
language plpgsql;

drop trigger if exists notify_tls_configs_node_groups_changes_trigger  on public.tls_configs_node_groups;
create trigger notify_tls_configs_node_groups_changes_trigger
    after update or insert or delete on public.tls_configs_node_groups
   for each row
execute procedure public.notify_tls_configs_node_groups_changes();