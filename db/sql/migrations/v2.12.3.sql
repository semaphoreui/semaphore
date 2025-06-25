alter table `task__output` drop `task`;

alter table `task__output` change `id` `id` bigint autoincrement not null