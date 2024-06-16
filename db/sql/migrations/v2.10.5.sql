create table task__stage(
    `id` integer primary key autoincrement,
    `task_id` int not null,
    `start` datetime not null,
    `start_output_id` int,
    `end` datetime,
    `end_output_id` int,
    `type` varchar(20) not null,
    foreign key (`task_id`) references project(`id`),
    foreign key (`start_output_id`) references task__output(`id`),
    foreign key (`end_output_id`) references task__output(`id`) on delete cascade
);