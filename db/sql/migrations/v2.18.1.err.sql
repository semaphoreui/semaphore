alter table `project__secret_storage` drop column `sync_enabled`;
alter table `project__secret_storage` drop column `sync_interval`;
alter table `project__secret_storage` drop column `last_synced_at`;
alter table `project__secret_storage` drop column `last_sync_failed_at`;

drop table `project__secret_storage__sync_path`;