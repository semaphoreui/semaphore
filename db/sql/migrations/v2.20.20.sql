alter table `project__inventory` add `variables` {{if .Mysql}}longtext{{else}}text{{end}} null;

create table `project__inventory_host`
(
    `id`           integer primary key autoincrement,
    `inventory_id` int not null,
    `project_id`   int not null,
    `name`         varchar(250){{if .Mysql}} character set utf8mb4 collate utf8mb4_bin{{end}} not null,
    `description`  {{if .Mysql}}longtext not null{{else}}text not null default ''{{end}},
    `enabled`      bool not null default true,
    `variables`    {{if .Mysql}}longtext not null{{else}}text not null default ''{{end}},
    `created`      datetime not null,

    unique (`inventory_id`, `name`),
    foreign key (`inventory_id`) references `project__inventory` (`id`) on delete cascade,
    foreign key (`project_id`) references `project` (`id`) on delete cascade
);

create index `project__inventory_host__project_id` on `project__inventory_host` (`project_id`);
