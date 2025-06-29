alter table `access_key` add `owner` varchar(20);
alter table `access_key` add `plain` text;
update access_key set `owner` = 'variable' where environment_id is not null and name like 'var.%';
update access_key set `owner` = 'environment' where environment_id is not null and name like 'env.%';
