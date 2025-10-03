-- Add compliance framework fields to projects table (SQLite)
ALTER TABLE projects ADD COLUMN compliance_framework TEXT;
ALTER TABLE projects ADD COLUMN compliance_os TEXT;
ALTER TABLE projects ADD COLUMN compliance_version TEXT;
ALTER TABLE projects ADD COLUMN enable_stig INTEGER DEFAULT 0;

-- Add indexes for compliance queries
CREATE INDEX idx_projects_compliance_framework ON projects(compliance_framework);
CREATE INDEX idx_projects_compliance_os ON projects(compliance_os);
CREATE INDEX idx_projects_enable_stig ON projects(enable_stig);
