-- Add project alert configuration table
CREATE TABLE project_alert_config (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    project_id INTEGER NOT NULL,
    alert_types TEXT NOT NULL,
    integrations TEXT NOT NULL,
    slack_config TEXT,
    teams_config TEXT,
    email_config TEXT,
    webhook_config TEXT,
    discord_config TEXT,
    pagerduty_config TEXT,
    created TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (project_id) REFERENCES projects(id) ON DELETE CASCADE,
    UNIQUE(project_id)
);

-- Create index for better performance
CREATE INDEX idx_project_alert_config_project_id ON project_alert_config(project_id);
