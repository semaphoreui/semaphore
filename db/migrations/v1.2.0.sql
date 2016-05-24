create table `user__token` (
	`id` varchar(32) not null primary key,
	`created` datetime not null,
	`expired` tinyint(1) not null default 0,
	`user_id` int(11) not null,

	foreign key (`user_id`) references user(`id`) on delete cascade
) ENGINE=InnoDB CHARSET=utf8;