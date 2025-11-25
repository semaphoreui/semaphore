-- Workflow feature migration
-- Creates tables for workflows, workflow nodes, workflow links, workflow runs, and workflow run nodes

create table `project__workflow` (
  `id` integer primary key autoincrement,
  `project_id` int not null,
  `name` varchar(255) not null,
  `description` text null,
  `created` datetime not null,
  `updated` datetime not null,

  foreign key (`project_id`) references project(`id`) on delete cascade
);

create table `project__workflow_node` (
  `id` integer primary key autoincrement,
  `workflow_id` int not null,
  `task_id` int null,
  `type` varchar(20) not null default 'task',
  `position_x` double not null default 0,
  `position_y` double not null default 0,
  `config_json` text null,

  foreign key (`workflow_id`) references project__workflow(`id`) on delete cascade,
  foreign key (`task_id`) references project__template(`id`) on delete set null
);

create table `project__workflow_link` (
  `id` integer primary key autoincrement,
  `workflow_id` int not null,
  `from_node_id` int not null,
  `to_node_id` int not null,
  `condition` varchar(20) not null default 'success',

  foreign key (`workflow_id`) references project__workflow(`id`) on delete cascade,
  foreign key (`from_node_id`) references project__workflow_node(`id`) on delete cascade,
  foreign key (`to_node_id`) references project__workflow_node(`id`) on delete cascade
);

create table `project__workflow_run` (
  `id` integer primary key autoincrement,
  `workflow_id` int not null,
  `status` varchar(20) not null default 'pending',
  `user_id` int null,
  `created` datetime not null,
  `start` datetime null,
  `end` datetime null,
  `message` text null,

  foreign key (`workflow_id`) references project__workflow(`id`) on delete cascade,
  foreign key (`user_id`) references `user`(`id`) on delete set null
);

create table `project__workflow_run_node` (
  `id` integer primary key autoincrement,
  `workflow_run_id` int not null,
  `workflow_node_id` int not null,
  `task_id` int null,
  `status` varchar(20) not null default 'pending',
  `created` datetime not null,
  `start` datetime null,
  `end` datetime null,
  `message` text null,

  foreign key (`workflow_run_id`) references project__workflow_run(`id`) on delete cascade,
  foreign key (`workflow_node_id`) references project__workflow_node(`id`) on delete cascade,
  foreign key (`task_id`) references `task`(`id`) on delete set null
);
