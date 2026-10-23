-- One row per integration considered for one webhook call: the history the Integrations page shows.
-- The alias and the integration name are kept as plain strings, and task_id/run_id carry no foreign
-- key, so a call stays readable after the alias, the integration or the task it started is deleted.
create table `project__integration_call`
(
    `id`               integer primary key autoincrement,
    `project_id`       int          not null,
    `integration_id`   int          null,
    `integration_name` varchar(255) not null,
    `alias`            varchar(255) not null,
    `created`          {{if .Mysql}}datetime(6){{else}}datetime{{end}} not null,
    `source_ip`        varchar(64)  not null,
    `method`           varchar(10)  not null,
    `outcome`          varchar(16)  not null,
    `reason`           varchar(255) not null,
    `duration`         int          not null,
    `task_id`          int          null,
    `run_id`           int          null,
    `workflow_id`      int          null,

    foreign key (`project_id`) references project (`id`) on delete cascade,
    foreign key (`integration_id`) references project__integration (`id`) on delete set null
){{if .Mysql}} default charset=utf8mb4{{end}};

-- The list pages backwards by `id desc` within one project, optionally filtered by alias.
create index `project__integration_call__project_id` on `project__integration_call` (`project_id`, `id`);
create index `project__integration_call__alias` on `project__integration_call` (`project_id`, `alias`);
