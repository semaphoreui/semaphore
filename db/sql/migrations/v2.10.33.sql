alter table `project__template_vault` change `vault_key_id` `vault_key_id` int;
alter table `project__template_vault` add `type` varchar(20) not null default 'password';
alter table `project__template_vault` add `script` text;
update `project__template_vault` set `type` = 'password' where `vault_key_id` IS NOT NULL;