CREATE TABLE IF NOT EXISTS rate_limits
(
    name                      text not null,
    limit_requests_per_second integer,
    priority                  integer,
    primary key(name, priority)
);

ALTER TABLE routes ADD COLUMN IF NOT EXISTS rate_limit_id text;
ALTER TABLE virtual_hosts ADD COLUMN IF NOT EXISTS rate_limit_id text;

-- pg notifications
create or replace function public.notify_rate_limits_changes() returns trigger as
$BODY$
begin
    if (TG_OP = 'DELETE') then
        perform pg_notify('notify_changes',
        '{"operation": "DELETE", "entity": "rate_limits", "name": ' || old.name || ', "priority": ' || old.priority::text || '}');
return old;
end if;
    perform pg_notify('notify_changes',
        '{"operation": "UPSERT", "entity": "rate_limits", "name": ' || new.name || ', "priority": ' || new.priority::text || ', "limit_requests_per_second": ' || new.limit_requests_per_second::text || '}');
return new;
end;
$BODY$
language plpgsql;

drop trigger IF exists notify_rate_limits_changes_trigger on public.rate_limits;
create trigger notify_rate_limits_changes_trigger
    after update or insert or delete on public.rate_limits
   for each row
execute procedure public.notify_rate_limits_changes();

