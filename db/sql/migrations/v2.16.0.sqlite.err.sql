alter table `project` drop `default_secret_storage_id`;
alter table `project__environment` drop `secret_storage_id`;
alter table `project__environment` drop `secret_storage_key_prefix`;
alter table `access_key` drop `storage_id`;
alter table `access_key` drop `source_storage_id`;
alter table `access_key` drop `source_storage_key`;

drop table project__secret_storage;

alter table `access_key` drop `owner`;
alter table `access_key` drop `plain`;