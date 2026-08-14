alter table `access_key` add `task_id` int null references task(`id`) on delete cascade;
alter table `access_key` add `expire_at` datetime null;
create index `access_key__task_id` on `access_key`(`task_id`);
