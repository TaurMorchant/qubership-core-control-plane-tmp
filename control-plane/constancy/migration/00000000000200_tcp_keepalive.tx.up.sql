CREATE TABLE IF NOT EXISTS tcp_keepalive (
    id serial primary key,
    probes int,
    _time int,
    _interval int
);

ALTER TABLE clusters
    ADD COLUMN IF NOT EXISTS tcp_keepalive_id integer references tcp_keepalive(id);

-- pg notifications thresholds
create or replace function public.notify_tcp_keepalive_changes() returns trigger as
$BODY$
begin
    if (TG_OP = 'DELETE') then
        perform pg_notify('notify_changes',
        '{"operation": "DELETE", "entity": "tcp_keepalive", "id": ' || old.id::text || '}');
return old;
end if;
    perform pg_notify('notify_changes',
        '{"operation": "UPSERT", "entity": "tcp_keepalive", "id": ' || new.id::text || '}');
return new;
end;
$BODY$
language plpgsql;

drop trigger IF exists notify_tcp_keepalive_changes_trigger on public.tcp_keepalive;
create trigger notify_tcp_keepalive_changes_trigger
    after update or insert or delete on public.tcp_keepalive
   for each row
execute procedure public.notify_tcp_keepalive_changes();
