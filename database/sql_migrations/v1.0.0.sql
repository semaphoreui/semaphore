alter table task add `debug` tinyint(1) not null default 0;

alter table project__template add `arguments` text null,
	add `override_args` tinyint(1) not null default 0;

alter table project__inventory add `ssh_key_id` int(11) not null,
	add foreign key (`ssh_key_id`) references access_key(`id`);

alter table task__output drop foreign key `task__output_ibfk_1`;
alter table task__output drop index `id`;
alter table task__output add key `task_id` (`task_id`);
alter table task__output add foreign key (`task_id`) references task(`id`) on delete cascade;