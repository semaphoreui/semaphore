ALTER TABLE project__inventory DROP FOREIGN KEY IF EXISTS project__inventory_ibfk_2;

alter table `project__inventory` drop column `key_id`;

ALTER TABLE project__template DROP FOREIGN KEY IF EXISTS project__template_ibfk_2;

alter table `project__template` drop column `ssh_key_id`;