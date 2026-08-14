drop index `access_key__task_id`;
alter table `access_key` drop `task_id`;
alter table `access_key` drop `expire_at`;