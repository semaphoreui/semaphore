create table `project__secret_sync` (
  `id` integer primary key autoincrement,

  `project_id`          int     not null,
  `storage_id`          int     not null,
  `environment_id`      int,
  `sync_enabled`        boolean not null default false,
  `sync_interval`       int     not null default 0,
  `last_synced_at`      datetime null,
  `last_sync_failed_at` datetime null,

  foreign key (`project_id`) references `project`(`id`) on delete cascade,
  foreign key (`storage_id`) references `project__secret_storage`(`id`) on delete cascade,
  foreign key (`environment_id`) references `project__environment`(`id`) on delete cascade
);

create table `project__secret_sync_path` (
  `id` integer primary key autoincrement,

  `sync_id`     int             not null,
  `path`        varchar(1000)   not null default '',
  `prefix`      varchar(1000)   not null default '',
  `separator`   varchar(20)     not null default '',

  foreign key (`sync_id`) references `project__secret_sync`(`id`) on delete cascade
);

alter table `access_key` add `synchronized` boolean not null default false;
