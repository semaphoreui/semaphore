create table task__stage
(
    `id`              integer primary key autoincrement,
    `task_id`         int      NOT NULL,
    `start`           datetime null,
    `start_output_id` bigint   null,
    `end`             datetime null,
    `end_output_id`   bigint   null,
    `type`            varchar(100),
    foreign key (`task_id`) references `task` (`id`) on delete cascade,
    foreign key (`start_output_id`) references `task__output` (`id`) on delete set null,
    foreign key (`end_output_id`) references `task__output` (`id`) on delete set null
);

create table task__stage_result
(
    `id`       integer primary key autoincrement,
    `task_id`  int NOT NULL,
    `stage_id` int NOT NULL,
    `json`     text,
    foreign key (`task_id`) references `task` (`id`) on delete cascade,
    foreign key (`stage_id`) references `task__stage` (`id`) on delete cascade
);
