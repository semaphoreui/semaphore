alter table `project__template` add `tasks` int not null default 0;
alter table `project__schedule` add `name` varchar(100);
alter table `project__schedule` add `disabled` boolean not null default false;