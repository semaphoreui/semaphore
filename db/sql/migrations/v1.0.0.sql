alter table task add `debug` boolean not null default false;

alter table `project__template` add `arguments` text null;
alter table `project__template` add `override_args` boolean not null default false;
alter table `project__inventory` add `ssh_key_id` int null references access_key(`id`);