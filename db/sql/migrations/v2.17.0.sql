-- Add hidden and type fields to project__view table
alter table project__view add column `hidden` boolean not null default false;
alter table project__view add column `type` varchar(20) not null default '';
alter table project__view add column `filter` varchar(1000);
alter table project__view add column `sort_column` varchar(100);
alter table project__view add column `sort_reverse` boolean not null default false;

-- Create All view with position -1 for each existing project
insert into project__view (project_id, title, position, hidden, type)
select
    p.id as project_id,
    'All' as title,
    -1 as position,
    false as hidden,
    'all' as type
from project p
where not exists (
    select 1 from project__view pv
    where pv.project_id = p.id and pv.type = 'all'
    );