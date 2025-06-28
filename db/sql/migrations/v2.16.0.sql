alter table `access_key` add `type` varchar(20);
alter table `access_key` add `plain` text;
update access_key set `type` = 'variable' where environment_id is not null and name like 'var.%';
update access_key set `type` = 'environment' where environment_id is not null and name like 'env.%';
