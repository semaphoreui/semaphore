alter table `project__secret_storage` add `sync_enabled` boolean not null default false;
alter table `project__secret_storage` add `sync_interval` int not null default 0;
alter table `project__secret_storage` add `last_synced_at` datetime null;
alter table `project__secret_storage` add `last_sync_failed_at` datetime null;

create table `project__secret_storage__sync_path` (
  `id` integer primary key autoincrement,

  storage_id    int             not null,
  path          varchar(1000)   not null default '',
  prefix        varchar(1000)   not null default '',
  `separator`   varchar(20)     not null default '',
  foreign key (`storage_id`) references `project__secret_storage`(`id`) on delete cascade
);

alter table `project__environment` add `sync_enabled` boolean not null default false;
alter table `project__environment` add `sync_interval` int not null default 0;
alter table `project__environment` add `last_synced_at` datetime null;
alter table `project__environment` add `last_sync_failed_at` datetime null;

create table `project__environment__sync_path` (
  `id` integer primary key autoincrement,

  environment_id int             not null,
  path           varchar(1000)   not null default '',
  prefix         varchar(1000)   not null default '',
  `separator`    varchar(20)     not null default '',
  foreign key (`environment_id`) references `project__environment`(`id`) on delete cascade
);
