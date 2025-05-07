update routes
set uuid = substring(uuid from 0 for 8) || '-' || substring(uuid from 8 for 4) || '-' || substring(uuid from 12 for 4) || '-' || substring(uuid from 16 for 4) || '-' || substring(uuid from 20)
where uuid not ilike '%-%-%-%-%';
