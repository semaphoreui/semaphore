drop index if exists `task__workflow_node_id`;
drop index if exists `task__workflow_run_id`;

alter table `task` drop column `artifacts`;
alter table `task` drop column `workflow_node_id`;
alter table `task` drop column `workflow_run_id`;

drop index if exists `project__workflow_approval__run_id`;
drop index if exists `project__workflow_approval__project_id`;
drop index if exists `project__workflow_approval__run_node`;
drop table if exists `project__workflow_approval`;

drop index if exists `project__workflow_run__workflow_template_id`;
drop index if exists `project__workflow_run__project_id`;
drop table if exists `project__workflow_run`;

drop index if exists `project__workflow_edge__destination_node_id`;
drop index if exists `project__workflow_edge__source_node_id`;
drop index if exists `project__workflow_edge__workflow_template_id`;
drop table if exists `project__workflow_edge`;

drop index if exists `project__workflow_node__template_id`;
drop index if exists `project__workflow_node__workflow_template_id`;
drop table if exists `project__workflow_node`;

drop index if exists `project__workflow_template__project_id`;
drop table if exists `project__workflow_template`;
