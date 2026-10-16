create table `project__proxy` (
    `id` integer primary key autoincrement,
    `project_id` int not null,
    `name` varchar(255) not null,
    `type` varchar(20) not null,
    `host` varchar(255) not null,
    `port` int,
    `user` varchar(255),
    `ssh_key_id` int,
    unique (`project_id`, `name`),

    foreign key (`project_id`) references `project`(`id`) on delete cascade,
    foreign key (`ssh_key_id`) references `access_key`(`id`) on delete set null
);

alter table `project__inventory` add `proxy_id` int references `project__proxy`(`id`) on delete set null;
