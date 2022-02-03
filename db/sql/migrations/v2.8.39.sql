delete from `project__template` where `removed` = true;

alter table `project__template` drop column `removed`;
