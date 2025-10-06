-- Add tags and labels fields to templates table
ALTER TABLE `templates` ADD COLUMN `tags` TEXT COMMENT 'JSON array of tag strings for flexible categorization and filtering';
ALTER TABLE `templates` ADD COLUMN `labels` TEXT COMMENT 'JSON array of label strings for hierarchical organization and color coding';

-- Add tags and labels fields to tasks table
ALTER TABLE `tasks` ADD COLUMN `tags` TEXT COMMENT 'JSON array of tag strings inherited from template';
ALTER TABLE `tasks` ADD COLUMN `labels` TEXT COMMENT 'JSON array of label strings inherited from template';

-- Add indexes for better performance on tag and label filtering
CREATE INDEX `idx_templates_tags` ON `templates` ((CAST(`tags` AS JSON)));
CREATE INDEX `idx_templates_labels` ON `templates` ((CAST(`labels` AS JSON)));
CREATE INDEX `idx_tasks_tags` ON `tasks` ((CAST(`tags` AS JSON)));
CREATE INDEX `idx_tasks_labels` ON `tasks` ((CAST(`labels` AS JSON)));
