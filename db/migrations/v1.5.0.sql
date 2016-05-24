CREATE TABLE `session` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `user_id` int(11) NOT NULL,
  `created` datetime NOT NULL,
  `last_active` datetime NOT NULL,
  `ip` varchar(15) NOT NULL DEFAULT '',
  `user_agent` text NOT NULL,
  `expired` tinyint(1) NOT NULL DEFAULT '0',
  PRIMARY KEY (`id`),
  KEY `user_id` (`user_id`),
  KEY `expired` (`expired`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;