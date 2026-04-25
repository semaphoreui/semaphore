alter table `user__token` add `expires_at` datetime null;
alter table `user__token` add `name` varchar(255) not null default '';