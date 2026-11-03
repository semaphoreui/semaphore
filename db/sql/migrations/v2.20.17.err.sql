{{if .Sqlite}}
update task set schedule_id = null where schedule_id in (select id from project__schedule where template_id is null);
{{end}}
delete from `project__schedule` where `template_id` is null;

{{if .Sqlite}}
create table project__schedule_old_2_20_17 (
    id               INTEGER PRIMARY KEY AUTOINCREMENT,
    template_id      INTEGER NOT NULL REFERENCES project__template(id) ON DELETE CASCADE,
    project_id       INTEGER NOT NULL REFERENCES project(id) ON DELETE CASCADE,
    cron_format      VARCHAR(255) NOT NULL,
    repository_id    INTEGER REFERENCES project__repository(id),
    last_commit_hash VARCHAR(64) NULL,
    name             VARCHAR(100) NOT NULL DEFAULT '',
    active           INTEGER NOT NULL DEFAULT 1,
    task_params_id   INTEGER REFERENCES project__task_params(id),
    run_at           DATETIME NULL,
    type             VARCHAR(20) NOT NULL DEFAULT '',
    delete_after_run BOOLEAN NOT NULL DEFAULT FALSE
);

insert into project__schedule_old_2_20_17 (id, template_id, project_id, cron_format, repository_id, last_commit_hash, name, active, task_params_id, run_at, type, delete_after_run)
    select id, template_id, project_id, cron_format, repository_id, last_commit_hash, name, active, task_params_id, run_at, type, delete_after_run from project__schedule;

drop table project__schedule;

alter table project__schedule_old_2_20_17 rename to project__schedule;

create index project__schedule__project_id on project__schedule(project_id);

create index project__schedule__repository_id on project__schedule(repository_id);

create index project__schedule__template_id on project__schedule(template_id);
{{else}}
alter table `project__schedule` drop foreign key `project__schedule__workflow_fk`;

alter table `project__schedule` drop column `workflow_id`;

alter table `project__schedule` change `template_id` `template_id` int not null;
{{end}}
