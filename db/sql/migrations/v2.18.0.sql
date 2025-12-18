create table `project__template_additional_repository`
(
    id            integer primary key autoincrement,
    template_id   int          not null,
    repository_id int          not null,
    path          varchar(255) not null,
    git_branch    varchar(255),
    position      int          not null default 0,

    foreign key (`template_id`) references `project__template` (`id`) on delete cascade,
    foreign key (`repository_id`) references `project__repository` (`id`) on delete cascade,

    unique (`template_id`, `path`)
);

create index `idx_template_additional_repo_template`
    on `project__template_additional_repository` (`template_id`);
