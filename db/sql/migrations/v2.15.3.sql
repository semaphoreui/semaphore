create table task__ansible_error(
    `id` integer primary key autoincrement,
    `task_id` int NOT NULL,
    `project_id` int NOT NULL,
    `task` varchar(250) NOT NULL,
    `error` varchar(1000) NOT NULL,
    foreign key (`project_id`) references `project`(`id`) on delete cascade,
    foreign key (`task_id`) references `task`(`id`) on delete cascade
);
create table task__ansible_host(
   `id` integer primary key autoincrement,
   `task_id` int NOT NULL,
   `project_id` int NOT NULL,
   `host` varchar(250) NOT NULL,
   `failed` int NOT NULL,
   `ignored` int NOT NULL,
   `changed` int NOT NULL,
   `ok` int NOT NULL,
   `rescued` int NOT NULL,
   `skipped` int NOT NULL,
   `unreachable` int NOT NULL,
   foreign key (`project_id`) references `project`(`id`) on delete cascade,
   foreign key (`task_id`) references `task`(`id`) on delete cascade
);