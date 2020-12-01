alter table task__output rename to task__output_backup;

create table task__output
(
    id integer primary key autoincrement,
    task_id int not null
        references task
            on delete cascade,
    task varchar(255) not null,
    time datetime not null,
    output longtext not null
);

insert into task__output(task_id, task, time, output) select * from task__output_backup;

drop table task__output_backup;
