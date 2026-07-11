drop index if exists `user__external_identity__provider_uid`;

alter table `user__external_identity` drop index `user__external_identity__provider_uid`;
