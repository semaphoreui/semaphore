create table `project__repository_submodule_credential` (
  `id` integer primary key autoincrement,
  `project_id` int not null,
  `repository_id` int not null,
  `access_key_id` int not null,
  `host` varchar(300) not null,

  foreign key (`project_id`) references `project`(`id`) on delete cascade,
  foreign key (`repository_id`) references `project__repository`(`id`) on delete cascade,
  foreign key (`access_key_id`) references `access_key`(`id`) on delete cascade
);

create unique index `project__repository_submodule_credential__repo_host`
  on `project__repository_submodule_credential`(`repository_id`, `host`);

alter table `project__workflow_node` add `delay_seconds` int null;

create table `project__workflow_delay` (
  `id` integer primary key autoincrement,
  `project_id` int not null,
  `workflow_run_id` int not null,
  `workflow_node_id` int not null,
  `status` varchar(30) not null,
  `resume_at` datetime not null,
  `created` datetime not null,
  `resolved` datetime null,

  foreign key (`project_id`) references `project`(`id`) on delete cascade,
  foreign key (`workflow_run_id`) references `project__workflow_run`(`id`) on delete cascade,
  foreign key (`workflow_node_id`) references `project__workflow_node`(`id`) on delete cascade,
  unique (`workflow_run_id`, `workflow_node_id`)
);

create index `project__workflow_delay__workflow_run_id` on `project__workflow_delay`(`workflow_run_id`);
create index `project__workflow_delay__status_resume_at` on `project__workflow_delay`(`status`, `resume_at`);
