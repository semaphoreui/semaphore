alter table project__template add column hosts_limit text default null after view_id;
alter table task modify hosts_limit text default null;
