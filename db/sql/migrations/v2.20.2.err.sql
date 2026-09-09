drop index `project__workflow_delay__status_resume_at`;
drop index `project__workflow_delay__workflow_run_id`;
drop table `project__workflow_delay`;
alter table `project__workflow_node` drop column `delay_seconds`;