alter table `project__secret_storage` add `sync_enabled` boolean not null default false;

create table `project__secret_storage__sync_path` (
  `id` integer primary key autoincrement,

  storage_id    int             not null,
  path          varchar(1000)   not null default '',
  prefix        varchar(1000)   not null default '',
  `separator`   varchar(20)     not null default '',

  foreign key (`storage_id`) references `project__secret_storage`(`id`) on delete cascade
);
