-- thresholds
CREATE TABLE IF NOT EXISTS thresholds (
    id serial primary key,
    max_connections int
);

-- circuit_breakers
CREATE TABLE IF NOT EXISTS circuit_breakers (
    id serial primary key,
    threshold_id integer references thresholds(id)
);

ALTER TABLE clusters
    ADD COLUMN IF NOT EXISTS circuit_breaker_id integer references circuit_breakers(id);

-- pg notifications thresholds
create or replace function public.notify_thresholds_changes() returns trigger as
$BODY$
begin
    if (TG_OP = 'DELETE') then
        perform pg_notify('notify_changes',
        '{"operation": "DELETE", "entity": "thresholds", "id": ' || old.id::text || '}');
return old;
end if;
    perform pg_notify('notify_changes',
        '{"operation": "UPSERT", "entity": "thresholds", "id": ' || new.id::text || '}');
return new;
end;
$BODY$
language plpgsql;

drop trigger IF exists notify_thresholds_changes_trigger on public.thresholds;
create trigger notify_thresholds_changes_trigger
    after update or insert or delete on public.thresholds
   for each row
execute procedure public.notify_thresholds_changes();

-- pg notifications thresholds
create or replace function public.notify_circuit_breakers_changes() returns trigger as
$BODY$
begin
    if (TG_OP = 'DELETE') then
        perform pg_notify('notify_changes',
        '{"operation": "DELETE", "entity": "circuit_breakers", "id": ' || old.id::text || '}');
return old;
end if;
    perform pg_notify('notify_changes',
        '{"operation": "UPSERT", "entity": "circuit_breakers", "id": ' || new.id::text || '}');
return new;
end;
$BODY$
language plpgsql;

drop trigger IF exists notify_circuit_breakers_changes_trigger on public.circuit_breakers;
create trigger notify_circuit_breakers_changes_trigger
    after update or insert or delete on public.circuit_breakers
   for each row
execute procedure public.notify_circuit_breakers_changes();