create table project__invite
(
    `id`              integer primary key autoincrement,
    `project_id`      int          not null,
    `user_id`         int null,
    `email`           varchar(255) null,
    `role`            varchar(50)  not null,
    `status`          varchar(50)  not null default 'pending',
    `token`           varchar(255) not null,
    `inviter_user_id` int          not null,
    `created`         datetime     not null,
    `expires_at`      datetime null,
    `accepted_at`     datetime null,

    foreign key (`project_id`) references project (`id`) on delete cascade,
    foreign key (`user_id`) references `user` (`id`) on delete cascade,
    foreign key (`inviter_user_id`) references `user` (`id`) on delete cascade,

    unique (`token`),
    unique (`project_id`, `user_id`),
    unique (`project_id`, `email`)
);