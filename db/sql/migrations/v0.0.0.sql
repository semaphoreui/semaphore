create table `user` (
	`id` integer primary key autoincrement,
	`created` datetime not null,
	`username` varchar(255) not null,
	`name` varchar(255) not null,
	`email` varchar(255) not null,
	`password` varchar(255) not null,

	unique (`username`),
	unique (`email`)
);

create table `project` (
	`id` integer primary key autoincrement,
	`created` datetime not null,
	`name` varchar(255) not null
);

create table `project__user` (
	`project_id` int not null,
	`user_id` int not null,
	`admin` boolean not null default false,

	unique (`project_id`, `user_id`),
	foreign key (`project_id`) references project(`id`) on delete cascade,
	foreign key (`user_id`) references `user`(`id`) on delete cascade
);

create table `access_key` (
	`id` integer primary key autoincrement,
	`name` varchar(255) not null,
	`type` varchar(255) not null,

	`project_id` int null,
	`key` text null,
	`secret` text null,

	foreign key (`project_id`) references project(`id`) on delete set null
);

create table `project__repository` (
	`id` integer primary key autoincrement,
	`project_id` int not null,
	`git_url` text not null,
	`ssh_key_id` int not null,

	foreign key (`project_id`) references project(`id`) on delete cascade,
	foreign key (`ssh_key_id`) references access_key(`id`)
);

create table `project__inventory` (
	`id` integer primary key autoincrement,
	`project_id` int not null,
	`type` varchar(255) not null,
	`key_id` int null,
	`inventory` longtext not null,

	foreign key (`project_id`) references project(`id`) on delete cascade,
	foreign key (`key_id`) references access_key(`id`)
);

create table `project__environment` (
	`id` integer primary key autoincrement,
	`project_id` int not null,
	`password` varchar(255) null,
	`json` longtext not null,

	foreign key (`project_id`) references project(`id`) on delete cascade
);

create table `project__template` (
	`id` integer primary key autoincrement,
	`ssh_key_id` int not null,
	`project_id` int not null,
	`inventory_id` int not null,
	`repository_id` int not null,
	`environment_id` int null,
	`playbook` varchar(255) not null,

	foreign key (`project_id`) references project(`id`) on delete cascade,
	foreign key (`ssh_key_id`) references access_key(`id`),
	foreign key (`inventory_id`) references project__inventory(`id`),
	foreign key (`repository_id`) references project__repository(`id`),
	foreign key (`environment_id`) references project__environment(`id`)
);

create table `project__template_schedule` (
	`template_id` int primary key,
	`cron_format` varchar(255) not null,

	foreign key (`template_id`) references project__template(`id`) on delete cascade
);

create table `task` (
	`id` integer primary key autoincrement,
	`template_id` int not null,
	`status` varchar(255) not null,
	`playbook` varchar(255) not null,
	`environment` longtext null,

	foreign key (`template_id`) references project__template(`id`) on delete cascade
);

create table `task__output` (
	`task_id` int not null,
	`task` varchar(255) not null,
	`time` datetime not null,
	`output` longtext not null,

	unique (`task_id`, `time`),
	foreign key (`task_id`) references task(`id`) on delete cascade
);
