{{if .Mysql}}
drop index `project__inventory_group__source_id` on `project__inventory_group`;
drop index `project__inventory_host__source_id` on `project__inventory_host`;
{{else}}
drop index if exists `project__inventory_group__source_id`;
drop index if exists `project__inventory_host__source_id`;
{{end}}

alter table `project__inventory_group` drop column `source_id`;
alter table `project__inventory_host` drop column `source_id`;

drop table `project__inventory_source`;
