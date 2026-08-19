alter table `access_key` add column `source_storage_type` varchar(10);
update `access_key` set source_storage_type = 'vault' where source_storage_id is not null;
update `access_key` set source_storage_type = 'env' where source_storage_id is null and source_storage_key is not null;

{{if .Postgresql}}
create index if not exists task__output_task_id_idx on task__output (task_id);
create index if not exists task_template_id_idx on task (template_id);
create index if not exists task_project_id_idx on task (project_id);
{{end}}
