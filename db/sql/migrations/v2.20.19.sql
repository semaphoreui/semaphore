create table `host`
(
    `id`                 integer primary key autoincrement,
    `name`               varchar(250){{if .Mysql}} character set utf8mb4 collate utf8mb4_bin{{end}} not null,
    `display_name`       varchar(250) not null,
    `first_seen`         datetime not null,
    `last_seen`          datetime not null,
    `last_automated`     datetime null,
    `runs`               int not null default 0,
    `last_status`        varchar(20) not null,
    `last_task_id`       int null,
    `last_project_id`    int null,
    `decommissioned`     datetime null,
    `decommission_count` int not null default 0,

    unique (`name`)
);

create table `host__project`
(
    `id`           integer primary key autoincrement,
    `host_id`      int not null,
    `project_id`   int not null,
    `first_seen`   datetime not null,
    `last_seen`    datetime not null,
    `runs`         int not null default 0,
    `last_status`  varchar(20) not null,
    `last_task_id` int null,

    unique (`host_id`, `project_id`),
    foreign key (`host_id`) references `host` (`id`) on delete cascade,
    foreign key (`project_id`) references `project` (`id`) on delete cascade
);

create index `host__project__project_id` on `host__project` (`project_id`);

create table `host__template`
(
    `id`           integer primary key autoincrement,
    `host_id`      int not null,
    `template_id`  int not null,
    `project_id`   int not null,
    `last_task_id` int null,

    unique (`host_id`, `template_id`),
    foreign key (`host_id`) references `host` (`id`) on delete cascade,
    foreign key (`template_id`) references `project__template` (`id`) on delete cascade
);

create index `host__template__project_id` on `host__template` (`project_id`, `host_id`);

alter table `task__ansible_host` add `host_id` int null;

create index `task__ansible_host__host_id` on `task__ansible_host` (`host_id`);

-- Hosts of past tasks except local connections, grouped by bytes on MySQL like the unique key on name.
insert into `host` (`name`, `display_name`, `first_seen`, `last_seen`, `runs`, `last_status`, `last_task_id`)
select min(lower(h.`host`)), min(h.`host`), min(t.`created`), max(t.`created`), count(distinct h.`task_id`), 'ok', max(h.`task_id`)
from `task__ansible_host` h
join `task` t on t.`id` = h.`task_id`
where lower(h.`host`) not in ('localhost', '127.0.0.1', '::1')
group by {{if .Mysql}}binary {{end}}lower(h.`host`);

update `task__ansible_host`
set `host_id` = (select `host`.`id` from `host` where `host`.`name` = lower(`task__ansible_host`.`host`));

update `host`
set `last_automated` = (
    select max(t.`created`)
    from `task__ansible_host` h
    join `task` t on t.`id` = h.`task_id`
    where h.`host_id` = `host`.`id` and h.`unreachable` = 0
);

-- The last status comes from any one row of the host in its last task, not the worst one.
update `host`
set `last_status` = (
        select case when h.`unreachable` > 0 then 'unreachable' when h.`failed` > 0 then 'failed' when h.`changed` > 0 then 'changed' else 'ok' end
        from `task__ansible_host` h
        where h.`host_id` = `host`.`id` and h.`task_id` = `host`.`last_task_id`
        limit 1
    ),
    `last_project_id` = (
        select h.`project_id`
        from `task__ansible_host` h
        where h.`host_id` = `host`.`id` and h.`task_id` = `host`.`last_task_id`
        limit 1
    );

insert into `host__project` (`host_id`, `project_id`, `first_seen`, `last_seen`, `runs`, `last_status`, `last_task_id`)
select h.`host_id`, h.`project_id`, min(t.`created`), max(t.`created`), count(distinct h.`task_id`), 'ok', max(h.`task_id`)
from `task__ansible_host` h
join `task` t on t.`id` = h.`task_id`
where h.`host_id` is not null
group by h.`host_id`, h.`project_id`;

update `host__project`
set `last_status` = (
    select case when h.`unreachable` > 0 then 'unreachable' when h.`failed` > 0 then 'failed' when h.`changed` > 0 then 'changed' else 'ok' end
    from `task__ansible_host` h
    where h.`host_id` = `host__project`.`host_id` and h.`task_id` = `host__project`.`last_task_id`
    limit 1
);

insert into `host__template` (`host_id`, `template_id`, `project_id`, `last_task_id`)
select h.`host_id`, t.`template_id`, h.`project_id`, max(h.`task_id`)
from `task__ansible_host` h
join `task` t on t.`id` = h.`task_id`
where h.`host_id` is not null
group by h.`host_id`, t.`template_id`, h.`project_id`;
