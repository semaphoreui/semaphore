create table runner
(
    id                  integer primary key autoincrement,
    project_id          int,
    token               varchar(255) not null,
    webhook             varchar(1000) not null default '',
    max_parallel_tasks  int not null default 0,

    foreign key (`project_id`) references project(`id`) on delete cascade
);