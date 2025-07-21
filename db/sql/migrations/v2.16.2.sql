create table project__task_params
(
    id           integer primary key autoincrement,

    variables    TEXT,
    project_id   int not null,
    cli_args     TEXT,
    inventory_id int,
    git_branch   varchar(255),
    params       TEXT,

    foreign key (`project_id`) references project (`id`) on delete cascade,
    foreign key (`inventory_id`) references project__inventory (`id`) on delete cascade
);
