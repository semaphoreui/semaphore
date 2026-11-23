alter table `task` drop column `effective_inventory_id`;
alter table `project__inventory_host` drop column `host_id`;
alter table `project__inventory_host` drop column `last_task_id`;
alter table `project__inventory_host` drop column `last_run`;
alter table `project__inventory_host` drop column `last_status`;
