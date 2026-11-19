drop table `project__workflow_run_carry`;
alter table `project__workflow_run` drop column `rerun_mode`;
alter table `project__workflow_run` drop column `rerun_of_run_id`;
