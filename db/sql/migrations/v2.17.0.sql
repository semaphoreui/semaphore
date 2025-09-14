-- Add hidden and type fields to project__view table
alter table project__view add column `hidden` boolean not null default false;
alter table project__view add column `type` varchar(20) not null default 'custom';

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

-- Remove any special All tab settings views that were previously created
delete from project__view where title = '__all_tab_settings__';