alter table `access_key` drop `storage_id`;
alter table `access_key` drop `source_storage_id`;
alter table `access_key` drop `source_storage_key`;

drop table project__secret_storage;

alter table `access_key` drop `owner`;
alter table `access_key` drop `plain`;