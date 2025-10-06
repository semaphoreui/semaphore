-- Add folder column to templates table for organizing templates
ALTER TABLE templates ADD COLUMN folder VARCHAR(255) NULL;

-- Create index for better performance when filtering by folder
CREATE INDEX idx_templates_folder ON templates(project_id, folder);
