alter table `role` add column `global_permissions` bigint not null default 0;

-- One row, updated at the start of every transaction that changes global roles, their assignments or the
-- admin flag of a user: the row lock serializes those transactions across HA nodes, so the "last
-- administrator" checks cannot race.
create table `global_role_state`
(
    `id`       int primary key,
    `revision` int not null default 1
);

insert into `global_role_state` (`id`, `revision`) values (1, 1);

create table `user__global_role`
(
    `id`      integer primary key autoincrement,
    `user_id` int not null,
    `role_id` int not null,

    unique (`user_id`, `role_id`),
    foreign key (`user_id`) references `user` (`id`) on delete cascade,
    foreign key (`role_id`) references `role` (`id`) on delete restrict
);

create index `user__global_role__role_id` on `user__global_role` (`role_id`);
