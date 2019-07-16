create or replace function now_utc() returns timestamp as $$
  select now() at time zone 'utc';
$$ language sql;