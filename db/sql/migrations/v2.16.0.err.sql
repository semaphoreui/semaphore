alter table `access_key` drop `storage_id`;
alter table `access_key` drop `linked_storage_id`;
alter table `access_key` drop `linked_storage_key`;

drop table secret_storage;

alter table `access_key` drop `owner`;
alter table `access_key` drop `plain`;