CREATE TABLE `event` (
  `project_id` int(11) DEFAULT NULL,
  `object_id` int(11) DEFAULT NULL,
  `object_type` varchar(20) DEFAULT '',
  `description` text,
  `created` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  KEY `project_id` (`project_id`),
  KEY `object_id` (`object_id`),
  KEY `created` (`created`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;