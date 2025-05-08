alter table project__template drop column allow_override_branch_in_task;
ALTER TABLE task ADD diff boolean NOT NULL DEFAULT false;
alter table task add `debug` boolean not null default false;
ALTER TABLE task ADD dry_run boolean NOT NULL DEFAULT false;
alter table `task` add column `hosts_limit` varchar(255) not null default '';

alter table runner drop column touched;
alter table runner drop column cleaning_requested;
