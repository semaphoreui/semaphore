create table `project__template_key` (
    `project_id` int not null,
    `template_id` int not null,
    `key_id` int not null,
    primary key (`template_id`, `key_id`),
    foreign key (`project_id`) references `project`(`id`) on delete cascade,
    foreign key (`template_id`) references `project__template`(`id`) on delete cascade,
    foreign key (`key_id`) references `access_key`(`id`) on delete cascade
);
