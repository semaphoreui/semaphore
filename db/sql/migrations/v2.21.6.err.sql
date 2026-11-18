alter table `project__view` drop column `show_template_stats`;

{{if .Postgresql}}
create index if not exists task_template_id_idx on task (template_id);
drop index if exists task__template_id__id;
{{end}}
