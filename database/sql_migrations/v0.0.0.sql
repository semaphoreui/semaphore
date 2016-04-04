create table user (
	`id` int(11) not null auto_increment primary key,
	`created` datetime not null default NOW(),
	`username` varchar(255) not null comment "Username, unique",
	`name` varchar(255) not null comment "Full name",
	`email` varchar(255) not null comment "Email, unique",
	`password` varchar(255) not null comment "Password",

	unique key `username` (`username`),
	unique key `email` (`email`)
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

	unique key `id` (`project_id`, `user_id`),
	foreign key (`project_id`) references project(`id`) on delete cascade,
	foreign key (`user_id`) references user(`id`) on delete cascade
) ENGINE=InnoDB CHARSET=utf8;

create table access_key (
	`id` int(11) not null primary key auto_increment,
	`name` varchar(255) not null,
	`type` varchar(255) not null comment 'aws/do/gcloud/ssh',

	`project_id` int(11) null,
	`key` text null,
	`secret` text null,

	foreign key (`project_id`) references project(`id`) on delete set null
) ENGINE=InnoDB CHARSET=utf8;

create table project__repository (
	`id` int(11) not null primary key auto_increment,
	`project_id` int(11) not null,
	`git_url` text not null,
	`ssh_key_id` int(11) not null,

	foreign key (`project_id`) references project(`id`) on delete cascade,
	foreign key (`ssh_key_id`) references access_key(`id`)
) ENGINE=InnoDB CHARSET=utf8;

create table project__inventory (
	`id` int(11) not null primary key auto_increment,
	`project_id` int(11) not null,
	`type` varchar(255) not null comment 'can be static/aws/do/gcloud',
	`key_id` int(11) null comment 'references keys to authenticate remote services',
	`inventory` longtext not null,

	foreign key (`project_id`) references project(`id`) on delete cascade,
	foreign key (`key_id`) references access_key(`id`)
) ENGINE=InnoDB CHARSET=utf8;

create table project__environment (
	`id` int(11) not null primary key auto_increment,
	`project_id` int(11) not null,
	`password` varchar(255) null,
	`json` longtext not null,

	foreign key (`project_id`) references project(`id`) on delete cascade
) ENGINE=InnoDB CHARSET=utf8;

create table project__template (
	`id` int(11) not null primary key auto_increment,
	`ssh_key_id` int(11) not null comment 'for accessing the inventory',
	`project_id` int(11) not null,
	`inventory_id` int(11) not null,
	`repository_id` int(11) not null,
	`environment_id` int(11) null,
	`playbook` varchar(255) not null comment 'playbook name (ansible.yml)',

	foreign key (`project_id`) references project(`id`) on delete cascade,
	foreign key (`ssh_key_id`) references access_key(`id`),
	foreign key (`inventory_id`) references project__inventory(`id`),
	foreign key (`repository_id`) references project__repository(`id`),
	foreign key (`environment_id`) references project__environment(`id`)
) ENGINE=InnoDB CHARSET=utf8;

create table project__template_schedule (
	`template_id` int(11) not null,
	`cron_format` varchar(255) not null,

	foreign key (`template_id`) references project__template(`id`) on delete cascade
) ENGINE=InnoDB CHARSET=utf8;

create table task (
	`id` int(11) not null primary key auto_increment,
	`template_id` int(11) not null,
	`status` varchar(255) not null,
	`playbook` varchar(255) not null comment 'override playbook name (ansible.yml)',
	`environment` longtext null comment 'override environment',

	foreign key (`template_id`) references project__template(`id`)
) ENGINE=InnoDB CHARSET=utf8;

create table task__output (
	`task_id` int(11) not null,
	`time` datetime not null default NOW(),
	`output` longtext not null,

	unique key `id` (`task_id`, `time`),
	foreign key (`task_id`) references task(`id`) on delete cascade
) ENGINE=InnoDB CHARSET=utf8;