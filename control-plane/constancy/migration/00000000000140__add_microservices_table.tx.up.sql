CREATE TABLE IF NOT EXISTS microservice_versions
(
    name               text not null,
    namespace          text,
    deployment_version text not null,
    initial_version    text not null,
    primary key(name, namespace, initial_version)
);

-- pg notifications
create or replace function public.notify_microservice_versions_changes() returns trigger as
$BODY$
begin
    if (TG_OP = 'DELETE') then
        perform pg_notify('notify_changes',
        '{"operation": "DELETE", "entity": "microservice_versions", "name": ' || old.name || ', "namespace": ' || old.namespace || ', "initial_version": ' || old.initial_version || '}');
return old;
end if;
    perform pg_notify('notify_changes',
        '{"operation": "UPSERT", "entity": "microservice_versions", "name": ' || new.name || ', "namespace": ' || new.namespace || ', "deployment_version": ' || new.deployment_version || ', "initial_version": ' || new.initial_version || '}');
return new;
end;
$BODY$
language plpgsql;

drop trigger IF exists notify_microservice_versions_changes_trigger on public.microservice_versions;
create trigger notify_microservice_versions_changes_trigger
    after update or insert or delete on public.microservice_versions
   for each row
execute procedure public.notify_microservice_versions_changes();
