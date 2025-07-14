create table user__email_otp(
    `id` integer primary key autoincrement,
    `user_id` int NOT NULL,
    `code` varchar(250) NOT NULL,
    `created` datetime NOT NULL,
    unique (`code`),
    foreign key (`user_id`) references `user`(`id`) on delete cascade
);