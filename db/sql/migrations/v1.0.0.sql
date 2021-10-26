alter table task add `debug` boolean not null default false;

alter table `project__template` add `arguments` text null;
alter table `project__template` add `override_args` boolean not null default false;
alter table `project__inventory` add `ssh_key_id` int null references access_key(`id`);

alter table `task__output` rename to `task__output_backup`;
create table `task__output`
(
    task_id int not null,
    task varchar(255) not null,
    time datetime not null,
    output longtext not null,

    foreign key (`task_id`) references task(`id`) on delete cascade
);
insert into `task__output` select * from `task__output_backup`;
drop table `task__output_backup`;
