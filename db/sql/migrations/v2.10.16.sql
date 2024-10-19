update `project__template` set `app` = 'ansible' where `app` = '';

alter table `project__template` change `app` `app` varchar(50) not null;
