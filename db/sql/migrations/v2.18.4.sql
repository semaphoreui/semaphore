create table `project__template_environment` (
    `project_id` int not null,
    `template_id` int not null,
    `environment_id` int not null,
    primary key (`template_id`, `environment_id`),
    foreign key (`project_id`) references `project`(`id`) on delete cascade,
    foreign key (`template_id`) references `project__template`(`id`) on delete cascade,
    foreign key (`environment_id`) references `project__environment`(`id`) on delete cascade
);

insert into `project__template_environment` (`project_id`, `template_id`, `environment_id`)
select `project_id`, `id` as template_id, `environment_id`
from `project__template` where `environment_id` is not null;

alter table `project__template` drop column `environment_id`;
