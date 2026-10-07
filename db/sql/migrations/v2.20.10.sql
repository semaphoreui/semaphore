{{if .Mysql}}
-- Restores the foreign keys on `user_id` that MySQL/MariaDB never received:
--   * task.user_id    — added in v1.7.0 without a foreign key
--   * session.user_id — created in v1.5.0 / rebuilt in v2.3.1 without one
--   * access_key.user_id — v2.10.15 used an inline `references` clause, which
--     MySQL parses but silently ignores (MariaDB honours it). It is added by
--     migration_2_20_10.PostApply only when missing.
-- SQLite got all of these in v2.19.14 (its own table rebuild), so it is skipped.
-- Orphaned rows are cleaned up first so the constraints can be validated.
delete from `session` where `user_id` not in (select `id` from `user`);
update `task` set `user_id` = null where `user_id` is not null and `user_id` not in (select `id` from `user`);
delete from `access_key` where `user_id` is not null and `user_id` not in (select `id` from `user`);

alter table `session` add constraint `session__user_fk` foreign key (`user_id`) references `user`(`id`) on delete cascade;
alter table `task` add constraint `task__user_fk` foreign key (`user_id`) references `user`(`id`) on delete set null;
{{end}}
{{if .Postgresql}}
-- Restores the foreign keys on `user_id` that PostgreSQL never received:
--   * task.user_id    — added in v1.7.0 without a foreign key
--   * session.user_id — created in v1.5.0 / rebuilt in v2.3.1 without one.
-- access_key.user_id already has one (PostgreSQL honours the inline
-- `references` of v2.10.15). SQLite got all of these in v2.19.14.
-- Orphaned rows are cleaned up first so the constraints can be validated.
delete from `session` where `user_id` not in (select `id` from `user`);
update `task` set `user_id` = null where `user_id` is not null and `user_id` not in (select `id` from `user`);

alter table `session` add constraint `session__user_fk` foreign key (`user_id`) references `user`(`id`) on delete cascade;
alter table `task` add constraint `task__user_fk` foreign key (`user_id`) references `user`(`id`) on delete set null;
{{end}}
