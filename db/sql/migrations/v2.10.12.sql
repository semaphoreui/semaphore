alter table `project__template` add `tasks` int not null default 0;
alter table `project__schedule` add `name` varchar(100) not null default '';
alter table `project__schedule` add `active` boolean not null default true;