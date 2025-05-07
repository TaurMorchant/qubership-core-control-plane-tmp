alter table header_matchers add column IF NOT EXISTS saferegexmatch varchar(255);
alter table header_matchers add column IF NOT EXISTS rangematch jsonb;
alter table header_matchers add column IF NOT EXISTS presentmatch boolean;
alter table header_matchers add column IF NOT EXISTS prefixmatch varchar(255);
alter table header_matchers add column IF NOT EXISTS suffixmatch varchar(255);
alter table header_matchers add column IF NOT EXISTS invertmatch boolean DEFAULT false;
