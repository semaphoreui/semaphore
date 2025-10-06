CREATE TABLE `task__file` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `task_id` int(11) NOT NULL,
  `project_id` int(11) NOT NULL,
  `filename` varchar(255) NOT NULL,
  `original_path` varchar(500) NOT NULL,
  `file_size` bigint(20) NOT NULL,
  `mime_type` varchar(100) NOT NULL,
  `checksum` varchar(32) NOT NULL,
  `created` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `idx_task_file_task_id` (`task_id`),
  KEY `idx_task_file_project_id` (`project_id`),
  KEY `idx_task_file_created` (`created`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

