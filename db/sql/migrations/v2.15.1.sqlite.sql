CREATE TABLE "option" (
                          "key"  VARCHAR(255) NOT NULL PRIMARY KEY,
                          value  VARCHAR(1000) NOT NULL
);

CREATE TABLE project (
                         id                 INTEGER PRIMARY KEY AUTOINCREMENT,
                         created            DATETIME   NOT NULL,
                         name               VARCHAR(255) NOT NULL,
                         alert              INTEGER    NOT NULL DEFAULT 0,
                         alert_chat         VARCHAR(30) NULL,
                         max_parallel_tasks INTEGER    NOT NULL DEFAULT 0,
                         type               VARCHAR(20) NULL DEFAULT ''
);

CREATE TABLE project__environment (
                                      id          INTEGER PRIMARY KEY AUTOINCREMENT,
                                      project_id  INTEGER NOT NULL REFERENCES project(id) ON DELETE CASCADE,
                                      password    VARCHAR(255) NULL,
                                      json        TEXT NOT NULL,
                                      name        VARCHAR(255) NULL,
                                      env         TEXT
);

CREATE INDEX project__environment__project__environment_project_id
    ON project__environment(project_id);

CREATE TABLE project__view (
                               id          INTEGER PRIMARY KEY AUTOINCREMENT,
                               title       VARCHAR(100) NOT NULL,
                               project_id  INTEGER NOT NULL REFERENCES project(id) ON DELETE CASCADE,
                               position    INTEGER NOT NULL
);

CREATE INDEX project__view__project__view_project_id
    ON project__view(project_id);

CREATE TABLE runner (
                        id                  INTEGER PRIMARY KEY AUTOINCREMENT,
                        project_id          INTEGER REFERENCES project(id) ON DELETE CASCADE,
                        token               VARCHAR(255) NOT NULL,
                        webhook             VARCHAR(1000) NOT NULL DEFAULT '',
                        max_parallel_tasks  INTEGER NOT NULL DEFAULT 0,
                        name                VARCHAR(100) NOT NULL DEFAULT '',
                        active              INTEGER NOT NULL DEFAULT 1,
                        public_key          TEXT NULL,
                        tag                 VARCHAR(200) NOT NULL DEFAULT '',
                        touched             DATETIME NULL,
                        cleaning_requested  DATETIME NULL
);

CREATE INDEX runner__runner__project_id
    ON runner(project_id);

CREATE TABLE session (
                         id              INTEGER PRIMARY KEY AUTOINCREMENT,
                         user_id         INTEGER NOT NULL,
                         created         DATETIME NOT NULL,
                         last_active     DATETIME NOT NULL,
                         ip              VARCHAR(39) NOT NULL DEFAULT '',
                         user_agent      TEXT NOT NULL,
                         expired         INTEGER NOT NULL DEFAULT 0,
                         verification_method INTEGER NOT NULL DEFAULT 0,
                         verified        INTEGER NOT NULL DEFAULT 0,
                         FOREIGN KEY(user_id) REFERENCES user(id)
);

CREATE INDEX session__session__expired
    ON session(expired);

CREATE INDEX session__session__user_id
    ON session(user_id);

CREATE TABLE user (
                      id        INTEGER PRIMARY KEY AUTOINCREMENT,
                      created   DATETIME NOT NULL,
                      username  VARCHAR(255) NOT NULL UNIQUE,
                      name      VARCHAR(255) NOT NULL,
                      email     VARCHAR(255) NOT NULL UNIQUE,
                      password  VARCHAR(255) NOT NULL,
                      alert     INTEGER NOT NULL DEFAULT 0,
                      external  INTEGER NOT NULL DEFAULT 0,
                      admin     INTEGER NOT NULL DEFAULT 1,
                      pro       INTEGER NOT NULL DEFAULT 0
);

CREATE TABLE access_key (
                            id             INTEGER PRIMARY KEY AUTOINCREMENT,
                            name           VARCHAR(255) NOT NULL,
                            type           VARCHAR(255) NOT NULL,
                            project_id     INTEGER REFERENCES project(id) ON DELETE SET NULL,
                            secret         TEXT NULL,
                            environment_id INTEGER REFERENCES project__environment(id) ON DELETE CASCADE,
                            user_id        INTEGER REFERENCES user(id) ON DELETE CASCADE
);

CREATE INDEX access_key__environment_id
    ON access_key(environment_id);

CREATE INDEX access_key__project_id
    ON access_key(project_id);

CREATE INDEX access_key__user_id
    ON access_key(user_id);

CREATE TABLE event (
                       id          INTEGER PRIMARY KEY AUTOINCREMENT,
                       project_id  INTEGER REFERENCES project(id) ON DELETE CASCADE,
                       object_id   INTEGER NULL,
                       object_type VARCHAR(20) NULL DEFAULT '',
                       description TEXT NULL,
                       created     DATETIME NOT NULL,
                       user_id     INTEGER REFERENCES user(id) ON DELETE SET NULL
);

CREATE INDEX event__project_id
    ON event(project_id);

CREATE INDEX event__user_id
    ON event(user_id);

CREATE TABLE event_backup_5784568 (
                                      project_id  INTEGER NULL,
                                      object_id   INTEGER NULL,
                                      object_type VARCHAR(20) NULL DEFAULT '',
                                      description TEXT NULL,
                                      created     DATETIME NOT NULL,
                                      user_id     INTEGER REFERENCES user(id) ON DELETE SET NULL
);

CREATE INDEX event_backup_5784568__user_id
    ON event_backup_5784568(user_id);

CREATE TABLE project__repository (
                                     id          INTEGER PRIMARY KEY AUTOINCREMENT,
                                     project_id  INTEGER NOT NULL REFERENCES project(id) ON DELETE CASCADE,
                                     git_url     TEXT NOT NULL,
                                     ssh_key_id  INTEGER NOT NULL REFERENCES access_key(id),
                                     name        VARCHAR(255) NULL,
                                     git_branch  VARCHAR(255) NOT NULL DEFAULT ''
);

CREATE INDEX project__repository__project_id
    ON project__repository(project_id);

CREATE INDEX project__repository__ssh_key_id
    ON project__repository(ssh_key_id);

CREATE TABLE project__inventory (
                                    id             INTEGER PRIMARY KEY AUTOINCREMENT,
                                    project_id     INTEGER NOT NULL REFERENCES project(id) ON DELETE CASCADE,
                                    type           VARCHAR(255) NOT NULL,
                                    inventory      TEXT NOT NULL,
                                    ssh_key_id     INTEGER REFERENCES access_key(id),
                                    name           VARCHAR(255) NULL,
                                    become_key_id  INTEGER REFERENCES access_key(id),
                                    template_id    INTEGER REFERENCES project__template(id) ON DELETE SET NULL,
                                    repository_id  INTEGER REFERENCES project__repository(id) ON DELETE SET NULL,
                                    runner_tag     VARCHAR(255) NULL
);

CREATE INDEX project__inventory__become_key_id
    ON project__inventory(become_key_id);

CREATE INDEX project__inventory__holder_id
    ON project__inventory(template_id);

CREATE INDEX project__inventory__project_id
    ON project__inventory(project_id);

CREATE INDEX project__inventory__repository_id
    ON project__inventory(repository_id);

CREATE INDEX project__inventory__ssh_key_id
    ON project__inventory(ssh_key_id);

CREATE TABLE project__template (
                                   id                            INTEGER PRIMARY KEY AUTOINCREMENT,
                                   project_id                    INTEGER NOT NULL REFERENCES project(id) ON DELETE CASCADE,
                                   inventory_id                  INTEGER REFERENCES project__inventory(id),
                                   repository_id                 INTEGER NOT NULL REFERENCES project__repository(id),
                                   environment_id                INTEGER REFERENCES project__environment(id),
                                   playbook                      VARCHAR(255) NOT NULL,
                                   arguments                     TEXT NULL,
                                   name                          VARCHAR(100) NOT NULL,
                                   description                   TEXT NULL,
                                   type                          VARCHAR(10) NOT NULL DEFAULT '',
                                   start_version                 VARCHAR(20) NULL,
                                   build_template_id             INTEGER REFERENCES project__template(id),
                                   view_id                       INTEGER REFERENCES project__view(id) ON DELETE SET NULL,
                                   survey_vars                   TEXT NULL,
                                   autorun                       INTEGER NULL DEFAULT 0,
                                   allow_override_args_in_task   INTEGER NOT NULL DEFAULT 0,
                                   suppress_success_alerts       INTEGER NOT NULL DEFAULT 0,
                                   app                           VARCHAR(50) NOT NULL,
                                   tasks                         INTEGER NOT NULL DEFAULT 0,
                                   git_branch                    VARCHAR(255) NULL,
                                   task_params                   TEXT NULL,
                                   runner_tag                    VARCHAR(50) NULL,
                                   allow_override_branch_in_task INTEGER NOT NULL DEFAULT 0
);

CREATE INDEX project__template__build_template_id
    ON project__template(build_template_id);

CREATE INDEX project__template__environment_id
    ON project__template(environment_id);

CREATE INDEX project__template__inventory_id
    ON project__template(inventory_id);

CREATE INDEX project__template__project_id
    ON project__template(project_id);

CREATE INDEX project__template__repository_id
    ON project__template(repository_id);

CREATE INDEX project__template__view_id
    ON project__template(view_id);

CREATE TABLE project__integration (
                                      id             INTEGER PRIMARY KEY AUTOINCREMENT,
                                      name           VARCHAR(255) NOT NULL,
                                      project_id     INTEGER NOT NULL REFERENCES project(id) ON DELETE CASCADE,
                                      template_id    INTEGER NOT NULL REFERENCES project__template(id) ON DELETE CASCADE,
                                      auth_method    VARCHAR(15) NOT NULL DEFAULT 'none',
                                      auth_secret_id INTEGER REFERENCES access_key(id) ON DELETE SET NULL,
                                      auth_header    VARCHAR(255) NULL,
                                      searchable     INTEGER NOT NULL DEFAULT 0,
                                      task_params    TEXT NULL
);

CREATE INDEX project__integration__auth_secret_id
    ON project__integration(auth_secret_id);

CREATE INDEX project__integration__project_id
    ON project__integration(project_id);

CREATE INDEX project__integration__template_id
    ON project__integration(template_id);

CREATE TABLE project__integration_alias (
                                            id             INTEGER PRIMARY KEY AUTOINCREMENT,
                                            alias          VARCHAR(50) NOT NULL UNIQUE,
                                            project_id     INTEGER NOT NULL REFERENCES project(id) ON DELETE CASCADE,
                                            integration_id INTEGER REFERENCES project__integration(id) ON DELETE CASCADE
);

CREATE INDEX project__integration_alias__integration_id
    ON project__integration_alias(integration_id);

CREATE INDEX project__integration_alias__project_id
    ON project__integration_alias(project_id);

CREATE TABLE project__integration_extract_value (
                                                    id             INTEGER PRIMARY KEY AUTOINCREMENT,
                                                    name           VARCHAR(255) NOT NULL,
                                                    integration_id INTEGER NOT NULL REFERENCES project__integration(id) ON DELETE CASCADE,
                                                    value_source   VARCHAR(255) NOT NULL,
                                                    body_data_type VARCHAR(255) NULL,
                                                    "key"          VARCHAR(255) NULL,
                                                    variable       VARCHAR(255) NULL,
                                                    variable_type  VARCHAR(255) NULL
);

CREATE INDEX project__integration_extract_value__integration_id
    ON project__integration_extract_value(integration_id);

CREATE TABLE project__integration_matcher (
                                              id             INTEGER PRIMARY KEY AUTOINCREMENT,
                                              name           VARCHAR(255) NOT NULL,
                                              integration_id INTEGER NOT NULL REFERENCES project__integration(id) ON DELETE CASCADE,
                                              match_type     VARCHAR(255) NULL,
                                              method         VARCHAR(255) NULL,
                                              body_data_type VARCHAR(255) NULL,
                                              "key"          VARCHAR(510) NULL,
                                              value          VARCHAR(510) NULL
);

CREATE INDEX project__integration_matcher__integration_id
    ON project__integration_matcher(integration_id);

CREATE TABLE project__schedule (
                                   id               INTEGER PRIMARY KEY AUTOINCREMENT,
                                   template_id      INTEGER NOT NULL REFERENCES project__template(id) ON DELETE CASCADE,
                                   project_id       INTEGER NOT NULL REFERENCES project(id) ON DELETE CASCADE,
                                   cron_format      VARCHAR(255) NOT NULL,
                                   repository_id    INTEGER REFERENCES project__repository(id),
                                   last_commit_hash VARCHAR(64) NULL,
                                   name             VARCHAR(100) NOT NULL DEFAULT '',
                                   active           INTEGER NOT NULL DEFAULT 1
);

CREATE INDEX project__schedule__project_id
    ON project__schedule(project_id);

CREATE INDEX project__schedule__repository_id
    ON project__schedule(repository_id);

CREATE INDEX project__schedule__template_id
    ON project__schedule(template_id);

CREATE TABLE project__template_vault (
                                         id           INTEGER PRIMARY KEY AUTOINCREMENT,
                                         project_id   INTEGER NOT NULL REFERENCES project(id) ON DELETE CASCADE,
                                         template_id  INTEGER NOT NULL REFERENCES project__template(id) ON DELETE CASCADE,
                                         vault_key_id INTEGER REFERENCES access_key(id) ON DELETE CASCADE,
                                         name         VARCHAR(255) NULL,
                                         type         VARCHAR(20) NOT NULL DEFAULT 'password',
                                         script       TEXT NULL,
                                         UNIQUE(template_id, vault_key_id, name)
);

CREATE INDEX project__template_vault__project_id
    ON project__template_vault(project_id);

CREATE INDEX project__template_vault__vault_key_id
    ON project__template_vault(vault_key_id);

CREATE TABLE project__terraform_inventory_alias (
                                                    alias        VARCHAR(100) PRIMARY KEY,
                                                    project_id   INTEGER NOT NULL REFERENCES project(id) ON DELETE CASCADE,
                                                    inventory_id INTEGER NOT NULL REFERENCES project__inventory(id) ON DELETE CASCADE,
                                                    auth_key_id  INTEGER NOT NULL REFERENCES access_key(id)
);

CREATE INDEX project__terraform_inventory_alias__auth_key_id
    ON project__terraform_inventory_alias(auth_key_id);

CREATE INDEX project__terraform_inventory_alias__inventory_id
    ON project__terraform_inventory_alias(inventory_id);

CREATE INDEX project__terraform_inventory_alias__project_id
    ON project__terraform_inventory_alias(project_id);

CREATE TABLE project__user (
                               project_id INTEGER NOT NULL REFERENCES project(id) ON DELETE CASCADE,
                               user_id    INTEGER NOT NULL REFERENCES user(id) ON DELETE CASCADE,
                               role       VARCHAR(50) NOT NULL DEFAULT 'manager',
                               UNIQUE(project_id, user_id)
);

CREATE INDEX project__user__user_id
    ON project__user(user_id);

CREATE TABLE task (
                      id              INTEGER PRIMARY KEY AUTOINCREMENT,
                      template_id     INTEGER NOT NULL REFERENCES project__template(id) ON DELETE CASCADE,
                      status          VARCHAR(255) NOT NULL,
                      playbook        VARCHAR(255) NOT NULL,
                      environment     TEXT NULL,
                      created         DATETIME NULL,
                      start           DATETIME NULL,
                      end             DATETIME NULL,
                      user_id         INTEGER REFERENCES user(id),
                      project_id      INTEGER REFERENCES project(id),
                      message         VARCHAR(250) NOT NULL DEFAULT '',
                      version         VARCHAR(20) NULL,
                      commit_hash     VARCHAR(64) NULL,
                      commit_message  VARCHAR(100) NOT NULL DEFAULT '',
                      build_task_id   INTEGER REFERENCES task(id) ON DELETE SET NULL,
                      arguments       TEXT NULL,
                      inventory_id    INTEGER REFERENCES project__inventory(id) ON DELETE SET NULL,
                      integration_id  INTEGER REFERENCES project__integration(id) ON DELETE SET NULL,
                      schedule_id     INTEGER REFERENCES project__schedule(id) ON DELETE SET NULL,
                      git_branch      VARCHAR(255) NULL,
                      params          TEXT NULL
);

CREATE INDEX task__integration_id
    ON task(integration_id);

CREATE INDEX task__inventory_id
    ON task(inventory_id);

CREATE INDEX task__project_id
    ON task(project_id);

CREATE INDEX task__schedule_id
    ON task(schedule_id);

CREATE INDEX task__template_id
    ON task(template_id);

CREATE TABLE project__terraform_inventory_state (
                                                    id           INTEGER PRIMARY KEY AUTOINCREMENT,
                                                    project_id   INTEGER NOT NULL REFERENCES project(id) ON DELETE CASCADE,
                                                    inventory_id INTEGER NOT NULL REFERENCES project__inventory(id) ON DELETE CASCADE,
                                                    state        TEXT     NOT NULL,
                                                    created      DATETIME NOT NULL,
                                                    task_id      INTEGER REFERENCES task(id) ON DELETE SET NULL
);

CREATE INDEX project__terraform_inventory_state__inventory_id
    ON project__terraform_inventory_state(inventory_id);

CREATE INDEX project__terraform_inventory_state__project_id
    ON project__terraform_inventory_state(project_id);

CREATE INDEX project__terraform_inventory_state__task_id
    ON project__terraform_inventory_state(task_id);

CREATE TABLE task__output (
                              id      INTEGER PRIMARY KEY AUTOINCREMENT,
                              task_id INTEGER NOT NULL REFERENCES task(id) ON DELETE CASCADE,
                              time    DATETIME NOT NULL,
                              output  TEXT NOT NULL
);

CREATE INDEX task__output__task__output_time_idx
    ON task__output(time);

CREATE INDEX task__output__task_id
    ON task__output(task_id);

CREATE TABLE task__stage (
                             id              INTEGER PRIMARY KEY AUTOINCREMENT,
                             task_id         INTEGER NOT NULL REFERENCES task(id) ON DELETE CASCADE,
                             start           DATETIME NULL,
                             start_output_id INTEGER REFERENCES task__output(id) ON DELETE SET NULL,
                             end             DATETIME NULL,
                             end_output_id   INTEGER REFERENCES task__output(id) ON DELETE SET NULL,
                             type            VARCHAR(100) NULL
);

CREATE INDEX task__stage__end_output_id
    ON task__stage(end_output_id);

CREATE INDEX task__stage__start_output_id
    ON task__stage(start_output_id);

CREATE INDEX task__stage__task_id
    ON task__stage(task_id);

CREATE TABLE task__stage_result (
                                    id       INTEGER PRIMARY KEY AUTOINCREMENT,
                                    task_id  INTEGER NOT NULL REFERENCES task(id) ON DELETE CASCADE,
                                    stage_id INTEGER NOT NULL REFERENCES task__stage(id) ON DELETE CASCADE,
                                    json     TEXT NULL
);

CREATE INDEX task__stage_result__stage_id
    ON task__stage_result(stage_id);

CREATE INDEX task__stage_result__task_id
    ON task__stage_result(task_id);

CREATE TABLE user__token (
                             id      VARCHAR(44)   NOT NULL PRIMARY KEY,
                             created DATETIME      NOT NULL,
                             expired INTEGER NOT NULL DEFAULT 0,
                             user_id INTEGER NOT NULL REFERENCES user(id) ON DELETE CASCADE
);

CREATE INDEX user__token__user_id
    ON user__token(user_id);

CREATE TABLE user__totp (
                            id            INTEGER PRIMARY KEY AUTOINCREMENT,
                            user_id       INTEGER NOT NULL UNIQUE REFERENCES user(id) ON DELETE CASCADE,
                            url           VARCHAR(250) NOT NULL,
                            recovery_hash VARCHAR(250) NOT NULL,
                            created       DATETIME NOT NULL
);