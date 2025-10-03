-- Add compliance framework fields to projects table
ALTER TABLE projects ADD COLUMN compliance_framework VARCHAR(50);
ALTER TABLE projects ADD COLUMN compliance_os VARCHAR(50);
ALTER TABLE projects ADD COLUMN compliance_version VARCHAR(50);
ALTER TABLE projects ADD COLUMN enable_stig BOOLEAN DEFAULT FALSE;

-- Add indexes for compliance queries
CREATE INDEX idx_projects_compliance_framework ON projects(compliance_framework);
CREATE INDEX idx_projects_compliance_os ON projects(compliance_os);
CREATE INDEX idx_projects_enable_stig ON projects(enable_stig);
