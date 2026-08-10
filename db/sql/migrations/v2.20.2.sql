create table `project__repository_submodule_credential` (
  `id` integer primary key autoincrement,
  `project_id` int not null,
  `repository_id` int not null,
  `access_key_id` int not null,
  `host` varchar(300) not null,

  foreign key (`project_id`) references `project`(`id`) on delete cascade,
  foreign key (`repository_id`) references `project__repository`(`id`) on delete cascade,
  foreign key (`access_key_id`) references `access_key`(`id`) on delete cascade
);

create unique index `project__repository_submodule_credential__repo_host`
  on `project__repository_submodule_credential`(`repository_id`, `host`);
