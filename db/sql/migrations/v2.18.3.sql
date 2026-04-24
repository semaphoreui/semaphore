create table `project__template_environment` (
    `project_id` int not null,
    `template_id` int not null,
    `environment_id` int not null,
    primary key (`template_id`, `environment_id`),
    foreign key (`project_id`) references `project`(`id`) on delete cascade,
    foreign key (`template_id`) references `project__template`(`id`) on delete cascade,
    foreign key (`environment_id`) references `project__environment`(`id`) on delete cascade
);
