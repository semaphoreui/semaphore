create table role
(
    `slug`        varchar(100) primary key not null,
    `name`        varchar(100)             not null,
    `permissions` bigint                   not null default 0,
    `project_id`  int,
    foreign key (`project_id`) references project (`id`) on delete cascade
);

create table project__template_role
(
    `id`          integer primary key autoincrement,
    `template_id` int          not null,
    `role_slug`   varchar(100) not null,
    `project_id`  int          not null,
    `permissions` bigint       not null default 0,

    foreign key (`template_id`) references project__template (`id`) on delete cascade,
    foreign key (`role_slug`) references role (`slug`) on delete cascade,
    foreign key (`project_id`) references project (`id`) on delete cascade,

    unique (`template_id`, `role_slug`)
);