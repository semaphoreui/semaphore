alter table `access_key` add `owner` varchar(20) default '' not null;
alter table `access_key` add `plain` text;
update access_key set `owner` = 'variable' where environment_id is not null and name like 'var.%';
update access_key set `owner` = 'environment' where environment_id is not null and name like 'env.%';

create table project__secret_storage (
  id integer primary key autoincrement,

  project_id    int             not null,
  name          varchar(100)    not null,
  type          varchar(20)     not null,
  params        text,
  readonly      boolean         not null default false,

  foreign key (`project_id`) references project(`id`) on delete cascade
);

alter table `access_key` add `storage_id` int null references `project__secret_storage`(`id`);

alter table `access_key` add `source_storage_id` int null references `project__secret_storage`(`id`);
alter table `access_key` add `source_storage_key` varchar(1000);