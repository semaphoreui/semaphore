alter table `runner` add column `registration_token` varchar(255);
alter table `runner` add column `registration_token_expires_at` datetime;
create index `runner__registration_token` on `runner` (`registration_token`);
