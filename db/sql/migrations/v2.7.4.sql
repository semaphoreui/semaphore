alter table `event` add `user_id` int null references `user`(`id`);
