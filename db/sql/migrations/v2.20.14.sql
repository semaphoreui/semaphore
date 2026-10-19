{{if .Sqlite}}
create table project__integration_new_2_20_14 (
    id             INTEGER PRIMARY KEY AUTOINCREMENT,
    name           VARCHAR(255) NOT NULL,
    project_id     INTEGER NOT NULL REFERENCES project(id) ON DELETE CASCADE,
    template_id    INTEGER NULL REFERENCES project__template(id) ON DELETE CASCADE,
    workflow_id    INTEGER NULL REFERENCES project__workflow_template(id) ON DELETE CASCADE,
    auth_method    VARCHAR(15) NOT NULL DEFAULT 'none',
    auth_secret_id INTEGER REFERENCES access_key(id) ON DELETE SET NULL,
    auth_header    VARCHAR(255) NULL,
    searchable     INTEGER NOT NULL DEFAULT 0,
    task_params_id INTEGER REFERENCES project__task_params(id)
);

insert into project__integration_new_2_20_14 (id, name, project_id, template_id, auth_method, auth_secret_id, auth_header, searchable, task_params_id)
    select id, name, project_id, template_id, auth_method, auth_secret_id, auth_header, searchable, task_params_id from project__integration;

drop table project__integration;

alter table project__integration_new_2_20_14 rename to project__integration;

create index project__integration__auth_secret_id on project__integration(auth_secret_id);

create index project__integration__project_id on project__integration(project_id);

create index project__integration__template_id on project__integration(template_id);

create index project__integration__workflow_id on project__integration(workflow_id);
{{else}}
alter table `project__integration` change `template_id` `template_id` int;

alter table `project__integration` add `workflow_id` int null;

alter table `project__integration` add constraint `project__integration__workflow_fk` foreign key (`workflow_id`) references `project__workflow_template`(`id`) on delete cascade;

create index `project__integration__workflow_id` on `project__integration`(`workflow_id`);
{{end}}

create table `group_role_mapping`
(
    `id`         integer primary key autoincrement,
    `provider`   varchar(100) not null,
    `group_name` varchar(255){{if .Mysql}} character set utf8mb4 collate utf8mb4_bin{{end}} not null,
    `project_id` int          not null,
    `role_id`    int          not null,
    `created`    datetime     not null,

    unique (`provider`, `group_name`, `project_id`, `role_id`),
    foreign key (`project_id`) references project (`id`) on delete cascade,
    foreign key (`role_id`) references `role` (`id`) on delete restrict
);

create index `group_role_mapping__role_id` on `group_role_mapping` (`role_id`);

create table `project__user_mapped_role`
(
    `project_id` int not null,
    `user_id`    int not null,
    `mapping_id` int not null,

    unique (`project_id`, `user_id`, `mapping_id`),
    foreign key (`project_id`) references project (`id`) on delete cascade,
    foreign key (`user_id`) references `user` (`id`) on delete cascade,
    foreign key (`mapping_id`) references `group_role_mapping` (`id`) on delete cascade
);

create index `project__user_mapped_role__user_id` on `project__user_mapped_role` (`user_id`);
create index `project__user_mapped_role__mapping_id` on `project__user_mapped_role` (`mapping_id`);
