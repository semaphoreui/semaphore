delete
from project__schedule
where template_id is null;

delete
from project__schedule
where (select count(*) from project__template where project__template.id = project__schedule.template_id) = 0;

delete
from project__schedule
where (select count(*) from project where project.id = project__schedule.project_id) = 0;

update project__schedule
set repository_id = null
where repository_id is not null
  and (select count(*) from project__repository where project__repository.id = project__schedule.repository_id) = 0;

alter table `project__schedule`
    rename to `project__schedule_backup_8436583`;

create table project__schedule
(
    id               integer primary key autoincrement,
    template_id      int          not null,
    project_id       int          not null,
    cron_format      varchar(255) not null,
    repository_id    int          null,
    last_commit_hash varchar(40)  null,

    foreign key (`template_id`) references project__template(`id`) on delete cascade,
    foreign key (`project_id`) references project(`id`) on delete cascade,
    foreign key (`repository_id`) references project__repository(`id`)
);

insert into project__schedule
select *
from project__schedule_backup_8436583;

drop table project__schedule_backup_8436583;
