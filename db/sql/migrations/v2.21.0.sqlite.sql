-- Add tags and labels fields to templates table
ALTER TABLE `templates` ADD COLUMN `tags` TEXT DEFAULT '[]';
ALTER TABLE `templates` ADD COLUMN `labels` TEXT DEFAULT '[]';

-- Add tags and labels fields to tasks table
ALTER TABLE `tasks` ADD COLUMN `tags` TEXT DEFAULT '[]';
ALTER TABLE `tasks` ADD COLUMN `labels` TEXT DEFAULT '[]';

-- Note: SQLite doesn't support JSON indexes in the same way as MySQL
-- The JSON functions will still work for filtering, but without dedicated indexes
