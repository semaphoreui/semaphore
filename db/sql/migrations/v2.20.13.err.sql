alter table `project__workflow_run` drop column `environment`;
alter table `project__workflow_node` drop column `survey_inputs`;
alter table `project__workflow_revision` drop column `survey_vars`;
