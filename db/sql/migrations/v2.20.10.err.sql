{{if .Mysql}}
-- The access_key.user_id foreign key added by migration_2_20_10.PostApply is
-- intentionally kept: MariaDB always had it, and the pre-2.20.10 schema is
-- valid with or without it.
alter table `task` drop foreign key `task__user_fk`;
alter table `session` drop foreign key `session__user_fk`;
{{end}}
{{if .Postgresql}}
alter table `task` drop foreign key `task__user_fk`;
alter table `session` drop foreign key `session__user_fk`;
{{end}}
