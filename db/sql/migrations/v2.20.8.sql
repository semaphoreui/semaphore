drop table if exists `project__invite_new`;
drop table if exists `project__template_role_new`;
drop table if exists `project__user_new`;
drop table if exists `role_new`;

create table `role_new`
(
    `id`          integer primary key autoincrement,
    `name`        varchar(100) not null,
    `permissions` bigint       not null default 0,
    `project_id`  int,
    `builtin_key` varchar(50),
    `slug`        varchar(100),

    foreign key (`project_id`) references project (`id`) on delete cascade,
    unique (`builtin_key`),
    check (`builtin_key` is null or `builtin_key` in ('owner', 'manager', 'task_runner', 'guest')),
    check (`builtin_key` is null or `project_id` is null)
);

create unique index `role_new__slug` on `role_new` (`slug`);

insert into `role_new` (`slug`, `name`, `permissions`, `project_id`, `builtin_key`)
values ('owner', 'Owner', 15, null, 'owner'),
       ('manager', 'Manager', 5, null, 'manager'),
       ('task_runner', 'Task Runner', 1, null, 'task_runner'),
       ('guest', 'Guest', 0, null, 'guest');

insert into `role_new` (`slug`, `name`, `permissions`, `project_id`, `builtin_key`)
select `slug`, `name`, `permissions`, `project_id`, null
from `role`;

create table `project__user_new`
(
    `project_id` int not null,
    `user_id`    int not null,
    `role_id`    int not null,

    unique (`project_id`, `user_id`),
    foreign key (`project_id`) references project (`id`) on delete cascade,
    foreign key (`user_id`) references `user` (`id`) on delete cascade,
    foreign key (`role_id`) references `role_new` (`id`) on delete restrict
);

create index `project__user_new__user_id` on `project__user_new` (`user_id`);

create table `project__template_role_new`
(
    `id`          integer primary key autoincrement,
    `template_id` int    not null,
    `role_id`     int    not null,
    `project_id`  int    not null,
    `permissions` bigint not null default 0,

    foreign key (`template_id`) references project__template (`id`) on delete cascade,
    foreign key (`role_id`) references `role_new` (`id`) on delete restrict,
    foreign key (`project_id`) references project (`id`) on delete cascade,
    unique (`template_id`, `role_id`)
);

create table `project__invite_new`
(
    `id`              integer primary key autoincrement,
    `project_id`      int          not null,
    `user_id`         int null,
    `email`           varchar(255) null,
    `role_id`         int          not null,
    `status`          varchar(50)  not null default 'pending',
    `token`           varchar(255) not null,
    `inviter_user_id` int          not null,
    `created`         datetime     not null,
    `expires_at`      datetime null,
    `accepted_at`     datetime null,

    foreign key (`project_id`) references project (`id`) on delete cascade,
    foreign key (`user_id`) references `user` (`id`) on delete cascade,
    foreign key (`role_id`) references `role_new` (`id`) on delete restrict,
    foreign key (`inviter_user_id`) references `user` (`id`) on delete cascade,
    unique (`token`),
    unique (`project_id`, `user_id`),
    unique (`project_id`, `email`)
);

insert into `project__user_new` (`project_id`, `user_id`, `role_id`)
select pu.`project_id`, pu.`user_id`, r.`id`
from `project__user` pu
join `role_new` r on r.`slug` = pu.`role`
where r.`project_id` is null
   or r.`project_id` = pu.`project_id`;

insert into `project__template_role_new` (`id`, `template_id`, `role_id`, `project_id`, `permissions`)
select tr.`id`, tr.`template_id`, r.`id`, tr.`project_id`, tr.`permissions`
from `project__template_role` tr
join `project__template` t on t.`id` = tr.`template_id`
join `role_new` r on r.`slug` = tr.`role_slug`
where tr.`project_id` = t.`project_id`
  and (r.`project_id` is null or r.`project_id` = t.`project_id`);

insert into `project__invite_new` (
    `id`,
    `project_id`,
    `user_id`,
    `email`,
    `role_id`,
    `status`,
    `token`,
    `inviter_user_id`,
    `created`,
    `expires_at`,
    `accepted_at`
)
select
    i.`id`,
    i.`project_id`,
    i.`user_id`,
    i.`email`,
    r.`id`,
    i.`status`,
    i.`token`,
    i.`inviter_user_id`,
    i.`created`,
    i.`expires_at`,
    i.`accepted_at`
from `project__invite` i
join `role_new` r on r.`slug` = i.`role`
where r.`builtin_key` is not null;

{{if .Postgresql}}
select setval(
    pg_get_serial_sequence('project__template_role_new', 'id'),
    coalesce((select max(`id`) from `project__template_role_new`), 1),
    exists(select 1 from `project__template_role_new`)
);

select setval(
    pg_get_serial_sequence('project__invite_new', 'id'),
    coalesce((select max(`id`) from `project__invite_new`), 1),
    exists(select 1 from `project__invite_new`)
);
{{end}}

{{if .Mysql}}
drop index `role_new__slug` on `role_new`;
{{else}}
drop index `role_new__slug`;
{{end}}

alter table `role_new` drop column `slug`;

{{if .Mysql}}
rename table
    `role` to `role_old`,
    `role_new` to `role`,
    `project__user` to `project__user_old`,
    `project__user_new` to `project__user`,
    `project__template_role` to `project__template_role_old`,
    `project__template_role_new` to `project__template_role`,
    `project__invite` to `project__invite_old`,
    `project__invite_new` to `project__invite`;
{{else}}
alter table `role` rename to `role_old`;
alter table `role_new` rename to `role`;
alter table `project__user` rename to `project__user_old`;
alter table `project__user_new` rename to `project__user`;
alter table `project__template_role` rename to `project__template_role_old`;
alter table `project__template_role_new` rename to `project__template_role`;
alter table `project__invite` rename to `project__invite_old`;
alter table `project__invite_new` rename to `project__invite`;
{{end}}

drop table `project__invite_old`;
drop table `project__template_role_old`;
drop table `project__user_old`;
drop table `role_old`;
