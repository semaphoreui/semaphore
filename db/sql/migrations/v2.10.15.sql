alter table `access_key` add `environment_id` int null references project__environment(`id`) on delete cascade;
alter table `access_key` add `user_id` int null references `user`(`id`) on delete cascade;
alter table `project__integration` add `task_params` text null;
alter table `task` add `schedule_id` int null references project__schedule(`id`) on delete set null;