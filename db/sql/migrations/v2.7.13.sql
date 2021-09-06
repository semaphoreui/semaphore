drop table project__template_schedule;

create table `project__schedule`
(
    `id` integer primary key autoincrement,
    `template_id` int references project__template (`id`) on delete cascade,
    `project_id` int not null references project (`id`) on delete cascade,
    `cron_format` varchar(255) not null
);
