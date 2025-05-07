create table if not exists stateful_session
(
    id serial PRIMARY KEY,

    enabled     boolean default true,
    cookie_name varchar(255) not null,
    cookie_ttl  int,
    cookie_path varchar(255),

    clusterName              varchar(255),
    namespace                varchar(255),
    deployment_version       varchar(50) references deployment_versions (version) not null,
    initialDeploymentVersion varchar(64) not null,

    gateways jsonb not null
);

ALTER TABLE routes DROP CONSTRAINT IF EXISTS routes_statefulsessionid_fk;
ALTER TABLE routes ADD COLUMN IF NOT EXISTS statefulsessionid integer UNIQUE;
ALTER TABLE routes ADD CONSTRAINT routes_statefulsessionid_fk FOREIGN KEY (statefulsessionid) REFERENCES stateful_session (id);

ALTER TABLE endpoints DROP CONSTRAINT IF EXISTS endpoints_statefulsessionid_fk;
ALTER TABLE endpoints ADD COLUMN IF NOT EXISTS statefulsessionid integer UNIQUE;
ALTER TABLE endpoints ADD CONSTRAINT endpoints_statefulsessionid_fk FOREIGN KEY (statefulsessionid) REFERENCES stateful_session (id);

create or replace function public.notify_stateful_session_changes() returns trigger as
$BODY$
begin
    if (TG_OP = 'DELETE') then
        perform pg_notify('notify_changes', '{"operation": "DELETE", "entity": "stateful_session", "id": ' || old.id::text || '}');
return old;
end if;
    perform pg_notify('notify_changes', '{"operation": "UPSERT", "entity": "stateful_session", "id": ' || new.id::text || '}');
return new;
end;
$BODY$
language plpgsql;

drop trigger IF exists notify_stateful_session_changes_trigger on public.stateful_session;
create trigger notify_stateful_session_changes_trigger
    after update or insert or delete on public.stateful_session
   for each row
execute procedure public.notify_stateful_session_changes();