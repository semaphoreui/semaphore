-- Global roles get their own table: a role grants project permissions, a global role grants instance-wide
-- permissions, and the two never mix in one row. Roles that carried global permissions are copied here
-- under the same slug, their global assignments follow, and the column leaves the role table.
create table `global_role`
(
    `id`          integer primary key autoincrement,
    `slug`        varchar(100) not null,
    `name`        varchar(100) not null,
    `permissions` bigint not null default 0,

    unique (`slug`)
);

insert into `global_role` (`slug`, `name`, `permissions`)
select `slug`, `name`, `global_permissions`
from `role`
where `project_id` is null and `builtin` = false and `global_permissions` <> 0;

create table `user__global_role_2_20_15`
(
    `id`      integer primary key autoincrement,
    `user_id` int not null,
    `role_id` int not null,

    unique (`user_id`, `role_id`),
    foreign key (`user_id`) references `user` (`id`) on delete cascade,
    foreign key (`role_id`) references `global_role` (`id`) on delete restrict
);

insert into `user__global_role_2_20_15` (`user_id`, `role_id`)
select a.`user_id`, g.`id`
from `user__global_role` as a
join `role` as r on r.`id` = a.`role_id`
join `global_role` as g on g.`slug` = r.`slug`;

drop table `user__global_role`;

alter table `user__global_role_2_20_15` rename to `user__global_role`;

create index `user__global_role__role_id` on `user__global_role` (`role_id`);

-- A role that existed only to carry global permissions has nothing left to do in a project.
delete from `role`
where `project_id` is null and `builtin` = false and `global_permissions` <> 0 and `permissions` = 0
  and not exists (select 1 from `project__user_role` as m where m.`role_id` = `role`.`id`)
  and not exists (select 1 from `project__template_role` as g where g.`role_id` = `role`.`id`);

alter table `role` drop column `global_permissions`;
