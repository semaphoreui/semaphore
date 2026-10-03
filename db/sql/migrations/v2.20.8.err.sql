drop index `project__workflow_run__revision_id`;
drop index `project__workflow_edge__revision_id`;
drop index `project__workflow_node__revision_id`;

{{if .Sqlite}}
alter table `project__workflow_run` drop column `revision_id`;
alter table `project__workflow_edge` drop column `revision_id`;
alter table `project__workflow_node` drop column `revision_id`;
{{else}}
alter table `project__workflow_run` drop foreign key `project__workflow_run__revision_fk`;
alter table `project__workflow_run` drop column `revision_id`;
alter table `project__workflow_edge` drop foreign key `project__workflow_edge__revision_fk`;
alter table `project__workflow_edge` drop column `revision_id`;
alter table `project__workflow_node` drop foreign key `project__workflow_node__revision_fk`;
alter table `project__workflow_node` drop column `revision_id`;
{{end}}

drop table `project__workflow_revision`;
