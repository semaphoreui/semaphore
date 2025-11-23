alter table `project__schedule` add column if not exists `run_at` datetime null;
alter table `project__schedule` add column if not exists `type` varchar(20) not null default 'cron';
update project__schedule set `type`='run_at' where `run_at` is not null and (`type`='' or `type` is null);
