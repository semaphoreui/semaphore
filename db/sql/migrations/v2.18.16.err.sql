PRAGMA foreign_keys=OFF;

drop index if exists `project__workflow_approval__run_id`;
drop index if exists `project__workflow_approval__project_id`;
drop index if exists `project__workflow_approval__run_node`;
drop table if exists `project__workflow_approval`;

create table `project__workflow_node_old` (
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

insert into `project__workflow_node_old` (`id`, `workflow_template_id`, `template_id`, `inventory_id`, `environment_id`)
select `id`, `workflow_template_id`, `template_id`, `inventory_id`, `environment_id`
from `project__workflow_node`;

drop table `project__workflow_node`;
alter table `project__workflow_node_old` rename to `project__workflow_node`;
create index `project__workflow_node__workflow_template_id` on `project__workflow_node`(`workflow_template_id`);
create index `project__workflow_node__template_id` on `project__workflow_node`(`template_id`);

PRAGMA foreign_keys=ON;
