alter table `task__output` add `stage_id` int null references `task__stage`(`id`);

{{if .Sqlite}}
drop index if exists task__stage__start_output_id;
drop index if exists task__stage__end_output_id;
{{end}}
-- The MySQL foreign keys on start_output_id / end_output_id are dropped in
-- migration_2_16_8.PreApply (auto-generated names differ across DB versions).

alter table `task__stage` drop column `start_output_id`;
alter table `task__stage` drop column `end_output_id`;