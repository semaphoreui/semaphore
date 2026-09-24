delete from `role` where `is_builtin` = true;

alter table `role` drop column `is_builtin`;
