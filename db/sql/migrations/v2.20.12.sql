create table `role_new`
(
    `id`          integer primary key autoincrement,
    `slug`        varchar(100) not null,
    `name`        varchar(100) not null,
    `permissions` bigint       not null default 0,
    `project_id`  int,
    `scope`       int          not null default 0,
    `builtin`     boolean      not null default false,

    foreign key (`project_id`) references project (`id`) on delete cascade,
    check (`scope` = coalesce(`project_id`, 0)),
    check (`builtin` = false or `project_id` is null)
);

insert into `role_new` (`slug`, `name`, `permissions`, `project_id`, `scope`, `builtin`)
values ('owner', 'Owner', 15, null, 0, true),
       ('manager', 'Manager', 5, null, 0, true),
       ('task_runner', 'Task Runner', 1, null, 0, true),
       ('guest', 'Guest', 0, null, 0, true);

insert into `role_new` (`slug`, `name`, `permissions`, `project_id`, `scope`, `builtin`)
select `slug`, `name`, `permissions`, `project_id`, coalesce(`project_id`, 0), false
from `role`;

create table `project__user_role`
(
    `project_id` int not null,
    `user_id`    int not null,
    `role_id`    int not null,

    unique (`project_id`, `user_id`, `role_id`),
    foreign key (`project_id`) references project (`id`) on delete cascade,
    foreign key (`user_id`) references `user` (`id`) on delete cascade,
    foreign key (`role_id`) references `role_new` (`id`) on delete restrict
);

create index `project__user_role__user_id` on `project__user_role` (`user_id`);
create index `project__user_role__role_id` on `project__user_role` (`role_id`);

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

-- An exact built-in slug always meant the built-in role, a custom role named like it never applied to members.
insert into `project__user_role` (`project_id`, `user_id`, `role_id`)
select pu.`project_id`, pu.`user_id`, r.`id`
from `project__user` pu
join `role_new` r on r.`slug` = pu.`role`
    and r.`builtin` = ({{if .Mysql}}binary {{end}}pu.`role` in ('owner', 'manager', 'task_runner', 'guest'))
    and (r.`project_id` is null or r.`project_id` = pu.`project_id`);

-- A member whose role does not exist had no permissions, the built-in guest role keeps exactly that.
insert into `project__user_role` (`project_id`, `user_id`, `role_id`)
select pu.`project_id`, pu.`user_id`, r.`id`
from `project__user` pu
join `role_new` r on r.`builtin` = true and r.`slug` = 'guest'
where not exists (
    select 1 from `project__user_role` pur where pur.`project_id` = pu.`project_id` and pur.`user_id` = pu.`user_id`
);

insert into `project__template_role_new` (`id`, `template_id`, `role_id`, `project_id`, `permissions`)
select tr.`id`, tr.`template_id`, r.`id`, tr.`project_id`, tr.`permissions`
from `project__template_role` tr
join `project__template` t on t.`id` = tr.`template_id` and t.`project_id` = tr.`project_id`
join `role_new` r on r.`slug` = tr.`role_slug`
    and r.`builtin` = ({{if .Mysql}}binary {{end}}tr.`role_slug` in ('owner', 'manager', 'task_runner', 'guest'))
    and (r.`project_id` is null or r.`project_id` = t.`project_id`);

{{if .Postgresql}}
select setval(
    pg_get_serial_sequence('project__template_role_new', 'id'),
    coalesce((select max(`id`) from `project__template_role_new`), 1),
    exists(select 1 from `project__template_role_new`)
);
{{end}}

{{if .Mysql}}
rename table
    `role` to `role_old`,
    `role_new` to `role`,
    `project__user` to `project__user_old`,
    `project__template_role` to `project__template_role_old`,
    `project__template_role_new` to `project__template_role`;
{{else}}
alter table `role` rename to `role_old`;
alter table `role_new` rename to `role`;
alter table `project__user` rename to `project__user_old`;
alter table `project__template_role` rename to `project__template_role_old`;
alter table `project__template_role_new` rename to `project__template_role`;
{{end}}
