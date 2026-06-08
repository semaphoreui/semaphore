create table `project__workflow_template` (
  `id` integer primary key autoincrement,
  `project_id` int not null,
  `name` varchar(255) not null,
  `description` text null,

  foreign key (`project_id`) references `project`(`id`) on delete cascade
);

create index `project__workflow_template__project_id` on `project__workflow_template`(`project_id`);

create table `project__workflow_node` (
  `id` integer primary key autoincrement,
  `workflow_template_id` int not null,
  `template_id` int not null,
  `inventory_id` int null,
  `environment_id` int null,

  foreign key (`workflow_template_id`) references `project__workflow_template`(`id`) on delete cascade,
  foreign key (`template_id`) references `project__template`(`id`) on delete cascade,
  foreign key (`inventory_id`) references `project__inventory`(`id`) on delete set null,
  foreign key (`environment_id`) references `project__environment`(`id`) on delete set null
);

create index `project__workflow_node__workflow_template_id` on `project__workflow_node`(`workflow_template_id`);
create index `project__workflow_node__template_id` on `project__workflow_node`(`template_id`);

create table `project__workflow_edge` (
  `id` integer primary key autoincrement,
  `workflow_template_id` int not null,
  `source_node_id` int not null,
  `destination_node_id` int not null,
  `condition` varchar(30) not null,

  foreign key (`workflow_template_id`) references `project__workflow_template`(`id`) on delete cascade,
  foreign key (`source_node_id`) references `project__workflow_node`(`id`) on delete cascade,
  foreign key (`destination_node_id`) references `project__workflow_node`(`id`) on delete cascade
);

create index `project__workflow_edge__workflow_template_id` on `project__workflow_edge`(`workflow_template_id`);
create index `project__workflow_edge__source_node_id` on `project__workflow_edge`(`source_node_id`);
create index `project__workflow_edge__destination_node_id` on `project__workflow_edge`(`destination_node_id`);

create table `project__workflow_run` (
  `id` integer primary key autoincrement,
  `project_id` int not null,
  `workflow_template_id` int not null,
  `status` varchar(30) not null,
  `start` datetime null,
  `end` datetime null,
  `root_task_id` int null,

  foreign key (`project_id`) references `project`(`id`) on delete cascade,
  foreign key (`workflow_template_id`) references `project__workflow_template`(`id`) on delete cascade,
  foreign key (`root_task_id`) references `task`(`id`) on delete set null
);

create index `project__workflow_run__project_id` on `project__workflow_run`(`project_id`);
create index `project__workflow_run__workflow_template_id` on `project__workflow_run`(`workflow_template_id`);

alter table `task` add column `workflow_run_id` int null;
alter table `task` add column `workflow_node_id` int null;

create index `task__workflow_run_id` on `task`(`workflow_run_id`);
create index `task__workflow_node_id` on `task`(`workflow_node_id`);
