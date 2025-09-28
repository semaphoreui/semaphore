create table role
(
    `id`          integer primary key autoincrement,
    `project_id`  int,
    `slug`        varchar(50)  not null,
    `name`        varchar(100) not null,
    `permissions` int          not null default 0,

    foreign key (`project_id`) references project (`id`) on delete cascade,

    unique (`slug`)
);

create table template__role
(
    `id`          integer primary key autoincrement,
    `template_id` int not null,
    `role_id`     int not null,
    `project_id`  int not null,
    `permissions` int not null default 0,

    foreign key (`template_id`) references template (`id`) on delete cascade,
    foreign key (`role_id`) references role (`id`) on delete cascade,
    foreign key (`project_id`) references project (`id`) on delete cascade,

    unique (`template_id`, `role_id`)
);

create table view__role
(
    `id`          integer primary key autoincrement,
    `view_id`     int not null,
    `role_id`     int not null,
    `project_id`  int not null,
    `permissions` int not null default 0,

    foreign key (`view_id`) references project__view (`id`) on delete cascade,
    foreign key (`role_id`) references role (`id`) on delete cascade,
    foreign key (`project_id`) references project (`id`) on delete cascade,

    unique (`view_id`, `role_id`)
);