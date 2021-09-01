alter table `project__inventory` add `become_key_id` int references access_key(`id`);
alter table `project__template` add `vault_pass_id` int references access_key(`id`);
