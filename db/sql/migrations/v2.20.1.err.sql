alter table `event` drop column `integration_id`;
alter table `event` drop column `action`;
alter table `event` drop column `ip`;
alter table `event` drop column `user_agent`;

drop index `access_key__task_id`;
alter table `access_key` drop `task_id`;
alter table `access_key` drop `expire_at`;