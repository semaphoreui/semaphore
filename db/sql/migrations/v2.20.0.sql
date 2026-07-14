create table `user__external_identity` (
  `id` integer primary key autoincrement,
  `user_id` int not null,
  `type` varchar(4) not null,
  `provider` varchar(64) not null,
  `external_uid` varchar(700) not null,
  `created` datetime not null,

  foreign key (`user_id`) references `user`(`id`) on delete cascade
);

create unique index `user__external_identity__type_provider_uid`
  on `user__external_identity`(`type`, `provider`, `external_uid`);
create index `user__external_identity__user_id`
  on `user__external_identity`(`user_id`);
