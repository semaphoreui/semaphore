alter table `project` add column `max_parallel_tasks` int not null default 0;
alter table `project__template` add column `suppress_success_alerts` bool not null default false;
