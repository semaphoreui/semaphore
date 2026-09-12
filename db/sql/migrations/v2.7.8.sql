-- The MySQL foreign keys on `key_id` / `ssh_key_id` are dropped in
-- migration_2_7_8.PreApply (the constraint name is auto-generated and differs
-- across DB versions). On Postgres dropping the column drops its FK as well.
alter table `project__inventory` drop column `key_id`;

alter table `project__template` drop column `ssh_key_id`;