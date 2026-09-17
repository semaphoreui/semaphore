create table audit_event (
{{if .Postgresql}}
    seq               bigserial primary key,
    occurred_at       timestamp with time zone not null,
{{else if .Mysql}}
    seq               bigint primary key auto_increment,
    occurred_at       datetime(6) not null,
{{else}}
    seq               integer primary key autoincrement,
    occurred_at       datetime not null,
{{end}}
    event_id          varchar(36) not null,
    schema_version    varchar(16) not null,
    event_code        varchar(255) not null,
    category          varchar(64) not null,
    type              varchar(64) not null,
    action            varchar(64) not null,
    outcome           varchar(64) not null,
    actor_type        varchar(64) not null,
    actor_id          varchar(255) null,
    actor_name        varchar(255) null,
    source_ip         varchar(255) null,
    source_user_agent varchar(1024) null,
    target_type       varchar(64) null,
    target_id         varchar(255) null,
    target_name       varchar(255) null,
    project_id        varchar(255) null,
    request_id        varchar(36) null,
    instance_id       varchar(255) not null,
    node_id           varchar(255) null,
    metadata          text null
);

create unique index audit_event__event_id on audit_event(event_id);

create table audit_export_state (
    destination_id   varchar(255) primary key,
    cursor_seq       bigint not null default 0,
    owner_id         varchar(36) null,
{{if .Postgresql}}
    lease_until      timestamp with time zone null,
    last_attempt_at  timestamp with time zone null,
    last_success_at  timestamp with time zone null,
{{else if .Mysql}}
    lease_until      datetime(6) null,
    last_attempt_at  datetime(6) null,
    last_success_at  datetime(6) null,
{{else}}
    lease_until      datetime null,
    last_attempt_at  datetime null,
    last_success_at  datetime null,
{{end}}
    lease_generation bigint not null default 0,
    last_error       varchar(1024) null
);
