-- Rollback puts aliases and authorization back on integrations. The original split cannot be
-- recovered, so it is derived: an endpoint reached by exactly one integration gives that
-- integration its aliases and auth as its own (searchable = false), an endpoint with several
-- integrations becomes project aliases and its integrations turn searchable, each with the auth
-- its group checked (its own override, else the endpoint's). An alias of an endpoint with no
-- integrations becomes a project alias that starts nothing, which is what it did.
alter table `project__integration` add `auth_method` varchar(15) not null default '';

alter table `project__integration` add `auth_secret_id` int null;

alter table `project__integration` add `auth_header` varchar(255) null;

alter table `project__integration` add `searchable` bool not null default false;

alter table `project__endpoint` add `integration_count` int not null default 0;

update `project__endpoint`
set `integration_count` = (select count(*)
                           from `project__integration` i
                           join `project__integration_group` g on g.`id` = i.`group_id`
                           where g.`endpoint_id` = `project__endpoint`.`id`);

update `project__integration`
set `auth_method`    = (select case when g.`auth_override` = true then g.`auth_method` else e.`auth_method` end
                        from `project__integration_group` g
                        join `project__endpoint` e on e.`id` = g.`endpoint_id`
                        where g.`id` = `project__integration`.`group_id`),
    `auth_secret_id` = (select case when g.`auth_override` = true then g.`auth_secret_id` else e.`auth_secret_id` end
                        from `project__integration_group` g
                        join `project__endpoint` e on e.`id` = g.`endpoint_id`
                        where g.`id` = `project__integration`.`group_id`),
    `auth_header`    = (select case when g.`auth_override` = true then g.`auth_header` else e.`auth_header` end
                        from `project__integration_group` g
                        join `project__endpoint` e on e.`id` = g.`endpoint_id`
                        where g.`id` = `project__integration`.`group_id`),
    `searchable`     = (select case when e.`integration_count` > 1 then true else false end
                        from `project__integration_group` g
                        join `project__endpoint` e on e.`id` = g.`endpoint_id`
                        where g.`id` = `project__integration`.`group_id`);

create table `project__integration_alias`
(
    `id`             integer primary key autoincrement,
    `alias`          varchar(50) not null,
    `project_id`     int not null,
    `integration_id` int,

    foreign key (`project_id`) references project (`id`) on delete cascade,
    foreign key (`integration_id`) references `project__integration` (`id`) on delete cascade,

    unique (`alias`)
);

create index `project__integration_alias__project_id` on `project__integration_alias` (`project_id`);
create index `project__integration_alias__integration_id` on `project__integration_alias` (`integration_id`);

insert into `project__integration_alias` (`alias`, `project_id`, `integration_id`)
select a.`alias`, a.`project_id`,
       case when e.`integration_count` = 1
            then (select i.`id`
                  from `project__integration` i
                  join `project__integration_group` g on g.`id` = i.`group_id`
                  where g.`endpoint_id` = e.`id`)
            else null end
from `project__endpoint_alias` a
join `project__endpoint` e on e.`id` = a.`endpoint_id`;

{{if .Sqlite}}
create table project__integration_old_2_21_11 (
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

insert into project__integration_old_2_21_11 (id, name, project_id, template_id, workflow_id, auth_method, auth_secret_id, auth_header, searchable, task_params_id)
    select id, name, project_id, template_id, workflow_id, auth_method, auth_secret_id, auth_header, searchable, task_params_id from project__integration;

drop table project__integration;

alter table project__integration_old_2_21_11 rename to project__integration;

create index project__integration__auth_secret_id on project__integration(auth_secret_id);

create index project__integration__project_id on project__integration(project_id);

create index project__integration__template_id on project__integration(template_id);

create index project__integration__workflow_id on project__integration(workflow_id);
{{else}}
alter table `project__integration` drop foreign key `project__integration__group_fk`;

alter table `project__integration` drop column `group_id`;

alter table `project__integration` add constraint `project__integration__auth_secret_fk` foreign key (`auth_secret_id`) references access_key (`id`) on delete set null;
{{end}}

drop table `project__integration_group`;
drop table `project__endpoint_alias`;
drop table `project__endpoint`;
