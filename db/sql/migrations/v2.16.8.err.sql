alter table `task__stage` add `start_output_id` bigint null references `task__output`(`id`);
alter table `task__stage` add `end_output_id` bigint null references `task__output`(`id`);

{{if .Sqlite}}
create index if not exists task__stage__start_output_id on `task__stage`(`start_output_id`);
create index if not exists task__stage__end_output_id on `task__stage`(`end_output_id`);
{{else if .Mysql}}
alter table `task__output` drop foreign key if exists `task__output_ibfk_2`;
{{end}}

alter table `task__output` drop column `stage_id`;
