alter table `project__workflow_revision` add column `survey_vars` text null;
alter table `project__workflow_node` add column `survey_inputs` text null;
alter table `project__workflow_run` add column `environment` text null;

drop table if exists `project__template_role_old`;
drop table if exists `project__user_old`;
drop table if exists `role_old`;
