alter table task add `debug` tinyint not null default 0;

alter table `project__template` add `arguments` text null;
alter table `project__template` add `override_args` tinyint not null default 0;
alter table `project__inventory` add `ssh_key_id` int null references access_key(`id`);

alter table `task__output` rename to `task__output_backup`;
create table `task__output`
(
    task_id int not null
        references task
            on delete cascade,
    task varchar(255) not null,
    time datetime not null,
    output longtext not null
);
insert into `task__output` select * from `task__output_backup`;
drop table `task__output_backup`;
