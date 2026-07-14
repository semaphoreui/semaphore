{{if .Mysql}}
alter table `user__external_identity` drop index `user__external_identity__type_provider_uid`;
{{else}}
drop index `user__external_identity__type_provider_uid`;
{{end}}

alter table `user__external_identity` drop column `type`;

create unique index `user__external_identity__provider_uid`
  on `user__external_identity`(`provider`, `external_uid`);
