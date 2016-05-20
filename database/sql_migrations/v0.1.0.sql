alter table user change `created` `created` datetime not null default current_timestamp;
alter table project change `created` `created` datetime not null default current_timestamp comment 'Created timestamp';
alter table task change `created` `created` datetime not null default current_timestamp;
alter table user__token change `created` `created` datetime not null default current_timestamp;