create table user (
	`id` int(11) not null auto_increment primary key,
	`created` datetime not null default NOW(),
	`username` varchar(255) not null comment "Username, unique",
	`name` varchar(255) not null comment "Full name",
	`email` varchar(255) not null comment "Email, unique",
	`password` varchar(255) not null comment "Password",

	UNIQUE KEY `username` (`username`),
	UNIQUE KEY `email` (`email`)
) ENGINE=InnoDB CHARSET=utf8;

create table project (
	`id` int(11) not null auto_increment primary key,
	`created` datetime not null default NOW() comment "Created timestamp",
	`name` varchar(255) not null comment "Project name"
) ENGINE=InnoDB CHARSET=utf8;

create table project__user (
	`project_id` int(11) not null,
	`user_id` varchar (255) not null comment "User ID",
	`admin` tinyint (1) not null default 0 comment 'Gives user god-like privileges',

	UNIQUE KEY `id` (`project_id`, `user_id`)
) ENGINE=InnoDB CHARSET=utf8;

