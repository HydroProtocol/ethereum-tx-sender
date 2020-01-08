create table launch_logs
(
    id          serial                not null
        constraint launch_logs_pkey
            primary key,
    created_at  timestamp with time zone,
    updated_at  timestamp with time zone,
    deleted_at  timestamp with time zone,
    "from"      text                  not null,
    "to"        text                  not null,
    value       text                  not null,
    gas_limit   bigint                not null,
    gas_used    bigint  default 0     not null,
    executed_at bigint  default 0     not null,
    status      text                  not null,
    gas_price   text                  not null,
    data        bytea                 not null,
    item_type   text                  not null,
    item_id     text                  not null,
    hash        text,
    err_msg     text,
    is_urgent   boolean default false not null,
    nonce       bigint
);

alter table launch_logs
    owner to launcher;

create index idx_launch_logs_deleted_at
    on launch_logs (deleted_at);

create index idx_launch_logs_from
    on launch_logs ("from");

create index idx_launch_logs_status
    on launch_logs (status);

create index idx_launch_logs_item
    on launch_logs (item_type, item_id);

create unique index uix_launch_logs_hash
    on launch_logs (hash);

create sequence block_number_serial;
