create table `group_role_mapping`
(
    `id`         integer primary key autoincrement,
    `provider`   varchar(100) not null,
    `group_name` varchar(255){{if .Mysql}} character set utf8mb4 collate utf8mb4_bin{{end}} not null,
    `project_id` int          not null,
    `role_id`    int          not null,
    `created`    datetime     not null,

    unique (`provider`, `group_name`, `project_id`, `role_id`),
    foreign key (`project_id`) references project (`id`) on delete cascade,
    foreign key (`role_id`) references `role` (`id`) on delete restrict
);

create index `group_role_mapping__role_id` on `group_role_mapping` (`role_id`);

create table `project__user_mapped_role`
(
    `project_id` int not null,
    `user_id`    int not null,
    `mapping_id` int not null,

    unique (`project_id`, `user_id`, `mapping_id`),
    foreign key (`project_id`) references project (`id`) on delete cascade,
    foreign key (`user_id`) references `user` (`id`) on delete cascade,
    foreign key (`mapping_id`) references `group_role_mapping` (`id`) on delete cascade
);

create index `project__user_mapped_role__user_id` on `project__user_mapped_role` (`user_id`);
create index `project__user_mapped_role__mapping_id` on `project__user_mapped_role` (`mapping_id`);
