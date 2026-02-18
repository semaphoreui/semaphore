alter table `project__schedule` add column `run_at` datetime null;
alter table `project__schedule` add column `type` varchar(20) not null default '';
alter table `project__schedule` add column `delete_after_run` boolean not null default false;