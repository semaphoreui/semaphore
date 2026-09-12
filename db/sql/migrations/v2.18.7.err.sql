drop index if exists `runner__registration_token`;
alter table `runner` drop column `registration_token`;
alter table `runner` drop column `registration_token_expires_at`;
