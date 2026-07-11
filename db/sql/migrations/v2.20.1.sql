alter table `user__external_identity` add `type` varchar(4) not null default 'oidc';

update `user__external_identity` set `type` = 'ldap' where `provider` = 'ldap';

{{if .Mysql}}
alter table `user__external_identity` drop index `user__external_identity__provider_uid`;
{{else}}
drop index `user__external_identity__provider_uid`;
{{end}}

create unique index `user__external_identity__type_provider_uid`
  on `user__external_identity`(`type`, `provider`, `external_uid`);
