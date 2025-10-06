CREATE TABLE `task__file` (
  `id` INTEGER PRIMARY KEY AUTOINCREMENT,
  `task_id` INTEGER NOT NULL,
  `project_id` INTEGER NOT NULL,
  `filename` TEXT NOT NULL,
  `original_path` TEXT NOT NULL,
  `file_size` INTEGER NOT NULL,
  `mime_type` TEXT NOT NULL,
  `checksum` TEXT NOT NULL,
  `created` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX `idx_task_file_task_id` ON `task__file` (`task_id`);
CREATE INDEX `idx_task_file_project_id` ON `task__file` (`project_id`);
CREATE INDEX `idx_task_file_created` ON `task__file` (`created`);

