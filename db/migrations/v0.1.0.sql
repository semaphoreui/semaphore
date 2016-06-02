alter table user change `created` `created` datetime not null;
alter table project change `created` `created` datetime not null comment 'Created timestamp';
alter table task change `created` `created` datetime not null;
alter table user__token change `created` `created` datetime not null;

alter table task drop foreign key `task_ibfk_1`;
alter table task add constraint `task_ibfk_1` foreign key (`template_id`) references `project__template` (`id`) on delete cascade;