{{if .Postgresql}}
drop index if exists task__output_task_id_idx;
drop index if exists task_template_id_idx;
drop index if exists task_project_id_idx;
{{end}}

alter table access_key drop column `source_storage_type`;