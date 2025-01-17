create table user__totp(
  `id` integer primary key autoincrement,
  `user_id` int NOT NULL,
  `url` varchar(250) NOT NULL,
  `created` datetime NOT NULL,
  unique (`user_id`),
  foreign key (`user_id`) references `user`(`id`) on delete cascade
);

alter table `session` add column verification_method int not null default 0;
alter table `session` add column verified boolean not null default false;