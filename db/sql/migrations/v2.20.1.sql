alter table `event` add column `integration_id` int;
alter table `event` add column `action` varchar(20);
alter table `event` add column `ip` varchar(45);
alter table `event` add column `user_agent` varchar(255);
