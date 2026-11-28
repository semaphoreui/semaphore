create table `project__inventory_source`
(
    `id`             integer primary key autoincrement,
    `project_id`     int not null,
    `inventory_id`   int not null,
    `name`           varchar(250){{if .Mysql}} character set utf8mb4 collate utf8mb4_bin{{end}} not null,
    `type`           varchar(50) not null,
    `repository_id`  int null,
    `key_id`         int null,
    `params`         {{if .Mysql}}longtext{{else}}text{{end}} null,
    `variables`      {{if .Mysql}}longtext not null{{else}}text not null default ''{{end}},
    `variable_count` int not null default 0,
    `status`         varchar(20) not null default '',
    `sync_started`   datetime null,
    `sync_run`       int not null default 0,
    `last_sync`      datetime null,
    `last_success`   datetime null,
    `last_error`     {{if .Mysql}}longtext not null{{else}}text not null default ''{{end}},
    `last_result`    {{if .Mysql}}longtext not null{{else}}text not null default ''{{end}},
    `created`        datetime not null,

    unique (`inventory_id`, `name`),
    foreign key (`inventory_id`) references `project__inventory` (`id`) on delete cascade,
    foreign key (`project_id`) references `project` (`id`) on delete cascade,
    foreign key (`repository_id`) references `project__repository` (`id`),
    foreign key (`key_id`) references `access_key` (`id`)
);

create index `project__inventory_source__project_id` on `project__inventory_source` (`project_id`);

alter table `project__inventory_host` add column `source_id` int null;
create index `project__inventory_host__source_id` on `project__inventory_host` (`source_id`);

alter table `project__inventory_group` add column `source_id` int null;
create index `project__inventory_group__source_id` on `project__inventory_group` (`source_id`);
