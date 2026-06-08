PRAGMA foreign_keys=OFF;

create table `project__workflow_node_new` (
  `id` integer primary key autoincrement,
  `workflow_template_id` int not null,
  `template_id` int not null default 0,
  `inventory_id` int null,
  `environment_id` int null,
  `kind` varchar(30) not null default 'task',
  `convergence_mode` varchar(30) not null default 'all',
  `approval_timeout` int null,
  `approval_message` text null,

  foreign key (`workflow_template_id`) references `project__workflow_template`(`id`) on delete cascade,
  foreign key (`inventory_id`) references `project__inventory`(`id`) on delete set null,
  foreign key (`environment_id`) references `project__environment`(`id`) on delete set null
);

insert into `project__workflow_node_new` (
  `id`,
  `workflow_template_id`,
  `template_id`,
  `inventory_id`,
  `environment_id`,
  `kind`,
  `convergence_mode`,
  `approval_timeout`,
  `approval_message`
)
select
  `id`,
  `workflow_template_id`,
  `template_id`,
  `inventory_id`,
  `environment_id`,
  'task',
  'all',
  null,
  null
from `project__workflow_node`;

drop table `project__workflow_node`;
alter table `project__workflow_node_new` rename to `project__workflow_node`;
create index `project__workflow_node__workflow_template_id` on `project__workflow_node`(`workflow_template_id`);
create index `project__workflow_node__template_id` on `project__workflow_node`(`template_id`);

create table `project__workflow_approval` (
  `id` integer primary key autoincrement,
  `project_id` int not null,
  `workflow_run_id` int not null,
  `workflow_node_id` int not null,
  `status` varchar(30) not null,
  `created` datetime not null,
  `resolved` datetime null,
  `resolved_by_user_id` int null,

  foreign key (`project_id`) references `project`(`id`) on delete cascade,
  foreign key (`workflow_run_id`) references `project__workflow_run`(`id`) on delete cascade,
  foreign key (`workflow_node_id`) references `project__workflow_node`(`id`) on delete cascade,
  foreign key (`resolved_by_user_id`) references `user`(`id`) on delete set null
);

create unique index `project__workflow_approval__run_node` on `project__workflow_approval`(`workflow_run_id`, `workflow_node_id`);
create index `project__workflow_approval__project_id` on `project__workflow_approval`(`project_id`);
create index `project__workflow_approval__run_id` on `project__workflow_approval`(`workflow_run_id`);

PRAGMA foreign_keys=ON;
