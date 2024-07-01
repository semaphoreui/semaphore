create table project__environment_secret (
      `id` integer primary key autoincrement,
      `name` varchar(255) not null,
      `environment_id` int not null,
      `secret_id` int not null,

      foreign key (`environment_id`) references project__environment(`id`) on delete cascade,
      foreign key (`secret_id`) references access_key(`id`) on delete cascade
);