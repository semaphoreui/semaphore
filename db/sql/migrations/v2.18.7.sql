{{if .Sqlite}}
create table project__repository_dg_tmp
(
    id         INTEGER
        primary key autoincrement,
    project_id INTEGER                 not null
        references project
            on delete cascade,
    git_url    TEXT                    not null,
    ssh_key_id INTEGER
        references access_key,
    name       VARCHAR(255),
    git_branch VARCHAR(255) default '' not null
);

insert into project__repository_dg_tmp(id, project_id, git_url, ssh_key_id, name, git_branch)
select id, project_id, git_url, ssh_key_id, name, git_branch
from project__repository;

drop table project__repository;

alter table project__repository_dg_tmp rename to project__repository;

create index project__repository__project_id on project__repository (project_id);

create index project__repository__ssh_key_id on project__repository (ssh_key_id);
{{else if .Mysql}}
alter table project__repository modify ssh_key_id int null;
{{else if .Postgresql}}
alter table public.project__repository alter column ssh_key_id drop not null;
{{end}}
