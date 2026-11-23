alter table `project__inventory_host` add column `last_status` varchar(20) null;
alter table `project__inventory_host` add column `last_run` datetime null;
alter table `project__inventory_host` add column `last_task_id` int null;
alter table `project__inventory_host` add column `host_id` int null;
alter table `task` add column `effective_inventory_id` int null;

create index `project__inventory_host__host_id` on `project__inventory_host` (`host_id`);
