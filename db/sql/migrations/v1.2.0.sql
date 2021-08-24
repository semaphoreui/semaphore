create table `user__token` (
	`id` varchar(32) not null primary key,
	`created` datetime not null,
	`expired` boolean not null default false,
	`user_id` int not null,

	foreign key (`user_id`) references `user`(`id`) on delete cascade
);
