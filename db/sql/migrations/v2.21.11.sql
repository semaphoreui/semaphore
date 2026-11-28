-- Webhook aliases and authorization move from integrations to endpoints and integration groups.
-- An endpoint holds the public aliases and the auth, a group picks an endpoint (and may override
-- its auth), an integration belongs to a group. legacy_integration_id remembers which integration
-- a row was made from while the data moves and is dropped at the end of this migration.
create table `project__endpoint`
(
    `id`                    integer primary key autoincrement,
    `project_id`            int          not null,
    `name`                  varchar(255) not null,
    `description`           text         null,
    `auth_method`           varchar(15)  not null default '',
    `auth_secret_id`        int          null,
    `auth_header`           varchar(255) null,
    `legacy_integration_id` int          null,

    foreign key (`project_id`) references project (`id`) on delete cascade,
    foreign key (`auth_secret_id`) references access_key (`id`) on delete set null
);

create index `project__endpoint__project_id` on `project__endpoint` (`project_id`);
create index `project__endpoint__auth_secret_id` on `project__endpoint` (`auth_secret_id`);

create table `project__endpoint_alias`
(
    `id`          integer primary key autoincrement,
    `alias`       varchar(50) not null,
    `project_id`  int         not null,
    `endpoint_id` int         not null,

    unique (`alias`),
    foreign key (`project_id`) references project (`id`) on delete cascade,
    foreign key (`endpoint_id`) references `project__endpoint` (`id`) on delete cascade
);

create index `project__endpoint_alias__project_id` on `project__endpoint_alias` (`project_id`);
create index `project__endpoint_alias__endpoint_id` on `project__endpoint_alias` (`endpoint_id`);

create table `project__integration_group`
(
    `id`                    integer primary key autoincrement,
    `project_id`            int          not null,
    `name`                  varchar(255) not null,
    `endpoint_id`           int          not null,
    `auth_override`         bool         not null default false,
    `auth_method`           varchar(15)  not null default '',
    `auth_secret_id`        int          null,
    `auth_header`           varchar(255) null,
    `legacy_integration_id` int          null,

    foreign key (`project_id`) references project (`id`) on delete cascade,
    foreign key (`endpoint_id`) references `project__endpoint` (`id`) on delete restrict,
    foreign key (`auth_secret_id`) references access_key (`id`) on delete set null
);

create index `project__integration_group__project_id` on `project__integration_group` (`project_id`);
create index `project__integration_group__endpoint_id` on `project__integration_group` (`endpoint_id`);
create index `project__integration_group__auth_secret_id` on `project__integration_group` (`auth_secret_id`);

alter table `project__integration` add `group_id` int null;

-- A project that had project-level aliases, or integrations reachable through them (searchable),
-- gets a "Default" endpoint with no auth: the project alias never checked anything itself.
insert into `project__endpoint` (`project_id`, `name`, `description`, `auth_method`, `auth_header`, `legacy_integration_id`)
select p.`id`, 'Default', 'Created on upgrade from the project aliases', '', '', null
from `project` p
where exists (select 1 from `project__integration_alias` a where a.`project_id` = p.`id` and a.`integration_id` is null)
   or exists (select 1 from `project__integration` i where i.`project_id` = p.`id` and i.`searchable` = true);

-- An integration with its own aliases becomes an endpoint of its own, carrying its auth.
-- 'none' was the column default, '' is what the application wrote: both mean no auth.
insert into `project__endpoint` (`project_id`, `name`, `description`, `auth_method`, `auth_secret_id`, `auth_header`, `legacy_integration_id`)
select i.`project_id`, i.`name`, '',
       case when i.`auth_method` = 'none' then '' else i.`auth_method` end,
       i.`auth_secret_id`, coalesce(i.`auth_header`, ''), i.`id`
from `project__integration` i
where i.`searchable` = false;

-- Project aliases go to the Default endpoint, an integration's own aliases to its endpoint.
-- An alias of a searchable integration matched nothing before (the lookup refused it) and is dropped.
insert into `project__endpoint_alias` (`alias`, `project_id`, `endpoint_id`)
select a.`alias`, a.`project_id`, e.`id`
from `project__integration_alias` a
join `project__endpoint` e on e.`project_id` = a.`project_id`
  and ((a.`integration_id` is null and e.`legacy_integration_id` is null)
    or a.`integration_id` = e.`legacy_integration_id`);

-- Every integration gets a group named after it. A searchable integration joins the Default
-- endpoint and keeps its own auth as an override; the others inherit from their own endpoint.
insert into `project__integration_group` (`project_id`, `name`, `endpoint_id`, `auth_override`, `auth_method`, `auth_secret_id`, `auth_header`, `legacy_integration_id`)
select i.`project_id`, i.`name`, e.`id`,
       case when i.`searchable` = true and i.`auth_method` not in ('', 'none') then true else false end,
       case when i.`searchable` = true and i.`auth_method` not in ('', 'none') then i.`auth_method` else '' end,
       case when i.`searchable` = true and i.`auth_method` not in ('', 'none') then i.`auth_secret_id` else null end,
       case when i.`searchable` = true and i.`auth_method` not in ('', 'none') then coalesce(i.`auth_header`, '') else '' end,
       i.`id`
from `project__integration` i
join `project__endpoint` e on e.`project_id` = i.`project_id`
  and ((i.`searchable` = true and e.`legacy_integration_id` is null)
    or (i.`searchable` = false and e.`legacy_integration_id` = i.`id`));

update `project__integration`
set `group_id` = (select g.`id` from `project__integration_group` g where g.`legacy_integration_id` = `project__integration`.`id`);

alter table `project__endpoint` drop column `legacy_integration_id`;
alter table `project__integration_group` drop column `legacy_integration_id`;

drop table `project__integration_alias`;

{{if .Sqlite}}
create table project__integration_new_2_21_11 (
    id             INTEGER PRIMARY KEY AUTOINCREMENT,
    name           VARCHAR(255) NOT NULL,
    project_id     INTEGER NOT NULL REFERENCES project(id) ON DELETE CASCADE,
    group_id       INTEGER NOT NULL REFERENCES project__integration_group(id) ON DELETE CASCADE,
    template_id    INTEGER NULL REFERENCES project__template(id) ON DELETE CASCADE,
    workflow_id    INTEGER NULL REFERENCES project__workflow_template(id) ON DELETE CASCADE,
    task_params_id INTEGER REFERENCES project__task_params(id)
);

insert into project__integration_new_2_21_11 (id, name, project_id, group_id, template_id, workflow_id, task_params_id)
    select id, name, project_id, group_id, template_id, workflow_id, task_params_id from project__integration;

drop table project__integration;

alter table project__integration_new_2_21_11 rename to project__integration;

create index project__integration__project_id on project__integration(project_id);

create index project__integration__group_id on project__integration(group_id);

create index project__integration__template_id on project__integration(template_id);

create index project__integration__workflow_id on project__integration(workflow_id);
{{else}}
alter table `project__integration` change `group_id` `group_id` int not null;

create index `project__integration__group_id` on `project__integration` (`group_id`);

alter table `project__integration` add constraint `project__integration__group_fk` foreign key (`group_id`) references `project__integration_group` (`id`) on delete cascade;

alter table `project__integration` drop column `auth_method`;

alter table `project__integration` drop column `auth_secret_id`;

alter table `project__integration` drop column `auth_header`;

alter table `project__integration` drop column `searchable`;
{{end}}
