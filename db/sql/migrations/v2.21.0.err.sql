delete from `access_key` where `owner` = 'mcp';

{{if .Mysql}}
alter table `access_key` drop foreign key `access_key__mcp_fk`;

drop index `access_key__mcp_id` on `access_key`;
{{else if .Postgresql}}
alter table `access_key` drop foreign key `access_key__mcp_fk`;

drop index if exists `access_key__mcp_id`;
{{else}}
drop index if exists `access_key__mcp_id`;
{{end}}

alter table `access_key` drop column `mcp_id`;

drop table `project__inventory_mcp`;
