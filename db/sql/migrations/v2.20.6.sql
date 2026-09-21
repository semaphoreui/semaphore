create table `project__alert` (
    `id` integer primary key autoincrement,
    `project_id` int not null,
    `name` varchar(255) not null,
    `type` varchar(20) not null,
    `enabled` boolean not null default true,
    `is_default` boolean not null default false,
    `chat_id` varchar(100),
    `thread_id` varchar(50),
    `url` text,
    `token` varchar(255),
    `recipients` text,
    `key_id` int,
    `body` longtext,
    unique (`project_id`, `name`),
    foreign key (`project_id`) references `project`(`id`) on delete cascade,
    foreign key (`key_id`) references `access_key`(`id`) on delete set null
);

create table `project__template_alert` (
    `project_id` int not null,
    `template_id` int not null,
    `alert_id` int not null,
    primary key (`template_id`, `alert_id`),
    foreign key (`project_id`) references `project`(`id`) on delete cascade,
    foreign key (`template_id`) references `project__template`(`id`) on delete cascade,
    foreign key (`alert_id`) references `project__alert`(`id`) on delete restrict
);

create table `project__schedule_alert` (
    `project_id` int not null,
    `schedule_id` int not null,
    `alert_id` int not null,
    primary key (`schedule_id`, `alert_id`),
    foreign key (`project_id`) references `project`(`id`) on delete cascade,
    foreign key (`schedule_id`) references `project__schedule`(`id`) on delete cascade,
    foreign key (`alert_id`) references `project__alert`(`id`) on delete restrict
);

alter table `project__template` add column `alert_on_success` boolean not null default true;
alter table `project__template` add column `alert_on_error` boolean not null default true;
alter table `project__template` add column `alert_mode` varchar(20) not null default 'default';

update `project__template` set `alert_on_success` = case when `suppress_success_alerts` then false else true end;
update `project__template` set `alert_on_error` = case when `suppress_error_alerts` then false else true end;

alter table `project__schedule` add column `alert_mode` varchar(20) not null default 'inherit';
alter table `project__schedule` add column `alert_on_success` boolean null;
alter table `project__schedule` add column `alert_on_error` boolean null;

alter table `task` add column `alert_snapshot` longtext;

create table `task__alert_send` (
    `task_id` int not null,
    `alert_id` int not null,
    `event` varchar(20) not null,
    `created` datetime not null,
    primary key (`task_id`, `alert_id`, `event`),
    foreign key (`task_id`) references `task`(`id`) on delete cascade
);
