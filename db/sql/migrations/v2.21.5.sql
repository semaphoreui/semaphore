-- A re-run is a new run that points at the run it repeats. `rerun_mode` is `all` (every node
-- runs again) or `failed` (nodes that succeeded are carried over, see below).
alter table `project__workflow_run` add column `rerun_of_run_id` int null;
alter table `project__workflow_run` add column `rerun_mode` varchar(20) null;

-- One row per node a `failed` re-run does not execute again: the node counts as succeeded in
-- `workflow_run_id`, and its result lives in `origin_run_id`, the run that executed it.
-- `task_id` is the task that ran there (null for an approval or delay node, and once the task
-- is deleted); the row, not the task, is what keeps the node from being launched again.
create table `project__workflow_run_carry` (
  `id` integer primary key autoincrement,
  `project_id` int not null,
  `workflow_run_id` int not null,
  `workflow_node_id` int not null,
  `origin_run_id` int not null,
  `task_id` int null,

  foreign key (`project_id`) references `project`(`id`) on delete cascade,
  foreign key (`workflow_run_id`) references `project__workflow_run`(`id`) on delete cascade,
  foreign key (`workflow_node_id`) references `project__workflow_node`(`id`) on delete cascade,
  foreign key (`task_id`) references `task`(`id`) on delete set null,
  unique (`workflow_run_id`, `workflow_node_id`)
);

create index `project__workflow_run_carry__task_id` on `project__workflow_run_carry`(`task_id`);
