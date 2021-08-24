CREATE TABLE `session` (
  `id` integer primary key autoincrement,
  `user_id` int NOT NULL,
  `created` datetime NOT NULL,
  `last_active` datetime NOT NULL,
  `ip` varchar(15) NOT NULL DEFAULT '',
  `user_agent` text NOT NULL,
  `expired` boolean NOT NULL DEFAULT false
);

CREATE INDEX `user_id` ON `session`(`user_id`);

CREATE INDEX `expired` ON `session`(`expired`);
