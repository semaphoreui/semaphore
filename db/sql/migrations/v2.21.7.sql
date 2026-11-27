create table `project__inventory_group`
(
    `id`           integer primary key autoincrement,
    `inventory_id` int not null,
    `project_id`   int not null,
    `name`         varchar(250){{if .Mysql}} character set utf8mb4 collate utf8mb4_bin{{end}} not null,
    `description`  {{if .Mysql}}longtext not null{{else}}text not null default ''{{end}},
    `variables`    {{if .Mysql}}longtext not null{{else}}text not null default ''{{end}},
    `variable_count` int not null default 0,
    `created`      datetime not null,

    unique (`inventory_id`, `name`),
    foreign key (`inventory_id`) references `project__inventory` (`id`) on delete cascade,
    foreign key (`project_id`) references `project` (`id`) on delete cascade
);

create index `project__inventory_group__project_id` on `project__inventory_group` (`project_id`);

create table `project__inventory_group_host`
(
    `group_id` int not null,
    `host_id`  int not null,

    primary key (`group_id`, `host_id`),
    foreign key (`group_id`) references `project__inventory_group` (`id`) on delete cascade,
    foreign key (`host_id`) references `project__inventory_host` (`id`) on delete cascade
);

create index `project__inventory_group_host__host_id` on `project__inventory_group_host` (`host_id`);

create table `project__inventory_group_child`
(
    `parent_id` int not null,
    `child_id`  int not null,

    primary key (`parent_id`, `child_id`),
    foreign key (`parent_id`) references `project__inventory_group` (`id`) on delete cascade,
    foreign key (`child_id`) references `project__inventory_group` (`id`) on delete cascade
);

create index `project__inventory_group_child__child_id` on `project__inventory_group_child` (`child_id`);
