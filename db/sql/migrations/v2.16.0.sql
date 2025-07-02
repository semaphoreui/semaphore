alter table `access_key` add `owner` varchar(20) default '' not null;
alter table `access_key` add `plain` text;
update access_key set `owner` = 'variable' where environment_id is not null and name like 'var.%';
update access_key set `owner` = 'environment' where environment_id is not null and name like 'env.%';

create table secret_storage (
  id integer primary key autoincrement,
  name varchar(100) not null,
  type varchar(20) not null
);

alter table `access_key` add `storage_id` int null references `secret_storage`(`id`);
alter table `access_key` add `linked_storage_id` int null references `secret_storage`(`id`);
alter table `access_key` add `linked_storage_key` varchar(1000);