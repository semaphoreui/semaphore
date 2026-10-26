-- PreRollback has checked that the new tables exist, so these are leftovers of a failed rollback on MySQL.
drop table if exists `project__template_role_legacy`;
drop table if exists `project__user`;
drop table if exists `role_legacy`;

create table `role_legacy`
(
    `slug`        varchar(100) primary key not null,
    `name`        varchar(100)             not null,
    `permissions` bigint                   not null default 0,
    `project_id`  int,
    foreign key (`project_id`) references project (`id`) on delete cascade
);

insert into `role_legacy` (`slug`, `name`, `permissions`, `project_id`)
select `slug`, `name`, `permissions`, `project_id`
from `role`
where `builtin` = false;

create table `project__user`
(
    `project_id` int          not null,
    `user_id`    int          not null,
    `role`       varchar(100) not null,

    unique (`project_id`, `user_id`),
    foreign key (`project_id`) references project (`id`) on delete cascade,
    foreign key (`user_id`) references `user` (`id`) on delete cascade
);

-- A member keeps the role with the most permissions, a built-in role wins a tie, then the role created first.
insert into `project__user` (`project_id`, `user_id`, `role`)
select pur.`project_id`, pur.`user_id`, r.`slug`
from `project__user_role` pur
join `role` r on r.`id` = pur.`role_id`
where r.`id` = (
    select r2.`id`
    from `project__user_role` pur2
    join `role` r2 on r2.`id` = pur2.`role_id`
    where pur2.`project_id` = pur.`project_id` and pur2.`user_id` = pur.`user_id`
    order by r2.`permissions` desc, r2.`builtin` desc, r2.`id`
    limit 1
);

create table `project__template_role_legacy`
(
    `id`          integer primary key autoincrement,
    `template_id` int          not null,
    `role_slug`   varchar(100) not null,
    `project_id`  int          not null,
    `permissions` bigint       not null default 0,

    foreign key (`template_id`) references project__template (`id`) on delete cascade,
    foreign key (`role_slug`) references `role_legacy` (`slug`) on delete cascade,
    foreign key (`project_id`) references project (`id`) on delete cascade,

    unique (`template_id`, `role_slug`)
);

insert into `project__template_role_legacy` (`id`, `template_id`, `role_slug`, `project_id`, `permissions`)
select tr.`id`, tr.`template_id`, r.`slug`, tr.`project_id`, tr.`permissions`
from `project__template_role` tr
join `role` r on r.`id` = tr.`role_id` and r.`builtin` = false;

{{if .Postgresql}}
select setval(
    pg_get_serial_sequence('project__template_role_legacy', 'id'),
    coalesce((select max(`id`) from `project__template_role_legacy`), 1),
    exists(select 1 from `project__template_role_legacy`)
);
{{end}}

-- The new tables leave under the upgrade's scratch names, so whatever a failure leaves behind, the upgrade drops first.
{{if .Mysql}}
rename table
    `role` to `role_new`,
    `role_legacy` to `role`,
    `project__template_role` to `project__template_role_new`,
    `project__template_role_legacy` to `project__template_role`,
    `project__user_role` to `project__user_role_new`;
{{else}}
alter table `role` rename to `role_new`;
alter table `role_legacy` rename to `role`;
alter table `project__template_role` rename to `project__template_role_new`;
alter table `project__template_role_legacy` rename to `project__template_role`;
alter table `project__user_role` rename to `project__user_role_new`;
{{end}}

drop table `project__user_role_new`;
drop table `project__template_role_new`;
drop table `role_new`;
