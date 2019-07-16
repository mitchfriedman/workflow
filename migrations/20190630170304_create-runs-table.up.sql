create table runs (
    uuid varchar(64) not null
        constraint run_workers_pkey
        primary key,

    rollback bool default false not null,

    job_name varchar(128) not null,
    scope varchar(128) not null,
    state varchar(32) not null,
    data jsonb,

    started timestamp default now_utc() not null,
    finished timestamp default null,
    last_step_complete timestamp default null,
    claimed_until timestamp default null,
    claimed_by varchar(64) default null
);

create unique index index_runs_on_uuid on runs(uuid);
create index index_runs_on_job_name on runs(job_name);
create index index_runs_on_scope on runs(scope);
create index index_runs_on_claimed_until on runs(claimed_until);
create index index_runs_on_claimed_by on runs(claimed_by);
