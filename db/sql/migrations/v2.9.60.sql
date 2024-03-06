create table project__integration (
  `id` integer primary key autoincrement,
  `name` varchar(255) not null,
  `project_id` int not null,
  `template_id` int not null,
  `auth_method` varchar(15) not null default 'none',
  `auth_secret_id` int,
  `auth_header` varchar(255),

  foreign key (`project_id`) references project(`id`) on delete cascade,
  foreign key (`template_id`) references project__template(`id`) on delete cascade,
  foreign key (`auth_secret_id`) references access_key(`id`) on delete set null
);

create table project__integration_extractor (
  `id` integer primary key autoincrement,
  `name` varchar(255) not null,
  `integration_id` int not null,

  foreign key (`integration_id`) references project__integration(`id`) on delete cascade
);

create table project__integration_extract_value (
  `id` integer primary key autoincrement,
  `name` varchar(255) not null,
  `extractor_id` int not null,
  `value_source` varchar(255) not null,
  `body_data_type` varchar(255) null,
  `key` varchar(255) null,
  `variable` varchar(255) null,

  foreign key (`extractor_id`) references project__integration_extractor(`id`) on delete cascade
);

create table project__integration_matcher (
  `id` integer primary key autoincrement,
  `name` varchar(255) not null,
  `extractor_id` int not null,
  `match_type` varchar(255) null,
  `method` varchar(255) null,
  `body_data_type` varchar(255) null,
  `key` varchar(510) null,
  `value` varchar(510) null,

  foreign key (`extractor_id`) references project__integration_extractor(`id`) on delete cascade
);
