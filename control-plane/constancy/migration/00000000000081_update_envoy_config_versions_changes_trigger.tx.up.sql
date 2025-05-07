DROP TRIGGER IF EXISTS notify_envoy_config_versions_changes_trigger on public.envoy_config_versions;

create CONSTRAINT trigger notify_envoy_config_versions_changes_trigger
    after update or insert or delete on public.envoy_config_versions
    DEFERRABLE INITIALLY DEFERRED
    for each row
execute procedure public.notify_envoy_config_versions_changes();