alter table `task` add `runner_id` int null references `runner`(`id`) on delete set null;
