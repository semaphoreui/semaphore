{{if .Mysql}}
drop index `task__ansible_host__host_id` on `task__ansible_host`;
{{else}}
drop index if exists `task__ansible_host__host_id`;
{{end}}
alter table `task__ansible_host` drop column `host_id`;
drop table `host__template`;
drop table `host__project`;
drop table `host`;
