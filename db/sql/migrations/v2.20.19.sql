create table `project__environment_secret`
(
    `id`             integer primary key autoincrement,
    `project_id`     int          not null,
    `environment_id` int          not null,
    `type`           varchar(10)  not null,
    `name`           varchar(255){{if .Mysql}} character set utf8mb4 collate utf8mb4_bin{{end}} not null,
    `key_id`         int          not null,

    unique (`environment_id`, `type`, `name`),
    foreign key (`project_id`) references `project` (`id`) on delete cascade,
    foreign key (`environment_id`) references `project__environment` (`id`) on delete cascade,
    foreign key (`key_id`) references `access_key` (`id`)
);

create index `project__environment_secret__key_id` on `project__environment_secret` (`key_id`);
