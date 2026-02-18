create table project__task_params
(
    id           integer primary key autoincrement,

    environment  TEXT,
    project_id   int not null,
    arguments    TEXT,
    inventory_id int,
    git_branch   varchar(255),
    params       TEXT,
    version      varchar(20),
    message      varchar(250),

    foreign key (`project_id`) references project (`id`) on delete cascade,
    foreign key (`inventory_id`) references project__inventory (`id`) on delete cascade
);

alter table project__integration drop task_params;
alter table project__schedule add task_params_id int references `project__task_params`(`id`);
alter table project__integration add task_params_id int references `project__task_params`(`id`);

