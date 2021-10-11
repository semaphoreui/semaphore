alter table `project__template` add `type` varchar(10) not null default 'task';
alter table `project__template` add `start_version` varchar(20) not null default '0.0.0';
alter table `task` add `version` varchar(20);