alter table `task` drop column `effective_inventory_id`;
{{if .Mysql}}drop index `project__inventory_host__host_id` on `project__inventory_host`;{{else}}drop index if exists `project__inventory_host__host_id`;{{end}}
alter table `project__inventory_host` drop column `host_id`;
alter table `project__inventory_host` drop column `last_task_id`;
alter table `project__inventory_host` drop column `last_run`;
alter table `project__inventory_host` drop column `last_status`;
