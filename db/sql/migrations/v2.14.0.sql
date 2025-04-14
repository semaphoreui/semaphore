alter table project__template add allow_override_branch_in_task bool not null default false;
alter table `task` drop column `diff`;
alter table `task` drop column `debug`;
alter table `task` drop column `dry_run`;
alter table `task` drop column `hosts_limit`;
alter table runner add touched datetime;
alter table runner add cleaning_requested datetime;
