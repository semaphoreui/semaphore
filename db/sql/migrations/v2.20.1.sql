alter table `user__external_identity` add `type` varchar(16) not null default 'oidc';

update `user__external_identity` set `type` = 'ldap' where `provider` = 'ldap';

drop index `user__external_identity__provider_uid`;

create unique index `user__external_identity__type_provider_uid`
  on `user__external_identity`(`type`, `provider`, `external_uid`);
