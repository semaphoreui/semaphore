create table `project__alert` (
  `id` integer primary key autoincrement,
  `project_id` int not null,
  `name` varchar(255) not null,
  `type` varchar(30) not null,
  `enabled` boolean not null default true,
  `is_default` boolean not null default false,
  `events` varchar(255) null,
  `chat_id` varchar(100) null,
  `thread_id` varchar(50) null,
  `url` text null,
  `recipients` text null,
  `key_id` int null,
  `params` longtext null,
  `body` longtext null,

  unique (`project_id`, `name`),
  foreign key (`project_id`) references `project`(`id`) on delete cascade,
  foreign key (`key_id`) references `access_key`(`id`)
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

alter table `project__template` add column `alert_mode` varchar(20) not null default 'default';
alter table `project__schedule` add column `alert_mode` varchar(20) not null default 'inherit';

alter table `task` add column `alert_snapshot` longtext null;

create table `task__alert_send` (
  `task_id` int not null,
  `destination` varchar(64) not null,
  `event` varchar(30) not null,
  `created` datetime not null,

  primary key (`task_id`, `destination`, `event`),
  foreign key (`task_id`) references `task`(`id`) on delete cascade
);

drop table if exists `event_backup_5784568`;
