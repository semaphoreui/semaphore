create table `project__template_vault` (
    `id` integer primary key autoincrement,
    `project_id` int not null,
    `template_id` int not null,
    `vault_key_id` int not null,
    `name` varchar(255),

    unique (`template_id`, `vault_key_id`, `name`),
    foreign key (`project_id`) references project(`id`) on delete cascade,
    foreign key (`template_id`) references project__template(`id`) on delete cascade,
    foreign key (`vault_key_id`) references `access_key`(`id`) on delete cascade
);

insert into `project__template_vault` (template_id, project_id, vault_key_id)
select `id` as template_id, project_id, vault_key_id
from `project__template` where `vault_key_id` is not null;

alter table `project__template` drop column `vault_key_id`;
