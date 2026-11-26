alter table `project__view` add `show_template_stats` boolean not null default false;

{{if .Postgresql}}
create index if not exists task__template_id__id on task (template_id, id);
drop index if exists task_template_id_idx;
{{end}}
