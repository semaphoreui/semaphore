create table project__terraform_inventory_alias(
  `alias` varchar(100) primary key,
  `project_id` int NOT NULL,
  `inventory_id` int NOT NULL,
  `auth_key_id` int NOT NULL,
  foreign key (`project_id`) references project(`id`) on delete cascade,
  foreign key (`inventory_id`) references project__inventory(`id`) on delete cascade,
  foreign key (`auth_key_id`) references access_key(`id`)
);

create table project__terraform_inventory_state(
  `id` integer primary key autoincrement,
  `project_id` int NOT NULL,
  `inventory_id` int NOT NULL,
  `state` text NOT NULL,
  `created` datetime NOT NULL,
  `task_id` int,
  foreign key (`task_id`) references task(`id`) on delete set null,
  foreign key (`project_id`) references project(`id`) on delete cascade,
  foreign key (`inventory_id`) references project__inventory(`id`) on delete cascade
);

alter table `project__inventory` change `holder_id` `template_id` int