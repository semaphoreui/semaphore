alter table `task__output` add `stage_id` int null references `task__stage`(`id`);

{{if .Sqlite}}
drop index if exists task__stage__start_output_id;
drop index if exists task__stage__end_output_id;
{{else if .Mysql}}
alter table `task__stage` drop foreign key if exists `task__stage_ibfk_2`;
alter table `task__stage` drop foreign key if exists `task__stage_ibfk_3`;
{{end}}

alter table `task__stage` drop column `start_output_id`;
alter table `task__stage` drop column `end_output_id`;