alter table `project__template` add `vault_key_id` int references access_key(`id`);
update `project__template` set `vault_key_id` = `vault_pass_id` where `vault_key_id` is null;
alter table `project__template` drop column `vault_pass_id`;