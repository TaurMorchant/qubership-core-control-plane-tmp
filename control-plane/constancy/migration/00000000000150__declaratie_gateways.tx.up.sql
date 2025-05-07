-- ext_authz_filters
CREATE TABLE IF NOT EXISTS ext_authz_filters
(
    name               text primary key,
    cluster_name       text not null,
    timeout            bigint,
    context_extensions jsonb,
    node_group text    not null
);

-- pg notifications
create or replace function public.notify_ext_authz_filters_changes() returns trigger as
$BODY$
begin
    if (TG_OP = 'DELETE') then
        perform pg_notify('notify_changes',
        '{"operation": "DELETE", "entity": "ext_authz_filters", "name": ' || old.name || '}');
return old;
end if;
    perform pg_notify('notify_changes',
        '{"operation": "UPSERT", "entity": "ext_authz_filters", "name": ' || new.name || '}');
return new;
end;
$BODY$
language plpgsql;

drop trigger IF exists notify_ext_authz_filters_changes_trigger on public.ext_authz_filters;
create trigger notify_ext_authz_filters_changes_trigger
    after update or insert or delete on public.ext_authz_filters
   for each row
execute procedure public.notify_ext_authz_filters_changes();

-- node_groups
ALTER TABLE node_groups ADD COLUMN IF NOT EXISTS gateway_type text;
ALTER TABLE node_groups ADD COLUMN IF NOT EXISTS forbid_virtual_hosts boolean;
