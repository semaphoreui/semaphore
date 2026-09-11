alter table `access_key` add `task_id` int null references task(`id`) on delete cascade;
alter table `access_key` add `expire_at` datetime null;
create index `access_key__task_id` on `access_key`(`task_id`);
alter table `event` add column `integration_id` int;
alter table `event` add column `action` varchar(20);
alter table `event` add column `ip` varchar(45);
alter table `event` add column `user_agent` varchar(255);
