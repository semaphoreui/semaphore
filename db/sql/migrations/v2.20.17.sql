{{if .Sqlite}}
create table project__schedule_new_2_20_17 (
    id               INTEGER PRIMARY KEY AUTOINCREMENT,
    template_id      INTEGER NULL REFERENCES project__template(id) ON DELETE CASCADE,
    workflow_id      INTEGER NULL REFERENCES project__workflow_template(id) ON DELETE CASCADE,
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

insert into project__schedule_new_2_20_17 (id, template_id, project_id, cron_format, repository_id, last_commit_hash, name, active, task_params_id, run_at, type, delete_after_run)
    select id, template_id, project_id, cron_format, repository_id, last_commit_hash, name, active, task_params_id, run_at, type, delete_after_run from project__schedule;

drop table project__schedule;

alter table project__schedule_new_2_20_17 rename to project__schedule;

create index project__schedule__project_id on project__schedule(project_id);

create index project__schedule__repository_id on project__schedule(repository_id);

create index project__schedule__template_id on project__schedule(template_id);

create index project__schedule__workflow_id on project__schedule(workflow_id);
{{else}}
alter table `project__schedule` change `template_id` `template_id` int;

alter table `project__schedule` add `workflow_id` int null;

alter table `project__schedule` add constraint `project__schedule__workflow_fk` foreign key (`workflow_id`) references `project__workflow_template`(`id`) on delete cascade;

create index `project__schedule__workflow_id` on `project__schedule`(`workflow_id`);
{{end}}
