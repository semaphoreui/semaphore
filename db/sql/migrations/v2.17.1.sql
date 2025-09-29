create table role
(
    `id`          integer primary key autoincrement,
    `slug`        varchar(50)  not null,
    `name`        varchar(100) not null,
    `permissions` bigint       not null default 0,

    unique (`slug`)
);

create table project__template_role
(
    `id`          integer primary key autoincrement,
    `template_id` int    not null,
    `role_id`     int    not null,
    `project_id`  int    not null,
    `permissions` bigint not null default 0,

    foreign key (`template_id`) references project__template (`id`) on delete cascade,
    foreign key (`role_id`) references role (`id`) on delete cascade,
    foreign key (`project_id`) references project (`id`) on delete cascade,

    unique (`template_id`, `role_id`)
);