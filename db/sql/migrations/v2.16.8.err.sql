alter table `task__stage` add `start_output_id` bigint null references `task__output`(`id`);
alter table `task__stage` add `end_output_id` bigint null references `task__output`(`id`);

{{if .Sqlite}}
create index if not exists task__stage__start_output_id on `task__stage`(`start_output_id`);
create index if not exists task__stage__end_output_id on `task__stage`(`end_output_id`);
{{end}}
-- The MySQL foreign key on stage_id is dropped in migration_2_16_8.PreRollback
-- (auto-generated names differ across DB versions).

alter table `task__output` drop column `stage_id`;
