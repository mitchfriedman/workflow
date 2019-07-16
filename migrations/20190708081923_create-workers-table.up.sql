create table workers (
  uuid varchar(64) not null
    constraint workers_pkey
    primary key,

  last_updated timestamp not null,
  lease_claimed_until timestamp not null
);
