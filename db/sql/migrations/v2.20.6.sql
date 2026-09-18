create table `project__host_config` (
    `id` integer primary key autoincrement,
    `project_id` int not null,
    `type` varchar(10) not null,
    `name` varchar(255) not null,
    `ssh_key_id` int not null,
    unique (`project_id`, `type`, `name`),

    foreign key (`project_id`) references `project`(`id`) on delete cascade,
    foreign key (`ssh_key_id`) references `access_key`(`id`)
);
