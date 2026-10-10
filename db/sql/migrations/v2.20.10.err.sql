alter table `project__workflow_edge` drop column `input_mappings`;
alter table `project__workflow_edge` drop column `input_mode`;

{{if .Mysql}}
drop index audit_event_created_idx on audit_event;
{{else}}
drop index if exists audit_event_created_idx;
{{end}}