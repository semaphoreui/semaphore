{{if .Sqlite}}
create table session_old_2_19_14 (
    id                  INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id             INTEGER NOT NULL,
    created             DATETIME NOT NULL,
    last_active         DATETIME NOT NULL,
    ip                  VARCHAR(39) NOT NULL DEFAULT '',
    user_agent          TEXT NOT NULL,
    expired             INTEGER NOT NULL DEFAULT 0,
    verification_method INTEGER NOT NULL DEFAULT 0,
    verified            INTEGER NOT NULL DEFAULT 0,
    FOREIGN KEY(user_id) REFERENCES user(id)
);

insert into session_old_2_19_14 (id, user_id, created, last_active, ip, user_agent, expired, verification_method, verified)
    select id, user_id, created, last_active, ip, user_agent, expired, verification_method, verified from session;

drop table session;

alter table session_old_2_19_14 rename to session;

create index session__session__expired on session(expired);

create index session__session__user_id on session(user_id);

create table task_old_2_20_2 (
    id               INTEGER PRIMARY KEY AUTOINCREMENT,
    template_id      INTEGER NOT NULL REFERENCES project__template(id) ON DELETE CASCADE,
    status           VARCHAR(255) NOT NULL,
    playbook         VARCHAR(255) NOT NULL,
    environment      TEXT NULL,
    created          DATETIME NULL,
    start            DATETIME NULL,
    end              DATETIME NULL,
    user_id          INTEGER REFERENCES user(id),
    project_id       INTEGER REFERENCES project(id),
    message          VARCHAR(250) NOT NULL DEFAULT '',
    version          VARCHAR(20) NULL,
    commit_hash      VARCHAR(64) NULL,
    commit_message   VARCHAR(100) NOT NULL DEFAULT '',
    build_task_id    INTEGER REFERENCES task(id) ON DELETE SET NULL,
    arguments        TEXT NULL,
    inventory_id     INTEGER REFERENCES project__inventory(id) ON DELETE SET NULL,
    integration_id   INTEGER REFERENCES project__integration(id) ON DELETE SET NULL,
    schedule_id      INTEGER REFERENCES project__schedule(id) ON DELETE SET NULL,
    git_branch       VARCHAR(255) NULL,
    params           TEXT NULL,
    runner_id        INTEGER NULL REFERENCES runner(id) ON DELETE SET NULL,
    workflow_run_id  INTEGER NULL,
    workflow_node_id INTEGER NULL,
    artifacts        TEXT NULL
);

insert into task_old_2_20_2 (id, template_id, status, playbook, environment, created, start, `end`, user_id, project_id, message, version, commit_hash, commit_message, build_task_id, arguments, inventory_id, integration_id, schedule_id, git_branch, params, runner_id, workflow_run_id, workflow_node_id, artifacts)
    select id, template_id, status, playbook, environment, created, start, `end`, user_id, project_id, message, version, commit_hash, commit_message, build_task_id, arguments, inventory_id, integration_id, schedule_id, git_branch, params, runner_id, workflow_run_id, workflow_node_id, artifacts
    from task;

drop table task;

alter table task_old_2_20_2 rename to task;

create index task__integration_id on task(integration_id);

create index task__inventory_id on task(inventory_id);

create index task__project_id on task(project_id);

create index task__schedule_id on task(schedule_id);

create index task__template_id on task(template_id);

create index task__workflow_run_id on task(workflow_run_id);

create index task__workflow_node_id on task(workflow_node_id);
{{end}}
